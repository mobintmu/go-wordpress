package poller

import (
	"context"
	"fmt"
	"go-wordpress/internal/example"
	"go-wordpress/internal/poller/dispatcher"
	"log"
	"time"

	"go.uber.org/fx"
)

/*
Fx App
 ├── Dispatcher (long-lived)
 │    └── Worker pools
 ├── Poller (long-lived background process)
 │    └── Infinite loop
 │         ├── Check condition (DB / cache / API)
 │         └── Dispatch job(s)
 └── Servers
*/

type Poller struct {
	Dispatcher *dispatcher.Service
	Interval   time.Duration
	ctx        context.Context
	cancel     context.CancelFunc
	OnTick     func(context.Context)
}

func New(
	dispatcher *dispatcher.Service,
) *Poller {
	p := &Poller{
		Dispatcher: dispatcher,
		Interval:   500 * time.Microsecond,
	}
	p.OnTick = p.DefaultTick // default behavior
	return p
}

func (p *Poller) Run(ctx context.Context) {
	ticker := time.NewTicker(p.Interval)
	defer ticker.Stop()

	for {
		fmt.Println("poller For loop ...")
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			p.OnTick(ctx)
		}
	}
}

func (p *Poller) DefaultTick(ctx context.Context) {
	// Example condition (replace with DB/cache logic)
	log.Println("DefaultTick is running ...")
	if !p.conditionMet(ctx) {
		return
	}

	j := example.NewExampleJob("job-123", "email")

	_ = p.Dispatcher.Dispatch(j)
}

func (p *Poller) conditionMet(ctx context.Context) bool {
	// DB / Redis / API / feature flag / backlog size
	return true
}

func RegisterLifecycle(lc fx.Lifecycle, p *Poller) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			fmt.Println("starting dispatcher...")
			p.ctx, p.cancel = context.WithCancel(context.Background())
			go p.Run(p.ctx)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			fmt.Println("stopping poller...")
			if p.cancel != nil {
				p.cancel()
			}
			return nil
		},
	})
}
