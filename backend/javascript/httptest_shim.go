package javascript

// httptest builds the pieces an HTTP handler can be exercised with: a request
// to hand it and a recorder to collect what it wrote. Neither one opens a
// socket, so both are answered here in full. The server constructors are not:
// a listener that binds a port, hands out a URL and has to be closed again is a
// different kind of thing from a recorder, and answering one with a value that
// never listens would put a claim in the program's output that Go would not
// make. A program that calls one is told it is not available rather than being
// handed something that pretends.
var httptestFuncs = map[string]string{
	"NewRecorder":           "go2jsHttptestNewRecorder",
	"NewRequest":            "go2jsHttptestNewRequest",
	"NewRequestWithContext": "go2jsHttptestNewRequest",
}

var httptestTypes = map[string]string{
	"ResponseRecorder": "go2jsHttptestResponseRecorder",
	"Response":         "go2jsHTTPResponseOf",
	"Request":          "go2jsHttptestRequest",
}

func httptestRuntimeSource() string {
	return `// go2jsHttptestRecorderBody is the buffer a recorder writes into. A Buffer
// grows as it is written to and reports what it holds as a string, which is what
// a program asks a recorded body for.
function go2jsHttptestRecorderBody() {
	const parts = [];

	return {
		Write: function(bytes) {
			parts.push(go2jsStringify(bytes));

			return Array.isArray(bytes) ? bytes.length : go2jsStringify(bytes).length;
		},
		WriteString: function(text) {
			parts.push(String(text));

			return String(text).length;
		},
		Bytes: function() {
			return go2jsStringToBytes(parts.join(""));
		},
		String: function() {
			return parts.join("");
		},
		Len: function() {
			return parts.join("").length;
		},
		Reset: function() {
			parts.length = 0;
		}
	};
}

// go2jsHttptestNewRecorder builds a recorder, which starts out answering 200
// with an empty body and no header written, the way httptest.NewRecorder does.
function go2jsHttptestNewRecorder() {
	const recorder = {
		Code: 200,
		HeaderMap: go2jsHTTPHeader(),
		Body: go2jsHttptestRecorderBody(),
		wroteHeader: false
	};

	// The header a handler writes to is the one the recorder reports, and a
	// header read before anything was written is still there to be written to.
	recorder.Header = function() {
		return recorder.HeaderMap;
	};

	recorder.WriteHeader = function(code) {
		// Only the first WriteHeader counts, the way it does in net/http: a
		// handler that calls it again has already had its answer sent.
		if (!recorder.wroteHeader) {
			recorder.wroteHeader = true;
			recorder.Code = Number(code);
		}
	};

	recorder.Write = function(bytes) {
		// Writing without a header sends the 200 that was there to begin with.
		recorder.WriteHeader(200);

		return recorder.Body.Write(bytes);
	};

	recorder.WriteString = function(text) {
		recorder.WriteHeader(200);

		return recorder.Body.WriteString(text);
	};

	recorder.Flush = function() {};

	// Result is the response the handler produced, taken at the moment it is
	// asked for, which is the moment a program stops writing to the recorder.
	recorder.Result = function() {
		return go2jsHTTPResponseOf(recorder);
	};

	return recorder;
}

function go2jsHTTPResponseOf(recorder) {
	return {
		Status: String(recorder.Code) + " " + go2jsHTTPStatusText(recorder.Code),
		StatusCode: recorder.Code,
		Header: recorder.HeaderMap,
		Body: go2jsStringToBytes(recorder.Body.String())
	};
}

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

	return text === undefined ? "Status " + code : text;
}

// go2jsHttptestNewRequest builds a request for a handler to be given. The
// context is accepted and dropped, because a request built here never outlives
// the handler it is handed to.
function go2jsHttptestNewRequest(method, target, body, context) {
	return go2jsHttptestRequest(method, target, body);
}

function go2jsHttptestRequest(method, target, body) {
	return {
		Method: String(method),
		URL: go2jsHttptestURL(target),
		Header: go2jsHTTPHeader(),
		Body: body,
		Host: go2jsHttptestURL(target).Host
	};
}

// go2jsHttptestURL breaks a target into the parts a handler reads. What a URL
// carries is spelled the way net/url spells it, so a request built here has the
// same fields a Go one does.
function go2jsHttptestURL(target) {
	const text = String(target);
	const parsed = text.includes("://") ? new URL(text) : new URL("http://" + text.replace(/^\//, ""));
	const search = text.indexOf("?");

	return {
		value: text,
		Scheme: text.includes("://") ? parsed.protocol.replace(":", "") : "",
		Host: parsed.host,
		Path: text.includes("://") ? parsed.pathname : (search === -1 ? text : text.slice(0, search)),
		RawQuery: search === -1 ? "" : text.slice(search + 1),
		Fragment: "",
		RawPath: "",
		String: function() {
			return text;
		}
	};
}

function go2jsHttptestResponseRecorder() {
	return go2jsHttptestNewRecorder();
}`
}

func init() {
	stdlibFuncMaps["net/http/httptest"] = httptestFuncs
	stdlibPkgAliases["net/http/httptest"] = []string{"httptest"}
	supportedStdlibPackages["net/http/httptest"] = funcSet(httptestFuncs)

	for name := range httptestTypes {
		packageTypes["net/http/httptest."+name] = httptestTypes[name]
	}
}
