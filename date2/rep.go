package date2

// This package represents date whose zero value is 1 CE. It stores years as
// int64 values and as such the minimum and maximum possible years are very
// large. The Golang time package is used both for representing dates and times
// and for doing timing operations. This package is oriented toward doing things
// with dates.

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/imarsman/datetime/gregorian"
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
const StartYear = epochYearGregorian - epochYearGregorian + 1

const absoluteMaxYear = math.MaxInt64
const absoluteZeroYear = math.MinInt64 + 1

func gregorianYear(inputYear int64) (year int64) {
	year = inputYear
	if year == 0 {
		year = 1
	} else if year == -1 {
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

	// days := 31
	// // Faster than if statement
	// switch d.month {
	// case 9:
	// 	// September
	// 	days = 30
	// case 4:
	// 	// April
	// 	days = 30
	// case 6:
	// 	// June
	// 	days = 30
	// case 11:
	// 	// November
	// 	days = 30
	// case 2:
	// 	// February
	// 	leapD, err := NewDate(year, 1, 1)
	// 	if err != nil {
	// 		return 0, err
	// 	}
	// 	isLeap := leapD.IsLeap()
	// 	if isLeap == false {
	// 		days = 28
	// 	} else {
	// 		days = 29
	// 	}
	// }

	// return days, nil
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

func yearsAndCentury(y int64) (int64, int) {
	negative := false
	if y < 0 {
		negative = true
		y = -y
	}
	var err error
	v := strconv.Itoa(int(y))
	var years int64 = 0
	var century int = 0
	if len(v) > 2 {
		century, err = strconv.Atoi(v[0:2])
		if err != nil {
			fmt.Println("Could not parse", v)
		}
		newVal, err := strconv.Atoi(v[len(v)-2:])
		if err != nil {
			fmt.Println("Could not parse", v)
		}
		if negative {
			years = -int64(newVal)
			century = -century
		} else {
			years = int64(newVal)
		}
	} else if len(v) == 2 {
		century = 0
		if err != nil {
			fmt.Println("Could not parse", v)
		}
		newVal, err := strconv.Atoi(v[len(v)-2:])
		if err != nil {
			fmt.Println("Could not parse", v)
		}
		if negative {
			years = -int64(newVal)
			century = -century
		} else {
			years = int64(newVal)
		}
	} else {
		century = 0
		newVal, err := strconv.Atoi(v[0:])
		if err != nil {
			fmt.Println("Could not parse", v)
		}
		if negative {
			years = -int64(newVal)
			century = -century
		} else {
			years = int64(newVal)
		}
	}
	return years, century
}

// https://en.wikipedia.org/wiki/Doomsday_rule
func anchorDay(y int64) int {
	years, century := yearsAndCentury(y)

	anchorDay := 0
	r := century % 4
	anchorDay = r
	switch r {
	case 0:
		anchorDay = 2
	case 1:
		anchorDay = 7
	case 2:
		anchorDay = 5
	case 3:
		anchorDay = 3
	}

	t := float64(years)
	if int(t)%2 != 0 {
		t += 11
	}
	t = t / 2
	if int(t)%2 != 0 {
		t += 11
	}
	t = 7 - float64((int(t) % 7))
	t = t + float64(anchorDay)
	if t > 7 {
		t = t - 7
	}

	return int(t)
}

var doomsdays map[int][]int = map[int][]int{
	1:  {3, 10, 17, 24, 21},
	2:  {4, 11, 18, 25},
	3:  {7, 14, 21, 28},
	4:  {4, 11, 18, 25},
	5:  {2, 9, 16, 23, 30},
	6:  {6, 13, 20, 27},
	7:  {4, 11, 18, 25},
	8:  {1, 8, 15, 22, 29},
	9:  {5, 12, 19, 26},
	10: {3, 10, 17, 24, 31},
	11: {7, 14, 21, 28},
	12: {5, 12, 19, 26},
}

func (d Date) closestDoomsDayProximity() (int, error) {
	isLeap := d.IsLeap()
	row := doomsdays[d.month]
	// targetDiff := 0
	var candidate int
	for i := 0; i < len(row); i++ {
		candidate = row[i]
		// diff := d.day - candidate
		diff := candidate - d.day
		// fmt.Println("candidate", candidate, candidate-d.day)
		if math.Abs(float64(diff)) <= 10 {
			// targetDiff = diff
			break
			// return candidate, nil
		}
	}
	// if targetDiff == 0 {
	// 	return 0, errors.New("Didn't find candidate")
	// }
	if isLeap {
		candidate = candidate + 1
		// 	targetDiff = targetDiff + 1
	}
	// fmt.Println("target day", targetDiff)

	// return targetDiff, nil
	return candidate, nil
}

// WeekDay the day of the week for date as specified by time.Weekday
// A Weekday specifies a day of the week (Sunday = 0, ...).
// The English name for time.Weekday can be obtained with time.Weekday.String()
func (d Date) WeekDay() (int, error) {
	err := d.validate()
	if err != nil {
		return 0, err
	}
	if d.year == 0 {
		return 0, errors.New("no year zero")
	}

	// Get first two digit integer value of year or 0
	twoDigitYear, _ := yearsAndCentury(d.year)
	a := math.Floor(float64(twoDigitYear) / 12)
	b := twoDigitYear % 12
	c := math.Floor(float64(b) / 4)

	twoDigitYarSum := int(a) + int(b) + int(c)
	// Get the anchor day for date year
	anchorDayForYear := anchorDay(d.year)
	// remainder := sum % 7

	// Find anchor day
	// if remainder > 7 {
	// 	remainder = remainder - 7
	// }

	anchorDayAdjustment := twoDigitYarSum % 7
	// fmt.Println("anchor date adjustment", anchorDayAdjustment)
	dayForDateAdjustment := anchorDayForYear + anchorDayAdjustment
	// if dayForDateAdjustment > 7 {
	// 	dayForDateAdjustment = dayForDateAdjustment - 7
	// }

	// d1Day, err := d.YearDay()
	// if err != nil {
	// 	return 0, err
	// }

	closestAnchorDayInMonth, err := d.closestDoomsDayProximity()
	// fmt.Println("closestDayInMonth", closestDayInMonth, "dayForDate", dayForDateAdjustment, "anchor day for year", anchorDayForYear)
	// var new = anchorDay

	var dayDiff int = d.day - closestAnchorDayInMonth
	if d.day > closestAnchorDayInMonth {
		dayDiff = closestAnchorDayInMonth - d.day
	}

	// targetWeekDay := ((closestAnchorDayInMonth + dayDiff) % 7) + 1
	// fmt.Println("target day", targetWeekDay)
	// fmt.Println("anchor day", anchorDayForYear, "closest anchor day", closestAnchorDayInMonth, "day", d.day, "diff", dayDiff, "targetWeekDay", targetWeekDay)
	var final int
	if dayDiff < 0 {
		fmt.Println("negative")
		if d.day > closestAnchorDayInMonth {
			fmt.Println("negative diff and day greater")
			// If day number is larger than the closest anchor day
			// Add to get to the day number
			// Negate dayDiff
			// e.g. 7 + -1 == 7 - 1
			final = int(math.Mod(float64(anchorDayForYear)+float64(-dayDiff), 7))
		} else {
			fmt.Println("negative diff and day less")
			// If day number is smaller than the closest anchor day
			// Subtract to get to the day number
			final = int(math.Mod(float64(anchorDayForYear)-float64(-dayDiff), 7))
			fmt.Printf("anchor day %d + daydiff %d\n", anchorDayForYear, dayDiff)
		}
	} else {
		fmt.Println("positive")
		// 7 -
		if d.day > closestAnchorDayInMonth {
			fmt.Println("postive diff and day greater")
			// If day number is larger than the closest anchor day
			// Add to get to the day number
			final = int(math.Mod(float64(anchorDayForYear)+float64(-dayDiff), 7))
		} else {
			fmt.Println("postive diff and day less")
			// If day number is smaller than the closest anchor day
			// Subtract to get to the day number
			final = int(math.Mod(float64(anchorDayForYear)-float64(-dayDiff), 7))
		}
	}
	fmt.Println("closestAnchorDay", closestAnchorDayInMonth, "day", d.day, "anchor day", anchorDayForYear, "diff", dayDiff, "final", final)
	// if math.Abs(float64(closestDayInMonth)) > 7 {
	// 	closestDayInMonth = closestDayInMonth % 7
	// }
	// dayForDate := dayForDateAdjustment
	// if dayForDate < 0 {
	// 	dayForDate--
	// }
	// if closestAnchorDayInMonth < 0 {
	// 	// fmt.Println("less than")
	// 	// fmt.Println("closestDayInMonth", closestDayInMonth, "dayForDate", dayForDate, "anchor day for year", anchorDayForYear, "target day", d.day)
	// 	closestAnchorDayInMonth = -closestAnchorDayInMonth
	// 	normalized := anchorDayForYear + closestAnchorDayInMonth
	// 	// fmt.Println("closestDayInMonth", closestDayInMonth, "dayForDate", dayForDate, "anchor day for year", anchorDayForYear, "target day", d.day)
	// 	// normalized = -normalized
	// 	// new = new + closest
	// 	if normalized%7 > 0 {
	// 		adjustment := normalized % 7
	// 		dayForDate = adjustment
	// 	}
	// 	// fmt.Println("closestDayInMonth", closestDayInMonth, "dayForDate", dayForDate, "anchor day for year", anchorDayForYear, "target day", d.day)
	// } else {
	// 	// fmt.Println("greater than")
	// 	// fmt.Println("closestDayInMonth", closestDayInMonth, "dayForDate", dayForDate, "anchor day for year", anchorDayForYear, "target day", d.day)
	// 	// dayForDate = dayForDate + closestDayInMonth
	// 	dayForDate = anchorDayForYear - closestAnchorDayInMonth
	// 	// if dayForDate > 7 {
	// 	// 	fmt.Println("dayfordate", dayForDate)
	// 	// 	// excess := dayForDate % 7
	// 	// 	// dayForDate = excess
	// 	// }
	// 	// if closestDayInMonth > d.day {
	// 	// 	dayForDate = dayForDate + closestDayInMonth
	// 	// } else {
	// 	// 	dayForDate = dayForDate - closestDayInMonth
	// 	// }
	// }
	// // fmt.Println("closestDayInMonth", closestDayInMonth, "dayForDate", dayForDate, "anchor day", anchorDayForYear)
	// // new = new + final
	// // if new > 7 {
	// // 	fmt.Println("new new", new)
	// // 	new = new % 7
	// // 	fmt.Println("new new", new)
	// // }
	// // new := closest - d.day
	// // d2, err := NewDate(d.year, d.month, closest)
	// // if err != nil {
	// // 	return 0, err
	// // }
	// // closestDay, err := d2.YearDay()
	// // if err != nil {
	// // 	return 0, err
	// // }
	// // diff := d1Day - closestDay

	// // fmt.Println("final", final, "diff", diff, "d1Day", d1Day, "closest", closestDay)
	// // newDay := final + diff
	// // fmt.Println("final", final, "new day", newDay, "diff", diff)
	// // fmt.Println("diff", diff, "new day", newDay, "remainder", remainder, "anchor", anchor, "d1YearDay", d1YearDay, "d2YearDay", d2YearDay)
	// // if newDay > 7 {
	// // 	newDay = newDay - 7
	// // }

	// // fmt.Println("distance between closest and date", diff, "date", d.String())

	return final, nil
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
	err := d.validate()
	if err != nil {
		return 0, err
	}
	// https://en.wikipedia.org/wiki/Determination_of_the_day_of_the_week
	// Formula by Gauss

	// Inputs
	// Year number A, month number M, day number D.
	// set C = A \ 100, Y = A % 100, and the value is
	// (1+5((Y−1)%4)+3(Y−1)+5(C%4))%7
	year := d.astronomicalYear()
	// year := d.year
	if d.year < 0 {
		// year--
		// year = year + 1
		// fmt.Println("year", year, "negative")
		// // year = -year
		// fmt.Println("year", year, "negative")
	}
	// if year < 0 {
	// 	year = -year
	// }
	// year := d.yearAbs()
	// if year < 0 {
	// 	year = -year
	// }
	c := year / 100
	// if c == 0 {
	// 	c = 1
	// }
	y := year % 100
	// if y == 0 {
	// 	y = 1
	// }
	result := (1 + 5*((y-1)%4) + 3*(y-1) + 5*(c%4)) % 7

	if result < 0 {
		result = -result
	}
	// fmt.Println("year", year, "y", y, "result", result, "c", c)

	return int(result), nil
}

func isoWeekOfYearForDate(doy int, dow time.Weekday) int {
	var woy int = (10 + doy - int(dow)) % 7
	return woy
}

// // encode returns the number of days elapsed from date zero to the date
// // corresponding to the given Time value.
// func encode(t time.Time) PeriodOfDays {
// 	// Compute the number of seconds elapsed since January 1, 1970 00:00:00
// 	// in the location specified by t and not necessarily UTC.
// 	// A Time value is represented internally as an offset from a UTC base
// 	// time; because we want to extract a date in the time zone specified
// 	// by t rather than in UTC, we need to compensate for the time zone
// 	// difference.
// 	_, offset := t.Zone()
// 	secs := t.Unix() + int64(offset)
// 	// Unfortunately operator / rounds towards 0, so negative values
// 	// must be handled differently
// 	if secs >= 0 {
// 		return PeriodOfDays(secs / secondsPerDay)
// 	}
// 	return -PeriodOfDays((secondsPerDay - 1 - secs) / secondsPerDay)
// }

// // decode returns the Time value corresponding to 00:00:00 UTC of the date
// // represented by d, the number of days elapsed since date zero.
// func decode(d PeriodOfDays) time.Time {
// 	secs := int64(d) * secondsPerDay
// 	return time.Unix(secs, 0).UTC()
// }
