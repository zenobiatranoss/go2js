"use strict";
(function() {
	if (typeof process !== "undefined" || typeof globalThis === "undefined") {
		return null;
	}

	const decoder = typeof TextDecoder === "function" ? new TextDecoder() : null;

	const lineWriter = function(log) {
		let pending = "";

		return {
			write(value) {
				pending += typeof value === "string" ? value : (decoder !== null ? decoder.decode(value) : String(value));

				let at = pending.indexOf("\n");

				while (at >= 0) {
					log(pending.slice(0, at));
					pending = pending.slice(at + 1);
					at = pending.indexOf("\n");
				}

				return true;
			},
		};
	};

	globalThis.process = {
		stdout: lineWriter((line) => console.log(line)),
		stderr: lineWriter((line) => console.warn(line)),
		env: {},
		argv: ["go2js"],
		platform: "browser",
		exit(code) {
			throw new Error("exit status " + code);
		},
		on() {},
		cwd() {
			return "/";
		},
		hrtime: {bigint() { return BigInt(Math.round(performance.now() * 1e6)); }},
	};

	return globalThis.process;
})();

const go2jsNativeMap = globalThis.Map;

const go2jsNativeSet = globalThis.Set;

const go2jsNativeDate = globalThis.Date;

const go2jsNativeTypeError = globalThis.TypeError;

const go2jsNativeRangeError = globalThis.RangeError;

const go2jsNativeError = globalThis.Error;

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

// go2jsDecodeUTF8Rune turns the bytes starting at a place into the rune they
// say and the number of bytes it took, the way unicode/utf8 does: a rune that
// is not there or is not well formed is the replacement character of size one.
function go2jsDecodeUTF8Rune(bytes, index) {
    const head = bytes[index] & 0xff;

    if (head < 0x80) {
        return [head, 1];
    }

    let count;
    let code;

    if (head < 0xc0) {
        return [0xfffd, 1];
    } else if (head < 0xe0) {
        count = 2;
        code = head & 0x1f;

        if (code < 2) {
            return [0xfffd, 1];
        }
    } else if (head < 0xf0) {
        count = 3;
        code = head & 0x0f;
    } else if (head < 0xf8) {
        count = 4;
        code = head & 0x07;

        if (code > 4) {
            return [0xfffd, 1];
        }
    } else {
        return [0xfffd, 1];
    }

    if (index + count > bytes.length) {
        return [0xfffd, 1];
    }

    for (let at = 1; at < count; at++) {
        const piece = bytes[index + at] & 0xff;

        if ((piece & 0xc0) !== 0x80) {
            return [0xfffd, 1];
        }

        code = (code << 6) | (piece & 0x3f);
    }

    if ((count === 2 && code < 0x80) ||
        (count === 3 && code < 0x800) ||
        (count === 4 && code < 0x10000) ||
        (code >= 0xd800 && code <= 0xdfff) ||
        code > 0x10ffff) {
        return [0xfffd, 1];
    }

    return [code, count];
}

// go2jsNewByteReader reads its bytes from the front and keeps the value whole,
// the way strings.Reader and bytes.Reader both do; the word it uses in an
// error tells which one it stands for. The bytes are the UTF-8 of the text for
// a string reader and the bytes of the slice for a byte reader, so Read hands
// back the same numbers the Go runtime would.
function go2jsNewByteReader(data, kind) {
    let bytes = go2jsToArray(data);
    let total = bytes.length;
    let position = 0;
    let prevRune = -1;
    const word = kind === "slice" ? "slice" : "string";

    function err(text) {
        return go2jsErrorsNew((kind === "slice" ? "bytes.Reader." : "strings.Reader.") + text);
    }

    return {
        __go2js_text: kind === "slice" ? null : go2jsBytesToString(bytes),
        __go2js_bytes: bytes,
        Read(target) {
            prevRune = -1;
            const slot = go2jsBytesReadSlot(target);

            if (position >= total || !slot) {
                return [0, go2jsIOEOF()];
            }

            const count = Math.min(slot.length, total - position);

            for (let index = 0; index < count; index++) {
                go2jsBytesStoreByte(target, index, bytes[position + index]);
            }

            position += count;
            return [count, null];
        },
        ReadAt(target, off) {
            prevRune = -1;
            off = Math.trunc(Number(off));

            if (off < 0) {
                return [0, err("ReadAt: negative offset")];
            }

            if (off >= total) {
                return [0, go2jsIOEOF()];
            }

            const slot = go2jsBytesReadSlot(target);

            if (!slot) {
                return [0, go2jsIOEOF()];
            }

            const count = Math.min(slot.length, total - off);

            for (let index = 0; index < count; index++) {
                go2jsBytesStoreByte(target, index, bytes[off + index]);
            }

            if (count < slot.length) {
                return [count, go2jsIOEOF()];
            }

            return [count, null];
        },
        ReadByte() {
            prevRune = -1;

            if (position >= total) {
                return [0, go2jsIOEOF()];
            }

            return [bytes[position++], null];
        },
        UnreadByte() {
            prevRune = -1;

            if (position <= 0) {
                return err("UnreadByte: at beginning of " + word);
            }

            position--;
            return null;
        },
        ReadRune() {
            if (position >= total) {
                prevRune = -1;
                return [0, 0, go2jsIOEOF()];
            }

            const decoded = go2jsDecodeUTF8Rune(bytes, position);
            const size = decoded[1];

            position += size;
            prevRune = position - size;
            return [decoded[0], size, null];
        },
        UnreadRune() {
            if (position <= 0) {
                return err("UnreadRune: at beginning of " + word);
            }

            if (prevRune < 0) {
                return err("UnreadRune: previous operation was not ReadRune");
            }

            position = prevRune;
            prevRune = -1;
            return null;
        },
        Seek(offset, whence) {
            offset = Math.trunc(Number(offset));
            const base = whence === 0 ? 0 : whence === 1 ? position : whence === 2 ? total : null;

            if (base === null) {
                return [0, err("Seek: invalid whence")];
            }

            const next = base + offset;

            if (next < 0) {
                return [0, err("Seek: negative position")];
            }

            position = next;
            prevRune = -1;
            return [next, null];
        },
        Size() {
            return total;
        },
        Len() {
            return Math.max(0, total - position);
        },
        Reset(next) {
            bytes = go2jsToArray(next);
            total = bytes.length;
            position = 0;
            prevRune = -1;
        },
        WriteTo(writer) {
            prevRune = -1;

            if (position >= total) {
                return [0, null];
            }

            const result = go2jsCallMethod(writer, "Write", bytes.slice(position));
            const wrote = Array.isArray(result) ? result[0] : Number(result);
            const failure = Array.isArray(result) ? result[1] : null;

            if (wrote > 0) {
                position += wrote;
            }

            return [wrote, failure];
        }
    };
}

// go2jsStringsNewReader reads text the way strings.NewReader does, by the UTF-8
// bytes of the string, and answers the whole reader's ways to step through it.
function go2jsStringsNewReader(value) {
    return go2jsNewByteReader(go2jsStringToBytes(go2jsStringify(value)), "string");
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

    return [chunks, null];
}

function go2jsMonthName(month) {
    const names = [
        "January", "February", "March", "April", "May", "June",
        "July", "August", "September", "October", "November", "December"
    ];

    // A month outside the year is not a month, and Go writes it out as the
    // number it was rather than as a name of its own. A month is a signed type
    // written through the unsigned it shares its bits with, so a month below
    // January reads back as the largest unsigned number it makes.
    if (month >= 1 && month <= 12) {
        return names[month - 1];
    }

    return "%!Month(" + String(go2jsWideWrap(month, "uint64")) + ")";
}

