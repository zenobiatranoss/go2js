"use strict";

const {
    go2jsAssert,
    go2jsBytesToString,
    go2jsCopy,
    go2jsImul,
    go2jsIndex,
    go2jsIntWrap,
    go2jsIsNil,
    go2jsLen,
    go2jsListNew,
    go2jsMap,
    go2jsMapDelete,
    go2jsMapGetOK,
    go2jsMapSet,
    go2jsMapTyped,
    go2jsMethodOn,
    go2jsMod,
    go2jsNamedMethodCall,
    go2jsNew,
    go2jsPrintln,
    go2jsRangeCopies,
    go2jsRangeMap,
    go2jsRangeSequence,
    go2jsRangeValue,
    go2jsRegisterMethod,
    go2jsRegisterTypeName,
    go2jsRegisterValueMethods,
    go2jsReportUncaughtPanic,
    go2jsRunMain,
    go2jsSetPackageName,
    go2jsSetSourceLines,
    go2jsShr,
    go2jsSliceAppend,
    go2jsSliceMake,
    go2jsSliceRange,
    go2jsSliceStore,
    go2jsSliceTyped,
    go2jsSprintf,
    go2jsStringToBytes,
    go2jsStringsBuilder,
    go2jsStringsRepeat,
    go2jsStructCopy,
    go2jsTyped,
    go2jsTypedNilGuard,
    go2jsWideAdd,
    go2jsWideMul,
    go2jsWideSub,
    go2jsWideToUnsigned,
    go2jsWideWrap,
    go2jsWriteOut,
} = require("./main.runtime.js");

