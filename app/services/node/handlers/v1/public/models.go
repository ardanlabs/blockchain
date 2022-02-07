package public

import "github.com/ardanlabs/blockchain/foundation/blockchain"

type balance struct {
	Address string `json:"address"`
	Balance uint   `json:"balance"`
}

type balances struct {
	LastestBlock string    `json:"lastest_block"`
	Uncommitted  int       `json:"uncommitted"`
	Balances     []balance `json:"balances"`
}

type tx struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value uint   `json:"value"`
	Tip   uint   `json:"tip"`
	Data  []byte `json:"data"`
	Gas   uint   `json:"gas"`
	Sig   string `json:"sig"`
}

type block struct {
	Header       blockchain.BlockHeader `json:"header"`
	Transactions []tx                   `json:"txs"`
}
