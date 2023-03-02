package database

import "database/sql"

func (d *Database) GetStudentByEmail(email string) (*Student, error) {
	var student Student
	var parentEmail, dietaryRestrictions sql.NullString
	var campusTour sql.NullBool
	err := d.Raw.QueryRow(`
		SELECT teamid, email, name, age, parentemail, previouslyparticipated,
			emailconfirmed, liabilitywaiver, computerusewaiver, multimediareleaseform,
			campustour, dietaryrestrictions
		FROM students
		WHERE email = $1
	`, email).Scan(&student.TeamID, &student.Email, &student.Name, &student.Age,
		&parentEmail, &student.PreviouslyParticipated, &student.EmailConfirmed,
		&student.LiabilitySigned, &student.ComputerUseWaiverSigned, &student.MultimediaReleaseForm,
		&campusTour, &dietaryRestrictions)
	if err != nil {
		return nil, err
	}

	if parentEmail.Valid {
		student.ParentEmail = parentEmail.String
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
