package public

type balance struct {
	Address string `json:"address"`
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
	Tip   uint   `json:"tip"`
	Data  []byte `json:"data"`
	Gas   uint   `json:"gas"`
	Sig   string `json:"sig"`
}

type block struct {
	ParentHash   string `json:"parent_hash"`
	Beneficiary  string `json:"beneficiary"`
	Difficulty   int    `json:"difficulty"`
	Number       uint64 `json:"number"`
	TotalTip     uint   `json:"total_tip"`
	TotalGas     uint   `json:"total_gas"`
	TimeStamp    uint64 `json:"timestamp"`
	Nonce        uint64 `json:"nonce"`
	Sig          string `json:"sig"`
	Transactions []tx   `json:"txs"`
}
