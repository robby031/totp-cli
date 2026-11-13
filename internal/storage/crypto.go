package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

const (
	vaultVersion      = 1
	kdfName           = "pbkdf2-sha256"
	cipherName        = "aes-256-gcm"
	kdfIterations     = 600_000
	saltSize          = 16
	keySize           = 32
	MinPasswordLength = 6
)

var (
	ErrInvalidPassword = errors.New("master password salah")
	errCorruptVault    = errors.New("file akun rusak atau formatnya tidak dikenali")
)

type vaultHeader struct {
	Version    int    `json:"version"`
	KDF        string `json:"kdf"`
	Iterations int    `json:"iterations"`
	Salt       string `json:"salt"`
	Cipher     string `json:"cipher"`
}

type vaultEnvelope struct {
	Header vaultHeader `json:"header"`
	Data   string      `json:"data"`
}

func newHeader(salt []byte) vaultHeader {
	return vaultHeader{
		Version:    vaultVersion,
		KDF:        kdfName,
		Iterations: kdfIterations,
		Salt:       base64.StdEncoding.EncodeToString(salt),
		Cipher:     cipherName,
	}
}

func deriveKey(password string, h vaultHeader) ([]byte, error) {
	if h.Version != vaultVersion || h.KDF != kdfName || h.Cipher != cipherName {
		return nil, errCorruptVault
	}
	salt, err := base64.StdEncoding.DecodeString(h.Salt)
	if err != nil || len(salt) < saltSize {
		return nil, errCorruptVault
	}
	if h.Iterations < 1 || h.Iterations > 10_000_000 {
		return nil, errCorruptVault
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, h.Iterations, keySize)
	if err != nil {
		return nil, fmt.Errorf("derive key: %w", err)
	}
	return key, nil
}

func sealVault(key []byte, h vaultHeader, plaintext []byte) (string, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("buat nonce: %w", err)
	}

	aad, err := json.Marshal(h)
	if err != nil {
		return "", fmt.Errorf("encode header: %w", err)
	}

	sealed := gcm.Seal(nil, nonce, plaintext, aad)
	raw := make([]byte, 0, len(nonce)+len(sealed))
	raw = append(raw, nonce...)
	raw = append(raw, sealed...)
	return base64.StdEncoding.EncodeToString(raw), nil
}

func openVault(key []byte, h vaultHeader, encoded string) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(raw) < gcm.NonceSize()+gcm.Overhead() {
		return nil, errCorruptVault
	}
	nonce, ciphertext := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]

	aad, err := json.Marshal(h)
	if err != nil {
		return nil, fmt.Errorf("encode header: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		return nil, ErrInvalidPassword
	}
	return plaintext, nil
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("buat cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("buat GCM: %w", err)
	}
	return gcm, nil
}
