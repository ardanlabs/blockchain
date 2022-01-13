package node

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"time"
)

const zeroHash = "00000000000000000000000000000000"

// BlockHeader represents common information required for
// each block.
type BlockHeader struct {
	PrevBlock string
	Number    uint64
	Time      uint64
	Nonce     uint64
}

// Block represents a set of transactions grouped together.
type Block struct {
	Header       BlockHeader
	Transactions []Tx
}

// NewBlock constructs a new BlockFS for persisting.
func NewBlock(prevBlock Block, transactions []Tx) Block {
	hash := zeroHash
	if prevBlock.Header.Number > 0 {
		hash = prevBlock.Hash()
	}

	return Block{
		Header: BlockHeader{
			PrevBlock: hash,
			Number:    prevBlock.Header.Number + 1,
			Time:      uint64(time.Now().Unix()),
		},
		Transactions: transactions,
	}
}

// Hash returns the unique hash for the block by marshaling
// the block into JSON and performing a hashing operation.
func (b Block) Hash() string {
	if b.Header.Number == 0 {
		return zeroHash
	}

	blockJson, err := json.Marshal(b)
	if err != nil {
		return zeroHash
	}

	hash := sha256.Sum256(blockJson)
	return hex.EncodeToString(hash[:])
}

// Converts a Block to a PeerBlock .
func (b Block) ToPeerBlock() PeerBlock {
	pb := PeerBlock{
		Header: PeerBlockHeader{
			PrevBlock: b.Header.PrevBlock,
			ThisBlock: b.Hash(),
			Number:    b.Header.Number,
			Time:      b.Header.Time,
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
func performPOW(ctx context.Context, b Block) (BlockFS, error) {
	for {
		// Did we timeout trying to solve the problem.
		if ctx.Err() != nil {
			return BlockFS{}, ctx.Err()
		}

		// Hash the block and check if we have solved the puzzle.
		hash := b.Hash()
		if !isHashSolved(hash) {

			// Choose a randon number so we can try again.
			const max = 1_000_000
			nBig, err := rand.Int(rand.Reader, big.NewInt(max))
			if err != nil {
				return BlockFS{}, err
			}
			b.Header.Nonce = nBig.Uint64()

			continue
		}

		// We found a solution to the POW.
		bfs := BlockFS{
			Hash:  hash,
			Block: b,
		}
		return bfs, nil
	}
}

// =============================================================================

// PeerBlockHeader represents what a block header looks like from any
// request to a node.
type PeerBlockHeader struct {
	PrevBlock string `json:"prev_block"`
	ThisBlock string `json:"this_block"`
	Number    uint64 `json:"number"`
	Time      uint64 `json:"time"`
	Nonce     uint32 `json:"nonce"`
}

// peerTx represents what a block looks like from any
// request to a node.
type PeerBlock struct {
	Header       PeerBlockHeader `json:"header"`
	Transactions []Tx            `json:"transactions"`
}

// Converts a PeerBlock to a Block.
func (pb PeerBlock) ToBlock() Block {
	b := Block{
		Header: BlockHeader{
			PrevBlock: pb.Header.PrevBlock,
			Number:    pb.Header.Number,
			Time:      pb.Header.Time,
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

// isHashSolved checks the hash to make sure it complies with
// the POW rules. Currently two leading 0's.
func isHashSolved(hash string) bool {
	if len(hash) != 64 {
		return false
	}

	if hash[:2] != "00" {
		return false
	}

	return true
}

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
