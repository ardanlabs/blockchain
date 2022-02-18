package blockchain

import (
	"context"
	"crypto/rand"
	"math"
	"math/big"
	"time"
)

// zeroHash represents a hash code of zeros.
const zeroHash string = "00000000000000000000000000000000"

// =============================================================================

// BlockHeader represents common information required for each block.
type BlockHeader struct {
	ParentHash   string `json:"parent_hash"`   // Hash of the previous block in the chain.
	MinerAddress string `json:"miner_address"` // The address of the miner who mined the block.
	Difficulty   int    `json:"difficulty"`    // Number of 0's needed to solve the hash solution.
	Number       uint64 `json:"number"`        // Block number in the chain.
	TotalTip     uint   `json:"total_tip"`     // Total tip paid by all senders as an incentive.
	TotalGas     uint   `json:"total_gas"`     // Total gas fee to recover computation costs paid by the sender.
	TimeStamp    uint64 `json:"timestamp"`     // Time the block was mined.
	Nonce        uint64 `json:"nonce"`         // Value identified to solve the hash solution.
}

// Block represents a group of transactions batched together.
type Block struct {
	Header       BlockHeader `json:"header"`
	Transactions []BlockTx   `json:"txs"`
}

// newBlock constructs a new BlockFS for persisting.
func newBlock(minerAddress string, difficulty int, transPerBlock int, parentBlock Block, txMempool *txMempool) Block {
	parentHash := zeroHash
	if parentBlock.Header.Number > 0 {
		parentHash = parentBlock.Hash()
	}

	// Copy the best transactions from the mempool for this new block.
	cpy := txMempool.copyBestByTip(transPerBlock)

	return Block{
		Header: BlockHeader{
			ParentHash:   parentHash,
			MinerAddress: minerAddress,
			Difficulty:   difficulty,
			Number:       parentBlock.Header.Number + 1,
			TimeStamp:    uint64(time.Now().Unix()),
		},
		Transactions: cpy,
	}
}

// Hash returns the unique hash for the Block.
func (b Block) Hash() string {
	if b.Header.Number == 0 {
		return zeroHash
	}

	return hash(b)
}

// =============================================================================

// blockFS represents what is written to the DB file.
type blockFS struct {
	Hash  string
	Block Block
}

// performPOW does the work of mining to find a valid hash for a specified
// block and returns a BlockFS ready to be written to disk.
func performPOW(ctx context.Context, difficulty int, b Block, ev EventHandler) (blockFS, time.Duration, error) {
	ev("worker: runMiningOperation: MINING: POW: started")
	defer ev("worker: runMiningOperation: MINING: POW: completed")

	for _, tx := range b.Transactions {
		ev("worker: runMiningOperation: MINING: POW: tx[%s]", tx.Hash())
	}

	t := time.Now()

	// Choose a random starting point for the nonce.
	nBig, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return blockFS{}, time.Since(t), ctx.Err()
	}
	b.Header.Nonce = nBig.Uint64()

	var attempts uint64
	for {
		attempts++
		if attempts%1_000_000 == 0 {
			ev("worker: runMiningOperation: MINING: POW: attempts[%d]", attempts)
		}

		// Did we timeout trying to solve the problem.
		if ctx.Err() != nil {
			ev("worker: runMiningOperation: MINING: POW: CANCELLED")
			return blockFS{}, time.Since(t), ctx.Err()
		}

		// Hash the block and check if we have solved the puzzle.
		hash := b.Hash()
		if !isHashSolved(difficulty, hash) {

			// I may want to track these nonce's to make sure I
			// don't try the same one twice.
			b.Header.Nonce++
			continue
		}

		// Did we timeout trying to solve the problem.
		if ctx.Err() != nil {
			ev("worker: runMiningOperation: MINING: POW: CANCELLED")
			return blockFS{}, time.Since(t), ctx.Err()
		}

		ev("worker: runMiningOperation: MINING: POW: SOLVED: prevBlk[%s]: newBlk[%s]", b.Header.ParentHash, b.Hash())
		ev("worker: runMiningOperation: MINING: POW: attempts[%d]", attempts)

		// We found a solution to the POW.
		bfs := blockFS{
			Hash:  hash,
			Block: b,
		}
		return bfs, time.Since(t), nil
	}
}

// isHashSolved checks the hash to make sure it complies with
// the POW rules. We need to match a difficulty number of 0's.
func isHashSolved(difficulty int, hash string) bool {
	const match = "00000000000000000"

	if len(hash) != 64 {
		return false
	}

	return hash[:difficulty] == match[:difficulty]
}
