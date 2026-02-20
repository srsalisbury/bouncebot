package daily

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

var validPlayerID = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

func validatePlayerID(playerID string) error {
	if !validPlayerID.MatchString(playerID) {
		return fmt.Errorf("invalid player ID")
	}
	return nil
}

// DayProgress tracks which puzzles a user has solved for a given day.
type DayProgress struct {
	Easy   bool `json:"easy"`
	Medium bool `json:"medium"`
	Hard   bool `json:"hard"`
}

// UserProgress maps dates to day progress.
type UserProgress map[string]DayProgress

// ProgressManager handles user progress storage and retrieval.
type ProgressManager struct {
	dataDir     string
	mu          sync.RWMutex
	cache       map[string]progressCacheEntry // playerID -> cached progress
	playerLocks sync.Map                      // playerID -> *sync.Mutex
}

type progressCacheEntry struct {
	progress     UserProgress
	lastAccessed time.Time
}

const maxProgressCacheSize = 1000

// NewProgressManager creates a new progress manager.
func NewProgressManager(dataDir string) *ProgressManager {
	return &ProgressManager{
		dataDir: dataDir,
		cache:   make(map[string]progressCacheEntry),
	}
}

func (pm *ProgressManager) getPlayerLock(playerID string) *sync.Mutex {
	v, _ := pm.playerLocks.LoadOrStore(playerID, &sync.Mutex{})
	return v.(*sync.Mutex)
}

// GetUserProgress loads progress for a player.
func (pm *ProgressManager) GetUserProgress(playerID string) (UserProgress, error) {
	if err := validatePlayerID(playerID); err != nil {
		return nil, err
	}

	// Check cache first
	pm.mu.RLock()
	if entry, ok := pm.cache[playerID]; ok {
		pm.mu.RUnlock()
		// Update last accessed time
		pm.mu.Lock()
		entry.lastAccessed = time.Now()
		pm.cache[playerID] = entry
		pm.mu.Unlock()
		return entry.progress, nil
	}
	pm.mu.RUnlock()

	// Load from disk
	path := pm.progressPath(playerID)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(UserProgress), nil
		}
		return nil, fmt.Errorf("failed to read progress file: %w", err)
	}

	var progress UserProgress
	if err := json.Unmarshal(data, &progress); err != nil {
		return nil, fmt.Errorf("failed to parse progress file: %w", err)
	}

	// Cache it
	pm.mu.Lock()
	pm.cache[playerID] = progressCacheEntry{
		progress:     progress,
		lastAccessed: time.Now(),
	}
	pm.evictCacheLocked()
	pm.mu.Unlock()

	return progress, nil
}

// SaveUserProgress saves progress for a player.
func (pm *ProgressManager) SaveUserProgress(playerID string, progress UserProgress) error {
	if err := validatePlayerID(playerID); err != nil {
		return err
	}

	path := pm.progressPath(playerID)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create progress directory: %w", err)
	}

	data, err := json.MarshalIndent(progress, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal progress: %w", err)
	}

	// Write atomically using temp file
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp progress file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename progress file: %w", err)
	}

	// Update cache
	pm.mu.Lock()
	pm.cache[playerID] = progressCacheEntry{
		progress:     progress,
		lastAccessed: time.Now(),
	}
	pm.evictCacheLocked()
	pm.mu.Unlock()

	return nil
}

// MarkSolved marks a puzzle as solved for a player.
func (pm *ProgressManager) MarkSolved(playerID, date, difficulty string) error {
	if err := validatePlayerID(playerID); err != nil {
		return err
	}

	// Per-player lock prevents concurrent read-modify-write races
	lock := pm.getPlayerLock(playerID)
	lock.Lock()
	defer lock.Unlock()

	progress, err := pm.GetUserProgress(playerID)
	if err != nil {
		return err
	}

	dayProgress := progress[date]
	switch difficulty {
	case DifficultyEasy:
		dayProgress.Easy = true
	case DifficultyMedium:
		dayProgress.Medium = true
	case DifficultyHard:
		dayProgress.Hard = true
	default:
		return fmt.Errorf("invalid difficulty: %s", difficulty)
	}
	progress[date] = dayProgress

	return pm.SaveUserProgress(playerID, progress)
}

// evictCacheLocked removes stale entries from the progress cache.
// Must be called with pm.mu held for writing.
func (pm *ProgressManager) evictCacheLocked() {
	if len(pm.cache) <= maxProgressCacheSize {
		return
	}
	// Evict entries not accessed in the last hour
	cutoff := time.Now().Add(-1 * time.Hour)
	for id, entry := range pm.cache {
		if entry.lastAccessed.Before(cutoff) {
			delete(pm.cache, id)
		}
	}
	// If still over limit, clear all
	if len(pm.cache) > maxProgressCacheSize {
		pm.cache = make(map[string]progressCacheEntry)
	}
}

// IsSolved checks if a puzzle has been solved.
func (pm *ProgressManager) IsSolved(playerID, date, difficulty string) (bool, error) {
	if err := validatePlayerID(playerID); err != nil {
		return false, err
	}

	progress, err := pm.GetUserProgress(playerID)
	if err != nil {
		return false, err
	}

	dayProgress, ok := progress[date]
	if !ok {
		return false, nil
	}

	switch difficulty {
	case DifficultyEasy:
		return dayProgress.Easy, nil
	case DifficultyMedium:
		return dayProgress.Medium, nil
	case DifficultyHard:
		return dayProgress.Hard, nil
	default:
		return false, fmt.Errorf("invalid difficulty: %s", difficulty)
	}
}

// progressPath returns the file path for a player's progress file.
// Uses first 2 chars of UUID as subdirectory for sharding.
func (pm *ProgressManager) progressPath(playerID string) string {
	prefix := playerID
	if len(prefix) >= 2 {
		prefix = prefix[:2]
	}
	return filepath.Join(pm.dataDir, "users", prefix, playerID+".json")
}
