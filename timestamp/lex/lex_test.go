package lex_test

import (
	"fmt"
	"testing"
	"time"

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

func TestParseTime(t *testing.T) {
	is := is.New(t)

	formats := []string{
		"20200102T122436Z",
		"20200102T122436-0000",
		"20200102T122436-0400",
		"2020-01-02T12:24:36-04:00",
		"2020-01-02T12:24:36Z",
		// This will work because colons are removed
		"20200102T12:24:36-04:00",
		// Colons removed so it will work
		"20200102T122436.123-04:00",
		// No zone allowed
		"20060102",
		// No zone allowed
		"2006/01/02",
		// No zone allowed
		"2006.01.02",
	}
	for _, f := range formats {
		ts, err := lex.Parse([]byte(f))
		count := 1000
		start := time.Now()
		// defer track(runningtime(fmt.Sprintf("Time to parse timestamp %dx", count)))
		for i := 0; i < count; i++ {
			ts, err = lex.Parse([]byte(f))
			is.NoErr(err)
		}

		t.Logf("Time to parse timestamp %s %dx = %v", f, count, time.Since(start))
		t.Log(ts)
	}
}

func TestParseFormats(t *testing.T) {
	is := is.NewRelaxed(t)

	formats := []string{
		"20200102T122436Z",
		"20200102T122436-0000",
		"20200102 122436-0000",
		"20200102122436-0000",
		"20200102T122436-0500",
		"2020-01-02T12:24:36-04:00",
		"2020-01-02T12:24:36Z",
		"2020-01-02T12-24-36Z",
		"2020-01-02T12-24-36Z",
		"20200102T12:24:36-05:00",
		"20200102T12:24:36-05",
		"20200102T122436.123-05:00",
		"20060102T150405.000Z",
		"20060102",
		"2006-01-02T15:04:05+0700",
	}
	for _, f := range formats {
		ts, err := lex.Parse([]byte(f))
		is.NoErr(err)

		tStr := ts.Format("20060102T150405.999999999-0700")
		t.Logf("Input %s, output %v", f, tStr)
	}
	_, err := lex.Parse([]byte("20060102Z"))
	is.True(err != nil)

	_, err = lex.Parse([]byte("20060102-0400"))
	is.True(err != nil)
}
