package main

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

func (a *Application) GetTeacherAddMemberTemplate(r *http.Request) map[string]any {
	user, err := a.GetLoggedInTeacher(r)
	if err != nil {
		a.Log.Error().Err(err).Msg("Failed to get logged in user")
		return nil
	}
	a.Log.Debug().Interface("user", user).Msg("found user")

	templateData := map[string]any{}

	teamID := r.URL.Query().Get("team_id")
	a.Log.Debug().Str("team_id", teamID).Msg("getting team")
	team, err := a.DB.GetTeam(user.Email, teamID)
	if err != nil {
		a.Log.Error().Err(err).Msg("Failed to get team")
		// TODO report this error to the user and email admin
		return nil
	}

	templateData["Team"] = team

	a.Log.Info().Interface("template_data", templateData).Msg("team edit template")

	return templateData
}

func (a *Application) HandleTeacherAddMember(w http.ResponseWriter, r *http.Request) {
	log := a.Log.With().Str("page_name", "teacher_add_member").Logger()
	user, err := a.GetLoggedInTeacher(r)
	if err != nil {
		// TODO indicate that they are logged out
		log.Error().Err(err).Msg("Failed to get logged in user")
		http.Redirect(w, r, "/register/teacher/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Error().Err(err).Msg("failed to parse form")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	studentName := r.Form.Get("student-name")
	studentAgeStr := r.Form.Get("student-age")
	studentEmail := r.Form.Get("student-email")
	previouslyParticipated := r.Form.Get("previously-participated") == "has"

	studentAge, err := strconv.Atoi(studentAgeStr)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse student age")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	parentEmail := r.Form.Get("parent-email")
	if studentAge < 18 && parentEmail == "" {
		log.Error().Err(err).Msg("parent email is required for students under 18")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	teamIDStr := r.URL.Query().Get("team_id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse team id")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Ensure the team exists and that the user is the owner
	team, err := a.DB.GetTeam(user.Email, teamIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get team")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(team.Members) >= 4 {
		log.Error().Err(err).Msg("Team already has 4 members")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	a.Log.Info().
		Str("student_name", studentName).
		Int("student_age", studentAge).
		Str("student_email", studentEmail).
		Str("parent_email", parentEmail).
		Str("team_id", teamIDStr).
		Msg("adding student")
	if err := a.DB.AddTeamMember(teamID, studentName, studentAge, studentEmail, previouslyParticipated, parentEmail); err != nil {
		log.Error().Err(err).Msg("Failed to add team member")
		// TODO report this error to the user and email admin
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/register/teacher/team/edit?team_id="+teamID.String(), http.StatusSeeOther)
}

func (a *Application) HandleTeacherDeleteMember(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	teamIDStr := r.URL.Query().Get("team_id")
	log := a.Log.With().
		Str("page_name", "teacher_delete_member").
		Str("team_id", teamIDStr).
		Str("email", email).
		Logger()
	user, err := a.GetLoggedInTeacher(r)
	if err != nil {
		// TODO indicate that they are logged out
		log.Error().Err(err).Msg("Failed to get logged in user")
		http.Redirect(w, r, "/register/teacher/login", http.StatusSeeOther)
		return
	}

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse team id")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Ensure the team exists and that the user is the owner
	_, err = a.DB.GetTeam(user.Email, teamIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get team")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := a.DB.RemoveTeamMember(teamID, email); err != nil {
		log.Error().Err(err).Msg("Failed to delete team member")
		// TODO report this error to the user and email admin
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/register/teacher/team/edit?team_id="+teamID.String(), http.StatusSeeOther)
}
