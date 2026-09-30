package javascript

func runtimeSource() string {
	return `
const go2jsNativeMap = globalThis.Map;
const go2jsNativeSet = globalThis.Set;
const go2jsNativeDate = globalThis.Date;
const go2jsNativeTypeError = globalThis.TypeError;
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
    return typeof os === "object" && typeof os.hostname === "function" ? os.hostname() : "";
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

// go2jsOSStatObject builds the FileInfo a stat call answers with. The methods
// are the ones Go's os.FileInfo declares, so a program that asks for a field it
// does not know here is told the same way a Go program would be.
function go2jsOSStatObject(stats) {
    return [{
        Size: function() {
            return stats.size;
        },
        IsDir: function() {
            return stats.isDirectory();
        },
        Mode: function() {
            return stats.mode;
        },
        ModTime: function() {
            return stats.mtime;
        }
    }, null];
}

// go2jsOSLstat describes a path without following a symbolic link, which is
// what Node's lstatSync does. A path that is not one answers the same as stat,
// which is what Go reports for it.
function go2jsOSLstat(path) {
    try {
        return go2jsOSStatObject(require("fs").lstatSync(String(path)));
    } catch (err) {
        return [null, err];
    }
}

function go2jsOSStat(path) {
    try {
        return go2jsOSStatObject(require("fs").statSync(String(path)));
    } catch (err) {
        return [null, err];
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
        const result = reader.Read(buffer);
        const read = Number(result[0]) || 0;

        if (read > 0) {
            chunks.push(...buffer.slice(0, read));
        }

        if (result[1] !== null && result[1] !== undefined) {
            if (result[1] === go2jsIOEOF()) {
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

function go2jsHTTPHeader() {
    const values = {};

    return {
        Set: function(key, value) {
            values[String(key).toLowerCase()] = String(value);
        },
        Get: function(key) {
            return values[String(key).toLowerCase()] || "";
        }
    };
}

// go2jsHTTPNoBody is the reader a request with nothing in its body is given. It
// reports no content and never ends, so it is not a reader over anything: every
// read asks again and the answer is always no bytes.
function go2jsHTTPNoBody() {
    return {
        Read: function() {
            return 0;
        },
        Close: function() {},
        __go2js_no_body: true
    };
}

function go2jsHTTPNewRequest(method, url, body) {
    try {
        return [{
            Method: String(method),
            URL: {
                value: String(url),
                String: function() {
                    return this.value;
                }
            },
            Header: go2jsHTTPHeader(),
            Body: body
        }, null];
    } catch (err) {
        return [null, err];
    }
}

function go2jsHTTPNewServeMux() {
    const routes = [];

    return {
        HandleFunc: function(pattern, handler) {
            routes.push({
                pattern: String(pattern),
                handler: handler
            });
        },
        Handler: function(request) {
            const path = request && request.URL
                ? new URL(request.URL.String()).pathname
                : "";

            for (const route of routes) {
                if (route.pattern === path) {
                    return [route.handler, route.pattern];
                }
            }

            return [function() {}, ""];
        }
    };
}

function go2jsTimeDate(year, month, day, hour, minute, second, nanosecond, location) {
    const date = new Date(0);

    date.setUTCFullYear(Number(year), Number(month) - 1, Number(day));
    date.setUTCHours(
        Number(hour),
        Number(minute),
        Number(second),
        Math.floor(Number(nanosecond) / 1000000)
    );

    return go2jsTimeValue(date);
}

// go2jsTimeMonth wraps a month number so it prints as its Go name while still
// comparing and converting like the underlying integer.
function go2jsTimeMonth(month) {
    return {
        __go2js_month: month,
        valueOf: function() {
            return this.__go2js_month;
        },
        String: function() {
            return go2jsMonthName(this.__go2js_month);
        }
    };
}

function go2jsMonthName(month) {
    const names = [
        "January", "February", "March", "April", "May", "June",
        "July", "August", "September", "October", "November", "December"
    ];

    return names[month - 1] || "";
}

// go2jsTimeZero is the zero time, which is January 1 of year 1 at midnight UTC.
// The year is out of the range that Date.UTC maps straight onto, so the
// instant is given as the number of milliseconds before the epoch instead.
function go2jsTimeZero() {
    return go2jsTimeValue(new Date(-62135596800000));
}

function go2jsTimeValue(date) {
    const self = {
        value: date,
        Year: function() {
            return this.value.getUTCFullYear();
        },
        Month: function() {
            return go2jsTimeMonth(this.value.getUTCMonth() + 1);
        },
        Day: function() {
            return this.value.getUTCDate();
        },
        Hour: function() {
            return this.value.getUTCHours();
        },
        Minute: function() {
            return this.value.getUTCMinutes();
        },
        Second: function() {
            return this.value.getUTCSeconds();
        },
        Nanosecond: function() {
            return this.value.getUTCMilliseconds() * 1000000;
        },
        Clock: function() {
            return [this.Hour(), this.Minute(), this.Second()];
        },
        Unix: function() {
            return Math.floor(this.value.getTime() / 1000);
        },
        UnixNano: function() {
            return this.value.getTime() * 1000000;
        },
        IsZero: function() {
            return this.value.getTime() === 0;
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
            return go2jsTimeFormat(this.value, String(layout));
        },
        String: function() {
            return this.value.toISOString();
        },
        Add: function(duration) {
            const next = new Date(this.value.getTime() + go2jsDurationNanos(duration) / 1000000);
            return go2jsTimeValue(next);
        },
        Sub: function(other) {
            return go2jsDuration((this.value.getTime() - go2jsTimeDateOf(other).getTime()) * 1000000);
        }
    };

    return self;
}

function go2jsTimeDateOf(value) {
    return value !== null && value !== undefined && value.value instanceof Date ? value.value : value;
}

function go2jsTimeFormat(date, layout) {
    const pad = function(value, width) {
        return String(value).padStart(width, "0");
    };

    return layout
        .replace(/2006/g, String(date.getUTCFullYear()))
        .replace(/01/g, pad(date.getUTCMonth() + 1, 2))
        .replace(/02/g, pad(date.getUTCDate(), 2))
        .replace(/15/g, pad(date.getUTCHours(), 2))
        .replace(/04/g, pad(date.getUTCMinutes(), 2))
        .replace(/05/g, pad(date.getUTCSeconds(), 2));
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
	const from = go2jsFilepathClean(base).split("/").filter(part => part !== "");
	const to = go2jsFilepathClean(target).split("/").filter(part => part !== "");

	let shared = 0;

	while (shared < from.length && shared < to.length && from[shared] === to[shared]) {
		shared++;
	}

	const parts = [];

	for (let i = shared; i < from.length; i++) {
		parts.push("..");
	}

	for (let i = shared; i < to.length; i++) {
		parts.push(to[i]);
	}

	return parts.join("/");
}

function go2jsFilepathSplit(value) {
	const path = String(value);
	const index = path.lastIndexOf("/");

	if (index < 0) {
		return ["", path];
	}

	if (index === 0) {
		return ["/", path.slice(1)];
	}

	return [path.slice(0, index), path.slice(index + 1)];
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

	return [new RegExp(source).test(go2jsStringify(name)), null];
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


function go2jsJSONMarshal(value, fields, omitEmpty) {
    try {
        return [JSON.stringify(go2jsJSONEncode(value, fields, omitEmpty)), null];
    } catch (err) {
        return [null, err];
    }
}

function go2jsJSONEncode(value, fields, omitEmpty) {
    if (value === null || value === undefined) {
        return null;
    }

    if (Array.isArray(value)) {
        return value.map(item => go2jsJSONEncode(item, fields, omitEmpty));
    }

    if (value instanceof go2jsNativeMap) {
        const out = {};
        for (const [key, item] of value.entries()) {
            out[go2jsStringify(key)] = go2jsJSONEncode(item, fields, omitEmpty);
        }
        return out;
    }

    if (typeof value !== "object" || value instanceof Error) {
        return value;
    }

    if (fields) {
        const out = {};
        let written = false;

        for (const field of Object.keys(fields)) {
            if (!(field in value)) {
                continue;
            }

            const name = fields[field];
            const item = value[field];

            if (omitEmpty && omitEmpty.indexOf(name) >= 0 && go2jsJSONEmpty(item)) {
                continue;
            }

            out[name] = go2jsJSONEncode(item, null, null);
            written = true;
        }

        if (written) {
            return out;
        }
    }

    const out = {};

    for (const key of Object.keys(value)) {
        out[key] = go2jsJSONEncode(value[key], null, null);
    }

    return out;
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

function go2jsJSONUnmarshal(data, target, fields, destination) {
    try {
        const value = JSON.parse(
            typeof data === "string"
                ? data
                : new TextDecoder().decode(Uint8Array.from(data))
        );

        if (target === null || target === undefined) {
            return null;
        }

        if (destination !== null && destination !== undefined && destination.map === true) {
            go2jsJSONDecodeMap(value, target, destination.value);
            return null;
        }
        if (typeof target === "object") {
            if (value !== null && typeof value === "object") {
                go2jsJSONDecode(value, target, fields);
            }
            return null;
        }

        return null;
    } catch (err) {
        return err;
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

function go2jsJSONCoerceAny(value) {
    if (value === null || value === undefined) {
        return null;
    }

    if (Array.isArray(value)) {
        return value.map(go2jsJSONCoerceAny);
    }

    if (typeof value === "object") {
        const map = go2jsMakeMap();

        for (const key of Object.keys(value)) {
            go2jsMapSet(map, key, go2jsJSONCoerceAny(value[key]));
        }

        return map;
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

function go2jsJSONDecodeMap(value, target, kind) {
    if (value === null || typeof value !== "object" || Array.isArray(value)) {
        return;
    }

    const map = go2jsJSONTargetMap(target);

    if (map === null) {
        return;
    }

    for (const key of Object.keys(value)) {
        go2jsMapSet(map, go2jsJSONCoerceString(key), go2jsJSONCoerce(value[key], kind));
    }
}

function go2jsJSONDecode(value, target, fields) {
    if (value === null || typeof value !== "object" || Array.isArray(value)) {
        return;
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

            if (!matched && key in target) {
                name = key;
            }
        }

        if (!(name in target)) {
            continue;
        }

        target[name] = value[key];
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
        total += part.length;
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
        return -1;
    }

    return this.data.shift();
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
	if (this.data.length === 0) {
		return go2jsBytesReadStringResult("", null);
	}

	const needle = go2jsToArray(delim);
	const limit = needle.length;

	if (limit === 0) {
		const out = this.data.slice();
		this.data = [];
		return go2jsBytesReadStringResult(go2jsBytesToString(out), null);
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

	return go2jsBytesReadStringResult("", go2jsEOFError());
};

go2jsBytesBuffer.prototype.ReadBytes = function(delim) {
	const result = this.ReadString(delim);

	return result[0] === "" && result[1] !== null
		? [null, result[1]]
		: [go2jsStringToBytes(result[0]), null];
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
		target.Write(go2jsStringToBytes(go2jsBytesToString(this.data)));
	}

	this.data = [];

	return [written, null];
};

go2jsBytesBuffer.prototype.ReadFrom = function(source) {
	if (!source || typeof source.Read !== "function") {
		return [0, null];
	}

	let total = 0;

	for (;;) {
		const chunk = go2jsStringToBytes("");
		const result = source.Read(chunk);
		const value = result && result.value !== undefined ? result.value : result;
		const err = result && result.err !== undefined ? result.err : null;

		if (value && value.length > 0) {
			this.data = this.data.concat(Array.from(value));
			total += value.length;
		}

		if (err !== null || value === null || value.length === 0) {
			break;
		}
	}

	return [total, null];
};

function go2jsBytesReadStringResult(value, err) {
	return [value, err];
}

function go2jsIOReadAllResult(n, err) {
	return [null, err];
}

function go2jsEOFError() {
	return go2jsSentinelError("EOF");
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
        return value.slice();
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

		return (...args) => fn.apply(captured, args);
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

	return (...args) => fn.apply(captured, args);
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

	return fn(receiver, ...args);
}

function go2jsInvokeMethod(receiver, method, args) {
	if (receiver === null || receiver === undefined) {
		throw new TypeError("method call on nil receiver");
	}

	if (receiver.__go2js_interface === true) {
		return go2jsInterfaceCall(receiver, method, ...args);
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

	return fn.apply(receiver, args);
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
				return function(...args) { return fn(target.value, ...args); };
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
		[go2jsPointerGet]: get,
		[go2jsPointerSet]: set
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
		return ctor.prototype[name].apply(null, args);
	}

	return receiver[name].apply(receiver, args);
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
	go2jsInitializers.push(assign);
}

function go2jsRunInitializers() {
	while (go2jsInitializers.length > 0) {
		go2jsInitializers.shift()();
	}
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
const go2jsInterfaces = Object.create(null);
const go2jsStructFormats = Object.create(null);
const go2jsTypeNames = Object.create(null);
const go2jsMethodSets = Object.create(null);
const go2jsStructFields = Object.create(null);
const go2jsTypeKinds = Object.create(null);

function go2jsRegisterTypeName(constructor, name, methods, fields, kind) {
	go2jsTypeNames[name] = constructor;

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

	if (value !== null && value !== undefined && value.__go2js_typed === true) {
		return value.type;
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
		return go2jsFormat(value[name]());
	}

	const ctor = value.constructor;

	if (ctor && typeof ctor.name === "string" && ctor.name !== "Object") {
		const registered = go2jsLookupTypeName(ctor.name);

		if (registered !== undefined && typeof go2jsMethodTable[registered + "." + name] === "function") {
			return go2jsFormat(go2jsMethodTable[registered + "." + name](value));
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
			return (...args) => go2jsMethodTable[registered + "." + name](receiver, ...args);
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

			if (!go2jsEqual(a[key], b[key])) {
				return false;
			}
		}

		return true;
	}

	return a === b;
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

function go2jsInterfaceCall(value, method, ...args) {
	if (value === null || value === undefined) {
		throw new TypeError(go2jsRuntimeErrorPrefix + "invalid memory address or nil pointer dereference");
	}

	if (value instanceof Error && method === "Error") {
		return value.message;
	}

	if (value.__go2js_interface !== true) {
		if (value.__go2js_pointer === true) {
			return go2jsInterfaceCall(value[go2jsPointerGet](), method, ...args);
		}

		if (typeof value[method] === "function") {
			return value[method](...args);
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

		if (typeof registered === "function") {
			return registered(target, ...args);
		}
	}

	let fn = target[method];
	let bound = false;

	if (typeof fn !== "function" && typeof value.type === "string") {
		fn = go2jsMethodTable[value.type + "." + method];
		bound = true;
	}

	if (typeof fn !== "function") {
		throw new TypeError("interface method " + method + " is not implemented");
	}

	if (bound) {
		return fn(target, ...args);
	}

	return fn.call(target, ...args);
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
	if (Array.isArray(value)) {
		return value.slice();
	}

	if (value instanceof go2jsNativeMap) {
		return new go2jsNativeMap(value);
	}

	if (value && typeof value === "object") {
		return Object.assign({}, value);
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
		throw new RangeError(go2jsRuntimeErrorPrefix + "index out of range [" + position + "]");
	}

	if (position >= go2jsLen(value)) {
		throw new RangeError(go2jsRuntimeErrorPrefix + "index out of range [" + position + "] with length " + go2jsLen(value));
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
		throw new RangeError(go2jsRuntimeErrorPrefix + "index out of range [" + position + "]");
	}

	const length = go2jsLen(value);

	if (position >= length) {
		throw new RangeError(go2jsRuntimeErrorPrefix + "index out of range [" + position + "] with length " + length);
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
function go2jsDivide(left, right) {
	if (right === 0) {
		throw new RangeError(go2jsRuntimeErrorPrefix + "integer divide by zero");
	}

	return Math.trunc(left / right);
}

// go2jsShl shifts a number the way Go does, where the count is however many
// bits a program asks for. A JavaScript shift is thirty-two bits wide whatever
// the count says, so a shift of sixty-two bits is a shift of thirty in
// JavaScript and the wrong answer twice over.
function go2jsShl(left, right) {
	const count = Math.trunc(Number(right));

	if (count < 0) {
		return go2jsShr(left, -count);
	}

	return Math.trunc(Number(left)) * 2 ** count;
}

// go2jsShr brings a number down by so many bits. A shift to the right on a
// signed number brings the sign down with it and one on an unsigned number
// brings down zeros, and the number itself says which it is: a negative one
// keeps its sign the whole way down.
function go2jsShr(left, right) {
	const count = Math.trunc(Number(right));
	const number = Math.trunc(Number(left));

	if (count <= 0) {
		return number;
	}

	return number < 0 ? -Math.floor(-number / 2 ** count) : Math.floor(number / 2 ** count);
}

// go2jsIntWrap keeps a whole number inside the width of the type it is for,
// which is what a Go variable of that type does without being asked. A number
// too wide for its type is not an error in Go, it is the number that fits, so
// an int8 one past its largest is the smallest int8 and a uint8 one below its
// smallest is the largest uint8.
function go2jsIntWrap(value, typeName) {
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
	if (right === 0) {
		throw new RangeError(go2jsRuntimeErrorPrefix + "integer divide by zero");
	}

	return left - Math.trunc(left / right) * right;
}

function go2jsStringByteAt(value, index) {
	if (typeof value !== "string") {
		return value[Math.trunc(index)];
	}

	return value.charCodeAt(Math.trunc(index));
}

function go2jsZeroArray(length) {
	const out = new Array(length);

	for (let index = 0; index < length; index++) {
		out[index] = 0;
	}

	Object.defineProperty(out, "__go2js_cap", {
		value: length,
		writable: true,
		configurable: true
	});

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
	if (!(value instanceof go2jsNativeTypeError)) {
		return null;
	}

	const message = String(value.message === undefined ? "" : value.message);

	if (/^Cannot (read|set) propert(?:y|ies) of (null|undefined)/.test(message)) {
		return go2jsRuntimeErrorPrefix + "invalid memory address or nil pointer dereference";
	}

	if (/^\w+ is not a function$/.test(message) || /of undefined$/.test(message)) {
		return go2jsRuntimeErrorPrefix + "invalid memory address or nil pointer dereference";
	}

	if (message.startsWith("index out of range") || message.startsWith("integer divide by zero")) {
		return go2jsRuntimeErrorPrefix + message;
	}

	return null;
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
	if (typeof value === "object") {
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
function go2jsNumericValue(value) {
	if (value !== null && value !== undefined && value.__go2js_pointer === true) {
		return go2jsNumericValue(go2jsDeref(value));
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
	if (value.__go2js_interface === true && value.value !== value) {
		if (raw !== true && typeof value.type === "string") {
			const errorer = go2jsMethodTable[value.type + ".Error"];

			if (typeof errorer === "function") {
				return errorer(value.value);
			}

			const stringer = go2jsMethodTable[value.type + ".String"];

			if (typeof stringer === "function") {
				return stringer(value.value);
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
		return value.String();
	}

	if (raw !== true && typeof value.Error === "function") {
		return value.Error();
	}

	const parts = [];
	const ctor = value.constructor;
	const fields = ctor && typeof ctor.name === "string" ? go2jsStructFormats[ctor.name] : null;

	for (const key of Object.keys(value)) {
		if (fields && typeof fields[key] === "string") {
			const formatter = go2jsMethodTable[fields[key]];

			if (typeof formatter === "function") {
				parts.push(formatter(value[key]));
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

		text += go2jsFormat(values[i]);
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
		const written = target.Write(go2jsStringToBytes(text));
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

			target.write(go2jsBytesToString(args[0]));
			return go2jsToArray(args[0]).length;
		};
	}

	if (typeof target[method] === "function") {
		const own = target[method];
		return function(...args) {
			return own.apply(target, args);
		};
	}

	const type = writer !== null && writer !== undefined && typeof writer.type === "string" ? writer.type : "";

	if (type !== "") {
		const registered = go2jsMethodTable[type + "." + method];
		if (typeof registered === "function") {
			return function(...args) {
				return registered.call(target, ...args);
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
			result += "%!w(" + go2jsDynamicTypeName(arg, go2jsTypedType(arg)) + "=" +
				go2jsFormatValue("v", spec.replace(/w$/, "v"), arg, true) + ")";
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

function go2jsGoSyntax(value) {
	if (value === null || value === undefined) {
		return "<nil>";
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

		return "map[" + go2jsGoTypeName(value) + "]" +
			"{" + entries.map(([key, item]) => go2jsGoSyntax(key) + ":" + go2jsGoSyntax(item)).join(", ") + "}";
	}

	if (Array.isArray(value)) {
		return "[]" + go2jsGoTypeName(value) + "{" + value.map(item => go2jsGoSyntax(item)).join(", ") + "}";
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
				return go2jsBytesToString(registered(operand.value));
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
			format = (...args) => method(receiver.value, ...args);
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
		format.call(receiver, state, rune);
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
	const boxed = operand !== null && operand !== undefined && operand.__go2js_interface === true;
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

	// A number too big for a double is held whole, and every verb over it reads
	// the digits off that whole number rather than off a rounded one, so it is
	// answered before anything tries to treat it as a plain number.
	if (typeof value === "bigint") {
		return go2jsBigintFormat(value, {verb: verb, flags: flags, precision: precision, width: parsed.width}, tagName, kind);
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
	// %T and %p answer for the whole operand, so they never reach the elements.
	const compound = verb === "T" || verb === "p" || (isBytes && go2jsVerbInSet(verb, "qsxX"))
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
				text = go2jsGoSyntax(value);
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

	if (verb === "c" || verb === "q" || verb === "s" || verb === "U") {
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
	} else if (verb === "e" || verb === "f" || verb === "g") {
		out = value.toString(10);
	} else {
		return "%!" + verb + "(" + (kind === "rune" ? "int32" : accepted) + "=" + value.toString(10) + ")";
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
				target.Write(text);
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

function go2jsStrconvItoa(value) {
	return String(Math.trunc(value));
}

function go2jsStrconvAtoi(s) {
	const value = parseInt(s, 10);
	if (Number.isNaN(value)) {
		return [0, new Error("strconv.Atoi: parsing " + JSON.stringify(s) + ": invalid syntax")];
	}
	return [value, null];
}

function go2jsStrconvParseInt(s, base, bitSize) {
	const value = parseInt(s, base || 10);
	if (Number.isNaN(value)) {
		return [0, new Error("strconv.ParseInt: parsing " + JSON.stringify(s) + ": invalid syntax")];
	}
	return [value, null];
}

function go2jsStrconvParseUint(s, base, bitSize) {
	const text = go2jsStringify(s).trim();
	const negative = text.startsWith("-");

	if (negative) {
		return [0, new Error('strconv.ParseUint: parsing "' + text + '": invalid syntax')];
	}

	const value = parseInt(text, base || 10);

	if (Number.isNaN(value) || value < 0) {
		return [0, new Error('strconv.ParseUint: parsing "' + text + '": invalid syntax')];
	}

	return [value, null];
}

function go2jsStrconvParseFloat(s, bitSize) {
	const value = parseFloat(s);
	if (Number.isNaN(value)) {
		return [0, new Error("strconv.ParseFloat: parsing " + JSON.stringify(s) + ": invalid syntax")];
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
		return ["", new Error("strconv.Unquote: invalid syntax")];
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

function go2jsStrconvFormatFloat(value, format, precision, bitSize) {
	const number = Number(value);

	format = go2jsStrconvFormatVerb(format);

	if (Number.isNaN(number)) {
		return "NaN";
	}

	if (!Number.isFinite(number)) {
		return number < 0 ? "-Inf" : "+Inf";
	}

	switch (format) {
	case "f":
		return precision >= 0 ? number.toFixed(precision) : String(number);
	case "e":
		return number.toExponential(Math.max(0, precision)).replace("E", "e");
	case "E":
		return number.toExponential(Math.max(0, precision)).replace("e", "E");
	case "g":
	case "G":
		return String(number);
	default:
		return String(number);
	}
}

function go2jsStrconvFormatBool(value) {
	return value ? "true" : "false";
}

function go2jsStrconvAppendInt(dst, value, base) {
	const text = Math.trunc(value).toString(base || 10);
	if (Array.isArray(dst)) {
		return dst.concat(Array.from(text, (ch) => ch.charCodeAt(0)));
	}
	return String(dst) + text;
}

function go2jsStrconvAppendFloat(dst, value, format, precision, bitSize) {
	const text = go2jsStrconvFormatFloat(value, format, precision, bitSize);
	if (Array.isArray(dst)) {
		return dst.concat(Array.from(text, (ch) => ch.charCodeAt(0)));
	}
	return String(dst) + text;
}

function go2jsStrconvAppendBool(dst, value) {
	const text = value ? "true" : "false";
	if (Array.isArray(dst)) {
		return dst.concat(Array.from(text, (ch) => ch.charCodeAt(0)));
	}
	return String(dst) + text;
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
			return function(...args) { return fn(target, ...args); };
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
		return fn();
	}

	return go2jsLen(this.data);
};

go2jsSortReverse.prototype.Less = function(i, j) {
	const fn = go2jsSortMethod(this.data, this.typeName, "Less");

	if (typeof fn === "function") {
		return fn(j, i);
	}

	return go2jsCompareValues(this.data[j], this.data[i]) < 0;
};

go2jsSortReverse.prototype.Swap = function(i, j) {
	const fn = go2jsSortMethod(this.data, this.typeName, "Swap");

	if (typeof fn === "function") {
		fn(i, j);
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

			while (current > 0 && less(current, current - 1)) {
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
			? less(index, index - 1)
			: items[index] < items[index - 1];

		if (outOfOrder) {
			return false;
		}
	}

	return true;
}

function go2jsSortSearch(values, target, less) {
	const items = go2jsToArray(values);
	let low = 0;
	let high = items.length;

	while (low < high) {
		const mid = Math.floor((low + high) / 2);

		if (less(target, items[mid])) {
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
		if (less(items[i], items[i - 1])) {
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


`
}
