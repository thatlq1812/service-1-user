# Service 1: User Service

> Authentication and user management microservice with JWT tokens and Redis blacklist

**Protocol:** gRPC  
**Port:** 50051  
**Database:** PostgreSQL + Redis

---

## Table of Contents

- [Quick Start](#quick-start)
- [Overview](#overview)
- [Setup Options](#setup-options)
  - [Option 1: Docker](#option-1-docker-recommended)
  - [Option 2: Terminal (Local)](#option-2-terminal-local-development)
- [Environment Configuration](#environment-configuration)
- [API Reference](#api-reference)
- [Database Schema](#database-schema)
- [Testing](#testing)
- [Troubleshooting](#troubleshooting)

---

## Quick Start

**For new users cloning the project:**

```bash
# Clone and setup
git clone https://github.com/thatlq1812/agrios.git
cd agrios

# Configure (optional - defaults work fine)
cp service-1-user/.env.example service-1-user/.env

# Start services
docker-compose up -d

# Wait for initialization
sleep 15

# Verify
docker logs agrios-user-service

# Test
grpcurl -plaintext localhost:50051 list
```

**Service will be running on port 50051**

See [Setup Options](#setup-options) for detailed instructions.

---

## Overview

User Service handles authentication and user management for the Agrios platform.

**Features:**
- JWT authentication with dual-token system (access + refresh)
- User CRUD operations
- Password hashing with bcrypt
- Token blacklist with Redis
- Token rotation for security
- Email uniqueness validation

**Technology Stack:**
- **Language:** Go 1.21+
- **Protocol:** gRPC
- **Database:** PostgreSQL 15
- **Cache:** Redis 7
- **Auth:** JWT with bcrypt

---

## Setup Options

### Option 1: Docker (Recommended)

**Prerequisites:**
- Docker 20.10+
- Docker Compose 1.29+
- Git

**Quick Start from Scratch:**

```bash
# 1. Clone repository
git clone https://github.com/thatlq1812/agrios.git
cd agrios

# 2. Configure environment (use defaults or customize)
cp service-1-user/.env.example service-1-user/.env
# Optional: Edit service-1-user/.env if needed

# 3. Start all services with Docker Compose
# This single command will:
#   - Pull and start PostgreSQL container (port 5432)
#   - Pull and start Redis container (port 6379)
#   - Build and start User Service container (port 50051)
#   - Create databases and run migrations automatically
docker-compose up -d

# 4. Wait for services to initialize and become healthy
sleep 15

# 5. Check service status
docker-compose ps

# Expected output:
# NAME                   STATUS
# agrios-postgres        Up (healthy)   <- PostgreSQL database
# agrios-redis           Up (healthy)   <- Redis cache
# agrios-user-service    Up (healthy)   <- User Service (gRPC)

# 6. View logs to confirm service is running
docker logs agrios-user-service --tail 20

# Expected output:
# Connected to PostgreSQL successfully
# Connected to Redis successfully
# User Service (gRPC) listening on port 50051
```

**Important Notes:**
- **PostgreSQL & Redis are started automatically** by Docker Compose - no manual installation needed
- Database tables are created automatically from `migrations/001_create_users_table.sql`
- All services run in isolated Docker containers with networking configured
- Data persists in Docker volumes even after stopping containers

**Database Migration:**

The database tables are created automatically when the service starts. The migration file `migrations/001_create_users_table.sql` is executed on first run.

**Verify PostgreSQL and Redis in Docker:**

```bash
# Check PostgreSQL is running and accessible
docker exec agrios-postgres psql -U postgres -c "SELECT version();"

# Check Redis is running and accessible  
docker exec agrios-redis redis-cli ping
# Expected: PONG

# View PostgreSQL databases
docker exec agrios-postgres psql -U postgres -c "\l"
# Should see: agrios_users, agrios_articles

# Check users table exists
docker exec agrios-postgres psql -U postgres -d agrios_users -c "\dt"
# Should see: users table

# Check Redis keys (token blacklist)
docker exec agrios-redis redis-cli KEYS "*"
```

**Verify User Service is Working:**

```bash
# Install grpcurl (if not already installed)
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List available services
grpcurl -plaintext localhost:50051 list

# Expected: user.UserService

# Test CreateUser
grpcurl -plaintext \
  -d '{"name":"Test User","email":"test@example.com","password":"TestPass123"}' \
  localhost:50051 user.UserService.CreateUser

# Expected: Success response with user data
```

**User Service Docker Details:**
```yaml
# From docker-compose.yml
user-service:
  build: ./service-1-user
  ports:
    - "50051:50051"
  depends_on:
    - postgres
    - redis
  environment:
    - DB_HOST=postgres
    - REDIS_HOST=redis
```

**Rebuild after code changes:**
```bash
docker-compose up -d --build user-service
```

---

### Option 2: Terminal (Local Development)

**Prerequisites:**
- Go 1.21+
- PostgreSQL 15+
- Redis 7+

#### Step 1: Install Dependencies

```bash
cd service-1-user

# Download Go dependencies
go mod download

# Verify dependencies
go mod verify
```

#### Step 2: Setup Database

**PostgreSQL:**
```bash
# Create database
psql -U postgres -c "CREATE DATABASE agrios_users;"

# Run migration
psql -U postgres -d agrios_users -f migrations/001_create_users_table.sql

# Verify
psql -U postgres -d agrios_users -c "\dt"
```

**Redis:**
```bash
# Start Redis server
redis-server

# In another terminal, verify
redis-cli ping
# Expected: PONG
```

#### Step 3: Configure Environment

```bash
# Copy example
cp .env.example .env

# Edit configuration
nano .env
```

**Required settings for local development:**
```env
# Database (local PostgreSQL)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=agrios_users

# Redis (local)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT
JWT_SECRET=your-super-secret-key
ACCESS_TOKEN_DURATION=15m
REFRESH_TOKEN_DURATION=168h

# Server
GRPC_PORT=50051
```

#### Step 4: Build and Run

```bash
# Build
go build -o bin/user-service ./cmd/server

# Run
./bin/user-service

# Or run directly
go run cmd/server/main.go
```

**Expected output:**
```
2025/12/05 10:00:00 Connected to PostgreSQL
2025/12/05 10:00:00 Connected to Redis
2025/12/05 10:00:00 User Service listening on :50051
```

#### Step 5: Verify Service

```bash
# Install grpcurl (if not installed)
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List available services
grpcurl -plaintext localhost:50051 list

# Expected output:
# grpc.reflection.v1alpha.ServerReflection
# user.UserService

# Test CreateUser
grpcurl -plaintext \
  -d '{"name":"Test User","email":"test@example.com","password":"pass123"}' \
  localhost:50051 user.UserService.CreateUser
```

---

## Environment Configuration

### Complete Environment Variables

```env
# Database Configuration
DB_HOST=localhost               # PostgreSQL host (use 'postgres' for Docker)
DB_PORT=5432                    # PostgreSQL port
DB_USER=postgres                # Database user
DB_PASSWORD=yourpassword        # Database password
DB_NAME=agrios_users            # Database name

# Redis Configuration
REDIS_HOST=localhost            # Redis host (use 'redis' for Docker)
REDIS_PORT=6379                 # Redis port
REDIS_PASSWORD=                 # Redis password (optional)

# JWT Configuration
JWT_SECRET=your-super-secret-key-change-in-production
ACCESS_TOKEN_DURATION=15m       # Access token expiration (15 minutes)
REFRESH_TOKEN_DURATION=168h     # Refresh token expiration (7 days = 168 hours)

# Server Configuration
GRPC_PORT=50051                 # gRPC server port
LOG_LEVEL=info                  # Logging level (debug, info, warn, error)
```

### Security Recommendations

**Production Settings:**
```env
# Use strong JWT secret (32+ characters)
JWT_SECRET=$(openssl rand -base64 32)

# Use secure database credentials
DB_PASSWORD=$(openssl rand -base64 16)

# Enable Redis password
REDIS_PASSWORD=$(openssl rand -base64 16)

# Adjust token durations for security
ACCESS_TOKEN_DURATION=15m       # Keep short
REFRESH_TOKEN_DURATION=168h     # 7 days maximum
```

---

## API Reference

### gRPC Service Definition

```protobuf
service UserService {
  // User Management
  rpc CreateUser (CreateUserRequest) returns (CreateUserResponse);
  rpc GetUser (GetUserRequest) returns (GetUserResponse);
  rpc UpdateUser (UpdateUserRequest) returns (UpdateUserResponse);
  rpc DeleteUser (DeleteUserRequest) returns (DeleteUserResponse);
  rpc ListUsers (ListUsersRequest) returns (ListUsersResponse);
  
  // Authentication
  rpc Login (LoginRequest) returns (LoginResponse);
  rpc ValidateToken (ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc RefreshToken (RefreshTokenRequest) returns (RefreshTokenResponse);
  rpc Logout (LogoutRequest) returns (LogoutResponse);
}
```

### 1. CreateUser

Register a new user account.

**Request:**
```bash
grpcurl -plaintext \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securepass123"
  }' \
  localhost:50051 user.UserService.CreateUser
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
      "createdAt": "2025-12-05T10:00:00Z",
      "updatedAt": "2025-12-05T10:00:00Z"
    }
  }
}
```

**Validation:**
- Name: Required, min 2 characters
- Email: Required, valid email format, unique
- Password: Required, min 8 characters

---

### 2. Login

Authenticate user and receive JWT tokens.

**Request:**
```bash
grpcurl -plaintext \
  -d '{
    "email": "john@example.com",
    "password": "securepass123"
  }' \
  localhost:50051 user.UserService.Login
```

**Response:**
```json
{
  "code": "000",
  "message": "Login successful",
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Token Details:**
- **Access Token**: 15 minutes validity, use for API requests
- **Refresh Token**: 7 days validity, use to get new access token

**Save tokens:**
```bash
LOGIN_RESP=$(grpcurl -plaintext \
  -d '{"email":"john@example.com","password":"securepass123"}' \
  localhost:50051 user.UserService.Login)

ACCESS_TOKEN=$(echo "$LOGIN_RESP" | grep 'accessToken' | cut -d'"' -f4)
REFRESH_TOKEN=$(echo "$LOGIN_RESP" | grep 'refreshToken' | cut -d'"' -f4)
```

---

### 3. ValidateToken

Check if access token is valid.

**Request:**
```bash
grpcurl -plaintext \
  -d "{\"token\":\"$ACCESS_TOKEN\"}" \
  localhost:50051 user.UserService.ValidateToken
```

**Response (Valid):**
```json
{
  "code": "000",
  "message": "success",
  "data": {
    "valid": true,
    "userId": 1,
    "email": "john@example.com"
  }
}
```

**Response (Invalid):**
```json
{
  "code": "004",
  "message": "Invalid or expired token",
  "data": {
    "valid": false
  }
}
```

---

### 4. RefreshToken

Get new access token using refresh token.

**Request:**
```bash
grpcurl -plaintext \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}" \
  localhost:50051 user.UserService.RefreshToken
```

**Response:**
```json
{
  "code": "000",
  "message": "success",
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Security Features:**
- Old refresh token is invalidated (token rotation)
- New refresh token issued with new expiration
- Old refresh token added to Redis blacklist

---

### 5. Logout

Invalidates tokens immediately for complete logout.

**Request (Complete Logout - Recommended):**
```bash
# Blacklist BOTH access and refresh tokens
grpcurl -plaintext \
  -d "{\"token\":\"$ACCESS_TOKEN\",\"refresh_token\":\"$REFRESH_TOKEN\"}" \
  localhost:50051 user.UserService.Logout
```

**Request (Access Token Only - Partial Logout):**
```bash
# Blacklist access token only (refresh token can still be used)
grpcurl -plaintext \
  -d "{\"token\":\"$ACCESS_TOKEN\"}" \
  localhost:50051 user.UserService.Logout
```

**Response:**
```json
{
  "code": "000",
  "message": "Logout successful"
}
```

**Implementation:**
- Access token added to Redis blacklist (required)
- Refresh token added to Redis blacklist (if provided)
- Both tokens cannot be reused after logout
- Tokens auto-deleted from blacklist when expired (TTL)
- Token remains invalid until natural expiration
- Refresh token should be discarded by client

---

### 6. GetUser

Retrieve user by ID.

**Request:**
```bash
grpcurl -plaintext \
  -d '{"id": 1}' \
  localhost:50051 user.UserService.GetUser
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
      "createdAt": "2025-12-05T10:00:00Z",
      "updatedAt": "2025-12-05T10:00:00Z"
    }
  }
}
```

---

### 7. UpdateUser

Update user information (partial update supported).

**Request:**
```bash
grpcurl -plaintext \
  -d '{
    "id": 1,
    "name": "John Smith",
    "email": "johnsmith@example.com"
  }' \
  localhost:50051 user.UserService.UpdateUser
```

**Response:**
```json
{
  "code": "000",
  "message": "User updated successfully",
  "data": {
    "user": {
      "id": 1,
      "name": "John Smith",
      "email": "johnsmith@example.com",
      "updatedAt": "2025-12-05T11:00:00Z"
    }
  }
}
```

**Note:** Only provided fields are updated (partial update)

---

### 8. DeleteUser

Delete user account.

**Request:**
```bash
grpcurl -plaintext \
  -d '{"id": 1}' \
  localhost:50051 user.UserService.DeleteUser
```

**Response:**
```json
{
  "code": "000",
  "message": "User deleted successfully"
}
```

---

### 9. ListUsers

Get paginated list of users.

**Request:**
```bash
grpcurl -plaintext \
  -d '{
    "page": 1,
    "page_size": 10
  }' \
  localhost:50051 user.UserService.ListUsers
```

**Response:**
```json
{
  "code": "000",
  "message": "success",
  "data": {
    "users": [
      {
        "id": 1,
        "name": "John Doe",
        "email": "john@example.com",
        "createdAt": "2025-12-05T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 10,
      "totalCount": 25,
      "totalPages": 3
    }
  }
}
```

---

## Database Schema

### Users Table

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
```

### Redis Keys

**Token Blacklist System:**
```
Key Format: blacklist:<full-token-string>
Value: "revoked"
TTL: Automatically set to remaining token lifetime
```

**Blacklist Algorithm:**

```
┌─────────────────────────────────────────────────────────────────┐
│                    Token Blacklist Flow                          │
└─────────────────────────────────────────────────────────────────┘

1. InvalidateToken(token) called by:
   ├─> Logout: Blacklist access token (required) + refresh token (optional)
   └─> RefreshToken: Blacklist old refresh token (automatic rotation)

2. Parse token without signature verification:
   token = jwt.ParseUnverified(tokenString)
   claims = token.Claims

3. Calculate remaining lifetime:
   timeRemaining = claims.ExpiresAt - time.Now()
   
   if timeRemaining <= 0:
       return nil  // Already expired, no need to blacklist

4. Store in Redis with TTL:
   key = "blacklist:" + tokenString
   value = "revoked"
   TTL = timeRemaining
   
   Redis.Set(key, value, TTL)

5. Auto-cleanup by Redis:
   - When TTL reaches 0, Redis automatically deletes key
   - No manual cleanup needed
   - Memory efficient
```

**Token Validation Algorithm:**

```
ValidateToken(token):
├─> 1. Check Redis blacklist:
│      key = "blacklist:" + token
│      if Redis.Exists(key):
│          return Error("token has been revoked")
│
├─> 2. Verify JWT signature:
│      claims, err = jwt.ParseWithClaims(token, secretKey)
│      if err:
│          return Error("invalid token signature")
│
├─> 3. Check token type:
│      if claims.TokenType != "access":
│          return Error("invalid token type")
│
└─> 4. Return claims:
       return claims, nil

ValidateRefreshToken(token):
├─> Same as ValidateToken but checks TokenType == "refresh"
```

**Complete Logout Flow:**

```
Logout(accessToken, refreshToken):
├─> 1. Validate access token (required):
│      if accessToken == "":
│          return Error("access token required")
│
├─> 2. Blacklist access token:
│      err = InvalidateToken(accessToken)
│      if err:
│          return Error("failed to invalidate access token")
│      → Redis: blacklist:<access_token> = "revoked" (TTL=15min)
│
├─> 3. Blacklist refresh token (if provided):
│      if refreshToken != "":
│          err = InvalidateToken(refreshToken)
│          if err:
│              return Error("failed to invalidate refresh token")
│          → Redis: blacklist:<refresh_token> = "revoked" (TTL=7days)
│
└─> 4. Return success:
       return {success: true}
```

**Token Rotation Flow (RefreshToken):**

```
RefreshToken(oldRefreshToken):
├─> 1. Validate old refresh token:
│      claims, err = ValidateRefreshToken(oldRefreshToken)
│      → Checks blacklist first, then signature
│      if err:
│          return Error("invalid or expired refresh token")
│
├─> 2. Blacklist old refresh token (security):
│      err = InvalidateToken(oldRefreshToken)
│      → Redis: blacklist:<old_refresh> = "revoked" (TTL=7days)
│
├─> 3. Generate new access token:
│      newAccessToken = GenerateToken(claims.UserID, claims.Email)
│      → Expiration: 15 minutes
│
├─> 4. Generate new refresh token:
│      newRefreshToken = GenerateRefreshToken(claims.UserID, claims.Email)
│      → Expiration: 7 days
│
└─> 5. Return new token pair:
       return {
           access_token: newAccessToken,
           refresh_token: newRefreshToken
       }
```

**How It Works:**
1. **Logout**: Blacklist both access + refresh tokens to prevent reuse
2. **RefreshToken**: Automatically blacklist old refresh token (token rotation)
3. **ValidateToken**: Check blacklist before signature verification
4. **Auto-cleanup**: Redis TTL automatically deletes expired entries
5. **No manual cleanup needed**: TTL mechanism handles everything

**Example Commands:**
```bash
# Check if token is blacklisted (from host)
docker-compose exec redis redis-cli GET "blacklist:<token>"

# List all blacklisted tokens
docker-compose exec redis redis-cli KEYS "blacklist:*"

# Count blacklist entries
docker-compose exec redis redis-cli KEYS "blacklist:*" | wc -l

# Check TTL (seconds until auto-deletion)
docker-compose exec redis redis-cli TTL "blacklist:<token>"
# Output: 900 = 15 minutes (access token)
# Output: 604800 = 7 days (refresh token)
# Output: -2 = key doesn't exist (already deleted)
```

**Automatic Cleanup Behavior:**
- Access tokens (15 min): Auto-deleted after 15 minutes
- Refresh tokens (7 days): Auto-deleted after 7 days
- TTL counts down automatically
- Memory efficient - old entries never accumulate
- No cronjobs or background workers needed

**Redis Commands:**
```bash
# Check if token is blacklisted (from host)
docker-compose exec redis redis-cli GET "blacklist:<token>"

# List all blacklisted tokens
docker-compose exec redis redis-cli KEYS "blacklist:*"

# Count blacklist entries
docker-compose exec redis redis-cli KEYS "blacklist:*" | wc -l

# Check TTL (seconds until auto-deletion)
docker-compose exec redis redis-cli TTL "blacklist:<token>"
# Output examples:
#   900 = 15 minutes remaining (access token)
#   604800 = 7 days remaining (refresh token)
#   -1 = no TTL set (should never happen)
#   -2 = key doesn't exist (already deleted or never blacklisted)
```

**Complete Verification Test:**
```bash
# 1. Login to get both tokens
LOGIN_RESP=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@test.com","password":"pass123"}')

# Extract tokens
ACCESS_TOKEN=$(echo $LOGIN_RESP | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
REFRESH_TOKEN=$(echo $LOGIN_RESP | grep -o '"refresh_token":"[^"]*' | cut -d'"' -f4)

echo "Access Token: ${ACCESS_TOKEN:0:50}..."
echo "Refresh Token: ${REFRESH_TOKEN:0:50}..."

# 2. Check blacklist before logout
echo "Blacklist count BEFORE logout:"
docker-compose exec redis redis-cli KEYS "blacklist:*" | wc -l

# 3. Complete logout (blacklist BOTH tokens)
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"$ACCESS_TOKEN\",\"refresh_token\":\"$REFRESH_TOKEN\"}"

# 4. Check blacklist after logout
echo "Blacklist count AFTER logout:"
docker-compose exec redis redis-cli KEYS "blacklist:*" | wc -l

# 5. Verify access token is blacklisted
echo "Try to validate access token (should FAIL):"
curl -s -X POST http://localhost:8080/api/v1/auth/validate \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"$ACCESS_TOKEN\"}"
# Expected: {"code":"016","message":"invalid or expired token"}

# 6. Verify refresh token is blacklisted
echo "Try to use refresh token (should FAIL):"
curl -s -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}"
# Expected: {"code":"016","message":"invalid or expired refresh token"}

# 7. Check TTL for both tokens
echo "Check TTL for blacklisted tokens:"
docker-compose exec redis redis-cli --scan --pattern "blacklist:*" | while read key; do
  TTL=$(docker-compose exec redis redis-cli TTL "$key" | tr -d '\r')
  echo "TTL: $TTL seconds (~$(($TTL / 60)) minutes)"
done
```

**Expected Results:**
- Blacklist count increases by 2 (access + refresh tokens)
- Access token validation fails: code "016"
- Refresh token usage fails: code "016"
- Access token TTL: ~900 seconds (15 minutes)
- Refresh token TTL: ~604800 seconds (7 days)
- After TTL expires: Redis auto-deletes both entries

---

## Testing

### Unit Tests

```bash
cd service-1-user

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/auth/...

# Verbose output
go test -v ./...
```

### Integration Tests

```bash
# Ensure service is running
docker-compose up -d user-service

# Run test script
bash ../scripts/test-user-service.sh
```

### Manual Testing Workflow

```bash
# 1. Create user
grpcurl -plaintext \
  -d '{"name":"Test","email":"test@example.com","password":"pass123"}' \
  localhost:50051 user.UserService.CreateUser

# 2. Login
LOGIN_RESP=$(grpcurl -plaintext \
  -d '{"email":"test@example.com","password":"pass123"}' \
  localhost:50051 user.UserService.Login)

ACCESS_TOKEN=$(echo "$LOGIN_RESP" | grep 'accessToken' | cut -d'"' -f4)
REFRESH_TOKEN=$(echo "$LOGIN_RESP" | grep 'refreshToken' | cut -d'"' -f4)

# 3. Validate token
grpcurl -plaintext \
  -d "{\"token\":\"$ACCESS_TOKEN\"}" \
  localhost:50051 user.UserService.ValidateToken

# 4. Get user
grpcurl -plaintext -d '{"id":1}' localhost:50051 user.UserService.GetUser

# 5. Update user
grpcurl -plaintext \
  -d '{"id":1,"name":"Updated Name"}' \
  localhost:50051 user.UserService.UpdateUser

# 6. Refresh token
grpcurl -plaintext \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}" \
  localhost:50051 user.UserService.RefreshToken

# 7. Logout
grpcurl -plaintext \
  -d "{\"token\":\"$ACCESS_TOKEN\"}" \
  localhost:50051 user.UserService.Logout

# 8. Verify token is blacklisted
grpcurl -plaintext \
  -d "{\"token\":\"$ACCESS_TOKEN\"}" \
  localhost:50051 user.UserService.ValidateToken
# Expected: valid=false
```

---

## Troubleshooting

### First Time Setup Issues

**Problem:** After cloning, service won't start

**Common issues and solutions:**

1. **Docker not running**
   ```bash
   # Check Docker status
   docker ps
   # If error: Start Docker Desktop
   ```

2. **PostgreSQL or Redis not accessible**
   ```bash
   # Check if containers are running
   docker-compose ps
   
   # Check PostgreSQL
   docker exec agrios-postgres psql -U postgres -c "SELECT 1;"
   
   # Check Redis
   docker exec agrios-redis redis-cli ping
   
   # If containers not running, start them:
   docker-compose up -d postgres redis
   sleep 10
   docker-compose up -d user-service
   ```

3. **Port conflicts (5432, 6379, 50051)**
   ```bash
   # Check if ports are already used
   netstat -ano | findstr :5432   # PostgreSQL
   netstat -ano | findstr :6379   # Redis
   netstat -ano | findstr :50051  # User Service
   
   # On Linux/Mac:
   lsof -i :5432
   lsof -i :6379
   lsof -i :50051
   
   # Solutions:
   # - Stop conflicting services
   # - Or change ports in docker-compose.yml
   ```

4. **Missing .env file**
   ```bash
   # Copy from example
   cp service-1-user/.env.example service-1-user/.env
   ```

5. **Old containers or volumes causing issues**
   ```bash
   # Stop and remove everything
   docker-compose down -v
   
   # Start fresh
   docker-compose up -d
   sleep 15
   ```

---

### Service Won't Start

**Problem:** Service fails to start

**Check logs:**
```bash
# Docker
docker logs agrios-user-service --tail 50

# Or follow logs in real-time
docker-compose logs -f user-service
```

**Common causes:**
1. Database not ready (wait 15 seconds after docker-compose up)
2. Redis not accessible
3. Port 50051 already in use
4. Missing environment variables

**Solutions:**
```bash
# Check all services are healthy
docker-compose ps

# Expected: All services show "Up (healthy)"

# Restart specific service
docker-compose restart user-service

# Full restart
docker-compose down && docker-compose up -d
```

---

### Database Connection Failed

**Problem:** `failed to connect to postgres` in User Service logs

**For Docker Setup (Recommended):**
```bash
# 1. Check PostgreSQL container is running
docker-compose ps postgres
# Should show: Up (healthy)

# 2. Check if database was created
docker exec agrios-postgres psql -U postgres -c "\l" | grep agrios_users

# 3. If database missing, recreate containers
docker-compose down -v
docker-compose up -d

# 4. Check PostgreSQL logs
docker logs agrios-postgres

# 5. Test connection from inside container
docker exec agrios-postgres psql -U postgres -d agrios_users -c "SELECT 1;"

# 6. Verify User Service can connect
docker logs agrios-user-service | grep -i postgres
# Should see: "Connected to PostgreSQL successfully"
```

**For Local Development:**
```bash
# 1. Verify database exists
psql -U postgres -l | grep agrios

# 2. Create database if missing
psql -U postgres -c "CREATE DATABASE agrios_users;"

# 3. Run migrations
psql -U postgres -d agrios_users -f migrations/001_create_users_table.sql

# 4. Check credentials in .env
cat .env | grep DB_

# 5. Test connection
psql -h localhost -U postgres -d agrios_users
```

---

### Redis Connection Failed

**Problem:** `failed to connect to redis` in User Service logs

**For Docker Setup (Recommended):**
```bash
# 1. Check Redis container is running
docker-compose ps redis
# Should show: Up (healthy)

# 2. Test Redis connection from inside container
docker exec agrios-redis redis-cli ping
# Expected: PONG

# 3. Check Redis logs
docker logs agrios-redis

# 4. Verify User Service can connect
docker logs agrios-user-service | grep -i redis
# Should see: "Connected to Redis successfully"

# 5. If issues persist, restart Redis
docker-compose restart redis
sleep 5
docker-compose restart user-service
```

**For Local Development:**
```bash
# 1. Check Redis is running
redis-cli ping

# 2. Start Redis if not running
redis-server

# 3. Check Redis configuration
cat .env | grep REDIS_

# 4. Test connection
redis-cli -h localhost -p 6379 ping
```

---

### Token Validation Fails

**Problem:** ValidateToken returns `valid=false`

**Possible causes:**
1. Token expired
2. Token blacklisted (after logout)
3. JWT_SECRET mismatch
4. Invalid token format

**Solutions:**
```bash
# 1. Check token expiration
# Access tokens expire after 15 minutes

# 2. Use refresh token to get new access token
grpcurl -plaintext \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}" \
  localhost:50051 user.UserService.RefreshToken

# 3. Check if token is blacklisted
redis-cli KEYS "blacklist:*"

# 4. Login again
grpcurl -plaintext \
  -d '{"email":"test@example.com","password":"pass123"}' \
  localhost:50051 user.UserService.Login
```

---

### Email Already Exists

**Problem:** `Email already registered`

**Solutions:**
```bash
# 1. Use different email
# 2. Delete existing user
grpcurl -plaintext -d '{"id":1}' localhost:50051 user.UserService.DeleteUser

# 3. Check existing users
grpcurl -plaintext -d '{"page":1,"page_size":10}' localhost:50051 user.UserService.ListUsers
```

---

## Project Structure

```
service-1-user/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── auth/
│   │   ├── jwt.go               # JWT token generation/validation
│   │   └── password.go          # Password hashing
│   ├── config/
│   │   └── config.go            # Configuration loading
│   ├── db/
│   │   ├── postgres.go          # PostgreSQL connection
│   │   └── redis.go             # Redis connection
│   ├── repository/
│   │   ├── user_repository.go   # Interface
│   │   ├── user_postgres.go     # Implementation
│   │   └── errors.go            # Custom errors
│   ├── response/
│   │   └── grpc_response.go     # Response builders
│   └── server/
│       └── user_server.go       # gRPC server implementation
├── proto/
│   ├── user_service.proto       # gRPC service definition
│   ├── user_service.pb.go       # Generated code
│   └── user_service_grpc.pb.go  # Generated gRPC code
├── migrations/
│   └── 001_create_users_table.sql
├── .env.example                 # Environment template
├── Dockerfile                   # Docker configuration
├── go.mod                       # Go dependencies
└── README.md                    # This file
```

---

## Development Commands

```bash
# Install dependencies
go mod download

# Update dependencies
go mod tidy

# Build
go build -o bin/user-service ./cmd/server

# Run
./bin/user-service

# Run with hot reload (requires air)
go install github.com/cosmtrek/air@latest
air

# Format code
go fmt ./...

# Lint code (requires golangci-lint)
golangci-lint run

# Generate proto files
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/user_service.proto

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# View coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Additional Resources

- **[Main Project README](../README.md)** - Complete platform documentation
- **[Architecture Guide](../docs/ARCHITECTURE_GUIDE.md)** - System design
- **[API Gateway](../service-3-gateway/README.md)** - REST API interface
- **[Article Service](../service-2-article/README.md)** - Content service

---

**Service Version:** 1.0.0  
**Last Updated:** December 5, 2025  
**Maintainer:** thatlq1812@gmail.com
