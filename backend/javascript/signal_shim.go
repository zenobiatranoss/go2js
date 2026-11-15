package javascript

import "strconv"

// os/signal delivers the signals the host process receives to a channel. Node
// reports them as events on the process, so a channel given to Notify is sent
// to, without waiting, whenever one of the signals it was asked for arrives.
// A program that waits on such a channel has to stay alive until then, so a
// registered channel keeps the event loop of the host from running dry.

var signalFuncs = map[string]string{
	"Notify":        "go2jsSignalNotify",
	"Stop":          "go2jsSignalStop",
	"Ignore":        "go2jsSignalIgnore",
	"Ignored":       "go2jsSignalIgnored",
	"Reset":         "go2jsSignalReset",
	"NotifyContext": "go2jsSignalNotifyContext",
}

// signalNames are the signals a program names, with the number and the text
// that Go gives each of them.
var signalNames = []struct {
	name   string
	number int
	text   string
}{
	{"SIGHUP", 1, "hangup"},
	{"SIGINT", 2, "interrupt"},
	{"SIGQUIT", 3, "quit"},
	{"SIGABRT", 6, "aborted"},
	{"SIGKILL", 9, "killed"},
	{"SIGUSR1", 10, "user defined signal 1"},
	{"SIGUSR2", 12, "user defined signal 2"},
	{"SIGPIPE", 13, "broken pipe"},
	{"SIGALRM", 14, "alarm clock"},
	{"SIGTERM", 15, "terminated"},
}

func init() {
	functions := make(map[string]string, len(signalFuncs))

	for name, value := range signalFuncs {
		functions[name] = value
	}

	stdlibFuncMaps["os/signal"] = functions
	supportedStdlibPackages["os/signal"] = funcSet(functions)

	for _, signal := range signalNames {
		packageVarValues["syscall."+signal.name] = "go2jsSignal_" + signal.name
		packageVarTypes["syscall."+signal.name] = "syscall.Signal"
	}

	packageVarValues["os.Interrupt"] = "go2jsSignal_SIGINT"
	packageVarValues["os.Kill"] = "go2jsSignal_SIGKILL"
	packageVarTypes["os.Interrupt"] = "syscall.Signal"
	packageVarTypes["os.Kill"] = "syscall.Signal"
}

func signalNumber(n int) string {
	return strconv.Itoa(n)
}

func signalRuntimeSource() string {
	table := ""

	for _, signal := range signalNames {
		table += "\t" + signalNumber(signal.number) + ": [\"" + signal.name + "\", \"" + signal.text + "\"],\n"
	}

	functions := ""

	for _, signal := range signalNames {
		functions += "function go2jsSignal_" + signal.name + "() {\n\treturn go2jsSignalValue(" + signalNumber(signal.number) + ");\n}\n\n"
	}

	return functions + `const go2jsSignalTable = {
` + table + `};

const go2jsSignalValues = {};
const go2jsSignalWatchers = [];

function go2jsSignalValue(number) {
	if (go2jsSignalValues[number] === undefined) {
		const known = go2jsSignalTable[number];
		const text = known === undefined ? "signal " + number : known[1];

		go2jsSignalValues[number] = {
			number: number,
			String: function() { return text; },
			Signal: function() {},
			__go2js_type_name: "syscall.Signal",
		};
	}

	return go2jsSignalValues[number];
}

function go2jsSignalNumberOf(value) {
	const inner = go2jsUnwrap(value);

	return inner !== null && typeof inner === "object" && typeof inner.number === "number" ? inner.number : Number(inner);
}

function go2jsSignalHostName(number) {
	const known = go2jsSignalTable[number];

	return known === undefined ? null : known[0];
}

let go2jsSignalResume = null;

// go2jsSignalIdle is asked when nothing can run. A program with a channel
// registered is waiting for a signal, which can only arrive once the loop has
// been given back to the host, so it is told to carry on from the signal.
function go2jsSignalIdle(resume) {
	if (resume === null) {
		go2jsSignalWatchers.length = 0;
		go2jsSignalRelease();

		return false;
	}

	if (go2jsSignalWatchers.length === 0) {
		return false;
	}

	go2jsSignalResume = resume;

	return true;
}

function go2jsSignalKeepAlive() {
	go2jsIdleHook = go2jsSignalIdle;

	if (go2jsSignalKeepAlive.timer === undefined && typeof setInterval === "function") {
		go2jsSignalKeepAlive.timer = setInterval(function() {}, 1 << 30);
	}
}

function go2jsSignalRelease() {
	if (go2jsSignalWatchers.length === 0 && go2jsSignalKeepAlive.timer !== undefined) {
		clearInterval(go2jsSignalKeepAlive.timer);
		go2jsSignalKeepAlive.timer = undefined;
	}
}

function go2jsSignalNotify(channel, ...rest) {
	if (channel === null || channel === undefined) {
		go2jsPanic("os/signal: Notify using nil channel");
	}

	if (typeof process === "undefined" || typeof process.on !== "function") {
		return;
	}

	const wanted = [];

	for (const entry of rest) {
		if (Array.isArray(entry)) {
			for (const item of entry) {
				wanted.push(go2jsSignalNumberOf(item));
			}
		} else if (entry !== null && entry !== undefined && typeof entry === "object" && !entry.number) {
			continue;
		} else {
			wanted.push(go2jsSignalNumberOf(entry));
		}
	}

	if (wanted.length === 0) {
		wanted.push(1, 2, 3, 15);
	}

	for (const number of wanted) {
		const host = go2jsSignalHostName(number);

		if (host === null || host === "SIGKILL") {
			continue;
		}

		const listener = function() {
			if (typeof channel === "function") {
				channel(go2jsSignalValue(number));
			} else {
				go2jsChanTrySend(channel, go2jsSignalValue(number));
			}

			if (go2jsSignalResume !== null) {
				const resume = go2jsSignalResume;

				go2jsSignalResume = null;
				resume();
			}
		};

		try {
			process.on(host, listener);
		} catch (e) {
			continue;
		}

		go2jsSignalWatchers.push({ channel: channel, host: host, listener: listener });
	}

	go2jsSignalKeepAlive();
}

function go2jsSignalStop(channel) {
	for (let index = go2jsSignalWatchers.length - 1; index >= 0; index--) {
		const watcher = go2jsSignalWatchers[index];

		if (watcher.channel === channel) {
			process.removeListener(watcher.host, watcher.listener);
			go2jsSignalWatchers.splice(index, 1);
		}
	}

	go2jsSignalRelease();
}

function go2jsSignalIgnore(...signals) {
	for (const signal of signals) {
		const host = go2jsSignalHostName(go2jsSignalNumberOf(signal));

		if (host !== null && host !== "SIGKILL" && typeof process !== "undefined") {
			process.on(host, function() {});
		}
	}
}

function go2jsSignalIgnored() {
	return false;
}

function go2jsSignalReset() {}

function go2jsSignalNotifyContext(parent, ...signals) {
	const made = go2jsContextWithCancel(parent);
	const target = function() {
		go2jsSignalStop(target);
		made[1]();
	};

	go2jsSignalNotify(target, ...signals);

	return [made[0], function() {
		go2jsSignalStop(target);
		made[1]();
	}];
}
`
}
