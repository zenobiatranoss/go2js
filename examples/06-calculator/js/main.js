"use strict";

const {
    go2jsCall,
    go2jsDeferInit,
    go2jsEqual,
    go2jsIndex,
    go2jsInterface,
    go2jsInterfaceCall,
    go2jsLen,
    go2jsMap,
    go2jsMapGetOK,
    go2jsMapTyped,
    go2jsMarkArray,
    go2jsMathMod,
    go2jsMethodOn,
    go2jsMethodOnYield,
    go2jsNew,
    go2jsPreempt,
    go2jsRangeSequence,
    go2jsRegisterInterface,
    go2jsRegisterMethod,
    go2jsRegisterStructFormat,
    go2jsRegisterTypeName,
    go2jsRegisterValueMethods,
    go2jsReportUncaughtPanic,
    go2jsRunInitializers,
    go2jsRunMain,
    go2jsRuneString,
    go2jsRunesToString,
    go2jsSetPackageName,
    go2jsSetSourceLines,
    go2jsSliceAppend,
    go2jsSliceMake,
    go2jsSliceRange,
    go2jsSliceStore,
    go2jsSliceTyped,
    go2jsSprintf,
    go2jsState,
    go2jsStrconvFormatFloat,
    go2jsStrconvParseFloat,
    go2jsStringToRunes,
    go2jsStringsContainsRune,
    go2jsStringsJoin,
    go2jsStructCopy,
    go2jsTyped,
    go2jsUnicodeIsDigit,
    go2jsUnicodeIsLetter,
    go2jsUnicodeIsSpace,
    go2jsWideAdd,
    go2jsWideWrap,
    go2jsWrapError,
    go2jsWriteOut,
} = require("./main.runtime.js");

