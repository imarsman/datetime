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
const StartYear = epochYearGregorian - epochYearGregorian

const absoluteMaxYear = math.MaxInt64
const absoluteZeroYear = math.MinInt64 + 1

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

func yearsAndCentury(year int64) (int64, int) {
	negative := false
	if year < 0 {
		negative = true
		year = -year
	}
	var err error
	v := strconv.Itoa(int(year))
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

	if year < 1000 && year > -1000 {
		year = 0
	}

	// fmt.Println("years", years, "century", century)
	return years, century
}

// https://dev.to/thormeier/algorithm-explained-the-doomsday-rule-or-figuring-out-if-november-24th-1763-was-a-tuesday-4lai
func anchorDayForCentury(y int64) int {
	yearPart := float64(y) / 100
	result := int((9 - (int64(math.Floor(yearPart))%4)*2) % 7)
	// if y < 0 && y > -100 {
	// 	result = 2
	// } else if y < 0 {
	// 	result--
	// 	if result == 0 {
	// 		result = 7
	// 	}
	// }
	fmt.Println("anchor day for century of ", y, "is", result)
	if result == 0 {
		result = 7
	}
	fmt.Println("anchor day for century of ", y, "is", result)

	return result
}

func (d Date) daysSince1Jan() int {
	d2 := d
	d2.day = 1
	d2.month = 1
	var startDateDays int64 = int64(daysTo1JanSinceEpoch(d2.year))
	if startDateDays < 0 {
		startDateDays = -startDateDays
	}

	// Get days since 1 Jan
	d3, _ := NewDate(d.year, 1, 1)
	total := 0
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

	return total
}

// https://dev.to/thormeier/algorithm-explained-the-doomsday-rule-or-figuring-out-if-november-24th-1763-was-a-tuesday-4lai
// https://en.wikipedia.org/wiki/Doomsday_rule
func anchorDayForYear(year int64) int {
	// years, _ := yearsAndCentury(y)

	// code := years / 4
	// code = years + code
	// code = code % 7

	// twoDigitYears := y % 100
	// fmt.Println("twodigityears", years)

	centuryAnchorDay := anchorDayForCentury(year)
	// fmt.Println("century anchor day", centuryAnchorDay)

	// yearPart := float64(y) / 100
	// centuryAnchorDay := int((9 - (int64(math.Floor(float64(y)/100))%4)*2) % 7)

	// c := int(math.Mod(float64(y), 100))
	// // fmt.Println("c", c)

	// var anchorDay int
	// // r := century % 4
	// r := int(math.Mod(float64(c), 4))
	// anchorDay = r
	// switch r {
	// case 0:
	// 	anchorDay = 2
	// case 1:
	// 	anchorDay = 7
	// case 2:
	// 	anchorDay = 5
	// case 3:
	// 	anchorDay = 3
	// }

	// // Get first two digit integer value of year or 0
	// // twoDigitYear, _ := yearsAndCentury(d.year)
	// a := math.Floor(float64(twoDigitYears) / 12)
	// b := twoDigitYears % 12
	// c2 := math.Floor(float64(b) / 4)

	// twoDigitYarSum := int(a) + int(b) + int(c2)

	// t := twoDigitYarSum % 7
	// fmt.Println("got", t)

	yy, century := yearsAndCentury(year)
	if year >= -100 && year <= 100 {
		fmt.Printf("year %d years %d century %d\n", year, yy, century)
		century = int(math.Abs(float64(century))) / 10
	}
	fmt.Printf("year %d years %d century %d\n", year, yy, century)

	// yy := y % 100 // Year, 1-2 digits

	anchorDay := (int(yy+int64(math.Floor(float64(yy)/4))) + centuryAnchorDay) % 7
	return anchorDay

	// t := int(twoDigitYears)
	// if int(t)%2 != 0 {
	// 	t += 11
	// }
	// t = t / 2
	// if int(t)%2 != 0 {
	// 	t += 11
	// }
	// t = 7 - (t % 7)
	// fmt.Println("t2", t)
	// t = t + centuryAnchorDay
	// if t > 7 {
	// 	t = t % 7
	// }
	// fmt.Println("t4", t)

	// return int(t)
}

var doomsdays map[int][]int = map[int][]int{
	int(January):   {3, 10, 17, 24, 31},
	int(February):  {7, 14, 21, 28},
	int(March):     {7, 14, 21, 28},
	int(April):     {4, 11, 18, 25},
	int(May):       {2, 9, 16, 23, 30},
	int(June):      {6, 13, 20, 27},
	int(July):      {4, 11, 18, 25},
	int(August):    {1, 8, 15, 22, 29},
	int(September): {5, 12, 19, 26},
	int(October):   {3, 10, 17, 24, 31},
	int(November):  {7, 14, 21, 28},
	int(December):  {5, 12, 19, 26},
}

