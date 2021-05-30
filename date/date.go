package date

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/imarsman/datetime/gregorian"
)

// PeriodOfDays describes a period of time measured in whole days. Negative values
// indicate days earlier than some mark.
// type PeriodOfDays int64

// // ZeroDays is the named zero value for PeriodOfDays.
// const ZeroDays PeriodOfDays = 0

// A Date represents a date under the (proleptic) Gregorian calendar as
// used by ISO 8601. This calendar uses astronomical year numbering,
// so it includes a year 0 and represents earlier years as negative numbers
// (i.e. year 0 is 1 CE; year -1 is 1 BCE, and so on).
//
// TODO: update this to reflect altered design
// A Date value requires 4 bytes of storage and can represent dates from
// Tue, 23 Jun -5,877,641 (5,877,642 BC) to Fri, 11 Jul 5,881,580.
// Dates outside that range will "wrap around".
//
// Programs using dates should typically store and pass them as values,
// not pointers.  That is, date variables and struct fields should be of
// type date.Date, not *date.Date.  A Date value can be used by
// multiple goroutines simultaneously.
//
// Date values can be compared using the Before, After, and Equal methods
// as well as the == and != operators.
//
// The Sub method subtracts two dates, returning the number of days between
// them. The Add method adds a Date and a number of days, producing a Date.
//
// The zero value of type Date is Thursday, January 1, 1970 (called 'the
// epoch'), based on Unix convention. As this date is unlikely to come up in
// practice, the IsZero method gives a simple way of detecting a date that
// has not been initialized explicitly.
//
// The first official date of the Gregorian calendar was Friday, October 15th
// 1582, quite unrelated to the epoch used here. The Date type does not
// distinguish between official Gregorian dates and earlier proleptic dates,
// which can also be represented when needed.
//

// Date a date holds values for year, month, and day as well as a flag tied to
// instantiation. The Gregorian calendar starts on year 1, not zero. The year
// before year 1 is 1 BCE, or -1. This can cause issues for calculating things
// like leap years for years before 1 CE. Years that are stored when creating a
// date instance using NewDate are adjusted to their proper Gregorian value.
// When calculations are required for things like leap year a version of the
// year for mathematical calculations is obtained. Since the year must in some
// circumstances be adjusted to be proper in a Gregorian sense a flag is set in
// the NewDate function to set safelyInstantiated to true rather than its
// default value of false. A check function can then determine if a date struct
// instance was accidentally created with new(Date) or Date{}. Calls that
// require this check will return an error.
type Date struct {
	year               int64
	month              int
	day                int
	safelyInstantiated bool
}

// NewDate returns the Date value corresponding to the given year, month, and day.
//
// The month and day may be outside their usual ranges and will be normalized
// during the conversion.
func NewDate(year int64, month int, day int) (Date, error) {
	d := Date{}

	d.year = gregorianYear(year)
	d.month = month
	d.day = day
	d.safelyInstantiated = true

	err := d.validate()
	if err != nil {
		return Date{}, err
	}

	return d, nil
}

func (d *Date) validate() error {
	if d.IsZero() || d.safelyInstantiated == false {
		return errors.New("date invalidly instantiated with new(Date) or Date{}")
	}
	if d.month < 1 || d.month > 12 {
		return fmt.Errorf("invalid month %d for date", d.month)
	}
	daysInMonth := d.daysInMonth()
	// Ensure valid day number
	if d.day > daysInMonth {
		return fmt.Errorf("invalid day %d for month %d", d.day, d.month)
	}

	return nil
}

func (d Date) yearAbs() int64 {
	if d.year < 0 {
		return -d.year
	}
	return d.year
}

func astronomicalYear(year int64) int64 {
	if year == 0 {
		return 1
	} else if year <= -1 {
		// fmt.Println("astronomical year", year)
		year++
		// fmt.Println("astronomical year", year)
	}

	return year
}

func (d Date) astronomicalYear() int64 {
	// year := d.year

	// if year == 1 {
	// 	return 1
	// } else if year <= -1 {
	// 	year++
	// }
	return astronomicalYear(d.year)
}

