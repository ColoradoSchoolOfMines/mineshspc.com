package internal

import "net/http"

func (a *Application) GetTeacherSchoolInfoTemplate(r *http.Request) map[string]any {
	user, err := a.GetLoggedInTeacher(r)
	if err != nil {
		a.Log.Warn().Err(err).Msg("Failed to get logged in user")
		return nil
	}
	a.Log.Info().Interface("user", user).Msg("found user")

	validated := user.SchoolName != "" || user.SchoolCity != "" || user.SchoolState != ""
	return map[string]any{
		"Username":    user.Name,
		"Validated":   validated,
		"SchoolName":  user.SchoolName,
		"SchoolCity":  user.SchoolCity,
		"SchoolState": user.SchoolState,
	}
}

func (a *Application) HandleTeacherSchoolInfo(w http.ResponseWriter, r *http.Request) {
	log := a.Log.With().Str("page_name", "teacher_school_info").Logger()
	user, err := a.GetLoggedInTeacher(r)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get logged in user")
		http.Redirect(w, r, "/register/teacher/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Err(err).Msg("failed to parse form")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	schoolName := r.FormValue("school-name")
	schoolCity := r.FormValue("school-city")
	schoolState := r.FormValue("school-state")

	if schoolName == "" || schoolCity == "" || schoolState == "" {
		a.ConfirmEmailRenderer(w, r, map[string]any{
			"Errors": map[string]any{
				"school-name":  schoolName == "",
				"school-city":  schoolCity == "",
				"school-state": schoolState == "",
			},
		})
		return
	}

	err = a.DB.SetTeacherSchoolInfo(r.Context(), user.Email, schoolName, schoolCity, schoolState)
	if err != nil {
		a.ConfirmEmailRenderer(w, r, map[string]any{
			"Errors": map[string]any{
				"general": "Failed to save school info. Please try again.",
			},
		})
		return
	}

	http.Redirect(w, r, "/register/teacher/teams", http.StatusSeeOther)
}
