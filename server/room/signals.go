package room

// Signal represents an action that should be taken by the orchestrator.
// Using a sealed interface pattern for type safety.
type Signal interface {
	signalMarker() // unexported method makes this a sealed interface
}

// BroadcastSignal indicates an event should be broadcast to clients.
type BroadcastSignal struct {
	Event BroadcastEvent
}

func (BroadcastSignal) signalMarker() {}

// EndGameSignal indicates the current game should end.
type EndGameSignal struct {
	RoomID string
}

func (EndGameSignal) signalMarker() {}

// StartNextGameSignal indicates the next game should start.
type StartNextGameSignal struct {
	RoomID string
}

func (StartNextGameSignal) signalMarker() {}

// StartTimerSignal indicates a disconnect timer should be started.
type StartTimerSignal struct {
	RoomID   string
	PlayerID string
}

func (StartTimerSignal) signalMarker() {}

// CancelTimerSignal indicates a disconnect timer should be cancelled.
type CancelTimerSignal struct {
	PlayerID string
}

func (CancelTimerSignal) signalMarker() {}

// BroadcastEvent is the specific event type to broadcast.
// Using a sealed interface pattern for type safety.
type BroadcastEvent interface {
	broadcastEventMarker() // unexported method makes this a sealed interface
}

// PlayerJoinedEvent is broadcast when a player joins a room.
type PlayerJoinedEvent struct {
	RoomID     string
	PlayerID   string
	PlayerName string
}

func (PlayerJoinedEvent) broadcastEventMarker() {}

// PlayerLeftEvent is broadcast when a player leaves a room.
type PlayerLeftEvent struct {
	RoomID   string
	PlayerID string
}

func (PlayerLeftEvent) broadcastEventMarker() {}

// GameStartedEvent is broadcast when a new game starts.
type GameStartedEvent struct {
	RoomID string
}

func (GameStartedEvent) broadcastEventMarker() {}

// PlayerFinishedSolvingEvent is broadcast when a player is done looking for solutions.
type PlayerFinishedSolvingEvent struct {
	RoomID   string
	PlayerID string
}

func (PlayerFinishedSolvingEvent) broadcastEventMarker() {}

// PlayerReadyForNextEvent is broadcast when a player is ready for the next game.
type PlayerReadyForNextEvent struct {
	RoomID   string
	PlayerID string
}

func (PlayerReadyForNextEvent) broadcastEventMarker() {}

// PlayerSolvedEvent is broadcast when a player submits a solution.
type PlayerSolvedEvent struct {
	RoomID    string
	PlayerID  string
	MoveCount int
}

func (PlayerSolvedEvent) broadcastEventMarker() {}

// SolutionRetractedEvent is broadcast when a player retracts their solution.
type SolutionRetractedEvent struct {
	RoomID   string
	PlayerID string
}

func (SolutionRetractedEvent) broadcastEventMarker() {}

// GameEndedEvent is broadcast when the game ends.
type GameEndedEvent struct {
	RoomID     string
	WinnerID   string
	WinnerName string
	Moves      []MovePayload
}

func (GameEndedEvent) broadcastEventMarker() {}
