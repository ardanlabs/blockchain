package mempool_test

import (
	"testing"

	"github.com/ardanlabs/blockchain/foundation/blockchain/mempool"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

func TestCRUD(t *testing.T) {
	type table struct {
		name string
		txs  []storage.BlockTx
		best []uint
	}

	tt := []table{
		{
			name: "basic",
			txs: []storage.BlockTx{
				{SignedTx: storage.SignedTx{UserTx: storage.UserTx{To: "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32", Tip: 10}}},
				{SignedTx: storage.SignedTx{UserTx: storage.UserTx{To: "0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4", Tip: 50}}},
				{SignedTx: storage.SignedTx{UserTx: storage.UserTx{To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 100}}},
				{SignedTx: storage.SignedTx{UserTx: storage.UserTx{To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 30}}},
			},
			best: []uint{100, 50, 30, 10},
		},
	}

	t.Log("Given the need to validate mempool api.")
	{
		for testID, tst := range tt {
			t.Logf("\tTest %d:\tWhen handling a set of transaction.", testID)
			{
				f := func(t *testing.T) {
					mp := mempool.New()

					for _, tx := range tst.txs {
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
						if tx.Tip != tst.best[i] {
							t.Logf("\t%s\tTest %d:\tgot: %d", failed, testID, tx.Tip)
							t.Logf("\t%s\tTest %d:\texp: %d", failed, testID, tst.best[i])
							t.Fatalf("\t%s\tTest %d:\tShould get back the right tip.", failed, testID)
						}
						t.Logf("\t%s\tTest %d:\tShould get back the right tip: %d", success, testID, tx.Tip)
					}

					mp.Delete(tst.txs[2])
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
