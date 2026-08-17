package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mau.fi/util/dbutil"
	"go.mau.fi/util/exerrors"
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
	TeamID                  uuid.UUID `column:"teamid"`
	Email                   string    `column:"email"`
	Name                    string    `column:"name"`
	Age                     int       `column:"age"`
	ParentEmail             string    `column:"parentemail"`
	Signatory               string    `column:"signatory"`
	PreviouslyParticipated  bool      `column:"previouslyparticipated"`
	EmailConfirmed          bool      `column:"emailconfirmed"`
	LiabilitySigned         bool      `column:"liabilitywaiver"`
	ComputerUseWaiverSigned bool      `column:"computerusewaiver"`

	CampusTour          bool   `column:"campustour"`
	DietaryRestrictions string `column:"dietaryrestrictions"`

	QRCodeSent bool `column:"qrcodesent"`
	CheckedIn  bool `column:"checkedin"`
}

const studentSelectColumns = `
	s.teamid, s.email, s.name, s.age, COALESCE(s.parentemail, ''), COALESCE(s.signatory, ''),
	s.previouslyparticipated, s.emailconfirmed, s.liabilitywaiver, s.computerusewaiver,
	COALESCE(s.campustour, false), COALESCE(s.dietaryrestrictions, ''), s.qrcodesent, s.checkedin
`

var studentColumns = []string{
	"teamid", "email", "name", "age", "parentemail", "signatory",
	"previouslyparticipated", "emailconfirmed", "liabilitywaiver", "computerusewaiver",
	"campustour", "dietaryrestrictions", "qrcodesent", "checkedin",
}
var scanStudentRow = exerrors.Must(dbutil.MakeReflectScanner[Student](studentColumns, dbutil.ReflectScanOptions{StructTag: "column"}))

func (s *Student) Scan(row dbutil.Scannable) (*Student, error) {
	return scanStudentRow(row)
}

func scanTeamRow(row dbutil.Scannable) (*Team, error) {
	var team Team
	var registrationTS int64
	err := row.Scan(&team.ID, &team.TeacherEmail, &team.Name, &team.Division, &team.InPerson, &team.DivisionExplanation, &team.SchoolName, &registrationTS)
	team.RegistrationTS = time.UnixMilli(registrationTS)
	return &team, err
}

func (t *Team) Scan(row dbutil.Scannable) (*Team, error) {
	return scanTeamRow(row)
}

func scanTeamWithTeacherNameRow(row dbutil.Scannable) (*TeamWithTeacherName, error) {
	var team Team
	var teamWithTeacherName TeamWithTeacherName
	var registrationTS int64
	err := row.Scan(&team.ID, &team.TeacherEmail, &team.Name, &team.Division, &team.InPerson, &team.DivisionExplanation, &team.SchoolName, &registrationTS, &teamWithTeacherName.TeacherName)
	team.RegistrationTS = time.UnixMilli(registrationTS)
	teamWithTeacherName.Team = &team
	return &teamWithTeacherName, err
}

func (twtn *TeamWithTeacherName) Scan(row dbutil.Scannable) (*TeamWithTeacherName, error) {
	return scanTeamWithTeacherNameRow(row)
}

func (d *Database) scanTeamStudents(ctx context.Context, team *Team) error {
	studentRows, err := d.DB.Query(ctx, `
		SELECT `+studentSelectColumns+`
		FROM students s
		WHERE s.teamid = ?
	`, team.ID)
	if err != nil {
		return err
	}
	defer studentRows.Close()
	for studentRows.Next() {
		s, err := scanStudentRow(studentRows)
		if err != nil {
			return err
		}
		team.Members = append(team.Members, *s)
	}
	return studentRows.Err()
}

func (d *Database) GetTeacherTeams(ctx context.Context, email string) ([]*Team, error) {
	teams, err := d.teamQH.QueryMany(ctx, `
		SELECT t.id, t.teacheremail, t.name, t.division, t.inperson, t.divisionexplanation, tt.schoolname, t.registration_ts
		FROM teams t
		JOIN teachers tt ON tt.email = t.teacheremail
		WHERE tt.email = ?
	`, email)
	if err != nil {
		return nil, err
	}
	for _, team := range teams {
		if err := d.scanTeamStudents(ctx, team); err != nil {
			return nil, err
		}
	}
	return teams, nil
}

func (d *Database) GetAdminTeamsWithTeacherName(ctx context.Context) ([]*TeamWithTeacherName, error) {
	teams, err := d.teamWithTeacherNameQH.QueryMany(ctx, `
		SELECT t.id, t.teacheremail, t.name, t.division, t.inperson, t.divisionexplanation, tt.schoolname, t.registration_ts, tt.name
		FROM teams t
		JOIN teachers tt ON tt.email = t.teacheremail
	`)
	if err != nil {
		return nil, err
	}
	for _, team := range teams {
		if err := d.scanTeamStudents(ctx, team.Team); err != nil {
			return nil, err
		}
	}
	return teams, nil
}

func (d *Database) GetTeam(ctx context.Context, email string, teamID uuid.UUID) (*Team, error) {
	team, err := d.teamQH.QueryOne(ctx, `
		SELECT t.id, t.teacheremail, t.name, t.division, t.inperson, t.divisionexplanation, tt.schoolname, t.registration_ts
		FROM teams t
		JOIN teachers tt ON tt.email = t.teacheremail
		WHERE tt.email = ?
		  AND t.id = ?
	`, email, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, sql.ErrNoRows
	}
	if err := d.scanTeamStudents(ctx, team); err != nil {
		return nil, err
	}
	return team, nil
}

func (d *Database) GetTeamNoMembers(ctx context.Context, teamID uuid.UUID) (*Team, error) {
	team, err := d.teamQH.QueryOne(ctx, `
		SELECT t.id, t.teacheremail, t.name, t.division, t.inperson, t.divisionexplanation, '', t.registration_ts
		FROM teams t
		WHERE t.id = ?
	`, teamID)
	if err == nil && team == nil {
		return nil, sql.ErrNoRows
	}
	return team, err
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
