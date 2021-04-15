package lex

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/timtadh/lexmachine"
	"github.com/timtadh/lexmachine/machines"
)

const (
	baseTimestampFormat  = "20060102T150405"
	zuluIndicatorFormat  = "Z"
	timezoneOffsetFormat = "-0700"
)

var tokens = []string{
	"DATE",
	"TIME",
	"SUBSECOND",
	"ZONE",
	"SHORTZONE",
	"ZULU",
}

var tokenIDs map[string]string // A map from the token names to their string ids

// TimestampParts the parts used while building a timestamp
type TimestampParts struct {
	CALCULATED string
	ORIGINAL   string
	YEAR       string
	MONTH      string
	DAY        string
	HOUR       string
	MINUTE     string
	SECOND     string
	SUBSECOND  string
	ZONE       string
}

// Timestamp a timestamp representation
type Timestamp struct {
	Parts TimestampParts
	TIME  time.Time
}

// NewTimestamp make new Timestamp struct. Can contain the parts for review down
// the line as well.
func NewTimestamp() Timestamp {
	return Timestamp{}
}

// NewTimestampParts get parts struct used while processing
func NewTimestampParts() TimestampParts {
	return TimestampParts{}
}

// A map of token names by int value
var tokmap map[string]int

// Lex Machine lexer
var lexer *lexmachine.Lexer

var reYMDPunctuation *regexp.Regexp

func init() {
	// Only for replacing in date portion
	reYMDPunctuation = regexp.MustCompile(`^(\d{4})[\-\.\/]?(\d{2})[\-\.\/]?(\d{2})(.*)`)

	tokmap = make(map[string]int)
	for id, name := range tokens {
		tokmap[name] = id
	}
	lexer = newLexer()
}

func getToken(tokenType int) lexmachine.Action {
	return func(s *lexmachine.Scanner, m *machines.Match) (interface{}, error) {
		return s.Token(tokenType, string(m.Bytes), m), nil
	}
}

func newLexer() *lexmachine.Lexer {
	lexer := lexmachine.NewLexer()
	// Assumes after first and second millennium
	lexer.Add([]byte(`[12]\d\d\d\d\d\d\d`), getToken(tokmap["DATE"]))
	// lexer.Add([]byte(`[\+\-]\d\d\d\d\d\d\d\d\d`), getToken(tokmap["DATE"]))
	lexer.Add([]byte(`\d\d\d\d\d\d`), getToken(tokmap["TIME"]))
	// A range of subsecond digit lengths are covered
	lexer.Add([]byte(`\.\d+`), getToken(tokmap["SUBSECOND"]))
	// Four digit zone
	lexer.Add([]byte(`[\-\+]\d\d\d\d`), getToken(tokmap["ZONE"]))
	// Allow for 2 digit zone
	lexer.Add([]byte(`[\-\+]\d\d`), getToken(tokmap["SHORTZONE"]))
	// Zulu (UTC) indicator
	lexer.Add([]byte(`[zZ]`), getToken(tokmap["ZULU"]))
	// Skip date/time separator
	lexer.Add([]byte(`[tT]`), skip)
	// Ignore spaces
	lexer.Add([]byte(` `), skip)

	err := lexer.CompileDFA()
	if err != nil {
		lexer = nil
	}
	return lexer
}

func skip(scan *lexmachine.Scanner, match *machines.Match) (interface{}, error) {
	return nil, nil
}

// ParseInLocation parse and default to location if zone not present
func ParseInLocation(bytes []byte, location *time.Location) (time.Time, error) {
	t, _, err := ParseInLocationGetParts(bytes, location)
	return t, err
}

// ParseInUTC only get time and error
func ParseInUTC(bytes []byte) (time.Time, error) {
	t, _, err := ParseInLocationGetParts(bytes, time.UTC)
	return t, err
}

// ParseInLocationGetParts parse and get the timestamp parts if more analysis is desired
func ParseInLocationGetParts(bytes []byte, location *time.Location) (time.Time, TimestampParts, error) {
	return scan(bytes, location)
}

