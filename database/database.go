package database

import (
	"embed"

	"go.mau.fi/util/dbutil"
	"go.mau.fi/util/exerrors"
)

//go:embed *.sql
var rawUpgrades embed.FS

type Database struct {
	DB *dbutil.Database
}

func NewDatabase(logger dbutil.DatabaseLogger, cfg dbutil.Config) (db *Database) {
	rawDB := exerrors.Must(dbutil.NewFromConfig("mineshspc", cfg, logger))
	rawDB.UpgradeTable = dbutil.BuildUpgradeTable().WithFS(rawUpgrades).Finish()
	db = &Database{
		DB: rawDB,
	}
	return
}
