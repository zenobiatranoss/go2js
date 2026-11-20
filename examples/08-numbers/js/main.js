"use strict";

const {
    go2jsBigInt,
    go2jsBigNewInt,
    go2jsCall,
    go2jsFloat32,
    go2jsFprint,
    go2jsImul,
    go2jsIndex,
    go2jsIndexCheck,
    go2jsIntWrap,
    go2jsLen,
    go2jsMap,
    go2jsMapGet,
    go2jsMapGetOK,
    go2jsMapSet,
    go2jsMapTyped,
    go2jsMathInf,
    go2jsMathNaN,
    go2jsMathRound,
    go2jsMathSin,
    go2jsMod,
    go2jsNamedMethodCall,
    go2jsNew,
    go2jsPrint,
    go2jsPrintln,
    go2jsPtr,
    go2jsRangeMap,
    go2jsRangeSequence,
    go2jsRegisterMethod,
    go2jsReportUncaughtPanic,
    go2jsRunMain,
    go2jsSetPackageName,
    go2jsSetSourceLines,
    go2jsSliceAppend,
    go2jsSliceCopy,
    go2jsSliceMake,
    go2jsSliceRange,
    go2jsSliceStore,
    go2jsSliceTyped,
    go2jsSprintf,
    go2jsStringSlice,
    go2jsStringToBytes,
    go2jsStringsBuilder,
    go2jsTyped,
    go2jsWideAdd,
    go2jsWideMul,
    go2jsWideSub,
    go2jsWideXor,
    go2jsWriteOut,
} = require("./main.runtime.js");

