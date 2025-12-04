package server

import (
	"context"
	"errors"
	"strings"

	"agrios/pkg/common"
	"service-1-user/internal/auth"
	"service-1-user/internal/repository"
	pb "service-1-user/proto"
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
	if req.Id < 0 {
		return &pb.GetUserResponse{
			Code:    common.CodeInvalidArgument,
			Message: "user ID must be positive",
		}, nil
	}

	user, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return &pb.GetUserResponse{
				Code:    common.CodeNotFound,
				Message: "user not found",
			}, nil
		}
		return &pb.GetUserResponse{
			Code:    common.CodeInternal,
			Message: "failed to get user",
		}, nil
	}

	return &pb.GetUserResponse{
		Code:    common.CodeSuccess,
		Message: "success",
		User:    user,
	}, nil
}

// CreateUser creates a new user with optional password
func (s *userServiceServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// 1. Validate input
	if req.Name == "" {
		return &pb.CreateUserResponse{
			Code:    common.CodeInvalidArgument,
			Message: "name is required",
		}, nil
	}

	if req.Email == "" {
		return &pb.CreateUserResponse{
			Code:    common.CodeInvalidArgument,
			Message: "email is required",
		}, nil
	}

	if !common.IsValidEmail(req.Email) {
		return &pb.CreateUserResponse{
			Code:    common.CodeInvalidArgument,
			Message: "invalid email format",
		}, nil
	}

	var user *pb.User
	var err error

	if req.Password != "" {
		passwordHash, err := auth.HashPassword(req.Password)
		if err != nil {
			return &pb.CreateUserResponse{
				Code:    common.CodeInternal,
				Message: "failed to hash password",
			}, nil
		}

		user, err = s.repo.CreateWithPassword(ctx, req.Name, req.Email, passwordHash)
		if err != nil {
			if isDuplicateError(err) {
				return &pb.CreateUserResponse{
					Code:    common.CodeAlreadyExists,
					Message: "email is already registered",
				}, nil
			}
			return &pb.CreateUserResponse{
				Code:    common.CodeInternal,
				Message: "email is already registered",
			}, nil
		}
	} else {
		// Create user without password (legacy support)
		user, err = s.repo.Create(ctx, req.Name, req.Email)
		if err != nil {
			if isDuplicateError(err) {
				return &pb.CreateUserResponse{
					Code:    common.CodeAlreadyExists,
					Message: "email is already registered",
				}, nil
			}
			return &pb.CreateUserResponse{
				Code:    common.CodeInternal,
				Message: "failed to create user",
			}, nil
		}
	}

	return &pb.CreateUserResponse{
		Code:    common.CodeSuccess,
		Message: "success",
		User:    user,
	}, nil
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
	if req.Id < 0 {
		return &pb.UpdateUserResponse{
			Code:    common.CodeInvalidArgument,
			Message: "user ID must be positive",
		}, nil
	}
	if req.Name == "" {
		return &pb.UpdateUserResponse{
			Code:    common.CodeInvalidArgument,
			Message: "name is required",
		}, nil
	}
	if req.Email == "" {
		return &pb.UpdateUserResponse{
			Code:    common.CodeInvalidArgument,
			Message: "invalid email format",
		}, nil
	}

	if !common.IsValidEmail(req.Email) {
		return &pb.UpdateUserResponse{
			Code:    common.CodeInvalidArgument,
			Message: "invalid email format",
		}, nil
	}

	// 3. Update user
	user, err := s.repo.Update(ctx, req.Id, req.Name, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return &pb.UpdateUserResponse{
				Code:    common.CodeNotFound,
				Message: "user not found",
			}, nil
		}
		if isDuplicateError(err) {
			return &pb.UpdateUserResponse{
				Code:    common.CodeAlreadyExists,
				Message: "email is already registered",
			}, nil
		}
		return &pb.UpdateUserResponse{
			Code:    common.CodeInternal,
			Message: "failed to update user",
		}, nil
	}

	return &pb.UpdateUserResponse{
		Code:    common.CodeSuccess,
		Message: "success",
		User:    user,
	}, nil
}

