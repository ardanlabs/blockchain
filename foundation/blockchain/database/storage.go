package database

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strconv"
	"sync"
)

type Storage interface {
	Write(block BlockFS) error
	GetBlock(num uint64) (*BlockFS, error)
	Foreach() Iterator
	Close() error
	Reset() error
}

type Iterator interface {
	Next() (*BlockFS, error)
	Done() bool
}

type JSONIterator struct {
	s       *JSONStorage
	done    bool
	scanner *bufio.Scanner
}

func (i *JSONIterator) Next() (*BlockFS, error) {
	if i.done {
		return nil, errors.New("done")
	}
	if !i.scanner.Scan() {
		i.done = true
		return nil, errors.New("done")
	}
	if err := i.scanner.Err(); err != nil {
		return nil, err
	}
	var blockFS BlockFS
	if err := json.Unmarshal(i.scanner.Bytes(), &blockFS); err != nil {
		return nil, err
	}
	return &blockFS, nil
}

func (i *JSONIterator) Done() bool {
	return i.done
}

type JSONStorage struct {
	dbPath string
	mu     sync.RWMutex
	dbFile *os.File
}

func NewJSONStorage(dbPath string) (*JSONStorage, error) {
	var dbFile *os.File
	var err error
	dbFile, err = os.OpenFile(dbPath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	if errors.Is(err, fs.ErrNotExist) {
		dbFile, err = os.OpenFile(dbPath, os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			return nil, err
		}
	}
	return &JSONStorage{
		dbFile: dbFile,
		dbPath: dbPath,
	}, nil

}

func (s *JSONStorage) Write(block BlockFS) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Marshal the block for writing to disk.
	blockFSJson, err := json.Marshal(block)
	if err != nil {
		return err
	}

	// Write the new block to the chain on disk.
	if _, err := s.dbFile.Write(append(blockFSJson, '\n')); err != nil {
		return err
	}
	return nil
}

func (s *JSONStorage) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Close and remove the current file.
	s.dbFile.Close()
	if err := os.Remove(s.dbPath); err != nil {
		return err
	}

	// Open a new blockchain database file with create.
	dbFile, err := os.OpenFile(s.dbPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	s.dbFile = dbFile

	return nil
}

func (s *JSONStorage) GetBlock(num uint64) (*BlockFS, error) {
	return nil, nil
}

func (s *JSONStorage) Foreach() Iterator {
	scanner := bufio.NewScanner(s.dbFile)
	return &JSONIterator{
		s:       s,
		scanner: scanner,
	}
}

func (s *JSONStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dbFile.Close()
	return nil
}

type FilesStorage struct {
	dbPath string
}

func NewFilesStorage(dbPath string) *FilesStorage {
	return &FilesStorage{dbPath: dbPath}
}

func (s *FilesStorage) getPath(blockNum uint64) string {
	name := strconv.FormatUint(blockNum, 10)
	return path.Join(s.dbPath, fmt.Sprintf("%s.json", name))
}

func (s *FilesStorage) Close() error {
	return nil
}

// Write adds a new block to the chain.
func (s *FilesStorage) Write(block BlockFS) error {
	// Marshal the block for writing to disk.
	data, err := json.MarshalIndent(block, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.OpenFile(s.getPath(block.Block.Number), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	// Write the new block to the chain on disk.
	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}

func (s *FilesStorage) GetBlock(num uint64) (*BlockFS, error) {
	f, err := os.OpenFile(s.getPath(num), os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	block := &BlockFS{}
	err = decoder.Decode(block)
	return block, err
}

func (s *FilesStorage) Foreach() Iterator {
	return &FilesIterator{s: s}
}

func (s *FilesStorage) Reset() error {
	return nil
}

type FilesIterator struct {
	current uint64
	s       *FilesStorage
	done    bool
}

func (i *FilesIterator) Next() (*BlockFS, error) {
	if i.done {
		return nil, errors.New("done")
	}
	i.current++
	block, err := i.s.GetBlock(i.current)
	if errors.Is(err, fs.ErrNotExist) {
		i.done = true
	}
	return block, err
}

func (i *FilesIterator) Done() bool {
	return i.done
}
