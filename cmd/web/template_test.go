package main

import (
	"testing"
	"time"

	"github.com/micahco/racket-connections/internal/assert"
)

func TestHumanDate(t *testing.T) {
	tests := []struct {
		name string
		tm   time.Time
		want string
	}{
		{
			name: "UTC",
			tm:   time.Date(2022, 3, 17, 10, 15, 0, 0, time.UTC),
			want: "17-Mar-2022",
		},
		{
			name: "Empty",
			tm:   time.Time{},
			want: "",
		},
		{
			name: "CET",
			tm:   time.Date(2022, 3, 17, 10, 15, 0, 0, time.FixedZone("CET", 1*60*60)),
			want: "17-Mar-2022",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := humanDate(tt.tm)
			assert.Equal(t, actual, tt.want)
		})
	}
}

func TestSinceDate(t *testing.T) {
	tests := []struct {
		name string
		tm   time.Time
		want string
	}{
		{
			name: "Today",
			tm:   time.Now(),
			want: "today",
		},
		{
			name: "Day",
			tm:   time.Now().Add(-24 * time.Hour),
			want: "1 day ago",
		},
		{
			name: "Days",
			tm:   time.Now().Add(-3 * 24 * time.Hour),
			want: "3 days ago",
		},
		{
			name: "Week",
			tm:   time.Now().Add(-7 * 24 * time.Hour),
			want: "1 week ago",
		},
		{
			name: "Week and some days",
			tm:   time.Now().Add(-10 * 24 * time.Hour),
			want: "1 week ago",
		},
		{
			name: "Weeks",
			tm:   time.Now().Add(-15 * 24 * time.Hour),
			want: "2 weeks ago",
		},
		// If it's more than a month, we just want humanDate
		{
			name: "More than a month ago",
			tm:   time.Date(2022, 3, 17, 10, 15, 0, 0, time.UTC),
			want: "17-Mar-2022",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := humanDate(tt.tm)
			assert.Equal(t, actual, tt.want)
		})
	}
}
