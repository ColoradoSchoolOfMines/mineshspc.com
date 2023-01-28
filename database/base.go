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
