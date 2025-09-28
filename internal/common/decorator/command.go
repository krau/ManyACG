package decorator

import "context"

type Command any
type Result any

type CommandHandler[C Command, R Result] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}
