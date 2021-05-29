package date2

// This package represents date whose zero value is 1 CE. It stores years as
// int64 values and as such the minimum and maximum possible years are very
// large. The Golang time package is used both for representing dates and times
// and for doing timing operations. This package is oriented toward doing things
// with dates.

import (
	"fmt"
	"math"
	"time"

	"github.com/imarsman/datetime/gregorian"
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

// The smallest possible year value
// const absoluteZeroYear = -(15 * 100 * 100 * 1000)

// const yearsToCE = 15 * 1000 * 1000 * 1000

// As far back as we will go - 15 billion years
// const zeroYear = absoluteZeroYear + 15*100*100*1000

// func bigYear(year int64) int64 {
// 	var bigYear int64 = astronomicalYear(year)
// 	// var bigYear int64 = year
// 	if bigYear <= 0 {
// 		return yearsToCE + year
// 	}
// 	return yearsToCE + year
// }

func countDaysForward(start int, count int64) int {
	total := int64(start) + count
	dow := total % 7
	if dow == 0 {
		dow = 7
	}

	return int(dow)
}

func countDaysBackward(start int, count int64) int {
	total := int64(start) + count
	dow := total % 7
	if dow < 0 {
		dow = -dow
	}
	if dow == 0 {
		dow = 7
	}

	return daysBackward[dow-1]
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

func (d Date) daysInMonth() (int, error) {
	err := d.validate()
	if err != nil {
		return 0, err
	}

	daysInMonth := gregorian.DaysInMonth[d.month]
	year := d.yearAbs()
	leapD, err := NewDate(year, 1, 1)
	isLeap := leapD.IsLeap()
	if isLeap && d.month == 2 {
		return 29, nil
	}
	return daysInMonth, nil
}

// YearDay returns the day of the year specified by d, in the range [1,365] for
// non-leap years, and [1,366] in leap years. The functionality should be the
// same as for the Go time.YearDay func.
func (d Date) YearDay() (int, error) {
	err := d.validate()
	if err != nil {
		return 0, err
	}
	var days int = 0
	copy := d
	copy.year = copy.astronomicalYear()
	for i := 1; i < 13; i++ {
		copy.month = i
		if copy.month > d.month {
			break
		}
		if copy.month == d.month {
			days += int(d.day)
			break
		}
		val, _ := copy.daysInMonth()
		days += int(val)
	}

	return days, nil
}

// For CE count forward from 1 Jan and for BCE count backward from 31 Dec
func (d Date) daysToDateFromAnchorDay() int {
	d2 := d
	d2.day = 1
	d2.month = 1
	var startDateDays int64 = int64(daysToAnchorDaySinceEpoch(d2.year))
	if startDateDays < 0 {
		startDateDays = -startDateDays
	}
	total := 0

	// Count forwards to date from 1 Jan
	if d.year > 0 {
		d3, _ := NewDate(d.year, 1, 1)
		for {
			daysInMonth, _ := d3.daysInMonth()
			if d3.month == d.month {
				total = total + d.day - 1
				return total
			}
			total = total + daysInMonth
			if d3.month >= 12 {
				break
			}
			d3.month++
		}
		// Count backwards to date from 31 Dec
	} else {
		d3, _ := NewDate(d.year, 12, 31)
		for {
			daysInMonth, _ := d3.daysInMonth()
			if d3.month == d.month {
				// fmt.Println("total", total)
				total = total + (daysInMonth - d.day)
				// fmt.Println("total", total)
				return total
			}
			total = total + daysInMonth
			if d3.month <= 1 {
				break
			}
			d3.month--
		}
	}

	return total
}

// daysToAnchorDaySinceEpoch takes a year and returns the number of days from
// the absolute epoch to the start of that year.
// This will work for CE but not for BCE
func daysToAnchorDaySinceEpoch(year int64) uint64 {
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
	leapDayCount = ((astronomicalYear / 4) - 1) - ((astronomicalYear / 100) - 1) + (astronomicalYear / 400)

	// TODO: See if there is interplay between the subtractions of 1 here and
	// the days to date count.
	// See https://www.thecalculatorsite.com/time/days-between-dates.php

	total := year * 365
	total -= 365
	total += leapDayCount
	fmt.Println("year", year, "leap days", leapDayCount, "total", total)

	if isLeap(astronomicalYear) {
		total--
	}

	return uint64(total)
}

// Weekday get day of week for a date
// See https://www.thecalculatorsite.com/time/days-between-dates.php
// - seems to calculate according to proleptic Gregorian calendar
func (d Date) Weekday() (int, error) {
	err := d.validate()
	if err != nil {
		return 0, err
	}

	year := d.year

	// This will count forward for CE and backward for BCE but give the results
	// in both cases as a positive integer.
	dateDays := int64(daysToAnchorDaySinceEpoch(year))

	daysSince1Jan := d.daysToDateFromAnchorDay()

	totalDays := dateDays + int64(daysSince1Jan)
	// fmt.Println("total days to", d.String(), totalDays)

	var dow int
	if year > 0 {
		// Count days forward, taking into account the starting day for
		// CE or BCE.
		dow = countDaysForward(firstDOWCE, totalDays)
	} else {
		// Count days backward, taking into account the starting day for
		// CE or BCE.
		dow = countDaysBackward(lastDOWBCE, totalDays)
	}

	return dow, nil
}

func isoWeekOfYearForDate(doy int, dow time.Weekday) int {
	var woy int = (10 + doy - int(dow)) % 7
	return woy
}
