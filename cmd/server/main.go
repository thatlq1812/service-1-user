package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"service-1-user/internal/db"
	"service-1-user/internal/repository"
	"service-1-user/internal/server"
	pb "service-1-user/proto"
)

func mustGetEnvInt32(key string, defaultValue int32) int32 {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.ParseInt(valStr, 10, 32)
	if err != nil {
		log.Printf("W: could not parse %s='%s' to int32. Using default value %d.", key, valStr, defaultValue)
		return defaultValue
	}
	return int32(val)
}

func mustGetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := time.ParseDuration(valStr)
	if err != nil {
		log.Fatalf("Error: Could not parse %s='%s' to time.Duration. Example format: 1h, 30m, 5s.", key, valStr)
		return defaultValue
	}
	return val
}

func LoadDBConfig() db.Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found.")
	}

	return db.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),

		MaxConns:        mustGetEnvInt32("DB_MAX_CONNS", 10),
		MinConns:        mustGetEnvInt32("DB_MIN_CONNS", 2),
		MaxConnLifetime: mustGetEnvDuration("DB_MAX_CONN_LIFETIME", time.Hour),
		MaxConnIdleTime: mustGetEnvDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
		ConnectTimeout:  mustGetEnvDuration("DB_CONNECT_TIMEOUT", 5*time.Second),
	}
}

func main() {

	// Load env variables from .env
	dbConfig := LoadDBConfig()

	// 1. Setup database

	if dbConfig.User == "" || dbConfig.DBName == "" {
		log.Fatalf("Missing required database configuration (DB_USER or DB_NAME) in environment.")
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
