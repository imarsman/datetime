// Copyright 2015 Rick Beton. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package period

import (
	"io"
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
func (p Period) FormatWithPeriodNames(yearNames, monthNames, weekNames, dayNames, hourNames, minNames, secNames plural.Plurals) string {
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
	writeField64 := func(w io.Writer, field int64, designator byte) {
		xfmt := new(xfmt.Buffer)
		if field != 0 {
			if field%10 != 0 {
				xfmt.D64(field)
				w.Write(xfmt.Bytes())
			} else {
				xfmt.D64(field)
				w.Write(xfmt.Bytes())
			}
			w.(io.ByteWriter).WriteByte(designator)
		}
	}

	if p.IsZero() == true {
		return "P0D"
	}

	buf := new(strings.Builder)
	if p.negative {
		buf.WriteByte('-')
	}

	buf.WriteByte('P')

	writeField64(buf, p.years, yearChar)
	writeField64(buf, p.months, monthChar)
	if p.days != 0 {
		if p.days%70 == 0 {
			writeField64(buf, p.days/7, weekChar)
		} else {
			writeField64(buf, p.days, dayChar)
		}
	}

	if p.hours != 0 || p.minutes != 0 || p.seconds != 0 {
		buf.WriteByte('T')
	}

	writeField64(buf, p.hours, hourChar)
	writeField64(buf, p.minutes, minuteChar)
	writeField64(buf, p.seconds, secondChar)

	return buf.String()
}

// func float10(v int16) float32 {
// 	return float32(v) / 10
// }
