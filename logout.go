package main

import (
	"net/http"
	"time"
)

func (a *Application) HandleTeacherLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "session_id", Value: "", Path: "/", Expires: time.Unix(0, 0)})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
