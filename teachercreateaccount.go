package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"net/http"
	"net/url"
	"strings"
	texttemplate "text/template"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

//go:embed emailtemplates/*
var emailTemplates embed.FS

func (a *Application) GetTeacherCreateAccountTemplate(r *http.Request) map[string]any {
	return map[string]any{
		"ReCaptchaSiteKey": a.Config.Recaptcha.SiteKey,
	}
}

type captchaResponse struct {
	Success     bool      `json:"success"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}

func (a *Application) verifyCaptcha(response string) error {
	form := url.Values{}
	form.Add("secret", a.Config.Recaptcha.SecretKey)
	form.Add("response", response)
	req, err := http.NewRequest(http.MethodPost, "https://www.google.com/recaptcha/api/siteverify", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var captchaResponse captchaResponse
	err = json.NewDecoder(resp.Body).Decode(&captchaResponse)
	if err != nil {
		return err
	}
	log.Info().Interface("resp", captchaResponse).Msg("captcha response")
	if !captchaResponse.Success {
		return errors.New("captcha failed")
	}
	return nil
}

func (a *Application) HandleTeacherCreateAccount(w http.ResponseWriter, r *http.Request) {
	log := a.Log.With().Str("page_name", "teacher_create_account").Logger()
	if err := r.ParseForm(); err != nil {
		log.Error().Err(err).Msg("failed to parse form")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	emailAddress := r.FormValue("email-address")
	name := r.FormValue("your-name")

	captchaResponse := r.FormValue("g-recaptcha-response")
	if err := a.verifyCaptcha(captchaResponse); err != nil {
		log.Error().Err(err).Msg("failed to verify captcha")
		w.WriteHeader(http.StatusBadRequest)
		a.TeacherCreateAccountRenderer(w, r, map[string]any{
			"Name":         name,
			"Email":        emailAddress,
			"CaptchaError": true,
		})
		return
	}

	err := a.DB.NewTeacher(name, emailAddress)
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr); sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique || sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
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
	signedTok, err := tok.SignedString(a.Config.ReadSecretKey())
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
	if !a.Config.RegistrationEnabled {
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

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

		return a.Config.ReadSecretKey(), nil
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
	jwtStr, err := jwt.SignedString(a.Config.ReadSecretKey())
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
