package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	pb "github.com/AaronBrownDev/distributed-rate-limiter/gen/ratelimiter/v1"
	grpcDelivery "github.com/AaronBrownDev/distributed-rate-limiter/internal/delivery/grpc"
	httpDelivery "github.com/AaronBrownDev/distributed-rate-limiter/internal/delivery/http"
	"github.com/AaronBrownDev/distributed-rate-limiter/internal/storage/redis"
	"github.com/AaronBrownDev/distributed-rate-limiter/internal/usecase"
	"google.golang.org/grpc"
)

func main() {
	exitCode := 0
	defer func() { os.Exit(exitCode) }()

	// Create context for shutdown signaling
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Get Redis configuration from environment
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	// Build Redis connection address
	redisAddress := fmt.Sprintf("%s:%s", redisHost, redisPort)
	keyPrefix := "ratelimit:"

	// Initialize Redis storage
	redisStorage, err := redis.NewRedisStorage(ctx, redisAddress, keyPrefix)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer func() {
		log.Println("Redis connection closed...")
		redisStorage.Close()
	}()
	// Create rate limiter service
	rateLimitService := usecase.NewRateLimiterService(redisStorage)

	// Get gRPC port from environment or use default
	grpcPort := 50051
	if envPort := os.Getenv("GRPC_PORT"); envPort != "" {
		grpcPort, err = strconv.Atoi(envPort)
		if err != nil {
			log.Printf("Invalid GRPC_PORT: %v", err)
			exitCode = 1
			return
		}
	}

	// Start gRPC server
	grpcServer, err := startGRPCServer(rateLimitService, grpcPort)
	if err != nil {
		log.Printf("Failed to start gRPC server: %v", err)
		exitCode = 1
		return
	}
	defer func() {
		log.Println("Shutting down gRPC server...")
		grpcServer.GracefulStop()
	}()
	// Get HTTP port from environment or use default
	apiPort := 8080
	if envPort := os.Getenv("API_PORT"); envPort != "" {
		apiPort, err = strconv.Atoi(envPort)
		if err != nil {
			log.Printf("Invalid API_PORT: %v", err)
			exitCode = 1
			return
		}
	}

	// Start HTTP server
	httpServer, err := startAPIServer(rateLimitService, apiPort)
	if err != nil {
		log.Printf("Failed to start HTTP server: %v", err)
		exitCode = 1
		return
	}
	defer func() {
		log.Println("Shutting down HTTP server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("Shutdown signal received")
}

// startAPIServer creates and starts the HTTP server.
func startAPIServer(rateLimitService *usecase.RateLimiterService, port int) (*http.Server, error) {
	handler := httpDelivery.NewHandler(rateLimitService)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Handler: mux,
	}

	// Create listener to detect port binding errors early
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	// Start server in goroutine
	go func() {
		log.Printf("HTTP server listening on :%d", port)
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	return server, nil
}

// startGRPCServer creates and starts the gRPC server.
func startGRPCServer(rateLimitService *usecase.RateLimiterService, port int) (*grpc.Server, error) {
	grpcServer := grpc.NewServer()
	pb.RegisterRateLimiterServiceServer(grpcServer, grpcDelivery.NewServer(rateLimitService))

	// Create listener to detect port binding errors early
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	// Start server in goroutine
	go func() {
		log.Printf("gRPC server listening on :%d", port)
		if err := grpcServer.Serve(listener); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	return grpcServer, nil
}