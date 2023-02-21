package database

func (d *Database) GetStudentByEmail(email string) (*Student, error) {
	var student Student
	err := d.Raw.QueryRow(`
		SELECT teamid, email, name, age, parentemail, previouslyparticipated,
			emailconfirmed, liabilitywaiver, computerusewaiver, multimediareleaseform
		FROM students
		WHERE email = $1
	`, email).Scan(&student.TeamID, &student.Email, &student.Name, &student.Age,
		&student.ParentEmail, &student.PreviouslyParticipated, &student.EmailConfirmed,
		&student.LiabilitySigned, &student.ComputerUseWaiverSigned, &student.MultimediaReleaseForm)
	return &student, err
}
