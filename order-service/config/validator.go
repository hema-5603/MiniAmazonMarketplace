package config

import "github.com/go-playground/validator/v10"

//CustomValidator wraps the go-playground validator to make it compatible with Echo
type CustomValidator struct{
	Validator *validator.Validate
}

//Validator executes the struct tags
func (cv *CustomValidator) Validate(i interface{}) error{
	return cv.Validator.Struct(i)
}