# date-fns v4 Reference

## Installation

```bash
npm install date-fns
# For time zone support in v4:
# date-fns v4 has built-in TZ support via tz() context (no separate package needed)
```

## format — Token Reference

```ts
import { format } from 'date-fns'

const d = new Date('2024-03-15T14:05:09')

// Year
format(d, 'yyyy')      // "2024"    — 4-digit year
format(d, 'yy')        // "24"      — 2-digit year

// Month
format(d, 'MM')        // "03"      — 2-digit month (01-12)
format(d, 'M')         // "3"       — month without leading zero
format(d, 'MMM')       // "Mar"     — short month name
format(d, 'MMMM')      // "March"   — full month name
format(d, 'MMMMM')     // "M"       — narrow month letter

// Day
format(d, 'dd')        // "15"      — 2-digit day
format(d, 'd')         // "15"      — day without leading zero
format(d, 'do')        // "15th"    — ordinal day
format(d, 'D')         // "75"      — day of year (1-366)

// Day of week
format(d, 'EEE')       // "Fri"     — short weekday
format(d, 'EEEE')      // "Friday"  — full weekday
format(d, 'EEEEE')     // "F"       — narrow weekday letter
format(d, 'e')         // "6"       — day of week (1=Mon)
format(d, 'i')         // "5"       — ISO day of week (1=Mon, 7=Sun)

// Hour
format(d, 'HH')        // "14"      — 24h, 2-digit (00-23)
format(d, 'H')         // "14"      — 24h without leading zero
format(d, 'hh')        // "02"      — 12h, 2-digit (01-12)
format(d, 'h')         // "2"       — 12h without leading zero
format(d, 'a')         // "PM"      — AM/PM
format(d, 'aaa')       // "pm"      — lowercase am/pm

// Minute / Second
format(d, 'mm')        // "05"      — 2-digit minute
format(d, 'ss')        // "09"      — 2-digit second

// Timezone
format(d, 'xxx')       // "+09:00"  — ISO 8601 offset
format(d, 'z')         // "UTC"     — short timezone name

// Common patterns
format(d, 'yyyy-MM-dd')              // "2024-03-15"
format(d, 'dd/MM/yyyy')              // "15/03/2024"
format(d, 'MMMM do, yyyy')           // "March 15th, 2024"
format(d, 'EEEE, MMMM d, yyyy')      // "Friday, March 15, 2024"
format(d, "yyyy-MM-dd'T'HH:mm:ss")   // "2024-03-15T14:05:09" (literal in quotes)
format(d, 'h:mm a')                  // "2:05 PM"
```

## parse & parseISO

```ts
import { parse, parseISO, isValid } from 'date-fns'

// parseISO — ISO 8601 strings
const d1 = parseISO('2024-03-15')             // Date
const d2 = parseISO('2024-03-15T14:05:09Z')   // Date (UTC)
const d3 = parseISO('2024-03-15T14:05:09+09:00')

// parse — custom format string
const d4 = parse('15/03/2024', 'dd/MM/yyyy', new Date())
const d5 = parse('March 15, 2024', 'MMMM d, yyyy', new Date())
const d6 = parse('2:05 PM', 'h:mm a', new Date())

// Validate
isValid(d4)              // true
isValid(new Date('invalid'))  // false
```

## Arithmetic (all functions are immutable — return new Date)

```ts
import {
  addDays, addWeeks, addMonths, addYears, addHours, addMinutes, addSeconds,
  subDays, subWeeks, subMonths, subYears,
} from 'date-fns'

const base = new Date('2024-03-15')

addDays(base, 7)           // 2024-03-22
addWeeks(base, 2)          // 2024-03-29
addMonths(base, 3)         // 2024-06-15
addYears(base, 1)          // 2025-03-15
addHours(base, 5)          // 2024-03-15T05:00:00
addMinutes(base, 90)       // 2024-03-15T01:30:00
addSeconds(base, 3600)     // 2024-03-15T01:00:00

subDays(base, 1)           // 2024-03-14
subMonths(base, 1)         // 2024-02-15
subYears(base, 5)          // 2019-03-15

// add/sub with duration object
import { add, sub } from 'date-fns'
add(base, { years: 1, months: 2, days: 3 })
sub(base, { weeks: 1, hours: 6 })
```

