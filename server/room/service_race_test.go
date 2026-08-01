package room

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/srsalisbury/bouncebot/model"
)

// TestService_StartNextGame_DoesNotDoubleFireOnConcurrentLeave reproduces a
// race in the "everyone ready -> start next round" transition. Starting the
// next round splits into two lock acquisitions around a potentially slow
// board search (selectNewGame unlocks before calling generateBoard, then
// re-locks to commit) - see service_signals.go's StartNextGameSignal case.
// room.ReadyForNext isn't cleared until the commit, so if a player who was
// already in ReadyForNext leaves while a round is mid-generation, the
// leave's own equality check (len(ReadyForNext) == len(Players)) is
// re-satisfied and fires a second, fully independent StartNextGameSignal -
// generating and committing a second board on top of the first.
func TestService_StartNextGame_DoesNotDoubleFireOnConcurrentLeave(t *testing.T) {
	svc := NewRoomService()

	room, aliceID, aliceToken := svc.Create("Alice", false)
	_ = aliceID
	_, _, bobToken, err := svc.Join(room.ID, "Bob")
	if err != nil {
		t.Fatalf("join bob: %v", err)
	}
	_, _, carolToken, err := svc.Join(room.ID, "Carol")
	if err != nil {
		t.Fatalf("join carol: %v", err)
	}

	var startedCount int32
	svc.SetOnGameStart(func(r *Room) { atomic.AddInt32(&startedCount, 1) })

	// Start the first game before configuring a minimum solution length, so
	// it takes generateBoard's fast (no-solve-call) path.
	if _, err := svc.StartGame(room.ID); err != nil {
		t.Fatalf("start game: %v", err)
	}
	atomic.StoreInt32(&startedCount, 0) // discount the initial StartGame

	// A minimum solution length above the fast-path threshold forces
	// generateBoard to actually call the (injected) solve function - this is
	// what widens the unlocked window during the *next* round's generation.
	if err := svc.UpdateRoomSettings(room.ID, aliceToken, RoomSettings{MinSolutionLength: 2}); err != nil {
		t.Fatalf("update settings: %v", err)
	}

	var solveCalls int32
	generating := make(chan struct{})
	release := make(chan struct{})
	svc.SetBoardSolveFunc(func(game *model.Game) ([]model.BotPosition, bool) {
		if atomic.AddInt32(&solveCalls, 1) == 1 {
			close(generating)
			<-release
		}
		return []model.BotPosition{{}, {}}, true
	})

	for _, tok := range []string{aliceToken, bobToken, carolToken} {
		if err := svc.MarkFinishedSolving(room.ID, tok); err != nil {
			t.Fatalf("mark finished: %v", err)
		}
	}

	if err := svc.MarkReadyForNext(room.ID, aliceToken); err != nil {
		t.Fatalf("alice ready: %v", err)
	}
	if err := svc.MarkReadyForNext(room.ID, bobToken); err != nil {
		t.Fatalf("bob ready: %v", err)
	}

	// Carol is the last to ready up, so this call synchronously drives
	// selectNewGame -> generateBoard, which blocks (via the injected solve
	// delay) until we release it below.
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := svc.MarkReadyForNext(room.ID, carolToken); err != nil {
			t.Error(err)
		}
	}()

	select {
	case <-generating:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for board generation to start")
	}

	// While the first round-start is still mid-generation (ReadyForNext not
	// yet cleared), Alice leaves. She was in ReadyForNext, so removing her
	// shrinks both ReadyForNext and Players by one, re-satisfying the
	// equality check and firing a second StartNextGameSignal.
	if err := svc.LeaveRoom(room.ID, aliceToken); err != nil {
		t.Fatalf("leave room: %v", err)
	}

	close(release)
	<-done

	if got := atomic.LoadInt32(&startedCount); got != 1 {
		t.Errorf("expected exactly 1 game start for a single ready-up transition, got %d (board was generated/committed more than once)", got)
	}
}
