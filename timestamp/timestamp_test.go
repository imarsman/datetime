package timestamp_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/imarsman/datetime/timestamp"
	"github.com/imarsman/datetime/timestamp/lex"
	"github.com/matryer/is"
)

func runningtime(s string) (string, time.Time) {
	fmt.Println("Start:	", s)
	return s, time.Now()
}

func track(s string, startTime time.Time) {
	endTime := time.Now()
	fmt.Println("End:	", s, "took", endTime.Sub(startTime))
}

func execute() {
	defer track(runningtime("execute"))
	time.Sleep(3 * time.Second)
}

// checkDate for use in parse checking. If the timestamp has no zone use the
// location passed in. Compare expected to the parssed value in the location
// pased in.
func checkDate(t *testing.T, input string, location *time.Location) (
	got string, parsed string, calculated time.Time, calculatedOffset time.Duration, defaultOffset time.Duration) {
	is := is.New(t)

	calculated, err := timestamp.ParseInLocation(input, location)
	if err != nil {
		t.Logf("Got error on parsing %v", err)
	}
	is.NoErr(err)
	// t.Log(calculated)

	parsed = timestamp.ISO8601MsecInLocation(calculated, calculated.Location())

	calculatedHours, calculatedMinutes, _ := timestamp.OffsetForTime(calculated)
	calculatedOffset = timestamp.OffsetDuration(calculatedHours, calculatedMinutes)

	inLoc := calculated.In(location)
	defaultHours, defaultMinutes, _ := timestamp.OffsetForLocation(
		inLoc.Year(), inLoc.Month(), inLoc.Day(), inLoc.Location().String())
	defaultOffset = timestamp.OffsetDuration(defaultHours, defaultMinutes)

	fmt.Printf("Input %s, calculated %v calculatedOffset %v defaultOffset %v\n", input, parsed, calculatedOffset, defaultOffset)

	return input, parsed, calculated, calculatedOffset, defaultOffset
}

