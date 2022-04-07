// Copyright 2017 Cameron Bergoon
// https://github.com/cbergoon/merkletree
// Licensed under the MIT License, see LICENCE file for details.

package merkle_test

import (
	"bytes"
	"crypto/md5"
	"hash"
	"testing"

	"github.com/ardanlabs/blockchain/foundation/blockchain/merkle"
)

// MD5 uses the md5 hashing algorithm for the merkle tree.
type MD5 struct {
	x string
}

// Hash hashes the values using md5.
func (t MD5) Hash() ([]byte, error) {
	h := md5.New()
	if _, err := h.Write([]byte(t.x)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// Equals tests for equality of two piece of data.
func (t MD5) Equals(other MD5) (bool, error) {
	return t.x == other.x, nil
}

// =============================================================================

func TestMD5_NewTreeWithHashingStrategy(t *testing.T) {
	for i := 0; i < len(tableMD5); i++ {
		tree, err := merkle.NewTree(tableMD5[i].data, merkle.WithHashStrategy[MD5](tableMD5[i].hashStrategy))
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableMD5[i].testCaseId, err)
		}
		if !bytes.Equal(tree.MerkleRoot, tableMD5[i].expectedHash) {
			t.Errorf("[case:%d] error: expected hash equal to %v got %v", tableMD5[i].testCaseId, tableMD5[i].expectedHash, tree.MerkleRoot)
		}
	}
}

func TestMD5_MerkleRoot(t *testing.T) {
	for i := 0; i < len(tableMD5); i++ {
		tree, err := merkle.NewTree(tableMD5[i].data, merkle.WithHashStrategy[MD5](tableMD5[i].hashStrategy))
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableMD5[i].testCaseId, err)
		}
		if !bytes.Equal(tree.MerkleRoot, tableMD5[i].expectedHash) {
			t.Errorf("[case:%d] error: expected hash equal to %v got %v", tableMD5[i].testCaseId, tableMD5[i].expectedHash, tree.MerkleRoot)
		}
	}
}

func TestMD5_RebuildTree(t *testing.T) {
	for i := 0; i < len(tableMD5); i++ {
		tree, err := merkle.NewTree(tableMD5[i].data, merkle.WithHashStrategy[MD5](tableMD5[i].hashStrategy))
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableMD5[i].testCaseId, err)
		}
		err = tree.RebuildTree()
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error:  %v", tableMD5[i].testCaseId, err)
		}
		if !bytes.Equal(tree.MerkleRoot, tableMD5[i].expectedHash) {
			t.Errorf("[case:%d] error: expected hash equal to %v got %v", tableMD5[i].testCaseId, tableMD5[i].expectedHash, tree.MerkleRoot)
		}
	}
}

func TestMD5_RebuildTreeWith(t *testing.T) {
	for i := 0; i < len(tableMD5)-1; i++ {
		tree, err := merkle.NewTree(tableMD5[i].data, merkle.WithHashStrategy[MD5](tableMD5[i].hashStrategy))
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableMD5[i].testCaseId, err)
		}
		err = tree.RebuildTreeWith(tableMD5[i+1].data)
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableMD5[i].testCaseId, err)
		}
		if !bytes.Equal(tree.MerkleRoot, tableMD5[i+1].expectedHash) {
			t.Errorf("[case:%d] error: expected hash equal to %v got %v", tableMD5[i].testCaseId, tableMD5[i+1].expectedHash, tree.MerkleRoot)
		}
	}
}

func TestMD5_VerifyTree(t *testing.T) {
	for i := 0; i < len(tableMD5); i++ {
		tree, err := merkle.NewTree(tableMD5[i].data, merkle.WithHashStrategy[MD5](tableMD5[i].hashStrategy))
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableMD5[i].testCaseId, err)
		}
		v1, err := tree.VerifyTree()
		if err != nil {
			t.Fatal(err)
		}
		if v1 != true {
			t.Errorf("[case:%d] error: expected tree to be valid", tableMD5[i].testCaseId)
		}
		tree.Root.Hash = []byte{1}
		tree.MerkleRoot = []byte{1}
		v2, err := tree.VerifyTree()
		if err != nil {
			t.Fatal(err)
		}
		if v2 != false {
			t.Errorf("[case:%d] error: expected tree to be invalid", tableMD5[i].testCaseId)
		}
	}
}

func TestMD5_VerifyData(t *testing.T) {
	for i := 0; i < len(tableMD5); i++ {
		tree, err := merkle.NewTree(tableMD5[i].data, merkle.WithHashStrategy[MD5](tableMD5[i].hashStrategy))
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableMD5[i].testCaseId, err)
		}
		if len(tableMD5[i].data) > 0 {
			v, err := tree.VerifyData(tableMD5[i].data[0])
			if err != nil {
				t.Fatal(err)
			}
			if !v {
				t.Errorf("[case:%d] error: expected valid data", tableMD5[i].testCaseId)
			}
		}
		if len(tableMD5[i].data) > 1 {
			v, err := tree.VerifyData(tableMD5[i].data[1])
			if err != nil {
				t.Fatal(err)
			}
			if !v {
				t.Errorf("[case:%d] error: expected valid content", tableMD5[i].testCaseId)
			}
		}
		if len(tableMD5[i].data) > 2 {
			v, err := tree.VerifyData(tableMD5[i].data[2])
			if err != nil {
				t.Fatal(err)
			}
			if !v {
				t.Errorf("[case:%d] error: expected valid content", tableMD5[i].testCaseId)
			}
		}
		if len(tableMD5[i].data) > 0 {
			tree.Root.Hash = []byte{1}
			tree.MerkleRoot = []byte{1}
			v, err := tree.VerifyData(tableMD5[i].data[0])
			if err != nil {
				t.Fatal(err)
			}
			if v {
				t.Errorf("[case:%d] error: expected invalid content", tableMD5[i].testCaseId)
			}
			if err := tree.RebuildTree(); err != nil {
				t.Fatal(err)
			}
		}
		v, err := tree.VerifyData(tableMD5[i].notInData)
		if err != nil {
			t.Fatal(err)
		}
		if v {
			t.Errorf("[case:%d] error: expected invalid content", tableMD5[i].testCaseId)
		}
	}
}

