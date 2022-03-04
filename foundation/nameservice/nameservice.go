// Package nameservice reads the zblock/accounts folder and creates a name
// service lookup for the ardan accounts.
package nameservice

import (
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
	"github.com/ethereum/go-ethereum/crypto"
)

// NameService maintains a map of accounts for name lookup.
type NameService struct {
	accounts map[storage.Account]string
}

// New constructs an Ardan Name Service with accounts from the zblock/accounts folder.
func New(root string) (*NameService, error) {
	ns := NameService{
		accounts: make(map[storage.Account]string),
	}

	fn := func(fileName string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walkdir failure: %w", err)
		}

		if path.Ext(fileName) != ".ecdsa" {
			return nil
		}

		privateKey, err := crypto.LoadECDSA(fileName)
		if err != nil {
			return err
		}

		account := storage.PublicKeyToAccount(privateKey.PublicKey)
		ns.accounts[account] = strings.TrimSuffix(path.Base(fileName), ".ecdsa")

		return nil
	}

	if err := filepath.Walk(root, fn); err != nil {
		return nil, fmt.Errorf("walking directory: %w", err)
	}

	return &ns, nil
}

// Lookup returns the name for the specified account.
func (ns *NameService) Lookup(account storage.Account) string {
	name, exists := ns.accounts[account]
	if !exists {
		return string(account)
	}
	return name
}

// Copy returns a copy of the map of names and accounts.
func (ns *NameService) Copy() map[storage.Account]string {
	cpy := make(map[storage.Account]string, len(ns.accounts))
	for account, name := range ns.accounts {
		cpy[account] = name
	}
	return cpy
}
