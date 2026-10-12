package pgtest

import (
	"net/url"

	"example.com/app/internal/migrations"
)

const (
	containerName = "app-pgtest-v1"
	image         = "postgres:17.6-alpine"
)

var (
	migrationFiles = migrations.FS
	migrate        = migrations.Up
	dsnParams      = url.Values{"pool_max_conns": {"4"}}
)
