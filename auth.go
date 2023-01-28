package main

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/database"
)

func (a *Application) GetLoggedInTeacher(r *http.Request) (*database.Teacher, error) {
	sessionTokenStr, err := r.Cookie("session_id")
	if err != nil {
		a.Log.Warn().Err(err).Msg("Failed to get session cookie")
		return nil, err
	}
	sessionToken, err := uuid.Parse(sessionTokenStr.Value)
	if err != nil {
		a.Log.Warn().Err(err).Msg("Failed to parse session cookie")
		return nil, err
	}
	user, err := a.DB.GetTeacherBySessionToken(sessionToken)
	if err != nil {
		a.Log.Warn().Err(err).
			Str("token", sessionToken.String()).
			Msg("couldn't find teacher with that session token")
		return nil, err
	}

	return user, nil
}
