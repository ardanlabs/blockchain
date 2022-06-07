package state_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage/memory"
	"github.com/ardanlabs/blockchain/foundation/events"
	"github.com/ardanlabs/blockchain/foundation/logger"
	"github.com/ethereum/go-ethereum/crypto"
)

// ================== TOOLKIT FOR TESTS =======================================

const (
	MINER1_ECDSA   = "8dc79feefd3b86e2f9991def0e5ccd9a5128e104682407b308594bc1032ac7f0"
	MINER2_ECDSA   = "5aed92a29e1014d83c1d8ac755878723d7e44d8dc129610d11b2022d09ad95bd"
	MINER3_ECDSA   = "ce07a51ad1d72084aed971b24042f320b4673e852b59eb550375b9eb6747d74a"
	PERSONA1_ECDSA = "9f332e3700d8fc2446eaf6d15034cf96e0c2745e40353deef032a5dbf1dfed93"
	PERSONA2_ECDSA = "aed31b6b5a341af8f27e66fb0b7633cf20fc27049e3eb7f6f623a4655b719ebb"
	PERSONA3_ECDSA = "601d7574860c135e9d3c1d52b0ee997404130edc2a1177c78fda92dd6a3dc2f7"
	GENESIS        = `{
		"date": "2021-12-17T00:00:00.000000000Z",
		"chain_id": 1,
		"trans_per_block": 10,
		"difficulty": 1,
		"mining_reward": 700,
		"gas_price": 15,
		"balances": {
			"0xF01813E4B85e178A83e29B8E7bF26BD830a25f32": 1000000,
			"0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4": 1000000
		}
	}`
)

type noopWorker struct {
}

func (n *noopWorker) Shutdown()                              {}
func (n *noopWorker) Sync()                                  {}
func (n *noopWorker) SignalStartMining()                     {}
func (n *noopWorker) SignalCancelMining()                    {}
func (n *noopWorker) SignalShareTx(blockTx database.BlockTx) {}

type signedTxOpts struct {
	keyHex string
	nonce  uint64
	to     string
	value  uint64
	tip    uint64
	data   []byte
}

func createSignedTX(opts signedTxOpts) (database.SignedTx, error) {

	privateKey, err := crypto.HexToECDSA(opts.keyHex)
	if err != nil {
		return database.SignedTx{}, err
	}

	toAccount, err := database.ToAccountID(opts.to)
	if err != nil {
		return database.SignedTx{}, err
	}

	const chainID = 1
	tx, err := database.NewTx(chainID, opts.nonce, toAccount, opts.value, opts.tip, opts.data)
	if err != nil {
		return database.SignedTx{}, err
	}

	return tx.Sign(privateKey)

}

func ifErrFailNow(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}

func newGenesis(t *testing.T) genesis.Genesis {
	minergenesis := genesis.Genesis{}
	err := json.Unmarshal([]byte(GENESIS), &minergenesis)
	ifErrFailNow(t, err)
	return minergenesis
}

// newTxFactory will create a method that will provide us database.SignedTx
// with random value at each call, but preserving control over nonce.
func newTxFactory(t *testing.T) func() database.SignedTx {
	nonce := uint64(0)
	mx := sync.Mutex{}
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	ret := func() database.SignedTx {
		mx.Lock()
		defer mx.Unlock()
		nonce++
		so := signedTxOpts{
			keyHex: PERSONA1_ECDSA,
			nonce:  nonce,
			to:     "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76",
			value:  r1.Uint64(),
			tip:    0,
			data:   []byte{},
		}
		ret, err := createSignedTX(so)
		if err != nil {
			t.Errorf("Error creating database transaction: %s", err.Error())
			t.FailNow()
			return database.SignedTx{}
		}
		return ret
	}

	return ret
}

// newMiner will create an in memory miner
func newMiner(t *testing.T, strkey string) *state.State {
	if strkey == "" {
		t.Fatalf("please provide a string w an ECDSA key as strkey")
	}
	var err error

	storage, err := memory.New()
	ifErrFailNow(t, err)

	key, err := crypto.HexToECDSA(MINER1_ECDSA)
	ifErrFailNow(t, err)

	evts := events.New()

	ev := func(v string, args ...any) {
		const websocketPrefix = "viewer:"

		s := fmt.Sprintf(v, args...)
		//log.Infow(s, "traceid", "00000000-0000-0000-0000-000000000000")
		if strings.HasPrefix(s, websocketPrefix) {
			evts.Send(s)
		}
	}

	ret, err := state.New(state.Config{
		BeneficiaryID:  database.PublicKeyToAccountID(key.PublicKey),
		Host:           "http://localhost:9080",
		Genesis:        newGenesis(t),
		Storage:        storage,
		SelectStrategy: "Tip",
		KnownPeers:     peer.NewPeerSet(),
		EvHandler:      ev,
	})
	ifErrFailNow(t, err)

	ret.Worker = &noopWorker{}
	return ret
}

