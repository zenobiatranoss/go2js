"use strict";

const {
    go2jsDeferInit,
    go2jsEqual,
    go2jsErrorsAs,
    go2jsErrorsIs,
    go2jsErrorsJoin,
    go2jsErrorsNew,
    go2jsGoexitSignal,
    go2jsInterfaceCall,
    go2jsIsNil,
    go2jsMap,
    go2jsMapGet,
    go2jsMapGetOK,
    go2jsMapSet,
    go2jsMapTyped,
    go2jsMethodOn,
    go2jsMethodOnYield,
    go2jsNew,
    go2jsPanic,
    go2jsPanicPayload,
    go2jsPrintln,
    go2jsPtr,
    go2jsRangeCopyEntries,
    go2jsRangeSequence,
    go2jsRecover,
    go2jsRegisterMethod,
    go2jsRegisterTypeName,
    go2jsRegisterValueMethods,
    go2jsReportUncaughtPanic,
    go2jsRunInitializers,
    go2jsRunMain,
    go2jsSetPackageName,
    go2jsSetSourceLines,
    go2jsSliceAppend,
    go2jsSliceTyped,
    go2jsSpreadArgs,
    go2jsSprintf,
    go2jsState,
    go2jsStringsJoin,
    go2jsStringsReplaceAll,
    go2jsWideAdd,
    go2jsWideSub,
    go2jsWideWrap,
    go2jsWrapError,
    go2jsWriteOut,
} = require("./main.runtime.js");

