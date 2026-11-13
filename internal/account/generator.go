package account

import (
	"fmt"
	"time"

	"github.com/robby031/genotp-go"
)

const (
	periodSeconds = 30
	maxSecretSize = 64
)

func GenerateCode(secretB32 string) (string, error) {
	buf := make([]byte, maxSecretSize)
	defer clear(buf)

	n, err := genotp.DecodeBase32(buf, secretB32)
	if err != nil {
		return "", fmt.Errorf("decode secret: %w", err)
	}

	code, err := genotp.GenTotpDefault(buf[:n])
	if err != nil {
		return "", fmt.Errorf("generate code: %w", err)
	}
	return code, nil
}

func RemainingSeconds(now time.Time) int {
	return periodSeconds - int(now.Unix()%periodSeconds)
}
