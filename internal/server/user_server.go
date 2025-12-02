package server

import (
	"context"
	"regexp"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"service-1-user/internal/auth"
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
func (s *userServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if req.Id <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "User ID must be positive, got %d", req.Id)
	}

	user, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, status.Errorf(codes.NotFound, "User with ID %d not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	return &pb.GetUserResponse{User: user}, nil
}

// CreateUser implement RPC CreateUser
func (s *userServiceServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
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

	// 3. Hash password if provided
	var user *pb.User
	var err error

	if req.Password != "" {
		// Hash the password
		passwordHash, err := auth.HashPassword(req.Password)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to hash password: %v", err)
		}

		// Create user with password
		user, err = s.repo.CreateWithPassword(ctx, req.Name, req.Email, passwordHash)
		if err != nil {
			if containsString(err.Error(), "duplicate") || containsString(err.Error(), "unique") {
				return nil, status.Errorf(codes.AlreadyExists, "Email %s is already registered", req.Email)
			}
			return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
		}
	} else {
		// Create user without password (legacy)
		user, err = s.repo.Create(ctx, req.Name, req.Email)
		if err != nil {
			if containsString(err.Error(), "duplicate") || containsString(err.Error(), "unique") {
				return nil, status.Errorf(codes.AlreadyExists, "Email %s is already registered", req.Email)
			}
			return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
		}
	}

	return &pb.CreateUserResponse{User: user}, nil
}

// Helper function to check valid email format (simple version)
var emailRegex = regexp.MustCompile("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$")

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
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
func (s *userServiceServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
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
	return &pb.UpdateUserResponse{User: user}, nil
}

// DeleteUser implement RPC DeleteUser
func (s *userServiceServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	// 1. Validate input
	if req.Id <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "User id must be positive")
	}
	// 2. Delete user
	err := s.repo.Delete(ctx, req.Id)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, status.Errorf(codes.NotFound, "User with ID %d not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}
	return &pb.DeleteUserResponse{Success: true}, nil
}

// ListUsers implement RPC ListUsers
func (s *userServiceServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	// 1. Validate pagination params
	pageSize := req.PageSize
	pageNumber := req.Page

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
		User:  users,
		Total: total,
	}, nil
}

// Login implement RPC Login
func (s *userServiceServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// 1. Validate input
	if req.Email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Email is required")
	}
	if req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Password is required")
	}

	// 2. Get user by email with password hash
	userWithPassword, err := s.repo.GetByEmailWithPassword(ctx, req.Email)
	if err != nil {
		if containsString(err.Error(), "no rows") {
			return nil, status.Errorf(codes.NotFound, "Invalid email or password")
		}
		return nil, status.Errorf(codes.Internal, "Failed to get user: %v", err)
	}

	// 3. Check password
	if !auth.CheckPassword(req.Password, userWithPassword.PasswordHash) {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid email or password")
	}

	// 4. Generate tokens
	accessToken, err := auth.GenerateToken(userWithPassword.User.Id, userWithPassword.User.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to generate access token: %v", err)
	}

	refreshToken, err := auth.GenerateRefreshToken(userWithPassword.User.Id, userWithPassword.User.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to generate refresh token: %v", err)
	}

	// 5. Return response
	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userWithPassword.User,
		Message:      "Login successful",
	}, nil
}

// ValidateToken implement RPC ValidateToken
func (s *userServiceServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	// 1. Validate input
	if req.Token == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Token is required")
	}

	// 2. Validate token
	claims, err := auth.ValidateToken(req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Valid:   false,
			Message: "Invalid or expired token",
		}, nil
	}

	// 3. Return response
	return &pb.ValidateTokenResponse{
		Valid:   true,
		UserId:  claims.UserID,
		Email:   claims.Email,
		Message: "Token is valid",
	}, nil
}

// Logout implement RPC Logout
func (s *userServiceServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	// For stateless JWT, logout is handled client-side by removing the token
	// In production, you might want to implement token blacklist using Redis

	return &pb.LogoutResponse{
		Success: true,
		Message: "Logout successful. Please remove token from client.",
	}, nil
}
