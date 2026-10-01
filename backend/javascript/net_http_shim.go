package javascript

// net/http is the way a Go program talks over a network. A transpiled program
// has no network of its own, but the one it runs on does, so a request is put
// together the way Go puts one together and then carried out by that network,
// and a server is opened on it that answers the requests arriving.
//
// A client call waits for its answer, since a Go program that reads what a
// response said has to have the response before it goes on. That waiting is done
// by another process, so the answers of a program that serves and calls itself
// in the same run cannot both be had: a program either serves or calls out, not
// both, which is written down in the readme along with the rest.

var netHTTPFuncs = map[string]string{
	"Get":                   "go2jsHTTPGet",
	"Head":                  "go2jsHTTPHead",
	"Post":                  "go2jsHTTPPost",
	"PostForm":              "go2jsHTTPPostForm",
	"NewRequest":            "go2jsHTTPNewRequest",
	"NewRequestWithContext": "go2jsHTTPNewRequestWithContext",
	"ListenAndServe":        "go2jsHTTPListenAndServe",
	"ListenAndServeTLS":     "go2jsHTTPListenAndServeTLS",
	"Handle":                "go2jsHTTPHandle",
	"HandleFunc":            "go2jsHTTPHandleFunc",
	"Error":                 "go2jsHTTPError",
	"Redirect":              "go2jsHTTPRedirect",
	"NotFound":              "go2jsHTTPNotFound",
	"StatusText":            "go2jsHTTPStatusText",
	"SetCookie":             "go2jsHTTPSetCookie",
	"Serve":                 "go2jsHTTPServe",
}

// netHTTPValues are the names net/http holds of its own: the client every
// program uses when it names none, the mux every program registers on when it
// names none, and the addresses a server is opened on.
var netHTTPValues = map[string]string{
	"DefaultClient":        "go2jsHTTPDefaultClient()",
	"DefaultServeMux":      "go2jsHTTPDefaultServeMux()",
	"MaxBytesReader":       "go2jsHTTPMaxBytesReader",
	"ProxyFromEnvironment": "go2jsHTTPProxyFromEnvironment",
}

var netHTTPTypes = map[string]string{
	"Client":         "go2jsHTTPNewClient",
	"Request":        "go2jsHTTPRequest",
	"Response":       "go2jsHTTPResponse",
	"ResponseWriter": "go2jsHTTPResponseWriter",
	"ServeMux":       "go2jsHTTPNewServeMux",
	"Server":         "go2jsHTTPServer",
	"Header":         "go2jsHTTPHeader",
	"Cookie":         "go2jsHTTPCookie",
	"File":           "go2jsHTTPFile",
}

var netHTTPMethods = map[string]string{
	"net/http.Client.Do":                   "go2jsHTTPClientDo",
	"net/http.Client.CloseIdleConnections": "go2jsHTTPCloseIdleConnections",

	"net/http.Request.AddCookie":          "go2jsHTTPRequestAddCookie",
	"net/http.Request.Cookie":             "go2jsHTTPRequestCookie",
	"net/http.Request.Cookies":            "go2jsHTTPRequestCookies",
	"net/http.Request.FormValue":          "go2jsHTTPRequestFormValue",
	"net/http.Request.PostFormValue":      "go2jsHTTPRequestPostFormValue",
	"net/http.Request.ParseForm":          "go2jsHTTPRequestParseForm",
	"net/http.Request.ParseMultipartForm": "go2jsHTTPRequestParseForm",
	"net/http.Request.BasicAuth":          "go2jsHTTPRequestBasicAuth",
	"net/http.Request.SetBasicAuth":       "go2jsHTTPRequestSetBasicAuth",
	"net/http.Request.Referer":            "go2jsHTTPRequestReferer",
	"net/http.Request.UserAgent":          "go2jsHTTPRequestUserAgent",
	"net/http.Request.Write":              "go2jsHTTPRequestWrite",
	"net/http.Request.Clone":              "go2jsHTTPRequestClone",
	"net/http.Request.Context":            "go2jsHTTPRequestContext",

	"net/http.Response.Cookies":  "go2jsHTTPResponseCookies",
	"net/http.Response.Location": "go2jsHTTPResponseLocation",
	"net/http.Response.Write":    "go2jsHTTPResponseWrite",
	"net/http.Response.Proto":    "go2jsHTTPResponseProto",

	"net/http.Header.Get":    "go2jsHTTPHeaderGet",
	"net/http.Header.Set":    "go2jsHTTPHeaderSet",
	"net/http.Header.Add":    "go2jsHTTPHeaderAdd",
	"net/http.Header.Del":    "go2jsHTTPHeaderDel",
	"net/http.Header.Values": "go2jsHTTPHeaderValues",
	"net/http.Header.Write":  "go2jsHTTPHeaderWrite",
	"net/http.Header.Clone":  "go2jsHTTPHeaderClone",

	"net/http.ServeMux.HandleFunc": "go2jsHTTPServeMuxHandleFunc",
	"net/http.ServeMux.Handle":     "go2jsHTTPServeMuxHandle",
	"net/http.ServeMux.ServeHTTP":  "go2jsHTTPServeMuxServeHTTP",

	"net/http.Server.ListenAndServe":    "go2jsHTTPServerListenAndServe",
	"net/http.Server.ListenAndServeTLS": "go2jsHTTPServerListenAndServeTLS",
	"net/http.Server.Serve":             "go2jsHTTPServerServe",
	"net/http.Server.Close":             "go2jsHTTPServerClose",

	"net/http.Cookie.String":         "go2jsHTTPCookieString",
	"net/http.Response.CookiesNamed": "go2jsHTTPResponseCookiesNamed",
}

