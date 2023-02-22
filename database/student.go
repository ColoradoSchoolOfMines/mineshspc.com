package database

import "database/sql"

func (d *Database) GetStudentByEmail(email string) (*Student, error) {
	var student Student
	var dietaryRestrictions sql.NullString
	err := d.Raw.QueryRow(`
		SELECT teamid, email, name, age, parentemail, previouslyparticipated,
			emailconfirmed, liabilitywaiver, computerusewaiver, multimediareleaseform,
			campustour, dietaryrestrictions
		FROM students
		WHERE email = $1
	`, email).Scan(&student.TeamID, &student.Email, &student.Name, &student.Age,
		&student.ParentEmail, &student.PreviouslyParticipated, &student.EmailConfirmed,
		&student.LiabilitySigned, &student.ComputerUseWaiverSigned, &student.MultimediaReleaseForm,
		&student.CampusTour, &dietaryRestrictions)
	if err != nil {
		return nil, err
	}

	if dietaryRestrictions.Valid {
		student.DietaryRestrictions = dietaryRestrictions.String
	}

	return &student, err
}

func (d *Database) ConfirmStudent(email string, campusTour bool, dietaryRestrictions string) error {
	_, err := d.Raw.Exec(`
		UPDATE students
		SET emailconfirmed = true, campustour = $1, dietaryrestrictions = $2
		WHERE email = $3
	`, campusTour, dietaryRestrictions, email)
	return err
}
