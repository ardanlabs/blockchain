package public

type balance struct {
	Address string `json:"address"`
	Balance uint   `json:"balance"`
}

type Balances struct {
	LastestBlock string    `json:"lastest_block"`
	Uncommitted  int       `json:"uncommitted"`
	Balances     []balance `json:"balances"`
}
