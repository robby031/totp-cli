package command

import (
	"fmt"
	"io"
	"time"

	"github.com/robby031/totp-cli/internal/account"
)

func writeCodes(w io.Writer, accounts []account.Account, now time.Time) error {
	remaining := account.RemainingSeconds(now)
	for _, acc := range accounts {
		code, err := account.GenerateCode(acc.Secret)
		if err != nil {
			return fmt.Errorf("akun %q: %w", acc.Name, err)
		}
		fmt.Fprintf(w, "%-20s %s  (%2ds)\n", acc.Name, code, remaining)
	}
	return nil
}
