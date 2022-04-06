// Copyright 2017 Cameron Bergoon
// https://github.com/cbergoon/merkletree
// Licensed under the MIT License, see LICENCE file for details.

package merkle_test

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"hash"
	"testing"

	"github.com/ardanlabs/blockchain/foundation/blockchain/merkle"
)

// TestContent implements the Content interface provided by merkletree and
// represents the content stored in the tree.
type TestMD5Content struct {
	x string
}

// Hash hashes the values of a TestContent.
func (t TestMD5Content) Hash() ([]byte, error) {
	h := md5.New()
	if _, err := h.Write([]byte(t.x)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// Equals tests for equality of two Contents.
func (t TestMD5Content) Equals(other TestMD5Content) (bool, error) {
	return t.x == other.x, nil
}

// =============================================================================

func TestNewTreeMD5Content(t *testing.T) {
	for i := 0; i < len(tableMD5Content); i++ {
		if !tableMD5Content[i].defaultHashStrategy {
			continue
		}
		tree, err := merkle.NewTree(tableMD5Content[i].contents)
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableMD5Content[i].testCaseId, err)
		}
		if !bytes.Equal(tree.MerkleRoot, tableMD5Content[i].expectedHash) {
			t.Errorf("[case:%d] error: expected hash equal to %v got %v", tableMD5Content[i].testCaseId, tableMD5Content[i].expectedHash, tree.MerkleRoot)
		}
	}
}

func TestNewTreeWithHashingStrategyMD5Content(t *testing.T) {
	for i := 0; i < len(tableMD5Content); i++ {
		tree, err := merkle.NewTree(tableMD5Content[i].contents, merkle.WithHashStrategy[TestMD5Content](tableMD5Content[i].hashStrategy))
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableMD5Content[i].testCaseId, err)
		}
		if !bytes.Equal(tree.MerkleRoot, tableMD5Content[i].expectedHash) {
			t.Errorf("[case:%d] error: expected hash equal to %v got %v", tableMD5Content[i].testCaseId, tableMD5Content[i].expectedHash, tree.MerkleRoot)
		}
	}
}

func TestMerkleTreeMD5Content_MerkleRoot(t *testing.T) {
	for i := 0; i < len(tableMD5Content); i++ {
		var tree *merkle.Tree[TestMD5Content]
		var err error
		if tableMD5Content[i].defaultHashStrategy {
			tree, err = merkle.NewTree(tableMD5Content[i].contents)
		} else {
			tree, err = merkle.NewTree(tableMD5Content[i].contents, merkle.WithHashStrategy[TestMD5Content](tableMD5Content[i].hashStrategy))
		}
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableMD5Content[i].testCaseId, err)
		}
		if !bytes.Equal(tree.MerkleRoot, tableMD5Content[i].expectedHash) {
			t.Errorf("[case:%d] error: expected hash equal to %v got %v", tableMD5Content[i].testCaseId, tableMD5Content[i].expectedHash, tree.MerkleRoot)
		}
	}
}

func TestMerkleTreeMD5Content_RebuildTree(t *testing.T) {
	for i := 0; i < len(tableMD5Content); i++ {
		var tree *merkle.Tree[TestMD5Content]
		var err error
		if tableMD5Content[i].defaultHashStrategy {
			tree, err = merkle.NewTree(tableMD5Content[i].contents)
		} else {
			tree, err = merkle.NewTree(tableMD5Content[i].contents, merkle.WithHashStrategy[TestMD5Content](tableMD5Content[i].hashStrategy))
		}
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableMD5Content[i].testCaseId, err)
		}
		err = tree.RebuildTree()
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error:  %v", tableMD5Content[i].testCaseId, err)
		}
		if !bytes.Equal(tree.MerkleRoot, tableMD5Content[i].expectedHash) {
			t.Errorf("[case:%d] error: expected hash equal to %v got %v", tableMD5Content[i].testCaseId, tableMD5Content[i].expectedHash, tree.MerkleRoot)
		}
	}
}

func TestMerkleTreeMD5Content_RebuildTreeWith(t *testing.T) {
	for i := 0; i < len(tableMD5Content)-1; i++ {
		if tableMD5Content[i].hashStrategyName != tableMD5Content[i+1].hashStrategyName {
			continue
		}
		var tree *merkle.Tree[TestMD5Content]
		var err error
		if tableMD5Content[i].defaultHashStrategy {
			tree, err = merkle.NewTree(tableMD5Content[i].contents)
		} else {
			tree, err = merkle.NewTree(tableMD5Content[i].contents, merkle.WithHashStrategy[TestMD5Content](tableMD5Content[i].hashStrategy))
		}
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableMD5Content[i].testCaseId, err)
		}
		err = tree.RebuildTreeWith(tableMD5Content[i+1].contents)
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableMD5Content[i].testCaseId, err)
		}
		if !bytes.Equal(tree.MerkleRoot, tableMD5Content[i+1].expectedHash) {
			t.Errorf("[case:%d] error: expected hash equal to %v got %v", tableMD5Content[i].testCaseId, tableMD5Content[i+1].expectedHash, tree.MerkleRoot)
		}
	}
}

