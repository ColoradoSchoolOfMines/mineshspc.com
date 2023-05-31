package internal

import (
	"net/http"
	"time"
)

func (a *Application) HandleTeacherLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "tok", Value: "", Path: "/", Expires: time.Unix(0, 0)})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