// scan read input and get time, the timestamp parts, and error
// https://blog.gopheracademy.com/advent-2017/lexmachine-advent/
func scan(input []byte, location *time.Location) (time.Time, TimestampParts, error) {
	tsParts := NewTimestampParts()

	timeStr := string(input)

	tsParts.ORIGINAL = timeStr

	// Works for dashes in dates and for just dates
	//   e.g. 2006-01-02
	// 	reYMDPunctuation = regexp.MustCompile(`^(\d{4})[\-\.\/]?(\d{2})[\-\.\/]?(\d{2})(.*)`)
	if strings.Count(timeStr, "-") > 1 || strings.Count(timeStr, "/") > 1 || strings.Count(timeStr, ".") > 1 {
		timeStr = reYMDPunctuation.ReplaceAllString(timeStr, "$1$2$3$4")
	}

	// If there is more than 1 dash left, decide what to do.
	//   e.g. 2021-01-02T00-00-00Z
	//        2021-01-02T00-00-00-04:00
	c := strings.Count(timeStr, "-")
	if c > 1 {
		// Assume bad time with dashes and potentially Z timezone
		if c == 2 {
			timeStr = strings.Replace(timeStr, "-", "", 2)
		}
		// Assume bad time with dashes and a negative UTC offset
		if c == 3 {
			timeStr = strings.Replace(timeStr, "-", "", 2)
		}
	}

	// Colons are not useful for parsing
	timeStr = strings.ReplaceAll(timeStr, ":", "")

	if lexer == nil {
		return time.Time{}, TimestampParts{}, errors.New("Lexer is nil. Something went wrong")
	}

	scanner, err := lexer.Scanner([]byte(timeStr))
	if err != nil {
		return time.Time{}, TimestampParts{}, errors.New("Problem converting " + timeStr)
	}

	for tk, err, eof := scanner.Next(); !eof; tk, err, eof = scanner.Next() {
		if err != nil {
			return time.Time{}, TimestampParts{}, errors.New("Problem converting " + timeStr)
		}
		token := tk.(*lexmachine.Token)

		switch token.Type {
		case tokmap["DATE"]:
			// ISO-8601 allows of a differing number of year digits with a
			// positive or negative offset (+/-).
			//   e.g. +20201-01-01T00:00:00Z
			//
			// This is not currently handled.
			// It could likely be handled by setting the year to a set numbr
			// (e.g. 1000) then computing the offet between that and the years
			// for the timestamp and following the formatting of the time value
			// do an addition or subtration of the previous amount calculated.
			v := token.Value.(string)
			tsParts.YEAR = v[0:4]
			tsParts.MONTH = v[4:6]
			tsParts.DAY = v[6:8]
		case tokmap["TIME"]:
			v := token.Value.(string)
			tsParts.HOUR = v[0:2]
			tsParts.MINUTE = v[2:4]
			tsParts.SECOND = v[4:6]
		case tokmap["SUBSECOND"]:
			v := token.Value.(string)
			tsParts.SUBSECOND = v
			// A zone with hours and minutes offset
		case tokmap["ZONE"]:
			// ZONE NOTE:
			// Note that RFC3339 requires a zone either as Z or an offset
			// The pattern 2006-01-02T15:04:05Z0700 is not meant to be a parser
			// pattern but rather a requirement specification for some kind of
			// zone to be present. This means that RFC3339 is stringent in that
			// it requires a zone indicator. Having no zone means that the date
			// is incorrectly specified.
			// https://stackoverflow.com/a/63321401/2694971
			v := token.Value.(string)
			tsParts.ZONE = v
			if v == "-0000" {
				tsParts.ZONE = "+0000"
			}
			// A zone with just two digits (hours) offset
		case tokmap["SHORTZONE"]:
			var b bytes.Buffer
			b.WriteString(token.Value.(string))
			b.WriteString("00")

			// v := token.Value.(string)
			tsParts.ZONE = b.String() // token.Value.(string) + "00"
			// Zulu (Z) offset (+0000)
		case tokmap["ZULU"]:
			tsParts.ZONE = "Z"
		}
	}

	// Allow for just a date with no time and no zone
	// This will assume UTC
	if tsParts.ZONE == "" {
		if tsParts.HOUR == "" {
			tsParts.HOUR = "00"
		}
		if tsParts.MINUTE == "" {
			tsParts.MINUTE = "00"
		}
		if tsParts.SECOND == "" {
			tsParts.SECOND = "00"
		}
	}

	canProceed := tsParts.YEAR != "" && tsParts.HOUR != ""

	if canProceed == true {
		str := tsParts.YEAR + tsParts.MONTH + tsParts.DAY + "T" + tsParts.HOUR + tsParts.MINUTE + tsParts.SECOND
		if tsParts.SUBSECOND != "" {
			str = str + tsParts.SUBSECOND
		}
		format := baseTimestampFormat
		if tsParts.SUBSECOND != "" {
			format = format + "." + strings.Repeat("9", len(tsParts.SUBSECOND)-1)
		}

		if tsParts.ZONE != "" {
			str = str + tsParts.ZONE

			switch tsParts.ZONE {
			case zuluIndicatorFormat:
				format = format + zuluIndicatorFormat
			default:
				format = format + timezoneOffsetFormat
			}
		}
		// If no zone found the location used will be the default passed in

		tsParts.CALCULATED = str
		var t time.Time

		// time.ParseInLocation will only use the location argument if the
		// timestamp has no zone offset information
		t, err = time.ParseInLocation(format, str, location)
		if err != nil {
			return time.Time{}, TimestampParts{}, errors.New("Could not parse timestamp")
		}

		return t, tsParts, nil
	}

	// If we got here we have a problem
	return time.Time{}, TimestampParts{}, errors.New("Could not parse timestamp")
}