go2jsSetPackageName("main");
go2jsSetSourceLines({9: "06-calculator/main.go:24:20", 10: "06-calculator/main.go:25:2", 39: "06-calculator/main.go:34:6", 40: "06-calculator/main.go:35:2", 41: "06-calculator/main.go:36:8", 42: "06-calculator/main.go:37:18", 43: "06-calculator/main.go:38:3", 44: "06-calculator/main.go:39:3", 45: "06-calculator/main.go:40:8", 47: "06-calculator/main.go:41:4", 50: "06-calculator/main.go:42:8", 52: "06-calculator/main.go:43:4", 53: "06-calculator/main.go:44:24", 57: "06-calculator/main.go:45:5", 59: "06-calculator/main.go:47:35", 60: "06-calculator/main.go:47:35", 61: "06-calculator/main.go:47:35", 62: "06-calculator/main.go:47:35", 63: "06-calculator/main.go:47:35", 64: "06-calculator/main.go:47:35", 65: "06-calculator/main.go:47:35", 66: "06-calculator/main.go:47:35", 67: "06-calculator/main.go:47:35", 68: "06-calculator/main.go:47:35", 69: "06-calculator/main.go:47:35", 72: "06-calculator/main.go:48:8", 74: "06-calculator/main.go:49:4", 75: "06-calculator/main.go:50:51", 79: "06-calculator/main.go:51:5", 81: "06-calculator/main.go:53:37", 82: "06-calculator/main.go:53:37", 83: "06-calculator/main.go:53:37", 84: "06-calculator/main.go:53:37", 85: "06-calculator/main.go:53:37", 86: "06-calculator/main.go:53:37", 87: "06-calculator/main.go:53:37", 88: "06-calculator/main.go:53:37", 89: "06-calculator/main.go:53:37", 90: "06-calculator/main.go:53:37", 91: "06-calculator/main.go:53:37", 94: "06-calculator/main.go:54:8", 96: "06-calculator/main.go:55:34", 97: "06-calculator/main.go:55:34", 98: "06-calculator/main.go:55:34", 99: "06-calculator/main.go:55:34", 100: "06-calculator/main.go:56:4", 105: "06-calculator/main.go:58:11", 106: "06-calculator/main.go:59:4", 111: "06-calculator/main.go:61:11", 112: "06-calculator/main.go:62:4", 117: "06-calculator/main.go:64:11", 118: "06-calculator/main.go:65:4", 123: "06-calculator/main.go:67:16", 128: "06-calculator/main.go:70:37", 129: "06-calculator/main.go:70:37", 130: "06-calculator/main.go:70:37", 131: "06-calculator/main.go:70:37", 204: "06-calculator/main.go:93:14", 205: "06-calculator/main.go:93:65", 210: "06-calculator/main.go:94:14", 211: "06-calculator/main.go:94:85", 216: "06-calculator/main.go:95:14", 217: "06-calculator/main.go:96:2", 218: "06-calculator/main.go:96:18", 220: "06-calculator/main.go:97:3", 223: "06-calculator/main.go:99:48", 228: "06-calculator/main.go:101:14", 229: "06-calculator/main.go:101:39", 234: "06-calculator/main.go:102:16", 235: "06-calculator/main.go:103:12", 236: "06-calculator/main.go:104:2", 242: "06-calculator/main.go:106:16", 243: "06-calculator/main.go:106:54", 249: "06-calculator/main.go:107:17", 250: "06-calculator/main.go:108:12", 251: "06-calculator/main.go:109:2", 252: "06-calculator/main.go:110:3", 254: "06-calculator/main.go:112:12", 257: "06-calculator/main.go:113:2", 258: "06-calculator/main.go:114:3", 260: "06-calculator/main.go:116:2", 263: "06-calculator/main.go:118:3", 268: "06-calculator/main.go:120:3", 273: "06-calculator/main.go:122:3", 278: "06-calculator/main.go:124:3", 279: "06-calculator/main.go:125:14", 281: "06-calculator/main.go:127:3", 286: "06-calculator/main.go:129:10", 291: "06-calculator/main.go:131:10", 295: "06-calculator/main.go:133:12", 301: "06-calculator/main.go:135:17", 302: "06-calculator/main.go:135:82", 303: "06-calculator/main.go:135:82", 304: "06-calculator/main.go:135:82", 315: "06-calculator/main.go:138:54", 317: "06-calculator/main.go:140:3", 318: "06-calculator/main.go:141:3", 319: "06-calculator/main.go:142:8", 321: "06-calculator/main.go:144:3", 323: "06-calculator/main.go:146:53", 326: "06-calculator/main.go:149:15", 327: "06-calculator/main.go:150:2", 328: "06-calculator/main.go:151:2", 329: "06-calculator/main.go:152:13", 331: "06-calculator/main.go:154:2", 332: "06-calculator/main.go:155:2", 333: "06-calculator/main.go:156:13", 334: "06-calculator/main.go:157:3", 335: "06-calculator/main.go:158:4", 337: "06-calculator/main.go:160:10", 339: "06-calculator/main.go:162:5", 340: "06-calculator/main.go:163:13", 342: "06-calculator/main.go:165:9", 348: "06-calculator/main.go:167:15", 349: "06-calculator/main.go:168:26", 350: "06-calculator/main.go:169:2", 351: "06-calculator/main.go:170:14", 353: "06-calculator/main.go:172:24", 380: "06-calculator/main.go:180:18", 381: "06-calculator/main.go:180:33", 390: "06-calculator/main.go:181:18", 391: "06-calculator/main.go:181:33", 392: "06-calculator/main.go:181:53", 393: "06-calculator/main.go:181:62", 409: "06-calculator/main.go:185:18", 410: "06-calculator/main.go:186:15", 411: "06-calculator/main.go:187:2", 412: "06-calculator/main.go:188:3", 414: "06-calculator/main.go:190:2", 418: "06-calculator/main.go:191:8", 419: "06-calculator/main.go:192:3", 420: "06-calculator/main.go:193:3", 421: "06-calculator/main.go:194:4", 423: "06-calculator/main.go:196:3", 424: "06-calculator/main.go:197:3", 425: "06-calculator/main.go:198:3", 426: "06-calculator/main.go:199:4", 428: "06-calculator/main.go:201:17", 431: "06-calculator/main.go:202:3", 432: "06-calculator/main.go:203:4", 434: "06-calculator/main.go:205:3", 435: "06-calculator/main.go:205:3", 436: "06-calculator/main.go:205:3", 437: "06-calculator/main.go:205:3", 438: "06-calculator/main.go:205:3", 445: "06-calculator/main.go:209:18", 446: "06-calculator/main.go:210:2", 447: "06-calculator/main.go:210:10", 449: "06-calculator/main.go:211:3", 450: "06-calculator/main.go:212:13", 451: "06-calculator/main.go:213:3", 454: "06-calculator/main.go:215:9", 460: "06-calculator/main.go:218:18", 461: "06-calculator/main.go:219:7", 462: "06-calculator/main.go:220:2", 465: "06-calculator/main.go:222:13", 466: "06-calculator/main.go:223:3", 467: "06-calculator/main.go:224:16", 469: "06-calculator/main.go:226:10", 474: "06-calculator/main.go:228:6", 475: "06-calculator/main.go:229:11", 477: "06-calculator/main.go:231:3", 478: "06-calculator/main.go:232:3", 479: "06-calculator/main.go:233:7", 483: "06-calculator/main.go:234:16", 484: "06-calculator/main.go:235:4", 485: "06-calculator/main.go:236:5", 487: "06-calculator/main.go:238:16", 488: "06-calculator/main.go:239:7", 489: "06-calculator/main.go:240:5", 491: "06-calculator/main.go:242:53", 492: "06-calculator/main.go:242:53", 493: "06-calculator/main.go:242:53", 494: "06-calculator/main.go:242:53", 495: "06-calculator/main.go:242:53", 496: "06-calculator/main.go:242:53", 497: "06-calculator/main.go:242:53", 500: "06-calculator/main.go:245:3", 501: "06-calculator/main.go:246:3", 506: "06-calculator/main.go:248:13", 507: "06-calculator/main.go:249:3", 508: "06-calculator/main.go:250:4", 510: "06-calculator/main.go:252:6", 511: "06-calculator/main.go:253:16", 513: "06-calculator/main.go:255:3", 517: "06-calculator/main.go:257:14", 523: "06-calculator/main.go:260:6", 524: "06-calculator/main.go:261:15", 525: "06-calculator/main.go:262:2", 526: "06-calculator/main.go:263:3", 528: "06-calculator/main.go:265:2", 529: "06-calculator/main.go:266:15", 532: "06-calculator/main.go:267:2", 533: "06-calculator/main.go:268:3", 535: "06-calculator/main.go:270:5", 536: "06-calculator/main.go:271:61", 537: "06-calculator/main.go:271:61", 538: "06-calculator/main.go:271:61", 539: "06-calculator/main.go:271:61", 540: "06-calculator/main.go:271:61", 541: "06-calculator/main.go:271:61", 542: "06-calculator/main.go:271:61", 543: "06-calculator/main.go:271:61", 545: "06-calculator/main.go:273:12", 548: "06-calculator/main.go:274:2", 549: "06-calculator/main.go:275:3", 551: "06-calculator/main.go:277:9", 554: "06-calculator/main.go:280:6", 555: "06-calculator/main.go:281:2", 556: "06-calculator/main.go:282:2", 557: "06-calculator/main.go:298:15", 558: "06-calculator/main.go:299:3", 559: "06-calculator/main.go:300:4", 560: "06-calculator/main.go:301:4", 562: "06-calculator/main.go:303:3"});
const NUM = 0;
const IDENT = 1;
const OP = 2;
const LPAREN = 3;
const RPAREN = 4;
const COMMA = 5;
const EOF = 6;

