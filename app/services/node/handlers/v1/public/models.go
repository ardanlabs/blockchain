package public

type balance struct {
	Account string `json:"account"`
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
	Data  string `json:"data"`
}
