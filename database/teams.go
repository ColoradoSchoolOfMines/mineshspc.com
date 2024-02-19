package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mau.fi/util/dbutil"
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
	SchoolName          string
	RegistrationTS      time.Time
}

type TeamWithTeacherName struct {
	*Team
	TeacherName string
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

	QRCodeSent bool
	CheckedIn  bool
}

func (d *Database) scanTeam(row dbutil.Scannable) (*Team, error) {
	var team Team
	var registrationTS int64
	err := row.Scan(&team.ID, &team.TeacherEmail, &team.Name, &team.Division, &team.InPerson, &team.DivisionExplanation, &team.SchoolName, &registrationTS)
	team.RegistrationTS = time.UnixMilli(registrationTS)
	return &team, err
}

func (d *Database) scanTeamWithTeacherName(row dbutil.Scannable) (*TeamWithTeacherName, error) {
	var team Team
	var teamWithTeacherName TeamWithTeacherName
	var registrationTS int64
	err := row.Scan(&team.ID, &team.TeacherEmail, &team.Name, &team.Division, &team.InPerson, &team.DivisionExplanation, &team.SchoolName, &registrationTS, &teamWithTeacherName.TeacherName)
	team.RegistrationTS = time.UnixMilli(registrationTS)
	teamWithTeacherName.Team = &team
	return &teamWithTeacherName, err
}

func (d *Database) scanTeamStudents(ctx context.Context, team *Team) error {
	studentRows, err := d.DB.Query(ctx, `
		SELECT s.email, s.name, s.age, s.parentemail, s.signatory, s.previouslyparticipated, s.emailconfirmed,
			s.liabilitywaiver, s.computerusewaiver, s.campustour, s.dietaryrestrictions, s.qrcodesent, s.checkedin
		FROM students s
		WHERE s.teamid = ?
	`, team.ID)
	if err != nil {
		return err
	}
	defer studentRows.Close()
	for studentRows.Next() {
		var s Student
		var parentEmail, signatory, dietaryRestrictions sql.NullString
		var campusTour sql.NullBool
		if err := studentRows.Scan(&s.Email, &s.Name, &s.Age, &parentEmail, &signatory, &s.PreviouslyParticipated, &s.EmailConfirmed,
			&s.LiabilitySigned, &s.ComputerUseWaiverSigned, &campusTour, &dietaryRestrictions, &s.QRCodeSent, &s.CheckedIn); err != nil {
			return err
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
	return nil
}

func (d *Database) scanTeamWithStudents(ctx context.Context, row dbutil.Scannable) (*Team, error) {
	team, err := d.scanTeam(row)
	if err != nil {
		return nil, err
	}

	if err := d.scanTeamStudents(ctx, team); err != nil {
		return nil, err
	}

	return team, err
}

func (d *Database) scanTeamWithStudentsAndTeacherName(ctx context.Context, row dbutil.Scannable) (*TeamWithTeacherName, error) {
	teamWithTeacherName, err := d.scanTeamWithTeacherName(row)
	if err != nil {
		return nil, err
	}

	if err := d.scanTeamStudents(ctx, teamWithTeacherName.Team); err != nil {
		return nil, err
	}

	return teamWithTeacherName, err
}

func (d *Database) GetTeacherTeams(ctx context.Context, email string) ([]*Team, error) {
	rows, err := d.DB.Query(ctx, `
		SELECT t.id, t.teacheremail, t.name, t.division, t.inperson, t.divisionexplanation, tt.schoolname, t.registration_ts
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
		team, err := d.scanTeamWithStudents(ctx, rows)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}

	return teams, err
}

func (d *Database) GetAdminTeamsWithTeacherName(ctx context.Context) ([]*TeamWithTeacherName, error) {
	rows, err := d.DB.Query(ctx, `
		SELECT t.id, t.teacheremail, t.name, t.division, t.inperson, t.divisionexplanation, tt.schoolname, t.registration_ts, tt.name
		FROM teams t
		JOIN teachers tt ON tt.email = t.teacheremail
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*TeamWithTeacherName

	for rows.Next() {
		team, err := d.scanTeamWithStudentsAndTeacherName(ctx, rows)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}

	return teams, err
}

func (d *Database) GetTeam(ctx context.Context, email string, teamID uuid.UUID) (*Team, error) {
	row := d.DB.QueryRow(ctx, `
		SELECT t.id, t.teacheremail, t.name, t.division, t.inperson, t.divisionexplanation, tt.schoolname, t.registration_ts
		FROM teams t
		JOIN teachers tt ON tt.email = t.teacheremail
		WHERE tt.email = ?
		  AND t.id = ?
	`, email, teamID)
	return d.scanTeamWithStudents(ctx, row)
}

func (d *Database) GetTeamNoMembers(ctx context.Context, teamID uuid.UUID) (*Team, error) {
	row := d.DB.QueryRow(ctx, `
		SELECT t.id, t.teacheremail, t.name, t.division, t.inperson, t.divisionexplanation, '', t.registration_ts
		FROM teams t
		WHERE t.id = ?
	`, teamID)
	return d.scanTeam(row)
}

func (d *Database) UpsertTeam(ctx context.Context, teacherEmail string, teamID uuid.UUID, name string, division Division, inPerson bool, divisionExplanation string) error {
	_, err := d.DB.Exec(ctx, `
		INSERT OR REPLACE INTO teams (id, teacheremail, name, division, inperson, divisionexplanation, registration_ts)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, teamID, teacherEmail, name, division, inPerson, divisionExplanation, time.Now().UnixMilli())
	return err
}

func (d *Database) AddTeamMember(ctx context.Context, teamID uuid.UUID, name string, studentAge int, studentEmail string, previouslyParticipated bool) error {
	_, err := d.DB.Exec(ctx, `
		INSERT INTO students (teamid, name, age, email, previouslyparticipated)
		VALUES (?, ?, ?, ?, ?)
	`, teamID, name, studentAge, studentEmail, previouslyParticipated)
	return err
}

func (d *Database) RemoveTeamMember(ctx context.Context, teamID uuid.UUID, studentEmail string) error {
	res, err := d.DB.Exec(ctx, `
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