go2jsSetPackageName("main");
go2jsSetSourceLines({3: "02-bank/main.go:10:17", 27: "02-bank/main.go:16:29", 28: "02-bank/main.go:17:9", 64: "02-bank/main.go:27:19", 65: "02-bank/main.go:28:32", 74: "02-bank/main.go:31:19", 75: "02-bank/main.go:32:2", 76: "02-bank/main.go:33:10", 78: "02-bank/main.go:35:2", 79: "02-bank/main.go:36:9", 81: "02-bank/main.go:38:2", 82: "02-bank/main.go:39:2", 83: "02-bank/main.go:40:2", 92: "02-bank/main.go:43:19", 93: "02-bank/main.go:44:2", 94: "02-bank/main.go:45:10", 96: "02-bank/main.go:47:2", 97: "02-bank/main.go:48:10", 98: "02-bank/main.go:48:10", 99: "02-bank/main.go:48:10", 100: "02-bank/main.go:48:10", 101: "02-bank/main.go:48:10", 103: "02-bank/main.go:50:2", 104: "02-bank/main.go:51:2", 105: "02-bank/main.go:52:2", 139: "02-bank/main.go:59:16", 140: "02-bank/main.go:60:2", 141: "02-bank/main.go:61:2", 142: "02-bank/main.go:62:3", 144: "02-bank/main.go:64:2", 145: "02-bank/main.go:65:2", 154: "02-bank/main.go:68:16", 169: "02-bank/main.go:69:8", 170: "02-bank/main.go:69:8", 171: "02-bank/main.go:70:3", 172: "02-bank/main.go:70:11", 174: "02-bank/main.go:71:10", 179: "02-bank/main.go:74:2", 180: "02-bank/main.go:75:2", 181: "02-bank/main.go:76:10", 185: "02-bank/main.go:78:2", 188: "02-bank/main.go:79:2", 189: "02-bank/main.go:80:10", 193: "02-bank/main.go:82:2", 194: "02-bank/main.go:82:12", 196: "02-bank/main.go:83:3", 201: "02-bank/main.go:85:2", 202: "02-bank/main.go:85:12", 204: "02-bank/main.go:86:3", 205: "02-bank/main.go:87:3", 206: "02-bank/main.go:88:3", 211: "02-bank/main.go:90:2", 247: "02-bank/main.go:93:6", 250: "02-bank/main.go:94:2", 251: "02-bank/main.go:95:2", 254: "02-bank/main.go:97:3", 257: "02-bank/main.go:98:7", 259: "02-bank/main.go:99:3", 262: "02-bank/main.go:100:7", 263: "02-bank/main.go:100:7", 264: "02-bank/main.go:100:7", 265: "02-bank/main.go:100:7", 266: "02-bank/main.go:100:7", 268: "02-bank/main.go:101:10", 273: "02-bank/main.go:103:3", 283: "02-bank/main.go:107:6", 285: "02-bank/main.go:108:2", 286: "02-bank/main.go:109:2", 287: "02-bank/main.go:110:2", 288: "02-bank/main.go:111:12", 289: "02-bank/main.go:112:2", 290: "02-bank/main.go:114:2", 291: "02-bank/main.go:114:2", 292: "02-bank/main.go:114:2", 293: "02-bank/main.go:114:2", 294: "02-bank/main.go:114:2", 295: "02-bank/main.go:114:2", 296: "02-bank/main.go:114:2", 297: "02-bank/main.go:114:2", 298: "02-bank/main.go:114:2", 299: "02-bank/main.go:114:2", 300: "02-bank/main.go:114:2", 301: "02-bank/main.go:114:2", 302: "02-bank/main.go:114:2", 303: "02-bank/main.go:114:2", 304: "02-bank/main.go:114:2", 305: "02-bank/main.go:114:2", 306: "02-bank/main.go:114:2", 307: "02-bank/main.go:114:2", 308: "02-bank/main.go:114:2", 309: "02-bank/main.go:114:2", 310: "02-bank/main.go:114:2", 311: "02-bank/main.go:114:2", 312: "02-bank/main.go:114:2", 313: "02-bank/main.go:114:2", 314: "02-bank/main.go:114:2", 315: "02-bank/main.go:114:2", 316: "02-bank/main.go:114:2", 317: "02-bank/main.go:114:2", 318: "02-bank/main.go:114:2", 319: "02-bank/main.go:114:2", 320: "02-bank/main.go:114:2", 321: "02-bank/main.go:114:2", 322: "02-bank/main.go:114:2", 323: "02-bank/main.go:114:2", 324: "02-bank/main.go:114:2", 325: "02-bank/main.go:114:2", 326: "02-bank/main.go:114:2", 327: "02-bank/main.go:114:2", 328: "02-bank/main.go:114:2", 329: "02-bank/main.go:114:2", 330: "02-bank/main.go:114:2", 331: "02-bank/main.go:114:2", 332: "02-bank/main.go:114:2", 333: "02-bank/main.go:114:2", 334: "02-bank/main.go:114:2", 335: "02-bank/main.go:114:2", 336: "02-bank/main.go:114:2", 337: "02-bank/main.go:114:2", 338: "02-bank/main.go:114:2", 339: "02-bank/main.go:114:2", 340: "02-bank/main.go:114:2", 341: "02-bank/main.go:114:2", 342: "02-bank/main.go:114:2", 343: "02-bank/main.go:114:2", 344: "02-bank/main.go:114:2", 345: "02-bank/main.go:114:2", 346: "02-bank/main.go:114:2", 347: "02-bank/main.go:114:2", 348: "02-bank/main.go:114:2", 349: "02-bank/main.go:114:2", 350: "02-bank/main.go:114:2", 351: "02-bank/main.go:114:2", 352: "02-bank/main.go:114:2", 353: "02-bank/main.go:114:2", 354: "02-bank/main.go:114:2", 355: "02-bank/main.go:114:2", 356: "02-bank/main.go:114:2", 357: "02-bank/main.go:114:2", 358: "02-bank/main.go:114:2", 359: "02-bank/main.go:114:2", 360: "02-bank/main.go:114:2", 361: "02-bank/main.go:114:2", 362: "02-bank/main.go:114:2", 363: "02-bank/main.go:114:2", 364: "02-bank/main.go:114:2", 365: "02-bank/main.go:114:2", 366: "02-bank/main.go:114:2", 367: "02-bank/main.go:114:2", 368: "02-bank/main.go:114:2", 369: "02-bank/main.go:114:2", 370: "02-bank/main.go:114:2", 371: "02-bank/main.go:114:2", 372: "02-bank/main.go:114:2", 373: "02-bank/main.go:114:2", 374: "02-bank/main.go:114:2", 375: "02-bank/main.go:114:2", 376: "02-bank/main.go:114:2", 377: "02-bank/main.go:114:2", 378: "02-bank/main.go:114:2", 379: "02-bank/main.go:114:2", 380: "02-bank/main.go:114:2", 381: "02-bank/main.go:114:2", 382: "02-bank/main.go:114:2", 383: "02-bank/main.go:114:2", 384: "02-bank/main.go:114:2", 385: "02-bank/main.go:114:2", 386: "02-bank/main.go:114:2", 387: "02-bank/main.go:114:2", 388: "02-bank/main.go:114:2", 389: "02-bank/main.go:114:2", 390: "02-bank/main.go:114:2", 391: "02-bank/main.go:114:2", 392: "02-bank/main.go:114:2", 393: "02-bank/main.go:114:2", 394: "02-bank/main.go:114:2", 395: "02-bank/main.go:114:2", 396: "02-bank/main.go:114:2", 397: "02-bank/main.go:114:2", 398: "02-bank/main.go:114:2", 399: "02-bank/main.go:114:2", 400: "02-bank/main.go:114:2", 401: "02-bank/main.go:114:2", 402: "02-bank/main.go:114:2", 403: "02-bank/main.go:114:2", 404: "02-bank/main.go:114:2", 405: "02-bank/main.go:125:2", 406: "02-bank/main.go:126:10", 407: "02-bank/main.go:127:74", 408: "02-bank/main.go:127:74", 409: "02-bank/main.go:127:74", 410: "02-bank/main.go:127:74", 411: "02-bank/main.go:127:74", 412: "02-bank/main.go:127:74", 413: "02-bank/main.go:127:74", 414: "02-bank/main.go:127:74", 415: "02-bank/main.go:127:74", 416: "02-bank/main.go:127:74", 417: "02-bank/main.go:127:74", 418: "02-bank/main.go:127:74", 420: "02-bank/main.go:129:2", 421: "02-bank/main.go:130:3", 422: "02-bank/main.go:131:51", 423: "02-bank/main.go:131:51", 424: "02-bank/main.go:131:51", 425: "02-bank/main.go:131:51", 427: "02-bank/main.go:133:12", 428: "02-bank/main.go:133:12", 429: "02-bank/main.go:133:12", 430: "02-bank/main.go:133:12", 431: "02-bank/main.go:134:63", 432: "02-bank/main.go:134:63", 433: "02-bank/main.go:134:63", 434: "02-bank/main.go:134:63"});
let ErrFrozen;
go2jsDeferInit(function*() {
    ErrFrozen = go2jsErrorsNew("account frozen");
}, 0);

