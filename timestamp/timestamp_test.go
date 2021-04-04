package timestamp

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// checkDate for use in parse checking
func checkDate(t *testing.T, input string, compare string) {
	v, err := ParseUTC(input)
	assert.Nil(t, err)

	ts := ISO8601LongMsec(v)
	t.Logf("Input %s, Expecting %s, Got %s", input, compare, ts)
	assert.Equal(t, compare, ts)
}

// func TestTimespanForDateRange(t *testing.T) {
// 	d1 := "2020-12-01"
// 	d2 := "2020-12-02"

// 	// t1, t2, err := TimespanForDateRange(d1, d2)
// 	// assert.Nil(t, err)

// 	t1String := ISO8601LongMsec(t1)
// 	t2String := ISO8601LongMsec(t2)

// 	t1Check, err := ParseUTC("2020-12-01T00:00:00Z")
// 	assert.Nil(t, err)

// 	t1CheckString := ISO8601LongMsec(t1Check)
// 	assert.Equal(t, t1String, t1CheckString)

// 	t2Check, err := ParseUTC("2020-12-02T00:00:00Z")
// 	assert.Nil(t, err)

// 	t2CheckString := ISO8601LongMsec(t2Check)
// 	assert.Equal(t, t2String, t2CheckString)

// 	t.Log("t1String", t1String)
// 	t.Log("t1CheckString", t1CheckString)
// 	t.Log("t2String", t2String)
// 	t.Log("t2CheckString", t2CheckString)
// }

// TestDateRange test the range date function
// func TestRangeDate(t *testing.T) {

// 	d1 := "2019-12-30"
// 	d2 := "2020-01-08"

// 	t.Log("Dates from", d1, "to", d2)

// 	start, err := TimeForDate(d1)
// 	assert.Nil(t, err)

// 	end, err := TimeForDate(d2)
// 	assert.Nil(t, err)

// 	dateMap := make(map[string]string)

// 	for rd := RangeDate(start, end); ; {
// 		date := rd()
// 		if date.IsZero() {
// 			break
// 		}
// 		//now let's create & sort an array with our map keys

// 		dateMap[DateForTime(date)] = ""
// 	}

// 	dates := make([]string, 0, len(dateMap))

// 	for k := range dateMap {
// 		dates = append(dates, k)
// 	}
// 	sort.Strings(dates)

// 	for _, k := range dates {
// 		t.Logf("Date: %s", k)
// 	}
// 	assert.Equal(t, 10, len(dates))
// 	assert.Equal(t, "2019-12-30", dates[0])
// 	assert.Equal(t, "2020-01-08", dates[len(dates)-1])
// }

// TestDateRangeFromDates test getting a date range and comparing the date range
// to the similar TimespanForDateRange value. They should be equal in terms of
// timestamp values.
// func TestDateRangeFromDates(t *testing.T) {
// 	d1 := "2020-01-01"
// 	d2 := "2021-01-10"

// 	r, err := DateRangeFromDates(d1, d2)
// 	assert.Nil(t, err)

// 	t.Log("Date range", r)

// 	assert.Equal(t, "2020-01-01/2021-01-10", r)

// 	t1, err := TimeForDate(d1)
// 	assert.Nil(t, err)
// 	t.Log("t1", t1)

// 	t2, err := TimeForDate(d2)
// 	assert.Nil(t, err)

// 	// Also check to ensure timespan method agrees
// 	ts1, ts2, err := TimespanForDateRange(d1, d2)
// 	assert.Nil(t, err)

// 	t.Logf("t1 %v t2 %v", t1, t2)
// 	t.Logf("ts1 %v ts2 %v", ts1, ts2)

// 	assert.Equal(t, t1, ts1)
// 	assert.Equal(t, t2, ts2)
// }

// func TestTimeDateOnly(t *testing.T) {
// 	time, err := TimeForDate("2020-01-01")
// 	assert.Nil(t, err)

// 	time2 := TimeDateOnly(time)

// 	t.Logf("time for date %v time date only %v", time, time2)
// 	assert.Equal(t, time, time2)
// }

