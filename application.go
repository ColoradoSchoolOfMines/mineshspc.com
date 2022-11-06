package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/* static/fonts/*
var staticFS embed.FS

type IAmA string

const (
	IAmATeacher IAmA = "Teacher"
	IAmAStudent IAmA = "Student"
)

type InterestForm struct {
	Email string `json:"email"`
	IAmA  IAmA   `json:"iama"`
}

type Application struct {
	Log           *zerolog.Logger
	EmailRegex    *regexp.Regexp
	Configuration Configuration

	Client *sheets.Service

	TeacherRegistrationCaptchas map[uuid.UUID]string
}

func NewApplication(log *zerolog.Logger) *Application {
	emailRegex, _ := regexp.Compile(`(?i)^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$`)

	return &Application{
		Log:        log,
		EmailRegex: emailRegex,

		TeacherRegistrationCaptchas: map[uuid.UUID]string{},
	}
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatal().Err(err).Msg("Unable to read authorization code")
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to retrieve token from web")
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to cache oauth token")
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func (app *Application) Authenticate() {
	ctx := context.Background()
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to read client secret file")
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to parse client secret file to config")
	}
	client := getClient(config)
	app.Client, err = sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to retrieve Sheets client")
	}
}

func (app *Application) ConfigureSheets() {
	spreadsheet, err := app.Client.Spreadsheets.Get(app.Configuration.SpreadsheetID).Do()
	if err != nil {
		log.Fatal().Err(err).
			Str("spreadsheet_id", app.Configuration.SpreadsheetID).
			Msg("Cannot access spreadsheet")
	}

	ensureExistsWithHeaders := func(title string, headers []interface{}) {
		log.Info().Str("title", title).Msg("Ensuring that sheet exists")
		var sheetID int64
		for _, s := range spreadsheet.Sheets {
			if s.Properties.Title == title {
				sheetID = s.Properties.SheetId
				break
			}
		}

		if sheetID == 0 {
			_, err := app.Client.Spreadsheets.BatchUpdate(
				app.Configuration.SpreadsheetID,
				&sheets.BatchUpdateSpreadsheetRequest{
					Requests: []*sheets.Request{
						{
							AddSheet: &sheets.AddSheetRequest{
								Properties: &sheets.SheetProperties{
									Title: title,
								},
							},
						},
					},
				}).Do()

			if err != nil {
				log.Fatal().Err(err).Str("title", title).Msg("Couldn't create sheet with title")
			}
			for _, s := range spreadsheet.Sheets {
				if s.Properties.Title == title {
					sheetID = s.Properties.SheetId
				}
			}
		}

		log.Info().Str("title", title).Msg("Ensuring that sheet with has the correct headers")

		topRow := fmt.Sprintf("%s!A1:%s1", title, string('A'+len(headers)-1))
		valueRange, err := app.Client.Spreadsheets.Values.Get(app.Configuration.SpreadsheetID, topRow).Do()
		if err != nil {
			log.Fatal().Err(err).Msg("Couldn't get the values of the first row")
		}

		// Only do anything if there's nothing in the top row
		if len(valueRange.Values) > 0 {
			return
		}

		_, err = app.Client.Spreadsheets.Values.Update(
			app.Configuration.SpreadsheetID, topRow, &sheets.ValueRange{
				Values: [][]interface{}{headers},
			},
		).ValueInputOption("RAW").Do()
		if err != nil {
			log.Fatal().Err(err).Msg("Couldn't update the header row")
		}
	}

	ensureExistsWithHeaders("Interest", []interface{}{"Student or Teacher", "Email"})
}

func (a *Application) RegisterInterest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, POST, OPTIONS")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPut {
		a.Log.Error().Msg("register_interest must be PUT")
		http.Error(w, "register_interest must be PUT", http.StatusBadRequest)
		return
	}
	var interestForm InterestForm
	err := json.NewDecoder(r.Body).Decode(&interestForm)
	if err != nil {
		a.Log.Error().Err(err).Msg("Failed to decode interest form")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !a.EmailRegex.MatchString(interestForm.Email) {
		a.Log.Error().Msg("Invalid email address")
		http.Error(w, "Invalid email address", http.StatusBadRequest)
	}

	// Now, actually add the user to the spreadsheet
	_, err = a.Client.Spreadsheets.Values.Append(
		a.Configuration.SpreadsheetID,
		"Interest",
		&sheets.ValueRange{
			Values: [][]interface{}{
				{interestForm.IAmA, interestForm.Email},
			},
		}).ValueInputOption("RAW").Do()
	if err != nil {
		a.Log.Error().Err(err).Msg("Failed to add user to the spreadsheet")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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
	r.HandleFunc("/rules/", ServeTemplate(a.Log, "rules.html", noArgs))
	r.HandleFunc("/register/", ServeTemplate(a.Log, "register.html", noArgs))
	r.HandleFunc("/register/teacher/", ServeTemplate(a.Log, "teacher.html", a.GetTeacherTemplate))
	r.HandleFunc("/register/student/", ServeTemplate(a.Log, "register.html", func(r *http.Request) any {
		return map[string]any{
			"RegistrationID":  uuid.New().String(),
			"CaptchaElements": []string{"ab", "cd", "ef", "gh", "ij"},
			"CaptchaIndexes":  []int{1, 3, 2},
		}
	}))
	r.HandleFunc("/faq/", ServeTemplate(a.Log, "faq.html", noArgs))
	r.HandleFunc("/archive/", ServeTemplate(a.Log, "archive.html", a.GetArchiveTemplate))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/", http.FileServer(http.FS(staticFS))))

	// Registration endpoints
	r.HandleFunc("/register/", a.RegisterInterest)

	//app.Authenticate()
	//app.ConfigureSheets()

	http.Handle("/", r)
	a.Log.Info().Msg("Listening on port 8090")
	http.ListenAndServe(":8090", nil)
}
