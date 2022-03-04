package public

import "github.com/ardanlabs/blockchain/foundation/blockchain/storage"

type info struct {
	Address storage.Address `json:"address"`
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
	FromAddress storage.Address `json:"from_address"`
	FromName    string          `json:"from_name"`
	Nonce       uint            `json:"nonce"`
	To          storage.Address `json:"to"`
	Value       uint            `json:"value"`
	Tip         uint            `json:"tip"`
	Data        []byte          `json:"data"`
	TimeStamp   uint64          `json:"timestamp"`
	Gas         uint            `json:"gas"`
	Sig         string          `json:"sig"`
}

type block struct {
	ParentHash   string          `json:"parent_hash"`
	MinerAddress storage.Address `json:"miner_address"`
	Difficulty   int             `json:"difficulty"`
	Number       uint64          `json:"number"`
	TotalTip     uint            `json:"total_tip"`
	TotalGas     uint            `json:"total_gas"`
	TimeStamp    uint64          `json:"timestamp"`
	Nonce        uint64          `json:"nonce"`
	Transactions []tx            `json:"txs"`
}
