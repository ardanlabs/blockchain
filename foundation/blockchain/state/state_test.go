package state_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage/memory"
	"github.com/ardanlabs/blockchain/foundation/blockchain/worker"
	"github.com/ardanlabs/blockchain/foundation/events"
	"github.com/ardanlabs/blockchain/foundation/logger"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	MINER_ECDSA = "8dc79feefd3b86e2f9991def0e5ccd9a5128e104682407b308594bc1032ac7f0"
)

func ifErrFailNow(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}

func Test_MineAndSyncBlock(t *testing.T) {

	log, err := logger.New("TEST")
	ifErrFailNow(t, err)
	defer log.Sync()

	storage, err := memory.New()
	ifErrFailNow(t, err)

	key, err := crypto.HexToECDSA(MINER_ECDSA)
	ifErrFailNow(t, err)

	evts := events.New()

	ev := func(v string, args ...any) {
		const websocketPrefix = "viewer:"

		s := fmt.Sprintf(v, args...)
		log.Infow(s, "traceid", "00000000-0000-0000-0000-000000000000")
		if strings.HasPrefix(s, websocketPrefix) {
			evts.Send(s)
		}
	}

	state, err := state.New(state.Config{
		BeneficiaryID:  database.PublicKeyToAccountID(key.PublicKey),
		Host:           "http://localhost:9080",
		Storage:        storage,
		SelectStrategy: "Tip",
		KnownPeers:     peer.NewPeerSet(),
		EvHandler:      ev,
	})
	ifErrFailNow(t, err)

	worker.Run(state, ev)
	log.Info("Done")
}
