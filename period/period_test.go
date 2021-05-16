package period_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/imarsman/datetime/period"
	"github.com/matryer/is"
)

const bechmarkBytesPerOp int64 = 10

//                Tests and benchmarks
// -----------------------------------------------------
// benchmark
//   go test -run=XXX -bench=. -benchmem
// Get allocation information and pipe to less
//   go build -gcflags '-m -m' ./*.go 2>&1 |less
// Run all tests
//   go test -v
// Run one test and do allocation profiling
//   go test -run=XXX -bench=IterativeISOTimestampLong -gcflags '-m' 2>&1 |less
// Run a specific test by function name pattern
//  go test -run=TestParsISOTimestamp
//
//  go test -run=XXX -bench=.
//  go test -bench=. -benchmem -memprofile memprofile.out -cpuprofile cpuprofile.out
//  go tool pprof -http=:8080 memprofile.out
//  go tool pprof -http=:8080 cpuprofile.out

func info(i int, m ...interface{}) string {
	if s, ok := m[0].(string); ok {
		m[0] = i
		return fmt.Sprintf("%d "+s, m...)
	}
	return fmt.Sprintf("%d %v", i, m[0])
}

//report start of running time tracking
func runningtime(s string) (string, time.Time) {
	fmt.Println("Start:	", s)
	return s, time.Now()
}

// report total running time
func track(s string, startTime time.Time) {
	endTime := time.Now()
	fmt.Println("End:	", s, "took", endTime.Sub(startTime))
}

// sample running time tracking
func execute() {
	defer track(runningtime("execute"))
	time.Sleep(3 * time.Second)
}

func TestPeriodParser(t *testing.T) {
	tests := []string{
		"P12Y1WT1H14M5S",
		"P3Y1WT1H14M",
		"P-3Y1WT1H14M",
		"P-3Y1WT1H1400M",
	}

	is := is.New(t)

	for _, test := range tests {
		p, err := period.Parse(test, true)
		is.NoErr(err)
		t.Logf("Got %20s %20s %20s", test, p.String(), p.Normalise(false).String())
	}

}

func TestParsePeriod(t *testing.T) {
	tests := []string{
		"P1Y",
		"P1Y1D",
		"P3Y",
		"P1M",
		"P2M",
		"P1W",
		// Showing shifting of years to months
		"P11Y",
		// Showing shifting of hours to minutes
		"PT11H",
		// Showing shifting of monts to days
		"P11M",
		"PT1H",
		// Showing shifting of minutes to seconds
		"P1MT1H31M",
		"PT1M",
		"PT1M5S",
		"PT1S",
		"PT1000S",
		"P1W",
		"P3Y1W",
		"P4W",
		"P2Y3M4W5D",
		"P3Y1WT1H14M",
		"P-3Y1WT1H14M",
		"P3Y1WT1H14M",
		"P-3Y1WT1H1400M",
		"P120Y120M200D",
		"P150Y150M200DT1H4M2000S",
		"P250Y150M200DT1H4M2000S",
	}

	is := is.New(t)

	for _, test := range tests {
		p, _ := period.Parse(test, true, true)
		d, _, err := p.Duration()
		is.NoErr(err)
		fmt.Printf("Input %-15s period %0-15s normalized %-20s duration %-15v\n",
			test, p.String(), p.Normalise(false).String(), d)
	}
}

func TestParsePeriodWithFractionalParts(t *testing.T) {
	is := is.New(t)

	tests := []string{
		"P1.5Y",
		"PT1.5S",
		"PT1.567S",
	}

	for _, test := range tests {
		p, _ := period.Parse(test, true, true)
		d, _, err := p.Duration()
		is.NoErr(err)
		fmt.Printf("Input %-15s period %0-15s normalized %-20s duration %-15v\n",
			test, p.String(), p.Normalise(false).String(), d)
	}

}

