// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package date

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
		-1, -2, -13, 1, 2, 4, 20, 40, 400,
	}
	for _, y := range list {
		days := daysToAnchorDaySinceEpoch(y)
		t.Logf("Days since epoch for year %d = %d", y, days)
	}
}

// https://www.thecalculatorsite.com/time/days-between-dates.php
func TestWeekday(t *testing.T) {
	is := is.New(t)

	year := int64(1970)
	var leapYearCount int64
	// if year < 100 {
	leapYearCount = (year / 4) - (year / 100) + (year / 400)
	t.Log(leapYearCount)

	var partList = []dateParts{
		// {-20, 1, 1},
		// {-10, 1, 1},
		// {-1, 12, 1},
		// {-1, 12, 2},
		// {-1, 12, 3},
		// {-1, 12, 4},
		// {-1, 12, 5},
		// {-1, 12, 6},
		// {-1, 12, 7},
		// {-1, 12, 8},
		// {-1, 12, 12},
		// {-1, 12, 13},
		// {-1, 12, 14},
		// {-1, 12, 15},
		// {-1, 12, 16},
		// {-1, 12, 17},
		// {-1, 12, 18},
		// {-1, 12, 19},
		// {-1, 12, 20},
		// {-1, 12, 21},
		// {-1, 12, 22},
		// {-1, 12, 23},
		// {-1, 12, 24},
		// {-1, 12, 25},
		// {-1, 12, 26},
		// {-1, 12, 27},
		// {-1, 12, 28},
		// {-1, 12, 29},
		// {-1, 12, 30},
		// {-1, 12, 31},
		{1, 1, 1},
		{20, 1, 1},
		{40, 1, 1},

		{1000, 1, 1},

		{1867, 7, 1},
		{1967, 7, 1},
		{1968, 5, 26},
		{1998, 8, 14},
		{2002, 4, 10},
		{2000, 1, 1},
		{2020, 1, 1},
		{2020, 5, 18},
		{2040, 1, 1},
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

// func TestGetBigYear(t *testing.T) {
// 	is := is.New(t)

// 	years := []int64{0, -1, -4, -5, -101, -201, -301, -401, -1001, -2001, 1, 4, 112, 200, 400, 1942, 2000, 1763}
// 	d, err := NewDate(1, 1, 1)
// 	is.NoErr(err)
// 	for i := 0; i < len(years); i++ {
// 		d.year = years[i]
// 		bigYear := d.getBigYear()
// 		isLeap := d.IsLeap()
// 		aYear := astronomicalYear(d.year)
// 		t.Logf("Got big year %-5d for year %-5d astronomical year %-5d Is leap %-5v", bigYear, d.year, aYear, isLeap)
// 	}

// }

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

func TestDaysToYearEnd(t *testing.T) {
	is := is.New(t)

	var partList = []datePartsWithVerify{
		{-1, 10, 1, 91},
		{-1, 12, 1, 30},
		{1, 12, 1, 0},
		{2021, 1, 30, 29},
	}

	for _, p := range partList {
		d, err := NewDate(p.y, p.m, p.d)
		daysToEnd := d.daysToDateFromAnchorDay()
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
			days = daysToAnchorDaySinceEpoch(year)
		}
	})

	b.Logf("Days since epoch to %d", days)
	is.True(days != 0)
}

// func BenchmarkWeekday(b *testing.B) {
// 	is := is.New(b)

// 	d, err := NewDate(2020, 3, 1)
// 	is.NoErr(err)

// 	var dow int

// 	b.ResetTimer()
// 	b.SetBytes(bechmarkBytesPerOp)
// 	b.ReportAllocs()
// 	b.SetParallelism(30)
// 	b.RunParallel(func(pb *testing.PB) {
// 		for pb.Next() {
// 			dow, err = d.Weekday()
// 		}
// 	})

// 	is.NoErr(err)
// 	b.Log("day of week for 1 January,", d.year, dow)
// 	is.True(dow != 0)
// }

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

// 10.98 ns/op      0 B/op   0 allocs/op
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
