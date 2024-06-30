package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/micahco/racket-connections/internal/models"
	"github.com/micahco/racket-connections/internal/validator"
)

const (
	PAGE_SIZE = 12
)

type postsQuery struct {
	Sport    []string
	Timeslot []models.Timeslot
}

type postsData struct {
	Query    postsQuery
	Days     []*models.DayOfWeek
	Times    []*models.TimeOfDay
	Sports   []*models.Sport
	Posts    []*models.PostCard
	NextPage string
	PrevPage string
}

func (app *application) handlePostsGet(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	sportsQuery := r.URL.Query()["sport"]
	for i := 0; i < len(sportsQuery); i++ {
		sportsQuery[i] = strings.ToLower(sportsQuery[i])
	}

	q := postsQuery{
		Sport:    sportsQuery,
		Timeslot: make([]models.Timeslot, 0),
	}

	days, _ := app.timeslots.Days()
	times, _ := app.timeslots.Times()
	s, _ := app.sports.All()

	for _, d := range days {
		for _, t := range times {
			key := fmt.Sprintf("%s-%s", d.Abbrev, t.Abbrev)
			if r.URL.Query().Get(key) == "on" {
				q.Timeslot = append(q.Timeslot, models.Timeslot{
					Day:  d,
					Time: t,
				})
			}
		}
	}

	limit := PAGE_SIZE + 1
	offset := (page - 1) * PAGE_SIZE
	p, err := app.posts.Fetch(q.Sport, q.Timeslot, limit, offset)
	if err != nil {
		app.serverError(w, err)

		return
	}

	var nextPage string
	if len(p) > PAGE_SIZE {
		p = p[:PAGE_SIZE]
		q := r.URL.Query()
		q.Set("page", strconv.Itoa(page+1))
		nextPage = fmt.Sprintf("/posts?%s", q.Encode())
	}

	var prevPage string
	if page > 1 {
		q := r.URL.Query()
		q.Set("page", strconv.Itoa(page-1))
		prevPage = fmt.Sprintf("/posts?%s", q.Encode())
	}

	data := postsData{
		Query:    q,
		Days:     days,
		Times:    times,
		Sports:   s,
		Posts:    p,
		NextPage: nextPage,
		PrevPage: prevPage,
	}

	app.render(w, r, http.StatusOK, "posts.html", data)
}

func (app *application) handlePostsPost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "create new post")
}

type postData struct {
	Post      *models.PostDetails
	Contacts  []*models.UserContact
	Days      []*models.DayOfWeek
	Times     []*models.TimeOfDay
	Timeslots []*models.Timeslot
}

func (app *application) handlePostsIdGet(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		clientError(w, http.StatusBadRequest)

		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	p, err := app.posts.GetDetails(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			clientError(w, http.StatusNotFound)
		} else {
			app.serverError(w, err)
		}

		return
	}

	c, err := app.contacts.All(p.UserID)
	if err != nil {
		app.serverError(w, err)

		return
	}

	timeslots, err := app.timeslots.User(p.UserID)
	if err != nil {
		app.serverError(w, err)

		return
	}

	d, _ := app.timeslots.Days()
	t, _ := app.timeslots.Times()

	data := postData{
		Post:      p,
		Contacts:  c,
		Days:      d,
		Times:     t,
		Timeslots: timeslots,
	}

	app.render(w, r, http.StatusOK, "post-details.html", data)
}

func (app *application) handlePostsIdEditGet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "edit post")
}

func (app *application) handlePostsIdEditPost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "edit post")
}

func (app *application) handlePostsIdDeleteGet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "delete post")
}

func (app *application) handlePostsIdDeletePost(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		clientError(w, http.StatusBadRequest)

		return
	}

	postID, err := strconv.Atoi(idParam)
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	suid, err := app.getSessionUserID(r)
	if err != nil {
		unauthorizedError(w)

		return
	}

	userID, err := app.posts.GetUserID(postID)
	if err != nil || suid != userID {
		unauthorizedError(w)

		return
	}

	app.posts.Delete(postID)

	f := FlashMessage{
		Type:    FlashSuccess,
		Message: "Successfully deleted post",
	}
	app.flash(r, f)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type newPostData struct {
	Sports []*models.Sport
	Skills []*models.SkillLevel
}

func (app *application) handlePostsNewGet(w http.ResponseWriter, r *http.Request) {
	sports, err := app.sports.All()
	if err != nil {
		app.serverError(w, err)

		return
	}

	skills, err := app.skills.All()
	if err != nil {
		app.serverError(w, err)

		return
	}

	data := newPostData{
		Sports: sports,
		Skills: skills,
	}

	app.render(w, r, http.StatusOK, "posts-new.html", data)
}

type newPostForm struct {
	sport      int
	skillLevel int
	comment    string
	validator.Validator
}

func (app *application) handlePostsNewPost(w http.ResponseWriter, r *http.Request) {
	userID, err := app.getSessionUserID(r)
	if err != nil {
		unauthorizedError(w)

		return
	}

	err = r.ParseForm()
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	sportID, err := strconv.Atoi(r.Form.Get("sport"))
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	// Check if user already has post with sport
	postID, err := app.posts.GetID(userID, sportID)
	if err != nil && !errors.Is(err, models.ErrNoRecord) {
		app.serverError(w, err)

		return
	}

	if postID != 0 {
		f := FlashMessage{
			Type:    FlashInfo,
			Message: "Already have post for sport.",
		}
		app.flash(r, f)

		url := fmt.Sprintf("/posts/%d", postID)
		http.Redirect(w, r, url, http.StatusSeeOther)

		return
	}

	skill, err := strconv.Atoi(r.Form.Get("skill-level"))
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	form := newPostForm{
		sport:      sportID,
		skillLevel: skill,
		comment:    r.Form.Get("comment"),
	}

	form.Validate(validator.MaxChars(form.comment, 254), "invalid comment: must be no more than 254 characters long")
	form.Validate(validator.PermittedInt(form.sport, 1, 2, 3, 4, 5, 6), "invalid sport")
	form.Validate(validator.PermittedInt(form.skillLevel, 1, 2, 3, 4, 5), "invalid skill level")

	if !form.IsValid() {
		validationError(w, form.Validator)

		return
	}

	postID, err = app.posts.Insert(userID, form.sport, form.skillLevel, form.comment)
	if err != nil {
		app.serverError(w, err)

		return
	}

	url := fmt.Sprintf("/posts/%d", postID)

	http.Redirect(w, r, url, http.StatusSeeOther)
}
