package room

import "log"

// processSignals interprets and executes signals.
// This is where the orchestration happens.
func (s *RoomService) processSignals(signals []Signal) {
	for _, sig := range signals {
		switch signal := sig.(type) {
		case BroadcastSignal:
			s.processBroadcast(signal.Event)

		case EndGameSignal:
			if err := s.withRoomLock(signal.RoomID, func(room *Room) ([]Signal, error) {
				return s.gameMgr.EndGame(room), nil
			}); err != nil {
				// Room vanished between the round ending and this signal being
				// processed (e.g. every player left during the gap) - the room
				// is gone either way, so there's nothing left to end, but log it
				// since it'd otherwise be an invisible dead end for whoever's
				// debugging a "stuck round" report.
				log.Printf("room %s: dropped EndGameSignal: %v", signal.RoomID, err)
			}

		case StartNextGameSignal:
			game, _, ok := s.selectNewGame(signal.RoomID)
			if !ok {
				// Room vanished while board generation was running unlocked -
				// same reasoning as EndGameSignal above. Note: room.NextGameStarting
				// is left permanently true, but harmlessly so, since the room
				// (and everyone in it) is already gone.
				log.Printf("room %s: dropped StartNextGameSignal: room not found during selectNewGame", signal.RoomID)
				continue
			}

			room, unlock := s.repo.GetWithLock(signal.RoomID)
			if room == nil {
				unlock()
				log.Printf("room %s: dropped StartNextGameSignal: room not found after selectNewGame", signal.RoomID)
				continue
			}
			s.gameMgr.PromotePendingPlayers(room)
			newSignals := s.gameMgr.CommitNewGame(room, game)
			// Make a copy for callback (room might be modified after unlock)
			roomCopy := *room
			unlock()
			s.processSignals(newSignals)
			// Notify about game start (for solver etc)
			if s.onGameStart != nil && roomCopy.CurrentGame != nil {
				s.onGameStart(&roomCopy)
			}

		case StartTimerSignal:
			gracePeriod := s.disconnectGracePeriod
			room, unlock := s.repo.GetWithLock(signal.RoomID)
			if room != nil && room.IsSinglePlayer {
				gracePeriod = s.soloDisconnectGracePeriod
			}
			unlock()
			s.timerMgr.StartTimer(
				signal.RoomID,
				signal.PlayerID,
				gracePeriod,
				s.onTimerFired,
			)

		case CancelTimerSignal:
			s.timerMgr.CancelTimer(signal.PlayerID)
		}
	}
}

func (s *RoomService) processBroadcast(event BroadcastEvent) {
	if s.broadcaster == nil {
		return
	}

	switch e := event.(type) {
	case PlayerJoinedEvent:
		s.broadcaster.BroadcastPlayerJoined(e.RoomID, e.PlayerID, e.PlayerName)
	case PlayerLeftEvent:
		s.broadcaster.BroadcastPlayerLeft(e.RoomID, e.PlayerID)
	case GameStartedEvent:
		s.broadcaster.BroadcastGameStarted(e.RoomID)
	case PlayerFinishedSolvingEvent:
		s.broadcaster.BroadcastPlayerFinishedSolving(e.RoomID, e.PlayerID)
	case PlayerReadyForNextEvent:
		s.broadcaster.BroadcastPlayerReadyForNext(e.RoomID, e.PlayerID)
	case PlayerSolvedEvent:
		s.broadcaster.BroadcastPlayerSolved(e.RoomID, e.PlayerID, e.MoveCount)
	case GameEndedEvent:
		s.broadcaster.BroadcastGameEnded(e.RoomID, e.WinnerID, e.WinnerName, e.Moves)
	case RoomSettingsChangedEvent:
		s.broadcaster.BroadcastRoomSettingsChanged(e.RoomID)
	}
}

func (s *RoomService) onTimerFired(roomID, playerID string) {
	s.RemovePlayer(roomID, playerID)
}

// withRoomLock acquires a lock on the room, executes fn, and processes any signals.
// Returns ErrRoomNotFound if the room doesn't exist.
// If fn returns an error, signals are not processed and the error is returned.
// The lock is released before processing signals to avoid deadlocks.
func (s *RoomService) withRoomLock(roomID string, fn func(room *Room) ([]Signal, error)) error {
	room, unlock := s.repo.GetWithLock(roomID)
	if room == nil {
		unlock()
		return ErrRoomNotFound
	}

	signals, err := fn(room)
	unlock() // Release lock before processing signals (signals may acquire locks)

	if err != nil {
		return err
	}

	s.processSignals(signals)
	return nil
}
