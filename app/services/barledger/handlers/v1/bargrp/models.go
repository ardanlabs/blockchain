package bargrp

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
