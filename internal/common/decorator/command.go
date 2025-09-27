package decorator

import "context"

type CommandHandler[C any, R any] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}
