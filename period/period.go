package period

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/imarsman/datetime/xfmt"
)

// Period a struct to define a period
type Period struct {
	negative                                            bool
	years, months, weeks, days, hours, minutes, seconds int64
}

const yearChar = 'Y'
const monthChar = 'M'
const weekChar = 'W'
const dayChar = 'D'
const hourChar = 'H'
const minuteChar = 'M'
const secondChar = 'S'
const periodChar = 'P'
const timeChar = 'T'
const negativeChar = '-'

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

// Between converts the span between two times to a period. Based on the Gregorian conversion
// algorithms of `time.Time`, the resultant period is precise.
//
// To improve precision, result is not always fully normalised; for time differences less than 3276 hours
// (about 4.5 months), it will contain zero in the years, months and days fields but the number of hours
// may be up to 3275; this reduces errors arising from the variable lengths of months. For larger time
// differences (greater than 3276 hours) the days, months and years fields are used as well.
//
// Remember that the resultant period does not retain any knowledge of the calendar, so any subsequent
// computations applied to the period can only be precise if they concern either the date (year, month,
// day) part, or the clock (hour, minute, second) part, but not both.
func Between(t1, t2 time.Time) (p Period) {
	sign := 1
	if t2.Before(t1) {
		t1, t2, sign = t2, t1, -1
	}

	if t1.Location() != t2.Location() {
		t2 = t2.In(t1.Location())
	}

	year, month, day, hour, min, sec, hundredth := daysDiff(t1, t2)

	if sign < 0 {
		p = NewPeriod(-year, -month, -day, -hour, -min, -sec)
		p.seconds -= int64(hundredth)
	} else {
		p = NewPeriod(year, month, day, hour, min, sec)
		p.seconds += int64(hundredth)
	}
	return
}

func daysDiff(t1, t2 time.Time) (year, month, day, hour, min, sec, hundredth int64) {
	duration := t2.Sub(t1)

	hh1, mm1, ss1 := t1.Clock()
	hh2, mm2, ss2 := t2.Clock()

	day = int64(duration / (24 * time.Hour))

	hour = int64(hh2 - hh1)
	min = int64(mm2 - mm1)
	sec = int64(ss2 - ss1)
	hundredth = int64(t2.Nanosecond()-t1.Nanosecond()) / 100000000

	// Normalize negative values
	if sec < 0 {
		sec += 60
		min--
	}

	if min < 0 {
		min += 60
		hour--
	}

	if hour < 0 {
		hour += 24
		// no need to reduce day - it's calculated differently.
	}

	// test 16bit storage limit (with 1 fixed decimal place)
	if day > 3276 {
		y1, m1, d1 := t1.Date()
		y2, m2, d2 := t2.Date()
		year = int64(y2 - y1)
		month = int64(m2 - m1)
		day = int64(d2 - d1)
	}

	return
}

// https://en.wikipedia.org/wiki/Year
// An average Gregorian year is 365.2425 days (52.1775 weeks, 8765.82 hours,
// 525949.2 minutes or 31556952 seconds). For this calendar, a common year is
// 365 days (8760 hours, 525600 minutes or 31536000 seconds), and a leap year is
// 366 days (8784 hours, 527040 minutes or 31622400 seconds)
// const daysPerYearE4 = 3652425   // 365.2425 days by the Gregorian rule
// const daysPerMonth = 30436875   // 30.436875 days per month
// const daysPerMonthE4 = 304369   // 30.4369 days per month
// const daysPerMonthE6 = 30436875 // 30.436875 days per month
// const daysPerMonthE6 = 30436875 // 30.436875 days per month
const daysPerMonthE6 = 30436875 // 30.436875 days per month

var oneDay = 24 * time.Hour

const oneMonthSeconds = 2628000
const oneMonthApprox = oneMonthSeconds * time.Second // 30.436875 days

const oneE4 = 10000 // 1e^5

const oneE5 = 100000 // 1e^5

const oneE6 = 1000000 // 1e^6

// More exact but rounds with small units
// const oneYearApprox = time.Duration(float64(365.2425*60*60*24)) * time.Second // 365.2425 days
const oneYearApprox = oneMonthSeconds * time.Second * 12 // 365.2425 days

