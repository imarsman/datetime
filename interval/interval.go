package interval

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/imarsman/datetime/period"
	"github.com/imarsman/datetime/timestamp"
)

// Interval an ISO-8601 interval
type Interval struct {
	Repeat        bool
	RepeatNumber  int
	StartTime     *time.Time
	StartDuration *period.Period

	EndTime     *time.Time
	EndDuration *period.Period
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
func Parse(iString string) (*Interval, error) {
	i := Interval{}

	re := regexp.MustCompile("\\/|[\\-]{2}")

	parts := re.FindAllString(iString, -1)
	// var p1 string
	// var p2 string

	if len(parts) != 2 {
		return nil, errors.New("Too many parts in interval")
	}
	if strings.HasPrefix(parts[0], "P") {
		p, err := period.ParseWithNormalise(parts[0], true)
		if err != nil {
			return nil, err
		}
		i.StartDuration = &p
	} else {
		ts, err := timestamp.ParseTimestampInUTC(parts[0])
		if err != nil {
			return nil, err
		}
		i.StartTime = ts
	}
	if strings.HasPrefix(parts[1], "P") {
		p, err := period.ParseWithNormalise(parts[1], true)
		if err != nil {
			return nil, err
		}
		i.StartDuration = &p
	} else {

	}

	return nil, nil
}
