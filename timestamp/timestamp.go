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
	"unicode"

	"github.com/imarsman/datetime/xfmt"
	// gocache "github.com/patrickmn/go-cache"
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

			This package will be automatically imported if you build with
			  -tags timetzdata
	*/// This will explicitly include tzdata in a build. See above for build flag.
	// You can do this in the main package if you choose.
	// _ "time/tzdata"
)

// Can view allocation analysis with
//   go build -gcflags '-m -m' timestamp.go 2>&1 |less

var reDigits *regexp.Regexp

// var zoneCache = gocache.New(1*time.Hour, 5*time.Minute)

var namedZoneTimeFormats = []string{
	"Monday, 02-Jan-06 15:04:05 MST",
	"Mon, 02 Jan 2006 15:04:05 MST",
}

// timeFormats a list of Golang time formats to cycle through. The first match
// will cause the loop through the formats to exit.
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

var timeFormats = []string{}           // A slice of time formats to be used if ISO parsing fails
var cachedZones map[int]*time.Location // A cache of zones by offset seconds

func init() {
	reDigits = regexp.MustCompile(`^\d+\.?\d+$`)
	timeFormats = append(timeFormats, nonISOTimeFormats...)
	// A cache for zones tied to offsets to save quite a bit of time and 3
	// allocations needed to get a fixed zone.
	cachedZones = make(map[int]*time.Location)
}

// runesToString join bytes with no allocation
//
// WriteRune is more complex than WriteByte so can't inline
// A small cost a few ns in testing is incurred for using a string builder.
// There are no heap allocations using strings.Builder.
func runesToString(runes ...rune) string {
	var sb = new(strings.Builder)
	for i := 0; i < len(runes); i++ {
		sb.WriteRune(runes[i])
	}
	return sb.String()
}

// bytesToString join bytes with no allocation
//
// can inline - strings.Builder WriteByte is less complex than WriteRune
// A small cost a few ns in testing is incurred for using a string builder.
// There are no heap allocations using strings.Builder.
func bytesToString(bytes ...byte) string {
	var sb = new(strings.Builder)
	for i := 0; i < len(bytes); i++ {
		sb.WriteByte(bytes[i])
	}
	return sb.String()
}

// OffsetForLocation get offset data for a named zone such a America/Tornto or EST
// or MST. Based on date the offset for a zone can differ, with, for example, an
// offset of -0500 for EST in the summer and -0400 for EST in the winter. This
// assumes that a year, month, and day is available and have been used to create
// the date to be analyzed. Based on this the offset for the supplied zone name
// is obtained. This has to be tested more, in particular the calculations to
// get the minutes.
//
// Get integer value of hours offset
//   hours = int(d.Hours())
//
// For 5.5 hours of offset or 0530
//  60 × 5.5 = 330 minutes total offset
//  330 % 60 = 30 minutes
//
// For an offset of 4.25 hours or 0415
//  60 × 4.25 = 255 minutes total offset
//  255 % 60 = 15 minutes
//
// If the zone is not recognized in Go's tzdata database an error will be
// returned.
//
// Can't inline
func OffsetForLocation(year int, month time.Month, day int, location string) (d time.Duration, err error) {
	l, err := time.LoadLocation(location)
	if err != nil {
		return 0, err
	}

	t := time.Date(year, month, day, 0, 0, 0, 0, l)
	d = OffsetForTime(t)

	return d, nil
}

// OffsetForTime the duration of the offset from UTC
// Inlines with
//   go build -gcflags '-m -m' timestamp.go 2>&1 |less
func OffsetForTime(t time.Time) (d time.Duration) {
	_, offset := t.Zone()

	d = time.Duration(offset) * time.Second

	return d
}

// ZoneFromHM get fixed zone from hour and minute offset
// A negative offsetH will result in a negative zone offset
//
// Inlines with
//   go build -gcflags '-m -m' timestamp.go 2>&1 |less
// and removal of fmt.Sprintf to make zone name
// location is a pointer and escapes to heap
func ZoneFromHM(offsetH, offsetM int) time.Location {
	if offsetM < 0 {
		offsetM = -offsetM
	}

	offsetSec := offsetH*60*60 + offsetM*60
	location := *time.FixedZone("FixedZone", offsetSec)

	return location
}

