"use strict";

const {
    go2jsAssert,
    go2jsDeref,
    go2jsHeapPop,
    go2jsHeapPush,
    go2jsIndex,
    go2jsIndexCheck,
    go2jsInterface,
    go2jsLen,
    go2jsMap,
    go2jsMapGet,
    go2jsMapGetOK,
    go2jsMapSet,
    go2jsMapTyped,
    go2jsMapUpdate,
    go2jsMarkArray,
    go2jsMethodOn,
    go2jsMethodOnYield,
    go2jsNamedMethodCall,
    go2jsNew,
    go2jsNilInterface,
    go2jsNilValue,
    go2jsPreempt,
    go2jsPrintln,
    go2jsRangeCopies,
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
    go2jsSortSlice,
    go2jsSortStrings,
    go2jsSprintf,
    go2jsState,
    go2jsStorePtr,
    go2jsStrCompare,
    go2jsStringsJoin,
    go2jsStructCopy,
    go2jsTyped,
    go2jsWideAdd,
    go2jsWideSub,
    go2jsWrapError,
    go2jsWriteOut,
} = require("./main.runtime.js");

go2jsSetPackageName("main");
go2jsSetSourceLines({38: "07-graphs/main.go:21:6", 39: "07-graphs/main.go:21:26", 40: "07-graphs/main.go:21:26", 41: "07-graphs/main.go:21:26", 48: "07-graphs/main.go:23:17", 49: "07-graphs/main.go:24:13", 50: "07-graphs/main.go:24:13", 51: "07-graphs/main.go:24:13", 52: "07-graphs/main.go:24:13", 53: "07-graphs/main.go:24:13", 54: "07-graphs/main.go:25:2", 57: "07-graphs/main.go:26:3", 60: "07-graphs/main.go:28:2", 61: "07-graphs/main.go:29:14", 62: "07-graphs/main.go:29:14", 63: "07-graphs/main.go:29:14", 64: "07-graphs/main.go:29:14", 65: "07-graphs/main.go:29:14", 75: "07-graphs/main.go:33:17", 76: "07-graphs/main.go:34:27", 77: "07-graphs/main.go:35:2", 78: "07-graphs/main.go:36:9", 80: "07-graphs/main.go:38:2", 81: "07-graphs/main.go:39:2", 90: "07-graphs/main.go:42:17", 92: "07-graphs/main.go:43:2", 93: "07-graphs/main.go:44:2", 94: "07-graphs/main.go:45:2", 95: "07-graphs/main.go:46:6", 96: "07-graphs/main.go:47:3", 97: "07-graphs/main.go:48:3", 98: "07-graphs/main.go:49:11", 99: "07-graphs/main.go:50:18", 100: "07-graphs/main.go:50:18", 101: "07-graphs/main.go:50:18", 102: "07-graphs/main.go:50:18", 103: "07-graphs/main.go:51:3", 104: "07-graphs/main.go:51:42", 106: "07-graphs/main.go:52:3", 107: "07-graphs/main.go:53:4", 108: "07-graphs/main.go:54:5", 109: "07-graphs/main.go:55:13", 113: "07-graphs/main.go:59:2", 142: "07-graphs/main.go:68:13", 144: "07-graphs/main.go:68:32", 149: "07-graphs/main.go:69:13", 151: "07-graphs/main.go:70:2", 152: "07-graphs/main.go:70:2", 153: "07-graphs/main.go:70:2", 154: "07-graphs/main.go:70:2", 155: "07-graphs/main.go:70:2", 160: "07-graphs/main.go:72:13", 162: "07-graphs/main.go:72:30", 169: "07-graphs/main.go:73:14", 171: "07-graphs/main.go:73:35", 172: "07-graphs/main.go:73:35", 173: "07-graphs/main.go:73:35", 174: "07-graphs/main.go:73:35", 179: "07-graphs/main.go:74:14", 181: "07-graphs/main.go:75:2", 182: "07-graphs/main.go:76:7", 183: "07-graphs/main.go:77:2", 184: "07-graphs/main.go:78:2", 185: "07-graphs/main.go:79:2", 190: "07-graphs/main.go:82:17", 191: "07-graphs/main.go:83:2", 192: "07-graphs/main.go:84:2", 193: "07-graphs/main.go:85:2", 194: "07-graphs/main.go:86:3", 196: "07-graphs/main.go:88:2", 197: "07-graphs/main.go:89:2", 202: "07-graphs/main.go:90:6", 206: "07-graphs/main.go:91:10", 207: "07-graphs/main.go:91:10", 208: "07-graphs/main.go:91:10", 209: "07-graphs/main.go:92:3", 210: "07-graphs/main.go:93:4", 212: "07-graphs/main.go:95:3", 213: "07-graphs/main.go:96:4", 216: "07-graphs/main.go:97:5", 217: "07-graphs/main.go:98:5", 218: "07-graphs/main.go:99:5", 219: "07-graphs/main.go:99:5", 220: "07-graphs/main.go:99:5", 221: "07-graphs/main.go:99:5", 226: "07-graphs/main.go:103:2", 232: "07-graphs/main.go:106:6", 233: "07-graphs/main.go:107:2", 234: "07-graphs/main.go:108:2", 235: "07-graphs/main.go:109:7", 237: "07-graphs/main.go:111:9", 244: "07-graphs/main.go:114:17", 245: "07-graphs/main.go:115:2", 246: "07-graphs/main.go:116:2", 247: "07-graphs/main.go:117:3", 250: "07-graphs/main.go:118:4", 253: "07-graphs/main.go:120:3", 254: "07-graphs/main.go:121:4", 257: "07-graphs/main.go:124:2", 258: "07-graphs/main.go:125:2", 259: "07-graphs/main.go:126:3", 260: "07-graphs/main.go:127:12", 263: "07-graphs/main.go:130:2", 264: "07-graphs/main.go:131:2", 265: "07-graphs/main.go:132:6", 266: "07-graphs/main.go:133:3", 267: "07-graphs/main.go:134:3", 268: "07-graphs/main.go:135:11", 269: "07-graphs/main.go:136:3", 270: "07-graphs/main.go:137:4", 271: "07-graphs/main.go:138:4", 272: "07-graphs/main.go:139:13", 273: "07-graphs/main.go:140:5", 277: "07-graphs/main.go:144:19", 278: "07-graphs/main.go:145:77", 279: "07-graphs/main.go:145:77", 280: "07-graphs/main.go:145:77", 281: "07-graphs/main.go:145:77", 283: "07-graphs/main.go:147:2", 313: "07-graphs/main.go:152:6", 314: "07-graphs/main.go:153:34", 315: "07-graphs/main.go:153:34", 316: "07-graphs/main.go:153:34", 317: "07-graphs/main.go:153:34", 318: "07-graphs/main.go:153:34", 319: "07-graphs/main.go:153:34", 320: "07-graphs/main.go:154:2", 321: "07-graphs/main.go:155:3", 325: "07-graphs/main.go:157:2", 332: "07-graphs/main.go:159:21", 333: "07-graphs/main.go:160:2", 334: "07-graphs/main.go:161:3", 335: "07-graphs/main.go:162:3", 337: "07-graphs/main.go:164:2", 346: "07-graphs/main.go:166:21", 347: "07-graphs/main.go:167:20", 348: "07-graphs/main.go:168:2", 349: "07-graphs/main.go:169:3", 351: "07-graphs/main.go:171:2", 352: "07-graphs/main.go:172:3", 354: "07-graphs/main.go:174:2", 355: "07-graphs/main.go:175:2", 357: "07-graphs/main.go:176:2", 366: "07-graphs/main.go:179:6", 367: "07-graphs/main.go:180:7", 368: "07-graphs/main.go:181:2", 369: "07-graphs/main.go:181:2", 370: "07-graphs/main.go:181:2", 371: "07-graphs/main.go:181:2", 372: "07-graphs/main.go:181:2", 373: "07-graphs/main.go:181:2", 374: "07-graphs/main.go:181:2", 375: "07-graphs/main.go:181:2", 376: "07-graphs/main.go:181:2", 377: "07-graphs/main.go:181:2", 378: "07-graphs/main.go:181:2", 379: "07-graphs/main.go:181:2", 380: "07-graphs/main.go:181:2", 381: "07-graphs/main.go:181:2", 382: "07-graphs/main.go:181:2", 383: "07-graphs/main.go:181:2", 384: "07-graphs/main.go:181:2", 385: "07-graphs/main.go:181:2", 386: "07-graphs/main.go:181:2", 387: "07-graphs/main.go:181:2", 388: "07-graphs/main.go:181:2", 389: "07-graphs/main.go:181:2", 390: "07-graphs/main.go:181:2", 391: "07-graphs/main.go:181:2", 392: "07-graphs/main.go:181:2", 393: "07-graphs/main.go:181:2", 394: "07-graphs/main.go:181:2", 395: "07-graphs/main.go:181:2", 396: "07-graphs/main.go:181:2", 397: "07-graphs/main.go:181:2", 398: "07-graphs/main.go:181:2", 399: "07-graphs/main.go:181:2", 400: "07-graphs/main.go:181:2", 401: "07-graphs/main.go:181:2", 402: "07-graphs/main.go:181:2", 403: "07-graphs/main.go:181:2", 404: "07-graphs/main.go:181:2", 405: "07-graphs/main.go:181:2", 406: "07-graphs/main.go:181:2", 407: "07-graphs/main.go:181:2", 408: "07-graphs/main.go:181:2", 409: "07-graphs/main.go:181:2", 410: "07-graphs/main.go:181:2", 411: "07-graphs/main.go:181:2", 412: "07-graphs/main.go:181:2", 413: "07-graphs/main.go:181:2", 414: "07-graphs/main.go:181:2", 415: "07-graphs/main.go:181:2", 416: "07-graphs/main.go:181:2", 417: "07-graphs/main.go:181:2", 418: "07-graphs/main.go:181:2", 419: "07-graphs/main.go:181:2", 420: "07-graphs/main.go:181:2", 421: "07-graphs/main.go:181:2", 422: "07-graphs/main.go:181:2", 423: "07-graphs/main.go:181:2", 424: "07-graphs/main.go:181:2", 425: "07-graphs/main.go:181:2", 426: "07-graphs/main.go:181:2", 427: "07-graphs/main.go:181:2", 428: "07-graphs/main.go:181:2", 429: "07-graphs/main.go:181:2", 430: "07-graphs/main.go:181:2", 431: "07-graphs/main.go:181:2", 432: "07-graphs/main.go:181:2", 433: "07-graphs/main.go:181:2", 434: "07-graphs/main.go:181:2", 435: "07-graphs/main.go:181:2", 436: "07-graphs/main.go:181:2", 437: "07-graphs/main.go:181:2", 438: "07-graphs/main.go:181:2", 439: "07-graphs/main.go:181:2", 440: "07-graphs/main.go:181:2", 441: "07-graphs/main.go:181:2", 442: "07-graphs/main.go:181:2", 443: "07-graphs/main.go:181:2", 444: "07-graphs/main.go:181:2", 445: "07-graphs/main.go:181:2", 446: "07-graphs/main.go:181:2", 447: "07-graphs/main.go:181:2", 448: "07-graphs/main.go:181:2", 449: "07-graphs/main.go:181:2", 450: "07-graphs/main.go:181:2", 451: "07-graphs/main.go:181:2", 452: "07-graphs/main.go:181:2", 453: "07-graphs/main.go:181:2", 454: "07-graphs/main.go:181:2", 455: "07-graphs/main.go:181:2", 456: "07-graphs/main.go:181:2", 457: "07-graphs/main.go:181:2", 458: "07-graphs/main.go:181:2", 459: "07-graphs/main.go:181:2", 460: "07-graphs/main.go:181:2", 461: "07-graphs/main.go:181:2", 462: "07-graphs/main.go:181:2", 463: "07-graphs/main.go:181:2", 464: "07-graphs/main.go:181:2", 465: "07-graphs/main.go:181:2", 466: "07-graphs/main.go:181:2", 467: "07-graphs/main.go:181:2", 468: "07-graphs/main.go:181:2", 469: "07-graphs/main.go:181:2", 470: "07-graphs/main.go:181:2", 471: "07-graphs/main.go:181:2", 472: "07-graphs/main.go:181:2", 473: "07-graphs/main.go:181:2", 474: "07-graphs/main.go:181:2", 475: "07-graphs/main.go:181:2", 476: "07-graphs/main.go:181:2", 477: "07-graphs/main.go:181:2", 478: "07-graphs/main.go:181:2", 479: "07-graphs/main.go:181:2", 480: "07-graphs/main.go:181:2", 481: "07-graphs/main.go:181:2", 482: "07-graphs/main.go:181:2", 483: "07-graphs/main.go:181:2", 484: "07-graphs/main.go:181:2", 485: "07-graphs/main.go:181:2", 486: "07-graphs/main.go:181:2", 487: "07-graphs/main.go:181:2", 488: "07-graphs/main.go:181:2", 489: "07-graphs/main.go:181:2", 490: "07-graphs/main.go:181:2", 491: "07-graphs/main.go:181:2", 492: "07-graphs/main.go:181:2", 493: "07-graphs/main.go:181:2", 494: "07-graphs/main.go:181:2", 495: "07-graphs/main.go:181:2", 496: "07-graphs/main.go:181:2", 497: "07-graphs/main.go:181:2", 498: "07-graphs/main.go:181:2", 499: "07-graphs/main.go:181:2", 500: "07-graphs/main.go:181:2", 501: "07-graphs/main.go:181:2", 502: "07-graphs/main.go:181:2", 503: "07-graphs/main.go:181:2", 504: "07-graphs/main.go:185:3", 506: "07-graphs/main.go:187:29", 507: "07-graphs/main.go:188:16", 508: "07-graphs/main.go:189:20", 509: "07-graphs/main.go:190:51", 510: "07-graphs/main.go:190:51", 511: "07-graphs/main.go:190:51", 512: "07-graphs/main.go:190:51", 514: "07-graphs/main.go:193:10", 515: "07-graphs/main.go:194:2", 516: "07-graphs/main.go:195:3", 518: "07-graphs/main.go:197:16", 519: "07-graphs/main.go:198:2", 520: "07-graphs/main.go:199:2", 521: "07-graphs/main.go:200:11", 522: "07-graphs/main.go:201:2", 523: "07-graphs/main.go:203:2", 545: "07-graphs/main.go:204:2", 546: "07-graphs/main.go:204:2", 547: "07-graphs/main.go:204:2", 548: "07-graphs/main.go:204:2", 549: "07-graphs/main.go:204:2", 550: "07-graphs/main.go:204:2", 551: "07-graphs/main.go:204:2", 552: "07-graphs/main.go:204:2", 553: "07-graphs/main.go:204:2", 554: "07-graphs/main.go:204:2", 555: "07-graphs/main.go:204:2", 556: "07-graphs/main.go:204:2", 557: "07-graphs/main.go:205:2", 558: "07-graphs/main.go:205:42", 560: "07-graphs/main.go:206:8", 561: "07-graphs/main.go:207:2", 562: "07-graphs/main.go:208:2", 563: "07-graphs/main.go:209:2", 564: "07-graphs/main.go:210:6", 565: "07-graphs/main.go:211:4", 566: "07-graphs/main.go:212:24", 569: "07-graphs/main.go:215:2"});