class InsufficientFunds {
    Need;
    Have;
    constructor() {
        this.Need = 0;
        this.Have = 0;
    }
}

go2jsRegisterTypeName(
    InsufficientFunds,
    "main.InsufficientFunds",
    ["Error"],
    [
        {Name: "Need", Tag: "", PkgPath: "", Anonymous: false, Type: "int", Kind: "int"},
        {Name: "Have", Tag: "", PkgPath: "", Anonymous: false, Type: "int", Kind: "int"}
    ],
    "struct"
);
go2jsRegisterValueMethods("main.InsufficientFunds", []);

InsufficientFunds.prototype.Error$ = function() {
    return go2jsSprintf("insufficient funds: need %d, have %d", this.Need, this.Have);
};
InsufficientFunds.prototype.Error = function*() {
    return InsufficientFunds.prototype.Error$.call(this);
};
go2jsRegisterMethod("*InsufficientFunds.Error", function*(receiver, ...args) {
    return yield* InsufficientFunds.prototype.Error.call(receiver, ...args);
}, 1);

class Account {
    ID;
    Balance;
    Frozen;
    History;
    constructor() {
        this.ID = "";
        this.Balance = 0;
        this.Frozen = false;
        this.History = null;
    }
}

go2jsRegisterTypeName(
    Account,
    "main.Account",
    ["Deposit", "Withdraw", "log"],
    [
        {Name: "ID", Tag: "", PkgPath: "", Anonymous: false, Type: "string", Kind: "string"},
        {Name: "Balance", Tag: "", PkgPath: "", Anonymous: false, Type: "int", Kind: "int"},
        {Name: "Frozen", Tag: "", PkgPath: "", Anonymous: false, Type: "bool", Kind: "bool"},
        {Name: "History", Tag: "", PkgPath: "", Anonymous: false, Type: "[]string", Kind: "slice"}
    ],
    "struct"
);
go2jsRegisterValueMethods("main.Account", []);

