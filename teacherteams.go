package main

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/database"
)

func (a *Application) GetTeacherTeamsTemplate(r *http.Request) map[string]any {
	user, err := a.GetLoggedInTeacher(r)
	if err != nil {
		a.Log.Error().Err(err).Msg("Failed to get logged in user")
		return nil
	}
	a.Log.Debug().Interface("user", user).Msg("found user")

	teams, err := a.DB.GetTeacherTeams(user.Email)
	if err != nil {
		a.Log.Error().Err(err).Msg("Failed to get teacher teams")
		// TODO report this error to the user and email admin
		return nil
	}

	return map[string]any{
		"Username":    user.Name,
		"SchoolName":  user.SchoolName,
		"SchoolCity":  user.SchoolCity,
		"SchoolState": user.SchoolState,
		"Teams":       teams,
	}
}

func (a *Application) GetTeacherTeamEditTemplate(r *http.Request) map[string]any {
	user, err := a.GetLoggedInTeacher(r)
	if err != nil {
		a.Log.Error().Err(err).Msg("Failed to get logged in user")
		return nil
	}
	a.Log.Debug().Interface("user", user).Msg("found user")

	templateData := map[string]any{
		"Username":    user.Name,
		"SchoolName":  user.SchoolName,
		"SchoolCity":  user.SchoolCity,
		"SchoolState": user.SchoolState,
	}

	teamID := r.URL.Query().Get("team_id")
	a.Log.Debug().Str("team_id", teamID).Msg("getting team")
	if teamID != "" {
		team, err := a.DB.GetTeam(user.Email, teamID)
		if err != nil {
			a.Log.Error().Err(err).Msg("Failed to get teacher teams")
			// TODO report this error to the user and email admin
			return nil
		}

		templateData["Team"] = team
	}

	a.Log.Info().Interface("template_data", templateData).Msg("team edit template")

	return templateData
}

func (a *Application) HandleTeacherTeamEdit(w http.ResponseWriter, r *http.Request) {
	log := a.Log.With().Str("page_name", "teacher_team_edit").Logger()
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

	teamName := r.Form.Get("team-name")
	inPerson := r.Form.Get("team-location") == "in-person"
	teamDivision, err := database.ParseDivision(r.Form.Get("team-division"))
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse team division")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	teamDivisionExplanation := r.Form.Get("team-division-explanation")

	teamIDStr := r.URL.Query().Get("team_id")
	var teamID uuid.UUID
	if teamIDStr == "" {
		// Create a team
		teamID = uuid.New()
	} else {
		// Update the team
		teamID, err = uuid.Parse(teamIDStr)
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse team id")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Verify that the in-person-ness of the team did not change
		team, err := a.DB.GetTeam(user.Email, teamIDStr)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get team")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if team.InPerson != inPerson {
			log.Error().Err(err).Msg("Cannot change in-person status of team")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if err := a.DB.UpsertTeam(user.Email, teamID, teamName, teamDivision, inPerson, teamDivisionExplanation); err != nil {
		log.Error().Err(err).Msg("Failed to upsert team")
		// TODO report this error to the user and email admin
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/register/teacher/team/edit?team_id="+teamID.String(), http.StatusSeeOther)
}