// TestParse parse all patterns anc compare with expected values
func TestParse(t *testing.T) {

	// It is possible to have a strring which is just digits that will be parsed
	// as a timestamp, incorrectly.
	t.Log(ParseUTC("2006010247"))

	// Get a unix timestamp we should not parse
	_, err := ParseUTC("1")
	assert.NotNil(t, err)

	// Get time value from parsed reference time
	unixBase, err := ParseUTC("2006-01-02T15:04:05.000+00:00")
	assert.Nil(t, err)

	// Use parsed reference time to create unix timestamp and nanosecond timestamp
	checkDate(t, fmt.Sprint(unixBase.UnixNano()), "2006-01-02T15:04:05.000+00:00")
	checkDate(t, fmt.Sprint(unixBase.Unix()), "2006-01-02T15:04:05.000+00:00")

	// RFC7232 - used in HTTP protocol
	checkDate(t, "Mon, 02 Jan 2006 15:04:05 GMT", "2006-01-02T15:04:05.000+00:00")

	// Short ISO-8601 timestamps with numerical zone offsets
	checkDate(t, "20060102T150405-0700", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "20060102T150405-0700", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "20060102T150405-07", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "20060102T150405.000+0000", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102T150405.000-0000", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102T150405.000-0700", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "20060102T150405.000+0700", "2006-01-02T08:04:05.000+00:00")
	checkDate(t, "20060102T150405.000000-0700", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "20060102T150405.999999999-0700", "2006-01-02T22:04:05.999+00:00")

	// Long ISO-8601 timestamps with numerical zone offsets
	checkDate(t, "2006-01-02T15:04:05-07:00", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15:04:05-07", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15:04:05.000-07:00", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15:04:05.000-07", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15:04:05.000000-07:00", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15:04:05.001000-07", "2006-01-02T22:04:05.001+00:00")
	checkDate(t, "2006-01-02T15:04:05.001000000-07:00", "2006-01-02T22:04:05.001+00:00")
	checkDate(t, "2006-01-02T15:04:05.999999999-07", "2006-01-02T22:04:05.999+00:00")

	// Short  ISO-8601 timestamps with zulu zone offsets
	checkDate(t, "20060102T150405Z", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102T150405.000Z", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102T150405.000000Z", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102T150405.000000000Z", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102T150405.001000000Z", "2006-01-02T15:04:05.001+00:00")
	checkDate(t, "20060102T150405.000100000Z", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102T150405.999999999Z", "2006-01-02T15:04:05.999+00:00")

	// Long date time with zulu zone offsets
	checkDate(t, "2006-01-02T15:04:05Z", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "2006-01-02T15:04:05.000Z", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "2006-01-02T15:04:05.000000Z", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "2006-01-02T15:04:05.999999999Z", "2006-01-02T15:04:05.999+00:00")

	// Just in case
	checkDate(t, "2006-01-02 15-04-05", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102150405", "2006-01-02T15:04:05.000+00:00")

	// Short ISO-8601 timestamps with no zone offset. Assume UTC.
	checkDate(t, "20060102T150405", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102T150405.000", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102T150405.000000", "2006-01-02T15:04:05.000+00:00")
	checkDate(t, "20060102T150405.999999999", "2006-01-02T15:04:05.999+00:00")

	// SQL
	checkDate(t, "2006-01-02 22:04:05", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02 22:04:05 -00", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02 22:04:05 +00", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02 22:04:05 -00:00", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02 22:04:05 +00:00", "2006-01-02T22:04:05.000+00:00")

	// Hopefully less likely to be found. Assume UTC.
	checkDate(t, "20060102", "2006-01-02T00:00:00.000+00:00")
	checkDate(t, "2006-01-02", "2006-01-02T00:00:00.000+00:00")
	checkDate(t, "2006/01/02", "2006-01-02T00:00:00.000+00:00")
	checkDate(t, "01/02/2006", "2006-01-02T00:00:00.000+00:00")
	checkDate(t, "1/2/2006", "2006-01-02T00:00:00.000+00:00")

	// Weird ones with improper separators
	checkDate(t, "2006-01-02T15-04-05-0700", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15-04-05.000-0700", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15-04-05.000000-0700", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15-04-05.999999999-0700", "2006-01-02T22:04:05.999+00:00")

	checkDate(t, "2006-01-02T15-04-05-07:00", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15-04-05.000-07:00", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15-04-05.000000-07:00", "2006-01-02T22:04:05.000+00:00")
	checkDate(t, "2006-01-02T15-04-05.999999999-07:00", "2006-01-02T22:04:05.999+00:00")
}

// TestIsPeriod check to see if various patterns are valid ISO-8601 periods
// func TestIsPeriod(t *testing.T) {
// 	isPeriod := IsPeriod("")
// 	assert.False(t, isPeriod)

// 	isPeriod = IsPeriod("P3Y4M5DT6H4M3S")
// 	assert.True(t, isPeriod)

// 	isPeriod = IsPeriod("P3y4m5dT6h4m3s")
// 	assert.True(t, isPeriod)

// 	isPeriod = IsPeriod("P3DDD4m5dT6h4m3s")
// 	assert.False(t, isPeriod)

// 	p, err := ParsePeriod("P3Y4M5DT6H4M3S")
// 	assert.Nil(t, err)

// 	t.Logf("Period %s", p)
// }

// TestParsePeriod parse a period
// func TestParsePeriod(t *testing.T) {
// 	pStr := "PT10M"
// 	p, err := ParsePeriod(pStr)
// 	assert.Nil(t, err)
// 	assert.Equal(t, p, pStr)
// }

// TestPeriodPositive test whether a priod
// func TestPeriodPositve(t *testing.T) {
// 	p, err := ParsePeriod("PT10M")
// 	assert.Nil(t, err)
// 	t.Logf("Is period %v", IsPeriod(p))

// 	pp, err := PeriodPositive(p)
// 	assert.Nil(t, err)

// 	pn, err := PeriodNegative(p)
// 	assert.Nil(t, err)

// 	assert.Equal(t, pp, p)
// 	assert.Equal(t, pn, "-PT10M")
// }

// TestOrdering check ordering call
func TestOrdering(t *testing.T) {
	t1, err1 := ParseUTC("20201210T223900-0500")
	assert.Nil(t, err1)

	t2, err2 := ParseUTC("20201211T223900-0500")
	assert.Nil(t, err2)

	assert.True(t, StartTimeIsBeforeEndTime(t1, t2))
	assert.False(t, StartTimeIsBeforeEndTime(t2, t1))
}

// func TestDurationToPeriod(t *testing.T) {
// 	t1String := "20201210T235959-0500"
// 	t1, err := ParseUTC(t1String)
// 	assert.Nil(t, err)

// 	t2String := "20211211T000000-0500"
// 	t2, err := ParseUTC(t2String)
// 	assert.Nil(t, err)

// 	d := t2.Sub(t1)

// 	pString := DurationToPeriod(d)
// 	p2, _ := period.NewOf(d)
// 	// Check that period obtained is the same as the one from period library
// 	assert.Equal(t, pString, p2.String())

// 	t.Logf("Period for duration between %s and %s %s", t1String, t2String, pString)
// 	t.Logf("Period obtained %s and from period library %s", pString, p2.String())
// 	assert.Equal(t, "P365DT1S", pString)
// }
