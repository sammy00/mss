package mss

import (
	"errors"
	"fmt"
	"math"

	"github.com/sammy00/mss/container/stack"
	"github.com/sammy00/mss/ots/winternitz"
)

// Node is a node in the Merkle tree
type Node struct {
	height uint32
	//index  int
	nu []byte
}

// String generates the string representation of node
func (node *Node) String() string {
	return fmt.Sprintf("{height: %v}", node.height)
}

// TreeHashStack is a stack tracing the running state
//	of the tree hash algo
type TreeHashStack struct {
	leaf      uint32       // zero-based index of starting leaf
	leafUpper uint32       // the global upper bound of leaves for this tree hash instance
	height    uint32       // height of the targeted Merkle tree
	nodeStack *stack.Stack // stacks store nodes for each Merkle sub-tree of height from 0 to H-1
}

// NewTreeHashStack makes a new tree hash instance
func NewTreeHashStack(startingLeaf, h uint32) *TreeHashStack {
	treeHashStack := new(TreeHashStack)
	treeHashStack.Init(startingLeaf, h)

	return treeHashStack
}

// Init initializes the tree hash instance to target specific height
//	and the range of leaves
func (th *TreeHashStack) Init(startingLeaf, h uint32) error {
	if 1 != startingLeaf%2 {
		return errors.New("invalid index of starting leaf")
	}

	th.leaf, th.leafUpper, th.height = startingLeaf, startingLeaf+(1<<h), h
	//th.leaf, th.height = startingLeaf, h
	th.nodeStack = stack.New() // clear up the stack

	return nil
}

// IsCompleted checks if the tree hash instance has completed
func (th *TreeHashStack) IsCompleted() bool {
	//return th.leaf >= uint32(1<<th.height)
	return th.leaf >= th.leafUpper
}

// LowestTailHeight returns the lowest height of tail nodes
//	in this tree hash instance
func (th *TreeHashStack) LowestTailHeight() uint32 {
	if th.nodeStack.Empty() {
		return th.height
	}
	if th.IsCompleted() {
		return math.MaxUint32
	}
	return th.Top().height
}

// Top returns the node in the top of the stack
func (th *TreeHashStack) Top() *Node {
	if th.nodeStack.Empty() {
		return nil
	}

	return th.nodeStack.Peek().(*Node)
}

// Update executes numOp updates on the instance, and
//	add on the new leaf derived by keyItr if necessary
func (th *TreeHashStack) Update(numOp uint32, keyItr *winternitz.SkPkIterator) {
	if th.IsCompleted() {
		return
	}

	for (numOp > 0) && !th.IsCompleted() {
		// may have nodes at the same height to merge
		if th.nodeStack.Len() >= 2 {
			e1, e2 := th.nodeStack.Peek2()
			node1 := e1.(*Node)
			node2 := e2.(*Node)

			// merge the nodes at the same height
			if node1.height == node2.height {
				th.nodeStack.Pop()
				th.nodeStack.Pop()

				th.nodeStack.Push(&Node{node1.height + 1, merge(node2.nu, node1.nu)})

				numOp--
				continue
			}
		}

		// invoke key generator to make a new leaf and
		//	add the new leaf to S
		sk, _ := keyItr.Next()
		th.nodeStack.Push(&Node{0, winternitz.HashPk(&sk.PublicKey)})
		th.leaf++
		numOp--
	}
}
