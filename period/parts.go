package period

import (
	"errors"
	"math"
	"time"
)

// Period a struct to define a period
type Period struct {
	negative                                            bool
	years, months, weeks, days, hours, minutes, seconds int64
	subseconds                                          int
}

const daysPerYearE4 = 3652425 // 365.2425 days by the Gregorian rule
const daysPerMonthE4 = 304369 // 30.4369 days per month
// const daysPerMonthE6 = 30436875 // 30.436875 days per month
const hundredMSDuration = 100 * time.Millisecond

// https://en.wikipedia.org/wiki/Year
// An average Gregorian year is 365.2425 days (52.1775 weeks, 8765.82 hours,
// 525949.2 minutes or 31556952 seconds). For this calendar, a common year is
// 365 days (8760 hours, 525600 minutes or 31536000 seconds), and a leap year is
// 366 days (8784 hours, 527040 minutes or 31622400 seconds)

const daysPerMonthE6 = 30436875 // 30.436875 days per month

const oneDay time.Duration = 24 * time.Hour // Number of nanoseconds in a day

const oneMonthSeconds = 2628000                                    // Number of seconds in a month
const oneMonthApprox time.Duration = oneMonthSeconds * time.Second // 30.436875 days

const oneE4 = 10000 // 1e^4

const oneE5 = 100000 // 1e^5

const oneE6 = 1000000 // 1e^6

// More exact but rounds with small units
// const oneYearApprox = time.Duration(float64(365.2425*60*60*24)) * time.Second // 365.2425 days
const oneYearApprox time.Duration = oneMonthSeconds * time.Second * 12 // Nanoseconds in 1 year

// NewPeriod create a new Period instance
func NewPeriod(years, months, days, hours, minutes, seconds int64) Period {
	p := Period{}
	p.years = years
	p.months = months
	p.days = days
	p.hours = hours
	p.minutes = minutes
	p.seconds = seconds
	p.negative = years < 0 || months < 0 || days < 0 || hours < 0 || minutes < 0 || seconds < 0
	if p.negative {
		p.years = int64(math.Abs(float64(p.years)))
		p.months = int64(math.Abs(float64(p.months)))
		p.days = int64(math.Abs(float64(p.days)))
		p.hours = int64(math.Abs(float64(p.hours)))
		p.minutes = int64(math.Abs(float64(p.minutes)))
		p.seconds = int64(math.Abs(float64(p.seconds)))
	}

	return p
}

// NewYMD creates a simple period without any fractional parts. The fields are initialised verbatim
// without any normalisation; e.g. 12 months will not become 1 year. Use the Normalise method if you
// need to.
//
// All the parameters must have the same sign (otherwise a panic occurs).
// Because this implementation uses int16 internally, the paramters must
// be within the range ± 2^16 / 10.
func NewYMD(years, months, days int64) Period {
	return NewPeriod(years, months, days, 0, 0, 0)
}

// NewHMS creates a simple period without any fractional parts. The fields are initialised verbatim
// without any normalisation; e.g. 120 seconds will not become 2 minutes. Use the Normalise method
// if you need to.
//
// All the parameters must have the same sign (otherwise a panic occurs).
// Because this implementation uses int16 internally, the paramters must
// be within the range ± 2^16 / 10.
func NewHMS(hours, minutes, seconds int64) Period {
	return NewPeriod(0, 0, 0, hours, minutes, seconds)
}

// Years get years for period with proper sign
func (p Period) Years() int64 {
	if p.IsNegative() {
		return -p.years
	}
	return p.years
}

// Months get months for period with proper sign
func (p Period) Months() int64 {
	if p.IsNegative() {
		return -p.months
	}
	return p.months
}

// Days get days for period with proper sign
func (p Period) Days() int64 {
	if p.IsNegative() {
		return -p.days
	}
	return p.days
}

