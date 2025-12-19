package grpc

import (
	"context"
	"errors"
	"time"

	pb "github.com/AaronBrownDev/distributed-rate-limiter/gen/ratelimiter/v1"
	"github.com/AaronBrownDev/distributed-rate-limiter/internal/storage"
	"github.com/AaronBrownDev/distributed-rate-limiter/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"	
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server implements the gRPC RateLimiterService server.
type Server struct {
	rls *usecase.RateLimiterService
	pb.UnimplementedRateLimiterServiceServer
}

// NewServer creates a new gRPC server with the provided rate limiter service.
func NewServer(usecaseServer *usecase.RateLimiterService) *Server {
	return &Server{
		rls: usecaseServer,
	}
}

// CheckRateLimit checks if a request is allowed and consumes tokens if permitted.
func (s *Server) CheckRateLimit(ctx context.Context, req *pb.CheckRateLimitRequest) (*pb.CheckRateLimitResponse, error) {
	var window time.Duration = time.Duration(req.WindowSeconds) * time.Second

	result, err := s.rls.CheckRateLimit(ctx, req.Key, req.Limit, window, req.Cost)
	if err != nil {
		return nil, handleError(err)
	}

	var retryAfterSeconds int64
	if !result.Allowed {
		retryAfterSeconds = int64(time.Until(result.ResetAt).Seconds())
	}

	return &pb.CheckRateLimitResponse{
		Allowed:           result.Allowed,
		Remaining:         result.Remaining,
		ResetAt:           timestamppb.New(result.ResetAt),
		Limit:             result.Limit,
		RetryAfterSeconds: retryAfterSeconds,
	}, nil
}

// GetStatus retrieves the current rate limit status without consuming tokens.
func (s *Server) GetStatus(ctx context.Context, req *pb.GetStatusRequest) (*pb.GetStatusResponse, error) {

	result, err := s.rls.GetStatus(ctx, req.Key, req.Limit)
	if err != nil {
		return nil, handleError(err)
	}

	return &pb.GetStatusResponse{
		Allowed:   result.Allowed,
		Current:   result.Limit - result.Remaining,
		Remaining: result.Remaining,
		ResetAt:   timestamppb.New(result.ResetAt),
		Limit:     result.Limit,
	}, nil
}


// ResetLimit clears the rate limit for the specified key.
func (s *Server) ResetLimit(ctx context.Context, req *pb.ResetLimitRequest) (*pb.ResetLimitResponse, error) {

	err := s.rls.ResetLimit(ctx, req.Key)
	if err != nil {
		return nil, handleError(err)
	}

	return &pb.ResetLimitResponse{}, nil
}

// Organizes invalid argument errors into a hashset for handleError func
var invalidArgs = map[error]struct{}{
	usecase.ErrInvalidKey:    {},
	usecase.ErrInvalidLimit:  {},
	usecase.ErrInvalidCost:   {},
	usecase.ErrInvalidWindow: {},
}

// handleError is a helper function for matching the error to its appropriate gRPC error status
func handleError(err error) error {
	if _, ok := invalidArgs[err]; ok {
		return status.Errorf(codes.InvalidArgument, "invalid argument: %v", err)
	} else if errors.Is(err, storage.ErrKeyNotFound) {
		return status.Errorf(codes.NotFound, "key not found")
	} else {
		return status.Errorf(codes.Internal, "internal server error: %v", err)
	}
}