go2jsSetPackageName("main");
go2jsSetSourceLines({2: "08-numbers/main.go:11:6", 3: "08-numbers/main.go:12:15", 4: "08-numbers/main.go:13:2", 5: "08-numbers/main.go:14:2", 6: "08-numbers/main.go:15:3", 7: "08-numbers/main.go:16:13", 8: "08-numbers/main.go:17:4", 9: "08-numbers/main.go:18:5", 13: "08-numbers/main.go:22:2", 20: "08-numbers/main.go:25:6", 21: "08-numbers/main.go:26:2", 22: "08-numbers/main.go:27:3", 24: "08-numbers/main.go:29:2", 31: "08-numbers/main.go:32:6", 32: "08-numbers/main.go:33:2", 33: "08-numbers/main.go:34:2", 34: "08-numbers/main.go:35:2", 35: "08-numbers/main.go:36:3", 38: "08-numbers/main.go:37:4", 41: "08-numbers/main.go:39:24", 42: "08-numbers/main.go:39:24", 43: "08-numbers/main.go:39:24", 44: "08-numbers/main.go:39:24", 45: "08-numbers/main.go:39:24", 46: "08-numbers/main.go:39:24", 47: "08-numbers/main.go:39:24", 48: "08-numbers/main.go:39:24", 49: "08-numbers/main.go:39:24", 50: "08-numbers/main.go:40:3", 52: "08-numbers/main.go:42:2", 59: "08-numbers/main.go:47:6", 60: "08-numbers/main.go:48:7", 61: "08-numbers/main.go:49:2", 62: "08-numbers/main.go:50:10", 64: "08-numbers/main.go:52:2", 71: "08-numbers/main.go:55:17", 72: "08-numbers/main.go:56:27", 73: "08-numbers/main.go:57:2", 74: "08-numbers/main.go:58:3", 75: "08-numbers/main.go:59:4", 76: "08-numbers/main.go:60:5", 81: "08-numbers/main.go:64:2", 86: "08-numbers/main.go:67:17", 87: "08-numbers/main.go:68:30", 88: "08-numbers/main.go:69:2", 89: "08-numbers/main.go:70:3", 90: "08-numbers/main.go:71:4", 93: "08-numbers/main.go:74:2", 98: "08-numbers/main.go:77:17", 100: "08-numbers/main.go:78:2", 101: "08-numbers/main.go:79:2", 102: "08-numbers/main.go:80:3", 103: "08-numbers/main.go:81:4", 104: "08-numbers/main.go:82:5", 106: "08-numbers/main.go:84:4", 107: "08-numbers/main.go:84:4", 108: "08-numbers/main.go:84:4", 109: "08-numbers/main.go:84:4", 110: "08-numbers/main.go:84:4", 111: "08-numbers/main.go:84:4", 113: "08-numbers/main.go:86:3", 115: "08-numbers/main.go:88:9", 120: "08-numbers/main.go:91:6", 121: "08-numbers/main.go:92:7", 122: "08-numbers/main.go:93:7", 123: "08-numbers/main.go:94:2", 124: "08-numbers/main.go:95:3", 126: "08-numbers/main.go:97:2", 127: "08-numbers/main.go:98:2", 128: "08-numbers/main.go:99:3", 129: "08-numbers/main.go:100:3", 130: "08-numbers/main.go:101:27", 131: "08-numbers/main.go:102:5", 134: "08-numbers/main.go:105:3", 135: "08-numbers/main.go:106:4", 137: "08-numbers/main.go:108:3", 138: "08-numbers/main.go:109:4", 141: "08-numbers/main.go:110:4", 143: "08-numbers/main.go:112:3", 144: "08-numbers/main.go:113:3", 145: "08-numbers/main.go:114:4", 146: "08-numbers/main.go:115:4", 147: "08-numbers/main.go:116:5", 152: "08-numbers/main.go:120:2", 159: "08-numbers/main.go:123:6", 160: "08-numbers/main.go:124:7", 161: "08-numbers/main.go:125:32", 162: "08-numbers/main.go:125:32", 163: "08-numbers/main.go:125:32", 164: "08-numbers/main.go:125:32", 165: "08-numbers/main.go:125:32", 166: "08-numbers/main.go:125:32", 167: "08-numbers/main.go:125:32", 168: "08-numbers/main.go:125:32", 169: "08-numbers/main.go:125:32", 170: "08-numbers/main.go:125:32", 171: "08-numbers/main.go:126:39", 172: "08-numbers/main.go:128:7", 173: "08-numbers/main.go:129:2", 174: "08-numbers/main.go:130:9", 175: "08-numbers/main.go:131:31", 176: "08-numbers/main.go:132:30", 177: "08-numbers/main.go:133:56", 178: "08-numbers/main.go:133:56", 179: "08-numbers/main.go:133:56", 180: "08-numbers/main.go:133:56", 181: "08-numbers/main.go:133:56", 182: "08-numbers/main.go:134:14", 183: "08-numbers/main.go:136:2", 189: "08-numbers/main.go:137:7", 190: "08-numbers/main.go:138:12", 191: "08-numbers/main.go:139:2", 198: "08-numbers/main.go:140:27", 199: "08-numbers/main.go:142:2", 200: "08-numbers/main.go:143:2", 201: "08-numbers/main.go:144:2", 202: "08-numbers/main.go:145:2", 203: "08-numbers/main.go:146:2", 204: "08-numbers/main.go:147:2", 205: "08-numbers/main.go:148:2", 206: "08-numbers/main.go:149:2", 207: "08-numbers/main.go:150:2", 208: "08-numbers/main.go:150:2", 209: "08-numbers/main.go:150:2", 210: "08-numbers/main.go:150:2", 211: "08-numbers/main.go:150:2", 212: "08-numbers/main.go:150:2", 213: "08-numbers/main.go:150:2", 214: "08-numbers/main.go:150:2", 215: "08-numbers/main.go:150:2", 216: "08-numbers/main.go:150:2", 217: "08-numbers/main.go:150:2", 218: "08-numbers/main.go:150:2", 219: "08-numbers/main.go:151:94", 220: "08-numbers/main.go:151:94", 221: "08-numbers/main.go:151:94", 222: "08-numbers/main.go:151:94", 223: "08-numbers/main.go:151:94", 224: "08-numbers/main.go:151:94", 225: "08-numbers/main.go:151:94", 226: "08-numbers/main.go:151:94", 227: "08-numbers/main.go:152:66", 228: "08-numbers/main.go:152:66", 229: "08-numbers/main.go:152:66", 230: "08-numbers/main.go:152:66", 231: "08-numbers/main.go:152:66", 232: "08-numbers/main.go:152:66", 233: "08-numbers/main.go:152:66", 234: "08-numbers/main.go:152:66", 235: "08-numbers/main.go:152:66", 236: "08-numbers/main.go:152:66", 237: "08-numbers/main.go:152:66", 238: "08-numbers/main.go:152:66", 239: "08-numbers/main.go:152:66", 240: "08-numbers/main.go:152:66", 241: "08-numbers/main.go:152:66", 242: "08-numbers/main.go:153:36", 243: "08-numbers/main.go:153:36", 244: "08-numbers/main.go:153:36", 245: "08-numbers/main.go:153:36", 246: "08-numbers/main.go:153:36", 247: "08-numbers/main.go:153:36", 248: "08-numbers/main.go:153:36", 249: "08-numbers/main.go:153:36", 250: "08-numbers/main.go:154:2", 251: "08-numbers/main.go:155:20", 252: "08-numbers/main.go:156:8", 253: "08-numbers/main.go:157:3", 255: "08-numbers/main.go:159:2"});

