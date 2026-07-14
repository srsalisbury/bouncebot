package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/srsalisbury/bouncebot/model"
	pb "github.com/srsalisbury/bouncebot/proto"
	"github.com/srsalisbury/bouncebot/server/daily"
	"github.com/srsalisbury/bouncebot/server/ratelimit"
	"github.com/srsalisbury/bouncebot/server/room"
)

var validUUID = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func validateDailyPlayerID(playerID string) error {
	if !validUUID.MatchString(playerID) {
		return fmt.Errorf("invalid player ID format")
	}
	return nil
}

func validateDailyDate(date string) error {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid date format")
	}
	today := time.Now().UTC()
	maxDate := today.AddDate(0, 0, 1)
	if parsed.After(maxDate) {
		return fmt.Errorf("cannot submit solutions for future dates")
	}
	return nil
}

const maxPlayerNameLength = 30

func validatePlayerName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("player name cannot be empty")
	}
	if len(name) > maxPlayerNameLength {
		return "", fmt.Errorf("player name too long (max %d characters)", maxPlayerNameLength)
	}
	return name, nil
}

type bounceBotServer struct {
	rooms              *room.RoomService
	getRoomLimiter     *ratelimit.Limiter
	createRoomLimiter  *ratelimit.Limiter
	submitDailyLimiter *ratelimit.Limiter
	dailyMgr           *daily.Manager
	dailyProgressMgr   *daily.ProgressManager
}

func NewBounceBotServer(rooms *room.RoomService, getRoomLimiter, createRoomLimiter, submitDailyLimiter *ratelimit.Limiter, dailyMgr *daily.Manager, dailyProgressMgr *daily.ProgressManager) *bounceBotServer {
	return &bounceBotServer{
		rooms:              rooms,
		getRoomLimiter:     getRoomLimiter,
		createRoomLimiter:  createRoomLimiter,
		submitDailyLimiter: submitDailyLimiter,
		dailyMgr:           dailyMgr,
		dailyProgressMgr:   dailyProgressMgr,
	}
}

func (s *bounceBotServer) CreateRoom(ctx context.Context, req *connect.Request[pb.CreateRoomRequest]) (*connect.Response[pb.CreateRoomResponse], error) {
	clientIP := ratelimit.ClientIPFromContext(ctx)
	if clientIP != "" && s.createRoomLimiter != nil && !s.createRoomLimiter.Allow(clientIP) {
		return nil, connect.NewError(connect.CodeResourceExhausted, nil)
	}

	playerName, err := validatePlayerName(req.Msg.PlayerName)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	r, playerID, sessionToken := s.rooms.Create(playerName, req.Msg.IsSinglePlayer)
	return connect.NewResponse(&pb.CreateRoomResponse{
		Room:         r.ToProto(),
		PlayerId:     playerID,
		SessionToken: sessionToken,
	}), nil
}

func (s *bounceBotServer) JoinRoom(_ context.Context, req *connect.Request[pb.JoinRoomRequest]) (*connect.Response[pb.JoinRoomResponse], error) {
	playerName, err := validatePlayerName(req.Msg.PlayerName)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	roomProto, playerID, sessionToken, err := s.rooms.Join(req.Msg.RoomId, playerName)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&pb.JoinRoomResponse{
		Room:         roomProto,
		PlayerId:     playerID,
		SessionToken: sessionToken,
	}), nil
}

