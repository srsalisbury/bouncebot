package daily

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

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
	dataDir string
	mu      sync.RWMutex
	cache   map[string]UserProgress // playerID -> progress
}

// NewProgressManager creates a new progress manager.
func NewProgressManager(dataDir string) *ProgressManager {
	return &ProgressManager{
		dataDir: dataDir,
		cache:   make(map[string]UserProgress),
	}
}

// GetUserProgress loads progress for a player.
func (pm *ProgressManager) GetUserProgress(playerID string) (UserProgress, error) {
	// Check cache first
	pm.mu.RLock()
	if progress, ok := pm.cache[playerID]; ok {
		pm.mu.RUnlock()
		return progress, nil
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
	pm.cache[playerID] = progress
	pm.mu.Unlock()

	return progress, nil
}

// SaveUserProgress saves progress for a player.
func (pm *ProgressManager) SaveUserProgress(playerID string, progress UserProgress) error {
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
	pm.cache[playerID] = progress
	pm.mu.Unlock()

	return nil
}

// MarkSolved marks a puzzle as solved for a player.
func (pm *ProgressManager) MarkSolved(playerID, date, difficulty string) error {
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

// IsSolved checks if a puzzle has been solved.
func (pm *ProgressManager) IsSolved(playerID, date, difficulty string) (bool, error) {
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
