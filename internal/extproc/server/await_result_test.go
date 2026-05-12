package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/singleflight"
)

// TestAwaitResult_CancellationWinsDeterministically verifies that awaitResult always
// returns ctx.Err() when both ctx.Done() and resChan are ready simultaneously.
//
// Both channels are populated before awaitResult runs, guaranteeing the simultaneous-
// ready case on every iteration. Without the ctx.Err() check in the resChan branch,
// Go's random select would return the token on roughly half of the 1000 iterations;
// with it, all iterations must return context.Canceled.
func TestAwaitResult_CancellationWinsDeterministically(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	for range 1000 {
		resChan := make(chan singleflight.Result, 1)
		resChan <- singleflight.Result{Val: "race-token"}

		_, err := awaitResult(ctx, resChan)
		require.ErrorIs(t, err, context.Canceled,
			"canceled context must return context error, not token")
	}
}
