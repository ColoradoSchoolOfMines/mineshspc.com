package internal

import (
	"fmt"
	htmltemplate "html/template"
	"net/http"
	"strings"
	texttemplate "text/template"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/templates"
	registerteacher "github.com/ColoradoSchoolOfMines/mineshspc.com/internal/templates/register/teacher"
)

type Issuer string

const (
	IssuerEmailLogin     Issuer = "email_login"
	IssuerSessionToken   Issuer = "session_token"
	IssuerStudentVerify  Issuer = "student_verify"
	IssuerSignForms      Issuer = "sign_forms"
	IssuerAdminLogin     Issuer = "admin_login"
	IssuerStudentQRCode  Issuer = "student_qrcode"
	IssuerVolunteerLogin Issuer = "volunteer_login"
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
	// TODO invent some way to make this a one-time token (maybe just add extra
	// random token in here and then store all of the tokens we have seen in
	// RAM)
	return jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer:    string(IssuerEmailLogin),
		Subject:   email,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})
}

func (a *Application) HandleTeacherLogin(w http.ResponseWriter, r *http.Request) {
	log := a.Log.With().Str("page_name", "teacher_create_account").Logger()
	ctx := log.WithContext(r.Context())
	if err := r.ParseForm(); err != nil {
		log.Err(err).Msg("failed to parse form")
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

	teacher, err := a.DB.GetTeacherByEmail(ctx, emailAddress)
	if err != nil {
		log.Warn().Err(err).Msg("failed to find teacher by email")
		templates.Base("Teacher Login", registerteacher.Login(emailAddress, registerteacher.LoginEmailDoesNotExist())).Render(ctx, w)
		return
	} else if !teacher.EmailConfirmed {
		log.Warn().Err(err).Msg("teacher email not confirmed, not sending login code to avoid amplification attacks")
		templates.Base("Teacher Login", registerteacher.Login(emailAddress, registerteacher.LoginEmailNotConfirmed())).Render(ctx, w)
		return
	}

	tok := a.CreateEmailLoginJWT(emailAddress)
	signedTok, err := tok.SignedString(a.Config.ReadSecretKey())
	if err != nil {
		log.Err(err).Msg("failed to sign email login token")
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
		log.Err(err).Msg("failed to send email")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		log.Info().Msg("sent email")
		http.SetCookie(w, &http.Cookie{Name: "email", Value: emailAddress, Path: "/"})
		http.Redirect(w, r, "/register/teacher/emaillogin", http.StatusSeeOther)
	}
}