func (s *bounceBotServer) GetRoom(ctx context.Context, req *connect.Request[pb.GetRoomRequest]) (*connect.Response[pb.Room], error) {
	// Apply rate limiting per IP (Option B from PLAYER_AUTHENTICATION_DESIGN.md)
	clientIP := ratelimit.ClientIPFromContext(ctx)
	if clientIP != "" && s.getRoomLimiter != nil && !s.getRoomLimiter.Allow(clientIP) {
		return nil, connect.NewError(connect.CodeResourceExhausted, nil)
	}

	roomProto, err := s.rooms.GetProto(req.Msg.RoomId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(roomProto), nil
}

func (s *bounceBotServer) StartGame(_ context.Context, req *connect.Request[pb.StartGameRequest]) (*connect.Response[pb.Room], error) {
	roomProto, err := s.rooms.StartGame(req.Msg.RoomId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(roomProto), nil
}

func (s *bounceBotServer) SubmitSolution(_ context.Context, req *connect.Request[pb.SubmitSolutionRequest]) (*connect.Response[pb.SubmitSolutionResponse], error) {
	moves := model.NewBotPositionsFromProto(req.Msg.Moves)
	solution, err := s.rooms.SubmitSolution(req.Msg.RoomId, req.Msg.SessionToken, moves)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Convert moves back to proto for response
	protoMoves := make([]*pb.BotPos, len(solution.Moves))
	for i, move := range solution.Moves {
		protoMoves[i] = move.ToProto()
	}

	return connect.NewResponse(&pb.SubmitSolutionResponse{
		Solution: &pb.PlayerSolution{
			PlayerId: solution.PlayerID,
			Moves:    protoMoves,
		},
	}), nil
}

func (s *bounceBotServer) MarkFinishedSolving(_ context.Context, req *connect.Request[pb.MarkFinishedSolvingRequest]) (*connect.Response[pb.MarkFinishedSolvingResponse], error) {
	err := s.rooms.MarkFinishedSolving(req.Msg.RoomId, req.Msg.SessionToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&pb.MarkFinishedSolvingResponse{
		Success: true,
	}), nil
}

func (s *bounceBotServer) MarkReadyForNext(_ context.Context, req *connect.Request[pb.MarkReadyForNextRequest]) (*connect.Response[pb.MarkReadyForNextResponse], error) {
	err := s.rooms.MarkReadyForNext(req.Msg.RoomId, req.Msg.SessionToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&pb.MarkReadyForNextResponse{
		Success: true,
	}), nil
}

func (s *bounceBotServer) UpdateRoomSettings(_ context.Context, req *connect.Request[pb.UpdateRoomSettingsRequest]) (*connect.Response[pb.UpdateRoomSettingsResponse], error) {
	settings := room.RoomSettings{
		ShowSolverMoveCount: req.Msg.Settings.ShowSolverMoveCount,
		ShowSolverSolutions: req.Msg.Settings.ShowSolverSolutions,
		MinSolutionLength:   int(req.Msg.Settings.MinSolutionLength),
	}
	err := s.rooms.UpdateRoomSettings(req.Msg.RoomId, req.Msg.SessionToken, settings)
	if err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}
	return connect.NewResponse(&pb.UpdateRoomSettingsResponse{
		Success: true,
	}), nil
}

func (s *bounceBotServer) LeaveRoom(_ context.Context, req *connect.Request[pb.LeaveRoomRequest]) (*connect.Response[pb.LeaveRoomResponse], error) {
	err := s.rooms.LeaveRoom(req.Msg.RoomId, req.Msg.SessionToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&pb.LeaveRoomResponse{
		Success: true,
	}), nil
}

func (s *bounceBotServer) BootPlayer(_ context.Context, req *connect.Request[pb.BootPlayerRequest]) (*connect.Response[pb.BootPlayerResponse], error) {
	err := s.rooms.BootPlayer(req.Msg.RoomId, req.Msg.SessionToken, req.Msg.TargetPlayerId)
	if err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}
	return connect.NewResponse(&pb.BootPlayerResponse{
		Success: true,
	}), nil
}

func (s *bounceBotServer) GetServerInfo(_ context.Context, _ *connect.Request[pb.GetServerInfoRequest]) (*connect.Response[pb.GetServerInfoResponse], error) {
	return connect.NewResponse(&pb.GetServerInfoResponse{
		DailyChallengeEnabled: s.dailyMgr != nil,
	}), nil
}

