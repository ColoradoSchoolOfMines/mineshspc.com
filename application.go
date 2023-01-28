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
	"github.com/sendgrid/sendgrid-go"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/database"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/* static/fonts/*
var staticFS embed.FS

type Application struct {
	Log           *zerolog.Logger
	DB            *database.Database
	EmailRegex    *regexp.Regexp
	Configuration Configuration

	LoginCodes           map[string]uuid.UUID
	ConfirmEmailRenderer func(w http.ResponseWriter, r *http.Request, extraData map[string]any)

	TeacherRegistrationCaptchas  map[uuid.UUID]string
	TeacherCreateAccountRenderer func(w http.ResponseWriter, r *http.Request, extraData map[string]any)

	TeacherSchoolInfoRenderer func(w http.ResponseWriter, r *http.Request, extraData map[string]any)

	SendGridClient *sendgrid.Client
}

func NewApplication(log *zerolog.Logger, db *sql.DB) *Application {
	return &Application{
		Log:        log,
		DB:         database.NewDatabase(db, log.With().Str("module", "database").Logger()),
		EmailRegex: regexp.MustCompile(`(?i)^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$`),

		LoginCodes:                  map[string]uuid.UUID{},
		TeacherRegistrationCaptchas: map[uuid.UUID]string{},
	}
}

func (a *Application) ServeTemplate(logger *zerolog.Logger, templateName string, generateTemplateData func(r *http.Request) map[string]any) func(w http.ResponseWriter, r *http.Request) {
	serveTemplateFn := a.ServeTemplateExtra(logger, templateName, generateTemplateData)
	return func(w http.ResponseWriter, r *http.Request) {
		serveTemplateFn(w, r, nil)
	}
}

func (a *Application) ServeTemplateExtra(logger *zerolog.Logger, templateName string, generateTemplateData func(r *http.Request) map[string]any) func(w http.ResponseWriter, r *http.Request, extraData map[string]any) {
	log := logger.With().Str("page_name", templateName).Logger()

	template, err := template.ParseFS(templateFS, "templates/base.html", "templates/partials/*", fmt.Sprintf("templates/%s", templateName))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse template")
	}

	parts := strings.Split(templateName, ".")

	return func(w http.ResponseWriter, r *http.Request, extraData map[string]any) {
		data := generateTemplateData(r)
		if data == nil {
			data = map[string]any{}
		}
		for k, v := range extraData {
			data[k] = v
		}
		templateData := map[string]any{
			"PageName": parts[0],
			"Data":     data,
		}
		err := template.ExecuteTemplate(w, "base.html", templateData)
		if err != nil {
			log.Error().Err(err).Msg("Failed to execute the template")
		}
	}
}

func (a *Application) Start() {
	a.DB.RunMigrations()

	a.Log.Info().Msg("connecting to sendgrid")
	a.SendGridClient = sendgrid.NewSendClient(a.Configuration.SendGridAPIKey)

	a.Log.Info().Msg("Starting router")
	r := mux.NewRouter().StrictSlash(true)

	noArgs := func(r *http.Request) map[string]any { return nil }

	// Static pages
	staticPages := map[string]struct {
		Template     string
		ArgGenerator func(r *http.Request) map[string]any
	}{
		"/":         {"home.html", noArgs},
		"/authors":  {"authors.html", noArgs},
		"/rules":    {"rules.html", noArgs},
		"/register": {"register.html", noArgs},
		"/faq":      {"faq.html", noArgs},
		"/archive":  {"archive.html", a.GetArchiveTemplate},

		"/register/teacher/login":       {"teacherlogin.html", noArgs},
		"/register/student/confirminfo": {"student.html", noArgs},
		"/register/parent/signforms":    {"parent.html", noArgs},
	}
	for path, templateInfo := range staticPages {
		r.HandleFunc(path, a.ServeTemplate(a.Log, templateInfo.Template, templateInfo.ArgGenerator)).Methods(http.MethodGet)
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

	// Registration renderers
	a.TeacherCreateAccountRenderer = a.ServeTemplateExtra(a.Log, "teachercreateaccount.html", a.GetTeacherRegistrationTemplate)
	a.TeacherSchoolInfoRenderer = a.ServeTemplateExtra(a.Log, "schoolinfo.html", a.GetTeacherSchoolInfoTemplate)
	registrationPages := map[string]func(w http.ResponseWriter, r *http.Request, extraData map[string]any){
		"/register/teacher/createaccount": a.TeacherCreateAccountRenderer,
		"/register/teacher/schoolinfo":    a.TeacherSchoolInfoRenderer,
	}
	for path, renderer := range registrationPages {
		r.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) { renderer(w, r, nil) }).Methods(http.MethodGet)
	}

	// Email confirmation code handling
	a.ConfirmEmailRenderer = a.ServeTemplateExtra(a.Log, "confirmemail.html", noArgs)
	r.HandleFunc("/register/teacher/confirmemail", a.HandleTeacherEmailConfirmation).Methods(http.MethodGet)

	// Form Post Handlers
	formHandlers := map[string]func(w http.ResponseWriter, r *http.Request){
		"/register/teacher/login":         a.HandleTeacherLogin,
		"/register/teacher/createaccount": a.HandleTeacherCreateAccount,
		"/register/teacher/schoolinfo":    a.HandleTeacherSchoolInfo,
	}
	for path, fn := range formHandlers {
		r.HandleFunc(path, fn).Methods(http.MethodPost)
	}

	http.Handle("/", r)

	a.Log.Info().Msg("Listening on port 8090")
	http.ListenAndServe(":8090", nil)
}