// IsCE is the year CE
func (d Date) IsCE() bool {
	// err := d.validate()
	// if err != nil {
	// 	return false, err
	// }
	return d.year > 0
}

// Year get year for date
func (d Date) Year() int64 {
	return d.year
}

// Month get month for year
func (d Date) Month() int {
	return d.month
}

// Day returns the day of the month specified by d.
// The first day of the month is 1.
func (d Date) Day() int {
	return d.day
}

func (d Date) daysInMonth() int {
	// validate calls this so there should be a valid month and year
	daysInMonth := gregorian.DaysInMonth[d.month]
	year := d.yearAbs()
	leapD := d
	leapD.year = year
	isLeap := leapD.IsLeap()
	if isLeap && d.month == 2 {
		return 29
	}

	return daysInMonth
}

// YearDay returns the day of the year specified by d, in the range [1,365] for
// non-leap years, and [1,366] in leap years. The functionality should be the
// same as for the Go time.YearDay func.
func (d Date) YearDay() int {
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
		val := copy.daysInMonth()
		days += int(val)
	}

	return days
}

// subtractDays add days to a date
// TODO: Fix this so that it works properly
func (d Date) subtractDays(subtract int) (date Date, err error) {
	d2 := d
	// fmt.Println("subtracing from", d.String(), subtract)
	newYear := false
	for {
		daysInMonth := d2.daysInMonth()
		if err != nil {
			return Date{}, err
		}
		// fmt.Println("date", d2.String(), d2.month == 1 && d2.day == 1)
		if d2.month == 1 && d2.day == 1 {
			// fmt.Println("equal")
			d2.year--
			newYear = true
		}
		if subtract > daysInMonth {
			// fmt.Println("d2.day", d2.day)
			if newYear {
				// fmt.Println("date", d2.String(), d2.month == 1 && d2.day == 1)
				d2.day = 1
				d2.month = 12

				newYear = false
			}
			// if d2.day == daysInMonth {
			// 	// fmt.Println("date", d2.String(), d2.month == 1 && d2.day == 1)
			// 	subtract = subtract - d2.day
			// 	// fmt.Println("d2.day", d2.day)
			// 	d2.day = 1

			// 	continue
			// } else {
			// 	// fmt.Println("date", d2.String(), d2.month == 1 && d2.day == 1)
			// 	// d2.day--
			// 	// fmt.Println("date", d2.String(), d2.month == 1 && d2.day == 1)
			// }
			fmt.Println("d2.day", d2.day)
			d2.month--
			fmt.Println("more subtract", subtract, "daysinmonth", daysInMonth, "d2.day", d2.day, "d2", d2.String())
			subtract = subtract - daysInMonth
			d2.day -= subtract
			fmt.Println("more subtract", subtract, "daysinmonth", daysInMonth, "d2.day", d2.day, "d2", d2.String())
			// fmt.Println("date", d2.String(), d2.month == 1 && d2.day == 1)
		} else {
			// subtract = subtract - d2.day
			// fmt.Println("date", d2.String(), d2.month == 1 && d2.day == 1)
			// curDay := d2.day
			// fmt.Println("subtract", subtract, "daysinmonth", daysInMonth, "d2.day", d2.day, "d2", d2.String())
			d2.day = d2.day - subtract
			// subtract = -curDay
			fmt.Println("less subtract", subtract, "daysinmonth", daysInMonth, "d2.day", d2.day, "d2", d2.String())
			fmt.Println("d2.day", d2.day)
			// fmt.Println("date", d2.String(), d2.month == 1 && d2.day == 1)
			subtract--
			if subtract == 0 {
				break
			}
		}
	}

	return d2, nil
}

func (d Date) daysToDateFromEpoch() uint64 {
	daysToAnchorDate := daysToAnchorDayFromEpoch(d.year)
	daysToDate := d.daysToDateFromAnchorDay()
	totalDays := daysToAnchorDate + uint64(daysToDate)

	return totalDays
}

