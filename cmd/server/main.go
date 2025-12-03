package main

import (
	"log"
	"net"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"agrios/pkg/common"
	"service-1-user/internal/auth"
	"service-1-user/internal/config"
	"service-1-user/internal/db"
	"service-1-user/internal/repository"
	"service-1-user/internal/server"
	pb "service-1-user/proto"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// 1. Load configuration
	cfg := config.Load()

	// 2. Setup database connection pool
	pool, err := db.NewPostgresPool(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("Connected to PostgreSQL successfully")

	// 3. Create repository
	userRepo := repository.NewUserPostgresRepository(pool)

	tokenManager := auth.NewTokenManager(
		cfg.JWTSecret,
		cfg.AccessTokenDuration,
		cfg.RefreshTokenDuration,
	)

	// 4. Setup gRPC server
	grpcServer := grpc.NewServer()

	// 5. Register service implementation
	userService := server.NewUserServiceServer(userRepo, tokenManager)
	pb.RegisterUserServiceServer(grpcServer, userService)

	// 6. Enable reflection for tools like grpcurl
	reflection.Register(grpcServer)

	// 7. Setup TCP listener
	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.GRPCPort, err)
	}

	log.Printf("User Service (gRPC) listening on port %s", cfg.GRPCPort)

	// 8. Start server in goroutine to handle graceful shutdown
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// 9. Wait for shutdown signal and perform graceful shutdown
	ctx := common.WaitForShutdown(cfg.ShutdownTimeout)

	log.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()

	<-ctx.Done()
	log.Println("Server stopped gracefully")
}
