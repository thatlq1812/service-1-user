# User Service

User authentication and management microservice for Agrios platform.

## Features

- User registration and authentication
- JWT token management (access + refresh tokens)
- Token blacklisting with Redis
- Password hashing with bcrypt
- gRPC API with wrapped response format

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Redis 7+

### Setup

1. **Clone and configure:**
```bash
git clone <repo-url>
cd service-1-user
cp .env.example .env
# Edit .env with your configuration
```

2. **Install dependencies:**
```bash
go mod download
```

3. **Setup database:**
```bash
# Create database
psql -U postgres -c "CREATE DATABASE agrios_users;"

# Run migrations
psql -U postgres -d agrios_users -f migrations/001_create_users_table.sql
```

4. **Run service:**
```bash
# Development
go run cmd/server/main.go

# Build and run
go build -o bin/user-service cmd/server/main.go
./bin/user-service
```

Service will start on port **50051** (gRPC).

### Docker

```bash
# Build
docker build -t user-service .

# Run
docker run -p 50051:50051 --env-file .env user-service
```

## API Documentation

### gRPC Methods

1. **CreateUser** - Register new user
2. **Login** - Authenticate and get tokens
3. **GetUser** - Get user by ID
4. **UpdateUser** - Update user information
5. **DeleteUser** - Delete user
6. **ListUsers** - Paginated user list
7. **ValidateToken** - Verify JWT token
8. **RevokeToken** - Logout (blacklist token)

### Response Format

All responses use wrapped format:
```json
{
  "code": "000",
  "message": "success",
  "data": {...}
}
```

**Error codes:**
- `000` - Success
- `003` - Invalid argument
- `005` - Not found
- `006` - Already exists
- `013` - Internal error
- `014` - Unauthenticated

### Testing

```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Create user
grpcurl -plaintext -d '{"name":"Test","email":"test@example.com","password":"pass123"}' localhost:50051 user.UserService.CreateUser

# Login
grpcurl -plaintext -d '{"email":"test@example.com","password":"pass123"}' localhost:50051 user.UserService.Login
```

## Configuration

Key environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | localhost | PostgreSQL host |
| `DB_PORT` | 5432 | PostgreSQL port |
| `DB_USER` | postgres | Database user |
| `DB_PASSWORD` | postgres | Database password |
| `DB_NAME` | agrios_users | Database name |
| `REDIS_ADDR` | localhost:6379 | Redis address |
| `JWT_SECRET` | *required* | JWT signing secret (change in production!) |
| `JWT_ACCESS_TOKEN_DURATION` | 15m | Access token lifetime |
| `JWT_REFRESH_TOKEN_DURATION` | 168h | Refresh token lifetime (7 days) |
| `GRPC_PORT` | 50051 | gRPC server port |

## Project Structure

```
service-1-user/
├── cmd/server/main.go          # Entry point
├── internal/
│   ├── auth/                   # JWT and password handling
│   ├── config/                 # Configuration loading
│   ├── db/                     # Database connections
│   ├── repository/             # Data access layer
│   ├── response/               # Response helpers
│   └── server/                 # gRPC server implementation
├── proto/                      # Protocol buffer definitions
├── migrations/                 # Database migrations
├── Dockerfile                  # Container image
└── .env.example               # Environment template

```

## Development

### Generate Proto Files

```bash
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/user_service.proto
```

### Database Migrations

Migrations are in `migrations/` directory. Apply manually:

```bash
psql -U postgres -d agrios_users -f migrations/001_create_users_table.sql
```

## Security Notes

**Important:**
- Change `JWT_SECRET` in production (use `openssl rand -base64 32`)
- Use strong database passwords
- Enable SSL/TLS for PostgreSQL and Redis in production
- Never commit `.env` file to git

## Troubleshooting

**Connection refused:**
- Check PostgreSQL is running: `pg_isready`
- Check Redis is running: `redis-cli ping`

**Database does not exist:**
```bash
psql -U postgres -c "CREATE DATABASE agrios_users;"
```

**Port already in use:**
```bash
# Find process using port 50051
netstat -ano | findstr :50051
# Kill or change GRPC_PORT in .env
```

## License

MIT License
