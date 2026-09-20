package internal

import (
	"net/http"
)

func (a *Application) GetArchiveTemplate(*http.Request) map[string]any {
	return map[string]any{
		"YearInfo": a.Config.Archive,
	}
}
