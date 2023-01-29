package main

import (
	"net/http"
)

func (a *Application) GetTeacherTeamsTemplate(r *http.Request) map[string]any {
	user, err := a.GetLoggedInTeacher(r)
	if err != nil {
		a.Log.Error().Err(err).Msg("Failed to get logged in user")
		return nil
	}
	a.Log.Info().Interface("user", user).Msg("found user")

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