// OffsetHM get hours and minutes for location offset from UTC
// Avoiding math.Abs and casting allows inlining in
//
//   go build -gcflags '-m -m' timestamp.go 2>&1 |less
func OffsetHM(d time.Duration) (hours, minutes int) {
	hours = int(d.Hours())
	minutes = int(d.Minutes()) % 60

	// Ensure minutes is positive
	if minutes < 0 {
		minutes = -minutes
	}
	minutes = minutes % 60

	return hours, minutes
}

// LocationOffsetString get an offset in HHMM format based on hours and
// minutes offset from UTC.
//
// For 5 hours and 30 minutes
//  0530
//
// For -5 hours and 30 minutes
//  -0500
//
// Inlines with
//   go build -gcflags '-m -m' timestamp.go 2>&1 |less
func LocationOffsetString(d time.Duration) string {
	return locationOffsetString(d, false)
}

// LocationOffsetStringDelimited get an offset in HHMM format based on hours and
// minutes offset from UTC.
//
// For 5 hours and 30 minutes
//  05:30
//
// For -5 hours and 30 minutes
//  -05:00
//
// Inlines with
//   go build -gcflags '-m -m' timestamp.go 2>&1 |less
func LocationOffsetStringDelimited(d time.Duration) string {
	return locationOffsetString(d, true)
}

// OffsetString get an offset in HHMM format based on hours and minutes offset
// from UTC.
//
// For 5 hours and 30 minutes
//  0530
//
// For -5 hours and 30 minutes
//  -0500
func locationOffsetString(d time.Duration, delimited bool) string {
	hours, minutes := OffsetHM(d)

	if delimited == false {
		return fmt.Sprintf("%+03d%02d", hours, minutes)
	}
	return fmt.Sprintf("%+03d:%02d", hours, minutes)
}

// RangeOverTimes returns a date range function over start date to end date inclusive.
// After the end of the range, the range function returns a zero date,
// date.IsZero() is true. If zones for start and end differ an error will be
// returned and needs to be checked for before time.IsZero().
//
// Note that this function has been modified to NOT change the location for the
// start and end time to UTC. This is in keeping with the avoidance of change to
// time locations passed into function. It is the responsibility of the caller
// to set location in keeping with the intended use of the function. The
// location used could affect the day values.
//
// Sample usage assuming building a map with empty string values:
/*
	t1 := time.Now()
	t2 := t1.Add(30 * 24 * time.Hour)

	m := make(map[string]string)

	var err error
	var newTime time.Time
	for rt := timestamp.RangeOverTimes(t1, t2); ; {
		newTime, err = rt()
		if err != nil {
			// Handle when there was an error in the input times
			break
		}
		if newTime.IsZero() {
			// Handle when the day range is done
			break
		}
		v := fmt.Sprintf("%04d-%02d-%02d", newTime.Year(), newTime.Month(), newTime.Day())
		m[v] = ""
	}

	if err != nil {
		// handle error due to non-equal UTC offsets
	}

	a := make([]string, 0, len(m))
	for v := range m {
		a = append(a, v)
	}
	sort.Strings(a)
	fmt.Println("Days in range")

	for _, v := range a {
		fmt.Println("Got", v)
	}
*/
// Can't inline
func RangeOverTimes(start, end time.Time) func() (time.Time, error) {
	_, startZone := start.Zone()
	_, endZone := end.Zone()

	if startZone != endZone {
		return func() (time.Time, error) {
			return time.Time{}, errors.New("Zones for start and end differ")
		}
	}

	y, m, d := start.Date()
	start = time.Date(y, m, d, 0, 0, 0, 0, start.Location())
	y, m, d = end.Date()
	end = time.Date(y, m, d, 0, 0, 0, 0, end.Location())

	return func() (time.Time, error) {
		if start.After(end) {
			return time.Time{}, nil
		}
		date := start
		start = start.AddDate(0, 0, 1)

		return date, nil
	}
}

// TimeDateOnly get date with zero time values
//
// Note that this function has been modified to NOT change the location for the
// start and end time to UTC. This is in keeping with the avoidance of change to
// time locations passed into function. It is the responsibility of the caller
// to set location in keeping with the intended use of the function.
//
// Can inline
func TimeDateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
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
//
// Can't inline
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

