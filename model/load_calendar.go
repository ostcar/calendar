package model

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const calURL = "https://kalender.evlks.de/json?vid=98"

// LoadEvents loads all calendar events.
func LoadEvents() ([]Event, error) {
	resp, err := http.Get(calURL)
	if err != nil {
		return nil, fmt.Errorf("fetch events: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		io.ReadAll(resp.Body)
		return nil, fmt.Errorf("got status %d, only accept 200", resp.StatusCode)
	}

	type jsonEvent struct {
		Veranstaltung struct {
			ID       string `json:"ID"`
			Start    string `json:"START_RFC"`
			Title    string `json:"_event_TITLE"`
			SubTitle string `json:"SUBTITLE"`
		} `json:"Veranstaltung"`
	}

	var jsonEvents []jsonEvent

	if err := json.NewDecoder(resp.Body).Decode(&jsonEvents); err != nil {
		return nil, fmt.Errorf("read and parse body: %w", err)
	}

	events := make([]Event, len(jsonEvents))
	for i, je := range jsonEvents {
		start, err := time.Parse("2006-01-02T15:04:05.000-07:00", je.Veranstaltung.Start)
		if err != nil {
			return nil, fmt.Errorf("parsing event %d: %w", i, err)
		}

		events[i] = Event{
			id:       je.Veranstaltung.ID,
			start:    start,
			Title:    je.Veranstaltung.Title,
			Subtitle: je.Veranstaltung.SubTitle,
		}
	}

	return events, nil
}