function* TokenKindString(k) {
    return go2jsIndex(go2jsMarkArray(["NUM", "IDENT", "OP", "LPAREN", "RPAREN", "COMMA", "EOF"]), k);
}

go2jsRegisterMethod("TokenKind.String", TokenKindString);

class Token {
    Kind;
    Text;
    Pos;
    constructor() {
        this.Kind = 0;
        this.Text = "";
        this.Pos = 0;
    }
}

go2jsRegisterStructFormat("Token", {"Kind": "TokenKind.String"});
go2jsRegisterTypeName(
    Token,
    "main.Token",
    [],
    [
        {Name: "Kind", Tag: "", PkgPath: "", Anonymous: false, Type: "main.TokenKind", Kind: "int"},
        {Name: "Text", Tag: "", PkgPath: "", Anonymous: false, Type: "string", Kind: "string"},
        {Name: "Pos", Tag: "", PkgPath: "", Anonymous: false, Type: "int", Kind: "int"}
    ],
    "struct"
);

function* lex(src) {
    let toks = null;
    let rs = go2jsStringToRunes(src);
    for (let i = 0; i < go2jsLen(rs);) {
        let r = go2jsIndex(rs, i);
        switch (true) {
            case go2jsUnicodeIsSpace(r):
                {
                i = go2jsWideAdd(i, 1, "int");
            }
                break;
            case go2jsUnicodeIsDigit(r) || r == 46:
                {
                let start = i;
                for (; i < go2jsLen(rs) && (go2jsUnicodeIsDigit(go2jsIndex(rs, i)) || go2jsIndex(rs, i) == 46);) {
                    if (++go2jsState.go2jsTicks > 20000) {
                        yield* go2jsPreempt();
                    }
                    i = go2jsWideAdd(i, 1, "int");
                }
                toks = go2jsSliceAppend(
                    toks,
                    Object.assign(
                        new Token(),
                        {
                            Kind: NUM,
                            Text: go2jsRunesToString(go2jsSliceRange(rs, start, i, undefined, false)),
                            Pos: start
                        }
                    )
                );
            }
                break;
            case go2jsUnicodeIsLetter(r):
                {
                let start = i;
                for (; i < go2jsLen(rs) && (go2jsUnicodeIsLetter(go2jsIndex(rs, i)) || go2jsUnicodeIsDigit(go2jsIndex(rs, i)));) {
                    if (++go2jsState.go2jsTicks > 20000) {
                        yield* go2jsPreempt();
                    }
                    i = go2jsWideAdd(i, 1, "int");
                }
                toks = go2jsSliceAppend(
                    toks,
                    Object.assign(
                        new Token(),
                        {
                            Kind: IDENT,
                            Text: go2jsRunesToString(go2jsSliceRange(rs, start, i, undefined, false)),
                            Pos: start
                        }
                    )
                );
            }
                break;
            case go2jsStringsContainsRune("+-*/^%", r):
                {
                toks = go2jsSliceAppend(
                    toks,
                    Object.assign(new Token(), {Kind: OP, Text: go2jsRuneString(r), Pos: i})
                );
                i = go2jsWideAdd(i, 1, "int");
            }
                break;
            case r == 40:
                {
                toks = go2jsSliceAppend(toks, Object.assign(new Token(), {Kind: LPAREN, Text: "(", Pos: i}));
                i = go2jsWideAdd(i, 1, "int");
            }
                break;
            case r == 41:
                {
                toks = go2jsSliceAppend(toks, Object.assign(new Token(), {Kind: RPAREN, Text: ")", Pos: i}));
                i = go2jsWideAdd(i, 1, "int");
            }
                break;
            case r == 44:
                {
                toks = go2jsSliceAppend(toks, Object.assign(new Token(), {Kind: COMMA, Text: ",", Pos: i}));
                i = go2jsWideAdd(i, 1, "int");
            }
                break;
            default:
                {
                return [null, go2jsWrapError("unexpected %q at %d", r, i)];
            }
                break;
        }
    }
    return [
        go2jsSliceAppend(toks, Object.assign(new Token(), {Kind: EOF, Text: "", Pos: go2jsLen(rs)})),
        null
    ];
}

