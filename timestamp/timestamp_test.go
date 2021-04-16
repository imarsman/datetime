package timestamp_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/imarsman/datetime/timestamp"
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

// checkDate for use in parse checking
func checkDate(t *testing.T, input string, compare string, location *time.Location) (string, string) {
	is := is.New(t)
	v, err := timestamp.ParseInLocation(input, location)
	is.NoErr(err)

	ts := timestamp.ISO8601Msec(v)
	fmt.Printf("Input %s, Expected %s, Got %s UTC In Location %s\n", input, compare, ts, location.String())
	return compare, ts
}

// TestParse parse all patterns anc compare with expected values
func TestParse(t *testing.T) {
	is := is.New(t)

	mst, err := time.LoadLocation("MST")
	is.NoErr(err)

	test, err := time.LoadLocation("EST")
	is.NoErr(err)
	est, err := time.LoadLocation("America/Toronto")
	is.NoErr(err)
	is.True(test != est)

	start := time.Now()

	// It is possible to have a string which is just digits that will be parsed
	// as a timestamp, incorrectly.
	t.Log(timestamp.ParseInUTC("2006010247"))

	// Get a unix timestamp we should not parse
	_, err = timestamp.ParseInUTC("1")
	is.True(err != nil) // Error should be true

	// Get time value from parsed reference time
	unixBase, err := timestamp.ParseInUTC("2006-01-02T15:04:05.000+00:00")
	is.NoErr(err)

	var expected, got string
	// Use parsed reference time to create unix timestamp and nanosecond timestamp
	expected, got = checkDate(t, fmt.Sprint(unixBase.UnixNano()), "2006-01-02T15:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, fmt.Sprint(unixBase.Unix()), "2006-01-02T15:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)

	// Short ISO-8601 timestamps with numerical zone offsets
	expected, got = checkDate(t, "20060102T150405-0700", "2006-01-02T22:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "20060102T150405-0700", "2006-01-02T22:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "20060102T150405-07", "2006-01-02T22:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "20060102T150405.000+0000", "2006-01-02T15:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "20060102T150405.000-0000", "2006-01-02T15:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "20060102T150405.000-0700", "2006-01-02T22:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "20060102T150405.000+0700", "2006-01-02T08:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "20060102T150405.000000-0700", "2006-01-02T22:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "20060102T150405.999999999-0700", "2006-01-02T22:04:05.999+00:00", time.UTC)
	is.Equal(expected, got)

	// Long ISO-8601 timestamps with numerical zone offsets
	expected, got = checkDate(t, "2006-01-02T15:04:05-07:00", "2006-01-02T22:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006-01-02T15:04:05-07", "2006-01-02T22:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006-01-02T15:04:05.000-07:00", "2006-01-02T22:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006-01-02T15:04:05.000-07", "2006-01-02T22:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006-01-02T15:04:05.000000-07:00", "2006-01-02T22:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006-01-02T15:04:05.001000-07", "2006-01-02T22:04:05.001+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006-01-02T15:04:05.001000000-07:00", "2006-01-02T22:04:05.001+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006-01-02T15:04:05.999999999-07", "2006-01-02T22:04:05.999+00:00", time.UTC)
	is.Equal(expected, got)

	// Short  ISO-8601 timestamps with UTC zone offsets
	expected, got = checkDate(t, "20060102T150405Z", "2006-01-02T15:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "20060102T150405.000Z", "2006-01-02T15:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "20060102T150405.000000Z", "2006-01-02T15:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "20060102T150405.000000000Z", "2006-01-02T15:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "20060102T150405.001000000Z", "2006-01-02T15:04:05.001+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "20060102T150405.000100000Z", "2006-01-02T15:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "20060102T150405.999999999Z", "2006-01-02T15:04:05.999+00:00", time.UTC)
	is.Equal(expected, got)

	// Long date time with UTC zone offsets
	expected, got = checkDate(t, "2006-01-02T15:04:05Z", "2006-01-02T15:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006-01-02T15:04:05.000Z", "2006-01-02T15:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006-01-02T15:04:05.000000Z", "2006-01-02T15:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006-01-02T15:04:05.999999999Z", "2006-01-02T15:04:05.999+00:00", time.UTC)
	is.Equal(expected, got)

	// Just in case
	expected, got = checkDate(t, "2006-01-02 15-04-05", "2006-01-02T15:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "20060102150405", "2006-01-02T15:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)

	// Short ISO-8601 timestamps with no zone offset. Assume UTC.
	expected, got = checkDate(t, "20060102T150405", "2006-01-02T15:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "20060102T150405.000", "2006-01-02T15:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "20060102T150405.000000", "2006-01-02T15:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "20060102T150405.999999999", "2006-01-02T15:04:05.999+00:00", time.UTC)
	is.Equal(expected, got)

	// SQL
	expected, got = checkDate(t, "2006-01-02 22:04:05", "2006-01-02T22:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	// MST is -0700 from UTC, so UTC will be 7 hours ahead
	expected, got = checkDate(t, "2006-01-02 22:04:05", "2006-01-03T05:04:05.000+00:00", mst)
	is.Equal(expected, got)
	// MST is -0500 from UTC, so UTC will be 5 hours ahead
	expected, got = checkDate(t, "2006-01-02 22:04:05", "2006-01-03T03:04:05.000+00:00", est)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006-01-02 22:04:05 -00", "2006-01-02T22:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	// The input has a timestamp so EST will not be applied
	expected, got = checkDate(t, "2006-01-02 22:04:05 -00", "2006-01-02T22:04:05.000+00:00", est)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006-01-02 22:04:05 +00", "2006-01-02T22:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006-01-02 22:04:05 -00:00", "2006-01-02T22:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006-01-02 22:04:05 +00:00", "2006-01-02T22:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)

	// Hopefully less likely to be found. Assume UTC.
	expected, got = checkDate(t, "20060102", "2006-01-02T00:00:00.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006-01-02", "2006-01-02T00:00:00.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006/01/02", "2006-01-02T00:00:00.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "01/02/2006", "2006-01-02T00:00:00.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "1/2/2006", "2006-01-02T00:00:00.000+00:00", time.UTC)
	is.Equal(expected, got)

	// Weird ones with improper separators
	expected, got = checkDate(t, "2006-01-02T15-04-05-0700", "2006-01-02T22:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006-01-02T15-04-05.000-0700", "2006-01-02T22:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006-01-02T15-04-05.000000-0700", "2006-01-02T22:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006-01-02T15-04-05.999999999-0700", "2006-01-02T22:04:05.999+00:00", time.UTC)
	is.Equal(expected, got)

	expected, got = checkDate(t, "2006-01-02T15-04-05-07:00", "2006-01-02T22:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006-01-02T15-04-05.000-07:00", "2006-01-02T22:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006-01-02T15-04-05.000000-07:00", "2006-01-02T22:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)
	expected, got = checkDate(t, "2006-01-02T15-04-05.999999999-07:00", "2006-01-02T22:04:05.999+00:00", time.UTC)
	is.Equal(expected, got)

	// RFC7232 - used in HTTP protocol
	expected, got = checkDate(t, "Mon, 02 Jan 2006 15:04:05 GMT", "2006-01-02T15:04:05.000+00:00", time.UTC)
	is.Equal(expected, got)

	// // RFC850
	// expected, got = checkDate(t, "Monday, 02-Jan-06 15:04:05 MST", "2006-01-02T22:04:05.000+00:00")

	// // RFC1123
	// expected, got = checkDate(t, "Mon, 02 Jan 2006 15:04:05 MST", "2006-01-02T22:04:05.000+00:00")

	// RFC1123Z
	expected, got = checkDate(t, "Mon, 02 Jan 2006 15:04:05 -0700", "2006-01-02T22:04:05.000+00:00", mst)
	is.Equal(expected, got)
	expected, got = checkDate(t, "Mon, 02 Jan 2006 15:04:05", "2006-01-02T22:04:05.000+00:00", mst)
	is.Equal(expected, got)
	// MST not used because a different offset is in the timestamp
	expected, got = checkDate(t, "Mon, 02 Jan 2006 15:04:05 -0600", "2006-01-02T21:04:05.000+00:00", mst)
	is.Equal(expected, got)

	// RFC822Z
	expected, got = checkDate(t, "02 Jan 06 15:04 -0700", "2006-01-02T22:04:00.000+00:00", time.UTC)
	is.Equal(expected, got)

	// Just in case
	// Will be offset 7 hours to get UTC
	expected, got = checkDate(t, "2006-01-02 15-04-05", "2006-01-02T22:04:05.000+00:00", mst)
	is.Equal(expected, got)
	// Will be offset 7 hours to get UTC
	expected, got = checkDate(t, "20060102150405", "2006-01-02T22:04:05.000+00:00", mst)
	is.Equal(expected, got)

	// Try modifying zone
	// Will be offset 7 hours to get UTC
	expected, got = checkDate(t, "Mon, 02 Jan 2006 15:04:05 -0700", "2006-01-02T22:04:05.000+00:00", mst)
	is.Equal(expected, got)
	// Will be offset 7 hours to get UTC
	expected, got = checkDate(t, "2006-01-02 15-04-05", "2006-01-02T22:04:05.000+00:00", mst)
	is.Equal(expected, got)

	// Try modifying zone
	// Will be offset 5 hours to get UTC
	expected, got = checkDate(t, "Mon, 02 Jan 2006 15:04:05 -0700", "2006-01-02T22:04:05.000+00:00", est)
	is.Equal(expected, got)
	// Will be offset 5 hours to get UTC
	expected, got = checkDate(t, "2006-01-02 15-04-05", "2006-01-02T20:04:05.000+00:00", est)
	is.Equal(expected, got)
	// EST not used because a different offset is in the timestamp
	expected, got = checkDate(t, "Mon, 02 Jan 2006 15:04:05 -0600", "2006-01-02T21:04:05.000+00:00", mst)
	is.Equal(expected, got)

	// RFC822Z
	expected, got = checkDate(t, "02 Jan 06 15:04 -0700", "2006-01-02T22:04:00.000+00:00", est)
	is.Equal(expected, got)

	// Just in case
	// Will be offset 5 hours to get UTC
	expected, got = checkDate(t, "2006-01-02 15-04-05", "2006-01-02T20:04:05.000+00:00", est)
	is.Equal(expected, got)
	// Will be offset 5 hours to get UTC
	expected, got = checkDate(t, "20060102150405", "2006-01-02T20:04:05.000+00:00", est)
	is.Equal(expected, got)

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
	t.Logf("Timestamp %s", timestamp.ISO8601Msec(unixBase))
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
		s = timestamp.ISO8601Msec(ts)
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
