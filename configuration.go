package main

import (
	"gopkg.in/yaml.v3"
)

type Configuration struct {
	SendGridAPIKey string `yaml:"sendgrid_api_key"`
	HealthcheckURL string `yaml:"healthcheck_url"`
}

func (c *Configuration) Parse(data []byte) error {
	return yaml.Unmarshal(data, c)
}
