package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strings"
	"time"
)

// CalEvents holds a collection of calendar events as returned by a provider
type CalEvents struct {
	Created time.Time  `json:"created"` // Date calendar events were retrieved
	NoDays  int        `json:"noDays"`  // Number of days retrieved
	Events  []CalEvent `json:"events"`  // List of calendar events
}

// CalEvent holds the calendar event details
type CalEvent struct {
	ID          string    `json:"id"`          // Identifier of the calendar
	Name        string    `json:"name"`        // Name of the calendar
	Start       time.Time `json:"start"`       // Start time
	End         time.Time `json:"end"`         // End time
	DayName     string    `json:"dayName"`     // Name of the day
	Time        string    `json:"time"`        // Starting time of the event
	Duration    string    `json:"duration"`    // Duration of the event
	Summary     string    `json:"summary"`     // Summary of the event
	Location    string    `json:"location"`    // Location of the event
	Description string    `json:"description"` // Description of the event
	Colour      string    `json:"colour"`      // Colour of the event
}

// ReadFromFile will read the calendar events from the specified file
func (c *CalEvents) ReadFromFile(path string) error {
	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		b, err := ioutil.ReadFile(path)
		if err == nil {
			err = json.Unmarshal(b, &c)
		}
	}
	return err
}

// WriteToFile will write the calendar settings to the specified file
func (c *CalEvents) WriteToFile(path string) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0666)
}

// GetDurationString returns the duration as a printable string
func GetDurationString(start time.Time, end time.Time) string {
	d := end.Sub(start)
	h := d.Hours()
	m := math.Mod(d.Minutes(), 60)

	r := ""
	if h > 0 {
		r = r + fmt.Sprintf("%2.fh", h)
	}
	if m != 0 {
		if r != "" {
			r = r + " "
		}
		r = r + fmt.Sprintf("%2.fm", m)
	}
	return strings.Trim(r, " ")
}
