package account

import (
	"errors"
	"strings"
)

type Account struct {
	Name   string `json:"name"`
	Secret string `json:"secret"`
}

var ErrEmptyName = errors.New("nama akun harus diisi")

func NormalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func NormalizeSecret(secret string) string {
	return strings.ToUpper(strings.Join(strings.Fields(secret), ""))
}
