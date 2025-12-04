package server

import (
	"context"
	"errors"
	"strings"

	"service-1-user/internal/auth"
	"service-1-user/internal/repository"
	"service-1-user/internal/response"
	pb "service-1-user/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultPageSize    = 10
	maxPageSize        = 100
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
		return nil, status.Error(codes.InvalidArgument, "user ID must be positive")
	}

	user, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, response.GRPCError(codes.NotFound, "user not found")
		}
		return nil, response.GRPCError(codes.Internal, "failed to get user")
	}

	return response.GetUserSuccess(user), nil
}

// CreateUser creates a new user with optional password
func (s *userServiceServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// Validate input
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if !isValidEmail(req.Email) {
		return nil, status.Error(codes.InvalidArgument, "invalid email format")
	}

	var user *pb.User
	var err error

	if req.Password != "" {
		passwordHash, err := auth.HashPassword(req.Password)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to hash password")
		}

		user, err = s.repo.CreateWithPassword(ctx, req.Name, req.Email, passwordHash)
		if err != nil {
			if isDuplicateError(err) {
				return nil, response.GRPCError(codes.AlreadyExists, "email is already registered")
			}
			return nil, response.GRPCError(codes.Internal, "failed to create user")
		}
	} else {
		// Create user without password (legacy support)
		user, err = s.repo.Create(ctx, req.Name, req.Email)
		if err != nil {
			if isDuplicateError(err) {
				return nil, response.GRPCError(codes.AlreadyExists, "email is already registered")
			}
			return nil, response.GRPCError(codes.Internal, "failed to create user")
		}
	}

	return response.CreateUserSuccess(user), nil
}

// isDuplicateError checks if error is related to duplicate key constraint violation
func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, errDuplicateKey) || strings.Contains(errMsg, errUniqueViolation)
}

// isValidEmail validates email format using regex
func isValidEmail(email string) bool {
	if len(email) == 0 {
		return false
	}
	// Simple email validation
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// UpdateUser updates user information (partial update supported)
// Only non-empty fields will be updated
func (s *userServiceServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	// Validate input
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user ID must be positive")
	}

	// At least one field must be provided
	if req.Name == "" && req.Email == "" && req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "at least one field must be provided for update")
	}

	// Prepare pointers for partial update
	var name, email, passwordHash *string

	// Process name
	if req.Name != "" {
		name = &req.Name
	}

	// Process email
	if req.Email != "" {
		if !isValidEmail(req.Email) {
			return nil, status.Error(codes.InvalidArgument, "invalid email format")
		}
		email = &req.Email
	}

	// Process password
	if req.Password != "" {
		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to hash password")
		}
		passwordHash = &hash
	}

	// Update user with only provided fields
	user, err := s.repo.PartialUpdate(ctx, req.Id, name, email, passwordHash)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, response.GRPCError(codes.NotFound, "user not found")
		}
		if errors.Is(err, repository.ErrEmailDuplicate) {
			return nil, response.GRPCError(codes.AlreadyExists, "email is already registered")
		}
		return nil, response.GRPCError(codes.Internal, "failed to update user")
	}

	return response.UpdateUserSuccess(user), nil
}

// DeleteUser deletes a user by ID
func (s *userServiceServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	// Validate input
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user ID must be positive")
	}

	// Delete user
	err := s.repo.Delete(ctx, req.Id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, response.GRPCError(codes.NotFound, "user not found")
		}
		return nil, response.GRPCError(codes.Internal, "failed to delete user")
	}

	return response.DeleteUserSuccess(), nil
}

// ListUsers retrieves a paginated list of users
func (s *userServiceServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	// Validate and normalize pagination parameters
	pageSize := req.PageSize
	pageNumber := req.Page

	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		return nil, status.Error(codes.InvalidArgument, "page_size too large")
	}
	if pageNumber < 0 {
		return nil, status.Error(codes.InvalidArgument, "page_number must be non-negative")
	}

	// Calculate offset (page_number 0 = first page with offset 0)
	offset := pageNumber * pageSize

	// Retrieve users from repository
	users, total, err := s.repo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, response.GRPCError(codes.Internal, "failed to list users")
	}

	// Calculate has_more
	hasMore := int64(pageNumber+1)*int64(pageSize) < int64(total)

	// Build and return response
	return response.ListUsersSuccess(users, int64(total), pageNumber, pageSize, hasMore), nil
}

// Login authenticates a user and returns JWT tokens
func (s *userServiceServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Validate input
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	// Get user by email with password hash
	userWithPassword, err := s.repo.GetByEmailWithPassword(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	// Verify password
	if !auth.CheckPassword(req.Password, userWithPassword.PasswordHash) {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}

	// Generate access and refresh tokens
	accessToken, err := s.tokenManager.GenerateToken(userWithPassword.User.Id, userWithPassword.User.Email)
	if err != nil {
		return nil, response.GRPCError(codes.Internal, "failed to generate access token")
	}

	refreshToken, err := s.tokenManager.GenerateRefreshToken(userWithPassword.User.Id, userWithPassword.User.Email)
	if err != nil {
		return nil, response.GRPCError(codes.Internal, "failed to generate refresh token")
	}

	// Return successful login response
	return response.LoginSuccess(accessToken, refreshToken, userWithPassword.User), nil
}

// ValidateToken verifies JWT token validity and returns claims
func (s *userServiceServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	// Validate input
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	// Validate and parse token
	claims, err := s.tokenManager.ValidateToken(ctx, req.Token)
	if err != nil {
		return nil, response.GRPCError(codes.Unauthenticated, "invalid or expired token")
	}

	// Return validation result with claims
	return response.ValidateTokenSuccess(true, int64(claims.UserID), claims.Email), nil
}

// Logout handles user logout (stateless JWT)
// Note: For stateless JWT, logout is handled client-side by removing the token.
// In production, consider implementing a token blacklist using Redis for added security.
func (s *userServiceServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	// Validate input
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	// Invalidate token
	err := s.tokenManager.InvalidateToken(ctx, req.Token)
	if err != nil {
		return nil, response.GRPCError(codes.Internal, "failed to logout")
	}

	return response.LogoutSuccess(), nil
}