// ParseUnixTS parse a timestamp directly, assuming input is some sort of UNIX
// timestamp. If the input is known to be a timestamp this will be faster than
// first trying to parse as other forms of timestamp. This function will handle
// timestamps in the form of seconds and nanoseconds delimited by a period.
//   e.g. 113621424536300000 becomes 1136214245.36300000
//
// Can't inline
func ParseUnixTS(timeStr string) (time.Time, error) {
	match := reDigits.MatchString(timeStr)

	timeStrLength := len(timeStr)
	// Only proceed if the incoming timestamp is a number with up to one
	// decimaal place. Otherwise return an error.
	if match == true {

		// Don't support timestamps less than 7 characters in length
		// to avoid strange date formats from being parsed.
		// Max would be 9999999, or Sun Apr 26 1970 17:46:39 GMT+0000
		if timeStrLength > 6 {
			toSend := timeStr
			// Break it into a format that has a period between second and
			// millisecond portions for the function.
			if timeStrLength > 10 {
				sec, nsec := timeStr[0:10], timeStr[11:timeStrLength-1]

				// Avoid heap allocation
				xfmtBuf := new(xfmt.Buffer)
				xfmtBuf.S(sec).S(".").S(nsec)
				b := xfmtBuf.Bytes()
				// l := len(b)

				toSend = bytesToString(b...)
				// toSend = string(b)
			}
			// Get seconds, nanoseconds, and error if there was a problem
			s, n, err := parseUnixTS(toSend)
			if err != nil {
				return time.Time{}, err
			}
			// If it was a unix seconds timestamp n will be zero. If it was a
			// nanoseconds timestamp there will be a nanoseconds portion that is not
			// zero.
			t := time.Unix(s, n)

			return t, nil
		}
	}
	xfmtBuf := new(xfmt.Buffer)
	// Avoid heap allocation
	xfmtBuf.S("Could not parse as UNIX timestamp ").S(timeStr)
	b := xfmtBuf.Bytes()
	// l := len(b)
	s := bytesToString(b...)
	// s := string(b)

	// return time.Time{}, fmt.Errorf("Could not parse as UNIX timestamp %s", timeStr)
	return time.Time{}, errors.New(s)
}

// ParseInUTC parse for all timestamps, defaulting to UTC, and return UTC zoned
// time
//
// Can inline
func ParseInUTC(timeStr string) (time.Time, error) {
	return parseTimestamp(timeStr, time.UTC, false)
}

// ParseISOInUTC parse limited to ISO timestamp formats and return UTC zoned time
//
// Can inline
func ParseISOInUTC(timeStr string) (time.Time, error) {
	return parseTimestamp(timeStr, time.UTC, true)
}

// ParseInLocation parse for all timestamp formats and default to location if
// there is no zone in the incoming timestamp. Return time adjusted to UTC.
//
// Can inline
func ParseInLocation(timeStr string, location *time.Location) (time.Time, error) {
	return parseTimestamp(timeStr, location, false)
}

// ParseISOInLocation parse limited to ISO timestamp formats, defaulting to
// location if there is no zone in the incoming timezone. Return time  adjusted
// to UTC.
//
// Can inline
func ParseISOInLocation(timeStr string, location *time.Location) (time.Time, error) {
	return parseTimestamp(timeStr, location, true)
}

