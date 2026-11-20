"use strict";

const {
    go2jsCall,
    go2jsErrorsNew,
    go2jsIndex,
    go2jsLen,
    go2jsMap,
    go2jsMapGet,
    go2jsMapGetOK,
    go2jsMapSet,
    go2jsMapTyped,
    go2jsMethodOnYield,
    go2jsNamedMethodCall,
    go2jsNew,
    go2jsPreempt,
    go2jsPrint,
    go2jsPrintln,
    go2jsRangeMap,
    go2jsRangeSequence,
    go2jsRegisterMethod,
    go2jsRegisterTypeName,
    go2jsRegisterValueMethods,
    go2jsReportUncaughtPanic,
    go2jsRunMain,
    go2jsSetPackageName,
    go2jsSetSourceLines,
    go2jsSliceAppend,
    go2jsSliceMake,
    go2jsSliceRange,
    go2jsSliceStore,
    go2jsSliceTyped,
    go2jsSlicesContains,
    go2jsSlicesIndex,
    go2jsSlicesMax,
    go2jsSlicesSort,
    go2jsSortStrings,
    go2jsSprint,
    go2jsSprintf,
    go2jsState,
    go2jsStringsFields,
    go2jsStringsToUpper,
    go2jsToNumber,
    go2jsTyped,
    go2jsWideMul,
    go2jsWideSub,
    go2jsWriteOut,
    go2jsZero,
} = require("./main.runtime.js");