// AddParts add in succession years, months, and days to a date
// - add years and then increment the date using the remainder days
// - add days for each month to date
// - add starting days to date
func (d Date) AddParts(years int64, months, days int) (newDate Date, remainder int64, err error) {
	dFinal := d

	// Add years by finding out how many days need to be added then adding the
	// number of years for the days and then the number of days remaining.
	if years > 0 {
		var totalDays uint64
		if d.year < 0 {
			// Get total days BCE and CE
			if d.year+years > 0 {
				startDays := d.daysToDateFromEpoch()

				dEnd := d
				// The year will cross the zero boundary
				dEnd.year = d.year + years + 1
				endDays := dEnd.daysToDateFromEpoch()

				totalDays = endDays + startDays
				newYears := totalDays / 365
				dFinal.year += int64(newYears) + 1
				remainder := totalDays % 365
				dFinal, err = dFinal.addDays(int(remainder))
			} else {
				startDays := d.daysToDateFromEpoch()

				dEnd := d
				// The year will cross the zero boundary
				dEnd.year = d.year - years
				endDays := dEnd.daysToDateFromEpoch()

				totalDays = endDays - startDays
				newYears := totalDays / 365
				dFinal.year += int64(newYears)
				remainder := totalDays % 365
				dFinal, err = dFinal.addDays(int(remainder))
			}
		} else {
			startDays := d.daysToDateFromEpoch()

			dEnd := d
			dEnd.year = d.year + years
			endDays := dEnd.daysToDateFromEpoch()

			totalDays = endDays - startDays
			newYears := totalDays / 365
			dFinal.year += int64(newYears)
			remainder := totalDays % 365
			dFinal, err = dFinal.addDays(int(remainder))
		}
	}
	// Month handling adds surplus days to dFinal
	if months > 0 {
		// Add months, which works by getting days for each month then adding
		// that many days to the date.
		dFinal, err = dFinal.addMonths(months)
		if err != nil {
			fmt.Println("error", err)
			return Date{}, 0, err
		}
	}

	// Add any days requested initially plus any days
	if days > 0 {
		dFinal, err = dFinal.addDays(days)
		if err != nil {
			fmt.Println("error", err)
			return Date{}, 0, err
		}
	}

	return dFinal, remainder, nil
}

// addMonths add months to a date
// TODO: decide if it would be good to add days
func (d Date) addMonths(add int) (d2 Date, err error) {
	d2 = d

	dNeutral := d2
	// Should probably use d.year
	dNeutral.year = 2019
	daysInMonth := dNeutral.daysInMonth()
	daysInMonth -= int(d2.day)
	for i := 0; i < add; i++ {
		d2, err = d2.addDays(daysInMonth)
		if err != nil {
			fmt.Println(err)
		}
	}

	// Do implicit validation of result
	_, err = NewDate(d2.year, d2.month, d2.day)
	if err != nil {
		return Date{}, err
	}

	return d2, nil
}

// addDays add days to a date
// Reasonably efficient given that it adds as many days at a time as possible.
func (d Date) addDays(add int) (date Date, err error) {
	d2 := d

	daysInMonth := d2.daysInMonth()
	if err != nil {
		return Date{}, err
	}
	for add > 0 {
		if add >= daysInMonth && d2.day <= daysInMonth {
			if d2.month >= 12 {
				d2.year++
				if d2.year == 0 {
					d2.year = 1
				}
				d2.month = 1
				d2.day = 1
				add = add - daysInMonth
				daysInMonth = d2.daysInMonth()
			} else {
				d2.month++
				d2.day = 1

				daysInMonth = d2.daysInMonth()

				daysTilEOM := daysInMonth - d2.day
				add = add - daysTilEOM

				if daysTilEOM == 0 {
					d2.day = 1
					d2.month++
					add--
				} else {
					add = add - daysTilEOM
					d2.month++
					d2.day = 1
					add--
				}
				if d2.month >= 12 {
					d2.year++
					if d2.year == 0 {
						d2.year = 1
					}
					d2.month = 1
				}
			}
			continue
		} else {
			if add+d2.day >= daysInMonth {
				add = add - d2.day
				d2.day = 1
				d2.month++
				if d2.month > 12 {
					d2.month = 1
					d2.year++
					if d2.year == 0 {
						d2.year = 1
					}
				}
				// Get days in month once month has had a chance to be corrected
				daysInMonth = d2.daysInMonth()
				// fmt.Println("new date", d2.String())

				continue
			}
			// There are fewer days to add than are remaining in the month
			d2.day += add
			add = 0

			break
		}
	}
	_, err = NewDate(d2.year, d2.month, d2.day)
	// Give a chance for flawed function logic to be found by invalid date
	if err != nil {
		return Date{}, err
	}

	return d2, nil
}

