package teacher_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/templates/register/teacher"
)

func renderTeacherLogin(t *testing.T, email string, errComponent templ.Component) string {
	t.Helper()
	ctx := context.Background()
	var buf bytes.Buffer
	if err := teacher.Login(email, errComponent).Render(ctx, &buf); err != nil {
		t.Fatalf("failed to render templ: %v", err)
	}
	return buf.String()
}

func TestTeacherLogin_NoError(t *testing.T) {
	html := renderTeacherLogin(t, "", nil)

	if !strings.Contains(html, "Don&#39;t have an account") && !strings.Contains(html, "Don't have an account") {
		t.Errorf("HTML should contain 'Don't have an account' message, got:\n%s", html)
	}
}

func TestTeacherLogin_EmailNotFound(t *testing.T) {
	html := renderTeacherLogin(t, "test@example.com", teacher.LoginEmailDoesNotExist())

	if !strings.Contains(html, "alert-danger") {
		t.Errorf("HTML should contain danger alert")
	}
	if !strings.Contains(html, "doesn&#39;t exist") && !strings.Contains(html, "doesn't exist") {
		t.Errorf("HTML should contain doesn't exist message, got:\n%s", html)
	}
}

func TestTeacherLogin_EmailNotConfirmed(t *testing.T) {
	html := renderTeacherLogin(t, "test@example.com", teacher.LoginEmailNotConfirmed())

	if !strings.Contains(html, "alert-danger") {
		t.Errorf("HTML should contain danger alert")
	}
	if !strings.Contains(html, "confirmed") {
		t.Errorf("HTML should contain 'confirmed' text")
	}
}
