package public

import (
	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
)

type act struct {
	Account database.AccountID `json:"account"`
	Name    string             `json:"name"`
	Balance uint               `json:"balance"`
	Nonce   uint               `json:"nonce"`
}

type actInfo struct {
	LastestBlock string `json:"lastest_block"`
	Uncommitted  int    `json:"uncommitted"`
	Accounts     []act  `json:"accounts"`
}

type tx struct {
	FromAccount database.AccountID `json:"from"`
	FromName    string             `json:"from_name"`
	To          database.AccountID `json:"to"`
	ToName      string             `json:"to_name"`
	Nonce       uint               `json:"nonce"`
	Value       uint               `json:"value"`
	Tip         uint               `json:"tip"`
	Data        []byte             `json:"data"`
	TimeStamp   uint64             `json:"timestamp"`
	GasPrice    uint               `json:"gas_price"`
	GasUnits    uint               `json:"gas_units"`
	Sig         string             `json:"sig"`
	Proof       []string           `json:"proof"`
	ProofOrder  []int64            `json:"proof_order"`
}

type block struct {
	PrevBlockHash string             `json:"prev_block_hash"`
	BeneficiaryID database.AccountID `json:"beneficiary"`
	Difficulty    uint               `json:"difficulty"`
	Number        uint64             `json:"number"`
	TimeStamp     uint64             `json:"timestamp"`
	Nonce         uint64             `json:"nonce"`
	TransRoot     string             `json:"trans_root"`
	Transactions  []tx               `json:"txs"`
}
