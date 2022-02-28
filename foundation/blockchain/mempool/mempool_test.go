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

func sign(tx storage.UserTx, gas uint) (storage.BlockTx, error) {
	pk, err := crypto.HexToECDSA("fae85851bdf5c9f49923722ce38f3c1defcfd3619ef5453230a58ad805499959")
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
	type table struct {
		name string
		txs  []storage.UserTx
		best []storage.UserTx
	}

	tt := []table{
		{
			name: "basic",
			txs: []storage.UserTx{
				{ID: 2, To: "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32", Tip: 10},
				{ID: 3, To: "0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4", Tip: 50},
				{ID: 4, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 100},
				{ID: 1, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 10},
			},
			best: []storage.UserTx{
				{ID: 4, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 100},
				{ID: 3, To: "0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4", Tip: 50},
				{ID: 2, To: "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32", Tip: 10},
				{ID: 1, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 10},
			},
		},
	}

	t.Log("Given the need to validate mempool api.")
	{
		for testID, tst := range tt {
			t.Logf("\tTest %d:\tWhen handling a set of transaction.", testID)
			{
				f := func(t *testing.T) {
					mp := mempool.New()

					for _, userTx := range tst.txs {
						tx, err := sign(userTx, 0)
						if err != nil {
							t.Fatalf("\t%s\tTest %d:\tShould be able to sign transaction.", failed, testID)
						}
						t.Logf("\t%s\tTest %d:\tShould be able to sign transaction.", success, testID)

						mp.Upsert(tx)
						t.Logf("\t%s\tTest %d:\tShould be able to add new transaction: %s", success, testID, tx.UniqueKey())
					}

					for i, tx := range mp.Copy() {
						if tx.To != tst.txs[i].To {
							t.Logf("\t%s\tTest %d:\tgot: %s", failed, testID, tx.To)
							t.Logf("\t%s\tTest %d:\texp: %s", failed, testID, tst.txs[i].To)
							t.Fatalf("\t%s\tTest %d:\tShould get back the right account.", failed, testID)
						}
						t.Logf("\t%s\tTest %d:\tShould get back the right account: %s", success, testID, tx.To[:6])
					}

					for i, tx := range mp.CopyBestByTip(4) {
						if tx.To != tst.best[i].To {
							t.Logf("\t%s\tTest %d:\tgot: %s", failed, testID, tx.To)
							t.Logf("\t%s\tTest %d:\texp: %s", failed, testID, tst.best[i].To)
							t.Fatalf("\t%s\tTest %d:\tShould get back the right tip.", failed, testID)
						}
						t.Logf("\t%s\tTest %d:\tShould get back the right tip: %d", success, testID, tx.Tip)
					}

					mp.Delete(mp.Copy()[1])
					l := len(mp.Copy())
					if l != 3 {
						t.Fatalf("\t%s\tTest %d:\tShould be able to remove a transaction.", failed, testID)
					}
					t.Logf("\t%s\tTest %d:\tShould be able to remove a transaction.", success, testID)

					mp.Truncate()
					l = len(mp.Copy())
					if l != 0 {
						t.Fatalf("\t%s\tTest %d:\tShould be able to truncate mempool.", failed, testID)
					}
					t.Logf("\t%s\tTest %d:\tShould be able to truncate mempool.", success, testID)
				}

				t.Run(tst.name, f)
			}
		}
	}
}