go2jsSetPackageName("main");
go2jsSetSourceLines({66: "10-lru-and-life/main.go:22:6", 67: "10-lru-and-life/main.go:23:29", 68: "10-lru-and-life/main.go:23:29", 69: "10-lru-and-life/main.go:23:29", 70: "10-lru-and-life/main.go:23:29", 71: "10-lru-and-life/main.go:23:29", 72: "10-lru-and-life/main.go:23:29", 73: "10-lru-and-life/main.go:23:29", 74: "10-lru-and-life/main.go:23:29", 75: "10-lru-and-life/main.go:23:29", 76: "10-lru-and-life/main.go:23:29", 83: "10-lru-and-life/main.go:26:15", 84: "10-lru-and-life/main.go:27:2", 87: "10-lru-and-life/main.go:28:3", 88: "10-lru-and-life/main.go:29:3", 89: "10-lru-and-life/main.go:30:3", 92: "10-lru-and-life/main.go:32:2", 93: "10-lru-and-life/main.go:33:2", 102: "10-lru-and-life/main.go:36:15", 104: "10-lru-and-life/main.go:37:2", 107: "10-lru-and-life/main.go:38:3", 109: "10-lru-and-life/main.go:39:3", 110: "10-lru-and-life/main.go:40:3", 113: "10-lru-and-life/main.go:42:15", 114: "10-lru-and-life/main.go:42:15", 115: "10-lru-and-life/main.go:42:15", 116: "10-lru-and-life/main.go:42:15", 117: "10-lru-and-life/main.go:42:15", 118: "10-lru-and-life/main.go:42:15", 119: "10-lru-and-life/main.go:42:15", 120: "10-lru-and-life/main.go:43:5", 121: "10-lru-and-life/main.go:44:11", 122: "10-lru-and-life/main.go:45:3", 123: "10-lru-and-life/main.go:46:3", 124: "10-lru-and-life/main.go:47:3", 126: "10-lru-and-life/main.go:49:2", 135: "10-lru-and-life/main.go:52:15", 136: "10-lru-and-life/main.go:53:2", 137: "10-lru-and-life/main.go:54:42", 138: "10-lru-and-life/main.go:55:9", 140: "10-lru-and-life/main.go:57:2", 149: "10-lru-and-life/main.go:62:6", 150: "10-lru-and-life/main.go:63:18", 151: "10-lru-and-life/main.go:64:2", 152: "10-lru-and-life/main.go:65:23", 153: "10-lru-and-life/main.go:66:3", 156: "10-lru-and-life/main.go:67:4", 159: "10-lru-and-life/main.go:70:2", 166: "10-lru-and-life/main.go:73:15", 167: "10-lru-and-life/main.go:74:2", 168: "10-lru-and-life/main.go:75:18", 169: "10-lru-and-life/main.go:76:2", 170: "10-lru-and-life/main.go:77:3", 171: "10-lru-and-life/main.go:78:4", 172: "10-lru-and-life/main.go:79:5", 176: "10-lru-and-life/main.go:83:2", 181: "10-lru-and-life/main.go:86:15", 182: "10-lru-and-life/main.go:87:21", 183: "10-lru-and-life/main.go:88:2", 184: "10-lru-and-life/main.go:89:26", 185: "10-lru-and-life/main.go:89:26", 186: "10-lru-and-life/main.go:89:26", 187: "10-lru-and-life/main.go:89:26", 188: "10-lru-and-life/main.go:89:26", 189: "10-lru-and-life/main.go:90:3", 190: "10-lru-and-life/main.go:91:9", 191: "10-lru-and-life/main.go:92:4", 194: "10-lru-and-life/main.go:95:2", 199: "10-lru-and-life/main.go:98:15", 200: "10-lru-and-life/main.go:99:2", 201: "10-lru-and-life/main.go:100:2", 202: "10-lru-and-life/main.go:101:3", 203: "10-lru-and-life/main.go:102:4", 204: "10-lru-and-life/main.go:103:5", 206: "10-lru-and-life/main.go:105:5", 209: "10-lru-and-life/main.go:108:3", 211: "10-lru-and-life/main.go:110:9", 232: "10-lru-and-life/main.go:115:15", 233: "10-lru-and-life/main.go:116:2", 234: "10-lru-and-life/main.go:117:2", 243: "10-lru-and-life/main.go:119:15", 244: "10-lru-and-life/main.go:119:55", 245: "10-lru-and-life/main.go:119:55", 246: "10-lru-and-life/main.go:119:55", 247: "10-lru-and-life/main.go:119:55", 248: "10-lru-and-life/main.go:119:55", 249: "10-lru-and-life/main.go:119:55", 258: "10-lru-and-life/main.go:121:6", 259: "10-lru-and-life/main.go:122:10", 260: "10-lru-and-life/main.go:122:10", 261: "10-lru-and-life/main.go:122:10", 262: "10-lru-and-life/main.go:122:10", 263: "10-lru-and-life/main.go:122:10", 264: "10-lru-and-life/main.go:122:10", 265: "10-lru-and-life/main.go:123:2", 266: "10-lru-and-life/main.go:124:20", 267: "10-lru-and-life/main.go:124:20", 268: "10-lru-and-life/main.go:124:20", 269: "10-lru-and-life/main.go:124:20", 270: "10-lru-and-life/main.go:124:20", 272: "10-lru-and-life/main.go:126:10", 273: "10-lru-and-life/main.go:127:2", 274: "10-lru-and-life/main.go:128:13", 276: "10-lru-and-life/main.go:130:2", 295: "10-lru-and-life/main.go:131:2", 296: "10-lru-and-life/main.go:132:2", 297: "10-lru-and-life/main.go:133:2", 298: "10-lru-and-life/main.go:134:6", 299: "10-lru-and-life/main.go:135:14", 300: "10-lru-and-life/main.go:136:3", 301: "10-lru-and-life/main.go:137:3", 302: "10-lru-and-life/main.go:138:4", 303: "10-lru-and-life/main.go:139:4", 304: "10-lru-and-life/main.go:140:12", 307: "10-lru-and-life/main.go:143:6", 308: "10-lru-and-life/main.go:144:19", 309: "10-lru-and-life/main.go:144:19", 310: "10-lru-and-life/main.go:144:19", 311: "10-lru-and-life/main.go:144:19", 312: "10-lru-and-life/main.go:144:19", 313: "10-lru-and-life/main.go:144:19", 314: "10-lru-and-life/main.go:144:19", 315: "10-lru-and-life/main.go:145:4", 317: "10-lru-and-life/main.go:147:22", 318: "10-lru-and-life/main.go:148:3", 319: "10-lru-and-life/main.go:149:3", 320: "10-lru-and-life/main.go:149:3", 321: "10-lru-and-life/main.go:149:3", 322: "10-lru-and-life/main.go:149:3", 323: "10-lru-and-life/main.go:149:3", 324: "10-lru-and-life/main.go:150:3", 325: "10-lru-and-life/main.go:150:3", 326: "10-lru-and-life/main.go:150:3", 327: "10-lru-and-life/main.go:150:3", 328: "10-lru-and-life/main.go:150:3", 329: "10-lru-and-life/main.go:151:11", 331: "10-lru-and-life/main.go:153:24", 332: "10-lru-and-life/main.go:154:2", 333: "10-lru-and-life/main.go:155:12", 335: "10-lru-and-life/main.go:157:2", 342: "10-lru-and-life/main.go:160:6", 343: "10-lru-and-life/main.go:161:7", 344: "10-lru-and-life/main.go:162:2", 345: "10-lru-and-life/main.go:163:3", 346: "10-lru-and-life/main.go:163:15", 348: "10-lru-and-life/main.go:164:4", 349: "10-lru-and-life/main.go:164:22", 351: "10-lru-and-life/main.go:165:5", 357: "10-lru-and-life/main.go:169:14", 358: "10-lru-and-life/main.go:169:14", 359: "10-lru-and-life/main.go:169:14", 360: "10-lru-and-life/main.go:169:14", 361: "10-lru-and-life/main.go:169:14", 362: "10-lru-and-life/main.go:169:14", 363: "10-lru-and-life/main.go:169:14", 364: "10-lru-and-life/main.go:171:7", 365: "10-lru-and-life/main.go:179:2", 366: "10-lru-and-life/main.go:180:3", 367: "10-lru-and-life/main.go:180:3", 368: "10-lru-and-life/main.go:180:3", 369: "10-lru-and-life/main.go:180:3", 370: "10-lru-and-life/main.go:181:7", 371: "10-lru-and-life/main.go:182:7", 373: "10-lru-and-life/main.go:185:2", 374: "10-lru-and-life/main.go:186:23", 375: "10-lru-and-life/main.go:187:3", 377: "10-lru-and-life/main.go:189:26", 378: "10-lru-and-life/main.go:189:26", 379: "10-lru-and-life/main.go:189:26", 380: "10-lru-and-life/main.go:189:26"});

