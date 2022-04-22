package database

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
	PrevBlockHash string    `json:"prev_block_hash"` // Bitcoin: Hash of the previous block in the chain.
	TimeStamp     uint64    `json:"timestamp"`       // Bitcoin: Time the block was mined.
	Nonce         uint64    `json:"nonce"`           // Bitcoin: Value identified to solve the hash solution.
	BeneficiaryID AccountID `json:"beneficiary"`     // Ethereum: The account who is receiving fees and tips.
	Difficulty    uint      `json:"difficulty"`      // Ethereum: Number of 0's needed to solve the hash solution.
	Number        uint64    `json:"number"`          // Ethereum: Block number in the chain.
	TransRoot     string    `json:"trans_root"`      // Bitcoin/Ethereum: Represents the merkle tree root hash for the transactions in this block.
}

// Block represents a group of transactions batched together.
type Block struct {
	Header BlockHeader
	Trans  *merkle.Tree[BlockTx]
}

// POW constructs a new Block and performs the work to find a nonce that
// solves the cryptographic POW puzzel.
func POW(ctx context.Context, beneficiaryID AccountID, difficulty uint, prevBlock Block, trans []BlockTx, evHandler func(v string, args ...any)) (Block, error) {

	// When mining the first block, the previous block's hash will be zero.
	prevBlockHash := signature.ZeroHash
	if prevBlock.Header.Number > 0 {
		prevBlockHash = prevBlock.Hash()
	}

	// Construct a merkle tree from the transaction for this block. The root
	// of this tree will be part of the block to be mined.
	tree, err := merkle.NewTree(trans)
	if err != nil {
		return Block{}, err
	}

	// Construct the block to be mined.
	nb := Block{
		Header: BlockHeader{
			PrevBlockHash: prevBlockHash,
			TimeStamp:     uint64(time.Now().UTC().Unix()), //
			Nonce:         0,                               // Will be identified by the POW algorithm.
			BeneficiaryID: beneficiaryID,
			Difficulty:    difficulty,
			Number:        prevBlock.Header.Number + 1,
			TransRoot:     tree.MerkleRootHex(),
		},
		Trans: tree,
	}

	// Peform the proof of work mining operation.
	if err := nb.performPOW(ctx, evHandler); err != nil {
		return Block{}, err
	}

	return nb, nil
}

// performPOW does the work of mining to find a valid hash for a specified
// block. Pointer semantics are being used since a nonce is being discovered.
func (b *Block) performPOW(ctx context.Context, ev func(v string, args ...any)) error {
	ev("worker: PerformPOW: MINING: started")
	defer ev("worker: PerformPOW: MINING: completed")

	// Log the transactions that are a part of this potential block.
	for _, tx := range b.Trans.Values() {
		ev("worker: PerformPOW: MINING: tx[%s]", tx)
	}

	// Choose a random starting point for the nonce. After this, the nonce
	// will be incremented by 1 until a solution is found by us or another node.
	nBig, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return ctx.Err()
	}
	b.Header.Nonce = nBig.Uint64()

	// Loop until we or another node finds a solution for the next block.
	var attempts uint64
	for {
		attempts++
		if attempts%1_000_000 == 0 {
			ev("worker: PerformPOW: MINING: attempts[%d]", attempts)
		}

		// Did we timeout trying to solve the problem.
		if ctx.Err() != nil {
			ev("worker: PerformPOW: MINING: CANCELLED")
			return ctx.Err()
		}

		// Hash the block and check if we have solved the puzzle.
		hash := b.Hash()
		if !isHashSolved(b.Header.Difficulty, hash) {
			b.Header.Nonce++
			continue
		}

		// Did we timeout trying to solve the problem.
		if ctx.Err() != nil {
			ev("worker: PerformPOW: MINING: CANCELLED")
			return ctx.Err()
		}

		ev("worker: PerformPOW: MINING: SOLVED: prevBlk[%s]: newBlk[%s]", b.Header.PrevBlockHash, hash)
		ev("worker: PerformPOW: MINING: attempts[%d]", attempts)

		return nil
	}
}

// Hash returns the unique hash for the Block.
func (b Block) Hash() string {
	if b.Header.Number == 0 {
		return signature.ZeroHash
	}

	// CORE NOTE: Hashing the block header and not the whole block so the blockchain
	// can be cryptographically checked by only needing block headers and not full
	// blocks with the transaction data. This will support the ability to have pruned
	// nodes in the future. A pruned node may only keep 1000 full blocks of data and
	// will still be capable of validating all new blocks and transactions in real time.

	return signature.Hash(b.Header)
}