func TestMerkleTreeMD5Content_VerifyTree(t *testing.T) {
	for i := 0; i < len(tableMD5Content); i++ {
		var tree *merkle.Tree[TestMD5Content]
		var err error
		if tableMD5Content[i].defaultHashStrategy {
			tree, err = merkle.NewTree(tableMD5Content[i].contents)
		} else {
			tree, err = merkle.NewTree(tableMD5Content[i].contents, merkle.WithHashStrategy[TestMD5Content](tableMD5Content[i].hashStrategy))
		}
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableMD5Content[i].testCaseId, err)
		}
		v1, err := tree.VerifyTree()
		if err != nil {
			t.Fatal(err)
		}
		if v1 != true {
			t.Errorf("[case:%d] error: expected tree to be valid", tableMD5Content[i].testCaseId)
		}
		tree.Root.Hash = []byte{1}
		tree.MerkleRoot = []byte{1}
		v2, err := tree.VerifyTree()
		if err != nil {
			t.Fatal(err)
		}
		if v2 != false {
			t.Errorf("[case:%d] error: expected tree to be invalid", tableMD5Content[i].testCaseId)
		}
	}
}

func TestMerkleTreeMD5Content_VerifyContent(t *testing.T) {
	for i := 0; i < len(tableMD5Content); i++ {
		var tree *merkle.Tree[TestMD5Content]
		var err error
		if tableMD5Content[i].defaultHashStrategy {
			tree, err = merkle.NewTree(tableMD5Content[i].contents)
		} else {
			tree, err = merkle.NewTree(tableMD5Content[i].contents, merkle.WithHashStrategy[TestMD5Content](tableMD5Content[i].hashStrategy))
		}
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableMD5Content[i].testCaseId, err)
		}
		if len(tableMD5Content[i].contents) > 0 {
			v, err := tree.VerifyContent(tableMD5Content[i].contents[0])
			if err != nil {
				t.Fatal(err)
			}
			if !v {
				t.Errorf("[case:%d] error: expected valid content", tableMD5Content[i].testCaseId)
			}
		}
		if len(tableMD5Content[i].contents) > 1 {
			v, err := tree.VerifyContent(tableMD5Content[i].contents[1])
			if err != nil {
				t.Fatal(err)
			}
			if !v {
				t.Errorf("[case:%d] error: expected valid content", tableMD5Content[i].testCaseId)
			}
		}
		if len(tableMD5Content[i].contents) > 2 {
			v, err := tree.VerifyContent(tableMD5Content[i].contents[2])
			if err != nil {
				t.Fatal(err)
			}
			if !v {
				t.Errorf("[case:%d] error: expected valid content", tableMD5Content[i].testCaseId)
			}
		}
		if len(tableMD5Content[i].contents) > 0 {
			tree.Root.Hash = []byte{1}
			tree.MerkleRoot = []byte{1}
			v, err := tree.VerifyContent(tableMD5Content[i].contents[0])
			if err != nil {
				t.Fatal(err)
			}
			if v {
				t.Errorf("[case:%d] error: expected invalid content", tableMD5Content[i].testCaseId)
			}
			if err := tree.RebuildTree(); err != nil {
				t.Fatal(err)
			}
		}
		v, err := tree.VerifyContent(tableMD5Content[i].notInContents)
		if err != nil {
			t.Fatal(err)
		}
		if v {
			t.Errorf("[case:%d] error: expected invalid content", tableMD5Content[i].testCaseId)
		}
	}
}

func TestMerkleTreeMD5Content_String(t *testing.T) {
	for i := 0; i < len(tableMD5Content); i++ {
		var tree *merkle.Tree[TestMD5Content]
		var err error
		if tableMD5Content[i].defaultHashStrategy {
			tree, err = merkle.NewTree(tableMD5Content[i].contents)
		} else {
			tree, err = merkle.NewTree(tableMD5Content[i].contents, merkle.WithHashStrategy[TestMD5Content](tableMD5Content[i].hashStrategy))
		}
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableMD5Content[i].testCaseId, err)
		}
		if tree.String() == "" {
			t.Errorf("[case:%d] error: expected not empty string", tableMD5Content[i].testCaseId)
		}
	}
}

