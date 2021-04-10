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
func checkDate(t *testing.T, input string, compare string) {
	is := is.New(t)
	v, err := timestamp.ParseUTC(input)
	is.NoErr(err)

	ts := timestamp.ISO8601LongMsec(v)
	t.Logf("Input %s, Expecting %s, Got %s", input, compare, ts)
	is.Equal(compare, ts)
}

// TestParse parse all patterns anc compare with expected values
func TestParse(t *testing.T) {
	is := is.New(t)

	start := time.Now()
	// It is possible to have a strring which is just digits that will be parsed
	// as a timestamp, incorrectly.
	t.Log(timestamp.ParseUTC("2006010247"))

	// Get a unix timestamp we should not parse
	_, err := timestamp.ParseUTC("1")
	is.True(err != nil) // Error should be true

	// Get time value from parsed reference time
	unixBase, err := timestamp.ParseUTC("2006-01-02T15:04:05.000+00:00")
	is.NoErr(err)

	// Use parsed reference time to create unix timestamp and nanosecond timestamp
	checkDate(t, fmt.Sprint(unixBase.UnixNano()), "2006-01-02T15:04:05.000+00:00")
	checkDate(t, fmt.Sprint(unixBase.Unix()), "2006-01-02T15:04:05.000+00:00")

	// RFC7232 - used in HTTP protocol
	checkDate(t, "Mon, 02 Jan 2006 15:04:05 GMT", "2006-01-02T15:04:05.000+00:00")

	// Short ISO-8601 timestamps with numerical zone offsets
	checkDate(t, "20060102T150405-0700", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "20060102T150405-0700", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "20060102T150405-07", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "20060102T150405.000+0000", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102T150405.000-0000", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102T150405.000-0700", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "20060102T150405.000+0700", "2006-01-02T08:04:05.000+00:00")
	checkDate(t, "20060102T150405.000000-0700", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "20060102T150405.999999999-0700", "2006-01-02T22:04:05.999+00:00")

	// Long ISO-8601 timestamps with numerical zone offsets
	checkDate(t, "2006-01-02T15:04:05-07:00", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15:04:05-07", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15:04:05.000-07:00", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15:04:05.000-07", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15:04:05.000000-07:00", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15:04:05.001000-07", "2006-01-02T22:04:05.001+00:00")
	checkDate(t, "2006-01-02T15:04:05.001000000-07:00", "2006-01-02T22:04:05.001+00:00")
	checkDate(t, "2006-01-02T15:04:05.999999999-07", "2006-01-02T22:04:05.999+00:00")

	// Short  ISO-8601 timestamps with UTC zone offsets
	checkDate(t, "20060102T150405Z", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102T150405.000Z", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102T150405.000000Z", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102T150405.000000000Z", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102T150405.001000000Z", "2006-01-02T15:04:05.001+00:00")
	checkDate(t, "20060102T150405.000100000Z", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102T150405.999999999Z", "2006-01-02T15:04:05.999+00:00")

	// Long date time with UTC zone offsets
	checkDate(t, "2006-01-02T15:04:05Z", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "2006-01-02T15:04:05.000Z", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "2006-01-02T15:04:05.000000Z", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "2006-01-02T15:04:05.999999999Z", "2006-01-02T15:04:05.999+00:00")

	// Just in case
	checkDate(t, "2006-01-02 15-04-05", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102150405", "2006-01-02T15:04:05.000+00:00")

	// Short ISO-8601 timestamps with no zone offset. Assume UTC.
	checkDate(t, "20060102T150405", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102T150405.000", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102T150405.000000", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102T150405.999999999", "2006-01-02T15:04:05.999+00:00")

	// SQL
	checkDate(t, "2006-01-02 22:04:05", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02 22:04:05 -00", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02 22:04:05 +00", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02 22:04:05 -00:00", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02 22:04:05 +00:00", "2006-01-02T22:04:05.000+00:00")

	// Hopefully less likely to be found. Assume UTC.
	checkDate(t, "20060102", "2006-01-02T00:00:00.000+00:00")
	checkDate(t, "2006-01-02", "2006-01-02T00:00:00.000+00:00")
	checkDate(t, "2006/01/02", "2006-01-02T00:00:00.000+00:00")
	checkDate(t, "01/02/2006", "2006-01-02T00:00:00.000+00:00")
	checkDate(t, "1/2/2006", "2006-01-02T00:00:00.000+00:00")

	// Weird ones with improper separators
	checkDate(t, "2006-01-02T15-04-05-0700", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15-04-05.000-0700", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15-04-05.000000-0700", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15-04-05.999999999-0700", "2006-01-02T22:04:05.999+00:00")

	checkDate(t, "2006-01-02T15-04-05-07:00", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15-04-05.000-07:00", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15-04-05.000000-07:00", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15-04-05.999999999-07:00", "2006-01-02T22:04:05.999+00:00")

	t.Logf("Took %v to check", time.Since(start))
}

func TestISOCompare(t *testing.T) {
	is := is.New(t)

	start := time.Now()
	// It is possible to have a strring which is just digits that will be parsed
	// as a timestamp, incorrectly.

	format := "2006-01-02T15:04:05-07:00"
	_, err := timestamp.ParseUTC(format)
	is.NoErr(err)
	count := 1000

	for i := 0; i < count; i++ {
		// Get a unix timestamp we should not parse
		_, err := timestamp.ParseUTC(format)
		is.NoErr(err) // Should parse with no error
	}

	t.Logf("Took %v to parse %s  %d times", time.Since(start), format, count)

	start = time.Now()

	format = "20060102T150405-0700"
	for i := 0; i < count; i++ {
		// Get a unix timestamp we should not parse
		_, err := timestamp.ParseUTC(format)
		is.NoErr(err) // Should parse with no error
	}
	t.Logf("Took %v to parse %s  %d times", time.Since(start), format, count)

	format = "2006-01-02T15:04:05-07:00"
	// format = "1/2/2006"
	for i := 0; i < count; i++ {
		// Get a unix timestamp we should not parse
		_, err := timestamp.ParseUTC(format)
		is.NoErr(err) // Should parse with no error
	}
	t.Logf("Took %v to check %s  %d times", time.Since(start), format, count)
}

// TestOrdering check ordering call
func TestOrdering(t *testing.T) {
	is := is.New(t)

	t1, err1 := timestamp.ParseUTC("20201210T223900-0500")
	is.NoErr(err1) // Should parse with no error

	t2, err2 := timestamp.ParseUTC("20201211T223900-0500")
	is.NoErr(err2) // Should parse with no error

	is.True(timestamp.StartTimeIsBeforeEndTime(t1, t2))  // Start before end
	is.True(!timestamp.StartTimeIsBeforeEndTime(t2, t1)) // Start not before end
}
func TestTime(t *testing.T) {
	is := is.New(t)

	var unixBase time.Time
	var err error
	count := 1000
	defer track(runningtime(fmt.Sprintf("Time to parse timestamp %dx", count)))
	for i := 0; i < count; i++ {
		unixBase, err = timestamp.ParseUTC("2006-01-02T15:04:05.000+00:00")
	}
	is.NoErr(err) // Should parse with no error
	t.Logf("Timestamp %s", timestamp.ISO8601LongMsec(unixBase))
}
