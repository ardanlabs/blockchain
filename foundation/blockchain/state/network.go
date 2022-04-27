package state

import (
	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/merkle"
)

// NetBlock represents what is serialized over the network.
type NetBlock struct {
	Hash   string               `json:"hash"`
	Header database.BlockHeader `json:"block"`
	Trans  []database.BlockTx   `json:"trans"`
}

// NewNetBlock constructs a block that can be serialized over the network.
func NewNetBlock(block database.Block) NetBlock {
	netBlock := NetBlock{
		Hash:   block.Hash(),
		Header: block.Header,
		Trans:  block.Trans.Values(),
	}

	return netBlock
}

// toDatabaseBlock converts a storage block into a database block.
func toDatabaseBlock(netBlock NetBlock) (database.Block, error) {
	tree, err := merkle.NewTree(netBlock.Trans)
	if err != nil {
		return database.Block{}, err
	}

	block := database.Block{
		Header: netBlock.Header,
		Trans:  tree,
	}

	return block, nil
}
