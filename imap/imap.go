package imap

import (
	"fmt"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func MoveUID(c *client.Client, uid uint32, dest string) error {
	caps, err := c.Capability()
	if err != nil {
		return fmt.Errorf("failed to get capabilities: %w", err)
	}
	if caps["MOVE"] {
		return moveNative(c, uid, dest)
	}
	return moveFallback(c, uid, dest)
}

func moveNative(c *client.Client, uid uint32, dest string) error {
	seq := singleSeqSet(uid)
	return c.UidMove(seq, dest)
}

func moveFallback(c *client.Client, uid uint32, dest string) error {
	seq := singleSeqSet(uid)
	if err := c.UidCopy(seq, dest); err != nil {
		return err
	}
	item := imap.FormatFlagsOp(imap.AddFlags, true)
	if err := c.UidStore(seq, item, []any{imap.DeletedFlag}, nil); err != nil {
		return err
	}
	return c.Expunge(nil)
}

func singleSeqSet(uid uint32) *imap.SeqSet {
	seq := new(imap.SeqSet)
	seq.AddNum(uid)
	return seq
}
