// internal/dispatcher/dispatcher.go
package dispatcher

//Producer → Dispatcher → Service Queue → Worker Pool → Job.Execute()
//One queue per service
//N workers per service
import (
	"context"
	"errors"
	"fmt"
	"go-wordpress/internal/poller/job"
	"go-wordpress/internal/poller/worker"
	"sync"

	"go.uber.org/fx"
)

type Dispatcher interface {
	Start()
	Stop()
	Dispatch(j job.Job) error
}

var ErrServiceNotRegistered = errors.New("service not registered")

type Service struct {
	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex

	queues  map[string]chan job.Job
	workers map[string][]worker.Worker
}

func New() *Service {
	ctx, cancel := context.WithCancel(context.Background())

	return &Service{
		ctx:     ctx,
		cancel:  cancel,
		queues:  make(map[string]chan job.Job),
		workers: make(map[string][]worker.Worker),
	}
}

func (d *Service) Register(
	service string,
	workerCount int,
	queueSize int,
) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.queues[service]; exists {
		fmt.Println("service already registered")
		return
	}

	queue := make(chan job.Job, queueSize)
	d.queues[service] = queue

	var workers []worker.Worker
	for i := 0; i < workerCount; i++ {
		w := worker.NewSimpleWorker(i+1, queue)
		workers = append(workers, w)
	}

	d.workers[service] = workers
}

func (d *Service) Start() {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, ws := range d.workers {
		for _, w := range ws {
			w.Start(d.ctx)
		}
	}
}

func (d *Service) Dispatch(j job.Job) error {
	// Prevent dispatch after stop
	select {
	case <-d.ctx.Done():
		return context.Canceled
	default:
	}

	d.mu.RLock()
	queue, ok := d.queues[j.Service()]
	d.mu.RUnlock()

	if !ok {
		return ErrServiceNotRegistered
	}

	// Safe send
	select {
	case queue <- j:
		return nil
	case <-d.ctx.Done():
		return context.Canceled
	}
}

func (d *Service) Stop() {
	d.cancel()

	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, q := range d.queues {
		close(q)
	}
}

func RegisterLifecycle(lc fx.Lifecycle, d *Service) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			fmt.Println("starting dispatcher...")
			d.Start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			fmt.Println("stopping dispatcher...")
			d.Stop()
			return nil
		},
	})
}

func RegisterServices(d *Service) {
	// Example services
	d.Register("email", 5, 100)
}
