package templates_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/contextkeys"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/templates"
)

func renderHome(t *testing.T, registrationEnabled bool) string {
	t.Helper()
	ctx := context.WithValue(context.Background(), contextkeys.ContextKeyRegistrationEnabled, registrationEnabled)
	var buf bytes.Buffer
	if err := templates.Home().Render(ctx, &buf); err != nil {
		t.Fatalf("failed to render templ: %v", err)
	}
	return buf.String()
}

func TestHome_RegistrationDisabled(t *testing.T) {
	html := renderHome(t, false)

	if !strings.Contains(html, "not yet open") {
		t.Errorf("HTML should contain 'not yet open', got:\n%s", html)
	}
}

func TestHome_RegistrationEnabled(t *testing.T) {
	html := renderHome(t, true)

	if !strings.Contains(html, "Register Now") {
		t.Errorf("HTML should contain 'Register Now', got:\n%s", html)
	}
}
