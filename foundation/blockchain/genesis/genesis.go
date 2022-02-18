// Package genesis maintains access to the genesis file.
package genesis

import (
	"encoding/json"
	"os"
	"time"
)

// Genesis represents the genesis file.
type Genesis struct {
	Date          time.Time       `json:"date"`
	ChainID       string          `json:"chain_id"`
	Difficulty    int             `json:"difficulty"`             // How difficult it needs to be to solve the work problem.
	TransPerBlock int             `json:"transactions_per_block"` // Number of transactions recorded in every block.
	MiningReward  uint            `json:"mining_reward"`          // Reward for mining a block.
	GasPrice      uint            `json:"gas_price"`              // Fee paid for each transaction mined into a block.
	Balances      map[string]uint `json:"balance_sheet"`
}

// =============================================================================

// Load opens and consumes the genesis file.
func Load() (Genesis, error) {
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
