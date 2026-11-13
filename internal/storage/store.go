package storage

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/robby031/totp-cli/internal/account"
)

var (
	ErrAccountExists   = errors.New("akun sudah ada")
	ErrAccountNotFound = errors.New("akun tidak ditemukan")
)

type PasswordProvider func(confirm bool) (string, error)

type Store struct {
	path     string
	password PasswordProvider
	key      []byte
	header   vaultHeader
}

func NewDefaultStore() (*Store, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("resolve config dir: %w", err)
	}
	return &Store{path: filepath.Join(dir, "totp-cli", "accounts.json")}, nil
}

func (s *Store) SetPasswordProvider(p PasswordProvider) {
	s.password = p
}

func (s *Store) List() ([]account.Account, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read accounts: %w", err)
	}

	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, nil
	}

	var envelope vaultEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("%w: %v", errCorruptVault, err)
	}

	plaintext, err := s.unlock(&envelope)
	if err != nil {
		return nil, err
	}
	return decodeAccounts(plaintext)
}

func (s *Store) Find(name string) (account.Account, error) {
	accounts, err := s.List()
	if err != nil {
		return account.Account{}, err
	}
	key := account.NormalizeName(name)
	for _, acc := range accounts {
		if acc.Name == key {
			return acc, nil
		}
	}
	return account.Account{}, ErrAccountNotFound
}

func (s *Store) Add(acc account.Account) error {
	accounts, err := s.List()
	if err != nil {
		return err
	}
	for _, existing := range accounts {
		if existing.Name == acc.Name {
			return ErrAccountExists
		}
	}
	return s.save(append(accounts, acc))
}

func (s *Store) Remove(name string) error {
	accounts, err := s.List()
	if err != nil {
		return err
	}
	key := account.NormalizeName(name)
	remaining := make([]account.Account, 0, len(accounts))
	for _, acc := range accounts {
		if acc.Name != key {
			remaining = append(remaining, acc)
		}
	}
	if len(remaining) == len(accounts) {
		return ErrAccountNotFound
	}
	return s.save(remaining)
}

func decodeAccounts(data []byte) ([]account.Account, error) {
	var accounts []account.Account
	if err := json.Unmarshal(data, &accounts); err != nil {
		return nil, fmt.Errorf("parse accounts: %w", err)
	}
	sort.Slice(accounts, func(i, j int) bool { return accounts[i].Name < accounts[j].Name })
	return accounts, nil
}

func (s *Store) unlock(envelope *vaultEnvelope) ([]byte, error) {
	if s.key != nil && s.header == envelope.Header {
		return openVault(s.key, envelope.Header, envelope.Data)
	}

	if s.password == nil {
		return nil, errors.New("master password belum disiapkan")
	}
	password, err := s.password(false)
	if err != nil {
		return nil, err
	}
	if password == "" {
		return nil, errors.New("master password tidak boleh kosong")
	}

	key, err := deriveKey(password, envelope.Header)
	if err != nil {
		return nil, err
	}
	plaintext, err := openVault(key, envelope.Header, envelope.Data)
	if err != nil {
		return nil, err
	}

	s.key = key
	s.header = envelope.Header
	return plaintext, nil
}

func (s *Store) generateKey() error {
	if s.password == nil {
		return errors.New("master password belum disiapkan")
	}

	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("buat salt: %w", err)
	}
	header := newHeader(salt)

	password, err := s.password(true)
	if err != nil {
		return err
	}
	if len(password) < MinPasswordLength {
		return fmt.Errorf("master password minimal %d karakter", MinPasswordLength)
	}

	key, err := deriveKey(password, header)
	if err != nil {
		return err
	}
	s.key = key
	s.header = header
	return nil
}

func (s *Store) save(accounts []account.Account) error {
	if s.key == nil {
		if err := s.generateKey(); err != nil {
			return err
		}
	}

	sort.Slice(accounts, func(i, j int) bool { return accounts[i].Name < accounts[j].Name })

	plaintext, err := json.Marshal(accounts)
	if err != nil {
		return fmt.Errorf("encode accounts: %w", err)
	}

	sealed, err := sealVault(s.key, s.header, plaintext)
	if err != nil {
		return fmt.Errorf("encrypt accounts: %w", err)
	}
	envelope := vaultEnvelope{Header: s.header, Data: sealed}

	data, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return fmt.Errorf("encode vault: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return fmt.Errorf("write accounts: %w", err)
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		return fmt.Errorf("replace accounts file: %w", err)
	}
	return nil
}
