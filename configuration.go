package main

import (
	"html/template"
	"os"

	"github.com/spf13/viper"
)

type Configuration struct {
	secretKeyBytes []byte

	Domain         string
	SendGridAPIKey string
	HealthcheckURL string
	HostedByHTML   template.HTML
	SecretKeyFile  string

	Recaptcha struct {
		SiteKey   string
		SecretKey string
	}
}

func InitConfiguration() Configuration {
	return Configuration{
		Domain:         viper.GetString("domain"),
		SendGridAPIKey: viper.GetString("sendgrid_api_key"),
		HealthcheckURL: viper.GetString("healthcheck_url"),
		HostedByHTML:   template.HTML(viper.GetString("hosted_by_html")),
		SecretKeyFile:  viper.GetString("secret_key_file"),
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
		var err error
		c.secretKeyBytes, err = os.ReadFile(c.SecretKeyFile)
		if err != nil {
			panic(err)
		}
	}
	return c.secretKeyBytes
}
