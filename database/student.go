package database

import (
	"context"
	"fmt"
)

func (d *Database) GetStudentByEmail(ctx context.Context, email string) (*Student, error) {
	return scanStudentRow(d.DB.QueryRow(ctx, `
		SELECT `+studentSelectColumns+`
		FROM students s
		WHERE s.email = $1
	`, email))
}

func (d *Database) ConfirmStudent(ctx context.Context, email string, campusTour bool, dietaryRestrictions, parentEmail string) error {
	_, err := d.DB.Exec(ctx, `
		UPDATE students
		SET emailconfirmed = true, campustour = $1, dietaryrestrictions = $2, parentemail = $3
		WHERE email = $4
	`, campusTour, dietaryRestrictions, parentEmail, email)
	return err
}

func (d *Database) SignFormsForStudent(ctx context.Context, email, signatory string, computerUse bool) error {
	computerUseQuery := ""
	if computerUse {
		computerUseQuery = "computerusewaiver = true,"
	}
	q := fmt.Sprintf(`
		UPDATE students
		SET liabilitywaiver = true, %s signatory = $1
		WHERE email = $2
	`, computerUseQuery)
	_, err := d.DB.Exec(ctx, q, signatory, email)
	return err
}

func (d *Database) GetAllDietaryRestrictions(ctx context.Context) ([]string, error) {
	rows, err := d.DB.Query(ctx, `
		SELECT dietaryrestrictions
		FROM students
		WHERE dietaryrestrictions != '' AND dietaryrestrictions IS NOT NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var dietaryRestrictions []string
	for rows.Next() {
		var restriction string
		if err = rows.Scan(&restriction); err != nil {
			return nil, err
		}
		dietaryRestrictions = append(dietaryRestrictions, restriction)
	}
	return dietaryRestrictions, nil
}

func (d *Database) MarkQRCodeSent(ctx context.Context, email string) error {
	_, err := d.DB.Exec(ctx, `
		UPDATE students
		SET qrcodesent = true
		WHERE email = $1
	`, email)
	return err
}

func (d *Database) CheckInStudent(ctx context.Context, email string) error {
	_, err := d.DB.Exec(ctx, `
		UPDATE students
		SET checkedin = true
		WHERE email = $1
	`, email)
	return err
}

func (d *Database) UncheckInStudent(ctx context.Context, email string) error {
	_, err := d.DB.Exec(ctx, `UPDATE students SET checkedin = false WHERE email = $1`, email)
	return err
}
