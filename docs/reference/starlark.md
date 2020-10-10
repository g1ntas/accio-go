# Starlark

Starlark is a dialect of Python which was created by [Bazel](https://bazel.build/) and is used as a scripting language for blueprints. Specification of the language can be found at [google/starlark-go/spec.md](https://github.com/google/starlark-go/blob/master/doc/spec.md).

**NOTE**:  
* Sets (data type) are not supported
* All scripts within tags must be indented

## Context variables

In Starlark scripts, context variables can be accessed through the special `vars` variable:
```
variable -name="someContextVariable" << "test" >>

filename << vars['someContextVariable'] >>
```

## Inline scripts
When using markup's inline body syntax, Starlark scripts are automatically returned, thus a return statement is not needed and will throw an error if used.

In example below, both variables are identical:
```
variable -name="name" <<
    return "John Doe"
>>

variable -name="inlineName" << "John Doe" >>
```

## Helpers

In addition to Starlark built-in functions described in the specification, Accio also brings some additional helper functions.

### strftime
`strftime(format, time=time())` converts `time` to a string as specified by the `format` argument. If `time`
is not provided, the current time as returned by `time()` is used. `format` must be a string.

The following POSIX compliant directives can be embedded in the format string:

| pattern | description |
|:--------|:------------|
| %A      | national representation of the full weekday name |
| %a      | national representation of the abbreviated weekday |
| %B      | national representation of the full month name |
| %b      | national representation of the abbreviated month name |
| %C      | (year / 100) as decimal number; single digits are preceded by a zero |
| %c      | national representation of time and date |
| %D      | equivalent to %m/%d/%y |
| %d      | day of the month as a decimal number (01-31) |
| %e      | the day of the month as a decimal number (1-31); single digits are preceded by a blank |
| %F      | equivalent to %Y-%m-%d |
| %H      | the hour (24-hour clock) as a decimal number (00-23) |
| %h      | same as %b |
| %I      | the hour (12-hour clock) as a decimal number (01-12) |
| %j      | the day of the year as a decimal number (001-366) |
| %k      | the hour (24-hour clock) as a decimal number (0-23); single digits are preceded by a blank |
| %l      | the hour (12-hour clock) as a decimal number (1-12); single digits are preceded by a blank |
| %M      | the minute as a decimal number (00-59) |
| %m      | the month as a decimal number (01-12) |
| %n      | a newline |
| %p      | national representation of either "ante meridiem" (a.m.)  or "post meridiem" (p.m.)  as appropriate. |
| %R      | equivalent to %H:%M |
| %r      | equivalent to %I:%M:%S %p |
| %S      | the second as a decimal number (00-60) |
| %T      | equivalent to %H:%M:%S |
| %t      | a tab |
| %U      | the week number of the year (Sunday as the first day of the week) as a decimal number (00-53) |
| %u      | the weekday (Monday as the first day of the week) as a decimal number (1-7) |
| %V      | the week number of the year (Monday as the first day of the week) as a decimal number (01-53) |
| %v      | equivalent to %e-%b-%Y |
| %W      | the week number of the year (Monday as the first day of the week) as a decimal number (00-53) |
| %w      | the weekday (Sunday as the first day of the week) as a decimal number (0-6) |
| %X      | national representation of the time |
| %x      | national representation of the date |
| %Y      | the year with century as a decimal number |
| %y      | the year without century as a decimal number (00-99) |
| %Z      | the time zone name |
| %z      | the time zone offset from UTC |
| %%      | a '%' |

Example:
```
# Returns '2020-10-09 11:54:16'
variable -name="datetime" <<
  return strftime("%Y-%m-%d %H:%M:%S")
>>
```

### time
`time()` returns the time in seconds since the Unix epoch in UTC timezone as an integer.

Example:
```
# Returns 1602237780
variable -name="timestamp" <<
  return time()
>>
```