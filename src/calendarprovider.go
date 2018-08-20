package main

// CalendarProvider defines an interface for Calendar providers
type CalendarProvider interface {
	GetEvents(maxEvents int) (CalEvents, error)
	SetConfig(c CalConfig)
}
