package astar

import "container/heap"

// PriorityQueue implements a min-heap for A* search.
// Nodes with lower priority values are dequeued first.
//
// Go's heap package provides heap operations, but requires you to implement
// the heap.Interface, which combines sort.Interface (Len, Less, Swap) with
// Push and Pop methods.
//
// The heap is a min-heap by default (smallest first), which is what we want
// for A* since we expand the node with the lowest f-score first.
type PriorityQueue struct {
	nodes []*stateNode
}

// Len returns the number of elements in the queue.
// Required by sort.Interface.
func (pq *PriorityQueue) Len() int {
	return len(pq.nodes)
}

// Less reports whether element i should come before element j.
// For a min-heap, we return true if i has lower priority (lower f-score).
// Required by sort.Interface.
func (pq *PriorityQueue) Less(i, j int) bool {
	// Lower heuristic = higher priority = should come first
	return pq.nodes[i].priority() < pq.nodes[j].priority()
}

// Swap exchanges elements i and j.
// Required by sort.Interface.
func (pq *PriorityQueue) Swap(i, j int) {
	pq.nodes[i], pq.nodes[j] = pq.nodes[j], pq.nodes[i]
}

// Push adds an element to the heap.
// Required by heap.Interface.
// Note: This is called by heap.Push, not directly. The parameter is interface{}
// because that's what heap.Interface requires.
func (pq *PriorityQueue) Push(x any) {
	pq.nodes = append(pq.nodes, x.(*stateNode))
}

// Pop removes and returns the last element (heap.Pop will swap min to end first).
// Required by heap.Interface.
// Note: heap.Pop calls this after moving the min element to the end,
// so we just remove and return the last element.
func (pq *PriorityQueue) Pop() any {
	old := pq.nodes
	n := len(old)
	node := old[n-1]
	old[n-1] = nil // avoid memory leak
	pq.nodes = old[0 : n-1]
	return node
}

// NewPriorityQueue creates an empty priority queue.
func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{
		nodes: make([]*stateNode, 0),
	}
	heap.Init(pq)
	return pq
}

// Enqueue adds a node to the priority queue.
// This is a convenience wrapper around heap.Push.
func (pq *PriorityQueue) Enqueue(node *stateNode) {
	heap.Push(pq, node)
}

// Dequeue removes and returns the highest-priority (lowest f-score) node.
// Returns nil if the queue is empty.
// This is a convenience wrapper around heap.Pop.
func (pq *PriorityQueue) Dequeue() *stateNode {
	if pq.Len() == 0 {
		return nil
	}
	return heap.Pop(pq).(*stateNode)
}

// IsEmpty returns true if the queue has no elements.
func (pq *PriorityQueue) IsEmpty() bool {
	return pq.Len() == 0
}
