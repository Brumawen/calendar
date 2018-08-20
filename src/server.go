package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	gopifinder "github.com/brumawen/gopi-finder/src"
	"github.com/gorilla/mux"
	"github.com/kardianos/service"
)

// Server defines the Calendar Web Service.
type Server struct {
	PortNo         int               // Port No the server will listen on
	VerboseLogging bool              // Verbose logging on/ off
	Timeout        int               // Timeout waiting for a response from an IP probe.  Defaults to 2 seconds.
	Config         *Config           // Configuration settings
	NoReg          bool              // Do not register with the finder server
	Finder         gopifinder.Finder // Finder client - used to find other devices
	exit           chan struct{}     // Exit flag
	shutdown       chan struct{}     // Shutdown complete flag
	http           *http.Server      // HTTP server
	router         *mux.Router       // HTTP router
	isregistering  bool              // Indicates that a registration is currently ongoing
}

// Start is called when the service is starting
func (s *Server) Start(v service.Service) error {
	s.logInfo("Service starting")

	// Make sure the working directory is the same as the application exe
	ap, err := os.Executable()
	if err != nil {
		s.logError("Error getting the executable path.", err.Error())
	} else {
		wd, err := os.Getwd()
		if err != nil {
			s.logError("Error getting current working directory.", err.Error())
		} else {
			ad := filepath.Dir(ap)
			s.logInfo("Current application path is", ad)
			if ad != wd {
				if err := os.Chdir(ad); err != nil {
					s.logError("Error chaning working directory.", err.Error())
				}
			}
		}
	}

	// Create a channel that will be used to block until the Stop signal is received
	s.exit = make(chan struct{})
	go s.run()
	return nil
}

// Stop is called when the service is stopping
func (s *Server) Stop(v service.Service) error {
	s.logInfo("Service stopping")
	// Close the channel, this will automatically release the block
	s.shutdown = make(chan struct{})
	close(s.exit)
	// Wait for the shutdown to complete
	_ = <-s.shutdown
	return nil
}

// run will start up and run the service and wait for a Stop signal
func (s *Server) run() {
	if s.PortNo < 0 {
		s.PortNo = 20513
	}
	s.Finder.Logger = logger
	s.Finder.VerboseLogging = service.Interactive()

	// Get the configuration
	if s.Config == nil {
		s.Config = &Config{}
	}
	s.Config.ReadFromFile("config.json")
	s.Config.SetDefaults()

	// Create a router
	s.router = mux.NewRouter().StrictSlash(true)
	s.router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./html/assets"))))

	// Add the controllers
	s.addController(new(LogController))
	s.addController(new(ConfigController))
	s.addController(new(CalendarController))

	// Create an HTTP server
	s.http = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.PortNo),
		Handler: s.router,
	}

	if s.NoReg {
		s.logInfo("Not registering service with Finder server.")
	} else {
		s.logInfo("Registering service with Finder server.")
		go func() {
			// Register service with the Finder server
			go s.RegisterService()
		}()
	}

	// Start the web server
	go func() {
		s.logInfo("Server listening on port", s.PortNo)
		if err := s.http.ListenAndServe(); err != nil {
			msg := err.Error()
			if !strings.Contains(msg, "http: Server closed") {
				s.logError("Error starting Web Server.", err.Error())
			}
		}
	}()

	// Wait for an exit signal
	_ = <-s.exit

	// Shutdown the HTTP server
	s.http.Shutdown(nil)

	s.logDebug("Shutdown complete")
	close(s.shutdown)
}

// AddController adds the specified web service controller to the Router
func (s *Server) addController(c Controller) {
	c.AddController(s.router, s)
}

// RegisterService will register the service with the devices on the network
func (s *Server) RegisterService() {
	if s.isregistering {
		return
	}
	s.isregistering = true
	isReg := false
	s.logDebug("Starting service registration.")
	for !isReg {
		s.logDebug("RegisterService: Getting device info")
		d, err := gopifinder.NewDeviceInfo()
		if err != nil {
			s.logError("Error getting device info.", err.Error())
		}
		s.logDebug("RegisterService: Creating service")
		sv := d.CreateService("WeatherForecast")
		sv.PortNo = s.PortNo

		if sv.IPAddress == "" {
			s.logDebug("RegisterService: No IP address found.")
		} else {
			s.logDebug("RegisterService: Using IP address", sv.IPAddress)
		}

		s.logDebug("Reg: Finding devices")
		_, err = s.Finder.FindDevices()
		if err != nil {
			s.logError("RegisterService: Error getting list of devices.", err.Error())
		} else {
			if len(s.Finder.Devices) == 0 {
				s.logDebug("RegisterService: Sleeping")
				time.Sleep(15 * time.Second)
			} else {
				// Register the services with the devices
				s.logDebug("RegisterService: Registering the service.")
				s.Finder.RegisterServices([]gopifinder.ServiceInfo{sv})
				isReg = true
			}
		}
	}
	s.logDebug("Completed service registration.")
	s.isregistering = false
}

// logDebug logs a debug message to the logger
func (s *Server) logDebug(v ...interface{}) {
	if s.VerboseLogging {
		a := fmt.Sprint(v)
		logger.Info("Server: [Dbg] ", a[1:len(a)-1])
	}
}

// logInfo logs an information message to the logger
func (s *Server) logInfo(v ...interface{}) {
	a := fmt.Sprint(v)
	logger.Info("Server: [Inf] ", a[1:len(a)-1])
}

// logError logs an error message to the logger
func (s *Server) logError(v ...interface{}) {
	a := fmt.Sprint(v)
	logger.Error("Server: [Err] ", a[1:len(a)-1])
}
