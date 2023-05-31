package website

import "embed"

//go:embed templates/*
var TemplateFS embed.FS

//go:embed static/* static/fonts/*
var StaticFS embed.FS
