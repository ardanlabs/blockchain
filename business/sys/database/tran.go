package database

const (
	TxDataReward = "reward"
)
const (
	TxStatusAccepted = "accepted"
	TxStatusError    = "error"
	TxStatusPending  = "pending"
)

// Tx represents a transaction in the database.
type Tx struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Value      uint   `json:"value"`
	Data       string `json:"data"`
	Status     string `json:"status"`
	StatusInfo string `json:"status_info"`
}

// NewTx constructs a new Tx for use.
func NewTx(from, to string, value uint, data string) Tx {
	return Tx{
		From:   from,
		To:     to,
		Value:  value,
		Data:   data,
		Status: TxStatusPending,
	}
}
