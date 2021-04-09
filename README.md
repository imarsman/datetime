# datetime parsing and handling library

Date and time functionality for ISO-8601 standard formats. 

The code for ISO-8601 date handling, ISO-8601 period handling, and intervals is
from https://github.com/rickb777/date. It is licences under the BSD-3-Clause
Licence. You can read this licence at the end of the LICENCE file for this
project.

Code for handling parsing of timestamps is written by Ian Marsman. It is
primarily aimed at parsing ISO-8601 dates but will fall back on alternate
formats such as formats for SQL timestamps if the initial attempt to parse
ISO-8601 format fails.

The ISO-8601 time format standards are broad and more comprehensive than is
supported by any of the packages in this project. Things like repeating periods
and years with more than four digits are not currently handled.

It is very possible that errors have been made in the implementation of the
various ISO-8601 formats handling. Please let the author know if you discover
any outright errors.