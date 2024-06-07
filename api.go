package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/users", makeHTPHandleFunc(s.handleUsers))
	router.HandleFunc("/posts", makeHTPHandleFunc(s.handlePosts))
	router.HandleFunc("/posts/{id}", makeHTPHandleFunc(s.handleGetPost))

	http.ListenAndServe(s.listenAddr, router)
}

// USERS
func (s *APIServer) handleUsers(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetAllUsers(w, r)
	case "POST":
		return s.handleCreateUser(w, r)
	case "PUT":
		return s.handleUpdateUser(w, r)
	case "DELETE":
		return s.handleDeleteUser(w, r)
	}
	return methodNotAllowed(r.Method)
}

func (s *APIServer) handleGetAllUsers(w http.ResponseWriter, r *http.Request) error {
	users, err := s.store.GetUsers()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, users)
}

func (s *APIServer) handleCreateUser(w http.ResponseWriter, r *http.Request) error {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	createdUser, err := s.store.CreateUser(req)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusCreated, createdUser)
}

func (s *APIServer) handleUpdateUser(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APIServer) handleDeleteUser(w http.ResponseWriter, r *http.Request) error {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}
	return nil
}

// POSTS
func (s *APIServer) handlePosts(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetAllPosts(w, r)
	case "POST":
		return s.handleCreatePost(w, r)
	case "PUT":
		return s.handleUpdatePost(w, r)
	case "DELETE":
		return s.handleUpdatePost(w, r)
	}
	return methodNotAllowed(r.Method)
}

func (s *APIServer) handleGetPost(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	fmt.Println(id)
	return nil
}

func (s *APIServer) handleGetAllPosts(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APIServer) handleCreatePost(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APIServer) handleUpdatePost(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APIServer) handleDeletePost(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// Helper functions
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type APIError struct {
	error
	status int
}

func NewAPIError(err error, status int) APIError {
	return APIError{err, status}
}

func methodNotAllowed(method string) APIError {
	err := fmt.Errorf("method not allowed: %s", method)
	return NewAPIError(err, http.StatusBadRequest)
}

func (e APIError) Unwrap() error {
	return e.error
}

func (e APIError) HTTPStatus() int {
	return e.status
}

func makeHTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			status := http.StatusInternalServerError
			var apiErr APIError
			if errors.As(err, &apiErr) {
				status = apiErr.HTTPStatus()
			}
			http.Error(w, err.Error(), status)
		}
	}
}