Account.prototype.log$ = function(format, ...args) {
    this.History = go2jsSliceAppend(this.History, go2jsSprintf(format, go2jsSpreadArgs(args)));
};
Account.prototype.log = function*(format, ...args) {
    return Account.prototype.log$.call(this, format, ...args);
};
go2jsRegisterMethod("*Account.log", function*(receiver, ...args) {
    return yield* Account.prototype.log.call(receiver, ...args);
}, 0);

Account.prototype.Deposit$ = function(n) {
    if (this.Frozen) {
        return go2jsWrapError("deposit to %s: %w", this.ID, ErrFrozen);
    }
    if (n <= 0) {
        go2jsPanic(go2jsSprintf("invalid deposit %d", n));
    }
    this.Balance = go2jsWideAdd(this.Balance, n, "int");
    go2jsMethodOn(this, Account, "log$", ["deposit %d", n]);
    return null;
};
Account.prototype.Deposit = function*(n) {
    return Account.prototype.Deposit$.call(this, n);
};
go2jsRegisterMethod("*Account.Deposit", function*(receiver, ...args) {
    return yield* Account.prototype.Deposit.call(receiver, ...args);
}, 1);

Account.prototype.Withdraw$ = function(n) {
    if (this.Frozen) {
        return go2jsWrapError("withdraw from %s: %w", this.ID, ErrFrozen);
    }
    if (this.Balance < n) {
        return go2jsWrapError(
            "withdraw from %s: %w",
            this.ID,
            go2jsNew(Object.assign(new InsufficientFunds(), {Need: n, Have: this.Balance}))
        );
    }
    this.Balance = go2jsWideSub(this.Balance, n, "int");
    go2jsMethodOn(this, Account, "log$", ["withdraw %d", n]);
    return null;
};
Account.prototype.Withdraw = function*(n) {
    return Account.prototype.Withdraw$.call(this, n);
};
go2jsRegisterMethod("*Account.Withdraw", function*(receiver, ...args) {
    return yield* Account.prototype.Withdraw.call(receiver, ...args);
}, 1);

class Bank {
    accounts;
    constructor() {
        this.accounts = null;
    }
}

go2jsRegisterTypeName(
    Bank,
    "main.Bank",
    ["Open", "Transfer"],
    [
        {
            Name: "accounts",
            Tag: "",
            PkgPath: "main",
            Anonymous: false,
            Type: "map[string]*main.Account",
            Kind: "map"
        }
    ],
    "struct"
);
go2jsRegisterValueMethods("main.Bank", []);

Bank.prototype.Open$ = function(id, initial) {
    let acc = go2jsNew(Object.assign(new Account(), {ID: id, Balance: initial}));
    if (go2jsIsNil(this.accounts)) {
        this.accounts = go2jsMapTyped("map[string]*main.Account", go2jsMap([]));
    }
    go2jsMapSet(this.accounts, id, acc);
    return acc;
};
Bank.prototype.Open = function*(id, initial) {
    return Bank.prototype.Open$.call(this, id, initial);
};
go2jsRegisterMethod("*Bank.Open", function*(receiver, ...args) {
    return yield* Bank.prototype.Open.call(receiver, ...args);
}, 1);

