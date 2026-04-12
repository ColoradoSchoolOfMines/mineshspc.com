package templates_test

import (
	"bytes"
	"context"
	"html/template"
	"strings"
	"testing"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/contextkeys"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/templates"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/templates/testhelpers"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/website"
)

func renderOldRegister(t *testing.T, registrationEnabled bool) string {
	t.Helper()
	tmpl, err := template.ParseFS(website.TemplateFS, "templates/base.html", "templates/partials/*", "templates/register.html")
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}
	data := map[string]any{
		"PageName":            "register",
		"Data":                map[string]any{},
		"HostedByHTML":        template.HTML("CS@Mines"),
		"RegistrationEnabled": registrationEnabled,
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "content", data); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}
	return buf.String()
}

func renderNewRegister(t *testing.T, registrationEnabled bool) string {
	t.Helper()
	ctx := context.WithValue(context.Background(), contextkeys.ContextKeyRegistrationEnabled, registrationEnabled)
	var buf bytes.Buffer
	if err := templates.Register().Render(ctx, &buf); err != nil {
		t.Fatalf("failed to render templ: %v", err)
	}
	return buf.String()
}

func TestRegisterEquivalence_RegistrationDisabled(t *testing.T) {
	oldHTML := renderOldRegister(t, false)
	newHTML := renderNewRegister(t, false)

	if !strings.Contains(oldHTML, "Registration is currently disabled") {
		t.Errorf("old HTML should contain disabled alert, got:\n%s", oldHTML)
	}
	if !strings.Contains(newHTML, "Registration is currently disabled") {
		t.Errorf("new HTML should contain disabled alert, got:\n%s", newHTML)
	}

	testhelpers.CompareHTML(t, oldHTML, newHTML)
}

func TestRegisterEquivalence_RegistrationEnabled(t *testing.T) {
	oldHTML := renderOldRegister(t, true)
	newHTML := renderNewRegister(t, true)

	if strings.Contains(oldHTML, "Registration is currently disabled") {
		t.Errorf("old HTML should NOT contain disabled alert when enabled")
	}
	if strings.Contains(newHTML, "Registration is currently disabled") {
		t.Errorf("new HTML should NOT contain disabled alert when enabled")
	}
	if !strings.Contains(oldHTML, "Register A Team") {
		t.Errorf("old HTML should contain register button, got:\n%s", oldHTML)
	}
	if !strings.Contains(newHTML, "Register A Team") {
		t.Errorf("new HTML should contain register button, got:\n%s", newHTML)
	}

	testhelpers.CompareHTML(t, oldHTML, newHTML)
}
