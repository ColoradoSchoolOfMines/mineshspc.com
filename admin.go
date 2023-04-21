package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/database"
)

func (a *Application) createAdminLoginJWT(email string) *jwt.Token {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer:    string(IssuerAdminLogin),
		Subject:   email,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(6 * time.Hour)),
	})
}

func (a *Application) isAdminByToken(tokenStr string) (bool, error) {
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

	if claims.Issuer != string(IssuerAdminLogin) {
		return false, fmt.Errorf("invalid student verify token: %w", err)
	}

	return true, nil
}

func (a *Application) GetAdminTeamsTemplate(r *http.Request) map[string]any {
	tok, err := r.Cookie("admin_token")
	if err != nil {
		a.Log.Warn().Err(err).Msg("failed to get admin token")
		return nil
	}

	if isAdmin, err := a.isAdminByToken(tok.Value); err != nil || !isAdmin {
		a.Log.Warn().Err(err).Msg("user is not admin!")
		return nil
	}

	teamsWithTeachers, err := a.DB.GetAdminTeamsWithTeacherName(r.Context())
	if err != nil {
		a.Log.Err(err).Msg("failed to get teams")
		return nil
	}

	var beginnerTeams, advancedTeams int
	var beginnerStudents, advancedStudents int
	var beginnerInPersonTeams, advancedInPersonTeams int
	var beginnerInPersonStudents, advancedInPersonStudents int
	var campusTour, emailConfirmed, formsSigned int
	for _, team := range teamsWithTeachers {
		if team.Division == database.DivisionBeginner {
			beginnerTeams++
			beginnerStudents += len(team.Members)
			if team.InPerson {
				beginnerInPersonStudents += len(team.Members)
				beginnerInPersonTeams++
			}
		} else {
			advancedTeams++
			advancedStudents += len(team.Members)
			if team.InPerson {
				advancedInPersonStudents += len(team.Members)
				advancedInPersonTeams++
			}
		}

		for _, member := range team.Members {
			if team.InPerson && member.CampusTour {
				campusTour++
			}

			if member.EmailConfirmed {
				emailConfirmed++
			}

			if member.LiabilitySigned {
				formsSigned++
			}
		}
	}

	return map[string]any{
		"Teams": teamsWithTeachers,
		"TeamStats": map[string]any{
			"Beginner": map[string]int{
				"InPerson": beginnerInPersonTeams,
				"Remote":   beginnerTeams - beginnerInPersonTeams,
				"Total":    beginnerTeams,
			},
			"Advanced": map[string]int{
				"InPerson": advancedInPersonTeams,
				"Remote":   advancedTeams - advancedInPersonTeams,
				"Total":    advancedTeams,
			},
			"TotalInPerson": beginnerInPersonTeams + advancedInPersonTeams,
			"TotalRemote":   (beginnerTeams + advancedTeams) - (beginnerInPersonTeams - advancedInPersonTeams),
		},
		"StudentStats": map[string]any{
			"Beginner": map[string]int{
				"InPerson": beginnerInPersonStudents,
				"Remote":   beginnerStudents - beginnerInPersonStudents,
				"Total":    beginnerStudents,
			},
			"Advanced": map[string]int{
				"InPerson": advancedInPersonStudents,
				"Remote":   advancedStudents - advancedInPersonStudents,
				"Total":    advancedStudents,
			},
			"TotalInPerson": beginnerInPersonStudents + advancedInPersonStudents,
			"TotalRemote":   (beginnerStudents + advancedStudents) - (beginnerInPersonStudents - advancedInPersonStudents),
		},
		"TotalTeams":             beginnerTeams + advancedTeams,
		"TotalStudents":          beginnerStudents + advancedStudents,
		"CampusTourStudents":     campusTour,
		"EmailConfirmedStudents": emailConfirmed,
		"FormsSignedStudents":    formsSigned,
	}
}

