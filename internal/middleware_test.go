package internal

import (
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
