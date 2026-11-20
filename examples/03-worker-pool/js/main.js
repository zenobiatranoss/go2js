"use strict";

const {
    go2jsChanRecv,
    go2jsChanRecvPair,
    go2jsChanRecvReady,
    go2jsChanSend,
    go2jsChannel,
    go2jsChannelClose,
    go2jsCopy,
    go2jsDuration,
    go2jsGo,
    go2jsGoexitSignal,
    go2jsIndex,
    go2jsMap,
    go2jsMapGet,
    go2jsMapTyped,
    go2jsMapUpdate,
    go2jsMod,
    go2jsMutex,
    go2jsMutexLock,
    go2jsMutexUnlock,
    go2jsNilInterface,
    go2jsPanicPayload,
    go2jsPreempt,
    go2jsPrintln,
    go2jsPtr,
    go2jsRangeCopies,
    go2jsRangeSequence,
    go2jsReady,
    go2jsRecover,
    go2jsRegisterMethod,
    go2jsRegisterTypeName,
    go2jsRegisterValueMethods,
    go2jsReportUncaughtPanic,
    go2jsRunMain,
    go2jsSelectWait,
    go2jsSetPackageName,
    go2jsSetSourceLines,
    go2jsSliceAppend,
    go2jsSliceTyped,
    go2jsSortSlice,
    go2jsSprintf,
    go2jsState,
    go2jsStructCopy,
    go2jsTimeAfter,
    go2jsTimeNewTicker,
    go2jsTimeZero,
    go2jsWaitGroup,
    go2jsWaitGroupAdd,
    go2jsWaitGroupDone,
    go2jsWaitGroupWait,
    go2jsWideAdd,
    go2jsWideMul,
    go2jsWideQuo,
    go2jsWriteOut,
} = require("./main.runtime.js");

