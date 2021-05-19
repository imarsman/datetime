package utility

import "strings"

// DigitCount count digits in an int64 number
func DigitCount(number int64) int64 {
	var count int64 = 0
	for number != 0 {
		number /= 10
		count++
	}
	return count
}

// BytesToString convert byte list to string with no allocation
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
