package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(port string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: ":" + port,
		store:      store,
	}
}

func (s *APIServer) Run() {
	http.HandleFunc("/login", makeHTPHandleFunc(s.handleLogin))
	http.HandleFunc("/signup", makeHTPHandleFunc(s.handleSignup))

	fmt.Printf("Listen and serve: http://localhost%s\n", s.listenAddr)
	http.ListenAndServe(s.listenAddr, nil)
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func WriteNoContent(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}

type APIError struct {
	error
	status int
}

func NewAPIError(err error, status int) APIError {
	return APIError{err, status}
}

func methodNotAllowedErr(method string) APIError {
	err := fmt.Errorf("method not allowed: %s", method)
	return NewAPIError(err, http.StatusMethodNotAllowed)
}

func (e APIError) Unwrap() error {
	return e.error
}

func (e APIError) HTTPStatus() int {
	return e.status
}

type apiFunc func(http.ResponseWriter, *http.Request) error

func makeHTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			status := http.StatusInternalServerError

			// check if error is API error
			var apiErr APIError
			if errors.As(err, &apiErr) {
				status = apiErr.HTTPStatus()
			}

			http.Error(w, err.Error(), status)
		}
	}
}

func withSessionAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("authenticating...")

		c, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				// If the cookie is not set, return an unauthorized status
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// For any other type of error, return a bad request status
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sessionID := c.Value

		// We then get the name of the user from our session map, where we set the session token
		userSession, err := s.GetSession(sessionID)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if userSession.isExpired() {
			//delete(sessions, sessionToken)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		handlerFunc(w, r)
	}
}
