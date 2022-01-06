package bargrp

import "github.com/ardanlabs/blockchain/business/sys/database"

type header struct {
	PrevBlock string
	Time      uint64
}

type block struct {
	Header       header
	Transactions []database.Tx
}
