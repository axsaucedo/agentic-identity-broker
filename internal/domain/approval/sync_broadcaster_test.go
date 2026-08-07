package approval

import (
	"testing"
	"time"
)

func TestApprovalSyncBroadcaster_SubscribeAndBroadcast(t *testing.T) {
	t.Run("subscriber receives broadcast", func(t *testing.T) {
		b := NewApprovalSyncBroadcaster(0) // no coalesce for this test
		ch := b.Subscribe()
		defer b.Unsubscribe(ch)

		b.Broadcast()

		select {
		case <-ch:
			// success
		case <-time.After(100 * time.Millisecond):
			t.Fatal("expected to receive broadcast")
		}
	})

	t.Run("multiple subscribers all receive broadcast", func(t *testing.T) {
		b := NewApprovalSyncBroadcaster(0)
		ch1 := b.Subscribe()
		ch2 := b.Subscribe()
		ch3 := b.Subscribe()
		defer b.Unsubscribe(ch1)
		defer b.Unsubscribe(ch2)
		defer b.Unsubscribe(ch3)

		b.Broadcast()

		for i, ch := range []<-chan struct{}{ch1, ch2, ch3} {
			select {
			case <-ch:
				// success
			case <-time.After(100 * time.Millisecond):
				t.Fatalf("subscriber %d did not receive broadcast", i)
			}
		}
	})

	t.Run("unsubscribed channel is closed and does not receive new broadcasts", func(t *testing.T) {
		b := NewApprovalSyncBroadcaster(0)
		ch := b.Subscribe()
		b.Unsubscribe(ch)

		// After unsubscribe, the channel is closed. Reads return immediately with zero value.
		// Verify the channel is indeed closed (this is the expected behavior).
		_, open := <-ch
		if open {
			t.Fatal("expected channel to be closed after Unsubscribe")
		}

		// Verify broadcasting doesn't panic when the channel has been unsubscribed and closed.
		b.Broadcast()
	})
}

func TestApprovalSyncBroadcaster_Coalesce(t *testing.T) {
	t.Run("coalesces rapid broadcasts within window", func(t *testing.T) {
		b := NewApprovalSyncBroadcaster(200 * time.Millisecond)
		ch := b.Subscribe()
		defer b.Unsubscribe(ch)

		// Fire multiple rapid broadcasts
		b.Broadcast()
		b.Broadcast()
		b.Broadcast()

		// Drain single coalesced notification
		select {
		case <-ch:
			// success — got coalesced notification
		case <-time.After(500 * time.Millisecond):
			t.Fatal("expected coalesced broadcast within window")
		}

		// Should NOT receive additional notifications (they were coalesced)
		select {
		case <-ch:
			t.Fatal("expected broadcasts to be coalesced into one")
		case <-time.After(100 * time.Millisecond):
			// success — no second notification
		}
	})

	t.Run("separate broadcasts outside coalesce window are independent", func(t *testing.T) {
		b := NewApprovalSyncBroadcaster(50 * time.Millisecond)
		ch := b.Subscribe()
		defer b.Unsubscribe(ch)

		b.Broadcast()

		select {
		case <-ch:
			// first broadcast received
		case <-time.After(200 * time.Millisecond):
			t.Fatal("expected first broadcast")
		}

		// Wait beyond coalesce window
		time.Sleep(100 * time.Millisecond)

		b.Broadcast()

		select {
		case <-ch:
			// second broadcast received
		case <-time.After(200 * time.Millisecond):
			t.Fatal("expected second broadcast after coalesce window expired")
		}
	})
}
