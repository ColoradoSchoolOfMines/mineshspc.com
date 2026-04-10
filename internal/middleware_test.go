package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	require.NoError(t, err)
	return signed
}

// --- parseTokenByIssuer ---

func TestParseTokenByIssuer_Empty(t *testing.T) {
	a := newTestApp()
	ok, err := a.parseTokenByIssuer("", IssuerAdminLogin)
	assert.False(t, ok)
	assert.Error(t, err)
}

func TestParseTokenByIssuer_CorrectIssuer(t *testing.T) {
	a := newTestApp()
	tok := makeSignedToken(t, IssuerAdminLogin, []byte(testSecretKey), time.Now().Add(time.Hour))
	ok, err := a.parseTokenByIssuer(tok, IssuerAdminLogin)
	assert.True(t, ok)
	assert.NoError(t, err)
}

func TestParseTokenByIssuer_WrongIssuer(t *testing.T) {
	a := newTestApp()
	tok := makeSignedToken(t, IssuerVolunteerLogin, []byte(testSecretKey), time.Now().Add(time.Hour))
	ok, err := a.parseTokenByIssuer(tok, IssuerAdminLogin)
	assert.False(t, ok)
	assert.Error(t, err)
}

func TestParseTokenByIssuer_ExpiredToken(t *testing.T) {
	a := newTestApp()
	tok := makeSignedToken(t, IssuerAdminLogin, []byte(testSecretKey), time.Now().Add(-time.Hour))
	ok, err := a.parseTokenByIssuer(tok, IssuerAdminLogin)
	assert.False(t, ok)
	assert.Error(t, err)
}

func TestParseTokenByIssuer_WrongSigningKey(t *testing.T) {
	a := newTestApp()
	tok := makeSignedToken(t, IssuerAdminLogin, []byte("different-key"), time.Now().Add(time.Hour))
	ok, err := a.parseTokenByIssuer(tok, IssuerAdminLogin)
	assert.False(t, ok)
	assert.Error(t, err)
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
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/admin/login", rec.Header().Get("Location"))
}

func TestAdminAuthMiddleware_InvalidToken(t *testing.T) {
	a := newTestApp()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/teams", nil)
	req.AddCookie(&http.Cookie{Name: "admin_token", Value: "not-a-jwt"})
	a.AdminAuthMiddleware(okHandler).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/admin/login", rec.Header().Get("Location"))
}

func TestAdminAuthMiddleware_WrongIssuer(t *testing.T) {
	a := newTestApp()
	tok := makeSignedToken(t, IssuerVolunteerLogin, []byte(testSecretKey), time.Now().Add(time.Hour))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/teams", nil)
	req.AddCookie(&http.Cookie{Name: "admin_token", Value: tok})
	a.AdminAuthMiddleware(okHandler).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/admin/login", rec.Header().Get("Location"))
}

func TestAdminAuthMiddleware_ValidToken(t *testing.T) {
	a := newTestApp()
	tok := makeSignedToken(t, IssuerAdminLogin, []byte(testSecretKey), time.Now().Add(time.Hour))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/teams", nil)
	req.AddCookie(&http.Cookie{Name: "admin_token", Value: tok})
	a.AdminAuthMiddleware(okHandler).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// --- VolunteerAuthMiddleware ---

func TestVolunteerAuthMiddleware_NoCookie(t *testing.T) {
	a := newTestApp()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/volunteer/scan", nil)
	a.VolunteerAuthMiddleware(okHandler).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/volunteer/login", rec.Header().Get("Location"))
}

func TestVolunteerAuthMiddleware_InvalidToken(t *testing.T) {
	a := newTestApp()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/volunteer/scan", nil)
	req.AddCookie(&http.Cookie{Name: "volunteer_token", Value: "not-a-jwt"})
	a.VolunteerAuthMiddleware(okHandler).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/volunteer/login", rec.Header().Get("Location"))
}

func TestVolunteerAuthMiddleware_WrongIssuer(t *testing.T) {
	a := newTestApp()
	tok := makeSignedToken(t, IssuerAdminLogin, []byte(testSecretKey), time.Now().Add(time.Hour))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/volunteer/scan", nil)
	req.AddCookie(&http.Cookie{Name: "volunteer_token", Value: tok})
	a.VolunteerAuthMiddleware(okHandler).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/volunteer/login", rec.Header().Get("Location"))
}

func TestVolunteerAuthMiddleware_ValidToken(t *testing.T) {
	a := newTestApp()
	tok := makeSignedToken(t, IssuerVolunteerLogin, []byte(testSecretKey), time.Now().Add(time.Hour))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/volunteer/scan", nil)
	req.AddCookie(&http.Cookie{Name: "volunteer_token", Value: tok})
	a.VolunteerAuthMiddleware(okHandler).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}
