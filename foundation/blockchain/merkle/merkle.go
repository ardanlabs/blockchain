// Copyright 2017 Cameron Bergoon
// https://github.com/cbergoon/merkletree
// Licensed under the MIT License, see LICENCE file for details.

// Package merkle provides an implementation of a merkel tree for validation
// support for the blockchain.
package merkle

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
)

// Hashable represents the data that is stored and verified by the tree. A type
// that implements this interface can be used as an item in the tree.
type Hashable[T any] interface {
	Hash() ([]byte, error)
	Equals(other T) (bool, error)
}

// =============================================================================

// Tree is the container for the tree. It holds a pointer to the root of the tree,
// a list of pointers to the leaf nodes, and the merkle root.
type Tree[T Hashable[T]] struct {
	Root         *Node[T]
	MerkleRoot   []byte
	Leafs        []*Node[T]
	hashStrategy func() hash.Hash
}

// WithHashStrategy allows the configuration of a different hash strategy than
// using the default strategy.
func WithHashStrategy[T Hashable[T]](hashStrategy func() hash.Hash) func(t *Tree[T]) {
	return func(t *Tree[T]) {
		t.hashStrategy = hashStrategy
	}
}

// NewTree creates a new Merkle Tree using the content.
func NewTree[T Hashable[T]](data []T, options ...func(t *Tree[T])) (*Tree[T], error) {
	var defaultHashStrategy = sha256.New
	t := Tree[T]{
		hashStrategy: defaultHashStrategy,
	}

	for _, option := range options {
		option(&t)
	}

	root, leafs, err := buildWithContent(data, &t)
	if err != nil {
		return nil, err
	}

	t.Root = root
	t.Leafs = leafs
	t.MerkleRoot = root.Hash

	return &t, nil
}

// GetMerklePath gets the Merkle path and indexes (left leaf or right leaf).
func (t *Tree[T]) GetMerklePath(data T) ([][]byte, []int64, error) {
	for _, node := range t.Leafs {
		ok, err := node.Data.Equals(data)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			continue
		}

		nodeParent := node.Parent
		var merklePath [][]byte
		var index []int64
		for nodeParent != nil {
			if bytes.Equal(nodeParent.Left.Hash, node.Hash) {
				merklePath = append(merklePath, nodeParent.Right.Hash)
				index = append(index, 1) // right leaf
			} else {
				merklePath = append(merklePath, nodeParent.Left.Hash)
				index = append(index, 0) // left leaf
			}
			node = nodeParent
			nodeParent = nodeParent.Parent
		}

		return merklePath, index, nil
	}

	return nil, nil, nil
}

// RebuildTree is a helper function that will rebuild the tree reusing only the
// content thatit holds in the leaves.
func (t *Tree[T]) RebuildTree() error {
	var data []T
	for _, node := range t.Leafs {
		data = append(data, node.Data)
	}

	root, leafs, err := buildWithContent(data, t)
	if err != nil {
		return err
	}

	t.Root = root
	t.Leafs = leafs
	t.MerkleRoot = root.Hash

	return nil
}

// RebuildTreeWith replaces the content of the tree and does a complete rebuild
// while the root of the tree will be replaced the MerkleTree completely survives
// this operation. Returns an error if the list of content cs contains no entries.
func (t *Tree[T]) RebuildTreeWith(data []T) error {
	root, leafs, err := buildWithContent(data, t)
	if err != nil {
		return err
	}

	t.Root = root
	t.Leafs = leafs
	t.MerkleRoot = root.Hash

	return nil
}

// VerifyTree verify tree validates the hashes at each level of the tree and
// returns true if the resulting hash at the root of the tree matches the
// resulting root hash; returns false otherwise.
func (t *Tree[T]) VerifyTree() (bool, error) {
	calculatedMerkleRoot, err := t.Root.verifyNode()
	if err != nil {
		return false, err
	}

	if bytes.Equal(t.MerkleRoot, calculatedMerkleRoot) {
		return true, nil
	}

	return false, nil
}

