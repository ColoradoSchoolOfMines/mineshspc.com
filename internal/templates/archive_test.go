package templates_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/templates"
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

func TestArchive(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer
	if err := templates.Archive(minimalArchiveFixture).Render(ctx, &buf); err != nil {
		t.Fatalf("failed to render templ: %v", err)
	}
	html := buf.String()

	if !strings.Contains(html, "2023") {
		t.Errorf("HTML should contain year 2023")
	}
	if !strings.Contains(html, "Team Alpha") {
		t.Errorf("HTML should contain Team Alpha")
	}
	if !strings.Contains(html, "Advanced Division Competition Winners") {
		t.Errorf("HTML should contain accordion header, got:\n%s", html)
	}
}
