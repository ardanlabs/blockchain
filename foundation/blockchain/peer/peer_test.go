package peer_test

import (
	"testing"

	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

// =============================================================================

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

	t.Log("Given the need to validate mempool api.")
	{
		for testID, tst := range tt {
			t.Logf("\tTest %d:\tWhen handling a set of transaction.", testID)
			{
				f := func(t *testing.T) {
					ps := peer.NewPeerSet()

					for _, peer := range tst.peers {
						t.Logf("\t%s\tTest %d:\tShould be able to add new peer: %s", success, testID, peer.Host)
						ps.Add(peer)
					}

					peers := ps.Copy("")
					if len(peers) != len(tst.peers) {
						t.Logf("\t%s\tTest %d:\tgot: %d", failed, testID, len(peers))
						t.Logf("\t%s\tTest %d:\texp: %d", failed, testID, len(tst.peers)-1)
						t.Fatalf("\t%s\tTest %d:\tShould get back the right peers.", failed, testID)
					}
					t.Logf("\t%s\tTest %d:\tShould get back the right peers: %d", success, testID, len(peers))

					peers = ps.Copy("host2")
					if len(peers) != len(tst.peers)-1 {
						t.Logf("\t%s\tTest %d:\tgot: %d", failed, testID, len(peers))
						t.Logf("\t%s\tTest %d:\texp: %d", failed, testID, len(tst.peers)-1)
						t.Fatalf("\t%s\tTest %d:\tShould get back the right peers.", failed, testID)
					}
					t.Logf("\t%s\tTest %d:\tShould get back the right peers: %d", success, testID, len(peers))
				}

				t.Run(tst.name, f)
			}
		}
	}
}