class Edge {
    To;
    Weight;
    constructor() {
        this.To = "";
        this.Weight = 0;
    }
}

go2jsRegisterTypeName(
    Edge,
    "main.Edge",
    [],
    [
        {Name: "To", Tag: "", PkgPath: "", Anonymous: false, Type: "string", Kind: "string"},
        {Name: "Weight", Tag: "", PkgPath: "", Anonymous: false, Type: "int", Kind: "int"}
    ],
    "struct"
);

class Graph {
    adj;
    constructor() {
        this.adj = null;
    }
}

go2jsRegisterTypeName(
    Graph,
    "main.Graph",
    ["AddEdge", "BFS", "Dijkstra", "Nodes", "TopoSort"],
    [{Name: "adj", Tag: "", PkgPath: "main", Anonymous: false, Type: "map[string][]main.Edge", Kind: "map"}],
    "struct"
);
go2jsRegisterValueMethods("main.Graph", []);

function NewGraph$() {
    return go2jsNew(
        Object.assign(new Graph(), {adj: go2jsMapTyped("map[string][]main.Edge", go2jsMap([]))})
    );
}

function* NewGraph() {
    return NewGraph$();
}

Graph.prototype.AddEdge$ = function(a, b, w, directed) {
    go2jsMapSet(
        this.adj,
        a,
        go2jsSliceAppend(go2jsMapGet(this.adj, a, null), Object.assign(new Edge(), {To: b, Weight: w}))
    );
    {
        let [, ok] = go2jsMapGetOK(this.adj, b, null)
        if (!ok) {
            go2jsMapSet(this.adj, b, null);
        }
    }
    if (!directed) {
        go2jsMapSet(
            this.adj,
            b,
            go2jsSliceAppend(go2jsMapGet(this.adj, b, null), Object.assign(new Edge(), {To: a, Weight: w}))
        );
    }
};
Graph.prototype.AddEdge = function*(a, b, w, directed) {
    return Graph.prototype.AddEdge$.call(this, a, b, w, directed);
};
go2jsRegisterMethod("*Graph.AddEdge", function*(receiver, ...args) {
    return yield* Graph.prototype.AddEdge.call(receiver, ...args);
}, 0);

