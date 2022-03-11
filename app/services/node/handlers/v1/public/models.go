package public

import (
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

type info struct {
	Account storage.Account `json:"account"`
	Name    string          `json:"name"`
	Balance uint            `json:"balance"`
	Nonce   uint            `json:"nonce"`
}

type actInfo struct {
	LastestBlock string `json:"lastest_block"`
	Uncommitted  int    `json:"uncommitted"`
	Accounts     []info `json:"accounts"`
}

type tx struct {
	FromAccount storage.Account `json:"from_account"`
	FromName    string          `json:"from_name"`
	Nonce       uint            `json:"nonce"`
	To          storage.Account `json:"to"`
	Value       uint            `json:"value"`
	Tip         uint            `json:"tip"`
	Data        []byte          `json:"data"`
	TimeStamp   uint64          `json:"timestamp"`
	Gas         uint            `json:"gas"`
	Sig         string          `json:"sig"`
}

type block struct {
	ParentHash   string          `json:"parent_hash"`
	MinerAccount storage.Account `json:"miner_account"`
	Difficulty   int             `json:"difficulty"`
	Number       uint64          `json:"number"`
	TotalTip     uint            `json:"total_tip"`
	TotalGas     uint            `json:"total_gas"`
	TimeStamp    uint64          `json:"timestamp"`
	Nonce        uint64          `json:"nonce"`
	Transactions []tx            `json:"txs"`
}
