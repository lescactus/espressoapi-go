package main

import (
	"embed"

	"github.com/lescactus/espressoapi-go/cmd"
)

//go:embed migrations/sql/*
var dbMigrations embed.FS

func main() {
	cmd.Execute(dbMigrations)
}
