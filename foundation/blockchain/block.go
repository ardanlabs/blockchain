package blockchain

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

// zeroHash represents a hash code of zeros.
const zeroHash string = "00000000000000000000000000000000"

// =============================================================================

// BlockHeader represents common information required for each block.
type BlockHeader struct {
	ParentHash  string `json:"parent_hash"` // Hash of the previous block in the chain.
	Beneficiary string `json:"beneficiary"` // Address who receives the reward and gas fee.
	Difficulty  int    `json:"difficulty"`  // Number of 0's needed to solve the hash solution.
	Number      uint64 `json:"number"`      // Block number in the chain.
	TotalTip    uint   `json:"total_tip"`   // Total tip paid by all senders as an incentive.
	TotalGas    uint   `json:"total_gas"`   // Total gas fee to recover computation costs paid by the sender.
	TimeStamp   uint64 `json:"timestamp"`   // Time the block was mined.
	Nonce       uint64 `json:"nonce"`       // Value identified to solve the hash solution.
}

// Block represents a group of transactions batched together.
type Block struct {
	Header       BlockHeader `json:"header"`
	Transactions []BlockTx   `json:"txs"`
}

// newBlock constructs a new BlockFS for persisting.
func newBlock(beneficiary string, difficulty int, transPerBlock int, parentBlock Block, txMempool txMempool) Block {
	parentHash := zeroHash
	if parentBlock.Header.Number > 0 {
		parentHash = parentBlock.Hash()
	}

	// Copy the best transactions from the mempool for this new block.
	cpy := txMempool.copyBestByTip(transPerBlock)

	return Block{
		Header: BlockHeader{
			ParentHash:  parentHash,
			Beneficiary: beneficiary,
			Difficulty:  difficulty,
			Number:      parentBlock.Header.Number + 1,
			TimeStamp:   uint64(time.Now().Unix()),
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

	data, err := json.Marshal(b)
	if err != nil {
		return zeroHash
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// Sign uses the specified private key to sign the user transaction.
func (b Block) Sign(privateKey *ecdsa.PrivateKey) (SignedBlock, error) {

	// Prepare the transaction for signing.
	tran, err := b.HashWithArdanStamp()
	if err != nil {
		return SignedBlock{}, err
	}

	// Sign the hash with the private key to produce a signature.
	sig, err := crypto.Sign(tran, privateKey)
	if err != nil {
		return SignedBlock{}, err
	}

	// Convert the 65 byte signature into the [R|S|V] format.
	v, r, s := toSignatureValues(sig)

	// Construct the signed block.
	signedBlock := SignedBlock{
		Block: b,
		V:     v,
		R:     r,
		S:     s,
	}

	return signedBlock, nil
}

// HashWithArdanStamp returns a hash of 32 bytes that represents this user
// transaction with the Ardan stamp embedded into the final hash.
func (b Block) HashWithArdanStamp() ([]byte, error) {

	// Marshal and hash the user data to validate the signature.
	txData, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}

	// Hash the transaction data into a 32 byte array. This will provide
	// a data length consistency with all transactions.
	txHash := crypto.Keccak256Hash(txData)

	// Convert the stamp into a slice of bytes. This stamp is
	// used so signatures we produce when signing transactions
	// are always unique to the Ardan blockchain.
	stamp := []byte("\x19Ardan Signed Message:\n32")

	// Hash the stamp and txHash together in a final 32 byte array
	// that represents the transaction data.
	tran := crypto.Keccak256Hash(stamp, txHash.Bytes())

	return tran.Bytes(), nil
}

// =============================================================================

// SignedBlock is a signed version of the block.
type SignedBlock struct {
	Block
	V *big.Int `json:"v"` // Recovery identifier, either 29 or 30 with ardanID.
	R *big.Int `json:"r"` // First coordinate of the ECDSA signature.
	S *big.Int `json:"s"` // Second coordinate of the ECDSA signature.
}

// VerifySignature verifies the signature conforms to our standards and
// is associated with the data claimed to be signed.
func (b SignedBlock) VerifySignature() error {

	// Check the recovery id is either 0 or 1.
	v := b.V.Uint64() - ardanID
	if v != 0 && v != 1 {
		return errors.New("invalid recovery id")
	}

	// Check the signature values are valid.
	if !crypto.ValidateSignatureValues(byte(v), b.R, b.S, false) {
		return errors.New("invalid signature values")
	}

	// Prepare the transaction for recovery and validation.
	tran, err := b.HashWithArdanStamp()
	if err != nil {
		return err
	}

	// Convert the [R|S|V] format into the original 65 bytes.
	sig := toSignatureBytes(b.V, b.R, b.S)

	// Capture the uncompressed public key associated with this signature.
	sigPublicKey, err := crypto.Ecrecover(tran, sig)
	if err != nil {
		return fmt.Errorf("ecrecover, %w", err)
	}

	// Check that the given public key created the signature over the data.
	rs := sig[:crypto.RecoveryIDOffset]
	if !crypto.VerifySignature(sigPublicKey, tran, rs) {
		return errors.New("invalid signature")
	}

	return nil
}

// Signature returns the signature as a string.
func (b SignedBlock) Signature() string {
	return "0x" + hex.EncodeToString(toSignatureBytesForDisplay(b.V, b.R, b.S))
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

		blocks = append(blocks, blockFS.SignedBlock)
		blockNum++
	}

	return blocks, nil
}
