package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	MIN_PASSWORD_LEN = 8
	MAX_PASSWORD_LEN = 72
)

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		t, _ := template.ParseFiles("login.html")
		t.Execute(w, nil)
		return nil
	case "POST":
		return s.handleLoginRequest(w, r)
	}
	return nil
}

func (s *APIServer) handleLoginRequest(w http.ResponseWriter, r *http.Request) error {
	// first, check if session id already present

	var req LoginRequest
	r.ParseForm()
	req.Email = r.Form.Get("email")
	req.Password = r.Form.Get("password")

	user, err := s.store.GetUserByEmail(req.Email)
	if err != nil {
		return permissionDenied()
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return permissionDenied()
	}

	s.createSession(w, user)

	// redirect...
	return WriteJSON(w, http.StatusOK, req)
}

func (s *APIServer) handleSignup(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		// redirect to login
	case "POST":
		return s.handleSignupRequest(w, r)
	}
	return nil
}

func (s *APIServer) handleSignupRequest(w http.ResponseWriter, r *http.Request) error {
	var req SignupRequest
	r.ParseForm()
	req.FirstName = r.Form.Get("firstName")
	req.LastName = r.Form.Get("lastName")
	req.Email = r.Form.Get("email")
	req.Password = r.Form.Get("password")

	if err := validatePassword(req.Password); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user, err := s.store.CreateUser(req, hash)
	if err != nil {
		return err
	}

	s.createSession(w, user)

	// redirect...
	return WriteJSON(w, http.StatusCreated, user)
}

func (s *APIServer) createSession(w http.ResponseWriter, user *User) {
	sesionID := uuid.NewString()
	expiry := time.Now().Add(time.Hour)

	s.store.CreateSession(sesionID, user.ID, expiry)

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sesionID,
		Expires: expiry,
	})
	fmt.Println("Session created:", sesionID)
}

// purposely ambiguous authentication error
func permissionDenied() APIError {
	err := errors.New("permission denied")
	return NewAPIError(err, http.StatusUnauthorized)
}

func validatePassword(p string) error {
	if len(p) < MIN_PASSWORD_LEN || len(p) > MIN_PASSWORD_LEN {
		err := fmt.Errorf("invalid password length: %d", len(p))
		return NewAPIError(err, http.StatusBadRequest)
	}

	return nil
}
