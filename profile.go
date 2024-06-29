package main

import (
	"net/http"

	"github.com/micahco/racket-connections/internal/models"
)

type profileData struct {
	Name      string
	Email     string
	Contacts  []*models.UserContact
	Days      []*models.DayOfWeek
	Times     []*models.TimeOfDay
	Timeslots []*models.Timeslot
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

	timeslots, err := app.timeslots.User(userID)
	if err != nil {
		app.serverError(w, err)

		return
	}

	days, _ := app.timeslots.Days()
	times, _ := app.timeslots.Times()

	data := profileData{
		Name:      p.Name,
		Email:     p.Email,
		Contacts:  c,
		Days:      days,
		Times:     times,
		Timeslots: timeslots,
	}

	app.render(w, r, http.StatusOK, "profile.html", data)
}