Graph.prototype.Nodes$ = function() {
    let out = go2jsSliceMake(0, go2jsLen(this.adj), "", true);
    for (const n of go2jsRangeMap(this.adj).keys()) {
        out = go2jsSliceAppend(out, n);
    }
    go2jsSortStrings(out);
    return out;
};
Graph.prototype.Nodes = function*() {
    return Graph.prototype.Nodes$.call(this);
};
go2jsRegisterMethod("*Graph.Nodes", function*(receiver, ...args) {
    return yield* Graph.prototype.Nodes.call(receiver, ...args);
}, 1);

Graph.prototype.BFS$ = function(start) {
    const go2jsReceiver_g = this;
    let seen = go2jsMapTyped("map[string]bool", go2jsMap([[start, true]]));
    let queue = go2jsSliceTyped("[]string", [start]);
    let order = null;
    for (; go2jsLen(queue) > 0;) {
        let n = go2jsIndex(queue, 0);
        queue = go2jsSliceRange(queue, 1, undefined, undefined, false);
        order = go2jsSliceAppend(order, n);
        let next = go2jsSliceAppend(
            go2jsNilValue("[]main.Edge", "slice"),
            ...(go2jsMapGet(this.adj, n, null) ?? [])
        );
        go2jsSortSlice(go2jsNilInterface(next, "[]main.Edge", "slice", true), function*(i, j) {
            return (go2jsStrCompare(go2jsIndex(next, i).To, go2jsIndex(next, j).To) < 0);
        });
        for (const e of go2jsRangeCopies(go2jsRangeSequence(next))) {
            if (!go2jsMapGet(seen, e.To, false)) {
                go2jsMapSet(seen, e.To, true);
                queue = go2jsSliceAppend(queue, e.To);
            }
        }
    }
    return order;
};
Graph.prototype.BFS = function*(start) {
    return Graph.prototype.BFS$.call(this, start);
};
go2jsRegisterMethod("*Graph.BFS", function*(receiver, ...args) {
    return yield* Graph.prototype.BFS.call(receiver, ...args);
}, 1);