Bank.prototype.Transfer = function*(from, to, n) {
    const go2jsReceiver_b = this;
    const go2jsDefers = [];
    let go2jsPanicValue;
    let go2jsRecovered = false;
    const go2jsRecover = () => {
        if (go2jsPanicValue === undefined || go2jsRecovered) return undefined;
        if (go2jsPanicValue === go2jsGoexitSignal) return undefined;
        go2jsRecovered = true;
        return go2jsPanicPayload(go2jsPanicValue);
    };
    let err = null;
    const go2jsReturnSignal = {};
    let go2jsReturning = false;
    try {
        (() => go2jsDefers.push(function*() {
            yield* function*() {
                {
                    let r = go2jsRecover()
                    if (go2jsEqual(r, null) === false) {
                        err = go2jsWrapError("transfer aborted: %v", r);
                    }
                }
            }()
        }))();
        let [src, ok] = go2jsMapGetOK(this.accounts, from, null);
        if (!ok) {
            err = go2jsWrapError("no account %q", from);
            go2jsReturning = true;
            throw go2jsReturnSignal;
        }
        const [go2jsM1_0, go2jsM1_1] = go2jsMapGetOK(this.accounts, to, null);
        let dst = go2jsM1_0;
        ok = go2jsM1_1;
        if (!ok) {
            err = go2jsWrapError("no account %q", to);
            go2jsReturning = true;
            throw go2jsReturnSignal;
        }
        {
            let err$1 = go2jsMethodOn(src, Account, "Withdraw$", [n])
            if (go2jsEqual(err$1, null) === false) {
                err = err$1;
                go2jsReturning = true;
                throw go2jsReturnSignal;
            }
        }
        {
            let err$2 = go2jsMethodOn(dst, Account, "Deposit$", [n])
            if (go2jsEqual(err$2, null) === false) {
                src.Balance = go2jsWideAdd(src.Balance, n, "int");
                go2jsMethodOn(src, Account, "log$", ["rollback %d", n]);
                err = err$2;
                go2jsReturning = true;
                throw go2jsReturnSignal;
            }
        }
        err = null;
        go2jsReturning = true;
        throw go2jsReturnSignal;
    } catch (go2jsCaught) {
        if (go2jsCaught !== go2jsReturnSignal) {
            go2jsPanicValue = go2jsCaught;
        }
    } finally {
        for (let i = go2jsDefers.length - 1; i >= 0; i--) {
            try {
                const go2jsPrevRecover = go2jsState.go2jsRecoverHook;
                go2jsState.go2jsRecoverHook = go2jsRecover;
                try {
                    yield* go2jsDefers[i]();
                } finally {
                    go2jsState.go2jsRecoverHook = go2jsPrevRecover;
                }
            } catch (go2jsDeferThrown) {
                if (go2jsDeferThrown !== go2jsReturnSignal) {
                    go2jsPanicValue = go2jsDeferThrown;
                    go2jsRecovered = false;
                }
            }
        }
        if (go2jsPanicValue !== undefined && !go2jsRecovered) {
            throw go2jsPanicValue;
        }
    }
    if (go2jsReturning || go2jsRecovered) {
        return err
    }
}
go2jsRegisterMethod("*Bank.Transfer", function*(receiver, ...args) {
    return yield* Bank.prototype.Transfer.call(receiver, ...args);
}, 1);

function classify$(err) {
    const go2jsReceiver_b = this;
    let addr1 = null;
    let short = null;
    switch (true) {
        case go2jsEqual(err, null):
            {
            return "ok";
        }
            break;
        case go2jsErrorsIs(err, ErrFrozen):
            {
            return "frozen";
        }
            break;
        case go2jsErrorsAs(
            err,
            (addr1 ??= go2jsPtr(() => short, value => short = value)),
            "*InsufficientFunds"
        ):
            {
            return go2jsSprintf("short by %d", go2jsWideSub(short.Need, short.Have, "int"));
        }
            break;
        default:
            {
            return "other";
        }
            break;
    }
}

function* classify(err) {
    return classify$(err);
}

