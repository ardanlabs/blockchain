package selector

import (
	"sort"

	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// tipSelect returns transactions with the best tip while respecting the nonce
// for each account/transaction.
var tipSelect = func(m map[storage.AccountID][]storage.BlockTx, howMany int) []storage.BlockTx {

	/*
		Bill: {Nonce: 2, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 250},
			  {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 150},
		Pavl: {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 200},
			  {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 75},
		Edua: {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 75},
			  {Nonce: 1, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 100},
	*/

	// Sort the transactions per account by nonce.
	for key := range m {
		if len(m[key]) > 1 {
			sort.Sort(byNonce(m[key]))
		}
	}

	/*
		Bill: {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 150},
		      {Nonce: 2, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 250},
		Pavl: {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 75},
		      {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 200},
		Edua: {Nonce: 1, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 100},
		      {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 75},
	*/

	// Pick the first transaction in the slice for each account. Each iteration
	// represents a new row of selections. Keep doing that until all the
	// transactions have been selected.
	var rows [][]storage.BlockTx
	for {
		var row []storage.BlockTx
		for key := range m {
			if len(m[key]) > 0 {
				row = append(row, m[key][0])
				m[key] = m[key][1:]
			}
		}
		if row == nil {
			break
		}
		rows = append(rows, row)
	}

	/*
		0: Bill: {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 150},
		0: Pavl: {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 75},
		0: Edua: {Nonce: 1, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 100},
		1: Bill: {Nonce: 2, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 250},
		1: Pavl: {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 200},
		1: Edua: {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 75},
	*/

	// Sort each row by tip unless we will take all transactions from that row
	// anyway. Then try to select the number of requested transactions. Keep
	// pulling transactions from each row until the amount of fulfilled or
	// there are no more transactions.
	final := []storage.BlockTx{}
done:
	for _, row := range rows {
		need := howMany - len(final)
		if len(row) > need {
			sort.Sort(byTip(row))
			final = append(final, row[:need]...)
			break done
		}
		final = append(final, row...)
	}

	/*
		0: Bill: {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 150},
		1: Pavl: {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 75},
		2: Edua: {Nonce: 1, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 100},
		3: Bill: {Nonce: 2, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 250},
	*/

	return final
}
