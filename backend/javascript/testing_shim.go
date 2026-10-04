package javascript

// testingFuncs are the functions of the testing package, which is what a test is
// run by: a program hands over what it found to be a test and the package says
// which of them passed, which failed and what each of them said.
var testingFuncs = map[string]string{
	"Short":     "go2jsTestingShort",
	"Verbose":   "go2jsTestingVerbose",
	"Init":      "go2jsTestingInit",
	"MainStart": "go2jsTestingMainStart",
	"Main":      "go2jsTestingMain",
}

// testingTypes are the types of the testing package the runtime answers with a
// value of its own, since a test is a value that carries what it has said and
// what it has failed at rather than a number or a string.
var testingTypes = map[string]string{
	"testing.T": "go2jsTestingT",
	"testing.B": "go2jsTestingB",
	"testing.M": "go2jsTestingM",

	// What a test is, a benchmark and an example are three are one value each,
	// and a value of one of them is made by naming it rather than by the runtime
	// making it, so a literal of one is built out of what it holds.
	"testing.InternalTest":      "go2jsTestingInternalTest",
	"testing.InternalBenchmark": "go2jsTestingInternalBenchmark",
	"testing.InternalExample":   "go2jsTestingInternalExample",
}

// testingMethods are the methods of the types of the testing package. They hang
// off the type they belong to, which is how a method of a type of the standard
// library is written out, and the same methods are registered for the interface
// as for the type since a test is as often handed to a helper as an interface as
// it is used as a value of the type.
var testingMethods = map[string]map[string]string{
	"testing.T": {
		"Cleanup":  "go2jsTestingCleanup",
		"Error":    "go2jsTestingError",
		"Errorf":   "go2jsTestingErrorf",
		"Fail":     "go2jsTestingFail",
		"FailNow":  "go2jsTestingFailNow",
		"Failed":   "go2jsTestingFailed",
		"Fatal":    "go2jsTestingFatal",
		"Fatalf":   "go2jsTestingFatalf",
		"Helper":   "go2jsTestingHelper",
		"Log":      "go2jsTestingLog",
		"Logf":     "go2jsTestingLogf",
		"Name":     "go2jsTestingName",
		"Parallel": "go2jsTestingParallel",
		"Run":      "go2jsTestingRun",
		"Setenv":   "go2jsTestingSetenv",
		"Chdir":    "go2jsTestingChdir",
		"TempDir":  "go2jsTestingTempDir",
		"Skip":     "go2jsTestingSkip",
		"SkipNow":  "go2jsTestingSkipNow",
		"Skipf":    "go2jsTestingSkipf",
		"Skipped":  "go2jsTestingSkipped",
		"Context":  "go2jsTestingContext",
		"Deadline": "go2jsTestingDeadline",
		"Output":   "go2jsTestingOutput",
	},
	"testing.B": {
		"Error":          "go2jsTestingBError",
		"Errorf":         "go2jsTestingBErrorf",
		"Fail":           "go2jsTestingBFail",
		"Failed":         "go2jsTestingBFailed",
		"Fatal":          "go2jsTestingBFatal",
		"Fatalf":         "go2jsTestingBFatalf",
		"Log":            "go2jsTestingBLog",
		"Logf":           "go2jsTestingBLogf",
		"Loop":           "go2jsTestingBLoop",
		"Name":           "go2jsTestingBName",
		"ResetTimer":     "go2jsTestingBResetTimer",
		"ReportAllocs":   "go2jsTestingBReportAllocs",
		"Run":            "go2jsTestingBRun",
		"SetBytes":       "go2jsTestingBSetBytes",
		"Skip":           "go2jsTestingBSkip",
		"SkipNow":        "go2jsTestingBSkipNow",
		"Skipf":          "go2jsTestingBSkipf",
		"Skipped":        "go2jsTestingBSkipped",
		"StartTimer":     "go2jsTestingBStartTimer",
		"StopTimer":      "go2jsTestingBStopTimer",
		"ReportMetric":   "go2jsTestingBReportMetric",
		"Cleanup":        "go2jsTestingBCleanup",
		"Helper":         "go2jsTestingBHelper",
		"SetParallelism": "go2jsTestingBSetParallelism",
	},
	"testing.M": {
		"Run": "go2jsTestingMRun",
	},
}

// testingTypesRegistered is every name the package answers for as a type rather
// than as a function, which is what a program reaches for when it names one.
var testingTypesRegistered = []string{
	"T", "B", "M", "TB", "F",
	"InternalTest", "InternalBenchmark", "InternalExample", "InternalFuzzTarget",
	"BenchmarkResult", "Coverage",
}

func init() {
	stdlibFuncMaps["testing"] = testingFuncs

	// The testing package is one the runtime answers itself, so what is added
	// here is folded into what it has rather than standing beside it: a package
	// is supported when everything in it is.
	supported := supportedStdlibPackages["testing"]

	if supported == nil {
		supported = map[string]bool{}
		supportedStdlibPackages["testing"] = supported
	}

	for name := range testingFuncs {
		supported[name] = true
	}

	for _, name := range testingTypesRegistered {
		supported[name] = true
	}

	for name, value := range testingTypes {
		packageTypes[name] = value
	}

	for typeName, methods := range testingMethods {
		for _, key := range []string{typeName, shortTestingKey(typeName)} {
			helpers := syncMethodHelpers[key]

			if helpers == nil {
				helpers = map[string]string{}
				syncMethodHelpers[key] = helpers
			}

			for method, helper := range methods {
				helpers[method] = helper
			}
		}

		for method := range methods {
			supported[method] = true
		}
	}
}

// shortTestingKey is the name of a type of the testing package as a method call
// finds it, which is the name of the package rather than the path it is imported
// by, since that is how a method call is looked up.
func shortTestingKey(typeName string) string {
	return "testing." + typeName[len("testing."):]
}

