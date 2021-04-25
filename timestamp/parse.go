package timestamp

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/imarsman/datetime/xfmt"
)

var reDigits *regexp.Regexp
var timeFormats = []string{} // A slice of time formats to be used if ISO parsing fails
var locationAtomic atomic.Value

var namedZoneTimeFormats = []string{
	"Monday, 02-Jan-06 15:04:05 MST",
	"Mon, 02 Jan 2006 15:04:05 MST",
}

// timeFormats a list of Golang time formats to cycle through. The first match
// will cause the loop through the formats to exit.
var nonISOTimeFormats = []string{

	// "Monday, 02-Jan-06 15:04:05 MST",
	// "Mon, 02 Jan 2006 15:04:05 MST",

	// RFC7232 - used in HTTP protocol
	"Mon, 02 Jan 2006 15:04:05 GMT",

	// RFC850
	// Unreliable to have Zone name known - don't try
	// "Monday, 02-Jan-06 15:04:05 MST",

	// RFC1123
	// Unreliable to have Zone name known - don't try
	// "Mon, 02 Jan 2006 15:04:05 MST",

	// RFC1123Z
	"Mon, 02 Jan 2006 15:04:05 -0700",

	"Mon, 02 Jan 2006 15:04:05",
	"Monday, 02-Jan-2006 15:04:05",

	// RFC822Z
	"02 Jan 06 15:04 -0700",

	// Just in case
	"2006-01-02 15-04-05",
	"20060102150405",

	// Stamp
	// Year not known - don't try
	// "Jan _2 15:04:05",

	// StampMilli
	// Year not known - don't try
	// "Jan _2 15:04:05.000",

	// StampMicro
	// Year not known - don't try
	// "Jan _2 15:04:05.000000",

	// StampNano
	// Year not known - don't try
	// "Jan _2 15:04:05.000000000",

	// Hopefully less likely to be found. Assume UTC.
	"20060102",
	"01/02/2006",
	"1/2/2006",
}

func init() {
	reDigits = regexp.MustCompile(`^\d+\.?\d+$`)
	timeFormats = append(timeFormats, nonISOTimeFormats...)
	// A cache for zones tied to offsets to save quite a bit of time and 3
	// allocations needed to get a fixed zone.
	// cachedZones := make(map[int]*time.Location)
	locationAtomic.Store(make(map[int]*time.Location))
}

// LocationFromOffset get a location based on the offset seconds from UTC. Uses a cache
// of locations based on offset.
func LocationFromOffset(offsetSec int) *time.Location {
	cachedZones := locationAtomic.Load().(map[int]*time.Location)
	var location *time.Location
	if l, ok := cachedZones[offsetSec]; ok {
		location = l
		// Given that zones are in at most 15 minute increments and can be
		// positive or negative there should only be so many.
		// https://time.is/time_zones
		// There are currently 37 observed UTC offsets in the world
		// (38 when Iran is on standard time).
		// Allow up to 50.
		// zoneMu.Lock()
		if len(cachedZones) > 50 {
			locationAtomic.Store(make(map[int]*time.Location))
		}
	} else {
		location = time.FixedZone("FixedZone", offsetSec)
		cachedZones[offsetSec] = location
		locationAtomic.Store(cachedZones)
	}

	return location
}

// RunesToString convert runes list to string with no allocation
//
// WriteRune is more complex than WriteByte so can't inline
//
// A small cost a few ns in testing is incurred for using a string builder.
// There are no heap allocations using strings.Builder.
func RunesToString(runes ...rune) string {
	var sb = new(strings.Builder)
	for i := 0; i < len(runes); i++ {
		sb.WriteRune(runes[i])
	}
	return sb.String()
}

// BytesToString convert byte list to string with no allocation
//
// can inline - strings.Builder WriteByte is less complex than WriteRune
//
// A small cost a few ns in testing is incurred for using a string builder.
// There are no heap allocations using strings.Builder.
func BytesToString(bytes ...byte) string {
	var sb = new(strings.Builder)
	for i := 0; i < len(bytes); i++ {
		sb.WriteByte(bytes[i])
	}
	return sb.String()
}

