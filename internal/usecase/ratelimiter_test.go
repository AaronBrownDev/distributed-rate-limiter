package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AaronBrownDev/distributed-rate-limiter/internal/storage"
)

type mockStorage struct {
	checkAndUpdateResult *storage.Result
	checkAndUpdateError  error
	getStatusResult      *storage.Result
	getStatusError       error
	resetError           error
}

func (m *mockStorage) CheckAndUpdate(ctx context.Context, key string, limit int64, window time.Duration, cost int64) (*storage.Result, error) {
	return m.checkAndUpdateResult, m.checkAndUpdateError
}

func (m *mockStorage) GetStatus(ctx context.Context, key string, limit int64) (*storage.Result, error) {
	return m.getStatusResult, m.getStatusError
}

func (m *mockStorage) Reset(ctx context.Context, key string) error {
	return m.resetError
}
func (m *mockStorage) Close() error {
	return nil
}

func TestRateLimiter_CheckRateLimit(t *testing.T) {

	tests := []struct {
		name        string
		inputKey    string
		inputLimit  int64
		inputWindow time.Duration
		inputCost   int64
		mockResult  *storage.Result
		mockError   error
		wantErr     error
		wantAllowed bool
	}{
		{
			name:        "valid request",
			inputKey:    "ratelimit:0001",
			inputLimit:  10,
			inputWindow: time.Second,
			inputCost:   1,
			mockResult: &storage.Result{
				Allowed:   true,
				Remaining: 9,
				Limit:     10,
			},
			mockError:   nil,
			wantErr:     nil,
			wantAllowed: true,
		},
		{
			name:        "empty key",
			inputKey:    "",
			inputLimit:  10,
			inputWindow: time.Second,
			inputCost:   1,
			wantErr:     ErrInvalidKey,
		},
		{
			name:        "negative limit",
			inputKey:    "ratelimit:0001",
			inputLimit:  -1,
			inputWindow: time.Second,
			inputCost:   1,
			wantErr:     ErrInvalidLimit,
		},
		{
			name:        "negative window",
			inputKey:    "ratelimit:0001",
			inputLimit:  10,
			inputWindow: time.Second * -1,
			inputCost:   1,
			wantErr:     ErrInvalidWindow,
		},
		{
			name:        "negative cost",
			inputKey:    "ratelimit:0001",
			inputLimit:  10,
			inputWindow: time.Second,
			inputCost:   -1,
			wantErr:     ErrInvalidCost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockStorage{
				checkAndUpdateResult: tt.mockResult,
				checkAndUpdateError:  tt.mockError,
			}

			service := NewRateLimiterService(mock)

			result, err := service.CheckRateLimit(context.Background(), tt.inputKey, tt.inputLimit, tt.inputWindow, tt.inputCost)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("CheckRateLimit() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && result.Allowed != tt.wantAllowed {
				t.Errorf("CheckRateLimit() allowed = %v, wantAllowed %v", result.Allowed, tt.wantAllowed)
			}
		})
	}
}

func TestRateLimiter_GetStatus(t *testing.T) {

	tests := []struct {
		name        string
		inputKey    string
		inputLimit  int64
		mockResult  *storage.Result
		mockError   error
		wantErr     error
		wantAllowed bool
	}{
		{
			name:       "valid",
			inputKey:   "ratelimit:0001",
			inputLimit: 10,
			mockResult: &storage.Result{
				Allowed:   true,
				Remaining: 10,
				Limit:     0,
			},
			mockError:   nil,
			wantErr:     nil,
			wantAllowed: true,
		},
		{
			name:       "empty key",
			inputKey:   "",
			inputLimit: 10,
			wantErr:    ErrInvalidKey,
		},
		{
			name:       "negative limit",
			inputKey:   "ratelimit:0001",
			inputLimit: -1,
			wantErr:    ErrInvalidLimit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockStorage{
				getStatusResult: tt.mockResult,
				getStatusError:  tt.mockError,
			}

			service := NewRateLimiterService(mock)

			result, err := service.GetStatus(context.Background(), tt.inputKey, tt.inputLimit)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GetStatus() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && result.Allowed != tt.wantAllowed {
				t.Errorf("GetStatus() allowed = %v, wantAllowed %v", result.Allowed, tt.wantAllowed)
			}
		})
	}
}

func TestRateLimiter_ResetLimit(t *testing.T) {

	tests := []struct {
		name       string
		inputKey   string
		mockResult *storage.Result
		mockError  error
		wantErr    error
	}{
		{
			name:       "valid",
			inputKey:   "ratelimit:0001",
			mockResult: &storage.Result{},
			mockError:  nil,
			wantErr:    nil,
		},
		{
			name:     "empty key",
			inputKey: "",
			wantErr:  ErrInvalidKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockStorage{
				resetError: tt.mockError,
			}

			service := NewRateLimiterService(mock)

			err := service.ResetLimit(context.Background(), tt.inputKey)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GetStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}