function sieve$(n) {
    let composite = go2jsSliceMake(go2jsWideAdd(n, 1, "int"), go2jsWideAdd(n, 1, "int"), false, true);
    let primes = null;
    for (let i = 2; i <= n; i = go2jsWideAdd(i, 1, "int")) {
        if (!go2jsIndex(composite, i)) {
            primes = go2jsSliceAppend(primes, i);
            for (let j = go2jsWideMul(i, i, "int"); j <= n; j = go2jsWideAdd(j, i, "int")) {
                go2jsSliceStore(composite, j, true);
            }
        }
    }
    return primes;
}

function* sieve(n) {
    return sieve$(n);
}

function gcd$(a, b) {
    for (; b != 0;) {
        [a, b] = [b, go2jsMod(a, b)];
    }
    return a;
}

function* gcd(a, b) {
    return gcd$(a, b);
}

function fibMemo$() {
    let memo = go2jsMapTyped("map[int]uint64", go2jsMap([[0, 0], [1, 1]]));
    let fib = null;
    fib = function*(n) {
        {
            let [v, ok] = go2jsMapGetOK(memo, n, 0)
            if (ok) {
                return v;
            }
        }
        go2jsMapSet(
            memo,
            n,
            go2jsWideAdd(
                yield* go2jsCall(fib, null, [go2jsWideSub(n, 1, "int")]),
                yield* go2jsCall(fib, null, [go2jsWideSub(n, 2, "int")]),
                "uint64"
            )
        );
        return go2jsMapGet(memo, n, 0);
    };
    return fib;
}

function* fibMemo() {
    return fibMemo$();
}

function NewMatrix$(r, c) {
    let m = go2jsSliceMake(r, r, null, true);
    for (const i of go2jsRangeMap(m).keys()) {
        go2jsSliceStore(m, i, go2jsSliceMake(c, c, 0, true));
    }
    return m;
}

function* NewMatrix(r, c) {
    return NewMatrix$(r, c);
}

function* go2jsMethodMatrixMul(m, o) {
    let out = NewMatrix$(go2jsLen(m), go2jsLen(go2jsIndex(o, 0)));
    for (const i of go2jsRangeMap(m).keys()) {
        for (const j of go2jsRangeMap(go2jsIndex(o, 0)).keys()) {
            for (const k of go2jsRangeMap(o).keys()) {
                go2jsIndexCheck(go2jsIndex(out, i), j);
                out[i][j] += go2jsIndex(go2jsIndex(m, i), k) * go2jsIndex(go2jsIndex(o, k), j);
            }
        }
    }
    return out;
}

go2jsRegisterMethod("Matrix.Mul", go2jsMethodMatrixMul);

