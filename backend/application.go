package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"

	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

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
	EmailRegex    *regexp.Regexp
	Configuration Configuration

	Client *sheets.Service
}

func NewApplication() *Application {
	emailRegex, _ := regexp.Compile(`(?i)^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$`)

	return &Application{
		EmailRegex: emailRegex,
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
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
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
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func (app *Application) Authenticate() {
	ctx := context.Background()
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)
	app.Client, err = sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}
}

func (app *Application) ConfigureSheets() {
	spreadsheet, err := app.Client.Spreadsheets.Get(app.Configuration.SpreadsheetID).Do()
	if err != nil {
		log.Fatalf("Cannot access spreadsheet with ID: %s: %v", app.Configuration.SpreadsheetID, err)
	}

	ensureExistsWithHeaders := func(title string, headers []interface{}) {
		log.Infof("Ensuring that sheet with title '%s' exists", title)
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
				log.Fatalf("Couldn't create sheet with title '%s': %v", title, err)
			}
			for _, s := range spreadsheet.Sheets {
				if s.Properties.Title == title {
					sheetID = s.Properties.SheetId
				}
			}
		}

		log.Infof("Ensuring that sheet with title '%s' has the correct headers", title)

		topRow := fmt.Sprintf("%s!A1:%s1", title, string('A'+len(headers)-1))
		valueRange, err := app.Client.Spreadsheets.Values.Get(app.Configuration.SpreadsheetID, topRow).Do()
		if err != nil {
			log.Fatal("Couldn't get the values of the first row: %v", err)
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
			log.Fatal("Couldn't update the header row: %v", err)
		}
	}

	ensureExistsWithHeaders("Interest", []interface{}{"Student or Teacher", "Email"})
}

func (app *Application) RegisterInterest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, POST, OPTIONS")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPut {
		log.Error("register_interest must be PUT")
		http.Error(w, "register_interest must be PUT", http.StatusBadRequest)
		return
	}
	var interestForm InterestForm
	err := json.NewDecoder(r.Body).Decode(&interestForm)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !app.EmailRegex.MatchString(interestForm.Email) {
		log.Error("Invalid email address")
		http.Error(w, "Invalid email address", http.StatusBadRequest)
	}

	// Now, actually add the user to the spreadsheet
	_, err = app.Client.Spreadsheets.Values.Append(
		app.Configuration.SpreadsheetID,
		"Interest",
		&sheets.ValueRange{
			Values: [][]interface{}{
				{interestForm.IAmA, interestForm.Email},
			},
		}).ValueInputOption("RAW").Do()
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (app *Application) RegisterHandlers() {
	http.HandleFunc("/register_interest", app.RegisterInterest)
}
