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
	Query     postsQuery
	Days      []*models.DayOfWeek
	Times     []*models.TimeOfDay
	Sports    []*models.Sport
	Posts     []*models.PostCard
	PageStart int
	PageEnd   int
	PageCount int
	NextPage  string
	PrevPage  string
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

	days, _ := app.models.Timeslot.Days()
	times, _ := app.models.Timeslot.Times()
	s, _ := app.models.Sport.All()

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
	c, p, err := app.models.Post.Fetch(q.Sport, q.Timeslot, limit, offset)
	if err != nil {
		app.serverError(w, r, err)

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

	count := int(c)
	start := offset + 1
	end := offset + PAGE_SIZE
	if end > count {
		end = count
	}

	data := postsData{
		Query:     q,
		Days:      days,
		Times:     times,
		Sports:    s,
		Posts:     p,
		NextPage:  nextPage,
		PrevPage:  prevPage,
		PageCount: count,
		PageStart: start,
		PageEnd:   end,
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
	IsOwner   bool
}

func (app *application) handlePostsIdGet(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		app.renderError(w, r, http.StatusBadRequest, "")

		return
	}

	postID, err := strconv.Atoi(idParam)
	if err != nil {
		app.renderError(w, r, http.StatusBadRequest, "")

		return
	}

	p, err := app.models.Post.GetDetails(postID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.renderError(w, r, http.StatusNotFound, "")
		} else {
			app.serverError(w, r, err)
		}

		return
	}

	c, err := app.models.Contact.All(p.UserID)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	timeslots, err := app.models.Timeslot.User(p.UserID)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	suid, err := app.getSessionUserID(r)
	if err != nil {
		unauthorizedError(w)

		return
	}

	d, _ := app.models.Timeslot.Days()
	t, _ := app.models.Timeslot.Times()

	data := postData{
		Post:      p,
		Contacts:  c,
		Days:      d,
		Times:     t,
		Timeslots: timeslots,
		IsOwner:   suid == p.UserID,
	}

	app.render(w, r, http.StatusOK, "post-details.html", data)
}

type postDeleteData struct {
	Post *models.PostDetails
}

func (app *application) handlePostsIdDeleteGet(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		app.renderError(w, r, http.StatusBadRequest, "")

		return
	}

	postID, err := strconv.Atoi(idParam)
	if err != nil {
		fmt.Println(2)
		app.renderError(w, r, http.StatusBadRequest, "")

		return
	}

	p, err := app.models.Post.GetDetails(postID)
	if err != nil {
		fmt.Println(3)
		app.renderError(w, r, http.StatusBadRequest, "")

		return
	}

	suid, err := app.getSessionUserID(r)
	if err != nil || suid != p.UserID {
		unauthorizedError(w)

		return
	}

	data := postDeleteData{
		Post: p,
	}

	app.render(w, r, http.StatusOK, "post-delete.html", data)
}

func (app *application) handlePostsIdDeletePost(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		app.renderError(w, r, http.StatusBadRequest, "")

		return
	}

	postID, err := strconv.Atoi(idParam)
	if err != nil {
		app.renderError(w, r, http.StatusBadRequest, "")

		return
	}

	userID, err := app.models.Post.GetUserID(postID)
	if err != nil {
		app.renderError(w, r, http.StatusBadRequest, "")

		return
	}

	suid, err := app.getSessionUserID(r)
	if err != nil || suid != userID {
		unauthorizedError(w)

		return
	}

	app.models.Post.Delete(postID)

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
	sports, err := app.models.Sport.All()
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	skills, err := app.models.Skill.All()
	if err != nil {
		app.serverError(w, r, err)

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
	suid, err := app.getSessionUserID(r)
	if err != nil {
		unauthorizedError(w)

		return
	}

	err = r.ParseForm()
	if err != nil {
		app.renderError(w, r, http.StatusBadRequest, "")

		return
	}

	sportID, err := strconv.Atoi(r.Form.Get("sport"))
	if err != nil {
		app.renderError(w, r, http.StatusBadRequest, "")

		return
	}

	// Check if user already has post with sport
	postID, err := app.models.Post.GetID(suid, sportID)
	if err != nil && !errors.Is(err, models.ErrNoRecord) {
		app.serverError(w, r, err)

		return
	}

	if postID != 0 {
		f := FlashMessage{
			Type:    FlashError,
			Message: "Unable to create post. You are only allowed to create one post per sport.",
		}
		app.flash(r, f)

		url := fmt.Sprintf("/posts/%d", postID)
		http.Redirect(w, r, url, http.StatusSeeOther)

		return
	}

	skill, err := strconv.Atoi(r.Form.Get("skill-level"))
	if err != nil {
		app.renderError(w, r, http.StatusBadRequest, "")

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

	postID, err = app.models.Post.Insert(suid, form.sport, form.skillLevel, form.comment)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	url := fmt.Sprintf("/posts/%d", postID)

	http.Redirect(w, r, url, http.StatusSeeOther)
}
