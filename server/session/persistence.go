// Package session provides multiplayer game session management.
package session

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

const (
	// DefaultDataFile is the default path for session data persistence.
	DefaultDataFile = "sessions.json"
	// DefaultAutoSaveInterval is the default interval for auto-saving sessions.
	DefaultAutoSaveInterval = 30 * time.Second
	// DefaultCleanupInterval is the default interval for cleaning up stale sessions.
	DefaultCleanupInterval = 1 * time.Hour
	// DefaultSessionMaxAge is the default max age before a session is cleaned up.
	DefaultSessionMaxAge = 24 * time.Hour
)

// persistedData is the JSON structure for saving sessions.
type persistedData struct {
	Sessions  map[string]*Session `json:"sessions"`
	SavedAt   time.Time           `json:"saved_at"`
	Version   int                 `json:"version"`
}

// Load loads sessions from the data file.
// Returns nil if the file doesn't exist or is empty.
func (store *Store) Load(filename string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	data, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		log.Printf("No session data file found at %s, starting fresh", filename)
		return nil
	}
	if err != nil {
		return err
	}

	if len(data) == 0 {
		log.Printf("Session data file is empty, starting fresh")
		return nil
	}

	var pd persistedData
	if err := json.Unmarshal(data, &pd); err != nil {
		return err
	}

	store.sessions = pd.Sessions
	if store.sessions == nil {
		store.sessions = make(map[string]*Session)
	}

	// Ensure Wins maps and LastActivityAt are initialized
	for _, sess := range store.sessions {
		if sess.Wins == nil {
			sess.Wins = make(map[string]int)
		}
		// For backward compatibility: if LastActivityAt is zero, use CreatedAt
		if sess.LastActivityAt.IsZero() {
			sess.LastActivityAt = sess.CreatedAt
		}
	}

	log.Printf("Loaded %d sessions from %s (saved at %s)", len(store.sessions), filename, pd.SavedAt.Format(time.RFC3339))
	return nil
}

// Save saves all sessions to the data file.
func (store *Store) Save(filename string) error {
	store.mu.RLock()
	defer store.mu.RUnlock()

	pd := persistedData{
		Sessions: store.sessions,
		SavedAt:  time.Now(),
		Version:  1,
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

	log.Printf("Saved %d sessions to %s", len(store.sessions), filename)
	return nil
}

// StartAutoSave starts a goroutine that periodically saves sessions.
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

// CleanupStaleSessions removes sessions that have been inactive for longer than maxAge.
// Returns the number of sessions removed.
func (store *Store) CleanupStaleSessions(maxAge time.Duration) int {
	store.mu.Lock()
	defer store.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	removed := 0

	for id, sess := range store.sessions {
		if sess.LastActivityAt.Before(cutoff) {
			delete(store.sessions, id)
			removed++
		}
	}

	if removed > 0 {
		log.Printf("Cleaned up %d stale sessions (inactive for >%v)", removed, maxAge)
	}

	return removed
}

// StartCleanup starts a goroutine that periodically removes stale sessions.
// Returns a channel that should be closed to stop cleanup.
func (store *Store) StartCleanup(interval, maxAge time.Duration) chan struct{} {
	stop := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				store.CleanupStaleSessions(maxAge)
			case <-stop:
				return
			}
		}
	}()

	return stop
}