go2jsSetPackageName("main");
go2jsSetSourceLines({45: "03-worker-pool/main.go:22:6", 46: "03-worker-pool/main.go:23:2", 47: "03-worker-pool/main.go:24:2", 48: "03-worker-pool/main.go:25:3", 49: "03-worker-pool/main.go:26:4", 51: "03-worker-pool/main.go:28:4", 53: "03-worker-pool/main.go:30:3", 55: "03-worker-pool/main.go:32:2", 62: "03-worker-pool/main.go:35:6", 73: "03-worker-pool/main.go:36:8", 74: "03-worker-pool/main.go:36:8", 75: "03-worker-pool/main.go:36:8", 76: "03-worker-pool/main.go:37:2", 82: "03-worker-pool/main.go:38:33", 83: "03-worker-pool/main.go:38:33", 84: "03-worker-pool/main.go:38:33", 85: "03-worker-pool/main.go:38:33", 110: "03-worker-pool/main.go:42:6", 111: "03-worker-pool/main.go:43:9", 112: "03-worker-pool/main.go:44:5", 113: "03-worker-pool/main.go:44:5", 124: "03-worker-pool/main.go:45:9", 125: "03-worker-pool/main.go:45:9", 126: "03-worker-pool/main.go:45:9", 127: "03-worker-pool/main.go:46:3", 128: "03-worker-pool/main.go:47:4", 153: "03-worker-pool/main.go:50:2", 160: "03-worker-pool/main.go:53:6", 161: "03-worker-pool/main.go:54:9", 162: "03-worker-pool/main.go:55:5", 163: "03-worker-pool/main.go:55:5", 174: "03-worker-pool/main.go:56:9", 175: "03-worker-pool/main.go:56:9", 176: "03-worker-pool/main.go:56:9", 177: "03-worker-pool/main.go:57:3", 183: "03-worker-pool/main.go:58:4", 208: "03-worker-pool/main.go:61:2", 215: "03-worker-pool/main.go:64:6", 216: "03-worker-pool/main.go:65:9", 217: "03-worker-pool/main.go:66:5", 218: "03-worker-pool/main.go:66:5", 229: "03-worker-pool/main.go:67:9", 230: "03-worker-pool/main.go:67:9", 231: "03-worker-pool/main.go:67:9", 232: "03-worker-pool/main.go:68:3", 238: "03-worker-pool/main.go:69:4", 239: "03-worker-pool/main.go:70:5", 265: "03-worker-pool/main.go:74:2", 293: "03-worker-pool/main.go:82:19", 305: "03-worker-pool/main.go:83:2", 306: "03-worker-pool/main.go:84:8", 307: "03-worker-pool/main.go:84:8", 308: "03-worker-pool/main.go:84:8", 309: "03-worker-pool/main.go:85:2", 336: "03-worker-pool/main.go:88:6", 348: "03-worker-pool/main.go:89:10", 349: "03-worker-pool/main.go:90:13", 350: "03-worker-pool/main.go:91:2", 351: "03-worker-pool/main.go:92:2", 352: "03-worker-pool/main.go:93:3", 353: "03-worker-pool/main.go:94:6", 354: "03-worker-pool/main.go:94:6", 355: "03-worker-pool/main.go:94:6", 357: "03-worker-pool/main.go:96:5", 358: "03-worker-pool/main.go:96:5", 359: "03-worker-pool/main.go:97:3", 360: "03-worker-pool/main.go:98:4", 361: "03-worker-pool/main.go:98:4", 362: "03-worker-pool/main.go:98:4", 363: "03-worker-pool/main.go:98:4", 365: "03-worker-pool/main.go:100:3", 368: "03-worker-pool/main.go:102:5", 369: "03-worker-pool/main.go:102:5", 370: "03-worker-pool/main.go:103:3", 371: "03-worker-pool/main.go:104:3", 374: "03-worker-pool/main.go:107:2", 375: "03-worker-pool/main.go:108:2", 381: "03-worker-pool/main.go:109:9", 383: "03-worker-pool/main.go:111:2", 384: "03-worker-pool/main.go:111:40", 386: "03-worker-pool/main.go:112:2", 387: "03-worker-pool/main.go:113:3", 389: "03-worker-pool/main.go:116:2", 390: "03-worker-pool/main.go:117:34", 396: "03-worker-pool/main.go:118:3", 398: "03-worker-pool/main.go:120:2", 399: "03-worker-pool/main.go:122:2", 400: "03-worker-pool/main.go:123:2", 401: "03-worker-pool/main.go:124:2", 402: "03-worker-pool/main.go:125:3", 403: "03-worker-pool/main.go:126:6", 404: "03-worker-pool/main.go:126:6", 415: "03-worker-pool/main.go:127:10", 416: "03-worker-pool/main.go:127:10", 417: "03-worker-pool/main.go:127:10", 418: "03-worker-pool/main.go:128:4", 419: "03-worker-pool/main.go:128:4", 420: "03-worker-pool/main.go:128:4", 421: "03-worker-pool/main.go:128:4", 422: "03-worker-pool/main.go:128:4", 423: "03-worker-pool/main.go:128:4", 448: "03-worker-pool/main.go:131:2", 449: "03-worker-pool/main.go:132:2", 450: "03-worker-pool/main.go:134:12", 451: "03-worker-pool/main.go:135:8", 452: "03-worker-pool/main.go:135:8", 453: "03-worker-pool/main.go:135:8", 454: "03-worker-pool/main.go:136:13", 455: "03-worker-pool/main.go:137:2", 456: "03-worker-pool/main.go:138:1", 460: "03-worker-pool/main.go:140:3", 480: "03-worker-pool/main.go:141:8", 482: "03-worker-pool/main.go:142:4", 486: "03-worker-pool/main.go:143:8", 488: "03-worker-pool/main.go:144:4", 493: "03-worker-pool/main.go:147:2", 494: "03-worker-pool/main.go:149:10", 495: "03-worker-pool/main.go:150:10", 496: "03-worker-pool/main.go:151:5", 497: "03-worker-pool/main.go:151:5", 498: "03-worker-pool/main.go:152:3", 500: "03-worker-pool/main.go:155:10", 518: "03-worker-pool/main.go:153:8", 520: "03-worker-pool/main.go:154:4", 524: "03-worker-pool/main.go:155:8", 526: "03-worker-pool/main.go:156:4", 530: "03-worker-pool/main.go:158:3", 533: "03-worker-pool/main.go:160:2", 534: "03-worker-pool/main.go:161:2", 535: "03-worker-pool/main.go:161:2", 536: "03-worker-pool/main.go:161:2", 537: "03-worker-pool/main.go:161:2", 538: "03-worker-pool/main.go:161:2", 539: "03-worker-pool/main.go:161:2"});

class Job {
    ID;
    Input;
    constructor() {
        this.ID = 0;
        this.Input = 0;
    }
}