function go2jsWeekdayName(weekday) {
    const names = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];

    // A day outside the week is not a day, and Go writes it out as the number it
    // was rather than as a name of its own. A day is a signed type written
    // through the unsigned it shares its bits with, so a day before Sunday reads
    // back as the largest unsigned number it makes.
    if (weekday >= 0 && weekday <= 6) {
        return names[weekday];
    }

    return "%!Weekday(" + String(go2jsWideWrap(weekday, "uint64")) + ")";
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
    // the clock of the host is not a moment nothing was ever told. Its location
    // is UTC, which is the name Go writes for it.
    zero.__go2js_time_zero = true;
    zero.__go2js_location = "UTC";

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
        Date: function() {
            return [this.Year(), this.Month(), this.Day()];
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
        ISOWeek: function() {
            // The ISO year and week are read from the Thursday of the week the
            // moment falls in, since the year that Thursday falls in is the year
            // the whole week belongs to.
            const local = fields(this);
            const day = (local.getUTCDay() + 6) % 7;
            const thursday = new Date(Date.UTC(local.getUTCFullYear(), local.getUTCMonth(), local.getUTCDate() + (3 - day)));
            const year = thursday.getUTCFullYear();
            const dayOfYear = Math.floor((thursday.getTime() - Date.UTC(year, 0, 1)) / 86400000) + 1;

            return [year, Math.floor((dayOfYear - 1) / 7) + 1];
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

            // Reckoning a moment from its fields keeps the finer part it was
            // holding, since none of those fields are a fraction of a second.
            next.__go2js_nsec = this.__go2js_nsec === undefined ? 0 : this.__go2js_nsec;
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
            // the epoch is a count of the second it sits before it. The whole
            // count is wider than a double keeps exactly, so it is worked out
            // as digits and narrowed back once it fits.
            return go2jsNarrow(BigInt(go2jsTimeQuotient(this.value.getTime(), 1000)) * 1000000000n + BigInt(go2jsTimeNanos(this)));
        },
        UnixMicro: function() {
            // The count of microseconds is the count of seconds and the whole
            // microseconds within one, cut off the nanoseconds below them.
            return go2jsNarrow(BigInt(go2jsTimeQuotient(this.value.getTime(), 1000)) * 1000000n + BigInt(Math.trunc(go2jsTimeNanos(this) / 1000)));
        },
        UnixMilli: function() {
            // The count of milliseconds is the count of seconds and the whole
            // milliseconds within one, which is the count of them the moment's
            // fields already hold.
            return go2jsNarrow(BigInt(go2jsTimeQuotient(this.value.getTime(), 1000)) * 1000n + BigInt(Math.trunc(go2jsTimeNanos(this) / 1000000)));
        },
        IsZero: function() {
            return this.value.getTime() === 0 || this.__go2js_time_zero === true;
        },
        Before: function(other) {
            return this.Compare(other) < 0;
        },
        After: function(other) {
            return this.Compare(other) > 0;
        },
        Equal: function(other) {
            return this.Compare(other) === 0;
        },
        Compare: function(other) {
            // Two moments read against the same monotonic clock are put in order
            // by that reading, which is finer than the wall clock; otherwise the
            // wall clock is read down to the nanosecond, since two moments within
            // the same millisecond are not the same moment.
            const that = go2jsTimeOf(other);

            if (this.__go2js_mono !== undefined && that.__go2js_mono !== undefined) {
                return this.__go2js_mono < that.__go2js_mono ? -1 : (this.__go2js_mono > that.__go2js_mono ? 1 : 0);
            }

            const here = go2jsTimeQuotient(this.value.getTime(), 1000);
            const there = go2jsTimeQuotient(that.value.getTime(), 1000);

            if (here !== there) {
                return here < there ? -1 : 1;
            }

            const fine = go2jsTimeNanos(this);
            const coarse = go2jsTimeNanos(that);

            return fine < coarse ? -1 : (fine > coarse ? 1 : 0);
        },
        Format: function(layout) {
            return go2jsTimeFormat(this.value, String(layout), this);
        },
        AppendFormat: function(b, layout) {
            return go2jsAppend(b, ...go2jsStringToBytes(go2jsTimeFormat(this.value, String(layout), this)));
        },
        MarshalText: function() {
            const text = go2jsTimeRFC3339Nano(this);

            if (text === null) {
                return [null, go2jsTimeYearRangeError("MarshalText")];
            }

            return [go2jsStringToBytes(text), null];
        },
        MarshalJSON: function() {
            const text = go2jsTimeRFC3339Nano(this);

            if (text === null) {
                return [null, go2jsTimeYearRangeError("MarshalJSON")];
            }

            return [go2jsStringToBytes("\"" + text + "\""), null];
        },
        UnmarshalText: function(data) {
            return go2jsTimeUnmarshal(this, go2jsRawText(data), false);
        },
        UnmarshalJSON: function(data) {
            return go2jsTimeUnmarshal(this, go2jsRawText(data), true);
        },
        String: function() {
            return go2jsTimeString(this);
        },
        Add: function(duration) {
            const nanos = Number(go2jsDurationNanos(duration));
            const held = this.__go2js_nsec === undefined ? 0 : this.__go2js_nsec;
            // A span is counted in whole milliseconds and in what is left over,
            // and what is left over is counted from the whole millisecond below
            // it rather than from zero, so a span before zero leaves a positive
            // remainder and the two parts always name the same span.
            const remainder = nanos % 1000000;
            const step = (nanos - remainder) / 1000000;
            let fine = held + remainder;
            let millis = this.value.getTime() + step;

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

// go2jsTimeRFC3339Nano is a moment written the way the JSON and text encoders
// write it, which is the layout with the fractional seconds written only as far
// as they go. A year that cannot be written in four digits has no text at all,
// which those encoders report rather than writing.
function go2jsTimeRFC3339Nano(t) {
    const year = t.Year();

    if (year < 0 || year > 9999) {
        return null;
    }

    return go2jsTimeFormat(t.value, "2006-01-02T15:04:05.999999999Z07:00", t);
}

function go2jsTimeYearRangeError(method) {
    return go2jsNameError(new Error("Time." + method + ": year outside of range [0,9999]"), "*errors.errorString");
}

// go2jsTimeUnmarshal reads a moment back out of its written form and gives the
// fields of the receiver the moment that was read. The JSON form arrives with
// the quotes a string value is written with and accepts the null that stands
// for a value that was not there, which leaves the receiver as it was.
function go2jsTimeUnmarshal(receiver, text, json) {
    if (json) {
        if (text === "null") {
            return null;
        }

        if (text.length < 2 || text[0] !== "\"" || text[text.length - 1] !== "\"") {
            return go2jsNameError(new Error("Time.UnmarshalJSON: input is not a JSON string"), "*errors.errorString");
        }

        text = text.slice(1, -1);
    }

    const [parsed, err] = go2jsTimeParse("2006-01-02T15:04:05.999999999Z07:00", text);

    if (err !== null) {
        return err;
    }

    receiver.value = parsed.value;
    receiver.__go2js_nsec = parsed.__go2js_nsec;
    receiver.__go2js_location = parsed.__go2js_location;
    receiver.__go2js_zone = parsed.__go2js_zone;

    return null;
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

    // A zone named by the offset a text carried answers by that offset, while a
    // zone fixed at an offset with a name of its own keeps the name it was
    // given, since both carry their offset as a number.
    if (typeof zone === "number") {
        return {name: name === "" || name === "UTC" ? go2jsTimeFormatZoneCompact(seconds) : name, zulu: zulu, number: number};
    }

    if (name === "UTC") {
        return {name: "UTC", zulu: zulu, number: number};
    }

    if (name === "Local") {
        return {name: go2jsTimeLocalZoneName(date), zulu: zulu, number: number};
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

    // The day of the year is a count of days from the first of January, which
    // the two pieces that ask for it write out.
    const yearDay = Math.floor((Date.UTC(year, month, day) - Date.UTC(year, 0, 1)) / 86400000) + 1;

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
            ["002", pad(yearDay, 3)],
            ["__2", "  ".slice(0, 3 - String(yearDay).length) + yearDay],
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

function go2jsURLQueryEscape(value) {
	return go2jsURLEncode(value, "query");
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

function go2jsJSONDelimText(delim) {
    const character = delim !== null && typeof delim === "object" && delim.__go2js_json_delim !== undefined
        ? delim.__go2js_json_delim
        : "";
    return character === "{" || character === "}" || character === "[" || character === "]" ? character : "";
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

	const bytes = result[0] === null || result[0] === undefined ? [] : go2jsToArray(result[0]);

	this.data = this.data.concat(bytes);

	return [bytes.length, result[1]];
};

function go2jsBytesReadStringResult(value, err) {
	return [value, err];
}

function go2jsEOFError() {
	return go2jsSentinelError("EOF")();
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

const go2jsMethodResults = Object.create(null);

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

// A span of time that was boxed as the type it was declared with holds a whole
// number rather than the shape that carries the method, so the way it writes
// itself is registered under the name of its type as well.
go2jsRegisterMethod("time.Duration.String", function(value) {
	return DurationString(value);
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

const go2jsValueMethodSets = Object.create(null);

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
	// Go writes the empty interface with a space between its keyword and its
	// braces, and a type name that carries one anywhere in it, as a slice or a
	// map of them does, keeps that space in every place it appears.
	return go2jsNormalizeTypeName(go2jsGoTypeNameRaw(value)).replace(/interface\{\}/g, "interface {}");
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

function go2jsRegisterMethod(name, fn, results) {
	go2jsMethodTable[name] = fn;

	// A method reached through reflect has to know how many values it hands
	// back, because a call hands them over as that many, and a method reached by
	// a plain call never asks.
	if (typeof results === "number") {
		go2jsMethodResults[name] = results;
	}
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

// go2jsValueAddress is the address fmt would print for a pointer, a channel or
// a function. None of the three is a value a JavaScript program can name an
// address for, so each is handed a stable number of its own the way a pointer
// is. A nil pointer, channel or function has no place to name and is written as
// the angle-bracket nil for the caller to turn into the zero address.
function go2jsValueAddress(value) {
	if (value !== null && value !== undefined && value.__go2js_pointer === true) {
		return go2jsPointerAddress(value);
	}

	if (typeof value === "function" ||
		(typeof value === "object" && value !== null && value.__go2js_channel === true)) {
		let address = go2jsPointerAddresses.get(value);

		if (address === undefined) {
			go2jsPointerAddressCount++;
			address = go2jsPointerAddressCount;
			go2jsPointerAddresses.set(value, address);
		}

		return "0x" + (0xc000000000 + address * 8).toString(16).padStart(12, "0");
	}

	return "<nil>";
}

// go2jsIsAddressOnlyValue reports whether a value is a channel or a function,
// which Go writes as an address rather than as the value behind it.
function go2jsIsAddressOnlyValue(value) {
	if (typeof value === "function") {
		return true;
	}

	return typeof value === "object" && value !== null && value.__go2js_channel === true;
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

function go2jsInterface(value, typeName, displayName, uncomparable) {
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

	if (uncomparable === true) {
		// A slice or a map or a function is never allowed to compare itself or
		// to hash itself, so the wrapper says it cannot so the equal and the
		// map key catch hold of it before a value is put side by side.
		wrapper.__go2js_uncomparable = true;
	}

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
	"Error": ["Error", "RuntimeError"],
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

// go2jsAnonymousInterfaceMethods reads the method names out of an interface
// written in place, as interface{M() int; N(x int)} is, which has no name to
// have been registered under.
function go2jsAnonymousInterfaceMethods(name) {
	const match = typeof name === "string" ? /^interface ?\{(.*)\}$/s.exec(name.trim()) : null;

	if (match === null || match[1].trim() === "") {
		return undefined;
	}

	const methods = [];
	let depth = 0;
	let start = 0;
	const body = match[1];

	const take = function(end) {
		const part = /^\s*([A-Za-z_][A-Za-z0-9_]*)\s*\(/.exec(body.slice(start, end));

		if (part !== null) {
			methods.push(part[1]);
		}
	};

	for (let i = 0; i < body.length; i++) {
		const ch = body[i];

		if (ch === "(" || ch === "{" || ch === "[") {
			depth++;
		} else if (ch === ")" || ch === "}" || ch === "]") {
			depth--;
		} else if (ch === ";" && depth === 0) {
			take(i);
			start = i + 1;
		}
	}

	take(body.length);

	return methods;
}

function go2jsInterfaceMethods(name) {
	const anonymous = go2jsAnonymousInterfaceMethods(name);

	if (anonymous !== undefined) {
		return anonymous;
	}

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
	// A slice or an array is the commonest thing asked. A slice keeps its length
	// in its state, which is a cheaper place to read it than through the proxy
	// that stands for it.
	if (typeof value === "object" && value !== null) {
		const state = go2jsSliceMeta.get(value);

		if (state !== undefined) {
			return state.length;
		}

		if (Array.isArray(value)) {
			return value.length;
		}
	}
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

function go2jsAppend(value, ...items) {
		if (value === null || value === undefined) {
			const result = [...items];
			Object.defineProperty(result, "__go2js_cap", {
				value: result.length,
				writable: true,
				configurable: true
			});
			return result;
		}
		if (!Array.isArray(value)) {
			throw new TypeError("go2jsAppend expects an array");
		}

		const result = value.concat(items);
		const required = result.length;
		const previousCap = value.__go2js_cap ?? value.length;
		const capacity = required <= previousCap ? previousCap : required;

		Object.defineProperty(result, "__go2js_cap", {
			value: capacity,
			writable: true,
			configurable: true
		});

		return result;
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

	// A value held in an interface is found by the type it holds and the value
	// it holds, which is how Go weighs one interface against another. A pointer
	// held in an interface is still told apart by which pointer it is, since two
	// pointers to equal values are two pointers.
	if (value.__go2js_interface === true) {
		const held = value.value === value ? value : value.value;

		if (held !== null && typeof held === "object" && held.__go2js_pointer === true) {
			return null;
		}

		return "iface:" + String(value.type) + ":" + go2jsMapKeyPart(held);
	}

	if (value.__go2js_pointer === true ||
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
		go2jsPanic(go2jsRuntimeFault("assignment to entry in nil map", "runtime.plainError"));
	}

	if (map.__go2js_nil === true) {
		go2jsPanic(go2jsRuntimeFault("assignment to entry in nil map", "runtime.plainError"));
	}

	if (!(map instanceof go2jsNativeMap)) {
		throw new TypeError("go2jsMapSet expects a Map");
	}
	map.set(go2jsMapKey(key), value);
}

// go2jsMapUpdate works an operation out on the entry a compound assignment is
// aimed at, because a lookup is a value JavaScript will not let an assignment
// be written to. A key that is not there yet reads as the element type's zero
// value, and the result is written back under the same key.
function go2jsMapUpdate(map, key, zero, operation) {
	if (map === null || map === undefined || map.__go2js_nil === true) {
		go2jsPanic(go2jsRuntimeFault("assignment to entry in nil map", "runtime.plainError"));
	}

	if (!(map instanceof go2jsNativeMap)) {
		throw new TypeError("go2jsMapUpdate expects a Map");
	}

	const k = go2jsMapKey(key);
	const current = map.has(k) ? map.get(k) : zero;
	map.set(k, operation(current));
}

function go2jsMapDelete(map, key) {
	if (map === null || map === undefined) {
		return;
	}

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

function go2jsIndex(value, index) {
	const state = go2jsSliceMeta.get(value);

	if (state !== undefined) {
		const position = Math.trunc(Number(index));

		if (position < 0) {
			throw new go2jsNativeRangeError(go2jsRuntimeErrorPrefix + "index out of range [" + position + "]");
		}

		if (position >= state.length) {
			throw new go2jsNativeRangeError(go2jsRuntimeErrorPrefix + "index out of range [" + position + "] with length " + state.length);
		}

		return state.data[state.offset + position];
	}

	const position = Math.trunc(Number(index));

	if (position < 0) {
		throw new go2jsNativeRangeError(go2jsRuntimeErrorPrefix + "index out of range [" + position + "]");
	}

	const length = go2jsLen(value);

	if (position >= length) {
		throw new go2jsNativeRangeError(go2jsRuntimeErrorPrefix + "index out of range [" + position + "] with length " + length);
	}

	return value[position];
}

function go2jsIndexCheck(value, index) {
	const state = go2jsSliceMeta.get(value);

	// The bounds of a slice are all that is asked, so it is not read from.
	if (state !== undefined && typeof index === "number" && index >= 0 && index < state.length) {
		return;
	}

	go2jsIndex(value, index);
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

// go2jsWideBits is the width a whole number of each narrow type is held in. A
// type not named here is held in the sixty four bits every other whole number
// type uses.
const go2jsWideBits = { "uint8": 8, byte: 8, "int8": 8, "uint16": 16, "int16": 16, "uint32": 32, "int32": 32, rune: 32 };

// go2jsWideUnsigned names the whole number types that never go below zero, since
// an answer below zero is given back as the largest number the width holds
// rather than as the negative number it was worked out as.
const go2jsWideUnsigned = {
	uint64: true, uint: true, uintptr: true, "uint8": true, byte: true, "uint16": true, "uint32": true
};

// go2jsWideWrap keeps a whole number inside the sixty four bits a signed or an
// unsigned number of that width is held in, which is what a Go variable of that
// type does without being asked. An answer too wide for its type is not a fault
// in Go, it is the number that fits, so a signed one past its largest is the
// smallest it has and an unsigned one below its smallest is the largest it has.
function go2jsWideWrap(value, typeName) {
	if (typeof value !== "bigint") {
		const number = Math.trunc(Number(value));

		if (Number.isInteger(number)) {
			// a double past the point where it holds every digit is worked out
			// as digits before it is given a width, since the digits it is
			// carrying are not the ones it was handed
			if (number > go2jsSafeInteger || number < -go2jsSafeInteger) {
				return go2jsWideWrap(BigInt(number), typeName);
			}

			// a number below zero that belongs to a type which never goes below
			// zero is the largest that type holds, which is wider than a double
			// can keep, so it is worked out as digits as well
			if (number < 0 && go2jsWideUnsigned[typeName] === true) {
				return go2jsWideWrap(BigInt(number), typeName);
			}
		}

		return number;
	}

	const unsigned = go2jsWideUnsigned[typeName] === true;
	const width = go2jsWideBits[typeName] === undefined ? 64 : go2jsWideBits[typeName];

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

// The operations a wide number is worked out with are named once here rather
// than written where they are used, so that adding two numbers does not make a
// new function to add them with each time.
function go2jsWidePlus(a, b) {
	return a + b;
}

function go2jsWideMinus(a, b) {
	return a - b;
}

function go2jsWideDivideWhole(a, b) {
	if (b === 0) {
		throw new go2jsNativeRangeError(go2jsRuntimeErrorPrefix + "integer divide by zero");
	}

	return Math.trunc(a / b);
}

function go2jsWideDivideBig(a, b) {
	if (b === 0n) {
		throw new go2jsNativeRangeError(go2jsRuntimeErrorPrefix + "integer divide by zero");
	}

	return a / b;
}

function go2jsWideAdd(left, right, typeName) {
	// Two plain numbers whose answer is a whole number that fits thirty-two bits
	// need none of what follows, and are by far the commonest case, so they are
	// answered first.
	if (typeof left === "number" && typeof right === "number") {
		const fast = left + right;

		if ((fast | 0) === fast && (fast >= 0 || go2jsWideUnsigned[typeName] !== true)) {
			return fast;
		}
	}

	return go2jsWideBinary(left, right, go2jsWidePlus, go2jsWidePlus, typeName);
}

// go2jsWideBinary keeps a whole number operation exact all the way out, which
// the plain operator cannot be once the answer is past the digits a double
// keeps. It is tried in doubles first, and only an answer that double cannot
// hold is worked out as digits, so a program that never leaves the numbers a
// double covers is not slowed down by the wider kind being there.
function go2jsWideBinary(left, right, whole, big, typeName) {
	left = go2jsDurationOperand(left);
	right = go2jsDurationOperand(right);

	if (go2jsIsWide(left) || go2jsIsWide(right)) {
		return go2jsWideWrap(big(go2jsWide(left), go2jsWide(right)), typeName);
	}

	const result = whole(left, right);

	if (Number.isSafeInteger(result)) {
		return go2jsWideWrap(result, typeName);
	}

	return go2jsWideWrap(big(BigInt(Math.trunc(left)), BigInt(Math.trunc(right))), typeName);
}

function go2jsWideSub(left, right, typeName) {
	// Two plain numbers whose answer is a whole number that fits thirty-two bits
	// need none of what follows, and are by far the commonest case, so they are
	// answered first.
	if (typeof left === "number" && typeof right === "number") {
		const fast = left - right;

		if ((fast | 0) === fast && (fast >= 0 || go2jsWideUnsigned[typeName] !== true)) {
			return fast;
		}
	}

	return go2jsWideBinary(left, right, go2jsWideMinus, go2jsWideMinus, typeName);
}

function go2jsWideQuo(left, right, typeName) {
	if (typeof left === "number" && typeof right === "number" && right !== 0) {
		const fast = Math.trunc(left / right);

		if ((fast | 0) === fast && (fast >= 0 || go2jsWideUnsigned[typeName] !== true)) {
			return fast;
		}
	}

	return go2jsWideBinary(left, right, go2jsWideDivideWhole, go2jsWideDivideBig, typeName);
}

// go2jsWideFloat turns a whole number into the nearest number a double holds,
function go2jsWideFloat(value) {
	value = go2jsDurationOperand(value);

	// reading a whole number as a double gives the nearest double to it, which is
	// the number the conversion gives in Go, digits and all the way out. A double
	// carries about sixteen digits exactly, so past that the answer is the nearest
	// to the whole number rather than the whole number itself, and that is what a
	// program asking for a float64 is asking for.
	return Number(value);
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

function go2jsDivide(left, right, typeName) {
	// a span of time is asked for the whole number it holds before the two are
	// divided, since it is that number and not the shape that does the work
	left = go2jsDurationOperand(left);
	right = go2jsDurationOperand(right);

	if (go2jsIsWide(left) || go2jsIsWide(right)) {
		return go2jsWideQuo(left, right, typeName);
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
		return go2jsRuntimeFault(reworded);
	}

	if (value instanceof go2jsNativeTypeError && /^interface conversion: /.test(String(value.message))) {
		return go2jsRuntimeFault(String(value.message).replace(/interface\{\}/g, "interface {}"), "*runtime.TypeAssertionError");
	}

	return value;
}

// go2jsRuntimeFault is what a program that recovers from a fault of the runtime
// is given: a value that is an error and a runtime.Error both, with the words Go
// uses for it.
function go2jsRuntimeFault(text, typeName) {
	const fault = new Error(text);

	go2jsNameError(fault, typeName !== undefined ? typeName : /^runtime error: (index|slice bounds)/.test(text) ? "runtime.boundsError" : "runtime.errorString");

	fault.RuntimeError = function() {};
	fault.Error = function() { return text; };

	return fault;
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
// go2jsStrCompare orders two strings the way Go orders them, which is over the
// bytes each is made of. JavaScript orders strings over the sixteen bit units it
// keeps them in, and a character outside the basic plane takes two of those
// units, so an astronaut emoji sorts before a character of the basic plane that
// Go puts it after. Reading both strings back as the bytes they hold puts the
// two orders back together.
function go2jsStrCompare(a, b) {
	const left = go2jsStringToBytes(a);
	const right = go2jsStringToBytes(b);
	const common = Math.min(left.length, right.length);

	for (let i = 0; i < common; i++) {
		if (left[i] !== right[i]) {
			return left[i] < right[i] ? -1 : 1;
		}
	}

	return left.length < right.length ? -1 : left.length > right.length ? 1 : 0;
}

function go2jsCompareValues(a, b) {
	if (typeof a === "number" && typeof b === "number") {
		return a < b ? -1 : a > b ? 1 : 0;
	}

	if (typeof a === "string" && typeof b === "string") {
		return go2jsStrCompare(a, b);
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

// go2jsFormatAddressLike says whether a nil of the named type reads as an
// address rather than as nothing at all, which is what a pointer, a channel and
// a function do: the verbs that count an address read them as zero, and only
// the rest see a nil.
function go2jsFormatAddressLike(typeName, shape) {
	if (shape === "pointer" || shape === "chan" || shape === "func") {
		return true;
	}

	if (typeof typeName !== "string") {
		return false;
	}

	return typeName.startsWith("*") || typeName.startsWith("chan ") ||
		typeName.startsWith("chan<- ") || typeName.startsWith("<-chan ") ||
		typeName.startsWith("func(");
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
			// The method is registered under the bare name of its type, and the
			// type of an element is written with the package it is from.
			const named = go2jsMethodTable[typeName + ".String"] || go2jsMethodTable[typeName.slice(typeName.lastIndexOf(".") + 1) + ".String"];

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

		// The type the box names is the type of what it holds, so it travels
		// with the value: a float held in an interface is still a float, and
		// the verbs must write it as one rather than as a whole number.
		return go2jsFormat(value.value, typeof value.type === "string" && value.type !== "" ? value.type : null, null, null, plus, nested, raw);
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

		// The type a map was written as says what it holds, so a float held in
		// one is written as a float rather than as the whole number it makes.
		const mapType = go2jsDeclaredTypeName(typeName) !== "" ? typeName : (value.__go2js_type || go2jsGoTypeName(value));
		const held = go2jsMapTypes(mapType);
		const parts = entries.map(([key, item]) => go2jsFormat(key, held.key || null, null, null, plus, true) + ":" + go2jsFormat(item, held.value || null, null, null, plus, true));

		return "map[" + parts.join(" ") + "]";
	}

	if (Array.isArray(value) || value instanceof Uint8Array) {
		const elementType = go2jsElementTypeName(go2jsDeclaredTypeName(typeName) !== "" ? typeName : (value.__go2js_type || go2jsGoTypeName(value)));
		const parts = [];

		for (let i = 0; i < value.length; i++) {
			parts.push(go2jsFormat(value[i], elementType || null, null, null, plus, true));
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

	// The declared type of each field travels with the value, so a nil slice
	// reads as [] and a float keeps its exponent, the way fmt reads a field by
	// the type it was declared with rather than by what it holds. A struct
	// written out in place carries the types on itself, having no class to.
	const goName = ctor && typeof ctor.name === "string" ? go2jsTypeNamesByConstructor.get(ctor) : undefined;
	let descriptors = goName !== undefined ? go2jsStructFields[goName] : undefined;

	if (descriptors === undefined && Array.isArray(value.__go2js_struct_fields)) {
		descriptors = value.__go2js_struct_fields;
	}

	const fieldTypes = Object.create(null);
	const fieldKinds = Object.create(null);

	if (descriptors !== undefined) {
		for (const descriptor of descriptors) {
			if (descriptor && typeof descriptor.Name === "string") {
				fieldTypes[descriptor.Name] = descriptor.Type || "";
				fieldKinds[descriptor.Name] = descriptor.Kind || "";
			}
		}
	}

	for (const key of Object.keys(value)) {
		if (fields && typeof fields[key] === "string") {
			const formatter = go2jsMethodTable[fields[key]];

			if (typeof formatter === "function") {
				parts.push(go2jsCallNow(formatter, null, [value[key]]));
				continue;
			}
		}

		const fieldType = fieldTypes[key] || "";
		const fieldKind = fieldKinds[key] || "";
		const held = fieldType !== "" && value[key] === null ? go2jsTyped(null, fieldType, fieldKind, fieldKind) : value[key];
		const text = go2jsFormat(held, fieldType !== "" ? fieldType : null, fieldKind !== "" ? fieldKind : null, fieldKind !== "" ? fieldKind : null, plus, true);

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

// go2jsSpreadValues returns the elements of a spread operand as they travel
// into the operand list, keeping the box each interface element was dressed
// in, because the box is where the type the verbs need to know is kept.
function go2jsSpreadValues(slice) {
	if (!Array.isArray(slice)) {
		return [slice];
	}

	return slice;
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

// go2jsNilValue gives a nil slice or a nil map a value of its own. An
// interface carries the type of what it holds, so a nil that has travelled
// through one still has to print as [] or map[] rather than <nil>. A plain
// array or Map keeps every slice and map operation working: both are empty,
// and writing through them throws first.
function go2jsNilValue(typeName, shape) {
	const value = shape === "map" ? new go2jsNativeMap() : [];

	value.__go2js_nil = true;
	value.__go2js_typed = true;
	value.type = typeName;
	value.shape = shape;
	// The value is its own wrapper, so anything that follows a wrapper through
	// to its value lands back on the slice or map rather than on nothing.
	value.value = value;

	return value;
}

// go2jsNilInterface gives the type to a nil slice or a nil map on its way into
// an interface, and keeps the name for a slice or a map that is really there so
// that %T can report it. A bare array or Map cannot recover its element type on
// its own, so a non-nil value is boxed the same way an interface{} value is.
function go2jsNilInterface(value, typeName, shape, uncomparable) {
	if (value === null || value === undefined) {
		return go2jsNilValue(typeName, shape);
	}

	// A nil that already has a value of its own is named here, because this is
	// the point that knows what the program called it.
	if (typeof value === "object" && value.__go2js_nil === true) {
		if (typeof value.type !== "string" || value.type === "") {
			value.type = typeName;
		}

		return value;
	}

	if (typeof value === "object" && (value.__go2js_interface === true || value.__go2js_typed === true)) {
		return value;
	}

	return go2jsInterface(value, typeName, typeName, uncomparable);
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
// a space between them, the way Go writes "% x" over a string or a byte slice,
// and the sharp flag gives every byte its own 0x when a space is there too but
// the whole run a single 0x otherwise, which is how "% #x" and "%#x" read.
function go2jsFormatHexBytes(value, upper, space, alt, precision) {
	const parts = [];
	// A precision cuts the bytes to that many before they are written, the way
	// fmt stops reading a string once its count is reached.
	const items = precision !== null && precision !== undefined && precision < go2jsToArray(value).length
		? go2jsToArray(value).slice(0, precision)
		: go2jsToArray(value);

	for (const item of items) {
		parts.push((Number(item) & 255).toString(16).padStart(2, "0"));
	}

	if (alt === true) {
		const prefix = upper ? "0X" : "0x";

		// A sharpened hex reads as one run of digits with a single 0x in front,
		// until the space flag asks for a gap between every byte, when each byte
		// takes its own 0x so the runs they make stay whole.
		const text = space
			? parts.map((part) => prefix + part).join(" ")
			: prefix + parts.join("");

		return upper ? text.toUpperCase() : text;
	}

	const text = parts.join(space ? " " : "");

	return upper ? text.toUpperCase() : text;
}

function go2jsIsByteArray(value) {
	// A slice of bytes the runtime holds is a typed array, which is a byte slice
	// however it is reached, even when it is empty.
	if (value instanceof Uint8Array) {
		return true;
	}

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

// go2jsTypeNeedsParens reports whether a type name is a composite type that Go
// wraps in parentheses when it names the nothing of the type, as in (*T)(nil)
// or (chan int)(nil). A plain name needs none: T(nil) is written as it stands.
function go2jsTypeNeedsParens(name) {
	const prefixes = ["*", "chan ", "chan<- ", "<-chan ", "func("];

	return prefixes.some((prefix) => name.startsWith(prefix));
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

		// A slice passed through an interface wears the box on the list itself,
		// so stepping down to what it holds lands on the very same list again.
		// The list stands for itself, so it is written rather than unwrapped.
		if (held !== value) {
			return go2jsGoSyntax(held, typeName);
		}
	}

	// A value with a GoString of its own writes itself.
	if (value !== null && value !== undefined && typeof value === "object" &&
		!Array.isArray(value) && !(value instanceof go2jsNativeMap)) {
		const receiver = value.__go2js_pointer === true ? go2jsDeref(value) : value;

		if (receiver !== null && receiver !== undefined && typeof receiver === "object") {
			const goString = go2jsNamedFormatMethod(receiver, "GoString");

			if (goString !== null) {
				return goString;
			}
		}
	}

	if (value === null || value === undefined) {
		// A value held in an interface has no value of its own when the value it
		// holds is nothing, so what is written is the type it was held as rather
		// than the nothing itself. An empty interface names itself, since it is
		// the type of the slot the nothing was found in.
		const declared = go2jsDeclaredTypeName(typeName);

		if (declared === "" && (typeName === "interface{}" || typeName === "interface {}" || typeName === "any")) {
			return "interface {}(nil)";
		}

		if (declared === "") {
			return "<nil>";
		}

		// Go wraps the type of a pointer or another compound type in
		// parentheses when it writes the nothing it points at, so a nil *T
		// reads (*T)(nil) and a nil channel reads (chan int)(nil).
		return go2jsTypeNeedsParens(declared) ? "(" + declared + ")(nil)" : declared + "(nil)";
	}

	if (value.__go2js_pointer === true) {
		// A pointer to a record is written as the record it aims at with the
		// ampersand in front, since the type is named there; a pointer to
		// anything else has nowhere but the parentheses to show its type, so
		// its address stands where the value would.
		if (go2jsPointerTargets(value)) {
			const targetType = go2jsDeclaredTypeName(typeName) !== "" ? typeName.replace(/^\*/, "") : "";

			return "&" + go2jsGoSyntax(go2jsDeref(value), targetType);
		}

		return "(" + go2jsGoTypeName(value) + ")(" + go2jsPointerAddress(value) + ")";
	}

	if (go2jsIsAddressOnlyValue(value)) {
		// A channel or a function has no fields to write, so Go shows the type
		// and the address the value sits at.
		const addressName = go2jsDeclaredTypeName(typeName) !== "" ? typeName : go2jsGoTypeName(value);

		return "(" + addressName + ")(" + go2jsValueAddress(value) + ")";
	}

	if (value instanceof Error) {
		return value.message;
	}

	if (typeof value === "string") {
		return JSON.stringify(value);
	}

	if (typeof value === "number") {
		// %#v writes a float the way %v does, so a float held in an interface
		// is written with the shortest digits that read back as it rather than
		// as the whole number its digits happen to make.
		if (go2jsIsFloatTypeName(typeName)) {
			return go2jsFormatFloatType(value, typeName, typeName);
		}

		return String(value);
	}

	if (typeof value === "boolean") {
		return String(value);
	}

	if (value instanceof go2jsNativeMap) {
		const held = go2jsMapTypes(go2jsDeclaredTypeName(typeName) !== "" ? typeName : (value.__go2js_type || go2jsGoTypeName(value)));
		const entries = Array.from(value.entries());

		entries.sort((a, b) => go2jsCompareValues(a[0], b[0]));

		// The name of a map already says what it holds keys and values of, so it
		// is written out as it stands rather than wrapped in brackets again, and
		// each key and value is written as the type the map holds it as.
		return go2jsCollectionTypeName(value, typeName) +
			"{" + entries.map(([key, item]) => go2jsGoSyntax(key, held.key) + ":" + go2jsGoSyntax(item, held.value)).join(", ") + "}";
	}

	if (Array.isArray(value)) {
		const elementType = go2jsElementTypeName(go2jsDeclaredTypeName(typeName) !== "" ? typeName : (value.__go2js_type || go2jsGoTypeName(value)));

		return go2jsCollectionTypeName(value, typeName) +
			"{" + value.map(item => go2jsGoSyntax(item, elementType)).join(", ") + "}";
	}

	const ctor = value.constructor;
	const goName = ctor && go2jsTypeNamesByConstructor.get(ctor);
	let descriptors = goName !== undefined ? go2jsStructFields[goName] : undefined;

	// A struct written out in place has no class of its own, so the name and
	// the type of each field travel on the value itself.
	if (descriptors === undefined && Array.isArray(value.__go2js_struct_fields)) {
		descriptors = value.__go2js_struct_fields;
	}

	const parts = [];

	for (const key of Object.keys(value)) {
		let fieldType = "";

		if (descriptors !== undefined) {
			for (const field of descriptors) {
				if (field.Name === key) {
					fieldType = field.Type || "";

					break;
				}
			}
		}

		parts.push(key + ":" + go2jsGoSyntax(value[key], fieldType));
	}

	return go2jsQualifiedTypeName(go2jsGoTypeName(value)) + "{" + parts.join(", ") + "}";
}

// go2jsMapTypes gives the name of the key and the name of the value a map type
// holds, so that a number kept inside one is written as the kind its type says
// rather than as the digits it happens to make.
function go2jsMapTypes(name) {
	const declared = go2jsDeclaredTypeName(name);

	if (!declared.startsWith("map[")) {
		return { key: "", value: "" };
	}

	let depth = 0;
	let i = 3;

	for (; i < declared.length; i++) {
		if (declared[i] === "[") {
			depth += 1;
		} else if (declared[i] === "]") {
			depth -= 1;

			if (depth === 0) {
				i += 1;

				break;
			}
		}
	}

	return { key: declared.slice(4, i - 1), value: declared.slice(i) };
}

// go2jsElementTypeName gives the name of the element a slice or an array holds.
function go2jsElementTypeName(name) {
	const declared = go2jsDeclaredTypeName(name);

	if (declared[0] !== "[") {
		return "";
	}

	const close = declared.indexOf("]");

	return close === -1 ? "" : declared.slice(close + 1);
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

	if (name === "" || name === "interface{}" || name === "interface {}" ||
		name === "any" || name === "null" ||
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
	// A number that does not name a rune is quoted as the replacement character,
	// which is what Go reads an invalid rune as before it writes the literal.
	if (code < 0 || code > 0x10FFFF || (code >= 0xD800 && code <= 0xDFFF)) {
		code = 0xFFFD;
	}

	if (!go2jsRuneIsPrintable(code)) {
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

	// fmt.Formatter's Format method takes the state and the verb, and it is the
	// only Format method fmt hands a verb to. A type such as time.Time carries a
	// Format method of its own that takes a layout string instead, and handing
	// the verb to that one writes nothing to the state, which would turn %v into
	// an empty string. Only a method shaped like fmt.Formatter's is used.
	if (format !== null && typeof receiver.Format === "function" && receiver.Format.length < 2) {
		format = null;
	}

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

// go2jsFormatPointerValue writes a pointer that has no method of its own. The
// verbs that read a whole number write the address itself, with a pointer to
// nothing written as the zero address, and the quoted verb reaches through a
// pointer to a record the way the other verbs' walk already does, with the
// ampersand in front. Every other verb refuses the pointer, naming its type
// beside the address, which is what fmt writes for an operand a verb cannot
// read.
function go2jsFormatPointerValue(verb, parsed, value, name, flags, precision) {
	const address = go2jsValueAddress(value);
	const nil = address === "<nil>";

	if (go2jsVerbInSet(verb, "dboxX")) {
		const number = nil ? 0 : parseInt(address.slice(2), 16);
		const spec = "%" + flags + (precision === null ? "" : "." + precision) + verb;

		return go2jsFormatValue(verb, spec, number);
	}

	if (verb === "q" && !nil && go2jsPointerTargets(value)) {
		const spec = "%" + flags + (parsed.width > 0 ? String(parsed.width) : "") +
			(precision === null ? "" : "." + precision) + "q";

		return "&" + go2jsFormatValue("q", spec, go2jsDeref(value));
	}

	return "%!" + verb + "(" + name + "=" + address + ")";
}

// go2jsFormatValue writes one operand with a verb. The raw flag says the operand
// is being shown as the rejected operand of a verb it does not take, and fmt
// shows such an operand as it is rather than by asking it to write itself. The
// nested flag says the operand sits inside another one, where a nil is written
// as the angle-bracket nil instead of refusing the verb the way a whole operand
// does.
function go2jsFormatValue(verb, spec, value, raw, nested) {
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

	// A nil sitting inside a compound is written as the angle-bracket nil for
	// whatever verb reads it, because Go has no value there to refuse the verb
	// against, and a nil that is the whole operand is the one case fmt does
	// refuse it for. A nil slice or map keeps the empty literal its shape names.
	// A pointer, a channel or a function is left to the cases below, which read
	// the nil they hold.
	if (nested === true && (value === null || value === undefined) && !go2jsFormatAddressLike(tagName, shape)) {
		return go2jsPad(go2jsNilFormat(tagName, kind, shape), parsed, false);
	}

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
		// for it. The type named is the one the value was declared as, so a
		// named kind reads as its own name rather than as the kind under it. A
		// rune is named int32, because that is the type fmt reflects on.
		const refusalName = go2jsDeclaredTypeName(tagName) !== "" ? go2jsDeclaredTypeName(tagName) : accepted;

		return "%!" + verb + "(" + (refusalName === "rune" ? "int32" : refusalName) + "=" +
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
	// %#v asks for Go syntax, which a String method has no say in.
	if (raw !== true && go2jsVerbInSet(verb, "vsxXq") && !(verb === "v" && flags.includes("#"))) {
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

		// A precision of zero asked of a whole zero writes nothing at all, only
		// the padding the width asks for, and the zero flag has nothing to fill.
		if (isInteger && precision === 0 && (num === 0 || Object.is(num, -0))) {
			return go2jsPad("", {...parsed, flags: parsed.flags.replace(/0/g, "")}, false);
		}

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

	// A channel or a function is an address to Go rather than a value with
	// parts, so it is answered here, before the verbs that walk a compound,
	// which would otherwise read the fields a channel happens to hold in
	// JavaScript. A nil channel or function reads as the zero address, and a
	// function or channel with no address counts one too. A pointer is left
	// below, where a pointer to a record is written as the record it aims at.
	if (go2jsIsAddressOnlyValue(value) && verb !== "T") {
		const chanOrFuncName = typeof accepted === "string" &&
			(accepted[0] === "*" || accepted.startsWith("chan ") || accepted.startsWith("<-chan ") ||
				accepted.startsWith("chan<- ") || accepted.startsWith("func("));
		const addressName = chanOrFuncName && go2jsDeclaredTypeName(accepted) !== ""
			? accepted
			: go2jsGoTypeName(value);

		if (verb === "v" || verb === "w") {
			if (!flags.includes("#")) {
				return go2jsPad(go2jsValueAddress(value), parsed, false);
			}
		} else if (verb === "p") {
			return go2jsPad(go2jsValueAddress(value), parsed, false);
		} else {
			return go2jsFormatPointerValue(verb, parsed, value, addressName, flags, precision);
		}
	}

	// A slice or an array applies the verb to each element and joins the results,
	// so it is answered before the verb's own case sees the whole value. A byte
	// slice is the exception: the quoted verbs and the hex verbs read it as a
	// string rather than as a list of numbers.
	// slice is the exception: the quoted verbs and the hex verbs read it as a
	// string rather than as a list of numbers.
	const isBytes = go2jsIsByteCompound(accepted);
	// The sharp flag with %v asks for the value as Go writes it, which is the
	// whole of it with its type in front rather than the elements one by one.
	// A nil slice written with a number verb is the empty list, whatever its
	// elements would have been.
	if ((value === null || value === undefined) && typeof accepted === "string" && accepted.startsWith("[]") &&
		go2jsVerbInSet(verb, "dxXobceEfFgG") && !isBytes) {
		return go2jsPad("[]", parsed, false);
	}

	const compound = go2jsFormatCompound(value, accepted, verb, flags, parsed.width, precision, isBytes);

	if (compound !== null) {
		return compound;
	}

	// A complex operand formats each of its parts with the verb and joins them
	// with a plus, the way fmt writes (re+imi), so it is answered first.
	if (go2jsIsComplexOperand(value, accepted)) {
		return go2jsFormatComplex(verb, parsed, value, accepted, flags, precision);
	}

	// A pointer, a slice, a map or a channel with nothing behind it has no
	// address to reach, and Go writes the zero address for it rather than the
	// word it uses for a value that is not there at all.
	if (verb === "p") {
		// A pointer is answered with the address it was handed, which stays the
		// same for the life of the value and has the shape of a Go address even
		// though the digits themselves cannot be Go's.
		if (go2jsIsAddressOnlyValue(value) ||
			(value !== null && value !== undefined && value.__go2js_pointer === true)) {
			return go2jsPad(go2jsValueAddress(value), parsed, false);
		}

		const empty = value === null || value === undefined ||
			(typeof value === "object" &&
				(value.__go2js_nil === true || value.__go2js_typed_nil === true));

		if (empty) {
			return go2jsPad("0x0", parsed, false);
		}
	}

	// A pointer with no method of its own is written by fmt as an address: the
	// verbs that read a whole number write the address itself, and every other
	// verb refuses the pointer naming its type. A pointer to nothing is the
	// zero address, and a pointer to a record is reached through the record it
	// aims at, which the walk above has already done. The plain, the signed,
	// the address and the type verbs keep their own paths.
	// A pointer, a channel and a function all carry no value of their own, so a
	// nil one of any of them is named by its type and read by the verb the way
	// fmt reads it: an address under the verbs that count one, a refusal under
	// the rest.
	const chanOrFuncName = typeof accepted === "string" &&
		(accepted[0] === "*" || accepted.startsWith("chan ") || accepted.startsWith("<-chan ") ||
			accepted.startsWith("chan<- ") || accepted.startsWith("func("));

	const addressable = value !== null && value !== undefined && value.__go2js_pointer === true
		? (go2jsDeclaredTypeName(accepted) !== "" ? accepted : go2jsGoTypeName(value))
		: (value === null || value === undefined) && chanOrFuncName
			? accepted
			: null;

	if (addressable !== null && verb !== "v" && verb !== "w" && verb !== "p" && verb !== "T") {
		return go2jsFormatPointerValue(verb, parsed, value, addressable, flags, precision);
	}

	switch (verb) {
		case "d":
			return numberText(intValue, String(Math.abs(intValue)));
		case "b":
			// A float formatted with %b is its IEEE754 expansion: the leading
			// bit of the significand, then the exponent as a power of two.
			if (accepted !== undefined && go2jsIsFloatTypeName(accepted)) {
				// The expansion carries its own minus for a negative value, so
				// the sign and the width are added around it the way they are
				// for a hex float, which also brings its own sign.
				return numberText(intValue, go2jsFormatFloatBits(Number(value), accepted), "", true);
			}

			return integerBody(intValue, 2, flags.includes("#") ? "0b" : "");
		case "o":
		case "O": {
			// The sharp flag puts a zero in front of the octal digits, but only
			// when the digits, cut to the precision that broadens them, do not
			// already start with one. %O always carries the 0o prefix, while %o
			// only does with #.
			const digits = Math.abs(Math.trunc(intValue)).toString(8);
			const broadened = precision !== null && precision > digits.length;
			const sharp = flags.includes("#") && intValue !== 0 && !broadened ? "0" : "";

			return integerBody(intValue, 8, (verb === "O" ? "0o" : "") + sharp);
		}
		case "x":
			if (accepted !== undefined && go2jsIsFloatTypeName(accepted)) {
				return numberText(intValue, go2jsFormatHexFloat(Number(value), precision, false, flags.includes("#")), "", true);
			}

			// A byte compound reads as a string of hex digits, and it is the
			// one compound whose hex verbs are answered whole rather than
			// element by element. A single byte is an array like any other, so
			// the reading is the shape of the value rather than a number of
			// it.
			if (Array.isArray(value) || value instanceof Uint8Array) {
				return go2jsPad(go2jsFormatHexBytes(value, false, flags.includes(" "), flags.includes("#"), precision), parsed, false);
			}

			if (typeof value === "string" || typeof value === "boolean") {
				return go2jsPad(go2jsFormatHexBytes(go2jsStringToBytes(value), false, flags.includes(" "), flags.includes("#"), precision), parsed, false);
			}

			return integerBody(intValue, 16, flags.includes("#") ? "0x" : "", false);
		case "X":
			if (accepted !== undefined && go2jsIsFloatTypeName(accepted)) {
				return numberText(intValue, go2jsFormatHexFloat(Number(value), precision, true, flags.includes("#")), "", true);
			}

			if (Array.isArray(value) || value instanceof Uint8Array) {
				return go2jsPad(go2jsFormatHexBytes(value, true, flags.includes(" "), flags.includes("#")), parsed, false);
			}

			if (typeof value === "string" || typeof value === "boolean") {
				return go2jsPad(go2jsFormatHexBytes(go2jsStringToBytes(value), true, flags.includes(" "), flags.includes("#"), precision), parsed, false);
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

			if (go2jsIsByteCompound(tagged) || go2jsIsByteCompound(kind) || (tagged === null && go2jsIsByteArray(value))) {
				text = go2jsBytesToString(value);
			} else {
				text = go2jsFormat(value, tagged, kind, shape);
			}

			if (precision !== null) {
				text = go2jsTruncateRunes(text, precision);
			}

			return go2jsPad(text, parsed, false);
		}
		case "v":
		case "w": {
			// A float under %v with a precision is written the way %g writes it
			// with that precision, which is the shortest form only when no
			// precision was asked for. The sharp flag keeps its own go-syntax
			// reading below.
			if (precision !== null && !flags.includes("#") && go2jsIsFloatTypeName(accepted)) {
				// %v keeps the space flag for a number but not the plus flag,
				// which under %v asks for a struct's field names rather than a
				// sign, so the plus is dropped before the number is written.
				return go2jsFormatG(Number(value), precision, {...parsed, flags: parsed.flags.replace(/\+/g, "")}, false);
			}

			let text;
			// A sign flag asks for a sign in front of a number that has none of
			// its own to show, so it only applies when the text rendered is the
			// number itself. A negative one already carries a minus, and a value
			// that is not a number, a string or a struct say, is left alone.
			let sign = "";

			if (flags.includes("#")) {
				text = go2jsGoSyntax(value, operand && typeof operand.__go2js_type_name === "string" && operand.__go2js_type_name !== "" ? operand.__go2js_type_name : boxedType);
			} else if (flags.includes("+") && !go2jsHasFormatMethod(value)) {
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
					text = go2jsTruncateRunes(text, precision);
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
				? (flags.includes("+") ? go2jsStrconvQuoteRuneWith(element, true, false) : go2jsQuoteRune(element))
				: go2jsFormatValue("q", inner, element, false, true));

			if (value instanceof go2jsNativeMap) {
				const parts = [];

				for (const key of value.keys()) {
					parts.push(go2jsQuoteText(go2jsBytesToString(key), flags, precision) + ":" + go2jsFormatValue("q", inner, value.get(key), false, true));
				}

				return go2jsPad("map[" + parts.join(" ") + "]", parsed, false);
			}

			// Go treats a byte slice as a string for the quoted verbs, and a byte
			// that is not part of a valid UTF8 sequence is written as \xNN rather
			// than being replaced the way a decoded string would be.
			if (go2jsIsByteCompound(accepted)) {
				return go2jsPad(go2jsQuoteBytesWithFlags(value, flags, precision), parsed, false);
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
				// The sharp flag is for a string, so a rune keeps its literal
				// unless the plus flag asks for the ASCII-only escapes.
				return go2jsPad(flags.includes("+") ? go2jsStrconvQuoteRuneWith(value, true, false) : go2jsQuoteRune(value), parsed, false);
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
					return go2jsPad(go2jsQuoteText(go2jsBytesToString(named), flags, precision), parsed, false);
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
					return go2jsPad(go2jsQuoteText(go2jsBytesToString(named), flags, precision), parsed, false);
				}
			}

			return go2jsPad(go2jsQuoteText(go2jsBytesToString(value), flags, precision), parsed, false);
		}
		case "t":
			return go2jsPad(value ? "true" : "false", parsed, false);
		case "c":
			return go2jsPad(go2jsRuneString(value), parsed, false);
		case "U": {
			// A whole number is read as the unsigned one of its width, so a
			// negative number names the code point its bits make rather than a
			// sign and a digit, and a number too wide for a double keeps every
			// bit it has.
			let number = typeof value === "bigint" ? value : BigInt(Math.trunc(Number(value)));

			if (number < 0n) {
				number = BigInt.asUintN(64, number);
			}

			let digits = number.toString(16).toUpperCase();

			while (digits.length < 4) {
				digits = "0" + digits;
			}

			// The # flag names the character the code point stands for, when it
			// can be shown, the way Go writes U+0041 'A'.
			if (flags.includes("#")) {
				const code = Number(number);

				if (code >= 0 && code <= 0x10FFFF && go2jsRuneIsPrintable(code)) {
					return go2jsPad("U+" + digits + " " + go2jsQuoteRune(code), parsed, false);
				}
			}

			return go2jsPad("U+" + digits, parsed, false);
		}
		case "T": {
			// A named slice of bytes or runes carries the slice it is built on
			// on the value itself, but %T reports the name it was declared with.
			if (typeof kind === "string" &&
				(go2jsIsByteCompound(kind) || kind === "[]rune" || kind === "[]int32") &&
				go2jsDeclaredTypeName(tagged) !== "") {
				return go2jsPad(tagged, parsed, false);
			}

			// reflect names the empty interface the way it was written before any
			// was an alias of it.
			return go2jsPad(go2jsGoTypeName(original).replace(/\bany\b/g, "interface {}"), parsed, false);
		}
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

// go2jsTextWidth counts the runes of a text, since a Go width counts runes and
// an astral character is two code units but one rune.
function go2jsTextWidth(text) {
	let width = 0;

	for (let i = 0; i < text.length; i++) {
		const code = text.charCodeAt(i);
		width++;

		if (code >= 0xd800 && code <= 0xdbff && i + 1 < text.length) {
			const next = text.charCodeAt(i + 1);

			if (next >= 0xdc00 && next <= 0xdfff) {
				i++;
			}
		}
	}

	return width;
}

function go2jsPad(text, parsed, integer) {
	const width = go2jsTextWidth(text);

	if (parsed.width <= width) {
		return text;
	}

	const fill = parsed.width - width;

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
		// A refused verb reports the operand the way %v would write it, with
		// neither the width nor the flags the refused verb asked for.
		const plain = {...parsed, flags: "", width: 0, precision: null};

		return "%!" + verb + "(" + typeName + "=" + go2jsFormatComplexText("v", plain, re, im, "", null) + ")";
	}

	return go2jsFormatComplexText(verb, parsed, re, im, flags, precision);
}

// go2jsFormatComplexText writes the (re+imi) body. Each part is a float written
// with the whole value's width, precision and flags, because that is how fmt
// writes a complex: the parentheses and the i are added around two padded
// numbers. The imaginary part always shows a sign of its own, and the plus flag
// under %v is the struct-field flag rather than a sign, so it is dropped.
function go2jsFormatComplexText(verb, parsed, re, im, flags, precision) {
	// Print and Println hand no format down, so the pair is written with no
	// width and no flags rather than read off something that was never passed.
	if (parsed === null || parsed === undefined) {
		parsed = {flags: "", width: 0, precision: null};
	}

	// %v renders a float the way %g does, with whatever precision was asked.
	const part = verb === "v" ? "g" : verb;
	const shown = (value, partFlags) => {
		const spec = "%" + partFlags + (parsed.width > 0 ? String(parsed.width) : "") +
			(precision === null ? "" : "." + precision) + part;

		// Both parts are floats in Go, so they are tagged as one, which is what
		// lets the float verbs accept them instead of refusing the part.
		return go2jsFormatValue(part, spec, go2jsTyped(value, "float64"));
	};
	const realFlags = verb === "v" ? flags.replace(/\+/g, "") : flags;
	const imagFlags = realFlags.includes("+") ? realFlags : realFlags + "+";

	return "(" + shown(re, realFlags) + shown(im, imagFlags) + "i)";
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
	if (accepted === "json.RawMessage") {
		return true;
	}

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

// go2jsBasicTypeName writes an element under the type Go names it by, since the
// aliases byte and rune are the same type as uint8 and int32 and fmt reports
// those rather than the alias a slice was declared with.
function go2jsBasicTypeName(name) {
	if (name === "byte") {
		return "uint8";
	}

	if (name === "rune") {
		return "int32";
	}

	return name;
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

// go2jsCompoundRecord reports whether a value stands for a struct that a verb
// can walk field by field: a plain record of Go fields, and not a slice, a map,
// a date, an error, a complex pair, a pointer, or a moment of the clock whose
// methods stand among its fields and would be read as fields of their own.
function go2jsCompoundRecord(value, accepted) {
	if (value === null || value === undefined || typeof value !== "object" ||
		Array.isArray(value) || value instanceof go2jsNativeMap ||
		value instanceof Error || value instanceof go2jsNativeDate ||
		value.__go2js_pointer === true || go2jsIsComplexOperand(value, accepted)) {
		return false;
	}

	// A struct has no method standing in its field list, so a record that
	// holds a function among its own properties is a wrapper of another kind.
	for (const item of Object.values(value)) {
		if (typeof item === "function") {
			return false;
		}
	}

	return true;
}

// go2jsFormatCompound walks a compound operand the way Go walks it for a verb
// that is read by its parts: a slice or an array by its elements, a struct by
// its fields, a map by its pairings. The verb with the flags, the width and the
// precision it was written with is applied to every part on its own, so a width
// pads each element rather than the whole, and the brackets are drawn around
// what came out, left without a width of their own. Without an operand to walk
// it says so by returning null, and so do the verbs that are answered by the
// shape of the whole value instead of its parts: %T and %p, %q, the sharpened
// and the signed %v, and the quoted and the hex verbs over bytes, which Go
// reads as one string.
function go2jsFormatCompound(value, accepted, verb, flags, width, precision, isBytes) {
	if (verb === "T" || verb === "p" || verb === "q" ||
		(verb === "v" && (flags.includes("#") || flags.includes("+"))) ||
		(isBytes && go2jsVerbInSet(verb, "qsxX"))) {
		return null;
	}

	// A verb that simply reads the value with no width and no precision is
	// already answered well for a struct or a map, so those two keep their own
	// paths until a width or a precision gives every part something to read.
	if (verb === "v" && width === 0 && precision === null &&
		(value instanceof go2jsNativeMap || go2jsCompoundRecord(value, accepted))) {
		return null;
	}

	const inner = "%" + flags + (width > 0 ? String(width) : "") +
		(precision === null ? "" : "." + precision) + verb;
	const parts = [];

	// A pointer to a record is walked through the record it aims at, with the
	// ampersand fmt writes in front, so %d on a *struct reads &{1 2} rather
	// than reporting the refusal a pointer could not answer on its own. A
	// walk that declines the record leaves the pointer to be read whole, so
	// nothing here is written beside the ampersand for it.
	if (value !== null && typeof value === "object" && value.__go2js_pointer === true) {
		const target = value[go2jsPointerGet]();

		if (go2jsCompoundRecord(target, accepted)) {
			const walked = go2jsFormatCompound(target, String(accepted).replace(/^\*/, ""), verb, flags, width, precision, isBytes);

			if (walked !== null) {
				return "&" + walked;
			}
		}

		return null;
	}

	// A slice or an array is walked for its elements, each one keeping the
	// element type the compound declared so a verb that is refused reports
	// uint8 rather than the int a bare number would infer.
	const elements = go2jsCompoundElements(value, accepted);

	if (elements !== null) {
		// The type the list itself carries names its element type better than
		// the one read off its first value, so a float a list holds is read as
		// a float even when it arrived without a declared type, as a field of
		// an unnamed struct does.
		const carried = value !== null && value !== undefined && typeof value.__go2js_type === "string" ? value.__go2js_type : null;
		const element = go2jsBasicTypeName(go2jsCompoundElementType(carried) ?? go2jsCompoundElementType(accepted));

		for (const item of elements) {
			parts.push(go2jsFormatValue(verb, inner, element === null ? item : go2jsTyped(item, element), false, true));
		}

		return "[" + parts.join(" ") + "]";
	}

	// A map walks its pairings with the keys in order. A key and a value take
	// the verb on their own, so a width or a precision reads on each part and
	// a verb a part refuses says so at that part.
	if (value instanceof go2jsNativeMap) {
		// The type a map was written as names the type of each key and each
		// value, so a value the map holds keeps the type it was declared with
		// and a verb reads it the way fmt reads the map's own element type.
		const mapType = go2jsDeclaredTypeName(accepted) !== "" ? accepted : (value.__go2js_type || go2jsGoTypeName(value));
		const types = go2jsMapTypes(mapType);
		const keyType = go2jsDeclaredTypeName(types.key) !== "" ? types.key : "";
		const valueType = go2jsDeclaredTypeName(types.value) !== "" ? types.value : "";
		const entries = Array.from(value.entries());
		entries.sort((a, b) => go2jsCompareValues(a[0], b[0]));

		for (const entry of entries) {
			const key = keyType === "" ? entry[0] : go2jsTyped(entry[0], keyType);
			const held = valueType === "" ? entry[1] : go2jsTyped(entry[1], valueType);

			parts.push(go2jsFormatValue(verb, inner, key, false, true) + ":" + go2jsFormatValue(verb, inner, held, false, true));
		}

		return "map[" + parts.join(" ") + "]";
	}

	// A struct walks its fields in declaration order. A field that can write
	// itself out is asked for it under the verbs that would read a string, and
	// the width and the precision it was written with are applied to the
	// answer, since fmt pads the result of a String method to the verb. Any
	// other field is read by the verb itself, and a verb the field refuses
	// reports the refusal rather than falling back to the shape of the value.
	if (go2jsCompoundRecord(value, accepted) && go2jsStructHasNoStringMethod(value)) {
		const ctor = value.constructor;
		const fields = ctor && typeof ctor.name === "string" ? go2jsStructFormats[ctor.name] : null;

		// The declared type of each field travels with it, so that a width or a
		// precision reaches what a field holds the way fmt reads the field's
		// type: a float inside a slice inside a struct is a float, and rounds
		// under the precision rather than printing in full.
		const goName = ctor && typeof ctor.name === "string" ? go2jsTypeNamesByConstructor.get(ctor) : undefined;
		let descriptors = goName !== undefined ? go2jsStructFields[goName] : undefined;

		// A struct written out in place has no class of its own, so the name and
		// the type of each field travel on the value itself.
		if (descriptors === undefined && Array.isArray(value.__go2js_struct_fields)) {
			descriptors = value.__go2js_struct_fields;
		}

		const fieldTypes = Object.create(null);
		const fieldKinds = Object.create(null);

		if (descriptors !== undefined) {
			for (const descriptor of descriptors) {
				fieldTypes[descriptor.Name] = descriptor.Type || "";
				fieldKinds[descriptor.Name] = descriptor.Kind || "";
			}
		}

		for (const key of Object.keys(value)) {
			if (fields && typeof fields[key] === "string" && go2jsVerbInSet(verb, "vsxXq")) {
				const formatter = go2jsMethodTable[fields[key]];

				if (typeof formatter === "function") {
					parts.push(go2jsPad(go2jsCallNow(formatter, null, [value[key]]), { width, flags, precision, verb }, false));
					continue;
				}
			}

			const fieldType = fieldTypes[key] || "";

			const fieldKind = fieldKinds[key] || "";
			parts.push(go2jsFormatValue(verb, inner, fieldType === "" ? value[key] : go2jsTyped(value[key], fieldType, fieldKind, fieldKind), false, true));
		}

		return "{" + parts.join(" ") + "}";
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
// exact; with one it is that many digits, rounded. The sharp flag that carries
// no precision draws the dot of a %#X even when the fraction has no digits, and
// pads the fraction of a %#x with zeros until it is four digits long, which is
// how the two sharpened verbs read in Go.
function go2jsFormatHexFloat(num, precision, upper, sharp) {
	const special = go2jsSpecialFloatText(num);

	if (special !== "") {
		return special;
	}

	const sharpened = sharp === true && precision === null;
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

	let fraction = text;

	if (sharpened && upper === false) {
		fraction = fraction.padEnd(Math.max(fraction.length, 4), "0");
	}

	const dot = fraction === "" && !(sharpened && upper) ? "" : ".";

	const body = sign + "0x1" + dot + fraction + "p" + (exponent < 0 ? "-" : "+") +
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

const go2jsCaseUpper = {
	97: 65,98: 66,99: 67,100: 68,101: 69,102: 70,
	103: 71,104: 72,105: 73,106: 74,107: 75,108: 76,
	109: 77,110: 78,111: 79,112: 80,113: 81,114: 82,
	115: 83,116: 84,117: 85,118: 86,119: 87,120: 88,
	121: 89,122: 90,181: 924,224: 192,225: 193,226: 194,
	227: 195,228: 196,229: 197,230: 198,231: 199,232: 200,
	233: 201,234: 202,235: 203,236: 204,237: 205,238: 206,
	239: 207,240: 208,241: 209,242: 210,243: 211,244: 212,
	245: 213,246: 214,248: 216,249: 217,250: 218,251: 219,
	252: 220,253: 221,254: 222,255: 376,257: 256,259: 258,
	261: 260,263: 262,265: 264,267: 266,269: 268,271: 270,
	273: 272,275: 274,277: 276,279: 278,281: 280,283: 282,
	285: 284,287: 286,289: 288,291: 290,293: 292,295: 294,
	297: 296,299: 298,301: 300,303: 302,305: 73,307: 306,
	309: 308,311: 310,314: 313,316: 315,318: 317,320: 319,
	322: 321,324: 323,326: 325,328: 327,331: 330,333: 332,
	335: 334,337: 336,339: 338,341: 340,343: 342,345: 344,
	347: 346,349: 348,351: 350,353: 352,355: 354,357: 356,
	359: 358,361: 360,363: 362,365: 364,367: 366,369: 368,
	371: 370,373: 372,375: 374,378: 377,380: 379,382: 381,
	383: 83,384: 579,387: 386,389: 388,392: 391,396: 395,
	402: 401,405: 502,409: 408,410: 573,414: 544,417: 416,
	419: 418,421: 420,424: 423,429: 428,432: 431,436: 435,
	438: 437,441: 440,445: 444,447: 503,453: 452,454: 452,
	456: 455,457: 455,459: 458,460: 458,462: 461,464: 463,
	466: 465,468: 467,470: 469,472: 471,474: 473,476: 475,
	477: 398,479: 478,481: 480,483: 482,485: 484,487: 486,
	489: 488,491: 490,493: 492,495: 494,498: 497,499: 497,
	501: 500,505: 504,507: 506,509: 508,511: 510,513: 512,
	515: 514,517: 516,519: 518,521: 520,523: 522,525: 524,
	527: 526,529: 528,531: 530,533: 532,535: 534,537: 536,
	539: 538,541: 540,543: 542,547: 546,549: 548,551: 550,
	553: 552,555: 554,557: 556,559: 558,561: 560,563: 562,
	572: 571,575: 11390,576: 11391,578: 577,583: 582,585: 584,
	587: 586,589: 588,591: 590,592: 11375,593: 11373,594: 11376,
	595: 385,596: 390,598: 393,599: 394,601: 399,603: 400,
	604: 42923,608: 403,609: 42924,611: 404,613: 42893,614: 42922,
	616: 407,617: 406,618: 42926,619: 11362,620: 42925,623: 412,
	625: 11374,626: 413,629: 415,637: 11364,640: 422,642: 42949,
	643: 425,647: 42929,648: 430,649: 580,650: 433,651: 434,
	652: 581,658: 439,669: 42930,670: 42928,837: 921,881: 880,
	883: 882,887: 886,891: 1021,892: 1022,893: 1023,940: 902,
	941: 904,942: 905,943: 906,945: 913,946: 914,947: 915,
	948: 916,949: 917,950: 918,951: 919,952: 920,953: 921,
	954: 922,955: 923,956: 924,957: 925,958: 926,959: 927,
	960: 928,961: 929,962: 931,963: 931,964: 932,965: 933,
	966: 934,967: 935,968: 936,969: 937,970: 938,971: 939,
	972: 908,973: 910,974: 911,976: 914,977: 920,981: 934,
	982: 928,983: 975,985: 984,987: 986,989: 988,991: 990,
	993: 992,995: 994,997: 996,999: 998,1001: 1000,1003: 1002,
	1005: 1004,1007: 1006,1008: 922,1009: 929,1010: 1017,1011: 895,
	1013: 917,1016: 1015,1019: 1018,1072: 1040,1073: 1041,1074: 1042,
	1075: 1043,1076: 1044,1077: 1045,1078: 1046,1079: 1047,1080: 1048,
	1081: 1049,1082: 1050,1083: 1051,1084: 1052,1085: 1053,1086: 1054,
	1087: 1055,1088: 1056,1089: 1057,1090: 1058,1091: 1059,1092: 1060,
	1093: 1061,1094: 1062,1095: 1063,1096: 1064,1097: 1065,1098: 1066,
	1099: 1067,1100: 1068,1101: 1069,1102: 1070,1103: 1071,1104: 1024,
	1105: 1025,1106: 1026,1107: 1027,1108: 1028,1109: 1029,1110: 1030,
	1111: 1031,1112: 1032,1113: 1033,1114: 1034,1115: 1035,1116: 1036,
	1117: 1037,1118: 1038,1119: 1039,1121: 1120,1123: 1122,1125: 1124,
	1127: 1126,1129: 1128,1131: 1130,1133: 1132,1135: 1134,1137: 1136,
	1139: 1138,1141: 1140,1143: 1142,1145: 1144,1147: 1146,1149: 1148,
	1151: 1150,1153: 1152,1163: 1162,1165: 1164,1167: 1166,1169: 1168,
	1171: 1170,1173: 1172,1175: 1174,1177: 1176,1179: 1178,1181: 1180,
	1183: 1182,1185: 1184,1187: 1186,1189: 1188,1191: 1190,1193: 1192,
	1195: 1194,1197: 1196,1199: 1198,1201: 1200,1203: 1202,1205: 1204,
	1207: 1206,1209: 1208,1211: 1210,1213: 1212,1215: 1214,1218: 1217,
	1220: 1219,1222: 1221,1224: 1223,1226: 1225,1228: 1227,1230: 1229,
	1231: 1216,1233: 1232,1235: 1234,1237: 1236,1239: 1238,1241: 1240,
	1243: 1242,1245: 1244,1247: 1246,1249: 1248,1251: 1250,1253: 1252,
	1255: 1254,1257: 1256,1259: 1258,1261: 1260,1263: 1262,1265: 1264,
	1267: 1266,1269: 1268,1271: 1270,1273: 1272,1275: 1274,1277: 1276,
	1279: 1278,1281: 1280,1283: 1282,1285: 1284,1287: 1286,1289: 1288,
	1291: 1290,1293: 1292,1295: 1294,1297: 1296,1299: 1298,1301: 1300,
	1303: 1302,1305: 1304,1307: 1306,1309: 1308,1311: 1310,1313: 1312,
	1315: 1314,1317: 1316,1319: 1318,1321: 1320,1323: 1322,1325: 1324,
	1327: 1326,1377: 1329,1378: 1330,1379: 1331,1380: 1332,1381: 1333,
	1382: 1334,1383: 1335,1384: 1336,1385: 1337,1386: 1338,1387: 1339,
	1388: 1340,1389: 1341,1390: 1342,1391: 1343,1392: 1344,1393: 1345,
	1394: 1346,1395: 1347,1396: 1348,1397: 1349,1398: 1350,1399: 1351,
	1400: 1352,1401: 1353,1402: 1354,1403: 1355,1404: 1356,1405: 1357,
	1406: 1358,1407: 1359,1408: 1360,1409: 1361,1410: 1362,1411: 1363,
	1412: 1364,1413: 1365,1414: 1366,4304: 7312,4305: 7313,4306: 7314,
	4307: 7315,4308: 7316,4309: 7317,4310: 7318,4311: 7319,4312: 7320,
	4313: 7321,4314: 7322,4315: 7323,4316: 7324,4317: 7325,4318: 7326,
	4319: 7327,4320: 7328,4321: 7329,4322: 7330,4323: 7331,4324: 7332,
	4325: 7333,4326: 7334,4327: 7335,4328: 7336,4329: 7337,4330: 7338,
	4331: 7339,4332: 7340,4333: 7341,4334: 7342,4335: 7343,4336: 7344,
	4337: 7345,4338: 7346,4339: 7347,4340: 7348,4341: 7349,4342: 7350,
	4343: 7351,4344: 7352,4345: 7353,4346: 7354,4349: 7357,4350: 7358,
	4351: 7359,5112: 5104,5113: 5105,5114: 5106,5115: 5107,5116: 5108,
	5117: 5109,7296: 1042,7297: 1044,7298: 1054,7299: 1057,7300: 1058,
	7301: 1058,7302: 1066,7303: 1122,7304: 42570,7545: 42877,7549: 11363,
	7566: 42950,7681: 7680,7683: 7682,7685: 7684,7687: 7686,7689: 7688,
	7691: 7690,7693: 7692,7695: 7694,7697: 7696,7699: 7698,7701: 7700,
	7703: 7702,7705: 7704,7707: 7706,7709: 7708,7711: 7710,7713: 7712,
	7715: 7714,7717: 7716,7719: 7718,7721: 7720,7723: 7722,7725: 7724,
	7727: 7726,7729: 7728,7731: 7730,7733: 7732,7735: 7734,7737: 7736,
	7739: 7738,7741: 7740,7743: 7742,7745: 7744,7747: 7746,7749: 7748,
	7751: 7750,7753: 7752,7755: 7754,7757: 7756,7759: 7758,7761: 7760,
	7763: 7762,7765: 7764,7767: 7766,7769: 7768,7771: 7770,7773: 7772,
	7775: 7774,7777: 7776,7779: 7778,7781: 7780,7783: 7782,7785: 7784,
	7787: 7786,7789: 7788,7791: 7790,7793: 7792,7795: 7794,7797: 7796,
	7799: 7798,7801: 7800,7803: 7802,7805: 7804,7807: 7806,7809: 7808,
	7811: 7810,7813: 7812,7815: 7814,7817: 7816,7819: 7818,7821: 7820,
	7823: 7822,7825: 7824,7827: 7826,7829: 7828,7835: 7776,7841: 7840,
	7843: 7842,7845: 7844,7847: 7846,7849: 7848,7851: 7850,7853: 7852,
	7855: 7854,7857: 7856,7859: 7858,7861: 7860,7863: 7862,7865: 7864,
	7867: 7866,7869: 7868,7871: 7870,7873: 7872,7875: 7874,7877: 7876,
	7879: 7878,7881: 7880,7883: 7882,7885: 7884,7887: 7886,7889: 7888,
	7891: 7890,7893: 7892,7895: 7894,7897: 7896,7899: 7898,7901: 7900,
	7903: 7902,7905: 7904,7907: 7906,7909: 7908,7911: 7910,7913: 7912,
	7915: 7914,7917: 7916,7919: 7918,7921: 7920,7923: 7922,7925: 7924,
	7927: 7926,7929: 7928,7931: 7930,7933: 7932,7935: 7934,7936: 7944,
	7937: 7945,7938: 7946,7939: 7947,7940: 7948,7941: 7949,7942: 7950,
	7943: 7951,7952: 7960,7953: 7961,7954: 7962,7955: 7963,7956: 7964,
	7957: 7965,7968: 7976,7969: 7977,7970: 7978,7971: 7979,7972: 7980,
	7973: 7981,7974: 7982,7975: 7983,7984: 7992,7985: 7993,7986: 7994,
	7987: 7995,7988: 7996,7989: 7997,7990: 7998,7991: 7999,8000: 8008,
	8001: 8009,8002: 8010,8003: 8011,8004: 8012,8005: 8013,8017: 8025,
	8019: 8027,8021: 8029,8023: 8031,8032: 8040,8033: 8041,8034: 8042,
	8035: 8043,8036: 8044,8037: 8045,8038: 8046,8039: 8047,8048: 8122,
	8049: 8123,8050: 8136,8051: 8137,8052: 8138,8053: 8139,8054: 8154,
	8055: 8155,8056: 8184,8057: 8185,8058: 8170,8059: 8171,8060: 8186,
	8061: 8187,8064: 8072,8065: 8073,8066: 8074,8067: 8075,8068: 8076,
	8069: 8077,8070: 8078,8071: 8079,8080: 8088,8081: 8089,8082: 8090,
	8083: 8091,8084: 8092,8085: 8093,8086: 8094,8087: 8095,8096: 8104,
	8097: 8105,8098: 8106,8099: 8107,8100: 8108,8101: 8109,8102: 8110,
	8103: 8111,8112: 8120,8113: 8121,8115: 8124,8126: 921,8131: 8140,
	8144: 8152,8145: 8153,8160: 8168,8161: 8169,8165: 8172,8179: 8188,
	8526: 8498,8560: 8544,8561: 8545,8562: 8546,8563: 8547,8564: 8548,
	8565: 8549,8566: 8550,8567: 8551,8568: 8552,8569: 8553,8570: 8554,
	8571: 8555,8572: 8556,8573: 8557,8574: 8558,8575: 8559,8580: 8579,
	9424: 9398,9425: 9399,9426: 9400,9427: 9401,9428: 9402,9429: 9403,
	9430: 9404,9431: 9405,9432: 9406,9433: 9407,9434: 9408,9435: 9409,
	9436: 9410,9437: 9411,9438: 9412,9439: 9413,9440: 9414,9441: 9415,
	9442: 9416,9443: 9417,9444: 9418,9445: 9419,9446: 9420,9447: 9421,
	9448: 9422,9449: 9423,11312: 11264,11313: 11265,11314: 11266,11315: 11267,
	11316: 11268,11317: 11269,11318: 11270,11319: 11271,11320: 11272,11321: 11273,
	11322: 11274,11323: 11275,11324: 11276,11325: 11277,11326: 11278,11327: 11279,
	11328: 11280,11329: 11281,11330: 11282,11331: 11283,11332: 11284,11333: 11285,
	11334: 11286,11335: 11287,11336: 11288,11337: 11289,11338: 11290,11339: 11291,
	11340: 11292,11341: 11293,11342: 11294,11343: 11295,11344: 11296,11345: 11297,
	11346: 11298,11347: 11299,11348: 11300,11349: 11301,11350: 11302,11351: 11303,
	11352: 11304,11353: 11305,11354: 11306,11355: 11307,11356: 11308,11357: 11309,
	11358: 11310,11359: 11311,11361: 11360,11365: 570,11366: 574,11368: 11367,
	11370: 11369,11372: 11371,11379: 11378,11382: 11381,11393: 11392,11395: 11394,
	11397: 11396,11399: 11398,11401: 11400,11403: 11402,11405: 11404,11407: 11406,
	11409: 11408,11411: 11410,11413: 11412,11415: 11414,11417: 11416,11419: 11418,
	11421: 11420,11423: 11422,11425: 11424,11427: 11426,11429: 11428,11431: 11430,
	11433: 11432,11435: 11434,11437: 11436,11439: 11438,11441: 11440,11443: 11442,
	11445: 11444,11447: 11446,11449: 11448,11451: 11450,11453: 11452,11455: 11454,
	11457: 11456,11459: 11458,11461: 11460,11463: 11462,11465: 11464,11467: 11466,
	11469: 11468,11471: 11470,11473: 11472,11475: 11474,11477: 11476,11479: 11478,
	11481: 11480,11483: 11482,11485: 11484,11487: 11486,11489: 11488,11491: 11490,
	11500: 11499,11502: 11501,11507: 11506,11520: 4256,11521: 4257,11522: 4258,
	11523: 4259,11524: 4260,11525: 4261,11526: 4262,11527: 4263,11528: 4264,
	11529: 4265,11530: 4266,11531: 4267,11532: 4268,11533: 4269,11534: 4270,
	11535: 4271,11536: 4272,11537: 4273,11538: 4274,11539: 4275,11540: 4276,
	11541: 4277,11542: 4278,11543: 4279,11544: 4280,11545: 4281,11546: 4282,
	11547: 4283,11548: 4284,11549: 4285,11550: 4286,11551: 4287,11552: 4288,
	11553: 4289,11554: 4290,11555: 4291,11556: 4292,11557: 4293,11559: 4295,
	11565: 4301,42561: 42560,42563: 42562,42565: 42564,42567: 42566,42569: 42568,
	42571: 42570,42573: 42572,42575: 42574,42577: 42576,42579: 42578,42581: 42580,
	42583: 42582,42585: 42584,42587: 42586,42589: 42588,42591: 42590,42593: 42592,
	42595: 42594,42597: 42596,42599: 42598,42601: 42600,42603: 42602,42605: 42604,
	42625: 42624,42627: 42626,42629: 42628,42631: 42630,42633: 42632,42635: 42634,
	42637: 42636,42639: 42638,42641: 42640,42643: 42642,42645: 42644,42647: 42646,
	42649: 42648,42651: 42650,42787: 42786,42789: 42788,42791: 42790,42793: 42792,
	42795: 42794,42797: 42796,42799: 42798,42803: 42802,42805: 42804,42807: 42806,
	42809: 42808,42811: 42810,42813: 42812,42815: 42814,42817: 42816,42819: 42818,
	42821: 42820,42823: 42822,42825: 42824,42827: 42826,42829: 42828,42831: 42830,
	42833: 42832,42835: 42834,42837: 42836,42839: 42838,42841: 42840,42843: 42842,
	42845: 42844,42847: 42846,42849: 42848,42851: 42850,42853: 42852,42855: 42854,
	42857: 42856,42859: 42858,42861: 42860,42863: 42862,42874: 42873,42876: 42875,
	42879: 42878,42881: 42880,42883: 42882,42885: 42884,42887: 42886,42892: 42891,
	42897: 42896,42899: 42898,42900: 42948,42903: 42902,42905: 42904,42907: 42906,
	42909: 42908,42911: 42910,42913: 42912,42915: 42914,42917: 42916,42919: 42918,
	42921: 42920,42933: 42932,42935: 42934,42937: 42936,42939: 42938,42941: 42940,
	42943: 42942,42945: 42944,42947: 42946,42952: 42951,42954: 42953,42961: 42960,
	42967: 42966,42969: 42968,42998: 42997,43859: 42931,43888: 5024,43889: 5025,
	43890: 5026,43891: 5027,43892: 5028,43893: 5029,43894: 5030,43895: 5031,
	43896: 5032,43897: 5033,43898: 5034,43899: 5035,43900: 5036,43901: 5037,
	43902: 5038,43903: 5039,43904: 5040,43905: 5041,43906: 5042,43907: 5043,
	43908: 5044,43909: 5045,43910: 5046,43911: 5047,43912: 5048,43913: 5049,
	43914: 5050,43915: 5051,43916: 5052,43917: 5053,43918: 5054,43919: 5055,
	43920: 5056,43921: 5057,43922: 5058,43923: 5059,43924: 5060,43925: 5061,
	43926: 5062,43927: 5063,43928: 5064,43929: 5065,43930: 5066,43931: 5067,
	43932: 5068,43933: 5069,43934: 5070,43935: 5071,43936: 5072,43937: 5073,
	43938: 5074,43939: 5075,43940: 5076,43941: 5077,43942: 5078,43943: 5079,
	43944: 5080,43945: 5081,43946: 5082,43947: 5083,43948: 5084,43949: 5085,
	43950: 5086,43951: 5087,43952: 5088,43953: 5089,43954: 5090,43955: 5091,
	43956: 5092,43957: 5093,43958: 5094,43959: 5095,43960: 5096,43961: 5097,
	43962: 5098,43963: 5099,43964: 5100,43965: 5101,43966: 5102,43967: 5103,
	65345: 65313,65346: 65314,65347: 65315,65348: 65316,65349: 65317,65350: 65318,
	65351: 65319,65352: 65320,65353: 65321,65354: 65322,65355: 65323,65356: 65324,
	65357: 65325,65358: 65326,65359: 65327,65360: 65328,65361: 65329,65362: 65330,
	65363: 65331,65364: 65332,65365: 65333,65366: 65334,65367: 65335,65368: 65336,
	65369: 65337,65370: 65338,66600: 66560,66601: 66561,66602: 66562,66603: 66563,
	66604: 66564,66605: 66565,66606: 66566,66607: 66567,66608: 66568,66609: 66569,
	66610: 66570,66611: 66571,66612: 66572,66613: 66573,66614: 66574,66615: 66575,
	66616: 66576,66617: 66577,66618: 66578,66619: 66579,66620: 66580,66621: 66581,
	66622: 66582,66623: 66583,66624: 66584,66625: 66585,66626: 66586,66627: 66587,
	66628: 66588,66629: 66589,66630: 66590,66631: 66591,66632: 66592,66633: 66593,
	66634: 66594,66635: 66595,66636: 66596,66637: 66597,66638: 66598,66639: 66599,
	66776: 66736,66777: 66737,66778: 66738,66779: 66739,66780: 66740,66781: 66741,
	66782: 66742,66783: 66743,66784: 66744,66785: 66745,66786: 66746,66787: 66747,
	66788: 66748,66789: 66749,66790: 66750,66791: 66751,66792: 66752,66793: 66753,
	66794: 66754,66795: 66755,66796: 66756,66797: 66757,66798: 66758,66799: 66759,
	66800: 66760,66801: 66761,66802: 66762,66803: 66763,66804: 66764,66805: 66765,
	66806: 66766,66807: 66767,66808: 66768,66809: 66769,66810: 66770,66811: 66771,
	66967: 66928,66968: 66929,66969: 66930,66970: 66931,66971: 66932,66972: 66933,
	66973: 66934,66974: 66935,66975: 66936,66976: 66937,66977: 66938,66979: 66940,
	66980: 66941,66981: 66942,66982: 66943,66983: 66944,66984: 66945,66985: 66946,
	66986: 66947,66987: 66948,66988: 66949,66989: 66950,66990: 66951,66991: 66952,
	66992: 66953,66993: 66954,66995: 66956,66996: 66957,66997: 66958,66998: 66959,
	66999: 66960,67000: 66961,67001: 66962,67003: 66964,67004: 66965,68800: 68736,
	68801: 68737,68802: 68738,68803: 68739,68804: 68740,68805: 68741,68806: 68742,
	68807: 68743,68808: 68744,68809: 68745,68810: 68746,68811: 68747,68812: 68748,
	68813: 68749,68814: 68750,68815: 68751,68816: 68752,68817: 68753,68818: 68754,
	68819: 68755,68820: 68756,68821: 68757,68822: 68758,68823: 68759,68824: 68760,
	68825: 68761,68826: 68762,68827: 68763,68828: 68764,68829: 68765,68830: 68766,
	68831: 68767,68832: 68768,68833: 68769,68834: 68770,68835: 68771,68836: 68772,
	68837: 68773,68838: 68774,68839: 68775,68840: 68776,68841: 68777,68842: 68778,
	68843: 68779,68844: 68780,68845: 68781,68846: 68782,68847: 68783,68848: 68784,
	68849: 68785,68850: 68786,71872: 71840,71873: 71841,71874: 71842,71875: 71843,
	71876: 71844,71877: 71845,71878: 71846,71879: 71847,71880: 71848,71881: 71849,
	71882: 71850,71883: 71851,71884: 71852,71885: 71853,71886: 71854,71887: 71855,
	71888: 71856,71889: 71857,71890: 71858,71891: 71859,71892: 71860,71893: 71861,
	71894: 71862,71895: 71863,71896: 71864,71897: 71865,71898: 71866,71899: 71867,
	71900: 71868,71901: 71869,71902: 71870,71903: 71871,93792: 93760,93793: 93761,
	93794: 93762,93795: 93763,93796: 93764,93797: 93765,93798: 93766,93799: 93767,
	93800: 93768,93801: 93769,93802: 93770,93803: 93771,93804: 93772,93805: 93773,
	93806: 93774,93807: 93775,93808: 93776,93809: 93777,93810: 93778,93811: 93779,
	93812: 93780,93813: 93781,93814: 93782,93815: 93783,93816: 93784,93817: 93785,
	93818: 93786,93819: 93787,93820: 93788,93821: 93789,93822: 93790,93823: 93791,
	125218: 125184,125219: 125185,125220: 125186,125221: 125187,125222: 125188,125223: 125189,
	125224: 125190,125225: 125191,125226: 125192,125227: 125193,125228: 125194,125229: 125195,
	125230: 125196,125231: 125197,125232: 125198,125233: 125199,125234: 125200,125235: 125201,
	125236: 125202,125237: 125203,125238: 125204,125239: 125205,125240: 125206,125241: 125207,
	125242: 125208,125243: 125209,125244: 125210,125245: 125211,125246: 125212,125247: 125213,
	125248: 125214,125249: 125215,125250: 125216,125251: 125217,
};

const go2jsCaseLower = {
	65: 97,66: 98,67: 99,68: 100,69: 101,70: 102,
	71: 103,72: 104,73: 105,74: 106,75: 107,76: 108,
	77: 109,78: 110,79: 111,80: 112,81: 113,82: 114,
	83: 115,84: 116,85: 117,86: 118,87: 119,88: 120,
	89: 121,90: 122,192: 224,193: 225,194: 226,195: 227,
	196: 228,197: 229,198: 230,199: 231,200: 232,201: 233,
	202: 234,203: 235,204: 236,205: 237,206: 238,207: 239,
	208: 240,209: 241,210: 242,211: 243,212: 244,213: 245,
	214: 246,216: 248,217: 249,218: 250,219: 251,220: 252,
	221: 253,222: 254,256: 257,258: 259,260: 261,262: 263,
	264: 265,266: 267,268: 269,270: 271,272: 273,274: 275,
	276: 277,278: 279,280: 281,282: 283,284: 285,286: 287,
	288: 289,290: 291,292: 293,294: 295,296: 297,298: 299,
	300: 301,302: 303,304: 105,306: 307,308: 309,310: 311,
	313: 314,315: 316,317: 318,319: 320,321: 322,323: 324,
	325: 326,327: 328,330: 331,332: 333,334: 335,336: 337,
	338: 339,340: 341,342: 343,344: 345,346: 347,348: 349,
	350: 351,352: 353,354: 355,356: 357,358: 359,360: 361,
	362: 363,364: 365,366: 367,368: 369,370: 371,372: 373,
	374: 375,376: 255,377: 378,379: 380,381: 382,385: 595,
	386: 387,388: 389,390: 596,391: 392,393: 598,394: 599,
	395: 396,398: 477,399: 601,400: 603,401: 402,403: 608,
	404: 611,406: 617,407: 616,408: 409,412: 623,413: 626,
	415: 629,416: 417,418: 419,420: 421,422: 640,423: 424,
	425: 643,428: 429,430: 648,431: 432,433: 650,434: 651,
	435: 436,437: 438,439: 658,440: 441,444: 445,452: 454,
	453: 454,455: 457,456: 457,458: 460,459: 460,461: 462,
	463: 464,465: 466,467: 468,469: 470,471: 472,473: 474,
	475: 476,478: 479,480: 481,482: 483,484: 485,486: 487,
	488: 489,490: 491,492: 493,494: 495,497: 499,498: 499,
	500: 501,502: 405,503: 447,504: 505,506: 507,508: 509,
	510: 511,512: 513,514: 515,516: 517,518: 519,520: 521,
	522: 523,524: 525,526: 527,528: 529,530: 531,532: 533,
	534: 535,536: 537,538: 539,540: 541,542: 543,544: 414,
	546: 547,548: 549,550: 551,552: 553,554: 555,556: 557,
	558: 559,560: 561,562: 563,570: 11365,571: 572,573: 410,
	574: 11366,577: 578,579: 384,580: 649,581: 652,582: 583,
	584: 585,586: 587,588: 589,590: 591,880: 881,882: 883,
	886: 887,895: 1011,902: 940,904: 941,905: 942,906: 943,
	908: 972,910: 973,911: 974,913: 945,914: 946,915: 947,
	916: 948,917: 949,918: 950,919: 951,920: 952,921: 953,
	922: 954,923: 955,924: 956,925: 957,926: 958,927: 959,
	928: 960,929: 961,931: 963,932: 964,933: 965,934: 966,
	935: 967,936: 968,937: 969,938: 970,939: 971,975: 983,
	984: 985,986: 987,988: 989,990: 991,992: 993,994: 995,
	996: 997,998: 999,1000: 1001,1002: 1003,1004: 1005,1006: 1007,
	1012: 952,1015: 1016,1017: 1010,1018: 1019,1021: 891,1022: 892,
	1023: 893,1024: 1104,1025: 1105,1026: 1106,1027: 1107,1028: 1108,
	1029: 1109,1030: 1110,1031: 1111,1032: 1112,1033: 1113,1034: 1114,
	1035: 1115,1036: 1116,1037: 1117,1038: 1118,1039: 1119,1040: 1072,
	1041: 1073,1042: 1074,1043: 1075,1044: 1076,1045: 1077,1046: 1078,
	1047: 1079,1048: 1080,1049: 1081,1050: 1082,1051: 1083,1052: 1084,
	1053: 1085,1054: 1086,1055: 1087,1056: 1088,1057: 1089,1058: 1090,
	1059: 1091,1060: 1092,1061: 1093,1062: 1094,1063: 1095,1064: 1096,
	1065: 1097,1066: 1098,1067: 1099,1068: 1100,1069: 1101,1070: 1102,
	1071: 1103,1120: 1121,1122: 1123,1124: 1125,1126: 1127,1128: 1129,
	1130: 1131,1132: 1133,1134: 1135,1136: 1137,1138: 1139,1140: 1141,
	1142: 1143,1144: 1145,1146: 1147,1148: 1149,1150: 1151,1152: 1153,
	1162: 1163,1164: 1165,1166: 1167,1168: 1169,1170: 1171,1172: 1173,
	1174: 1175,1176: 1177,1178: 1179,1180: 1181,1182: 1183,1184: 1185,
	1186: 1187,1188: 1189,1190: 1191,1192: 1193,1194: 1195,1196: 1197,
	1198: 1199,1200: 1201,1202: 1203,1204: 1205,1206: 1207,1208: 1209,
	1210: 1211,1212: 1213,1214: 1215,1216: 1231,1217: 1218,1219: 1220,
	1221: 1222,1223: 1224,1225: 1226,1227: 1228,1229: 1230,1232: 1233,
	1234: 1235,1236: 1237,1238: 1239,1240: 1241,1242: 1243,1244: 1245,
	1246: 1247,1248: 1249,1250: 1251,1252: 1253,1254: 1255,1256: 1257,
	1258: 1259,1260: 1261,1262: 1263,1264: 1265,1266: 1267,1268: 1269,
	1270: 1271,1272: 1273,1274: 1275,1276: 1277,1278: 1279,1280: 1281,
	1282: 1283,1284: 1285,1286: 1287,1288: 1289,1290: 1291,1292: 1293,
	1294: 1295,1296: 1297,1298: 1299,1300: 1301,1302: 1303,1304: 1305,
	1306: 1307,1308: 1309,1310: 1311,1312: 1313,1314: 1315,1316: 1317,
	1318: 1319,1320: 1321,1322: 1323,1324: 1325,1326: 1327,1329: 1377,
	1330: 1378,1331: 1379,1332: 1380,1333: 1381,1334: 1382,1335: 1383,
	1336: 1384,1337: 1385,1338: 1386,1339: 1387,1340: 1388,1341: 1389,
	1342: 1390,1343: 1391,1344: 1392,1345: 1393,1346: 1394,1347: 1395,
	1348: 1396,1349: 1397,1350: 1398,1351: 1399,1352: 1400,1353: 1401,
	1354: 1402,1355: 1403,1356: 1404,1357: 1405,1358: 1406,1359: 1407,
	1360: 1408,1361: 1409,1362: 1410,1363: 1411,1364: 1412,1365: 1413,
	1366: 1414,4256: 11520,4257: 11521,4258: 11522,4259: 11523,4260: 11524,
	4261: 11525,4262: 11526,4263: 11527,4264: 11528,4265: 11529,4266: 11530,
	4267: 11531,4268: 11532,4269: 11533,4270: 11534,4271: 11535,4272: 11536,
	4273: 11537,4274: 11538,4275: 11539,4276: 11540,4277: 11541,4278: 11542,
	4279: 11543,4280: 11544,4281: 11545,4282: 11546,4283: 11547,4284: 11548,
	4285: 11549,4286: 11550,4287: 11551,4288: 11552,4289: 11553,4290: 11554,
	4291: 11555,4292: 11556,4293: 11557,4295: 11559,4301: 11565,5024: 43888,
	5025: 43889,5026: 43890,5027: 43891,5028: 43892,5029: 43893,5030: 43894,
	5031: 43895,5032: 43896,5033: 43897,5034: 43898,5035: 43899,5036: 43900,
	5037: 43901,5038: 43902,5039: 43903,5040: 43904,5041: 43905,5042: 43906,
	5043: 43907,5044: 43908,5045: 43909,5046: 43910,5047: 43911,5048: 43912,
	5049: 43913,5050: 43914,5051: 43915,5052: 43916,5053: 43917,5054: 43918,
	5055: 43919,5056: 43920,5057: 43921,5058: 43922,5059: 43923,5060: 43924,
	5061: 43925,5062: 43926,5063: 43927,5064: 43928,5065: 43929,5066: 43930,
	5067: 43931,5068: 43932,5069: 43933,5070: 43934,5071: 43935,5072: 43936,
	5073: 43937,5074: 43938,5075: 43939,5076: 43940,5077: 43941,5078: 43942,
	5079: 43943,5080: 43944,5081: 43945,5082: 43946,5083: 43947,5084: 43948,
	5085: 43949,5086: 43950,5087: 43951,5088: 43952,5089: 43953,5090: 43954,
	5091: 43955,5092: 43956,5093: 43957,5094: 43958,5095: 43959,5096: 43960,
	5097: 43961,5098: 43962,5099: 43963,5100: 43964,5101: 43965,5102: 43966,
	5103: 43967,5104: 5112,5105: 5113,5106: 5114,5107: 5115,5108: 5116,
	5109: 5117,7312: 4304,7313: 4305,7314: 4306,7315: 4307,7316: 4308,
	7317: 4309,7318: 4310,7319: 4311,7320: 4312,7321: 4313,7322: 4314,
	7323: 4315,7324: 4316,7325: 4317,7326: 4318,7327: 4319,7328: 4320,
	7329: 4321,7330: 4322,7331: 4323,7332: 4324,7333: 4325,7334: 4326,
	7335: 4327,7336: 4328,7337: 4329,7338: 4330,7339: 4331,7340: 4332,
	7341: 4333,7342: 4334,7343: 4335,7344: 4336,7345: 4337,7346: 4338,
	7347: 4339,7348: 4340,7349: 4341,7350: 4342,7351: 4343,7352: 4344,
	7353: 4345,7354: 4346,7357: 4349,7358: 4350,7359: 4351,7680: 7681,
	7682: 7683,7684: 7685,7686: 7687,7688: 7689,7690: 7691,7692: 7693,
	7694: 7695,7696: 7697,7698: 7699,7700: 7701,7702: 7703,7704: 7705,
	7706: 7707,7708: 7709,7710: 7711,7712: 7713,7714: 7715,7716: 7717,
	7718: 7719,7720: 7721,7722: 7723,7724: 7725,7726: 7727,7728: 7729,
	7730: 7731,7732: 7733,7734: 7735,7736: 7737,7738: 7739,7740: 7741,
	7742: 7743,7744: 7745,7746: 7747,7748: 7749,7750: 7751,7752: 7753,
	7754: 7755,7756: 7757,7758: 7759,7760: 7761,7762: 7763,7764: 7765,
	7766: 7767,7768: 7769,7770: 7771,7772: 7773,7774: 7775,7776: 7777,
	7778: 7779,7780: 7781,7782: 7783,7784: 7785,7786: 7787,7788: 7789,
	7790: 7791,7792: 7793,7794: 7795,7796: 7797,7798: 7799,7800: 7801,
	7802: 7803,7804: 7805,7806: 7807,7808: 7809,7810: 7811,7812: 7813,
	7814: 7815,7816: 7817,7818: 7819,7820: 7821,7822: 7823,7824: 7825,
	7826: 7827,7828: 7829,7838: 223,7840: 7841,7842: 7843,7844: 7845,
	7846: 7847,7848: 7849,7850: 7851,7852: 7853,7854: 7855,7856: 7857,
	7858: 7859,7860: 7861,7862: 7863,7864: 7865,7866: 7867,7868: 7869,
	7870: 7871,7872: 7873,7874: 7875,7876: 7877,7878: 7879,7880: 7881,
	7882: 7883,7884: 7885,7886: 7887,7888: 7889,7890: 7891,7892: 7893,
	7894: 7895,7896: 7897,7898: 7899,7900: 7901,7902: 7903,7904: 7905,
	7906: 7907,7908: 7909,7910: 7911,7912: 7913,7914: 7915,7916: 7917,
	7918: 7919,7920: 7921,7922: 7923,7924: 7925,7926: 7927,7928: 7929,
	7930: 7931,7932: 7933,7934: 7935,7944: 7936,7945: 7937,7946: 7938,
	7947: 7939,7948: 7940,7949: 7941,7950: 7942,7951: 7943,7960: 7952,
	7961: 7953,7962: 7954,7963: 7955,7964: 7956,7965: 7957,7976: 7968,
	7977: 7969,7978: 7970,7979: 7971,7980: 7972,7981: 7973,7982: 7974,
	7983: 7975,7992: 7984,7993: 7985,7994: 7986,7995: 7987,7996: 7988,
	7997: 7989,7998: 7990,7999: 7991,8008: 8000,8009: 8001,8010: 8002,
	8011: 8003,8012: 8004,8013: 8005,8025: 8017,8027: 8019,8029: 8021,
	8031: 8023,8040: 8032,8041: 8033,8042: 8034,8043: 8035,8044: 8036,
	8045: 8037,8046: 8038,8047: 8039,8072: 8064,8073: 8065,8074: 8066,
	8075: 8067,8076: 8068,8077: 8069,8078: 8070,8079: 8071,8088: 8080,
	8089: 8081,8090: 8082,8091: 8083,8092: 8084,8093: 8085,8094: 8086,
	8095: 8087,8104: 8096,8105: 8097,8106: 8098,8107: 8099,8108: 8100,
	8109: 8101,8110: 8102,8111: 8103,8120: 8112,8121: 8113,8122: 8048,
	8123: 8049,8124: 8115,8136: 8050,8137: 8051,8138: 8052,8139: 8053,
	8140: 8131,8152: 8144,8153: 8145,8154: 8054,8155: 8055,8168: 8160,
	8169: 8161,8170: 8058,8171: 8059,8172: 8165,8184: 8056,8185: 8057,
	8186: 8060,8187: 8061,8188: 8179,8486: 969,8490: 107,8491: 229,
	8498: 8526,8544: 8560,8545: 8561,8546: 8562,8547: 8563,8548: 8564,
	8549: 8565,8550: 8566,8551: 8567,8552: 8568,8553: 8569,8554: 8570,
	8555: 8571,8556: 8572,8557: 8573,8558: 8574,8559: 8575,8579: 8580,
	9398: 9424,9399: 9425,9400: 9426,9401: 9427,9402: 9428,9403: 9429,
	9404: 9430,9405: 9431,9406: 9432,9407: 9433,9408: 9434,9409: 9435,
	9410: 9436,9411: 9437,9412: 9438,9413: 9439,9414: 9440,9415: 9441,
	9416: 9442,9417: 9443,9418: 9444,9419: 9445,9420: 9446,9421: 9447,
	9422: 9448,9423: 9449,11264: 11312,11265: 11313,11266: 11314,11267: 11315,
	11268: 11316,11269: 11317,11270: 11318,11271: 11319,11272: 11320,11273: 11321,
	11274: 11322,11275: 11323,11276: 11324,11277: 11325,11278: 11326,11279: 11327,
	11280: 11328,11281: 11329,11282: 11330,11283: 11331,11284: 11332,11285: 11333,
	11286: 11334,11287: 11335,11288: 11336,11289: 11337,11290: 11338,11291: 11339,
	11292: 11340,11293: 11341,11294: 11342,11295: 11343,11296: 11344,11297: 11345,
	11298: 11346,11299: 11347,11300: 11348,11301: 11349,11302: 11350,11303: 11351,
	11304: 11352,11305: 11353,11306: 11354,11307: 11355,11308: 11356,11309: 11357,
	11310: 11358,11311: 11359,11360: 11361,11362: 619,11363: 7549,11364: 637,
	11367: 11368,11369: 11370,11371: 11372,11373: 593,11374: 625,11375: 592,
	11376: 594,11378: 11379,11381: 11382,11390: 575,11391: 576,11392: 11393,
	11394: 11395,11396: 11397,11398: 11399,11400: 11401,11402: 11403,11404: 11405,
	11406: 11407,11408: 11409,11410: 11411,11412: 11413,11414: 11415,11416: 11417,
	11418: 11419,11420: 11421,11422: 11423,11424: 11425,11426: 11427,11428: 11429,
	11430: 11431,11432: 11433,11434: 11435,11436: 11437,11438: 11439,11440: 11441,
	11442: 11443,11444: 11445,11446: 11447,11448: 11449,11450: 11451,11452: 11453,
	11454: 11455,11456: 11457,11458: 11459,11460: 11461,11462: 11463,11464: 11465,
	11466: 11467,11468: 11469,11470: 11471,11472: 11473,11474: 11475,11476: 11477,
	11478: 11479,11480: 11481,11482: 11483,11484: 11485,11486: 11487,11488: 11489,
	11490: 11491,11499: 11500,11501: 11502,11506: 11507,42560: 42561,42562: 42563,
	42564: 42565,42566: 42567,42568: 42569,42570: 42571,42572: 42573,42574: 42575,
	42576: 42577,42578: 42579,42580: 42581,42582: 42583,42584: 42585,42586: 42587,
	42588: 42589,42590: 42591,42592: 42593,42594: 42595,42596: 42597,42598: 42599,
	42600: 42601,42602: 42603,42604: 42605,42624: 42625,42626: 42627,42628: 42629,
	42630: 42631,42632: 42633,42634: 42635,42636: 42637,42638: 42639,42640: 42641,
	42642: 42643,42644: 42645,42646: 42647,42648: 42649,42650: 42651,42786: 42787,
	42788: 42789,42790: 42791,42792: 42793,42794: 42795,42796: 42797,42798: 42799,
	42802: 42803,42804: 42805,42806: 42807,42808: 42809,42810: 42811,42812: 42813,
	42814: 42815,42816: 42817,42818: 42819,42820: 42821,42822: 42823,42824: 42825,
	42826: 42827,42828: 42829,42830: 42831,42832: 42833,42834: 42835,42836: 42837,
	42838: 42839,42840: 42841,42842: 42843,42844: 42845,42846: 42847,42848: 42849,
	42850: 42851,42852: 42853,42854: 42855,42856: 42857,42858: 42859,42860: 42861,
	42862: 42863,42873: 42874,42875: 42876,42877: 7545,42878: 42879,42880: 42881,
	42882: 42883,42884: 42885,42886: 42887,42891: 42892,42893: 613,42896: 42897,
	42898: 42899,42902: 42903,42904: 42905,42906: 42907,42908: 42909,42910: 42911,
	42912: 42913,42914: 42915,42916: 42917,42918: 42919,42920: 42921,42922: 614,
	42923: 604,42924: 609,42925: 620,42926: 618,42928: 670,42929: 647,
	42930: 669,42931: 43859,42932: 42933,42934: 42935,42936: 42937,42938: 42939,
	42940: 42941,42942: 42943,42944: 42945,42946: 42947,42948: 42900,42949: 642,
	42950: 7566,42951: 42952,42953: 42954,42960: 42961,42966: 42967,42968: 42969,
	42997: 42998,65313: 65345,65314: 65346,65315: 65347,65316: 65348,65317: 65349,
	65318: 65350,65319: 65351,65320: 65352,65321: 65353,65322: 65354,65323: 65355,
	65324: 65356,65325: 65357,65326: 65358,65327: 65359,65328: 65360,65329: 65361,
	65330: 65362,65331: 65363,65332: 65364,65333: 65365,65334: 65366,65335: 65367,
	65336: 65368,65337: 65369,65338: 65370,66560: 66600,66561: 66601,66562: 66602,
	66563: 66603,66564: 66604,66565: 66605,66566: 66606,66567: 66607,66568: 66608,
	66569: 66609,66570: 66610,66571: 66611,66572: 66612,66573: 66613,66574: 66614,
	66575: 66615,66576: 66616,66577: 66617,66578: 66618,66579: 66619,66580: 66620,
	66581: 66621,66582: 66622,66583: 66623,66584: 66624,66585: 66625,66586: 66626,
	66587: 66627,66588: 66628,66589: 66629,66590: 66630,66591: 66631,66592: 66632,
	66593: 66633,66594: 66634,66595: 66635,66596: 66636,66597: 66637,66598: 66638,
	66599: 66639,66736: 66776,66737: 66777,66738: 66778,66739: 66779,66740: 66780,
	66741: 66781,66742: 66782,66743: 66783,66744: 66784,66745: 66785,66746: 66786,
	66747: 66787,66748: 66788,66749: 66789,66750: 66790,66751: 66791,66752: 66792,
	66753: 66793,66754: 66794,66755: 66795,66756: 66796,66757: 66797,66758: 66798,
	66759: 66799,66760: 66800,66761: 66801,66762: 66802,66763: 66803,66764: 66804,
	66765: 66805,66766: 66806,66767: 66807,66768: 66808,66769: 66809,66770: 66810,
	66771: 66811,66928: 66967,66929: 66968,66930: 66969,66931: 66970,66932: 66971,
	66933: 66972,66934: 66973,66935: 66974,66936: 66975,66937: 66976,66938: 66977,
	66940: 66979,66941: 66980,66942: 66981,66943: 66982,66944: 66983,66945: 66984,
	66946: 66985,66947: 66986,66948: 66987,66949: 66988,66950: 66989,66951: 66990,
	66952: 66991,66953: 66992,66954: 66993,66956: 66995,66957: 66996,66958: 66997,
	66959: 66998,66960: 66999,66961: 67000,66962: 67001,66964: 67003,66965: 67004,
	68736: 68800,68737: 68801,68738: 68802,68739: 68803,68740: 68804,68741: 68805,
	68742: 68806,68743: 68807,68744: 68808,68745: 68809,68746: 68810,68747: 68811,
	68748: 68812,68749: 68813,68750: 68814,68751: 68815,68752: 68816,68753: 68817,
	68754: 68818,68755: 68819,68756: 68820,68757: 68821,68758: 68822,68759: 68823,
	68760: 68824,68761: 68825,68762: 68826,68763: 68827,68764: 68828,68765: 68829,
	68766: 68830,68767: 68831,68768: 68832,68769: 68833,68770: 68834,68771: 68835,
	68772: 68836,68773: 68837,68774: 68838,68775: 68839,68776: 68840,68777: 68841,
	68778: 68842,68779: 68843,68780: 68844,68781: 68845,68782: 68846,68783: 68847,
	68784: 68848,68785: 68849,68786: 68850,71840: 71872,71841: 71873,71842: 71874,
	71843: 71875,71844: 71876,71845: 71877,71846: 71878,71847: 71879,71848: 71880,
	71849: 71881,71850: 71882,71851: 71883,71852: 71884,71853: 71885,71854: 71886,
	71855: 71887,71856: 71888,71857: 71889,71858: 71890,71859: 71891,71860: 71892,
	71861: 71893,71862: 71894,71863: 71895,71864: 71896,71865: 71897,71866: 71898,
	71867: 71899,71868: 71900,71869: 71901,71870: 71902,71871: 71903,93760: 93792,
	93761: 93793,93762: 93794,93763: 93795,93764: 93796,93765: 93797,93766: 93798,
	93767: 93799,93768: 93800,93769: 93801,93770: 93802,93771: 93803,93772: 93804,
	93773: 93805,93774: 93806,93775: 93807,93776: 93808,93777: 93809,93778: 93810,
	93779: 93811,93780: 93812,93781: 93813,93782: 93814,93783: 93815,93784: 93816,
	93785: 93817,93786: 93818,93787: 93819,93788: 93820,93789: 93821,93790: 93822,
	93791: 93823,125184: 125218,125185: 125219,125186: 125220,125187: 125221,125188: 125222,
	125189: 125223,125190: 125224,125191: 125225,125192: 125226,125193: 125227,125194: 125228,
	125195: 125229,125196: 125230,125197: 125231,125198: 125232,125199: 125233,125200: 125234,
	125201: 125235,125202: 125236,125203: 125237,125204: 125238,125205: 125239,125206: 125240,
	125207: 125241,125208: 125242,125209: 125243,125210: 125244,125211: 125245,125212: 125246,
	125213: 125247,125214: 125248,125215: 125249,125216: 125250,125217: 125251,
};

function go2jsStringsToUpper(s) {
	let out = "";
	for (const part of go2jsStringRuneParts(s)) {
		const mapped = go2jsCaseUpper[part.rune];
		out += String.fromCodePoint(mapped === undefined ? part.rune : mapped);
	}
	return out;
}

function go2jsStringsToLower(s) {
	let out = "";
	for (const part of go2jsStringRuneParts(s)) {
		const cp = part.rune;
		const mapped = go2jsCaseLower[cp];
		out += String.fromCodePoint(mapped === undefined ? cp : mapped);
	}
	return out;
}

function go2jsStringsTrimSpace(s) {
	return s.trim();
}

function go2jsStringsSplit(s, sep) {
	if (sep.length === 0) {
		return Array.from(s);
	}
	return s.split(sep);
}

function go2jsStringsRepeat(s, count) {
	return s.repeat(count);
}

function go2jsStringsFields(s) {
	return s.split(/\s+/).filter((part) => part.length > 0);
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
function go2jsQuoteBytes(value, asciiOnly, graphicOnly) {
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

		go2jsStrconvAppendEscapedRune(out, rune, '"', asciiOnly, graphicOnly);
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
	// A Go string is a run of bytes, and a byte that is not part of a valid UTF8
	// sequence has nowhere of its own to sit in a JavaScript string. Reading the
	// string back as the bytes it is made of lets each one be escaped as Go
	// escapes it, whether or not it spells a rune.
	return go2jsQuoteBytes(go2jsStringToBytes(value), asciiOnly, graphicOnly);
}

function go2jsStrconvQuote(s) {
	return go2jsStrconvQuoteWith(s, false, false);
}

// go2jsCanBackquote says whether a string can be written as a raw string, the
// way strconv.CanBackquote decides it: every rune must be printable, a
// backquote or a carriage return rules it out, and so does a byte that spells
// no rune.
function go2jsCanBackquote(text) {
	for (let i = 0; i < text.length; i++) {
		const unit = text.charCodeAt(i);

		if (unit === 0xFEFF || unit === 0x60 || unit === 0x0D) {
			return false;
		}

		if (unit === 0x09 || (unit >= 0x20 && unit <= 0x7E)) {
			continue;
		}

		if (unit >= 0xD800 && unit <= 0xDBFF) {
			// A leading surrogate is the top of a rune only when a trailing one
			// follows, and the rune it makes is printable unless it is the byte
			// order mark.
			const low = i + 1 < text.length ? text.charCodeAt(i + 1) : 0;

			if (low < 0xDC00 || low > 0xDFFF) {
				return false;
			}

			i++;
			continue;
		}

		if (unit >= 0xDC00 && unit <= 0xDFFF) {
			return false;
		}

		if (unit >= 0x80) {
			continue;
		}

		return false;
	}

	return true;
}

// go2jsTruncateRunes cuts a string to a number of runes, the way fmt's precision
// does before a quoted verb writes it, without splitting a rune in two.
function go2jsTruncateRunes(text, precision) {
	if (precision === null || precision === undefined || precision < 0) {
		return text;
	}

	let seen = 0;
	let i = 0;

	while (i < text.length && seen < precision) {
		const unit = text.charCodeAt(i);

		if (unit >= 0xD800 && unit <= 0xDBFF && i + 1 < text.length) {
			const low = text.charCodeAt(i + 1);

			if (low >= 0xDC00 && low <= 0xDFFF) {
				i += 2;
				seen++;
				continue;
			}
		}

		i++;
		seen++;
	}

	return text.slice(0, i);
}

// go2jsQuoteText writes a string under the quoted verb, where the sharp flag
// asks for a raw string when one can be written and the plus flag asks for the
// escapes an ASCII-only literal needs. The precision cuts the string to that
// many runes before it is written.
function go2jsQuoteText(text, flags, precision) {
	text = go2jsTruncateRunes(text, precision);

	if (flags.includes("#") && go2jsCanBackquote(text)) {
		const backquote = String.fromCharCode(96);

		return backquote + text + backquote;
	}

	return go2jsStrconvQuoteWith(text, flags.includes("+"), false);
}

// go2jsQuoteBytesWithFlags writes the bytes of a string for the quoted verb,
// keeping a byte that spells no rune as the escape it needs rather than reading
// it back as the replacement character.
function go2jsQuoteBytesWithFlags(value, flags, precision) {
	// A precision cuts the bytes to that many runes before they are written, so
	// they are read back as the string they spell and cut there.
	if (precision !== null && precision !== undefined && precision >= 0) {
		value = go2jsStringToBytes(go2jsTruncateRunes(go2jsBytesToString(value), precision));
	}

	const text = go2jsBytesToString(value);

	if (flags.includes("#") && go2jsCanBackquote(text)) {
		const backquote = String.fromCharCode(96);

		return backquote + text + backquote;
	}

	return go2jsQuoteBytes(value, flags.includes("+"), false);
}

function go2jsStrconvQuoteRuneWith(value, asciiOnly, graphicOnly) {
	const out = ["'"];

	go2jsStrconvAppendEscapedRune(out, Number(value), "'", asciiOnly, graphicOnly);
	out.push("'");

	return out.join("");
}

function go2jsSortInts(values) {
	if (values === null || values === undefined) {
		return;
	}

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

	// An interface holds the value it was given next to the name of that
	// value's type, and the methods belong to the name rather than to the box,
	// so the box is opened before the method is looked for.
	if (target.__go2js_interface === true) {
		return go2jsSortMethod(target.value, target.type, method);
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

// The kinds sort names are a slice underneath, so a value of one is a plain
// array with nowhere of its own to keep its methods. The methods are reached
// through the table by the name of the type, which is how an interface that
// holds one calls them, and how reflect sees them.
function go2jsSortSliceLess(values, i, j, stringly, floatly) {
	const left = values[i];
	const right = values[j];

	if (stringly) {
		return String(left) < String(right);
	}

	if (floatly) {
		return left < right || (Number.isNaN(Number(left)) && !Number.isNaN(Number(right)));
	}

	return Number(left) < Number(right);
}

function go2jsSortSliceSwap(values, i, j) {
	const held = values[i];
	values[i] = values[j];
	values[j] = held;
}

go2jsRegisterMethod("IntSlice.Len", function(values) { return go2jsLen(values); });

go2jsRegisterMethod("IntSlice.Less", function(values, i, j) { return go2jsSortSliceLess(values, i, j, false, false); });

go2jsRegisterMethod("IntSlice.Swap", go2jsSortSliceSwap);

go2jsRegisterMethod("StringSlice.Len", function(values) { return go2jsLen(values); });

go2jsRegisterMethod("StringSlice.Less", function(values, i, j) { return go2jsSortSliceLess(values, i, j, true, false); });

go2jsRegisterMethod("StringSlice.Swap", go2jsSortSliceSwap);

go2jsRegisterMethod("Float64Slice.Len", function(values) { return go2jsLen(values); });

go2jsRegisterMethod("Float64Slice.Less", function(values, i, j) { return go2jsSortSliceLess(values, i, j, false, true); });

go2jsRegisterMethod("Float64Slice.Swap", go2jsSortSliceSwap);

function go2jsUnicodeCodePoint(value) {
	if (typeof value === "string") {
		const runes = Array.from(value);
		return runes.length === 0 ? -1 : runes[0].codePointAt(0);
	}

	if (!Number.isInteger(value)) {
		return -1;
	}

	return value;
}

function go2jsUnicodeRune(value) {
	const r = go2jsUnicodeCodePoint(value);

	if (r < 0 || r > 0x10ffff || (r >= 0xd800 && r <= 0xdfff)) {
		return "";
	}

	return String.fromCodePoint(r);
}

function go2jsUnicodeIsDigit(value) {
	const ch = go2jsUnicodeRune(value);
	return ch !== "" && /^\p{Nd}$/u.test(ch);
}

function go2jsUnicodeIsLetter(value) {
	const ch = go2jsUnicodeRune(value);
	return ch !== "" && /^\p{L}$/u.test(ch);
}

function go2jsUnicodeIsSpace(value) {
	const ch = go2jsUnicodeRune(value);
	return ch !== "" && /^\p{White_Space}$/u.test(ch);
}

const go2jsCaseTitle = {
	97: 65,98: 66,99: 67,100: 68,101: 69,102: 70,
	103: 71,104: 72,105: 73,106: 74,107: 75,108: 76,
	109: 77,110: 78,111: 79,112: 80,113: 81,114: 82,
	115: 83,116: 84,117: 85,118: 86,119: 87,120: 88,
	121: 89,122: 90,181: 924,224: 192,225: 193,226: 194,
	227: 195,228: 196,229: 197,230: 198,231: 199,232: 200,
	233: 201,234: 202,235: 203,236: 204,237: 205,238: 206,
	239: 207,240: 208,241: 209,242: 210,243: 211,244: 212,
	245: 213,246: 214,248: 216,249: 217,250: 218,251: 219,
	252: 220,253: 221,254: 222,255: 376,257: 256,259: 258,
	261: 260,263: 262,265: 264,267: 266,269: 268,271: 270,
	273: 272,275: 274,277: 276,279: 278,281: 280,283: 282,
	285: 284,287: 286,289: 288,291: 290,293: 292,295: 294,
	297: 296,299: 298,301: 300,303: 302,305: 73,307: 306,
	309: 308,311: 310,314: 313,316: 315,318: 317,320: 319,
	322: 321,324: 323,326: 325,328: 327,331: 330,333: 332,
	335: 334,337: 336,339: 338,341: 340,343: 342,345: 344,
	347: 346,349: 348,351: 350,353: 352,355: 354,357: 356,
	359: 358,361: 360,363: 362,365: 364,367: 366,369: 368,
	371: 370,373: 372,375: 374,378: 377,380: 379,382: 381,
	383: 83,384: 579,387: 386,389: 388,392: 391,396: 395,
	402: 401,405: 502,409: 408,410: 573,414: 544,417: 416,
	419: 418,421: 420,424: 423,429: 428,432: 431,436: 435,
	438: 437,441: 440,445: 444,447: 503,452: 453,454: 453,
	455: 456,457: 456,458: 459,460: 459,462: 461,464: 463,
	466: 465,468: 467,470: 469,472: 471,474: 473,476: 475,
	477: 398,479: 478,481: 480,483: 482,485: 484,487: 486,
	489: 488,491: 490,493: 492,495: 494,497: 498,499: 498,
	501: 500,505: 504,507: 506,509: 508,511: 510,513: 512,
	515: 514,517: 516,519: 518,521: 520,523: 522,525: 524,
	527: 526,529: 528,531: 530,533: 532,535: 534,537: 536,
	539: 538,541: 540,543: 542,547: 546,549: 548,551: 550,
	553: 552,555: 554,557: 556,559: 558,561: 560,563: 562,
	572: 571,575: 11390,576: 11391,578: 577,583: 582,585: 584,
	587: 586,589: 588,591: 590,592: 11375,593: 11373,594: 11376,
	595: 385,596: 390,598: 393,599: 394,601: 399,603: 400,
	604: 42923,608: 403,609: 42924,611: 404,613: 42893,614: 42922,
	616: 407,617: 406,618: 42926,619: 11362,620: 42925,623: 412,
	625: 11374,626: 413,629: 415,637: 11364,640: 422,642: 42949,
	643: 425,647: 42929,648: 430,649: 580,650: 433,651: 434,
	652: 581,658: 439,669: 42930,670: 42928,837: 921,881: 880,
	883: 882,887: 886,891: 1021,892: 1022,893: 1023,940: 902,
	941: 904,942: 905,943: 906,945: 913,946: 914,947: 915,
	948: 916,949: 917,950: 918,951: 919,952: 920,953: 921,
	954: 922,955: 923,956: 924,957: 925,958: 926,959: 927,
	960: 928,961: 929,962: 931,963: 931,964: 932,965: 933,
	966: 934,967: 935,968: 936,969: 937,970: 938,971: 939,
	972: 908,973: 910,974: 911,976: 914,977: 920,981: 934,
	982: 928,983: 975,985: 984,987: 986,989: 988,991: 990,
	993: 992,995: 994,997: 996,999: 998,1001: 1000,1003: 1002,
	1005: 1004,1007: 1006,1008: 922,1009: 929,1010: 1017,1011: 895,
	1013: 917,1016: 1015,1019: 1018,1072: 1040,1073: 1041,1074: 1042,
	1075: 1043,1076: 1044,1077: 1045,1078: 1046,1079: 1047,1080: 1048,
	1081: 1049,1082: 1050,1083: 1051,1084: 1052,1085: 1053,1086: 1054,
	1087: 1055,1088: 1056,1089: 1057,1090: 1058,1091: 1059,1092: 1060,
	1093: 1061,1094: 1062,1095: 1063,1096: 1064,1097: 1065,1098: 1066,
	1099: 1067,1100: 1068,1101: 1069,1102: 1070,1103: 1071,1104: 1024,
	1105: 1025,1106: 1026,1107: 1027,1108: 1028,1109: 1029,1110: 1030,
	1111: 1031,1112: 1032,1113: 1033,1114: 1034,1115: 1035,1116: 1036,
	1117: 1037,1118: 1038,1119: 1039,1121: 1120,1123: 1122,1125: 1124,
	1127: 1126,1129: 1128,1131: 1130,1133: 1132,1135: 1134,1137: 1136,
	1139: 1138,1141: 1140,1143: 1142,1145: 1144,1147: 1146,1149: 1148,
	1151: 1150,1153: 1152,1163: 1162,1165: 1164,1167: 1166,1169: 1168,
	1171: 1170,1173: 1172,1175: 1174,1177: 1176,1179: 1178,1181: 1180,
	1183: 1182,1185: 1184,1187: 1186,1189: 1188,1191: 1190,1193: 1192,
	1195: 1194,1197: 1196,1199: 1198,1201: 1200,1203: 1202,1205: 1204,
	1207: 1206,1209: 1208,1211: 1210,1213: 1212,1215: 1214,1218: 1217,
	1220: 1219,1222: 1221,1224: 1223,1226: 1225,1228: 1227,1230: 1229,
	1231: 1216,1233: 1232,1235: 1234,1237: 1236,1239: 1238,1241: 1240,
	1243: 1242,1245: 1244,1247: 1246,1249: 1248,1251: 1250,1253: 1252,
	1255: 1254,1257: 1256,1259: 1258,1261: 1260,1263: 1262,1265: 1264,
	1267: 1266,1269: 1268,1271: 1270,1273: 1272,1275: 1274,1277: 1276,
	1279: 1278,1281: 1280,1283: 1282,1285: 1284,1287: 1286,1289: 1288,
	1291: 1290,1293: 1292,1295: 1294,1297: 1296,1299: 1298,1301: 1300,
	1303: 1302,1305: 1304,1307: 1306,1309: 1308,1311: 1310,1313: 1312,
	1315: 1314,1317: 1316,1319: 1318,1321: 1320,1323: 1322,1325: 1324,
	1327: 1326,1377: 1329,1378: 1330,1379: 1331,1380: 1332,1381: 1333,
	1382: 1334,1383: 1335,1384: 1336,1385: 1337,1386: 1338,1387: 1339,
	1388: 1340,1389: 1341,1390: 1342,1391: 1343,1392: 1344,1393: 1345,
	1394: 1346,1395: 1347,1396: 1348,1397: 1349,1398: 1350,1399: 1351,
	1400: 1352,1401: 1353,1402: 1354,1403: 1355,1404: 1356,1405: 1357,
	1406: 1358,1407: 1359,1408: 1360,1409: 1361,1410: 1362,1411: 1363,
	1412: 1364,1413: 1365,1414: 1366,5112: 5104,5113: 5105,5114: 5106,
	5115: 5107,5116: 5108,5117: 5109,7296: 1042,7297: 1044,7298: 1054,
	7299: 1057,7300: 1058,7301: 1058,7302: 1066,7303: 1122,7304: 42570,
	7545: 42877,7549: 11363,7566: 42950,7681: 7680,7683: 7682,7685: 7684,
	7687: 7686,7689: 7688,7691: 7690,7693: 7692,7695: 7694,7697: 7696,
	7699: 7698,7701: 7700,7703: 7702,7705: 7704,7707: 7706,7709: 7708,
	7711: 7710,7713: 7712,7715: 7714,7717: 7716,7719: 7718,7721: 7720,
	7723: 7722,7725: 7724,7727: 7726,7729: 7728,7731: 7730,7733: 7732,
	7735: 7734,7737: 7736,7739: 7738,7741: 7740,7743: 7742,7745: 7744,
	7747: 7746,7749: 7748,7751: 7750,7753: 7752,7755: 7754,7757: 7756,
	7759: 7758,7761: 7760,7763: 7762,7765: 7764,7767: 7766,7769: 7768,
	7771: 7770,7773: 7772,7775: 7774,7777: 7776,7779: 7778,7781: 7780,
	7783: 7782,7785: 7784,7787: 7786,7789: 7788,7791: 7790,7793: 7792,
	7795: 7794,7797: 7796,7799: 7798,7801: 7800,7803: 7802,7805: 7804,
	7807: 7806,7809: 7808,7811: 7810,7813: 7812,7815: 7814,7817: 7816,
	7819: 7818,7821: 7820,7823: 7822,7825: 7824,7827: 7826,7829: 7828,
	7835: 7776,7841: 7840,7843: 7842,7845: 7844,7847: 7846,7849: 7848,
	7851: 7850,7853: 7852,7855: 7854,7857: 7856,7859: 7858,7861: 7860,
	7863: 7862,7865: 7864,7867: 7866,7869: 7868,7871: 7870,7873: 7872,
	7875: 7874,7877: 7876,7879: 7878,7881: 7880,7883: 7882,7885: 7884,
	7887: 7886,7889: 7888,7891: 7890,7893: 7892,7895: 7894,7897: 7896,
	7899: 7898,7901: 7900,7903: 7902,7905: 7904,7907: 7906,7909: 7908,
	7911: 7910,7913: 7912,7915: 7914,7917: 7916,7919: 7918,7921: 7920,
	7923: 7922,7925: 7924,7927: 7926,7929: 7928,7931: 7930,7933: 7932,
	7935: 7934,7936: 7944,7937: 7945,7938: 7946,7939: 7947,7940: 7948,
	7941: 7949,7942: 7950,7943: 7951,7952: 7960,7953: 7961,7954: 7962,
	7955: 7963,7956: 7964,7957: 7965,7968: 7976,7969: 7977,7970: 7978,
	7971: 7979,7972: 7980,7973: 7981,7974: 7982,7975: 7983,7984: 7992,
	7985: 7993,7986: 7994,7987: 7995,7988: 7996,7989: 7997,7990: 7998,
	7991: 7999,8000: 8008,8001: 8009,8002: 8010,8003: 8011,8004: 8012,
	8005: 8013,8017: 8025,8019: 8027,8021: 8029,8023: 8031,8032: 8040,
	8033: 8041,8034: 8042,8035: 8043,8036: 8044,8037: 8045,8038: 8046,
	8039: 8047,8048: 8122,8049: 8123,8050: 8136,8051: 8137,8052: 8138,
	8053: 8139,8054: 8154,8055: 8155,8056: 8184,8057: 8185,8058: 8170,
	8059: 8171,8060: 8186,8061: 8187,8064: 8072,8065: 8073,8066: 8074,
	8067: 8075,8068: 8076,8069: 8077,8070: 8078,8071: 8079,8080: 8088,
	8081: 8089,8082: 8090,8083: 8091,8084: 8092,8085: 8093,8086: 8094,
	8087: 8095,8096: 8104,8097: 8105,8098: 8106,8099: 8107,8100: 8108,
	8101: 8109,8102: 8110,8103: 8111,8112: 8120,8113: 8121,8115: 8124,
	8126: 921,8131: 8140,8144: 8152,8145: 8153,8160: 8168,8161: 8169,
	8165: 8172,8179: 8188,8526: 8498,8560: 8544,8561: 8545,8562: 8546,
	8563: 8547,8564: 8548,8565: 8549,8566: 8550,8567: 8551,8568: 8552,
	8569: 8553,8570: 8554,8571: 8555,8572: 8556,8573: 8557,8574: 8558,
	8575: 8559,8580: 8579,9424: 9398,9425: 9399,9426: 9400,9427: 9401,
	9428: 9402,9429: 9403,9430: 9404,9431: 9405,9432: 9406,9433: 9407,
	9434: 9408,9435: 9409,9436: 9410,9437: 9411,9438: 9412,9439: 9413,
	9440: 9414,9441: 9415,9442: 9416,9443: 9417,9444: 9418,9445: 9419,
	9446: 9420,9447: 9421,9448: 9422,9449: 9423,11312: 11264,11313: 11265,
	11314: 11266,11315: 11267,11316: 11268,11317: 11269,11318: 11270,11319: 11271,
	11320: 11272,11321: 11273,11322: 11274,11323: 11275,11324: 11276,11325: 11277,
	11326: 11278,11327: 11279,11328: 11280,11329: 11281,11330: 11282,11331: 11283,
	11332: 11284,11333: 11285,11334: 11286,11335: 11287,11336: 11288,11337: 11289,
	11338: 11290,11339: 11291,11340: 11292,11341: 11293,11342: 11294,11343: 11295,
	11344: 11296,11345: 11297,11346: 11298,11347: 11299,11348: 11300,11349: 11301,
	11350: 11302,11351: 11303,11352: 11304,11353: 11305,11354: 11306,11355: 11307,
	11356: 11308,11357: 11309,11358: 11310,11359: 11311,11361: 11360,11365: 570,
	11366: 574,11368: 11367,11370: 11369,11372: 11371,11379: 11378,11382: 11381,
	11393: 11392,11395: 11394,11397: 11396,11399: 11398,11401: 11400,11403: 11402,
	11405: 11404,11407: 11406,11409: 11408,11411: 11410,11413: 11412,11415: 11414,
	11417: 11416,11419: 11418,11421: 11420,11423: 11422,11425: 11424,11427: 11426,
	11429: 11428,11431: 11430,11433: 11432,11435: 11434,11437: 11436,11439: 11438,
	11441: 11440,11443: 11442,11445: 11444,11447: 11446,11449: 11448,11451: 11450,
	11453: 11452,11455: 11454,11457: 11456,11459: 11458,11461: 11460,11463: 11462,
	11465: 11464,11467: 11466,11469: 11468,11471: 11470,11473: 11472,11475: 11474,
	11477: 11476,11479: 11478,11481: 11480,11483: 11482,11485: 11484,11487: 11486,
	11489: 11488,11491: 11490,11500: 11499,11502: 11501,11507: 11506,11520: 4256,
	11521: 4257,11522: 4258,11523: 4259,11524: 4260,11525: 4261,11526: 4262,
	11527: 4263,11528: 4264,11529: 4265,11530: 4266,11531: 4267,11532: 4268,
	11533: 4269,11534: 4270,11535: 4271,11536: 4272,11537: 4273,11538: 4274,
	11539: 4275,11540: 4276,11541: 4277,11542: 4278,11543: 4279,11544: 4280,
	11545: 4281,11546: 4282,11547: 4283,11548: 4284,11549: 4285,11550: 4286,
	11551: 4287,11552: 4288,11553: 4289,11554: 4290,11555: 4291,11556: 4292,
	11557: 4293,11559: 4295,11565: 4301,42561: 42560,42563: 42562,42565: 42564,
	42567: 42566,42569: 42568,42571: 42570,42573: 42572,42575: 42574,42577: 42576,
	42579: 42578,42581: 42580,42583: 42582,42585: 42584,42587: 42586,42589: 42588,
	42591: 42590,42593: 42592,42595: 42594,42597: 42596,42599: 42598,42601: 42600,
	42603: 42602,42605: 42604,42625: 42624,42627: 42626,42629: 42628,42631: 42630,
	42633: 42632,42635: 42634,42637: 42636,42639: 42638,42641: 42640,42643: 42642,
	42645: 42644,42647: 42646,42649: 42648,42651: 42650,42787: 42786,42789: 42788,
	42791: 42790,42793: 42792,42795: 42794,42797: 42796,42799: 42798,42803: 42802,
	42805: 42804,42807: 42806,42809: 42808,42811: 42810,42813: 42812,42815: 42814,
	42817: 42816,42819: 42818,42821: 42820,42823: 42822,42825: 42824,42827: 42826,
	42829: 42828,42831: 42830,42833: 42832,42835: 42834,42837: 42836,42839: 42838,
	42841: 42840,42843: 42842,42845: 42844,42847: 42846,42849: 42848,42851: 42850,
	42853: 42852,42855: 42854,42857: 42856,42859: 42858,42861: 42860,42863: 42862,
	42874: 42873,42876: 42875,42879: 42878,42881: 42880,42883: 42882,42885: 42884,
	42887: 42886,42892: 42891,42897: 42896,42899: 42898,42900: 42948,42903: 42902,
	42905: 42904,42907: 42906,42909: 42908,42911: 42910,42913: 42912,42915: 42914,
	42917: 42916,42919: 42918,42921: 42920,42933: 42932,42935: 42934,42937: 42936,
	42939: 42938,42941: 42940,42943: 42942,42945: 42944,42947: 42946,42952: 42951,
	42954: 42953,42961: 42960,42967: 42966,42969: 42968,42998: 42997,43859: 42931,
	43888: 5024,43889: 5025,43890: 5026,43891: 5027,43892: 5028,43893: 5029,
	43894: 5030,43895: 5031,43896: 5032,43897: 5033,43898: 5034,43899: 5035,
	43900: 5036,43901: 5037,43902: 5038,43903: 5039,43904: 5040,43905: 5041,
	43906: 5042,43907: 5043,43908: 5044,43909: 5045,43910: 5046,43911: 5047,
	43912: 5048,43913: 5049,43914: 5050,43915: 5051,43916: 5052,43917: 5053,
	43918: 5054,43919: 5055,43920: 5056,43921: 5057,43922: 5058,43923: 5059,
	43924: 5060,43925: 5061,43926: 5062,43927: 5063,43928: 5064,43929: 5065,
	43930: 5066,43931: 5067,43932: 5068,43933: 5069,43934: 5070,43935: 5071,
	43936: 5072,43937: 5073,43938: 5074,43939: 5075,43940: 5076,43941: 5077,
	43942: 5078,43943: 5079,43944: 5080,43945: 5081,43946: 5082,43947: 5083,
	43948: 5084,43949: 5085,43950: 5086,43951: 5087,43952: 5088,43953: 5089,
	43954: 5090,43955: 5091,43956: 5092,43957: 5093,43958: 5094,43959: 5095,
	43960: 5096,43961: 5097,43962: 5098,43963: 5099,43964: 5100,43965: 5101,
	43966: 5102,43967: 5103,65345: 65313,65346: 65314,65347: 65315,65348: 65316,
	65349: 65317,65350: 65318,65351: 65319,65352: 65320,65353: 65321,65354: 65322,
	65355: 65323,65356: 65324,65357: 65325,65358: 65326,65359: 65327,65360: 65328,
	65361: 65329,65362: 65330,65363: 65331,65364: 65332,65365: 65333,65366: 65334,
	65367: 65335,65368: 65336,65369: 65337,65370: 65338,66600: 66560,66601: 66561,
	66602: 66562,66603: 66563,66604: 66564,66605: 66565,66606: 66566,66607: 66567,
	66608: 66568,66609: 66569,66610: 66570,66611: 66571,66612: 66572,66613: 66573,
	66614: 66574,66615: 66575,66616: 66576,66617: 66577,66618: 66578,66619: 66579,
	66620: 66580,66621: 66581,66622: 66582,66623: 66583,66624: 66584,66625: 66585,
	66626: 66586,66627: 66587,66628: 66588,66629: 66589,66630: 66590,66631: 66591,
	66632: 66592,66633: 66593,66634: 66594,66635: 66595,66636: 66596,66637: 66597,
	66638: 66598,66639: 66599,66776: 66736,66777: 66737,66778: 66738,66779: 66739,
	66780: 66740,66781: 66741,66782: 66742,66783: 66743,66784: 66744,66785: 66745,
	66786: 66746,66787: 66747,66788: 66748,66789: 66749,66790: 66750,66791: 66751,
	66792: 66752,66793: 66753,66794: 66754,66795: 66755,66796: 66756,66797: 66757,
	66798: 66758,66799: 66759,66800: 66760,66801: 66761,66802: 66762,66803: 66763,
	66804: 66764,66805: 66765,66806: 66766,66807: 66767,66808: 66768,66809: 66769,
	66810: 66770,66811: 66771,66967: 66928,66968: 66929,66969: 66930,66970: 66931,
	66971: 66932,66972: 66933,66973: 66934,66974: 66935,66975: 66936,66976: 66937,
	66977: 66938,66979: 66940,66980: 66941,66981: 66942,66982: 66943,66983: 66944,
	66984: 66945,66985: 66946,66986: 66947,66987: 66948,66988: 66949,66989: 66950,
	66990: 66951,66991: 66952,66992: 66953,66993: 66954,66995: 66956,66996: 66957,
	66997: 66958,66998: 66959,66999: 66960,67000: 66961,67001: 66962,67003: 66964,
	67004: 66965,68800: 68736,68801: 68737,68802: 68738,68803: 68739,68804: 68740,
	68805: 68741,68806: 68742,68807: 68743,68808: 68744,68809: 68745,68810: 68746,
	68811: 68747,68812: 68748,68813: 68749,68814: 68750,68815: 68751,68816: 68752,
	68817: 68753,68818: 68754,68819: 68755,68820: 68756,68821: 68757,68822: 68758,
	68823: 68759,68824: 68760,68825: 68761,68826: 68762,68827: 68763,68828: 68764,
	68829: 68765,68830: 68766,68831: 68767,68832: 68768,68833: 68769,68834: 68770,
	68835: 68771,68836: 68772,68837: 68773,68838: 68774,68839: 68775,68840: 68776,
	68841: 68777,68842: 68778,68843: 68779,68844: 68780,68845: 68781,68846: 68782,
	68847: 68783,68848: 68784,68849: 68785,68850: 68786,71872: 71840,71873: 71841,
	71874: 71842,71875: 71843,71876: 71844,71877: 71845,71878: 71846,71879: 71847,
	71880: 71848,71881: 71849,71882: 71850,71883: 71851,71884: 71852,71885: 71853,
	71886: 71854,71887: 71855,71888: 71856,71889: 71857,71890: 71858,71891: 71859,
	71892: 71860,71893: 71861,71894: 71862,71895: 71863,71896: 71864,71897: 71865,
	71898: 71866,71899: 71867,71900: 71868,71901: 71869,71902: 71870,71903: 71871,
	93792: 93760,93793: 93761,93794: 93762,93795: 93763,93796: 93764,93797: 93765,
	93798: 93766,93799: 93767,93800: 93768,93801: 93769,93802: 93770,93803: 93771,
	93804: 93772,93805: 93773,93806: 93774,93807: 93775,93808: 93776,93809: 93777,
	93810: 93778,93811: 93779,93812: 93780,93813: 93781,93814: 93782,93815: 93783,
	93816: 93784,93817: 93785,93818: 93786,93819: 93787,93820: 93788,93821: 93789,
	93822: 93790,93823: 93791,125218: 125184,125219: 125185,125220: 125186,125221: 125187,
	125222: 125188,125223: 125189,125224: 125190,125225: 125191,125226: 125192,125227: 125193,
	125228: 125194,125229: 125195,125230: 125196,125231: 125197,125232: 125198,125233: 125199,
	125234: 125200,125235: 125201,125236: 125202,125237: 125203,125238: 125204,125239: 125205,
	125240: 125206,125241: 125207,125242: 125208,125243: 125209,125244: 125210,125245: 125211,
	125246: 125212,125247: 125213,125248: 125214,125249: 125215,125250: 125216,125251: 125217,
};

go2jsRegisterMethod("encoding/json.Delim.String", go2jsJSONDelimText);

const go2jsSliceMeta = new WeakMap();

const go2jsArrayMark = new WeakSet();

// go2jsSliceIndex answers the whole number a property name stands for when it
// names one of a slice's elements, and null when it names anything else. Only
// a canonical index is taken, so "length", "1x" and "01" name properties
// rather than elements, and the name is read digit by digit rather than
// turned into a number and back again.
function go2jsSliceIndex(property) {
	if (typeof property !== "string" || property.length === 0) {
		return null;
	}

	let start = 0;

	if (property.charCodeAt(0) === 45) {
		if (property.length === 1) {
			return null;
		}

		start = 1;
	}

	if (property.charCodeAt(start) === 48 && property.length > start + 1) {
		return null;
	}

	let index = 0;

	for (let i = start; i < property.length; i++) {
		const code = property.charCodeAt(i);

		if (code < 48 || code > 57) {
			return null;
		}

		index = index * 10 + (code - 48);
	}

	return property.charCodeAt(0) === 45 ? -index : index;
}

function go2jsSliceView(data, offset, length, capacity) {
	if (data === null || data === undefined) {
		return null;
	}

	if (offset < 0 || length < 0 || capacity < length || offset + capacity > data.length) {
		throw new RangeError("invalid slice bounds");
	}

	const state = {
		data,
		offset,
		length,
		capacity
	};

	const target = [];

	const proxy = new Proxy(target, {
		get(target, property, receiver) {
			if (property === "__go2js_slice") {
				return true;
			}

			if (property === "length") {
				return state.length;
			}

			if (property === Symbol.iterator) {
				return function*() {
					for (let i = 0; i < state.length; i++) {
						yield state.data[state.offset + i];
					}
				};
			}

			if (property === "entries") {
				return function*() {
					for (let i = 0; i < state.length; i++) {
						yield [i, state.data[state.offset + i]];
					}
				};
			}

			const index = go2jsSliceIndex(property);
			if (index !== null) {
				if (index < 0 || index >= state.length) {
					return undefined;
				}
				return state.data[state.offset + index];
			}

			return Reflect.get(target, property, receiver);
		},

		has(target, property) {
			const index = go2jsSliceIndex(property);

			if (index !== null) {
				return index >= 0 && index < state.length;
			}

			return Reflect.has(target, property);
		},

		ownKeys(target) {
			const keys = [];

			for (let i = 0; i < state.length; i++) {
				keys.push(String(i));
			}

			return keys.concat(Reflect.ownKeys(target).filter(key => key !== "length"));
		},

		getOwnPropertyDescriptor(target, property) {
			const index = go2jsSliceIndex(property);

			if (index !== null) {
				if (index < 0 || index >= state.length) {
					return undefined;
				}

				return {
					value: state.data[state.offset + index],
					writable: true,
					enumerable: true,
					configurable: true
				};
			}

			return Reflect.getOwnPropertyDescriptor(target, property);
		},

		set(target, property, value) {
			const index = go2jsSliceIndex(property);
			if (index !== null) {
				if (index < 0) {
					throw new go2jsNativeRangeError(go2jsRuntimeErrorPrefix + "index out of range [" + index + "]");
				}

				if (index >= state.length) {
					throw new go2jsNativeRangeError(go2jsRuntimeErrorPrefix + "index out of range [" + index + "] with length " + state.length);
				}
				state.data[state.offset + index] = value;
				return true;
			}

			if (property === "length") {
				throw new TypeError("cannot assign slice length");
			}

			return Reflect.set(target, property, value);
		}
	});

	go2jsSliceMeta.set(proxy, state);
	return proxy;
}

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

function go2jsSliceMake(length, capacity, zero, sharedZero) {
	length = Math.trunc(Number(length));
	capacity = Math.trunc(Number(capacity));

	if (length > 4294967295) {
		throw new RangeError(go2jsRuntimeErrorPrefix + "makeslice: len out of range");
	}

	if (capacity > 4294967295) {
		throw new RangeError(go2jsRuntimeErrorPrefix + "makeslice: cap out of range");
	}

	// A slice can only be made with a length that is not negative and a
	// capacity that can hold it, and Go says which of the two went wrong.
	if (length < 0) {
		throw new RangeError(go2jsRuntimeErrorPrefix + "makeslice: len out of range");
	}

	if (capacity < 0 || length > capacity) {
		throw new RangeError(go2jsRuntimeErrorPrefix + "makeslice: cap out of range");
	}

	const data = new Array(capacity);

	// A zero value that stands for itself fills the whole room at once, which
	// keeps a slice of numbers from making a function call for every element.
	// One that has to be made fresh, like a struct or an array, is asked for
	// once for each place it fills.
	if (sharedZero) {
		data.fill(zero);
	} else {
		for (let i = 0; i < capacity; i++) {
			data[i] = zero();
		}
	}

	return go2jsSliceView(data, 0, length, capacity);
}

function go2jsSliceAppend(value, ...items) {
	if (value === null || value === undefined) {
		// Nothing added to nothing is still nil.
		if (items.length === 0) {
			return null;
		}

		value = go2jsSliceView([], 0, 0, 0);
	}

	const state = go2jsSliceState(value);

	if (!state) {
		throw new TypeError("go2jsSliceAppend expects a slice");
	}

	if (items.length === 0) {
		if (go2jsSliceMeta.has(value)) {
			return go2jsSliceView(
				state.data,
				state.offset,
				state.length,
				state.capacity
			);
		}
		return value;
	}

	const required = state.length + items.length;

	if (required <= state.capacity) {
		for (let i = 0; i < items.length; i++) {
			state.data[state.offset + state.length + i] = items[i];
		}

		return go2jsSliceView(
			state.data,
			state.offset,
			required,
			state.capacity
		);
	}

	let capacity = state.capacity > 0 ? state.capacity * 2 : 1;

	while (capacity < required) {
		capacity *= 2;
	}

	const data = Array.from(
		{length: capacity},
		() => undefined
	);

	for (let i = 0; i < state.length; i++) {
		data[i] = go2jsStructFieldCopy(state.data[state.offset + i]);
	}

	for (let i = 0; i < items.length; i++) {
		data[state.length + i] = go2jsStructFieldCopy(items[i]);
	}

	return go2jsSliceView(data, 0, required, capacity);
}

function go2jsSliceRange(value, low, high, max, isArray) {
	if (value === null || value === undefined) {
		if ((low === undefined || low === 0) &&
			(high === undefined || high === 0) &&
			(max === undefined || max === 0)) {
			return null;
		}

		throw new RangeError("slice of nil slice");
	}

	const state = go2jsSliceState(value);
	if (!state) {
		throw new TypeError("value is not sliceable");
	}

	const start = low === undefined ? 0 : Math.trunc(low);
	const end = high === undefined ? state.length : Math.trunc(high);
	const limit = max === undefined ? state.capacity : Math.trunc(max);

	// The bounds fail in the order Go asks about them, so a slice that crosses
	// itself shows its own two indexes before the room it runs past, and the
	// message ends with the length when the values lie in an array or the
	// capacity when they lie in a slice.
	const prefix = go2jsRuntimeErrorPrefix + "slice bounds out of range ";
	const room = isArray ? state.length : state.capacity;
	const width = isArray ? "length" : "capacity";

	if (start < 0 || start > end) {
		if (max !== undefined) {
			throw new RangeError(prefix +
				(start < 0 ? "[" + start + "::]" : "[" + start + ":" + end + ":]"));
		}

		throw new RangeError(prefix +
			(start < 0 ? "[" + start + ":]" : "[" + start + ":" + end + "]"));
	}

	if (max !== undefined) {
		if (end < 0 || end > limit) {
			throw new RangeError(prefix +
				(end < 0 ? "[:" + end + ":]" : "[:" + end + ":" + limit + "]"));
		}

		if (limit > room) {
			throw new RangeError(prefix +
				"[::" + limit + "] with " + width + " " + room);
		}
	} else if (end < 0 || end > limit) {
		throw new RangeError(prefix +
			(end < 0 ? "[:" + end + "]" : "[:" + end + "] with " + width + " " + room));
	}

	return go2jsSliceView(
		state.data,
		state.offset + start,
		end - start,
		limit - start
	);
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

	// A map or a set is storage reached by reference rather than a record of
	// fields to write out, so Go copies the reference and hands back the same
	// one. Copying one field by field would leave a hollow shell that no longer
	// knows what it holds.
	if (value instanceof go2jsNativeMap || value instanceof go2jsNativeSet) {
		return value;
	}

	const embedded = typeof value.__go2js_embedded === "undefined"
		? null
		: value.__go2js_embedded;

	const copy = Object.create(Object.getPrototypeOf(value));

	// The name a struct carries is not a field of it but what the struct is, so
	// it is kept rather than dropped: a copy that lost it would be written out
	// as nothing in particular, where %T and %#v have to say what it is.
	if (typeof value.__go2js_type_name === "string" && value.__go2js_type_name !== "") {
		Object.defineProperty(copy, "__go2js_type_name", {
			value: value.__go2js_type_name,
			writable: true,
			configurable: true,
			enumerable: false
		});
	}

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

// go2jsRangeCopies hands back the elements of a sequence as values of their
// own, which is what ranging over a slice or an array of records gives: the
// element is a copy, and writing to it is a write to the copy rather than a
// write to the element the sequence is holding. A sequence of numbers is left
// alone, because a number is a value already and copying it costs more than
// naming it.
function* go2jsRangeCopies(sequence) {
	for (const item of sequence) {
		yield go2jsStructFieldCopy(item);
	}
}

function go2jsRangeSequence(value) {
	if (typeof value === "function") {
		return go2jsRangeYielded(value);
	}

	if (go2jsNaNKeys === true && value instanceof go2jsNativeMap) {
		return go2jsRangeNaNKeys(value);
	}

	return value === null || value === undefined ? [] : value;
}

// go2jsRangeYielded runs a sequence of values, which is a function that hands
// each of its values to whoever asked for them, and keeps what it handed over.
// A sequence is run to its end before the first value is used, because a yield
// in the middle of a function cannot be left half finished and taken up again
// where it stopped, so a sequence that never ends is one this cannot walk.
// The sequence is written as a generator, so it is run here rather than called,
// and the callback it calls its values through answers with whether to go on.
function go2jsRangeYielded(sequence) {
	const handed = [];

	go2jsCallNow(sequence, null, [function (...items) {
		handed.push(items);

		return true;
	}]);

	return handed;
}

function go2jsRangeMap(value) {
	return value === null || value === undefined ? new go2jsNativeMap() : value;
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

// go2jsPreemptMark is what a loop that has run for a while hands up when it gives
// the turn to the other goroutines. It is told apart from the turn a goroutine
// gives up by waiting, because a loop that is only being fair has not waited, and
// so is not a step toward the limit on steps, and is not what a lone goroutine
// is to be called asleep for.
const go2jsPreemptMark = {};

var go2jsTicks = 0;

function* go2jsPreempt() {
	go2jsTicks = 0;

	yield go2jsPreemptMark;
}

var go2jsMainTask = null;

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
	// that is waiting on the same channel ready, so the select waiting for what
	// this one is now offering is woken to look again. A goroutine parked in a
	// receive is what a select's send case was waiting for, and a goroutine
	// parked in a send is what a select's receive case was waiting for.
	if (kind === "receiver" || kind === "sender") {
		go2jsWakeSelects(channels, kind === "receiver" ? "sender" : "receiver");
	}

	return go2jsParked;
}

// go2jsSelectWaits reports whether the parked select task is waiting with a case
// that sends on the channel (kind "sender") or receives from it (kind
// "receiver").
function go2jsSelectWaits(task, channel, kind) {
	const cases = kind === "sender" ? task.selectSends : task.selectRecvs;

	return cases !== undefined && cases.indexOf(channel) >= 0;
}

// go2jsWakeSelects makes runnable every select parked with a case of the kind
// named on one of the channels. A select whose other cases were already ready
// is not picked here: it looks again when it runs, and chooses then.
function go2jsWakeSelects(channels, kind) {
	const selects = [];

	for (const channel of channels) {
		if (channel === null || channel === undefined) {
			continue;
		}

		for (const entry of channel.blocked) {
			if (entry.parked === "select" && go2jsSelectWaits(entry, channel, kind)) {
				if (selects.indexOf(entry) < 0) {
					selects.push(entry);
				}
			}
		}
	}

	for (const entry of selects) {
		go2jsRunnable(entry);
	}
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
	task.selectSends = undefined;
	task.selectRecvs = undefined;
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

function go2jsHasBlocked(channel, kind) {
	for (const task of channel.blocked) {
		if (task.parked === kind) {
			return true;
		}
	}

	return false;
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

		if (step.value === go2jsPreemptMark) {
			go2jsWaitForWork();

			continue;
		}

		if (!go2jsWaitForWork()) {
			throw go2jsFatalError("all goroutines are asleep - deadlock!");
		}
	}
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

		// A goroutine that is only being kept from taking every turn is no reason
		// for a timer that has come due to wait, so they are looked at between turns.
		if (go2jsTimers.size > 0) {
			go2jsWakeDueTimers();
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

		if (step.value !== go2jsPreemptMark) {
			steps++;
		}

		if (step.done) {
			task.finished = true;
			go2jsLiveTasks--;

			// A Go program ends when main returns, whatever the other goroutines
			// are still doing.
			if (task === go2jsMainTask) {
				go2jsReady.length = 0;

				return steps + 1;
			}

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

// go2jsStackMegabytes is how big a stack the program asks to run on. It is set
// by the compiler when it is told to, and is nothing otherwise.
var go2jsStackMegabytes = 0;

// go2jsRelaunchWithStack runs the program again on a thread with the stack it
// asked for, and says so, so that this thread does not run main as well. It is
// only done for the script Node was started on, since that is the one script
// that can be started again, and a program that was loaded by another one is
// left to run where it is.
function go2jsRelaunchWithStack() {
	if (go2jsStackMegabytes <= 0 || typeof process === "undefined" || typeof process.getBuiltinModule !== "function") {
		return false;
	}

	if (process.versions === undefined || process.versions.node === undefined || typeof process.argv[1] !== "string") {
		return false;
	}

	const threads = process.getBuiltinModule("node:worker_threads");

	if (threads === undefined || threads === null || !threads.isMainThread) {
		return false;
	}

	let entry;

	try {
		const fs = process.getBuiltinModule("node:fs");
		const url = process.getBuiltinModule("node:url");

		entry = fs.realpathSync(process.argv[1]);

		const stack = String(new Error().stack);

		if (!stack.includes(entry) && !stack.includes(url.pathToFileURL(entry).href)) {
			return false;
		}
	} catch (error) {
		return false;
	}

	const worker = new threads.Worker(entry, {
		argv: process.argv.slice(2),
		resourceLimits: { stackSizeMb: go2jsStackMegabytes },
	});

	worker.on("error", function(error) {
		process.stderr.write(String(error !== null && error !== undefined && error.stack !== undefined ? error.stack : error) + "\n");
		process.exitCode = 2;
	});

	worker.on("exit", function(code) {
		if (process.exitCode === undefined) {
			process.exitCode = code;
		}
	});

	return true;
}

// go2jsRunMain runs main as a goroutine too, so that everything else in the
// program is written the same way and the scheduler has one place it all goes.
function go2jsRunMain(generator) {
	if (go2jsRelaunchWithStack()) {
		return;
	}

	const main = go2jsStart(generator);

	go2jsMainTask = main;

	go2jsRunLoop(main);
}

// go2jsIdleHook is set by a library that can wake the program from outside, as
// a signal can. When nothing is runnable it is asked whether it will do so; if
// it will, the loop is left where it is and picked up again from there.
var go2jsIdleHook = null;

function go2jsRunLoop(main) {
	while (go2jsLiveTasks > 0) {
		if (main.finished) {
			go2jsReady.length = 0;
			break;
		}

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

		if (go2jsIdleHook !== null && go2jsIdleHook(function() {
			try {
				go2jsRunLoop(main);
			} catch (e) {
				go2jsReportUncaughtPanic(e);
			}
		})) {
			return;
		}

		throw go2jsFatalError("all goroutines are asleep - deadlock!");
	}

	if (go2jsIdleHook !== null) {
		go2jsIdleHook(null);
	}
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

// go2jsRWMutexRLock takes a lock to read, which waits only for the writers. A
// writer waiting for the lock asks for it to be given to no reader until it is
// through, which is what Go does to keep a reader from starving it.
function* go2jsRWMutexRLock(mutex) {
	while (mutex.writer || go2jsHasBlocked(mutex, "mutexwriter")) {
		yield go2jsPark([mutex], "mutexreader", undefined);

		if (go2jsRunning.answered) {
			go2jsRunning.answered = false;

			return;
		}
	}

	mutex.readers++;
}

function go2jsRWMutexRUnlock(mutex) {
	if (mutex.readers === 0) {
		throw new Error("sync: RUnlock of unlocked RWMutex");
	}

	mutex.readers--;
	go2jsRWMutexPromote(mutex);
}

function* go2jsRWMutexLock(mutex) {
	while (mutex.writer || mutex.readers > 0) {
		yield go2jsPark([mutex], "mutexwriter", undefined);

		if (go2jsRunning.answered) {
			go2jsRunning.answered = false;

			return;
		}
	}

	mutex.writer = true;
}

function go2jsRWMutexUnlock(mutex) {
	if (!mutex.writer) {
		throw new Error("sync: unlock of unlocked RWMutex");
	}

	mutex.writer = false;
	go2jsRWMutexPromote(mutex);
}

// go2jsRWMutexPromote gives the lock to a writer waiting for it, or else to
// every reader waiting for it once no writer holds it.
function go2jsRWMutexPromote(mutex) {
	if (mutex.writer || mutex.readers > 0) {
		return;
	}

	const writer = go2jsTakeBlocked(mutex, "mutexwriter");

	if (writer !== null) {
		mutex.writer = true;
		go2jsAnswered(writer);

		return;
	}

	while (go2jsHasBlocked(mutex, "mutexreader")) {
		mutex.readers++;
		go2jsAnswered(go2jsTakeBlocked(mutex, "mutexreader"));
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

// Methods on the named scalar type time.Duration are emitted as flat functions.
function DurationString(d) {
	return go2jsDurationString(go2jsDurationNanos(d));
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

function go2jsOutputText(value, suffix, separator) {
	if (value === undefined) {
		return "";
	}

	// The operands are written by the same formatter Println and Print use, so
	// a byte slice, a named type or an error reads the same to whichever writer
	// it is sent.
	if (Array.isArray(value)) {
		const spaced = separator !== undefined && separator !== null && separator !== "";

		return go2jsJoinOperands(value, spaced) + suffix;
	}

	return go2jsJoinOperands([value], false) + suffix;
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

function go2jsBufioText(source) {
	source = go2jsUnwrap(source);

	if (source !== null && source !== undefined && typeof source.__go2js_text === "string") {
		return source.__go2js_text;
	}

	// A byte reader keeps the numbers it reads as bytes, which spell the text a
	// reader over it would read the same way they do for a string reader.
	if (source !== null && source !== undefined && source.__go2js_bytes) {
		return go2jsBytesToString(source.__go2js_bytes);
	}

	// A reader or a scanner reads from what it is given, which for a file is
	// the text left in it. What it read is kept on the source, so that reading
	// it twice reads it once.
	if (source !== null && source !== undefined && typeof source.__go2js_readAll === "function") {
		const text = go2jsBufioTextOf(source.__go2js_readAll());

		source.__go2js_text = text;

		return text;
	}

	// A buffer keeps the text it reads from as bytes, so what a reader reads is
	// what those bytes spell rather than the shape of the buffer holding them.
	if (source !== null && source !== undefined && typeof source.String === "function" &&
		source.data !== undefined) {
		const text = source.String();

		source.__go2js_text = text;

		return text;
	}

	return go2jsStringify(source);
}

// go2jsBufioTextOf is the text a read stands for, which is a string once the
// bytes it was read into have been spelled out.
function go2jsBufioTextOf(value) {
	if (typeof value === "string") {
		return value;
	}

	if (value !== null && value !== undefined && typeof value.String === "function") {
		return value.String();
	}

	return go2jsBytesToString(value);
}

// go2jsBufioNewScanner reads through everything the reader has and hands it out
// a piece at a time, the piece being whatever the split function it is given
// says one is. The split function is one of the package's own until a program
// asks for another, which is the way Go starts a scanner too.
function go2jsBufioNewScanner(source) {
	const bytes = go2jsStringToBytes(go2jsBufioText(source));
	let position = 0;
	let token = null;
	let err = null;
	let split = go2jsBufioScanLines;

	return {
		Split: function (fn) {
			if (typeof fn === "function") {
				split = fn;
			}

			return null;
		},
		Scan: function () {
			// Everything the reader had is already in hand, so every piece the
			// split function is asked about is the last of what is left, which
			// is what lets it hand back a token with no separator after it.
			const rest = bytes.slice(position);
			const result = split(rest, true);
			const advance = Number(result[0]) || 0;
			const piece = result[1];
			const failure = result[2];

			if (failure !== null && failure !== undefined) {
				err = failure;

				return false;
			}

			position += advance;

			if (piece !== null && piece !== undefined) {
				token = piece;

				return true;
			}

			return false;
		},
		Text: function () {
			return token === null ? "" : go2jsBufioBytesToString(token);
		},
		Bytes: function () {
			return token === null ? [] : token;
		},
		Buffer: function () {
			return token === null ? [] : token;
		},
		Err: function () {
			return err;
		},
	};
}

// go2jsBufioBytesToString is the text the bytes of a token stand for.
function go2jsBufioBytesToString(bytes) {
	return go2jsBytesToString(go2jsToArray(bytes));
}

// go2jsBufioDropCR takes a carriage return off the end of a line, which is what
// a line ending in the two characters of a Windows line ending leaves over.
function go2jsBufioDropCR(bytes) {
	if (bytes.length > 0 && bytes[bytes.length - 1] === 13) {
		return bytes.slice(0, -1);
	}

	return bytes;
}

// go2jsBufioScanLines is bufio.ScanLines: one token for each line, with a
// carriage return before the newline dropped.
function go2jsBufioScanLines(data, atEOF) {
	const bytes = go2jsToArray(data);

	if (atEOF && bytes.length === 0) {
		return [0, null, null];
	}

	for (let index = 0; index < bytes.length; index++) {
		if (bytes[index] === 10) {
			return [index + 1, go2jsBufioDropCR(bytes.slice(0, index)), null];
		}
	}

	if (atEOF) {
		return [bytes.length, go2jsBufioDropCR(bytes), null];
	}

	return [0, null, null];
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

function go2jsStringsFieldsFunc(value, predicate) {
	const parts = go2jsStringRuneParts(value);
	const out = [];
	let current = "";

	for (const char of parts) {
		if (go2jsCallNow(predicate, null, [char.rune])) {
			if (current !== "") {
				out.push(current);
				current = "";
			}

			continue;
		}

		current += char.text;
	}

	if (current !== "") {
		out.push(current);
	}

	return out;
}

function go2jsStringsTitle(value) {
	let out = "";
	let prev = 0x20;

	for (const char of go2jsStringRuneParts(value)) {
		const r = char.rune;

		if (go2jsStringsTitleSeparator(prev)) {
			const t = go2jsCaseTitle[r];
			out += String.fromCodePoint(t === undefined ? r : t);
		} else if (r === 0xfffd) {
			// A byte that is not the UTF-8 of a rune reads as RuneError, and
			// what is written back is that rune rather than the raw byte.
			out += "\ufffd";
		} else {
			out += char.text;
		}

		prev = r;
	}

	return out;
}

function go2jsStringsTitleSeparator(r) {
	if (r <= 0x7f) {
		if ((r >= 0x30 && r <= 0x39) || (r >= 0x61 && r <= 0x7a) || (r >= 0x41 && r <= 0x5a) || r === 0x5f) {
			return false;
		}

		return true;
	}

	if (go2jsUnicodeIsLetter(r) || go2jsUnicodeIsDigit(r)) {
		return false;
	}

	return go2jsUnicodeIsSpace(r);
}

function go2jsSortOrderCompare(left, right, less) {
	const result = go2jsCallNow(less, null, [left, right]);

	if (typeof result === "boolean") {
		return result ? -1 : 1;
	}

	if (result < 0) {
		return -1;
	}

	if (result > 0) {
		return 1;
	}

	return left - right;
}

function go2jsSortIndexStable(a, less) {
	const order = a.map((value, index) => index);

	order.sort((left, right) => go2jsSortOrderCompare(left, right, less));

	const sorted = order.map((index) => a[index]);

	for (let index = 0; index < a.length; index++) {
		a[index] = sorted[index];
	}
}

function go2jsSortSlice(a, less) {
	go2jsSortIndexStable(a, less);
}

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

function go2jsSentinelError(message, typeName, key) {
	return function go2jsSentinelErrorValue() {
		const name = typeName === undefined || typeName === null || typeName === "" ? "*errors.errorString" : typeName;
		const cacheKey = (key === undefined || key === null || key === "" ? name : key) + "\u0000" + message;
		let error = go2jsSentinelErrors.get(cacheKey);

		if (error === undefined) {
			error = go2jsNameError(new Error(message), name);

			const named = go2jsErrnoSentinelMessages.get(message);

			if (named !== undefined) {
				error.Is = function (other) {
					return go2jsErrorMessage(other) === named;
				};
			}

			go2jsSentinelErrors.set(cacheKey, error);
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

function go2jsTimeParse(layout, value) {
	const text = String(value);
	const parts = { year: 0, month: 1, day: 1, hour: 0, minute: 0, second: 0, milli: 0, nsec: 0, offset: null, pm: null };

	const refuse = function() {
		return [go2jsTimeZero(), go2jsTimeParseChunkError(String(layout), text, currentElement, currentHold)];
	};

	// A field read with a value it cannot hold is refused by naming the field,
	// which is the way Go turns such a reading down, and the moment read is the
	// zero moment rather than nothing at all.
	const rangeError = function(field) {
		return [go2jsTimeZero(), go2jsSentinelError("parsing time " + go2jsStrconvQuote(text) + ": " + field + " out of range")()];
	};

	// The piece of the layout a reading is being taken from and the text it is
	// being read out of are kept so that a refusal can tell which of them refused
	// the other, which is the way Go names the two pieces it could not match.
	let currentElement = "";
	let currentHold = "";

	const digits = function(min, max) {
		const found = /^[0-9]+/.exec(text.slice(position));

		if (found === null) {
			return null;
		}

		if (found[0].length < min) {
			return null;
		}

		// A piece without a leading zero takes as many of the digits before it
		// as it holds, up to two, while one that asks for a width takes exactly
		// that many, so a month or a day is not read as only its first digit.
		const read = found[0].length < max ? found[0].length : max;

		position += read;

		return parseInt(found[0].slice(0, read), 10);
	};

	// A fraction of a second that is written where the layout does not ask for
	// one is still the fraction of the second that was just read, the way Go
	// reads it, so a text carries what it was written with rather than what the
	// layout happened to ask for.
	const readFraction = function() {
		if (text[position] !== ".") {
			return false;
		}

		position += 1;

		const found = /^[0-9]+/.exec(text.slice(position));

		if (found === null) {
			return false;
		}

		const read = found[0].length > 9 ? found[0].slice(0, 9) : found[0];

		position += read.length;

		parts.nsec = parseInt((read + "000000000").slice(0, 9), 10);
		parts.milli = Math.floor(parts.nsec / 1000000);

		return true;
	};

	const named = function(names) {
		for (let i = 0; i < names.length; i++) {
			const full = names[i];

			for (const short of [full, full.slice(0, 3)]) {
				if (text.slice(position, position + short.length) === short) {
					position += short.length;

					return i;
				}
			}
		}

		return -1;
	};

	let position = 0;
	let layoutAt = 0;
	const layoutText = String(layout);

	while (layoutAt < layoutText.length) {
		const rest = layoutText.slice(layoutAt);

		currentHold = text.slice(position);

		// The pieces are read longest first, the same order a layout is written
		// out in, so that a name is not read as a shorter name it begins with.
		const pieces = ["January", "Jan", "Monday", "Mon", "2006", "_2", "01", "15", "03", "04", "05", "02",
			"PM", "pm", "MST", "Z07:00", "Z0700", "-07:00", "-0700", "06", "1", "2", "3", "4", "5"];

		let piece = null;

		for (const candidate of pieces) {
			if (rest.startsWith(candidate)) {
				piece = candidate;
				break;
			}
		}

		if (piece === null) {
			const fraction = /^\.([0-9]+)/.exec(rest);

			if (fraction !== null) {
				const width = fraction[1].length;
				const optional = fraction[1][0] === "9";

				currentElement = "." + fraction[1];

				// A fraction is only in the text when the text writes the dot
				// that begins it, so a dot the text does not have leaves a piece
				// of nines with no digits to read.
				const found = text[position] === "." ? /^[0-9]+/.exec(text.slice(position + 1)) : null;

				// A piece of nines asks for as many digits of a second as happen
				// to be written, so a text that writes none of them is read as a
				// whole second rather than refused.
				if (found === null) {
					if (optional) {
						layoutAt += 1 + fraction[1].length;
						continue;
					}

					return refuse();
				}

				position += 1;

				// A piece of nines takes as many of the digits of a second as were
				// written down and no more, while a piece of zeros asks for the
				// number of digits it carries, so a text with fewer or more of them
				// than it was asked for is refused rather than read as another second.
				if (found[0].length > 9 || (!optional && found[0].length !== width)) {
					return refuse();
				}

				const read = found[0].slice(0, optional ? found[0].length : width);

				position += read.length;

				layoutAt += 1 + fraction[1].length;

				parts.nsec = parseInt((read + "000000000").slice(0, 9), 10);
				parts.milli = Math.floor(parts.nsec / 1000000);

				continue;
			}

			// Anything else in a layout is written out as it is, so it has to be
			// there for the text to be read as the layout says it is written. A
			// run of such characters is taken together, the way Go reads the run
			// of text before the next piece of a layout, and a space in it is the
			// one character Go lets the text leave out.
			const ch = rest[0];

			if (ch === "Z") {
				currentElement = "Z";

				if (text[position] === "Z") {
					position += 1;
					layoutAt += 1;
					continue;
				}

				const zone = /^[+-][0-9]{2}:?[0-9]{2}/.exec(text.slice(position));

				if (zone === null) {
					return refuse();
				}

				position += zone[0].length;
				layoutAt += 1;

				continue;
			}

			let literalEnd = layoutAt;

			while (literalEnd < layoutText.length) {
				const tail = layoutText.slice(literalEnd);

				if (tail[0] === "." && /^\.([0-9]+)/.test(tail)) {
					break;
				}

				if (pieces.some(candidate => tail.startsWith(candidate))) {
					break;
				}

				literalEnd++;
			}

			const literal = layoutText.slice(layoutAt, literalEnd);

			currentElement = literal;

			// A run of spaces in the layout stands for a run of spaces in the
			// text, and the text may leave them out only when it has nothing
			// left at all, which is the way Go matches the spaces before a
			// piece rather than letting a non-space stand for them.
			let step = 0;

			while (step < literal.length) {
				if (literal[step] === " ") {
					if (position < text.length && text[position] !== " ") {
						currentHold = text.slice(position);

						return refuse();
					}

					while (step < literal.length && literal[step] === " ") {
						step += 1;
					}

					while (text[position] === " ") {
						position += 1;
					}

					continue;
				}

				if (position >= text.length || text[position] !== literal[step]) {
					currentHold = text.slice(position);

					return refuse();
				}

				position += 1;
				step += 1;
			}

			layoutAt = literalEnd;

			continue;
		}

		// A piece that reads what it stands for takes it from the text as it
		// goes and so has already moved past itself, while one that is written
		// out as it is has its own length to move past.
		currentElement = piece;

		switch (piece) {
		case "January":
		case "Jan": {
			const month = named(go2jsTimeMonths);

			if (month < 0) {
				return refuse();
			}

			parts.month = month + 1;
			break;
		}
		case "Monday":
		case "Mon":
			if (named(go2jsTimeDays) < 0) {
				return refuse();
			}

			break;
		case "2006": {
			const year = digits(4, 4);

			if (year === null) {
				return refuse();
			}

			parts.year = year;
			break;
		}
		case "06": {
			const year = digits(2, 2);

			if (year === null) {
				return refuse();
			}

			parts.year = year < 69 ? 2000 + year : 1900 + year;
			break;
		}
		case "01": {
			const month = digits(2, 2);

			if (month === null) {
				return refuse();
			}

			if (month < 1 || month > 12) {
				return rangeError("month");
			}

			parts.month = month;
			break;
		}
		case "02": {
			const day = digits(2, 2);

			if (day === null) {
				return refuse();
			}

			parts.day = day;
			break;
		}
		case "_2": {
			if (text[position] !== " ") {
				return refuse();
			}

			position += 1;

			const day = digits(1, 2);

			if (day === null) {
				return refuse();
			}

			parts.day = day;
			break;
		}
		case "15": {
			const hour = digits(2, 2);

			if (hour === null) {
				return refuse();
			}

			if (hour < 0 || hour >= 24) {
				return rangeError("hour");
			}

			parts.hour = hour;
			break;
		}
		case "03": {
			const hour = digits(2, 2);

			if (hour === null) {
				return refuse();
			}

			if (hour < 0 || hour > 12) {
				return rangeError("hour");
			}

			parts.hour = hour;
			break;
		}
		case "04": {
			const minute = digits(2, 2);

			if (minute === null) {
				return refuse();
			}

			if (minute < 0 || minute >= 60) {
				return rangeError("minute");
			}

			parts.minute = minute;
			break;
		}
		case "05": {
			const second = digits(2, 2);

			if (second === null) {
				return refuse();
			}

			if (second < 0 || second >= 60) {
				return rangeError("second");
			}

			parts.second = second;
			break;
		}
		case "1": {
			const month = digits(1, 2);

			if (month === null) {
				return refuse();
			}

			if (month < 1 || month > 12) {
				return rangeError("month");
			}

			parts.month = month;
			break;
		}
		case "2": {
			const day = digits(1, 2);

			if (day === null) {
				return refuse();
			}

			parts.day = day;
			break;
		}
		case "3": {
			const hour = digits(1, 2);

			if (hour === null) {
				return refuse();
			}

			if (hour < 0 || hour > 12) {
				return rangeError("hour");
			}

			parts.hour = hour;
			break;
		}
		case "4": {
			const minute = digits(1, 2);

			if (minute === null) {
				return refuse();
			}

			if (minute < 0 || minute >= 60) {
				return rangeError("minute");
			}

			parts.minute = minute;
			break;
		}
		case "5": {
			const second = digits(1, 2);

			if (second === null) {
				return refuse();
			}

			if (second < 0 || second >= 60) {
				return rangeError("second");
			}

			parts.second = second;
			break;
		}
		case "MST": {
			const zone = /^[A-Za-z]{2,5}/.exec(text.slice(position));

			if (zone === null) {
				return refuse();
			}

			position += zone[0].length;
			break;
		}
		case "PM":
		case "pm": {
			const marker = piece === "PM" ? /^(AM|PM)/ : /^(am|pm)/;

			if (marker.exec(text.slice(position)) === null) {
				return refuse();
			}

			position += 2;

			parts.pm = text[position - 2] === "P";

			if (parts.pm && parts.hour > 12) {
				return refuse();
			}

			break;
		}
		case "Z07:00":
		case "Z0700":
		case "-07:00":
		case "-0700": {
			const wanted = piece[0] === "Z" ? "Z" : /^[+-]/;
			let zone = wanted === "Z"
				? (/^Z/.test(text.slice(position)) ? "Z" : (/^[+-][0-9]{2}:?[0-9]{2}/.exec(text.slice(position)) || [null])[0])
				: (/^[+-][0-9]{2}:?[0-9]{2}/.exec(text.slice(position)) || [null])[0];

			// A fraction of a second may stand between the seconds and the zone,
			// which is where a layout that never asked for one lets it stand. It
			// is read as the seconds just gone read it, so what is left is what a
			// zone is read against, which is the value Go reports as unmatched.
			if (zone === null && text[position] === ".") {
				if (!readFraction()) {
					return refuse();
				}

				currentHold = text.slice(position);

				zone = wanted === "Z"
					? (/^Z/.test(text.slice(position)) ? "Z" : (/^[+-][0-9]{2}:?[0-9]{2}/.exec(text.slice(position)) || [null])[0])
					: (/^[+-][0-9]{2}:?[0-9]{2}/.exec(text.slice(position)) || [null])[0];
			}

			if (zone === null) {
				return refuse();
			}

			// "Z" stands for just itself in the text while the whole piece stands
			// in the layout, so the layout pointer moves the whole piece at the
			// bottom of the loop and the text pointer moves the letter it was.
			position += String(zone).length;

			// An offset is a time of day east or west of UTC, so an hour past the
			// day or a minute past the hour is not an offset and is turned down by
			// the part of it that went past the range it can hold.
			if (zone !== "Z") {
				const offset = /^([+-])([0-9]{2}):?([0-9]{2})$/.exec(zone);

				if (offset !== null) {
					if (parseInt(offset[2], 10) > 23) {
						return rangeError("time zone offset hour");
					}

					if (parseInt(offset[3], 10) > 59) {
						return rangeError("time zone offset minute");
					}
				}
			}

			parts.offset = zone === "Z" ? null : go2jsParseZoneSeconds(zone);
			break;
		}
		default:
			break;
		}

		layoutAt += piece.length;
	}

	if (position !== text.length) {
		// A text may still hold a fraction of a second the layout never asked
		// for, which Go reads it for, so it is taken here rather than refused.
		if (!(text[position] === "." && readFraction() && position === text.length)) {
			return [go2jsTimeZero(), go2jsTimeParseExtraError(layoutText, text, text.slice(position))];
		}
	}

	if (parts.pm !== null) {
		if (parts.pm && parts.hour < 12) {
			parts.hour += 12;
		}

		if (!parts.pm && parts.hour === 12) {
			parts.hour = 0;
		}
	}

	// The day is the one field whose range depends on another, since a month
	// carries a different number of days, and a leap year carries one more in
	// February, so it is turned down once the whole reading is known.
	const daysIn = function(month, year) {
		if (month === 2) {
			const leap = (year % 4 === 0 && year % 100 !== 0) || year % 400 === 0;

			return leap ? 29 : 28;
		}

		if (month === 4 || month === 6 || month === 9 || month === 11) {
			return 30;
		}

		return 31;
	};

	if (parts.day < 1 || parts.day > daysIn(parts.month, parts.year)) {
		return rangeError("day");
	}

	let date = new Date(0);

	date.setUTCFullYear(parts.year, parts.month - 1, parts.day);
	date.setUTCHours(parts.hour, parts.minute, parts.second, parts.milli);

	// A zone named in the text shifts the instant so that the clock the moment
	// carries reads in that zone as the clock the text carried it in.
	if (parts.offset !== null && parts.offset !== 0) {
		date = new Date(date.getTime() - parts.offset * 1000);
	}

	const parsed = go2jsTimeValue(date);

	parsed.__go2js_nsec = parts.nsec % 1000000;
	parsed.__go2js_location = "UTC";
	parsed.__go2js_zone = parts.offset;

	return [parsed, null];
}

// go2jsParseZoneSeconds is the number of seconds a written zone is east of UTC,
// which the zone captured from a text keeps, so that the moment is read in the
// zone the text named it in rather than only telling it apart from its clock.
function go2jsParseZoneSeconds(zone) {
	if (zone === "Z") {
		return 0;
	}

	const match = /^([+-])([0-9]{2}):?([0-9]{2})$/.exec(zone);

	if (match === null) {
		return 0;
	}

	const sign = match[1] === "-" ? -1 : 1;

	return sign * (parseInt(match[2], 10) * 3600 + parseInt(match[3], 10) * 60);
}

// go2jsTimeParseChunkError is the refusal of a piece of text that is not written
// the way the piece of the layout it was read against says it is written, and it
// tells which piece of text and which piece of the layout refused each other.
function go2jsTimeParseChunkError(layout, value, layoutChunk, valueChunk) {
	return go2jsSentinelError("parsing time " + JSON.stringify(value) + " as " +
		JSON.stringify(layout) + ": cannot parse " + JSON.stringify(valueChunk) +
		" as " + JSON.stringify(layoutChunk))();
}

// go2jsTimeParseExtraError is the refusal of a text that carries more than the
// layout asked for, which Go names as the text left over rather than as a piece
// of the layout it could not match.
function go2jsTimeParseExtraError(layout, value, extra) {
	return go2jsSentinelError("parsing time " + JSON.stringify(value) +
		": extra text: " + JSON.stringify(extra))();
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

// go2jsStringSlice slices on Go byte offsets. Working on the decoded bytes
// keeps the result aligned with Go even when a bound lands inside a rune.
function go2jsStringSlice(value, low, high) {
	const bytes = go2jsStringToBytes(go2jsBytesToString(value));
	const start = low === undefined ? 0 : Math.trunc(low);
	const end = high === undefined ? bytes.length : Math.trunc(high);

	if (start < 0 || start > end) {
		throw new RangeError(go2jsRuntimeErrorPrefix + "slice bounds out of range " +
			(start < 0 ? "[" + start + ":]" : "[" + start + ":" + end + "]"));
	}

	if (end > bytes.length) {
		throw new RangeError(go2jsRuntimeErrorPrefix + "slice bounds out of range [:" + end + "] with length " + bytes.length);
	}

	return go2jsBytesToString(bytes.slice(start, end));
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

function go2jsStringToRunes(value) {
	return go2jsStringRuneParts(value).map(part => part.rune);
}

// go2jsStringRuneParts walks a Go string one rune at a time the way a range
// does, pairing each rune with the piece of the string it was written as. A
// byte that is not the UTF-8 of a rune reads as the replacement rune and keeps
// its one byte, which is what Go reads there and what keeps the byte when a
// text is built back up from the parts.
function go2jsStringRuneParts(value) {
	const text = go2jsRawText(value);
	const parts = [];
	let i = 0;

	while (i < text.length) {
		const unit = text.charCodeAt(i);

		if (unit >= 0xdc00 && unit <= 0xdfff && (i === 0 || text.charCodeAt(i - 1) < 0xd800 || text.charCodeAt(i - 1) > 0xdbff)) {
			parts.push({ rune: 0xfffd, text: text[i] });
			i += 1;
			continue;
		}

		let code = unit;
		let width = 1;

		if (unit >= 0xd800 && unit <= 0xdbff && i + 1 < text.length) {
			const low = text.charCodeAt(i + 1);

			if (low >= 0xdc00 && low <= 0xdfff) {
				code = (unit - 0xd800) * 0x400 + (low - 0xdc00) + 0x10000;
				width = 2;
			}
		}

		parts.push({ rune: code, text: text.slice(i, i + width) });
		i += width;
	}

	return parts;
}

function go2jsRunesToString(value) {
	if (typeof value === "string") {
		return value;
	}

	if (value === null || value === undefined) {
		return "";
	}

	return Array.from(value).map(item => go2jsRuneString(item)).join("");
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

function go2jsOSFileMethods() {
	if (go2jsOSFileMethods.registered) {
		return;
	}

	go2jsOSFileMethods.registered = true;

	const close = function() {
		return null;
	};

	const descriptor = function() {
		return typeof this.fd === "number" ? this.fd : 0;
	};

	const attach = function(stream, label) {
		if (stream === null || typeof stream !== "object") {
			return;
		}

		const named = function() {
			return label;
		};

		for (const name of ["Fd", "Name", "Close"]) {
			if (Object.prototype.hasOwnProperty.call(stream, name)) {
				continue;
			}

			const value = name === "Name" ? named : name === "Close" ? close : descriptor;

			Object.defineProperty(stream, name, {
				value: value,
				enumerable: false,
				writable: true,
				configurable: true
			});
		}
	};

	for (const name of ["*os.File", "*File"]) {
		go2jsRegisterMethod(name + ".Write", go2jsOSFileWrite);
		go2jsRegisterMethod(name + ".WriteString", go2jsOSFileWriteString);
		go2jsRegisterMethod(name + ".Name", go2jsOSFileName);
		go2jsRegisterMethod(name + ".Close", close);
		go2jsRegisterMethod(name + ".Fd", descriptor);
	}

	// The standard input is read from the descriptor it stands for, because a
	// read of it is a read of the stream behind it and the stream object of the
	// host has no read of its own.
	const readStandardInput = function(buffer) {
		let filled = 0;
		const start = go2jsOSStreamPosition(process.stdin);

		while (filled < buffer.length) {
			let chunk = 0;

			// A read of the descriptor of the host is handed the bytes as a view
			// over them, since that is what it reads into, and what it filled is
			// copied back into the place the program asked for.
			const view = Buffer.from(buffer.slice(filled, buffer.length));

			try {
				try {
					chunk = require("fs").readSync(0, view, 0, view.length, start + filled);
				} catch (positioned) {
					// a pipe has no place to read from, so it is read from where it is
					if (positioned.code !== "ESPIPE") {
						throw positioned;
					}

					chunk = require("fs").readSync(0, view, 0, view.length, null);
				}
			} catch (err) {
				// Nothing left to hand over is the end of the text rather than a
				// failure of the read, which is what a read reports it as.
				if (err.code === "EOF" || err.code === "EAGAIN") {
					break;
				}

				return [filled, go2jsOSHostError(err, "read", "/dev/stdin")];
			}

			if (chunk <= 0) {
				break;
			}

			for (let i = 0; i < chunk; i++) {
				buffer[filled + i] = view[i];
			}

			filled += chunk;
		}

		go2jsOSStreamSetPosition(process.stdin, start + filled);

		return [filled, filled > 0 ? null : go2jsIOEOF()];
	};

	// What a scanner or a reader reads the standard input through is the whole
	// of it, which is read from the descriptor it stands for.
	Object.defineProperty(process.stdin, "__go2js_readAll", {
		value: function () {
			return go2jsScanStdinRest();
		},
		enumerable: false,
		writable: true,
		configurable: true
	});

	attach(process.stdout, "/dev/stdout");
	attach(process.stderr, "/dev/stderr");
	attach(process.stdin, "/dev/stdin");

	Object.defineProperty(process.stdin, "Read", {
		value: readStandardInput,
		enumerable: false,
		writable: true,
		configurable: true
	});
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
				const listed = lookup(target);

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

function go2jsReflectType(go2jsKind, go2jsName, go2jsElem, go2jsKey, go2jsMethods, go2jsLen, go2jsOuts, go2jsVariadic, go2jsIns) {
	const type = {
		__go2js_reflectType: true,
		kind: go2jsKind,
		name: go2jsName || "",
		elem: go2jsElem || null,
		key: go2jsKey || null,
		methods: go2jsMethods || [],
		len: typeof go2jsLen === "number" ? go2jsLen : null,
		// A function type carries what it takes and what it hands back as the
		// types of each, and whether it takes a trailing run of arguments,
		// because reflect has to answer for every one of them. The types are
		// descriptors of their own, so a program that asks about one gets a type
		// it can go on asking about.
		ins: Array.isArray(go2jsIns) ? go2jsIns : null,
		outs: Array.isArray(go2jsOuts) ? go2jsOuts : [],
		variadic: go2jsVariadic === true
	};

	// A descriptor is usually written out at the top of the generated program,
	// which is above the registrations that say what a named type is made of, so
	// what a descriptor needs to know about a named type is looked up when it is
	// wanted rather than when it is written. A lookup that comes back with
	// nothing is tried again rather than remembered, because the registration it
	// was looking for has yet to run.
	if (go2jsKind === "struct") {
		go2jsReflectLazyList(type, "fields", function(go2jsType) {
			return go2jsReflectFieldsOfName(go2jsType.name);
		});
	}

	if (!Array.isArray(go2jsMethods) || go2jsMethods.length === 0) {
		go2jsReflectLazyList(type, "methods", go2jsReflectMethodsForType);
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

	// A program may hold a type behind an interface of its own that names one of
	// the methods of reflect.Type, so the descriptor answers a call made on it.
	Object.setPrototypeOf(type, go2jsReflectTypePrototype());

	return type;
}

let go2jsReflectTypeProtoCache = null;

function go2jsReflectTypePrototype() {
	if (go2jsReflectTypeProtoCache === null) {
		go2jsReflectTypeProtoCache = {};

		for (const name of Object.keys(go2jsReflectTypeMethods)) {
			const fn = go2jsReflectTypeMethods[name];

			Object.defineProperty(go2jsReflectTypeProtoCache, name, {
				value: function(...args) {
					return go2jsCallNow(fn, null, [this, ...args]);
				},
				enumerable: false,
				writable: true,
				configurable: true
			});
		}
	}

	return go2jsReflectTypeProtoCache;
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
	// A pointer reaches every method of the type it points at, including the
	// ones declared with a pointer receiver, so the whole set is taken rather
	// than the value set the element may carry on its own.
	const methods = type !== null && type !== undefined && type.name !== ""
		? go2jsReflectMethodsOfName(type.name)
		: go2jsReflectMethodsOf(type);

	return go2jsReflectType("ptr", "", type, null, methods.length === 0 ? go2jsReflectMethodsOf(type) : methods);
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
	const descriptor = go2jsReflectType(go2jsReflectTypeKind(bare), name, null, null, []);

	// A descriptor read off a value has to say which fields the value has, or
	// NumField and Field have nothing to answer with, because nothing else about
	// the descriptor names them. The emitter's list is the one to believe, and
	// the value settles it when the type is not one the emitter described.
	if (descriptor.kind === "struct" && !Array.isArray(descriptor.fields)) {
		descriptor.fields = go2jsReflectFieldDescriptorsOf(bare);
	}

	return descriptor;
}

// go2jsReflectTypeFromName builds a type descriptor from the name of a type. A
// value that reached reflect through an interface carries the name of what it
// holds, and a name written out in full says what the type is where reading it
// back off the value does not, because a pointer cell, an empty map, and an
// untyped nil all say nothing at all about what they hold. The empty interface
// says nothing either way here, because what a value passed as one turned out
// to be is said by the value and not by the interface it passed through.
function go2jsReflectTypeFromName(name) {
	if (typeof name !== "string") {
		return null;
	}

	if (go2jsIsInterfaceTypeName(name.trim())) {
		return null;
	}

	return go2jsReflectDeclaredType(name);
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
		return go2jsReflectType(go2jsReflectTypeKindOfName(text), text.slice(text.lastIndexOf(".") + 1), null, null, []);
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

// go2jsReflectValueMethodsOfName looks up the methods a value of a named type
// carries, which is the ones declared with a value receiver. A type that never
// registered a value set is one the emitter did not describe, so the whole set
// is the best that can be said of it.
function go2jsReflectValueMethodsOfName(name) {
	if (typeof name !== "string" || name === "") {
		return [];
	}

	const short = name.slice(name.lastIndexOf(".") + 1);

	for (const registered of Object.keys(go2jsValueMethodSets)) {
		if (registered === name || registered.slice(registered.lastIndexOf(".") + 1) === short) {
			return go2jsValueMethodSets[registered];
		}
	}

	return go2jsReflectMethodsOfName(name);
}

// go2jsReflectMethodsForType reports the method set a type descriptor carries,
// which is the whole set for a pointer and the value set for everything else,
// because a pointer reaches every method of the type it points at and a value
// reaches only the ones declared with a value receiver.
function go2jsReflectMethodsForType(type) {
	if (type === null || type === undefined) {
		return [];
	}

	if (type.kind === "ptr") {
		return go2jsReflectMethodsOfName(type.name);
	}

	return go2jsReflectValueMethodsOfName(type.name);
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

// go2jsReflectKindName is reflect.Kind's own String, which is the kind's name.
// A kind is kept as that name, so the method has only to hand it back.
function go2jsReflectKindName(kind) {
	return String(kind);
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

// go2jsReflectValueFieldByIndex follows a path of field indexes down through
// nested structs, which is how a field promoted from an embedded struct is
// reached. Each step keeps the getter and setter of the one before it, so a
// write to the field at the end reaches the storage the caller owns.
function go2jsReflectValueFieldByIndex(receiver, index) {
	if (!Array.isArray(index)) {
		throw new TypeError("reflect: FieldByIndex of non-slice index");
	}

	let current = receiver;

	for (const at of index) {
		current = go2jsReflectValueField(current, at);
	}

	return current;
}

// go2jsReflectValueFieldByName looks a struct field up by name on a value,
// reaching the fields of an embedded struct as Go does, which is the field as a
// value, and the zero Value when the struct carried none of that name.
function go2jsReflectValueFieldByName(receiver, name) {
	const found = go2jsReflectTypeFieldByName(go2jsReflectValueType(receiver), name);

	if (found[1] !== true) {
		return go2jsReflectValue(undefined, go2jsReflectType("invalid", "", null, null, []));
	}

	return go2jsReflectValueFieldByIndex(receiver, found[0].Index);
}

function go2jsReflectValueIndex(receiver, index) {
	const target = go2jsReflectRead(receiver);
	const at = Math.trunc(Number(index));

	// A string is indexed by bytes rather than by the code units the runtime
	// stores it as, and an element of one is the unsigned byte at that place.
	if (typeof target === "string") {
		const bytes = go2jsStringToBytes(target);

		if (at < 0 || at >= bytes.length) {
			go2jsPanic("reflect: string index out of range");
		}

		return go2jsReflectValue(bytes[at], go2jsReflectType("uint8", "uint8", null, null, []));
	}

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

// go2jsReflectValueNumMethod reports how many methods a value can call, which
// is the whole set for a pointer and the value set for everything else, because
// a value cannot call a method declared with a pointer receiver.
function go2jsReflectValueNumMethod(receiver) {
	return go2jsReflectMethodsOf(go2jsReflectValueType(receiver)).length;
}

// go2jsReflectValueMethod reports one method of a value by index, as a bound
// method value that can be called, which is what Go hands back rather than a
// bare name.
function go2jsReflectValueMethod(receiver, index) {
	const methods = go2jsReflectMethodsOf(go2jsReflectValueType(receiver));
	const at = Math.trunc(Number(index));

	if (at < 0 || at >= methods.length) {
		throw new RangeError("reflect: Method index out of range");
	}

	return go2jsReflectBoundMethod(receiver, methods[at]);
}

// go2jsReflectMethodKeys lists the keys a method of a value could be recorded
// under. A method declared with a pointer receiver rides under a star, and a
// pointer reaches value methods too, so both spellings are tried.
function go2jsReflectMethodKeys(type, name) {
	const owner = type !== null && type !== undefined && type.kind === "ptr" && type.elem !== null ? type.elem : type;
	const bare = owner !== null && owner !== undefined ? owner.name : "";

	if (typeof bare !== "string" || bare === "") {
		return [];
	}

	return ["*" + bare + "." + name, bare + "." + name];
}

// go2jsReflectBoundMethod binds a method to the value it was reached on, which
// is the value a program then calls. A method the table never recorded cannot
// be called, and comes back as a value that reads as invalid.
function go2jsReflectBoundMethod(receiver, name) {
	const value = go2jsReflectRead(receiver);
	const keys = go2jsReflectMethodKeys(go2jsReflectValueType(receiver), name);

	for (const key of keys) {
		const fn = go2jsMethodTable[key];

		if (typeof fn === "function") {
			const results = typeof go2jsMethodResults[key] === "number" ? go2jsMethodResults[key] : null;

			return {
				__go2js_reflectValue: true,
				v: {
					__go2js_callable: true,
					__go2js_callable_results: results,
					__go2js_callable_value: (...args) => go2jsCallNow(fn, null, [value, ...args]),
					__go2js_signature: go2jsMethodSignatures[key] !== undefined ? go2jsMethodSignatures[key][0] : null
				},
				// A method is a function, and its type carries how many results it
				// hands back so a call boxes the right number of them.
				t: go2jsReflectType("func", "", null, null, [], null, results === null ? 1 : results, false)
			};
		}
	}

	return go2jsReflectValue(undefined, go2jsReflectType("invalid", "", null, null, []));
}

// go2jsReflectBoxResults wraps what a call handed back into the slice of Values
// Go returns. A call with no results boxes nothing, a call with one result boxes
// it whole, and a call with more than one has to take them apart, because the
// runtime hands several results over as a run rather than as one value.
function go2jsReflectBoxResults(results, raw) {
	if (results === 0) {
		return [];
	}

	if (results === null || results === 1) {
		return [go2jsReflectValue(raw, go2jsReflectTypeOf(raw))];
	}

	const tuple = Array.isArray(raw) ? raw : [];

	return tuple.map((item) => go2jsReflectValue(item, go2jsReflectTypeOf(item)));
}

function go2jsReflectValueIsValid(receiver) {
	if (!go2jsReflectValueBox(receiver)) {
		return receiver !== null && receiver !== undefined;
	}

	return receiver.v !== null && receiver.v !== undefined;
}

function go2jsReflectValueCanSet(receiver) {
	// Only a value that stands over storage can be written to, which is the same
	// as being able to have its address taken.
	return go2jsReflectValueCanAddr(receiver);
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

// go2jsReflectValueMapRange walks a map's entries one at a time, which is what
// MapRange hands back: an iterator whose Next moves to the next entry and whose
// Key and Value read the entry it stands on.
function go2jsReflectValueMapRange(receiver) {
	const keys = go2jsReflectValueMapKeys(receiver);
	let at = 0;

	return {
		Next: function() {
			if (at >= keys.length) {
				return false;
			}

			at++;

			return true;
		},
		Key: function() {
			return keys[at - 1];
		},
		Value: function() {
			return go2jsReflectValueMapIndex(receiver, keys[at - 1]);
		},
		Reset: function() {
			at = 0;
		}
	};
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

// go2jsReflectValueBytes reads the bytes a slice of bytes stands for, which is
// how a program reaches the storage behind a byte slice.
function go2jsReflectValueBytes(receiver) {
	if (go2jsReflectKind(receiver) !== "slice") {
		go2jsPanic("reflect.Value.Bytes of non-slice type");
	}

	const value = go2jsReflectUnwrap(go2jsReflectRead(receiver));

	if (value instanceof Uint8Array) {
		return value;
	}

	if (Array.isArray(value)) {
		return Uint8Array.from(value);
	}

	return new Uint8Array(0);
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

// go2jsReflectFieldWithIndex is the field at a position, carrying the index
// path reflect reports for it, which for a field of a struct is the position
// itself and for one reached through an embedded struct is the path down to it.
function go2jsReflectFieldWithIndex(field, index) {
	const described = typeof field === "string" ? {Name: field} : field;
	const described2 = Object.assign({}, described, {Index: index});

	return go2jsReflectStructField(described2);
}

// go2jsReflectTypeFieldsOf lays out a struct type's fields in the order reflect
// reports them, each with the one-element index that reaches it.
function go2jsReflectTypeFieldsOf(type) {
	// A descriptor that names its own fields answers from that list, because a
	// zero value of the type would be built without them.
	const raw = Array.isArray(type.fields)
		? type.fields
		: go2jsReflectStructFieldNames(go2jsReflectUnwrap(go2jsReflectZeroOf(type)));

	const fields = [];

	for (let index = 0; index < raw.length; index++) {
		fields.push(go2jsReflectFieldWithIndex(raw[index], [index]));
	}

	return fields;
}

function go2jsReflectTypeField(receiver, index) {
	const fields = go2jsReflectTypeFieldsOf(receiver);

	if (index < 0 || index >= fields.length) {
		throw new RangeError("reflect: Field index out of range");
	}

	return fields[index];
}

// go2jsReflectTypeFieldByIndex walks a path of field indexes down through
// embedded structs, which is how a type names a field promoted from one.
function go2jsReflectTypeFieldByIndex(receiver, index) {
	const path = Array.from(index === null || index === undefined ? [] : index, (step) => Math.trunc(Number(step)));
	let type = receiver;
	let field = null;

	for (const step of path) {
		field = go2jsReflectTypeField(type, step);

		if (field === null || field === undefined) {
			return go2jsReflectStructField({Name: ""});
		}

		type = field.Type;
	}

	return field === null ? go2jsReflectStructField({Name: ""}) : Object.assign({}, field, {Index: path});
}

// go2jsReflectTypeFieldByName looks a struct field up by name, reaching the
// fields of an embedded struct as Go does and stopping at the shallowest depth
// that names one. A name carried more than once at that depth is ambiguous, and
// a name carried by none of them is not there, both of which answer the same
// way: the field is not one the type has.
function go2jsReflectTypeFieldByName(receiver, name) {
	if (go2jsReflectTypeKindOf(receiver) !== "struct") {
		throw new TypeError("reflect: FieldByName of non-struct type " + go2jsReflectTypeString(receiver));
	}

	let level = [{type: receiver, prefix: []}];

	while (level.length > 0) {
		const next = [];
		let found = null;
		let count = 0;

		for (const entry of level) {
			const fields = go2jsReflectTypeFieldsOf(entry.type);

			for (let index = 0; index < fields.length; index++) {
				const field = fields[index];
				const path = entry.prefix.concat([index]);

				if (field.Name === name) {
					count++;

					if (count === 1) {
						found = Object.assign({}, field, {Index: path});
					}
				}

				// The fields of an embedded struct are searched a level down, so
				// a name promoted from it reads as the embedded struct's own.
				if (field.Anonymous && go2jsReflectKind(field.Type) === "struct") {
					next.push({type: field.Type, prefix: path});
				}
			}
		}

		if (count === 1) {
			return [found, true];
		}

		if (count > 1) {
			return [go2jsReflectStructField({Name: ""}), false];
		}

		level = next;
	}

	return [go2jsReflectStructField({Name: ""}), false];
}

// go2jsReflectTypeMethodByName looks a method up by name, which is the method
// and whether the type carried one of that name.
function go2jsReflectTypeMethodByName(receiver, name) {
	const methods = go2jsReflectMethodsOf(go2jsReflectValueType(receiver));
	const index = methods.indexOf(name);

	if (index < 0) {
		return [{Name: "", Type: null, Index: 0}, false];
	}

	return [go2jsReflectTypeMethod(receiver, index), true];
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
// set in, which is alphabetical by name. A method is a function, so its type
// is reported as one; the signature is not spelled out, because the only thing
// a program asks of a method type is that it is a function.
function go2jsReflectTypeMethod(receiver, index) {
	const methods = go2jsReflectMethodsOf(go2jsReflectValueType(receiver));

	if (index < 0 || index >= methods.length) {
		throw new RangeError("reflect: Method index out of range");
	}

	const name = methods[index];
	const type = go2jsReflectValueType(receiver);
	const found = go2jsReflectFindMethod(type, name);
	let func = go2jsReflectValue(undefined, go2jsReflectType("invalid", "", null, null, []));

	if (found !== null) {
		const signature = go2jsMethodSignatures[found.key];

		func = {
			__go2js_reflectValue: true,
			v: {
				__go2js_callable: true,
				__go2js_callable_results: found.results,
				__go2js_callable_value: (...args) => go2jsCallNow(found.fn, null, args),
				__go2js_signature: signature !== undefined ? signature[1] : null
			},
			t: go2jsReflectType("func", "", null, null, [], null, found.results === null ? 1 : found.results, false)
		};
	}

	return {Name: name, Type: go2jsReflectType("func", "", null, null, []), Func: func, Index: index};
}

const go2jsMethodSignatures = Object.create(null);

// go2jsReflectFindMethod finds the function a method of a type was recorded
// under, with the key it was recorded by and how many results it hands back.
function go2jsReflectFindMethod(type, name) {
	for (const key of go2jsReflectMethodKeys(type, name)) {
		const fn = go2jsMethodTable[key];

		if (typeof fn === "function") {
			return {key: key, fn: fn, results: typeof go2jsMethodResults[key] === "number" ? go2jsMethodResults[key] : null};
		}
	}

	return null;
}

// go2jsReflectTypeIsVariadic reports whether a function type takes a trailing
// run of arguments, which the emitter recorded when it wrote the type.
function go2jsReflectTypeIsVariadic(receiver) {
	const type = go2jsReflectValueType(receiver);

	return type !== null && type !== undefined && type.variadic === true;
}

// go2jsReflectFuncParts reports the argument types and the result types a
// function type was written with, which is what In, Out, NumIn and NumOut all
// read, and the number of results an older descriptor kept on its own.
function go2jsReflectFuncParts(type) {
	if (type === null || type === undefined) {
		return {ins: [], outs: []};
	}

	return {
		ins: Array.isArray(type.ins) ? type.ins : [],
		outs: Array.isArray(type.outs) ? type.outs : []
	};
}

function go2jsReflectTypeNumIn(receiver) {
	return go2jsReflectFuncParts(go2jsReflectValueType(receiver)).ins.length;
}

function go2jsReflectTypeIn(receiver, index) {
	const ins = go2jsReflectFuncParts(go2jsReflectValueType(receiver)).ins;
	const at = Math.trunc(Number(index));

	if (at < 0 || at >= ins.length) {
		throw new RangeError("reflect: In index out of range");
	}

	return ins[at];
}

function go2jsReflectTypeNumOut(receiver) {
	return go2jsReflectFuncParts(go2jsReflectValueType(receiver)).outs.length;
}

function go2jsReflectTypeOut(receiver, index) {
	const outs = go2jsReflectFuncParts(go2jsReflectValueType(receiver)).outs;
	const at = Math.trunc(Number(index));

	if (at < 0 || at >= outs.length) {
		throw new RangeError("reflect: Out index out of range");
	}

	return outs[at];
}

// go2jsReflectTypeComparable reports whether values of a type can be compared,
// which is false for a slice, a map and a function and true for everything
// else, down through the elements of an array and the fields of a struct.
function go2jsReflectTypeComparable(receiver) {
	return go2jsReflectComparable(go2jsReflectValueType(receiver));
}

function go2jsReflectComparable(type) {
	const kind = go2jsReflectKind(type);

	if (kind === "slice" || kind === "map" || kind === "func") {
		return false;
	}

	if (kind === "array") {
		const elem = go2jsReflectElem(type);

		return elem !== null && elem !== undefined && go2jsReflectComparable(elem);
	}

	if (kind === "struct") {
		const fields = Array.isArray(type.fields) ? type.fields : [];

		for (const field of fields) {
			const fieldType = go2jsReflectFieldTypeOf(field);

			if (fieldType !== null && fieldType !== undefined && !go2jsReflectComparable(fieldType)) {
				return false;
			}
		}
	}

	return true;
}

// go2jsReflectNumericKind reports whether a kind is one of the numeric kinds,
// which are the ones two of are always convertible to one another.
function go2jsReflectNumericKind(kind) {
	return kind === "int" || kind === "int8" || kind === "int16" || kind === "int32" || kind === "int64" ||
		kind === "uint" || kind === "uint8" || kind === "uint16" || kind === "uint32" || kind === "uint64" ||
		kind === "uintptr" || kind === "float32" || kind === "float64" ||
		kind === "complex64" || kind === "complex128";
}

// go2jsReflectIntegerKind reports whether a kind is one of the integer kinds,
// which are the numbers Go turns into a string of the rune they name.
function go2jsReflectIntegerKind(kind) {
	return kind === "int" || kind === "int8" || kind === "int16" || kind === "int32" || kind === "int64" ||
		kind === "uint" || kind === "uint8" || kind === "uint16" || kind === "uint32" || kind === "uint64" ||
		kind === "uintptr";
}

// go2jsReflectByteOrRuneSlice reports whether a type is a slice of bytes or a
function go2jsReflectByteOrRuneSlice(type) {
	if (type === null || type === undefined || type.kind !== "slice") {
		return false;
	}

	const elem = type.elem;

	return elem !== null && elem !== undefined && (elem.kind === "uint8" || elem.kind === "int32");
}

// go2jsReflectSameUnderlying reports whether two types are made of the same
// thing, which is what makes a named type and its bare form hold one another.
function go2jsReflectSameUnderlying(left, right) {
	if (left === null || left === undefined || right === null || right === undefined || left.kind !== right.kind) {
		return false;
	}

	if (left.kind === "slice" || left.kind === "ptr") {
		return go2jsReflectSameUnderlying(left.elem, right.elem);
	}

	if (left.kind === "array") {
		return left.len === right.len && go2jsReflectSameUnderlying(left.elem, right.elem);
	}

	if (left.kind === "map") {
		return go2jsReflectSameUnderlying(left.key, right.key) && go2jsReflectSameUnderlying(left.elem, right.elem);
	}

	if (left.kind === "struct") {
		const leftFields = go2jsReflectTypeFieldsOf(left);
		const rightFields = go2jsReflectTypeFieldsOf(right);

		if (leftFields.length !== rightFields.length) {
			return false;
		}

		for (let index = 0; index < leftFields.length; index++) {
			if (leftFields[index].Name !== rightFields[index].Name) {
				return false;
			}

			if (!go2jsReflectSameUnderlying(leftFields[index].Type, rightFields[index].Type)) {
				return false;
			}
		}

		return true;
	}

	if (left.kind === "func") {
		return go2jsReflectTypeString(left) === go2jsReflectTypeString(right);
	}

	return true;
}

// go2jsReflectTypeAssignableTo reports whether a value of one type may be put
// where the other is wanted: the same type, a type that carries every method of
// an interface, or a type sharing an underlying form with one that is not named.
function go2jsReflectTypeAssignableTo(receiver, target) {
	const value = go2jsReflectValueType(receiver);
	const want = go2jsReflectValueType(target);

	if (go2jsReflectTypeString(value) === go2jsReflectTypeString(want)) {
		return true;
	}

	if (want.kind === "interface") {
		return go2jsReflectTypeImplements(value, want);
	}

	if ((value.name === "" || want.name === "") && go2jsReflectSameUnderlying(value, want)) {
		return true;
	}

	return false;
}

// go2jsReflectTypeConvertibleTo reports whether a type can be turned into
// another, which every assignable pair can and so can the pairs Go names:
// numeric to numeric, string to string, and string to and from a slice of
// bytes or runes.
function go2jsReflectTypeConvertibleTo(receiver, target) {
	const value = go2jsReflectValueType(receiver);
	const want = go2jsReflectValueType(target);

	if (go2jsReflectTypeAssignableTo(value, want)) {
		return true;
	}

	if (go2jsReflectNumericKind(value.kind) && go2jsReflectNumericKind(want.kind)) {
		return true;
	}

	if (value.kind === "string" && want.kind === "string") {
		return true;
	}

	if (value.kind === "string" && go2jsReflectByteOrRuneSlice(want)) {
		return true;
	}

	if (go2jsReflectByteOrRuneSlice(value) && want.kind === "string") {
		return true;
	}

	// An integer becomes the one-rune string that names its code point, which is
	// the one conversion between a number and a string that Go allows.
	if (go2jsReflectIntegerKind(value.kind) && want.kind === "string") {
		return true;
	}

	return go2jsReflectSameUnderlying(value, want);
}

// go2jsReflectValueConvert turns a value into another type, which the pairs Go
// calls convertible are turned between. A number keeps its value under the new
// kind, a string comes and goes as a slice of bytes or runes, and everything
// else moves across unchanged.
function go2jsReflectValueConvert(receiver, target) {
	const want = go2jsReflectValueType(target);
	const source = go2jsReflectValueType(receiver);
	const value = go2jsReflectRead(receiver);

	if (want.kind === "string") {
		if (source.kind === "slice") {
			const items = go2jsReflectUnwrap(value);

			if (source.elem !== null && source.elem !== undefined && source.elem.kind === "int32") {
				return go2jsReflectValue(String.fromCodePoint.apply(null, Array.from(items, (rune) => Number(rune))), want);
			}

			return go2jsReflectValue(go2jsBytesToString(go2jsReflectUnwrap(value)), want);
		}

		if (go2jsReflectNumericKind(source.kind)) {
			return go2jsReflectValue(String.fromCodePoint(Number(value)), want);
		}
	}

	if (want.kind === "slice" && source.kind === "string" && go2jsReflectByteOrRuneSlice(want)) {
		if (want.elem.kind === "int32") {
			return go2jsReflectValue(Array.from(String(value), (ch) => ch.codePointAt(0)), want);
		}

		return go2jsReflectValue(go2jsStringToBytes(String(value)), want);
	}

	return go2jsReflectValue(value, want);
}

// go2jsReflectValueCanConvert reports whether Convert would turn a value into
// the type asked for, which is whether the two types are convertible.
function go2jsReflectValueCanConvert(receiver, target) {
	if (!go2jsReflectValueIsValid(receiver)) {
		return false;
	}

	return go2jsReflectTypeConvertibleTo(go2jsReflectValueType(receiver), target);
}

// go2jsReflectValueMethodByName looks a method up by name, which reports the
// same shape as Method does so a caller cannot tell the two apart.
function go2jsReflectValueMethodByName(receiver, name) {
	const methods = go2jsReflectMethodsOf(go2jsReflectValueType(receiver));

	if (methods.indexOf(name) < 0) {
		return go2jsReflectValue(undefined, go2jsReflectType("invalid", "", null, null, []));
	}

	return go2jsReflectBoundMethod(receiver, name);
}

// go2jsReflectValueCall calls a function or bound method value, handing back
// the results as the slice of Values Go returns. The arguments arrive as the
// slice Call was given, so a nil slice is no arguments at all.
function go2jsReflectValueCall(receiver, args) {
	const value = go2jsReflectRead(receiver);
	const given = args === null || args === undefined ? [] : args;
	// Call is handed a slice of Values and passes what they hold, so each one is
	// read back out before the call is made.
	const list = Array.isArray(given) ? given.map(go2jsReflectRead) : [go2jsReflectRead(given)];

	if (value !== null && value !== undefined && value.__go2js_callable === true) {
		return go2jsReflectBoxResults(value.__go2js_callable_results, value.__go2js_callable_value(...list));
	}

	if (typeof value === "function") {
		const type = go2jsReflectValueType(receiver);
		const results = type !== null && type !== undefined && Array.isArray(type.outs) ? type.outs.length : null;

		return go2jsReflectBoxResults(results, go2jsCallNow(value, null, list));
	}

	throw new TypeError("reflect: call of reflect.Value.Call on a value that is not callable");
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
	// Only a value standing over storage has an address. A value read off
	// another value, such as one a call handed back, has no storage of its own,
	// so taking its address is refused rather than answered with a fresh cell.
	if (typeof receiver.get !== "function" || typeof receiver.set !== "function") {
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
	Cap: go2jsReflectValueCap,
	FieldByName: go2jsReflectValueFieldByName,
	FieldByIndex: go2jsReflectValueFieldByIndex,
	Kind: go2jsReflectKind,
	Type: go2jsReflectValueType,
	Set: go2jsReflectValueSet,
	SetZero: go2jsReflectValueSetZero,
	SetString: go2jsReflectValueSet,
	SetInt: go2jsReflectValueSetInt,
	SetUint: go2jsReflectValueSetUint,
	SetFloat: go2jsReflectValueSet,
	SetBool: go2jsReflectValueSet,
	NumField: go2jsReflectValueNumField,
	Field: go2jsReflectValueField,
	Index: go2jsReflectValueIndex,
	NumMethod: go2jsReflectValueNumMethod,
	Method: go2jsReflectValueMethod,
	IsValid: go2jsReflectValueIsValid,
	IsNil: go2jsReflectValueIsNil,
	IsZero: go2jsReflectValueIsZero,
	CanSet: go2jsReflectValueCanSet,
	CanInterface: go2jsReflectValueCanInterface,
	CanAddr: go2jsReflectValueCanAddr,
	CanCompare: go2jsReflectValueCanCompare,
	Comparable: go2jsReflectValueComparable,
	Convert: go2jsReflectValueConvert,
	CanConvert: go2jsReflectValueCanConvert,
	IsExported: go2jsReflectValueIsExported,
	Addr: go2jsReflectValueAddr,
	UnsafeAddr: go2jsReflectValueAddr,
	Pointer: go2jsReflectValuePointer,
	MethodByName: go2jsReflectValueMethodByName,
	Call: go2jsReflectValueCall,
	Bytes: go2jsReflectValueBytes,
	SetLen: go2jsReflectValueSetLen,
	MapIndex: go2jsReflectValueMapIndex,
	MapKeys: go2jsReflectValueMapKeys,
	MapRange: go2jsReflectValueMapRange,
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
	FieldByIndex: go2jsReflectTypeFieldByIndex,
	FieldByName: go2jsReflectTypeFieldByName,
	MethodByName: go2jsReflectTypeMethodByName,
	Elem: go2jsReflectTypeElemOf,
	String: go2jsReflectTypeString,
	Size: go2jsReflectTypeSize,
	Implements: go2jsReflectTypeImplements,
	Key: go2jsReflectTypeKey,
	Len: go2jsReflectTypeLen,
	NumMethod: go2jsReflectTypeNumMethod,
	Method: go2jsReflectTypeMethod,
	Comparable: go2jsReflectTypeComparable,
	AssignableTo: go2jsReflectTypeAssignableTo,
	ConvertibleTo: go2jsReflectTypeConvertibleTo,
	NumIn: go2jsReflectTypeNumIn,
	In: go2jsReflectTypeIn,
	NumOut: go2jsReflectTypeNumOut,
	Out: go2jsReflectTypeOut,
	IsVariadic: go2jsReflectTypeIsVariadic
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

	// A kind is a value of a named type of its own, so its String is reached the
	// way any other method is even though the value behind it is a plain name.
	go2jsMethodTable["reflect.Kind.String"] = go2jsReflectKindName;

	for (const [name, fn] of Object.entries(go2jsReflectTypeMethods)) {
		go2jsMethodTable["reflect.Type." + name] = fn;
	}
}

let go2jsReflectRegistered = false;

// go2jsReflectNew allocates a zero value of the given type and returns an
// addressable Value for it, which is what a Go pointer to a fresh T provides.
function go2jsReflectNew(type) {
	// The value reflect.New hands back is a pointer to the fresh storage, so it
	// reads as a pointer and Elem of it reads back as the type that was asked
	// for.
	return go2jsReflectValue(go2jsReflectNewBox(go2jsReflectUnwrap(go2jsReflectZeroOf(type))), go2jsReflectPtrTo(type));
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
		// The box is pointed at, so the value it holds is the type the pointer
		// points at rather than the pointer itself.
		return go2jsReflectAddressable(receiver.v, go2jsReflectElem(receiver.t));
	}

	if (receiver.t === null || receiver.t === undefined) {
		return go2jsReflectValue(receiver.v, go2jsReflectElem(go2jsReflectTypeOf(receiver.v)));
	}

	if (receiver.t.kind === "interface") {
		// An interface value keeps what it holds and the name of the type it
		// holds it as, so Elem gives that value back under the type it was
		// wrapped with rather than the type of the interface variable.
		const cell = receiver.v;
		let inner = cell;
		let named = null;

		if (cell !== null && cell !== undefined && cell.__go2js_interface === true) {
			inner = cell.value;
			named = typeof cell.type === "string" ? go2jsReflectTypeFromName(cell.type) : null;
		} else if (cell !== null && cell !== undefined && typeof cell === "object" && cell.v !== undefined) {
			inner = cell.v;
		}

		return go2jsReflectValue(inner, named !== null ? named : go2jsReflectTypeOf(inner));
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

// go2jsReflectValueSetZero writes the zero value of a value's own type into it,
// which is what SetZero does.
function go2jsReflectValueSetZero(receiver) {
	const type = go2jsReflectValueType(receiver);

	go2jsReflectValueSet(receiver, go2jsReflectValue(go2jsReflectUnwrap(go2jsReflectZeroOf(type)), type));
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
	const value = go2jsReflectUnwrap(go2jsReflectRead(receiver));

	if (value !== null && typeof value === "object" && value.__go2js_callable === true && typeof value.__go2js_callable_value === "function") {
		const call = value.__go2js_callable_value;
		const typed = function*(...args) {
			return call(...args);
		};

		if (typeof value.__go2js_signature === "string") {
			return go2jsInterface(typed, value.__go2js_signature, value.__go2js_signature);
		}

		return typed;
	}

	return value;
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

	// A string is counted in bytes, the way Go counts it, rather than in the
	// code units the runtime happens to store it as.
	if (typeof value === "string") {
		return go2jsStringByteLength(value);
	}

	if (value === null || value === undefined) {
		return 0;
	}

	if (value instanceof go2jsNativeMap) {
		return value.size;
	}

	return value.length !== undefined ? value.length : 0;
}

// go2jsReflectValueCap reports the capacity of a slice, an array or a channel.
// A slice here is the run of its elements, so its capacity is its length.
function go2jsReflectValueCap(receiver) {
	const kind = go2jsReflectKind(receiver);

	if (kind !== "slice" && kind !== "array" && kind !== "chan") {
		throw new TypeError("reflect: Cap of non-array, non-slice value");
	}

	// A slice made with a capacity past its length carries that capacity beside
	// it, because a plain list reports only the length it happens to have.
	const value = go2jsReflectUnwrap(go2jsReflectRead(receiver));

	if (value !== null && value !== undefined && typeof value === "object" && typeof value.__go2js_cap === "number") {
		return value.__go2js_cap;
	}

	return go2jsReflectValueLen(receiver);
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
	// A value can have its address taken only when it stands over storage, which
	// a box over a plain value does not, however valid the value it holds is.
	return go2jsReflectValueIsValid(receiver) &&
		typeof receiver.get === "function" && typeof receiver.set === "function";
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

// A value is comparable when its type is, which a struct with a field that is
// not comparable is not, however plainly comparable the rest of it looks.
function go2jsReflectValueComparable(receiver) {
	if (!go2jsReflectValueIsValid(receiver)) {
		return false;
	}

	return go2jsReflectComparable(go2jsReflectValueType(receiver));
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

	if (name.endsWith("$")) {
		name = name.slice(0, -1);
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

	if (payload === null || payload === undefined) {
		return "panic called with nil argument";
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
	// A fault the runtime raised by name is one of Go's too, whatever kind of
	// error the engine made it out of: an out of range slice, a failed
	// assertion, a channel closed twice.
	const named = thrown instanceof Error && (
		thrown.__go2js_error_name !== undefined ||
		String(thrown.message).startsWith("runtime error: ") ||
		String(thrown.message).startsWith("interface conversion: "));

	if (value === undefined && !named && go2jsRuntimeErrorText(thrown) === null) {
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

go2jsBigInt.prototype.MulRange = function(a, b) {
	// MulRange is the product of every whole number from a to b, one when the
	// range is empty and zero when it takes in zero.
	const low = BigInt(a);
	const high = BigInt(b);

	if (low > high) {
		this.value = 1n;
	} else if (low <= 0n && high >= 0n) {
		this.value = 0n;
	} else {
		let product = 1n;

		for (let n = low; n <= high; n++) {
			product *= n;
		}

		this.value = product;
	}

	return this;
};

go2jsBigInt.prototype.Binomial = function(n, k) {
	// Binomial is n over k, worked out as a product that stays whole at each step.
	let top = BigInt(n);
	let choose = BigInt(k);

	if (choose > top - choose) {
		choose = top - choose;
	}

	let result = 1n;

	for (let i = 1n; i <= choose; i++) {
		result = result * (top - choose + i) / i;
	}

	this.value = choose < 0n ? 0n : result;

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

	// The standard streams of the host carry no Write of their own; the file
	// the program sees is written through the one the os package registered.
	if (typeof target.Write !== "function" && (target === process.stdout || target === process.stderr)) {
		go2jsOSFileMethods();
		go2jsOSFileWrite(target, go2jsStringToBytes(text));

		return null;
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

// A file that was never opened is not a file, and everything asked of it is
// refused the same way Go refuses it: with the error that stands for an
// argument that is not one.
function go2jsOSFileInvalid() {
	const error = new Error("invalid argument");

	error.__go2js_errno = true;
	error.errno = 0;

	return error;
}

function go2jsOSFileOf(value) {
	if (value !== null && value !== undefined && value.__go2js_osfile === true) {
		return value;
	}

	return null;
}

function go2jsOSFileStream(value) {
	if (value === process.stdout) {
		return process.stdout;
	}

	if (value === process.stderr) {
		return process.stderr;
	}

	return null;
}

function go2jsOSFileWrite(file, buffer) {
	const value = go2jsUnwrap(file);
	const stream = go2jsOSFileStream(value);
	const bytes = go2jsToArray(buffer);

	if (stream !== null) {
		stream.write(go2jsBytesToString(buffer));
		return [bytes.length, null];
	}

	const target = go2jsOSFileOf(value);

	if (target === null) {
		// A file that is not there is refused, where the streams of the process
		// take what they are given when no file was named at all.
		if (file !== null && file !== undefined) {
			return [0, go2jsOSFileInvalid()];
		}

		go2jsWriteOut(process.stdout, go2jsBytesToString(buffer));

		return [bytes.length, null];
	}

	const result = go2jsOSFileWriteAt(target, buffer, target.position);

	// A write that went through moves the file along with it, and one that did
	// not leaves it where it was.
	if (result[1] === null) {
		target.position += bytes.length;
	}

	return result;
}

function go2jsOSFileWriteString(file, value) {
	return go2jsOSFileWrite(file, go2jsStringToBytes(value));
}

function go2jsOSFileName(file) {
	const value = go2jsUnwrap(file);

	if (value === process.stdout) {
		return "/dev/stdout";
	}

	if (value === process.stderr) {
		return "/dev/stderr";
	}

	if (value === process.stdin) {
		return "/dev/stdin";
	}

	const target = go2jsOSFileOf(value);

	return target === null ? "/dev/stdout" : target.path;
}

// go2jsOSStreamPosition is where reading one of the streams of the process
// stands, which is kept on the stream rather than left to the descriptor
// underneath, so that a seek of one can put it back.
function go2jsOSStreamPosition(stream) {
	return typeof stream.__go2js_position === "number" ? stream.__go2js_position : 0;
}

function go2jsOSStreamSetPosition(stream, position) {
	stream.__go2js_position = position;
}

// A write at a place leaves where the file stands alone, which is what tells it
// from a write that moves it along.
function go2jsOSFileWriteAt(file, buffer, offset) {
	const target = go2jsOSFileOf(go2jsUnwrap(file));

	if (target === null || target.closed === true) {
		return [0, go2jsOSFileInvalid()];
	}

	const bytes = go2jsToArray(buffer);
	const at = Number(offset);

	try {
		require("fs").writeSync(target.fd, Buffer.from(bytes), 0, bytes.length, target.append ? null : at);
	} catch (err) {
		return [0, go2jsOSHostError(err, "write", target.path)];
	}

	return [bytes.length, null];
}

// go2jsIsGoexit reports whether a thrown value is a goroutine being told to
// end rather than a panic, because the two are unwound the same way and only
// one of them is a failure.
function go2jsIsGoexit(thrown) {
	return thrown !== null && thrown !== undefined && thrown.__go2js_goexit === true;
}

const go2jsScanStdin = {text: "", pos: 0, ended: false, decoder: null};

// go2jsScanStdinRest is everything left of the standard input, which is what a
// reader or a scanner over it reads from. What a Scan function read ahead of
// what it used is the front of it.
function go2jsScanStdinRest() {
	const ahead = go2jsScanStdin.text.slice(go2jsScanStdin.pos);
	const parts = [];
	const fs = require("fs");

	go2jsScanStdin.text = "";
	go2jsScanStdin.pos = 0;

	while (!go2jsScanStdin.ended) {
		const chunk = Buffer.alloc(1 << 16);
		let count = 0;

		try {
			count = fs.readSync(0, chunk, 0, chunk.length, null);
		} catch (err) {
			if (err.code === "EAGAIN") {
				Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, 5);
				continue;
			}

			count = 0;
		}

		if (count <= 0) {
			go2jsScanStdin.ended = true;
			break;
		}

		parts.push(chunk.subarray(0, count));
	}

	const decoder = go2jsScanStdin.decoder === null ? new TextDecoder("utf-8") : go2jsScanStdin.decoder;

	return ahead + decoder.decode(Buffer.concat(parts));
}

// go2js:replaces
function go2jsMapKey(key) {
	if (key !== null && typeof key === "object" && key.__go2js_uncomparable === true) {
		throw new Error(go2jsRuntimeErrorPrefix + "hash of unhashable type " + (key.type || ""));
	}

	// A NaN is not equal to anything, itself included, so every one of them is
	// a key of its own that no lookup finds again.
	if (typeof key === "number" && key !== key) {
		go2jsNaNKeys = true;

		return new Number(NaN);
	}

	const signature = go2jsMapKeySignature(key);

	if (signature === null) {
		return key;
	}

	// A key with a NaN inside is no more equal to itself than a NaN is.
	if (signature.indexOf("numberNaN") !== -1) {
		return Object.assign(Object.create(Object.getPrototypeOf(key)), key);
	}

	let shape = go2jsMapKeyShapes.get(signature);

	if (shape === undefined) {
		go2jsMapKeyShapes.set(signature, key);

		return key;
	}

	return shape;
}

var go2jsNaNKeys = false;

// A map that holds a NaN is walked with the NaN handed back as the number it
// stands for, which the key it was stored under is not.
function go2jsRangeNaNKeys(map) {
	const unwrap = (key) => (key instanceof Number ? NaN : key);

	return {
		entries: () => Array.from(map.entries(), (entry) => [unwrap(entry[0]), entry[1]]),
		keys: () => Array.from(map.keys(), unwrap),
		values: () => Array.from(map.values()),
		[Symbol.iterator]: () => Array.from(map.keys(), unwrap)[Symbol.iterator](),
		get size() {
			return map.size;
		},
	};
}

function go2jsRuntimeErrorText(value) {
	const raised = value instanceof go2jsNativeError &&
		typeof value.message === "string" && value.message.startsWith(go2jsRuntimeErrorPrefix);

	if (!(value instanceof go2jsNativeTypeError) && !(value instanceof go2jsNativeRangeError) && !raised) {
		return null;
	}

	const message = String(value.message === undefined ? "" : value.message);

	if (/^Cannot (read|set) propert(?:y|ies) of (null|undefined)/.test(message)) {
		return go2jsRuntimeErrorPrefix + "invalid memory address or nil pointer dereference";
	}

	if (/^\w+ is not a function$/.test(message) || /of undefined$/.test(message)) {
		return go2jsRuntimeErrorPrefix + "invalid memory address or nil pointer dereference";
	}

	const fault = message.startsWith(go2jsRuntimeErrorPrefix)
		? message.slice(go2jsRuntimeErrorPrefix.length)
		: message;

	if (fault.startsWith("index out of range") || fault.startsWith("integer divide by zero") || raised) {
		return go2jsRuntimeErrorPrefix + fault;
	}

	return null;
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
		// A typed nil pointer is still a pointer of its type, and Go runs the
		// method of that type on the nil receiver.
		if (typeof value.type === "string" && value.type.startsWith("*")) {
			const nilMethod = go2jsMethodTable[value.type + "." + method] || go2jsMethodTable[value.type.slice(1) + "." + method];

			if (typeof nilMethod === "function") {
				return yield* go2jsCall(nilMethod, null, [null, ...args]);
			}
		}

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

	// A lock held in an interface (a sync.Locker, the L of a condition) is told
	// from a read-write one by what it keeps, and is locked the way the call that
	// names it directly would.
	if (typeof fn !== "function") {
		let lock = target;

		if (lock.__go2js_pointer === true) {
			lock = lock[go2jsPointerGet]();
		}

		if (lock !== null && typeof lock === "object" && Array.isArray(lock.blocked)) {
			const readWrite = "readers" in lock;

			if (method === "Lock") {
				return yield* (readWrite ? go2jsRWMutexLock(lock) : go2jsMutexLock(lock));
			}

			if (method === "Unlock") {
				return readWrite ? go2jsRWMutexUnlock(lock) : go2jsMutexUnlock(lock);
			}

			if (method === "RLock" && readWrite) {
				return yield* go2jsRWMutexRLock(lock);
			}

			if (method === "RUnlock" && readWrite) {
				return go2jsRWMutexRUnlock(lock);
			}
		}

		throw new TypeError("interface method " + method + " is not implemented");
	}

	if (registered) {
		return yield* go2jsCall(fn, null, [target, ...args]);
	}

	return yield* go2jsCall(fn, target, args);
}

function go2jsGenerator(value) {
	// A generator answers to next, throw and return together. A struct of the
	// program that has a method called next is not one, and running it as one
	// would wait on a task that never was.
	return value !== null && typeof value === "object" && typeof value.next === "function" && typeof value.throw === "function" && typeof value.return === "function";
}
if (typeof go2jsStringsBuilder === "function") { go2jsRegisterTypeName(go2jsStringsBuilder, "strings.Builder"); }


const go2jsState = {
    get go2jsTicks() {
        return go2jsTicks;
    },
    set go2jsTicks(value) {
        go2jsTicks = value;
    },
};

module.exports = {
    go2jsBufioNewScanner,
    go2jsFprint,
    go2jsIndex,
    go2jsIndexCheck,
    go2jsInterface,
    go2jsLen,
    go2jsMap,
    go2jsMapGet,
    go2jsMapTyped,
    go2jsMapUpdate,
    go2jsNilInterface,
    go2jsPreempt,
    go2jsPrintln,
    go2jsPtr,
    go2jsRangeCopies,
    go2jsRangeMap,
    go2jsRangeSequence,
    go2jsRegisterTypeName,
    go2jsReportUncaughtPanic,
    go2jsRunMain,
    go2jsRunesToString,
    go2jsSetPackageName,
    go2jsSetSourceLines,
    go2jsSliceAppend,
    go2jsSliceMake,
    go2jsSliceRange,
    go2jsSliceTyped,
    go2jsSortInts,
    go2jsSortSlice,
    go2jsSprintf,
    go2jsState,
    go2jsStrCompare,
    go2jsStringSlice,
    go2jsStringToRunes,
    go2jsStringsBuilder,
    go2jsStringsFields,
    go2jsStringsFieldsFunc,
    go2jsStringsNewReader,
    go2jsStringsRepeat,
    go2jsStringsSplit,
    go2jsStringsTitle,
    go2jsStringsToLower,
    go2jsStringsToUpper,
    go2jsStringsTrimSpace,
    go2jsTyped,
    go2jsUnicodeIsLetter,
    go2jsWideAdd,
    go2jsWideFloat,
    go2jsWideSub,
    go2jsWriteOut,
};
