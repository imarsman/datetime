// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package date2

import (
	"testing"

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

// func same(d Date, t time.Time) bool {
// 	yd, wd := d.ISOWeek()
// 	yt, wt := t.ISOWeek()
// 	return d.Year() == t.Year() &&
// 		d.Month() == t.Month() &&
// 		d.Day() == t.Day() &&
// 		d.Weekday() == t.Weekday() &&
// 		d.YearDay() == t.YearDay() &&
// 		yd == yt && wd == wt
// }

type dateParts struct {
	y int64
	m int
	d int
}

type datePartsWithVerify struct {
	y int64
	m int
	d int
	v int
}

func TestMaxMinDates(t *testing.T) {
	d := Max()
	t.Log("max year", d.year, "month", d.month, "day", d.day)
	d = Min()
	t.Log("min year", d.year, "month", d.month, "day", d.day)
}
func TestDayOfYear(t *testing.T) {
	is := is.New(t)

	type datePartsWithVerify struct {
		y int64
		m int
		d int
		v int
	}

	var partList = []datePartsWithVerify{
		{-1000, 5, 18, 0},
		// Last day of BCE
		{-1, 12, 31, 365},
		// First day of CE
		{1, 1, 1, 1},
		{2019, 5, 18, 0},
		{2020, 3, 18, 0},
		{2021, 3, 1, 0},
		{2020, 1, 1, 0},
		{2020, 3, 1, 0},
	}

	for _, p := range partList {
		d, err := NewDate(p.y, p.m, p.d)
		dayOfYear, err := d.YearDay()
		if p.v != 0 {
			// is.Equal(p.v, dayOfYear)
		}
		is.NoErr(err)
		t.Log("Days into year", dayOfYear, "for year", d.Year(), "month", d.Month(), "day", d.Day())
	}
}

func TestDaysInMonth(t *testing.T) {
	is := is.New(t)
	type datePartsWithVerify struct {
		y int64
		m int
		d int
		v int
	}

	var partList = []datePartsWithVerify{
		{2020, 1, 18, 31},
		{2020, 2, 18, 29},
		{2021, 2, 18, 28},
	}
	for _, p := range partList {
		d, err := NewDate(p.y, p.m, p.d)
		days, err := d.daysInMonth()
		is.Equal(days, p.v)
		is.NoErr(err)
		t.Log("Days in year", d.year, "month", d.month, "days", days)
	}
}

func TestDaysSinceEpoch(t *testing.T) {
	// is := is.New(t)

	var list = []int64{
		-1, -2, -13, 1, 2,
	}
	for _, y := range list {
		days := daysTo1JanSinceEpoch(y)
		t.Logf("Days since epoch for year %d = %d", y, days)
	}
}

func TestWeekday(t *testing.T) {
	is := is.New(t)

	year := int64(1970)
	var leapYearCount int64
	// if year < 100 {
	leapYearCount = (year / 4) - (year / 100) + (year / 400)
	t.Log(leapYearCount)

	var partList = []dateParts{
		{-15000000000, 1, 1},
		// {-10, 1, 1},
		{-1, 1, 1},
		{1, 1, 1},
		// {1, 1, 2},
		// {2, 1, 1},
		// {3, 1, 1},
		{400, 1, 1},
		// {5, 1, 1},
		// {40, 1, 1},
		// {60, 1, 1},
		// {100, 1, 1},
		{1000, 1, 1},
		// {1020, 1, 1},
		// {1021, 1, 1},
		// {1022, 1, 1},
		// {1023, 1, 1},
		// {1065, 1, 1},
		// {1066, 1, 1},
		// {1100, 1, 1},

		// {1, 1, 1},
		// {1, 1, 7},
		// {1, 1, 21},
		// {1, 1, 22},
		// {10, 1, 1},
		// {14, 1, 1},

		// Problematic
		// {1065, 1, 1},
		// // Problematic
		// {1066, 1, 1},
		// // Problematic
		// {1066, 10, 14},
		// {1900, 1, 1},
		// {1960, 1, 1},
		// {1964, 1, 1},
		// {1967, 1, 1},
		// {1968, 1, 1},
		// {1968, 3, 1},
		// {1969, 1, 1},
		// {1970, 1, 1},
		// {1970, 1, 15},
		// {1970, 2, 1},
		{1970, 1, 1},
		// {1971, 1, 1},
		// {2010, 1, 1},
		// {2018, 1, 1},
		// {2018, 1, 2},
		{2018, 1, 15},
		{2020, 1, 1},
		// {2019, 5, 18},
		// {2020, 5, 18},
		{2030, 1, 1},
		// {2040, 1, 1},
		// {2060, 1, 1},
		// {2061, 1, 1},
		// {3060, 1, 1},
		// {2021, 5, 18},
	}

	var err error
	var d Date

	for _, p := range partList {
		d, err = NewDate(p.y, p.m, p.d)
		dow, err := d.Weekday()
		is.NoErr(err)
		t.Log("Day of week for", d.String(), dow)
	}

	is.NoErr(err)
}

// func TestDoomsday(t *testing.T) {

// 	for i := -10; i < 50; i++ {
// 		t.Log(i, anchorDayForYear(int64(i)))
// 	}
// 	t.Log("2005", anchorDayForYear(2005))
// 	t.Log("1968", anchorDayForYear(1968))
// 	t.Log("2100", anchorDayForYear(2100))
// }

// func TestAnchorDay(t *testing.T) {
// 	is := is.New(t)

// 	type datePartsWithVerify struct {
// 		y int64
// 		m int
// 		d int
// 		v int
// 	}

// 	var partList = []datePartsWithVerify{
// 		{-1000, 1, 1, 0},
// 		{1968, 5, 26, 0},
// 		// This one is wrong
// 		{-3, 1, 1, 0},
// 		{-1, 1, 1, 0},
// 		{1, 1, 1, 0},
// 		// Unix epoch
// 		{1970, 1, 1, 6},
// 		{1998, 8, 14, 6},
// 		{2002, 4, 10, 4},
// 		{2018, 5, 18, 3},
// 		{2019, 5, 18, 4},
// 		{2020, 5, 18, 6},
// 		{2021, 5, 18, 7},
// 	}

// 	var err error
// 	var d Date

// 	for _, p := range partList {
// 		d, err = NewDate(p.y, p.m, p.d)
// 		is.NoErr(err)
// 		// dow, err := d.WeekDay()
// 		anchorDay := anchorDayForYear(d.year)
// 		is.NoErr(err)
// 		// Allow for exploraty use without failing
// 		if p.v != 0 {
// 			is.Equal(anchorDay, p.v)
// 		}
// 		t.Log(d.String(), "anchor day", anchorDay)
// 	}
// }

// func TestWeekDay2(t *testing.T) {
// 	is := is.New(t)

// 	type datePartsWithVerify struct {
// 		y int64
// 		m int
// 		d int
// 		v int
// 	}

// 	var partList = []datePartsWithVerify{
// 		{1066, 10, 14, 0},
// 		{100, 1, 1, 0},
// 		{1000, 1, 1, 0},
// 		// {1300, 1, 1, 0},
// 		// {1400, 1, 1, 0},
// 		{1500, 1, 1, 0},
// 		{1500, 1, 2, 0},
// 		// {1600, 1, 1, 0},
// 		// {1969, 1, 1, 0},
// 		// {1970, 1, 1, 0},
// 		{2000, 1, 1, 0},
// 		// {2021, 1, 1, 0},
// 	}

// 	var err error
// 	var d Date

// 	for _, p := range partList {
// 		d, err = NewDate(p.y, p.m, p.d)
// 		is.NoErr(err)
// 		// weekDay, err := d.WeekDay()
// 		// anchorDay := anchorDay(d.year)
// 		is.NoErr(err)
// 		// Allow for exploraty use without failing
// 		if p.v != 0 {
// 			// is.Equal(anchorDay, p.v)
// 		}
// 		dow, _ := d.DayOfWeek()
// 		t.Log(d.String(), "week day", dow)
// 	}

// 	// d, _ := NewDate(1970, 1, 3)
// 	// t.Logf("days since 1 jan for %s %d", d.String(), d.daysSince1Jan())
// 	// t.Log("day of week", d.dayOfWeek2())
// 	// d, _ = NewDate(1969, 1, 1)
// 	// t.Logf("days since 1 jan for %s %d", d.String(), d.daysSince1Jan())
// 	// t.Log("day of week", d.dayOfWeek2())
// 	// d, _ = NewDate(1990, 1, 1)
// 	// t.Logf("days since 1 jan for %s %d", d.String(), d.daysSince1Jan())
// 	// t.Log("day of week", d.dayOfWeek2())
// 	// d, _ = NewDate(2021, 1, 1)
// 	// t.Logf("days since 1 jan for %s %d", d.String(), d.daysSince1Jan())
// 	// t.Log("day of week", d.dayOfWeek2())
// 	// d, _ = NewDate(1, 1, 1)
// 	// t.Logf("days since 1 jan for %s %d", d.String(), d.daysSince1Jan())
// 	// t.Log("day of week", d.dayOfWeek2())
// }

// func TestWeekDay(t *testing.T) {
// 	is := is.New(t)

// 	type datePartsWithVerify struct {
// 		y int64
// 		m int
// 		d int
// 		v int
// 	}

// 	var partList = []datePartsWithVerify{
// 		// {-1000, 1, 1, 0},
// 		// {1968, 5, 26, 0},
// 		// This one is wrong
// 		// {-3, 1, 1, 0},
// 		{-101, 1, 1, 0},
// 		{-100, 1, 1, 0},
// 		{-99, 1, 1, 0},
// 		{-1, 1, 1, 0},
// 		{1, 1, 1, 0},
// 		// {10, 1, 1, 0},
// 		{99, 1, 1, 0},
// 		{100, 1, 1, 0},
// 		{101, 1, 1, 0},
// 		// {900, 1, 1, 0},
// 		// {1000, 1, 1, 0},
// 		// {1000, 3, 1, 0},
// 		// {1001, 1, 1, 0},
// 		// {1099, 1, 1, 0},
// 		// {1100, 1, 1, 0},
// 		// {15, 1, 1, 0},
// 		// Unix epoch
// 		// {1000, 1, 1, 6},
// 		// {1010, 1, 1, 6},
// 		// {1861, 4, 12, 6},
// 		// {2000, 1, 1, 6},
// 		// {2001, 1, 1, 6},
// 		// {2002, 1, 1, 6},
// 		// {2004, 1, 1, 6},
// 		// {1010, 1, 1, 6},
// 		// {1500, 1, 1, 6},
// 		// {1970, 1, 1, 6},
// 		// {1998, 8, 14, 6},
// 		// {2005, 4, 10, 4},
// 		// {2018, 5, 18, 3},
// 		// {2019, 5, 18, 4},
// 		// {2020, 5, 18, 6},
// 		// {2040, 5, 18, 6},
// 		// {2102, 1, 1, 6},
// 		// {3000, 1, 1, 6},
// 		// {2021, 4, 1, 7},
// 		{4000, 1, 1, 0},
// 		{5000, 1, 1, 0},
// 		{5010, 1, 1, 0},
// 		/*
// 		   closest 1 final 5 anchor day 7
// 		   new new 2
// 		       date_test.go:246: 2021-05-01 week day 2
// 		*/
// 		// {2021, 5, 1, 7},
// 		// {2021, 5, 18, 7},
// 	}

// 	var err error
// 	var d Date

// 	for _, p := range partList {
// 		d, err = NewDate(p.y, p.m, p.d)
// 		is.NoErr(err)
// 		weekDay, err := d.WeekDay()
// 		// anchorDay := anchorDay(d.year)
// 		is.NoErr(err)
// 		// Allow for exploraty use without failing
// 		if p.v != 0 {
// 			// is.Equal(anchorDay, p.v)
// 		}
// 		t.Log(d.String(), "week day", weekDay)
// 	}
// }

func TestGetCenturyAnchorDay(t *testing.T) {
	years := []int64{0, -1, -100, -200, -300, -1000, 112, 200, 1942, 2000, 1763}

	for i := 0; i < len(years); i++ {
		d := anchorDayForCentury(years[i])
		t.Logf("Got %d for %d", d, years[i])
	}

}

func TestWeekdayCount(t *testing.T) {
	// Jan 2021
	day := countDaysForward(5, 30)
	t.Log("30 days forward from 1 Jan with start day of 5", day)
	day = countDaysForward(1, 27)
	t.Log("27 days forward from 1 Feb with start day of 1", day)
	day = countDaysBackward(7, 30)
	t.Log("30 days Backward from 31 Jan with start day of 7", day)
	day = countDaysBackward(7, 27)
	t.Log("27 days Backward from 28 Feb with start day of 7", day)
}

func TestGetBigYear(t *testing.T) {
	is := is.New(t)

	years := []int64{0, -1, -4, -5, -101, -201, -301, -401, -1001, -2001, 1, 4, 112, 200, 400, 1942, 2000, 1763}
	d, err := NewDate(1, 1, 1)
	is.NoErr(err)
	for i := 0; i < len(years); i++ {
		d.year = years[i]
		bigYear := d.getBigYear()
		isLeap := d.IsLeap()
		aYear := astronomicalYear(d.year)
		t.Logf("Got big year %-5d for year %-5d astronomical year %-5d Is leap %-5v", bigYear, d.year, aYear, isLeap)
	}

}

func TestIsLeap(t *testing.T) {
	is := is.New(t)

	type yearWithVerify struct {
		y int64
		v bool
	}

	tests := []yearWithVerify{
		// {-4, true},
		// {-5, false},
		// // Consider what to do with year zero
		// {0, true},
		// {-1, false},
		{1000, false},
		{2000, true},
		{3000, false},
		{1984, true},
		{2004, true},
	}

	for _, item := range tests {
		d, err := NewDate(item.y, 1, 1)
		is.NoErr(err)
		isLeap := d.IsLeap()
		// Comment out to try more
		is.Equal(isLeap, item.v)
		is.Equal(true, true)
		t.Logf("mathematical year %-5d gregorian year %-5d isLeap %-5v verify %-5v", d.astronomicalYear(), d.year, isLeap, item.v)
	}
}

// TODO: verify this works with BCE years
func TestWeekNumer(t *testing.T) {

	iterate := func(t *testing.T, start, end int64) {
		is := is.New(t)
		is.True(start < end)

		for i := start; i < end; i++ {
			d, _ := NewDate(i, 1, 1)
			weeks := d.ISOWeeksInYear()
			// if weeks == 53 {
			t.Log("weeks in year", d.year, weeks)
			// }
		}
	}

	t.Log("Printing all years with 53 weeks. The rest have 52")
	iterate(t, -50, -1)
	iterate(t, 0, 50)
	iterate(t, 1000, 1010)
	iterate(t, 2000, 2021)
}

// func TestFromGregorianYear(t *testing.T) {
// 	tests := []int64{
// 		-1000,
// 		1939,
// 		1970,
// 		1960,
// 	}

// 	for i := 0; i < len(tests); i++ {
// 		year := fromGregorianYear(tests[i])
// 		t.Log(tests[i], "year", year)
// 	}
// }

func TestToGregorianYear(t *testing.T) {
	tests := []int64{
		0,
		-1000,
		1939,
		1970,
		1960,
	}

	for i := 0; i < len(tests); i++ {
		// year := fromGregorianYear(tests[i])
		year := gregorianYear(tests[i])
		t.Log(tests[i], "year", year, "isCE")
	}
}

func TestAddDate(t *testing.T) {
	type datePartsWithAdd struct {
		y    int64
		m    int
		d    int
		addY int64
		addM int
		addD int
	}

	var partList = []datePartsWithAdd{
		{y: 1968, m: 05, d: 26, addY: 52, addM: 0, addD: 0},
		{y: 2002, m: 4, d: 10, addY: 52, addM: 0, addD: 0},
		{y: 2019, m: 3, d: 1, addY: 3, addM: 10, addD: 1},
		{y: 2019, m: 1, d: 1, addY: 1, addM: 0, addD: 0},
		{y: 2019, m: 1, d: 1, addY: 0, addM: 12, addD: 0},
		{y: 2019, m: 1, d: 1, addY: 14, addM: 0, addD: 0},
		{y: 2019, m: 1, d: 1, addY: 1000, addM: 0, addD: 0},
	}

	for _, dt := range partList {
		d, _ := NewDate(dt.y, dt.m, dt.d)
		t.Logf("Pre %s", d.String())
		t.Logf("Add %d years,  %d months, %d days", dt.addY, dt.addM, dt.addD)
		d, _, _ = d.AddParts(dt.addY, dt.addM, dt.addD)
		t.Logf("Post %s", d.String())
	}
}

func TestSubtractDays(t *testing.T) {
	d, _ := NewDate(2000, 2, 15)
	t.Log(d.String())
	d, _ = d.subtractDays(32)
	t.Log(d.String())

	is := is.New(t)

	var partList = []dateParts{
		// {-1, 10, 1},
		// {-1, 12, 1},
		{2021, 1, 30},
	}

	for _, p := range partList {
		d, err := NewDate(p.y, p.m, p.d)
		d2, _ := d.subtractDays(30)
		is.NoErr(err)
		// if p.v != 0 {
		// 	is.Equal(daysToEnd, p.v)
		// }
		t.Log("Date", d.String(), "date 2", d2.String())
	}

}

func TestDaysSinceYearEnd(t *testing.T) {
	is := is.New(t)

	var partList = []datePartsWithVerify{
		{-1, 10, 1, 91},
		{-1, 12, 1, 30},
		{2021, 1, 30, 29},
	}

	for _, p := range partList {
		d, err := NewDate(p.y, p.m, p.d)
		daysToEnd := d.daysToYearEnd()
		is.NoErr(err)
		if p.v != 0 {
			is.Equal(daysToEnd, p.v)
		}
		t.Log("Date", d.String(), "to year end", daysToEnd)
	}
}

func TestString(t *testing.T) {
	var partList = []dateParts{
		{2018, 3, 1},
		{2019, 3, 1},
		{2020, 3, 1},
		{2021, 3, 1},
		{1, 3, 1},
		{1000000, 3, 1},
		{-1000000, 3, 1},
	}

	// var err error
	// var d Date

	for _, p := range partList {
		d, _ := NewDate(p.y, p.m, p.d)
		t.Log(d.String())
	}
}

func BenchmarkAddPartsLong(b *testing.B) {
	is := is.New(b)

	var err error
	var d Date

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			d, _ = NewDate(2019, 3, 1)
			d, _, err = d.AddParts(100000000, 0, 0)
		}
	})

	b.Log("caculated", d.String())
	is.NoErr(err)
}

