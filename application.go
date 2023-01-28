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

type TemplateData[T any] struct {
	PageName string
	Data     T
}

func ServeTemplate[T any](logger *zerolog.Logger, templateName string, generateTemplateData func(r *http.Request) T) func(w http.ResponseWriter, r *http.Request) {
	log := logger.With().Str("page_name", templateName).Logger()

	template, err := template.ParseFS(templateFS, "templates/base.html", "templates/partials/*", fmt.Sprintf("templates/%s", templateName))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse template")
	}

	parts := strings.Split(templateName, ".")

	return func(w http.ResponseWriter, r *http.Request) {
		templateData := TemplateData[T]{
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
	r.HandleFunc("/", ServeTemplate(a.Log, "home.html", noArgs))
	r.HandleFunc("/authors/", ServeTemplate(a.Log, "authors.html", noArgs))
	r.HandleFunc("/rules/", ServeTemplate(a.Log, "rules.html", noArgs))
	r.HandleFunc("/register/", ServeTemplate(a.Log, "register.html", noArgs))
	r.HandleFunc("/faq/", ServeTemplate(a.Log, "faq.html", noArgs))
	r.HandleFunc("/archive/", ServeTemplate(a.Log, "archive.html", a.GetArchiveTemplate))

	// Serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/", http.FileServer(http.FS(staticFS))))

	r.HandleFunc("/register/teacher", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.URL.String()+"/createaccount", http.StatusTemporaryRedirect)
	})
	r.HandleFunc("/register/teacher/login", ServeTemplate(a.Log, "teacherlogin.html", noArgs))
	r.HandleFunc("/register/teacher/createaccount", ServeTemplate(a.Log, "teachercreateaccount.html", a.GetTeacherRegistrationTemplate))
	r.HandleFunc("/register/student", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.URL.String()+"/confirminfo", http.StatusTemporaryRedirect)
	})
	r.HandleFunc("/register/student/confirminfo", ServeTemplate(a.Log, "student.html", noArgs))
	r.HandleFunc("/register/parent", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.URL.String()+"/signforms", http.StatusTemporaryRedirect)
	})
	r.HandleFunc("/register/parent/signforms", ServeTemplate(a.Log, "parent.html", noArgs))

	http.Handle("/", r)

	a.Log.Info().Msg("Listening on port 8090")
	http.ListenAndServe(":8090", nil)
}
