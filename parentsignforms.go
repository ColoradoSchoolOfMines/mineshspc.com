package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/database"
)

func (a *Application) getStudentBySignFormsToken(tokenStr string) (*database.Student, error) {
	if tokenStr == "" {
		return nil, errors.New("no token")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return a.Config.ReadSecretKey(), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse sign forms token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !token.Valid || !ok {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	if claims.Issuer != string(IssuerSignForms) {
		return nil, fmt.Errorf("invalid sign forms token: %w", err)
	}

	return a.DB.GetStudentByEmail(claims.Subject)
}

func (a *Application) GetParentSignFormsTemplate(r *http.Request) map[string]any {
	tok := r.URL.Query().Get("tok")
	student, err := a.getStudentBySignFormsToken(tok)
	if err != nil {
		a.Log.Warn().Err(err).Msg("failed to get student from token")
		return nil
	}

	team, err := a.DB.GetTeamNoMembers(student.TeamID)
	if err != nil {
		a.Log.Error().Err(err).Msg("failed to get student's team")
		return nil
	}

	teacher, err := a.DB.GetTeacherByEmail(team.TeacherEmail)
	if err != nil {
		a.Log.Error().Err(err).Msg("failed to get teacher from DB")
		return nil
	}

	return map[string]any{
		"Accepted": student.ComputerUseWaiverSigned && student.MultimediaReleaseForm && student.LiabilitySigned,
		"Student":  student,
		"Teacher":  teacher,
		"Token":    tok,
	}
}

func (a *Application) HandleParentSignForms(w http.ResponseWriter, r *http.Request) {
	log := a.Log.With().Str("page_name", "sign_forms").Logger()
	tok := r.URL.Query().Get("tok")
	student, err := a.getStudentBySignFormsToken(tok)
	if err != nil {
		log.Error().Err(err).Msg("failed to get student from sign forms token")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Error().Err(err).Msg("failed to parse form")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !r.Form.Has("liability") || !r.Form.Has("technology-use") || !r.Form.Has("technology-resource-access") {
		log.Error().Msg("liability not signed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	parentName := r.Form.Get("parent-name")
	if parentName == "" {
		log.Error().Msg("parent name not provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = a.DB.SignFormsForStudent(student.Email, parentName); err != nil {
		log.Error().Err(err).Msg("failed to confirm student")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/register/parent/signforms?tok="+tok, http.StatusSeeOther)
}
