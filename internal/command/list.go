package command

import "fmt"

func (r *Runner) runList() error {
	accounts, err := r.store.List()
	if err != nil {
		return err
	}
	for _, acc := range accounts {
		fmt.Fprintln(r.out, acc.Name)
	}
	return nil
}
