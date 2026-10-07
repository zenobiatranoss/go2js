package javascript

var slicesFuncs = map[string]string{
	"BinarySearch":     "go2jsSlicesBinarySearch",
	"BinarySearchFunc": "go2jsSlicesBinarySearchFunc",
	"Clone":            "go2jsSlicesClone",
	"Compact":          "go2jsSlicesCompact",
	"CompactFunc":      "go2jsSlicesCompactFunc",
	"Contains":         "go2jsSlicesContains",
	"ContainsFunc":     "go2jsSlicesContainsFunc",
	"Delete":           "go2jsSlicesDelete",
	"Equal":            "go2jsSlicesEqual",
	"EqualFunc":        "go2jsSlicesEqualFunc",
	"Index":            "go2jsSlicesIndex",
	"IndexFunc":        "go2jsSlicesIndexFunc",
	"Insert":           "go2jsSlicesInsert",
	"Max":              "go2jsSlicesMax",
	"MaxFunc":          "go2jsSlicesMaxFunc",
	"Min":              "go2jsSlicesMin",
	"MinFunc":          "go2jsSlicesMinFunc",
	"Reverse":          "go2jsSlicesReverse",
	"Sort":             "go2jsSlicesSort",
	"SortFunc":         "go2jsSlicesSortFunc",
	"SortStableFunc":   "go2jsSlicesSortStableFunc",
	"DeleteFunc":       "go2jsSlicesDeleteFunc",
	"Grow":             "go2jsSlicesGrow",
	"Clip":             "go2jsSlicesClip",
	"Collect":          "go2jsSlicesCollect",
	"Repeat":           "go2jsSlicesRepeat",
	"Concat":           "go2jsSlicesConcat",
	"All":              "go2jsSlicesAll",
	"Values":           "go2jsSlicesValues",
	"Backward":         "go2jsSlicesBackward",
	"AppendSeq":        "go2jsSlicesAppendSeq",
	"Sorted":           "go2jsSlicesSorted",
	"Chunk":            "go2jsSlicesChunk",
	"Clear":            "go2jsSlicesClear",
	"SortedFunc":       "go2jsSlicesSortedFunc",
	"SortedStableFunc": "go2jsSlicesSortedStableFunc",
	"Compare":          "go2jsSlicesCompare",
	"CompareFunc":      "go2jsSlicesCompareFunc",
	"IsSorted":         "go2jsSlicesIsSorted",
	"IsSortedFunc":     "go2jsSlicesIsSortedFunc",
	"After":            "go2jsSlicesAfter",
	"Before":           "go2jsSlicesBefore",
	"Replace":          "go2jsSlicesReplace",
}

var mapsFuncs = map[string]string{
	"Clone":      "go2jsMapsClone",
	"Copy":       "go2jsMapsCopy",
	"DeleteFunc": "go2jsMapsDeleteFunc",
	"Equal":      "go2jsMapsEqual",
	"EqualFunc":  "go2jsMapsEqualFunc",
	"Keys":       "go2jsMapsKeys",
	"Values":     "go2jsMapsValues",
	"All":        "go2jsMapsAll",
	"Collect":    "go2jsMapsCollect",
	"Clear":      "go2jsMapsClear",
	"Insert":     "go2jsMapsInsert",
}

var base64Funcs = map[string]string{
	"NewEncoder": "go2jsBase64NewEncoder",
	"NewDecoder": "go2jsBase64NewDecoder",
}

var logFuncs = map[string]string{
	"Print":     "go2jsLogPrint",
	"Printf":    "go2jsLogPrintf",
	"Println":   "go2jsLogPrintln",
	"Fatal":     "go2jsLogFatal",
	"Fatalf":    "go2jsLogFatalf",
	"Fatalln":   "go2jsLogFatalln",
	"Panic":     "go2jsLogPanic",
	"Panicf":    "go2jsLogPanicf",
	"Panicln":   "go2jsLogPanicln",
	"Output":    "go2jsLogOutputMessage",
	"SetOutput": "go2jsLogSetOutput",
	"New":       "go2jsLogNew",
	"SetFlags":  "go2jsLogSetFlags",
	"SetPrefix": "go2jsLogSetPrefix",
	"Flags":     "go2jsLogFlags",
	"Prefix":    "go2jsLogPrefix",
	"Writer":    "go2jsLogWriterValue",
	"Default":   "go2jsLogDefault",
}

var logFlagConstants = map[string]string{
	"Ldate":         "1",
	"Ltime":         "2",
	"Lmicroseconds": "4",
	"Llongfile":     "8",
	"Lshortfile":    "16",
	"LUTC":          "32",
	"Lmsgprefix":    "64",
	"LstdFlags":     "3",
}

var utf16Funcs = map[string]string{
	"Encode":      "go2jsUTF16Encode",
	"Decode":      "go2jsUTF16Decode",
	"IsSurrogate": "go2jsUTF16IsSurrogate",
	"DecodeRune":  "go2jsUTF16DecodeRune",
	"EncodeRune":  "go2jsUTF16EncodeRune",
	"RuneLen":     "go2jsUTF16RuneLen",
	"AppendRune":  "go2jsUTF16AppendRune",
}

var packageVarMethods = map[string]string{
	"base64.Encoding.EncodeToString": "go2jsBase64EncodeToString",
	"base64.Encoding.DecodeString":   "go2jsBase64DecodeString",
	"base64.Encoding.EncodedLen":     "go2jsBase64EncodedLen",
	"base64.Encoding.DecodedLen":     "go2jsBase64DecodedLen",
	"base64.Encoding.Encode":         "go2jsBase64EncoderWrite",
	"base64.Encoding.Decode":         "go2jsBase64DecoderRead",
	"base64.Encoding.Strict":         "go2jsBase64Strict",
	"base64.NewEncoder.Encode":       "go2jsBase64EncoderWrite",
	"base64.NewDecoder.Decode":       "go2jsBase64DecoderRead",
	"fs.FileMode.String":             "go2jsFileModeStringMethod",
	"fs.FileMode.IsDir":              "go2jsFileModeIsDirMethod",
	"fs.FileMode.IsRegular":          "go2jsFileModeIsRegularMethod",
	"fs.FileMode.Type":               "go2jsFileModeTypeMethod",
	"fs.FileMode.Perm":               "go2jsFileModePermMethod",
}

var packageVarValues = map[string]string{
	"base64.StdEncoding":       "go2jsBase64StdEncoding",
	"base64.URLEncoding":       "go2jsBase64URLEncoding",
	"base64.RawStdEncoding":    "go2jsBase64RawStdEncoding",
	"base64.RawURLEncoding":    "go2jsBase64RawURLEncoding",
	"log.Default":              "go2jsLogDefault",
	"io.Discard":               "go2jsIODiscard",
	"context.Canceled":         "go2jsContextCanceled",
	"context.DeadlineExceeded": "go2jsContextDeadlineExceeded",
	"io.EOF":                   "go2jsIOEOFError",
	"io.ErrUnexpectedEOF":      `go2jsSentinelError("unexpected EOF")`,

	"os.ErrNotExist":         `go2jsSentinelError("file does not exist")`,
	"os.ErrExist":            `go2jsSentinelError("file already exists")`,
	"os.ErrClosed":           `go2jsSentinelError("file already closed")`,
	"os.ErrPermission":       `go2jsSentinelError("permission denied")`,
	"os.ErrDeadlineExceeded": `go2jsSentinelError("i/o timeout")`,

	"syscall.ENOENT": `go2jsSentinelError("no such file or directory", "syscall.Errno")`,
	"syscall.EEXIST": `go2jsSentinelError("file exists", "syscall.Errno")`,
	"syscall.EACCES": `go2jsSentinelError("permission denied", "syscall.Errno")`,
	"syscall.EPERM":  `go2jsSentinelError("operation not permitted", "syscall.Errno")`,
	"syscall.EINVAL": `go2jsSentinelError("invalid argument", "syscall.Errno")`,

	"os.Args": `go2jsOSArgs`,

	"filepath.Separator":     `47`,
	"os.PathSeparator":       `47`,
	"os.PathListSeparator":   `58`,
	"filepath.ListSeparator": `58`,
}

