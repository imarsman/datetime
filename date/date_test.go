// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package date

import (
	"fmt"
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
		dayOfYear := d.YearDay()
		if p.v != 0 {
			// is.Equal(p.v, dayOfYear)
		}
		is.NoErr(err)
		t.Log("Days into year", dayOfYear, "for year", d.Year(), "month", d.Month(), "day", d.Day())
	}
}

func TestDaysToDateFromAnchorDay(t *testing.T) {
	is := is.New(t)
	type datePartsWithVerify struct {
		y int64
		m int
		d int
		v int
	}

	var partList = []datePartsWithVerify{
		{-6, 3, 2, 0},
		{-6, 1, 2, 0},
		{-5, 1, 2, 0},
		{-5, 3, 2, 0},
		{-1, 3, 2, 0},
		// {-4, 4, 1, 0},
		// {-5, 4, 1, 0},
		// {-4, 1, 31, 0},
		// {-5, 1, 31, 0},
		// {-5, 8, 31, 0},
		// {-4, 8, 31, 0},
		// {-298, 10, 04, 0},
		// {-298, 10, 03, 0},

		// {-800, 12, 31, 0},
		// {-10, 1, 1, 0},
		// {-10, 1, 2, 0},
		// {-9, 12, 1, 0},
		// {-5, 12, 1, 0},
		// {-6, 12, 1, 0},
		// {-1, 1, 18, 0},
		// {-1, 12, 1, 0},
		// {-2, 12, 31, 0},
		// {-4, 12, 1, 0},
		// {-5, 12, 30, 0},
		// {-5, 3, 31, 0},
		// {-5, 1, 1, 365},
		// {-2, 1, 1, 30},
		// {800, 1, 1, 0},
		// {401, 2, 1, 0},
		// {401, 4, 30, 0},

		// {8, 12, 1, 0},
		// {4, 2, 28, 0},
		// {4, 12, 1, 0},
		// {1, 12, 1, 0},
		// {4, 1, 1, 0},
		// {4, 1, 31, 0},
		// {3, 5, 1, 0},
		// {3, 2, 28, 0},
		// {5, 1, 1, 0},
		// {400, 1, 1, 0},
		// {800001, 12, 31, 0},
	}
	for _, p := range partList {
		d, err := NewDate(p.y, p.m, p.d)
		is.NoErr(err)
		daysToAnchor := daysToAnchorDayFromEpoch(d.year)
		days := d.daysToDateFromAnchorDay()
		t.Log(d, "days to anchor date", daysToAnchor, "days to date", days)
		t.Log(d, "days", days)
		if p.v != 0 {
			// is.Equal(days, p.v)
		}
		sum := daysToAnchor + int64(days)
		t.Log("Days from anchor date to date", days, "days to anchor", daysToAnchor, "total from epoch", sum, "for", d.String())
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
		is.NoErr(err)
		days := d.DaysInMonth()
		is.Equal(days, p.v)
		t.Log("Days in year", d.year, "month", d.month, "days", days)
	}
}

func TestDaysSinceEpoch(t *testing.T) {
	// is := is.New(t)

	var list = []int64{
		-1, -2, -13, 1, 2, 4, 20, 40, 400,
	}
	for _, y := range list {
		days := daysToAnchorDayFromEpoch(y)
		t.Logf("Days since epoch for year %d = %d", y, days)
	}
}