go2jsSetPackageName("main");
go2jsSetSourceLines({2: "05-generics/main.go:17:6", 3: "05-generics/main.go:18:2", 4: "05-generics/main.go:19:2", 5: "05-generics/main.go:20:3", 7: "05-generics/main.go:22:2", 10: "05-generics/main.go:25:6", 11: "05-generics/main.go:26:22", 12: "05-generics/main.go:27:2", 13: "05-generics/main.go:28:21", 15: "05-generics/main.go:30:2", 18: "05-generics/main.go:33:6", 19: "05-generics/main.go:34:2", 20: "05-generics/main.go:35:2", 21: "05-generics/main.go:36:6", 22: "05-generics/main.go:37:10", 25: "05-generics/main.go:40:2", 28: "05-generics/main.go:43:6", 29: "05-generics/main.go:44:2", 30: "05-generics/main.go:45:2", 31: "05-generics/main.go:46:9", 33: "05-generics/main.go:48:2", 36: "05-generics/main.go:51:6", 37: "05-generics/main.go:52:2", 38: "05-generics/main.go:53:5", 39: "05-generics/main.go:54:3", 41: "05-generics/main.go:56:2", 42: "05-generics/main.go:57:2", 43: "05-generics/main.go:58:15", 44: "05-generics/main.go:59:4", 47: "05-generics/main.go:62:2", 50: "05-generics/main.go:65:6", 51: "05-generics/main.go:66:2", 52: "05-generics/main.go:67:2", 53: "05-generics/main.go:68:8", 54: "05-generics/main.go:69:12", 56: "05-generics/main.go:71:2", 75: "05-generics/main.go:76:20", 76: "05-generics/main.go:76:42", 82: "05-generics/main.go:77:20", 83: "05-generics/main.go:78:2", 84: "05-generics/main.go:79:5", 85: "05-generics/main.go:80:16", 87: "05-generics/main.go:82:15", 88: "05-generics/main.go:83:21", 89: "05-generics/main.go:83:21", 90: "05-generics/main.go:83:21", 91: "05-generics/main.go:83:21", 92: "05-generics/main.go:83:21", 93: "05-generics/main.go:83:21", 94: "05-generics/main.go:83:21", 95: "05-generics/main.go:84:2", 101: "05-generics/main.go:86:19", 102: "05-generics/main.go:86:38", 129: "05-generics/main.go:93:21", 130: "05-generics/main.go:93:46", 152: "05-generics/main.go:97:6", 153: "05-generics/main.go:98:2", 156: "05-generics/main.go:100:28", 157: "05-generics/main.go:100:44", 163: "05-generics/main.go:101:28", 164: "05-generics/main.go:102:23", 165: "05-generics/main.go:103:2", 166: "05-generics/main.go:104:10", 168: "05-generics/main.go:106:2", 169: "05-generics/main.go:107:28", 170: "05-generics/main.go:108:2", 171: "05-generics/main.go:109:3", 173: "05-generics/main.go:111:2", 179: "05-generics/main.go:116:17", 180: "05-generics/main.go:116:33", 181: "05-generics/main.go:116:33", 182: "05-generics/main.go:116:33", 183: "05-generics/main.go:116:33", 184: "05-generics/main.go:116:33", 185: "05-generics/main.go:116:33", 190: "05-generics/main.go:117:17", 191: "05-generics/main.go:117:33", 192: "05-generics/main.go:117:33", 193: "05-generics/main.go:117:33", 194: "05-generics/main.go:117:33", 195: "05-generics/main.go:117:33", 196: "05-generics/main.go:117:33", 197: "05-generics/main.go:117:48", 202: "05-generics/main.go:118:6", 203: "05-generics/main.go:119:2", 204: "05-generics/main.go:120:2", 205: "05-generics/main.go:121:3", 207: "05-generics/main.go:123:2", 208: "05-generics/main.go:124:3", 210: "05-generics/main.go:126:2", 236: "05-generics/main.go:136:6", 237: "05-generics/main.go:137:2", 238: "05-generics/main.go:138:52", 239: "05-generics/main.go:138:52", 240: "05-generics/main.go:138:52", 241: "05-generics/main.go:138:52", 242: "05-generics/main.go:138:52", 243: "05-generics/main.go:138:52", 244: "05-generics/main.go:138:52", 245: "05-generics/main.go:138:52", 246: "05-generics/main.go:138:52", 247: "05-generics/main.go:139:13", 248: "05-generics/main.go:139:41", 250: "05-generics/main.go:140:23", 251: "05-generics/main.go:140:23", 252: "05-generics/main.go:140:23", 253: "05-generics/main.go:140:58", 255: "05-generics/main.go:141:14", 256: "05-generics/main.go:141:73", 258: "05-generics/main.go:142:14", 259: "05-generics/main.go:142:14", 260: "05-generics/main.go:142:14", 261: "05-generics/main.go:142:14", 262: "05-generics/main.go:142:14", 263: "05-generics/main.go:142:14", 264: "05-generics/main.go:142:14", 265: "05-generics/main.go:142:14", 266: "05-generics/main.go:142:14", 267: "05-generics/main.go:142:14", 268: "05-generics/main.go:142:14", 269: "05-generics/main.go:142:14", 270: "05-generics/main.go:142:14", 271: "05-generics/main.go:144:2", 272: "05-generics/main.go:144:2", 273: "05-generics/main.go:144:2", 274: "05-generics/main.go:144:2", 275: "05-generics/main.go:144:2", 276: "05-generics/main.go:144:2", 277: "05-generics/main.go:144:2", 278: "05-generics/main.go:144:2", 279: "05-generics/main.go:144:2", 280: "05-generics/main.go:144:2", 281: "05-generics/main.go:145:15", 282: "05-generics/main.go:145:50", 284: "05-generics/main.go:146:2", 285: "05-generics/main.go:147:11", 286: "05-generics/main.go:147:50", 288: "05-generics/main.go:148:2", 289: "05-generics/main.go:149:12", 290: "05-generics/main.go:149:52", 292: "05-generics/main.go:150:12", 293: "05-generics/main.go:151:2", 294: "05-generics/main.go:152:12", 296: "05-generics/main.go:154:2", 297: "05-generics/main.go:155:2", 298: "05-generics/main.go:156:12", 299: "05-generics/main.go:156:51", 301: "05-generics/main.go:157:62", 302: "05-generics/main.go:157:62", 303: "05-generics/main.go:157:62", 304: "05-generics/main.go:157:62", 305: "05-generics/main.go:157:62", 306: "05-generics/main.go:157:62", 307: "05-generics/main.go:157:62", 308: "05-generics/main.go:157:62", 309: "05-generics/main.go:157:62", 310: "05-generics/main.go:157:62", 311: "05-generics/main.go:157:98", 314: "05-generics/main.go:160:2", 315: "05-generics/main.go:161:20", 316: "05-generics/main.go:162:3", 318: "05-generics/main.go:164:6", 322: "05-generics/main.go:165:11", 323: "05-generics/main.go:166:3", 325: "05-generics/main.go:168:12", 326: "05-generics/main.go:169:2", 327: "05-generics/main.go:171:8", 328: "05-generics/main.go:172:2", 329: "05-generics/main.go:173:2", 330: "05-generics/main.go:174:2", 331: "05-generics/main.go:175:14", 332: "05-generics/main.go:175:14", 333: "05-generics/main.go:175:14", 334: "05-generics/main.go:177:2", 335: "05-generics/main.go:177:2", 336: "05-generics/main.go:177:2", 337: "05-generics/main.go:177:2", 338: "05-generics/main.go:178:2", 339: "05-generics/main.go:179:2", 340: "05-generics/main.go:180:2", 341: "05-generics/main.go:181:2", 342: "05-generics/main.go:182:7", 343: "05-generics/main.go:183:32", 344: "05-generics/main.go:183:32", 345: "05-generics/main.go:183:32", 346: "05-generics/main.go:183:32", 347: "05-generics/main.go:183:32", 348: "05-generics/main.go:184:55"});

