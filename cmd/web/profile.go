package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/micahco/racket-connections/internal/models"
	"github.com/micahco/racket-connections/internal/validator"
)

type profileData struct {
	Name      string
	Email     string
	Contacts  []*models.UserContact
	Days      []*models.DayOfWeek
	Times     []*models.TimeOfDay
	Timeslots []*models.Timeslot
	Posts     []*models.ProfilePost
}

func (app *application) handleProfileGet(w http.ResponseWriter, r *http.Request) {
	suid, err := app.getSessionUserID(r)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	u, err := app.models.User.GetProfile(suid)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	contacts, err := app.models.Contact.UserContacts(suid)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	timeslots, err := app.models.Timeslot.User(suid)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	posts, err := app.models.Post.User(suid)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	days, _ := app.models.Timeslot.Days()
	times, _ := app.models.Timeslot.Times()

	data := profileData{
		Name:      u.Name,
		Email:     u.Email,
		Contacts:  contacts,
		Days:      days,
		Times:     times,
		Timeslots: timeslots,
		Posts:     posts,
	}

	app.render(w, r, http.StatusOK, "profile.html", data)
}

func (app *application) handleProfileDeleteGet(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "profile-delete.html", nil)
}

func (app *application) handleProfileDeletePost(w http.ResponseWriter, r *http.Request) {
	suid, err := app.getSessionUserID(r)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	err = app.models.User.Delete(suid)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	err = app.sessionManager.Clear(r.Context())
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	f := FlashMessage{
		Type:    FlashSuccess,
		Message: "Succesfully closed account",
	}
	app.flash(r, f)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type profileContactsData struct {
	Contacts []*models.UserContact
	Methods  []*models.ContactMethod
}

func (app *application) handleProfileContactsGet(w http.ResponseWriter, r *http.Request) {
	suid, err := app.getSessionUserID(r)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	c, err := app.models.Contact.UserContacts(suid)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	m, _ := app.models.Contact.Methods()

	data := profileContactsData{
		Contacts: c,
		Methods:  m,
	}

	app.render(w, r, http.StatusOK, "profile-contacts.html", data)
}

type newContactForm struct {
	contactMethod string
	contactValue  string
	validator.Validator
}

func (app *application) handleProfileContactsPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.renderError(w, r, http.StatusBadRequest, "")

		return
	}

	form := newContactForm{
		contactMethod: r.Form.Get("contact-method"),
		contactValue:  r.Form.Get("contact-value"),
	}

	switch form.contactMethod {
	case "email":
		form.Validate(validator.Matches(form.contactValue, validator.EmailRX), "invalid contact email: must be a valid email address")
		form.Validate(validator.MaxChars(form.contactValue, 254), "invalid contact email: must be no more than 254 characters long")
	case "phone":
		form.Validate(validator.Matches(form.contactValue, validator.PhoneRX), "invalid contact phone: must be a valid phone number")
	case "other":
		// TODO: cleanse for anything malicous
	default:
		app.renderError(w, r, http.StatusBadRequest, "")

		return
	}

	if !form.IsValid() {
		validationError(w, form.Validator)

		return
	}

	suid, err := app.getSessionUserID(r)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	userContacts, err := app.models.Contact.UserContacts(suid)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	for _, c := range userContacts {
		if c.Value == form.contactValue {
			f := FlashMessage{
				Type:    FlashError,
				Message: "Unable to add contact method: duplicate value",
			}
			app.flash(r, f)

			http.Redirect(w, r, "/profile/contacts", http.StatusSeeOther)

			return
		}
	}

	methodID, err := app.models.Contact.MethodID(form.contactMethod)
	if err != nil {
		app.serverError(w, r, err)
	}

	app.models.Contact.Insert(form.contactValue, suid, methodID)

	http.Redirect(w, r, "/profile/contacts", http.StatusSeeOther)
}

func (app *application) handleProfileContactsDeletePost(w http.ResponseWriter, r *http.Request) {
	suid, err := app.getSessionUserID(r)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	c, err := app.models.Contact.UserContacts(suid)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	if len(c) == 1 {
		f := FlashMessage{
			Type:    FlashError,
			Message: "Unable to delete contact method: minimum one required.",
		}
		app.flash(r, f)

		http.Redirect(w, r, "/profile/contacts", http.StatusSeeOther)

		return
	}

	err = r.ParseForm()
	if err != nil {
		app.renderError(w, r, http.StatusBadRequest, "")

		return
	}

	id, err := strconv.Atoi(r.Form.Get("id"))
	if err != nil {
		app.renderError(w, r, http.StatusBadRequest, "")

		return
	}

	err = app.models.Contact.Delete(id)
	if err != nil {
		app.renderError(w, r, http.StatusBadRequest, "")

		return
	}

	http.Redirect(w, r, "/profile/contacts", http.StatusSeeOther)
}

type profileAvailabilityData struct {
	Days      []*models.DayOfWeek
	Times     []*models.TimeOfDay
	Timeslots []*models.Timeslot
}

func (app *application) handleProfileAvailabilityGet(w http.ResponseWriter, r *http.Request) {
	suid, err := app.getSessionUserID(r)
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

	data := profileAvailabilityData{
		Days:      days,
		Times:     times,
		Timeslots: timeslots,
	}

	app.render(w, r, http.StatusOK, "profile-availability.html", data)
}

func (app *application) handleProfileAvailabilityPost(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.renderError(w, r, http.StatusBadRequest, "")

		return
	}

	suid, err := app.getSessionUserID(r)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	err = app.models.Timeslot.DeleteUser(suid)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	days, _ := app.models.Timeslot.Days()
	times, _ := app.models.Timeslot.Times()
	for _, d := range days {
		for _, t := range times {
			key := fmt.Sprintf("%s-%s", d.Abbrev, t.Abbrev)
			if r.Form.Get(key) == "on" {
				err = app.models.Timeslot.Insert(suid, d.ID, t.ID)
				if err != nil {
					app.serverError(w, r, err)
				}
			}
		}
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