class Node {}

go2jsRegisterInterface("Node", ["Eval", "String"]);

class Unary {
    Op;
    X;
    constructor() {
        this.Op = "";
        this.X = null;
    }
}

go2jsRegisterTypeName(
    Unary,
    "main.Unary",
    ["Eval", "String"],
    [
        {Name: "Op", Tag: "", PkgPath: "", Anonymous: false, Type: "string", Kind: "string"},
        {Name: "X", Tag: "", PkgPath: "", Anonymous: false, Type: "main.Node", Kind: "interface"}
    ],
    "struct"
);
go2jsRegisterValueMethods("main.Unary", ["Eval", "String"]);

class Binary {
    Op;
    L;
    R;
    constructor() {
        this.Op = "";
        this.L = null;
        this.R = null;
    }
}

go2jsRegisterTypeName(
    Binary,
    "main.Binary",
    ["Eval", "String"],
    [
        {Name: "Op", Tag: "", PkgPath: "", Anonymous: false, Type: "string", Kind: "string"},
        {Name: "L", Tag: "", PkgPath: "", Anonymous: false, Type: "main.Node", Kind: "interface"},
        {Name: "R", Tag: "", PkgPath: "", Anonymous: false, Type: "main.Node", Kind: "interface"}
    ],
    "struct"
);
go2jsRegisterValueMethods("main.Binary", ["Eval", "String"]);

