package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAntigravityClaudeTransientResourceError(t *testing.T) {
	tests := []struct {
		name   string
		model  string
		status int
		body   string
		want   bool
	}{
		{
			name:   "model capacity exhausted",
			model:  "claude-sonnet-4-6",
			status: 503,
			body:   `{"error":{"status":"UNAVAILABLE","message":"No capacity available","details":[{"reason":"MODEL_CAPACITY_EXHAUSTED"}]}}`,
			want:   true,
		},
		{
			name:   "overloaded 529",
			model:  "claude-sonnet-4-6",
			status: 529,
			body:   `{"error":{"message":"upstream overloaded"}}`,
			want:   true,
		},
		{
			name:   "quota exhausted",
			model:  "claude-sonnet-4-6",
			status: 429,
			body:   `{"error":{"status":"RESOURCE_EXHAUSTED","message":"QUOTA_EXHAUSTED"}}`,
			want:   false,
		},
		{
			name:   "rate limit exceeded",
			model:  "claude-sonnet-4-6",
			status: 429,
			body:   `{"error":{"status":"RESOURCE_EXHAUSTED","message":"RATE_LIMIT_EXCEEDED"}}`,
			want:   false,
		},
		{
			name:   "validation required",
			model:  "claude-sonnet-4-6",
			status: 403,
			body:   `{"error":{"status":"PERMISSION_DENIED","reason":"VALIDATION_REQUIRED"}}`,
			want:   false,
		},
		{
			name:   "non Claude model unchanged",
			model:  "gemini-3.7-flash",
			status: 503,
			body:   `{"error":{"message":"No capacity available"}}`,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, isAntigravityClaudeTransientResourceError(tt.model, tt.status, []byte(tt.body)))
		})
	}
}

func TestNewAntigravityTransientStreamFailoverErrorClaudeBudget(t *testing.T) {
	err := newAntigravityTransientStreamFailoverError("claude-opus-4-6-thinking")
	require.True(t, err.RetryableOnSameAccount)
	require.True(t, err.RequestScopedTransient)
	require.True(t, err.SafeToFailoverAfterWrite)
	require.Equal(t, antigravityClaudeTransientRetryMaxRetries, err.SameAccountRetryMax)
	require.WithinDuration(t, time.Now().Add(antigravityClaudeTransientRetryWindow), err.SameAccountRetryDeadline, time.Second)
}

func TestNewAntigravityTransientStreamFailoverErrorGeminiKeepsLegacyBudget(t *testing.T) {
	err := newAntigravityTransientStreamFailoverError("gemini-3.7-flash")
	require.True(t, err.RetryableOnSameAccount)
	require.False(t, err.RequestScopedTransient)
	require.Zero(t, err.SameAccountRetryMax)
	require.True(t, err.SameAccountRetryDeadline.IsZero())
}
