package main

import (
	"embed"

	"github.com/lescactus/espressoapi-go/cmd"
)

//go:embed migrations/sql/*
var dbMigrations embed.FS

//go:embed docs/swagger.json
var swagger embed.FS

func main() {
	cmd.Execute(&dbMigrations, swagger)
}
