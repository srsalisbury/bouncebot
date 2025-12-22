// Package room provides multiplayer game room management.
package room

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// persistedData is the JSON structure for saving rooms.
type persistedData struct {
	Rooms   map[string]*Room `json:"rooms"`
	SavedAt time.Time        `json:"saved_at"`
	Version int              `json:"version"`
}

// Load loads rooms from the data file.
// Returns nil if the file doesn't exist or is empty.
func (store *Store) Load(filename string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	data, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		log.Printf("No room data file found at %s, starting fresh", filename)
		return nil
	}
	if err != nil {
		return err
	}

	if len(data) == 0 {
		log.Printf("Room data file is empty, starting fresh")
		return nil
	}

	var pd persistedData
	if err := json.Unmarshal(data, &pd); err != nil {
		return err
	}

	store.rooms = pd.Rooms
	if store.rooms == nil {
		store.rooms = make(map[string]*Room)
	}

	// Ensure Wins maps and LastActivityAt are initialized
	for _, room := range store.rooms {
		if room.Wins == nil {
			room.Wins = make(map[string]int)
		}
		// For backward compatibility: if LastActivityAt is zero, use CreatedAt
		if room.LastActivityAt.IsZero() {
			room.LastActivityAt = room.CreatedAt
		}
	}

	log.Printf("Loaded %d rooms from %s (saved at %s)", len(store.rooms), filename, pd.SavedAt.Format(time.RFC3339))
	return nil
}

// Save saves all rooms to the data file.
func (store *Store) Save(filename string) error {
	store.mu.RLock()
	defer store.mu.RUnlock()

	pd := persistedData{
		Rooms:   store.rooms,
		SavedAt: time.Now(),
		Version: 1,
	}

	data, err := json.MarshalIndent(pd, "", "  ")
	if err != nil {
		return err
	}

	// Write to temp file first, then rename for atomicity
	tmpFile := filename + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	if err := os.Rename(tmpFile, filename); err != nil {
		os.Remove(tmpFile)
		return err
	}

	log.Printf("Saved %d rooms to %s", len(store.rooms), filename)
	return nil
}

// StartAutoSave starts a goroutine that periodically saves rooms.
// Returns a channel that should be closed to stop auto-saving.
func (store *Store) StartAutoSave(filename string, interval time.Duration) chan struct{} {
	stop := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := store.Save(filename); err != nil {
					log.Printf("Auto-save failed: %v", err)
				}
			case <-stop:
				// Final save before stopping
				if err := store.Save(filename); err != nil {
					log.Printf("Final save failed: %v", err)
				}
				return
			}
		}
	}()

	return stop
}

// CleanupStaleRooms removes rooms that have been inactive for longer than maxAge.
// Returns the number of rooms removed.
func (store *Store) CleanupStaleRooms(maxAge time.Duration) int {
	store.mu.Lock()
	defer store.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	removed := 0

	for id, room := range store.rooms {
		if room.LastActivityAt.Before(cutoff) {
			delete(store.rooms, id)
			removed++
		}
	}

	if removed > 0 {
		log.Printf("Cleaned up %d stale rooms (inactive for >%v)", removed, maxAge)
	}

	return removed
}

// StartCleanup starts a goroutine that periodically removes stale rooms.
// Returns a channel that should be closed to stop cleanup.
func (store *Store) StartCleanup(interval, maxAge time.Duration) chan struct{} {
	stop := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				store.CleanupStaleRooms(maxAge)
			case <-stop:
				return
			}
		}
	}()

	return stop
}
