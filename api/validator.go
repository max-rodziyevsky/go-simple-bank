package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/max-rodziyevsky/go-simple-bank/util"
)

var validCurrency validator.Func = func(fl validator.FieldLevel) bool {
	if currency, ok := fl.Field().Interface().(string); ok {
		// check currency is supported
		return util.IsSupportedCurrency(currency)
	}
	return false
}
