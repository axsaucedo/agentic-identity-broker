package approval

import (
	"testing"
	"time"
)

func TestApprovalRateLimiter_AllowCreation(t *testing.T) {
	t.Run("allows creation when under rate limit", func(t *testing.T) {
		rl := NewApprovalRateLimiter(50, 10)
		if !rl.AllowCreation("user@example.com", "agent-1") {
			t.Fatal("expected creation to be allowed")
		}
	})

	t.Run("blocks creation when rate limit exceeded", func(t *testing.T) {
		rl := NewApprovalRateLimiter(50, 2)
		for i := 0; i < 2; i++ {
			if !rl.AllowCreation("user@example.com", "agent-1") {
				t.Fatalf("expected creation %d to be allowed", i)
			}
		}
		if rl.AllowCreation("user@example.com", "agent-1") {
			t.Fatal("expected creation to be denied after exhausting rate limit")
		}
	})

	t.Run("different pairs have independent rate limits", func(t *testing.T) {
		rl := NewApprovalRateLimiter(50, 2)
		for i := 0; i < 2; i++ {
			rl.AllowCreation("user@example.com", "agent-1")
		}
		if !rl.AllowCreation("user@example.com", "agent-2") {
			t.Fatal("expected different agent pair to be allowed")
		}
		if !rl.AllowCreation("other@example.com", "agent-1") {
			t.Fatal("expected different principal pair to be allowed")
		}
	})

	t.Run("rate recovers over time", func(t *testing.T) {
		rl := NewApprovalRateLimiter(50, 60)
		for i := 0; i < 60; i++ {
			rl.AllowCreation("user@example.com", "agent-1")
		}
		if rl.AllowCreation("user@example.com", "agent-1") {
			t.Fatal("expected denial after exhausting burst")
		}
		time.Sleep(1100 * time.Millisecond)
		if !rl.AllowCreation("user@example.com", "agent-1") {
			t.Fatal("expected creation to be allowed after token replenishment")
		}
	})
}

func TestApprovalRateLimiter_PairKey(t *testing.T) {
	t.Run("generates consistent keys", func(t *testing.T) {
		rl := NewApprovalRateLimiter(50, 10)
		rl.AllowCreation("user@example.com", "agent-1")
		if !rl.AllowCreation("user@example.com", "agent-1") {
			t.Fatal("expected same pair to reuse limiter")
		}
	})
}
