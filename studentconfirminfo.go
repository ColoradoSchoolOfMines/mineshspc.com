package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/database"
)

func (a *Application) getStudentByToken(tokenStr string) (*database.Student, error) {
	if tokenStr == "" {
		return nil, errors.New("no token")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return a.Config.ReadGetSecretKey(), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse student confirmation token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !token.Valid || !ok {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	if claims.Issuer != string(IssuerStudentVerify) {
		return nil, fmt.Errorf("invalid student verify token: %w", err)
	}

	return a.DB.GetStudentByEmail(claims.Subject)
}

func (a *Application) GetStudentConfirmInfoTemplate(r *http.Request) map[string]any {
	tok := r.URL.Query().Get("tok")
	student, err := a.getStudentByToken(tok)
	if err != nil {
		a.Log.Warn().Err(err).Msg("failed to get student from token")
		return nil
	}

	team, err := a.DB.GetTeamNoMembers(student.TeamID)
	if err != nil {
		a.Log.Error().Err(err).Msg("failed to get student's team")
		return nil
	}

	return map[string]any{
		"Confirmed": student.EmailConfirmed,
		"Student":   student,
		"Team":      team,
		"Token":     tok,
	}
}

func (a *Application) HandleStudentConfirmEmail(w http.ResponseWriter, r *http.Request) {
	log := a.Log.With().Str("page_name", "student_confirm_email").Logger()
	tok := r.URL.Query().Get("tok")
	student, err := a.getStudentByToken(tok)
	if err != nil {
		log.Warn().Err(err).Msg("failed to get student from token")
		a.StudentConfirmInfoRenderer(w, r, map[string]any{
			"Error": "Failed to find any such student.",
		})
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Error().Err(err).Msg("failed to parse form")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	team, err := a.DB.GetTeamNoMembers(student.TeamID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get student's team")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Form.Has("confirm-info-correct") {
		student.EmailConfirmed = true
	}

	if team.InPerson {
		student.CampusTour = r.Form.Has("campus-tour")
		student.DietaryRestrictions = r.FormValue("dietary-restrictions")
	}

	if err = a.DB.ConfirmStudent(student.Email, student.CampusTour, student.DietaryRestrictions); err != nil {
		log.Error().Err(err).Msg("failed to confirm student")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Info().Interface("s", student).Msg("student confirmed")

	a.StudentConfirmInfoRenderer(w, r, map[string]any{
		"Confirmed": student.EmailConfirmed,
		"Student":   student,
		"Team":      team,
		"Token":     tok,
	})
}
