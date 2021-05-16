package period

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/cockroachdb/apd"
	"github.com/imarsman/datetime/xfmt"
)

const yearChar = 'Y'

// const monthChar = 'M'
const weekChar = 'W'
const dayChar = 'D'
const hourChar = 'H'
const minuteChar = 'I'
const monthChar = 'O'
const minuteMonthChar = 'M'
const secondChar = 'S'
const periodChar = 'P'
const timeChar = 'T'
const negativeChar = '-'
const dotChar = '.'
const commaChar = ','

// NewOf converts a time duration to a Period, and also indicates whether the conversion is precise.
// Any time duration that spans more than ± 3276 hours will be approximated by assuming that there
// are 24 hours per day, 365.2425 days per year (as per Gregorian calendar rules), and a month
// being 1/12 of that (approximately 30.4369 days).
//
// The result is not always fully normalised; for time differences less than 3276 hours (about 4.5 months),
// it will contain zero in the years, months and days fields but the number of days may be up to 3275; this
// reduces errors arising from the variable lengths of months. For larger time differences, greater than
// 3276 hours, the days, months and years fields are used as well.
func NewOf(duration time.Duration) (p Period, precise bool) {
	var sign int64 = 1
	d := duration
	if duration < 0 {
		sign = -1
		d = -duration
	}

	sign10 := sign * 10

	totalHours := int64(d / time.Hour)

	// check for 16-bit overflow - occurs near the 4.5 month mark
	if totalHours < 3277 {
		// simple HMS case
		minutes := d % time.Hour / time.Minute
		seconds := d % time.Minute / hundredMSDuration
		return NewPeriod(0, 0, 0, sign10*totalHours, sign10*int64(minutes), sign*int64(seconds)), true
	}

	totalDays := totalHours / 24 // ignoring daylight savings adjustments

	if totalDays < 3277 {
		hours := totalHours - totalDays*24
		minutes := d % time.Hour / time.Minute
		seconds := d % time.Minute / hundredMSDuration
		return NewPeriod(0, 0, sign10*totalDays, sign10*hours, sign10*int64(minutes), sign*int64(seconds)), false
	}

	// TODO it is uncertain whether this is too imprecise and should be improved
	years := (oneE4 * totalDays) / daysPerYearE4
	months := ((oneE4 * totalDays) / daysPerMonthE4) - (12 * years)
	hours := totalHours - totalDays*24
	totalDays = ((totalDays * oneE4) - (daysPerMonthE4 * months) - (daysPerYearE4 * years)) / oneE4
	return NewPeriod(sign10*years, sign10*months, sign10*totalDays, sign10*hours, 0, 0), false
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
	t1GTt2 := true
	if t2.Before(t1) {
		t1, t2, t1GTt2 = t2, t1, false
	}

	if t1.Location() != t2.Location() {
		t2 = t2.In(t1.Location())
	}

	year, month, day, hour, min, sec, hundredth := daysDiff(t1, t2)

	p = NewPeriod(year, month, day, hour, min, sec)
	p.seconds = int64(hundredth)

	p.negative = t1GTt2 == false

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

	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	year = int64(y2 - y1)
	month = int64(m2 - m1)
	day = int64(d2 - d1)

	return
}

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
func (p *Period) Simplify(precise bool, th ...int) *Period {
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

func (p *Period) doSimplify(precise bool, monthMax, hourMax, minuteMax, secondMax int64) *Period {
	// What is this for?
	// if p.years%10 != 0 {
	// 	return p
	// }

	ap, neg := p.absNeg()

	// single year is dropped if there are some months
	if ap.years == 10 && 0 < ap.months && ap.months <= monthMax && ap.days == 0 {
		ap.months += 120
		ap.years = 0
	}

	// if ap.months%10 != 0 && ap.months > 10 {
	if ap.months > 10 {
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
func (p *Period) condNegate(neg bool) *Period {
	if neg {
		return p.Negate()
	}
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
func (p *Period) Normalise(precise bool) *Period {
	n := p.normalise(precise)
	return n
}

func (p *Period) normalise(precise bool) *Period {
	return p.rippleUp(precise).AdjustToRight(precise)
	// return p.rippleUp(precise)
}

// This can overflow with very large input values
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

	// Return empty duration with error if overflow
	if hourNumber < 0 && minuteNumber < 0 && secondNumber < 0 {
		return time.Duration(0), errors.New("Hour, minute, and second duration exceeds maximum")
	}

	return hourminutesecondDuration, nil
}

// This can overflow with very large input values
func ymdApproxDuration(p Period) (time.Duration, error) {
	yearDuration := time.Duration(p.years) * oneYearApproxNS
	monthDuration := time.Duration(p.months) * oneMonthApproxNS
	dayDuration := time.Duration(p.days) * oneDayNS
	yearMonthDayDuration := yearDuration + monthDuration + dayDuration

	yearNumber := int64(yearMonthDayDuration / oneYearApproxNS)
	remainder := int64(yearMonthDayDuration % oneYearApproxNS)

	monthNumber := int64(remainder / int64(oneMonthApproxNS))
	remainder = int64(yearMonthDayDuration % oneMonthApproxNS)

	dayNumber := int64(remainder / int64(oneDayNS))

	// Return empty duration with error if overflow
	if yearNumber < 0 && monthNumber < 0 && dayNumber < 0 {
		return time.Duration(0), errors.New("Year, month, and day duration exceeds maximum")
	}

	return yearMonthDayDuration, nil
}

// rippleUp move values up through category if boundaries passed Since this call
// uses durations to handle moving things around it is possible for the duration
// for an hour, minute, and second or year, month, and day duraton to overflow
// the duration's maximum value, which is 292.471208677536 years in nanoseconds.
// This is currently considered an acceptable cost as if there is an overflow it
// is handled by leaving the overflowing values alone.
// Precise as true will result in no adjustment for years, months, and days.
func (p *Period) rippleUp(precise bool) *Period {
	hourminutesecondDuration, err := hmsDuration(*p)
	if err == nil {
		hourNumber := int64(hourminutesecondDuration / time.Hour)
		remainder := int64(hourminutesecondDuration % time.Hour)

		minuteNumber := remainder / int64(time.Minute)
		remainder = int64(hourminutesecondDuration % time.Minute)

		secondNumber := remainder / int64(time.Second)

		p.hours = hourNumber
		p.minutes = minuteNumber
		p.seconds = secondNumber
	}

	if !precise {
		yearMonthDayDuration, err := ymdApproxDuration(*p)
		if err == nil {
			yearNumber := int64(yearMonthDayDuration / oneYearApproxNS)
			remainder := int64(yearMonthDayDuration % oneYearApproxNS)

			monthNumber := int64(remainder / int64(oneMonthApproxNS))
			remainder = int64(yearMonthDayDuration % oneMonthApproxNS)

			dayNumber := int64(remainder / int64(oneDayNS))

			p.years = yearNumber
			p.months = monthNumber
			p.days = dayNumber

		}
	}
	return p
}

// AdjustToRight attempts to remove fractions in higher-order fields by moving their value to the
// next-lower-order field.
//
// The average number of days in a month is 30.436875 days, so getting a period
// with months in nit will make the adjustment approximate.
// Note that precise is a cascade of normalize from parse.
func (p *Period) AdjustToRight(precise bool) *Period {
	// remember that the fields are all fixed-point 1E1

	y10 := p.years % 10
	if p.years > 10 {
		p.months += y10 * 12
		p.years = (p.years / 10) * 10
	}

	m10 := p.months % 10
	if !precise && p.months > 10 {
		p.days += (m10 * daysPerMonthE6) / oneE6
		p.months = (p.months / 10) * 10
	}

	d10 := p.days % 10
	if p.days > 10 {
		p.hours += d10 * 24
		p.days = (p.days / 10) * 10
	}

	hh10 := p.hours % 10
	if p.hours > 10 {
		p.minutes += hh10 * 60
		p.hours = (p.hours / 10) * 10
	}

	mm10 := p.minutes % 10
	if p.minutes > 10 {
		p.seconds += mm10 * 60
		p.minutes = (p.minutes / 10) * 10
	}

	return p
}

// AdditionsFromDecimalSection break down decimal section and get allocations to
// various parts
func AdditionsFromDecimalSection(part rune, whole, fractional int64) (
	years, months, days, hours, minutes, seconds int64, subseconds int, err error) {

	const oneMillion int64 = 1000000

	// Count digits in an integer
	var digitCount = func(number int64) int64 {
		var count int64 = 0
		for number != 0 {
			number /= 10
			count++
		}
		return count
	}

	// Check for whether a rune is a time part
	var isTimePart = func(r rune) bool {
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
		case minuteChar:
			return true
		case secondChar:
			return true
		default:
			return false
		}
	}

	// Exit with invalid characters
	if !isTimePart(part) {
		err := fmt.Errorf("Invalid time part %v to float", part)
		return years, months, days, hours, minutes, seconds, subseconds, err
	}

	// Tested to work with up to 15 billion years. Max for each part is
	// multipleied by 1000 because we are not using nanosecond but rather
	// millisecond level accuracy.
	// An error will be returned by the apd library if the precision is insufficient.
	// apcContext := apd.BaseContext.WithPrecision(200) // context for large
	// calculations if necessary

	// Go formatter will warn about

	const msOneYearApprox = oneYearApproxNS / time.Duration(oneMillion)  // a year of milliseconds
	const msOneMonthApprox = oneYearApproxNS / time.Duration(oneMillion) // a month of milliseconds
	const msOneDay = oneDayNS / time.Duration(oneMillion)                // a day of milliseconds
	const msOneHour = time.Hour / time.Duration(oneMillion)              // an hour of milliseconds
	const msOneMinute = time.Minute / time.Duration(oneMillion)          // a minute of milliseconds
	const msOneSecond = time.Second / time.Duration(oneMillion)          // a second of milliseconds

	const maxYears = 290 * 1000                // Maximum years before failing over to apd
	const maxMonths = maxYears * 12 * 1000     // Maximum months before failing over to apd
	const maxDays = maxYears * 12 * 365 * 1000 // Maximum hours before failing over to apd

	const maxHours = maxYears * 12 * 365 * 24 * 1000             // Maximum hours before failing over to apd
	const maxMinutes = maxYears * 12 * 365 * 24 * 60 * 1000      // Maximum minutes before failing over to apd
	const maxSeconds = maxYears * 12 * 365 * 24 * 60 * 60 * 1000 // Maximum seconds before failing over to apd

	var multiplier int64 = 0 // relative value to multiply by based on period part

	// Get an apd library fetch of too-large period part
	var getPart = func(multiplier int64, whole int64) (int64, error) {
		appContext := apd.BaseContext.WithPrecision(200) // context for large calculations if necessary
		var fullValueAPD = apd.New(0, 0)
		_, err = appContext.Mul(fullValueAPD, apd.New(whole, 0), apd.New(multiplier, 0))
		if err != nil {
			return 0, err
		}
		resultAPD := apd.New(0, 0)
		_, err := appContext.QuoInteger(resultAPD, fullValueAPD, apd.New(int64(multiplier), 0))
		if err != nil {
			return 0, err
		}
		wholeValue, err := resultAPD.Int64()
		if err != nil {
			return 0, err
		}

		return wholeValue, nil
	}

	// Find the section that applies and set it
	if part == yearChar {
		// multiplier = int64(oneYearApproxNS) / oneMillion
		multiplier = int64(msOneYearApprox)
		// Only use arbitrary precision decimals if we would overflow an int64
		if whole > maxYears {
			years, err = getPart(multiplier, whole)
			if err != nil {

			}
		} else {
			fullValue := whole * multiplier
			years = fullValue / int64(multiplier)
		}
	} else if part == monthChar {
		// multiplier = int64(oneMonthApproxNS) / oneMillion
		multiplier = int64(msOneMonthApprox)
		// Only use arbitrary precision decimals if we would overflow an int64
		if whole > maxMonths {
			months, err = getPart(multiplier, whole)
			if err != nil {
				return 0, 0, 0, 0, 0, 0, 0, err
			}
		} else {
			fullValue := whole * multiplier
			months = fullValue / int64(multiplier)
		}
	} else if part == dayChar {
		// multiplier = int64(oneDayNS) / oneMillion
		multiplier = int64(msOneDay)
		// Only use arbitrary precision decimals if we would overflow an int64
		if whole > maxDays {
			days, err = getPart(multiplier, whole)
			if err != nil {
				return 0, 0, 0, 0, 0, 0, 0, err
			}
		} else {
			fullValue := whole * multiplier
			days = fullValue / int64(multiplier)
		}
	} else if part == hourChar {
		// multiplier = int64(time.Hour) / oneMillion
		multiplier = int64(msOneHour)
		// Only use arbitrary precision decimals if we would overflow an int64
		if whole > maxHours {
			hours, err = getPart(multiplier, whole)
			if err != nil {
				return 0, 0, 0, 0, 0, 0, 0, err
			}
		} else {
			fullValue := whole * multiplier
			hours = fullValue / int64(multiplier)
		}
	} else if part == minuteChar {
		multiplier = int64(msOneMinute)
		// Only use arbitrary precision decimals if we would overflow an int64
		if whole > maxMinutes {
			minutes, err = getPart(multiplier, whole)
			if err != nil {
				return 0, 0, 0, 0, 0, 0, 0, err
			}
		} else {
			fullValue := whole * multiplier
			minutes = fullValue / int64(multiplier)
		}
	} else if part == secondChar {
		multiplier = int64(msOneSecond)
		// Only use arbitrary precision decimals if we would overflow an int64
		if whole > maxSeconds {
			seconds, err = getPart(multiplier, whole)
			if err != nil {
				return 0, 0, 0, 0, 0, 0, 0, err
			}
		} else {
			fullValue := whole * multiplier
			seconds = fullValue / int64(multiplier)
		}
	}

	var fractionalFloat = float64(fractional) / math.Pow(10, float64(digitCount(fractional))) // decimal value of post

	// We have already figured out the multiplier so don't need to do that again
	var postNano = int64(fractionalFloat * float64(multiplier)) // nanosecond value for fractional part

	// Overflow of an int64 occurs with about 192 years. The most that a portion
	// of a time period could be is just under one year. It is safe to do int64
	// calculations here.
	years += postNano / int64(msOneYearApprox)
	remainder := postNano % int64(msOneYearApprox)

	months += remainder / int64(msOneMonthApprox)
	remainder = postNano % int64(msOneMonthApprox)

	days += remainder / int64(msOneDay)
	remainder = postNano % int64(msOneDay)

	hours += remainder / int64(msOneHour)
	remainder = postNano % int64(msOneHour)

	minutes += remainder / int64(msOneMinute)
	remainder = postNano % int64(msOneMinute)

	seconds += remainder / int64(msOneSecond)
	remainder = postNano % int64(msOneSecond)

	subseconds += int(remainder)

	return years, months, days, hours, minutes, seconds, subseconds, nil
}

// MustParse is as per Parse except that it panics if the string cannot be parsed.
// This is intended for setup code; don't use it for user inputs.
// By default, the value is normalised.
// Normalisation can be disabled using the optional flag.
func MustParse(value string, normalise bool, precise ...bool) Period {
	d, err := Parse(value, normalise, precise...)
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
// Note that this will end up flagging precise in AjustRight.
func Parse(period string, normalize bool, precise ...bool) (Period, error) {
	usePrecise := false
	if len(precise) > 0 {
		usePrecise = precise[0]
	}
	return ParseWithNormalise(period, normalize, usePrecise)
}

// ParseWithNormalise parses strings that specify periods using ISO-8601 rules
// with an option to specify whether to normalise parsed period components.
//
// This method is deprecated and should not be used. It may be removed in a
// future version.
func ParseWithNormalise(period string, normalise bool, precise bool) (Period, error) {
	if period == "" || period == "-" || period == "+" {
		return Period{}, fmt.Errorf("period.ParseWithNormalise: cannot parse a blank string as a period")
	}

	if period == "P0" {
		p := new(Period)
		// p.Input = "P0"
		return *p, nil
	}

	p, err := parse(period, normalise, precise)
	if err != nil {
		return Period{}, err
	}

	return p, nil
}

// GetParts get the parts of a period
func parse(input string, normalise bool, precise bool) (Period, error) {

	var orig = input
	var period = Period{}

	var activePart []rune
	var decimalPart []rune
	var decimalSection rune
	var inDecimal bool

	// Ordering of period parts to allow some detection of malformed periods

	const (
		yearRank   = iota // rank order for year part
		monthRank         // rank order for month part
		weekRank          // rank order for week part
		dayRank           // rank order for day part
		hourRank          // rank order for hour part
		minuteRank        // rank order for minute part
		secondRank        // rand order for second part
	)

	input = strings.ToUpper(input)

	var isTime bool

	checkRank := func(old, new int) (int, error) {
		if old > new {
			return 0, fmt.Errorf("Ranks out of order %d and %d", old, new)
		}
		return new, nil
	}

	isValidChar := func(r rune) bool {
		switch r {
		case yearChar:
			return true
			// Double duty
		case minuteMonthChar:
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

	var inPeriod bool = false
	var maxSection int
	var decimalIn int

	var currentRank = yearRank

	for i, r := range input {
		maxSection = i

		if unicode.IsDigit(r) {
			activePart = append(activePart, r)

			continue
		}

		// Deal with decimal delimiter and continue gathering digits after
		// adding delimiter to activePart. Exit if decimalPart already allocated.
		if r == dotChar || r == commaChar {
			if len(decimalPart) > 0 {
				xfmt := new(xfmt.Buffer)
				msg := xfmt.S("period.parse: only one decimal section allowed ").S(orig)

				return Period{}, errors.New(string(msg.Bytes()))
			}
			inDecimal = true
			activePart = append(activePart, '.')

			continue
		}

		// Hnadle non digits
		if isValidChar(r) == true {
			// We should be just starting when we find this
			if r == periodChar {
				if inPeriod == false {
					inPeriod = true

					continue
				}
				// We have already found a period character P so this is an error
				xfmt := new(xfmt.Buffer)
				msg := xfmt.S("period.parse: only one period indicator allowed ").S(orig)

				return Period{}, errors.New(string(msg.Bytes()))
			}
			// Allow negative signs throughout. Ignore but set period negative state
			if r == negativeChar {
				period.negative = true

				continue
			}

			if r == timeChar {
				// If we are already in a time section this is an error
				if isTime == true {
					xfmt := new(xfmt.Buffer)
					msg := xfmt.S("period.parse: time must only be indicated once ").S(orig)
					return Period{}, errors.New(string(msg.Bytes()))
				}
				// Set state as being in time
				isTime = true

				continue
			}

			var intVal int64

			// If we have reached the end of the decimal part, clean up and continue
			if inDecimal == true {
				decimalPart = append(decimalPart, activePart...)
			} else {
				s := RunesToString(activePart...)
				var err error
				intVal, err = strconv.ParseInt(s, 10, 64)
				if err != nil {
					return Period{}, err
				}
			}

			if r == yearChar {
				if isTime == true {
					xfmt := new(xfmt.Buffer)
					msg := xfmt.S("period.parse: non time part after time declared ").S(orig)
					return Period{}, errors.New(string(msg.Bytes()))
				}
				var err error
				// Check ordering relative to previous
				currentRank, err = checkRank(currentRank, yearRank)
				if err != nil {
					return Period{}, err
				}

				if inDecimal == false {
					period.years = intVal
				} else {
					decimalSection = r
					decimalIn = maxSection
				}
			} else if r == minuteMonthChar {
				if isTime == false {
					var err error
					// Check ordering relative to previous
					currentRank, err = checkRank(currentRank, monthRank)
					if err != nil {
						return Period{}, err
					}

					if inDecimal == false {
						period.months = intVal
					} else {
						decimalSection = monthChar
						decimalIn = maxSection
					}
				} else {
					var err error
					// Check ordering relative to previous
					currentRank, err = checkRank(currentRank, minuteRank)
					if err != nil {
						return Period{}, err
					}

					if inDecimal == false {
						period.minutes = intVal
					} else {
						decimalSection = monthChar
						decimalIn = maxSection
					}
				}
			} else if r == weekChar {
				if isTime == true {
					xfmt := new(xfmt.Buffer)
					msg := xfmt.S("period.parse: non time part after time declared ").S(orig)
					return Period{}, errors.New(string(msg.Bytes()))
				}
				var err error
				// Check ordering relative to previous
				currentRank, err = checkRank(currentRank, weekRank)
				if err != nil {
					return Period{}, err
				}

				if inDecimal == false {
					period.weeks = intVal
				} else {
					decimalSection = r
					decimalIn = maxSection
				}
			} else if r == dayChar {
				if isTime == true {
					xfmt := new(xfmt.Buffer)
					msg := xfmt.S("period.parse: non time part after time declared ").S(orig)
					return Period{}, errors.New(string(msg.Bytes()))
				}
				var err error
				// Check ordering relative to previous
				currentRank, err = checkRank(currentRank, dayRank)
				if err != nil {
					return Period{}, err
				}

				if inDecimal == false {
					period.days = intVal
				} else {
					decimalSection = r
					decimalIn = maxSection
				}
			} else if r == hourChar {
				var err error
				// Check ordering relative to previous
				currentRank, err = checkRank(currentRank, hourRank)
				if err != nil {
					return Period{}, err
				}

				if inDecimal == false {
					period.hours = intVal
				} else {
					decimalSection = r
					decimalIn = maxSection
				}
			} else if r == minuteMonthChar {
				var err error
				// Check ordering relative to previous
				currentRank, err = checkRank(currentRank, minuteRank)
				if err != nil {
					return Period{}, err
				}

				if inDecimal == false {
					period.minutes = intVal
				} else {
					decimalSection = r
					decimalIn = maxSection
				}
			} else if r == secondChar {
				var err error
				// Check ordering relative to previous
				currentRank, err = checkRank(currentRank, secondRank)
				if err != nil {
					return Period{}, err
				}

				if inDecimal == false {
					period.seconds = intVal
				} else {
					decimalSection = r
					decimalIn = maxSection
				}
			}

			if inDecimal == true {
				inPeriod = false
				inDecimal = false
			}

			activePart = make([]rune, 0, 0)

			continue
		}

		xfmt := new(xfmt.Buffer)
		msg := xfmt.S("period.parse: character").C(r).S("is not valid")
		return Period{}, errors.New(string(msg.Bytes()))
	}

	period.days += period.weeks * 7
	// Zero out weeks as we have put them in days
	period.weeks = 0

	if len(decimalPart) > 0 {
		if decimalIn != maxSection {
			return Period{}, fmt.Errorf("period.parse: decimal must end in last character %dnot in %d", maxSection, decimalIn)
		}
		fmt.Println("period.parse: got decimal", RunesToString(decimalPart...), "in section", decimalSection)
	}

	if normalise == true {
		period = *period.Normalise(precise)
	}

	// Prints Size of period.Period struct: 64 bytes
	// for P130Y200D
	// fmt.Printf("Size of %T struct: %d bytes\n", *period, unsafe.Sizeof(*period))

	return period, nil
}
