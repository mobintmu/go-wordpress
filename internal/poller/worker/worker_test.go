package worker

import (
	"context"
	"errors"
	"fmt"
	"go-wordpress/internal/poller/job"
	"sync"
	"testing"
	"time"
)

type mockJob struct {
	id      string
	service string
	timeout int

	mu       sync.Mutex
	executed bool
	execErr  error
	done     chan struct{}
}

func newMockJob(id int, timeout int) *mockJob {
	fmt.Println("creating new mock job id ", id)
	return &mockJob{
		id:      fmt.Sprintf("job-%d", id),
		service: "test-service",
		done:    make(chan struct{}),
		timeout: timeout,
	}
}

func (m *mockJob) ID() string {
	return m.id
}

func (m *mockJob) Service() string {
	return m.service
}

func (m *mockJob) Execute(ctx context.Context) error {
	fmt.Println("wait for 3 seconds before executing job m : ", m.id)
	time.Sleep(3 * time.Second)
	m.mu.Lock()
	m.executed = true
	m.mu.Unlock()

	close(m.done)
	return m.execErr
}

func TestSimpleWorker_ExecutesJob(t *testing.T) {
	jobQueue := make(chan job.Job, 1)
	worker := NewSimpleWorker(1, jobQueue)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker.Start(ctx)

	j := newMockJob(1, 0)
	jobQueue <- j

	select {
	case <-j.done:
		// success
	case <-time.After(1 * time.Second):
		t.Fatal("job was not executed")
	}

	j.mu.Lock()
	defer j.mu.Unlock()
	if !j.executed {
		t.Fatal("expected job to be executed")
	}
}

func TestSimpleWorker_StopPreventsExecution(t *testing.T) {
	jobQueue := make(chan job.Job, 1)
	worker := NewSimpleWorker(1, jobQueue)

	ctx := context.Background()
	worker.Start(ctx)
	worker.Stop()

	j := newMockJob(1, 1)
	jobQueue <- j

	select {
	case <-j.done:
		t.Fatal("job should not be executed after Stop()")
	case <-time.After(200 * time.Millisecond):
		// success
	}
}

func TestSimpleWorker_ExecutesMultipleJobs(t *testing.T) {
	jobQueue := make(chan job.Job, 2)
	worker := NewSimpleWorker(1, jobQueue)

	ctx := context.Background()
	worker.Start(ctx)

	j1 := newMockJob(1, 1)
	j2 := newMockJob(2, 1)

	jobQueue <- j1
	jobQueue <- j2

	//blocker
	select {
	case <-j1.done:
	case <-time.After(4 * time.Second):
		t.Fatal("job1 not executed")
	}

	//blocker
	select {
	case <-j2.done:
	case <-time.After(4 * time.Second):
		t.Fatal("job2 not executed")
	}
}

func TestSimpleWorker_ExecutesMultipleJobs2(t *testing.T) {
	jobQueue := make(chan job.Job, 2)
	worker := NewSimpleWorker(1, jobQueue)

	ctx := context.Background()
	worker.Start(ctx)

	j1 := newMockJob(1, 1)
	j2 := newMockJob(2, 1)

	jobQueue <- j1
	jobQueue <- j2

	//blocker
	select {
	case <-j1.done:
	case <-time.After(4 * time.Second):
		t.Fatal("job1 not executed")
	}

	//blocker
	select {
	case <-j2.done:
	case <-time.After(6 * time.Second):
		t.Fatal("job2 not executed")
	}
}

func TestSimpleWorker_JobErrorDoesNotStopWorker(t *testing.T) {
	jobQueue := make(chan job.Job, 2)
	worker := NewSimpleWorker(1, jobQueue)

	ctx := context.Background()
	worker.Start(ctx)

	j1 := newMockJob(1, 1)
	j1.execErr = errors.New("fail")

	j2 := newMockJob(2, 1)

	jobQueue <- j1
	jobQueue <- j2

	<-j1.done
	<-j2.done // worker still alive
}
