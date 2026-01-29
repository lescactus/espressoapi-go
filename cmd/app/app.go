package app

import (
	"embed"
	"io/fs"

	"github.com/jmoiron/sqlx"
	"github.com/lescactus/espressoapi-go/internal/config"
	"github.com/rs/zerolog"
)

type Application struct {
	Db           *sqlx.DB
	Cfg          *config.App
	Logger       *zerolog.Logger
	MigrationsFS *embed.FS
	SwaggerFS    fs.FS
	TemplatesFS  *embed.FS
}

var App *Application