// DeleteUser deletes a user by ID
func (s *userServiceServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	// 1. Validate input
	if req.Id <= 0 {
		return &pb.DeleteUserResponse{
			Code:    common.CodeInvalidArgument,
			Message: "user ID must be positive",
		}, nil
	}

	// 2. Delete user
	err := s.repo.Delete(ctx, req.Id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return &pb.DeleteUserResponse{
				Code:    common.CodeNotFound,
				Message: "user not found",
			}, nil
		}
		return &pb.DeleteUserResponse{
			Code:    common.CodeInternal,
			Message: "failed to delete user",
		}, nil
	}

	return &pb.DeleteUserResponse{
		Code:    common.CodeSuccess,
		Message: "success",
		Success: true,
	}, nil
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
		return &pb.ListUsersResponse{
			Code:    common.CodeInvalidArgument,
			Message: "page_size too large",
		}, nil
	}
	if pageNumber < 0 {
		return &pb.ListUsersResponse{
			Code:    common.CodeInvalidArgument,
			Message: "page_number must be non-negative",
		}, nil
	}

	// Calculate offset (page_number 0 = first page with offset 0)
	offset := pageNumber * pageSize

	// 2. Retrieve users from repository
	users, total, err := s.repo.List(ctx, pageSize, offset)
	if err != nil {
		return &pb.ListUsersResponse{
			Code:    common.CodeInternal,
			Message: "failed to list users",
		}, nil
	}

	// 3. Calculate has_more
	hasMore := int64(pageNumber+1)*int64(pageSize) < int64(total)

	// 4. Build and return response
	return &pb.ListUsersResponse{
		Code:    common.CodeSuccess,
		Message: "success",
		Users:   users,
		Total:   int64(total),
		Page:    pageNumber,
		Size:    pageSize,
		HasMore: hasMore,
	}, nil
}

// Login authenticates a user and returns JWT tokens
func (s *userServiceServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// 1. Validate input
	if req.Email == "" {
		return &pb.LoginResponse{
			Code:    common.CodeInvalidArgument,
			Message: "email is required",
		}, nil
	}
	if req.Password == "" {
		return &pb.LoginResponse{
			Code:    common.CodeInvalidArgument,
			Message: "password is required",
		}, nil
	}

	// 2. Get user by email with password hash
	userWithPassword, err := s.repo.GetByEmailWithPassword(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return &pb.LoginResponse{
				Code:    common.CodeUnauthorized,
				Message: "invalid email or password",
			}, nil
		}
		return &pb.LoginResponse{
			Code:    common.CodeInternal,
			Message: "failed to get user",
		}, nil
	}

	// 3. Verify password
	if !auth.CheckPassword(req.Password, userWithPassword.PasswordHash) {
		return &pb.LoginResponse{
			Code:    common.CodeUnauthorized,
			Message: "invalid email or password",
		}, nil
	}

	// 4. Generate access and refresh tokens
	accessToken, err := s.tokenManager.GenerateToken(userWithPassword.User.Id, userWithPassword.User.Email)
	if err != nil {
		return &pb.LoginResponse{
			Code:    common.CodeInternal,
			Message: "failed to generate access token",
		}, nil
	}

	refreshToken, err := s.tokenManager.GenerateRefreshToken(userWithPassword.User.Id, userWithPassword.User.Email)
	if err != nil {
		return &pb.LoginResponse{
			Code:    common.CodeInternal,
			Message: "failed to generate refresh token",
		}, nil
	}

	// 5. Return successful login response
	return &pb.LoginResponse{
		Code:         common.CodeSuccess,
		Message:      "success",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userWithPassword.User,
	}, nil
}

// ValidateToken verifies JWT token validity and returns claims
func (s *userServiceServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	// 1. Validate input
	if req.Token == "" {
		return &pb.ValidateTokenResponse{
			Code:    common.CodeInvalidArgument,
			Message: "token is required",
			Valid:   false,
		}, nil
	}

	// 2. Validate and parse token
	claims, err := s.tokenManager.ValidateToken(ctx, req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Code:    common.CodeUnauthorized,
			Message: "invalid or expired token",
			Valid:   false,
		}, nil
	}

	// 3. Return validation result with claims
	return &pb.ValidateTokenResponse{
		Code:    common.CodeSuccess,
		Message: "success",
		Valid:   true,
		UserId:  int64(claims.UserID),
		Email:   claims.Email,
	}, nil
}

// Logout handles user logout (stateless JWT)
// Note: For stateless JWT, logout is handled client-side by removing the token.
// In production, consider implementing a token blacklist using Redis for added security.
func (s *userServiceServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	// validate input
	if req.Token == "" {
		return &pb.LogoutResponse{
			Code:    common.CodeInvalidArgument,
			Message: "token is required",
		}, nil
	}

	// invalidate token
	err := s.tokenManager.InvalidateToken(ctx, req.Token)
	if err != nil {
		return &pb.LogoutResponse{
			Code:    common.CodeInternal,
			Message: "failed to logout",
		}, nil
	}

	return &pb.LogoutResponse{
		Code:    common.CodeSuccess,
		Message: "success",
		Success: true,
	}, nil
}
