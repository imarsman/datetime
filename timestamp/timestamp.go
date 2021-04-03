package timestamp

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/imarsman/datetime/date"
	"github.com/imarsman/datetime/timespan"

	"github.com/imarsman/datetime/period"
)

// timeFormats a list of Golang time formats to cycle through. The first match
// will cause the loop through the formats to exit.
var timeFormats = []string{

	// RFC7232 - used in HTTP protocol
	"Mon, 02 Jan 2006 15:04:05 GMT",

	// Short ISO-8601 timestamps with numerical zone offsets
	"20060102T150405-0700",
	"20060102T150405-07",
	"20060102T150405.999-0700",
	"20060102T150405.999-07",
	"20060102T150405.999999-0700",
	"20060102T150405.999999-07",
	"20060102T150405.999999999-0700",
	"20060102T150405.999999999-07",

	// Long ISO-8601 timestamps with numerical zone offsets
	"2006-01-02T15:04:05-07:00",
	"2006-01-02T15:04:05.999-07:00",
	"2006-01-02T15:04:05.999-07",
	"2006-01-02T15:04:05.999999-07:00",
	"2006-01-02T15:04:05.999999-07",
	"2006-01-02T15:04:05.999999999-07:00",
	"2006-01-02T15:04:05.999999999-07",

	// Short  ISO-8601 timestamps with zulu zone offsets
	"20060102T150405Z",
	"20060102T150405.999Z",
	"20060102T150405.999Z",
	"20060102T150405.999999999Z",

	// Long ISO-8601 timestamps with zulu zone offsets
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05.999Z",
	"2006-01-02T15:04:05.999999Z",
	"2006-01-02T15:04:05.999999999Z",

	// Just in case
	"2006-01-02 15-04-05",
	"20060102150405",

	// Short ISO-8601 timestamps with no zone offset. Assume UTC.
	"20060102T150405",
	"20060102T150405.999",
	"20060102T150405.999999",
	"20060102T150405.999999999",

	// SQL
	"2006-01-02 15:04:05",
	"2006-01-02 15:04:05 -07",
	"2006-01-02 15:04:05 -07:00",

	// Hopefully less likely to be found. Assume UTC.
	"20060102",
	"2006-01-02",
	"2006/01/02",
	"01/02/2006",
	"1/2/2006",

	// Weird ones with improper separators
	"2006-01-02T15-04-05Z",
	"2006-01-02T15-04-05.999Z",
	"2006-01-02T15-04-05.999999Z",
	"2006-01-02T15-04-05.999999999Z",

	// Weird ones with improper separators
	"2006-01-02T15-04-05-0700",
	"2006-01-02T15-04-05-07",
	"2006-01-02T15-04-05.000-0700",
	"2006-01-02T15-04-05.000-07",
	"2006-01-02T15-04-05.000000-0700",
	"2006-01-02T15-04-05.000000-07",
	"2006-01-02T15-04-05.999999999-0700",
	"2006-01-02T15-04-05.999999999-07",

	"2006-01-02T15-04-05-07:00",
	"2006-01-02T15-04-05.999-07:00",
	"2006-01-02T15-04-05.999-07",
	"2006-01-02T15-04-05.999999-07:00",
	"2006-01-02T15-04-05.999999-07",
	"2006-01-02T15-04-05.999999999-07:00",
	"2006-01-02T15-04-05.999999999-07",

	// time.RFC3339,
	"2006-01-02T15:04:05Z07:00",
}

// TimespanForDateRange parse an ISO8601 date range which normally is written like
//   yyyy-mm-dd/yyyy-mm-dd,
//
// indicating start/end. The returned values will be the first nanosecond of the
// start date and the first nannosecond following the end of the range.
func TimespanForDateRange(start, end string) (time.Time, time.Time, error) {
	d1, err := date.Parse("2006-01-02", start)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("error when parsing start %v", err)
	}
	d2, err := date.Parse("2006-01-02", end)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("error when parsing end %v", err)
	}
	ts := timespan.NewTimeSpan(d1.In(time.UTC), d2.In(time.UTC))
	s := ts.Start()
	e := ts.End()

	return s, e, nil
}

