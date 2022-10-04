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

	fromPavel = "0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4"
	fromBill  = "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32"
	fromEd    = "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0"
)

func TestAdvancedSort(t *testing.T) {
	tran := func(nonce uint64, from string, hexKey string, tip uint64, ts time.Time) database.BlockTx {
		const toID = "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76"

		tx, err := sign(hexKey, database.Tx{Nonce: nonce, FromID: database.AccountID(from), ToID: toID, Tip: tip})
		if err != nil {
			t.Fatalf("Should be able to sign transaction: %s", tx)
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
				tran(1, fromPavel, signPavel, 1, now),
				tran(2, fromPavel, signPavel, 2, now),
				tran(3, fromPavel, signPavel, 3, now),
				tran(4, fromPavel, signPavel, 3, now),

				tran(1, fromBill, signBill, 1, now),
				tran(2, fromBill, signBill, 4, now),
				tran(3, fromBill, signBill, 1, now),
			},
			howMany: 4,
			best: []database.BlockTx{
				tran(1, fromPavel, signPavel, 1, now),
				tran(2, fromPavel, signPavel, 2, now),
				tran(3, fromPavel, signPavel, 3, now),
				tran(4, fromPavel, signPavel, 3, now),
			},
		},
		{
			name: "one from another account",
			txs: []database.BlockTx{
				tran(0, fromPavel, signPavel, 25, now),
				tran(1, fromPavel, signPavel, 75, now),
				tran(2, fromPavel, signPavel, 50, now),

				tran(0, fromBill, signBill, 1, now),
				tran(1, fromBill, signBill, 5, now),
				tran(2, fromBill, signBill, 6, now),

				tran(0, fromEd, signEd, 5, now),
				tran(1, fromEd, signEd, 6, now),
				tran(2, fromEd, signEd, 7, now),
			},
			howMany: 4,
			best: []database.BlockTx{
				tran(0, fromPavel, signPavel, 25, now),
				tran(1, fromPavel, signPavel, 75, now),
				tran(2, fromPavel, signPavel, 50, now),

				tran(0, fromEd, signEd, 5, now),
			},
		},
		{
			name: "unblock big fee",
			txs: []database.BlockTx{
				tran(0, fromPavel, signPavel, 1, now),
				tran(1, fromPavel, signPavel, 1, now),
				tran(2, fromPavel, signPavel, 50, now),

				tran(0, fromBill, signBill, 1, now),
				tran(1, fromBill, signBill, 15, now),
				tran(2, fromBill, signBill, 16, now),

				tran(0, fromEd, signEd, 5, now),
				tran(1, fromEd, signEd, 6, now),
				tran(2, fromEd, signEd, 7, now),
			},
			howMany: 4,
			best: []database.BlockTx{
				tran(0, fromPavel, signPavel, 1, now),
				tran(1, fromPavel, signPavel, 1, now),
				tran(2, fromPavel, signPavel, 50, now),

				tran(0, fromEd, signEd, 5, now),
			},
		},
	}

	for _, tst := range tt {
		f := func(t *testing.T) {
			m := make(map[database.AccountID][]database.BlockTx)
			for _, tx := range tst.txs {
				m[tx.FromID] = append(m[tx.FromID], tx)
			}

			sort, err := selector.Retrieve(selector.StrategyTipAdvanced)
			if err != nil {
				t.Fatalf("Test %s:\tShould be able to get sort strategy function: %s", tst.name, err)
			}

			txs := sort(m, tst.howMany)
			if len(tst.txs) > tst.howMany && len(txs) < tst.howMany {
				t.Fatalf("Test %s:\tShould to get %d after sort, but got %d", tst.name, tst.howMany, len(txs))
			}
			for _, exp := range tst.best {
				expFrom := exp.FromID
				found := false
				for _, tx := range txs {
					if exp.Nonce == tx.Nonce && expFrom == tx.FromID {
						found = true
						break
					}
				}

				if !found {
					t.Fatalf("Test %s:\tShould get back the right from/nonce: %s/%d", tst.name, expFrom, exp.Nonce)
				}
			}
		}

		t.Run(tst.name, f)
	}
}