func BenchmarkAddPartsShort(b *testing.B) {
	is := is.New(b)

	var err error
	var d Date

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			d, _ = NewDate(2019, 3, 1)
			d, _, err = d.AddParts(1, 3, 12)
		}
	})

	b.Log("caculated", d.String())
	is.NoErr(err)
}

func BenchmarkDaysSinceEpoch(b *testing.B) {
	is := is.New(b)

	var days uint64
	var year int64

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			days = daysTo1JanSinceEpoch(year)
		}
	})

	b.Logf("Days since epoch to %d", days)
	is.True(days != 0)
}

func BenchmarkDayOfWeek1Jan(b *testing.B) {
	is := is.New(b)

	d, err := NewDate(2020, 3, 1)
	is.NoErr(err)

	var dow int

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			dow, err = d.Weekday()
		}
	})

	is.NoErr(err)
	b.Log("day of week for 1 January,", d.year, dow)
	is.True(dow != 0)
}

//  2.220 ns/op	  0 B/op   0 allocs/op
func BenchmarkDaysInMonth(b *testing.B) {
	is := is.New(b)
	var err error

	d, err := NewDate(2020, 2, 1)
	is.NoErr(err)

	var dayOfYear int

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			dayOfYear, err = d.daysInMonth()
		}
	})

	is.NoErr(err)
	b.Log("days in month", dayOfYear)
	is.True(dayOfYear != 0)
}

