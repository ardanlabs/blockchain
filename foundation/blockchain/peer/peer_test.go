package peer_test

import (
	"testing"

	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
)

func Test_CRUD(t *testing.T) {
	type table struct {
		name  string
		peers []peer.Peer
	}

	tt := []table{
		{
			name:  "basic",
			peers: []peer.Peer{{Host: "host1"}, {Host: "host2"}, {Host: "host3"}},
		},
	}

	for _, tst := range tt {
		f := func(t *testing.T) {
			ps := peer.NewPeerSet()

			for _, peer := range tst.peers {
				ps.Add(peer)
			}

			peers := ps.Copy("")
			if len(peers) != len(tst.peers) {
				t.Logf("Test %s:\tgot: %d", tst.name, len(peers))
				t.Logf("Test %s:\texp: %d", tst.name, len(tst.peers)-1)
				t.Fatalf("Test %s:\tShould get back the right peers.", tst.name)
			}

			peers = ps.Copy("host2")
			if len(peers) != len(tst.peers)-1 {
				t.Logf("Test %s:\tgot: %d", tst.name, len(peers))
				t.Logf("Test %s:\texp: %d", tst.name, len(tst.peers)-1)
				t.Fatalf("Test %s:\tShould get back the right peers.", tst.name)
			}
		}

		t.Run(tst.name, f)
	}
}
