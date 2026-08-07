package approval

import (
	"sync"

	"golang.org/x/time/rate"
)

// ApprovalRateLimiter enforces per-(principal, agent_id) rate limits on approval creation.
// It uses an in-memory token bucket per pair via golang.org/x/time/rate.
type ApprovalRateLimiter struct {
	mu         sync.Mutex
	limiters   map[string]*rate.Limiter
	maxPending int
	ratePerMin int
}

// NewApprovalRateLimiter creates a new rate limiter with the specified constraints.
func NewApprovalRateLimiter(maxPending, ratePerMin int) *ApprovalRateLimiter {
	return &ApprovalRateLimiter{
		limiters:   make(map[string]*rate.Limiter),
		maxPending: maxPending,
		ratePerMin: ratePerMin,
	}
}

// AllowCreation checks if a new approval creation is allowed for the given pair
// based on the per-minute rate limit. Returns true if allowed, false if rate limited.
func (rl *ApprovalRateLimiter) AllowCreation(principal, agentID string) bool {
	key := principal + "|" + agentID

	rl.mu.Lock()
	limiter, ok := rl.limiters[key]
	if !ok {
		limiter = rate.NewLimiter(rate.Limit(float64(rl.ratePerMin)/60.0), rl.ratePerMin)
		rl.limiters[key] = limiter
	}
	rl.mu.Unlock()

	return limiter.Allow()
}

// MaxPending returns the configured maximum pending approvals per pair.
func (rl *ApprovalRateLimiter) MaxPending() int {
	return rl.maxPending
}
