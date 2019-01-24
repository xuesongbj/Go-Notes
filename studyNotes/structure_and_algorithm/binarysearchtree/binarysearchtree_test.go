package binarysearchtree

import (
	"fmt"
	"testing"
)

var bst Tree

func FillTree(bst *Tree) {
	bst.Insert(8, "8")
	bst.Insert(10, "10")
	bst.Insert(6, "6")
	bst.Insert(3, "3")
	bst.Insert(7, "7")
	bst.Insert(12, "12")
	bst.Insert(5, "5")
	bst.Insert(1, "1")
	bst.Insert(2, "2")
	bst.Insert(17, "17")
	bst.Insert(28, "28")
	bst.Insert(14, "14")
}

// isSameSlice returns true if the 2 slices are identical
func isSameSlice(a, b []string) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestInsert(t *testing.T) {
	FillTree(&bst)

	bst.Insert(100, "100")
}

func TestTraverseAllNodes(t *testing.T) {
	var req = []string{
		"1",
		"2",
		"3",
		"5",
		"6",
		"7",
		"8",
		"10",
		"12",
		"14",
		"17",
		"28",
		"100"}
	rst := make([]string, 0, 10)

	bst.TraverseAllNodes(func(i interface{}) {
		rst = append(rst, fmt.Sprintf("%s", i))
	})

	if !isSameSlice(rst, req) {
		t.Errorf("Traversal order incorrect, got %v", rst)
	}
}
func TestPreOrderTraverse(t *testing.T) {
	var req = []string{
		"8",
		"6",
		"3",
		"1",
		"2",
		"5",
		"7",
		"10",
		"12",
		"17",
		"14",
		"28",
		"100",
	}

	rst := make([]string, 0, 10)
	bst.PreOrderTraverse(func(i interface{}) {
		rst = append(rst, fmt.Sprintf("%s", i))
	})

	if !isSameSlice(rst, req) {
		t.Errorf("Traversal order incorrect, got %v instead of %v", rst, req)
	}
}

func TestPostOrderTraverse(t *testing.T) {
	var req = []string{
		"2",
		"1",
		"5",
		"3",
		"7",
		"6",
		"14",
		"100",
		"28",
		"17",
		"12",
		"10",
		"8",
	}
	rst := make([]string, 0, 10)
	bst.PostOrderTraverse(func(i interface{}) {
		rst = append(rst, fmt.Sprintf("%s", i))
	})

	if !isSameSlice(rst, req) {
		t.Errorf("Traversal order incorrect, got %v instead of %v", rst, req)
	}
}

func TestMin(t *testing.T) {
	const mixNum = "1"

	min := bst.Min()
	if min.(string) != "1" {
		t.Errorf("Min should be 1.")
	}
}
func TestMax(t *testing.T) {
	const maxNum = "100"

	max := bst.Max()
	if max.(string) != "100" {
		t.Errorf("Max shuld be 100.")
	}
}

func TestSearch(t *testing.T) {
	if !bst.Search(1) || !bst.Search(8) || !bst.Search(12) {
		t.Errorf("Search not working")
	}
}

func TestRemove(t *testing.T) {
	bst.Remove(1)

	const minNum = "2"

	min := bst.Min()
	if min.(string) != "2" {
		t.Errorf("min should be 2")
	}
}
