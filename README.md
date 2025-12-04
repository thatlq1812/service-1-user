# User Service

gRPC microservice for user management and authentication.

**Port:** `50051` | **Protocol:** gRPC

## Quick Start

```bash
# Setup database
psql -U postgres -c "CREATE DATABASE agrios_users;"
psql -U postgres -d agrios_users -f migrations/001_create_users_table.sql

# Run
cp .env.example .env  # Edit with your config
go run cmd/server/main.go
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | localhost | PostgreSQL host |
| `DB_PORT` | 5432 | PostgreSQL port |
| `DB_NAME` | agrios_users | Database name |
| `REDIS_ADDR` | localhost:6379 | Redis (for token blacklist) |
| `JWT_SECRET` | *required* | JWT signing secret |
| `GRPC_PORT` | 50051 | gRPC server port |

---

# API Reference

## User Management

### 1. CreateUser

Register a new user account.

**Request:**
```protobuf
message CreateUserRequest {
  string name = 1;      // Required
  string email = 2;     // Required, unique
  string password = 3;  // Required
}
```

**Example:**
```bash
grpcurl -plaintext -d '{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "secret123"
}' localhost:50051 user.UserService/CreateUser
```

**Response:**
```json
{
  "code": "000",
  "message": "success",
  "data": {
    "user": {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    }
  }
}
```

---

### 2. GetUser

Get user by ID.

**Request:**
```protobuf
message GetUserRequest {
  int32 id = 1;  // Required
}
```

**Example:**
```bash
grpcurl -plaintext -d '{"id": 1}' localhost:50051 user.UserService/GetUser
```

**Response:**
```json
{
  "code": "000",
  "message": "success",
  "data": {
    "user": {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    }
  }
}
```

---

### 3. UpdateUser

Update user information (partial update supported).

**Request:**
```protobuf
message UpdateUserRequest {
  int32 id = 1;               // Required
  optional string name = 2;    // Optional
  optional string email = 3;   // Optional
  optional string password = 4; // Optional
}
```

**Example:**
```bash
grpcurl -plaintext -d '{
  "id": 1,
  "name": "John Updated"
}' localhost:50051 user.UserService/UpdateUser
```

**Response:**
```json
{
  "code": "000",
  "message": "success",
  "data": {
    "user": {
      "id": 1,
      "name": "John Updated",
      "email": "john@example.com",
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T12:00:00Z"
    }
  }
}
```

---

### 4. DeleteUser

Delete user account.

**Request:**
```protobuf
message DeleteUserRequest {
  int32 id = 1;  // Required
}
```

**Example:**
```bash
grpcurl -plaintext -d '{"id": 1}' localhost:50051 user.UserService/DeleteUser
```

**Response:**
```json
{
  "code": "000",
  "message": "success",
  "data": {
    "success": true
  }
}
```

---

### 5. ListUsers

Get paginated user list.

**Request:**
```protobuf
message ListUsersRequest {
  int32 page = 1;       // Default: 1
  int32 page_size = 2;  // Default: 10
}
```

**Example:**
```bash
grpcurl -plaintext -d '{
  "page": 1,
  "page_size": 10
}' localhost:50051 user.UserService/ListUsers
```

**Response:**
```json
{
  "code": "000",
  "message": "success",
  "data": {
    "users": [
      {"id": 1, "name": "John Doe", "email": "john@example.com"},
      {"id": 2, "name": "Jane Smith", "email": "jane@example.com"}
    ],
    "total": 25,
    "page": 1,
    "size": 10,
    "has_more": true
  }
}
```

---

## Authentication

### 6. Login

Authenticate user and get JWT tokens.

**Request:**
```protobuf
message LoginRequest {
  string email = 1;     // Required
  string password = 2;  // Required
}
```

**Example:**
```bash
grpcurl -plaintext -d '{
  "email": "john@example.com",
  "password": "secret123"
}' localhost:50051 user.UserService/Login
```

**Response:**
```json
{
  "code": "000",
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com"
    }
  }
}
```

---

### 7. ValidateToken

Validate JWT token (used by other services).

**Request:**
```protobuf
message ValidateTokenRequest {
  string token = 1;  // Required
}
```

**Example:**
```bash
grpcurl -plaintext -d '{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}' localhost:50051 user.UserService/ValidateToken
```

**Response (valid):**
```json
{
  "code": "000",
  "message": "success",
  "data": {
    "valid": true,
    "user_id": 1,
    "email": "john@example.com"
  }
}
```

**Response (invalid):**
```json
{
  "code": "005",
  "message": "invalid or expired token",
  "data": {
    "valid": false
  }
}
```

---

### 8. Logout

Invalidate token (add to Redis blacklist).

**Request:**
```protobuf
message LogoutRequest {
  string token = 1;  // Required
}
```

**Example:**
```bash
grpcurl -plaintext -d '{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}' localhost:50051 user.UserService/Logout
```

**Response:**
```json
{
  "code": "000",
  "message": "success",
  "data": {
    "success": true
  }
}
```

---

## Error Codes

| Code | Description |
|------|-------------|
| `000` | Success |
| `001` | Email already exists |
| `002` | Validation error |
| `003` | User not found |
| `004` | Invalid credentials |
| `005` | Invalid/expired token |
| `500` | Internal server error |
