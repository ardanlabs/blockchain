package chain

import (
	"crypto/sha256"
	"encoding/json"
)

// BlockHeader represents common information required for
// each block.
type BlockHeader struct {
	ParentHash [32]byte // Parent block reference.
	Time       uint64
}

// Block represents a set of transactions grouped together.
type Block struct {
	Header  BlockHeader
	Payload []Tx // New transactions only (payload).
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
