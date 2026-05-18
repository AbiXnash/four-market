package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		if len(password) < 8 {
			return false
		}
		hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
		hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
		hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
		return hasUpper && hasLower && hasDigit
	})
}

func Struct(s any) error {
	return validate.Struct(s)
}

func Var(field any, tag string) error {
	return validate.Var(field, tag)
}
