package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// GCalendar is a calendar provider for a Google Calendar.
type GCalendar struct {
	CalConfig CalConfig // Selected Calendar Configuration
}

// SetConfig sets the configuration for this calendar provider
func (g *GCalendar) SetConfig(c CalConfig) {
	g.CalConfig = c
}

// GetEvents returns the calendar events
func (g *GCalendar) GetEvents(noDays int) (CalEvents, error) {
	if noDays <= 0 {
		noDays = 4
	}
	evts := CalEvents{
		Created: time.Now(),
		NoDays:  noDays,
	}
	lastFName := fmt.Sprintf("lastevents_%s.json", g.CalConfig.ID)

	client, err := g.getClient()
	if err != nil {
		evts.ReadFromFile(lastFName)
		return evts, fmt.Errorf("Error getting client. %s", err.Error())
	}

	srv, err := calendar.New(client)
	if err != nil {
		evts.ReadFromFile(lastFName)
		return evts, fmt.Errorf("Error creating calendar. %s", err.Error())
	}

	timeMin := time.Now().Format(time.RFC3339)
	timeMax := time.Now().Add(time.Duration(noDays*24) * time.Hour).Format(time.RFC3339)
	events, err := srv.Events.List("primary").
		ShowDeleted(false).SingleEvents(true).
		TimeMin(timeMin).TimeMax(timeMax).
		OrderBy("startTime").Do()
	if err != nil {
		evts.ReadFromFile(lastFName)
		return evts, fmt.Errorf("Error retrieving calendar events. %s", err.Error())
	}

	// b, err := json.Marshal(events)
	// ioutil.WriteFile("events.json", b, 0666)

	for _, item := range events.Items {
		if st, err := g.getTime(item.Start.DateTime, item.Start.Date); err == nil {
			if et, err := g.getTime(item.End.DateTime, item.End.Date); err == nil {
				evts.Events = append(evts.Events, CalEvent{
					ID:          g.CalConfig.ID,
					Name:        g.CalConfig.Name,
					Start:       st,
					End:         et,
					DayName:     st.Weekday().String(),
					Time:        st.Format("15:04"),
					Duration:    GetDurationString(st, et),
					Summary:     item.Summary,
					Description: item.Description,
					Location:    item.Location,
				})
			}
		}
	}

	// Save a copy of these events
	evts.WriteToFile(lastFName)

	return evts, nil
}

func (g *GCalendar) getTime(a string, b string) (time.Time, error) {
	l := "2006-01-02T15:04:05-07:00"
	if a == "" {
		return time.Parse(l, b)
	}
	return time.Parse(l, a)
}

func (g *GCalendar) getClient() (*http.Client, error) {
	// Read the credentials
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		return nil, errors.New("Error reading credentials.json file. " + err.Error())
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, errors.New("Error parsing credentials.json file. " + err.Error())
	}

	tokenFile := fmt.Sprintf("Token_%s.json", g.CalConfig.ID)
	token, err := g.getTokenFromFile(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("Error reading token file for %s. %s", g.CalConfig.Name, err.Error())
	}

	return config.Client(context.Background(), token), nil
}

// Retrieves a token from a local file.
func (g *GCalendar) getTokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}
