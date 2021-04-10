package timestamp

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/imarsman/datetime/timestamp/lex"
)

var isoTimeFormats = []string{
	// Short ISO-8601 timestamps with numerical zone offsets
	"20060102T150405-0700",
	"20060102T150405.999-0700",
	"20060102T150405.999999-0700",
	"20060102T150405.999999999-0700",

	"20060102T150405-07",
	"20060102T150405.999-07",
	"20060102T150405.999999-07",
	"20060102T150405.999999999-07",

	// Long ISO-8601 timestamps with numerical zone offsets
	// "2006-01-02T15:04:05-07:00",
	// "2006-01-02T15:04:05.999-07:00",
	// "2006-01-02T15:04:05.999-07",
	// "2006-01-02T15:04:05.999999-07:00",
	// "2006-01-02T15:04:05.999999-07",
	// "2006-01-02T15:04:05.999999999-07:00",
	// "2006-01-02T15:04:05.999999999-07",

	// Short  ISO-8601 timestamps with UTC zone offsets
	"20060102T150405Z",
	"20060102T150405.999Z",
	"20060102T150405.999Z",
	"20060102T150405.999999999Z",

	// Long ISO-8601 timestamps with UTC zone offsets
	// "2006-01-02T15:04:05Z",
	// "2006-01-02T15:04:05.999Z",
	// "2006-01-02T15:04:05.999999Z",
	// "2006-01-02T15:04:05.999999999Z",
}

// timeFormats a list of Golang time formats to cycle through. The first match
// will cause the loop through the formats to exit.
var nonISOTimeFormats = []string{

	// RFC7232 - used in HTTP protocol
	"Mon, 02 Jan 2006 15:04:05 GMT",

	// Just in case
	"2006-01-02 15-04-05",
	"20060102150405",

	// Short ISO-8601 timestamps with no zone offset. Assume UTC.
	"20060102T150405",
	"20060102T150405.999",
	"20060102T150405.999999",
	"20060102T150405.999999999",

	// SQL
	"20060102 150405",
	"20060102 150405 -07",
	"20060102 150405 -07:00",

	"2006-01-02 15:04:05",
	"2006-01-02 15:04:05 -07",
	"2006-01-02 15:04:05 -07:00",

	"2006-01-02 15:04:05-07",
	"2006-01-02 15:04:05.000-07",
	"2006-01-02 15:04:05.000000-07",
	"2006-01-02 15:04:05.000000000-07",

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

	// Actually, Golang's parse can't parse the RFC2229 format as defined as a
	// constant by the Go time library.
	// time.RFC3339,
	// "2006-01-02T15:04:05Z07:00",
}

var timeFormats = []string{}

func init() {
	timeFormats = append(timeFormats, isoTimeFormats...)
	timeFormats = append(timeFormats, nonISOTimeFormats...)
}

// RangeOverTimes returns a date range function over start date to end date inclusive.
// After the end of the range, the range function returns a zero date,
// date.IsZero() is true.
//
// Sample usage assuming building a map with empty string values:
/*
  for rd := dateutils.RangeOverTimes(start, end); ; {
	 date := rd()
	 if date.IsZero() {
	   break
	 }
	 indicesForDays[getIndexForDate(*date)] = ""
  }
*/
func RangeOverTimes(start, end time.Time) func() time.Time {
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

// ParseUTC parse for all timestamps and return UTC zoned time
func ParseUTC(timeStr string) (time.Time, error) {
	return parseTimestampInLocation(timeStr, time.UTC)
}

// ParseISOUTC parse for ISO timestamp formats and return UTC zoned time
func ParseISOUTC(timeStr string) (time.Time, error) {
	return parseTimestampInLocation(timeStr, time.UTC)
}

// ParseInLocation parse for all timestamp formats and return time with specific
// time library location
func ParseInLocation(timeStr string, loc *time.Location) (time.Time, error) {
	return parseTimestampInLocation(timeStr, loc)
}

// ParseISOInLocation parse for ISO timestamp formats and return time with
// UTC zoned time
func ParseISOInLocation(timeStr string, loc *time.Location) (time.Time, error) {
	return parseTimestampInLocation(timeStr, loc)
}

// ParseTimestampInLocation and return time with specific time library location
func parseTimestampInLocation(timeStr string, loc *time.Location) (time.Time, error) {
	original := timeStr

	// Try ISO parsing first. The lexer is tolerant of some inconsistency in
	// format that is not ISO-8601 compliant, such as dashes where there should
	// be colons and a space instead of a T to separate date and time.
	t, err := lex.Parse([]byte(timeStr))
	if err == nil {
		return t, nil
	}

	s := nonISOTimeFormats
	for _, format := range s {
		t, err := time.Parse(format, original)
		if err == nil {
			t = t.In(loc)

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
		t := time.Unix(s, n).In(loc)

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

// StartTimeIsBeforeEndTime if time 1 is before time 2 return true, else false
func StartTimeIsBeforeEndTime(t1 time.Time, t2 time.Time) bool {
	t1 = t1.In(time.UTC)
	t2 = t2.In(time.UTC)

	return t2.Unix()-t1.Unix() > 0
}
