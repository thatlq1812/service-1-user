package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"service-1-user/internal/repository"
	pb "service-1-user/proto"
)

// userServiceServer implements UserServiceServer interface
type userServiceServer struct {
	pb.UnimplementedUserServiceServer
	repo repository.UserRepository
}

// NewUserServiceServer create server
func NewUserServiceServer(repo repository.UserRepository) pb.UserServiceServer {
	return &userServiceServer{
		repo: repo,
	}
}

// GetUser implement RPC GetUser
func (s *userServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	// 1. Validate input
	if req.Id <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "User ID must be positive, got %d", req.Id)
	}

	// 2. Call repository
	user, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		// Check if not found error
		if err.Error() == "no rows in result set" {
			return nil, status.Errorf(codes.NotFound, "User with ID %d not found", req.Id)
		}

		// Other errors are internal
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	// 3. Return success
	return user, nil
}

// CreateUser implement RPC CreateUser
func (s *userServiceServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	//1. Validate input
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Name is required")
	}
	if req.Email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Email is required")
	}

	// 2. Basic email validation
	if !isValidEmail(req.Email) {
		return nil, status.Errorf(codes.InvalidArgument, "Email format is invalid: %s", req.Email)
	}

	// 3. Call repository
	user, err := s.repo.Create(ctx, req.Name, req.Email)
	if err != nil {
		//Check if duplicate email (unique constraint violation)
		if containsString(err.Error(), "duplicate") || containsString(err.Error(), "unique") {
			return nil, status.Errorf(codes.AlreadyExists, "Email %s is already registered", req.Email)
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}
	return user, nil
}

// Helper function to check valid email format (simple version)
func isValidEmail(email string) bool {
	// Sumple check: constain @ and .
	hasAt := false
	hasDot := false
	for _, c := range email {
		if c == '@' {
			hasAt = true
		}
		if c == '.' {
			hasDot = true
		}
	}
	return hasAt && hasDot && len(email) >= 5
}

// Helper function to check if a substring is in a string
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && hasSubstring(s, substr))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// UpdateUser implement RPC UpdateUser
func (s *userServiceServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	// 1. Validate input
	if req.Id <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "User id must be positive")
	}
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Name is required")
	}
	if req.Email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Email is required")
	}

	// 2. Call repository
	user, err := s.repo.Update(ctx, req.Id, req.Name, req.Email)
	if err != nil {
		// Check if not found
		if err.Error() == "no rows in result set" {
			return nil, status.Errorf(codes.NotFound, "User with ID %d not found", req.Id)
		}
		//Check if duplicate email
		if containsString(err.Error(), "duplicate") || containsString(err.Error(), "unique") {
			return nil, status.Errorf(codes.AlreadyExists, "Email %s is already registered", req.Email)
		}
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}
	return user, nil
}

// DeleteUser implement RPC DeleteUser
func (s *userServiceServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.User, error) {
	// 1. Validate input
	if req.Id <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "User id must be positive")
	}
	// 2. Get user first (to return in response)
	user, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, status.Errorf(codes.NotFound, "User with ID %d not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}
	// 3. Detele user
	err = s.repo.Delete(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}
	return user, nil
}

// ListUsers implement RPC ListUsers
func (s *userServiceServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	// 1. Validate pagination params
	pageSize := req.PageSize
	pageNumber := req.PageNumber

	// Calculate limit and offset from page_size and page_number
	if pageSize <= 0 {
		pageSize = 10 // Default page size
	}
	if pageSize > 100 {
		return nil, status.Errorf(codes.InvalidArgument, "page_size too large, max 100, got %d", pageSize)
	}
	if pageNumber < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "page_number must be non-negative, got %d", pageNumber)
	}

	// Convert page_number to offset
	// page_number 0 = first page (offset 0)
	// page_number 1 = second page (offset = page_size)
	offset := pageNumber * pageSize

	// 2. Call repository
	users, total, err := s.repo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	// 3. Build response
	return &pb.ListUsersResponse{
		Users: users,
		Total: total,
	}, nil
}
