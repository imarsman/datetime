package timestamp

import (
	"errors"
	"net/http"
	"time"

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
func OffsetForLocation(year int, month time.Month, day int, locationName string) (d time.Duration, err error) {
	l, err := time.LoadLocation(locationName)
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

// ZoneFromHM get fixed zone from hour and minute offset
// A negative offsetH will result in a negative zone offset
func ZoneFromHM(offsetH, offsetM int) *time.Location {
	if offsetM < 0 {
		offsetM = -offsetM
	}

	// Must be passed a value equivalent to total seconds for hours and minutes
	location := LocationFromOffset(offsetH*60*60 + offsetM*60)

	return location
}

// OffsetHM get hours and minutes for location offset from UTC
// Avoiding math.Abs and casting allows inlining in
func OffsetHM(d time.Duration) (offsetH, offsetM int) {
	offsetH = int(d.Hours())
	offsetM = int(d.Minutes()) % 60

	// Ensure minutes is positive
	if offsetM < 0 {
		offsetM = -offsetM
	}
	offsetM = offsetM % 60

	return offsetH, offsetM
}

// LocationOffsetString get an offset in HHMM format based on hours and
// minutes offset from UTC.
//
// For 5 hours and 30 minutes
//  0530
//
// For -5 hours and 30 minutes
//  -0500
func LocationOffsetString(d time.Duration) (string, error) {
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
func LocationOffsetStringDelimited(d time.Duration) (string, error) {
	return locationOffsetString(d, true)
}

// TwoDigitOffset get digit offset for hours and minutes. This is designed
// solely to help with calculating offset strings for timestamps without using
// fmt.Sprintf, which causes allocations.
func TwoDigitOffset(in int, addPrefix bool) (string, error) {
	if in > 99 || in < -99 {
		return "", errors.New("Out of range")
	}

	var prefix rune = '+'

	if in < 0 {
		prefix = '-'
		in = -in
	}

	var fr rune = rune('0' + int(in/10))
	var lr rune = rune('0' + in%10)

	if addPrefix == true {
		return RunesToString(prefix, fr, lr), nil
	}
	return RunesToString(fr, lr), nil
}

// OffsetString get an offset in HHMM format based on hours and minutes offset
// from UTC.
//
// For 5 hours and 30 minutes
//  0530
//
// For -5 hours and 30 minutes
//  -0500
func locationOffsetString(d time.Duration, delimited bool) (string, error) {
	offsetH, offsetM := OffsetHM(d)

	xfmt := new(xfmt.Buffer)

	h, err := TwoDigitOffset(offsetH, true)
	if err != nil {
		return "", err
	}
	xfmt.S(h)
	if delimited == true {
		xfmt.C(':')
	}
	m, err := TwoDigitOffset(offsetM, false)
	if err != nil {
		return "", err
	}
	xfmt.S(m)

	return BytesToString(xfmt.Bytes()...), nil
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
func ISO8601CompactUTC(t time.Time) string {
	return t.Format("20060102T150405-0700")
}

// ISO8601CompactMsecUTC ISO-8601 timestamp with no seconds
//   "20060102T150405.000-0700"
//
// Result will be in whatever the location the incoming time is set to. If UTC
// is desired set location to time.UTC first
func ISO8601CompactMsecUTC(t time.Time) string {
	return t.Format("20060102T150405.000-0700")
}

// ISO8601UTC ISO-8601 timestamp long format string result
//   "2006-01-02T15:04:05-07:00"
//
// Result will be in whatever the location the incoming time is set to. If UTC
// is desired set location to time.UTC first
func ISO8601UTC(t time.Time) string {
	return t.Format("2006-01-02T15:04:05-07:00")
}

// ISO8601MsecUTC ISO-8601 longtimestamp with msec
//   "2006-01-02T15:04:05.000-07:00"
//
// Result will be in whatever the location the incoming time is set to. If UTC
// is desired set location to time.UTC first
func ISO8601MsecUTC(t time.Time) string {
	return t.Format("2006-01-02T15:04:05.000-07:00")
}

// ISO8601InLocation timestamp long format string result in location
//   "2006-01-02T15:04:05-07:00"
//
// Result will be in whatever the location the incoming time is set to. If UTC
// is desired set location to time.UTC first
func ISO8601InLocation(t time.Time, location *time.Location) string {
	return t.Format("2006-01-02T15:04:05-07:00")
}

// ISO8601MsecInLocation ISO-8601 longtimestamp with msec in location
//   "2006-01-02T15:04:05.000-07:00"
//
// Result will be in whatever the location the incoming time is set to. If UTC
// is desired set location to time.UTC first
func ISO8601MsecInLocation(t time.Time, location *time.Location) string {
	return t.Format("2006-01-02T15:04:05.000-07:00")
}

// ISO8601CompactInLocation timestamp with no sub seconds in location
//   "20060102T150405-0700"
//
// Result will be in whatever the location the incoming time is set to. If UTC
// is desired set location to time.UTC first
func ISO8601CompactInLocation(t time.Time, location *time.Location) string {
	return t.Format("20060102T150405-0700")
}

// ISO8601CompactMsecInLocation timestamp with no seconds in location
//   "20060102T150405.000-0700"
//
// Result will be in whatever the location the incoming time is set to. If UTC
// is desired set location to time.UTC first
func ISO8601CompactMsecInLocation(t time.Time, location *time.Location) string {
	return t.Format("20060102T150405.000-0700")
}

// StartTimeIsBeforeEndTime if time 1 is before time 2 return true, else false
func StartTimeIsBeforeEndTime(t1 time.Time, t2 time.Time) bool {
	return t2.Unix()-t1.Unix() > 0
}
