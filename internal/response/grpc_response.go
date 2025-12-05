package response

import (
	pb "github.com/thatlq1812/service-1-user/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Standard response codes mapping
const (
	CodeSuccess            = "000" // Success
	CodeUnknownError       = "002" // Unknown error
	CodeInvalidRequest     = "003" // Invalid request
	CodeNotFound           = "005" // Not found
	CodeAlreadyExists      = "006" // Already exists
	CodePermissionDenied   = "007" // Permission denied
	CodeInternalError      = "013" // Internal error
	CodeUnauthenticated    = "014" // Authentication required
	CodeServiceUnavailable = "015" // Service unavailable
	CodeUnauthorized       = "016" // Unauthorized
)

// MapGRPCCodeToString converts gRPC status code to our standard string code
func MapGRPCCodeToString(code codes.Code) string {
	switch code {
	case codes.OK:
		return CodeSuccess
	case codes.InvalidArgument:
		return CodeInvalidRequest
	case codes.NotFound:
		return CodeNotFound
	case codes.AlreadyExists:
		return CodeAlreadyExists
	case codes.PermissionDenied:
		return CodePermissionDenied
	case codes.Unauthenticated:
		return CodeUnauthenticated
	case codes.Internal:
		return CodeInternalError
	case codes.Unavailable:
		return CodeServiceUnavailable
	case codes.FailedPrecondition:
		return CodeUnauthorized
	default:
		return CodeUnknownError
	}
}

// User Service Response Helpers

func CreateUserSuccess(user *pb.User) *pb.CreateUserResponse {
	return &pb.CreateUserResponse{
		Code:    CodeSuccess,
		Message: "User created successfully",
		Data: &pb.CreateUserData{
			User: user,
		},
	}
}

func GetUserSuccess(user *pb.User) *pb.GetUserResponse {
	return &pb.GetUserResponse{
		Code:    CodeSuccess,
		Message: "User retrieved successfully",
		Data: &pb.GetUserData{
			User: user,
		},
	}
}

func UpdateUserSuccess(user *pb.User) *pb.UpdateUserResponse {
	return &pb.UpdateUserResponse{
		Code:    CodeSuccess,
		Message: "User updated successfully",
		Data: &pb.UpdateUserData{
			User: user,
		},
	}
}

func DeleteUserSuccess() *pb.DeleteUserResponse {
	return &pb.DeleteUserResponse{
		Code:    CodeSuccess,
		Message: "User deleted successfully",
		Data: &pb.DeleteUserData{
			Success: true,
		},
	}
}

func ListUsersSuccess(users []*pb.User, total int64, page, size int32, hasMore bool) *pb.ListUsersResponse {
	return &pb.ListUsersResponse{
		Code:    CodeSuccess,
		Message: "Users listed successfully",
		Data: &pb.ListUsersData{
			Users:   users,
			Total:   total,
			Page:    page,
			Size:    size,
			HasMore: hasMore,
		},
	}
}

func LoginSuccess(accessToken, refreshToken string) *pb.LoginResponse {
	return &pb.LoginResponse{
		Code:    CodeSuccess,
		Message: "Login successful",
		Data: &pb.LoginData{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}
}

func ValidateTokenSuccess(valid bool, userID int64, email string) *pb.ValidateTokenResponse {
	return &pb.ValidateTokenResponse{
		Code:    CodeSuccess,
		Message: "Token validated successfully",
		Data: &pb.ValidateTokenData{
			Valid:  valid,
			UserId: userID,
			Email:  email,
		},
	}
}

func LogoutSuccess() *pb.LogoutResponse {
	return &pb.LogoutResponse{
		Code:    CodeSuccess,
		Message: "Logout successful",
		Data: &pb.LogoutData{
			Success: true,
		},
	}
}

func RefreshTokenSuccess(accessToken, refreshToken string) *pb.RefreshTokenResponse {
	return &pb.RefreshTokenResponse{
		Code:    CodeSuccess,
		Message: "Token refreshed successfully",
		Data: &pb.RefreshTokenData{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}
}

// Error response helper
func GRPCError(code codes.Code, message string) error {
	// Add hints based on code
	hint := ""
	switch code {
	case codes.InvalidArgument:
		hint = " Check input parameters for validity."
	case codes.NotFound:
		hint = " Verify the resource ID exists."
	case codes.Unauthenticated:
		hint = " Provide valid authentication credentials."
	case codes.PermissionDenied:
		hint = " Ensure you have the required permissions."
	case codes.Internal:
		hint = " Contact support if the issue persists."
	default:
		hint = ""
	}
	fullMessage := message + hint
	return status.Error(code, fullMessage)
}

// Error with custom code
func GRPCErrorWithCode(code codes.Code, message string) error {
	return status.Error(code, message)
}