// RangeDate returns a date range function over start date to end date inclusive.
// After the end of the range, the range function returns a zero date,
// date.IsZero() is true.
//
// Sample usage assuming building a map with empty string values:
/*
  for rd := dateutils.RangeDate(start, end); ; {
	 date := rd()
	 if date.IsZero() {
	   break
	 }
	 indicesForDays[getIndexForDate(*date)] = ""
  }
*/
func RangeDate(start, end time.Time) func() time.Time {
	start = start.In(time.UTC)
	end = end.In(time.UTC)

	y, m, d := start.Date()
	start = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	y, m, d = end.Date()
	end = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	return func() time.Time {
		if start.After(end) {
			return time.Time{}
		}
		date := start
		start = start.AddDate(0, 0, 1)

		return date
	}
}

// DatesInRange get dates in range.
func DatesInRange(d1, d2 string) ([]string, error) {
	start, err := TimeForDate(d1)
	if err != nil {
		return nil, err
	}

	end, err := TimeForDate(d2)
	if err != nil {
		return nil, err
	}

	dates := make(map[string]string)

	for rd := RangeDate(start, end); ; {
		date := rd()
		if date.IsZero() {
			break
		}
		dates[DateForTime(date)] = ""
	}

	keys := make([]string, 0, len(dates))

	for k := range dates {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys, nil
}

// DateRangeFromDates get date range between two dates
//   e.g. 2020-01-01/2021/01/02
// There is a lot of flexibility in time intevals in ISO-8601. This function
// only supports two dates. The first value is the start of the range. The
// second value is the date following the end of the range.
func DateRangeFromDates(d1, d2 string) (string, error) {
	if !IsDate(d1) {
		return "", fmt.Errorf("%s is not a valid date", d1)
	}
	if !IsDate(d2) {
		return "", fmt.Errorf("%s is not a valid date", d2)
	}
	if d1 != d2 {
		if !(d1 < d2) {
			return "", fmt.Errorf("%s is not >= %s", d1, d2)
		}
	}

	return d1 + "/" + d2, nil
}

// TimeDateOnly get date with zero time values
func TimeDateOnly(t time.Time) time.Time {
	t = t.In(time.UTC)

	newT := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return newT
}

// parseUnixTS returns seconds and nanoseconds from a timestamp that has the
// format "%d.%09d", time.Unix(), int64(time.Nanosecond()))
// if the incoming nanosecond portion is longer or shorter than 9 digits it is
// converted to nanoseconds. The expectation is that the seconds and
// seconds will be used to create a time variable.
//
// For example:
/*
	var t int64
    seconds, nanoseconds, err := ParseTimestamp("1136073600.000000001",0)
    if err == nil {
		t = time.Unix(seconds, nanoseconds)
	}
*/
// From https://github.com/moby/moby/blob/master/api/types/time/timestamp.go
// Part of Docker, under Apache licence.
func parseUnixTS(value string) (int64, int64, error) {
	sa := strings.SplitN(value, ".", 2)
	s, err := strconv.ParseInt(sa[0], 10, 64)
	if err != nil {
		return s, 0, err
	}
	if len(sa) != 2 {
		return s, 0, nil
	}
	n, err := strconv.ParseInt(sa[1], 10, 64)
	if err != nil {
		return s, n, err
	}
	// should already be in nanoseconds but just in case convert n to nanoseconds
	n = int64(float64(n) * math.Pow(float64(10), float64(9-len(sa[1]))))
	return s, n, nil
}

// ParseTimestampInUTC parse time string and return UTC zoned time
func ParseTimestampInUTC(timeStr string) (time.Time, error) {
	return ParseTimestampInLocation(timeStr, *time.UTC)
}

// ParseTimestampInLocation and return time with specific time library location
func ParseTimestampInLocation(timeStr string, loc time.Location) (time.Time, error) {
	// Continue on for non unix timestamp patterns
	for _, format := range timeFormats {
		t, err := time.Parse(format, timeStr)
		if err == nil {
			t = t.In(&loc)

			return t, nil
		}
	}

	// Deal with oddball unix timestamp
	match, err := regexp.MatchString("^\\d+$", timeStr)
	if err != nil {
		return time.Time{}, errors.New("Could not parse time")
	}
	// Don't support timestamps less than 7 characters in length
	// to avoid strange date formats from being parsed.
	// Max would be 9999999, or Sun Apr 26 1970 17:46:39 GMT+0000
	if match == true && len(timeStr) > 6 {
		toSend := timeStr
		// Break it into a format that has a period between second and
		// millisecond portions for the function.
		if len(timeStr) > 10 {
			sec, nsec := timeStr[0:10], timeStr[11:len(timeStr)-1]
			toSend = sec + "." + nsec
		}
		// Get seconds, nanoseconds, and error if there was a problem
		s, n, err := parseUnixTS(toSend)
		if err != nil {
			return time.Time{}, err
		}
		// If it was a unix seconds timestamp n will be zero. If it was a
		// nanoseconds timestamp there will be a nanoseconds portion that is not
		// zero.
		t := time.Unix(s, n).In(&loc)

		return t, nil
	}

	return time.Time{}, fmt.Errorf("Could not parse time %s", timeStr)
}

// RFC7232 get format used for http headers
//   "Mon, 02 Jan 2006 15:04:05 GMT"
//
// TimeFormat is the time format to use when generating times in HTTP headers.
// It is like time.RFC1123 but hard-codes GMT as the time zone. The time being
// formatted must be in UTC for Format to generate the correct format. This is
// done in the function before the call to format.
func RFC7232(t time.Time) string {
	t = t.In(time.UTC)

	return t.Format(http.TimeFormat)
}

// ISO8601Short ISO-8601 timestamp with no seconds
//   "20060102T150405-0700"
func ISO8601Short(t time.Time) string {
	t = t.In(time.UTC)

	return t.Format("20060102T150405-0700")
}

// ISO8601ShortMsec ISO-8601 timestamp with no seconds
//   "20060102T150405.000-0700"
func ISO8601ShortMsec(t time.Time) string {
	t = t.In(time.UTC)

	return t.Format("20060102T150405.000-0700")
}

// ISO8601Long ISO-8601 timestamp long format string result
//   "2006-01-02T15:04:05-07:00"
func ISO8601Long(t time.Time) string {
	t = t.In(time.UTC)

	return t.Format("2006-01-02T15:04:05-07:00")
}

// ISO8601LongMsec ISO-8601 longtimestamp with msec
//   "2006-01-02T15:04:05.000-07:00"
func ISO8601LongMsec(t time.Time) string {
	t = t.In(time.UTC)

	return t.Format("2006-01-02T15:04:05.000-07:00")
}

// IsPeriod is input a valid ISO-8601 period format
func IsPeriod(input string) bool {
	// Convert to uppercase as this would be an irritating source of errors
	_, err := period.ParseWithNormalise(strings.ToUpper(input), true)
	return err == nil
}

// ParsePeriod get period string
func ParsePeriod(input string) (string, error) {
	if IsPeriod(input) == false {
		return "", fmt.Errorf("invalid period %s", input)
	}
	// Convert to uppercase as this would be an irritating source of errors
	p, _ := period.ParseWithNormalise(strings.ToUpper(input), true)

	return p.String(), nil
}

// DurationToPeriod get period string from time.Duration
func DurationToPeriod(d time.Duration) string {
	p, _ := period.NewOf(d)
	return p.String()
}

// PeriodPositive get positive value of a period
func PeriodPositive(input string) (string, error) {
	if IsPeriod(input) == false {
		return "", fmt.Errorf("invalid period %s", input)
	}
	p, _ := period.ParseWithNormalise(strings.ToUpper(input), true)
	if p.IsNegative() {
		p = p.Negate()
	}

	return p.String(), nil
}

// PeriodNegative negate a period
func PeriodNegative(input string) (string, error) {
	if IsPeriod(input) == false {
		return "", fmt.Errorf("invalid period %s", input)
	}
	p, _ := period.ParseWithNormalise(strings.ToUpper(input), true)
	if p.IsPositive() {
		p = p.Negate()
	}

	return p.String(), nil
}

// TimeAddPeriod add a period to a time
func TimeAddPeriod(t time.Time, input string) (time.Time, error) {
	t = t.In(time.UTC)

	if IsPeriod(input) == false {
		fmt.Printf("%s is not a valid period\n", input)
		return time.Time{}, fmt.Errorf("invalid period %s", input)
	}

	p, err := period.ParseWithNormalise(strings.ToUpper(input), true)
	if err != nil {
		fmt.Printf("Got error parsing period %v\n", err)
		return time.Time{}, err
	}

	newTime, _ := p.AddTo(t)
	newTime = newTime.In(time.UTC)

	return newTime, nil
}

// TimeSubtractPeriod subtract a period from a time. If incoming period is negative
// it will be handled properly (by making sure it is subtracted.
func TimeSubtractPeriod(t time.Time, input string) (time.Time, error) {
	if IsPeriod(input) == false {
		return time.Time{}, fmt.Errorf("Invalid period %s", input)
	}

	p, err := period.ParseWithNormalise(strings.ToUpper(input), true)
	if err != nil {
		return time.Time{}, err
	}

	newTime, _ := p.Negate().AddTo(t)
	if p.IsNegative() == true {
		newTime, _ = p.AddTo(t)
	}

	newTime = newTime.In(time.UTC)
	return newTime, nil
}

// StartTimeIsBeforeEndTime if time 1 is before time 2 return true, else false
func StartTimeIsBeforeEndTime(t1 time.Time, t2 time.Time) bool {
	t1 = t1.In(time.UTC)
	t2 = t2.In(time.UTC)

	return t2.Unix()-t1.Unix() > 0
}

// StartDateIsBeforeEndDate if date 1 is before date 2 return true, else false
func StartDateIsBeforeEndDate(d1, d2 string) (bool, error) {
	if !IsDate(d1) {
		return false, errors.New(d1 + " is not valid")
	} else if !IsDate(d2) {
		return false, errors.New(d2 + " is not valid")
	}

	return d1 < d2, nil
}

// DatesEqual are two dates equal
func DatesEqual(d1, d2 string) (bool, error) {
	if !IsDate(d1) {
		return false, errors.New(d1 + " is not valid")
	} else if !IsDate(d2) {
		return false, errors.New(d2 + " is not valid")
	}

	return d1 == d2, nil
}

// PositivePeriodBetween get a positive period value between two times
// Can be negated to get negative period.
func PositivePeriodBetween(t1 time.Time, t2 time.Time) string {
	t1 = t1.In(time.UTC)
	t2 = t2.In(time.UTC)

	p := period.Between(t1, t2)
	if p.IsNegative() == true {
		p = p.Negate()
	}
	p = period.New(p.Years(), p.Months(), p.Days(), p.Hours(), p.Minutes(), p.Seconds())

	return p.String()
}

// NegativePeriodBetween get a negative period value between two times.
// Can be negated to get negative period.
func NegativePeriodBetween(t1 time.Time, t2 time.Time) string {
	t1 = t1.In(time.UTC)
	t2 = t2.In(time.UTC)

	p := period.Between(t1, t2)
	if p.IsPositive() == true {
		p = p.Negate()
	}
	p = period.New(p.Years(), p.Months(), p.Days(), p.Hours(), p.Minutes(), p.Seconds())

	return p.String()
}

// PeriodBetweenTimes get a period value between two times
func PeriodBetweenTimes(t1 time.Time, t2 time.Time) string {
	t1 = t1.In(time.UTC)
	t2 = t2.In(time.UTC)

	p := period.Between(t1, t2)
	p = period.New(p.Years(), p.Months(), p.Days(), p.Hours(), p.Minutes(), p.Seconds())

	return p.String()
}

// TimeForDate get time for date string.
// The returned value will have zero values for all time parts
func TimeForDate(ds string) (time.Time, error) {
	d, err := date.Parse(date.ISO8601, ds)
	if err != nil {
		// t := time.Now()
		// d := date.New(t.Year(), t.Month(), t.Day())
		// t2 := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
		return time.Time{}, err
	}
	t2 := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
	return t2, nil
}

// IsDate check if a date string is valid
//   Format 2006-01-02
func IsDate(ds string) bool {
	_, err := date.Parse("2006-01-02", ds)
	return err == nil
}

// DateForTime get date string for time
//   Format 2006-01-02
func DateForTime(t time.Time) string {
	t = t.In(time.UTC)

	d := date.New(t.Year(), t.Month(), t.Day())
	return d.Format(date.ISO8601)
}

// IsSameDate check whether the y, m, and d portions of a timestamp are
// equivalent to a date string. Time is converted to UTC first.
func IsSameDate(base string, t time.Time) bool {
	t = t.In(time.UTC)

	ds := date.New(t.Year(), t.Month(), t.Day()).Format("200")
	return base == ds
}

// Period1LessThanPeriod2 check whether p1 is less than p2
//   e.g. P3D < PT23H
func Period1LessThanPeriod2(p1S string, p2S string) (bool, error) {
	p1, err := period.ParseWithNormalise(p1S, true)
	if err != nil {
		return false, fmt.Errorf("problem parsing %s %v", p1S, err)
	}

	p2, err := period.ParseWithNormalise(p2S, true)
	if err != nil {
		return false, fmt.Errorf("problem parsing %s %v", p2S, err)
	}
	check := p1.DurationApprox() < p2.DurationApprox()

	return check, nil
}