class Call {
    Name;
    Args;
    constructor() {
        this.Name = "";
        this.Args = null;
    }
}

go2jsRegisterTypeName(
    Call,
    "main.Call",
    ["Eval", "String"],
    [
        {Name: "Name", Tag: "", PkgPath: "", Anonymous: false, Type: "string", Kind: "string"},
        {Name: "Args", Tag: "", PkgPath: "", Anonymous: false, Type: "[]main.Node", Kind: "slice"}
    ],
    "struct"
);
go2jsRegisterValueMethods("main.Call", ["Eval", "String"]);

function* NumEval(n) {
    return [n, null];
}

go2jsRegisterMethod("Num.Eval", NumEval);

function* NumString(n) {
    return go2jsStrconvFormatFloat(n, 103, go2jsWideWrap(-1, "int"), 64);
}

go2jsRegisterMethod("Num.String", NumString);

function* VarEval(v, env) {
    {
        let [x, ok] = go2jsMapGetOK(env, v, 0)
        if (ok) {
            return [x, null];
        }
    }
    return [0, go2jsWrapError("undefined variable %s", v)];
}

go2jsRegisterMethod("Var.Eval", VarEval);

function* VarString(v) {
    return v;
}

go2jsRegisterMethod("Var.String", VarString);

Unary.prototype.Eval = function*(env) {
    let [x, err] = yield* go2jsInterfaceCall(this.X, "Eval", env);
    return [-x, err];
}
go2jsRegisterMethod("Unary.Eval", function*(receiver, ...args) {
    return yield* Unary.prototype.Eval.call(receiver, ...args);
}, 2);

Unary.prototype.String = function*() {
    return "(" + this.Op + (yield* go2jsInterfaceCall(this.X, "String")) + ")";
}
go2jsRegisterMethod("Unary.String", function*(receiver, ...args) {
    return yield* Unary.prototype.String.call(receiver, ...args);
}, 1);

Binary.prototype.Eval = function*(env) {
    let [l, err] = yield* go2jsInterfaceCall(this.L, "Eval", env);
    if (go2jsEqual(err, null) === false) {
        return [0, err];
    }
    const [go2jsM1_0, go2jsM1_1] = yield* go2jsInterfaceCall(this.R, "Eval", env);
    let r = go2jsM1_0;
    err = go2jsM1_1;
    if (go2jsEqual(err, null) === false) {
        return [0, err];
    }
    switch (this.Op) {
        case "+":
            {
            return [l + r, null];
        }
            break;
        case "-":
            {
            return [l - r, null];
        }
            break;
        case "*":
            {
            return [l * r, null];
        }
            break;
        case "/":
            {
            if (r == 0) {
                return [0, go2jsWrapError("division by zero")];
            }
            return [l / r, null];
        }
            break;
        case "%":
            {
            return [go2jsMathMod(l, r), null];
        }
            break;
        case "^":
            {
            return [Math.pow(l, r), null];
        }
            break;
    }
    return [0, go2jsWrapError("bad operator %s", this.Op)];
}
go2jsRegisterMethod("Binary.Eval", function*(receiver, ...args) {
    return yield* Binary.prototype.Eval.call(receiver, ...args);
}, 2);

