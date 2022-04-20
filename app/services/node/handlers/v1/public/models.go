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
	Gas         uint               `json:"gas"`
	Sig         string             `json:"sig"`
	Hash        string             `json:"hash"`
	Proof       []string           `json:"proof"`
	ProofOrder  []int64            `json:"proof_order"`
}

type block struct {
	ParentHash   string             `json:"parent_hash"`
	MinerAccount database.AccountID `json:"miner_account"`
	Difficulty   int                `json:"difficulty"`
	Number       uint64             `json:"number"`
	TimeStamp    uint64             `json:"timestamp"`
	Nonce        uint64             `json:"nonce"`
	MerkleRoot   string             `json:"merkle_root"`
	Transactions []tx               `json:"txs"`
}
