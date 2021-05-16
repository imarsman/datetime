# datetime parsing and handling library

Date and time functionality for ISO-8601 standard formats.

This library was initially written in an attempt to allow for the flexible
parsing of a range of input timestamp formats with an emphasis on ISO-8601
timestamps. The parsing produces Golang time.Time values. The desire to build
this functionality arose from an earlier project written in Java where timestamp
inputs were in varying formats that had to be read in reliably to produce
timestamps that could be used in the production of output documents.

Handling of ISO-8601 dates and timespans is included using the code found at
https://github.com/rickb777/date. That code is for the most part included
without modification. A separate period handling package is being tested that
replaces the one from rickb777. Until testing is finished the original period
parsing package will be kept. Care has been taken to avoid producing incorrect
durations when the spans evaluated exceed the maximum values for Golang's
duration type (int64), around 290 years for adjustment of years, months, and
days or hours, minutes, and seconds. The conversion to a duration will result in
an error if an overflow occurs. When allocating sub portions such as years the
calculations are done in terms of milliseconds instead of nanoseconds to allow
for very fast non-overflow int64 handling of periods up to about 290,000,000
years.  This is part of support for a fractional portion of a period section to
the level of milliseconds with trailing zeros in the franctional seconds part
removed. The fractional conversion has been tested up to a value of 15 billion
years. If a fractional part exceeds the maximum int64 size an arbitrary
precision decimal library is used.

The timestamp parsing of ISO-8601 timestamps is weighted in favour of allowing
for some non-compliant formatting of parsed input as long as the compliance
issues do not allow acceptance of ambiguous input. The library takes the
approach of initially trying to tokenize ISO-8601 format timestamps. If that
step fails a number of non-ISO-8601 formats are tried in a loop in a best effort
to parse the incoming timestamp. Finally, if the incoming timestamp is in a Unix
timestamp format an attempt is made to parse it as such.

None of the ISO-8601 time types are comprehensive in their handling of the
possible formats and variations in formats. The ISO-8601 standard is a large and
complex one and the full specification costs money. The goal is to provide
useful coverage of the ISO-8601 standard.

The code for ISO-8601 date, timespan handling is licenced under the BSD-3-Clause
Licence. You can read this licence at the end of the LICENCE file for this
project. The timestamp parsing, date, period, and timespan support are used here
in separate packages, for date, period, timespan, and timestamp.

Although testing has been done to attempt to ensure correct handling of the
types represented here, there are likely errors. Please submit any errors as
issues for this project. More work will be done to test the accuracy and
handling of the packages. More tests will be added to ensure that the full set
of functionality for each package is covered.
