package merkle_test

import (
	"bytes"
	"crypto/sha256"
	"hash"
	"testing"

	"github.com/ardanlabs/blockchain/foundation/blockchain/merkle"
)

// TestSHA256Content implements the Content interface provided by merkletree and
// represents the content stored in the tree.
type TestSHA256Content struct {
	x string
}

// Hash hashes the values of a TestSHA256Content.
func (t TestSHA256Content) Hash() ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(t.x)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// Equals tests for equality of two Contents.
func (t TestSHA256Content) Equals(other TestSHA256Content) (bool, error) {
	return t.x == other.x, nil
}

// =============================================================================

func TestNewTreeSHA256Content(t *testing.T) {
	for i := 0; i < len(tableSHA256Content); i++ {
		if !tableSHA256Content[i].defaultHashStrategy {
			continue
		}
		tree, err := merkle.NewTree(tableSHA256Content[i].contents)
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableSHA256Content[i].testCaseId, err)
		}
		if !bytes.Equal(tree.MerkleRoot, tableSHA256Content[i].expectedHash) {
			t.Errorf("[case:%d] error: expected hash equal to %v got %v", tableSHA256Content[i].testCaseId, tableSHA256Content[i].expectedHash, tree.MerkleRoot)
		}
	}
}

func TestNewTreeWithHashingStrategySHA256Content(t *testing.T) {
	for i := 0; i < len(tableSHA256Content); i++ {
		tree, err := merkle.NewTree(tableSHA256Content[i].contents, merkle.WithHashStrategy[TestSHA256Content](tableSHA256Content[i].hashStrategy))
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableSHA256Content[i].testCaseId, err)
		}
		if !bytes.Equal(tree.MerkleRoot, tableSHA256Content[i].expectedHash) {
			t.Errorf("[case:%d] error: expected hash equal to %v got %v", tableSHA256Content[i].testCaseId, tableSHA256Content[i].expectedHash, tree.MerkleRoot)
		}
	}
}

func TestMerkleTreeSHA256Content_MerkleRoot(t *testing.T) {
	for i := 0; i < len(tableSHA256Content); i++ {
		var tree *merkle.Tree[TestSHA256Content]
		var err error
		if tableSHA256Content[i].defaultHashStrategy {
			tree, err = merkle.NewTree(tableSHA256Content[i].contents)
		} else {
			tree, err = merkle.NewTree(tableSHA256Content[i].contents, merkle.WithHashStrategy[TestSHA256Content](tableSHA256Content[i].hashStrategy))
		}
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableSHA256Content[i].testCaseId, err)
		}
		if !bytes.Equal(tree.MerkleRoot, tableSHA256Content[i].expectedHash) {
			t.Errorf("[case:%d] error: expected hash equal to %v got %v", tableSHA256Content[i].testCaseId, tableSHA256Content[i].expectedHash, tree.MerkleRoot)
		}
	}
}

func TestMerkleTreeSHA256Content_RebuildTree(t *testing.T) {
	for i := 0; i < len(tableSHA256Content); i++ {
		var tree *merkle.Tree[TestSHA256Content]
		var err error
		if tableSHA256Content[i].defaultHashStrategy {
			tree, err = merkle.NewTree(tableSHA256Content[i].contents)
		} else {
			tree, err = merkle.NewTree(tableSHA256Content[i].contents, merkle.WithHashStrategy[TestSHA256Content](tableSHA256Content[i].hashStrategy))
		}
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableSHA256Content[i].testCaseId, err)
		}
		err = tree.RebuildTree()
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error:  %v", tableSHA256Content[i].testCaseId, err)
		}
		if !bytes.Equal(tree.MerkleRoot, tableSHA256Content[i].expectedHash) {
			t.Errorf("[case:%d] error: expected hash equal to %v got %v", tableSHA256Content[i].testCaseId, tableSHA256Content[i].expectedHash, tree.MerkleRoot)
		}
	}
}

func TestMerkleTreeSHA256Content_RebuildTreeWith(t *testing.T) {
	for i := 0; i < len(tableSHA256Content)-1; i++ {
		if tableSHA256Content[i].hashStrategyName != tableSHA256Content[i+1].hashStrategyName {
			continue
		}
		var tree *merkle.Tree[TestSHA256Content]
		var err error
		if tableSHA256Content[i].defaultHashStrategy {
			tree, err = merkle.NewTree(tableSHA256Content[i].contents)
		} else {
			tree, err = merkle.NewTree(tableSHA256Content[i].contents, merkle.WithHashStrategy[TestSHA256Content](tableSHA256Content[i].hashStrategy))
		}
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableSHA256Content[i].testCaseId, err)
		}
		err = tree.RebuildTreeWith(tableSHA256Content[i+1].contents)
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableSHA256Content[i].testCaseId, err)
		}
		if !bytes.Equal(tree.MerkleRoot, tableSHA256Content[i+1].expectedHash) {
			t.Errorf("[case:%d] error: expected hash equal to %v got %v", tableSHA256Content[i].testCaseId, tableSHA256Content[i+1].expectedHash, tree.MerkleRoot)
		}
	}
}

