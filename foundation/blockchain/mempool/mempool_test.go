package mempool_test

import (
	"testing"

	"github.com/ardanlabs/blockchain/foundation/blockchain/mempool"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
	"github.com/ethereum/go-ethereum/crypto"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

func sign(hexKey string, tx storage.UserTx, gas uint) (storage.BlockTx, error) {
	pk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		return storage.BlockTx{}, err
	}

	signedTx, err := tx.Sign(pk)
	if err != nil {
		return storage.BlockTx{}, err
	}

	return storage.NewBlockTx(signedTx, gas), nil
}

func TestCRUD(t *testing.T) {
	type user struct {
		userTx storage.UserTx
		hexKey string
	}
	type table struct {
		name string
		txs  []user
		best map[string]storage.UserTx
	}

	tt := []table{
		{
			name: "tip",
			txs: []user{
				{
					userTx: storage.UserTx{Nonce: 2, To: "0x0000000000000000000000000000000000000000", Tip: 250},
					hexKey: "9f332e3700d8fc2446eaf6d15034cf96e0c2745e40353deef032a5dbf1dfed93",
				},
				{
					userTx: storage.UserTx{Nonce: 2, To: "0x1111111111111111111111111111111111111111", Tip: 200},
					hexKey: "fae85851bdf5c9f49923722ce38f3c1defcfd3619ef5453230a58ad805499959",
				},
				{
					userTx: storage.UserTx{Nonce: 2, To: "0x2222222222222222222222222222222222222222", Tip: 75},
					hexKey: "aed31b6b5a341af8f27e66fb0b7633cf20fc27049e3eb7f6f623a4655b719ebb",
				},
				{
					userTx: storage.UserTx{Nonce: 1, To: "0x3333333333333333333333333333333333333333", Tip: 150},
					hexKey: "9f332e3700d8fc2446eaf6d15034cf96e0c2745e40353deef032a5dbf1dfed93",
				},
				{
					userTx: storage.UserTx{Nonce: 1, To: "0x4444444444444444444444444444444444444444", Tip: 75},
					hexKey: "fae85851bdf5c9f49923722ce38f3c1defcfd3619ef5453230a58ad805499959",
				},
				{
					userTx: storage.UserTx{Nonce: 1, To: "0x5555555555555555555555555555555555555555", Tip: 100},
					hexKey: "aed31b6b5a341af8f27e66fb0b7633cf20fc27049e3eb7f6f623a4655b719ebb",
				},
			},
			best: map[string]storage.UserTx{
				"0x3333333333333333333333333333333333333333": {Nonce: 1, Tip: 150},
				"0x5555555555555555555555555555555555555555": {Nonce: 1, Tip: 100},
				"0x4444444444444444444444444444444444444444": {Nonce: 1, Tip: 75},
				"0x0000000000000000000000000000000000000000": {Nonce: 2, Tip: 250},
			},
		},
	}

	t.Log("Given the need to validate mempool api.")
	{
		for testID, tst := range tt {
			t.Logf("\tTest %d:\tWhen handling a set of transaction.", testID)
			{
				f := func(t *testing.T) {
					mp, err := mempool.New()
					if err != nil {
						t.Fatalf("\t%s\tTest %d:\tShould be able to construct a mempool: %s", failed, testID, err)
					}

					for _, user := range tst.txs {
						tx, err := sign(user.hexKey, user.userTx, 0)
						if err != nil {
							t.Fatalf("\t%s\tTest %d:\tShould be able to sign/upsert transaction: %s", failed, testID, tx)
						}
						t.Logf("\t%s\tTest %d:\tShould be able to sign/upsert transaction: %s", success, testID, tx)

						mp.Upsert(tx)
					}

					txs := mp.PickBest(4)
					if len(txs) != 4 {
						t.Fatalf("\t%s\tTest %d:\tShould get back the 4 transactions.", failed, testID)
					}
					t.Logf("\t%s\tTest %d:\tShould get back the 4 transactions", success, testID)

					for _, tx := range txs {
						if _, exists := tst.best[tx.To]; !exists {
							t.Fatalf("\t%s\tTest %d:\tShould get back the right addr/tip: %s/%d", failed, testID, tx.To, tx.Tip)
						}
						t.Logf("\t%s\tTest %d:\tShould get back the right addr/tip: %s/%d", success, testID, tx.To, tx.Tip)
					}

					mp.Delete(txs[1])
					txs = mp.PickBest(-1)
					if len(txs) != len(tst.txs)-1 {
						t.Logf("\t%s\tTest %d:\tgot: %d", failed, testID, len(txs))
						t.Logf("\t%s\tTest %d:\texp: %d", failed, testID, len(tst.txs)-1)
						t.Fatalf("\t%s\tTest %d:\tShould be able to remove a transaction.", failed, testID)
					}
					t.Logf("\t%s\tTest %d:\tShould be able to remove a transaction.", success, testID)

					mp.Truncate()
					txs = mp.PickBest(-1)
					if len(txs) != 0 {
						t.Fatalf("\t%s\tTest %d:\tShould be able to truncate mempool.", failed, testID)
					}
					t.Logf("\t%s\tTest %d:\tShould be able to truncate mempool.", success, testID)
				}

				t.Run(tst.name, f)
			}
		}
	}
}