func extendedRuntimeSource() string {
	return `
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
			if (go2jsCallNow(compare, null, [scratch[right], scratch[left]]) < 0) {
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
		const result = go2jsCallNow(compare, null, [x, y]);

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
		if (go2jsCallNow(predicate, null, [a[index]])) {
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

// go2jsSlicesCompare is the ordering of two slices, read the way slices.Compare
// reads one: the first pair that differ decides, and equal prefixes fall back
// on the lengths. A nil slice is the empty slice, so Compare(go2jsSlice(), nil)
// is the same 0 Go calls it.
function go2jsSlicesCompare(a, b) {
	const alen = a === null || a === undefined ? 0 : a.length;
	const blen = b === null || b === undefined ? 0 : b.length;
	const common = Math.min(alen, blen);

	for (let index = 0; index < common; index++) {
		const result = go2jsCompareValues(a[index], b[index]);

		if (result !== 0) {
			return result;
		}
	}

	if (alen < blen) {
		return -1;
	}

	if (alen > blen) {
		return 1;
	}

	return 0;
}

// go2jsSlicesCompareFunc is the ordering of two slices given by the comparison
// handed in for one, in the way slices.CompareFunc reads it.
function go2jsSlicesCompareFunc(a, b, compare) {
	const alen = a === null || a === undefined ? 0 : a.length;
	const blen = b === null || b === undefined ? 0 : b.length;
	const common = Math.min(alen, blen);

	for (let index = 0; index < common; index++) {
		const result = go2jsCallNow(compare, null, [a[index], b[index]]);

		if (result !== 0) {
			return result;
		}
	}

	if (alen < blen) {
		return -1;
	}

	if (alen > blen) {
		return 1;
	}

	return 0;
}

// go2jsSlicesIsSorted reports whether the slice is ordered in the way
// slices.IsSorted reads an order, which is the way a < b is read for the
// values themselves.
function go2jsSlicesIsSorted(a) {
	for (let next = 1; next < a.length; next++) {
		if (go2jsCompareValues(a[next], a[next - 1]) < 0) {
			return false;
		}
	}

	return true;
}

// go2jsSlicesIsSortedFunc reports whether the slice is ordered by the
// comparison handed in, the way slices.IsSortedFunc reads one.
function go2jsSlicesIsSortedFunc(a, compare) {
	for (let next = 1; next < a.length; next++) {
		if (go2jsCallNow(compare, null, [a[next], a[next - 1]]) < 0) {
			return false;
		}
	}

	return true;
}

// go2jsSlicesMax is the greatest value of a slice, the way slices.Max reads
// one.
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
		if (go2jsCallNow(compare, null, [item, best]) > 0) {
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
		if (go2jsCallNow(compare, null, [item, best]) < 0) {
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
	return a.filter(item => !go2jsCallNow(predicate, null, [item]));
}

function go2jsSlicesGrow(a, count) {
	if (a.length === 0) {
		return [];
	}

	return a;
}

// go2jsSlicesClear empties a slice, which is what a slice full of zeroes is set
// to and what a slice that had something in it gives back its elements of.
function go2jsSlicesClear(a) {
	if (a === null || a === undefined) {
		return a;
	}

	a.length = 0;
}

// go2jsSlicesAfter and go2jsSlicesBefore are the slice views that start after an
// index or end before one, which is what a slice expression of three indices
// stands for: the elements between the two bounds.
function go2jsSlicesAfter(a, index) {
	return a.slice(index);
}

function go2jsSlicesBefore(a, index) {
	return a.slice(0, index);
}

// go2jsSlicesReplace answers a slice with some of its own elements replaced by
// others, which is the assignment to an element of a slice rather than to the
// slice itself.
function go2jsSlicesReplace(a, index, values) {
	a[index] = values;
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



// A sequence is a function that hands each of its values to whoever asked for
// them, so slices.Values and slices.All answer with one and the functions that
// take a sequence call it with a function of their own. A sequence that is
// walked to its end is walked eagerly, since a yield in the middle of a
// function cannot be left half finished and taken up again where it stopped.

function go2jsSlicesValues(source) {
	return function (handOff) {
		if (source === null || source === undefined) {
			return;
		}

		for (let i = 0; i < source.length; i++) {
			if (!go2jsCallNow(handOff, null, [source[i]])) {
				return;
			}
		}
	};
}

function go2jsSlicesAll(source) {
	return function (handOff) {
		if (source === null || source === undefined) {
			return;
		}

		for (let i = 0; i < source.length; i++) {
			if (!go2jsCallNow(handOff, null, [i, source[i]])) {
				return;
			}
		}
	};
}

function go2jsSlicesBackward(source) {
	return function (handOff) {
		if (source === null || source === undefined) {
			return;
		}

		for (let i = source.length - 1; i >= 0; i--) {
			if (!go2jsCallNow(handOff, null, [i, source[i]])) {
				return;
			}
		}
	};
}

// go2jsSlicesChunk cuts a slice into slices of at most a number of elements
// each, and a slice whose own length does not divide evenly leaves the last
// chunk with what is left rather than padding it out to the same size.
function go2jsSlicesChunk(a, n) {
	const whole = go2jsToArray(a);

	if (n < 1) {
		return null;
	}

	if (whole.length === 0) {
		return [];
	}

	const out = [];

	for (let start = 0; start < whole.length; start += n) {
		out.push(whole.slice(start, start + n));
	}

	return out;
}

function go2jsSlicesCollect(source) {
	if (source === null || source === undefined) {
		return [];
	}

	if (typeof source === "function") {
		const out = [];

		go2jsCallNow(source, null, [function (value) {
			out.push(value);

			return true;
		}]);

		return out;
	}

	if (Array.isArray(source)) {
		return source.slice();
	}

	return Array.from(source);
}

function go2jsSlicesAppendSeq(destination, source) {
	if (source === null || source === undefined) {
		return destination;
	}

	if (typeof source === "function") {
		go2jsCallNow(source, null, [function (value) {
			destination.push(value);

			return true;
		}]);

		return destination;
	}

	for (const item of source) {
		destination.push(item);
	}

	return destination;
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

		if (go2jsCallNow(compare, null, [a[mid], target]) < 0) {
			low = mid + 1;
		} else {
			high = mid;
		}
	}

	return [low, low < a.length && go2jsCallNow(compare, null, [a[low], target]) === 0];
}

// go2jsMapsAll is the sequence of every key of a map along with the value it
// stands for, which is the order Go walks a map in when the walk comes from a
// range over this sequence rather than from a range over the map itself.
function go2jsMapsAll(m) {
	// A walk stops as soon as it is asked to, so the sequence hands back a
	// question rather than an answer and the walk asks it once per entry. The
	// key and the value arrive as two arguments of that question rather than
	// as one pair, which is how a walk over this sequence is written.
	return function (yielded) {
		for (const [key, value] of go2jsMapEntries(m)) {
			if (yielded(key, value) === false) {
				return false;
			}
		}

		return true;
	};
}

// go2jsMapsCollect gathers the pairs a sequence of keys and values walks past
// into a map of its own, which is the map the sequence stands for.
function go2jsMapsCollect(source) {
	if (source === null || source === undefined) {
		return go2jsMap([]);
	}

	if (typeof source === "function") {
		const entries = [];

		go2jsCallNow(source, null, [function (key, value) {
			entries.push([key, value]);

			return true;
		}]);

		return go2jsMap(entries);
	}

	if (Array.isArray(source)) {
		return go2jsMap(source);
	}

	return go2jsMap([]);
}

// go2jsMapsClear empties a map, which is what a map literal is given before the
// entries of one are put into it and what a map is left as once its keys have all
// been deleted.
function go2jsMapsClear(m) {
	for (const key of go2jsMapKeys(m)) {
		go2jsMapDelete(m, key);
	}
}

function go2jsMapsKeys(m) {
	return go2jsMapKeys(m).sort(go2jsCompareValues);
}

function go2jsMapsValues(m) {
	return go2jsMapsKeys(m).map(key => go2jsMapGet(m, key, null));
}

// go2jsMapsInsert puts the pairs a sequence of keys and values walks past into
// a map, replacing a pair that is already held, the way maps.Insert does.
function go2jsMapsInsert(target, sequence) {
	if (sequence === null || sequence === undefined) {
		return target;
	}

	if (typeof sequence === "function") {
		go2jsCallNow(sequence, null, [function (key, value) {
			go2jsMapSet(target, key, value);

			return true;
		}]);
	} else if (Array.isArray(sequence)) {
		for (const entry of sequence) {
			if (Array.isArray(entry)) {
				go2jsMapSet(target, entry[0], entry[1]);
			}
		}
	}

	return target;
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
		if (go2jsCallNow(predicate, null, [entry[0], entry[1]])) {
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
	go2jsLogWriteTo(go2jsLogStandardWriter, text);
}

function go2jsLogDefault() {
	return go2jsLogDefaultLogger;
}

function go2jsLogSetOutput(writer) {
	go2jsLogStandardWriter = go2jsUnwrap(writer) === undefined ? process.stderr : writer;

	return null;
}

function go2jsLogWriterValue() {
	return go2jsLogStandardWriter;
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

function go2jsLogWriteStandard(text) {
	go2jsLogOutput(go2jsLogDecorate(text));
}

function go2jsLogBuildStandard() {
	return {
		Print(...values) {
			go2jsLogWriteStandard(go2jsLogEnsureNewline(go2jsSprint(...values)));
		},
		Printf(format, ...values) {
			go2jsLogWriteStandard(go2jsLogEnsureNewline(go2jsSprintf(format, ...values)));
		},
		Println(...values) {
			go2jsLogWriteStandard(go2jsLogSprintln(values));
		},
		Fatal(...values) {
			go2jsLogWriteStandard(go2jsLogEnsureNewline(go2jsSprint(...values)));
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
			const text = go2jsSprint(...values);
			go2jsLogWriteStandard(go2jsLogEnsureNewline(text));
			go2jsPanic(text);
		},
		Panicf(format, ...values) {
			const text = go2jsSprintf(format, ...values);
			go2jsLogWriteStandard(go2jsLogEnsureNewline(text));
			go2jsPanic(text);
		},
		Panicln(...values) {
			const text = go2jsLogSprintln(values);
			go2jsLogWriteStandard(text);
			go2jsPanic(text);
		},
		Output(calldepth, text) {
			go2jsLogWriteStandard(go2jsLogEnsureNewline(go2jsStringify(text)));
			return null;
		},
		SetOutput(writer) {
			go2jsLogSetOutput(writer);
		},
		Writer() {
			return go2jsLogStandardWriter;
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
	if (text === "") {
		return "\n";
	}

	if (text.endsWith("\n")) {
		return text;
	}

	return text + "\n";
}

function go2jsLogPrint(...values) {
	go2jsLogOutput(go2jsLogDecorate(go2jsLogEnsureNewline(go2jsSprint(...values))));
}

function go2jsLogPrintf(format, ...values) {
	go2jsLogOutput(go2jsLogDecorate(go2jsLogEnsureNewline(go2jsSprintf(format, ...values))));
}

function go2jsLogPrintln(...values) {
	go2jsLogOutput(go2jsLogDecorate(go2jsLogSprintln(values)));
}

function go2jsLogOutputMessage(calldepth, text) {
	go2jsLogOutput(go2jsLogDecorate(go2jsLogEnsureNewline(go2jsStringify(text))));
	return null;
}

function go2jsLogPanic(...values) {
	const text = go2jsSprint(...values);
	go2jsLogOutput(go2jsLogDecorate(go2jsLogEnsureNewline(text)));
	go2jsPanic(text);
}

function go2jsLogPanicf(format, ...values) {
	const text = go2jsSprintf(format, ...values);
	go2jsLogOutput(go2jsLogDecorate(go2jsLogEnsureNewline(text)));
	go2jsPanic(text);
}

function go2jsLogPanicln(...values) {
	const text = go2jsLogSprintln(values);
	go2jsLogOutput(go2jsLogDecorate(text));
	go2jsPanic(text);
}

function go2jsLogFatal(...values) {
	go2jsLogOutput(go2jsLogDecorate(go2jsLogEnsureNewline(go2jsSprint(...values))));
	process.exit(1);
}

function go2jsLogFatalf(format, ...values) {
	go2jsLogOutput(go2jsLogDecorate(go2jsLogEnsureNewline(go2jsSprintf(format, ...values))));
	process.exit(1);
}

function go2jsLogFatalln(...values) {
	go2jsLogOutput(go2jsLogDecorate(go2jsLogSprintln(values)));
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
		go2jsCallNow(inner.Write, inner, [go2jsStringToBytes(text)]);
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
			const stamp = go2jsLogStamp(new Date(), this.flags);

			go2jsLogWriteTo(this.writer, (this.flags & 64) !== 0 ? stamp + this.prefix + chunk : this.prefix + stamp + chunk);
		},
		Print(...values) {
			this.write(go2jsLogEnsureNewline(go2jsSprint(...values)));
		},
		Printf(format, ...values) {
			this.write(go2jsLogEnsureNewline(go2jsSprintf(format, ...values)));
		},
		Println(...values) {
			this.write(go2jsLogSprintln(values));
		},
		Fatal(...values) {
			this.write(go2jsLogEnsureNewline(go2jsSprint(...values)));
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
			const message = go2jsSprint(...values);
			this.write(go2jsLogEnsureNewline(message));
			go2jsPanic(message);
		},
		Panicf(format, ...values) {
			const message = go2jsSprintf(format, ...values);
			this.write(go2jsLogEnsureNewline(message));
			go2jsPanic(message);
		},
		Panicln(...values) {
			const message = go2jsLogSprintln(values);
			this.write(message);
			go2jsPanic(message);
		},
		Output(calldepth, message) {
			this.write(go2jsLogEnsureNewline(go2jsStringify(message)));
			return null;
		},
		SetOutput(target) {
			this.writer = target;
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
var go2jsLogStandardWriter = go2jsLogStandardWriterDefault();
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

function go2jsUTF16RuneLen(value) {
	const r = go2jsUnicodeCodePoint(value);

	if (r >= 0 && r < 0xd800) {
		return 1;
	}

	if (r >= 0xe000 && r <= 0xffff) {
		return 1;
	}

	if (r >= 0x10000 && r <= 0x10ffff) {
		return 2;
	}

	return -1;
}

function go2jsUTF16AppendRune(units, value) {
	const start = units === null || units === undefined ? [] : Array.from(units);
	const r = go2jsUnicodeCodePoint(value);

	if ((r >= 0 && r < 0xd800) || (r >= 0xe000 && r <= 0xffff)) {
		start.push(r);
		return start;
	}

	if (r >= 0x10000 && r <= 0x10ffff) {
		const adjusted = r - 0x10000;

		start.push(0xd800 + (adjusted >> 10));
		start.push(0xdc00 + (adjusted & 0x3ff));
		return start;
	}

	// a rune out of range or a surrogate is written as the replacement char
	start.push(0xfffd);
	return start;
}
`
}

