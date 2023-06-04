package internal

import (
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/sendgrid/sendgrid-go"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/database"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/config"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/website"
)

type Application struct {
	Log        *zerolog.Logger
	DB         *database.Database
	EmailRegex *regexp.Regexp
	Config     config.Configuration

	ConfirmEmailRenderer       func(w http.ResponseWriter, r *http.Request, extraData map[string]any)
	TeacherLoginRenderer       func(w http.ResponseWriter, r *http.Request, extraData map[string]any)
	EmailLoginRenderer         func(w http.ResponseWriter, r *http.Request, extraData map[string]any)
	StudentConfirmInfoRenderer func(w http.ResponseWriter, r *http.Request, extraData map[string]any)
	TeamAddMemberRenderer      func(w http.ResponseWriter, r *http.Request, extraData map[string]any)

	TeacherCreateAccountRenderer func(w http.ResponseWriter, r *http.Request, extraData map[string]any)

	SendGridClient *sendgrid.Client
}

func NewApplication(log *zerolog.Logger, db *database.Database) *Application {
	return &Application{
		Log:        log,
		DB:         db,
		EmailRegex: regexp.MustCompile(`(?i)^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$`),
		Config:     config.InitConfiguration(),
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

	template, err := template.ParseFS(website.TemplateFS, "templates/base.html", "templates/partials/*", fmt.Sprintf("templates/%s", templateName))
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
		user, err := a.GetLoggedInTeacher(r)
		if err == nil {
			data["Username"] = user.Name
		}

		templateData := map[string]any{
			"PageName":            parts[0],
			"Data":                data,
			"HostedByHTML":        a.Config.HostedByHTML,
			"RegistrationEnabled": a.Config.RegistrationEnabled,
		}
		if err := template.ExecuteTemplate(w, "base.html", templateData); err != nil {
			log.Err(err).Msg("Failed to execute the template")
		}
	}
}

type renderInfo struct {
	RenderFn           func(w http.ResponseWriter, r *http.Request, extraData map[string]any)
	RedirectIfLoggedIn bool
}

