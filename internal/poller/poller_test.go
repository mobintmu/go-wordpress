package poller_test

import (
	"context"
	"go-wordpress/internal/example"
	"go-wordpress/internal/poller"
	"go-wordpress/internal/poller/dispatcher"
	"log"
	"sync/atomic"
	"testing"
	"time"
)

// TestPollerIntegration_FullLifecycle tests the complete poller -> dispatcher -> worker flow
func TestPollerIntegration_FullLifecycle(t *testing.T) {
	// Setup
	disp := dispatcher.New()
	disp.Register("email", 3, 100)
	disp.Start()

	pollerInstance := poller.New(disp)
	pollerInstance.Interval = 100 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run poller
	done := make(chan struct{})
	go func() {
		pollerInstance.Run(ctx)
		close(done)
	}()

	// Let poller dispatch jobs
	time.Sleep(400 * time.Millisecond)

	// Cleanup
	cancel()
	disp.Stop()

	select {
	case <-done:
		log.Println("Poller stopped gracefully")
	case <-time.After(5 * time.Second):
		t.Fatal("poller did not stop within timeout")
	}

	t.Log("✓ Full lifecycle completed successfully")
}

func TestPollerIntegration_WithConditionalDispatching(t *testing.T) {
	disp := dispatcher.New()
	disp.Register("email", 2, 50)
	disp.Start()
	defer disp.Stop()

	var tickCount atomic.Int32

	p := poller.New(disp)
	p.Interval = 50 * time.Millisecond

	// Inject conditional behavior
	p.OnTick = func(ctx context.Context) {
		count := tickCount.Add(1)
		if count%2 == 0 {
			j := example.NewExampleJob("job-123", "email")
			_ = disp.Dispatch(j)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	go p.Run(ctx)

	time.Sleep(2 * time.Second)
	cancel()

	count := tickCount.Load()
	t.Logf("Total ticks: %d", count)

	if count < 3 {
		t.Fatalf("expected at least 3 ticks, got %d", count)
	}
}
