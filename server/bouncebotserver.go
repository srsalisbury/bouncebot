package main

import (
	"context"

	"connectrpc.com/connect"
	"github.com/srsalisbury/bouncebot/model"
	pb "github.com/srsalisbury/bouncebot/proto"
	"github.com/srsalisbury/bouncebot/server/room"
)

type bounceBotServer struct {
	rooms *room.RoomService
}

func NewBounceBotServer(rooms *room.RoomService) *bounceBotServer {
	return &bounceBotServer{rooms: rooms}
}

func (s *bounceBotServer) MakeGame(_ context.Context, req *connect.Request[pb.MakeGameRequest]) (*connect.Response[pb.Game], error) {
	game := model.Game1()
	return connect.NewResponse(game.ToProto()), nil
}

func (s *bounceBotServer) CreateRoom(_ context.Context, req *connect.Request[pb.CreateRoomRequest]) (*connect.Response[pb.Room], error) {
	r := s.rooms.Create(req.Msg.PlayerName)
	return connect.NewResponse(r.ToProto()), nil
}

func (s *bounceBotServer) JoinRoom(_ context.Context, req *connect.Request[pb.JoinRoomRequest]) (*connect.Response[pb.Room], error) {
	r, err := s.rooms.Join(req.Msg.RoomId, req.Msg.PlayerName)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(r.ToProto()), nil
}

func (s *bounceBotServer) GetRoom(_ context.Context, req *connect.Request[pb.GetRoomRequest]) (*connect.Response[pb.Room], error) {
	r, err := s.rooms.Get(req.Msg.RoomId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(r.ToProto()), nil
}

func (s *bounceBotServer) StartGame(_ context.Context, req *connect.Request[pb.StartGameRequest]) (*connect.Response[pb.Room], error) {
	r, err := s.rooms.StartGame(req.Msg.RoomId, req.Msg.UseFixedBoard)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(r.ToProto()), nil
}

func (s *bounceBotServer) SubmitSolution(_ context.Context, req *connect.Request[pb.SubmitSolutionRequest]) (*connect.Response[pb.SubmitSolutionResponse], error) {
	moves := model.NewBotPositionsFromProto(req.Msg.Moves)
	solution, err := s.rooms.SubmitSolution(req.Msg.RoomId, req.Msg.PlayerId, moves)
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

func (s *bounceBotServer) RetractSolution(_ context.Context, req *connect.Request[pb.RetractSolutionRequest]) (*connect.Response[pb.RetractSolutionResponse], error) {
	err := s.rooms.RetractSolution(req.Msg.RoomId, req.Msg.PlayerId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&pb.RetractSolutionResponse{
		Success: true,
	}), nil
}

func (s *bounceBotServer) MarkFinishedSolving(_ context.Context, req *connect.Request[pb.MarkFinishedSolvingRequest]) (*connect.Response[pb.MarkFinishedSolvingResponse], error) {
	err := s.rooms.MarkFinishedSolving(req.Msg.RoomId, req.Msg.PlayerId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&pb.MarkFinishedSolvingResponse{
		Success: true,
	}), nil
}

func (s *bounceBotServer) MarkReadyForNext(_ context.Context, req *connect.Request[pb.MarkReadyForNextRequest]) (*connect.Response[pb.MarkReadyForNextResponse], error) {
	err := s.rooms.MarkReadyForNext(req.Msg.RoomId, req.Msg.PlayerId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&pb.MarkReadyForNextResponse{
		Success: true,
	}), nil
}
