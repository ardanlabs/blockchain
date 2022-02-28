package storage

import (
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/signature"
)

// =============================================================================

// BlockHeader represents common information required for each block.
type BlockHeader struct {
	ParentHash   string `json:"parent_hash"`   // Hash of the previous block in the chain.
	MinerAddress string `json:"miner_address"` // The address of the miner who mined the block.
	Difficulty   int    `json:"difficulty"`    // Number of 0's needed to solve the hash solution.
	Number       uint64 `json:"number"`        // Block number in the chain.
	TotalTip     uint   `json:"total_tip"`     // Total tip paid by all senders as an incentive.
	TotalGas     uint   `json:"total_gas"`     // Total gas fee to recover computation costs paid by the sender.
	TimeStamp    uint64 `json:"timestamp"`     // Time the block was mined.
	Nonce        uint64 `json:"nonce"`         // Value identified to solve the hash solution.
}

// Block represents a group of transactions batched together.
type Block struct {
	Header       BlockHeader `json:"header"`
	Transactions []BlockTx   `json:"txs"`
}

// NewBlock constructs a new BlockFS for persisting.
func NewBlock(minerAddress string, difficulty int, transPerBlock int, parentBlock Block, trans []BlockTx) Block {
	parentHash := signature.ZeroHash
	if parentBlock.Header.Number > 0 {
		parentHash = parentBlock.Hash()
	}

	return Block{
		Header: BlockHeader{
			ParentHash:   parentHash,
			MinerAddress: minerAddress,
			Difficulty:   difficulty,
			Number:       parentBlock.Header.Number + 1,
			TimeStamp:    uint64(time.Now().UTC().Unix()),
		},
		Transactions: trans,
	}
}

// Hash returns the unique hash for the Block.
func (b Block) Hash() string {
	if b.Header.Number == 0 {
		return signature.ZeroHash
	}

	return signature.Hash(b)
}

// =============================================================================

// BlockFS represents what is written to the DB file.
type BlockFS struct {
	Hash  string
	Block Block
}
