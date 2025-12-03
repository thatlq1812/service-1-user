package server

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"agrios/pkg/common"
	"service-1-user/internal/auth"
	"service-1-user/internal/repository"
	pb "service-1-user/proto"
)

const (
	defaultPageSize    = 10
	maxPageSize        = 100
	errNoRows          = "no rows in result set"
	errDuplicateKey    = "duplicate"
	errUniqueViolation = "unique"
)

// userServiceServer implements UserServiceServer interface
type userServiceServer struct {
	pb.UnimplementedUserServiceServer
	repo         repository.UserRepository
	tokenManager *auth.TokenManager
}

// NewUserServiceServer create server
func NewUserServiceServer(repo repository.UserRepository, tokenManager *auth.TokenManager) pb.UserServiceServer {
	return &userServiceServer{
		repo:         repo,
		tokenManager: tokenManager,
	}
}

// GetUser retrieves a user by ID
func (s *userServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if req.Id <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "user ID must be positive, got %d", req.Id)
	}

	user, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "user with ID %d not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	return &pb.GetUserResponse{User: user}, nil
}

// CreateUser creates a new user with optional password
func (s *userServiceServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// 1. Validate input
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	// 2. Validate email format
	if !common.IsValidEmail(req.Email) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid email format: %s", req.Email)
	}

	// 3. Hash password if provided and create user
	var user *pb.User
	var err error

	if req.Password != "" {
		// Hash the password
		passwordHash, err := auth.HashPassword(req.Password)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
		}

		// Create user with password
		user, err = s.repo.CreateWithPassword(ctx, req.Name, req.Email, passwordHash)
		if err != nil {
			if isDuplicateError(err) {
				return nil, status.Errorf(codes.AlreadyExists, "email %s is already registered", req.Email)
			}
			return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
		}
	} else {
		// Create user without password (legacy support)
		user, err = s.repo.Create(ctx, req.Name, req.Email)
		if err != nil {
			if isDuplicateError(err) {
				return nil, status.Errorf(codes.AlreadyExists, "email %s is already registered", req.Email)
			}
			return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
		}
	}

	return &pb.CreateUserResponse{User: user}, nil
}

// isDuplicateError checks if error is related to duplicate key constraint violation
func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, errDuplicateKey) || strings.Contains(errMsg, errUniqueViolation)
}

// UpdateUser updates user information
func (s *userServiceServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	// 1. Validate input
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user ID must be positive")
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	// 2. Validate email format
	if !common.IsValidEmail(req.Email) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid email format: %s", req.Email)
	}

	// 3. Update user
	user, err := s.repo.Update(ctx, req.Id, req.Name, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "user with ID %d not found", req.Id)
		}
		if isDuplicateError(err) {
			return nil, status.Errorf(codes.AlreadyExists, "email %s is already registered", req.Email)
		}
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &pb.UpdateUserResponse{User: user}, nil
}

// DeleteUser deletes a user by ID
func (s *userServiceServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	// 1. Validate input
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user ID must be positive")
	}

	// 2. Delete user
	err := s.repo.Delete(ctx, req.Id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "user with ID %d not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &pb.DeleteUserResponse{Success: true}, nil
}

// ListUsers retrieves a paginated list of users
func (s *userServiceServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	// 1. Validate and normalize pagination parameters
	pageSize := req.PageSize
	pageNumber := req.Page

	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		return nil, status.Errorf(codes.InvalidArgument, "page_size too large, max %d, got %d", maxPageSize, pageSize)
	}
	if pageNumber < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "page_number must be non-negative, got %d", pageNumber)
	}

	// Calculate offset (page_number 0 = first page with offset 0)
	offset := pageNumber * pageSize

	// 2. Retrieve users from repository
	users, total, err := s.repo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	// 3. Build and return response
	return &pb.ListUsersResponse{
		User:  users,
		Total: total,
	}, nil
}

// Login authenticates a user and returns JWT tokens
func (s *userServiceServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// 1. Validate input
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	// 2. Get user by email with password hash
	userWithPassword, err := s.repo.GetByEmailWithPassword(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	// 3. Verify password
	if !auth.CheckPassword(req.Password, userWithPassword.PasswordHash) {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}

	// 4. Generate access and refresh tokens
	accessToken, err := s.tokenManager.GenerateToken(userWithPassword.User.Id, userWithPassword.User.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate access token: %v", err)
	}

	refreshToken, err := s.tokenManager.GenerateRefreshToken(userWithPassword.User.Id, userWithPassword.User.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate refresh token: %v", err)
	}

	// 5. Return successful login response
	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userWithPassword.User,
		Message:      "login successful",
	}, nil
}

// ValidateToken verifies JWT token validity and returns claims
func (s *userServiceServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	// 1. Validate input
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	// 2. Validate and parse token
	claims, err := s.tokenManager.ValidateToken(req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Valid:   false,
			Message: "invalid or expired token",
		}, nil
	}

	// 3. Return validation result with claims
	return &pb.ValidateTokenResponse{
		Valid:   true,
		UserId:  claims.UserID,
		Email:   claims.Email,
		Message: "token is valid",
	}, nil
}

// Logout handles user logout (stateless JWT)
// Note: For stateless JWT, logout is handled client-side by removing the token.
// In production, consider implementing a token blacklist using Redis for added security.
func (s *userServiceServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	return &pb.LogoutResponse{
		Success: true,
		Message: "logout successful, please remove token from client",
	}, nil
}
