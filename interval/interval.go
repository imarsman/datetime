package interval

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/imarsman/datetime/isodate"
	"github.com/imarsman/datetime/period"
	"github.com/imarsman/datetime/timestamp"
)

// Interval an ISO-8601 interval
type Interval struct {
	Repeat       bool
	RepeatNumber int
	StartTime    *time.Time
	StartPeriod  *period.Period

	EndTime   *time.Time
	EndPeriod *period.Period
}

// IsRepeating is the interval repeating
func (i *Interval) IsRepeating() bool {
	return i.Repeat
}

// RepeatCount the number of repeats for the interval
func (i *Interval) RepeatCount() int {
	return i.RepeatNumber
}

// RepeatInfinite is the number of repeats infinite
func (i *Interval) RepeatInfinite() bool {
	return i.Repeat == true && i.RepeatNumber == 0
}

// Parse parse an interval and populate new Interval struct
func Parse(iString string) (Interval, error) {
	i := Interval{}

	re := regexp.MustCompile("\\/|[\\-]{2}")

	parts := re.FindAllString(iString, -1)
	// var p1 string
	// var p2 string

	if len(parts) != 2 {
		return Interval{}, errors.New("Too many parts in interval")
	}
	if strings.HasPrefix(parts[0], "P") {
		p, err := period.ParseWithNormalise(parts[0], true)
		if err != nil {
			return Interval{}, err
		}
		i.StartPeriod = &p

	} else {
		dt, err := isodate.ParseISO(parts[0])
		if err != nil {
			t, err := timestamp.ParseUTC(parts[0])
			if err != nil {
				return Interval{}, err
			}
			i.StartTime = &t
		}
		t := dt.UTC()
		i.StartTime = &t
	}
	if strings.HasPrefix(parts[1], "P") {
		p, err := period.ParseWithNormalise(parts[1], true)
		if err != nil {
			return Interval{}, err
		}
		i.StartPeriod = &p
	} else {
		t, err := timestamp.ParseUTC(parts[1])
		if err != nil {
			return Interval{}, err
		}
		i.EndTime = &t
	}
	if i.StartTime == nil && i.EndTime == nil {
		return Interval{}, errors.New("One of start or end must be a time")
	}
	if i.StartTime == nil {
		offset := i.EndTime.Add(i.EndPeriod.Abs().Negate().DurationApprox())
		i.StartTime = &offset
	}
	if i.EndTime == nil {
		offset := i.EndTime.Add(i.EndPeriod.Abs().DurationApprox())
		i.EndTime = &offset
	}

	return Interval{}, nil
}
