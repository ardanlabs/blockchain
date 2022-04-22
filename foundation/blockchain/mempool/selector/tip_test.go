package selector_test

import (
	"testing"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/mempool/selector"
	"github.com/ethereum/go-ethereum/crypto"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

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

func TestTipSort(t *testing.T) {
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
			name: "one from second cycle",
			txs: []database.BlockTx{
				tran(0, signPavel, 25, now),
				tran(1, signPavel, 75, now),
				tran(2, signPavel, 50, now),

				tran(0, signBill, 10, now),
				tran(1, signBill, 5, now),
				tran(2, signBill, 75, now),

				tran(0, signEd, 5, now),
				tran(1, signEd, 50, now),
				tran(2, signEd, 25, now),
			},
			howMany: 4,
			best: []database.BlockTx{
				tran(0, signPavel, 25, now),
				tran(1, signPavel, 75, now),
				tran(0, signBill, 10, now),
				tran(0, signEd, 5, now),
			},
		},
		{
			name: "whole two cycles",
			txs: []database.BlockTx{
				tran(0, signPavel, 25, now),
				tran(1, signPavel, 75, now),
				tran(2, signPavel, 50, now),

				tran(0, signBill, 10, now),
				tran(1, signBill, 5, now),
				tran(2, signBill, 75, now),

				tran(0, signEd, 5, now),
				tran(1, signEd, 50, now),
				tran(2, signEd, 25, now),
			},
			howMany: 6,
			best: []database.BlockTx{
				tran(0, signPavel, 25, now),
				tran(1, signPavel, 75, now),
				tran(0, signBill, 10, now),
				tran(1, signBill, 5, now),
				tran(0, signEd, 5, now),
				tran(1, signEd, 50, now),
			},
		},
		{
			name: "take all",
			txs: []database.BlockTx{
				tran(0, signPavel, 25, now),
				tran(1, signPavel, 75, now),
				tran(2, signPavel, 50, now),
				tran(0, signBill, 10, now),
				tran(1, signBill, 5, now),
				tran(2, signBill, 75, now),
				tran(0, signEd, 5, now),
				tran(1, signEd, 50, now),
				tran(2, signEd, 25, now),
			},
			howMany: 15,
			best: []database.BlockTx{
				tran(0, signPavel, 25, now),
				tran(1, signPavel, 75, now),
				tran(2, signPavel, 50, now),
				tran(0, signBill, 10, now),
				tran(1, signBill, 5, now),
				tran(2, signBill, 75, now),
				tran(0, signEd, 5, now),
				tran(1, signEd, 50, now),
				tran(2, signEd, 25, now),
			},
		},
		{
			name: "first two",
			txs: []database.BlockTx{
				tran(0, signPavel, 25, now),
				tran(1, signPavel, 75, now),
				tran(2, signPavel, 50, now),
				tran(0, signBill, 10, now),
				tran(1, signBill, 5, now),
				tran(2, signBill, 75, now),
				tran(0, signEd, 5, now),
				tran(1, signEd, 50, now),
				tran(2, signEd, 25, now),
			},
			howMany: 2,
			best: []database.BlockTx{
				tran(0, signPavel, 25, now),
				tran(0, signBill, 10, now),
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

					sort, err := selector.Retrieve(selector.StrategyTip)
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
