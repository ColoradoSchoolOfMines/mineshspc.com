package teacher_test

import (
	"bytes"
	"context"
	"html/template"
	"strings"
	"testing"

	"github.com/a-h/templ"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/templates/register/teacher"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/templates/testhelpers"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/website"
)

func renderOldTeacherLogin(t *testing.T, data map[string]any) string {
	t.Helper()
	tmpl, err := template.ParseFS(website.TemplateFS, "templates/base.html", "templates/partials/*", "templates/teacherlogin.html")
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}
	outerData := map[string]any{
		"PageName":            "register",
		"Data":                data,
		"HostedByHTML":        template.HTML("CS@Mines"),
		"RegistrationEnabled": false,
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "content", outerData); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}
	return buf.String()
}

func renderNewTeacherLogin(t *testing.T, email string, errComponent templ.Component) string {
	t.Helper()
	ctx := context.Background()
	var buf bytes.Buffer
	if err := teacher.Login(email, errComponent).Render(ctx, &buf); err != nil {
		t.Fatalf("failed to render templ: %v", err)
	}
	return buf.String()
}

func TestTeacherLoginEquivalence_NoError(t *testing.T) {
	oldHTML := renderOldTeacherLogin(t, map[string]any{})
	newHTML := renderNewTeacherLogin(t, "", nil)

	if !strings.Contains(oldHTML, "Don&#39;t have an account") && !strings.Contains(oldHTML, "Don't have an account") {
		t.Errorf("old HTML should contain 'Don't have an account' message, got:\n%s", oldHTML)
	}
	if !strings.Contains(newHTML, "Don&#39;t have an account") && !strings.Contains(newHTML, "Don't have an account") {
		t.Errorf("new HTML should contain 'Don't have an account' message")
	}

	testhelpers.CompareHTML(t, oldHTML, newHTML)
}

func TestTeacherLoginEquivalence_EmailNotFound(t *testing.T) {
	oldHTML := renderOldTeacherLogin(t, map[string]any{
		"EmailNotFound": true,
		"Email":         "test@example.com",
	})
	newHTML := renderNewTeacherLogin(t, "test@example.com", teacher.LoginEmailDoesNotExist())

	if !strings.Contains(oldHTML, "alert-danger") {
		t.Errorf("old HTML should contain danger alert")
	}
	if !strings.Contains(newHTML, "alert-danger") {
		t.Errorf("new HTML should contain danger alert")
	}
	if !strings.Contains(oldHTML, "doesn&#39;t exist") && !strings.Contains(oldHTML, "doesn't exist") {
		t.Errorf("old HTML should contain doesn't exist message, got:\n%s", oldHTML)
	}

	testhelpers.CompareHTML(t, oldHTML, newHTML)
}

func TestTeacherLoginEquivalence_EmailNotConfirmed(t *testing.T) {
	oldHTML := renderOldTeacherLogin(t, map[string]any{
		"EmailNotConfirmed": true,
		"Email":             "test@example.com",
	})
	newHTML := renderNewTeacherLogin(t, "test@example.com", teacher.LoginEmailNotConfirmed())

	if !strings.Contains(oldHTML, "alert-danger") {
		t.Errorf("old HTML should contain danger alert")
	}
	if !strings.Contains(newHTML, "alert-danger") {
		t.Errorf("new HTML should contain danger alert")
	}
	if !strings.Contains(oldHTML, "confirmed") {
		t.Errorf("old HTML should contain 'confirmed' text")
	}
	if !strings.Contains(newHTML, "confirmed") {
		t.Errorf("new HTML should contain 'confirmed' text")
	}

	testhelpers.CompareHTML(t, oldHTML, newHTML)
}
