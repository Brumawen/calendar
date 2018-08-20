package main

import (
	"errors"
	"fmt"
	"math"
	"testing"
	"time"
)

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

func TestMoo(t *testing.T) {
	s := time.Date(2018, 8, 19, 1, 0, 0, 0, time.Local)
	e := time.Date(2018, 8, 19, 2, 0, 0, 0, time.Local)
	d := e.Sub(s)
	h := d.Hours()
	m := math.Mod(d.Minutes(), 60)
	fmt.Printf("%2.fh %2.fm\r\n", h, m)
}