function* go2jsMethodMatrixTranspose(m) {
    let out = NewMatrix$(go2jsLen(go2jsIndex(m, 0)), go2jsLen(m));
    for (const i of go2jsRangeMap(m).keys()) {
        for (const j of go2jsRangeMap(go2jsIndex(m, i)).keys()) {
            go2jsSliceStore(go2jsIndex(out, j), i, go2jsIndex(go2jsIndex(m, i), j));
        }
    }
    return out;
}

go2jsRegisterMethod("Matrix.Transpose", go2jsMethodMatrixTranspose);

function* go2jsMethodMatrixString(m) {
    let addr1 = null;
    let sb = new go2jsStringsBuilder();
    for (const row of go2jsRangeSequence(m)) {
        for (const [j, v] of go2jsRangeSequence(row).entries()) {
            if (j > 0) {
                sb.WriteByte(32);
            }
            go2jsFprint(
                (addr1 ??= go2jsPtr(() => sb, value => sb = value)),
                [go2jsSprintf("%6.1f", go2jsTyped(v, "float64"))],
                "",
                ""
            );
        }
        sb.WriteByte(10);
    }
    return sb.String();
}

go2jsRegisterMethod("Matrix.String", go2jsMethodMatrixString);

function det$(m) {
    let n = go2jsLen(m);
    let a = NewMatrix$(n, n);
    for (const i of go2jsRangeMap(m).keys()) {
        go2jsSliceCopy(go2jsIndex(a, i), go2jsIndex(m, i));
    }
    let d = 1.0;
    for (let c = 0; c < n; c = go2jsWideAdd(c, 1, "int")) {
        let p = c;
        for (let r = go2jsWideAdd(c, 1, "int"); r < n; r = go2jsWideAdd(r, 1, "int")) {
            if (Math.abs(go2jsIndex(go2jsIndex(a, r), c)) > Math.abs(go2jsIndex(go2jsIndex(a, p), c))) {
                p = r;
            }
        }
        if (go2jsIndex(go2jsIndex(a, p), c) == 0) {
            return 0;
        }
        if (p != c) {
            go2jsIndexCheck(a, p);
            go2jsIndexCheck(a, c);
            [a[p], a[c]] = [go2jsIndex(a, c), go2jsIndex(a, p)];
            d = -d;
        }
        d *= go2jsIndex(go2jsIndex(a, c), c);
        for (let r = go2jsWideAdd(c, 1, "int"); r < n; r = go2jsWideAdd(r, 1, "int")) {
            let f = go2jsIndex(go2jsIndex(a, r), c) / go2jsIndex(go2jsIndex(a, c), c);
            for (let k = c; k < n; k = go2jsWideAdd(k, 1, "int")) {
                go2jsIndexCheck(go2jsIndex(a, r), k);
                a[r][k] -= f * go2jsIndex(go2jsIndex(a, c), k);
            }
        }
    }
    return d;
}

function* det(m) {
    return det$(m);
}

