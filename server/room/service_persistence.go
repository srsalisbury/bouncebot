package room

import (
	"log"
	"time"
)

// ---- Persistence Methods ----

// Load loads rooms from the data file.
func (s *RoomService) Load(filename string) error {
	rooms, err := s.persistence.Load(filename)
	if err != nil {
		return err
	}
	s.repo.Replace(rooms)
	return nil
}

// Save saves all rooms to the data file.
func (s *RoomService) Save(filename string) error {
	return s.persistence.Save(filename, s.repo.All())
}

// StartAutoSave starts a goroutine that periodically saves rooms.
func (s *RoomService) StartAutoSave(filename string, interval time.Duration) chan struct{} {
	stop := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := s.Save(filename); err != nil {
					log.Printf("Auto-save failed: %v", err)
				}
			case <-stop:
				// Final save before stopping
				if err := s.Save(filename); err != nil {
					log.Printf("Final save failed: %v", err)
				}
				return
			}
		}
	}()

	return stop
}

// CleanupStaleRooms removes rooms that have been inactive for longer than maxAge.
func (s *RoomService) CleanupStaleRooms(maxAge time.Duration) int {
	stale := s.persistence.FindStaleRooms(s.repo.All(), maxAge)
	for _, id := range stale {
		s.repo.Delete(id)
	}

	if len(stale) > 0 {
		log.Printf("Cleaned up %d stale rooms (inactive for >%v)", len(stale), maxAge)
	}

	return len(stale)
}

// StartCleanup starts a goroutine that periodically removes stale rooms.
func (s *RoomService) StartCleanup(interval, maxAge time.Duration) chan struct{} {
	stop := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.CleanupStaleRooms(maxAge)
			case <-stop:
				return
			}
		}
	}()

	return stop
}
