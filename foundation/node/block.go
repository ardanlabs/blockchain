package node

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"time"
)

// BlockHeader represents common information required for
// each block.
type BlockHeader struct {
	PrevBlock string `json:"prev_block"`
	Number    uint64 `json:"number"`
	Time      uint64 `json:"time"`
	Nonce     uint64 `json:"nonce"`
}

// Block represents a set of transactions grouped together.
type Block struct {
	Header       BlockHeader `json:"header"`
	Transactions []Tx        `json:"txs"`
}

// NewBlock constructs a new BlockFS for persisting.
func NewBlock(prevBlock Block, txs map[string]Tx) Block {
	hash := zeroHash
	if prevBlock.Header.Number > 0 {
		hash = prevBlock.Hash()
	}

	// Store a copy of the transaction to support
	// state changes that don't get written to the mempool.
	cpy := make([]Tx, 0, len(txs))
	for _, tx := range txs {
		cpy = append(cpy, tx)
	}

	return Block{
		Header: BlockHeader{
			PrevBlock: hash,
			Number:    prevBlock.Header.Number + 1,
			Time:      uint64(time.Now().Unix()),
		},
		Transactions: cpy,
	}
}

// Hash returns the unique hash for the block by marshaling
// the block into JSON and performing a hashing operation.
func (b Block) Hash() string {
	if b.Header.Number == 0 {
		return zeroHash
	}

	return generateHash(b)
}

// ToPeerBlock converts a Block to a PeerBlock .
func (b Block) ToPeerBlock() PeerBlock {
	pb := PeerBlock{
		Header: PeerBlockHeader{
			PrevBlock: b.Header.PrevBlock,
			ThisBlock: b.Hash(),
			Number:    b.Header.Number,
			Time:      b.Header.Time,
			Nonce:     b.Header.Nonce,
		},
		Transactions: b.Transactions,
	}

	return pb
}

// BlocksToPeerBlocks converts a slice of blocks to peerblocks.
func BlocksToPeerBlocks(blocks []Block) []PeerBlock {
	pbs := make([]PeerBlock, len(blocks))
	for i, block := range blocks {
		pbs[i] = block.ToPeerBlock()
	}

	return pbs
}

// =============================================================================

// BlockFS represents what is written to the DB file.
type BlockFS struct {
	Hash  string
	Block Block
}

// performPOW does the work to find a valid hash for
// this block.
func performPOW(ctx context.Context, b Block, ev EventHandler) (BlockFS, time.Duration, error) {
	t := time.Now()

	// Choose a random starting point for the nonce.
	nBig, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return BlockFS{}, time.Since(t), ctx.Err()
	}
	nonce := nBig.Uint64()

	var attempts uint64
	for {
		attempts++
		if attempts%1_000_000 == 0 {
			ev("bcWorker: runMiningOperation: miningG: mining attempts[%d]", attempts)
		}

		// Did we timeout trying to solve the problem.
		if ctx.Err() != nil {
			return BlockFS{}, time.Since(t), ctx.Err()
		}

		// Hash the block and check if we have solved the puzzle.
		hash := b.Hash()
		if !isHashSolved(hash) {

			// I may want to track these nonce's to make sure I
			// don't try the same one twice.
			nonce++
			b.Header.Nonce = nonce
			continue
		}

		// Did we timeout trying to solve the problem.
		if ctx.Err() != nil {
			return BlockFS{}, time.Since(t), ctx.Err()
		}

		ev("bcWorker: runMiningOperation: miningG: mining final attempts[%d]", attempts)

		// We found a solution to the POW.
		bfs := BlockFS{
			Hash:  hash,
			Block: b,
		}
		return bfs, time.Since(t), nil
	}
}

// isHashSolved checks the hash to make sure it complies with
// the POW rules. Currently six leading 0's.
func isHashSolved(hash string) bool {
	if len(hash) != 64 {
		return false
	}

	match := "000000"
	return hash[:len(match)] == match
}

// =============================================================================

// PeerBlockHeader represents what a block header looks like from any
// request to a node.
type PeerBlockHeader struct {
	PrevBlock string `json:"prev_block"`
	ThisBlock string `json:"this_block"`
	Number    uint64 `json:"number"`
	Time      uint64 `json:"time"`
	Nonce     uint64 `json:"nonce"`
}

// PeerBlock represents what a block looks like from any
// request to a node.
type PeerBlock struct {
	Header       PeerBlockHeader `json:"header"`
	Transactions []Tx            `json:"transactions"`
}

// ToBlock converts a PeerBlock to a Block.
func (pb PeerBlock) ToBlock() Block {
	b := Block{
		Header: BlockHeader{
			PrevBlock: pb.Header.PrevBlock,
			Number:    pb.Header.Number,
			Time:      pb.Header.Time,
			Nonce:     pb.Header.Nonce,
		},
		Transactions: pb.Transactions,
	}

	return b
}

// PeerBlocksToBlocks converts a slice of peer blocks to blocks.
func PeerBlocksToBlocks(pbs []PeerBlock) []Block {
	blocks := make([]Block, len(pbs))
	for i, pb := range pbs {
		blocks[i] = pb.ToBlock()
	}

	return blocks
}

// =============================================================================

// loadBlocksFromDisk the current set of blocks/transactions. In a real
// world situation this would require a lot of memory.
func loadBlocksFromDisk(dbPath string) ([]Block, error) {
	dbFile, err := os.Open(dbPath)
	if err != nil {
		return nil, err
	}
	defer dbFile.Close()

	var blockNum int
	var blocks []Block
	scanner := bufio.NewScanner(dbFile)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		var blockFS BlockFS
		if err := json.Unmarshal(scanner.Bytes(), &blockFS); err != nil {
			return nil, err
		}

		if blockFS.Block.Hash() != blockFS.Hash {
			return nil, fmt.Errorf("block %d has been changed", blockNum)
		}

		blocks = append(blocks, blockFS.Block)
		blockNum++
	}

	return blocks, nil
}
