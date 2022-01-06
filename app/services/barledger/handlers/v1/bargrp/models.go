package bargrp

import "github.com/ardanlabs/blockchain/business/sys/database"

type newTX struct {
	Status string `json:"status"`
	From   string `json:"from" validate:"required"`
	To     string `json:"to" validate:"required"`
	Value  uint   `json:"value" validate:"required"`
	Data   string `json:"data"`
}

type bals struct {
	Account string
	Balance uint
}

type header struct {
	PrevBlock string
	Time      uint64
}

type block struct {
	Header       header
	Transactions []database.Tx
}
