package blockchain

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"time"
)

// zeroHash represents a hash code of zeros.
const zeroHash string = "00000000000000000000000000000000"

// =============================================================================

// BlockHeader represents common information required for each block.
type BlockHeader struct {
	ParentHash string `json:"parent_hash"` // Hash of the previous block in the chain.
	Difficulty int    `json:"difficulty"`  // Number of 0's needed to solve the hash solution.
	Number     uint64 `json:"number"`      // Block number in the chain.
	TotalTip   uint   `json:"total_tip"`   // Total tip paid by all senders as an incentive.
	TotalGas   uint   `json:"total_gas"`   // Total gas fee to recover computation costs paid by the sender.
	TimeStamp  uint64 `json:"timestamp"`   // Time the block was mined.
	Nonce      uint64 `json:"nonce"`       // Value identified to solve the hash solution.
}

// Block represents a group of transactions batched together.
type Block struct {
	Header       BlockHeader `json:"header"`
	Transactions []BlockTx   `json:"txs"`
}

// newBlock constructs a new BlockFS for persisting.
func newBlock(difficulty int, transPerBlock int, parentBlock SignedBlock, txMempool *txMempool) Block {
	parentHash := zeroHash
	if parentBlock.Header.Number > 0 {
		parentHash = parentBlock.Hash()
	}

	// Copy the best transactions from the mempool for this new block.
	cpy := txMempool.copyBestByTip(transPerBlock)

	return Block{
		Header: BlockHeader{
			ParentHash: parentHash,
			Difficulty: difficulty,
			Number:     parentBlock.Header.Number + 1,
			TimeStamp:  uint64(time.Now().Unix()),
		},
		Transactions: cpy,
	}
}

// Sign uses the specified private key to sign the user transaction.
func (b Block) Sign(privateKey *ecdsa.PrivateKey) (SignedBlock, error) {

	// Sign the hash with the private key to produce a signature.
	v, r, s, err := sign(b, privateKey)
	if err != nil {
		return SignedBlock{}, err
	}

	// Construct the signed block.
	signedBlock := SignedBlock{
		Block: b,
		V:     v,
		R:     r,
		S:     s,
	}

	return signedBlock, nil
}

// =============================================================================

// SignedBlock is a signed version of the block.
type SignedBlock struct {
	Block
	V *big.Int `json:"v"` // Recovery identifier, either 29 or 30 with ardanID.
	R *big.Int `json:"r"` // First coordinate of the ECDSA signature.
	S *big.Int `json:"s"` // Second coordinate of the ECDSA signature.
}

// Hash returns the unique hash for the Block.
func (b Block) Hash() string {
	if b.Header.Number == 0 {
		return zeroHash
	}

	return hash(b)
}

// VerifySignature verifies the signature conforms to our standards and
// is associated with the data claimed to be signed.
func (b SignedBlock) VerifySignature() error {
	return verifySignature(b.Block, b.V, b.R, b.S)
}

// FromAddress extracts the address for the account that signed the transaction.
func (b SignedBlock) FromAddress() (string, error) {
	return fromAddress(b.Block, b.V, b.R, b.S)
}

// Signature returns the signature as a string.
func (b SignedBlock) SignatureString() string {
	return signatureString(b.V, b.R, b.S)
}

// =============================================================================

// blockFS represents what is written to the DB file.
type blockFS struct {
	Hash        string
	SignedBlock SignedBlock
}

// performPOW does the work of mining to find a valid hash for a specified
// block and returns a BlockFS ready to be written to disk.
func performPOW(ctx context.Context, difficulty int, b Block, privateKey *ecdsa.PrivateKey, ev EventHandler) (blockFS, time.Duration, error) {
	ev("worker: runMiningOperation: MINING: POW: started")
	defer ev("worker: runMiningOperation: MINING: POW: completed")

	for _, tx := range b.Transactions {
		ev("worker: runMiningOperation: MINING: POW: tx[%s]", tx.Hash())
	}

	t := time.Now()

	// Choose a random starting point for the nonce.
	nBig, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return blockFS{}, time.Since(t), ctx.Err()
	}
	b.Header.Nonce = nBig.Uint64()

	var attempts uint64
	for {
		attempts++
		if attempts%1_000_000 == 0 {
			ev("worker: runMiningOperation: MINING: POW: attempts[%d]", attempts)
		}

		// Did we timeout trying to solve the problem.
		if ctx.Err() != nil {
			ev("worker: runMiningOperation: MINING: POW: CANCELLED")
			return blockFS{}, time.Since(t), ctx.Err()
		}

		// Hash the block and check if we have solved the puzzle.
		hash := b.Hash()
		if !isHashSolved(difficulty, hash) {

			// I may want to track these nonce's to make sure I
			// don't try the same one twice.
			b.Header.Nonce++
			continue
		}

		// Did we timeout trying to solve the problem.
		if ctx.Err() != nil {
			ev("worker: runMiningOperation: MINING: POW: CANCELLED")
			return blockFS{}, time.Since(t), ctx.Err()
		}

		ev("worker: runMiningOperation: MINING: POW: SOLVED: prevBlk[%s]: newBlk[%s]", b.Header.ParentHash, b.Hash())
		ev("worker: runMiningOperation: MINING: POW: attempts[%d]", attempts)

		// Sign the block for integrity and to let others know we
		// get the reward and fees.
		signedBlock, err := b.Sign(privateKey)
		if err != nil {
			return blockFS{}, time.Since(t), ctx.Err()
		}

		// We found a solution to the POW.
		bfs := blockFS{
			Hash:        hash,
			SignedBlock: signedBlock,
		}
		return bfs, time.Since(t), nil
	}
}

// isHashSolved checks the hash to make sure it complies with
// the POW rules. We need to match a difficulty number of 0's.
func isHashSolved(difficulty int, hash string) bool {
	const match = "00000000000000000"

	if len(hash) != 64 {
		return false
	}

	return hash[:difficulty] == match[:difficulty]
}

// =============================================================================

// loadBlocksFromDisk the current set of blocks/transactions. In a real
// world situation this would require a lot of memory.
func loadBlocksFromDisk(dbPath string) ([]SignedBlock, error) {
	dbFile, err := os.Open(dbPath)
	if err != nil {
		return nil, err
	}
	defer dbFile.Close()

	var blockNum int
	var blocks []SignedBlock
	scanner := bufio.NewScanner(dbFile)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		var blockFS blockFS
		if err := json.Unmarshal(scanner.Bytes(), &blockFS); err != nil {
			return nil, err
		}

		if blockFS.SignedBlock.Hash() != blockFS.Hash {
			return nil, fmt.Errorf("block %d has been changed", blockNum)
		}

		if err := blockFS.SignedBlock.VerifySignature(); err != nil {
			return nil, fmt.Errorf("block %d has bad signature", blockNum)
		}

		blocks = append(blocks, blockFS.SignedBlock)
		blockNum++
	}

	return blocks, nil
}
