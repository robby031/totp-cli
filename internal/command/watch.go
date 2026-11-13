package command

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"totp-cli/internal/account"
)

const clearScreen = "\033[H\033[J"

func (r *Runner) runWatch(args []string) error {
	accounts, err := r.selectAccounts(args)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return r.watchCodes(ctx, accounts)
}

func (r *Runner) watchCodes(ctx context.Context, accounts []account.Account) error {
	for {
		now := time.Now()
		if err := r.renderFrame(accounts, now); err != nil {
			return err
		}

		timer := time.NewTimer(time.Until(now.Truncate(time.Second).Add(time.Second)))
		select {
		case <-ctx.Done():
			timer.Stop()
			fmt.Fprintln(r.out)
			return nil
		case <-timer.C:
		}
	}
}

func (r *Runner) renderFrame(accounts []account.Account, now time.Time) error {
	var frame bytes.Buffer
	frame.WriteString(clearScreen)
	if err := writeCodes(&frame, accounts, now); err != nil {
		return err
	}
	frame.WriteString("\nCtrl+C untuk keluar\n")

	_, err := r.out.Write(frame.Bytes())
	return err
}
