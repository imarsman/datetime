package date2

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
type PeriodOfDays int64

// ZeroDays is the named zero value for PeriodOfDays.
const ZeroDays PeriodOfDays = 0

// A Date represents a date under the (proleptic) Gregorian calendar as
// used by ISO 8601. This calendar uses astronomical year numbering,
// so it includes a year 0 and represents earlier years as negative numbers
// (i.e. year 0 is 1 BC; year -1 is 2 BC, and so on).
//
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
type Date struct {
	year  int64
	month int64
	day   int64
	ce    bool
}

func (d Date) yearAbs() int64 {
	if d.year < 0 {
		return -d.year
	}
	return d.year
}

// Validate validates that the date is basically OK
func (d Date) Validate() error {
	if d.IsZero() {
		return errors.New("Zero month or day")
	}
	return nil
}

func (d *Date) clean() error {
	if d.year <= 0 {
		d.year--
	}
	return nil
}

// NewDate returns the Date value corresponding to the given year, month, and day.
//
// The month and day may be outside their usual ranges and will be normalized
// during the conversion.
func NewDate(year int64, month int64, day int64) (Date, error) {
	d := Date{}

	var ce bool
	d.year, ce = gregorianYear(year)
	d.month = month
	d.day = day
	d.ce = ce

	err := d.clean()
	if err != nil {
		return Date{}, err
	}
	err = d.Validate()
	if err != nil {
		return Date{}, err
	}

	return d, nil
}

// IsCE is the year CE
func (d Date) IsCE() (bool, error) {
	err := d.clean()
	if err != nil {
		return false, err
	}
	return d.year > 0, nil
}

// Year get year for date
func (d Date) Year() int64 {
	return d.year
}

// Month get month for year
func (d Date) Month() int64 {
	return d.month
}

// Day returns the day of the month specified by d.
// The first day of the month is 1.
func (d Date) Day() int64 {
	return d.day
}

