package blockchain

import (
	"encoding/json"
	"os"
	"time"
)

// Genesis represents the genesis file.
type Genesis struct {
	Date         time.Time    `json:"date"`
	ChainID      string       `json:"chain_id"`
	Difficulty   int          `json:"difficulty"`    // The number of preceding 0's needed for a hash.
	ReadyToMine  int          `json:"ready_to_mine"` // Number of transactions needed to mine a block.
	MiningReward uint         `json:"mining_reward"` // The reward for mining a block.
	GasPrice     uint         `json:"gas_price"`     // Price of Gas for a single transaction.
	Balances     BalanceSheet `json:"balance_sheet"`
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
