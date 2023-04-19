package main

import (
	"context"
	"encoding/base64"
	"fmt"
	htmltemplate "html/template"
	"net/http"
	"strings"
	texttemplate "text/template"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	qrcode "github.com/skip2/go-qrcode"
)

func (a *Application) getStudentQRCodeURL(email string) (string, error) {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer:  string(IssuerStudentQRCode),
		Subject: email,
	})

	signedTok, err := tok.SignedString(a.Config.ReadSecretKey())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/volunteer/scan?tok=%s", a.Config.Domain, signedTok), nil
}

func (a *Application) getStudentQRCodeImage(email string) ([]byte, error) {
	url, err := a.getStudentQRCodeURL(email)
	if err != nil {
		return nil, err
	}
	return qrcode.Encode(url, qrcode.Medium, 256)
}

func (a *Application) sendQRCodeEmail(ctx context.Context, studentName, email string) error {
	subject := "Mines HSPC Ticket"
	log := zerolog.Ctx(ctx).With().
		Str("component", "send_email").
		Interface("to", email).
		Str("subject", subject).
		Logger()

	log.Info().Msg("sending email")

	qrcodeBytes, err := a.getStudentQRCodeImage(email)
	if err != nil {
		a.Log.Err(err).Msg("failed to get student QR code image")
		return err
	}
	// Base64 encode the qrcodeBytes
	qrcodeBase64 := base64.StdEncoding.EncodeToString(qrcodeBytes)

	templateData := map[string]any{
		"StudentName":  studentName,
		"QRCodeBase64": qrcodeBase64,
	}

	var plainTextContent, htmlContent strings.Builder
	texttemplate.Must(texttemplate.ParseFS(emailTemplates, "emailtemplates/ticket.txt")).Execute(&plainTextContent, templateData)
	htmltemplate.Must(htmltemplate.ParseFS(emailTemplates, "emailtemplates/ticket.html")).Execute(&htmlContent, templateData)

	from := mail.NewEmail("Mines HSPC Support", "support@mineshspc.com")
	message := mail.NewSingleEmail(from, subject, mail.NewEmail("", email), plainTextContent.String(), htmlContent.String())
	message.ReplyTo = from

	attachment := mail.NewAttachment()
	attachment.SetFilename("ticket.png")
	attachment.SetContent(qrcodeBase64)
	message.AddAttachment(attachment)

	log.Debug().Interface("pt", plainTextContent.String()).Interface("html", htmlContent.String()).Msg("sending email")

	resp, err := a.SendGridClient.Send(message)
	if err != nil {
		log.Err(err).Msg("failed to send email")
		return err
	} else if resp.StatusCode != http.StatusAccepted {
		log.Error().
			Int("status_code", resp.StatusCode).
			Str("response_body", resp.Body).
			Msg("error sending email")
		return fmt.Errorf("error sending email (error code %d)", resp.StatusCode)
	} else {
		log.Info().
			Int("status_code", resp.StatusCode).
			Msg("successfully sent email")
		return nil
	}
}