Binary.prototype.String = function*() {
    return "(" + (yield* go2jsInterfaceCall(this.L, "String")) + " " + this.Op + " " + (
        yield* go2jsInterfaceCall(this.R, "String")
    ) + ")";
}
go2jsRegisterMethod("Binary.String", function*(receiver, ...args) {
    return yield* Binary.prototype.String.call(receiver, ...args);
}, 1);

let funcs;
go2jsDeferInit(function*() {
    funcs = go2jsMapTyped("map[string]func(args []float64) (float64, error)", go2jsMap([["sqrt", function*(
        a
    ) {
        return [Math.sqrt(go2jsIndex(a, 0)), null];
    }], ["max", function*(a) {
        let m = go2jsIndex(a, 0);
        for (const x of go2jsRangeSequence(go2jsSliceRange(a, 1, undefined, undefined, false))) {
            m = Math.max(m, x);
        }
        return [m, null];
    }], ["abs", function*(a) {
        return [Math.abs(go2jsIndex(a, 0)), null];
    }]]));
}, 0);
Call.prototype.Eval = function*(env) {
    let [f, ok] = go2jsMapGetOK(funcs, this.Name, null);
    if (!ok) {
        return [0, go2jsWrapError("unknown function %s", this.Name)];
    }
    let args = null;
    for (const a of go2jsRangeSequence(this.Args)) {
        let [v, err] = yield* go2jsInterfaceCall(a, "Eval", env);
        if (go2jsEqual(err, null) === false) {
            return [0, err];
        }
        args = go2jsSliceAppend(args, v);
    }
    if (go2jsLen(args) == 0) {
        return [0, go2jsWrapError("%s needs arguments", this.Name)];
    }
    return yield* go2jsCall(f, null, [args]);
}
go2jsRegisterMethod("Call.Eval", function*(receiver, ...args) {
    return yield* Call.prototype.Eval.call(receiver, ...args);
}, 2);

Call.prototype.String = function*() {
    let parts = go2jsSliceMake(go2jsLen(this.Args), go2jsLen(this.Args), "", true);
    for (const [i, a] of go2jsRangeSequence(this.Args).entries()) {
        go2jsSliceStore(parts, i, yield* go2jsInterfaceCall(a, "String"));
    }
    return this.Name + "(" + go2jsStringsJoin(parts, ", ") + ")";
}
go2jsRegisterMethod("Call.String", function*(receiver, ...args) {
    return yield* Call.prototype.String.call(receiver, ...args);
}, 1);

class Parser {
    toks;
    pos;
    constructor() {
        this.toks = null;
        this.pos = 0;
    }
}

go2jsRegisterTypeName(
    Parser,
    "main.Parser",
    ["expr", "next", "peek", "primary", "unary"],
    [
        {Name: "toks", Tag: "", PkgPath: "main", Anonymous: false, Type: "[]main.Token", Kind: "slice"},
        {Name: "pos", Tag: "", PkgPath: "main", Anonymous: false, Type: "int", Kind: "int"}
    ],
    "struct"
);
go2jsRegisterValueMethods("main.Parser", []);

Parser.prototype.peek$ = function() {
    return go2jsStructCopy(go2jsIndex(this.toks, this.pos));
};
Parser.prototype.peek = function*() {
    return Parser.prototype.peek$.call(this);
};
go2jsRegisterMethod("*Parser.peek", function*(receiver, ...args) {
    return yield* Parser.prototype.peek.call(receiver, ...args);
}, 1);

