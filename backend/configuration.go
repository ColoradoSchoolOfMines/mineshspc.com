package main

import (
	"gopkg.in/yaml.v3"
)

type Configuration struct {
	SpreadsheetID string `yaml:"spreadsheet_id"`
}

func (c *Configuration) Parse(data []byte) error {
	return yaml.Unmarshal(data, c)
}
