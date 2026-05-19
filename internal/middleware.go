package internal

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
)

func (a *Application) parseTokenByIssuer(tokenStr string, issuer Issuer) (bool, error) {
	if tokenStr == "" {
		return false, fmt.Errorf("no token")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.Config.ReadSecretKey(), nil
	})
	if err != nil {
		return false, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !token.Valid || !ok {
		return false, fmt.Errorf("invalid token")
	}

	if claims.Issuer != string(issuer) {
		return false, fmt.Errorf("wrong issuer: got %s, want %s", claims.Issuer, issuer)
	}

	return true, nil
}

func (a *Application) AdminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tok, err := r.Cookie("admin_token")
		if err != nil {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}

		if isAdmin, err := a.parseTokenByIssuer(tok.Value, IssuerAdminLogin); err != nil || !isAdmin {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *Application) VolunteerAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tok, err := r.Cookie("volunteer_token")
		if err != nil {
			http.Redirect(w, r, "/volunteer/login", http.StatusSeeOther)
			return
		}

		if isVolunteer, err := a.parseTokenByIssuer(tok.Value, IssuerVolunteerLogin); err != nil || !isVolunteer {
			http.Redirect(w, r, "/volunteer/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}
