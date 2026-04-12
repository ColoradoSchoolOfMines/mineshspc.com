package templates_test

import (
	"bytes"
	"context"
	"html/template"
	"io/fs"
	"strings"
	"testing"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/contextkeys"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/templates"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/templates/testhelpers"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/website"
)

func renderOldHome(t *testing.T, registrationEnabled bool) string {
	t.Helper()
	tmpl, err := template.ParseFS(website.TemplateFS, "templates/base.html", "templates/partials/*", "templates/home.html")
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}
	data := map[string]any{
		"PageName":            "home",
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

func renderNewHome(t *testing.T, registrationEnabled bool) string {
	t.Helper()
	ctx := context.WithValue(context.Background(), contextkeys.ContextKeyRegistrationEnabled, registrationEnabled)
	var buf bytes.Buffer
	if err := templates.Home().Render(ctx, &buf); err != nil {
		t.Fatalf("failed to render templ: %v", err)
	}
	return buf.String()
}

// Ensure website.TemplateFS has the expected file
func TestWebsiteTemplateFS(t *testing.T) {
	entries, err := fs.ReadDir(website.TemplateFS, "templates")
	if err != nil {
		t.Fatalf("failed to read template dir: %v", err)
	}
	found := false
	for _, e := range entries {
		if e.Name() == "home.html" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("home.html not found in website.TemplateFS")
	}
}

func TestHomeEquivalence_RegistrationDisabled(t *testing.T) {
	oldHTML := renderOldHome(t, false)
	newHTML := renderNewHome(t, false)

	// Sanity checks
	if !strings.Contains(oldHTML, "not yet open") {
		t.Errorf("old HTML should contain 'not yet open', got:\n%s", oldHTML)
	}
	if !strings.Contains(newHTML, "not yet open") {
		t.Errorf("new HTML should contain 'not yet open', got:\n%s", newHTML)
	}

	testhelpers.CompareHTML(t, oldHTML, newHTML)
}

func TestHomeEquivalence_RegistrationEnabled(t *testing.T) {
	oldHTML := renderOldHome(t, true)
	newHTML := renderNewHome(t, true)

	// Sanity checks
	if !strings.Contains(oldHTML, "Register Now") {
		t.Errorf("old HTML should contain 'Register Now', got:\n%s", oldHTML)
	}
	if !strings.Contains(newHTML, "Register Now") {
		t.Errorf("new HTML should contain 'Register Now', got:\n%s", newHTML)
	}

	testhelpers.CompareHTML(t, oldHTML, newHTML)
}
