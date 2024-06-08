package main

import "time"

type Session struct {
	sessionID string
	userID    int
	expiry    time.Time
}

func (s Session) isExpired() bool {
	return s.expiry.Before(time.Now())
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignupRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	CreatedAt    time.Time `json:"createdAt"`
	PasswordHash string    `json:"-"`
}
