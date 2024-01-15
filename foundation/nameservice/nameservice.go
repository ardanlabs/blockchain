// Package nameservice reads the zblock/accounts folder and creates a name
// service lookup for the ardan accounts.
package nameservice

import (
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ethereum/go-ethereum/crypto"
)

// NameService maintains a map of accounts for name lookup.
type NameService struct {
	accounts map[database.AccountID]string
}

// New constructs an Ardan Name Service with accounts from the zblock/accounts folder.
func New(root string) (*NameService, error) {
	ns := NameService{
		accounts: make(map[database.AccountID]string),
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

		accountID := database.PublicKeyToAccountID(privateKey.PublicKey)
		ns.accounts[accountID] = strings.TrimSuffix(filepath.Base(fileName), ".ecdsa")

		return nil
	}

	if err := filepath.Walk(root, fn); err != nil {
		return nil, fmt.Errorf("walking directory: %w", err)
	}

	return &ns, nil
}

// Lookup returns the name for the specified account.
func (ns *NameService) Lookup(accountID database.AccountID) string {
	name, exists := ns.accounts[accountID]
	if !exists {
		return string(accountID)
	}
	return name
}

// Copy returns a copy of the map of names and accounts.
func (ns *NameService) Copy() map[database.AccountID]string {
	accounts := make(map[database.AccountID]string, len(ns.accounts))
	for account, name := range ns.accounts {
		accounts[account] = name
	}
	return accounts
}