func TestMerkleTreeSHA256Content_VerifyTree(t *testing.T) {
	for i := 0; i < len(tableSHA256Content); i++ {
		var tree *merkle.Tree[TestSHA256Content]
		var err error
		if tableSHA256Content[i].defaultHashStrategy {
			tree, err = merkle.NewTree(tableSHA256Content[i].contents)
		} else {
			tree, err = merkle.NewTree(tableSHA256Content[i].contents, merkle.WithHashStrategy[TestSHA256Content](tableSHA256Content[i].hashStrategy))
		}
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableSHA256Content[i].testCaseId, err)
		}
		v1, err := tree.VerifyTree()
		if err != nil {
			t.Fatal(err)
		}
		if v1 != true {
			t.Errorf("[case:%d] error: expected tree to be valid", tableSHA256Content[i].testCaseId)
		}
		tree.Root.Hash = []byte{1}
		tree.MerkleRoot = []byte{1}
		v2, err := tree.VerifyTree()
		if err != nil {
			t.Fatal(err)
		}
		if v2 != false {
			t.Errorf("[case:%d] error: expected tree to be invalid", tableSHA256Content[i].testCaseId)
		}
	}
}

func TestMerkleTreeSHA256Content_VerifyContent(t *testing.T) {
	for i := 0; i < len(tableSHA256Content); i++ {
		var tree *merkle.Tree[TestSHA256Content]
		var err error
		if tableSHA256Content[i].defaultHashStrategy {
			tree, err = merkle.NewTree(tableSHA256Content[i].contents)
		} else {
			tree, err = merkle.NewTree(tableSHA256Content[i].contents, merkle.WithHashStrategy[TestSHA256Content](tableSHA256Content[i].hashStrategy))
		}
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableSHA256Content[i].testCaseId, err)
		}
		if len(tableSHA256Content[i].contents) > 0 {
			v, err := tree.VerifyContent(tableSHA256Content[i].contents[0])
			if err != nil {
				t.Fatal(err)
			}
			if !v {
				t.Errorf("[case:%d] error: expected valid content", tableSHA256Content[i].testCaseId)
			}
		}
		if len(tableSHA256Content[i].contents) > 1 {
			v, err := tree.VerifyContent(tableSHA256Content[i].contents[1])
			if err != nil {
				t.Fatal(err)
			}
			if !v {
				t.Errorf("[case:%d] error: expected valid content", tableSHA256Content[i].testCaseId)
			}
		}
		if len(tableSHA256Content[i].contents) > 2 {
			v, err := tree.VerifyContent(tableSHA256Content[i].contents[2])
			if err != nil {
				t.Fatal(err)
			}
			if !v {
				t.Errorf("[case:%d] error: expected valid content", tableSHA256Content[i].testCaseId)
			}
		}
		if len(tableSHA256Content[i].contents) > 0 {
			tree.Root.Hash = []byte{1}
			tree.MerkleRoot = []byte{1}
			v, err := tree.VerifyContent(tableSHA256Content[i].contents[0])
			if err != nil {
				t.Fatal(err)
			}
			if v {
				t.Errorf("[case:%d] error: expected invalid content", tableSHA256Content[i].testCaseId)
			}
			if err := tree.RebuildTree(); err != nil {
				t.Fatal(err)
			}
		}
		v, err := tree.VerifyContent(tableSHA256Content[i].notInContents)
		if err != nil {
			t.Fatal(err)
		}
		if v {
			t.Errorf("[case:%d] error: expected invalid content", tableSHA256Content[i].testCaseId)
		}
	}
}

func TestMerkleTreeSHA256Content_String(t *testing.T) {
	for i := 0; i < len(tableSHA256Content); i++ {
		var tree *merkle.Tree[TestSHA256Content]
		var err error
		if tableSHA256Content[i].defaultHashStrategy {
			tree, err = merkle.NewTree(tableSHA256Content[i].contents)
		} else {
			tree, err = merkle.NewTree(tableSHA256Content[i].contents, merkle.WithHashStrategy[TestSHA256Content](tableSHA256Content[i].hashStrategy))
		}
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableSHA256Content[i].testCaseId, err)
		}
		if tree.String() == "" {
			t.Errorf("[case:%d] error: expected not empty string", tableSHA256Content[i].testCaseId)
		}
	}
}

