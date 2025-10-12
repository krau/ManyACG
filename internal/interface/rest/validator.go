package rest

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type structValidator struct {
	validate *validator.Validate
}

func (v *structValidator) Validate(out any) error {
	err := v.validate.Struct(out)
	if err == nil {
		return nil
	}

	if errs, ok := err.(validator.ValidationErrors); ok {
		t := reflect.TypeOf(out)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		first := errs[0]
		field, _ := t.FieldByName(first.StructField())
		msg := field.Tag.Get("message")
		if msg == "" {
			// msg = first.Error()
			msg = fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", first.Field(), first.Tag())
		}
		result := ValidationError{
			Field:   first.Field(),
			Message: msg,
		}
		return fiber.NewError(fiber.StatusBadRequest, result.Message)
	}
	return err
}

func NewStructValidator() *structValidator {
	validate := validator.New()
	validate.RegisterValidation("objectid", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		log.Debug("validating objectid", "field", fl.FieldName(), "value", s)
		_, err := objectuuid.ObjectIDFromHex(s)
		return err == nil
	})
	return &structValidator{validate: validate}
}
