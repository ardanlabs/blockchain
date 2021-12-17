package db

import "time"

const (
	TxTypeReward = "reward"
)

// Genesis represents the genesis file.
type Genesis struct {
	Date     time.Time       `json:"date"`
	ChainID  string          `json:"chain_id"`
	Balances map[string]uint `json:"balances"`
}

// Tx represents a transaction in the database.
type Tx struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value uint   `json:"value"`
	Data  string `json:"data"`
}

func (t Tx) IsReward() bool {
	return t.Data == TxTypeReward
}
