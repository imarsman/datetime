package date

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/imarsman/datetime/gregorian"
)

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
// require this check will return an error if validation fails.
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

// validate validate a date to ensure it is a proper date.
func (d *Date) validate() error {
	if d.IsZero() || d.safelyInstantiated == false {
		return errors.New("date invalidly instantiated with new(Date) or Date{}")
	}
	if d.month < 1 || d.month > 12 {
		return fmt.Errorf("invalid month %d for date", d.month)
	}
	daysInMonth := d.DaysInMonth()
	// Ensure valid day number
	if d.day > daysInMonth {
		return fmt.Errorf("invalid day %d for month %d", d.day, d.month)
	}

	return nil
}

// yearAbs get positive value for year.
func (d Date) yearAbs() int64 {
	if d.year < 0 {
		return -d.year
	}
	return d.year
}

// AstronomicalYear get the year, which for ce years starts at 0 instead of -1.
// This means that 1 BCE is astronomical year 0.
func (d Date) AstronomicalYear() int64 {
	return astronomicalYear(d.year)
}

// IsCE is the year CE
func (d Date) IsCE() bool {
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

// DaysInMonth the number of days for a month taking the year into consideration.
func (d Date) DaysInMonth() int {
	// validate calls this so there should be a valid month and year
	daysInMonth := gregorian.DaysInMonth[d.month]
	year := d.AstronomicalYear()
	leapD := d
	leapD.year = year
	isLeap := leapD.IsLeap()
	if isLeap && d.month == 2 {
		return 29
	}

	return daysInMonth
}

// AtYearDay get date at day of year
func (d Date) AtYearDay(day int) Date {
	var days int = 0
	copy := d
	copy.year = copy.AstronomicalYear()
	for i := 1; i < 13; i++ {
		copy.month = i
		if copy.month > d.month {
			break
		}
		if copy.month == d.month {
			days += int(d.day)
			break
		}
		val := copy.DaysInMonth()
		days += int(val)
	}

	return Date{}
}

// YearDay returns the day of the year specified by d, in the range [1,365] for
// non-leap years, and [1,366] in leap years. The functionality should be the
// same as for the Go time.YearDay func.
func (d Date) YearDay() int {
	var days int = 0
	copy := d
	copy.year = copy.AstronomicalYear()
	for i := 1; i < 13; i++ {
		copy.month = i
		if copy.month > d.month {
			break
		}
		if copy.month == d.month {
			days += d.day
			break
		}
		val := copy.DaysInMonth()
		days += val
	}

	return days
}

// AddDays add days to a date
//
// Using chunks of 400 years allows the avoidance of huge cyles of a year at at
// time for very large values (millions of years and more).
// The deduction in days for 400 year chunks is not significantly different for
// different numbers of chunks.
//
// Most of the difference in time is made up in cycling though remainder days
// after accounting for the 400 year chunks. The goal is to cap the maximum
// amount of work to be done.
//
// https://www.thecalculatorsite.com/time/days-between-dates.php
//
// one chunk of 400 years is 146097 days
//
// - 800000-12-31 - remainder of 0 days
//   15.78 ns/op   633.84 MB/s   0 B/op   0 allocs/op
// - 2001-12-31 - remainder of 364 days
//   18.47 ns/op   541.31 MB/s   0 B/op   0 allocs/op
// - 800-12-01 - remainder of 36494 days
//   926.2 ns/op    10.80 MB/s   0 B/op   0 allocs/op
func (d Date) AddDays(days int64) (Date, error) {
	copy := d
	// fmt.Println("days", days)
	// startingDays := days

	/*
		How chunkDays is calculated

		var startingYears int64 = 400
		var ay int64 = startingYears // Make sure to use astromical years
		var leapDayCount int64 = (ay / 4) - (ay / 100) + (ay / 400)
		var chunkDays int64 = 365*400 + leapDayCount
	*/

	// A chunk of 400 years of days, including leap days
	// The Gregorian calendar has cyles of 400 years. Every four years there is
	// a leap day, minus one leap day on years divisible by 100 and adding back
	// a leap day for years divisible by 400.
	const chunk400YearDays int64 = 146097

	const chunk100YearDays int64 = 365*100 + 24

	var chunks400Years bool

	chunks400 := days / chunk400YearDays
	chunk400YearRemainder := days % chunk400YearDays

	chunks100 := chunk400YearRemainder / chunk100YearDays
	chunk100YearRemainder := chunk400YearRemainder / chunk100YearDays

	if d.year < 0 {
		copy.year = d.year
	}
	if copy.year < 0 {
		days -= 2
	}

	var totalCunkYears int64 = 1

	if chunks400 > 0 {
		chunks400Years = true
		chunkTotal := chunks400 * chunk400YearDays

		if copy.year < 0 {
			copy.year -= chunks400 * 400
			days -= chunkTotal
		} else {
			copy.year += chunks400 * 400
			days -= chunkTotal
		}
		totalCunkYears += chunks400 * 400
		// Introduce a leap year compensation that we will have to deal with
		// more later.
		if copy.IsLeap() && chunk400YearRemainder > 0 {
			days++
		}
	}
	if chunk400YearRemainder == 0 {
		copy.month--
	}

	days = chunk400YearRemainder + chunk100YearRemainder

	if chunks100 > 0 {
		if chunks400Years == true && copy.year < 0 {
			days--
		}

		chunkTotal := chunks100 * chunk100YearDays

		if copy.year < 0 {
			copy.year -= chunks100 * 100
			days -= chunkTotal
		} else {
			copy.year += chunks100 * 100
			days -= chunkTotal + (chunks100 * 2)
		}
	}

	days = days + chunk100YearRemainder

	var err error
	if days > 0 {
		if copy.year < 0 {
			days++
		}
		if d.year > 0 {
			// days--
			daysInMonth := copy.DaysInMonth()
			for days > 0 {
				daysInMonth = copy.DaysInMonth()
				if int64(daysInMonth) <= days {
					copy.month++
					if copy.month > 12 {
						copy.month = 1
						copy.day = 1
						copy.year++
						daysInMonth = copy.DaysInMonth()
						days -= int64(daysInMonth)

						continue
					}
					days -= int64(daysInMonth)
					copy.day = 1
				} else {
					copy.day += int(days)
					days = 0
					break
				}
			}
		} else {
			daysInMonth := copy.DaysInMonth()
			for {
				if int64(daysInMonth) <= days {
					days -= int64(daysInMonth)
					copy.month--
					daysInMonth = copy.DaysInMonth()
					copy.day = daysInMonth
					if copy.month < 1 {
						copy.month = 12
						copy.day = 31
						copy.year--
						daysInMonth = copy.DaysInMonth()
						days -= int64(daysInMonth)

						continue
					}
					days -= int64(daysInMonth)
					copy.day = 1
				} else {
					copy.day -= int(days)
					if copy.day == 0 {
						copy.day++
					}
					break
				}
			}
			if copy.IsLeap() && copy.day > 1 {
				copy.day--
			}
		}
	} else {
		// Year day counts are of the total days in a year, and do not start at
		// day one. Thus, with no remainder after 400 and 100 year chunks have
		// been removed, there will be one day to add.
		// e.g. 900-12-31
		copy.month++
		if copy.month > 12 {
			copy.year++
			copy.month = 1
			copy.day = 1
		} else {
			copy.day = 1
		}
	}

	// Have not discovered the reson for consistent errors. Most likely to do
	// with chunks of 400 and 100 years.
	if copy.year > 0 {
		for {
			if (copy.year-1)%400 == 0 && copy.month == 2 && copy.day == 1 {
				copy.year--
				copy.month = 12
				copy.day = 31
				break
			} else if (copy.year-1)%100 == 0 && copy.month == 2 && copy.day == 1 {
				copy.month = 1
				copy.day = 1
				break
			}
			break
		}
	}

	err = copy.validate()
	if err != nil {
		return Date{}, err
	}

	return copy, nil
}

// FromDays get date from days from epoch
func FromDays(days int64, ce bool) (Date, error) {
	d, err := NewDate(1, 1, 1)
	if err != nil {
		return Date{}, err
	}
	if ce != true {
		d.year = -d.year
		d.month = 12
		d.day = 31
	}

	d, err = d.AddDays(days)
	if err != nil {
		return Date{}, err
	}

	return d, nil
}

func (d Date) daysToDateFromEpoch() int64 {
	// To 1 Jan for CE or 31 Dec for BCE
	daysToAnchorDate := daysToAnchorDayFromEpoch(d.year)
	// fmt.Println("days to anchor day", daysToAnchorDate)

	daysToDate := d.daysToDateFromAnchorDay()
	// fmt.Println("days to date", daysToDate, d)
	totalDays := daysToAnchorDate + int64(daysToDate)

	return totalDays
}

// daysOneYearFromDate add a year and if the intervening days have a leap year days
// return 366, else return 365.
// This is much cheaper than calculating days since epoch to date.
func (d Date) daysOneYearFromDate() int {
	copy := d
	if copy.year > 0 {
		isLeap := copy.IsLeap()
		if isLeap == false {
			copy.year++
			isLeap = copy.IsLeap()
			if isLeap {
				if copy.month > 2 {
					return 365
				}
				if copy.month == 2 && copy.day == 29 {
					return 366
				}
			}
		} else {
			if copy.month < 2 {
				return 366
			}
			if copy.month == 2 && copy.day < 29 {
				return 366
			}
		}
	} else {
		isLeap := copy.IsLeap()
		copy.year--
		if !isLeap {
			isLeap = copy.IsLeap()
			if isLeap {
				if copy.month < 2 {
					return 365
				}
				if copy.month > 2 {
					return 366
				}
				if copy.month == 2 && copy.day == 29 {
					return 366
				}
			}
		} else {
			if copy.month < 2 {
				return 366
			}
			if copy.month > 2 {
				return 365
			}
			if copy.month == 2 && copy.day < 29 {
				return 366
			}
		}
	}

	return 365
}

// AddParts add in succession years, months, and days to a date
// - add years and then increment the date using the remainder days
// - add sum of days for months to date
// - add days days to date
func (d Date) AddParts(years int64, months, days int) (copy Date, remainder int64, err error) {
	copy = d

	// Add years. Surplus days are added to result
	if years > 0 {
		copy, err = copy.AddYears(years)
	}
	// Month handling adding surplus days to result
	if months > 0 {
		copy, err = copy.AddMonths(months)
		if err != nil {
			return Date{}, 0, err
		}
	}

	// Add days
	if days > 0 {
		copy, err = copy.AddDays(int64(days))
	}

	return copy, remainder, nil
}

// AddYears add years to date
func (d Date) AddYears(years int64) (copy Date, err error) {
	copy = d

	var totalDays int64
	if d.year < 0 {
		// TODO: test this
		if d.year+years > 0 {
			startDays := d.daysToDateFromEpoch()

			dEnd := d
			// The year will cross the zero boundary
			dEnd.year = d.year + years + 1 // Crossing CE boundary
			endDays := dEnd.daysToDateFromEpoch()

			totalDays = endDays + startDays
			newYears := totalDays / 365
			copy.year += newYears + 1 // Crossing CE boundary
			remainder := totalDays % 365
			copy, err = copy.AddDays(remainder)
		} else {
			// Get years with end in CE

			startDays := d.daysToDateFromEpoch()

			dEnd := d
			// The year will cross the zero boundary
			dEnd.year = d.year - years
			endDays := dEnd.daysToDateFromEpoch()

			totalDays = endDays - startDays
			newYears := totalDays / 365
			copy.year += newYears
			remainder := totalDays % 365
			copy, err = copy.AddDays(remainder)
		}
	} else {
		startDays := d.daysToDateFromEpoch()

		dEnd := d
		dEnd.year = d.year + years
		endDays := dEnd.daysToDateFromEpoch()

		totalDays = endDays - startDays
		newYears := totalDays / 365
		copy.year += newYears
		remainder := totalDays % 365
		copy, err = copy.AddDays(remainder)
	}

	return copy, nil
}

// AddMonths add months to a date
// TODO: decide if it would be good to add days
func (d Date) AddMonths(add int) (copy Date, err error) {
	copy = d

	// Benchmark adding 1000 months to 2019-01-01
	// With adding years:    122.3 ns/op	  81.76 MB/s   0 B/op   0 allocs/op
	// Without adding year:  284.4 ns/op	  35.16 MB/s   0 B/op   0 allocs/op
	// With and without the years add produces same result
	if add > 12 {
		for {
			// Asking for year days using IsLeap logic is much much faster than
			// counting days from epoch to date.
			daysBetween := copy.daysOneYearFromDate()
			if daysBetween < 0 {
				break
			}
			if add >= 12 {
				copy, err = copy.AddYears(1)
				if err != nil {
					return Date{}, err
				}
				// Subtract 12 months from the number of months to add
				add -= 12
				if add < 12 {
					break
				}
			} else {
				break
			}
		}
	}

	// TODO: add years if there are multiples of 12 months
	// daysToAdd := 0
	for i := 0; i < add; i++ {
		if copy.month == 12 {
			copy.year++
			copy.month = 1
			copy.day = 1
		} else {
			if copy.month == 2 {
				daysInMonth := copy.DaysInMonth()
				if daysInMonth == 29 {
					copy, err = copy.AddDays(1)
					if err != nil {
						return Date{}, err
					}
				}
			}
			copy.month++
			copy.day = 1
		}
	}

	// Do validation of result
	err = copy.validate()
	if err != nil {
		return Date{}, err
	}

	return copy, nil
}

// DaysTo get the number of days to a date
func (d Date) DaysTo(d2 Date) int64 {
	days1 := d.daysToDateFromEpoch()
	days2 := d2.daysToDateFromEpoch()

	return int64(days2 - days1)
}

// AtOrPastLeapDay is the date at or past leap day
// Written assuming that counting days is backwards for BCE.
func (d Date) AtOrPastLeapDay() bool {
	if d.IsLeap() {
		if d.year > 0 {
			if d.month == 1 {
				return false
			}
			if d.month == 2 {
				if d.day < 29 {
					return false
				}
			}
			return true
		}
		if d.month == 1 {
			return false
		}
		if d.month == 2 {
			if d.day < 29 {
				return false
			}
		}
		return true

	}
	return false
}

// For CE count forward from 1 Jan and for BCE count backward from 31 Dec
func (d Date) daysToDateFromAnchorDay() int {
	days := 0

	// Count forwards to date from 1 Jan
	d2 := d
	d2.month = 1
	d2.day = 1
	for {
		daysInMonth := d2.DaysInMonth()
		// If months is February on leap year we get the proper number of days
		if d2.month == d.month {
			if days == 0 {
				return d.day - 1
			}
			days += d.day - 1
			if days == -1 {
				days = 1
			}

			break
		}
		days += daysInMonth
		d2.month++
		daysInMonth = d2.DaysInMonth()
		if d2.month > 12 {
			break
		}
	}

	// TODO: double check using test
	// CE years
	if d.year > 0 {
		if d2.IsLeap() {
			// Remove extra day for days in Feb before the 29th CE leap yearr
			if d.month == 1 {
				days--
			}
			if d.month == 2 && d.day < 29 {
				days--
			}
		}
		// BCE years
	} else {
		days = 365 - days
		if d2.IsLeap() {
			// Remove extra day if previous to February 29s
			if d.month > 2 {
				days--
			}
		}
	}

	return days
}

// Weekday get day of week for a date for the proleptic Gregorian calendar. This
// means that gregorian dates will be returned even for years outside of the use
// of the Gregorian calendar.
// See https://www.thecalculatorsite.com/time/days-between-dates.php
// - seems to calculate according to proleptic Gregorian calendar
func (d Date) Weekday() int {
	// This will count forward for CE and backward for BCE but give the results
	// in both cases as a positive integer. The anchor date for CE is 1-1-1. The
	// anchor date for BCE is -1-12-31. Thus the days to anchor date will differ
	// between CE and BCE as CE counts forward through the year and BCE
	// effectively counts from 1 Jan as well but has to subtract from the number
	// of days in the year.

	// Use function that adds up the days to the date
	totalDays := d.daysToDateFromEpoch()

	var dow int
	if d.year > 0 {
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

// ISOWeekOfYear get ISO week of year. Not working properly currently.
func (d Date) ISOWeekOfYear() int {
	doy := d.daysToDateFromAnchorDay()
	d2 := d
	d2.month = 1
	d2.day = 1
	wd := d.Weekday()
	fmt.Println("doy", doy, "wd", wd)

	return isoWeekOfYearForDate(doy, time.Weekday(wd))
}

// ISOWeeksInYear get number of ISO weeks in year
func (d Date) ISOWeeksInYear() int {
	year := d.year
	if year < 0 {
		year = -year
	}

	dMathematical, _ := NewDate(year, 1, 1)
	year = dMathematical.AstronomicalYear()

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
	dim := d.DaysInMonth()

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
	// fmt.Println(year, isLeap)

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
