package bargrp

import "github.com/ardanlabs/blockchain/business/sys/database"

type newTx struct {
	From  string `json:"from" validate:"required"`
	To    string `json:"to" validate:"required"`
	Value uint   `json:"value" validate:"required"`
	Data  string `json:"data"`
}

type balance struct {
	Account string `json:"account"`
	Balance uint   `json:"balance"`
}

type balances struct {
	LastestBlock string    `json:"lastest_block"`
	Uncommitted  int       `json:"uncommitted"`
	Balances     []balance `json:"balances"`
}

type blockHeader struct {
	PrevBlock string `json:"prev_block"`
	ThisBlock string `json:"this_block"`
	Time      uint64 `json:"time"`
}

type block struct {
	Header       blockHeader   `json:"header"`
	Transactions []database.Tx `json:"transactions"`
}