class entry {
    node;
    dist;
    constructor() {
        this.node = "";
        this.dist = 0;
    }
}

go2jsRegisterTypeName(
    entry,
    "main.entry",
    [],
    [
        {Name: "node", Tag: "", PkgPath: "main", Anonymous: false, Type: "string", Kind: "string"},
        {Name: "dist", Tag: "", PkgPath: "main", Anonymous: false, Type: "int", Kind: "int"}
    ],
    "struct"
);

function* go2jsMethodpqLen(p) {
    const go2jsReceiver_g = this;
    return go2jsLen(p);
}

go2jsRegisterMethod("pq.Len", go2jsMethodpqLen);

function* go2jsMethodpqLess(p, i, j) {
    const go2jsReceiver_g = this;
    return go2jsIndex(p, i).dist < go2jsIndex(p, j).dist || (
        go2jsIndex(p, i).dist == go2jsIndex(p, j).dist && (
            go2jsStrCompare(go2jsIndex(p, i).node, go2jsIndex(p, j).node) < 0
        )
    );
}

go2jsRegisterMethod("pq.Less", go2jsMethodpqLess);

function* go2jsMethodpqSwap(p, i, j) {
    const go2jsReceiver_g = this;
    go2jsIndexCheck(p, i);
    go2jsIndexCheck(p, j);
    [p[i], p[j]] = [go2jsIndex(p, j), go2jsIndex(p, i)];
}