// ParseTimestampInLocation parse timestamp, defaulting to location if there is
// no zone in the incoming timestamp, and return time ajusted to the incoming
// location.
//
// Can't inline due to use of range but it's too complex anyway.
func parseTimestamp(timeStr string, location *time.Location, isoOnly bool) (time.Time, error) {
	timeStr = strings.TrimSpace(timeStr)
	original := timeStr

	// Check to see if the incoming data is a series of digits or digits with a
	// single decimal place.

	isTS := false
	if reDigits.MatchString(timeStr) {
		// A 20060101 date will have 10 digits
		// A 20060102060708 timestamp will have 14 digits
		// A Unix timetamp will have 10 digits
		// A Unix nanosecond timestamp will have 19 digits
		l := len(timeStr)
		if l != 8 && l != 14 {
			isTS = true
		}
	}

	// Try ISO parsing first. The lexer is tolerant of some inconsistency in
	// format that is not ISO-8601 compliant, such as dashes where there should
	// be colons and a space instead of a T to separate date and time.
	if isTS == false {
		t, err := ParseISOTimestamp(timeStr, location)
		// t, err := lex.ParseInLocation([]byte(timeStr), location)
		if err == nil {
			return t, nil
		}
	}

	// If only iso format patterns should be tried leave now
	if isoOnly == true {

		xfmtBuf := new(xfmt.Buffer)
		// Avoid heap allocation
		xfmtBuf.S("Could not parse as ISO timestamp ").S(timeStr)

		b := xfmtBuf.Bytes()
		// l := len(b)
		s := bytesToString(b...)
		// s := string(b)

		return time.Time{}, errors.New(s)
	}

	if isTS == true {
		t, err := ParseUnixTS(timeStr)
		if err == nil {
			return t.In(location), nil
		}
		xfmtBuf := new(xfmt.Buffer)
		// Avoid heap allocation
		xfmtBuf.S("Could not parse as UNIX timestamp ").S(timeStr)

		b := xfmtBuf.Bytes()
		// l := len(b)
		// s := string(b)
		s := bytesToString(b...)

		return time.Time{}, errors.New(s)
	}

	// If not a unix type timestamp try alternate non-iso timestamp formats
	s := nonISOTimeFormats
	for _, format := range s {
		// If no zone in timestamp use location
		t, err := time.ParseInLocation(format, original, location)
		if err == nil {
			return t, nil
		}
	}

	xfmtBuf := new(xfmt.Buffer)
	xfmtBuf.S("Could not parse with other timestamp patterns ").S(timeStr)

	b := xfmtBuf.Bytes()
	// l := len(b)
	errMsg := bytesToString(b...)
	// errMsg := string(b)

	return time.Time{}, errors.New(errMsg)
}

// RFC7232 get format used for http headers
//   "Mon, 02 Jan 2006 15:04:05 GMT"
//
// TimeFormat is the time format to use when generating times in HTTP headers.
// It is like time.RFC1123 but hard-codes GMT as the time zone. The time being
// formatted must be in UTC for Format to generate the correct format. This is
// done in the function before the call to format.
//
// Can't inline as the format must be in GMT
func RFC7232(t time.Time) string {
	t = t.In(time.UTC)
	return t.Format(http.TimeFormat)
}

// ISO8601CompactUTC ISO-8601 timestamp with no sub seconds
//   "20060102T150405-0700"
//
// Result will be in whatever the location the incoming time is set to. If UTC
// is desired set location to time.UTC first
//
// Can inline if not converted first to UTC
func ISO8601CompactUTC(t time.Time) string {
	return t.Format("20060102T150405-0700")
}

// ISO8601CompactMsecUTC ISO-8601 timestamp with no seconds
//   "20060102T150405.000-0700"
//
// Result will be in whatever the location the incoming time is set to. If UTC
// is desired set location to time.UTC first
//
// Can inline if not converted first to UTC
func ISO8601CompactMsecUTC(t time.Time) string {
	return t.Format("20060102T150405.000-0700")
}

// ISO8601UTC ISO-8601 timestamp long format string result
//   "2006-01-02T15:04:05-07:00"
//
// Result will be in whatever the location the incoming time is set to. If UTC
// is desired set location to time.UTC first
//
// Can inline if not converted first to UTC
func ISO8601UTC(t time.Time) string {
	return t.Format("2006-01-02T15:04:05-07:00")
}

// ISO8601MsecUTC ISO-8601 longtimestamp with msec
//   "2006-01-02T15:04:05.000-07:00"
//
// Result will be in whatever the location the incoming time is set to. If UTC
// is desired set location to time.UTC first
//
// Can inline if not converted first to UTC
func ISO8601MsecUTC(t time.Time) string {
	return t.Format("2006-01-02T15:04:05.000-07:00")
}

