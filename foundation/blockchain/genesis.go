package blockchain

import (
	"encoding/json"
	"os"
	"time"
)

// Genesis represents the genesis file.
type Genesis struct {
	Date     time.Time    `json:"date"`
	ChainID  string       `json:"chain_id"`
	Balances BalanceSheet `json:"balance_sheet"`
}

// =============================================================================

// loadGenesis opens and consumes the genesis file.
func loadGenesis() (Genesis, error) {
	path := "zblock/genesis.json"
	content, err := os.ReadFile(path)
	if err != nil {
		return Genesis{}, err
	}

	var genesis Genesis
	err = json.Unmarshal(content, &genesis)
	if err != nil {
		return Genesis{}, err
	}

	return genesis, nil
}
