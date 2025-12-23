package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/AaronBrownDev/distributed-rate-limiter/internal/storage"
	"github.com/AaronBrownDev/distributed-rate-limiter/internal/usecase"
)

// CheckRateLimitRequest represents a request to check if a rate limit allows a request.
type CheckRateLimitRequest struct {
	Key           string `json:"key"`
	Limit         int64  `json:"limit"`
	WindowSeconds int64  `json:"window_seconds"`
	Cost          int64  `json:"cost"`
}

// CheckRateLimitResponse contains the result of a rate limit check.
type CheckRateLimitResponse struct {
	Allowed           bool   `json:"allowed"`
	Remaining         int64  `json:"remaining"`
	ResetAt           string `json:"reset_at"`
	Limit             int64  `json:"limit"`
	RetryAfterSeconds int64  `json:"retry_after_seconds"`
}

// GetStatusResponse contains the current status of a rate limit.
type GetStatusResponse struct {
	Allowed   bool   `json:"allowed"`
	Current   int64  `json:"current"`
	Remaining int64  `json:"remaining"`
	ResetAt   string `json:"reset_at"`
	Limit     int64  `json:"limit"`
}

// Handler provides HTTP request handlers for the rate limiter service.
type Handler struct {
	rls *usecase.RateLimiterService
}

// NewHandler creates a new HTTP handler with the provided rate limiter service.
func NewHandler(usecaseService *usecase.RateLimiterService) *Handler {
	return &Handler{
		rls: usecaseService,
	}
}

// CheckRateLimit checks if a request is allowed and consumes tokens if permitted.
func (h *Handler) CheckRateLimit(w http.ResponseWriter, r *http.Request) {
	// Check if method is POST
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Decode request into CheckRateLimitRequest struct
	var req CheckRateLimitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Calculate window
	window := time.Duration(req.WindowSeconds) * time.Second

	// Call service layer
	result, err := h.rls.CheckRateLimit(r.Context(), req.Key, req.Limit, window, req.Cost)
	if err != nil {
		handleServerError(w, err)
		return
	}

	// Calculate retry timer
	var retryAfterSeconds int64
	if !result.Allowed {
		retryAfterSeconds = int64(time.Until(result.ResetAt).Seconds())
	}

	// Build output response
	response := CheckRateLimitResponse{
		Allowed:           result.Allowed,
		Remaining:         result.Remaining,
		Limit:             result.Limit,
		ResetAt:           result.ResetAt.Format(time.RFC3339),
		RetryAfterSeconds: retryAfterSeconds,
	}

	// Send response back

	// Header metadata
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(response.Limit, 10))
	w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(response.Remaining, 10))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))
	if !response.Allowed {
		w.Header().Set("Retry-After", strconv.FormatInt(retryAfterSeconds, 10))
	}

	// Write appropriate status code
	if result.Allowed {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusTooManyRequests)
	}

	// Encode response
	json.NewEncoder(w).Encode(response)
}

// GetStatus retrieves the current rate limit status without consuming tokens.
func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	// Check if method is GET
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Get parameters
	key := r.URL.Query().Get("key")
	limitStr := r.URL.Query().Get("limit")

	// Validate parameters
	if key == "" {
		writeError(w, http.StatusBadRequest, "key parameter required")
		return
	}
	if limitStr == "" {
		writeError(w, http.StatusBadRequest, "limit parameter required")
		return
	}

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "limit not integer")
		return
	}

	// Call service layers
	result, err := h.rls.GetStatus(r.Context(), key, limit)
	if err != nil {
		handleServerError(w, err)
		return
	}

	// Build response
	response := GetStatusResponse{
		Allowed:   result.Allowed,
		Current:   result.Limit - result.Remaining,
		Remaining: result.Remaining,
		ResetAt:   result.ResetAt.Format(time.RFC3339),
		Limit:     result.Limit,
	}

	// Send response back
	// Header metadata
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(response.Limit, 10))
	w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(response.Remaining, 10))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))

	// Returns OK - always success, no consumed tokens
	w.WriteHeader(http.StatusOK)

	// Encode response
	json.NewEncoder(w).Encode(response)
}

// ResetLimit clears the rate limit for the specified key.
func (h *Handler) ResetLimit(w http.ResponseWriter, r *http.Request) {
	// Check if method is DELETE
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Get parameters
	key := r.URL.Query().Get("key")
	if key == "" {
		writeError(w, http.StatusBadRequest, "key parameter required")
		return
	}

	// Call service layer
	if err := h.rls.ResetLimit(r.Context(), key); err != nil {
		handleServerError(w, err)
		return
	}

	// Returns success - no content
	w.WriteHeader(http.StatusNoContent)
}

// RegisterRoutes registers all HTTP handler routes with the provided ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/v1/limit/check", h.CheckRateLimit)
	mux.HandleFunc("/v1/limit/status", h.GetStatus)
	mux.HandleFunc("/v1/limit/reset", h.ResetLimit)
}

// writeError sends a JSON error response with the specified status code and message.
func writeError(w http.ResponseWriter, statusCode int, errorMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": errorMsg})
}

// invalidArgs maps usecase validation errors for quick error type checking.
var invalidArgs = map[error]struct{}{
	usecase.ErrInvalidKey:    {},
	usecase.ErrInvalidLimit:  {},
	usecase.ErrInvalidCost:   {},
	usecase.ErrInvalidWindow: {},
}

// handleServerError converts internal errors to appropriate HTTP status codes.
func handleServerError(w http.ResponseWriter, err error) {
	if _, ok := invalidArgs[err]; ok {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("bad request: %v", err))
	} else if errors.Is(err, storage.ErrKeyNotFound) {
		writeError(w, http.StatusNotFound, "key not found")
	} else {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("internal server error: %v", err))
	}
}