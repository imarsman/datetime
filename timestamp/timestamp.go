package timestamp

import (
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
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

var reDigits *regexp.Regexp

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

var timeFormats = []string{}

func init() {
	reDigits = regexp.MustCompile(`^\d+\.?\d+$`)
	timeFormats = append(timeFormats, nonISOTimeFormats...)
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
// If the zone is not recognized in Go's tzdata database an error will be returned.
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
func OffsetForTime(t time.Time) (d time.Duration) {
	_, offset := t.Zone()

	d = time.Duration(offset) * time.Second

	return d
}

// OffsetFromHM get fixed zone from hour and minute offset
func OffsetFromHM(offsetH, offsetM int) *time.Location {
	offsetSec := offsetH*60*60 + offsetM*60
	o := time.FixedZone("FIXED", offsetSec)

	return o
}

// OffsetHM get hours and minutes for location offset from UTC
func OffsetHM(d time.Duration) (hours, minutes int) {
	hours = int(d.Hours())
	minutes = int(math.Abs(float64(int(d.Minutes()) % 60)))

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

// ParseUnixTS parse a timestamp directly, assuming input is some sort of UNIX
// timestamp. If the input is known to be a timestamp this will be faster than
// first trying to parse as other forms of timestamp. This function will handle
// timestamps in the form of seconds and nanoseconds delimited by a period.
//   e.g. 113621424536300000 becomes 1136214245.36300000
func ParseUnixTS(timeStr string) (time.Time, error) {
	match := reDigits.MatchString(timeStr)

	// Only proceed if the incoming timestamp is a number with up to one
	// decimaal place. Otherwise return an error.
	if match == true {

		// Don't support timestamps less than 7 characters in length
		// to avoid strange date formats from being parsed.
		// Max would be 9999999, or Sun Apr 26 1970 17:46:39 GMT+0000
		if len(timeStr) > 6 {
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
			t := time.Unix(s, n)

			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("Could not parse as UNIX timestamp %s", timeStr)
}

// ParseInUTC parse for all timestamps, defaulting to UTC, and return UTC zoned time
func ParseInUTC(timeStr string) (time.Time, error) {
	return parseTimestamp(timeStr, time.UTC, false)
}

// ParseISOInUTC parse limited to ISO timestamp formats and return UTC zoned time
func ParseISOInUTC(timeStr string) (time.Time, error) {
	return parseTimestamp(timeStr, time.UTC, true)
}

// ParseInLocation parse for all timestamp formats and default to location if
// there is no zone in the incoming timestamp. Return time adjusted to UTC.
func ParseInLocation(timeStr string, location *time.Location) (time.Time, error) {
	return parseTimestamp(timeStr, location, false)
}

// ParseISOInLocation parse limited to ISO timestamp formats, defaulting to
// location if there is no zone in the incoming timezone. Return time  adjusted
// to UTC.
func ParseISOInLocation(timeStr string, location *time.Location) (time.Time, error) {
	return parseTimestamp(timeStr, location, true)
}

// ParseTimestampInLocation parse timestamp, defaulting to location if there is
// no zone in the incoming timestamp, and return time ajusted to the incoming
// location.
func parseTimestamp(timeStr string, location *time.Location, isoOnly bool) (time.Time, error) {
	timeStr = strings.TrimSpace(timeStr)
	original := timeStr

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

	if isTS == false {
		t, err := ParseISOTimestamp(timeStr, location)
		// t, err := lex.ParseInLocation([]byte(timeStr), location)
		if err == nil {
			return t, nil
		}
	}

	t, err := ParseUnixTS(timeStr)
	if err == nil {
		return t.In(location), nil
	}
	// Check to see if the incoming data is a series of digits or digits with a
	// single decimal place.

	// Try ISO parsing first. The lexer is tolerant of some inconsistency in
	// format that is not ISO-8601 compliant, such as dashes where there should
	// be colons and a space instead of a T to separate date and time.

	// If only iso format patterns should be tried leave now
	if isoOnly == true {
		return time.Time{}, err
	}

	// If not a unix type timestamp try alternate non-iso timestamp formats
	s := nonISOTimeFormats
	for _, format := range s {
		// If no zone in timestamp use location
		t, err := time.ParseInLocation(format, original, location)
		if err == nil {
			// t = t.In(time.UTC)
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("Could not parse as UNIX timestamp %s", timeStr)
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

// ISO8601CompactUTC ISO-8601 timestamp with no sub seconds
//   "20060102T150405-0700"
func ISO8601CompactUTC(t time.Time) string {
	t = t.In(time.UTC)

	return t.Format("20060102T150405-0700")
}

// ISO8601CompactMsecUTC ISO-8601 timestamp with no seconds
//   "20060102T150405.000-0700"
func ISO8601CompactMsecUTC(t time.Time) string {
	t = t.In(time.UTC)

	return t.Format("20060102T150405.000-0700")
}

// ISO8601UTC ISO-8601 timestamp long format string result
//   "2006-01-02T15:04:05-07:00"
func ISO8601UTC(t time.Time) string {
	t = t.In(time.UTC)

	return t.Format("2006-01-02T15:04:05-07:00")
}

// ISO8601MsecUTC ISO-8601 longtimestamp with msec
//   "2006-01-02T15:04:05.000-07:00"
func ISO8601MsecUTC(t time.Time) string {
	t = t.In(time.UTC)

	return t.Format("2006-01-02T15:04:05.000-07:00")
}

// ISO8601InLocation timestamp long format string result in location
//   "2006-01-02T15:04:05-07:00"
func ISO8601InLocation(t time.Time, location *time.Location) string {
	t = t.In(location)

	return t.Format("2006-01-02T15:04:05-07:00")
}

// ISO8601MsecInLocation ISO-8601 longtimestamp with msec in location
//   "2006-01-02T15:04:05.000-07:00"
func ISO8601MsecInLocation(t time.Time, location *time.Location) string {
	t = t.In(location)

	return t.Format("2006-01-02T15:04:05.000-07:00")
}

// ISO8601CompactInLocation timestamp with no sub seconds in location
//   "20060102T150405-0700"
func ISO8601CompactInLocation(t time.Time, location *time.Location) string {
	t = t.In(location)

	return t.Format("20060102T150405-0700")
}

// ISO8601CompactMsecInLocation timestamp with no seconds in location
//   "20060102T150405.000-0700"
func ISO8601CompactMsecInLocation(t time.Time, location *time.Location) string {
	t = t.In(location)

	return t.Format("20060102T150405.000-0700")
}

// StartTimeIsBeforeEndTime if time 1 is before time 2 return true, else false
func StartTimeIsBeforeEndTime(t1 time.Time, t2 time.Time) bool {
	// t1 = t1.In(time.UTC)
	// t2 = t2.In(time.UTC)

	return t2.Unix()-t1.Unix() > 0
}

// ParseISOTimestamp parse an ISO timetamp iteratively. The reult will be in the
// zone for the timestamp or if there is no zone offset in the incoming
// timestamp the incoming location will bue used. It is the responsibility of
// further steps to standardize to a specific zone offset.
func ParseISOTimestamp(timeStr string, location *time.Location) (time.Time, error) {
	var t time.Time

	// Define sections that can change.
	var currentSection int = 0
	var emptySection int = 0

	// Defined sections. Use iota since the incrementing values correspond to
	// the incremental section processing and give each const a separate value.
	const (
		yearSection = iota + 1
		monthSection
		daySection
		hourSection
		minuteSection
		secondSection
		subsecondSection
		zoneSection
		afterSection
	)

	// Define required lengths for sections. Used quite a bit to both stop
	// allocating to a timestamp part and to ensure that parts are fully
	// allocated.
	const (
		yearMax      int = 4
		monthMax     int = 2
		dayMax       int = 2
		hourMax      int = 2
		minuteMax    int = 2
		secondMax    int = 2
		subsecondMax int = 9
		zoneMax      int = 4
	)

	// Define whether offset is positive for later offset calculation.
	var offsetPositive bool = false

	// Define the varous part to hold values for year, month, etc.
	var yearParts, monthParts, dayParts []rune
	var hourParts, minuteParts, secondParts []rune
	var subsecondParts []rune
	var zoneParts []rune

	var unparsed string

	// A function to handle adding to a slice if it is not above capacity and
	// flagging when it has reached capacity. Runs same speed when inline and is
	// only used here. Return a flag indicating if a timestamp part has reached
	// its max capacity.
	var addIf = func(part *[]rune, add rune, max int) bool {
		if len(*part) < max {
			*part = append(*part, add)
		}
		if len(*part) == max {
			return true
		}
		return false
	}

	// Loop through runes in time string and decide what to do with each.
	for _, r := range timeStr {
		orig := r
		if unicode.IsDigit(r) {
			switch currentSection {
			case emptySection:
				currentSection = yearSection
				done := addIf(&yearParts, r, yearMax)
				if done == true {
					currentSection = monthSection
				}
			case yearSection:
				done := addIf(&yearParts, r, yearMax)
				if done == true {
					currentSection = monthSection
				}
			case monthSection:
				done := addIf(&monthParts, r, monthMax)
				if done == true {
					currentSection = daySection
				}
			case daySection:
				done := addIf(&dayParts, r, dayMax)
				if done == true {
					currentSection = hourSection
				}
			case hourSection:
				done := addIf(&hourParts, r, hourMax)
				if done == true {
					currentSection = minuteSection
				}
			case minuteSection:
				done := addIf(&minuteParts, r, minuteMax)
				if done == true {
					currentSection = secondSection
				}
			case secondSection:
				done := addIf(&secondParts, r, secondMax)
				if done == true {
					currentSection = subsecondSection
				}
			case subsecondSection:
				done := addIf(&subsecondParts, r, subsecondMax)
				if done == true {
					currentSection = zoneSection
				}
			case zoneSection:
				done := addIf(&zoneParts, r, zoneMax)
				if done == true {
					// We could exit here but we can continue to more accurately
					// report bad date parts if we allow things to continue.
					currentSection = afterSection
				}
			default:
				unparsed = unparsed + string(orig)
			}
		} else if r == '.' {
			// There could be extraneous decimal characters.
			if currentSection != subsecondSection {
				continue
			}
			currentSection = subsecondSection
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
			if currentSection == zoneSection || currentSection == subsecondSection {
				// define offset as zero for hours and minutes
				zoneParts = []rune{'0', '0', '0', '0'}
				break
			} else {
				unparsed = unparsed + string(orig)
			}
			// Ignore spaces
		} else if unicode.IsSpace(r) {
			continue
		} else {
			// We haven't dealt with valid characters so prepare for erroor
			// notAllocated = notAllocated + 1
			unparsed = unparsed + string(orig)
		}
	}

	// If we've found characters not allocated, error.
	if len(unparsed) > 0 {
		return time.Time{}, fmt.Errorf(
			"got unparsed caracters %s in input %s",
			strings.Join(strings.Split(unparsed, ""), ","), timeStr)
	}

	zoneFound := false
	zoneLen := len(zoneParts)

	// If length < 4
	if zoneLen < zoneMax {
		zoneFound = true
		if zoneLen == 1 || zoneLen == 3 {
			return time.Time{}, fmt.Errorf("Zone is of length %d wich is not enough to detect zone", len(zoneParts))
		} else if zoneLen == 0 {
			zoneFound = false
			zoneParts = append(zoneParts, '0', '0', '0', '0')
		} else if zoneLen == 2 {
			zoneParts = append(zoneParts, '0', '0')
		}
	} else {
		zoneFound = true
	}

	// Allow for just dates and convert to timestamp with zero valued time
	// parts. Since we are fixing it here it will pass the next tests if nothing
	// else is wrong or missing.
	if len(hourParts) == 0 && len(minuteParts) == 0 && len(secondParts) == 0 {
		for i := 0; i < 2; i++ {
			hourParts = append(hourParts, '0')
			minuteParts = append(minuteParts, '0')
			secondParts = append(secondParts, '0')
		}
	}

	// Error if any part does not contain enough characters. This could happen
	// easily if for instance a year had 2 digits instead of 4. If this happened
	// year would take 4 digits, month would take 2, day would take 2, hour
	// would take 2, minute would take 2, and second would get none. We are thus
	// requiring that all date and time parts be fully allocated event if we
	// can't tell where the problem started.
	if len(yearParts) != yearMax {
		return time.Time{}, fmt.Errorf("Input %s has year length %d needs %d", timeStr, len(yearParts), yearMax)
	}
	if len(monthParts) != monthMax {
		return time.Time{}, fmt.Errorf("Input %s has month length %d needs %d", timeStr, len(monthParts), monthMax)
	}
	if len(dayParts) != dayMax {
		return time.Time{}, fmt.Errorf("Input %s has day length %d needs %d", timeStr, len(dayParts), dayMax)
	}
	if len(hourParts) != hourMax {
		return time.Time{}, fmt.Errorf("Input %s has hour length %d needs %d", timeStr, len(hourParts), hourMax)
	}
	if len(minuteParts) != minuteMax {
		return time.Time{}, fmt.Errorf("Input %s has minute length %d needs %d", timeStr, len(minuteParts), minuteMax)
	}
	if len(secondParts) != secondMax {
		return time.Time{}, fmt.Errorf("Input %s has second length %d needs %d", timeStr, len(secondParts), secondMax)
	}

	// We already only put digits into the parts so Atoi should be fine in all
	// cases. The problem would have been with an incorrect number of digits in
	// a part, which would have been caught above.

	// Get year int value from yearParts rune slice
	y, err := strconv.Atoi(string(yearParts))
	if err != nil {
		return time.Time{}, err
	}
	// Get month int value from monthParts rune slice
	m, err := strconv.Atoi(string(monthParts))
	if err != nil {
		return time.Time{}, err
	}
	// Get day int value from dayParts rune slice
	d, err := strconv.Atoi(string(dayParts))
	if err != nil {
		return time.Time{}, err
	}
	// Get hour int value from hourParts rune slice
	h, err := strconv.Atoi(string(hourParts))
	if err != nil {
		return time.Time{}, err
	}
	// Get minute int value from minParts rune slice
	mn, err := strconv.Atoi(string(minuteParts))
	if err != nil {
		return time.Time{}, err
	}
	// Get second int value from secondParts rune slice
	s, err := strconv.Atoi(string(secondParts))
	if err != nil {
		return time.Time{}, err
	}

	// Handle subseconds if that slice is nonempty
	var subsecond int = 0
	if len(subsecondParts) > 0 {
		subsecond, err = strconv.Atoi(string(subsecondParts))
		if err != nil {
			return time.Time{}, err
		}
		// Calculate subseconds in terms of nanosecond if the length is less
		// than the full length for nanoseconds since that is what the time.Date
		// function is expecting.
		if len(subsecondParts) < subsecondMax {
			// 10^ whatever extra decimal place count is missing from 9
			subsecond = int(float64(subsecond) * (math.Pow(10, (float64(subsecondMax) - float64(len(subsecondParts))))))
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
		t = time.Date(y, time.Month(m), d, h, mn, s, int(subsecond), location)
	} else {
		// Evaluate offset from the timestamp value
		offsetH, err := strconv.Atoi(string(zoneParts[0:2]))
		if err != nil {
			return time.Time{}, err
		}

		offsetM, err := strconv.Atoi(string(zoneParts[2:]))
		if err != nil {
			return time.Time{}, err
		}

		// Evaluate seconds of offset from UTC
		// offsetSec := offsetH*60*60 + offsetM*60

		// If offset is 00:00 use UTC
		if offsetH == 0 && offsetM == 0 {
			t = time.Date(y, time.Month(m), d, h, mn, s, int(subsecond), time.UTC)
		} else {
			// The +/- in the timestamp was used to set offsetPositive
			if offsetPositive == true {
				// Create fixed zone using seconds offset
				fixedZone := OffsetFromHM(offsetH, offsetM)
				// fixedZone := time.FixedZone("FIXEDZONE", offsetSec)
				// Offset by fixed offset zone
				t = time.Date(y, time.Month(m), d, h, mn, s, int(subsecond), fixedZone)
			} else {
				// Create fixed zone using negative seconds offset
				fixedZone := OffsetFromHM(-offsetH, offsetM)
				// fixedZone := time.FixedZone("FIXEDZONE", -offsetSec)
				// Offset by fixed offset zone
				t = time.Date(y, time.Month(m), d, h, mn, s, int(subsecond), fixedZone)
			}
		}
	}

	// Note that the time returned will have the zone offset of the incoming
	// timestamp or the default passed in if there was no offset in the timestamp.
	return t, nil
}
