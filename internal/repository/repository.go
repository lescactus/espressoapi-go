package repository

import (
	"context"

	"github.com/lescactus/espressoapi-go/internal/models/sql"
)

type SheetRepository interface {
	CreateSheet(ctx context.Context, sheet *sql.Sheet) error
	GetSheetById(ctx context.Context, id int) (*sql.Sheet, error)
	GetSheetByName(ctx context.Context, name string) (*sql.Sheet, error)
	GetAllSheets(ctx context.Context) ([]sql.Sheet, error)
	UpdateSheetById(ctx context.Context, id int, sheet *sql.Sheet) (*sql.Sheet, error)
	DeleteSheetById(ctx context.Context, id int) error
	Ping(ctx context.Context) error
}

type BeansRepository interface {
	CreateBeans(ctx context.Context, beans *sql.Beans) error
	GetBeansById(ctx context.Context, id int) (*sql.Beans, error)
	GetAllBeans(ctx context.Context) ([]sql.Beans, error)
	UpdateBeansById(ctx context.Context, id int, beans *sql.Beans) (*sql.Beans, error)
	DeleteBeansById(ctx context.Context, id int) error
	Ping(ctx context.Context) error
}
