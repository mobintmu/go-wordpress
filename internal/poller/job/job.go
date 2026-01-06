package job

import "context"

type Job interface {
	ID() string
	Service() string
	Execute(ctx context.Context) error
}
