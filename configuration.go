package main

import (
	"html/template"
	"os"

	"gopkg.in/yaml.v3"
)

type Configuration struct {
	secretKeyBytes []byte

	Domain         string        `yaml:"domain"`
	SendGridAPIKey string        `yaml:"sendgrid_api_key"`
	HealthcheckURL string        `yaml:"healthcheck_url"`
	HostedByHTML   template.HTML `yaml:"hosted_by_html"`
	SecretKeyFile  string        `yaml:"secret_key_file"`
}

func (c *Configuration) Parse(data []byte) error {
	return yaml.Unmarshal(data, c)
}

func (c *Configuration) ReadGetSecretKey() []byte {
	if len(c.secretKeyBytes) == 0 {
		var err error
		c.secretKeyBytes, err = os.ReadFile(c.SecretKeyFile)
		if err != nil {
			panic(err)
		}
	}
	return c.secretKeyBytes
}
