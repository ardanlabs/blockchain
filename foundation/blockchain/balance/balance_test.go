package balance_test

import (
	"testing"

	"github.com/ardanlabs/blockchain/foundation/blockchain/balance"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

func TestCRUD(t *testing.T) {
	type account struct {
		account string
		value   uint
	}
	type table struct {
		total    uint
		accounts []account
	}

	tt := []table{
		{
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
				sheet := balance.NewSheet(nil)
				for _, acct := range tst.accounts {
					sheet.ApplyValue(acct.account, acct.value)
				}

				values := sheet.Values()
				for _, acct := range tst.accounts {
					if _, exists := values[acct.account]; !exists {
						t.Errorf("\t%s\tTest %d:\tShould be able to find account: %s", failed, testID, acct.account)
					} else {
						t.Logf("\t%s\tTest %d:\tShould be able to find account: %s", success, testID, acct.account)
					}
				}

				var total uint
				for _, v := range values {
					total += v
				}

				if total != tst.total {
					t.Errorf("\t%s\tTest %d:\tShould be able to have the correct total.", failed, testID)
					t.Logf("got: %d", total)
					t.Logf("exp: %d", tst.total)
				} else {
					t.Logf("\t%s\tTest %d:\tShould be able to have the correct total of %d.", success, testID, tst.total)
				}
			}
		}
	}
}
