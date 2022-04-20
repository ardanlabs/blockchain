// Copyright 2017 Cameron Bergoon
// https://github.com/cbergoon/merkletree
// Licensed under the MIT License, see LICENCE file for details.
// This code has been cleaned up, refactored, and turned into generics.

// Package merkle provides an implementation of a merkel tree for validation
// support for the blockchain.
package merkle

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
)

// Hashable represents the behavior concrete data must exhibit to be used in
// the merkle tree.
type Hashable[T any] interface {
	Hash() ([]byte, error)
	Equals(other T) bool
}

// =============================================================================

// Tree represents a merkle tree that uses data of some type T that exhibits the
// behavior defined by the Hashable constraint.
type Tree[T Hashable[T]] struct {
	Root         *Node[T]
	MerkleRoot   []byte
	Leafs        []*Node[T]
	hashStrategy func() hash.Hash
}

// WithHashStrategy is used to change the default hash strategy of using sha256
// when constructing a new tree.
func WithHashStrategy[T Hashable[T]](hashStrategy func() hash.Hash) func(t *Tree[T]) {
	return func(t *Tree[T]) {
		t.hashStrategy = hashStrategy
	}
}

// NewTree constructs a new merkle tree that uses data of some type T that
// exhibits the behavior defined by the Hashable interface.
func NewTree[T Hashable[T]](values []T, options ...func(t *Tree[T])) (*Tree[T], error) {
	var defaultHashStrategy = sha256.New

	t := Tree[T]{
		hashStrategy: defaultHashStrategy,
	}

	for _, option := range options {
		option(&t)
	}

	if err := t.GenerateTree(values); err != nil {
		return nil, err
	}

	return &t, nil
}

// GenerateTree constructs the leafs and nodes of the tree from the specified
// data. If the tree has been generated previously, the tree is re-generated
// from scratch.
func (t *Tree[T]) GenerateTree(values []T) error {
	if len(values) == 0 {
		return errors.New("cannot construct tree with no content")
	}

	var leafs []*Node[T]
	for _, value := range values {
		hash, err := value.Hash()
		if err != nil {
			return err
		}

		leafs = append(leafs, &Node[T]{
			Hash:  hash,
			Value: value,
			leaf:  true,
			Tree:  t,
		})
	}

	if len(leafs)%2 == 1 {
		duplicate := &Node[T]{
			Hash:  leafs[len(leafs)-1].Hash,
			Value: leafs[len(leafs)-1].Value,
			leaf:  true,
			dup:   true,
			Tree:  t,
		}
		leafs = append(leafs, duplicate)
	}

	root, err := buildIntermediate(leafs, t)
	if err != nil {
		return err
	}

	t.Root = root
	t.Leafs = leafs
	t.MerkleRoot = root.Hash

	return nil
}

// RebuildTree is a helper function that will rebuild the tree reusing only the
// data that it currently holds in the leaves.
func (t *Tree[T]) RebuildTree() error {
	var data []T
	for _, node := range t.Leafs {
		data = append(data, node.Value)
	}

	if err := t.GenerateTree(data); err != nil {
		return err
	}

	return nil
}

// MerkleProof returns the set of hashes and the order of joining those hashes
// for proving a transaction is in the tree.
func (t *Tree[T]) MerkleProof(data T) ([][]byte, []int64, error) {
	for _, node := range t.Leafs {
		if !node.Value.Equals(data) {
			continue
		}

		var merkleProof [][]byte
		var order []int64
		nodeParent := node.Parent

		for nodeParent != nil {
			if bytes.Equal(nodeParent.Left.Hash, node.Hash) {
				merkleProof = append(merkleProof, nodeParent.Right.Hash)
				order = append(order, 1) // right leaf, concat second.
			} else {
				merkleProof = append(merkleProof, nodeParent.Left.Hash)
				order = append(order, 0) // left leaf, concat first.
			}
			node = nodeParent
			nodeParent = nodeParent.Parent
		}

		return merkleProof, order, nil
	}

	return nil, nil, errors.New("unable to find data in tree")
}

// VerifyTree validates the hashes at each level of the tree and returns true
// if the resulting hash at the root of the tree matches the resulting root hash.
func (t *Tree[T]) VerifyTree() error {
	calculatedMerkleRoot, err := t.Root.verifyNode()
	if err != nil {
		return err
	}

	if !bytes.Equal(t.MerkleRoot, calculatedMerkleRoot) {
		return errors.New("root hashe invalid")

	}

	return nil
}

// VerifyData indicates whether a given piece of data is in the tree and if the
// hashes are valid for that data. Returns true if the expected merkle root is
// equivalent to the merkle root calculated on the critical path for a given
// piece of data.
func (t *Tree[T]) VerifyData(data T) error {
	for _, node := range t.Leafs {
		if !node.Value.Equals(data) {
			continue
		}

		currentParent := node.Parent
		for currentParent != nil {
			rightBytes, err := currentParent.Right.CalculateNodeHash()
			if err != nil {
				return err
			}

			leftBytes, err := currentParent.Left.CalculateNodeHash()
			if err != nil {
				return err
			}

			h := t.hashStrategy()
			if _, err := h.Write(append(leftBytes, rightBytes...)); err != nil {
				return err
			}

			if !bytes.Equal(h.Sum(nil), currentParent.Hash) {
				return errors.New("merkle root is not equivalent to the merkle root calculated on the critical path")
			}

			currentParent = currentParent.Parent
		}

		return nil
	}

	return errors.New("merkle root is not equivalent to the merkle root calculated on the critical path")
}

// Values returns a slice of unique values stores in the tree. The last
// value in the tree might be a duplicate when there are an odd number of
// actual values.
func (t *Tree[T]) Values() []T {
	var values []T
	for _, tx := range t.Leafs {
		values = append(values, tx.Value)
	}

	l := len(t.Leafs)
	if bytes.Equal(t.Leafs[l-1].Hash, t.Leafs[l-2].Hash) {
		return values[:l-1]
	}

	return values
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
// immediate relationships, a hash, the data if it is a leaf, and other metadata.
type Node[T Hashable[T]] struct {
	Tree   *Tree[T]
	Parent *Node[T]
	Left   *Node[T]
	Right  *Node[T]
	Hash   []byte
	Value  T
	leaf   bool
	dup    bool
}

// verifyNode walks down the tree until hitting a leaf, calculating the hash at
// each level and returning the resulting hash of the node.
func (n *Node[T]) verifyNode() ([]byte, error) {
	if n.leaf {
		return n.Value.Hash()
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
		return n.Value.Hash()
	}

	h := n.Tree.hashStrategy()
	if _, err := h.Write(append(n.Left.Hash, n.Right.Hash...)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// String returns a string representation of the node.
func (n *Node[T]) String() string {
	return fmt.Sprintf("%t %t %v %v", n.leaf, n.dup, n.Hash, n.Value)
}

// =============================================================================

// ToHex converts a hash in bytes to a hex encoded string.
func ToHex(hash []byte) string {
	return fmt.Sprintf("0x%s", hex.EncodeToString(hash))
}

// buildIntermediate is a helper function that for a given list of leaf nodes,
// constructs the intermediate and root levels of the tree. Returns the resulting
// root node of the tree.
func buildIntermediate[T Hashable[T]](nl []*Node[T], t *Tree[T]) (*Node[T], error) {
	var nodes []*Node[T]

	for i := 0; i < len(nl); i += 2 {
		left, right := i, i+1
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
