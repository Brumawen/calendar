package main

import (
	"errors"
	"testing"
)

func TestCanRetrieveCalendarEvents(t *testing.T) {

}

func TestCanGetGoogleEvents(t *testing.T) {
	c := CalConfig{
		ID:   "testok",
		Name: "Test",
	}
	g := GCalendar{CalConfig: c}
	l, err := g.GetEvents(14)
	if err != nil {
		t.Error(err)
	}
	if len(l.Events) == 0 {
		t.Error(errors.New("No events returned"))
	}
}

func TestCanGetGoogleEvents1(t *testing.T) {
	c := CalConfig{
		ID:   "b8125164-12eb-4f96-b7b1-96cf48d35a99",
		Name: "Test",
	}
	g := GCalendar{CalConfig: c}
	l, err := g.GetEvents(14)
	if err != nil {
		t.Error(err)
	}
	if len(l.Events) == 0 {
		t.Error(errors.New("No events returned"))
	}
}

func TestCanGetLastGoogleEvents(t *testing.T) {
	c := CalConfig{
		ID:   "test",
		Name: "Test",
	}
	g := GCalendar{CalConfig: c}
	l, err := g.GetEvents(10)
	if err == nil {
		t.Error(errors.New("No error returned"))
	}
	if len(l.Events) == 0 {
		t.Error(errors.New("No events returned"))
	}
}
