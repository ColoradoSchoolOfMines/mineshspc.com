package database

import (
	"embed"

	"go.mau.fi/util/dbutil"
)

//go:embed *.sql
var rawUpgrades embed.FS

type Database struct {
	DB *dbutil.Database
}

func NewDatabase(db *dbutil.Database) *Database {
	db.UpgradeTable = dbutil.BuildUpgradeTable().WithFS(rawUpgrades).Finish()
	return &Database{DB: db}
}
