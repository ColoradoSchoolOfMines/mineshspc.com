package database

import "database/sql"

func (d *Database) IsEmailAdmin(email string) (bool, error) {
	var isAdmin bool
	err := d.Raw.QueryRow("SELECT true FROM admins WHERE email = $1", email).Scan(&isAdmin)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
