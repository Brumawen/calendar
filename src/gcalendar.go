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

	uuid "github.com/satori/go.uuid"
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

// RemovedConfig is used to clean up after config has been removed
func (g *GCalendar) RemovedConfig(c CalConfig) error {
	tokenFile := fmt.Sprintf("Token_%s.json", c.ID)
	if _, err := os.Stat(tokenFile); err == nil {
		return os.Remove(tokenFile)
	}
	return nil
}

// ProviderName returns the name of the provider
func (g *GCalendar) ProviderName() string {
	return "Google"
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
					Colour:      g.CalConfig.Colour,
				})
			}
		}
	}

	// Save a copy of these events
	evts.WriteToFile(lastFName)

	return evts, nil
}

// GetAuthenticateURL returns the URL that will be used to choose the calendar and
// authenticate with Google.
func (g *GCalendar) GetAuthenticateURL() (string, error) {
	config, err := g.getConfig()
	if err != nil {
		return "", err
	}

	url := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	return url, nil
}

// ValidateConfig validates the configuration change for the calendar
// and returns the calendar configuation ready to save
func (g *GCalendar) ValidateConfig(c CalConfig) (CalConfig, error) {
	if c.ID == "" {
		return c, errors.New("ID must be specified")
	}
	if c.Name == "" {
		return c, errors.New("Name must be specified")
	}
	if c.Colour == "" {
		return c, errors.New("Colour must be specified")
	}
	return c, nil
}

// ValidateNewConfig validates the new configuration values
// and returns the calendar configuration ready to save
func (g *GCalendar) ValidateNewConfig(c NewCalConfig) (CalConfig, error) {
	cc := CalConfig{}

	// Validate the token
	config, err := g.getConfig()
	if err != nil {
		return cc, err
	}

	if c.Name == "" {
		return cc, errors.New("Name must be specified")
	}
	if c.AuthCode == "" {
		return cc, errors.New("Authentication code must be specified")
	}
	if c.Colour == "" {
		return cc, errors.New("Colour must be selected")
	}

	// Get the token
	token, err := config.Exchange(oauth2.NoContext, c.AuthCode)
	if err != nil {
		m := fmt.Sprintf("Unable to retrieve authentication token. %v", err)
		return cc, errors.New(m)
	}

	// Generate a new ID
	uuid, err := uuid.NewV4()
	if err != nil {
		return cc, errors.New("Error creating GUID. " + err.Error())
	}
	cc.ID = uuid.String()
	cc.Name = c.Name
	cc.Provider = g.ProviderName()
	cc.Colour = c.Colour

	// Save the token
	err = g.saveTokenToFile(cc.ID, token)
	if err != nil {
		return cc, err
	}
	return cc, nil
}

func (g *GCalendar) getTime(a string, b string) (time.Time, error) {
	l := "2006-01-02T15:04:05-07:00"
	if a == "" {
		return time.Parse(l, b)
	}
	return time.Parse(l, a)
}

func (g *GCalendar) getConfig() (*oauth2.Config, error) {
	// Read the credentials
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		return nil, errors.New("Error reading credentials.json file. " + err.Error())
	}

	return google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
}

func (g *GCalendar) getClient() (*http.Client, error) {
	config, err := g.getConfig()
	if err != nil {
		return nil, err
	}

	token, err := g.getTokenFromFile(g.CalConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("Error reading token file for %s. %s", g.CalConfig.Name, err.Error())
	}

	return config.Client(context.Background(), token), nil
}

// Retrieves a token from a local file.
func (g *GCalendar) getTokenFromFile(id string) (*oauth2.Token, error) {
	tokenFile := fmt.Sprintf("Token_%s.json", id)
	f, err := os.Open(tokenFile)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

func (g *GCalendar) saveTokenToFile(id string, token *oauth2.Token) error {
	tokenFile := fmt.Sprintf("Token_%s.json", id)
	f, err := os.OpenFile(tokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("Unable to cache oauth token: %v", err)
	}
	json.NewEncoder(f).Encode(token)
	return nil
}
