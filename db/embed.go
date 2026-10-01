package db

import "embed"

// Migrations includes both dialects; each release adds matching version files.
//
//go:embed sqlite/migrations/*.sql postgres/migrations/*.sql
var Migrations embed.FS
