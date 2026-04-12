package partials_test

import (
	"bytes"
	"context"
	"html/template"
	"strings"
	"testing"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/database"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/contextkeys"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/templates/partials"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/templates/testhelpers"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/website"
)

func renderOldNavbar(t *testing.T, pageName string, registrationEnabled bool, username string) string {
	t.Helper()
	tmpl, err := template.ParseFS(website.TemplateFS, "templates/base.html", "templates/partials/*")
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}
	data := map[string]any{
		"PageName":            pageName,
		"Data":                map[string]any{"Username": username},
		"HostedByHTML":        template.HTML("CS@Mines"),
		"RegistrationEnabled": registrationEnabled,
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "navbar", data); err != nil {
		t.Fatalf("failed to execute navbar template: %v", err)
	}
	return buf.String()
}

func renderNewNavbar(t *testing.T, pageName partials.PageName, registrationEnabled bool, username string) string {
	t.Helper()
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.ContextKeyPageName, pageName)
	ctx = context.WithValue(ctx, contextkeys.ContextKeyRegistrationEnabled, registrationEnabled)
	if username != "" {
		ctx = context.WithValue(ctx, contextkeys.ContextKeyLoggedInTeacher, &database.Teacher{Name: username})
	}
	var buf bytes.Buffer
	if err := partials.Navbar().Render(ctx, &buf); err != nil {
		t.Fatalf("failed to render navbar: %v", err)
	}
	return buf.String()
}

func TestNavbarEquivalence_NoActivePage_RegistrationDisabled_NotLoggedIn(t *testing.T) {
	oldHTML := renderOldNavbar(t, "", false, "")
	newHTML := renderNewNavbar(t, "", false, "")

	// Should not contain register link
	if strings.Contains(oldHTML, "id=\"registration-link\"") {
		t.Errorf("old HTML should not have registration link when disabled")
	}
	if strings.Contains(newHTML, "id=\"registration-link\"") {
		t.Errorf("new HTML should not have registration link when disabled")
	}
	// Should contain teacher login
	if !strings.Contains(oldHTML, "Teacher Login") {
		t.Errorf("old HTML should contain Teacher Login")
	}
	if !strings.Contains(newHTML, "Teacher Login") {
		t.Errorf("new HTML should contain Teacher Login")
	}

	testhelpers.CompareHTML(t, oldHTML, newHTML)
}

func TestNavbarEquivalence_HomePage_RegistrationEnabled(t *testing.T) {
	oldHTML := renderOldNavbar(t, "home", true, "")
	newHTML := renderNewNavbar(t, partials.PageNameHome, true, "")

	// Should contain register link
	if !strings.Contains(oldHTML, "id=\"registration-link\"") {
		t.Errorf("old HTML should have registration link when enabled, got:\n%s", oldHTML)
	}
	if !strings.Contains(newHTML, "id=\"registration-link\"") {
		t.Errorf("new HTML should have registration link when enabled")
	}
	// Home link should be active
	if !strings.Contains(oldHTML, "home-link-active") {
		t.Errorf("old HTML should have home-link-active class")
	}
	if !strings.Contains(newHTML, "home-link-active") {
		t.Errorf("new HTML should have home-link-active class")
	}

	testhelpers.CompareHTML(t, oldHTML, newHTML)
}

func TestNavbarEquivalence_LoggedIn(t *testing.T) {
	oldHTML := renderOldNavbar(t, "home", true, "Jane Smith")
	newHTML := renderNewNavbar(t, partials.PageNameHome, true, "Jane Smith")

	if !strings.Contains(oldHTML, "Jane Smith") {
		t.Errorf("old HTML should contain username")
	}
	if !strings.Contains(newHTML, "Jane Smith") {
		t.Errorf("new HTML should contain username")
	}
	if !strings.Contains(oldHTML, "Logout") {
		t.Errorf("old HTML should contain Logout link")
	}
	if !strings.Contains(newHTML, "Logout") {
		t.Errorf("new HTML should contain Logout link")
	}

	testhelpers.CompareHTML(t, oldHTML, newHTML)
}
