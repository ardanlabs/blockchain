// Package storage handles all the lower level support for maintaining the
// blockchain on disk.
package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
)

// Storage manages reading and writing of blocks to storage.
type Storage struct {
	dbPath string
	dbFile *os.File
	mu     sync.Mutex
}

// New provides access to blockchain storage.
func New(dbPath string) (*Storage, error) {

	// Open the blockchain database file with append.
	dbFile, err := os.OpenFile(dbPath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	strg := Storage{
		dbPath: dbPath,
		dbFile: dbFile,
	}

	return &strg, nil
}

// Close cleanly releases the storage area.
func (str *Storage) Close() {
	str.mu.Lock()
	defer str.mu.Unlock()

	str.dbFile.Close()
}

// Reset create a new storage area for the blockchain to start new.
func (str *Storage) Reset() error {
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

// Write adds a new block to the chain.
func (str *Storage) Write(block BlockFS) error {
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

// ReadAllBlocks loads all existing blocks from storage into memory. In a real
// world situation this would require a lot of memory.
func (str *Storage) ReadAllBlocks(evHandler func(v string, args ...any), validate bool) ([]Block, error) {
	dbFile, err := os.Open(str.dbPath)
	if err != nil {
		return nil, err
	}
	defer dbFile.Close()

	var blocks []Block
	var latestBlock Block
	scanner := bufio.NewScanner(dbFile)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		var blockFS BlockFS
		if err := json.Unmarshal(scanner.Bytes(), &blockFS); err != nil {
			return nil, err
		}

		block, err := ToBlock(blockFS)
		if err != nil {
			return nil, err
		}

		if validate {
			if err := block.ValidateBlock(latestBlock, evHandler); err != nil {
				return nil, err
			}
		}

		blocks = append(blocks, block)
		latestBlock = block
	}

	return blocks, nil
}
