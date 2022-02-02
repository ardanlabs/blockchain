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

type Tx struct {
	To    string `json:"to"`
	Value uint   `json:"value"`
	Tip   uint   `json:"tip"`
	Data  string `json:"data"`
}

type SignedTx struct {
	Transaction Tx     `json:"transaction"`
	Signature   []byte `json:"signature"`
}