var doomsdaysLeap map[int][]int = map[int][]int{
	int(January):  {4, 11, 18, 25},
	int(February): {1, 8, 15, 22, 29},
}

func (d Date) nearestDoomsday() int {
	isLeapYear := d.IsLeap()
	switch d.month {
	case 1:
		if !isLeapYear {
			return 4
		}
		return 3
	case 2:
		if !isLeapYear {
			return 29
		}
		return 28
	case 3:
		return 0
	case 4:
		return 4
	case 5:
		return 9
	case 6:
		return 6
	case 7:
		return 11
	case 8:
		return 8
	case 9:
		return 5
	case 10:
		return 10
	case 11:
		return 7
	case 12:
		return 12
	}

	// return [
	//     1 => !$isLeapYear ? 3 : 4,
	//     2 => !$isLeapYear ? 28 : 29,
	//     3 => 0,
	//     4 => 4,
	//     5 => 9,
	//     6 => 6,
	//     7 => 11,
	//     8 => 8,
	//     9 => 5,
	//     10 => 10,
	//     11 => 7,
	//     12 => 12,
	// ][$m];
	return 0
}

func (d Date) closestDoomsdayProximity() (int, error) {
	isLeap := d.IsLeap()
	// fmt.Println("year", d.year, "is leap", isLeap)
	row := doomsdays[d.month]
	if isLeap && (d.month == int(January) || d.month == int(February)) {
		row = doomsdaysLeap[d.month]
	}
	var candidate int
	for i := 0; i < len(row); i++ {
		candidate = row[i]
		diff := candidate - d.day
		if math.Abs(float64(diff)) <= 10 {
			break
		}
	}
	// // if targetDiff == 0 {
	// // 	return 0, errors.New("Didn't find candidate")
	// // }
	// if isLeap && d.month == 2 {
	// 	candidate = candidate + 1
	// 	// 	targetDiff = targetDiff + 1
	// }
	fmt.Println("closestDoomsdayProximity", candidate, "to", d.String())

	// return targetDiff, nil
	return candidate, nil
}

// daysTo1JanSinceEpoch takes a year and returns the number of days from
// the absolute epoch to the start of that year.
// This will work for CE but not for BCE
func daysTo1JanSinceEpoch(year int64) uint64 {
	var leapYearCount int64

	// - add all years divisible by 4
	// - subtract all years divisible by 100
	// - add back all years divisible by 400
	leapYearCount = (year / 4) - (year / 100) + (year / 400)

	total := year * 365
	total -= 365
	total += leapYearCount

	// d, _ := NewDate(year, 1, 1)
	// if d.IsLeap() {
	// 	total--
	// }

	if isLeap(year) {
		total--
	}

	return uint64(total)
}

// Weekday get day of week for a date
func (d Date) Weekday() (int, error) {
	err := d.validate()
	if err != nil {
		return 0, err
	}
	// https://en.wikipedia.org/wiki/Determination_of_the_day_of_the_week
	// Formula by Gauss

	year := d.astronomicalYear()
	if d.year < 0 {
		return 0, errors.New("Pre CE years not handled yet")
	}

	// var anchorDay int64 = 6
	// rollover := 7 - anchorDay
	dAnchorDays := int64(daysTo1JanSinceEpoch(1))

	dateDays := int64(daysTo1JanSinceEpoch(year))

	daysSince1Jan := d.daysSince1Jan()
	// fmt.Println("days since 1 Jan", daysSince1Jan)
	// Likely to account for the fact that the first year of 1-1 CE acts like a
	// Monday. This needs to be understood better.
	daysSince1Jan++

	// var diff int64
	// fmt.Println("anchorDay", anchorDay, "dateDays", dateDays, "anchorDays", dAnchorDays, "daysSince1Jan",
	// 	daysSince1Jan, "diff", diff)

	// Currently anchor days will be zero
	totalDays := dateDays + dAnchorDays
	// totalDays := dateDays
	// fmt.Println("total days", totalDays, "days since 1 Jan", daysSince1Jan)

	// Add in number of days into year
	totalDays += int64(daysSince1Jan)
	// fmt.Println("total days", totalDays)

	// The 1 Jan value will be one too high as the leap day for that year will
	// be factored in but we also factor in the leap day with daysSince1Jan
	dow := totalDays
	dow = dow % 7
	if year <= 100 {
		dow = dow + 5 // Add from Monday to Friday (1-1-1 is a Friday)
	}
	if d.IsLeap() {
		dow++
	}
	// if final == 0 {
	// 	fmt.Println("adding")
	// 	final = 7
	// }

	// fmt.Println("final", dow)

	return int(dow), nil
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
