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

// --- Admin route protection ---

func TestRouting_AdminHome_NoCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	assertRedirectsTo(t, doRequest(router, http.MethodGet, "/admin"), "/admin/login")
}

func TestRouting_AdminHome_WrongCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	rec := doRequest(router, http.MethodGet, "/admin",
		&http.Cookie{Name: "volunteer_token", Value: volunteerToken(t)})
	assertRedirectsTo(t, rec, "/admin/login")
}

func TestRouting_AdminHome_ValidCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	rec := doRequest(router, http.MethodGet, "/admin",
		&http.Cookie{Name: "admin_token", Value: adminToken(t)})
	assertNotRedirectTo(t, rec, "/admin/login")
}

func TestRouting_AdminTeams_NoCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	assertRedirectsTo(t, doRequest(router, http.MethodGet, "/admin/teams"), "/admin/login")
}

func TestRouting_AdminTeams_WrongCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	rec := doRequest(router, http.MethodGet, "/admin/teams",
		&http.Cookie{Name: "volunteer_token", Value: volunteerToken(t)})
	assertRedirectsTo(t, rec, "/admin/login")
}

func TestRouting_AdminTeams_ValidCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	rec := doRequest(router, http.MethodGet, "/admin/teams",
		&http.Cookie{Name: "admin_token", Value: adminToken(t)})
	assertNotRedirectTo(t, rec, "/admin/login")
}

func TestRouting_AdminDietaryRestrictions_NoCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	assertRedirectsTo(t, doRequest(router, http.MethodGet, "/admin/dietaryrestrictions"), "/admin/login")
}

func TestRouting_AdminDietaryRestrictions_ValidCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	rec := doRequest(router, http.MethodGet, "/admin/dietaryrestrictions",
		&http.Cookie{Name: "admin_token", Value: adminToken(t)})
	assertNotRedirectTo(t, rec, "/admin/login")
}

func TestRouting_AdminAPI_NoCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	// Spot-check a few API endpoints
	for _, path := range []string{
		"/admin/api/resendstudentemail",
		"/admin/api/sendqrcodes",
		"/admin/api/kattis/teams",
	} {
		rec := doRequest(router, http.MethodGet, path)
		assertRedirectsTo(t, rec, "/admin/login")
	}
}

func TestRouting_AdminAPI_ValidCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	for _, path := range []string{
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
	rec := doRequest(router, http.MethodGet, "/admin/login")
	assertNotRedirectTo(t, rec, "/admin/login")
}

// --- Volunteer route protection ---

func TestRouting_VolunteerScan_NoCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	assertRedirectsTo(t, doRequest(router, http.MethodGet, "/volunteer/scan"), "/volunteer/login")
}

func TestRouting_VolunteerScan_WrongCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	rec := doRequest(router, http.MethodGet, "/volunteer/scan",
		&http.Cookie{Name: "admin_token", Value: adminToken(t)})
	assertRedirectsTo(t, rec, "/volunteer/login")
}

func TestRouting_VolunteerScan_ValidCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	rec := doRequest(router, http.MethodGet, "/volunteer/scan",
		&http.Cookie{Name: "volunteer_token", Value: volunteerToken(t)})
	assertNotRedirectTo(t, rec, "/volunteer/login")
}

func TestRouting_VolunteerCheckin_NoCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	assertRedirectsTo(t, doRequest(router, http.MethodGet, "/volunteer/checkin"), "/volunteer/login")
}

func TestRouting_VolunteerCheckin_WrongCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	rec := doRequest(router, http.MethodGet, "/volunteer/checkin",
		&http.Cookie{Name: "admin_token", Value: adminToken(t)})
	assertRedirectsTo(t, rec, "/volunteer/login")
}

func TestRouting_VolunteerCheckin_ValidCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	rec := doRequest(router, http.MethodGet, "/volunteer/checkin",
		&http.Cookie{Name: "volunteer_token", Value: volunteerToken(t)})
	assertNotRedirectTo(t, rec, "/volunteer/login")
}

// --- Volunteer unprotected routes ---

func TestRouting_VolunteerLogin_NoCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	rec := doRequest(router, http.MethodGet, "/volunteer/login")
	assertNotRedirectTo(t, rec, "/volunteer/login")
}

func TestRouting_VolunteerHome_NoCookie(t *testing.T) {
	router := newTestAppWithDB(t).BuildRouter()
	rec := doRequest(router, http.MethodGet, "/volunteer")
	assertNotRedirectTo(t, rec, "/volunteer/login")
}
