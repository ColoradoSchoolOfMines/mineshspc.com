package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/database"
)

func (a *Application) getStudentBySignFormsToken(ctx context.Context, tokenStr string) (*database.Student, error) {
	if tokenStr == "" {
		return nil, errors.New("no token")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
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

	return a.DB.GetStudentByEmail(ctx, claims.Subject)
}

func (a *Application) GetParentSignFormsTemplate(r *http.Request) map[string]any {
	ctx := r.Context()
	tok := r.URL.Query().Get("tok")
	student, err := a.getStudentBySignFormsToken(ctx, tok)
	if err != nil {
		a.Log.Warn().Err(err).Msg("failed to get student from token")
		return nil
	}

	team, err := a.DB.GetTeamNoMembers(ctx, student.TeamID)
	if err != nil {
		a.Log.Err(err).Msg("failed to get student's team")
		return nil
	}

	teacher, err := a.DB.GetTeacherByEmail(ctx, team.TeacherEmail)
	if err != nil {
		a.Log.Err(err).Msg("failed to get teacher from DB")
		return nil
	}

	accepted := student.LiabilitySigned
	if team.InPerson {
		accepted = accepted && student.ComputerUseWaiverSigned
	}

	return map[string]any{
		"Accepted":               accepted,
		"Student":                student,
		"Teacher":                teacher,
		"NeedsComputerUseWaiver": team.InPerson,
		"Token":                  tok,
	}
}

func (a *Application) HandleParentSignForms(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := a.Log.With().Str("page_name", "sign_forms").Logger()
	tok := r.URL.Query().Get("tok")
	student, err := a.getStudentBySignFormsToken(ctx, tok)
	if err != nil {
		log.Warn().Err(err).Msg("failed to get student from sign forms token")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	team, err := a.DB.GetTeamNoMembers(ctx, student.TeamID)
	if err != nil {
		a.Log.Err(err).Msg("failed to get student's team")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log = log.With().Str("student_email", student.Email).Logger()

	if err := r.ParseForm(); err != nil {
		log.Err(err).Msg("failed to parse form")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !r.Form.Has("liability-waiver") {
		log.Warn().Msg("liability form not accepted")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if team.InPerson && !r.Form.Has("computer-use-waiver") {
		log.Warn().Msg("liability form not accepted")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	parentName := r.Form.Get("parent-name")
	if parentName == "" {
		log.Warn().Msg("parent name not provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = a.DB.SignFormsForStudent(ctx, student.Email, parentName, team.InPerson); err != nil {
		log.Err(err).Msg("failed to sign forms for student")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Info().Any("student", student).Msg("signed forms for student")

	http.Redirect(w, r, "/register/parent/signforms?tok="+tok, http.StatusSeeOther)
}