class LRU {
    cap;
    order;
    items;
    hits;
    misses;
    constructor() {
        this.cap = 0;
        this.order = null;
        this.items = null;
        this.hits = 0;
        this.misses = 0;
    }
}

go2jsRegisterTypeName(
    LRU,
    "main.LRU",
    ["Get", "Keys", "Put"],
    [
        {Name: "cap", Tag: "", PkgPath: "main", Anonymous: false, Type: "int", Kind: "int"},
        {
            Name: "order",
            Tag: "",
            PkgPath: "main",
            Anonymous: false,
            Type: "*container/list.List",
            Kind: "ptr"
        },
        {
            Name: "items",
            Tag: "",
            PkgPath: "main",
            Anonymous: false,
            Type: "map[string]*container/list.Element",
            Kind: "map"
        },
        {Name: "hits", Tag: "", PkgPath: "main", Anonymous: false, Type: "int", Kind: "int"},
        {Name: "misses", Tag: "", PkgPath: "main", Anonymous: false, Type: "int", Kind: "int"}
    ],
    "struct"
);
go2jsRegisterValueMethods("main.LRU", []);

class kv {
    key;
    val;
    constructor() {
        this.key = "";
        this.val = 0;
    }
}

go2jsRegisterTypeName(
    kv,
    "main.kv",
    [],
    [
        {Name: "key", Tag: "", PkgPath: "main", Anonymous: false, Type: "string", Kind: "string"},
        {Name: "val", Tag: "", PkgPath: "main", Anonymous: false, Type: "int", Kind: "int"}
    ],
    "struct"
);

function NewLRU$(n) {
    return go2jsNew(
        Object.assign(
            new LRU(),
            {
                cap: n,
                order: go2jsListNew(),
                items: go2jsMapTyped("map[string]*container/list.Element", go2jsMap([]))
            }
        )
    );
}

function* NewLRU(n) {
    return NewLRU$(n);
}