function* main() {
    const go2jsReceiver_b = this;
    let bank = go2jsNew(Object.assign(new Bank(), {}));
    go2jsMethodOn(bank, Bank, "Open$", ["alice", 100]);
    go2jsMethodOn(bank, Bank, "Open$", ["bob", 20]);
    let frozen = go2jsMethodOn(bank, Bank, "Open$", ["carol", 500]);
    frozen.Frozen = true;
    let tests = go2jsSliceTyped("[]struct{from string; to string; amount int}", [(() => {
        const __v = {"from": "alice", "to": "bob", "amount": 30};
        Object.defineProperty(
            __v,
            "__go2js_type_name",
            {value: "struct { from string; to string; amount int }"}
        );
        Object.defineProperty(
            __v,
            "__go2js_struct_fields",
            {
                value: [
                    {Name: "from", Type: "string", Anonymous: false},
                    {Name: "to", Type: "string", Anonymous: false},
                    {Name: "amount", Type: "int", Anonymous: false}
                ]
            }
        );
        return __v;
    })(), (() => {
        const __v = {"from": "bob", "to": "alice", "amount": 80};
        Object.defineProperty(
            __v,
            "__go2js_type_name",
            {value: "struct { from string; to string; amount int }"}
        );
        Object.defineProperty(
            __v,
            "__go2js_struct_fields",
            {
                value: [
                    {Name: "from", Type: "string", Anonymous: false},
                    {Name: "to", Type: "string", Anonymous: false},
                    {Name: "amount", Type: "int", Anonymous: false}
                ]
            }
        );
        return __v;
    })(), (() => {
        const __v = {"from": "alice", "to": "carol", "amount": 10};
        Object.defineProperty(
            __v,
            "__go2js_type_name",
            {value: "struct { from string; to string; amount int }"}
        );
        Object.defineProperty(
            __v,
            "__go2js_struct_fields",
            {
                value: [
                    {Name: "from", Type: "string", Anonymous: false},
                    {Name: "to", Type: "string", Anonymous: false},
                    {Name: "amount", Type: "int", Anonymous: false}
                ]
            }
        );
        return __v;
    })(), (() => {
        const __v = {"from": "dave", "to": "bob", "amount": 1};
        Object.defineProperty(
            __v,
            "__go2js_type_name",
            {value: "struct { from string; to string; amount int }"}
        );
        Object.defineProperty(
            __v,
            "__go2js_struct_fields",
            {
                value: [
                    {Name: "from", Type: "string", Anonymous: false},
                    {Name: "to", Type: "string", Anonymous: false},
                    {Name: "amount", Type: "int", Anonymous: false}
                ]
            }
        );
        return __v;
    })(), (() => {
        const __v = {"from": "alice", "to": "bob", "amount": go2jsWideWrap(-5, "int")};
        Object.defineProperty(
            __v,
            "__go2js_type_name",
            {value: "struct { from string; to string; amount int }"}
        );
        Object.defineProperty(
            __v,
            "__go2js_struct_fields",
            {
                value: [
                    {Name: "from", Type: "string", Anonymous: false},
                    {Name: "to", Type: "string", Anonymous: false},
                    {Name: "amount", Type: "int", Anonymous: false}
                ]
            }
        );
        return __v;
    })(), (() => {
        const __v = {"from": "carol", "to": "bob", "amount": 5};
        Object.defineProperty(
            __v,
            "__go2js_type_name",
            {value: "struct { from string; to string; amount int }"}
        );
        Object.defineProperty(
            __v,
            "__go2js_struct_fields",
            {
                value: [
                    {Name: "from", Type: "string", Anonymous: false},
                    {Name: "to", Type: "string", Anonymous: false},
                    {Name: "amount", Type: "int", Anonymous: false}
                ]
            }
        );
        return __v;
    })()]);
    for (const [i, t] of go2jsRangeCopyEntries(go2jsRangeSequence(tests).entries())) {
        let err = yield* go2jsMethodOnYield(bank, Bank, "Transfer", [t.from, t.to, t.amount]);
        go2jsWriteOut(
            process.stdout,
            go2jsSprintf(
                "%d. %s -> %s %d: %-14s %v\n",
                go2jsWideAdd(i, 1, "int"),
                t.from,
                t.to,
                t.amount,
                classify$(err),
                err
            )
        );
    }
    for (const id of go2jsRangeSequence(go2jsSliceTyped("[]string", ["alice", "bob", "carol"]))) {
        let acc = go2jsMapGet(bank.accounts, id, null);
        go2jsWriteOut(
            process.stdout,
            go2jsSprintf("%-6s %4d  [%s]\n", id, acc.Balance, go2jsStringsJoin(acc.History, "; "))
        );
    }
    let joined = go2jsErrorsJoin(
        ErrFrozen,
        go2jsNew(Object.assign(new InsufficientFunds(), {Need: 5, Have: 1}))
    );
    go2jsPrintln(
        go2jsErrorsIs(joined, ErrFrozen),
        go2jsStringsReplaceAll(yield* go2jsInterfaceCall(joined, "Error"), "\n", " | ")
    );
}

try {
    go2jsRunMain((function*() {
        go2jsRunInitializers();
        return yield* main();
    })());
} catch (e) {
    go2jsReportUncaughtPanic(e);
}
