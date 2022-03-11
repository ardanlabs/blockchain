package selector_test

import (
	"testing"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/mempool/selector"
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

	walletTx, err := tx.Sign(pk)
	if err != nil {
		return storage.BlockTx{}, err
	}

	signedTx, err := walletTx.ToSignedTx()
	if err != nil {
		return storage.BlockTx{}, err
	}

	return storage.NewBlockTx(signedTx, gas), nil
}

func TestTipSort(t *testing.T) {
	tran := func(nonce uint, hexKey string, tip uint, ts time.Time) storage.BlockTx {
		const to = "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76"

		tx, err := sign(hexKey, storage.UserTx{Nonce: nonce, To: to, Tip: tip}, 0)
		if err != nil {
			t.Fatalf("\t%s \tShould be able to sign transaction: %s", failed, tx)
		}
		return tx
	}

	type test struct {
		name    string
		txs     []storage.BlockTx
		howMany int
		best    []storage.BlockTx
	}

	signPavel := "fae85851bdf5c9f49923722ce38f3c1defcfd3619ef5453230a58ad805499959"
	signBill := "9f332e3700d8fc2446eaf6d15034cf96e0c2745e40353deef032a5dbf1dfed93"
	signEd := "aed31b6b5a341af8f27e66fb0b7633cf20fc27049e3eb7f6f623a4655b719ebb"

	now := time.Now()
	tt := []test{
		{
			name: "one from second cycle",
			txs: []storage.BlockTx{
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
			best: []storage.BlockTx{
				tran(0, signPavel, 25, now),
				tran(1, signPavel, 75, now),
				tran(0, signBill, 10, now),
				tran(0, signEd, 5, now),
			},
		},
		{
			name: "whole two cycles",
			txs: []storage.BlockTx{
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
			best: []storage.BlockTx{
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
			txs: []storage.BlockTx{
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
			best: []storage.BlockTx{
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
			txs: []storage.BlockTx{
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
			best: []storage.BlockTx{
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
					m := make(map[storage.Account][]storage.BlockTx)
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
					for _, tx := range txs {
						gotFrom, err := tx.FromAccount()
						if err != nil {
							t.Fatalf("\t%s\tTest %d:\tShould be able to get from account: %s", failed, testID, err)
						}

						found := false
						for _, exp := range tst.best {
							expFrom, err := exp.FromAccount()
							if err != nil {
								t.Fatalf("\t%s\tTest %d:\tShould be able to get from account: %s", failed, testID, err)
							}

							if exp.Nonce == tx.Nonce && expFrom == gotFrom {
								found = true
								break
							}
						}

						if !found {
							t.Fatalf("\t%s\tTest %d:\tShould get back the right from/nonce: %s/%d", failed, testID, gotFrom, tx.Nonce)
						}
						t.Logf("\t%s\tTest %d:\tShould get back the right from/nonce: %s/%d", success, testID, gotFrom, tx.Nonce)
					}
				}

				t.Run(tst.name, f)
			}
		}
	}
}
