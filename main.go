package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func main() {
	cfg := loadConfig("config.yaml")
	fmt.Println("config loaded")

	pollInterval, _ := time.ParseDuration(cfg.PollInterval)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	for _, acc := range cfg.Accounts {
		go processAccount(ctx, acc, pollInterval)
	}

	<-ctx.Done()
	fmt.Println("shutdown")
}

func processAccount(ctx context.Context, acc IMAPAccount, pollInterval time.Duration) {
	fmt.Printf("[%s] connecting to %s \n", acc.Name, acc.Host)
	var c *client.Client
	var err error

retry:
	c, err = client.DialTLS(acc.Host, nil)
	if err != nil {
		fmt.Printf("[%s] dial error: %v, retrying in 10s \n", acc.Name, err)
		select {
		case <-time.After(10 * time.Second):
			goto retry
		case <-ctx.Done():
			return
		}
	}
	defer c.Logout()

	if err := c.Login(acc.Username, acc.Password); err != nil {
		fmt.Printf("[%s] login error: %v \n", acc.Name, err)
		return
	}
	fmt.Printf("[%s] logged in \n", acc.Name)

	sel, err := c.Select(acc.Inbox, false)
	if err != nil {
		fmt.Printf("[%s] select inbox error: %v \n", acc.Name, err)
		return
	}
	fmt.Printf("[%s] %s selected, %d msgs \n", acc.Name, acc.Inbox, sel.Messages)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		if ctx.Err() != nil {
			return
		}
		fmt.Printf("[%s] processing mailbox \n", acc.Name)
		if err := handleMailbox(ctx, c, acc); err != nil {
			fmt.Printf("[%s] process error: %v \n", acc.Name, err)
			time.Sleep(5 * time.Second)
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}

func handleMailbox(ctx context.Context, c *client.Client, acc IMAPAccount) error {
	_, err := c.Select(acc.Inbox, false)
	if err != nil {
		return fmt.Errorf("[%s] reselect: %w", acc.Name, err)
	}

	criteria := imap.NewSearchCriteria()
	if acc.SearchUnseenOnly {
		criteria.WithoutFlags = []string{imap.SeenFlag}
	} else {
		since := time.Now().Add(-24 * time.Hour)
		criteria.Since = since
	}

	uids, err := c.Search(criteria)
	if err != nil {
		return fmt.Errorf("[%s] search: %w", acc.Name, err)
	}

	if len(uids) == 0 {
		return nil
	}

	count := 0
	for _, uid := range uids {
		count++

		oneSeqSet := new(imap.SeqSet)
		oneSeqSet.AddNum(uid)

		fetchItems := []imap.FetchItem{imap.FetchUid, imap.FetchFlags, imap.FetchItem("BODY.PEEK[]")}

		msgCh := make(chan *imap.Message, 1)
		done := make(chan error, 1)

		seqSet := new(imap.SeqSet)
		seqSet.AddNum(uid)
		go func() {
			done <- c.Fetch(seqSet, fetchItems, msgCh)
		}()

		var fetchMsg *imap.Message
		select {
		case msg := <-msgCh:
			fetchMsg = msg
		case err := <-done:
			if err != nil {
				fmt.Printf("[%s] error fetching UID %v: %v\n", acc.Name, uid, err)
				continue
			}
		case <-time.After(10 * time.Second):
			fmt.Printf("[%s] timeout fetching UID %v\n", acc.Name, uid)
			continue
		}

		if fetchMsg == nil {
			fmt.Printf("[%s] no message received for UID %v\n", acc.Name, uid)
			continue
		}

		section := &imap.BodySectionName{}
		r := fetchMsg.GetBody(section)
		if r == nil {
			fmt.Printf("[%s] uid %d: no body\n", acc.Name, fetchMsg.Uid)
			continue
		}
		var raw bytes.Buffer
		if _, err := io.Copy(&raw, r); err != nil {
			fmt.Printf("[%s] error reading message: %v\n", acc.Name, err)
			continue
		}

		fmt.Printf("[%s] uid %d: processing mail (size: %d)\n",
			acc.Name, fetchMsg.Uid, raw.Len())

		isSpam, score, required, saErr := checkWithSpamc(raw.Bytes())
		if saErr != nil {
			fmt.Printf("[%s] uid %d: spamc error: %v\n", acc.Name, fetchMsg.Uid, saErr)
			continue
		}

		fmt.Printf("[%s] uid %d: score=%.2f required=%.2f -> spam=%v\n",
			acc.Name, fetchMsg.Uid, score, required, isSpam)

		if isSpam {
			if err := moveUID(c, fetchMsg.Uid, acc.SpamFolder); err != nil {
				fmt.Printf("[%s] uid %d: move error: %v\n", acc.Name, fetchMsg.Uid, err)
				continue
			}
			fmt.Printf("[%s] uid %d: moved to folder %v\n", acc.Name, fetchMsg.Uid, acc.SpamFolder)
		}
	}

	fmt.Printf("[%s] mails processed: %d\n", acc.Name, count)
	return nil
}

func checkWithSpamc(raw []byte) (isSpam bool, score, required float64, err error) {
	cmd := exec.Command("docker", "exec", "-i", "spamassassin", "spamc", "-c")
	stdin, _ := cmd.StdinPipe()
	go func() {
		defer stdin.Close()
		stdin.Write(raw)
	}()

	out, errRun := cmd.CombinedOutput()
	if errRun != nil {
		if ee, ok := errRun.(*exec.ExitError); ok {
			isSpam = (ee.ExitCode() == 1)
		} else {
			return false, 0, 0, fmt.Errorf("spamc run: %w", errRun)
		}
	}

	fmt.Sscanf(string(bytes.TrimSpace(out)), "%f/%f", &score, &required)
	if !isSpam && score >= required {
		isSpam = true
	}
	return
}

func moveUID(c *client.Client, uid uint32, dest string) error {
	caps, err := c.Capability()
	if err == nil && caps["MOVE"] {
		seq := new(imap.SeqSet)
		seq.AddNum(uid)
		return c.UidMove(seq, dest)
	}

	seq := new(imap.SeqSet)
	seq.AddNum(uid)
	if err := c.UidCopy(seq, dest); err != nil {
		return err
	}
	item := imap.FormatFlagsOp(imap.AddFlags, true)
	if err := c.UidStore(seq, item, []interface{}{imap.DeletedFlag}, nil); err != nil {
		return err
	}
	return c.Expunge(nil)
}
