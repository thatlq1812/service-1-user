#!/bin/bash
# Setup script for User Service

set -e

echo "🚀 Setting up User Service..."

# 1. Check prerequisites
echo "1️⃣ Checking prerequisites..."
command -v go >/dev/null 2>&1 || { echo "❌ Go is not installed"; exit 1; }
command -v psql >/dev/null 2>&1 || { echo "❌ PostgreSQL client is not installed"; exit 1; }
echo "✅ Prerequisites OK"

# 2. Create .env from example if not exists
if [ ! -f .env ]; then
    echo "2️⃣ Creating .env file..."
    cp .env.example .env
    echo "⚠️  IMPORTANT: Edit .env and set your JWT_SECRET!"
    echo "   Generate with: openssl rand -base64 32"
else
    echo "2️⃣ .env file already exists"
fi

# 3. Install Go dependencies
echo "3️⃣ Installing Go dependencies..."
go mod download
echo "✅ Dependencies installed"

# 4. Setup database
echo "4️⃣ Setting up database..."
read -p "   Create database? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    psql -U postgres -c "CREATE DATABASE agrios_users;" 2>/dev/null || echo "   Database already exists"
    psql -U postgres -d agrios_users -f migrations/001_create_users_table.sql
    echo "✅ Database setup complete"
else
    echo "⏭️  Skipping database setup"
fi

# 5. Build service
echo "5️⃣ Building service..."
go build -o bin/user-service cmd/server/main.go
echo "✅ Build complete"

echo ""
echo "🎉 Setup complete!"
echo ""
echo "Next steps:"
echo "  1. Edit .env and set JWT_SECRET: openssl rand -base64 32"
echo "  2. Start service: ./bin/user-service"
echo "  3. Or run directly: go run cmd/server/main.go"
echo ""
