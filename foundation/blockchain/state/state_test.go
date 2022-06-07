package state_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage/memory"
	"github.com/ethereum/go-ethereum/crypto"
)

// ============================ TESTS CASES ===================================

// Test_MineAndSyncBlock is the simple happy path. We do a transaction, mine a
// block and offer it to another miner. No issues should be found.
func Test_MineAndSyncBlock(t *testing.T) {
	node1 := newNode(MINER1_PRIVATEKEY, t)
	node2 := newNode(MINER2_PRIVATEKEY, t)

	tx := database.Tx{
		ChainID: CHAIN_ID,
		Nonce:   1,
		ToID:    KENNEDY_ACCOUNTID,
		Value:   1,
		Tip:     0,
		Data:    nil,
	}

	signedTx := newSignedTx(tx, JACK_PRIVATEKEY, t)
	if err := node1.UpsertWalletTransaction(signedTx); err != nil {
		t.Fatalf("Error upserting wallet transaction: %v", err)
	}

	blk, err := node1.MineNewBlock(context.Background())
	if err != nil {
		t.Fatalf("Error mining new block: %v", err)
	}

	err = node2.ProcessProposedBlock(blk)
	if err != nil {
		t.Fatalf("Error proposing new block: %v", err)
	}
}

// Test_MineAndErrChainForkedDetection will create 2 nodes, mine some blocks on
// node 1, then provide the blocks to node 2. Some blocks won't be forwarded to
// node 2. This scenario should raise a database.ErrChainForked.
func Test_MineAndErrChainForkedDetection(t *testing.T) {
	node1 := newNode(MINER1_PRIVATEKEY, t)
	node2 := newNode(MINER2_PRIVATEKEY, t)

	// Let's add 15 blocks to Node1 starting with Nonce 1.
	var blocks []database.Block

	for i := 1; i <= 15; i++ {
		tx := database.Tx{
			ChainID: CHAIN_ID,
			Nonce:   uint64(i),
			ToID:    KENNEDY_ACCOUNTID,
			Value:   1,
			Tip:     0,
			Data:    nil,
		}

		signedTx := newSignedTx(tx, JACK_PRIVATEKEY, t)
		if err := node1.UpsertWalletTransaction(signedTx); err != nil {
			t.Fatalf("Error upserting wallet transaction: %v", err)
		}

		blk, err := node1.MineNewBlock(context.Background())
		if err != nil {
			t.Fatalf("Error mining new block: %v", err)
		}

		blocks = append(blocks, blk)
	}

	// Let's add the first 10 blocks to node2, then skip blocks 11 and 12,
	// then try to add block 13. Remember zero indexing.

	for i, blk := range blocks[:13] {
		switch {
		case i < 10:
			if err := node2.ProcessProposedBlock(blk); err != nil {
				t.Fatalf("Error proposing new block %d: %v", i, err)
			}

		case i == 10 || i == 11:
			continue

		case i == 12:
			err := node2.ProcessProposedBlock(blk)
			if !errors.Is(err, database.ErrChainForked) {
				t.Fatal("Error handling missing blocks - should have received" +
					" ErrChainForked error")
			}
		}
	}
}

// Test_MineAndForceMissingBlock - in this scenario we will create
// 2 Miners, mine some blocks on Miner 1, and then, provide the blocks to Miner
// 2 but one block will be missing. A error with the message: "this block is not
// the next number, got 12, exp 11". is expected as result.
func Test_MineAndForceMissingBlock(t *testing.T) {
	node1 := newNode(MINER1_PRIVATEKEY, t)
	node2 := newNode(MINER2_PRIVATEKEY, t)

	// Let's add 15 blocks to Node1 starting with Nonce 1.
	var blocks []database.Block

	for i := 1; i <= 15; i++ {
		tx := database.Tx{
			ChainID: CHAIN_ID,
			Nonce:   uint64(i),
			ToID:    KENNEDY_ACCOUNTID,
			Value:   1,
			Tip:     0,
			Data:    nil,
		}

		signedTx := newSignedTx(tx, JACK_PRIVATEKEY, t)
		if err := node1.UpsertWalletTransaction(signedTx); err != nil {
			t.Fatalf("Error upserting wallet transaction: %v", err)
		}

		blk, err := node1.MineNewBlock(context.Background())
		if err != nil {
			t.Fatalf("Error mining new block: %v", err)
		}

		blocks = append(blocks, blk)
	}

	// Let's add the first 10 blocks to node2, then skip blocks 11 and 12,
	// then try to add block 13. Remember zero indexing.

	for i, blk := range blocks[:13] {
		switch {
		case i < 10:
			if err := node2.ProcessProposedBlock(blk); err != nil {
				t.Fatalf("Error proposing new block %d: %v", i, err)
			}

		case i == 10:
			continue

		case i == 11:
			err := node2.ProcessProposedBlock(blk)
			if err == nil {
				t.Fatal(
					"Error handling missing block - and error should have " +
						"been triggered when processing 11th block")
			}
		}
	}
}

