package database

import (
	"fmt"

	"github.com/google/uuid"
)

type Division string

const (
	DivisionBeginner Division = "Beginner"
	DivisionAdvanced Division = "Advanced"
)

func ParseDivision(s string) (Division, error) {
	switch s {
	case "Beginner":
		return DivisionBeginner, nil
	case "Advanced":
		return DivisionAdvanced, nil
	default:
		return "", fmt.Errorf("invalid division: %s", s)
	}
}

type Team struct {
	ID                  uuid.UUID
	TeacherEmail        string
	Name                string
	Division            Division
	DivisionExplanation string
	InPerson            bool
	Members             []Student
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

func (d *Database) scanTeam(row Scannable) (*Team, error) {
	var team Team
	err := row.Scan(&team.ID, &team.TeacherEmail, &team.Name, &team.Division, &team.InPerson, &team.DivisionExplanation)
	return &team, err
}

func (d *Database) scanTeamWithStudents(row Scannable) (*Team, error) {
	team, err := d.scanTeam(row)

	studentRows, err := d.Raw.Query(`
		SELECT s.email, s.name, s.parentemail, s.previouslyparticipated, s.emailconfirmed,
			s.computerusewaiver, s.multimediareleaseform
		FROM students s
		WHERE s.teamid = ?
	`, team.ID)
	if err != nil {
		return nil, err
	}
	defer studentRows.Close()
	for studentRows.Next() {
		var student Student
		if err := studentRows.Scan(&student.Email, &student.Name, &student.ParentEmail, &student.PreviouslyParticipated,
			&student.EmailConfirmed, &student.ComputerUseWaiverSigned, &student.MultimediaReleaseForm); err != nil {
			return nil, err
		}

		team.Members = append(team.Members, student)
	}

	return team, err
}

func (d *Database) GetTeacherTeams(email string) ([]*Team, error) {
	rows, err := d.Raw.Query(`
		SELECT t.id, t.teacheremail, t.name, t.division, t.inperson, t.divisionexplanation
		FROM teams t
		JOIN teachers tt ON tt.email = t.teacheremail
		WHERE tt.email = ?
	`, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*Team

	for rows.Next() {
		team, err := d.scanTeamWithStudents(rows)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}

	return teams, err
}

func (d *Database) GetTeam(email string, teamID string) (*Team, error) {
	row := d.Raw.QueryRow(`
		SELECT t.id, t.teacheremail, t.name, t.division, t.inperson, t.divisionexplanation
		FROM teams t
		JOIN teachers tt ON tt.email = t.teacheremail
		WHERE tt.email = ?
		  AND t.id = ?
	`, email, teamID)
	return d.scanTeamWithStudents(row)
}

func (d *Database) UpsertTeam(teacherEmail string, teamID uuid.UUID, name string, division Division, inPerson bool, divisionExplanation string) error {
	_, err := d.Raw.Exec(`
		INSERT OR REPLACE INTO teams (id, teacheremail, name, division, inperson, divisionexplanation)
		VALUES (?, ?, ?, ?, ?, ?)
	`, teamID, teacherEmail, name, division, inPerson, divisionExplanation)
	return err
}
