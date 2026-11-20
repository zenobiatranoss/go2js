"use strict";

const {
    go2jsBufioNewScanner,
    go2jsFprint,
    go2jsIndex,
    go2jsIndexCheck,
    go2jsInterface,
    go2jsLen,
    go2jsMap,
    go2jsMapGet,
    go2jsMapTyped,
    go2jsMapUpdate,
    go2jsNilInterface,
    go2jsPreempt,
    go2jsPrintln,
    go2jsPtr,
    go2jsRangeCopies,
    go2jsRangeMap,
    go2jsRangeSequence,
    go2jsRegisterTypeName,
    go2jsReportUncaughtPanic,
    go2jsRunMain,
    go2jsRunesToString,
    go2jsSetPackageName,
    go2jsSetSourceLines,
    go2jsSliceAppend,
    go2jsSliceMake,
    go2jsSliceRange,
    go2jsSliceTyped,
    go2jsSortInts,
    go2jsSortSlice,
    go2jsSprintf,
    go2jsState,
    go2jsStrCompare,
    go2jsStringSlice,
    go2jsStringToRunes,
    go2jsStringsBuilder,
    go2jsStringsFields,
    go2jsStringsFieldsFunc,
    go2jsStringsNewReader,
    go2jsStringsRepeat,
    go2jsStringsSplit,
    go2jsStringsTitle,
    go2jsStringsToLower,
    go2jsStringsToUpper,
    go2jsStringsTrimSpace,
    go2jsTyped,
    go2jsUnicodeIsLetter,
    go2jsWideAdd,
    go2jsWideFloat,
    go2jsWideSub,
    go2jsWriteOut,
} = require("./main.runtime.js");

go2jsSetPackageName("main");
go2jsSetSourceLines({23: "04-word-stats/main.go:25:6", 24: "04-word-stats/main.go:26:28", 25: "04-word-stats/main.go:27:11", 33: "04-word-stats/main.go:31:6", 35: "04-word-stats/main.go:32:30", 36: "04-word-stats/main.go:32:30", 37: "04-word-stats/main.go:32:30", 38: "04-word-stats/main.go:33:2", 39: "04-word-stats/main.go:34:2", 40: "04-word-stats/main.go:35:2", 41: "04-word-stats/main.go:36:2", 42: "04-word-stats/main.go:37:6", 46: "04-word-stats/main.go:38:11", 47: "04-word-stats/main.go:39:6", 48: "04-word-stats/main.go:40:4", 50: "04-word-stats/main.go:42:3", 51: "04-word-stats/main.go:43:21", 52: "04-word-stats/main.go:44:4", 53: "04-word-stats/main.go:45:4", 54: "04-word-stats/main.go:46:15", 55: "04-word-stats/main.go:47:12", 57: "04-word-stats/main.go:48:16", 58: "04-word-stats/main.go:49:5", 62: "04-word-stats/main.go:54:31", 63: "04-word-stats/main.go:55:2", 64: "04-word-stats/main.go:56:10", 66: "04-word-stats/main.go:58:2", 67: "04-word-stats/main.go:59:3", 68: "04-word-stats/main.go:60:4", 70: "04-word-stats/main.go:62:3", 72: "04-word-stats/main.go:65:71", 73: "04-word-stats/main.go:65:71", 74: "04-word-stats/main.go:65:71", 75: "04-word-stats/main.go:65:71", 76: "04-word-stats/main.go:66:72", 77: "04-word-stats/main.go:66:72", 78: "04-word-stats/main.go:66:72", 79: "04-word-stats/main.go:66:72", 80: "04-word-stats/main.go:66:72", 81: "04-word-stats/main.go:66:72", 82: "04-word-stats/main.go:66:72", 83: "04-word-stats/main.go:66:72", 84: "04-word-stats/main.go:67:2", 85: "04-word-stats/main.go:68:2", 86: "04-word-stats/main.go:69:53", 87: "04-word-stats/main.go:69:53", 88: "04-word-stats/main.go:69:53", 89: "04-word-stats/main.go:69:53", 91: "04-word-stats/main.go:72:25", 92: "04-word-stats/main.go:73:2", 93: "04-word-stats/main.go:74:10", 95: "04-word-stats/main.go:76:2", 96: "04-word-stats/main.go:77:2", 97: "04-word-stats/main.go:78:2", 98: "04-word-stats/main.go:79:38", 99: "04-word-stats/main.go:79:38", 100: "04-word-stats/main.go:79:38", 101: "04-word-stats/main.go:79:38", 102: "04-word-stats/main.go:79:38", 103: "04-word-stats/main.go:79:38", 104: "04-word-stats/main.go:79:38", 105: "04-word-stats/main.go:79:38", 106: "04-word-stats/main.go:79:38", 108: "04-word-stats/main.go:82:25", 109: "04-word-stats/main.go:83:74", 110: "04-word-stats/main.go:83:74", 111: "04-word-stats/main.go:83:74", 112: "04-word-stats/main.go:83:74", 113: "04-word-stats/main.go:83:74", 114: "04-word-stats/main.go:83:74", 115: "04-word-stats/main.go:84:9", 116: "04-word-stats/main.go:85:17", 117: "04-word-stats/main.go:86:3", 121: "04-word-stats/main.go:88:61", 122: "04-word-stats/main.go:88:61", 123: "04-word-stats/main.go:88:61", 124: "04-word-stats/main.go:88:61", 125: "04-word-stats/main.go:88:61", 126: "04-word-stats/main.go:89:2", 127: "04-word-stats/main.go:90:2", 128: "04-word-stats/main.go:91:3", 129: "04-word-stats/main.go:92:4", 131: "04-word-stats/main.go:94:32", 132: "04-word-stats/main.go:94:32", 133: "04-word-stats/main.go:94:32", 134: "04-word-stats/main.go:94:32", 135: "04-word-stats/main.go:94:32", 136: "04-word-stats/main.go:94:32", 137: "04-word-stats/main.go:94:32", 138: "04-word-stats/main.go:94:32", 139: "04-word-stats/main.go:94:32", 140: "04-word-stats/main.go:94:32", 141: "04-word-stats/main.go:94:32", 142: "04-word-stats/main.go:94:32", 143: "04-word-stats/main.go:94:32", 144: "04-word-stats/main.go:94:32", 145: "04-word-stats/main.go:94:32", 146: "04-word-stats/main.go:94:32", 148: "04-word-stats/main.go:96:14"});
const text = "It was the best of times, it was the worst of times, it was the age of wisdom,\nit was the age of foolishness, it was the epoch of belief, it was the epoch of incredulity,\nit was the season of Light, it was the season of Darkness, it was the spring of hope,\nit was the winter of despair.\n\nWe had everything before us, we had nothing before us; we were all going direct to Heaven,\nwe were all going direct the other way.";

