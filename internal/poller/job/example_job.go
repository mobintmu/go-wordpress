package job

import "time"

type BaseJob struct {
	JobID       string
	ServiceName string
	CreatedAt   time.Time
	Payload     any
}

func (j *BaseJob) ID() string {
	return j.JobID
}

func (j *BaseJob) Service() string {
	return j.ServiceName
}