// Abs converts a negative period to a positive one.
func (p Period) Abs() Period {
	a, _ := p.absNeg()
	return a
}

func (p Period) absNeg() (Period, bool) {
	if p.IsNegative() {
		p.negative = false
		return p, true
	}
	return p, false
}

// Simplify applies some heuristic simplifications with the objective of reducing the number
// of non-zero fields and thus making the rendered form simpler. It should be applied to
// a normalised period, otherwise the results may be unpredictable.
//
// Note that months and days are never combined, due to the variability of month lengths.
// Days and hours are only combined when imprecise behaviour is selected; this is due to
// daylight savings transitions, during which there are more than or fewer than 24 hours
// per day.
//
// The following transformation rules are applied in order:
//
// * P1YnM becomes 12+n months for 0 < n <= 6
// * P1DTnH becomes 24+n hours for 0 < n <= 6 (unless precise is true)
// * PT1HnM becomes 60+n minutes for 0 < n <= 10
// * PT1MnS becomes 60+n seconds for 0 < n <= 10
//
// At each step, if a fraction exists and would affect the calculation, the transformations
// stop. Also, when not precise,
//
// * for periods of at least ten years, month proper fractions are discarded
// * for periods of at least a year, day proper fractions are discarded
// * for periods of at least a month, hour proper fractions are discarded
// * for periods of at least a day, minute proper fractions are discarded
// * for periods of at least an hour, second proper fractions are discarded
//
// The thresholds can be set using the varargs th parameter. By default, the thresholds a,
// b, c, d are 6 months, 6 hours, 10 minutes, 10 seconds respectively as listed in the rules
// above.
//
// * No thresholds is equivalent to 6, 6, 10, 10.
// * A single threshold a is equivalent to a, a, a, a.
// * Two thresholds a, b are equivalent to a, a, b, b.
// * Three thresholds a, b, c are equivalent to a, b, c, c.
// * Four thresholds a, b, c, d are used as provided.
//
func (p Period) Simplify(precise bool, th ...int) Period {
	switch len(th) {
	case 0:
		return p.doSimplify(precise, 60, 60, 100, 100)
	case 1:
		return p.doSimplify(precise, int64(th[0]*10), int64(th[0]*10), int64(th[0]*10), int64(th[0]*10))
	case 2:
		return p.doSimplify(precise, int64(th[0]*10), int64(th[0]*10), int64(th[1]*10), int64(th[1]*10))
	case 3:
		return p.doSimplify(precise, int64(th[0]*10), int64(th[1]*10), int64(th[2]*10), int64(th[2]*10))
	default:
		return p.doSimplify(precise, int64(th[0]*10), int64(th[1]*10), int64(th[2]*10), int64(th[3]*10))
	}
}

func (p Period) doSimplify(precise bool, monthMax, hourMax, minuteMax, secondMax int64) Period {
	if p.years%10 != 0 {
		return p
	}

	ap, neg := p.absNeg()

	// single year is dropped if there are some months
	if ap.years == 10 && 0 < ap.months && ap.months <= monthMax && ap.days == 0 {
		ap.months += 120
		ap.years = 0
	}

	if ap.months%10 != 0 && ap.months > 10 {
		// month fraction is dropped for periods of at least ten years (1:120)
		months := ap.months / 10
		if !precise && ap.years >= 100 && months == 0 {
			ap.months = 0
		}
		return ap.condNegate(neg)
	}

	if ap.days%10 != 0 && ap.days > 10 {
		// day fraction is dropped for periods of at least a year (1:365)
		days := ap.days / 10
		if !precise && (ap.years > 0 || ap.months >= 120) && days == 0 {
			ap.days = 0
		}
		return ap.condNegate(neg)
	}

	if !precise && ap.days == 10 && ap.years == 0 && ap.months == 0 && 0 < ap.hours && ap.hours <= hourMax {
		ap.hours += 240
		ap.days = 0
	}

	if ap.hours%10 != 0 && ap.hours > 10 {
		// hour fraction is dropped for periods of at least a month (1:720)
		hours := ap.hours / 10
		if !precise && (ap.years > 0 || ap.months > 0 || ap.days >= 300) && hours == 0 {
			ap.hours = 0
		}
		return ap.condNegate(neg)
	}

	if ap.hours == 10 && 0 < ap.minutes && ap.minutes <= minuteMax {
		ap.minutes += 600
		ap.hours = 0
	}

	if ap.minutes%10 != 0 && ap.minutes > 10 {
		// minute fraction is dropped for periods of at least a day (1:1440)
		minutes := ap.minutes / 10
		if !precise &&
			(ap.years > 0 || ap.months > 0 || ap.days > 0 || ap.hours >= 240) && minutes == 0 {
			ap.minutes = 0
		}
		return ap.condNegate(neg)
	}

	if ap.minutes == 10 && ap.hours == 0 && 0 < ap.seconds && ap.seconds <= secondMax {
		ap.seconds += 600
		ap.minutes = 0
	}

	if ap.seconds%10 != 0 {
		// second fraction is dropped for periods of at least an hour (1:3600)
		seconds := ap.seconds / 10
		if !precise &&
			(ap.years > 0 || ap.months > 0 || ap.days > 0 || ap.hours > 0 || ap.minutes >= 600) && seconds == 0 {
			ap.seconds = 0
		}
	}

	return ap.condNegate(neg)
}

