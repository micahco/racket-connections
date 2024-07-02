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
	suid, err := app.getSessionUserID(r)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	p, err := app.models.User.GetProfile(suid)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	c, err := app.models.Contact.All(suid)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	timeslots, err := app.models.Timeslot.User(suid)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	days, _ := app.models.Timeslot.Days()
	times, _ := app.models.Timeslot.Times()

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
