package database

import (
	"context"
	"database/sql"

	"maunium.net/go/mautrix/util/dbutil"
)

type Teacher struct {
	Name           string
	Email          string
	EmailConfirmed bool
	EmailAllowance int
	SchoolName     string
	SchoolCity     string
	SchoolState    string
}

func (d *Database) NewTeacher(ctx context.Context, name, email string) error {
	_, err := d.DB.ExecContext(ctx, "INSERT INTO teachers (name, email) VALUES (?, ?)", name, email)
	return err
}

func (d *Database) SetEmailConfirmed(ctx context.Context, email string) error {
	_, err := d.DB.ExecContext(ctx, "UPDATE teachers SET emailconfirmed = TRUE WHERE email = ?", email)
	return err
}

func (d *Database) scanTeacher(row dbutil.Scannable) (*Teacher, error) {
	var schoolName, schoolCity, schoolState sql.NullString
	var t Teacher
	if err := row.Scan(&t.Name, &t.Email, &t.EmailConfirmed, &t.EmailAllowance, &schoolName, &schoolCity, &schoolState); err != nil {
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

func (d *Database) GetTeacherByEmail(ctx context.Context, email string) (*Teacher, error) {
	row := d.DB.QueryRowContext(ctx, `
		SELECT t.name, t.email, t.emailconfirmed, t.emailallowance, t.schoolname, t.schoolcity, t.schoolstate
		FROM teachers t
		WHERE t.email = ?
	`, email)
	return d.scanTeacher(row)
}

func (d *Database) SetTeacherSchoolInfo(ctx context.Context, email, schoolName, schoolCity, schoolState string) error {
	_, err := d.DB.ExecContext(ctx, `
		UPDATE teachers
		SET schoolname = ?, schoolcity = ?, schoolstate = ?
		WHERE email = ?
	`, schoolName, schoolCity, schoolState, email)
	return err
}

func (d *Database) DecrementEmailAllowance(ctx context.Context, email string) error {
	_, err := d.DB.ExecContext(ctx, `
		UPDATE teachers
		SET emailallowance = emailallowance - 1
		WHERE email = ?
	`, email)
	return err
}
