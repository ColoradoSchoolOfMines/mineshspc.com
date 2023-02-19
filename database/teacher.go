package database

import (
	"database/sql"
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

func (d *Database) scanTeacher(row Scannable) (*Teacher, error) {
	var schoolName, schoolCity, schoolState sql.NullString
	var t Teacher
	if err := row.Scan(&t.Name, &t.Email, &t.EmailConfirmed, &schoolName, &schoolCity, &schoolState); err != nil {
		return nil, err
	}

	if schoolName.Valid {
		t.SchoolName = schoolName.String
	}
	if schoolCity.Valid {
		t.SchoolCity = schoolCity.String
	}
	if schoolState.Valid {
		t.SchoolState = schoolState.String
	}

	return &t, nil
}

func (d *Database) GetTeacherByEmail(email string) (*Teacher, error) {
	row := d.Raw.QueryRow(`
		SELECT t.name, t.email, t.emailconfirmed, t.schoolname, t.schoolcity, t.schoolstate
		FROM teachers t
		WHERE t.email = ?
	`, email)
	return d.scanTeacher(row)
}

func (d *Database) SetTeacherSchoolInfo(email, schoolName, schoolCity, schoolState string) error {
	_, err := d.Raw.Exec(`
		UPDATE teachers
		SET schoolname = ?, schoolcity = ?, schoolstate = ?
		WHERE email = ?
	`, schoolName, schoolCity, schoolState, email)
	return err
}
