package command

import (
	"errors"

	"github.com/robby031/totp-cli/internal/account"
)

var errNoAccounts = errors.New("belum ada akun, gunakan: totp add <name>")

func (r *Runner) selectAccounts(args []string) ([]account.Account, error) {
	if len(args) > 1 {
		return nil, errUsage
	}

	if len(args) == 1 {
		acc, err := r.store.Find(args[0])
		if err != nil {
			return nil, friendlyError(err, args[0])
		}
		return []account.Account{acc}, nil
	}

	accounts, err := r.store.List()
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return nil, errNoAccounts
	}
	return accounts, nil
}
