package database_test

import (
	"errors"
	"testing"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
	"github.com/ethereum/go-ethereum/crypto"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

// =============================================================================

func Test_Transactions(t *testing.T) {
	type table struct {
		name        string
		miner       database.AccountID
		minerReward uint
		gas         uint
		balances    map[string]uint
		final       map[database.AccountID]uint
		txs         []database.UserTx
	}

	tt := []table{
		{
			name:        "basic",
			miner:       "0xFef311483Cc040e1A89fb9bb469eeB8A70935EF8",
			minerReward: 100,
			gas:         80,
			balances: map[string]uint{
				"0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4": 1000,
				"0xF01813E4B85e178A83e29B8E7bF26BD830a25f32": 0,
				"0xFef311483Cc040e1A89fb9bb469eeB8A70935EF8": 0,
			},
			final: map[database.AccountID]uint{
				"0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4": 540,
				"0xF01813E4B85e178A83e29B8E7bF26BD830a25f32": 200,
				"0xFef311483Cc040e1A89fb9bb469eeB8A70935EF8": 360,
			},
			txs: []database.UserTx{
				{
					ToID:  "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32",
					Value: 100,
					Tip:   50,
				},
				{
					ToID:  "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32",
					Value: 100,
					Tip:   50,
				},
			},
		},
	}

	t.Log("Given the need to validate the transactions.")
	{
		for testID, tst := range tt {
			t.Logf("\tTest %d:\tWhen handling a set of database.", testID)
			{
				f := func(t *testing.T) {
					db, err := database.New("", genesis.Genesis{MiningReward: tst.minerReward, Balances: tst.balances}, nil)
					if err != nil {
						t.Fatalf("\t%s\tTest %d:\tShould be able to open database: %v", failed, testID, err)
					}
					t.Logf("\t%s\tTest %d:\tShould be able to open database.", success, testID)

					for _, tx := range tst.txs {
						blockTx, err := sign(tx, tst.gas)
						if err != nil {
							t.Fatalf("\t%s\tTest %d:\tShould be able to sign transaction: %v", failed, testID, err)
						}
						t.Logf("\t%s\tTest %d:\tShould be able to sign transaction.", success, testID)

						if err := db.ApplyTransaction(tst.miner, blockTx); err != nil {
							t.Fatalf("\t%s\tTest %d:\tShould be able to apply transaction: %v", failed, testID, err)
						}
						t.Logf("\t%s\tTest %d:\tShould be able to apply transaction.", success, testID)
					}

					db.ApplyMiningReward(tst.miner)
					t.Logf("\t%s\tTest %d:\tShould be able to apply miner reward.", success, testID)

					accounts := db.CopyAccounts()
					for account, info := range accounts {
						finalValue, exists := tst.final[account]
						if !exists {
							t.Errorf("\t%s\tTest %d:\tShould have account %s in balances.", failed, testID, account)
						} else {
							t.Logf("\t%s\tTest %d:\tShould have account %s in balances.", success, testID, account)
						}

						if finalValue != info.Balance {
							t.Errorf("\t%s\tTest %d:\tShould have correct balances for %s.", failed, testID, account)
							t.Logf("\t%s\tTest %d:\tgot: %d", failed, testID, info.Balance)
							t.Logf("\t%s\tTest %d:\texp: %d", failed, testID, finalValue)
						} else {
							t.Logf("\t%s\tTest %d:\tShould have correct balances for %s.", success, testID, account)
						}
					}
				}

				t.Run(tst.name, f)
			}
		}
	}
}

func TestNonceValidation(t *testing.T) {
	type table struct {
		name        string
		minerReward uint
		gas         uint
		balances    map[string]uint
		txs         []database.UserTx
		results     []error
	}

	tt := []table{
		{
			name:        "basic",
			minerReward: 100,
			balances:    map[string]uint{},
			txs: []database.UserTx{
				{
					Nonce: 5,
					ToID:  "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32",
				},
				{
					Nonce: 3,
					ToID:  "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32",
				},
				{
					Nonce: 6,
					ToID:  "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32",
				},
			},
			results: []error{nil, errors.New("error"), nil},
		},
	}

	t.Log("Given the need to validate new transactions use a proper nonce.")
	{
		for testID, tst := range tt {
			t.Logf("\tTest %d:\tWhen handling a set of transactions.", testID)
			{
				db, err := database.New("", genesis.Genesis{MiningReward: tst.minerReward, Balances: tst.balances}, nil)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to open database: %v", failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to open database.", success, testID)

				for i, tx := range tst.txs {
					blockTx, err := sign(tx, tst.gas)
					if err != nil {
						t.Fatalf("\t%s\tTest %d:\tShould be able to sign transaction: %v", failed, testID, err)
					}
					t.Logf("\t%s\tTest %d:\tShould be able to sign transaction.", success, testID)

					err = db.ValidateNonce(blockTx.SignedTx)
					if (tst.results[i] == nil && err != nil) || (tst.results[i] != nil && err == nil) {
						t.Fatalf("\t%s\tTest %d:\tShould be able to validate nonce correctly.", failed, testID)
					}
					t.Logf("\t%s\tTest %d:\tShould be able to validate nonce correctly.", success, testID)

					err = db.ApplyTransaction("test", blockTx)
					if (tst.results[i] == nil && err != nil) || (tst.results[i] != nil && err == nil) {
						t.Fatalf("\t%s\tTest %d:\tShould be able to apply transaction.", failed, testID)
					}
				}
			}
		}
	}
}

// =============================================================================

func sign(tx database.UserTx, gas uint) (database.BlockTx, error) {
	pk, err := crypto.HexToECDSA("fae85851bdf5c9f49923722ce38f3c1defcfd3619ef5453230a58ad805499959")
	if err != nil {
		return database.BlockTx{}, err
	}

	signedTx, err := tx.Sign(pk)
	if err != nil {
		return database.BlockTx{}, err
	}

	return database.NewBlockTx(signedTx, gas), nil
}