Parser.prototype.next$ = function() {
    let t = go2jsStructCopy(go2jsIndex(this.toks, this.pos));
    this.pos = go2jsWideAdd(this.pos, 1, "int");
    return go2jsStructCopy(t);
};
Parser.prototype.next = function*() {
    return Parser.prototype.next$.call(this);
};
go2jsRegisterMethod("*Parser.next", function*(receiver, ...args) {
    return yield* Parser.prototype.next.call(receiver, ...args);
}, 1);

let prec;
go2jsDeferInit(function*() {
    prec = go2jsMapTyped(
        "map[string]int",
        go2jsMap([["+", 1], ["-", 1], ["*", 2], ["/", 2], ["%", 2], ["^", 3]])
    );
}, 1);
Parser.prototype.expr = function*(minPrec) {
    let [left, err] = yield* go2jsMethodOnYield(this, Parser, "unary", []);
    if (go2jsEqual(err, null) === false) {
        return [null, err];
    }
    for (;;) {
        if (++go2jsState.go2jsTicks > 20000) {
            yield* go2jsPreempt();
        }
        let t = go2jsStructCopy(go2jsMethodOn(this, Parser, "peek$", []));
        let [pr, ok] = go2jsMapGetOK(prec, t.Text, 0);
        if (t.Kind != OP || !ok || pr < minPrec) {
            return [left, null];
        }
        go2jsMethodOn(this, Parser, "next$", []);
        let nextMin = go2jsWideAdd(pr, 1, "int");
        if (t.Text == "^") {
            nextMin = pr;
        }
        const [go2jsM2_0, go2jsM2_1] = yield* go2jsMethodOnYield(this, Parser, "expr", [nextMin]);
        let right = go2jsM2_0;
        let err$1 = go2jsM2_1;
        if (go2jsEqual(err$1, null) === false) {
            return [null, err$1];
        }
        left = go2jsInterface(
            Object.assign(new Binary(), {Op: t.Text, L: left, R: right}),
            "Binary",
            "main.Binary"
        );
    }
}
go2jsRegisterMethod("*Parser.expr", function*(receiver, ...args) {
    return yield* Parser.prototype.expr.call(receiver, ...args);
}, 2);

Parser.prototype.unary = function*() {
    {
        let t = go2jsMethodOn(this, Parser, "peek$", [])
        if (t.Kind == OP && t.Text == "-") {
            go2jsMethodOn(this, Parser, "next$", []);
            let [x, err] = yield* go2jsMethodOnYield(this, Parser, "unary", []);
            return [go2jsInterface(Object.assign(new Unary(), {Op: "-", X: x}), "Unary", "main.Unary"), err];
        }
    }
    return yield* go2jsMethodOnYield(this, Parser, "primary", []);
}
go2jsRegisterMethod("*Parser.unary", function*(receiver, ...args) {
    return yield* Parser.prototype.unary.call(receiver, ...args);
}, 2);

