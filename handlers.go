package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/micahco/racket-connections/internal/crypto"
	"github.com/micahco/racket-connections/internal/models"
	"github.com/micahco/racket-connections/internal/validator"
	"github.com/micahco/racket-connections/ui"
)

func refresh(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func unauthorizedError(w http.ResponseWriter) {
	http.Error(w, "permission denied", http.StatusUnauthorized)
}

func validationError(w http.ResponseWriter, v validator.Validator) {
	http.Error(w, v.Errors(), http.StatusUnprocessableEntity)
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	clientError(w, http.StatusNotFound)
}

func (app *application) handleStatic() http.Handler {
	if app.isProduction {
		return http.FileServer(http.FS(ui.Files))
	}

	fs := http.FileServer(http.Dir("./ui/static/"))
	return http.StripPrefix("/static", fs)
}

func (app *application) handleFavicon(w http.ResponseWriter, r *http.Request) {
	if app.isProduction {
		http.ServeFileFS(w, r, ui.Files, "static/favicon.ico")

		return
	}

	http.ServeFile(w, r, "./ui/static/favicon.ico")
}

func (app *application) handleRootGet(w http.ResponseWriter, r *http.Request) {
	if app.isAuthenticated(r) {
		http.Redirect(w, r, "/posts/latest", http.StatusSeeOther)

		return
	}

	app.render(w, r, http.StatusOK, "login.html", nil)
}

func (app *application) handleAuthLoginGet(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type authLoginForm struct {
	email    string
	password string
	validator.Validator
}

func (app *application) handleAuthLoginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	form := authLoginForm{
		email:    r.Form.Get("email"),
		password: r.Form.Get("password"),
	}

	form.Validate(validator.NotBlank(form.email), "invalid email: cannot be blank")
	form.Validate(validator.Matches(form.email, validator.EmailRX), "invalid email: must be a valid email address")
	form.Validate(validator.NotBlank(form.password), "invalid password: cannot be blank")

	if !form.IsValid() {
		validationError(w, form.Validator)

		return
	}

	id, err := app.users.Authenticate(form.email, form.password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			unauthorizedError(w)
		} else {
			app.serverError(w, err)
		}

		return
	}

	err = app.login(r, id)
	if err != nil {
		app.serverError(w, err)

		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) handleAuthLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.logout(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) handleAuthSignupGet(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type authSignupForm struct {
	email string
	validator.Validator
}

func (app *application) handleAuthSignupPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	form := authSignupForm{email: r.Form.Get("email")}

	form.Validate(validator.NotBlank(form.email), "invalid email: cannot be blank")
	form.Validate(validator.Matches(form.email, validator.EmailRX), "invalid email: must be a valid email address")
	form.Validate(validator.MaxChars(form.email, 254), "invalid email: must be no more than 254 characters long")
	form.Validate(validator.PermittedEmailDomain(form.email, "oregonstate.edu"), "invalid email: must be an OSU email address")

	if !form.IsValid() {
		validationError(w, form.Validator)

		return
	}

	// Check if link verification has already been created
	v, err := app.verifications.Get(form.email)
	if err != nil && err != models.ErrNoRecord {
		app.serverError(w, err)

		return
	}

	// Consistent flash message
	f := FlashMessage{
		Type:    FlashInfo,
		Message: "A link to activate your account has been sent to the email address provided.",
	}

	// Don't send a new link if less than 5 minutes since last
	if v != nil {
		min := 5 * time.Minute
		if time.Since(v.CreatedAt) < min {
			app.flash(r, f)

			refresh(w, r)

			return
		}
	}

	token, err := crypto.GenerateRandomString(32)
	if err != nil {
		app.serverError(w, err)

		return
	}

	url := fmt.Sprintf("%s/auth/create?token=%s", app.url, token)
	html := fmt.Sprintf("<p>Please follow the link below to create your account:</p>"+
		"<a href=\"%s\">%s</a>", url, url)

	err = app.sendMail(form.email, "Email verification", html)
	if err != nil {
		app.serverError(w, err)

		return
	}

	err = app.verifications.Insert(token, form.email)
	if err != nil {
		app.serverError(w, err)

		return
	}

	app.sessionManager.Clear(r.Context())
	app.sessionManager.RenewToken(r.Context())
	app.sessionManager.Put(r.Context(), verificationEmailSessionKey, form.email)

	app.flash(r, f)

	refresh(w, r)
}

type authCreateData struct {
	HasSessionEmail bool
}

func (app *application) handleAuthCreateGet(w http.ResponseWriter, r *http.Request) {
	queryToken := r.URL.Query().Get("token")
	if queryToken == "" {
		unauthorizedError(w)

		return
	}

	app.sessionManager.Put(r.Context(), verificationTokenSessionKey, queryToken)

	exists := app.sessionManager.Exists(r.Context(), verificationEmailSessionKey)
	data := authCreateData{
		HasSessionEmail: exists,
	}

	app.render(w, r, http.StatusOK, "auth-create.html", data)
}

type authCreateForm struct {
	name     string
	email    string
	password string
	validator.Validator
}

var ExpiredTokenFlash = FlashMessage{
	Type:    FlashError,
	Message: "Expired verification token.",
}

func (app *application) handleAuthCreatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	form := authCreateForm{
		name:     r.Form.Get("name"),
		email:    r.Form.Get("email"),
		password: r.Form.Get("password"),
	}

	email := app.sessionManager.GetString(r.Context(), verificationEmailSessionKey)
	if form.email != "" {
		email = form.email
	}

	form.Validate(validator.NotBlank(form.name), "invalid name: cannot be blank")
	form.Validate(validator.NotBlank(email), "invalid email: cannot be blank")
	form.Validate(validator.Matches(email, validator.EmailRX), "invalid email: must be a valid email address")
	form.Validate(validator.MaxChars(email, 254), "invalid email: must be no more than 254 characters long")
	form.Validate(validator.NotBlank(form.password), "invalid password: cannot be blank")
	form.Validate(validator.MinChars(form.password, 8), "invalid password: must be at least 8 characters long")
	form.Validate(validator.MaxChars(form.password, 72), "invalid password: must be no more than 72 characters long")

	if !form.IsValid() {
		validationError(w, form.Validator)

		return
	}

	token := app.sessionManager.GetString(r.Context(), verificationTokenSessionKey)
	if token == "" {
		unauthorizedError(w)

		return
	}

	err = app.verifications.Verify(token, email)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			unauthorizedError(w)
		} else if errors.Is(err, models.ErrExpiredVerification) {
			app.flash(r, ExpiredTokenFlash)

			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			app.serverError(w, err)
		}

		return
	}

	err = app.verifications.Purge(email)
	if err != nil {
		app.serverError(w, err)

		return
	}

	id, err := app.users.Insert(form.name, email, form.password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			unauthorizedError(w)
		} else {
			app.serverError(w, err)
		}

		return
	}

	app.sessionManager.Clear(r.Context())
	err = app.login(r, id)
	if err != nil {
		app.serverError(w, err)

		return
	}

	http.Redirect(w, r, "/profile/contact/new", http.StatusSeeOther)
}

