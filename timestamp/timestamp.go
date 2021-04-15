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

	// https://golang.org/pkg/time/tzdata/
	/*
		    Package tzdata provides an embedded copy of the timezone database.
		    If this package is imported anywhere in the program, then if the
		    time package cannot find tzdata files on the system, it will use
		    this embedded information.

			Importing this package will increase the size of a program by about
			450 KB.

			This package should normally be imported by a program's main
			package, not by a library. Libraries normally shouldn't decide
			whether to include the timezone database in a program.

			This package will be automatically imported if you build with -tags
			timetzdata.
	*/
	// This will explicitly include tzdata in a build. See above for build flag.
	// You can do this in the main package if you choose.
	// _ "time/tzdata"

	"github.com/imarsman/datetime/timestamp/lex"
)

var namedZoneTimeFormats = []string{
	"Monday, 02-Jan-06 15:04:05 MST",
	"Mon, 02 Jan 2006 15:04:05 MST",
}

// timeFormats a list of Golang time formats to cycle through. The first match
// will cause the loop through the formats to exit.
//
// A mapping of time zone names to UTC offsets could be made, but there is
// unreliability of offset for some cases based on DST state.
//   e.g. Name, Location, offset
//   MT - Mountain Time, North America, UTC -7:00/-6:00
var nonISOTimeFormats = []string{

	// "Monday, 02-Jan-06 15:04:05 MST",
	// "Mon, 02 Jan 2006 15:04:05 MST",

	// RFC7232 - used in HTTP protocol
	"Mon, 02 Jan 2006 15:04:05 GMT",

	// RFC850
	// Unreliable to have Zone name known - don't try
	// "Monday, 02-Jan-06 15:04:05 MST",

	// RFC1123
	// Unreliable to have Zone name known - don't try
	// "Mon, 02 Jan 2006 15:04:05 MST",

	// RFC1123Z
	"Mon, 02 Jan 2006 15:04:05 -0700",

	"Mon, 02 Jan 2006 15:04:05",
	"Monday, 02-Jan-2006 15:04:05",

	// RFC822Z
	"02 Jan 06 15:04 -0700",

	// Just in case
	"2006-01-02 15-04-05",
	"20060102150405",

	// Stamp
	// Year not known - don't try
	// "Jan _2 15:04:05",

	// StampMilli
	// Year not known - don't try
	// "Jan _2 15:04:05.000",

	// StampMicro
	// Year not known - don't try
	// "Jan _2 15:04:05.000000",

	// StampNano
	// Year not known - don't try
	// "Jan _2 15:04:05.000000000",

	// Hopefully less likely to be found. Assume UTC.
	"20060102",
	"01/02/2006",
	"1/2/2006",
}

var timeFormats = []string{}

func init() {
	timeFormats = append(timeFormats, nonISOTimeFormats...)
}

// LocationForZone get a location from a named zone
func LocationForZone(zone string) (*time.Location, error) {
	location, err := time.LoadLocation(strings.TrimSpace(zone))
	if err != nil {
		return nil, err
	}
	return location, nil
}

// OffsetForZone get offset data for a zone. This assumes that a UTC time is
// available to start with then adjust for the zone.
func OffsetForZone(t time.Time, zone string) (hours, minutes int, err error) {
	location, err := time.LoadLocation(zone)
	if err != nil {
		return 0, 0, err
	}

	_, tzOffset := t.In(time.UTC).In(location).Zone()

	d, err := time.ParseDuration(fmt.Sprint(tzOffset) + "s")
	hours = int(d.Hours())
	m := int(int64(d.Minutes()) % 60)
	minutes = int(math.Abs(float64(m)))

	return hours, minutes, nil
}

// OffsetString get an offset in HHMM format based on hours and minutes offset
func OffsetString(hours, minutes int) string {
	return fmt.Sprintf("%+03d%02d", hours, minutes)
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

// ParseInUTC parse for all timestamps, defaulting to UTC, and return UTC zoned time
func ParseInUTC(timeStr string) (time.Time, error) {
	return parseTimestamp(timeStr, time.UTC, false)
}

// ParseISOInUTC parse limited to ISO timestamp formats and return UTC zoned time
func ParseISOInUTC(timeStr string) (time.Time, error) {
	return parseTimestamp(timeStr, time.UTC, true)
}

// ParseInLocation parse for all timestamp formats and return time with specific
// time library location
func ParseInLocation(timeStr string, location *time.Location) (time.Time, error) {
	return parseTimestamp(timeStr, location, false)
}

// ParseISOInLocation parse limited to ISO timestamp formats and return time with
// UTC zoned time
func ParseISOInLocation(timeStr string, location *time.Location) (time.Time, error) {
	return parseTimestamp(timeStr, location, true)
}

// ParseTimestampInLocation and return time with specific time library location
func parseTimestamp(timeStr string, location *time.Location, isoOnly bool) (time.Time, error) {
	original := timeStr

	// Try ISO parsing first. The lexer is tolerant of some inconsistency in
	// format that is not ISO-8601 compliant, such as dashes where there should
	// be colons and a space instead of a T to separate date and time.
	t, err := lex.ParseInLocation([]byte(timeStr), location)
	if err == nil {
		return t, nil
	}

	if isoOnly == true {
		return time.Time{}, errors.New("No ISO format match")
	}

	// Don't try to
	zoneMatch, err := regexp.MatchString("\\s\\w+\\/\\w+", timeStr)
	if err != nil {
		return time.Time{}, err
	}
	if zoneMatch == true {
		re := regexp.MustCompile("^.*(\\s\\w{2,}\\/\\w{2,}).*$")
		zone := re.ReplaceAllString(timeStr, "$1")
		return time.Time{}, fmt.Errorf("Can't parse time zone for location %s", strings.TrimSpace(zone))
	}

	s := nonISOTimeFormats
	for _, format := range s {
		// If no zone in timestamp use location
		t, err := time.ParseInLocation(format, original, location)
		if err == nil {
			t = t.In(time.UTC)
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
		t := time.Unix(s, n).In(location)

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

	// var b []byte
	// return string(t.AppendFormat(b, "2006-01-02T15:04:05.000-07:00"))
	return t.Format("2006-01-02T15:04:05.000-07:00")
}

// func ISO8601LongMsec2(t time.Time) string {
// 	t = t.In(time.UTC)

// 	// var b []byte
// 	// return string(t.AppendFormat(b, "2006-01-02T15:04:05.000-07:00"))
// 	return t.Format("2006-01-02T15:04:05.000-07:00")
// }

// StartTimeIsBeforeEndTime if time 1 is before time 2 return true, else false
func StartTimeIsBeforeEndTime(t1 time.Time, t2 time.Time) bool {
	t1 = t1.In(time.UTC)
	t2 = t2.In(time.UTC)

	return t2.Unix()-t1.Unix() > 0
}
