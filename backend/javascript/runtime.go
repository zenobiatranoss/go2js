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

// go2jsFilepathHasPrefix reads a name as one beginning with the text of another,
// which is a question about the text and not about the parts of a path.
function go2jsFilepathHasPrefix(value, prefix) {
	return go2jsStringsHasPrefix(value, prefix);
}

// go2jsFilepathIsLocal reads a name as one kept inside the tree it is read
// against: not empty, not rooted, and with no part that steps out of it.
function go2jsFilepathIsLocal(value) {
	const path = String(value);

	if (path === "" || path.startsWith("/")) {
		return false;
	}

	let steps = false;

	for (const part of path.split("/")) {
		if (part === "." || part === "..") {
			steps = true;
			break;
		}
	}

	const cleaned = steps ? go2jsFilepathClean(path) : path;

	return cleaned !== ".." && !cleaned.startsWith("../");
}

// go2jsFilepathLocalize reads a slash separated name as a name of the machine's
// paths, which for the machine this runtime stands for is the same text unless
// the name is not a clean relative one or holds a byte a name cannot hold.
function go2jsFilepathLocalize(value) {
	const path = String(value);

	if (!go2jsValidPath(path) || path.indexOf("\u0000") >= 0) {
		return ["", go2jsSentinelError("invalid path")()];
	}

	return [path, null];
}

