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

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

//go:embed emailtemplates/*
var emailTemplates embed.FS

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func (a *Application) GetTeacherCreateAccountTemplate(r *http.Request) map[string]any {
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
		log.Warn().Err(err).Msg("invalid captcha")
		a.TeacherCreateAccountRenderer(w, r, map[string]any{
			"Name":         name,
			"Email":        emailAddress,
			"CaptchaError": "Invalid captcha",
		})
		return
	}

	err = a.DB.NewTeacher(name, emailAddress)
	if errors.Is(err, sqlite3.ErrConstraintUnique) {
		log.Warn().Err(err).Msg("account already exists")
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

	tok := a.CreateEmailLoginJWT(emailAddress)
	signedTok, err := tok.SignedString(a.Config.ReadGetSecretKey())
	if err != nil {
		log.Error().Err(err).Msg("failed to sign email login token")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	templateData := map[string]any{
		"Name":       name,
		"ConfirmURL": fmt.Sprintf("%s/register/teacher/emaillogin?tok=%s", a.Config.Domain, signedTok),
	}

	var plainTextContent, htmlContent strings.Builder
	texttemplate.Must(texttemplate.ParseFS(emailTemplates, "emailtemplates/teachercreateaccount.txt")).Execute(&plainTextContent, templateData)
	htmltemplate.Must(htmltemplate.ParseFS(emailTemplates, "emailtemplates/teachercreateaccount.html")).Execute(&htmlContent, templateData)

	err = a.SendEmail(log, "Confirm Email to Log In to Mines HSPC Registration",
		mail.NewEmail(name, emailAddress),
		plainTextContent.String(),
		htmlContent.String())
	if err != nil {
		log.Error().Err(err).Msg("failed to send email")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		log.Info().Msg("successfully sent email")
		http.SetCookie(w, &http.Cookie{Name: "email", Value: emailAddress, Path: "/"})
		http.Redirect(w, r, "/register/teacher/confirmemail", http.StatusSeeOther)
	}
}

func (a *Application) HandleTeacherEmailLogin(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("tok")
	if tokenStr == "" {
		emailCookie, err := r.Cookie("email")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Just render the "check your email" page
		a.ConfirmEmailRenderer(w, r, map[string]any{"Email": emailCookie.Value})
		return
	}

	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return a.Config.ReadGetSecretKey(), nil
	})
	if err != nil {
		a.Log.Error().Err(err).Msg("failed to parse email login token")
		// TODO check if the error is that it is expired, and if so, then do
		// something nicer for the user.
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !token.Valid || !ok {
		a.Log.Error().Interface("token", token).Msg("failed to validate token")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if claims.Issuer != string(IssuerEmailLogin) {
		a.Log.Error().Interface("token", token).Msg("invalid token issuer, should be login")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	a.Log.Info().Str("sub", claims.Subject).Msg("confirmed email")
	err = a.DB.SetEmailConfirmed(claims.Subject)
	if err != nil {
		a.Log.Error().Err(err).Msg("failed to set email confirmed")
	}

	teacher, err := a.DB.GetTeacherByEmail(claims.Subject)
	if err != nil {
		a.Log.Error().Err(err).Msg("failed to get teacher from DB")
		return
	}

	jwt, expires := a.CreateSessionJWT(claims.Subject)
	jwtStr, err := jwt.SignedString(a.Config.ReadGetSecretKey())
	if err != nil {
		a.Log.Error().Err(err).Msg("failed to sign JWT")
		return
	}
	a.Log.Info().Interface("jwt", jwt).Str("jwt_str", jwtStr).Msg("signed JWT")
	http.SetCookie(w, &http.Cookie{Name: "tok", Value: jwtStr, Path: "/", Expires: expires})

	if teacher.SchoolName == "" || teacher.SchoolCity == "" || teacher.SchoolState == "" {
		http.Redirect(w, r, "/register/teacher/schoolinfo", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/register/teacher/teams", http.StatusSeeOther)
	}
}

func (a *Application) CreateSessionJWT(email string) (*jwt.Token, time.Time) {
	// TODO invent some way to make this a one-time token
	expires := time.Now().Add(24 * time.Hour)
	return jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer:    string(IssuerSessionToken),
		Subject:   email,
		ExpiresAt: jwt.NewNumericDate(expires),
	}), expires
}
