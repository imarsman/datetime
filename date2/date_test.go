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
	m int64
	d int64
}

func TestMaxMinDates(t *testing.T) {
	d := Max()
	t.Log("max year", d.year, "month", d.month, "day", d.day)
	d = Min()
	t.Log("min year", d.year, "month", d.month, "day", d.day)
}
func TestDayOfYear(t *testing.T) {
	is := is.New(t)
	// test := []Date

	var partList = []dateParts{
		{-1000, 5, 18},
		{1000, 3, 1},
		{2019, 5, 18},
		{2020, 3, 18},
		{2021, 3, 1},
	}

	for _, p := range partList {
		d, err := NewDate(p.y, p.m, p.d)
		dayOfYear, err := d.YearDay()
		is.NoErr(err)
		t.Log("Days into year", dayOfYear, "for year", d.Year(), "month", d.Month(), "day", d.Day())
	}
}

func TestDaysInMonth(t *testing.T) {
	is := is.New(t)

	var partList = []dateParts{
		{2020, 2, 18},
		{2021, 2, 18},
	}
	for _, p := range partList {
		d, err := NewDate(p.y, p.m, p.d)
		days, err := d.daysInMonth()
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
		days := daysSinceEpoch(y)
		t.Logf("Days since epoch for year %d = %d", y, days)
	}
}

func TestDayOfWeek1Jan(t *testing.T) {
	is := is.New(t)

	var partList = []dateParts{
		{-1000, 5, 18},
		{2018, 5, 18},
		{2019, 5, 18},
		{2020, 5, 18},
		{2021, 5, 18},
	}

	var err error
	var d Date

	for _, p := range partList {
		d, err = NewDate(p.y, p.m, p.d)
		dow, err := d.dayOfWeek1Jan()
		is.NoErr(err)
		t.Log("Day of week 1 Jan for", d.year, dow)
	}

	is.NoErr(err)
}

func TestDayOfWeek(t *testing.T) {
	is := is.New(t)

	var partList = []dateParts{
		{2018, 3, 1},
		{2019, 3, 1},
		{2020, 3, 1},
		{2021, 3, 1},
	}

	var err error
	var d Date

	for _, p := range partList {
		d, err = NewDate(p.y, p.m, p.d)
		is.NoErr(err)
		dow, err := d.WeekDay()
		is.NoErr(err)
		t.Log("Day of week", d.year, d.month, d.day, dow)
	}
}

func TestIsLeap(t *testing.T) {
	tests := []int64{
		0,
		1000,
		2000,
		3000,
		1984,
		2000,
		2004,
	}

	for _, y := range tests {
		isLeap := IsLeap(y)
		t.Log("year", y, "isLeap", isLeap)
	}
}

func TestWeekNumer(t *testing.T) {

	iterate := func(t *testing.T, start, end int64) {
		is := is.New(t)
		is.True(start < end)

		for i := start; i < end; i++ {

			weeks := isoWeeksInYear(i)
			if weeks == 53 {
				t.Log("weeks in year", i, weeks)
			}
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
		year, isCE := gregorianYear(tests[i])
		t.Log(tests[i], "year", year, "isCE", isCE)
	}
}

func TestAddDate(t *testing.T) {
	type datePartsWithAdd struct {
		y    int64
		m    int64
		d    int64
		addY int64
		addM int64
		addD int64
	}

	var partList = []datePartsWithAdd{
		{2019, 3, 1, 3, 10, 1},
		{2019, 1, 1, 1, 0, 0},
		{2019, 1, 1, 0, 12, 0},
		{2019, 1, 1, 14, 0, 0},
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
	d, _ = d.SubtractDays(32)
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

func BenchmarkAddParts(b *testing.B) {
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
			d, _, err = d.AddParts(1000, 0, 0)
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
			days = daysSinceEpoch(year)
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
			dow, err = d.dayOfWeek1Jan()
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

	d, err := NewDate(2020, 3, 1)
	is.NoErr(err)

	var dayOfYear int64

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

	var dayOfYear int64

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
func BenchmarkDayOfWeek(b *testing.B) {
	is := is.New(b)
	var err error

	d, err := NewDate(2020, 3, 1)
	is.NoErr(err)
	var dayOfWeek int

	b.ResetTimer()
	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			dayOfWeek, err = d.WeekDay()
		}
	})
	is.NoErr(err)

	is.True(dayOfWeek != 0)
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
