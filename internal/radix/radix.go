// Package radix implements a generic radix tree. This is based primarily on this original implementation
// in this project:
//
// https://github.com/armon/go-radix
//
// The original was not generic, but this one is. I also cut out some extraneous operations I didn't feel I needed.
package radix

import (
	"sort"
	"strings"
)

// WalkFunc is used when walking the tree. Takes a
// key and value, returning if iteration should
// be terminated.
type WalkFunc[T any] func(s string, v T) bool

// leafNode is used to represent a value
type leafNode[T any] struct {
	key string
	val T
}

// edge is used to represent an edge node
type edge[T any] struct {
	label byte
	node  *node[T]
}

type node[T any] struct {
	// leaf is used to store possible leaf
	leaf *leafNode[T]

	// prefix is the common prefix we ignore
	prefix string

	// Edges should be stored in-order for iteration.
	// We avoid a fully materialized slice to save memory,
	// since in most cases we expect to be sparse
	edges edges[T]
}

func (n *node[T]) isLeaf() bool {
	return n.leaf != nil
}

func (n *node[T]) addEdge(e edge[T]) {
	num := len(n.edges)
	idx := sort.Search(num, func(i int) bool {
		return n.edges[i].label >= e.label
	})

	n.edges = append(n.edges, edge[T]{})
	copy(n.edges[idx+1:], n.edges[idx:])
	n.edges[idx] = e
}

func (n *node[T]) updateEdge(label byte, node *node[T]) {
	num := len(n.edges)
	idx := sort.Search(num, func(i int) bool {
		return n.edges[i].label >= label
	})
	if idx < num && n.edges[idx].label == label {
		n.edges[idx].node = node
		return
	}
	panic("replacing missing edge")
}

func (n *node[T]) getEdge(label byte) *node[T] {
	num := len(n.edges)
	idx := sort.Search(num, func(i int) bool {
		return n.edges[i].label >= label
	})
	if idx < num && n.edges[idx].label == label {
		return n.edges[idx].node
	}
	return nil
}

func (n *node[T]) delEdge(label byte) {
	num := len(n.edges)
	idx := sort.Search(num, func(i int) bool {
		return n.edges[i].label >= label
	})
	if idx < num && n.edges[idx].label == label {
		copy(n.edges[idx:], n.edges[idx+1:])
		n.edges[len(n.edges)-1] = edge[T]{}
		n.edges = n.edges[:len(n.edges)-1]
	}
}

type edges[T any] []edge[T]

func (e edges[T]) Len() int {
	return len(e)
}

func (e edges[T]) Less(i, j int) bool {
	return e[i].label < e[j].label
}

func (e edges[T]) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e edges[T]) Sort() {
	sort.Sort(e)
}

// Tree implements a radix tree. It's a map-like structure that lets you efficiently find exact instances of keys
// as well as anything matching prefixes.
type Tree[T any] struct {
	root         *node[T]
	size         int
	defaultValue T
}

// New returns an empty radix Tree.
func New[T any]() Tree[T] {
	return Tree[T]{root: &node[T]{}}
}

// Len is used to return the number of elements in the tree
func (tree *Tree[T]) Len() int {
	return tree.size
}

// longestPrefix finds the length of the shared prefix of two strings.
func longestPrefix(k1, k2 string) int {
	maxLength := len(k1)
	if l := len(k2); l < maxLength {
		maxLength = l
	}
	var i int
	for i = 0; i < maxLength; i++ {
		if k1[i] != k2[i] {
			break
		}
	}
	return i
}