function* Sum(__go2js_type0, xs) {
    let total = go2jsZero(__go2js_type0);
    for (const x of go2jsRangeSequence(xs)) {
        total += x;
    }
    return total;
}

function* Map$go2js(__go2js_type0, __go2js_type1, xs, f) {
    let out = go2jsSliceMake(0, go2jsLen(xs), null, true);
    for (const x of go2jsRangeSequence(xs)) {
        out = go2jsSliceAppend(out, yield* go2jsCall(f, null, [x]));
    }
    return out;
}

function* Filter(__go2js_type0, xs, keep) {
    let out = null;
    for (const x of go2jsRangeSequence(xs)) {
        if (yield* go2jsCall(keep, null, [x])) {
            out = go2jsSliceAppend(out, x);
        }
    }
    return out;
}

function* Reduce(__go2js_type0, __go2js_type1, xs, init, f) {
    let acc = init;
    for (const x of go2jsRangeSequence(xs)) {
        acc = yield* go2jsCall(f, null, [acc, x]);
    }
    return acc;
}

function* MaxBy(__go2js_type0, __go2js_type1, xs, key) {
    let best = go2jsZero(__go2js_type0);
    if (go2jsLen(xs) == 0) {
        return [best, false];
    }
    best = go2jsIndex(xs, 0);
    for (const x of go2jsRangeSequence(go2jsSliceRange(xs, 1, undefined, undefined, false))) {
        if ((yield* go2jsCall(key, null, [x])) > (yield* go2jsCall(key, null, [best]))) {
            best = x;
        }
    }
    return [best, true];
}

function* GroupBy(__go2js_type0, __go2js_type1, xs, key) {
    let out = go2jsMapTyped("map[K][]T", go2jsMap([]));
    for (const x of go2jsRangeSequence(xs)) {
        let k = yield* go2jsCall(key, null, [x]);
        go2jsMapSet(out, k, go2jsSliceAppend(go2jsMapGet(out, k, null), x));
    }
    return out;
}

class Stack {
    items;
    constructor() {
        this.items = null;
    }
}

go2jsRegisterTypeName(
    Stack,
    "main.Stack",
    ["Len", "Pop", "Push"],
    [{Name: "items", Tag: "", PkgPath: "main", Anonymous: false, Type: "[]T", Kind: "slice"}],
    "struct"
);
go2jsRegisterValueMethods("main.Stack", ["Len"]);

Stack.prototype.Push = function*(v) {
    this.items = go2jsSliceAppend(this.items, v);
}
go2jsRegisterMethod("*Stack.Push", function*(receiver, ...args) {
    return yield* Stack.prototype.Push.call(receiver, ...args);
}, 0);

Stack.prototype.Pop = function*() {
    let zero = go2jsZero("T");
    if (go2jsLen(this.items) == 0) {
        return [zero, go2jsErrorsNew("empty stack")];
    }
    let v = go2jsIndex(this.items, go2jsWideSub(go2jsLen(this.items), 1, "int"));
    this.items = go2jsSliceRange(
        this.items,
        undefined,
        go2jsWideSub(go2jsLen(this.items), 1, "int"),
        undefined,
        false
    );
    return [v, null];
}
go2jsRegisterMethod("*Stack.Pop", function*(receiver, ...args) {
    return yield* Stack.prototype.Pop.call(receiver, ...args);
}, 2);

Stack.prototype.Len = function*() {
    return go2jsLen(this.items);
}
go2jsRegisterMethod("Stack.Len", function*(receiver, ...args) {
    return yield* Stack.prototype.Len.call(receiver, ...args);
}, 1);

class Pair {
    Key;
    Val;
    constructor() {
        this.Key = null;
        this.Val = null;
    }
}

go2jsRegisterTypeName(
    Pair,
    "main.Pair",
    ["String"],
    [
        {Name: "Key", Tag: "", PkgPath: "", Anonymous: false, Type: "K", Kind: "interface"},
        {Name: "Val", Tag: "", PkgPath: "", Anonymous: false, Type: "V", Kind: "interface"}
    ],
    "struct"
);
go2jsRegisterValueMethods("main.Pair", ["String"]);

