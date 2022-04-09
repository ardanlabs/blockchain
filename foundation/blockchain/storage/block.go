package storage

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/merkle"
	"github.com/ardanlabs/blockchain/foundation/blockchain/signature"
)

// ErrChainForked is returned from validateNextBlock if another node's chain
// is two or more blocks ahead of ours.
var ErrChainForked = errors.New("blockchain forked, start resync")

// =============================================================================

// BlockHeader represents common information required for each block.
type BlockHeader struct {
	ParentHash   string  `json:"parent_hash"`   // Hash of the previous block in the chain.
	MinerAccount Account `json:"miner_account"` // The account of the miner who mined the block.
	Difficulty   int     `json:"difficulty"`    // Number of 0's needed to solve the hash solution.
	Number       uint64  `json:"number"`        // Block number in the chain.
	MerkleRoot   string  `json:"merkle_root"`   // Represents the merkle tree root hash for the transactions in this block.
	TimeStamp    uint64  `json:"timestamp"`     // Time the block was mined.
	Nonce        uint64  `json:"nonce"`         // Value identified to solve the hash solution.
}

// Block represents a group of transactions batched together.
type Block struct {
	Header       BlockHeader `json:"header"`
	Transactions []BlockTx   `json:"txs"`
}

// NewBlock constructs a new BlockFS for persisting.
func NewBlock(minerAccount Account, difficulty int, transPerBlock int, parentBlock Block, trans []BlockTx) (Block, error) {
	parentHash := signature.ZeroHash
	if parentBlock.Header.Number > 0 {
		parentHash = parentBlock.Hash()
	}

	tree, err := merkle.NewTree(trans)
	if err != nil {
		return Block{}, err
	}

	nb := Block{
		Header: BlockHeader{
			ParentHash:   parentHash,
			MinerAccount: minerAccount,
			Difficulty:   difficulty,
			Number:       parentBlock.Header.Number + 1,
			MerkleRoot:   tree.MerkelRootHex(),
			TimeStamp:    uint64(time.Now().UTC().Unix()),
		},
		Transactions: trans,
	}

	return nb, nil
}

// Hash returns the unique hash for the Block.
func (b Block) Hash() string {
	if b.Header.Number == 0 {
		return signature.ZeroHash
	}

	// Using the block header because the data is smaller for the unmarshal
	// operation and the merkel root will let us validate no transaction
	// has been hampered with.

	return signature.Hash(b.Header)
}