// IntPow calculates n to the mth power. Since the result is an int, it is
// assumed that m is a positive power. Don't use for negative values of m.
// https://stackoverflow.com/questions/64108933/how-to-use-math-pow-with-integers-in-golang
func intPow(n, m int) int {
	var result int = n
	switch m {
	case 0:
		return 1
	}
	for i := 2; i <= m; i++ {
		result *= n
	}
	return result
}

// StringToInt convert a string with no decimal spaces to int.
func StringToInt(input string) (int, error) {
	var runes []rune = []rune(input)
	var l int = len(runes)
	if l > 9 {
		return 0, errors.New("length greater than 9")
	}

	var result int = 0
	for i := 0; i < l; i++ {
		r := runes[i]
		if unicode.IsDigit(r) {
			var num int = int(r) - 48 // 48 is 0 so any digit will be an offset from that
			var pow int = intPow(10, l-i-1)
			result = result + num*pow
		} else {
			return 0, errors.New("non digit in input")
		}
	}
	return result, nil
}

// ParseInUTC parse for all timestamps, defaulting to UTC, and return UTC zoned
// time
func ParseInUTC(timeStr string) (time.Time, error) {
	return parseTimestamp(timeStr, time.UTC, false)
}

// ParseISOInUTC parse limited to ISO timestamp formats and return UTC zoned time
func ParseISOInUTC(timeStr string) (time.Time, error) {
	return parseTimestamp(timeStr, time.UTC, true)
}

// ParseInLocation parse for all timestamp formats and default to location if
// there is no zone in the incoming timestamp. Return time adjusted to UTC.
func ParseInLocation(timeStr string, location *time.Location) (time.Time, error) {
	return parseTimestamp(timeStr, location, false)
}

// ParseISOInLocation parse limited to ISO timestamp formats, defaulting to
// location if there is no zone in the incoming timezone. Return time  adjusted
// to UTC.
func ParseISOInLocation(timeStr string, location *time.Location) (time.Time, error) {
	return parseTimestamp(timeStr, location, true)
}

// ParseTimestampInLocation parse timestamp, defaulting to location if there is
// no zone in the incoming timestamp, and return time ajusted to the incoming
// location.
//
// Can't inline due to use of range but it's too complex anyway.
func parseTimestamp(timeStr string, location *time.Location, isoOnly bool) (time.Time, error) {
	timeStr = strings.TrimSpace(timeStr)
	var original string = timeStr

	// Check to see if the incoming data is a series of digits or digits with a
	// single decimal place.

	var isTS bool = false
	if reDigits.MatchString(timeStr) {
		// A 20060101 date will have 10 digits
		// A 20060102060708 timestamp will have 14 digits
		// A Unix timetamp will have 10 digits
		// A Unix nanosecond timestamp will have 19 digits
		l := len(timeStr)
		if l != 8 && l != 14 {
			isTS = true
		}
	}

	// Try ISO parsing first. The lexer is tolerant of some inconsistency in
	// format that is not ISO-8601 compliant, such as dashes where there should
	// be colons and a space instead of a T to separate date and time.
	if isTS == false {
		t, err := ParseISOTimestamp(timeStr, location)
		if err == nil {
			return t, nil
		}
	}

	// If only iso format patterns should be tried leave now
	if isoOnly == true {

		xfmtBuf := new(xfmt.Buffer)
		// Avoid heap allocation
		xfmtBuf.S("Could not parse as ISO timestamp ").S(timeStr)

		return time.Time{}, errors.New(BytesToString(xfmtBuf.Bytes()...))
	}

	if isTS == true {
		t, err := ParseUnixTS(timeStr)
		if err != nil {
			xfmtBuf := new(xfmt.Buffer)
			// Avoid heap allocation
			xfmtBuf.S("Could not parse as UNIX timestamp ").S(timeStr)

			return time.Time{}, errors.New(BytesToString(xfmtBuf.Bytes()...))
		}

		return t.In(location), nil
	}

	// If not a unix type timestamp try alternate non-iso timestamp formats
	s := nonISOTimeFormats
	for _, format := range s {
		// If no zone in timestamp use location
		t, err := time.ParseInLocation(format, original, location)
		if err == nil {
			return t, nil
		}
	}

	xfmtBuf := new(xfmt.Buffer)
	xfmtBuf.S("Could not parse with other timestamp patterns ").S(timeStr)

	return time.Time{}, errors.New(BytesToString(xfmtBuf.Bytes()...))
}

