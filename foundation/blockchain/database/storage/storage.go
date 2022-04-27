// Package storage handles all the lower level support for reading and writing
// blocks to disk.
package storage

import (
	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/merkle"
)

// block represents what serialized to disk.
type block struct {
	Hash   string               `json:"hash"`
	Header database.BlockHeader `json:"block"`
	Trans  []database.BlockTx   `json:"trans"`
}

// newBlock constructs a block that can be serialized to disk.
func newBlock(dbBlock database.Block) block {
	block := block{
		Hash:   dbBlock.Hash(),
		Header: dbBlock.Header,
		Trans:  dbBlock.Trans.Values(),
	}

	return block
}

// toDatabaseBlock converts a storage block into a database block.
func toDatabaseBlock(block block) (database.Block, error) {
	tree, err := merkle.NewTree(block.Trans)
	if err != nil {
		return database.Block{}, err
	}

	dbBlock := database.Block{
		Header: block.Header,
		Trans:  tree,
	}

	return dbBlock, nil
}
