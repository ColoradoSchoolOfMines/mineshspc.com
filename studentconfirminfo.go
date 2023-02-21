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
		a.Log.Warn().Err(err).Msg("failed to get student's team")
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
	student, err := a.getStudentByToken(r.URL.Query().Get("tok"))

	a.Log.Debug().Interface("student", student).Err(err).Msg("student confirm email")
}
