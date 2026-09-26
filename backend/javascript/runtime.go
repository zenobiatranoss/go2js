package javascript

func runtimeSource() string {
	return `

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
            return go2jsBytesToString(value).replace(new RegExp(source, "g"), go2jsStringify(replacement));
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

function go2jsJSONMarshal(value) {
    try {
        return [JSON.stringify(value), null];
    } catch (err) {
        return [null, err];
    }
}

function go2jsJSONUnmarshal(data, target) {
    try {
        const value = JSON.parse(
            typeof data === "string"
                ? data
                : new TextDecoder().decode(Uint8Array.from(data))
        );

        if (target !== null && target !== undefined && typeof target === "object") {
            if (value !== null && typeof value === "object") {
                Object.assign(target, value);
            }
            return [null];
        }

        return [null];
    } catch (err) {
        return [err];
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

function go2jsBytesBuffer() {
    this.data = [];
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

    if (value instanceof Map) {
        return new Map(value);
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

	if (value instanceof Map) {
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

	if (typeof value === "number") {
		return "number";
	}

	if (typeof value === "string") {
		return "string";
	}

	if (typeof value === "boolean") {
		return "bool";
	}

	if (Array.isArray(value)) {
		return "slice";
	}

	if (typeof value === "object") {
		return value.constructor?.name || "object";
	}

	return typeof value;
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
	if (value instanceof Map || value instanceof Set) {
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
	if (value instanceof Map || value instanceof Set) {
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
	return new Map();
}

function go2jsMap(entries) {
	const map = new Map();

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
	if (!(map instanceof Map)) {
		return zero;
	}

	const value = map.get(key);

	// A missing key yields the element type's zero value, not null.
	return value === undefined ? zero : value;
}

function go2jsMapGetOK(map, key, zero) {
	if (!(map instanceof Map) || !map.has(key)) {
		return [zero, false];
	}

	return [map.get(key), true];
}

function go2jsMapSet(map, key, value) {
	if (!(map instanceof Map)) {
		throw new TypeError("go2jsMapSet expects a Map");
	}
	map.set(key, value);
}

function go2jsMapDelete(map, key) {
	if (!(map instanceof Map)) {
		throw new TypeError("go2jsMapDelete expects a Map");
	}
	map.delete(key);
}

function go2jsMapHas(map, key) {
	if (!(map instanceof Map)) {
		return false;
	}
	return map.has(key);
}

function go2jsMapKeys(map) {
	if (!(map instanceof Map)) {
		return [];
	}
	return Array.from(map.keys());
}

function go2jsMapValues(map) {
	if (!(map instanceof Map)) {
		return [];
	}
	return Array.from(map.values());
}

function go2jsMapEntries(map) {
	if (!(map instanceof Map)) {
		return [];
	}
	return Array.from(map.entries());
}

function go2jsCopy(value) {
	if (Array.isArray(value)) {
		return value.slice();
	}

	if (value instanceof Map) {
		return new Map(value);
	}

	if (value && typeof value === "object") {
		return Object.assign({}, value);
	}

	return value;
}

function go2jsClear(value) {
	if (value instanceof Map || value instanceof Set) {
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
	if (value instanceof Map) {
		value.delete(key);
		return;
	}

	if (value && typeof value === "object") {
		delete value[key];
	}
}

function go2jsContains(value, item) {
	if (value instanceof Map || value instanceof Set) {
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

	if (value instanceof Map || value instanceof Set) {
		return Array.from(value);
	}

	if (typeof value[Symbol.iterator] === "function") {
		return Array.from(value);
	}

	return [value];
}

function go2jsRange(value) {
	if (value instanceof Map || value instanceof Set) {
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
	if (value instanceof Map) {
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

	if (value instanceof Map) {
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

	if (value instanceof Map) {
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
			return integerBody(intValue, 16, flags.includes("#") ? "0x" : "", false);
		case "X":
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