class WordCount {
    Word;
    Count;
    constructor() {
        this.Word = "";
        this.Count = 0;
    }
}

go2jsRegisterTypeName(
    WordCount,
    "main.WordCount",
    [],
    [
        {Name: "Word", Tag: "", PkgPath: "", Anonymous: false, Type: "string", Kind: "string"},
        {Name: "Count", Tag: "", PkgPath: "", Anonymous: false, Type: "int", Kind: "int"}
    ],
    "struct"
);

function tokenize$(line) {
    return go2jsStringsFieldsFunc(go2jsStringsToLower(line), function*(r) {
        return !go2jsUnicodeIsLetter(r) && r != 39;
    });
}

function* tokenize(line) {
    return tokenize$(line);
}

function* main() {
    let addr1 = null;
    let scanner = go2jsBufioNewScanner(
        go2jsInterface(go2jsStringsNewReader(text), "*Reader", "*strings.Reader")
    );
    let counts = go2jsMapTyped("map[string]int", go2jsMap([]));
    let [lines, words, letters] = [0, 0, 0];
    let longest = "";
    let lengths = go2jsMapTyped("map[int]int", go2jsMap([]));
    for (; scanner.Scan();) {
        if (++go2jsState.go2jsTicks > 20000) {
            yield* go2jsPreempt();
        }
        let line = scanner.Text();
        if (go2jsStringsTrimSpace(line) == "") {
            continue;
        }
        lines = go2jsWideAdd(lines, 1, "int");
        for (const w of go2jsRangeSequence(tokenize$(line))) {
            go2jsMapUpdate(counts, w, 0, (current) => go2jsWideAdd(current, 1, "int"));
            words = go2jsWideAdd(words, 1, "int");
            letters = go2jsWideAdd(letters, go2jsLen(w), "int");
            const lhs2 = go2jsLen(w);
            go2jsMapUpdate(lengths, lhs2, 0, (current) => go2jsWideAdd(current, 1, "int"));
            if (go2jsLen(w) > go2jsLen(longest)) {
                longest = w;
            }
        }
    }
    let list = go2jsSliceMake(0, go2jsLen(counts), () => new WordCount(), false);
    for (const [w, c] of go2jsRangeSequence(counts).entries()) {
        list = go2jsSliceAppend(list, Object.assign(new WordCount(), {Word: w, Count: c}));
    }
    go2jsSortSlice(go2jsNilInterface(list, "[]main.WordCount", "slice", true), function*(i, j) {
        if (go2jsIndex(list, i).Count != go2jsIndex(list, j).Count) {
            return go2jsIndex(list, i).Count > go2jsIndex(list, j).Count;
        }
        return (go2jsStrCompare(go2jsIndex(list, i).Word, go2jsIndex(list, j).Word) < 0);
    });
    go2jsWriteOut(
        process.stdout,
        go2jsSprintf("lines=%d words=%d unique=%d letters=%d\n", lines, words, go2jsLen(counts), letters)
    );
    go2jsWriteOut(
        process.stdout,
        go2jsSprintf(
            "average word length %.2f, longest %q\n",
            go2jsTyped(go2jsWideFloat(letters) / go2jsWideFloat(words), "float64"),
            longest
        )
    );
    go2jsPrintln("top words:");
    for (const wc of go2jsRangeCopies(go2jsRangeSequence(go2jsSliceRange(list, undefined, 8, undefined, false)))) {
        go2jsWriteOut(
            process.stdout,
            go2jsSprintf("  %-10s %3d %s\n", wc.Word, wc.Count, go2jsStringsRepeat("#", wc.Count))
        );
    }
    let keys = go2jsSliceMake(0, go2jsLen(lengths), 0, true);
    for (const k of go2jsRangeMap(lengths).keys()) {
        keys = go2jsSliceAppend(keys, k);
    }
    go2jsSortInts(keys);
    go2jsPrintln("length histogram:");
    for (const k of go2jsRangeSequence(keys)) {
        go2jsWriteOut(
            process.stdout,
            go2jsSprintf(
                "  %2d | %s (%d)\n",
                k,
                go2jsStringsRepeat("*", go2jsMapGet(lengths, k, 0)),
                go2jsMapGet(lengths, k, 0)
            )
        );
    }
    let title = go2jsStringsTitle(go2jsStringsToLower("the quick brown fox"));
    go2jsPrintln(
        title,
        go2jsStringsToUpper("héllo wörld"),
        go2jsLen("héllo"),
        go2jsLen(go2jsStringToRunes("héllo"))
    );
    let rev = go2jsStringToRunes("stressed");
    for (let [i, j] = [0, go2jsWideSub(go2jsLen(rev), 1, "int")]; i < j; [i, j] = [go2jsWideAdd(i, 1, "int"), go2jsWideSub(j, 1, "int")]) {
        go2jsIndexCheck(rev, i);
        go2jsIndexCheck(rev, j);
        [rev[i], rev[j]] = [go2jsIndex(rev, j), go2jsIndex(rev, i)];
    }
    go2jsPrintln(
        go2jsRunesToString(rev),
        go2jsIndex(go2jsStringsFields("  a  b   c "), 2),
        go2jsTyped(go2jsStringsSplit("a,b,,c", ","), "[]string", "", "slice")
    );
    let sb = new go2jsStringsBuilder();
    for (const [i, w] of go2jsRangeSequence(go2jsSliceTyped("[]string", ["alpha", "beta", "gamma"])).entries()) {
        if (i > 0) {
            sb.WriteString(" | ");
        }
        go2jsFprint(
            (addr1 ??= go2jsPtr(() => sb, value => sb = value)),
            [
                go2jsSprintf(
                    "%d:%s",
                    i,
                    go2jsStringsToUpper(go2jsStringSlice(w, undefined, 1)) + go2jsStringSlice(
                        w,
                        1,
                        undefined
                    )
                )
            ],
            "",
            ""
        );
    }
    go2jsPrintln(sb.String());
}

try {
    go2jsRunMain((function*() {
        return yield* main();
    })());
} catch (e) {
    go2jsReportUncaughtPanic(e);
}
