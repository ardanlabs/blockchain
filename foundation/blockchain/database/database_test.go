package database_test

import (
	"errors"
	"testing"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
	"github.com/ethereum/go-ethereum/crypto"
)

func Test_Transactions(t *testing.T) {
	type table struct {
		name        string
		miner       database.AccountID
		minerReward uint64
		gas         uint64
		balances    map[string]uint64
		final       map[database.AccountID]uint64
		txs         []database.Tx
	}

	tt := []table{
		{
			name:        "basic",
			miner:       "0xFef311483Cc040e1A89fb9bb469eeB8A70935EF8",
			minerReward: 100,
			gas:         80,
			balances: map[string]uint64{
				"0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4": 1000,
				"0xF01813E4B85e178A83e29B8E7bF26BD830a25f32": 0,
				"0xFef311483Cc040e1A89fb9bb469eeB8A70935EF8": 0,
			},
			final: map[database.AccountID]uint64{
				"0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4": 540,
				"0xF01813E4B85e178A83e29B8E7bF26BD830a25f32": 200,
				"0xFef311483Cc040e1A89fb9bb469eeB8A70935EF8": 360,
			},
			txs: []database.Tx{
				{
					ChainID: 1,
					Nonce:   1,
					FromID:  "0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4",
					ToID:    "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32",
					Value:   100,
					Tip:     50,
				},
				{
					ChainID: 1,
					Nonce:   2,
					FromID:  "0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4",
					ToID:    "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32",
					Value:   100,
					Tip:     50,
				},
			},
		},
	}

	for _, tst := range tt {
		f := func(t *testing.T) {
			db, err := database.New(genesis.Genesis{ChainID: 1, MiningReward: tst.minerReward, Balances: tst.balances}, MockStorage{}, nil)
			if err != nil {
				t.Fatalf("Test %s:\tShould be able to open database: %v", tst.name, err)
			}

			for _, tx := range tst.txs {
				blockTx, err := sign(tx, tst.gas)
				if err != nil {
					t.Fatalf("Test %s:\tShould be able to sign transaction: %v", tst.name, err)
				}

				if err := db.ApplyTransaction(database.Block{Header: database.BlockHeader{BeneficiaryID: tst.miner}}, blockTx); err != nil {
					t.Fatalf("Test %s:\tShould be able to apply transaction: %v", tst.name, err)
				}
			}

			db.ApplyMiningReward(database.Block{Header: database.BlockHeader{BeneficiaryID: tst.miner, MiningReward: tst.minerReward}})

			accounts := db.Copy()
			for account, info := range accounts {
				finalValue, exists := tst.final[account]
				if !exists {
					t.Errorf("Test %s:\tShould have account %s in balances.", tst.name, account)
				}

				if finalValue != info.Balance {
					t.Errorf("Test %s:\tShould have correct balances for %s.", tst.name, account)
					t.Logf("Test %s:\tgot: %d", tst.name, info.Balance)
					t.Logf("Test %s:\texp: %d", tst.name, finalValue)
				}
			}
		}

		t.Run(tst.name, f)
	}
}

func TestNonceValidation(t *testing.T) {
	type table struct {
		name        string
		miner       database.AccountID
		minerReward uint64
		gas         uint64
		balances    map[string]uint64
		txs         []database.Tx
		results     []error
	}

	tt := []table{
		{
			name:        "basic",
			miner:       "0xFef311483Cc040e1A89fb9bb469eeB8A70935EF8",
			minerReward: 100,
			balances:    map[string]uint64{"0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4": 10},
			txs: []database.Tx{
				{
					ChainID: 1,
					Nonce:   1,
					FromID:  "0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4",
					ToID:    "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32",
				},
				{
					ChainID: 1,
					Nonce:   5,
					FromID:  "0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4",
					ToID:    "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32",
				},
				{
					ChainID: 1,
					Nonce:   2,
					FromID:  "0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4",
					ToID:    "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32",
				},
			},
			results: []error{nil, errors.New("error"), nil},
		},
	}

	for _, tst := range tt {
		db, err := database.New(genesis.Genesis{ChainID: 1, MiningReward: tst.minerReward, Balances: tst.balances}, MockStorage{}, nil)
		if err != nil {
			t.Fatalf("Test %s:\tShould be able to open database: %v", tst.name, err)
		}

		for i, tx := range tst.txs {
			blockTx, err := sign(tx, tst.gas)
			if err != nil {
				t.Fatalf("Test %s:\tShould be able to sign transaction: %v", tst.name, err)
			}

			err = db.ApplyTransaction(database.Block{Header: database.BlockHeader{BeneficiaryID: tst.miner}}, blockTx)
			if (tst.results[i] == nil && err != nil) || (tst.results[i] != nil && err == nil) {
				t.Fatalf("Test %s:\tShould be able to apply transaction : %s", tst.name, err)
			}
		}
	}
}

// =============================================================================

func sign(tx database.Tx, gas uint64) (database.BlockTx, error) {
	pk, err := crypto.HexToECDSA("fae85851bdf5c9f49923722ce38f3c1defcfd3619ef5453230a58ad805499959")
	if err != nil {
		return database.BlockTx{}, err
	}

	signedTx, err := tx.Sign(pk)
	if err != nil {
		return database.BlockTx{}, err
	}

	return database.NewBlockTx(signedTx, gas, 1), nil
}

// =============================================================================

type MockIterator struct{}

func (mi MockIterator) Next() (database.BlockData, error) {
	return database.BlockData{}, nil
}

func (mi MockIterator) Done() bool {
	return true
}

type MockStorage struct{}

func (ms MockStorage) Write(block database.BlockData) error {
	return nil
}

func (ms MockStorage) GetBlock(num uint64) (database.BlockData, error) {
	return database.BlockData{}, nil
}

func (ms MockStorage) ForEach() database.Iterator {
	return &MockIterator{}
}

func (ms MockStorage) Close() error {
	return nil
}

func (ms MockStorage) Reset() error {
	return nil
}
