package date2

// This package represents date whose zero value is 1 CE. It stores years as
// int64 values and as such the minimum and maximum possible years are very
// large. The Golang time package is used both for representing dates and times
// and for doing timing operations. This package is oriented toward doing things
// with dates.

import (
	"math"
	"time"
)

// Month a month - int
type Month int

const (
	// January the month of January
	January Month = 1 + iota
	// February the month of February
	February
	// March the month of March
	March
	// April the month of April
	April
	// May the month of May
	May
	// June the month of June
	June
	// July the month of July
	July
	// August the month of August
	August
	// September the month of September
	September
	// October the month of October
	October
	// November the month of November
	November
	// December the month of December
	December
)

// Weekday a weekday (int)
type Weekday int

// From Golang time package
const (
	// Monday the day Monday
	Monday Weekday = 1 + iota
	// Tuesday the day Tuesday
	Tuesday
	// Wednesday the day Wednesday
	Wednesday
	// Thursday the day Thursday
	Thursday
	// Friday the day Friday
	Friday
	// Saturday the day Saturday
	Saturday
	// Sunday the day Sunday
	Sunday
)

// From Golang time package
const (
	secondsPerMinute = 60
	secondsPerHour   = 60 * secondsPerMinute
	secondsPerDay    = 24 * secondsPerHour
	secondsPerWeek   = 7 * secondsPerDay
	daysPer400Years  = 365*400 + 97
	daysPer100Years  = 365*100 + 24
	daysPer4Years    = 365*4 + 1
)

const billion = 1000000000

const epochYearGregorian int64 = 1970

// StartYear the beginning of the universe
const StartYear = epochYearGregorian - epochYearGregorian

const absoluteMaxYear = math.MaxInt64
const absoluteZeroYear = math.MinInt64

// Gregorian epock year
// const zeroYear int64 = 0
// const zeroMonth int64 = 1
// const zeroDay int64 = 1

func gregorianYear(inputYear int64) (year int64, isCE bool) {
	year = inputYear
	if year == 0 {
		return 1, true
	}
	if year > (StartYear) {
		year = StartYear + year
	} else {
		year = StartYear - -year
	}
	return year, year > 0
}

// encode returns the number of days elapsed from date zero to the date
// corresponding to the given Time value.
func encode(t time.Time) PeriodOfDays {
	// Compute the number of seconds elapsed since January 1, 1970 00:00:00
	// in the location specified by t and not necessarily UTC.
	// A Time value is represented internally as an offset from a UTC base
	// time; because we want to extract a date in the time zone specified
	// by t rather than in UTC, we need to compensate for the time zone
	// difference.
	_, offset := t.Zone()
	secs := t.Unix() + int64(offset)
	// Unfortunately operator / rounds towards 0, so negative values
	// must be handled differently
	if secs >= 0 {
		return PeriodOfDays(secs / secondsPerDay)
	}
	return -PeriodOfDays((secondsPerDay - 1 - secs) / secondsPerDay)
}

// decode returns the Time value corresponding to 00:00:00 UTC of the date
// represented by d, the number of days elapsed since date zero.
func decode(d PeriodOfDays) time.Time {
	secs := int64(d) * secondsPerDay
	return time.Unix(secs, 0).UTC()
}
