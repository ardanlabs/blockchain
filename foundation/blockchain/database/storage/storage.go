// Package storage handles all the lower level support for reading and writing
// blocks to disk.
package storage

import (
	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/merkle"
)

// Block represents what serialized to disk and over the network.
type Block struct {
	Hash   string               `json:"hash"`
	Header database.BlockHeader `json:"block"`
	Trans  []database.BlockTx   `json:"trans"`
}

// NewBlock constructs a block that can be serialized to disk and network.
func NewBlock(dbBlock database.Block) Block {
	block := Block{
		Hash:   dbBlock.Hash(),
		Header: dbBlock.Header,
		Trans:  dbBlock.Trans.Values(),
	}

	return block
}

// ToDatabaseBlock converts a storage block into a database block.
func ToDatabaseBlock(block Block) (database.Block, error) {
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
