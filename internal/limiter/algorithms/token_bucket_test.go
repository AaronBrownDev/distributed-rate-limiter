package algorithms

import (
	"testing"
	"time"
)

func TestTokenBucket_NewTokenBucket_Capacity(t *testing.T) {
	var testCapacity int64 = 5
	tb := NewTokenBucket(testCapacity, 1, 100*time.Millisecond)

	if tb.getCurrentTokens() != testCapacity {
		t.Errorf("expected tokens to be at max capacity, got %d instead of %d", tb.getCurrentTokens(), testCapacity)
	}
}

func TestTokenBucket_Allow(t *testing.T) {
	type allowCall struct {
		requestTokens int64
		waitTime      time.Duration
		wantedBool    bool
		wantedTokens  int64
	}

	tests := []struct {
		name         string
		capacity     int64
		refillRate   int64
		refillPeriod time.Duration
		calls        []allowCall
	}{
		{
			name:         "Allow within initial capacity",
			capacity:     5,
			refillRate:   1,
			refillPeriod: 100 * time.Millisecond,
			calls: []allowCall{
				{
					requestTokens: 3,
					waitTime:      0,
					wantedBool:    true,
					wantedTokens:  2,
				},
			},
		},
		{
			name:         "Reject when insufficient tokens",
			capacity:     5,
			refillRate:   1,
			refillPeriod: 100 * time.Millisecond,
			calls: []allowCall{
				// Call 1
				{
					requestTokens: 5,
					waitTime:      0,
					wantedBool:    true,
					wantedTokens:  0,
				},
				// Call 2
				{
					requestTokens: 1,
					waitTime:      0,
					wantedBool:    false,
					wantedTokens:  0,
				},
			},
		},
		{
			name:         "Refill after time",
			capacity:     5,
			refillRate:   1,
			refillPeriod: 100 * time.Millisecond,
			calls: []allowCall{
				{
					requestTokens: 5,
					waitTime:      0,
					wantedBool:    true,
					wantedTokens:  0,
				},
				{
					requestTokens: 1,
					waitTime:      0,
					wantedBool:    false,
					wantedTokens:  0,
				},
				{
					requestTokens: 1,
					waitTime:      101 * time.Millisecond,
					wantedBool:    true,
					wantedTokens:  0,
				},
			},
		},
		{
			name:         "Multiple refill periods",
			capacity:     5,
			refillRate:   2,
			refillPeriod: 100 * time.Millisecond,
			calls: []allowCall{
				{
					requestTokens: 5,
					waitTime:      0,
					wantedBool:    true,
					wantedTokens:  0,
				},
				{
					requestTokens: 4,
					waitTime:      201 * time.Millisecond,
					wantedBool:    true,
					wantedTokens:  0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tb := NewTokenBucket(tt.capacity, tt.refillRate, tt.refillPeriod)

			for _, call := range tt.calls {

				if call.waitTime > 0 {
					time.Sleep(call.waitTime)
				}

				success, tokens := tb.Allow(call.requestTokens)

				if success != call.wantedBool {
					t.Errorf("got success=%v, want %v", success, call.wantedBool)
				}

				if tokens != call.wantedTokens {
					t.Errorf("got tokens=%d, want %d", tokens, call.wantedTokens)
				}
			}
		})
	}
}
