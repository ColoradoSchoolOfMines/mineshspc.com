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

	teacherQH             *dbutil.QueryHelper[*Teacher]
	studentQH             *dbutil.QueryHelper[*Student]
	teamQH                *dbutil.QueryHelper[*Team]
	teamWithTeacherNameQH *dbutil.QueryHelper[*TeamWithTeacherName]
}

func NewDatabase(logger dbutil.DatabaseLogger, cfg dbutil.Config) (db *Database) {
	rawDB := exerrors.Must(dbutil.NewFromConfig("mineshspc", cfg, logger))
	rawDB.UpgradeTable = dbutil.BuildUpgradeTable().WithFS(rawUpgrades).Finish()
	db = &Database{
		DB: rawDB,

		teacherQH:             dbutil.MakeQueryHelperReflect[*Teacher](rawDB),
		studentQH:             dbutil.MakeQueryHelperReflect[*Student](rawDB),
		teamQH:                dbutil.MakeQueryHelperReflect[*Team](rawDB),
		teamWithTeacherNameQH: dbutil.MakeQueryHelperReflect[*TeamWithTeacherName](rawDB),
	}
	return
}
