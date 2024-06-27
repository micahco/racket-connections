package main

import (
	"net/http"

	"github.com/micahco/racket-connections/internal/models"
)

type profileData struct {
	Name     string
	Email    string
	Contacts []*models.UserContact
	Times    []*models.Timeslot
}

func (app *application) handleProfileGet(w http.ResponseWriter, r *http.Request) {
	userID, err := app.getSessionUserID(r)
	if err != nil {
		app.serverError(w, err)

		return
	}

	p, err := app.users.GetProfile(userID)
	if err != nil {
		app.serverError(w, err)

		return
	}

	c, err := app.contacts.All(userID)
	if err != nil {
		app.serverError(w, err)

		return
	}

	t, err := app.timeslots.All(userID)
	if err != nil {
		app.serverError(w, err)

		return
	}

	data := profileData{
		Name:     p.Name,
		Email:    p.Email,
		Contacts: c,
		Times:    t,
	}

	app.render(w, r, http.StatusOK, "profile.html", data)
}
