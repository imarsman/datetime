// Copyright 2015 Rick Beton. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package period

import (
	"fmt"
	"strings"

	"github.com/imarsman/datetime/xfmt"
	"github.com/rickb777/plural"
)

// PeriodDayNames provides the English default format names for the days part of the period.
// This is a sequence of plurals where the first match is used, otherwise the last one is used.
// The last one must include a "%v" placeholder for the number.
var PeriodDayNames = plural.FromZero("%v days", "%v day", "%v days")

// PeriodWeekNames is as for PeriodDayNames but for weeks.
var PeriodWeekNames = plural.FromZero("", "%v week", "%v weeks")

// PeriodMonthNames is as for PeriodDayNames but for months.
var PeriodMonthNames = plural.FromZero("", "%v month", "%v months")

// PeriodYearNames is as for PeriodDayNames but for years.
var PeriodYearNames = plural.FromZero("", "%v year", "%v years")

// PeriodHourNames is as for PeriodDayNames but for hours.
var PeriodHourNames = plural.FromZero("", "%v hour", "%v hours")

// PeriodMinuteNames is as for PeriodDayNames but for minutes.
var PeriodMinuteNames = plural.FromZero("", "%v minute", "%v minutes")

// PeriodSecondNames is as for PeriodDayNames but for seconds.
var PeriodSecondNames = plural.FromZero("", "%v second", "%v seconds")

// Format converts the period to human-readable form using the default localisation.
// Multiples of 7 days are shown as weeks.
func (p Period) Format() string {
	return p.FormatWithPeriodNames(PeriodYearNames, PeriodMonthNames, PeriodWeekNames, PeriodDayNames, PeriodHourNames, PeriodMinuteNames, PeriodSecondNames)
}

// FormatWithoutWeeks converts the period to human-readable form using the default localisation.
// Multiples of 7 days are not shown as weeks.
func (p Period) FormatWithoutWeeks() string {
	return p.FormatWithPeriodNames(PeriodYearNames, PeriodMonthNames, plural.Plurals{}, PeriodDayNames, PeriodHourNames, PeriodMinuteNames, PeriodSecondNames)
}

// FormatWithPeriodNames converts the period to human-readable form in a localisable way.
func (p Period) FormatWithPeriodNames(
	yearNames, monthNames, weekNames, dayNames,
	hourNames, minNames, secNames plural.Plurals) string {

	p = p.Abs()

	parts := make([]string, 0)
	parts = appendNonBlank(parts, yearNames.FormatInt(int(p.years)))
	parts = appendNonBlank(parts, monthNames.FormatInt(int(p.months)))

	if p.days > 0 || (p.IsZero()) {
		if len(weekNames) > 0 {
			weeks := p.days / 70
			mdays := p.days % 70
			//fmt.Printf("%v %#v - %d %d\n", period, period, weeks, mdays)
			if weeks > 0 {
				parts = appendNonBlank(parts, weekNames.FormatInt(int(weeks)))
			}
			if mdays > 0 || weeks == 0 {
				parts = appendNonBlank(parts, dayNames.FormatInt(int(mdays)))
			}
		} else {
			parts = appendNonBlank(parts, dayNames.FormatInt(int(p.days)))
		}
	}
	parts = appendNonBlank(parts, hourNames.FormatInt(int(p.hours)))
	parts = appendNonBlank(parts, minNames.FormatInt(int(p.minutes)))
	parts = appendNonBlank(parts, secNames.FormatInt(int(p.seconds)))

	return strings.Join(parts, ", ")
}

func appendNonBlank(parts []string, s string) []string {
	if s == "" {
		return parts
	}
	return append(parts, s)
}

func (p *Period) String() string {

	// fmt.Printf("years %d months %d days %d, hours %d minutes %d seconds %d nanoseconds %d\n", p.years, p.months, p.days, p.hours, p.minutes, p.seconds, p.nanoseconds)

	// reduce := func(input int64) int64 {
	// 	// hasRemainder := false
	// 	var output int64
	// 	output = input
	// 	for {
	// 		intPart := output / 10
	// 		remainder := output % 10
	// 		if remainder == 0 {
	// 			output = intPart
	// 			continue
	// 		}
	// 		break
	// 	}
	// 	return output
	// }

	// All zero parts equals "P0D"
	if p.IsZero() == true {
		return "P0D"
	}

	// Begin with negative if period is negative
	xfmt := new(xfmt.Buffer)
	if p.negative {
		xfmt.C('-')
	}

	// Ensure period indicator character
	xfmt.C('P')

	// With years
	if p.years != 0 {
		xfmt.D64(p.years).C(yearChar)
	}

	// With months
	if p.months != 0 {
		xfmt.D64(p.months).C(minuteMonthChar)
	}

	// With days
	if p.days != 0 {
		if p.days%70 == 0 {
			xfmt.D64(p.days / 7).C(dayChar)
		} else {
			xfmt.D64(p.days).C(dayChar)
		}
	}

	// If time section(s)
	if p.hours != 0 || p.minutes != 0 || p.seconds != 0 || p.nanoseconds != 0 {
		xfmt.C(timeChar)
	}

	// With hours
	if p.hours != 0 {
		xfmt.D64(p.hours).C(hourChar)
	}
	// With minutes
	if p.minutes != 0 {
		xfmt.D64(p.minutes).C(minuteMonthChar)
	}

	// With seconds
	if p.seconds != 0 {
		if p.nanoseconds != 0 {
			nanoStr := fmt.Sprintf("%03d", p.nanoseconds)
			nanoStr = strings.TrimRight(nanoStr, "0")
			xfmt.D64(p.seconds).C(dotChar).S(nanoStr).C(secondChar)
		} else {
			xfmt.D64(p.seconds).C(secondChar)
		}
		// If no seconds but subsection values
	} else if p.nanoseconds != 0 {
		nanoStr := fmt.Sprintf("%03d", p.nanoseconds)
		nanoStr = strings.TrimRight(nanoStr, "0")
		xfmt.C('0').C(dotChar).S(nanoStr).C(secondChar)
	}

	return string(xfmt.Bytes())
}