// ValidateBlock takes a block and validates it to be included into the blockchain.
func (b Block) ValidateBlock(previousBlock Block, evHandler func(v string, args ...any)) error {
	evHandler("storage: ValidateBlock: validate: blk[%d]: check: chain is not forked", b.Header.Number)

	// The node who sent this block has a chain that is two or more blocks ahead
	// of ours. This means there has been a fork and we are on the wrong side.
	nextNumber := previousBlock.Header.Number + 1
	if b.Header.Number >= (nextNumber + 2) {
		return ErrChainForked
	}

	evHandler("storage: ValidateBlock: validate: blk[%d]: check: block difficulty is the same or greater than parent block difficulty", b.Header.Number)

	if b.Header.Difficulty < previousBlock.Header.Difficulty {
		return fmt.Errorf("block difficulty is less than previous block difficulty, parent %d, block %d", previousBlock.Header.Difficulty, b.Header.Difficulty)
	}

	evHandler("storage: ValidateBlock: validate: blk[%d]: check: block hash has been solved", b.Header.Number)

	hash := b.Hash()
	if !isHashSolved(b.Header.Difficulty, hash) {
		return fmt.Errorf("%s invalid block hash", hash)
	}

	evHandler("storage: ValidateBlock: validate: blk[%d]: check: block number is the next number", b.Header.Number)

	if b.Header.Number != nextNumber {
		return fmt.Errorf("this block is not the next number, got %d, exp %d", b.Header.Number, nextNumber)
	}

	evHandler("storage: ValidateBlock: validate: blk[%d]: check: parent hash does match parent block", b.Header.Number)

	if b.Header.PrevBlockHash != previousBlock.Hash() {
		return fmt.Errorf("parent block hash doesn't match our known parent, got %s, exp %s", b.Header.PrevBlockHash, previousBlock.Hash())
	}

	if previousBlock.Header.TimeStamp > 0 {
		evHandler("storage: ValidateBlock: validate: blk[%d]: check: block's timestamp is greater than parent block's timestamp", b.Header.Number)

		parentTime := time.Unix(int64(previousBlock.Header.TimeStamp), 0)
		blockTime := time.Unix(int64(b.Header.TimeStamp), 0)
		if !blockTime.After(parentTime) {
			return fmt.Errorf("block timestamp is before parent block, parent %s, block %s", parentTime, blockTime)
		}

		// This is a check that Ethereum does but we can't because we don't run all the time.

		// evHandler("storage: ValidateBlock: validate: blk[%d]: check: block is less than 15 minutes apart from parent block", b.Header.Number)

		// dur := blockTime.Sub(parentTime)
		// if dur.Seconds() > time.Duration(15*time.Second).Seconds() {
		// 	return fmt.Errorf("block is older than 15 minutes, duration %v", dur)
		// }
	}

	evHandler("storage: ValidateBlock: validate: blk[%d]: check: merkle root does match transactions", b.Header.Number)

	if b.Header.TransRoot != b.Trans.MerkleRootHex() {
		return fmt.Errorf("merkle root does not match transactions, got %s, exp %s", b.Trans.MerkleRootHex(), b.Header.TransRoot)
	}

	return nil
}

// isHashSolved checks the hash to make sure it complies with
// the POW rules. We need to match a difficulty number of 0's.
func isHashSolved(difficulty uint, hash string) bool {
	const match = "00000000000000000"

	if len(hash) != 64 {
		return false
	}

	return hash[:difficulty] == match[:difficulty]
}

// =============================================================================

// BlockFS represents what is written to the DB file.
type BlockFS struct {
	Hash  string      `json:"hash"`
	Block BlockHeader `json:"block"`
	Trans []BlockTx   `json:"trans"`
}

// NewBlockFS constructs the value to serialize to disk.
func NewBlockFS(block Block) BlockFS {
	bfs := BlockFS{
		Hash:  block.Hash(),
		Block: block.Header,
		Trans: block.Trans.Values(),
	}

	return bfs
}

// ToBlock converts a BlockFS into a Block.
func ToBlock(blockFS BlockFS) (Block, error) {
	tree, err := merkle.NewTree(blockFS.Trans)
	if err != nil {
		return Block{}, err
	}

	nb := Block{
		Header: blockFS.Block,
		Trans:  tree,
	}

	return nb, nil
}
