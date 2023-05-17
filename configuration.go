package main

import (
	"html/template"
	"os"

	"github.com/spf13/viper"
)

type Configuration struct {
	secretKeyBytes []byte

	DevMode bool

	Domain              string
	SendGridAPIKey      string
	HealthcheckURL      string
	HostedByHTML        template.HTML
	RegistrationEnabled bool

	Recaptcha struct {
		SiteKey   string
		SecretKey string
	}
}

func InitConfiguration() Configuration {
	return Configuration{
		DevMode: viper.GetBool("dev_mode"),

		Domain:              viper.GetString("domain"),
		SendGridAPIKey:      viper.GetString("sendgrid_api_key"),
		HealthcheckURL:      viper.GetString("healthcheck_url"),
		HostedByHTML:        template.HTML(viper.GetString("hosted_by_html")),
		RegistrationEnabled: viper.GetBool("registration_enabled"),
		Recaptcha: struct {
			SiteKey   string
			SecretKey string
		}{
			SiteKey:   viper.GetString("recaptcha.site_key"),
			SecretKey: viper.GetString("recaptcha.secret_key"),
		},
	}
}

func (c *Configuration) ReadSecretKey() []byte {
	if len(c.secretKeyBytes) == 0 {
		secretKey := viper.GetString("jwt_secret_key")
		if secretKey != "" {
			return []byte(secretKey)
		}

		var err error
		c.secretKeyBytes, err = os.ReadFile(viper.GetString("jwt_secret_key_file"))
		if err != nil {
			panic(err)
		}
	}
	return c.secretKeyBytes
}
