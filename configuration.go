package main

import (
	"html/template"
	"os"

	"gopkg.in/yaml.v3"
)

type RecaptchaConfig struct {
	SiteKey   string `yaml:"site_key"`
	SecretKey string `yaml:"secret_key"`
}

type Configuration struct {
	secretKeyBytes []byte

	Domain         string        `yaml:"domain"`
	SendGridAPIKey string        `yaml:"sendgrid_api_key"`
	HealthcheckURL string        `yaml:"healthcheck_url"`
	HostedByHTML   template.HTML `yaml:"hosted_by_html"`
	SecretKeyFile  string        `yaml:"secret_key_file"`

	Recaptcha RecaptchaConfig `yaml:"recaptcha"`
}

func (c *Configuration) Parse(data []byte) error {
	return yaml.Unmarshal(data, c)
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
