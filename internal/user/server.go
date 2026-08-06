package user

import (
	"context"
	userv1 "ecommerce/gen/user/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	userv1.UnimplementedUserServiceServer
	store *Store
}

func NewServer(store *Store) *Server {
	return &Server{store: store}
}

func (s *Server) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	u, err := s.store.Create(req.GetEmail(), req.GetFullName(), req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.AlreadyExists, "create user: %v", err)
	}
	return &userv1.CreateUserResponse{User: toProto(u)}, nil
}

func (s *Server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	u, ok := s.store.Get(req.GetUserId())
	if !ok {
		return nil, status.Errorf(codes.NotFound, "user %q not found", req.GetUserId())
	}
	return &userv1.GetUserResponse{User: toProto(u)}, nil
}

func (s *Server) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	users := s.store.List()
	out := make([]*userv1.User, 0, len(users))
	for _, u := range users {
		out = append(out, toProto(u))
	}
	return &userv1.ListUsersResponse{Users: out}, nil
}

func toProto(u User) *userv1.User {
	return &userv1.User{
		UserId:    u.ID,
		Email:     u.Email,
		FullName:  u.FullName,
		CreatedAt: timestamppb.New(u.CreatedAt),
	}
}