// Insert is used to add a new entry or update an existing entry. Returns true if an existing record is updated.
func (tree *Tree[T]) Insert(s string, v T) (T, bool) {
	var parent *node[T]
	n := tree.root
	search := s
	for {
		// Handle key exhaustion
		if len(search) == 0 {
			if n.isLeaf() {
				old := n.leaf.val
				n.leaf.val = v
				return old, true
			}

			n.leaf = &leafNode[T]{
				key: s,
				val: v,
			}
			tree.size++
			return tree.defaultValue, false
		}

		// Look for the edge
		parent = n
		n = n.getEdge(search[0])

		// No edge, create one
		if n == nil {
			e := edge[T]{
				label: search[0],
				node: &node[T]{
					leaf: &leafNode[T]{
						key: s,
						val: v,
					},
					prefix: search,
				},
			}
			parent.addEdge(e)
			tree.size++
			return tree.defaultValue, false
		}

		// Determine the longest prefix of the search key on match
		commonPrefix := longestPrefix(search, n.prefix)
		if commonPrefix == len(n.prefix) {
			search = search[commonPrefix:]
			continue
		}

		// Split the node
		tree.size++
		child := &node[T]{
			prefix: search[:commonPrefix],
		}
		parent.updateEdge(search[0], child)

		// Restore the existing node
		child.addEdge(edge[T]{
			label: n.prefix[commonPrefix],
			node:  n,
		})
		n.prefix = n.prefix[commonPrefix:]

		// Create a new leaf node
		leaf := &leafNode[T]{
			key: s,
			val: v,
		}

		// If the new key is a subset, add to this node
		search = search[commonPrefix:]
		if len(search) == 0 {
			child.leaf = leaf
			return tree.defaultValue, false
		}

		// Create a new edge for the node
		child.addEdge(edge[T]{
			label: search[0],
			node: &node[T]{
				leaf:   leaf,
				prefix: search,
			},
		})
		return tree.defaultValue, false
	}
}

// Delete is used to remove a node from the tree, returning the previous value and if it was deleted.
func (tree *Tree[T]) Delete(s string) (T, bool) {
	var parent *node[T]
	var label byte
	n := tree.root
	search := s
	for {
		// Check for key exhaustion
		if len(search) == 0 {
			if !n.isLeaf() {
				break
			}
			goto DELETE
		}

		// Look for an edge
		parent = n
		label = search[0]
		n = n.getEdge(label)
		if n == nil {
			break
		}

		// Consume the search prefix
		if strings.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):]
		} else {
			break
		}
	}
	return tree.defaultValue, false

DELETE:
	// Delete the leaf
	leaf := n.leaf
	n.leaf = nil
	tree.size--

	// Check if we should delete this node from the parent
	if parent != nil && len(n.edges) == 0 {
		parent.delEdge(label)
	}

	// Check if we should merge this node
	if n != tree.root && len(n.edges) == 1 {
		n.mergeChild()
	}

	// Check if we should merge the parent's other child
	parent.delEdge(n.prefix[0])
	if parent != nil && parent != tree.root && len(parent.edges) == 1 && !parent.isLeaf() {
		parent.mergeChild()
	}

	return leaf.val, true
}

// DeletePrefix is used to delete the subtree under a prefix
// Returns how many nodes were deleted
// Use this to delete large subtrees efficiently
func (tree *Tree[T]) DeletePrefix(s string) int {
	return tree.deletePrefix(nil, tree.root, s)
}

// delete does a recursive deletion
func (tree *Tree[T]) deletePrefix(parent, n *node[T], prefix string) int {
	// Check for key exhaustion
	if len(prefix) == 0 {
		// Remove the leaf node
		subTreeSize := 0
		// recursively walk from all edges of the node to be deleted
		recursiveWalk(n, func(s string, v T) bool {
			subTreeSize++
			return false
		})
		if n.isLeaf() {
			n.leaf = nil
		}
		n.edges = nil // deletes the entire subtree

		// Check if we should merge the parent's other child
		if parent != nil && parent != tree.root && len(parent.edges) == 1 && !parent.isLeaf() {
			parent.mergeChild()
		}
		tree.size -= subTreeSize
		return subTreeSize
	}

	// Look for an edge
	label := prefix[0]
	child := n.getEdge(label)
	if child == nil || (!strings.HasPrefix(child.prefix, prefix) && !strings.HasPrefix(prefix, child.prefix)) {
		return 0
	}

	// Consume the search prefix
	if len(child.prefix) > len(prefix) {
		prefix = prefix[len(prefix):]
	} else {
		prefix = prefix[len(child.prefix):]
	}
	return tree.deletePrefix(n, child, prefix)
}

func (n *node[T]) mergeChild() {
	e := n.edges[0]
	child := e.node
	n.prefix = n.prefix + child.prefix
	n.leaf = child.leaf
	n.edges = child.edges
}

// Get is used to look up a specific key, returning the value and if it was found.
func (tree *Tree[T]) Get(s string) (T, bool) {
	n := tree.root
	search := s
	for {
		// Check for key exhaustion
		if len(search) == 0 {
			if n.isLeaf() {
				return n.leaf.val, true
			}
			break
		}

		// Look for an edge
		n = n.getEdge(search[0])
		if n == nil {
			break
		}

		// Consume the search prefix
		if strings.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):]
		} else {
			break
		}
	}
	return tree.defaultValue, false
}

