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

func signEd(tx storage.UserTx, gas uint) (storage.BlockTx, error) {
	pk, err := crypto.HexToECDSA("aed31b6b5a341af8f27e66fb0b7633cf20fc27049e3eb7f6f623a4655b719ebb")
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
			name: "tip",
			txs: []storage.UserTx{
				{Nonce: 2, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 250},
				{Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 200},
				{Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 75},
				{Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 150},
				{Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 75},
				{Nonce: 1, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 100},
			},
			best: []storage.UserTx{
				{Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 150},
				{Nonce: 1, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 100},
				{Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 75},
				{Nonce: 2, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 250},
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

					sign := []func(tx storage.UserTx, gas uint) (storage.BlockTx, error){
						signBill,
						signPavel,
						signEd,
					}

					var signIdx int
					for _, userTx := range tst.txs {
						var tx storage.BlockTx
						var err error

						tx, err = sign[signIdx](userTx, 0)
						signIdx++
						if signIdx == 3 {
							signIdx = 0
						}

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

					for i, tx := range txs {
						if tx.To != tst.best[i].To {
							t.Logf("\t%s\tTest %d:\tgot: %s, Nonce: %d", failed, testID, tx.To, tx.Nonce)
							t.Logf("\t%s\tTest %d:\texp: %s, Nonce: %d", failed, testID, tst.best[i].To, tst.best[i].Nonce)
							t.Fatalf("\t%s\tTest %d:\tShould get back the right tip/id.", failed, testID)
						}
						t.Logf("\t%s\tTest %d:\tShould get back the right tip/Nonce: %d/%d", success, testID, tx.Tip, tx.Nonce)
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
