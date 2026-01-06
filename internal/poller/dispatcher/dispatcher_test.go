package dispatcher

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

type mockJob struct {
	id      string
	service string

	mu       sync.Mutex
	executed bool
	done     chan struct{}
}

func newMockJob(service string) *mockJob {
	return &mockJob{
		id:      "job-1",
		service: service,
		done:    make(chan struct{}),
	}
}

func (m *mockJob) ID() string {
	return m.id
}

func (m *mockJob) Service() string {
	return m.service
}

func (m *mockJob) Execute(ctx context.Context) error {
	fmt.Println("3 second before execute job id", m.id)
	time.Sleep(3 * time.Second)

	m.mu.Lock()
	m.executed = true
	m.mu.Unlock()

	close(m.done)
	return nil
}

func TestDispatcher_Register(t *testing.T) {
	d := New()

	d.Register("email", 2, 10)

	if _, ok := d.queues["email"]; !ok {
		t.Fatal("queue was not created")
	}

	if len(d.workers["email"]) != 2 {
		t.Fatalf("expected 2 workers, got %d", len(d.workers["email"]))
	}
}

func TestDispatcher_Dispatch_ExecutesJob(t *testing.T) {
	d := New()

	d.Register("email", 1, 1)
	d.Start()

	job := newMockJob("email")

	err := d.Dispatch(job)
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}

	select {
	case <-job.done:
		// success
	case <-time.After(5 * time.Second):
		t.Fatal("job was not executed")
	}

	job.mu.Lock()
	defer job.mu.Unlock()

	if !job.executed {
		t.Fatal("expected job to be executed")
	}
}

func TestDispatcher_Dispatch_MultipleJobs(t *testing.T) {
	d := New()

	d.Register("email", 1, 2)
	d.Start()

	job1 := newMockJob("email")
	job2 := newMockJob("email")

	_ = d.Dispatch(job1)
	_ = d.Dispatch(job2)

	select {
	case <-job1.done:
	case <-time.After(5 * time.Second):
		t.Fatal("job1 not executed")
	}

	select {
	case <-job2.done:
	case <-time.After(5 * time.Second):
		t.Fatal("job2 not executed")
	}
}

func TestDispatcher_Stop(t *testing.T) {
	d := New()

	d.Register("email", 1, 1)
	d.Start()
	d.Stop()

	job := newMockJob("email")

	err := d.Dispatch(job)
	if err == nil {
		t.Fatal("expected error or panic prevention after Stop")
	}
}
