// Copyright 2017 Cameron Bergoon
// https://github.com/cbergoon/merkletree
// Licensed under the MIT License, see LICENCE file for details.

package merkle_test

import (
	"bytes"
	"crypto/sha256"
	"hash"
	"testing"

	"github.com/ardanlabs/blockchain/foundation/blockchain/merkle"
)

// Data uses the sha256 hashing algorithm for the merkle tree.
type Data struct {
	x string
}

// Hash hashes the values using sha256.
func (d Data) Hash() ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(d.x)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// Equals tests for equality of two piece of data.
func (d Data) Equals(other Data) bool {
	return d.x == other.x
}

// =============================================================================

func Test_NewTreeWithDefault(t *testing.T) {
	for i := 0; i < len(table); i++ {
		tree, err := merkle.NewTree(table[i].data)
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", table[i].testCaseID, err)
		}
		if !bytes.Equal(tree.MerkleRoot, table[i].expectedHash) {
			t.Errorf("[case:%d] error: expected hash equal to %v got %v", table[i].testCaseID, table[i].expectedHash, tree.MerkleRoot)
		}
	}
}

func Test_NewTreeWithHashingStrategy(t *testing.T) {
	for i := 0; i < len(table); i++ {
		tree, err := merkle.NewTree(table[i].data, merkle.WithHashStrategy[Data](table[i].hashStrategy))
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", table[i].testCaseID, err)
		}
		if !bytes.Equal(tree.MerkleRoot, table[i].expectedHash) {
			t.Errorf("[case:%d] error: expected hash equal to %v got %v", table[i].testCaseID, table[i].expectedHash, tree.MerkleRoot)
		}
	}
}

func Test_MerkleRoot(t *testing.T) {
	for i := 0; i < len(table); i++ {
		tree, err := merkle.NewTree(table[i].data, merkle.WithHashStrategy[Data](table[i].hashStrategy))
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", table[i].testCaseID, err)
		}
		if !bytes.Equal(tree.MerkleRoot, table[i].expectedHash) {
			t.Errorf("[case:%d] error: expected hash equal to %v got %v", table[i].testCaseID, table[i].expectedHash, tree.MerkleRoot)
		}
	}
}

func Test_RebuildTree(t *testing.T) {
	for i := 0; i < len(table); i++ {
		tree, err := merkle.NewTree(table[i].data, merkle.WithHashStrategy[Data](table[i].hashStrategy))
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", table[i].testCaseID, err)
		}
		err = tree.RebuildTree()
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error:  %v", table[i].testCaseID, err)
		}
		if !bytes.Equal(tree.MerkleRoot, table[i].expectedHash) {
			t.Errorf("[case:%d] error: expected hash equal to %v got %v", table[i].testCaseID, table[i].expectedHash, tree.MerkleRoot)
		}
	}
}

func Test_RebuildTreeWith(t *testing.T) {
	for i := 0; i < len(table)-1; i++ {
		tree, err := merkle.NewTree(table[i].data, merkle.WithHashStrategy[Data](table[i].hashStrategy))
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", table[i].testCaseID, err)
		}
		err = tree.GenerateTree(table[i+1].data)
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", table[i].testCaseID, err)
		}
		if !bytes.Equal(tree.MerkleRoot, table[i+1].expectedHash) {
			t.Errorf("[case:%d] error: expected hash equal to %v got %v", table[i].testCaseID, table[i+1].expectedHash, tree.MerkleRoot)
		}
	}
}

func Test_VerifyTree(t *testing.T) {
	for i := 0; i < len(table); i++ {
		tree, err := merkle.NewTree(table[i].data, merkle.WithHashStrategy[Data](table[i].hashStrategy))
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", table[i].testCaseID, err)
		}
		if err := tree.VerifyTree(); err != nil {
			t.Errorf("[case:%d] error: expected tree to be valid", table[i].testCaseID)
		}
		tree.Root.Hash = []byte{1}
		tree.MerkleRoot = []byte{1}
		if err := tree.VerifyTree(); err == nil {
			t.Errorf("[case:%d] error: expected tree to be invalid", table[i].testCaseID)
		}
	}
}

func Test_VerifyData(t *testing.T) {
	for i := 0; i < len(table); i++ {
		tree, err := merkle.NewTree(table[i].data, merkle.WithHashStrategy[Data](table[i].hashStrategy))
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", table[i].testCaseID, err)
		}
		if len(table[i].data) > 0 {
			if err := tree.VerifyData(table[i].data[0]); err != nil {
				t.Errorf("[case:%d] error: expected valid content", table[i].testCaseID)
			}
		}
		if len(table[i].data) > 1 {
			if err := tree.VerifyData(table[i].data[1]); err != nil {
				t.Errorf("[case:%d] error: expected valid content", table[i].testCaseID)
			}
		}
		if len(table[i].data) > 2 {
			if err := tree.VerifyData(table[i].data[2]); err != nil {
				t.Errorf("[case:%d] error: expected valid content", table[i].testCaseID)
			}
		}
		if len(table[i].data) > 0 {
			tree.Root.Hash = []byte{1}
			tree.MerkleRoot = []byte{1}
			if err := tree.VerifyData(table[i].data[0]); err == nil {
				t.Errorf("[case:%d] error: expected invalid content", table[i].testCaseID)
			}
			if err := tree.RebuildTree(); err != nil {
				t.Fatal(err)
			}
		}
		if err := tree.VerifyData(table[i].notInContents); err == nil {
			t.Errorf("[case:%d] error: expected invalid content", table[i].testCaseID)
		}
	}
}

