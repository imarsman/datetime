# datetime parsing and handling library

Date and time functionality for ISO-8601 standard formats.

This library was initially written in an attempt to allow for the flexible
parsing of a range of input timestamp formats with an emphasis on ISO-8601
timestamps. The parsing produces Golang time.Time values. The desire to build
this functionality arose from an earlier project written in Java where timestamp
inputs were in varying formats that had to be read in reliably to produce
timestamps that could be used in the production of output documents.

None of the ISO-8601 time types implemented here are comprehensive in their
handling of the possible formats and variations in formats. The ISO-8601
standard is a large and complex one and the full specification costs money. The
goal is to provide useful coverage of the ISO-8601 standard.

Handling of ISO-8601 dates and timespans is included using the code found at
https://github.com/rickb777/date. That code is for the most part included
without modification. A separate period handling package is being tested that
replaces the one from rickb777. Until testing is finished the original period
parsing package will be kept. 

In the case of periods, care has been taken to avoid producing incorrect
durations when the spans evaluated exceed the maximum values for Golang's
duration type (int64), around 290 years for adjustment of years, months, and
days or hours, minutes, and seconds. Durations that would overflow will produce
an error. When allocating sub portions such as years the calculations are done
in terms of milliseconds instead of nanoseconds to allow for very fast
non-overflow int64 handling of periods up to about 290,000,000 years.  This is
part of support for a fractional portion of a period section to the level of
milliseconds with trailing zeros in the franctional seconds part removed. The
fractional conversion has been tested up to a value of 15 billion years. If a
fractional part exceeds the maximum int64 size an arbitrary precision decimal
library is used but it can deal with the larger values.

The timestamp parsing of ISO-8601 timestamps is weighted in favour of allowing
for some non-compliant formatting of parsed input as long as the compliance
issues do not allow acceptance of ambiguous input. The library takes the
approach of initially trying to tokenize ISO-8601 format timestamps. If that
step fails a number of non-ISO-8601 formats are tried in a loop in a best effort
to parse the incoming timestamp. If the incoming timestamp is in a Unix
timestamp format an attempt is made to parse it as such.

For an interesting task a date package is being worked on which has no time
portion. I am interested in being able to represent and make calculations on
dates with a very large range. Most time libraries are optimized for very fast
and accurate handling of time and date oriented operations and representing time
accurately. I am interested in exploring the representation of dates in such a
way that there will be no practical minimum or maximum date that be represented.
This is as much of a learning excercise as anything else and I will see where it
leads me. The dates are for the Gregorian calendar, which when extended to
represent dates prior to the adoption of the Gregorian calendar is called a
proleptic Gregorian calendar. The goal of the date library is to accurately
allow for calculations such as the addition of years, months, and days to a date
with an accurate accounting for things like leap years. So far that is going
pretty well but more testing will be required.

So far, although the underlying int64 type used to represent years can hold
years in a very wide range, including well before 1 BCE, calculations of things
like days including leap days and day of the week is written with positive year
numbers in mind. More work will need to be done to deal with negative years,
likely by using a year many years BCE as year zero so that all year calculations
can use addition rather than subtraction. The main goal for now is to correctly
handle CE year calculations.

There may be some value in making a clock package that could be paired with the
date package to allow for handling of dates and times outside of the normal
maximum and minimum date ranges built into programming languages like Go. Such a
combination would allow for the representation of ISO-8601 periods added to
arbitrary dates and times. Currently the period library will overflow the Go
time library when asked to represent a duration outside of the maximum value for
nanoseconds, based on the limits of the int64 type. This would be effectively
similar to using a larger numerical type like int256 or some library that
implemented the handling of large numbers like this.

The code for ISO-8601 date, gregorian calendar calculations, and timespan
handling is licenced under the BSD-3-Clause Licence. You can read this licence
at the end of the LICENCE file for this project.

Although testing has been done to attempt to ensure correct handling of the
types represented here, there are likely errors. Please submit any errors as
issues for this project. More work will be done to test the accuracy and
handling of the packages. More tests will be added to ensure that the full set
of functionality for each package is covered.

See: https://stackoverflow.com/questions/25065055/what-is-the-maximum-time-time-in-go/32620397#32620397
