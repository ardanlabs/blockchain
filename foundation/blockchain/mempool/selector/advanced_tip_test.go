package selector_test

import (
	"testing"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/mempool/selector"
)

var (
	signPavel = "fae85851bdf5c9f49923722ce38f3c1defcfd3619ef5453230a58ad805499959"
	signBill  = "9f332e3700d8fc2446eaf6d15034cf96e0c2745e40353deef032a5dbf1dfed93"
	signEd    = "aed31b6b5a341af8f27e66fb0b7633cf20fc27049e3eb7f6f623a4655b719ebb"
)

func TestAdvancedSort(t *testing.T) {
	tran := func(nonce uint64, hexKey string, tip uint64, ts time.Time) database.BlockTx {
		const toID = "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76"

		tx, err := sign(hexKey, database.Tx{Nonce: nonce, ToID: toID, Tip: tip})
		if err != nil {
			t.Fatalf("\t%s \tShould be able to sign transaction: %s", failed, tx)
		}
		return tx
	}

	type test struct {
		name    string
		txs     []database.BlockTx
		howMany int
		best    []database.BlockTx
	}

	now := time.Now()
	tt := []test{
		{
			name: "all from first account",
			txs: []database.BlockTx{
				tran(1, signPavel, 1, now),
				tran(2, signPavel, 2, now),
				tran(3, signPavel, 3, now),
				tran(4, signPavel, 3, now),

				tran(1, signBill, 1, now),
				tran(2, signBill, 4, now),
				tran(3, signBill, 1, now),
			},
			howMany: 4,
			best: []database.BlockTx{
				tran(1, signPavel, 1, now),
				tran(2, signPavel, 2, now),
				tran(3, signPavel, 3, now),
				tran(4, signPavel, 3, now),
			},
		},
		{
			name: "one from another account",
			txs: []database.BlockTx{
				tran(0, signPavel, 25, now),
				tran(1, signPavel, 75, now),
				tran(2, signPavel, 50, now),

				tran(0, signBill, 1, now),
				tran(1, signBill, 5, now),
				tran(2, signBill, 6, now),

				tran(0, signEd, 5, now),
				tran(1, signEd, 6, now),
				tran(2, signEd, 7, now),
			},
			howMany: 4,
			best: []database.BlockTx{
				tran(0, signPavel, 25, now),
				tran(1, signPavel, 75, now),
				tran(2, signPavel, 50, now),

				tran(0, signEd, 5, now),
			},
		},
		{
			name: "unblock big fee",
			txs: []database.BlockTx{
				tran(0, signPavel, 1, now),
				tran(1, signPavel, 1, now),
				tran(2, signPavel, 50, now),

				tran(0, signBill, 1, now),
				tran(1, signBill, 15, now),
				tran(2, signBill, 16, now),

				tran(0, signEd, 5, now),
				tran(1, signEd, 6, now),
				tran(2, signEd, 7, now),
			},
			howMany: 4,
			best: []database.BlockTx{
				tran(0, signPavel, 1, now),
				tran(1, signPavel, 1, now),
				tran(2, signPavel, 50, now),

				tran(0, signEd, 5, now),
			},
		},
	}

	t.Log("Given the need to pick best transactions from mempool.")
	{
		for testID, tst := range tt {
			t.Logf("\tTest %d:\tWhen handling a set of transaction.", testID)
			{
				f := func(t *testing.T) {
					m := make(map[database.AccountID][]database.BlockTx)
					for _, tx := range tst.txs {
						from, err := tx.FromAccount()
						if err != nil {
							t.Fatalf("\t%s\tTest %d:\tShould be able to get from account: %s", failed, testID, err)
						}

						m[from] = append(m[from], tx)
					}

					sort, err := selector.Retrieve(selector.StrategyTipAdvanced)
					if err != nil {
						t.Fatalf("\t%s\tTest %d:\tShould be able to get sort strategy function: %s", failed, testID, err)
					}

					txs := sort(m, tst.howMany)
					if len(tst.txs) > tst.howMany && len(txs) < tst.howMany {
						t.Fatalf("\t%s\tTest %d:\tShould to get %d after sort, but got %d", failed, testID, tst.howMany, len(txs))
					}
					for _, exp := range tst.best {
						expFrom, err := exp.FromAccount()
						if err != nil {
							t.Fatalf("\t%s\tTest %d:\tShould be able to get from account: %s", failed, testID, err)
						}

						found := false
						for _, tx := range txs {
							gotFrom, err := tx.FromAccount()
							if err != nil {
								t.Fatalf("\t%s\tTest %d:\tShould be able to get from account: %s", failed, testID, err)
							}

							if exp.Nonce == tx.Nonce && expFrom == gotFrom {
								found = true
								break
							}
						}

						if !found {
							t.Fatalf("\t%s\tTest %d:\tShould get back the right from/nonce: %s/%d", failed, testID, expFrom, exp.Nonce)
						}
						t.Logf("\t%s\tTest %d:\tShould get back the right from/nonce: %s/%d", success, testID, expFrom, exp.Nonce)
					}
				}

				t.Run(tst.name, f)
			}
		}
	}
}