// LongestPrefix is like Get, but instead of an exact match, it will return the longest prefix match.
func (tree *Tree[T]) LongestPrefix(s string) (string, T, bool) {
	var last *leafNode[T]
	n := tree.root
	search := s
	for {
		// Look for a leaf node
		if n.isLeaf() {
			last = n.leaf
		}

		// Check for key exhaustion
		if len(search) == 0 {
			break
		}

		// Look for an edge
		n = n.getEdge(search[0])
		if n == nil {
			break
		}

		// Consume the search prefix
		if strings.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):]
		} else {
			break
		}
	}
	if last != nil {
		return last.key, last.val, true
	}
	return "", tree.defaultValue, false
}

// Minimum is used to return the minimum value in the tree.
func (tree *Tree[T]) Minimum() (string, T, bool) {
	n := tree.root
	for {
		if n.isLeaf() {
			return n.leaf.key, n.leaf.val, true
		}
		if len(n.edges) > 0 {
			n = n.edges[0].node
		} else {
			break
		}
	}
	return "", tree.defaultValue, false
}

// Maximum is used to return the maximum value in the tree.
func (tree *Tree[T]) Maximum() (string, T, bool) {
	n := tree.root
	for {
		if num := len(n.edges); num > 0 {
			n = n.edges[num-1].node
			continue
		}
		if n.isLeaf() {
			return n.leaf.key, n.leaf.val, true
		}
		break
	}
	return "", tree.defaultValue, false
}

// Walk visits every not in the tree, invoking your callback function for each one.
func (tree *Tree[T]) Walk(fn WalkFunc[T]) {
	recursiveWalk(tree.root, fn)
}

// WalkPrefix is used to visit just the nodes whose key starts with the given prefix.
func (tree *Tree[T]) WalkPrefix(prefix string, fn WalkFunc[T]) {
	n := tree.root
	search := prefix
	for {
		// Check for key exhaustion
		if len(search) == 0 {
			recursiveWalk(n, fn)
			return
		}

		// Look for an edge
		n = n.getEdge(search[0])
		if n == nil {
			return
		}

		// Consume the search prefix
		if strings.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):]
			continue
		}
		if strings.HasPrefix(n.prefix, search) {
			// Child may be under our search prefix
			recursiveWalk(n, fn)
		}
		return
	}
}

// WalkPath is used to walk the tree, but only visiting nodes
// from the root down to a given leaf. Where WalkPrefix walks
// all the entries *under* the given prefix, this walks the
// entries *above* the given prefix.
func (tree *Tree[T]) WalkPath(path string, fn WalkFunc[T]) {
	n := tree.root
	search := path
	for {
		// Visit the leaf values if any
		if n.leaf != nil && fn(n.leaf.key, n.leaf.val) {
			return
		}

		// Check for key exhaustion
		if len(search) == 0 {
			return
		}

		// Look for an edge
		n = n.getEdge(search[0])
		if n == nil {
			return
		}

		// Consume the search prefix
		if strings.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):]
		} else {
			break
		}
	}
}

// recursiveWalk is used to do a pre-order walk of a node
// recursively. Returns true if the walk should be aborted
func recursiveWalk[T any](n *node[T], fn WalkFunc[T]) bool {
	// Visit the leaf values if any
	if n.leaf != nil && fn(n.leaf.key, n.leaf.val) {
		return true
	}

	// Recurse on the children
	i := 0
	k := len(n.edges) // keeps track of number of edges in previous iteration
	for i < k {
		e := n.edges[i]
		if recursiveWalk(e.node, fn) {
			return true
		}
		// It is a possibility that the WalkFunc modified the node we are
		// iterating on. If there are no more edges, mergeChild happened,
		// so the last edge became the current node n, on which we'll
		// iterate one last time.
		if len(n.edges) == 0 {
			return recursiveWalk(n, fn)
		}
		// If there are now less edges than in the previous iteration,
		// then do not increment the loop index, since the current index
		// points to a new edge. Otherwise, get to the next index.
		if len(n.edges) >= k {
			i++
		}
		k = len(n.edges)
	}
	return false
}