func (s *bounceBotServer) GetDailyChallenge(_ context.Context, req *connect.Request[pb.GetDailyChallengeRequest]) (*connect.Response[pb.GetDailyChallengeResponse], error) {
	if s.dailyMgr == nil {
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("daily challenges are not enabled"))
	}

	if err := validateDailyPlayerID(req.Msg.PlayerId); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Calculate user's local date based on timezone offset
	now := time.Now().UTC()
	offset := time.Duration(req.Msg.TimezoneOffsetMinutes) * time.Minute
	localTime := now.Add(-offset) // Subtract offset since JS gives minutes behind UTC
	date := localTime.Format("2006-01-02")

	// Get puzzles for the date
	puzzles, err := s.dailyMgr.GetPuzzlesForDate(date)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if puzzles == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("no puzzles for date %s", date))
	}

	// Get user progress
	progress, err := s.dailyProgressMgr.GetUserProgress(req.Msg.PlayerId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	dayProgress := progress[date]

	// Build response puzzles
	puzzleInfos := make([]*pb.DailyPuzzleInfo, 0, 3)

	// Helper to build puzzle info
	buildPuzzleInfo := func(difficulty string, puzzle *daily.DailyPuzzle, solved bool) *pb.DailyPuzzleInfo {
		if puzzle == nil {
			return nil
		}
		gameStr := strings.Join(puzzle.Game, "\n")
		game, err := model.ParseGameString(gameStr)
		if err != nil {
			return nil
		}
		info := &pb.DailyPuzzleInfo{
			Difficulty: difficulty,
			Game:       game.ToProto(),
			Solved:     solved,
		}
		if solved {
			info.OptimalMoves = int32(puzzle.OptimalMoves)
		}
		return info
	}

	if info := buildPuzzleInfo(daily.DifficultyEasy, puzzles.Easy, dayProgress.Easy); info != nil {
		puzzleInfos = append(puzzleInfos, info)
	}
	if info := buildPuzzleInfo(daily.DifficultyMedium, puzzles.Medium, dayProgress.Medium); info != nil {
		puzzleInfos = append(puzzleInfos, info)
	}
	if info := buildPuzzleInfo(daily.DifficultyHard, puzzles.Hard, dayProgress.Hard); info != nil {
		puzzleInfos = append(puzzleInfos, info)
	}

	// Calculate seconds until next local midnight
	localMidnight := time.Date(localTime.Year(), localTime.Month(), localTime.Day()+1, 0, 0, 0, 0, time.UTC)
	secondsUntilReset := int32(localMidnight.Sub(localTime).Seconds())

	return connect.NewResponse(&pb.GetDailyChallengeResponse{
		Date:              date,
		Puzzles:           puzzleInfos,
		SecondsUntilReset: secondsUntilReset,
	}), nil
}

func (s *bounceBotServer) SubmitDailySolution(ctx context.Context, req *connect.Request[pb.SubmitDailySolutionRequest]) (*connect.Response[pb.SubmitDailySolutionResponse], error) {
	if s.dailyMgr == nil {
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("daily challenges are not enabled"))
	}

	clientIP := ratelimit.ClientIPFromContext(ctx)
	if clientIP != "" && s.submitDailyLimiter != nil && !s.submitDailyLimiter.Allow(clientIP) {
		return nil, connect.NewError(connect.CodeResourceExhausted, nil)
	}

	if err := validateDailyPlayerID(req.Msg.PlayerId); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Validate the submitted date
	if err := validateDailyDate(req.Msg.Date); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Get puzzles for the date
	puzzles, err := s.dailyMgr.GetPuzzlesForDate(req.Msg.Date)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if puzzles == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("no puzzles for date %s", req.Msg.Date))
	}

	// Get the puzzle for the requested difficulty
	var puzzle *daily.DailyPuzzle
	switch req.Msg.Difficulty {
	case daily.DifficultyEasy:
		puzzle = puzzles.Easy
	case daily.DifficultyMedium:
		puzzle = puzzles.Medium
	case daily.DifficultyHard:
		puzzle = puzzles.Hard
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid difficulty: %s", req.Msg.Difficulty))
	}
	if puzzle == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("puzzle not found for difficulty %s", req.Msg.Difficulty))
	}

	// Parse the game
	gameStr := strings.Join(puzzle.Game, "\n")
	game, err := model.ParseGameString(gameStr)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Validate the solution by simulating the moves
	moves := model.NewBotPositionsFromProto(req.Msg.Moves)
	if valid, _ := game.CheckSolution(moves); !valid {
		return connect.NewResponse(&pb.SubmitDailySolutionResponse{
			Correct:       false,
			NewCompletion: false,
		}), nil
	}

	// Check if already solved
	alreadySolved, err := s.dailyProgressMgr.IsSolved(req.Msg.PlayerId, req.Msg.Date, req.Msg.Difficulty)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Mark as solved
	if !alreadySolved {
		if err := s.dailyProgressMgr.MarkSolved(req.Msg.PlayerId, req.Msg.Date, req.Msg.Difficulty); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	return connect.NewResponse(&pb.SubmitDailySolutionResponse{
		Correct:       true,
		NewCompletion: !alreadySolved,
	}), nil
}