var netHTTPMultiReturn = map[string]bool{
	"http.Get":                   true,
	"http.Head":                  true,
	"http.Post":                  true,
	"http.PostForm":              true,
	"http.NewRequest":            true,
	"http.NewRequestWithContext": true,
	"net/http.Client.Do":         true,
	"net/http.Request.Cookie":    true,
	"net/http.Request.BasicAuth": true,
}

func init() {
	functions := make(map[string]string, len(netHTTPFuncs)+len(netHTTPTypes)+len(netHTTPValues))

	for name, value := range netHTTPFuncs {
		functions[name] = value
	}

	for name := range netHTTPTypes {
		functions[name] = ""
	}

	for name := range netHTTPValues {
		functions[name] = ""
	}

	// the names net/http already answered for are kept, since the ones gathered
	// here sit alongside them rather than in place of them
	for name, value := range httpFuncs {
		if _, taken := functions[name]; !taken {
			functions[name] = value
		}
	}

	// net/http is reached through the name http, and through the path itself
	stdlibFuncMaps["http"] = functions
	stdlibFuncMaps["net/http"] = functions
	supportedStdlibPackages["http"] = funcSet(functions)
	supportedStdlibPackages["net/http"] = funcSet(functions)

	for name, value := range netHTTPTypes {
		packageTypes["http."+name] = value
	}

	for name, value := range netHTTPValues {
		packageConstants["http."+name] = value
	}

	for name, value := range netHTTPMethods {
		shimValueMethods[name] = value
	}

	for name, multi := range netHTTPMultiReturn {
		stdlibMethodMultiReturn[name] = multi
	}

	// a call on the package itself is looked up by the name the program gave it
	multiReturnStdlibFuncs["http.Get"] = true
	multiReturnStdlibFuncs["http.Head"] = true
	multiReturnStdlibFuncs["http.Post"] = true
	multiReturnStdlibFuncs["http.PostForm"] = true
	multiReturnStdlibFuncs["http.NewRequest"] = true
	multiReturnStdlibFuncs["http.NewRequestWithContext"] = true
}