// ISO8601InLocation timestamp long format string result in location
//   "2006-01-02T15:04:05-07:00"
//
// Result will be in whatever the location the incoming time is set to. If UTC
// is desired set location to time.UTC first
//
// Can inline if not converted first to UTC
func ISO8601InLocation(t time.Time, location *time.Location) string {
	return t.Format("2006-01-02T15:04:05-07:00")
}

// ISO8601MsecInLocation ISO-8601 longtimestamp with msec in location
//   "2006-01-02T15:04:05.000-07:00"
//
// Result will be in whatever the location the incoming time is set to. If UTC
// is desired set location to time.UTC first
//
// Can inline if not converted first to UTC
func ISO8601MsecInLocation(t time.Time, location *time.Location) string {
	return t.Format("2006-01-02T15:04:05.000-07:00")
}

// ISO8601CompactInLocation timestamp with no sub seconds in location
//   "20060102T150405-0700"
//
// Result will be in whatever the location the incoming time is set to. If UTC
// is desired set location to time.UTC first
//
// Can inline if not converted first to UTC
func ISO8601CompactInLocation(t time.Time, location *time.Location) string {
	return t.Format("20060102T150405-0700")
}

// ISO8601CompactMsecInLocation timestamp with no seconds in location
//   "20060102T150405.000-0700"
//
// Result will be in whatever the location the incoming time is set to. If UTC
// is desired set location to time.UTC first
//
// Can inline if not converted first to UTC
func ISO8601CompactMsecInLocation(t time.Time, location *time.Location) string {
	return t.Format("20060102T150405.000-0700")
}

// StartTimeIsBeforeEndTime if time 1 is before time 2 return true, else false
func StartTimeIsBeforeEndTime(t1 time.Time, t2 time.Time) bool {
	return t2.Unix()-t1.Unix() > 0
}

