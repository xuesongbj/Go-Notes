package binarysearchtree

import "sync"

// Node a single node that composes the tree
type Node struct {
	Key   int         `json:"key"`
	Value interface{} `json:"value"`
	Left  *Node       `json:"left"`
	Right *Node       `json:"right"`
}

// Tree the binary search tree(Thread safe)
type Tree struct {
	Root *Node

	Lock sync.RWMutex
}

// Insert inserts the value t in the tree
func (t *Tree) Insert(key int, value interface{}) {
	t.Lock.Lock()
	defer t.Lock.Unlock()

	n := &Node{key, value, nil, nil}

	if t.Root == nil {
		t.Root = n
	} else {
		insertNode(t.Root, n)
	}
}

// internal function to find the correct place for a node in a tree
func insertNode(node, newNode *Node) {
	if newNode.Key < node.Key {
		if node.Left == nil {
			node.Left = newNode
		} else {
			insertNode(node.Left, newNode)
		}
	} else {
		if node.Right == nil {
			node.Right = newNode
		} else {
			insertNode(node.Right, newNode)
		}
	}
}

// TraverseAllNodes visits all nodes with in-order traversing
func (t *Tree) TraverseAllNodes(f func(interface{})) {
	t.Lock.Lock()
	defer t.Lock.Unlock()

	traverseAllNodes(t.Root, f)
}

// internal recursive function to traverse in order
func traverseAllNodes(n *Node, f func(interface{})) {
	if n != nil {
		traverseAllNodes(n.Left, f)
		f(n.Value)
		traverseAllNodes(n.Right, f)
	}
}

// PreOrderTraverse visits all nodes with pre-order traversin
func (t *Tree) PreOrderTraverse(f func(interface{})) {
	t.Lock.Lock()
	defer t.Lock.Unlock()

	preOrderTraverse(t.Root, f)
}

// internal recursive function to traverse pre order
func preOrderTraverse(n *Node, f func(interface{})) {
	if n != nil {
		f(n.Value)
		preOrderTraverse(n.Left, f)
		preOrderTraverse(n.Right, f)
	}
}

// PostOrderTraverse  visits all nodes with post-order traversing
func (t *Tree) PostOrderTraverse(f func(interface{})) {
	t.Lock.Lock()
	defer t.Lock.Unlock()

	postOrderTraverse(t.Root, f)
}

// internal recursive function to traverse post order
func postOrderTraverse(n *Node, f func(interface{})) {
	if n != nil {
		postOrderTraverse(n.Left, f)
		postOrderTraverse(n.Right, f)
		f(n.Value)
	}
}

// Min returns the value with min value stored in the tree
func (t *Tree) Min() interface{} {
	t.Lock.Lock()
	defer t.Lock.Unlock()

	n := t.Root
	if n == nil {
		return nil
	}

	for {
		if n.Left == nil {
			return n.Value
		}
		n = n.Left
	}
}

// Max returns the value with max value stored in the tree
func (t *Tree) Max() interface{} {
	t.Lock.Lock()
	defer t.Lock.Unlock()

	n := t.Root
	if n == nil {
		return nil
	}

	for {
		if n.Right == nil {
			return n.Value
		}
		n = n.Right
	}
}

// Search returns true if the value t exists in the tree
func (t *Tree) Search(key int) bool {
	t.Lock().Lock()
	defer t.Lock.Unlock()

	return search(bst.Root, key)
}

// internal recursive function to search an value in the tree
func search(n *Node, key int) bool {
	if n == nil {
		return
	}

	if key < n.Key {
		return search(n.Left, key)
	}

	if key > n.Key {
		return search(n.Right, key)
	}

	return true
}

// Remove removes the value with key `key` from the tree
func (t *Tree) Remove(key int) {
	t.Lock.Lock()
	defer t.Lock.Unlock()

	remove(t, key)
}

// internal recursive function to remove an value
func remove(node *Node, key int) *Node {
	if node == nil {
		return
	}

	if key < node.Key {
		node.Left = remove(node.Left, key)
		return node
	}

	if key > node.Key {
		node.Right = remove(node.Right, key)
		return node
	}

	if node.Left == nil && node.Rigth == nil {
		node = nil
		return
	}

	if node.Left == nil {
		node = node.Rigth
		return node
	}

	if node.Right == nil {
		node = node.Left
		return node
	}

	leftMostRightSide := node.Right
	for {
		// Find smallest value on the right side.
		if leftMostRightSide != nil && leftMostRightSide.Left != nil {
			leftMostRightSide = leftMostRightSide.Left
		} else {
			break
		}
	}
	node.Key, node.Value = leftMostRightSide.Key, leftMostRightSide.Value
	node.Right = remove(node.Right, node.Key)
	return node
}
