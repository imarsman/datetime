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
		-1, 1000, 2000, absoluteZeroYear + 10,
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
		// {-10, 1, 1},
		// {-1, 1, 1},
		// {1, 1, 1},
		{1, 1, 2},
		// {2, 1, 1},
		// {3, 1, 1},
		{400, 1, 1},
		// {5, 1, 1},
		// {40, 1, 1},
		// {60, 1, 1},
		// {100, 1, 1},
		// {1000, 1, 1},
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
		{1, 1, 22},
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
		// {1970, 1, 1},
		// {1971, 1, 1},
		// {2010, 1, 1},
		// {2018, 1, 1},
		// {2018, 1, 2},
		{2018, 1, 15},
		// {2020, 1, 1},
		// {2019, 5, 18},
		// {2020, 5, 18},
		// {2030, 1, 1},
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

func TestGetBigYear(t *testing.T) {
	is := is.New(t)

	years := []int64{0, -1, -5, -101, -201, -301, -401, -1001, -2001, 112, 200, 1942, 2000, 1763}
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

// func TestNew(t *testing.T) {
// 	cases := []string{
// 		"0000-01-01T00:00:00+00:00",
// 		"0001-01-01T00:00:00+00:00",
// 		"1614-01-01T01:02:03+04:00",
// 		"1970-01-01T00:00:00+00:00",
// 		"1815-12-10T05:06:07+00:00",
// 		"1901-09-10T00:00:00-05:00",
// 		"1998-09-01T00:00:00-08:00",
// 		"2000-01-01T00:00:00+00:00",
// 		"9999-12-31T00:00:00+00:00",
// 	}
// 	for _, c := range cases {
// 		tIn, err := time.Parse(time.RFC3339, c)
// 		if err != nil {
// 			t.Errorf("New(%v) cannot parse input: %v", c, err)
// 			continue
// 		}
// 		dOut := New(tIn.Year(), tIn.Month(), tIn.Day())
// 		if !same(dOut, tIn) {
// 			t.Errorf("New(%v) == %v, want date of %v", c, dOut, tIn)
// 		}
// 		dOut = NewAt(tIn)
// 		if !same(dOut, tIn) {
// 			t.Errorf("NewAt(%v) == %v, want date of %v", c, dOut, tIn)
// 		}
// 	}
// }

// func TestDaysSinceEpoch(t *testing.T) {
// 	zero := Date{}.DaysSinceEpoch()
// 	if zero != 0 {
// 		t.Errorf("Non zero %v", zero)
// 	}
// 	today := Today()
// 	days := today.DaysSinceEpoch()
// 	copy1 := NewOfDays(days)
// 	copy2 := days.Date()
// 	if today != copy1 || days == 0 {
// 		t.Errorf("Today == %v, want date of %v", today, copy1)
// 	}
// 	if today != copy2 || days == 0 {
// 		t.Errorf("Today == %v, want date of %v", today, copy2)
// 	}
// }

// func TestToday(t *testing.T) {
// 	today := Today()
// 	now := time.Now()
// 	if !same(today, now) {
// 		t.Errorf("Today == %v, want date of %v", today, now)
// 	}
// 	today = TodayUTC()
// 	now = time.Now().UTC()
// 	if !same(today, now) {
// 		t.Errorf("TodayUTC == %v, want date of %v", today, now)
// 	}
// 	cases := []int{-10, -5, -3, 0, 1, 4, 8, 12}
// 	for _, c := range cases {
// 		location := time.FixedZone("zone", c*60*60)
// 		today = TodayIn(location)
// 		now = time.Now().In(location)
// 		if !same(today, now) {
// 			t.Errorf("TodayIn(%v) == %v, want date of %v", c, today, now)
// 		}
// 	}
// }

// func TestTime(t *testing.T) {
// 	cases := []struct {
// 		d Date
// 	}{
// 		{New(-1234, time.February, 5)},
// 		{New(0, time.April, 12)},
// 		{New(1, time.January, 1)},
// 		{New(1946, time.February, 4)},
// 		{New(1970, time.January, 1)},
// 		{New(1976, time.April, 1)},
// 		{New(1999, time.December, 1)},
// 		{New(1111111, time.June, 21)},
// 	}
// 	zones := []int{-12, -10, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 8, 12}
// 	for _, c := range cases {
// 		d := c.d
// 		tUTC := d.UTC()
// 		if !same(d, tUTC) {
// 			t.Errorf("TimeUTC(%v) == %v, want date part %v", d, tUTC, d)
// 		}
// 		if tUTC.Location() != time.UTC {
// 			t.Errorf("TimeUTC(%v) == %v, want %v", d, tUTC.Location(), time.UTC)
// 		}
// 		tLocal := d.Local()
// 		if !same(d, tLocal) {
// 			t.Errorf("TimeLocal(%v) == %v, want date part %v", d, tLocal, d)
// 		}
// 		if tLocal.Location() != time.Local {
// 			t.Errorf("TimeLocal(%v) == %v, want %v", d, tLocal.Location(), time.Local)
// 		}
// 		for _, z := range zones {
// 			location := time.FixedZone("zone", z*60*60)
// 			tInLoc := d.In(location)
// 			if !same(d, tInLoc) {
// 				t.Errorf("TimeIn(%v) == %v, want date part %v", d, tInLoc, d)
// 			}
// 			if tInLoc.Location() != location {
// 				t.Errorf("TimeIn(%v) == %v, want %v", d, tInLoc.Location(), location)
// 			}
// 		}
// 	}
// }

// func TestPredicates(t *testing.T) {
// 	// The list of case dates must be sorted in ascending order
// 	cases := []struct {
// 		d Date
// 	}{
// 		{New(-1234, time.February, 5)},
// 		{New(0, time.April, 12)},
// 		{New(1, time.January, 1)},
// 		{New(1946, time.February, 4)},
// 		{New(1970, time.January, 1)},
// 		{New(1976, time.April, 1)},
// 		{New(1999, time.December, 1)},
// 		{New(1111111, time.June, 21)},
// 	}
// 	for i, ci := range cases {
// 		di := ci.d
// 		for j, cj := range cases {
// 			dj := cj.d
// 			testPredicate(t, di, dj, di.Equal(dj), i == j, "Equal")
// 			testPredicate(t, di, dj, di.Before(dj), i < j, "Before")
// 			testPredicate(t, di, dj, di.After(dj), i > j, "After")
// 			testPredicate(t, di, dj, di == dj, i == j, "==")
// 			testPredicate(t, di, dj, di != dj, i != j, "!=")
// 		}
// 	}

// 	// Test IsZero
// 	zero := Date{}
// 	if !zero.IsZero() {
// 		t.Errorf("IsZero(%v) == false, want true", zero)
// 	}
// 	today := Today()
// 	if today.IsZero() {
// 		t.Errorf("IsZero(%v) == true, want false", today)
// 	}
// }

// func testPredicate(t *testing.T, di, dj Date, p, q bool, m string) {
// 	if p != q {
// 		t.Errorf("%s(%v, %v) == %v, want %v\n%v", m, di, dj, p, q, debug.Stack())
// 	}
// }

// func TestArithmetic(t *testing.T) {
// 	cases := []struct {
// 		d Date
// 	}{
// 		{New(-1234, time.February, 5)},
// 		{New(0, time.April, 12)},
// 		{New(1, time.January, 1)},
// 		{New(1946, time.February, 4)},
// 		{New(1970, time.January, 1)},
// 		{New(1976, time.April, 1)},
// 		{New(1999, time.December, 1)},
// 		{New(1111111, time.June, 21)},
// 	}
// 	offsets := []PeriodOfDays{-1000000, -9999, -555, -99, -22, -1, 0, 1, 22, 99, 555, 9999, 1000000}
// 	for _, c := range cases {
// 		di := c.d
// 		for _, days := range offsets {
// 			dj := di.Add(days)
// 			days2 := dj.Sub(di)
// 			if days2 != days {
// 				t.Errorf("AddSub(%v,%v) == %v, want %v", di, days, days2, days)
// 			}
// 			d3 := dj.Add(-days)
// 			if d3 != di {
// 				t.Errorf("AddNeg(%v,%v) == %v, want %v", di, days, d3, di)
// 			}
// 			eMin1 := min(di.day, dj.day)
// 			aMin1 := di.Min(dj)
// 			if aMin1.day != eMin1 {
// 				t.Errorf("%v.Max(%v) is %s", di, dj, aMin1)
// 			}
// 			eMax1 := max(di.day, dj.day)
// 			aMax1 := di.Max(dj)
// 			if aMax1.day != eMax1 {
// 				t.Errorf("%v.Max(%v) is %s", di, dj, aMax1)
// 			}
// 		}
// 	}
// }

// func TestAddDate(t *testing.T) {
// 	cases := []struct {
// 		d                   Date
// 		years, months, days int
// 		expected            Date
// 	}{
// 		{New(1970, time.January, 1), 1, 2, 3, New(1971, time.March, 4)},
// 		{New(1999, time.September, 28), 6, 4, 2, New(2006, time.January, 30)},
// 		{New(1999, time.September, 28), 0, 0, 3, New(1999, time.October, 1)},
// 		{New(1999, time.September, 28), 0, 1, 3, New(1999, time.October, 31)},
// 	}
// 	for _, c := range cases {
// 		di := c.d
// 		dj := di.AddDate(c.years, c.months, c.days)
// 		if dj != c.expected {
// 			t.Errorf("%v AddDate(%v,%v,%v) == %v, want %v", di, c.years, c.months, c.days, dj, c.expected)
// 		}
// 		dk := dj.AddDate(-c.years, -c.months, -c.days)
// 		if dk != di {
// 			t.Errorf("%v AddDate(%v,%v,%v) == %v, want %v", dj, -c.years, -c.months, -c.days, dk, di)
// 		}
// 	}
// }

// func TestAddPeriod(t *testing.T) {
// 	cases := []struct {
// 		in       Date
// 		delta    period.Period
// 		expected Date
// 	}{
// 		{New(1970, time.January, 1), period.NewYMD(0, 0, 0), New(1970, time.January, 1)},
// 		{New(1971, time.January, 1), period.NewYMD(10, 0, 0), New(1981, time.January, 1)},
// 		{New(1972, time.January, 1), period.NewYMD(0, 10, 0), New(1972, time.November, 1)},
// 		{New(1972, time.January, 1), period.NewYMD(0, 24, 0), New(1974, time.January, 1)},
// 		{New(1973, time.January, 1), period.NewYMD(0, 0, 10), New(1973, time.January, 11)},
// 		{New(1973, time.January, 1), period.NewYMD(0, 0, 365), New(1974, time.January, 1)},
// 		{New(1974, time.January, 1), period.NewHMS(1, 2, 3), New(1974, time.January, 1)},
// 		// note: the period is not normalised so the HMS is ignored even though it's more than one day
// 		{New(1975, time.January, 1), period.NewHMS(24, 2, 3), New(1975, time.January, 1)},
// 	}
// 	for i, c := range cases {
// 		out := c.in.AddPeriod(c.delta)
// 		if out != c.expected {
// 			t.Errorf("%d: %v.AddPeriod(%v) == %v, want %v", i, c.in, c.delta, out, c.expected)
// 		}
// 	}
// }

// func min(a, b PeriodOfDays) PeriodOfDays {
// 	if a < b {
// 		return a
// 	}
// 	return b
// }

// func max(a, b PeriodOfDays) PeriodOfDays {
// 	if a > b {
// 		return a
// 	}
// 	return b
// }

// See main testin in period_test.go
// func TestIsLeap(t *testing.T) {
// 	cases := []struct {
// 		year     int
// 		expected bool
// 	}{
// 		{2000, true},
// 		{2001, false},
// 	}
// 	for _, c := range cases {
// 		got := IsLeap(c.year)
// 		if got != c.expected {
// 			t.Errorf("TestIsLeap(%d) == %v, want %v", c.year, got, c.expected)
// 		}
// 	}
// }

// func TestDaysIn(t *testing.T) {
// 	cases := []struct {
// 		year     int
// 		month    time.Month
// 		expected int
// 	}{
// 		{2000, time.January, 31},
// 		{2000, time.February, 29},
// 		{2001, time.February, 28},
// 		{2001, time.April, 30},
// 	}
// 	for _, c := range cases {
// 		got1 := DaysIn(c.year, c.month)
// 		if got1 != c.expected {
// 			t.Errorf("DaysIn(%d, %d) == %v, want %v", c.year, c.month, got1, c.expected)
// 		}
// 		d := New(c.year, c.month, 1)
// 		got2 := d.LastDayOfMonth()
// 		if got2 != c.expected {
// 			t.Errorf("DaysIn(%d) == %v, want %v", c.year, got2, c.expected)
// 		}
// 	}
// }

// func TestDatesInRange(t *testing.T) {
// 	is := is.New(t)

// 	d1, err := ParseISO("2019-12-31")
// 	is.NoErr(err)
// 	t.Log("d1", d1)

// 	d2, err := ParseISO("2020-01-09")
// 	is.NoErr(err)
// 	t.Log("d2", d2)

// 	dates, err := DatesInRange(d1, d2)
// 	is.NoErr(err)

// 	for _, v := range dates {
// 		t.Logf("Date: %v", v)
// 	}
// }

// func TestTiming(t *testing.T) {
// 	is := is.New(t)

// 	start := time.Now()
// 	format := "2019-12-31"
// 	count := 1000
// 	for i := 0; i < count; i++ {
// 		// Get a unix timestamp we should not parse
// 		_, err := ParseISO(format)
// 		is.NoErr(err)
// 	}

// 	t.Logf("Took %v to check %s  %d times", time.Since(start), format, count)
// }

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
