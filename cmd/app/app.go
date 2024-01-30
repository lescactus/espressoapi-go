package app

import (
	"io/fs"

	"github.com/jmoiron/sqlx"
	"github.com/lescactus/espressoapi-go/internal/config"
	"github.com/rs/zerolog"
)

type Application struct {
	Db           *sqlx.DB
	Cfg          *config.App
	Logger       *zerolog.Logger
	MigrationsFS *fs.FS
}

var App *Application
