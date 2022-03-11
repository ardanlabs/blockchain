package public

import (
	"encoding/hex"
	"math/big"

	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

type info struct {
	Account storage.Account `json:"account"`
	Name    string          `json:"name"`
	Balance uint            `json:"balance"`
	Nonce   uint            `json:"nonce"`
}

type actInfo struct {
	LastestBlock string `json:"lastest_block"`
	Uncommitted  int    `json:"uncommitted"`
	Accounts     []info `json:"accounts"`
}

type tx struct {
	FromAccount storage.Account `json:"from_account"`
	FromName    string          `json:"from_name"`
	Nonce       uint            `json:"nonce"`
	To          storage.Account `json:"to"`
	Value       uint            `json:"value"`
	Tip         uint            `json:"tip"`
	Data        []byte          `json:"data"`
	TimeStamp   uint64          `json:"timestamp"`
	Gas         uint            `json:"gas"`
	Sig         string          `json:"sig"`
}

type block struct {
	ParentHash   string          `json:"parent_hash"`
	MinerAccount storage.Account `json:"miner_account"`
	Difficulty   int             `json:"difficulty"`
	Number       uint64          `json:"number"`
	TotalTip     uint            `json:"total_tip"`
	TotalGas     uint            `json:"total_gas"`
	TimeStamp    uint64          `json:"timestamp"`
	Nonce        uint64          `json:"nonce"`
	Transactions []tx            `json:"txs"`
}

type walletTx struct {
	Nonce uint   `json:"nonce"` // Unique id for the transaction supplied by the user.
	To    string `json:"to"`    // Account receiving the benefit of the transaction.
	Value uint   `json:"value"` // Monetary value received from this transaction.
	Tip   uint   `json:"tip"`   // Tip offered by the sender as an incentive to mine this transaction.
	Data  []byte `json:"data"`  // Extra data related to the transaction.
	Sig   string `json:"sig"`   // Raw signature of the account who signed the transaction.
}

func (w walletTx) toSignedTx() (storage.SignedTx, error) {
	to, err := storage.ToAccount(w.To)
	if err != nil {
		return storage.SignedTx{}, err
	}

	sig, err := hex.DecodeString(w.Sig[2:])
	if err != nil {
		return storage.SignedTx{}, err
	}

	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:64])
	v := new(big.Int).SetBytes([]byte{sig[64]})

	signedTx := storage.SignedTx{
		UserTx: storage.UserTx{
			Nonce: w.Nonce,
			To:    to,
			Value: w.Value,
			Tip:   w.Tip,
			Data:  w.Data,
		},
		R: r,
		S: s,
		V: v,
	}

	return signedTx, nil
}
