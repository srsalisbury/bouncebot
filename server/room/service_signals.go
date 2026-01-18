package room

// processSignals interprets and executes signals.
// This is where the orchestration happens.
func (s *RoomService) processSignals(signals []Signal) {
	for _, sig := range signals {
		switch signal := sig.(type) {
		case BroadcastSignal:
			s.processBroadcast(signal.Event)

		case EndGameSignal:
			_ = s.withRoomLock(signal.RoomID, func(room *Room) ([]Signal, error) {
				return s.gameMgr.EndGame(room), nil
			})

		case StartNextGameSignal:
			room, unlock := s.repo.GetWithLock(signal.RoomID)
			if room != nil {
				newSignals := s.gameMgr.StartNextGame(room)
				// Make a copy for callback (room might be modified after unlock)
				roomCopy := *room
				unlock()
				s.processSignals(newSignals)
				// Notify about game start (for solver etc)
				if s.onGameStart != nil && roomCopy.CurrentGame != nil {
					s.onGameStart(&roomCopy)
				}
			} else {
				unlock()
			}

		case StartTimerSignal:
			s.timerMgr.StartTimer(
				signal.RoomID,
				signal.PlayerID,
				s.disconnectGracePeriod,
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