LRU.prototype.Get$ = function(k) {
    {
        let [e, ok] = go2jsMapGetOK(this.items, k, null)
        if (ok) {
            this.order.MoveToFront(e);
            this.hits = go2jsWideAdd(this.hits, 1, "int");
            return [go2jsAssert(e.Value, "*kv", "interface {}").val, true];
        }
    }
    this.misses = go2jsWideAdd(this.misses, 1, "int");
    return [0, false];
};
LRU.prototype.Get = function*(k) {
    return LRU.prototype.Get$.call(this, k);
};
go2jsRegisterMethod("*LRU.Get", function*(receiver, ...args) {
    return yield* LRU.prototype.Get.call(receiver, ...args);
}, 2);

LRU.prototype.Put$ = function(k, v) {
    let evicted = "";
    {
        let [e, ok] = go2jsMapGetOK(this.items, k, null)
        if (ok) {
            const lhs1 = go2jsAssert(e.Value, "*kv", "interface {}");
            lhs1.val = v;
            this.order.MoveToFront(e);
            return "";
        }
    }
    go2jsMapSet(
        this.items,
        k,
        this.order.PushFront(
            go2jsTypedNilGuard(go2jsNew(Object.assign(new kv(), {key: k, val: v})), "*main.kv")
        )
    );
    if (this.order.Len() > this.cap) {
        let last = this.order.Back();
        this.order.Remove(last);
        go2jsMapDelete(this.items, go2jsAssert(last.Value, "*kv", "interface {}").key);
        return go2jsAssert(last.Value, "*kv", "interface {}").key;
    }
    return "";
};
LRU.prototype.Put = function*(k, v) {
    return LRU.prototype.Put$.call(this, k, v);
};
go2jsRegisterMethod("*LRU.Put", function*(receiver, ...args) {
    return yield* LRU.prototype.Put.call(receiver, ...args);
}, 1);

LRU.prototype.Keys$ = function() {
    let out = null;
    for (let e = this.order.Front(); !go2jsIsNil(e); e = e.Next()) {
        out = go2jsSliceAppend(out, go2jsAssert(e.Value, "*kv", "interface {}").key);
    }
    return out;
};
LRU.prototype.Keys = function*() {
    return LRU.prototype.Keys$.call(this);
};
go2jsRegisterMethod("*LRU.Keys", function*(receiver, ...args) {
    return yield* LRU.prototype.Keys.call(receiver, ...args);
}, 1);

function ParseGrid$(...rows) {
    let g = go2jsSliceMake(go2jsLen(rows), go2jsLen(rows), null, true);
    for (const [y, r] of go2jsRangeSequence(rows).entries()) {
        go2jsSliceStore(g, y, go2jsSliceMake(go2jsLen(r), go2jsLen(r), false, true));
        for (const range2 of go2jsRangeValue(r)) {
            let x = go2jsCopy(range2[0]);
            let c = go2jsCopy(range2[1]);
            go2jsSliceStore(go2jsIndex(g, y), x, c == 35);
        }
    }
    return g;
}

function* ParseGrid(...rows) {
    return ParseGrid$(...rows);
}

function* go2jsMethodGridNeighbours(g, x, y) {
    let n = 0;
    let [h, w] = [go2jsLen(g), go2jsLen(go2jsIndex(g, 0))];
    for (let dy = go2jsWideWrap(-1, "int"); dy <= 1; dy = go2jsWideAdd(dy, 1, "int")) {
        for (let dx = go2jsWideWrap(-1, "int"); dx <= 1; dx = go2jsWideAdd(dx, 1, "int")) {
            if ((dx != 0 || dy != 0) && go2jsIndex(go2jsIndex(g, go2jsMod((go2jsWideAdd(go2jsWideAdd(y, dy, "int"), h, "int")), h)), go2jsMod((go2jsWideAdd(go2jsWideAdd(x, dx, "int"), w, "int")), w))) {
                n = go2jsWideAdd(n, 1, "int");
            }
        }
    }
    return n;
}

go2jsRegisterMethod("Grid.Neighbours", go2jsMethodGridNeighbours);

