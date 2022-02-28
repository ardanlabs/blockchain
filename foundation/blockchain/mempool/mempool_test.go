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

func signBill(tx storage.UserTx, gas uint) (storage.BlockTx, error) {
	pk, err := crypto.HexToECDSA("9f332e3700d8fc2446eaf6d15034cf96e0c2745e40353deef032a5dbf1dfed93")
	if err != nil {
		return storage.BlockTx{}, err
	}

	signedTx, err := tx.Sign(pk)
	if err != nil {
		return storage.BlockTx{}, err
	}

	return storage.NewBlockTx(signedTx, gas), nil
}

func signPavel(tx storage.UserTx, gas uint) (storage.BlockTx, error) {
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
				{ID: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 100},
				{ID: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 75},
				{ID: 2, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 150},
				{ID: 2, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 250},
				{ID: 3, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 200},
				{ID: 3, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 75},
			},
			best: []storage.UserTx{
				{ID: 2, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 250},
				{ID: 3, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 200},
				{ID: 2, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 150},
				{ID: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 100},
				{ID: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 75},
				{ID: 3, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 75},
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

					for i, userTx := range tst.txs {
						var tx storage.BlockTx
						var err error
						if i%2 == 0 {
							tx, err = signBill(userTx, 0)
						} else {
							tx, err = signPavel(userTx, 0)
						}

						if err != nil {
							t.Fatalf("\t%s\tTest %d:\tShould be able to sign transaction.", failed, testID)
						}
						t.Logf("\t%s\tTest %d:\tShould be able to sign transaction.", success, testID)

						mp.Upsert(tx)
						t.Logf("\t%s\tTest %d:\tShould be able to add new transaction: %s", success, testID, tx.UniqueKey())
					}

					if len(mp.Copy()) != len(tst.txs) {
						t.Logf("\t%s\tTest %d:\tgot: %d", failed, testID, len(mp.Copy()))
						t.Logf("\t%s\tTest %d:\texp: %d", failed, testID, len(tst.txs))
						t.Fatalf("\t%s\tTest %d:\tShould get back the right number of transactions.", failed, testID)
					}
					t.Logf("\t%s\tTest %d:\tShould get back the right number of transactions.", success, testID)

					for i, tx := range mp.CopyBestByTip(6) {
						if tx.To != tst.best[i].To {
							t.Logf("\t%s\tTest %d:\tgot: %s, idx: %d", failed, testID, tx.To, i)
							t.Logf("\t%s\tTest %d:\texp: %s, idx: %d", failed, testID, tst.best[i].To, i)
							t.Fatalf("\t%s\tTest %d:\tShould get back the right tip/id.", failed, testID)
						}
						t.Logf("\t%s\tTest %d:\tShould get back the right tip/id: %d/%d", success, testID, tx.Tip, tx.ID)
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