func TestMerkleTreeMD5Content_MerklePath(t *testing.T) {
	for i := 0; i < len(tableMD5Content); i++ {
		var tree *merkle.Tree[TestMD5Content]
		var err error
		if tableMD5Content[i].defaultHashStrategy {
			tree, err = merkle.NewTree(tableMD5Content[i].contents)
		} else {
			tree, err = merkle.NewTree(tableMD5Content[i].contents, merkle.WithHashStrategy[TestMD5Content](tableMD5Content[i].hashStrategy))
		}
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableMD5Content[i].testCaseId, err)
		}
		for j := 0; j < len(tableMD5Content[i].contents); j++ {
			merklePath, index, _ := tree.GetMerklePath(tableMD5Content[i].contents[j])

			hash, err := tree.Leafs[j].CalculateNodeHash()
			if err != nil {
				t.Errorf("[case:%d] error: calculateNodeHash error: %v", tableMD5Content[i].testCaseId, err)
			}
			h := sha256.New()
			for k := 0; k < len(merklePath); k++ {
				if index[k] == 1 {
					hash = append(hash, merklePath[k]...)
				} else {
					hash = append(merklePath[k], hash...)
				}
				if _, err := h.Write(hash); err != nil {
					t.Errorf("[case:%d] error: Write error: %v", tableMD5Content[i].testCaseId, err)
				}
				hash, err = calHash(hash, tableMD5Content[i].hashStrategy)
				if err != nil {
					t.Errorf("[case:%d] error: calHash error: %v", tableMD5Content[i].testCaseId, err)
				}
			}
			if !bytes.Equal(tree.MerkleRoot, hash) {
				t.Errorf("[case:%d] error: expected hash equal to %v got %v", tableMD5Content[i].testCaseId, hash, tree.MerkleRoot)
			}
		}
	}
}

// =============================================================================

var tableMD5Content = []struct {
	testCaseId          int
	hashStrategy        func() hash.Hash
	hashStrategyName    string
	defaultHashStrategy bool
	contents            []TestMD5Content
	expectedHash        []byte
	notInContents       TestMD5Content
}{
	{
		testCaseId:          5,
		hashStrategy:        md5.New,
		hashStrategyName:    "md5",
		defaultHashStrategy: false,
		contents: []TestMD5Content{
			{x: "Hello"}, {x: "Hi"}, {x: "Hey"}, {x: "Hola"},
		},
		notInContents: TestMD5Content{x: "NotInTestTable"},
		expectedHash:  []byte{217, 158, 206, 52, 191, 78, 253, 233, 25, 55, 69, 142, 254, 45, 127, 144},
	},
	{
		testCaseId:          6,
		hashStrategy:        md5.New,
		hashStrategyName:    "md5",
		defaultHashStrategy: false,
		contents: []TestMD5Content{
			{x: "Hello"}, {x: "Hi"}, {x: "Hey"},
		},
		notInContents: TestMD5Content{x: "NotInTestTable"},
		expectedHash:  []byte{145, 228, 171, 107, 94, 219, 221, 171, 7, 195, 206, 128, 148, 98, 59, 76},
	},
	{
		testCaseId:          7,
		hashStrategy:        md5.New,
		hashStrategyName:    "md5",
		defaultHashStrategy: false,
		contents: []TestMD5Content{
			{x: "Hello"}, {x: "Hi"}, {x: "Hey"}, {x: "Greetings"}, {x: "Hola"},
		},
		notInContents: TestMD5Content{x: "NotInTestTable"},
		expectedHash:  []byte{167, 200, 229, 62, 194, 247, 117, 12, 206, 194, 90, 235, 70, 14, 100, 100},
	},
	{
		testCaseId:          8,
		hashStrategy:        md5.New,
		hashStrategyName:    "md5",
		defaultHashStrategy: false,
		contents: []TestMD5Content{
			{x: "123"}, {x: "234"}, {x: "345"}, {x: "456"}, {x: "1123"}, {x: "2234"}, {x: "3345"}, {x: "4456"},
		},
		notInContents: TestMD5Content{x: "NotInTestTable"},
		expectedHash:  []byte{8, 36, 33, 50, 204, 197, 82, 81, 207, 74, 6, 60, 162, 209, 168, 21},
	},
	{
		testCaseId:          9,
		hashStrategy:        md5.New,
		hashStrategyName:    "md5",
		defaultHashStrategy: false,
		contents: []TestMD5Content{
			{x: "123"}, {x: "234"}, {x: "345"}, {x: "456"}, {x: "1123"}, {x: "2234"}, {x: "3345"}, {x: "4456"}, {x: "5567"},
		},
		notInContents: TestMD5Content{x: "NotInTestTable"},
		expectedHash:  []byte{158, 85, 181, 191, 25, 250, 251, 71, 215, 22, 68, 68, 11, 198, 244, 148},
	},
}

// =============================================================================

func calHash(hash []byte, hashStrategy func() hash.Hash) ([]byte, error) {
	h := hashStrategy()
	if _, err := h.Write(hash); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
