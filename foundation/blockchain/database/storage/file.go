package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strconv"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
)

// Ardan represents the storage implementation for reading and storing blocks
// in their own separate files on disk. This implements the database.Storage
// interface.
type Ardan struct {
	dbPath string
}

// NewArdan constructs an Ardan value for use.
func NewArdan(dbPath string) (*Ardan, error) {
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return nil, err
	}

	return &Ardan{dbPath: dbPath}, nil
}

// Close in this implementation has nothing to do since a new file is
// written to disk for each now block and then immediately closed.
func (ard *Ardan) Close() error {
	return nil
}

// Write takes the specified database blocks and stores it on disk in a
// file labeled with the block number.
func (ard *Ardan) Write(dbBlock database.Block) error {

	// Need to convert the block to the storage format.
	block := NewBlock(dbBlock)

	// Marshal the block for writing to disk in a more human readable format.
	data, err := json.MarshalIndent(block, "", "  ")
	if err != nil {
		return err
	}

	// Create a new file for this block and name it based on the block number.
	f, err := os.OpenFile(ard.getPath(block.Header.Number), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write the new block to disk.
	if _, err := f.Write(data); err != nil {
		return err
	}

	return nil
}

// GetBlock searches the blockchain on disk to locate and return the
// contents of the specified block by number.
func (ard *Ardan) GetBlock(num uint64) (database.Block, error) {

	// Open the block file for the specified number.
	f, err := os.OpenFile(ard.getPath(num), os.O_RDONLY, 0600)
	if err != nil {
		return database.Block{}, err
	}
	defer f.Close()

	// Decode the contents of the block.
	var block Block
	if err := json.NewDecoder(f).Decode(&block); err != nil {
		return database.Block{}, err
	}

	// Return the block as a database block.
	return ToDatabaseBlock(block)
}

// ForEach returns an iterator to walk through all the blocks on
// disk starting with block number 1.
func (ard *Ardan) ForEach() database.Iterator {
	return &ArdanIterator{storage: ard}
}

// Reset will clear out the blockchain on disk.
func (ard *Ardan) Reset() error {
	return nil
}

// getPath forms the path to the specified block.
func (ard *Ardan) getPath(blockNum uint64) string {
	name := strconv.FormatUint(blockNum, 10)
	return path.Join(ard.dbPath, fmt.Sprintf("%s.json", name))
}

// ArdanIterator represents the iteration implementation for walking
// through and reading blocks on disk. This implements the database
// Iterator interface.
type ArdanIterator struct {
	storage *Ardan // Access to the Ardan storage API.
	current uint64 // Currenet block number being iterated over.
	eoc     bool   // Represents the iterator is at the end of the chain.
}

// Next retrieves the next block from disk.
func (ai *ArdanIterator) Next() (database.Block, error) {
	if ai.eoc {
		return database.Block{}, errors.New("end of chain")
	}

	ai.current++
	block, err := ai.storage.GetBlock(ai.current)
	if errors.Is(err, fs.ErrNotExist) {
		ai.eoc = true
	}

	return block, nil
}

// Done returns the end of chain value.
func (ai *ArdanIterator) Done() bool {
	return ai.eoc
}
