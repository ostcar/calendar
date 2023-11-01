package model

import (
	"fmt"
	"time"
)

// Model holds the global data for the calendar.
type Model struct {
	location *time.Location
	events   map[string][]Event
}

// New initializes a model.
func New(location *time.Location, events []Event) *Model {
	eventMap := make(map[string][]Event, len(events))
	for _, e := range events {
		attr := dayAttr(e.start)
		eventMap[attr] = append(eventMap[attr], e)
	}
	return &Model{
		location: location,
		events:   eventMap,
	}
}

// ThisMonth returns the current month.
func (m Model) ThisMonth() Month {
	n := time.Now()
	return m.newMonth(n.Year(), n.Month())
}

// Month represents one month.
type Month struct {
	model *Model

	year  int
	month time.Month
}

func (m Model) newMonth(year int, month time.Month) Month {
	return Month{
		model: &m,
		year:  year,
		month: month,
	}
}

// Name returns a string representing the month.
func (m Month) Name() string {
	return fmt.Sprintf("%s %d", germanMonth(m.month), m.year)
}

// Weeks returns all weeks of the month.
func (m Month) Weeks() []Week {
	// Find monday monday <= the monday day of the month.
	monday := time.Date(m.year, m.month, 1, 0, 0, 0, 0, m.model.location)
	for monday.Weekday() != time.Monday {
		monday = monday.Add(-24 * time.Hour)
	}

	var weeks []Week
	before := m.month - 1
	if before == 0 {
		before = time.December
	}

	for monday.Month() == before || monday.Month() == m.month {
		weeks = append(weeks, m.newWeek(monday))
		monday = monday.Add(7 * 24 * time.Hour)
	}

	return weeks
}

// Next returns the next month.
func (m Month) Next() Month {
	year := m.year
	month := m.month + 1
	if month > time.December {
		month = time.January
		year++
	}

	return Month{
		model: m.model,
		year:  year,
		month: month,
	}
}

// Previous returns the previous month.
func (m Month) Previous() Month {
	year := m.year
	month := m.month - 1
	if month < time.January {
		month = time.December
		year--
	}

	return Month{
		model: m.model,
		year:  year,
		month: month,
	}
}

// Attr returns a string representation of the month.
func (m Month) Attr() string {
	return fmt.Sprintf("%d-%d", m.year, m.month)
}

// MonthFromAttr returns a Month from an attr.
func (m Model) MonthFromAttr(attr string) (Month, error) {
	var year int
	var month time.Month
	if n, err := fmt.Sscanf(attr, "%d-%d", &year, &month); n != 2 || err != nil {
		return Month{}, fmt.Errorf("invalid attr %s", attr)
	}

	if month < time.January || month > time.December {
		return Month{}, fmt.Errorf("invalid attr %s", attr)
	}

	return Month{
		model: &m,
		year:  year,
		month: month,
	}, nil
}

// Week ...
type Week struct {
	model  *Model
	monday time.Time
}

func (m Month) newWeek(monday time.Time) Week {
	return Week{
		model:  m.model,
		monday: monday,
	}
}

// Days returns the days of the month
func (w Week) Days() []Day {
	days := make([]Day, 7)
	for i := 0; i < 7; i++ {
		days[i] = w.newDay(w.monday.Add(time.Duration(i) * 24 * time.Hour))
	}
	return days
}

// Day ...
type Day struct {
	model *Model
	time  time.Time
}

func (w Week) newDay(start time.Time) Day {
	return Day{
		model: w.model,
		time:  start,
	}
}

// Number returns a number between 1 and 31.
func (d Day) Number() int {
	return d.time.Day()
}

// IsToday returns true, if this is the day.
func (d Day) IsToday() bool {
	yearA, monthA, dayA := time.Now().Date()
	yearB, monthB, dayB := d.time.Date()

	return yearA == yearB && monthA == monthB && dayA == dayB
}

// InMonth tells, if the day is in the monath.
func (d Day) InMonth(month Month) bool {
	return month.month == d.time.Month()
}

func dayAttr(t time.Time) string {
	return t.Format("2006-01-02")
}

// Events returns all events for that day.
func (d Day) Events() []Event {
	return d.model.events[dayAttr(d.time)]
}

func germanMonth(m time.Month) string {
	switch m {
	case time.January:
		return "Januar"
	case time.February:
		return "Februar"
	case time.March:
		return "März"
	case time.April:
		return "April"
	case time.May:
		return "Mai"
	case time.June:
		return "Juni"
	case time.July:
		return "Juli"
	case time.August:
		return "August"
	case time.September:
		return "September"
	case time.October:
		return "Oktober"
	case time.November:
		return "November"
	case time.December:
		return "Dezember"
	}
	panic(fmt.Sprintf("Invalid Month %s. I hope Go introduces enums.", m))
}

// Event is shown in a day.
type Event struct {
	id       string
	start    time.Time
	Title    string
	Subtitle string
}

// Time returns the start time of the event.
func (e Event) Time() string {
	return e.start.Format("15:04")
}
