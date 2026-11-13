package command

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/robby031/totp-cli/internal/storage"
	"golang.org/x/term"
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

func (r *Runner) readVisibleAndClear(prompt string) (string, error) {
	fmt.Fprint(r.errOut, prompt)

	file, ok := r.in.(*os.File)
	if !ok || !term.IsTerminal(int(file.Fd())) {
		line, err := r.reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		}
		return strings.TrimSpace(line), nil
	}

	fd := int(file.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer term.Restore(fd, oldState)

	var line []byte
	readByte := func() (byte, error) {
		var buf [1]byte
		n, err := file.Read(buf[:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, io.EOF
		}
		return buf[0], nil
	}

	for {
		b, err := readByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return strings.TrimSpace(string(line)), nil
			}
			return "", err
		}

		switch b {
		case '\r', '\n':
			fmt.Fprint(r.errOut, "\r\033[2K")
			return strings.TrimSpace(string(line)), nil
		case 0x03:
			fmt.Fprint(r.errOut, "\r\033[2K")
			return "", errors.New("input dibatalkan")
		case 0x04:
			if len(line) == 0 {
				return "", io.EOF
			}
		case 0x7f, 0x08:
			if len(line) > 0 {
				line = line[:len(line)-1]
				fmt.Fprint(r.errOut, "\b \b")
			}
		case 0x1b:
			skipEscapeSequence(readByte)
		default:
			if b < 0x20 {
				continue
			}
			line = append(line, b)
			fmt.Fprintf(r.errOut, "%c", b)
		}
	}
}

func skipEscapeSequence(readByte func() (byte, error)) {
	b, err := readByte()
	if err != nil {
		return
	}
	if b != '[' && b != 'O' {
		return
	}
	for {
		b, err = readByte()
		if err != nil {
			return
		}
		if b >= 0x40 && b <= 0x7e {
			return
		}
	}
}