func TestGetParts(t *testing.T) {
	is := is.New(t)

	type periodParts struct {
		part      rune
		pre, post int64
	}

	parts := []periodParts{
		{'S', 1, 1},           // 1.iseconds - should give 100 ms
		{'S', 13, 1575},       // 13.1575 seconds - should give 157 ms
		{'I', 13, 575},        // 13.575 minutes - should give 575 ms
		{'H', 200, 5},         // 200 hours and 30 minutes
		{'Y', 260, 5},         // 260 years and 6 months
		{'Y', 15000000000, 5}, // 15 billion years (and 6 months) - over threshold
	}

	for _, part := range parts {
		years, months, days, hours, minutes, seconds, subseconds, err := period.AdditionsFromDecimalSection(part.part, part.pre, part.post)
		t.Logf("years %d, months %d, days %d, hours %d, minutes %d, seconds %d, subseconds %d, err %v",
			years, months, days, hours, minutes, seconds, subseconds, err)

		is.NoErr(err)
	}
}

// No use of arbitrary precision decimals
// With 'I', 13, 575
// 15.77 ns/op   633.97 MB/s   0 B/op   0 allocs/op
func BenchmarkGetAdditions(b *testing.B) {
	is := is.New(b)

	var err error
	var years, months, days, hours, minutes, seconds int64
	var subseconds int

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			years, months, days, hours, minutes, seconds, subseconds, err = period.AdditionsFromDecimalSection(
				'I', 13, 575)
		}
	})

	b.Logf("years %d, months %d, days %d, hours %d, minutes %d, seconds %d, subseconds %d, err %v",
		years, months, days, hours, minutes, seconds, subseconds, err)

	is.NoErr(err) // Parsing should not have caused an error
}

// Force use of arbitrary precision decimals
// With 'Y', 30000, 575
// 384.3 ns/op   26.02 MB/s   640 B/op   19 allocs/op
func BenchmarkGetAdditionsLong(b *testing.B) {
	is := is.New(b)

	var err error
	var years, months, days, hours, minutes, seconds int64
	var subseconds int

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			years, months, days, hours, minutes, seconds, subseconds, err = period.AdditionsFromDecimalSection(
				'Y', 30000000000, 575)
		}
	})

	b.Logf("years %d, months %d, days %d, hours %d, minutes %d, seconds %d, subseconds %d, err %v",
		years, months, days, hours, minutes, seconds, subseconds, err)

	is.NoErr(err) // Parsing should not have caused an error
}

func BenchmarkGetAdditionsSubThreshold(b *testing.B) {
	is := is.New(b)

	var err error
	var years, months, days, hours, minutes, seconds int64
	var subseconds int

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			years, months, days, hours, minutes, seconds, subseconds, err = period.AdditionsFromDecimalSection(
				'Y', 30000, 575)
		}
	})

	b.Logf("years %d, months %d, days %d, hours %d, minutes %d, seconds %d, subseconds %d, err %v",
		years, months, days, hours, minutes, seconds, subseconds, err)

	is.NoErr(err) // Parsing should not have caused an error
}

// TestParsePeriodBad parse intentionally incorrect periods
func TestParsePeriodBad(t *testing.T) {
	tests := []string{
		"P300YT1H4M2000S",
		"P3YT2629999H",
	}

	is := is.New(t)

	// var err error
	for _, test := range tests {
		p, _ := period.Parse(test, false)
		d, _, err := p.Duration()
		is.True(err != nil)
		fmt.Printf("Input %-15s period %0-15s normalized %-20s duration %-15v\n",
			test, p.String(), p.Normalise(true).String(), d)
	}

}

// Slower with more allocations with a complex period
// 209.8 ns/op   47.67 MB/s	  160 B/op	  16 allocs/op
func BenchmarkParsePeriodLong(b *testing.B) {
	is := is.New(b)

	var p period.Period
	var err error

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p, err = period.Parse("P250Y150M200DT1H4M2000S", true)
		}
	})

	b.Log(p.String())
	is.True(p != period.Period{})
	is.NoErr(err) // Parsing should not have caused an error
}

func BenchmarkParsePeriodFractional(b *testing.B) {
	is := is.New(b)

	var p period.Period
	var err error

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p, err = period.Parse("PT1.5S", true)
		}
	})

	b.Log(p.String())
	is.True(p != period.Period{})
	is.NoErr(err) // Parsing should not have caused an error
}

// Faster with fewer allocations with a simple period
// 41.72 ns/op   239.68 MB/s	  16 B/op   2 allocs/op
func BenchmarkParsePeriodShort(b *testing.B) {
	is := is.New(b)

	var p period.Period
	var err error

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p, err = period.Parse("P1M", true)
		}
	})

	b.Log(p.String())
	is.True(p != period.Period{})
	is.NoErr(err) // Parsing should not have caused an error
}
