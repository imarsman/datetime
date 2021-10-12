package period_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/imarsman/datetime/period"
	"github.com/matryer/is"

	// "golang.org/x/text/language"
	"golang.org/x/text/message"
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

func TestSum(t *testing.T) {
	is := is.New(t)
	var sum float64 = 0
	for i := 1; i <= 100; i++ {
		sum += float64(i) / 2
	}
	t.Log("sum", sum)
	is.True(sum == 2525)
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
		"PT1S",
		"PT1M",
		"PT1H",
		"P1D",
		"P1W",
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
		// Showing shifting of minutes to seconds
		"P1MT1H31M",
		"PT1M5S",
		"PT1000S",
		"PT0.1S",
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
		"P250000Y150M200DT1H4M2000S",
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
		"P0D",
		"PT0.5S",
		// Missing leading zero
		"PT.5S",
		"PT0.5M",
		// Missing leading zero
		"PT.5M",
		"PT0.5H",
		// Missing leading zero
		"PT.5H",
		// Use comma instead
		"PT1,5M",
		"PT1.5M",
		"P1.5M",
		"P1.5Y",
		"P260.5W",
		"P1.5M",
		"P1.5D",
		"PT1.5H",
		"PT1.5M",
		"PT1.5S",
		"PT1.05S",
		"PT1.500S",
		"PT1.567S",
		"PT1H14M",
	}

	for _, test := range tests {
		p, _ := period.Parse(test, true, true)
		d, _, err := p.Duration()
		is.NoErr(err)
		fmt.Printf("Input %-15s period %0-15s normalized %-20s duration %-15v\n",
			test, p.String(), p.Normalise(false).String(), d)
	}

}

// TestParsePeriodBad parse intentionally incorrect periods
func TestParsePeriodBad(t *testing.T) {
	tests := []string{
		"P4M300YT1H4M2000S",
		"P30000YT2629999.5H16M",
		"PT1H1Y",
	}

	is := is.New(t)

	for _, test := range tests {
		p, err := period.Parse(test, false)
		t.Log(err)
		is.True(err != nil)
		is.Equal(p.String(), "P0D")
	}

}

// No use of arbitrary precision decimals
// With 'I', 13, 575
// 15.77 ns/op   0 B/op   0 allocs/op
func BenchmarkGetAdditions(b *testing.B) {
	is := is.New(b)

	var err error
	var years, months, days, hours, minutes, seconds int64
	var subseconds int64

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

func TestGetFractionalParts(t *testing.T) {
	is := is.New(t)

	type periodParts struct {
		part rune
		pre  int64
		post float64
	}

	parts := []periodParts{
		{'S', 1, .1},           // 1.1 seconds - should give 100 ms
		{'S', 13, .1575},       // 13.1575 seconds - should give 157 ms
		{'I', 13, .575},        // 13.575 minutes - should give 575 ms
		{'H', 200, .5},         // 200 hours and 30 minutes
		{'Y', 260, .5},         // 260 years and 6 months
		{'W', 260, .5},         // 260 weeks and .5 week
		{'Y', 15000000000, .5}, // 15 billion years (and 6 months) - over threshold
	}

	p := message.NewPrinter(message.MatchLanguage("en"))

	for _, part := range parts {
		years, months, days, hours, minutes, seconds, subseconds, err := period.AdditionsFromDecimalSection(part.part, part.pre, part.post)
		is.NoErr(err)
		t.Log(p.Sprintf("part %s pre %d post %f years %d, months %d, days %d, hours %d, minutes %d, seconds %d, subseconds %d, err %v",
			string(part.part), part.pre, part.post, years, months, days, hours, minutes, seconds, subseconds, err))
	}
}

// Force use of arbitrary precision decimals
// With 'Y', 30000, 575
// 384.3 ns/op   26.02 MB/s   640 B/op   19 allocs/op
func BenchmarkGetAdditionsLong(b *testing.B) {
	is := is.New(b)

	var err error
	var years, months, days, hours, minutes, seconds int64
	var subseconds int64

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			years, months, days, hours, minutes, seconds, subseconds, err = period.AdditionsFromDecimalSection(
				'Y', 200000000, 575)
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
	var subseconds int64

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

// Slower with more allocations with a complex period
// 148.0 ns/op    48 B/op   6 allocs/op
// Was
// 209.8 ns/op   160 B/op  16 allocs/op
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
			p, err = period.Parse("P200150M200DT1H4M2000S", true)
		}
	})

	b.Log(p.String())
	is.True(p != period.Period{})
	is.NoErr(err) // Parsing should not have caused an error
}

func BenchmarkParsePeriodLongFractional(b *testing.B) {
	is := is.New(b)

	var p period.Period
	var err error

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p, err = period.Parse("P200000000.5Y", true)
		}
	})

	b.Log(p.String())
	is.True(p != period.Period{})
	is.NoErr(err) // Parsing should not have caused an error
}

// 87.91 ns/op 	 40 B/op   2 allocs/o
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
			p, err = period.Parse("PT1H1.05S", true)
		}
	})

	b.Log(p.String())
	is.True(p != period.Period{})
	is.NoErr(err) // Parsing should not have caused an error
}

// Faster with fewer allocations with a simple period
// 59.95 ns/op	  8 B/op   1 allocs/op
// 41.72 ns/op   16 B/op   2 allocs/op
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
			p, err = period.Parse("P1M1S", true)
		}
	})

	b.Log(p.String())
	is.True(p != period.Period{})
	is.NoErr(err) // Parsing should not have caused an error
}
