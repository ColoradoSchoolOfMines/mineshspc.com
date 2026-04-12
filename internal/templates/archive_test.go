package templates_test

import (
	"bytes"
	"context"
	"html/template"
	"strings"
	"testing"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/templates"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/templates/testhelpers"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/website"
)

// minimalArchiveFixture is a small fixture for testing: 1 year, 2 result categories, 2 teams each.
var minimalArchiveFixture = []templates.YearInfo{
	{
		Year: 2023,
		RecapParagraphs: []string{
			"This is the first recap paragraph.",
			"This is the second recap paragraph.",
		},
		Links: []templates.Link{
			{URL: "/static/solutions.pdf", Title: "Solution Slides"},
			{URL: "https://kattis.com/problems", Title: "Problems"},
		},
		Results: []templates.CompetitionResult{
			{
				Name:      "Advanced Division",
				Shortname: "Advanced",
				Teams: []templates.WinningTeam{
					{Place: "1st", Name: "Team Alpha", School: "Alpha High School", Location: "Denver, CO"},
					{Place: "2nd", Name: "Team Beta", School: "Beta Academy", Location: "Boulder, CO"},
				},
			},
			{
				Name:      "Beginner Division",
				Shortname: "Beginner",
				Teams: []templates.WinningTeam{
					{Place: "1st", Name: "Team Gamma", School: "Gamma School", Location: "Fort Collins, CO"},
					{Place: "2nd", Name: "Team Delta", School: "Delta Institute", Location: "Aurora, CO"},
				},
			},
		},
	},
}

// htmlArchiveFixture mirrors minimalArchiveFixture but uses plain string URLs for the HTML template.
var htmlArchiveFixture = map[string]any{
	"YearInfo": []map[string]any{
		{
			"Year": 2023,
			"RecapParagraphs": []string{
				"This is the first recap paragraph.",
				"This is the second recap paragraph.",
			},
			"Links": []map[string]any{
				{"URL": "/static/solutions.pdf", "Title": "Solution Slides"},
				{"URL": "https://kattis.com/problems", "Title": "Problems"},
			},
			"Results": []map[string]any{
				{
					"Name":      "Advanced Division",
					"Shortname": "Advanced",
					"Teams": []map[string]any{
						{"Place": "1st", "Name": "Team Alpha", "School": "Alpha High School", "Location": "Denver, CO"},
						{"Place": "2nd", "Name": "Team Beta", "School": "Beta Academy", "Location": "Boulder, CO"},
					},
				},
				{
					"Name":      "Beginner Division",
					"Shortname": "Beginner",
					"Teams": []map[string]any{
						{"Place": "1st", "Name": "Team Gamma", "School": "Gamma School", "Location": "Fort Collins, CO"},
						{"Place": "2nd", "Name": "Team Delta", "School": "Delta Institute", "Location": "Aurora, CO"},
					},
				},
			},
		},
	},
}

func renderOldArchive(t *testing.T) string {
	t.Helper()
	tmpl, err := template.ParseFS(website.TemplateFS, "templates/base.html", "templates/partials/*", "templates/archive.html")
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}
	data := map[string]any{
		"PageName":            "archive",
		"Data":                htmlArchiveFixture,
		"HostedByHTML":        template.HTML("CS@Mines"),
		"RegistrationEnabled": false,
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "content", data); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}
	return buf.String()
}

func renderNewArchive(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	var buf bytes.Buffer
	if err := templates.Archive(minimalArchiveFixture).Render(ctx, &buf); err != nil {
		t.Fatalf("failed to render templ: %v", err)
	}
	return buf.String()
}

func TestArchiveEquivalence(t *testing.T) {
	oldHTML := renderOldArchive(t)
	newHTML := renderNewArchive(t)

	// Sanity checks
	if !strings.Contains(oldHTML, "2023") {
		t.Errorf("old HTML should contain year 2023")
	}
	if !strings.Contains(newHTML, "2023") {
		t.Errorf("new HTML should contain year 2023")
	}
	if !strings.Contains(oldHTML, "Team Alpha") {
		t.Errorf("old HTML should contain Team Alpha")
	}
	if !strings.Contains(newHTML, "Team Alpha") {
		t.Errorf("new HTML should contain Team Alpha")
	}
	if !strings.Contains(oldHTML, "Advanced Division Competition Winners") {
		t.Errorf("old HTML should contain accordion header, got:\n%s", oldHTML)
	}
	if !strings.Contains(newHTML, "Advanced Division Competition Winners") {
		t.Errorf("new HTML should contain accordion header")
	}

	testhelpers.CompareHTML(t, oldHTML, newHTML)
}
