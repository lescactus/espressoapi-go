package repository

import "errors"

var (
	ErrSheetAlreadyExists = errors.New("sheet already exists")
	ErrSheetDoesNotExist  = errors.New("sheet does not exists")

	ErrBeansDoesNotExist = errors.New("beans does not exists")
)
