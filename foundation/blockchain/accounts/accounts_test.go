package accounts_test

import (
	"errors"
	"testing"

	"github.com/ardanlabs/blockchain/foundation/blockchain/accounts"
	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
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

func TestTransactions(t *testing.T) {
	type table struct {
		name        string
		miner       storage.Account
		minerReward uint
		gas         uint
		balances    map[storage.Account]uint
		final       map[storage.Account]uint
		txs         []storage.UserTx
	}

	tt := []table{
		{
			name:        "basic",
			miner:       "miner",
			minerReward: 100,
			gas:         80,
			balances: map[storage.Account]uint{
				"0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4": 1000,
				"0xF01813E4B85e178A83e29B8E7bF26BD830a25f32": 0,
				"miner": 0,
			},
			final: map[storage.Account]uint{
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
					accounts := accounts.New(genesis.Genesis{MiningReward: tst.minerReward, Balances: tst.balances})

					for _, tx := range tst.txs {
						blktx, err := sign(tx, tst.gas)
						if err != nil {
							t.Fatalf("\t%s\tTest %d:\tShould be able to sign transaction.", failed, testID)
						}
						t.Logf("\t%s\tTest %d:\tShould be able to sign transaction.", success, testID)

						if err := accounts.ApplyTransaction(tst.miner, blktx); err != nil {
							t.Fatalf("\t%s\tTest %d:\tShould be able to apply transaction.", failed, testID)
						}
						t.Logf("\t%s\tTest %d:\tShould be able to apply transaction.", success, testID)
					}

					accounts.ApplyMiningReward(tst.miner)
					t.Logf("\t%s\tTest %d:\tShould be able to apply miner reward.", success, testID)

					cpyAccount := accounts.Copy()
					for account, info := range cpyAccount {
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
		balances    map[storage.Account]uint
		txs         []storage.UserTx
		results     []error
	}

	tt := []table{
		{
			name:        "basic",
			minerReward: 100,
			balances:    map[storage.Account]uint{},
			txs: []storage.UserTx{
				{
					Nonce: 5,
					To:    "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32",
				},
				{
					Nonce: 3,
					To:    "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32",
				},
				{
					Nonce: 6,
					To:    "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32",
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
				accounts := accounts.New(genesis.Genesis{MiningReward: tst.minerReward, Balances: tst.balances})

				for i, tx := range tst.txs {
					blktx, err := sign(tx, tst.gas)
					if err != nil {
						t.Fatalf("\t%s\tTest %d:\tShould be able to sign transaction.", failed, testID)
					}
					t.Logf("\t%s\tTest %d:\tShould be able to sign transaction.", success, testID)

					err = accounts.ValidateNonce(blktx.SignedTx)
					if (tst.results[i] == nil && err != nil) || (tst.results[i] != nil && err == nil) {
						t.Fatalf("\t%s\tTest %d:\tShould be able to validate nonce correctly.", failed, testID)
					}
					t.Logf("\t%s\tTest %d:\tShould be able to validate nonce correctly.", success, testID)

					err = accounts.ApplyTransaction("test", blktx)
					if (tst.results[i] == nil && err != nil) || (tst.results[i] != nil && err == nil) {
						t.Fatalf("\t%s\tTest %d:\tShould be able to apply transaction.", failed, testID)
					}
				}
			}
		}
	}
}