function* go2jsMethodGridStep(g) {
    let next = go2jsSliceMake(go2jsLen(g), go2jsLen(g), null, true);
    for (const y of go2jsRangeMap(g).keys()) {
        go2jsSliceStore(
            next,
            y,
            go2jsSliceMake(go2jsLen(go2jsIndex(g, y)), go2jsLen(go2jsIndex(g, y)), false, true)
        );
        for (const x of go2jsRangeMap(go2jsIndex(g, y)).keys()) {
            let n = go2jsNamedMethodCall(g, "Grid", "Neighbours", x, y);
            go2jsSliceStore(go2jsIndex(next, y), x, n == 3 || (go2jsIndex(go2jsIndex(g, y), x) && n == 2));
        }
    }
    return next;
}

go2jsRegisterMethod("Grid.Step", go2jsMethodGridStep);

function* go2jsMethodGridString(g) {
    let sb = new go2jsStringsBuilder();
    for (const row of go2jsRangeSequence(g)) {
        for (const c of go2jsRangeSequence(row)) {
            if (c) {
                sb.WriteByte(35);
            } else {
                sb.WriteByte(46);
            }
        }
        sb.WriteByte(10);
    }
    return sb.String();
}

go2jsRegisterMethod("Grid.String", go2jsMethodGridString);

class RNG {
    state;
    constructor() {
        this.state = 0;
    }
}

go2jsRegisterTypeName(
    RNG,
    "main.RNG",
    ["Intn", "Next"],
    [{Name: "state", Tag: "", PkgPath: "main", Anonymous: false, Type: "uint32", Kind: "uint32"}],
    "struct"
);
go2jsRegisterValueMethods("main.RNG", []);

RNG.prototype.Next$ = function() {
    this.state = go2jsIntWrap(go2jsIntWrap(go2jsImul(this.state, 1664525), "uint32") + 1013904223, "uint32");
    return go2jsIntWrap(go2jsShr(this.state, 8), "uint32");
};
RNG.prototype.Next = function*() {
    return RNG.prototype.Next$.call(this);
};
go2jsRegisterMethod("*RNG.Next", function*(receiver, ...args) {
    return yield* RNG.prototype.Next.call(receiver, ...args);
}, 1);

RNG.prototype.Intn$ = function(n) {
    return Math.trunc(
        go2jsIntWrap(
            go2jsMod(go2jsMethodOn(this, RNG, "Next$", []), go2jsWideToUnsigned(n, "uint32")),
            "uint32"
        )
    );
};
RNG.prototype.Intn = function*(n) {
    return RNG.prototype.Intn$.call(this, n);
};
go2jsRegisterMethod("*RNG.Intn", function*(receiver, ...args) {
    return yield* RNG.prototype.Intn.call(receiver, ...args);
}, 1);

