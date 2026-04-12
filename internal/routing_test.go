package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mau.fi/util/dbutil"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/database"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/config"
)

func newTestAppWithDB(t *testing.T) *Application {
	t.Helper()
	log := zerolog.Nop()

	rawDB, err := dbutil.NewFromConfig("mineshspc", dbutil.Config{
		PoolConfig: dbutil.PoolConfig{
			Type:         "sqlite3",
			URI:          ":memory:",
			MaxOpenConns: 1,
			MaxIdleConns: 1,
		},
	}, dbutil.ZeroLogger(zerolog.Nop()))
	require.NoError(t, err)

	db := database.NewDatabase(rawDB)
	require.NoError(t, db.DB.Upgrade(context.Background()))

	cfg := config.Configuration{}
	cfg.JWTSecretKey = testSecretKey
	return NewApplication(&log, cfg, db)
}

func adminToken(t *testing.T) string {
	t.Helper()
	return makeSignedToken(t, IssuerAdminLogin, []byte(testSecretKey), time.Now().Add(time.Hour))
}

func volunteerToken(t *testing.T) string {
	t.Helper()
	return makeSignedToken(t, IssuerVolunteerLogin, []byte(testSecretKey), time.Now().Add(time.Hour))
}

func doRequest(router http.Handler, method, path string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	router.ServeHTTP(rec, req)
	return rec
}

func assertRedirectsTo(t *testing.T, rec *httptest.ResponseRecorder, location string) {
	t.Helper()
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, location, rec.Header().Get("Location"))
}

func assertNotRedirectTo(t *testing.T, rec *httptest.ResponseRecorder, location string) {
	t.Helper()
	if rec.Code == http.StatusSeeOther || rec.Code == http.StatusFound || rec.Code == http.StatusTemporaryRedirect {
		assert.NotEqual(t, location, rec.Header().Get("Location"),
			"expected route NOT to redirect to %s but it did", location)
	}
}

// --- Teacher route protection ---

func TestRouting_Teacher_NoCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	for _, path := range []string{
		"/register/teacher/schoolinfo",
		"/register/teacher/teams",
		"/register/teacher/team/edit",
		"/register/teacher/team/addmember",
	} {
		assertRedirectsTo(t, doRequest(router, http.MethodGet, path), "/register/teacher/login")
	}
}

// --- Admin route protection ---
//
// Tests cover: no cookie, malformed token, wrong-issuer token (volunteer
// token used on admin route), and valid token. Each case is verified on
// one representative route; the remaining admin routes are spot-checked
// for the no-cookie case to confirm they're all wired up.

func TestRouting_Admin_NoCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	for _, path := range []string{
		"/admin",
		"/admin/teams",
		"/admin/dietaryrestrictions",
		"/admin/api/resendstudentemail",
		"/admin/api/sendqrcodes",
		"/admin/api/kattis/teams",
		"/admin/api/zoom/breakout",
		"/admin/api/team-list",
	} {
		assertRedirectsTo(t, doRequest(router, http.MethodGet, path), "/admin/login")
	}
}

func TestRouting_Admin_MalformedToken(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	rec := doRequest(router, http.MethodGet, "/admin/teams",
		&http.Cookie{Name: "admin_token", Value: "not-a-jwt"})
	assertRedirectsTo(t, rec, "/admin/login")
}

func TestRouting_Admin_WrongIssuer(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	rec := doRequest(router, http.MethodGet, "/admin/teams",
		&http.Cookie{Name: "admin_token", Value: volunteerToken(t)})
	assertRedirectsTo(t, rec, "/admin/login")
}

func TestRouting_Admin_ValidToken(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	for _, path := range []string{
		"/admin",
		"/admin/teams",
		"/admin/dietaryrestrictions",
		"/admin/api/kattis/teams",
		"/admin/api/zoom/breakout",
		"/admin/api/team-list",
	} {
		rec := doRequest(router, http.MethodGet, path,
			&http.Cookie{Name: "admin_token", Value: adminToken(t)})
		assertNotRedirectTo(t, rec, "/admin/login")
	}
}

// --- Admin unprotected routes ---

func TestRouting_AdminLogin_NoCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	assertNotRedirectTo(t, doRequest(router, http.MethodGet, "/admin/login"), "/admin/login")
}

// --- Volunteer route protection ---
//
// Same structure: all protected routes checked for no-cookie, then one
// representative route for malformed/wrong-issuer/valid-token.

func TestRouting_Volunteer_NoCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	for _, path := range []string{
		"/volunteer/scan",
		"/volunteer/checkin",
	} {
		assertRedirectsTo(t, doRequest(router, http.MethodGet, path), "/volunteer/login")
	}
}

func TestRouting_Volunteer_MalformedToken(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	rec := doRequest(router, http.MethodGet, "/volunteer/scan",
		&http.Cookie{Name: "volunteer_token", Value: "not-a-jwt"})
	assertRedirectsTo(t, rec, "/volunteer/login")
}

func TestRouting_Volunteer_WrongIssuer(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	rec := doRequest(router, http.MethodGet, "/volunteer/scan",
		&http.Cookie{Name: "volunteer_token", Value: adminToken(t)})
	assertRedirectsTo(t, rec, "/volunteer/login")
}

func TestRouting_Volunteer_ValidToken(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	for _, path := range []string{
		"/volunteer/scan",
		"/volunteer/checkin",
	} {
		rec := doRequest(router, http.MethodGet, path,
			&http.Cookie{Name: "volunteer_token", Value: volunteerToken(t)})
		assertNotRedirectTo(t, rec, "/volunteer/login")
	}
}

// --- Volunteer unprotected routes ---

func TestRouting_VolunteerPublic_NoCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	for _, path := range []string{"/volunteer", "/volunteer/login"} {
		assertNotRedirectTo(t, doRequest(router, http.MethodGet, path), "/volunteer/login")
	}
}
