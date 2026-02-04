package grpc

import (
	"context"
	"fmt"

	usersvc "github.com/codex-k8s/project-example/services/internal/users/internal/domain/service"
	"github.com/codex-k8s/project-example/services/internal/users/internal/domain/types/entity"
	usersgen "github.com/codex-k8s/project-example/services/internal/users/internal/transport/grpc/generated/users/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server implements the generated UsersService gRPC server.
type Server struct {
	usersgen.UnimplementedUsersServiceServer
	svc *usersvc.Service
}

// Register registers the gRPC server implementation.
func Register(s *grpc.Server, svc *usersvc.Service) {
	usersgen.RegisterUsersServiceServer(s, &Server{svc: svc})
}

// Register handles UsersService.Register.
func (s *Server) Register(ctx context.Context, req *usersgen.RegisterRequest) (*usersgen.RegisterResponse, error) {
	u, err := s.svc.Register(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, err
	}
	return &usersgen.RegisterResponse{User: toProtoUser(u)}, nil
}

// Authenticate handles UsersService.Authenticate.
func (s *Server) Authenticate(ctx context.Context, req *usersgen.AuthenticateRequest) (*usersgen.AuthenticateResponse, error) {
	u, err := s.svc.Authenticate(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, err
	}
	return &usersgen.AuthenticateResponse{User: toProtoUser(u)}, nil
}

// GetUser handles UsersService.GetUser.
func (s *Server) GetUser(ctx context.Context, req *usersgen.GetUserRequest) (*usersgen.GetUserResponse, error) {
	u, err := s.svc.GetUser(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &usersgen.GetUserResponse{User: toProtoUser(u)}, nil
}

func toProtoUser(u entity.User) *usersgen.User {
	ts := timestamppb.New(u.CreatedAt)
	if err := ts.CheckValid(); err != nil {
		// This is an invariant violation; the boundary interceptor will map it to Internal.
		panic(fmt.Errorf("invalid created_at timestamp: %w", err))
	}
	return &usersgen.User{
		Id:        u.ID,
		Username:  u.Username,
		CreatedAt: ts,
	}
}
