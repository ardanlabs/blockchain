// Package memory implements the ability to read and write blocks to memory
// using a slice.
package memory

import (
	"errors"
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
)

// Memory represents the serialization implementation for reading and storing
// blocks in memory using a slice. This implements the database.Storage
// interface.
type Memory struct {
	mu     sync.RWMutex
	blocks []database.BlockData
}

// New constructs an Memory value for use.
func New() (*Memory, error) {
	return &Memory{}, nil
}

// Close in this implementation has nothing to do since everything
// is in memory.
func (m *Memory) Close() error {
	return nil
}

// Write takes the specified database blocks and stores it in memory.
func (m *Memory) Write(blockData database.BlockData) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	l := len(m.blocks)
	if l+1 != int(blockData.Header.Number) {
		return errors.New("block is out of order")
	}

	m.blocks = append(m.blocks, blockData)

	return nil
}

// GetBlock searches the blockchain to locate and return the contents of
// the specified block by number.
func (m *Memory) GetBlock(num uint64) (database.BlockData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	l := uint64(len(m.blocks))
	if l == 0 || num >= l {
		return database.BlockData{}, errors.New("block does not exist")
	}

	return m.blocks[num], nil
}

// ForEach returns an iterator to walk through all the blocks
// starting with block number 1.
func (m *Memory) ForEach() database.Iterator {
	return &memoryIterator{storage: m}
}

// Reset will clear out the blockchain on disk.
func (m *Memory) Reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.blocks = []database.BlockData{}
	return nil
}

// =============================================================================

// memoryIterator represents the iteration implementation for walking
// through and reading blocks on disk. This implements the database
// Iterator interface.
type memoryIterator struct {
	storage *Memory // Access to the storage API.
	current uint64  // Currenet block number being iterated over.
	eoc     bool    // Represents the iterator is at the end of the chain.
}

// Next retrieves the next block from disk.
func (mi *memoryIterator) Next() (database.BlockData, error) {
	if mi.eoc {
		return database.BlockData{}, errors.New("end of chain")
	}

	blockData, err := mi.storage.GetBlock(mi.current)
	if err != nil {
		mi.eoc = true
	}

	mi.current++

	return blockData, err
}

// Done returns the end of chain value.
func (mi *memoryIterator) Done() bool {
	return mi.eoc
}
