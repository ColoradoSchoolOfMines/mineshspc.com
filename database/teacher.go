package database

import (
	"time"

	"github.com/google/uuid"
)

func (d *Database) NewTeacher(name, email string) error {
	_, err := d.Raw.Exec("INSERT INTO teachers (name, email) VALUES (?, ?)", name, email)
	return err
}

func (d *Database) SetEmailConfirmed(email string) error {
	_, err := d.Raw.Exec("UPDATE teachers SET emailconfirmed = TRUE WHERE email = ?", email)
	return err
}

func (d *Database) NewTeacherSession(email string, sessionToken uuid.UUID, expiresAt time.Time) error {
	_, err := d.Raw.Exec(`
		INSERT INTO sessions (email, token, expires)
		VALUES (?, ?, ?)
	`, email, sessionToken.String(), expiresAt.UnixMilli())
	return err
}
