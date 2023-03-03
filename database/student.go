package database

import "database/sql"

func (d *Database) GetStudentByEmail(email string) (*Student, error) {
	var student Student
	var parentEmail, signatory, dietaryRestrictions sql.NullString
	var campusTour sql.NullBool
	err := d.Raw.QueryRow(`
		SELECT teamid, email, name, age, parentemail, signatory, previouslyparticipated,
			emailconfirmed, liabilitywaiver, computerusewaiver, multimediareleaseform,
			campustour, dietaryrestrictions
		FROM students
		WHERE email = $1
	`, email).Scan(&student.TeamID, &student.Email, &student.Name, &student.Age,
		&parentEmail, &signatory, &student.PreviouslyParticipated, &student.EmailConfirmed,
		&student.LiabilitySigned, &student.ComputerUseWaiverSigned, &student.MultimediaReleaseForm,
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

func (d *Database) ConfirmStudent(email string, campusTour bool, dietaryRestrictions, parentEmail string) error {
	_, err := d.Raw.Exec(`
		UPDATE students
		SET emailconfirmed = true, campustour = $1, dietaryrestrictions = $2, parentemail = $3
		WHERE email = $3
	`, campusTour, dietaryRestrictions, email, parentEmail)
	return err
}

func (d *Database) SignFormsForStudent(email, signatory string) error {
	_, err := d.Raw.Exec(`
		UPDATE students
		SET liabilitywaiver = true, computerusewaiver = true, multimediareleaseform = true, signatory = $1
		WHERE email = $2
	`, signatory, email)
	return err
}