function maze$(w, h, rng) {
    let grid = go2jsSliceMake(
        go2jsWideAdd(go2jsWideMul(h, 2, "int"), 1, "int"),
        go2jsWideAdd(go2jsWideMul(h, 2, "int"), 1, "int"),
        null,
        true
    );
    for (const y of go2jsRangeMap(grid).keys()) {
        go2jsSliceStore(
            grid,
            y,
            go2jsStringToBytes(go2jsStringsRepeat("#", go2jsWideAdd(go2jsWideMul(w, 2, "int"), 1, "int")))
        );
    }
    let seen = go2jsSliceMake(h, h, null, true);
    for (const i of go2jsRangeMap(seen).keys()) {
        go2jsSliceStore(seen, i, go2jsSliceMake(w, w, false, true));
    }
    class cell {
        x;
        y;
        constructor() {
            this.x = 0;
            this.y = 0;
        }
    }
    go2jsRegisterTypeName(
        cell,
        "main.cell",
        [],
        [
            {Name: "x", Tag: "", PkgPath: "main", Anonymous: false, Type: "int", Kind: "int"},
            {Name: "y", Tag: "", PkgPath: "main", Anonymous: false, Type: "int", Kind: "int"}
        ],
        "struct"
    );

    let stack = go2jsSliceTyped("[]main.cell", [Object.assign(new cell(), {x: 0, y: 0})]);
    go2jsSliceStore(go2jsIndex(seen, 0), 0, true);
    go2jsSliceStore(go2jsIndex(grid, 1), 1, 32);
    for (; go2jsLen(stack) > 0;) {
        let c = go2jsStructCopy(go2jsIndex(stack, go2jsWideSub(go2jsLen(stack), 1, "int")));
        let opts = null;
        for (const d of go2jsRangeCopies(go2jsRangeSequence(go2jsSliceTyped("[]main.cell", [Object.assign(new cell(), {x: 1, y: 0}), Object.assign(new cell(), {x: go2jsWideWrap(-1, "int"), y: 0}), Object.assign(new cell(), {x: 0, y: 1}), Object.assign(new cell(), {x: 0, y: go2jsWideWrap(-1, "int")})])))) {
            let [nx, ny] = [go2jsWideAdd(c.x, d.x, "int"), go2jsWideAdd(c.y, d.y, "int")];
            if (nx >= 0 && ny >= 0 && nx < w && ny < h && !go2jsIndex(go2jsIndex(seen, ny), nx)) {
                opts = go2jsSliceAppend(opts, Object.assign(new cell(), {x: nx, y: ny}));
            }
        }
        if (go2jsLen(opts) == 0) {
            stack = go2jsSliceRange(
                stack,
                undefined,
                go2jsWideSub(go2jsLen(stack), 1, "int"),
                undefined,
                false
            );
            continue;
        }
        let n = go2jsStructCopy(go2jsIndex(opts, go2jsMethodOn(rng, RNG, "Intn$", [go2jsLen(opts)])));
        go2jsSliceStore(go2jsIndex(seen, n.y), n.x, true);
        go2jsSliceStore(
            go2jsIndex(grid, go2jsWideAdd(go2jsWideMul(n.y, 2, "int"), 1, "int")),
            go2jsWideAdd(go2jsWideMul(n.x, 2, "int"), 1, "int"),
            32
        );
        go2jsSliceStore(
            go2jsIndex(grid, go2jsWideAdd(go2jsWideAdd(c.y, n.y, "int"), 1, "int")),
            go2jsWideAdd(go2jsWideAdd(c.x, n.x, "int"), 1, "int"),
            32
        );
        stack = go2jsSliceAppend(stack, go2jsStructCopy(n));
    }
    let out = go2jsSliceMake(go2jsLen(grid), go2jsLen(grid), "", true);
    for (const [i, row] of go2jsRangeSequence(grid).entries()) {
        go2jsSliceStore(out, i, go2jsBytesToString(row));
    }
    return out;
}

function* maze(w, h, rng) {
    return maze$(w, h, rng);
}

function* main() {
    let c = NewLRU$(3);
    for (const k of go2jsRangeSequence(go2jsSliceTyped("[]string", ["a", "b", "c", "a", "d", "b", "e", "a"]))) {
        {
            let [, ok] = go2jsMethodOn(c, LRU, "Get$", [k])
            if (!ok) {
                {
                    let ev = go2jsMethodOn(c, LRU, "Put$", [k, go2jsLen(k)])
                    if (ev != "") {
                        go2jsWriteOut(process.stdout, go2jsSprintf("put %s evicts %s\n", k, ev));
                    }
                }
            }
        }
    }
    go2jsPrintln(
        go2jsTyped(go2jsMethodOn(c, LRU, "Keys$", []), "[]string", "", "slice"),
        "hits",
        c.hits,
        "misses",
        c.misses
    );
    let g = ParseGrid$("..........", "...#......", "....#.....", "..###.....", "..........", "..........");
    for (let gen = 0; gen < 3; gen = go2jsWideAdd(gen, 1, "int")) {
        go2jsWriteOut(
            process.stdout,
            go2jsSprintf("generation %d\n%s", gen, yield* go2jsMethodGridString(g))
        );
        g = go2jsNamedMethodCall(g, "Grid", "Step");
        g = go2jsNamedMethodCall(g, "Grid", "Step");
    }
    let rng = go2jsNew(Object.assign(new RNG(), {state: 42}));
    for (const line of go2jsRangeSequence(maze$(10, 5, rng))) {
        go2jsPrintln(line);
    }
    go2jsPrintln(
        go2jsTyped(go2jsMethodOn(rng, RNG, "Next$", []), "uint32"),
        go2jsMethodOn(rng, RNG, "Intn$", [100])
    );
}

try {
    go2jsRunMain((function*() {
        return yield* main();
    })());
} catch (e) {
    go2jsReportUncaughtPanic(e);
}