// parseUnixTS returns seconds and nanoseconds from a timestamp that has the
// format "%d.%09d", time.Unix(), int64(time.Nanosecond()))
// if the incoming nanosecond portion is longer or shorter than 9 digits it is
// converted to nanoseconds. The expectation is that the seconds and
// seconds will be used to create a time variable.
//
// For example:
/*
	var t int64
    seconds, nanoseconds, err := ParseTimestamp("1136073600.000000001",0)
    if err == nil {
		t = time.Unix(seconds, nanoseconds)
	}
*/
// From https://github.com/moby/moby/blob/master/api/types/time/timestamp.go
// Part of Docker, under Apache licence.
//
// Can't inline
func parseUnixTS(timeStr string) (int64, int64, error) {
	sa := strings.SplitN(timeStr, ".", 2)

	// Parse the first portion
	s, err := strconv.ParseInt(sa[0], 10, 64)
	if err != nil {
		return s, 0, err
	}

	// If there is no second part assume nanoseconds is zero
	if len(sa) != 2 {
		return s, 0, nil
	}

	// Parse the first portion
	n, err := strconv.ParseInt(sa[1], 10, 64)
	if err != nil {
		return s, n, err
	}

	// big package cumbersome for this purpose and slower
	// var nb = big.NewInt(n)
	// var i = big.NewInt(10)
	// var e = big.NewInt(int64(9 - len(sa[1])))
	// var r = i.Exp(i, e, nil)

	// bi := nb.Mul(nb, r)
	// s = bi.Int64()

	// should already be in nanoseconds but just in case convert n to
	// nanoseconds. math.Pow works well here.
	n = int64(float64(n) * math.Pow(float64(10), float64(9-len(sa[1]))))

	return s, n, nil
}

// ParseUnixTS parse a timestamp directly, assuming input is some sort of UNIX
// timestamp. If the input is known to be a timestamp this will be faster than
// first trying to parse as other forms of timestamp. This function will handle
// timestamps in the form of seconds and nanoseconds delimited by a period.
//   e.g. 113621424536300000 becomes 1136214245.36300000
//
// Can't inline
func ParseUnixTS(timeStr string) (time.Time, error) {
	match := reDigits.MatchString(timeStr)

	timeStrLength := len(timeStr)
	// Only proceed if the incoming timestamp is a number with up to one
	// decimaal place. Otherwise return an error.
	if match == true {

		// Don't support timestamps less than 7 characters in length
		// to avoid strange date formats from being parsed.
		// Max would be 9999999, or Sun Apr 26 1970 17:46:39 GMT+0000
		if timeStrLength > 6 {
			toSend := timeStr
			// Break it into a format that has a period between second and
			// millisecond portions for the function.
			if timeStrLength > 10 {
				sec, nsec := timeStr[0:10], timeStr[11:timeStrLength-1]

				// Avoid heap allocation
				xfmtBuf := new(xfmt.Buffer)
				xfmtBuf.S(sec).S(".").S(nsec)

				toSend = BytesToString(xfmtBuf.Bytes()...)
			}
			// Get seconds, nanoseconds, and error if there was a problem
			s, n, err := parseUnixTS(toSend)
			if err != nil {
				return time.Time{}, err
			}
			// If it was a unix seconds timestamp n will be zero. If it was a
			// nanoseconds timestamp there will be a nanoseconds portion that is not
			// zero.
			t := time.Unix(s, n)

			return t, nil
		}
	}
	xfmtBuf := new(xfmt.Buffer)
	// Avoid heap allocation
	xfmtBuf.S("Could not parse as UNIX timestamp ").S(timeStr)
	b := xfmtBuf.Bytes()

	// return time.Time{}, fmt.Errorf("Could not parse as UNIX timestamp %s", timeStr)
	return time.Time{}, errors.New(BytesToString(b...))
}

