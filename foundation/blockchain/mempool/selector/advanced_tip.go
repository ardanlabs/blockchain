package selector

import (
	"sort"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
)

// advancedTipSelect returns transactions with the best tip while respecting the nonce
// for each account/transaction. This strategy takes into account high-value transactions
// that happens to be stuck on a low-nonce transaction with a low tip price.
var advancedTipSelect = func(m map[database.AccountID][]database.BlockTx, howMany int) []database.BlockTx {
	final := []database.BlockTx{}

	// Sort the transactions per account by nonce.
	for key := range m {
		if len(m[key]) > 1 {
			sort.Sort(byNonce(m[key]))
		}
	}

	at := newAdvancedTips(m, howMany)
	for from, num := range at.findBest() {
		for i := 0; i < num; i++ {
			final = append(final, m[from][i])
		}
	}

	return final
}

// =============================================================================

type advancedTips struct {
	howMany   int
	bestTip   uint
	bestPos   map[database.AccountID]int
	groupTips map[database.AccountID][]uint
	groups    []database.AccountID
}

func newAdvancedTips(m map[database.AccountID][]database.BlockTx, howMany int) *advancedTips {
	groupTips := map[database.AccountID][]uint{}
	groups := []database.AccountID{}

	for from := range m {
		groupTips[from] = []uint{0}
		groups = append(groups, from)
	}

	for from, group := range m {
		for i, tx := range group {
			if i > howMany {
				break
			}
			groupTips[from] = append(groupTips[from], tx.Tip+groupTips[from][i])
		}
	}

	return &advancedTips{
		howMany:   howMany,
		groupTips: groupTips,
		groups:    groups,
	}
}

func (at *advancedTips) findBest() map[database.AccountID]int {
	at.findBestTransactions(0, 0, at.howMany, at.bestPos, 0)
	return at.bestPos
}

func (at *advancedTips) findBestTransactions(groupID, pos int, left int, currPos map[database.AccountID]int, prevTip uint) {
	if prevTip > at.bestTip {
		at.bestTip = prevTip
		at.bestPos = currPos
	}

	if groupID >= len(at.groups) {
		return
	}
	from := at.groups[groupID]

	for pos, tip := range at.groupTips[from] {
		if left-pos < 0 {
			break
		}

		newCurrPos := copyMap(currPos)
		newCurrPos[from] = pos
		at.findBestTransactions(groupID+1, pos, left-pos, newCurrPos, prevTip+tip)
	}
}

// =============================================================================

func copyMap(m map[database.AccountID]int) map[database.AccountID]int {
	newCurrPos := map[database.AccountID]int{}
	for from, pos := range m {
		newCurrPos[from] = pos
	}

	return newCurrPos
}
