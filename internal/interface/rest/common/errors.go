package common

import "github.com/gofiber/fiber/v3"

type Response[T any] struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type Error struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Response() *Response[any] {
	return &Response[any]{
		Status:  e.Status,
		Message: e.Message,
	}
}

func NewError(status int, message string) *Error {
	return &Error{
		Status:  status,
		Message: message,
	}
}

func NewSuccess[T any](data T) *Response[T] {
	return &Response[T]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    data,
	}
}