// ParseISOTimestamp parse an ISO timetamp iteratively. The reult will be in the
// zone for the timestamp or if there is no zone offset in the incoming
// timestamp the incoming location will bue used. It is the responsibility of
// further steps to standardize to a specific zone offset.
func ParseISOTimestamp(timeStr string, location *time.Location) (time.Time, error) {
	// Define sections that can change.

	const maxLength int = 35
	timeStrLength := len(timeStr)

	if timeStrLength > maxLength {
		// Avoid allocations that would occur with fmt.Sprintf
		xfmtBuf := new(xfmt.Buffer)
		xfmtBuf.S("Input ").S(timeStr[0:35]).S("... length is ").D(timeStrLength).S(" and > max of ").D(maxLength)

		// errors.New escapes to heap
		return time.Time{}, errors.New(BytesToString(xfmtBuf.Bytes()...))
	}

	// Needs to not be a const since it gets reassigned
	var currentSection int = 0 // value for current section

	// Define sections that are constant. Use iota since the incrementing values
	// correspond to the incremental section processing and give each const a
	// separate value.

	const (
		emptySection     int = iota // value for empty section
		yearSection                 // year - four digits
		monthSection                // month - 2 digits
		daySection                  // day - 2 digits
		hourSection                 // hour - 2 digits
		minuteSection               // minute - 2 digits
		secondSection               // second - 2 digits
		subsecondSection            // subsecond 1-9 digits
		zoneSection                 // zone +/-HHMM or Z
		afterSection                // after - when done
	)

	// Define whether offset is positive for later offset calculation.

	var offsetPositive bool = false // is offset from UTC positive

	// Define the varous part to hold values for year, month, etc. Make initial
	// size 0 and capacity enough to avoid shuffling when appending.

	var (
		yearParts      = make([]rune, 0, 4) // year digit parts
		monthParts     = make([]rune, 0, 2) // month digit parts
		dayParts       = make([]rune, 0, 2) // day digit parts
		hourParts      = make([]rune, 0, 2) // hour digit parts
		minuteParts    = make([]rune, 0, 2) // minute digit parts
		secondParts    = make([]rune, 0, 2) // second digit parts
		subsecondParts = make([]rune, 0, 9) // subsecond digit parts
		zoneParts      = make([]rune, 0, 4) // zone parts
	)

	const (
		yearMax      int = 4 // max length for year
		monthMax     int = 2 // max length for month number
		dayMax       int = 2 // max length for day number
		hourMax      int = 2 // max length for hour number
		minuteMax    int = 2 // max length for minute number
		secondMax    int = 2 // max length for second number
		subsecondMax int = 9 // max length for subsecond number
		zoneMax      int = 4 // max length for zone
	)

	// A function to handle adding to a slice if it is not above capacity and
	// flagging when it has reached capacity. Runs same speed when inline and is
	// only used here. Return a flag indicating if a timestamp part has reached
	// its max capacity plus the modified slice to avoid issues due to
	// appending.
	var addIf = func(part []rune, add rune, max int) ([]rune, bool) {
		if len(part) < max {
			part = append(part, add)
		}
		if len(part) == max {
			return part, true
		}
		return part, false
	}

	var isZero = func(part ...rune) bool {
		for i := 0; i < len(part); i++ {
			if part[i] != '0' {
				return false
			}
		}
		return true
	}

	var unparsed []string      // string representation of unparsed runes and their positions
	var partAtMax bool = false // flag indicating current part is filled

	// Loop through runes in time string and decide what to do with each.
	for i, r := range timeStr {
		orig := r
		if unicode.IsDigit(r) {
			switch currentSection {
			// Initially no section is active
			case emptySection:
				currentSection = yearSection
				yearParts, partAtMax = addIf(yearParts, r, yearMax)
				if partAtMax == true {
					currentSection = monthSection
				}
				// Year section is used until full
			case yearSection:
				yearParts, partAtMax = addIf(yearParts, r, yearMax)
				if partAtMax == true {
					currentSection = monthSection
				}
				// Month section is used until full
			case monthSection:
				monthParts, partAtMax = addIf(monthParts, r, monthMax)
				if partAtMax == true {
					currentSection = daySection
				}
				// Day section is used until full
			case daySection:
				dayParts, partAtMax = addIf(dayParts, r, dayMax)
				if partAtMax == true {
					currentSection = hourSection
				}
				// Hour section is used until full
			case hourSection:
				hourParts, partAtMax = addIf(hourParts, r, hourMax)
				if partAtMax == true {
					currentSection = minuteSection
				}
				// Minute section is used until full
			case minuteSection:
				minuteParts, partAtMax = addIf(minuteParts, r, minuteMax)
				if partAtMax == true {
					currentSection = secondSection
				}
				// Second section is used until full
			case secondSection:
				secondParts, partAtMax = addIf(secondParts, r, secondMax)
				if partAtMax == true {
					currentSection = subsecondSection
				}
				// Subsecond section is used until full
			case subsecondSection:
				subsecondParts, partAtMax = addIf(subsecondParts, r, subsecondMax)
				if partAtMax == true {
					currentSection = zoneSection
				}
				// Zone section is used until full
			case zoneSection:
				// Add to zone
				zoneParts, partAtMax = addIf(zoneParts, r, zoneMax)
				if partAtMax == true {
					// We could exit here but we can continue to more accurately
					// report bad date parts if we allow things to continue.
					currentSection = afterSection
				}
			default:
				// Default to bad input

				// Avoid allocations that would occur with fmt.Sprintf
				xfmtBuf := new(xfmt.Buffer)
				xfmtBuf.S("'").C(orig).S("'").C('@').D(i)

				unparsed = append(unparsed, BytesToString(xfmtBuf.Bytes()...))
			}
			// If the current section is not for subseconds skip
		} else if r == '.' {
			// There could be extraneous decimal characters.
			if currentSection != subsecondSection {
				continue
			}
			// currentSection = subsecondSection
		} else if r == '-' || r == '+' {
			// Selectively define offset possitivity
			if currentSection == subsecondSection {
				offsetPositive = (r == '+')
				currentSection = zoneSection
			}
			// Valid but not useful for parsing
		} else if unicode.ToUpper(r) == 'T' || r == ':' || r == '/' {
			continue
			// Zulu offset
		} else if unicode.ToUpper(r) == 'Z' {
			// define offset as zero for hours and minutes
			if currentSection == zoneSection || currentSection == subsecondSection {
				zoneParts = append(zoneParts, '0', '0', '0', '0')
				break
			} else {
				// Assume bad input

				// Avoid allocations that would occur with fmt.Sprintf
				xfmtBuf := new(xfmt.Buffer)
				xfmtBuf.S("'").C(orig).S("'").C('@').D(i)

				unparsed = append(unparsed, BytesToString(xfmtBuf.Bytes()...))
			}
			// Ignore spaces
		} else if unicode.IsSpace(r) {
			continue
		} else {
			// Catch-all for characters not allowed

			// Avoid allocations that would occur with fmt.Sprintf
			xfmtBuf := new(xfmt.Buffer)
			xfmtBuf.S("'").C(orig).S("'").C('@').D(i)

			unparsed = append(unparsed, BytesToString(xfmtBuf.Bytes()...))
		}
	}

	// If we've found characters not allocated, error.
	if len(unparsed) > 0 {
		// Avoid allocations that would occur with fmt.Sprintf
		xfmtBuf := new(xfmt.Buffer)
		xfmtBuf.S("got unparsed caracters ").S(strings.Join(unparsed, ",")).S(" in input ").S(timeStr)

		// errors.New escapes to heap
		return time.Time{}, errors.New(BytesToString(xfmtBuf.Bytes()...))
	}

	zoneFound := false        // has time zone been found
	zoneLen := len(zoneParts) // length of the zone found

	// If length < 4
	if zoneLen < zoneMax {
		zoneFound = true
		// A zone with 1 or 3 characters is ambiguous
		if zoneLen == 1 || zoneLen == 3 {
			// Avoid allocations that would occur with fmt.Sprintf
			xfmtBuf := new(xfmt.Buffer)
			xfmtBuf.S("Zone is of length ").D(zoneLen).S(" wich is not enough to detect zone")

			// errors.New escapes to heap
			return time.Time{}, errors.New(BytesToString(xfmtBuf.Bytes()...))

			// With no zone assume UTC and set all offset characters to 0
		} else if zoneLen == 0 {
			zoneFound = false
			zoneParts = append(zoneParts, '0', '0', '0', '0')
		} else if zoneLen == 2 {
			// Zone of length 2 needs padding to set minute offset
			zoneParts = append(zoneParts, '0', '0')
		}
	} else {
		// Zone is found. Used later when setting location
		zoneFound = true
	}

	yearLen := len(yearParts)
	monthLen := len(monthParts)
	dayLen := len(dayParts)

	hourLen := len(hourParts)
	minuteLen := len(minuteParts)
	secondLen := len(secondParts)

	// This does not need to be recalculated
	subsecondLen := len(subsecondParts)
	// This will need to be recalculated
	zoneLen = len(zoneParts)

	// Allow for just dates and convert to timestamp with zero valued time
	// parts. Since we are fixing it here it will pass the next tests if nothing
	// else is wrong or missing.
	if hourLen == 0 && minuteLen == 0 && secondLen == 0 {
		hourParts = append(hourParts, '0', '0')
		minuteParts = append(minuteParts, '0', '0')
		secondParts = append(secondParts, '0', '0')

		hourLen, minuteLen, secondLen = hourMax, minuteMax, secondMax
	}

	// Error if any part does not contain enough characters. This could happen
	// easily if for instance a year had 2 digits instead of 4. If this happened
	// year would take 4 digits, month would take 2, day would take 2, hour
	// would take 2, minute would take 2, and second would get none. We are thus
	// requiring that all date and time parts be fully allocated even if we
	// can't tell where the problem started.
	if yearLen != yearMax {
		// errors.New escapes to heap
		return time.Time{}, errors.New("Input year length is not 4")
	}
	if monthLen != monthMax {
		// errors.New escapes to heap
		return time.Time{}, errors.New("Input month length is not 2")
	}
	if dayLen != dayMax {
		// errors.New escapes to heap
		return time.Time{}, errors.New("Input day length is not 2")
	}
	if hourLen != hourMax {
		// errors.New escapes to heap
		return time.Time{}, errors.New("Input hour length is not 2")
	}
	if minuteLen != minuteMax {
		// errors.New escapes to heap
		return time.Time{}, errors.New("Input minute length is not 2")
	}
	if secondLen != secondMax {
		// errors.New escapes to heap
		return time.Time{}, errors.New("Input second length is not 2")
	}

	// We already only put digits into the parts so Atoi should be fine in all
	// cases. The problem would have been with an incorrect number of digits in
	// a part, which would have been caught above.

	var err error
	var y, m, d, h, mn, s int
	y, m, d, h, mn, s = 0, 0, 0, 0, 0, 0

	// Get year int value from yearParts rune slice
	// Should not error since only digits were place in slice
	// If zero can avoid an allocation and time
	if isZero(yearParts...) == false {
		y, err = strconv.Atoi(RunesToString(yearParts...))
		// y, err = StringToInt(RunesToString(yearParts...))
		if err != nil {
			return time.Time{}, err
		}
	}

	// Get month int value from monthParts rune slice
	// Should not error since only digits were place in slice
	// If zero can avoid an allocation and time
	if isZero(monthParts...) == false {
		m, err = strconv.Atoi(RunesToString(monthParts...))
		// m, err = StringToInt(RunesToString(monthParts...))
		if err != nil {
			return time.Time{}, err
		}
	}

	// Get day int value from dayParts rune slice
	// Should not error since only digits were place in slice
	// If zero can avoid an allocation and time
	if isZero(dayParts...) == false {
		d, err = strconv.Atoi(RunesToString(dayParts...))
		// d, err = StringToInt(RunesToString(dayParts...))
		if err != nil {
			return time.Time{}, err
		}
	}

	// Get hour int value from hourParts rune slice
	// Should not error since only digits were place in slice
	// If zero can avoid an allocation and time
	if isZero(hourParts...) == false {
		// h, err = StringToInt(RunesToString(hourParts...))
		h, err = strconv.Atoi(RunesToString(hourParts...))
		if err != nil {
			return time.Time{}, err
		}
	}

	// Get minute int value from minParts rune slice
	// Should not error since only digits were place in slice
	// If zero can avoid an allocation and time
	if isZero(minuteParts...) == false {
		// mn, err = StringToInt(RunesToString(minuteParts...))
		mn, err = strconv.Atoi(RunesToString(minuteParts...))
		if err != nil {
			return time.Time{}, err
		}
	}

	// Get second int value from secondParts rune slice
	// Should not error since only digits were place in slice
	// If zero can avoid an allocation and time
	if isZero(secondParts...) == false {
		// s, err = StringToInt(RunesToString(secondParts...))
		s, err = strconv.Atoi(RunesToString(secondParts...))
		if err != nil {
			return time.Time{}, err
		}
	}

	var subseconds int = 0 // default subsecond value is 0

	// Handle subseconds if that slice is nonempty
	// There would have been an error if the length of subsecond parts was
	// greater than subsecondMax
	if subsecondLen > 0 {
		// If zero can avoid an allocation and time
		if isZero(subsecondParts...) == false {
			// subseconds, err = StringToInt(RunesToString(subsecondParts...))
			subseconds, err = strconv.Atoi(RunesToString(subsecondParts...))
			if err != nil {
				return time.Time{}, err
			}
			// Calculate subseconds in terms of nanosecond if the length is less
			// than the full length for nanoseconds since that is what the time.Date
			// function is expecting.
			if subsecondLen < subsecondMax {
				// 10^ whatever extra decimal place count is missing from 9
				// This has been tried 3 ways
				// - with a custom intPow function
				// - with math.Pow
				// - with the big package
				//
				// - using math.Pow seems to be quite consistent
				// - using intPow seems to be most consistently fast
				// - going with math.Pow because it's almost identical and is very
				//   tested

				// var i = big.NewInt(int64(subseconds))
				// var e = big.NewInt(int64(subsecondMax - subsecondLen))
				// bi := i.Exp(i, e, nil)
				// subseconds = int(bi.Int64())

				// subseconds = intPow(subseconds, subsecondMax-subsecondLen)

				subseconds = int(
					subseconds *
						int(math.Pow(10, (float64(subsecondMax-subsecondLen)))))
			}
		}
	}

	// NOTE:
	// We have already ensured that all parts have the correct number of digits.
	// don't worry about ensuring that the values of months, days, hours,
	// minutes, etc. are being too large within their digit span. The Go time
	// package increments higher values as appropriate. For instance a value of
	// 60 seconds would force an addition to the minute and all the way up to
	// the year for 2020-12-31T59:59:60-0000

	offsetZero := isZero(zoneParts...)

	// Create timestamp based on parts with proper offsset

	// If no zone was found in scan use default location
	if zoneFound == false {
		return time.Date(y, time.Month(m), d, h, mn, s, subseconds, location), nil
	}

	if offsetZero == true {
		return time.Date(y, time.Month(m), d, h, mn, s, subseconds, time.UTC), nil
	}

	var offsetH int = 0 // starting state for offset hours
	var offsetM int = 0 // starting state for offset minutes

	hourOffsetParts := zoneParts[0:2]
	// Can avoid allocations by skipping this
	if isZero(hourOffsetParts...) == false {
		// Evaluate hour offset from the timestamp value
		// Should not error since only digits were place in slice
		// offsetH, err = StringToInt(RunesToString(hourOffsetParts...))
		offsetH, err = strconv.Atoi(RunesToString(hourOffsetParts...))
		if err != nil {
			return time.Time{}, err
		}
	}

	minuteOffsetParts := zoneParts[2:]
	// Can avoid allocations by skipping this
	if isZero(minuteOffsetParts...) == false {
		// Evaluate minute offset from the timestamp value
		// Should not error since only digits were place in slice
		// offsetM, err = StringToInt(RunesToString(minuteOffsetParts...))
		offsetM, err = strconv.Atoi(RunesToString(minuteOffsetParts...))
		if err != nil {
			return time.Time{}, err
		}
	}

	// Set offset based on hours and minutes
	offsetSec := offsetH*60*60 + offsetM*60

	// The +/- in the timestamp was used to set offsetPositive
	// Negate it if offset is not positive
	if offsetPositive == false {
		offsetSec = -offsetSec
	}

	// Don't allow offset minutes not in 15 minute increment
	switch offsetM {
	case 0:
	case 15:
	case 30:
	case 45:
	default:
		// Avoid allocations that would occur with fmt.Sprintf
		xfmtBuf := new(xfmt.Buffer)
		xfmtBuf.S("UTC offset minutes ").D(offsetM).S(" not in a 15 minute increment")

		// errors.New escapes to heap
		return time.Time{}, errors.New(BytesToString(xfmtBuf.Bytes()...))
	}

	return time.Date(y, time.Month(m), d, h, mn, s, subseconds, LocationFromOffset(offsetSec)), nil
}