go2jsRegisterMethod("pq.Swap", go2jsMethodpqSwap);

function* go2jsMethodpqPush(p, x) {
    const go2jsReceiver_g = this;
    go2jsStorePtr(
        p,
        go2jsSliceAppend(go2jsDeref(p), go2jsStructCopy(go2jsAssert(x, "entry", "interface {}")))
    );
}

go2jsRegisterMethod("pq.Push", go2jsMethodpqPush);

function* go2jsMethodpqPop(p) {
    const go2jsReceiver_g = this;
    let old = go2jsDeref(p);
    let n = go2jsLen(old);
    let x = go2jsStructCopy(go2jsIndex(old, go2jsWideSub(n, 1, "int")));
    go2jsStorePtr(p, go2jsSliceRange(old, undefined, go2jsWideSub(n, 1, "int"), undefined, false));
    return go2jsStructCopy(x);
}

go2jsRegisterMethod("pq.Pop", go2jsMethodpqPop);

Graph.prototype.Dijkstra = function*(src) {
    let dist = go2jsMapTyped("map[string]int", go2jsMap([]));
    let prev = go2jsMapTyped("map[string]string", go2jsMap([]));
    for (const n of go2jsRangeMap(this.adj).keys()) {
        go2jsMapSet(dist, n, 9223372036854775807n);
    }
    go2jsMapSet(dist, src, 0);
    let h = go2jsNew(yield* (function*() {
        const literal1 = [new entry()];
        literal1[0] = Object.assign(new entry(), {node: src, dist: 0});
        return literal1;
    })());
    for (; go2jsNamedMethodCall(h, "pq", "Len") > 0;) {
        if (++go2jsState.go2jsTicks > 20000) {
            yield* go2jsPreempt();
        }
        let cur = go2jsStructCopy(
            go2jsAssert(go2jsHeapPop(go2jsInterface(h, "*pq", "*main.pq")), "entry", "interface {}")
        );
        if (cur.dist > go2jsMapGet(dist, cur.node, 0)) {
            continue;
        }
        for (const e of go2jsRangeCopies(go2jsRangeSequence(go2jsMapGet(this.adj, cur.node, null)))) {
            {
                let nd = go2jsWideAdd(cur.dist, e.Weight, "int")
                if (nd < go2jsMapGet(dist, e.To, 0)) {
                    go2jsMapSet(dist, e.To, nd);
                    go2jsMapSet(prev, e.To, cur.node);
                    go2jsHeapPush(
                        go2jsInterface(h, "*pq", "*main.pq"),
                        Object.assign(new entry(), {node: e.To, dist: nd})
                    );
                }
            }
        }
    }
    return [dist, prev];
}
go2jsRegisterMethod("*Graph.Dijkstra", function*(receiver, ...args) {
    return yield* Graph.prototype.Dijkstra.call(receiver, ...args);
}, 2);