// SubtractDays add days to a date
func (d Date) SubtractDays(subtract int64) (date Date, err error) {
	d2 := d
	fmt.Println("subtracing from", d.String(), subtract)
	newYear := false
	for {
		daysInMonth, err := d2.daysInMonth()
		if err != nil {
			return Date{}, err
		}
		// fmt.Println("date", d2.String(), d2.month == 1 && d2.day == 1)
		if d2.month == 1 && d2.day == 1 {
			fmt.Println("equal")
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

// AddDays add days to a date
func (d Date) AddDays(add int64) (date Date, err error) {
	d2 := d
	// newYear := false
	dNeutral := d2
	dNeutral.year = 2019
	daysInMonth, _ := dNeutral.daysInMonth()
	for add > 0 {
		// We will deal with leap days elsewhere
		if err != nil {
			return Date{}, err
		}
		// if d2.month >= 12 {
		// 	newYear = true
		// }
		if add >= daysInMonth && d2.day <= daysInMonth {
			if d2.month >= 12 {
				d2.year++
				d2.month = 1
				d2.day = 1
				// newYear = false
				add = add - daysInMonth
				dNeutral = d2
				dNeutral.year = 2019
				daysInMonth, _ = dNeutral.daysInMonth()
			} else {
				d2.month++

				if d2.month > 12 {
					fmt.Println("months exceeded")
				}
				daysTilEOM := daysInMonth - d2.day
				add = add - daysTilEOM

				dNeutral = d2
				dNeutral.year = 2019
				daysInMonth, _ = dNeutral.daysInMonth()

				if daysTilEOM == 0 {
					d2.day = 1
					add--
				} else {
					add = add - daysTilEOM
					d2.day = 1
					add--
				}
			}
		} else {
			if add+d2.day >= daysInMonth {
				add = add - d2.day
				d2.day = 1
				d2.month++
				// New year
				if d2.month > 12 {
					d2.month = 1
					d2.year++
					// newYear = true
				}

				continue
			}
			var newAdd int64
			var newDay int64
			if add < d2.day {
				newAdd = d2.day - add
				newDay = d2.day + add
				d2.day = newDay
				add = newAdd
			} else {
				newAdd = add - d2.day
				newDay = d2.day + add
				d2.day = newDay
				add = newAdd
			}

			break
		}
	}

	return d2, nil
}

// AddMonths add months to a date
// TODO: decide if it would be good to add days
func (d Date) AddMonths(add int64) (d2 Date, err error) {
	d2 = d
	dNeutral := d2
	dNeutral.year = 2019
	daysInMonth, _ := dNeutral.daysInMonth()
	daysInMonth -= d2.day
	for i := 0; int64(i) < add; i++ {
		d2, _ = d2.AddDays(daysInMonth)
	}

	return d2, nil
}

// AddParts add the number of months to a starting month
// function. The remainder can be used to extend calculations to
// time parts. The time package AddParts call does not mutate the reciever by
// acting on a pointer so this one does the same.
// This call can be expensive with large units to add.
func (d Date) AddParts(years, months, days int64) (Date, int64, error) {
	dFinal := d

	var remainder int64

	if years > 0 {
		dFinal, _ = dFinal.addYears(years)
	}
	if months > 0 {
		dFinal, _ = dFinal.AddMonths(months)
	}

	if days > 0 {
		dFinal, _ = dFinal.AddDays(days)
	}

	extraDays := (dFinal.year - d.year) * 365

	// Add in extra days due to leap years
	var startDateDays int64 = int64(daysSinceEpoch(d.year))
	if startDateDays < 0 {
		startDateDays = -startDateDays
	}
	var endDateDays int64 = int64(daysSinceEpoch(dFinal.year))
	if endDateDays < 0 {
		endDateDays = -endDateDays
	}
	// fmt.Println("start days", startDateDays, "enddays", endDateDays)

	// The difference between the number of days up to the 1 January of the
	// date started with and the number of days to the 1 January of the date we
	// ended with.
	extraDays = (startDateDays - endDateDays) - int64(extraDays)

	// fmt.Println("extra days", extraDays)
	dFinal, _ = dFinal.AddDays(extraDays)

	return dFinal, remainder, nil
}

// TODO: Decide whether to account for leap days
func (d Date) addYears(years int64) (Date, error) {
	d2 := d
	// fmt.Println("add years", years)
	d2.year += years

	// fmt.Println("years added", d2.String())

	// var oldDays int64 = int64(-daysSinceEpoch(d.year))
	// var newDays int64 = int64(-daysSinceEpoch(dFinal.year))
	// yearDays := (dFinal.year - d.year) * 365
	// var diff int64 = (oldDays - newDays) - int64(yearDays)

	// fmt.Println("old date", dFinal.String())
	// // dFinal, _ = dFinal.AddDays(diff)
	// fmt.Println("added date", dFinal.String())

	// for {
	// 	monthDays, _ := d3.daysInMonth()
	// 	if days+monthDays < days {
	// 		days = days + monthDays
	// 	} else {
	// 		diff := days - monthDays
	// 		days += diff
	// 		break
	// 	}
	// }

	return d2, nil
}

func (d Date) daysInMonth() (int64, error) {
	err := d.clean()
	if err != nil {
		return 0, err
	}

	year := d.yearAbs()

	days := 31
	// Faster than if statement
	switch d.month {
	case 9:
		// September
		days = 30
	case 4:
		// April
		days = 30
	case 6:
		// June
		days = 30
	case 11:
		// November
		days = 30
	case 2:
		// February
		isLeap := IsLeap(year)
		if isLeap == false {
			days = 28
		} else {
			days = 29
		}
	}

	return int64(days), nil
}

// YearDay returns the day of the year specified by d, in the range [1,365] for
// non-leap years, and [1,366] in leap years. The functionality should be the
// same as for the Go time.YearDay func.
func (d Date) YearDay() (int64, error) {
	err := d.clean()
	if err != nil {
		return 0, err
	}
	var days int64 = 0
	copy := d
	for i := 1; i < 13; i++ {
		copy.month = int64(i)
		if copy.month > d.month {
			break
		}
		if copy.month == d.month {
			days += d.day
			break
		}
		val, _ := copy.daysInMonth()
		days += val
	}

	return days, nil
}

// WeekDay the day of the week for date as specified by time.Weekday
// A Weekday specifies a day of the week (Sunday = 0, ...).
// The English name for time.Weekday can be obtained with time.Weekday.String()
func (d Date) WeekDay() (int, error) {
	err := d.clean()
	if err != nil {
		return 0, err
	}
	dow1Jan, _ := d.dayOfWeek1Jan()
	dow := dow1Jan

	doy, _ := d.YearDay()

	dow--
	for i := 0; i < int(doy); i++ {
		if dow == 7 {
			dow = 1
		} else {
			dow++
		}
	}

	return dow, nil
}

// daysSinceEpoch takes a year and returns the number of days from
// the absolute epoch to the start of that year.
// This is basically (year - zeroYear) * 365, but accounting for leap days.
// From Go time package
func daysSinceEpoch(year int64) uint64 {
	y := uint64(int64(year) - absoluteZeroYear)

	// Add in days from 400-year cycles.
	n := y / 400
	y -= 400 * n
	d := daysPer400Years * n

	// Add in 100-year cycles.
	n = y / 100
	y -= 100 * n
	d += daysPer100Years * n

	// Add in 4-year cycles.
	n = y / 4
	y -= 4 * n
	d += daysPer4Years * n

	// Add in non-leap years.
	n = y
	d += 365 * n

	return d
}

func (d Date) dayOfWeek1Jan() (int, error) {
	err := d.clean()
	if err != nil {
		return 0, err
	}
	// https://en.wikipedia.org/wiki/Determination_of_the_day_of_the_week
	// Formula by Gauss

	// Inputs
	// Year number A, month number M, day number D.
	// set C = A \ 100, Y = A % 100, and the value is
	// (1+5((Y−1)%4)+3(Y−1)+5(C%4))%7
	year := d.yearAbs()
	if year < 0 {
		year = -year
	}
	c := year / 100
	y := year % 100
	result := (1 + 5*((y-1)%4) + 3*(y-1) + 5*(c%4)) % 7

	return int(result), nil
}

func isoWeekOfYearForDate(doy int, dow time.Weekday) int {
	var woy int = (10 + doy - int(dow)) % 7
	return woy
}

// ISOWeeksInYear get number of ISO weeks in year
func (d Date) ISOWeeksInYear() int {
	return isoWeeksInYear(d.year)
}

func isoWeeksInYear(year int64) int {
	if year < 0 {
		year = -year
	}
	year = gregorian.AdjustYear(year)

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
	d, _ := NewDate(int64(t.Year()), int64(t.Month()), int64(t.Day()))

	return d
}

// Min returns the smallest representable date.
func Min() Date {
	gy, _ := gregorianYear(math.MinInt64)
	d, _ := NewDate(gy, 1, 1)

	return d
}

// Max returns the largest representable date.
func Max() Date {
	gy, _ := gregorianYear(math.MaxInt64)
	d, _ := NewDate(gy, 1, 1)

	return d
}

// Date returns the year, month, and day of d.
// The first day of the month is 1.
func (d Date) Date() (year int64, month int64, day int64) {
	return d.year, d.month, d.day
}

// LastDayOfMonth returns the last day of the month specified by d.
// The first day of the month is 1.
func (d Date) LastDayOfMonth() (int64, error) {
	dim, err := d.daysInMonth()
	if err != nil {
		return 0, err
	}
	return dim, nil
}

// // Month returns the month of the year specified by d.
// func (d Date) Month() time.Month {
// 	return d.month
// }

// Year returns the year specified by d.
// func (d Date) Year() int {
// 	t := decode(d.day)
// 	return t.Year()
// }

// Weekday returns the day of the week specified by d.
func (d Date) Weekday() time.Weekday {
	// Date zero, January 1, 1970, fell on a Thursday
	wdayZero := time.Thursday
	// Taking into account potential for overflow and negative offset
	return time.Weekday((int32(wdayZero) + int32(d.day)%7 + 7) % 7)
}

// ISOWeek returns the ISO 8601 year and week number in which d occurs.
// Week ranges from 1 to 53. Jan 01 to Jan 03 of year n might belong to
// week 52 or 53 of year n-1, and Dec 29 to Dec 31 might belong to week 1
// of year n+1.
// func (d Date) ISOWeek() (year, week int) {
// 	t := decode(d.day)
// 	return t.ISOWeek()
// }

// IsZero reports whether t represents the zero date.
func (d Date) IsZero() bool {
	return d.day == 0
}

// Equal reports whether d and u represent the same date.
func (d Date) Equal(u Date) bool {
	return d.day == u.day
}

// IsBefore reports whether the date d is before u.
func (d Date) IsBefore(u Date) bool {
	return d.day < u.day
}

// IsAfter reports whether the date d is after u.
func (d Date) IsAfter(u Date) bool {
	return d.day > u.day
}

// MinDate returns the earlier of two dates.
func (d Date) MinDate(u Date) Date {
	if d.day > u.day {
		return u
	}
	return d
}

// MaxDate returns the later of two dates.
func (d Date) MaxDate(u Date) Date {
	if d.day < u.day {
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

// IsLeap simply tests whether a given year is a leap year, using the Gregorian calendar algorithm.
func IsLeap(year int64) bool {
	return gregorian.IsLeap(year)
}

// DaysIn gives the number of days in a given month, according to the Gregorian calendar.
func DaysIn(year int64, month time.Month) int {
	return gregorian.DaysIn(year, month)
}

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