func TestMerkleTreeSHA256Content_MerklePath(t *testing.T) {
	for i := 0; i < len(tableSHA256Content); i++ {
		var tree *merkle.Tree[TestSHA256Content]
		var err error
		if tableSHA256Content[i].defaultHashStrategy {
			tree, err = merkle.NewTree(tableSHA256Content[i].contents)
		} else {
			tree, err = merkle.NewTree(tableSHA256Content[i].contents, merkle.WithHashStrategy[TestSHA256Content](tableSHA256Content[i].hashStrategy))
		}
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", tableSHA256Content[i].testCaseId, err)
		}
		for j := 0; j < len(tableSHA256Content[i].contents); j++ {
			merklePath, index, _ := tree.GetMerklePath(tableSHA256Content[i].contents[j])

			hash, err := tree.Leafs[j].CalculateNodeHash()
			if err != nil {
				t.Errorf("[case:%d] error: calculateNodeHash error: %v", tableSHA256Content[i].testCaseId, err)
			}
			h := sha256.New()
			for k := 0; k < len(merklePath); k++ {
				if index[k] == 1 {
					hash = append(hash, merklePath[k]...)
				} else {
					hash = append(merklePath[k], hash...)
				}
				if _, err := h.Write(hash); err != nil {
					t.Errorf("[case:%d] error: Write error: %v", tableSHA256Content[i].testCaseId, err)
				}
				hash, err = calHash(hash, tableSHA256Content[i].hashStrategy)
				if err != nil {
					t.Errorf("[case:%d] error: calHash error: %v", tableSHA256Content[i].testCaseId, err)
				}
			}
			if !bytes.Equal(tree.MerkleRoot, hash) {
				t.Errorf("[case:%d] error: expected hash equal to %v got %v", tableSHA256Content[i].testCaseId, hash, tree.MerkleRoot)
			}
		}
	}
}

// =============================================================================

var tableSHA256Content = []struct {
	testCaseId          int
	hashStrategy        func() hash.Hash
	hashStrategyName    string
	defaultHashStrategy bool
	contents            []TestSHA256Content
	expectedHash        []byte
	notInContents       TestSHA256Content
}{
	{
		testCaseId:          0,
		hashStrategy:        sha256.New,
		hashStrategyName:    "sha256",
		defaultHashStrategy: true,
		contents: []TestSHA256Content{
			{
				x: "Hello",
			},
			{
				x: "Hi",
			},
			{
				x: "Hey",
			},
			{
				x: "Hola",
			},
		},
		notInContents: TestSHA256Content{x: "NotInTestTable"},
		expectedHash:  []byte{95, 48, 204, 128, 19, 59, 147, 148, 21, 110, 36, 178, 51, 240, 196, 190, 50, 178, 78, 68, 187, 51, 129, 240, 44, 123, 165, 38, 25, 208, 254, 188},
	},
	{
		testCaseId:          1,
		hashStrategy:        sha256.New,
		hashStrategyName:    "sha256",
		defaultHashStrategy: true,
		contents: []TestSHA256Content{
			{
				x: "Hello",
			},
			{
				x: "Hi",
			},
			{
				x: "Hey",
			},
		},
		notInContents: TestSHA256Content{x: "NotInTestTable"},
		expectedHash:  []byte{189, 214, 55, 197, 35, 237, 92, 14, 171, 121, 43, 152, 109, 177, 136, 80, 194, 57, 162, 226, 56, 2, 179, 106, 255, 38, 187, 104, 251, 63, 224, 8},
	},
	{
		testCaseId:          2,
		hashStrategy:        sha256.New,
		hashStrategyName:    "sha256",
		defaultHashStrategy: true,
		contents: []TestSHA256Content{
			{
				x: "Hello",
			},
			{
				x: "Hi",
			},
			{
				x: "Hey",
			},
			{
				x: "Greetings",
			},
			{
				x: "Hola",
			},
		},
		notInContents: TestSHA256Content{x: "NotInTestTable"},
		expectedHash:  []byte{46, 216, 115, 174, 13, 210, 55, 39, 119, 197, 122, 104, 93, 144, 112, 131, 202, 151, 41, 14, 80, 143, 21, 71, 140, 169, 139, 173, 50, 37, 235, 188},
	},
	{
		testCaseId:          3,
		hashStrategy:        sha256.New,
		hashStrategyName:    "sha256",
		defaultHashStrategy: true,
		contents: []TestSHA256Content{
			{
				x: "123",
			},
			{
				x: "234",
			},
			{
				x: "345",
			},
			{
				x: "456",
			},
			{
				x: "1123",
			},
			{
				x: "2234",
			},
			{
				x: "3345",
			},
			{
				x: "4456",
			},
		},
		notInContents: TestSHA256Content{x: "NotInTestTable"},
		expectedHash:  []byte{30, 76, 61, 40, 106, 173, 169, 183, 149, 2, 157, 246, 162, 218, 4, 70, 153, 148, 62, 162, 90, 24, 173, 250, 41, 149, 173, 121, 141, 187, 146, 43},
	},
	{
		testCaseId:          4,
		hashStrategy:        sha256.New,
		hashStrategyName:    "sha256",
		defaultHashStrategy: true,
		contents: []TestSHA256Content{
			{
				x: "123",
			},
			{
				x: "234",
			},
			{
				x: "345",
			},
			{
				x: "456",
			},
			{
				x: "1123",
			},
			{
				x: "2234",
			},
			{
				x: "3345",
			},
			{
				x: "4456",
			},
			{
				x: "5567",
			},
		},
		notInContents: TestSHA256Content{x: "NotInTestTable"},
		expectedHash:  []byte{143, 37, 161, 192, 69, 241, 248, 56, 169, 87, 79, 145, 37, 155, 51, 159, 209, 129, 164, 140, 130, 167, 16, 182, 133, 205, 126, 55, 237, 188, 89, 236},
	},
}