// Hours get hours for period with proper sign
func (p Period) Hours() int64 {
	if p.IsNegative() {
		return -p.hours
	}
	return p.hours
}

// Minutes get minutes for period with proper sign
func (p Period) Minutes() int64 {
	if p.IsNegative() {
		return -p.minutes
	}
	return p.minutes
}

// Seconds get seconds for period with proper sign
func (p Period) Seconds() int64 {
	if p.IsNegative() {
		return -p.seconds
	}
	return p.seconds
}

// IsNegative is period negative
func (p *Period) IsNegative() bool {
	return p.negative == true
}

// IsPositive returns true if its negative property is false
func (p *Period) IsPositive() bool {
	return p.negative == false
}

// Negate changes the sign of the period.
func (p *Period) Negate() *Period {
	if p.IsNegative() {
		p.negative = false
		return p
	}
	p.negative = true

	return p
}

// Duration converts a period to the equivalent duration in nanoseconds.
// A flag is also returned that is true when the conversion was precise and
// false otherwise.
//
// When the period specifies hours, minutes and seconds only, the result is
// precise. however, when the period specifies years, months and days, it is
// impossible to be precise because the result may depend on knowing date and
// timezone information, so the duration is estimated on the basis of a year
// being 365.2425 days as per Gregorian calendar rules) and a month being 1/12
// of a that; days are all assumed to be 24 hours long.
func (p Period) Duration() (time.Duration, bool, error) {
	// remember that the fields are all fixed-point 1E1
	tdE6, err := ymdApproxDuration(p)
	if err != nil {
		return time.Duration(0), false, err
	}
	stE3, err := hmsDuration(p)
	if err != nil {
		return time.Duration(0), false, err
	}

	if p.negative == true {
		return -(tdE6 + stE3), tdE6 == 0, nil
	}
	return (tdE6 + stE3), tdE6 == 0, nil
}

func hmsDuration(p Period) (time.Duration, error) {
	hourDuration := time.Duration(p.hours) * time.Hour
	minuteDuration := time.Duration(p.minutes) * time.Minute
	secondDuration := time.Duration(p.seconds) * time.Second
	hourminutesecondDuration := (hourDuration + minuteDuration + secondDuration)

	hourNumber := int64(hourminutesecondDuration / time.Hour)
	remainder := int64(hourminutesecondDuration % time.Hour)

	minuteNumber := remainder / int64(time.Minute)
	remainder = int64(hourminutesecondDuration % time.Minute)

	secondNumber := remainder / int64(time.Second)

	if hourNumber < 0 && minuteNumber < 0 && secondNumber < 0 {
		return time.Duration(0), errors.New("Hour, minute, and second duration exceeds maximum")
	}

	// remember that the fields are all fixed-point 1E1
	// and these are divided by 1E1
	// hhE3 := time.Duration(period.hours) * time.Hour
	// mmE3 := time.Duration(period.minutes) * time.Minute
	// ssE3 := time.Duration(period.seconds) * time.Second
	// return hhE3 + mmE3 + ssE3
	return hourminutesecondDuration, nil
}

func ymdApproxDuration(p Period) (time.Duration, error) {
	yearDuration := time.Duration(p.years) * oneYearApprox
	monthDuration := time.Duration(p.months) * oneMonthApprox
	dayDuration := time.Duration(p.days) * oneDay
	yearMonthDayDuration := yearDuration + monthDuration + dayDuration

	yearNumber := int64(yearMonthDayDuration / oneYearApprox)
	remainder := int64(yearMonthDayDuration % oneYearApprox)

	monthNumber := int64(remainder / int64(oneMonthApprox))
	remainder = int64(yearMonthDayDuration % oneMonthApprox)

	dayNumber := int64(remainder / int64(oneDay))

	if yearNumber < 0 && monthNumber < 0 && dayNumber < 0 {
		return time.Duration(0), errors.New("Year, month, and day duration exceeds maximum")
	}

	return yearMonthDayDuration, nil
}