Pair.prototype.String = function*() {
    return go2jsSprintf("%v=%v", go2jsTyped(this.Key, "K"), go2jsTyped(this.Val, "V"));
}
go2jsRegisterMethod("Pair.String", function*(receiver, ...args) {
    return yield* Pair.prototype.String.call(receiver, ...args);
}, 1);

class OrderedMap {
    m;
    constructor() {
        this.m = null;
    }
}

go2jsRegisterTypeName(
    OrderedMap,
    "main.OrderedMap",
    ["Pairs", "Set"],
    [{Name: "m", Tag: "", PkgPath: "main", Anonymous: false, Type: "map[K]V", Kind: "map"}],
    "struct"
);
go2jsRegisterValueMethods("main.OrderedMap", []);

function* NewOrderedMap(__go2js_type0, __go2js_type1) {
    return go2jsNew(Object.assign(new OrderedMap(), {m: go2jsMapTyped("map[K]V", go2jsMap([]))}));
}

OrderedMap.prototype.Set = function*(k, v) {
    go2jsMapSet(this.m, k, v);
}
go2jsRegisterMethod("*OrderedMap.Set", function*(receiver, ...args) {
    return yield* OrderedMap.prototype.Set.call(receiver, ...args);
}, 0);

OrderedMap.prototype.Pairs = function*() {
    let keys = go2jsSliceMake(0, go2jsLen(this.m), null, true);
    for (const k of go2jsRangeMap(this.m).keys()) {
        keys = go2jsSliceAppend(keys, k);
    }
    go2jsSlicesSort(keys);
    let out = go2jsSliceMake(go2jsLen(keys), go2jsLen(keys), () => new Pair(), false);
    for (const [i, k] of go2jsRangeSequence(keys).entries()) {
        go2jsSliceStore(out, i, Object.assign(new Pair(), {Key: k, Val: go2jsMapGet(this.m, k, null)}));
    }
    return out;
}
go2jsRegisterMethod("*OrderedMap.Pairs", function*(receiver, ...args) {
    return yield* OrderedMap.prototype.Pairs.call(receiver, ...args);
}, 1);

function* go2jsMethodSet$go2jsAdd(s, v) {
    go2jsMapSet(s, v, (() => {
        const __v = {};
        Object.defineProperty(__v, "__go2js_type_name", {value: "struct {}"});
        Object.defineProperty(__v, "__go2js_struct_fields", {value: []});
        return __v;
    })());
}

go2jsRegisterMethod("Set.Add", go2jsMethodSet$go2jsAdd);

function* go2jsMethodSet$go2jsHas(s, v) {
    let [, ok] = go2jsMapGetOK(s, v, (() => {
        const __v = {};
        Object.defineProperty(__v, "__go2js_type_name", {value: "struct {}"});
        Object.defineProperty(__v, "__go2js_struct_fields", {value: []});
        return __v;
    })());
    return ok;
}

go2jsRegisterMethod("Set.Has", go2jsMethodSet$go2jsHas);

function* Union(__go2js_type0, a, b) {
    let out = go2jsMapTyped("main.Set[T]", go2jsMap([]));
    for (const k of go2jsRangeMap(a).keys()) {
        go2jsNamedMethodCall(out, "Set", "Add", k);
    }
    for (const k of go2jsRangeMap(b).keys()) {
        go2jsNamedMethodCall(out, "Set", "Add", k);
    }
    return out;
}

class Person {
    Name;
    Age;
    City;
    constructor() {
        this.Name = "";
        this.Age = 0;
        this.City = "";
    }
}

go2jsRegisterTypeName(
    Person,
    "main.Person",
    [],
    [
        {Name: "Name", Tag: "", PkgPath: "", Anonymous: false, Type: "string", Kind: "string"},
        {Name: "Age", Tag: "", PkgPath: "", Anonymous: false, Type: "int", Kind: "int"},
        {Name: "City", Tag: "", PkgPath: "", Anonymous: false, Type: "string", Kind: "string"}
    ],
    "struct"
);