Parser.prototype.primary = function*() {
    let t = go2jsStructCopy(go2jsMethodOn(this, Parser, "next$", []));
    switch (t.Kind) {
        case NUM:
            {
            let [f, err] = go2jsStrconvParseFloat(t.Text, 64);
            if (go2jsEqual(err, null) === false) {
                return [null, go2jsWrapError("bad number %q", t.Text)];
            }
            return [go2jsInterface(f, "Num", "main.Num"), null];
        }
            break;
        case IDENT:
            {
            if (go2jsMethodOn(this, Parser, "peek$", []).Kind != LPAREN) {
                return [go2jsInterface(t.Text, "Var", "main.Var"), null];
            }
            go2jsMethodOn(this, Parser, "next$", []);
            let call = Object.assign(new Call(), {Name: t.Text});
            for (; go2jsMethodOn(this, Parser, "peek$", []).Kind != RPAREN;) {
                if (++go2jsState.go2jsTicks > 20000) {
                    yield* go2jsPreempt();
                }
                let [arg, err] = yield* go2jsMethodOnYield(this, Parser, "expr", [1]);
                if (go2jsEqual(err, null) === false) {
                    return [null, err];
                }
                call.Args = go2jsSliceAppend(call.Args, arg);
                if (go2jsMethodOn(this, Parser, "peek$", []).Kind == COMMA) {
                    go2jsMethodOn(this, Parser, "next$", []);
                } else if (go2jsMethodOn(this, Parser, "peek$", []).Kind != RPAREN) {
                    return [
                        null,
                        go2jsWrapError(
                            "expected , or ) at %d",
                            go2jsMethodOn(this, Parser, "peek$", []).Pos
                        )
                    ];
                }
            }
            go2jsMethodOn(this, Parser, "next$", []);
            return [go2jsInterface(go2jsStructCopy(call), "main.Call", "main.Call", true), null];
        }
            break;
        case LPAREN:
            {
            let [n, err] = yield* go2jsMethodOnYield(this, Parser, "expr", [1]);
            if (go2jsEqual(err, null) === false) {
                return [null, err];
            }
            if (go2jsMethodOn(this, Parser, "next$", []).Kind != RPAREN) {
                return [null, go2jsWrapError("missing )")];
            }
            return [n, null];
        }
            break;
    }
    return [null, go2jsWrapError("unexpected %v %q at %d", yield* TokenKindString(t.Kind), t.Text, t.Pos)];
}
go2jsRegisterMethod("*Parser.primary", function*(receiver, ...args) {
    return yield* Parser.prototype.primary.call(receiver, ...args);
}, 2);

function* Calc(src, env) {
    let [toks, err] = yield* lex(src);
    if (go2jsEqual(err, null) === false) {
        return ["", err];
    }
    let p = go2jsNew(Object.assign(new Parser(), {toks: toks}));
    const [go2jsM3_0, go2jsM3_1] = yield* go2jsMethodOnYield(p, Parser, "expr", [1]);
    let tree = go2jsM3_0;
    err = go2jsM3_1;
    if (go2jsEqual(err, null) === false) {
        return ["", err];
    }
    if (go2jsMethodOn(p, Parser, "peek$", []).Kind != EOF) {
        return [
            "",
            go2jsWrapError(
                "trailing %q at %d",
                go2jsMethodOn(p, Parser, "peek$", []).Text,
                go2jsMethodOn(p, Parser, "peek$", []).Pos
            )
        ];
    }
    const [go2jsM4_0, go2jsM4_1] = yield* go2jsInterfaceCall(tree, "Eval", env);
    let v = go2jsM4_0;
    err = go2jsM4_1;
    if (go2jsEqual(err, null) === false) {
        return ["", err];
    }
    return [go2jsSprintf("%s = %.6g", tree, go2jsTyped(v, "float64")), null];
}

function* main() {
    let env = go2jsMapTyped("map[string]float64", go2jsMap([["pi", Math.PI], ["x", 4], ["rate", 0.07]]));
    for (const src of go2jsRangeSequence(go2jsSliceTyped("[]string", ["1 + 2 * 3", "(1 + 2) * 3", "2 ^ 3 ^ 2", "-x ^ 2 + 10", "sqrt(16) + max(1, x, 3) * 2", "100 * (1 + rate) ^ 10", "17 % 5 + abs(-2.5)", "2 * pi * x", "1 / (x - 4)", "3 + ", "y * 2", "foo(1)", "4 $ 4", "(1 + 2"]))) {
        let [out, err] = yield* Calc(src, env);
        if (go2jsEqual(err, null) === false) {
            go2jsWriteOut(process.stdout, go2jsSprintf("%-30s error: %v\n", src, err));
            continue;
        }
        go2jsWriteOut(process.stdout, go2jsSprintf("%-30s %s\n", src, out));
    }
}

try {
    go2jsRunMain((function*() {
        go2jsRunInitializers();
        return yield* main();
    })());
} catch (e) {
    go2jsReportUncaughtPanic(e);
}
