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

func TestGetParts(t *testing.T) {

	// is := is.New(t)
	years, months, days, hours, minutes, seconds, subseconds, err := period.AdditionsFromDecimalSection('S', 13, 575)
	t.Logf("years %d, months %d, days %d, hours %d, minutes %d, seconds %d, subseconds %d, err %v",
		years, months, days, hours, minutes, seconds, subseconds, err)
}

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
		// is.NoErr(err)
	}

}

func BenchmarkParsePeriod(b *testing.B) {
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
