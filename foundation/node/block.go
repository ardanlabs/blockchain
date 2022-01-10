package node

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
}

// Block represents a set of transactions grouped together.
type Block struct {
	Header       BlockHeader
	Transactions []Tx
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

// BlockFS represents what is written to the DB file.
type BlockFS struct {
	Hash  string
	Block Block
}

// NewBlockFS constructs a new BlockFS for persisting.
func NewBlockFS(prevBlock Block, transactions []Tx) (BlockFS, error) {
	hash := zeroHash
	if prevBlock.Header.Number > 0 {
		hash = prevBlock.Hash()
	}

	block := Block{
		Header: BlockHeader{
			PrevBlock: hash,
			Number:    prevBlock.Header.Number + 1,
			Time:      uint64(time.Now().Unix()),
		},
		Transactions: transactions,
	}

	blockFS := BlockFS{
		Hash:  block.Hash(),
		Block: block,
	}

	return blockFS, nil
}

// =============================================================================

// PeerBlockHeader represents what a block header looks like from any
// request to a node.
type PeerBlockHeader struct {
	PrevBlock string `json:"prev_block"`
	ThisBlock string `json:"this_block"`
	Number    uint64 `json:"number"`
	Time      uint64 `json:"time"`
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

// loadBlocksFromDisk the current set of blocks/transactions.
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

// func toNodeBlock(block block) node.Block {
// 	nBlock := node.Block{
// 		Header: node.BlockHeader{
// 			PrevBlock: hashToBytes(block.Header.PrevBlock),
// 			Number:    block.Header.Number,
// 			Time:      uint64(block.Header.Time.Unix()),
// 		},
// 		Transactions: toNodeTxs(block.Transactions),
// 	}

// 	return nBlock
// }
