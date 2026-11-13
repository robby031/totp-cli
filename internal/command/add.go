package command

import (
	"errors"
	"fmt"

	"totp-cli/internal/account"
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

	secret, err := r.readLine("Secret (Base32): ", true)
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
	fmt.Fprintf(r.out, "Account %q added\n", acc.Name)
	return nil
}
