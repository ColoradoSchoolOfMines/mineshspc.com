package main

import (
	"net/http"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/database"
	"github.com/google/uuid"
)

func (a *Application) GetTeacherTeamsTemplate(r *http.Request) map[string]any {
	user, err := a.GetLoggedInTeacher(r)
	if err != nil {
		a.Log.Error().Err(err).Msg("Failed to get logged in user")
		return nil
	}
	a.Log.Info().Interface("user", user).Msg("found user")

	teams, err := a.DB.GetTeacherTeams(user.Email)
	if err != nil {
		a.Log.Error().Err(err).Msg("Failed to get teacher teams")
		// TODO report this error to the user and email admin
		return nil
	}

	teams = append(teams, database.Team{
		ID:           uuid.New(),
		TeacherEmail: "ohea@ohea.com",
		Name:         "Cool Team",
		Division:     "Advanced",
		InPerson:     true,
		Members: []database.Student{
			{
				TeamID:                  uuid.New(),
				Email:                   "test.student@somewhere.edu",
				Name:                    "Test Student",
				ParentEmail:             "parent@coolcompany.com",
				PreviouslyParticipated:  false,
				EmailConfirmed:          true,
				LiabilitySigned:         true,
				ComputerUseWaiverSigned: true,
				MultimediaReleaseForm:   false,
			},
			{
				TeamID:                  uuid.New(),
				Email:                   "other.student@somewhere.edu",
				Name:                    "Other Student",
				ParentEmail:             "parent2@coolcompany.com",
				PreviouslyParticipated:  false,
				EmailConfirmed:          true,
				LiabilitySigned:         false,
				ComputerUseWaiverSigned: false,
				MultimediaReleaseForm:   false,
			},
		},
	}, database.Team{
		ID:           uuid.New(),
		TeacherEmail: "ohea@ohea.com",
		Name:         "Not In Person Team",
		Division:     "Advanced",
		InPerson:     false,
		Members: []database.Student{
			{
				TeamID:                  uuid.New(),
				Email:                   "test.student@somewhere.edu",
				Name:                    "Test Student",
				ParentEmail:             "parent@coolcompany.com",
				PreviouslyParticipated:  false,
				EmailConfirmed:          true,
				LiabilitySigned:         true,
				ComputerUseWaiverSigned: true,
				MultimediaReleaseForm:   false,
			},
			{
				TeamID:                  uuid.New(),
				Email:                   "other.student@somewhere.edu",
				Name:                    "Other Student",
				ParentEmail:             "parent2@coolcompany.com",
				PreviouslyParticipated:  false,
				EmailConfirmed:          true,
				LiabilitySigned:         false,
				ComputerUseWaiverSigned: false,
				MultimediaReleaseForm:   false,
			},
		},
	})

	return map[string]any{
		"Username":    user.Name,
		"SchoolName":  user.SchoolName,
		"SchoolCity":  user.SchoolCity,
		"SchoolState": user.SchoolState,
		"Teams":       teams,
	}
}
