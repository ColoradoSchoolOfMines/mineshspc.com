package main

import (
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/* static/fonts/*
var staticFS embed.FS

type Application struct {
	Log           *zerolog.Logger
	DB            *sql.DB
	EmailRegex    *regexp.Regexp
	Configuration Configuration

	TeacherRegistrationCaptchas map[uuid.UUID]string
}

func NewApplication(log *zerolog.Logger, db *sql.DB) *Application {
	return &Application{
		Log:        log,
		DB:         db,
		EmailRegex: regexp.MustCompile(`(?i)^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$`),

		TeacherRegistrationCaptchas: map[uuid.UUID]string{},
	}
}

type TemplateData struct {
	PageName string
	Data     any
}

func ServeTemplate(logger *zerolog.Logger, templateName string, generateTemplateData func(r *http.Request) any) func(w http.ResponseWriter, r *http.Request) {
	log := logger.With().Str("page_name", templateName).Logger()

	template, err := template.ParseFS(templateFS, "templates/base.html", "templates/partials/*", fmt.Sprintf("templates/%s", templateName))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse template")
	}

	parts := strings.Split(templateName, ".")

	return func(w http.ResponseWriter, r *http.Request) {
		templateData := TemplateData{
			PageName: parts[0],
			Data:     generateTemplateData(r),
		}
		err := template.ExecuteTemplate(w, "base.html", templateData)
		if err != nil {
			log.Error().Err(err).Msg("Failed to execute the template")
		}
	}
}

func (a *Application) Start() {
	a.Log.Info().Msg("Starting router")
	r := mux.NewRouter().StrictSlash(true)

	noArgs := func(r *http.Request) any { return nil }

	// Static pages
	staticPages := map[string]struct {
		Template     string
		ArgGenerator func(r *http.Request) any
	}{
		"/":         {"home.html", noArgs},
		"/authors":  {"authors.html", noArgs},
		"/rules":    {"rules.html", noArgs},
		"/register": {"register.html", noArgs},
		"/faq":      {"faq.html", noArgs},
		"/archive":  {"archive.html", a.GetArchiveTemplate},

		"/register/teacher/login":         {"teacherlogin.html", noArgs},
		"/register/teacher/createaccount": {"teachercreateaccount.html", a.GetTeacherRegistrationTemplate},
		"/register/student/confirminfo":   {"student.html", noArgs},
		"/register/parent/signforms":      {"parent.html", noArgs},
	}
	for path, templateInfo := range staticPages {
		r.HandleFunc(path, ServeTemplate(a.Log, templateInfo.Template, templateInfo.ArgGenerator)).Methods(http.MethodGet)
	}

	// Serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/", http.FileServer(http.FS(staticFS)))).Methods(http.MethodGet)

	// Registration pages
	redirects := map[string]string{
		"/register/teacher": "/register/teacher/createaccount",
		"/register/student": "/register/student/confirminfo",
		"/register/parent":  "/register/parent/signforms",
	}
	for path, redirectPath := range redirects {
		r.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, redirectPath, http.StatusTemporaryRedirect)
		}).Methods(http.MethodGet)
	}

	http.Handle("/", r)

	a.Log.Info().Msg("Listening on port 8090")
	http.ListenAndServe(":8090", nil)
}
