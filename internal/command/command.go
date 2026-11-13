package command

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"totp-cli/internal/storage"
)

const usage = `Usage:
  totp add <name>       Tambah akun
  totp code [name]      Tampilkan kode sekali untuk satu akun, atau semua akun
  totp watch [name]     Pantau kode secara realtime
  totp list             Tampilkan daftar akun
  totp remove <name>    Hapus akun`

var errUsage = errors.New(usage)

type Runner struct {
	store  *storage.Store
	in     io.Reader
	out    io.Writer
	errOut io.Writer
	reader *bufio.Reader
}

func NewRunner(store *storage.Store, in io.Reader, out, errOut io.Writer) *Runner {
	return &Runner{
		store:  store,
		in:     in,
		out:    out,
		errOut: errOut,
		reader: bufio.NewReader(in),
	}
}

func (r *Runner) Run(args []string) error {
	if len(args) == 0 {
		return r.runCode(nil)
	}

	switch args[0] {
	case "add":
		return r.runAdd(args[1:])
	case "code":
		return r.runCode(args[1:])
	case "watch":
		return r.runWatch(args[1:])
	case "list":
		return r.runList()
	case "remove":
		return r.runRemove(args[1:])
	case "help":
		return errUsage
	default:
		return errUsage
	}
}

func requireName(args []string) (string, error) {
	if len(args) != 1 {
		return "", errUsage
	}
	return args[0], nil
}

func friendlyError(err error, name string) error {
	switch {
	case errors.Is(err, storage.ErrAccountNotFound):
		return fmt.Errorf("akun %q tidak ditemukan", name)
	case errors.Is(err, storage.ErrAccountExists):
		return fmt.Errorf("akun %q sudah ada", name)
	default:
		return err
	}
}
