package errors

import "errors"

var (
	ErrSheetAlreadyExists = errors.New("sheet already exists")
	ErrSheetDoesNotExist  = errors.New("sheet does not exists")

	ErrRoasterAlreadyExists = errors.New("roaster already exists")
	ErrRoasterDoesNotExist  = errors.New("roaster does not exists")

	ErrBeansDoesNotExist = errors.New("beans does not exists")

	ErrShotDoesNotExist     = errors.New("shot does not exists")
	ErrShotRatingOutOfRange = errors.New("shot rating is out of range. Must be between 0.0 and 10.0")
)
