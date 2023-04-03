package main

import (
	"context"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	texttemplate "text/template"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func (a *Application) GetTeacherAddMemberTemplate(r *http.Request) map[string]any {
	user, err := a.GetLoggedInTeacher(r)
	if err != nil {
		a.Log.Warn().Err(err).Msg("Failed to get logged in user")
		return nil
	}
	a.Log.Debug().Interface("user", user).Msg("found user")

	templateData := map[string]any{}

	teamIDStr := r.URL.Query().Get("team_id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		a.Log.Warn().Err(err).Msg("Failed to parse team id")
		return nil
	}
	a.Log.Debug().Str("team_id", teamIDStr).Msg("getting team")
	team, err := a.DB.GetTeam(r.Context(), user.Email, teamID)
	if err != nil {
		a.Log.Error().Err(err).Msg("Failed to get team")
		// TODO report this error to the user and email admin
		return nil
	}

	templateData["Team"] = team

	a.Log.Info().Interface("template_data", templateData).Msg("team edit template")

	return templateData
}

func (a *Application) sendStudentEmail(ctx context.Context, studentEmail, studentName, teacherName, teamName string) error {
	log := zerolog.Ctx(ctx).With().Str("action", "sendStudentEmail").Logger()

	tok := a.CreateStudentVerifyJWT(studentEmail)
	signedTok, err := tok.SignedString(a.Config.ReadSecretKey())
	if err != nil {
		log.Error().Err(err).Msg("Failed to create student verify url")
		return err
	}
	templateData := map[string]any{
		"Name":        studentName,
		"TeacherName": teacherName,
		"TeamName":    teamName,
		"VerifyURL":   fmt.Sprintf("%s/register/student/confirminfo?tok=%s", a.Config.Domain, signedTok),
	}

	var plainTextContent, htmlContent strings.Builder
	texttemplate.Must(texttemplate.ParseFS(emailTemplates, "emailtemplates/studentverify.txt")).Execute(&plainTextContent, templateData)
	htmltemplate.Must(htmltemplate.ParseFS(emailTemplates, "emailtemplates/studentverify.html")).Execute(&htmlContent, templateData)

	err = a.SendEmail(log, "Confirm Mines HSPC Registration",
		mail.NewEmail(studentName, studentEmail),
		plainTextContent.String(),
		htmlContent.String())
	if err != nil {
		log.Error().Err(err).Msg("failed to send email")
		return err
	}
	log.Info().Msg("sent email")
	return nil
}

var ageRegex = regexp.MustCompile(`^(\d+)$`)

func (a *Application) HandleTeacherAddMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := a.Log.With().Str("page_name", "teacher_add_member").Logger()
	user, err := a.GetLoggedInTeacher(r)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get logged in user")
		http.Redirect(w, r, "/register/teacher/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Error().Err(err).Msg("failed to parse form")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	studentName := r.FormValue("student-name")
	studentAgeStr := r.FormValue("student-age")
	studentEmail := r.FormValue("student-email")
	previouslyParticipated := r.FormValue("previously-participated") == "has"

	if !ageRegex.MatchString(studentAgeStr) {
		a.TeamAddMemberRenderer(w, r, map[string]any{
			"Error": map[string]any{
				"General": htmltemplate.HTML("Please enter an integer age without decimal places."),
			},
			"StudentName":            studentName,
			"StudentEmail":           studentEmail,
			"PreviouslyParticipated": previouslyParticipated,
		})
		return
	}

	studentAge, err := strconv.Atoi(studentAgeStr)
	if err != nil {
		log.Warn().Err(err).Msg("failed to parse student age")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if user.EmailAllowance <= 0 {
		log.Warn().Msg("User has no email allowance")
		a.TeamAddMemberRenderer(w, r, map[string]any{
			"Error": map[string]any{
				"General": htmltemplate.HTML(
					`You have reached your quota for sent emails. Please email
					<a href="mailto:support@mineshspc.com">support@mineshspc.com</a>
					if you need to add more members to any of your teams.`),
			},
			"StudentName":            studentName,
			"StudentAge":             studentAge,
			"StudentEmail":           studentEmail,
			"PreviouslyParticipated": previouslyParticipated,
		})
		return
	}

	err = a.DB.DecrementEmailAllowance(ctx, user.Email)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decrement email allowance")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	teamIDStr := r.URL.Query().Get("team_id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to parse team id")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Ensure the team exists and that the user is the owner
	team, err := a.DB.GetTeam(ctx, user.Email, teamID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get team")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(team.Members) >= 4 {
		log.Warn().Err(err).Msg("Team already has 4 members")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log = log.With().
		Str("student_name", studentName).
		Int("student_age", studentAge).
		Str("student_email", studentEmail).
		Bool("previously_participated", previouslyParticipated).
		Str("team_id", teamIDStr).
		Logger()
	log.Info().Msg("adding student")
	if err := a.DB.AddTeamMember(ctx, teamID, studentName, studentAge, studentEmail, previouslyParticipated); err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr); sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique || sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
			a.TeamAddMemberRenderer(w, r, map[string]any{
				"Error": map[string]any{
					"General": "That email address has already added to a team.",
				},
				"StudentName":            studentName,
				"StudentAge":             studentAge,
				"StudentEmail":           studentEmail,
				"PreviouslyParticipated": previouslyParticipated,
			})
			return
		}

		log.Error().Err(err).Msg("Failed to add team member")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send email to student
	if err := a.sendStudentEmail(ctx, studentEmail, studentName, user.Name, team.Name); err != nil {
		log.Error().Err(err).Msg("failed to send student email")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/register/teacher/team/edit?team_id="+teamID.String(), http.StatusSeeOther)
}

func (a *Application) CreateStudentVerifyJWT(email string) *jwt.Token {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer:  string(IssuerStudentVerify),
		Subject: email,
	})
}

func (a *Application) HandleTeacherDeleteMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !a.Config.RegistrationEnabled {
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	email := r.URL.Query().Get("email")
	teamIDStr := r.URL.Query().Get("team_id")
	log := a.Log.With().
		Str("page_name", "teacher_delete_member").
		Str("team_id", teamIDStr).
		Str("email", email).
		Logger()
	user, err := a.GetLoggedInTeacher(r)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get logged in user")
		http.Redirect(w, r, "/register/teacher/login", http.StatusSeeOther)
		return
	}

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to parse team id")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Ensure the team exists and that the user is the owner
	_, err = a.DB.GetTeam(ctx, user.Email, teamID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get team")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := a.DB.RemoveTeamMember(ctx, teamID, email); err != nil {
		log.Error().Err(err).Msg("Failed to delete team member")
		// TODO report this error to the user and email admin
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/register/teacher/team/edit?team_id="+teamID.String(), http.StatusSeeOther)
}
