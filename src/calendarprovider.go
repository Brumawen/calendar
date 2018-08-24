package main

// CalendarProvider defines an interface for Calendar providers
type CalendarProvider interface {
	GetEvents(maxEvents int) (CalEvents, error)
	ProviderName() string
	SetConfig(c CalConfig)
	ValidateConfig(c CalConfig) (CalConfig, error)
	ValidateNewConfig(c NewCalConfig) (CalConfig, error)
}
