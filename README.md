# datetime parsing and handling library

Time functionality for ISO-8601 standard formats.

This library was initially written in an attempt to allow for the flexible
parsing of a range of input timestamp formats with an emphasis on ISO-8601
timestamps. The parsing produces Golang time.Time values. The desire to build
this functionality arose from an earlier project written in Java where timestamp
inputs were in varying formats that had to be read in reliably to produce
timestamps that could be used in the production of output documents.

Currently timestamp and well as ISO-8601 period parsing are implemented. 

For timestamp parsing, if the format to be parsed to Golang time is consistent
it would be easier and possibly faster to just use the Golang Parse function.
The value of the timestamp parsing is in its flexibility in handling varying
input formats with relatively low overhead.

For ISO-8601 period parsing, correct handling and normalization of periods with
good performance is implemented. Where period parts would overflow the value for
that part (e.g. int64) the CockroachDB arbitrary precision decimal library is
used. This is less efficient and uses more resources than integer arithmetic but
is avoided unless it is required. Periods are calculated with a maximum
precision of milliseconds.

In the case of periods, care has been taken to avoid producing incorrect
durations when the spans evaluated exceed the maximum values for Golang's
duration type (int64), around 290 years for adjustment of years, months, and
days or hours, minutes, and seconds. Durations that would overflow will produce
an error. As also mentioned above, when allocating sub portions such as years
the calculations are done in terms of milliseconds instead of nanoseconds to
allow for very fast non-overflow int64 handling of periods up to about
290,000,000 years.  This is part of support for a fractional portion of a period
section to the level of milliseconds with trailing zeros in the franctional
seconds part removed. The fractional conversion has been tested up to a value of
15 billion years.

The timestamp parsing of ISO-8601 timestamps is weighted in favour of allowing
for some non-compliant formatting of parsed input as long as the compliance
issues do not allow acceptance of ambiguous input. The library takes the
approach of initially trying to tokenize ISO-8601 format timestamps. If that
step fails a number of non-ISO-8601 formats are tried in a loop in a best effort
to parse the incoming timestamp. If the incoming timestamp is in a Unix
timestamp format an attempt is made to parse it as such.

The develop branch contains work done on a date package which is based on a zero
date of 1 CE. This package needs more work and is not included in the main
branch. When completed the date package will allow data calculations with spans
in the billions of years using the proleptic Julian calendar.

Although testing has been done to attempt to ensure correct handling of the
types represented here, there are likely errors. Please submit any errors as
issues for this project. More work will be done to test the accuracy and
handling of the packages. More tests will be added to ensure that the full set
of functionality for each package is covered.

See: https://stackoverflow.com/questions/25065055/what-is-the-maximum-time-time-in-go/32620397#32620397