// ValidateBlock takes a block and validates it to be included into the blockchain.
func (b Block) ValidateBlock(parentBlock Block, evHandler func(v string, args ...any)) (string, error) {
	evHandler("storage: ValidateBlock: validate: blk[%d]: check: chain is not forked", b.Header.Number)

	// The node who sent this block has a chain that is two or more blocks ahead
	// of ours. This means there has been a fork and we are on the wrong side.
	nextNumber := parentBlock.Header.Number + 1
	if b.Header.Number >= (nextNumber + 2) {
		return signature.ZeroHash, ErrChainForked
	}

	evHandler("storage: ValidateBlock: validate: blk[%d]: check: block difficulty is the same or greater than parent block difficulty", b.Header.Number)

	if b.Header.Difficulty < parentBlock.Header.Difficulty {
		return signature.ZeroHash, fmt.Errorf("block difficulty is less than parent block difficulty, parent %d, block %d", parentBlock.Header.Difficulty, b.Header.Difficulty)
	}

	evHandler("storage: ValidateBlock: validate: blk[%d]: check: block hash has been solved", b.Header.Number)

	hash := b.Hash()
	if !isHashSolved(b.Header.Difficulty, hash) {
		return signature.ZeroHash, fmt.Errorf("%s invalid block hash", hash)
	}

	evHandler("storage: ValidateBlock: validate: blk[%d]: check: block number is the next number", b.Header.Number)

	if b.Header.Number != nextNumber {
		return signature.ZeroHash, fmt.Errorf("this block is not the next number, got %d, exp %d", b.Header.Number, nextNumber)
	}

	evHandler("storage: ValidateBlock: validate: blk[%d]: check: parent hash does match parent block", b.Header.Number)

	if b.Header.ParentHash != parentBlock.Hash() {
		return signature.ZeroHash, fmt.Errorf("parent block hash doesn't match our known parent, got %s, exp %s", b.Header.ParentHash, parentBlock.Hash())
	}

	if parentBlock.Header.TimeStamp > 0 {
		evHandler("storage: ValidateBlock: validate: blk[%d]: check: block's timestamp is greater than parent block's timestamp", b.Header.Number)

		parentTime := time.Unix(int64(parentBlock.Header.TimeStamp), 0)
		blockTime := time.Unix(int64(b.Header.TimeStamp), 0)
		if !blockTime.After(parentTime) {
			return signature.ZeroHash, fmt.Errorf("block timestamp is before parent block, parent %s, block %s", parentTime, blockTime)
		}

		// This is a check that Ethereum does but we can't because we don't run all the time.

		// evHandler("storage: ValidateBlock: validate: blk[%d]: check: block is less than 15 minutes apart from parent block", b.Header.Number)

		// dur := blockTime.Sub(parentTime)
		// if dur.Seconds() > time.Duration(15*time.Second).Seconds() {
		// 	return signature.ZeroHash, fmt.Errorf("block is older than 15 minutes, duration %v", dur)
		// }
	}

	evHandler("storage: ValidateBlock: validate: blk[%d]: check: could create merkle tree from transactions", b.Header.Number)

	tree, err := merkle.NewTree(b.Transactions)
	if err != nil {
		return signature.ZeroHash, fmt.Errorf("unable to create merkle tree from transactions, %w", err)
	}

	evHandler("storage: ValidateBlock: validate: blk[%d]: check: merkle root does match transactions", b.Header.Number)

	if b.Header.MerkleRoot != tree.MerkelRootHex() {
		return signature.ZeroHash, fmt.Errorf("merkle root does not match transactions, got %s, exp %s", tree.MerkelRootHex(), b.Header.MerkleRoot)
	}

	return hash, nil
}

// PerformPOW does the work of mining to find a valid hash for a specified
// block and returns a BlockFS ready to be written to disk.
func (b Block) PerformPOW(ctx context.Context, difficulty int, ev func(v string, args ...any)) (BlockFS, time.Duration, error) {
	ev("worker: PerformPOW: MINING: started")
	defer ev("worker: PerformPOW: MINING: completed")

	for _, tx := range b.Transactions {
		ev("worker: PerformPOW: MINING: tx[%s]", tx)
	}

	t := time.Now()

	// Choose a random starting point for the nonce.
	nBig, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return BlockFS{}, time.Since(t), ctx.Err()
	}
	b.Header.Nonce = nBig.Uint64()

	var attempts uint64
	for {
		attempts++
		if attempts%1_000_000 == 0 {
			ev("worker: PerformPOW: MINING: attempts[%d]", attempts)
		}

		// Did we timeout trying to solve the problem.
		if ctx.Err() != nil {
			ev("worker: PerformPOW: MINING: CANCELLED")
			return BlockFS{}, time.Since(t), ctx.Err()
		}

		// Hash the block and check if we have solved the puzzle.
		hash := b.Hash()
		if !isHashSolved(difficulty, hash) {
			b.Header.Nonce++
			continue
		}

		// Did we timeout trying to solve the problem.
		if ctx.Err() != nil {
			ev("worker: PerformPOW: MINING: CANCELLED")
			return BlockFS{}, time.Since(t), ctx.Err()
		}

		ev("worker: PerformPOW: MINING: SOLVED: prevBlk[%s]: newBlk[%s]", b.Header.ParentHash, hash)
		ev("worker: PerformPOW: MINING: attempts[%d]", attempts)

		// We found a solution to the POW.
		bfs := BlockFS{
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

// =============================================================================

// BlockFS represents what is written to the DB file.
type BlockFS struct {
	Hash  string
	Block Block
}
