package room

import (
	"strings"
	"sync"
	"testing"
)

func TestRepository_Create(t *testing.T) {
	repo := NewRoomRepository()

	room := repo.Create("Alice")

	if room.ID == "" {
		t.Error("expected room ID to be set")
	}
	if len(room.ID) != 4 {
		t.Errorf("expected 4-character room ID, got %d", len(room.ID))
	}
	if len(room.Players) != 1 {
		t.Errorf("expected 1 player, got %d", len(room.Players))
	}
	if room.Players[0].Name != "Alice" {
		t.Errorf("expected player name 'Alice', got '%s'", room.Players[0].Name)
	}
	if room.Players[0].ID == "" {
		t.Error("expected player ID to be set")
	}
	if room.Players[0].Status != PlayerStatusConnected {
		t.Errorf("expected player status 'connected', got '%s'", room.Players[0].Status)
	}
	if room.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if room.LastActivityAt.IsZero() {
		t.Error("expected LastActivityAt to be set")
	}
	if room.Wins == nil {
		t.Error("expected Wins map to be initialized")
	}
}

func TestRepository_Create_UniqueIDs(t *testing.T) {
	repo := NewRoomRepository()

	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		room := repo.Create("Player")
		if ids[room.ID] {
			t.Errorf("duplicate room ID generated: %s", room.ID)
		}
		ids[room.ID] = true
	}

	if repo.Count() != 100 {
		t.Errorf("expected 100 rooms, got %d", repo.Count())
	}
}

func TestRepository_Get(t *testing.T) {
	repo := NewRoomRepository()

	created := repo.Create("Alice")

	room := repo.Get(created.ID)
	if room == nil {
		t.Fatal("expected to find room")
	}
	if room.ID != created.ID {
		t.Errorf("expected room ID '%s', got '%s'", created.ID, room.ID)
	}
}

func TestRepository_Get_NotFound(t *testing.T) {
	repo := NewRoomRepository()

	room := repo.Get("XXXX")
	if room != nil {
		t.Error("expected nil for nonexistent room")
	}
}

func TestRepository_Get_CaseInsensitive(t *testing.T) {
	repo := NewRoomRepository()

	created := repo.Create("Alice")
	lowercaseID := strings.ToLower(created.ID)

	room := repo.Get(lowercaseID)
	if room == nil {
		t.Fatal("expected case-insensitive lookup to work")
	}
	if room.ID != created.ID {
		t.Errorf("expected room ID '%s', got '%s'", created.ID, room.ID)
	}
}

func TestRepository_GetWithLock(t *testing.T) {
	repo := NewRoomRepository()

	created := repo.Create("Alice")

	room, unlock := repo.GetWithLock(created.ID)
	if room == nil {
		t.Fatal("expected to find room")
	}
	if room.ID != created.ID {
		t.Errorf("expected room ID '%s', got '%s'", created.ID, room.ID)
	}

	// Modify room while holding lock
	room.GamesPlayed = 5
	unlock()

	// Verify modification persisted
	room2 := repo.Get(created.ID)
	if room2.GamesPlayed != 5 {
		t.Errorf("expected GamesPlayed to be 5, got %d", room2.GamesPlayed)
	}
}

func TestRepository_GetWithLock_NotFound(t *testing.T) {
	repo := NewRoomRepository()

	room, unlock := repo.GetWithLock("XXXX")
	if room != nil {
		t.Error("expected nil for nonexistent room")
	}
	// unlock should be a no-op, not panic
	unlock()
}

func TestRepository_GetWithLock_CaseInsensitive(t *testing.T) {
	repo := NewRoomRepository()

	created := repo.Create("Alice")
	lowercaseID := strings.ToLower(created.ID)

	room, unlock := repo.GetWithLock(lowercaseID)
	defer unlock()

	if room == nil {
		t.Fatal("expected case-insensitive lookup to work")
	}
}

func TestRepository_Delete(t *testing.T) {
	repo := NewRoomRepository()

	room := repo.Create("Alice")
	roomID := room.ID

	if repo.Count() != 1 {
		t.Errorf("expected 1 room, got %d", repo.Count())
	}

	repo.Delete(roomID)

	if repo.Count() != 0 {
		t.Errorf("expected 0 rooms after delete, got %d", repo.Count())
	}
	if repo.Get(roomID) != nil {
		t.Error("expected room to be deleted")
	}
}