// condNegate conditionally negate
func (p Period) condNegate(neg bool) Period {
	if neg {
		return p.Negate()
	}
	return p
}

// IsNegative is period negative
func (p Period) IsNegative() bool {
	return p.negative == true
}

// IsPositive returns true if its negative property is false
func (p Period) IsPositive() bool {
	return p.negative == false
}

// Negate changes the sign of the period.
func (p Period) Negate() Period {
	if p.IsNegative() {
		// p.Input = strings.ReplaceAll(p.Input, "-", "")
		p.negative = false
		return p
	}
	p.negative = true
	// p.Input = "-" + p.Input
	return p
}

// RunesToString convert runes list to string with no allocation
//
// WriteRune is more complex than WriteByte so can't inline
//
// A small cost a few ns in testing is incurred for using a string builder.
// There are no heap allocations using strings.Builder.
func RunesToString(runes ...rune) string {
	var sb = new(strings.Builder)
	for i := 0; i < len(runes); i++ {
		sb.WriteRune(runes[i])
	}
	return sb.String()
}

// IsZero is period emtpy
func (p Period) IsZero() bool {
	return p.Years() == 0 && p.Months() == 0 && p.Days() == 0 && p.Hours() == 0 && p.Minutes() == 0 && p.Seconds() == 0
}

// Duration converts a period to the equivalent duration in nanoseconds.
// A flag is also returned that is true when the conversion was precise and false otherwise.
//
// When the period specifies hours, minutes and seconds only, the result is precise.
// however, when the period specifies years, months and days, it is impossible to be precise
// because the result may depend on knowing date and timezone information, so the duration
// is estimated on the basis of a year being 365.2425 days as per Gregorian calendar rules)
// and a month being 1/12 of a that; days are all assumed to be 24 hours long.
func (p Period) Duration() (time.Duration, bool) {
	// remember that the fields are all fixed-point 1E1
	tdE6 := time.Duration(totalDaysApproxE7(p))
	stE3 := totalSecondsE3(p)
	if p.negative == true {
		return -(tdE6 + stE3), tdE6 == 0
	}
	return (tdE6 + stE3), tdE6 == 0
}

func totalSecondsE3(period Period) time.Duration {
	// remember that the fields are all fixed-point 1E1
	// and these are divided by 1E1
	hhE3 := time.Duration(period.hours) * time.Hour
	mmE3 := time.Duration(period.minutes) * time.Minute
	ssE3 := time.Duration(period.seconds) * time.Second
	return hhE3 + mmE3 + ssE3
}

func totalDaysApproxE7(period Period) time.Duration {
	// remember that the fields are all fixed-point 1E1
	ydE6 := time.Duration(period.years) * oneYearApprox
	mdE6 := time.Duration(period.months) * oneMonthApprox
	ddE6 := time.Duration(period.days) * oneDay

	return ydE6 + mdE6 + ddE6
}

