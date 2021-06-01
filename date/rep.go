package date

// This package represents date whose zero value is 1 CE. It stores years as
// int64 values and as such the minimum and maximum possible years are very
// large. The Golang time package is used both for representing dates and times
// and for doing timing operations. This package is oriented toward doing things
// with dates.

import (
	"math"
	"time"
)

const lastDOWBCE int = 7 // 31 December, 1 BCE is Friday

const firstDOWCE int = 1 // 1 January, 1 CE is Saturday

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

var daysForward = [...]int{1, 2, 3, 4, 5, 6, 7}

var daysBackward = [...]int{7, 6, 5, 4, 3, 2, 1}

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

// The largest possible year value
const absoluteMaxYear = math.MaxInt64

// The maximum year that will be dealt with
const maxYear = 15 * 100 * 100 * 1000
const minYear = -maxYear

func countDaysForward(start int, count int64) int {
	total := int64(start) + count
	dow := total % 7
	if dow == 0 {
		dow = 7
	}

	return int(dow)
}

// Not working properly yet - off by a day in tests
func countDaysBackward(start int, count int64) int {
	// When the count is less than or equal to the start day the result will be
	// a simple counting back of days.
	if count <= int64(start) {
		dow := int64(start) - count
		if dow == 0 {
			dow = 7
		}
		return int(dow)
	}
	// The count assumes a day of 1 but the actual starting day can be different
	total := count - int64(start)

	// Results will range from 0-6
	dow := total % 7

	return daysBackward[dow]
}

func gregorianYear(inputYear int64) (year int64) {
	year = inputYear
	if year == 0 {
		year = 1
	} else if year == -1 {
		// Redundant
		year = -1
	}

	return year
}

// daysToAnchorDayFromEpoch takes a year and returns the number of days from
// the absolute epoch to the start of that year.
// This will work for CE but not for BCE
func daysToAnchorDayFromEpoch(year int64) int64 {
	var leapDayCount int64

	if year < 0 {
		year = -year
	}

	// Leap is calculated based on astronomical year
	astronomicalYear := astronomicalYear(year)
	if year < 0 {
		year = -year
	}

	// - add all years divisible by 4
	// - subtract all years divisible by 100
	// - add back all years divisible by 400
	leapDayCount = (astronomicalYear / 4) - (astronomicalYear / 100) + (astronomicalYear / 400)

	// TODO: See if there is interplay between the subtractions of 1 here and
	// the days to date count.
	// See https://www.thecalculatorsite.com/time/days-between-dates.php

	total := year * 365
	total -= 365

	total += leapDayCount

	if isLeap(astronomicalYear) {
		total--
	}

	return total
}

// This is orphaned and should not be
func isoWeekOfYearForDate(doy int, dow time.Weekday) int {
	var woy int = (10 + doy - int(dow)) % 7
	return woy
}
