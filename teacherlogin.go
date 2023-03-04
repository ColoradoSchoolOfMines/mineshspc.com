package main

import (
	"fmt"
	htmltemplate "html/template"
	"net/http"
	"strings"
	texttemplate "text/template"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Issuer string

const (
	IssuerEmailLogin    Issuer = "email_login"
	IssuerSessionToken  Issuer = "session_token"
	IssuerStudentVerify Issuer = "student_verify"
	IssuerSignForms     Issuer = "sign_forms"
)

func (a *Application) GetEmailLoginTemplate(r *http.Request) map[string]any {
	emailCookie, err := r.Cookie("email")
	if err != nil {
		return nil
	}
	return map[string]any{
		"Email": emailCookie.Value,
	}
}

func (a *Application) CreateEmailLoginJWT(email string) *jwt.Token {
	// TODO invent some way to make this a one-time token
	return jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer:    string(IssuerEmailLogin),
		Subject:   email,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})
}

func (a *Application) HandleTeacherLogin(w http.ResponseWriter, r *http.Request) {
	log := a.Log.With().Str("page_name", "teacher_create_account").Logger()
	if err := r.ParseForm(); err != nil {
		log.Error().Err(err).Msg("failed to parse form")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	emailAddress := r.FormValue("email-address")
	if emailAddress == "" {
		a.Log.Warn().Msg("no email address provided in request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log = log.With().Str("email", emailAddress).Logger()

	teacher, err := a.DB.GetTeacherByEmail(emailAddress)
	if err != nil {
		log.Warn().Err(err).Msg("failed to find teacher by email")
		a.TeacherLoginRenderer(w, r, map[string]any{
			"Email":         emailAddress,
			"EmailNotFound": true,
		})
		return
	} else if !teacher.EmailConfirmed {
		log.Warn().Err(err).Msg("teacher email not confirmed, not sending login code to avoid amplification attacks")
		a.TeacherLoginRenderer(w, r, map[string]any{
			"Email":             emailAddress,
			"EmailNotConfirmed": true,
		})
		return
	}

	tok := a.CreateEmailLoginJWT(emailAddress)
	signedTok, err := tok.SignedString(a.Config.ReadSecretKey())
	if err != nil {
		log.Error().Err(err).Msg("failed to sign email login token")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	templateData := map[string]any{
		"Name":     teacher.Name,
		"LoginURL": fmt.Sprintf("%s/register/teacher/emaillogin?tok=%s", a.Config.Domain, signedTok),
	}

	var plainTextContent, htmlContent strings.Builder
	texttemplate.Must(texttemplate.ParseFS(emailTemplates, "emailtemplates/teacherlogin.txt")).Execute(&plainTextContent, templateData)
	htmltemplate.Must(htmltemplate.ParseFS(emailTemplates, "emailtemplates/teacherlogin.html")).Execute(&htmlContent, templateData)

	err = a.SendEmail(log, "Log in to Mines HSPC Registration",
		mail.NewEmail(teacher.Name, emailAddress),
		plainTextContent.String(),
		htmlContent.String())
	if err != nil {
		log.Error().Err(err).Msg("failed to send email")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		log.Info().Msg("sent email")
		http.SetCookie(w, &http.Cookie{Name: "email", Value: emailAddress, Path: "/"})
		http.Redirect(w, r, "/register/teacher/emaillogin", http.StatusSeeOther)
	}
}