func (p *Period) validate() error {
	var f []string
	if p.years > math.MaxInt64 {
		f = append(f, "years")
	}
	if p.months > math.MaxInt64 {
		f = append(f, "months")
	}
	if p.days > math.MaxInt64 {
		f = append(f, "days")
	}
	if p.hours > math.MaxInt64 {
		f = append(f, "hours")
	}
	if p.minutes > math.MaxInt64 {
		f = append(f, "minutes")
	}
	if p.seconds > math.MaxInt64 {
		f = append(f, "seconds")
	}

	if len(f) > 0 {
		// if p.Input == "" {
		// 	p.Input = p.String()
		// }
		return fmt.Errorf("integer overflow occurred in %s", strings.Join(f, ","))
	}

	return nil
}

// Normalise normalise period
func (p Period) Normalise(precise bool) *Period {
	n := p.normalise(precise)
	return n
}

func (p *Period) normalise(precise bool) *Period {
	return p.rippleUp(precise).moveFractionToRight()
}

// rippleUp move values up through category if boundaries passed
func (p *Period) rippleUp(precise bool) *Period {
	// remember that the fields are all fixed-point 1E1

	p.minutes += (p.seconds / 600) * 10
	p.seconds = p.seconds % 600

	p.hours += (p.minutes / 600) * 10
	p.minutes = p.minutes % 600
	// fmt.Println("hours", p.hours, "minutes", p.minutes)

	// 32670-(32670/60)-(32670/3600) = 32760 - 546 - 9.1 = 32204.9
	if !precise || p.hours > 32204 {
		p.days += (p.hours / 240) * 10
		p.hours = p.hours % 240
	}

	if !precise || p.days > 32760 {
		dE6 := p.days * oneE5
		p.months += (dE6 / daysPerMonthE6) * 10
		p.days = (dE6 % daysPerMonthE6) / oneE5
	}

	p.years += (p.months / 120) * 10
	p.months = p.months % 120
	// fmt.Println("hours", p.hours, "minutes", p.minutes)

	return p
}

// moveFractionToRight attempts to remove fractions in higher-order fields by moving their value to the
// next-lower-order field.
//
// The average number of days in a month is 30.436875 days, so getting a period
// with months in nit will make the adjustment approximate.
func (p *Period) moveFractionToRight() *Period {
	// remember that the fields are all fixed-point 1E1

	y10 := p.years % 10
	if y10 != 0 && p.years > 10 && (p.months != 0 || p.days != 0 || p.hours != 0 || p.minutes != 0 || p.seconds != 0) {
		p.months += y10 * 12
		p.years = (p.years / 10) * 10
	}

	m10 := p.months % 10
	if m10 != 0 && p.months > 10 && (p.days != 0 || p.hours != 0 || p.minutes != 0 || p.seconds != 0) {
		p.days += (m10 * daysPerMonthE6) / oneE6
		p.months = (p.months / 10) * 10
	}

	d10 := p.days % 10
	if d10 != 0 && p.days > 10 && (p.hours != 0 || p.minutes != 0 || p.seconds != 0) {
		p.hours += d10 * 24
		p.days = (p.days / 10) * 10
	}

	hh10 := p.hours % 10
	if hh10 != 0 && p.hours > 10 && (p.minutes != 0 || p.seconds != 0) {
		p.minutes += hh10 * 60
		// fmt.Println("minutes", p.minutes)
	}

	mm10 := p.minutes % 10
	if mm10 != 0 && p.minutes > 10 && p.seconds != 0 {
		p.seconds += mm10 * 60
		p.minutes = (p.minutes / 10) * 10
	}

	return p
}

// MustParse is as per Parse except that it panics if the string cannot be parsed.
// This is intended for setup code; don't use it for user inputs.
// By default, the value is normalised.
// Normalisation can be disabled using the optional flag.
func MustParse(value string, normalise ...bool) Period {
	d, err := Parse(value, normalise...)
	if err != nil {
		panic(err)
	}
	return d
}

