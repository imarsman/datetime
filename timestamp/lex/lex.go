package lex

import (
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
	// A range of subsecond digit lengths
	lexer.Add([]byte(`\.\d`), getToken(tokmap["SUBSECOND"]))
	lexer.Add([]byte(`\.\d\d`), getToken(tokmap["SUBSECOND"]))
	lexer.Add([]byte(`\.\d\d\d`), getToken(tokmap["SUBSECOND"]))
	lexer.Add([]byte(`\.\d\d\d\d`), getToken(tokmap["SUBSECOND"]))
	lexer.Add([]byte(`\.\d\d\d\d\d`), getToken(tokmap["SUBSECOND"]))
	lexer.Add([]byte(`\.\d\d\d\d\d\d`), getToken(tokmap["SUBSECOND"]))
	lexer.Add([]byte(`\.\d\d\d\d\d\d\d`), getToken(tokmap["SUBSECOND"]))
	lexer.Add([]byte(`\.\d\d\d\d\d\d\d\d`), getToken(tokmap["SUBSECOND"]))
	lexer.Add([]byte(`\.\d\d\d\d\d\d\d\d\d`), getToken(tokmap["SUBSECOND"]))
	// Assumes after first and second millennium
	lexer.Add([]byte(`[12]\d\d\d\d\d\d\d`), getToken(tokmap["DATE"]))
	// lexer.Add([]byte(`[\+\-]\d\d\d\d\d\d\d\d\d`), getToken(tokmap["DATE"]))
	lexer.Add([]byte(`\d\d\d\d\d\d`), getToken(tokmap["TIME"]))
	// Four digit zone
	lexer.Add([]byte(`[\-\+]\d\d\d\d`), getToken(tokmap["ZONE"]))
	// Allow for 2 digit zone
	lexer.Add([]byte(`[\-\+]\d\d`), getToken(tokmap["ZONE"]))
	// Zulu (UTC) indicator
	lexer.Add([]byte(`Z`), getToken(tokmap["ZONE"]))
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

// Parse only get time and error
func Parse(bytes []byte) (time.Time, error) {
	t, _, err := ParseGetParts(bytes)
	return t, err
}

// ParseGetParts parse and get the timestamp parts if more analysis is desired
func ParseGetParts(bytes []byte) (time.Time, TimestampParts, error) {
	return scan(bytes)
}

// scan read input and get time, the timestamp parts, and error
// https://blog.gopheracademy.com/advent-2017/lexmachine-advent/
func scan(bytes []byte) (time.Time, TimestampParts, error) {
	tsp := NewTimestampParts()

	timeStr := string(bytes)

	tsp.ORIGINAL = timeStr

	timeStr = strings.ToUpper(timeStr)

	// Works for dashes in dates and for just dates
	//   e.g. 2006-01-02
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
	// If there are two dashes assume they are for a bad timestamp with dashes
	if strings.Count(timeStr, "-") == 2 {
		timeStr = strings.ReplaceAll(timeStr, "-", "")
	}

	// Colons are not useful for parsing
	timeStr = strings.ReplaceAll(timeStr, ":", "")

	// Colons are not useful for parsing
	// timeStr = strings.ReplaceAll(timeStr, " ", "")

	if lexer == nil {
		return time.Time{}, TimestampParts{}, errors.New("Lexer is nil. Something went wrong")
	}

	scanner, err := lexer.Scanner([]byte(timeStr))
	if err != nil {
		// fmt.Println(err)
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
			tsp.YEAR = v[0:4]
			tsp.MONTH = v[4:6]
			tsp.DAY = v[6:8]
		case tokmap["TIME"]:
			v := token.Value.(string)
			tsp.HOUR = v[0:2]
			tsp.MINUTE = v[2:4]
			tsp.SECOND = v[4:6]
		case tokmap["SUBSECOND"]:
			v := token.Value.(string)
			// if len(v[1:]) == 2 {
			// 	v = v + "0"
			// 	// fmt.Println(v)
			// }
			tsp.SUBSECOND = v
			// fmt.Println(tsp.SUBSECOND)
		case tokmap["ZONE"]:
			// Note that RFC3339 requires a zone either as Z or an offset
			// The pattern 2006-01-02T15:04:05Z0700 is not meant to be a parser
			// pattern but rather a requirement specification for some kind of
			// zone to be present. This means that RFC3339 is stringent in that
			// it requires a zone indicator. Having no zone means that the date
			// is incorrectly specified.
			// https://stackoverflow.com/a/63321401/2694971
			v := token.Value.(string)
			if len(v) == 3 {
				v = v + "00"
			}
			tsp.ZONE = v
			if v == "-0000" {
				tsp.ZONE = "+0000"
			}
		}
	}

	// Allow for just a date with no time and no zone
	// This will assume UTC
	if tsp.ZONE == "" {
		if tsp.HOUR == "" {
			tsp.HOUR = "00"
		}
		if tsp.MINUTE == "" {
			tsp.MINUTE = "00"
		}
		if tsp.SECOND == "" {
			tsp.SECOND = "00"
		}
	}

	canProceed := tsp.YEAR != "" && tsp.HOUR != ""

	if canProceed == true {
		str := tsp.YEAR + tsp.MONTH + tsp.DAY + "T" + tsp.HOUR + tsp.MINUTE + tsp.SECOND
		if tsp.SUBSECOND != "" {
			str = str + tsp.SUBSECOND
		}
		format := baseTimestampFormat
		if tsp.SUBSECOND != "" {
			format = format + "." + strings.Repeat("0", len(tsp.SUBSECOND)-1)
		}

		if tsp.ZONE != "" {
			str = str + tsp.ZONE

			switch tsp.ZONE {
			case zuluIndicatorFormat:
				format = format + zuluIndicatorFormat
			default:
				format = format + timezoneOffsetFormat
			}
		} else {
			str = str + zuluIndicatorFormat

			format = format + zuluIndicatorFormat
		}

		tsp.CALCULATED = str
		t, err := time.Parse(format, str)

		if err != nil {
			return time.Time{}, TimestampParts{}, errors.New("Could not parse timestamp")
		}

		t = t.In(time.UTC)

		return t, tsp, nil
	}

	// If we got here we have a problem
	return time.Time{}, TimestampParts{}, errors.New("Could not parse timestamp")
}
