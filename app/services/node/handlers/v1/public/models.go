package public

import "github.com/ardanlabs/blockchain/foundation/node"

type balance struct {
	Account node.Account `json:"account"`
	Balance uint         `json:"balance"`
}

type balances struct {
	LastestBlock node.Hash `json:"lastest_block"`
	Uncommitted  int       `json:"uncommitted"`
	Balances     []balance `json:"balances"`
}

type tx struct {
	From  node.Account `json:"from"`
	To    node.Account `json:"to"`
	Value uint         `json:"value"`
	Data  string       `json:"data"`
}
