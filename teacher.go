package main

import (
	"embed"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"math/rand"
	"net/http"
	"strings"
	texttemplate "text/template"
	"time"

	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

//go:embed emailtemplates/*
var emailTemplates embed.FS

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func (a *Application) GetTeacherRegistrationTemplate(r *http.Request) map[string]any {
	captchaElements := make([]string, 5)
	for i := range captchaElements {
		captchaElements[i] = string(alphabet[rand.Intn(len(alphabet))]) + string(alphabet[rand.Intn(len(alphabet))])
	}

	captchaAnswer := ""
	captchaIndexes := make([]int, 3)
	for i := range captchaIndexes {
		index := rand.Intn(5)
		captchaIndexes[i] = index
		captchaAnswer += captchaElements[index]
	}

	registrationID := uuid.New()
	a.TeacherRegistrationCaptchas[registrationID] = captchaAnswer
	a.Log.Info().
		Str("registration_id", registrationID.String()).
		Interface("captcha_elements", captchaElements).
		Interface("captcha_indexes", captchaIndexes).
		Interface("captcha_answer", captchaAnswer).
		Msg("created captcha for teacher registration")

	go func() {
		time.Sleep(24 * time.Hour)
		if _, ok := a.TeacherRegistrationCaptchas[registrationID]; ok {
			a.Log.Info().
				Str("registration_id", registrationID.String()).
				Msg("expiring registration")
			delete(a.TeacherRegistrationCaptchas, registrationID)
		}
	}()

	return map[string]any{
		"RegistrationID":  registrationID,
		"CaptchaElements": captchaElements,
		"CaptchaIndexes":  captchaIndexes,
	}
}

func (a *Application) HandleTeacherLogin(w http.ResponseWriter, r *http.Request) {
	// TODO implement this. Make sure to check that the email was confirmed
	// before allowing to avoid replay attacks.
}

func (a *Application) HandleTeacherCreateAccount(w http.ResponseWriter, r *http.Request) {
	log := a.Log.With().Str("page_name", "teacher_create_account").Logger()
	if err := r.ParseForm(); err != nil {
		log.Error().Err(err).Msg("failed to parse form")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	registrationIDString := r.Form.Get("registration-id")

	emailAddress := r.Form.Get("email-address")
	name := r.Form.Get("your-name")

	registrationID, err := uuid.Parse(registrationIDString)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse registration id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer delete(a.TeacherRegistrationCaptchas, registrationID)

	if captcha, ok := a.TeacherRegistrationCaptchas[registrationID]; !ok || captcha != r.Form.Get("captcha") {
		log.Error().Err(err).Msg("invalid captcha")
		a.TeacherCreateAccountRenderer(w, r, map[string]any{
			"Name":         name,
			"Email":        emailAddress,
			"CaptchaError": "Invalid captcha",
		})
		return
	}

	err = a.DB.NewTeacher(name, emailAddress)
	if errors.Is(err, sqlite3.ErrConstraintUnique) {
		log.Error().Err(err).Msg("account already exists")
		a.TeacherCreateAccountRenderer(w, r, map[string]any{
			"Name":        name,
			"Email":       emailAddress,
			"EmailExists": true,
		})
		return
	} else if err != nil {
		log.Error().Err(err).Msg("failed to create new teacher account")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	loginCode := uuid.New()
	a.LoginCodes[emailAddress] = loginCode
	a.Log.Info().Str("email", emailAddress).Interface("login_code", loginCode).Msg("created a login code")
	go func() {
		time.Sleep(time.Hour)
		if _, ok := a.LoginCodes[emailAddress]; ok {
			a.Log.Info().Str("email", emailAddress).Msg("expiring login code")
			delete(a.LoginCodes, emailAddress)
		}
	}()

	from := mail.NewEmail("Mines HSPC", "noreply@mineshspc.com")
	to := mail.NewEmail(name, emailAddress)
	subject := "Confirm Email to Log In to Mines HSPC Registration"

	templateData := map[string]any{
		"Name":       name,
		"ConfirmURL": fmt.Sprintf("https://mineshspc.com/register/teacher/confirmemail?login_code=%s", loginCode),
	}

	var plainTextContent, htmlContent strings.Builder
	texttemplate.Must(texttemplate.ParseFS(emailTemplates, "emailtemplates/teachercreateaccount.txt")).Execute(&plainTextContent, templateData)
	htmltemplate.Must(htmltemplate.ParseFS(emailTemplates, "emailtemplates/teachercreateaccount.html")).Execute(&htmlContent, templateData)

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
		http.Redirect(w, r, "/register/teacher/confirmemail", http.StatusSeeOther)
	}
}

func (a *Application) HandleTeacherEmailConfirmation(w http.ResponseWriter, r *http.Request) {
	emailCookie, err := r.Cookie("email")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// If there is no login code, then this was a redirect from the login/create account page.
	loginCode := r.URL.Query().Get("login_code")
	if loginCode == "" {
		a.ConfirmEmailRenderer(w, r, map[string]any{"Email": emailCookie.Value})
		return
	}

	// Verify the login code and give them a session.
	if loginCodeUUID, err := uuid.Parse(loginCode); err != nil {
		a.Log.Error().Err(err).Msg("failed to parse login code")
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if storedLoginCode, ok := a.LoginCodes[emailCookie.Value]; !ok || storedLoginCode != loginCodeUUID {
		a.Log.Error().Err(err).Msg("invalid login code")
		a.ConfirmEmailRenderer(w, r, map[string]any{
			"Error": "Invalid or expired login code",
		})
		return
	} else {
		a.Log.Info().Str("email", emailCookie.Value).Msg("confirmed email")
		err = a.DB.SetEmailConfirmed(emailCookie.Value)
		if err != nil {
			a.Log.Error().Err(err).Msg("failed to set email confirmed")
		}
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
	http.Redirect(w, r, "/teacher/register/schoolinfo", http.StatusSeeOther)
}

func (a *Application) GetTeacherSchoolInfoTemplate(r *http.Request) map[string]any {
	user, err := a.GetLoggedInTeacher(r)
	if err != nil {
		a.Log.Error().Err(err).Msg("Failed to get logged in user")
		return nil
	}
	a.Log.Info().Interface("user", user).Msg("found user")

	validated := ""
	if user.SchoolName != "" || user.SchoolCity != "" || user.SchoolState != "" {
		validated = "validated"
	}
	return map[string]any{
		"Username":    user.Name,
		"Validated":   validated,
		"SchoolName":  user.SchoolName,
		"SchoolCity":  user.SchoolCity,
		"SchoolState": user.SchoolState,
	}
}

func (a *Application) HandleTeacherSchoolInfo(w http.ResponseWriter, r *http.Request) {
	log := a.Log.With().Str("page_name", "teacher_school_info").Logger()
	user, err := a.GetLoggedInTeacher(r)
	if err != nil {
		// TODO indicate that they are logged out
		http.Redirect(w, r, "/register/teacher/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Error().Err(err).Msg("failed to parse form")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	schoolName := r.Form.Get("school-name")
	schoolCity := r.Form.Get("school-city")
	schoolState := r.Form.Get("school-state")

	if schoolName == "" || schoolCity == "" || schoolState == "" {
		a.ConfirmEmailRenderer(w, r, map[string]any{
			"Errors": map[string]any{
				"school-name":  schoolName == "",
				"school-city":  schoolCity == "",
				"school-state": schoolState == "",
			},
		})
		return
	}

	err = a.DB.SetTeacherSchoolInfo(user.Email, schoolName, schoolCity, schoolState)
	if err != nil {
		a.ConfirmEmailRenderer(w, r, map[string]any{
			"Errors": map[string]any{
				"general": "Failed to save school info. Please try again.",
			},
		})
		return
	}

	http.Redirect(w, r, "/teacher/register/teams", http.StatusSeeOther)
}
