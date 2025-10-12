package common

import (
	"github.com/gofiber/fiber/v3"
)

const (
	StateKeyService = "serv"
	StateKeyConfig  = "cfg"
)

func GetState[T any](ctx fiber.Ctx, key string) (T, bool) {
	val, ok := ctx.App().State().Get(key)
	if !ok {
		var zero T
		return zero, false
	}
	vv, ok := val.(T)
	if !ok {
		var zero T
		return zero, false
	}
	return vv, true
}

func MustGetState[T any](ctx fiber.Ctx, key string) T {
	val, ok := GetState[T](ctx, key)
	if !ok {
		panic("failed to get state: " + key)
	}
	return val
}
