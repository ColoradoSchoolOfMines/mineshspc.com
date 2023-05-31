package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog/log"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/database"
)

func (a *Application) createVolunteerLoginJWT(email string) *jwt.Token {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer:    string(IssuerVolunteerLogin),
		Subject:   email,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(6 * time.Hour)),
	})
}

func (a *Application) isVolunteerByToken(tokenStr string) (bool, error) {
	if tokenStr == "" {
		return false, errors.New("no token")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return a.Config.ReadSecretKey(), nil
	})
	if err != nil {
		return false, fmt.Errorf("failed to parse admin token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !token.Valid || !ok {
		return false, fmt.Errorf("failed to validate token: %w", err)
	}

	if claims.Issuer != string(IssuerVolunteerLogin) {
		return false, fmt.Errorf("invalid student verify token: %w", err)
	}

	return true, nil
}

func (a *Application) HandleVolunteerEmailLogin(w http.ResponseWriter, r *http.Request) {
	tok := r.URL.Query().Get("tok")
	log.Info().Str("token", tok).Msg("got token")
	isVolunteer, err := a.isVolunteerByToken(tok)
	if err != nil || !isVolunteer {
		a.Log.Warn().Msg("failed to get volunteer")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "volunteer_token", Value: tok, Path: "/"})
	http.Redirect(w, r, "/volunteer/scan", http.StatusSeeOther)
}

func (a *Application) HandleVolunteerLogin(w http.ResponseWriter, r *http.Request) {
	log := a.Log.With().Str("page_name", "volunteer_login").Logger()
	if err := r.ParseForm(); err != nil {
		log.Err(err).Msg("failed to parse form")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	emailAddress := r.FormValue("email")
	if emailAddress == "" {
		a.Log.Warn().Msg("no email address provided in request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log = log.With().Str("email", emailAddress).Logger()

	isVolunteer, err := a.DB.IsEmailVolunteer(r.Context(), emailAddress)
	if err != nil {
		log.Warn().Err(err).Msg("failed to find volunteer by email")
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if !isVolunteer {
		log.Warn().Err(err).Msg("teacher email not confirmed, not sending login code to avoid amplification attacks")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tok := a.createVolunteerLoginJWT(emailAddress)
	signedTok, err := tok.SignedString(a.Config.ReadSecretKey())
	if err != nil {
		log.Err(err).Msg("failed to sign email login token")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	plainTextContent := `Please click the following link to log in to your Mines HSPC volunteer account:

	` + fmt.Sprintf("%s/volunteer/emaillogin?tok=%s", a.Config.Domain, signedTok)

	err = a.SendEmail(log, "Log in as a Mines HSPC Volunteer",
		mail.NewEmail("", emailAddress),
		plainTextContent,
		"")
	if err != nil {
		log.Err(err).Msg("failed to send email")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		log.Info().Msg("sent email")
		http.SetCookie(w, &http.Cookie{Name: "volunteer_email", Value: emailAddress, Path: "/"})
		w.Write([]byte("check your email for a login link"))
	}
}

func (a *Application) GetVolunteerScanTemplate(r *http.Request) map[string]any {
	ctx := r.Context()
	tok, err := r.Cookie("volunteer_token")
	if err != nil {
		a.Log.Warn().Err(err).Msg("failed to get volunteer token")
		return nil
	}

	if isVolunteer, err := a.isVolunteerByToken(tok.Value); err != nil || !isVolunteer {
		a.Log.Warn().Err(err).Msg("user is not volunteer!")
		return nil
	}

	res := map[string]any{
		"LoggedInAsVolunteer": true,
	}

	if adminTok, err := r.Cookie("admin_token"); err == nil {
		if isAdmin, err := a.isAdminByToken(adminTok.Value); err == nil && isAdmin {
			res["LoggedInAsAdmin"] = true
		}
	}

	studentCheckInToken := r.URL.Query().Get("tok")
	student, err := a.getStudentByQRToken(ctx, studentCheckInToken)
	if err != nil {
		a.Log.Warn().Err(err).Msg("failed to get student by token")
		return res
	}
	res["Token"] = studentCheckInToken

	team, err := a.DB.GetTeamNoMembers(ctx, student.TeamID)
	if err != nil {
		a.Log.Warn().Err(err).Msg("failed to get team")
		return res
	}

	res["TeamName"] = team.Name

	if !team.InPerson {
		res["NotInPerson"] = true
		return res
	}

	res["AllGood"] = student.EmailConfirmed &&
		student.LiabilitySigned &&
		student.ComputerUseWaiverSigned

	res["Student"] = student

	return res
}

func (a *Application) HandleVolunteerCheckIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tok, err := r.Cookie("volunteer_token")
	if err != nil {
		a.Log.Warn().Err(err).Msg("failed to get volunteer token")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if isVolunteer, err := a.isVolunteerByToken(tok.Value); err != nil || !isVolunteer {
		a.Log.Warn().Err(err).Msg("user is not volunteer!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	studentSignInToken := r.URL.Query().Get("tok")
	student, err := a.getStudentByQRToken(ctx, studentSignInToken)
	if err != nil {
		a.Log.Warn().Err(err).Msg("failed to get student by token")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !student.CheckedIn {
		a.DB.CheckInStudent(ctx, student.Email)
	}

	http.Redirect(w, r, fmt.Sprintf("/volunteer/scan?tok=%s", studentSignInToken), http.StatusSeeOther)
}

func (a *Application) getStudentByQRToken(ctx context.Context, tokenStr string) (*database.Student, error) {
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

	if claims.Issuer != string(IssuerStudentQRCode) {
		return nil, fmt.Errorf("invalid student verify token: %w", err)
	}

	return a.DB.GetStudentByEmail(ctx, claims.Subject)
}