## Comparison

```ts
import {
  isAfter, isBefore, isEqual,
  differenceInDays, differenceInMonths, differenceInYears,
  differenceInHours, differenceInMinutes,
  compareAsc, compareDesc,
} from 'date-fns'

const a = new Date('2024-01-01')
const b = new Date('2024-06-15')

isAfter(b, a)              // true  — b is after a
isBefore(a, b)             // true  — a is before b
isEqual(a, a)              // true

differenceInDays(b, a)     // 166   — b - a in days
differenceInMonths(b, a)   // 5
differenceInYears(b, a)    // 0
differenceInHours(b, a)    // 3984

// Sort arrays of dates
dates.sort(compareAsc)     // ascending
dates.sort(compareDesc)    // descending
```

## Boundaries

```ts
import {
  startOfDay, endOfDay,
  startOfWeek, endOfWeek,
  startOfMonth, endOfMonth,
  startOfYear, endOfYear,
  startOfHour, startOfMinute,
} from 'date-fns'

const d = new Date('2024-03-15T14:35:22')

startOfDay(d)              // 2024-03-15T00:00:00.000
endOfDay(d)                // 2024-03-15T23:59:59.999

startOfWeek(d)             // 2024-03-10 (Sunday by default)
startOfWeek(d, { weekStartsOn: 1 })  // 2024-03-11 (Monday)
endOfWeek(d, { weekStartsOn: 1 })    // 2024-03-17

startOfMonth(d)            // 2024-03-01T00:00:00.000
endOfMonth(d)              // 2024-03-31T23:59:59.999

startOfYear(d)             // 2024-01-01T00:00:00.000
endOfYear(d)               // 2024-12-31T23:59:59.999
```

## Formatting Helpers

```ts
import {
  formatDistance, formatDistanceToNow, formatRelative,
  intlFormat,
} from 'date-fns'

const past = new Date('2024-01-01')
const now  = new Date('2024-03-15')

formatDistance(past, now)                        // "about 2 months"
formatDistance(past, now, { addSuffix: true })   // "about 2 months ago"
formatDistanceToNow(past, { addSuffix: true })   // "about 2 months ago"

formatRelative(past, now)                        // "01/01/2024" (> 6 days ago)

// Intl-based formatting (locale-aware without importing locale)
intlFormat(d, { year: 'numeric', month: 'long', day: 'numeric' })
```

## Locale Support

```ts
import { format } from 'date-fns'
import { ko } from 'date-fns/locale'

format(new Date(), 'EEEE, MMMM do', { locale: ko })   // "금요일, 3월 15일"
formatDistanceToNow(date, { locale: ko, addSuffix: true })  // "2개월 전"
```

## v4 Time Zone Support (tz option)

```ts
import { formatInTimeZone, toZonedTime, fromZonedTime } from 'date-fns-tz'
// Note: date-fns-tz v3+ works with date-fns v4

// Format in a specific timezone
formatInTimeZone(new Date(), 'Asia/Seoul', 'yyyy-MM-dd HH:mm:ss')
// "2024-03-15 23:05:09"

// Convert UTC Date to zoned time (for display)
const seoulTime = toZonedTime(new Date(), 'Asia/Seoul')
format(seoulTime, 'HH:mm:ss')

// Convert zoned time input to UTC Date (for storage)
const utcDate = fromZonedTime('2024-03-15 14:00:00', 'Asia/Seoul')
```

## Common Patterns

```ts
// Is date in current month?
import { isSameMonth } from 'date-fns'
isSameMonth(date, new Date())

// Days in month
import { getDaysInMonth } from 'date-fns'
getDaysInMonth(new Date('2024-02-01'))   // 29 (leap year)

// Week number
import { getWeek, getISOWeek } from 'date-fns'
getISOWeek(new Date('2024-03-15'))       // 11

// Clamp date to range
import { clamp, max, min } from 'date-fns'
clamp(date, { start: minDate, end: maxDate })
max([date1, date2, date3])   // latest
min([date1, date2, date3])   // earliest

// Check same day
import { isSameDay } from 'date-fns'
isSameDay(new Date('2024-03-15'), new Date('2024-03-15T23:59:59'))  // true
```
