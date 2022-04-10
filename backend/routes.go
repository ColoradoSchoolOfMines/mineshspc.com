package main

import (
	"encoding/json"
	"net/http"
	"regexp"
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
	EmailRegex *regexp.Regexp
}

func NewApplication() *Application {
	emailRegex, _ := regexp.Compile(`(?i)^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$`)
	return &Application{
		EmailRegex: emailRegex,
	}
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
		http.Error(w, "register_interest must be PUT", http.StatusBadRequest)
		return
	}
	var interestForm InterestForm
	err := json.NewDecoder(r.Body).Decode(&interestForm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !app.EmailRegex.MatchString(interestForm.Email) {
		http.Error(w, "Invalid email address", http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	app := NewApplication()
	http.HandleFunc("/register_interest", app.RegisterInterest)
	http.ListenAndServe(":8090", nil)
}
