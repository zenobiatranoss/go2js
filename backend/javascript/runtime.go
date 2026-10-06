package javascript

func runtimeSource() string {
	return `
const go2jsNativeMap = globalThis.Map;
const go2jsNativeSet = globalThis.Set;
const go2jsNativeDate = globalThis.Date;
const go2jsNativeTypeError = globalThis.TypeError;
const go2jsNativeRangeError = globalThis.RangeError;
const go2jsNativeError = globalThis.Error;

// The words Go puts in front of every fault it raises on its own.
const go2jsRuntimeErrorPrefix = "runtime error: ";

function go2jsFloat(value) {
	if (Number.isNaN(value)) {
		return "NaN";
	}

	if (value === Infinity) {
		return "+Inf";
	}

	if (value === -Infinity) {
		return "-Inf";
	}

	const text = Math.abs(value).toExponential(6);
	const match = text.match(/^([0-9]\.[0-9]{6})e([+-])([0-9]+)$/);

	if (!match) {
		return (value < 0 ? "-" : "+") + text;
	}

	return (value < 0 ? "-" : "+") + match[1] + "e" + match[2] + match[3].padStart(3, "0");
}

var go2jsTextEncoder = new TextEncoder();
var go2jsStrictDecoder = new TextDecoder("utf-8", { fatal: true });
var go2jsLenientDecoder = new TextDecoder("utf-8");

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

function go2jsOSGetpid() {
    return typeof process === "object" && typeof process.pid === "number" ? process.pid : 1;
}

function go2jsOSGetppid() {
    return typeof process === "object" && typeof process.ppid === "number" ? process.ppid : 1;
}

function go2jsOSGetuid() {
    return typeof process === "object" && typeof process.getuid === "function" ? process.getuid() : -1;
}

function go2jsOSGeteuid() {
    return typeof process === "object" && typeof process.geteuid === "function" ? process.geteuid() : -1;
}

function go2jsOSGetgid() {
    return typeof process === "object" && typeof process.getgid === "function" ? process.getgid() : -1;
}

function go2jsOSGetegid() {
    return typeof process === "object" && typeof process.getegid === "function" ? process.getegid() : -1;
}

function go2jsOSHostname() {
    try {
        return [require("os").hostname(), null];
    } catch (err) {
        return ["", go2jsOSHostError(err, "hostname", "")];
    }
}

function go2jsOSExecutable() {
    return typeof process === "object" && typeof process.execPath === "string" ? process.execPath : "";
}

function go2jsOSGetenv(name) {
    return process.env[String(name)] || "";
}

function go2jsOSSetenv(name, value) {
    process.env[String(name)] = String(value);
    return null;
}

// go2jsOSUnsetenv takes a name out of the environment, and a name that is not
// there is not an error: Go answers nothing either way.
function go2jsOSUnsetenv(name) {
    delete process.env[String(name)];
    return null;
}

// go2jsOSLookupEnv answers a name and whether it was there at all, which is the
// difference between a name set to nothing and a name that is not set.
function go2jsOSLookupEnv(name) {
    const value = process.env[String(name)];

    if (value === undefined) {
        return ["", false];
    }

    return [value, true];
}

// go2jsOSUserHomeDir is the directory a program keeps a user's own files in,
// which is the one the environment names and the one the system has of its own
// accord.
function go2jsOSUserHomeDir() {
    try {
        const home = require("os").homedir();

        if (typeof home === "string" && home !== "") {
            return [home, null];
        }
    } catch (err) {
        return ["", go2jsOSPathError(err, "user home directory")];
    }

    return ["", new Error("neither $HOME nor $USERPROFILE is set")];
}

// go2jsOSUserCacheDir is the directory a program keeps what it has worked out
// between runs in, which is the one the environment names and the one the
// system has of its own accord.
function go2jsOSUserCacheDir() {
    try {
        if (typeof require("os").tmpdir === "function") {
            const cache = require("os").tmpdir();

            if (typeof cache === "string" && cache !== "") {
                return [cache, null];
            }
        }
    } catch (err) {
        return ["", go2jsOSPathError(err, "user cache directory")];
    }

    return ["", new Error("neither $XDG_CACHE_HOME nor $HOME is set")];
}

// go2jsOSStatObject builds the FileInfo a stat call answers with, which is the
// same thing a file answers a stat of its own with.
function go2jsOSStatObject(stats, path) {
    return [go2jsOSFileInfo(path === undefined || path === null ? "" : String(path), stats), null];
}

// go2jsOSLstat describes a path without following a symbolic link, which is
// what Node's lstatSync does. A path that is not one answers the same as stat,
// which is what Go reports for it.
function go2jsOSLstat(path) {
    try {
        return go2jsOSStatObject(require("fs").lstatSync(String(path)), path);
    } catch (err) {
        return [null, go2jsOSHostError(err, "lstat", String(path))];
    }
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

// go2jsTimeFixedZone is a zone kept at a number of seconds east of UTC, which is
// what a name and an offset make without asking the host for anything.
function go2jsTimeFixedZone(name, offset) {
    return go2jsTimeLocation(String(name), Number(offset));
}

// go2jsTimeLoadLocation is a zone the host knows by name, which is asked of the
// host rather than carried along, since where the zones of the world are is not
// something a translation carries. A name the host does not know has no zone,
// and Go answers for one that cannot be found with an error.
function go2jsTimeLoadLocation(name) {
    const wanted = String(name);

    if (wanted === "" || wanted === "UTC") {
        return [go2jsTimeLocation("UTC"), null];
    }

    if (wanted === "Local") {
        return [go2jsTimeHostLocation(), null];
    }

    if (go2jsTimeZoneFormat(wanted) === null) {
        return [null, go2jsTimeLoadLocationError(wanted)];
    }

    return [go2jsTimeLocation(wanted, wanted), null];
}

function go2jsTimeLoadLocationError(name) {
    return go2jsNameError(new Error("unknown time zone " + name), "*errors.errorString");
}

// MonthString and WeekdayString are the String methods of the two named numbers
// the time package keeps a name for, which a method called on a value of either
// type reaches under the name its type and the method make between them.
function MonthString(month) {
	return go2jsMonthName(month);
}

function WeekdayString(weekday) {
	return go2jsWeekdayName(weekday);
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

// go2jsTimeHostLocation is the zone of the machine, which is the zone Go means
// by Local, named after wherever the machine keeps its clock rather than left
// standing as a word for it.
function go2jsTimeHostLocation() {
    const name = go2jsTimeHostZoneName();

    return go2jsTimeLocation(name, name === "UTC" ? null : name);
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
// what moving a moment about carries with it.
function go2jsTimeNanosCarry(value, next) {
    next.__go2js_nsec = value.__go2js_nsec === undefined ? 0 : value.__go2js_nsec;

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
            return this.value.toISOString();
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

            next.__go2js_location = this.__go2js_location;
            next.__go2js_zone = this.__go2js_zone;

            return next;
        },
        Round: function(duration) {
            const rounded = go2jsTimeRounded(this, duration, true);
            const next = go2jsTimeValue(new Date(rounded.millis));

            next.__go2js_nsec = rounded.nsec;

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

// go2jsRegexpCompile mirrors regexp.Compile, which reports a bad pattern as an
// error rather than panicking.
function go2jsRegexpCompile(pattern) {
    try {
        return [go2jsRegexpNew(pattern), null];
    } catch (cause) {
        return [null, new Error("error parsing regexp: " + go2jsStringify(cause.message))];
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

function go2jsFilepathAbs(value) {
	const path = String(value);

	if (go2jsFilepathIsAbs(path)) {
		return [go2jsFilepathClean(path), null];
	}

	let base = "/";

	try {
		if (typeof process !== "undefined" && process.cwd) {
			base = process.cwd();
		}
	} catch (err) {
		base = "/";
	}

	return [go2jsFilepathClean(base + "/" + path), null];
}

function go2jsFilepathIsAbs(value) {
	return String(value).startsWith("/");
}

function go2jsFilepathRel(base, target) {
	const from = go2jsFilepathClean(base);
	const to = go2jsFilepathClean(target);

	if (from === to) {
		return [".", null];
	}

	// A pair of paths only reaches one another when both name a location the
	// same way, so a root in one and not the other is refused.
	if (go2jsFilepathIsAbs(from) !== go2jsFilepathIsAbs(to)) {
		return ["", go2jsErrorsNew("Rel: can't make " + String(target) + " relative to " + String(base))];
	}

	const fromParts = from.split("/").filter(part => part !== "");
	const toParts = to.split("/").filter(part => part !== "");

	let shared = 0;

	while (shared < fromParts.length && shared < toParts.length && fromParts[shared] === toParts[shared]) {
		shared++;
	}

	// A path cannot step above a directory it was never inside of.
	if (fromParts[shared] === "..") {
		return ["", go2jsErrorsNew("Rel: can't make " + String(target) + " relative to " + String(base))];
	}

	const parts = [];

	for (let i = shared; i < fromParts.length; i++) {
		parts.push("..");
	}

	for (let i = shared; i < toParts.length; i++) {
		parts.push(toParts[i]);
	}

	return [parts.join("/"), null];
}

function go2jsFilepathSplit(value) {
	const path = String(value);
	const index = path.lastIndexOf("/");

	// The directory a split leaves behind still carries the separator that
	// divided it from the file, so joining the two gives the path again.
	if (index < 0) {
		return ["", path];
	}

	return [path.slice(0, index + 1), path.slice(index + 1)];
}

function go2jsFilepathToSlash(value) {
	return String(value).split("\\").join("/");
}

// A pattern names the files it matches with the marks of a shell, and each of
// them stands for one thing: a star stands for any run of characters that are
// not a separator, a question for one such character, and a class for one of
// the characters it lists. The pattern is written as a pattern of the same
// shape rather than matched a mark at a time, because the marks are only ever
// matched together.
function go2jsFilepathMatch(pattern, name) {
	const compiled = go2jsFilepathPatternSource(pattern);

	if (compiled[1] !== null) {
		return compiled;
	}

	return [new RegExp(compiled[0]).test(go2jsStringify(name)), null];
}

// go2jsFilepathPatternSource writes a pattern as one of the same shape, which is
// what a name is asked against, and reports a pattern that cannot be written.
function go2jsFilepathPatternSource(pattern) {
	let source = "^";

	for (let index = 0; index < pattern.length; index++) {
		const char = pattern[index];

		if (char === "*") {
			source += "[^/]*";
			continue;
		}

		if (char === "?") {
			source += "[^/]";
			continue;
		}

		if (char === "\\") {
			if (index + 1 >= pattern.length) {
				return [false, go2jsErrorsNew("syntax error in pattern")];
			}

			source += go2jsFilepathQuote(pattern[index + 1]);
			index++;
			continue;
		}

		if (char === "[") {
			let cursor = index + 1;
			let negated = false;
			let body = "";

			if (pattern[cursor] === "^") {
				negated = true;
				cursor++;
			}

			// A class is a list of characters and of ranges between them, and a
			// range is written with a mark between its ends. A mark that has
			// nothing to end is not a range, and neither is a class with nothing
			// in it, because there is nothing a single character could be told
			// apart from.
			for (let ranges = 0; ; ranges++) {
				if (pattern[cursor] === "]" && ranges > 0) {
					cursor++;
					break;
				}

				const lo = go2jsFilepathClassChar(pattern, cursor);

				if (lo === null) {
					return [false, go2jsErrorsNew("syntax error in pattern")];
				}

				cursor += Number(lo[0].slice(1));
				let piece = lo[1];

				if (pattern[cursor] === "-") {
					const hi = go2jsFilepathClassChar(pattern, cursor + 1);

					if (hi === null) {
						return [false, go2jsErrorsNew("syntax error in pattern")];
					}

					cursor += 1 + Number(hi[0].slice(1));
					piece += "-" + hi[1];
				}

				body += piece;
			}

			source += "[" + (negated ? "^" : "") + body + "]";
			index = cursor - 1;
			continue;
		}

		source += go2jsFilepathQuote(char);
	}

	source += "$";

	return [source, null];
}

// go2jsFilepathClassChar reads one character of a class as it was written,
// which may be a mark that stands for the character after it, and reports what
// to write for it along with how much of the pattern it took. A mark that
// stands for the end of the class, or for a range, is not a character, because
// neither is a thing a class can hold.
function go2jsFilepathClassChar(pattern, index) {
	const char = pattern[index];

	if (char === undefined || char === "]" || char === "-") {
		return null;
	}

	if (char === "\\" && pattern[index + 1] === undefined) {
		return null;
	}

	const taken = char === "\\" ? 2 : 1;

	if (taken === 2) {
		return ["\u00002", go2jsFilepathClassQuote(pattern[index + 1])];
	}

	return ["\u00001", go2jsFilepathClassQuote(char)];
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

// The marks of a file in Go are in the places Go keeps them, and the host takes
// only the marks of what may be done with it, which are the same nine.
function go2jsFileModeToHost(mode) {
	return mode & 511;
}

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

// A mark of a file is asked of as the mark it is, which is the number behind it.
function go2jsFileModeStringMethod(mode) {
	return go2jsFileModeString(go2jsFileModeBits(mode));
}

function go2jsFileModeIsDirMethod(mode) {
	return (go2jsFileModeBits(mode) & 2147483648) !== 0;
}

function go2jsFileModeIsRegularMethod(mode) {
	return (go2jsFileModeBits(mode) & 2401763328) === 0;
}

function go2jsFileModeTypeMethod(mode) {
	return go2jsFileMode(go2jsFileModeBits(mode) & 2401763328);
}

function go2jsFileModePermMethod(mode) {
	return go2jsFileMode(go2jsFileModeBits(mode) & 511);
}

// go2jsFilepathClassQuote writes one character of a class, where a character
// that closes the class, or starts it, or is a mark of its own stands for
// itself.
function go2jsFilepathClassQuote(char) {
	return "]$^-^\\".indexOf(char) === -1 ? char : "\\" + char;
}

function go2jsFilepathQuote(char) {
	return /[.*+?^${}()|[\]\\]/.test(char) ? "\\" + char : char;
}

function go2jsFilepathVolumeName(value) {
	return "";
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

function go2jsURLPathEscape(value) {
	return go2jsURLEncode(value, "segment");
}

function go2jsURLFragmentEscape(value) {
	return go2jsURLEncode(value, "fragment");
}

function go2jsURLQueryUnescape(value) {
	return go2jsURLDecode(value, "query");
}

function go2jsURLPathUnescape(value) {
	return go2jsURLDecode(value, "segment");
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

function go2jsURLUser(name) {
	return go2jsURLUserinfo(go2jsStringify(name), "", false);
}

function go2jsURLUserPassword(name, password) {
	return go2jsURLUserinfo(go2jsStringify(name), go2jsStringify(password), true);
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

// A request that names what it wants is either an absolute URL or a path from
// the root, so a URI is read the same way and then asked whether it is one.
function go2jsURLParseRequestURI(value) {
	const parsed = go2jsURLParse(value);
	const url = parsed[0];

	if (parsed[1] !== null) {
		return parsed;
	}

	if (url.Scheme === "" && !go2jsStringify(value).startsWith("/")) {
		return [null, go2jsNameError(
			new Error("parse " + JSON.stringify(go2jsStringify(value)) +
				": invalid URI for request"), "*url.Error")];
	}

	return [url, null];
}

function go2jsURLParseQuery(raw) {
	let failure = null;
	const values = go2jsURLQueryFromRaw(raw, (message) => {
		if (failure === null) {
			failure = message;
		}
	});

	if (failure !== null) {
		return [values, go2jsErrorsNew(failure)];
	}

	return [values, null];
}

// Joining a path to a URL gives the URL back as it is written, so what is
// returned here is the whole of it rather than the URL.
function go2jsURLJoinPath(base, ...elems) {
	const [url, err] = go2jsURLParse(base);

	if (err !== null) {
		return ["", go2jsURLError("parse", go2jsStringify(base), err)];
	}

	return [go2jsURLString(url.JoinPath(...elems)), null];
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

// go2jsJSONRawMessage is a json.RawMessage made from the text given it, which is
// all a raw message is: the text of a value, held as that text.
function go2jsJSONRawMessage(value) {
    if (typeof value === "string") {
        return value;
    }

    return go2jsBytesToString(value);
}

// go2jsJSONMarshalAsText writes out a value that is itself text rather than a
// value JSON describes: a json.RawMessage, which is the text of a value that
// has not been read yet, and a json.Number, which is the text of a number. The
// text is written as it stands rather than as a string, and a json.Number is
// written as a number only when it is one, since a text that is not a number is
// not one to write.
function go2jsJSONMarshalAsText(value, typeName, prefix, indent, escapeHTML) {
    const text = go2jsJSONNumberText(value);

    try {
        if (typeName === "json.RawMessage") {
            // A raw message is the text of a value that has not been read yet,
            // so what it is written out as is the value that text describes
            // with the room taken out of it, and with the room put back by
            // whatever the caller laid the rest of the value out with. Every
            // token is written as it was written, so the digits of a number
            // inside it are the digits it was written with.
            const compact = go2jsJSONCompact(text);

            if (prefix === "" && indent === "") {
                return [escapeHTML === false ? compact : go2jsJSONEscapeHTML(compact), null];
            }

            return [go2jsJSONEscapeHTML(go2jsJSONIndent(compact, prefix, indent)), null];
        }

        if (!go2jsJSONNumberLiteral(text)) {
            return [null, new Error("json: invalid number literal " + go2jsJSONWritten(text))];
        }

        // A number is written as the digits it holds, since a number a double
        // holds is written as the double and a number with more digits than a
        // double keeps would lose them.
        return [escapeHTML === false ? text : go2jsJSONEscapeHTML(text), null];
    } catch (err) {
        return [null, new Error("json: error calling MarshalJSON for type " + typeName + ": invalid character")];
    }
}

// go2jsJSONEncoderEncodeAsText is an encoder handed a value that is itself text
// rather than a value JSON describes, which is written as the text it stands
// for and with whatever the encoder was set up to write with.
function go2jsJSONEncoderEncodeAsText(encoder, value, typeName) {
    const written = go2jsJSONMarshalAsText(value, typeName, encoder.prefix, encoder.indent, encoder.escapeHTML);

    if (written[1] !== null && written[1] !== undefined) {
        return written[1];
    }

    return go2jsJSONWrite(encoder.writer, written[0] + "\n");
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

// go2jsJSONScanStep is the state machine encoding/json uses to validate JSON
// text, so that the errors json.Valid, json.Compact and json.Indent report are
// the errors Go reports. The machine follows the scanner Go ships: statuses
// mirror the opcodes scanContinue through scanError, and a parseState stack
// holds what an object key, an object value, or an array value is expected
// next. A space is fed to the machine once the text is out, exactly the way
// Go's scanner.eof does.
const go2jsJSONScanContinue = 0;
const go2jsJSONScanBeginLiteral = 1;
const go2jsJSONScanBeginObject = 2;
const go2jsJSONScanObjectKey = 3;
const go2jsJSONScanObjectValue = 4;
const go2jsJSONScanEndObject = 5;
const go2jsJSONScanBeginArray = 6;
const go2jsJSONScanArrayValue = 7;
const go2jsJSONScanEndArray = 8;
const go2jsJSONScanSkipSpace = 9;
const go2jsJSONScanEnd = 10;
const go2jsJSONScanError = 11;

const go2jsJSONParseObjectKey = 0;
const go2jsJSONParseObjectValue = 1;
const go2jsJSONParseArrayValue = 2;

function go2jsJSONScanNew() {
    return { step: "beginValue", endTop: false, err: null, depth: 0, parse: [] };
}

function go2jsJSONScanSpace(character) {
    return character === 0x20 || character === 0x09 || character === 0x0a || character === 0x0d;
}

function go2jsJSONQuoteChar(character) {
    if (character === 0x27) {
        return "'\\''";
    }

    if (character === 0x22) {
        return "'\"'";
    }

    let body;

    if (character === 0x5c) {
        body = "\\\\";
    } else if (character >= 0x20 && character <= 0x7e) {
        body = String.fromCharCode(character);
    } else if (character === 0x07) {
        body = "\\a";
    } else if (character === 0x08) {
        body = "\\b";
    } else if (character === 0x09) {
        body = "\\t";
    } else if (character === 0x0a) {
        body = "\\n";
    } else if (character === 0x0b) {
        body = "\\v";
    } else if (character === 0x0c) {
        body = "\\f";
    } else if (character === 0x0d) {
        body = "\\r";
    } else {
        body = "\\x" + character.toString(16).padStart(2, "0");
    }

    return "'" + body + "'";
}

function go2jsJSONScanFail(scanner, character, context) {
    scanner.err = "invalid character " + go2jsJSONQuoteChar(character) + " " + context;

    return go2jsJSONScanError;
}

function go2jsJSONScanStep(scanner, character) {
    switch (scanner.step) {
        case "beginValue":
            return go2jsJSONScanStateBeginValue(scanner, character);

        case "beginValueOrEmpty":
            return go2jsJSONScanStateBeginValueOrEmpty(scanner, character);

        case "beginStringOrEmpty":
            return go2jsJSONScanStateBeginStringOrEmpty(scanner, character);

        case "beginString":
            return go2jsJSONScanStateBeginString(scanner, character);

        case "inString":
            return go2jsJSONScanStateInString(scanner, character);

        case "inStringEsc":
            return go2jsJSONScanStateInStringEsc(scanner, character);

        case "inStringEscU":
        case "inStringEscU1":
        case "inStringEscU12":
        case "inStringEscU123":
            return go2jsJSONScanStateInStringEscU(scanner, character);

        case "neg":
            return go2jsJSONScanStateNeg(scanner, character);

        case "zero":
            return go2jsJSONScanStateZero(scanner, character);

        case "one":
            return go2jsJSONScanStateOne(scanner, character);

        case "dot":
            return go2jsJSONScanStateDot(scanner, character);

        case "dot0":
            return go2jsJSONScanStateDot0(scanner, character);

        case "e":
            return go2jsJSONScanStateE(scanner, character);

        case "esign":
            return go2jsJSONScanStateESign(scanner, character);

        case "e0":
            return go2jsJSONScanStateE0(scanner, character);

        case "t":
            return go2jsJSONScanStateT(scanner, character);

        case "tr":
            return go2jsJSONScanStateTr(scanner, character);

        case "tru":
            return go2jsJSONScanStateTru(scanner, character);

        case "f":
            return go2jsJSONScanStateF(scanner, character);

        case "fa":
            return go2jsJSONScanStateFa(scanner, character);

        case "fal":
            return go2jsJSONScanStateFal(scanner, character);

        case "fals":
            return go2jsJSONScanStateFals(scanner, character);

        case "n":
            return go2jsJSONScanStateN(scanner, character);

        case "nu":
            return go2jsJSONScanStateNu(scanner, character);

        case "nul":
            return go2jsJSONScanStateNul(scanner, character);

        case "endValue":
            return go2jsJSONScanStateEndValue(scanner, character);

        case "endTop":
            return go2jsJSONScanStateEndTop(scanner, character);

        default:
            return go2jsJSONScanError;
    }
}

function go2jsJSONScanStateBeginValue(scanner, character) {
    if (go2jsJSONScanSpace(character)) {
        return go2jsJSONScanSkipSpace;
    }

    if (character === 0x7b) {
        scanner.step = "beginStringOrEmpty";
        scanner.parse.push(go2jsJSONParseObjectKey);

        return go2jsJSONScanBeginObject;
    }

    if (character === 0x5b) {
        scanner.step = "beginValueOrEmpty";
        scanner.parse.push(go2jsJSONParseArrayValue);

        return go2jsJSONScanBeginArray;
    }

    if (character === 0x22) {
        scanner.step = "inString";

        return go2jsJSONScanBeginLiteral;
    }

    if (character === 0x2d) {
        scanner.step = "neg";

        return go2jsJSONScanBeginLiteral;
    }

    if (character === 0x30) {
        scanner.step = "zero";

        return go2jsJSONScanBeginLiteral;
    }

    if (character === 0x74) {
        scanner.step = "t";

        return go2jsJSONScanBeginLiteral;
    }

    if (character === 0x66) {
        scanner.step = "f";

        return go2jsJSONScanBeginLiteral;
    }

    if (character === 0x6e) {
        scanner.step = "n";

        return go2jsJSONScanBeginLiteral;
    }

    if (character >= 0x31 && character <= 0x39) {
        scanner.step = "one";

        return go2jsJSONScanBeginLiteral;
    }

    return go2jsJSONScanFail(scanner, character, "looking for beginning of value");
}

function go2jsJSONScanStateBeginValueOrEmpty(scanner, character) {
    if (go2jsJSONScanSpace(character)) {
        return go2jsJSONScanSkipSpace;
    }

    if (character === 0x5d) {
        return go2jsJSONScanStateEndValue(scanner, character);
    }

    return go2jsJSONScanStateBeginValue(scanner, character);
}

function go2jsJSONScanStateBeginStringOrEmpty(scanner, character) {
    if (go2jsJSONScanSpace(character)) {
        return go2jsJSONScanSkipSpace;
    }

    if (character === 0x7d) {
        scanner.parse[scanner.parse.length - 1] = go2jsJSONParseObjectValue;

        return go2jsJSONScanStateEndValue(scanner, character);
    }

    return go2jsJSONScanStateBeginString(scanner, character);
}

function go2jsJSONScanStateBeginString(scanner, character) {
    if (go2jsJSONScanSpace(character)) {
        return go2jsJSONScanSkipSpace;
    }

    if (character === 0x22) {
        scanner.step = "inString";

        return go2jsJSONScanBeginLiteral;
    }

    return go2jsJSONScanFail(scanner, character, "looking for beginning of object key string");
}

function go2jsJSONScanStateInString(scanner, character) {
    if (character === 0x22) {
        scanner.step = "endValue";

        return go2jsJSONScanContinue;
    }

    if (character === 0x5c) {
        scanner.step = "inStringEsc";

        return go2jsJSONScanContinue;
    }

    if (character < 0x20) {
        return go2jsJSONScanFail(scanner, character, "in string literal");
    }

    return go2jsJSONScanContinue;
}

function go2jsJSONScanStateInStringEsc(scanner, character) {
    if (character === 0x22 || character === 0x2f || character === 0x5c
        || character === 0x62 || character === 0x66 || character === 0x6e
        || character === 0x72 || character === 0x74) {
        scanner.step = "inString";

        return go2jsJSONScanContinue;
    }

    if (character === 0x75) {
        scanner.step = "inStringEscU";

        return go2jsJSONScanContinue;
    }

    return go2jsJSONScanFail(scanner, character, "in string escape code");
}

function go2jsJSONScanStateInStringEscU(scanner, character) {
    const hex = character >= 0x30 && character <= 0x39
        || character >= 0x61 && character <= 0x66
        || character >= 0x41 && character <= 0x46;

    if (hex) {
        if (scanner.step === "inStringEscU") {
            scanner.step = "inStringEscU1";
        } else if (scanner.step === "inStringEscU1") {
            scanner.step = "inStringEscU12";
        } else if (scanner.step === "inStringEscU12") {
            scanner.step = "inStringEscU123";
        } else {
            scanner.step = "inString";
        }

        return go2jsJSONScanContinue;
    }

    return go2jsJSONScanFail(scanner, character, "in \\u hexadecimal character escape");
}

function go2jsJSONScanStateNeg(scanner, character) {
    if (character === 0x30) {
        scanner.step = "zero";

        return go2jsJSONScanContinue;
    }

    if (character >= 0x31 && character <= 0x39) {
        scanner.step = "one";

        return go2jsJSONScanContinue;
    }

    return go2jsJSONScanFail(scanner, character, "in numeric literal");
}

function go2jsJSONScanStateZero(scanner, character) {
    if (character === 0x2e) {
        scanner.step = "dot";

        return go2jsJSONScanContinue;
    }

    if (character === 0x65 || character === 0x45) {
        scanner.step = "e";

        return go2jsJSONScanContinue;
    }

    return go2jsJSONScanStateEndValue(scanner, character);
}

function go2jsJSONScanStateOne(scanner, character) {
    if (character >= 0x30 && character <= 0x39) {
        return go2jsJSONScanContinue;
    }

    return go2jsJSONScanStateZero(scanner, character);
}

function go2jsJSONScanStateDot(scanner, character) {
    if (character >= 0x30 && character <= 0x39) {
        scanner.step = "dot0";

        return go2jsJSONScanContinue;
    }

    return go2jsJSONScanFail(scanner, character, "after decimal point in numeric literal");
}

function go2jsJSONScanStateDot0(scanner, character) {
    if (character >= 0x30 && character <= 0x39) {
        return go2jsJSONScanContinue;
    }

    if (character === 0x65 || character === 0x45) {
        scanner.step = "e";

        return go2jsJSONScanContinue;
    }

    return go2jsJSONScanStateEndValue(scanner, character);
}

function go2jsJSONScanStateE(scanner, character) {
    if (character === 0x2b || character === 0x2d) {
        scanner.step = "esign";

        return go2jsJSONScanContinue;
    }

    return go2jsJSONScanStateESign(scanner, character);
}

function go2jsJSONScanStateESign(scanner, character) {
    if (character >= 0x30 && character <= 0x39) {
        scanner.step = "e0";

        return go2jsJSONScanContinue;
    }

    return go2jsJSONScanFail(scanner, character, "in exponent of numeric literal");
}

function go2jsJSONScanStateE0(scanner, character) {
    if (character >= 0x30 && character <= 0x39) {
        return go2jsJSONScanContinue;
    }

    return go2jsJSONScanStateEndValue(scanner, character);
}

function go2jsJSONScanStateT(scanner, character) {
    if (character === 0x72) {
        scanner.step = "tr";

        return go2jsJSONScanContinue;
    }

    return go2jsJSONScanFail(scanner, character, "in literal true (expecting 'r')");
}

function go2jsJSONScanStateTr(scanner, character) {
    if (character === 0x75) {
        scanner.step = "tru";

        return go2jsJSONScanContinue;
    }

    return go2jsJSONScanFail(scanner, character, "in literal true (expecting 'u')");
}

function go2jsJSONScanStateTru(scanner, character) {
    if (character === 0x65) {
        scanner.step = "endValue";

        return go2jsJSONScanContinue;
    }

    return go2jsJSONScanFail(scanner, character, "in literal true (expecting 'e')");
}

function go2jsJSONScanStateF(scanner, character) {
    if (character === 0x61) {
        scanner.step = "fa";

        return go2jsJSONScanContinue;
    }

    return go2jsJSONScanFail(scanner, character, "in literal false (expecting 'a')");
}

function go2jsJSONScanStateFa(scanner, character) {
    if (character === 0x6c) {
        scanner.step = "fal";

        return go2jsJSONScanContinue;
    }

    return go2jsJSONScanFail(scanner, character, "in literal false (expecting 'l')");
}

function go2jsJSONScanStateFal(scanner, character) {
    if (character === 0x73) {
        scanner.step = "fals";

        return go2jsJSONScanContinue;
    }

    return go2jsJSONScanFail(scanner, character, "in literal false (expecting 's')");
}

function go2jsJSONScanStateFals(scanner, character) {
    if (character === 0x65) {
        scanner.step = "endValue";

        return go2jsJSONScanContinue;
    }

    return go2jsJSONScanFail(scanner, character, "in literal false (expecting 'e')");
}

function go2jsJSONScanStateN(scanner, character) {
    if (character === 0x75) {
        scanner.step = "nu";

        return go2jsJSONScanContinue;
    }

    return go2jsJSONScanFail(scanner, character, "in literal null (expecting 'u')");
}

function go2jsJSONScanStateNu(scanner, character) {
    if (character === 0x6c) {
        scanner.step = "nul";

        return go2jsJSONScanContinue;
    }

    return go2jsJSONScanFail(scanner, character, "in literal null (expecting 'l')");
}

function go2jsJSONScanStateNul(scanner, character) {
    if (character === 0x6c) {
        scanner.step = "endValue";

        return go2jsJSONScanContinue;
    }

    return go2jsJSONScanFail(scanner, character, "in literal null (expecting 'l')");
}

function go2jsJSONScanStateEndTop(scanner, character) {
    if (!go2jsJSONScanSpace(character)) {
        return go2jsJSONScanFail(scanner, character, "after top-level value");
    }

    return go2jsJSONScanEnd;
}
function go2jsJSONScanStateEndValue(scanner, character) {
    const n = scanner.parse.length;

    if (n === 0) {
        scanner.step = "endTop";
        scanner.endTop = true;

        return go2jsJSONScanStep(scanner, character);
    }

    if (go2jsJSONScanSpace(character)) {
        scanner.step = "endValue";

        return go2jsJSONScanSkipSpace;
    }

    const parsed = scanner.parse[n - 1];

    if (parsed === go2jsJSONParseObjectKey) {
        if (character === 0x3a) {
            scanner.parse[n - 1] = go2jsJSONParseObjectValue;
            scanner.step = "beginValue";

            return go2jsJSONScanObjectKey;
        }

        return go2jsJSONScanFail(scanner, character, "after object key");
    }

    if (parsed === go2jsJSONParseObjectValue) {
        if (character === 0x2c) {
            scanner.parse[n - 1] = go2jsJSONParseObjectKey;
            scanner.step = "beginString";

            return go2jsJSONScanObjectValue;
        }

        if (character === 0x7d) {
            scanner.parse.pop();
            go2jsJSONScanPop(scanner);

            return go2jsJSONScanEndObject;
        }

        return go2jsJSONScanFail(scanner, character, "after object key:value pair");
    }

    if (parsed === go2jsJSONParseArrayValue) {
        if (character === 0x2c) {
            scanner.step = "beginValue";

            return go2jsJSONScanArrayValue;
        }

        if (character === 0x5d) {
            scanner.parse.pop();
            go2jsJSONScanPop(scanner);

            return go2jsJSONScanEndArray;
        }

        return go2jsJSONScanFail(scanner, character, "after array element");
    }

    return go2jsJSONScanFail(scanner, 0, "");
}

function go2jsJSONScanPop(scanner) {
    const n = scanner.parse.length;

    if (n === 0) {
        scanner.step = "endTop";
        scanner.endTop = true;
    } else {
        scanner.step = "endValue";
    }
}

function go2jsJSONScanEOF(scanner) {
    if (scanner.err !== null) {
        return go2jsJSONScanError;
    }

    if (scanner.endTop) {
        return go2jsJSONScanEnd;
    }

    go2jsJSONScanStep(scanner, 0x20);

    if (scanner.endTop) {
        return go2jsJSONScanEnd;
    }

    if (scanner.err === null) {
        scanner.err = "unexpected end of JSON input";
    }

    return go2jsJSONScanError;
}

// go2jsJSONScanErrorOf validates the text of a JSON value the way the Go
// scanner does and returns the error Go would report, or nothing for a value
// the scanner accepts. It is the whole check json.Valid, json.Compact and
// json.Indent make before they act.
function go2jsJSONScanErrorOf(text) {
    const scanner = go2jsJSONScanNew();

    for (let i = 0; i < text.length; i++) {
        if (go2jsJSONScanStep(scanner, text.charCodeAt(i)) === go2jsJSONScanError) {
            return scanner.err;
        }
    }

    if (go2jsJSONScanEOF(scanner) === go2jsJSONScanError) {
        return scanner.err;
    }

    return "";
}

// go2jsJSONValid reports whether the text is a value the Go scanner accepts.
function go2jsJSONIsValid(data) {
    return go2jsJSONScanErrorOf(go2jsJSONByteText(data)) === "";
}

// go2jsJSONIndentTo lays the text of a value out the way encoding/json's
// appendIndent does, byte for byte, so that a space left at the end of the text
// survives it and an empty object or an empty array is not broken across
// lines. It returns the error Go would report for text the scanner rejects.
function go2jsJSONIndentTo(text, prefix, indent) {
    const scanner = go2jsJSONScanNew();
    let out = "";
    let needIndent = false;
    let depth = 0;

    for (let i = 0; i < text.length; i++) {
        const character = text.charCodeAt(i);
        const v = go2jsJSONScanStep(scanner, character);

        if (v === go2jsJSONScanSkipSpace) {
            continue;
        }

        if (v === go2jsJSONScanError) {
            return [out, scanner.err];
        }

        if (needIndent && v !== go2jsJSONScanEndObject && v !== go2jsJSONScanEndArray) {
            needIndent = false;
            depth++;

            out += "\n" + prefix + indent.repeat(depth);
        }

        if (v === go2jsJSONScanContinue) {
            out += String.fromCharCode(character);

            continue;
        }

        switch (character) {
            case 0x7b: // {
            case 0x5b: // [
                needIndent = true;
                out += String.fromCharCode(character);
                break;

            case 0x2c: // ,
                out += String.fromCharCode(character);
                out += "\n" + prefix + indent.repeat(depth);
                break;

            case 0x3a: // :
                out += ": ";
                break;

            case 0x7d: // }
            case 0x5d: // ]
                if (needIndent) {
                    needIndent = false;
                } else {
                    depth--;
                    out += "\n" + prefix + indent.repeat(depth);
                }

                out += String.fromCharCode(character);
                break;

            default:
                out += String.fromCharCode(character);
        }
    }

    if (go2jsJSONScanEOF(scanner) === go2jsJSONScanError) {
        return [out, scanner.err];
    }

    return [out, null];
}

// go2jsJSONCompactTo reports the compacted form of a value text Go would
// report for it, or the error Go would report instead. The text is validated
// the way encoding/json's appendCompact scans it before the room between
// tokens is taken out, so that a text that is not a whole JSON value is never
// written as a smaller one.
function go2jsJSONCompactTo(text) {
    const error = go2jsJSONScanErrorOf(String(text));

    if (error !== "") {
        return [null, error];
    }

    return [go2jsJSONCompact(text), null];
}

// go2jsJSONByteText is the text the bytes of a value stand for, read the way
// json.Unmarshal reads the bytes it is handed.
function go2jsJSONByteText(data) {
    if (typeof data === "string") {
        return data;
    }

    return new TextDecoder().decode(Uint8Array.from(data));
}

// go2jsJSONCompactToBuffer writes the compacted form of a value text to a
// buffer, the way json.Compact does, and returns the error Go would report
// for text that is not a whole value.
function go2jsJSONCompactToBuffer(dst, src) {
    const result = go2jsJSONCompactTo(go2jsJSONByteText(src));

    if (result[1] !== null && result[1] !== undefined) {
        return result[1];
    }

    return go2jsJSONWrite(dst, result[0]);
}

// go2jsJSONIndentToBuffer writes an indented form of a value text to a buffer,
// the way json.Indent does, and returns the error Go would report for text
// that is not a whole value. The text is laid out exactly as encoding/json
// lays one out, byte for byte.
function go2jsJSONIndentToBuffer(dst, src, prefix, indent) {
    const result = go2jsJSONIndentTo(go2jsJSONByteText(src), String(prefix), String(indent));

    if (result[1] !== null && result[1] !== undefined) {
        return result[1];
    }

    return go2jsJSONWrite(dst, result[0]);
}

// go2jsJSONHTMLEscapeToBuffer writes a value text to a buffer with the five
// characters that make JSON unsafe to embed in HTML escaped, the way
// json.HTMLEscape does. The text is not validated, any more than Go
// validates one.
function go2jsJSONHTMLEscapeToBuffer(dst, src) {
    return go2jsJSONWrite(dst, go2jsJSONEscapeHTML(go2jsJSONByteText(src)));
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

// A json.Number is the text a number was written with, so what the type's own
// methods read are the methods of the text it holds. The names carry the Java
// Script spelling of Number because Number is a global of the language itself.
function Number$go2jsString(value) {
    return go2jsJSONNumberText(value);
}

// Number$go2jsInt64 is the number as a whole number the width of a machine, read
// the way strconv reads one, so what is wrong with text that is not a whole
// number is said the way strconv says it.
function Number$go2jsInt64(value) {
    const text = go2jsJSONNumberText(value);

    if (/^[+-]?\d+$/.test(text)) {
        const parsed = Number(text);

        if (Number.isSafeInteger(parsed)) {
            return [parsed, null];
        }

        return [0, new Error("strconv.ParseInt: parsing " + JSON.stringify(text) + ": value out of range")];
    }

    return [0, new Error("strconv.ParseInt: parsing " + JSON.stringify(text) + ": invalid syntax")];
}

// Number$go2jsFloat64 is the number as a double, read the way strconv reads one.
function Number$go2jsFloat64(value) {
    const text = go2jsJSONNumberText(value);

    if (/^[+-]?(\d+(\.\d*)?|\.\d+)([eE][+-]?\d+)?$/.test(text)) {
        return [Number(text), null];
    }

    return [0, new Error("strconv.ParseFloat: parsing " + JSON.stringify(text) + ": invalid syntax")];
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


function go2jsJSONNewEncoder(writer) {
    return {
        __go2js_json_encoder: true,
        writer: writer,
        prefix: "",
        indent: "",
        escapeHTML: true
    };
}

function go2jsJSONEncoderEncode(encoder, value, fields, omitEmpty, stringFields, textFields) {
    try {
        const encoded = go2jsJSONEncode(value, fields, omitEmpty, stringFields, textFields, encoder.prefix, encoder.indent);
        const text = go2jsJSONText(encoded, encoder.prefix, encoder.indent, encoder.escapeHTML);

        return go2jsJSONWrite(encoder.writer, text + "\n");
    } catch (err) {
        return err;
    }
}

// go2jsJSONWrite hands text to an io.Writer as the bytes it stands for, the way
// a program that writes a string to a file does.
function go2jsJSONWrite(writer, text) {
    const target = go2jsUnwrap(writer);

    if (target === null || target === undefined) {
        return go2jsJSONError("json: writer is nil");
    }

    if (typeof target.Write !== "function") {
        return go2jsJSONError("json: writer does not implement io.Writer");
    }

    const result = go2jsCallNow(target.Write, target, [go2jsStringToBytes(text)]);

    if (Array.isArray(result)) {
        return result[1] === null || result[1] === undefined ? null : result[1];
    }

    return result === undefined ? null : result;
}

function go2jsJSONEncoderSetIndent(encoder, prefix, indent) {
    encoder.prefix = String(prefix);
    encoder.indent = String(indent);

    return null;
}

function go2jsJSONEncoderSetEscapeHTML(encoder, escape) {
    encoder.escapeHTML = escape !== false;

    return null;
}

// go2jsJSONNewDecoder returns a decoder that reads values out of an io.Reader.
// The text it has read is held onto between calls, so a stream of values that
// arrive without anything separating them can be read one after another.
function go2jsJSONNewDecoder(reader) {
    return {
        __go2js_json_decoder: true,
        reader: reader,
        buffer: "",
        position: 0,
        eof: false,
        stack: []
    };
}

// go2jsJSONDecoderFill reads whatever more the reader has to offer. A reader
// that has nothing left sets the flag instead of asking again.
function go2jsJSONDecoderFill(decoder) {
    const reader = go2jsUnwrap(decoder.reader);

    if (reader === null || reader === undefined || typeof reader.Read !== "function") {
        decoder.eof = true;
        return;
    }

    const chunk = new Array(4096);
    const result = go2jsCallNow(reader.Read, reader, [chunk]);
    const read = Number(result[0]) || 0;
    const err = result[1] === undefined ? null : result[1];

    if (read > 0) {
        decoder.buffer += new TextDecoder().decode(Uint8Array.from(chunk.slice(0, read)));
    }

    if (err !== null || read <= 0) {
        decoder.eof = true;
    }
}

// go2jsJSONDecoderSkipSpace leaves past the space between values, reading more
// when the text in hand ends before the next value starts.
function go2jsJSONDecoderSkipSpace(decoder) {
    for (;;) {
        while (decoder.position < decoder.buffer.length && /\s/.test(decoder.buffer[decoder.position])) {
            decoder.position++;
        }

        if (decoder.position < decoder.buffer.length || decoder.eof) {
            return;
        }

        go2jsJSONDecoderFill(decoder);
    }
}

// go2jsJSONDecoderOne reads the text of the value that starts at the position.
// The text ends where the value ends rather than where the next whitespace is,
// which is what lets a stream of values sit next to each other.
function go2jsJSONDecoderOne(decoder) {
    go2jsJSONDecoderSkipSpace(decoder);

    if (decoder.position >= decoder.buffer.length) {
        // a stream that has run out is the end of the input, not a failure
        return [null, go2jsIOEOF()];
    }

    const start = decoder.position;
    const text = decoder.buffer;
    const first = text[start];
    let depth = 0;
    let inString = false;
    let escaped = false;

    // a value that is not wrapped in brackets ends at the space or the comma
    // after it, while one that is wrapped ends where its last bracket closes
    const bracketed = first === "{" || first === "[";

    for (let index = start; index < text.length; index++) {
        const ch = text[index];

        if (inString) {
            if (escaped) {
                escaped = false;
            } else if (ch === "\\") {
                escaped = true;
            } else if (ch === "\"") {
                inString = false;
            }

            continue;
        }

        if (ch === "\"") {
            inString = true;
            continue;
        }

        if (ch === "{" || ch === "[") {
            depth++;
            continue;
        }

        if (ch === "}" || ch === "]") {
            if (depth === 0) {
                return [null, go2jsJSONError("invalid character after top-level value")];
            }

            depth--;

            if (depth === 0) {
                return [text.slice(start, index + 1), null];
            }

            continue;
        }

        if (!bracketed && index > start && (/\s/.test(ch) || ch === ",")) {
            return [text.slice(start, index), null];
        }
    }

    if (inString || depth > 0) {
        return [null, go2jsJSONError("unexpected EOF")];
    }

    return [text.slice(start), null];
}

function go2jsJSONDecoderDecode(decoder, target, fields, destination, stringFields, fieldTypes) {
    const [text, err] = go2jsJSONDecoderOne(decoder);

    if (err !== null) {
        return err;
    }

    decoder.position += text.length;

    return go2jsJSONUnmarshal(text, target, fields, destination, stringFields, fieldTypes);
}

// go2jsJSONDecoderMore reports whether another value is left to read, which is
// what a loop over a stream asks before it reads one.
function go2jsJSONDecoderMore(decoder) {
    go2jsJSONDecoderSkipSpace(decoder);

    return decoder.position < decoder.buffer.length;
}

// go2jsJSONDecoderBuffered is the text the decoder has read but not yet handed
// out, which is what a program writes to the start of what follows.
function go2jsJSONDecoderBuffered(decoder) {
    return go2jsJSONCoerceString(decoder.buffer.slice(decoder.position));
}

// go2jsJSONDecoderToken hands back one piece of the document at a time.
function go2jsJSONDecoderToken(decoder) {
    for (;;) {
        go2jsJSONDecoderSkipSpace(decoder);
        if (decoder.position >= decoder.buffer.length) {
            return [null, go2jsIOEOF()];
        }
        const buffer = decoder.buffer;
        const character = buffer[decoder.position];
        if (character === "," || character === ":") {
            decoder.position++;
            continue;
        }
        if (character === "{" || character === "[" || character === "}" || character === "]") {
            decoder.position++;
            if (character === "{" || character === "[") {
                decoder.stack.push(character);
            } else {
                const opened = decoder.stack.pop();
                if (opened === undefined || go2jsJSONDelimPair(opened) !== character) {
                    return [null, go2jsJSONSyntaxError("unexpected " + character + " in JSON input")];
                }
            }
            return [go2jsJSONDelim(character), null];
        }
        const start = decoder.position;
        const end = go2jsJSONSkipValue(buffer, start);
        if (end <= start) {
            return [null, go2jsJSONSyntaxError("invalid character in JSON input")];
        }
        decoder.position = end;
        const literal = buffer.slice(start, end);
        if (literal[0] === '"') {
            return [JSON.parse(literal), null];
        }
        if (literal === "true" || literal === "false") {
            return [literal === "true", null];
        }
        if (literal === "null") {
            return [null, null];
        }
        if (go2jsJSONNumberLiteral(literal)) {
            const num = Number(literal);
            return [Number.isFinite(num) ? num : literal, null];
        }
        return [null, go2jsJSONSyntaxError("invalid character in JSON input")];
    }
}

function go2jsJSONDelimPair(opened) {
    return opened === "{" ? "}" : "]";
}

function go2jsJSONDelim(character) {
    return {__go2js_json_delim: character};
}

function go2jsJSONDelimText(delim) {
    const character = delim !== null && typeof delim === "object" && delim.__go2js_json_delim !== undefined
        ? delim.__go2js_json_delim
        : "";
    return character === "{" || character === "}" || character === "[" || character === "]" ? character : "";
}

function go2jsJSONDelimCode(delim) {
    return delim !== null && typeof delim === "object" && delim.__go2js_json_delim !== undefined
        ? delim.__go2js_json_delim.charCodeAt(0)
        : 0;
}

function go2jsJSONSyntaxError(message) {
    return go2jsJSONError(message);
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
    this.parts.push(String.fromCodePoint(Number(value)));
    return [go2jsStringByteLength(String.fromCodePoint(Number(value))), null];
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
    this.WriteString(String.fromCodePoint(value));
    return [go2jsStringByteLength(String.fromCodePoint(value)), null];
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

function go2jsIOReadAllResult(n, err) {
	return [null, err];
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

function go2jsCloneValue(value) {
    if (value === null || value === undefined) {
        return value;
    }

	if (Array.isArray(value)) {
		const copy = value.slice();
		for (let i = 0; i < copy.length; i++) {
			copy[i] = go2jsCopy(copy[i]);
		}
		return copy;
	}

    if (value instanceof go2jsNativeMap) {
        return new go2jsNativeMap(value);
    }

    if (typeof value !== "object") {
        return value;
    }

    const embedded = value.__go2js_embedded;
    const cloned = Object.assign(
        Object.create(Object.getPrototypeOf(value)),
        value
    );

    if (Array.isArray(embedded)) {
        for (const name of embedded) {
            const current = value[name];
            if (current === null || current === undefined) {
                continue;
            }

            if (current.__go2js_pointer === true || current.__go2js_interface === true) {
                cloned[name] = current;
            } else {
                cloned[name] = go2jsCloneValue(current);
            }
        }

        return go2jsEmbedProxy(cloned, embedded.slice());
    }

    return cloned;
}

function go2jsMethodValue(receiver, method, copyReceiver) {
	if (receiver === null || receiver === undefined) {
		throw new TypeError("method value on nil receiver");
	}

	if (receiver.__go2js_interface === true) {
		let target = receiver.value;

		if (target === null || target === undefined) {
			throw new TypeError("method value on nil interface");
		}

		if (target.__go2js_pointer === true) {
			target = target[go2jsPointerGet]();

			if (target === null || target === undefined) {
				throw new TypeError("method value on nil pointer");
			}
		}

		const captured = copyReceiver ? go2jsCloneValue(target) : target;
		const fn = captured[method];

		if (typeof fn !== "function") {
			throw new TypeError("method " + method + " is not implemented");
		}

		return (...args) => go2jsCallNow(fn, captured, args);
	}

	let target = receiver;

	if (target.__go2js_pointer === true) {
		target = target[go2jsPointerGet]();

		if (target === null || target === undefined) {
			throw new TypeError("method value on nil pointer");
		}
	}

	const captured = copyReceiver ? go2jsCloneValue(target) : target;
	const fn = captured[method];

	if (typeof fn !== "function") {
		throw new TypeError("method " + method + " is not implemented");
	}

	return (...args) => go2jsCallNow(fn, captured, args);
}

function go2jsMethodExpression(method, copyReceiver) {
	return (receiver, ...args) => {
		if (receiver === null || receiver === undefined) {
			throw new TypeError("method expression on nil receiver");
		}

		if (receiver.__go2js_interface === true) {
			return go2jsMethodValue(receiver, method, copyReceiver)(...args);
		}

		let target = receiver;

		if (copyReceiver && target.__go2js_pointer === true) {
			target = target[go2jsPointerGet]();
		}

		if (copyReceiver) {
			target = go2jsCloneValue(target);
		}

		return go2jsInvokeMethod(target, method, args);
	};
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

function go2jsInvokeMethod(receiver, method, args) {
	if (receiver === null || receiver === undefined) {
		throw new TypeError("method call on nil receiver");
	}

	if (receiver.__go2js_interface === true) {
		return go2jsInterfaceCallNow(receiver, method, ...args);
	}

	if (receiver.__go2js_pointer === true) {
		const target = receiver[go2jsPointerGet]();

		if (target === null || target === undefined) {
			throw new TypeError("method call on nil pointer");
		}
	}

	const fn = receiver[method];

	if (typeof fn !== "function") {
		throw new TypeError("method " + method + " is not implemented");
	}

	return go2jsCallNow(fn, receiver, args);
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

function go2jsStorePtr(ptr, value) {
	if (ptr === null || ptr === undefined) {
		throw new TypeError("invalid pointer assignment");
	}
	if (!go2jsPointerAccessor(ptr, go2jsPointerSet)) {
		throw new TypeError("value is not a pointer");
	}
	ptr[go2jsPointerSet](value);
}

function go2jsToString(value) {
	return String(value);
}

function go2jsToNumber(value) {
	return Number(value);
}

function go2jsToBool(value) {
	return Boolean(value);
}

function go2jsToRune(value) {
	return String.fromCodePoint(value);
}

// go2jsMethodOn calls a pointer receiver method. Go lets the method body see a
// nil receiver, so the call must reach the method instead of failing on a null
// object the way a plain JavaScript property access would.
function go2jsMethodOn(receiver, ctor, name, args) {
	if (receiver === null || receiver === undefined) {
		return go2jsCallNow(ctor.prototype[name], null, args);
	}

	return go2jsCallNow(receiver[name], receiver, args);
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

function go2jsNewTypeOf(value) {
	if (value !== null && value !== undefined && value.__go2js_new_type !== undefined) {
		return value.__go2js_new_type;
	}

	if (value !== null && value !== undefined && value.__go2js_pointer === true) {
		const pointed = value[go2jsPointerGet]();

		return pointed === null || pointed === undefined ? "" : "*" + go2jsNewTypeOf(pointed);
	}

	return "";
}

// go2jsInitializers holds the assignments a package level variable would have
// made while the program was being written. Go makes them after every
// declaration in the package is in place, so an initializer there is allowed to
// name a type that another file declares further down. A JavaScript assignment
// at the top of a file would run before that class exists, so the assignments
// wait here and are made in the order they were written once the program has
// been loaded.
const go2jsInitializers = [];

function go2jsDeferInit(assign) {
	go2jsInitializers.push(function () { return go2jsRunNow(assign()); });
}

function go2jsRunInitializers() {
	while (go2jsInitializers.length > 0) {
		go2jsInitializers.shift()();
	}
}

// go2jsRunSync runs a body written as a generator where a plain call stands,
// which is what the compiler writes in front of main: there is no goroutine of
// its own there to hand the turn to, so the body is run through.
function go2jsRunSync(body) {
	return go2jsRunNow(body());
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

function go2jsRegisterStructFormat(name, fields) {
	go2jsStructFormats[name] = fields;
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

// A pointer that is nil is still a value Go knows the type of once an interface
// holds it, so an interface holding one is not empty. Reading anything through
// it reaches memory the program does not have, which Go stops the program for,
// and a trap that stops it is what stands in for that missing memory.
const go2jsTypedNilPointers = new Map();

function go2jsTypedNilPointer(typeName) {
	let boxed = go2jsTypedNilPointers.get(typeName);

	if (boxed === undefined) {
		const stop = () => {
			throw new TypeError(go2jsRuntimeErrorPrefix + "invalid memory address or nil pointer dereference");
		};

		// The keys the runtime itself asks a value about are let through, since
		// they are the type of the value rather than a field of it, and a name
		// beginning with the runtime's own mark is never a field either.
		const internal = key => {
			return typeof key === "symbol" || key === "constructor" ||
				key === "toString" || key === "valueOf" ||
				(typeof key === "string" && key.startsWith("__go2js"));
		};

		boxed = new Proxy({}, {
			get(target, key) {
				if (key === "__go2js_typed_nil") {
					return true;
				}

				if (key === "__go2js_type_name") {
					return typeName;
				}

				if (internal(key)) {
					return target[key];
				}

				return stop();
			},
			set(target, key, value) {
				if (internal(key)) {
					target[key] = value;

					return true;
				}

				return stop();
			},
			has(target, key) {
				return internal(key);
			},
			ownKeys() {
				return [];
			},
			getOwnPropertyDescriptor() {
				return undefined;
			}
		});

		go2jsTypedNilPointers.set(typeName, boxed);
	}

	return boxed;
}

function go2jsIsTypedNilPointer(value) {
	return value !== null && value !== undefined &&
		value.__go2js_typed_nil === true;
}

// go2jsTypedNilGuard boxes a nil pointer on the way into an interface, and
// leaves a pointer that has something to point at exactly as it is.
function go2jsTypedNilGuard(value, typeName) {
	if (value === null || value === undefined) {
		return go2jsTypedNilPointer(typeName);
	}

	return value;
}

function go2jsTypeOf(value) {
	if (value === null || value === undefined) {
		return "nil";
	}

	if (value.__go2js_typed_nil === true) {
		return value.__go2js_type_name;
	}

	if (value.__go2js_pointer === true) {
		const pointed = value[go2jsPointerGet]();

		return pointed === null || pointed === undefined ? "nil" : "*" + go2jsTypeOf(pointed);
	}

	if (value.__go2js_interface === true) {
		return value.type || "unknown";
	}

	// A value built by the runtime in place of one Go names a type for, such as
	// the error a command leaves behind, says so by that name, and a question
	// about its type is answered with it rather than with the class it happens
	// to be built from.
	if (typeof value.__go2js_type === "string" && value.__go2js_type !== "") {
		return value.__go2js_type;
	}

	// A type of the standard library is a class of the runtime standing for it,
	// and the Go name of that type is registered against the class, so a value of
	// one is asked about by the type it is rather than by the class it was built
	// from.
	if (typeof value === "object") {
		const registered = go2jsTypeNamesByConstructor.get(value.constructor);

		if (typeof registered === "string" && registered !== "") {
			return registered;
		}
	}

	if (value instanceof Error) {
		return "error";
	}

	if (typeof value === "number") {
		return Number.isInteger(value) ? "int" : "float64";
	}

	if (typeof value === "string") {
		return "string";
	}

	if (typeof value === "boolean") {
		return "bool";
	}

	if (Array.isArray(value)) {
		if (value.length === 0) {
			return "[]interface {}";
		}

		return "[]" + go2jsTypeOf(value[0]);
	}

	// A typed array is the one a byte slice is written as, since a program keeps
	// writing bytes into it, so it carries the name of the slice it stands for
	// rather than the name of the array class it is built on.
	const viewName = go2jsViewTypeName(value);

	if (viewName !== null) {
		return viewName;
	}

	if (value instanceof go2jsNativeMap) {
		if (value.__go2js_type !== undefined) {
			return value.__go2js_type;
		}

		if (value.size === 0) {
			return "map[interface {}]interface {}";
		}

		return "map[" + go2jsTypeOf(go2jsFirstKey(value)) + "]" + go2jsTypeOf(go2jsFirstValue(value));
	}

	if (typeof value === "object") {
		const ctor = value.constructor;

		if (ctor && typeof ctor.name === "string" && ctor.name !== "Object") {
			const registered = go2jsLookupTypeName(ctor.name);

			return registered !== undefined ? registered : ctor.name;
		}

		return "struct {}";
	}

	return "interface {}";
}

// go2jsViewTypeName gives back the name of the slice a typed array stands for,
// or nothing when the value is not a typed array at all.
function go2jsViewTypeName(value) {
	if (value === null || value === undefined || typeof value !== "object" ||
		typeof ArrayBuffer === "undefined" || !ArrayBuffer.isView(value)) {
		return null;
	}

	switch (value.constructor !== undefined && value.constructor !== null ? value.constructor.name : "") {
	case "Uint8Array":
	case "Uint8ClampedArray":
		return "[]uint8";
	case "Int8Array":
		return "[]int8";
	case "Uint16Array":
		return "[]uint16";
	case "Int16Array":
		return "[]int16";
	case "Uint32Array":
		return "[]uint32";
	case "Int32Array":
		return "[]int32";
	case "Float32Array":
		return "[]float32";
	case "Float64Array":
		return "[]float64";
	default:
		return null;
	}
}

function go2jsShortTypeName(name) {
	name = go2jsNormalizeTypeName(name);

	const pointer = name.startsWith("*");
	const trimmed = pointer ? name.slice(1) : name;
	const parts = trimmed.split(".");
	const short = parts[parts.length - 1];

	return pointer ? "*" + short : short;
}

function go2jsSwitchTypeOf(value) {
	return go2jsShortTypeName(go2jsTypeOf(value));
}

// go2jsSwitchCaseIndex reports which of the labels a type switch names is the
// one the examined value matches, in the order they were written, because the
// first clause that matches is the clause that runs. A label that names an
// interface matches every value carrying its methods, not only a value whose
// own type is that interface, and a label that names the empty interface
// matches anything but a nil.
function go2jsSwitchCaseIndex(value, labels) {
	const name = go2jsSwitchTypeOf(value);

	for (let index = 0; index < labels.length; index++) {
		const label = labels[index];

		if (label === "nil") {
			if (name === "nil") {
				return index;
			}

			continue;
		}

		if (label === "interface{}" || label === "interface {}" || label === "any") {
			if (name !== "nil") {
				return index;
			}

			continue;
		}

		if (name === label || go2jsSatisfiesInterface(value, label)) {
			return index;
		}
	}

	return -1;
}

function go2jsSameTypeName(actual, expected) {
	if (actual === expected) {
		return true;
	}

	if (typeof actual !== "string" || typeof expected !== "string") {
		return false;
	}

	return go2jsShortTypeName(actual) === go2jsShortTypeName(expected);
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

// go2jsRegisterInterface records the method set of an interface type so a type
// assertion can check method satisfaction instead of only exact type names.
function go2jsRegisterInterface(name, methods) {
	go2jsInterfaces[go2jsInterfaceKey(name)] = methods;
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

function go2jsSatisfiesInterface(value, name) {
	if (value === null || value === undefined) {
		return false;
	}

	const methods = go2jsInterfaceMethods(name);

	if (methods === undefined) {
		return false;
	}

	for (const method of methods) {
		if (go2jsLookupMember(value, method) === undefined) {
			return false;
		}
	}

	return true;
}

function go2jsAssert(value, typeName) {
	if (value !== null && value !== undefined &&
		value.__go2js_interface === true) {
		if (go2jsSameTypeName(value.type, typeName)) {
			return value.value;
		}

		if (go2jsSatisfiesInterface(value, typeName)) {
			return value;
		}

		throw new TypeError(
			"interface conversion: " + value.type + " is not " + typeName
		);
	}

	const actual = go2jsTypeOf(value);

	if (go2jsSameTypeName(actual, typeName)) {
		return value;
	}

	if (go2jsSatisfiesInterface(value, typeName)) {
		return value;
	}

	throw new TypeError("interface conversion failed");
}

function go2jsAssertOK(value, typeName) {
    if (value !== null && value !== undefined &&
        value.__go2js_interface === true &&
        go2jsSameTypeName(value.type, typeName)) {
        return [value.value, true];
    }

    if (value !== null && value !== undefined && go2jsSameTypeName(go2jsTypeOf(value), typeName)) {
        return [value, true];
    }

    if (go2jsSatisfiesInterface(value, typeName)) {
		return [value, true];
	}

    return [null, false];
}

// go2jsMin and go2jsMax answer with the smallest and the largest of the values
// they are given, which is the same answer whichever of the two is found first
// only when there is nothing to tell them apart. An operand that is not a
// number at all, such as a NaN, decides the answer on its own: a NaN among the
// values gives a NaN, and the two zeroes are told apart by the one written
// with a minus in front, which is the smaller of them.
function go2jsMin(...values) {
	let best = values[0];

	for (let index = 1; index < values.length; index++) {
		const order = go2jsMinOrder(best, values[index]);

		if (Number.isNaN(order)) {
			return NaN;
		}

		if (order > 0) {
			best = values[index];
		}
	}

	return best;
}

function go2jsMax(...values) {
	let best = values[0];

	for (let index = 1; index < values.length; index++) {
		const order = go2jsMinOrder(best, values[index]);

		if (Number.isNaN(order)) {
			return NaN;
		}

		if (order < 0) {
			best = values[index];
		}
	}

	return best;
}

// go2jsMinOrder puts two values in order. A value that is not a plain number
// is asked for the one it stands for, which is what a duration and a moment
// each answer as, and the answer is one of the values as it was given rather
// than the number behind it, so a duration stays a duration.
function go2jsMinOrder(left, right) {
	if (typeof left === "string" || typeof right === "string") {
		return left < right ? -1 : (left > right ? 1 : 0);
	}

	const a = Number(left);
	const b = Number(right);

	if (Number.isNaN(a) || Number.isNaN(b)) {
		return NaN;
	}

	if (a === 0 && b === 0) {
		return Object.is(a, -0) || Object.is(b, -0) ? -1 : 1;
	}

	return a < b ? -1 : (a > b ? 1 : 0);
}

// go2jsClear empties what it is given: a map is left with nothing in it at all,
// while a slice keeps its length and is left holding nothing but zero values,
// because a slice of a length is still that slice afterwards.
function go2jsClear(value) {
	const target = go2jsUnwrap(value);

	if (target === null || target === undefined) {
		return null;
	}

	if (target instanceof go2jsNativeMap) {
		target.clear();

		return null;
	}

	if (Array.isArray(target)) {
		target.fill(0);

		return null;
	}

	if (target instanceof go2jsNativeSet) {
		target.clear();

		return null;
	}

	for (const key of Object.keys(target)) {
		delete target[key];
	}

	return null;
}

function go2jsComplex(re, im) {
	return { re: Number(re), im: Number(im) };
}

function go2jsComplexValue(value) {
	if (value !== null && typeof value === "object" && "re" in value && "im" in value) {
		return value;
	}
	return { re: Number(value), im: 0 };
}

function go2jsComplexConvert(value) {
	return go2jsComplexValue(value);
}

function go2jsComplexAdd(a, b) {
	a = go2jsComplexValue(a);
	b = go2jsComplexValue(b);
	return { re: a.re + b.re, im: a.im + b.im };
}

function go2jsComplexSub(a, b) {
	a = go2jsComplexValue(a);
	b = go2jsComplexValue(b);
	return { re: a.re - b.re, im: a.im - b.im };
}

function go2jsComplexMul(a, b) {
	a = go2jsComplexValue(a);
	b = go2jsComplexValue(b);
	return {
		re: a.re * b.re - a.im * b.im,
		im: a.re * b.im + a.im * b.re
	};
}

function go2jsComplexDiv(a, b) {
	a = go2jsComplexValue(a);
	b = go2jsComplexValue(b);
	const denominator = b.re * b.re + b.im * b.im;
	return {
		re: (a.re * b.re + a.im * b.im) / denominator,
		im: (a.im * b.re - a.re * b.im) / denominator
	};
}

function go2jsComplexNeg(value) {
	value = go2jsComplexValue(value);
	return { re: -value.re, im: -value.im };
}

function go2jsComplexEqual(a, b) {
	a = go2jsComplexValue(a);
	b = go2jsComplexValue(b);
	return a.re === b.re && a.im === b.im;
}

function go2jsReal(value) {
	return go2jsComplexValue(value).re;
}

function go2jsImag(value) {
	return go2jsComplexValue(value).im;
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

function go2jsCap(value) {
	if (value === null || value === undefined) {
		return 0;
	}
	if (value.__go2js_pointer === true) {
		return go2jsCap(value[go2jsPointerGet]());
	}
	if (value instanceof go2jsNativeMap || value instanceof go2jsNativeSet) {
		return value.size;
	}
	if (Array.isArray(value)) {
		return value.length;
	}
	if (typeof value === "string") {
		return go2jsStringByteLength(value);
	}
	return 0;
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

function go2jsMake(type, size, capacity) {
	if (typeof size === "number") {
		if (typeof capacity === "number" && capacity > size) {
			const value = new Array(capacity);
			value.length = size;
			return value;
		}
		return new Array(size);
	}
	return [];
}

function go2jsMakeSlice(size, capacity) {
		if (typeof size !== "number") {
			return [];
		}

		const actualCapacity = typeof capacity === "number" ? capacity : size;
		const value = new Array(size);

		Object.defineProperty(value, "__go2js_cap", {
			value: actualCapacity,
			writable: true,
			configurable: true
		});

		return value;
}

function go2jsMakeMap() {
	return new go2jsNativeMap();
}

function go2jsBinaryBigEndian() {
    return "be";
}

function go2jsBinaryLittleEndian() {
    return "le";
}

function go2jsBinaryPutUint(order, width, target, value) {
    let number = Number(value);

    if (!Number.isFinite(number) || number < 0) {
        number = 0;
    }

    number = Math.trunc(number) % Math.pow(2, width);

    if (number < 0) {
        number += Math.pow(2, width);
    }

    const bytes = new Array(width);

    for (let index = width - 1; index >= 0; index--) {
        bytes[index] = number % 256;
        number = Math.floor(number / 256);
    }

    if (order === "le") {
        bytes.reverse();
    }

    for (let index = 0; index < width; index++) {
        target[index] = bytes[index] & 255;
    }
}

function go2jsBinaryUint(order, width, source) {
    let result = 0;

    if (order === "le") {
        for (let index = width - 1; index >= 0; index--) {
            result = result * 256 + (Number(source[index]) & 255);
        }

        return result;
    }

    for (let index = 0; index < width; index++) {
        result = result * 256 + (Number(source[index]) & 255);
    }

    return result;
}

function go2jsBinaryUvarint(buffer) {
    let value = 0;
    let shift = 0;

    for (let index = 0; index < buffer.length; index++) {
        const current = Number(buffer[index]) & 255;

        if (current < 0x80) {
            return [value + current * Math.pow(2, shift), index + 1];
        }

        value += (current & 0x7f) * Math.pow(2, shift);
        shift += 7;
    }

    return [value, 0];
}

function go2jsBinaryVarint(buffer) {
    const [value, read] = go2jsBinaryUvarint(buffer);

    if (read === 0) {
        return [0, 0];
    }

    return value % 2 === 0 ? [-(value / 2), read] : [(value + 1) / 2, read];
}

function go2jsBinaryPutUvarint(buffer, value) {
    let remaining = Math.trunc(Number(value));

    if (!Number.isFinite(remaining) || remaining < 0) {
        remaining = 0;
    }

    for (let index = 0; index < buffer.length; index++) {
        if (remaining < 0x80) {
            buffer[index] = remaining & 255;
            return index + 1;
        }

        buffer[index] = (remaining & 0x7f) | 0x80;
        remaining = Math.floor(remaining / 128);
    }

    return 0;
}

function go2jsBinaryPutVarint(buffer, value) {
    const number = Math.trunc(Number(value));

    return go2jsBinaryPutUvarint(buffer, number < 0 ? -2 * number : 2 * number);
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

function go2jsMapGetOK(map, key, zero) {
	if (!(map instanceof go2jsNativeMap) || !map.has(go2jsMapKey(key))) {
		return [zero, false];
	}

	return [map.get(go2jsMapKey(key)), true];
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

// go2jsMapUpdate works an operation out on the entry a compound assignment is
// aimed at, because a lookup is a value JavaScript will not let an assignment
// be written to. A key that is not there yet reads as the element type's zero
// value, and the result is written back under the same key.
function go2jsMapUpdate(map, key, zero, operation) {
	if (map === null || map === undefined || map.__go2js_nil === true) {
		go2jsPanic("assignment to entry in nil map");
	}

	if (!(map instanceof go2jsNativeMap)) {
		throw new TypeError("go2jsMapUpdate expects a Map");
	}

	const k = go2jsMapKey(key);
	const current = map.has(k) ? map.get(k) : zero;
	map.set(k, operation(current));
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

function go2jsMapValues(map) {
	if (!(map instanceof go2jsNativeMap)) {
		return [];
	}
	return Array.from(map.values());
}

function go2jsMapEntries(map) {
	if (!(map instanceof go2jsNativeMap)) {
		return [];
	}
	return Array.from(map.entries());
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



function go2jsDelete(value, key) {
	if (value instanceof go2jsNativeMap) {
		value.delete(key);
		return;
	}

	if (value && typeof value === "object") {
		delete value[key];
	}
}

function go2jsContains(value, item) {
	if (value instanceof go2jsNativeMap || value instanceof go2jsNativeSet) {
		return value.has(item);
	}

	if (Array.isArray(value) || typeof value === "string") {
		return value.includes(item);
	}

	return false;
}

function go2jsCloneArray(value) {
	if (!Array.isArray(value)) {
		return [];
	}
	return value.slice();
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

function go2jsRange(value) {
	if (value instanceof go2jsNativeMap || value instanceof go2jsNativeSet) {
		return value.entries();
	}

	if (Array.isArray(value) || typeof value === "string") {
		return value.entries();
	}

	return Object.entries(value || {});
}

// go2jsIndex reads one element of a slice or an array the way Go does, which
// means refusing an index that is not there rather than answering undefined.
// go2jsIndexPtr is the address of one element of a slice or an array. The
// element and the index it was reached by are both read where the address is
// taken, because that is where Go reads them, and a program that holds the
// address afterwards reads and writes that element rather than whichever one
// sits at the index later on.
function go2jsIndexPtr(owner, index, typeName) {
	const array = go2jsMaterializeValue(go2jsUnwrap(owner));
	const at = Math.trunc(Number(index));

	return go2jsPtr(
		function() {
			return go2jsIndex(array, at);
		},
		function(next) {
			go2jsIndexSet(array, at, next);
		},
		typeName
	);
}

// go2jsIndexSet writes one element of a slice or an array, which is the other
// half of reading one by go2jsIndex.
function go2jsIndexSet(value, index, next) {
	const position = Math.trunc(Number(index));

	if (position < 0) {
		throw new go2jsNativeRangeError(go2jsRuntimeErrorPrefix + "index out of range [" + position + "]");
	}

	if (position >= go2jsLen(value)) {
		throw new go2jsNativeRangeError(go2jsRuntimeErrorPrefix + "index out of range [" + position + "] with length " + go2jsLen(value));
	}

	if (value instanceof Uint8Array) {
		value[position] = Number(next) & 0xff;
		return;
	}

	value[position] = next;
}

function go2jsIndex(value, index) {
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

// go2jsIndexCheck guards an assignment to a slice or an array element, where
// the check has to stand beside the assignment rather than inside it.
function go2jsIndexCheck(value, index) {
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

function go2jsWideSub(left, right, typeName) {
	return go2jsWideBinary(left, right, (a, b) => a - b, typeName);
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

function go2jsWideRem(left, right, typeName) {
	return go2jsWideBinary(left, right, (a, b) => {
		if (b === 0) {
			throw new go2jsNativeRangeError(go2jsRuntimeErrorPrefix + "integer divide by zero");
		}

		return a % b;
	}, typeName);
}

function go2jsWideAnd(left, right, typeName) {
	return go2jsWideBinary(left, right, (a, b) => a & b, typeName);
}

function go2jsWideOr(left, right, typeName) {
	return go2jsWideBinary(left, right, (a, b) => a | b, typeName);
}

function go2jsWideXor(left, right, typeName) {
	return go2jsWideBinary(left, right, (a, b) => a ^ b, typeName);
}

function go2jsWideAndNot(left, right, typeName) {
	left = go2jsDurationOperand(left);
	right = go2jsDurationOperand(right);

	// Turning a number inside out in JavaScript is a thirty-two bit thing, so
	// marks above the thirty-second are cleared through whole numbers rather
	// than through an operator that would lose every one of them.
	if (go2jsIsWide(left) || go2jsIsWide(right) ||
		go2jsFitsThirtyTwoBits(left) === false || go2jsFitsThirtyTwoBits(right) === false) {
		return go2jsWideBinary(go2jsWide(left), go2jsWide(right), (a, b) => a & ~b, typeName);
	}

	// The marks to clear are a subset of the marks held, so taking them off by
	// what they are worth clears them without borrowing from the ones above.
	return go2jsWideWrap(left - (left & right), typeName);
}

// go2jsFitsThirtyTwoBits reports whether a whole number is one a thirty-two bit
// operation reaches every mark of.
function go2jsFitsThirtyTwoBits(value) {
	const number = Number(value);

	return Number.isInteger(number) && number >= -2147483648 && number <= 4294967295;
}

// go2jsWideFloat turns a whole number into the nearest number a double holds,
// which is the same number the conversion gives in Go, digits and all the way
// out. A double carries about sixteen digits exactly, so past that the answer
// is the nearest one to it rather than the whole number asked for.
function go2jsWideFloat(value) {
	value = go2jsDurationOperand(value);

	// reading a whole number as a double gives the nearest double to it, which is
	// the number the conversion gives in Go, digits and all the way out. A double
	// carries about sixteen digits exactly, so past that the answer is the nearest
	// to the whole number rather than the whole number itself, and that is what a
	// program asking for a float64 is asking for.
	return Number(value);
}


// go2jsWideToSigned turns a whole number into a signed number of the width the
// type asks for, keeping every digit it has when they all fit.
function go2jsWideToSigned(value, typeName) {
	value = go2jsDurationOperand(value);

	const number = typeof value === "bigint" ? value : go2jsWide(value);

	return go2jsNarrow(go2jsIntWrap(number, typeName || "int64"));
}

// go2jsWideToUnsigned is the same for an unsigned type, where a number below the
// smallest is not a fault but the largest the type has.
function go2jsWideToUnsigned(value, typeName) {
	value = go2jsDurationOperand(value);

	const number = typeof value === "bigint" ? value : go2jsWide(value);
	const wrapped = go2jsIntWrap(number, typeName || "uint64");

	return go2jsNarrow(typeof wrapped === "bigint" ? go2jsWideWrap(wrapped, typeName || "uint64") : wrapped);
}

// go2jsWideCompare answers which of two whole numbers comes first, whichever
// kind either of them is. A double cannot hold every whole number, so a number
// past the safe range is compared as digits rather than as a number, which is
// the same answer and the right one.
function go2jsWideCompare(left, right) {
	left = go2jsDurationOperand(left);
	right = go2jsDurationOperand(right);

	if (typeof left === "bigint" && typeof right === "bigint") {
		return left < right ? -1 : (left > right ? 1 : 0);
	}

	// one side is a double and the other a whole number, so the double is read
	// as digits only when it stands outside the range a double holds exactly
	if (typeof left === "number" && typeof right === "bigint") {
		if (Number.isInteger(left) && left >= -go2jsSafeInteger && left <= go2jsSafeInteger) {
			return left < Number(right) ? -1 : (left > Number(right) ? 1 : 0);
		}

		return left < Number(right) ? -1 : (left > Number(right) ? 1 : 0);
	}

	if (typeof left === "bigint" && typeof right === "number") {
		return Number(left) < right ? -1 : (Number(left) > right ? 1 : 0);
	}

	return left < right ? -1 : (left > right ? 1 : 0);
}

// go2jsWideEqual answers whether two whole numbers are the same number, in
// either kind and in a mixture of the two. A double past the range a double
// holds exactly is not equal to a whole number that stands for the same digits,
// since the double is a different number once the digits are gone.
function go2jsWideEqual(left, right) {
	left = go2jsDurationOperand(left);
	right = go2jsDurationOperand(right);

	if (typeof left === "bigint" || typeof right === "bigint") {
		return go2jsWideCompare(left, right) === 0;
	}

	return left === right;
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

// go2jsShl shifts a number the way Go does, where the count is however many
// bits a program asks for. A JavaScript shift is thirty-two bits wide whatever
// the count says, so a shift of sixty-two bits is a shift of thirty in
// JavaScript and the wrong answer twice over.
// A shift lands inside the sixty four bits the number is held in, the same way
// any other operation on it does, and the name of the type it is for says how
// wide those bits are. A number with no type of its own to land in is left
// wherever it lands.
function go2jsShl(left, right, typeName) {
	left = go2jsDurationOperand(left);
	right = go2jsDurationOperand(right);

	const count = Math.trunc(Number(right));

	if (count < 0) {
		return go2jsShr(left, -count, typeName);
	}

	// a shift that lands outside the range a double holds exactly is worked out
	// as digits, since a double is already carrying more of them than it can
	// keep by the time a shift of so many bits has been asked for
	if (count > 30 && (go2jsIsWide(left) || Math.abs(Math.trunc(Number(left))) > go2jsSafeInteger / 2 ** count)) {
		return go2jsShlWide(go2jsWide(left) << BigInt(count), typeName);
	}

	return go2jsShlWide(Math.trunc(Number(left)) * 2 ** count, typeName);
}

function go2jsShlWide(value, typeName) {
	return typeName === undefined ? go2jsNarrow(value) : go2jsWideWrap(value, typeName);
}

// go2jsShr brings a number down by so many bits. A shift to the right on a
// signed number brings the sign down with it and one on an unsigned number
// brings down zeros, and the number itself says which it is: a negative one
// keeps its sign the whole way down.
function go2jsShr(left, right, typeName) {
	left = go2jsDurationOperand(left);
	right = go2jsDurationOperand(right);

	const count = Math.trunc(Number(right));

	if (count <= 0) {
		return go2jsShlWide(go2jsIsWide(left) ? go2jsWide(left) : Math.trunc(Number(left)), typeName);
	}

	if (go2jsIsWide(left)) {
		return go2jsShlWide(go2jsWide(left) >> BigInt(count), typeName);
	}

	// a number that is a whole number of digits is brought down by dropping the
	// digits on the right, and one that is not whole comes down by rounding
	// towards the floor, which keeps the ones the digits stand for. Rounding
	// towards zero instead would turn -1 into 0 for a shift of one, which is not
	// what a shift does to a negative number in Go.
	return go2jsShlWide(Math.floor(Math.trunc(Number(left)) / 2 ** count), typeName);
}

// go2jsIntWrap keeps a whole number inside the width of the type it is for,
// which is what a Go variable of that type does without being asked. A number
// too wide for its type is not an error in Go, it is the number that fits, so
// an int8 one past its largest is the smallest int8 and a uint8 one below its
// smallest is the largest uint8.
// go2jsIntWidthName gives back the name of a whole number as a width rather than
// as a name of its own, since a name a package gives a whole number says what
// the number is and the width it is held in is the one underneath it.
function go2jsIntWidthName(typeName) {
	switch (typeName) {
		case "int8":
		case "int16":
		case "int32":
		case "int64":
		case "int":
		case "uint8":
		case "uint16":
		case "uint32":
		case "uint64":
		case "uint":
		case "uintptr":
		case "byte":
		case "rune":
			return typeName;
	}

	const kind = go2jsTypeKinds[typeName];

	return typeof kind === "string" ? kind : typeName;
}

function go2jsIntWrap(value, typeName) {
	// A file mode is a set of marks rather than a number, so marks put on it or
	// taken off it leave it a mode rather than a plain number.
	if (typeName === "fs.FileMode" || typeName === "os.FileMode" || typeName === "io/fs.FileMode" || typeName === "fs.FileMode " || typeName === "io/fs.FileMode ") {
		return go2jsFileMode(go2jsFileModeBits(value));
	}

	value = go2jsDurationOperand(value);

	// A type named over a whole number is held as the number it stands for
	// rather than as the name it was given, since a name says what a value is
	// and not how wide it is.
	typeName = go2jsIntWidthName(typeName);

	// a whole number with more digits than a double keeps is given the width of
	// its type as digits, since the digits a double is carrying past that point
	// are not the ones the number was given
	if (typeof value === "bigint") {
		return go2jsNarrow(go2jsWideWrap(value, typeName));
	}

	const number = Math.trunc(Number(value));
	let bits = 0;
	let signed = true;

	switch (typeName) {
		case "int8":
			bits = 8;
			break;
		case "int16":
			bits = 16;
			break;
		case "int32":
		case "rune":
			bits = 32;
			break;
		case "int64":
		case "int":
			bits = 64;
			break;
		case "uint8":
		case "byte":
			bits = 8;
			signed = false;
			break;
		case "uint16":
			bits = 16;
			signed = false;
			break;
		case "uint32":
			bits = 32;
			signed = false;
			break;
		case "uint64":
		case "uint":
		case "uintptr":
			bits = 64;
			signed = false;
			break;
		default:
			return number;
	}

	const top = 2 ** bits;
	const kept = ((number % top) + top) % top;

	return signed && kept > top / 2 - 1 ? kept - top : kept;
}

// go2jsMod takes the remainder the way Go does, which is the one left over
// after a quotient cut off toward zero, rather than the JavaScript remainder
// that keeps the sign of the dividend.
function go2jsMod(left, right) {
	left = go2jsDurationOperand(left);
	right = go2jsDurationOperand(right);

	if (go2jsIsWide(left) || go2jsIsWide(right)) {
		return go2jsWideRem(left, right);
	}

	if (right === 0) {
		throw new go2jsNativeRangeError(go2jsRuntimeErrorPrefix + "integer divide by zero");
	}

	return left - Math.trunc(left / right) * right;
}

function go2jsStringByteAt(value, index) {
	if (typeof value !== "string") {
		return value[Math.trunc(index)];
	}

	return value.charCodeAt(Math.trunc(index));
}

function go2jsZeroArray(length, zeroFactory) {
	const out = new Array(length);
	const factory = typeof zeroFactory === "function" ? zeroFactory : () => 0;

	for (let index = 0; index < length; index++) {
		out[index] = factory();
	}

	Object.defineProperty(out, "__go2js_cap", {
		value: length,
		writable: true,
		configurable: true
	});

	go2jsArrayMark.add(out);
	return out;
}

function go2jsZeroValue(type) {
	switch (type) {
	case "string":
		return "";
	case "bool":
		return false;
	case "number":
		return 0;
	default:
		return null;
	}
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

function go2jsPrint(...values) {
	go2jsWriteOut(process.stdout, go2jsJoinOperands(values, false));
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
function go2jsWriteDestination(destination, text) {
	const target = go2jsUnwrap(destination);

	if (target === null || target === undefined) {
		throw new Error("io: writer is nil");
	}

	if (typeof target.write === "function") {
		target.write(text);
		return [text.length, null];
	}

	if (typeof target.Write === "function") {
		const written = go2jsCallNow(target.Write, target, [go2jsStringToBytes(text)]);
		return Array.isArray(written) ? written : [written, null];
	}

	throw new Error("io: writer does not implement Write");
}

function go2jsSprint(...values) {
	return go2jsJoinOperands(values, false);
}

function go2jsSprintln(...values) {
	return go2jsJoinOperands(values, true) + "\n";
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

function go2jsFprintf(writer, format, args) {
	return go2jsFprint(writer, [go2jsSprintf(format, ...args)], "", "");
}

function go2jsFprintln(writer, values) {
	return go2jsFprint(writer, values, "\n", " ");
}

// go2jsSpreadArgs marks a variadic slice so go2jsSprintf can splice its elements
// into the operand list, which is what a trailing ... means in Go.
function go2jsSpreadArgs(slice) {
	return {__go2js_spread: true, slice: slice};
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

// go2jsSprintfWrapping writes a format that is allowed to wrap, which is the
// shape Errorf gives its format. fmt lets a %w wrap there, and only for an
// operand that is an error.
function go2jsSprintfWrapping(format, args) {
	return go2jsFormatForPrinter(format, args, true);
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
function go2jsNilInterface(value, typeName, shape) {
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

	return go2jsInterface(value, typeName, typeName);
}

function go2jsIsNil(value) {
	if (value === null || value === undefined) {
		return true;
	}

	return value.__go2js_nil === true || value.__go2js_typed_nil === true;
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
			const special = go2jsSpecialFloatText(num);

			if (special !== "") {
				return go2jsPad(special, parsed, false);
			}

			return numberText(num, go2jsFormatFixed(Math.abs(num), precision === null ? 6 : precision));
		}
		case "e":
		case "E": {
			const num = Number(value);
			const special = go2jsSpecialFloatText(num);

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
			return go2jsFormatG(Number(value), precision, parsed);
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
			return go2jsPad(String.fromCharCode(value), parsed, false);
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

// go2jsFormatFixed writes a number in positional notation with a fixed number of
// fraction digits. Above 1e21 toFixed falls back to an exponent, so the value is
// written out from the shortest form that reads back as the same float, which is
// what Go prints before it pads the fraction.
function go2jsFormatFixed(abs, precision) {
	if (abs < 1e21) {
		return abs.toFixed(precision);
	}

	const { rawExponent, mantissa } = go2jsFloatBits(abs);
	const significand = rawExponent === 0 ? mantissa : mantissa | (1n << 52n);
	const scale = (rawExponent === 0 ? 1 : rawExponent) - 1023 - 52;
	const shift = scale < 0 ? BigInt(-scale) : 0n;
	const units = 10n ** BigInt(precision);

	// The value is the significand scaled by a power of two, so a value with a
	// negative scale has digits left over after the point.
	let whole = shift === 0n ? significand << BigInt(scale) : significand >> shift;
	let below = shift === 0n ? 0n : (significand & ((1n << shift) - 1n)) * units;
	const guard = shift === 0n ? 0n : 1n << shift;
	let rounded = below >> shift;

	// The digits that do not fit are rounded off, half away from zero, and a
	// round that carries past the point moves the whole number up.
	if (shift > 0n && below - (rounded << shift) >= guard / 2n) {
		rounded += 1n;
	}

	if (rounded >= units) {
		rounded -= units;
		whole += 1n;
	}

	return precision === 0
		? whole.toString()
		: whole.toString() + "." + rounded.toString().padStart(precision, "0");
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

// go2jsFloat32 gives a result back the number of digits a float32 keeps, since a
// float64 holds digits a float32 has no room for and a number carrying digits it
// has no room for is a number that is not the one the program asked for.
function go2jsFloat32(value) {
	return Math.fround(value);
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

	const [mantissa, exponent] = Math.abs(value).toExponential(precision).split("e");
	const exp = parseInt(exponent, 10);
	const expSign = exp < 0 ? "-" : "+";
	const expDigits = String(Math.abs(exp)).padStart(2, "0");
	const body = mantissa + "e" + expSign + expDigits;
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

// go2jsFormatG renders %g, using the shortest representation when no precision
// is given and switching to the exponent form outside Go's -4..eprec window.
function go2jsFormatG(value, precision, parsed) {
	if (value === 0) {
		return go2jsPadNumber("0", "", "0", parsed);
	}

	const parts = go2jsDecimalParts(value);

	if (parts === null) {
		return go2jsPad(String(value), parsed, false);
	}

	const exp = parts.dp - 1;
	// Go uses the exponent form when exp < -4 or exp >= eprec, and picks
	// eprec = 6 whenever the shortest representation was requested.
	const eprec = precision === null ? 6 : precision;

	if (exp < -4 || exp >= eprec) {
		const mantissaDigits = precision === null ? parts.digits.length - 1 : precision - 1;

		return go2jsFormatE(value, Math.max(mantissaDigits, 0), parsed);
	}

	const body = go2jsFixedFromParts(parts, precision === null ? parts.digits.length : precision);
	const prefix = value < 0 ? "-" : parsed.flags.includes("+") ? "+" : parsed.flags.includes(" ") ? " " : "";

	return go2jsPadNumber(prefix + body, prefix, body, parsed);
}

// go2jsFixedFromParts renders a decimal point at parts.dp using exactly
// significant digits.
function go2jsFixedFromParts(parts, significant) {
	const digits = parts.digits.padEnd(significant, "0");

	if (parts.dp <= 0) {
		return "0." + "0".repeat(-parts.dp) + digits;
	}

	if (parts.dp >= significant) {
		return digits + "0".repeat(parts.dp - significant);
	}

	return digits.slice(0, parts.dp) + "." + digits.slice(parts.dp);
}

function go2jsStringsContains(s, substr) {
	return s.includes(substr);
}

function go2jsStringsHasPrefix(s, prefix) {
	return s.startsWith(prefix);
}

function go2jsStringsHasSuffix(s, suffix) {
	return s.endsWith(suffix);
}

function go2jsStringsUTF8Index(text, utf16Index) {
	if (utf16Index < 0) {
		return utf16Index;
	}

	let bytes = 0;

	for (let offset = 0; offset < utf16Index; offset++) {
		const code = text.codePointAt(offset);

		if (code < 0x80) {
			bytes += 1;
		} else if (code < 0x800) {
			bytes += 2;
		} else {
			bytes += code < 0x10000 ? 3 : 4;
		}

		if (code >= 0x10000) {
			offset++;
		}
	}

	return bytes;
}

function go2jsStringsIndex(s, substr) {
	return go2jsStringsUTF8Index(s, s.indexOf(substr));
}

function go2jsStringsToUpper(s) {
	return s.toUpperCase();
}

function go2jsStringsToLower(s) {
	return s.toLowerCase();
}

function go2jsStringsTrimSpace(s) {
	return s.trim();
}

function go2jsStringsTrim(s, cutset) {
	const chars = cutset.split("");
	let start = 0;
	let end = s.length;

	while (start < end && chars.includes(s[start])) {
		start++;
	}

	while (end > start && chars.includes(s[end - 1])) {
		end--;
	}

	return s.slice(start, end);
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

function go2jsStringsReplace(s, old, replacement, count) {
	if (count < 0) {
		return s.split(old).join(replacement);
	}

	let result = s;
	for (let i = 0; i < count; i++) {
		result = result.replace(old, replacement);
	}
	return result;
}

function go2jsStringsReplaceAll(s, old, replacement) {
	return s.split(old).join(replacement);
}

function go2jsStringsRepeat(s, count) {
	return s.repeat(count);
}

function go2jsStringsFields(s) {
	return s.split(/\s+/).filter((part) => part.length > 0);
}

function go2jsStringsTrimPrefix(s, prefix) {
	if (s.startsWith(prefix)) {
		return s.slice(prefix.length);
	}
	return s;
}

function go2jsStringsTrimSuffix(s, suffix) {
	if (s.endsWith(suffix)) {
		return s.slice(0, -suffix.length);
	}
	return s;
}

function go2jsStringsCut(s, sep) {
	const index = s.indexOf(sep);
	if (index < 0) {
		return [s, "", false];
	}
	return [s.slice(0, index), s.slice(index + sep.length), true];
}

function go2jsStringsCutPrefix(s, prefix) {
	if (!s.startsWith(prefix)) {
		return [s, false];
	}
	return [s.slice(prefix.length), true];
}

function go2jsStringsCutSuffix(s, suffix) {
	if (!s.endsWith(suffix)) {
		return [s, false];
	}
	return [s.slice(0, -suffix.length), true];
}

function go2jsStringsEqualFold(a, b) {
	return a.toLowerCase() === b.toLowerCase();
}

function go2jsStringsCompare(a, b) {
	if (a < b) {
		return -1;
	}
	if (a > b) {
		return 1;
	}
	return 0;
}

function go2jsStringsToValidUTF8(s) {
	return s;
}

function go2jsStringsLastIndex(s, substr) {
	return go2jsStringsUTF8Index(s, s.lastIndexOf(substr));
}

function go2jsMathInf(sign) {
	return sign >= 0 ? Infinity : -Infinity;
}

function go2jsMathNaN() {
	return NaN;
}

function go2jsHexEncode(dst, src) {
	const target = go2jsUnwrap(dst);
	const bytes = go2jsHexBytes(src);

	if (go2jsLen(target) < bytes.length * 2) {
		throw new Error("encoding/hex: buffer too small");
	}

	const text = go2jsHexEncodeToString(bytes);

	for (let index = 0; index < text.length; index++) {
		target[index] = text.charCodeAt(index);
	}

	return text.length;
}

function go2jsHexEncodeToString(src) {
	const bytes = go2jsHexBytes(src);
	let out = "";

	for (const b of bytes) {
		out += b.toString(16).padStart(2, "0");
	}

	return out;
}

function go2jsHexDecodeString(s) {
	const out = [];

	for (let i = 0; i + 1 < s.length; i += 2) {
		const byte = parseInt(s.slice(i, i + 2), 16);

		if (Number.isNaN(byte)) {
			throw new Error("encoding/hex: invalid byte: " + s.slice(i, i + 2));
		}

		out.push(byte);
	}

	return out;
}

function go2jsHexEncodedLen(n) {
	return n * 2;
}

function go2jsHexDecodedLen(s) {
	return s.length >> 1;
}

function go2jsHexDump(src) {
	const bytes = go2jsHexBytes(src);
	let out = "";

	for (let base = 0; base < bytes.length; base += 16) {
		const chunk = bytes.slice(base, base + 16);
		const left = chunk.slice(0, 8);
		const right = chunk.slice(8);

		out += base.toString(16).padStart(8, "0");
		out += "  ";
		out += go2jsHexColumns(left);
		out += "  ";
		out += go2jsHexColumns(right);
		out += "  |";
		out += chunk.map(b => (b >= 0x20 && b < 0x7f) ? String.fromCharCode(b) : ".").join("");
		out += "|\n";
	}

	return out;
}

function go2jsHexColumns(chunk) {
	return chunk
		.map(b => b.toString(16).padStart(2, "0"))
		.join(" ")
		.padEnd(23, " ");
}

function go2jsHexDumper(dst) {
	const target = go2jsHexDestination(dst);

	return {
		type: "*hex.Dumper",
		Write(p) {
			const bytes = go2jsHexBytes(p);
			const text = go2jsHexDump(bytes);

			if (target !== null && target !== undefined && typeof target.Write === "function") {
				go2jsCallNow(target.Write, target, [text]);
			} else {
				for (let i = 0; i < text.length; i++) {
					target[i] = text.charCodeAt(i);
				}
			}

			return [bytes.length, null];
		},
		Close() {
			return null;
		}
	};
}

function go2jsHexDestination(dst) {
	let value = dst;

	for (let depth = 0; depth < 8; depth++) {
		value = go2jsUnwrap(value);

		if (value !== null && value !== undefined && value.__go2js_interface === true) {
			value = value.value;
			continue;
		}

		if (value !== null && value !== undefined && value.__go2js_pointer === true) {
			value = go2jsDeref(value);
			continue;
		}

		return value;
	}

	return value;
}

function go2jsHexBytes(src) {
	if (src === null || src === undefined) {
		return [];
	}

	if (Array.isArray(src)) {
		return src.map((b) => b & 0xff);
	}

	const text = go2jsStringify(src);
	const out = [];

	for (let i = 0; i < text.length; i++) {
		out.push(text.charCodeAt(i) & 0xff);
	}

	return out;
}

function go2jsStringsCount(s, substr) {
	if (substr === "") {
		return s.length + 1;
	}
	return s.split(substr).length - 1;
}

function go2jsStrconvItoa(value, base) {
	base = base === undefined || base === 0 ? 10 : base;

	// A count wider than a double holds exactly keeps its own kind of whole
	// number, and the digits of that number are the digits Go writes.
	if (typeof value === "bigint") {
		return value.toString(base);
	}

	return Math.trunc(value).toString(base);
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

function go2jsStrconvParseInt(s, base, bitSize) {
	const text = go2jsStringify(s).trim();

	base = base === undefined || base === 0 ? 10 : base;

	const parsed = go2jsStrconvParseInteger(text, base);

	if (parsed === null) {
		return [0, new Error('strconv.ParseInt: parsing "' + text + '": invalid syntax')];
	}

	// A signed count at the full sixty four bits keeps the digits it was given,
	// because a double cannot hold them and rounding them would answer with a
	// number Go never wrote.
	if (bitSize === 64 || bitSize === 0 || bitSize === undefined) {
		const digits = text.replace(/^[+-]/, "");

		if (base === 10 && /^[0-9]+$/.test(digits)) {
			const exact = BigInt(digits);

			if (text.startsWith("-")) {
				if (exact <= 9223372036854775808n) {
					return [-exact, null];
				}

				return [0, new Error('strconv.ParseInt: parsing "' + text + '": value out of range')];
			}

			if (exact <= 9223372036854775807n) {
				return [exact, null];
			}

			return [0, new Error('strconv.ParseInt: parsing "' + text + '": value out of range')];
		}

		if (base === 16 && /^[0-9a-fA-F]+$/.test(digits)) {
			const exact = BigInt("0x" + digits);

			return text.startsWith("-") ? [-exact, null] : [exact, null];
		}
	}

	if (!go2jsStrconvFitsWidth(parsed, bitSize)) {
		return [0, new Error('strconv.ParseInt: parsing "' + text + '": value out of range')];
	}

	return [parsed, null];
}

function go2jsStrconvParseUint(s, base, bitSize) {
	const text = go2jsStringify(s).trim();

	if (text.startsWith("-") || text.startsWith("+")) {
		return [0, new Error('strconv.ParseUint: parsing "' + text + '": invalid syntax')];
	}

	// A count wider than a double can hold exactly is a count JavaScript keeps as
	// a whole number of its own kind, which is where the digits are kept.
	const digits = go2jsStrconvParseInteger(text, base);

	if (digits === null) {
		return [0, new Error('strconv.ParseUint: parsing "' + text + '": invalid syntax')];
	}

	if (bitSize === 64 || bitSize === 0 || bitSize === undefined) {
		const exact = BigInt(text);

		if (exact >= 0n && exact <= 18446744073709551615n) {
			return [exact, null];
		}
	}

	const value = digits;

	if (value < 0 || !go2jsStrconvFitsWidth(value, bitSize)) {
		return [0, new Error('strconv.ParseUint: parsing "' + text + '": value out of range')];
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

function go2jsStrconvParseBool(s) {
	if (s === "true" || s === "1" || s === "t" || s === "T") {
		return [true, null];
	}
	if (s === "false" || s === "0" || s === "f" || s === "F") {
		return [false, null];
	}
	return [false, new Error("strconv.ParseBool: parsing " + JSON.stringify(s) + ": invalid syntax")];
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

function go2jsStrconvAppendEscapedRune(out, code, quote, asciiOnly) {
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
	} else if (go2jsStrconvRuneIsPrint(code)) {
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

	let prefix;
	let width;

	if (code < 0x20 || code === 0x7f) {
		prefix = "\\x";
		width = 2;
	} else if (!Number.isInteger(code) || code < 0 || code > 0x10ffff) {
		prefix = "\\u";
		width = 4;
		code = 0xfffd;
	} else if (code < 0x10000) {
		prefix = "\\u";
		width = 4;
	} else {
		prefix = "\\U";
		width = 8;
	}

	out.push(prefix);

	for (let shift = (width - 1) * 4; shift >= 0; shift -= 4) {
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

		go2jsStrconvAppendEscapedRune(out, rune, '"', false);
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

function go2jsStrconvQuoteWith(value, asciiOnly) {
	const out = ['"'];

	for (const char of String(value)) {
		go2jsStrconvAppendEscapedRune(out, char.codePointAt(0), '"', asciiOnly);
	}

	out.push('"');

	return out.join("");
}

function go2jsStrconvQuote(s) {
	return go2jsStrconvQuoteWith(s, false);
}

function go2jsStrconvQuoteToASCII(s) {
	return go2jsStrconvQuoteWith(s, true);
}

function go2jsStrconvUnquote(s) {
	try {
		if (s.length >= 2 && s[0] === String.fromCharCode(96) && s[s.length - 1] === String.fromCharCode(96)) {
			return [s.slice(1, -1), null];
		}
		return [JSON.parse(s), null];
	} catch (err) {
		// The error a quoted string that cannot be read gives back is the plain
		// one Go hands back, with nothing said about which function ran.
		return ["", new Error("invalid syntax")];
	}
}

function go2jsStrconvFormatVerb(value) {
	if (typeof value === "string") {
		return value;
	}

	if (value === null || value === undefined) {
		return "";
	}

	const code = Number(value);

	if (!Number.isInteger(code) || code < 0 || code > 0x10ffff) {
		return String(value);
	}

	return String.fromCodePoint(code);
}

// go2jsStrconvExponent writes an exponent the way Go writes one: a sign and at
// least two digits, because a one digit exponent is written 0 in front of it.
function go2jsStrconvExponent(text, upper) {
	return text.replace(/[eE]([+-]?\d+)$/, (_, digits) => {
		const sign = digits.startsWith("-") ? "-" : "+";
		const rest = digits.replace(/^[+-]/, "");

		return (upper ? "E" : "e") + sign + (rest.length < 2 ? "0" + rest : rest);
	});
}

function go2jsStrconvFormatFloat(value, format, precision, bitSize) {
	let number = Number(value);

	format = go2jsStrconvFormatVerb(format);

	if (typeof value === "bigint") {
		return value.toString();
	}

	if (Number.isNaN(number)) {
		return "NaN";
	}

	if (!Number.isFinite(number)) {
		return number < 0 ? "-Inf" : "+Inf";
	}

	// A number that was asked about as a smaller kind of number is rounded to
	// that kind first, so the digits written are the digits it keeps.
	if (bitSize === 32) {
		number = Math.fround(number);
	}

	switch (format) {
	case "f":
		return precision >= 0 ? number.toFixed(precision) : String(number);
	case "e":
		return go2jsStrconvExponent(number.toExponential(precision < 0 ? 6 : precision), false);
	case "E":
		return go2jsStrconvExponent(number.toExponential(precision < 0 ? 6 : precision), true);
	case "g":
	case "G":
		// The shortest form is the number as JavaScript writes it, and it is
		// written as an exponent once the position of the first digit is far
		// enough from the point, which is the same distance Go uses.
		const text = go2jsStrconvShortestFloat(number);

		if (precision >= 0 && (format === "g" ? precision !== 0 : precision !== 0)) {
			const fixed = format === "g" ? number.toPrecision(precision) : number.toExponential(precision - 1);

			return format === "g" ? go2jsStrconvExponent(fixed, false) : go2jsStrconvExponent(fixed, true);
		}

		return go2jsStrconvExponent(text, format === "G");
	default:
		return String(number);
	}
}

// go2jsStrconvShortestFloat writes a number the way the shortest form of Go's
// strconv writes it, which switches to an exponent once the digits run far from
// the point in either direction.
function go2jsStrconvShortestFloat(number) {
	if (number === 0) {
		return String(number);
	}

	const magnitude = Math.abs(number);
	const exponent = Math.floor(Math.log10(magnitude));
	const digits = String(number).replace(/^-/, "").replace(".", "").replace(/e.*$/, "").replace(/0+$/, "") || "0";

	// Go writes the exponent form once the first digit is left of one past the
	// point for a large number, or right of four under it for a small one.
	if (exponent < -4 || exponent >= 6) {
		const mantissa = number / Math.pow(10, exponent);
		const rounded = Number(mantissa.toPrecision(17));

		return String(rounded) + "e" + (exponent < 0 ? "-" : "+") + Math.abs(exponent);
	}

	return String(number);
}

function go2jsStrconvFormatBool(value) {
	return value ? "true" : "false";
}

// go2jsStrconvAppendText writes text onto the end of a byte slice, starting a
// slice of its own when there is not one yet.
function go2jsStrconvAppendText(dst, text) {
	const bytes = Array.isArray(dst) ? dst.slice() : [];

	for (const ch of String(text)) {
		bytes.push(ch.charCodeAt(0));
	}

	return bytes;
}

function go2jsStrconvAppendInt(dst, value, base) {
	const text = go2jsStrconvItoa(value, base || 10);

	return go2jsStrconvAppendText(dst, text);
}

function go2jsStrconvAppendFloat(dst, value, format, precision, bitSize) {
	const text = go2jsStrconvFormatFloat(value, format, precision, bitSize);

	return go2jsStrconvAppendText(dst, text);
}

function go2jsStrconvAppendBool(dst, value) {
	return go2jsStrconvAppendText(dst, value ? "true" : "false");
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

function go2jsSortInterface(data, typeName) {
	if (data === null || data === undefined) {
		return;
	}

	const length = go2jsSortMethod(data, typeName, "Len");
	const less = go2jsSortMethod(data, typeName, "Less");
	const swap = go2jsSortMethod(data, typeName, "Swap");

	if (typeof length === "function" && typeof less === "function" && typeof swap === "function") {
		const count = Number(length());

		for (let index = 1; index < count; index++) {
			let current = index;

			while (current > 0 && go2jsCallNow(less, null, [current, current - 1])) {
				swap(current, current - 1);
				current--;
			}
		}

		return;
	}

	const items = go2jsToArray(data);

	items.sort((a, b) => (a < b ? -1 : a > b ? 1 : 0));

	if (Array.isArray(data)) {
		data.length = 0;
		data.push(...items);
	}
}

function go2jsSortIsSortedInterface(data, typeName) {
	const less = go2jsSortMethod(data, typeName, "Less");
	const items = go2jsToArray(data);

	for (let index = 1; index < items.length; index++) {
		const outOfOrder = typeof less === "function"
			? go2jsCallNow(less, null, [index, index - 1])
			: items[index] < items[index - 1];

		if (outOfOrder) {
			return false;
		}
	}

	return true;
}

function go2jsSortSearch(count, found) {
	const total = Number(count);
	let low = 0;
	let high = total;

	while (low < high) {
		const mid = Math.floor((low + high) / 2);

		// The function says whether the item at a position sorts at or after
		// what is being looked for, and the search keeps the half it describes.
		if (go2jsCallNow(found, null, [mid])) {
			high = mid;
		} else {
			low = mid + 1;
		}
	}

	return low;
}

function go2jsSortIsSorted(values, less) {
	const items = go2jsToArray(values);

	for (let i = 1; i < items.length; i++) {
		if (go2jsCallNow(less, null, [items[i], items[i - 1]])) {
			return false;
		}
	}

	return true;
}

function go2jsSortFloat64s(values) {
	values.sort((a, b) => a - b);
}

function go2jsSortStrings(values) {
	values.sort();
}

function go2jsMathSignbit(value) {
	return Object.is(value, -0) || value < 0 || value === -Infinity;
}

function go2jsMathIsInf(value, sign) {
	if (sign > 0) {
		return value === Infinity;
	}
	if (sign < 0) {
		return value === -Infinity;
	}
	return value === Infinity || value === -Infinity;
}


function go2jsUTF8Byte(value, index) {
	return Number(value[index]) & 0xff;
}

function go2jsUTF8Continuation(value) {
	return (value & 0xc0) === 0x80;
}

function go2jsUTF8Width(bytes, index) {
	const length = bytes.length;
	if (index >= length) {
		return 0;
	}

	const b0 = go2jsUTF8Byte(bytes, index);

	if (b0 <= 0x7f) {
		return 1;
	}

	if (b0 >= 0xc2 && b0 <= 0xdf) {
		if (index + 1 >= length) {
			return 0;
		}
		const b1 = go2jsUTF8Byte(bytes, index + 1);
		return go2jsUTF8Continuation(b1) ? 2 : 0;
	}

	if (b0 >= 0xe0 && b0 <= 0xef) {
		if (index + 2 >= length) {
			return 0;
		}

		const b1 = go2jsUTF8Byte(bytes, index + 1);
		const b2 = go2jsUTF8Byte(bytes, index + 2)

		if (!go2jsUTF8Continuation(b1) || !go2jsUTF8Continuation(b2)) {
			return 0;
		}

		if (b0 === 0xe0 && b1 < 0xa0) {
			return 0;
		}

		if (b0 === 0xed && b1 >= 0xa0) {
			return 0;
		}

		return 3;
	}

	if (b0 >= 0xf0 && b0 <= 0xf4) {
		if (index + 3 >= length) {
			return 0;
		}

		const b1 = go2jsUTF8Byte(bytes, index + 1);
		const b2 = go2jsUTF8Byte(bytes, index + 2);
		const b3 = go2jsUTF8Byte(bytes, index + 3);

		if (!go2jsUTF8Continuation(b1) ||
			!go2jsUTF8Continuation(b2) ||
			!go2jsUTF8Continuation(b3)) {
			return 0;
		}

		if (b0 === 0xf0 && b1 < 0x90) {
			return 0;
		}

		if (b0 === 0xf4 && b1 > 0x8f) {
			return 0;
		}

		return 4;
	}

	return 0;
}

function go2jsUTF8RuneCount(value) {
	const bytes = Array.from(value);
	let count = 0;
	let index = 0;

	while (index < bytes.length) {
		const width = go2jsUTF8Width(bytes, index);
		if (width > 0) {
			index += width;
		} else {
			index++;
		}
		count++;
	}

	return count;
}

function go2jsUTF8Valid(value) {
	const bytes = Array.from(value);
	let index = 0;

	while (index < bytes.length) {
		const width = go2jsUTF8Width(bytes, index);
		if (width === 0) {
			return false;
		}
		index += width;
	}

	return true;
}

function go2jsUTF8FullRune(value) {
	const bytes = Array.from(value);
	if (bytes.length === 0) {
		return false;
	}

	const b0 = go2jsUTF8Byte(bytes, 0);

	if (b0 <= 0x7f) {
		return true;
	}

	if (b0 >= 0xc2 && b0 <= 0xdf) {
		return bytes.length >= 2;
	}

	if (b0 >= 0xe0 && b0 <= 0xef) {
		return bytes.length >= 3;
	}

	if (b0 >= 0xf0 && b0 <= 0xf4) {
		return bytes.length >= 4;
	}

	return true;
}

function go2jsUTF8FullRuneInString(s) {
	return s.length > 0;
}


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

function go2jsUnicodeIsControl(value) {
	const ch = go2jsUnicodeRune(value);
	return ch !== "" && /^\p{Cc}$/u.test(ch);
}

function go2jsUnicodeIsDigit(value) {
	const ch = go2jsUnicodeRune(value);
	return ch !== "" && /^\p{Nd}$/u.test(ch);
}

function go2jsUnicodeIsGraphic(value) {
	const ch = go2jsUnicodeRune(value);
	return ch !== "" && /^[\p{L}\p{M}\p{N}\p{P}\p{S}\p{Zs}]$/u.test(ch);
}

function go2jsUnicodeIsLetter(value) {
	const ch = go2jsUnicodeRune(value);
	return ch !== "" && /^\p{L}$/u.test(ch);
}

function go2jsUnicodeIsLower(value) {
	const ch = go2jsUnicodeRune(value);
	return ch !== "" && /^\p{Ll}$/u.test(ch);
}

function go2jsUnicodeIsMark(value) {
	const ch = go2jsUnicodeRune(value);
	return ch !== "" && /^\p{M}$/u.test(ch);
}

function go2jsUnicodeIsNumber(value) {
	const ch = go2jsUnicodeRune(value);
	return ch !== "" && /^\p{N}$/u.test(ch);
}

function go2jsUnicodeIsPrint(value) {
	const ch = go2jsUnicodeRune(value);
	return ch !== "" && (value === 0x20 || /^[\p{L}\p{M}\p{N}\p{P}\p{S}]$/u.test(ch));
}

function go2jsUnicodeIsPunct(value) {
	const ch = go2jsUnicodeRune(value);
	return ch !== "" && /^\p{P}$/u.test(ch);
}

function go2jsUnicodeIsSpace(value) {
	const ch = go2jsUnicodeRune(value);
	return ch !== "" && /^\p{White_Space}$/u.test(ch);
}

function go2jsUnicodeIsSymbol(value) {
	const ch = go2jsUnicodeRune(value);
	return ch !== "" && /^\p{S}$/u.test(ch);
}

function go2jsUnicodeIsTitle(value) {
	const ch = go2jsUnicodeRune(value);
	return ch !== "" && /^\p{Lt}$/u.test(ch);
}

function go2jsUnicodeIsUpper(value) {
	const ch = go2jsUnicodeRune(value);
	return ch !== "" && /^\p{Lu}$/u.test(ch);
}

function go2jsUnicodeMapCase(value, mode) {
	const r = go2jsUnicodeCodePoint(value);
	const ch = go2jsUnicodeRune(value);

	if (ch === "") {
		return r;
	}

	let mapped;

	switch (mode) {
	case "lower":
		mapped = ch.toLowerCase();
		break;
	case "title":
		mapped = ch.toUpperCase();
		break;
	default:
		mapped = ch.toUpperCase();
		break;
	}

	const runes = Array.from(mapped);

	if (runes.length !== 1) {
		return r;
	}

	return runes[0].codePointAt(0);
}

function go2jsUnicodeToLower(value) {
	return go2jsUnicodeMapCase(value, "lower");
}

function go2jsUnicodeToTitle(value) {
	return go2jsUnicodeMapCase(value, "title");
}

function go2jsUnicodeToUpper(value) {
	return go2jsUnicodeMapCase(value, "upper");
}

function go2jsUTF8RuneCountInString(s) {
	return Array.from(s).length;
}

function go2jsUTF8RuneLen(value) {
	const r = go2jsUnicodeCodePoint(value);

	if (r < 0 || r > 0x10ffff || (r >= 0xd800 && r <= 0xdfff)) {
		return -1;
	}

	if (r <= 0x7f) {
		return 1;
	}

	if (r <= 0x7ff) {
		return 2;
	}

	if (r <= 0xffff) {
		return 3;
	}

	return 4;
}

// A rune is read out of a string by its first character, which is what a string
// of runes holds one rune at a time, and a byte or a byte slice is read out of
// it by the width that leads the bytes there.
function go2jsUTF8DecodeRuneInString(s) {
	const text = String(s);

	if (text.length === 0) {
		return [0xfffd, 0];
	}

	const code = text.codePointAt(0);

	return [code, text.length >= 2 && code > 0xffff ? 2 : 1];
}

function go2jsUTF8DecodeLastRuneInString(s) {
	const text = String(s);

	if (text.length === 0) {
		return [0xfffd, 0];
	}

	// the runes of a string are read from its end, since the last one is the
	// last character, and a character of two code units is the one before it
	let at = text.length - 1;

	if (at > 0 && text.charCodeAt(at) < 0xdc00 || (at > 0 && text.charCodeAt(at) >= 0xdc00 && text.charCodeAt(at) <= 0xdfff)) {
		if (text.charCodeAt(at) >= 0xdc00 && text.charCodeAt(at) <= 0xdfff) {
			at--;
		}
	}

	const code = text.codePointAt(at);

	return [code, code > 0xffff ? 2 : 1];
}

function go2jsUTF8DecodeRune(value) {
	const bytes = Array.from(value);

	if (bytes.length === 0) {
		return [0xfffd, 0];
	}

	return go2jsUTF8DecodeRuneInString(Buffer.from(bytes).toString("utf8"));
}

function go2jsUTF8DecodeLastRune(value) {
	const text = Buffer.from(Array.from(value)).toString("utf8");

	return go2jsUTF8DecodeLastRuneInString(text);
}

// go2jsUTF8AppendRune writes a rune onto the end of a byte slice, which is the
// rune written down as the bytes it stands for.
function go2jsUTF8AppendRune(bytes, value) {
	const start = bytes === null || bytes === undefined ? [] : Array.from(bytes);

	return Array.from(Buffer.concat([Buffer.from(start), Buffer.from(String.fromCodePoint(go2jsUnicodeCodePoint(value)), "utf8")]));
}

function go2jsUTF8RuneStart(value) {
	return (Number(value) & 0xc0) !== 0x80;
}

function go2jsUTF8ValidRune(value) {
	const r = go2jsUnicodeCodePoint(value);
	return r >= 0 &&
		r <= 0x10ffff &&
		!(r >= 0xd800 && r <= 0xdfff);
}

function go2jsUTF8ValidString(s) {
	for (let i = 0; i < s.length; i++) {
		const value = s.charCodeAt(i);

		if (value >= 0xd800 && value <= 0xdbff) {
			if (i + 1 >= s.length) {
				return false;
			}

			const next = s.charCodeAt(i + 1);

			if (next < 0xdc00 || next > 0xdfff) {
				return false;
			}

			i++;
			continue;
		}

		if (value >= 0xdc00 && value <= 0xdfff) {
			return false;
		}
	}

	return true;
}

go2jsRegisterMethod("encoding/json.Delim.String", go2jsJSONDelimText);

`
}
