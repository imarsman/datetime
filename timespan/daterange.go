// Copyright 2015 Rick Beton. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package timespan

import (
	"fmt"
	"time"

	"github.com/imarsman/datetime/isodate"
	"github.com/imarsman/datetime/period"
)

const minusOneNano time.Duration = -1

// DateRange carries a isodate and a number of days and describes a range between two isodates.
type DateRange struct {
	mark isodate.Date
	days isodate.PeriodOfDays
}

// NewDateRangeOf assembles a new isodate range from a start time and a duration, discarding
// the precise time-of-day information. The start time includes a location, which is not
// necessarily UTC. The duration can be negative.
func NewDateRangeOf(start time.Time, duration time.Duration) DateRange {
	sd := isodate.NewAt(start)
	ed := isodate.NewAt(start.Add(duration))
	return DateRange{sd, isodate.PeriodOfDays(ed.Sub(sd))}
}

// NewDateRange assembles a new isodate range from two isodates. These are half-open, so
// if start and end are the same, the range spans zero (not one) day. Similarly, if they
// are on subsequent days, the range is one isodate (not two).
// The result is normalised.
func NewDateRange(start, end isodate.Date) DateRange {
	if end.Before(start) {
		return DateRange{end, isodate.PeriodOfDays(start.Sub(end))}
	}
	return DateRange{start, isodate.PeriodOfDays(end.Sub(start))}
}

// NewYearOf constructs the range encompassing the whole year specified.
func NewYearOf(year int) DateRange {
	start := isodate.New(year, time.January, 1)
	end := isodate.New(year+1, time.January, 1)
	return DateRange{start, isodate.PeriodOfDays(end.Sub(start))}
}

// NewMonthOf constructs the range encompassing the whole month specified for a given year.
// It handles leap years correctly.
func NewMonthOf(year int, month time.Month) DateRange {
	start := isodate.New(year, month, 1)
	endT := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)
	end := isodate.NewAt(endT)
	return DateRange{start, isodate.PeriodOfDays(end.Sub(start))}
}

// EmptyRange constructs an empty range. This is often a useful basis for
// further operations but note that the end isodate is undefined.
func EmptyRange(day isodate.Date) DateRange {
	return DateRange{day, 0}
}

// OneDayRange constructs a range of exactly one day. This is often a useful basis for
// further operations. Note that the last isodate is the same as the start isodate.
func OneDayRange(day isodate.Date) DateRange {
	return DateRange{day, 1}
}

// DayRange constructs a range of n days.
//
// Note that n can be negative. In this case, the specified day will be the end day,
// which is outside of the half-open range; the last day will be the day before the
// day specified.
func DayRange(day isodate.Date, n isodate.PeriodOfDays) DateRange {
	if n < 0 {
		return DateRange{day.Add(n), -n}
	}
	return DateRange{day, n}
}

// Days returns the period represented by this range. This will never be negative.
func (dateRange DateRange) Days() isodate.PeriodOfDays {
	if dateRange.days < 0 {
		return -dateRange.days
	}
	return dateRange.days
}

// IsZero returns true if this has a zero start isodate and the the range is empty.
// Usually this is because the range was created via the zero value.
func (dateRange DateRange) IsZero() bool {
	return dateRange.days == 0 && dateRange.mark.IsZero()
}

// IsEmpty returns true if this has a starting isodate but the range is empty (zero days).
func (dateRange DateRange) IsEmpty() bool {
	return dateRange.days == 0
}

// Start returns the earliest isodate represented by this range.
func (dateRange DateRange) Start() isodate.Date {
	if dateRange.days < 0 {
		return dateRange.mark.Add(isodate.PeriodOfDays(1 + dateRange.days))
	}
	return dateRange.mark
}

// Last returns the last isodate (inclusive) represented by this range. Be careful because
// if the range is empty (i.e. has zero days), then the last is undefined so an empty isodate
// is returned. Therefore it is often more useful to use End() instead of Last().
// See also IsEmpty().
func (dateRange DateRange) Last() isodate.Date {
	if dateRange.days < 0 {
		return dateRange.mark // because mark is at the end
	} else if dateRange.days == 0 {
		return isodate.Date{}
	}
	return dateRange.mark.Add(dateRange.days - 1)
}

