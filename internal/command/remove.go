package command

import "fmt"

func (r *Runner) runRemove(args []string) error {
	name, err := requireName(args)
	if err != nil {
		return err
	}
	if err := r.store.Remove(name); err != nil {
		return friendlyError(err, name)
	}
	fmt.Fprintf(r.out, "akun %q dihapus\n", name)
	return nil
}
