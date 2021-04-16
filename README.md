# datetime parsing and handling library

Date and time functionality for ISO-8601 standard formats.

This library was initially written in an attempt to allow for the flexible
parsing of a range of input timestamp formats with an emphasis on ISO-8601
timestamps. The parsing produces Golang time.Time values. The desire to build
this functionality arose from an earlier project written in Java where timestamp
inputs were in varying formats that had to be read in reliably to produce
timestamps that could be used in the production of output documents.

Handling of ISO-8601 dates, periods, and timespans is included using the code
found at https://github.com/rickb777/date. That code is for the most part
included without modification as it is evidently well optimized and the result
of a great deal of thought and careful attention to accuracy. A period lexing
package is included but was mostly written for comparison purposes with the
existing parsing functionality. It will likely be removed as the existing
parsing is about 3x faster and has already benefitted from careful attention to
proper handling.

The timestamp parsing of ISO-8601 timestamps is weighted in favour of allowing
for some non-compliant formatting of parsed input as long as the compliance
issues do allow acceptance of ambiguous input. The library takes the approach of
initially trying to tokenize ISO-8601 format timestamps. If that step fails a
number of non-ISO-8601 formats are tried in a loop in a best effort to parse the
incoming timestamp. Finally, if the incoming timestamp is in a Unix timestamp
format an attempt is made to parse it as such.

None of the ISO-8601 time types are comprehensive in their handling of the
possible formats and variations in formats. The ISO-8601 standard is a large one.

The code for ISO-8601 date, period, and timespan handling is licenced under the
BSD-3-Clause Licence. You can read this licence at the end of the LICENCE file
for this project. The timestamp parsing, date, period, and timespan support are
used here in separate packages, for date, period, timespan, and timestamp.

Although testing has been done to attempt to ensure correct handling of the
types represented here, there are likely errors. Please submit any errors as
issues for this project.
