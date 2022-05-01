package request

import (
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

type Validator struct {
	v *validator.Validate
}

func NewValidator() *Validator {
	return &Validator{v: validator.New()}
}

func (v *Validator) Validate(i interface{}) error {
	if err := v.v.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return nil
}