go2jsRegisterTypeName(
    Job,
    "main.Job",
    [],
    [
        {Name: "ID", Tag: "", PkgPath: "", Anonymous: false, Type: "int", Kind: "int"},
        {Name: "Input", Tag: "", PkgPath: "", Anonymous: false, Type: "int", Kind: "int"}
    ],
    "struct"
);

class Result {
    JobID;
    Worker;
    Output;
    constructor() {
        this.JobID = 0;
        this.Worker = 0;
        this.Output = 0;
    }
}

go2jsRegisterTypeName(
    Result,
    "main.Result",
    [],
    [
        {Name: "JobID", Tag: "", PkgPath: "", Anonymous: false, Type: "int", Kind: "int"},
        {Name: "Worker", Tag: "", PkgPath: "", Anonymous: false, Type: "int", Kind: "int"},
        {Name: "Output", Tag: "", PkgPath: "", Anonymous: false, Type: "int", Kind: "int"}
    ],
    "struct"
);

function collatzSteps$(n) {
    let steps = 0;
    for (; n != 1;) {
        if (go2jsMod(n, 2) == 0) {
            n = go2jsWideQuo(n, 2, "int");
        } else {
            n = go2jsWideAdd(go2jsWideMul(3, n, "int"), 1, "int");
        }
        steps = go2jsWideAdd(steps, 1, "int");
    }
    return steps;
}

function* collatzSteps(n) {
    return collatzSteps$(n);
}

function* worker(id, jobs, results, wg) {
    const go2jsDefers = [];
    let go2jsPanicValue;
    let go2jsRecovered = false;
    const go2jsRecover = () => {
        if (go2jsPanicValue === undefined || go2jsRecovered) return undefined;
        if (go2jsPanicValue === go2jsGoexitSignal) return undefined;
        go2jsRecovered = true;
        return go2jsPanicPayload(go2jsPanicValue);
    };
    try {
        ((go2jsDeferRecv0) => go2jsDefers.push(function*() {
            go2jsWaitGroupDone(go2jsDeferRecv0)
        }))(wg);
        const rangechan3 = jobs;
        for (;;) {
            const received2 = yield* go2jsChanRecvPair(rangechan3);
            if (!received2[1]) break;
            const recv1 = received2[0];
            let job = go2jsCopy(recv1);
            yield* go2jsChanSend(
                results,
                Object.assign(new Result(), {JobID: job.ID, Worker: id, Output: collatzSteps$(job.Input)})
            );
        }
    } catch (go2jsCaught) {
        go2jsPanicValue = go2jsCaught;
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
                go2jsPanicValue = go2jsDeferThrown;
                go2jsRecovered = false;
            }
        }
        if (go2jsPanicValue !== undefined && !go2jsRecovered) {
            throw go2jsPanicValue;
        }
    }
}

function generate$(n) {
    let out = go2jsChannel(0);
    go2jsGo(function*() {
        yield* function*() {
            const go2jsDefers = [];
            let go2jsPanicValue;
            let go2jsRecovered = false;
            const go2jsRecover = () => {
                if (go2jsPanicValue === undefined || go2jsRecovered) return undefined;
                if (go2jsPanicValue === go2jsGoexitSignal) return undefined;
                go2jsRecovered = true;
                return go2jsPanicPayload(go2jsPanicValue);
            };
            try {
                ((go2jsDeferArg0) => go2jsDefers.push(function*() {
                    go2jsChannelClose(go2jsDeferArg0)
                }))(out);
                for (let i = 1; i <= n; i = go2jsWideAdd(i, 1, "int")) {
                    yield* go2jsChanSend(out, i);
                }
            } catch (go2jsCaught) {
                go2jsPanicValue = go2jsCaught;
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
                        go2jsPanicValue = go2jsDeferThrown;
                        go2jsRecovered = false;
                    }
                }
                if (go2jsPanicValue !== undefined && !go2jsRecovered) {
                    throw go2jsPanicValue;
                }
            }
        }()
    });
    return out;
}

function* generate(n) {
    return generate$(n);
}

