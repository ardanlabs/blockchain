package public

type balance struct {
	Account string `json:"account"`
	Balance uint   `json:"balance"`
}

type Balances struct {
	LastestBlock string    `json:"lastest_block"`
	Uncommitted  int       `json:"uncommitted"`
	Balances     []balance `json:"balances"`
}

type Tx struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value uint   `json:"value"`
	Tip   uint   `json:"tip"`
	Data  string `json:"data"`
}

type SignedTx struct {
	Transaction Tx     `json:"transaction"`
	Signature   []byte `json:"signature"`
}
