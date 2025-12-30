package room

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// PersistenceManager handles saving and loading room state.
type PersistenceManager interface {
	// Load loads rooms from the given file.
	// Returns an empty map if the file doesn't exist.
	Load(filename string) (map[string]*Room, error)

	// Save saves the given rooms to the file.
	Save(filename string, rooms map[string]*Room) error

	// FindStaleRooms returns room IDs that have been inactive longer than maxAge.
	FindStaleRooms(rooms map[string]*Room, maxAge time.Duration) []string
}

// persistenceManager is the concrete implementation of PersistenceManager.
type persistenceManager struct{}

// NewPersistenceManager creates a new PersistenceManager.
func NewPersistenceManager() PersistenceManager {
	return &persistenceManager{}
}

// persistedData is the JSON structure for saving rooms.
type persistedData struct {
	Rooms   map[string]*Room `json:"rooms"`
	SavedAt time.Time        `json:"saved_at"`
	Version int              `json:"version"`
}

func (pm *persistenceManager) Load(filename string) (map[string]*Room, error) {
	data, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		log.Printf("No room data file found at %s, starting fresh", filename)
		return make(map[string]*Room), nil
	}
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		log.Printf("Room data file is empty, starting fresh")
		return make(map[string]*Room), nil
	}

	var pd persistedData
	if err := json.Unmarshal(data, &pd); err != nil {
		return nil, err
	}

	rooms := pd.Rooms
	if rooms == nil {
		rooms = make(map[string]*Room)
	}

	// Ensure Wins maps and LastActivityAt are initialized
	for _, room := range rooms {
		if room.Wins == nil {
			room.Wins = make(map[string]int)
		}
		// For backward compatibility: if LastActivityAt is zero, use CreatedAt
		if room.LastActivityAt.IsZero() {
			room.LastActivityAt = room.CreatedAt
		}
	}

	log.Printf("Loaded %d rooms from %s (saved at %s)", len(rooms), filename, pd.SavedAt.Format(time.RFC3339))
	return rooms, nil
}

func (pm *persistenceManager) Save(filename string, rooms map[string]*Room) error {
	pd := persistedData{
		Rooms:   rooms,
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

	log.Printf("Saved %d rooms to %s", len(rooms), filename)
	return nil
}

func (pm *persistenceManager) FindStaleRooms(rooms map[string]*Room, maxAge time.Duration) []string {
	cutoff := time.Now().Add(-maxAge)
	var stale []string

	for id, room := range rooms {
		if room.LastActivityAt.Before(cutoff) {
			stale = append(stale, id)
		}
	}

	return stale
}