func (app *application) handleAuthResetGet(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "auth-reset.html", nil)
}

type authResetForm struct {
	email string
	validator.Validator
}

func (app *application) handleAuthResetPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	form := authResetForm{email: r.Form.Get("email")}

	form.Validate(validator.NotBlank(form.email), "invalid email: cannot be blank")
	form.Validate(validator.MaxChars(form.email, 254), "invalid email: must be no more than 254 characters long")

	if !form.IsValid() {
		validationError(w, form.Validator)

		return
	}

	exists, err := app.users.ExistsEmail(form.email)
	if err != nil {
		app.serverError(w, err)
		return
	}

	f := FlashMessage{
		Type:    FlashInfo,
		Message: "A link to reset your password has been emailed to the address provided.",
	}

	if !exists {
		app.flash(r, f)

		refresh(w, r)

		return
	}

	token, err := crypto.GenerateRandomString(32)
	if err != nil {
		app.serverError(w, err)

		return
	}

	url := fmt.Sprintf("%s/auth/reset/update?token=%s", app.url, token)
	html := fmt.Sprintf("<p>Please follow the link below to reset your password:</p>"+
		"<a href=\"%s\">%s</a>", url, url)

	err = app.sendMail(form.email, "Reset password", html)
	if err != nil {
		app.serverError(w, err)

		return
	}

	err = app.verifications.Insert(token, form.email)
	if err != nil {
		app.serverError(w, err)

		return
	}

	app.sessionManager.Clear(r.Context())
	app.sessionManager.RenewToken(r.Context())
	app.sessionManager.Put(r.Context(), resetEmailSessionKey, form.email)

	app.flash(r, f)

	refresh(w, r)
}

func (app *application) handleAuthResetUpdateGet(w http.ResponseWriter, r *http.Request) {
	queryToken := r.URL.Query().Get("token")
	if queryToken == "" {
		unauthorizedError(w)

		return
	}

	app.sessionManager.Put(r.Context(), resetTokenSessionKey, queryToken)

	app.render(w, r, http.StatusOK, "auth-reset.html", nil)
}

type authResetUpdateForm struct {
	email    string
	password string
	validator.Validator
}

