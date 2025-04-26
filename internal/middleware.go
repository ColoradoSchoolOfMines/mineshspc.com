package internal

import "net/http"

func (a *Application) AdminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tok, err := r.Cookie("admin_token")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		if isAdmin, err := a.isAdminByToken(tok.Value); err != nil || !isAdmin {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}
