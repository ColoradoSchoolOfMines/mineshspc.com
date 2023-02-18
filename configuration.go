package main

import (
	"html/template"

	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Domain         string        `yaml:"domain"`
	SendGridAPIKey string        `yaml:"sendgrid_api_key"`
	HealthcheckURL string        `yaml:"healthcheck_url"`
	HostedByHTML   template.HTML `yaml:"hosted_by_html"`
}

func (c *Configuration) Parse(data []byte) error {
	return yaml.Unmarshal(data, c)
}