function path$(prev, to) {
    let p = null;
    for (let n = to; n != ""; n = go2jsMapGet(prev, n, "")) {
        p = go2jsSliceAppend(go2jsSliceTyped("[]string", [n]), ...(p ?? []));
    }
    return go2jsStringsJoin(p, " -> ");
}

function* path(prev, to) {
    return path$(prev, to);
}

Graph.prototype.TopoSort$ = function() {
    let indeg = go2jsMapTyped("map[string]int", go2jsMap([]));
    for (const [n, es] of go2jsRangeSequence(this.adj).entries()) {
        {
            let [, ok] = go2jsMapGetOK(indeg, n, 0)
            if (!ok) {
                go2jsMapSet(indeg, n, 0);
            }
        }
        for (const e of go2jsRangeCopies(go2jsRangeSequence(es))) {
            go2jsMapUpdate(indeg, e.To, 0, (current) => go2jsWideAdd(current, 1, "int"));
        }
    }
    let ready = null;
    for (const [n, d] of go2jsRangeSequence(indeg).entries()) {
        if (d == 0) {
            ready = go2jsSliceAppend(ready, n);
        }
    }
    go2jsSortStrings(ready);
    let order = null;
    for (; go2jsLen(ready) > 0;) {
        let n = go2jsIndex(ready, 0);
        ready = go2jsSliceRange(ready, 1, undefined, undefined, false);
        order = go2jsSliceAppend(order, n);
        for (const e of go2jsRangeCopies(go2jsRangeSequence(go2jsMapGet(this.adj, n, null)))) {
            go2jsMapUpdate(indeg, e.To, 0, (current) => go2jsWideSub(current, 1, "int"));
            if (go2jsMapGet(indeg, e.To, 0) == 0) {
                ready = go2jsSliceAppend(ready, e.To);
                go2jsSortStrings(ready);
            }
        }
    }
    if (go2jsLen(order) != go2jsLen(this.adj)) {
        return [
            null,
            go2jsWrapError("cycle detected after %d of %d nodes", go2jsLen(order), go2jsLen(this.adj))
        ];
    }
    return [order, null];
};
Graph.prototype.TopoSort = function*() {
    return Graph.prototype.TopoSort$.call(this);
};
go2jsRegisterMethod("*Graph.TopoSort", function*(receiver, ...args) {
    return yield* Graph.prototype.TopoSort.call(receiver, ...args);
}, 2);

class UnionFind {
    parent;
    size;
    constructor() {
        this.parent = null;
        this.size = null;
    }
}

go2jsRegisterTypeName(
    UnionFind,
    "main.UnionFind",
    ["Find", "Union"],
    [
        {Name: "parent", Tag: "", PkgPath: "main", Anonymous: false, Type: "[]int", Kind: "slice"},
        {Name: "size", Tag: "", PkgPath: "main", Anonymous: false, Type: "[]int", Kind: "slice"}
    ],
    "struct"
);
go2jsRegisterValueMethods("main.UnionFind", []);

function NewUF$(n) {
    let u = go2jsNew(
        Object.assign(
            new UnionFind(),
            {parent: go2jsSliceMake(n, n, 0, true), size: go2jsSliceMake(n, n, 0, true)}
        )
    );
    for (const i of go2jsRangeMap(u.parent).keys()) {
        go2jsIndexCheck(u.parent, i);
        go2jsIndexCheck(u.size, i);
        [u.parent[i], u.size[i]] = [i, 1];
    }
    return u;
}

function* NewUF(n) {
    return NewUF$(n);
}

UnionFind.prototype.Find$ = function(x) {
    for (; go2jsIndex(this.parent, x) != x;) {
        go2jsSliceStore(this.parent, x, go2jsIndex(this.parent, go2jsIndex(this.parent, x)));
        x = go2jsIndex(this.parent, x);
    }
    return x;
};
UnionFind.prototype.Find = function*(x) {
    return UnionFind.prototype.Find$.call(this, x);
};
go2jsRegisterMethod("*UnionFind.Find", function*(receiver, ...args) {
    return yield* UnionFind.prototype.Find.call(receiver, ...args);
}, 1);

