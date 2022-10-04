package mempool_test

import (
	"testing"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/mempool"
	"github.com/ethereum/go-ethereum/crypto"
)

func Test_CRUD(t *testing.T) {
	type user struct {
		Tx     database.Tx
		hexKey string
	}
	type table struct {
		name string
		txs  []user
		best map[database.AccountID]database.Tx
	}

	tt := []table{
		{
			name: "tip",
			txs: []user{
				{
					Tx:     database.Tx{Nonce: 2, FromID: "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32", ToID: "0x0000000000000000000000000000000000000000", Tip: 250},
					hexKey: "9f332e3700d8fc2446eaf6d15034cf96e0c2745e40353deef032a5dbf1dfed93",
				},
				{
					Tx:     database.Tx{Nonce: 2, FromID: "0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4", ToID: "0x1111111111111111111111111111111111111111", Tip: 200},
					hexKey: "fae85851bdf5c9f49923722ce38f3c1defcfd3619ef5453230a58ad805499959",
				},
				{
					Tx:     database.Tx{Nonce: 2, FromID: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", ToID: "0x2222222222222222222222222222222222222222", Tip: 75},
					hexKey: "aed31b6b5a341af8f27e66fb0b7633cf20fc27049e3eb7f6f623a4655b719ebb",
				},
				{
					Tx:     database.Tx{Nonce: 1, FromID: "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32", ToID: "0x3333333333333333333333333333333333333333", Tip: 150},
					hexKey: "9f332e3700d8fc2446eaf6d15034cf96e0c2745e40353deef032a5dbf1dfed93",
				},
				{
					Tx:     database.Tx{Nonce: 1, FromID: "0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4", ToID: "0x4444444444444444444444444444444444444444", Tip: 75},
					hexKey: "fae85851bdf5c9f49923722ce38f3c1defcfd3619ef5453230a58ad805499959",
				},
				{
					Tx:     database.Tx{Nonce: 1, FromID: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", ToID: "0x5555555555555555555555555555555555555555", Tip: 100},
					hexKey: "aed31b6b5a341af8f27e66fb0b7633cf20fc27049e3eb7f6f623a4655b719ebb",
				},
			},
			best: map[database.AccountID]database.Tx{
				"0x3333333333333333333333333333333333333333": {Nonce: 1, Tip: 150},
				"0x5555555555555555555555555555555555555555": {Nonce: 1, Tip: 100},
				"0x4444444444444444444444444444444444444444": {Nonce: 1, Tip: 75},
				"0x0000000000000000000000000000000000000000": {Nonce: 2, Tip: 250},
			},
		},
	}

	for _, tst := range tt {
		f := func(t *testing.T) {
			mp, err := mempool.New()
			if err != nil {
				t.Fatalf("Test %s:\tShould be able to construct a mempool: %s", tst.name, err)
			}

			for _, user := range tst.txs {
				tx, err := sign(user.hexKey, user.Tx)
				if err != nil {
					t.Fatalf("Test %s:\tShould be able to sign/upsert transaction: %s", tst.name, tx)
				}

				mp.Upsert(tx)
			}

			txs := mp.PickBest(4)
			if len(txs) != 4 {
				t.Fatalf("Test %s:\tShould get back the 4 transactions.", tst.name)
			}

			for _, tx := range txs {
				if _, exists := tst.best[tx.ToID]; !exists {
					t.Fatalf("Test %s:\tShould get back the right account/tip: %s/%d", tst.name, tx.ToID, tx.Tip)
				}
			}

			mp.Delete(txs[1])
			txs = mp.PickBest()
			if len(txs) != len(tst.txs)-1 {
				t.Logf("Test %s:\tgot: %d", tst.name, len(txs))
				t.Logf("Test %s:\texp: %d", tst.name, len(tst.txs)-1)
				t.Fatalf("Test %s:\tShould be able to remove a transaction.", tst.name)
			}

			mp.Truncate()
			txs = mp.PickBest()
			if len(txs) != 0 {
				t.Fatalf("Test %s:\tShould be able to truncate mempool.", tst.name)
			}
		}

		t.Run(tst.name, f)
	}
}

// =============================================================================

func sign(hexKey string, tx database.Tx) (database.BlockTx, error) {
	pk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		return database.BlockTx{}, err
	}

	signedTx, err := tx.Sign(pk)
	if err != nil {
		return database.BlockTx{}, err
	}

	return database.NewBlockTx(signedTx, 0, 0), nil
}
