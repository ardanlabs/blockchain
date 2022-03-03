package mempool_test

import (
	"testing"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/mempool"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

func TestSign(t *testing.T) {
	tx1, err := sign("9f332e3700d8fc2446eaf6d15034cf96e0c2745e40353deef032a5dbf1dfed93", storage.UserTx{Nonce: 2, To: to, Tip: 50}, 0)
	if err != nil {
		t.Fatalf("\t%s \tShould be able to sign transaction: %s %+v", failed, tx1, err)
	}
	tx2, err := sign("9f332e3700d8fc2446eaf6d15034cf96e0c2745e40353deef032a5dbf1dfed93", storage.UserTx{Nonce: 2, To: to, Tip: 75}, 0)
	if err != nil {
		t.Fatalf("\t%s \tShould be able to sign transaction: %s", failed, tx1)
	}
	from1, err := tx1.FromAddress()
	if err != nil {
		t.Fatalf("\t%s \tShould be able to get from address: %s %+v", failed, tx1, err)
	}
	from2, err := tx2.FromAddress()
	if err != nil {
		t.Fatalf("\t%s \tShould be able to get from address: %s %v", failed, tx2, err)
	}
	if from1 != from2 {
		t.Fatalf("\t%s \tfrom addresses not equal: %s vs %s", failed, from1, from2)
	}
}

func TestSimpleSort(t *testing.T) {
	tran := func(nonce uint, hexKey string, tip uint, ts time.Time) storage.BlockTx {

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
	signPavel := "9f332e3700d8fc2446eaf6d15034cf96e0c2745e40353deef032a5dbf1dfed93"
	signBill := "fae85851bdf5c9f49923722ce38f3c1defcfd3619ef5453230a58ad805499959"
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
				// tran(2, signPavel, 50, now),

				tran(0, signBill, 10, now),
				// tran(1, signBill, 5, now),
				// tran(2, signBill, 75, now),

				tran(0, signEd, 5, now),
				// tran(1, signEd, 50, now),
				// tran(2, signEd, 25, now),
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
				// tran(2, signPavel, 50, now),

				tran(0, signBill, 10, now),
				tran(1, signBill, 5, now),
				// tran(2, signBill, 75, now),

				tran(0, signEd, 5, now),
				tran(1, signEd, 50, now),
				// tran(2, signEd, 25, now),
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
				// tran(1, signPavel, 75, now),
				// tran(2, signPavel, 50, now),

				tran(0, signBill, 10, now),
				// tran(1, signBill, 5, now),
				// tran(2, signBill, 75, now),

				// tran(0, signEd, 5, now),
				// tran(1, signEd, 50, now),
				// tran(2, signEd, 25, now),
			},
		},
	}
	t.Log("Given the need to pick best transactions from mempool .")
	for testID, tst := range tt {
		t.Logf("\tTest %d:\tWhen handling a set of transaction.", testID)
		f := func(t *testing.T) {
			m := make(map[string][]storage.BlockTx)
			for _, tx := range tst.txs {
				from, err := tx.FromAddress()
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to get from address: %s", failed, testID, tx)
				}
				m[from] = append(m[from], tx)
			}

			txs := mempool.SimpleSort(m, tst.howMany)
			for _, tx := range txs {
				fromAddress, err := tx.FromAddress()
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to get from address: %s", failed, testID, tx)
				}
				found := false
				for _, exp := range tst.best {
					expFrom, err := exp.FromAddress()
					if err != nil {
						t.Fatalf("\t%s\tTest %d:\tShould be able to get from address: %s", failed, testID, tx)
					}
					if exp.Nonce == tx.Nonce && expFrom == fromAddress {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("\t%s\tTest %d:\tShould get back the right from/nonce: %s/%d", failed, testID, fromAddress, tx.Nonce)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the right from/nonce: %s/%d", success, testID, fromAddress, tx.Nonce)
			}

		}
		t.Run(tst.name, f)
	}
}
