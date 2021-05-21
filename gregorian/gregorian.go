package gregorian

// IsLeap simply tests whether a given year is a leap year, using the Gregorian calendar algorithm.

// AdjustYear deal with year zero
// func AdjustYear(year int64) int64 {
// 	if year == 0 {
// 		return 1
// 	}

// 	return year
// }

// DaysInYear gives the number of days in a given year, according to the Gregorian calendar.
// func DaysInYear(year int64) int {
// 	if IsLeap(year) {
// 		return 366
// 	}

// 	return 365
// }

// DaysIn gives the number of days in a given month, according to the Gregorian calendar.

// daaysInMonth number of days in each month indexed by month number using zero
// padding
var daysInMonth = []int{
	0,
	31, // January
	28,
	31, // March
	30,
	31, // May
	30,
	31, // July
	31,
	30, // September
	31,
	30, // November
	31,
}
