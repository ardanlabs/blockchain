package selector

import (
	"sort"

	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

type advancedTips struct {
	bestPos   map[storage.Account]int
	bestTip   uint
	groupTips map[storage.Account][]uint
	groups    []storage.Account
	howMany   int
}

func newAdvancedTips(m map[storage.Account][]storage.BlockTx, howMany int) *advancedTips {
	groupTips := map[storage.Account][]uint{}
	groups := []storage.Account{}
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
	return &advancedTips{groupTips: groupTips, groups: groups, howMany: howMany}
}

func (a *advancedTips) FindBest() map[storage.Account]int {
	a.findBest(0, 0, a.howMany, a.bestPos, 0)
	return a.bestPos
}

func (a *advancedTips) findBest(groupID, pos int, left int, currPos map[storage.Account]int, prevTip uint) {
	if prevTip > a.bestTip {
		a.bestTip = prevTip
		a.bestPos = currPos
	}

	if groupID >= len(a.groups) {
		return
	}
	from := a.groups[groupID]

	for pos, tip := range a.groupTips[from] {
		if left-pos < 0 {
			break
		}
		newCurrPos := copyMap(currPos)
		newCurrPos[from] = pos
		a.findBest(groupID+1, pos, left-pos, newCurrPos, prevTip+tip)
	}
}

func copyMap(m map[storage.Account]int) map[storage.Account]int {
	newCurrPos := map[storage.Account]int{}
	for from, pos := range m {
		newCurrPos[from] = pos
	}
	return newCurrPos
}

// advancedSelect returns transactions with the best tip while respecting the nonce
// for each account/transaction.
var advancedSelect = func(m map[storage.Account][]storage.BlockTx, howMany int) []storage.BlockTx {
	final := []storage.BlockTx{}

	// Sort the transactions per account by nonce.
	for key := range m {
		if len(m[key]) > 1 {
			sort.Sort(byNonce(m[key]))
		}
	}

	a := newAdvancedTips(m, howMany)
	for from, num := range a.FindBest() {
		for i := 0; i < num; i++ {
			final = append(final, m[from][i])
		}
	}

	return final
}
