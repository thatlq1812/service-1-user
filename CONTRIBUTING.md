# Contributing to User Service

Thank you for considering contributing to the User Service!

## Development Setup

1. **Fork and clone:**
```bash
git clone https://github.com/YOUR_USERNAME/service-1-user.git
cd service-1-user
```

2. **Setup environment:**
```bash
./setup.sh
```

3. **Create feature branch:**
```bash
git checkout -b feature/your-feature-name
```

## Code Standards

### Go Style
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Run linter: `golangci-lint run`

### Commit Messages
Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add user search endpoint
fix: handle nil pointer in GetUser
docs: update API documentation
chore: upgrade dependencies
test: add tests for JWT validation
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `refactor`: Code restructuring
- `test`: Adding tests
- `chore`: Maintenance tasks

### Response Format
All new endpoints must use wrapped response format:
```go
return &pb.Response{
    Code:    response.MapGRPCCodeToString(codes.OK),
    Message: "success",
    Data:    &pb.ResponseData{...},
}, nil
```

## Testing

### Unit Tests
```bash
go test ./...
```

### Integration Tests
```bash
# Start dependencies
docker-compose up -d postgres redis

# Run tests
go test ./... -tags=integration
```

### Manual Testing
```bash
# Start service
go run cmd/server/main.go

# Test with grpcurl
grpcurl -plaintext -d '{"email":"test@example.com","password":"pass"}' \
  localhost:50051 user.UserService.Login
```

## Pull Request Process

1. **Update documentation** - If adding features, update README.md
2. **Add tests** - Include unit tests for new code
3. **Update CHANGELOG.md** - Document your changes
4. **Clean commits** - Squash WIP commits
5. **Pass CI** - Ensure all checks pass

### PR Template
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
How was this tested?

## Checklist
- [ ] Code follows project style
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
```

## Database Migrations

**Adding new migration:**
1. Create file: `migrations/XXX_description.sql`
2. Test locally: `psql -U postgres -d agrios_users -f migrations/XXX_description.sql`
3. Document in README.md

**Migration guidelines:**
- Always provide rollback steps
- Test on sample data
- Consider backward compatibility

## Proto Changes

**Modifying .proto files:**
1. Edit `proto/user_service.proto`
2. Regenerate code:
```bash
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/user_service.proto
```
3. Update affected handlers
4. Maintain backward compatibility if possible

## Security

**Reporting vulnerabilities:**
- DO NOT open public issues for security issues
- Email: security@example.com (replace with actual)
- Include: description, impact, reproduction steps

**Security checklist:**
- [ ] No hardcoded credentials
- [ ] Input validation on all endpoints
- [ ] SQL injection prevention (use parameterized queries)
- [ ] JWT secrets not committed
- [ ] Error messages don't leak sensitive info

## Code Review

Reviewers will check:
- Code quality and style
- Test coverage
- Documentation completeness
- Security considerations
- Performance implications

## Questions?

- Open an issue for bugs/features
- Discussions for questions
- Check existing issues first

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
