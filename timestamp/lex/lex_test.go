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
		"2006-01-02",
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
		// Space instead of T
		"20200102 122436-0000",
		// No T delimiter
		"20200102122436-0000",
		"20200102T122436-0500",
		"2020-01-02T12:24:36-04:00",
		"2020-01-02T12:24:36Z",
		// Incorrect dashes between time elements
		"2020-01-02T12-24-36Z",
		"2020-01-02T12-24-36-0400",

		// No dashes to maatch coolons
		"20200102T12:24:36-05:00",
		// Hour only offset
		"20200102T12:24:36-05",
		// No delimiters between date and time elements but colon for offset
		"20200102T122436.123-05:00",
		// Will show no decimals
		"20060102T150405.000Z",
		// Will have one decimal place
		"20060102T150405.100Z",
		// Will have two decimal places
		"20060102T150405.120Z",
		// Will have three decimal places
		"20060102T150405.123Z",
		// Will have four decimal places
		"20060102T150405.1234Z",
		// Will have five decimal places
		"20060102T150405.12345Z",
		// Will have six decimal places
		"20060102T150405.123456Z",
		// Will have seven decimal places
		"20060102T150405.1234567Z",
		// Will have eight decimal places
		"20060102T150405.12345678Z",
		// Will have nine decimal places
		"20060102T150405.123456789Z",
		"20060102",
		"2006-01-02T15:04:05+0700",
		"1997-01-31 09:26:56.66 +02:00",
		// Will have eight decimal places
		"20060102T150405.123456000-0400",
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

// Run as
//  go test -run=XXX -bench=.
func BenchmarkTest(b *testing.B) {
	_, err := lex.Parse([]byte("20200102T122436-0400"))
	if err != nil {
		b.Log(err)
	}
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := lex.Parse([]byte("20200102T122436-0400"))
			if err != nil {
				b.Log(err)
			}
		}
	})
}