// End returns the isodate following the last isodate of the range. End can be considered to
// be the exclusive end, i.e. the final value of a half-open range.
//
// If the range is empty (i.e. has zero days), then the start isodate is returned, this being
// also the (half-open) end value in that case. This is more useful than the undefined result
// returned by Last() for empty ranges.
func (dateRange DateRange) End() isodate.Date {
	if dateRange.days < 0 {
		return dateRange.mark.Add(1) // because mark is at the end
	}
	return dateRange.mark.Add(dateRange.days)
}

// Normalise ensures that the number of days is zero or positive.
// The normalised isodate range is returned;
// in this value, the mark isodate is the same as the start isodate.
func (dateRange DateRange) Normalise() DateRange {
	if dateRange.days < 0 {
		return DateRange{dateRange.mark.Add(dateRange.days), -dateRange.days}
	}
	return dateRange
}

// ShiftBy moves the isodate range by moving both the start and end isodates similarly.
// A negative parameter is allowed.
func (dateRange DateRange) ShiftBy(days isodate.PeriodOfDays) DateRange {
	if days == 0 {
		return dateRange
	}
	newMark := dateRange.mark.Add(days)
	return DateRange{newMark, dateRange.days}
}

// ExtendBy extends (or reduces) the isodate range by moving the end isodate.
// A negative parameter is allowed and this may cause the range to become inverted
// (i.e. the mark isodate becomes the end isodate instead of the start isodate).
func (dateRange DateRange) ExtendBy(days isodate.PeriodOfDays) DateRange {
	if days == 0 {
		return dateRange
	}
	return DateRange{dateRange.mark, dateRange.days + days}.Normalise()
}

// ShiftByPeriod moves the isodate range by moving both the start and end isodates similarly.
// A negative parameter is allowed.
//
// Any time component is ignored. Therefore, be careful with periods containing
// more that 24 hours in the hours/minutes/seconds fields. These will not be
// normalised for you; if you want this behaviour, call delta.Normalise(false)
// on the input parameter.
//
// For example, PT24H adds nothing, whereas P1D adds one day as expected. To
// convert a period such as PT24H to its equivalent P1D, use
// delta.Normalise(false) as the input.
func (dateRange DateRange) ShiftByPeriod(delta period.Period) DateRange {
	if delta.IsZero() {
		return dateRange
	}
	newMark := dateRange.mark.AddPeriod(delta)
	//fmt.Printf("mark + %v : %v -> %v", delta, DateRange.mark, newMark)
	return DateRange{newMark, dateRange.days}
}

// ExtendByPeriod extends (or reduces) the isodate range by moving the end isodate.
// A negative parameter is allowed and this may cause the range to become inverted
// (i.e. the mark isodate becomes the end isodate instead of the start isodate).
func (dateRange DateRange) ExtendByPeriod(delta period.Period) DateRange {
	if delta.IsZero() {
		return dateRange
	}
	newEnd := dateRange.End().AddPeriod(delta)
	//fmt.Printf("%v, end + %v : %v -> %v", DateRange.mark, delta, DateRange.End(), newEnd)
	return NewDateRange(dateRange.Start(), newEnd)
}

// String describes the isodate range in human-readable form.
func (dateRange DateRange) String() string {
	norm := dateRange.Normalise()
	switch norm.days {
	case 0:
		return fmt.Sprintf("0 days at %v", norm.mark)
	case 1:
		return fmt.Sprintf("1 day on %v", norm.mark)
	default:
		return fmt.Sprintf("%v days from %v to %v", norm.days, norm.Start(), norm.Last())
	}
}

// Contains tests whether the isodate range contains a specified isodate.
// Empty isodate ranges (i.e. zero days) never contain anything.
func (dateRange DateRange) Contains(d isodate.Date) bool {
	if dateRange.days == 0 {
		return false
	}
	return !(d.Before(dateRange.Start()) || d.After(dateRange.Last()))
}

