package database

import (
	"context"
	"database/sql"
	"fmt"
)

func (d *Database) GetStudentByEmail(ctx context.Context, email string) (*Student, error) {
	var student Student
	var parentEmail, signatory, dietaryRestrictions sql.NullString
	var campusTour sql.NullBool
	err := d.DB.QueryRowContext(ctx, `
		SELECT teamid, email, name, age, parentemail, signatory, previouslyparticipated,
			emailconfirmed, liabilitywaiver, computerusewaiver,
			campustour, dietaryrestrictions
		FROM students
		WHERE email = $1
	`, email).Scan(&student.TeamID, &student.Email, &student.Name, &student.Age,
		&parentEmail, &signatory, &student.PreviouslyParticipated, &student.EmailConfirmed,
		&student.LiabilitySigned, &student.ComputerUseWaiverSigned,
		&campusTour, &dietaryRestrictions)
	if err != nil {
		return nil, err
	}

	if parentEmail.Valid {
		student.ParentEmail = parentEmail.String
	}

	if signatory.Valid {
		student.Signatory = signatory.String
	}

	if dietaryRestrictions.Valid {
		student.DietaryRestrictions = dietaryRestrictions.String
	}

	if campusTour.Valid {
		student.CampusTour = campusTour.Bool
	}

	return &student, err
}

func (d *Database) ConfirmStudent(ctx context.Context, email string, campusTour bool, dietaryRestrictions, parentEmail string) error {
	_, err := d.DB.ExecContext(ctx, `
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
	_, err := d.DB.ExecContext(ctx, q, signatory, email)
	return err
}

func (d *Database) GetAllDietaryRestrictions(ctx context.Context) ([]string, error) {
	rows, err := d.DB.QueryContext(ctx, `
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