function* main() {
    let ints = go2jsSliceTyped("[]int", [5, 3, 8, 1, 9, 2]);
    go2jsPrintln(
        yield* Sum("int", ints),
        go2jsTyped(yield* Sum("float64", go2jsSliceTyped("[]float64", [1.5, 2.25])), "float64"),
        go2jsTyped(
            yield* Sum("float64", go2jsSliceTyped("[]main.Celsius", [20.5, 1.5])),
            "main.Celsius",
            "float64"
        )
    );
    let squares = yield* Map$go2js("int", "int", ints, function*(n) {
        return go2jsWideMul(n, n, "int");
    });
    go2jsPrintln(go2jsTyped(squares, "[]int", "", "slice"), go2jsTyped(yield* Filter("int", squares, function*(
        n
    ) {
        return n > 10;
    }), "[]int", "", "slice"));
    go2jsPrintln(yield* Reduce("int", "string", ints, "", function*(a, n) {
        return a + go2jsSprint(n);
    }));
    go2jsPrintln(
        go2jsTyped(
            yield* Map$go2js(
                "string",
                "string",
                go2jsSliceTyped("[]string", ["go", "js"]),
                go2jsStringsToUpper
            ),
            "[]string",
            "",
            "slice"
        )
    );
    let people = go2jsSliceTyped(
        "[]main.Person",
        [
            Object.assign(new Person(), {Name: "Ann", Age: 31, City: "Oslo"}),
            Object.assign(new Person(), {Name: "Bob", Age: 25, City: "Rome"}),
            Object.assign(new Person(), {Name: "Cid", Age: 47, City: "Oslo"}),
            Object.assign(new Person(), {Name: "Dee", Age: 38, City: "Rome"}),
            Object.assign(new Person(), {Name: "Eve", Age: 29, City: "Kyiv"})
        ]
    );
    let [oldest] = yield* MaxBy("struct", "int", people, function*(p) {
        return p.Age;
    });
    go2jsPrintln("oldest:", go2jsTyped(oldest, "main.Person"));
    let [, ok] = yield* MaxBy("struct", "int", go2jsSliceTyped("[]main.Person", []), function*(p) {
        return p.Age;
    });
    go2jsPrintln("empty:", ok);
    let groups = yield* GroupBy("struct", "string", people, function*(p) {
        return p.City;
    });
    let cities = go2jsSliceMake(0, 0, "", true);
    for (const c of go2jsRangeMap(groups).keys()) {
        cities = go2jsSliceAppend(cities, c);
    }
    go2jsSortStrings(cities);
    for (const c of go2jsRangeSequence(cities)) {
        let names = yield* Map$go2js("struct", "string", go2jsMapGet(groups, c, null), function*(p) {
            return p.Name;
        });
        go2jsWriteOut(process.stdout, go2jsSprintf("%-5s %v avg age %.1f\n", c, go2jsTyped(
            names,
            "[]string",
            "",
            "slice"
        ), go2jsTyped(go2jsToNumber(yield* Sum("int", yield* Map$go2js("struct", "int", go2jsMapGet(
            groups,
            c,
            null
        ), function*(p) {
            return p.Age;
        }))) / go2jsToNumber(go2jsLen(go2jsMapGet(groups, c, null))), "float64")));
    }
    let st = new Stack();
    for (const w of go2jsRangeSequence(go2jsStringsFields("one two three"))) {
        yield* st.Push(w);
    }
    for (; (yield* st.Len()) > 0;) {
        if (++go2jsState.go2jsTicks > 20000) {
            yield* go2jsPreempt();
        }
        let [v] = yield* st.Pop();
        go2jsPrint(v, " ");
    }
    let [, err] = yield* st.Pop();
    go2jsPrintln(err);
    let om = yield* go2jsCall(NewOrderedMap, null, ["string", "int"]);
    yield* go2jsMethodOnYield(om, OrderedMap, "Set", ["pear", 3]);
    yield* go2jsMethodOnYield(om, OrderedMap, "Set", ["apple", 7]);
    yield* go2jsMethodOnYield(om, OrderedMap, "Set", ["fig", 1]);
    go2jsPrintln(
        go2jsTyped(yield* go2jsMethodOnYield(om, OrderedMap, "Pairs", []), "[]main.Pair", "", "slice")
    );
    let [a, b] = [
        go2jsMapTyped("main.Set[int]", go2jsMap([])),
        go2jsMapTyped("main.Set[int]", go2jsMap([]))
    ];
    go2jsNamedMethodCall(a, "Set", "Add", 1);
    go2jsNamedMethodCall(a, "Set", "Add", 2);
    go2jsNamedMethodCall(b, "Set", "Add", 2);
    go2jsNamedMethodCall(b, "Set", "Add", 3);
    let u = yield* Union("int", a, b);
    go2jsPrintln(
        go2jsLen(u),
        go2jsNamedMethodCall(u, "Set", "Has", 3),
        go2jsNamedMethodCall(u, "Set", "Has", 4)
    );
    go2jsPrintln(go2jsSlicesMax(ints), go2jsSlicesIndex(ints, 8), go2jsSlicesContains(ints, 7));
}

try {
    go2jsRunMain((function*() {
        return yield* main();
    })());
} catch (e) {
    go2jsReportUncaughtPanic(e);
}
