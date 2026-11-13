package command

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"totp-cli/internal/storage"
)

func (r *Runner) Password(confirm bool) (string, error) {
	prompt := "Master password: "
	if confirm {
		prompt = "Buat master password baru: "
	}

	password, err := r.readLine(prompt, true)
	if err != nil {
		return "", err
	}
	if password == "" {
		return "", errors.New("master password tidak boleh kosong")
	}

	if !confirm {
		return password, nil
	}
	if len(password) < storage.MinPasswordLength {
		return "", fmt.Errorf("master password minimal %d karakter", storage.MinPasswordLength)
	}

	again, err := r.readLine("Ulangi master password: ", true)
	if err != nil {
		return "", err
	}
	if again != password {
		return "", errors.New("master password tidak sama")
	}
	return password, nil
}

func (r *Runner) readLine(prompt string, hide bool) (string, error) {
	fmt.Fprint(r.errOut, prompt)

	if hide {
		if file, ok := r.in.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
			line, err := term.ReadPassword(int(file.Fd()))
			fmt.Fprintln(r.errOut)
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(string(line)), nil
		}
	}

	line, err := r.reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
