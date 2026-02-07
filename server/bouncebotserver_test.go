package main

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/srsalisbury/bouncebot/model"
	pb "github.com/srsalisbury/bouncebot/proto"
	"github.com/srsalisbury/bouncebot/server/room"
)

func TestBounceBotServer_CreateRoom(t *testing.T) {
	svc := room.NewRoomService()
	server := NewBounceBotServer(svc)

	req := connect.NewRequest(&pb.CreateRoomRequest{
		PlayerName:     "Alice",
		IsSinglePlayer: false,
	})

	resp, err := server.CreateRoom(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Msg.Room.Id == "" {
		t.Error("expected room ID to be set")
	}
	if resp.Msg.PlayerId == "" {
		t.Error("expected player ID to be set")
	}
	if resp.Msg.SessionToken == "" {
		t.Error("expected session token to be set")
	}
	if len(resp.Msg.Room.Players) != 1 {
		t.Errorf("expected 1 player, got %d", len(resp.Msg.Room.Players))
	}
	if resp.Msg.Room.Players[0].Name != "Alice" {
		t.Errorf("expected player name 'Alice', got '%s'", resp.Msg.Room.Players[0].Name)
	}
}

func TestBounceBotServer_CreateRoom_SinglePlayer(t *testing.T) {
	svc := room.NewRoomService()
	server := NewBounceBotServer(svc)

	req := connect.NewRequest(&pb.CreateRoomRequest{
		PlayerName:     "Solo",
		IsSinglePlayer: true,
	})

	resp, err := server.CreateRoom(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.Msg.Room.IsSinglePlayer {
		t.Error("expected single player room")
	}
}

func TestBounceBotServer_JoinRoom(t *testing.T) {
	svc := room.NewRoomService()
	server := NewBounceBotServer(svc)

	// Create a room first
	r, _, _ := svc.Create("Alice", false)

	req := connect.NewRequest(&pb.JoinRoomRequest{
		RoomId:     r.ID,
		PlayerName: "Bob",
	})

	resp, err := server.JoinRoom(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Msg.PlayerId == "" {
		t.Error("expected player ID to be set")
	}
	if resp.Msg.SessionToken == "" {
		t.Error("expected session token to be set")
	}
	if len(resp.Msg.Room.Players) != 2 {
		t.Errorf("expected 2 players, got %d", len(resp.Msg.Room.Players))
	}
}

func TestBounceBotServer_JoinRoom_NotFound(t *testing.T) {
	svc := room.NewRoomService()
	server := NewBounceBotServer(svc)

	req := connect.NewRequest(&pb.JoinRoomRequest{
		RoomId:     "NONEXISTENT",
		PlayerName: "Bob",
	})

	_, err := server.JoinRoom(context.Background(), req)
	if err == nil {
		t.Error("expected error for nonexistent room")
	}
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("expected NotFound code, got %v", connect.CodeOf(err))
	}
}

func TestBounceBotServer_GetRoom(t *testing.T) {
	svc := room.NewRoomService()
	server := NewBounceBotServer(svc)

	r, _, _ := svc.Create("Alice", false)

	req := connect.NewRequest(&pb.GetRoomRequest{
		RoomId: r.ID,
	})

	resp, err := server.GetRoom(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Msg.Id != r.ID {
		t.Errorf("expected room ID '%s', got '%s'", r.ID, resp.Msg.Id)
	}
}

func TestBounceBotServer_GetRoom_NotFound(t *testing.T) {
	svc := room.NewRoomService()
	server := NewBounceBotServer(svc)

	req := connect.NewRequest(&pb.GetRoomRequest{
		RoomId: "NONEXISTENT",
	})

	_, err := server.GetRoom(context.Background(), req)
	if err == nil {
		t.Error("expected error for nonexistent room")
	}
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("expected NotFound code, got %v", connect.CodeOf(err))
	}
}

func TestBounceBotServer_StartGame(t *testing.T) {
	svc := room.NewRoomService()
	server := NewBounceBotServer(svc)

	r, _, _ := svc.Create("Alice", false)

	req := connect.NewRequest(&pb.StartGameRequest{
		RoomId: r.ID,
	})

	resp, err := server.StartGame(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Msg.CurrentGame == nil {
		t.Error("expected current game to be set")
	}
	if resp.Msg.GameStartedAt == nil {
		t.Error("expected game started time to be set")
	}
}

func TestBounceBotServer_StartGame_NotFound(t *testing.T) {
	svc := room.NewRoomService()
	server := NewBounceBotServer(svc)

	req := connect.NewRequest(&pb.StartGameRequest{
		RoomId: "NONEXISTENT",
	})

	_, err := server.StartGame(context.Background(), req)
	if err == nil {
		t.Error("expected error for nonexistent room")
	}
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("expected NotFound code, got %v", connect.CodeOf(err))
	}
}

func TestBounceBotServer_SubmitSolution(t *testing.T) {
	svc := room.NewRoomService()
	server := NewBounceBotServer(svc)

	r, _, sessionToken := svc.Create("Alice", false)
	svc.StartGame(r.ID)

	// Use Game1 for a known valid solution
	r, _ = svc.Get(r.ID)
	r.CurrentGame = model.Game1()

	validMoves := model.Game1OptimalSolution()
	protoMoves := make([]*pb.BotPos, len(validMoves))
	for i, m := range validMoves {
		protoMoves[i] = m.ToProto()
	}

	req := connect.NewRequest(&pb.SubmitSolutionRequest{
		RoomId:       r.ID,
		SessionToken: sessionToken,
		Moves:        protoMoves,
	})

	resp, err := server.SubmitSolution(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Msg.Solution.PlayerId != r.Players[0].ID {
		t.Errorf("expected player ID '%s', got '%s'", r.Players[0].ID, resp.Msg.Solution.PlayerId)
	}
	if len(resp.Msg.Solution.Moves) != len(validMoves) {
		t.Errorf("expected %d moves, got %d", len(validMoves), len(resp.Msg.Solution.Moves))
	}
}

func TestBounceBotServer_SubmitSolution_Invalid(t *testing.T) {
	svc := room.NewRoomService()
	server := NewBounceBotServer(svc)

	r, _, sessionToken := svc.Create("Alice", false)
	svc.StartGame(r.ID)
	r, _ = svc.Get(r.ID)
	r.CurrentGame = model.Game1()

	// Submit an invalid solution (wrong position)
	invalidMoves := []*pb.BotPos{
		{Id: 0, Pos: &pb.Position{X: 5, Y: 6}},
	}

	req := connect.NewRequest(&pb.SubmitSolutionRequest{
		RoomId:       r.ID,
		SessionToken: sessionToken,
		Moves:        invalidMoves,
	})

	_, err := server.SubmitSolution(context.Background(), req)
	if err == nil {
		t.Error("expected error for invalid solution")
	}
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("expected InvalidArgument code, got %v", connect.CodeOf(err))
	}
}

func TestBounceBotServer_MarkFinishedSolving(t *testing.T) {
	svc := room.NewRoomService()
	server := NewBounceBotServer(svc)

	r, _, sessionToken := svc.Create("Alice", false)
	svc.StartGame(r.ID)

	req := connect.NewRequest(&pb.MarkFinishedSolvingRequest{
		RoomId:       r.ID,
		SessionToken: sessionToken,
	})

	resp, err := server.MarkFinishedSolving(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.Msg.Success {
		t.Error("expected success to be true")
	}
}

func TestBounceBotServer_MarkFinishedSolving_NotFound(t *testing.T) {
	svc := room.NewRoomService()
	server := NewBounceBotServer(svc)

	req := connect.NewRequest(&pb.MarkFinishedSolvingRequest{
		RoomId:       "NONEXISTENT",
		SessionToken: "invalid-token",
	})

	_, err := server.MarkFinishedSolving(context.Background(), req)
	if err == nil {
		t.Error("expected error for nonexistent room")
	}
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("expected NotFound code, got %v", connect.CodeOf(err))
	}
}

func TestBounceBotServer_MarkReadyForNext(t *testing.T) {
	svc := room.NewRoomService()
	server := NewBounceBotServer(svc)

	r, _, sessionToken := svc.Create("Alice", false)

	req := connect.NewRequest(&pb.MarkReadyForNextRequest{
		RoomId:       r.ID,
		SessionToken: sessionToken,
	})

	resp, err := server.MarkReadyForNext(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.Msg.Success {
		t.Error("expected success to be true")
	}
}

func TestBounceBotServer_MarkReadyForNext_NotFound(t *testing.T) {
	svc := room.NewRoomService()
	server := NewBounceBotServer(svc)

	req := connect.NewRequest(&pb.MarkReadyForNextRequest{
		RoomId:       "NONEXISTENT",
		SessionToken: "invalid-token",
	})

	_, err := server.MarkReadyForNext(context.Background(), req)
	if err == nil {
		t.Error("expected error for nonexistent room")
	}
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("expected NotFound code, got %v", connect.CodeOf(err))
	}
}

func TestBounceBotServer_UpdateRoomSettings(t *testing.T) {
	svc := room.NewRoomService()
	server := NewBounceBotServer(svc)

	r, _, sessionToken := svc.Create("Alice", false)

	req := connect.NewRequest(&pb.UpdateRoomSettingsRequest{
		RoomId:       r.ID,
		SessionToken: sessionToken,
		Settings: &pb.RoomSettings{
			ShowSolverMoveCount: true,
			ShowSolverSolutions: true,
		},
	})

	resp, err := server.UpdateRoomSettings(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.Msg.Success {
		t.Error("expected success to be true")
	}

	// Verify settings were applied
	r, _ = svc.Get(r.ID)
	if !r.Settings.ShowSolverMoveCount {
		t.Error("expected ShowSolverMoveCount to be true")
	}
	if !r.Settings.ShowSolverSolutions {
		t.Error("expected ShowSolverSolutions to be true")
	}
}

func TestBounceBotServer_UpdateRoomSettings_NotHost(t *testing.T) {
	svc := room.NewRoomService()
	server := NewBounceBotServer(svc)

	r, _, _ := svc.Create("Alice", false)
	_, _, bobSessionToken, _ := svc.Join(r.ID, "Bob")

	// Bob (not host) tries to change settings
	req := connect.NewRequest(&pb.UpdateRoomSettingsRequest{
		RoomId:       r.ID,
		SessionToken: bobSessionToken,
		Settings: &pb.RoomSettings{
			ShowSolverMoveCount: true,
		},
	})

	_, err := server.UpdateRoomSettings(context.Background(), req)
	if err == nil {
		t.Error("expected error when non-host updates settings")
	}
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Errorf("expected PermissionDenied code, got %v", connect.CodeOf(err))
	}
}

func TestBounceBotServer_UpdateRoomSettings_NotFound(t *testing.T) {
	svc := room.NewRoomService()
	server := NewBounceBotServer(svc)

	req := connect.NewRequest(&pb.UpdateRoomSettingsRequest{
		RoomId:       "NONEXISTENT",
		SessionToken: "invalid-token",
		Settings:     &pb.RoomSettings{},
	})

	_, err := server.UpdateRoomSettings(context.Background(), req)
	if err == nil {
		t.Error("expected error for nonexistent room")
	}
}