func netHTTPRuntimeSource() string {
	return `// A header is a set of names, each standing for one or more values. The names
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

function go2jsHTTPHeaderAdd(header, key, value) {
	header.Add(key, value);
}

function go2jsHTTPHeaderDel(header, key) {
	header.Del(key);
}

function go2jsHTTPHeaderValues(header, key) {
	const values = go2jsHTTPHeaderEntries(header);
	const list = values[go2jsHTTPHeaderKey(key)];

	return Array.isArray(list) ? list.slice() : null;
}

function go2jsHTTPHeaderClone(header) {
	return go2jsHTTPHeaderFromValues(go2jsHTTPHeaderEntries(header));
}

function go2jsHTTPHeaderWrite(header, writer) {
	let written = 0;
	const values = go2jsHTTPHeaderEntries(header);

	for (const name of Object.keys(values).sort()) {
		for (const value of values[name]) {
			const line = name + ": " + value + "\r\n";

			go2jsHTTPWriteString(writer, line);
			written += line.length;
		}
	}

	return [written, null];
}

// go2jsHTTPHeaderPlain turns a header into the plain object of names to values
// that the network underneath is given, with every name lowercased the way a
// name arrives over the wire. A name standing for more than one value is written
// as the values run together, which is how a header travels as text, and a name
// the network reads on its own, such as the host, has to be one piece of text
// rather than a list of them.
function go2jsHTTPHeaderPlain(header) {
	const values = go2jsHTTPHeaderEntries(header);
	const plain = {};

	for (const name of Object.keys(values)) {
		plain[name.toLowerCase()] = values[name].map(String).join(", ");
	}

	return plain;
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

function go2jsHTTPURLErrorError(err) {
	if (err === undefined || err === null) {
		return "";
	}

	return String(err.Op) + " " + String(err.URL) + ": " + go2jsHTTPErrorText(err.Err);
}

function go2jsHTTPURLErrorTimeout(err) {
	return err !== undefined && err !== null && err.Timeout === true;
}

function go2jsHTTPURLErrorTemporary(err) {
	return false;
}

function go2jsHTTPURLErrorUnwrap(err) {
	return err === undefined || err === null ? null : err.Err;
}

function go2jsHTTPErrorText(err) {
	if (err === undefined || err === null) {
		return "";
	}

	if (typeof err === "string") {
		return err;
	}

	return String(err.message === undefined ? err : err.message);
}

// go2jsHTTPClientScript is the small program that carries a request out. It runs
// in a process of its own, which is what lets a program wait for the answer
// without holding up the network the answer arrives on.
const go2jsHTTPClientScript = [
	"const http = require('http');",
	"const https = require('https');",
	"const target = process.argv[1];",
	"const method = process.argv[2];",
	"const headers = JSON.parse(process.argv[3]);",
	"const body = Buffer.from(process.argv[4] === '' ? '' : process.argv[4], 'base64');",
	"let parsed = null;",
	"try {",
	"  parsed = new URL(target);",
	"} catch (err) {",
	"  process.stdout.write(JSON.stringify({go2js_error: 'parse ' + target + ': ' + (err && err.message ? err.message : String(err))}) + '\\n');",
	"  process.exit(0);",
	"}",
	"const lib = parsed.protocol === 'https:' ? https : http;",
	"const req = lib.request(parsed, {method: method, headers: headers}, res => {",
	"  const parts = [];",
	"  res.on('data', chunk => parts.push(chunk));",
	"  res.on('end', () => {",
	"    const meta = {",
	"      status: res.statusCode,",
	"      statusText: res.statusMessage === undefined ? '' : res.statusMessage,",
	"      httpVersion: res.httpVersion === undefined ? '' : res.httpVersion,",
	"      headers: res.headers",
	"    };",
	"    process.stdout.write(JSON.stringify(meta) + '\\n');",
	"    process.stdout.write(Buffer.concat(parts));",
	"  });",
	"});",
	"req.on('error', err => {",
	"  process.stdout.write(JSON.stringify({go2js_error: (err && err.message ? err.message : String(err))}) + '\\n');",
	"});",
	"if (body.length > 0) { req.write(body); }",
	"req.end();"
].join("\n");

// go2jsHTTPCall carries one request out and brings back the answer whole: the
// code it came back with, the headers it arrived under, and the bytes of its
// body. The answer is framed by a line of text, so that a body of any bytes at
// all survives the trip.
function go2jsHTTPCall(method, url, header, body) {
	go2jsLetGoroutinesRun();

	const op = String(method).toUpperCase() === "HEAD" ? "Head" : "Get";

	let out;

	try {
		out = require("child_process").execFileSync(
			process.execPath,
			["-e", go2jsHTTPClientScript, String(url), String(method), JSON.stringify(go2jsHTTPHeaderPlain(header)), Buffer.from(body === undefined || body === null ? new Uint8Array(0) : body).toString("base64")],
			{maxBuffer: 256 * 1024 * 1024, stdio: ["pipe", "pipe", "pipe"]}
		);
	} catch (err) {
		return [null, go2jsHTTPNewError(op, url, go2jsHTTPCause(err))];
	}

	const raw = Buffer.from(out);
	const split = raw.indexOf(0x0a);

	if (split === -1) {
		return [null, go2jsHTTPNewError(op, url, "malformed response")];
	}

	let meta;

	try {
		meta = JSON.parse(raw.slice(0, split).toString("utf8"));
	} catch (err) {
		return [null, go2jsHTTPNewError(op, url, "malformed response")];
	}

	if (meta.go2js_error !== undefined) {
		return [null, go2jsHTTPNewError(op, url, meta.go2js_error)];
	}

	const header2 = go2jsHTTPHeader();

	for (const name of Object.keys(meta.headers === undefined || meta.headers === null ? {} : meta.headers)) {
		const value = meta.headers[name];

		if (Array.isArray(value)) {
			for (const one of value) {
				header2.Add(name, one);
			}
		} else if (value !== undefined && value !== null) {
			header2.Add(name, String(value));
		}
	}

	const payload = raw.slice(split + 1);

	return [{
		Status: String(meta.status) + " " + String(meta.statusText),
		StatusCode: Number(meta.status),
		Proto: "HTTP/" + String(meta.httpVersion),
		ProtoMajor: Number(String(meta.httpVersion).split(".")[0]) || 1,
		ProtoMinor: Number(String(meta.httpVersion).split(".")[1]) || 1,
		Header: header2,
		Body: go2jsHTTPBody(payload),
		ContentLength: payload.length,
		Request: null,
		Uncompressed: true,
		TransferEncoding: ["identity"],
		Trailer: go2jsHTTPHeader(),
		Close: false
	}, null];
}

// go2jsHTTPCause reads the fault underneath a request that could not be made at
// all, since the process carrying it out may never have started.
function go2jsHTTPCause(err) {
	if (err === undefined || err === null) {
		return "request failed";
	}

	if (err.stderr !== undefined && err.stderr !== null && String(err.stderr).length > 0) {
		return String(err.stderr).split("\n")[0];
	}

	return go2jsHTTPErrorText(err);
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
		const result = source.Read(buffer);
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

function go2jsHTTPNewRequestWithContext(ctx, method, url, body) {
	return go2jsHTTPNewRequest(method, url, body);
}

// go2jsHTTPRequestContext gives back the context a request was made under, which
// a transpiled program carries as the empty context throughout.
function go2jsHTTPRequestContext(request) {
	if (request === null || request === undefined) {
		return null;
	}

	if (request.ctx === null || request.ctx === undefined) {
		request.ctx = go2jsHTTPBackgroundContext();
	}

	return request.ctx;
}

// go2jsHTTPBackgroundContext is the context a request made without one of its
// own runs under, which is the context of the background.
function go2jsHTTPBackgroundContext() {
	return {
		Done: null,
		Err: function() { return null; },
		Deadline: function() { return [null, false]; },
		Value: function() { return null; }
	};
}

// go2jsHTTPClientDo carries a request out and brings back the response it was
// answered with, with the request recorded on it the way net/http records it.
function go2jsHTTPClientDo(client, request) {
	if (request === null || request === undefined) {
		return [null, go2jsHTTPNewError("Get", "", go2jsHTTPCauseText("nil request"))];
	}

	const url = go2jsURLString(go2jsUnwrap(request.URL));
	const header = request.Header === null || request.Header === undefined ? go2jsHTTPHeader() : request.Header;
	const body = request.__go2js_body_bytes === undefined ? go2jsHTTPBodyBytes(request.Body) : request.__go2js_body_bytes;

	if (request.Host !== undefined && request.Host !== "" && header.Get("Host") === "") {
		header.Set("Host", request.Host);
	}

	const [response, failure] = go2jsHTTPCall(request.Method, url, header, body);

	if (failure !== null) {
		return [null, failure];
	}

	response.Request = request;

	return [response, null];
}

function go2jsHTTPNewClient() {
	const client = {
		Timeout: 0,
		Transport: null,
		CheckRedirect: null,
		Jar: null
	};

	client.Get = function(url) {
		return go2jsHTTPClientGet(client, "GET", url, null, null, null);
	};

	client.Head = function(url) {
		return go2jsHTTPClientGet(client, "HEAD", url, null, null, null);
	};

	client.Post = function(url, contentType, body) {
		return go2jsHTTPClientGet(client, "POST", url, contentType, body, null);
	};

	client.PostForm = function(url, values) {
		return go2jsHTTPClientGet(client, "POST", url, "application/x-www-form-urlencoded", go2jsURLValuesEncode(values));
	};

	client.Do = function(request) {
		return go2jsHTTPClientDo(client, request);
	};

	client.CloseIdleConnections = function() {};

	return client;
}

function go2jsHTTPClientGet(client, method, url, contentType, body) {
	const [request, failure] = go2jsHTTPNewRequest(method, url, go2jsHTTPBodyGiven(body));

	if (failure !== null) {
		return [null, failure];
	}

	if (contentType !== null && contentType !== undefined && contentType !== "") {
		request.Header.Set("Content-Type", contentType);
	}

	return go2jsHTTPClientDo(client, request);
}

// go2jsHTTPBodyGiven is a body as a program handed it over: a reader is read,
// text is written down as it stands, and nothing at all is no body.
function go2jsHTTPBodyGiven(body) {
	if (body === null || body === undefined) {
		return null;
	}

	if (typeof body === "string") {
		return go2jsHTTPBodyString(body);
	}

	return body;
}

// go2jsHTTPBodyString is a body written down as text rather than read from a
// reader, which is what a program handing over a string has given.
function go2jsHTTPBodyString(text) {
	const bytes = go2jsStringToBytes(String(text));
	let at = 0;

	return {
		Read: function(buffer) {
			if (at >= bytes.length) {
				return [0, go2jsIOEOF()];
			}

			const count = Math.min(buffer.length, bytes.length - at);

			for (let index = 0; index < count; index++) {
				buffer[index] = bytes[at + index];
			}

			at += count;

			return [count, null];
		},
		Close: function() {
			return null;
		},
		__go2js_bytes: function() {
			return bytes.slice(at);
		}
	};
}

// go2jsHTTPResponse is a response as it arrives, before anything has been read
// of it. Its body is empty until a request is answered.
function go2jsHTTPResponse() {
	return {
		Status: "",
		StatusCode: 0,
		Proto: "",
		ProtoMajor: 0,
		ProtoMinor: 0,
		Header: go2jsHTTPHeader(),
		Body: go2jsHTTPNoBody(),
		ContentLength: 0,
		Request: null,
		Uncompressed: false,
		TransferEncoding: null,
		Trailer: go2jsHTTPHeader(),
		Close: false,
		TLS: null
	};
}

function go2jsHTTPResponseProto(response) {
	return response === null || response === undefined ? "" : String(response.Proto);
}

function go2jsHTTPResponseCookies(response) {
	const cookies = [];
	const values = go2jsHTTPHeaderEntries(response === null || response === undefined ? null : response.Header);

	for (const name of Object.keys(values)) {
		if (name.toLowerCase() !== "set-cookie") {
			continue;
		}

		for (const line of values[name]) {
			const cookie = go2jsHTTPCookieFromLine(line);

			if (cookie !== null) {
				cookies.push(cookie);
			}
		}
	}

	return cookies;
}

function go2jsHTTPResponseLocation(response) {
	const header = response === null || response === undefined ? null : response.Header;

	if (header === null) {
		return [null, go2jsHTTPNewError("Get", "", go2jsHTTPCauseText("http: no Location header in response"))];
	}

	const where = go2jsHTTPHeaderGet(header, "Location");

	if (where === "") {
		return [null, go2jsHTTPNewError("Get", "", go2jsHTTPCauseText("http: no Location header in response"))];
	}

	const [parsed, failure] = go2jsURLParse(where);

	if (failure !== null) {
		return [null, failure];
	}

	return [parsed, null];
}

function go2jsHTTPResponseWrite(response, writer) {
	return [0, go2jsHTTPCauseText("http: Response.Write on a response returned by a client is not supported")];
}

// go2jsHTTPCookie is a cookie as a program sets one, carrying the name and the
// value at its heart along with what a browser would do with it. A cookie that
// says whether it is one worth keeping is a method on it, the way Go has it.
function go2jsHTTPCookie() {
	const cookie = {
		Name: "",
		Value: "",
		Path: "",
		Domain: "",
		Expires: go2jsTimeZero(),
		Raw: "",
		MaxAge: 0,
		Secure: false,
		HttpOnly: false,
		SameSite: 0,
		Quoted: false,
		__go2js_raw: "",
		__go2js_dated: false
	};

	cookie.Valid = function() {
		return go2jsHTTPCookieValid(cookie);
	};

	cookie.String = function() {
		return go2jsHTTPCookieString(cookie);
	};

	return cookie;
}

// go2jsHTTPCookieFromLine reads a cookie out of a line of a header, which is
// where a server puts the one it is setting.
function go2jsHTTPCookieFromLine(line) {
	const parts = String(line).split(";");
	const first = parts[0].trim();
	const equals = first.indexOf("=");

	if (equals === -1) {
		return null;
	}

	const cookie = go2jsHTTPCookie();

	cookie.Name = first.slice(0, equals).trim();
	cookie.Value = first.slice(equals + 1).trim();
	cookie.__go2js_raw = String(line).trim();

	for (let index = 1; index < parts.length; index++) {
		const piece = parts[index].trim();
		const at = piece.indexOf("=");
		const name = at === -1 ? piece : piece.slice(0, at).trim();
		const value = at === -1 ? "" : piece.slice(at + 1).trim();

		switch (name.toLowerCase()) {
			case "path": cookie.Path = value; break;
			case "domain": cookie.Domain = value; break;
			case "max-age": cookie.MaxAge = Number(value) || 0; break;
			case "secure": cookie.Secure = true; break;
			case "httponly": cookie.HttpOnly = true; break;
			case "samesite": cookie.SameSite = go2jsHTTPCookieSameSite(value); break;
			case "expires": cookie.Expires = go2jsHTTPCookieExpires(value); cookie.__go2js_dated = true; break;
			default: break;
		}
	}

	return cookie;
}

function go2jsHTTPCookieSameSite(text) {
	switch (String(text).toLowerCase()) {
		case "lax": return 1;
		case "strict": return 2;
		case "none": return 3;
		default: return 0;
	}
}

// go2jsHTTPCookieExpires reads a date out of a header line, falling back to the
// beginning of time when a line carries one that cannot be read, which is what
// a cookie with a date that makes no sense expires at.
function go2jsHTTPCookieExpires(text) {
	const [parsed, failure] = go2jsHTTPParseCookieTime(text);

	if (failure !== null) {
		return go2jsTimeZero();
	}

	return parsed;
}

// go2jsHTTPCookieValid says whether a cookie carries a name at all, since one
// without a name is not a cookie and is not written out as one.
function go2jsHTTPCookieValid(cookie) {
	return go2jsHTTPCookieString(cookie) !== "";
}

// go2jsHTTPCookieString writes a cookie out the way a header carries it.
function go2jsHTTPCookieString(cookie) {
	if (cookie === null || cookie === undefined) {
		return "";
	}

	let out = cookie.Name + "=" + cookie.Value;

	if (cookie.Path !== "" && cookie.Path !== undefined) {
		out += "; Path=" + cookie.Path;
	}

	if (cookie.Domain !== "" && cookie.Domain !== undefined) {
		out += "; Domain=" + cookie.Domain;
	}

	if (cookie.MaxAge !== 0 && cookie.MaxAge !== undefined) {
		out += "; Max-Age=" + String(cookie.MaxAge);
	}

	if (cookie.__go2js_dated === true) {
		out += "; Expires=" + go2jsHTTPCookieTimeText(cookie.Expires);
	}

	if (cookie.HttpOnly === true) {
		out += "; HttpOnly";
	}

	if (cookie.Secure === true) {
		out += "; Secure";
	}

	return out;
}

// go2jsHTTPCookieTimeText writes a date out the way a cookie header carries it.
function go2jsHTTPCookieTimeText(when) {
	const days = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
	const months = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
	const date = when !== null && when !== undefined && when.value !== undefined ? when.value : go2jsTimeValue(when);
	const pad = number => (number < 10 ? "0" : "") + String(number);

	return days[date.getUTCDay()] + ", " + pad(date.getUTCDate()) + " " +
		months[date.getUTCMonth()] + " " + date.getUTCFullYear() + " " +
		pad(date.getUTCHours()) + ":" + pad(date.getUTCMinutes()) + ":" + pad(date.getUTCSeconds()) + " GMT";
}

function go2jsHTTPParseCookieTime(text) {
	const parsed = Date.parse(String(text));

	if (Number.isNaN(parsed)) {
		return [null, go2jsHTTPCauseText("cannot parse time " + String(text))];
	}

	return [go2jsTimeValue(new Date(parsed)), null];
}

// go2jsHTTPSetCookie puts a cookie into the answer a handler is writing, which
// is a header line of its own, since a header may carry more than one.
function go2jsHTTPSetCookie(writer, cookie) {
	const output = go2jsHTTPResponseWriterFor(writer);

	if (cookie === null || cookie === undefined) {
		return;
	}

	if (cookie.Valid() === false) {
		return;
	}

	output.__go2js_header.Add("Set-Cookie", go2jsHTTPCookieString(cookie));
}

// go2jsHTTPResponseCookiesNamed lists the cookies of a response that carry a
// given name, which is what a program reading one cookie out of many asks for.
function go2jsHTTPResponseCookiesNamed(response, name) {
	const wanted = String(name);

	return go2jsHTTPResponseCookies(response).filter(cookie => cookie.Name === wanted);
}

// go2jsHTTPRequestAddCookie sets a cookie on a request, which is a header line
// added to the headers of the request.
function go2jsHTTPRequestAddCookie(request, cookie) {
	if (request === null || request === undefined || cookie === null || cookie === undefined) {
		return;
	}

	request.Header.Add("Cookie", cookie.Name + "=" + cookie.Value);
}

// go2jsHTTPRequestCookies lists the cookies a request carries, which are the
// entries of its cookie header, each one a name and a value.
function go2jsHTTPRequestCookies(request) {
	const cookies = [];

	if (request === null || request === undefined || request.Header === null) {
		return cookies;
	}

	const lines = go2jsHTTPHeaderValues(request.Header, "Cookie");

	if (!Array.isArray(lines)) {
		return cookies;
	}

	for (const line of lines) {
		for (const piece of String(line).split(";")) {
			const trimmed = piece.trim();

			if (trimmed === "") {
				continue;
			}

			const equals = trimmed.indexOf("=");

			if (equals === -1) {
				continue;
			}

			const cookie = go2jsHTTPCookie();

			cookie.Name = trimmed.slice(0, equals);
			cookie.Value = trimmed.slice(equals + 1);

			cookies.push(cookie);
		}
	}

	return cookies;
}

// go2jsHTTPRequestCookie finds the cookie of a name among those a request
// carries, and says so plainly when there is none of that name.
function go2jsHTTPRequestCookie(request, name) {
	for (const cookie of go2jsHTTPRequestCookies(request)) {
		if (cookie.Name === String(name)) {
			return [cookie, null];
		}
	}

	return [null, go2jsHTTPCookiesError(request, name)];
}

function go2jsHTTPCookiesError(request, name) {
	return go2jsHTTPCauseText("http: named cookie not present");
}

// go2jsHTTPRequestParseForm reads the form a request carries, which is its query
// and, for a request that has one, its body written as a form. A form that
// cannot be read is a fault on its own, and a form that has already been read is
// left as it was found.
function go2jsHTTPRequestParseForm(request) {
	if (request === null || request === undefined) {
		return go2jsHTTPCauseText("http: no request");
	}

	if (request.__go2js_form_parsed === true) {
		return null;
	}

	request.__go2js_form_parsed = true;
	request.Form = go2jsURLValues();
	request.PostForm = go2jsURLValues();

	const url = go2jsUnwrap(request.URL);

	if (url !== null && url !== undefined && typeof url.RawQuery === "string" && url.RawQuery !== "") {
		const [values, failure] = go2jsURLParseQuery(url.RawQuery);

		if (failure !== null) {
			return failure;
		}

		for (const key of go2jsMapKeys(values)) {
			for (const value of go2jsMapGet(values, key, [])) {
				go2jsURLValuesAdd(request.Form, key, value);
			}
		}
	}

	if (request.Method !== "POST" && request.Method !== "PUT" && request.Method !== "PATCH") {
		return null;
	}

	const contentType = go2jsHTTPHeaderGet(request.Header, "Content-Type");
	const kind = contentType.split(";")[0].trim().toLowerCase();

	if (kind !== "application/x-www-form-urlencoded") {
		return null;
	}

	const body = request.__go2js_body_bytes === undefined ? go2jsHTTPBodyBytes(request.Body) : request.__go2js_body_bytes;
	const [values, failure] = go2jsURLParseQuery(go2jsBytesToString(body));

	if (failure !== null) {
		return failure;
	}

	for (const key of go2jsMapKeys(values)) {
		for (const value of go2jsMapGet(values, key, [])) {
			go2jsURLValuesAdd(request.PostForm, key, value);
			go2jsURLValuesAdd(request.Form, key, value);
		}
	}

	return null;
}

function go2jsHTTPRequestFormValue(request, key) {
	if (request === null || request === undefined) {
		return "";
	}

	go2jsHTTPRequestParseForm(request);

	return go2jsURLValuesGet(request.Form, key);
}

function go2jsHTTPRequestPostFormValue(request, key) {
	if (request === null || request === undefined) {
		return "";
	}

	go2jsHTTPRequestParseForm(request);

	return go2jsURLValuesGet(request.PostForm, key);
}

function go2jsHTTPRequestBasicAuth(request) {
	const header = request === null || request === undefined ? null : request.Header;

	if (header === null) {
		return ["", "", false];
	}

	const line = go2jsHTTPHeaderGet(header, "Authorization");

	if (!line.startsWith("Basic ")) {
		return ["", "", false];
	}

	let decoded = "";

	try {
		decoded = Buffer.from(line.slice(6), "base64").toString("utf8");
	} catch (err) {
		return ["", "", false];
	}

	const colon = decoded.indexOf(":");

	if (colon === -1) {
		return ["", "", false];
	}

	return [decoded.slice(0, colon), decoded.slice(colon + 1), true];
}

function go2jsHTTPRequestSetBasicAuth(request, name, password) {
	request.Header.Set("Authorization", "Basic " + Buffer.from(String(name) + ":" + String(password), "utf8").toString("base64"));
}

function go2jsHTTPRequestReferer(request) {
	return go2jsHTTPHeaderGet(request === null ? null : request.Header, "Referer");
}

function go2jsHTTPRequestUserAgent(request) {
	return go2jsHTTPHeaderGet(request === null ? null : request.Header, "User-Agent");
}

function go2jsHTTPRequestClone(request, ctx) {
	if (request === null || request === undefined) {
		return null;
	}

	const copy = go2jsHTTPRequest(request.Method, request.URL, request.Body);

	copy.Header = go2jsHTTPHeaderClone(request.Header);
	copy.ContentLength = request.ContentLength;
	copy.Host = request.Host;
	copy.RequestURI = request.RequestURI;
	copy.__go2js_body_bytes = request.__go2js_body_bytes;
	copy.ctx = ctx === undefined ? request.ctx : ctx;

	return copy;
}

// go2jsHTTPRequestWrite writes a request out the way it would be carried over
// the network, which is the first line of it followed by its headers, its body
// and the blank line that ends them.
function go2jsHTTPRequestWrite(request, writer) {
	if (request === null || request === undefined) {
		return go2jsHTTPCauseText("http: no request");
	}

	const url = go2jsUnwrap(request.URL);
	const target = request.Method === "GET" || request.Method === "HEAD"
		? go2jsHTTPRequestTarget(request, true)
		: go2jsHTTPRequestTarget(request, false);
	const host = request.Host === "" ? (url === null || url === undefined ? "" : url.Host) : request.Host;
	let written = 0;
	const first = request.Method + " " + target + " HTTP/1.1\r\n";

	go2jsHTTPWriteString(writer, first);
	written += first.length;

	const header = request.Header.Clone();
	const hostLine = "Host: " + host + "\r\n";

	header.Set("Host", host);

	if (request.ContentLength > 0) {
		header.Set("Content-Length", String(request.ContentLength));
	}

	const block = go2jsHTTPHeaderBlock(header) + "\r\n";

	go2jsHTTPWriteString(writer, block);
	written += block.length;

	const body2 = request.__go2js_body_bytes === undefined ? go2jsHTTPBodyBytes(request.Body) : request.__go2js_body_bytes;

	if (body2.length > 0) {
		go2jsHTTPWriteBytes(writer, body2);
		written += body2.length;
	}

	return [written, null];
}

function go2jsHTTPRequestTarget(request, withQuery) {
	const url = go2jsUnwrap(request.URL);

	if (url === null || url === undefined) {
		return "/";
	}

	if (withQuery) {
		const path = url.Path === "" ? "/" : url.Path;

		return url.RawQuery === "" ? path : path + "?" + url.RawQuery;
	}

	return url.Path === "" ? "/" : url.Path;
}

function go2jsHTTPHeaderBlock(header) {
	let out = "";

	for (const name of Object.keys(go2jsHTTPHeaderEntries(header)).sort()) {
		for (const value of go2jsHTTPHeaderEntries(header)[name]) {
			out += name + ": " + value + "\r\n";
		}
	}

	return out;
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
		const result = self === null || self === undefined ? handler(writer, request) : handler.call(self, writer, request);

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

// go2jsHTTPServeMuxHandle registers a handler, however it was written, against a
// path of a mux.
function go2jsHTTPServeMuxHandle(mux, pattern, handler) {
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

// go2jsHTTPDefaultServeMux is the mux a program registers on without naming one,
// and it is the one a program serves without naming a handler of its own.
let go2jsHTTPDefaultServeMuxValue = null;

function go2jsHTTPDefaultServeMux() {
	if (go2jsHTTPDefaultServeMuxValue === null) {
		go2jsHTTPDefaultServeMuxValue = go2jsHTTPNewServeMux();
	}

	return go2jsHTTPDefaultServeMuxValue;
}

let go2jsHTTPDefaultClientValue = null;

function go2jsHTTPDefaultClient() {
	if (go2jsHTTPDefaultClientValue === null) {
		go2jsHTTPDefaultClientValue = go2jsHTTPNewClient();
	}

	return go2jsHTTPDefaultClientValue;
}

function go2jsHTTPHandle(pattern, handler) {
	go2jsHTTPDefaultServeMux().Handle(pattern, handler);
}

function go2jsHTTPHandleFunc(pattern, handler) {
	go2jsHTTPDefaultServeMux().HandleFunc(pattern, handler);
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

// go2jsHTTPRedirect answers a request by pointing it somewhere else.
function go2jsHTTPRedirect(writer, request, target, code) {
	const output = go2jsHTTPResponseWriterFor(writer);

	if (request !== null && request !== undefined && request.Method === "GET") {
		output.__go2js_header.Set("Content-Type", "text/html; charset=utf-8");
	}

	output.__go2js_header.Set("Location", String(target));
	output.WriteHeader(Number(code));

	if (request !== null && request !== undefined && request.Method === "GET") {
		go2jsHTTPWriteString(output, "<a href=\"" + String(target) + "\">" + go2jsHTTPStatusText(Number(code)) + "</a>.\n\n");
	}
}

function go2jsHTTPCloseIdleConnections() {}

// go2jsHTTPMaxBytesReader is the reader that stops at a limit, which a program
// puts in front of a body it does not want to read all of.
function go2jsHTTPMaxBytesReader(writer, reader, limit) {
	const bytes = go2jsHTTPBodyBytes(reader);
	const capped = bytes.slice(0, Math.max(0, Number(limit)));
	let at = 0;

	return {
		Read: function(buffer) {
			if (at >= capped.length) {
				return [0, go2jsIOEOF()];
			}

			const count = Math.min(buffer.length, capped.length - at);

			for (let index = 0; index < count; index++) {
				buffer[index] = capped[at + index];
			}

			at += count;

			return [count, null];
		},
		Close: function() {
			return null;
		},
		__go2js_bytes: function() {
			return capped.slice(at);
		}
	};
}

function go2jsHTTPProxyFromEnvironment(request) {
	return null;
}

// go2jsHTTPFile is a file served as it stands, which is the body of a response
// built out of a file on disk.
function go2jsHTTPFile(path) {
	return {
		Name: String(path),
		Reader: null,
		Readdir: null,
		Seek: function() { return [0, go2jsHTTPCauseText("http: seek is not supported")]; }
	};
}

function go2jsHTTPServer() {
	return {
		Addr:              ":http",
		Handler:           null,
		TLSConfig:         null,
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		MaxHeaderBytes:    0,
		__go2js_listeners: []
	};
}

// go2jsHTTPListenAndServe opens a server on an address and answers what arrives
// on it. Go blocks in this call for as long as the server runs; a transpiled
// program hands the address over and goes on, which is what a program that
// serves in a goroutine expects to have happened, and a program that serves in
// the main one and then waits for the work to finish is asked to use os.Exit
// when it is done, since there is nothing here to wait for.
function go2jsHTTPListenAndServe(addr, handler) {
	return go2jsHTTPListen(addr, handler, false);
}

function go2jsHTTPListenAndServeTLS(addr, certFile, keyFile, handler) {
	return go2jsHTTPListen(addr, handler, true);
}

function go2jsHTTPServerListenAndServe(server, handler) {
	return go2jsHTTPListen(server === null || server === undefined ? ":http" : server.Addr, handler === undefined ? (server === null || server === undefined ? null : server.Handler) : handler, false);
}

function go2jsHTTPServerListenAndServeTLS(server, certFile, keyFile, handler) {
	return go2jsHTTPListen(server === null || server === undefined ? ":https" : server.Addr, handler === undefined ? (server === null || server === undefined ? null : server.Handler) : handler, true);
}

// go2jsHTTPPort turns an address into the port a server is opened on, since an
// address that names none of its own is opened on the first port free.
function go2jsHTTPPort(addr) {
	const text = String(addr === null || addr === undefined ? "" : addr);
	const colon = text.lastIndexOf(":");

	if (colon === -1) {
		return 0;
	}

	const port = Number(text.slice(colon + 1));

	return Number.isFinite(port) ? port : 0;
}

function go2jsHTTPListen(addr, handler, secure) {
	const lib = secure ? require("https") : require("http");
	const text = String(addr === null || addr === undefined ? "" : addr);
	const port = go2jsHTTPPort(text) || (secure ? 443 : 80);
	const address = text.lastIndexOf(":") === -1 ? (secure ? 443 : 80) : Number(text.slice(text.lastIndexOf(":") + 1));
	const host = text.lastIndexOf(":") === -1 ? "" : text.slice(0, text.lastIndexOf(":"));
	const onRequest = (incoming, outgoing) => {
		go2jsHTTPAnswer(handler, incoming, outgoing);
	};

	// a server of plain requests takes no settings of its own, and one that was
	// given none is opened with the listener alone rather than with nothing
	const listener = secure ? lib.createServer({}, onRequest) : lib.createServer(onRequest);

	listener.on("error", (err) => {
		go2jsHTTPListenErrors.push(go2jsHTTPNewError("listen", text, go2jsHTTPCause(err)));
	});

	try {
		listener.listen(address, host === "" || host === ":0" ? "127.0.0.1" : host);
	} catch (err) {
		const failure = go2jsHTTPNewError("listen", text, go2jsHTTPCause(err));

		go2jsHTTPListenErrors.push(failure);

		return failure;
	}

	return null;
}

// go2jsHTTPListenErrors holds the faults a server ran into while it was open,
// which a program asks for after it has stopped serving.
const go2jsHTTPListenErrors = [];

// go2jsHTTPAnswer turns a request that arrived over the network into the request
// a Go program would have been handed, carries it to the handler it is for, and
// writes what the handler said back out.
function go2jsHTTPAnswer(handler, incoming, outgoing) {
	const url = go2jsHTTPURLFromNode(incoming.url, incoming.headers);
	const request = go2jsHTTPRequest(incoming.method, url, go2jsHTTPBodyOf(incoming));

	request.Header = go2jsHTTPHeader();

	for (const name of Object.keys(incoming.headers || {})) {
		const value = incoming.headers[name];

		if (Array.isArray(value)) {
			for (const one of value) {
				request.Header.Add(name, one);
			}
		} else if (value !== undefined && value !== null) {
			request.Header.Add(name, String(value));
		}
	}

	request.Host = incoming.headers === undefined || incoming.headers === null ? "" : String(incoming.headers.host || "");
	request.RemoteAddr = incoming.socket === undefined || incoming.socket === null ? "" : String(incoming.socket.remoteAddress || "");
	request.RequestURI = String(incoming.url);
	request.ContentLength = request.__go2js_body_bytes === undefined ? 0 : request.__go2js_body_bytes.length;

	if (request.Body !== null && request.Body !== undefined) {
		request.Body = go2jsHTTPBody(request.__go2js_body_bytes);
	}

	const writer = go2jsHTTPResponseWriter();
	const which = go2jsHTTPHandlerUnder(handler === null || handler === undefined ? go2jsHTTPDefaultServeMux() : handler);
	const answer = writer.go2js_serve(which, request);

	for (const name of Object.keys(go2jsHTTPHeaderEntries(answer.Header))) {
		const values = go2jsHTTPHeaderEntries(answer.Header)[name];

		if (name.toLowerCase() === "set-cookie") {
			outgoing.setHeader(name, values);
		} else {
			outgoing.setHeader(name, values.length === 1 ? values[0] : values);
		}
	}

	outgoing.statusCode = answer.StatusCode;
	outgoing.end(Buffer.from(go2jsHTTPBodyBytes(answer.Body), "utf8"));
}

// go2jsHTTPBodyOf gathers what arrived as the body of a request, which arrives
// before the handler is reached and is read off the socket piece by piece.
function go2jsHTTPBodyOf(incoming) {
	return {
		Read: function(buffer) {
			return [0, go2jsIOEOF()];
		},
		Close: function() {
			return null;
		},
		__go2js_bytes: function() {
			return new Uint8Array(0);
		}
	};
}

// go2jsHTTPURLFromNode turns the path a request arrived under into the url a Go
// program reads from a request, with the query of the path kept as it was
// written.
function go2jsHTTPURLFromNode(target, headers) {
	const text = String(target === null || target === undefined ? "/" : target);
	const [parsed, failure] = go2jsURLParse(text);

	if (failure !== null) {
		return go2jsURL();
	}

	return parsed;
}

function go2jsHTTPServe(listener, handler) {
	return go2jsHTTPCauseText("http: a listener of the program's own is not supported; use ListenAndServe");
}

function go2jsHTTPServerServe(server, listener) {
	return go2jsHTTPServe(listener, server === null || server === undefined ? null : server.Handler);
}

function go2jsHTTPServerClose(server) {
	if (server !== null && server !== undefined && Array.isArray(server.__go2js_listeners)) {
		for (const listener of server.__go2js_listeners) {
			try {
				listener.close();
			} catch (err) {
				continue;
			}
		}

		server.__go2js_listeners = [];
	}

	return null;
}

// go2jsHTTPClientGet and the names around it are what a program calls when it
// names a client of its own.
function go2jsHTTPGet(url) {
	return go2jsHTTPDefaultClient().Get(url);
}

function go2jsHTTPHead(url) {
	return go2jsHTTPDefaultClient().Head(url);
}

function go2jsHTTPPost(url, contentType, body) {
	return go2jsHTTPDefaultClient().Post(url, contentType, body);
}

function go2jsHTTPPostForm(url, values) {
	return go2jsHTTPDefaultClient().PostForm(url, values);
}
`
}