func (a *Application) GetAdminDietaryRestrictionsTemplate(r *http.Request) map[string]any {
	tok, err := r.Cookie("admin_token")
	if err != nil {
		a.Log.Warn().Err(err).Msg("failed to get admin token")
		return nil
	}

	if isAdmin, err := a.isAdminByToken(tok.Value); err != nil || !isAdmin {
		a.Log.Warn().Err(err).Msg("user is not admin!")
		return nil
	}

	dietaryRestrictions, err := a.DB.GetAllDietaryRestrictions(r.Context())
	if err != nil {
		a.Log.Err(err).Msg("failed to get dietary restrictions")
		return nil
	}

	return map[string]any{
		"DietaryRestrictions": dietaryRestrictions,
	}
}

func (a *Application) HandleAdminEmailLogin(w http.ResponseWriter, r *http.Request) {
	tok := r.URL.Query().Get("tok")
	log.Info().Str("token", tok).Msg("got token")
	isAdmin, err := a.isAdminByToken(tok)
	if err != nil || !isAdmin {
		a.Log.Warn().Msg("failed to get admin")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "admin_token", Value: tok, Path: "/"})
	http.Redirect(w, r, "/admin/teams", http.StatusSeeOther)
}

func (a *Application) HandleAdminLogin(w http.ResponseWriter, r *http.Request) {
	log := a.Log.With().Str("page_name", "admin_login").Logger()
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

	isAdmin, err := a.DB.IsEmailAdmin(r.Context(), emailAddress)
	if err != nil {
		log.Warn().Err(err).Msg("failed to find admin by email")
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if !isAdmin {
		log.Warn().Err(err).Msg("teacher email not confirmed, not sending login code to avoid amplification attacks")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tok := a.createAdminLoginJWT(emailAddress)
	signedTok, err := tok.SignedString(a.Config.ReadSecretKey())
	if err != nil {
		log.Err(err).Msg("failed to sign email login token")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	plainTextContent := `Please click the following link to log in to your Mines HSPC admin account:

	` + fmt.Sprintf("%s/admin/emaillogin?tok=%s", a.Config.Domain, signedTok)

	err = a.SendEmail(log, "Log in to Mines HSPC Admin",
		mail.NewEmail("", emailAddress),
		plainTextContent,
		"")
	if err != nil {
		log.Err(err).Msg("failed to send email")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		log.Info().Msg("sent email")
		http.SetCookie(w, &http.Cookie{Name: "admin_email", Value: emailAddress, Path: "/"})
		w.Write([]byte("check your email for a login link"))
	}
}

func (a *Application) HandleResendStudentEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tok, err := r.Cookie("admin_token")
	if err != nil {
		a.Log.Warn().Err(err).Msg("failed to get admin token")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if isAdmin, err := a.isAdminByToken(tok.Value); err != nil || !isAdmin {
		a.Log.Warn().Err(err).Msg("user is not admin!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		a.Log.Warn().Msg("no email address provided in query string")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	student, err := a.DB.GetStudentByEmail(ctx, email)
	if err != nil {
		a.Log.Warn().Err(err).Msg("failed to get student by email")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	teacher, err := a.DB.GetTeacherForTeam(ctx, student.TeamID)
	if err != nil {
		log.Err(err).Msg("failed to get student's teacher")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	team, err := a.DB.GetTeamNoMembers(ctx, student.TeamID)
	if err != nil {
		log.Err(err).Msg("failed to get student's team")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := a.sendStudentEmail(ctx, student.Email, student.Name, teacher.Name, team.Name, false); err != nil {
		a.Log.Err(err).Msg("failed to send student email")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	page := r.URL.Query().Get("page")
	if page == "volunteer" {
		http.Redirect(w, r, "/volunteer/scan", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/admin/teams", http.StatusSeeOther)
	}
}

func (a *Application) HandleResendParentEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tok, err := r.Cookie("admin_token")
	if err != nil {
		a.Log.Warn().Err(err).Msg("failed to get admin token")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if isAdmin, err := a.isAdminByToken(tok.Value); err != nil || !isAdmin {
		a.Log.Warn().Err(err).Msg("user is not admin!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		a.Log.Warn().Msg("no email address provided in query string")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	student, err := a.DB.GetStudentByEmail(ctx, email)
	if err != nil {
		a.Log.Warn().Err(err).Msg("failed to get student by email")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := a.sendParentEmail(ctx, student, false); err != nil {
		a.Log.Err(err).Msg("failed to send parent email")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	page := r.URL.Query().Get("page")
	if page == "volunteer" {
		http.Redirect(w, r, "/volunteer/scan", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/admin/teams", http.StatusSeeOther)
	}
}

func (a *Application) HandleGetStudentEmailConfirmationLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tok, err := r.Cookie("admin_token")
	if err != nil {
		a.Log.Warn().Err(err).Msg("failed to get admin token")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if isAdmin, err := a.isAdminByToken(tok.Value); err != nil || !isAdmin {
		a.Log.Warn().Err(err).Msg("user is not admin!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		a.Log.Warn().Msg("no email address provided in query string")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = a.DB.GetStudentByEmail(ctx, email)
	if err != nil {
		a.Log.Warn().Err(err).Msg("failed to get student by email")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	confirmationLink, err := a.getStudentConfirmEmailLink(email)
	if err != nil {
		a.Log.Err(err).Msg("failed to get student confirmation link")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte(confirmationLink))
}

func (a *Application) HandleGetParentEmailConfirmationLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tok, err := r.Cookie("admin_token")
	if err != nil {
		a.Log.Warn().Err(err).Msg("failed to get admin token")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if isAdmin, err := a.isAdminByToken(tok.Value); err != nil || !isAdmin {
		a.Log.Warn().Err(err).Msg("user is not admin!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		a.Log.Warn().Msg("no email address provided in query string")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = a.DB.GetStudentByEmail(ctx, email)
	if err != nil {
		a.Log.Warn().Err(err).Msg("failed to get student by email")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	signURL, err := a.getParentSignFormsLink(email)
	if err != nil {
		a.Log.Err(err).Msg("failed to get parent sign forms link")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte(signURL))
}

func (a *Application) HandleSendEmailConfirmationReminders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := hlog.FromRequest(r).With().Str("action", "send_email_confirmation_reminders").Logger()
	tok, err := r.Cookie("admin_token")
	if err != nil {
		log.Warn().Err(err).Msg("failed to get admin token")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if isAdmin, err := a.isAdminByToken(tok.Value); err != nil || !isAdmin {
		log.Warn().Err(err).Msg("user is not admin!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	teamsWithTeachers, err := a.DB.GetAdminTeamsWithTeacherName(ctx)
	if err != nil {
		log.Err(err).Msg("failed to get teams with teachers")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, team := range teamsWithTeachers {
		for _, member := range team.Members {
			if member.EmailConfirmed {
				w.Write([]byte(fmt.Sprintf("Not resending confirmation info to %s because they already finished confirming\n", member.Email)))
				continue
			}
			w.Write([]byte(fmt.Sprintf("Resending confirmation email to %s\n", member.Email)))
			go func(team *database.TeamWithTeacherName, member database.Student) {
				ctx := log.WithContext(context.Background())
				err := a.sendStudentEmail(ctx, member.Email, member.Name, team.TeacherName, team.Name, true)
				if err != nil {
					log.Err(err).Msg("failed to send student email")
					return
				}
			}(team, member)
		}
	}
}

func (a *Application) HandleSendParentReminders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := hlog.FromRequest(r).With().Str("action", "send_parent_reminders").Logger()

	tok, err := r.Cookie("admin_token")
	if err != nil {
		log.Warn().Err(err).Msg("failed to get admin token")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if isAdmin, err := a.isAdminByToken(tok.Value); err != nil || !isAdmin {
		log.Warn().Err(err).Msg("user is not admin!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	teamsWithTeachers, err := a.DB.GetAdminTeamsWithTeacherName(ctx)
	if err != nil {
		log.Err(err).Msg("failed to get teams with teachers")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, team := range teamsWithTeachers {
		for _, member := range team.Members {
			if !member.EmailConfirmed {
				w.Write([]byte(fmt.Sprintf("Not sending sign forms email for %s because they already signed the forms because the student hasn't confirmed their email yet\n", member.Email)))
				continue
			}
			if member.LiabilitySigned {
				w.Write([]byte(fmt.Sprintf("Not resending sign forms email to %s (for %s) because they already signed the forms\n", member.ParentEmail, member.Email)))
				continue
			}
			w.Write([]byte(fmt.Sprintf("Resending sign forms email to %s (for %s)\n", member.ParentEmail, member.Email)))
			go func(member database.Student) {
				ctx := log.WithContext(context.Background())
				err := a.sendParentEmail(ctx, &member, true)
				if err != nil {
					log.Err(err).Msg("failed to send parent email")
					return
				}
			}(member)
		}
	}
}

func (a *Application) HandleSendQRCodes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := hlog.FromRequest(r).With().Str("action", "send_parent_reminders").Logger()
	tok, err := r.Cookie("admin_token")
	if err != nil {
		log.Warn().Err(err).Msg("failed to get admin token")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if isAdmin, err := a.isAdminByToken(tok.Value); err != nil || !isAdmin {
		log.Warn().Err(err).Msg("user is not admin!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	teamsWithTeachers, err := a.DB.GetAdminTeamsWithTeacherName(ctx)
	if err != nil {
		log.Err(err).Msg("failed to get teams with teachers")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, team := range teamsWithTeachers {
		for _, member := range team.Members {
			if member.QRCodeSent {
				w.Write([]byte(fmt.Sprintf("Not sending QR code to %s since we already sent to that email\n", member.Email)))
			} else if member.EmailConfirmed {
				w.Write([]byte(fmt.Sprintf("Sending QR code to %s\n", member.Email)))

				go func(member database.Student) {
					ctx := log.WithContext(context.Background())
					err := a.sendQRCodeEmail(ctx, member.Name, member.Email)
					if err != nil {
						log.Err(err).Msg("failed to send QR code email")
						return
					}

					err = a.DB.MarkQRCodeSent(ctx, member.Email)
					if err != nil {
						log.Err(err).Msg("failed to mark QR code sent")
					}
				}(member)
			} else {
				w.Write([]byte(fmt.Sprintf("Not sending QR code to %s since it's not confirmed\n", member.Email)))
			}
		}
	}
}

func (a *Application) HandleKattisParticipantsExport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tok, err := r.Cookie("admin_token")
	if err != nil {
		a.Log.Warn().Err(err).Msg("failed to get admin token")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if isAdmin, err := a.isAdminByToken(tok.Value); err != nil || !isAdmin {
		a.Log.Warn().Err(err).Msg("user is not admin!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	teamsWithTeachers, err := a.DB.GetAdminTeamsWithTeacherName(ctx)
	if err != nil {
		a.Log.Err(err).Msg("failed to get teams with teachers")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	divStr := r.URL.Query().Get("div")
	division, err := database.ParseDivision(divStr)
	if err != nil {
		a.Log.Err(err).Msg("invalid division")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer := csv.NewWriter(w)
	for _, team := range teamsWithTeachers {
		if team.Division != division {
			continue
		}
		for _, member := range team.Members {
			parts := []string{
				member.Name,
				member.Email,
				team.Name,
				"CONTESTANT",
				"",
				"",
				"",
			}
			writer.Write(parts)
		}
	}
	writer.Flush()
}

func (a *Application) HandleKattisTeamsExport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tok, err := r.Cookie("admin_token")
	if err != nil {
		a.Log.Warn().Err(err).Msg("failed to get admin token")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if isAdmin, err := a.isAdminByToken(tok.Value); err != nil || !isAdmin {
		a.Log.Warn().Err(err).Msg("user is not admin!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	teamsWithTeachers, err := a.DB.GetAdminTeamsWithTeacherName(ctx)
	if err != nil {
		a.Log.Err(err).Msg("failed to get teams with teachers")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	divStr := r.URL.Query().Get("div")
	division, err := database.ParseDivision(divStr)
	if err != nil {
		a.Log.Err(err).Msg("invalid division")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer := csv.NewWriter(w)
	for _, team := range teamsWithTeachers {
		if team.Division != division {
			continue
		}

		siteName := "Colorado School of Mines"
		if !team.InPerson {
			siteName = "Remote"
		}
		parts := []string{
			team.Name,
			siteName,
		}
		writer.Write(parts)
	}
	writer.Flush()
}
