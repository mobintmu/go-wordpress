package worker

import "context"

type Worker interface {
	Start(ctx context.Context)
	Stop()
}