// 8.101 ns/op   0 B/op	  0 allocs/op
func BenchmarkDayOfYear(b *testing.B) {
	is := is.New(b)
	var err error

	d, err := NewDate(2020, 3, 1)
	is.NoErr(err)

	var dayOfYear int

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			dayOfYear, err = d.YearDay()
		}
	})
	is.NoErr(err)

	b.Log("day of year", dayOfYear)
	is.True(dayOfYear != 0)
}

// 10.25 ns/op   0 B/op   0 allocs/op
// func BenchmarkDayOfWeek(b *testing.B) {
// 	is := is.New(b)
// 	var err error

// 	d, err := NewDate(2020, 3, 1)
// 	is.NoErr(err)
// 	var dayOfWeek int

// 	b.ResetTimer()
// 	b.SetBytes(bechmarkBytesPerOp)
// 	b.ReportAllocs()
// 	b.SetParallelism(30)
// 	b.RunParallel(func(pb *testing.PB) {
// 		for pb.Next() {
// 			dayOfWeek, err = d.WeekDay()
// 		}
// 	})
// 	is.NoErr(err)

// 	is.True(dayOfWeek != 0)
// }

// 11.81 ns/op   0 B/op	  0 allocs/op
func BenchmarkNewDate(b *testing.B) {
	is := is.New(b)
	var err error
	var d Date
	is.NoErr(err)

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			d, err = NewDate(2020, 3, 1)
		}
	})
	is.NoErr(err)
	is.True(d.IsZero() == false)
	b.Log(d)
}

func BenchmarkWeekday(b *testing.B) {
	is := is.New(b)
	var err error
	d, err := NewDate(2020, 3, 1)
	is.NoErr(err)
	var val int

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			val, _ = d.Weekday()
		}
	})
	is.True(val > 0)
	is.NoErr(err)
	is.True(d.IsZero() == false)
	b.Log(d, val)
}