// https://www.thecalculatorsite.com/time/days-between-dates.php
func TestWeekday(t *testing.T) {
	is := is.New(t)

	year := int64(1970)
	var leapYearCount int64
	leapYearCount = (year / 4) - (year / 100) + (year / 400)
	t.Log(leapYearCount)

	var partList = []dateParts{
		{-2, 12, 31},
		{-1, 12, 10},
		{-1, 12, 17},
		{-1, 12, 24},
		{-1, 12, 31},
		{-1, 1, 1},
		{1, 1, 1},
		{2, 12, 31},
		{20, 1, 1},
		{40, 1, 1},
		{1000, 1, 1},
		{1867, 7, 1},
		{1967, 7, 1},
		{1968, 5, 26},
		{1998, 8, 14},
		{2002, 4, 10},
		{1985, 7, 30},
		{2000, 1, 1},
		{2020, 1, 1},
		{2020, 5, 18},
		{2040, 1, 1},
	}

	var err error
	var d Date

	for _, p := range partList {
		d, err = NewDate(p.y, p.m, p.d)
		dow := d.Weekday()
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

func TestDaysTo(t *testing.T) {
	is := is.New(t)

	d1, err := NewDate(2019, 6, 1)
	is.NoErr(err)
	d2, err := NewDate(2020, 6, 1)
	daysBetween := d1.DaysTo(d2)

	t.Log("days between", daysBetween)
}

func TestIsLeap(t *testing.T) {
	is := is.New(t)

	type yearWithVerify struct {
		y int64
		v bool
	}

	tests := []yearWithVerify{
		{-5, true},
		{1000, false},
		{2000, true},
		{3000, false},
		{1984, true},
		{2004, true},
		{8000, true},
	}

	for _, item := range tests {
		d, err := NewDate(item.y, 1, 1)
		is.NoErr(err)
		isLeap := d.IsLeap()
		// Comment out to try more
		is.Equal(isLeap, item.v)
		is.Equal(true, true)
		t.Logf("mathematical year %-5d gregorian year %-5d isLeap %-5v verify %-5v", d.AstronomicalYear(), d.year, isLeap, item.v)
	}
}

func TestWeekOfYear(t *testing.T) {
	is := is.New(t)
	d, err := NewDate(2022, 7, 30)
	is.NoErr(err)

	woy := d.ISOWeekOfYear()
	t.Log("week of year for", d, woy)
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

func TestToGregorianYear(t *testing.T) {
	tests := []int64{
		0,
		-1000,
		1939,
		1970,
		1960,
	}

	for i := 0; i < len(tests); i++ {
		year := gregorianYear(tests[i])
		t.Log(tests[i], "year", year, "isCE")
	}
}

func TestAddParts(t *testing.T) {
	is := is.New(t)

	type datePartsWithAdd struct {
		y    int64
		m    int
		d    int
		addY int64
		addM int
		addD int
	}

	var partList = []datePartsWithAdd{
		{y: 1968, m: 5, d: 26, addY: 52, addM: 0, addD: 0},
		{y: 2002, m: 4, d: 10, addY: 52, addM: 0, addD: 0},
		{y: 2019, m: 3, d: 1, addY: 3, addM: 10, addD: 1},
		{y: 2019, m: 1, d: 1, addY: 1, addM: 0, addD: 0},
		{y: 2019, m: 1, d: 1, addY: 0, addM: 12, addD: 0},
		{y: 2019, m: 1, d: 1, addY: 14, addM: 0, addD: 0},
		{y: 2019, m: 1, d: 1, addY: 10, addM: 0, addD: 0},
		{y: 2019, m: 1, d: 1, addY: 1000, addM: 0, addD: 0},
		{y: 2019, m: 1, d: 1, addY: 0, addM: 1000, addD: 0},
		{y: 2019, m: 6, d: 1, addY: 0, addM: 0, addD: 1000000},
		{y: 2018, m: 6, d: 1, addY: 0, addM: 0, addD: 365},
		{y: 2019, m: 6, d: 1, addY: 0, addM: 0, addD: 365},
		// Covers a leap year. End date shoold have same month and day as start.
		{y: 2020, m: 2, d: 1, addY: 0, addM: 0, addD: 366},
		{y: -4, m: 1, d: 1, addY: 10, addM: 0, addD: 0},
		{y: -10, m: 1, d: 1, addY: 1, addM: 0, addD: 0},
	}

	for _, dt := range partList {
		d, err := NewDate(dt.y, dt.m, dt.d)
		is.NoErr(err)
		t.Logf("Pre %s", d.String())
		t.Logf("Add %d years,  %d months, %d days", dt.addY, dt.addM, dt.addD)
		d, _, _ = d.AddParts(dt.addY, dt.addM, dt.addD)
		t.Logf("Post %s", d.String())
	}
}

func TestAddDays(t *testing.T) {
	is := is.New(t)

	// date_test.go:413: not equal - new date -101-12-31 date from days -101-12-29 days to date 36524 starting date -100-12-31
	// date_test.go:413: not equal - new date -101-12-31 date from days -101-12-29 days to date 36524 starting date -100-12-31
	// Ensure that cyles go through all probably multiples of 400
	dayCount := 1 * 40
	d, err := NewDate(-297, 10, 1)
	var newDate Date
	is.NoErr(err)
	startDate := d
	t.Log("Trying a series of", dayCount, "days starting with year", d.year)
	foundErr := false
	var lastDay Date
	var errorsFound int = 0
	var endDate Date

	for i := 0; i < dayCount; i++ {
		if d.year < 0 {
			d, err = NewDate(startDate.year, 11, 1)
		} else {
			d, err = NewDate(startDate.year, 1, 1)
		}
		is.NoErr(err)
		addedDays, err := d.AddDays(int64(i + 1))
		is.NoErr(err)
		daysToDate := addedDays.daysToDateFromEpoch()
		newDate, err = FromDays(daysToDate, d.year > 0)
		// t.Log("adding", i+1, "days to", d, "new date from adding", newDate, "days to date", daysToDate, "new date from days", newDate)
		is.NoErr(err)
		lastDay = newDate
		if i > 999 && (i+1)%100000 == 0 {
			t.Logf("On day %d of %d date %s...\n", i+1, dayCount, newDate)
		}
		if addedDays.String() != newDate.String() {
			t.Log("not equal - new date", addedDays.String(), "added days", i+1, "date from days", newDate.String(), "days to date", daysToDate, "starting date", d)
			foundErr = true
			errorsFound++
		}
		endDate = newDate
	}
	if foundErr == false {
		t.Log("no errors found for", dayCount, "day increments. Last date", lastDay.String(), "years from", startDate, "to", endDate)
	} else {
		t.Log("Found", errorsFound, "errors out of ", dayCount, "Check from", startDate, "to", endDate)
	}
	fmt.Println("Day check finished!")
}

func TestDaysToDateParts(t *testing.T) {
	is := is.New(t)

	d, err := NewDate(-8, 3, 2)
	// d, err := NewDate(-6, 3, 2)
	// d, err := NewDate(-1, 3, 2)
	// d, err := NewDate(-1, 3, 2)
	// d, err := NewDate(-5, 3, 2)
	is.NoErr(err)
	t.Log(d, "days to date from anchor day", d.daysToDateFromAnchorDay())
	daysToDate := d.daysToDateFromEpoch()
	ce := d.year > 0
	newDate, err := FromDays(daysToDate, ce)
	is.NoErr(err)
	t.Log("date", d, "new date", newDate)
}

func TestDateFromDays(t *testing.T) {
	is := is.New(t)

	var partList = []datePartsWithVerify{
		// {-1, 1, 31, 0},
		// {401, 2, 1, 0},
		// {401, 6, 1, 0},
		// {801, 1, 1, 0},
		// {1201, 1, 1, 0},
		// {1201, 9, 1, 0},
		// {-298, 10, 04, 0},
		// {-2800, 1, 1, 0},
		// {-1600, 10, 4, 0},
		// {-1200, 10, 3, 0},
		// {-800, 6, 3, 0},
		// {-800, 2, 3, 0},
		// {-800, 10, 3, 0},
		// {-401, 6, 3, 0},
		// {-401, 2, 3, 0},
		// {-400, 10, 3, 0},
		// {-400, 2, 3, 0},
		// {-400, 6, 3, 0},
		// {-298, 10, 3, 0},
		// {-298, 10, 10, 0},
		{-1, 3, 2, 0},
		{-1, 4, 2, 0},
		// {-1, 4, 20, 0},
		// {-1, 5, 2, 0},
		// {-1, 5, 1, 0},
		// {-1, 5, 31, 0},
		// {-1, 10, 1, 0},
		// {-1, 12, 30, 0},
		// {101, 1, 1, 0},
		// {201, 1, 1, 0},
		// {400, 1, 1, 0},
		// {500, 1, 1, 0},
		// {600, 1, 1, 0},
		// {401, 1, 1, 0},
		// {2000, 1, 1, 0},
		// Incorrect
		{2021, 6, 10, 0},
		{2400, 1, 1, 0},
		{3000, 1, 1, 0},
		{4000, 1, 1, 0},
		{10001, 1, 1, 0},
		// Incorrect
		{10400, 2, 1, 0},
		// Incorrect
		{10400, 2, 2, 0},
		// Incorrect
		{10400, 6, 2, 0},
		// Incorrect
		{10401, 6, 2, 0},
	}

	for _, p := range partList {
		d, err := NewDate(p.y, p.m, p.d)
		daysToDate := d.daysToDateFromEpoch()
		ce := d.year > 0
		newDate, err := FromDays(daysToDate, ce)
		is.NoErr(err)
		if d.String() != newDate.String() {
			t.Log("not equal", d.String(), newDate.String())
			// is.Equal(d.String(), newDate.String())
		}
		t.Logf("starting date %-12s days to date %-12d date from days %-10s", d.String(), daysToDate, newDate.String())
	}

}

// Test date on day
func TestOnDay(t *testing.T) {
	is := is.New(t)

	type ydverify struct {
		y int64
		d int
	}

	var ymdList = []ydverify{
		{2021, 10},
		{2021, 100},
		{2021, 364},
	}

	for _, p := range ymdList {
		d, err := OnDay(p.y, p.d)
		is.NoErr(err)
		t.Log("Date", d.String())
	}

}

func TestDaysToOneYear(t *testing.T) {
	var partList = []datePartsWithVerify{
		{2019, 3, 1, 366},
		{2019, 1, 1, 365},
		{2020, 12, 1, 365},
		{2020, 1, 1, 366},
		{-1, 12, 31, 365},
		{-4, 12, 31, 365},
		{-4, 2, 28, 366},
		{-4, 1, 28, 366},
	}

	for _, p := range partList {
		d, _ := NewDate(p.y, p.m, p.d)
		daysToOneYear := d.DaysOneYearFromDate()
		t.Log(d.String(), "days to one year from", daysToOneYear)
	}
}

func TestDaysToDate(t *testing.T) {
	is := is.New(t)

	var partList = []datePartsWithVerify{
		// {-400, 2, 3, 0},
		// {-401, 2, 3, 0},
		// {-1, 2, 3, 0},
		// {-1, 2, 28, 0},
		// {-1, 1, 1, 0},
		// {-1, 1, 31, 0},
		// {-1, 10, 1, 91},
		// {-1, 12, 1, 30},
		// {-1, 12, 30, 30},
		// {1, 12, 1, 334},
		// {590, 2, 26, 0},
		{2020, 3, 11, 0},
		{2020, 4, 9, 0},
		{2020, 4, 11, 0},
		// {2020, 4, 12, 0},
		// {2001, 1, 1, 0},
		// {2001, 1, 2, 0},
		// {2001, 1, 3, 0},
		// {2001, 2, 3, 0},
		// {2021, 1, 30, 737819},
		// {5107, 2, 28, 0},
		// {10401, 6, 2, 0},
	}

	for _, p := range partList {
		d, err := NewDate(p.y, p.m, p.d)
		daysToEnd := d.daysToDateFromEpoch()
		is.NoErr(err)
		// if uint64(p.v) != 0 {
		// 	is.Equal(daysToEnd, uint64(p.v))
		// }
		t.Log("Date", d.String(), "days to date", daysToEnd)
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

	for _, p := range partList {
		d, _ := NewDate(p.y, p.m, p.d)
		t.Log(d.String())
	}
}

// 21.90 ns/op   16 B/op   1 allocs/op
func BenchmarkDateFromDays(b *testing.B) {
	is := is.New(b)
	var d Date

	d, err := NewDate(800000, 12, 21)
	// d, err := NewDate(800, 12, 1)
	// d, err := NewDate(2001, 12, 31)
	is.NoErr(err)
	ce := d.year > 0
	d2 := d
	daysToDate := d.daysToDateFromEpoch()

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			d2, err = FromDays(daysToDate, ce)
			is.NoErr(err)
		}
	})
	b.Log("starting date", d, "new date", d2)

}

func BenchmarkAddPartsLongMonths(b *testing.B) {
	is := is.New(b)

	var err error
	var d Date
	d, _ = NewDate(2019, 1, 1)
	d2 := Date{}
	b.Log("starting date", d.String())
	b.Logf("Add %d years %d months %d days", 0, 1000, 0)

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			d2, _, err = d.AddParts(0, 1000, 0)
		}
	})

	b.Log("caculated", d2.String())
	is.NoErr(err)
}

