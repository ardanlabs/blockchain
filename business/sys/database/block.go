package database

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// BlockHeader represents common information required for
// each block.
type BlockHeader struct {
	PrevBlock [32]byte
	Time      uint64
}

// Block represents a set of transactions grouped together.
type Block struct {
	Header       BlockHeader
	Transactions []Tx
}

// Hash returns the unique hash for the block by marshaling
// the block into JSON and performing a hashing operation.
func (b Block) Hash() ([32]byte, error) {
	blockJson, err := json.Marshal(b)
	if err != nil {
		return [32]byte{}, err
	}

	return sha256.Sum256(blockJson), nil
}

// BlockFS represents what is written to the DB file.
type BlockFS struct {
	Hash  [32]byte
	Block Block
}

// NewBlockFS constructs a new BlockFS for persisting.
func NewBlockFS(prevBlock [32]byte, transactions []Tx) (BlockFS, error) {
	block := Block{
		Header: BlockHeader{
			PrevBlock: prevBlock,
			Time:      uint64(time.Now().Unix()),
		},
		Transactions: transactions,
	}

	hash, err := block.Hash()
	if err != nil {
		return BlockFS{}, err
	}

	blockFS := BlockFS{
		Hash:  hash,
		Block: block,
	}

	return blockFS, nil
}

// =============================================================================

// loadBlocks the current set of blocks/transactions.
func loadBlocks(dbPath string) ([]Block, error) {
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

		hash, err := blockFS.Block.Hash()
		if err != nil {
			return nil, err
		}

		if hash != blockFS.Hash {
			return nil, fmt.Errorf("block %d has been changed", blockNum)
		}

		blocks = append(blocks, blockFS.Block)
		blockNum++
	}

	return blocks, nil
}
