package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

type TeacherRegistrationTemplate struct {
	RegistrationID  uuid.UUID
	CaptchaElements []string
	CaptchaIndexes  []int
}

func (a *Application) GetTeacherTemplate(r *http.Request) TeacherRegistrationTemplate {
	captchaElements := make([]string, 5)
	for i := range captchaElements {
		captchaElements[i] = string(alphabet[rand.Intn(len(alphabet))]) + string(alphabet[rand.Intn(len(alphabet))])
	}

	captchaAnswer := ""
	captchaIndexes := make([]int, 3)
	for i := range captchaIndexes {
		index := rand.Intn(4)
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
		a.Log.Info().
			Str("registration_id", registrationID.String()).
			Msg("expiring registration")
		delete(a.TeacherRegistrationCaptchas, registrationID)
	}()

	return TeacherRegistrationTemplate{
		RegistrationID:  registrationID,
		CaptchaElements: captchaElements,
		CaptchaIndexes:  captchaIndexes,
	}
}