func (app *application) handleAuthResetUpdatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	form := authResetUpdateForm{
		email:    r.Form.Get("email"),
		password: r.Form.Get("password"),
	}

	email := app.sessionManager.GetString(r.Context(), resetEmailSessionKey)
	if form.email != "" {
		email = form.email
	}

	form.Validate(validator.NotBlank(email), "invalid email: cannot be blank")
	form.Validate(validator.Matches(email, validator.EmailRX), "invalid email: must be a valid email address")
	form.Validate(validator.MaxChars(email, 254), "invalid email: must be no more than 254 characters long")
	form.Validate(validator.NotBlank(form.password), "invalid password: cannot be blank")
	form.Validate(validator.MinChars(form.password, 8), "invalid password: must be at least 8 characters long")
	form.Validate(validator.MaxChars(form.password, 72), "invalid password: must be no more than 72 characters long")

	if !form.IsValid() {
		validationError(w, form.Validator)

		return
	}

	token := app.sessionManager.GetString(r.Context(), resetTokenSessionKey)
	if token == "" {
		unauthorizedError(w)

		return
	}

	err = app.verifications.Verify(token, email)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			unauthorizedError(w)
		} else if errors.Is(err, models.ErrExpiredVerification) {
			app.flash(r, ExpiredTokenFlash)

			http.Redirect(w, r, "/auth/reset", http.StatusSeeOther)
		} else {
			app.serverError(w, err)
		}

		return
	}

	err = app.verifications.Purge(email)
	if err != nil {
		app.serverError(w, err)

		return
	}

	err = app.users.UpdatePassword(email, form.password)
	if err != nil {
		app.serverError(w, err)

		return
	}

	app.sessionManager.Clear(r.Context())

	f := FlashMessage{
		Type:    FlashSuccess,
		Message: "Successfully updated password. Please login.",
	}
	app.flash(r, f)

	http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
}

func (app *application) handleProfileGet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "show user profile")
}

type newContactData struct {
	ContactMethods []*models.ContactMethod
	RefererPath    string
}

func (app *application) handleProfileContactNewGet(w http.ResponseWriter, r *http.Request) {
	refPath := ""

	ref, err := url.Parse(r.Referer())
	if err == nil {
		refPath = ref.Path
	}

	m, err := app.contacts.AllMethods()
	if err != nil {
		app.serverError(w, err)
	}

	data := newContactData{
		ContactMethods: m,
		RefererPath:    refPath,
	}

	app.render(w, r, http.StatusOK, "profile-contact-new.html", data)
}

type newContactForm struct {
	referer string
	value   string
	validator.Validator
}

func (app *application) handleProfileContactNewPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	methodID, err := strconv.Atoi(r.Form.Get("contact-method"))
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	form := newContactForm{
		referer: r.Form.Get("referer"),
		value:   r.Form.Get("contact-value"),
	}

	form.Validate(validator.NotBlank(form.value), "invalid value: cannot be blank")

	if methodID == 1 {
		form.Validate(validator.Matches(form.value, validator.EmailRX), "invalid email: must be a valid email address")
		form.Validate(validator.MaxChars(form.value, 254), "invalid email: must be no more than 254 characters long")
	} else {
		form.Validate(validator.MaxChars(form.value, 100), "invalid value: must be no more than 100 characters long")
	}

	if !form.IsValid() {
		validationError(w, form.Validator)

		return
	}

	suid, err := app.getSessionUserID(r)
	if err != nil {
		unauthorizedError(w)

		return
	}

	err = app.contacts.Insert(suid, methodID, form.value)
	if err != nil {
		app.serverError(w, err)

		return
	}

	redirectURL := "/profile"

	if form.referer == "/auth/create" {
		redirectURL = "/"
	} else if form.referer == "/posts/new" {
		redirectURL = form.referer
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (app *application) handleProfileEditGet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "show edit profile form")
}

func (app *application) handleProfileEditPost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "update user profile")
}

func (app *application) handleProfileDeletePost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "delete user profile")
}

func (app *application) handlePostsGet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "show all posts")
}

func (app *application) handlePostsPost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "create new post")
}

type postsLatestData struct {
	Map    map[int][]*models.PostDetails
	Sports []*models.Sport
}

func (app *application) handlePostsLatestGet(w http.ResponseWriter, r *http.Request) {
	m, err := app.posts.Latest()
	if err != nil {
		app.serverError(w, err)

		return
	}

	s, err := app.sports.All()
	if err != nil {
		app.serverError(w, err)

		return
	}

	data := postsLatestData{
		Map:    m,
		Sports: s,
	}

	app.render(w, r, http.StatusOK, "posts-latest.html", data)
}

type postData struct {
	Post          *models.PostDetails
	SessionUserID int
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

	userID, err := app.getSessionUserID(r)
	if err != nil {
		unauthorizedError(w)

		return
	}

	data := postData{
		Post:          p,
		SessionUserID: userID,
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
	Sports     []*models.Sport
	SportQuery string
	Skills     []*models.SkillLevel
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

	q := r.URL.Query().Get("sport")

	data := newPostData{
		Sports:     sports,
		SportQuery: strings.ToLower(q),
		Skills:     skills,
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
		url := fmt.Sprintf("/posts/%d", postID)
		http.Redirect(w, r, url, http.StatusConflict)

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

	postID, err = app.posts.Insert(form.comment, form.skillLevel, userID, form.sport)
	if err != nil {
		app.serverError(w, err)

		return
	}

	url := fmt.Sprintf("/posts/%d", postID)

	http.Redirect(w, r, url, http.StatusSeeOther)
}