func Test_String(t *testing.T) {
	for i := 0; i < len(table); i++ {
		tree, err := merkle.NewTree(table[i].data, merkle.WithHashStrategy[Data](table[i].hashStrategy))
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", table[i].testCaseID, err)
		}
		if tree.String() == "" {
			t.Errorf("[case:%d] error: expected not empty string", table[i].testCaseID)
		}
	}
}

func Test_MerklePath(t *testing.T) {
	for i := 0; i < len(table); i++ {
		tree, err := merkle.NewTree(table[i].data, merkle.WithHashStrategy[Data](table[i].hashStrategy))
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", table[i].testCaseID, err)
		}
		for j := 0; j < len(table[i].data); j++ {
			merkleProof, index, _ := tree.MerkleProof(table[i].data[j])

			hash, err := tree.Leafs[j].CalculateNodeHash()
			if err != nil {
				t.Errorf("[case:%d] error: calculateNodeHash error: %v", table[i].testCaseID, err)
			}
			h := sha256.New()
			for k := 0; k < len(merkleProof); k++ {
				if index[k] == 1 {
					hash = append(hash, merkleProof[k]...)
				} else {
					hash = append(merkleProof[k], hash...)
				}
				if _, err := h.Write(hash); err != nil {
					t.Errorf("[case:%d] error: Write error: %v", table[i].testCaseID, err)
				}
				hash, err = calHash(hash, table[i].hashStrategy)
				if err != nil {
					t.Errorf("[case:%d] error: calHash error: %v", table[i].testCaseID, err)
				}
			}
			if !bytes.Equal(tree.MerkleRoot, hash) {
				t.Errorf("[case:%d] error: expected hash equal to %v got %v", table[i].testCaseID, hash, tree.MerkleRoot)
			}
		}
	}
}

// =============================================================================

func calHash(hash []byte, hashStrategy func() hash.Hash) ([]byte, error) {
	h := hashStrategy()
	if _, err := h.Write(hash); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// =============================================================================

var table = []struct {
	testCaseID    int
	hashStrategy  func() hash.Hash
	data          []Data
	expectedHash  []byte
	notInContents Data
}{
	{
		testCaseID:   1,
		hashStrategy: sha256.New,
		data: []Data{
			{x: "Hello"}, {x: "Hi"}, {x: "Hey"}, {x: "Hola"},
		},
		notInContents: Data{x: "NotInTestTable"},
		expectedHash:  []byte{95, 48, 204, 128, 19, 59, 147, 148, 21, 110, 36, 178, 51, 240, 196, 190, 50, 178, 78, 68, 187, 51, 129, 240, 44, 123, 165, 38, 25, 208, 254, 188},
	},
	{
		testCaseID:   2,
		hashStrategy: sha256.New,
		data: []Data{
			{x: "Hello"}, {x: "Hi"}, {x: "Hey"},
		},
		notInContents: Data{x: "NotInTestTable"},
		expectedHash:  []byte{189, 214, 55, 197, 35, 237, 92, 14, 171, 121, 43, 152, 109, 177, 136, 80, 194, 57, 162, 226, 56, 2, 179, 106, 255, 38, 187, 104, 251, 63, 224, 8},
	},
	{
		testCaseID:   2,
		hashStrategy: sha256.New,
		data: []Data{
			{x: "Hello"}, {x: "Hi"}, {x: "Hey"}, {x: "Greetings"}, {x: "Hola"},
		},
		notInContents: Data{x: "NotInTestTable"},
		expectedHash:  []byte{46, 216, 115, 174, 13, 210, 55, 39, 119, 197, 122, 104, 93, 144, 112, 131, 202, 151, 41, 14, 80, 143, 21, 71, 140, 169, 139, 173, 50, 37, 235, 188},
	},
	{
		testCaseID:   3,
		hashStrategy: sha256.New,
		data: []Data{
			{x: "123"}, {x: "234"}, {x: "345"}, {x: "456"}, {x: "1123"}, {x: "2234"}, {x: "3345"}, {x: "4456"},
		},
		notInContents: Data{x: "NotInTestTable"},
		expectedHash:  []byte{30, 76, 61, 40, 106, 173, 169, 183, 149, 2, 157, 246, 162, 218, 4, 70, 153, 148, 62, 162, 90, 24, 173, 250, 41, 149, 173, 121, 141, 187, 146, 43},
	},
	{
		testCaseID:   4,
		hashStrategy: sha256.New,
		data: []Data{
			{x: "123"}, {x: "234"}, {x: "345"}, {x: "456"}, {x: "1123"}, {x: "2234"}, {x: "3345"}, {x: "4456"}, {x: "5567"},
		},
		notInContents: Data{x: "NotInTestTable"},
		expectedHash:  []byte{143, 37, 161, 192, 69, 241, 248, 56, 169, 87, 79, 145, 37, 155, 51, 159, 209, 129, 164, 140, 130, 167, 16, 182, 133, 205, 126, 55, 237, 188, 89, 236},
	},
}