var packageVarTypes = map[string]string{
	"base64.StdEncoding":       "base64.Encoding",
	"base64.URLEncoding":       "base64.Encoding",
	"base64.RawStdEncoding":    "base64.Encoding",
	"base64.RawURLEncoding":    "base64.Encoding",
	"log.Default":              "log.Logger",
	"io.Discard":               "io.Writer",
	"context.Canceled":         "error",
	"context.DeadlineExceeded": "error",
	"io.EOF":                   "error",
	"io.ErrUnexpectedEOF":      "error",

	"os.Args": "[]string",

	"filepath.Separator":     "rune",
	"os.PathSeparator":       "rune",
	"filepath.ListSeparator": "rune",
	"os.PathListSeparator":   "rune",

	"os.ErrNotExist":         "error",
	"os.ErrExist":            "error",
	"os.ErrDeadlineExceeded": "error",
	"os.ErrClosed":           "error",
	"os.ErrPermission":       "error",

	"syscall.ENOENT": "error",
	"syscall.EEXIST": "error",
	"syscall.EACCES": "error",
	"syscall.EPERM":  "error",
	"syscall.EINVAL": "error",
}

var packageVarPaths = map[string]string{
	"base64":  "encoding/base64",
	"io":      "io",
	"log":     "log",
	"context": "context",
}

func sortSliceShimRuntimeSource() string {
	return `function go2jsSortSliceInPlace(values, compare) {
	const state = go2jsSliceState(values);

	if (state === null) {
		return values;
	}

	const window = [];

	for (let index = 0; index < state.length; index++) {
		window.push(state.data[state.offset + index]);
	}

	window.sort(compare);

	for (let index = 0; index < state.length; index++) {
		state.data[state.offset + index] = window[index];
	}

	return values;
}

function go2jsSortStringSlice(values) {
	return go2jsSortSliceInPlace(values, (left, right) => (String(left) < String(right) ? -1 : String(left) > String(right) ? 1 : 0));
}

function go2jsSortIntSlice(values) {
	return go2jsSortSliceInPlace(values, (left, right) => (Number(left) < Number(right) ? -1 : Number(left) > Number(right) ? 1 : 0));
}

function go2jsSortFloat64Slice(values) {
	return go2jsSortSliceInPlace(values, (left, right) => (Number(left) < Number(right) ? -1 : Number(left) > Number(right) ? 1 : 0));
}
`
}

func bytesToStringRuntimeSource() string {
	return `
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
	return Array.from(go2jsRawText(value)).map(char => char.codePointAt(0));
}

// A slice converted from nothing is nil, and nil answers to the question, so
// the empty value it is given says that much while it still behaves as the
// slice it is.
function go2jsNilBytes() {
	return go2jsNilValue("", "slice");
}

function go2jsNilSlice() {
	return go2jsNilValue("", "slice");
}

function go2jsStringToUTF16(value) {
	if (typeof value === "string") {
		return Array.from(value).map(character => character.charCodeAt(0));
	}

	if (value === null || value === undefined) {
		return [];
	}

	return Array.from(value).map(item => Number(item));
}

function go2jsBytesCopy(value) {
	return go2jsStringToBytes(go2jsBytesToString(value));
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
`
}

func extendedStdlibFuncs() {
	extend := func(target map[string]string, entries map[string]string) {
		for name, helper := range entries {
			target[name] = helper
		}
	}

	extend(stringsFuncs, map[string]string{
		"Clone":         "go2jsStringsClone",
		"ContainsAny":   "go2jsStringsContainsAny",
		"ContainsFunc":  "go2jsStringsContainsFunc",
		"IndexAny":      "go2jsStringsIndexAny",
		"IndexByte":     "go2jsStringsIndexByte",
		"IndexFunc":     "go2jsStringsIndexFunc",
		"LastIndexAny":  "go2jsStringsLastIndexAny",
		"LastIndexByte": "go2jsStringsLastIndexByte",
		"FieldsFunc":    "go2jsStringsFieldsFunc",
		"SplitAfter":    "go2jsStringsSplitAfter",
		"SplitAfterN":   "go2jsStringsSplitAfterN",
		"TrimFunc":      "go2jsStringsTrimFunc",
		"TrimLeftFunc":  "go2jsStringsTrimLeftFunc",
		"TrimRightFunc": "go2jsStringsTrimRightFunc",
		"TrimLeft":      "go2jsStringsTrimLeft",
		"TrimRight":     "go2jsStringsTrimRight",
		"Title":         "go2jsStringsTitle",
		"Map":           "go2jsStringsMap",
		"CutSuffix":     "go2jsStringsCutSuffix",
		"ContainsRune":  "go2jsStringsContainsRune",
		"IndexRune":     "go2jsStringsIndexRune",
		"Builder":       "go2jsStringsBuilder",
		"NewReplacer":   "go2jsStringsNewReplacer",
	})

	extend(strconvFuncs, map[string]string{
		"AppendQuote":     "go2jsStrconvAppendQuote",
		"QuoteToASCII":    "go2jsStrconvQuoteToASCII",
		"QuoteRune":       "go2jsStrconvQuoteRune",
		"AppendQuoteRune": "go2jsStrconvAppendQuoteRune",
		"IsPrint":         "go2jsStrconvIsPrint",
		"IsGraphic":       "go2jsStrconvIsGraphic",
		"FormatComplex":   "go2jsStrconvFormatFloat",
	})

	extend(mathFuncs, map[string]string{
		"Cbrt":       "go2jsMathCbrt",
		"Mod":        "go2jsMathMod",
		"Remainder":  "go2jsMathRemainder",
		"Hypot":      "go2jsMathHypot",
		"Gamma":      "go2jsMathGamma",
		"Logb":       "go2jsMathLogb",
		"Sinh":       "go2jsMathSinh",
		"Cosh":       "go2jsMathCosh",
		"Tanh":       "go2jsMathTanh",
		"Asinh":      "go2jsMathAsinh",
		"Acosh":      "go2jsMathAcosh",
		"Atanh":      "go2jsMathAtanh",
		"Copysign":   "go2jsMathCopysign",
		"Dim":        "go2jsMathDim",
		"Modf":       "go2jsMathModf",
		"Sincos":     "go2jsMathSincos",
		"MaxInt":     "go2jsMathMaxInt",
		"MinInt":     "go2jsMathMinInt",
		"MaxInt8":    "go2jsMathMaxInt8",
		"MinInt8":    "go2jsMathMinInt8",
		"MaxInt16":   "go2jsMathMaxInt16",
		"MinInt16":   "go2jsMathMinInt16",
		"MaxInt32":   "go2jsMathMaxInt32",
		"MinInt32":   "go2jsMathMinInt32",
		"MaxInt64":   "go2jsMathMaxInt64",
		"MinInt64":   "go2jsMathMinInt64",
		"MaxUint":    "go2jsMathMaxUint",
		"MaxUint8":   "go2jsMathMaxUint8",
		"MaxUint16":  "go2jsMathMaxUint16",
		"MaxUint32":  "go2jsMathMaxUint32",
		"MaxUint64":  "go2jsMathMaxUint64",
		"MaxFloat32": "go2jsMathMaxFloat32",
		"MaxFloat64": "go2jsMathMaxFloat64",
		"SqrtPhi":    "go2jsMathSqrtPhi",
		"Ln10":       "go2jsMathLn10",
		"Log2E":      "go2jsMathLog2E",
		"Ln2":        "go2jsMathLn2",
		"Phi":        "go2jsMathPhi",
		"SqrtE":      "go2jsMathSqrtE",
		"SqrtPi":     "go2jsMathSqrtPi",
		"Sqrt2":      "go2jsMathSqrt2",
	})

	extend(sortFuncs, map[string]string{
		"Slice":             "go2jsSortSlice",
		"SliceStable":       "go2jsSortSliceStable",
		"StringsAreSorted":  "go2jsSortStringsAreSorted",
		"IntsAreSorted":     "go2jsSortIntsAreSorted",
		"Float64sAreSorted": "go2jsSortFloat64sAreSorted",
		"SearchStrings":     "go2jsSortSearchStrings",
		"SearchInts":        "go2jsSortSearchInts",
		"SearchFloat64s":    "go2jsSortSearchFloat64s",
		"Sort":              "go2jsSortSlice",
	})

	extend(osFuncs, map[string]string{
		"ReadFile":     "go2jsOSReadFile",
		"WriteFile":    "go2jsOSWriteFile",
		"Open":         "go2jsOSOpen",
		"OpenFile":     "go2jsOSOpenFile",
		"Create":       "go2jsOSCreate",
		"Remove":       "go2jsOSRemove",
		"RemoveAll":    "go2jsOSRemoveAll",
		"Rename":       "go2jsOSRename",
		"CreateTemp":   "go2jsOSCreateTemp",
		"MkdirTemp":    "go2jsOSMkdirTemp",
		"MkdirAll":     "go2jsOSMkdirAll",
		"Chmod":        "go2jsOSChmod",
		"Truncate":     "go2jsOSTruncate",
		"Symlink":      "go2jsOSSymlink",
		"Readlink":     "go2jsOSReadlink",
		"Link":         "go2jsOSLink",
		"Chtimes":      "go2jsOSChtimes",
		"Hostname":     "go2jsOSHostname",
		"Executable":   "go2jsOSExecutable",
		"TempDir":      "go2jsOSTempDir",
		"Getpid":       "go2jsOSGetpid",
		"Exit":         "go2jsOSExit",
		"IsNotExist":   "go2jsOSIsNotExist",
		"IsPermission": "go2jsOSIsPermission",
		"IsExist":      "go2jsOSIsExist",
		"IsTimeout":    "go2jsOSIsTimeout",
		"Getwd":        "go2jsOSGetwd",
		"Chdir":        "go2jsOSChdir",
		"ReadDir":      "go2jsOSReadDir",
		"Mkdir":        "go2jsOSMkdir",
	})

	extend(bytesFuncs, map[string]string{
		"Contains":        "go2jsBytesContains",
		"ContainsAny":     "go2jsBytesContainsAny",
		"ContainsFunc":    "go2jsBytesContainsFunc",
		"Count":           "go2jsBytesCount",
		"EqualFold":       "go2jsBytesEqualFold",
		"HasPrefix":       "go2jsBytesHasPrefix",
		"HasSuffix":       "go2jsBytesHasSuffix",
		"Index":           "go2jsBytesIndex",
		"IndexAny":        "go2jsBytesIndexAny",
		"IndexByte":       "go2jsBytesIndexByte",
		"Join":            "go2jsBytesJoin",
		"LastIndex":       "go2jsBytesLastIndex",
		"Repeat":          "go2jsBytesRepeat",
		"Replace":         "go2jsBytesReplace",
		"ReplaceAll":      "go2jsBytesReplaceAll",
		"Split":           "go2jsBytesSplit",
		"SplitN":          "go2jsBytesSplitN",
		"Title":           "go2jsBytesTitle",
		"ToLower":         "go2jsBytesToLower",
		"ToUpper":         "go2jsBytesToUpper",
		"Trim":            "go2jsBytesTrim",
		"TrimSpace":       "go2jsBytesTrimSpace",
		"TrimPrefix":      "go2jsBytesTrimPrefix",
		"TrimSuffix":      "go2jsBytesTrimSuffix",
		"Fields":          "go2jsBytesFields",
		"NewBuffer":       "go2jsBytesNewBuffer",
		"NewBufferString": "go2jsBytesNewBufferString",
		"NewReader":       "go2jsBytesNewReader",
		"Runes":           "go2jsBytesRunes",
		"CutPrefix":       "go2jsBytesCutPrefix",
		"CutSuffix":       "go2jsBytesCutSuffix",
		"Compare":         "go2jsBytesCompare",
		"Cut":             "go2jsBytesCut",
		"Clone":           "go2jsBytesClone",
		"ContainsRune":    "go2jsBytesContainsRune",
		"ToTitle":         "go2jsBytesToTitle",
		"Map":             "go2jsBytesMap",
		"SplitAfter":      "go2jsBytesSplitAfter",
		"LastIndexAny":    "go2jsBytesLastIndexAny",
	})

	extend(ioFuncs, map[string]string{
		"ReadFull":    "go2jsIOReadFull",
		"WriteString": "go2jsIOWriteString",
		"NopCloser":   "go2jsIONopCloser",
		"MultiReader": "go2jsIOMultiReader",
		"MultiWriter": "go2jsIOMultiWriter",
		"LimitReader": "go2jsIOLimitReader",
		"TeeReader":   "go2jsIOTeeReader",
		"Copy":        "go2jsIOCopy",
		"ReadAll":     "go2jsIOReadAll",
	})

	extend(errorsFuncs, map[string]string{
		"As":     "go2jsErrorsAs",
		"Is":     "go2jsErrorsIs",
		"Join":   "go2jsErrorsJoin",
		"Unwrap": "go2jsErrorsUnwrap",
		"New":    "go2jsErrorsNew",
	})

	extend(utf8Funcs, map[string]string{
		"ValidString":            "go2jsUTF8ValidString",
		"Valid":                  "go2jsUTF8Valid",
		"RuneCountInString":      "go2jsUTF8RuneCountInString",
		"RuneLen":                "go2jsUTF8RuneLen",
		"RuneCount":              "go2jsUTF8RuneCount",
		"RuneStart":              "go2jsUTF8RuneStart",
		"DecodeRuneInString":     "go2jsUTF8DecodeRuneInString",
		"DecodeRune":             "go2jsUTF8DecodeRune",
		"AppendRune":             "go2jsUTF8AppendRune",
		"DecodeLastRuneInString": "go2jsUTF8DecodeLastRuneInString",
	})

	extend(randFuncs, map[string]string{
		"Float32": "go2jsRandFloat64",
		"Uint32":  "go2jsRandUint32",
		"Uint64":  "go2jsRandInt63",
	})

	extend(timeFuncs, map[string]string{
		"Since":         "go2jsTimeSince",
		"Until":         "go2jsTimeUntil",
		"Now":           "go2jsTimeNow",
		"Sleep":         "go2jsTimeSleep",
		"Unix":          "go2jsTimeFromUnix",
		"Parse":         "go2jsTimeParse",
		"After":         "go2jsTimeAfter",
		"NewTimer":      "go2jsTimeNewTimer",
		"NewTicker":     "go2jsTimeNewTicker",
		"AfterFunc":     "go2jsTimeAfterFunc",
		"ParseDuration": "go2jsParseDuration",
		"Tick":          "go2jsTimeTick",
		"FixedZone":     "go2jsTimeFixedZone",
		"LoadLocation":  "go2jsTimeLoadLocation",
	})

}

