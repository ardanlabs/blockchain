package blockchain

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// storage manages reading and writing of blocks to storage.
type storage struct {
	dbPath string
	dbFile *os.File
	mu     sync.Mutex
}

// newStorage provides access to blockchain storage.
func newStorage(dbPath string) (*storage, error) {

	// Open the blockchain database file with append.
	dbFile, err := os.OpenFile(dbPath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	strg := storage{
		dbPath: dbPath,
		dbFile: dbFile,
	}

	return &strg, nil
}

// Close cleanly releases the storage area.
func (str *storage) close() {
	str.mu.Lock()
	defer str.mu.Unlock()

	str.dbFile.Close()
}

// =============================================================================

// reset create a new storage area for the blockchain to start new.
func (str *storage) reset() error {
	str.mu.Lock()
	defer str.mu.Unlock()

	// Close and remove the current file.
	str.dbFile.Close()
	os.Remove(str.dbPath)

	// Open a new blockchain database file with create.
	dbFile, err := os.OpenFile(str.dbPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}

	str.dbFile = dbFile

	return nil
}

// write adds a new block to the chain.
func (str *storage) write(block blockFS) error {
	str.mu.Lock()
	defer str.mu.Unlock()

	// Marshal the block for writing to disk.
	blockFSJson, err := json.Marshal(block)
	if err != nil {
		return err
	}

	// Write the new block to the chain on disk.
	if _, err := str.dbFile.Write(append(blockFSJson, '\n')); err != nil {
		return err
	}

	return nil
}

// =============================================================================

// readAllBlocks loads all existing blocks from storage into memory. In a real
// world situation this would require a lot of memory.
func (str *storage) readAllBlocks() ([]Block, error) {
	dbFile, err := os.Open(str.dbPath)
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

		var blockFS blockFS
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
