package database

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog"
)

type Database struct {
	Raw *sql.DB
	Log *zerolog.Logger
}

func NewDatabase(db *sql.DB, logger zerolog.Logger) *Database {
	return &Database{
		Raw: db,
		Log: &logger,
	}
}

func (d *Database) RunMigrations() {
	d.Log.Info().Msg("running migrations")

	migrations := []string{
		`
			CREATE TABLE IF NOT EXISTS teachers (
				email          TEXT NOT NULL PRIMARY KEY,
				name           TEXT NOT NULL,
				emailconfirmed BOOLEAN NOT NULL DEFAULT FALSE,

				-- school info is 1:1 with teachers
				schoolname     TEXT,
				schoolcity     TEXT,
				schoolstate    TEXT
			);
		`,
		`
			CREATE TABLE IF NOT EXISTS sessions (
				email   TEXT NOT NULL,
				token   TEXT NOT NULL,
				expires INTEGER NOT NULL,
				PRIMARY KEY (email, token)
			);
		`,
		`
			CREATE TABLE IF NOT EXISTS teams (
				id           TEXT NOT NULL PRIMARY KEY,
				teacheremail TEXT NOT NULL,
				name         TEXT NOT NULL,
				division     TEXT NOT NULL,
				inperson     BOOLEAN NOT NULL,

				UNIQUE (teacheremail, name, division)
			);
		`,
		`
			CREATE TABLE IF NOT EXISTS students (
				email                  TEXT NOT NULL PRIMARY KEY,
				teamid                 TEXT NOT NULL,
				name                   TEXT NOT NULL,
				parentemail            TEXT NOT NULL,
				previouslyparticipated BOOLEAN NOT NULL DEFAULT FALSE,
				emailconfirmed         BOOLEAN NOT NULL DEFAULT FALSE,
				liabilitywaiver        BOOLEAN NOT NULL DEFAULT FALSE,
				computerusewaiver      BOOLEAN NOT NULL DEFAULT FALSE,
				multimediareleaseform  BOOLEAN NOT NULL DEFAULT FALSE
			)
		`,
	}
	txn, err := d.Raw.BeginTx(context.TODO(), nil)
	if err != nil {
		d.Log.Fatal().Err(err).Msg("failed to begin transaction")
	}
	for i, m := range migrations {
		d.Log.Info().Int("migration", i+1).Msg("running migration")
		_, err = txn.Exec(m)
		if err != nil {
			d.Log.Fatal().Err(err).Int("migration_number", i+1).Msg("failed to run migration")
			err = txn.Rollback()
			if err != nil {
				d.Log.Fatal().Err(err).Msg("failed to rollback transaction")
			}
			return
		}
	}
	if err = txn.Commit(); err != nil {
		d.Log.Fatal().Err(err).Msg("failed to commit transaction")
	}
}
