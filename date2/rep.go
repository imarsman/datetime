package date2

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"math"
	"time"
)

const secondsPerDay = 60 * 60 * 24

const billion = 1000000000

const epochYearGregorian int64 = 1970

// StartYear the beginning of the universe
// const StartYear = -billion * 15
const StartYear = epochYearGregorian - epochYearGregorian

// Gregorian epock year
const zeroYear int64 = 0
const zeroMonth int64 = 1
const zeroDay int64 = 1

// func fromGregorianYear(year int64) int64 {
// 	if year < 0 {
// 		// return StartYear + epochYearGregorian - year
// 		return StartYear - year
// 	}
// 	return StartYear + (epochYearGregorian - year)
// }

func gregorianYear(inputYear int64) (year int64, isCE bool) {
	// if year > 0 {
	// 	return 0, errors.New("parse.toGregorianYear: ncoming value must be < 0")
	// }
	// var diff int64
	// if year > (StartYear - epochYearGregorian) {
	// 	diff = StartYear - year + epochYearGregorian
	// } else {
	// 	diff = year - StartYear - epochYearGregorian
	// }
	year = inputYear
	if year == 0 {
		return 1, true
	}
	if year > (StartYear) {
		year = StartYear + year
	} else {
		year = StartYear - int64(math.Abs(float64(year)))
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