// Parse parses strings that specify periods using ISO-8601 rules.
//
// In addition, a plus or minus sign can precede the period, e.g. "-P10D"
//
// By default, the value is normalised, e.g. multiple of 12 months become years
// so "P24M" is the same as "P2Y". However, this is done without loss of precision,
// so for example whole numbers of days do not contribute to the months tally
// because the number of days per month is variable.
//
// Normalisation can be disabled using the optional flag.
//
// The zero value can be represented in several ways: all of the following
// are equivalent: "P0Y", "P0M", "P0W", "P0D", "PT0H", PT0M", PT0S", and "P0".
// The canonical zero is "P0D".
func Parse(period string, normalise ...bool) (Period, error) {
	return ParseWithNormalise(period, len(normalise) == 0 || normalise[0])
}

// ParseWithNormalise parses strings that specify periods using ISO-8601 rules
// with an option to specify whether to normalise parsed period components.
//
// This method is deprecated and should not be used. It may be removed in a
// future version.
func ParseWithNormalise(period string, normalise bool) (Period, error) {
	if period == "" || period == "-" || period == "+" {
		return Period{}, fmt.Errorf("cannot parse a blank string as a period")
	}

	if period == "P0" {
		p := new(Period)
		// p.Input = "P0"
		return *p, nil
	}

	p, err := parse(period, normalise)
	if err != nil {
		return Period{}, err
	}

	return p, nil
}

// Parse parse a period
func parse(input string, normalise bool) (Period, error) {
	var orig = input

	var parts []rune

	period := new(Period)
	input = strings.ToUpper(input)
	// period.Input = input

	var isTime bool

	isValidChar := func(r rune) bool {
		switch r {
		case yearChar:
			return true
		case monthChar:
			return true
		case weekChar:
			return true
		case dayChar:
			return true
		case hourChar:
			return true
		// minuteChar same as monthChar
		case secondChar:
			return true
		case periodChar:
			return true
		case negativeChar:
			return true
		case timeChar:
			return true
		default:
			return false
		}
	}

	var periodFound bool = false

	for _, c := range input {
		c = unicode.ToUpper(c)

		// fmt.Println(string(c))
		if unicode.IsDigit(c) {
			parts = append(parts, c)
			continue
		}

		// Hnadle non digits
		if isValidChar(c) == true {
			if c == periodChar {
				if periodFound == false {
					periodFound = true
					continue
				}
				xfmt := new(xfmt.Buffer)
				msg := xfmt.S("only one period indicator allowed ").S(orig)
				return Period{}, errors.New(string(msg.Bytes()))
			}
			if c == negativeChar {
				period.negative = true
				continue
			}
			if c == timeChar {
				if isTime == true {
					xfmt := new(xfmt.Buffer)
					msg := xfmt.S("time must only be indicated once ").S(orig)
					return Period{}, errors.New(string(msg.Bytes()))
				}
				isTime = true
				continue
			}

			s := RunesToString(parts...)
			intVal, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return Period{}, err
			}
			if c == yearChar {
				if isTime == true {
					xfmt := new(xfmt.Buffer)
					msg := xfmt.S("non time part after time declared ").S(orig)
					return Period{}, errors.New(string(msg.Bytes()))
				}
				period.years = intVal
			} else if c == monthChar || c == minuteChar {
				if isTime == false {
					period.months = intVal
				} else {
					period.minutes = intVal
				}
			} else if c == weekChar {
				if isTime == true {
					xfmt := new(xfmt.Buffer)
					msg := xfmt.S("non time part after time declared ").S(orig)
					return Period{}, errors.New(string(msg.Bytes()))
				}
				period.weeks = intVal
			} else if c == dayChar {
				if isTime == true {
					xfmt := new(xfmt.Buffer)
					msg := xfmt.S("non time part after time declared ").S(orig)
					return Period{}, errors.New(string(msg.Bytes()))
				}
				period.days = intVal
			} else if c == hourChar {
				period.hours = intVal
			} else if c == minuteChar {
				period.minutes = intVal
			} else if c == secondChar {
				period.seconds = intVal
			}
			parts = make([]rune, 0, 0)
			continue
		}

		xfmt := new(xfmt.Buffer)
		msg := xfmt.S("character").C(c).S("is not valid")
		return Period{}, errors.New(string(msg.Bytes()))
	}

	period.days += period.weeks * 7
	// Zero out weeks as we have put them in days
	period.weeks = 0

	if normalise == true {
		period = period.Normalise(true)
	}

	return *period, nil
}