func BenchmarkAddPartsLongDays(b *testing.B) {
	is := is.New(b)

	var err error
	var d Date
	d, _ = NewDate(2019, 3, 1)
	d2 := d
	b.Log("starting date", d.String())
	b.Logf("Add %d years %d months %d days", 0, 0, 1000000)

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			d2, _, err = d.AddParts(0, 0, 1000000)
		}
	})

	b.Log("caculated", d2.String())
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

// 4.297 ns/op   0 B/op	  0 allocs/op
func BenchmarkDaysSinceEpoch(b *testing.B) {
	is := is.New(b)

	var days int64
	var year int64

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			days = daysToAnchorDayFromEpoch(year)
		}
	})

	b.Logf("Days since epoch to %d", days)
	is.True(days != 0)
}

//  7.621 ns/op   0 B/op   0 allocs/op
func BenchmarkDaysToOneYearFromDate(b *testing.B) {
	is := is.New(b)

	d, err := NewDate(2020, 1, 1)
	is.NoErr(err)
	var days int

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			days = d.DaysOneYearFromDate()
		}
	})

	b.Logf("Days to one year from %s %d", d.String(), days)
	is.True(days != 0)
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
			dayOfYear = d.DaysInMonth()
		}
	})

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
			dayOfYear = d.YearDay()
		}
	})
	is.NoErr(err)

	b.Log("day of year for", d.String(), dayOfYear)
	is.True(dayOfYear != 0)
}

func BenchmarkFromDays(b *testing.B) {
	is := is.New(b)
	var err error

	d, err := NewDate(-2020, 1, 1)
	var offset int64 = 365
	var d2 Date
	is.NoErr(err)

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			d2, err = d.AddDays(offset)
			is.NoErr(err)
		}
	})
	is.NoErr(err)

	b.Log("Starting date", d.String(), offset, "days since", d2.String())
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
	d, err := NewDate(200000, 3, 1)
	is.NoErr(err)
	var val int

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			val = d.Weekday()
		}
	})
	is.True(val > 0)
	is.NoErr(err)
	is.True(d.IsZero() == false)
	b.Log(d, val)
}
