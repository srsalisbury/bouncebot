package astar

import (
	"testing"

	"github.com/srsalisbury/bouncebot/model"
)

func TestPriorityQueue(t *testing.T) {
	pq := NewPriorityQueue()

	// Create nodes with different priorities
	// priority = len(moves) + heuristic
	node1 := &stateNode{
		bots:      map[model.BotId]model.Position{0: {X: 0, Y: 0}},
		moves:     make([]model.BotPosition, 3), // g=3
		heuristic: 2,                            // h=2, f=5
	}
	node2 := &stateNode{
		bots:      map[model.BotId]model.Position{0: {X: 1, Y: 1}},
		moves:     make([]model.BotPosition, 1), // g=1
		heuristic: 1,                            // h=1, f=2
	}
	node3 := &stateNode{
		bots:      map[model.BotId]model.Position{0: {X: 2, Y: 2}},
		moves:     make([]model.BotPosition, 2), // g=2
		heuristic: 5,                            // h=5, f=7
	}

	// Add in non-sorted order
	pq.Enqueue(node1) // f=5
	pq.Enqueue(node3) // f=7
	pq.Enqueue(node2) // f=2

	if pq.Len() != 3 {
		t.Errorf("Expected length 3, got %d", pq.Len())
	}

	// Should dequeue in priority order (lowest first)
	first := pq.Dequeue()
	if first.priority() != 2 {
		t.Errorf("Expected first priority 2, got %d", first.priority())
	}

	second := pq.Dequeue()
	if second.priority() != 5 {
		t.Errorf("Expected second priority 5, got %d", second.priority())
	}

	third := pq.Dequeue()
	if third.priority() != 7 {
		t.Errorf("Expected third priority 7, got %d", third.priority())
	}

	if !pq.IsEmpty() {
		t.Error("Expected queue to be empty")
	}

	// Dequeue from empty should return nil
	if pq.Dequeue() != nil {
		t.Error("Expected nil from empty queue")
	}
}
