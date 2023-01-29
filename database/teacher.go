package database

import (
	"time"

	"github.com/google/uuid"
)

type Teacher struct {
	Name           string
	Email          string
	EmailConfirmed bool
	SchoolName     string
	SchoolCity     string
	SchoolState    string
}

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

func (d *Database) GetTeacherByEmail(email string) (*Teacher, error) {
	row := d.Raw.QueryRow(`
		SELECT t.name, t.email, t.emailconfirmed, t.schoolname, t.schoolcity, t.schoolstate
		FROM teachers t
		WHERE t.email = ?
	`, email)

	var t Teacher
	if err := row.Scan(&t.Name, &t.Email, &t.EmailConfirmed, &t.SchoolName, &t.SchoolCity, &t.SchoolState); err != nil {
		return nil, err
	}
	return &t, nil
}

func (d *Database) GetTeacherBySessionToken(sessionToken uuid.UUID) (*Teacher, error) {
	row := d.Raw.QueryRow(`
		SELECT t.name, t.email, t.emailconfirmed, t.schoolname, t.schoolcity, t.schoolstate
		FROM teachers t
		JOIN sessions s ON s.email = t.email
		WHERE s.token = ?
			AND s.expires >= ?
	`, sessionToken.String(), time.Now().UnixMilli())

	var t Teacher
	if err := row.Scan(&t.Name, &t.Email, &t.EmailConfirmed, &t.SchoolName, &t.SchoolCity, &t.SchoolState); err != nil {
		return nil, err
	}
	return &t, nil
}

func (d *Database) SetTeacherSchoolInfo(email, schoolName, schoolCity, schoolState string) error {
	_, err := d.Raw.Exec(`
		UPDATE teachers
		SET schoolname = ?, schoolcity = ?, schoolstate = ?
		WHERE email = ?
	`, schoolName, schoolCity, schoolState, email)
	return err
}
