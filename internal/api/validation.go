package api

import (
	"github.com/beevik/guid"
	"github.com/go-playground/validator/v10"
)

var validGUID = func(fl validator.FieldLevel) bool {
	_, err := guid.ParseString(fl.Field().String())
	return err == nil
}

func RegisterCustomeValidators(v *validator.Validate) {
	v.RegisterValidation("guid", validGUID)
}
