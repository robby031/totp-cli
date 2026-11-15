package command

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/robby031/totp-cli/internal/account"
)

var (
	watchTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))

	watchSubStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Italic(true)

	watchCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("244")).
			Padding(0, 1)

	watchHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Italic(true)

	normalValueStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#179299"))

	warnValueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#dc8a78"))

	criticalValueStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#fe640b"))
)

func styleForRemaining(remaining int) lipgloss.Style {
	switch {
	case remaining <= 5:
		return criticalValueStyle
	case remaining <= 10:
		return warnValueStyle
	default:
		return normalValueStyle
	}
}

func writeCodes(w io.Writer, accounts []account.Account, now time.Time) error {
	remaining := account.RemainingSeconds(now)
	valueStyle := styleForRemaining(remaining)

	for _, acc := range accounts {
		code, err := account.GenerateCode(acc.Secret)
		if err != nil {
			return fmt.Errorf("akun %q: %w", acc.Name, err)
		}

		styledCode := valueStyle.Render(code)
		styledCountdown := valueStyle.Render(fmt.Sprintf("[%2ds]", remaining))

		fmt.Fprintf(w, "%-20s %s  %s\n", acc.Name, styledCode, styledCountdown)
	}
	return nil
}

func (r *Runner) renderFrame(accounts []account.Account, now time.Time) error {
	var codesBuf bytes.Buffer
	if err := writeCodes(&codesBuf, accounts, now); err != nil {
		return err
	}

	header := watchTitleStyle.Render("TOTP WATCH")
	sub := watchSubStyle.Render(now.Format("15:04:05"))
	card := watchCardStyle.Render(strings.TrimRight(codesBuf.String(), "\n"))
	hint := watchHintStyle.Render("Ctrl+C untuk keluar")

	var frame bytes.Buffer
	frame.WriteString(clearScreen)
	frame.WriteString(header)
	frame.WriteString("\n")
	frame.WriteString(sub)
	frame.WriteString("\n")
	frame.WriteString(card)
	frame.WriteString("\n")
	frame.WriteString(hint)
	frame.WriteString("\n")

	_, err := r.out.Write(frame.Bytes())
	return err
}