// go2jsValidPath reads a name as one the file tree allows: the current
// directory, or a name of parts none of which is empty, the current directory
// or the one above it.
function go2jsValidPath(path) {
	if (path === ".") {
		return true;
	}

	if (path === "") {
		return false;
	}

	for (const part of path.split("/")) {
		if (part === "" || part === "." || part === "..") {
			return false;
		}
	}

	return true;
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

// The separator the runtime names paths with is the one it draws them with, so a
// path already written with that separator has nothing to change, and the two
// ways of writing one are the same way.
function go2jsFilepathToSlash(value) {
	return String(value);
}

function go2jsFilepathFromSlash(value) {
	return String(value);
}

// A pattern names the files it matches with the marks of a shell, and each of
// them stands for one thing: a star stands for any run of characters that are
// not a separator, a question for one such character, and a class for one of
// the characters it lists. The pattern is written as a pattern of the same
// shape rather than matched a mark at a time, because the marks are only ever
// matched together.
function go2jsFilepathMatch(pattern, name, badPattern) {
	const compiled = go2jsFilepathPatternSource(pattern, badPattern);

	if (compiled[1] !== null) {
		return compiled;
	}

	return [new RegExp(compiled[0]).test(go2jsStringify(name)), null];
}

// path and filepath each name their own malformed-pattern sentinel, and both
// are the same error string, so each is memoized under its own name to keep
// them the distinct values Go keeps them as.
function go2jsPathBadPattern() {
	return go2jsSentinelError("syntax error in pattern", "*errors.errorString", "path.ErrBadPattern")();
}

function go2jsFilepathBadPattern() {
	return go2jsSentinelError("syntax error in pattern", "*errors.errorString", "filepath.ErrBadPattern")();
}

// go2jsPathMatch names files with the pattern language of a shell, which is the
// same language whether the names are read as paths of the machine or as paths
// of the slash, since the slash is the separator the runtime names them with.
function go2jsPathMatch(pattern, name) {
	return go2jsFilepathMatch(pattern, name, go2jsPathBadPattern);
}

// go2jsFilepathSplitList cuts a list of paths apart on the mark that separates
// them, which is a colon on the systems this runtime stands for, and reads the
// empty text as the empty list rather than as one empty name.
function go2jsFilepathSplitList(value) {
	const text = go2jsStringify(value);

	if (text === "") {
		return go2jsZeroArray(0, () => "");
	}

	return String(value).split(":");
}

// go2jsFilepathPatternSource writes a pattern as one of the same shape, which is
// what a name is asked against, and reports a pattern that cannot be written.
function go2jsFilepathPatternSource(pattern, badPattern) {
	const bad = badPattern === undefined || badPattern === null ? go2jsFilepathBadPattern : badPattern;
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
				return [false, bad()];
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
					return [false, bad()];
				}

				cursor += Number(lo[0].slice(1));
				let piece = lo[1];

				if (pattern[cursor] === "-") {
					const hi = go2jsFilepathClassChar(pattern, cursor + 1);

					if (hi === null) {
						return [false, bad()];
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

	// RequestURI is the URL as it is written on a request line: the path it
	// asks for, or a slash when it names none, with the query after it. An
	// opaque URL is its own request target, and one that begins with two
	// slashes is given the scheme back so the two are not mistaken for an
	// authority.
	url.RequestURI = function() {
		let result = url.Opaque;

		if (result === "") {
			result = url.EscapedPath();

			if (result === "") {
				result = "/";
			}
		} else if (result.startsWith("//")) {
			result = url.Scheme + ":" + result;
		}

		if (url.ForceQuery || url.RawQuery !== "") {
			result += "?" + url.RawQuery;
		}

		return result;
	};

	url.EscapedFragment = function() {
		const escaped = url.RawFragment;

		return escaped !== "" && go2jsURLValidEncoded(escaped, "fragment") ? escaped : go2jsURLFragmentEscape(url.Fragment);
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
	const value = go2jsStringify(text);

	for (let index = 0; index < value.length; index++) {
		const char = value[index];

		// A percent is written with the two digits that name the byte, so it is
		// read with them rather than asked whether it escapes on its own.
		if (char === "%") {
			const named = value.slice(index + 1, index + 3);

			if (named.length < 2 || !/^[0-9A-Fa-f]{2}$/.test(named)) {
				return false;
			}

			index += 2;
			continue;
		}

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


function go2jsJSONMarshal(value, fields, omitEmpty, stringFields, textFields, prefix, indent, types) {
    try {
        const encoded = go2jsJSONEncode(value, fields, omitEmpty, stringFields, textFields, prefix, indent, types);

        // the text a marshal writes is the bytes of a []byte, so it is kept as
        // bytes rather than as the text itself, since a byte slice and a string
        // are read differently even when they hold the same characters
        return [go2jsStringToBytes(go2jsJSONText(encoded, prefix, indent, true)), null];
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

function go2jsJSONEncode(value, fields, omitEmpty, stringFields, textFields, prefix, indent, types) {
    // A nil slice or a nil map travels as a value of its own that still says
    // which it is, and JSON writes every nil as nothing at all.
    if (value !== null && value !== undefined && typeof value === "object" && value.__go2js_nil === true) {
        return null;
    }

    // A pointer is written as the value it points at, and a nil pointer is
    // written as nothing, which is what the value behind it being nothing
    // comes down to.
    while (value !== null && value !== undefined && typeof value === "object" && value.__go2js_pointer === true) {
        value = go2jsDeref(value);

        if (types !== null && types !== undefined && types.kind === "ptr") {
            types = types.elem;
        }
    }

    if (value === null || value === undefined) {
        return null;
    }

    // A number JSON cannot write is refused rather than written as nothing,
    // since a NaN and an infinity have no text a JSON number is written with.
    if (typeof value === "number") {
        if (Number.isNaN(value)) {
            throw new Error("json: unsupported value: NaN");
        }

        if (value === Infinity) {
            throw new Error("json: unsupported value: +Inf");
        }

        if (value === -Infinity) {
            throw new Error("json: unsupported value: -Inf");
        }

        return value;
    }

    // A slice of bytes is written as the base64 of its bytes rather than as
    // the list of numbers it holds, which is what a []byte is for JSON.
    if (types !== null && types !== undefined && types.elem !== null && types.elem !== undefined && types.elem.kind === "byte" &&
        (Array.isArray(value) || value instanceof Uint8Array)) {
        return go2jsBase64EncodeToString(go2jsBase64StdEncoding(), value);
    }

    if (Array.isArray(value)) {
        const elem = types !== null && types !== undefined && types.elem !== undefined ? types.elem : null;

        return value.map(item => go2jsJSONEncode(item, null, null, null, null, prefix, indent, elem));
    }

    if (value instanceof go2jsNativeMap) {
        const elem = types !== null && types !== undefined && types.elem !== undefined ? types.elem : null;

        // A map's keys are written in the order of their names rather than in
        // the order they were put in, so a map and its text stand in one
        // order however the map was built.
        const entries = [];

        for (const [key, item] of value.entries()) {
            entries.push([go2jsStringify(key), item]);
        }

        entries.sort((left, right) => left[0] < right[0] ? -1 : left[0] > right[0] ? 1 : 0);

        const out = {};

        for (const [key, item] of entries) {
            out[key] = go2jsJSONEncode(item, null, null, null, null, prefix, indent, elem);
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

    // what each field holds was written beside the value when it was written
    // by a program, so a field holding bytes or a pointer of its own is known
    // here rather than found out by looking the value over
    const held = types !== null && types !== undefined && types.types !== undefined ? types.types : null;

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

            const fieldTypes = held !== null && held[field] !== undefined ? held[field] : null;

            out[name] = go2jsJSONEncode(item, null, null, null, null, prefix, indent, fieldTypes);
        }

        return out;
    }

    const out = {};

    for (const key of Object.keys(value)) {
        const fieldTypes = held !== null && held[key] !== undefined ? held[key] : null;

        out[key] = go2jsJSONEncode(value[key], null, null, null, null, prefix, indent, fieldTypes);
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

        // The text is checked the way Go checks it before any of it is read, so
        // a text that is not a value fails with the character Go names and the
        // room it was found in rather than with whatever the host parser says.
        const scanError = go2jsJSONScanErrorOf(text);

        if (scanError !== "") {
            return go2jsJSONError(scanError);
        }

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
        return go2jsJSONCoerceAny(value, text, path);
    }

    // a string where a slice of bytes is read is the base64 of those bytes
    // rather than a value of another kind, so it is read as bytes and not
    // refused for being a string
    if (descriptor.kind === "slice" && descriptor.elem !== null && descriptor.elem !== undefined &&
        descriptor.elem.kind === "byte" && typeof value === "string") {
        return go2jsJSONBase64Bytes(value);
    }

    // a value of another kind than the destination holds is refused, and the
    // read goes on so that what did fit is still there beside the failure
    if (descriptor.kind !== "ptr" && descriptor.kind !== "any" && !go2jsJSONFitsKind(value, descriptor.kind)) {
        go2jsJSONSaveError(go2jsJSONError("json: cannot unmarshal " + go2jsJSONValueKind(value) +
            " into Go value of type " + go2jsJSONGoTypeName(descriptor)));

        return go2jsJSONZeroValue(descriptor);
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
        return go2jsJSONCoerce(value, descriptor.kind, text, path);
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

function go2jsJSONEncoderEncode(encoder, value, fields, omitEmpty, stringFields, textFields, types) {
    try {
        const encoded = go2jsJSONEncode(value, fields, omitEmpty, stringFields, textFields, encoder.prefix, encoder.indent, types);
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
        useNumber: false,
        stack: []
    };
}

// go2jsJSONDecoderUseNumber asks a decoder to keep the text of every number it
// reads rather than the number the text stands for, since a number no double
// holds the digits of is kept by holding on to how it was written.
function go2jsJSONDecoderUseNumber(decoder) {
    decoder.useNumber = true;
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

    // A decoder that was asked for the text of numbers has it while the read
    // runs, and one that was not leaves the flag as it found it.
    const previous = go2jsJSONNumberAsText;
    go2jsJSONNumberAsText = decoder.useNumber === true;

    try {
        return go2jsJSONUnmarshal(text, target, fields, destination, stringFields, fieldTypes);
    } finally {
        go2jsJSONNumberAsText = previous;
    }
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

// A read a decoder asked for the text of numbers keeps every number as the text
// it was written with rather than as the double that text stands for.
let go2jsJSONNumberAsText = false;

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
    case "array":
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

// go2jsJSONGoTypeName is the type a destination was declared with, named the way
// Go names it, which is what a read that did not fit says the value was refused
// into. A pointer is named by what it points at, since that is what the value
// was refused into.
function go2jsJSONGoTypeName(descriptor) {
    if (descriptor === null || descriptor === undefined) {
        return "value";
    }

    if (descriptor.kind === "ptr") {
        return go2jsJSONGoTypeName(descriptor.elem);
    }

    if (typeof descriptor.go === "string" && descriptor.go !== "") {
        return descriptor.go;
    }

    return go2jsJSONTypeName(descriptor.kind);
}

// go2jsJSONFieldPath names a field the way Go names one in a read that did not
// fit: the struct type it was declared in and the name JSON knows it by.
function go2jsJSONFieldPath(target, fields, name, key) {
    let label = "";

    if (target !== null && target !== undefined) {
        const found = go2jsGoTypeNameRaw(target);

        if (typeof found === "string" && found !== "") {
            const dot = found.lastIndexOf(".");
            label = dot >= 0 ? found.slice(dot + 1) : found;
        }
    }

    let jsonName = key;

    if (fields !== null && fields !== undefined && fields[name] !== undefined) {
        jsonName = fields[name];
    }

    return label === "" ? jsonName : label + "." + jsonName;
}

// go2jsJSONBase64Bytes is the bytes a JSON string stands for, since a string
// where a slice of bytes is read is the base64 of those bytes. A string that is
// not base64 is refused with the same words Go refuses it with.
function go2jsJSONBase64Bytes(value) {
    const decoded = go2jsBase64DecodeString(go2jsBase64StdEncoding(), value);

    if (decoded[1] !== null && decoded[1] !== undefined) {
        go2jsJSONSaveError(go2jsJSONError(decoded[1].message));

        return [];
    }

    return decoded[0];
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
function go2jsJSONCoerceAny(value, text, path) {
    if (value === null || value === undefined) {
        return null;
    }

    if (Array.isArray(value)) {
        return go2jsJSONAnyList(value.map((item, index) => go2jsJSONCoerceAny(item, text, go2jsJSONPath(path, String(index)))));
    }

    if (typeof value === "object") {
        const map = go2jsJSONAnyMap(go2jsMakeMap());

        for (const key of Object.keys(value)) {
            go2jsMapSet(map, key, go2jsJSONCoerceAny(value[key], text, go2jsJSONPath(path, key)));
        }

        return map;
    }

    if (typeof value === "number") {
        // A number a decoder asked to keep as text is kept as the text it was
        // written with, which is the only way a number no double holds the
        // digits of keeps them.
        if (go2jsJSONNumberAsText && text !== null && text !== undefined) {
            const written = go2jsJSONRawText(text, path);

            if (written !== null && written !== undefined && go2jsJSONNumberLiteral(written)) {
                return go2jsInterface(written, "json.Number", "json.Number", false);
            }
        }

        return go2jsJSONAnyNumber(value);
    }

    return value;
}

function go2jsJSONCoerce(value, kind, text, path) {
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

    return go2jsJSONCoerceAny(value, text, path);
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
        // text is read back into the value it stands for, and a value written
        // without that text around it is refused the way Go refuses it
        if (stringFields && stringFields.indexOf(name) >= 0) {
            const marked = fieldTypes !== null && fieldTypes !== undefined && fieldTypes[name] !== undefined ? fieldTypes[name] : null;

            if (typeof item !== "string") {
                if (item !== null && item !== undefined) {
                    go2jsJSONSaveError(go2jsJSONError("json: invalid use of ,string struct tag, trying to unmarshal unquoted value into " + go2jsJSONGoTypeName(marked)));
                }

                continue;
            }

            try {
                target[name] = JSON.parse(item);
            } catch (err) {
                go2jsJSONSaveError(go2jsJSONError("json: invalid use of ,string struct tag, trying to unmarshal " + JSON.stringify(item) + " into " + go2jsJSONGoTypeName(marked)));
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

            // a string read into a slice of bytes is the base64 of those bytes,
            // so it is not refused for being a string
            const byteText = declared !== null && declared !== undefined && declared.kind === "slice" &&
                declared.elem !== null && declared.elem !== undefined && declared.elem.kind === "byte" &&
                typeof item === "string";

            if (!byteText && declared !== null && declared !== undefined && declared.kind !== undefined &&
                !go2jsJSONFitsKind(item, go2jsJSONElemKind(declared))) {
                go2jsJSONSaveError(go2jsJSONError("json: cannot unmarshal " + go2jsJSONValueKind(item) +
                    " into Go struct field " + go2jsJSONFieldPath(target, fields, name, key) + " of type " + go2jsJSONGoTypeName(declared)));
                continue;
            }

            target[name] = go2jsJSONStore(item, null, declared, text, go2jsJSONPath(path, key));
            continue;
        }

        // A field the program declared no type for holds a value nothing is
        // known about, which is read as such rather than as the plain value the
        // text happened to be written with.
        target[name] = go2jsJSONCoerceAny(item, text, go2jsJSONPath(path, key));
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

	const bytes = result[0] === null || result[0] === undefined ? [] : go2jsToArray(result[0]);

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

function go2jsToRune(value) {
	return go2jsRuneString(value);
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

function go2jsEqual(a, b, deep) {
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
		// A slice or a map or a function is never a pair of sides that can be
		// called equal, so even the same value on both sides is a panic in Go.
		// reflect.DeepEqual does not compare by the == of Go, so it is allowed
		// to look at one of them as deep as it goes.
		if (deep !== true && a !== null && typeof a === "object" && a.__go2js_uncomparable === true && a.__go2js_interface === true) {
			throw new Error(go2jsRuntimeErrorPrefix + "comparing uncomparable type " + (a.type || ""));
		}

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

		// The two hold the same shape, so a value that cannot be compared would
		// have to be, which is a panic in Go. The wrapper holds that up before
		// a slice is asked to weigh itself element by element, unless the
		// compare is the deep one of reflect.DeepEqual, which is allowed to.
		if (deep !== true && (a.__go2js_uncomparable === true || b.__go2js_uncomparable === true)) {
			throw new Error(go2jsRuntimeErrorPrefix + "comparing uncomparable type " + (a.type || ""));
		}

		// An array that is its own interface wrapper keeps its elements
		// directly, so unwrapping would only arrive back at the same object.
		if (a.value !== a && b.value !== b) {
			return go2jsEqual(a.value, b.value, deep);
		}
	} else if (ai || bi) {
		// One side is a value that was stored in an interface and the other is
		// not. Go compares the value the interface holds against the value on
		// the other side, so the wrapper is peeled off and the two values are
		// compared, which leaves a value of a different type unequal. A bare
		// number or string or boolean has no type of its own on the way in, so
		// the only box it can be equal to is one that names the plain type a
		// JavaScript value of that kind already is: an int64 is never an int,
		// no matter how they line up as numbers.
		const wrapped = ai ? a : b;
		const plain = ai ? b : a;

		if (wrapped.value === wrapped) {
			return false;
		}

		if (typeof plain !== "object" || plain === null) {
			if (typeof plain === "number") {
				if (Number.isInteger(plain)) {
					if (wrapped.type !== "int") {
						return false;
					}
				} else if (wrapped.type !== "float64") {
					return false;
				}
			} else if (typeof plain === "string" && wrapped.type !== "string") {
				return false;
			} else if (typeof plain === "boolean" && wrapped.type !== "bool") {
				return false;
			}
		}

		return go2jsEqual(wrapped.value, plain, deep);
	}

	if (Array.isArray(a) || Array.isArray(b)) {
		if (!Array.isArray(a) || !Array.isArray(b) || a.length !== b.length) {
			return false;
		}

		for (let index = 0; index < a.length; index++) {
			if (!go2jsEqual(a[index], b[index], deep)) {
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

	// A map is a whole in itself rather than a set of properties that happen
	// to be on it, so it compares by walking the entries one by one: same
	// number of them, and each value in one is equal to the value the same key
	// holds in the other.
	if (a instanceof go2jsNativeMap || b instanceof go2jsNativeMap) {
		if (!(a instanceof go2jsNativeMap) || !(b instanceof go2jsNativeMap) || a.size !== b.size) {
			return false;
		}

		for (const entry of a) {
			const key = entry[0];
			if (!b.has(key) || !go2jsEqual(entry[1], b.get(key), deep)) {
				return false;
			}
		}

		return true;
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

			if (!go2jsEqual(a[key], b[key], deep)) {
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

// go2jsDeepEqual compares the way reflect.DeepEqual does, which is the deep
// way: a slice, a map, or a function is not refused the way the == of Go
// refuses it, but is unwound until nothing but plain values are left.
function go2jsDeepEqual(left, right) {
	return go2jsEqual(left, right, true);
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
	return go2jsTypeOf(value);
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

		// A label names a type the way the source wrote it: an imported type
		// carries its package and a local one does not, while the runtime's name
		// for a type is qualified by the package it was declared in. An exact
		// name is therefore tried first, then an agreement of the two names cut
		// to their last part, and finally the labels that name an interface,
		// which match on the methods the value carries.
		if (name === label ||
			go2jsShortTypeName(name) === go2jsShortTypeName(label) ||
			go2jsSatisfiesInterface(value, label)) {
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

function go2jsAssert(value, typeName, interfaceName) {
	if (value !== null && value !== undefined &&
		value.__go2js_interface === true) {
		if (go2jsSameTypeName(value.type, typeName)) {
			return value.value;
		}

		if (go2jsSatisfiesInterface(value, typeName)) {
			return value;
		}

		// A broken assertion says what the value is and what it was asked to
		// be, the same way Go writes it out, with the interface it was read
		// from in front.
		throw new TypeError(
			"interface conversion: " + interfaceName + " is " + value.type + ", not " + typeName
		);
	}

	const actual = go2jsTypeOf(value);

	if (go2jsSameTypeName(actual, typeName)) {
		return value;
	}

	if (go2jsSatisfiesInterface(value, typeName)) {
		return value;
	}

	throw new TypeError(
		"interface conversion: " + interfaceName + " is " + actual + ", not " + typeName
	);
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

// go2jsBinarySizes is the count of bytes a value of each kind is written in,
// which is the width its type names rather than the width the number it holds
// suggests.
const go2jsBinarySizes = {
	bool: 1,
	int8: 1, uint8: 1,
	int16: 2, uint16: 2,
	int32: 4, uint32: 4, float32: 4,
	int64: 8, uint64: 8, float64: 8, int: 8, uint: 8, uintptr: 8
};

// go2jsBinaryWrite writes a value in the bytes its type names, which is what Go
// writes: a scalar is written once, and a run of them is written one after the
// other. The count of bytes each is written in is read from its type, since the
// number alone does not say whether it was a uint32 or a uint64.
function go2jsBinaryWrite(writer, order, data, kind) {
	let bytes = [];

	if (kind.length > 2 && kind.slice(0, 2) === "[]") {
		const element = kind.slice(2);

		for (const value of go2jsToArray(data)) {
			bytes = bytes.concat(go2jsBinaryScalar(order, value, element));
		}
	} else {
		bytes = go2jsBinaryScalar(order, data, kind);
	}

	const written = go2jsCallMethod(writer, "Write", bytes);

	return Array.isArray(written) && written.length > 1 ? written[1] : null;
}

// go2jsBinaryScalar is one value written in the bytes of the type named, in the
// order asked for.
function go2jsBinaryScalar(order, value, kind) {
	if (kind === "bool") {
		return [go2jsUnwrap(value) === true ? 1 : 0];
	}

	if (kind.slice(0, 5) === "float") {
		return go2jsBinaryFloat(order, Number(go2jsUnwrap(value)), kind === "float32" ? 4 : 8);
	}

	return go2jsBinaryInteger(order, go2jsUnwrap(value), go2jsBinarySizes[kind] === undefined ? 8 : go2jsBinarySizes[kind]);
}

// go2jsBinaryInteger is a whole number written in the bytes of a width, in the
// two's complement bits of that width, so a negative number is the bits it names
// rather than a sign, and most significant byte first.
function go2jsBinaryInteger(order, value, width) {
	const number = typeof value === "bigint" ? value : BigInt(Math.trunc(Number(value)));
	const modulus = 1n << BigInt(width * 8);
	let wrapped = ((number % modulus) + modulus) % modulus;
	const bytes = new Array(width);

	for (let index = width - 1; index >= 0; index--) {
		bytes[index] = Number(wrapped & 0xffn);
		wrapped >>= 8n;
	}

	if (order === "le") {
		bytes.reverse();
	}

	return bytes;
}

// go2jsBinaryFloat is a number written in the bits IEEE 754 gives its width, in
// the order asked for, which is how Go writes a floating point number.
function go2jsBinaryFloat(order, value, width) {
	const view = new DataView(new ArrayBuffer(width));

	if (width === 4) {
		view.setFloat32(0, Number(value), order === "le");
	} else {
		view.setFloat64(0, Number(value), order === "le");
	}

	const bytes = new Array(width);

	for (let index = 0; index < width; index++) {
		bytes[index] = view.getUint8(index);
	}

	return bytes;
}

function go2jsBinaryPutUint(order, width, target, value) {
    let number = Number(value);

    if (!Number.isFinite(number) || number < 0) {
        number = 0;
    }

    const modulus = Math.pow(256, width);

    number = Math.trunc(number) % modulus;

    if (number < 0) {
        number += modulus;
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

    return value % 2 === 0 ? [value / 2, read] : [-(value + 1) / 2, read];
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

// go2jsBinaryPutVarint writes a whole number in the zigzag form Go writes one
// in, which puts the small negative numbers next to the small positive ones: a
// negative number is written as twice the number before it with the low bit set.
function go2jsBinaryPutVarint(buffer, value) {
    const number = Math.trunc(Number(value));

    return go2jsBinaryPutUvarint(buffer, number < 0 ? -2 * number - 1 : 2 * number);
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

// go2jsMapKey hands back the one object every key of that shape shares, so a key
// is found by what it holds. A key that is not a struct or an array is left
// alone, since a string or a number is already found by what it is.
function go2jsMapKey(key) {
	// A slice or a map or a function has no way of being hashed, so a key of
	// one of them is a panic the same way it is in Go, before its contents
	// are mistaken for something that could be told apart.
	if (key !== null && typeof key === "object" && key.__go2js_uncomparable === true) {
		throw new Error(go2jsRuntimeErrorPrefix + "hash of unhashable type " + (key.type || ""));
	}

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

function go2jsWideAdd(left, right, typeName) {
	return go2jsWideBinary(left, right, (a, b) => a + b, (a, b) => a + b, typeName);
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
	return go2jsWideBinary(left, right, (a, b) => a - b, (a, b) => a - b, typeName);
}

function go2jsWideMul(left, right, typeName) {
	return go2jsWideBinary(left, right, (a, b) => a * b, (a, b) => a * b, typeName);
}

function go2jsWideQuo(left, right, typeName) {
	return go2jsWideBinary(left, right, (a, b) => {
		if (b === 0) {
			throw new go2jsNativeRangeError(go2jsRuntimeErrorPrefix + "integer divide by zero");
		}

		return Math.trunc(a / b);
	}, (a, b) => {
		if (b === 0n) {
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

		return a - Math.trunc(a / b) * b;
	}, (a, b) => {
		if (b === 0n) {
			throw new go2jsNativeRangeError(go2jsRuntimeErrorPrefix + "integer divide by zero");
		}

		return a % b;
	}, typeName);
}

// go2jsWideBitwise runs a marks operation over two whole numbers as digits,
// since a JavaScript operator on numbers touches only the low thirty two of
// them, and every mark of a wider number lives above where it reaches.
function go2jsWideBitwise(left, right, whole, typeName) {
	left = go2jsDurationOperand(left);
	right = go2jsDurationOperand(right);

	return go2jsWideWrap(whole(go2jsWide(left), go2jsWide(right)), typeName);
}

function go2jsWideAnd(left, right, typeName) {
	return go2jsWideBitwise(left, right, (a, b) => a & b, typeName);
}

function go2jsWideOr(left, right, typeName) {
	return go2jsWideBitwise(left, right, (a, b) => a | b, typeName);
}

function go2jsWideXor(left, right, typeName) {
	return go2jsWideBitwise(left, right, (a, b) => a ^ b, typeName);
}

function go2jsWideAndNot(left, right, typeName) {
	return go2jsWideBitwise(left, right, (a, b) => a & ~b, typeName);
}

function go2jsImul(left, right) {
	return Math.imul(go2jsDurationOperand(left), go2jsDurationOperand(right)) >>> 0;
}

// go2jsFitsThirtyTwoBits reports whether a whole number is one a thirty-two bit
// operation reaches every mark of.
function go2jsFitsThirtyTwoBits(value) {
	const number = Number(value);

	return Number.isInteger(number) && number >= -2147483648 && number <= 4294967295;
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

// go2jsWideFloat32 turns a whole number into the nearest number a float32
// holds, which is the same number the conversion gives in Go, brought down to
// the twenty four digits a float32 keeps.
function go2jsWideFloat32(value) {
	return Math.fround(Number(go2jsDurationOperand(value)));
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
	// keep by the time a shift of so many bits has been asked for. A number
	// held as digits to begin with is shifted as digits whatever the count is.
	if (go2jsIsWide(left) ||
		(count > 30 && Math.abs(Math.trunc(Number(left))) > go2jsSafeInteger / 2 ** count)) {
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

	if (Array.isArray(value)) {
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

	if (descriptors !== undefined) {
		for (const descriptor of descriptors) {
			if (descriptor && typeof descriptor.Name === "string") {
				fieldTypes[descriptor.Name] = descriptor.Type || "";
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
		const held = fieldType !== "" && value[key] === null ? go2jsTyped(null, fieldType) : value[key];
		const text = go2jsFormat(held, fieldType !== "" ? fieldType : null, null, null, plus, true);

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

			return go2jsPad(go2jsGoTypeName(original), parsed, false);
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

		if (descriptors !== undefined) {
			for (const descriptor of descriptors) {
				fieldTypes[descriptor.Name] = descriptor.Type || "";
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

			parts.push(go2jsFormatValue(verb, inner, fieldType === "" ? value[key] : go2jsTyped(value[key], fieldType), false, true));
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
	for (const ch of s) {
		const cp = ch.codePointAt(0);
		const mapped = go2jsCaseUpper[cp];
		out += String.fromCodePoint(mapped === undefined ? cp : mapped);
	}
	return out;
}

function go2jsStringsToLower(s) {
	let out = "";
	for (const ch of s) {
		const cp = ch.codePointAt(0);
		const mapped = go2jsCaseLower[cp];
		out += String.fromCodePoint(mapped === undefined ? cp : mapped);
	}
	return out;
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
	if (sep.length === 0) {
		return Array.from(s);
	}
	return s.split(sep);
}

function go2jsStringsJoin(parts, sep) {
	return parts.join(sep);
}

// go2jsStringsReplace puts a new piece of text in place of an old one, at most
// a given number of times counting from the front, the way Go does it: the
// search for the next old piece begins after the piece just written, so a
// replacement that itself holds the old piece is left alone. An empty old piece
// stands between every rune, and a count below zero means every one.
function go2jsStringsReplace(s, old, replacement, count) {
	if (old === replacement || count === 0) {
		return s;
	}

	if (old.length === 0) {
		const runes = Array.from(s);
		const limit = count < 0 ? runes.length + 1 : Math.min(count, runes.length + 1);
		const out = [];

		for (let i = 0; i < limit; i++) {
			out.push(replacement);

			if (i < runes.length) {
				out.push(runes[i]);
			}
		}
		for (let i = limit; i < runes.length; i++) {
			out.push(runes[i]);
		}

		return out.join("");
	}

	let result = "";
	let start = 0;
	let replaced = 0;

	while (count < 0 || replaced < count) {
		const index = s.indexOf(old, start);

		if (index < 0) {
			break;
		}

		result += s.slice(start, index) + replacement;
		start = index + old.length;
		replaced++;
	}

	return result + s.slice(start);
}

function go2jsStringsReplaceAll(s, old, replacement) {
	return go2jsStringsReplace(s, old, replacement, -1);
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
	return go2jsStrCompare(a, b);
}

// go2jsStringsToValidUTF8 replaces every byte that does not begin a valid UTF8
// sequence with the replacement, which is the way Go reads a string that may
// hold bytes that are not text. A valid sequence that spells the replacement
// character itself is left as it is, since it is a rune Go keeps.
function go2jsStringsToValidUTF8(s, replacement) {
	const bytes = go2jsStringToBytes(s);
	const fill = go2jsStringToBytes(replacement);
	const out = [];

	for (let i = 0; i < bytes.length;) {
		const width = go2jsUtf8Width(bytes, i);

		if (width === 0) {
			for (const item of fill) {
				out.push(item);
			}

			i += 1;
			continue;
		}

		for (let j = 0; j < width; j++) {
			out.push(bytes[i + j]);
		}

		i += width;
	}

	return go2jsBytesToString(out);
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

// go2jsHexValue reads one hex digit for what it stands for, and reports a
// character that is not one with a number no digit could be.
function go2jsHexValue(code) {
	if (code >= 0x30 && code <= 0x39) {
		return code - 0x30;
	}

	if (code >= 0x41 && code <= 0x46) {
		return code - 0x41 + 10;
	}

	if (code >= 0x61 && code <= 0x66) {
		return code - 0x61 + 10;
	}

	return -1;
}

// go2jsHexInvalidByte names the character that stopped a reading, written the
// way the language writes a printed rune.
function go2jsHexInvalidByte(byte) {
	const hex = byte.toString(16).toUpperCase().padStart(4, "0");
	const quoted = byte >= 0x20 && byte !== 0x7f ? " '" + String.fromCharCode(byte) + "'" : "";

	return go2jsNameError(new Error("encoding/hex: invalid byte: U+" + hex + quoted), "hex.InvalidByteError");
}

function go2jsHexDecodeString(s) {
	const src = go2jsStringify(s);
	const out = [];
	let index = 1;

	for (; index < src.length; index += 2) {
		const high = go2jsHexValue(src.charCodeAt(index - 1));
		const low = go2jsHexValue(src.charCodeAt(index));

		// A character that is not a hex digit stops the reading where it is,
		// keeping what was read before it, the way go hands back the bytes it
		// managed together with the fault.
		if (high < 0) {
			return [out, go2jsHexInvalidByte(src.charCodeAt(index - 1))];
		}

		if (low < 0) {
			return [out, go2jsHexInvalidByte(src.charCodeAt(index))];
		}

		out.push((high << 4) | low);
	}

	// A reading that stops with half a byte left is not a fault in any character
	// it has, so it is told apart from one that named a bad character.
	if (src.length % 2 === 1) {
		const last = src.charCodeAt(src.length - 1);

		if (go2jsHexValue(last) < 0) {
			return [out, go2jsHexInvalidByte(last)];
		}

		return [out, go2jsNameError(new Error("encoding/hex: odd length hex string"), "*errors.errorString")];
	}

	return [out, null];
}

function go2jsHexEncodedLen(n) {
	return Math.trunc(Number(n)) * 2;
}

function go2jsHexDecodedLen(n) {
	return Math.trunc(Number(n) / 2);
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
	if (substr.length === 0) {
		return Array.from(s).length + 1;
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

// go2jsStrconvSyntaxError, go2jsStrconvRangeError and go2jsStrconvBaseError build
// the errors Go raises while reading a value, worded as Go words them so a program
// that prints an error prints what Go prints.
function go2jsStrconvSyntaxError(name, text) {
	return new Error("strconv." + name + ": parsing " + JSON.stringify(go2jsStringify(text)) + ": invalid syntax");
}

function go2jsStrconvRangeError(name, text) {
	return new Error("strconv." + name + ": parsing " + JSON.stringify(go2jsStringify(text)) + ": value out of range");
}

function go2jsStrconvBaseError(name, text, base) {
	return new Error("strconv." + name + ": parsing " + JSON.stringify(go2jsStringify(text)) + ": invalid base " + base);
}

// go2jsStrconvIsRangeError reports whether an error is the one Go raises for a
// value too wide for the width it was read at.
function go2jsStrconvIsRangeError(err) {
	return err instanceof Error && /value out of range$/.test(err.message);
}

// go2jsStrconvDigitValue reads what a digit in a number means, in any base up to
// thirty-six, and answers -1 for a character that is not a digit.
function go2jsStrconvDigitValue(ch) {
	const code = ch.charCodeAt(0);

	if (code >= 48 && code <= 57) {
		return code - 48;
	}

	if (code >= 97 && code <= 122) {
		return code - 87;
	}

	if (code >= 65 && code <= 90) {
		return code - 55;
	}

	return -1;
}

// go2jsStrconvUnderscoreOK reports whether the underscores in a number are where
// Go lets them be: between two digits, and only when the number was written with
// no base of its own. Go's own rule is kept whole, so a number Go refuses is
// refused here too.
function go2jsStrconvUnderscoreOK(text) {
	let saw = "^";
	let i = 0;
	let body = text;

	if (body.length >= 1 && (body[0] === "-" || body[0] === "+")) {
		body = body.slice(1);
	}

	let hex = false;
	if (body.length >= 2 && body[0] === "0" && (body[1] === "b" || body[1] === "B" || body[1] === "o" || body[1] === "O" || body[1] === "x" || body[1] === "X")) {
		i = 2;
		saw = "0";
		hex = body[1] === "x" || body[1] === "X";
	}

	for (; i < body.length; i++) {
		const ch = body[i];

		if ((ch >= "0" && ch <= "9") || (hex && ((ch >= "a" && ch <= "f") || (ch >= "A" && ch <= "F")))) {
			saw = "0";
			continue;
		}

		if (ch === "_") {
			if (saw !== "0") {
				return false;
			}
			saw = "_";
			continue;
		}

		if (saw === "_") {
			return false;
		}

		saw = "!";
	}

	return saw !== "_";
}

// go2jsStrconvAtoi reads a signed decimal whole number, as Go's strconv.Atoi
// does, and answers the widest count that fits a sixty-four bit int along with
// the error Go raises for a number that does not.
function go2jsStrconvAtoi(s) {
	return go2jsStrconvParseSigned(go2jsStringify(s), 10, 64, "Atoi");
}

// go2jsStrconvParseUnsigned reads a whole number in any base from two to
// thirty-six, as Go's strconv.ParseUint does: Go's own base inference, prefix
// and underscore rules are followed, and a value too wide for the width asked is
// answered with the widest value that width holds along with the range error.
function go2jsStrconvParseUnsigned(text, base, bitSize) {
	if (text.length === 0) {
		return [0, go2jsStrconvSyntaxError("ParseUint", text)];
	}

	let body = text;
	const base0 = base === 0;
	let resolved = base;

	if (base >= 2 && base <= 36) {
		// A base named outright is used as it is.
	} else if (base0) {
		resolved = 10;

		if (body[0] === "0") {
			const second = body.length > 1 ? body[1].toLowerCase() : "";

			if (body.length >= 3 && second === "b") {
				resolved = 2;
				body = body.slice(2);
			} else if (body.length >= 3 && second === "o") {
				resolved = 8;
				body = body.slice(2);
			} else if (body.length >= 3 && second === "x") {
				resolved = 16;
				body = body.slice(2);
			} else {
				resolved = 8;
				body = body.slice(1);
			}
		}
	} else {
		return [0, go2jsStrconvBaseError("ParseUint", text, base)];
	}

	if (bitSize === 0 || bitSize === undefined || bitSize === null) {
		bitSize = 64;
	}

	const maxVal = (1n << BigInt(bitSize)) - 1n;
	const bigBase = BigInt(resolved);
	let n = 0n;
	let underscores = false;

	for (const ch of body) {
		if (ch === "_") {
			if (!base0) {
				return [0, go2jsStrconvSyntaxError("ParseUint", text)];
			}

			underscores = true;
			continue;
		}

		const digit = go2jsStrconvDigitValue(ch);

		if (digit < 0 || digit >= resolved) {
			return [0, go2jsStrconvSyntaxError("ParseUint", text)];
		}

		n = n * bigBase + BigInt(digit);

		if (n > maxVal) {
			return [maxVal, go2jsStrconvRangeError("ParseUint", text)];
		}
	}

	if (underscores && !go2jsStrconvUnderscoreOK(text)) {
		return [0, go2jsStrconvSyntaxError("ParseUint", text)];
	}

	return [n, null];
}

// go2jsStrconvParseSigned reads a whole number with a sign, as Go's strconv does,
// by reading the digits without the sign and then fitting them to the width once
// the sign is known.
function go2jsStrconvParseSigned(text, base, bitSize, name) {
	if (text.length === 0) {
		return [0, go2jsStrconvSyntaxError(name, text)];
	}

	let body = text;
	let neg = false;

	if (body[0] === "+") {
		body = body.slice(1);
	} else if (body[0] === "-") {
		neg = true;
		body = body.slice(1);
	}

	const size = bitSize === 0 || bitSize === undefined || bitSize === null ? 64 : bitSize;
	const [unsigned, err] = go2jsStrconvParseUnsigned(body, base, size);

	if (err !== null && !go2jsStrconvIsRangeError(err)) {
		if (/invalid base/.test(err.message)) {
			return [0, go2jsStrconvBaseError(name, text, base)];
		}

		return [0, go2jsStrconvSyntaxError(name, text)];
	}

	const cutoff = 1n << BigInt(size - 1);

	if (!neg && unsigned >= cutoff) {
		return [cutoff - 1n, go2jsStrconvRangeError(name, text)];
	}

	if (neg && unsigned > cutoff) {
		return [-cutoff, go2jsStrconvRangeError(name, text)];
	}

	return [neg ? -unsigned : unsigned, null];
}

function go2jsStrconvParseInt(s, base, bitSize) {
	return go2jsStrconvParseSigned(go2jsStringify(s), base, bitSize, "ParseInt");
}

function go2jsStrconvParseUint(s, base, bitSize) {
	return go2jsStrconvParseUnsigned(go2jsStringify(s), base, bitSize);
}

// go2jsStrconvSpecial reads the names Go accepts for the values that are not
// finite: an infinity, with or without a sign, and a NaN, which Go takes only
// without a sign. It answers the value and how many characters it read, or null.
function go2jsStrconvSpecial(s) {
	if (s.length === 0) {
		return null;
	}

	let i = 0;
	let sign = 1;
	let hadSign = false;

	if (s[0] === "+" || s[0] === "-") {
		hadSign = true;
		if (s[0] === "-") {
			sign = -1;
		}
		i = 1;
	}

	if (i >= s.length) {
		return null;
	}

	const first = s[i].toLowerCase();

	if (first === "i") {
		const rest = s.slice(i).toLowerCase();
		const target = "infinity";
		let n = 0;

		while (n < rest.length && n < target.length && rest[n] === target[n]) {
			n++;
		}

		if (n > 3 && n < 8) {
			n = 3;
		}

		if (n === 3 || n === 8) {
			return [sign * Infinity, i + n];
		}

		return null;
	}

	if (first === "n" && !hadSign && s.slice(i, i + 3).toLowerCase() === "nan") {
		return [NaN, i + 3];
	}

	return null;
}

// go2jsStrconvReadFloat reads the digits of a floating point number the way Go's
// strconv does, keeping the mantissa as a whole number of its own kind so that
// even the widest hexadecimal literal keeps every digit. It answers where the
// number ended, whether it was written in hexadecimal, and the parts the value is
// built from.
function go2jsStrconvReadFloat(s) {
	let underscores = false;
	let i = 0;
	let neg = false;

	if (i >= s.length) {
		return { ok: false, n: 0 };
	}

	if (s[i] === "+") {
		i++;
	} else if (s[i] === "-") {
		neg = true;
		i++;
	}

	let base = 10;
	let maxMantDigits = 19;
	let expChar = "e";
	let hex = false;

	if (i + 2 < s.length && s[i] === "0" && (s[i + 1] === "x" || s[i + 1] === "X")) {
		base = 16;
		maxMantDigits = 16;
		i += 2;
		expChar = "p";
		hex = true;
	}

	let sawdot = false;
	let sawdigits = false;
	let nd = 0;
	let ndMant = 0;
	let dp = 0;
	let mantissa = 0n;
	let trunc = false;

	loop: for (; i < s.length; i++) {
		const c = s[i];

		if (c === "_") {
			underscores = true;
			continue;
		}

		if (c === ".") {
			if (sawdot) {
				break loop;
			}

			sawdot = true;
			dp = nd;
			continue;
		}

		if (c >= "0" && c <= "9") {
			sawdigits = true;

			if (c === "0" && nd === 0) {
				dp--;
				continue;
			}

			nd++;

			if (ndMant < maxMantDigits) {
				mantissa = mantissa * BigInt(base) + BigInt(c.charCodeAt(0) - 48);
				ndMant++;
			} else if (c !== "0") {
				trunc = true;
			}

			continue;
		}

		if (base === 16 && ((c >= "a" && c <= "f") || (c >= "A" && c <= "F"))) {
			sawdigits = true;
			nd++;

			// Every hexadecimal digit is kept, however many there are, so that a
			// literal with more bits than a double holds is rounded from all of it.
			mantissa = mantissa * 16n + BigInt(go2jsStrconvDigitValue(c));
			ndMant++;

			continue;
		}

		break;
	}

	if (!sawdigits) {
		return { ok: false, n: i };
	}

	if (!sawdot) {
		dp = nd;
	}

	if (base === 16) {
		dp *= 4;
		ndMant *= 4;
	}

	if (i < s.length && s[i].toLowerCase() === expChar) {
		i++;

		if (i >= s.length) {
			return { ok: false, n: i };
		}

		let esign = 1;

		if (s[i] === "+") {
			i++;
		} else if (s[i] === "-") {
			i++;
			esign = -1;
		}

		if (i >= s.length || s[i] < "0" || s[i] > "9") {
			return { ok: false, n: i };
		}

		let e = 0;

		for (; i < s.length && ((s[i] >= "0" && s[i] <= "9") || s[i] === "_"); i++) {
			if (s[i] === "_") {
				underscores = true;
				continue;
			}

			if (e < 10000) {
				e = e * 10 + (s.charCodeAt(i) - 48);
			}
		}

		dp += e * esign;
	} else if (base === 16) {
		return { ok: false, n: i };
	}

	let exp = 0;

	if (mantissa !== 0n) {
		exp = dp - ndMant;
	}

	if (underscores && !go2jsStrconvUnderscoreOK(s.slice(0, i))) {
		return { ok: false, n: i };
	}

	return { ok: true, n: i, hex: hex, neg: neg, mantissa: mantissa, exp: exp, trunc: trunc };
}

// go2jsStrconvScaleFloat rounds a value written as a whole mantissa times a
// power of two to the nearest value of the width asked, rounding to even when
// the value sits exactly between two of them. The mantissa is kept whole so that
// even a hexadecimal literal with more bits than a double holds is rounded once,
// from all of it, the way Go rounds it.
function go2jsStrconvScaleFloat(mantissa, exp2, bits32) {
	if (mantissa === 0n) {
		return 0;
	}

	const mantbits = bits32 ? 23 : 52;
	const bias = bits32 ? 127 : 1023;
	const minNormal = 1 - bias;
	const minSub = minNormal - mantbits;

	const n = mantissa.toString(2).length;
	const e = exp2 + n - 1;

	if (e > bias) {
		return Infinity;
	}

	let significand;

	if (e >= minNormal) {
		const shift = n - 1 - mantbits;
		const unitExp = e - mantbits;

		if (shift <= 0) {
			significand = mantissa << BigInt(-shift);
		} else {
			significand = go2jsStrconvRoundBits(mantissa, shift);
		}

		return Number(significand) * (2 ** unitExp);
	}

	// A value below the smallest normal number is written in units of the
	// smallest positive value there is, and any bits past that unit are dropped.
	const shift = minSub - exp2;

	if (shift > n) {
		return 0;
	}

	if (shift <= 0) {
		significand = mantissa << BigInt(-shift);
	} else {
		significand = go2jsStrconvRoundBits(mantissa, shift);
	}

	return Number(significand) * (2 ** minSub);
}

// go2jsStrconvRoundBits drops the low bits of a mantissa, rounding the rest to
// the nearest whole number and to even when it is exactly between two of them.
function go2jsStrconvRoundBits(mantissa, shift) {
	const q = mantissa >> BigInt(shift);
	const rem = mantissa - (q << BigInt(shift));
	const half = 1n << BigInt(shift - 1);

	if (rem > half || (rem === half && (q & 1n) === 1n)) {
		return q + 1n;
	}

	return q;
}

// go2jsStrconvAtofHex rounds a hexadecimal floating point number to the nearest
// value of the width asked.
function go2jsStrconvAtofHex(neg, mantissa, exp, bits32) {
	let value = go2jsStrconvScaleFloat(mantissa, exp, bits32);

	if (neg) {
		value = -value;
	}

	// A value that reached infinity is one Go reports as out of range; a value
	// that sank to zero is not, and is answered as zero without an error.
	return [value, !Number.isFinite(value)];
}

function go2jsStrconvParseFloat(s, bitSize) {
	const text = go2jsStringify(s);
	const bits32 = bitSize === 32;

	const special = go2jsStrconvSpecial(text);

	if (special !== null) {
		if (special[1] !== text.length) {
			return [0, go2jsStrconvSyntaxError("ParseFloat", text)];
		}

		return [bits32 ? Math.fround(special[0]) : special[0], null];
	}

	const parsed = go2jsStrconvReadFloat(text);

	if (!parsed.ok || parsed.n !== text.length) {
		return [0, go2jsStrconvSyntaxError("ParseFloat", text)];
	}

	if (parsed.hex) {
		const [value, overflow] = go2jsStrconvAtofHex(parsed.neg, parsed.mantissa, parsed.exp, bits32);

		if (overflow) {
			return [value, go2jsStrconvRangeError("ParseFloat", text)];
		}

		return [value, null];
	}

	// JavaScript reads a decimal floating point number with the same rounding Go
	// does, so once the underscores Go allows are gone the text can be handed to
	// it whole.
	const value = Number(text.replace(/_/g, ""));

	if (!Number.isFinite(value)) {
		return [value, go2jsStrconvRangeError("ParseFloat", text)];
	}

	if (bits32) {
		const rounded = Math.fround(value);

		if (!Number.isFinite(rounded)) {
			return [rounded, go2jsStrconvRangeError("ParseFloat", text)];
		}

		return [rounded, null];
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
	base = base === undefined || base === 0 ? 10 : base;

	// A count wider than a double holds exactly keeps its own kind of whole
	// number, and Go writes the digits of that number, not of a rounded one.
	if (typeof value === "bigint") {
		return value.toString(base);
	}

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

function go2jsStrconvQuoteToASCII(s) {
	return go2jsStrconvQuoteWith(s, true, false);
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

// go2jsStrconvUnhex turns a hexadecimal digit byte into the value it stands
// for, or -1 when the byte is not a digit.
function go2jsStrconvUnhex(code) {
	if (code >= 48 && code <= 57) {
		return code - 48;
	}

	if (code >= 97 && code <= 102) {
		return code - 97 + 10;
	}

	if (code >= 65 && code <= 70) {
		return code - 65 + 10;
	}

	return -1;
}

// go2jsStrconvSpellsValidUTF8 says whether every byte from one point up to
// another spells a run of valid UTF8.
function go2jsStrconvSpellsValidUTF8(bytes, from, to) {
	for (let i = from; i < to;) {
		const width = go2jsUtf8Width(bytes, i);

		if (width === 0 || i + width > to) {
			return false;
		}

		i += width;
	}

	return true;
}

// go2jsStrconvUnquoteCharBytes reads the rune that leads a quoted string of
// bytes, the way strconv.UnquoteChar does: a plain byte is handed back as it
// is, a backslash escape is told apart as the rune it stands for, and the
// point after it is handed back along with whether the rune took more than one
// byte to write. Bytes are read, not text, because a string may hold bytes that
// spell no rune at all.
function go2jsStrconvUnquoteCharBytes(bytes, at, quote) {
	const bad = () => [0, false, at, false];

	if (at >= bytes.length) {
		return bad();
	}

	const first = bytes[at];

	if (first === quote && (quote === 39 || quote === 34)) {
		return bad();
	}

	if (first >= 0x80) {
		const width = go2jsUtf8Width(bytes, at);

		if (width === 0) {
			return [0xfffd, true, at + 1, true];
		}

		return [go2jsUtf8Rune(bytes, at, width), true, at + width, true];
	}

	if (first !== 92) {
		return [first, false, at + 1, true];
	}

	if (at + 1 >= bytes.length) {
		return bad();
	}

	const esc = bytes[at + 1];
	let next = at + 2;
	let value = 0;
	let multibyte = false;

	switch (esc) {
	case 97:
		value = 7;
		break;
	case 98:
		value = 8;
		break;
	case 102:
		value = 12;
		break;
	case 110:
		value = 10;
		break;
	case 114:
		value = 13;
		break;
	case 116:
		value = 9;
		break;
	case 118:
		value = 11;
		break;
	case 120:
	case 117:
	case 85: {
		const size = esc === 120 ? 2 : esc === 117 ? 4 : 8;

		if (bytes.length - next < size) {
			return bad();
		}

		let v = 0;

		for (let i = 0; i < size; i++) {
			const digit = go2jsStrconvUnhex(bytes[next + i]);

			if (digit < 0) {
				return bad();
			}

			v = v * 16 + digit;
		}

		next += size;

		if (esc === 120) {
			value = v;
			break;
		}

		if (!go2jsValidRune(v)) {
			return bad();
		}

		value = v;
		multibyte = true;
		break;
	}
	case 48:
	case 49:
	case 50:
	case 51:
	case 52:
	case 53:
	case 54:
	case 55: {
		let v = esc - 48;

		if (bytes.length - next < 2) {
			return bad();
		}

		for (let i = 0; i < 2; i++) {
			const digit = bytes[next + i] - 48;

			if (digit < 0 || digit > 7) {
				return bad();
			}

			v = v * 8 + digit;
		}

		next += 2;

		if (v > 255) {
			return bad();
		}

		value = v;
		break;
	}
	case 92:
		value = 92;
		break;
	case 39:
	case 34:
		if (esc !== quote) {
			return bad();
		}

		value = esc;
		break;
	default:
		return bad();
	}

	return [value, multibyte, next, true];
}

// go2jsStrconvUnquoteBytes reads a quoted string at the front of a run of
// bytes and hands back the value, how many bytes it read and whether it read
// one. When asked to unescape it hands back the value with quotes and escapes
// taken off, and otherwise the quoted text just as it stands, which is what
// strconv.QuotedPrefix wants.
function go2jsStrconvUnquoteBytes(bytes, unescape) {
	const fail = () => [null, 0, false];

	if (bytes.length < 2) {
		return fail();
	}

	const quote = bytes[0];
	let end = -1;

	for (let i = 1; i < bytes.length; i++) {
		if (bytes[i] === quote) {
			end = i + 1;
			break;
		}
	}

	if (quote === 96) {
		if (end < 0) {
			return fail();
		}

		if (!unescape) {
			return [bytes.slice(0, end), end, true];
		}

		const out = [];

		for (let i = 1; i < end - 1; i++) {
			if (bytes[i] !== 13) {
				out.push(bytes[i]);
			}
		}

		return [out, end, true];
	}

	if (quote !== 34 && quote !== 39) {
		return fail();
	}

	if (end < 0) {
		return fail();
	}

	let clean = true;

	for (let i = 0; i < end; i++) {
		if (bytes[i] === 92 || bytes[i] === 10) {
			clean = false;
			break;
		}
	}

	if (clean) {
		let valid;

		if (quote === 34) {
			valid = go2jsStrconvSpellsValidUTF8(bytes, 1, end - 1);
		} else {
			const width = go2jsUtf8Width(bytes, 1);

			valid = width > 0 && width === end - 2;
		}

		if (valid) {
			if (unescape) {
				return [bytes.slice(1, end - 1), end, true];
			}

			return [bytes.slice(0, end), end, true];
		}
	}

	const out = [];
	let i = 1;

	while (i < bytes.length && bytes[i] !== quote) {
		const step = go2jsStrconvUnquoteCharBytes(bytes, i, quote);

		if (bytes[i] === 10 || !step[3]) {
			return fail();
		}

		i = step[2];

		if (unescape) {
			if (step[0] < 0x80 || !step[1]) {
				out.push(step[0] & 0xff);
			} else {
				for (const byte of go2jsTextEncoder.encode(String.fromCodePoint(step[0]))) {
					out.push(byte);
				}
			}
		}

		if (quote === 39) {
			break;
		}
	}

	if (!(i < bytes.length && bytes[i] === quote)) {
		return fail();
	}

	i += 1;

	if (unescape) {
		return [out, i, true];
	}

	return [bytes.slice(0, i), i, true];
}

function go2jsStrconvUnquote(s) {
	const bytes = go2jsStringToBytes(s);
	const result = go2jsStrconvUnquoteBytes(bytes, true);

	if (!result[2] || result[1] !== bytes.length) {
		return ["", new Error("invalid syntax")];
	}

	return [go2jsBytesToString(result[0]), null];
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

function go2jsStrconvFormatFloat(value, format, precision, bitSize) {
	let number = Number(value);

	format = go2jsStrconvFormatVerb(format);

	if (typeof value === "bigint") {
		return value.toString();
	}

	if (typeof value === "bigint") {
		return value.toString();
	}

	// A number that was asked about as a smaller kind of number is rounded to
	// that kind first, so the digits written are the digits it keeps and a value
	// past the range of that kind becomes an infinity.
	if (bitSize === 32) {
		number = Math.fround(number);
	}

	if (Number.isNaN(number)) {
		return "NaN";
	}

	if (!Number.isFinite(number)) {
		return number < 0 ? "-Inf" : "+Inf";
	}

	switch (format) {
	case "x":
	case "X":
		return go2jsStrconvFormatHex(number, format, precision, bitSize);
	case "b":
		return go2jsStrconvFormatBinary(number, bitSize);
	}

	return go2jsStrconvBigFtoa(number, format, precision, bitSize);
}

// go2jsStrconvBigFtoa writes a number in one of the decimal forms Go's strconv
// writes, by carrying the value as an exact multiple precision decimal and
// rounding it the way Go rounds. Every digit it prints is a digit Go prints,
// including the shortest form, the sign of a negative zero, and the tie that
// Go breaks to even.
function go2jsStrconvBigFtoa(number, format, precision, bitSize) {
	const flt = go2jsStrconvFloatInfo(bitSize);
	const bits = go2jsStrconvFloatBits(number, bitSize);
	const mb = BigInt(flt.mantbits);
	const eb = BigInt(flt.expbits);
	const neg = (bits & (1n << (mb + eb))) !== 0n;
	let exp = Number((bits >> mb) & ((1n << eb) - 1n));
	let mant = bits & ((1n << mb) - 1n);

	if (exp === 0) {
		exp++;
	} else {
		mant |= 1n << mb;
	}

	exp += flt.bias;
	const shortest = precision < 0;
	const digs = go2jsDecAssign(mant);

	go2jsDecShift(digs, exp - flt.mantbits);

	if (shortest) {
		go2jsDecRoundShortest(digs, mant, exp, bitSize);

		switch (format) {
		case "e":
		case "E":
			precision = Math.max(digs.nd - 1, 0);
			break;
		case "f":
			precision = Math.max(digs.nd - digs.dp, 0);
			break;
		case "g":
		case "G":
			precision = digs.nd;
			break;
		}
	} else {
		switch (format) {
		case "e":
		case "E":
			go2jsDecRound(digs, precision + 1);
			break;
		case "f":
			go2jsDecRound(digs, digs.dp + precision);
			break;
		case "g":
		case "G":
			if (precision === 0) {
				precision = 1;
			}

			go2jsDecRound(digs, precision);
			break;
		}
	}

	return go2jsStrconvFormatDigits(shortest, neg, digs, precision, format);
}

// go2jsDecAssign is Go's decimal Assign: the digits of an unsigned whole number
// with the point after the last one, and trailing zeros trimmed.
function go2jsDecAssign(value) {
	const str = value.toString();

	const a = { d: str.split(""), nd: str.length, dp: str.length, trunc: false };

	go2jsDecTrim(a);

	return a;
}

// go2jsDecTrim is Go's decimal trim: trailing zeros carry no meaning, and a
// number with no digits left keeps a point at zero.
function go2jsDecTrim(a) {
	while (a.nd > 0 && a.d[a.nd - 1] === "0") {
		a.nd--;
	}

	if (a.nd === 0) {
		a.dp = 0;
	}
}

// go2jsDecShift multiplies the decimal by a power of two: left for a positive
// shift, right for a negative one. The work is done exactly with whole numbers,
// since a decimal cannot hold 1/2 exactly but can hold every power of five.
function go2jsDecShift(a, k) {
	if (a.nd === 0) {
		return;
	}

	const digits = BigInt(a.d.slice(0, a.nd).join(""));
	const nd = a.nd;
	const dp = a.dp;
	let scaled;
	let scale;

	if (k > 0) {
		scaled = digits * (2n ** BigInt(k));
		scale = 0;
	} else {
		const m = -k;
		scaled = digits * (5n ** BigInt(m));
		scale = m;
	}

	let str = scaled.toString();
	const cap = 2000;

	if (str.length > cap) {
		if (/[1-9]/.test(str.slice(cap))) {
			a.trunc = true;
		}

		str = str.slice(0, cap);
	}

	a.d = str.split("");
	a.nd = str.length;
	a.dp = dp - nd - scale + str.length;
	go2jsDecTrim(a);
}

// go2jsDecRoundShortest is Go's roundShortest: it walks the decimal digits
// against the halfway points on either side of the number and stops at the
// shortest run that still names it, rounding to the nearer when both fit.
function go2jsDecRoundShortest(d, mant, exp, bitSize) {
	const flt = go2jsStrconvFloatInfo(bitSize);

	if (mant === 0n) {
		d.nd = 0;
		return;
	}

	const minexp = flt.bias + 1;

	if (exp > minexp && 332 * (d.dp - d.nd) >= 100 * (exp - flt.mantbits)) {
		return;
	}

	const upper = go2jsDecAssign(mant*2n + 1n);
	go2jsDecShift(upper, exp - flt.mantbits - 1);

	let mantlo;
	let explo;

	if (mant > (1n << BigInt(flt.mantbits)) || exp === minexp) {
		mantlo = mant - 1n;
		explo = exp;
	} else {
		mantlo = mant*2n - 1n;
		explo = exp - 1;
	}

	const lower = go2jsDecAssign(mantlo*2n + 1n);
	go2jsDecShift(lower, explo - flt.mantbits - 1);
	const inclusive = mant % 2n === 0n;
	let upperdelta = 0;

	for (let ui = 0; ; ui++) {
		const mi = ui - upper.dp + d.dp;

		if (mi >= d.nd) {
			break;
		}

		const li = ui - upper.dp + lower.dp;
		let l = "0";

		if (li >= 0 && li < lower.nd) {
			l = lower.d[li];
		}

		let m = "0";

		if (mi >= 0) {
			m = d.d[mi];
		}

		let u = "0";

		if (ui < upper.nd) {
			u = upper.d[ui];
		}

		const okdown = l !== m || (inclusive && li + 1 === lower.nd);

		if (upperdelta === 0 && m.charCodeAt(0) + 1 < u.charCodeAt(0)) {
			upperdelta = 2;
		} else if (upperdelta === 0 && m !== u) {
			upperdelta = 1;
		} else if (upperdelta === 1 && (m !== "9" || u !== "0")) {
			upperdelta = 2;
		}

		const okup = upperdelta > 0 && (inclusive || upperdelta > 1 || ui + 1 < upper.nd);

		if (okdown && okup) {
			go2jsDecRound(d, mi + 1);
			return;
		}

		if (okdown) {
			go2jsDecRoundDown(d, mi + 1);
			return;
		}

		if (okup) {
			go2jsDecRoundUp(d, mi + 1);
			return;
		}
	}
}

// go2jsDecShouldRoundUp decides whether to round up when keeping nd digits,
// and breaks an exact halfway to even.
function go2jsDecShouldRoundUp(a, nd) {
	if (nd < 0 || nd >= a.nd) {
		return false;
	}

	if (a.d[nd] === "5" && nd + 1 === a.nd) {
		if (a.trunc) {
			return true;
		}

		return nd > 0 && (a.d[nd-1].charCodeAt(0) - 48) % 2 !== 0;
	}

	return a.d[nd] >= "5";
}

// go2jsDecRound keeps nd digits.
function go2jsDecRound(a, nd) {
	if (nd < 0 || nd >= a.nd) {
		return;
	}

	if (go2jsDecShouldRoundUp(a, nd)) {
		go2jsDecRoundUp(a, nd);
	} else {
		go2jsDecRoundDown(a, nd);
	}
}

// go2jsDecRoundDown keeps the first nd digits.
function go2jsDecRoundDown(a, nd) {
	if (nd < 0 || nd >= a.nd) {
		return;
	}

	a.nd = nd;
	go2jsDecTrim(a);
}

// go2jsDecRoundUp keeps the first nd digits, adding one to the last.
function go2jsDecRoundUp(a, nd) {
	if (nd < 0 || nd >= a.nd) {
		return;
	}

	for (let i = nd - 1; i >= 0; i--) {
		if (a.d[i] !== "9") {
			a.d[i] = String.fromCharCode(a.d[i].charCodeAt(0) + 1);
			a.nd = i + 1;
			return;
		}
	}

	a.d = ["1"];
	a.nd = 1;
	a.dp++;
}

// go2jsStrconvFormatDigits arranges digits in the form the requested verb asks
// for, following Go's strconv.
function go2jsStrconvFormatDigits(shortest, neg, digs, precision, format) {
	if (format === "e" || format === "E") {
		return go2jsStrconvFmtE(neg, digs, precision, format);
	}

	if (format === "f") {
		return go2jsStrconvFmtF(neg, digs, precision);
	}

	let eprec = precision;

	if (eprec > digs.nd && digs.nd >= digs.dp) {
		eprec = digs.nd;
	}

	if (shortest) {
		eprec = 6;
	}

	const exp = digs.dp - 1;

	if (exp < -4 || exp >= eprec) {
		if (precision > digs.nd) {
			precision = digs.nd;
		}

		return go2jsStrconvFmtE(neg, digs, precision - 1, format === "G" ? "E" : "e");
	}

	if (precision > digs.dp) {
		precision = digs.nd;
	}

	return go2jsStrconvFmtF(neg, digs, Math.max(precision - digs.dp, 0));
}

// go2jsStrconvFmtE writes digits as -d.dddde±dd.
function go2jsStrconvFmtE(neg, digs, precision, format) {
	const d = digs.d;
	let out = neg ? "-" : "";

	out += digs.nd !== 0 ? d[0] : "0";

	if (precision > 0) {
		out += ".";

		let i = 1;
		const m = Math.min(digs.nd, precision + 1);

		if (i < m) {
			out += d.slice(i, m).join("");
			i = m;
		}

		for (; i <= precision; i++) {
			out += "0";
		}
	}

	out += format;

	let exp = digs.nd === 0 ? 0 : digs.dp - 1;

	out += exp < 0 ? "-" : "+";
	exp = Math.abs(exp);

	if (exp < 10) {
		out += "0" + String(exp);
	} else if (exp < 100) {
		out += String(exp);
	} else {
		out += String(exp);
	}

	return out;
}

// go2jsStrconvFmtF writes digits as -ddd.dddd.
function go2jsStrconvFmtF(neg, digs, precision) {
	const d = digs.d;
	let out = neg ? "-" : "";

	if (digs.dp > 0) {
		const m = Math.min(digs.nd, digs.dp);

		out += d.slice(0, m).join("");

		for (let i = m; i < digs.dp; i++) {
			out += "0";
		}
	} else {
		out += "0";
	}

	if (precision > 0) {
		out += ".";

		for (let i = 0; i < precision; i++) {
			const j = digs.dp + i;

			out += j >= 0 && j < digs.nd ? d[j] : "0";
		}
	}

	return out;
}

// go2jsStrconvFloatInfo is the shape of an IEEE754 number: how many bits hold
// the significand and the exponent, and how far the exponent is biased.
function go2jsStrconvFloatInfo(bitSize) {
	return bitSize === 32
		? { mantbits: 23, expbits: 8, bias: -127 }
		: { mantbits: 52, expbits: 11, bias: -1023 };
}

// go2jsStrconvFloatBits reads the raw bits a number keeps in the given width.
function go2jsStrconvFloatBits(number, bitSize) {
	const buf = new ArrayBuffer(bitSize === 32 ? 4 : 8);
	const view = new DataView(buf);

	if (bitSize === 32) {
		view.setFloat32(0, number, true);

		return BigInt(view.getUint32(0, true));
	}

	view.setFloat64(0, number, true);

	return view.getBigUint64(0, true);
}

// go2jsStrconvFormatBinary writes a number the way Go's 'b' format writes it:
// its binary significand in decimal, a p, and the power of two it is scaled by.
function go2jsStrconvFormatBinary(number, bitSize) {
	const info = go2jsStrconvFloatInfo(bitSize);
	const bits = go2jsStrconvFloatBits(number, bitSize);
	const signMask = 1n << BigInt(info.mantbits + info.expbits);
	const neg = (bits & signMask) !== 0n;
	let exp = Number((bits >> BigInt(info.mantbits)) & ((1n << BigInt(info.expbits)) - 1n));
	let mant = bits & ((1n << BigInt(info.mantbits)) - 1n);

	if (exp === (1 << info.expbits) - 1) {
		return "NaN";
	}

	if (exp === 0) {
		exp++;
	} else {
		mant |= 1n << BigInt(info.mantbits);
	}

	exp += info.bias;
	exp -= info.mantbits;

	return (neg ? "-" : "") + mant.toString() + "p" + (exp >= 0 ? "+" : "") + String(exp);
}

// go2jsStrconvFormatHex writes a number the way the hexadecimal forms of Go's
// strconv write it: the leading digit, the fraction, and a binary exponent of
// at least two digits.
function go2jsStrconvFormatHex(number, format, precision, bitSize) {
	const info = go2jsStrconvFloatInfo(bitSize);
	const bits = go2jsStrconvFloatBits(number, bitSize);
	const mb = BigInt(info.mantbits);
	const eb = BigInt(info.expbits);
	const signMask = 1n << (mb + eb);
	const neg = (bits & signMask) !== 0n;
	let exp = Number((bits >> mb) & ((1n << eb) - 1n));
	let mant = bits & ((1n << mb) - 1n);

	if (exp === (1 << info.expbits) - 1) {
		return "NaN";
	}

	if (exp === 0) {
		exp++;
	} else {
		mant |= 1n << mb;
	}

	exp += info.bias;

	if (mant === 0n) {
		exp = 0;
	}

	// Put the leading 1, if any, at bit 60.
	mant = (mant << (60n - mb)) & go2jsUint64Mask;
	while (mant !== 0n && (mant & (1n << 60n)) === 0n) {
		mant = (mant << 1n) & go2jsUint64Mask;
		exp--;
	}

	// Round when the caller asked for a fraction short enough to round.
	if (precision >= 0 && precision < 15) {
		const shift = BigInt(precision * 4);
		const extra = (mant << shift) & ((1n << 60n) - 1n);
		mant >>= 60n - shift;
		if ((extra | (mant & 1n)) > (1n << 59n)) {
			mant++;
		}
		mant = (mant << (60n - shift)) & go2jsUint64Mask;
		if ((mant & (1n << 61n)) !== 0n) {
			mant >>= 1n;
			exp++;
		}
	}

	const lower = format === "x";
	const hex = lower ? "0123456789abcdef" : "0123456789ABCDEF";
	let out = neg ? "-" : "";
	out += "0" + format + String((mant >> 60n) & 1n);

	mant = (mant << 4n) & go2jsUint64Mask;

	if (precision < 0 && mant !== 0n) {
		out += ".";

		while (mant !== 0n) {
			out += hex[Number((mant >> 60n) & 15n)];
			mant = (mant << 4n) & go2jsUint64Mask;
		}
	} else if (precision > 0) {
		out += ".";

		for (let i = 0; i < precision; i++) {
			out += hex[Number((mant >> 60n) & 15n)];
			mant = (mant << 4n) & go2jsUint64Mask;
		}
	}

	out += lower ? "p" : "P";
	out += exp < 0 ? "-" : "+";
	out += go2jsStrconvExponentDigits(Math.abs(exp));

	return out;
}

// go2jsStrconvExponentDigits writes the digits of an exponent, which Go writes
// with at least two digits.
function go2jsStrconvExponentDigits(value) {
	return value < 10 ? "0" + String(value) : String(value);
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
	const length = go2jsSortMethod(data, typeName, "Len");
	let items = null;
	let count;

	// A value that answers for its own length says how many items it holds,
	// which is what a reversed or wrapped interface needs; a plain slice says
	// it by being one.
	if (typeof length === "function") {
		count = Number(go2jsCallNow(length, null, []));
	} else {
		items = go2jsToArray(data);
		count = items.length;
	}

	for (let index = 1; index < count; index++) {
		const outOfOrder = typeof less === "function"
			? go2jsCallNow(less, null, [index, index - 1])
			: items[index] < items[index - 1];

		if (outOfOrder) {
			return false;
		}
	}

	return true;
}

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
	if (values === null || values === undefined) {
		return;
	}

	values.sort((a, b) => a - b);
}

function go2jsSortStrings(values) {
	if (values === null || values === undefined) {
		return;
	}

	values.sort((a, b) => go2jsStrCompare(a, b));
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

function go2jsUnicodeMapCase(value, mode) {
	const r = go2jsUnicodeCodePoint(value);
	const ch = go2jsUnicodeRune(value);

	if (ch === "") {
		return r;
	}

	let mapped;

	switch (mode) {
	case "lower":
		mapped = go2jsCaseLower[r];
		break;
	case "title":
		mapped = go2jsCaseTitle[r];
		break;
	default:
		mapped = go2jsCaseUpper[r];
		break;
	}

	return mapped === undefined ? r : mapped;
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

// go2jsUTF8RuneDecodeSize is how many bytes a rune is written as, which is what
// a decode answers as its size, and how wide an error rune sits where a lone
// surrogate or a number out of range is read.
function go2jsUTF8RuneDecodeSize(code) {
	if (code <= 0x7f) {
		return 1;
	}

	if (code <= 0x7ff) {
		return 2;
	}

	if (code <= 0xffff) {
		if (code >= 0xd800 && code <= 0xdfff) {
			return 1;
		}

		return 3;
	}

	if (code <= 0x10ffff) {
		return 4;
	}

	return 1;
}

// A rune is read out of a string by its first character, which is what a string
// of runes holds one rune at a time, and the size is how many bytes a rune of
// it is written as in UTF-8.
function go2jsUTF8DecodeRuneInString(s) {
	const text = String(s);

	if (text.length === 0) {
		return [0xfffd, 0];
	}

	const code = text.codePointAt(0);

	// A code unit of the surrogate range stands on its own only when it is a
	// byte that spells no rune, and Go reads it as the replacement character.
	if (code >= 0xd800 && code <= 0xdfff) {
		return [0xfffd, 1];
	}

	return [code, go2jsUTF8RuneDecodeSize(code)];
}

function go2jsUTF8DecodeLastRuneInString(s) {
	const text = String(s);

	if (text.length === 0) {
		return [0xfffd, 0];
	}

	// the runes of a string are read from its end, since the last one is the
	// last character, and a character of two code units is the one before it
	let at = text.length - 1;

	if (at > 0) {
		const unit = text.charCodeAt(at);

		if (unit >= 0xdc00 && unit <= 0xdfff) {
			at--;
		}
	}

	const code = text.codePointAt(at);

	// a lone high surrogate left at the front is a byte that spells no rune
	if (code >= 0xd800 && code <= 0xdfff) {
		return [0xfffd, 1];
	}

	return [code, go2jsUTF8RuneDecodeSize(code)];
}

// go2jsUTF8DecodeRune reads the rune that leads a byte slice, the way
// utf8.DecodeRune does: an empty slice gives the replacement character and no
// width, and a byte that does not begin a rune gives the replacement character
// and the width of one. Bytes are read one by one, since reading them as text
// would fold every byte that spells no rune into the replacement character and
// lose how wide it was.
function go2jsUTF8DecodeRune(value) {
	const bytes = Array.from(value, (item) => Number(item) & 255);

	if (bytes.length === 0) {
		return [0xfffd, 0];
	}

	const width = go2jsUtf8Width(bytes, 0);

	if (width === 0) {
		return [0xfffd, 1];
	}

	const rune = go2jsUtf8Rune(bytes, 0, width);

	if (width > 1 && rune === 0) {
		return [0xfffd, 1];
	}

	return [rune, width];
}

// go2jsUTF8DecodeLastRune reads the rune that ends a byte slice, stepping back
// over the bytes a rune may be written with before reading the one that leads
// them.
function go2jsUTF8DecodeLastRune(value) {
	const bytes = Array.from(value, (item) => Number(item) & 255);
	const end = bytes.length;

	if (end === 0) {
		return [0xfffd, 0];
	}

	let start = end - 1;

	if (bytes[start] < 0x80) {
		return [bytes[start], 1];
	}

	const limit = end - 4 < 0 ? 0 : end - 4;

	for (start--; start >= limit; start--) {
		if ((bytes[start] & 0xc0) !== 0x80) {
			break;
		}
	}

	if (start < 0) {
		start = 0;
	}

	const width = go2jsUtf8Width(bytes, start);

	if (width === 0 || start + width !== end) {
		return [0xfffd, 1];
	}

	const rune = go2jsUtf8Rune(bytes, start, width);

	if (width > 1 && rune === 0) {
		return [0xfffd, 1];
	}

	return [rune, width];
}

// go2jsUTF8AppendRune writes a rune onto the end of a byte slice, which is the
// rune written down as the bytes it stands for.
function go2jsUTF8AppendRune(bytes, value) {
	const start = bytes === null || bytes === undefined ? [] : Array.from(bytes);

	return Array.from(Buffer.concat([Buffer.from(start), Buffer.from(String.fromCodePoint(go2jsUnicodeCodePoint(value)), "utf8")]));
}

// go2jsUTF8EncodeRune writes a rune as the bytes it stands for into the start
// of a byte slice and says how many bytes it took, the way unicode/utf8 does:
// a rune out of range (a negative one, one past the last, or a surrogate) is
// written as the error rune, and if the slice is too small to hold it, the
// write reaches past it the way Go does.
function go2jsUTF8EncodeRune(p, value) {
	let r = go2jsUnicodeCodePoint(value);
	let size = 1;

	if (r < 0) {
		r = 0x110000;
	}

	if (r <= 0x7f) {
		size = 1;
	} else if (r <= 0x7ff) {
		size = 2;
	} else if (r <= 0xffff && !(r >= 0xd800 && r <= 0xdfff)) {
		size = 3;
	} else if (r > 0xffff && r <= 0x10ffff) {
		size = 4;
	} else {
		size = 3;
	}

	if (p === null || p === undefined || go2jsLen(p) < size) {
		go2jsPanic("index out of range");
	}

	if (size === 1) {
		p[0] = r;
	} else if (size === 2) {
		p[0] = 0xc0 | (r >> 6);
		p[1] = 0x80 | (r & 0x3f);
	} else if (size === 3 && r <= 0xffff && !(r >= 0xd800 && r <= 0xdfff)) {
		p[0] = 0xe0 | (r >> 12);
		p[1] = 0x80 | ((r >> 6) & 0x3f);
		p[2] = 0x80 | (r & 0x3f);
	} else if (size === 4) {
		p[0] = 0xf0 | (r >> 18);
		p[1] = 0x80 | ((r >> 12) & 0x3f);
		p[2] = 0x80 | ((r >> 6) & 0x3f);
		p[3] = 0x80 | (r & 0x3f);
	} else {
		p[0] = 0xef;
		p[1] = 0xbf;
		p[2] = 0xbd;
	}

	return size;
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
