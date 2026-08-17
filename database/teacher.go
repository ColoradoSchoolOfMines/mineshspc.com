package database

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"go.mau.fi/util/dbutil"
	"go.mau.fi/util/exerrors"
)

type Teacher struct {
	Name           string `column:"name"`
	Email          string `column:"email"`
	EmailConfirmed bool   `column:"emailconfirmed"`
	EmailAllowance int    `column:"emailallowance"`
	SchoolName     string `column:"schoolname"`
	SchoolCity     string `column:"schoolcity"`
	SchoolState    string `column:"schoolstate"`
}

const teacherSelectColumns = `
	t.name, t.email, t.emailconfirmed, t.emailallowance,
	COALESCE(t.schoolname, ''), COALESCE(t.schoolcity, ''), COALESCE(t.schoolstate, '')
`

var teacherColumns = []string{"name", "email", "emailconfirmed", "emailallowance", "schoolname", "schoolcity", "schoolstate"}
var scanTeacherRow = exerrors.Must(dbutil.MakeReflectScanner[Teacher](teacherColumns, dbutil.ReflectScanOptions{StructTag: "column"}))

func (t *Teacher) Scan(row dbutil.Scannable) (*Teacher, error) {
	return scanTeacherRow(row)
}

func (d *Database) NewTeacher(ctx context.Context, name, email string) error {
	_, err := d.DB.Exec(ctx, "INSERT INTO teachers (name, email) VALUES (?, ?)", name, email)
	return err
}

func (d *Database) SetEmailConfirmed(ctx context.Context, email string) error {
	_, err := d.DB.Exec(ctx, "UPDATE teachers SET emailconfirmed = TRUE WHERE email = ?", email)
	return err
}

func (d *Database) GetAllTeachers(ctx context.Context) ([]*Teacher, error) {
	return d.teacherQH.QueryMany(ctx, `
		SELECT `+teacherSelectColumns+`
		FROM teachers t
		ORDER BY t.name
	`)
}

func (d *Database) SetEmailAllowance(ctx context.Context, email string, allowance int) error {
	_, err := d.DB.Exec(ctx, `
		UPDATE teachers
		SET emailallowance = ?
		WHERE email = ?
	`, allowance, email)
	return err
}

func (d *Database) GetTeacherByEmail(ctx context.Context, email string) (*Teacher, error) {
	t, err := d.teacherQH.QueryOne(ctx, `
		SELECT `+teacherSelectColumns+`
		FROM teachers t
		WHERE t.email = ?
	`, email)
	if err == nil && t == nil {
		return nil, sql.ErrNoRows
	}
	return t, err
}

func (d *Database) GetTeacherForTeam(ctx context.Context, teamID uuid.UUID) (*Teacher, error) {
	t, err := d.teacherQH.QueryOne(ctx, `
		SELECT `+teacherSelectColumns+`
		FROM teachers t
		JOIN teams tea ON tea.teacheremail = t.email
		WHERE tea.id = ?
	`, teamID)
	if err == nil && t == nil {
		return nil, sql.ErrNoRows
	}
	return t, err
}

func (d *Database) SetTeacherSchoolInfo(ctx context.Context, email, schoolName, schoolCity, schoolState string) error {
	_, err := d.DB.Exec(ctx, `
		UPDATE teachers
		SET schoolname = ?, schoolcity = ?, schoolstate = ?
		WHERE email = ?
	`, schoolName, schoolCity, schoolState, email)
	return err
}

func (d *Database) DecrementEmailAllowance(ctx context.Context, email string) error {
	_, err := d.DB.Exec(ctx, `
		UPDATE teachers
		SET emailallowance = emailallowance - 1
		WHERE email = ?
	`, email)
	return err
}
