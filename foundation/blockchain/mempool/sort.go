package mempool

import (
	"sort"

	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

type SortStrategy func(transactions map[string][]storage.BlockTx, howMany int) []storage.BlockTx

// SimpleSort returns a list of the best transactions for the next
// mining operation. The caller specifies how many transactions they want.
// Pass -1 for all the transactions.
// The algorithm focuses on the transactions with the best tip while
// respecting the nonce for each address/transaction.
var SimpleSort SortStrategy = func(m map[string][]storage.BlockTx, howMany int) []storage.BlockTx {

	/*
		Bill: {Nonce: 2, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 250},
			  {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 150},
		Pavl: {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 200},
			  {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 75},
		Edua: {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 75},
			  {Nonce: 1, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 100},
	*/

	// Sort the transactions per address by nonce.
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

	// Pick the first transaction in the slice for each address. Each iteration
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
	var final []storage.BlockTx
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

// batch_size = 6
//
// transactions = [
//     {"from": 1, "fee": 1, "number": 1},
//     {"from": 1, "fee": 2, "number": 2},
//     {"from": 1, "fee": 3, "number": 3},
//     {"from": 1, "fee": 3, "number": 4},
//     {"from": 2, "fee": 1, "number": 1},
//     {"from": 2, "fee": 4, "number": 2},
//     {"from": 2, "fee": 1, "number": 3},
// ]
//
// grouped_transactions = dict()
// for t in transactions:
//     # group by "from"
//     grouped_transactions.setdefault(t["from"], list()).append(t)
//
// for v in grouped_transactions.values():
//     # order by "number"
//     v.sort(key=lambda x: x["number"])
//
// groups = list(grouped_transactions.keys())
// group_fees = [[0] for _ in range(len(groups))]
// for idx, group_name in enumerate(groups):
//     g_fees = group_fees[idx]
//     for i, t in enumerate(grouped_transactions[group_name]):
//         if i >= batch_size:
//             break
//         g_fees.append(t["fee"] + g_fees[-1])
//
// print(group_fees)
//
// ctx = dict(best_position=[0] * len(groups), best_fee=-1)
//
// print("group_fees", group_fees)

// def find_best_fee(group_idx, remainder, current_position, prev_fee):
//     if group_idx >= len(group_fees):
//         return
//     if remainder <= 0:
//         return
//     for pos, g_fee in enumerate(group_fees[group_idx]):
//         new_remainder = remainder - pos
//         if new_remainder < 0:
//             break
//         position = list(current_position)
//         position[group_idx] = pos
//         fee = prev_fee + group_fees[group_idx][pos]
//         if fee > ctx["best_fee"]:
//             ctx["best_fee"] = fee
//             ctx["best_position"] = position
//         find_best_fee(group_idx + 1, new_remainder, position, fee)
//
//
// find_best_fee(0, batch_size, [0] * len(groups), 0)
//
// print(ctx)
