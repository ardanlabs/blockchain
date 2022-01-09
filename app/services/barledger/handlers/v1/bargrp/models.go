package bargrp

import (
	"fmt"
	"time"

	"github.com/ardanlabs/blockchain/foundation/node"
)

type newTx struct {
	From  string `json:"from" validate:"required"`
	To    string `json:"to" validate:"required"`
	Value uint   `json:"value" validate:"required"`
	Data  string `json:"data"`
}

type tx struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Value      uint   `json:"value"`
	Data       string `json:"data"`
	Status     string `json:"status"`
	StatusInfo string `json:"status_info"`
}

func toTxs(nTxs []node.Tx) []tx {
	tsx := make([]tx, len(nTxs))
	for i := range nTxs {
		tsx[i] = tx(nTxs[i])
	}
	return tsx
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
	PrevBlock string    `json:"prev_block"`
	ThisBlock string    `json:"this_block"`
	Number    uint64    `json:"number"`
	Time      time.Time `json:"time"`
}

type block struct {
	Header       blockHeader `json:"header"`
	Transactions []tx        `json:"transactions"`
}

func toBlock(nBlock node.Block) block {
	block := block{
		Header: blockHeader{
			PrevBlock: fmt.Sprintf("%x", nBlock.Header.PrevBlock),
			ThisBlock: fmt.Sprintf("%x", nBlock.Hash()),
			Number:    nBlock.Header.Number,
			Time:      time.Unix(int64(nBlock.Header.Time), 0),
		},
		Transactions: toTxs(nBlock.Transactions),
	}

	return block
}