func testingShimRuntimeSource() string {
	return `
// go2jsTestingFailNow is what ends a test that has said enough. A test that has
// failed stops there rather than going on to say more about a state it is no
// longer in, and nothing after the call runs at all: this is thrown rather than
// returned so that the deferred calls of the frames being left run on the way
// out, which is what leaving a test early means.
const go2jsTestingFailNowSignal = {__go2js_testing_failnow: true};

// go2jsTestingState is what the run of a set of tests is holding: how it was
// asked to run, what it has run so far and what it is running now.
const go2jsTestingState = {
	verbose: false,
	short: false,
	runPattern: "",
	benchPattern: "",
	passed: 0,
	failed: 0,
	skipped: 0,
	ran: 0,
	benchTime: 1000,
	topLevel: [],
	benchmarks: [],
	started: 0,
	output: [],
	cleanups: [],
	parallel: false,
};

// go2jsTestingFlag reads what the run was asked for out of the arguments of the
// program. The names are the ones the testing package registers, and the shorter
// spellings are taken as well so that a run can be asked for the way it is
// usually asked for.
function go2jsTestingReadFlags(args) {
	const wanted = [
		["test.v", "v", "verbose"],
		["test.short", "short", "short"],
		["test.run", "run", "runPattern"],
		["test.bench", "bench", "benchPattern"],
		["test.count", "count", "count"],
		["test.timeout", "timeout", "timeout"],
		["test.benchtime", "benchtime", "benchTime"],
	];

	for (let i = 0; i < args.length; i++) {
		let name = String(args[i]).replace(/^--?/, "");
		let value = null;

		const equals = name.indexOf("=");

		if (equals >= 0) {
			value = name.slice(equals + 1);
			name = name.slice(0, equals);
		}

		for (const entry of wanted) {
			if (name !== entry[0] && name !== entry[1]) {
				continue;
			}

			if (entry[2] === "verbose") {
				go2jsTestingState.verbose = value === null || value !== "false";
			} else if (value !== null) {
				go2jsTestingState[entry[2]] = value;
			}

			break;
		}
	}
}

// go2jsTestingReadFlagsOnce reads the arguments the first time it is asked to,
// which is when a run begins rather than when a test is registered.
let go2jsTestingFlagsRead = false;

function go2jsTestingInit() {
	if (go2jsTestingFlagsRead) {
		return;
	}

	go2jsTestingFlagsRead = true;
	go2jsTestingReadFlags(typeof process !== "undefined" && process.argv !== undefined ? process.argv.slice(2) : []);
	return null;
}

// go2jsTestingShort says whether the run was asked to be short, which is what a
// test asks before spending a long time on something a short run leaves out.
function go2jsTestingShort() {
	return go2jsTestingState.short === true || go2jsTestingState.short === "true";
}

// go2jsTestingVerbose says whether the run was asked to say what every test did
// rather than only what failed.
function go2jsTestingVerbose() {
	return go2jsTestingState.verbose === true;
}

// go2jsTestingOut is a line said by the run itself, which goes out as the run
// goes out: what a test said is told apart from what the run says.
function go2jsTestingOut(text) {
	go2jsTestingState.output.push(String(text));

	if (typeof process !== "undefined" && process.stdout !== undefined) {
		process.stdout.write(String(text) + "\n");
	}
}

// go2jsTestingElapsed is how long something has been running, which is said in
// the seconds Go says it in, to the same hundredths.
function go2jsTestingElapsed(since) {
	return ((Date.now() - since) / 1000).toFixed(2) + "s";
}

// go2jsTestingMatch says whether a name is one the run was asked for. The
// pattern is one pattern per element of the name, as it is in Go, so that a test
// and the test under it can each be asked for by their own.
function go2jsTestingMatch(pattern, name) {
	if (pattern === "" || pattern === undefined || pattern === null) {
		return true;
	}

	const parts = String(pattern).split("/");
	const elements = String(name).split("/");

	// The name is read element by element against the pattern read element by
	// element, and a name with fewer elements than the pattern is held to as far
	// as it goes: a test under another is asked for by the name it has under it,
	// and the test it is under is asked for by its own, which is one element of
	// it. Where the pattern runs out, what is left of the name is not held to
	// anything.
	for (let i = 0; i < elements.length; i++) {
		if (i >= parts.length) {
			break;
		}

		if (parts[i] === "") {
			continue;
		}

		// The pattern is held to the name as it is written: it is the reader of the
		// name who says what a pattern has to be anchored to, the way it is in Go,
		// where a pattern matches the part of a name it is given and is held to the
		// whole of that part only where it says so itself.
		let matched;

		try {
			matched = new RegExp(parts[i]).test(elements[i]);
		} catch (error) {
			matched = false;
		}

		if (!matched) {
			return false;
		}
	}

	return true;
}

// go2jsTestingInternalTest is what a run of tests is handed: the name of a test
// and the function that is one.
function go2jsTestingInternalTest() {
	return {Name: "", F: null};
}

// go2jsTestingInternalBenchmark is what a run of benchmarks is handed, which is a
// benchmark and the name it is told of by.
function go2jsTestingInternalBenchmark() {
	return {Name: "", F: null};
}

// go2jsTestingInternalExample is what an example is: a function, the name it is
// told of by, and what it is held to having printed.
function go2jsTestingInternalExample() {
	return {Name: "", F: null, Output: "", Unordered: false};
}

// go2jsTestingT is a test: what it is called, what it has said, what it has
// failed at and what is left to run when it is through.
function go2jsTestingT(name, parent) {
	return {
		kind: "T",
		name: name === undefined || name === null ? "" : String(name),
		parent: parent === undefined ? null : parent,
		logs: [],
		cleanups: [],
		children: [],
		said: 0,
		report: "",
		failed: false,
		skipped: false,
		helper: false,
		parallel: false,
		started: Date.now(),
		env: [],
		dir: null,
	};
}

// go2jsTestingB is a benchmark: a test that is run often enough for how long it
// takes to mean something.
function go2jsTestingB(name, parent) {
	const benchmark = go2jsTestingT(name, parent);

	benchmark.kind = "B";
	benchmark.bytes = 0;
	benchmark.metrics = {};
	benchmark.running = true;
	benchmark.elapsed = 0;

	return benchmark;
}

// go2jsTestingM is what a whole run is held by: what was found to be a test, and
// what a run of them came to.
function go2jsTestingM() {
	return {
		kind: "M",
		tests: [],
		benchmarks: [],
		examples: [],
		fuzzTargets: [],
		code: 0,
		ran: false,
	};
}

// go2jsTestingFullName is the name of a test as it is told of, which is the name
// it was given on its own with the names of the tests it is under in front of it.
function go2jsTestingFullName(test) {
	const names = [];

	for (let at = test; at !== null && at !== undefined; at = at.parent) {
		names.unshift(at.name);
	}

	return names.join("/");
}

// go2jsTestingLog is a line of what a test has to say. What a test said is held
// until it is reported: a test that passes said nothing anyone asked to hear, and
// a test that fails says all of it, including what it said before it knew it had
// failed.
function go2jsTestingLog(test, text) {
	test.logs.push(go2jsTestingWhere() + String(text));

	if (go2jsTestingState.verbose) {
		go2jsTestingSayLogs(test);
	}
}

// go2jsTestingWhere is where in the source a test said what it said from, said
// the way Go says where a test said something from: the file it came from and the
// line it came from it, as a path inside the directory the tests were written in
// rather than the whole of it, since that is the directory the run was given.
//
// What a test said is said from the first line of the source that said it, and
// the lines under that line are said as they were written, without a place of
// their own, which is how Go says what a test said.
function go2jsTestingWhere() {
	if (typeof go2jsStackLines !== "function" || typeof go2jsSourcePlace !== "function") {
		return "";
	}

	const dir = typeof process !== "undefined" && process.env !== undefined && process.env.GO2JS_TEST_DIR !== undefined
		? String(process.env.GO2JS_TEST_DIR)
		: "";

	for (const frame of go2jsStackLines()) {
		const match = /^\s*at (?:.*? )?\(?([^()]*):(\d+):\d+\)?$/.exec(String(frame));

		if (match === null) {
			continue;
		}

		// A line of the program the emitter wrote is a line of the Go source it
		// wrote it from, and a line of the runtime is in no table of lines at all,
		// which is what tells one from the other here: a subtest is a function
		// with no name of its own, so a frame of it is found by where it was
		// written rather than by what it is called.
		const place = go2jsSourcePlace(Number(match[2]));

		if (place === null) {
			continue;
		}

		let file = String(place.file);

		if (dir !== "" && file.indexOf(dir + "/") === 0) {
			file = file.slice(dir.length + 1);
		}

		return file + ":" + place.line + ": ";
	}

	return "";
}

// go2jsTestingSayLogs says what a test has said that has not been said yet, which
// is said as a test is run when the run asked to hear what it is doing.
function go2jsTestingSayLogs(test) {
	for (const line of go2jsTestingLogs(test)) {
		go2jsTestingOut("    " + line);
	}
}

// go2jsTestingFormat is what a test was told as the format it was given says it,
// which is what a call that takes a format string means by what it was given.
function go2jsTestingFormat(format, args) {
	if (format === undefined || format === null) {
		return go2jsTestingOperands(args);
	}

	return go2jsSprintf(String(format), ...args);
}

// go2jsTestingOperands is what a call that takes no format says of what it was
// given: the operands written one after another with a space between them, which
// is how Go prints what it was handed rather than how it shapes a format.
function go2jsTestingOperands(args) {
	return Array.from(args === undefined || args === null ? [] : args)
		.map(item => go2jsOutputText([item], "", ""))
		.join(" ");
}

// go2jsTestingErrorf is what a test found wrong, said as the format it was given
// says it.
function go2jsTestingErrorf(test, format, ...args) {
	go2jsTestingFail(test);
	go2jsTestingLog(test, go2jsTestingFormat(format, args));
}

// go2jsTestingLogf is what a test has to say, said as the format it was given
// says it.
function go2jsTestingLogf(test, format, ...args) {
	go2jsTestingLog(test, go2jsTestingFormat(format, args));
}

// go2jsTestingFatalf is a test ending there over something wrong, said as the
// format it was given says it.
function go2jsTestingFatalf(test, format, ...args) {
	go2jsTestingLog(test, go2jsTestingFormat(format, args));
	go2jsTestingFailNow(test);
}

// go2jsTestingSkipf is a test leaving itself out, said as the format it was given
// says it.
function go2jsTestingSkipf(test, format, ...args) {
	go2jsTestingSkipNow(test, go2jsTestingFormat(format, args));
}

// go2jsTestingSkip is a test leaving itself out, and why.
function go2jsTestingSkip(test, ...args) {
	go2jsTestingSkipNow(test, go2jsTestingOperands(args));
}

// go2jsTestingError says a test has found something wrong, and carries on: what
// comes after it may be worth saying too.
function go2jsTestingError(test, ...args) {
	go2jsTestingFail(test);
	go2jsTestingLog(test, go2jsTestingFormat(args[0], args.slice(1)));
}

// go2jsTestingFail marks a test as failed without saying anything about it,
// which is what a test does when what it found is not what it was after.
function go2jsTestingFail(test) {
	for (let at = test; at !== null && at !== undefined; at = at.parent) {
		at.failed = true;
	}

	return null;
}

// go2jsTestingFailed says whether a test failed, which is a test failing or any of
// the tests under it failing: a test is held to what it did and what it is made of
// did, and a test is not held to what a test it is under did, since what that one
// did is that one's to answer for.
function go2jsTestingFailed(test) {
	if (test.failed) {
		return true;
	}

	for (const child of test.children) {
		if (go2jsTestingFailed(child)) {
			return true;
		}
	}

	return false;
}

// go2jsTestingFailNow marks a test as failed and ends it there, which is what a
// test does when what comes after it would only be said about a state the test
// is no longer in.
function go2jsTestingFailNow(test) {
	go2jsTestingFail(test);
	throw go2jsTestingFailNowSignal;
}

// go2jsTestingFatal says a test has found something wrong and ends it there.
function go2jsTestingFatal(test, ...args) {
	go2jsTestingLog(test, go2jsTestingFormat(args[0], args.slice(1)));
	go2jsTestingFailNow(test);
}

// go2jsTestingSkip says a test is leaving itself out, and why, and ends it there
// as though it had passed.
function go2jsTestingSkipNow(test, ...args) {
	if (args.length > 0) {
		go2jsTestingLog(test, go2jsTestingFormat(args[0], args.slice(1)));
	}

	test.skipped = true;
	throw go2jsTestingFailNowSignal;
}

// go2jsTestingSkipped says whether a test has left itself out, which is a thing a
// test does and not a thing it does to whatever it is under: a test whose test
// was left out is a test that passed.
function go2jsTestingSkipped(test) {
	return test.skipped === true;
}

// go2jsTestingCleanup is something to do once a test is through, which runs in
// the other order than it was asked for, since what was asked for last is the
// first thing left to undo.
function go2jsTestingCleanup(test, fn) {
	test.cleanups.push(fn);
}

// go2jsTestingRunCleanups does what a test left to be done, and keeps doing it
// even when one of them fails, since the rest are still to be done.
function go2jsTestingRunCleanups(test) {
	for (let i = test.cleanups.length - 1; i >= 0; i--) {
		try {
			go2jsCallNow(test.cleanups[i], null, []);
		} catch (error) {
			if (error === go2jsTestingFailNowSignal) {
				continue;
			}

			go2jsTestingLog(test, "cleanup failed: " + go2jsTestingPanicText(error));
		}
	}

	test.cleanups = [];
}

// go2jsTestingPanicText is what a thrown value says about itself, which is what
// a test is told when something in it was thrown rather than returned.
function go2jsTestingPanicText(thrown) {
	if (thrown === null || thrown === undefined) {
		return "panic: nil";
	}

	if (thrown.__go2js_panic !== undefined && thrown.__go2js_panic !== null) {
		return "panic: " + go2jsStringify(thrown.__go2js_panic);
	}

	if (typeof thrown === "string") {
		return "panic: " + thrown;
	}

	if (thrown.message !== undefined) {
		return "panic: " + String(thrown.message);
	}

	return "panic: " + go2jsStringify(thrown);
}

// go2jsTestingHelper marks a function as one that is only there to be called by
// a test, which is what keeps the name of a test to what the test itself says.
function go2jsTestingHelper() {
	return null;
}

// go2jsTestingName is the name of a test as the run knows it by.
function go2jsTestingName(test) {
	return go2jsTestingFullName(test);
}

// go2jsTestingOutput is everything a test has said, which is what a test that
// failed is shown of what it said before it did.
function go2jsTestingOutput(test) {
	return test.logs.join("\n");
}

// go2jsTestingBLogf, go2jsTestingBErrorf and go2jsTestingBFatalf are what a
// benchmark has to say, said as the format it was given says it, which is what
// the calls of a test are: a benchmark is a test that is run often enough for how
// long it takes to mean something.
function go2jsTestingBLogf(benchmark, format, ...args) {
	go2jsTestingLog(benchmark, go2jsTestingFormat(format, args));
}

function go2jsTestingBErrorf(benchmark, format, ...args) {
	go2jsTestingErrorf(benchmark, format, ...args);
}

function go2jsTestingBFatalf(benchmark, format, ...args) {
	go2jsTestingFatalf(benchmark, format, ...args);
}

// go2jsTestingBRun is a benchmark under a benchmark, which is run as a test of its
// own is: the name it is given is read as the name of the benchmark it is under.
function go2jsTestingBRun(benchmark, name, fn) {
	return go2jsTestingRun(benchmark, name, fn);
}

// go2jsTestingReportMachine says what a run of benchmarks ran on, which is what
// its numbers are numbers of: the machine and the package they were run against,
// told as they are told where benchmarks are run.
function go2jsTestingReportMachine() {
	let machine = "js";

	try {
		machine = require("os").platform() + "/" + require("os").arch();
	} catch (error) {
		machine = "js";
	}

	let cpu = "unknown";

	try {
		const list = require("os").cpus();

		if (list !== undefined && list !== null && list.length > 0) {
			cpu = String(list[0].model);
		}
	} catch (error) {
		cpu = "unknown";
	}

	const parts = machine.split("/");
	const pack = typeof process !== "undefined" && process.env !== undefined && process.env.GO2JS_TEST_PKG !== undefined
		? String(process.env.GO2JS_TEST_PKG)
		: "";

	go2jsTestingOut("goos: " + parts[0]);
	go2jsTestingOut("goarch: " + go2jsTestingGoarch(parts[1]));

	if (pack !== "") {
		go2jsTestingOut("pkg: " + pack);
	}

	go2jsTestingOut("cpu: " + cpu);
}

// go2jsTestingGoarch is the architecture a run was on, said as Go says it, which
// is not always how the engine names it: the architecture Node runs on is x64
// where Go calls the same one amd64, and an architecture neither of them names is
// said as it came rather than as nothing at all.
function go2jsTestingGoarch(arch) {
	switch (arch) {
	case "":
		return "unknown";
	case "x64":
		return "amd64";
	case "arm64":
		return "arm64";
	case "ia32":
		return "386";
	default:
		return arch;
	}
}

// go2jsTestingParallel says a test is to be run alongside the others rather than
// after them. Everything here runs in one turn of one loop, so a test that asks
// for it is run where it stands: what it asked for is which tests run beside it,
// and there is only this one loop to run them in.
function go2jsTestingParallel(test) {
	test.parallel = true;
	go2jsTestingState.parallel = true;
	return null;
}

// go2jsTestingRun runs what a test has under it, which is a test of its own and
// is named for the test it is under and the name it was given.
function go2jsTestingRun(test, name, fn) {
	const full = go2jsTestingFullName(test);
	const pattern = go2jsTestingState.runPattern;

	if (!go2jsTestingMatch(pattern, full + "/" + name)) {
		return true;
	}

	const child = test.kind === "B" ? go2jsTestingB(name, test) : go2jsTestingT(name, test);

	test.children.push(child);

	if (child.kind === "B") {
		go2jsTestingBRunOne(child, fn);

		return !go2jsTestingFailed(child);
	}

	go2jsTestingRunChild(child, fn);

	return !go2jsTestingFailed(child);
}

// go2jsTestingAnnounce says a test is being run, which is said when the run was
// asked to be told of what it is doing.
function go2jsTestingAnnounce(test) {
	test.started = Date.now();

	if (go2jsTestingState.verbose) {
		go2jsTestingOut("=== RUN   " + go2jsTestingFullName(test));
	}
}

// go2jsTestingRunOne runs one test and says how it came to be, which is what a run
// of tests is: a list of tests, each of which was run and came to something.
function go2jsTestingRunOne(test, fn) {
	go2jsTestingAnnounce(test);

	try {
		go2jsCallNow(fn, null, [test]);
	} catch (error) {
		if (error !== go2jsTestingFailNowSignal) {
			go2jsTestingFail(test);
			go2jsTestingLog(test, go2jsTestingPanicText(error));
		}
	}

	go2jsTestingRunCleanups(test);
	go2jsTestingRestore(test);
	go2jsTestingCount(test);
	go2jsTestingReport(test);

	return !go2jsTestingFailed(test);
}

// go2jsTestingRunChild runs a test that is under another one, which is reported
// by the test it is under rather than by the run, since how a test came to be is
// told under the line of the test that was run.
function go2jsTestingRunChild(test, fn) {
	go2jsTestingAnnounce(test);

	try {
		go2jsCallNow(fn, null, [test]);
	} catch (error) {
		if (error !== go2jsTestingFailNowSignal) {
			go2jsTestingFail(test);
			go2jsTestingLog(test, go2jsTestingPanicText(error));
		}
	}

	go2jsTestingRunCleanups(test);
	go2jsTestingRestore(test);
	go2jsTestingCount(test);
}

// go2jsTestingCount holds a test to account for, which is how a run of tests comes
// to a number: a test that left itself out is held to account for as one that was
// not run, and a test that failed as one that did not come to.
function go2jsTestingCount(test) {
	if (go2jsTestingSkipped(test)) {
		go2jsTestingState.skipped++;
	} else if (go2jsTestingFailed(test)) {
		go2jsTestingState.failed++;
	} else {
		go2jsTestingState.passed++;
	}
}

// go2jsTestingReport says how a test came to be, and how the tests under it came
// to be, which is said under it since that is where they were run from.
function go2jsTestingReport(test) {
	for (const line of go2jsTestingReportLines(test, 0)) {
		go2jsTestingOut(line);
	}
}

// go2jsTestingReportLines is how a test and everything under it came to be, as
// the lines it is said in: a test is said on a line of its own, then what it said
// under that, then the tests under it said under that, each a level further in,
// which is how a test is told of the tests it is made of.
function go2jsTestingReportLines(test, depth) {
	const pad = depth === 0 ? "" : "    ".repeat(depth);
	const took = go2jsTestingElapsed(test.started);
	const name = go2jsTestingFullName(test);
	const lines = [];

	// A test that failed is said whether or not the run asked to be told of what
	// it is doing, since a run that failed is one thing to hear about. A test that
	// came to is said only when it was asked, since a run that passed is one thing
	// to hear about however many tests it ran.
	if (go2jsTestingFailed(test)) {
		lines.push(pad + "--- FAIL: " + name + " (" + took + ")");
	} else if (go2jsTestingState.verbose) {
		lines.push(pad + "--- " + (go2jsTestingSkipped(test) ? "SKIP: " : "PASS: ") + name + " (" + took + ")");
	}

	if (go2jsTestingFailed(test)) {
		for (const line of go2jsTestingLogs(test)) {
			lines.push(pad + "    " + line);
		}
	}

	for (const child of test.children) {
		for (const line of go2jsTestingReportLines(child, depth + 1)) {
			lines.push(line);
		}
	}

	return lines;
}

// go2jsTestingLogs is what a test said that has not been said yet, said a line at
// a time: what a test said is held until it is reported, since a test that passed
// said nothing anyone asked to hear, while a test that failed says all of it.
function go2jsTestingLogs(test) {
	const lines = [];

	while (test.said < test.logs.length) {
		const line = test.logs[test.said];

		test.said++;

		for (const part of line.split("\n")) {
			lines.push(part);
		}
	}

	return lines;
}

// go2jsTestingRestore puts back what a test changed about the run, since a test
// is a guest of it: a value of the environment and the directory it stands in.
function go2jsTestingRestore(test) {
	for (let i = test.env.length - 1; i >= 0; i--) {
		const entry = test.env[i];

		if (entry.was === null) {
			delete process.env[entry.name];
		} else {
			process.env[entry.name] = entry.was;
		}
	}

	test.env = [];

	if (test.dir !== null) {
		process.chdir(test.dir);
		test.dir = null;
	}
}

// go2jsTestingSetenv sets a value of the environment for as long as the test is
// running, and puts back what was there before once the test is through.
function go2jsTestingSetenv(test, name, value) {
	const key = String(name);
	const was = Object.prototype.hasOwnProperty.call(process.env, key) ? process.env[key] : null;

	test.env.push({name: key, was: was});
	process.env[key] = value === undefined || value === null ? "" : String(value);

	return null;
}

// go2jsTestingChdir stands the run in a directory for as long as the test is
// running, and stands it back where it was once the test is through.
function go2jsTestingChdir(test, dir) {
	if (test.dir === null) {
		test.dir = process.cwd();
	}

	process.chdir(String(dir));

	return null;
}

// go2jsTestingTempDir is a directory of the test's own to write in, which is
// taken away with the test rather than left behind.
function go2jsTestingTempDir(test, pattern) {
	const wanted = pattern === undefined || pattern === null || pattern === ""
		? String(test.name).replace(/[^A-Za-z0-9]+/g, "") + "-"
		: String(pattern);

	const base = go2jsOSTempDir();
	let dir = "";

	for (let attempt = 0; attempt < 100; attempt++) {
		const suffix = Math.random().toString(36).slice(2, 8);
		dir = base.replace(/\/+$/, "") + "/" + go2jsTestingExpandTempPattern(wanted, suffix);

		try {
			require("fs").mkdirSync(dir, {recursive: true});

			const cleanup = test.cleanups[test.cleanups.length - 1];

			go2jsTestingCleanup(test, () => {
				require("fs").rmSync(dir, {recursive: true, force: true});
			});

			if (cleanup !== undefined) {
				test.cleanups[test.cleanups.length - 1] = test.cleanups[test.cleanups.length - 1];
			}

			return dir;
		} catch (error) {
			continue;
		}
	}

	throw new Error("testing: cannot make a temporary directory");
}

// go2jsTestingExpandTempPattern is what the pattern of a temporary directory
// asks for: a star stands for what makes it different from another one.
function go2jsTestingExpandTempPattern(pattern, suffix) {
	let out = "";
	let stars = 0;

	for (const ch of String(pattern)) {
		if (ch === "*") {
			stars++;
			continue;
		}

		if (stars > 0) {
			out += suffix;
			stars--;
		}

		out += ch;
	}

	return out + (stars > 0 ? suffix : "");
}

// go2jsTestingContext is the context a test is cancelled with, which is the one
// of the run itself until the run is told otherwise.
function go2jsTestingContext() {
	return go2jsContextBackground();
}

// go2jsTestingDeadline is when a test is out of time, which here it never is: a
// test runs in one turn of one loop and is ended by what it does rather than by
// how long it has been running.
function go2jsTestingDeadline() {
	return [go2jsTimeZero(), false];
}

// go2jsTestingMainStart is what a run of tests is started with: everything that
// was found to be one.
function go2jsTestingMainStart(deps, tests, benchmarks, fuzzTargets, examples) {
	go2jsTestingInit();

	const m = go2jsTestingM();

	m.tests = tests === undefined || tests === null ? [] : Array.from(tests);
	m.benchmarks = benchmarks === undefined || benchmarks === null ? [] : Array.from(benchmarks);
	m.fuzzTargets = fuzzTargets === undefined || fuzzTargets === null ? [] : Array.from(fuzzTargets);
	m.examples = examples === undefined || examples === null ? [] : Array.from(examples);

	return m;
}

// go2jsTestingMain runs everything that was handed over and ends the program
// with what the run came to, which is what a program that ran tests ends with.
function go2jsTestingMain(matchString, tests, benchmarks, examples) {
	const m = go2jsTestingMainStart(null, tests, benchmarks, [], examples);

	go2jsOSExit(go2jsTestingMRun(m));
}

// go2jsTestingMRun runs every test that was handed over and answers with what
// the run came to: nothing but failures is a run that passed.
function go2jsTestingMRun(m) {
	if (m.ran) {
		return m.code;
	}

	m.ran = true;
	go2jsTestingInit();
	go2jsTestingState.started = Date.now();

	const times = go2jsTestingTimes();

	for (let round = 0; round < times; round++) {
		go2jsTestingRunOnce(m);
	}

	m.code = go2jsTestingState.failed > 0 ? 1 : 0;

	if (m.code === 0 && go2jsTestingState.ran === 0) {
		go2jsTestingOut("testing: warning: no tests to run\n");
	}

	go2jsTestingOut(m.code === 0 ? "PASS" : "FAIL");

	return m.code;
}

// go2jsTestingTimes is how many times a run was asked to run, which is once unless
// it was asked for more.
function go2jsTestingTimes() {
	const asked = parseInt(go2jsTestingState.count, 10);

	return asked > 0 ? asked : 1;
}

// go2jsTestingRunOnce is one run of what was found: the tests, and the benchmarks
// if a run of benchmarks was asked for, and the examples.
function go2jsTestingRunOnce(m) {

	for (const entry of m.tests) {
		if (!go2jsTestingMatch(go2jsTestingState.runPattern, String(entry.Name))) {
			continue;
		}

		go2jsTestingRan();
		go2jsTestingRunOne(go2jsTestingT(String(entry.Name), null), entry.F);
	}

	if (go2jsTestingState.benchPattern !== "") {
		go2jsTestingReportMachine();

		for (const entry of m.benchmarks) {
			if (!go2jsTestingMatch(go2jsTestingState.benchPattern, String(entry.Name))) {
				continue;
			}

			go2jsTestingRan();
			go2jsTestingBRunOne(go2jsTestingB(String(entry.Name), null), entry.F);
		}
	}

	for (const entry of m.examples) {
		if (!go2jsTestingMatch(go2jsTestingState.runPattern, String(entry.Name))) {
			continue;
		}

		go2jsTestingRan();
		go2jsTestingRunExample(entry);
	}
}

// go2jsTestingRan holds that something was run, which is what a run of tests that
// ran nothing says: a warning, since a run of tests that ran no tests has told
// whoever ran it nothing about the tests they thought they were running.
function go2jsTestingRan() {
	go2jsTestingState.ran++;
}


// go2jsTestingRunExample runs an example, which is a function whose output is
// what is compared, so what it printed is what it is held to.
function go2jsTestingRunExample(entry) {
	const name = String(entry.Name);
	const printed = [];

	go2jsTestingOut("=== RUN   " + name);

	const started = Date.now();
	const restore = go2jsTestingCaptureOutput(printed);

	try {
		go2jsCallNow(entry.F, null, []);
	} catch (error) {
		restore();
		throw error;
	}

	restore();

	const wanted = entry.Output === undefined || entry.Output === null
		? ""
		: go2jsOutputText([entry.Output], "", "\n");

	const got = printed.join("");

	if (got === wanted) {
		go2jsTestingState.passed++;
		go2jsTestingOut("--- PASS: " + name + " (" + go2jsTestingElapsed(started) + ")");
		return;
	}

	// An example that printed what it was held to passed, and one that did not is
	// told what it printed and what it was held to, which is the whole of what is
	// wrong with it.
	go2jsTestingState.failed++;
	go2jsTestingOut("--- FAIL: " + name + " (" + go2jsTestingElapsed(started) + ")");
	go2jsTestingOut("got:");
	go2jsTestingOut(go2jsTestingTrimEnd(got));
	go2jsTestingOut("want:");
	go2jsTestingOut(go2jsTestingTrimEnd(wanted));
}

// go2jsTestingTrimEnd is what was printed without the newlines that ended it,
// which is how what an example printed is told: line by line, as it was printed,
// rather than as one string of it.
function go2jsTestingTrimEnd(text) {
	return String(text).replace(/[\r\n]+$/, "");
}

// go2jsTestingCaptureOutput collects what is written while a function runs, which
// is what an example is compared by.
function go2jsTestingCaptureOutput(into) {
	if (typeof process === "undefined" || process.stdout === undefined) {
		return () => {};
	}

	const write = process.stdout.write.bind(process.stdout);

	process.stdout.write = (chunk, ...rest) => {
		into.push(typeof chunk === "string" ? chunk : Buffer.from(chunk).toString("utf8"));
		return true;
	};

	return () => {
		process.stdout.write = write;
	};
}

// go2jsTestingBError says a benchmark found something wrong, which does not stop
// it: a benchmark that keeps running is a benchmark whose numbers mean little.
function go2jsTestingBError(b, ...args) {
	go2jsTestingBFail(b);
	go2jsTestingBLog(b, go2jsTestingFormat(args[0], args.slice(1)));
}

function go2jsTestingBFatal(b, ...args) {
	go2jsTestingBLog(b, go2jsTestingFormat(args[0], args.slice(1)));
	go2jsTestingFailNow(b);
}

function go2jsTestingBLog(b, ...args) {
	return go2jsTestingLog(b, go2jsTestingFormat(args[0], args.slice(1)));
}

function go2jsTestingBFail(b) {
	return go2jsTestingFail(b);
}

function go2jsTestingBFailed(b) {
	return go2jsTestingFailed(b);
}

function go2jsTestingBName(b) {
	return go2jsTestingFullName(b);
}

function go2jsTestingBSkipped(b) {
	return go2jsTestingSkipped(b);
}

function go2jsTestingBSkipNow(b, ...args) {
	return go2jsTestingSkipNow(b, ...args);
}

function go2jsTestingBSkip(b, ...args) {
	return go2jsTestingSkipNow(b, ...args);
}

function go2jsTestingBSkipf(b, format, ...args) {
	return go2jsTestingSkipNow(b, format, ...args);
}

function go2jsTestingBCleanup(b, fn) {
	return go2jsTestingCleanup(b, fn);
}

function go2jsTestingBHelper() {
	return null;
}

function go2jsTestingBReportAllocs(b) {
	b.allocations = true;
	return null;
}

function go2jsTestingBSetParallelism() {
	return null;
}

function go2jsTestingBSetBytes(b, n) {
	b.bytes = Number(n);
	return null;
}

function go2jsTestingBReportMetric(b, n, unit) {
	b.metrics[String(unit)] = Number(n);
	return null;
}

// go2jsTestingBNow is how long a benchmark has been running for, in nanoseconds,
// which is what the clock of the host is read for.
function go2jsTestingBNow() {
	// A benchmark is timed in the clock that counts most finely there is, since
	// what it is measuring is often over in less time than a clock that counts in
	// milliseconds can tell apart, and what it is told is turned back into
	// milliseconds where it is said.
	return Number(process.hrtime.bigint()) / 1e6;
}

function go2jsTestingBStartTimer(b) {
	if (b.running) {
		return null;
	}

	b.elapsed += go2jsTestingBNow() - b.started;
	b.started = go2jsTestingBNow();
	b.running = true;

	return null;
}

function go2jsTestingBStopTimer(b) {
	if (b.running) {
		b.elapsed += go2jsTestingBNow() - b.started;
		b.running = false;
	}

	return null;
}

function go2jsTestingBResetTimer(b) {
	b.started = go2jsTestingBNow();
	b.running = true;
	b.elapsed = 0;
	return null;
}

// go2jsTestingBRunName runs what a benchmark has under it.
function go2jsTestingBRunName(b, fn) {
	go2jsTestingBRunOne(b, fn);
	return !go2jsTestingFailed(b);
}

// go2jsTestingBRunOne runs a benchmark as many times over as it takes to be worth
// having run, and says what it cost, which is the whole of what a benchmark is
// for: a number of what one of them costs, and a number of how many of them were
// run to make it.
function go2jsTestingBRunOne(b, fn) {
	const target = go2jsTestingBenchTime();

	go2jsTestingState.verbose && go2jsTestingOut("=== RUN   " + go2jsTestingFullName(b));

	// A benchmark is run once, then twice as many times as it was, and so on
	// until a run took as long as the run was asked to take: the number of times a
	// benchmark is run is decided by how long it took to run it once, which is
	// what Go does to find out what a thing costs. The last run is the one that is
	// reported, since it is the one that was run for as long as it was asked for.
	let n = 1;
	let elapsed = 0;

	go2jsTestingState.benchmarks++;

	try {
		for (;;) {
			go2jsTestingBResetTimer(b);
			go2jsTestingRunBenchmarkBody(b, fn, n);

			elapsed = go2jsTestingBElapsed(b);

			if (elapsed >= target) {
				break;
			}

			n *= 2;

			if (n >= go2jsTestingBenchCeiling) {
				break;
			}
		}
	} catch (error) {
		if (error !== go2jsTestingFailNowSignal) {
			go2jsTestingBFail(b);
			go2jsTestingBLog(b, go2jsTestingPanicText(error));
		}
	}

	go2jsTestingBStopTimer(b);
	go2jsTestingRunCleanups(b);
	go2jsTestingReportBenchmark(b);

	return !go2jsTestingFailed(b);
}

// go2jsTestingRunBenchmarkBody runs what a benchmark is once for as many times as
// it was asked to be run, which is either a loop over b.N of whatever it does, or
// a loop over b.Loop, which is what a benchmark written the newer way asks for.
function go2jsTestingRunBenchmarkBody(b, fn, n) {
	// A benchmark is run once for as many times as it is asked to be run, since a
	// benchmark written the older way loops over b.N of what it does and one
	// written the newer way loops until b.Loop says it has been run enough.
	// Either way the loop is the benchmark's own, so what it did is counted as it
	// does it rather than by how many times it was handed the whole of it.
	b.N = n;
	b.loop = {count: 0, started: 0};
	b.iterations = n;

	go2jsCallNow(fn, null, [b]);

	if (b.loop.count > 0) {
		b.iterations = b.loop.count;
	}
}

// go2jsTestingBenchTime is how long a benchmark is run for, which is what it was
// asked for, said as a duration in the writing Go writes one in: a number with a
// unit under it, which is turned here into the milliseconds a clock of this kind
// counts in. A run that asked for none is run for as long as a benchmark is run
// for when nothing was asked.
function go2jsTestingBenchTime() {
	const asked = go2jsTestingState.benchTime;

	if (asked === undefined || asked === null || asked === "") {
		return 1000;
	}

	return go2jsTestingDurationMs(String(asked), 1000);
}

// go2jsTestingDurationMs is a duration written the way Go writes one read as the
// milliseconds it stands for, since a clock of this kind counts in them. A
// duration written in a way that cannot be read is read as what a run that asked
// for none is run for, rather than as no time at all.
function go2jsTestingDurationMs(text, fallback) {
	const match = /^([0-9]+(?:\.[0-9]+)?)(ns|us|µs|μs|ms|s|m|h)?$/.exec(text.trim());

	if (match === null) {
		return fallback;
	}

	const amount = Number(match[1]);

	switch (match[2] === undefined ? "s" : match[2]) {
	case "ns":
		return amount / 1e6;
	case "us":
	case "µs":
	case "μs":
		return amount / 1e3;
	case "ms":
		return amount;
	case "s":
		return amount * 1e3;
	case "m":
		return amount * 6e4;
	default:
		return amount * 36e5;
	}
}

// go2jsTestingBenchCeiling is how many times a benchmark is run at the most: the
// number of times a benchmark is decided by how long it took, and a benchmark that
// costs next to nothing would run for ever deciding that. A billion times is where
// Go stops counting, which is as far as a program of this kind is counted.
const go2jsTestingBenchCeiling = 1e9;

// go2jsTestingBElapsed is how long the benchmark has been running, which is what
// its numbers are numbers of: the time it has spent running rather than the time
// that has passed since it started, since a benchmark may stop and start its own
// timer as it goes.
function go2jsTestingBElapsed(b) {
	const since = b.running ? go2jsTestingBNow() - b.started : 0;

	return (b.elapsed === undefined ? 0 : b.elapsed) + since;
}

// go2jsTestingBLoop is a benchmark asking whether to go on, which is what a
// benchmark written the newer way asks for rather than asking for a number of
// times to run: it goes on until a run has taken as long as the run was asked to
// take, and every call to it is one more time the benchmark has run.
function go2jsTestingBLoop(b) {
	if (b.loop === undefined || b.loop === null) {
		b.loop = {count: 0, started: 0};
	}

	if (b.loop.count === 0) {
		b.loop.started = go2jsTestingBNow();
		b.loop.started = b.started;
	}

	b.loop.count++;
	b.N = b.loop.count;
	b.iterations = b.loop.count;

	return go2jsTestingBElapsed(b) < go2jsTestingBenchTime();
}

// go2jsTestingReportBenchmark says what a benchmark cost each time it was run,
// which is what a benchmark is asked for: how many times it was run, and what one
// of those times cost.
function go2jsTestingReportBenchmark(b) {
	const took = go2jsTestingBElapsed(b);
	const rounds = b.iterations === undefined || b.iterations === 0 ? 1 : b.iterations;
	// What one of them cost is said in nanoseconds, since that is what a cost per
	// operation is said in, and what was timed is in milliseconds.
	const each = ((took / rounds) * 1e6).toFixed(2);
	const procs = typeof navigator !== "undefined" && navigator.hardwareConcurrency
		? navigator.hardwareConcurrency
		: (typeof require === "function" ? require("os").cpus().length : 1);
	let line = String(b.name) + "-" + procs + "\t" + rounds + "\t" + each + " ns/op";

	// What a benchmark says it moved is said as a rate, since that is what a
	// benchmark told a size is asked for: the bytes it moved over how long it was
	// run for, as megabytes a second, which is a mebibyte of them a second.
	if (b.bytes !== undefined && b.bytes > 0 && took > 0) {
		line += "\t" + ((b.bytes * rounds) / took / 1024).toFixed(2) + " MB/s";
	}

	// What a benchmark asked to be told what each of them allocated is told as no
	// allocations at all, since a JavaScript program does not hand out the memory a
	// Go program does, so there is nothing here to count.
	if (b.allocations === true) {
		line += "\t0 B/op\t0 allocs/op";
	}

	for (const unit of Object.keys(b.metrics)) {
		line += "\t" + b.metrics[unit] + " " + unit;
	}

	if (go2jsTestingFailed(b)) {
		go2jsTestingState.failed++;

		go2jsTestingOut("--- FAIL: " + go2jsTestingFullName(b));

		for (const line of go2jsTestingLogs(b)) {
			go2jsTestingOut("    " + line);
		}

		return;
	}

	go2jsTestingOut(line);
}

`
}
