package example

import (
	"context"
	"log"
	"math/rand"
	"time"
)

type ExampleJob struct {
	id      string
	service string
}

func NewExampleJob(id string, service string) *ExampleJob {
	return &ExampleJob{
		id:      id,
		service: service,
	}
}

func (j *ExampleJob) ID() string {
	return j.id
}

func (j *ExampleJob) Service() string {
	return j.service
}

func (j *ExampleJob) Execute(ctx context.Context) error {
	// real work goes here
	log.Printf("executing job id=%s service=%s\n", j.id, j.service)
	// Seed the random number generator (important for different results each run)
	rand.Seed(time.Now().UnixNano())

	// Generate a random sleep duration between 1 and 5 seconds
	sleepSecs := rand.Intn(10) + 5 // 1 to 5 inclusive
	sleepSecs = 1
	log.Printf("Sleeping for %d seconds...executing job id=%s service=%s \n", sleepSecs, j.id, j.service)

	// Sleep for the random duration
	time.Sleep(time.Duration(sleepSecs) * time.Second)
	log.Printf("Awake...executing job id=%s service=%s \n", j.id, j.service)

	return nil
}
