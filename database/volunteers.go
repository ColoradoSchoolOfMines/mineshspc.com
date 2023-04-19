package database

import (
	"context"
	"database/sql"
)

func (d *Database) IsEmailVolunteer(ctx context.Context, email string) (bool, error) {
	// TODO this check probably should not be in database layer, but whatever
	if isAdmin, err := d.IsEmailAdmin(ctx, email); err == nil && isAdmin {
		// Admins are always volunteers
		return true, nil
	}

	var isVolunteer bool
	err := d.DB.QueryRowContext(ctx, "SELECT true FROM volunteers WHERE email = $1", email).Scan(&isVolunteer)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
