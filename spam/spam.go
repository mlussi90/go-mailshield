package spam

import (
	"bytes"
	"fmt"
	"os/exec"
)

func CheckWithSpamc(raw []byte) (bool, float64, float64, error) {
	cmd := exec.Command("docker", "exec", "-i", "spamassassin", "spamc", "-c")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return false, 0, 0, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	go func() {
		defer stdin.Close()
		_, _ = stdin.Write(raw)
	}()

	out, errRun := cmd.CombinedOutput()
	isSpam := false
	var score, required float64

	if errRun != nil {
		if ee, ok := errRun.(*exec.ExitError); ok {
			isSpam = (ee.ExitCode() == 1)
		} else {
			return false, 0, 0, fmt.Errorf("spamc run: %w", errRun)
		}
	}

	if _, err := fmt.Sscanf(string(bytes.TrimSpace(out)), "%f/%f", &score, &required); err != nil {
		return false, 0, 0, fmt.Errorf("failed to parse spamc output: %w", err)
	}

	if !isSpam && score >= required {
		isSpam = true
	}

	return isSpam, score, required, nil
}
