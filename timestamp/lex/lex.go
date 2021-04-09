package lex

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/timtadh/lexmachine"
	"github.com/timtadh/lexmachine/machines"
)

var tokens = []string{
	"DATE",
	"TIME",
	"DASH",
	"HYPHEN",
	"COLON",
	"SUBSECOND",
	"ZONE",
	// "MINUS",
	// "SPACE",
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
	// PLUSMINUS  string
	// ZONEOFFSET string
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

var tokmap map[string]int
var lexer *lexmachine.Lexer
var reYMDDash *regexp.Regexp

func init() {
	reYMDDash = regexp.MustCompile(`^(\d{4})[\-\.\/]?(\d{2})[\-\.\/]?(\d{2})(.*)`)

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
	lexer.Add([]byte(`\.\d\d\d`), getToken(tokmap["SUBSECOND"]))
	lexer.Add([]byte(`\.\d\d\d\d\d\d`), getToken(tokmap["SUBSECOND"]))
	lexer.Add([]byte(`\.\d\d\d\d\d\d\d\d\d`), getToken(tokmap["SUBSECOND"]))
	lexer.Add([]byte(`([12]\d\d\d\d\d\d\d)`), getToken(tokmap["DATE"]))
	lexer.Add([]byte(`(\d\d\d\d\d\d)`), getToken(tokmap["TIME"]))
	lexer.Add([]byte(`[\-\+]\d\d\d\d|\Z`), getToken(tokmap["ZONE"]))
	// Allow for 2 digit zone
	lexer.Add([]byte(`[\-\+]\d\d|\Z`), getToken(tokmap["ZONE"]))
	lexer.Add([]byte(`Z`), getToken(tokmap["ZONE"]))
	lexer.Add([]byte(`:`), skip)
	lexer.Add([]byte(` `), skip)

	err := lexer.Compile()
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
	// This will work for just dates
	if strings.Count(timeStr, "-") > 1 || strings.Count(timeStr, "/") > 1 || strings.Count(timeStr, ".") > 1 {
		timeStr = reYMDDash.ReplaceAllString(timeStr, "$1$2$3$4")
	}

	timeStr = strings.ReplaceAll(timeStr, ":", "")
	timeStr = strings.Replace(timeStr, "T", "", 1)

	if lexer == nil {
		return time.Time{}, TimestampParts{}, errors.New("Lexer is nil. Something went wrong")
	}

	scanner, err := lexer.Scanner([]byte(timeStr))
	if err != nil {
		// fmt.Println(err)
		return time.Time{}, TimestampParts{}, errors.New("Problem converting" + timeStr)
	}

	for tk, err, eof := scanner.Next(); !eof; tk, err, eof = scanner.Next() {
		if err != nil {
			return time.Time{}, TimestampParts{}, errors.New("Problem converting" + timeStr)
		}
		token := tk.(*lexmachine.Token)

		switch token.Type {
		case tokmap["DATE"]:
			v := token.Value.(string)
			tsp.YEAR = v[0:4]
			tsp.MONTH = v[4:6]
			tsp.DAY = v[6:8]
		case tokmap["TIME"]:
			v := token.Value.(string)
			v = strings.ReplaceAll(v, ":", "")
			tsp.HOUR = v[0:2]
			tsp.MINUTE = v[2:4]
			tsp.SECOND = v[4:6]
		case tokmap["SUBSECOND"]:
			v := token.Value.(string)
			tsp.SUBSECOND = v
		case tokmap["ZONE"]:
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
		format := "20060102T150405"
		if tsp.SUBSECOND != "" {
			format = format + "." + strings.Repeat("9", len(tsp.SUBSECOND)-1)
		}

		if tsp.ZONE != "" {
			str = str + tsp.ZONE

			switch tsp.ZONE {
			case "Z":
				format = format + "Z"
			default:
				format = format + "-0700"
			}
		} else {
			str = str + "Z"

			format = format + "Z"
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
