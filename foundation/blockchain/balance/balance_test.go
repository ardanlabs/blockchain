package balance_test

import (
	"testing"

	"github.com/ardanlabs/blockchain/foundation/blockchain/balance"
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

func TestTransactions(t *testing.T) {
	type table struct {
		name        string
		miner       string
		minerReward uint
		gas         uint
		sheet       map[string]uint
		final       map[string]uint
		txs         []storage.UserTx
	}

	tt := []table{
		{
			name:        "basic",
			miner:       "miner",
			minerReward: 100,
			gas:         80,
			sheet: map[string]uint{
				"0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4": 1000,
				"0xF01813E4B85e178A83e29B8E7bF26BD830a25f32": 0,
				"miner": 0,
			},
			final: map[string]uint{
				"0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4": 540,
				"0xF01813E4B85e178A83e29B8E7bF26BD830a25f32": 200,
				"miner": 360,
			},
			txs: []storage.UserTx{
				{
					To:    "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32",
					Value: 100,
					Tip:   50,
				},
				{
					To:    "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32",
					Value: 100,
					Tip:   50,
				},
			},
		},
	}

	t.Log("Given the need to validate the transactions.")
	{
		for testID, tst := range tt {
			t.Logf("\tTest %d:\tWhen handling a set of accounts.", testID)
			{
				f := func(t *testing.T) {
					sheet := balance.NewSheet(tst.minerReward, tst.sheet)

					for _, tx := range tst.txs {
						blktx, err := sign(tx, tst.gas)
						if err != nil {
							t.Fatalf("\t%s\tTest %d:\tShould be able to sign transaction.", failed, testID)
						}
						t.Logf("\t%s\tTest %d:\tShould be able to sign transaction.", success, testID)

						if err := sheet.ApplyTransaction(tst.miner, blktx); err != nil {
							t.Fatalf("\t%s\tTest %d:\tShould be able to apply transaction.", failed, testID)
						}
						t.Logf("\t%s\tTest %d:\tShould be able to apply transaction.", success, testID)
					}

					sheet.ApplyMiningReward(tst.miner)
					t.Logf("\t%s\tTest %d:\tShould be able to apply miner reward.", success, testID)

					values := sheet.Values()
					for addr, value := range values {
						finalValue, exists := tst.final[addr]
						if !exists {
							t.Errorf("\t%s\tTest %d:\tShould have account %s in balances.", failed, testID, addr)
						} else {
							t.Logf("\t%s\tTest %d:\tShould have account %s in balances.", success, testID, addr)
						}

						if finalValue != value {
							t.Errorf("\t%s\tTest %d:\tShould have correct balances for %s.", failed, testID, addr)
							t.Logf("\t%s\tTest %d:\tgot: %d", failed, testID, value)
							t.Logf("\t%s\tTest %d:\texp: %d", failed, testID, finalValue)
						} else {
							t.Logf("\t%s\tTest %d:\tShould have correct balances for %s.", success, testID, addr)
						}
					}
				}

				t.Run(tst.name, f)
			}
		}
	}
}