// VerifyContent indicates whether a given content is in the tree and the hashes
// are valid for that content. Returns true if the expected Merkle Root is
// equivalent to the Merkle root calculated on the critical path for a given
// content. Returns true if valid and false otherwise.
func (t *Tree[T]) VerifyContent(data T) (bool, error) {
	for _, node := range t.Leafs {
		ok, err := node.Data.Equals(data)
		if err != nil {
			return false, err
		}
		if !ok {
			continue
		}

		currentParent := node.Parent
		for currentParent != nil {
			rightBytes, err := currentParent.Right.CalculateNodeHash()
			if err != nil {
				return false, err
			}

			leftBytes, err := currentParent.Left.CalculateNodeHash()
			if err != nil {
				return false, err
			}

			h := t.hashStrategy()
			if _, err := h.Write(append(leftBytes, rightBytes...)); err != nil {
				return false, err
			}

			if !bytes.Equal(h.Sum(nil), currentParent.Hash) {
				return false, nil
			}

			currentParent = currentParent.Parent
		}

		return true, nil
	}

	return false, nil
}

// String returns a string representation of the tree. Only leaf nodes are
// included in the output.
func (t *Tree[T]) String() string {
	s := ""

	for _, l := range t.Leafs {
		s += fmt.Sprint(l)
		s += "\n"
	}

	return s
}

// =============================================================================

// Node represents a node, root, or leaf in the tree. It stores pointers to its
// immediate relationships, a hash, the content stored if it is a leaf, and
// other metadata.
type Node[T Hashable[T]] struct {
	Tree   *Tree[T]
	Parent *Node[T]
	Left   *Node[T]
	Right  *Node[T]
	Hash   []byte
	Data   T
	leaf   bool
	dup    bool
}

// verifyNode walks down the tree until hitting a leaf, calculating the hash at
// each level and returning the resulting hash of Node n.
func (n *Node[T]) verifyNode() ([]byte, error) {
	if n.leaf {
		return n.Data.Hash()
	}

	rightBytes, err := n.Right.verifyNode()
	if err != nil {
		return nil, err
	}

	leftBytes, err := n.Left.verifyNode()
	if err != nil {
		return nil, err
	}

	h := n.Tree.hashStrategy()
	if _, err := h.Write(append(leftBytes, rightBytes...)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// CalculateNodeHash is a helper function that calculates the hash of the node.
func (n *Node[T]) CalculateNodeHash() ([]byte, error) {
	if n.leaf {
		return n.Data.Hash()
	}

	h := n.Tree.hashStrategy()
	if _, err := h.Write(append(n.Left.Hash, n.Right.Hash...)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// String returns a string representation of the node.
func (n *Node[T]) String() string {
	return fmt.Sprintf("%t %t %v %v", n.leaf, n.dup, n.Hash, n.Data)
}

// =============================================================================

// buildWithContent is a helper function that for a given set of Contents,
// generates a corresponding tree and returns the root node, a list of leaf
// nodes, and a possible error. Returns an error if cs contains no Contents.
func buildWithContent[T Hashable[T]](data []T, t *Tree[T]) (*Node[T], []*Node[T], error) {
	if len(data) == 0 {
		return nil, nil, errors.New("cannot construct tree with no content")
	}

	var leafs []*Node[T]
	for _, dt := range data {
		hash, err := dt.Hash()
		if err != nil {
			return nil, nil, err
		}

		leafs = append(leafs, &Node[T]{
			Hash: hash,
			Data: dt,
			leaf: true,
			Tree: t,
		})
	}

	if len(leafs)%2 == 1 {
		duplicate := &Node[T]{
			Hash: leafs[len(leafs)-1].Hash,
			Data: leafs[len(leafs)-1].Data,
			leaf: true,
			dup:  true,
			Tree: t,
		}
		leafs = append(leafs, duplicate)
	}

	root, err := buildIntermediate(leafs, t)
	if err != nil {
		return nil, nil, err
	}

	return root, leafs, nil
}

// buildIntermediate is a helper function that for a given list of leaf nodes,
// constructs the intermediate and root levels of the tree. Returns the resulting
// root node of the tree.
func buildIntermediate[T Hashable[T]](nl []*Node[T], t *Tree[T]) (*Node[T], error) {
	var nodes []*Node[T]

	for i := 0; i < len(nl); i += 2 {
		var left, right int = i, i + 1
		if i+1 == len(nl) {
			right = i
		}

		h := t.hashStrategy()
		chash := append(nl[left].Hash, nl[right].Hash...)
		if _, err := h.Write(chash); err != nil {
			return nil, err
		}

		n := Node[T]{
			Left:  nl[left],
			Right: nl[right],
			Hash:  h.Sum(nil),
			Tree:  t,
		}

		nodes = append(nodes, &n)
		nl[left].Parent = &n
		nl[right].Parent = &n

		if len(nl) == 2 {
			return &n, nil
		}
	}

	return buildIntermediate(nodes, t)
}