// For CE count forward from 1 Jan and for BCE count backward from 31 Dec
func (d Date) daysToDateFromAnchorDay() int {
	d2 := d
	d2.day = 1
	d2.month = 1
	var startDateDays int64 = int64(daysToAnchorDayFromEpoch(d2.year))
	if startDateDays < 0 {
		startDateDays = -startDateDays
	}
	total := 0

	// Count forwards to date from 1 Jan
	if d.year > 0 {
		// Ignore error because whatever created attached date was validated already
		d3, _ := NewDate(d.year, 1, 1)
		for {
			daysInMonth := d3.daysInMonth()
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
			// Ignore error because whatever created attached date was validated already
			daysInMonth := d3.daysInMonth()
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

// Weekday get day of week for a date
// See https://www.thecalculatorsite.com/time/days-between-dates.php
// - seems to calculate according to proleptic Gregorian calendar
func (d Date) Weekday() int {
	year := d.year

	// This will count forward for CE and backward for BCE but give the results
	// in both cases as a positive integer.
	dateDays := int64(daysToAnchorDayFromEpoch(year))

	daysFromAnchorDay := d.daysToDateFromAnchorDay()

	totalDays := dateDays + int64(daysFromAnchorDay)
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

	return dow
}

// ISOWeeksInYear get number of ISO weeks in year
func (d Date) ISOWeeksInYear() int {
	year := d.year
	if year < 0 {
		year = -year
	}

	dMathematical, _ := NewDate(year, 1, 1)
	year = dMathematical.astronomicalYear()

	p := math.Mod(float64(year+(year/4)-(year/100)+(year/400)), 7)
	weeks := 52
	if p == 4 || p-1 == 3 {
		weeks++
	}

	return weeks
}

// Today returns today's date according to the current local time.
func Today() Date {
	t := time.Now()
	d, _ := NewDate(int64(t.Year()), int(t.Month()), t.Day())

	return d
}

// Min returns the smallest representable date.
func Min() Date {
	year := gregorianYear(minYear)
	d, _ := NewDate(year, 1, 1)

	return d
}

// Max returns the largest representable date.
func Max() Date {
	year := gregorianYear(maxYear)
	d, _ := NewDate(year, 1, 1)

	return d
}

// Date returns the year, month, and day of d.
// The first day of the month is 1.
func (d Date) Date() (year int64, month int, day int) {
	return d.year, d.month, d.day
}

// LastDayOfMonth returns the last day of the month specified by d.
// The first day of the month is 1.
func (d Date) LastDayOfMonth() int {
	dim := d.daysInMonth()

	return dim
}

// IsZero reports whether t represents the zero date.
func (d Date) IsZero() bool {
	return (d.year == 0 && d.month == 0 && d.day == 0)
}

// Equal reports whether d and u represent the same date.
func (d Date) Equal(u Date) bool {
	return d.year == u.year &&
		d.month == u.month &&
		d.day == u.day
}

// IsBefore reports whether the date d is before u.
func (d Date) IsBefore(u Date) bool {
	dSum := d.year*1000 + int64(d.month*100) + int64(d.day)
	uSum := u.year*1000 + int64(u.month*100) + int64(u.day)

	return dSum < uSum
}

// IsAfter reports whether the date d is after u.
func (d Date) IsAfter(u Date) bool {
	dSum := d.year*1000 + int64(d.month*100) + int64(d.day)
	uSum := u.year*1000 + int64(u.month*100) + int64(u.day)

	return dSum > uSum
}

// MinDate returns the earlier of two dates.
func (d Date) MinDate(u Date) Date {
	if u.IsBefore(d) {
		return u
	}
	return d
}

// MaxDate returns the later of two dates.
func (d Date) MaxDate(u Date) Date {
	if u.IsAfter(d) {
		return u
	}
	return d
}

// Add returns the date d plus the given number of days. The parameter may be negative.
// func (d Date) Add(days PeriodOfDays) Date {
// 	// 	return Date{d.day + days}
// 	return Date{}
// }

// AddDate returns the date corresponding to adding the given number of years,
// months, and days to d. For example, AddData(-1, 2, 3) applied to
// January 1, 2011 returns March 4, 2010.
//
// AddDate normalizes its result in the same way that Date does,
// so, for example, adding one month to October 31 yields
// December 1, the normalized form for November 31.
//
// The addition of all fields is performed before normalisation of any; this can affect
// the result. For example, adding 0y 1m 3d to September 28 gives October 31 (not
// November 1).
// func (d Date) AddDate(years, months, days int) Date {
// 	// t := decode(d.day).AddDate(years, months, days)
// 	// return Date{encode(t)}
// 	return Date{}
// }

// AddPeriod returns the date corresponding to adding the given period. If the
// period's fields are be negative, this results in an earlier date.
//
// Any time component is ignored. Therefore, be careful with periods containing
// more that 24 hours in the hours/minutes/seconds fields. These will not be
// normalised for you; if you want this behaviour, call delta.Normalise(false)
// on the input parameter.
//
// For example, PT24H adds nothing, whereas P1D adds one day as expected. To
// convert a period such as PT24H to its equivalent P1D, use
// delta.Normalise(false) as the input.
//
// See the description for AddDate.
// func (d Date) AddPeriod(delta period.Period) (Date, error) {
// 	newDate, err := d.AddDate(delta.Years(), delta.Months(), delta.Days())
// 	return newDate, err
// }

// Sub returns d-u as the number of days between the two dates.
// func (d Date) Sub(u Date) (days PeriodOfDays) {
// 	return d.day - u.day
// }

// DaysSinceEpoch returns the number of days since the epoch (1st January 1970), which may be negative.
// func (d Date) DaysSinceEpoch() (days PeriodOfDays) {
// 	return d.day
// }

// DaysIn days in a month
// func (d Date) DaysIn() int {
// 	if d.month == int(time.February) && d.IsLeap() {
// 		return 29
// 	}
// 	switch d.month {
// 	case 9:
// 		return 30
// 	case 4:
// 		return 30
// 	case 6:
// 		return 30
// 	case 11:
// 		return 30
// 	}

// 	return 31
// }

func isLeap(year int64) bool {
	year = astronomicalYear(year)
	if year == 0 {
		return false
	}

	isLeap := year%4 == 0 && (year%100 != 0 || year%400 == 0)

	return isLeap
}

// IsLeap simply tests whether a given year is a leap year, using the Gregorian calendar algorithm.
func (d Date) IsLeap() bool {
	// NOTES
	// -1 BCE is a leap year
	// -5 BCE is a leap year
	// -401 BCE is a leap year

	return isLeap(d.year)
}

// DaysIn gives the number of days in a given month, according to the Gregorian calendar.
// func DaysIn(year int64, month time.Month) int {
// 	if month == time.February && IsLeap(year) {
// 		return 29
// 	}

// 	return daysInMonth[month]
// }

// DatesInRange get dates in range.
// func DatesInRange(d1, d2 Date) ([]Date, error) {
// 	start := d1.UTC()
// 	end := d2.UTC()

// 	dateMap := make(map[Date]string)

// 	for rd := timestamp.RangeOverTimes(start, end); ; {
// 		t, err := rd()
// 		if err != nil {
// 			return nil, err
// 		}

// 		if t.IsZero() {
// 			break
// 		}
// 		dateMap[NewAt(t)] = ""
// 	}

// 	dates := make([]Date, 0, len(dateMap))

// 	// Build output array
// 	for k := range dateMap {
// 		dates = append(dates, k)
// 	}

// 	// Sort output by date
// 	sort.Slice(dates[:], func(i, j int) bool {
// 		return dates[i].Before(dates[j])
// 	})

// 	return dates, nil
// }
