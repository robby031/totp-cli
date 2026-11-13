package command

import (
	"errors"
	"fmt"

	"github.com/robby031/totp-cli/internal/account"
)

func (r *Runner) runAdd(args []string) error {
	name, err := requireName(args)
	if err != nil {
		return err
	}

	acc := account.Account{Name: account.NormalizeName(name)}
	if acc.Name == "" {
		return account.ErrEmptyName
	}

	secret, err := r.readVisibleAndClear("Secret (Base32): ")
	if err != nil {
		return err
	}
	if secret == "" {
		return errors.New("secret is required")
	}
	acc.Secret = account.NormalizeSecret(secret)

	if _, err := account.GenerateCode(acc.Secret); err != nil {
		return fmt.Errorf("invalid secret: %w", err)
	}

	if err := r.store.Add(acc); err != nil {
		return friendlyError(err, name)
	}
	fmt.Fprintf(r.out, "Akun %q ditambahkan\n", acc.Name)
	return nil
}
