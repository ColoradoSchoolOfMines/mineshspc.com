package database

import (
	"context"
	"database/sql"
)

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
