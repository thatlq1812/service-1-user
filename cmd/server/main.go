package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"service-1-user/internal/db"
	"service-1-user/internal/repository"
	"service-1-user/internal/server"
	pb "service-1-user/proto"
)

func main() {
	// 1. Setup database
	dbConfig := db.Config{
		Host:     "127.0.0.1",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		DBName:   "agrios_users",
	}

	pool, err := db.NewPostgresPool(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("Connected to PostgreSQL")

	// 2. Create repository
	userRepo := repository.NewUserPostgresRepository(pool)

	// 3. Setup gRPC server
	grpcServer := grpc.NewServer()

	// 4. Register service
	userService := server.NewUserServiceServer(userRepo)
	pb.RegisterUserServiceServer(grpcServer, userService)

	// 5. Enable reflection (for grpcurl)
	reflection.Register(grpcServer)

	// 6. Listen on port 50051
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen on port 50051: %v", err)
	}

	log.Println("gRPC server listening on :50051")

	// 7. Start serving (blocking call)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