// ParseISOTimestamp parse an ISO timetamp iteratively. The reult will be in the
// zone for the timestamp or if there is no zone offset in the incoming
// timestamp the incoming location will bue used. It is the responsibility of
// further steps to standardize to a specific zone offset.
//
//  go build -gcflags '-m' timestamp.go
//
// Can't inline
// location leaks
func ParseISOTimestamp(timeStr string, location *time.Location) (time.Time, error) {
	// Define sections that can change.

	const maxLength int = 35
	timeStrLength := len(timeStr)

	if timeStrLength > maxLength {
		xfmtBuf := new(xfmt.Buffer)
		xfmtBuf.S("Input ").S(timeStr[0:35]).S("... length is ").D(timeStrLength).S(" and > max of ").D(maxLength)

		b := xfmtBuf.Bytes()
		s := bytesToString(b...)
		// s := string(b)

		return time.Time{}, errors.New(s)
	}

	// Needs to not be a const since it gets reassigned
	var currentSection int = 0 // value for current section

	// Define sections that are constant. Use iota since the incrementing values
	// correspond to the incremental section processing and give each const a
	// separate value.

	const (
		emptySection     int = iota // value for empty section
		yearSection                 // year - four digits
		monthSection                // month - 2 digits
		daySection                  // day - 2 digits
		hourSection                 // hour - 2 digits
		minuteSection               // minute - 2 digits
		secondSection               // second - 2 digits
		subsecondSection            // subsecond 1-9 digits
		zoneSection                 // zone +/-HHMM or Z
		afterSection                // after - when done
	)

	// Define whether offset is positive for later offset calculation.

	var offsetPositive bool = false // is offset from UTC positive

	// Define the varous part to hold values for year, month, etc. Make initial
	// size 0 and capacity enough to avoid shuffling when appending.

	var (
		yearParts      = make([]rune, 0, 4) // year digit parts
		monthParts     = make([]rune, 0, 2) // month digit parts
		dayParts       = make([]rune, 0, 2) // day digit parts
		hourParts      = make([]rune, 0, 2) // hour digit parts
		minuteParts    = make([]rune, 0, 2) // minute digit parts
		secondParts    = make([]rune, 0, 2) // second digit parts
		subsecondParts = make([]rune, 0, 9) // subsecond digit parts
		zoneParts      = make([]rune, 0, 4) // zone parts
	)

	const (
		yearMax      int = 4 // max length for year
		monthMax     int = 2 // max length for month number
		dayMax       int = 2 // max length for day number
		hourMax      int = 2 // max length for hour number
		minuteMax    int = 2 // max length for minute number
		secondMax    int = 2 // max length for second number
		subsecondMax int = 9 // max length for subsecond number
		zoneMax      int = 4 // max length for zone
	)

	var unparsed []string // string representation of unparsed runes and their positions

	var partAtMax bool     // flag indicating current part is filled
	var offsetNonzero bool // flag indicating that full zone offset is zero

	// A function to handle adding to a slice if it is not above capacity and
	// flagging when it has reached capacity. Runs same speed when inline and is
	// only used here. Return a flag indicating if a timestamp part has reached
	// its max capacity plus the modified slice to avoid issues due to
	// appending.
	var addIf = func(part []rune, add rune, max int) ([]rune, bool) {
		if len(part) < max {
			part = append(part, add)
		}
		if len(part) == max {
			return part, true
		}
		return part, false
	}

	// Loop through runes in time string and decide what to do with each.
	for i, r := range timeStr {
		orig := r
		if unicode.IsDigit(r) {
			switch currentSection {
			// Initially no section is active
			case emptySection:
				currentSection = yearSection
				yearParts, partAtMax = addIf(yearParts, r, yearMax)
				if partAtMax == true {
					currentSection = monthSection
				}
				// Year section is used until full
			case yearSection:
				yearParts, partAtMax = addIf(yearParts, r, yearMax)
				if partAtMax == true {
					currentSection = monthSection
				}
				// Month section is used until full
			case monthSection:
				monthParts, partAtMax = addIf(monthParts, r, monthMax)
				if partAtMax == true {
					currentSection = daySection
				}
				// Day section is used until full
			case daySection:
				dayParts, partAtMax = addIf(dayParts, r, dayMax)
				if partAtMax == true {
					currentSection = hourSection
				}
				// Hour section is used until full
			case hourSection:
				hourParts, partAtMax = addIf(hourParts, r, hourMax)
				if partAtMax == true {
					currentSection = minuteSection
				}
				// Minute section is used until full
			case minuteSection:
				minuteParts, partAtMax = addIf(minuteParts, r, minuteMax)
				if partAtMax == true {
					currentSection = secondSection
				}
				// Second section is used until full
			case secondSection:
				secondParts, partAtMax = addIf(secondParts, r, secondMax)
				if partAtMax == true {
					currentSection = subsecondSection
				}
				// Subsecond section is used until full
			case subsecondSection:
				subsecondParts, partAtMax = addIf(subsecondParts, r, subsecondMax)
				if partAtMax == true {
					currentSection = zoneSection
				}
				// Zone section is used until full
			case zoneSection:
				// Allow for offset calculations later to be avoided
				if r != '0' {
					offsetNonzero = true
				}
				// Add to zone
				zoneParts, partAtMax = addIf(zoneParts, r, zoneMax)
				if partAtMax == true {
					// We could exit here but we can continue to more accurately
					// report bad date parts if we allow things to continue.
					currentSection = afterSection
				}
				// Default to bad input
			default:
				xfmtBuf := new(xfmt.Buffer)
				xfmtBuf.S("'").C(orig).S("'").C('@').D(i)
				b := xfmtBuf.Bytes()
				// l := len(b)
				// s := new(string)
				s := bytesToString(b...)

				unparsed = append(unparsed, s)
			}
			// If the current section is not for subseconds skip
		} else if r == '.' {
			// There could be extraneous decimal characters.
			if currentSection != subsecondSection {
				continue
			}
			// currentSection = subsecondSection
		} else if r == '-' || r == '+' {
			// Selectively define offset possitivity
			if currentSection == subsecondSection {
				offsetPositive = (r == '+')
				currentSection = zoneSection
			}
			// Valid but not useful for parsing
		} else if unicode.ToUpper(r) == 'T' || r == ':' || r == '/' {
			continue
			// Zulu offset
		} else if unicode.ToUpper(r) == 'Z' {
			// define offset as zero for hours and minutes
			if currentSection == zoneSection || currentSection == subsecondSection {
				zoneParts = append(zoneParts, '0', '0', '0', '0')
				break
			} else {
				// Assume bad input
				xfmtBuf := new(xfmt.Buffer)
				xfmtBuf.S("'").C(orig).S("'").C('@').D(i)
				b := xfmtBuf.Bytes()
				s := bytesToString(b...)

				unparsed = append(unparsed, s)
			}
			// Ignore spaces
		} else if unicode.IsSpace(r) {
			continue
		} else {
			// Catch-all for characters not allowed
			xfmtBuf := new(xfmt.Buffer)
			xfmtBuf.S("'").C(orig).S("'").C('@').D(i)
			b := xfmtBuf.Bytes()
			s := bytesToString(b...)

			unparsed = append(unparsed, s)
		}
	}

	// If we've found characters not allocated, error.
	if len(unparsed) > 0 {
		xfmtBuf := new(xfmt.Buffer)
		xfmtBuf.S("got unparsed caracters ").S(strings.Join(unparsed, ",")).S(" in input ").S(timeStr)
		b := xfmtBuf.Bytes()
		s := bytesToString(b...)

		e := errors.New(s)

		return time.Time{}, e
	}

	zoneFound := false        // has time zone been found
	zoneLen := len(zoneParts) // length of the zone found

	// If length < 4
	if zoneLen < zoneMax {
		zoneFound = true
		// A zone with 1 or 3 characters is ambiguous
		if zoneLen == 1 || zoneLen == 3 {
			xfmtBuf := new(xfmt.Buffer)
			xfmtBuf.S("Zone is of length ").D(zoneLen).S(" wich is not enough to detect zone")
			b := xfmtBuf.Bytes()
			s := bytesToString(b...)

			return time.Time{}, errors.New(s)
			// With no zone assume UTC and set all offset characters to 0
		} else if zoneLen == 0 {
			zoneFound = false
			zoneParts = append(zoneParts, '0', '0', '0', '0')
			// Zone of length 2 needs padding to set minute offset
		} else if zoneLen == 2 {
			zoneParts = append(zoneParts, '0', '0')
		}
	} else {
		// Zone is found. Used later when setting location
		zoneFound = true
	}

	// Allow for just dates and convert to timestamp with zero valued time
	// parts. Since we are fixing it here it will pass the next tests if nothing
	// else is wrong or missing.
	if len(hourParts) == 0 && len(minuteParts) == 0 && len(secondParts) == 0 {
		hourParts = append(hourParts, '0', '0')
		minuteParts = append(minuteParts, '0', '0')
		secondParts = append(secondParts, '0', '0')
	}

	// Error if any part does not contain enough characters. This could happen
	// easily if for instance a year had 2 digits instead of 4. If this happened
	// year would take 4 digits, month would take 2, day would take 2, hour
	// would take 2, minute would take 2, and second would get none. We are thus
	// requiring that all date and time parts be fully allocated even if we
	// can't tell where the problem started.
	if len(yearParts) != yearMax {
		return time.Time{}, errors.New("Input year length is not 4")
	}
	if len(monthParts) != monthMax {
		return time.Time{}, errors.New("Input month length is not 2")
	}
	if len(dayParts) != dayMax {
		return time.Time{}, errors.New("Input day length is not 2")
	}
	if len(hourParts) != hourMax {
		return time.Time{}, errors.New("Input hour length is not 2")
	}
	if len(minuteParts) != minuteMax {
		return time.Time{}, errors.New("Input minute length is not 2")
	}
	if len(secondParts) != secondMax {
		return time.Time{}, errors.New("Input second length is not 2")
	}

	// We already only put digits into the parts so Atoi should be fine in all
	// cases. The problem would have been with an incorrect number of digits in
	// a part, which would have been caught above.

	// Get year int value from yearParts rune slice
	// Should not error since only digits were place in slice
	y, err := strconv.Atoi(runesToString(yearParts...))
	// y, err := strconv.Atoi(string(yearParts))
	if err != nil {
		return time.Time{}, err
	}
	// Get month int value from monthParts rune slice
	// Should not error since only digits were place in slice
	m, err := strconv.Atoi(runesToString(monthParts...))
	// m, err := strconv.Atoi(string(monthParts))
	if err != nil {
		return time.Time{}, err
	}
	// Get day int value from dayParts rune slice
	// Should not error since only digits were place in slice
	d, err := strconv.Atoi(runesToString(dayParts...))
	// d, err := strconv.Atoi(string(dayParts))
	if err != nil {
		return time.Time{}, err
	}
	// Get hour int value from hourParts rune slice
	// Should not error since only digits were place in slice
	h, err := strconv.Atoi(runesToString(hourParts...))
	// h, err := strconv.Atoi(string(hourParts))
	if err != nil {
		return time.Time{}, err
	}
	// Get minute int value from minParts rune slice
	// Should not error since only digits were place in slice
	mn, err := strconv.Atoi(runesToString(minuteParts...))
	// mn, err := strconv.Atoi(string(minuteParts))
	if err != nil {
		return time.Time{}, err
	}
	// Get second int value from secondParts rune slice
	// Should not error since only digits were place in slice
	s, err := strconv.Atoi(runesToString(secondParts...))
	// s, err := strconv.Atoi(string(secondParts))
	if err != nil {
		return time.Time{}, err
	}

	var subseconds int = 0 // default subsecond value is 0

	subsecondLen := len(subsecondParts)

	// Handle subseconds if that slice is nonempty
	// There would have been an error if the length of subsecond parts was
	// greater than subsecondMax
	if subsecondLen > 0 {
		subseconds, err = strconv.Atoi(runesToString(subsecondParts...))
		// subseconds, err = strconv.Atoi(string(subsecondParts))
		if err != nil {
			return time.Time{}, err
		}
		// Calculate subseconds in terms of nanosecond if the length is less
		// than the full length for nanoseconds since that is what the time.Date
		// function is expecting.
		if subsecondLen < subsecondMax {
			// 10^ whatever extra decimal place count is missing from 9
			subseconds = int(
				float64(subseconds) *
					(math.Pow(10, (float64(subsecondMax) - float64(subsecondLen)))))
		}
	}

	// NOTE:
	// We have already ensured that all parts have the correct number of digits.
	// don't worry about ensuring that the values of months, days, hours,
	// minutes, etc. are being too large within their digit span. The Go time
	// package increments higher values as appropriate. For instance a value of
	// 60 seconds would force an addition to the minute and all the way up to
	// the year for 2020-12-31T59:59:60-0000.

	// Create timestamp based on parts with proper offsset

	// If no zone was found in scan use default location
	if zoneFound == false {
		return time.Date(y, time.Month(m), d, h, mn, s, subseconds, location), nil
	}

	var offsetM int = 0
	var offsetH int = 0

	if offsetNonzero {
		// Evaluate minute offset from the timestamp value
		// Should not error since only digits were place in slice
		offsetM, err = strconv.Atoi(runesToString(zoneParts[2:]...))
		// offsetM, err = strconv.Atoi(string(zoneParts[2:]))
		if err != nil {
			return time.Time{}, err
		}

		// Evaluate hour offset from the timestamp value
		// Should not error since only digits were place in slice
		offsetH, err = strconv.Atoi(runesToString(zoneParts[0:2]...))
		// offsetH, err = strconv.Atoi(string(zoneParts[0:2]))
		if err != nil {
			return time.Time{}, err
		}
	} else {
		return time.Date(y, time.Month(m), d, h, mn, s, subseconds, time.UTC), nil
	}

	var offsetSec int

	// The +/- in the timestamp was used to set offsetPositive
	switch offsetPositive {
	case true:
		offsetSec = offsetH*60*60 + offsetM*60
	default:
		offsetSec = -offsetH*60*60 + offsetM*60
	}

	var zone *time.Location

	// Using a cache for locations saves 3 allocations and over 170 bytes in benchmark
	if val, ok := cachedZones[offsetSec]; ok {
		zone = val
		// Given that zones are in at most 15 minute increments and can be
		// positive or negative there should only be so many.
		// https://time.is/time_zones
		// There are currently 37 observed UTC offsets in the world (38 when
		// Iran is on standard time).
		// Allow up to 50.
		if len(cachedZones) > 50 {
			cachedZones = make(map[int]*time.Location)
		}
	} else {
		zone = time.FixedZone("FixedZone", offsetSec)
		cachedZones[offsetSec] = zone
	}

	return time.Date(y, time.Month(m), d, h, mn, s, subseconds, zone), nil
}