func TestRepository_Delete_CaseInsensitive(t *testing.T) {
	repo := NewRoomRepository()

	room := repo.Create("Alice")
	lowercaseID := strings.ToLower(room.ID)

	repo.Delete(lowercaseID)

	if repo.Count() != 0 {
		t.Errorf("expected 0 rooms after delete, got %d", repo.Count())
	}
}

func TestRepository_All(t *testing.T) {
	repo := NewRoomRepository()

	repo.Create("Alice")
	repo.Create("Bob")
	repo.Create("Charlie")

	all := repo.All()
	if len(all) != 3 {
		t.Errorf("expected 3 rooms, got %d", len(all))
	}
}

func TestRepository_All_ReturnsCopy(t *testing.T) {
	repo := NewRoomRepository()

	room := repo.Create("Alice")

	all := repo.All()
	// Modifying the returned map should not affect the repository
	delete(all, room.ID)

	if repo.Count() != 1 {
		t.Error("expected original repository to be unchanged")
	}
}

func TestRepository_Replace(t *testing.T) {
	repo := NewRoomRepository()

	repo.Create("Alice")
	repo.Create("Bob")

	// Replace with new rooms
	newRooms := map[string]*Room{
		"TEST": {ID: "TEST", Players: []Player{{ID: "p1", Name: "Charlie"}}},
	}
	repo.Replace(newRooms)

	if repo.Count() != 1 {
		t.Errorf("expected 1 room after replace, got %d", repo.Count())
	}
	if repo.Get("TEST") == nil {
		t.Error("expected new room to exist")
	}
}

func TestRepository_Replace_Nil(t *testing.T) {
	repo := NewRoomRepository()

	repo.Create("Alice")
	repo.Replace(nil)

	if repo.Count() != 0 {
		t.Errorf("expected 0 rooms after nil replace, got %d", repo.Count())
	}
}

func TestRepository_Replace_CreatesLocks(t *testing.T) {
	repo := NewRoomRepository()

	newRooms := map[string]*Room{
		"TEST": {ID: "TEST", Players: []Player{{ID: "p1", Name: "Alice"}}},
	}
	repo.Replace(newRooms)

	// Should be able to get lock for replaced room
	room, unlock := repo.GetWithLock("TEST")
	if room == nil {
		t.Fatal("expected to find room with lock")
	}
	unlock()
}

func TestRepository_Count(t *testing.T) {
	repo := NewRoomRepository()

	if repo.Count() != 0 {
		t.Errorf("expected 0 rooms initially, got %d", repo.Count())
	}

	repo.Create("Alice")
	if repo.Count() != 1 {
		t.Errorf("expected 1 room, got %d", repo.Count())
	}

	repo.Create("Bob")
	if repo.Count() != 2 {
		t.Errorf("expected 2 rooms, got %d", repo.Count())
	}
}

func TestRepository_Concurrent(t *testing.T) {
	repo := NewRoomRepository()

	// Create initial room
	room := repo.Create("Alice")
	roomID := room.ID

	// Run concurrent operations
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(3)

		// Concurrent reads
		go func() {
			defer wg.Done()
			repo.Get(roomID)
		}()

		// Concurrent creates
		go func() {
			defer wg.Done()
			repo.Create("Player")
		}()

		// Concurrent All()
		go func() {
			defer wg.Done()
			repo.All()
		}()
	}
	wg.Wait()

	// Should have 101 rooms (1 initial + 100 created)
	if repo.Count() != 101 {
		t.Errorf("expected 101 rooms, got %d", repo.Count())
	}
}

func TestRepository_GetWithLock_Concurrent(t *testing.T) {
	repo := NewRoomRepository()

	room := repo.Create("Alice")
	roomID := room.ID

	// Multiple goroutines trying to modify the same room
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			r, unlock := repo.GetWithLock(roomID)
			if r != nil {
				r.GamesPlayed++
				unlock()
			}
		}(i)
	}
	wg.Wait()

	// Should have exactly 10 increments
	finalRoom := repo.Get(roomID)
	if finalRoom.GamesPlayed != 10 {
		t.Errorf("expected GamesPlayed to be 10, got %d", finalRoom.GamesPlayed)
	}
}