function square$(in$go2js) {
    let out = go2jsChannel(0);
    go2jsGo(function*() {
        yield* function*() {
            const go2jsDefers = [];
            let go2jsPanicValue;
            let go2jsRecovered = false;
            const go2jsRecover = () => {
                if (go2jsPanicValue === undefined || go2jsRecovered) return undefined;
                if (go2jsPanicValue === go2jsGoexitSignal) return undefined;
                go2jsRecovered = true;
                return go2jsPanicPayload(go2jsPanicValue);
            };
            try {
                ((go2jsDeferArg0) => go2jsDefers.push(function*() {
                    go2jsChannelClose(go2jsDeferArg0)
                }))(out);
                const rangechan6 = in$go2js;
                for (;;) {
                    const received5 = yield* go2jsChanRecvPair(rangechan6);
                    if (!received5[1]) break;
                    const recv4 = received5[0];
                    let v = go2jsCopy(recv4);
                    yield* go2jsChanSend(out, go2jsWideMul(v, v, "int"));
                }
            } catch (go2jsCaught) {
                go2jsPanicValue = go2jsCaught;
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
                        go2jsPanicValue = go2jsDeferThrown;
                        go2jsRecovered = false;
                    }
                }
                if (go2jsPanicValue !== undefined && !go2jsRecovered) {
                    throw go2jsPanicValue;
                }
            }
        }()
    });
    return out;
}

function* square(in$go2js) {
    return square$(in$go2js);
}

function filterOdd$(in$go2js) {
    let out = go2jsChannel(0);
    go2jsGo(function*() {
        yield* function*() {
            const go2jsDefers = [];
            let go2jsPanicValue;
            let go2jsRecovered = false;
            const go2jsRecover = () => {
                if (go2jsPanicValue === undefined || go2jsRecovered) return undefined;
                if (go2jsPanicValue === go2jsGoexitSignal) return undefined;
                go2jsRecovered = true;
                return go2jsPanicPayload(go2jsPanicValue);
            };
            try {
                ((go2jsDeferArg0) => go2jsDefers.push(function*() {
                    go2jsChannelClose(go2jsDeferArg0)
                }))(out);
                const rangechan9 = in$go2js;
                for (;;) {
                    const received8 = yield* go2jsChanRecvPair(rangechan9);
                    if (!received8[1]) break;
                    const recv7 = received8[0];
                    let v = go2jsCopy(recv7);
                    if (go2jsMod(v, 2) == 1) {
                        yield* go2jsChanSend(out, v);
                    }
                }
            } catch (go2jsCaught) {
                go2jsPanicValue = go2jsCaught;
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
                        go2jsPanicValue = go2jsDeferThrown;
                        go2jsRecovered = false;
                    }
                }
                if (go2jsPanicValue !== undefined && !go2jsRecovered) {
                    throw go2jsPanicValue;
                }
            }
        }()
    });
    return out;
}

function* filterOdd(in$go2js) {
    return filterOdd$(in$go2js);
}

class Counter {
    mu;
    n;
    constructor() {
        this.mu = go2jsMutex();
        this.n = null;
    }
}

go2jsRegisterTypeName(
    Counter,
    "main.Counter",
    ["Inc"],
    [
        {Name: "mu", Tag: "", PkgPath: "main", Anonymous: false, Type: "sync.Mutex", Kind: "struct"},
        {Name: "n", Tag: "", PkgPath: "main", Anonymous: false, Type: "map[string]int", Kind: "map"}
    ],
    "struct"
);
go2jsRegisterValueMethods("main.Counter", []);

Counter.prototype.Inc = function*(key) {
    const go2jsReceiver_c = this;
    const go2jsDefers = [];
    let go2jsPanicValue;
    let go2jsRecovered = false;
    const go2jsRecover = () => {
        if (go2jsPanicValue === undefined || go2jsRecovered) return undefined;
        if (go2jsPanicValue === go2jsGoexitSignal) return undefined;
        go2jsRecovered = true;
        return go2jsPanicPayload(go2jsPanicValue);
    };
    try {
        yield* go2jsMutexLock(this.mu);
        ((go2jsDeferRecv0) => go2jsDefers.push(function*() {
            go2jsMutexUnlock(go2jsDeferRecv0)
        }))(this.mu);
        go2jsMapUpdate(this.n, key, 0, (current) => go2jsWideAdd(current, 1, "int"));
    } catch (go2jsCaught) {
        go2jsPanicValue = go2jsCaught;
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
                go2jsPanicValue = go2jsDeferThrown;
                go2jsRecovered = false;
            }
        }
        if (go2jsPanicValue !== undefined && !go2jsRecovered) {
            throw go2jsPanicValue;
        }
    }
}
go2jsRegisterMethod("*Counter.Inc", function*(receiver, ...args) {
    return yield* Counter.prototype.Inc.call(receiver, ...args);
}, 0);

