package errors

import "errors"

var (
	ErrSheetAlreadyExists = errors.New("sheet already exists")
	ErrSheetDoesNotExist  = errors.New("sheet does not exists")
	ErrSheetNameIsEmpty   = errors.New("sheet name is empty")

	ErrRoasterAlreadyExists = errors.New("roaster already exists")
	ErrRoasterDoesNotExist  = errors.New("roaster does not exists")
	ErrRoasterNameIsEmpty   = errors.New("roaster name is empty")

	ErrBeansDoesNotExist         = errors.New("beans does not exists")
	ErrBeansForeignKeyConstraint = errors.New("beans foreign key constraint failed")
	ErrBeansIsNil                = errors.New("beans is nil")
	ErrBeansNameIsEmpty          = errors.New("beans name is empty")

	ErrShotDoesNotExist         = errors.New("shot does not exists")
	ErrShotRatingOutOfRange     = errors.New("shot rating is out of range. Must be between 0.0 and 10.0")
	ErrShotForeignKeyConstraint = errors.New("shot foreign key constraint failed")
)