func (a *Application) Start() {
	a.Log.Info().Msg("connecting to sendgrid")
	a.SendGridClient = sendgrid.NewSendClient(a.Config.SendGridAPIKey)

	a.Log.Info().Msg("Starting router")

	r := chi.NewRouter()

	noArgs := func(r *http.Request) map[string]any { return nil }

	// Static pages
	staticPages := map[string]struct {
		Template     string
		ArgGenerator func(r *http.Request) map[string]any
	}{
		"/":         {"home.html", noArgs},
		"/info":     {"info.html", noArgs},
		"/authors":  {"authors.html", noArgs},
		"/rules":    {"rules.html", noArgs},
		"/register": {"register.html", noArgs},
		"/faq":      {"faq.html", noArgs},
		"/archive":  {"archive.html", a.GetArchiveTemplate},
	}
	for path, templateInfo := range staticPages {
		r.Get(path, a.ServeTemplate(a.Log, templateInfo.Template, templateInfo.ArgGenerator))
	}

	// Serve static files
	r.Handle("/static/*", http.FileServer(http.FS(website.StaticFS)))

	// Redirect pages
	redirects := map[string]string{
		"/register/teacher": "/register/teacher/createaccount",
		"/register/student": "/",
		"/register/parent":  "/",
	}
	for path, redirectPath := range redirects {
		redirFn := func(redirectPath string) func(w http.ResponseWriter, r *http.Request) {
			return func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, redirectPath, http.StatusTemporaryRedirect)
			}
		}
		r.Get(path, redirFn(redirectPath))
	}

	// Registration renderers
	a.TeacherLoginRenderer = a.ServeTemplateExtra(a.Log, "teacherlogin.html", a.GetEmailLoginTemplate)
	a.TeacherCreateAccountRenderer = a.ServeTemplateExtra(a.Log, "teachercreateaccount.html", a.GetTeacherCreateAccountTemplate)
	a.ConfirmEmailRenderer = a.ServeTemplateExtra(a.Log, "confirmemail.html", a.GetEmailLoginTemplate)
	a.EmailLoginRenderer = a.ServeTemplateExtra(a.Log, "emaillogin.html", a.GetEmailLoginTemplate)
	a.StudentConfirmInfoRenderer = a.ServeTemplateExtra(a.Log, "student.html", a.GetStudentConfirmInfoTemplate)
	a.TeamAddMemberRenderer = a.ServeTemplateExtra(a.Log, "teamaddmember.html", a.GetTeacherAddMemberTemplate)
	registrationPages := map[string]renderInfo{
		"/register/teacher/confirmemail":   {a.ConfirmEmailRenderer, true},
		"/register/teacher/createaccount":  {a.TeacherCreateAccountRenderer, true},
		"/register/teacher/login":          {a.TeacherLoginRenderer, true},
		"/register/teacher/schoolinfo":     {a.ServeTemplateExtra(a.Log, "schoolinfo.html", a.GetTeacherSchoolInfoTemplate), false},
		"/register/teacher/teams":          {a.ServeTemplateExtra(a.Log, "teams.html", a.GetTeacherTeamsTemplate), false},
		"/register/teacher/team/edit":      {a.ServeTemplateExtra(a.Log, "teamedit.html", a.GetTeacherTeamEditTemplate), false},
		"/register/teacher/team/addmember": {a.TeamAddMemberRenderer, false},

		// Student
		"/register/student/confirminfo": {a.StudentConfirmInfoRenderer, false},

		// Parent
		"/register/parent/signforms": {a.ServeTemplateExtra(a.Log, "parent.html", a.GetParentSignFormsTemplate), false},
	}
	for path, rend := range registrationPages {
		renderFn := func(path string, rend renderInfo) func(w http.ResponseWriter, r *http.Request) {
			return func(w http.ResponseWriter, r *http.Request) {
				if teacher, err := a.GetLoggedInTeacher(r); err != nil {
					a.Log.Info().Err(err).
						Bool("redirect_if_logged_in", rend.RedirectIfLoggedIn).
						Msg("Failed to get logged in teacher")
					if rend.RedirectIfLoggedIn && path != "/register/teacher/login" && path != "/register/teacher/createaccount" && path != "/register/teacher/confirmemail" {
						http.Redirect(w, r, "/register/teacher/login", http.StatusTemporaryRedirect)
					}
				} else if rend.RedirectIfLoggedIn && teacher != nil {
					if teacher.SchoolCity == "" || teacher.SchoolName == "" || teacher.SchoolState == "" {
						http.Redirect(w, r, "/register/teacher/schoolinfo", http.StatusSeeOther)
					} else {
						http.Redirect(w, r, "/register/teacher/teams", http.StatusSeeOther)
					}
					return
				}
				rend.RenderFn(w, r, nil)
			}
		}
		r.Get(path, renderFn(path, rend))
	}

	// Delete Team member
	r.Get("/register/teacher/team/delete", a.HandleTeacherDeleteMember)

	// Email confirmation code handling
	r.Get("/register/teacher/emaillogin", a.HandleTeacherEmailLogin)

	// Logout
	r.Get("/register/teacher/logout", a.HandleTeacherLogout)

	// Form Post Handlers
	formHandlers := map[string]func(w http.ResponseWriter, r *http.Request){
		"/register/teacher/login":          a.HandleTeacherLogin,
		"/register/teacher/createaccount":  a.HandleTeacherCreateAccount,
		"/register/teacher/schoolinfo":     a.HandleTeacherSchoolInfo,
		"/register/teacher/team/edit":      a.HandleTeacherTeamEdit,
		"/register/teacher/team/addmember": a.HandleTeacherAddMember,
		"/register/student/confirminfo":    a.HandleStudentConfirmEmail,
		"/register/parent/signforms":       a.HandleParentSignForms,
	}
	for path, fn := range formHandlers {
		renderFn := func(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
			return func(w http.ResponseWriter, r *http.Request) {
				handler(w, r)
			}
		}
		r.Post(path, renderFn(fn))
	}

	// Admin pages
	r.Get("/admin", a.ServeTemplate(a.Log, "adminhome.html", noArgs))
	r.Get("/admin/login", a.ServeTemplate(a.Log, "adminlogin.html", noArgs))
	r.Get("/admin/emaillogin", a.HandleAdminEmailLogin)
	r.Post("/admin/emaillogin", a.HandleAdminLogin)
	r.Get("/admin/resendstudentemail", a.HandleResendStudentEmail)
	r.Get("/admin/resendparentemail", a.HandleResendParentEmail)
	r.Get("/admin/confirmationlink/student", a.HandleGetStudentEmailConfirmationLink)
	r.Get("/admin/confirmationlink/parent", a.HandleGetParentEmailConfirmationLink)
	r.Get("/admin/dietaryrestrictions", a.ServeTemplate(a.Log, "admindietaryrestrictions.html", a.GetAdminDietaryRestrictionsTemplate))
	r.Get("/admin/teams", a.ServeTemplate(a.Log, "adminteams.html", a.GetAdminTeamsTemplate))
	r.Get("/admin/sendemailconfirmationreminders", a.HandleSendEmailConfirmationReminders)
	r.Get("/admin/sendparentreminders", a.HandleSendParentReminders)
	r.Get("/admin/sendqrcodes", a.HandleSendQRCodes)
	r.Get("/admin/kattis/participants", a.HandleKattisParticipantsExport)
	r.Get("/admin/kattis/teams", a.HandleKattisTeamsExport)
	r.Get("/admin/zoom/breakout", a.HandleZoomBreakoutExport)

	// Volunteer pages
	r.Get("/volunteer", a.ServeTemplate(a.Log, "volunteerhome.html", noArgs))
	r.Get("/volunteer/login", a.ServeTemplate(a.Log, "volunteerlogin.html", noArgs))
	r.Get("/volunteer/emaillogin", a.HandleVolunteerEmailLogin)
	r.Post("/volunteer/emaillogin", a.HandleVolunteerLogin)
	r.Get("/volunteer/scan", a.ServeTemplate(a.Log, "volunteerscan.html", a.GetVolunteerScanTemplate))
	r.Get("/volunteer/checkin", a.HandleVolunteerCheckIn)

	var handler http.Handler = r
	handler = hlog.RequestIDHandler("request_id", "RequestID")(handler)
	handler = hlog.NewHandler(*a.Log)(handler)

	http.Handle("/", r)

	a.Log.Info().Msg("Listening on port 8090")
	http.ListenAndServe(":8090", handler)
}