// StartUTC assumes that the start isodate is a UTC isodate and gets the start time of that isodate, as UTC.
// It returns midnight on the first day of the range.
func (dateRange DateRange) StartUTC() time.Time {
	return dateRange.Start().UTC()
}

// EndUTC assumes that the end isodate is a UTC isodate and returns the time a nanosecond after the end time
// in a specified location. Along with StartUTC, this gives a 'half-open' range where the start
// is inclusive and the end is exclusive.
func (dateRange DateRange) EndUTC() time.Time {
	return dateRange.End().UTC()
}

// ContainsTime tests whether a given local time is within the isodate range. The time range is
// from midnight on the start day to one nanosecond before midnight on the day after the end isodate.
// Empty isodate ranges (i.e. zero days) never contain anything.
//
// If a calculation needs to be 'half-open' (i.e. the end isodate is exclusive), simply use the
// expression 'dateRange.ExtendBy(-1).ContainsTime(t)'
func (dateRange DateRange) ContainsTime(t time.Time) bool {
	if dateRange.days == 0 {
		return false
	}
	utc := t.In(time.UTC)
	return !(utc.Before(dateRange.StartUTC()) || dateRange.EndUTC().Add(minusOneNano).Before(utc))
}

// Merge combines two isodate ranges by calculating a isodate range that just encompasses them both.
// There are two special cases.
//
// Firstly, if one range is entirely contained within the other range, the larger of the two is
// returned. Otherwise, the result is from the start of the earlier one to the end of the later
// one, even if the two ranges don't overlap.
//
// Secondly, if either range is the zero value (see IsZero), it is excluded from the merge and
// the other range is returned unchanged.
func (dateRange DateRange) Merge(otherRange DateRange) DateRange {
	if otherRange.IsZero() {
		return dateRange
	}
	if dateRange.IsZero() {
		return otherRange
	}
	minStart := dateRange.Start().Min(otherRange.Start())
	maxEnd := dateRange.End().Max(otherRange.End())
	return NewDateRange(minStart, maxEnd)
}

// Duration computes the duration (in nanoseconds) from midnight at the start of the isodate
// range up to and including the very last nanosecond before midnight on the end day.
// The calculation is for UTC, which does not have daylight saving and every day has 24 hours.
//
// If the range is greater than approximately 290 years, the result will hard-limit to the
// minimum or maximum possible duration (see time.Sub(t)).
func (dateRange DateRange) Duration() time.Duration {
	return dateRange.End().UTC().Sub(dateRange.Start().UTC())
}

// DurationIn computes the duration (in nanoseconds) from midnight at the start of the isodate
// range up to and including the very last nanosecond before midnight on the end day.
// The calculation is for the specified location, which may have daylight saving, so not every day
// necessarily has 24 hours. If the isodate range spans the day the clocks are changed, this is
// taken into account.
//
// If the range is greater than approximately 290 years, the result will hard-limit to the
// minimum or maximum possible duration (see time.Sub(t)).
func (dateRange DateRange) DurationIn(loc *time.Location) time.Duration {
	return dateRange.EndTimeIn(loc).Sub(dateRange.StartTimeIn(loc))
}

// StartTimeIn returns the start time in a specified location.
func (dateRange DateRange) StartTimeIn(loc *time.Location) time.Time {
	return dateRange.Start().In(loc)
}

// EndTimeIn returns the nanosecond after the end time in a specified location. Along with
// StartTimeIn, this gives a 'half-open' range where the start is inclusive and the end is
// exclusive.
func (dateRange DateRange) EndTimeIn(loc *time.Location) time.Time {
	return dateRange.End().In(loc)
}

// TimeSpanIn obtains the time span corresponding to the isodate range in a specified location.
// The result is normalised.
func (dateRange DateRange) TimeSpanIn(loc *time.Location) TimeSpan {
	s := dateRange.StartTimeIn(loc)
	d := dateRange.DurationIn(loc)
	return TimeSpan{s, d}
}
