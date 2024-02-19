package database

import (
	"context"
	"database/sql"
)

func (d *Database) IsEmailAdmin(ctx context.Context, email string) (bool, error) {
	var isAdmin bool
	err := d.DB.QueryRow(ctx, "SELECT true FROM admins WHERE email = $1", email).Scan(&isAdmin)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
