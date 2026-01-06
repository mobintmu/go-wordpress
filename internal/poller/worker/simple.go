package worker

import (
	"context"
	"log"

	"go-wordpress/internal/poller/job"
)

type SimpleWorker struct {
	id       int
	jobQueue <-chan job.Job
	cancel   context.CancelFunc
}

func NewSimpleWorker(
	id int,
	jobQueue <-chan job.Job,
) *SimpleWorker {
	return &SimpleWorker{
		id:       id,
		jobQueue: jobQueue,
	}
}

func (w *SimpleWorker) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	go func() {
		log.Printf("[worker-%d] started\n", w.id)

		for {
			//blocker
			select {
			case <-ctx.Done():
				log.Printf("[worker-%d] stopped\n", w.id)
				return

			case j := <-w.jobQueue:
				w.handleJob(ctx, j)
			}
		}
	}()
}

func (w *SimpleWorker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
}

func (w *SimpleWorker) handleJob(ctx context.Context, j job.Job) {
	log.Printf(
		"[worker-%d] executing job id=%s service=%s\n",
		w.id,
		j.ID(),
		j.Service(),
	)

	if err := j.Execute(ctx); err != nil {
		log.Printf(
			"[worker-%d] job failed id=%s error=%v\n",
			w.id,
			j.ID(),
			err,
		)
	}
}