function* main() {
    const go2jsReceiver_c = this;
    const go2jsDefers = [];
    let go2jsPanicValue;
    let go2jsRecovered = false;
    const go2jsRecover = () => {
        if (go2jsPanicValue === undefined || go2jsRecovered) return undefined;
        if (go2jsPanicValue === go2jsGoexitSignal) return undefined;
        go2jsRecovered = true;
        return go2jsPanicPayload(go2jsPanicValue);
    };
    try {
        let jobs = go2jsChannel(10);
        let results = go2jsChannel(10);
        let wg = go2jsWaitGroup();
        for (let w = 1; w <= 4; w = go2jsWideAdd(w, 1, "int")) {
            go2jsWaitGroupAdd(wg, 1);
            go2jsGo(function*() {
                yield* worker(w, jobs, results, go2jsPtr(() => wg, value => wg = value))
            });
        }
        go2jsGo(function*() {
            yield* function*() {
                for (let i = 1; i <= 20; i = go2jsWideAdd(i, 1, "int")) {
                    yield* go2jsChanSend(
                        jobs,
                        Object.assign(new Job(), {ID: i, Input: go2jsWideMul(i, 7, "int")})
                    );
                }
                go2jsChannelClose(jobs);
            }()
        });
        go2jsGo(function*() {
            yield* function*() {
                yield* go2jsWaitGroupWait(wg);
                go2jsChannelClose(results);
            }()
        });
        let all = null;
        const rangechan12 = results;
        for (;;) {
            const received11 = yield* go2jsChanRecvPair(rangechan12);
            if (!received11[1]) break;
            const recv10 = received11[0];
            let r = go2jsCopy(recv10);
            all = go2jsSliceAppend(all, go2jsStructCopy(r));
        }
        go2jsSortSlice(go2jsNilInterface(all, "[]main.Result", "slice", true), function*(i, j) {
            return go2jsIndex(all, i).JobID < go2jsIndex(all, j).JobID;
        });
        for (const r of go2jsRangeCopies(go2jsRangeSequence(all))) {
            go2jsWriteOut(process.stdout, go2jsSprintf("job %2d -> %3d steps\n", r.JobID, r.Output));
        }
        let sum = 0;
        const rangechan15 = filterOdd$(square$(generate$(10)));
        for (;;) {
            const received14 = yield* go2jsChanRecvPair(rangechan15);
            if (!received14[1]) break;
            const recv13 = received14[0];
            let v = go2jsCopy(recv13);
            sum = go2jsWideAdd(sum, v, "int");
        }
        go2jsPrintln("sum of odd squares up to 10:", sum);
        let c = Object.assign(new Counter(), {n: go2jsMapTyped("map[string]int", go2jsMap([]))});
        let wg2 = go2jsWaitGroup();
        for (let i = 0; i < 50; i = go2jsWideAdd(i, 1, "int")) {
            go2jsWaitGroupAdd(wg2, 1);
            go2jsGo(function*() {
                yield* function*(i) {
                    const go2jsDefers = [];
                    let go2jsPanicValue;
                    let go2jsRecovered = false;
                    const go2jsRecover = () => {
                        if (go2jsPanicValue === undefined || go2jsRecovered) return undefined;
                        if (go2jsPanicValue === go2jsGoexitSignal) return undefined;
                        go2jsRecovered = true;
                        return go2jsPanicPayload(go2jsPanicValue);
                    };
                    try {
                        ((go2jsDeferRecv0) => go2jsDefers.push(function*() {
                            go2jsWaitGroupDone(go2jsDeferRecv0)
                        }))(wg2);
                        yield* c.Inc(
                            go2jsIndex(
                                go2jsSliceTyped("[]string", ["red", "green", "blue"]),
                                go2jsMod(i, 3)
                            )
                        );
                    } catch (go2jsCaught) {
                        go2jsPanicValue = go2jsCaught;
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
                                go2jsPanicValue = go2jsDeferThrown;
                                go2jsRecovered = false;
                            }
                        }
                        if (go2jsPanicValue !== undefined && !go2jsRecovered) {
                            throw go2jsPanicValue;
                        }
                    }
                }(i)
            });
        }
        yield* go2jsWaitGroupWait(wg2);
        go2jsPrintln(go2jsMapGet(c.n, "red", 0), go2jsMapGet(c.n, "green", 0), go2jsMapGet(c.n, "blue", 0));
        let ticker = go2jsTimeNewTicker(go2jsDuration(go2jsWideMul(10, go2jsDuration(1000000), "int64")));
        ((go2jsDeferRecv0) => go2jsDefers.push(function*() {
            go2jsDeferRecv0.Stop()
        }))(ticker);
        let timeout = go2jsTimeAfter(go2jsDuration(go2jsWideMul(55, go2jsDuration(1000000), "int64")));
        let ticks = 0;
        loop: for (;;) {
            if (++go2jsState.go2jsTicks > 20000) {
                yield* go2jsPreempt();
            }
            {
                const go2jsSelCh0 = ticker.C;
                const go2jsSelCh1 = timeout;
                let go2jsPicked = -1;
                while (true) {
                    const go2jsReady = [];
                    if (go2jsChanRecvReady(go2jsSelCh0)) {
                        go2jsReady.push(0);
                    }
                    if (go2jsChanRecvReady(go2jsSelCh1)) {
                        go2jsReady.push(1);
                    }
                    if (go2jsReady.length === 0) {
                        yield* go2jsSelectWait([], [go2jsSelCh0, go2jsSelCh1]);
                        continue;
                    }
                    go2jsPicked = go2jsReady[Math.floor(Math.random() * go2jsReady.length)];
                    break;
                }
                if (go2jsPicked === 0) {
                    yield* go2jsChanRecv(go2jsSelCh0, go2jsTimeZero());
                    go2jsSelBody$0: {
                        ticks = go2jsWideAdd(ticks, 1, "int");
                    }
                }
                if (go2jsPicked === 1) {
                    yield* go2jsChanRecv(go2jsSelCh1, go2jsTimeZero());
                    go2jsSelBody$1: {
                        break loop;
                    }
                }
            }
        }
        go2jsPrintln("ticker fired repeatedly:", ticks >= 3);
        let done = go2jsChannel(0);
        let data = go2jsChannel(1);
        go2jsGo(function*() {
            yield* function*() {
                {
                    const go2jsSelCh0 = data;
                    const go2jsSelCh1 = go2jsTimeAfter(go2jsDuration(1000000000));
                    let go2jsPicked = -1;
                    while (true) {
                        const go2jsReady = [];
                        if (go2jsChanRecvReady(go2jsSelCh0)) {
                            go2jsReady.push(0);
                        }
                        if (go2jsChanRecvReady(go2jsSelCh1)) {
                            go2jsReady.push(1);
                        }
                        if (go2jsReady.length === 0) {
                            yield* go2jsSelectWait([], [go2jsSelCh0, go2jsSelCh1]);
                            continue;
                        }
                        go2jsPicked = go2jsReady[Math.floor(Math.random() * go2jsReady.length)];
                        break;
                    }
                    if (go2jsPicked === 0) {
                        let m = yield* go2jsChanRecv(go2jsSelCh0, "");
                        go2jsSelBody$2: {
                            go2jsPrintln("got", m);
                        }
                    }
                    if (go2jsPicked === 1) {
                        yield* go2jsChanRecv(go2jsSelCh1, go2jsTimeZero());
                        go2jsSelBody$3: {
                            go2jsPrintln("too slow");
                        }
                    }
                }
                go2jsChannelClose(done);
            }()
        });
        yield* go2jsChanSend(data, "hello from select");
        yield* go2jsChanRecv(done, (() => {
            const __v = {};
            Object.defineProperty(__v, "__go2js_type_name", {value: "struct {}"});
            Object.defineProperty(__v, "__go2js_struct_fields", {value: []});
            return __v;
        })());
    } catch (go2jsCaught) {
        go2jsPanicValue = go2jsCaught;
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
                go2jsPanicValue = go2jsDeferThrown;
                go2jsRecovered = false;
            }
        }
        if (go2jsPanicValue !== undefined && !go2jsRecovered) {
            throw go2jsPanicValue;
        }
    }
}

try {
    go2jsRunMain((function*() {
        return yield* main();
    })());
} catch (e) {
    go2jsReportUncaughtPanic(e);
}
