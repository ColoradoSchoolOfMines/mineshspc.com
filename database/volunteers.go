package database

import (
	"context"
	"database/sql"
)

func (d *Database) GetAllVolunteers(ctx context.Context) ([]string, error) {
	rows, err := d.DB.Query(ctx, "SELECT email FROM volunteers ORDER BY email")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}
	return emails, rows.Err()
}

func (d *Database) AddVolunteer(ctx context.Context, email string) error {
	_, err := d.DB.Exec(ctx, "INSERT OR IGNORE INTO volunteers (email) VALUES ($1)", email)
	return err
}

func (d *Database) RemoveVolunteer(ctx context.Context, email string) error {
	_, err := d.DB.Exec(ctx, "DELETE FROM volunteers WHERE email = $1", email)
	return err
}

func (d *Database) IsEmailVolunteer(ctx context.Context, email string) (bool, error) {
	var isVolunteer bool
	err := d.DB.QueryRow(ctx, "SELECT true FROM volunteers WHERE email = $1", email).Scan(&isVolunteer)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
