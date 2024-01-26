package errors

import "errors"

var (
	ErrSheetAlreadyExists = errors.New("sheet already exists")
	ErrSheetDoesNotExist  = errors.New("sheet does not exists")

	ErrRoasterAlreadyExists = errors.New("roaster already exists")
	ErrRoasterDoesNotExist  = errors.New("roaster does not exists")

	ErrBeansDoesNotExist = errors.New("beans does not exists")
)
