package partials_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/database"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/contextkeys"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/templates/partials"
)

func renderNavbar(t *testing.T, pageName partials.PageName, registrationEnabled bool, username string) string {
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

func TestNavbar_NoActivePage_RegistrationDisabled(t *testing.T) {
	html := renderNavbar(t, "", false, "")

	if strings.Contains(html, "id=\"registration-link\"") {
		t.Errorf("HTML should not have registration link when disabled")
	}
	if !strings.Contains(html, "Teacher Login") {
		t.Errorf("HTML should contain Teacher Login")
	}
}

func TestNavbar_HomePage_RegistrationEnabled(t *testing.T) {
	html := renderNavbar(t, partials.PageNameHome, true, "")

	if !strings.Contains(html, "id=\"registration-link\"") {
		t.Errorf("HTML should have registration link when enabled, got:\n%s", html)
	}
	if !strings.Contains(html, "home-link-active") {
		t.Errorf("HTML should have home-link-active class")
	}
}

func TestNavbar_LoggedIn(t *testing.T) {
	html := renderNavbar(t, partials.PageNameHome, true, "Jane Smith")

	if !strings.Contains(html, "Jane Smith") {
		t.Errorf("HTML should contain username")
	}
	if !strings.Contains(html, "Logout") {
		t.Errorf("HTML should contain Logout link")
	}
}
