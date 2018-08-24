package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

// ConfigController handles the Web Methods for configuring the module.
type ConfigController struct {
	Srv *Server
}

// ConfigPageData holds the data used to write to the configuration page.
type ConfigPageData struct {
	GoogleURL string
	Calendars []CalConfig
}

// AddController adds the controller routes to the router
func (c *ConfigController) AddController(router *mux.Router, s *Server) {
	c.Srv = s
	router.Path("/config.html").Handler(http.HandlerFunc(c.handleConfigWebPage))
	router.Methods("GET").Path("/config/get").Name("GetConfig").
		Handler(Logger(c, http.HandlerFunc(c.handleGetConfig)))
	router.Methods("GET").Path("/config/get/{id}").Name("GetCalendar").
		Handler(Logger(c, http.HandlerFunc(c.handleGetCalendar)))
	router.Methods("POST").Path("/config/add").Name("AddCalendar").
		Handler(Logger(c, http.HandlerFunc(c.handleAddCalendar)))
	router.Methods("POST").Path("/config/update").Name("UpdateCalendar").
		Handler(Logger(c, http.HandlerFunc(c.handleUpdateCalendar)))
	router.Methods("POST").Path("/config/remove/{id}").Name("RemoveCalendar").
		Handler(Logger(c, http.HandlerFunc(c.handleRemoveCalendar)))
}

func (c *ConfigController) handleConfigWebPage(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("./html/config.html"))

	gc := new(GCalendar)
	gurl, err := gc.GetAuthenticateURL()
	if err != nil {
		http.Error(w, "Error getting Google Authentication URL.", 500)
		return
	}

	v := ConfigPageData{
		Calendars: c.Srv.Config.Calendars,
		GoogleURL: gurl,
	}

	if err := t.Execute(w, v); err != nil {
		http.Error(w, "Error getting web page. "+err.Error(), 500)
	}
}

func (c *ConfigController) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	if err := c.Srv.Config.WriteTo(w); err != nil {
		http.Error(w, "Error serializing configuration. "+err.Error(), 500)
	}
}

func (c *ConfigController) handleGetCalendar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Calendar identifier not specified", 500)
		return
	}

	for _, i := range c.Srv.Config.Calendars {
		if i.ID == id {
			if err := i.WriteTo(w); err != nil {
				http.Error(w, "Error serializing calendar configuration. "+err.Error(), 500)
			}
			return
		}
	}
	http.Error(w, "Invalid calendar identifier", 500)
}

func (c *ConfigController) handleAddCalendar(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	nc := NewCalConfig{
		Name:     r.Form.Get("addName"),
		Colour:   r.Form.Get("addColour"),
		Provider: r.Form.Get("addProvider"),
	}

	// Check name or colour does not already exist
	for _, i := range c.Srv.Config.Calendars {
		if i.Name == nc.Name {
			http.Error(w, "This name has already been used.  Please select another name.", 500)
			return
		}
		if i.Colour == nc.Colour {
			http.Error(w, "This colour has already been used.  Please select another colour.", 500)
			return
		}
	}

	// Get the calendar provider
	p, err := c.getCalendarProvider(nc.Provider)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	switch nc.Provider {
	case "Google":
		nc.AuthCode = r.Form.Get("addGoogleCode")
	}
	// Validate the config
	cc, err := p.ValidateNewConfig(nc)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// Append the new configuration
	c.Srv.Config.Calendars = append(c.Srv.Config.Calendars, cc)
	err = c.Srv.Config.WriteToFile("config.json")
	if err != nil {
		m := fmt.Sprintf("Error saving config.json file. %s", err.Error())
		c.LogError(m)
		http.Error(w, m, 500)
	}
}

func (c *ConfigController) handleUpdateCalendar(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	id := r.Form.Get("updID")
	if id == "" {
		http.Error(w, "Calendar ID not specified", 500)
		return
	}

	cals := []CalConfig{}
	found := true
	for _, i := range c.Srv.Config.Calendars {
		if i.ID == id {
			cc := CalConfig{
				ID:       id,
				Name:     r.Form.Get("updName"),
				Provider: i.Provider,
				Colour:   r.Form.Get("updColour"),
			}

			p, err := c.getCalendarProvider(i.Provider)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			cc, err = p.ValidateConfig(cc)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

			cals = append(cals, cc)
			found = true
		} else {
			cals = append(cals, i)
		}
	}

	if found {
		c.Srv.Config.Calendars = cals
		if err := c.Srv.Config.WriteToFile("config.json"); err != nil {
			m := fmt.Sprintf("Error writing config.json file. %s", err.Error())
			c.LogError(m)
			http.Error(w, m, 500)
		}
	} else {
		http.Error(w, "Invalid calendar identifier", 500)
	}
}

func (c *ConfigController) handleRemoveCalendar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Calendar identifier not specified", 500)
		return
	}

	cl := []CalConfig{}
	ri := CalConfig{}
	found := false
	for _, i := range c.Srv.Config.Calendars {
		if i.ID == id {
			c.LogInfo(fmt.Sprintf("Calendar %s removed.", i.Name))
			ri = i
			found = true
		} else {
			cl = append(cl, i)
		}
	}
	if found {
		c.Srv.Config.Calendars = cl
		if err := c.Srv.Config.WriteToFile("config.json"); err != nil {
			m := fmt.Sprintf("Error writing config.json file. %s", err.Error())
			c.LogError(m)
			http.Error(w, m, 500)
		} else {
			if p, err := c.getCalendarProvider(ri.Provider); err == nil {
				err = p.RemovedConfig(ri)
				if err != nil {
					m := fmt.Sprintf("Error cleaning up %s for removed config item %s. %s", p.ProviderName(), ri.ID, err.Error())
					c.LogError(m)
					m = fmt.Sprintf("%v", ri)
					c.LogError(m)
				}
			}
		}

	} else {
		http.Error(w, "Invalid calendar identifier", 500)
	}
}

func (c *ConfigController) getCalendarProvider(provider string) (CalendarProvider, error) {
	switch provider {
	case "Google":
		return new(GCalendar), nil
	default:
		m := fmt.Sprintf("Invalid Calendar provider '%s'", provider)
		return nil, errors.New(m)
	}
}

// LogInfo is used to log information messages for this controller.
func (c *ConfigController) LogInfo(v ...interface{}) {
	a := fmt.Sprint(v)
	logger.Info("ConfigController: [Inf] ", a[1:len(a)-1])
}

// LogError is used to log error messages for this controller.
func (c *ConfigController) LogError(v ...interface{}) {
	a := fmt.Sprint(v)
	logger.Error("ConfigController: [Err] ", a[1:len(a)-1])
}
