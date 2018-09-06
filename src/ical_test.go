package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/brumawen/ical"
)

func TestCanGetIcalFeed(t *testing.T) {
	resp, err := http.Get("https://calendar.google.com/calendar/ical/en.sa%23holiday%40group.v.calendar.google.com/public/basic.ics")

	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		t.Error(err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	if len(b) == 0 {
		t.Error("No feed returned.")
	}
}

func TestCanParseIcalFeed(t *testing.T) {
	resp, err := http.Get("https://calendar.google.com/calendar/ical/en.sa%23holiday%40group.v.calendar.google.com/public/basic.ics")
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		t.Error(err)
	}
	c, err := ical.Parse(resp.Body, time.Local)
	if err != nil {
		t.Error(err)
	}
	if len(c.Events) == 0 {
		t.Error("No events parsed.")
	}
	for _, e := range c.Events {
		fmt.Printf("%v, %v, %s\r\n", e.StartDate, e.EndDate, e.Summary)
	}
}

func TestCanGetICalEvents(t *testing.T) {
	url := "https://calendar.google.com/calendar/ical/en.sa%23holiday%40group.v.calendar.google.com/public/basic.ics"
	c := CalConfig{
		ID:   "testical",
		Name: "Test iCal",
		URL:  fmt.Sprintf("%s\n%s", url, url),
	}

	p := ICalFeed{CalConfig: c}
	l, err := p.GetEvents(365)
	if err != nil {
		t.Error(err)
	}
	if len(l.Events) == 0 {
		t.Error(errors.New("No events returned"))
	} else if len(l.Events) != 21 {
		t.Errorf("Wrong number of events returned. Expected %d, got %d", 21, len(l.Events))
	}

}
