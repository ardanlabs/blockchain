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

	blockTx := storage.BlockTx{
		SignedTx: signedTx,
		Gas:      gas,
	}

	return blockTx, nil
}

func TestTransactions(t *testing.T) {
	type table struct {
		name  string
		miner string
		gas   uint
		sheet map[string]uint
		final map[string]uint
		txs   []storage.UserTx
	}

	tt := []table{
		{
			name:  "basic",
			miner: "miner",
			gas:   80,
			sheet: map[string]uint{
				"0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4": 1000,
				"0xF01813E4B85e178A83e29B8E7bF26BD830a25f32": 0,
				"miner": 0,
			},
			final: map[string]uint{
				"0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4": 540,
				"0xF01813E4B85e178A83e29B8E7bF26BD830a25f32": 200,
				"miner": 260,
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
					sheet := balance.NewSheet(tst.sheet)

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

func TestCRUD(t *testing.T) {
	type account struct {
		address string
		value   uint
	}
	type table struct {
		name     string
		total    uint
		accounts []account
	}

	tt := []table{
		{
			name:  "basic",
			total: 200,
			accounts: []account{
				{"0xF01813E4B85e178A83e29B8E7bF26BD830a25f32", 100},
				{"0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4", 100},
			},
		},
	}

	t.Log("Given the need to validate the CRUD API.")
	{
		for testID, tst := range tt {
			t.Logf("\tTest %d:\tWhen handling a set of accounts.", testID)
			{
				f := func(t *testing.T) {
					sheet := balance.NewSheet(nil)
					for _, acct := range tst.accounts {
						sheet.ApplyValue(acct.address, acct.value)
					}

					values := sheet.Values()
					for _, acct := range tst.accounts {
						if _, exists := values[acct.address]; !exists {
							t.Errorf("\t%s\tTest %d:\tShould be able to find account: %s", failed, testID, acct.address)
						} else {
							t.Logf("\t%s\tTest %d:\tShould be able to find account: %s", success, testID, acct.address)
						}
					}

					var total uint
					for _, v := range values {
						total += v
					}

					if total != tst.total {
						t.Errorf("\t%s\tTest %d:\tShould be able to have the correct total.", failed, testID)
						t.Logf("\t%s\tTest %d:\tgot: %d", failed, testID, total)
						t.Logf("\t%s\tTest %d:\texp: %d", failed, testID, tst.total)
					} else {
						t.Logf("\t%s\tTest %d:\tShould be able to have the correct total of %d.", success, testID, tst.total)
					}
				}

				t.Run(tst.name, f)
			}
		}
	}
}
