package bargrp

import (
	"time"

	"github.com/ardanlabs/blockchain/foundation/database"
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

func toTxs(dbTxs []database.Tx) []tx {
	tsx := make([]tx, len(dbTxs))
	for i := range dbTxs {
		tsx[i] = tx(dbTxs[i])
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
	Time      time.Time `json:"time"`
}

type block struct {
	Header       blockHeader `json:"header"`
	Transactions []tx        `json:"transactions"`
}
