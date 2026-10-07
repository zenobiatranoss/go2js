"use strict";
const go2jsNativeMap = globalThis.Map;

const go2jsNativeSet = globalThis.Set;

const go2jsNativeDate = globalThis.Date;

const go2jsNativeTypeError = globalThis.TypeError;

const go2jsNativeRangeError = globalThis.RangeError;

// The words Go puts in front of every fault it raises on its own.
const go2jsRuntimeErrorPrefix = "runtime error: ";

var go2jsTextEncoder = new TextEncoder();

var go2jsStrictDecoder = new TextDecoder("utf-8", { fatal: true });

// Decoding reads a run of bytes as the runes it spells, and a byte that is not
// the UTF-8 of a rune as the byte it is, since a Go string keeps it either way.
function go2jsDecodeBytes(bytes) {
	if (bytes.length === 0) {
		return "";
	}

	try {
		// Text is the common case, and the platform reads it far faster than a
		// walk over every byte of it could.
		return go2jsStrictDecoder.decode(bytes);
	} catch (error) {
		return go2jsDecodeBytesByteByByte(bytes);
	}
}

function go2jsDecodeBytesByteByByte(bytes) {
	let text = "";
	let i = 0;

	while (i < bytes.length) {
		const first = bytes[i];
		let codePoint;
		let size;

		if (first < 0x80) {
			codePoint = first;
			size = 1;
		} else if ((first & 0xe0) === 0xc0) {
			codePoint = first & 0x1f;
			size = 2;
		} else if ((first & 0xf0) === 0xe0) {
			codePoint = first & 0x0f;
			size = 3;
		} else if ((first & 0xf8) === 0xf0) {
			codePoint = first & 0x07;
			size = 4;
		} else {
			text += go2jsRawByteUnit(first);
			i++;
			continue;
		}

		if (i + size > bytes.length) {
			text += go2jsRawByteUnit(first);
			i++;
			continue;
		}

		let wellFormed = true;

		for (let offset = 1; offset < size; offset++) {
			const next = bytes[i + offset];

			if ((next & 0xc0) !== 0x80) {
				wellFormed = false;
				break;
			}

			codePoint = (codePoint << 6) | (next & 0x3f);
		}

		// A rune has to be as long as it was written to be, has to be a rune at
		// all, and has to be one the encoding is allowed to spell. A sequence
		// that is not any of those is a byte and not a character.
		if (!wellFormed || codePoint > 0x10ffff || (codePoint >= 0xd800 && codePoint <= 0xdfff) ||
			(size === 2 && codePoint < 0x80) || (size === 3 && codePoint < 0x800) || (size === 4 && codePoint < 0x10000)) {
			text += go2jsRawByteUnit(first);
			i++;
			continue;
		}

		text += String.fromCodePoint(codePoint);
		i += size;
	}

	return text;
}

// A Go string is a run of bytes, and a JavaScript string is a run of runes, so
// a byte that is not the UTF-8 of a rune has nowhere of its own to sit in one.
// Those bytes are held as lone low surrogates, which no rune is ever written
// with, and are read back as the very bytes they were wherever they are
// measured, sliced, indexed or written out.
function go2jsRawByteUnit(byte) {
	return String.fromCharCode(0xdc00 + (byte & 255));
}

function go2jsHasRawBytes(value) {
	for (let i = 0; i < value.length; i++) {
		const unit = value.charCodeAt(i);

		// A low surrogate is the other half of a rune when the unit before it is
		// a high surrogate, and a raw byte of its own when nothing is.
		if (unit >= 0xdc00 && unit <= 0xdfff && (i === 0 || value.charCodeAt(i - 1) < 0xd800 || value.charCodeAt(i - 1) > 0xdbff)) {
			return true;
		}
	}

	return false;
}

function go2jsStringToBytes(value) {
	const bytes = [];
	let start = 0;
	let i = 0;

	while (i < value.length) {
		const unit = value.charCodeAt(i);

		if (unit >= 0xdc00 && unit <= 0xdfff && (i === 0 || value.charCodeAt(i - 1) < 0xd800 || value.charCodeAt(i - 1) > 0xdbff)) {
			if (i > start) {
				go2jsAppendEncodedBytes(bytes, value.slice(start, i));
			}

			bytes.push(unit - 0xdc00);
			i++;
			start = i;
			continue;
		}

		i++;
	}

	// Text with no raw byte in it is the case for nearly every string, and is
	// left to the platform to write out whole.
	if (start === 0) {
		return Array.from(go2jsTextEncoder.encode(value));
	}

	if (start < value.length) {
		go2jsAppendEncodedBytes(bytes, value.slice(start));
	}

	return bytes;
}

function go2jsAppendEncodedBytes(bytes, text) {
	const encoded = go2jsTextEncoder.encode(text);

	for (let i = 0; i < encoded.length; i++) {
		bytes.push(encoded[i]);
	}
}

function go2jsOSArgs() {
    return process.argv.slice(1);
}

function go2jsOSGetenv(name) {
    return process.env[String(name)] || "";
}

function go2jsOSSetenv(name, value) {
    process.env[String(name)] = String(value);
    return null;
}

// go2jsOSStatObject builds the FileInfo a stat call answers with, which is the
// same thing a file answers a stat of its own with.
function go2jsOSStatObject(stats, path) {
    return [go2jsOSFileInfo(path === undefined || path === null ? "" : String(path), stats), null];
}

function go2jsOSStat(path) {
    try {
        return go2jsOSStatObject(require("fs").statSync(String(path)), path);
    } catch (err) {
        return [null, go2jsOSHostError(err, "stat", String(path))];
    }
}

function go2jsStringsNewReader(value) {
    let position = 0;
    const text = String(value);

    return {
        __go2js_text: text,
        Read: function(buffer) {
            if (position >= text.length) {
                return [0, go2jsIOEOF()];
            }

            const count = Math.min(buffer.length, text.length - position);

            for (let i = 0; i < count; i++) {
                buffer[i] = text.charCodeAt(position + i);
            }

            position += count;
            return [count, null];
        }
    };
}

// go2jsUnwrap returns the concrete value behind an interface wrapper so
// runtime helpers can call methods on it.
function go2jsUnwrap(value) {
	if (value !== null && value !== undefined && value.__go2js_interface === true) {
		return value.value;
	}

	return value;
}

function go2jsIOReadAll(reader) {
	reader = go2jsUnwrap(reader);

	const chunks = [];
    const buffer = new Array(4096);

    while (true) {
        const result = go2jsCallNow(reader.Read, reader, [buffer]);
        const read = Number(result[0]) || 0;

        if (read > 0) {
            chunks.push(...buffer.slice(0, read));
        }

        if (result[1] !== null && result[1] !== undefined) {
            // The end of a text is not a failure of reading it, whether it was
            // reported as the end itself or as an error that ends in it.
            if (result[1] === go2jsIOEOF() || go2jsErrorsIs(result[1], go2jsIOEOF())) {
                break;
            }

            return [null, result[1]];
        }

        if (read <= 0) {
            break;
        }
    }

    return [String.fromCharCode(...chunks), null];
}

function go2jsTimeDate(year, month, day, hour, minute, second, nanosecond, location) {
    const name = go2jsTimeLocationName(location);

    // The pieces are the clock of the location rather than of UTC, so the moment
    // is the one at which that clock shows them. A zone that moves its offset
    // across the year is asked twice, since the offset at the moment found for
    // one offset may not be the offset that moment actually falls under.
    let date = new Date(0);

    date.setUTCFullYear(Number(year), Number(month) - 1, Number(day));
    date.setUTCHours(
        Number(hour),
        Number(minute),
        Number(second),
        Math.floor(Number(nanosecond) / 1000000)
    );

    const offset = go2jsTimeLocationOffsetSeconds(location, date);

    if (offset === 0) {
        const at = go2jsTimeValue(date);

        at.__go2js_nsec = go2jsTimeFine(Number(nanosecond));
        at.__go2js_location = name;
        at.__go2js_zone = location !== null && location !== undefined && typeof location === "object"
            ? location.__go2js_zone
            : null;

        return at;
    }

    let instant = new Date(date.getTime() - offset * 1000);
    const second_ = go2jsTimeLocationOffsetSeconds(location, instant);

    if (second_ !== offset) {
        instant = new Date(date.getTime() - second_ * 1000);
    }

    const at = go2jsTimeValue(instant);

    at.__go2js_nsec = go2jsTimeFine(Number(nanosecond));
    at.__go2js_location = name;
    at.__go2js_zone = location !== null && location !== undefined && typeof location === "object"
        ? location.__go2js_zone
        : null;

    return at;
}

// go2jsTimeFine is what a count of nanoseconds holds finer than the millisecond
// the host keeps a moment to, brought inside the count of a millisecond it is
// counted in.
function go2jsTimeFine(nanosecond) {
    if (!Number.isFinite(nanosecond)) {
        return 0;
    }

    return ((Math.trunc(nanosecond) % 1000000) + 1000000) % 1000000;
}

function go2jsMonthName(month) {
    const names = [
        "January", "February", "March", "April", "May", "June",
        "July", "August", "September", "October", "November", "December"
    ];

    return names[month - 1] || "";
}

function go2jsWeekdayName(weekday) {
    const names = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];

    return names[weekday] || "";
}

// go2jsTimeLocation names where a moment is being read, which is what a location
// prints and what a moment read in it carries.
// A location is a name and, where the name stands for one, what it stands for:
// a number of seconds east of UTC for a zone fixed at one, the name of a zone
// the host knows for a zone it keeps, and nothing at all for UTC itself and for
// the zone of the machine, which the host answers for.
function go2jsTimeLocation(name, zone) {
    return {
        __go2js_location: name,
        __go2js_zone: zone === undefined ? null : zone,
        String: function() {
            return this.__go2js_location;
        }
    };
}

// go2jsTimeZoneFormat is the host's way of reading a moment in a zone it knows,
// which is where the offset of that zone at that moment is read from. One
// formatter is kept for a zone, since making one costs more than asking it a
// question does.
const go2jsTimeZoneFormats = new Map();

function go2jsTimeZoneFormat(name) {
    if (go2jsTimeZoneFormats.has(name)) {
        return go2jsTimeZoneFormats.get(name);
    }

    let formatter = null;

    try {
        formatter = new Intl.DateTimeFormat("en-US", {
            timeZone: name,
            hourCycle: "h23",
            year: "numeric",
            month: "2-digit",
            day: "2-digit",
            hour: "2-digit",
            minute: "2-digit",
            second: "2-digit"
        });
    } catch (error) {
        formatter = null;
    }

    go2jsTimeZoneFormats.set(name, formatter);

    return formatter;
}

// go2jsTimeZoneNameLocales are the bundles a zone is looked for a name in. The
// name of a zone is carried in the bundle of the place it belongs to rather than
// in one bundle for the whole world, so Berlin is named in the British bundle and
// not in the American one, and a zone is asked of each until one answers with
// something better than an offset.
const go2jsTimeZoneNameLocales = ["und", "en-GB", "en-US", "en-IN", "ja-JP", "pt-BR", "en-AU"];

const go2jsTimeZoneNameFormats = new Map();

function go2jsTimeZoneNameFormat(name) {
    if (go2jsTimeZoneNameFormats.has(name)) {
        return go2jsTimeZoneNameFormats.get(name);
    }

    const formats = [];

    for (const locale of go2jsTimeZoneNameLocales) {
        try {
            formats.push(new Intl.DateTimeFormat(locale, {timeZone: name, timeZoneName: "short"}));
        } catch (error) {
            break;
        }
    }

    go2jsTimeZoneNameFormats.set(name, formats);

    return formats;
}

// go2jsTimeZoneAbbreviation is the name a zone is known by at a moment, which is
// what a layout naming a zone writes out and what Zone answers with. A bundle
// that answers with an offset has no name to give, and a zone with no name is
// written out as the offset it keeps, which is what Go writes for such a zone.
function go2jsTimeZoneAbbreviation(name, date) {
    for (const formatter of go2jsTimeZoneNameFormat(name)) {
        for (const part of formatter.formatToParts(date)) {
            if (part.type !== "timeZoneName") {
                continue;
            }

            const short = String(part.value).trim();

            if (short !== "" && !/^GMT[+-]/.test(short)) {
                return short;
            }
        }
    }

    return null;
}

// go2jsTimeZoneOffsetSeconds is the number of seconds a zone is east of UTC at
// a moment, which is the difference between the moment as UTC reads it and the
// same moment as that zone reads it.
function go2jsTimeZoneOffsetSeconds(name, date) {
    const formatter = go2jsTimeZoneFormat(name);

    if (formatter === null) {
        return 0;
    }

    const read = {};

    for (const part of formatter.formatToParts(date)) {
        if (part.type !== "literal") {
            read[part.type] = parseInt(part.value, 10);
        }
    }

    if (!Number.isFinite(read.year) || !Number.isFinite(read.month) ||
        !Number.isFinite(read.day) || !Number.isFinite(read.hour)) {
        return 0;
    }

    const asRead = Date.UTC(read.year, read.month - 1, read.day, read.hour, read.minute || 0, read.second || 0);
    const asUTC = Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), date.getUTCDate(),
        date.getUTCHours(), date.getUTCMinutes(), date.getUTCSeconds());

    // A zone east of UTC reads the moment later than UTC does, so the wall clock
    // of the zone is ahead of UTC by however far east it keeps.
    return Math.round((asRead - asUTC) / 1000);
}

// go2jsTimeLocationOffsetSeconds is the offset of a location at a moment. A zone
// fixed at a number of seconds is always that far east, a zone the host knows
// is asked, and the zone of the machine is the one the moment itself answers.
function go2jsTimeLocationOffsetSeconds(location, date) {
    const zone = location !== null && location !== undefined ? location.__go2js_zone : null;

    if (zone === null || zone === undefined) {
        const name = go2jsTimeLocationName(location);

        if (name === "UTC") {
            return 0;
        }

        if (name === "Local") {
            const host = go2jsTimeHostZoneName();

            if (host !== "UTC") {
                return go2jsTimeZoneOffsetSeconds(host, date);
            }
        }

        return go2jsTimeZoneOffsetSeconds(name, date);
    }

    if (typeof zone === "number") {
        return zone;
    }

    return go2jsTimeZoneOffsetSeconds(zone, date);
}

// go2jsTimeHostZoneName is the name of the zone the machine keeps its clock in.
// A machine told which zone to keep it in is named by that, since Go keeps the
// name it was given rather than renaming it to whatever the host calls the same
// place; a machine left to itself is named by what the host resolved.
function go2jsTimeHostZoneName() {
    const told = go2jsTimeToldZoneName();

    if (told !== null) {
        return told;
    }

    let name = null;

    try {
        name = new Intl.DateTimeFormat().resolvedOptions().timeZone;
    } catch (error) {
        name = null;
    }

    return name === undefined || name === null || name === "" ? "UTC" : String(name);
}

// go2jsTimeToldZoneName is the zone the machine was told to keep its clock in, or
// nothing for a machine that was told nothing, or nothing for a telling the host
// does not know a zone by, which is a rule of its own rather than a name.
function go2jsTimeToldZoneName() {
    let told = null;

    try {
        told = process.env.TZ;
    } catch (error) {
        told = null;
    }

    if (told === undefined || told === null) {
        return null;
    }

    // A telling may name the system database with a mark in front of it, as the
    // C library allows, which is not part of the name.
    let name = String(told);

    if (name.startsWith(":")) {
        name = name.slice(1);
    }

    if (name === "" || go2jsTimeZoneFormat(name) === null) {
        return null;
    }

    return name;
}

// go2jsTimeQuotient is the whole part of one count divided by another, read
// without trusting the division: a count long enough for the division to round
// can answer with a quotient a whole part above the one it names, which would
// put a moment a second later than it is.
function go2jsTimeQuotient(value, divisor) {
    let whole = Math.floor(value / divisor);

    if (whole * divisor > value) {
        whole -= 1;
    } else if ((whole + 1) * divisor <= value) {
        whole += 1;
    }

    return whole;
}

// go2jsTimeNanos is the nanoseconds a moment carries within its second. The host
// keeps a moment to the millisecond, so what is finer than that is carried beside
// it, and Go counts a second in nanoseconds rather than in milliseconds.
function go2jsTimeNanos(value) {
    return value.value.getUTCMilliseconds() * 1000000 + (value.__go2js_nsec === undefined ? 0 : value.__go2js_nsec);
}

// go2jsTimeNanosCarry is the finer part of a moment put on a new one, which is
// what moving a moment about carries with it, and so is the reading of the
// monotonic clock a moment was made against.
function go2jsTimeNanosCarry(value, next) {
    next.__go2js_nsec = value.__go2js_nsec === undefined ? 0 : value.__go2js_nsec;

    if (value.__go2js_mono !== undefined) {
        next.__go2js_mono = value.__go2js_mono;
    }

    return next;
}

// go2jsTimeLocalDate is the moment as a location reads it, and it is what the
// fields of a moment are read from: the fields are the ones a clock in that
// location would show, and the moment itself has not moved.
function go2jsTimeLocalDate(date, location) {
    if (date === null || date === undefined || !(date instanceof Date)) {
        return date;
    }

    const offset = go2jsTimeLocationOffsetSeconds(location, date);

    return offset === 0 ? date : new Date(date.getTime() + offset * 1000);
}

// go2jsTimeLocationName is the name a location prints, which is the name it was
// given and UTC for the location a moment carries when it was never told.
function go2jsTimeLocationName(location) {
    if (location === null || location === undefined) {
        return "UTC";
    }

    if (typeof location === "object" && typeof location.__go2js_location === "string") {
        return location.__go2js_location;
    }

    return String(location);
}

// go2jsTimeZero is the zero time, which is January 1 of year 1 at midnight UTC.
// The year is out of the range that Date.UTC maps straight onto, so the
// instant is given as the number of milliseconds before the epoch instead.
function go2jsTimeZero() {
    const zero = go2jsTimeValue(new Date(-62135596800000));

    // The zero time is a moment like any other, so it is marked as this one
    // rather than only being the moment it stands for, since a moment far from
    // the clock of the host is not a moment nothing was ever told.
    zero.__go2js_time_zero = true;

    return zero;
}

// go2jsTimeMonoOrigin is the moment the process was born on the clock it is
// measured by, from which the reading a moment is made against is counted, so
// that a monotonic reading is a count from the birth of the process the way Go
// counts one.
const go2jsTimeMonoOrigin = Date.now() * 1000000;

function go2jsTimeValue(date) {
    // A moment that already carries the methods of a time is left as it is,
    // since wrapping it again would bury the moment it holds.
    if (date !== null && date !== undefined && date.value instanceof Date) {
        return date;
    }

    // The fields of a moment are the ones a clock in its own location would
    // show, so they are read from the moment as that location reads it rather
    // than from the moment as UTC reads it.
    const fields = function(self) {
        return go2jsTimeLocalDate(self.value, self);
    };

    const self = {
        value: date,
        Year: function() {
            return fields(this).getUTCFullYear();
        },
        Month: function() {
            return fields(this).getUTCMonth() + 1;
        },
        Day: function() {
            return fields(this).getUTCDate();
        },
        Hour: function() {
            return fields(this).getUTCHours();
        },
        Minute: function() {
            return fields(this).getUTCMinutes();
        },
        Second: function() {
            return fields(this).getUTCSeconds();
        },
        Nanosecond: function() {
            return go2jsTimeNanos(this);
        },
        Clock: function() {
            return [this.Hour(), this.Minute(), this.Second()];
        },
        YearDay: function() {
            const local = fields(this);
            const year = local.getUTCFullYear();
            const start = Date.UTC(year, 0, 1);

            return Math.floor((Date.UTC(year, local.getUTCMonth(), local.getUTCDate()) - start) / 86400000) + 1;
        },
        Weekday: function() {
            return fields(this).getUTCDay();
        },
        UTC: function() {
            const next = go2jsTimeNanosCarry(this, go2jsTimeValue(this.value));

            next.__go2js_location = "UTC";
            next.__go2js_zone = null;

            return next;
        },
        Local: function() {
            const next = go2jsTimeNanosCarry(this, go2jsTimeValue(this.value));

            next.__go2js_location = go2jsTimeHostZoneName();
            next.__go2js_zone = next.__go2js_location === "UTC" ? null : next.__go2js_location;

            return next;
        },
        In: function(location) {
            const next = go2jsTimeNanosCarry(this, go2jsTimeValue(this.value));

            next.__go2js_location = go2jsTimeLocationName(location);
            next.__go2js_zone = location !== null && location !== undefined &&
                typeof location === "object" ? location.__go2js_zone : null;

            return next;
        },
        Location: function() {
            return go2jsTimeLocation(
               	this.__go2js_location === undefined ? "UTC" : this.__go2js_location,
               	this.__go2js_zone === undefined ? null : this.__go2js_zone
            );
        },
        Zone: function() {
            return [go2jsTimeFormatZone(this.value, this).name, go2jsTimeLocationOffsetSeconds(this, this.value)];
        },
        AddDate: function(years, months, days) {
            const date = new Date(this.value.getTime());
            date.setUTCFullYear(date.getUTCFullYear() + Number(years), date.getUTCMonth() + Number(months), date.getUTCDate() + Number(days));
            const next = go2jsTimeValue(date);

            next.__go2js_location = this.__go2js_location;
            next.__go2js_zone = this.__go2js_zone;

            return next;
        },
        Unix: function() {
            return Math.floor(this.value.getTime() / 1000);
        },
        UnixNano: function() {
            // The count of nanoseconds since the epoch is counted in seconds
            // and then in what is left of the second, which for a moment before
            // the epoch is a count of the second it sits before it.
            return go2jsTimeQuotient(this.value.getTime(), 1000) * 1000000000 + go2jsTimeNanos(this);
        },
        IsZero: function() {
            return this.value.getTime() === 0 || this.__go2js_time_zero === true;
        },
        Before: function(other) {
            return this.value.getTime() < go2jsTimeDateOf(other).getTime();
        },
        After: function(other) {
            return this.value.getTime() > go2jsTimeDateOf(other).getTime();
        },
        Equal: function(other) {
            return this.value.getTime() === go2jsTimeDateOf(other).getTime();
        },
        Format: function(layout) {
            return go2jsTimeFormat(this.value, String(layout), this);
        },
        String: function() {
            return go2jsTimeString(this);
        },
        Add: function(duration) {
            const nanos = go2jsDurationNanos(duration);
            const held = this.__go2js_nsec === undefined ? 0 : this.__go2js_nsec;
            let fine = held + (nanos % 1000000);
            let millis = this.value.getTime() + go2jsTimeQuotient(nanos, 1000000);

            // A moment moved past the end of a second carries what it was holding
            // into the next second rather than dropping it, which is what moving a
            // moment by a duration does in Go as well.
            if (fine >= 1000000) {
                fine -= 1000000;
                millis += 1;
            } else if (fine < 0) {
                fine += 1000000;
                millis -= 1;
            }

            const next = go2jsTimeValue(new Date(millis));

            next.__go2js_nsec = fine;

            // A moment moved by a duration keeps the reading of the monotonic
            // clock it carried, moved by the same duration, which is what Go
            // keeps of a moment it moves.
            if (this.__go2js_mono !== undefined) {
                next.__go2js_mono = this.__go2js_mono + nanos;
            }

            next.__go2js_location = this.__go2js_location;
            next.__go2js_zone = this.__go2js_zone;

            return next;
        },
        Sub: function(other) {
            // The span between two moments is counted in nanoseconds, so it is
            // counted in whole seconds and in what each of them holds within its
            // second rather than in the milliseconds the two are kept to.
            const that = go2jsTimeOf(other);
            const here = go2jsTimeQuotient(this.value.getTime(), 1000);
            const there = go2jsTimeQuotient(that.value.getTime(), 1000);

            return go2jsDuration((here - there) * 1000000000 + go2jsTimeNanos(this) - go2jsTimeNanos(that));
        },
        Truncate: function(duration) {
            const rounded = go2jsTimeRounded(this, duration, false);
            const next = go2jsTimeValue(new Date(rounded.millis));

            next.__go2js_nsec = rounded.nsec;

            if (this.__go2js_mono !== undefined) {
                next.__go2js_mono = this.__go2js_mono;
            }

            next.__go2js_location = this.__go2js_location;
            next.__go2js_zone = this.__go2js_zone;

            return next;
        },
        Round: function(duration) {
            const rounded = go2jsTimeRounded(this, duration, true);
            const next = go2jsTimeValue(new Date(rounded.millis));

            next.__go2js_nsec = rounded.nsec;

            if (this.__go2js_mono !== undefined) {
                next.__go2js_mono = this.__go2js_mono;
            }

            next.__go2js_location = this.__go2js_location;
            next.__go2js_zone = this.__go2js_zone;

            return next;
        }
    };

    return self;
}

// go2jsTimeRoundedTo gives a moment rounded to a multiple of a span counted from
// the zero time, which is where Go counts the multiples of a span from. A span
// under half a millisecond rounds to the millisecond the host can hold, since a
// moment it cannot hold cannot be written down.
// go2jsTimeOf is a moment as it is held, whether it was handed over as one or as
// the date standing behind one.
function go2jsTimeOf(value) {
    const raw = go2jsUnwrap(value);

    if (raw !== null && typeof raw === "object" && raw.value instanceof Date) {
        return raw;
    }

    return go2jsTimeValue(raw instanceof Date ? raw : new Date(Number(raw)));
}

// go2jsTimeRounded is a moment cut back to a multiple of a span, counted from the
// zero time as Go counts one. A span finer than a millisecond is measured in the
// nanoseconds a moment holds within its second, which the host does not keep and
// which is carried beside the moment for that reason.
function go2jsTimeRounded(value, duration, nearest) {
    const held = value.__go2js_nsec === undefined ? 0 : value.__go2js_nsec;
    const span = Math.trunc(Number(go2jsDurationNanos(duration)));

    if (!(span > 0)) {
        return {millis: value.value.getTime(), nsec: held};
    }

    // A span of a millisecond or more is measured in milliseconds, which a moment
    // is kept to, and leaves nothing of the second behind: a moment cut back to
    // one does not hold what it was holding.
    if (span >= 1000000) {
        return {millis: go2jsTimeRoundedTo(value.value, duration, nearest), nsec: 0};
    }

    // A moment counted in nanoseconds since the zero time is a count far past what
    // a whole number of doubles holds, so a span finer than a millisecond is
    // counted in whole numbers rather than in ones that lose their ends.
    const since = BigInt(value.value.getTime() - go2jsTimeZeroMillis) * 1000000n + BigInt(held);
    const width = BigInt(span);
    let rest = since % width;

    if (rest < 0n) {
        rest += width;
    }

    let counted = since - rest;

    if (nearest && width - rest < rest) {
        counted += width;
    }

    return {
        millis: go2jsTimeZeroMillis + Number(counted / 1000000n),
        nsec: Number(((counted % 1000000n) + 1000000n) % 1000000n)
    };
}

function go2jsTimeRoundedTo(date, duration, nearest) {
    const span = Number(go2jsDurationNanos(duration)) / 1000000;

    if (!(span > 0)) {
        return date.getTime();
    }

    const since = date.getTime() - go2jsTimeZeroMillis;
    const counted = Math.floor(since / span) * span;
    const rounded = nearest && since - counted >= span / 2 ? counted + span : counted;

    return go2jsTimeZeroMillis + rounded;
}

// go2jsTimeZeroMillis is the moment Go counts from: January 1 of year 1 at
// midnight UTC.
const go2jsTimeZeroMillis = -62135596800000;

function go2jsTimeDateOf(value) {
    return value !== null && value !== undefined && value.value instanceof Date ? value.value : value;
}

// go2jsTimeMonths and go2jsTimeDays are the names a layout asks for by name,
// which are written out in full rather than taken from a Date, whose names are
// the ones of the host rather than the ones a Go program reads.
const go2jsTimeMonths = [
    "January", "February", "March", "April", "May", "June",
    "July", "August", "September", "October", "November", "December"
];

const go2jsTimeDays = [
    "Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"
];

// go2jsTimeString is a moment written the way Go writes it when it is printed,
// which is the layout "2006-01-02 15:04:05.999999999 -0700 MST" followed, for a
// moment that carries a reading of the monotonic clock, by how far into that
// clock the moment was made, as " m=+ddd.fffffffff".
function go2jsTimeString(t) {
    let text = go2jsTimeFormat(t.value, "2006-01-02 15:04:05.999999999 -0700 MST", t);

    if (t.__go2js_mono === undefined) {
        return text;
    }

    let mono = t.__go2js_mono;
    let sign = "+";

    if (mono < 0) {
        sign = "-";
        mono = -mono;
    }

    const seconds = Math.trunc(mono / 1000000000);
    const nanos = mono % 1000000000;
    const minutes = Math.trunc(seconds / 1000000000);

    let suffix = " m=" + sign;
    let width = 0;

    // The reading is written in up to three groups of nine digits, of which the
    // first is left out when it is what holds the whole of the count, the way Go
    // leaves it out for a process that began in living memory.
    if (minutes !== 0) {
        suffix += String(minutes);
        width = 9;
    }

    suffix += String(seconds).padStart(width, "0") + "." + String(nanos).padStart(9, "0");

    return text + suffix;
}

// go2jsTimeFormat writes a time the way a layout asks for it. A layout is read
// one piece at a time and what a piece stands for is written out as it is, so
// that digits standing for a year are never read again as a piece standing for
// a month, which is what a layout replacing every piece in turn would do.
// go2jsTimeFormatZone is how a moment's location is written out: the name the
// location answers to, the zulu a layout asks a zone for when it stands at
// zero, and the signed number of hours and minutes east of UTC.
// go2jsTimeFormatZone is how the location of a moment is written out: the name
// the location answers to, and the offset in the two shapes a layout asks for,
// which differ only in whether they carry the mark separating hours from minutes
// and whether a zone standing at zero is written as the letter for it.
function go2jsTimeFormatZone(date, location) {
    const name = go2jsTimeLocationName(location);
    const seconds = go2jsTimeLocationOffsetSeconds(location, date);
    const zone = location !== null && location !== undefined ? location.__go2js_zone : null;
    const number = go2jsTimeFormatZoneNumber(seconds);
    const zulu = seconds === 0 ? "Z" : number;

    if (name === "UTC") {
        return {name: "UTC", zulu: zulu, number: number};
    }

    if (name === "Local") {
        return {name: go2jsTimeLocalZoneName(date), zulu: zulu, number: number};
    }

    if (typeof zone === "number") {
        return {name: name === "" ? number : name, zulu: zulu, number: number};
    }

    const known = go2jsTimeZoneAbbreviation(name, date);

    return {name: known === null ? go2jsTimeFormatZoneCompact(seconds) : known, zulu: zulu, number: number};
}

// go2jsTimeFormatZoneCompact is the offset of a zone written the way a zone
// without a name is named, which is the offset without the mark between its
// hours and its minutes.
function go2jsTimeFormatZoneCompact(seconds) {
    return go2jsTimeFormatZoneNumber(seconds).replace(":", "");
}

// go2jsTimeFormatZoneNumber is the signed number of hours and minutes a zone is
// east of UTC, which is what Go writes out for a zone that has no name.
function go2jsTimeFormatZoneNumber(seconds) {
    const sign = seconds < 0 ? "-" : "+";
    const held = Math.abs(seconds);

    return sign + String(Math.floor(held / 3600)).padStart(2, "0") +
        ":" + String(Math.floor((held % 3600) / 60)).padStart(2, "0");
}

// go2jsTimeLocalZoneName is the name the zone of the machine answers to, which
// Go names by the offset it keeps rather than by anything the host calls it.
function go2jsTimeLocalZoneName(date) {
    const seconds = -date.getTimezoneOffset() * 60;

    if (seconds === 0) {
        return "UTC";
    }

    const sign = seconds < 0 ? "-" : "+";
    const held = Math.abs(seconds);

    return sign + String(Math.floor(held / 3600)).padStart(2, "0") +
        String(Math.floor((held % 3600) / 60)).padStart(2, "0");
}

function go2jsTimeFormat(date, layout, location) {
    const pad = function(value, width) {
        return String(value).padStart(width, "0");
    };

    // The pieces are what a clock in the location of the moment would show, and
    // the zone of the moment is written out from that same location, so a layout
    // naming a zone names the one the moment is in rather than UTC.
    const local = go2jsTimeLocalDate(date, location);
    const year = local.getUTCFullYear();
    const month = local.getUTCMonth();
    const day = local.getUTCDate();
    const hour = local.getUTCHours();
    const minute = local.getUTCMinutes();
    const second = local.getUTCSeconds();
    const milli = local.getUTCMilliseconds();
    const hour12 = hour % 12 === 0 ? 12 : hour % 12;
    const monthName = go2jsTimeMonths[month];
    const dayName = go2jsTimeDays[local.getUTCDay()];
    const zone = go2jsTimeFormatZone(date, location);

    let out = "";
    let index = 0;

    while (index < layout.length) {
        const rest = layout.slice(index);

        // The pieces are read longest first, so that a name is not read as a
        // shorter name it begins with and a two digit piece is not read as the
        // one digit piece it begins with.
        const pieces = [
            ["January", monthName],
            ["Jan", monthName.slice(0, 3)],
            ["Monday", dayName],
            ["Mon", dayName.slice(0, 3)],
            ["2006", pad(year, 4)],
            ["06", pad(year % 100, 2)],
            ["_2", day < 10 ? " " + day : String(day)],
            ["01", pad(month + 1, 2)],
            ["15", pad(hour, 2)],
            ["03", pad(hour12, 2)],
            ["04", pad(minute, 2)],
            ["05", pad(second, 2)],
            ["02", pad(day, 2)],
            ["PM", hour < 12 ? "AM" : "PM"],
            ["pm", hour < 12 ? "am" : "pm"],
            ["MST", zone.name],
            ["Z07:00", zone.zulu],
            ["Z0700", zone.zulu.replace(":", "")],
            ["-07:00", zone.number],
            ["-0700", zone.number.replace(":", "")],
            ["1", String(month + 1)],
            ["2", String(day)],
            ["3", String(hour12)],
            ["4", String(minute)],
            ["5", String(second)]
        ];

        let matched = null;

        for (const [piece, text] of pieces) {
            if (rest.startsWith(piece)) {
                matched = [piece, text];
                break;
            }
        }

        if (matched !== null) {
            out += matched[1];
            index += matched[0].length;
            continue;
        }

        // A fraction of a second is written as the digits it has. A layout
        // asking for as many of them as the piece carries is answered with that
        // many, zeros included, and one asking for nines is answered with the
        // digits there are, since the nines are what a piece of them means is
        // not written down.
        const fraction = /^\.([0-9]+)/.exec(rest);

        if (fraction !== null) {
            const width = fraction[1].length;
            const trimmed = fraction[1][0] === "9";
            const nanos = date.getUTCMilliseconds() * 1000000 +
                (location !== null && location !== undefined && location.__go2js_nsec !== undefined ? location.__go2js_nsec : 0);
            let digits = String(nanos).padStart(9, "0");

            digits = trimmed ? digits.slice(0, width).replace(/0+$/, "") : digits.slice(0, width);

            out += digits === "" ? "" : "." + digits;
            index += 1 + width;
            continue;
        }

        out += layout[index];
        index += 1;
    }

    return out;
}

function go2jsRegexpNew(pattern) {
    const source = String(pattern);

    // Go validates the pattern when it is compiled, so an invalid expression has
    // to fail here rather than on first use.
    go2jsRegexpNewRegExp(source);

    // Every index method reports Go byte offsets, so the subject is decoded once
    // per call and translated on the way out.
    const exec = (value, flags) => {
        const text = go2jsBytesToString(value);
        const match = go2jsRegexpNewRegExp(source, flags).exec(text);

        return { match: match, offset: go2jsByteOffsetMap(text) };
    };

    return {
        pattern: source,
        MatchString(value) {
            return go2jsRegexpNewRegExp(source).test(String(value));
        },
        Find(value) {
            return go2jsRegexpByteSubmatch(exec(value).match);
        },
        FindString(value) {
            const match = exec(value).match;
            return match === null ? "" : match[0];
        },
        FindStringIndex(value) {
            const result = exec(value);
            return go2jsRegexpIndex(result.match, result.offset);
        },
        FindIndex(value) {
            const result = exec(value);
            return go2jsRegexpIndex(result.match, result.offset);
        },
        FindStringSubmatch(value) {
            return go2jsRegexpSubmatch(exec(value).match);
        },
        FindStringSubmatchIndex(value) {
            const result = exec(value, "d");
            return go2jsRegexpSubmatchIndex(result.match, result.offset);
        },
        FindSubmatch(value) {
            return go2jsRegexpByteSubmatch(exec(value).match);
        },
        FindSubmatchIndex(value) {
            const result = exec(value, "d");
            return go2jsRegexpByteSubmatchIndex(result.match, result.offset);
        },
        FindAllStringSubmatch(value, limit) {
            return go2jsRegexpFindAllSubmatch(source, value, limit);
        },
        FindAllStringSubmatchIndex(value, limit) {
            return go2jsRegexpFindAllSubmatchIndex(source, value, limit);
        },
        FindAllSubmatch(value, limit) {
            return go2jsRegexpFindAllByteSubmatch(source, value, limit);
        },
        FindAllSubmatchIndex(value, limit) {
            return go2jsRegexpFindAllSubmatchIndex(source, value, limit);
        },
        FindAllString(value, limit) {
            return go2jsRegexpFindAllString(source, value, limit);
        },
        FindAllStringIndex(value, limit) {
            return go2jsRegexpFindAllIndex(source, value, limit);
        },
        FindAllIndex(value, limit) {
            return go2jsRegexpFindAllIndex(source, value, limit);
        },
        ReplaceAllString(value, replacement) {
            return go2jsRegexpReplaceAllString(source, value, replacement);
        },
        ReplaceAll(value, replacement) {
            return go2jsBytesToString(value).replace(go2jsRegexpNewRegExp(source, "g"), go2jsRegexpExpand(replacement, source));
        },
        Replace(value, replacement) {
            return go2jsStringify(value).replace(go2jsRegexpNewRegExp(source), go2jsRegexpExpand(replacement, source));
        },
        ReplaceLiteral(value, replacement) {
            return go2jsStringify(value).replace(go2jsRegexpNewRegExp(source), () => go2jsStringify(replacement));
        },
        ReplaceAllLiteral(value, replacement) {
            return go2jsStringify(value).replace(go2jsRegexpNewRegExp(source, "g"), () => go2jsStringify(replacement));
        },
        ReplaceAllStringFunc(value, replacer) {
            return String(value).replace(go2jsRegexpNewRegExp(source, "g"), match => go2jsCallNow(replacer, null, [match]));
        },
        ReplaceAllFunc(value, replacer) {
            return go2jsBytesToString(value).replace(go2jsRegexpNewRegExp(source, "g"), match => go2jsCallNow(replacer, null, [match]));
        },
        ReplaceStringFunc(value, replacer) {
            return String(value).replace(go2jsRegexpNewRegExp(source), match => go2jsCallNow(replacer, null, [match]));
        },
        ReplaceFunc(value, replacer) {
            return go2jsStringify(value).replace(go2jsRegexpNewRegExp(source), match => go2jsCallNow(replacer, null, [match]));
        },
        Split(value, limit) {
            return go2jsRegexpSplit(source, value, limit);
        },
        String() {
            return source;
        },
        NumSubexp() {
            return go2jsRegexpNewRegExp(source + "|").exec("").length - 1;
        },
        SubexpNames() {
            return go2jsRegexpSubmatchNames(source);
        },
        SubexpIndex(name) {
            const wanted = go2jsStringify(name);

            if (wanted === "") {
                return -1;
            }

            const names = go2jsRegexpSubmatchNames(source);

            for (let i = 0; i < names.length; i++) {
                if (names[i] === wanted) {
                    return i;
                }
            }

            return -1;
        }

    };
}

function go2jsRegexpMustCompile(pattern) {
    try {
        return go2jsRegexpNew(pattern);
    } catch (cause) {
        throw new Error("regexp: Compile(" + JSON.stringify(go2jsStringify(pattern)) + "): " + go2jsStringify(cause.message));
    }
}

function go2jsFilepathClean(value) {
    let path = String(value).replace(/\/+/g, "/");

    const absolute = path.startsWith("/");
    const parts = [];

    for (const part of path.split("/")) {
        if (!part || part === ".") {
            continue;
        }

        if (part === "..") {
            if (parts.length > 0 && parts[parts.length - 1] !== "..") {
                parts.pop();
            } else if (!absolute) {
                parts.push("..");
            }
            continue;
        }

        parts.push(part);
    }

    path = parts.join("/");

    if (absolute) {
        path = "/" + path;
    }

    return path || (absolute ? "/" : ".");
}

// The marks of a mode are written in the order Go writes them: the letters of
// what sort of file it is from the highest mark down, and then the nine marks
// of what may be done with it, a mark that is not set being a dash.
const go2jsFileModeLetters = "dalTLDpSugct?";

const go2jsFileModeRights = "rwxrwxrwx";

function go2jsFileModeString(mode) {
	let text = "";

	for (let index = 0; index < go2jsFileModeLetters.length; index++) {
		if ((mode & Math.pow(2, 31 - index)) !== 0) {
			text += go2jsFileModeLetters[index];
		}
	}

	// A file that is only a file has no letter to say what it is, and the place
	// one would take is a dash rather than nothing at all.
	if (text === "") {
		text = "-";
	}

	for (let index = 0; index < go2jsFileModeRights.length; index++) {
		text += (mode & Math.pow(2, 8 - index)) !== 0 ? go2jsFileModeRights[index] : "-";
	}

	return text;
}

// go2jsIsFileMode reports whether a value is a mode of marks rather than a
// number, which is what a file mode is.
function go2jsIsFileMode(value) {
	if (value === null || value === undefined || typeof value !== "object") {
		return false;
	}

	const name = value.__go2js_type_name === undefined ? value.type : value.__go2js_type_name;

	return name === "io/fs.FileMode" || name === "fs.FileMode";
}

// go2jsFileMode reads the bits of a mode off whatever a mode is written as,
// which is a number behind a box or the box itself.
function go2jsFileModeBits(mode) {
	if (mode !== null && mode !== undefined && typeof mode === "object" && mode.__go2js_typed === true) {
		return Number(mode.value) >>> 0;
	}

	return Number(mode) >>> 0;
}

// The marks of a file in the filesystem of the host are in the places the
// filesystem itself keeps them, and Go keeps its own in places of its own, so a
// mode read from one is moved into the marks of the other.
const go2jsFileModeHostTypes = [
	[16384, 2147483648], [40960, 134217728], [4096, 33554432], [49152, 16777216],
	[8192, 34505216], [24576, 67108864], [32768, 0]
];

function go2jsFileModeFromHost(mode) {
	let bits = mode & 511;

	if ((mode & 2048) !== 0) {
		bits |= 8388608;
	}

	if ((mode & 1024) !== 0) {
		bits |= 4194304;
	}

	if ((mode & 512) !== 0) {
		bits |= 1048576;
	}

	for (const [host, own] of go2jsFileModeHostTypes) {
		if ((mode & 61440) === host) {
			bits |= own;
			break;
		}
	}

	return bits;
}

// go2jsFileMode is a mark of a file as its own type rather than as a number,
// which is what lets it write itself out and be asked what it is.
function go2jsFileMode(bits) {
	// The marks of a mode are thirty-two of them and no more, so a mode read
	// through an operator that works in signed ones is read as what it is.
	const mode = Number(bits) >>> 0;

	const boxed = go2jsTyped({
		// A mode is a named number of its own, and a verb that asks what it is
		// says so whether it is handed the mode or what the mode holds.
		type: "fs.FileMode",
		valueOf: function () {
			return mode;
		},
		String: function () {
			return go2jsFileModeString(mode);
		},
		IsDir: function () {
			return (mode & 2147483648) !== 0;
		},
		IsRegular: function () {
			return (mode & 2401763328) === 0;
		},
		Type: function () {
			return go2jsFileMode(mode & 2401763328);
		},
		Perm: function () {
			return go2jsFileMode(mode & 511);
		}
	}, "fs.FileMode", "uint32");

	// The box is what an operator is handed, so the marks behind it are read
	// through the box rather than through what it holds.
	boxed.valueOf = function () {
		return mode;
	};

	return boxed;
}

function go2jsFilepathJoin(...parts) {
    return go2jsFilepathClean(
        parts
            .filter(part => part !== null && part !== undefined && String(part) !== "")
            .map(part => String(part))
            .join("/")
    );
}

function go2jsFilepathBase(value) {
    const path = go2jsFilepathClean(value);
    if (path === "/" || path === ".") {
        return path === "/" ? "/" : ".";
    }

    const index = path.lastIndexOf("/");
    return index === -1 ? path : path.slice(index + 1);
}

function go2jsFilepathDir(value) {
    const path = go2jsFilepathClean(value);

    if (path === "/") {
        return "/";
    }

    const index = path.lastIndexOf("/");

    if (index === -1) {
        return ".";
    }

    if (index === 0) {
        return "/";
    }

    return path.slice(0, index);
}

function go2jsFilepathExt(value) {
    const base = go2jsFilepathBase(value);
    const index = base.lastIndexOf(".");

    if (index <= 0) {
        return "";
    }

    return base.slice(index);
}

// A URL is written out of the parts it is made of, so the parts are kept as
// they were given and the whole is assembled from them, the way the language
// writes a URL from a scheme, a user, a host, a path, a query and a fragment.
function go2jsURL() {
	const url = {
		Scheme: "",
		Opaque: "",
		User: null,
		Host: "",
		Path: "",
		RawPath: "",
		ForceQuery: false,
		RawQuery: "",
		Fragment: "",
		RawFragment: ""
	};

	url.IsAbs = function() {
		return url.Scheme !== "";
	};

	url.EscapedPath = function() {
		return go2jsURLEscapedPath(url);
	};

	url.Hostname = function() {
		// An address written in full keeps its brackets around the host in a
		// URL, and a name is asked for without them.
		const name = go2jsURLHostPort(url.Host)[0];

		return name.startsWith("[") && name.endsWith("]") ? name.slice(1, -1) : name;
	};

	url.Port = function() {
		return go2jsURLHostPort(url.Host)[1];
	};

	url.Query = function() {
		return go2jsURLQueryFromRaw(url.RawQuery, null);
	};

	url.String = function() {
		return go2jsURLString(url);
	};

	// A password is a secret, so a URL written for reading shows where one was
	// without saying what it was.
	url.Redacted = function() {
		const shown = go2jsURLCopy(url);

		if (shown.User !== null && shown.User.Password() !== "") {
			shown.User = go2jsURLUserinfo(shown.User.Username(), "xxxxx", true);
		}

		return go2jsURLString(shown);
	};

	url.JoinPath = function(...elems) {
		const joined = go2jsURL();
		const parts = [go2jsURLEscapedPath(url)].concat(elems.map((part) => go2jsStringify(part)));
		let path = parts[0];

		if (!path.startsWith("/")) {
			path = "/" + path;
		}

		// The elements are joined and then cleaned, so a step that walks back
		// out of a directory takes the directory with it.
		const kept = [];

		for (const segment of path.split("/")) {
			if (segment === "" || segment === ".") {
				continue;
			}

			if (segment === "..") {
				kept.pop();
				continue;
			}

			kept.push(segment);
		}

		path = "/" + kept.join("/");

		// A path written to end at a separator keeps it, because that is where
		// the last element was added.
		if (go2jsStringify(parts[parts.length - 1]).endsWith("/") && !path.endsWith("/")) {
			path += "/";
		}

		for (let index = 1; index < parts.length; index++) {
			const element = go2jsStringify(parts[index]);

			if (element === "") {
				continue;
			}

			path = path.replace(/\/$/, "") + "/" + element;
		}

		joined.Scheme = url.Scheme;
		joined.Opaque = url.Opaque;
		joined.User = url.User;
		joined.Host = url.Host;
		joined.RawQuery = url.RawQuery;
		joined.ForceQuery = url.ForceQuery;
		joined.Fragment = url.Fragment;
		joined.RawFragment = url.RawFragment;

		const [escaped, decoded] = go2jsURLSetPath(joined, path);

		joined.Path = decoded;
		joined.RawPath = escaped;

		return joined;
	};

	return url;
}

function go2jsURLCopy(url) {
	const copy = go2jsURL();

	for (const field of ["Scheme", "Opaque", "User", "Host", "Path", "RawPath",
		"ForceQuery", "RawQuery", "Fragment", "RawFragment"]) {
		copy[field] = url[field];
	}

	return copy;
}

// go2jsURLShouldEscape reports whether a character has to be written as the
// bytes of its name, which depends on where in the URL it stands: a path keeps
// the marks a path is written with, a query keeps none of them, and a host
// keeps the ones a host is allowed to hold.
function go2jsURLShouldEscape(char, mode) {
	if (/[A-Za-z0-9]/.test(char)) {
		return false;
	}

	if (char === "-" || char === "_" || char === "." || char === "~") {
		return false;
	}

	if (mode === "host" || mode === "zone") {
		return "!$&'()*+,;=:[]<>\"".indexOf(char) === -1;
	}

	if ("$&+,/:;=?@".indexOf(char) !== -1) {
		if (mode === "path") {
			return char === "?";
		}

		if (mode === "segment") {
			return char === "/" || char === ";" || char === "," || char === "?";
		}

		if (mode === "user") {
			return char === "@" || char === "/" || char === "?" || char === ":";
		}

		return mode === "query";
	}

	if (mode === "fragment") {
		return char !== "!" && char !== "(" && char !== ")" && char !== "*";
	}

	return true;
}

function go2jsURLEncode(value, mode) {
	const text = go2jsStringify(value);
	let out = "";

	// A character outside the basic plane stands for several bytes, and each of
	// them is written by name.
	for (const char of text) {
		if (mode === "query" && char === " ") {
			out += "+";
			continue;
		}

		if (!go2jsURLShouldEscape(char, mode)) {
			out += char;
			continue;
		}

		for (const byte of new TextEncoder().encode(char)) {
			out += "%" + byte.toString(16).toUpperCase().padStart(2, "0");
		}
	}

	return out;
}

// go2jsURLDecode undoes the writing by name. The bytes are gathered first and
// read as text at the end, because a name written one byte at a time is a
// character rather than the bytes of one.
function go2jsURLDecode(value, mode) {
	const text = go2jsStringify(value);
	const bytes = [];
	let index = 0;

	while (index < text.length) {
		const char = text[index];

		if (char === "%") {
			const named = text.slice(index + 1, index + 3);

			if (named.length < 2 || !/^[0-9A-Fa-f]{2}$/.test(named)) {
				return ["", go2jsNameError(
					new Error("invalid URL escape " + JSON.stringify(text.slice(index, index + 3))),
					"url.EscapeError")];
			}

			bytes.push(parseInt(named, 16));
			index += 3;
			continue;
		}

		// A space is written as a plus in a query, because that is how a query is
		// written, and nowhere else.
		if (char === "+" && mode === "query") {
			bytes.push(0x20);
			index++;
			continue;
		}

		for (const byte of new TextEncoder().encode(char)) {
			bytes.push(byte);
		}

		index++;
	}

	return [new TextDecoder().decode(new Uint8Array(bytes)), null];
}

function go2jsURLQueryEscape(value) {
	return go2jsURLEncode(value, "query");
}

function go2jsURLFragmentEscape(value) {
	return go2jsURLEncode(value, "fragment");
}

function go2jsURLQueryUnescape(value) {
	return go2jsURLDecode(value, "query");
}

function go2jsURLEscapedPath(url) {
	if (url.RawPath !== "" && go2jsURLValidEncoded(url.RawPath, "path")) {
		const [path, err] = go2jsURLDecode(url.RawPath, "path");

		if (err === null && path === url.Path) {
			return url.RawPath;
		}
	}

	if (url.Path === "*") {
		return "*";
	}

	return go2jsURLEncode(url.Path, "path");
}

// go2jsURLValidEncoded reports whether text is written the way the mode writes
// it, which is what tells a path that was written on purpose from one that was
// written by escaping a path.
function go2jsURLValidEncoded(text, mode) {
	for (const char of text) {
		if (go2jsURLShouldEscape(char, mode)) {
			return false;
		}
	}

	return true;
}

function go2jsURLString(url) {
	let out = "";

	if (url.Scheme !== "") {
		out += url.Scheme + ":";
	}

	if (url.Opaque !== "") {
		out += url.Opaque;
	} else {
		if (url.Scheme !== "" || url.Host !== "" || url.User !== null) {
			if (url.Host !== "" || url.Path !== "" || url.User !== null) {
				out += "//";
			}

			if (url.User !== null && url.User !== undefined) {
				out += go2jsURLUserinfoString(url.User) + "@";
			}

			if (url.Host !== "") {
				out += go2jsURLEncode(url.Host, "host");
			}
		}

		const path = go2jsURLEscapedPath(url);

		// A host is followed by a path that starts at a separator, and a path
		// whose first segment names a scheme is told from a file name by the
		// directory in front of it.
		if (path !== "" && path[0] !== "/" && url.Host !== "") {
			out += "/";
		}

		if (out === "" && path.split("/")[0].indexOf(":") !== -1) {
			out += "./";
		}

		out += path;
	}

	if (url.ForceQuery === true || url.RawQuery !== "") {
		out += "?" + url.RawQuery;
	}

	if (url.Fragment !== "") {
		out += "#" + (url.RawFragment !== "" ? url.RawFragment : go2jsURLFragmentEscape(url.Fragment));
	}

	return out;
}

// go2jsURLHostPort splits a host into the name and the port, keeping the
// brackets of an address that is written in full around the name.
function go2jsURLHostPort(host) {
	const text = go2jsStringify(host);

	if (text.startsWith("[")) {
		const close = text.indexOf("]");

		if (close === -1) {
			return [text, ""];
		}

		const name = text.slice(0, close + 1);
		const rest = text.slice(close + 1);

		return rest.startsWith(":") ? [name, rest.slice(1)] : [name, ""];
	}

	const colon = text.lastIndexOf(":");

	if (colon === -1) {
		return [text, ""];
	}

	return [text.slice(0, colon), text.slice(colon + 1)];
}

function go2jsURLInvalidHost(host) {
	const text = go2jsStringify(host);
	const [, port] = go2jsURLHostPort(text);

	if (port !== "" && !/^[0-9]*$/.test(port)) {
		return go2jsNameError(new Error("invalid port " + JSON.stringify(port === "" ? ":" : ":" + port) + " after host"),
			"*url.Error");
	}

	if (text.indexOf("[") !== -1 || text.indexOf("]") !== -1) {
		if (!/^\[[0-9A-Fa-f:.]+\]/.test(text)) {
			return go2jsNameError(new Error("invalid character in host name"), "*url.Error");
		}
	}

	return null;
}

function go2jsURLParse(value) {
	const text = go2jsStringify(value);
	const url = go2jsURL();

	if (text === "") {
		return [url, null];
	}

	let rest = text;
	const hash = rest.indexOf("#");

	if (hash !== -1) {
		const written = rest.slice(hash + 1);

		rest = rest.slice(0, hash);

		const [fragment, fragmentErr] = go2jsURLDecode(written, "fragment");

		if (fragmentErr !== null) {
			return [null, go2jsURLError("parse", text, fragmentErr)];
		}

		url.Fragment = fragment;
		url.RawFragment = go2jsURLFragmentEscape(fragment) === written ? "" : written;
	}

	// A scheme is the name in front of the first colon, as long as the colon
	// comes before any separator and the name is one a scheme may be spelled
	// with. A name in capitals is a path rather than a scheme, because schemes
	// are written in small letters.
	const colon = rest.indexOf(":");
	const separator = rest.search(/[/?#]/);
	let scheme = "";

	if (colon !== -1 && (separator === -1 || colon < separator)) {
		const candidate = rest.slice(0, colon);

		if (/^[A-Za-z][A-Za-z0-9+\-.]*$/.test(candidate) && candidate === candidate.toLowerCase()) {
			scheme = candidate;
			rest = rest.slice(colon + 1);
		}
	}

	url.Scheme = scheme;

	if (rest.startsWith("//")) {
		rest = rest.slice(2);

		let end = rest.length;

		for (let index = 0; index < rest.length; index++) {
			if (rest[index] === "/" || rest[index] === "?") {
				end = index;
				break;
			}
		}

		const authority = rest.slice(0, end);

		rest = rest.slice(end);

		// The user is written before the last at sign, because a host may hold
		// an at sign of its own inside a name.
		const at = authority.lastIndexOf("@");

		if (at !== -1) {
			url.User = go2jsURLParseUserinfo(authority.slice(0, at));
			url.Host = authority.slice(at + 1);
		} else {
			url.Host = authority;
		}

		const invalid = go2jsURLInvalidHost(url.Host);

		if (invalid !== null) {
			return [null, go2jsURLError("parse", text, invalid)];
		}
	} else if (!rest.startsWith("/")) {
		if (scheme !== "") {
			// What stands behind a scheme with no separator before it says what
			// the URL is about rather than where it leads, so it is kept whole.
			url.Opaque = rest;

			return [url, null];
		}

		if (rest.split("/")[0].indexOf(":") !== -1) {
			return [null, go2jsURLError("parse", text,
				go2jsNameError(new Error("first path segment in URL cannot contain colon"), "*url.Error"))];
		}
	}

	const question = rest.indexOf("?");

	if (question !== -1) {
		url.ForceQuery = true;
		url.RawQuery = rest.slice(question + 1);
		rest = rest.slice(0, question);
	}

	const [escaped, path, pathErr] = go2jsURLPathParts(rest);

	if (pathErr !== null) {
		return [null, go2jsURLError("parse", text, pathErr)];
	}

	url.RawPath = escaped;
	url.Path = path;

	return [url, null];
}

// go2jsURLPathParts reads a path as it was written and as it reads, and reports
// which of the two to keep: the writing is kept when it says something the
// escaping would not have said on its own.
function go2jsURLPathParts(raw) {
	if (raw === "") {
		return ["", "", null];
	}

	const [path, err] = go2jsURLDecode(raw, "path");

	if (err !== null) {
		return ["", "", err];
	}

	return [go2jsURLEncode(path, "path") === raw ? "" : raw, path, null];
}

function go2jsURLSetPath(url, raw) {
	const [escaped, path, err] = go2jsURLPathParts(raw);

	if (err !== null) {
		throw err;
	}

	return [escaped, path];
}

function go2jsURLError(op, url, err) {
	return go2jsNameError(new Error(op + " " + JSON.stringify(url) + ": " + go2jsStringify(err && err.message ? err.message : err)), "*url.Error");
}

function go2jsURLParseUserinfo(text) {
	const colon = text.indexOf(":");

	if (colon === -1) {
		return go2jsURLUserinfo(text, "", false);
	}

	return go2jsURLUserinfo(text.slice(0, colon), text.slice(colon + 1), true);
}

function go2jsURLUserinfo(name, password, hasPassword) {
	const info = {
		name: name,
		password: password,
		hasPassword: hasPassword
	};

	// The name and the password are held as they were written and are given
	// back as they read, which is the other way round for a URL.
	info.Username = function() {
		const [decoded, err] = go2jsURLDecode(name, "user");

		return err === null ? decoded : name;
	};

	info.Password = function() {
		if (!hasPassword) {
			return "";
		}

		const [decoded, err] = go2jsURLDecode(password, "user");

		return err === null ? decoded : password;
	};

	info.String = function() {
		let out = go2jsURLEncode(name, "user");

		if (hasPassword) {
			out += ":" + go2jsURLEncode(password, "user");
		}

		return out;
	};

	return info;
}

function go2jsURLUserinfoString(info) {
	if (info === null || info === undefined) {
		return "";
	}

	return typeof info.String === "function" ? info.String() : go2jsStringify(info);
}

// go2jsURLQueryFromRaw reads a query into a set of values, and leaves out any
// pair that does not read, the way a query is read: a pair that cannot be read
// is no pair at all.
function go2jsURLQueryFromRaw(raw, fail) {
	const values = go2jsURLValues();
	const text = go2jsStringify(raw);

	if (text === "") {
		return values;
	}

	for (const pair of text.split("&")) {
		if (pair === "") {
			continue;
		}

		// A semicolon separated a pair once, and is refused now rather than
		// taken for a separator.
		if (pair.indexOf(";") !== -1) {
			if (typeof fail === "function") {
				fail("invalid semicolon separator in query");
			}

			continue;
		}

		const equals = pair.indexOf("=");
		const rawKey = equals === -1 ? pair : pair.slice(0, equals);
		const rawValue = equals === -1 ? "" : pair.slice(equals + 1);
		const [key, keyErr] = go2jsURLQueryUnescape(rawKey);
		const [value, valueErr] = go2jsURLQueryUnescape(rawValue);

		if (keyErr !== null || valueErr !== null) {
			if (typeof fail === "function") {
				fail(keyErr !== null ? keyErr.message : valueErr.message);
			}

			continue;
		}

		go2jsURLValuesAdd(values, key, value);
	}

	return values;
}

function go2jsURLValues() {
	return go2jsMapTyped("net/url.Values", go2jsMap([]));
}

function go2jsURLValuesGet(values, key) {
	const list = go2jsMapGet(values, go2jsStringify(key), null);

	return Array.isArray(list) && list.length > 0 ? go2jsStringify(list[0]) : "";
}

function go2jsURLValuesSet(values, key, value) {
	go2jsMapSet(values, go2jsStringify(key), [go2jsStringify(value)]);
}

function go2jsURLValuesAdd(values, key, value) {
	const name = go2jsStringify(key);
	const list = go2jsMapGet(values, name, null);

	if (!Array.isArray(list)) {
		go2jsMapSet(values, name, [go2jsStringify(value)]);
		return;
	}

	list.push(go2jsStringify(value));
}

function go2jsURLValuesDel(values, key) {
	go2jsMapDelete(values, go2jsStringify(key));
}

function go2jsURLValuesHas(values, key) {
	return go2jsMapHas(values, go2jsStringify(key));
}

// A query is written with its keys in order, because a query that reads the
// same has to be written the same way.
function go2jsURLValuesEncode(values) {
	const keys = go2jsMapKeys(values).map((key) => go2jsStringify(key)).sort();
	let out = "";

	for (const key of keys) {
		const list = go2jsMapGet(values, key, []);
		const pairs = Array.isArray(list) ? list : [list];

		for (const value of pairs) {
			if (out !== "") {
				out += "&";
			}

			out += go2jsURLQueryEscape(key) + "=" + go2jsURLQueryEscape(value);
		}
	}

	return out;
}

function go2jsJSONMarshal(value, fields, omitEmpty, stringFields, textFields, prefix, indent) {
    try {
        const encoded = go2jsJSONEncode(value, fields, omitEmpty, stringFields, textFields, prefix, indent);

        return [go2jsJSONText(encoded, prefix, indent, true), null];
    } catch (err) {
        return [null, err];
    }
}

// go2jsJSONEscapeHTML writes the five characters that make JSON unsafe to
// embed in HTML as the escapes encoding/json writes for them. JSON.stringify
// leaves every one of them alone, so they are swapped for their escapes once
// the text is written, which never puts one outside a string.
function go2jsJSONEscapeHTML(text) {
    if (typeof text !== "string") {
        return text;
    }

    return text
        .split("<").join("\\u003c")
        .split(">").join("\\u003e")
        .split("&").join("\\u0026")
        .split("\u2028").join("\\u2028")
        .split("\u2029").join("\\u2029");
}

// go2jsJSONText writes an encoded value out. An indent of nothing writes the
// value with no room between tokens at all, while any other indent starts a
// new line for every part and pads it, putting the prefix in front of every
// line but the first.
function go2jsJSONText(encoded, prefix, indent, escapeHTML) {
    let text;

    if (prefix === "" && indent === "") {
        text = JSON.stringify(encoded);
    } else {
        text = JSON.stringify(encoded, null, indent === "" ? " " : indent);

        if (indent === "") {
            text = text.split("\n").map(line => line.replace(/^ +/, "")).join("\n");
        }
    }

    // the text of a value that was written rather than made goes back in before
    // the escapes are written, since a number and a raw message both say more
    // than the double and the value made again out of what was read would
    if (go2jsJSONLiteralWritten) {
        text = go2jsJSONPutLiterals(text);
    }

    // the prefix is not part of the JSON, so it is added after the escapes
    text = escapeHTML === false ? text : go2jsJSONEscapeHTML(text);

    if (prefix !== "" && indent !== "") {
        const lines = text.split("\n");

        for (let index = 1; index < lines.length; index++) {
            lines[index] = prefix + lines[index];
        }

        text = lines.join("\n");
    }

    return text;
}

function go2jsJSONEncode(value, fields, omitEmpty, stringFields, textFields, prefix, indent) {
    if (value === null || value === undefined) {
        return null;
    }

    if (Array.isArray(value)) {
        return value.map(item => go2jsJSONEncode(item, null, null, null, null, prefix, indent));
    }

    if (value instanceof go2jsNativeMap) {
        const out = {};
        for (const [key, item] of value.entries()) {
            out[go2jsStringify(key)] = go2jsJSONEncode(item, null, null, null, null, prefix, indent);
        }
        return out;
    }

    if (typeof value !== "object" || value instanceof Error) {
        return value;
    }

    // a value that stands for a struct is read for the fields it was declared
    // with, which is how a struct nested in another is written with the names
    // its own tags gave it rather than with the names of its fields
    if (fields === null || fields === undefined) {
        const own = go2jsJSONStructMapping(value);

        if (own !== null) {
            fields = own.fields;
            omitEmpty = own.omitEmpty;
            stringFields = own.stringFields;
            textFields = own.textFields;
        }
    }

    if (fields) {
        const out = {};

        for (const field of Object.keys(fields)) {
            if (!(field in value)) {
                continue;
            }

            const name = fields[field];
            const item = value[field];

            if (omitEmpty && omitEmpty.indexOf(name) >= 0 && go2jsJSONEmpty(item)) {
                continue;
            }

            // a field a tag marked ",string" is written as the text of the
            // value rather than as the value itself
            if (stringFields && stringFields.indexOf(field) >= 0) {
                out[name] = JSON.stringify(item);
                continue;
            }

            // a field declared to hold text rather than a value is written as
            // the text it holds, since writing it as a value would write it in
            // quotes and make of it a string holding that text
            if (textFields !== null && textFields !== undefined && textFields[field] !== undefined) {
                out[name] = go2jsJSONTextValue(item, textFields[field], prefix, indent);
                continue;
            }

            out[name] = go2jsJSONEncode(item, null, null, null, null);
        }

        return out;
    }

    const out = {};

    for (const key of Object.keys(value)) {
        out[key] = go2jsJSONEncode(value[key], null, null, null, null, prefix, indent);
    }

    return out;
}

// go2jsJSONFold is a name as JSON is said to match it, which is a name that
// differs from another only in case and so is that other name.
function go2jsJSONFold(name) {
    let out = "";

    for (let i = 0; i < name.length; i++) {
        let code = name.charCodeAt(i);

        if (code >= 65 && code <= 90) {
            code = code + 32;
        }

        out += String.fromCharCode(code);
    }

    return out;
}

// go2jsJSONStructMapping is the fields a struct value is written with: the name
// each field was declared under and the name its tag gave it, along with the
// fields a tag asked to leave out when empty. It is nothing for a value that
// does not stand for a struct.
function go2jsJSONStructMapping(value) {
    const candidates = go2jsJSONFieldCandidates(value, 0, new Set());

    if (candidates === null) {
        return null;
    }

    return go2jsJSONReduceCandidates(candidates);
}

// go2jsJSONMappingTypes is what each field of a struct value holds, read off the
// candidates its own fields make up and so flattened and settled by the same
// rules as its names. A field holding something JSON has no shape for is left
// out, and reading into it falls back to whatever the text held.
function go2jsJSONMappingTypes(value) {
    const candidates = go2jsJSONFieldCandidates(value, 0, new Set());

    if (candidates === null) {
        return null;
    }

    const fields = go2jsJSONReduceCandidates(candidates);

    if (fields === null || fields === undefined || fields.fields === undefined) {
        return null;
    }

    const types = {};

    for (const candidate of candidates) {
        if (fields.fields[candidate.name] === undefined || types[candidate.name] !== undefined) {
            continue;
        }

        const descriptor = go2jsJSONTypeDescriptor(candidate.type);

        if (descriptor === null || descriptor === undefined || descriptor.kind === "any") {
            continue;
        }

        types[candidate.name] = descriptor;
    }

    return types;
}

// go2jsJSONTypeDescriptor is what a destination of the named type holds, read
// off the type name a program registered the field with. A type that stands for
// a struct is described by the name it was declared under and by nothing else,
// because a value of it is built from that very type and read through that
// type's own fields, which is as much as any depth of it needs to know.
function go2jsJSONTypeDescriptor(typeName) {
    if (typeof typeName !== "string" || typeName === "") {
        return null;
    }

    const text = typeName;

    if (text[0] === "*") {
        return {kind: "ptr", elem: go2jsJSONTypeDescriptor(text.slice(1))};
    }

    if (text.startsWith("map[")) {
        const close = text.indexOf("]");

        if (close < 0) {
            return null;
        }

        return {kind: "map", elem: go2jsJSONTypeDescriptor(text.slice(close + 1))};
    }

    if (text[0] === "[" && text[1] !== "]") {
        const close = text.indexOf("]");

        if (close < 0) {
            return null;
        }

        const length = parseInt(text.slice(1, close), 10);

        return {kind: "array", len: isNaN(length) ? 0 : length, elem: go2jsJSONTypeDescriptor(text.slice(close + 1))};
    }

    if (text.startsWith("[]")) {
        return {kind: "slice", elem: go2jsJSONTypeDescriptor(text.slice(2))};
    }

    if (text === "string" || text === "[]byte" || text.startsWith("string")) {
        return {kind: "string"};
    }

    if (text === "bool") {
        return {kind: "bool"};
    }

    if (text === "int" || text.startsWith("int") || text === "uint" || text.startsWith("uint") ||
        text === "byte" || text === "rune" || text.startsWith("float") ||
        text.startsWith("complex")) {
        return {kind: "number"};
    }

    if (text === "any" || text.startsWith("interface{")) {
        return {kind: "any"};
    }

    if (Array.isArray(go2jsStructFields[text])) {
        return {kind: "lazy", name: text};
    }

    return {kind: "any"};
}

// go2jsJSONFieldCandidates reads the fields a struct value is written with at
// the depth they were found. A field promoted out of an embedded struct is one
// level deeper than a field declared on the value itself, and an embedded
// struct a tag gave no name is walked into rather than written under its own.
function go2jsJSONFieldCandidates(value, depth, seen) {
    if (value === null || typeof value !== "object") {
        return null;
    }

    const name = go2jsTypeNamesByConstructor.get(value.constructor);

    if (name === undefined || go2jsTypeKinds[name] !== "struct") {
        return null;
    }

    const descriptors = go2jsStructFields[name];

    if (!Array.isArray(descriptors)) {
        return null;
    }

    // a struct that reaches itself through an embedded pointer is walked once,
    // which is as much as the shallower fields it repeats are worth
    if (seen.has(value)) {
        return null;
    }

    seen.add(value);

    const candidates = [];

    for (const descriptor of descriptors) {
        // a field the package did not export is not written, which a package
        // path on it is the mark of
        if (descriptor.PkgPath !== "" && descriptor.Anonymous !== true) {
            continue;
        }

        const tag = go2jsJSONTag(descriptor.Tag, descriptor.Name);

        if (tag.skip) {
            continue;
        }

        if (descriptor.Anonymous === true && tag.named === false) {
            const inner = go2jsJSONFieldCandidates(value[descriptor.Name], depth + 1, seen);

            if (inner !== null) {
                seen.delete(value[descriptor.Name]);

                for (const candidate of inner) {
                    candidates.push(candidate);
                }

                continue;
            }
        }

        candidates.push({
            name: descriptor.Name,
            json: tag.name,
            depth: depth,
            omit: tag.omit,
            string: tag.string,
            type: descriptor.Type,
        });
    }

    return candidates;
}

// go2jsJSONReduceCandidates settles the name a field is written under: a
// shallower field hides a deeper one, and a name two fields at the same depth
// disagree over is left out.
function go2jsJSONReduceCandidates(candidates) {
    const minDepth = {};
    const count = {};

    for (const candidate of candidates) {
        if (minDepth[candidate.json] === undefined || candidate.depth < minDepth[candidate.json]) {
            minDepth[candidate.json] = candidate.depth;
        }
    }

    for (const candidate of candidates) {
        if (candidate.depth === minDepth[candidate.json]) {
            count[candidate.json] = (count[candidate.json] || 0) + 1;
        }
    }

    const fields = {};
    const omitEmpty = [];
    const stringFields = [];
    const textFields = [];

    for (const candidate of candidates) {
        if (candidate.depth !== minDepth[candidate.json] || count[candidate.json] !== 1) {
            continue;
        }

        if (fields[candidate.name] !== undefined) {
            continue;
        }

        fields[candidate.name] = candidate.json;

        if (candidate.omit) {
            omitEmpty.push(candidate.json);
        }

        const written = go2jsJSONTextTypeName(candidate.type);

        if (written !== "") {
            textFields[candidate.name] = written;
        }

        if (candidate.string) {
            stringFields.push(candidate.name);
        }
    }

    return {
        fields: fields,
        omitEmpty: omitEmpty,
        stringFields: stringFields,
        textFields: textFields,
    };
}

// go2jsJSONTag reads the name and the options a struct tag gave a field. A tag
// that names nothing, or that names the field as left out, is not one to write
// the field with.
function go2jsJSONTag(tag, fallback) {
    const match = /(?:^|\s)json:"([^"]*)"/.exec(tag === null || tag === undefined ? "" : tag);

    if (match === null) {
        return {name: fallback, named: false, omit: false, string: false, skip: false};
    }

    const parts = match[1].split(",");
    const name = parts[0].trim();
    const omit = parts.slice(1).some((option) => option.trim() === "omitempty");
    const asString = parts.slice(1).some((option) => option.trim() === "string");

    if (name === "-") {
        return {name: fallback, named: true, omit: omit, string: asString, skip: true};
    }

    return {name: name === "" ? fallback : name, named: name !== "", omit: omit, string: asString, skip: false};
}

function go2jsJSONEmpty(value) {
    if (value === null || value === undefined) {
        return true;
    }

    if (typeof value === "boolean") {
        return !value;
    }

    if (typeof value === "number") {
        return value === 0;
    }

    if (typeof value === "string") {
        return value === "";
    }

    if (value !== null && value !== undefined && value.__go2js_pointer === true) {
        return go2jsJSONEmpty(value[go2jsPointerGet]());
    }

    if (Array.isArray(value) || typeof value === "object") {
        return go2jsLen(value) === 0;
    }

    return false;
}

function go2jsJSONUnmarshal(data, target, fields, destination, stringFields, fieldTypes) {
    go2jsJSONFirstError = null;

    try {
        const text = typeof data === "string" ? data : new TextDecoder().decode(Uint8Array.from(data));
        const value = JSON.parse(text);

        if (target === null || target === undefined) {
            return null;
        }

        if (destination !== null && destination !== undefined && destination.kind !== undefined) {
            go2jsJSONAssign(target, go2jsJSONStore(value, target, destination, text, ""));
            return go2jsJSONFirstError;
        }

        if (typeof target === "object") {
            if (value !== null && typeof value === "object") {
                go2jsJSONDecode(value, target, fields, stringFields, fieldTypes, text, "");
            }
            return go2jsJSONFirstError;
        }

        return null;
    } catch (err) {
        return err;
    }
}

// go2jsJSONTextTypeName names the two encoding/json types whose value is the text
// of something rather than something JSON describes: a RawMessage holds the text
// of a value that has not been read yet, and a Number holds the text of a
// number. Anything else is not one of them, and is written as the value it is.
function go2jsJSONTextTypeName(typeName) {
    if (typeName === "json.RawMessage") {
        return "json.RawMessage";
    }

    if (typeName === "json.Number") {
        return "json.Number";
    }

    return "";
}

// go2jsJSONTextValue is a value written as the text it was written with rather
// than as what that text describes. JSON.stringify writes every number it is
// given as a double, which is the same number unless the number has more digits
// than a double keeps or more digits than it needs, and it cannot write a raw
// message at all because the text of one is not a value. Both are therefore
// written under a name of their own, and that name is put back to the text it
// stood for once the value around it has been written out.
let go2jsJSONLiteralCount = 0;

const go2jsJSONLiteralTexts = {};

let go2jsJSONLiteralWritten = false;

function go2jsJSONLiteral(text, prefix, indent) {
    const name = "@@go2js-json-literal-" + go2jsJSONLiteralCount++ + "@@";

    go2jsJSONLiteralTexts[name] = {text: text, prefix: prefix, indent: indent};
    go2jsJSONLiteralWritten = true;

    return name;
}

// go2jsJSONTextValue is the text a field declared to hold text is written as. A
// raw message is the text of a value that has not been read yet, so it is
// written out as the value that text describes and laid out with whatever the
// writer around it is laid out with; a number is the digits of a number and is
// written as those digits, which is the only way a number no double keeps the
// digits of is kept.
function go2jsJSONTextValue(value, typeName, prefix, indent) {
    const text = go2jsJSONNumberText(value);

    if (typeName === "json.Number") {
        // A number that holds no digits is not a number at all, and writing one
        // is what the writer says is wrong rather than what it writes.
        if (!go2jsJSONNumberLiteral(text)) {
            throw new Error("json: invalid number literal " + go2jsJSONWritten(text));
        }

        return go2jsJSONLiteral(text, prefix, indent);
    }

    // a raw message holding nothing is nothing, which is what the writer writes
    // for a value it has nothing of
    if (text === "") {
        return null;
    }

    try {
        return go2jsJSONLiteral(go2jsJSONCompact(text), prefix, indent);
    } catch (err) {
        return text;
    }
}

// go2jsJSONPutLiterals puts back the text of every value that was written rather
// than made, which JSON.stringify wrote out under a name of its own. A value
// that runs to more than one line is lined up under the column its name was
// written at, since that is the column its text begins at.
function go2jsJSONPutLiterals(text) {
    go2jsJSONLiteralWritten = false;

    return text.replace(/"@@go2js-json-literal-(\d+)@@"/g, (match, which, offset) => {
        const name = "@@go2js-json-literal-" + which + "@@";

        if (!Object.prototype.hasOwnProperty.call(go2jsJSONLiteralTexts, name)) {
            return match;
        }

        const literal = go2jsJSONLiteralTexts[name];

        delete go2jsJSONLiteralTexts[name];

        if (literal.prefix === "" && literal.indent === "") {
            return literal.text;
        }

        const laid = go2jsJSONIndent(literal.text, "", literal.indent);
        const lineStart = text.lastIndexOf("\n", offset) + 1;
        const leading = /^[ \t]*/.exec(text.slice(lineStart, offset))[0];

        return laid.split("\n").join("\n" + leading);
    });
}

// go2jsJSONCompact is the text of a value with the room between its tokens taken
// out, which is what a raw message is written out as. Only the room between
// tokens goes: a string is kept whole, so the escapes and the spacing written
// inside one survive, and a number is kept whole, so the digits written for one
// survive.
function go2jsJSONCompact(text) {
    let out = "";
    let index = 0;

    while (index < text.length) {
        const character = text[index];

        if (character === '"') {
            const end = go2jsJSONSkipValue(text, index);

            out += text.slice(index, end);
            index = end;
            continue;
        }

        if (character === " " || character === "\t" || character === "\n" || character === "\r") {
            index++;
            continue;
        }

        out += character;
        index++;
    }

    return out;
}

// go2jsJSONIndent lays the text of a value out with room between its tokens: a
// token that opens an object or an array, and one that opens a member or an
// element, each begins a line of its own, and a closing one ends the last. Every
// token is written as it was written, so the digits of a number and the escapes
// of a string are left exactly as the text held them.
function go2jsJSONIndent(text, prefix, indent) {
    let out = "";
    let depth = 0;
    let index = 0;

    const line = () => {
        out += "\n" + prefix + indent.repeat(depth);
    };

    while (index < text.length) {
        const character = text[index];

        if (character === " " || character === "\t" || character === "\n" || character === "\r") {
            index++;
            continue;
        }

        if (character === '"') {
            const end = go2jsJSONSkipValue(text, index);

            out += text.slice(index, end);
            index = end;
            continue;
        }

        if (character === ",") {
            out += ",";
            index++;
            line();
            continue;
        }

        if (character === ":") {
            out += ": ";
            index++;
            continue;
        }

        if (character === "{" || character === "[") {
            const end = go2jsJSONSkipValue(text, index);

            // an object and an array holding nothing are written as the two
            // characters that open and close one, with no line between them
            if (end - index <= 2) {
                out += text.slice(index, end);
                index = end;
                continue;
            }

            depth++;
            out += character;
            index++;
            line();
            continue;
        }

        if (character === "}" || character === "]") {
            depth--;
            line();
            out += character;
            index++;
            continue;
        }

        const end = go2jsJSONSkipValue(text, index);

        out += text.slice(index, end);
        index = end;
    }

    return out;
}

// go2jsJSONWritten is the text of a value made out of the value itself, for a
// value whose own text the document could not be asked for. It is the same
// value written out again, which is what a program is given when it asked to
// hold something as text and the text cannot be read, and which is why it is
// the last resort and not the first.
function go2jsJSONWritten(value) {
    return JSON.stringify(value);
}

// go2jsJSONNumberLiteral reports whether the text is a number as JSON writes one,
// which is what a number read into a json.Number is required to have been
// written as.
function go2jsJSONNumberLiteral(text) {
    if (typeof text !== "string" || text === "") {
        return false;
    }

    let index = 0;

    if (text[0] === "-") {
        index++;
    }

    const first = text[index];

    if (first !== undefined && first >= "0" && first <= "9") {
        return true;
    }

    return false;
}

// go2jsJSONNumberText is the digits of a number as they are written, which for a
// number a double already holds are the digits of the double and for one it did
// not are whatever it can be written back as.
function go2jsJSONNumberText(value) {
    if (typeof value === "string") {
        return value;
    }

    if (typeof value === "number") {
        return String(value);
    }

    if (value === null || value === undefined) {
        return "";
    }

    return String(value);
}

// value holding it, and the name it was written under within it. A name is set
// apart by a character no JSON may hold in a name, so a name that is a number
// or that holds one is still told apart from an element of an array.
function go2jsJSONPath(path, name) {
    return path === "" || path === null || path === undefined ? name : path + "\u0000" + name;
}

// go2jsJSONSkipSpace gives the index of the next character that is not space,
// which is where the value after the space begins.
function go2jsJSONSkipSpace(text, index) {
    while (index < text.length) {
        const code = text.charCodeAt(index);

        if (code !== 32 && code !== 9 && code !== 10 && code !== 13) {
            break;
        }

        index++;
    }

    return index;
}

// go2jsJSONSkipValue gives the index just past the value written at the index,
// which is the extent of the text of that value. A string runs to its closing
// quote, stepping over the quote behind a backslash, and an object and an array
// run to the match for the one they opened with, running over any string they
// hold because a bracket or a brace in a string is only text.
function go2jsJSONSkipValue(text, index) {
    index = go2jsJSONSkipSpace(text, index);

    if (text[index] === '"') {
        index++;

        while (index < text.length) {
            if (text[index] === "\\") {
                index += 2;
                continue;
            }

            if (text[index] === '"') {
                return index + 1;
            }

            index++;
        }

        return index;
    }

    if (text[index] === "{" || text[index] === "[") {
        let depth = 0;

        while (index < text.length) {
            const character = text[index];

            if (character === '"') {
                index = go2jsJSONSkipValue(text, index);
                continue;
            }

            if (character === "{" || character === "[") {
                depth++;
            } else if (character === "}" || character === "]") {
                depth--;

                if (depth === 0) {
                    return index + 1;
                }
            }

            index++;
        }

        return index;
    }

    // a number and the words a value may be written as run to the end of
    // themselves, which is the first character that cannot be part of one
    while (index < text.length) {
        const character = text[index];

        if (character === "," || character === "}" || character === "]" ||
            character === " " || character === "\t" || character === "\n" || character === "\r") {
            break;
        }

        index++;
    }

    return index;
}

// go2jsJSONRawText is the text that was written for the value the place names,
// or nothing when the document does not hold a value there. The value itself is
// the whole of the answer, since what is wanted is the text of a value and not
// a part of one.
function go2jsJSONRawText(text, path) {
    if (typeof text !== "string" || typeof path !== "string") {
        return null;
    }

    const wanted = path === "" ? [] : path.split("\u0000");
    let index = go2jsJSONSkipSpace(text, 0);

    for (const step of wanted) {
        index = go2jsJSONSkipSpace(text, index);

        if (text[index] === "{") {
            index = go2jsJSONSeekMember(text, index + 1, step);

            if (index < 0) {
                return null;
            }

            continue;
        }

        if (text[index] === "[") {
            const position = go2jsJSONSeekElement(text, index + 1, step);

            if (position < 0) {
                return null;
            }

            index = position;
            continue;
        }

        // the place named is not written as a value holding others, so there is
        // nothing there to read
        return null;
    }

    return text.slice(go2jsJSONSkipSpace(text, index), go2jsJSONSkipValue(text, index));
}

// go2jsJSONSeekMember gives the index of the value written under the named key
// of the object that opened at the index, or nothing when the object named no
// such key. A key is read back the way JSON reads it, so a key written with
// escapes names the same key as the same key written without them.
function go2jsJSONSeekMember(text, index, wanted) {
    while (index < text.length) {
        index = go2jsJSONSkipSpace(text, index);

        if (text[index] !== '"') {
            return -1;
        }

        const start = index;
        index = go2jsJSONSkipValue(text, index);
        let name;

        try {
            name = JSON.parse(text.slice(start, index));
        } catch (err) {
            return -1;
        }

        index = go2jsJSONSkipSpace(text, index);

        if (text[index] !== ":") {
            return -1;
        }

        index = go2jsJSONSkipSpace(text, index + 1);

        if (name === wanted) {
            return index;
        }

        index = go2jsJSONSkipValue(text, index);
        index = go2jsJSONSkipSpace(text, index);

        if (text[index] === ",") {
            index++;
        }
    }

    return -1;
}

// go2jsJSONSeekElement gives the index of the value written at the named place
// of the array that opened at the index, or nothing when the array is shorter
// than that.
function go2jsJSONSeekElement(text, index, wanted) {
    let place = 0;
    index = go2jsJSONSkipSpace(text, index);

    while (index < text.length && text[index] !== "]") {
        if (place === Number(wanted)) {
            return index;
        }

        index = go2jsJSONSkipValue(text, index);
        index = go2jsJSONSkipSpace(text, index);

        if (text[index] === ",") {
            index = go2jsJSONSkipSpace(text, index + 1);
        }

        place++;
    }

    return -1;
}

// go2jsJSONZeroStruct builds an anonymous struct with the fields it was
// declared with, each holding what that kind of field holds before anything has
// been read into it, since a field is only read into a place that stands for it
// already.
function go2jsJSONZeroStruct(descriptor) {
    const out = {};

    if (descriptor === null || descriptor === undefined || descriptor.fields === null || descriptor.fields === undefined) {
        return out;
    }

    const types = descriptor.types === null || descriptor.types === undefined ? null : descriptor.types;

    for (const name of Object.keys(descriptor.fields)) {
        out[name] = go2jsJSONZeroValue(types === null ? null : types[name]);
    }

    return out;
}

// go2jsJSONFillZeroFields gives an anonymous struct the fields it was declared
// with that nothing stands for yet, leaving the ones that already stand for
// something as they are.
function go2jsJSONFillZeroFields(struct, descriptor) {
    if (struct === null || struct === undefined || descriptor === null || descriptor === undefined ||
        descriptor.fields === null || descriptor.fields === undefined) {
        return;
    }

    const types = descriptor.types === null || descriptor.types === undefined ? null : descriptor.types;

    for (const name of Object.keys(descriptor.fields)) {
        if (!(name in struct)) {
            struct[name] = go2jsJSONZeroValue(types === null ? null : types[name]);
        }
    }
}

// go2jsJSONZeroValue is what a field holds before the text has anything to say
// about it, which for a struct is a struct of its own with its own fields.
function go2jsJSONZeroValue(descriptor) {
    if (descriptor === null || descriptor === undefined || descriptor.kind === undefined) {
        return null;
    }

    switch (descriptor.kind) {
    case "struct":
    case "lazy":
        return go2jsJSONZeroStruct(descriptor);
    case "slice":
        return [];
    case "map":
        return {};
    case "number":
        return 0;
    case "string":
        return "";
    case "bool":
        return false;
    default:
        return null;
    }
}

// go2jsJSONStore builds the value a destination of the given kind holds out of
// the value the text held, and hands it back for the caller to put where that
// destination stands.
function go2jsJSONStore(value, target, descriptor, text, path) {
    if (descriptor === null || descriptor === undefined || descriptor.kind === undefined) {
        return go2jsJSONCoerceAny(value);
    }

    switch (descriptor.kind) {
    case "ptr":
        // the value behind the pointer is built in place, which for a map and
        // for a plain value is what the thing behind it already knows how to
        // take, and for a struct and a slice is done below
        return go2jsJSONStore(value, target, descriptor.elem, text, path);

    case "struct": {
        if (value === null || typeof value !== "object" || Array.isArray(value)) {
            return null;
        }

        // the value is read into the place that already stands for the struct,
        // and into one built for it when nothing does, which is what reading
        // into a pointer with nothing behind it comes down to
        let struct = go2jsJSONPointerTarget(target);

        // A struct with a name of its own is read through that type, so a value
        // carrying no type is built from the one it was declared under. An
        // anonymous struct is described by the mapping written beside it rather
        // than by a type, so whatever already stands for it is read into.
        if (struct === null || struct === undefined ||
            (descriptor.name !== "" && go2jsJSONStructMapping(struct) === null)) {
            const constructor = descriptor.name === "" ? undefined : go2jsTypeNames[descriptor.name];
            struct = constructor !== undefined && constructor !== null ? new constructor() : go2jsJSONZeroStruct(descriptor);
        } else {
            go2jsJSONFillZeroFields(struct, descriptor);
        }

        go2jsJSONDecode(value, struct, descriptor.fields, descriptor.string, descriptor.types, text, path);

        return struct;
    }

    case "lazy": {
        if (value === null || typeof value !== "object" || Array.isArray(value)) {
            return null;
        }

        // A struct that holds itself is described by the name it was declared
        // under and by nothing else, because the value it stands for is built
        // from that very type and read through that type's own fields. Any depth
        // of one is read the same way, so a type holding itself is described
        // once and read as many times as the text holds of it.
        const constructor = descriptor.name === "" ? undefined : go2jsTypeNames[descriptor.name];
        const struct = constructor !== undefined && constructor !== null ? new constructor() : {};

        // the fields are read off the value the type declared, so a type that
        // holds itself is described once and read as many times as the text
        // holds of it, however deep that text goes
        go2jsJSONDecode(value, struct, null, null, go2jsJSONMappingTypes(struct), text, path);

        return struct;
    }

    case "raw":
        // A raw message is the text that was written for the value rather than
        // the value that text stands for, so the spacing, the escapes and the
        // digits it was written with all come back as they were written. The
        // text is read off the document, which is the only place it is still
        // written; making the value again out of what it was made of would give
        // the same value and not the same text.
        if (text !== null && text !== undefined) {
            const written = go2jsJSONRawText(text, path);

            if (written !== null) {
                return written;
            }
        }

        return go2jsJSONWritten(value);

    case "num": {
        // A number is held as the digits it was written with, because a number
        // no double holds the digits of is kept by holding on to how it was
        // written, so the text is read back off the document rather than made
        // out of the double JSON.parse gave.
        if (text !== null && text !== undefined) {
            const written = go2jsJSONRawText(text, path);

            if (written !== null && go2jsJSONNumberLiteral(written)) {
                return written;
            }
        }

        return go2jsJSONNumberText(value);
    }

    case "slice":
        return go2jsJSONStoreSlice(value, descriptor.elem, text, path);
    case "array":
        return go2jsJSONStoreArray(value, go2jsJSONPointerTarget(target), descriptor.elem, descriptor.len, text, path);

    case "map":
        return go2jsJSONStoreMap(value, target, descriptor.elem, text, path);

    default:
        return go2jsJSONCoerce(value, descriptor.kind);
    }
}

// go2jsJSONStoreSlice builds a slice of what the descriptor holds, one element
// for every one the text had.
function go2jsJSONStoreSlice(value, elem, text, path) {
    if (!Array.isArray(value)) {
        return [];
    }

    const out = new Array(value.length);

    for (let index = 0; index < value.length; index++) {
        out[index] = go2jsJSONStore(value[index], null, elem, text, go2jsJSONPath(path, String(index)));
    }

    return out;
}

// go2jsJSONStoreArray builds an array of what the descriptor holds. An array
// keeps the room it was declared with, so a text too long for it is refused and
// one too short leaves the rest of it as it was declared.
function go2jsJSONStoreArray(value, target, elem, length, text, path) {
    const out = go2jsToArray(target);

    if (!Array.isArray(value) || value.length > length) {
        return out;
    }

    out.length = length;

    for (let index = 0; index < length; index++) {
        out[index] = go2jsJSONStore(value[index], null, elem, text, go2jsJSONPath(path, String(index)));
    }

    return out;
}

// go2jsJSONStoreMap builds a map of what the descriptor holds, one entry for
// every key the text named.
function go2jsJSONStoreMap(value, target, elem, text, path) {
    if (value === null || typeof value !== "object" || Array.isArray(value)) {
        return go2jsJSONTargetMap(target);
    }

    const map = go2jsJSONTargetMap(target);

    if (map === null) {
        // nothing stood for the map yet, which is what a field of it that has
        // not been set comes down to, so one is built to hold what the text
        // named
        return go2jsJSONStoreMap(value, go2jsMakeMap(), elem, text, path);
    }

    for (const key of Object.keys(value)) {
        go2jsMapSet(map, go2jsJSONCoerceString(key), go2jsJSONStore(value[key], null, elem, text, go2jsJSONPath(path, key)));
    }

    return map;
}

// go2jsJSONPointerTarget is the value behind a pointer, which is given one to
// stand in for when it has none yet.
function go2jsJSONPointerTarget(target) {
    if (target === null || target === undefined || target.__go2js_pointer !== true) {
        return target;
    }

    const current = target[go2jsPointerGet]();

    if (current === null || current === undefined || typeof current !== "object") {
        const created = {};
        target[go2jsPointerSet](created);

        return created;
    }

    return current;
}

// go2jsJSONAssign puts a built value where the destination stands, which is
// behind a pointer or the value itself.
function go2jsJSONAssign(target, value) {
    if (target !== null && target !== undefined && target.__go2js_pointer === true) {
        target[go2jsPointerSet](value);
    }
}

function go2jsJSONDelimText(delim) {
    const character = delim !== null && typeof delim === "object" && delim.__go2js_json_delim !== undefined
        ? delim.__go2js_json_delim
        : "";
    return character === "{" || character === "}" || character === "[" || character === "]" ? character : "";
}

function go2jsJSONError(message) {
    return go2jsSentinelError(message)();
}

// A read keeps the first failure it ran into, since that is the one a program
// is told about, and it keeps reading so that a partial value is still there
// for whatever the program does with the fields that did fit.
let go2jsJSONFirstError = null;

function go2jsJSONSaveError(err) {
    if (go2jsJSONFirstError === null) {
        go2jsJSONFirstError = err;
    }
}

// go2jsJSONValueKind names what the text held, the way a read that was handed
// something of the wrong kind describes what it was handed.
function go2jsJSONValueKind(value) {
    if (value === null || value === undefined) {
        return "null";
    }

    if (Array.isArray(value)) {
        return "array";
    }

    switch (typeof value) {
    case "string":
        return "string";
    case "number":
    case "bigint":
        return "number";
    case "boolean":
        return "bool";
    case "object":
        return "object";
    default:
        return "value";
    }
}

// go2jsJSONFitsKind answers whether what the text held can stand for a field of
// the kind it was declared to hold. Nothing in the text stands for anything, so
// a null fits whatever it is read into, which is how Go reads a null into a
// number and leaves it as it was.
function go2jsJSONFitsKind(value, kind) {
    if (value === null || value === undefined) {
        return true;
    }

    switch (kind) {
    case "number":
        return typeof value === "number" || typeof value === "bigint";
    case "string":
        return typeof value === "string";
    case "bool":
        return typeof value === "boolean";
    case "slice":
        return Array.isArray(value);
    case "map":
        return value !== null && typeof value === "object" && !Array.isArray(value);
    case "struct":
    case "lazy":
        return value !== null && typeof value === "object" && !Array.isArray(value);
    default:
        return true;
    }
}

// go2jsJSONElemKind is the kind of what a destination holds, which for a
// pointer is the kind of what it points at.
function go2jsJSONElemKind(descriptor) {
    if (descriptor === null || descriptor === undefined || descriptor.kind === undefined) {
        return null;
    }

    return descriptor.kind === "ptr" ? go2jsJSONElemKind(descriptor.elem) : descriptor.kind;
}

function go2jsJSONTypeName(kind) {
    switch (kind) {
    case "number":
        return "number";
    case "string":
        return "string";
    case "bool":
        return "bool";
    case "slice":
        return "slice";
    case "map":
        return "map";
    default:
        return "value";
    }
}

function go2jsJSONCoerceString(value) {
    if (value === null || value === undefined) {
        return "";
    }

    if (typeof value === "string") {
        return value;
    }

    return String(value);
}

// A number the text held where no type was asked for is a float64, whatever it
// was written with, since that is the one number type the empty interface gives
// it. It is said so by the wrapper it is held in, which is how any other value
// held in an interface says what it is.
function go2jsJSONAnyNumber(value) {
    return go2jsInterface(value, "float64");
}

// A map the text described where no type was asked for holds values nothing is
// known about, so it says that rather than naming a type off the first of them.
function go2jsJSONAnyMap(map) {
    map.__go2js_type = "map[string]interface {}";

    return map;
}

// A list the text described where no type was asked for is a slice of values
// nothing is known about, so it says so rather than naming a type off the
// first of them, which is what lets a type assertion over it be made.
function go2jsJSONAnyList(list) {
    Object.defineProperty(list, "__go2js_type", {
        value: "[]interface {}",
        enumerable: false,
        writable: true,
        configurable: true
    });

    return list;
}

// go2jsJSONCoerceAny is a value of the text read as a value nothing is known
// about: a number is a float64, an object is a map of them, and an array is a
// list of whatever its own elements are.
function go2jsJSONCoerceAny(value) {
    if (value === null || value === undefined) {
        return null;
    }

    if (Array.isArray(value)) {
        return go2jsJSONAnyList(value.map(go2jsJSONCoerceAny));
    }

    if (typeof value === "object") {
        const map = go2jsJSONAnyMap(go2jsMakeMap());

        for (const key of Object.keys(value)) {
            go2jsMapSet(map, key, go2jsJSONCoerceAny(value[key]));
        }

        return map;
    }

    if (typeof value === "number") {
        return go2jsJSONAnyNumber(value);
    }

    return value;
}

function go2jsJSONCoerce(value, kind) {
    if (kind === "string") {
        return go2jsJSONCoerceString(value);
    }

    if (kind === "number") {
        if (typeof value === "number") {
            return value;
        }

        const parsed = Number(value);

        return Number.isNaN(parsed) ? 0 : parsed;
    }

    if (kind === "bool") {
        return Boolean(value);
    }

    return go2jsJSONCoerceAny(value);
}

function go2jsJSONTargetMap(target) {
    const direct = target instanceof go2jsNativeMap ? target : null;

    if (direct !== null) {
        return direct;
    }

    if (target !== null && target !== undefined && target.__go2js_pointer === true) {
        const current = target[go2jsPointerGet]();

        if (current instanceof go2jsNativeMap) {
            return current;
        }

        const created = go2jsMakeMap();
        target[go2jsPointerSet](created);

        return created;
    }

    return null;
}

function go2jsJSONDecode(value, target, fields, stringFields, fieldTypes, text, path) {
    if (value === null || typeof value !== "object" || Array.isArray(value)) {
        return;
    }

    // a struct stands in the target with the fields it was declared with, so a
    // struct being read into is read through its own tags rather than through
    // the tags of the struct that happened to hold it
    const own = go2jsJSONStructMapping(target);

    if (own !== null) {
        if (fields === null || fields === undefined) {
            fields = own.fields;
        }

        if (stringFields === null || stringFields === undefined) {
            stringFields = own.stringFields;
        }
    }

    for (const key of Object.keys(value)) {
        let name = key;

        if (fields) {
            let matched = false;

            for (const field of Object.keys(fields)) {
                if (fields[field] === key && field in target) {
                    name = field;
                    matched = true;
                    break;
                }
            }

            if (!matched) {
                // a field a tag named differently than it is spelled only in
                // case is still that field, and the text has always been taken
                // for it, so it is looked for once more that way before it is
                // looked for at all
                for (const field of Object.keys(fields)) {
                    if (go2jsJSONFold(fields[field]) === go2jsJSONFold(key) && field in target) {
                        name = field;
                        matched = true;
                        break;
                    }
                }
            }

            if (!matched && key in target) {
                name = key;
            }
        }

        if (!(name in target)) {
            continue;
        }

        const current = target[name];
        const item = value[key];

        // a field a tag marked ",string" holds the text of the value, so the
        // text is read back into the value it stands for
        if (stringFields && stringFields.indexOf(name) >= 0 && typeof item === "string") {
            try {
                target[name] = JSON.parse(item);
            } catch (err) {
                target[name] = item;
            }
            continue;
        }

        // a field that stands for a struct of its own is written into the
        // struct that is already there, so the tags of that struct name the
        // fields the JSON holds rather than the field names themselves
        if (current !== null && typeof current === "object" && !Array.isArray(current) &&
            item !== null && typeof item === "object" && !Array.isArray(item) &&
            go2jsJSONStructMapping(current) !== null) {
            go2jsJSONDecode(item, current, null, null, go2jsJSONMappingTypes(current), text, path);
            continue;
        }

        // a field is built as the type it was declared to hold, which for a
        // field the program declared a type for is that type and for one it did
        // not is whatever the text held
        if (fieldTypes !== null && fieldTypes !== undefined && fieldTypes[name] !== undefined) {
            const declared = fieldTypes[name];

            if (declared !== null && declared !== undefined && declared.kind !== undefined &&
                !go2jsJSONFitsKind(item, go2jsJSONElemKind(declared))) {
                go2jsJSONSaveError(go2jsJSONError("json: cannot unmarshal " + go2jsJSONValueKind(item) +
                    " into Go struct field " + name + " of type " + go2jsJSONTypeName(go2jsJSONElemKind(declared))));
                continue;
            }

            target[name] = go2jsJSONStore(item, null, declared, text, go2jsJSONPath(path, key));
            continue;
        }

        // A field the program declared no type for holds a value nothing is
        // known about, which is read as such rather than as the plain value the
        // text happened to be written with.
        target[name] = go2jsJSONCoerceAny(item);
    }
}

function go2jsStringsBuilder() {
    this.parts = [];
}

// A builder is told how much of what it was given it took, and never has any
// reason to refuse it, so the writing methods answer with a count and nothing
// to complain about.
go2jsStringsBuilder.prototype.WriteString = function(value) {
    this.parts.push(String(value));
    return [go2jsStringByteLength(String(value)), null];
};

go2jsStringsBuilder.prototype.Write = function(value) {
    this.parts.push(go2jsBytesToString(value));
    return [go2jsToArray(value).length, null];
};

go2jsStringsBuilder.prototype.WriteRune = function(value) {
    const text = go2jsRuneString(value);

    this.parts.push(text);
    return [go2jsStringByteLength(text), null];
};

go2jsStringsBuilder.prototype.WriteByte = function(value) {
    this.parts.push(String.fromCharCode(Number(value) & 255));
    return null;
};

go2jsStringsBuilder.prototype.String = function() {
    return this.parts.join("");
};

go2jsStringsBuilder.prototype.Len = function() {
    let total = 0;

    for (const part of this.parts) {
        total += go2jsStringByteLength(part);
    }

    return total;
};

go2jsStringsBuilder.prototype.Reset = function() {
    this.parts.length = 0;
};

go2jsStringsBuilder.prototype.Grow = function() {
};

go2jsStringsBuilder.prototype.Cap = function() {
    return this.Len();
};

function go2jsBytesBuffer(initial) {
    this.data = initial === null || initial === undefined ? [] : go2jsToArray(initial);
}

go2jsBytesBuffer.prototype.Write = function(value) {
    if (value === null || value === undefined) {
        return [0, null];
    }

    if (typeof value === "string") {
        value = go2jsStringToBytes(value);
    } else if (ArrayBuffer.isView(value)) {
        value = Array.from(value);
    } else if (Array.isArray(value)) {
        value = value.slice();
    } else {
        throw new TypeError("bytes.Buffer.Write expects byte data");
    }

    this.data.push(...value);
    return [value.length, null];
};

go2jsBytesBuffer.prototype.String = function() {
    return new TextDecoder().decode(Uint8Array.from(this.data));
};

go2jsBytesBuffer.prototype.Bytes = function() {
    return this.data.slice();
};

go2jsBytesBuffer.prototype.Reset = function() {
    this.data.length = 0;
};

go2jsBytesBuffer.prototype.WriteString = function(value) {
    this.Write(value);
    return [go2jsStringByteLength(go2jsStringify(value)), null];
};

go2jsBytesBuffer.prototype.WriteByte = function(value) {
    this.data.push(value & 0xff);
    return null;
};

go2jsBytesBuffer.prototype.WriteRune = function(value) {
    const text = go2jsRuneString(value);

    this.WriteString(text);
    return [go2jsStringByteLength(text), null];
};

go2jsBytesBuffer.prototype.Len = function() {
    return this.data.length;
};

go2jsBytesBuffer.prototype.Cap = function() {
    return this.data.length;
};

go2jsBytesBuffer.prototype.Grow = function(n) {
    return this;
};

go2jsBytesBuffer.prototype.Truncate = function(n) {
    if (n < this.data.length) {
        this.data.length = Math.max(0, n);
    }
};

go2jsBytesBuffer.prototype.UnreadByte = function() {
    this.data.pop();
};

go2jsBytesBuffer.prototype.ReadByte = function() {
    if (this.data.length === 0) {
        // A buffer with nothing left in it is emptied of its room and answers
        // the end of the input rather than a byte it does not have.
        this.data = [];

        return [0, go2jsIOEOF()];
    }

    return [this.data.shift(), null];
};

go2jsBytesBuffer.prototype.Available = function() {
	return this.data.length;
};

go2jsBytesBuffer.prototype.AvailableBuffer = function() {
	const out = new go2jsBytesBuffer();

	out.data = this.data.slice(this.data.length);

	return out;
};

go2jsBytesBuffer.prototype.Read = function(target) {
	const slot = go2jsBytesReadSlot(target);

	if (this.data.length === 0 || !slot) {
		if (target) {
			return [0, go2jsIOEOF()];
		}

		return 0;
	}

	const n = Math.min(this.data.length, slot.length);

	for (let i = 0; i < n; i++) {
		go2jsBytesStoreByte(target, i, this.data[i]);
	}

	this.data = this.data.slice(n);

	if (target) {
		return [n, null];
	}

	return n;
};

function go2jsBytesReadSlot(target) {
	if (!target) {
		return null;
	}

	if (Array.isArray(target) || ArrayBuffer.isView(target)) {
		return target;
	}

	if (Array.isArray(target.data)) {
		return target.data;
	}

	return null;
}

function go2jsBytesStoreByte(target, index, value) {
	if (!target) {
		return;
	}

	const state = go2jsSliceState(target);

	if (state !== null) {
		const at = state.offset + index;

		if (at >= state.offset && at < state.offset + state.length) {
			state.data[at] = value & 255;
		}

		return;
	}

	const slot = go2jsBytesReadSlot(target);

	if (slot !== null && index < slot.length) {
		slot[index] = value & 255;
	}
}

go2jsBytesBuffer.prototype.ReadRune = function() {
	if (this.data.length === 0) {
		return [go2jsRuneEOF, 0];
	}

	const first = this.data[0];

	if (first < 0x80) {
		this.data.shift();
		return [first, 1];
	}

	let width = 0;
	if ((first & 0xe0) === 0xc0) {
		width = 2;
	} else if ((first & 0xf0) === 0xe0) {
		width = 3;
	} else if ((first & 0xf8) === 0xf0) {
		width = 4;
	} else {
		this.data.shift();
		return [0xfffd, 1];
	}

	if (this.data.length < width) {
		this.data = [];
		return [0xfffd, 1];
	}

	const slice = this.data.slice(0, width);
	this.data = this.data.slice(width);

	const text = go2jsBytesToString(slice);
	const decoded = Array.from(text);
	const rune = decoded.length > 0 ? decoded[0].codePointAt(0) : 0xfffd;

	return [rune, width];
};

go2jsBytesBuffer.prototype.UnreadRune = function() {
	return null;
};

go2jsBytesBuffer.prototype.ReadString = function(delim) {
	const needle = go2jsToArray(delim);
	const limit = needle.length;

	if (limit === 0) {
		const out = this.data.slice();
		this.data = [];
		return go2jsBytesReadStringResult(go2jsBytesToString(out), go2jsEOFError());
	}

	for (let i = 0; i + limit <= this.data.length; i++) {
		let match = true;
		for (let j = 0; j < limit; j++) {
			if (this.data[i + j] !== needle[j]) {
				match = false;
				break;
			}
		}
		if (match) {
			const out = this.data.slice(0, i + limit);
			this.data = this.data.slice(i + limit);
			return go2jsBytesReadStringResult(go2jsBytesToString(out), null);
		}
	}

	// A delimiter that never came means the text ran out, so what was there is
	// handed back with the end of it rather than dropped, which is what makes a
	// read of the last piece of a text report the end along with the piece.
	const rest = this.data.slice();
	this.data = [];

	return go2jsBytesReadStringResult(go2jsBytesToString(rest), go2jsEOFError());
};

go2jsBytesBuffer.prototype.ReadBytes = function(delim) {
	const result = this.ReadString(delim);

	// What was read before the end arrived is handed back along with the end,
	// and only a read that found nothing at all has nothing to hand back.
	if (result[0] === "") {
		return [null, result[1]];
	}

	return [go2jsStringToBytes(result[0]), result[1]];
};

go2jsBytesBuffer.prototype.Next = function(n) {
	const count = n === undefined ? 1 : n;

	if (count <= 0) {
		return go2jsStringToBytes("");
	}

	const out = this.data.slice(0, count);
	this.data = this.data.slice(count);

	return out;
};

go2jsBytesBuffer.prototype.WriteTo = function(target) {
	if (target && target.__go2js_discard === true) {
		const written = this.data.length;
		this.data = [];
		return [written, null];
	}

	if (!target || typeof target.Write !== "function") {
		return [0, null];
	}

	const written = this.data.length;

	if (written > 0) {
		go2jsCallNow(target.Write, target, [go2jsStringToBytes(go2jsBytesToString(this.data))]);
	}

	this.data = [];

	return [written, null];
};

go2jsBytesBuffer.prototype.ReadFrom = function(source) {
	const result = go2jsIOReadAll(source);

	if (result[0] === null || result[0] === undefined) {
		return [0, result[1]];
	}

	const text = String(result[0]);
	const bytes = go2jsStringToBytes(text);

	this.data = this.data.concat(bytes);

	return [bytes.length, result[1]];
};

function go2jsBytesReadStringResult(value, err) {
	return [value, err];
}

function go2jsEOFError() {
	return go2jsSentinelError("EOF")();
}

function go2jsBytesEqual(a, b) {
    // A nil byte slice is an empty one, so equality is decided by the bytes and
    // not by which of the two spellings of nothing was used.
    if (a === null || a === undefined) {
        a = [];
    }

    if (b === null || b === undefined) {
        b = [];
    }

    if (ArrayBuffer.isView(a)) {
        a = Array.from(a);
    }

    if (ArrayBuffer.isView(b)) {
        b = Array.from(b);
    }

    if (!Array.isArray(a) || !Array.isArray(b) || a.length !== b.length) {
        return false;
    }

    for (let i = 0; i < a.length; i++) {
        if (a[i] !== b[i]) {
            return false;
        }
    }

    return true;
}

function go2jsNamedMethodCall(receiver, typeName, method, ...args) {
	const fn = go2jsMethodTable[typeName + "." + method];

	if (typeof fn !== "function") {
		throw new TypeError("method " + typeName + "." + method + " is not implemented");
	}

	// A registered method takes the value it works on as its first argument,
	// rather than as the value it is called on, since it is reached by name
	// rather than as a member of the value.
	return go2jsCallNow(fn, null, [receiver, ...args]);
}

// go2jsLookupMember reads a method or field from a plain value, a pointer box,
// or an interface holder so promoted members resolve like direct ones.
function go2jsLookupMember(target, name) {
	if (target === null || target === undefined) {
		return undefined;
	}

	if (target.__go2js_typed_nil === true) {
		const methods = go2jsInterfaceMethods(target.__go2js_type_name);

		if (methods === undefined || methods.indexOf(name) < 0) {
			return undefined;
		}

		return function go2jsTypedNilMethod() {
			throw new TypeError(go2jsRuntimeErrorPrefix + "invalid memory address or nil pointer dereference");
		};
	}

	if (target.__go2js_pointer === true) {
		return go2jsLookupMember(target[go2jsPointerGet](), name);
	}

	if (target.__go2js_interface === true) {
		if (typeof target.type === "string") {
			const fn = go2jsMethodTable[target.type + "." + name];

			if (typeof fn === "function") {
				return function(...args) { return go2jsCallNow(fn, null, [target.value, ...args]); };
			}
		}

		return go2jsLookupMember(target.value, name);
	}

	const value = target[name];

	if (value === undefined) {
		return undefined;
	}

	return typeof value === "function" ? value.bind(target) : value;
}

function go2jsEmbedProxy(target, embedded) {
    return new Proxy(target, {
        get(target, property, receiver) {
		if (property === "__go2js_embedded") {
			return embedded;
		}

		// Marker flags describe the outer value, so an embedded pointer must
		// not leak its own __go2js_pointer through the promotion proxy.
		if (typeof property === "string" && property.startsWith("__go2js")) {
			return Reflect.get(target, property, receiver);
		}


            if (Reflect.has(target, property)) {
                return Reflect.get(target, property, receiver);
            }

            for (const entry of embedded) {
                const current = typeof entry === "string" ? target[entry] : entry;

                if (current === null || current === undefined) {
                    continue;
                }

                const value = go2jsLookupMember(current, property);

                if (value !== undefined) {
                    return value;
                }
            }

            return undefined;
        },

        set(target, property, value, receiver) {
            if (property === "__go2js_embedded") {
                return false;
            }

            if (Reflect.has(target, property)) {
                return Reflect.set(target, property, value, receiver);
            }

            for (const entry of embedded) {
                const current = typeof entry === "string" ? target[entry] : entry;

                if (
                    current !== null &&
                    current !== undefined &&
                    property in Object(current)
                ) {
                    current[property] = value;
                    return true;
                }
            }

            return Reflect.set(target, property, value, receiver);
        },

        has(target, property) {
            if (property === "__go2js_embedded") {
                return true;
            }

            if (Reflect.has(target, property)) {
                return true;
            }

            for (const entry of embedded) {
                const current = typeof entry === "string" ? target[entry] : entry;

                if (
                    current !== null &&
                    current !== undefined &&
                    property in Object(current)
                ) {
                    return true;
                }
            }

            return false;
        }
    });
}

// A pointer reaches the value it points at through an accessor of its own, and
// that accessor cannot be held under a plain property name: a value may well
// have a method called Get or Set, and a pointer that answered for it would call
// the pointer in place of the method. A symbol cannot collide with a name any Go
// method can be given, because no name reaches a symbol.
const go2jsPointerGet = Symbol("go2js.get");

const go2jsPointerSet = Symbol("go2js.set");

function go2jsPointerAccessor(ptr, accessor) {
	return ptr !== null && ptr !== undefined && typeof ptr[accessor] === "function";
}

function go2jsPtr(get, set, typeName) {
	const pointer = {
		[go2jsPointerGet]: function () { return go2jsRunNow(get()); },
		[go2jsPointerSet]: function (value) { return go2jsRunNow(set(value)); }
	};

	if (typeName !== undefined) {
		pointer.__go2js_new_type = typeName;
	}

	return new Proxy(pointer, {
		get(target, property, receiver) {
			if (property === go2jsPointerGet || property === go2jsPointerSet) {
				return Reflect.get(target, property);
			}

			if (property === "__go2js_pointer") {
				return true;
			}

			if (property === "__go2js_new_type") {
				return target.__go2js_new_type;
			}

			const value = target[go2jsPointerGet]();
			if (value === null || value === undefined) {
				return undefined;
			}

			if (typeof value !== "object" && typeof value !== "function") {
				return undefined;
			}

			return Reflect.get(value, property, value);
		},

		set(target, property, value) {
			const current = target[go2jsPointerGet]();
			if (current === null || current === undefined) {
				throw new TypeError("cannot assign through nil pointer");
			}

			return Reflect.set(current, property, value);
		},

		has(target, property) {
			const value = target[go2jsPointerGet]();
			return value !== null && value !== undefined && property in Object(value);
		}
	});
}

function go2jsDeref(ptr) {
	if (ptr === null || ptr === undefined) {
		throw new TypeError("invalid pointer dereference");
	}
	if (!go2jsPointerAccessor(ptr, go2jsPointerGet)) {
		throw new TypeError("value is not a pointer");
	}
	return ptr[go2jsPointerGet]();
}

// go2jsRuneString is the code point a value names, written as the character it
// is, and what a value that names no code point names instead: the mark of a
// run that is not a rune, which is what Go puts in its place.
function go2jsRuneString(value) {
	const codePoint = Number(value);

	if (codePoint < 0 || codePoint > 0x10ffff || (codePoint >= 0xd800 && codePoint <= 0xdfff)) {
		return "\uFFFD";
	}

	return String.fromCodePoint(codePoint);
}

function go2jsNew(value, typeName) {
	return go2jsPtr(
		function() {
			return value;
		},
		function(next) {
			value = next;
		},
		typeName
	);
}

const go2jsMethodTable = Object.create(null);

// A set of query values is a map, so what it holds is read and written as one,
// and what the language asks of it is answered from there.
go2jsRegisterMethod("Values.Get", go2jsURLValuesGet);

go2jsRegisterMethod("Values.Set", go2jsURLValuesSet);

go2jsRegisterMethod("Values.Add", go2jsURLValuesAdd);

go2jsRegisterMethod("Values.Del", go2jsURLValuesDel);

go2jsRegisterMethod("Values.Has", go2jsURLValuesHas);

go2jsRegisterMethod("Values.Encode", go2jsURLValuesEncode);

// A month and a weekday are written as the number they are, so the name they
// print is registered under the name of their type for a value that was boxed
// as that type and so has nowhere of its own to keep the method.
go2jsRegisterMethod("time.Month.String", function(value) {
	return go2jsMonthName(Number(go2jsUnwrap(value)));
});

go2jsRegisterMethod("time.Weekday.String", function(value) {
	return go2jsWeekdayName(Number(go2jsUnwrap(value)));
});

const go2jsInterfaces = Object.create(null);

const go2jsStructFormats = Object.create(null);

const go2jsTypeNames = Object.create(null);

// go2jsTypeNamesByConstructor is the other way round from go2jsTypeNames: a
// value stands on the class it was made from rather than on the name it was
// declared under, and a struct being written has to be named before the fields
// it was declared with can be found.
const go2jsTypeNamesByConstructor = new WeakMap();

const go2jsMethodSets = Object.create(null);

const go2jsStructFields = Object.create(null);

const go2jsTypeKinds = Object.create(null);

function go2jsRegisterTypeName(constructor, name, methods, fields, kind) {
	go2jsTypeNames[name] = constructor;
	go2jsTypeNamesByConstructor.set(constructor, name);

	// The method set travels with the name, because a type that reflect reads
	// off a value rather than off the program has to be told what it can do.
	if (Array.isArray(methods)) {
		go2jsMethodSets[name] = methods;
	}

	// What the fields are travels with it for the same reason: NumField and
	// Field have nothing else to answer with for a type read off a value. A
	// field is more than a name, because a program reads a tag, a package path
	// and a type off one, so a descriptor is kept whole and a bare name is
	// turned into one that says only the name.
	if (Array.isArray(fields)) {
		go2jsStructFields[name] = fields.map((field) =>
			(typeof field === "string" ? {Name: field, Tag: "", PkgPath: "", Anonymous: false, Type: "", Kind: ""} : field));
	}

	// The kind travels with it for the same reason as well: a type reached
	// through a name rather than through a value has no value to read a kind
	// from, and reflect answers about the kind of everything it is given.
	if (typeof kind === "string" && kind !== "") {
		go2jsTypeKinds[name] = kind;
	}
}

// go2jsGoTypeName reports the Go type name of a value for %T.
function go2jsGoTypeName(value) {
	// a verb asking for the type of a value asks for the type as it stands on
	// its own, which is the name without the second one byte and rune have, and
	// every way the name is worked out below is one that can come out needing
	// that, so it is settled once at the end rather than in each of them
	return go2jsNormalizeTypeName(go2jsGoTypeNameRaw(value));
}

function go2jsGoTypeNameRaw(value) {
	// Every value of a type is the one type behind them all, which is a name
	// kept out of reach rather than one a program can ask for, so a descriptor
	// answers with it instead of with the type it describes.
	if (value !== null && value !== undefined && value.__go2js_reflectType === true) {
		return "*reflect.rtype";
	}

	// A box says what it holds rather than the place it is held, because the
	// type a verb asks for is the type of the value and not the type standing
	// over it. A value that answers for its own type says so by carrying the
	// name, and only a value behind an interface is marked that way.
	if (value !== null && value !== undefined && typeof value === "object") {
		const held = value.__go2js_typed === true || value.__go2js_interface === true ? value.value : value;

		if (held !== null && held !== undefined && typeof held.__go2js_type === "string" && held.__go2js_type !== "") {
			return held.__go2js_type;
		}
	}

	if (value !== null && value !== undefined && value.__go2js_typed === true) {
		return value.type;
	}

	// a value that answers for its own type carries the name itself
	if (value !== null && value !== undefined && typeof value === "object") {
		const own = value.__go2js_type_name;

		if (typeof own === "string" && own !== "") {
			return own;
		}

		if (value.__go2js_json_delim !== undefined) {
			return "encoding/json.Delim";
		}
	}

	// A box standing for a nil pointer carries the name of the pointer type, so
	// a verb that asks for the type of the value asks the box rather than the
	// interface that happens to hold it.
	if (value !== null && value !== undefined && value.__go2js_typed_nil === true) {
		return value.__go2js_type_name;
	}

	if (value === null || value === undefined) {
		return "<nil>";
	}

	if (typeof value === "boolean") {
		return "bool";
	}

	if (typeof value === "string" || value instanceof String) {
		return "string";
	}

	if (typeof value === "number" || value instanceof Number) {
		return Number.isInteger(value) ? "int" : "float64";
	}

	if (value instanceof Error) {
		// An error carries the name of the Go type that made it, because one
		// Error stands for all of them here.
		return value.__go2js_error_name !== undefined ? value.__go2js_error_name : "error";
	}

	if (value.__go2js_pointer === true) {
		const pointed = value[go2jsPointerGet]();

		return pointed === null || pointed === undefined ? "*nil" : "*" + go2jsGoTypeName(pointed);
	}

	if (value.__go2js_interface === true) {
		// The error interface says only "error". What %T wants to know is what is
		// really inside it, and context.DeadlineExceeded is not a plain error.
		if (value.type === "error") {
			const inner = go2jsInterfaceValue(value);

			if (inner !== null && inner instanceof Error && typeof inner.__go2js_error_name === "string" && inner.__go2js_error_name !== "") {
				return inner.__go2js_error_name;
			}
		}

		if (typeof value.__go2js_type_name === "string") {
			return value.__go2js_type_name;
		}

		return value.type;
	}

	if (Array.isArray(value)) {
		// A list that was made from a declaration of its own knows what it was
		// declared to hold, and one that was not holds values nothing is known
		// about, so it says that rather than naming a type off the first of them.
		if (typeof value.__go2js_type === "string" && value.__go2js_type !== "") {
			return value.__go2js_type;
		}

		return "[]interface {}";
	}

	if (value instanceof go2jsNativeMap) {
		// A map that was made from a declaration of its own knows the key and the
		// value it was declared to hold, and an empty one has nothing to be read
		// for, so the declaration is what says what it is.
		if (typeof value.__go2js_type === "string" && value.__go2js_type !== "") {
			return value.__go2js_type;
		}

		if (value.size === 0) {
			return "map[interface {}]interface {}";
		}

		return "map[" + go2jsGoTypeName(go2jsFirstKey(value)) + "]" + go2jsGoTypeName(go2jsFirstValue(value));
	}

	const ctor = value.constructor;

	if (ctor && typeof ctor.name === "string") {
		const registered = go2jsLookupTypeName(ctor.name);

		if (registered !== undefined) {
			return registered;
		}
	}

	if (ctor && typeof ctor.name === "string" && ctor.name !== "Object") {
		return "main." + ctor.name;
	}

	return "interface {}";
}

function go2jsHasFormatMethod(value) {
	const receiver = value !== null && value !== undefined && value.__go2js_pointer === true
		? go2jsDeref(value)
		: value;

	if (receiver === null || receiver === undefined || typeof receiver !== "object") {
		return false;
	}

	if (receiver.__go2js_interface === true || Array.isArray(receiver) || receiver instanceof go2jsNativeMap) {
		return true;
	}

	return go2jsNamedFormatMethod(receiver, "Error") !== null || go2jsNamedFormatMethod(receiver, "String") !== null;
}

// go2jsFormatFields writes a value the way %+v does, which is the plain walk
// with a name in front of every field.
function go2jsFormatFields(value) {
	return go2jsFormat(value, null, null, null, true, false);
}

function go2jsNamedFormatMethod(value, name) {
	if (value.__go2js_interface === true || value.__go2js_pointer === true) {
		return null;
	}

	if (typeof value[name] === "function") {
		return go2jsFormat(go2jsCallNow(value[name], value, []));
	}
	// A box standing for a named type carries its own way of writing itself
	// out, which is what a type with a String method of its own is shown by.
	// A pointer that points at nothing is asked what it is rather than what it
	// holds, since looking for a method on it would be reaching through a
	// pointer that is not there.
	if (value.__go2js_typed === true && value.value !== null && value.value !== undefined &&
		go2jsIsTypedNilPointer(value.value) === false && typeof value.value[name] === "function") {
		return go2jsFormat(go2jsCallNow(value.value[name], value.value, []));
	}

	const ctor = value.constructor;

	if (ctor && typeof ctor.name === "string" && ctor.name !== "Object") {
		const registered = go2jsLookupTypeName(ctor.name);

		if (registered !== undefined && typeof go2jsMethodTable[registered + "." + name] === "function") {
			return go2jsFormat(go2jsCallNow(go2jsMethodTable[registered + "." + name], null, [value]));
		}
	}

	const goType = go2jsGoTypeNameRaw(value);
	if (typeof goType === "string" && goType !== "" && goType !== "object") {
		const fn = go2jsMethodTable[goType + "." + name];
		if (typeof fn === "function") {
			return go2jsFormat(go2jsCallNow(fn, null, [value]));
		}
	}

	return null;
}

// go2jsLookupNamedMethod finds a method of a value's own type and hands it back
// un-called, which is what a method that takes arguments needs. A method of the
// type the emitter wrote as a class is reached on the value, and one it wrote as
// a registered function, because a method of a type that is not a class has
// nowhere to live, is reached through the table it registered itself in.
function go2jsLookupNamedMethod(receiver, name) {
	if (receiver === null || receiver === undefined || go2jsIsTypedNilPointer(receiver)) {
		return null;
	}

	if (receiver.__go2js_pointer === true) {
		receiver = go2jsDeref(receiver);
	}

	if (receiver.__go2js_interface === true) {
		return null;
	}

	if (typeof receiver[name] === "function") {
		return receiver[name].bind(receiver);
	}

	const ctor = receiver.constructor;

	if (ctor && typeof ctor.name === "string" && ctor.name !== "Object") {
		const registered = go2jsLookupTypeName(ctor.name);

		if (registered !== undefined && typeof go2jsMethodTable[registered + "." + name] === "function") {
			return (...args) => go2jsCallNow(go2jsMethodTable[registered + "." + name], null, [receiver, ...args]);
		}
	}

	return null;
}

function go2jsLookupTypeName(jsName) {
	for (const name of Object.keys(go2jsTypeNames)) {
		if (go2jsTypeNames[name].name === jsName) {
			return name;
		}
	}

	return undefined;
}

function go2jsFirstKey(value) {
	for (const [key] of value) {
		return key;
	}

	return null;
}

function go2jsFirstValue(value) {
	for (const [, item] of value) {
		return item;
	}

	return null;
}

function go2jsRegisterMethod(name, fn) {
	go2jsMethodTable[name] = fn;
}

// go2jsPointerAddress gives a pointer a stable address to print. A real Go
// address is the one thing a translation cannot reproduce, so each pointer is
// handed a number of its own that stays the same for the life of the value,
// which is what lets a printed address still be compared with the next one.
const go2jsPointerAddresses = new WeakMap();

let go2jsPointerAddressCount = 0;

function go2jsPointerAddress(pointer) {
	// A pointer that points at nothing has no address to write, so fmt names it
	// rather than pointing at a place in a heap that is not there.
	let target;

	if (!go2jsPointerAccessor(pointer, go2jsPointerGet)) {
		return "<nil>";
	}

	try {
		target = pointer[go2jsPointerGet]();
	} catch (e) {
		target = null;
	}

	if (target === null || target === undefined) {
		return "<nil>";
	}

	let address = go2jsPointerAddresses.get(pointer);

	if (address === undefined) {
		go2jsPointerAddressCount++;
		address = go2jsPointerAddressCount;
		go2jsPointerAddresses.set(pointer, address);
	}

	// Go writes a heap address in twelve hex digits, so the same width is used
	// here to keep the shape of the output right.
	let text = (0xc000000000 + address * 8).toString(16);

	return "0x" + text.padStart(12, "0");
}

// go2jsPointerTargets reports whether a pointer names a struct, an array, a
// slice or a map, which are the ones fmt writes with a leading & rather than as
// a bare address.
function go2jsPointerTargets(pointer) {
	let value;

	if (!go2jsPointerAccessor(pointer, go2jsPointerGet)) {
		return false;
	}

	try {
		value = pointer[go2jsPointerGet]();
	} catch (error) {
		return false;
	}

	// Only a pointer to a struct, an array, a slice or a map is written with a
	// leading &. A pointer to a pointer is written as a bare address, since the
	// value it points at is itself only an address.
	if (value !== null && value !== undefined && value.__go2js_pointer === true) {
		return false;
	}

	return value !== null && value !== undefined && typeof value === "object";
}

function go2jsInterface(value, typeName, displayName) {
	// An interface names no type of its own, so the type of the value it holds
	// is the one that has to be written down. Reading the name of the interface
	// back off the wrapper would claim that every value stored in one has the
	// same type, and a value compared against a plain one of another type would
	// never be equal.
	if (go2jsIsInterfaceTypeName(typeName)) {
		const named = go2jsGoTypeName(value);

		if (named !== "" && !go2jsIsInterfaceTypeName(named)) {
			typeName = named;

			if (displayName === "any" || displayName === "interface {}" || displayName === "interface{}") {
				displayName = named;
			}
		}
	}

	const wrapper = {
		__go2js_interface: true,
		type: typeName,
		value: value
	};

	if (typeof displayName === "string" && displayName !== "" && displayName !== typeName) {
		wrapper.__go2js_type_name = displayName;
	}

	if (typeName !== undefined && typeName !== null && typeName !== "" && value !== null && value !== undefined && typeof value === "object") {
		// A value that already knows what it really is gives the better answer to
		// %T than the interface it is being dressed up for. Wrapping
		// context.DeadlineExceeded as an error must not turn it back into a plain
		// error, the way a hand wrapped in a box is still a hand.
		const inner = go2jsInterfaceValue(value);
		const known = inner !== null && typeof inner === "object" ? inner.__go2js_error_name : undefined;

		if (known === undefined || known === null || known === "") {
			wrapper.__go2js_error_name = String(typeName).replace(/^\*/, "");
		} else {
			wrapper.__go2js_error_name = known;
		}
	}

	if (Array.isArray(value) && typeof typeName === "string" && typeName.startsWith("[")) {
		for (const [key, entry] of Object.entries(wrapper)) {
			Object.defineProperty(value, key, {
				value: entry,
				enumerable: false,
				writable: true,
				configurable: true
			});
		}

		Object.defineProperty(value, "value", {
			value: value,
			enumerable: false,
			writable: true,
			configurable: true
		});

		return value;
	}

	return wrapper;
}

// go2jsSameReflectType reports whether two type descriptors describe one type.
// The kind, the name, the length and what the type is made of are all of what
// makes a type the type it is.
function go2jsSameReflectType(left, right) {
	if (left === right) {
		return true;
	}

	if (left === null || left === undefined || right === null || right === undefined) {
		return false;
	}

	return left.kind === right.kind &&
		(left.name || "") === (right.name || "") &&
		(left.len === null ? null : left.len) === (right.len === null ? null : right.len) &&
		go2jsSameReflectType(left.elem, right.elem) &&
		go2jsSameReflectType(left.key, right.key);
}

function go2jsEqual(a, b) {
	// A pointer that is nil and is held in an interface is a value of its own
	// rather than nothing, so the interface holding it is not nil even though
	// the pointer is, and it is the interface that is being compared here.
	const aNil = a === null || a === undefined;
	const bNil = b === null || b === undefined;

	if (aNil && bNil) {
		return true;
	}

	if (aNil || bNil) {
		return false;
	}

	if (a === b) {
		return true;
	}

	// A type read off a value is not the very object the program declared, so
	// two types are compared by what they describe, which is what decides
	// whether they are one type.
	if (a.__go2js_reflectType === true && b.__go2js_reflectType === true) {
		return go2jsSameReflectType(a, b);
	}

	const ai = a.__go2js_interface === true;
	const bi = b.__go2js_interface === true;

	if (ai && bi) {
		if (a.type !== b.type) {
			return false;
		}

		// An array that is its own interface wrapper keeps its elements
		// directly, so unwrapping would only arrive back at the same object.
		if (a.value !== a && b.value !== b) {
			return go2jsEqual(a.value, b.value);
		}
	} else if (ai || bi) {
		// One side is a value that was stored in an interface and the other is
		// not. Go compares the value the interface holds against the value on
		// the other side, so the wrapper is peeled off and the two values are
		// compared, which leaves a value of a different type unequal.
		const wrapped = ai ? a : b;
		const plain = ai ? b : a;

		if (wrapped.value === wrapped) {
			return false;
		}

		return go2jsEqual(wrapped.value, plain);
	}

	if (Array.isArray(a) || Array.isArray(b)) {
		if (!Array.isArray(a) || !Array.isArray(b) || a.length !== b.length) {
			return false;
		}

		for (let index = 0; index < a.length; index++) {
			if (!go2jsEqual(a[index], b[index])) {
				return false;
			}
		}

		return true;
	}

	// Two errors are two names for two conditions, and they are one condition
	// only when they are one value, since what an error holds is a message
	// rather than a set of fields to be read one by one.
	if (go2jsIsErrorValue(a) && go2jsIsErrorValue(b)) {
		return false;
	}

	if (typeof a === "object" || typeof b === "object") {
		if (typeof a !== typeof b || a === null || b === null) {
			return false;
		}

		const ak = Object.keys(a);
		const bk = Object.keys(b);

		if (ak.length !== bk.length) {
			return false;
		}

		for (const key of ak) {
			if (!Object.prototype.hasOwnProperty.call(b, key)) {
				return false;
			}

			// What a value can be done with is not part of the value: two values
			// of a type are one value when what they hold is one, and the methods
			// of the type are the same for both of them.
			if (typeof a[key] === "function" && typeof b[key] === "function") {
				continue;
			}

			if (!go2jsEqual(a[key], b[key])) {
				return false;
			}
		}

		return true;
	}

	return a === b;
}

// go2jsIsErrorValue reports whether a value is an error rather than some other
// value that happens to be built the same way.
function go2jsIsErrorValue(value) {
	if (value === null || value === undefined || typeof value !== "object") {
		return false;
	}

	if (value instanceof Error) {
		return true;
	}

	return typeof value.Error === "function" && typeof value.message === "string";
}

function go2jsInterfaceValue(value) {
	if (value === null || value === undefined) {
		return null;
	}

	if (value.__go2js_interface === true) {
		return value.value;
	}

	return value;
}

// The method behind an interface is a Go function, so it is a generator: the
// call is delegated, which lets the method wait on a channel without taking the
// caller down with it, and lets the caller wait on the method.
// go2jsInterfaceCallNow is the interface call the runtime itself makes, which
// is one it makes from inside something that cannot wait: the method is run to
// its end here, answering whatever it waits on by running the goroutines that
// are runnable until it is through.
function go2jsInterfaceCallNow(value, method, ...args) {
	return go2jsRunNow(go2jsInterfaceCall(value, method, ...args));
}

function* go2jsInterfaceCall(value, method, ...args) {
	if (value === null || value === undefined) {
		throw new TypeError(go2jsRuntimeErrorPrefix + "invalid memory address or nil pointer dereference");
	}

	if (value instanceof Error && method === "Error") {
		return value.message;
	}

	if (value.__go2js_interface !== true) {
		if (value.__go2js_pointer === true) {
			return yield* go2jsInterfaceCall(value[go2jsPointerGet](), method, ...args);
		}

		if (typeof value[method] === "function") {
			return yield* go2jsCall(value[method], value, args);
		}

		throw new TypeError("value is not an interface");
	}

	const target = value.value;

	if (target === null || target === undefined) {
		throw new TypeError(go2jsRuntimeErrorPrefix + "invalid memory address or nil pointer dereference");
	}

	if (typeof value.type === "string" && target.__go2js_pointer === true) {
		const key = value.type.startsWith("*")
			? value.type.slice(1) + "." + method
			: value.type + "." + method;
		const registered = go2jsMethodTable[key];

		// A method reached through the table is a wrapper that takes the value
		// it works on as its first argument, so that is how it is called here.
		if (typeof registered === "function") {
			return yield* go2jsCall(registered, null, [target, ...args]);
		}
	}

	let fn = target[method];
	let registered = false;

	if (typeof fn !== "function" && typeof value.type === "string") {
		fn = go2jsMethodTable[value.type + "." + method];
		registered = true;
	}

	if (typeof fn !== "function") {
		throw new TypeError("interface method " + method + " is not implemented");
	}

	if (registered) {
		return yield* go2jsCall(fn, null, [target, ...args]);
	}

	return yield* go2jsCall(fn, target, args);
}

function go2jsIsTypedNilPointer(value) {
	return value !== null && value !== undefined &&
		value.__go2js_typed_nil === true;
}

// go2jsNormalizeTypeName gives a type the one name Go has for it, since byte is
// uint8 and rune is int32 under two names each, and a type written one way is
// asked about by the other as often as not.
function go2jsNormalizeTypeName(name) {
	if (typeof name !== "string" || name === "") {
		return name;
	}

	// byte and rune are the other names for uint8 and int32, and a name given
	// for its own sake is not one of them, so a name a package stands behind is
	// left as it stands.
	return name
		.replace(/(?<![.\w])byte(?![.\w])/g, "uint8")
		.replace(/(?<![.\w])rune(?![.\w])/g, "int32");
}

function go2jsInterfaceKey(name) {
	if (typeof name !== "string") {
		return "";
	}

	const trimmed = name.startsWith("*") ? name.slice(1) : name;
	const dot = trimmed.lastIndexOf(".");

	return dot === -1 ? trimmed : trimmed.slice(dot + 1);
}

// go2jsBuiltinInterfaces records the method set of the interfaces a program can
// name without declaring them, so a type assertion or a type switch clause that
// names one can succeed for a type carrying its methods. A name the program
// declares itself is registered over these, because a declaration of its own
// is the one that counts.
const go2jsBuiltinInterfaces = {
	"error": ["Error"],
	"rand.Source": ["Int63", "Seed"],
	"sort.Interface": ["Len", "Less", "Swap"],
	"Stringer": ["String"],
	"GoStringer": ["GoString"],
	"Formatter": ["Format"],
	"State": ["Flag", "Width", "Precision", "Write"],
	"Reader": ["Read"],
	"Writer": ["Write"],
	"ReadWriter": ["Read", "Write"],
	"ReadCloser": ["Read", "Close"],
	"WriteCloser": ["Write", "Close"],
	"ReadWriteCloser": ["Read", "Write", "Close"],
	"Closer": ["Close"],
	"Seeker": ["Seek"],
	"StringWriter": ["WriteString"],
};

function go2jsInterfaceMethods(name) {
	const key = go2jsInterfaceKey(name);

	if (key === "") {
		return undefined;
	}

	const methods = go2jsInterfaces[key];

	return methods !== undefined ? methods : go2jsBuiltinInterfaces[key];
}

function go2jsComplexValue(value) {
	if (value !== null && typeof value === "object" && "re" in value && "im" in value) {
		return value;
	}
	return { re: Number(value), im: 0 };
}

function go2jsLen(value) {
	if (value === null || value === undefined) {
		return 0;
	}
	if (value.__go2js_pointer === true) {
		return go2jsLen(value[go2jsPointerGet]());
	}
	if (value instanceof go2jsNativeMap || value instanceof go2jsNativeSet) {
		return value.size;
	}
	if (typeof value === "string") {
		return go2jsStringByteLength(value);
	}
	if (Array.isArray(value)) {
		return value.length;
	}
	if (typeof value === "object") {
		return Object.keys(value).length;
	}
	return 0;
}

// go2jsStringByteLength mirrors Go's len(string), which counts UTF-8 bytes
// rather than UTF-16 code units.
function go2jsStringByteLength(s) {
	let bytes = 0;

	for (let i = 0; i < s.length; i++) {
		const unit = s.charCodeAt(i);

		// A byte that is not the UTF-8 of a rune is a byte of its own, held as a
		// lone surrogate, and takes up room on its own.
		if (unit >= 0xdc00 && unit <= 0xdfff && (i === 0 || s.charCodeAt(i - 1) < 0xd800 || s.charCodeAt(i - 1) > 0xdbff)) {
			bytes += 1;
			continue;
		}

		let code = unit;

		if (unit >= 0xd800 && unit <= 0xdbff && i + 1 < s.length) {
			const low = s.charCodeAt(i + 1);

			if (low >= 0xdc00 && low <= 0xdfff) {
				code = (unit - 0xd800) * 0x400 + (low - 0xdc00) + 0x10000;
				i++;
			}
		}

		if (code < 0x80) {
			bytes += 1;
		} else if (code < 0x800) {
			bytes += 2;
		} else if (code < 0x10000) {
			bytes += 3;
		} else {
			bytes += 4;
		}
	}

	return bytes;
}

function go2jsMakeMap() {
	return new go2jsNativeMap();
}

function go2jsMapTypeName(typeName) {
	return String(typeName).replace(/\bany\b/g, "interface {}");
}

// go2jsSliceTyped says what a list was written as, the same way a map literal
// is marked with the type it was written as, so that a printed slice carries
// the type Go writes rather than one read off the first value it holds.
function go2jsSliceTyped(typeName, list) {
	if (list !== null && list !== undefined && typeName !== undefined && typeName !== null && typeName !== "") {
		Object.defineProperty(list, "__go2js_type", {
			value: typeName,
			writable: true,
			configurable: true,
			enumerable: false
		});
	}

	return list;
}

function go2jsMapTyped(typeName, map) {
	if (map !== null && map !== undefined && typeName !== undefined && typeName !== null && typeName !== "") {
		Object.defineProperty(map, "__go2js_type", {
			value: go2jsMapTypeName(typeName),
			writable: true,
			configurable: true,
			enumerable: false
		});
	}

	return map;
}

// A key that is a struct or an array is found by what it holds rather than by
// which object it happens to be, so two keys built alike reach the same entry
// and a key built apart does not. JavaScript tells objects apart by identity, so
// the first key of a shape is kept and every later key of that shape is given
// the one that was kept.
const go2jsMapKeyShapes = new Map();

function go2jsMapKeySignature(value) {
	if (Array.isArray(value)) {
		let signature = "[";

		for (let index = 0; index < value.length; index++) {
			signature += "|" + go2jsMapKeyPart(value[index]);
		}

		return signature + "]";
	}

	if (value === null || typeof value !== "object") {
		return null;
	}

	if (value.__go2js_pointer === true || value.__go2js_interface === true ||
		value.__go2js_reflectValue === true || value.__go2js_reflectType === true) {
		return null;
	}

	const ctor = value.constructor;
	const name = ctor !== undefined && ctor !== null && typeof ctor.name === "string" ? ctor.name : "";
	const keys = Object.keys(value).sort();
	let signature = "{" + name;

	for (const key of keys) {
		signature += "|" + key + "=" + go2jsMapKeyPart(value[key]);
	}

	return signature + "}";
}

function go2jsMapKeyPart(value) {
	if (value === null || value === undefined) {
		return " nil";
	}

	if (typeof value === "object" || typeof value === "function") {
		const nested = go2jsMapKeySignature(value);

		return nested === null ? " obj" : " " + nested;
	}

	return typeof value.charAt === "function" ? "s" + value : typeof value + value;
}

// go2jsMapKey hands back the one object every key of that shape shares, so a key
// is found by what it holds. A key that is not a struct or an array is left
// alone, since a string or a number is already found by what it is.
function go2jsMapKey(key) {
	const signature = go2jsMapKeySignature(key);

	if (signature === null) {
		return key;
	}

	let shape = go2jsMapKeyShapes.get(signature);

	if (shape === undefined) {
		go2jsMapKeyShapes.set(signature, key);

		return key;
	}

	return shape;
}

function go2jsMap(entries) {
	const map = new go2jsNativeMap();

	if (!entries) {
		return map;
	}
	for (const entry of entries) {
		if (!Array.isArray(entry) || entry.length < 2) {
			continue;
		}
		map.set(go2jsMapKey(entry[0]), entry[1]);
	}

	return map;
}

function go2jsMapGet(map, key, zero) {
	if (!(map instanceof go2jsNativeMap)) {
		return zero;
	}

	const value = map.get(go2jsMapKey(key));

	// A missing key yields the element type's zero value, not null.
	return value === undefined ? zero : value;
}

function go2jsMapSet(map, key, value) {
	if (map === null || map === undefined) {
		go2jsPanic("assignment to entry in nil map");
	}

	if (map.__go2js_nil === true) {
		go2jsPanic("assignment to entry in nil map");
	}

	if (!(map instanceof go2jsNativeMap)) {
		throw new TypeError("go2jsMapSet expects a Map");
	}
	map.set(go2jsMapKey(key), value);
}

function go2jsMapDelete(map, key) {
	if (!(map instanceof go2jsNativeMap)) {
		throw new TypeError("go2jsMapDelete expects a Map");
	}
	map.delete(go2jsMapKey(key));
}

function go2jsMapHas(map, key) {
	if (!(map instanceof go2jsNativeMap)) {
		return false;
	}
	return map.has(go2jsMapKey(key));
}

function go2jsMapKeys(map) {
	if (!(map instanceof go2jsNativeMap)) {
		return [];
	}
	return Array.from(map.keys());
}

function go2jsCopy(value) {
	// A value of a type of Go that stands for the value behind a pointer is one
	// value rather than a record of fields, so copying it is the same value seen
	// twice rather than a second one written out beside it.
	if (value !== null && value !== undefined && value.__go2js_shared === true) {
		return value;
	}

	// A slice is a name for storage somebody else owns, so a copy of one is the
	// same name rather than a second run of elements. The test is the metadata a
	// slice carries and not the shape of the value, because an array is the same
	// shape and is a value of its own.
	if (value !== null && value !== undefined && go2jsSliceMeta.has(value)) {
		return value;
	}

	if (Array.isArray(value)) {
		if (!go2jsArrayMark.has(value)) {
			return value;
		}
		const copy = value.slice();
		for (let i = 0; i < copy.length; i++) {
			copy[i] = go2jsCopy(copy[i]);
		}
		go2jsArrayMark.add(copy);
		return copy;
	}

	if (value instanceof go2jsNativeMap) {
		return new go2jsNativeMap(value);
	}

	if (value && typeof value === "object") {
		if (value.__go2js_pointer === true || value instanceof go2jsNativeDate) {
			return value;
		}
		if (typeof value.constructor === "function" && value.constructor !== Object && value.constructor !== Array) {
			if (typeof go2jsStructCopy === "function") {
				return go2jsStructCopy(value);
			}
		}
		const copy = {};
		for (const key of Object.keys(value)) {
			copy[key] = go2jsCopy(value[key]);
		}
		return copy;
	}

	return value;
}

function go2jsToArray(value) {
	if (value === null || value === undefined) {
		return [];
	}

	if (Array.isArray(value)) {
		return value.slice();
	}

	if (value instanceof go2jsNativeMap || value instanceof go2jsNativeSet) {
		return Array.from(value);
	}

	if (typeof value[Symbol.iterator] === "function") {
		return Array.from(value);
	}

	return [value];
}

// go2jsDivide keeps an integer division honest: a divisor of zero stops the
// program where Go stops it rather than answering Infinity or NaN.
// A whole number Go can be asked for goes past the point where a double holds
// every one of its numbers, and JavaScript has a second kind of whole number for
// that, one that keeps every digit it is given. Both kinds are read here so an
// operation works out the same whichever it is handed, and so a number that fits
// a double is left as one rather than paid for twice over.
const go2jsSafeInteger = 9007199254740991;

// go2jsIsWide reports whether a value is a whole number too wide for a double,
// which is the only thing the second kind is used for here.
function go2jsIsWide(value) {
	return typeof value === "bigint";
}

// go2jsWide reads a value as a whole number of the wider kind, whichever kind it
// came in as. A number that is already whole and within reach is turned over
// exactly, and one that has a fraction in it is left alone rather than rounded
// away, because a fraction is not something a whole number operation can have.
function go2jsWide(value) {
	if (typeof value === "bigint") {
		return value;
	}

	return BigInt(Math.trunc(Number(value)));
}

// go2jsNarrow reads a whole number back as a double once it fits one, so a
// program that only ever sees small numbers never notices any of this.
function go2jsNarrow(value) {
	if (typeof value !== "bigint") {
		return value;
	}

	if (value <= go2jsSafeInteger && value >= -go2jsSafeInteger) {
		return Number(value);
	}

	return value;
}

// go2jsWideWrap keeps a whole number inside the sixty four bits a signed or an
// unsigned number of that width is held in, which is what a Go variable of that
// type does without being asked. An answer too wide for its type is not a fault
// in Go, it is the number that fits, so a signed one past its largest is the
// smallest it has and an unsigned one below its smallest is the largest it has.
function go2jsWideWrap(value, typeName) {
	if (typeof value !== "bigint") {
		const number = Math.trunc(Number(value));

		// a double past the point where it holds every digit is worked out as
		// digits before it is given a width, since the digits it is carrying are
		// not the ones it was handed
		if (Number.isInteger(number) && (number > go2jsSafeInteger || number < -go2jsSafeInteger)) {
			return go2jsWideWrap(BigInt(number), typeName);
		}

		return number;
	}

	// the width the type is held in, and whether it goes below zero at all
	const bits = { "uint8": 8, byte: 8, "int8": 8, "uint16": 16, "int16": 16, "uint32": 32, "int32": 32, rune: 32 };
	const unsigned = typeName === "uint64" || typeName === "uint" || typeName === "uintptr" ||
		typeName === "uint8" || typeName === "byte" || typeName === "uint16" || typeName === "uint32";
	const width = bits[typeName] === undefined ? 64 : bits[typeName];

	if (width === 64) {
		return go2jsNarrow(unsigned ? (value & go2jsUint64Mask) : go2jsWrapSigned64(value));
	}

	const modulus = 1n << BigInt(width);
	const wrapped = unsigned ? (value & (modulus - 1n)) : ((value % modulus) + modulus) % modulus;

	return go2jsNarrow(unsigned ? wrapped : (wrapped >= modulus >> 1n ? wrapped - modulus : wrapped));
}

const go2jsUint64Mask = (1n << 64n) - 1n;

function go2jsWrapSigned64(value) {
	const wrapped = ((value % go2jsUint64One) + go2jsUint64One) % go2jsUint64One;

	return wrapped >= go2jsSignBit64 ? wrapped - go2jsUint64One : wrapped;
}

const go2jsUint64One = 1n << 64n;

const go2jsSignBit64 = 1n << 63n;

// go2jsWideBinary runs an operation over two whole numbers and gives back
// whichever kind of number the answer is one of. A number that fits a double is
// read back as one, so a program working in the range a double covers is not
// slowed down or made to answer differently by a range it never reaches.
function go2jsWideBinary(left, right, whole, typeName) {
	left = go2jsDurationOperand(left);
	right = go2jsDurationOperand(right);

	if (!go2jsIsWide(left) && !go2jsIsWide(right)) {
		return go2jsWideWrap(whole(left, right), typeName);
	}

	return go2jsWideWrap(whole(go2jsWide(left), go2jsWide(right)), typeName);
}

function go2jsWideAdd(left, right, typeName) {
	return go2jsWideBinary(left, right, (a, b) => a + b, typeName);
}

function go2jsWideMul(left, right, typeName) {
	return go2jsWideBinary(left, right, (a, b) => a * b, typeName);
}

function go2jsWideQuo(left, right, typeName) {
	return go2jsWideBinary(left, right, (a, b) => {
		if (b === 0) {
			throw new go2jsNativeRangeError(go2jsRuntimeErrorPrefix + "integer divide by zero");
		}

		return a / b;
	}, typeName);
}

// go2jsDurationOperand reads the whole number out of a span of time and leaves
// every other value as it stands, so a helper that works in whole numbers can
// take one of either without being told which it has.
function go2jsDurationOperand(value) {
	if (go2jsIsDuration(value)) {
		return value.nanoseconds;
	}

	return value;
}

function go2jsDivide(left, right) {
	// a span of time is asked for the whole number it holds before the two are
	// divided, since it is that number and not the shape that does the work
	left = go2jsDurationOperand(left);
	right = go2jsDurationOperand(right);

	if (go2jsIsWide(left) || go2jsIsWide(right)) {
		return go2jsWideQuo(left, right);
	}

	if (right === 0) {
		throw new go2jsNativeRangeError(go2jsRuntimeErrorPrefix + "integer divide by zero");
	}

	return Math.trunc(left / right);
}

function go2jsPanic(value) {
	const error = value instanceof Error ? value : new Error(go2jsStringify(value));

	if (error.__go2js_panic_value === undefined) {
		error.__go2js_panic_value = value;
	}

	throw error;
}

// Go says a runtime fault in the same words however it was reached, so a
// JavaScript error about a missing value is given Go's wording on the way out.
function go2jsRuntimeErrorText(value) {
	// An index past the end of something and a division by nothing are faults
	// the engine raises on its own behalf, in the same way that reaching
	// through a standing-for-nothing value is, and Go words all three the same
	// way however they were reached.
	if (!(value instanceof go2jsNativeTypeError) && !(value instanceof go2jsNativeRangeError)) {
		return null;
	}

	const message = String(value.message === undefined ? "" : value.message);

	if (/^Cannot (read|set) propert(?:y|ies) of (null|undefined)/.test(message)) {
		return go2jsRuntimeErrorPrefix + "invalid memory address or nil pointer dereference";
	}

	if (/^\w+ is not a function$/.test(message) || /of undefined$/.test(message)) {
		return go2jsRuntimeErrorPrefix + "invalid memory address or nil pointer dereference";
	}

	// A fault the runtime raised carries the words Go puts in front of one
	// already, so the fault itself is what is judged and the words are put
	// back on the way out rather than left where they were found.
	const fault = message.startsWith(go2jsRuntimeErrorPrefix)
		? message.slice(go2jsRuntimeErrorPrefix.length)
		: message;

	if (fault.startsWith("index out of range") || fault.startsWith("integer divide by zero")) {
		return go2jsRuntimeErrorPrefix + fault;
	}

	return null;
}

// A fault the runtime raises about itself rather than about the program, such as
// every goroutine being asleep with nothing left to wake one, is not a panic:
// Go ends the program with a fatal error instead, and the difference is what
// tells the two apart on the way out.
function go2jsFatalError(text) {
	const error = new Error(text);
	error.__go2js_fatal = text;

	return error;
}

function go2jsPanicPayload(value) {
	if (value !== null && value !== undefined && value.__go2js_panic_value !== undefined) {
		return value.__go2js_panic_value;
	}

	const reworded = go2jsRuntimeErrorText(value);

	if (reworded !== null) {
		return reworded;
	}

	return value;
}

function go2jsRecover() {
	return undefined;
}

function go2jsStringify(value) {
	if (value === null || value === undefined) {
		return "<nil>";
	}
	if (value instanceof Error) {
		return value.message;
	}
	if (value.__go2js_interface === true) {
		return go2jsStringify(value.value);
	}
	if (value instanceof go2jsNativeMap) {
		const parts = [];
		for (const [key, val] of value) {
			parts.push(go2jsStringify(key) + ":" + go2jsStringify(val));
		}
		return "map[" + parts.join(" ") + "]";
	}
	if (Array.isArray(value)) {
		const parts = [];

		for (let i = 0; i < value.length; i++) {
			parts.push(go2jsStringify(value[i]));
		}

		return "[" + parts.join(" ") + "]";
	}
	if (typeof value === "object" && value !== null) {
		if (value.__go2js_json_delim !== undefined) {
			return String(value.__go2js_json_delim);
		}
		const parts = [];
		for (const key of Object.keys(value)) {
			parts.push(key + ":" + go2jsStringify(value[key]));
		}
		return "{" + parts.join(" ") + "}";
	}
	return String(value);
}

// go2jsCompareValues orders map keys the way Go's fmt package does.
function go2jsCompareValues(a, b) {
	if (typeof a === "number" && typeof b === "number") {
		return a < b ? -1 : a > b ? 1 : 0;
	}

	const left = go2jsFormat(a);
	const right = go2jsFormat(b);

	return left < right ? -1 : left > right ? 1 : 0;
}

// go2jsNumericValue coerces a value for the integer verbs without throwing on
// non numeric operands, because the switch below is shared by every verb.
// go2jsIsDuration reports whether a value is a span of time, which is a shape
// around a whole number of nanoseconds rather than a number of its own. The name
// is read first and on its own, since a box standing for a nil pointer answers
// only to the keys the runtime itself asks about.
function go2jsIsDuration(value) {
	return value !== null && value !== undefined && typeof value === "object" &&
		value.__go2js_type_name === "time.Duration";
}

function go2jsNumericValue(value) {
	if (value !== null && value !== undefined && value.__go2js_pointer === true) {
		return go2jsNumericValue(go2jsDeref(value));
	}

	// a span of time is asked for the whole number it holds, since that is what
	// a verb wanting digits writes down
	if (go2jsIsDuration(value)) {
		return Number(value.nanoseconds);
	}

	// a mark of a file is asked for the whole number of marks it holds, since
	// that is what a verb wanting digits writes down
	if (go2jsIsFileMode(value)) {
		return go2jsFileModeBits(value);
	}

	if (typeof value === "number") {
		return value;
	}

	if (typeof value === "bigint") {
		return Number(value);
	}

	if (typeof value === "string") {
		return Number(value);
	}

	return NaN;
}

function go2jsNilFormat(typeName, kind, shape) {
	// A nil map and a nil slice still print as their empty literal, while nil
	// pointers, channels and functions print as <nil>. A named type such as
	// Buffer carries its composite kind in shape instead of a bracket in its
	// name.
	if (shape === "map" || (typeof typeName === "string" && typeName.startsWith("map["))) {
		return "map[]";
	}

	if (shape === "slice" || (typeof typeName === "string" && typeName.startsWith("[]"))) {
		return "[]";
	}

	return "<nil>";
}

// go2jsFormat writes a value the way %v does, or the way %+v does when plus is
// set, which is the flag that puts a name in front of every field. The nested
// flag says the value sits inside another one, and that is what decides whether
// a pointer is written as an address or with a leading &. The raw flag says the
// value is being shown as the rejected operand of a verb it does not take, and
// fmt shows such an operand as it is rather than by asking it to write itself.
function go2jsFormat(value, typeName, kind, shape, plus, nested, raw) {
	if (typeof typeName !== "string") {
		typeName = go2jsTypedType(value);
	}

	if (typeof kind !== "string") {
		kind = go2jsTypedKind(value);
	}

	if (typeof shape !== "string") {
		shape = go2jsTypedShape(value);
	}

	value = go2jsUntyped(value);

	// A month and a weekday are numbers that are written down under a name, and
	// fmt writes the name rather than the number. Only the verbs that hand a
	// value over to what it says of itself do so, which is what the raw flag,
	// standing for %#v, says is not wanted here.
	if (raw !== true) {
		const calendar = go2jsCalendarName(typeName);

		if (calendar !== null && typeof value === "number") {
			return calendar(value);
		}

		// A number that is a type of its own is written as whatever that type
		// says it is, so that a level of a package says the name of a level rather
		// than the number behind one.
		if (typeof value === "number" && typeof typeName === "string" && typeName !== "") {
			const named = go2jsMethodTable[typeName + ".String"];

			if (typeof named === "function") {
				return go2jsFormat(go2jsCallNow(named, null, [value]));
			}
		}
	}

	// A number too big for a double is held as the whole number it is, and the
	// verbs are written against the number itself, so a plain number of that
	// kind is shown as the digits it is made of. Which verb is in play is worked
	// out further down, where the reading of the value it has is made, so this
	// only covers the verbs that are the same either way.
	if (typeof value === "bigint") {
		return go2jsBigintFormat(value, {verb: "v", width: 0, precision: -1, plus: false, space: false, zero: false, left: false, alt: false}, typeName, kind);
	}

	// A nil slice or a nil map is given an empty value of its own so its type
	// can travel, but it is still nil, and nil is what the verbs are written
	// against.
	if (value !== null && typeof value === "object" && value.__go2js_nil === true) {
		value = null;
	}

	if (value === null || value === undefined) {
		return go2jsNilFormat(typeName, kind, shape);
	}

	// A nil pointer held in an interface prints as the nil it is, the same as a
	// nil pointer anywhere else, even though the interface holding it is not
	// itself empty.
	if (value.__go2js_typed_nil === true) {
		return go2jsNilFormat("ptr", kind, shape);
	}

	// A value reflect made is a window onto a value rather than a value of its
	// own, so it is written as what the window holds, which is what a verb over
	// a reflect.Value shows in Go.
	if (value !== null && value !== undefined && value.__go2js_reflectValue === true) {
		return go2jsFormat(typeof value.get === "function" ? value.get() : value.v, null, null, null, plus, nested, raw);
	}

	if (value !== null && value !== undefined && value.__go2js_reflectType === true) {
		return typeof value.String === "function" ? value.String() : "";
	}

	if (value instanceof Error) {
		// An error stands for a struct holding its message, so showing the value
		// as it is means showing that struct behind the pointer fmt writes.
		return raw === true ? "&{" + value.message + "}" : value.message;
	}

	if (raw !== true && value.__go2js_interface !== true) {
		const receiver = value.__go2js_pointer === true ? go2jsDeref(value) : value;

		if (receiver !== null && receiver !== undefined && typeof receiver === "object" &&
			!Array.isArray(receiver) && !(receiver instanceof go2jsNativeMap)) {
			const named = go2jsNamedFormatMethod(receiver, "Error");

			if (named !== null) {
				return named;
			}

			const stringer = go2jsNamedFormatMethod(receiver, "String");

			if (stringer !== null) {
				return stringer;
			}
		}
	}

	// A slice or an array carries the wrapper on itself, so there is nothing
	// left to unwrap and it formats as the composite it is rather than as a
	// value pointing at itself.
	if ((value.__go2js_interface === true || value.__go2js_typed === true) && value.value !== value) {
		if (raw !== true && typeof value.type === "string") {
			const errorer = go2jsMethodTable[value.type + ".Error"];

			if (typeof errorer === "function") {
				return go2jsCallNow(errorer, null, [value.value]);
			}

			const stringer = go2jsMethodTable[value.type + ".String"];

			if (typeof stringer === "function") {
				return go2jsCallNow(stringer, null, [value.value]);
			}
		}

		return go2jsFormat(value.value, null, null, null, plus, nested, raw);
	}

	if (value.__go2js_pointer === true) {
		// fmt writes a pointer to a struct, an array, a slice or a map with a
		// leading &, but a pointer that sits inside a struct, an array, a slice
		// or a map is written as a bare address, because &{} there would be
		// ambiguous with the value it points at.
		if (nested || !go2jsPointerTargets(value)) {
			return go2jsPointerAddress(value);
		}

		return "&" + go2jsFormat(go2jsDeref(value), null, null, null, plus, false, raw);
	}


	// A complex operand is a pair of parts, so %v writes it in parentheses with
	// an i rather than as the struct its two fields would otherwise make.
	if (go2jsIsComplexTypeName(typeName)) {
		const { re, im } = go2jsComplexValue(value);

		return go2jsFormatComplexText("v", null, re, im, "", null);
	}

	if (value instanceof go2jsNativeMap) {
		// fmt sorts map keys, so the output must be deterministic.
		const entries = Array.from(value.entries());
		entries.sort((a, b) => go2jsCompareValues(a[0], b[0]));

		const parts = entries.map(([key, item]) => go2jsFormat(key, null, null, null, plus, true) + ":" + go2jsFormat(item, null, null, null, plus, true));

		return "map[" + parts.join(" ") + "]";
	}

	if (Array.isArray(value)) {
		const parts = [];

		for (let i = 0; i < value.length; i++) {
			parts.push(go2jsFormat(value[i], null, null, null, plus, true));
		}

		return "[" + parts.join(" ") + "]";
	}

	switch (typeof value) {
	case "string":
		return value;
	case "number":
		// A float is a float even when it holds a whole number, and Go writes
		// 1e+15 where a whole number would say 1000000000000000. Only the
		// integer types keep their digits, so that a large int64 stays
		// readable instead of turning into an exponent.
		if (go2jsIsFloatTypeName(typeName) || (kind !== null && kind !== undefined && kind !== "int" && kind !== "" && go2jsIsFloatTypeName(kind))) {
			return go2jsFormatFloatType(value, typeName, kind);
		}

		return Number.isInteger(value) ? String(value) : go2jsFormatFloatDefault(value);
	case "boolean":
		return String(value);
	}

	if (raw !== true && typeof value.String === "function") {
		return go2jsCallNow(value.String, value, []);
	}

	if (raw !== true && typeof value.Error === "function") {
		return go2jsCallNow(value.Error, value, []);
	}

	const parts = [];
	const ctor = value.constructor;
	const fields = ctor && typeof ctor.name === "string" ? go2jsStructFormats[ctor.name] : null;

	for (const key of Object.keys(value)) {
		if (fields && typeof fields[key] === "string") {
			const formatter = go2jsMethodTable[fields[key]];

			if (typeof formatter === "function") {
				parts.push(go2jsCallNow(formatter, null, [value[key]]));
				continue;
			}
		}

		const text = go2jsFormat(value[key], null, null, null, plus, true);

		parts.push(plus === true ? key + ":" + text : text);
	}

	return "{" + parts.join(" ") + "}";
}

function go2jsPrintln(...values) {
	go2jsWriteOut(process.stdout, go2jsJoinOperands(values, true) + "\n");
}

function go2jsJoinOperands(values, alwaysSpace) {
	let text = "";

	for (let i = 0; i < values.length; i++) {
		if (i > 0) {
			if (alwaysSpace || (!isStringOperand(values[i - 1]) && !isStringOperand(values[i]))) {
				text += " ";
			}
		}

		// Println and Print have no format to read a type from, so the type an
		// operand is boxed as is what a verb would have been given.
		text += go2jsFormat(values[i], go2jsTypedType(values[i]), go2jsTypedKind(values[i]), go2jsTypedShape(values[i]));
	}

	return text;
}

function isStringOperand(value) {
	value = go2jsUntyped(value);

	if (value === null || value === undefined) {
		return false;
	}

	if (value.__go2js_interface === true) {
		return typeof value.value === "string";
	}

	return typeof value === "string";
}

// Text on its way out is written as bytes, because a string is bytes and a
// byte that is not the UTF-8 of a rune has to arrive as itself.
function go2jsWriteOut(stream, text) {
	if (typeof text === "string" && go2jsHasRawBytes(text)) {
		stream.write(Buffer.from(go2jsStringToBytes(text)));
		return;
	}

	stream.write(text);
}

// go2jsOutputBytes is a run of bytes as the standard output takes it. Bytes that
// spell text are handed over as the text they spell, which the host writes far
// faster than a walk over every byte of it, and bytes that do not spell text are
// handed over as the bytes they are, since writing a byte that is not the UTF-8
// of a rune writes that byte and not the character that stands in for it.
function go2jsOutputBytes(value) {
	const bytes = Uint8Array.from(Array.from(go2jsToArray(value), item => Number(item) & 255));

	try {
		return go2jsStrictDecoder.decode(bytes);
	} catch (error) {
		return Buffer.from(bytes);
	}
}

function go2jsWriterMethod(writer, method) {
	const target = go2jsUnwrap(writer);

	if (target === null || target === undefined) {
		return null;
	}

	if (target === process.stdout || target === process.stderr || target === process.stdin) {
		return function(...args) {
			if (method === "WriteString") {
				target.write(go2jsBytesToString(args[0]));
				return go2jsStringify(args[0]).length;
			}

			target.write(go2jsOutputBytes(args[0]));
			return go2jsToArray(args[0]).length;
		};
	}

	if (typeof target[method] === "function") {
		const own = target[method];
		return function(...args) {
			return go2jsCallNow(own, target, args);
		};
	}

	const type = writer !== null && writer !== undefined && typeof writer.type === "string" ? writer.type : "";

	if (type !== "") {
		const registered = go2jsMethodTable[type + "." + method];
		if (typeof registered === "function") {
			return function(...args) {
				return go2jsCallNow(registered, null, [target, ...args]);
			};
		}
	}

	return null;
}

function go2jsFprint(writer, values, suffix, separator) {
	const write = go2jsWriterMethod(writer, "Write");

	if (write === null) {
		throw new Error("fmt.Fprint: writer does not implement Write");
	}

	const written = write(go2jsStringToBytes(go2jsOutputText(values, suffix, separator)));

	return Array.isArray(written) ? written[0] : written;
}

// go2jsSpreadValues unwraps the interface elements of a spread operand, which
// the emitter boxes so their dynamic type survives.
function go2jsSpreadValues(slice) {
	if (!Array.isArray(slice)) {
		return [slice];
	}

	return slice.map((element) => (element !== null && typeof element === "object" &&
		element.__go2js_interface === true ? element.value : element));
}

// go2jsSprintf writes a format and its operands, which is what Printf and
// Sprintf and friends do. A %w among them is turned down, because fmt only lets
// that verb wrap in the one place that says so.
function go2jsSprintf(format, ...args) {
	return go2jsFormatForPrinter(format, args, false);
}

function go2jsFormatForPrinter(format, args, wrapErrs) {
	let result = "";

	// A trailing ... hands over a marked slice; its elements become the operands
	// that follow the ones already given.
	if (args.length > 0 && args[args.length - 1] !== null &&
		typeof args[args.length - 1] === "object" && args[args.length - 1].__go2js_spread === true) {
		const last = args.pop();

		args = args.concat(go2jsSpreadValues(last.slice));
	}

	// The one based operand number the next verb will read, counting down as
	// fmt does: the first unindexed verb lands on operand one. claimed records
	// which operands the verbs have used, so the unused ones can be reported.
	let argIndex = 0;
	let walked = 0;
	let reached = 0;

	// fmt stops trusting the operand numbers once one of them names something
	// that is not there, and every verb left in the format says so.
	let goodArgNum = true;
	// Once a format names its operands, fmt gives up on saying which of the
	// operands went unused, because it cannot tell.
	let reordered = false;
	// True while the verb being read follows an index, which fmt only allows
	// when the width or the precision is read from an operand of its own.
	let afterIndex = false;

	for (let i = 0; i < format.length; i++) {
		const ch = format[i];

		if (ch !== "%") {
			result += ch;
			continue;
		}

		i++;
		let spec = "%";

		// An explicit index names the operand directly. It may stand in for the
		// width, for the precision or for the operand the verb reads, and the
		// place it was written in the format is what says which of them it is.
		let explicitIndex = 0;
		let starRead = false;
		// An index in front of the width has to stay there: digits or a dot
		// after it would leave fmt unable to tell who they belong to.
		let seenDot = false;
		// fmt reads its flags before the width and its precision, so a flag
		// written after one of them is the verb instead and whatever follows
		// it in the format is text.
		let seenWidth = false;

		while (i < format.length) {
			const flag = format[i];

			if (flag === "[" && format[i + 1] !== "]") {
				const close = format.indexOf("]", i + 1);

				if (close === -1) {
					break;
				}

				const digits = format.slice(i + 1, close);

				if (digits.length === 0 || !/^\d+$/.test(digits)) {
					break;
				}

				const named = Number(digits);
				reordered = true;

				if (named < 1 || named > args.length) {
					// fmt keeps the walk where it was and stops believing the
					// rest of the numbers, so every verb from the one here on
					// reports a bad index.
					goodArgNum = false;
				} else {
					explicitIndex = named;

					if (!seenDot) {
						afterIndex = true;
					}
				}

				i = close + 1;
				continue;
			}

			if ("+-# ".includes(flag)) {
				if (seenWidth) {
					break;
				}

				spec += flag;
				i++;
				continue;
			}

			// A width or a precision written out in digits has to come before
			// an index, because fmt has no way to tell which of the two the
			// operand belongs to. "%[1]6.2f" is a bad index, not a width.
			if (flag === "0" || (flag >= "1" && flag <= "9")) {
				if (afterIndex) {
					goodArgNum = false;
				}

				// A leading zero is the flag that pads with zeros; a digit after
				// that is the width itself, and the width ends the flags.
				if (flag !== "0") {
					seenWidth = true;
				}

				spec += flag;
				i++;
				continue;
			}

			if (flag === ".") {
				if (afterIndex) {
					goodArgNum = false;
				}

				seenDot = true;
				seenWidth = true;
				spec += flag;
				i++;
				continue;
			}

			// A "*" width or precision reads its operand from the cursor and
			// leaves the cursor on the next one, so the verb that follows reads
			// the operand after the star.
			if (flag === "*") {
				seenWidth = true;

				if (explicitIndex > 0) {
					argIndex = explicitIndex;
					explicitIndex = 0;
				} else {
					argIndex++;
				}

				if (argIndex < 1) {
					argIndex = 1;
				}

				// A width or a precision is read as an integer and nothing else,
				// so an operand of any other kind leaves the verb with none.
				const read = argIndex <= args.length
					? go2jsStarWidth(args[argIndex - 1])
					: null;

				if (read === null) {
					result += seenDot ? "%!(BADPREC)" : "%!(BADWIDTH)";
				}

				if (argIndex > args.length) {
					// The cursor stays where the star found nothing, so the verb
					// that follows is left looking past the end of the operands
					// as well and reports its own missing one.
					argIndex = args.length + 1;
					walked = 0;
				} else {
					// The verb that follows this star reads the operand after
					// the one the star used, so the cursor is left on it
					// already, whether or not the star made sense of it.
					walked = 1;
					reached = argIndex;
					starRead = true;
				}

				afterIndex = false;

				if (read === null) {
					// A precision that never arrived is not a precision, so the
					// dot in front of it goes too. A width simply never appears.
					if (seenDot) {
						spec = spec.replace(/\.$/, "");
					}
				} else if (read >= 0) {
					spec += String(read);
				}

				i++;
				continue;
			}

			break;
		}

		const verb = format[i];
		spec += verb;

		starRead = false;

		if (verb === "%") {
			result += "%";
			continue;
		}

		// fmt walks the operands with a cursor that holds the operand the next
		// verb reads. A verb with no index consumes the current one and moves
		// on, while an explicit index points the cursor at that operand and
		// leaves it there, so "%d %d" walks forwards and "%d %[1]d" repeats the
		// first operand.
		let namedOperand = false;

		if (explicitIndex > 0) {
			// An explicit index points the cursor at that operand. Naming one
			// also discards the walk, because fmt then has no way to know which
			// operands the unindexed verbs in between consumed.
			argIndex = explicitIndex;
			explicitIndex = 0;
			walked = 0;
			namedOperand = true;
		} else if (!starRead) {
			argIndex++;
			walked = 1;
		}

		if (argIndex < 1) {
			argIndex = 1;
		}

		// The reach follows the cursor rather than the highest index seen: an
		// explicit index that points back at an earlier operand rewinds it, so
		// "%d %[1]d" leaves the operands past the cursor looking unused again.
		reached = argIndex;

		// The rest parameter holds the operands only, so the one based operand
		// number N lives at args[N-1].
		const arg = args[argIndex - 1];

		// An operand past the end of the argument list is reported as missing,
		// the way fmt words the error. An index that names an operand which is
		// not there is a bad index instead, because fmt never walked to it.
		if (!goodArgNum) {
			result += "%!" + verb + "(BADINDEX)";
			argIndex = 1;
			afterIndex = false;
			continue;
		}

		if (argIndex > args.length) {
			result += "%!" + verb + "(" + (namedOperand ? "BADINDEX" : "MISSING") + ")";
			argIndex = 1;
			continue;
		}

		afterIndex = false;

		// A %w is written exactly as the %v beside it would be, because that is
		// all fmt does with it: the cause is taken from the operand and the text
		// still comes from the verb a Formatter is asked about. Handing the method
		// the w itself would report a verb fmt never passes on. It wraps only
		// where fmt says it may, and only around an error, so every other use is
		// a verb no operand takes.
		if (verb === "w" && !(wrapErrs === true && go2jsIsErrorOperand(arg))) {
			// A nil operand is written as the name of its type alone, with
			// nothing after an equals sign, because there is no value to write
			// there. Any other operand is written the way a verb no operand
			// takes is written: its type and its value beside one another.
			result += "%!w(" + go2jsDynamicTypeName(arg, go2jsTypedType(arg)) +
				(arg === null || arg === undefined
					? ""
					: "=" + go2jsFormatValue("v", spec.replace(/w$/, "v"), arg, true)) + ")";
		} else {
			result += verb === "w"
				? go2jsFormatValue("v", spec.replace(/w$/, "v"), arg)
				: go2jsFormatValue(verb, spec, arg);
		}
	}

	// fmt reports the operands the cursor never passed, unless the format named
	// its operands, in which case it cannot tell which ones went unused.
	if (walked > 0 && !reordered && reached < args.length) {
		const extra = args.slice(reached).map((value) => {
			const text = go2jsFormat(value, go2jsTypedType(value), go2jsTypedKind(value), go2jsTypedShape(value));

			return go2jsInferTypeName(value, go2jsTypedType(value)) + "=" + text;
		});

		result += "%!(EXTRA " + extra.join(", ") + ")";
	}

	return result;
}

// go2jsStarWidth reads an operand that stands in for a width or a precision.
// Only an integer counts, the way fmt reads it, and the answer is null when
// the operand is anything else, including a float that happens to be whole.
function go2jsStarWidth(operand) {
	const name = go2jsInferTypeName(operand, go2jsTypedType(operand));

	if (!go2jsIsIntegerTypeName(name)) {
		return null;
	}

	const value = Number(go2jsUntyped(operand));

	if (!Number.isInteger(value) || Math.abs(value) > Number.MAX_SAFE_INTEGER) {
		return null;
	}

	return value;
}

// go2jsStripWrappers reaches the value a chain of wrappers stands for. Retyping
// an already wrapped value has to go all the way in, because a wrapper whose
// own value is a wrapper of itself sends every walk of it round in circles.
function go2jsStripWrappers(value) {
	let current = value;

	while (current !== null && typeof current === "object" && (current.__go2js_typed === true || current.__go2js_interface === true)) {
		// A slice or an array is its own interface wrapper, so following its
		// value once more would come back to where this started.
		if (current.value === current) {
			break;
		}

		current = current.value;
	}

	return current;
}

// go2jsIsInterfaceTypeName reports whether a name is one of the ways the emitter
// spells an empty interface, which carries no name of its own.
function go2jsIsInterfaceTypeName(name) {
	return name === "" || name === "any" || name === "interface{}" || name === "interface {}";
}

function go2jsTyped(value, type, kind, shape) {
	// A span of time already carries the name of its own type and a way of
	// writing itself out, so wrapping it would bury both of them.
	if (go2jsIsDuration(value)) {
		return value;
	}

	// A value that already stands as the type it is being given again keeps the
	// way of writing itself out that came with it.
	if (value !== null && value !== undefined && typeof value === "object" && value.__go2js_typed === true && value.type === type) {
		return value;
	}

	// A nil slice or a nil map is nil, and the empty value it carries says only
	// what it is, so the name asked for here is the one that gets written down.
	if (value !== null && typeof value === "object" && value.__go2js_nil === true) {
		const named = typeof value.type === "string" && value.type !== "" ? value.type : type;
		const shaped = typeof shape === "string" && shape !== "" ? shape : value.shape;

		return {__go2js_typed: true, value: null, type: named, shape: shaped};
	}

	if (value !== null && typeof value === "object" && (value.__go2js_typed === true || value.__go2js_interface === true)) {
		// An interface carries no name of its own, so a type that arrived with
		// the value says more than the name of the interface it sits in.
		if (go2jsIsInterfaceTypeName(type)) {
			if (typeof value.type === "string" && value.type !== "") {
				return value;
			}
		}

		value = go2jsStripWrappers(value);
	} else if (go2jsIsInterfaceTypeName(type)) {
		// A value that leaves through an interface is reported by the type it
		// really has, and a value that carries no tag of its own has to be read
		// for it, because the name of the interface says only where it sits.
		const named = go2jsGoTypeName(value);

		if (typeof named === "string" && named !== "" && !go2jsIsInterfaceTypeName(named)) {
			type = named;
		}
	}

	const wrapper = {__go2js_typed: true, value: value, type: type};

	if (typeof kind === "string" && kind !== "") {
		wrapper.kind = kind;
	}

	// The shape records the composite kind of a named type, so a nil Buffer
	// still prints as [] rather than <nil>.
	if (typeof shape === "string" && shape !== "") {
		wrapper.shape = shape;
	}

	// A type that says how it is written out has that written on the box, so a
	// method called on the value reaches the same answer printing it would give.
	if (typeof type === "string" && type !== "") {
		const stringer = go2jsMethodTable[type + ".String"];

		if (typeof stringer === "function") {
			wrapper.String = function() {
				return go2jsCallNow(stringer, null, [this.value]);
			};
		}
	}

	return wrapper;
}

function go2jsTypedShape(value) {
	if (value !== null && value !== undefined && value.__go2js_typed === true) {
		return value.shape;
	}

	return null;
}

function go2jsUntyped(value) {
	if (value !== null && value !== undefined && value.__go2js_typed === true) {
		return value.value;
	}

	return value;
}

function go2jsTypedType(value) {
	if (value !== null && value !== undefined && value.__go2js_typed === true) {
		return value.type;
	}

	return null;
}

function go2jsTypedKind(value) {
	if (value !== null && value !== undefined && value.__go2js_typed === true &&
		typeof value.kind === "string") {
		return value.kind;
	}

	return null;
}

// go2jsFormatHexBytes writes bytes as two hex digits each. The space flag puts
// a space between them, the way Go writes "% x" over a string or a byte slice.
function go2jsFormatHexBytes(value, upper, space) {
	const parts = [];

	for (const item of go2jsToArray(value)) {
		parts.push((Number(item) & 255).toString(16).padStart(2, "0"));
	}

	const text = parts.join(space ? " " : "");

	return upper ? text.toUpperCase() : text;
}

function go2jsIsByteArray(value) {
	if (!Array.isArray(value) || value.length === 0) {
		return false;
	}

	for (const item of value) {
		if (typeof item !== "number" || !Number.isInteger(item) || item < 0 || item > 255) {
			return false;
		}
	}

	return true;
}

function go2jsInferTypeName(value, tagged) {
	if (typeof tagged === "string") {
		return tagged;
	}

	if (value === null || value === undefined) {
		return "<nil>";
	}

	if (value.__go2js_interface === true) {
		return typeof value.type === "string" ? value.type : "interface {}";
	}

	if (value instanceof Error) {
		return "error";
	}

	if (typeof value === "string") {
		return "string";
	}

	if (typeof value === "boolean") {
		return "bool";
	}

	if (typeof value === "number") {
		return Number.isInteger(value) ? "int" : "float64";
	}

	if (go2jsIsByteArray(value)) {
		return "[]uint8";
	}

	if (Array.isArray(value)) {
		return "[]int";
	}

	if (typeof value === "object") {
		const name = go2jsGoTypeName(value);

		if (typeof name === "string" && name !== "" && name !== "object") {
			return name;
		}

		return "struct {}";
	}

	return "interface {}";
}

// go2jsIsIntegerTypeName reports whether a Go type name is one of the integer
// kinds, which includes rune and byte.
function go2jsIsIntegerTypeName(typeName) {
	switch (typeName) {
	case "int":
	case "int8":
	case "int16":
	case "int32":
	case "int64":
	case "uint":
	case "uint8":
	case "uint16":
	case "uint32":
	case "uint64":
	case "uintptr":
	case "rune":
	case "byte":
		return true;
	default:
		return false;
	}
}

// go2jsIsFloatTypeName reports whether a Go type name is one of the float kinds.
function go2jsIsFloatTypeName(typeName) {
	return typeName === "float32" || typeName === "float64";
}

// go2jsIsComplexTypeName reports whether a Go type name is one of the complex
// kinds.
function go2jsIsComplexTypeName(typeName) {
	return typeName === "complex64" || typeName === "complex128";
}

// go2jsVerbAccepts reports whether a verb is defined for a Go type. The table
// below lists, for each operand kind, the verbs fmt refuses; a verb that is not
// named accepts everything. A compound operand is never refused as a whole,
// because fmt expands it and then formats each element with the verb.
function go2jsVerbAccepts(verb, typeName) {
	if (!go2jsIsBasicScalarName(typeName)) {
		return true;
	}

	if (typeName === "<nil>") {
		// A nil operand only has %v and %T, so every other verb is a bad verb.
		return verb === "v" || verb === "T";
	}

	if (go2jsIsIntegerTypeName(typeName)) {
		return go2jsVerbInSet(verb, "bcdoOqxUXvT");
	}

	if (go2jsIsFloatTypeName(typeName) || go2jsIsComplexTypeName(typeName)) {
		return go2jsVerbInSet(verb, "befgEFGXxXvT");
	}

	if (typeName === "string") {
		return go2jsVerbInSet(verb, "qsxXvT");
	}

	if (typeName === "bool") {
		return go2jsVerbInSet(verb, "tvT");
	}

	return verb === "v" || verb === "T";
}

// go2jsVerbInSet reports whether a one character verb appears in a set of
// accepted verbs.
function go2jsVerbInSet(verb, accepted) {
	return verb !== "" && accepted.indexOf(verb) >= 0;
}

// go2jsIsBasicScalarName reports whether a Go type name is one fmt formats
// directly, as opposed to a compound operand that fmt expands first. A string
// and a nil operand are both formatted directly, and a verb that does not fit
// either of them is a bad verb.
function go2jsIsBasicScalarName(typeName) {
	switch (typeName) {
	case "bool":
	case "int":
	case "int8":
	case "int16":
	case "int32":
	case "int64":
	case "uint":
	case "uint8":
	case "uint16":
	case "uint32":
	case "uint64":
	case "uintptr":
	case "float32":
	case "float64":
	case "complex64":
	case "complex128":
	case "rune":
	case "byte":
	case "string":
	case "<nil>":
		return true;
	default:
		return false;
	}
}

function go2jsGoSyntax(value, typeName) {
	// A value held in an interface is written out as the value it holds, named
	// by the type it was held as, since the wrapper itself is not written.
	if (value !== null && value !== undefined &&
		(value.__go2js_typed === true || value.__go2js_interface === true)) {
		const held = value.value;

		if (go2jsDeclaredTypeName(typeName) === "" &&
			typeof value.type === "string" && value.type !== "") {
			typeName = value.type;
		}

		return go2jsGoSyntax(held, typeName);
	}

	if (value === null || value === undefined) {
		// A value held in an interface has no value of its own when the value it
		// holds is nothing, so what is written is the type it was held as rather
		// than the nothing itself.
		const declared = go2jsDeclaredTypeName(typeName);

		return declared === "" ? "<nil>" : declared + "(nil)";
	}

	if (value.__go2js_pointer === true) {
		return "(" + go2jsGoTypeName(value) + ")(" + go2jsPointerAddress(value) + ")";
	}

	if (value instanceof Error) {
		return value.message;
	}

	if (typeof value === "string") {
		return JSON.stringify(value);
	}

	if (typeof value === "number" || typeof value === "boolean") {
		return String(value);
	}

	if (value instanceof go2jsNativeMap) {
		const entries = Array.from(value.entries());
		entries.sort((a, b) => go2jsCompareValues(a[0], b[0]));

		// The name of a map already says what it holds keys and values of, so it
		// is written out as it stands rather than wrapped in brackets again.
		return go2jsCollectionTypeName(value, typeName) +
			"{" + entries.map(([key, item]) => go2jsGoSyntax(key) + ":" + go2jsGoSyntax(item)).join(", ") + "}";
	}

	if (Array.isArray(value)) {
		return go2jsCollectionTypeName(value, typeName) +
			"{" + value.map(item => go2jsGoSyntax(item)).join(", ") + "}";
	}

	const ctor = value.constructor;
	const name = ctor && typeof ctor.name === "string" ? ctor.name : "";
	const fields = name !== "" ? go2jsStructFormats[name] : null;
	const parts = [];

	for (const key of Object.keys(value)) {
		const fieldType = fields && typeof fields[key] === "string" ? fields[key] : "";
		parts.push(key + ":" + go2jsGoSyntax(value[key]));

		if (fieldType === "") {
			continue;
		}
	}

	return go2jsQualifiedTypeName(go2jsGoTypeName(value)) + "{" + parts.join(", ") + "}";
}

// go2jsCollectionTypeName is the name a list or a map is written out under. A
// value held in an interface carries the type it was held as, which is the one
// to write, while a value of its own is named by what it holds.
function go2jsCollectionTypeName(value, typeName) {
	const declared = go2jsDeclaredTypeName(typeName);

	return declared === "" ? go2jsGoTypeName(value) : declared;
}

// go2jsDeclaredTypeName is the type a value was declared as, or nothing when
// what is offered names nothing in particular. An interface names no type of its
// own, and a value the host holds in a slot of its own is given a name that says
// where it sits rather than what it is.
function go2jsDeclaredTypeName(name) {
	if (typeof name !== "string") {
		return "";
	}

	if (name === "" || name === "interface {}" || name === "null" ||
		name === "undefined" || name === "unknown" || name === "Object" || name === "Array") {
		return "";
	}

	return name;
}

// go2jsCalendarName is the helper that writes a month or a weekday under the
// name it is known by, or nothing when the type is neither of them.
function go2jsCalendarName(typeName) {
	if (typeName === "time.Month") {
		return go2jsMonthName;
	}

	if (typeName === "time.Weekday") {
		return go2jsWeekdayName;
	}

	return null;
}

function go2jsQualifiedTypeName(name) {
	if (name === "" || name === "Object" || name === "Array") {
		return "";
	}

	return name;
}

const go2jsRuneShortEscapes = {
	0x07: "\\a", 0x08: "\\b", 0x0C: "\\f", 0x0A: "\\n",
	0x0D: "\\r", 0x09: "\\t", 0x0B: "\\v"
};

// go2jsRuneIsPrintable mirrors the rule strconv uses to decide whether a rune
// can be shown literally or has to be escaped. Only ASCII is printable as a
// whole; above that a rune counts only when it belongs to a letter, mark,
// number, punctuation or symbol category, which keeps the space separators and
// the format characters escaped exactly as Go escapes them.
function go2jsRuneIsPrintable(code) {
	if (code === 0x20) {
		return true;
	}

	if (code < 0x20 || code === 0x7F) {
		return false;
	}

	if (code < 0x7F) {
		return true;
	}

	const category = (String.fromCodePoint(code) || "").replace(/[\p{L}\p{M}\p{N}\p{P}\p{S}]/u, "");

	return category.length === 0;
}

// go2jsQuoteRuneBody renders a rune the way strconv.QuoteRune does, using the
// short escapes Go prefers before falling back to \u.
function go2jsQuoteRuneBody(code) {
	if (code < 0 || code > 0x10FFFF || (code >= 0xD800 && code <= 0xDFFF) ||
		!go2jsRuneIsPrintable(code)) {
		const short = go2jsRuneShortEscapes[code];

		if (short !== undefined) {
			return short;
		}

		// strconv escapes an ASCII control as \xNN, anything else below the
		// supplementary planes as \uNNNN and the rest as \UNNNNNNNN, always
		// with lower case hex digits.
		if (code < 0x20 || code === 0x7F) {
			return "\\x" + code.toString(16).padStart(2, "0");
		}

		if (code > 0xFFFF) {
			return "\\U" + code.toString(16).padStart(8, "0");
		}

		return "\\u" + code.toString(16).padStart(4, "0");
	}

	const ch = String.fromCodePoint(code);

	// Go escapes the quote and the backslash so the result stays unambiguous.
	if (ch === "'" || ch === "\\") {
		return "\\" + ch;
	}

	return ch;
}

function go2jsQuoteRune(code) {
	return "'" + go2jsQuoteRuneBody(code) + "'";
}

// go2jsIsGoStructType reports whether a resolved Go type name denotes a struct,
// which is the only shape fmt expands into a braced field list for %q.
function go2jsIsGoStructType(typeName) {
	if (typeof typeName !== "string" || typeName === "") {
		return false;
	}

	return typeName.startsWith("struct") || /^[A-Za-z_][A-Za-z0-9_.]*\.[A-Za-z_][A-Za-z0-9_]*$/.test(typeName) ||
		typeName.includes("struct {");
}

// go2jsStructHasNoStringMethod keeps a struct that defines String from being
// expanded field by field: fmt quotes the result of the method instead.
function go2jsStructHasNoStringMethod(value) {
	return typeof value.String !== "function" && typeof value.Error !== "function";
}

// The flags fmt.State reports on, in the order Go numbers them, because a
// Format method compares against those numbers and a translation cannot renumber
// them without changing what the method it is calling asks for.
// go2jsSelfWrittenText returns what an error or a Stringer would write for a
// value, or nothing when the value is neither. The value behind an interface box
// or a pointer is the one the method belongs to, because the box and the
// pointer are not the type that carries it.
function go2jsSelfWrittenText(operand, value) {
	// A value boxed as an interface carries the name of the type it holds, and
	// that name is where a method of a type the emitter wrote as a number or a
	// string is registered, because such a value has nowhere of its own to keep
	// it.
	if (operand !== null && operand !== undefined && operand.__go2js_interface === true &&
		typeof operand.type === "string") {
		for (const method of ["Error", "String"]) {
			const registered = go2jsMethodTable[operand.type + "." + method];

			if (typeof registered === "function") {
				return go2jsBytesToString(go2jsCallNow(registered, null, [operand.value]));
			}
		}
	}

	let target = operand;

	if (target !== null && target !== undefined && target.__go2js_interface === true) {
		target = target.value;
	}

	if (target !== null && target !== undefined && target.__go2js_pointer === true) {
		target = go2jsDeref(target);
	}

	if (target === null || target === undefined) {
		return null;
	}

	const written = go2jsNamedFormatMethod(target, "Error") ?? go2jsNamedFormatMethod(target, "String");

	return written === null ? null : go2jsBytesToString(written);
}

// go2jsDynamicTypeName names a value by the type it really is rather than by the
// interface it was boxed as, which is the name fmt writes into the text of a
// verb that was refused.
function go2jsDynamicTypeName(value, tagName) {
	if (value !== null && value !== undefined && value.__go2js_interface === true) {
		value = value.value;
	}

	return go2jsGoTypeName(value);
}

// go2jsIsErrorOperand reports whether a value is an error, which is the only
// thing a %w has a cause to take from.
function go2jsIsErrorOperand(value) {
	if (value === null || value === undefined) {
		return false;
	}

	if (value.__go2js_interface === true) {
		return go2jsIsErrorOperand(value.value);
	}

	// The Error method sits on the type a pointer points at rather than on the
	// pointer, so the pointer is followed to reach it.
	if (value.__go2js_pointer === true) {
		return go2jsIsErrorOperand(go2jsDeref(value));
	}

	return value instanceof Error || go2jsNamedFormatMethod(value, "Error") !== null;
}

// go2jsFmtFlags lists the flag characters a verb may be written with. A method
// asking about a flag asks with the character itself, the way fmt reads it, so
// the answer is about the character rather than about a bit of its own.
const go2jsFmtFlags = ["-", "+", "#", " ", "0"];

// go2jsFormatSelf hands a value to the Format method of its own type, which is
// what a type that carries a Format method means by it. The state it is given
// is what fmt would pass, and what it writes is what the verb produces. A value
// with no such method reports nothing, which sends the verb on to its own case.
function go2jsFormatSelf(value, verb, flags, width, precision) {
	const receiver = go2jsUntyped(value);

	if (receiver === null || receiver === undefined || go2jsIsTypedNilPointer(receiver)) {
		return null;
	}

	let format = go2jsLookupNamedMethod(receiver, "Format");

	if (format === null && receiver.__go2js_interface === true && typeof receiver.type === "string") {
		const method = go2jsMethodTable[receiver.type + ".Format"];

		if (typeof method === "function") {
			format = (...args) => go2jsCallNow(method, null, [receiver.value, ...args]);
		}
	}

	if (format === null) {
		return null;
	}

	const state = go2jsFmtState(verb, flags, width, precision, receiver);

	// Go hands a Format method the verb as the rune it stands for, and a method
	// that writes it into its own output does so as a code point, so it is
	// passed as the number the letter is rather than as the letter.
	const rune = String(verb).codePointAt(0);

	try {
		go2jsCallNow(format, receiver, [state, rune]);
	} catch (error) {
		// A Format method that gives up is reported the way fmt reports a panic
		// inside one, rather than being allowed to end the program.
		const text = error !== null && error !== undefined && error.message !== undefined
			? String(error.message)
			: String(error);

		return "%!" + verb + "(PANIC=Format method: " + text + ")";
	}

	return go2jsBytesToString(state.buffer);
}

// go2jsFmtState builds the state fmt hands to a Format method. It answers the
// three questions the method asks about how it was called, and takes the bytes
// the method writes, which are the bytes the verb stands for.
function go2jsFmtState(verb, flags, width, precision, value) {
	const written = go2jsFmtFlags.filter(flag => flags.includes(flag));
	const buffer = [];

	const state = {
		__go2js_fmtState: true,
		verb: verb,
		buffer: buffer,
		Flag: function(flag) {
			return written.includes(typeof flag === "string" ? flag : String.fromCharCode(flag));
		},
		// Width and Precision report whether they were written at all, which is
		// the second value a method that asks for them is given.
		Width: function() {
			return typeof width === "number" ? [width, true] : [0, false];
		},
		Precision: function() {
			return typeof precision === "number" ? [precision, true] : [0, false];
		},
		// fmt hands a Format method its output as bytes, so what arrives is
		// taken as the byte values it already is. A method that writes a string
		// instead is given the same reading, the way it is in Go.
		Write: function(bytes) {
			const written = typeof bytes === "string" ? go2jsStringToBytes(bytes) : Array.from(bytes);

			for (const byte of written) {
				buffer.push(Number(byte) & 255);
			}

			return [written.length, null];
		}
	};

	// fmt writes to a state through the io.Writer it is, so a method that holds
	// it as one reaches Write without knowing what it is.
	if (value !== null && value !== undefined) {
		state.value = value;
	}

	return state;
}

// go2jsFormatValue writes one operand with a verb. The raw flag says the operand
// is being shown as the rejected operand of a verb it does not take, and fmt
// shows such an operand as it is rather than by asking it to write itself.
function go2jsFormatValue(verb, spec, value, raw) {
	const parsed = go2jsParseFormatSpec(spec);
	const flags = parsed.flags;
	const precision = parsed.precision;
	const tagged = go2jsTypedType(value);
	const original = value;

	// An operand boxed as an interface is written as what it holds, and the type
	// it holds is the type the verb is asked about. The box is kept alongside,
	// because a String or an Error method is reached through the type the box
	// names rather than through the value it wraps.
	const operand = value;
	const boxed = operand !== null && operand !== undefined &&
		(operand.__go2js_interface === true || operand.__go2js_typed === true);
	const boxedType = boxed && typeof operand.type === "string" ? operand.type : null;

	value = go2jsUnwrap(go2jsUntyped(value));
	value = go2jsMaterializeValue(value);

	// A value reflect made is a window onto a value rather than a value of its
	// own, so a verb is answered by what the window holds, and %T names the
	// window rather than what it looks through.
	if (value !== null && value !== undefined && value.__go2js_reflectValue === true) {
		if (verb === "T") {
			return "reflect.Value";
		}

		value = typeof value.get === "function" ? value.get() : value.v;
		value = go2jsMaterializeValue(go2jsUnwrap(go2jsUntyped(value)));
	} else if (verb !== "T" && value !== null && value !== undefined && value.__go2js_reflectType === true) {
		// every value of a type is the one type behind them all, and %T is the
		// one verb that asks what a value is rather than what it says, so it is
		// answered above and not here
		return typeof value.String === "function" ? value.String() : "";
	}

	const kind = go2jsTypedKind(original);
	const shape = go2jsTypedShape(original);
	const tagName = tagged || boxedType;

	// A value of a type that carries a whole number of its own, such as the Int of
	// math/big, is written as the number it holds rather than as the holder that
	// holds it, because such a type answers for itself and counts the number in
	// the base the verb names. A verb written for a string is counted in ten like
	// the rest, since a type that answers for itself writes the number whichever
	// way it is asked about it, and a fault is named after the type rather than
	// after the pointer it is reached through, which is how Go names it.
	if (value !== null && value !== undefined && value.__go2js_whole_number === true &&
		typeof value.value === "bigint") {
		const answered = verb === "s" ? "d" : (verb === "O" ? "o" : verb);
		// %T asks what the value is rather than what it says, and it is reached
		// through the pointer it is held by, so the pointer is part of the name.
		const named = verb === "T" || typeof tagName !== "string" || tagName === ""
			? tagName
			: tagName.replace(/^\*/, "");

		return go2jsBigintFormat(value.value,
			{verb: answered, flags: flags, precision: precision, width: parsed.width},
			named, kind);
	}

	// A number too big for a double is held whole, and every verb over it reads
	// the digits off that whole number rather than off a rounded one, so it is
	// answered before anything tries to treat it as a plain number.
	if (typeof value === "bigint") {
		return go2jsBigintFormat(value, {verb: verb, flags: flags, precision: precision, width: parsed.width}, tagName, kind);
	}

	// A month and a weekday are numbers that are written down under a name, and
	// fmt hands them to what they say of themselves rather than reading them as
	// the whole numbers they are, so the verbs that accept a string are answered
	// here rather than refused against the number underneath.
	const calendar = go2jsCalendarName(tagName);

	if (calendar !== null && typeof value === "number" && go2jsVerbInSet(verb, "vsq")) {
		if (verb === "q") {
			return go2jsPad(JSON.stringify(calendar(value)), parsed, false);
		}

		if (verb === "s") {
			return go2jsPad(calendar(value), parsed, false);
		}
	}

	const accepted = kind || go2jsInferTypeName(value, tagName);

	if (verb !== "%" && !go2jsVerbAccepts(verb, accepted)) {
		// A nil operand has no value to show, so fmt names the type alone.
		if (accepted === "<nil>") {
			return "%!" + verb + "(<nil>)";
		}

		// fmt shows the rejected operand with %v while keeping the width and the
		// precision the verb was written with, and without asking it to write
		// itself, because an operand reporting a fault is not the place to look
		// for it. A rune is named int32, because that is the type fmt reflects
		// on.
		return "%!" + verb + "(" + (accepted === "rune" ? "int32" : accepted) + "=" +
			go2jsFormatValue("v", spec.replace(verb + "$", "v"), value, true) + ")";
	}

	// A type that formats itself is asked to, and what it writes is what the
	// verb produces. It goes before the verb's own case, because a type that
	// takes the work over has the last word on how it is written. A width or a
	// precision that was not written is passed as nothing rather than as zero,
	// because a method asks for them to find out whether they were given.
	const selfFormatted = raw === true
		? null
		: go2jsFormatSelf(
			original,
			verb,
			flags,
			parsed.width > 0 ? parsed.width : null,
			precision
		);

	if (selfFormatted !== null) {
		return selfFormatted;
	}

	// An error or a Stringer is asked to write itself for the verbs that can
	// read a string at all, which is the set fmt consults them for. Every other
	// verb is answered by the value itself, so %d on a named integer is the
	// number and not what its String method would have said.
	if (raw !== true && go2jsVerbInSet(verb, "vsxXq")) {
		const text = go2jsSelfWrittenText(operand, value);

		if (text !== null) {
			return go2jsFormatValue(verb, spec, text);
		}
	}

	// Renders the sign prefix and zero-pads the digits that follow it. A body
	// written in hex for a float brings its own sign with it, and bodyHasSign
	// says so, because the sign is only added once.
	const numberText = (num, body, prefixOverride, bodyHasSign) => {
		const negative = num < 0 || Object.is(num, -0);
		const sign = negative
			? "-"
			: flags.includes("+")
				? "+"
				: flags.includes(" ")
					? " "
					: "";

		// The sign comes first even when a base prefix is written, so %# x on
		// 255 reads " 0xff" rather than "0x ff".
		const prefix = prefixOverride === undefined
			? sign
			: (bodyHasSign === true && negative ? "" : sign) + prefixOverride;

		// A precision on an integer is a minimum number of digits, and the zeros
		// go in before the base prefix, the way Go writes 0x00ff. A float spends
		// its precision on the fraction it was already rendered with, so the
		// digits are left as they are.
		const isInteger = go2jsIsIntegerTypeName(accepted);
		const digits = precision !== null && isInteger ? go2jsZeroPadDigits(body, precision) : body;

		return go2jsPadNumber(prefix + digits, prefix, digits, parsed, isInteger);
	};

	const integerBody = (num, base, prefix, upper) => {
		const magnitude = Math.abs(Math.trunc(num));
		let body = magnitude.toString(base);

		if (upper) {
			body = body.toUpperCase();
		}

		// The width counts the base prefix, so a prefixed body is padded through
		// its full length rather than the digits alone.
		return numberText(num, body, prefix === "" ? undefined : prefix);
	};

	const intValue = Math.trunc(go2jsNumericValue(value));

	// A slice or an array applies the verb to each element and joins the results,
	// so it is answered before the verb's own case sees the whole value. A byte
	// slice is the exception: the quoted verbs and the hex verbs read it as a
	// string rather than as a list of numbers.
	const isBytes = go2jsIsByteCompound(accepted);
	// The sharp flag with %v asks for the value as Go writes it, which is the
	// whole of it with its type in front rather than the elements one by one.
	const compound = verb === "T" || verb === "p" || (verb === "v" && flags.includes("#")) ||
		(isBytes && go2jsVerbInSet(verb, "qsxX"))
		? null
		: go2jsCompoundElements(value, accepted);

	if (compound !== null) {
		const inner = "%" + flags + (precision === null ? "" : "." + precision) + verb;
		// An element keeps the element type of the compound, so a verb that is
		// refused reports uint8 rather than the int a bare number would infer.
		const element = go2jsCompoundElementType(accepted);
		const parts = compound.map((item) => go2jsFormatValue(verb, inner, element === null ? item : go2jsTyped(item, element)));

		return go2jsPad("[" + parts.join(" ") + "]", parsed, false);
	}

	// A complex operand formats each of its parts with the verb and joins them
	// with a plus, the way fmt writes (re+imi), so it is answered first.
	if (go2jsIsComplexOperand(value, accepted)) {
		return go2jsFormatComplex(verb, parsed, value, accepted, flags, precision);
	}

	switch (verb) {
		case "d":
			return numberText(intValue, String(Math.abs(intValue)));
		case "b":
			// A float formatted with %b is its IEEE754 expansion: the leading
			// bit of the significand, then the exponent as a power of two.
			if (accepted !== undefined && go2jsIsFloatTypeName(accepted)) {
				return go2jsFormatFloatBits(Number(value), accepted);
			}

			return integerBody(intValue, 2, flags.includes("#") ? "0b" : "");
		case "o":
			return integerBody(intValue, 8, flags.includes("#") ? "0" : "");
		case "O":
			// %O always carries the 0o prefix, while %o only does with #.
			return integerBody(intValue, 8, "0o");
		case "x":
			if (accepted !== undefined && go2jsIsFloatTypeName(accepted)) {
				return numberText(intValue, go2jsFormatHexFloat(Number(value), precision, false), "", true);
			}

			if ((Array.isArray(value) || value instanceof Uint8Array) && !Number.isInteger(Number(value))) {
				return go2jsFormatHexBytes(value, false, flags.includes(" "));
			}

			if (typeof value === "string" || typeof value === "boolean") {
				return go2jsFormatHexBytes(go2jsStringToBytes(value), false, flags.includes(" "));
			}

			return integerBody(intValue, 16, flags.includes("#") ? "0x" : "", false);
		case "X":
			if (accepted !== undefined && go2jsIsFloatTypeName(accepted)) {
				return numberText(intValue, go2jsFormatHexFloat(Number(value), precision, true), "", true);
			}

			if ((Array.isArray(value) || value instanceof Uint8Array) && !Number.isInteger(Number(value))) {
				return go2jsFormatHexBytes(value, true, flags.includes(" "));
			}

			if (typeof value === "string" || typeof value === "boolean") {
				return go2jsFormatHexBytes(go2jsStringToBytes(value), true, flags.includes(" "));
			}

			return integerBody(intValue, 16, flags.includes("#") ? "0X" : "", true);
		case "f":
		case "F": {
			// %F is %f spelled in upper case; it has no exponent form to fold, so
			// the two share everything but the letter.
			const num = Number(value);
			const special = go2jsSpecialFloat(num, parsed);

			if (special !== "") {
				return go2jsPad(special, parsed, false);
			}

			return numberText(num, go2jsFormatFixed(Math.abs(num), precision === null ? 6 : precision, flags.includes("#")));
		}
		case "e":
		case "E": {
			const num = Number(value);
			const special = go2jsSpecialFloat(num, parsed);

			if (special !== "") {
				return go2jsPad(special, parsed, false);
			}

			if (verb === "e") {
				return go2jsFormatE(num, precision === null ? 6 : precision, parsed);
			}

			const upper = go2jsFormatE(num, precision === null ? 6 : precision, parsed);
			const exponent = upper.indexOf("e");

			return exponent === -1 ? upper : upper.slice(0, exponent) + "E" + upper.slice(exponent + 1);
		}
		case "g":
			return go2jsFormatG(Number(value), precision, parsed, false);
		case "G":
			return go2jsFormatG(Number(value), precision, parsed, true);
		case "s": {
			let text;

			if (go2jsIsByteCompound(tagged) || (tagged === null && go2jsIsByteArray(value))) {
				text = go2jsBytesToString(value);
			} else {
				text = go2jsFormat(value, tagged, kind, shape);
			}

			if (precision !== null) {
				text = text.slice(0, precision);
			}

			return go2jsPad(text, parsed, false);
		}
		case "v":
		case "w": {
			let text;
			// A sign flag asks for a sign in front of a number that has none of
			// its own to show, so it only applies when the text rendered is the
			// number itself. A negative one already carries a minus, and a value
			// that is not a number, a string or a struct say, is left alone.
			let sign = "";

			if (flags.includes("#")) {
				text = go2jsGoSyntax(value, boxedType);
			} else if (flags.includes("+") && !go2jsHasFormatMethod(operand)) {
				text = go2jsFormatFields(value);
			} else {
				text = go2jsFormat(operand, tagged || boxedType, kind, shape, false, false, raw);

				if (typeof value === "number" && /^[0-9]/.test(text) && flags.includes(" ")) {
					sign = " ";
				}
			}

			// A precision means a minimum number of digits for an integer, a
			// truncation for a string, and nothing at all for the other kinds.
			const isInteger = go2jsIsIntegerTypeName(accepted);

			if (precision !== null) {
				if (accepted === "string") {
					text = text.slice(0, precision);
				} else if (isInteger && go2jsDigitsOf(text) < precision) {
					text = go2jsZeroPadDigits(text, precision);
				}
			}

			// A sign stands outside the width, so the width and a padding of
			// zeros are counted from the digits alone.
			return sign === ""
				? go2jsPad(text, parsed, isInteger)
				: go2jsPadNumber(sign + text, sign, text, parsed, isInteger);
		}
		case "q": {
			// A nil slice or a nil map still prints as its empty literal under
			// the quoted verbs, but a nil byte slice is quoted like the empty
			// string it is, and a nil pointer keeps its <nil>.
			if (value === null || value === undefined) {
				if (go2jsIsByteCompound(accepted)) {
					return go2jsPad(go2jsStrconvQuote(""), parsed, false);
				}

				const nilText = go2jsNilFormat(accepted, kind, shape);

				if (nilText !== "<nil>") {
					return go2jsPad(nilText, parsed, false);
				}
			}

			// A compound operand is formatted element by element, the way fmt
			// applies the verb to each field rather than to the whole value.
			const inner = "%" + flags + (precision === null ? "" : "." + precision) + "q";
			const quoteElement = (element) => (typeof element === "number" && Number.isInteger(element)
				? go2jsQuoteRune(element)
				: go2jsFormatValue("q", inner, element));

			if (value instanceof go2jsNativeMap) {
				const parts = [];

				for (const key of value.keys()) {
					parts.push(go2jsStrconvQuote(go2jsBytesToString(key)) + ":" + go2jsFormatValue("q", inner, value.get(key)));
				}

				return go2jsPad("map[" + parts.join(" ") + "]", parsed, false);
			}

			// Go treats a byte slice as a string for the quoted verbs, and a byte
			// that is not part of a valid UTF8 sequence is written as \xNN rather
			// than being replaced the way a decoded string would be.
			if (go2jsIsByteCompound(accepted)) {
				return go2jsPad(go2jsQuoteBytes(value), parsed, false);
			}

			if (Array.isArray(value)) {
				const parts = [];

				for (const element of value) {
					parts.push(quoteElement(element));
				}

				return go2jsPad("[" + parts.join(" ") + "]", parsed, false);
			}

			// A struct value is a class instance whose fields were assigned in
			// declaration order, which is the order fmt reports them in. Only a
			// real struct takes this path: a Stringer, an interface wrapper or a
			// plain map-like object is formatted through its own String method.
			if (value !== null && typeof value === "object" && !(value instanceof Error) &&
				!(value instanceof go2jsNativeSet) && !value.__go2js_interface &&
				!value.__go2js_pointer && go2jsIsGoStructType(accepted) &&
				go2jsStructHasNoStringMethod(value)) {
				const parts = [];

				for (const key of Object.keys(value)) {
					parts.push(quoteElement(value[key]));
				}

				return go2jsPad("{" + parts.join(" ") + "}", parsed, false);
			}

			if (typeof value === "number" && Number.isInteger(value)) {
				return go2jsPad(go2jsQuoteRune(value), parsed, false);
			}

			// An interface operand is quoted as the dynamic value it holds, the
			// way fmt reaches through the interface before applying the verb.
			if (operand !== null && typeof operand === "object" && operand.__go2js_interface === true) {
				return go2jsFormatValue("q", "%" + flags + (precision === null ? "" : "." + precision) + "q",
					operand.value);
			}

			// A named function type can still carry a String method, and Go
			// quotes what that method returns rather than the function itself.
			if (typeof value === "function") {
				const named = go2jsNamedFormatMethod(value, "String");

				if (named !== null) {
					return go2jsPad(go2jsStrconvQuote(go2jsBytesToString(named)), parsed, false);
				}
			}

			// An error or Stringer is quoted through its own method, which is
			// what fmt does for a pointer or interface operand.
			if (value !== null && typeof value === "object" &&
				!(value instanceof go2jsNativeMap) && !Array.isArray(value) &&
				!(value instanceof Uint8Array)) {
				const receiver = value.__go2js_pointer === true ? go2jsDeref(value) : value;
				const named = go2jsNamedFormatMethod(receiver, "Error") ?? go2jsNamedFormatMethod(receiver, "String");

				if (named !== null) {
					return go2jsPad(go2jsStrconvQuote(go2jsBytesToString(named)), parsed, false);
				}
			}

			return go2jsPad(go2jsStrconvQuote(go2jsBytesToString(value)), parsed, false);
		}
		case "t":
			return go2jsPad(value ? "true" : "false", parsed, false);
		case "c":
			return go2jsPad(go2jsRuneString(value), parsed, false);
		case "U": {
			let digits = Math.trunc(Number(value)).toString(16).toUpperCase();

			while (digits.length < 4) {
				digits = "0" + digits;
			}

			return go2jsPad("U+" + digits, parsed, false);
		}
		case "T":
			return go2jsPad(go2jsGoTypeName(original), parsed, false);
		default:
			return go2jsStringify(value);
	}
}

function go2jsParseFormatSpec(spec) {
	let i = 1;
	let flags = "";

	while (i < spec.length && "+-# 0".includes(spec[i])) {
		flags += spec[i];
		i++;
	}

	let width = "";

	while (i < spec.length && spec[i] >= "0" && spec[i] <= "9") {
		width += spec[i];
		i++;
	}

	let precision = null;

	if (i < spec.length && spec[i] === ".") {
		i++;

		let digits = "";

		while (i < spec.length && spec[i] >= "0" && spec[i] <= "9") {
			digits += spec[i];
			i++;
		}

		precision = digits === "" ? 0 : parseInt(digits, 10);
	}

	return {
		flags,
		width: width === "" ? 0 : parseInt(width, 10),
		precision,
		verb: spec[i],
	};
}

// go2jsPad widens a rendered value to the width the verb asked for. A zero flag
// fills with zeros in front of the digits but behind any sign, so a negative
// value reads -0042 rather than 00-42. An integer that was given a precision
// takes spaces instead, because a precision and a zero flag together leave the
// flag with nothing to say: %05.3d of 42 is "  042" while %05d is "00042".
// go2jsBigintFormat writes a number that is too big for a double the way fmt
// writes an integer, reading the digits off the whole number rather than off a
// rounded one. A negative number written in a base other than ten keeps the sign
// in front rather than borrowing from the digits, which is what fmt does for an
// integer of a signed type.
function go2jsBigintFormat(value, parsed, typeName, kind) {
	const spec = parsed === undefined ? {verb: "v", width: 0, precision: null, flags: ""} : parsed;
	const verb = spec.verb === undefined ? "v" : spec.verb;
	const flags = spec.flags === undefined ? "" : spec.flags;
	const precision = spec.precision === undefined ? null : spec.precision;
	const accepted = typeName || kind || "int";

	// a number of this size is a whole number, so the verbs written for a float
	// take it as one, and the ones that only a float answers say so
	if (go2jsVerbInSet(verb, "efgEFG")) {
		return go2jsFormatValue(verb, spec.width > 0 ? "%" + flags + spec.width + (precision === null ? "" : "." + precision) + verb : "%" + flags + verb,
			go2jsTyped(Number(value), "float64"));
	}

	if (verb === "T") {
		return go2jsFormatValue("T", "%T", go2jsTyped(0, accepted));
	}

	// these three read a number as the character it stands for. %s does not: it
	// is written for a string, and fmt says so when a number is given to it.
	if (verb === "c" || verb === "q" || verb === "U") {
		return go2jsFormatValue(verb, "%" + verb, go2jsTyped(Number(value), "rune"));
	}

	if (verb === "p") {
		return "0x0";
	}

	let out;

	if (verb === "v" || verb === "d" || verb === undefined) {
		out = value.toString(10);
	} else if (verb === "b" || verb === "B") {
		out = value.toString(2);
	} else if (verb === "o") {
		out = value.toString(8);
	} else if (verb === "x" || verb === "X") {
		out = value.toString(16);
	} else {
		return "%!" + verb + "(" + (accepted === "rune" ? "int32" : accepted) + "=" + value.toString(10) + ")";
	}

	// only %X asks for capital digits, and %x with the sharp flag keeps its
	// prefix and its digits in the same case
	if (verb === "X") {
		out = out.toUpperCase();
	}

	// a base prefix is written in front of the digits when the sharp flag asks
	// for it, and %X has none of its own
	let prefix = "";

	if (flags.includes("#")) {
		if (verb === "b" || verb === "B") {
			prefix = "0b";
		} else if (verb === "o") {
			prefix = "0";
		} else if (verb === "x" && value < 0n) {
			prefix = "-0x";
		} else if (verb === "x") {
			prefix = "0x";
		}
	}

	// the sharp flag leaves a value of zero in front of its prefix rather than
	// turning the prefix into the whole of the number
	const body = prefix === "" ? out : (out === "0" && value === 0n ? "0" : prefix + out);

	return go2jsPad(body, {width: spec.width || 0, flags: flags.includes("+") || flags.includes(" ") || flags.includes("#") ? flags + "+" : flags, precision: precision}, true);
}

function go2jsPad(text, parsed, integer) {
	if (parsed.width <= text.length) {
		return text;
	}

	const fill = parsed.width - text.length;

	if (parsed.flags.includes("-")) {
		return text + " ".repeat(fill);
	}

	if (parsed.flags.includes("0") && !(integer === true && parsed.precision !== null)) {
		const sign = text.startsWith("-") || text.startsWith("+") ? text[0] : "";

		return sign + text.slice(sign.length).padStart(text.length - sign.length + fill, "0");
	}

	return " ".repeat(fill) + text;
}

// go2jsFormatComplex writes a complex value the way fmt does: both parts take
// the verb, the imaginary part is written with an explicit sign in front of its
// magnitude, and the pair is wrapped in parentheses with an i.
function go2jsFormatComplex(verb, parsed, value, accepted, flags, precision) {
	const typeName = go2jsIsComplexTypeName(accepted) ? accepted : "complex128";
	const { re, im } = go2jsComplexValue(value);

	if (!go2jsVerbAccepts(verb, typeName)) {
		// A refused verb reports the operand the way %v would write it.
		return "%!" + verb + "(" + typeName + "=" + go2jsFormatComplexText("v", parsed, re, im, "", null) + ")";
	}

	return go2jsPad(go2jsFormatComplexText(verb, parsed, re, im, flags, precision), parsed, false);
}

// go2jsFormatComplexText writes the (re+imi) body. The imaginary part is written
// as a magnitude with its sign kept out of the part itself, because the sign
// joins the two parts rather than belonging to the imaginary one. A width, a
// sign and a sharp flag belong to the whole value, so neither part carries them.
function go2jsFormatComplexText(verb, parsed, re, im, flags, precision) {
	// %v and %g write each part in its shortest form and ignore a precision,
	// which is how a complex reads in Go.
	const part = verb === "v" || verb === "g" || verb === "G" ? "v" : verb;
	const spec = part === "v" || precision === null ? "%v" : "%." + precision + part;
	// Both parts are floats in Go, so they are tagged as one, which is what lets
	// the float verbs accept them instead of refusing the part.
	const asFloat = (part) => go2jsTyped(part, "float64");
	const real = go2jsFormatValue(part, spec, asFloat(re));
	const negative = im < 0 || Object.is(im, -0);
	const magnitude = negative ? -im : im;
	const imaginary = go2jsFormatValue(part, spec, asFloat(magnitude));

	return "(" + real + (negative ? "-" : "+") + imaginary + "i)";
}

// go2jsIsSliceTypeName reports whether a type name names a slice, and
// go2jsIsArrayTypeName whether it names an array.
function go2jsIsSliceTypeName(typeName) {
	return typeName !== null && typeName !== undefined && typeName.startsWith("[]");
}

function go2jsIsArrayTypeName(typeName) {
	return typeName !== null && typeName !== undefined && /^\[\d+\]/.test(typeName);
}

// go2jsIsComplexOperand reports whether an operand is a complex value, which is
// a pair of a real and an imaginary part.
function go2jsIsComplexOperand(value, accepted) {
	if (go2jsIsComplexTypeName(accepted)) {
		return true;
	}

	return value !== null && typeof value === "object" && "re" in value && "im" in value;
}

// go2jsIsByteCompound reports whether a type name is a slice or an array of
// bytes, which Go treats as a string for the quoted verbs and the hex verbs.
function go2jsIsByteCompound(accepted) {
	return accepted === "[]byte" || accepted === "[]uint8" ||
		accepted !== null && accepted !== undefined && /^\[\d+\](byte|uint8)$/.test(accepted);
}

// go2jsCompoundElementType returns the element type of a slice or an array type
// name, so an element can keep the type the compound declared it with.
function go2jsCompoundElementType(accepted) {
	if (go2jsIsSliceTypeName(accepted)) {
		return accepted.slice(2);
	}

	const array = accepted !== null && accepted !== undefined ? /^\[(\d+)\](.*)$/.exec(accepted) : null;

	return array === null ? null : array[2];
}

// go2jsCompoundElements returns the elements of a slice or an array operand, or
// null when the operand is not one, so a verb can be applied to each of them the
// way fmt expands a compound operand before formatting.
function go2jsCompoundElements(value, accepted) {
	if (Array.isArray(value) || value instanceof Uint8Array) {
		return Array.from(value);
	}

	if (go2jsIsSliceTypeName(accepted) || go2jsIsArrayTypeName(accepted)) {
		return value === null || value === undefined ? null : [];
	}

	return null;
}

// go2jsFloatBits splits a double into the parts the hex and binary verbs need.
// The words are read big endian so the sign, the exponent and the top of the
// mantissa all sit in the first word, which keeps the bit layout obvious.
function go2jsFloatBits(num) {
	const buffer = new DataView(new ArrayBuffer(8));

	buffer.setFloat64(0, num, false);

	const high = buffer.getUint32(0, false);
	const low = buffer.getUint32(4, false);

	return {
		negative: (high & 0x80000000) !== 0,
		rawExponent: (high >>> 20) & 0x7ff,
		mantissa: (BigInt(high & 0xfffff) << 32n) | BigInt(low),
	};
}

// go2jsSpecialFloatText writes the text Go uses for a value that is not finite.
// An infinity keeps its sign, a NaN does not.
function go2jsSpecialFloatText(num) {
	return Number.isNaN(num) ? "NaN" : num === Infinity ? "+Inf" : num === -Infinity ? "-Inf" : "";
}

// go2jsSpecialFloat writes a value that is not finite the way fmt writes it for
// the flags asked: a NaN takes a sign when one is asked and a space otherwise,
// an infinity becomes a space flanked sign only when a space is asked and no
// sign is, else keeping the sign it carries.
function go2jsSpecialFloat(num, parsed) {
	const special = go2jsSpecialFloatText(num);
	const plus = parsed.flags.includes("+");
	const space = parsed.flags.includes(" ");

	if (Number.isNaN(num)) {
		return plus ? "+NaN" : space ? " NaN" : "NaN";
	}

	if (space && !plus && special[0] === "+") {
		return " " + special.slice(1);
	}

	return special;
}

// go2jsFormatFixed writes a number in positional notation with a fixed number of
// fraction digits. Above 1e21 toFixed falls back to an exponent, so the value is
// written out from the shortest form that reads back as the same float, which is
// what Go prints before it pads the fraction.
function go2jsFormatFixed(abs, precision, sharp) {
	if (abs === 0) {
		let body = "0";

		if (precision > 0) {
			body += "." + "0".repeat(precision);
		}

		return sharp === true && precision === 0 ? body + "." : body;
	}

	// The fraction is cut at the precision with the exact digits the value
	// holds, rounding a tie to an even neighbor the way strconv's decimal does,
	// then written as fmtF writes it.
	const exact = go2jsExactDigits(abs);
	let digits = exact.digits;
	let dp = exact.dp;
	const roundAt = dp + precision;

	if (roundAt >= 0 && roundAt < digits.length) {
		const result = go2jsRoundDigits(digits, roundAt);
		digits = result.digits;

		if (result.carry) {
			dp += 1;
		}
	}

	let body = go2jsFormatGsfFixed(digits, digits.length, dp, precision);

	return sharp === true && precision === 0 && !body.includes(".") ? body + "." : body;
}

// magnitude returns the value without its sign, which is what a fraction and an
// exponent are written against.
function magnitude(num) {
	return num < 0 || Object.is(num, -0) ? -num : num;
}

// go2jsHexFraction writes the bits below a value's leading 1 as the hex digits
// they occupy, with the trailing zeros that carry no value left off.
function go2jsHexFraction(below, top) {
	if (below === 0n) {
		return "";
	}

	const digits = Math.ceil(top / 4);
	let text = (below << BigInt(4 * digits - top)).toString(16);

	while (text.length > 0 && text.endsWith("0")) {
		text = text.slice(0, -1);
	}

	return text;
}

// go2jsFormatHexFloat writes a float the way C99 and Go write it for %x and %X:
// a leading 1, a dot, the fraction digits and the exponent in binary with a p
// marker. Without a precision the fraction is the value's own bits, so it is
// exact; with one it is that many digits, rounded.
function go2jsFormatHexFloat(num, precision, upper) {
	const special = go2jsSpecialFloatText(num);

	if (special !== "") {
		return special;
	}

	const sign = num < 0 || Object.is(num, -0) ? "-" : "";
	const value = magnitude(num);
	const { rawExponent, mantissa } = go2jsFloatBits(value);

	// A double is its 52 bit significand times a power of two. A normal value
	// keeps the hidden leading bit, a subnormal does not and so starts one binade
	// lower, and a zero has no significand at all.
	const scale = (rawExponent === 0 ? 1 : rawExponent) - 1023 - 52;

	// A normal value keeps a hidden leading bit, so a zero mantissa is not a zero
	// value: 1.0 has a mantissa of nothing and reads 0x1p+00.
	if (magnitude(num) === 0) {
		const body = sign + "0x0" + (precision === null || precision === 0 ? "" : "." + "0".repeat(precision)) + "p+00";

		return upper ? body.toUpperCase() : body;
	}

	// A normal value carries a hidden leading bit, so the significand is the
	// mantissa with that bit added back, and the leading 1 of the literal is the
	// highest bit it holds. Everything below that bit is the fraction.
	const significand = rawExponent === 0 ? mantissa : mantissa | (1n << 52n);
	const top = significand.toString(2).length - 1;
	let exponent = scale + top;
	let text;

	if (precision === null) {
		text = go2jsHexFraction(significand - (1n << BigInt(top)), top);
	} else {
		const units = Math.pow(2, precision * 4);
		const rounded = Math.round((value / Math.pow(2, exponent) - 1) * units);

		// A round that carries past the leading 1 moves the point instead.
		if (rounded >= units) {
			exponent += 1;
			text = "0".repeat(precision);
		} else {
			text = rounded === 0 ? "0".repeat(precision) : BigInt(rounded).toString(16).padStart(precision, "0");
		}
	}

	const body = sign + "0x1" + (text === "" ? "" : "." + text) + "p" + (exponent < 0 ? "-" : "+") +
		(Math.abs(exponent) < 10 ? "0" : "") + Math.abs(exponent);

	return upper ? body.toUpperCase() : body;
}

// go2jsFloat32Bits splits a float32 into its exponent and mantissa, which are
// narrower than a double's and so report different significands.
function go2jsFloat32Bits(value) {
	const buffer = new DataView(new ArrayBuffer(4));

	buffer.setFloat32(0, Math.fround(value), false);

	const word = buffer.getUint32(0, false);

	return {
		rawExponent: (word >>> 23) & 0xff,
		mantissa: BigInt(word & 0x7fffff),
	};
}

// go2jsFormatFloatBits writes the IEEE754 expansion of a float: its significand
// as a decimal count of the units the exponent names.
function go2jsFormatFloatBits(num, typeName) {
	const special = go2jsSpecialFloatText(num);

	if (special !== "") {
		return special;
	}

	// A float32 is written from its own 24 bit significand and exponent rather
	// than from the 52 bit ones a double would report.
	const single = typeName === "float32";
	const parts = single ? go2jsFloat32Bits(magnitude(num)) : go2jsFloatBits(magnitude(num));
	const { rawExponent, mantissa } = parts;
	const width = single ? 23 : 52;
	const bias = single ? 127 : 1023;

	// The significand is the mantissa with the hidden leading bit a normal value
	// keeps, and the exponent is the power of two the value is scaled by, which is
	// the stored exponent minus the bias and the width of the mantissa. A
	// subnormal and a zero start one binade lower because they have no hidden bit.
	const hidden = 1n << BigInt(width);
	const significand = rawExponent === 0 ? mantissa : mantissa | hidden;
	const scale = (rawExponent === 0 ? 1 : rawExponent) - bias - width;
	const sign = num < 0 || Object.is(num, -0) ? "-" : "";

	return sign + significand.toString(10) + "p" + (scale < 0 ? "-" : "+") + Math.abs(scale).toString(10);
}

// go2jsDigitsOf counts the digits of a rendered integer, ignoring the sign and
// any base prefix.
function go2jsDigitsOf(text) {
	const digits = text.replace(/^[-+ ]/, "").replace(/^0[xXobB]/, "");

	return digits.length;
}

// go2jsZeroPadDigits pads an already rendered integer with leading zeros until
// it has at least the wanted number of digits.
function go2jsZeroPadDigits(text, wanted) {
	const sign = text.startsWith("-") ? "-" : "";
	const body = sign === "" ? text : text.slice(1);
	const prefixMatch = body.match(/^0[xXobB]/);
	const prefix = prefixMatch === null ? "" : prefixMatch[0];
	const digits = prefix === "" ? body : body.slice(prefix.length);

	return sign + prefix + digits.padStart(wanted, "0");
}

// go2jsPadNumber pads a rendered number out to its width. A zero flag fills with
// zeros in front of the digits but behind any sign or base prefix, which is how
// Go writes -003.142 and 0x00ff. An integer that was given a precision takes
// spaces instead, because a precision and a zero flag together leave the flag
// with nothing to say: %05.3d of 42 is "  042" while %05d is "00042".
function go2jsPadNumber(text, prefix, body, parsed, integer) {
	// A base prefix sits outside the width, so %08x with a sharp flag writes
	// 0x000000ff, which is the eight digits with the prefix in front of them.
	const digits = /^[0][xXoObB]/.test(prefix) ? body.length : text.length;

	if (parsed.width <= digits) {
		return text;
	}

	const fill = parsed.width - digits;

	if (parsed.flags.includes("-")) {
		return text + " ".repeat(fill);
	}

	if (parsed.flags.includes("0") && !(integer === true && parsed.precision !== null)) {
		return prefix + body.padStart(body.length + fill, "0");
	}

	return " ".repeat(fill) + text;
}

// go2jsFormatFloatType writes a float the way the width it is kept in writes it,
// since a float32 has fewer digits to give and the shortest form that reads
// back as the same float32 is shorter than the one that reads back as the same
// float64.
function go2jsFormatFloatType(value, typeName, kind) {
	const single = go2jsIsFloat32TypeName(typeName) || go2jsIsFloat32TypeName(kind);

	return single ? go2jsFormatFloat32Default(value) : go2jsFormatFloatDefault(value);
}

function go2jsIsFloat32TypeName(name) {
	return typeof name === "string" && name === "float32";
}

// go2jsFormatFloat32Default mirrors fmt's %v rule for float32: the fewest
// digits that read back as the same float32. A float32 holds one bit fewer
// digit than a double does, so a value carried in one is written in fewer
// digits than the same value carried in a double would be.
function go2jsFormatFloat32Default(value) {
	const num = Number(value);

	if (Number.isNaN(num) || num === Infinity || num === -Infinity) {
		return go2jsFormatFloatDefault(num);
	}

	if (num === 0) {
		return Object.is(num, -0) ? "-0" : "0";
	}

	if (Math.fround(num) !== num) {
		return go2jsFormatFloatDefault(num);
	}

	for (let digits = 1; digits <= 9; digits++) {
		const text = num.toExponential(digits - 1);
		const rounded = Number(text);

		if (Math.fround(rounded) === num) {
			const parts = text.split("e");
			const digits = parts[0].replace(".", "");

			return go2jsWriteFloatDigits(num < 0 ? "-" : "", digits, parseInt(parts[1], 10));
		}
	}

	return go2jsFormatFloatDefault(num);
}

// go2jsFormatFloatDefault mirrors fmt's %v rule for float64: the shortest
// round-trip digits, switching to scientific notation when the decimal exponent
// is below -4 or at least 6.
function go2jsFormatFloatDefault(value) {
	const num = Number(value);

	if (Number.isNaN(num)) {
		return "NaN";
	}

	if (num === Infinity) {
		return "+Inf";
	}

	if (num === -Infinity) {
		return "-Inf";
	}

	if (num === 0) {
		return Object.is(num, -0) ? "-0" : "0";
	}

	const parts = Math.abs(num).toExponential().split("e");
	const digits = parts[0].replace(".", "");
	const exp10 = parseInt(parts[1], 10);

	return go2jsWriteFloatDigits(num < 0 ? "-" : "", digits, exp10);
}

// go2jsWriteFloatDigits writes a number from the digits that stand for it and
// the place the point goes, switching to scientific notation when the decimal
// exponent is below -4 or at least 6, which is the rule fmt's %v writes by.
function go2jsWriteFloatDigits(sign, digits, exp10) {
	if (exp10 < -4 || exp10 >= 6) {
		let body = digits[0];

		if (digits.length > 1) {
			body += "." + digits.slice(1);
		}

		const exponentSign = exp10 < 0 ? "-" : "+";
		const exponent = String(Math.abs(exp10)).padStart(2, "0");

		return sign + body + "e" + exponentSign + exponent;
	}

	if (exp10 >= 0) {
		if (digits.length > exp10 + 1) {
			return sign + digits.slice(0, exp10 + 1) + "." + digits.slice(exp10 + 1);
		}

		return sign + digits + "0".repeat(exp10 + 1 - digits.length);
	}

	return sign + "0." + "0".repeat(-exp10 - 1) + digits;
}

function go2jsFormatE(value, precision, parsed) {
	if (!Number.isFinite(value)) {
		return go2jsPad(String(value), parsed, false);
	}

	const abs = magnitude(value);
	let body;

	if (abs === 0) {
		body = "0";

		if (precision > 0) {
			body += "." + "0".repeat(precision);
		}

		body += "e+00";
	} else {
		// The mantissa is rounded to the precision plus the one digit it leads
		// with, again to an even neighbor on a tie, and then written as fmtE
		// writes it with the exponent the point sits at.
		const exact = go2jsExactDigits(abs);
		let digits = exact.digits;
		let dp = exact.dp;
		const roundAt = precision + 1;

		if (roundAt < digits.length) {
			const result = go2jsRoundDigits(digits, roundAt);
			digits = result.digits;

			if (result.carry) {
				dp += 1;
			}
		}

		body = go2jsFormatTimeExponent(digits, digits.length, dp, precision, false);
	}

	// The sharp flag keeps the point even where a precision of zero leaves
	// nothing after it, so 1.e+00 and 1. are written with the point that holds.
	if (parsed.flags.includes("#") && precision === 0) {
		body = body.replace("e", ".e");
	}

	const prefix = value < 0 || Object.is(value, -0) ? "-" : parsed.flags.includes("+") ? "+" : parsed.flags.includes(" ") ? " " : "";

	return go2jsPadNumber(prefix + body, prefix, body, parsed);
}

// go2jsDecimalParts splits a JS shortest representation into significant digits
// and a decimal point position, so %g can pick between %e and %f like Go does.
// The value equals 0.<digits> * 10^dp.
function go2jsDecimalParts(value) {
	const text = String(Math.abs(value));

	if (text === "Infinity" || text === "NaN") {
		return null;
	}

	let mantissa = text;
	let exponent = 0;

	if (text.includes("e")) {
		const [mantissaPart, exponentPart] = text.split("e");
		mantissa = mantissaPart;
		exponent = parseInt(exponentPart, 10);
	}

	const [intPart, fracPart = ""] = mantissa.split(".");
	const raw = intPart + fracPart;
	const fractionDigits = fracPart.length;
	const firstSignificant = raw.search(/[1-9]/);

	if (firstSignificant < 0) {
		return { digits: "0", dp: 1 };
	}

	const significant = raw.slice(firstSignificant);
	const digits = significant.replace(/0+$/, "") || "0";
	// Trailing zeros stripped above scale the value rather than change dp.
	const dp = significant.length - fractionDigits + exponent;

	return { digits, dp };
}

// go2jsExactDigits writes the digits a finite float holds exactly, with the
// place of its point. A float is a whole number times a power of two, and a
// power of two is a whole number of fives over a power of ten, so the exact
// digits are the product gathered that way, with the point cut at the same
// distance the power of ten lays it.
function go2jsExactDigits(num) {
    const {rawExponent, mantissa} = go2jsFloatBits(num);

    const k = rawExponent === 0 ? -1074 : rawExponent - 1075;
    const significant = rawExponent === 0 ? mantissa : (1n << 52n) | mantissa;

    let digits;
    let dp;

    if (k >= 0) {
        digits = (significant << BigInt(k)).toString();
        dp = digits.length;
    } else {
        const m = -k;
        const product = significant * 5n ** BigInt(m);
        const text = product.toString();

        digits = text;
        dp = text.length - m;
    }

    // The point never moves when trailing zeros are set aside: they lie at the
    // end of the number, not in front of any digit that the point counts.
    while (digits.length > 1 && digits.endsWith("0")) {
        digits = digits.slice(0, -1);
    }

    return {digits, dp};
}

// go2jsFormatG renders %g, using the shortest representation when no precision
// is given and switching to the exponent form outside Go's -4..eprec window.
// Rendered as %G it spells the exponent with an E rather than an e, which is the
// only letter the two forms disagree about.
function go2jsFormatG(value, precision, parsed, upper) {
    // The sharp flag keeps the zeros of the fraction and forces a point even
    // where a number has none of its own, and the count of digits it keeps is
    // the precision the sharp flag asks for.
    const sharp = parsed.flags.includes("#");
    const sharpDigits = precision === null ? 6 : precision;

    // A whole zero is written as a single zero in the fixed form, which the
    // path for a number with digits would not reach on its own.
    if (value === 0) {
        const body = sharp ? go2jsRetainSharp("0", sharpDigits) : "0";

        return go2jsFormatGsfSign(body, value, parsed);
    }

    // A precision asked as zero is the precision of one, which Go reads the
    // same way, while the sharp flag keeps the zeros the letter itself asked.
    const significant = precision === null ? null : Math.max(precision, 1);

    // A negative precision means the shortest form that reads back as the same
    // number, for which the count of digits is the count the number holds.
    const shortest = significant === null;

    let digits;
    let dp;

    if (shortest) {
        const parts = go2jsDecimalParts(value);

        if (parts === null) {
            return go2jsPad(go2jsSpecialFloat(value, parsed), parsed, false);
        }

        digits = parts.digits;
        dp = parts.dp;
    } else {
        // A precision rounds the number itself, which is not the shortest text
        // that names it but the exact sum it holds in binary, so the exact
        // digits are read off its bits before they are rounded.
        if (!Number.isFinite(value)) {
            return go2jsPad(go2jsSpecialFloat(value, parsed), parsed, false);
        }

        const exact = go2jsExactDigits(Math.abs(value));

        digits = exact.digits;
        dp = exact.dp;
    }

    let nd = digits.length;

    if (!shortest && nd > significant) {
        // The digits are rounded to the requested count of significant figures,
        // the way a run of nines rounds up and over into a single leading one
        // that moves the point with it, as Go moves it.
        const carried = go2jsRoundDigits(digits, significant);
        digits = carried.digits;
        dp = dp + (carried.carry ? 1 : 0);
        nd = digits.length;
    }

    let prec = shortest ? nd : significant;
    let eprec = prec;

    // A precision finer than a number can hold asks for only what the number
    // holds, as does a precision coarser than one: both give the door of the
    // shortest form, which is the number's own length.
    if (eprec > nd && nd >= dp) {
        eprec = nd;
    }

    if (shortest) {
        eprec = 6;
    }

    const exp = dp - 1;

    // Go writes a number as an exponent when its position in the number line is
    // far from the point, and the number itself for everything between.
    let body;

    if (exp < -4 || exp >= eprec) {
        if (prec > nd) {
            prec = nd;
        }

        body = go2jsFormatTimeExponent(digits, nd, dp, prec - 1, upper);
    } else {
        if (prec > dp) {
            prec = nd;
        }

        body = go2jsFormatGsfFixed(digits, nd, dp, Math.max(prec - dp, 0));
    }

    if (sharp) {
        body = go2jsRetainSharp(body, sharpDigits);
    }

    return go2jsFormatGsfSign(body, value, parsed);
}

// go2jsRetainSharp is the mark of the sharp flag on a number: a point where the
// number has none and zeros kept up to the precision's count of significant
// digits, which is the way fmt keeps them when it is asked to.
function go2jsRetainSharp(text, digits) {
    let remaining = digits;
    let sawNonzero = false;
    let hasDecimalPoint = false;
    let out = "";
    let tail = "";

    for (let i = 0; i < text.length; i++) {
        const c = text[i];

        if (c === ".") {
            hasDecimalPoint = true;
            out += c;
        } else if (c === "e" || c === "E") {
            tail = text.slice(i);
            break;
        } else {
            out += c;

            if (c !== "0") {
                sawNonzero = true;
            }

            if (sawNonzero) {
                remaining--;
            }
        }
    }

    if (!hasDecimalPoint) {
        if (out === "0") {
            remaining--;
        }

        out += ".";
    }

    while (remaining > 0) {
        out += "0";
        remaining--;
    }

    return out + tail;
}

// go2jsFormatGsfSign puts the sign a number asks for in front of the text of it,
// and pads the two of them to the width asked for.
function go2jsFormatGsfSign(body, value, parsed) {
    const prefix = value < 0 || Object.is(value, -0) ? "-" : parsed.flags.includes("+") ? "+" : parsed.flags.includes(" ") ? " " : "";

    return go2jsPadNumber(prefix + body, prefix, body, parsed);
}

// go2jsRoundDigits rounds a decimal digit string down to n significant digits.
// The digit cut off decides the direction, with an exact half rounding to an
// even neighbor, and a run of nines rounds up to a single leading one that the
// caller's point moves with.
function go2jsRoundDigits(string, n) {
    const cut = string[n];

    if (cut === "5" && n + 1 === string.length) {
        const before = string.charCodeAt(n - 1) - 48;

        if (before % 2 === 1) {
            return go2jsRoundDigitsUp(string, n);
        }

        let digits = string.slice(0, n);

        while (digits.length > 0 && digits.endsWith("0")) {
            digits = digits.slice(0, -1);
        }

        return {digits: digits.length === 0 ? "0" : digits, carry: false};
    }

    if (cut >= "5") {
        return go2jsRoundDigitsUp(string, n);
    }

    let digits = string.slice(0, n);

    while (digits.length > 0 && digits.endsWith("0")) {
        digits = digits.slice(0, -1);
    }

    return {digits: digits.length === 0 ? "0" : digits, carry: false};
}

function go2jsRoundDigitsUp(string, n) {
    let digits = string.slice(0, n);

    for (let i = n - 1; i >= 0; i--) {
        if (digits.charCodeAt(i) < 57) {
            digits = digits.slice(0, i) + String.fromCharCode(digits.charCodeAt(i) + 1) + digits.slice(i + 1);
            return {digits: digits.slice(0, i + 1), carry: false};
        }
    }

    return {digits: "1", carry: true};
}

// go2jsFormatTimeExponent writes a rounded decimal as Go's fmtE writes it: the
// first digit, a point and the digits that follow it up to the precision, and
// the distance of the point from the one place, its exponent, as a count of
// powers of ten with a sign.
function go2jsFormatTimeExponent(digits, nd, dp, prec, upper) {
    let body = digits[0];

    if (prec > 0) {
        body += ".";

        for (let i = 1; i <= prec; i++) {
            body += i < nd ? digits[i] : "0";
        }
    }

    const magnitude = dp - 1;
    const positive = magnitude >= 0;
    const text = String(Math.abs(magnitude)).padStart(2, "0");
    const letter = upper ? "E" : "e";

    return body + letter + (positive ? "+" : "-") + text;
}

// go2jsFormatGsfFixed writes a rounded decimal as Go's fmtF writes it: the whole
// part read up to the point, and the fraction down to the precision, with zeros
// where no digit is held.
function go2jsFormatGsfFixed(digits, nd, dp, prec) {
    let body = "";

    if (dp > 0) {
        const whole = Math.min(nd, dp);

        body += digits.slice(0, whole);

        for (let i = whole; i < dp; i++) {
            body += "0";
        }
    } else {
        body += "0";
    }

    if (prec > 0) {
        body += ".";

        for (let i = 0; i < prec; i++) {
            const at = dp + i;

            body += at >= 0 && at < nd ? digits[at] : "0";
        }
    }

    return body;
}

function go2jsStringsContains(s, substr) {
	return s.includes(substr);
}

function go2jsStringsToUpper(s) {
	return s.toUpperCase();
}

function go2jsStringsSplit(s, sep) {
	if (sep === "") {
		return s.split("");
	}
	return s.split(sep);
}

function go2jsStringsJoin(parts, sep) {
	return parts.join(sep);
}

function go2jsStringsReplaceAll(s, old, replacement) {
	return s.split(old).join(replacement);
}

// go2jsStrconvParseInteger reads a whole number in the way Go's strconv does:
// nothing but the digits of the number and an optional sign may be there, and
// JavaScript's own parsing is far more forgiving, so the text is checked first.
function go2jsStrconvParseInteger(text, base) {
	if (typeof text === "string") {
		text = text.trim();
	} else {
		text = go2jsStringify(text);
	}

	const sign = /^[+-]/.test(text) ? text[0] : "";

	if (sign) {
		text = text.slice(1);
	}

	if (text.length === 0) {
		return null;
	}

	base = base === undefined || base === 0 ? 10 : base;

	// A base that is not one of the ones Go accepts names no number at all.
	if (![2, 8, 10, 16].includes(base)) {
		return null;
	}

	const digits = base === 16 ? /^[0-9a-fA-F]+$/ : base === 10 ? /^[0-9]+$/ : base === 8 ? /^[0-7]+$/ : /^[01]+$/;

	if (!digits.test(text)) {
		return null;
	}

	if (sign === "+" && /^0/.test(text)) {
		return null;
	}

	let value;

	if (base === 10) {
		value = Number(text);
	} else {
		value = parseInt(text, base);
	}

	if (typeof value !== "number" || !Number.isFinite(value)) {
		return null;
	}

	return sign === "-" ? -value : value;
}

// go2jsStrconvFitsWidth reports whether a number is inside the width it was
// asked to be read at, because Go answers a count that does not fit with an
// error rather than with the number it could not hold.
function go2jsStrconvFitsWidth(value, bitSize) {
	switch (bitSize) {
	case 0:
	case 8:
		return value >= -128 && value <= 255;
	case 16:
		return value >= -32768 && value <= 65535;
	case 32:
		return value >= -2147483648 && value <= 4294967295;
	default:
		return true;
	}
}

function go2jsStrconvAtoi(s) {
	const value = go2jsStrconvParseInteger(s, 10);

	if (value === null) {
		return [0, new Error("strconv.Atoi: parsing " + JSON.stringify(s) + ": invalid syntax")];
	}

	if (!go2jsStrconvFitsWidth(value, 64)) {
		return [0, new Error("strconv.Atoi: parsing " + JSON.stringify(s) + ": value out of range")];
	}

	return [value, null];
}

function go2jsStrconvParseFloat(s, bitSize) {
	const text = go2jsStringify(s).trim();

	// JavaScript reads a whole number as an integer no matter which fraction was
	// asked for, so the fraction is asked about on its own.
	if (!/^[+-]?(\d+\.?\d*|\.\d+)([eE][+-]?\d+)?$/.test(text)) {
		return [0, new Error('strconv.ParseFloat: parsing "' + text + '": invalid syntax')];
	}

	const value = parseFloat(text);

	if (Number.isNaN(value)) {
		return [0, new Error('strconv.ParseFloat: parsing "' + text + '": invalid syntax')];
	}

	if (bitSize === 32) {
		return [Math.fround(value), null];
	}

	return [value, null];
}

function go2jsStrconvFormatInt(value, base) {
	return Math.trunc(value).toString(base);
}

function go2jsStrconvHexDigit(value) {
	return "0123456789abcdef"[value];
}

function go2jsStrconvRuneIsPrint(code) {
	if (code === 0x20) {
		return true;
	}

	if (code < 0x20 || code === 0x7f) {
		return false;
	}

	const char = String.fromCodePoint(code);

	return /^[\p{L}\p{M}\p{N}\p{P}\p{S}]$/u.test(char);
}

// go2jsStrconvRuneIsGraphic reports whether a code point is among the spaces a
// Go program would print as it is even though it is not a printing character,
// which is what separates the graphic runes from the printable ones.
function go2jsStrconvRuneIsGraphic(code) {
	if (code > 0xffff) {
		return false;
	}

	switch (code) {
	case 0x00a0:
	case 0x1680:
	case 0x2000:
	case 0x2001:
	case 0x2002:
	case 0x2003:
	case 0x2004:
	case 0x2005:
	case 0x2006:
	case 0x2007:
	case 0x2008:
	case 0x2009:
	case 0x200a:
	case 0x202f:
	case 0x205f:
	case 0x3000:
		return true;
	default:
		return false;
	}
}

// go2jsValidRune reports whether a number names a code point that a rune can
// stand for: a value within the rune space that is not a surrogate, which is
// what Go would replace with the replacement character.
function go2jsValidRune(code) {
	return Number.isInteger(code) && code >= 0 && code <= 0x10ffff && (code < 0xd800 || code > 0xdfff);
}

function go2jsStrconvAppendEscapedRune(out, code, quote, asciiOnly, graphicOnly) {
	if (!go2jsValidRune(code)) {
		code = 0xfffd;
	}

	const char = String.fromCodePoint(code);

	if (char === quote || char === "\\") {
		out.push("\\", char);
		return;
	}

	if (asciiOnly) {
		if (code < 0x80 && go2jsStrconvRuneIsPrint(code)) {
			out.push(char);
			return;
		}
	} else if (go2jsStrconvRuneIsPrint(code) || (graphicOnly && go2jsStrconvRuneIsGraphic(code))) {
		out.push(char);
		return;
	}

	switch (code) {
	case 0x07:
		out.push("\\a");
		return;
	case 0x08:
		out.push("\\b");
		return;
	case 0x0c:
		out.push("\\f");
		return;
	case 0x0a:
		out.push("\\n");
		return;
	case 0x0d:
		out.push("\\r");
		return;
	case 0x09:
		out.push("\\t");
		return;
	case 0x0b:
		out.push("\\v");
		return;
	}

	if (code < 0x20 || code === 0x7f) {
		out.push("\\x");
		out.push(go2jsStrconvHexDigit(code >> 4));
		out.push(go2jsStrconvHexDigit(code & 0xf));
		return;
	}

	if (code < 0x10000) {
		out.push("\\u");
	} else {
		out.push("\\U");
	}

	for (let shift = code < 0x10000 ? 12 : 28; shift >= 0; shift -= 4) {
		out.push(go2jsStrconvHexDigit((code >> shift) & 0xf));
	}
}

// go2jsUtf8Rune decodes the UTF8 sequence of a given width that starts at a
// byte, or returns 0 when the sequence is one a rune cannot stand for: an
// overlong encoding, a surrogate or a value past the last rune.
function go2jsUtf8Rune(bytes, at, width) {
	const head = bytes[at];
	const tail = (index) => bytes[at + index] & 0x3f;
	let rune;

	if (width === 1) {
		return head;
	}

	if (width === 2) {
		rune = ((head & 0x1f) << 6) | tail(1);
	} else if (width === 3) {
		rune = ((head & 0x0f) << 12) | (tail(1) << 6) | tail(2);
	} else {
		rune = ((head & 0x07) << 18) | (tail(1) << 12) | (tail(2) << 6) | tail(3);
	}

	const shortest = width === 2 ? 0x80 : width === 3 ? 0x800 : 0x10000;

	return rune < shortest || rune > 0x10ffff || rune >= 0xd800 && rune <= 0xdfff ? 0 : rune;
}

// go2jsQuoteBytes writes a byte slice the way %q writes a string: each rune is
// escaped on its own, and a byte that does not begin a valid UTF8 sequence is
// written as \xNN because it stands for no rune at all.
function go2jsQuoteBytes(value) {
	const bytes = Array.from(value, (item) => Number(item) & 255);
	const out = ['"'];

	for (let i = 0; i < bytes.length;) {
		const width = go2jsUtf8Width(bytes, i);
		let rune;

		if (width === 0) {
			out.push("\\x" + bytes[i].toString(16).padStart(2, "0"));
			i += 1;
			continue;
		}

		rune = go2jsUtf8Rune(bytes, i, width);

		// A sequence that decodes to a value a rune cannot hold stands for no
		// rune, so each of its bytes is written on its own.
		if (rune === 0) {
			for (let j = 0; j < width; j++) {
				out.push("\\x" + bytes[i + j].toString(16).padStart(2, "0"));
			}

			i += width;
			continue;
		}

		go2jsStrconvAppendEscapedRune(out, rune, '"', false, false);
		i += width;
	}

	out.push('"');

	return out.join("");
}

// go2jsUtf8Width returns the length of the UTF8 sequence that starts at a byte,
// or 0 when the byte does not begin one.
function go2jsUtf8Width(bytes, at) {
	const first = bytes[at];

	if (first < 0x80) {
		return 1;
	}

	const width = first >= 0xf0 ? 4 : first >= 0xe0 ? 3 : first >= 0xc0 ? 2 : 0;

	if (width === 0 || at + width > bytes.length) {
		return 0;
	}

	for (let i = 1; i < width; i++) {
		if ((bytes[at + i] & 0xc0) !== 0x80) {
			return 0;
		}
	}

	// A sequence that is well formed but names a rune Go cannot encode is still
	// a sequence, and the escaper decides what to write for it.
	return width;
}

function go2jsStrconvQuoteWith(value, asciiOnly, graphicOnly) {
	const out = ['"'];

	for (const char of String(value)) {
		go2jsStrconvAppendEscapedRune(out, char.codePointAt(0), '"', asciiOnly, graphicOnly);
	}

	out.push('"');

	return out.join("");
}

function go2jsStrconvQuote(s) {
	return go2jsStrconvQuoteWith(s, false, false);
}

function go2jsSortInts(values) {
	values.sort((a, b) => a - b);
}

function go2jsSortReverse(data, typeName) {
	this.data = data;
	this.typeName = typeName;
}

function go2jsSortTableMethod(target, typeName, method) {
	if (typeof typeName !== "string" || typeName === "") {
		return undefined;
	}

	for (const key of [typeName + "." + method, "*" + typeName + "." + method]) {
		const fn = go2jsMethodTable[key];

		if (typeof fn === "function") {
			return function(...args) { return go2jsCallNow(fn, null, [target, ...args]); };
		}
	}

	return undefined;
}

function go2jsSortMethod(target, typeName, method) {
	if (target === null || target === undefined) {
		return undefined;
	}

	if (target.__go2js_pointer === true) {
		const inner = target[go2jsPointerGet]();

		if (inner !== null && inner !== undefined && typeof inner[method] === "function") {
			return inner[method].bind(inner);
		}

		return go2jsSortTableMethod(target, typeName, method);
	}

	if (typeof target[method] === "function") {
		return target[method].bind(target);
	}

	return go2jsSortTableMethod(target, typeName, method);
}

go2jsSortReverse.prototype.Len = function() {
	const fn = go2jsSortMethod(this.data, this.typeName, "Len");

	if (typeof fn === "function") {
		return go2jsCallNow(fn, null, []);
	}

	return go2jsLen(this.data);
};

go2jsSortReverse.prototype.Less = function(i, j) {
	const fn = go2jsSortMethod(this.data, this.typeName, "Less");

	if (typeof fn === "function") {
		return go2jsCallNow(fn, null, [j, i]);
	}

	return go2jsCompareValues(this.data[j], this.data[i]) < 0;
};

go2jsSortReverse.prototype.Swap = function(i, j) {
	const fn = go2jsSortMethod(this.data, this.typeName, "Swap");

	if (typeof fn === "function") {
		go2jsCallNow(fn, null, [i, j]);
		return;
	}

	const items = this.data;
	const tmp = items[i];
	items[i] = items[j];
	items[j] = tmp;
};

function go2jsSortStrings(values) {
	values.sort();
}

go2jsRegisterMethod("encoding/json.Delim.String", go2jsJSONDelimText);

const go2jsSliceMeta = new WeakMap();

const go2jsArrayMark = new WeakSet();

function go2jsSliceState(value) {
	if (value === null || value === undefined) {
		return null;
	}

	const meta = go2jsSliceMeta.get(value);
	if (meta) {
		return meta;
	}

	if (Array.isArray(value)) {
		return {
			data: value,
			offset: 0,
			length: value.length,
			capacity: value.length
		};
	}

	return null;
}

function go2jsStructCopy(value) {
	if (value === null || value === undefined || typeof value !== "object") {
		return value;
	}

	if (value.__go2js_pointer === true) {
		return value;
	}

	// A value of a type of Go that stands for the value behind a pointer is one
	// value and not a record of fields to write out: Go shares what such a type
	// holds when the type is copied and says the type is not to be copied at all,
	// so a copy is the same value seen twice rather than a second one beside it.
	if (value.__go2js_shared === true) {
		return value;
	}

	// A date keeps its value inside itself, where a copy cannot reach it, so a
	// copied date stops being one the moment the copy is made. It is already
	// the value it stands for, so there is nothing here to copy.
	if (value instanceof go2jsNativeDate) {
		return value;
	}

	// A fault is not a record of fields to write out either, since what it is
	// worth saying is the text it says and a copy of that keeps nothing.
	if (value instanceof Error) {
		return value;
	}

	const embedded = typeof value.__go2js_embedded === "undefined"
		? null
		: value.__go2js_embedded;

	const copy = Object.create(Object.getPrototypeOf(value));

	for (const key of Object.keys(value)) {
		copy[key] = go2jsStructFieldCopy(value[key]);
	}

	if (embedded !== null) {
		return go2jsEmbedProxy(copy, embedded);
	}

	return copy;
}

// go2jsStructFieldCopy copies one field of a struct. A field holding a struct is
// a value of its own rather than a name for one somewhere else, so it is copied
// as well, and a copy of a struct is a copy of everything it is made of. A field
// holding a slice, a map or a pointer is not a value of its own but a reference
// to storage somebody else owns, and Go copies the reference and not the
// storage, so those are carried over as they stand.
function go2jsStructFieldCopy(item) {
	if (item === null || item === undefined || typeof item !== "object") {
		return item;
	}

	// A description of a type is the one description of it rather than a copy of
	// one, and copying it field by field drops the parts of it that are not
	// fields, which is where the way it writes itself is kept. It is carried
	// over as it stands for the same reason a pointer is.
	if (item.__go2js_pointer === true || item.__go2js_reflectValue === true ||
		item.__go2js_reflectType === true || item.__go2js_typed === true ||
		item.__go2js_interface === true || item instanceof go2jsNativeDate) {
		return item;
	}

	// A fault is reached through the text it says rather than through fields it
	// holds, and copying one field by field leaves it saying nothing at all, so
	// the fault a field carries is the fault that was there.
	if (item instanceof Error) {
		return item;
	}

	if (Array.isArray(item) || item.__go2js_nil === true) {
		return go2jsCopy(item);
	}

	return go2jsStructCopy(item);
}

function go2jsMaterializeValue(value) {
	if (value === null || value === undefined) {
		return value;
	}

	if (typeof value === "object" && go2jsSliceMeta.has(value)) {
		const state = go2jsSliceState(value);

		return state.data.slice(state.offset, state.offset + state.length);
	}

	return value;
}

// go2jsTimers holds the channels a timer is waiting on, so that a select with
// nothing else to do can wait for the moment one of them comes due rather than
// declaring every goroutine asleep.
const go2jsTimers = new Map();

// go2jsTimerDue says whether a timer has come due. A channel that stands for a
// timer is not ready until the moment it names, which is what lets a select
// choose the work that finished over the one that merely has a deadline.
function go2jsTimerDue(channel) {
	if (channel.timerDeadline === undefined) {
		return false;
	}

	if (Date.now() < channel.timerDeadline) {
		return false;
	}

	go2jsTimers.delete(channel);
	channel.timerDeadline = undefined;

	// A timer that runs a function has no value to send: the function is what
	// the timer was asked for, and the moment it fires at is not a result. The
	// function runs as a goroutine of its own, which is what the timer is
	// written in Go to mean.
	if (typeof channel.timerCallback === "function") {
		const run = channel.timerCallback;

		delete channel.timerCallback;
		go2jsGo(run);

		return true;
	}

	channel.buffer.push(go2jsTimeValue(new Date(channel.timerDue)));
	go2jsWake(channel);

	// A ticker sends again once its period is up, so the deadline it leaves
	// behind is the next one rather than none at all. A ticker that was stopped
	// has no period left and is done.
	if (channel.timerPeriod !== null && channel.timerPeriod !== undefined) {
		go2jsTimerArm(channel);
	}

	return true;
}

// go2jsTimerArm gives a channel a deadline a period on from now, in the two
// forms the schedule keeps it in: the moment a send carries, and the millis-
// second moment a select compares against the clock.
function go2jsTimerArm(channel) {
	channel.timerDeadline = Date.now() + channel.timerPeriod / 1000000;
	channel.timerDue = go2jsTimeValue(new Date(channel.timerDeadline));
	go2jsTimers.set(channel, channel.timerDue);
}

function go2jsWaitForEarliestTimer() {
	let earliest = Infinity;

	for (const channel of go2jsTimers.keys()) {
		if (channel.timerDeadline < earliest) {
			earliest = channel.timerDeadline;
		}
	}

	if (!isFinite(earliest)) {
		return false;
	}

	const wait = earliest - Date.now();

	if (wait > 0) {
		try {
			require("child_process").execFileSync("sleep", [String(wait / 1000)]);
		} catch (err) {
			// A wait that could not be taken still leaves the deadline to be
			// reached, and the loop comes back round to it.
		}
	}

	return true;
}

// go2jsPark suspends the goroutine that is running until one of the channels
// it names changes in a way that might let its operation through. The task is
// registered on every one of them, so a change to any of them makes it
// runnable again, and the operation is then tried once more from the top.
const go2jsParked = {};

function go2jsPark(channels, kind, value) {
	const task = go2jsRunning;

	if (task === null || task === undefined) {
		throw new Error("go2js: blocking operation outside a goroutine");
	}

	for (const channel of channels) {
		channel.blocked.push(task);
		task.waiting.push(channel);
	}

	task.parked = kind;
	task.parkedValue = value;

	// A goroutine that parks to send or to receive makes a case of a select
	// that is waiting on the same channel ready, so any select parked there is
	// woken to look again. Go hands a value straight from one goroutine to
	// another without either of them waiting again, and a select is a
	// goroutine that is waiting to be given one.
	if (kind === "receiver" || kind === "sender") {
		const selects = [];

		for (const channel of channels) {
			for (const entry of channel.blocked) {
				if (entry.parked === "select") {
					selects.push(entry);
				}
			}
		}

		for (const entry of selects) {
			go2jsRunnable(entry);
		}
	}

	return go2jsParked;
}

// go2jsWake makes every goroutine waiting on the channel runnable, since the
// channel it was waiting on has just changed.
function go2jsWake(channel) {
	if (channel.blocked.length === 0) {
		return;
	}

	const waiting = channel.blocked;

	channel.blocked = [];

	for (const task of waiting) {
		go2jsRunnable(task);
	}
}

function go2jsRunnable(task) {
	if (task.finished || go2jsReady.indexOf(task) >= 0) {
		return;
	}

	for (const channel of task.waiting) {
		const index = channel.blocked.indexOf(task);

		if (index >= 0) {
			channel.blocked.splice(index, 1);
		}
	}

	task.waiting = [];
	task.parked = null;
	go2jsReady.push(task);
}

// go2jsTakeBlocked takes a goroutine that is waiting on the channel for the
// kind of operation named, so that a send can be handed to a receiver or a
// receive answered by a sender without either of them waiting again.
function go2jsTakeBlocked(channel, kind) {
	for (let index = 0; index < channel.blocked.length; index++) {
		const task = channel.blocked[index];

		if (task.parked === kind) {
			channel.blocked.splice(index, 1);
			task.waiting = task.waiting.filter((entry) => entry !== channel);
			return task;
		}
	}

	return null;
}

// A Go function is written as a generator, so calling one is stepping it: the
// call is delegated with yield*, which runs the callee as part of the caller's
// own goroutine and lets either of them wait. A value the compiler could not
// tell the shape of is called through go2jsCall instead, which runs a generator
// and passes through anything else unchanged.

// go2jsGenerator says whether a value handed back from a call is a goroutine
// that has been started but not run.
function go2jsGenerator(value) {
	return value !== null && typeof value === "object" && typeof value.next === "function";
}

function* go2jsCall(fn, self, args) {
	const result = fn.apply(self, args);

	if (go2jsGenerator(result)) {
		return yield* result;
	}

	return result;
}

// go2jsCallNow calls a value the runtime was handed as a callback. The
// callback runs to its end here, and if it waits on something then the
// goroutines that can be run are run until it is through, which is what a
// callback such as a comparison or a visitor has to be given: nothing here
// waits with it, so the wait it makes is one the scheduler answers.
function go2jsCallNow(fn, self, args) {
	const result = fn.apply(self, args);

	if (go2jsGenerator(result)) {
		return go2jsRunNow(result);
	}

	return result;
}

function go2jsRunNow(generator) {
	// A thunk the compiler could not see the shape of may well turn out not to
	// be a generator, and then there is nothing to run: the value it produced is
	// the answer.
	if (!go2jsGenerator(generator)) {
		return generator;
	}

	for (;;) {
		const step = generator.next();

		if (step.done) {
			return step.value;
		}

		if (!go2jsWaitForWork()) {
			throw go2jsFatalError("all goroutines are asleep - deadlock!");
		}
	}
}

function go2jsWaitGroup() {
	return {
		__go2js_channel: true,
		blocked: [],
		capacity: 0,
		buffer: [],
		closed: false,
		count: 0,
		handoff: []
	};
}

function go2jsWaitGroupAdd(group, delta) {
	group.count += delta;

	if (group.count < 0) {
		throw new Error("sync: negative WaitGroup counter");
	}

	// A wait group is reached at zero, which is the moment every goroutine
	// waiting on it is free to carry on.
	if (group.count === 0) {
		go2jsWake(group);
	}
}

function go2jsWaitGroupDone(group) {
	go2jsWaitGroupAdd(group, -1);
}

// A task is one goroutine. The body of every Go function is written as a
// generator, so a task is that generator together with the channels its
// current operation is waiting on. Running the task means giving the generator
// its next step; where it stops is where it waits until a channel changes.
function go2jsTask(generator) {
	return {
		generator,
		waiting: [],
		parked: null,
		parkedValue: undefined,
		answered: false,
		finished: false
	};
}

// go2jsReady holds the goroutines that can be stepped right now, and
// go2jsRunning holds the one being stepped, so that a goroutine that is waiting
// knows what it is waiting as.
const go2jsReady = [];

let go2jsRunning = null;

let go2jsLiveTasks = 0;

function go2jsStep() {
	let steps = 0;

	while (go2jsReady.length > 0) {
		if (steps >= 1000000) {
			go2jsReady.length = 0;
			throw new Error("go2js: goroutine task limit exceeded");
		}

		const task = go2jsReady.shift();

		go2jsRunning = task;

		let step;

		try {
			step = task.generator.next();
		} catch (thrown) {
			go2jsRunning = null;
			task.finished = true;
			go2jsLiveTasks--;

			// A goroutine that was told to end stops there, and the rest of them
			// carry on. Anything else is a failure and belongs to the program, so
			// it is left to travel up.
			if (go2jsIsGoexit(thrown)) {
				steps++;
				continue;
			}

			go2jsReady.length = 0;
			throw thrown;
		}

		go2jsRunning = null;
		steps++;

		if (step.done) {
			task.finished = true;
			go2jsLiveTasks--;
			continue;
		}

		// A step that stopped without parking its goroutine asked for the turn
		// to be handed to another one, so it goes back into the queue.
		if (step.value !== go2jsParked) {
			go2jsReady.push(task);
		}
	}

	return steps;
}

// go2jsWakeDueTimers wakes what is waiting on a timer that has come due. A
// timer is noticed by the goroutine that is waiting on it when that goroutine
// runs, so reaching the moment a timer names is what makes it runnable again.
function go2jsWakeDueTimers() {
	const now = Date.now();

	for (const channel of go2jsTimers.keys()) {
		if (channel.timerDeadline <= now) {
			// A timer that runs a function has nothing waiting on the channel it
			// stands for, so it is fired from here rather than left for a
			// goroutine that may never come to look at it.
			if (typeof channel.timerCallback === "function") {
				go2jsTimerDue(channel);
			} else {
				go2jsWake(channel);
			}
		}
	}
}

// go2jsWaitForWork runs the goroutines that can be run, and if none of them can
// then waits for the moment the earliest timer comes due. It says whether
// there was anything at all to wait for.
function go2jsWaitForWork() {
	if (go2jsStep() > 0) {
		return true;
	}

	if (!go2jsWaitForEarliestTimer()) {
		return false;
	}

	go2jsWakeDueTimers();

	return true;
}

function go2jsStart(generator) {
	const task = go2jsTask(generator);

	go2jsLiveTasks++;
	go2jsReady.push(task);

	return task;
}

// go2jsGo writes "go f(x)": f runs as a goroutine of its own, starting from the
// moment it was written rather than straight away. What it is handed is the
// generator function holding the body, which is started and then left to be
// stepped by the scheduler like any other goroutine.
function go2jsGo(body) {
	go2jsStart(body());
}

// go2jsRunMain runs main as a goroutine too, so that everything else in the
// program is written the same way and the scheduler has one place it all goes.
function go2jsRunMain(generator) {
	const main = go2jsStart(generator);

	while (go2jsLiveTasks > 0) {
		if (go2jsWaitForWork()) {
			continue;
		}

		// Nothing is runnable and nothing is coming due. A program whose main
		// has already returned has ended, as a Go program does, and the
		// goroutines still waiting are left where they are; a program still
		// inside main has every goroutine asleep, which Go calls a deadlock and
		// says so.
		if (main.finished) {
			go2jsReady.length = 0;
			break;
		}

		throw go2jsFatalError("all goroutines are asleep - deadlock!");
	}
}

function* go2jsWaitGroupWait(group) {
	// Waiting on a wait group is parking on it, which the group wakes when the
	// count it holds comes back to zero.
	while (group.count > 0) {
		yield go2jsPark([group], "waitgroup", undefined);
	}
}

function go2jsMutex() {
	return { locked: false, blocked: [] };
}

// go2jsMutexLock takes a mutex, waiting for it when another goroutine is
// holding it. The wait is a park on the mutex itself, which the unlock wakes,
// so a goroutine that has to wait for a lock goes to the back of the queue with
// the goroutines that were waiting before it rather than finding the lock held
// under it.
function* go2jsMutexLock(mutex) {
	while (mutex.locked) {
		yield go2jsPark([mutex], "mutex", undefined);

		// The unlock that woke this goroutine gave the lock to it as it went, so
		// it holds the lock already and has nothing left to wait for.
		if (go2jsRunning.answered) {
			go2jsRunning.answered = false;

			return;
		}
	}

	mutex.locked = true;
}

function go2jsMutexUnlock(mutex) {
	if (!mutex.locked) {
		throw new Error("sync: unlock of unlocked mutex");
	}

	mutex.locked = false;

	// The goroutine that was waiting first for the lock is given it as the lock
	// is let go, so it goes straight from one holder to the next without a
	// moment of it free for anyone else to take.
	const next = go2jsTakeBlocked(mutex, "mutex");

	if (next !== null) {
		mutex.locked = true;
		go2jsAnswered(next);
	}
}

// go2jsAnswered lets a goroutine that was waiting be run again with what it was
// waiting for already granted, so it carries on from where it stopped rather
// than looking for it a second time.
function go2jsAnswered(task) {
	task.parked = null;
	task.parkedValue = undefined;
	task.answered = true;
	go2jsRunnable(task);
}

// go2jsErrorMethodCall invokes an error chain method such as Is or Unwrap on a
// value, a pointer box, an interface holder, or an embedded promotion proxy.
function go2jsErrorMethodCall(err, name, ...args) {
	if (err === null || err === undefined) {
		return undefined;
	}

	let receiver = err;

	if (err.__go2js_pointer === true) {
		receiver = err[go2jsPointerGet]();

		if (receiver === null || receiver === undefined) {
			return undefined;
		}
	}

	if (err.__go2js_interface === true) {
		if (typeof err.type === "string") {
			const fn = go2jsMethodTable[err.type + "." + name];

			if (typeof fn === "function") {
				return go2jsCallNow(fn, null, [err.value, ...args]);
			}
		}

		receiver = err.value;
	}

	if (receiver === null || receiver === undefined) {
		return undefined;
	}

	if (typeof receiver === "object" || typeof receiver === "function") {
		const fn = receiver[name];

		if (typeof fn === "function") {
			return go2jsCallNow(fn, receiver, args);
		}
	}

	return undefined;
}

// go2jsErrorUnwrapAll returns the direct children of an error chain node, which
// covers cause fields, errors.Join results, and Unwrap() returning one or many
// errors.
function go2jsErrorUnwrapAll(err) {
	if (err === null || err === undefined) {
		return [];
	}

	const unwrapped = go2jsErrorMethodCall(err, "Unwrap");

	if (unwrapped !== undefined && unwrapped !== null) {
		return Array.isArray(unwrapped) ? unwrapped.filter(item => item !== null && item !== undefined) : [unwrapped];
	}

	if (err.cause !== undefined && err.cause !== null) {
		return [err.cause];
	}

	if (Array.isArray(err.joined)) {
		return err.joined.filter(item => item !== null && item !== undefined);
	}

	return [];
}

function go2jsErrorsIs(err, target) {
	if (err === target) {
		return true;
	}

	if (err === null || err === undefined || target === null || target === undefined) {
		return err === target;
	}

	if (go2jsErrorMethodCall(err, "Is", target) === true) {
		return true;
	}

	return go2jsErrorUnwrapAll(err).some(part => go2jsErrorsIs(part, target));
}

function go2jsErrorMessage(err) {
	if (err === null || err === undefined) {
		return "<nil>";
	}

	// An error is asked of itself what it says, which a value carrying an Error
	// method of its own is asked the same way, so that joining errors says what
	// each of them says rather than what it is built from.
	return go2jsErrorString(go2jsUnwrap(err));
}

function go2jsDuration(nanoseconds) {
	// a span of time that is already one keeps the shape it has, since giving
	// it a second would bury the first inside it
	if (go2jsIsDuration(nanoseconds)) {
		return nanoseconds;
	}

	return {
		nanoseconds: nanoseconds,
		// the name of the type is kept where a verb asking for it can find it,
		// since the object is its own kind rather than something wrapped
		__go2js_type_name: "time.Duration",
		valueOf() {
			return this.nanoseconds;
		},
		String() {
			return go2jsDurationString(this.nanoseconds);
		},
		Nanoseconds() {
			return this.nanoseconds;
		},
		Microseconds() {
			return Math.trunc(this.nanoseconds / 1000);
		},
		Milliseconds() {
			return Math.trunc(this.nanoseconds / 1000000);
		},
		Seconds() {
			return this.nanoseconds / 1000000000;
		},
		Minutes() {
			return this.nanoseconds / 60000000000;
		},
		Hours() {
			return this.nanoseconds / 3600000000000;
		},
		Abs() {
			return DurationAbs(this);
		},
		Truncate(m) {
			return DurationTruncate(this, m);
		},
		Round(m) {
			return DurationRound(this, m);
		}
	};
}

// go2jsDurationNanos accepts either a duration object or a raw nanosecond count,
// and gives the count back as it is held, since a count with more digits than a
// double keeps is held as digits and must not be brought down to a double here.
function go2jsDurationNanos(value) {
	if (go2jsIsDuration(value)) {
		return value.nanoseconds;
	}

	if (typeof value === "bigint") {
		return value;
	}

	return Number(value);
}

// go2jsDurationNanosAsNumber gives the count back as a double, which is what
// the helpers that divide the count and answer with a fraction need.
function go2jsDurationNanosAsNumber(value) {
	return Number(go2jsDurationNanos(value));
}

function go2jsDurationDecimal(value) {
	if (Number.isInteger(value)) {
		return String(value);
	}

	let text = value.toFixed(9);

	while (text.endsWith("0")) {
		text = text.slice(0, -1);
	}

	if (text.endsWith(".")) {
		text = text.slice(0, -1);
	}

	return text;
}

// A span of time too many digits to be a double is put in words the same way,
// working the digits themselves rather than rounding them away.
function go2jsDurationStringBig(nanoseconds) {
	if (nanoseconds === 0n) {
		return "0s";
	}

	const sign = nanoseconds < 0n ? "-" : "";
	let rest = nanoseconds < 0n ? -nanoseconds : nanoseconds;
	const billion = 1000000000n;
	const minute = 60000000000n;
	const hour = 3600000000000n;

	if (rest < billion) {
		let scale = 1n;
		let unit = "ns";

		if (rest >= 1000000n) {
			scale = 1000000n;
			unit = "ms";
		} else if (rest >= 1000n) {
			scale = 1000n;
			unit = "\u00b5s";
		}

		return sign + go2jsFractionDigits(rest, scale) + unit;
	}

	const hours = rest / hour;
	rest -= hours * hour;

	const minutes = rest / minute;
	rest -= minutes * minute;

	let text = "";

	if (hours > 0n) {
		text += hours + "h";
	}

	if (hours > 0n || minutes > 0n) {
		text += minutes + "m";
	}

	return sign + text + go2jsFractionDigits(rest, billion) + "s";
}

// go2jsFractionDigits writes a fraction of a whole part out the way Go writes
// one, in digits, with the zeros on the right left off and none at all when
// there is nothing to say. The unit is left to the caller, which is the one
// that knows what is being measured.
function go2jsFractionDigits(numerator, denominator) {
	const whole = numerator / denominator;
	const fraction = (numerator % denominator).toString().padStart(denominator.toString().length - 1, "0").replace(/0+$/, "");

	if (fraction === "") {
		return String(whole);
	}

	return String(whole) + "." + fraction;
}

function go2jsDurationString(nanoseconds) {
	if (typeof nanoseconds === "bigint") {
		return go2jsDurationStringBig(nanoseconds);
	}

	if (nanoseconds === 0) {
		return "0s";
	}

	const sign = nanoseconds < 0 ? "-" : "";
	let rest = Math.abs(nanoseconds);

	if (rest < 1000000000) {
		let scale = 1;
		let unit = "ns";

		if (rest >= 1000000) {
			scale = 1000000;
			unit = "ms";
		} else if (rest >= 1000) {
			scale = 1000;
			unit = "\u00b5s";
		}

		return sign + go2jsDurationDecimal(rest / scale) + unit;
	}

	const hours = Math.floor(rest / 3600000000000);
	rest -= hours * 3600000000000;

	const minutes = Math.floor(rest / 60000000000);
	rest -= minutes * 60000000000;

	let text = "";

	if (hours > 0) {
		text += hours + "h";
	}

	if (hours > 0 || minutes > 0) {
		text += minutes + "m";
	}

	return sign + text + go2jsDurationDecimal(rest / 1000000000) + "s";
}

function go2jsErrorString(value) {
	if (value === null || value === undefined) {
		return "<nil>";
	}

	if (value instanceof Error) {
		return value.message;
	}

	if (value.__go2js_interface === true) {
		return go2jsErrorString(value.value);
	}

	if (typeof value === "object" && typeof value.Error === "function") {
		return go2jsCallNow(value.Error, value, []);
	}

	return go2jsStringify(value);
}

function go2jsOperandIsString(value) {
	return typeof value === "string";
}

function go2jsOutputText(value, suffix, separator) {
	if (value === undefined) {
		return "";
	}

	if (Array.isArray(value)) {
		if (separator === undefined || separator === null || separator === "") {
			let out = "";

			for (let index = 0; index < value.length; index++) {
				if (index > 0 && !go2jsOperandIsString(value[index - 1]) && !go2jsOperandIsString(value[index])) {
					out += " ";
				}

				out += go2jsErrorString(value[index]);
			}

			return out + suffix;
		}

		return value.map(item => go2jsErrorString(item)).join(separator) + suffix;
	}

	return go2jsErrorString(value) + suffix;
}

function DurationAbs(d) {
	const span = go2jsDurationNanos(d);

	if (span >= 0) {
		return go2jsDuration(span);
	}

	// the smallest count there is has no positive counterpart inside the width
	// it is held in, so its own absolute value is the largest there is
	if (span === -go2jsSignBit64) {
		return go2jsDuration(go2jsSignBit64 - 1n);
	}

	return go2jsDuration(-span);
}

function DurationTruncate(d, multiple) {
	const step = go2jsDurationNanosAsNumber(multiple);

	if (step === 0) {
		return go2jsDuration(go2jsDurationNanos(d));
	}

	return go2jsDuration(Math.trunc(go2jsDurationNanosAsNumber(d) / step) * step);
}

function DurationRound(d, multiple) {
	const step = go2jsDurationNanosAsNumber(multiple);

	if (step === 0) {
		return go2jsDuration(go2jsDurationNanos(d));
	}

	const value = go2jsDurationNanosAsNumber(d);
	const half = step / 2;
	let offset = value % step;

	if (offset < 0) {
		offset += step;
	}

	const rounded = offset >= half ? value + (step - offset) : value - offset;

	return go2jsDuration(rounded);
}

function go2jsLogOutput(text) {
	go2jsLogWriteTo(go2jsLogStandardWriter, text);
}

function go2jsLogStandardWriterDefault() {
	return process.stderr;
}

function go2jsLogPad(value, width) {
	const text = String(value);

	return text.length >= width ? text : "0".repeat(width - text.length) + text;
}

function go2jsLogStamp(now, flag) {
	let year, month, day, hour, minute, second;

	if ((flag & 32) !== 0) {
		year = now.getUTCFullYear();
		month = now.getUTCMonth() + 1;
		day = now.getUTCDate();
		hour = now.getUTCHours();
		minute = now.getUTCMinutes();
		second = now.getUTCSeconds();
	} else {
		year = now.getFullYear();
		month = now.getMonth() + 1;
		day = now.getDate();
		hour = now.getHours();
		minute = now.getMinutes();
		second = now.getSeconds();
	}

	let stamp = "";

	if ((flag & 7) !== 0) {
		if ((flag & 1) !== 0) {
			stamp += go2jsLogPad(year, 4) + "/" + go2jsLogPad(month, 2) + "/" + go2jsLogPad(day, 2) + " ";
		}

		if ((flag & 6) !== 0) {
			stamp += go2jsLogPad(hour, 2) + ":" + go2jsLogPad(minute, 2) + ":" + go2jsLogPad(second, 2);

			if ((flag & 4) !== 0) {
				stamp += "." + go2jsLogPad((flag & 32) !== 0 ? now.getUTCMilliseconds() : now.getMilliseconds(), 6);
			}

			stamp += " ";
		}
	}

	return stamp;
}

function go2jsLogDecorate(text) {
	const stamp = go2jsLogStamp(new Date(), go2jsLogStandardFlags);

	return (go2jsLogStandardFlags & 64) !== 0 ? "" : go2jsLogCurrentPrefix
		+ stamp
		+ ((go2jsLogStandardFlags & 64) !== 0 ? go2jsLogCurrentPrefix : "")
		+ text;
}

function go2jsLogWriteTo(target, text) {
	const inner = go2jsUnwrap(target);

	if (inner === null || inner === undefined) {
		process.stderr.write(text);
		return;
	}

	if (inner === process.stdout || inner === process.stderr || inner === process.stdin) {
		inner.write(text);
		return;
	}

	if (typeof inner.write === "function") {
		inner.write(text);
		return;
	}

	if (typeof inner.Write === "function") {
		go2jsCallNow(inner.Write, inner, [go2jsStringToBytes(text)]);
		return;
	}

	process.stderr.write(text);
}

var go2jsLogStandardFlags = 3;

var go2jsLogCurrentPrefix = "";

var go2jsLogStandardWriter = go2jsLogStandardWriterDefault();

function go2jsCallMethod(target, method, ...args) {
	if (target === null || target === undefined) {
		throw new TypeError("call of method " + method + " on nil interface");
	}

	if (target.__go2js_interface === true) {
		return go2jsInterfaceCallNow(target, method, ...args);
	}

	const fn = target[method];

	if (typeof fn !== "function") {
		throw new TypeError("method " + method + " is not implemented");
	}

	return go2jsCallNow(fn, target, args);
}

// A sentinel is one value that every name for it stands for, so os.ErrNotExist
// and fs.ErrNotExist compare equal and errors.Is finds one inside a chain. They
// are kept by their message, which is the one thing all the names agree on.
const go2jsSentinelErrors = new Map();

// An errno and the sentinel that names the same condition are two names for one
// thing, so a chain that carries the errno carries the sentinel with it, which
// is what makes errors.Is find the sentinel in the error of a path that is not
// there.
const go2jsErrnoSentinelMessages = new Map([
	["no such file or directory", "file does not exist"],
	["permission denied", "permission denied"],
	["operation not permitted", "permission denied"],
	["file exists", "file already exists"]
]);

// go2jsErrnoSentinel is the errno a code stands for, which is the value a path
// error unwraps to and the value errors.Is matches a sentinel against.
function go2jsErrnoSentinel(code) {
	switch (code) {
	case "ENOENT":
		return go2jsSentinelError("no such file or directory", "syscall.Errno")();
	case "EEXIST":
		return go2jsSentinelError("file exists", "syscall.Errno")();
	case "EACCES":
		return go2jsSentinelError("permission denied", "syscall.Errno")();
	case "EPERM":
		return go2jsSentinelError("operation not permitted", "syscall.Errno")();
	default:
		return null;
	}
}

function go2jsSentinelError(message, typeName) {
	return function go2jsSentinelErrorValue() {
		const name = typeName === undefined || typeName === null || typeName === "" ? "*errors.errorString" : typeName;
		const key = name + "\u0000" + message;
		let error = go2jsSentinelErrors.get(key);

		if (error === undefined) {
			error = go2jsNameError(new Error(message), name);

			const named = go2jsErrnoSentinelMessages.get(message);

			if (named !== undefined) {
				error.Is = function (other) {
					return go2jsErrorMessage(other) === named;
				};
			}

			go2jsSentinelErrors.set(key, error);
		}

		return error;
	};
}

function go2jsIOEOF() {
	return go2jsSentinelError("EOF")();
}

// go2jsNameError records the Go type an error value really has, because a
// translation has one Error for them all and %T would otherwise report the same
// name for every one of them.
function go2jsNameError(err, typeName) {
	if (err !== null && err !== undefined && (typeof err === "object" || typeof err === "function")) {
		err.__go2js_error_name = typeName;
	}

	return err;
}

function go2jsErrorsNew(value) {
	return go2jsNameError(new Error(go2jsStringify(value)), "*errors.errorString");
}

function go2jsTimeNow() {
	const now = go2jsTimeValue(new Date());

	now.__go2js_location = go2jsTimeHostZoneName();
	now.__go2js_zone = now.__go2js_location === "UTC" ? null : now.__go2js_location;

	// A moment made by reading the clock of the machine carries how far the
	// clock has run since the process was born, so that printing it tells the
	// distance into the process the way Go tells it.
	now.__go2js_mono = Math.max(0, Date.now() * 1000000 - go2jsTimeMonoOrigin);

	return now;
}

go2jsRegisterMethod("error.Error", function(value) {
	const inner = value !== null && value !== undefined && value.value !== undefined
		? value.value
		: value;

	return inner instanceof Error ? inner.message : String(inner);
});

go2jsRegisterMethod("error.Unwrap", function() {
	return null;
});

// go2jsRegexpClassCategories lists the one and two letter general category
// codes RE2 accepts. Everything else after \p{ is a script name, which
// JavaScript spells Script=Name.
const go2jsRegexpClassCategories = new Set([
	"L", "Lu", "Ll", "Lt", "Lm", "Lo",
	"M", "Mn", "Mc", "Me",
	"N", "Nd", "Nl", "No",
	"P", "Pc", "Pd", "Ps", "Pe", "Pi", "Pf", "Po",
	"S", "Sm", "Sc", "Sk", "So",
	"Z", "Zl", "Zp", "Zs",
	"C", "Cc", "Cf", "Co", "Cs", "Cn"
]);

function go2jsRegexpPropertyName(name) {
	if (go2jsRegexpClassCategories.has(name)) {
		return "General_Category=" + name;
	}

	return "Script=" + name;
}

// RE2 spells the ASCII classes with escapes, and JavaScript agrees on \d and
// \w but not on \s, which in JavaScript also covers Unicode spaces.
const go2jsRegexpAsciiSpace = "\\t\\n\\f\\r ";

const go2jsRegexpPosixClasses = {
	"[:alpha:]": "A-Za-z",
	"[:digit:]": "0-9",
	"[:alnum:]": "A-Za-z0-9",
	"[:upper:]": "A-Z",
	"[:lower:]": "a-z",
	"[:space:]": "\\t\\n\\f\\r ",
	"[:blank:]": "\\t ",
	"[:punct:]": "!-/:-@\\[-\\x60{-~",
	"[:print:]": "\\x20-\\x7e",
	"[:graph:]": "\\x21-\\x7e",
	"[:cntrl:]": "\\x00-\\x1f\\x7f",
	"[:xdigit:]": "0-9A-Fa-f"
};

function go2jsRegexpPosixClass(name) {
	return go2jsRegexpPosixClasses[name];
}

// go2jsRegexpASCIIClasses rewrites the classes where RE2 and JavaScript
// disagree, tracking whether the pattern is inside a bracket expression so the
// replacement can omit the brackets.
function go2jsRegexpASCIIClasses(source) {
	let out = "";
	let inClass = false;

	for (let i = 0; i < source.length; i++) {
		const ch = source[i];

		if (ch === "\\" && i + 1 < source.length) {
			const next = source[i + 1];

			if (next === "s" || next === "S") {
				const body = go2jsRegexpAsciiSpace;

				if (next === "S") {
					out += inClass ? "\\\\S" : "[^" + body + "]";
				} else {
					out += inClass ? body : "[" + body + "]";
				}

				i++;
				continue;
			}

			if (next === "d" || next === "D" || next === "w" || next === "W") {
				// The u flag already keeps \d and \w ASCII only.
				out += ch + next;
				i++;
				continue;
			}

			out += ch + next;
			i++;
			continue;
		}

		if (ch === "[") {
			// [[:alpha:]] is a bracket expression holding a POSIX class.
			if (source[i + 1] === "[" && source[i + 2] === ":") {
				const close = source.indexOf(":]" + "]", i + 3);
				const name = close === -1 ? "" : source.slice(i + 1, close + 2);
				const body = go2jsRegexpPosixClass(name);

				if (body !== undefined) {
					out += "[" + body + "]";
					i = close + 2;
					continue;
				}
			}

			inClass = true;
			out += ch;
			continue;
		}

		if (ch === "]" && inClass) {
			inClass = false;
			out += ch;
			continue;
		}

		out += ch;
	}

	return out;
}

// go2jsRegexpPattern rewrites RE2 constructs that JavaScript does not accept.
function go2jsRegexpPattern(pattern) {
	let source = go2jsStringify(pattern);
	let flags = "";

	// RE2 gives an error to the Perl tricks that JavaScript would quietly take:
	// a reference to a numbered group, or a look that goes ahead or behind.
	const refused = source.match(/\\[0-9]|\(\?P=|\(\?[=!<>|(]/);

	if (refused) {
		const offending = refused[0];

		throw new Error(offending.startsWith("\\") ? "invalid escape sequence: " + offending : "invalid or unsupported Perl syntax: " + offending);
	}

	// Leading and embedded flag groups such as (?i) become regexp flags.
	source = source.replace(/\(\?([imsU]+)\)/g, (all, group) => {
		for (const flag of group) {
			if (flag === "i" || flag === "m" || flag === "s") {
				if (!flags.includes(flag)) {
					flags += flag;
				}
			}
		}

		return "";
	});

	// (?P<name>re) is RE2 syntax for the named group (?<name>re).
	source = source.replace(/\(\?P</g, "(?<");
	source = source.replace(/\\A/g, "^").replace(/\\z/g, "$").replace(/\\Z/g, "$");
	// RE2 also accepts the single letter shorthand \pL for \p{L}.
	source = source.replace(/\\([pP])([A-Za-z])(?![A-Za-z{])/g, (all, kind, name) => "\\" + kind + "{" + name + "}");
	source = source.replace(/\\([pP])\{([^}=]+)\}/g, (all, kind, name) => "\\" + kind + "{" + go2jsRegexpPropertyName(name) + "}");
	source = go2jsRegexpASCIIClasses(source);

	return {source: source, flags: flags};
}

function go2jsRegexpNewRegExp(pattern, flags) {
	const parsed = go2jsRegexpPattern(pattern);

	if (flags === undefined || flags === null) {
		flags = "";
	}

	let merged = "u" + parsed.flags;

	for (const flag of flags) {
		if (!merged.includes(flag)) {
			merged += flag;
		}
	}

	return new RegExp(parsed.source, merged);
}

function go2jsRegexpSubmatchNames(pattern) {
	const source = go2jsRegexpPattern(pattern).source;

	// The whole match is the first name a pattern has, and it has none, and
	// every group after it contributes a name whether it was given one or not.
	const names = [""];
	let inClass = false;

	for (let index = 0; index < source.length; index++) {
		const ch = source[index];

		if (ch === "\\") {
			index++;
			continue;
		}

		if (inClass) {
			if (ch === "]") {
				inClass = false;
			}

			continue;
		}

		if (ch === "[") {
			inClass = true;
			continue;
		}

		if (ch !== "(") {
			continue;
		}

		// A group that opens with a mark is one of the forms that does not
		// capture, unless it is a mark followed by the name of a capture.
		if (source[index + 1] !== "?") {
			names.push("");
			continue;
		}

		const named = /^\(\?P?<([A-Za-z_][A-Za-z0-9_]*)>/.exec(source.slice(index));

		if (named !== null) {
			names.push(named[1]);
		}
	}

	return names;
}

function go2jsRegexpFindAllString(pattern, value, limit) {
	const source = go2jsStringify(value);
	const regex = go2jsRegexpNewRegExp(pattern, "g");
	const out = [];

	for (;;) {
		const match = regex.exec(source);

		if (match === null) {
			break;
		}

		out.push(match[0]);

		if (limit !== undefined && limit >= 0 && out.length >= limit) {
			break;
		}
	}

	return out;
}

function go2jsUtf8Length(codePoint) {
	if (codePoint < 0x80) {
		return 1;
	}

	if (codePoint < 0x800) {
		return 2;
	}

	if (codePoint < 0x10000) {
		return 3;
	}

	return 4;
}

// go2jsByteOffsetMap converts a JavaScript UTF-16 index into the UTF-8 byte
// offset Go reports. JavaScript strings are UTF-16, so any match that touches a
// non ASCII rune is offset by a different amount than Go would report.
function go2jsByteOffsetMap(text) {
	let ascii = true;

	for (let i = 0; i < text.length; i++) {
		if (text.charCodeAt(i) > 0x7f) {
			ascii = false;
			break;
		}
	}

	// Plain ASCII text needs no translation at all.
	if (ascii) {
		return (index) => index;
	}

	const offsets = new Int32Array(text.length + 1);
	let byte = 0;
	let i = 0;

	while (i < text.length) {
		const codePoint = text.codePointAt(i);
		const width = codePoint > 0xffff ? 2 : 1;

		offsets[i] = byte;
		byte += go2jsUtf8Length(codePoint);

		if (width === 2) {
			// A match can never start or end between the two halves of a
			// surrogate pair, but the slot still has to hold a value.
			offsets[i + 1] = byte;
		}

		i += width;
	}

	offsets[text.length] = byte;

	return (index) => (index >= 0 && index < offsets.length ? offsets[index] : -1);
}

function go2jsRegexpIndex(match, offset) {
	if (match === null || match === undefined) {
		return null;
	}

	if (offset === undefined) {
		offset = (index) => index;
	}

	return [offset(match.index), offset(match.index + match[0].length)];
}

function go2jsRegexpSubmatch(match) {
	// A match without capture groups still reports the whole match.
	if (match === null || match === undefined) {
		return null;
	}

	const out = [match[0]];

	for (let i = 1; i < match.length; i++) {
		out.push(match[i] === undefined ? "" : match[i]);
	}

	return out;
}

function go2jsRegexpSubmatchIndex(match, offset) {
	// Requires a regexp built with the "d" flag so capture offsets are available.
	if (match === null || match === undefined) {
		return null;
	}

	if (offset === undefined) {
		offset = (index) => index;
	}

	const out = [offset(match.index), offset(match.index + match[0].length)];
	const indices = match.indices || [];

	for (let i = 1; i < match.length; i++) {
		if (indices[i] === undefined) {
			out.push(-1, -1);
			continue;
		}

		out.push(offset(indices[i][0]), offset(indices[i][1]));
	}

	return out;
}

function go2jsRegexpByteSubmatch(match) {
	const groups = go2jsRegexpSubmatch(match);

	if (groups === null) {
		return null;
	}

	return groups.map((group) => go2jsStringToBytes(group));
}

function go2jsRegexpByteSubmatchIndex(match, offset) {
	// Byte and string variants share the offsets, because the subject was
	// decoded from the same bytes the caller passed in.
	return go2jsRegexpSubmatchIndex(match, offset);
}

function go2jsRegexpFindAllSubmatch(pattern, value, limit) {
	const source = go2jsBytesToString(value);
	const regex = go2jsRegexpNewRegExp(pattern, "g");
	const out = [];

	for (;;) {
		const match = regex.exec(source);

		if (match === null) {
			break;
		}

		out.push(go2jsRegexpSubmatch(match));

		if (limit !== undefined && limit >= 0 && out.length >= limit) {
			break;
		}
	}

	return out;
}

function go2jsRegexpFindAllSubmatchIndex(pattern, value, limit) {
	const source = go2jsBytesToString(value);
	const offset = go2jsByteOffsetMap(source);
	const regex = go2jsRegexpNewRegExp(pattern, "gd");
	const out = [];

	for (;;) {
		const match = regex.exec(source);

		if (match === null) {
			break;
		}

		out.push(go2jsRegexpSubmatchIndex(match, offset));

		if (limit !== undefined && limit >= 0 && out.length >= limit) {
			break;
		}
	}

	return out;
}

// go2jsRegexpFindAllIndex backs FindAllStringIndex and its byte counterparts.
function go2jsRegexpFindAllIndex(pattern, value, limit) {
	const source = go2jsBytesToString(value);
	const offset = go2jsByteOffsetMap(source);
	const regex = go2jsRegexpNewRegExp(pattern, "g");
	const out = [];

	for (;;) {
		const match = regex.exec(source);

		if (match === null) {
			break;
		}

		out.push(go2jsRegexpIndex(match, offset));

		if (limit !== undefined && limit >= 0 && out.length >= limit) {
			break;
		}
	}

	return out;
}

function go2jsRegexpFindAllByteSubmatch(pattern, value, limit) {
	return go2jsRegexpFindAllSubmatch(pattern, value, limit).map((groups) =>
		groups.map((group) => go2jsStringToBytes(group)));
}

function go2jsRegexpReplaceAllString(pattern, value, replacement) {
	return go2jsStringify(value).replace(
		go2jsRegexpNewRegExp(pattern, "g"),
		go2jsRegexpExpand(go2jsStringify(replacement), pattern)
	);
}

function go2jsRegexpExpand(replacement, pattern) {
	let text = go2jsStringify(replacement)
		.replace(/\$(\d+)/g, (match, index) => "$" + (Number(index) === 0 ? "&" : index))
		.replace(/\$\{(\w+)\}/g, (match, name) => "$" + (name === "0" ? "&" : name));

	// Go spells a named group as $name, while JavaScript needs $<name>.
	if (typeof pattern === "string" && pattern !== "") {
		const names = go2jsRegexpSubmatchNames(pattern);

		for (const name of names) {
			if (name === "" || /^\d+$/.test(name)) {
				continue;
			}

			text = text.split("$" + name).join("$<" + name + ">");
		}
	}

	return text;
}

function go2jsRegexpSplit(pattern, value, limit) {
	const text = go2jsBytesToString(value);

	if (limit === 0) {
		return [];
	}

	// A positive limit keeps the trailing remainder in one piece, so the split
	// is done by hand instead of with String.prototype.split.
	const regex = go2jsRegexpNewRegExp(pattern, "gd");
	const parts = [];

	let last = 0;
	let match = regex.exec(text);

	while (match !== null) {
		const start = match.indices[0][0];
		const end = match.indices[0][1];

		parts.push(text.slice(last, start));

		if (limit > 0 && parts.length === limit - 1) {
			parts.push(text.slice(end));
			return parts;
		}

		last = end;
		match = regex.exec(text);
	}

	parts.push(text.slice(last));

	return parts;
}

function go2jsRawText(value) {
	if (typeof value === "string") {
		return value;
	}

	if (value === null || value === undefined) {
		return "";
	}

	if (value.__go2js_interface === true) {
		return go2jsRawText(value.value);
	}

	if (value.__go2js_pointer === true) {
		return go2jsRawText(value[go2jsPointerGet]());
	}

	if (value.__go2js_text !== undefined) {
		return value.__go2js_text;
	}


	if (Array.isArray(value)) {
		return go2jsBytesToString(value);
	}

	if (value instanceof Uint8Array) {
		return new TextDecoder().decode(value);
	}

	if (typeof value.String === "function") {
		return value.String();
	}

	return String(value);
}

function go2jsBytesToString(value) {
	if (typeof value === "string") {
		return value;
	}

	if (value === null || value === undefined) {
		return "";
	}

	if (Array.isArray(value)) {
		return go2jsDecodeBytes(Uint8Array.from(value.map(item => Number(item) & 255)));
	}

	if (value instanceof Uint8Array) {
		return go2jsDecodeBytes(value);
	}

	if (typeof value.String === "function") {
		return value.String();
	}

	return String(value);
}

// go2jsReflectTypeLen reports the length of a type, which is the one kind of
// type that has a length of its own rather than one that was handed to it. A
// type is not a value, so a slice or a string has none to report either.
function go2jsReflectTypeLen(receiver) {
	const type = go2jsReflectValueType(receiver);

	if (type.kind !== "array" || typeof type.len !== "number") {
		throw new TypeError("reflect: Len of non-array type " + go2jsReflectTypeString(type));
	}

	return type.len;
}

// go2jsReflectLazyList gives a descriptor a list it has to go and look up, and
// works it out the first time something reads it. The list is kept once it has
// been found, and a lookup that finds nothing is left to be tried again, which
// is what lets a descriptor be written out before the registrations it reads.
function go2jsReflectLazyList(target, name, lookup) {
	let value = null;
	let found = false;

	Object.defineProperty(target, name, {
		get: function() {
			if (!found) {
				const listed = lookup(target.name);

				if (listed !== null && listed !== undefined) {
					value = listed;
					found = true;
				}
			}

			return found ? value : null;
		},
		set: function(next) {
			value = next;
			found = true;
		},
		enumerable: true,
		configurable: true
	});
}

function go2jsReflectType(go2jsKind, go2jsName, go2jsElem, go2jsKey, go2jsMethods, go2jsLen) {
	const type = {
		__go2js_reflectType: true,
		kind: go2jsKind,
		name: go2jsName || "",
		elem: go2jsElem || null,
		key: go2jsKey || null,
		methods: go2jsMethods || [],
		len: typeof go2jsLen === "number" ? go2jsLen : null
	};

	// A descriptor is usually written out at the top of the generated program,
	// which is above the registrations that say what a named type is made of, so
	// what a descriptor needs to know about a named type is looked up when it is
	// wanted rather than when it is written. A lookup that comes back with
	// nothing is tried again rather than remembered, because the registration it
	// was looking for has yet to run.
	if (go2jsKind === "struct") {
		go2jsReflectLazyList(type, "fields", go2jsReflectFieldsOfName);
	}

	if (!Array.isArray(go2jsMethods) || go2jsMethods.length === 0) {
		go2jsReflectLazyList(type, "methods", go2jsReflectMethodsOfName);
	}

	// A type writes itself as the name it has, so a program that prints one, or
	// hands it to a verb that takes a string, is given that name rather than the
	// parts of the descriptor. It is kept out of the way of the fields the
	// descriptor is read for, which are its own.
	Object.defineProperty(type, "String", {
		value: () => go2jsReflectTypeString(type),
		enumerable: false,
		writable: true,
		configurable: true
	});

	return type;
}

// go2jsReflectMethodsOf reports the method set a type descriptor carries. It is
// what Implements compares, so a type that never went through the emitter and
// was read off a value instead has an empty set of them.
function go2jsReflectMethodsOf(type) {
	if (type === null || type === undefined || !Array.isArray(type.methods)) {
		return [];
	}

	return type.methods;
}

// go2jsReflectPtrTo reports the pointer to a type, the way reflect.PtrTo does.
// A pointer type has no name of its own, and every method of the type it points
// at is in its method set, so the methods travel across unchanged.
function go2jsReflectPtrTo(type) {
	return go2jsReflectType("ptr", "", type, null, go2jsReflectMethodsOf(type));
}

// go2jsReflectTypeImplements reports whether a type carries every method an
// interface asks for. Go compares the two method sets by name, because one type
// cannot have two methods of the same name, so the comparison here is the same.
function go2jsReflectTypeImplements(type, iface) {
	const wanted = go2jsReflectMethodsOf(iface);

	if (wanted.length === 0) {
		// Every type answers an interface that asks for nothing.
		return true;
	}

	const have = go2jsReflectMethodsOf(type);

	return wanted.every((name) => have.indexOf(name) >= 0);
}

function go2jsReflectTypeKind(value) {
	if (value === null || value === undefined) {
		return "invalid";
	}

	if (typeof value === "boolean") {
		return "bool";
	}

	if (typeof value === "string") {
		return "string";
	}

	if (typeof value === "number") {
		return Number.isInteger(value) ? "int" : "float64";
	}

	if (typeof value === "function") {
		return "func";
	}

	if (Array.isArray(value)) {
		return "slice";
	}

	// A map is checked before the pointer cell, because go2jsNativeMap builds on
	// Map and so carries the get and set methods a cell is recognised by.
	if (value instanceof go2jsNativeMap) {
		return "map";
	}

	if (go2jsPointerAccessor(value, go2jsPointerGet) && go2jsPointerAccessor(value, go2jsPointerSet)) {
		// A map or a slice that reached reflect by reference is handed over in a
		// cell so that a write reaches the storage its holder owns, and that cell
		// is not a pointer. A real one says so, and only a real one is a pointer.
		if (value.__go2js_pointer !== true) {
			const target = go2jsStripWrappers(value[go2jsPointerGet]());

			if (target instanceof go2jsNativeMap) {
				return "map";
			}

			if (Array.isArray(target)) {
				return "slice";
			}
		}

		return "ptr";
	}

	if (value.__go2js_interface === true) {
		return "interface";
	}

	if (value instanceof Error) {
		return "error";
	}

	if (typeof value === "object") {
		return "struct";
	}

	return "invalid";
}

function go2jsReflectNameOfValue(value) {
	if (value === null || value === undefined) {
		return "";
	}

	if (value instanceof Error) {
		return "error";
	}

	if (Array.isArray(value)) {
		return "";
	}

	if (typeof value === "object" && value.constructor && value.constructor.name &&
		!value.constructor.name.startsWith("go2js")) {
		// Go spells a type with the package it belongs to in front of it, and the
		// emitter recorded that spelling, so it is the one to report.
		return go2jsRegisteredTypeName(value.constructor.name);
	}

	return "";
}

function go2jsReflectTypeOf(value, type) {
	if (type !== null && type !== undefined) {
		return type;
	}

	// A value that came through an interface is wrapped in the type it was
	// given there, and it is the value underneath that says what it is.
	const bare = go2jsStripWrappers(value);
	const name = go2jsReflectNameOfValue(bare);
	const descriptor = go2jsReflectType(go2jsReflectTypeKind(bare), name, null, null, go2jsReflectMethodsOfName(name));

	// A descriptor read off a value has to say which fields the value has, or
	// NumField and Field have nothing to answer with, because nothing else about
	// the descriptor names them. The emitter's list is the one to believe, and
	// the value settles it when the type is not one the emitter described.
	if (descriptor.kind === "struct" && !Array.isArray(descriptor.fields)) {
		descriptor.fields = go2jsReflectFieldDescriptorsOf(bare);
	}

	return descriptor;
}

// go2jsReflectDeclaredType builds a descriptor for a type the program said out
// loud, which is the static type of a field rather than the type of a value.
// The empty interface is a type here rather than the absence of one, because a
// field of interface type is an interface whatever it happens to be holding.
function go2jsReflectDeclaredType(name) {
	if (typeof name !== "string") {
		return null;
	}

	const text = name.trim();

	if (text === "") {
		return null;
	}

	if (go2jsIsInterfaceTypeName(text)) {
		return go2jsReflectType("interface", "", null, null, []);
	}

	if (text[0] === "*") {
		const elem = go2jsReflectDeclaredType(text.slice(1));

		return elem === null ? null : go2jsReflectType("ptr", "", elem, null, go2jsReflectMethodsOfName(text));
	}

	if (text.startsWith("[]")) {
		const elem = go2jsReflectDeclaredType(text.slice(2));

		return elem === null ? null : go2jsReflectType("slice", "", elem, null, []);
	}

	if (text[0] === "[") {
		const end = text.indexOf("]");
		const length = Number(text.slice(1, end));

		if (end > 1 && Number.isInteger(length) && length >= 0) {
			const elem = go2jsReflectDeclaredType(text.slice(end + 1));

			return elem === null ? null : go2jsReflectType("array", "", elem, null, [], length);
		}
	}

	if (text.startsWith("map[")) {
		const end = text.indexOf("]");

		if (end > 4) {
			const key = go2jsReflectDeclaredType(text.slice(4, end));
			const elem = go2jsReflectDeclaredType(text.slice(end + 1));

			if (key !== null && elem !== null) {
				return go2jsReflectType("map", "", elem, key, []);
			}
		}
	}

	if (text.startsWith("chan ")) {
		const elem = go2jsReflectDeclaredType(text.slice(5));

		return elem === null ? null : go2jsReflectType("chan", "", elem, null, []);
	}

	if (text.startsWith("func(") && text.endsWith(")")) {
		return go2jsReflectType("func", "", null, null, []);
	}

	if (typeof go2jsReflectBasicTypes[text] === "string") {
		return go2jsReflectType(go2jsReflectBasicTypes[text], text, null, null, []);
	}

	// A name that is not a type written out in full is the name of a named
	// type, and what kind of type it is was recorded when the program said so.
	if (/^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*$/.test(text)) {
		return go2jsReflectType(go2jsReflectTypeKindOfName(text), text.slice(text.lastIndexOf(".") + 1), null, null, go2jsReflectMethodsOfName(text));
	}

	return null;
}

// go2jsReflectFieldTypeOf builds the type a struct says one of its fields is.
// The kind is there to fall back on, because a type written out in full says
// more than the kind does and a name the emitter never described leaves the
// kind as the only thing left to go by.
function go2jsReflectFieldTypeOf(described) {
	if (described === null || described === undefined) {
		return null;
	}

	const declared = go2jsReflectDeclaredType(described.Type);
	const kind = typeof described.Kind === "string" && described.Kind !== "" ? described.Kind : "";

	if (declared !== null) {
		// The kind recorded beside the type is the one the program said, and it
		// settles a name that was registered without saying what kind of type it
		// is, which is what a named interface comes down to: it has a name and a
		// method set, and nothing that says whether it is a struct or a slice.
		if (declared.kind === "invalid" && kind !== "" && kind !== "invalid") {
			declared.kind = kind;
		}

		return declared;
	}

	return kind !== "" && kind !== "invalid" ? go2jsReflectType(kind, "", null, null, []) : null;
}

// go2jsReflectMethodsOfName looks up the method set the emitter recorded for a
// named type. A type that was read off a value rather than off the program has
// to be told what it can do, because nothing else about it says.
function go2jsReflectMethodsOfName(name) {
	if (typeof name !== "string" || name === "") {
		return [];
	}

	const short = name.slice(name.lastIndexOf(".") + 1);

	for (const registered of Object.keys(go2jsMethodSets)) {
		if (registered === name || registered.slice(registered.lastIndexOf(".") + 1) === short) {
			return go2jsMethodSets[registered];
		}
	}

	return [];
}

// go2jsReflectClassOfName reports the class a name was registered for. A name
// is looked up whole first and then by the last part of it, because a descriptor
// read off a value holds the bare name while the registry is keyed by the
// qualified one.
function go2jsReflectClassOfName(name) {
	if (typeof name !== "string" || name === "") {
		return null;
	}

	if (typeof go2jsTypeNames[name] === "function") {
		return go2jsTypeNames[name];
	}

	const short = name.slice(name.lastIndexOf(".") + 1);

	for (const registered of Object.keys(go2jsTypeNames)) {
		if (registered.slice(registered.lastIndexOf(".") + 1) === short) {
			return go2jsTypeNames[registered];
		}
	}

	return null;
}

// go2jsReflectTypeKindsOfName reports the kind a name was registered under. A
// name is looked up whole first and then by the last part of it, which is the
// same way a method set is found, so both agree on which type a name means.
function go2jsReflectTypeKindOfName(name) {
	if (typeof name !== "string" || name === "") {
		return "invalid";
	}

	if (typeof go2jsTypeKinds[name] === "string") {
		return go2jsTypeKinds[name];
	}

	const short = name.slice(name.lastIndexOf(".") + 1);

	for (const registered of Object.keys(go2jsTypeKinds)) {
		if (registered.slice(registered.lastIndexOf(".") + 1) === short) {
			return go2jsTypeKinds[registered];
		}
	}

	return "invalid";
}

// go2jsReflectFieldsOfName reports the field names a name was registered with,
// looked up the way the method set is.
function go2jsReflectFieldsOfName(name) {
	if (typeof name !== "string" || name === "") {
		return null;
	}

	if (Array.isArray(go2jsStructFields[name])) {
		return go2jsStructFields[name];
	}

	const short = name.slice(name.lastIndexOf(".") + 1);

	for (const registered of Object.keys(go2jsStructFields)) {
		if (registered.slice(registered.lastIndexOf(".") + 1) === short) {
			return go2jsStructFields[registered];
		}
	}

	return null;
}

// go2jsReflectFieldNamesOf reduces a list of field descriptors to the names
// alone, which is what a program that only counts or indexes fields needs.
function go2jsReflectFieldNamesOf(fields) {
	return fields.map((field) => (typeof field === "string" ? field : field.Name));
}

// go2jsReflectFieldDescriptorsOf describes the fields of a value the way the
// emitter described the type they belong to, falling back on what the value
// itself carries for a type the emitter said nothing about.
function go2jsReflectFieldDescriptorsOf(value) {
	const names = go2jsReflectStructFieldNames(value);

	return names.map((name) => ({Name: name, Tag: "", PkgPath: "", Anonymous: false}));
}

// go2jsReflectBasicTypes is the set of type names that are written out the
// same way everywhere they appear, so a name is read as the kind it names.
const go2jsReflectBasicTypes = {
	bool: "bool",
	string: "string",
	int: "int",
	int8: "int8",
	int16: "int16",
	int32: "int32",
	int64: "int64",
	uint: "uint",
	uint8: "uint8",
	uint16: "uint16",
	uint32: "uint32",
	uint64: "uint64",
	uintptr: "uintptr",
	float32: "float32",
	float64: "float64",
	complex64: "complex64",
	complex128: "complex128",
	byte: "uint8",
	rune: "int32"
};

function go2jsReflectValue(value, type) {
	return {__go2js_reflectValue: true, v: value, t: type || null};
}

// go2jsReflectRead returns the current storage of a value. An addressable value
// reads through its getter, so a write made by Set is visible to the readers.
function go2jsReflectRead(receiver) {
	if (!go2jsReflectValueBox(receiver)) {
		return receiver;
	}

	// A reflect box carries its own get and set, which are plain methods on a plain
// object rather than the accessors of a pointer, so they are called by name.
	if (typeof receiver.get === "function") {
		return receiver.get();
	}

	return receiver.v;
}

// go2jsReflectAddressable wraps a New cell so that writes to it reach the
// storage the cell points at, which is how a pointer to a value behaves.
function go2jsReflectAddressable(cell, type) {
	return {
		__go2js_reflectValue: true,
		v: cell.target,
		t: type || null,
		get: function() {
			return cell.target;
		},
		set: function(next) {
			cell.target = next;
		}
	};
}

function go2jsReflectValueBox(value) {
	return value !== null && value !== undefined && value.__go2js_reflectValue === true;
}

function go2jsReflectUnwrap(value) {
	if (!go2jsReflectValueBox(value)) {
		return value;
	}

	// A box standing over a slot holds the value as it stood when the box was
	// made, and a write through that box can have moved it since, so the getter
	// is asked when there is one. What is left is the value of a box that is not
	// standing over a slot at all, which is all a box over a plain value has.
	return typeof value.get === "function" ? value.get() : value.v;
}

function go2jsReflectKind(receiver) {
	// Kind is shared by reflect.Type and reflect.Value, and a type descriptor
	// already carries its kind.
	if (receiver !== null && receiver !== undefined && receiver.__go2js_reflectType === true) {
		return receiver.kind;
	}

	if (!go2jsReflectValueBox(receiver)) {
		return go2jsReflectTypeKind(receiver);
	}

	if (receiver.t !== null && receiver.t !== undefined) {
		return receiver.t.kind;
	}

	return go2jsReflectTypeKind(receiver.v);
}

function go2jsReflectValueType(receiver) {
	// A type descriptor is already a type, so deriving one from it would only
	// describe the descriptor object itself.
	if (receiver !== null && receiver !== undefined && receiver.__go2js_reflectType === true) {
		return receiver;
	}

	if (!go2jsReflectValueBox(receiver)) {
		return go2jsReflectTypeOf(receiver);
	}

	if (receiver.t !== null && receiver.t !== undefined) {
		return receiver.t;
	}

	return go2jsReflectTypeOf(receiver.v);
}

// go2jsRegisteredTypeName finds the registered spelling of a bare type name.
function go2jsRegisteredTypeName(name) {
	for (const registered of Object.keys(go2jsTypeNames)) {
		const short = registered.slice(registered.lastIndexOf(".") + 1);

		if (short === name) {
			return registered;
		}
	}

	return name;
}

// go2jsReflectStructFieldNames lists the fields of a struct value in
// declaration order, which is the order reflect reports them in.
function go2jsReflectStructFieldNames(value) {
	if (value === null || value === undefined || typeof value !== "object") {
		return [];
	}

	if (value.__go2js_reflectValue === true) {
		return go2jsReflectStructFieldNames(value.v);
	}

	if (value.__go2js_reflectNew === true) {
		return go2jsReflectStructFieldNames(value.target);
	}

	const ctor = value.constructor;

	if (ctor && typeof ctor.name === "string") {
		// The emitter records the fields of every type it emits, which is the
		// list reflect has to report even for a value that has none of them set.
		const registered = go2jsStructFields[go2jsRegisteredTypeName(ctor.name)];

		if (Array.isArray(registered)) {
			return go2jsReflectFieldNamesOf(registered);
		}

		const fields = go2jsStructFormats[ctor.name];

		if (Array.isArray(fields)) {
			return fields.map((field) => (typeof field === "string" ? field : field.name));
		}
	}

	return Object.keys(value);
}

function go2jsReflectFieldAt(value, index) {
	// A Value box, a New box and a bare struct all stand for the same storage,
	// so the field list is read through whichever wrapper is present.
	let target = value;

	if (target !== null && target !== undefined && target.__go2js_reflectValue === true) {
		target = target.v;
	}

	if (go2jsReflectIsNewBox(target)) {
		target = target.target;
	}

	const names = go2jsReflectStructFieldNames(target);

	if (index < 0 || index >= names.length) {
		throw new RangeError("reflect: Field index out of range");
	}

	// A field is reached through a getter and a setter so that Set on it writes
	// into the storage the caller owns, the way a Go pointer to a field does.
	const slot = names[index];

	return {
		name: slot,
		value: target[slot],
		get: function() {
			return target[slot];
		},
		set: function(next) {
			target[slot] = next;
		}
	};
}

function go2jsReflectValueNumField(receiver) {
	// The wrappers a value travels through are not fields of it, so the count
	// has to be taken from the value itself, the same way the type of it is.
	return go2jsReflectStructFieldNames(go2jsStripWrappers(go2jsReflectUnwrap(receiver))).length;
}

// go2jsReflectFieldTypeAt reports the type a struct's own descriptor says one
// of its fields is, and reports nothing at all for a struct the emitter never
// described, which is a value whose type was read off it rather than off the
// program.
function go2jsReflectFieldTypeAt(receiver, index) {
	if (!go2jsReflectValueBox(receiver)) {
		return null;
	}

	const fields = receiver.t !== null && receiver.t !== undefined ? receiver.t.fields : null;

	if (!Array.isArray(fields) || index < 0 || index >= fields.length) {
		return null;
	}

	return go2jsReflectFieldTypeOf(fields[index]);
}

function go2jsReflectValueField(receiver, index) {
	// The field is one of the value's own, so the wrappers it travelled through
	// are stepped over rather than read as fields of their own.
	const field = go2jsReflectFieldAt(go2jsStripWrappers(go2jsReflectUnwrap(receiver)), index);

	// A field is the type its own struct says it is, and the value standing in
	// it says nothing at all when that value is the zero of a type that reads
	// as nothing, which is what a nil slice and a nil map both are. So the type
	// the struct declared is asked first, and the value settles it only for a
	// struct that was described by no type of its own.
	const type = go2jsReflectFieldTypeAt(receiver, index) || go2jsReflectTypeOf(field.value);

	return {
		__go2js_reflectValue: true,
		get: field.get,
		set: field.set,
		v: field.value,
		t: type
	};
}

function go2jsReflectValueIndex(receiver, index) {
	const target = go2jsReflectRead(receiver);
	const at = Math.trunc(Number(index));

	// An element of a slice or an array is reached by way of the container it
	// stands in, so a write to it goes back the same way. A box holding only the
	// element would be a copy, and a program could write to it all afternoon
	// without the container ever hearing of it, which is the one thing an
	// element of a slice in Go is not.
	if (target === null || target === undefined || typeof target !== "object") {
		const at2 = target !== null && target !== undefined && target[at] !== undefined ? target[at] : target.at(at);

		return go2jsReflectValue(at2, go2jsReflectTypeOf(at2));
	}

	// An element of a slice view is not a place in the array the view was cut
	// from, because the view is a run of elements handed back on its own. The
	// element is therefore written by putting the run back with the element
	// changed, which is the only way through to the storage behind it.
	if (receiver !== null && receiver !== undefined && receiver.__go2js_sliceView === true) {
		const elementOf = receiver.t !== null && receiver.t !== undefined ? go2jsReflectElem(receiver.t) : null;

		return {
			__go2js_reflectValue: true,
			get: function() {
				return go2jsReflectRead(receiver)[at];
			},
			set: function(next) {
				const whole = go2jsReflectRead(receiver).slice();

				whole[at] = next;
				receiver.set(whole);
			},
			v: target[at],
			t: elementOf !== null ? elementOf : go2jsReflectTypeOf(target[at])
		};
	}

	// The container says what its elements are, and the element says what it is
	// holding, which is nothing at all when the element is the zero of a type
	// that reads as nothing.
	const element = receiver !== null && receiver !== undefined && receiver.t !== null && receiver.t !== undefined
		? go2jsReflectElem(receiver.t)
		: null;

	return {
		__go2js_reflectValue: true,
		get: function() {
			return target[at];
		},
		set: function(next) {
			target[at] = next;
		},
		v: target[at],
		t: element !== null ? element : go2jsReflectTypeOf(target[at])
	};
}

function go2jsReflectValueNumMethod() {
	return 0;
}

function go2jsReflectValueIsValid(receiver) {
	if (!go2jsReflectValueBox(receiver)) {
		return receiver !== null && receiver !== undefined;
	}

	return receiver.v !== null && receiver.v !== undefined;
}

function go2jsReflectValueCanSet(receiver) {
	return go2jsReflectValueBox(receiver) && receiver.v !== null && receiver.v !== undefined &&
		typeof receiver.v === "object";
}

// go2jsReflectValueMapIndex reads an entry of a map. A key the map does not
// hold has no value to answer with, so it answers with the zero Value, which is
// what reflect gives a program that asked for one.
function go2jsReflectValueMapIndex(receiver, key) {
	const target = go2jsReflectRead(receiver);
	const element = receiver !== null && receiver !== undefined && receiver.t !== null && receiver.t !== undefined
		? go2jsReflectElem(receiver.t)
		: null;

	if (target instanceof go2jsNativeMap && go2jsReflectValueIsValid(key)) {
		const at = go2jsReflectUnwrap(go2jsReflectRead(key));
		const found = target.get(at);

		if (found !== undefined) {
			return go2jsReflectValue(found, element || go2jsReflectTypeOf(found));
		}
	}

	return go2jsReflectValue(null, element !== null ? element : go2jsReflectType("invalid", "", null, null, []));
}

// go2jsReflectValueMapKeys lists the keys a map holds, which is every key it
// has an entry for, in the order it holds them.
function go2jsReflectValueMapKeys(receiver) {
	const target = go2jsReflectRead(receiver);
	const element = receiver !== null && receiver !== undefined && receiver.t !== null && receiver.t !== undefined &&
		receiver.t.key !== null && receiver.t.key !== undefined
		? receiver.t.key
		: null;
	const keys = [];

	if (target instanceof go2jsNativeMap) {
		for (const key of target.keys()) {
			keys.push(go2jsReflectValue(key, element || go2jsReflectTypeOf(key)));
		}
	}

	return keys;
}

// go2jsReflectValueSetMapIndex writes an entry of a map, and a value that is
// not a value at all is the way a program asks for the entry to be removed.
function go2jsReflectValueSetMapIndex(receiver, key, value) {
	// The map is read through the box rather than taken from it, because a box
	// that was handed a fresh map by Set kept the one it had, and reading it is
	// what tells a write the map it is writing into.
	const target = go2jsReflectRead(receiver);

	if (!go2jsReflectValueIsValid(key)) {
		go2jsPanic("reflect: SetMapIndex on zero key");
	}

	if (!(target instanceof go2jsNativeMap)) {
		go2jsPanic("reflect: SetMapIndex on non-map value");
	}

	const at = go2jsReflectUnwrap(go2jsReflectRead(key));

	if (go2jsReflectValueBox(value) && !go2jsReflectValueIsValid(value)) {
		go2jsMapDelete(target, at);
		return;
	}

	go2jsMapSet(target, at, go2jsReflectUnwrap(go2jsReflectRead(value)));
}

// go2jsReflectValueSlice takes a slice of a slice or an array, sharing the
// storage behind it, so a write to what it hands back is a write to the slice
// it came from.
function go2jsReflectValueSlice(receiver, low, high, max) {
	const target = go2jsReflectRead(receiver);

	if (!Array.isArray(target)) {
		go2jsPanic("reflect.Value.Slice: slice of unaddressable array or string");
	}

	const from = low === undefined ? 0 : Math.trunc(Number(low));
	const to = high === undefined ? target.length : Math.trunc(Number(high));

	if (from < 0 || to > target.length || to < from) {
		go2jsPanic("reflect.Value.Slice: slice bounds out of range");
	}

	const from_t = receiver !== null && receiver !== undefined ? receiver.t : null;
	const element = from_t !== null ? go2jsReflectElem(from_t) : null;

	// The value handed back is a slice, so it is described as one, keeping the
	// name and the type of what it was cut from where the container said them.
	// Describing it as its element type instead makes a program that asks what
	// kind of thing it now has answer with the kind of the things inside it.
	const sliced = from_t !== null && from_t.kind === "slice"
		? from_t
		: go2jsReflectType("slice", "", element, null, []);

	// A slice cut from a slice shares the storage it was cut out of, so a write
	// through it has to land in that storage. Cutting a run of elements out of an
	// array and handing the run back on its own gives a copy, and a write through
	// the copy is a write the array never hears of, which is the one thing a
	// slice in Go is not.
	return {
		__go2js_reflectValue: true,
		__go2js_sliceView: true,
		get: function() {
			return target.slice(from, to);
		},
		set: function(next) {
			const elements = Array.isArray(next) ? next : [];

			// Only the elements the slice covers are written, because a slice
			// holds its own length rather than the length of what it was cut
			// from, and the storage past that length is left as it was.
			for (let i = 0; i < elements.length && from + i < to; i++) {
				target[from + i] = elements[i];
			}
		},
		v: target.slice(from, to),
		t: sliced
	};
}

// go2jsReflectValueAppend adds values to the slice a Value stands for, which is
// how a program grows one without knowing how much room it has.
function go2jsReflectValueAppend(receiver, ...values) {
	const target = go2jsReflectRead(receiver);
	const items = Array.isArray(target) ? target.slice() : [];

	for (const value of values) {
		items.push(go2jsReflectUnwrap(go2jsReflectRead(value)));
	}

	const element = receiver !== null && receiver !== undefined && receiver.t !== null && receiver.t !== undefined
		? go2jsReflectElem(receiver.t)
		: null;

	return go2jsReflectValue(items, element !== null ? element : go2jsReflectTypeOf(items[items.length - 1]));
}

// go2jsReflectValueCopy copies values from one slice into another, which is what
// copy does, and answers with how many were copied.
function go2jsReflectValueCopy(receiver, source) {
	const target = go2jsReflectRead(receiver);
	const from = go2jsReflectUnwrap(go2jsReflectRead(source));
	const length = Math.min(Array.isArray(target) ? target.length : 0, Array.isArray(from) ? from.length : 0);

	for (let i = 0; i < length; i++) {
		target[i] = from[i];
	}

	return length;
}

// go2jsReflectValueNew allocates a zero value of the type the receiver names
// and hands back a pointer to it, which is what a program gets for New.
function go2jsReflectValueNew(receiver) {
	if (!go2jsReflectValueBox(receiver) || receiver.t === null || receiver.t === undefined) {
		go2jsPanic("reflect: New of zero Value");
	}

	return go2jsReflectNew(receiver.t);
}

// go2jsReflectIntBits reports how wide an integer kind is, and reports zero for
// a kind that is not an integer at all, which is the one width that is never
// too narrow. A width is all it takes to say whether a number fits, because
// every integer kind differs from its neighbours only in how many bits it has
// and in whether those bits read as signed.
function go2jsReflectIntBits(kind) {
	switch (kind) {
		case "int8":
		case "uint8":
			return 8;
		case "int16":
		case "uint16":
			return 16;
		case "int32":
		case "uint32":
			return 32;
		case "int64":
		case "uint64":
		case "uint":
		case "uintptr":
			return 64;
		default:
			return 0;
	}
}

// go2jsReflectIntIsSigned reports whether a width of that many bits reads as
// signed, which is the half of a kind's name after the u.
function go2jsReflectIntIsSigned(kind) {
	return kind.charAt(0) !== "u";
}

// go2jsReflectValueOverflowInt reports whether a whole number is one the type
// cannot hold, which is what a program that asks before it writes is told. A
// number of a shape the type has no answer for fits as well as anything.
function go2jsReflectValueOverflowInt(receiver, value) {
	const kind = go2jsReflectKind(receiver);
	const bits = go2jsReflectIntBits(kind);

	if (bits === 0) {
		return false;
	}

	const number = Math.trunc(Number(value));
	const top = 2 ** (bits - 1);

	return go2jsReflectIntIsSigned(kind)
		? number < -top || number > top - 1
		: number < 0 || number > 2 * top - 1;
}

// go2jsReflectValueOverflowUint reports the same for a number read as unsigned,
// which is the question a program asks when it has an unsigned number in hand
// and a signed type to put it in. That type keeps the numbers up to its own
// largest, and not the ones past it.
function go2jsReflectValueOverflowUint(receiver, value) {
	const kind = go2jsReflectKind(receiver);
	const bits = go2jsReflectIntBits(kind);

	if (bits === 0) {
		return false;
	}

	const number = Math.trunc(Number(value));
	const top = 2 ** (bits - 1);

	return go2jsReflectIntIsSigned(kind)
		? number < 0 || number > top - 1
		: number < 0 || number > 2 * top - 1;
}

// go2jsReflectNarrowInt keeps the low bits of a whole number the way a type
// narrower than the number does. A program that writes a number too wide means
// the number that fits, so the value stored is not the one that was asked for.
function go2jsReflectNarrowInt(kind, value) {
	const bits = go2jsReflectIntBits(kind);
	const number = Math.trunc(Number(value));

	if (bits === 0) {
		return number;
	}

	const top = 2 ** bits;
	const kept = ((number % top) + top) % top;

	return go2jsReflectIntIsSigned(kind) && kept > top / 2 - 1 ? kept - top : kept;
}

// go2jsReflectValueSetInt stores a whole number, narrowed to the width of the
// type it goes into.
function go2jsReflectValueSetInt(receiver, value) {
	go2jsReflectValueSet(receiver, go2jsReflectNarrowInt(go2jsReflectKind(receiver), value));
}

// go2jsReflectValueSetUint stores an unsigned number, narrowed to the width of
// the type it goes into.
function go2jsReflectValueSetUint(receiver, value) {
	go2jsReflectValueSet(receiver, go2jsReflectNarrowInt(go2jsReflectKind(receiver), value));
}

function go2jsReflectValueSetLen(receiver, length) {
	const target = go2jsReflectUnwrap(receiver);

	if (Array.isArray(target)) {
		target.length = Math.trunc(Number(length));
	}
}

function go2jsReflectTypeName(receiver) {
	return receiver !== null && receiver !== undefined && receiver.name ? receiver.name : "";
}

function go2jsReflectTypeKindOf(receiver) {
	// A type descriptor already knows its kind, so it does not have to be derived
	// from a value the way go2jsReflectTypeKind has to.
	if (receiver !== null && receiver !== undefined && receiver.__go2js_reflectType === true) {
		return receiver.kind;
	}

	return go2jsReflectTypeKind(receiver);
}

function go2jsReflectTypeNumField(receiver) {
	if (Array.isArray(receiver.fields)) {
		return receiver.fields.length;
	}

	return go2jsReflectStructFieldNames(go2jsReflectZeroOf(receiver).v).length;
}

function go2jsReflectTypeField(receiver, index) {
	// A descriptor that names its own fields answers from that list, because a
	// zero value of the type would be built without them.
	if (Array.isArray(receiver.fields)) {
		if (index < 0 || index >= receiver.fields.length) {
			throw new RangeError("reflect: Field index out of range");
		}

		return go2jsReflectStructField(receiver.fields[index]);
	}

	const field = go2jsReflectFieldAt(go2jsReflectUnwrap(go2jsReflectZeroOf(receiver)), index);

	return go2jsReflectStructField({Name: field.name, Tag: "", PkgPath: "", Anonymous: false});
}

function go2jsReflectTypeElemOf(receiver) {
	return go2jsReflectElem(go2jsReflectValueType(receiver));
}

function go2jsReflectTypeSize() {
	return 0;
}

// go2jsReflectTypeKey reports the key type of a map, which is the only kind
// that has one.
function go2jsReflectTypeKey(receiver) {
	const type = go2jsReflectValueType(receiver);

	if (type.key === null || type.key === undefined) {
		throw new TypeError("reflect: Key of non-map type " + go2jsReflectTypeString(type));
	}

	return type.key;
}

// go2jsReflectTypeNumMethod reports how many methods a type carries.
function go2jsReflectTypeNumMethod(receiver) {
	return go2jsReflectMethodsOf(go2jsReflectValueType(receiver)).length;
}

// go2jsReflectTypeMethod reports one method by the order Go sorted the method
// set in, which is alphabetical by name.
function go2jsReflectTypeMethod(receiver, index) {
	const methods = go2jsReflectMethodsOf(go2jsReflectValueType(receiver));

	if (index < 0 || index >= methods.length) {
		throw new RangeError("reflect: Method index out of range");
	}

	return {Name: methods[index], Type: null, Index: index};
}

// go2jsReflectValueMethodByName looks a method up by name, which reports the
// same shape as Method does so a caller cannot tell the two apart.
function go2jsReflectValueMethodByName(receiver, name) {
	const methods = go2jsReflectMethodsOf(go2jsReflectValueType(receiver));
	const index = methods.indexOf(name);

	if (index < 0) {
		return {__go2js_reflectValue: true, v: {}, t: go2jsReflectType("invalid", "", null, null, [])};
	}

	return {Name: name, Type: null, Index: index};
}

// go2jsReflectValueCall calls a method or function value. A method reached by
// name is looked up in the table the emitter filled in, and a function value is
// called through the cell that holds it, so both go the way they were declared.
function go2jsReflectValueCall(receiver, ...args) {
	const value = go2jsReflectRead(receiver);

	if (value !== null && value !== undefined && value.__go2js_callable === true) {
		return value.__go2js_callable_value(...args);
	}

	const name = receiver && receiver.__go2js_methodName;
	const fn = typeof name === "string" ? go2jsMethodTable[name] : undefined;

	if (typeof fn !== "function") {
		throw new TypeError("reflect: call of reflect.Value.Call on a value that is not callable");
	}

	return go2jsCallNow(fn, null, [value, ...args]);
}

// go2jsReflectValueAddr is the address of a value, which is a pointer to the
// storage the value was read from when there is any, and a new cell holding the
// value otherwise.
// go2jsReflectValueAddr is the address of a value, which is a pointer to the
// storage it lives in. The pointer it hands back still reaches that storage, so
// what it points at can be written to, and a method a program finds on the
// pointer is a method of what it points at, which is how a pointer answers for
// the interface its element type implements.
function go2jsReflectValueAddr(receiver) {
	if (!go2jsReflectValueBox(receiver)) {
		go2jsPanic("reflect: Addr of unaddressable value");
	}

	const type = go2jsReflectPtrTo(go2jsReflectValueType(receiver));
	const name = type !== null && type !== undefined && type.elem !== null && type.elem !== undefined && type.elem.name
		? "*" + type.elem.name
		: undefined;

	if (typeof receiver.get === "function" && typeof receiver.set === "function") {
		// The address keeps the storage the value lives in, so a write through
		// the pointer is a write to the value the address was taken of.
		let storage = receiver.get();

		return go2jsReflectValue(go2jsPtr(() => storage, next => storage = next, name), type);
	}

	return go2jsReflectValue(go2jsNew(go2jsReflectRead(receiver), name), type);
}

// go2jsReflectValuePointer reports an address as a number, which for a value
// built by a translation is a number it keeps for the life of the value, the
// same way it prints a pointer.
function go2jsReflectValuePointer(receiver) {
	return go2jsPointerAddress(go2jsReflectRead(receiver));
}

// go2jsReflectStructField describes one field of a struct: its name, its type
// and its tag, which is what reflect.Type.Field reports.
// go2jsReflectStructField reports one field of a struct the way reflect does.
// A field is more than a name: a program reads a tag off it, a private field is
// one with a package path set, an embedded field is one that is anonymous, and
// the type of a field is what a program keys its decoding on. A bare name is
// taken to be a field that says only the name.
function go2jsReflectStructField(field) {
	const described = typeof field === "string" ? {Name: field} : field;

	return {
		__go2js_reflectStructField: true,
		Name: described.Name,
		Tag: described.Tag || "",
		PkgPath: described.PkgPath || "",
		Anonymous: described.Anonymous === true,
		Index: described.Index || [],
		Offset: described.Offset || 0,
		Type: go2jsReflectFieldTypeOf(described),
		IsExported: () => described.PkgPath === ""
	};
}

// The methods of reflect.Value and reflect.Type are reached through the method
// table, keyed by the registered type name, because those two names resolve to
// the runtime helper functions rather than to generated classes.
const go2jsReflectValueMethods = {
	Elem: go2jsReflectValueElem,
	Interface: go2jsReflectValueInterface,
	String: go2jsReflectValueString,
	Int: go2jsReflectValueInt,
	Uint: go2jsReflectValueUint,
	Float: go2jsReflectValueFloat,
	Bool: go2jsReflectValueBool,
	Len: go2jsReflectValueLen,
	Kind: go2jsReflectKind,
	Type: go2jsReflectValueType,
	Set: go2jsReflectValueSet,
	SetString: go2jsReflectValueSet,
	SetInt: go2jsReflectValueSetInt,
	SetUint: go2jsReflectValueSetUint,
	SetFloat: go2jsReflectValueSet,
	SetBool: go2jsReflectValueSet,
	NumField: go2jsReflectValueNumField,
	Field: go2jsReflectValueField,
	Index: go2jsReflectValueIndex,
	NumMethod: go2jsReflectValueNumMethod,
	IsValid: go2jsReflectValueIsValid,
	IsNil: go2jsReflectValueIsNil,
	IsZero: go2jsReflectValueIsZero,
	CanSet: go2jsReflectValueCanSet,
	CanInterface: go2jsReflectValueCanInterface,
	CanAddr: go2jsReflectValueCanAddr,
	CanCompare: go2jsReflectValueCanCompare,
	IsExported: go2jsReflectValueIsExported,
	Addr: go2jsReflectValueAddr,
	UnsafeAddr: go2jsReflectValueAddr,
	Pointer: go2jsReflectValuePointer,
	MethodByName: go2jsReflectValueMethodByName,
	Call: go2jsReflectValueCall,
	SetLen: go2jsReflectValueSetLen,
	MapIndex: go2jsReflectValueMapIndex,
	MapKeys: go2jsReflectValueMapKeys,
	MapRange: go2jsReflectValueMapKeys,
	SetMapIndex: go2jsReflectValueSetMapIndex,
	Slice: go2jsReflectValueSlice,
	Append: go2jsReflectValueAppend,
	Copy: go2jsReflectValueCopy,
	New: go2jsReflectValueNew,
	OverflowInt: go2jsReflectValueOverflowInt,
	OverflowUint: go2jsReflectValueOverflowUint
};

// go2jsReflectStructTagGet reports the value of one key in a struct tag, which
// is how a program reads a tag that names a key and gives it a value. A key the
// tag does not carry is reported as not being there, and a key standing on its
// own without a value reads as the empty string, which is what the tag says.
function go2jsReflectStructTagGet(receiver, key) {
	// A tag is reached as a string most of the time and as a field descriptor
	// the rest, because a field carries one while a bare tag does not.
	const tag = typeof receiver === "string"
		? receiver
		: (receiver === null || receiver === undefined ? "" : String(receiver.Tag || ""));

	if (tag === "") {
		return "";
	}

	for (const part of tag.split(/\s+/)) {
		if (part === "") {
			continue;
		}

		// A tag is written as a key, a colon, and the value in quotes, and a key
		// standing on its own says the value is the empty string.
		const colon = part.indexOf(":");
		const name = colon < 0 ? part : part.slice(0, colon);

		if (name !== key) {
			continue;
		}

		if (colon < 0) {
			return "";
		}

		const value = part.slice(colon + 1);

		return value.charAt(0) === "\"" && value.charAt(value.length - 1) === "\""
			? value.slice(1, value.length - 1)
			: value;
	}

	return "";
}

const go2jsReflectTypeMethods = {
	Name: go2jsReflectTypeName,
	Kind: go2jsReflectTypeKindOf,
	NumField: go2jsReflectTypeNumField,
	Field: go2jsReflectTypeField,
	Elem: go2jsReflectTypeElemOf,
	String: go2jsReflectTypeString,
	Size: go2jsReflectTypeSize,
	Implements: go2jsReflectTypeImplements,
	Key: go2jsReflectTypeKey,
	Len: go2jsReflectTypeLen,
	NumMethod: go2jsReflectTypeNumMethod,
	Method: go2jsReflectTypeMethod
};

function go2jsReflectRegisterMethods() {
	if (go2jsReflectRegistered) {
		return;
	}

	go2jsReflectRegistered = true;

	for (const [name, fn] of Object.entries(go2jsReflectValueMethods)) {
		go2jsMethodTable["reflect.Value." + name] = fn;
	}

	// The tag of a field is itself reached as a value, because reflect gives it
	// a type of its own and a program calls Get on it.
	go2jsMethodTable["reflect.StructTag.Get"] = go2jsReflectStructTagGet;

	for (const [name, fn] of Object.entries(go2jsReflectTypeMethods)) {
		go2jsMethodTable["reflect.Type." + name] = fn;
	}
}

let go2jsReflectRegistered = false;

// go2jsReflectNew allocates a zero value of the given type and returns an
// addressable Value for it, which is what a Go pointer to a fresh T provides.
function go2jsReflectNew(type) {
	return go2jsReflectValue(go2jsReflectNewBox(go2jsReflectUnwrap(go2jsReflectZeroOf(type))), type);
}

// go2jsReflectNewBox holds the allocated storage so a later Set on the
// pointer target is visible through Elem, the way writing through a Go pointer
// is.
function go2jsReflectNewBox(zero) {
	return {__go2js_reflectNew: true, target: zero};
}

function go2jsReflectIsNewBox(value) {
	return value !== null && value !== undefined && value.__go2js_reflectNew === true;
}

function go2jsReflectElem(type) {
	return type !== null && type !== undefined && type.elem !== null && type.elem !== undefined ?
		type.elem :
		go2jsReflectType("invalid", "", null, null);
}

function go2jsReflectValueElem(receiver) {
	if (!go2jsReflectValueBox(receiver)) {
		return go2jsReflectValue(receiver, go2jsReflectElem(go2jsReflectTypeOf(receiver)));
	}

	// A value produced by New points at fresh storage, so Elem yields that
	// storage and writes through it are kept.
	if (go2jsReflectIsNewBox(receiver.v)) {
		return go2jsReflectAddressable(receiver.v, receiver.t);
	}

	if (receiver.t === null || receiver.t === undefined) {
		return go2jsReflectValue(receiver.v, go2jsReflectElem(go2jsReflectTypeOf(receiver.v)));
	}

	if (receiver.t.kind === "interface") {
		const inner = receiver.v !== null && receiver.v !== undefined ? receiver.v.v : receiver.v;
		return go2jsReflectValue(inner, go2jsReflectTypeOf(inner));
	}

	// A pointer that was described without saying what it points at takes the
	// type of what it points at, which is the only thing left that says. A
	// pointer that reached reflect by way of an interface is described from the
	// cell standing for it, and a cell knows how to be read and not what is
	// behind it.
	const elemType = receiver.t.kind === "ptr" && !receiver.t.elem && receiver.v !== null && receiver.v !== undefined &&
		go2jsPointerAccessor(receiver.v, go2jsPointerGet)
		? go2jsReflectTypeOf(go2jsStripWrappers(receiver.v[go2jsPointerGet]()))
		: go2jsReflectElem(receiver.t);

	// Elem of a pointer is addressable, so a write to what it points at reaches
	// the pointee the same way a Go pointer to a field does.
	if (receiver.t.kind === "ptr" && receiver.v !== null && receiver.v !== undefined &&
		go2jsPointerAccessor(receiver.v, go2jsPointerGet) && go2jsPointerAccessor(receiver.v, go2jsPointerSet)) {
		const pointer = receiver.v;

		return {
			__go2js_reflectValue: true,
			get: function() {
				return pointer[go2jsPointerGet]();
			},
			set: function(next) {
				pointer[go2jsPointerSet](next);
			},
			v: pointer[go2jsPointerGet](),
			t: elemType
		};
	}

	return go2jsReflectValue(receiver.v, elemType);
}

function go2jsReflectValueSet(receiver, value) {
	if (!go2jsReflectValueBox(receiver)) {
		throw new TypeError("reflect: Set using unaddressable value");
	}

	const next = go2jsReflectUnwrap(value);

	if (typeof receiver.set === "function" && typeof receiver.get === "function") {
		receiver.set(next);
		return;
	}

	// A slot reached through a wrapper keeps its setter one level down.
	if (receiver.v !== null && receiver.v !== undefined &&
		go2jsPointerAccessor(receiver.v, go2jsPointerGet) && go2jsPointerAccessor(receiver.v, go2jsPointerSet)) {
		receiver.v[go2jsPointerSet](next);
		return;
	}

	if (receiver.v !== null && receiver.v !== undefined && typeof receiver.v === "object") {
		receiver.v.__go2js_reflectValue = true;
		receiver.v.v = next;
		receiver.v.t = receiver.t;
		return;
	}

	throw new TypeError("reflect: Set using unaddressable value");
}

function go2jsReflectValueInterface(receiver) {
	return go2jsReflectUnwrap(go2jsReflectRead(receiver));
}

function go2jsReflectValueString(receiver) {
	if (!go2jsReflectValueBox(receiver)) {
		return String(receiver);
	}

	const value = go2jsReflectUnwrap(go2jsReflectRead(receiver));

	return typeof value === "string" ? value : String(value);
}

function go2jsReflectValueInt(receiver) {
	const value = go2jsReflectUnwrap(go2jsReflectRead(receiver));

	if (typeof value === "bigint") {
		return value;
	}

	return Math.trunc(Number(value) || 0);
}

function go2jsReflectValueUint(receiver) {
	return Number(go2jsReflectValueInt(receiver));
}

function go2jsReflectValueFloat(receiver) {
	return Number(go2jsReflectUnwrap(go2jsReflectRead(receiver)));
}

function go2jsReflectValueBool(receiver) {
	return Boolean(go2jsReflectUnwrap(go2jsReflectRead(receiver)));
}

function go2jsReflectValueLen(receiver) {
	const value = go2jsReflectUnwrap(go2jsReflectRead(receiver));

	if (value === null || value === undefined) {
		return 0;
	}

	if (value instanceof go2jsNativeMap) {
		return value.size;
	}

	return value.length !== undefined ? value.length : 0;
}

function go2jsReflectValueIsNil(receiver) {
	const value = go2jsReflectUnwrap(go2jsReflectRead(receiver));

	return value === null || value === undefined;
}

// go2jsReflectValueIsZero reports whether a value is the zero value of its type.
// Every kind has its own zero, and a value that was never set holds it already,
// so the kinds that hold what they were given are compared against the zero of
// that kind rather than against undefined.
function go2jsReflectValueIsZero(receiver) {
	if (!go2jsReflectValueIsValid(receiver)) {
		panic("reflect: IsZero of an invalid Value");
	}

	const kind = go2jsReflectKind(receiver);
	const value = go2jsReflectUnwrap(go2jsReflectRead(receiver));

	if (value === null || value === undefined) {
		return true;
	}

	switch (kind) {
		case "bool":
			return value === false;
		case "int":
		case "int8":
		case "int16":
		case "int32":
		case "int64":
		case "uint":
		case "uint8":
		case "uint16":
		case "uint32":
		case "uint64":
		case "uintptr":
		case "float32":
		case "float64":
		case "complex64":
		case "complex128":
			return Number(value) === 0;
		case "string":
			return String(value) === "";
		case "map":
		case "slice":
			return go2jsReflectValueLen(receiver) === 0;
		case "ptr":
		case "unsafePointer":
		case "chan":
		case "func":
		case "interface":
			return value === null || value === undefined || value === false;
	}

	// A struct is the zero value while none of its fields differ from the zero
	// value of their own type, which is what a program that builds one field at
	// a time comes out as.
	if (kind === "struct") {
		for (const field of go2jsStructFieldsOf(value)) {
			if (!go2jsReflectValueIsZero(go2jsReflectValue(value[field], null))) {
				return false;
			}
		}

		return true;
	}

	if (kind === "array") {
		for (let i = 0; i < go2jsReflectValueLen(receiver); i++) {
			if (!go2jsReflectValueIsZero(receiver.Index(i))) {
				return false;
			}
		}

		return true;
	}

	return false;
}

// go2jsStructFieldsOf lists the fields of a value, whether it is a plain object
// or a value the emitter described field by field.
function go2jsStructFieldsOf(value) {
	if (value === null || typeof value !== "object") {
		return [];
	}

	if (Array.isArray(value.__go2js_fields)) {
		return value.__go2js_fields.map(name => value[name]);
	}

	return Object.keys(value).filter(name => !name.startsWith("__go2js_")).map(name => value[name]);
}

// A translation hands out every value it can reach, so nothing is withheld the
// way an unexported field is in Go. A value that came out of a pointer is
// addressable, the same way one that came out of a slice is.
function go2jsReflectValueCanInterface(receiver) {
	return go2jsReflectValueIsValid(receiver);
}

function go2jsReflectValueCanAddr(receiver) {
	return go2jsReflectValueIsValid(receiver) && receiver.__go2js_reflectValue === true;
}

// A slice, a map and a function cannot be compared in Go, and the rest can, so
// the answer follows the kind rather than the value.
function go2jsReflectValueCanCompare(receiver) {
	if (!go2jsReflectValueIsValid(receiver)) {
		return false;
	}

	const kind = go2jsReflectKind(receiver);

	return kind !== "slice" && kind !== "map" && kind !== "func";
}

// A field of a struct built by the emitter is written under its own name, and
// an unexported one of the standard library is written under the name Go gives
// it, so the first letter of the name settles whether it was exported.
function go2jsReflectValueIsExported(receiver) {
	const name = receiver && receiver.__go2js_fieldName;

	if (typeof name !== "string" || name === "") {
		return true;
	}

	const first = name[0];

	return first === first.toUpperCase() && first !== first.toLowerCase();
}

function go2jsReflectTypeString(receiver) {
	if (receiver === null || receiver === undefined) {
		return "invalid";
	}

	if (receiver.__go2js_reflectType === true) {
		if (receiver.name !== "") {
			// A named type is named after the package it was declared in, which
			// the registry knows and the descriptor does not carry.
			const ctor = go2jsReflectClassOfName(receiver.name);

			if (ctor !== null && typeof ctor.name === "string") {
				const registered = go2jsRegisteredTypeName(ctor.name);

				if (typeof registered === "string" && registered !== "") {
					return registered;
				}
			}

			return receiver.name;
		}

		if (receiver.kind === "ptr" && receiver.elem) {
			return "*" + go2jsReflectTypeString(receiver.elem);
		}

		if (receiver.kind === "slice" && receiver.elem) {
			return "[]" + go2jsReflectTypeString(receiver.elem);
		}

		if (receiver.kind === "map" && receiver.key) {
			return "map[" + go2jsReflectTypeString(receiver.key) + "]" +
				go2jsReflectTypeString(receiver.elem);
		}

		if (receiver.kind === "array" && typeof receiver.len === "number" && receiver.elem) {
			return "[" + receiver.len + "]" + go2jsReflectTypeString(receiver.elem);
		}

		if (receiver.kind === "chan" && receiver.elem) {
			return "chan " + go2jsReflectTypeString(receiver.elem);
		}

		// The kind names are the ones a program reads them by, and an interface
		// and a channel are the two that are written out in full where a kind is
		// only a name for it.
		if (receiver.kind === "interface") {
			return "interface {}";
		}

		return receiver.kind;
	}

	return go2jsReflectTypeOf(receiver).kind;
}

function go2jsReflectZeroOf(receiver) {
	return go2jsReflectZeroOfType(go2jsReflectValueType(receiver));
}

function go2jsReflectZeroOfType(type) {
	if (type === null || type === undefined) {
		return go2jsReflectValue(null, null);
	}

	switch (type.kind) {
	case "bool":
		return go2jsReflectValue(false, type)
	case "string":
		return go2jsReflectValue("", type)
	case "int":
	case "int8":
	case "int16":
	case "int32":
	case "int64":
	case "uint":
	case "uint8":
	case "uint16":
	case "uint32":
	case "uint64":
	case "uintptr":
	case "float32":
	case "float64":
		return go2jsReflectValue(0, type)
	case "slice":
		return go2jsReflectValue([], type)
	case "map":
		return go2jsReflectValue(new go2jsNativeMap(), type)
	case "func":
	case "chan":
	case "interface":
	case "struct":
		// A struct has to be a real instance of its class, otherwise reflect
		// cannot name its fields.
		return go2jsReflectValue(go2jsReflectInstanceOf(type), type)
	case "ptr":
		return go2jsReflectValue(null, type)
	default:
		return go2jsReflectValue(null, type)
	}
}

// go2jsReflectInstanceOf builds the zero value of a named type from the class
// the emitter generated for it.
function go2jsReflectInstanceOf(type) {
	if (type === null || type === undefined || !type.name) {
		return null;
	}

	// The descriptor holds the bare type name while the registry is keyed by the
	// qualified one, so the lookup has to try both spellings, and then either
	// name of the same type, because which one it was handed is not something
	// the runtime gets to choose.
	const ctor = go2jsReflectClassOfName(type.name) || go2jsTypeNames[go2jsRegisteredTypeName(type.name)];

	if (typeof ctor !== "function") {
		return null;
	}

	try {
		return new ctor();
	} catch (cause) {
		return null;
	}
}

go2jsReflectRegisterMethods();

// go2jsSourceLines maps a line number in the generated program to the place in
// the Go source it came from, so that a stack trace can name the Go places a
// panic came from.
let go2jsSourceLines = null;

// go2jsPackageName is the package the program was declared in. A Go frame is
// written under the package the function it names was declared in, and the name
// the emitter wrote is the name on its own.
let go2jsPackageName = "";

function go2jsSetPackageName(name) {
	go2jsPackageName = String(name);
}

function go2jsSetSourceLines(lines) {
	// the lines named in the table are counted from the first line of the
	// program, which is the line after the one the table was written at. A
	// program of several files carries a table for each of them, and each names
	// the lines of the program as a whole, so the tables are read into one.
	const at = go2jsWrittenLine();

	if (go2jsSourceLines === null) {
		go2jsSourceLines = {};
	}

	const from = at === null ? 0 : at;

	for (const line of Object.keys(lines)) {
		go2jsSourceLines[from + Number(line)] = lines[line];
	}
}

// go2jsWrittenLine is the line of the generated program that asked for this.
// Every frame above the caller is one of the runtime's own, however many of them
// there happen to be, so the first frame that is not one is the program's.
function go2jsWrittenLine() {
	for (const line of go2jsStackLines()) {
		const match = /^\s*at (?:(.*?) )?\(?[^()]*:(\d+):\d+\)?$/.exec(line);

		if (match === null) {
			continue;
		}

		const name = match[1] === undefined ? "" : match[1];

		if (name === "" || name.startsWith("go2js")) {
			continue;
		}

		return Number(match[2]);
	}

	return null;
}

function go2jsLookupSourceLine(line) {
	if (go2jsSourceLines === null) {
		return "";
	}

	const pos = go2jsSourceLines[line];

	return pos === undefined ? "" : pos;
}

// go2jsSourcePlace is where a line of the generated program came from in the Go
// source. The emitter writes the place a line came from as file:line:column,
// which is read back here as the file and the line, the way Go writes a frame.
function go2jsSourcePlace(line) {
	const written = go2jsLookupSourceLine(line);

	if (written === "") {
		return null;
	}

	const match = /^(.*):(\d+):\d+$/.exec(written);

	if (match === null) {
		return null;
	}

	return {file: match[1], line: Number(match[2])};
}

// go2jsGoFrameName is the name a frame of a Go program is written under. The
// name the emitter wrote is the name the program declared, unless the frame is
// one of the module loader's or of the runtime's own, and a frame of the runtime
// is not a frame of the program.
function go2jsGoFrameName(name) {
	if (typeof name !== "string" || name === "") {
		return "";
	}

	if (name.startsWith("go2js") || name.startsWith("Object.") || name.startsWith("Module.") ||
		name.startsWith("Function.") || name.startsWith("node:") || name.startsWith("internal/")) {
		return "";
	}

	if (go2jsPackageName !== "" && name.indexOf(".") < 0) {
		return go2jsPackageName + "." + name;
	}

	return name;
}

// go2jsStackLines is the stack as the program stands on it, one frame to a line.
function go2jsStackLines() {
	const error = new Error();

	return String(error.stack === undefined ? "" : error.stack).split("\n");
}

// go2jsGoFrameOf is one frame of a JavaScript stack read back as a place in the
// Go source, and is nothing when the frame is not one the program has a name
// for or was not written out of the Go source at all.
function go2jsGoFrameOf(line) {
	const match = /^\s*at (?:(.*?) )?\(?([^()]*):(\d+):\d+\)?$/.exec(String(line));

	if (match === null) {
		return null;
	}

	const name = go2jsGoFrameName(match[1] === undefined ? "" : match[1]);

	if (name === "") {
		return null;
	}

	const place = go2jsSourcePlace(Number(match[3]));

	if (place === null) {
		return null;
	}

	return {name: name, file: place.file, line: place.line};
}

// go2jsGoFrames is the trace of the places the program is standing on, read off
// a JavaScript stack and read back into the Go source. A frame is written as
// "    at name (file:line:column)", and the line in it is a line of the
// generated program, which is what the table of lines is for. The frames of the
// goroutine runner and of the runtime itself are left out, since a Go program
// does not show them either.
function go2jsGoFrames(stack) {
	const frames = [];

	for (const line of stack === undefined ? go2jsStackLines() : stack) {
		const frame = go2jsGoFrameOf(line);

		if (frame !== null) {
			frames.push(frame);
		}
	}

	return frames;
}

// go2jsFrameText is one frame written the way Go writes it: the name of the
// function and where in the source it was, on two lines.
function go2jsFrameText(frame) {
	return frame.name + "()\n\t" + frame.file + ":" + frame.line + "\n";
}

// go2jsPanicTrace is what a Go program prints when a panic is not recovered:
// the fault, a blank line, the goroutine it was on, and the frames of the
// program that ran on it.
function go2jsPanicTrace(thrown, value) {
	return "panic: " + go2jsPanicMessage(value) + "\n\ngoroutine 1 [running]:\n" +
		go2jsFramesText(go2jsGoFrames(go2jsThrownLines(thrown)));
}

// go2jsFramesText is the frames of the program, one after another, which is the
// part a trace of the program shares with a trace of anything else.
function go2jsFramesText(frames) {
	let text = "";

	for (const frame of frames) {
		text += go2jsFrameText(frame);
	}

	return text;
}

// go2jsThrownLines is the stack a thrown value carries, which is the stack it was
// thrown from rather than the stack of whoever is looking at it now.
function go2jsThrownLines(thrown) {
	if (thrown === null || thrown === undefined || thrown.stack === undefined) {
		return go2jsStackLines();
	}

	return String(thrown.stack).split("\n");
}

// go2jsPanicMessage is the fault as Go writes it in the first line of a panic,
// which is the value the program panicked with and nothing else.
function go2jsPanicMessage(value) {
	const payload = go2jsPanicPayload(value);

	if (payload instanceof Error) {
		return payload.message;
	}

	return go2jsStringify(payload);
}

// go2jsReportUncaughtPanic writes the trace of a panic that nothing recovered,
// and ends the program the way a Go program ends when a fault is not dealt
// with: with a status of two.
function go2jsReportUncaughtPanic(thrown) {
	const value = thrown === null || thrown === undefined ? undefined : thrown.__go2js_panic_value;

	// A fault the runtime raised about itself rather than about the program is
	// a fatal error, which Go writes with its own words in place of a panic and
	// ends the program the same way.
	if (thrown !== null && thrown !== undefined && thrown.__go2js_fatal !== undefined) {
		process.stderr.write("fatal error: " + thrown.__go2js_fatal + "\n\ngoroutine 1 [running]:\n" +
			go2jsFramesText(go2jsGoFrames(go2jsThrownLines(thrown))));
		process.exit(2);
	}

	// A fault of the runtime itself is written the way Go writes it as well, so a
	// member reached through a standing-for-nothing value is the nil dereference
	// Go would have raised rather than the wording of the engine underneath.
	if (value === undefined && go2jsRuntimeErrorText(thrown) === null) {
		throw thrown;
	}

	process.stderr.write(go2jsPanicTrace(thrown, value === undefined ? thrown : value));
	process.exit(2);
}

// go2jsBigInt is the holder of a whole number of the math/big package. Go says an
// Int is a value of its own and never a name for one, so a program that copies a
// pointer to an Int has two pointers to the one value, which is what two holders
// standing for one number cannot say; a pointer is therefore held as the holder
// itself and a copy of a pointer is a copy of that holder.
function go2jsBigInt() {
	// The holder says it holds a whole number of its own, which is what tells the
	// writing of one to count the number rather than to name the holder, the way
	// a type of Go that answers for itself is written rather than described.
	this.__go2js_whole_number = true;
	this.value = 0n;
}

// A whole number of no width is the one number rather than a copy of it, since
// every method of an Int is written on a pointer to it and the holder is that
// pointer, so a copy of a pointer is a copy of the pointer and not a second Int.
go2jsBigInt.prototype.__go2js_shared = true;

// go2jsBigIntDigits answers the value of a digit written as a letter, which is
// how the digits above nine are written from ten up, and -1 for anything that is
// no digit at all.
function go2jsBigIntDigit(character) {
	if (character >= "0" && character <= "9") {
		return character.charCodeAt(0) - 48;
	}

	const lowered = character.toLowerCase();
	if (lowered >= "a" && lowered <= "z") {
		return lowered.charCodeAt(0) - 97 + 10;
	}

	return -1;
}

// go2jsBigIntValue is the number inside a value of the package, however it was
// handed over. A value that has been through an interface is held in a box, and
// the box is a wrapper around the number rather than the number itself, and a
// value of nothing is a number of no value at all, which is as good as zero.
function go2jsBigIntValue(value) {
	if (value !== null && value !== undefined && value.__go2js_interface === true) {
		value = value.value;
	}

	if (value === null || value === undefined) {
		return 0n;
	}

	if (value !== null && value !== undefined && typeof value === "object" &&
		typeof value.value === "bigint") {
		return value.value;
	}

	return BigInt(value);
}

// go2jsBigIntOf is the number a pointer to an Int holds. A method on a pointer to
// nothing is a fault in Go rather than an answer, which is what reaching for the
// number behind nothing is.
function go2jsBigIntOf(receiver) {
	if (receiver === null || receiver === undefined) {
		throw new TypeError(go2jsRuntimeErrorPrefix + "invalid memory address or nil pointer dereference");
	}

	return go2jsBigIntValue(receiver);
}

// go2jsBigIntDivisor is the number a division is by, and dividing by zero is a
// fault in Go rather than an answer.
function go2jsBigIntDivisor(divisor) {
	const value = go2jsBigIntOf(divisor);

	if (value === 0n) {
		throw new go2jsNativeRangeError(go2jsRuntimeErrorPrefix + "division by zero");
	}

	return value;
}

// go2jsBigIntEuclidParts is what a division leaves three numbers of: what was
// divided, the divisor it was divided by, and the remainder left over, which is
// never negative and never as large as the divisor whatever the sign of either.
// JavaScript leaves the remainder with the sign of what was divided, so it is
// taken in the width of the divisor and brought up out of the negative side.
function go2jsBigIntEuclidParts(x, y) {
	const dividend = go2jsBigIntOf(x);
	const divisor = go2jsBigIntDivisor(y);
	const width = divisor < 0n ? -divisor : divisor;
	const remainder = ((dividend % width) + width) % width;

	return [dividend, divisor, remainder];
}

// go2jsBigIntLowBits is what a machine number of a width answers for a number
// wider than that width: the bits it holds at the bottom of the number, brought
// round about the ends of the width where they run past the top of it.
function go2jsBigIntLowBits(value, half, span) {
	const held = ((value % span) + span) % span;

	return held > half ? held - span : held;
}

// go2jsBigIntAbs is a number with nothing under it, which is a number below
// nothing standing the other way up rather than the number of no value at all.
function go2jsBigIntAbs(value) {
	return value < 0n ? -value : value;
}

// go2jsBigIntMod is a number counted off another as many times as it takes, which
// is the count left over in the width of the other however either is written, and
// where the other is no value at all there is nothing to count off and the count
// is the number itself.
function go2jsBigIntMod(value, modulus) {
	if (modulus === 0n) {
		return value;
	}

	return go2jsBigIntEuclidParts(value, {value: modulus})[2];
}

// go2jsBigIntGiven is whether a pointer to an Int was handed over at all, which
// is what tells a pointer to nothing from a pointer to a number of no value: the
// first is nothing to write into and the second is written like any other.
function go2jsBigIntGiven(value) {
	return value !== null && value !== undefined;
}

// go2jsBigIntInto is a number written into the Int a pointer points at, and a
// pointer to nothing is a fault in Go rather than an answer, the same as reading
// through one is.
function go2jsBigIntInto(pointer, held) {
	if (!go2jsBigIntGiven(pointer)) {
		throw new TypeError(go2jsRuntimeErrorPrefix + "invalid memory address or nil pointer dereference");
	}

	pointer.value = held;

	return pointer;
}

// go2jsBigIntBezout is a number that divides two without a remainder together with
// the count of the first of them that says so, both taken from the sizes of the
// two, since a count of one number belongs to it however it is written: a*x + b*y
// comes to the divisor for some y, and the count of the first is what is left of
// the divisor when the first multiplied by it is taken away, divided by the second.
function go2jsBigIntBezout(a, b) {
	const left = go2jsBigIntAbs(a);
	const right = go2jsBigIntAbs(b);
	let held = left;
	let other = right;
	let count = 1n;
	let step = 0n;

	// The two are brought down to nothing by the division that leaves nothing
	// behind, keeping beside the count how many times the first was taken in to
	// reach what is held.
	while (other !== 0n) {
		const quotient = held / other;

		[held, other] = [other, held - quotient * other];
		[count, step] = [step, count - quotient * step];
	}

	return [held, count];
}

// go2jsBigIntCount is a count of places or bits, which Go counts in a machine
// word and hands over to the package as a number, and JavaScript has no count
// wider than a double, so the count is taken as the number it stands for.
function go2jsBigIntCount(value) {
	const held = typeof value === "bigint" ? value : BigInt(Math.trunc(Number(value)));

	// A count below nothing is a fault in Go, since there is no place below the
	// first one to count from and no bit below the bottom one to count up to.
	if (held < 0n) {
		throw new go2jsNativeRangeError("negative count");
	}

	return held;
}

// go2jsBigIntSign is -1, 0 or 1 as a number is below, equal to or above nothing.
function go2jsBigIntSign(value) {
	return value < 0n ? -1 : (value > 0n ? 1 : 0);
}

// go2jsBigIntWidth is how wide a number has to be written for two of them to hold
// between them every bit either has, and a width is a number of bits rather than
// a number, so the count is a plain one.
function go2jsBigIntWidth(x, y) {
	const left = go2jsBigIntValue(x) < 0n ? -go2jsBigIntValue(x) : go2jsBigIntValue(x);
	const right = go2jsBigIntValue(y) < 0n ? -go2jsBigIntValue(y) : go2jsBigIntValue(y);

	return Number(left.toString(2).length > right.toString(2).length
		? left.toString(2).length
		: right.toString(2).length) + 1;
}

// go2jsBigIntBits is two numbers put together bit by bit. A number below nothing
// has no bits of its own in JavaScript, but Go writes a number below nothing in
// twos complement as far as the width it is written to, and both numbers are
// therefore written at a width that holds every bit of either before they are put
// together, so that a bit set in a number below nothing is set in the result the
// way Go sets it.
function go2jsBigIntBits(x, y, width, combine) {
	const held = BigInt.asUintN(width, go2jsBigIntValue(x));

	return BigInt.asIntN(width, combine(held, BigInt.asUintN(width, go2jsBigIntValue(y))));
}

go2jsBigInt.prototype.Set = function(value) {
	this.value = go2jsBigIntValue(value);

	return this;
};

go2jsBigInt.prototype.SetInt64 = function(value) {
	this.value = go2jsBigIntValue(value);

	return this;
};

go2jsBigInt.prototype.SetUint64 = function(value) {
	const held = go2jsBigIntValue(value);

	// A count of no width has no negative side and no upper one beyond the width
	// it is written to, so a number outside those is not a count at all.
	if (held < 0n || held > 18446744073709551615n) {
		throw new go2jsNativeRangeError("argument to SetUint64 is not a uint64");
	}

	this.value = held;

	return this;
};

go2jsBigInt.prototype.Add = function(x, y) {
	this.value = go2jsBigIntOf(x) + go2jsBigIntOf(y);

	return this;
};

go2jsBigInt.prototype.Sub = function(x, y) {
	this.value = go2jsBigIntOf(x) - go2jsBigIntOf(y);

	return this;
};

go2jsBigInt.prototype.Mul = function(x, y) {
	this.value = go2jsBigIntOf(x) * go2jsBigIntOf(y);

	return this;
};

go2jsBigInt.prototype.Quo = function(x, y) {
	// Quo is the quotient as it falls, cut off at the point and no further, so
	// the remainder is left holding the sign of what was divided.
	this.value = go2jsBigIntOf(x) / go2jsBigIntDivisor(y);

	return this;
};

go2jsBigInt.prototype.Rem = function(x, y) {
	this.value = go2jsBigIntOf(x) % go2jsBigIntDivisor(y);

	return this;
};

go2jsBigInt.prototype.Div = function(x, y) {
	// Div is a division that leaves nothing behind but a remainder no smaller than
	// nothing and no larger than the divisor however the divisor is written, so it
	// counts down to the divisor rather than off at the point: the remainder is
	// what is left of the dividend in the width of the divisor and the quotient is
	// whatever takes the rest of it away.
	const [dividend, divisor, remainder] = go2jsBigIntEuclidParts(x, y);

	this.value = (dividend - remainder) / divisor;

	return this;
};

go2jsBigInt.prototype.Mod = function(x, y) {
	// Mod is what such a division leaves over, so it too is never negative and
	// never as large as the divisor, whatever the sign of what was divided.
	const [, , remainder] = go2jsBigIntEuclidParts(x, y);

	this.value = remainder;

	return this;
};

go2jsBigInt.prototype.DivMod = function(x, y, modulus) {
	// DivMod is a division that leaves nothing behind written twice: the quotient
	// into the number it is called on and the remainder into the one given.
	const [dividend, divisor, remainder] = go2jsBigIntEuclidParts(x, y);

	this.value = (dividend - remainder) / divisor;

	go2jsBigIntInto(modulus, remainder);

	return [this, modulus];
};

go2jsBigInt.prototype.QuoRem = function(x, y, remainder) {
	// QuoRem is the division that falls at the point, written twice, where the
	// remainder is left holding the sign of what was divided.
	const dividend = go2jsBigIntOf(x);
	const divisor = go2jsBigIntDivisor(y);

	this.value = dividend / divisor;

	go2jsBigIntInto(remainder, dividend % divisor);

	return [this, remainder];
};

go2jsBigInt.prototype.Exp = function(x, count, modulus) {
	// Exp is the number raised to a power, and where a modulus is given it is the
	// number left after raising it as many times as the power says, counted off
	// the modulus each time so that nothing wider than the modulus is ever held.
	// A modulus of no value at all is no modulus, and a power of no value or less
	// leaves the number of one, since anything raised to nothing is one.
	const base = go2jsBigIntOf(x);
	const power = go2jsBigIntValue(count);
	const held = go2jsBigIntValue(modulus);
	let width = held < 0n ? -held : held;

	if (width < 1n) {
		width = 0n;
	}

	if (power <= 0n) {
		this.value = width === 0n ? 1n : 1n % width;

		return this;
	}

	// The power is walked from its top bit down, doubling what is held and
	// multiplying it by the base wherever the bit stands for one, since a power
	// of no width may be far wider than anything counted in one go.
	let held2 = width === 0n ? 1n : 1n % width;
	let square = width === 0n ? base : base % width;
	let left = power;

	while (left > 0n) {
		if ((left & 1n) === 1n) {
			held2 = width === 0n ? held2 * square : (held2 * square) % width;
		}

		square = width === 0n ? square * square : (square * square) % width;
		left >>= 1n;
	}

	// What is left of a power counted off a modulus is counted off in the width of
	// that modulus, so a power below nothing comes out above nothing, which is
	// where Go puts it as well.
	this.value = width === 0n ? held2 : (held2 < 0n ? held2 + width : held2);

	return this;
};

go2jsBigInt.prototype.GCD = function(x, y, a, b) {
	// GCD is the largest number that divides both without a remainder, and it is
	// written as no value at all or above however the two are written. x and y are
	// written so that the divisor is a*x + b*y, which is what makes a division of
	// a by the divisor come out with nothing in it, and where one of the two given
	// is no value at all the other answers for itself beside a count of one or
	// nothing, since nothing is a divisor of everything.
	const left = go2jsBigIntOf(a);
	const right = go2jsBigIntOf(b);

	if (go2jsBigIntAbs(left) === 0n || go2jsBigIntAbs(right) === 0n) {
		this.value = go2jsBigIntAbs(left) === 0n ? go2jsBigIntAbs(right) : go2jsBigIntAbs(left);

		if (go2jsBigIntGiven(x)) {
			go2jsBigIntInto(x, go2jsBigIntAbs(left) === 0n ? 0n : BigInt(go2jsBigIntSign(left)));
		}

		if (go2jsBigIntGiven(y)) {
			go2jsBigIntInto(y, go2jsBigIntAbs(right) === 0n ? 0n : BigInt(go2jsBigIntSign(right)));
		}

		return this;
	}

	// The two counts are found from the sizes of the two, since the count of one
	// belongs to one whichever way it is written, and the count beside the other is
	// then whatever takes the divisor off the whole of it.
	const [divisor, count] = go2jsBigIntBezout(left, right);

	this.value = divisor;

	if (go2jsBigIntGiven(x)) {
		go2jsBigIntInto(x, left < 0n ? -count : count);
	}

	if (go2jsBigIntGiven(y)) {
		const written = left < 0n ? -count : count;
		const rest = divisor - left * written;

		go2jsBigIntInto(y, rest / right);
	}

	return this;
};

go2jsBigInt.prototype.ModInverse = function(value, modulus) {
	// ModInverse is the count that undoes the number, which multiplies it into one
	// left over after being counted off the modulus as many times as it takes. A
	// modulus is counted in its own width, whatever its sign, and a number below
	// nothing is counted off the modulus first, so that what is counted off is the
	// number as it stands among the numbers of that width. Two numbers that share
	// something have no such count at all, and Go hands back nothing rather than
	// a count, leaving the number it was called on as it was.
	const held = go2jsBigIntOf(value);
	const written = go2jsBigIntDivisor(modulus);
	const width = written < 0n ? -written : written;
	const number = held < 0n ? go2jsBigIntMod(held, width) : held;
	const [divisor, count] = go2jsBigIntBezout(number, width);

	if (divisor !== 1n) {
		return null;
	}

	// The count comes out of the division wherever it lands, and what matters is
	// the count left over in the width of the modulus, so a count below nothing is
	// brought up by the width of the modulus once, which is as much as it takes.
	this.value = count < 0n ? count + width : count;

	return this;
};

go2jsBigInt.prototype.Neg = function(x) {
	this.value = -go2jsBigIntOf(x);

	return this;
};

go2jsBigInt.prototype.Abs = function(x) {
	const value = go2jsBigIntOf(x);

	this.value = value < 0n ? -value : value;

	return this;
};

go2jsBigInt.prototype.Not = function(x) {
	// Not is every bit turned over, which for a number written in twos complement
	// is the number one below the number itself, and JavaScript writes numbers
	// below nothing the same way, so the two agree without a width.
	this.value = ~go2jsBigIntOf(x);

	return this;
};

go2jsBigInt.prototype.And = function(x, y) {
	const width = go2jsBigIntWidth(x, y);

	this.value = go2jsBigIntBits(x, y, width, (left, right) => left & right);

	return this;
};

go2jsBigInt.prototype.Or = function(x, y) {
	const width = go2jsBigIntWidth(x, y);

	this.value = go2jsBigIntBits(x, y, width, (left, right) => left | right);

	return this;
};

go2jsBigInt.prototype.Xor = function(x, y) {
	const width = go2jsBigIntWidth(x, y);

	this.value = go2jsBigIntBits(x, y, width, (left, right) => left ^ right);

	return this;
};

go2jsBigInt.prototype.AndNot = function(x, y) {
	// AndNot is the bits of the first that the second does not have set, which
	// is the first with the bits of the second turned over put through it.
	const width = go2jsBigIntWidth(x, y);

	this.value = go2jsBigIntBits(x, y, width, (left, right) => left & ~right);

	return this;
};

go2jsBigInt.prototype.Lsh = function(x, count) {
	// Lsh is the number with that many bits of nothing standing on the right of
	// it, which is the number multiplied by two as many times.
	this.value = go2jsBigIntOf(x) << go2jsBigIntCount(count);

	return this;
};

go2jsBigInt.prototype.Rsh = function(x, count) {
	// Rsh is the number with that many bits of it falling off the right, and the
	// bits that fall are made up by the sign repeated, so a number below nothing
	// divided by two as many times still counts down rather than off at nothing.
	this.value = go2jsBigIntOf(x) >> go2jsBigIntCount(count);

	return this;
};

go2jsBigInt.prototype.Bit = function(index) {
	// Bit is one bit of the number at a place counted from the bottom, and a
	// number below nothing is written with as many bits above it as it has below,
	// so every bit of it above its width stands as it stands in twos complement,
	// which is what a shift of a number below nothing does in JavaScript as well.
	return Number((this.value >> go2jsBigIntCount(index)) & 1n);
};

go2jsBigInt.prototype.TrailingZeroBits = function() {
	// The trailing bits of nothing are no bits of the number at all, so a number
	// of no value has as many of them as a machine can count in one go.
	const value = this.value < 0n ? -this.value : this.value;

	// A number of no value at all has no bits of its own at all, so none of them
	// stand at the bottom of it and none is to be counted off the bottom.
	if (value === 0n) {
		return 0;
	}

	let held = 0n;

	while (((value >> held) & 1n) === 0n) {
		held += 1n;
	}

	return Number(held);
};

go2jsBigInt.prototype.Cmp = function(x) {
	// Cmp says which is the greater, so the number held is set against the one
	// given rather than the other way round.
	return go2jsBigIntSign(this.value - go2jsBigIntOf(x));
};

go2jsBigInt.prototype.CmpAbs = function(x) {
	// CmpAbs sets the two numbers by their size rather than by where they stand,
	// so two numbers the same size apart are neither above one another however
	// they are written.
	const held = go2jsBigIntAbs(this.value);
	const other = go2jsBigIntAbs(go2jsBigIntOf(x));

	return go2jsBigIntSign(held - other);
};

go2jsBigInt.prototype.Sign = function() {
	return go2jsBigIntSign(this.value);
};

go2jsBigInt.prototype.String = function() {
	return this.value.toString();
};

go2jsBigInt.prototype.Text = function(base) {
	const radix = Number(base);

	// A base is a place to count in, and a place of one, of none or of more than
	// the thirty six digits there are letters for is no place to count in at all.
	if (!Number.isInteger(radix) || radix < 2 || radix > 36) {
		throw new go2jsNativeRangeError("invalid base");
	}

	return this.value.toString(radix);
};

go2jsBigInt.prototype.SetString = function(text, base) {
	const value = go2jsBigIntParse(String(text), Number(base));

	// Text that is no number at all hands back nothing at all, because there is no
	// number in it to point to and what would have been read into cannot be told
	// from one number to another, and Go says what is left in a number that was
	// not read is not to be relied on.
	if (value === null) {
		return [null, false];
	}

	this.value = value;

	return [this, true];
};

go2jsBigInt.prototype.Int64 = function() {
	// A machine number of sixty four bits is read off the bottom of the number and
	// wraps about the ends of that width where the number is wider than it, which
	// is what the machine holds of such a number, and the sign of the number is
	// put back afterwards.
	const value = go2jsBigIntLowBits(this.value, 9223372036854775807n, 18446744073709551616n);

	return go2jsNarrow(value);
};

go2jsBigInt.prototype.Uint64 = function() {
	// A count of no width is read off the bottom of the number with the sign left
	// off, so a number below nothing answers with what is under it and a number
	// wider than the sixty four bits answers with what is left at the bottom.
	const value = this.value < 0n ? -this.value : this.value;

	return go2jsNarrow(value & 18446744073709551615n);
};

go2jsBigInt.prototype.BitLen = function() {
	// The bits a number holds are counted from the top down rather than written
	// out, because writing a number wider than a double as text costs a character
	// for every bit of it and a shift of thirty two at a time costs none for the
	// bits below where it stops.
	const value = this.value < 0n ? -this.value - 1n : this.value;
	let held = 0n;

	while (value >> BigInt(held) > 0xffffffffn) {
		held += 32n;
	}

	return Number(held) + 32 - Math.clz32(Number(value >> held));
};

go2jsBigInt.prototype.IsInt64 = function() {
	return this.value >= -9223372036854775808n && this.value <= 9223372036854775807n;
};

go2jsBigInt.prototype.IsUint64 = function() {
	return this.value >= 0n && this.value <= 18446744073709551615n;
};

go2jsBigInt.prototype.Bytes = function() {
	// Bytes is the number written as bytes from the heaviest to the lightest,
	// which is how a number travels between machines and out of a program, and it
	// is the number under nothing rather than the number itself, so a number below
	// nothing is written as what is under it.
	const value = go2jsBigIntAbs(this.value);

	if (value === 0n) {
		return [];
	}

	const held = value.toString(16);
	const written = held.length % 2 === 0 ? held : "0" + held;
	const answered = [];

	for (let i = 0; i < written.length; i += 2) {
		answered.push(parseInt(written.slice(i, i + 2), 16));
	}

	return answered;
};

go2jsBigInt.prototype.SetBytes = function(bytes) {
	// SetBytes is the number read back out of bytes written from the heaviest to
	// the lightest, with nothing above nothing counted as a byte of no value.
	const source = bytes === null || bytes === undefined ? [] : Array.from(bytes);
	let held = 0n;

	for (const item of source) {
		const byte = BigInt(item);

		if (byte < 0n || byte > 255n) {
			throw new go2jsNativeRangeError("byte out of range in SetBytes");
		}

		held = (held << 8n) | byte;
	}

	this.value = held;

	return this;
};

go2jsBigInt.prototype.Float64 = function() {
	// Float64 is the double standing nearest the number, with the way it stands
	// said as well: below where the double falls short of it, above where it goes
	// past it, and exact where it lands on the number itself.
	const held = Number(this.value);

	if (!Number.isFinite(held)) {
		return [held, "Below"];
	}

	const below = BigInt(held);
	const above = held < 0 ? below - 1n : below + 1n;

	if (below === this.value) {
		return [held, "Exact"];
	}

	return [held, below > this.value ? "Above" : "Below"];
};

// go2jsBigIntParse reads a number written as text in a given base, and answers
// nothing at all rather than a number when the text is not one: a base of zero is
// read off the front of the text rather than named, and every base has to be one
// there are digits for.
function go2jsBigIntParse(text, base) {
	// A base is a place to count in, and there are digits for the places from two
	// to the thirty six and no letters for any other, so a base outside those is a
	// fault in Go rather than a number that is read differently.
	if (base !== 0 && (base < 2 || base > 36)) {
		throw new go2jsNativeRangeError("invalid number base " + base);
	}

	if (!Number.isInteger(base)) {
		return null;
	}

	// A sign is a part of the number and not of the base it is written in, so it
	// is taken off before the rest is read and put back afterwards.
	let sign = 1n;
	let rest = text;

	if (rest.startsWith("+") || rest.startsWith("-")) {
		sign = rest.startsWith("-") ? -1n : 1n;
		rest = rest.slice(1);
	}

	let value = null;

	if (base === 0) {
		const lowered = rest.toLowerCase();

		if (lowered.startsWith("0b")) {
			value = go2jsBigIntDigits(go2jsBigIntAfterPrefix(lowered.slice(2)), 2, true);
		} else if (lowered.startsWith("0o")) {
			value = go2jsBigIntDigits(go2jsBigIntAfterPrefix(lowered.slice(2)), 8, true);
		} else if (lowered.startsWith("0x")) {
			value = go2jsBigIntDigits(go2jsBigIntAfterPrefix(lowered.slice(2)), 16, true);
		} else if (lowered.length > 1 && lowered.startsWith("0")) {
			value = go2jsBigIntDigits(go2jsBigIntAfterPrefix(lowered.slice(1)), 8, true);
		} else {
			// No prefix was written, so there is no base standing between an
			// underscore and the first digit, and one standing first is no number.
			value = go2jsBigIntDigits(lowered, 10, true);
		}
	} else {
		value = go2jsBigIntDigits(rest, base, false);
	}

	return value === null ? null : sign * value;
}

// go2jsBigIntAfterPrefix takes the digits of a number whose base was read off its
// front, where an underscore may stand between the base and the first digit, the
// way Go writes a number apart from the base it is in.
function go2jsBigIntAfterPrefix(text) {
	return text.startsWith("_") ? text.slice(1) : text;
}

// go2jsBigIntDigits reads the digits of a number in a base that was settled on
// already, and answers nothing at all rather than a number when a digit is not
// one of the base or an underscore stands where Go would not have it stand.
function go2jsBigIntDigits(text, base, underscoresAllowed) {
	if (base < 2 || base > 36 || text === "") {
		return null;
	}

	let written = "";

	for (let index = 0; index < text.length; index++) {
		const character = text[index];

		if (character === "_") {
			// An underscore stands for nothing at all, and is written between two
			// digits and only where the base was read off the front of the number
			// rather than named outright, so one at either end of the digits or two
			// of them running together is not a number.
			if (!underscoresAllowed || index === 0 || index === text.length - 1 ||
				text[index - 1] === "_") {
				return null;
			}

			continue;
		}

		const digit = go2jsBigIntDigit(character);

		if (digit < 0 || digit >= base) {
			return null;
		}

		written += character;
	}

	if (written === "") {
		return null;
	}

	const radix = BigInt(base);
	let value = 0n;

	for (let index = 0; index < written.length; index++) {
		value = value * radix + BigInt(go2jsBigIntDigit(written[index]));
	}

	return value;
}

// The kinds of value a record can carry are the kinds Go names, in the order it
// names them, where any comes first because a value nobody told what it holds
// holds nothing in particular.
const go2jsSlogKind = {
	Any: 0,
	Bool: 1,
	Duration: 2,
	Float64: 3,
	Int64: 4,
	String: 5,
	Time: 6,
	Uint64: 7,
	Group: 8,
	LogValuer: 9,
};

function go2jsSlogKindName(kind) {
	switch (kind) {
	case go2jsSlogKind.Any: return "Any";
	case go2jsSlogKind.Bool: return "Bool";
	case go2jsSlogKind.Duration: return "Duration";
	case go2jsSlogKind.Float64: return "Float64";
	case go2jsSlogKind.Int64: return "Int64";
	case go2jsSlogKind.String: return "String";
	case go2jsSlogKind.Time: return "Time";
	case go2jsSlogKind.Uint64: return "Uint64";
	case go2jsSlogKind.Group: return "Group";
	case go2jsSlogKind.LogValuer: return "LogValuer";
	}

	return "<unknown slog.Kind>";
}

// go2jsSlogLevelName is the name a level is written under, which is the name of
// the level itself when it is one of the four Go names and the name of the level
// it stands beside followed by how far it stands from it when it is not, the way
// Go writes DEBUG+1 for a level one step above DEBUG.
function go2jsSlogLevelName(level) {
	const held = Number(level) | 0;
	const named = function(base, from) {
		const off = held - from;

		return off === 0 ? base : base + (off < 0 ? "" : "+") + off;
	};

	if (held < go2jsSlogKind.Any) {
		return named("DEBUG", -4);
	}

	if (held < 4) {
		return named("INFO", 0);
	}

	if (held < 8) {
		return named("WARN", 4);
	}

	return named("ERROR", 8);
}

// go2jsSlogLevelOf is the level a holder of one reports, which is what a handler
// asked for the lowest level it writes asks for.
function go2jsSlogLevelOf(leveler) {
	leveler = go2jsUnwrap(go2jsUntyped(leveler));
	if (leveler !== null && leveler !== undefined && typeof leveler.Level === "function") {
		return leveler.Level() | 0;
	}

	return Number(leveler) | 0;
}

// go2jsSlogLevelNumber is a level as the number behind its name, read through
// whatever it was wrapped in on the way here.
function go2jsSlogLevelNumber(level) {
	return Number(go2jsUnwrap(go2jsUntyped(level))) | 0;
}

// go2jsSlogValue holds one value of a record together with the kind of value it
// is, since a number in JavaScript is a number whichever kind Go gave it, and a
// moment or a span of time is a value JavaScript has no kind of its own for.
function go2jsSlogValue(kind, num, str, any) {
	// A value built with nothing said to it is a value of no kind in particular,
	// which is what a value of the program that was never filled in is.
	this.kind = kind === undefined ? go2jsSlogKind.Any : kind;
	this.num = num === undefined ? 0 : num;
	this.str = str === undefined ? "" : str;
	this.any = any === undefined ? null : any;
}

go2jsSlogValue.prototype.Kind = function() {
	return this.kind;
};

go2jsSlogValue.prototype.Any = function() {
	switch (this.kind) {
	case go2jsSlogKind.String:
		return this.str;
	case go2jsSlogKind.Int64:
	case go2jsSlogKind.Uint64:
	case go2jsSlogKind.Duration:
		return Number(this.num);
	case go2jsSlogKind.Float64:
		return Number(this.num);
	case go2jsSlogKind.Bool:
		return this.num === true || this.num === 1;
	case go2jsSlogKind.Time:
		return this.any;
	case go2jsSlogKind.Group:
		return this.any.slice();
	}

	return this.any;
};

// A reader that is handed a value of another kind than the one it reads is a
// reader that was asked the wrong question, which Go answers by panicking.
go2jsSlogValue.prototype.Bool = function() {
	go2jsSlogValueWantKind(this, go2jsSlogKind.Bool);

	return this.num === true || this.num === 1;
};

go2jsSlogValue.prototype.Duration = function() {
	go2jsSlogValueWantKind(this, go2jsSlogKind.Duration);

	return go2jsDuration(this.num);
};

go2jsSlogValue.prototype.Float64 = function() {
	go2jsSlogValueWantKind(this, go2jsSlogKind.Float64);

	return Number(this.num);
};

go2jsSlogValue.prototype.Int64 = function() {
	go2jsSlogValueWantKind(this, go2jsSlogKind.Int64);

	return Number(this.num);
};

go2jsSlogValue.prototype.Uint64 = function() {
	go2jsSlogValueWantKind(this, go2jsSlogKind.Uint64);

	return Number(this.num);
};

go2jsSlogValue.prototype.Time = function() {
	go2jsSlogValueWantKind(this, go2jsSlogKind.Time);

	return this.any;
};

go2jsSlogValue.prototype.Group = function() {
	go2jsSlogValueWantKind(this, go2jsSlogKind.Group);

	return this.any;
};

// go2jsSlogValueWantKind is the complaint of a reader that was handed a value of
// another kind than the one it reads.
function go2jsSlogValueWantKind(value, wanted) {
	if (value.kind !== wanted) {
		go2jsPanic("Value kind is " + go2jsSlogKindName(value.kind) + ", not " +
			go2jsSlogKindName(wanted));
	}
}

go2jsSlogValue.prototype.GroupValue = function() {
	return this.any.slice();
};

go2jsSlogValue.prototype.LogValue = function() {
	return this.any;
};

go2jsSlogValue.prototype.String = function() {
	return go2jsSlogValueText(this);
};

go2jsSlogValue.prototype.Equal = function(other) {
	other = go2jsSlogUnwrapValue(other);

	if (other === null || other === undefined || other.kind !== this.kind) {
		return false;
	}

	if (this.kind === go2jsSlogKind.String) {
		return this.str === other.str;
	}

	if (this.kind === go2jsSlogKind.Group) {
		if (this.any.length !== other.any.length) {
			return false;
		}

		return this.any.every((attr, index) => go2jsSlogAttrEqual(attr, other.any[index]));
	}

	if (this.kind === go2jsSlogKind.Time) {
		return go2jsSlogTimeText(this.any) === go2jsSlogTimeText(other.any);
	}

	return go2jsSlogValueText(this) === go2jsSlogValueText(other);
};

go2jsSlogValue.prototype.Resolve = function() {
	let value = this;

	// A value the program asked to be asked again is asked again and again, and
	// a value that keeps asking stops at the point where Go stops asking too, so
	// that a value which never settles is answered rather than waited on.
	for (let count = 0; count < 100; count++) {
		if (value.kind !== go2jsSlogKind.LogValuer) {
			return value;
		}

		value = go2jsSlogLogValue(value.any);
	}

	return go2jsSlogAnyValueText("LogValue called too many times on Value of type " + go2jsSlogTypeNameOf(this.any));
};

// go2jsSlogLogValue asks a value of the program for the value it stands for.
function go2jsSlogLogValue(valuer) {
	try {
		valuer = go2jsUnwrap(go2jsUntyped(valuer));

		return go2jsSlogAsValue(go2jsSlogInvoke(valuer.LogValue, valuer, []));
	} catch (err) {
		// What the value said on its way down is not what is reported, since
		// what it did not reach says more than what it said.
		return go2jsSlogAnyValueText("LogValue panicked\n" + go2jsSlogPanicSite(valuer));
	}
}

// go2jsSlogPanicSite is where a value that would not answer was asked, since a
// fault a value raised says nothing of the fault that raised it.
function go2jsSlogPanicSite(valuer) {
	const name = go2jsSlogTypeNameOf(valuer);

	if (name === "<nil>" || name === "") {
		return "";
	}

	return "called from " + name;
}

// go2jsSlogAsValue reads a value that answers for being one, which is how a
// value the program made with the constructor of a Value of the package comes
// back holding what it holds.
function go2jsSlogAsValue(value) {
	if (value instanceof go2jsSlogValue) {
		return value;
	}

	return go2jsSlogAnyValue(value, null);
}

// go2jsSlogTypeNameOf is the name Go gives a value, which is the name a value
// that writes itself as text is asked for under.
function go2jsSlogTypeNameOf(value) {
	if (value === null || value === undefined) {
		return "<nil>";
	}

	const named = go2jsGoTypeName(value);

	if (named !== undefined && named !== null && named !== "") {
		return String(named);
	}

	return typeof value;
}

// go2jsSlogValueText is a value written the way Go writes one with fmt.Sprint,
// which is how a value of a kind of its own is written inside a group.
function go2jsSlogValueText(value) {
	switch (value.kind) {
	case go2jsSlogKind.String:
		return value.str;
	case go2jsSlogKind.Int64:
	case go2jsSlogKind.Uint64:
		return String(value.num);
	case go2jsSlogKind.Float64:
		return go2jsSprintf("%v", Number(value.num));
	case go2jsSlogKind.Bool:
		return value.num === true || value.num === 1 ? "true" : "false";
	case go2jsSlogKind.Duration:
		return go2jsDurationStringBig(BigInt(value.num));
	case go2jsSlogKind.Time:
		return go2jsSlogTimeText(value.any);
	case go2jsSlogKind.Group:
		return "[" + value.any.map(attr => attr.String()).join(" ") + "]";
	case go2jsSlogKind.Any:
	case go2jsSlogKind.LogValuer:
		return go2jsSprintf("%v", value.any);
	}

	return "<unknown slog.Value>";
}

function go2jsSlogStringValue(value) {
	return new go2jsSlogValue(go2jsSlogKind.String, 0,
		go2jsRawText(go2jsSlogTextOf(value)), null);
}

function go2jsSlogInt64Value(value) {
	return new go2jsSlogValue(go2jsSlogKind.Int64, go2jsSlogBigInt(go2jsUnwrap(go2jsUntyped(value))), "", null);
}

function go2jsSlogUint64Value(value) {
	return new go2jsSlogValue(go2jsSlogKind.Uint64, go2jsSlogBigInt(go2jsUnwrap(go2jsUntyped(value))), "", null);
}

function go2jsSlogFloat64Value(value) {
	return new go2jsSlogValue(go2jsSlogKind.Float64, Number(go2jsUnwrap(go2jsUntyped(value))), "", null);
}

function go2jsSlogBoolValue(value) {
	value = go2jsUnwrap(go2jsUntyped(value));

	return new go2jsSlogValue(go2jsSlogKind.Bool, value === true || value === 1, "", null);
}

function go2jsSlogTimeValue(value) {
	return new go2jsSlogValue(go2jsSlogKind.Time, 0, "",
		go2jsUnwrap(go2jsUntyped(value)));
}

function go2jsSlogDurationValue(value) {
	return new go2jsSlogValue(go2jsSlogKind.Duration, go2jsSlogBigInt(go2jsSlogNumberOf(value)), "", null);
}

// The attributes of a group of values are given one after another, or as the
// one list of them the program already had, since either says the same thing.
function go2jsSlogGroupValue(...attrs) {
	return new go2jsSlogValue(go2jsSlogKind.Group, 0, "", go2jsSlogFlattenAttrs(attrs));
}

function go2jsSlogFlattenAttrs(attrs) {
	const flat = [];

	for (const attr of attrs) {
		if (Array.isArray(attr)) {
			flat.push(...attr);
			continue;
		}

		if (attr !== undefined) {
			flat.push(attr);
		}
	}

	return flat;
}

// go2jsSlogAnyValue is a value read out of one of Go's own shapes, which the
// kinds of the program are told apart by, and which a value nothing was said
// about is read the way the shape it has reads itself.
function go2jsSlogAnyValue(value, types) {
	const type = types === null || types === undefined ? "" : String(types);

	switch (type) {
	case "string":
		return go2jsSlogStringValue(value);
	case "bool":
		return go2jsSlogBoolValue(value);
	case "int":
	case "int8":
	case "int16":
	case "int32":
	case "int64":
	case "untyped int":
	case "untyped rune":
		return go2jsSlogInt64Value(value);
	case "uint":
	case "uint8":
	case "uint16":
	case "uint32":
	case "uint64":
	case "uintptr":
	case "untyped uint":
		return go2jsSlogUint64Value(value);
	case "float32":
	case "float64":
	case "untyped float":
	case "untyped complex":
		return go2jsSlogFloat64Value(value);
	case "time.Duration":
		return go2jsSlogDurationValue(value);
	case "time.Time":
		return go2jsSlogTimeValue(value);
	case "slog.Value":
	case "log/slog.Value":
		value = go2jsUnwrap(go2jsUntyped(value));

		return value instanceof go2jsSlogValue ? value : go2jsSlogAnyValueText(String(value));
	case "slog.Attr":
	case "log/slog.Attr":
		// An attribute handed over as a value is kept whole, since what the
		// words would have written of it is written as the pair it stands for.
		value = go2jsUnwrap(go2jsUntyped(value));

		return value instanceof go2jsSlogAttr
			? new go2jsSlogValue(go2jsSlogKind.Any, 0, "", value)
			: go2jsSlogAnyValueText(String(value));
	case "[]slog.Attr":
	case "[]log/slog.Attr":
		return go2jsSlogGroupValue(go2jsSlogAttrsOf(go2jsUnwrap(go2jsUntyped(value))));
	}

	// A value of a shape the runtime has of its own is a number, and a number
	// that counts without a remainder is the whole number Go would have written
	// it as, while one that carries a remainder is a length that is not whole.
	if (type === "") {
		if (typeof value === "string") {
			return go2jsSlogStringValue(value);
		}

		if (typeof value === "boolean") {
			return go2jsSlogBoolValue(value);
		}

		if (typeof value === "bigint") {
			return go2jsSlogInt64Value(value);
		}

		if (typeof value === "number") {
			return Number.isInteger(value) ? go2jsSlogInt64Value(value) : go2jsSlogFloat64Value(value);
		}
	}

	// A value the program gave a LogValue method to is asked what it stands for
	// before it is written, and the kind of it is the kind that asking gives it.
	if (go2jsSlogIsLogValuer(value)) {
		return new go2jsSlogValue(go2jsSlogKind.LogValuer, 0, "", go2jsUnwrap(go2jsUntyped(value)));
	}

	return new go2jsSlogValue(go2jsSlogKind.Any, 0, "", value);
}

// go2jsSlogIsLogValuer is whether a value answers for being one, since a value
// the program gave a LogValue method to is asked what it stands for before it is
// written, rather than being written as the value it holds.
function go2jsSlogIsLogValuer(value) {
	value = go2jsUnwrap(go2jsUntyped(value));

	return value !== null && value !== undefined && typeof value === "object" &&
		typeof value.LogValue === "function";
}

// go2jsSlogAnyValueText is a value the program holds that stands for a fault,
// which Go keeps as a value of its own so that writing it out says what is
// wrong rather than that something is.
function go2jsSlogAnyValueText(text) {
	return new go2jsSlogValue(go2jsSlogKind.Any, 0, "", go2jsSlogErrorValue(text));
}

// go2jsSlogBigInt is a whole number of any width, since a number of more digits
// than a double holds is still a whole number and must not be rounded to fit.
function go2jsSlogBigInt(value) {
	if (typeof value === "bigint") {
		return value;
	}

	return BigInt(Math.trunc(Number(value) || 0));
}

// go2jsSlogNumberOf is the number a value stands for, however the runtime was
// handed it, since a typed or boxed value is a number written in a wrapper.
function go2jsSlogNumberOf(value) {
	return Number(go2jsUnwrap(go2jsUntyped(value)));
}

// go2jsSlogTextOf is the text a value stands for, whether it was written as one
// already or is something that says what it is.
function go2jsSlogTextOf(value) {
	value = go2jsUnwrap(go2jsUntyped(value));

	if (value === null || value === undefined) {
		return value === null ? "<nil>" : "<nil>";
	}

	if (typeof value === "string") {
		return value;
	}

	if (typeof value === "object" && typeof value.String === "function") {
		return String(go2jsSlogInvoke(value.String, value, []));
	}

	return String(value);
}

// go2jsSlogAttr holds a key with the value it stands for, which is the pair a
// record carries beside its message.
function go2jsSlogAttr(key, value) {
	this.Key = key === undefined ? "" : go2jsRawText(key);
	this.Value = value instanceof go2jsSlogValue ? value : new go2jsSlogValue(go2jsSlogKind.Any, 0, "", value);
}

go2jsSlogAttr.prototype.String = function() {
	return this.Key + "=" + this.Value.String();
};

go2jsSlogAttr.prototype.Equal = function(other) {
	return go2jsSlogAttrEqual(this, other);
};

go2jsSlogAttr.prototype.Value = function() {
	return this.Value;
};

function go2jsSlogAttrEqual(one, other) {
	one = go2jsSlogUnwrapAttr(one);
	other = go2jsSlogUnwrapAttr(other);

	if (one === null || one === undefined || other === null || other === undefined) {
		return false;
	}

	return one.Key === other.Key && one.Value.Equal(other.Value);
}

// An attribute or a value of the program can arrive wrapped in whatever it was
// typed as, and what a comparison is asked about is what is inside.
function go2jsSlogUnwrapAttr(attr) {
	attr = go2jsUnwrap(go2jsUntyped(attr));

	return attr instanceof go2jsSlogAttr ? attr : null;
}

function go2jsSlogUnwrapValue(value) {
	value = go2jsUnwrap(go2jsUntyped(value));

	return value instanceof go2jsSlogValue ? value : null;
}

// go2jsSlogAttrEmpty is an attribute nobody said anything about, which is an
// empty key with a value that holds nothing, and which is written nowhere.
function go2jsSlogAttrEmpty(attr) {
	if (attr.Key !== "") {
		return false;
	}

	const any = attr.Value.any;

	return attr.Value.kind === go2jsSlogKind.Any &&
		(any === null || any === undefined || any === go2jsInterface(null));
}

// go2jsSlogEmptyGroup is a group with nothing in it, which Go leaves out of a
// record because a group nobody filled says nothing a reader would want.
function go2jsSlogEmptyGroup(value) {
	return value.kind === go2jsSlogKind.Group && value.Group().length === 0;
}

function go2jsSlogString(key, value) {
	return new go2jsSlogAttr(key, go2jsSlogStringValue(value));
}

function go2jsSlogInt(key, value) {
	return new go2jsSlogAttr(key, go2jsSlogInt64Value(value));
}

function go2jsSlogTime(key, value) {
	return new go2jsSlogAttr(key, go2jsSlogTimeValue(value));
}

function go2jsSlogAny(key, value, type) {
	return new go2jsSlogAttr(key, go2jsSlogAnyValue(value, type));
}

// go2jsSlogLevelAttr is the level of a record as the attribute a program may
// rewrite, which carries the name the level is written under rather than the
// number behind it, since that is what a level writes as.
function go2jsSlogLevelAttr(key, level) {
	return new go2jsSlogAttr(key, go2jsSlogAnyValue(go2jsSlogLevelHolder(level), "log/slog.Level"));
}

// go2jsSlogLevelHolder is a level as a value of its own, which writes as the
// name it is known by however it is written: as text, as JSON, or as itself.
function go2jsSlogLevelHolder(level) {
	const name = go2jsSlogLevelName(level);

	return {
		level: Number(level) | 0,
		MarshalText: function() {
			return name;
		},
		MarshalJSON: function() {
			return "\"" + go2jsSlogEscapeJSON(name) + "\"";
		},
		String: function() {
			return name;
		},
	};
}

// go2jsSlogAttrsOf reads the attributes a slice of them holds, since a slice the
// program made is a list the runtime walks the same way it walks one of its own.
function go2jsSlogAttrsOf(value) {
	if (Array.isArray(value)) {
		return value.map(attr => go2jsSlogAsAttr(attr));
	}

	return [];
}

function go2jsSlogAsAttr(value) {
	value = go2jsUntyped(value);
	if (value instanceof go2jsSlogAttr) {
		return value;
	}

	return new go2jsSlogAttr("!BADKEY", value);
}

// go2jsSlogArgsToAttrs reads the trailing arguments of a log call as the pairs
// they stand for, where a name in front of a value makes the pair and anything
// else stands for a value whose name was left out.
function go2jsSlogArgsToAttrs(args, types) {
	const attrs = [];
	let index = 0;

	while (index < args.length) {
		const value = args[index];

		if (typeof value === "string") {
			if (index + 1 === args.length) {
				attrs.push(new go2jsSlogAttr("!BADKEY", go2jsSlogStringValue(value)));
				break;
			}

			attrs.push(new go2jsSlogAttr(value, go2jsSlogAnyValue(args[index + 1],
				types === null || types === undefined ? null : types[index + 1])));

			index += 2;
			continue;
		}

		const v2 = go2jsUntyped(value);
		if (v2 instanceof go2jsSlogAttr) {
			attrs.push(v2);
			index++;
			continue;
		}

		attrs.push(new go2jsSlogAttr("!BADKEY",
			go2jsSlogAnyValue(value, types === null || types === undefined ? null : types[index])));

		index++;
	}

	return attrs;
}

// go2jsSlogRecord is one record of what a program logged: when it happened, how
// loud it was, what was said and what was said beside it.
function go2jsSlogRecord(time, level, message, pc) {
	this.Time = time === undefined || time === null ? go2jsTimeZero() : time;
	this.Message = message === undefined ? "" : go2jsRawText(message);
	this.Level = Number(level) | 0;
	this.PC = pc === undefined || pc === null ? 0 : Number(pc);

	// The attributes of a record are kept as one list: Go keeps the first few
	// inside the record and the rest beside it, which is an arrangement a list
	// does not need.
	this.__go2js_attrs = [];
	this.__go2js_source = "";
}

go2jsSlogRecord.prototype.NumAttrs = function() {
	return this.__go2js_attrs.length;
};

go2jsSlogRecord.prototype.Attrs = function(visit) {
	// A visitor is a function of the program behind whatever it was handed
	// as, and it is called the way a callback is called: run to its end here,
	// waiting where it waits.
	const seen = go2jsUnwrap(go2jsUntyped(visit));

	for (const attr of this.__go2js_attrs.slice()) {
		if (go2jsCallNow(seen, null, [attr]) === false) {
			return;
		}
	}
};

go2jsSlogRecord.prototype.Clone = function() {
	const clone = new go2jsSlogRecord(this.Time, this.Level, this.Message, this.PC);

	clone.__go2js_attrs = this.__go2js_attrs.slice();
	clone.__go2js_source = this.__go2js_source;

	return clone;
};

go2jsSlogRecord.prototype.AddAttrs = function(attrs) {
	for (const attr of go2jsSlogAttrsOf(attrs === undefined ? [] : attrs)) {
		if (!go2jsSlogEmptyGroup(attr.Value)) {
			this.__go2js_attrs.push(attr);
		}
	}
};

go2jsSlogRecord.prototype.Add = function(args, types) {
	for (const attr of go2jsSlogArgsToAttrs(args === undefined ? [] : args, types)) {
		if (!go2jsSlogEmptyGroup(attr.Value)) {
			this.__go2js_attrs.push(attr);
		}
	}
};

// go2jsSlogLevelVar holds a level that can be changed while a program runs,
// which is how the lowest level a handler writes is moved without stopping.
function go2jsSlogLevelVar(level) {
	this.level = go2jsSlogNumberOf(level) | 0;
}

go2jsSlogLevelVar.prototype.Level = function() {
	return this.level | 0;
};

go2jsSlogLevelVar.prototype.Set = function(level) {
	this.level = go2jsSlogNumberOf(level) | 0;
};

go2jsSlogLevelVar.prototype.String = function() {
	return go2jsSlogLevelName(this.level);
};

go2jsSlogLevelVar.prototype.Swap = function(level) {
	const held = this.level | 0;

	this.level = Number(level) | 0;

	return held;
};

go2jsSlogLevelVar.prototype.CompareAndSwap = function(wanted, level) {
	if ((this.level | 0) !== (Number(wanted) | 0)) {
		return false;
	}

	this.level = Number(level) | 0;

	return true;
};

// A level is a named number, so a verb or a writer that asks what a value is
// asks the level the name it is written under, whichever of the two names of
// the package the value is holding was written.
go2jsRegisterMethod("slog.Level.String", go2jsSlogLevelStringMethod);

go2jsRegisterMethod("log/slog.Level.String", go2jsSlogLevelStringMethod);

go2jsRegisterMethod("slog.Level.MarshalText", go2jsSlogLevelStringMethod);

go2jsRegisterMethod("log/slog.Level.MarshalText", go2jsSlogLevelStringMethod);

go2jsRegisterMethod("slog.Level.MarshalJSON", go2jsSlogLevelMarshalJSONMethod);

go2jsRegisterMethod("log/slog.Level.MarshalJSON", go2jsSlogLevelMarshalJSONMethod);

// A kind is a named number too, and fmt writes the name of the kind rather
// than the number behind one.
go2jsRegisterMethod("slog.Kind.String", go2jsSlogKindName);

go2jsRegisterMethod("log/slog.Kind.String", go2jsSlogKindName);

function go2jsSlogLevelStringMethod(level) {
	return go2jsSlogLevelName(level);
}

function go2jsSlogLevelMarshalJSONMethod(level) {
	return go2jsSlogStringValue(go2jsSlogLevelName(level));
}

// go2jsSlogHandler is one of the two handlers the package builds itself: what it
// writes, where it writes it and how it was asked to write it, together with the
// attributes and groups a program built it up with before the record arrived.
function go2jsSlogHandler(json, writer, opts) {
	const handler = {
		__go2js_slog_builtin: true,
		json: json,
		writer: writer,
		opts: opts === undefined || opts === null ? {} : opts,
		preformatted: "",
		groupPrefix: "",
		groups: [],
		nOpenGroups: 0,
	};

	handler.Enabled = function(ctx, level) {
		void ctx;
		return go2jsSlogHandlerEnabled(this, level);
	};

	handler.Handle = function(ctx, record) {
		void ctx;
		return go2jsSlogHandlerHandle(this, record);
	};

	handler.WithAttrs = function(attrs) {
		return go2jsSlogHandlerWithAttrs(this, attrs);
	};

	handler.WithGroup = function(name) {
		return go2jsSlogHandlerWithGroup(this, name);
	};

	return handler;
}

// go2jsSlogHandlerEnabled is whether a record of this level is one the handler
// writes, which it is once it is as loud as the lowest level it was asked for.
function go2jsSlogHandlerEnabled(handler, level) {
	const asked = handler.opts === undefined || handler.opts === null ? null : handler.opts;

	if (asked === null || asked.Level === undefined || asked.Level === null) {
		return (Number(level) | 0) >= 0;
	}

	return (Number(level) | 0) >= go2jsSlogLevelOf(asked.Level);
}

// go2jsSlogHandlerClone is a copy of a handler that can be given attributes of
// its own, which is how a logger is built up without disturbing the one it came
// from.
function go2jsSlogHandlerClone(handler) {
	const clone = go2jsSlogHandler(handler.json, handler.writer, handler.opts);

	clone.preformatted = handler.preformatted;
	clone.groupPrefix = handler.groupPrefix;
	clone.groups = handler.groups.slice();
	clone.nOpenGroups = handler.nOpenGroups;

	return clone;
}

// go2jsSlogHandlerWithAttrs is a handler that writes the attributes it already
// wrote before the ones it is given now, written once and for all, since they
// are the same on every record that follows.
function go2jsSlogHandlerWithAttrs(handler, attrs) {
	const held = go2jsSlogAttrsOf(attrs);

	// Attributes that are all groups nobody filled write nothing, so a list of
	// nothing but them leaves the handler standing as it does.
	if (go2jsSlogCountEmptyGroups(held) === held.length) {
		return handler;
	}

	const next = go2jsSlogHandlerClone(handler);
	const state = go2jsSlogState(next, held !== null ? next.preformatted : "");
	state.prefix.push(next.groupPrefix);
	go2jsSlogOpenGroups(state);

	const mark = state.buf.length;

	if (go2jsSlogAppendAttrs(state, held)) {
		next.preformatted = state.buf.join("");
		next.groupPrefix = state.prefix.join("");
		next.nOpenGroups = next.groups.length;
	} else {
		next.preformatted = state.buf.slice(0, mark).join("");
	}

	return next;
}

// go2jsSlogHandlerWithGroup is a handler that writes everything it is given
// under the name of one more group.
function go2jsSlogHandlerWithGroup(handler, name) {
	const next = go2jsSlogHandlerClone(handler);

	next.groups.push(go2jsRawText(name));

	return next;
}

// go2jsSlogHandlerHandle writes one record, which is one line however many
// attributes the record carries, since a line that ran on would be read as a
// record of its own.
function go2jsSlogHandlerHandle(handler, record) {
	const state = go2jsSlogState(handler, "");

	if (handler.json) {
		go2jsSlogWrite(state, "{");
	}

	const replace = go2jsSlogReplaceOf(handler);
	const groups = state.groups;

	// The attributes a handler writes about a record are not inside any group,
	// so a program asked to rewrite one of them is told of no group at all.
	state.groups = null;

	if (!go2jsSlogTimeIsZero(record.Time)) {
		if (replace === null) {
			go2jsSlogAppendKey(state, "time");
			go2jsSlogAppendTime(state, record.Time);
		} else {
			go2jsSlogAppendAttr(state, go2jsSlogTime("time", record.Time));
		}
	}

	if (replace === null) {
		go2jsSlogAppendKey(state, "level");
		go2jsSlogAppendString(state, go2jsSlogLevelName(record.Level));
	} else {
		go2jsSlogAppendAttr(state, go2jsSlogLevelAttr("level", record.Level));
	}

	if (handler.opts !== undefined && handler.opts !== null && handler.opts.AddSource === true) {
		go2jsSlogAppendAttr(state, go2jsSlogAny("source", go2jsSlogSourceOf(record.__go2js_source), ""));
	}

	if (replace === null) {
		go2jsSlogAppendKey(state, "msg");
		go2jsSlogAppendString(state, record.Message);
	} else {
		go2jsSlogAppendAttr(state, go2jsSlogString("msg", record.Message));
	}

	state.groups = groups;
	go2jsSlogAppendNonBuiltIns(state, record);
	go2jsSlogWrite(state, "\n");

	return go2jsSlogWriteLine(handler, state.buf.join(""));
}

// go2jsSlogReplaceOf is the function a program gave for rewriting attributes,
// which is none at all unless it gave one.
function go2jsSlogReplaceOf(handler) {
	const opts = handler.opts === undefined || handler.opts === null ? null : handler.opts;

	if (opts === null || typeof opts.ReplaceAttr !== "function") {
		return null;
	}

	return opts.ReplaceAttr;
}

// go2jsSlogState is the writing of one record as it stands, which is gathered as
// pieces so that a group found to hold nothing can be taken back off the end.
function go2jsSlogState(handler, written) {
	const opts = handler.opts === undefined || handler.opts === null ? null : handler.opts;
	const state = {
		h: handler,
		buf: written === "" ? [] : [written],
		sep: "",
		prefix: [],
		groups: null,
	};

	if (opts !== null && typeof opts.ReplaceAttr === "function") {
		state.groups = handler.groups.slice(0, handler.nOpenGroups);
	}

	if (written !== "") {
		state.sep = handler.json ? "," : " ";

		if (handler.json && written.endsWith("{")) {
			state.sep = "";
		}
	}

	return state;
}

function go2jsSlogWrite(state, text) {
	state.buf.push(text);
}

function go2jsSlogOpenGroups(state) {
	for (const name of state.h.groups.slice(state.h.nOpenGroups)) {
		go2jsSlogOpenGroup(state, name);
	}
}

function go2jsSlogOpenGroup(state, name) {
	if (state.h.json) {
		go2jsSlogAppendKey(state, name);
		go2jsSlogWrite(state, "{");
		state.sep = "";
	} else {
		state.prefix.push(name, ".");
	}

	if (state.groups !== null) {
		state.groups.push(name);
	}
}

function go2jsSlogCloseGroup(state, name) {
	if (state.h.json) {
		go2jsSlogWrite(state, "}");
	} else {
		state.prefix.splice(state.prefix.length - 2, 2);
	}

	state.sep = state.h.json ? "," : " ";

	if (state.groups !== null) {
		state.groups.pop();
	}

	void name;
}

// go2jsSlogAppendAttrs writes a list of attributes and says whether anything of
// them was written, since a list of nothing but groups nobody filled writes
// nothing at all.
function go2jsSlogAppendAttrs(state, attrs) {
	let written = false;

	for (const attr of attrs) {
		if (go2jsSlogAppendAttr(state, attr)) {
			written = true;
		}
	}

	return written;
}

// go2jsSlogAppendAttr writes one attribute, after asking the program whether it
// wants it written another way, and says whether anything was written for it.
function go2jsSlogAppendAttr(state, attr) {
	let held = new go2jsSlogAttr(attr.Key, attr.Value.Resolve());
	const replace = go2jsSlogReplaceOf(state.h);

	if (replace !== null && held.Value.kind !== go2jsSlogKind.Group) {
		held = go2jsSlogAsAttr(go2jsSlogInvoke(replace, null,
			[state.groups === null ? null : state.groups.slice(), held]));
		held = new go2jsSlogAttr(held.Key, held.Value.Resolve());
	}

	if (go2jsSlogAttrEmpty(held)) {
		return false;
	}

	// The place a record was made is a value of its own, written as the place it
	// was made rather than as the fields it is made of, unless the record is
	// being written as JSON, where those fields are the shape of the writing.
	if (held.Value.kind === go2jsSlogKind.Any) {
		const source = go2jsSlogSourceParts(held.Value.any);

		if (source !== null) {
			held = new go2jsSlogAttr(held.Key, state.h.json
				? go2jsSlogGroupValue([
					go2jsSlogString("function", source.Function),
					go2jsSlogString("file", source.File),
					go2jsSlogInt("line", source.Line),
				])
				: go2jsSlogStringValue(source.File + ":" + source.Line));
		}
	}

	if (held.Value.kind === go2jsSlogKind.Group) {
		const attrs = held.Value.Group();

		// A group with nothing in it is not written, and a group that turns out
		// to hold nothing once the program has had its say about it is taken back
		// off the end of what was written.
		if (attrs.length > 0) {
			const mark = state.buf.length;

			if (held.Key !== "") {
				go2jsSlogOpenGroup(state, held.Key);
			}

			if (!go2jsSlogAppendAttrs(state, attrs)) {
				state.buf.length = mark;
				return false;
			}

			if (held.Key !== "") {
				go2jsSlogCloseGroup(state, held.Key);
			}
		}
	} else {
		go2jsSlogAppendKey(state, held.Key);
		go2jsSlogAppendValue(state, held.Value);
	}

	return true;
}

// go2jsSlogAppendNonBuiltIns writes the attributes a program gave before the
// record was made, which were written once and for all, and then those the record
// carries itself, which are inside the groups the handler was given.
function go2jsSlogAppendNonBuiltIns(state, record) {
	const handler = state.h;

	if (handler.preformatted !== "") {
		go2jsSlogWrite(state, state.sep);
		go2jsSlogWrite(state, handler.preformatted);
		state.sep = handler.json ? "," : " ";

		if (handler.json && handler.preformatted.endsWith("{")) {
			state.sep = "";
		}
	}

	// A record with no attributes writes no group at all, since a group nobody
	// put anything in is not part of what the record said.
	let open = handler.nOpenGroups;

	if (record.NumAttrs() > 0) {
		state.prefix.push(handler.groupPrefix);

		const mark = state.buf.length;

		go2jsSlogOpenGroups(state);
		open = handler.groups.length;

		let empty = true;

		for (const attr of record.__go2js_attrs) {
			if (go2jsSlogAppendAttr(state, attr)) {
				empty = false;
			}
		}

		if (empty) {
			state.buf.length = mark;
			open = handler.nOpenGroups;
		}
	}

	if (handler.json) {
		for (let count = 0; count < open; count++) {
			go2jsSlogWrite(state, "}");
		}

		go2jsSlogWrite(state, "}");
	}
}

function go2jsSlogAppendKey(state, key) {
	go2jsSlogWrite(state, state.sep);
	go2jsSlogAppendString(state, state.prefix.join("") + key);

	if (state.h.json) {
		go2jsSlogWrite(state, ":");
	} else {
		go2jsSlogWrite(state, "=");
	}

	state.sep = state.h.json ? "," : " ";
}

function go2jsSlogAppendString(state, text) {
	go2jsSlogWrite(state, state.h.json
		? "\"" + go2jsSlogEscapeJSON(text) + "\""
		: go2jsSlogQuoted(text));
}

// go2jsSlogAppendValue writes the value of an attribute, which is written the
// way the handler was built to write values.
function go2jsSlogAppendValue(state, value) {
	try {
		if (state.h.json) {
			go2jsSlogAppendJSONValue(state, value);
		} else {
			go2jsSlogAppendTextValue(state, value);
		}
	} catch (err) {
		go2jsSlogAppendString(state, "!ERROR:" + go2jsSlogErrorText(err));
	}
}

// go2jsSlogAppendTextValue writes a value the way Go writes one into a line of
// key=value pairs: a string and a moment are written as themselves, a value of
// no kind of its own is written the way the words would write it, and a value
// that has a kind is written as that kind writes it.
// go2jsSlogIsByteSlice is whether a value is a run of bytes, which a handler
// writes as the words it stands for rather than as the numbers behind them.
function go2jsSlogIsByteSlice(value) {
	const name = go2jsGoTypeNameRaw(value);

	return (name === "[]byte" || name === "[]uint8" || name === "[]uint8 " ||
		name === "[]byte ") && go2jsSlogBytesOf(go2jsUnwrap(go2jsUntyped(value))) !== null;
}

// go2jsSlogBytesOf is the numbers a run of bytes holds, whether the runtime
// keeps them as a list or as the text they were written from.
function go2jsSlogBytesOf(value) {
	if (value === null || value === undefined) {
		return null;
	}

	if (Array.isArray(value)) {
		return value;
	}

	if (typeof value === "string") {
		return go2jsStringToBytes(value);
	}

	if (typeof value === "object") {
		if (Array.isArray(value.data)) {
			return value.data.slice(value.offset, value.offset + value.length);
		}

		if (typeof value.valueOf === "function") {
			const held = value.valueOf();

			if (Array.isArray(held)) {
				return held;
			}
		}
	}

	return null;
}

function go2jsSlogAppendTextValue(state, value) {
	switch (value.kind) {
	case go2jsSlogKind.String:
		go2jsSlogAppendString(state, value.str);
		return;
	case go2jsSlogKind.Time:
		go2jsSlogAppendTime(state, value.any);
		return;
	case go2jsSlogKind.Any: {
		const any = go2jsUnwrap(go2jsUntyped(value.any));

		// An attribute held as a value is written as the pair it stands for,
		// since that is what it says of itself when it is shown.
		if (any instanceof go2jsSlogAttr) {
			go2jsSlogAppendString(state, go2jsSlogInvoke(any.String, any, []));
			return;
		}

		// A run of bytes is written as the words it stands for, since that is
		// what a run of bytes is for, and writing out the numbers behind them
		// says nothing that a reader wanted.
		if (go2jsSlogIsByteSlice(value.any)) {
			go2jsSlogWrite(state, go2jsSprintf("%q", go2jsBytesToString(go2jsSlogBytesOf(any))));
			return;
		}

		const written = any !== null && any !== undefined &&
			typeof any === "object" && typeof any.MarshalText === "function"
			? go2jsSlogInvoke(any.MarshalText, any, [])
			: null;

		if (written !== null) {
			go2jsSlogAppendString(state, go2jsRawText(written));
			return;
		}

		go2jsSlogAppendString(state, go2jsSprintf("%+v", any));
		return;
	}
	}

	go2jsSlogWrite(state, go2jsSlogValueText(value));
}

function go2jsSlogAppendJSONValue(state, value) {
	switch (value.kind) {
	case go2jsSlogKind.String:
		go2jsSlogAppendString(state, value.str);
		return;
	case go2jsSlogKind.Int64:
	case go2jsSlogKind.Uint64:
		go2jsSlogWrite(state, String(value.num));
		return;
	case go2jsSlogKind.Duration:
		go2jsSlogWrite(state, String(value.num));
		return;
	case go2jsSlogKind.Float64:
		go2jsSlogWrite(state, go2jsSlogJSONFloat(Number(value.num)));
		return;
	case go2jsSlogKind.Bool:
		go2jsSlogWrite(state, value.num === true || value.num === 1 ? "true" : "false");
		return;
	case go2jsSlogKind.Time:
		go2jsSlogWrite(state, "\"" + go2jsSlogTimeText(value.any, "2006-01-02T15:04:05.999999999Z07:00") + "\"");
		return;
	case go2jsSlogKind.Any: {
		const any = value.any;

		// An error an attribute holds is written as the text it says, since that
		// is what the words of a fault are, unless it says how to write itself as
		// something else first.
		if (go2jsIsErrorValue(any) && typeof any.MarshalJSON !== "function") {
			go2jsSlogAppendString(state, go2jsRawText(go2jsSlogErrorText(any)));
			return;
		}

		if (any !== null && any !== undefined && typeof any === "object" &&
			typeof any.MarshalJSON === "function") {
			go2jsSlogWrite(state, go2jsRawText(go2jsSlogInvoke(any.MarshalJSON, any, [])));
			return;
		}

		go2jsSlogWrite(state, go2jsSlogJSONText(any));
		return;
	}
	}

	go2jsSlogAppendString(state, "!ERROR:bad kind: " + go2jsSlogKindName(value.kind));
}

// go2jsSlogAppendTime writes a moment the way RFC 3339 writes it with the
// thousandths of a second counted out in full.
function go2jsSlogAppendTime(state, time) {
	go2jsSlogWrite(state, go2jsSlogHandlerOf(state).json
		? "\"" + go2jsSlogTimeText(time, "2006-01-02T15:04:05.999999999Z07:00") + "\""
		: go2jsSlogTimeText(time, "2006-01-02T15:04:05.000Z07:00"));
}

function go2jsSlogHandlerOf(state) {
	return state.h;
}

// go2jsSlogJSONFloat is a length written the way encoding/json writes one,
// which is in full until the number is so large or so small that an exponent is
// shorter than the digits it stands for.
function go2jsSlogJSONFloat(value) {
	if (!isFinite(value)) {
		return go2jsSlogQuoted("!ERROR:json: unsupported value: " +
			(Number.isNaN(value) ? "NaN" : (value > 0 ? "+Inf" : "-Inf")));
	}

	const held = Math.abs(value);

	if (held === 0) {
		return Object.is(value, -0) ? "-0" : "0";
	}

	if (held < 1e-6 || held >= 1e21) {
		const written = value.toExponential().split("e");
		const sign = written[1][0] === "-" ? "-" : "+";
		const step = written[1].slice(1).replace(/^0+/, "");

		return written[0] + "e" + sign + step;
	}

	return String(value);
}

// go2jsSlogJSONText is a value of no kind of its own written as JSON writes it,
// which is what a handler told to write JSON writes for a value it knows nothing
// about.
function go2jsSlogJSONText(value) {
	if (value === null || value === undefined) {
		return "null";
	}

	switch (typeof value) {
	case "string":
		return "\"" + go2jsSlogEscapeJSON(value) + "\"";
	case "boolean":
		return value ? "true" : "false";
	case "bigint":
		return String(value);
	case "number":
		return go2jsSlogJSONFloat(value);
	case "function":
		return go2jsSlogQuoted("!ERROR:json: unsupported type: function");
	case "symbol":
		return go2jsSlogQuoted("!ERROR:json: unsupported type: symbol");
	}

	if (value.value instanceof Date) {
		return "\"" + go2jsSlogTimeText(value) + "\"";
	}

	if (Array.isArray(value)) {
		return "[" + value.map(item => go2jsSlogJSONText(item)).join(",") + "]";
	}

	if (value instanceof go2jsNativeMap) {
		const parts = [];

		for (const [key, item] of value.entries()) {
			parts.push("\"" + go2jsSlogEscapeJSON(go2jsStringify(key)) + "\":" + go2jsSlogJSONText(item));
		}

		return "{" + parts.join(",") + "}";
	}

	if (value instanceof go2jsSlogAttr) {
		return "\"" + go2jsSlogEscapeJSON(value.Key) + "\":" + go2jsSlogJSONText(value.Value.Any());
	}

	const parts = [];

	for (const key of Object.keys(value)) {
		if (key.startsWith("__go2js")) {
			continue;
		}

		parts.push("\"" + go2jsSlogEscapeJSON(key) + "\":" + go2jsSlogJSONText(value[key]));
	}

	return "{" + parts.join(",") + "}";
}

// go2jsSlogQuoted is a piece of a line written as it is when nothing in it needs
// quoting, and quoted the way strconv quotes when something does.
function go2jsSlogQuoted(text) {
	return go2jsSlogNeedsQuoting(text) ? go2jsSprintf("%q", text) : text;
}

// go2jsSlogNeedsQuoting is whether a piece of a line carries something that would
// be read as more or less than the piece itself, which is what a space, a mark of
// equality, a mark of quotation or anything no page shows would.
function go2jsSlogNeedsQuoting(text) {
	if (text === "") {
		return true;
	}

	return go2jsSlogUnprintable.test(text);
}

const go2jsSlogUnprintable = /[\x00-\x1f\x7f"'\\= \p{Cc}\p{Cf}\p{Cs}\p{Co}\p{Cn}\p{Zl}\p{Zp}]/u;

// go2jsSlogEscapeJSON is a string written as JSON writes one, where the marks
// that end a string and the one that escapes the character after it are the only
// ones written in full, and the two characters that end a line inside a string
// are written out even though JSON does not ask for it, since JSON is often
// read as JavaScript.
function go2jsSlogEscapeJSON(text) {
	const hex = "0123456789abcdef";
	let out = "";
	let start = 0;

	for (let index = 0; index < text.length; index++) {
		const held = text[index];
		const code = text.charCodeAt(index);

		if (code < 0x80) {
			// A character a page shows is written as it is, save for the two that
			// end a string and the one that stands for the character after it.
			if (code >= 0x20 && code <= 0x7e && held !== "\"" && held !== "\\") {
				continue;
			}

			out += text.slice(start, index);

			switch (held) {
			case "\\":
			case "\"":
				out += "\\" + held;
				break;
			case "\n":
				out += "\\n";
				break;
			case "\r":
				out += "\\r";
				break;
			case "\t":
				out += "\\t";
				break;
			default:
				out += "\\u00" + hex[code >> 4] + hex[code & 0xf];
			}

			start = index + 1;
			continue;
		}

		if (code >= 0xd800 && code <= 0xdbff && index + 1 < text.length) {
			const next = text.charCodeAt(index + 1);

			if (next >= 0xdc00 && next <= 0xdfff) {
				continue;
			}
		}

		if (code === 0x2028 || code === 0x2029) {
			out += text.slice(start, index) + "\\u202" + hex[code & 0xf];
			start = index + 1;
		}
	}

	return start < text.length ? out + text.slice(start) : out;
}

// go2jsSlogWriteLine is a line handed to the writer a program gave, which is one
// call to Write however many attributes the record carried.
function go2jsSlogWriteLine(handler, text) {
	// The line of a record written through no handler of the program's is left
	// to the logger of the log package, since that is what writes it there.
	if (handler.__go2js_slog_default === true) {
		go2jsLogOutput(go2jsLogDecorate(text));

		return null;
	}

	const target = go2jsUnwrap(handler.writer);

	if (target === null || target === undefined) {
		return go2jsSlogErrorValue("slog: writer is nil");
	}

	if (typeof target.Write !== "function") {
		return go2jsSlogErrorValue("slog: writer does not implement io.Writer");
	}

	const written = go2jsCallNow(target.Write, target, [go2jsStringToBytes(text)]);

	if (Array.isArray(written)) {
		return written[1] === null || written[1] === undefined ? null : written[1];
	}

	return written === undefined ? null : written;
}

// go2jsSlogTimeText is a moment written as a layout writes it, where a layout
// named for a moment is how Go writes one.
function go2jsSlogTimeText(time, layout) {
	const wanted = layout === undefined || layout === null ? "2006-01-02 15:04:05.999999999 -0700 MST" : layout;

	if (time === null || time === undefined) {
		return "";
	}

	if (time.value instanceof Date) {
		return go2jsTimeFormat(time.value, wanted, time);
	}

	if (typeof time.Format === "function") {
		return go2jsRawText(time.Format(wanted));
	}

	return go2jsRawText(time);
}

function go2jsSlogTimeIsZero(time) {
	if (time === null || time === undefined) {
		return true;
	}

	if (typeof time.IsZero === "function") {
		return time.IsZero() === true;
	}

	return time.value instanceof Date && time.__go2js_time_zero === true;
}

// go2jsSlogSourceOf is the place a program stood when it asked for a record to
// be written, read back into the parts Go keeps a place in.
function go2jsSlogSourceOf(text) {
	const held = String(text === null || text === undefined ? "" : text);

	if (held === "") {
		return null;
	}

	const at = held.lastIndexOf(":");

	if (at < 0) {
		return {Function: "", File: held, Line: 0};
	}

	return {Function: "", File: held.slice(0, at), Line: Number(held.slice(at + 1))};
}

// go2jsSlogSourceParts reads a place out of a value, whether the program made one
// or the runtime read one back out of the text it was given.
function go2jsSlogSourceParts(value) {
	if (value === null || value === undefined || typeof value !== "object") {
		return null;
	}

	if (typeof value.File === "string" && typeof value.Line === "number") {
		return {Function: String(value.Function === undefined ? "" : value.Function), File: value.File, Line: value.Line};
	}

	if (typeof value.__go2js_file === "string" && typeof value.__go2js_line === "number") {
		return {Function: "", File: value.__go2js_file, Line: value.__go2js_line};
	}

	return null;
}

function go2jsSlogCountEmptyGroups(attrs) {
	let count = 0;

	for (const attr of attrs) {
		if (go2jsSlogEmptyGroup(attr.Value)) {
			count++;
		}
	}

	return count;
}

// go2jsSlogInvoke asks a value of the program to do something and reads what it
// answered, since a method the program declared may be one that waits and a
// method of two answers comes back as a pair of them.
function go2jsSlogInvoke(target, self, args) {
	if (target === null || target === undefined || typeof target !== "function") {
		return null;
	}

	const answered = go2jsCallNow(target, self, args === undefined ? [] : args);

	return Array.isArray(answered) ? answered[0] : answered;
}

// go2jsSlogError is a fault the package stands for on its own account, which is
// what a record says went wrong when the program said so with words.
function go2jsSlogErrorValue(text) {
	return go2jsErrorsNew(String(text));
}

function go2jsSlogErrorText(err) {
	return go2jsErrorMessage(err);
}

// go2jsSlogLogger writes records to the handler it was given, which is the one
// place a logger decides what a record looks like rather than what it says.
function go2jsSlogLogger(handler) {
	return {
		handler: handler === undefined || handler === null ? go2jsSlogDefaultHandler() : handler,
		Handler: function() {
			return this.handler;
		},
		Enabled: function(ctx, level) {
			return go2jsSlogLoggerEnabled(this, ctx, level);
		},
		With: function(args, types) {
			return go2jsSlogLoggerWith(this, args, types);
		},
		WithGroup: function(name) {
			return go2jsSlogLoggerWithGroup(this, name);
		},
		Log: function(ctx, level, message, args, types, source) {
			return go2jsSlogWriteRecord(this, ctx, level, message,
				go2jsSlogArgsToAttrs(args === undefined ? [] : args, types), source);
		},
		LogAttrs: function(ctx, level, message, attrs, source) {
			return go2jsSlogWriteRecord(this, ctx, level, message, attrs, source);
		},
		Debug: function(message, args, types, source) {
			return this.Log(null, -4, message, args, types, source);
		},
		Info: function(message, args, types, source) {
			return this.Log(null, 0, message, args, types, source);
		},
		Warn: function(message, args, types, source) {
			return this.Log(null, 4, message, args, types, source);
		},
		Error: function(message, args, types, source) {
			return this.Log(null, 8, message, args, types, source);
		},
	};
}

// go2jsSlogLoggerEnabled is whether the handler behind a logger writes a record
// of this level, which is the handler's own answer rather than the logger's.
function go2jsSlogLoggerEnabled(logger, ctx, level) {
	const handler = go2jsUnwrap(logger.handler);

	if (handler === null || handler === undefined || typeof handler.Enabled !== "function") {
		return true;
	}

	return go2jsSlogInvoke(handler.Enabled, handler, [ctx, Number(level) | 0]) !== false;
}

// go2jsSlogLoggerWith is a logger that writes what this one writes and then
// writes these attributes as well, which the handler remembers rather than the
// logger, since they are the same on every record that follows.
function go2jsSlogLoggerWith(logger, ...rest) {
	let types = null;
	let args = [];
	if (rest.length > 0) {
		if (Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}
	const handler = go2jsUnwrap(logger.handler);
	const attrs = go2jsSlogArgsToAttrs(args, types);

	if (handler === null || handler === undefined || typeof handler.WithAttrs !== "function") {
		return new go2jsSlogLogger(handler);
	}

	return new go2jsSlogLogger(go2jsSlogInvoke(handler.WithAttrs, handler, [attrs]));
}

// go2jsSlogLoggerWithGroup is a logger that writes everything it is given under
// one more group name.
function go2jsSlogLoggerWithGroup(logger, name) {
	const handler = go2jsUnwrap(logger.handler);

	if (handler === null || handler === undefined || typeof handler.WithGroup !== "function") {
		return new go2jsSlogLogger(handler);
	}

	return new go2jsSlogLogger(go2jsSlogInvoke(handler.WithGroup, handler, [go2jsRawText(name)]));
}

// go2jsSlogWriteRecord writes one record, which is first asked of the handler
// whether it is one the handler writes at all, since a record of a level the
// handler drops is not written and never reaches the writer.
function go2jsSlogWriteRecord(logger, ctx, level, message, attrs, source) {
	const handler = go2jsUnwrap(logger.handler);

	if (handler === null || handler === undefined || typeof handler.Handle !== "function") {
		return;
	}

	// A level is a named number, so it is read through whatever it was wrapped
	// in on the way here.
	level = go2jsSlogLevelNumber(level);

	if (!go2jsSlogLoggerEnabled(logger, ctx, level)) {
		return;
	}

	const record = new go2jsSlogRecord(go2jsTimeNow(), level, message, 0);

	record.__go2js_source = source === undefined || source === null ? "" : String(source);
	record.AddAttrs(attrs === undefined || attrs === null ? [] : attrs);

	go2jsSlogInvoke(handler.Handle, handler, [ctx, record]);
}

function go2jsSlogLoggerError(logger, message, ...rest) {
	let types = null;
	let source = undefined;
	let args = [];
	if (rest.length > 0) {
		const last = rest[rest.length - 1];
		if (typeof last === "string" && last.indexOf(":") >= 0) {
			source = last;
			rest = rest.slice(0, rest.length - 1);
		}
		if (rest.length > 0 && Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}
	return go2jsSlogWriteRecord(logger, null, 8, message, go2jsSlogArgsToAttrs(args, types), source);
}

// go2jsSlogDefaultLogger is the logger a program logs through when it names no
// logger of its own, which writes through the logger of the log package and so
// is written where that one is written.
const go2jsSlogDefaultLogger = new go2jsSlogLogger(go2jsSlogDefaultHandler());

// go2jsSlogLogLoggerLevel is the lowest level the default logger writes, which
// the log package decides and a program may move.
let go2jsSlogLogLoggerLevel = 0;

// go2jsSlogDefaultHandler is the handler behind the default logger, which is not
// one of the two the package builds, since it leaves the writing of the line to
// the log package and only gathers what goes on it.
function go2jsSlogDefaultHandler() {
	const handler = {
		__go2js_slog_default: true,
		opts: {},
		preformatted: "",
		groupPrefix: "",
		groups: [],
		nOpenGroups: 0,
	};

	handler.Enabled = function(ctx, level) {
		void ctx;
		return (Number(level) | 0) >= go2jsSlogLogLoggerLevel;
	};

	handler.Handle = function(ctx, record) {
		void ctx;

		const state = go2jsSlogState(this, go2jsSlogLevelName(record.Level) + " " + record.Message);

		state.sep = " ";
		go2jsSlogAppendNonBuiltIns(state, record);

		return go2jsSlogWriteLine(this, state.buf.join("") + "\n");
	};

	handler.WithAttrs = function(attrs) {
		return go2jsSlogHandlerWithAttrs(this, attrs);
	};

	handler.WithGroup = function(name) {
		return go2jsSlogHandlerWithGroup(this, name);
	};

	return handler;
}

// go2jsSlogSet is where the default logger stands, which is a name rather than a
// constant so that a program which was given a logger of its own logs through
// that one from then on.
let go2jsSlogDefaultHeld = go2jsSlogDefaultLogger;

function go2jsSlogSet(logger) {
	go2jsSlogDefaultHeld = logger;
}

function go2jsSlogError(message, ...rest) {
	let types = null;
	let source = undefined;
	let args = [];
	if (rest.length > 0) {
		const last = rest[rest.length - 1];
		if (typeof last === "string" && last.indexOf(":") >= 0) {
			source = last;
			rest = rest.slice(0, rest.length - 1);
		}
		if (rest.length > 0 && Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}
	return go2jsSlogLoggerError(go2jsSlogDefaultHeld, message, ...args, types, source);
}

function go2jsSlogHandlerOptions() {}

go2jsRegisterTypeName(go2jsSlogHandlerOptions, "slog.HandlerOptions");

go2jsSlogHandlerOptions.prototype.Level = undefined;

go2jsSlogHandlerOptions.prototype.AddSource = false;

go2jsSlogHandlerOptions.prototype.ReplaceAttr = undefined;

// go2jsHTTPStatusText is the phrase net/http writes after a status code. The
// table is the one net/http keeps, so a response reads the way it would in Go.
const go2jsHTTPStatusTexts = {
	200: "OK",
	100: "Continue",
	101: "Switching Protocols",
	102: "Processing",
	103: "Early Hints",
	200: "OK",
	201: "Created",
	202: "Accepted",
	203: "Non-Authoritative Information",
	204: "No Content",
	205: "Reset Content",
	206: "Partial Content",
	207: "Multi-Status",
	208: "Already Reported",
	226: "IM Used",
	300: "Multiple Choices",
	301: "Moved Permanently",
	302: "Found",
	303: "See Other",
	304: "Not Modified",
	305: "Use Proxy",
	307: "Temporary Redirect",
	308: "Permanent Redirect",
	400: "Bad Request",
	401: "Unauthorized",
	402: "Payment Required",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	406: "Not Acceptable",
	407: "Proxy Authentication Required",
	408: "Request Timeout",
	409: "Conflict",
	410: "Gone",
	411: "Length Required",
	412: "Precondition Failed",
	413: "Request Entity Too Large",
	414: "Request URI Too Long",
	415: "Unsupported Media Type",
	416: "Requested Range Not Satisfiable",
	417: "Expectation Failed",
	418: "I'm a teapot",
	421: "Misdirected Request",
	422: "Unprocessable Entity",
	423: "Locked",
	424: "Failed Dependency",
	425: "Too Early",
	426: "Upgrade Required",
	428: "Precondition Required",
	429: "Too Many Requests",
	431: "Request Header Fields Too Large",
	451: "Unavailable For Legal Reasons",
	500: "Internal Server Error",
	501: "Not Implemented",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Timeout",
	505: "HTTP Version Not Supported",
	506: "Variant Also Negotiates",
	507: "Insufficient Storage",
	508: "Loop Detected",
	510: "Not Extended",
	511: "Network Authentication Required",
};

function go2jsHTTPStatusText(code) {
	const text = go2jsHTTPStatusTexts[code];

	return text === undefined ? "" : text;
}

// An errno the host reports is said the way Go says it, rather than in the
// wording of the host, which names the same condition a different way.
const go2jsErrnoMessages = new Map([
	["ENOENT", "no such file or directory"],
	["EEXIST", "file exists"],
	["EACCES", "permission denied"],
	["EPERM", "operation not permitted"],
	["EISDIR", "is a directory"],
	["ENOTDIR", "not a directory"],
	["ENOTEMPTY", "directory not empty"],
	["EMFILE", "too many open files"],
	["ELOOP", "too many levels of symbolic links"],
	["EINVAL", "invalid argument"],
	["EBADF", "bad file descriptor"],
	["ENOSPC", "no space left on device"],
	["EXDEV", "invalid cross-device link"],
	["ENOTTY", "inappropriate ioctl for device"]
]);

// An error the host reports is written as an error of a path: what was being
// done, which path, and the condition that ended it, which is what a Go program
// reading one is given. An error of a link is told of two names rather than one,
// and carries the old and the new rather than a path.
function go2jsOSHostError(err, syscall, path, old) {
	if (err === null || err === undefined) {
		return null;
	}

	const named = go2jsErrnoMessages.get(err.code);

	if (named === undefined) {
		return err;
	}

	if (old !== undefined && old !== null) {
		return go2jsOSLinkError(named, err.code, syscall, old, path);
	}

	return go2jsNameError(go2jsOSError(named, syscall, path), "*fs.PathError");
}

// go2jsOSLinkError is what a link that was refused or could not be made answers
// with: what was being done, the name that was pointed at, the name that was made
// to point at it, and the condition that ended it.
function go2jsOSLinkError(message, code, op, old, path) {
	const error = new Error(String(op) + " " + String(old) + " " + String(path) + ": " + String(message));

	error.__go2js_errno = true;
	error.code = code;
	error.op = op;
	error.old = old;
	error.path = path;
	error.syscall = op;

	const errno = go2jsErrnoSentinel(code);

	if (errno !== null) {
		error.cause = errno;
	}

	return go2jsNameError(error, "*fs.LinkError");
}

// An error from a file is told by what was being done to which path, which is
// how one error is told from another that reads the same.
function go2jsOSError(message, syscall, path, target) {
	const text = String(message);
	const notExist = /no such file or directory|not exist|ENOENT/i.test(text);
	const error = new Error(String(syscall) + " " + String(path) + ": " + text);

	error.__go2js_errno = true;
	error.code = notExist ? "ENOENT" : text;
	error.syscall = syscall;
	error.path = String(path);
	error.errno = notExist ? 2 : 0;

	// An error of a path carries the errno that ended it, which is what a chain
	// unwraps to and what tells errors.Is that the condition it stands for is
	// the one a sentinel of the os package names.
	const errno = go2jsErrnoSentinel(error.code);

	if (errno !== null) {
		error.cause = errno;
	}

	return error;
}

// go2jsOSFileInfo is what a stat answers with: a set of methods rather than the
// fields the host happens to keep the same information in, so that a program
// asks a file what it wants to know the way it asks one in Go.
function go2jsOSFileInfo(path, stats) {
	const text = String(path);
	const info = {
		__go2js_type: "*os.fileStat",
		Name: function () {
			return text.slice(text.lastIndexOf("/") + 1);
		},
		Size: function () {
			return Number(stats.size);
		},
		IsDir: function () {
			return typeof stats.isDirectory === "function" ? stats.isDirectory() : false;
		},
		Mode: function () {
			// What may be done with a file is a mode of marks rather than a
			// number of bits, since that is what a program asks about, and the
			// marks of the host are in the places Go keeps its own.
			return go2jsFileMode(go2jsFileModeFromHost(stats.mode === undefined ? 0 : stats.mode));
		},
		ModTime: function () {
			const when = stats.mtime instanceof Date ? stats.mtime : new Date(Number(stats.mtime));

			return go2jsTimeValue(isNaN(when.getTime()) ? new Date(0) : when);
		}
	};

	return info;
}

// go2jsGoexitSignal is what ends a goroutine that was told to stop. It is not a
// panic: it is thrown so that the deferred calls of the frames it leaves behind
// run on the way out, and the goroutine runner tells the two apart.
const go2jsGoexitSignal = {__go2js_goexit: true};

// go2jsIsGoexit reports whether a thrown value is a goroutine being told to
// end rather than a panic, because the two are unwound the same way and only
// one of them is a failure.
function go2jsIsGoexit(thrown) {
	return thrown !== null && thrown !== undefined && thrown.__go2js_goexit === true;
}

// A header is a set of names, each standing for one or more values. The names
// are written the way net/http writes them, with the first letter of each word
// of the name in capitals, so that a name asked for in any letter is found.
function go2jsHTTPHeaderKey(name) {
	const words = String(name).split("-");

	for (let index = 0; index < words.length; index++) {
		if (words[index] === "") {
			continue;
		}

		words[index] = words[index].charAt(0).toUpperCase() + words[index].slice(1);
	}

	return words.join("-");
}

function go2jsHTTPHeader() {
	const values = {};

	// The values are also held as plain properties under the names Go writes,
	// so that a program reaching into a header for one of them finds it there.
	const entry = function(key) {
		const name = go2jsHTTPHeaderKey(key);

		if (values[name] === undefined) {
			values[name] = [];
		}

		return values[name];
	};

	const header = {
		Set: function(key, value) {
			values[go2jsHTTPHeaderKey(key)] = [String(value)];
		},
		Add: function(key, value) {
			entry(key).push(String(value));
		},
		Get: function(key) {
			const list = values[go2jsHTTPHeaderKey(key)];

			return Array.isArray(list) && list.length > 0 ? list[0] : "";
		},
		Values: function(key) {
			const list = values[go2jsHTTPHeaderKey(key)];

			return Array.isArray(list) ? list.slice() : null;
		},
		Del: function(key) {
			delete values[go2jsHTTPHeaderKey(key)];
		},
		Clone: function() {
			const copy = go2jsHTTPHeader();

			for (const name of Object.keys(values)) {
				copy.Add(name, values[name].slice());
			}

			return copy;
		},
		Write: function(writer) {
			let written = 0;

			for (const name of Object.keys(values).sort()) {
				for (const value of values[name]) {
					const line = name + ": " + value + "\r\n";

					go2jsHTTPWriteString(writer, line);
					written += line.length;
				}
			}

			return [written, null];
		},
		__go2js_values: function() {
			return values;
		},
		__go2js_own: function() {
			return go2jsHTTPHeaderFromValues(values);
		}
	};

	return header;
}

// go2jsHTTPHeaderFromValues holds a set of header values as plain properties, so
// that a program reading a header by name finds the list of values waiting there
// under the name Go writes it in.
function go2jsHTTPHeaderFromValues(values) {
	const header = go2jsHTTPHeader();

	for (const name of Object.keys(values)) {
		header[name] = values[name].slice();
	}

	return header;
}

// go2jsHTTPHeaderEntries lists the names a header carries along with the values
// under each of them, which is what a header written out or cloned is made of.
function go2jsHTTPHeaderEntries(header) {
	if (header === null || header === undefined) {
		return {};
	}

	if (typeof header.__go2js_values === "function") {
		return header.__go2js_values();
	}

	// a header that reached here as a plain object of names to values is read
	// the way it stands
	const values = {};

	for (const name of Object.keys(header)) {
		if (name.startsWith("__go2js") || typeof header[name] === "function") {
			continue;
		}

		values[name] = Array.isArray(header[name]) ? header[name] : [String(header[name])];
	}

	return values;
}

function go2jsHTTPHeaderGet(header, key) {
	const values = go2jsHTTPHeaderEntries(header);
	const list = values[go2jsHTTPHeaderKey(key)];

	return Array.isArray(list) && list.length > 0 ? list[0] : "";
}

function go2jsHTTPHeaderSet(header, key, value) {
	header.Set(key, value);
}

// go2jsHTTPNewError is the fault a request that never reached anything gives
// back. It is named the way net/http names it, carrying the operation that
// failed and the fault underneath it.
function go2jsHTTPNewError(op, url, cause) {
	const err = new Error(op + " " + String(url) + ": " + String(cause));

	err.__go2js_http_error = true;
	err.__go2js_type = "*url.Error";
	err.name = "Error";
	err.Op = String(op);
	err.URL = String(url);
	err.Err = cause === undefined || cause === null ? null : cause;

	return err;
}

// go2jsHTTPBody is the reader a response body is read through. It hands out the
// bytes as they are asked for and reports the end of them when there are none
// left, which is what a reader over a fixed run of bytes does.
function go2jsHTTPBody(payload) {
	const bytes = Buffer.isBuffer(payload) ? payload : Buffer.from(payload === undefined || payload === null ? new Uint8Array(0) : payload);
	let at = 0;
	let closed = false;

	const body = {
		Read: function(buffer) {
			if (closed || at >= bytes.length) {
				return [0, go2jsIOEOF()];
			}

			const wanted = buffer === undefined || buffer === null ? bytes.length - at : buffer.length;
			const count = Math.min(wanted, bytes.length - at);

			for (let index = 0; index < count; index++) {
				buffer[index] = bytes[at + index];
			}

			at += count;

			return [count, null];
		},
		Close: function() {
			closed = true;
			at = bytes.length;

			return null;
		},
		__go2js_bytes: function() {
			return bytes.slice(at);
		},
		__go2js_length: function() {
			return Math.max(0, bytes.length - at);
		}
	};

	return body;
}

// go2jsHTTPBodyBytes reads all that is left of a body, which is what a program
// hands to io.ReadAll when it would rather have the bytes outright.
function go2jsHTTPBodyBytes(reader) {
	const source = go2jsUnwrap(reader);

	if (source === null || source === undefined) {
		return new Uint8Array(0);
	}

	if (typeof source === "string") {
		return go2jsStringToBytes(source);
	}

	if (source instanceof Uint8Array) {
		return source;
	}

	if (source.__go2js_no_body === true) {
		return new Uint8Array(0);
	}

	if (typeof source.__go2js_bytes === "function") {
		return source.__go2js_bytes();
	}

	if (typeof source.Bytes === "function") {
		return source.Bytes();
	}

	if (typeof source.Read !== "function") {
		return new Uint8Array(0);
	}

	const chunks = [];
	const buffer = new Array(8192).fill(0);

	for (;;) {
		const result = go2jsCallNow(source.Read, source, [buffer]);
		const read = Number(Array.isArray(result) ? result[0] : result) || 0;

		if (read > 0) {
			chunks.push(buffer.slice(0, read));
		}

		if (read <= 0) {
			break;
		}
	}

	let total = 0;

	for (const chunk of chunks) {
		total += chunk.length;
	}

	const out = new Uint8Array(total);
	let at = 0;

	for (const chunk of chunks) {
		out.set(chunk, at);
		at += chunk.length;
	}

	return out;
}

// go2jsHTTPNoBody is the reader a request with nothing in its body is given. It
// reports no content and never ends, so it is not a reader over anything: every
// read asks again and the answer is always no bytes.
function go2jsHTTPNoBody() {
	return {
		Read: function() {
			return [0, go2jsIOEOF()];
		},
		Close: function() {
			return null;
		},
		__go2js_no_body: true
	};
}

// go2jsHTTPRequest holds a request as it is written down before it is carried
// out, and as it arrives after it has been carried out.
function go2jsHTTPRequest(method, url, body) {
	return {
		Method: String(method),
		URL: url,
		Proto: "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header: go2jsHTTPHeader(),
		Body: body === undefined ? null : body,
		ContentLength: 0,
		Host: "",
		Form: go2jsURLValues(),
		PostForm: go2jsURLValues(),
		MultipartForm: null,
		Trailer: go2jsHTTPHeader(),
		RemoteAddr: "",
		RequestURI: "",
		TransferEncoding: null,
		Close: false,
		TLS: null,
		Response: null,
		ctx: null
	};
}

// go2jsHTTPNewRequest builds a request out of a method, a url and a body. The
// url is parsed here rather than when the request is carried out, so that a url
// that is not a url is said so before anything is attempted.
function go2jsHTTPNewRequest(method, rawURL, body) {
	const text = String(rawURL === undefined || rawURL === null ? "" : rawURL);

	if (!/^[A-Za-z][A-Za-z0-9+.-]*:/.test(text)) {
		return [null, go2jsHTTPNewError("parse", text === "" ? "" : text, go2jsHTTPCauseText("parse " + (text === "" ? "" : text) + ": empty url"))];
	}

	const [parsed, failure] = go2jsURLParse(text);

	if (failure !== null) {
		return [null, failure];
	}

	const request = go2jsHTTPRequest(method, parsed, body);

	if (body !== null && body !== undefined) {
		const bytes = go2jsHTTPBodyBytes(body);

		request.ContentLength = bytes.length;
		request.__go2js_body_bytes = bytes;
	} else {
		request.Body = go2jsHTTPNoBody();
		request.__go2js_body_bytes = new Uint8Array(0);
	}

	request.Host = parsed.Host;

	if (text.startsWith("http://") || text.startsWith("https://")) {
		request.RequestURI = parsed.RequestURI === undefined ? (parsed.Path === "" ? "/" : parsed.Path) + (parsed.RawQuery === "" ? "" : "?" + parsed.RawQuery) : parsed.RequestURI;
	} else {
		// a request written for a server rather than for the network carries
		// only the part of the url that names what is being asked for
		const cut = text.indexOf("://");
		const afterScheme = cut === -1 ? text : text.slice(cut + 3);
		const slash = afterScheme.indexOf("/");

		request.RequestURI = slash === -1 ? "/" : afterScheme.slice(slash);
	}

	return [request, null];
}

function go2jsHTTPCauseText(text) {
	const err = new Error(String(text));

	err.__go2js_type = "error";
	err.name = "Error";

	return err;
}

// go2jsHTTPWriteString hands a run of text to whatever a handler is writing to,
// which is a response writer of the server or a writer of the program.
function go2jsHTTPWriteString(writer, text) {
	if (writer === null || writer === undefined) {
		return 0;
	}

	if (typeof writer.__go2js_write === "function") {
		return writer.__go2js_write(go2jsStringToBytes(text));
	}

	if (typeof writer.write === "function") {
		writer.write(text);

		return text.length;
	}

	if (typeof writer.Write === "function") {
		go2jsCallMethod(writer, "Write", go2jsStringToBytes(text));

		return text.length;
	}

	return 0;
}

function go2jsHTTPWriteBytes(writer, bytes) {
	if (writer === null || writer === undefined) {
		return 0;
	}

	if (typeof writer.__go2js_write === "function") {
		return writer.__go2js_write(bytes);
	}

	if (typeof writer.Write === "function") {
		go2jsCallMethod(writer, "Write", bytes);

		return bytes.length;
	}

	return 0;
}

// go2jsHTTPResponseWriter is what a handler writes its answer through. What a
// handler writes is held until the answer is done, since a program that reads
// what a handler said reads it back through a recorder.
function go2jsHTTPResponseWriter() {
	const writer = {
		Code: 200,
		__go2js_header: go2jsHTTPHeader(),
		wrote: false,
		parts: [],
		finished: false,
		Request: null
	};

	// The headers of an answer are asked for rather than read, since that is
	// what net/http says of them, and a handler written against a writer whose
	// headers were a field would be told otherwise.
	writer.Header = function() {
		return writer.__go2js_header;
	};

	writer.__go2js_write = function(bytes) {
		writer.WriteHeader(200);

		const text = Buffer.isBuffer(bytes) ? bytes.toString("utf8") : go2jsBytesToString(bytes);

		writer.parts.push(text);

		return Array.isArray(bytes) ? bytes.length : (bytes === null || bytes === undefined ? 0 : bytes.length);
	};

	writer.WriteHeader = function(code) {
		// Only the first code counts, the way it does in net/http: a handler
		// that names another one has already had its answer sent.
		if (writer.wrote) {
			return;
		}

		writer.wrote = true;
		writer.Code = Number(code);
	};

	writer.Write = function(bytes) {
		writer.WriteHeader(200);

		const length = bytes === null || bytes === undefined ? 0 : (Array.isArray(bytes) ? bytes.length : bytes.length);

		writer.parts.push(Buffer.isBuffer(bytes) ? bytes.toString("utf8") : go2jsBytesToString(bytes));

		return [length, null];
	};

	writer.WriteString = function(text) {
		writer.WriteHeader(200);
		writer.parts.push(String(text));

		return [String(text).length, null];
	};

	writer.Flush = function() {};

	// go2jsHTTPServeHTTP carries a request to the handler it is for, and then
	// turns what the handler wrote into the answer that goes back.
	writer.go2js_serve = function(handler, request) {
		writer.Request = request;

		const given = go2jsHTTPHandlerUnder(handler);

		let answered = false;

		if (given !== null && given !== undefined) {
			if (typeof given === "function") {
				answered = go2jsHTTPInvokeHandler(given, writer, request) === true;
			} else if (typeof given.ServeHTTP === "function") {
				answered = go2jsHTTPInvokeHandler(given.ServeHTTP, writer, request, given) === true;
			}
		}

		if (!writer.wrote) {
			writer.WriteHeader(writer.Code);
		}

		const answer = {
			Status: String(writer.Code) + " " + go2jsHTTPStatusText(writer.Code),
			StatusCode: writer.Code,
			Proto: "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header: writer.__go2js_header,
			Body: go2jsHTTPBody(Buffer.from(writer.parts.join(""), "utf8")),
			ContentLength: Buffer.byteLength(writer.parts.join(""), "utf8"),
			Request: request,
			Uncompressed: true,
			TransferEncoding: ["identity"],
			Trailer: go2jsHTTPHeader(),
			Close: false
		};

		writer.finished = true;
		writer.parts = [];

		return answer;
	};

	return writer;
}

// go2jsHTTPInvokeHandler calls a handler with the writer and the request, and
// says whether it answered. A handler that is written the way Go writes one is
// called with both, and a handler that returns a value of its own has answered
// with that value instead.
function go2jsHTTPInvokeHandler(handler, writer, request, self) {
	try {
		const result = self === null || self === undefined
			? go2jsCallNow(handler, null, [writer, request])
			: go2jsCallNow(handler, self, [writer, request]);

		if (result !== null && result !== undefined && result !== true) {
			// a handler that answered with a response of its own is written out
			// as the answer, which is what a handler written as a function of
			// the request alone comes back with
			if (typeof result.StatusCode === "number") {
				writer.WriteHeader(result.StatusCode);

				if (result.Body !== null && result.Body !== undefined) {
					go2jsHTTPWriteBytes(writer, go2jsHTTPBodyBytes(result.Body));
				}
			}
		}

		return true;
	} catch (err) {
		if (err !== null && err !== undefined && err.__go2js_panic === true) {
			throw err.value;
		}

		throw err;
	}
}

// go2jsHTTPResponseWriterFor is the writer handed to a handler, which is the
// writer itself when it is one and a fresh one otherwise.
function go2jsHTTPResponseWriterFor(value) {
	if (value !== null && value !== undefined && value.__go2js_interface === true) {
		return go2jsHTTPResponseWriterFor(value.value);
	}

	if (value !== null && value !== undefined && value.__go2js_pointer === true) {
		return go2jsHTTPResponseWriterFor(value[go2jsPointerGet]());
	}

	if (value !== null && value !== undefined && typeof value.__go2js_write === "function") {
		return value;
	}

	return go2jsHTTPResponseWriter();
}

// go2jsHTTPServeMuxHandleFunc registers a handler written as a function of the
// writer and the request against a path.
function go2jsHTTPServeMuxHandleFunc(mux, pattern, handler) {
	return go2jsHTTPHandleOn(mux, pattern, handler);
}

// go2jsHTTPServeMux holds the handlers a program registered against the paths
// they answer, and hands each request to the one registered for its path.
function go2jsHTTPNewServeMux() {
	const routes = [];
	const mux = {};

	// HandleFunc registers a handler written as a function of the writer and
	// the request, which is the shape most handlers are written in.
	mux.HandleFunc = function(pattern, handler) {
		go2jsHTTPHandleOn(mux, pattern, handler);
	};

	mux.Handle = function(pattern, handler) {
		routes.push({pattern: String(pattern), handler: handler});
	};

	mux.ServeHTTP = function(writer, request) {
		go2jsHTTPServeMuxServeHTTP(mux, writer, request);
	};

	mux.__go2js_routes = routes;

	return mux;
}

// go2jsHTTPHandlerUnder is a handler as it stands, with the boxes a value is
// carried in taken off: a handler handed to net/http as an interface or through
// a pointer is the same handler once it is unwrapped, and what answers a
// request is found on the value itself.
function go2jsHTTPHandlerUnder(handler) {
	let current = handler;

	for (let step = 0; step < 16; step++) {
		if (current === null || current === undefined) {
			return current;
		}

		if (current.__go2js_interface === true) {
			current = current.value;

			continue;
		}

		if (current.__go2js_pointer === true) {
			current = current[go2jsPointerGet]();

			continue;
		}

		return current;
	}

	return current;
}

// go2jsHTTPHandleOn registers a handler against a path of a mux, which is the
// one registration net/http has. What a program hands over is kept as it stands,
// since a handler written as a function and one written as an object that
// answers a request are told apart by what they are rather than by where they
// came from.
function go2jsHTTPHandleOn(mux, pattern, handler) {
	if (mux === null || mux === undefined || !Array.isArray(mux.__go2js_routes)) {
		return;
	}

	mux.__go2js_routes.push({pattern: String(pattern), handler: handler});
}

// go2jsHTTPServeMuxMatch finds the handler registered for a path, preferring the
// longest pattern that matches it, which is the one a program meant.
function go2jsHTTPServeMuxMatch(routes, path) {
	let best = null;

	for (const route of routes) {
		if (!go2jsHTTPServeMuxMatches(route.pattern, path)) {
			continue;
		}

		if (best === null || route.pattern.length > best.pattern.length) {
			best = route;
		}
	}

	return best;
}

// go2jsHTTPServeMuxMatches says whether a pattern stands for a path, which it
// does when it names the path outright, or ends in a slash and stands for
// everything written under it.
function go2jsHTTPServeMuxMatches(pattern, path) {
	if (pattern === path) {
		return true;
	}

	if (pattern === "/") {
		return true;
	}

	if (pattern.endsWith("/")) {
		return path.startsWith(pattern);
	}

	return false;
}

function go2jsHTTPServeMuxServeHTTP(mux, writer, request) {
	const target = writer === null || writer === undefined ? null : writer;
	const output = go2jsHTTPResponseWriterFor(target);
	const routes = mux === null || mux === undefined ? [] : (mux.__go2js_routes || []);
	const path = go2jsHTTPRequestPath(request);
	const route = go2jsHTTPServeMuxMatch(routes, path);

	if (route === null) {
		go2jsHTTPNotFound(output, request);

		return;
	}

	// A handler is a function of the writer and the request, or an object that
	// has a way of answering one, and it is looked for with the boxes a handler
	// may be carried in taken off first, since a handler handed over as an
	// interface is the same handler underneath.
	const handler = go2jsHTTPHandlerUnder(route.handler);

	if (typeof handler === "function") {
		go2jsHTTPInvokeHandler(handler, output, request);

		return;
	}

	if (handler !== null && handler !== undefined && typeof handler.ServeHTTP === "function") {
		go2jsHTTPInvokeHandler(handler.ServeHTTP, output, request, handler);
	}
}

// go2jsHTTPRequestPath is the part of a request that names what is being asked
// for, which for a request that arrived is the path of its url.
function go2jsHTTPRequestPath(request) {
	if (request === null || request === undefined) {
		return "/";
	}

	if (typeof request.RequestURI === "string" && request.RequestURI !== "") {
		const cut = request.RequestURI.indexOf("?");

		return cut === -1 ? request.RequestURI : request.RequestURI.slice(0, cut);
	}

	const url = go2jsUnwrap(request.URL);

	if (url === null || url === undefined) {
		return "/";
	}

	return url.Path === "" ? "/" : url.Path;
}

// go2jsHTTPError answers a request with a fault in place of what was asked for.
function go2jsHTTPError(writer, message, code) {
	const output = go2jsHTTPResponseWriterFor(writer);
	const text = String(message) + "\n";

	output.__go2js_header.Del("Content-Length");
	output.__go2js_header.Set("Content-Type", "text/plain; charset=utf-8");
	output.__go2js_header.Set("X-Content-Type-Options", "nosniff");
	output.WriteHeader(Number(code));
	go2jsHTTPWriteString(output, text);
}

function go2jsHTTPNotFound(writer, request) {
	go2jsHTTPError(writer, "404 page not found", 404);
}

class User {
    ID;
    Name;
    Email;
    constructor() {
        this.ID = 0;
        this.Name = "";
        this.Email = "";
    }
}
go2jsRegisterTypeName(User, "main.User", [], [{Name: "ID", Tag: "json:\"id\"", PkgPath: "", Anonymous: false, Type: "int", Kind: "int"}, {Name: "Name", Tag: "json:\"name\"", PkgPath: "", Anonymous: false, Type: "string", Kind: "string"}, {Name: "Email", Tag: "json:\"email\"", PkgPath: "", Anonymous: false, Type: "string", Kind: "string"}], "struct");


if (typeof go2jsBytesBuffer === "function") { go2jsRegisterTypeName(go2jsBytesBuffer, "bytes.Buffer"); }
if (typeof go2jsHTTPHeader === "function") { go2jsRegisterTypeName(go2jsHTTPHeader, "http.Header"); }
if (typeof go2jsHTTPNewServeMux === "function") { go2jsRegisterTypeName(go2jsHTTPNewServeMux, "http.ServeMux"); }
if (typeof go2jsMutex === "function") { go2jsRegisterTypeName(go2jsMutex, "sync.Mutex"); }
if (typeof go2jsWaitGroup === "function") { go2jsRegisterTypeName(go2jsWaitGroup, "sync.WaitGroup"); }
if (typeof go2jsURL === "function") { go2jsRegisterTypeName(go2jsURL, "url.URL"); }

go2jsSetPackageName("main");
go2jsSetSourceLines({1: "/home/ollama/Desktop/go2js/examples/realworld/main.go:27:6", 2: "/home/ollama/Desktop/go2js/examples/realworld/main.go:28:2", 3: "/home/ollama/Desktop/go2js/examples/realworld/main.go:30:26", 4: "/home/ollama/Desktop/go2js/examples/realworld/main.go:31:27", 5: "/home/ollama/Desktop/go2js/examples/realworld/main.go:32:26", 6: "/home/ollama/Desktop/go2js/examples/realworld/main.go:33:37", 9: "/home/ollama/Desktop/go2js/examples/realworld/main.go:36:6", 10: "/home/ollama/Desktop/go2js/examples/realworld/main.go:37:7", 11: "/home/ollama/Desktop/go2js/examples/realworld/main.go:38:7", 12: "/home/ollama/Desktop/go2js/examples/realworld/main.go:40:2", 13: "/home/ollama/Desktop/go2js/examples/realworld/main.go:41:2", 14: "/home/ollama/Desktop/go2js/examples/realworld/main.go:42:2", 15: "/home/ollama/Desktop/go2js/examples/realworld/main.go:44:24", 16: "/home/ollama/Desktop/go2js/examples/realworld/main.go:45:57", 19: "/home/ollama/Desktop/go2js/examples/realworld/main.go:48:6", 20: "/home/ollama/Desktop/go2js/examples/realworld/main.go:49:10", 21: "/home/ollama/Desktop/go2js/examples/realworld/main.go:50:10", 22: "/home/ollama/Desktop/go2js/examples/realworld/main.go:51:7", 23: "/home/ollama/Desktop/go2js/examples/realworld/main.go:53:2", 26: "/home/ollama/Desktop/go2js/examples/realworld/main.go:56:6", 28: "/home/ollama/Desktop/go2js/examples/realworld/main.go:57:2", 29: "/home/ollama/Desktop/go2js/examples/realworld/main.go:63:15", 30: "/home/ollama/Desktop/go2js/examples/realworld/main.go:64:2", 31: "/home/ollama/Desktop/go2js/examples/realworld/main.go:65:23", 32: "/home/ollama/Desktop/go2js/examples/realworld/main.go:67:2", 33: "/home/ollama/Desktop/go2js/examples/realworld/main.go:68:8", 34: "/home/ollama/Desktop/go2js/examples/realworld/main.go:70:2", 35: "/home/ollama/Desktop/go2js/examples/realworld/main.go:71:2", 38: "/home/ollama/Desktop/go2js/examples/realworld/main.go:74:6", 39: "/home/ollama/Desktop/go2js/examples/realworld/main.go:75:10", 40: "/home/ollama/Desktop/go2js/examples/realworld/main.go:77:2", 41: "/home/ollama/Desktop/go2js/examples/realworld/main.go:78:2", 42: "/home/ollama/Desktop/go2js/examples/realworld/main.go:79:2", 43: "/home/ollama/Desktop/go2js/examples/realworld/main.go:80:28", 44: "/home/ollama/Desktop/go2js/examples/realworld/main.go:82:2", 45: "/home/ollama/Desktop/go2js/examples/realworld/main.go:83:2", 46: "/home/ollama/Desktop/go2js/examples/realworld/main.go:84:2", 47: "/home/ollama/Desktop/go2js/examples/realworld/main.go:86:29", 50: "/home/ollama/Desktop/go2js/examples/realworld/main.go:89:6", 51: "/home/ollama/Desktop/go2js/examples/realworld/main.go:90:7", 52: "/home/ollama/Desktop/go2js/examples/realworld/main.go:92:31", 53: "/home/ollama/Desktop/go2js/examples/realworld/main.go:93:30", 56: "/home/ollama/Desktop/go2js/examples/realworld/main.go:96:6", 57: "/home/ollama/Desktop/go2js/examples/realworld/main.go:97:2", 58: "/home/ollama/Desktop/go2js/examples/realworld/main.go:99:2", 59: "/home/ollama/Desktop/go2js/examples/realworld/main.go:101:2", 60: "/home/ollama/Desktop/go2js/examples/realworld/main.go:103:2", 61: "/home/ollama/Desktop/go2js/examples/realworld/main.go:104:2", 62: "/home/ollama/Desktop/go2js/examples/realworld/main.go:106:2", 65: "/home/ollama/Desktop/go2js/examples/realworld/main.go:109:6", 66: "/home/ollama/Desktop/go2js/examples/realworld/main.go:110:7", 67: "/home/ollama/Desktop/go2js/examples/realworld/main.go:112:32", 68: "/home/ollama/Desktop/go2js/examples/realworld/main.go:113:31", 69: "/home/ollama/Desktop/go2js/examples/realworld/main.go:114:31", 70: "/home/ollama/Desktop/go2js/examples/realworld/main.go:115:33", 73: "/home/ollama/Desktop/go2js/examples/realworld/main.go:118:6", 74: "/home/ollama/Desktop/go2js/examples/realworld/main.go:119:26", 75: "/home/ollama/Desktop/go2js/examples/realworld/main.go:120:31", 76: "/home/ollama/Desktop/go2js/examples/realworld/main.go:122:2", 77: "/home/ollama/Desktop/go2js/examples/realworld/main.go:123:2", 78: "/home/ollama/Desktop/go2js/examples/realworld/main.go:125:25", 79: "/home/ollama/Desktop/go2js/examples/realworld/main.go:127:15", 80: "/home/ollama/Desktop/go2js/examples/realworld/main.go:128:2", 81: "/home/ollama/Desktop/go2js/examples/realworld/main.go:130:2", 82: "/home/ollama/Desktop/go2js/examples/realworld/main.go:131:32", 86: "/home/ollama/Desktop/go2js/examples/realworld/main.go:135:6", 87: "/home/ollama/Desktop/go2js/examples/realworld/main.go:136:9", 88: "/home/ollama/Desktop/go2js/examples/realworld/main.go:138:25", 89: "/home/ollama/Desktop/go2js/examples/realworld/main.go:139:29", 92: "/home/ollama/Desktop/go2js/examples/realworld/main.go:142:6", 93: "/home/ollama/Desktop/go2js/examples/realworld/main.go:143:12", 94: "/home/ollama/Desktop/go2js/examples/realworld/main.go:145:15", 95: "/home/ollama/Desktop/go2js/examples/realworld/main.go:147:21", 96: "/home/ollama/Desktop/go2js/examples/realworld/main.go:148:2", 99: "/home/ollama/Desktop/go2js/examples/realworld/main.go:151:6", 100: "/home/ollama/Desktop/go2js/examples/realworld/main.go:152:14", 101: "/home/ollama/Desktop/go2js/examples/realworld/main.go:158:2", 102: "/home/ollama/Desktop/go2js/examples/realworld/main.go:159:3", 103: "/home/ollama/Desktop/go2js/examples/realworld/main.go:160:3", 105: "/home/ollama/Desktop/go2js/examples/realworld/main.go:163:2", 106: "/home/ollama/Desktop/go2js/examples/realworld/main.go:164:2", 107: "/home/ollama/Desktop/go2js/examples/realworld/main.go:166:2", 108: "/home/ollama/Desktop/go2js/examples/realworld/main.go:167:27", 109: "/home/ollama/Desktop/go2js/examples/realworld/main.go:168:34", 110: "/home/ollama/Desktop/go2js/examples/realworld/main.go:169:30", 113: "/home/ollama/Desktop/go2js/examples/realworld/main.go:172:6", 114: "/home/ollama/Desktop/go2js/examples/realworld/main.go:173:9", 115: "/home/ollama/Desktop/go2js/examples/realworld/main.go:175:2", 116: "/home/ollama/Desktop/go2js/examples/realworld/main.go:176:3", 119: "/home/ollama/Desktop/go2js/examples/realworld/main.go:179:9", 120: "/home/ollama/Desktop/go2js/examples/realworld/main.go:181:22", 121: "/home/ollama/Desktop/go2js/examples/realworld/main.go:182:2", 124: "/home/ollama/Desktop/go2js/examples/realworld/main.go:185:6", 125: "/home/ollama/Desktop/go2js/examples/realworld/main.go:186:12", 126: "/home/ollama/Desktop/go2js/examples/realworld/main.go:187:2", 129: "/home/ollama/Desktop/go2js/examples/realworld/main.go:190:6", 130: "/home/ollama/Desktop/go2js/examples/realworld/main.go:191:9", 131: "/home/ollama/Desktop/go2js/examples/realworld/main.go:202:28", 132: "/home/ollama/Desktop/go2js/examples/realworld/main.go:203:29", 133: "/home/ollama/Desktop/go2js/examples/realworld/main.go:204:27", 134: "/home/ollama/Desktop/go2js/examples/realworld/main.go:205:28", 135: "/home/ollama/Desktop/go2js/examples/realworld/main.go:206:33", 136: "/home/ollama/Desktop/go2js/examples/realworld/main.go:208:11", 137: "/home/ollama/Desktop/go2js/examples/realworld/main.go:210:29", 138: "/home/ollama/Desktop/go2js/examples/realworld/main.go:211:32", 141: "/home/ollama/Desktop/go2js/examples/realworld/main.go:214:6", 142: "/home/ollama/Desktop/go2js/examples/realworld/main.go:215:2", 143: "/home/ollama/Desktop/go2js/examples/realworld/main.go:216:2", 144: "/home/ollama/Desktop/go2js/examples/realworld/main.go:218:2", 145: "/home/ollama/Desktop/go2js/examples/realworld/main.go:220:2", 146: "/home/ollama/Desktop/go2js/examples/realworld/main.go:221:3", 147: "/home/ollama/Desktop/go2js/examples/realworld/main.go:223:6", 157: "/home/ollama/Desktop/go2js/examples/realworld/main.go:224:10", 158: "/home/ollama/Desktop/go2js/examples/realworld/main.go:226:4", 159: "/home/ollama/Desktop/go2js/examples/realworld/main.go:227:4", 160: "/home/ollama/Desktop/go2js/examples/realworld/main.go:228:4", 174: "/home/ollama/Desktop/go2js/examples/realworld/main.go:232:2", 175: "/home/ollama/Desktop/go2js/examples/realworld/main.go:234:2", 178: "/home/ollama/Desktop/go2js/examples/realworld/main.go:237:6", 179: "/home/ollama/Desktop/go2js/examples/realworld/main.go:238:2", 180: "/home/ollama/Desktop/go2js/examples/realworld/main.go:239:2", 181: "/home/ollama/Desktop/go2js/examples/realworld/main.go:241:2", 182: "/home/ollama/Desktop/go2js/examples/realworld/main.go:242:2", 183: "/home/ollama/Desktop/go2js/examples/realworld/main.go:244:2", 184: "/home/ollama/Desktop/go2js/examples/realworld/main.go:245:2", 185: "/home/ollama/Desktop/go2js/examples/realworld/main.go:247:2", 186: "/home/ollama/Desktop/go2js/examples/realworld/main.go:248:2", 187: "/home/ollama/Desktop/go2js/examples/realworld/main.go:250:2", 188: "/home/ollama/Desktop/go2js/examples/realworld/main.go:251:2", 189: "/home/ollama/Desktop/go2js/examples/realworld/main.go:253:2", 190: "/home/ollama/Desktop/go2js/examples/realworld/main.go:254:2", 191: "/home/ollama/Desktop/go2js/examples/realworld/main.go:256:2", 192: "/home/ollama/Desktop/go2js/examples/realworld/main.go:257:2", 193: "/home/ollama/Desktop/go2js/examples/realworld/main.go:259:2", 194: "/home/ollama/Desktop/go2js/examples/realworld/main.go:260:2", 195: "/home/ollama/Desktop/go2js/examples/realworld/main.go:262:2", 196: "/home/ollama/Desktop/go2js/examples/realworld/main.go:263:2", 197: "/home/ollama/Desktop/go2js/examples/realworld/main.go:265:2", 198: "/home/ollama/Desktop/go2js/examples/realworld/main.go:266:2", 199: "/home/ollama/Desktop/go2js/examples/realworld/main.go:268:2", 200: "/home/ollama/Desktop/go2js/examples/realworld/main.go:269:2", 201: "/home/ollama/Desktop/go2js/examples/realworld/main.go:271:2", 202: "/home/ollama/Desktop/go2js/examples/realworld/main.go:272:2", 203: "/home/ollama/Desktop/go2js/examples/realworld/main.go:274:2", 204: "/home/ollama/Desktop/go2js/examples/realworld/main.go:275:2", 205: "/home/ollama/Desktop/go2js/examples/realworld/main.go:277:2", 206: "/home/ollama/Desktop/go2js/examples/realworld/main.go:278:2", 207: "/home/ollama/Desktop/go2js/examples/realworld/main.go:280:2"});
function* testStrings() {
    let s = "go2js standard library test";
    go2jsPrintln("strings:", go2jsStringsToUpper(s));
    go2jsPrintln("contains:", go2jsStringsContains(s, "standard"));
    go2jsPrintln("replace:", go2jsStringsReplaceAll(s, "library", "stdlib"));
    go2jsPrintln("split:", go2jsStringsJoin(go2jsStringsSplit("one,two,three", ","), "|"));
}

function* testBytes() {
    let a = go2jsStringToBytes("hello ");
    let b = go2jsStringToBytes("world");
    let buf = new go2jsBytesBuffer();
    buf.Write(a);
    buf.Write(b);
    go2jsPrintln("bytes:", buf.String());
    go2jsPrintln("bytes equal:", go2jsBytesEqual(go2jsStringToBytes("abc"), go2jsStringToBytes("abc")));
}

function* testStrconv() {
    let [n] = go2jsStrconvAtoi("12345");
    let [f] = go2jsStrconvParseFloat("12.5", 64);
    let b = go2jsStrconvFormatInt(999, 10);
    go2jsPrintln("strconv:", n, go2jsTyped(f, "float64"), b);
}

function* testJSON() {
    let $go2js_addr_1 = null;
    let u = Object.assign(new User(), {ID: 42, Name: "Alice", Email: "alice@example.com"});
    let [data, err] = go2jsJSONMarshal(u, {"ID": "id", "Name": "name", "Email": "email"}, null, null, null, "", "");
    go2jsPrintln("json error:", err);
    go2jsPrintln("json:", go2jsBytesToString(data));
    let decoded = new User();
    err = go2jsJSONUnmarshal(data, ($go2js_addr_1 ??= go2jsPtr(() => decoded, value => decoded = value)), {"ID": "id", "Name": "name", "Email": "email"}, {"kind":"ptr","elem":{"kind":"struct","name":"main.User","fields":{"ID": "id", "Name": "name", "Email": "email"},"string":null,"types":{"ID": {"kind":"number"}, "Name": {"kind":"string"}, "Email": {"kind":"string"}}}}, null);
    go2jsPrintln("json decoded:", decoded.ID, decoded.Name, decoded.Email);
    go2jsPrintln("json decode error:", err);
}

function* testURL() {
    let [u] = go2jsURLParse("https://example.com/users?id=42&name=alice");
    go2jsPrintln("url scheme:", u.Scheme);
    go2jsPrintln("url host:", u.Host);
    go2jsPrintln("url path:", u.Path);
    go2jsPrintln("url query:", go2jsNamedMethodCall(u.Query(), "Values", "Get", "name"));
    let values = go2jsMapTyped("net/url.Values", go2jsMap([]));
    go2jsNamedMethodCall(values, "Values", "Set", "page", "2");
    go2jsNamedMethodCall(values, "Values", "Set", "limit", "50");
    go2jsPrintln("url values:", go2jsNamedMethodCall(values, "Values", "Encode"));
}

function* testRegexp() {
    let r = go2jsRegexpMustCompile("go[0-9]+js");
    go2jsPrintln("regexp match:", r.MatchString("this is go2js"));
    go2jsPrintln("regexp find:", r.FindString("project go2js works"));
}

function* testSort() {
    let values = go2jsSliceTyped("[]int", [9, 2, 7, 1, 5, 3]);
    go2jsSortInts(values);
    go2jsPrintln("sort:", go2jsTyped(values, "[]int", "", "slice"));
    let names = go2jsSliceTyped("[]string", ["charlie", "alice", "bob"]);
    go2jsSortStrings(names);
    go2jsPrintln("sort strings:", go2jsTyped(names, "[]string", "", "slice"));
}

function* testFilepath() {
    let p = go2jsFilepathJoin("tmp", "go2js", "test.txt");
    go2jsPrintln("filepath base:", go2jsFilepathBase(p));
    go2jsPrintln("filepath dir:", go2jsFilepathDir(p));
    go2jsPrintln("filepath ext:", go2jsFilepathExt(p));
    go2jsPrintln("filepath clean:", go2jsFilepathClean("./tmp/../tmp/test.txt"));
}

function* testOS() {
    go2jsPrintln("os args:", go2jsLen(go2jsOSArgs()));
    go2jsPrintln("os separator:", go2jsRuneString(47));
    let name = "GO2JS_TEST_VALUE";
    go2jsOSSetenv(name, "hello");
    go2jsPrintln("os env:", go2jsOSGetenv(name));
    let [info, err] = go2jsOSStat("stdlib_realworld.go");
    go2jsPrintln("os stat exists:", go2jsEqual(err, null));
    if (info != null) {
        go2jsPrintln("os stat size:", (yield* go2jsInterfaceCall(info, "Size")) > 0);
    }
}

function* testErrors() {
    let err = go2jsErrorsNew("test error");
    go2jsPrintln("errors:", (yield* go2jsInterfaceCall(err, "Error")));
    go2jsPrintln("errors nil:", go2jsEqual(null, null));
}

function* testIO() {
    let reader = go2jsStringsNewReader("hello from io");
    let [data, err] = go2jsIOReadAll(go2jsInterface(reader, "*Reader", "*strings.Reader"));
    go2jsPrintln("io:", go2jsBytesToString(data));
    go2jsPrintln("io error:", err);
}

function* testHTTP() {
    let [req, err] = go2jsHTTPNewRequest("GET", "https://example.com/api/test?value=42", null);
    if (go2jsEqual(err, null) === false) {
        go2jsPrintln("http request error:", err);
        return;
    }
    go2jsHTTPHeaderSet(req.Header, "User-Agent", "go2js-test");
    go2jsHTTPHeaderSet(req.Header, "X-Test", "hello");
    go2jsPrintln("http method:", req.Method);
    go2jsPrintln("http url:", req.URL.String());
    go2jsPrintln("http user-agent:", go2jsHTTPHeaderGet(req.Header, "User-Agent"));
    go2jsPrintln("http x-test:", go2jsHTTPHeaderGet(req.Header, "X-Test"));
}

function* testHTTPServer() {
    let mux = go2jsHTTPNewServeMux();
    go2jsHTTPServeMuxHandleFunc(mux, "/hello", function*(w, r) {
        go2jsFprint(w, ["hello"], "", "");
    }
);
    let req = (yield* httptestRequest("/hello"));
    let [handler, pattern] = mux.Handler(req);
    go2jsPrintln("http route:", go2jsEqual(handler, null) === false, pattern);
}

function* httptestRequest(path) {
    let [req] = go2jsHTTPNewRequest("GET", "http://localhost" + path, null);
    return req;
}

function* testTime() {
    let now = go2jsStructCopy(go2jsTimeDate(2026, 9, 26, 12, 30, 45, 0, go2jsTimeLocation("UTC")));
    go2jsPrintln("time year:", now.Year());
    go2jsPrintln("time month:", go2jsTyped(now.Month(), "time.Month", "int"));
    go2jsPrintln("time day:", now.Day());
    go2jsPrintln("time hour:", now.Hour());
    go2jsPrintln("time formatted:", now.Format("2006-01-02 15:04:05"));
    let later = go2jsStructCopy(now.Add(go2jsDuration(go2jsWideAdd(go2jsDuration(go2jsWideMul(2, go2jsDuration(3600000000000), "int64")), go2jsDuration(go2jsWideMul(30, go2jsDuration(60000000000), "int64")), "int64"))));
    go2jsPrintln("time later:", later.Format("15:04"));
    go2jsPrintln("time duration:", go2jsTyped(later.Sub(go2jsStructCopy(now)), "time.Duration", "int64"));
}

function* testConcurrency() {
    let wg = go2jsWaitGroup();
    let mu = go2jsMutex();
    let total = 0;
    for (let i = 0; i < 10; i++) {
        go2jsWaitGroupAdd(wg, 1);
        go2jsGo(function* () { (yield* function*() {
            const go2jsDefers = [];
            let go2jsPanicValue;
            let go2jsRecovered = false;            const go2jsRecover = () => {
                if (go2jsPanicValue === undefined || go2jsRecovered) return undefined;
                if (go2jsPanicValue === go2jsGoexitSignal) return undefined;
                go2jsRecovered = true;
                return go2jsPanicPayload(go2jsPanicValue);
            };
            try {
                (() => go2jsDefers.push(function* () { go2jsWaitGroupDone(wg) }))();
                yield* go2jsMutexLock(mu);
                                total = go2jsWideAdd(total, 1, "int");
                go2jsMutexUnlock(mu);
            } catch (go2jsCaught) {
                go2jsPanicValue = go2jsCaught;
            } finally {
                for (let i = go2jsDefers.length - 1; i >= 0; i--) {
                    yield* go2jsDefers[i]();
                }
                if (go2jsPanicValue !== undefined && !go2jsRecovered) {
                    throw go2jsPanicValue;
                }
            }
        }
()) });
    }
    yield* go2jsWaitGroupWait(wg);
    go2jsPrintln("sync total:", total);
}

function* main() {
    go2jsPrintln("=== strings ===");
    (yield* testStrings());
    go2jsPrintln("=== bytes ===");
    (yield* testBytes());
    go2jsPrintln("=== strconv ===");
    (yield* testStrconv());
    go2jsPrintln("=== json ===");
    (yield* testJSON());
    go2jsPrintln("=== url ===");
    (yield* testURL());
    go2jsPrintln("=== regexp ===");
    (yield* testRegexp());
    go2jsPrintln("=== sort ===");
    (yield* testSort());
    go2jsPrintln("=== filepath ===");
    (yield* testFilepath());
    go2jsPrintln("=== os ===");
    (yield* testOS());
    go2jsPrintln("=== errors ===");
    (yield* testErrors());
    go2jsPrintln("=== io ===");
    (yield* testIO());
    go2jsPrintln("=== http ===");
    (yield* testHTTP());
    go2jsPrintln("=== time ===");
    (yield* testTime());
    go2jsPrintln("=== sync ===");
    (yield* testConcurrency());
    go2jsPrintln("=== done ===");
}

try { go2jsRunMain((function* () { return yield* main(); })()); } catch (e) { go2jsReportUncaughtPanic(e); }
