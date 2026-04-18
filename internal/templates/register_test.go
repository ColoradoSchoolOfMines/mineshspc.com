package templates_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/contextkeys"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/templates"
)

func renderRegister(t *testing.T, registrationEnabled bool) string {
	t.Helper()
	ctx := context.WithValue(context.Background(), contextkeys.ContextKeyRegistrationEnabled, registrationEnabled)
	var buf bytes.Buffer
	if err := templates.Register().Render(ctx, &buf); err != nil {
		t.Fatalf("failed to render templ: %v", err)
	}
	return buf.String()
}

func TestRegister_RegistrationDisabled(t *testing.T) {
	html := renderRegister(t, false)

	if !strings.Contains(html, "Registration is currently disabled") {
		t.Errorf("HTML should contain disabled alert, got:\n%s", html)
	}
}

func TestRegister_RegistrationEnabled(t *testing.T) {
	html := renderRegister(t, true)

	if strings.Contains(html, "Registration is currently disabled") {
		t.Errorf("HTML should NOT contain disabled alert when enabled")
	}
	if !strings.Contains(html, "Register A Team") {
		t.Errorf("HTML should contain register button, got:\n%s", html)
	}
}