func extendedRuntimeSource2() string {
	return `
function go2jsStringsClone(value) {
	return go2jsStringify(value);
}

function go2jsStringsContainsAny(value, chars) {
	const set = Array.from(go2jsStringify(chars));

	return Array.from(go2jsStringify(value)).some(char => set.includes(char));
}

function go2jsStringsIndexAny(value, chars) {
	const text = go2jsStringify(value);
	const set = Array.from(go2jsStringify(chars));
	const { chars: runes, starts } = go2jsStringsByteOffsets(text);

	for (let index = 0; index < runes.length; index++) {
		if (set.includes(runes[index])) {
			return starts[index];
		}
	}

	return -1;
}

function go2jsStringsLastIndexAny(value, chars) {
	const text = go2jsStringify(value);
	const set = Array.from(go2jsStringify(chars));
	const { chars: runes, starts } = go2jsStringsByteOffsets(text);

	for (let index = runes.length - 1; index >= 0; index--) {
		if (set.includes(runes[index])) {
			return starts[index];
		}
	}

	return -1;
}

function go2jsStringsIndexByte(value, b) {
	return go2jsStringsIndexRune(value, b);
}

function go2jsStringsLastIndexByte(value, b) {
	return go2jsStringsLastIndexRune(value, b);
}

function go2jsStringsIndexRune(value, r) {
	const target = String.fromCodePoint(Number(r));
	const { chars, starts } = go2jsStringsByteOffsets(go2jsStringify(value));

	for (let index = 0; index < chars.length; index++) {
		if (chars[index] === target) {
			return starts[index];
		}
	}

	return -1;
}

function go2jsStringsLastIndexRune(value, r) {
	const target = String.fromCodePoint(Number(r));
	const { chars, starts } = go2jsStringsByteOffsets(go2jsStringify(value));

	for (let index = chars.length - 1; index >= 0; index--) {
		if (chars[index] === target) {
			return starts[index];
		}
	}

	return -1;
}

function go2jsStringsByteOffsets(text) {
	const chars = Array.from(text);
	const starts = [];
	let offset = 0;

	for (const char of chars) {
		starts.push(offset);

		const code = char.codePointAt(0);

		if (code < 0x80) {
			offset += 1;
		} else if (code < 0x800) {
			offset += 2;
		} else {
			offset += code < 0x10000 ? 3 : 4;
		}
	}

	return { chars, starts };
}

function go2jsStringsIndexFunc(value, predicate) {
	const { chars, starts } = go2jsStringsByteOffsets(go2jsStringify(value));

	for (let index = 0; index < chars.length; index++) {
		if (go2jsCallNow(predicate, null, [chars[index].codePointAt(0)])) {
			return starts[index];
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
		if (go2jsCallNow(predicate, null, [char.codePointAt(0)])) {
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

	while (start < end && go2jsCallNow(predicate, null, [chars[start].codePointAt(0)])) {
		start++;
	}

	while (end > start && go2jsCallNow(predicate, null, [chars[end - 1].codePointAt(0)])) {
		end--;
	}

	return chars.slice(start, end).join("");
}

function go2jsStringsTrimLeftFunc(value, predicate) {
	const chars = Array.from(go2jsStringify(value));
	let start = 0;

	while (start < chars.length && go2jsCallNow(predicate, null, [chars[start].codePointAt(0)])) {
		start++;
	}

	return chars.slice(start).join("");
}

function go2jsStringsTrimRightFunc(value, predicate) {
	const chars = Array.from(go2jsStringify(value));
	let end = chars.length;

	while (end > 0 && go2jsCallNow(predicate, null, [chars[end - 1].codePointAt(0)])) {
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
		.map(char => go2jsCallNow(mapping, null, [char.codePointAt(0)]))
		.filter(code => code >= 0)
		.map(code => String.fromCodePoint(code))
		.join("");
}

function go2jsStringsContainsRune(value, r) {
	return go2jsStringify(value).includes(String.fromCodePoint(Number(r)));
}

function go2jsStringsNewReplacer(...args) {
	if (args.length % 2 === 1) {
		go2jsPanic("strings.NewReplacer: odd argument count");
	}

	// A replacement is looked for at each place in the text, and the first
	// pair in the argument order whose pattern fits there is the one that
	// stands, the way Go's trie picks the pair with the highest priority. The
	// empty pattern fits every place, and after it has stood an empty
	// replacement only one byte of the text is walked on, so a place never
	// gets two of them. The text is walked byte by byte, the way Go walks it,
	// so an empty pattern stands between the bytes of a rune as much as
	// between the runes.
	const olds = [];
	const news = [];

	for (let index = 0; index + 1 < args.length; index += 2) {
		olds.push(go2jsStringToBytes(go2jsStringify(args[index])));
		news.push(go2jsStringToBytes(go2jsStringify(args[index + 1])));
	}

	function patternFits(pattern, bytes, at) {
		for (let index = 0; index < pattern.length; index++) {
			if (bytes[at + index] !== pattern[index]) {
				return false;
			}
		}

		return true;
	}

	function replacedBytes(text) {
		const bytes = go2jsStringToBytes(go2jsStringify(text));
		const out = [];
		let position = 0;
		let last = 0;
		let prevEmpty = false;

		while (position <= bytes.length) {
			let match = -1;
			let keylen = 0;

			for (let index = 0; index < olds.length; index++) {
				const pattern = olds[index];

				if (prevEmpty && pattern.length === 0) {
					continue;
				}

				if (pattern.length === 0 || patternFits(pattern, bytes, position)) {
					match = index;
					keylen = pattern.length;
					break;
				}
			}

			if (match >= 0) {
				for (let index = last; index < position; index++) {
					out.push(bytes[index]);
				}

				for (const piece of news[match]) {
					out.push(piece);
				}

				prevEmpty = keylen === 0;
				position += keylen;
				last = position;
				continue;
			}

			prevEmpty = false;
			position++;
		}

		if (last !== bytes.length) {
			for (let index = last; index < bytes.length; index++) {
				out.push(bytes[index]);
			}
		}

		return out;
	}

	return {
		Replace(text) {
			return new TextDecoder().decode(Uint8Array.from(replacedBytes(text)));
		},
		WriteString(writer, text) {
			const result = go2jsCallMethod(writer, "Write", replacedBytes(text));
			const wrote = Array.isArray(result) ? result[0] : Number(result);
			const failure = Array.isArray(result) ? result[1] : null;
			return [wrote, failure];
		},
		ReplaceAll(text) {
			return this.Replace(text);
		},
	};
}

function go2jsStrconvAppendQuote(target, value) {
	return go2jsStrconvAppendBytes(target, go2jsStrconvQuote(value));
}

function go2jsStrconvQuoteRune(value) {
	return go2jsStrconvQuoteRuneWith(value, false, false);
}

function go2jsStrconvAppendQuoteRune(target, value) {
	return go2jsStrconvAppendBytes(target, go2jsStrconvQuoteRune(value));
}

// go2jsStrconvRuneText turns a rune into the one character it stands for, so a
// question asked about a rune is asked about the character it names rather than
// about the number that names it.
function go2jsStrconvRuneText(value) {
	if (typeof value === "number" && Number.isInteger(value)) {
		if (value < 0 || value > 1114111) {
			return "�";
		}

		return String.fromCodePoint(value);
	}

	return go2jsStringify(value);
}

function go2jsStrconvIsPrint(value) {
	const text = go2jsStrconvRuneText(value);

	return text.length > 0 && !/[\x00-\x1f\x7f]/.test(text);
}

function go2jsStrconvIsGraphic(value) {
	const text = go2jsStrconvRuneText(value);

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

// go2jsSortFind is the lowest position the comparison handed in answers with a
// 0 or less at, and whether the value at the position answers with a 0, the
// way sort.Find reads one. The comparison is the one Go asks for: it returns
// less than 0 when the target comes before the entry, 0 when they are equal,
// and more than 0 when the entry comes first. The search is binary, so the
// position is found in the same turns the Go search takes.
function go2jsSortFind(count, compare) {
	const total = Number(count);
	let low = 0;
	let high = total;

	while (low < high) {
		const mid = Math.floor((low + high) / 2);

		if (go2jsCallNow(compare, null, [mid]) > 0) {
			low = mid + 1;
		} else {
			high = mid;
		}
	}

	return [low, low < total && go2jsCallNow(compare, null, [low]) === 0];
}








function go2jsOSExit(code) {
	process.exit(code === undefined ? 0 : Number(code));
}

const go2jsErrNotExistMessage = "file does not exist";
const go2jsErrPermissionMessage = "permission denied";
const go2jsErrExistMessage = "file already exists";
const go2jsErrDeadlineMessage = "i/o timeout";

function go2jsOSIsNotExist(err) {
	if (err === null || err === undefined) {
		return false;
	}

	const value = go2jsUnwrap(err);

	if (value === null || value === undefined) {
		return false;
	}

	if (value.code === "ENOENT" || value.errno === 2 || value.errno === -2) {
		return true;
	}

	return go2jsErrorMessage(value) === go2jsErrNotExistMessage;
}

function go2jsOSIsPermission(err) {
	if (err === null || err === undefined) {
		return false;
	}

	const value = go2jsUnwrap(err);

	if (value === null || value === undefined) {
		return false;
	}

	if (value.code === "EACCES" || value.code === "EPERM" || value.errno === 13 || value.errno === -13) {
		return true;
	}

	return go2jsErrorMessage(value) === go2jsErrPermissionMessage;
}

function go2jsOSIsExist(err) {
	if (err === null || err === undefined) {
		return false;
	}

	const value = go2jsUnwrap(err);

	if (value === null || value === undefined) {
		return false;
	}

	if (value.code === "EEXIST" || value.errno === 17 || value.errno === -17) {
		return true;
	}

	return go2jsErrorMessage(value) === go2jsErrExistMessage;
}

function go2jsOSIsTimeout(err) {
	if (err === null || err === undefined) {
		return false;
	}

	const value = go2jsUnwrap(err);

	if (value === null || value === undefined) {
		return false;
	}

	if (value.code === "ETIMEDOUT" || value.errno === 110 || value.errno === -110) {
		return true;
	}

	if (value.timeout === true) {
		return true;
	}

	return go2jsErrorMessage(value) === go2jsErrDeadlineMessage;
}

function go2jsOSGetwd() {
	try {
		return [require("process").cwd(), null];
	} catch (err) {
		return ["", go2jsOSHostError(err, "getwd", "")];
	}
}

function go2jsOSChdir(path) {
	try {
		require("process").chdir(String(path));
		return null;
	} catch (err) {
		return go2jsOSWrapNodeError(err);
	}
}

// go2jsOSDirEntry is what a directory says about one name in it, asked of as the
// methods a Go program asks it of rather than as the fields of the host.
function go2jsOSDirEntry(dir, name, isDirectory) {
	const known = typeof isDirectory === "boolean" ? isDirectory : null;
	const entry = {
		__go2js_type: "*os.unixDirent",
		Name: function () {
			return String(name);
		},
		IsDir: function () {
			if (known !== null) {
				return known;
			}

			try {
				return require("fs").statSync(String(dir) + "/" + String(name)).isDirectory();
			} catch (err) {
				return false;
			}
		},
		Type: function () {
			// A file that is only a file has no mark of its own to say so with,
			// and a directory is the one that is marked as being one.
			let mode = 0;

			if (entry.IsDir()) {
				mode = 2147483648;
			} else {
				try {
					if (require("fs").lstatSync(String(dir) === "" || String(dir) === "." ? String(name) : (String(dir).endsWith("/") ? String(dir) + String(name) : String(dir) + "/" + String(name))).isSymbolicLink()) {
						mode = 67108864;
					}
				} catch (err) {
					mode = 0;
				}
			}

			// The mode a file is given is asked of by its marks, and the marks
			// are written the way a file mode is written when it is shown.
			return go2jsFileMode(mode);
		},
		// What an entry says about itself is asked of as the pair a Go method
		// answers with, since asking costs a look at the filesystem.
		Info: function () {
			try {
				const stats = require("fs").lstatSync(String(dir) + "/" + String(name));

				return [go2jsOSFileInfo(String(name), stats), null];
			} catch (err) {
				return [null, go2jsOSHostError(err, "lstat", String(dir) + "/" + String(name))];
			}
		}
	};

	return entry;
}

function go2jsOSReadDir(path) {
	const target = String(path);

	try {
		const names = require("fs").readdirSync(target).sort();
		const entries = names.map(name => go2jsOSDirEntry(target, name));

		return [entries, null];
	} catch (err) {
		return [null, go2jsOSHostError(err, "readdirent", target)];
	}
}

function go2jsOSWrapNodeError(err) {
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
	const set = go2jsStringToBytes(go2jsStringify(chars));

	return go2jsToArray(a).some(item => set.includes(item));
}

// go2jsBytesContainsRune answers whether a byte slice holds a rune, which is a
// question about the characters of the slice rather than about its bytes.
function go2jsBytesContainsRune(a, r) {
	const bytes = go2jsToArray(a);
	const code = Number(r);
	const wanted = go2jsStringToBytes(String.fromCodePoint(code));

	for (let index = 0; index < bytes.length; index++) {
		// A character outside the plain range is written in more than one byte,
		// and those bytes are compared against the bytes the rune is written as.
		let matched = true;

		for (let offset = 0; offset < wanted.length; offset++) {
			if (bytes[index + offset] !== wanted[offset]) {
				matched = false;
				break;
			}
		}

		if (matched) {
			return true;
		}
	}

	return false;
}

function go2jsBytesToTitle(a) {
	return go2jsStringToBytes(go2jsRawText(a).toUpperCase());
}

// go2jsBytesMap writes each byte of a slice through a function, and a function
// that answers a minus one takes the byte out of the slice rather than
// replacing it.
function go2jsBytesMap(mapping, a) {
	if (typeof mapping !== "function") {
		return null;
	}

	const out = [];

	for (const byte of go2jsToArray(a)) {
		// The mapping is a function of the program rather than one of the
		// runtime, so it is called the way the runtime calls any of those.
		const replaced = go2jsCallNow(mapping, null, [byte]);

		if (Number(replaced) < 0) {
			continue;
		}

		out.push(Number(replaced) & 255);
	}

	return out;
}

// go2jsBytesSplitAfter splits a slice in two at each separator and keeps the
// separator at the end of the piece before it, which is the difference between
// this and a split that leaves it out.
function go2jsBytesSplitAfter(a, sep) {
	const whole = go2jsToArray(a);
	const separator = go2jsToArray(sep);
	const out = [];

	if (separator.length === 0) {
		return [whole];
	}

	let start = 0;

	while (start <= whole.length) {
		const at = go2jsBytesIndex(whole.slice(start), separator);

		if (at < 0) {
			out.push(whole.slice(start));
			break;
		}

		out.push(whole.slice(start, start + at + separator.length));
		start += at + separator.length;
	}

	return out;
}

// go2jsBytesLastIndexAny answers where the last of a set of bytes sits, and a
// byte that is not there at all is answered as one that is not there.
function go2jsBytesLastIndexAny(a, chars) {
	const set = go2jsStringToBytes(go2jsStringify(chars));
	const haystack = go2jsToArray(a);

	for (let index = haystack.length - 1; index >= 0; index--) {
		if (set.includes(haystack[index])) {
			return index;
		}
	}

	return -1;
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
		.map(part => go2jsStringToBytes(part));
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
	return go2jsNewByteReader(go2jsToArray(data), "slice");
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

// go2jsBytesCut splits a byte slice in two at the first place a separator sits,
// and hands back the part before it and the part after it along with whether
// the separator was there at all. A separator that is not there leaves the slice
// whole and says so.
function go2jsBytesCut(a, sep) {
	const whole = go2jsToArray(a);
	const separator = go2jsToArray(sep);
	const at = separator.length === 0 ? 0 : go2jsBytesIndex(whole, separator);

	if (at < 0) {
		return [whole, [], false];
	}

	return [whole.slice(0, at), whole.slice(at + separator.length), true];
}

// go2jsBytesClone copies a byte slice, and the copy is a slice of its own so
// writing onto it leaves the one it was copied from as it was.
function go2jsBytesClone(b) {
	return go2jsToArray(b).slice();
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
		return go2jsInterfaceCallNow(target, method, ...args);
	}

	const fn = target[method];

	if (typeof fn !== "function") {
		throw new TypeError("method " + method + " is not implemented");
	}

	return go2jsCallNow(fn, target, args);
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

function go2jsCryptoRandomBytes(count) {
	const bytes = new Uint8Array(count);

	if (typeof crypto === "object" && typeof crypto.getRandomValues === "function") {
		crypto.getRandomValues(bytes);

		return bytes;
	}

	for (let index = 0; index < count; index++) {
		bytes[index] = Math.floor(Math.random() * 256);
	}

	return bytes;
}

function go2jsCryptoRandRead(buffer) {
	const target = go2jsUnwrap(buffer);
	const length = go2jsLen(target);
	const random = go2jsCryptoRandomBytes(length);

	for (let index = 0; index < length; index++) {
		target[index] = random[index];
	}

	return [length, null];
}

function go2jsCryptoRandReader() {
	return {
		Read: go2jsCryptoRandRead
	};
}

function go2jsIOEOFError() {
	return go2jsIOEOF();
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

function go2jsSentinelError(message, typeName) {
	return function go2jsSentinelErrorValue() {
		const name = typeName === undefined || typeName === null || typeName === "" ? "*errors.errorString" : typeName;
		const key = name + "\u0000" + message;
		let error = go2jsSentinelErrors.get(key);

		if (error === undefined) {
			error = go2jsNameError(new Error(message), name);

			const named = go2jsErrnoSentinelMessages.get(message);

			if (named !== undefined) {
				error.Is = function (other) {
					return go2jsErrorMessage(other) === named;
				};
			}

			go2jsSentinelErrors.set(key, error);
		}

		return error;
	};
}

function go2jsIOEOF() {
	return go2jsSentinelError("EOF")();
}

function go2jsIODiscard() {
	const writer = go2jsInterface({
		Write(buffer) {
			return go2jsToArray(buffer).length;
		},
	}, "io.Writer");

	writer.__go2js_discard = true;

	return writer;
}

function go2jsIOReadFull(reader, buffer) {
	let total = 0;
	const length = go2jsLen(buffer);

	while (total < length) {
		const chunk = go2jsSliceFrom(buffer, total);
		const result = go2jsCallMethod(reader, "Read", chunk);

		const read = Array.isArray(result) ? result[0] : Number(result);

		if (!read || read <= 0) {
			return [total, go2jsIOUnexpectedEOFErorror()];
		}

		total += read;
	}

	return [total, null];
}
function go2jsSliceFrom(value, start) {
	const state = go2jsSliceState(value);

	if (!state) {
		return go2jsToArray(value).slice(start);
	}

	return go2jsSliceView(state.data, state.offset + start, state.length - start, state.capacity - start);
}

function go2jsIOUnexpectedEOFErorror() {
	return new Error("unexpected EOF");
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

function go2jsErrorsAs(err, target, wanted) {
	if (err === null || err === undefined) {
		return false;
	}

	// A target that is not a pointer, or a pointer standing for nothing, is
	// refused before anything is looked at in the error.
	if (go2jsIsNilTarget(target)) {
		go2jsPanic("errors: target must be a non-nil pointer");
	}

	if (wanted === undefined || wanted === null || wanted === "" || wanted === "any") {
		const declared = go2jsNewTypeOf(target);

		wanted = declared !== "" ? declared : go2jsErrorNameOf(target);
	}

	wanted = String(wanted).replace(/^\*/, "");

	if (wanted === "error" || wanted === "any" || wanted === "interface {}") {
		go2jsStoreErrorTarget(target, err);

		return true;
	}

	if (go2jsSameTypeName(go2jsErrorNameOf(err), wanted)) {
		go2jsStoreErrorTarget(target, err);

		return true;
	}

	// The target may name an interface rather than a type of its own, and an
	// error is stored in it whenever it has every method the interface asks
	// for, which is the question Go asks before it stores one.
	if (go2jsInterfaceMethods(wanted) !== undefined && go2jsSatisfiesInterface(err, wanted)) {
		go2jsStoreErrorTarget(target, err);

		return true;
	}

	return go2jsErrorUnwrapAll(err).some(part => go2jsErrorsAs(part, target, wanted));
}

// go2jsErrorsAsCheckTarget judges a target whose type only the value carries.
// Go looks at the type the value actually has, so a plain value, a pointer
// standing for nothing, and a pointer to a type that cannot hold an error each
// draw their own complaint. A target that is none of those is handed back
// untouched, and the call goes on with it.
function go2jsErrorsAsCheckTarget(target) {
	if (go2jsIsNilTarget(target)) {
		throw new TypeError("errors: target must be a non-nil pointer");
	}

	const pointer = target.__go2js_pointer === true;
	let declared = go2jsNewTypeOf(target);

	// A value of interface type carries the name of its own type with it, and
	// that is the type Go judges the target by.
	if ((declared === "" || declared === undefined) && target.__go2js_interface === true) {
		declared = typeof target.type === "string" ? target.type : target.__go2js_type_name;
	}

	if (declared === "" || declared === undefined) {
		if (pointer) {
			// The address of a variable says nothing about what is in it yet.
			return target;
		}

		declared = go2jsErrorNameOf(target);
	}

	if (typeof declared !== "string" || declared === "") {
		return target;
	}

	let element = declared;

	if (element.startsWith("*")) {
		element = element.slice(1);
	} else if (!pointer) {
		// Whatever it is, it is not a pointer, and Go asks for one.
		go2jsPanic("errors: target must be a non-nil pointer");
	}

	// What the pointer points at is the type that has to hold an error, and it
	// has to be that type itself: Go does not take an address on the way in. A
	// pointer to a pointer is one JavaScript value here, so for one of those the
	// method on the type behind it is the one that counts.
	if (go2jsMethodTable[element + ".Error"] !== undefined) {
		return target;
	}

	if (pointer && go2jsMethodTable["*" + element + ".Error"] !== undefined) {
		return target;
	}

	go2jsPanic("errors: *target must be interface or implement error");
}

// go2jsIsNilTarget reports whether a target handed to errors.As is missing,
// which covers a nil interface as well as a nil pointer of any depth.
function go2jsIsNilTarget(target) {
	if (target === null || target === undefined) {
		return true;
	}

	if (target.__go2js_pointer === true) {
		return target.get === undefined && target.value === null;
	}

	return false;
}

function go2jsErrorNameOf(value) {
	if (value === null || value === undefined) {
		return "";
	}

	if (value.__go2js_error_name !== undefined) {
		return value.__go2js_error_name;
	}

	if (value.__go2js_interface === true) {
		return value.__go2js_error_name === undefined
			? go2jsErrorNameOf(value.value)
			: value.__go2js_error_name;
	}

	if (value instanceof Error) {
		return "error";
	}

	if (typeof value === "object") {
		const ctor = value.constructor;

		if (ctor && typeof ctor.name === "string" && ctor.name !== "Object") {
			const registered = go2jsLookupTypeName(ctor.name);

			if (registered !== undefined) {
				return String(registered).replace(/^\*/, "");
			}
		}
	}

	return typeof value;
}

function go2jsStoreErrorTarget(target, value) {
	if (target === null || typeof target !== "object") {
		return;
	}

	value = go2jsUnwrap(value);

	if (go2jsPointerAccessor(target, go2jsPointerSet)) {
		target[go2jsPointerSet](value);
		return;
	}

	try {
		target.value = value;
	} catch (err) {
		void err;
	}
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
	return go2jsRandSourceUint32(go2jsRandDefault());
}

function go2jsTimeSince(t) {
	return go2jsTimeSub(go2jsTimeNow(), t);
}

function go2jsTimeUntil(t) {
	return go2jsTimeSub(t, go2jsTimeNow());
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

// A sleep is a receive on a channel that comes due at the moment it names, so
// it is written as one: the goroutine sleeping waits there like any other, and
// the scheduler decides when the moment has arrived.
function* go2jsTimeSleep(d) {
	yield* go2jsChanRecv(go2jsTimeAfter(d));
}

// go2jsTimeFromUnix is the moment a count of seconds since the epoch names,
// counted from the epoch rather than from a year, with the nanoseconds beside
// it brought inside the second they belong to the way Go brings them, and kept
// in the zone of the machine rather than in UTC, since that is the zone Go
// counts from.
function go2jsTimeFromUnix(sec, nsec) {
	let seconds = Math.trunc(Number(sec));
	let rest = Math.trunc(Number(nsec === undefined || nsec === null ? 0 : nsec));

	if (rest < 0 || rest >= 1000000000) {
		const whole = Math.trunc(rest / 1000000000);

		seconds += whole;
		rest -= whole * 1000000000;

		if (rest < 0) {
			rest += 1000000000;
			seconds--;
		}
	}

	const moment = go2jsTimeValue(new Date(seconds * 1000 + Math.floor(rest / 1000000)));

	moment.__go2js_nsec = rest % 1000000;
	moment.__go2js_location = go2jsTimeHostZoneName();
	moment.__go2js_zone = moment.__go2js_location === "UTC" ? null : moment.__go2js_location;

	return moment;
}

// go2jsTimeParse reads a moment out of text the way a layout says it is
// written. A layout is read one piece at a time and each piece takes what it
// stands for from the text where it stands, so text that does not hold what a
// piece asks for is refused rather than read as something else.
function go2jsTimeParse(layout, value) {
	const text = String(value);
	const parts = { year: 0, month: 1, day: 1, hour: 0, minute: 0, second: 0, milli: 0, nsec: 0, offset: null, pm: null };

	const refuse = function() {
		return [null, go2jsTimeParseError(String(layout), text)];
	};

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

	const literal = function(wanted) {
		if (text.slice(position, position + wanted.length) !== wanted) {
			return false;
		}

		position += wanted.length;

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

				position += 1;

				const found = /^[0-9]+/.exec(text.slice(position));

				if (found === null) {
					return refuse();
				}

				// A piece of nines takes as many of the digits of a second as were
				// written down and no more, while a piece of zeros asks for the
				// number of digits it carries, so a text with fewer or more of them
				// than it was asked for is refused rather than read as another second.
				if (found[0].length > 9 || (!optional && found[0].length !== width)) {
					return [null, go2jsTimeParseChunkError(layoutText, text, "." + fraction[1], "." + found[0])];
				}

				const read = found[0].slice(0, optional ? found[0].length : width);

				position += read.length;

				layoutAt += 1 + fraction[1].length;

				parts.nsec = parseInt((read + "000000000").slice(0, 9), 10);
				parts.milli = Math.floor(parts.nsec / 1000000);

				continue;
			}

			// Anything else in a layout is written out as it is, so it has to be
			// there for the text to be read as the layout says it is written.
			const ch = rest[0];

			if (ch === "Z") {
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

			if (!literal(ch)) {
				return refuse();
			}

			layoutAt += 1;

			continue;
		}

		// A piece that reads what it stands for takes it from the text as it
		// goes and so has already moved past itself, while one that is written
		// out as it is has its own length to move past.
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

			parts.hour = hour;
			break;
		}
		case "03": {
			const hour = digits(2, 2);

			if (hour === null) {
				return refuse();
			}

			parts.hour = hour;
			break;
		}
		case "04": {
			const minute = digits(2, 2);

			if (minute === null) {
				return refuse();
			}

			parts.minute = minute;
			break;
		}
		case "05": {
			const second = digits(2, 2);

			if (second === null) {
				return refuse();
			}

			parts.second = second;
			break;
		}
		case "1": {
			const month = digits(1, 2);

			if (month === null) {
				return refuse();
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

			parts.hour = hour;
			break;
		}
		case "4": {
			const minute = digits(1, 2);

			if (minute === null) {
				return refuse();
			}

			parts.minute = minute;
			break;
		}
		case "5": {
			const second = digits(1, 2);

			if (second === null) {
				return refuse();
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
			// which is where a layout that never asked for one lets it stand.
			if (zone === null && text[position] === ".") {
				if (!readFraction()) {
					return refuse();
				}

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
			return refuse();
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

	if (parts.month < 1 || parts.month > 12 || parts.day < 1 || parts.day > 31 ||
		parts.hour > 23 || parts.minute > 59 || parts.second > 59) {
		return refuse();
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

function go2jsTimeParseError(layout, value) {
	return go2jsSentinelError("parsing time " + JSON.stringify(value) + " as " +
		JSON.stringify(layout) + ": cannot parse")();
}

// An errno is a number that says what a condition was, and it is asked of
// itself what it says and whether it is the condition a sentinel names, which
// is what an error of a path carrying one is unwrapped to and matched against.
function ErrnoError(value) {
	return go2jsErrorMessage(value);
}

function ErrnoIs(value, target) {
	return go2jsErrorsIs(value, target);
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

// go2jsTimeAfter returns a channel that carries the moment it comes due. The
// channel starts empty and only becomes ready at that moment, so a select that
// has real work waiting chooses the work rather than the deadline.
// go2jsTimeStopChannel takes a timer or ticker back out of the schedule, which
// is what a Stop is. A timer that has already come due has left the schedule
// on its own, so there is nothing left to take out and Stop says it stopped
// nothing, the way Go does.
function go2jsTimeStopChannel(channel) {
	if (channel === null || channel === undefined || channel.timerDeadline === undefined) {
		return false;
	}

	go2jsTimers.delete(channel);
	channel.timerDeadline = undefined;

	return true;
}

// go2jsTimeNewTimer waits for a duration and then sends the moment on a
// channel, the way time.NewTimer does. The timer holds that channel so a Stop
// can take it back out of the schedule before it fires.
function go2jsTimeNewTimer(d) {
	const channel = go2jsTimeAfter(d);

	// A Stop says whether it was the one that kept the timer from firing, so a
	// timer that has already fired, or that was stopped before, reports that it
	// stopped nothing.
	return {C: channel, timerChannel: channel, Stop: function() {
		return go2jsTimeStopChannel(channel);
	}};
}

// go2jsTimeNewTicker sends the moment on a channel over and over, every
// duration, the way time.NewTicker does. The channel is the one the schedule
// watches, so a channel that is drained and armed again is the same channel and
// a consumer that missed a tick sees the next one rather than falling behind.
function go2jsTimeNewTicker(d) {
	const channel = go2jsChannel(1);

	channel.timerPeriod = go2jsDurationNanos(d);
	go2jsTimerArm(channel);

	return {C: channel, timerChannel: channel, Stop: function() {
		channel.timerPeriod = null;
		go2jsTimeStopChannel(channel);
	}};
}

// go2jsTimeAfterFunc runs a function once, a duration from now, and hands back a
// timer whose Stop keeps the function from running when it is called in time.
function go2jsTimeAfterFunc(d, fn) {
	const channel = go2jsTimeAfter(d);
	const timer = {timerChannel: channel, ran: false};

	timer.Stop = function() {
		if (timer.ran) {
			return false;
		}

		return go2jsTimeStopChannel(channel);
	};

	// The channel is drained by the schedule, so the function is run from there
	// rather than from a timer of its own, which keeps one clock for the whole
	// program. The timer starts it as a goroutine of its own, which is what
	// Go writes it to mean.
	channel.timerCallback = function* () {
		timer.ran = true;
		yield* go2jsCall(fn, null, []);
	};

	return timer;
}

function go2jsTimeAfter(d) {
	const channel = go2jsChannel(1);
	const due = go2jsTimeAdd(go2jsTimeNow(), d);

	channel.timerDeadline = go2jsTimeDateOf(due).getTime();
	channel.timerDue = due;
	go2jsTimers.set(channel, due);

	return channel;
}

function go2jsTimeTick(d) {
	return go2jsTimeAfter(d);
}

// go2jsRegexpClassCategories lists the one and two letter general category
// codes RE2 accepts. Everything else after \p{ is a script name, which
// JavaScript spells Script=Name.
const go2jsRegexpClassCategories = new Set([
	"L", "Lu", "Ll", "Lt", "Lm", "Lo",
	"M", "Mn", "Mc", "Me",
	"N", "Nd", "Nl", "No",
	"P", "Pc", "Pd", "Ps", "Pe", "Pi", "Pf", "Po",
	"S", "Sm", "Sc", "Sk", "So",
	"Z", "Zl", "Zp", "Zs",
	"C", "Cc", "Cf", "Co", "Cs", "Cn"
]);

function go2jsRegexpPropertyName(name) {
	if (go2jsRegexpClassCategories.has(name)) {
		return "General_Category=" + name;
	}

	return "Script=" + name;
}

// RE2 spells the ASCII classes with escapes, and JavaScript agrees on \d and
// \w but not on \s, which in JavaScript also covers Unicode spaces.
const go2jsRegexpAsciiSpace = "\\t\\n\\f\\r ";

const go2jsRegexpPosixClasses = {
	"[:alpha:]": "A-Za-z",
	"[:digit:]": "0-9",
	"[:alnum:]": "A-Za-z0-9",
	"[:upper:]": "A-Z",
	"[:lower:]": "a-z",
	"[:space:]": "\\t\\n\\f\\r ",
	"[:blank:]": "\\t ",
	"[:punct:]": "!-/:-@\\[-\\x60{-~",
	"[:print:]": "\\x20-\\x7e",
	"[:graph:]": "\\x21-\\x7e",
	"[:cntrl:]": "\\x00-\\x1f\\x7f",
	"[:xdigit:]": "0-9A-Fa-f"
};

function go2jsRegexpPosixClass(name) {
	return go2jsRegexpPosixClasses[name];
}

// go2jsRegexpASCIIClasses rewrites the classes where RE2 and JavaScript
// disagree, tracking whether the pattern is inside a bracket expression so the
// replacement can omit the brackets.
function go2jsRegexpASCIIClasses(source) {
	let out = "";
	let inClass = false;

	for (let i = 0; i < source.length; i++) {
		const ch = source[i];

		if (ch === "\\" && i + 1 < source.length) {
			const next = source[i + 1];

			if (next === "s" || next === "S") {
				const body = go2jsRegexpAsciiSpace;

				if (next === "S") {
					out += inClass ? "\\\\S" : "[^" + body + "]";
				} else {
					out += inClass ? body : "[" + body + "]";
				}

				i++;
				continue;
			}

			if (next === "d" || next === "D" || next === "w" || next === "W") {
				// The u flag already keeps \d and \w ASCII only.
				out += ch + next;
				i++;
				continue;
			}

			out += ch + next;
			i++;
			continue;
		}

		if (ch === "[") {
			// [[:alpha:]] is a bracket expression holding a POSIX class.
			if (source[i + 1] === "[" && source[i + 2] === ":") {
				const close = source.indexOf(":]" + "]", i + 3);
				const name = close === -1 ? "" : source.slice(i + 1, close + 2);
				const body = go2jsRegexpPosixClass(name);

				if (body !== undefined) {
					out += "[" + body + "]";
					i = close + 2;
					continue;
				}
			}

			inClass = true;
			out += ch;
			continue;
		}

		if (ch === "]" && inClass) {
			inClass = false;
			out += ch;
			continue;
		}

		out += ch;
	}

	return out;
}

// go2jsRegexpPattern rewrites RE2 constructs that JavaScript does not accept.
function go2jsRegexpPattern(pattern) {
	let source = go2jsStringify(pattern);
	let flags = "";

	// RE2 gives an error to the Perl tricks that JavaScript would quietly take:
	// a reference to a numbered group, or a look that goes ahead or behind.
	const refused = source.match(/\\[0-9]|\(\?P=|\(\?[=!<>|(]/);

	if (refused) {
		const offending = refused[0];

		throw new Error(offending.startsWith("\\") ? "invalid escape sequence: " + offending : "invalid or unsupported Perl syntax: " + offending);
	}

	// Leading and embedded flag groups such as (?i) become regexp flags.
	source = source.replace(/\(\?([imsU]+)\)/g, (all, group) => {
		for (const flag of group) {
			if (flag === "i" || flag === "m" || flag === "s") {
				if (!flags.includes(flag)) {
					flags += flag;
				}
			}
		}

		return "";
	});

	// (?P<name>re) is RE2 syntax for the named group (?<name>re).
	source = source.replace(/\(\?P</g, "(?<");
	source = source.replace(/\\A/g, "^").replace(/\\z/g, "$").replace(/\\Z/g, "$");
	// RE2 also accepts the single letter shorthand \pL for \p{L}.
	source = source.replace(/\\([pP])([A-Za-z])(?![A-Za-z{])/g, (all, kind, name) => "\\" + kind + "{" + name + "}");
	source = source.replace(/\\([pP])\{([^}=]+)\}/g, (all, kind, name) => "\\" + kind + "{" + go2jsRegexpPropertyName(name) + "}");
	source = go2jsRegexpASCIIClasses(source);

	return {source: source, flags: flags};
}

function go2jsRegexpNewRegExp(pattern, flags) {
	const parsed = go2jsRegexpPattern(pattern);

	if (flags === undefined || flags === null) {
		flags = "";
	}

	let merged = "u" + parsed.flags;

	for (const flag of flags) {
		if (!merged.includes(flag)) {
			merged += flag;
		}
	}

	return new RegExp(parsed.source, merged);
}

function go2jsRegexpSubmatchNames(pattern) {
	const source = go2jsRegexpPattern(pattern).source;

	// The whole match is the first name a pattern has, and it has none, and
	// every group after it contributes a name whether it was given one or not.
	const names = [""];
	let inClass = false;

	for (let index = 0; index < source.length; index++) {
		const ch = source[index];

		if (ch === "\\") {
			index++;
			continue;
		}

		if (inClass) {
			if (ch === "]") {
				inClass = false;
			}

			continue;
		}

		if (ch === "[") {
			inClass = true;
			continue;
		}

		if (ch !== "(") {
			continue;
		}

		// A group that opens with a mark is one of the forms that does not
		// capture, unless it is a mark followed by the name of a capture.
		if (source[index + 1] !== "?") {
			names.push("");
			continue;
		}

		const named = /^\(\?P?<([A-Za-z_][A-Za-z0-9_]*)>/.exec(source.slice(index));

		if (named !== null) {
			names.push(named[1]);
		}
	}

	return names;
}

// go2jsRegexpMatchString mirrors regexp.MatchString, which returns a bool and an
// error, so the shim uses the [value, error] tuple convention.
function go2jsRegexpMatchString(pattern, value) {
	return [go2jsRegexpNewRegExp(pattern).test(go2jsStringify(value)), null];
}



function go2jsRegexpFindAllString(pattern, value, limit) {
	const source = go2jsStringify(value);
	const regex = go2jsRegexpNewRegExp(pattern, "g");
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

// go2jsStringIndexMap relates Go byte offsets to JavaScript UTF-16 positions.
function go2jsStringIndexMap(text) {
	let ascii = true;

	for (let i = 0; i < text.length; i++) {
		if (text.charCodeAt(i) > 0x7f) {
			ascii = false;
			break;
		}
	}

	if (ascii) {
		return { byteLength: text.length, index: (offset) => offset };
	}

	const starts = new Map();
	let byte = 0;
	let i = 0;

	while (i < text.length) {
		const codePoint = text.codePointAt(i);

		starts.set(byte, i);
		byte += go2jsUtf8Length(codePoint);
		i += codePoint > 0xffff ? 2 : 1;
	}

	starts.set(byte, text.length);

	return {
		byteLength: byte,
		// A JavaScript string cannot hold half a UTF-8 sequence, so an offset
		// inside a rune rounds down to the start of that rune.
		index(offset) {
			let candidate = offset;

			while (candidate > 0 && !starts.has(candidate)) {
				candidate--;
			}

			return starts.get(candidate);
		}
	};
}

// go2jsStringSlice slices on Go byte offsets. Working on the decoded bytes
// keeps the result aligned with Go even when a bound lands inside a rune.
function go2jsStringSlice(value, low, high) {
	const bytes = go2jsStringToBytes(go2jsBytesToString(value));
	const start = low === undefined ? 0 : Math.trunc(low);
	const end = high === undefined ? bytes.length : Math.trunc(high);

	if (start < 0 || end < start || end > bytes.length) {
		throw new RangeError("slice bounds out of range");
	}

	return go2jsBytesToString(bytes.slice(start, end));
}

// go2jsStringIndexByte returns the byte at a Go string offset. Indexing a string
// in Go yields a byte, not a rune, so an offset inside a multi byte character
// returns that character's intermediate byte.
function go2jsStringIndexByte(value, offset) {
	const bytes = go2jsStringToBytes(go2jsBytesToString(value));
	const position = Math.trunc(offset);

	if (position < 0) {
		throw new RangeError(go2jsRuntimeErrorPrefix + "index out of range [" + position + "]");
	}

	if (position >= bytes.length) {
		throw new RangeError(go2jsRuntimeErrorPrefix + "index out of range [" + position + "] with length " + bytes.length);
	}

	return bytes[position];
}

function go2jsUtf8Length(codePoint) {
	if (codePoint < 0x80) {
		return 1;
	}

	if (codePoint < 0x800) {
		return 2;
	}

	if (codePoint < 0x10000) {
		return 3;
	}

	return 4;
}

// go2jsByteOffsetMap converts a JavaScript UTF-16 index into the UTF-8 byte
// offset Go reports. JavaScript strings are UTF-16, so any match that touches a
// non ASCII rune is offset by a different amount than Go would report.
function go2jsByteOffsetMap(text) {
	let ascii = true;

	for (let i = 0; i < text.length; i++) {
		if (text.charCodeAt(i) > 0x7f) {
			ascii = false;
			break;
		}
	}

	// Plain ASCII text needs no translation at all.
	if (ascii) {
		return (index) => index;
	}

	const offsets = new Int32Array(text.length + 1);
	let byte = 0;
	let i = 0;

	while (i < text.length) {
		const codePoint = text.codePointAt(i);
		const width = codePoint > 0xffff ? 2 : 1;

		offsets[i] = byte;
		byte += go2jsUtf8Length(codePoint);

		if (width === 2) {
			// A match can never start or end between the two halves of a
			// surrogate pair, but the slot still has to hold a value.
			offsets[i + 1] = byte;
		}

		i += width;
	}

	offsets[text.length] = byte;

	return (index) => (index >= 0 && index < offsets.length ? offsets[index] : -1);
}

function go2jsRegexpIndex(match, offset) {
	if (match === null || match === undefined) {
		return null;
	}

	if (offset === undefined) {
		offset = (index) => index;
	}

	return [offset(match.index), offset(match.index + match[0].length)];
}

function go2jsRegexpSubmatch(match) {
	// A match without capture groups still reports the whole match.
	if (match === null || match === undefined) {
		return null;
	}

	const out = [match[0]];

	for (let i = 1; i < match.length; i++) {
		out.push(match[i] === undefined ? "" : match[i]);
	}

	return out;
}

function go2jsRegexpSubmatchIndex(match, offset) {
	// Requires a regexp built with the "d" flag so capture offsets are available.
	if (match === null || match === undefined) {
		return null;
	}

	if (offset === undefined) {
		offset = (index) => index;
	}

	const out = [offset(match.index), offset(match.index + match[0].length)];
	const indices = match.indices || [];

	for (let i = 1; i < match.length; i++) {
		if (indices[i] === undefined) {
			out.push(-1, -1);
			continue;
		}

		out.push(offset(indices[i][0]), offset(indices[i][1]));
	}

	return out;
}

function go2jsRegexpByteSubmatch(match) {
	const groups = go2jsRegexpSubmatch(match);

	if (groups === null) {
		return null;
	}

	return groups.map((group) => go2jsStringToBytes(group));
}

function go2jsRegexpByteSubmatchIndex(match, offset) {
	// Byte and string variants share the offsets, because the subject was
	// decoded from the same bytes the caller passed in.
	return go2jsRegexpSubmatchIndex(match, offset);
}

function go2jsRegexpFindAllSubmatch(pattern, value, limit) {
	const source = go2jsBytesToString(value);
	const regex = go2jsRegexpNewRegExp(pattern, "g");
	const out = [];

	for (;;) {
		const match = regex.exec(source);

		if (match === null) {
			break;
		}

		out.push(go2jsRegexpSubmatch(match));

		if (limit !== undefined && limit >= 0 && out.length >= limit) {
			break;
		}
	}

	return out;
}

function go2jsRegexpFindAllSubmatchIndex(pattern, value, limit) {
	const source = go2jsBytesToString(value);
	const offset = go2jsByteOffsetMap(source);
	const regex = go2jsRegexpNewRegExp(pattern, "gd");
	const out = [];

	for (;;) {
		const match = regex.exec(source);

		if (match === null) {
			break;
		}

		out.push(go2jsRegexpSubmatchIndex(match, offset));

		if (limit !== undefined && limit >= 0 && out.length >= limit) {
			break;
		}
	}

	return out;
}

// go2jsRegexpFindAllIndex backs FindAllStringIndex and its byte counterparts.
function go2jsRegexpFindAllIndex(pattern, value, limit) {
	const source = go2jsBytesToString(value);
	const offset = go2jsByteOffsetMap(source);
	const regex = go2jsRegexpNewRegExp(pattern, "g");
	const out = [];

	for (;;) {
		const match = regex.exec(source);

		if (match === null) {
			break;
		}

		out.push(go2jsRegexpIndex(match, offset));

		if (limit !== undefined && limit >= 0 && out.length >= limit) {
			break;
		}
	}

	return out;
}

function go2jsRegexpFindAllByteSubmatch(pattern, value, limit) {
	return go2jsRegexpFindAllSubmatch(pattern, value, limit).map((groups) =>
		groups.map((group) => go2jsStringToBytes(group)));
}

function go2jsRegexpReplaceAllString(pattern, value, replacement) {
	return go2jsStringify(value).replace(
		go2jsRegexpNewRegExp(pattern, "g"),
		go2jsRegexpExpand(go2jsStringify(replacement), pattern)
	);
}

function go2jsRegexpExpand(replacement, pattern) {
	let text = go2jsStringify(replacement)
		.replace(/\$(\d+)/g, (match, index) => "$" + (Number(index) === 0 ? "&" : index))
		.replace(/\$\{(\w+)\}/g, (match, name) => "$" + (name === "0" ? "&" : name));

	// Go spells a named group as $name, while JavaScript needs $<name>.
	if (typeof pattern === "string" && pattern !== "") {
		const names = go2jsRegexpSubmatchNames(pattern);

		for (const name of names) {
			if (name === "" || /^\d+$/.test(name)) {
				continue;
			}

			text = text.split("$" + name).join("$<" + name + ">");
		}
	}

	return text;
}

function go2jsRegexpSplit(pattern, value, limit) {
	const text = go2jsBytesToString(value);

	if (limit === 0) {
		return [];
	}

	// A positive limit keeps the trailing remainder in one piece, so the split
	// is done by hand instead of with String.prototype.split.
	const regex = go2jsRegexpNewRegExp(pattern, "gd");
	const parts = [];

	let last = 0;
	let match = regex.exec(text);

	while (match !== null) {
		const start = match.indices[0][0];
		const end = match.indices[0][1];

		parts.push(text.slice(last, start));

		if (limit > 0 && parts.length === limit - 1) {
			parts.push(text.slice(end));
			return parts;
		}

		last = end;
		match = regex.exec(text);
	}

	parts.push(text.slice(last));

	return parts;
}

function go2jsRegexpQuoteMeta(value) {
	return go2jsStringify(value).replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

`
}

func osStdioRuntimeSource() string {
	return `

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
				chunk = require("fs").readSync(0, view, 0, view.length, start + filled);
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
			const chunks = [];
			const buffer = new Array(4096);

			for (;;) {
				const result = readStandardInput(buffer);

				if (result[0] <= 0) {
					break;
				}

				chunks.push(...buffer.slice(0, result[0]));
			}

			return String.fromCharCode(...chunks);
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

function go2jsOSStdin() {
	go2jsOSFileMethods();

	return process.stdin;
}

function go2jsOSStdout() {
	go2jsOSFileMethods();

	return process.stdout;
}

function go2jsOSStderr() {
	go2jsOSFileMethods();

	return process.stderr;
}
`
}

func runeRuntimeSource() string {
	return `
function go2jsRuneCodePoint(value) {
	return String(go2jsStringify(value)).codePointAt(0);
}
`
}
