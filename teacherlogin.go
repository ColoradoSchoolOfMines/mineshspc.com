package main

import (
	"fmt"
	htmltemplate "html/template"
	"net/http"
	"strings"
	texttemplate "text/template"
	"time"

	"github.com/google/uuid"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func (a *Application) HandleTeacherLogin(w http.ResponseWriter, r *http.Request) {
	log := a.Log.With().Str("page_name", "teacher_create_account").Logger()
	if err := r.ParseForm(); err != nil {
		log.Error().Err(err).Msg("failed to parse form")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	emailAddress := r.Form.Get("email-address")
	if emailAddress == "" {
		a.Log.Error().Msg("no email address provided in request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log = log.With().Str("email", emailAddress).Logger()

	teacher, err := a.DB.GetTeacherByEmail(emailAddress)
	if err != nil {
		log.Error().Err(err).Msg("failed to find teacher by email")
		a.TeacherLoginRenderer(w, r, map[string]any{
			"Email":         emailAddress,
			"EmailNotFound": true,
		})
		return
	} else if !teacher.EmailConfirmed {
		log.Error().Err(err).Msg("teacher email not confirmed, not sending login code to avoid mirror attacks")
		a.TeacherLoginRenderer(w, r, map[string]any{
			"Email":             emailAddress,
			"EmailNotConfirmed": true,
		})
		return
	}

	from := mail.NewEmail("Mines HSPC", "noreply@mineshspc.com")
	to := mail.NewEmail(teacher.Name, emailAddress)
	subject := "Log In to Mines HSPC Registration"

	templateData := map[string]any{
		"Name":     teacher.Name,
		"LoginURL": fmt.Sprintf("https://mineshspc.com/register/teacher/emaillogin?login_code=%s", a.CreateLoginCode(emailAddress)),
	}

	var plainTextContent, htmlContent strings.Builder
	texttemplate.Must(texttemplate.ParseFS(emailTemplates, "emailtemplates/teacherlogin.txt")).Execute(&plainTextContent, templateData)
	htmltemplate.Must(htmltemplate.ParseFS(emailTemplates, "emailtemplates/teacherlogin.html")).Execute(&htmlContent, templateData)

	message := mail.NewSingleEmail(from, subject, to, plainTextContent.String(), htmlContent.String())
	message.ReplyTo = mail.NewEmail("Mines HSPC Team", "team@mineshspc.com")
	resp, err := a.SendGridClient.Send(message)
	if err != nil {
		log.Error().Err(err).Msg("failed to send email")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if resp.StatusCode != http.StatusAccepted {
		log.Error().
			Int("status_code", resp.StatusCode).
			Str("to", emailAddress).
			Str("response_body", resp.Body).
			Msg("error sending email")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		log.Info().
			Int("status_code", resp.StatusCode).
			Str("to", emailAddress).
			Msg("successfully sent email")
		http.SetCookie(w, &http.Cookie{Name: "email", Value: emailAddress, Path: "/"})
		http.Redirect(w, r, "/register/teacher/emaillogin", http.StatusSeeOther)
	}
}

func (a *Application) HandleEmailLogin(w http.ResponseWriter, r *http.Request) {
	emailCookie, err := r.Cookie("email")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// If there is no login code, then this was a redirect from the login/create account page.
	loginCode := r.URL.Query().Get("login_code")
	if loginCode == "" {
		a.EmailLoginRenderer(w, r, map[string]any{"Email": emailCookie.Value})
		return
	}

	// Verify the login code and give them a session.
	if loginCodeUUID, err := uuid.Parse(loginCode); err != nil {
		a.Log.Error().Err(err).Msg("failed to parse login code")
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if storedLoginCode, ok := a.LoginCodes[emailCookie.Value]; !ok || storedLoginCode != loginCodeUUID {
		a.Log.Error().Err(err).Msg("invalid login code")
		a.EmailLoginRenderer(w, r, map[string]any{
			"Error": "Invalid or expired login code",
		})
		return
	} else {
		delete(a.LoginCodes, emailCookie.Value)
	}

	sessionID := uuid.New()
	expires := time.Now().AddDate(0, 0, 1)
	err = a.DB.NewTeacherSession(emailCookie.Value, sessionID, expires)
	if err != nil {
		a.Log.Error().Err(err).Msg("failed to create new teacher session")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "session_id", Value: sessionID.String(), Path: "/", Expires: expires})

	teacher, err := a.DB.GetTeacherByEmail(emailCookie.Value)
	if err != nil {
		a.Log.Error().Err(err).Msg("failed to get teacher from DB")
		return
	}

	if teacher.SchoolName == "" || teacher.SchoolCity == "" || teacher.SchoolState == "" {
		http.Redirect(w, r, "/register/teacher/schoolinfo", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/register/teacher/teams", http.StatusSeeOther)
	}
}