function* main() {
    let p = sieve$(100);
    go2jsPrintln(
        go2jsLen(p),
        go2jsTyped(go2jsSliceRange(p, undefined, 10, undefined, false), "[]int", "", "slice"),
        go2jsTyped(
            go2jsSliceRange(p, go2jsWideSub(go2jsLen(p), 3, "int"), undefined, undefined, false),
            "[]int",
            "",
            "slice"
        )
    );
    go2jsPrintln(gcd$(84, 36), gcd$(17, 5), go2jsTyped(yield* go2jsCall(fibMemo$(), null, [90]), "uint64"));
    let f = go2jsNew(new go2jsBigInt(), "Int").MulRange(1, 30);
    go2jsPrintln("30! =", go2jsTyped(f, "*big.Int", "", "pointer"));
    let two = go2jsBigNewInt(2);
    let pow = go2jsNew(new go2jsBigInt(), "Int").Exp(two, go2jsBigNewInt(200), null);
    go2jsPrintln("2^200 =", go2jsTyped(pow, "*big.Int", "", "pointer"), pow.BitLen());
    let [q, r] = go2jsNew(new go2jsBigInt(), "Int").DivMod(
        pow,
        go2jsBigNewInt(1000003),
        go2jsNew(new go2jsBigInt(), "Int")
    );
    go2jsPrintln(go2jsStringSlice(q.String(), undefined, 12), go2jsTyped(r, "*big.Int", "", "pointer"));
    let a = yield* (function*() {
        const literal2 = [null, null];
        literal2[0] = go2jsSliceTyped("[]float64", [1, 2, 3]);
        literal2[1] = go2jsSliceTyped("[]float64", [4, 5, 6]);
        return literal2;
    })();
    let b = go2jsNamedMethodCall(a, "Matrix", "Transpose");
    go2jsPrint(yield* go2jsMethodMatrixString(go2jsNamedMethodCall(a, "Matrix", "Mul", b)));
    let sq = yield* (function*() {
        const literal3 = [null, null, null];
        literal3[0] = go2jsSliceTyped("[]float64", [2, -1, 0]);
        literal3[1] = go2jsSliceTyped("[]float64", [-1, 2, -1]);
        literal3[2] = go2jsSliceTyped("[]float64", [0, -1, 2]);
        return literal3;
    })();
    go2jsWriteOut(process.stdout, go2jsSprintf("det=%.3f\n", go2jsTyped(det$(sq), "float64")));
    let u8 = 250;
    u8 = go2jsIntWrap(u8 + 10, "uint8");
    let i8 = 127;
    i8 = go2jsIntWrap(i8 + 1, "int8");
    let u32 = 2147483648;
    u32 = go2jsIntWrap(go2jsImul(u32, 2), "uint32");
    let i64 = 9223372036854775807n;
    i64 = go2jsWideAdd(i64, 1, "int64");
    go2jsPrintln(
        go2jsTyped(u8, "uint8"),
        go2jsTyped(i8, "int8"),
        go2jsTyped(u32, "uint32"),
        go2jsTyped(i64, "int64"),
        -3,
        -1,
        3,
        -4,
        4,
        5
    );
    go2jsPrintln(
        go2jsTyped(Math.sqrt(2), "float64"),
        go2jsTyped(Math.floor(-2.5), "float64"),
        go2jsTyped(go2jsMathRound(2.5), "float64"),
        go2jsTyped(Math.trunc(-2.7), "float64"),
        go2jsTyped(go2jsMathInf(1), "float64"),
        go2jsTyped(go2jsMathNaN(), "float64")
    );
    go2jsWriteOut(
        process.stdout,
        go2jsSprintf(
            "%.4f %.4f %e %08.3f %x %o %b %c %U\n",
            go2jsTyped(go2jsMathSin(1), "float64"),
            go2jsTyped(Math.log(10), "float64"),
            go2jsTyped(123456.789, "float64"),
            go2jsTyped(3.14159, "float64"),
            255,
            64,
            10,
            go2jsTyped(71, "rune"),
            go2jsTyped(71, "rune")
        )
    );
    go2jsPrintln(
        go2jsTyped(0.3, "float64"),
        go2jsTyped(go2jsFloat32(go2jsFloat32(0.1) + go2jsFloat32(0.2)), "float32"),
        go2jsTyped(1e21, "float64"),
        go2jsTyped(1e20, "float64"),
        go2jsTyped(100000000.0, "float64"),
        go2jsTyped(2.5, "float64")
    );
    let h = 14695981039346656037n;
    for (const c of go2jsRangeSequence(go2jsStringToBytes("hello"))) {
        h = go2jsWideXor(h, Math.trunc(c), "uint64");
        h = go2jsWideMul(h, 1099511628211, "uint64");
    }
    go2jsPrintln("fnv1a:", go2jsTyped(h, "uint64"));
}

try {
    go2jsRunMain((function*() {
        return yield* main();
    })());
} catch (e) {
    go2jsReportUncaughtPanic(e);
}
