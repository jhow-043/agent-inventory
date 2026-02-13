// Package token handles device token persistence on the local filesystem.
package token

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const tokenFileName = "device.token"

// Store manages reading and writing the device authentication token.
type Store struct {
	dir string
}

// NewStore creates a new token store with the given data directory.
func NewStore(dataDir string) *Store {
	return &Store{dir: dataDir}
}

// Load reads the stored token from disk. Returns empty string if no token exists.
func (s *Store) Load() (string, error) {
	data, err := os.ReadFile(s.path())
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read token: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// Save writes the token to disk, creating the data directory if needed.
func (s *Store) Save(tok string) error {
	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}
	if err := os.WriteFile(s.path(), []byte(tok), 0600); err != nil {
		return fmt.Errorf("write token: %w", err)
	}
	return nil
}

// Delete removes the stored token file.
func (s *Store) Delete() error {
	if err := os.Remove(s.path()); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete token: %w", err)
	}
	return nil
}

func (s *Store) path() string {
	return filepath.Join(s.dir, tokenFileName)
}
