package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/gorilla/mux"
)

// CalendarController handles the Web Methods for reading calendars.
type CalendarController struct {
	Srv *Server
}

// AddController adds the controller routes to the router
func (c *CalendarController) AddController(router *mux.Router, s *Server) {
	c.Srv = s
	router.Methods("GET").Path("/calendar/get").Name("GetNames").
		Handler(Logger(c, http.HandlerFunc(c.handleGetNames)))
	router.Methods("GET").Path("/calendar/get/{noDays}").Name("GetCalendars").
		Handler(Logger(c, http.HandlerFunc(c.handleGetCalendars)))
}

func (c *CalendarController) handleGetNames(w http.ResponseWriter, r *http.Request) {
	l := []CalName{}

	for _, i := range c.Srv.Config.Calendars {
		l = append(l, CalName{
			Name:   i.Name,
			Colour: i.Colour,
		})
	}

	if b, err := json.Marshal(l); err != nil {
		m := fmt.Sprintf("Error serializing calendar names. %s", err.Error())
		c.LogError(m)
		http.Error(w, m, 500)
	} else {
		w.Header().Set("content-type", "application/json")
		w.Write(b)
	}
}

func (c *CalendarController) handleGetCalendars(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	d := vars["noDays"]
	noDays := 4
	if d != "" {
		if i, err := strconv.Atoi(d); err == nil {
			noDays = i
		}
	}
	el := []CalEvent{}
	for _, calConfig := range c.Srv.Config.Calendars {
		p, err := c.getCalendarProvider(calConfig)
		if err != nil {
			c.LogError(fmt.Sprintf("Error retrieving calendar provider for %s. %s", calConfig.Name, err.Error()))
		} else {
			evts, err := p.GetEvents(noDays)
			if err != nil {
				c.LogError(fmt.Sprintf("Error retrieving calendar events for %s. %s", calConfig.Name, err.Error()))
			} else {
				for _, e := range evts.Events {
					el = append(el, e)
				}
			}
		}
	}
	sort.Slice(el, func(i, j int) bool {
		return el[j].Start.After(el[i].Start)
	})

	if b, err := json.Marshal(el); err != nil {
		m := fmt.Sprintf("Error serializing calendar events. %s", err.Error())
		c.LogError(m)
		http.Error(w, m, 500)
	} else {
		w.Header().Set("content-type", "application/json")
		w.Write(b)
	}
}

func (c *CalendarController) getCalendarProvider(cc CalConfig) (CalendarProvider, error) {
	switch cc.Provider {
	case "Google":
		// Google Calendar
		gc := new(GCalendar)
		gc.SetConfig(cc)
		return gc, nil
	case "iCal":
		ic := new(ICalFeed)
		ic.SetConfig(cc)
		return ic, nil
	default:
		m := fmt.Sprintf("Invalid Calendar provider '%s'", cc.Provider)
		c.LogError(m)
		return nil, errors.New(m)
	}
}

// LogInfo is used to log information messages for this controller.
func (c *CalendarController) LogInfo(v ...interface{}) {
	a := fmt.Sprint(v)
	logger.Info("CalendarController: [Inf] ", a[1:len(a)-1])
}

// LogError is used to log information messages for this controller.
func (c *CalendarController) LogError(v ...interface{}) {
	a := fmt.Sprint(v)
	logger.Info("CalendarController: [Err] ", a[1:len(a)-1])
}
