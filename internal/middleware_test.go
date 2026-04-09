package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/config"
)

const testSecretKey = "test-secret-key"

func newTestApp() *Application {
	log := zerolog.Nop()
	cfg := config.Configuration{}
	cfg.JWTSecretKey = testSecretKey
	return NewApplication(&log, cfg, nil)
}

func makeSignedToken(t *testing.T, issuer Issuer, key []byte, expiry time.Time) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer:    string(issuer),
		Subject:   "test@example.com",
		ExpiresAt: jwt.NewNumericDate(expiry),
	})
	signed, err := tok.SignedString(key)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}

// --- parseTokenByIssuer ---

func TestParseTokenByIssuer_Empty(t *testing.T) {
	a := newTestApp()
	ok, err := a.parseTokenByIssuer("", IssuerAdminLogin)
	if ok || err == nil {
		t.Error("expected error for empty token")
	}
}

func TestParseTokenByIssuer_CorrectIssuer(t *testing.T) {
	a := newTestApp()
	tok := makeSignedToken(t, IssuerAdminLogin, []byte(testSecretKey), time.Now().Add(time.Hour))
	ok, err := a.parseTokenByIssuer(tok, IssuerAdminLogin)
	if !ok || err != nil {
		t.Errorf("expected valid token to pass: %v", err)
	}
}

func TestParseTokenByIssuer_WrongIssuer(t *testing.T) {
	a := newTestApp()
	tok := makeSignedToken(t, IssuerVolunteerLogin, []byte(testSecretKey), time.Now().Add(time.Hour))
	ok, err := a.parseTokenByIssuer(tok, IssuerAdminLogin)
	if ok || err == nil {
		t.Error("expected error for wrong issuer")
	}
}

func TestParseTokenByIssuer_ExpiredToken(t *testing.T) {
	a := newTestApp()
	tok := makeSignedToken(t, IssuerAdminLogin, []byte(testSecretKey), time.Now().Add(-time.Hour))
	ok, err := a.parseTokenByIssuer(tok, IssuerAdminLogin)
	if ok || err == nil {
		t.Error("expected error for expired token")
	}
}

func TestParseTokenByIssuer_WrongSigningKey(t *testing.T) {
	a := newTestApp()
	tok := makeSignedToken(t, IssuerAdminLogin, []byte("different-key"), time.Now().Add(time.Hour))
	ok, err := a.parseTokenByIssuer(tok, IssuerAdminLogin)
	if ok || err == nil {
		t.Error("expected error for token signed with wrong key")
	}
}

// --- AdminAuthMiddleware ---

var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestAdminAuthMiddleware_NoCookie(t *testing.T) {
	a := newTestApp()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/teams", nil)
	a.AdminAuthMiddleware(okHandler).ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/admin/login" {
		t.Errorf("expected redirect to /admin/login, got %q", loc)
	}
}

func TestAdminAuthMiddleware_InvalidToken(t *testing.T) {
	a := newTestApp()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/teams", nil)
	req.AddCookie(&http.Cookie{Name: "admin_token", Value: "not-a-jwt"})
	a.AdminAuthMiddleware(okHandler).ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/admin/login" {
		t.Errorf("expected redirect to /admin/login, got %q", loc)
	}
}

func TestAdminAuthMiddleware_WrongIssuer(t *testing.T) {
	a := newTestApp()
	tok := makeSignedToken(t, IssuerVolunteerLogin, []byte(testSecretKey), time.Now().Add(time.Hour))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/teams", nil)
	req.AddCookie(&http.Cookie{Name: "admin_token", Value: tok})
	a.AdminAuthMiddleware(okHandler).ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
}

func TestAdminAuthMiddleware_ValidToken(t *testing.T) {
	a := newTestApp()
	tok := makeSignedToken(t, IssuerAdminLogin, []byte(testSecretKey), time.Now().Add(time.Hour))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/teams", nil)
	req.AddCookie(&http.Cookie{Name: "admin_token", Value: tok})
	a.AdminAuthMiddleware(okHandler).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// --- VolunteerAuthMiddleware ---

func TestVolunteerAuthMiddleware_NoCookie(t *testing.T) {
	a := newTestApp()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/volunteer/scan", nil)
	a.VolunteerAuthMiddleware(okHandler).ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/volunteer/login" {
		t.Errorf("expected redirect to /volunteer/login, got %q", loc)
	}
}

func TestVolunteerAuthMiddleware_InvalidToken(t *testing.T) {
	a := newTestApp()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/volunteer/scan", nil)
	req.AddCookie(&http.Cookie{Name: "volunteer_token", Value: "not-a-jwt"})
	a.VolunteerAuthMiddleware(okHandler).ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/volunteer/login" {
		t.Errorf("expected redirect to /volunteer/login, got %q", loc)
	}
}

func TestVolunteerAuthMiddleware_WrongIssuer(t *testing.T) {
	a := newTestApp()
	tok := makeSignedToken(t, IssuerAdminLogin, []byte(testSecretKey), time.Now().Add(time.Hour))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/volunteer/scan", nil)
	req.AddCookie(&http.Cookie{Name: "volunteer_token", Value: tok})
	a.VolunteerAuthMiddleware(okHandler).ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
}

func TestVolunteerAuthMiddleware_ValidToken(t *testing.T) {
	a := newTestApp()
	tok := makeSignedToken(t, IssuerVolunteerLogin, []byte(testSecretKey), time.Now().Add(time.Hour))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/volunteer/scan", nil)
	req.AddCookie(&http.Cookie{Name: "volunteer_token", Value: tok})
	a.VolunteerAuthMiddleware(okHandler).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