// ================== TOOLKIT FOR TESTS =======================================

const (
	MINER1_PRIVATEKEY = "8dc79feefd3b86e2f9991def0e5ccd9a5128e104682407b308594bc1032ac7f0"
	MINER2_PRIVATEKEY = "5aed92a29e1014d83c1d8ac755878723d7e44d8dc129610d11b2022d09ad95bd"
	MINER3_PRIVATEKEY = "ce07a51ad1d72084aed971b24042f320b4673e852b59eb550375b9eb6747d74a"
	JACK_PRIVATEKEY   = "9f332e3700d8fc2446eaf6d15034cf96e0c2745e40353deef032a5dbf1dfed93"
	JILL_PRIVATEKEY   = "aed31b6b5a341af8f27e66fb0b7633cf20fc27049e3eb7f6f623a4655b719ebb"
	SAMMY_PRIVATEKEY  = "601d7574860c135e9d3c1d52b0ee997404130edc2a1177c78fda92dd6a3dc2f7"

	KENNEDY_ACCOUNTID = database.AccountID("0xF01813E4B85e178A83e29B8E7bF26BD830a25f32")
	PAVEL_ACCOUNTID   = database.AccountID("0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4")
	CESAR_ACCOUNTID   = database.AccountID("0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76")
	BABA_ACCOUNTID    = database.AccountID("0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9")
	ED_ACCOUNTID      = database.AccountID("0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0")
	MINER1_ACCOUNTID  = database.AccountID("0xFef311483Cc040e1A89fb9bb469eeB8A70935EF8")
	MINER2_ACCOUNTID  = database.AccountID("0xb8Ee4c7ac4ca3269fEc242780D7D960bd6272a61")

	NONCE_ZERO = 0
	CHAIN_ID   = 1
)

// =============================================================================

// noopWorker implements the Worker interface which does nothing.
type noopWorker struct{}

func (n noopWorker) Shutdown() {}

func (n noopWorker) Sync() {}

func (n noopWorker) SignalStartMining() {}

func (n noopWorker) SignalCancelMining() {}

func (n noopWorker) SignalShareTx(blockTx database.BlockTx) {}

// =============================================================================

// newGenesis will create a new Genesis.
func newGenesis() genesis.Genesis {
	g := genesis.Genesis{
		Date:          time.Now().Add(time.Hour * 24 * -365),
		ChainID:       CHAIN_ID,
		TransPerBlock: 10,
		Difficulty:    1,
		MiningReward:  700,
		GasPrice:      15,
		Balances: map[string]uint64{
			"0xF01813E4B85e178A83e29B8E7bF26BD830a25f32": 1000000,
			"0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4": 1000000,
		},
	}

	return g
}

// newSignedTx constructs a signed transaction.
func newSignedTx(tx database.Tx, hexKey string, t *testing.T) database.SignedTx {
	privateKey, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		t.Fatalf("Error constructing private key: %v", err)
	}

	signedTx, err := tx.Sign(privateKey)
	if err != nil {
		t.Fatalf("Error signing transaction: %v", err)
	}

	return signedTx
}

// newNode will create an in memory miner.
func newNode(hexKey string, t *testing.T) *state.State {
	if hexKey == "" {
		t.Fatalf("Error with hexKey being empty.")
	}

	privateKey, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		t.Fatalf("Error constructing private key: %v", err)
	}

	storage, err := memory.New()
	if err != nil {
		t.Fatalf("Error setting up memory storage: %v", err)
	}

	state, err := state.New(state.Config{
		BeneficiaryID:  database.PublicKeyToAccountID(privateKey.PublicKey),
		Host:           "http://localhost:9080",
		Genesis:        newGenesis(),
		Storage:        storage,
		SelectStrategy: "Tip",
		KnownPeers:     peer.NewPeerSet(),
		EvHandler:      func(v string, args ...any) {},
	})
	if err != nil {
		t.Fatalf("Error constructing node state: %v", err)
	}

	state.Worker = noopWorker{}
	return state
}
