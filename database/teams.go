package database

import "github.com/google/uuid"

type Division string

const (
	DivisionBeginner Division = "Beginner"
	DivisionAdvanced Division = "Advanced"
)

type Team struct {
	ID           uuid.UUID
	TeacherEmail string
	Name         string
	Division     Division
	InPerson     bool
	Members      []Student
}

type Student struct {
	TeamID                  uuid.UUID
	Email                   string
	Name                    string
	ParentEmail             string
	PreviouslyParticipated  bool
	EmailConfirmed          bool
	LiabilitySigned         bool
	ComputerUseWaiverSigned bool
	MultimediaReleaseForm   bool
}

func (d *Database) GetTeacherTeams(email string) ([]Team, error) {
	rows, err := d.Raw.Query(`
		SELECT t.id, t.teacheremail, t.name, t.division
		FROM teams t
		JOIN teachers tt ON tt.email = t.teacheremail
		WHERE tt.email = ?
	`, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []Team

	for rows.Next() {
		var team Team
		if err := rows.Scan(&team.ID, &team.TeacherEmail, &team.Name, &team.Division); err != nil {
			return nil, err
		}
		studentRows, err := d.Raw.Query(`
			SELECT s.email, s.name, s.parentemail, s.previouslyparticipated, s.emailconfirmed,
				s.computerusewaiversigned, s.multimediareleaseform
			FROM students s
			WHERE s.teamid = ?
		`)
		if err != nil {
			return nil, err
		}
		defer studentRows.Close()
		for studentRows.Next() {
			var student Student
			if err := rows.Scan(&student.Email, &student.Name, &student.ParentEmail, &student.PreviouslyParticipated,
				&student.EmailConfirmed, &student.ComputerUseWaiverSigned, &student.MultimediaReleaseForm); err != nil {
				return nil, err
			}

			team.Members = append(team.Members, student)
		}
		teams = append(teams, team)
	}

	return teams, err
}
