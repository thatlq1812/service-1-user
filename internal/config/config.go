package config

import (
	"agrios/pkg/common"
	"service-1-user/internal/db"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	GRPCPort        string
	ShutdownTimeout time.Duration

	// JWT
	JWTSecret            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration

	DB db.Config
}

func Load() *Config {
	return &Config{
		// Server Config
		GRPCPort:        common.GetEnvString("GRPC_PORT", "50051"),
		ShutdownTimeout: common.GetEnvDuration("SHUTDOWN_TIMEOUT", 10*time.Second),

		// JWT Config
		JWTSecret:            common.GetEnvString("JWT_SECRET", "insecure-default-secret-change-this"), // default value for Dev
		AccessTokenDuration:  common.GetEnvDuration("ACCESS_TOKEN_DURATION", 24*time.Hour),
		RefreshTokenDuration: common.GetEnvDuration("REFRESH_TOKEN_DURATION", 7*24*time.Hour),

		// Database Config
		DB: db.Config{
			Host:     common.GetEnvString("DB_HOST", "localhost"),
			Port:     common.GetEnvString("DB_PORT", "5432"),
			User:     common.MustGetEnvString("DB_USER"),
			Password: common.MustGetEnvString("DB_PASSWORD"),
			DBName:   common.MustGetEnvString("DB_NAME"),

			MaxConns:        common.GetEnvInt32("DB_MAX_CONNS", 10),
			MinConns:        common.GetEnvInt32("DB_MIN_CONNS", 2),
			MaxConnLifetime: common.GetEnvDuration("DB_MAX_CONN_LIFETIME", time.Hour),
			MaxConnIdleTime: common.GetEnvDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
			ConnectTimeout:  common.GetEnvDuration("DB_CONNECT_TIMEOUT", 5*time.Second),
		},
	}
}