func TestMD5_String(t *testing.T) {
	for i := 0; i < len(tableMD5); i++ {
		tree, err := merkle.NewTree(tableMD5[i].data, merkle.WithHashStrategy[MD5](tableMD5[i].hashStrategy))
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableMD5[i].testCaseId, err)
		}
		if tree.String() == "" {
			t.Errorf("[case:%d] error: expected not empty string", tableMD5[i].testCaseId)
		}
	}
}

func TestMD5_MerklePath(t *testing.T) {
	for i := 0; i < len(tableMD5); i++ {
		tree, err := merkle.NewTree(tableMD5[i].data, merkle.WithHashStrategy[MD5](tableMD5[i].hashStrategy))
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableMD5[i].testCaseId, err)
		}
		for j := 0; j < len(tableMD5[i].data); j++ {
			merklePath, index, _ := tree.MerklePath(tableMD5[i].data[j])

			hash, err := tree.Leafs[j].CalculateNodeHash()
			if err != nil {
				t.Errorf("[case:%d] error: calculateNodeHash error: %v", tableMD5[i].testCaseId, err)
			}
			h := md5.New()
			for k := 0; k < len(merklePath); k++ {
				if index[k] == 1 {
					hash = append(hash, merklePath[k]...)
				} else {
					hash = append(merklePath[k], hash...)
				}
				if _, err := h.Write(hash); err != nil {
					t.Errorf("[case:%d] error: Write error: %v", tableMD5[i].testCaseId, err)
				}
				hash, err = calHash(hash, tableMD5[i].hashStrategy)
				if err != nil {
					t.Errorf("[case:%d] error: calHash error: %v", tableMD5[i].testCaseId, err)
				}
			}
			if !bytes.Equal(tree.MerkleRoot, hash) {
				t.Errorf("[case:%d] error: expected hash equal to %v got %v", tableMD5[i].testCaseId, hash, tree.MerkleRoot)
			}
		}
	}
}

// =============================================================================

var tableMD5 = []struct {
	testCaseId   int
	hashStrategy func() hash.Hash
	data         []MD5
	expectedHash []byte
	notInData    MD5
}{
	{
		testCaseId:   1,
		hashStrategy: md5.New,
		data: []MD5{
			{x: "Hello"}, {x: "Hi"}, {x: "Hey"}, {x: "Hola"},
		},
		notInData:    MD5{x: "NotInTestTable"},
		expectedHash: []byte{217, 158, 206, 52, 191, 78, 253, 233, 25, 55, 69, 142, 254, 45, 127, 144},
	},
	{
		testCaseId:   2,
		hashStrategy: md5.New,
		data: []MD5{
			{x: "Hello"}, {x: "Hi"}, {x: "Hey"},
		},
		notInData:    MD5{x: "NotInTestTable"},
		expectedHash: []byte{145, 228, 171, 107, 94, 219, 221, 171, 7, 195, 206, 128, 148, 98, 59, 76},
	},
	{
		testCaseId:   3,
		hashStrategy: md5.New,
		data: []MD5{
			{x: "Hello"}, {x: "Hi"}, {x: "Hey"}, {x: "Greetings"}, {x: "Hola"},
		},
		notInData:    MD5{x: "NotInTestTable"},
		expectedHash: []byte{167, 200, 229, 62, 194, 247, 117, 12, 206, 194, 90, 235, 70, 14, 100, 100},
	},
	{
		testCaseId:   4,
		hashStrategy: md5.New,
		data: []MD5{
			{x: "123"}, {x: "234"}, {x: "345"}, {x: "456"}, {x: "1123"}, {x: "2234"}, {x: "3345"}, {x: "4456"},
		},
		notInData:    MD5{x: "NotInTestTable"},
		expectedHash: []byte{8, 36, 33, 50, 204, 197, 82, 81, 207, 74, 6, 60, 162, 209, 168, 21},
	},
	{
		testCaseId:   5,
		hashStrategy: md5.New,
		data: []MD5{
			{x: "123"}, {x: "234"}, {x: "345"}, {x: "456"}, {x: "1123"}, {x: "2234"}, {x: "3345"}, {x: "4456"}, {x: "5567"},
		},
		notInData:    MD5{x: "NotInTestTable"},
		expectedHash: []byte{158, 85, 181, 191, 25, 250, 251, 71, 215, 22, 68, 68, 11, 198, 244, 148},
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
