package decorator

import "context"

type Command any
type Result any
type Query any

type CommandHandler[C Command] interface {
	Handle(ctx context.Context, cmd C) error
}

type QueryHandler[Q Query, R Result] interface {
	Handle(ctx context.Context, query Q) (R, error)
}
