"use strict";
const go2jsNativeMap = globalThis.Map;
const go2jsNativeSet = globalThis.Set;

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

function go2jsStringToBytes(value) {
    return Array.from(new TextEncoder().encode(value));
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

function go2jsOSStat(path) {
    try {
        const stats = require("fs").statSync(String(path));

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
                return [0, "EOF"];
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

        if (result[1] !== null && result[1] !== undefined) {
            if (result[1] === "EOF") {
                break;
            }
            return [null, result[1]];
        }

        if (result[0] <= 0) {
            break;
        }

        chunks.push(...buffer.slice(0, result[0]));
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

    return {
        pattern: source,
        MatchString(value) {
            return new RegExp(source).test(String(value));
        },
        Find(value) {
            const match = new RegExp(source).exec(go2jsBytesToString(value));
            return match === null ? null : go2jsStringToBytes(match[0]);
        },
        FindString(value) {
            const match = new RegExp(source).exec(go2jsBytesToString(value));
            return match === null ? "" : match[0];
        },
        FindStringIndex(value) {
            const match = new RegExp(source).exec(go2jsBytesToString(value));
            return match === null ? null : [match.index, match.index + match[0].length];
        },
        FindAllString(value, limit) {
            return go2jsRegexpFindAllString(source, value, limit);
        },
        FindAllStringIndex(value, limit) {
            const regex = new RegExp(source, "g");
            const text = go2jsBytesToString(value);
            const out = [];

            for (;;) {
                const match = regex.exec(text);

                if (match === null) {
                    break;
                }

                out.push([match.index, match.index + match[0].length]);

                if (limit >= 0 && out.length >= limit) {
                    break;
                }
            }

            return out;
        },
        ReplaceAllString(value, replacement) {
            return go2jsRegexpReplaceAllString(source, value, replacement);
        },
        ReplaceAll(value, replacement) {
            return go2jsBytesToString(value).replace(new RegExp(source, "g"), go2jsRegexpExpand(replacement));
        },
        Replace(value, replacement) {
            return go2jsStringify(value).replace(new RegExp(source), go2jsRegexpExpand(replacement));
        },
        ReplaceLiteral(value, replacement) {
            return go2jsStringify(value).replace(new RegExp(source), () => go2jsStringify(replacement));
        },
        ReplaceAllLiteral(value, replacement) {
            return go2jsStringify(value).replace(new RegExp(source, "g"), () => go2jsStringify(replacement));
        },
        Split(value, limit) {
            return go2jsRegexpSplit(source, value, limit);
        },
        String() {
            return source;
        },
        NumSubexp() {
            return new RegExp(source + "|").exec("").length - 1;
        }
    };
}

function go2jsRegexpMustCompile(pattern) {
    return go2jsRegexpNew(pattern);
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

function go2jsFilepathMatch(pattern, name) {
	const source = String(pattern)
		.split("")
		.map(char => {
			if (char === "*") {
				return ".*";
			}

			if ("\\.[]{}()+-^$|?+".indexOf(char) >= 0) {
				return "\\" + char;
			}

			return char;
		})
		.join("");

	return new RegExp("^" + source + "$").test(String(name));
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

function go2jsURLParse(value) {
    try {
        const parsed = new URL(value);

        return [{
            Scheme: parsed.protocol.replace(/:$/, ""),
            Host: parsed.host,
            Path: parsed.pathname,
            RawQuery: parsed.search.replace(/^\?/, ""),
            Query: function() {
                return go2jsURLValuesFromSearchParams(parsed.searchParams);
            }
        }, null];
    } catch (err) {
        return [null, err];
    }
}

function go2jsURLValuesFromSearchParams(params) {
    const values = {};

    for (const [key, value] of params.entries()) {
        if (!values[key]) {
            values[key] = [];
        }
        values[key].push(value);
    }

    values.Get = function(key) {
        const value = values[key];
        return value && value.length > 0 ? value[0] : "";
    };

    values.Set = function(key, value) {
        values[key] = [String(value)];
    };

    values.Encode = function() {
        const params = new URLSearchParams();

        for (const key of Object.keys(values)) {
            if (key === "Get" || key === "Set" || key === "Encode") {
                continue;
            }

            for (const value of values[key]) {
                params.append(key, value);
            }
        }

        return params.toString();
    };

    return values;
}

function go2jsURLValues() {
    const values = {};

    values.Get = function(key) {
        const value = values[key];
        return value && value.length > 0 ? value[0] : "";
    };

    values.Set = function(key, value) {
        values[key] = [String(value)];
    };

    values.Encode = function() {
        const params = new URLSearchParams();

        for (const key of Object.keys(values)) {
            if (key === "Get" || key === "Set" || key === "Encode") {
                continue;
            }

            for (const value of values[key]) {
                params.append(key, value);
            }
        }

        return params.toString();
    };

    return values;
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
        return go2jsJSONEmpty(value.get());
    }

    if (Array.isArray(value) || typeof value === "object") {
        return go2jsLen(value) === 0;
    }

    return false;
}

function go2jsJSONUnmarshal(data, target, fields) {
    try {
        const value = JSON.parse(
            typeof data === "string"
                ? data
                : new TextDecoder().decode(Uint8Array.from(data))
        );

        if (target !== null && target !== undefined && typeof target === "object") {
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

go2jsStringsBuilder.prototype.WriteString = function(value) {
    this.parts.push(String(value));
    return this.parts.length;
};

go2jsStringsBuilder.prototype.Write = function(value) {
    this.parts.push(go2jsBytesToString(value));
    return go2jsToArray(value).length;
};

go2jsStringsBuilder.prototype.WriteRune = function(value) {
    this.parts.push(String.fromCodePoint(Number(value)));
    return 1;
};

go2jsStringsBuilder.prototype.WriteByte = function(value) {
    this.parts.push(String.fromCharCode(Number(value) & 255));
    return 1;
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
        return 0;
    }

    if (typeof value === "string") {
        value = Array.from(new TextEncoder().encode(value));
    } else if (ArrayBuffer.isView(value)) {
        value = Array.from(value);
    } else if (Array.isArray(value)) {
        value = value.slice();
    } else {
        throw new TypeError("bytes.Buffer.Write expects byte data");
    }

    this.data.push(...value);
    return value.length;
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
    return go2jsStringByteLength(go2jsStringify(value));
};

go2jsBytesBuffer.prototype.WriteByte = function(value) {
    this.data.push(value & 0xff);
    return 1;
};

go2jsBytesBuffer.prototype.WriteRune = function(value) {
    return this.WriteString(String.fromCodePoint(value));
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

function go2jsBytesEqual(a, b) {
    if (a === null || a === undefined || b === null || b === undefined) {
        return a === b;
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
			target = target.get();

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
		target = target.get();

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
			target = target.get();
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

	if (receiver !== null && receiver !== undefined && receiver.__go2js_pointer === true) {
		return fn(receiver.get(), ...args);
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
		const target = receiver.get();

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

function go2jsEmbedProxy(target, embedded) {
    return new Proxy(target, {
        get(target, property, receiver) {
            if (property === "__go2js_embedded") {
                return embedded;
            }

            if (Reflect.has(target, property)) {
                return Reflect.get(target, property, receiver);
            }

            for (const entry of embedded) {
                const current = typeof entry === "string" ? target[entry] : entry;

                if (current === null || current === undefined) {
                    continue;
                }

                const value = current[property];

                if (value !== undefined) {
                    if (typeof value === "function") {
                        return value.bind(current);
                    }

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

function go2jsPtr(get, set) {
	const pointer = {
		get,
		set
	};

	return new Proxy(pointer, {
		get(target, property, receiver) {
			if (property === "get" || property === "set") {
				return Reflect.get(target, property, receiver);
			}

			if (property === "__go2js_pointer") {
				return true;
			}

			const value = target.get();
			if (value === null || value === undefined) {
				return undefined;
			}

			return Reflect.get(value, property, value);
		},

		set(target, property, value) {
			const current = target.get();
			if (current === null || current === undefined) {
				throw new TypeError("cannot assign through nil pointer");
			}

			return Reflect.set(current, property, value);
		},

		has(target, property) {
			const value = target.get();
			return value !== null && value !== undefined && property in Object(value);
		}
	});
}

function go2jsDeref(ptr) {
	if (ptr === null || ptr === undefined) {
		throw new TypeError("invalid pointer dereference");
	}
	if (typeof ptr.get !== "function") {
		throw new TypeError("value is not a pointer");
	}
	return ptr.get();
}

function go2jsStorePtr(ptr, value) {
	if (ptr === null || ptr === undefined) {
		throw new TypeError("invalid pointer assignment");
	}
	if (typeof ptr.set !== "function") {
		throw new TypeError("value is not a pointer");
	}
	ptr.set(value);
}

function go2jsNew(value) {
	return go2jsPtr(
		function() {
			return value;
		},
		function(next) {
			value = next;
		}
	);
}

const go2jsMethodTable = Object.create(null);
const go2jsStructFormats = Object.create(null);
const go2jsTypeNames = Object.create(null);

function go2jsRegisterTypeName(constructor, name) {
	go2jsTypeNames[name] = constructor;
}

// go2jsGoTypeName reports the Go type name of a value for %T.
function go2jsGoTypeName(value) {
	if (value === null || value === undefined) {
		return "<nil>";
	}

	if (typeof value === "boolean") {
		return "bool";
	}

	if (typeof value === "number") {
		return Number.isInteger(value) ? "int" : "float64";
	}

	if (value instanceof Error) {
		return "error";
	}

	if (value.__go2js_interface === true) {
		return value.type;
	}

	if (Array.isArray(value)) {
		return "[]interface {}";
	}

	if (value instanceof go2jsNativeMap) {
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

function go2jsInterface(value, typeName) {
	return {
		__go2js_interface: true,
		type: typeName,
		value: value
	};
}


function go2jsEqual(a, b) {
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

	const ai = a.__go2js_interface === true;
	const bi = b.__go2js_interface === true;

	if (ai && bi) {
		if (a.type !== b.type) {
			return false;
		}

		return go2jsEqual(a.value, b.value);
	}

	if (ai || bi) {
		return false;
	}

	if (Array.isArray(a) || Array.isArray(b)) {
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
		throw new TypeError("call of method on nil interface");
	}

	if (value instanceof Error && method === "Error") {
		return value.message;
	}

	if (value.__go2js_interface !== true) {
		throw new TypeError("value is not an interface");
	}

	const target = value.value;

	if (target === null || target === undefined) {
		throw new TypeError("call of method on nil interface");
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

function go2jsTypeOf(value) {
	if (value === null || value === undefined) {
		return "nil";
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

	if (value instanceof go2jsNativeMap) {
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

function go2jsAssert(value, typeName) {
	if (value !== null && value !== undefined &&
		value.__go2js_interface === true) {
		if (value.type === typeName) {
			return value.value;
		}

		throw new TypeError(
			"interface conversion: " + value.type + " is not " + typeName
		);
	}

	const actual = go2jsTypeOf(value);

	if (actual === typeName) {
		return value;
	}

	throw new TypeError("interface conversion failed");
}

function go2jsAssertOK(value, typeName) {
	if (value !== null && value !== undefined &&
		value.__go2js_interface === true &&
		value.type === typeName) {
		return [value.value, true];
	}

	return [null, false];
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
		return go2jsLen(value.get());
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

	for (const ch of s) {
		const code = ch.codePointAt(0);

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
		return go2jsCap(value.get());
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

function go2jsMap(entries) {
	const map = new go2jsNativeMap();

	if (!entries) {
		return map;
	}

	for (const entry of entries) {
		if (!Array.isArray(entry) || entry.length < 2) {
			continue;
		}
		map.set(entry[0], entry[1]);
	}

	return map;
}

function go2jsMapGet(map, key, zero) {
	if (!(map instanceof go2jsNativeMap)) {
		return zero;
	}

	const value = map.get(key);

	// A missing key yields the element type's zero value, not null.
	return value === undefined ? zero : value;
}

function go2jsMapGetOK(map, key, zero) {
	if (!(map instanceof go2jsNativeMap) || !map.has(key)) {
		return [zero, false];
	}

	return [map.get(key), true];
}

function go2jsMapSet(map, key, value) {
	if (!(map instanceof go2jsNativeMap)) {
		throw new TypeError("go2jsMapSet expects a Map");
	}
	map.set(key, value);
}

function go2jsMapDelete(map, key) {
	if (!(map instanceof go2jsNativeMap)) {
		throw new TypeError("go2jsMapDelete expects a Map");
	}
	map.delete(key);
}

function go2jsMapHas(map, key) {
	if (!(map instanceof go2jsNativeMap)) {
		return false;
	}
	return map.has(key);
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

function go2jsClear(value) {
	if (value instanceof go2jsNativeMap || value instanceof go2jsNativeSet) {
		value.clear();
		return;
	}

	if (Array.isArray(value)) {
		value.length = 0;
		return;
	}

	if (value && typeof value === "object") {
		for (const key of Object.keys(value)) {
			delete value[key];
		}
	}
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

function go2jsPanicPayload(value) {
	if (value !== null && value !== undefined && value.__go2js_panic_value !== undefined) {
		return value.__go2js_panic_value;
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

function go2jsFormat(value) {
	if (value === null || value === undefined) {
		return "<nil>";
	}

	if (value instanceof Error) {
		return value.message;
	}

	if (value.__go2js_interface === true) {
		if (typeof value.type === "string") {
			const stringer = go2jsMethodTable[value.type + ".String"];

			if (typeof stringer === "function") {
				return stringer(value.value);
			}
		}

		return go2jsFormat(value.value);
	}

	if (value.__go2js_pointer === true) {
		return "&" + go2jsFormat(go2jsDeref(value));
	}

	if (value instanceof go2jsNativeMap) {
		// fmt sorts map keys, so the output must be deterministic.
		const entries = Array.from(value.entries());
		entries.sort((a, b) => go2jsCompareValues(a[0], b[0]));

		const parts = entries.map(([key, item]) => go2jsFormat(key) + ":" + go2jsFormat(item));

		return "map[" + parts.join(" ") + "]";
	}

	if (Array.isArray(value)) {
		const parts = [];

		for (let i = 0; i < value.length; i++) {
			parts.push(go2jsFormat(value[i]));
		}

		return "[" + parts.join(" ") + "]";
	}

	switch (typeof value) {
	case "string":
		return value;
	case "number":
	case "boolean":
		return String(value);
	}

	if (typeof value.String === "function") {
		return value.String();
	}

	if (typeof value.Error === "function") {
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

		parts.push(go2jsFormat(value[key]));
	}

	return "{" + parts.join(" ") + "}";
}

function go2jsPrintln(...values) {
	process.stdout.write(go2jsJoinOperands(values, true) + "\n");
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
	if (value === null || value === undefined) {
		return false;
	}

	if (value.__go2js_interface === true) {
		return typeof value.value === "string";
	}

	return typeof value === "string";
}

function go2jsPrint(...values) {
	process.stdout.write(go2jsJoinOperands(values, false));
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

function go2jsFprint(writer, values, suffix, separator) {
	const target = go2jsUnwrap(writer);

	if (target === null || target === undefined || typeof target.Write !== "function") {
		throw new Error("fmt.Fprint: writer does not implement Write");
	}

	const written = target.Write(go2jsStringToBytes(go2jsOutputText(values, suffix, separator)));

	return Array.isArray(written) ? written[0] : written;
}

function go2jsFprintf(writer, format, args) {
	return go2jsFprint(writer, [go2jsSprintf(format, ...args)], "", "");
}

function go2jsFprintln(writer, values) {
	return go2jsFprint(writer, values, "\n", " ");
}

function go2jsSprintf(format, ...args) {
	let result = "";
	let argIndex = 0;

	for (let i = 0; i < format.length; i++) {
		const ch = format[i];

		if (ch !== "%") {
			result += ch;
			continue;
		}

		i++;
		let spec = "%";

		while (i < format.length && "+-# 0123456789.".includes(format[i])) {
			spec += format[i];
			i++;
		}

		const verb = format[i];
		spec += verb;

		if (verb === "%") {
			result += "%";
			continue;
		}

		const arg = args[argIndex];
		argIndex++;
		result += go2jsFormatValue(verb, spec, arg);
	}

	return result;
}

function go2jsTyped(value, type) {
	return {__go2js_typed: true, value: value, type: type};
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

function go2jsFormatHexBytes(value, upper) {
	let text = "";

	for (const item of go2jsToArray(value)) {
		text += (Number(item) & 255).toString(16).padStart(2, "0");
	}

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

function go2jsVerbAccepts(verb, typeName) {
	switch (verb) {
		case "d":
		case "b":
		case "o":
		case "c":
			return typeName === "int" || typeName === "int8" || typeName === "int16" ||
				typeName === "int32" || typeName === "int64" || typeName === "uint" ||
				typeName === "uint8" || typeName === "uint16" || typeName === "uint32" ||
				typeName === "uint64" || typeName === "uintptr" || typeName === "rune" ||
				typeName === "byte";
		case "e":
		case "E":
		case "f":
		case "F":
		case "g":
		case "G":
			return typeName === "int" || typeName === "int8" || typeName === "int16" ||
				typeName === "int32" || typeName === "int64" || typeName === "uint" ||
				typeName === "uint8" || typeName === "uint16" || typeName === "uint32" ||
				typeName === "uint64" || typeName === "uintptr" || typeName === "rune" ||
				typeName === "byte" || typeName === "float32" || typeName === "float64";
		case "s":
		case "q":
			return !go2jsIsBasicScalarName(typeName) || typeName === "string";
		case "x":
		case "X":
			return true;
		case "v":
		case "T":
			return true;
		case "t":
			return typeName === "bool";
		default:
			return true;
	}
}

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
		return true;
	default:
		return false;
	}
}

function go2jsGoSyntax(value) {
	if (value === null || value === undefined) {
		return "<nil>";
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

function go2jsFormatValue(verb, spec, value) {
	const parsed = go2jsParseFormatSpec(spec);
	const flags = parsed.flags;
	const precision = parsed.precision;
	const tagged = go2jsTypedType(value);

	value = go2jsUntyped(value);
	value = go2jsMaterializeValue(value);

	if (verb !== "%" && !go2jsVerbAccepts(verb, go2jsInferTypeName(value, tagged))) {
		return "%!" + verb + "(" + go2jsInferTypeName(value, tagged) + "=" + go2jsFormat(value) + ")";
	}

	// Renders the sign prefix and zero-pads the digits that follow it.
	const numberText = (num, body, prefixOverride) => {
		const prefix = prefixOverride !== undefined
			? prefixOverride
			: num < 0
				? "-"
				: flags.includes("+")
					? "+"
					: flags.includes(" ")
						? " "
						: "";

		return go2jsPadNumber(prefix + body, prefix, body, parsed);
	};

	const integerBody = (num, base, prefix, upper) => {
		const magnitude = Math.abs(Math.trunc(num));
		let body = magnitude.toString(base);

		if (upper) {
			body = body.toUpperCase();
		}

		if (prefix !== "") {
			body = prefix + body;
		}

		return numberText(num, body);
	};

	const intValue = Math.trunc(Number(value));

	switch (verb) {
		case "d":
			return numberText(intValue, String(Math.abs(intValue)));
		case "b":
			return integerBody(intValue, 2, flags.includes("#") ? "0b" : "");
		case "o":
			return integerBody(intValue, 8, flags.includes("#") ? "0" : "");
		case "x":
			if ((Array.isArray(value) || value instanceof Uint8Array) && !Number.isInteger(Number(value))) {
				return go2jsFormatHexBytes(value, false);
			}
			return integerBody(intValue, 16, flags.includes("#") ? "0x" : "", false);
		case "X":
			if ((Array.isArray(value) || value instanceof Uint8Array) && !Number.isInteger(Number(value))) {
				return go2jsFormatHexBytes(value, true);
			}
			return integerBody(intValue, 16, flags.includes("#") ? "0X" : "", true);
		case "f": {
			const num = Number(value);
			const body = Math.abs(num).toFixed(precision === null ? 6 : precision);

			return numberText(num, body);
		}
		case "e":
			return go2jsFormatE(Number(value), precision === null ? 6 : precision, parsed);
		case "E": {
			const upper = go2jsFormatE(Number(value), precision === null ? 6 : precision, parsed);
			const exponent = upper.indexOf("e");

			return exponent === -1 ? upper : upper.slice(0, exponent) + "E" + upper.slice(exponent + 1);
		}
		case "g":
			return go2jsFormatG(Number(value), precision, parsed);
		case "s": {
			let text;

			if (tagged === "[]uint8" || (tagged === null && go2jsIsByteArray(value))) {
				text = go2jsBytesToString(value);
			} else {
				text = go2jsFormat(value);
			}

			if (precision !== null) {
				text = text.slice(0, precision);
			}

			return go2jsPad(text, parsed, false);
		}
		case "v": {
			let text = flags.includes("#") ? go2jsGoSyntax(value) : go2jsFormat(value);

			if (precision !== null) {
				text = text.slice(0, precision);
			}

			return go2jsPad(text, parsed, false);
		}
		case "q":
			return go2jsPad(JSON.stringify(go2jsStringify(value)), parsed, false);
		case "t":
			return go2jsPad(value ? "true" : "false", parsed, false);
		case "c":
			return go2jsPad(String.fromCharCode(value), parsed, false);
		case "T":
			return go2jsPad(go2jsGoTypeName(value), parsed, false);
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

function go2jsPad(text, parsed, numeric) {
	if (parsed.width <= text.length) {
		return text;
	}

	const fill = parsed.width - text.length;

	if (parsed.flags.includes("-")) {
		return text + " ".repeat(fill);
	}

	if (numeric && parsed.flags.includes("0")) {
		return text.padStart(parsed.width, "0");
	}

	return " ".repeat(fill) + text;
}

// go2jsPadNumber keeps any sign or base prefix in front of the zero padding.
function go2jsPadNumber(text, prefix, body, parsed) {
	if (parsed.width <= text.length) {
		return text;
	}

	const fill = parsed.width - text.length;

	if (parsed.flags.includes("-")) {
		return text + " ".repeat(fill);
	}

	if (parsed.flags.includes("0")) {
		return prefix + body.padStart(body.length + fill, "0");
	}

	return " ".repeat(fill) + text;
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
	const prefix = value < 0 ? "-" : parsed.flags.includes("+") ? "+" : parsed.flags.includes(" ") ? " " : "";

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

function go2jsStringsIndex(s, substr) {
	return s.indexOf(substr);
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
	return s.lastIndexOf(substr);
}

function go2jsMathInf(sign) {
	return sign >= 0 ? Infinity : -Infinity;
}

function go2jsMathNaN() {
	return NaN;
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

function go2jsHexBytes(src) {
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

function go2jsStrconvQuote(s) {
	return JSON.stringify(s);
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

function go2jsSortReverse(data) {
	this.data = data;
}

go2jsSortReverse.prototype.Len = function() {
	return go2jsLen(this.data);
};

go2jsSortReverse.prototype.Less = function(i, j) {
	return go2jsRawCompare(this.data[j], this.data[i]) < 0;
};

go2jsSortReverse.prototype.Swap = function(i, j) {
	const items = this.data;
	const tmp = items[i];
	items[i] = items[j];
	items[j] = tmp;
};

function go2jsSortInterface(data) {
	if (data instanceof go2jsSortReverse) {
		const items = go2jsToArray(data.data);

		items.sort((a, b) => (a < b ? -1 : a > b ? 1 : 0));
		items.reverse();

		if (Array.isArray(data.data)) {
			data.data.length = 0;
			data.data.push(...items);
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

const go2jsSliceMeta = new WeakMap();

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

			const index = Number(property);
			if (Number.isInteger(index) && String(index) === String(property)) {
				if (index < 0 || index >= state.length) {
					return undefined;
				}
				return state.data[state.offset + index];
			}

			return Reflect.get(target, property, receiver);
		},

		has(target, property) {
			const index = Number(property);

			if (Number.isInteger(index) && String(index) === String(property)) {
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
			const index = Number(property);

			if (Number.isInteger(index) && String(index) === String(property)) {
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
			const index = Number(property);
			if (Number.isInteger(index) && String(index) === String(property)) {
				if (index < 0 || index >= state.length) {
					throw new RangeError("slice index out of range");
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

function go2jsSliceLen(value) {
	if (value === null || value === undefined) {
		return 0;
	}

	const state = go2jsSliceState(value);
	return state ? state.length : 0;
}

function go2jsSliceCap(value) {
	if (value === null || value === undefined) {
		return 0;
	}

	const state = go2jsSliceState(value);
	return state ? state.capacity : 0;
}

function go2jsSliceMake(length, capacity, zeroFactory) {
	length = Math.trunc(length);
	capacity = Math.trunc(capacity);

	if (length < 0 || capacity < 0 || length > capacity) {
		throw new RangeError("invalid slice length or capacity");
	}

	const data = Array.from(
		{length: capacity},
		() => zeroFactory()
	);

	return go2jsSliceView(data, 0, length, capacity);
}

function go2jsSliceAppend(value, ...items) {
	if (value === null || value === undefined) {
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
		data[i] = state.data[state.offset + i];
	}

	for (let i = 0; i < items.length; i++) {
		data[state.length + i] = items[i];
	}

	return go2jsSliceView(data, 0, required, capacity);
}

function go2jsSliceRange(value, low, high, max) {
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

	if (start < 0 || end < start || end > limit || limit > state.capacity) {
		throw new RangeError("slice bounds out of range");
	}

	return go2jsSliceView(
		state.data,
		state.offset + start,
		end - start,
		limit - start
	);
}

function go2jsSliceLiteral(zeroFactory, entries) {
	let length = 0;
	let next = 0;

	for (const entry of entries) {
		const index = entry[0] === null ? next : Math.trunc(entry[0]);

		if (index < 0) {
			throw new RangeError("negative slice literal index");
		}

		length = Math.max(length, index + 1);
		next = index + 1;
	}

	const data = Array.from(
		{length},
		() => zeroFactory()
	);

	next = 0;

	for (const entry of entries) {
		const index = entry[0] === null ? next : Math.trunc(entry[0]);
		data[index] = entry[1];
		next = index + 1;
	}

	return data;
}

function go2jsArrayLiteral(length, zeroFactory, entries) {
	const data = Array.from(
		{length},
		() => zeroFactory()
	);

	let next = 0;

	for (const entry of entries) {
		const index = entry[0] === null ? next : Math.trunc(entry[0]);

		if (index < 0 || index >= length) {
			throw new RangeError("array literal index out of range");
		}

		data[index] = entry[1];
		next = index + 1;
	}

	return data;
}

function go2jsStructCopy(value) {
	if (value === null || value === undefined || typeof value !== "object") {
		return value;
	}

	if (value.__go2js_pointer === true) {
		return value;
	}

	const embedded = typeof value.__go2js_embedded === "undefined"
		? null
		: value.__go2js_embedded;

	const copy = Object.create(Object.getPrototypeOf(value));

	for (const key of Object.keys(value)) {
		copy[key] = value[key];
	}

	if (embedded !== null) {
		return go2jsEmbedProxy(copy, embedded);
	}

	return copy;
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

function go2jsArrayCopy(value) {
	if (!Array.isArray(value)) {
		throw new TypeError("array copy expects an array");
	}

	return value.slice();
}

function go2jsSliceCopy(dst, src) {
	const destination = go2jsSliceState(dst);
	const source = go2jsSliceState(src);

	if (!destination || !source) {
		throw new TypeError("copy expects slices");
	}

	const count = Math.min(destination.length, source.length);
	const values = Array.from(
		{length: count},
		(_, i) => source.data[source.offset + i]
	);

	for (let i = 0; i < count; i++) {
		destination.data[destination.offset + i] = values[i];
	}

	return count;
}

function go2jsChannel(capacity) {
	return {
		__go2js_channel: true,
		buffer: [],
		capacity: typeof capacity === "number" && capacity > 0 ? Math.trunc(capacity) : 0,
		closed: false,
		receivers: [],
		senders: []
	};
}

function go2jsChanLen(channel) {
	return channel.buffer.length;
}

function go2jsChanCap(channel) {
	return channel.capacity;
}

function go2jsChannelClosed(channel) {
	return channel.closed && channel.buffer.length === 0;
}

function go2jsChannelClosedPanic(operation) {
	throw new Error("send on closed channel");
}

function go2jsChanRecvPair(channel) {
	while (true) {
		if (channel.buffer.length > 0) {
			const value = channel.buffer.shift();
			go2jsChannelPump(channel);
			return [value, true];
		}

		if (channel.closed) {
			return [go2jsChannelZero, false];
		}

		if (!go2jsProgress()) {
			throw new Error("go2js: no goroutine can unblock this channel receive");
		}
	}
}

function go2jsChanRecv(channel) {
	return go2jsChanRecvPair(channel)[0];
}

function go2jsChanTryRecv(channel) {
	if (channel.buffer.length > 0) {
		const value = channel.buffer.shift();
		go2jsChannelPump(channel);
		return [value, true];
	}

	if (channel.closed) {
		return [go2jsChannelZero, false];
	}

	return null;
}

function go2jsChanSend(channel, value) {
	while (true) {
		if (channel.closed) {
			go2jsChannelClosedPanic("send");
		}

		if (channel.capacity === 0 || channel.buffer.length < channel.capacity) {
			channel.buffer.push(value);
			go2jsChannelPump(channel);
			return;
		}

		if (!go2jsProgress()) {
			throw new Error("go2js: no goroutine can unblock this channel send");
		}
	}
}

function go2jsChanTrySend(channel, value) {
	if (channel.closed) {
		go2jsChannelClosedPanic("send");
	}

	if (channel.capacity === 0) {
		return false;
	}

	if (channel.buffer.length >= channel.capacity) {
		return false;
	}

	channel.buffer.push(value);
	go2jsChannelPump(channel);
	return true;
}

function go2jsChannelPump(channel) {
	if (channel.buffer.length === 0) {
		return;
	}

	const receiver = channel.receivers.shift();

	if (receiver === undefined) {
		return;
	}

	receiver.resolve({ channel, value: channel.buffer.shift() });
}

function go2jsChannelClose(channel) {
	if (channel.closed) {
		throw new Error("close of closed channel");
	}

	channel.closed = true;

	const waiting = channel.receivers.splice(0, channel.receivers.length);

	for (const receiver of waiting) {
		receiver.resolve({ channel, value: go2jsChannelZero, closed: true });
	}

	const blocked = channel.senders.splice(0, channel.senders.length);

	for (const sender of blocked) {
		sender.reject(new Error("send on closed channel"));
	}
}

function go2jsChannelZero() {
	return undefined;
}

function go2jsChannelRange(channel) {
	return {
		[Symbol.iterator]() {
			return {
				next() {
					const result = go2jsChanTryRecv(channel);

					if (result === null) {
						if (channel.closed) {
							return { done: true, value: undefined };
						}
						return { done: false, value: undefined };
					}

					if (result[1] === false) {
						return { done: true, value: undefined };
					}

					return { done: false, value: result[0] };
				}
			};
		}
	};
}

function go2jsWaitGroup() {
	return {
		count: 0,
		waiters: []
	};
}

const go2jsTasks = [];
let go2jsDraining = false;

function go2jsWaitGroupAdd(group, delta) {
	group.count += delta;

	if (group.count < 0) {
		throw new Error("sync: negative WaitGroup counter");
	}

	if (group.count === 0) {
		const waiting = group.waiters.splice(0, group.waiters.length);

		for (const waiter of waiting) {
			waiter.resolve();
		}
	}
}

function go2jsWaitGroupDone(group) {
	go2jsWaitGroupAdd(group, -1);
}

function go2jsRunTasks() {
	if (go2jsDraining) {
		return 0;
	}

	go2jsDraining = true;

	let executed = 0;

	try {
		while (go2jsTasks.length > 0) {
			if (executed >= 1000000) {
				throw new Error("go2js: goroutine task limit exceeded");
			}

			const task = go2jsTasks.shift();
			task();
			executed++;
		}
	} finally {
		go2jsDraining = false;
	}

	return executed;
}

function go2jsGo(task) {
	go2jsTasks.push(task);
	queueMicrotask(() => {
		go2jsRunTasks();
	});
}

function go2jsProgress() {
	if (go2jsTasks.length === 0) {
		return false;
	}

	return go2jsRunTasks() > 0;
}

function go2jsWaitGroupWait(group) {
	let budget = 1000000;

	while (group.count > 0) {
		if (budget-- === 0) {
			throw new Error("go2js: wait group did not reach zero");
		}

		go2jsRunTasks();
	}
}

function go2jsMutex() {
	return { locked: false, waiters: [] };
}

function go2jsMutexLock(mutex) {
	if (!mutex.locked) {
		mutex.locked = true;
		return;
	}

	throw new Error("sync.Mutex: already locked");
}

function go2jsMutexUnlock(mutex) {
	if (!mutex.locked) {
		throw new Error("sync: unlock of unlocked mutex");
	}

	mutex.locked = false;

	const waiting = mutex.waiters.shift();

	if (waiting !== undefined) {
		mutex.locked = true;
		waiting.resolve();
	}
}

function go2jsRWMutex() {
	return { readers: 0, writer: false, readersWaiting: [], writersWaiting: [] };
}

function go2jsRWMutexRLock(mutex) {
	if (mutex.writer) {
		throw new Error("sync: RLock of write-locked mutex");
	}

	mutex.readers++;
}

function go2jsRWMutexRUnlock(mutex) {
	if (mutex.readers === 0) {
		throw new Error("sync: RUnlock of unlocked RWMutex");
	}

	mutex.readers--;

	if (mutex.readers === 0) {
		go2jsRWMutexPromote(mutex);
	}
}

function go2jsRWMutexLock(mutex) {
	if (mutex.writer || mutex.readers > 0) {
		throw new Error("sync: Lock already held");
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

function go2jsRWMutexPromote(mutex) {
	if (mutex.writer || mutex.readers > 0) {
		return;
	}

	if (mutex.writersWaiting.length > 0) {
		const writer = mutex.writersWaiting.shift();
		mutex.writer = true;
		writer.resolve();
		return;
	}

	while (mutex.readersWaiting.length > 0) {
		mutex.readers++;
		mutex.readersWaiting.shift().resolve();
	}
}

function go2jsOnce() {
	return { done: false };
}

function go2jsOnceDo(once, fn) {
	if (once.done) {
		return;
	}

	once.done = true;
	fn();
}

function go2jsSyncMap() {
	return { entries: new go2jsNativeMap() };
}

function go2jsSyncMapStore(store, key, value) {
	store.entries.set(key, value);
}

function go2jsSyncMapLoad(store, key) {
	if (!store.entries.has(key)) {
		return [null, false];
	}

	return [store.entries.get(key), true];
}

function go2jsSyncMapLoadOrStore(store, key, value) {
	if (store.entries.has(key)) {
		return [store.entries.get(key), true];
	}

	store.entries.set(key, value);
	return [value, false];
}

function go2jsSyncMapLoadAndDelete(store, key) {
	if (!store.entries.has(key)) {
		return [null, false];
	}

	const value = store.entries.get(key);
	store.entries.delete(key);
	return [value, true];
}

function go2jsSyncMapDelete(store, key) {
	store.entries.delete(key);
}

function go2jsSyncMapSwap(store, key, value) {
	const previous = store.entries.get(key);
	store.entries.set(key, value);
	return previous;
}

function go2jsSyncMapCompareAndSwap(store, key, old, next) {
	if (!store.entries.has(key)) {
		if (old !== undefined && old !== null) {
			return false;
		}
	}

	if (store.entries.get(key) !== old) {
		return false;
	}

	store.entries.set(key, next);
	return true;
}

function go2jsSyncMapRange(store, fn) {
	for (const [key, value] of store.entries) {
		fn(key, value);
	}
}

function go2jsSyncMapLen(store) {
	return store.entries.size;
}

function go2jsAtomicCell(target) {
	if (target !== null && target !== undefined &&
		typeof target.get === "function" && typeof target.set === "function") {
		return {
			get value() {
				return target.get();
			},
			set value(next) {
				target.set(next);
			}
		};
	}

	return { value: target };
}

function go2jsAtomicLoad(cell) {
	return cell.value;
}

function go2jsAtomicStore(cell, value) {
	cell.value = value;
}

function go2jsAtomicAdd(cell, delta) {
	cell.value = cell.value + delta;
	return cell.value;
}

function go2jsAtomicSwap(cell, value) {
	const previous = cell.value;
	cell.value = value;
	return previous;
}

function go2jsAtomicCompareAndSwap(cell, expected, next) {
	if (cell.value !== expected) {
		return false;
	}

	cell.value = next;
	return true;
}

function go2jsErrorsIs(err, target) {
	if (err === target) {
		return true;
	}

	if (err === null || err === undefined || target === null || target === undefined) {
		return err === target;
	}

	if (err instanceof Error && target instanceof Error && err.message === target.message) {
		return true;
	}

	if (err === null || err === undefined) {
		return false;
	}

	if (err.cause !== undefined) {
		return go2jsErrorsIs(err.cause, target);
	}

	if (Array.isArray(err.joined)) {
		return err.joined.some(part => go2jsErrorsIs(part, target));
	}

	return false;
}

function go2jsWrapError(format, ...args) {
	const error = new Error(go2jsSprintf(format, ...args));

	for (let i = 0; i < args.length; i++) {
		if (format.includes("%w")) {
			const verbs = format.match(/%[+#0 -.]*[0-9.]*[a-zA-Z]/g) || [];
			const position = verbs.indexOf("%w");

			if (position === i) {
				error.cause = args[i];
				break;
			}
		}
	}

	return error;
}

function go2jsErrorsUnwrap(err) {
	if (err === null || err === undefined) {
		return null;
	}

	if (err.cause !== undefined) {
		return err.cause;
	}

	return null;
}

function go2jsErrorsJoin(...errs) {
	const parts = errs.flat().filter(item => item !== null && item !== undefined);

	if (parts.length === 0) {
		return null;
	}

	if (parts.length === 1) {
		return parts[0];
	}

	const error = new Error(parts.map(part => go2jsErrorMessage(part)).join("\n"));
	error.joined = parts;

	return error;
}

function go2jsErrorMessage(err) {
	if (err === null || err === undefined) {
		return "<nil>";
	}

	err = go2jsUnwrap(err);

	if (err === null || err === undefined) {
		return "<nil>";
	}

	return typeof err === "object" && err.message !== undefined ? err.message : String(err);
}

function go2jsDuration(nanoseconds) {
	return {
		nanoseconds: nanoseconds,
		valueOf() {
			return this.nanoseconds;
		},
		String() {
			return go2jsDurationString(this.nanoseconds);
		}
	};
}

// go2jsDurationNanos accepts either a duration object or a raw nanosecond count.
function go2jsDurationNanos(value) {
	if (value !== null && value !== undefined && typeof value.nanoseconds === "number") {
		return value.nanoseconds;
	}

	return Number(value);
}

// Methods on the named scalar type time.Duration are emitted as flat functions.
function DurationString(d) {
	return go2jsDurationString(go2jsDurationNanos(d));
}

function DurationNanoseconds(d) {
	return go2jsDurationNanos(d);
}

function DurationMicroseconds(d) {
	return Math.trunc(go2jsDurationNanos(d) / 1000);
}

function DurationMilliseconds(d) {
	return Math.trunc(go2jsDurationNanos(d) / 1000000);
}

function DurationSeconds(d) {
	return go2jsDurationNanos(d) / 1000000000;
}

function DurationMinutes(d) {
	return go2jsDurationNanos(d) / 60000000000;
}

function DurationHours(d) {
	return go2jsDurationNanos(d) / 3600000000000;
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

function go2jsDurationString(nanoseconds) {
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
		return value.Error();
	}

	return go2jsStringify(value);
}

function go2jsStdoutWrite(values, suffix, separator) {
	process.stdout.write(go2jsOutputText(values, suffix, separator));
}

function go2jsStderrWrite(values, suffix, separator) {
	process.stderr.write(go2jsOutputText(values, suffix, separator));
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

function go2jsExit(code) {
	process.exit(code === undefined ? 0 : Number(code));
}

function go2jsBufioText(source) {
	source = go2jsUnwrap(source);

	if (source !== null && source !== undefined && typeof source.__go2js_text === "string") {
		return source.__go2js_text;
	}

	return go2jsStringify(source);
}

function go2jsBufioNewReader(source) {
	let position = 0;
	const text = go2jsBufioText(source);

	function takeUntil(separator, keepSeparator) {
		const stop = text.indexOf(separator, position);

		if (stop === -1) {
			const rest = text.slice(position);
			position = text.length;
			return rest;
		}

		const end = keepSeparator ? stop + separator.length : stop;
		const chunk = text.slice(position, end);
		position = end;

		return chunk;
	}

	return {
		ReadString: function (separator) {
			return [takeUntil(go2jsStringify(separator), true), null];
		},
		ReadLine: function () {
			if (position >= text.length) {
				return null;
			}

			const line = takeUntil("\n", false);
			return line.endsWith("\r") ? line.slice(0, -1) : line;
		},
	};
}

function go2jsBufioNewWriter(destination) {
	const target = go2jsUnwrap(destination);

	let pending = "";

	return {
		Write: function (chunk) {
			const text = go2jsBytesToString(chunk);
			pending += text;
			return [text.length, null];
		},
		WriteString: function (chunk) {
			const text = go2jsStringify(chunk);
			pending += text;
			return text.length;
		},
		Flush: function () {
			if (pending === "") {
				return null;
			}

			const written = go2jsWriteDestination(target, pending);
			pending = "";

			return written;
		},
	};
}

function go2jsBufioNewScanner(source) {
	const lines = go2jsBufioSplitLines(go2jsBufioText(source));
	let index = 0;
	let line = "";

	return {
		Scan: function () {
			if (index >= lines.length) {
				return false;
			}

			line = lines[index];
			index++;

			return true;
		},
		Text: function () {
			return line;
		},
		Bytes: function () {
			return go2jsStringToBytes(line);
		},
		Buffer: function () {
			return line;
		},
		Err: function () {
			return null;
		},
	};
}

function go2jsBufioSplitLines(text) {
	if (text === "") {
		return [];
	}

	const lines = text.split("\n");

	if (lines.length > 0 && lines[lines.length - 1] === "") {
		lines.pop();
	}

	return lines.map(line => (line.endsWith("\r") ? line.slice(0, -1) : line));
}

function go2jsBufioScanLines() {
	return 0;
}

const go2jsRandState = { seed: 0x2545f491 };

function go2jsRandSeed(value) {
	go2jsRandState.seed = (value >>> 0) || 0x2545f491;
}

function go2jsRandNext() {
	go2jsRandState.seed = (Math.imul(go2jsRandState.seed, 1103515245) + 12345) >>> 0;
	return go2jsRandState.seed;
}

function go2jsRandInt63() {
	return go2jsRandNext();
}

function go2jsRandInt63n(limit) {
	if (limit <= 0) {
		go2jsPanic("invalid argument to Int63n");
	}

	return go2jsRandNext() % limit;
}

function go2jsRandIntn(limit) {
	if (limit <= 0) {
		go2jsPanic("invalid argument to Intn");
	}

	return go2jsRandInt63n(limit);
}

function go2jsRandInt() {
	return go2jsRandNext() % 0x7fffffff;
}

function go2jsRandFloat64() {
	return go2jsRandNext() / 4294967296;
}

function go2jsRandShuffle(values) {
	for (let i = values.length - 1; i > 0; i--) {
		const j = go2jsRandIntn(i + 1);
		const swap = values[i];
		values[i] = values[j];
		values[j] = swap;
	}

	return values;
}

function go2jsSlicesSort(a) {
	if (a === null || a === undefined) {
		return;
	}

	const items = [];

	for (let i = 0; i < go2jsLen(a); i++) {
		items.push(a[i]);
	}

	items.sort(go2jsCompareValues);

	for (let i = 0; i < items.length; i++) {
		a[i] = items[i];
	}
}

function go2jsSlicesMergeSort(a, compare) {
	const scratch = new Array(a.length);

	const merge = (lo, mid, hi) => {
		for (let index = lo; index < hi; index++) {
			scratch[index] = a[index];
		}

		let left = lo;
		let right = mid;
		let out = lo;

		while (left < mid && right < hi) {
			if (compare(scratch[right], scratch[left]) < 0) {
				a[out++] = scratch[right++];
			} else {
				a[out++] = scratch[left++];
			}
		}

		while (left < mid) {
			a[out++] = scratch[left++];
		}

		while (right < hi) {
			a[out++] = scratch[right++];
		}
	};

	const sort = (lo, hi) => {
		if (hi - lo < 2) {
			return;
		}

		const mid = (lo + hi) >> 1;

		sort(lo, mid);
		sort(mid, hi);
		merge(lo, mid, hi);
	};

	sort(0, a.length);
}

function go2jsSlicesSortFunc(a, compare) {
	a.sort((x, y) => {
		const result = compare(x, y);

		return result < 0 ? -1 : result > 0 ? 1 : 0;
	});
}

function go2jsSlicesSortStableFunc(a, compare) {
	go2jsSlicesMergeSort(a, compare);
}

function go2jsSlicesReverse(a) {
	a.reverse();
}

function go2jsSlicesContains(a, value) {
	for (const item of a) {
		if (go2jsEqual(item, value)) {
			return true;
		}
	}

	return false;
}

function go2jsSlicesIndex(a, value) {
	for (let index = 0; index < a.length; index++) {
		if (go2jsEqual(a[index], value)) {
			return index;
		}
	}

	return -1;
}

function go2jsSlicesIndexFunc(a, predicate) {
	for (let index = 0; index < a.length; index++) {
		if (predicate(a[index])) {
			return index;
		}
	}

	return -1;
}

function go2jsSlicesContainsFunc(a, predicate) {
	return go2jsSlicesIndexFunc(a, predicate) >= 0;
}

function go2jsSlicesIndexAny(a, values) {
	for (let index = 0; index < a.length; index++) {
		if (values.includes(a[index])) {
			return index;
		}
	}

	return -1;
}


function go2jsSlicesEqual(a, b) {
	if (a === b) {
		return true;
	}

	if (a === null || b === null || a === undefined || b === undefined) {
		return false;
	}

	if (a.length !== b.length) {
		return false;
	}

	for (let index = 0; index < a.length; index++) {
		if (!go2jsEqual(a[index], b[index])) {
			return false;
		}
	}

	return true;
}

function go2jsSlicesEqualFunc(a, b, equals) {
	if (a === b) {
		return true;
	}

	if (a === null || b === null || a === undefined || b === undefined) {
		return false;
	}

	if (a.length !== b.length) {
		return false;
	}

	for (let index = 0; index < a.length; index++) {
		if (!equals(a[index], b[index])) {
			return false;
		}
	}

	return true;
}

function go2jsSlicesMax(a) {
	if (a.length === 0) {
		throw new Error("slices.Max: empty list");
	}

	let best = a[0];

	for (const item of a) {
		if (go2jsCompareValues(item, best) > 0) {
			best = item;
		}
	}

	return best;
}

function go2jsSlicesMin(a) {
	if (a.length === 0) {
		throw new Error("slices.Min: empty list");
	}

	let best = a[0];

	for (const item of a) {
		if (go2jsCompareValues(item, best) < 0) {
			best = item;
		}
	}

	return best;
}


function go2jsSlicesMaxFunc(a, compare) {
	if (a.length === 0) {
		throw new Error("slices.MaxFunc: empty list");
	}

	let best = a[0];

	for (const item of a) {
		if (compare(item, best) > 0) {
			best = item;
		}
	}

	return best;
}

function go2jsSlicesMinFunc(a, compare) {
	if (a.length === 0) {
		throw new Error("slices.MinFunc: empty list");
	}

	let best = a[0];

	for (const item of a) {
		if (compare(item, best) < 0) {
			best = item;
		}
	}

	return best;
}

function go2jsSlicesClone(a) {
	return a === null || a === undefined ? a : a.slice();
}

function go2jsSlicesSorted(source) {
	return go2jsSlicesCollect(source).slice().sort(go2jsCompareValues);
}

function go2jsSlicesCompact(a) {
	const out = [];

	for (const item of a) {
		if (out.length === 0 || !go2jsEqual(out[out.length - 1], item)) {
			out.push(item);
		}
	}

	return out;
}

function go2jsSlicesCompactFunc(a, equals) {
	const out = [];

	for (const item of a) {
		if (out.length === 0 || !equals(out[out.length - 1], item)) {
			out.push(item);
		}
	}

	return out;
}

function go2jsSlicesInsert(a, index, ...values) {
	const clamped = Math.max(0, Math.min(index, a.length));

	return a.slice(0, clamped).concat(values, a.slice(clamped));
}

function go2jsSlicesDelete(a, start, end) {
	const low = Math.max(0, Math.min(start, a.length));
	const high = Math.max(low, Math.min(end, a.length));

	return a.slice(0, low).concat(a.slice(high));
}

function go2jsSlicesDeleteFunc(a, predicate) {
	return a.filter(item => !predicate(item));
}

function go2jsSlicesGrow(a, count) {
	if (a.length === 0) {
		return [];
	}

	return a;
}

function go2jsSlicesClip(a) {
	if (a.length === 0) {
		return [];
	}

	return a;
}

function go2jsSlicesRepeat(value, count) {
	return new Array(count).fill(value);
}

function go2jsSlicesConcat(...lists) {
	const out = [];

	for (const list of lists) {
		for (const item of list) {
			out.push(item);
		}
	}

	return out;
}



function go2jsSlicesAll(source, predicate) {
	for (const item of go2jsSlicesCollect(source)) {
		if (!predicate(item)) {
			return false;
		}
	}

	return true;
}




function go2jsSlicesCollect(source) {
	if (source === null || source === undefined) {
		return [];
	}

	if (typeof source === "function") {
		const out = [];

		for (const value of source()) {
			out.push(value);
		}

		return out;
	}

	if (Array.isArray(source)) {
		return source.slice();
	}

	return Array.from(source);
}

function go2jsSlicesBinarySearch(a, target) {
	let low = 0;
	let high = a.length;

	while (low < high) {
		const mid = (low + high) >>> 1;

		if (go2jsCompareValues(a[mid], target) < 0) {
			low = mid + 1;
		} else {
			high = mid;
		}
	}

	return [low, low < a.length && go2jsEqual(a[low], target)];
}

function go2jsSlicesBinarySearchFunc(a, target, compare) {
	let low = 0;
	let high = a.length;

	while (low < high) {
		const mid = (low + high) >>> 1;

		if (compare(a[mid], target) < 0) {
			low = mid + 1;
		} else {
			high = mid;
		}
	}

	return [low, low < a.length && compare(a[low], target) === 0];
}

function go2jsMapsKeys(m) {
	return go2jsMapKeys(m).sort(go2jsCompareValues);
}

function go2jsMapsValues(m) {
	return go2jsMapsKeys(m).map(key => go2jsMapGet(m, key, null));
}

function go2jsMapsClone(m) {
	return go2jsMap(go2jsMapEntries(m));
}


function go2jsMapsCopy(destination, ...sources) {
	for (const source of sources) {
		if (source === null || source === undefined) {
			continue;
		}

		for (const entry of go2jsMapEntries(source)) {
			go2jsMapSet(destination, entry[0], entry[1]);
		}
	}

	return destination;
}

function go2jsMapsDeleteFunc(m, predicate) {
	for (const entry of go2jsMapEntries(m)) {
		if (predicate(entry[0], entry[1])) {
			go2jsMapDelete(m, entry[0]);
		}
	}
}

function go2jsMapsEqual(a, b) {
	if (a === b) {
		return true;
	}

	if (!(a instanceof go2jsNativeMap) || !(b instanceof go2jsNativeMap)) {
		return false;
	}

	if (a.size !== b.size) {
		return false;
	}

	for (const entry of go2jsMapEntries(a)) {
		if (!go2jsMapHas(b, entry[0])) {
			return false;
		}

		if (!go2jsEqual(entry[1], go2jsMapGet(b, entry[0], null))) {
			return false;
		}
	}

	return true;
}

function go2jsMapsEqualFunc(a, b, equals) {
	if (a === b) {
		return true;
	}

	if (!(a instanceof go2jsNativeMap) || !(b instanceof go2jsNativeMap)) {
		return false;
	}

	if (a.size !== b.size) {
		return false;
	}

	for (const entry of go2jsMapEntries(a)) {
		if (!go2jsMapHas(b, entry[0])) {
			return false;
		}

		if (!equals(entry[1], go2jsMapGet(b, entry[0], null))) {
			return false;
		}
	}

	return true;
}

const go2jsBase64Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
const go2jsBase64URLAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_";

function go2jsBase64NewEncoding(alphabet, padding) {
	return {
		alphabet: alphabet,
		padding: padding,
		EncodeToString(value) {
			return go2jsBase64EncodeToString(this, value);
		},
		DecodeString(value) {
			return go2jsBase64DecodeString(this, value);
		},
		EncodedLen(n) {
			return go2jsBase64EncodedLen(this, n);
		},
		DecodedLen(n) {
			return go2jsBase64DecodedLen(n);
		},
		Strict() {
			return go2jsBase64NewEncoding(this.alphabet, this.padding);
		}
	};
}

function go2jsBase64StdEncoding() {
	return go2jsBase64NewEncoding(go2jsBase64Alphabet, "=");
}

function go2jsBase64URLEncoding() {
	return go2jsBase64NewEncoding(go2jsBase64URLAlphabet, "=");
}

function go2jsBase64RawStdEncoding() {
	return go2jsBase64NewEncoding(go2jsBase64Alphabet, "");
}

function go2jsBase64RawURLEncoding() {
	return go2jsBase64NewEncoding(go2jsBase64URLAlphabet, "");
}

function go2jsBase64Strict() {
	return go2jsBase64StdEncoding();
}

function go2jsBase64EncodeToString(encoding, value) {
	const bytes = go2jsToArray(value);
	let out = "";

	for (let index = 0; index < bytes.length; index += 3) {
		const first = bytes[index];
		const second = bytes[index + 1];
		const third = bytes[index + 2];
		const chunk = (first << 16) | ((second === undefined ? 0 : second) << 8) | (third === undefined ? 0 : third);

		out += encoding.alphabet[(chunk >> 18) & 63];
		out += encoding.alphabet[(chunk >> 12) & 63];
		out += second === undefined ? encoding.padding : encoding.alphabet[(chunk >> 6) & 63];
		out += third === undefined ? encoding.padding : encoding.alphabet[chunk & 63];
	}

	return out;
}

function go2jsBase64DecodeString(encoding, value) {
	let text = go2jsStringify(value).replace(/=+$/, "");

	if (encoding === undefined || encoding === null) {
		encoding = /[-_]/.test(text) ? go2jsBase64URLEncoding() : go2jsBase64StdEncoding();
	}

	const out = [];
	let buffer = 0;
	let bits = 0;

	for (const char of text) {
		const index = encoding.alphabet.indexOf(char);

		if (index < 0) {
			return [out, new Error("illegal base64 data at input byte " + Array.from(text).indexOf(char))];
		}

		buffer = (buffer << 6) | index;
		bits += 6;

		if (bits >= 8) {
			bits -= 8;
			out.push((buffer >> bits) & 255);
		}
	}

	return [out, null];
}

function go2jsBase64EncodedLen(encoding, n) {
	if (n === undefined) {
		n = encoding;
	}

	const value = Number(n);

	return Math.ceil(value / 3) * 4;
}

function go2jsBase64DecodedLen(encoding, n) {
	if (n === undefined) {
		n = encoding;
	}

	const value = Number(n);

	return Math.floor(value * 3 / 4);
}

function go2jsBase64NewEncoder() {
	return go2jsBase64StdEncoding();
}

function go2jsBase64NewDecoder() {
	return go2jsBase64StdEncoding();
}

function go2jsBase64EncoderWrite(encoding, value) {
	return go2jsBase64EncodeToString(encoding, value);
}

function go2jsBase64DecoderRead(encoding, value) {
	return go2jsBase64DecodeString(encoding, value)[0];
}

function go2jsLogFlags() {
	return go2jsLogStandardFlags;
}

function go2jsLogPrefix() {
	return go2jsLogCurrentPrefix;
}

function go2jsLogWriter() {
	return process.stderr;
}

function go2jsLogSprintln(values) {
	let out = "";

	for (let index = 0; index < values.length; index++) {
		if (index > 0) {
			out += " ";
		}

		out += go2jsFormat(values[index]);
	}

	return out + "\n";
}

function go2jsLogOutput(text) {
	process.stderr.write(text);
}

function go2jsLogDefault() {
	return go2jsLogDefaultLogger;
}

function go2jsLogStandardWriter() {
	return process.stderr;
}

function go2jsLogStamp() {
	const now = new Date();
	const pad = value => String(value).padStart(2, "0");
	let stamp = "";

	if ((go2jsLogStandardFlags & 2) === 2) {
		stamp += pad(now.getHours()) + ":" + pad(now.getMinutes()) + ":" + pad(now.getSeconds()) + " ";
	}

	if ((go2jsLogStandardFlags & 1) === 1) {
		stamp = now.getFullYear() + "/" + pad(now.getMonth() + 1) + "/" + pad(now.getDate()) + " " + stamp;
	}

	return stamp;
}

function go2jsLogDecorate(text) {
	return go2jsLogCurrentPrefix + go2jsLogStamp() + text;
}

function go2jsLogWriteStandard(text) {
	process.stderr.write(go2jsLogDecorate(text));
}

function go2jsLogBuildStandard() {
	return {
		Print(...values) {
			go2jsLogWriteStandard(go2jsLogEnsureNewline(go2jsSprint(values)));
		},
		Printf(format, ...values) {
			go2jsLogWriteStandard(go2jsLogEnsureNewline(go2jsSprintf(format, ...values)));
		},
		Println(...values) {
			go2jsLogWriteStandard(go2jsLogSprintln(values));
		},
		Fatal(...values) {
			go2jsLogWriteStandard(go2jsLogEnsureNewline(go2jsSprint(values)));
			process.exit(1);
		},
		Fatalf(format, ...values) {
			go2jsLogWriteStandard(go2jsLogEnsureNewline(go2jsSprintf(format, ...values)));
			process.exit(1);
		},
		Fatalln(...values) {
			go2jsLogWriteStandard(go2jsLogSprintln(values));
			process.exit(1);
		},
		Panic(...values) {
			go2jsPanic(go2jsSprint(values));
		},
		Panicf(format, ...values) {
			go2jsPanic(go2jsSprintf(format, ...values));
		},
		Writer() {
			return go2jsLogStandardWriter();
		},
		SetPrefix(value) {
			go2jsLogCurrentPrefix = go2jsRawText(value);
		},
		Prefix() {
			return go2jsLogCurrentPrefix;
		},
		SetFlags(value) {
			go2jsLogStandardFlags = Number(value) | 0;
		},
		Flags() {
			return go2jsLogStandardFlags;
		},
	};
}

function go2jsLogSetFlags(value) {
	go2jsLogStandardFlags = Number(value) | 0;
}

function go2jsLogSetPrefix(value) {
	go2jsLogCurrentPrefix = go2jsStringify(value);
}


function go2jsLogEnsureNewline(text) {
	if (text === "" || text.endsWith("\n")) {
		return text;
	}

	return text + "\n";
}

function go2jsLogPrint(...values) {
	go2jsLogOutput(go2jsLogEnsureNewline(go2jsSprint(values)));
}

function go2jsLogPrintf(format, ...values) {
	go2jsLogOutput(go2jsLogEnsureNewline(go2jsSprintf(format, ...values)));
}

function go2jsLogPrintln(...values) {
	go2jsLogOutput(go2jsLogSprintln(values));
}

function go2jsLogPanic(...values) {
	go2jsPanic(go2jsSprint(values));
}

function go2jsLogPanicf(format, ...values) {
	go2jsPanic(go2jsSprintf(format, ...values));
}

function go2jsLogFatal(...values) {
	go2jsLogOutput(go2jsLogEnsureNewline(go2jsSprint(values)));
	process.exit(1);
}

function go2jsLogFatalf(format, ...values) {
	go2jsLogOutput(go2jsLogEnsureNewline(go2jsSprintf(format, ...values)));
	process.exit(1);
}

function go2jsLogFatalln(...values) {
	go2jsLogOutput(go2jsLogSprintln(values));
	process.exit(1);
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
		inner.Write(go2jsStringToBytes(text));
		return;
	}

	process.stderr.write(text);
}

function go2jsLogNew(writer, prefix, ...rest) {
	const text = go2jsRawText(prefix);
	const logger = {
		writer: writer === undefined ? process.stderr : writer,
		prefix: text,
		flags: rest.length > 0 ? Number(rest[0]) | 0 : go2jsLogStandardFlags,
		write(chunk) {
			go2jsLogWriteTo(this.writer, this.prefix + chunk);
		},
		Print(...values) {
			this.write(go2jsLogEnsureNewline(go2jsSprint(values)));
		},
		Printf(format, ...values) {
			this.write(go2jsLogEnsureNewline(go2jsSprintf(format, ...values)));
		},
		Println(...values) {
			this.write(go2jsLogSprintln(values));
		},
		Fatal(...values) {
			this.write(go2jsLogEnsureNewline(go2jsSprint(values)));
			process.exit(1);
		},
		Fatalf(format, ...values) {
			this.write(go2jsLogEnsureNewline(go2jsSprintf(format, ...values)));
			process.exit(1);
		},
		Fatalln(...values) {
			this.write(go2jsLogSprintln(values));
			process.exit(1);
		},
		Panic(...values) {
			go2jsPanic(go2jsSprint(values));
		},
		Panicf(format, ...values) {
			go2jsPanic(go2jsSprintf(format, ...values));
		},
		Writer() {
			return this.writer;
		},
		SetPrefix(value) {
			this.prefix = go2jsRawText(value);
		},
		Prefix() {
			return this.prefix;
		},
		SetFlags(value) {
			this.flags = Number(value) | 0;
		},
		Flags() {
			return this.flags;
		},
	};

	return logger;
}

var go2jsLogStandardFlags = 3;
var go2jsLogCurrentPrefix = "";
var go2jsLogDefaultLogger = go2jsLogBuildStandard();

function go2jsUTF16Encode(value) {
	const units = [];

	for (const code of go2jsCodePoints(value)) {
		if (code < 0x10000) {
			units.push(code);
			continue;
		}

		const adjusted = code - 0x10000;

		units.push(0xd800 + (adjusted >> 10));
		units.push(0xdc00 + (adjusted & 0x3ff));
	}

	return units;
}

function go2jsUTF16Decode(units) {
	let out = "";

	for (let index = 0; index < units.length; index++) {
		const unit = Number(units[index]);

		if (unit >= 0xd800 && unit <= 0xdbff && index + 1 < units.length) {
			const low = Number(units[index + 1]);

			if (low >= 0xdc00 && low <= 0xdfff) {
				out += String.fromCharCode(unit, low);
				index++;
				continue;
			}
		}

		out += String.fromCharCode(unit);
	}

	return out;
}

function go2jsUTF16IsSurrogate(value) {
	const unit = Number(value);

	return unit >= 0xd800 && unit <= 0xdfff;
}

function go2jsUTF16DecodeRune(r1, r2) {
	const high = Number(r1);
	const low = Number(r2);

	if (high < 0xd800 || high > 0xdbff || low < 0xdc00 || low > 0xdfff) {
		return 0xfffd;
	}

	return 0x10000 + ((high - 0xd800) << 10) + (low - 0xdc00);
}

function go2jsUTF16EncodeRune(value) {
	const code = Number(value);

	if (code < 0 || code > 0x10ffff) {
		return [0xfffd, 0xfffd];
	}

	if (code < 0x10000) {
		return [code, 0xfffd];
	}

	const adjusted = code - 0x10000;

	return [0xd800 + (adjusted >> 10), 0xdc00 + (adjusted & 0x3ff)];
}

function go2jsStringsClone(value) {
	return go2jsStringify(value);
}

function go2jsStringsContainsAny(value, chars) {
	const set = Array.from(go2jsStringify(chars));

	return Array.from(go2jsStringify(value)).some(char => set.includes(char));
}

function go2jsStringsIndexAny(value, chars) {
	const text = go2jsStringify(value);
	const set = go2jsStringify(chars);

	for (let index = 0; index < text.length; index++) {
		if (set.includes(text[index])) {
			return index;
		}
	}

	return -1;
}

function go2jsStringsLastIndexAny(value, chars) {
	const text = go2jsStringify(value);
	const set = go2jsStringify(chars);

	for (let index = text.length - 1; index >= 0; index--) {
		if (set.includes(text[index])) {
			return index;
		}
	}

	return -1;
}

function go2jsStringsIndexByte(value, b) {
	return go2jsStringify(value).indexOf(String.fromCodePoint(Number(b)));
}

function go2jsStringsLastIndexByte(value, b) {
	return go2jsStringify(value).lastIndexOf(String.fromCodePoint(Number(b)));
}

function go2jsStringsIndexFunc(value, predicate) {
	const chars = Array.from(go2jsStringify(value));

	for (let index = 0; index < chars.length; index++) {
		if (predicate(chars[index].codePointAt(0))) {
			return index;
		}
	}

	return -1;
}

function go2jsStringsContainsFunc(value, predicate) {
	return go2jsStringsIndexFunc(value, predicate) >= 0;
}

function go2jsStringsFieldsFunc(value, predicate) {
	const chars = Array.from(go2jsStringify(value));
	const out = [];
	let current = "";

	for (const char of chars) {
		if (predicate(char.codePointAt(0))) {
			if (current !== "") {
				out.push(current);
				current = "";
			}

			continue;
		}

		current += char;
	}

	if (current !== "") {
		out.push(current);
	}

	return out;
}

function go2jsStringsSplitAfter(value, sep) {
	const parts = go2jsStringify(value).split(sep);

	return parts.map((part, index) => index === parts.length - 1 ? part : part + sep);
}

function go2jsStringsSplitAfterN(value, sep, n) {
	const parts = go2jsStringify(value).split(sep);

	if (n < 0) {
		return go2jsStringsSplitAfter(value, sep);
	}

	if (parts.length <= n) {
		return go2jsStringsSplitAfter(value, sep);
	}

	const head = parts.slice(0, n - 1).map(part => part + sep);

	return head.concat(parts.slice(n - 1).join(sep));
}

function go2jsStringsTrimFunc(value, predicate) {
	const chars = Array.from(go2jsStringify(value));
	let start = 0;
	let end = chars.length;

	while (start < end && predicate(chars[start].codePointAt(0))) {
		start++;
	}

	while (end > start && predicate(chars[end - 1].codePointAt(0))) {
		end--;
	}

	return chars.slice(start, end).join("");
}

function go2jsStringsTrimLeftFunc(value, predicate) {
	const chars = Array.from(go2jsStringify(value));
	let start = 0;

	while (start < chars.length && predicate(chars[start].codePointAt(0))) {
		start++;
	}

	return chars.slice(start).join("");
}

function go2jsStringsTrimRightFunc(value, predicate) {
	const chars = Array.from(go2jsStringify(value));
	let end = chars.length;

	while (end > 0 && predicate(chars[end - 1].codePointAt(0))) {
		end--;
	}

	return chars.slice(0, end).join("");
}

function go2jsStringsTrimLeft(value, cutset) {
	return go2jsStringsTrimLeftFunc(value, code => go2jsStringify(cutset).includes(String.fromCodePoint(code)));
}

function go2jsStringsTrimRight(value, cutset) {
	return go2jsStringsTrimRightFunc(value, code => go2jsStringify(cutset).includes(String.fromCodePoint(code)));
}

function go2jsStringsTitle(value) {
	let out = "";
	let previousIsSeparator = true;

	for (const char of go2jsStringify(value)) {
		if (previousIsSeparator) {
			out += char.toUpperCase();
		} else {
			out += char.toLowerCase();
		}

		previousIsSeparator = /[^\p{L}\p{N}]/u.test(char);
	}

	return out;
}

function go2jsStringsMap(mapping, value) {
	return Array.from(go2jsRawText(value))
		.map(char => mapping(char.codePointAt(0)))
		.filter(code => code >= 0)
		.map(code => String.fromCodePoint(code))
		.join("");
}

function go2jsStringsContainsRune(value, r) {
	return go2jsStringify(value).includes(String.fromCodePoint(Number(r)));
}

function go2jsStringsIndexRune(value, r) {
	return go2jsStringify(value).indexOf(String.fromCodePoint(Number(r)));
}

function go2jsStringsNewReplacer(...args) {
	const pairs = [];

	for (let index = 0; index + 1 < args.length; index += 2) {
		pairs.push([go2jsStringify(args[index]), go2jsStringify(args[index + 1])]);
	}

	return {
		Replace(text) {
			let out = go2jsStringify(text);

			for (const [from, to] of pairs) {
				if (from === "") {
					continue;
				}

				out = out.split(from).join(to);
			}

			return out;
		},
		ReplaceAll(text) {
			return this.Replace(text);
		},
	};
}

function go2jsStrconvAppendQuote(target, value) {
	target.push(go2jsStrconvQuote(value));

	return target;
}

function go2jsStrconvQuoteRune(value) {
	return go2jsStrconvQuote(String.fromCodePoint(Number(value)));
}

function go2jsStrconvAppendQuoteRune(target, value) {
	target.push(go2jsStrconvQuoteRune(value));

	return target;
}

function go2jsStrconvIsPrint(value) {
	const text = go2jsStringify(value);

	return text.length > 0 && !/[\x00-\x1f\x7f]/.test(text);
}

function go2jsStrconvIsGraphic(value) {
	const text = go2jsStringify(value);

	return text.length > 0 && !/[\x00-\x1f\x7f]/.test(text);
}

function go2jsMathCbrt(value) {
	return Math.cbrt(Number(value));
}

function go2jsMathMod(x, y) {
	return Number(x) % Number(y);
}

function go2jsMathRemainder(x, y) {
	const a = Number(x);
	const b = Number(y);

	if (b === 0) {
		return NaN;
	}

	const quotient = a / b;
	const rounded = Math.round(quotient);
	const diff = a - rounded * b;

	return diff === 0 ? 0 : diff;
}

function go2jsMathHypot(...values) {
	return Math.hypot(...values.map(Number));
}

function go2jsMathGamma(value) {
	return Math.exp(Math.log(Number(value)));
}

function go2jsMathLogb(value) {
	return Math.log2(Number(value));
}

function go2jsMathSinh(value) {
	return Math.sinh(Number(value));
}

function go2jsMathCosh(value) {
	return Math.cosh(Number(value));
}

function go2jsMathTanh(value) {
	return Math.tanh(Number(value));
}

function go2jsMathAsinh(value) {
	return Math.asinh(Number(value));
}

function go2jsMathAcosh(value) {
	return Math.acosh(Number(value));
}

function go2jsMathAtanh(value) {
	return Math.atanh(Number(value));
}

function go2jsMathCopysign(x, y) {
	const magnitude = Math.abs(Number(x));

	return Number(y) < 0 || Object.is(Number(y), -0) ? -magnitude : magnitude;
}

function go2jsMathDim(x, y) {
	return Math.max(Number(x) - Number(y), 0);
}

function go2jsMathModf(value) {
	const n = Number(value);
	const integer = Math.trunc(n);

	return [integer, n - integer];
}

function go2jsMathSincos(value) {
	const n = Number(value);

	return [Math.sin(n), Math.cos(n)];
}

function go2jsMathMaxInt() { return 9223372036854775807; }
function go2jsMathMinInt() { return -9223372036854775808; }
function go2jsMathMaxInt8() { return 127; }
function go2jsMathMinInt8() { return -128; }
function go2jsMathMaxInt16() { return 32767; }
function go2jsMathMinInt16() { return -32768; }
function go2jsMathMaxInt32() { return 2147483647; }
function go2jsMathMinInt32() { return -2147483648; }
function go2jsMathMaxInt64() { return 9223372036854775807; }
function go2jsMathMinInt64() { return -9223372036854775808; }
function go2jsMathMaxUint() { return 18446744073709551615; }
function go2jsMathMaxUint8() { return 255; }
function go2jsMathMaxUint16() { return 65535; }
function go2jsMathMaxUint32() { return 4294967295; }
function go2jsMathMaxUint64() { return 18446744073709551615; }
function go2jsMathMaxFloat32() { return 3.4028234663852886e+38; }

function go2jsSortOrderCompare(left, right, less) {
	const result = less(left, right);

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

function go2jsSortSliceStable(a, less) {
	go2jsSortIndexStable(a, less);
}

function go2jsSortStringsAreSorted(a) {
	for (let index = 1; index < a.length; index++) {
		if (go2jsCompareValues(a[index - 1], a[index]) > 0) {
			return false;
		}
	}

	return true;
}

function go2jsSortIntsAreSorted(a) {
	return go2jsSortStringsAreSorted(a);
}

function go2jsSortFloat64sAreSorted(a) {
	return go2jsSortStringsAreSorted(a);
}

function go2jsSortSearchStrings(a, target) {
	return go2jsSlicesBinarySearch(a, target)[0];
}

function go2jsSortSearchInts(a, target) {
	return go2jsSlicesBinarySearch(a, target)[0];
}

function go2jsSortSearchFloat64s(a, target) {
	return go2jsSlicesBinarySearch(a, target)[0];
}

function go2jsOSReadFile(path) {
	try {
		return [require("fs").readFileSync(String(path)), null];
	} catch (err) {
		return [null, go2jsOSError(err)];
	}
}

function go2jsOSWriteFile(path, data) {
	try {
		require("fs").writeFileSync(String(path), go2jsToArray(data));
		return null;
	} catch (err) {
		return go2jsOSError(err);
	}
}

function go2jsOSOpen(name) {
	try {
		const handle = require("fs").openSync(String(name), "r");

		return {
			read(buffer, offset, length) {
				const out = Buffer.alloc(length);
				const read = require("fs").readSync(handle, out, 0, length, offset);
				for (let index = 0; index < read; index++) {
					buffer[offset + index] = out[index];
				}
				return read;
			},
			write(buffer) {
				return require("fs").writeSync(handle, Buffer.from(go2jsToArray(buffer)));
			},
			Close() {
				require("fs").closeSync(handle);
			},
			Name() {
				return String(name);
			},
		};
	} catch (err) {
		return go2jsOSError(err);
	}
}

function go2jsOSCreate(name) {
	return go2jsOSOpen(name);
}

function go2jsOSRemove(name) {
	try {
		require("fs").unlinkSync(String(name));
		return null;
	} catch (err) {
		return go2jsOSError(err);
	}
}

function go2jsOSMkdirAll(name) {
	try {
		require("fs").mkdirSync(String(name), {recursive: true});
		return null;
	} catch (err) {
		return go2jsOSError(err);
	}
}

function go2jsOSHostname() {
	try {
		return require("os").hostname();
	} catch (err) {
		return "";
	}
}

function go2jsOSExecutable() {
	try {
		return require("process").execPath;
	} catch (err) {
		return "";
	}
}

function go2jsOSTempDir() {
	try {
		return require("os").tmpdir();
	} catch (err) {
		return "/tmp";
	}
}

function go2jsOSGetpid() {
	return process.pid;
}

function go2jsOSExit(code) {
	process.exit(code === undefined ? 0 : Number(code));
}

function go2jsOSIsNotExist(err) {
	return err !== null && err !== undefined && err.code === "ENOENT";
}

function go2jsOSGetwd() {
	try {
		return require("process").cwd();
	} catch (err) {
		return "";
	}
}

function go2jsOSChdir(path) {
	try {
		require("process").chdir(String(path));
		return null;
	} catch (err) {
		return go2jsOSError(err);
	}
}

function go2jsOSReadDir(path) {
	try {
		return require("fs").readdirSync(String(path)).sort();
	} catch (err) {
		return null;
	}
}

function go2jsOSError(err) {
	const wrapped = new Error(err.message);

	wrapped.code = err.code;

	return wrapped;
}

function go2jsBytesContains(a, sub) {
	return go2jsBytesIndex(a, sub) >= 0;
}

function go2jsBytesIndex(a, sub) {
	const haystack = go2jsToArray(a);
	const needle = go2jsToArray(sub);

	if (needle.length === 0) {
		return 0;
	}

	for (let index = 0; index + needle.length <= haystack.length; index++) {
		let matched = true;

		for (let offset = 0; offset < needle.length; offset++) {
			if (haystack[index + offset] !== needle[offset]) {
				matched = false;
				break;
			}
		}

		if (matched) {
			return index;
		}
	}

	return -1;
}

function go2jsBytesLastIndex(a, sub) {
	const haystack = go2jsToArray(a);
	const needle = go2jsToArray(sub);

	if (needle.length === 0) {
		return haystack.length;
	}

	for (let index = haystack.length - needle.length; index >= 0; index--) {
		let matched = true;

		for (let offset = 0; offset < needle.length; offset++) {
			if (haystack[index + offset] !== needle[offset]) {
				matched = false;
				break;
			}
		}

		if (matched) {
			return index;
		}
	}

	return -1;
}

function go2jsBytesCount(a, sub) {
	if (go2jsToArray(sub).length === 0) {
		return go2jsToArray(a).length + 1;
	}

	let count = 0;
	let index = 0;

	const haystack = go2jsToArray(a);

	while (index < haystack.length) {
		const found = go2jsBytesIndex(haystack.slice(index), sub);

		if (found < 0) {
			break;
		}

		count++;
		index += found + go2jsToArray(sub).length;
	}

	return count;
}

function go2jsBytesContainsAny(a, chars) {
	const set = go2jsToArray(chars);

	return go2jsToArray(a).some(item => set.includes(item));
}

function go2jsBytesIndexAny(a, chars) {
	const set = go2jsToArray(chars);

	return go2jsToArray(a).findIndex(item => set.includes(item));
}

function go2jsBytesIndexByte(a, b) {
	return go2jsToArray(a).indexOf(Number(b) & 255);
}

function go2jsBytesContainsFunc(a, predicate) {
	return go2jsToArray(a).some(predicate);
}

function go2jsBytesEqualFold(a, b) {
	return go2jsBytesToLower(go2jsToArray(a)).join("") === go2jsBytesToLower(go2jsToArray(b)).join("");
}

function go2jsBytesHasPrefix(a, prefix) {
	const haystack = go2jsToArray(a);
	const needle = go2jsToArray(prefix);

	if (needle.length > haystack.length) {
		return false;
	}

	for (let index = 0; index < needle.length; index++) {
		if (haystack[index] !== needle[index]) {
			return false;
		}
	}

	return true;
}

function go2jsBytesHasSuffix(a, suffix) {
	const haystack = go2jsToArray(a);
	const needle = go2jsToArray(suffix);

	if (needle.length > haystack.length) {
		return false;
	}

	return go2jsBytesEqual(a.slice(haystack.length - needle.length), needle);
}

function go2jsBytesJoin(chunks, sep) {
	const out = [];

	for (const chunk of chunks) {
		for (const item of go2jsToArray(chunk)) {
			out.push(item);
		}
	}

	if (sep === undefined || go2jsToArray(sep).length === 0) {
		return out;
	}

	const joined = [];
	const separator = go2jsToArray(sep);

	chunks.forEach((chunk, index) => {
		if (index > 0) {
			joined.push(...separator);
		}

		joined.push(...go2jsToArray(chunk));
	});

	return joined;
}

function go2jsBytesRepeat(b, count) {
	const out = [];

	for (let index = 0; index < count; index++) {
		out.push(...go2jsToArray(b));
	}

	return out;
}

function go2jsBytesReplace(a, old, replacement, n) {
	const haystack = go2jsToArray(a);
	const needle = go2jsToArray(old);
	const limit = n < 0 ? Number.POSITIVE_INFINITY : n;
	const out = [];
	let index = 0;
	let replaced = 0;

	while (index < haystack.length) {
		if (replaced < limit && go2jsBytesEqual(haystack.slice(index, index + needle.length), needle) && needle.length > 0) {
			out.push(...go2jsToArray(replacement));
			index += needle.length;
			replaced++;
			continue;
		}

		out.push(haystack[index]);
		index++;
	}

	return out;
}

function go2jsBytesReplaceAll(a, old, replacement) {
	return go2jsBytesReplace(a, old, replacement, -1);
}

function go2jsBytesSplit(a, sep) {
	return go2jsRawText(a)
		.split(go2jsRawText(sep))
		.map(part => Array.from(new TextEncoder().encode(part)));
}

function go2jsBytesSplitN(a, sep, n) {
	const parts = go2jsBytesSplit(a, sep);

	if (n < 0) {
		return parts;
	}

	return parts.slice(0, n);
}

function go2jsBytesTitle(a) {
	return go2jsStringToBytes(go2jsStringsTitle(go2jsRawText(a)));
}


function go2jsBytesToLower(a) {
	return go2jsStringToBytes(go2jsRawText(a).toLowerCase());
}

function go2jsBytesToUpper(a) {
	return go2jsStringToBytes(go2jsRawText(a).toUpperCase());
}

function go2jsBytesTrim(a, cutset) {
	return go2jsStringToBytes(go2jsStringsTrim(go2jsRawText(a), cutset));
}

function go2jsBytesTrimSpace(a) {
	return go2jsStringToBytes(go2jsRawText(a).replace(/^\s+/, "").replace(/\s+$/, ""));
}

function go2jsBytesTrimPrefix(a, prefix) {
	const text = go2jsRawText(a);
	const head = go2jsRawText(prefix);

	return go2jsBytesHasPrefix(a, prefix) ? go2jsStringToBytes(text.slice(head.length)) : go2jsStringToBytes(text);
}

function go2jsBytesTrimSuffix(a, suffix) {
	const text = go2jsRawText(a);
	const tail = go2jsRawText(suffix);

	return go2jsBytesHasSuffix(a, suffix) ? go2jsStringToBytes(text.slice(0, text.length - tail.length)) : go2jsStringToBytes(text);
}

function go2jsBytesFields(a) {
	return go2jsBytesSplit(go2jsBytesTrimSpace(a), go2jsStringToBytes(" "));
}

function go2jsBytesNewBuffer(data) {
	return new go2jsBytesBuffer(go2jsToArray(data));
}

function go2jsBytesNewBufferString(text) {
	return new go2jsBytesBuffer(go2jsStringToBytes(go2jsStringify(text)));
}

function go2jsBytesNewReader(data) {
	return go2jsStringsNewReader(go2jsRawText(data));
}

function go2jsBytesRunes(a) {
	return go2jsStringToRunes(go2jsRawText(a));
}

function go2jsBytesCutPrefix(a, prefix) {
	if (!go2jsBytesHasPrefix(a, prefix)) {
		return [go2jsToArray(a), false];
	}

	return [go2jsToArray(a).slice(go2jsToArray(prefix).length), true];
}

function go2jsBytesCutSuffix(a, suffix) {
	if (!go2jsBytesHasSuffix(a, suffix)) {
		return [go2jsToArray(a), false];
	}

	return [go2jsToArray(a).slice(0, go2jsToArray(a).length - go2jsToArray(suffix).length), true];
}

function go2jsBytesCompare(a, b) {
	return go2jsCompareValues(go2jsToArray(a), go2jsToArray(b));
}

function go2jsCallMethod(target, method, ...args) {
	if (target === null || target === undefined) {
		throw new TypeError("call of method " + method + " on nil interface");
	}

	if (target.__go2js_interface === true) {
		return go2jsInterfaceCall(target, method, ...args);
	}

	const fn = target[method];

	if (typeof fn !== "function") {
		throw new TypeError("method " + method + " is not implemented");
	}

	return fn.apply(target, args);
}

function go2jsIOWriteString(writer, value) {
	return go2jsCallMethod(writer, "Write", go2jsStringToBytes(value));
}

function go2jsIONopCloser(reader) {
	return go2jsInterface({
		Read(buffer) {
			return go2jsCallMethod(reader, "Read", buffer);
		},
		Close() {
			return null;
		},
	}, "io.ReadCloser");
}

function go2jsIOMultiReader(...readers) {
	let index = 0;

	return go2jsInterface({
		Read(buffer) {
			while (index < readers.length) {
				const result = go2jsCallMethod(readers[index], "Read", buffer);

				if (Array.isArray(result) && result[0] > 0) {
					return result;
				}

				index++;
			}

			return [0, go2jsIOEOFError()];
		},
	}, "io.Reader");
}

function go2jsIOMultiWriter(...writers) {
	return go2jsInterface({
		Write(buffer) {
			for (const writer of writers) {
				go2jsCallMethod(writer, "Write", buffer);
			}

			return go2jsToArray(buffer).length;
		},
	}, "io.Writer");
}

function go2jsIOLimitReader(reader, limit) {
	let remaining = Number(limit);

	return go2jsInterface({
		Read(buffer) {
			if (remaining <= 0) {
				return [0, go2jsIOEOFError()];
			}

			const window = buffer.length > remaining ? buffer.slice(0, remaining) : buffer;
			const result = go2jsCallMethod(reader, "Read", window);

			if (Array.isArray(result) && typeof result[0] === "number") {
				const count = result[0] > remaining ? remaining : result[0];

				remaining -= count;

				for (let index = 0; index < count; index++) {
					buffer[index] = window[index];
				}

				return [count, result[1]];
			}

			return result;
		},
	}, "io.Reader");
}

function go2jsIOTeeReader(reader, writer) {
	return go2jsInterface({
		Read(buffer) {
			const result = go2jsCallMethod(reader, "Read", buffer);

			if (Array.isArray(result) && result[0] > 0) {
				go2jsCallMethod(writer, "Write", go2jsToArray(buffer).slice(0, result[0]));
			}

			return result;
		},
	}, "io.Reader");
}

function go2jsIOCopy(destination, source) {
	const buffer = new Array(32 * 1024).fill(0);
	let total = 0;

	for (;;) {
		const result = go2jsCallMethod(source, "Read", buffer);

		const read = Array.isArray(result) ? result[0] : Number(result);

		if (!read || read <= 0) {
			break;
		}

		go2jsCallMethod(destination, "Write", buffer.slice(0, read));
		total += read;
	}

	return [total, null];
}

function go2jsIOEOFError() {
	return "EOF";
}

function go2jsSentinelError(message) {
	return function go2jsSentinelErrorValue() {
		return message;
	};
}

function go2jsIOEOF() {
	return "EOF";
}

function go2jsIODiscard() {
	return go2jsInterface({
		Write(buffer) {
			return go2jsToArray(buffer).length;
		},
	}, "io.Writer");
}

function go2jsIOReadFull(reader, buffer) {
	let total = 0;
	const target = go2jsToArray(buffer);

	while (total < target.length) {
		const chunk = new Array(target.length - total).fill(0);
		const result = go2jsCallMethod(reader, "Read", chunk);

		const read = Array.isArray(result) ? result[0] : Number(result);

		if (!read || read <= 0) {
			return [total, go2jsIOUnexpectedEOFErorror()];
		}

		for (let index = 0; index < read; index++) {
			target[total + index] = chunk[index];
		}

		total += read;
	}

	return [total, null];
}

function go2jsIOUnexpectedEOFErorror() {
	return new Error("unexpected EOF");
}

function go2jsErrorsNew(value) {
	return new Error(go2jsStringify(value));
}

function go2jsErrorsAs(err, target) {
	if (err === null || err === undefined) {
		return null;
	}

	if (target === null || typeof target !== "object") {
		return err;
	}

	const name = target.__go2js_error_name;

	if (name !== undefined && err.name === name) {
		return err;
	}

	return null;
}

function go2jsRandPerm(n) {
	const out = Array.from({length: n}, (_, index) => index);

	for (let index = out.length - 1; index > 0; index--) {
		const swap = go2jsRandIntn(index + 1);
		const held = out[index];

		out[index] = out[swap];
		out[swap] = held;
	}

	return out;
}

function go2jsRandUint32() {
	return go2jsRandNext();
}

function go2jsTimeSince(t) {
	return go2jsTimeSub(go2jsTimeNow(), t);
}

function go2jsTimeUntil(t) {
	return go2jsTimeSub(t, go2jsTimeNow());
}

function go2jsTimeNow() {
	return new Date();
}

function go2jsTimeSleep(d) {
	const ms = go2jsDurationNanos(d) / 1e6;

	if (ms > 0) {
		try {
			require("child_process").execFileSync("sleep", [String(ms / 1000)]);
		} catch (err) {
		}
	}
}

function go2jsTimeUnix(value) {
	return go2jsTimeValue(value).getTime() / 1000;
}

function go2jsTimeParse(layout, value) {
	return [go2jsTimeValue(new Date(String(value))), null];
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

function go2jsTimeAdd(base, duration) {
	const date = go2jsTimeDateOf(base);

	return go2jsTimeValue(new Date(date.getTime() + go2jsDurationNanos(duration) / 1000000));
}

function go2jsTimeSub(left, right) {
	return go2jsDuration(
		(go2jsTimeDateOf(left).getTime() - go2jsTimeDateOf(right).getTime()) * 1000000
	);
}

function go2jsTimeAfter(d) {
	const channel = go2jsChannel(1);

	go2jsChanSend(channel, go2jsTimeAdd(go2jsTimeNow(), d));

	return channel;
}

function go2jsTimeTick(d) {
	return go2jsTimeNow();
}

function go2jsRegexpMatchString(pattern, value) {
	return new RegExp(go2jsStringify(pattern)).test(go2jsStringify(value));
}



function go2jsRegexpFindAllString(pattern, value, limit) {
	const source = go2jsStringify(value);
	const regex = new RegExp(go2jsStringify(pattern), "g");
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

function go2jsRegexpReplaceAllString(pattern, value, replacement) {
	return go2jsStringify(value).replace(
		new RegExp(go2jsStringify(pattern), "g"),
		go2jsRegexpExpand(go2jsStringify(replacement))
	);
}

function go2jsRegexpExpand(replacement) {
	return go2jsStringify(replacement)
		.replace(/\$(\d+)/g, (match, index) => "$" + (Number(index) === 0 ? "&" : index))
		.replace(/\$\{(\w+)\}/g, (match, name) => "$" + (name === "0" ? "&" : name));
}

function go2jsRegexpSplit(pattern, value, limit) {
	const parts = go2jsStringify(value).split(new RegExp(go2jsStringify(pattern)));

	if (limit === undefined || limit < 0) {
		return parts;
	}

	return parts.slice(0, limit);
}

function go2jsRegexpQuoteMeta(value) {
	return go2jsStringify(value).replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function go2jsCodePoints(value) {
	if (Array.isArray(value)) {
		return value.map(item => Number(item));
	}

	if (typeof value === "string") {
		return Array.from(value).map(char => char.codePointAt(0));
	}

	if (value === null || value === undefined) {
		return [];
	}

	return Array.from(go2jsRawText(value)).map(char => char.codePointAt(0));
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
		return go2jsRawText(value.get());
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
	return Array.from(go2jsRawText(value)).map(char => char.codePointAt(0));
}

function go2jsRunesToString(value) {
	if (typeof value === "string") {
		return value;
	}

	if (value === null || value === undefined) {
		return "";
	}

	return Array.from(value).map(item => String.fromCodePoint(Number(item))).join("");
}

function go2jsBytesToString(value) {
	if (typeof value === "string") {
		return value;
	}

	if (value === null || value === undefined) {
		return "";
	}

	if (Array.isArray(value)) {
		return new TextDecoder().decode(Uint8Array.from(value.map(item => Number(item) & 255)));
	}

	if (value instanceof Uint8Array) {
		return new TextDecoder().decode(value);
	}

	if (typeof value.String === "function") {
		return value.String();
	}

	return String(value);
}

const go2jsContextStates = new WeakMap();

function go2jsContextState(cancelled, err, deadline) {
	return {
		done: cancelled === true,
		err: err || null,
		values: new go2jsNativeMap(),
		deadline: deadline || 0,
		callbacks: []
	};
}

function go2jsContextWrap(state) {
	const wrapped = go2jsInterface({
		Err() {
			return state.err;
		},
		Done() {
			if (!state.done) {
				return null;
			}

			const channel = go2jsChannel(0);
			channel.closed = true;

			return channel;
		},
		Deadline() {
			if (state.deadline === 0) {
				return [null, false];
			}

			return [go2jsTimeValue(new Date(state.deadline)), true];
		},
		Value(key) {
			let current = state;

			while (current !== null && current !== undefined) {
				if (current.values !== null && current.values !== undefined &&
					current.values.has(key)) {
					return current.values.get(key);
				}

				current = current.parent;
			}

			return null;
		},
		String() {
			return "context.Background";
		}
	}, "context.Context");

	go2jsContextStates.set(wrapped, state);

	return wrapped;
}

function go2jsContextBackground() {
	return go2jsContextWrap(go2jsContextState(false, null, 0));
}

function go2jsContextWithValue(parent, key, value) {
	const base = go2jsContextStateOf(parent);
	const child = go2jsContextState(base.done, base.err, base.deadline);
	child.values = new go2jsNativeMap(base.values);
	child.values.set(key, value);
	child.parent = base;

	return go2jsContextWrap(child);
}

function go2jsContextWithCancel(parent) {
	const base = go2jsContextStateOf(parent);
	const child = go2jsContextState(false, null, base.deadline);
	child.parent = base;
	child.cancel = go2jsCancelFunc(child);

	return [go2jsContextWrap(child), go2jsCancelFunc(child)];
}

function go2jsCancelFunc(state) {
	const cancel = () => {
		if (state.done) {
			return;
		}

		state.done = true;
		state.err = go2jsContextCanceled();

		for (const fn of state.callbacks || []) {
			fn(state.err);
		}
	};

	cancel.__go2js_context_cancel = true;

	return cancel;
}

function go2jsContextWithDeadline(parent, deadline) {
	const base = go2jsContextStateOf(parent);
	const at = Number(deadline);
	const child = go2jsContextState(at <= Date.now(), null, at === 0 ? 0 : at);
	child.parent = base;

	if (child.done) {
		child.err = go2jsContextDeadlineExceeded();
	}

	return [go2jsContextWrap(child), go2jsCancelFunc(child)];
}

function go2jsContextWithTimeout(parent, timeout) {
	return go2jsContextWithDeadline(parent, Date.now() + Number(timeout));
}

function go2jsContextStateOf(ctx) {
	if (ctx !== null && ctx !== undefined && go2jsContextStates.has(ctx)) {
		return go2jsContextStates.get(ctx);
	}

	return go2jsContextState(false, null, 0);
}

function go2jsContextAfterFunc(ctx, fn) {
	go2jsContextStateOf(ctx).callbacks.push(fn);

	return 2;
}

var go2jsContextCanceledError = null;
var go2jsContextDeadlineExceededError = null;

function go2jsContextCanceled() {
	if (go2jsContextCanceledError === null) {
		go2jsContextCanceledError = go2jsInterface(new Error("context canceled"), "error");
	}

	return go2jsContextCanceledError;
}

function go2jsContextDeadlineExceeded() {
	if (go2jsContextDeadlineExceededError === null) {
		go2jsContextDeadlineExceededError = go2jsInterface(new Error("context deadline exceeded"), "error");
	}

	return go2jsContextDeadlineExceededError;
}

function go2jsContextCause(ctx) {
	return go2jsContextStateOf(ctx).err;
}

function go2jsAtomicTyped(initial, numeric) {
	const cell = {value: initial === undefined ? (numeric ? 0 : null) : initial};

	cell.__go2js_atomic = true;

	return cell;
}

function go2jsAtomicInt64Type() {
	return go2jsAtomicTyped(0, true);
}

function go2jsAtomicInt32Type() {
	return go2jsAtomicTyped(0, true);
}

function go2jsAtomicUint64Type() {
	return go2jsAtomicTyped(0, true);
}

function go2jsAtomicUint32Type() {
	return go2jsAtomicTyped(0, true);
}

function go2jsAtomicBoolType() {
	return go2jsAtomicTyped(false, false);
}

function go2jsAtomicValueType() {
	return go2jsAtomicTyped(null, false);
}

function go2jsAtomicTypeAdd(cell, delta) {
	cell.value = Number(cell.value) + Number(delta);
	return cell.value;
}

function go2jsAtomicTypeLoad(cell) {
	return cell.value;
}

function go2jsAtomicTypeStore(cell, value) {
	cell.value = value;
}

function go2jsAtomicTypeSwap(cell, value) {
	const previous = cell.value;
	cell.value = value;
	return previous;
}

function go2jsAtomicTypeCompareAndSwap(cell, oldValue, newValue) {
	if (cell.value !== oldValue) {
		return false;
	}

	cell.value = newValue;
	return true;
}

function go2jsAtomicBoolTypeLoad(cell) {
	return cell.value;
}

function go2jsAtomicBoolTypeStore(cell, value) {
	cell.value = Boolean(value);
}

function go2jsAtomicBoolTypeSwap(cell, value) {
	const previous = cell.value;
	cell.value = Boolean(value);
	return previous;
}

function go2jsAtomicBoolTypeCompareAndSwap(cell, oldValue, newValue) {
	if (cell.value !== Boolean(oldValue)) {
		return false;
	}

	cell.value = Boolean(newValue);
	return true;
}

function go2jsAtomicValueTypeLoad(cell) {
	return cell.value;
}

function go2jsAtomicValueTypeStore(cell, value) {
	cell.value = value;
}

function go2jsAtomicValueTypeSwap(cell, value) {
	const previous = cell.value;
	cell.value = value;
	return previous;
}

function go2jsAtomicValueTypeCompareAndSwap(cell, oldValue, newValue) {
	if (cell.value !== oldValue) {
		return false;
	}

	cell.value = newValue;
	return true;
}

function go2jsFlagState() {
	if (go2jsFlagState.current === undefined) {
		go2jsFlagState.current = {
			values: new go2jsNativeMap(),
			order: [],
			args: [],
			parsed: false
		};
	}

	return go2jsFlagState.current;
}

function go2jsFlagDeclare(name, usage, value, target) {
	const state = go2jsFlagState();

	if (!state.values.has(name)) {
		state.order.push(name);
	}

	const entry = {name: name, usage: usage, value: value, target: target || null};

	state.values.set(name, entry);

	return entry;
}

function go2jsFlagBind(entry, text) {
	if (typeof entry.value === "boolean") {
		entry.value = text !== "false" && text !== "0" && text !== "FALSE" && text !== "False";
	} else if (typeof entry.value === "number") {
		entry.value = Number(text) | 0;
	} else if (typeof entry.value === "object" && entry.value !== null) {
		entry.value = go2jsDuration(text);
	} else {
		entry.value = text;
	}

	const target = entry.target;

	if (target !== null && target !== undefined) {
		if (target.__go2js_pointer === true) {
			target.set(entry.value);
		} else {
			target.value = entry.value;
		}
	}
}

function go2jsFlagString(name, value, usage) {
	go2jsFlagDeclare(name, usage, value === undefined ? "" : String(value));
	return go2jsFlagCell(go2jsFlagState().values.get(name));
}

function go2jsFlagStringVar(target, name, value, usage) {
	const text = value === undefined ? "" : String(value);
	go2jsFlagDeclare(name, usage, text, target);
	go2jsFlagSetValue(target, text);
}

function go2jsFlagInt(name, value, usage) {
	go2jsFlagDeclare(name, usage, Number(value) | 0);
	return go2jsFlagCell(go2jsFlagState().values.get(name));
}

function go2jsFlagIntVar(target, name, value, usage) {
	const number = Number(value) | 0;
	go2jsFlagDeclare(name, usage, number, target);
	go2jsFlagSetValue(target, number);
}

function go2jsFlagInt64(name, value, usage) {
	go2jsFlagDeclare(name, usage, Number(value) | 0);
	return go2jsFlagCell(go2jsFlagState().values.get(name));
}

function go2jsFlagInt64Var(target, name, value, usage) {
	const number = Number(value) | 0;
	go2jsFlagDeclare(name, usage, number, target);
	go2jsFlagSetValue(target, number);
}

function go2jsFlagUint64(name, value, usage) {
	return go2jsFlagInt64(name, value, usage);
}

function go2jsFlagUint64Var(target, name, value, usage) {
	go2jsFlagInt64Var(target, name, value, usage);
}

function go2jsFlagFloat64(name, value, usage) {
	go2jsFlagDeclare(name, usage, Number(value));
	return go2jsFlagCell(go2jsFlagState().values.get(name));
}

function go2jsFlagFloat64Var(target, name, value, usage) {
	const number = Number(value);
	go2jsFlagDeclare(name, usage, number, target);
	go2jsFlagSetValue(target, number);
}

function go2jsFlagBool(name, value, usage) {
	go2jsFlagDeclare(name, usage, Boolean(value));
	return go2jsFlagCell(go2jsFlagState().values.get(name));
}

function go2jsFlagBoolVar(target, name, value, usage) {
	const flag = Boolean(value);
	go2jsFlagDeclare(name, usage, flag, target);
	go2jsFlagSetValue(target, flag);
}

function go2jsFlagDuration(name, value, usage) {
	go2jsFlagDeclare(name, usage, go2jsDuration(value === undefined ? 0 : value));
	return go2jsFlagCell(go2jsFlagState().values.get(name));
}

function go2jsFlagDurationVar(target, name, value, usage) {
	const duration = value === undefined ? 0 : value;
	go2jsFlagDeclare(name, usage, go2jsDuration(duration), target);
	go2jsFlagSetValue(target, go2jsDuration(duration));
}

function go2jsFlagParse(argumentsList) {
	const state = go2jsFlagState();
	const list = argumentsList === undefined
		? (typeof process !== "undefined" ? process.argv.slice(2) : [])
		: Array.from(argumentsList);

	state.args = [];
	state.parsed = true;

	for (let index = 0; index < list.length; index++) {
		const arg = String(list[index]);

		if (arg === "--") {
			state.args = state.args.concat(list.slice(index + 1).map(String));
			break;
		}

		if (!arg.startsWith("-") || arg === "-") {
			state.args.push(arg);
			continue;
		}

		let name = arg.replace(/^--?/, "");
		let text;

		const equals = name.indexOf("=");

		if (equals !== -1) {
			text = name.slice(equals + 1);
			name = name.slice(0, equals);
		}

		const entry = state.values.get(name);

		if (entry === undefined) {
			continue;
		}

		if (text === undefined) {
			if (index + 1 < list.length && !String(list[index + 1]).startsWith("-")) {
				text = String(list[++index]);
			} else {
				text = "true";
			}
		}

		go2jsFlagBind(entry, text);
	}
}

function go2jsFlagParsed() {
	return go2jsFlagState().parsed;
}

function go2jsFlagArgs() {
	return go2jsFlagState().args.slice();
}

function go2jsFlagNArg() {
	return go2jsFlagState().args.length;
}

function go2jsFlagNFlag() {
	return go2jsFlagState().order.length;
}

function go2jsFlagLookup(name) {
	const entry = go2jsFlagState().values.get(String(name));

	if (entry === undefined) {
		return go2jsInterface(null, "*flag.Flag");
	}

	return go2jsInterface({
		Name() {
			return entry.name;
		},
		Usage() {
			return entry.usage;
		},
		ValueString() {
			return String(entry.value);
		},
		String() {
			return String(entry.value);
		}
	}, "*flag.Flag");
}

function go2jsFlagSet(name, value) {
	const entry = go2jsFlagState().values.get(String(name));

	if (entry === undefined) {
		return null;
	}

	go2jsFlagBind(entry, typeof value === "string" ? value : String(value));

	return null;
}

function go2jsFlagCell(entry) {
	return go2jsPtr(
		() => entry.value,
		value => go2jsFlagBind(entry, typeof value === "string" ? value : String(value))
	);
}

function go2jsFlagSetValue(target, value) {
	if (target === null || target === undefined) {
		return;
	}

	if (target.__go2js_pointer === true) {
		target.set(value);
		return;
	}

	target.value = value;
}

function go2jsFlagVisit(fn) {
	for (const name of go2jsFlagState().order) {
		fn(go2jsFlagLookup(name));
	}
}

function go2jsFlagVar(value, name, usage) {
	go2jsFlagDeclare(String(name), usage, value);
}

function go2jsFlagPrintDefaults() {
	const state = go2jsFlagState();

	for (const name of state.order) {
		const entry = state.values.get(name);
		process.stderr.write("  -" + name + " " + entry.usage + "\n");
	}
}

function go2jsSHA256Constants() {
	return [
		0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
		0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
		0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
		0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
		0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
		0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
		0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
		0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2
	];
}

function go2jsRotr(value, bits) {
	return (value >>> bits) | (value << (32 - bits));
}

function go2jsSHA256Block(state, block) {
	const k = go2jsSHA256Constants();
	const w = new Array(64);

	for (let index = 0; index < 16; index++) {
		const offset = index * 4;
		w[index] = ((block[offset] << 24) | (block[offset + 1] << 16) | (block[offset + 2] << 8) | block[offset + 3]) >>> 0;
	}

	for (let index = 16; index < 64; index++) {
		const s0 = go2jsRotr(w[index - 15], 7) ^ go2jsRotr(w[index - 15], 18) ^ (w[index - 15] >>> 3);
		const s1 = go2jsRotr(w[index - 2], 17) ^ go2jsRotr(w[index - 2], 19) ^ (w[index - 2] >>> 10);
		w[index] = (w[index - 16] + s0 + w[index - 7] + s1) >>> 0;
	}

	let [a, b, c, d, e, f, g, h] = state;

	for (let index = 0; index < 64; index++) {
		const s1 = go2jsRotr(e, 6) ^ go2jsRotr(e, 11) ^ go2jsRotr(e, 25);
		const ch = (e & f) ^ (~e & g);
		const temp1 = (h + s1 + ch + k[index] + w[index]) >>> 0;
		const s0 = go2jsRotr(a, 2) ^ go2jsRotr(a, 13) ^ go2jsRotr(a, 22);
		const maj = (a & b) ^ (a & c) ^ (b & c);
		const temp2 = (s0 + maj) >>> 0;

		h = g;
		g = f;
		f = e;
		e = (d + temp1) >>> 0;
		d = c;
		c = b;
		b = a;
		a = (temp1 + temp2) >>> 0;
	}

	const next = [a, b, c, d, e, f, g, h];

	for (let index = 0; index < 8; index++) {
		state[index] = (state[index] + next[index]) >>> 0;
	}
}

function go2jsSHA256Init() {
	return [0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a, 0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19];
}

function go2jsSHA256Pad(data, absorbed) {
	const bitLength = (data.length + (absorbed || 0)) * 8;
	const padded = data.slice();
	padded.push(0x80);

	while (padded.length % 64 !== 56) {
		padded.push(0);
	}

	const high = Math.floor(bitLength / 0x100000000);
	const low = bitLength >>> 0;

	padded.push((high >>> 24) & 255, (high >>> 16) & 255, (high >>> 8) & 255, high & 255);
	padded.push((low >>> 24) & 255, (low >>> 16) & 255, (low >>> 8) & 255, low & 255);

	return padded;
}

function go2jsSHA256Digest(data, state, absorbed) {
	const current = state === undefined ? go2jsSHA256Init() : state.slice();
	const padded = go2jsSHA256Pad(Array.from(data, item => Number(item) & 255), absorbed);

	for (let offset = 0; offset < padded.length; offset += 64) {
		go2jsSHA256Block(current, padded.slice(offset, offset + 64));
	}

	const out = [];

	for (let index = 0; index < 8; index++) {
		const word = current[index];
		out.push((word >>> 24) & 255, (word >>> 16) & 255, (word >>> 8) & 255, word & 255);
	}

	return out;
}

function go2jsSHA256Sum256(data) {
	return go2jsSHA256Digest(go2jsToArray(data));
}



function go2jsSHA256New() {
	const state = go2jsSHA256Init();

	return go2jsInterface({
		Size() {
			return 32;
		},
		BlockSize() {
			return 64;
		},
		Reset() {
			const fresh = go2jsSHA256Init();
			for (let index = 0; index < 8; index++) {
				state[index] = fresh[index];
			}

			state.buffer = [];
			state.hashed = go2jsSHA256Init();
			state.length = 0;
		},
		Write(data) {
			const bytes = Array.from(go2jsToArray(data), item => Number(item) & 255);
			const combined = go2jsSHA256Buffered(state, bytes);
			return bytes.length;
		},
		Sum(target) {
			return go2jsToArray(target).concat(
				go2jsSHA256Digest(state.buffer || [], state.hashed || go2jsSHA256Init(), state.length || 0));
		}
	}, "hash.Hash");
}

function go2jsSHA256Buffered(state, bytes) {
	if (state.buffer === undefined) {
		state.buffer = [];
		state.hashed = go2jsSHA256Init();
		state.length = 0;
	}

	const combined = state.buffer.concat(bytes);

	for (let offset = 0; offset + 64 <= combined.length; offset += 64) {
		go2jsSHA256Block(state.hashed, combined.slice(offset, offset + 64));
		state.length += 64;
	}

	state.buffer = combined.slice(state.length);
	return state;
}

function go2jsRotl(value, bits) {
	return (value << bits) | (value >>> (32 - bits));
}

function go2jsSHA1Block(state, block) {
	const w = new Array(80);

	for (let index = 0; index < 16; index++) {
		const offset = index * 4;
		w[index] = ((block[offset] << 24) | (block[offset + 1] << 16) | (block[offset + 2] << 8) | block[offset + 3]) >>> 0;
	}

	for (let index = 16; index < 80; index++) {
		w[index] = go2jsRotl(w[index - 3] ^ w[index - 8] ^ w[index - 14] ^ w[index - 16], 1);
	}

	let [a, b, c, d, e] = state;

	for (let index = 0; index < 80; index++) {
		let f;
		let k;

		if (index < 20) {
			f = (b & c) | (~b & d);
			k = 0x5a827999;
		} else if (index < 40) {
			f = b ^ c ^ d;
			k = 0x6ed9eba1;
		} else if (index < 60) {
			f = (b & c) | (b & d) | (c & d);
			k = 0x8f1bbcdc;
		} else {
			f = b ^ c ^ d;
			k = 0xca62c1d6;
		}

		const temp = (go2jsRotl(a, 5) + f + e + k + w[index]) >>> 0;
		e = d;
		d = c;
		c = go2jsRotl(b, 30);
		b = a;
		a = temp;
	}

	const next = [a, b, c, d, e];

	for (let index = 0; index < 5; index++) {
		state[index] = (state[index] + next[index]) >>> 0;
	}
}

function go2jsSHA1Init() {
	return [0x67452301, 0xefcdab89, 0x98badcfe, 0x10325476, 0xc3d2e1f0];
}

function go2jsSHA1Digest(data) {
	const bitLength = Array.from(data, item => Number(item) & 255).length * 8;
	const bytes = Array.from(data, item => Number(item) & 255);
	const padded = bytes.slice();
	padded.push(0x80);

	while (padded.length % 64 !== 56) {
		padded.push(0);
	}

	const high = Math.floor(bitLength / 0x100000000);
	const low = bitLength >>> 0;

	padded.push((high >>> 24) & 255, (high >>> 16) & 255, (high >>> 8) & 255, high & 255);
	padded.push((low >>> 24) & 255, (low >>> 16) & 255, (low >>> 8) & 255, low & 255);

	const state = go2jsSHA1Init();

	for (let offset = 0; offset < padded.length; offset += 64) {
		go2jsSHA1Block(state, padded.slice(offset, offset + 64));
	}

	const out = [];

	for (let index = 0; index < 5; index++) {
		const word = state[index];
		out.push((word >>> 24) & 255, (word >>> 16) & 255, (word >>> 8) & 255, word & 255);
	}

	return out;
}

function go2jsSHA1Sum(data) {
	return go2jsSHA1Digest(go2jsToArray(data));
}



function go2jsSHA1New() {
	return go2jsSimpleHash(go2jsSHA1Digest, 20);
}

function go2jsMD5Constants() {
	if (go2jsMD5Constants.cache === undefined) {
		const k = new Array(64);

		for (let index = 0; index < 64; index++) {
			k[index] = Math.floor(Math.abs(Math.sin(index + 1)) * 0x100000000);
		}

		go2jsMD5Constants.cache = k;
	}

	return go2jsMD5Constants.cache;
}

function go2jsMD5Rotl(value, bits) {
	return (value << bits) | (value >>> (32 - bits));
}

function go2jsMD5Block(state, block) {
	const k = go2jsMD5Constants();
	const shifts = [7, 12, 17, 22, 7, 12, 17, 22, 7, 12, 17, 22, 7, 12, 17, 22,
		5, 9, 14, 20, 5, 9, 14, 20, 5, 9, 14, 20, 5, 9, 14, 20,
		4, 11, 16, 23, 4, 11, 16, 23, 4, 11, 16, 23, 4, 11, 16, 23,
		6, 10, 15, 21, 6, 10, 15, 21, 6, 10, 15, 21, 6, 10, 15, 21];

	const m = new Array(16);

	for (let index = 0; index < 16; index++) {
		const offset = index * 4;
		m[index] = (block[offset] | (block[offset + 1] << 8) | (block[offset + 2] << 16) | (block[offset + 3] << 24)) >>> 0;
	}

	let [a, b, c, d] = state;

	for (let index = 0; index < 64; index++) {
		let f;
		let g;

		if (index < 16) {
			f = (b & c) | (~b & d);
			g = index;
		} else if (index < 32) {
			f = (d & b) | (~d & c);
			g = (5 * index + 1) % 16;
		} else if (index < 48) {
			f = b ^ c ^ d;
			g = (3 * index + 5) % 16;
		} else {
			f = c ^ (b | ~d);
			g = (7 * index) % 16;
		}

		const temp = d;
		d = c;
		c = b;

		const sum = (a + f + k[index] + m[g]) >>> 0;
		b = (b + go2jsMD5Rotl(sum, shifts[index])) >>> 0;
		a = temp;
	}

	const next = [a, b, c, d];

	for (let index = 0; index < 4; index++) {
		state[index] = (state[index] + next[index]) >>> 0;
	}
}

function go2jsMD5Digest(data) {
	const bytes = Array.from(data, item => Number(item) & 255);
	const bitLength = bytes.length * 8;
	const padded = bytes.slice();
	padded.push(0x80);

	while (padded.length % 64 !== 56) {
		padded.push(0);
	}

	const low = bitLength >>> 0;
	const high = Math.floor(bitLength / 0x100000000);

	padded.push(low & 255, (low >>> 8) & 255, (low >>> 16) & 255, (low >>> 24) & 255);
	padded.push(high & 255, (high >>> 8) & 255, (high >>> 16) & 255, (high >>> 24) & 255);

	const state = [0x67452301, 0xefcdab89, 0x98badcfe, 0x10325476];

	for (let offset = 0; offset < padded.length; offset += 64) {
		go2jsMD5Block(state, padded.slice(offset, offset + 64));
	}

	const out = [];

	for (let index = 0; index < 4; index++) {
		const word = state[index];
		out.push(word & 255, (word >>> 8) & 255, (word >>> 16) & 255, (word >>> 24) & 255);
	}

	return out;
}

function go2jsMD5Sum(data) {
	return go2jsMD5Digest(go2jsToArray(data));
}



function go2jsMD5New() {
	return go2jsSimpleHash(go2jsMD5Digest, 16);
}

function go2jsSimpleHash(digest, size) {
	const chunks = [];

	return go2jsInterface({
		Size() {
			return size;
		},
		BlockSize() {
			return 64;
		},
		Reset() {
			chunks.length = 0;
		},
		Write(data) {
			const bytes = Array.from(go2jsToArray(data), item => Number(item) & 255);
			for (const byte of bytes) {
				chunks.push(byte);
			}
			return bytes.length;
		},
		Sum(target) {
			return go2jsToArray(target).concat(digest(chunks));
		}
	}, "hash.Hash");
}

function go2jsHashCall(hash, method, ...args) {
	if (hash !== null && hash !== undefined && hash.__go2js_interface === true) {
		return go2jsInterfaceCall(hash, method, ...args);
	}

	return hash[method](...args);
}

function go2jsHashBytes(hashFn, bytes) {
	const hash = hashFn();
	go2jsHashCall(hash, "Write", bytes);
	return go2jsHashCall(hash, "Sum", []);
}

function go2jsHMACNew(hashFn, key) {
	const blockSize = go2jsHashCall(hashFn(), "BlockSize");
	let block = Array.from(go2jsToArray(key), item => Number(item) & 255);

	if (block.length > blockSize) {
		block = go2jsHashBytes(hashFn, block);
	}

	while (block.length < blockSize) {
		block.push(0);
	}

	const innerKey = new Array(blockSize);
	const outerKey = new Array(blockSize);

	for (let index = 0; index < blockSize; index++) {
		innerKey[index] = block[index] ^ 0x36;
		outerKey[index] = block[index] ^ 0x5c;
	}

	const chunks = [];

	return go2jsInterface({
		Size() {
			return go2jsHashBytes(hashFn, []).length;
		},
		BlockSize() {
			return blockSize;
		},
		Reset() {
			chunks.length = 0;
		},
		Write(data) {
			const bytes = Array.from(go2jsToArray(data), item => Number(item) & 255);
			for (const byte of bytes) {
				chunks.push(byte);
			}
			return bytes.length;
		},
		Sum(target) {
			const inner = go2jsHashBytes(hashFn, innerKey.concat(chunks));
			return go2jsToArray(target).concat(go2jsHashBytes(hashFn, outerKey.concat(inner)));
		}
	}, "hash.Hash");
}

function go2jsCSVParse(text) {
	const rows = [];
	let row = [];
	let field = "";
	let quoted = false;
	const source = go2jsRawText(text);

	for (let index = 0; index < source.length; index++) {
		const ch = source[index];

		if (quoted) {
			if (ch === '"') {
				if (source[index + 1] === '"') {
					field += '"';
					index++;
				} else {
					quoted = false;
				}
			} else {
				field += ch;
			}
			continue;
		}

		if (ch === '"') {
			quoted = true;
		} else if (ch === ",") {
			row.push(field);
			field = "";
		} else if (ch === "\n") {
			row.push(field);
			rows.push(row);
			row = [];
			field = "";
		} else if (ch !== "\r") {
			field += ch;
		}
	}

	if (field !== "" || row.length > 0) {
		row.push(field);
		rows.push(row);
	}

	return rows;
}

function go2jsCSVRecord(row) {
	return row;
}

function go2jsCSVNewReader(text) {
	const reader = {
		rows: go2jsCSVParse(go2jsRawText(text)),
		index: 0
	};

	reader.Read = function() {
		if (reader.index >= reader.rows.length) {
			return [null, "EOF"];
		}

		return [go2jsCSVRecord(reader.rows[reader.index++]), null];
	};

	reader.ReadAll = function() {
		const out = [];

		for (;;) {
			const result = reader.Read();

			if (result[0] === null) {
				if (result[1] === "EOF") {
					return [out, null];
				}

				return [out, result[1]];
			}

			out.push(result[0]);
		}
	};

	reader.FieldsPerRecord = -1;

	let handle = reader;

	return go2jsPtr(
		() => handle,
		value => {
			handle = value;
		}
	);
}

function go2jsCSVReadAll(source) {
	const reader = go2jsCSVNewReader(source);

	return reader.ReadAll();
}

function go2jsCSVNewWriter(target) {
	const writer = {pending: ""};

	writer.write = function(text) {
		if (target === null || target === undefined) {
			writer.pending += text;
			return;
		}

		const sink = go2jsUnwrap(target);

		if (sink !== null && sink !== undefined && typeof sink.WriteString === "function") {
			sink.WriteString(text);
			return;
		}

		if (sink !== null && sink !== undefined && typeof sink.Write === "function") {
			sink.Write(text);
			return;
		}

		writer.pending += text;
	};

	writer.Write = function(record) {
		const values = go2jsToArray(record);
		const line = values.map(value => {
			const item = go2jsRawText(value);

			return /[",\n\r]/.test(item) ? '"' + item.replace(/"/g, '""') + '"' : item;
		}).join(",") + "\n";

		writer.write(line);

		return null;
	};

	writer.Flush = function() {
		return null;
	};

	writer.Error = function() {
		return null;
	};

	writer.String = function() {
		return writer.pending;
	};

	let handle = writer;

	return go2jsPtr(
		() => handle,
		value => {
			handle = value;
		}
	);
}

function go2jsCSVWriteAll(target, records) {
	const writer = go2jsCSVNewWriter(target);

	for (const record of go2jsToArray(records)) {
		writer.Write(record);
	}

	writer.Flush();

	return null;
}

function go2jsHeapSift(items, index) {
	const less = go2jsHeapLess;

	for (;;) {
		const left = 2 * index + 1;
		const right = left + 1;
		let smallest = index;

		if (left < items.length && less(items[left], items[smallest])) {
			smallest = left;
		}

		if (right < items.length && less(items[right], items[smallest])) {
			smallest = right;
		}

		if (smallest === index) {
			return;
		}

		const swap = items[index];
		items[index] = items[smallest];
		items[smallest] = swap;
		index = smallest;
	}
}

function go2jsHeapLess(left, right) {
	const leftLess = go2jsUnwrap(left);
	const rightLess = go2jsUnwrap(right);

	if (typeof leftLess.Less === "function") {
		return leftLess.Less(rightLess);
	}

	return go2jsCompareValues(leftLess, rightLess) < 0;
}

function go2jsHeapItems(target) {
	let value = target;

	for (let depth = 0; depth < 8; depth++) {
		if (value === null || value === undefined) {
			return [];
		}

		if (value.__go2js_interface === true) {
			value = value.value;
			continue;
		}

		if (value.__go2js_pointer === true) {
			value = value.get();
			continue;
		}

		break;
	}

	if (Array.isArray(value)) {
		return value;
	}

	if (value !== null && value !== undefined && Array.isArray(value.items)) {
		return value.items;
	}

	return [];
}

function go2jsHeapSiftUp(items, index) {
	while (index > 0) {
		const parent = Math.floor((index - 1) / 2);

		if (!go2jsHeapLess(items[index], items[parent])) {
			break;
		}

		const swap = items[index];
		items[index] = items[parent];
		items[parent] = swap;
		index = parent;
	}
}

function go2jsHeapInit(target) {
	const items = go2jsHeapItems(target);

	for (let index = Math.floor(items.length / 2) - 1; index >= 0; index--) {
		go2jsHeapSift(items, index);
	}
}

function go2jsHeapPush(target, value) {
	const items = go2jsHeapItems(target);
	items.push(value);
	go2jsHeapSiftUp(items, items.length - 1);
}

function go2jsHeapPop(target) {
	const items = go2jsHeapItems(target);

	if (items.length === 0) {
		return null;
	}

	const top = items[0];
	const last = items.pop();

	if (items.length > 0) {
		items[0] = last;
		go2jsHeapSift(items, 0);
	}

	return top;
}

function go2jsHeapPeek(target) {
	const items = go2jsHeapItems(target);
	return items.length === 0 ? null : items[0];
}

function go2jsHeapRemove(target, index) {
	const items = go2jsHeapItems(target);
	const removed = items[index];
	items[index] = items[items.length - 1];
	items.pop();
	go2jsHeapSift(items, index);
	return removed;
}

function go2jsHeapFix(target, index) {
	const items = go2jsHeapItems(target);

	if (index > 0 && go2jsHeapLess(items[index], items[Math.floor((index - 1) / 2)])) {
		go2jsHeapSiftUp(items, index);
		return;
	}

	go2jsHeapSift(items, index);
}

function go2jsListNode(list, value) {
	const node = {Value: value, next: null, prev: null, list: list};

	node.Next = function() {
		if (node.next === null || node.next === node.list.root) {
			return null;
		}

		return go2jsListHandle(node.next);
	};

	node.Prev = function() {
		if (node.prev === null || node.prev === node.list.root) {
			return null;
		}

		return go2jsListHandle(node.prev);
	};

	return node;
}

function go2jsListHandle(node) {
	let held = node;

	return go2jsPtr(
		() => held,
		value => {
			held = value;
		}
	);
}

function go2jsListNew() {
	const root = go2jsListNode(null, null);
	const list = {len: 0};

	root.next = root;
	root.prev = root;
	root.list = list;
	list.root = root;

	list.PushBack = function(value) {
		const node = go2jsListNode(list, value);
		node.prev = root.prev;
		node.next = root;
		root.prev.next = node;
		root.prev = node;
		list.len++;
		return go2jsListHandle(node);
	};

	list.PushFront = function(value) {
		const node = go2jsListNode(list, value);
		node.next = root.next;
		node.prev = root;
		root.next.prev = node;
		root.next = node;
		list.len++;
		return go2jsListHandle(node);
	};

	list.InsertBefore = function(mark, value) {
		const pivot = mark.get();
		const node = go2jsListNode(list, value);
		node.prev = pivot.prev;
		node.next = pivot;
		pivot.prev.next = node;
		pivot.prev = node;
		list.len++;
		return go2jsListHandle(node);
	};

	list.InsertAfter = function(mark, value) {
		const pivot = mark.get();
		const node = go2jsListNode(list, value);
		node.next = pivot.next;
		node.prev = pivot;
		pivot.next.prev = node;
		pivot.next = node;
		list.len++;
		return go2jsListHandle(node);
	};

	list.Remove = function(handle) {
		const node = handle.get();
		node.prev.next = node.next;
		node.next.prev = node.prev;
		list.len--;
		return go2jsListHandle(node);
	};

	list.MoveToFront = function(handle) {
		const node = handle.get();
		node.prev.next = node.next;
		node.next.prev = node.prev;
		node.prev = root;
		node.next = root.next;
		root.next.prev = node;
		root.next = node;
	};

	list.MoveToBack = function(handle) {
		const node = handle.get();
		node.next.prev = node.prev;
		node.prev.next = node.next;
		node.next = root;
		node.prev = root.prev;
		root.prev.next = node;
		root.prev = node;
	};

	list.Len = function() {
		return list.len;
	};

	list.Front = function() {
		return root.next === root ? null : go2jsListHandle(root.next);
	};

	list.Back = function() {
		return root.prev === root ? null : go2jsListHandle(root.prev);
	};

	let handle = list;

	return go2jsPtr(
		() => handle,
		value => {
			handle = value;
		}
	);
}
class Address {
    City;
    Zip;
    Note;
    constructor() {
        this.City = "";
        this.Zip = 0;
        this.Note = "";
    }
}
go2jsRegisterTypeName(Address, "main.Address");

Address.prototype.Omit = function(field) {
    switch (field) {
        case "note":
            this.Note = "";
            break;
    }
}

class Person {
    Name;
    Age;
    constructor() {
        this.Name = "";
        this.Age = 0;
    }
}
go2jsRegisterTypeName(Person, "main.Person");

function Map(__go2js_type0, __go2js_type1, in$go2js, fn) {
    let out = go2jsSliceMake(0, go2jsLen(in$go2js), () => null);
    for (const v of in$go2js) {
        out = go2jsSliceAppend(out, fn(v));
    }
    return out;
}

function Filter(__go2js_type0, in$go2js, keep) {
    let out = [];
    for (const v of in$go2js) {
        if (keep(v)) {
            out = go2jsSliceAppend(out, v);
        }
    }
    return out;
}

function Reduce(__go2js_type0, __go2js_type1, in$go2js, init, fn) {
    let acc = init;
    for (const v of in$go2js) {
        acc = fn(acc, v);
    }
    return acc;
}

class Stack {
    items;
    constructor() {
        this.items = null;
    }
}
go2jsRegisterTypeName(Stack, "main.Stack");

Stack.prototype.Push = function(v) {
    this.items = go2jsSliceAppend(this.items, v);
}

Stack.prototype.Pop = function() {
    let last = go2jsLen(this.items) - 1;
    let v = this.items[last];
    this.items = go2jsSliceRange(this.items, undefined, last, undefined);
    return v;
}

Stack.prototype.Len = function() {
    return go2jsLen(this.items);
}

class Pair {
    First;
    Second;
    constructor() {
        this.First = null;
        this.Second = null;
    }
}
go2jsRegisterTypeName(Pair, "main.Pair");

function MinOf(__go2js_type0, a, b) {
    if (a < b) {
        return a;
    }
    return b;
}

function MaxOf(__go2js_type0, a, b) {
    if (a > b) {
        return a;
    }
    return b;
}

function SortInts(values) {
    go2jsSortInts(values);
}

const Sunday = 0;
const Monday = 1;
const Tuesday = 2;
function WeekdayString(d) {
    return ["Sunday", "Monday", "Tuesday"][d];
}
go2jsRegisterMethod("Weekday.String", WeekdayString);

class Shape {}

class Rect {
    W;
    H;
    tag;
    constructor() {
        this.W = 0;
        this.H = 0;
        this.tag = "";
    }
}
go2jsRegisterTypeName(Rect, "main.Rect");

Rect.prototype.Area = function() {
    return this.W * this.H;
}

Rect.prototype.Perimeter = function() {
    return 2 * (this.W + this.H);
}

class Named {
    Name;
    constructor() {
        this.Name = "";
    }
}
go2jsRegisterTypeName(Named, "main.Named");

Named.prototype.Describe = function() {
    return "named:" + this.Name;
}

class Describable {}

function fib(n) {
    if (n < 2) {
        return n;
    }
    return fib(n - 1) + fib(n - 2);
}

function gotoCounter(limit) {
    let go2jsPC = 0;
    go2js_dispatch: while (go2jsPC >= 0) {
        switch (go2jsPC) {
            case 0:
                var count = 0;
                go2jsPC = 1;
                break;
            case 1:
                loop:
                if (count < limit) {
                    count++;
                    go2jsPC = 1;
                    continue go2js_dispatch;
                }
                go2jsPC = 2;
                break;
            case 2:
                return count;
                go2jsPC = 3;
                break;
            default:
                go2jsPC = -1;
        }
    }
}

function divide(a, b) {
    if (b == 0) {
        return [0, new Error("division by zero")];
    }
    return [Math.trunc((a / b)), null];
}

function namedResult(fail) {
    const go2jsDefers = [];
    let go2jsPanicValue;
    let go2jsRecovered = false;    const go2jsRecover = () => {
        if (go2jsPanicValue === undefined || go2jsRecovered) return undefined;
        go2jsRecovered = true;
        return go2jsPanicPayload(go2jsPanicValue);
    };
    let value = 0;
    let err = null;
    const go2jsReturnSignal = {};
    let go2jsReturning = false;
    try {
        (() => go2jsDefers.push(() => function() {
            {
                let r = go2jsRecover()
                if (go2jsEqual(r, null) === false) {
                    err = go2jsWrapError("recovered: %v", r);
                }
            }
        }
()))();
        if (fail) {
            go2jsPanic("kaboom");
        }
        value = 7;
        err = null;
        go2jsReturning = true;
        throw go2jsReturnSignal;
    } catch (go2jsCaught) {
        if (go2jsCaught !== go2jsReturnSignal) {
            go2jsPanicValue = go2jsCaught;
        }
    } finally {
        for (let i = go2jsDefers.length - 1; i >= 0; i--) {
            go2jsDefers[i]();
        }
        if (go2jsPanicValue !== undefined && !go2jsRecovered) {
            throw go2jsPanicValue;
        }
    }
if (go2jsReturning || go2jsRecovered) {
    return [value, err]
}
}

function main() {
    const go2jsDefers = [];
    let go2jsPanicValue;
    let go2jsRecovered = false;    const go2jsRecover = () => {
        if (go2jsPanicValue === undefined || go2jsRecovered) return undefined;
        go2jsRecovered = true;
        return go2jsPanicPayload(go2jsPanicValue);
    };
    try {
        go2jsPrintln("=== basics ===");
        const pi = 314159/100000;
        const big = 1048576;
        let s = "go2js";
        let r = 233;
        let b = 0xFF;
        let f = 2.5;
        let i8 = -128;
        let u16 = 65535;
        let any1 = 42;
        go2jsPrintln(pi, big, s, r, b, f, i8, u16, any1);
        go2jsPrintln(WeekdayString(Sunday), WeekdayString(Monday), WeekdayString(Tuesday), WeekdayString(Tuesday), WeekdayString(Sunday));
        process.stdout.write(go2jsSprintf("%d|%s|%v|%q|%t|%c|%x|%X|%o|%b|%e|%f|%%\n", 42, "str", [1, 2], "quoted", true, 65, 255, 255, 8, 5, 1234.5, 1.5));
        process.stdout.write(go2jsSprintf("%5d|%-5d|%05d|%+d\n", 42, 42, 42, 42));
        process.stdout.write(go2jsSprintf("%8.3f|%-8.3f|%s|%q\n", 3.14159, 3.14159, "pad", "pad"));
        go2jsPrintln(1.0, 2.0, 1e3, 1.5e-3);
        go2jsPrintln();
        go2jsPrintln("=== conversions & operators ===");
        let a = 17;
        let d = 5;
        go2jsPrintln(a + d, a - d, a * d, Math.trunc((a / d)), a % d);
        go2jsPrintln(a & d, a | d, a ^ d, (a & ~d), a << 2, a >> 2, ~a);
        go2jsPrintln(a > d, a < d, a >= d, a <= d, a == d, a != d);
        go2jsPrintln(!true, -a, +a);
        let f1 = 7.5;
        let f2 = 2.0;
        go2jsPrintln(f1 + f2, f1 - f2, f1 * f2, f1 / f2);
        let n = a;
        let fl = Number(a) / Number(d);
        go2jsPrintln(n, fl, Math.trunc(fl), Math.trunc(a) * Math.trunc(b));
        let s1 = "foo";
        let s2 = "bar";
        go2jsPrintln(s1 + s2, s1 == s2, s1 < s2, go2jsLen(s1), go2jsLen(s2));
        let c1 = 97;
        let c2 = 98;
        go2jsPrintln(c1, c2, c1 < c2, String.fromCodePoint(c1) + String.fromCodePoint(c2));
        go2jsPrintln(go2jsStrconvItoa(123), go2jsStrconvQuote("hi"), go2jsStrconvFormatInt(255, 16));
        let [parsed, perr] = go2jsStrconvAtoi("456");
        go2jsPrintln(parsed, perr);
        let [bad, berr] = go2jsStrconvAtoi("abc");
        go2jsPrintln(bad, go2jsEqual(berr, null) === false);
        go2jsPrintln();
        go2jsPrintln("=== control flow ===");
        let score = 87;
        if (score >= 90) {
            go2jsPrintln("grade A");
        }
        else         if (score >= 80) {
            go2jsPrintln("grade B");
        }
        else         if (score >= 70) {
            go2jsPrintln("grade C");
        }
        else {
            go2jsPrintln("grade F");
        }
        {
            let t = score
            switch (true) {
                case t > 100:
                    go2jsPrintln("impossible");
                    break;
                case t > 50:
                    go2jsPrintln("passed");
                    break;
                default:
                    go2jsPrintln("failed");
                    break;
            }
        }
        {
            let weekday = Tuesday
            switch (weekday) {
                case Sunday:
                    go2jsPrintln("start of week");
                    break;
                case Monday:
                case Tuesday:
                    go2jsPrintln("weekday", Math.trunc(weekday));
                    break;
                default:
                    go2jsPrintln("other");
                    break;
            }
        }
        switch (go2jsTypeOf(any1)) {
            case "int":
                {
                    let v = go2jsInterfaceValue(any1);
                    go2jsPrintln("int", v + 1);
                    }
            break;
            case "string":
                {
                    let v = go2jsInterfaceValue(any1);
                    go2jsPrintln("string", v);
                    }
            break;
            case "nil":
                {
                    let v = go2jsInterfaceValue(any1);
                    go2jsPrintln("nil");
                    }
            break;
            default:
                {
                    go2jsPrintln("unknown");
                    }
            break;
        }
        {
            let s = s1
            if (go2jsLen(s) > 2) {
                go2jsPrintln("short", s.slice(0, 2), s.slice(1));
            }
        }
        let sum = 0;
        for (let i = 0; i < 5; i++) {
            sum += i;
        }
        go2jsPrintln("sum", sum);
        let i = 0;
        for (; i < 3; ) {
            i++;
        }
        go2jsPrintln("while", i);
        let j = 0;
        for (; ; ) {
            j++;
            if (j > 2) {
                break;
            }
        }
        go2jsPrintln("forever", j);
        outer:
        for (let x = 0; x < 3; x++) {
            for (let y = 0; y < 3; y++) {
                if (y == 2) {
                    continue outer;
                }
                if (x == 2) {
                    break outer;
                }
                go2jsPrint(x, y, " ");
            }
        }
        go2jsPrintln("labels");
        go2jsPrintln("goto", gotoCounter(3), gotoCounter(0));
        go2jsPrintln();
        go2jsPrintln("=== functions ===");
        go2jsPrintln(fib(10));
        let [quotient, err] = divide(10, 2);
        go2jsPrintln(quotient, err);
        [, err] = divide(1, 0);
        go2jsPrintln(err);
        let [value, verr] = namedResult(false);
        go2jsPrintln(value, verr);
        [value, verr] = namedResult(true);
        go2jsPrintln(value, go2jsEqual(verr, null) === false);
        let multi = function(x, y) {
            return x * y;
        }
;
        go2jsPrintln(multi(6, 7));
        let counter = function() {
            let n$1 = 0;
            return function() {
                n$1++;
                return n$1;
            }
;
        }
();
        go2jsPrintln(counter(), counter(), counter());
        let variadic = function(prefix, ...values) {
            let total = 0;
            for (const v of values) {
                total += v;
            }
            return prefix + ":" + go2jsStrconvItoa(total);
        }
;
        go2jsPrintln(variadic("sum", 1, 2, 3));
        go2jsPrintln(variadic("none", []));
        let apply = function(fn, v) {
            return fn(v);
        }
;
        go2jsPrintln(apply(function(x) {
            return x * x;
        }
, 9));
        let forward = multi;
        go2jsPrintln(forward(2, 3));
        (function() {
            const go2jsDefers = [];
            let go2jsPanicValue;
            let go2jsRecovered = false;            const go2jsRecover = () => {
                if (go2jsPanicValue === undefined || go2jsRecovered) return undefined;
                go2jsRecovered = true;
                return go2jsPanicPayload(go2jsPanicValue);
            };
            try {
                ((go2jsDeferArg0) => go2jsDefers.push(() => go2jsPrintln(go2jsDeferArg0)))("deferred inner");
                go2jsPrintln("inner body");
            } catch (go2jsCaught) {
                go2jsPanicValue = go2jsCaught;
            } finally {
                for (let i = go2jsDefers.length - 1; i >= 0; i--) {
                    go2jsDefers[i]();
                }
                if (go2jsPanicValue !== undefined && !go2jsRecovered) {
                    throw go2jsPanicValue;
                }
            }
        }
());
        for (let k = 0; k < 3; k++) {
            ((go2jsDeferArg0, go2jsDeferArg1) => go2jsDefers.push(() => go2jsPrintln(go2jsDeferArg0, go2jsDeferArg1)))("defer loop", k);
        }
        go2jsPrintln();
        go2jsPrintln("=== structs & pointers ===");
        let rect = Object.assign(new Rect(), {W: 3, H: 4, tag: "first"});
        go2jsPrintln(rect.Area(), rect.Perimeter(), rect.tag);
        let shape = go2jsInterface(go2jsStructCopy(rect), "Rect");
        go2jsPrintln(go2jsInterfaceCall(shape, "Area"), go2jsInterfaceCall(shape, "Perimeter"));
        let ptr = go2jsPtr(() => rect, value => rect = value);
        ptr.W = 10;
        go2jsPrintln(rect.W, ptr.W, (go2jsDeref(ptr)).H);
        let nilPtr = null;
        go2jsPrintln(nilPtr == null);
        let values = [go2jsNew(Object.assign(new Rect(), {W: 1, H: 1})), go2jsNew(Object.assign(new Rect(), {W: 2, H: 2}))];
        go2jsPrintln(values[0].Area(), values[1].Area());
        let anon = {X: 1, Y: 2, Tag: "anon"};
        go2jsPrintln(anon.X, anon.Y, anon.Tag);
        let positional = {X: 7, Y: 8};
        go2jsPrintln(positional.X, positional.Y);
        let other = Object.assign(new Rect(), {W: 3, H: 4, tag: "first"});
        go2jsPrintln(rect == other, rect == Object.assign(new Rect(), {W: 3, H: 4, tag: "first"}));
        let copied = go2jsStructCopy(rect);
        copied.W = 99;
        go2jsPrintln(rect.W, copied.W);
        let arr = go2jsArrayCopy([1, 2, 3, 4]);
        let arrPtr = go2jsPtr(() => arr, value => arr = value);
        arrPtr[0] = 100;
        go2jsPrintln(arr, go2jsLen(arr), go2jsLen(arrPtr), go2jsCap(arrPtr));
        go2jsPrintln();
        go2jsPrintln("=== interfaces & errors ===");
        let describable = go2jsInterface(Object.assign(new Named(), {Name: "go"}), "Named");
        go2jsPrintln(go2jsInterfaceCall(describable, "Describe"));
        let empty = null;
        go2jsPrintln(go2jsEqual(empty, null));
        let shape2 = null;
        let [, isNil] = go2jsAssertOK(shape2, "Shape");
        go2jsPrintln(isNil);
        let [rect2, ok] = go2jsAssertOK(shape, "Rect");
        go2jsPrintln(ok, rect2.W);
        {
                        const [go2jsM1_0, go2jsM1_1] = go2jsAssertOK(describable, "Named");
            let named = go2jsM1_0;
            let ok$2 = go2jsM1_1;

            if (ok$2) {
                go2jsPrintln("asserted", named.Name);
            }
        }
        switch (go2jsTypeOf(describable)) {
            case "Named":
                {
                    let v = go2jsInterfaceValue(describable);
                    go2jsPrintln("type switch named", v.Name);
                    }
            break;
            case "Describable":
                {
                    let v = go2jsInterfaceValue(describable);
                    go2jsPrintln("type switch describable");
                    }
            break;
        }
        let base = new Error("base failure");
        let wrapped = go2jsWrapError("context: %w", base);
        go2jsPrintln(wrapped);
        go2jsPrintln(go2jsErrorsIs(wrapped, base), go2jsErrorsIs(base, wrapped));
        let joined = go2jsErrorsJoin(new Error("first"), new Error("second"));
        go2jsPrintln(joined);
        go2jsPrintln(go2jsStringsContains(go2jsInterfaceCall(joined, "Error"), "first"), go2jsStringsContains(go2jsInterfaceCall(joined, "Error"), "second"));
        go2jsPrintln();
        go2jsPrintln("=== collections ===");
        let slice = [1, 2, 3];
        slice = go2jsSliceAppend(slice, 4, 5);
        go2jsPrintln(slice, go2jsLen(slice), go2jsSliceCap(slice) >= 5);
        let prealloc = go2jsSliceMake(3, 10, () => 0);
        prealloc[2] = 9;
        go2jsPrintln(prealloc, go2jsLen(prealloc));
        let nilSlice = null;
        go2jsPrintln(nilSlice == null, go2jsLen(nilSlice), go2jsSliceAppend(nilSlice, "x"));
        let sub = go2jsSliceRange(slice, 1, 3, undefined);
        go2jsPrintln(sub, go2jsLen(sub), go2jsSliceCap(sub) <= go2jsSliceCap(slice));
        sub[0] = 20;
        go2jsPrintln(slice[1]);
        let copiedSlice = go2jsSliceMake(go2jsLen(slice), go2jsLen(slice), () => 0);
        go2jsSliceCopy(copiedSlice, slice);
        copiedSlice[0] = 100;
        go2jsPrintln(slice[0], copiedSlice[0]);
        let twoD = [[1, 2], [3, 4]];
        go2jsPrintln(twoD, twoD[1][0], go2jsLen(twoD));
        let strSlice = ["a", "b", "c"];
        go2jsPrintln(strSlice[0] + strSlice[2], go2jsStringsJoin(strSlice, "-"));
        let strMap = go2jsMap([["one", 1], ["two", 2]]);
        go2jsMapSet(strMap, "three", 3);
        go2jsMapDelete(strMap, "one");
        go2jsPrintln(go2jsLen(strMap), go2jsMapGet(strMap, "two", 0), go2jsMapGet(strMap, "missing", 0));
        let intMap = go2jsMap([[1, "a"], [2, "b"]]);
        go2jsPrintln(go2jsMapGet(intMap, 1, ""), go2jsMapGet(intMap, 2, ""), go2jsMapGet(intMap, 3, "") == "");
        let set = go2jsMap([]);
        go2jsMapSet(set, "x", true);
        go2jsPrintln(go2jsMapGet(set, "x", false), go2jsMapGet(set, "y", false), go2jsLen(set));
        go2jsPrintln();
        go2jsPrintln("=== generics ===");
        go2jsPrintln(Map("int", "string", [1, 2, 3], function(v) {
            return go2jsStrconvItoa(v * 2);
        }
));
        go2jsPrintln(Filter("int", [1, 2, 3, 4], function(v) {
            return v % 2 == 0;
        }
));
        go2jsPrintln(Reduce("int", "int", [1, 2, 3, 4], 0, function(acc, v) {
            return acc + v;
        }
));
        let stack = Object.assign(new Stack(), {});
        stack.Push(1);
        stack.Push(2);
        go2jsPrintln(stack.Len(), stack.Pop(), stack.Len());
        let pair = Object.assign(new Pair(), {First: "age", Second: 30});
        go2jsPrintln(pair.First, pair.Second);
        go2jsPrintln(MinOf("int", 3, 7), MaxOf("int", 3, 7), MinOf("float64", 2.5, 1.5));
        let numbers = [5, 2, 8];
        SortInts(numbers);
        go2jsPrintln(numbers);
        go2jsPrintln();
        go2jsPrintln("=== strings ===");
        let text = "Hello, Go 2JS World";
        go2jsPrintln(go2jsStringsToUpper(text), go2jsStringsToLower(text));
        go2jsPrintln(go2jsStringsContains(text, "Go"), go2jsStringsHasPrefix(text, "Hello"), go2jsStringsHasSuffix(text, "World"));
        go2jsPrintln(go2jsStringsIndex(text, "Go"), go2jsStringsLastIndex(text, "o"), go2jsStringsIndex(text, "zzz"));
        go2jsPrintln(go2jsStringsSplit("a,b,c", ","), go2jsStringsSplit("abc", ""));
        go2jsPrintln(go2jsStringsJoin(["x", "y", "z"], "+"));
        go2jsPrintln(go2jsStringsReplace("aaa", "a", "b", 2), go2jsStringsReplaceAll("aaa", "a", "c"));
        go2jsPrintln(go2jsStringsRepeat("ab", 3), go2jsStringsTrimSpace("  padded  "), go2jsStringsTrim("xxhixx", "x"));
        go2jsPrintln(go2jsStringsTrimPrefix("prefix-body", "prefix-"), go2jsStringsTrimSuffix("body.suffix", ".suffix"));
        go2jsPrintln(go2jsStringsFields("  one   two  three "));
        let [before, after, found] = go2jsStringsCut("key=value", "=");
        go2jsPrintln(before, after, found);
        let upper = go2jsStringsMap(function(r) {
            if (r == 32) {
                return 45;
            }
            return r;
        }
, "a b c");
        go2jsPrintln(upper);
        go2jsPrintln(go2jsStringsEqualFold("Go", "GO"), go2jsStringsCount("cheese", "e"), go2jsStringsTitle("go2js"));
        let [cutA, cutAFound] = go2jsStringsCutPrefix("v1.2.3", "v");
        let [cutB, cutBFound] = go2jsStringsCutSuffix("file.go", ".go");
        go2jsPrintln(cutA, cutAFound, cutB, cutBFound);
        go2jsPrintln();
        go2jsPrintln("=== numbers & math ===");
        go2jsPrintln(go2jsStrconvFormatInt(255, 2), go2jsStrconvFormatFloat(1.5, 102, 3, 64));
        go2jsPrintln(...(go2jsStrconvParseInt("-42", 10, 64)));
        go2jsPrintln(...(go2jsStrconvParseFloat("3.25", 64)));
        go2jsPrintln(...(go2jsStrconvParseBool("true")));
        go2jsPrintln(go2jsStrconvQuote("with \"quotes\""), go2jsStrconvFormatBool(false));
        process.stdout.write(go2jsSprintf("%.2f %.2f %.2f %.2f\n", Math.sqrt(16), Math.abs(-3.5), Math.floor(2.7), Math.ceil(2.1)));
        process.stdout.write(go2jsSprintf("%.4f %.4f %.4f\n", Math.pow(2, 10), Math.max(3, 9), Math.min(3, 9)));
        process.stdout.write(go2jsSprintf("%.4f %.4f\n", Math.trunc(2.9), Math.round(2.5)));
        process.stdout.write(go2jsSprintf("%.4f %.4f\n", Math.log(Math.E), Math.exp(1)));
        go2jsPrintln(Number.MAX_SAFE_INTEGER > 0, Math.PI > 3.14);
        go2jsPrintln();
        go2jsPrintln("=== time ===");
        let start = go2jsStructCopy(go2jsTimeDate(2024, 3, 15, 10, 30, 0, 0, 0));
        let end = go2jsStructCopy(start.Add(go2jsDuration(90 * go2jsDuration(60000000000))));
        go2jsPrintln(start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute());
        go2jsPrintln(DurationMinutes(end.Sub(go2jsStructCopy(start))));
        go2jsPrintln(start.Add(go2jsDuration(24 * go2jsDuration(3600000000000))).Day());
        go2jsPrintln(DurationMilliseconds(Math.trunc(1500)));
        go2jsPrintln(start.Before(go2jsStructCopy(end)), end.After(go2jsStructCopy(start)), start.Equal(go2jsStructCopy(start)));
        let [parsedTime, terr] = go2jsTimeParse("2006-01-02", "2024-03-15");
        go2jsPrintln(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(), terr);
        go2jsPrintln(go2jsDuration(1000000000), go2jsDuration(go2jsDuration(1000000) * 250));
        go2jsPrintln();
        go2jsPrintln("=== containers & sorting ===");
        let list = go2jsListNew();
        list.PushBack("a");
        list.PushBack("b");
        list.PushFront("start");
        go2jsPrintln(list.Len());
        for (let e = list.Front(); e != null; e = e.Next()) {
            go2jsPrint(go2jsAssert(e.Value, "string"), " ");
        }
        go2jsPrintln();
        let front = list.Front();
        list.Remove(front);
        go2jsPrintln(list.Len(), go2jsAssert(list.Front().Value, "string"));
        let nums = [5, 2, 9, 1];
        go2jsSortInts(nums);
        go2jsPrintln(nums);
        let desc = [5, 2, 9, 1];
        go2jsSortInterface(new go2jsSortReverse(desc));
        go2jsPrintln(desc);
        let strs = ["pear", "apple", "fig"];
        go2jsSortStrings(strs);
        go2jsPrintln(strs);
        go2jsPrintln(go2jsSortSearchInts(nums, 9), go2jsSortIsSorted(go2jsInterface(nums, "IntSlice")));
        let people = [Object.assign(new Person(), {Name: "Cara", Age: 30}), Object.assign(new Person(), {Name: "Ana", Age: 25}), Object.assign(new Person(), {Name: "Bob", Age: 35})];
        go2jsSortSlice(people, function(i, j) {
            return people[i].Age < people[j].Age;
        }
);
        go2jsPrintln(people[0].Name, people[1].Name, people[2].Name);
        go2jsPrintln();
        go2jsPrintln("=== maps & slices packages ===");
        let ages = go2jsMap([["ana", 25], ["bob", 35]]);
        go2jsPrintln(go2jsSlicesContains([1, 2, 3], 2), go2jsSlicesIndex([1, 2, 3], 3));
        go2jsPrintln(go2jsSlicesMax([3, 7, 2]), go2jsSlicesMin([3, 7, 2]));
        go2jsSlicesSort(slice);
        go2jsPrintln(go2jsSlicesEqual([1, 2], [1, 2]), slice);
        let keyList = go2jsSliceMake(0, go2jsLen(ages), () => "");
        for (const k of ages.keys()) {
            keyList = go2jsSliceAppend(keyList, k);
        }
        go2jsSortStrings(keyList);
        for (const k of keyList) {
            go2jsPrint(k, "=", go2jsMapGet(ages, k, 0), " ");
        }
        go2jsPrintln();
        let [, hasAna] = go2jsMapGetOK(ages, "ana", 0);
        go2jsPrintln(hasAna, go2jsLen(ages));
        go2jsPrintln();
        go2jsPrintln("=== io, bytes & csv ===");
        let buf = new go2jsBytesBuffer();
        buf.WriteString("hello ");
        buf.WriteString("world");
        go2jsPrintln(buf.String(), buf.Len());
        go2jsPrintln(go2jsBytesContains(go2jsStringToBytes("seafood"), go2jsStringToBytes("foo")), go2jsBytesEqual(go2jsStringToBytes("ab"), go2jsStringToBytes("ab")));
        let reader = go2jsStringsNewReader("line one\nline two");
        let scanner = go2jsBufioNewScanner(go2jsInterface(reader, "*Reader"));
        for (; scanner.Scan(); ) {
            go2jsPrint(scanner.Text(), "|");
        }
        go2jsPrintln();
        let out = new go2jsBytesBuffer();
        let writer = go2jsBufioNewWriter(go2jsInterface(go2jsPtr(() => out, value => out = value), "*Buffer"));
        writer.WriteString("buffered line\n");
        writer.Flush();
        go2jsPrint(out.String());
        let records = [["name", "age"], ["ana", "25"], ["bob", "35"]];
        let csvOut = new go2jsBytesBuffer();
        let csvWriter = go2jsCSVNewWriter(go2jsInterface(go2jsPtr(() => csvOut, value => csvOut = value), "*Buffer"));
        for (const row of records) {
            csvWriter.Write(row);
        }
        csvWriter.Flush();
        go2jsPrint(go2jsStringsTrimSpace(csvOut.String()));
        let csvReader = go2jsCSVNewReader(go2jsInterface(go2jsStringsNewReader("x,1\ny,2"), "*Reader"));
        let [csvRows, rerr] = csvReader.ReadAll();
        if (go2jsEqual(rerr, null) === false) {
            go2jsPrintln("csv error", rerr);
            return;
        }
        go2jsPrintln(csvRows);
        go2jsPrintln();
        go2jsPrintln("=== regex & unicode ===");
        let re = go2jsRegexpMustCompile("[a-z]+[0-9]+");
        go2jsPrintln(re.MatchString("abc123"), re.FindString("xx abc123 yy"));
        go2jsPrintln(re.FindAllString("a1 b2 c3", -1));
        go2jsPrintln(go2jsRegexpMustCompile("^\\d+$").MatchString("12345"));
        go2jsPrintln(re.ReplaceAllString("a1 b2", "<$0>"));
        go2jsPrintln(go2jsUnicodeIsLetter(97), go2jsUnicodeIsDigit(53), go2jsUnicodeIsSpace(32));
        go2jsPrintln(go2jsUnicodeToUpper(120), go2jsUnicodeToLower(88));
        let upperRunes = go2jsStringToRunes(go2jsStringsToUpper("héllo"));
        go2jsPrintln(go2jsRunesToString(upperRunes), go2jsLen(upperRunes));
        go2jsPrintln();
        go2jsPrintln("=== json & crypto ===");
        let addr = Object.assign(new Address(), {City: "Paris", Zip: 75000, Note: ""});
        let [encoded, jerr] = go2jsJSONMarshal(addr, {"City": "city", "Zip": "zip", "Note": "note"}, ["note"]);
        go2jsPrintln(go2jsBytesToString(encoded), jerr);
        let decoded = new Address();
        let uerr = go2jsJSONUnmarshal(go2jsStringToBytes("{\"city\":\"Rome\",\"zip\":1,\"note\":\"n\"}"), go2jsPtr(() => decoded, value => decoded = value), {"City": "city", "Zip": "zip", "Note": "note"});
        go2jsPrintln(decoded.City, decoded.Zip, decoded.Note, uerr);
        decoded.Omit("note");
        let [roundTrip, ] = go2jsJSONMarshal(decoded, {"City": "city", "Zip": "zip", "Note": "note"}, ["note"]);
        go2jsPrintln(go2jsBytesToString(roundTrip));
        let md5sum = go2jsArrayCopy(go2jsMD5Sum(go2jsStringToBytes("go2js")));
        process.stdout.write(go2jsSprintf("%x\n", md5sum));
        let sha = go2jsArrayCopy(go2jsSHA256Sum256(go2jsStringToBytes("go2js")));
        process.stdout.write(go2jsSprintf("%x\n", go2jsSliceRange(sha, undefined, 8, undefined)));
        go2jsPrintln();
        go2jsPrintln("=== paths ===");
        let path = go2jsFilepathJoin("a", "b", "c.txt");
        go2jsPrintln(path, go2jsFilepathDir(path), go2jsFilepathBase(path));
        go2jsPrintln(go2jsFilepathExt(path), go2jsFilepathClean("a/./b/../c"));
        let [abs, aerr] = go2jsFilepathAbs("rel/file.txt");
        go2jsPrintln(go2jsFilepathIsAbs(abs), go2jsEqual(aerr, null));
        let sep = String.fromCodePoint(47);
        go2jsPrintln(sep, go2jsStringsCount(path, sep));
        go2jsPrintln();
        go2jsPrintln("=== concurrency ===");
        let results = go2jsChannel(3);
        let wg = go2jsWaitGroup();
        for (let i$3 = 1; i$3 <= 3; i$3++) {
            go2jsWaitGroupAdd(wg, 1);
            go2jsGo(() => function(n) {
                const go2jsDefers = [];
                let go2jsPanicValue;
                let go2jsRecovered = false;                const go2jsRecover = () => {
                    if (go2jsPanicValue === undefined || go2jsRecovered) return undefined;
                    go2jsRecovered = true;
                    return go2jsPanicPayload(go2jsPanicValue);
                };
                try {
                    (() => go2jsDefers.push(() => go2jsWaitGroupDone(wg)))();
                    go2jsChanSend(results, n * n);
                } catch (go2jsCaught) {
                    go2jsPanicValue = go2jsCaught;
                } finally {
                    for (let i = go2jsDefers.length - 1; i >= 0; i--) {
                        go2jsDefers[i]();
                    }
                    if (go2jsPanicValue !== undefined && !go2jsRecovered) {
                        throw go2jsPanicValue;
                    }
                }
            }
(i$3));
        }
        go2jsWaitGroupWait(wg);
        go2jsChannelClose(results);
        let total = 0;
        for (const $go2js_recv_1 of go2jsChannelRange(results)) {
            let v$4 = $go2js_recv_1;
            total += v$4;
        }
        go2jsPrintln("squares sum", total);
        let messages = go2jsChannel(0);
        go2jsGo(() => function() {
            go2jsChanSend(messages, "ping");
        }
());
        go2jsPrintln(go2jsChanRecv(messages));
        let done = go2jsChannel(0);
        let ticker = go2jsTimeAfter(go2jsDuration(10 * go2jsDuration(1000000)));
        go2jsChanRecv(ticker);
        go2jsGo(() => function() {
            go2jsChannelClose(done);
        }
());
        go2jsChanRecv(done);
        go2jsPrintln("done signaled");
        let mu = go2jsMutex();
        let mutexCounter = 0;
        let muWg = go2jsWaitGroup();
        for (let i$5 = 0; i$5 < 5; i$5++) {
            go2jsWaitGroupAdd(muWg, 1);
            go2jsGo(() => function() {
                const go2jsDefers = [];
                let go2jsPanicValue;
                let go2jsRecovered = false;                const go2jsRecover = () => {
                    if (go2jsPanicValue === undefined || go2jsRecovered) return undefined;
                    go2jsRecovered = true;
                    return go2jsPanicPayload(go2jsPanicValue);
                };
                try {
                    (() => go2jsDefers.push(() => go2jsWaitGroupDone(muWg)))();
                    go2jsMutexLock(mu);
                    mutexCounter++;
                    go2jsMutexUnlock(mu);
                } catch (go2jsCaught) {
                    go2jsPanicValue = go2jsCaught;
                } finally {
                    for (let i = go2jsDefers.length - 1; i >= 0; i--) {
                        go2jsDefers[i]();
                    }
                    if (go2jsPanicValue !== undefined && !go2jsRecovered) {
                        throw go2jsPanicValue;
                    }
                }
            }
());
        }
        go2jsWaitGroupWait(muWg);
        go2jsPrintln("counter", mutexCounter);
        let once = go2jsOnce();
        let onceValue = 0;
        for (let i$6 = 0; i$6 < 3; i$6++) {
            go2jsOnceDo(once, function() {
                onceValue++;
            }
);
        }
        go2jsPrintln("once", onceValue);
        let atomicCounter = 0;
        let atomWg = go2jsWaitGroup();
        for (let i$7 = 0; i$7 < 4; i$7++) {
            go2jsWaitGroupAdd(atomWg, 1);
            go2jsGo(() => function() {
                const go2jsDefers = [];
                let go2jsPanicValue;
                let go2jsRecovered = false;                const go2jsRecover = () => {
                    if (go2jsPanicValue === undefined || go2jsRecovered) return undefined;
                    go2jsRecovered = true;
                    return go2jsPanicPayload(go2jsPanicValue);
                };
                try {
                    (() => go2jsDefers.push(() => go2jsWaitGroupDone(atomWg)))();
                    go2jsAtomicAdd(go2jsAtomicCell(go2jsPtr(() => atomicCounter, value => atomicCounter = value)), 2);
                } catch (go2jsCaught) {
                    go2jsPanicValue = go2jsCaught;
                } finally {
                    for (let i = go2jsDefers.length - 1; i >= 0; i--) {
                        go2jsDefers[i]();
                    }
                    if (go2jsPanicValue !== undefined && !go2jsRecovered) {
                        throw go2jsPanicValue;
                    }
                }
            }
());
        }
        go2jsWaitGroupWait(atomWg);
        go2jsPrintln("atomic", go2jsAtomicLoad(go2jsAtomicCell(go2jsPtr(() => atomicCounter, value => atomicCounter = value))));
        let [ctx, cancel] = go2jsContextWithCancel();
        let ctxResult = go2jsChannel(1);
        cancel();
        go2jsGo(() => function() {
            go2jsChanRecv(go2jsInterfaceCall(ctx, "Done"));
            go2jsChanSend(ctxResult, go2jsInterfaceCall(go2jsInterfaceCall(ctx, "Err"), "Error"));
        }
());
        go2jsPrintln(go2jsChanRecv(ctxResult));
        go2jsPrintln(go2jsEqual(go2jsInterfaceCall(ctx, "Err"), go2jsContextCanceled()), go2jsErrorsIs(go2jsInterfaceCall(ctx, "Err"), go2jsContextCanceled()));
        go2jsPrintln();
        go2jsPrintln("=== end ===");
    } catch (go2jsCaught) {
        go2jsPanicValue = go2jsCaught;
    } finally {
        for (let i = go2jsDefers.length - 1; i >= 0; i--) {
            go2jsDefers[i]();
        }
        if (go2jsPanicValue !== undefined && !go2jsRecovered) {
            throw go2jsPanicValue;
        }
    }
}

main();
