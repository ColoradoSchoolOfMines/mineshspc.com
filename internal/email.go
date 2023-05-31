package internal

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func (a *Application) SendEmail(log zerolog.Logger, subject string, to *mail.Email, plainTextContent, htmlContent string) error {
	log = log.With().
		Str("component", "send_email").
		Interface("to", to).
		Str("subject", subject).
		Logger()

	if a.Config.DevMode {
		fmt.Printf("=== EMAIL ===\nTo: %s\nSubject: %s\n\n%s\n\n", to, subject, plainTextContent)

		return nil
	}

	from := mail.NewEmail("Mines HSPC Support", "support@mineshspc.com")
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	message.ReplyTo = from
	resp, err := a.SendGridClient.Send(message)
	if err != nil {
		log.Err(err).Msg("failed to send email")
		return err
	} else if resp.StatusCode != http.StatusAccepted {
		log.Error().
			Int("status_code", resp.StatusCode).
			Interface("to", to).
			Str("response_body", resp.Body).
			Msg("error sending email")
		return fmt.Errorf("error sending email (error code %d)", resp.StatusCode)
	} else {
		log.Info().
			Int("status_code", resp.StatusCode).
			Interface("to", to).
			Msg("successfully sent email")
		return nil
	}
}
