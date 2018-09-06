package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/brumawen/ical"
	uuid "github.com/satori/go.uuid"
)

// ICalFeed is a calendar provider for a iCal web feed.
type ICalFeed struct {
	CalConfig CalConfig // Selected Calendar Configuration
}

// SetConfig sets the configuration for this calendar provider
func (p *ICalFeed) SetConfig(c CalConfig) {
	p.CalConfig = c
}

// RemovedConfig is used to clean up after config has been removed
func (p *ICalFeed) RemovedConfig(c CalConfig) error {
	return nil
}

// ProviderName returns the name of the provider
func (p *ICalFeed) ProviderName() string {
	return "iCal"
}

// GetEvents returns the calendar events
func (p *ICalFeed) GetEvents(noDays int) (CalEvents, error) {
	evts := CalEvents{
		Created: time.Now(),
		NoDays:  noDays,
	}
	lastFName := fmt.Sprintf("lastevents_%s.json", p.CalConfig.ID)

	// Split the URL by lines
	urls := strings.Split(strings.Replace(p.CalConfig.URL, "\r", "", -1), "\n")

	for _, u := range urls {
		resp, err := http.Get(strings.TrimSpace(u))
		if resp != nil {
			defer resp.Body.Close()
			resp.Close = true
		}
		if err != nil {
			evts.ReadFromFile(lastFName)
			return evts, fmt.Errorf("Error getting feed. %s", err.Error())
		}
		c, err := ical.Parse(resp.Body, time.Local)
		if err != nil {
			evts.ReadFromFile(lastFName)
			return evts, fmt.Errorf("Error parsing feed. %s", err.Error())
		}
		ts := time.Now()
		te := time.Now().Add(time.Duration(noDays*24) * time.Hour)
		if len(c.Events) != 0 {
			for _, e := range c.Events {
				if e.StartDate.Equal(ts) || (e.StartDate.After(ts) && e.StartDate.Before(te)) {
					// Check if we have already loaded the event
					exists := false
					for _, x := range evts.Events {
						if x.UID == e.UID {
							exists = true
							break
						}
					}
					if !exists {
						// New event
						evts.Events = append(evts.Events, CalEvent{
							ID:          p.CalConfig.ID,
							Name:        p.CalConfig.Name,
							UID:         e.UID,
							Start:       e.StartDate,
							End:         e.EndDate,
							DayName:     e.StartDate.Weekday().String(),
							Time:        e.StartDate.Format("15:04"),
							Duration:    GetDurationString(e.StartDate, e.EndDate),
							Summary:     e.Summary,
							Description: e.Description,
							Colour:      p.CalConfig.Colour,
						})
					}
				}
			}
		}
	}
	evts.EventCount = len(evts.Events)

	// Save a copy of these events
	evts.WriteToFile(lastFName)
	return evts, nil
}

// ValidateConfig validates the configuration change for the calendar
// and returns the calendar configuation ready to save
func (p *ICalFeed) ValidateConfig(c CalConfig) (CalConfig, error) {
	if c.ID == "" {
		return c, errors.New("ID must be specified")
	}
	if c.Name == "" {
		return c, errors.New("Name must be specified")
	}
	if c.Colour == "" {
		return c, errors.New("Colour must be specified")
	}
	if c.URL == "" {
		return c, errors.New("URL must be specified")
	}
	return c, nil
}

// ValidateNewConfig validates the new configuration values
// and returns the calendar configuration ready to save
func (p *ICalFeed) ValidateNewConfig(c NewCalConfig) (CalConfig, error) {
	cc := CalConfig{}

	if c.Name == "" {
		return cc, errors.New("Name must be specified")
	}
	if c.URL == "" {
		return cc, errors.New("URL must be specified")
	}
	if c.Colour == "" {
		return cc, errors.New("Colour must be selected")
	}

	// Generate a new ID
	uuid, err := uuid.NewV4()
	if err != nil {
		return cc, errors.New("Error creating GUID. " + err.Error())
	}
	cc.ID = uuid.String()
	cc.Name = c.Name
	cc.Provider = p.ProviderName()
	cc.Colour = c.Colour
	cc.URL = c.URL

	return cc, nil
}
