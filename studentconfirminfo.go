package main

import (
	"errors"
	"fmt"
	htmltemplate "html/template"
	"net/http"
	"strings"
	texttemplate "text/template"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/database"
)

func (a *Application) getStudentByToken(tokenStr string) (*database.Student, error) {
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

func (a *Application) CreateSignFormsJWT(email string) *jwt.Token {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer:  string(IssuerSignForms),
		Subject: email,
	})
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

	sendEmail := false

	log.Info().Interface("student", student).Msg("confirming email")

	if !student.EmailConfirmed {
		if r.Form.Has("confirm-info-correct") {
			student.EmailConfirmed = true
		} else {
			log.Warn().Msg("student did not confirm info")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if student.Age < 18 && student.ParentEmail == "" {
			parentEmail := r.FormValue("parent-email")
			if parentEmail == "" {
				log.Warn().Err(err).Msg("parent email is required for students under 18")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			student.ParentEmail = parentEmail
		}
		sendEmail = true
	}

	log.Info().Interface("send_email", sendEmail).Interface("student", student).Msg("done confirming")

	if team.InPerson {
		student.CampusTour = r.Form.Has("campus-tour")
		student.DietaryRestrictions = r.FormValue("dietary-restrictions")
	}

	if err = a.DB.ConfirmStudent(student.Email, student.CampusTour, student.DietaryRestrictions, student.ParentEmail); err != nil {
		log.Error().Err(err).Msg("failed to confirm student")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Info().Interface("s", student).Msg("student confirmed")

	if sendEmail {
		toAddress := student.ParentEmail
		if student.Age >= 18 {
			toAddress = student.Email
		}

		tok := a.CreateSignFormsJWT(student.Email)
		signedTok, err := tok.SignedString(a.Config.ReadSecretKey())
		if err != nil {
			log.Error().Err(err).Msg("failed to sign email login token")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		templateData := map[string]any{
			"Student": student,
			"SignURL": fmt.Sprintf("%s/register/parent/signforms?tok=%s", a.Config.Domain, signedTok),
		}

		var plainTextContent, htmlContent strings.Builder
		texttemplate.Must(texttemplate.ParseFS(emailTemplates, "emailtemplates/forms.txt")).Execute(&plainTextContent, templateData)
		htmltemplate.Must(htmltemplate.ParseFS(emailTemplates, "emailtemplates/forms.html")).Execute(&htmlContent, templateData)

		err = a.SendEmail(log, "Sign forms to participate in Mines HSPC",
			mail.NewEmail("", toAddress),
			plainTextContent.String(),
			htmlContent.String())
		if err != nil {
			log.Error().Err(err).Msg("failed to send email")
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			log.Info().Msg("successfully sent email")
		}
	}

	a.StudentConfirmInfoRenderer(w, r, map[string]any{
		"Confirmed": student.EmailConfirmed,
		"Student":   student,
		"Team":      team,
		"Token":     tok,
	})
}
