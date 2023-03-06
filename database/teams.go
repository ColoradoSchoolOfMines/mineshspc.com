package database

import (
	"database/sql"
	"errors"
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
	Age                     int
	ParentEmail             string
	Signatory               string
	PreviouslyParticipated  bool
	EmailConfirmed          bool
	LiabilitySigned         bool
	ComputerUseWaiverSigned bool

	CampusTour          bool
	DietaryRestrictions string
}

func (d *Database) scanTeam(row Scannable) (*Team, error) {
	var team Team
	err := row.Scan(&team.ID, &team.TeacherEmail, &team.Name, &team.Division, &team.InPerson, &team.DivisionExplanation)
	return &team, err
}

func (d *Database) scanTeamWithStudents(row Scannable) (*Team, error) {
	team, err := d.scanTeam(row)
	if err != nil {
		return nil, err
	}

	studentRows, err := d.Raw.Query(`
		SELECT s.email, s.name, s.age, s.parentemail, s.signatory, s.previouslyparticipated, s.emailconfirmed,
			s.liabilitywaiver, s.computerusewaiver, s.campustour, s.dietaryrestrictions
		FROM students s
		WHERE s.teamid = ?
	`, team.ID)
	if err != nil {
		return nil, err
	}
	defer studentRows.Close()
	for studentRows.Next() {
		var s Student
		var parentEmail, signatory, dietaryRestrictions sql.NullString
		var campusTour sql.NullBool
		if err := studentRows.Scan(&s.Email, &s.Name, &s.Age, &parentEmail, &signatory, &s.PreviouslyParticipated,
			&s.EmailConfirmed, &s.LiabilitySigned, &s.ComputerUseWaiverSigned, &campusTour, &dietaryRestrictions); err != nil {
			return nil, err
		}

		if parentEmail.Valid {
			s.ParentEmail = parentEmail.String
		}

		if signatory.Valid {
			s.Signatory = signatory.String
		}

		if dietaryRestrictions.Valid {
			s.DietaryRestrictions = dietaryRestrictions.String
		}

		if campusTour.Valid {
			s.CampusTour = campusTour.Bool
		}

		team.Members = append(team.Members, s)
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

func (d *Database) GetTeam(email string, teamID uuid.UUID) (*Team, error) {
	row := d.Raw.QueryRow(`
		SELECT t.id, t.teacheremail, t.name, t.division, t.inperson, t.divisionexplanation
		FROM teams t
		JOIN teachers tt ON tt.email = t.teacheremail
		WHERE tt.email = ?
		  AND t.id = ?
	`, email, teamID)
	return d.scanTeamWithStudents(row)
}

func (d *Database) GetTeamNoMembers(teamID uuid.UUID) (*Team, error) {
	row := d.Raw.QueryRow(`
		SELECT t.id, t.teacheremail, t.name, t.division, t.inperson, t.divisionexplanation
		FROM teams t
		WHERE t.id = ?
	`, teamID)
	return d.scanTeam(row)
}

func (d *Database) UpsertTeam(teacherEmail string, teamID uuid.UUID, name string, division Division, inPerson bool, divisionExplanation string) error {
	_, err := d.Raw.Exec(`
		INSERT OR REPLACE INTO teams (id, teacheremail, name, division, inperson, divisionexplanation)
		VALUES (?, ?, ?, ?, ?, ?)
	`, teamID, teacherEmail, name, division, inPerson, divisionExplanation)
	return err
}

func (d *Database) AddTeamMember(teamID uuid.UUID, name string, studentAge int, studentEmail string, previouslyParticipated bool) error {
	_, err := d.Raw.Exec(`
		INSERT INTO students (teamid, name, age, email, previouslyparticipated)
		VALUES (?, ?, ?, ?, ?)
	`, teamID, name, studentAge, studentEmail, previouslyParticipated)
	return err
}

func (d *Database) RemoveTeamMember(teamID uuid.UUID, studentEmail string) error {
	res, err := d.Raw.Exec(`
		DELETE FROM students
		WHERE teamid = ?
			AND email = ?
	`, teamID, studentEmail)
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); err != nil {
		return err
	} else if affected != 1 {
		return errors.New("incorrect number of rows affected on delete from students table")
	}
	return nil
}