// TestParse parse all patterns and compare with expected values. Input
// tmestamps are parsed and have the timestamp value applied if available,
// otherwise the passed in location is used.
func TestParse(t *testing.T) {
	is := is.New(t)

	mst, err := time.LoadLocation("MST")
	is.NoErr(err)

	test, err := time.LoadLocation("EST")
	is.NoErr(err)
	est, err := time.LoadLocation("America/Toronto")
	is.NoErr(err)
	is.True(test != est)

	utcST, err := timestamp.ParseInUTC("2006-01-02T15:04:05.000+00:00")
	is.NoErr(err)
	utcST = utcST.In(time.UTC)
	offsetH, offsetM, err := timestamp.OffsetForLocation(utcST.Year(), utcST.Month(), utcST.Day(), time.UTC.String())
	is.NoErr(err)
	utcOffset := timestamp.OffsetDuration(offsetH, offsetM)

	estST, err := timestamp.ParseInUTC("2006-01-02T15:04:05.000+00:00")
	is.NoErr(err)
	estST = estST.In(est)
	offsetH, offsetM, err = timestamp.OffsetForLocation(estST.Year(), estST.Month(), estST.Day(), est.String())
	is.NoErr(err)
	estSTOffset := timestamp.OffsetDuration(offsetH, offsetM)

	estDST, err := timestamp.ParseInUTC("2006-07-02T15:04:05.000+00:00")
	is.NoErr(err)
	estDST = estDST.In(est)
	offsetH, offsetM, err = timestamp.OffsetForLocation(estDST.Year(), estDST.Month(), estDST.Day(), est.String())
	is.NoErr(err)
	estDSTOffset := timestamp.OffsetDuration(offsetH, offsetM)

	mstST, err := timestamp.ParseInUTC("2006-07-02T15:04:05.000+00:00")
	is.NoErr(err)
	mstST = mstST.In(mst)
	offsetH, offsetM, err = timestamp.OffsetForLocation(mstST.Year(), mstST.Month(), mstST.Day(), mst.String())
	is.NoErr(err)
	mstOffset := timestamp.OffsetDuration(offsetH, offsetM)

	start := time.Now()

	// It is possible to have a string which is just digits that will be parsed
	// as a timestamp, incorrectly.
	_, err = timestamp.ParseInUTC("2006010247")
	is.NoErr(err)

	// Get a unix timestamp we should not parse
	_, err = timestamp.ParseInUTC("1")
	is.True(err != nil) // Error should be true

	// Get time value from parsed reference time
	unixBase, err := timestamp.ParseInUTC("2006-01-02T15:04:05.000+00:00")
	is.NoErr(err)

	var sent, tStr string
	var res time.Time
	var resOffset, defOffset time.Duration
	is.True(sent == "")
	is.True(tStr == "")
	is.True(res == time.Time{})

	// This could be parsed as ISO-8601
	sent, tStr, res, resOffset, defOffset = checkDate(t, fmt.Sprint(unixBase.UnixNano()), time.UTC)
	is.Equal(resOffset, defOffset)
	// UTC timestamp stayed UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, fmt.Sprint(unixBase.Unix()), time.UTC)
	is.Equal(resOffset, defOffset)
	// UTC timestamp converted to MST
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, fmt.Sprint(unixBase.Unix()), mst)
	is.Equal(resOffset, defOffset)
	is.Equal(mstOffset, resOffset)

	// Handling of a leap second. The minute value rolls over to the next minute
	// when the second value is 60.
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150460-0700", time.UTC)
	is.True(resOffset != defOffset)

	// Should be offset corresponding to Mountain Standard Time
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150460-0700", mst)
	is.Equal(resOffset, defOffset)
	is.Equal(mstOffset, resOffset)

	// Should be offset corresponding to Eastern Standard time
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102240000-0500", est)
	is.Equal(resOffset, defOffset)
	// Result offset is EST
	is.Equal(estSTOffset, resOffset)

	// Should be offset corresponding to Estern Daylight Savings Time
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060702240000-0400", est)
	// Is DST
	is.Equal(estDSTOffset, resOffset)

	// Short ISO-8601 timestamps with numerical zone offsets
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405-0700", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405-0700", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405-07", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000+0000", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000-0000", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000-0700", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000+0700", time.UTC)
	is.True(resOffset != defOffset)
	// Result is not MST
	is.True(mstOffset != resOffset)
	// Result is UTC+07:00
	d, err := time.ParseDuration("7h")
	is.NoErr(err)
	is.True(resOffset == d)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000000-0700", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000000+0330", time.UTC)
	is.True(resOffset != defOffset)
	// Result is UTC+03:30
	d, err = time.ParseDuration("3h30m")
	is.NoErr(err)
	is.True(resOffset == d)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.999999999-0700", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	// Long ISO-8601 timestamps with numerical zone offsets
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05-07:00", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05-07", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05.000-07:00", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05.000-07", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05.000000-07:00", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05.001000-07", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05.001000000-07:00", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05.999999999-07", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	// Short  ISO-8601 timestamps with UTC zone offsets
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000000Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000000000Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.001000000Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000100000Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.999999999Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	// Long date time with UTC zone offsets
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05.000Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05.000000Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05.999999999Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	// Just in case
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 15-04-05", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102150405", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	// Short ISO-8601 timestamps with no zone offset. Assume UTC.
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000000", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.999999999", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	// SQL
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 22:04:05", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	// MST is -0700 from UTC, so UTC will be 7 hours ahead
	// Should be offset corresponding to Mountain Standard Time
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 22:04:05", mst)
	is.Equal(resOffset, defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	// MST is -0500 from UTC, so UTC will be 5 hours ahead
	// Should be offset corresponding to Eastern Daylight Savings Time
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 22:04:05", est)
	is.Equal(resOffset, defOffset)
	// Result is EST
	is.Equal(estSTOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 22:04:05 -00", time.UTC)
	is.Equal(resOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	// The input has a timestamp so EST will not be applied
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 22:04:05 -00", est)
	is.True(resOffset != defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 22:04:05 +00", time.UTC)
	is.Equal(resOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 22:04:05 -00:00", time.UTC)
	is.Equal(resOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 22:04:05 +00:00", time.UTC)
	is.Equal(resOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	// Hopefully less likely to be found. Assume UTC.
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102", time.UTC)
	is.Equal(resOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02", time.UTC)
	is.Equal(resOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006/01/02", time.UTC)
	is.Equal(resOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "01/02/2006", time.UTC)
	is.Equal(resOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "1/2/2006", time.UTC)
	is.Equal(resOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	// Weird ones with improper separators
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15-04-05-0700", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(resOffset, mstOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15-04-05.000-0700", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(resOffset, mstOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15-04-05.000000-0700", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(resOffset, mstOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15-04-05.999999999-0700", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(resOffset, mstOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15-04-05-07:00", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(resOffset, mstOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15-04-05.000-07:00", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(resOffset, mstOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15-04-05.000000-07:00", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(resOffset, mstOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15-04-05.999999999-07:00", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(resOffset, mstOffset)

	// RFC7232 - used in HTTP protocol
	sent, tStr, res, resOffset, defOffset = checkDate(t, "Mon, 02 Jan 2006 15:04:05 GMT", time.UTC)
	is.Equal(resOffset, defOffset)
	// Result is MST
	is.Equal(utcOffset, resOffset)

	// RFC1123Z
	sent, tStr, res, resOffset, defOffset = checkDate(t, "Mon, 02 Jan 2006 15:04:05 -0700", mst)
	is.Equal(mstOffset, resOffset)
	// Result is MST
	is.Equal(resOffset, defOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "Mon, 02 Jan 2006 15:04:05", est)
	is.Equal(estSTOffset, resOffset)
	// Result is EST
	is.Equal(resOffset, defOffset)

	// RFC822Z
	sent, tStr, res, resOffset, defOffset = checkDate(t, "02 Jan 06 15:04 -0700", time.UTC)
	is.Equal(mstOffset, resOffset)
	// Result is MST
	is.True(resOffset != defOffset)

	// Just in case
	// Will be offset 7 hours to get UTC
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 15-04-05", mst)
	is.Equal(mstOffset, resOffset)
	// Result is MST
	is.Equal(resOffset, defOffset)

	// Will be offset 7 hours to get UTC
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102150405", mst)
	is.Equal(mstOffset, resOffset)
	// Result is MST
	is.Equal(resOffset, defOffset)

	// Try modifying zone
	// Will be offset 7 hours to get UTC
	sent, tStr, res, resOffset, defOffset = checkDate(t, "Mon, 02 Jan 2006 15:04:05 -0700", mst)
	is.Equal(mstOffset, resOffset)
	// Result is MST
	is.Equal(resOffset, defOffset)

	// Will be offset 7 hours to get UTC
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 15-04-05", mst)
	is.Equal(mstOffset, resOffset)
	// Result is MST
	is.Equal(resOffset, defOffset)

	// Try modifying zone
	// Will be offset 5 hours to get UTC
	sent, tStr, res, resOffset, defOffset = checkDate(t, "Mon, 02 Jan 2006 15:04:05 -0700", est)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	// Will be offset 5 hours to get UTC
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 15-04-05", est)
	is.Equal(estSTOffset, resOffset)
	// Result is EST
	is.Equal(resOffset, defOffset)

	// EST not used because a different offset is in the timestamp
	sent, tStr, res, resOffset, defOffset = checkDate(t, "Mon, 02 Jan 2006 15:04:05 -0600", mst)
	is.True(resOffset != defOffset)
	// Result is UTC-06:00
	d, err = time.ParseDuration("-6h")
	is.NoErr(err)
	is.True(resOffset == d)

	// RFC822Z
	sent, tStr, res, resOffset, defOffset = checkDate(t, "02 Jan 06 15:04 -0700", est)
	is.True(resOffset != defOffset)
	is.Equal(mstOffset, resOffset)

	// Just in case
	// Will be offset 5 hours to get UTC
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 15-04-05", est)
	is.Equal(estSTOffset, resOffset)
	is.Equal(resOffset, defOffset)

	// Will be offset 5 hours to get UTC
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102150405", est)
	is.Equal(estSTOffset, resOffset)
	is.Equal(resOffset, defOffset)

	t.Logf("Took %v to check", time.Since(start))
}

func TestISOCompare(t *testing.T) {
	is := is.New(t)

	start := time.Now()
	// It is possible to have a strring which is just digits that will be parsed
	// as a timestamp, incorrectly.

	ts := "2006-01-02T15:04:05-07:00"
	_, err := timestamp.ParseISOInUTC(ts)
	is.NoErr(err)
	count := 1000

	for i := 0; i < count; i++ {
		// Get a unix timestamp we should not parse
		_, err := timestamp.ParseInUTC(ts)
		is.NoErr(err) // Should parse with no error
	}

	t.Logf("Took %v to parse %s %d times", time.Since(start), ts, count)

	start = time.Now()

	ts = "20060102T150405-0700"
	for i := 0; i < count; i++ {
		// Get a unix timestamp we should not parse
		_, err := timestamp.ParseInUTC(ts)
		is.NoErr(err) // Should parse with no error
	}
	t.Logf("Took %v to parse %s %d times", time.Since(start), ts, count)
}

// TestOrdering check ordering call
func TestOrdering(t *testing.T) {
	is := is.New(t)

	t1, err1 := timestamp.ParseInUTC("20201210T223900-0500")
	is.NoErr(err1) // Should parse with no error

	t2, err2 := timestamp.ParseInUTC("20201211T223900-0500")
	is.NoErr(err2) // Should parse with no error

	is.True(timestamp.StartTimeIsBeforeEndTime(t1, t2))  // Start before end
	is.True(!timestamp.StartTimeIsBeforeEndTime(t2, t1)) // Start not before end
}

// Test how long it take to parse a timestamp 1,000 times
func TestTime(t *testing.T) {
	is := is.New(t)

	var unixBase time.Time
	var err error
	count := 1000
	defer track(runningtime(fmt.Sprintf("Time to parse timestamp %dx", count)))
	for i := 0; i < count; i++ {
		unixBase, err = timestamp.ParseInUTC("2006-01-02T15:04:05.000+00:00")
	}
	is.NoErr(err) // Should parse with no error
	t.Logf("Timestamp %s", timestamp.ISO8601MsecUTC(unixBase))
}

func TestFormat(t *testing.T) {
	is := is.New(t)
	ts, err := timestamp.ParseInUTC("2006-01-02T15:04:05.000+00:00")
	is.NoErr(err) // Should parse with no error

	// var unixBase time.Time
	var s string
	// var err error
	count := 1000
	defer track(runningtime(fmt.Sprintf("Time to format timestamp %dx", count)))
	for i := 0; i < count; i++ {
		s = timestamp.ISO8601MsecUTC(ts)
	}
	t.Logf("Timestamp %s", s)
}

var locations = []string{
	"MST",
	"America/New_York",
	"UTC",
	"Asia/Kabul",
	"America/St_Johns",
	"Europe/London",
	"America/Argentina/San_Luis",
	"Asia/Calcutta",
	"Asia/Tokyo",
	"America/Toronto",
}

// TestOffsetForZones test to get the offset for dates with a named zone. This
// could be more accurately done by removing the zone information from the
// timestamp string but normally this sort of opperation would be needed when
// timetamps were available without zone information but the location was known.
func TestOffsetForZones(t *testing.T) {
	is := is.New(t)

	var hours, minutes int
	var err error
	t1, err := timestamp.ParseInUTC("20200101T000000Z")
	is.NoErr(err)
	t2, err := timestamp.ParseInUTC("20200701T000000Z")
	is.NoErr(err)
	defer track(runningtime(fmt.Sprintf("Time to get offset information for %d locations/dates", len(locations)*2)))
	for _, location := range locations {
		for _, tNext := range []time.Time{t1, t2} {
			hours, minutes, err = timestamp.OffsetForLocation(tNext.Year(), tNext.Month(), tNext.Day(), location)
			is.NoErr(err)
			offset := timestamp.LocationOffsetString(hours, minutes)
			fmt.Printf("zone %s time %v offset %s\n", location, tNext, offset)
		}
	}
}

// Test how long it takes to get timezone information 1,000 times.
func TestZoneTime(t *testing.T) {
	is := is.New(t)

	zone := "Canada/Newfoundland"
	count := 1000
	defer track(runningtime(fmt.Sprintf("Time to get zone information %dx", count)))
	// var utcTime time.Time
	var hours, minutes int
	var err error
	for i := 0; i < count; i++ {
		hours, minutes, err = timestamp.OffsetForLocation(2006, 1, 1, zone)
		_ = timestamp.LocationOffsetString(hours, minutes)
	}
	is.NoErr(err)
	offset := timestamp.LocationOffsetStringDelimited(hours, minutes)
	t.Logf("start zone %s offset %s hours %d minutes %d offset %s error %v", zone, offset, hours, minutes, offset, err)
}

// Test seprate call to parse a Unix timestamp.
func TestParseUnixTimestamp(t *testing.T) {
	is := is.New(t)

	var err error
	var t1, t2 time.Time

	now := time.Now()
	ts1 := fmt.Sprint(now.UnixNano())
	t.Logf("Nano timestamp string %s len %d", ts1, len(ts1))
	ts2 := fmt.Sprint(now.Unix())

	count := 1000
	defer track(runningtime(fmt.Sprintf("Time to parse two timestamps %dx", count*2)))
	// var utcTime time.Time
	for i := 0; i < count; i++ {
		t1, err = timestamp.ParseUnixTS(ts1)
		t2, err = timestamp.ParseUnixTS(ts2)
	}
	is.True(t1 != time.Time{})
	is.True(t2 != time.Time{})
	is.NoErr(err)
}

func TestParseLocation(t *testing.T) {
	t1, _ := time.Parse("2006-01-02T15:04:05-0700", "2006-02-02T15:04:05-0700")

	t.Logf("Got time %v at location %s", t1.Format(time.RFC1123), t1.Location())
}

func TestParsISOTimestamp(t *testing.T) {
	is := is.New(t)

	var err error
	count := 1000
	var ts time.Time

	formats := []string{
		"20060102T010101",
		"20060102T010101.123456789",
		"20060102T010101-0400",
		"20060102t010101-0400",
		"20060102T010101+0400",
		"2006-01-02T01:01:01-04:00",
		"2006-01-02T01:01:01-06:00",
		"2006-01-02T18:01:01+01:00",
		"2006-01-02T18-01-01+0100",
		"2006/01/02T18.01.01+01:00",
		// Let 29th on leap year through
		"2000-02-29T12-01-01+0100",
	}

	badFormats := []string{
		// Bad month
		"2006-13-02T40-01-01+0100",
		// bad day
		"2006-02-30T12-01-01+0100",
		// Bad hours
		"2006-01-02T40-01-01+0100",
		// Bad minutes
		"2006-01-02T11-60-01+0100",
		// Bad seconds
		"2006-01-02T11-30-61+0100",
		"20060102T010101-04000",
		"20060102T0101Z01Z",
		"2006w01s02T18a01b01c01:00",
		"2006-01-02T18:01:01b01:00",
	}

	for _, in := range formats {
		ts, err = timestamp.ParseISOTimestamp(in, time.UTC)
		t.Logf("input %s ts %v", in, ts)
		// is.NoErr(err)
	}

	for _, in := range badFormats {
		ts, err = timestamp.ParseISOTimestamp(in, time.UTC)
		t.Logf("input %s error %v", in, err)
		// is.True(err != nil)
	}

	defer track(runningtime(fmt.Sprintf("Time to process ISO timestamp %dx", count*2)))
	for i := 0; i < count; i++ {
		ts, err = timestamp.ParseISOTimestamp("20060102T010101.", time.UTC)
		is.NoErr(err)
	}
	is.NoErr(err)
	t.Log("ts", ts)
}

// Run as
//  go test -run=XXX -bench=.
//  go test -bench=. -benchmem -memprofile memprofile.out -cpuprofile cpuprofile.out
//  go tool pprof -http=:8080 memprofile.out
//  go tool pprof -http=:8080 cpuprofile.out
func BenchmarkUnixTimestampTest(b *testing.B) {
	is := is.New(b)

	var err error
	var t1 time.Time

	now := time.Now()
	ts1 := fmt.Sprint(now.Unix())

	b.SetBytes(2)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			t1, err = timestamp.ParseUnixTS(ts1)
			if err != nil {
				b.Log(err)
			}
		}
	})
	is.True(t1 != time.Time{})
	is.NoErr(err)
}

// Run as
//  go test -run=XXX -bench=.
//  go test -bench=. -benchmem -memprofile memprofile.out -cpuprofile cpuprofile.out
//  go tool pprof -http=:8080 memprofile.out
//  go tool pprof -http=:8080 cpuprofile.out
func BenchmarkUnixTimestampNanoTest(b *testing.B) {
	is := is.New(b)

	var err error
	var t1 time.Time

	now := time.Now()
	ts1 := fmt.Sprint(now.UnixNano())

	b.SetBytes(2)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			t1, err = timestamp.ParseUnixTS(ts1)
			if err != nil {
				b.Log(err)
			}
		}
	})
	is.True(t1 != time.Time{})
	is.NoErr(err)
}

func BenchmarkIterativeISOTimestampTest(b *testing.B) {
	is := is.New(b)

	var err error
	var t1 time.Time

	b.SetBytes(2)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			t1, err = timestamp.ParseISOTimestamp("20060102T010101.", time.UTC)
			if err != nil {
				b.Log(err)
			}
		}
	})
	is.True(t1 != time.Time{})
	is.NoErr(err)
}
func BenchmarkLexedISOTimestampTest(b *testing.B) {
	is := is.New(b)

	var err error
	var t1 time.Time

	// now := time.Now()
	// ts1 := fmt.Sprint(now.UnixNano())

	b.SetBytes(2)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			t1, err = lex.ParseInLocation([]byte("20060102T010101"), time.UTC)
			if err != nil {
				b.Log(err)
			}
		}
	})
	is.True(t1 != time.Time{})
	is.NoErr(err)
}