UnionFind.prototype.Union$ = function(a, b) {
    [a, b] = [go2jsMethodOn(this, UnionFind, "Find$", [a]), go2jsMethodOn(this, UnionFind, "Find$", [b])];
    if (a == b) {
        return false;
    }
    if (go2jsIndex(this.size, a) < go2jsIndex(this.size, b)) {
        [a, b] = [b, a];
    }
    go2jsSliceStore(this.parent, b, a);
    go2jsIndexCheck(this.size, a);
    this.size[a] = go2jsWideAdd(this.size[a], go2jsIndex(this.size, b), "int");
    return true;
};
UnionFind.prototype.Union = function*(a, b) {
    return UnionFind.prototype.Union$.call(this, a, b);
};
go2jsRegisterMethod("*UnionFind.Union", function*(receiver, ...args) {
    return yield* UnionFind.prototype.Union.call(receiver, ...args);
}, 1);

function* main() {
    let g = NewGraph$();
    for (const e of go2jsRangeCopies(go2jsRangeSequence(go2jsSliceTyped("[]struct{a string; b string; w int}", [(() => {
        const __v = {"a": "A", "b": "B", "w": 7};
        Object.defineProperty(__v, "__go2js_type_name", {value: "struct { a string; b string; w int }"});
        Object.defineProperty(
            __v,
            "__go2js_struct_fields",
            {
                value: [
                    {Name: "a", Type: "string", Anonymous: false},
                    {Name: "b", Type: "string", Anonymous: false},
                    {Name: "w", Type: "int", Anonymous: false}
                ]
            }
        );
        return __v;
    })(), (() => {
        const __v = {"a": "A", "b": "C", "w": 9};
        Object.defineProperty(__v, "__go2js_type_name", {value: "struct { a string; b string; w int }"});
        Object.defineProperty(
            __v,
            "__go2js_struct_fields",
            {
                value: [
                    {Name: "a", Type: "string", Anonymous: false},
                    {Name: "b", Type: "string", Anonymous: false},
                    {Name: "w", Type: "int", Anonymous: false}
                ]
            }
        );
        return __v;
    })(), (() => {
        const __v = {"a": "A", "b": "F", "w": 14};
        Object.defineProperty(__v, "__go2js_type_name", {value: "struct { a string; b string; w int }"});
        Object.defineProperty(
            __v,
            "__go2js_struct_fields",
            {
                value: [
                    {Name: "a", Type: "string", Anonymous: false},
                    {Name: "b", Type: "string", Anonymous: false},
                    {Name: "w", Type: "int", Anonymous: false}
                ]
            }
        );
        return __v;
    })(), (() => {
        const __v = {"a": "B", "b": "C", "w": 10};
        Object.defineProperty(__v, "__go2js_type_name", {value: "struct { a string; b string; w int }"});
        Object.defineProperty(
            __v,
            "__go2js_struct_fields",
            {
                value: [
                    {Name: "a", Type: "string", Anonymous: false},
                    {Name: "b", Type: "string", Anonymous: false},
                    {Name: "w", Type: "int", Anonymous: false}
                ]
            }
        );
        return __v;
    })(), (() => {
        const __v = {"a": "B", "b": "D", "w": 15};
        Object.defineProperty(__v, "__go2js_type_name", {value: "struct { a string; b string; w int }"});
        Object.defineProperty(
            __v,
            "__go2js_struct_fields",
            {
                value: [
                    {Name: "a", Type: "string", Anonymous: false},
                    {Name: "b", Type: "string", Anonymous: false},
                    {Name: "w", Type: "int", Anonymous: false}
                ]
            }
        );
        return __v;
    })(), (() => {
        const __v = {"a": "C", "b": "D", "w": 11};
        Object.defineProperty(__v, "__go2js_type_name", {value: "struct { a string; b string; w int }"});
        Object.defineProperty(
            __v,
            "__go2js_struct_fields",
            {
                value: [
                    {Name: "a", Type: "string", Anonymous: false},
                    {Name: "b", Type: "string", Anonymous: false},
                    {Name: "w", Type: "int", Anonymous: false}
                ]
            }
        );
        return __v;
    })(), (() => {
        const __v = {"a": "C", "b": "F", "w": 2};
        Object.defineProperty(__v, "__go2js_type_name", {value: "struct { a string; b string; w int }"});
        Object.defineProperty(
            __v,
            "__go2js_struct_fields",
            {
                value: [
                    {Name: "a", Type: "string", Anonymous: false},
                    {Name: "b", Type: "string", Anonymous: false},
                    {Name: "w", Type: "int", Anonymous: false}
                ]
            }
        );
        return __v;
    })(), (() => {
        const __v = {"a": "D", "b": "E", "w": 6};
        Object.defineProperty(__v, "__go2js_type_name", {value: "struct { a string; b string; w int }"});
        Object.defineProperty(
            __v,
            "__go2js_struct_fields",
            {
                value: [
                    {Name: "a", Type: "string", Anonymous: false},
                    {Name: "b", Type: "string", Anonymous: false},
                    {Name: "w", Type: "int", Anonymous: false}
                ]
            }
        );
        return __v;
    })(), (() => {
        const __v = {"a": "E", "b": "F", "w": 9};
        Object.defineProperty(__v, "__go2js_type_name", {value: "struct { a string; b string; w int }"});
        Object.defineProperty(
            __v,
            "__go2js_struct_fields",
            {
                value: [
                    {Name: "a", Type: "string", Anonymous: false},
                    {Name: "b", Type: "string", Anonymous: false},
                    {Name: "w", Type: "int", Anonymous: false}
                ]
            }
        );
        return __v;
    })()])))) {
        go2jsMethodOn(g, Graph, "AddEdge$", [e.a, e.b, e.w, false]);
    }
    go2jsPrintln("BFS from A:", go2jsTyped(go2jsMethodOn(g, Graph, "BFS$", ["A"]), "[]string", "", "slice"));
    let [dist, prev] = yield* go2jsMethodOnYield(g, Graph, "Dijkstra", ["A"]);
    for (const n of go2jsRangeSequence(go2jsMethodOn(g, Graph, "Nodes$", []))) {
        go2jsWriteOut(
            process.stdout,
            go2jsSprintf("A to %s: %2d via %s\n", n, go2jsMapGet(dist, n, 0), path$(prev, n))
        );
    }
    let deps = NewGraph$();
    for (const d of go2jsRangeCopies(go2jsRangeSequence(go2jsSliceTyped("[][2]string", [go2jsMarkArray(["shirt", "tie"]), go2jsMarkArray(["tie", "jacket"]), go2jsMarkArray(["trousers", "shoes"]), go2jsMarkArray(["trousers", "belt"]), go2jsMarkArray(["belt", "jacket"]), go2jsMarkArray(["socks", "shoes"]), go2jsMarkArray(["shirt", "belt"])])))) {
        go2jsMethodOn(deps, Graph, "AddEdge$", [go2jsIndex(d, 0), go2jsIndex(d, 1), 1, true]);
    }
    let [order, err] = go2jsMethodOn(deps, Graph, "TopoSort$", []);
    go2jsPrintln("dress order:", go2jsTyped(order, "[]string", "", "slice"), err);
    go2jsMethodOn(deps, Graph, "AddEdge$", ["jacket", "shirt", 1, true]);
    [, err] = go2jsMethodOn(deps, Graph, "TopoSort$", []);
    go2jsPrintln("with cycle:", err);
    class wedge {
        a;
        b;
        w;
        constructor() {
            this.a = 0;
            this.b = 0;
            this.w = 0;
        }
    }
    go2jsRegisterTypeName(
        wedge,
        "main.wedge",
        [],
        [
            {Name: "a", Tag: "", PkgPath: "main", Anonymous: false, Type: "int", Kind: "int"},
            {Name: "b", Tag: "", PkgPath: "main", Anonymous: false, Type: "int", Kind: "int"},
            {Name: "w", Tag: "", PkgPath: "main", Anonymous: false, Type: "int", Kind: "int"}
        ],
        "struct"
    );

    let edges = go2jsSliceTyped(
        "[]main.wedge",
        [
            Object.assign(new wedge(), {a: 0, b: 1, w: 4}),
            Object.assign(new wedge(), {a: 0, b: 2, w: 3}),
            Object.assign(new wedge(), {a: 1, b: 2, w: 1}),
            Object.assign(new wedge(), {a: 1, b: 3, w: 2}),
            Object.assign(new wedge(), {a: 2, b: 3, w: 4}),
            Object.assign(new wedge(), {a: 3, b: 4, w: 2}),
            Object.assign(new wedge(), {a: 4, b: 5, w: 6})
        ]
    );
    go2jsSortSlice(go2jsNilInterface(edges, "[]main.wedge", "slice", true), function*(i, j) {
        return go2jsIndex(edges, i).w < go2jsIndex(edges, j).w;
    });
    let uf = NewUF$(6);
    let total = 0;
    let used = null;
    for (const e of go2jsRangeCopies(go2jsRangeSequence(edges))) {
        if (go2jsMethodOn(uf, UnionFind, "Union$", [e.a, e.b])) {
            total = go2jsWideAdd(total, e.w, "int");
            used = go2jsSliceAppend(used, go2jsSprintf("%d-%d", e.a, e.b));
        }
    }
    go2jsPrintln("MST weight", total, go2jsTyped(used, "[]string", "", "slice"));
}

try {
    go2jsRunMain((function*() {
        return yield* main();
    })());
} catch (e) {
    go2jsReportUncaughtPanic(e);
}
