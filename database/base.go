package database

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog"
)

type Scannable interface {
	Scan(dest ...interface{}) error
}

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
				emailallowance INTEGER NOT NULL DEFAULT 16,

				-- school info is 1:1 with teachers
				schoolname     TEXT,
				schoolcity     TEXT,
				schoolstate    TEXT
			);
		`,
		`
			CREATE TABLE IF NOT EXISTS teams (
				id                  TEXT NOT NULL,
				teacheremail        TEXT NOT NULL,
				name                TEXT NOT NULL,
				division            TEXT NOT NULL,
				divisionexplanation TEXT NOT NULL,
				inperson            BOOLEAN NOT NULL,

				PRIMARY KEY (id, teacheremail)
			);
		`,
		`
			CREATE TABLE IF NOT EXISTS students (
				email                  TEXT NOT NULL PRIMARY KEY,
				teamid                 TEXT NOT NULL,
				name                   TEXT NOT NULL,
				age                    INTEGER NOT NULL,
				parentemail            TEXT,
				signatory              TEXT,
				dietaryrestrictions    TEXT,
				campustour             BOOLEAN,
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
