package command

import "time"

func (r *Runner) runCode(args []string) error {
	accounts, err := r.selectAccounts(args)
	if err != nil {
		return err
	}
	return writeCodes(r.out, accounts, time.Now())
}