// ============================ TESTS CASES ===================================

// Test_MineAndSyncBlock - Simple happy path. We do a transaction, mine a
// block and offer it to another miner - no issues should be found.
func Test_MineAndSyncBlock(t *testing.T) {

	log, err := logger.New("TEST")
	ifErrFailNow(t, err)
	defer log.Sync()

	// Miner 1

	state1 := newMiner(t, MINER1_ECDSA)
	state2 := newMiner(t, MINER2_ECDSA)
	txFactory := newTxFactory(t)

	txOpts := txFactory()

	state1.UpsertWalletTransaction(txOpts)

	// Let them interact
	blk, err := state1.MineNewBlock(context.Background())
	ifErrFailNow(t, err)

	err = state2.ProcessProposedBlock(blk)
	ifErrFailNow(t, err)

	log.Info("Done")
}

// Test_MineAndSyncBlocksNotRespectingOrder - in this scenario we will create
// 2 Miners, mine some blocks on Miner 1, and then, provide the blocks to Miner 2,
// but some blocks will be missing and we expect to see it raising a database.ErrChainForked
func Test_MineAndForceRescynError(t *testing.T) {

	log, err := logger.New("TEST")
	ifErrFailNow(t, err)
	defer log.Sync()

	state1 := newMiner(t, MINER1_ECDSA)

	state2 := newMiner(t, MINER2_ECDSA)

	txFactory := newTxFactory(t)

	// lets play with states now
	blocks := make([]database.Block, 0)
	for i := 0; i < 20; i++ {
		txOpts := txFactory()
		err = state1.UpsertWalletTransaction(txOpts)
		ifErrFailNow(t, err)
		blk, err := state1.MineNewBlock(context.Background())
		ifErrFailNow(t, err)
		blocks = append(blocks, blk)
	}

	for i, b := range blocks {

		if i == 12 {
			err = state2.ProcessProposedBlock(b)
			if err == nil {
				t.Fatal("Should have failed here, fork issue")
			} else if err == database.ErrChainForked {
				t.Logf("Failling gracefully due to fork in blocks: %s", err.Error())
				return
			} else {
				t.Fatalf("Ohoh - we gotta an unplanned error: %s", err.Error())
			}
		} else if i != 10 && i != 11 {
			err = state2.ProcessProposedBlock(b)
			if err != nil {
				t.Fatal(err)
			}
		} else {
			t.Log("This is the 10th, lets skip it and see.")
		}
	}
}

// Test_MineAndSyncBlocksNotRespectingOrder - in this scenario we will create
// 2 Miners, mine some blocks on Miner 1, and then, provide the blocks to Miner 2,
// but some blocks will be missing and we expect to see it raising a database.ErrChainForked
func Test_MineAndForceMissingBlock(t *testing.T) {

	log, err := logger.New("TEST")
	ifErrFailNow(t, err)
	defer log.Sync()

	state1 := newMiner(t, MINER1_ECDSA)

	state2 := newMiner(t, MINER2_ECDSA)

	txFactory := newTxFactory(t)

	// lets play with states now
	blocks := make([]database.Block, 0)
	for i := 0; i < 20; i++ {
		txOpts := txFactory()
		err = state1.UpsertWalletTransaction(txOpts)
		ifErrFailNow(t, err)
		blk, err := state1.MineNewBlock(context.Background())
		ifErrFailNow(t, err)
		blocks = append(blocks, blk)
	}

	for i, b := range blocks {

		if i == 11 {
			err = state2.ProcessProposedBlock(b)
			if err == nil {
				t.Fatal("Should have failed here, fork issue")
			} else if err == database.ErrChainForked {
				t.Fatalf("Ohoh - we gotta an unplanned error: %s", err.Error())
				return
			} else {
				t.Logf("Nice, there should be an error here: %s", err.Error())
				return
			}
		} else if i != 10 {
			err = state2.ProcessProposedBlock(b)
			if err != nil {
				t.Fatal(err)
			}
		} else {
			t.Log("This is the 10th, lets skip it and see.")
		}
	}
}
