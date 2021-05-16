// Copyright 2015 Rick Beton. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package period

import (
	"time"
)

// Add adds two periods together. Use this method along with Negate in order to subtract periods.
//
// The result is not normalised and may overflow arithmetically (to make this unlikely, use Normalise on
// the inputs before adding them).
func (p Period) Add(that Period) Period {
	return NewPeriod(
		p.Years()+that.Years(),
		p.Months()+that.Months(),
		p.Days()+that.Days(),
		p.Hours()+that.Hours(),
		p.Minutes()+that.Minutes(),
		p.Seconds()+that.Seconds(),
	)
}

//-------------------------------------------------------------------------------------------------

// AddTo adds the period to a time, returning the result.
// A flag is also returned that is true when the conversion was precise and false otherwise.
//
// When the period specifies hours, minutes and seconds only, the result is precise.
// Also, when the period specifies whole years, months and days (i.e. without fractions), the
// result is precise. However, when years, months or days contains fractions, the result
// is only an approximation (it assumes that all days are 24 hours and every year is 365.2425
// days, as per Gregorian calendar rules).
// TODO: test this
func (p Period) AddTo(t time.Time) (time.Time, bool, error) {
	// wholeYears := (p.years % 1) == 0
	// wholeMonths := (p.months % 1) == 0
	// wholeDays := (p.days % 1) == 0

	// if wholeYears && wholeMonths && wholeDays {
	if p.years > 0 || p.months > 0 || p.days > 0 {
		// in this case, time.AddDate provides an exact solution
		stE3, err := hmsDuration(p)
		if err != nil {
			return time.Time{}, false, err
		}

		// t1 := t.AddDate(int(p.years/10), int(p.months/10), int(p.days/10))
		t1 := t.AddDate(int(p.years), int(p.months), int(p.days))
		return t1.Add(stE3 * time.Millisecond), true, nil
	}

	d, precise, err := p.Duration()
	if err != nil {
		return time.Time{}, false, err
	}

	return t.Add(d), precise, nil
}

//-------------------------------------------------------------------------------------------------

// Scale a period by a multiplication factor. Obviously, this can both enlarge and shrink it,
// and change the sign if negative. The result is normalised, but integer overflows are silently
// ignored.
//
// Bear in mind that the internal representation is limited by fixed-point arithmetic with two
// decimal places; each field is only int16.
//
// Known issue: scaling by a large reduction factor (i.e. much less than one) doesn't work properly.
func (p Period) Scale(factor float64) *Period {
	result, _ := p.ScaleWithOverflowCheck(factor)
	return result
}

// ScaleWithOverflowCheck a period by a multiplication factor. Obviously, this can both enlarge and shrink it,
// and change the sign if negative. The result is normalised. An error is returned if integer overflow
// happened.
//
// Bear in mind that the internal representation is limited by fixed-point arithmetic with one
// decimal place; each field is only int16.
//
// Known issue: scaling by a large reduction factor (i.e. much less than one) doesn't work properly.
func (p Period) ScaleWithOverflowCheck(factor float64) (*Period, error) {
	if -0.5 < factor && factor < 0.5 {
		d, pr1, err := p.Duration()
		if err != nil {
			return &Period{}, err
		}

		mul := float64(d) * float64(factor)
		p2, pr2 := NewOf(time.Duration(mul))
		return p2.Normalise(pr1 && pr2), nil
	}

	y := int64(float64(p.years) * factor)
	m := int64(float64(p.months) * factor)
	d := int64(float64(p.days) * factor)
	hh := int64(float64(p.hours) * factor)
	mm := int64(float64(p.minutes) * factor)
	ss := int64(float64(p.seconds) * factor)
	subsec := int64(float64(p.subseconds) * factor)

	newPeriod := NewPeriod(y, m, d, hh, mm, ss)
	newPeriod.subseconds = int(subsec)
	newPeriod.negative = p.IsNegative()

	return newPeriod.Normalise(true), nil
}
