"use strict";

const {
    go2jsCall,
    go2jsEmbedProxy,
    go2jsEqual,
    go2jsIndex,
    go2jsInterface,
    go2jsInterfaceCall,
    go2jsLen,
    go2jsMap,
    go2jsMapDelete,
    go2jsMapGetOK,
    go2jsMapSet,
    go2jsMapTyped,
    go2jsMapUpdate,
    go2jsMethodOn,
    go2jsMethodOnYield,
    go2jsNew,
    go2jsNilInterface,
    go2jsPrintln,
    go2jsRangeCopies,
    go2jsRangeMap,
    go2jsRangeSequence,
    go2jsRegisterInterface,
    go2jsRegisterMethod,
    go2jsRegisterTypeName,
    go2jsRegisterValueMethods,
    go2jsReportUncaughtPanic,
    go2jsRunMain,
    go2jsSetPackageName,
    go2jsSetSourceLines,
    go2jsSliceAppend,
    go2jsSliceMake,
    go2jsSliceTyped,
    go2jsSortSlice,
    go2jsSprintf,
    go2jsStrCompare,
    go2jsStringsRepeat,
    go2jsSwitchCaseIndex,
    go2jsTyped,
    go2jsWideAdd,
    go2jsWideSub,
    go2jsWrapError,
    go2jsWriteOut,
} = require("./main.runtime.js");

go2jsSetPackageName("main");
go2jsSetSourceLines({30: "01-inventory/main.go:23:15", 31: "01-inventory/main.go:23:35", 40: "01-inventory/main.go:24:15", 41: "01-inventory/main.go:24:35", 50: "01-inventory/main.go:25:15", 51: "01-inventory/main.go:25:35", 60: "01-inventory/main.go:26:15", 61: "01-inventory/main.go:26:42", 95: "01-inventory/main.go:34:15", 96: "01-inventory/main.go:35:43", 124: "01-inventory/main.go:43:17", 125: "01-inventory/main.go:44:39", 198: "01-inventory/main.go:61:6", 199: "01-inventory/main.go:61:34", 200: "01-inventory/main.go:61:34", 201: "01-inventory/main.go:61:34", 208: "01-inventory/main.go:63:23", 209: "01-inventory/main.go:64:2", 210: "01-inventory/main.go:64:24", 212: "01-inventory/main.go:65:3", 213: "01-inventory/main.go:66:3", 216: "01-inventory/main.go:68:12", 223: "01-inventory/main.go:71:23", 224: "01-inventory/main.go:72:2", 225: "01-inventory/main.go:73:2", 226: "01-inventory/main.go:74:10", 228: "01-inventory/main.go:76:2", 229: "01-inventory/main.go:77:10", 231: "01-inventory/main.go:79:2", 232: "01-inventory/main.go:80:2", 233: "01-inventory/main.go:81:3", 235: "01-inventory/main.go:83:2", 244: "01-inventory/main.go:86:23", 246: "01-inventory/main.go:87:28", 247: "01-inventory/main.go:88:2", 248: "01-inventory/main.go:89:10", 250: "01-inventory/main.go:91:2", 251: "01-inventory/main.go:91:48", 253: "01-inventory/main.go:92:2", 262: "01-inventory/main.go:95:23", 263: "01-inventory/main.go:96:2", 264: "01-inventory/main.go:97:2", 265: "01-inventory/main.go:98:29", 267: "01-inventory/main.go:100:2", 273: "01-inventory/main.go:103:6", 274: "01-inventory/main.go:104:2", 275: "01-inventory/main.go:105:14", 276: "01-inventory/main.go:106:2", 277: "01-inventory/main.go:107:82", 278: "01-inventory/main.go:107:82", 279: "01-inventory/main.go:107:82", 280: "01-inventory/main.go:107:82", 281: "01-inventory/main.go:107:82", 282: "01-inventory/main.go:107:82", 283: "01-inventory/main.go:107:82", 284: "01-inventory/main.go:107:82", 285: "01-inventory/main.go:107:82", 286: "01-inventory/main.go:107:82", 288: "01-inventory/main.go:109:2", 291: "01-inventory/main.go:112:6", 292: "01-inventory/main.go:113:9", 293: "01-inventory/main.go:114:2", 294: "01-inventory/main.go:114:2", 295: "01-inventory/main.go:114:2", 296: "01-inventory/main.go:114:2", 297: "01-inventory/main.go:114:2", 298: "01-inventory/main.go:114:2", 299: "01-inventory/main.go:114:2", 300: "01-inventory/main.go:114:2", 301: "01-inventory/main.go:114:2", 302: "01-inventory/main.go:114:2", 303: "01-inventory/main.go:114:2", 304: "01-inventory/main.go:114:2", 305: "01-inventory/main.go:114:2", 306: "01-inventory/main.go:114:2", 307: "01-inventory/main.go:114:2", 308: "01-inventory/main.go:114:2", 309: "01-inventory/main.go:114:2", 310: "01-inventory/main.go:114:2", 311: "01-inventory/main.go:114:2", 312: "01-inventory/main.go:114:2", 313: "01-inventory/main.go:115:2", 314: "01-inventory/main.go:115:2", 315: "01-inventory/main.go:115:2", 316: "01-inventory/main.go:115:2", 317: "01-inventory/main.go:115:2", 318: "01-inventory/main.go:115:2", 319: "01-inventory/main.go:115:2", 320: "01-inventory/main.go:115:2", 321: "01-inventory/main.go:115:2", 322: "01-inventory/main.go:115:2", 323: "01-inventory/main.go:115:2", 324: "01-inventory/main.go:115:2", 325: "01-inventory/main.go:115:2", 326: "01-inventory/main.go:115:2", 327: "01-inventory/main.go:115:2", 328: "01-inventory/main.go:115:2", 329: "01-inventory/main.go:115:2", 330: "01-inventory/main.go:115:2", 331: "01-inventory/main.go:115:2", 332: "01-inventory/main.go:115:2", 333: "01-inventory/main.go:115:2", 334: "01-inventory/main.go:115:2", 335: "01-inventory/main.go:115:2", 336: "01-inventory/main.go:116:2", 337: "01-inventory/main.go:116:2", 338: "01-inventory/main.go:116:2", 339: "01-inventory/main.go:116:2", 340: "01-inventory/main.go:116:2", 341: "01-inventory/main.go:116:2", 342: "01-inventory/main.go:116:2", 343: "01-inventory/main.go:116:2", 344: "01-inventory/main.go:116:2", 345: "01-inventory/main.go:116:2", 346: "01-inventory/main.go:116:2", 347: "01-inventory/main.go:116:2", 348: "01-inventory/main.go:116:2", 349: "01-inventory/main.go:116:2", 350: "01-inventory/main.go:116:2", 351: "01-inventory/main.go:116:2", 352: "01-inventory/main.go:116:2", 353: "01-inventory/main.go:116:2", 354: "01-inventory/main.go:116:2", 355: "01-inventory/main.go:117:2", 356: "01-inventory/main.go:117:2", 357: "01-inventory/main.go:117:2", 358: "01-inventory/main.go:117:2", 359: "01-inventory/main.go:117:2", 360: "01-inventory/main.go:117:2", 361: "01-inventory/main.go:117:2", 362: "01-inventory/main.go:117:2", 363: "01-inventory/main.go:117:2", 364: "01-inventory/main.go:117:2", 365: "01-inventory/main.go:117:2", 366: "01-inventory/main.go:117:2", 367: "01-inventory/main.go:117:2", 368: "01-inventory/main.go:117:2", 369: "01-inventory/main.go:117:2", 370: "01-inventory/main.go:117:2", 371: "01-inventory/main.go:117:2", 372: "01-inventory/main.go:117:2", 373: "01-inventory/main.go:117:2", 374: "01-inventory/main.go:118:2", 375: "01-inventory/main.go:118:2", 376: "01-inventory/main.go:118:2", 377: "01-inventory/main.go:118:2", 378: "01-inventory/main.go:118:2", 379: "01-inventory/main.go:118:2", 380: "01-inventory/main.go:118:2", 381: "01-inventory/main.go:118:2", 382: "01-inventory/main.go:118:2", 383: "01-inventory/main.go:118:2", 384: "01-inventory/main.go:118:2", 385: "01-inventory/main.go:118:2", 386: "01-inventory/main.go:118:2", 387: "01-inventory/main.go:118:2", 388: "01-inventory/main.go:118:2", 389: "01-inventory/main.go:118:2", 390: "01-inventory/main.go:118:2", 391: "01-inventory/main.go:118:2", 392: "01-inventory/main.go:118:2", 393: "01-inventory/main.go:119:2", 394: "01-inventory/main.go:119:2", 395: "01-inventory/main.go:119:2", 396: "01-inventory/main.go:119:2", 397: "01-inventory/main.go:119:2", 398: "01-inventory/main.go:119:2", 399: "01-inventory/main.go:119:2", 400: "01-inventory/main.go:119:2", 401: "01-inventory/main.go:119:2", 402: "01-inventory/main.go:119:2", 403: "01-inventory/main.go:119:2", 404: "01-inventory/main.go:119:2", 405: "01-inventory/main.go:119:2", 406: "01-inventory/main.go:119:2", 407: "01-inventory/main.go:119:2", 408: "01-inventory/main.go:119:2", 409: "01-inventory/main.go:119:2", 410: "01-inventory/main.go:119:2", 411: "01-inventory/main.go:119:2", 412: "01-inventory/main.go:120:2", 413: "01-inventory/main.go:120:2", 414: "01-inventory/main.go:120:2", 415: "01-inventory/main.go:120:2", 416: "01-inventory/main.go:120:2", 417: "01-inventory/main.go:120:2", 418: "01-inventory/main.go:120:2", 419: "01-inventory/main.go:120:2", 420: "01-inventory/main.go:120:2", 421: "01-inventory/main.go:120:2", 422: "01-inventory/main.go:120:2", 423: "01-inventory/main.go:120:2", 424: "01-inventory/main.go:120:2", 425: "01-inventory/main.go:120:2", 426: "01-inventory/main.go:120:2", 427: "01-inventory/main.go:120:2", 428: "01-inventory/main.go:120:2", 429: "01-inventory/main.go:120:2", 430: "01-inventory/main.go:120:2", 431: "01-inventory/main.go:122:19", 432: "01-inventory/main.go:122:77", 433: "01-inventory/main.go:122:77", 434: "01-inventory/main.go:122:77", 435: "01-inventory/main.go:122:77", 436: "01-inventory/main.go:122:77", 437: "01-inventory/main.go:122:77", 439: "01-inventory/main.go:123:36", 440: "01-inventory/main.go:124:57", 441: "01-inventory/main.go:124:57", 442: "01-inventory/main.go:124:57", 444: "01-inventory/main.go:127:42", 445: "01-inventory/main.go:127:42", 446: "01-inventory/main.go:127:42", 447: "01-inventory/main.go:127:42", 448: "01-inventory/main.go:127:42", 449: "01-inventory/main.go:127:42", 450: "01-inventory/main.go:127:42", 451: "01-inventory/main.go:128:2", 452: "01-inventory/main.go:128:2", 453: "01-inventory/main.go:128:2", 454: "01-inventory/main.go:128:2", 455: "01-inventory/main.go:128:2", 456: "01-inventory/main.go:128:2", 457: "01-inventory/main.go:128:2", 458: "01-inventory/main.go:128:2", 459: "01-inventory/main.go:128:2", 460: "01-inventory/main.go:128:2", 461: "01-inventory/main.go:128:2", 462: "01-inventory/main.go:128:2", 463: "01-inventory/main.go:128:2", 464: "01-inventory/main.go:128:2", 465: "01-inventory/main.go:128:2", 466: "01-inventory/main.go:128:2", 467: "01-inventory/main.go:128:2", 468: "01-inventory/main.go:128:2", 469: "01-inventory/main.go:128:2", 470: "01-inventory/main.go:128:2", 471: "01-inventory/main.go:128:2", 472: "01-inventory/main.go:128:2", 473: "01-inventory/main.go:128:2", 474: "01-inventory/main.go:128:2", 475: "01-inventory/main.go:128:2", 476: "01-inventory/main.go:128:2", 477: "01-inventory/main.go:128:2", 478: "01-inventory/main.go:128:2", 479: "01-inventory/main.go:128:2", 480: "01-inventory/main.go:128:2", 481: "01-inventory/main.go:128:2", 482: "01-inventory/main.go:128:2", 483: "01-inventory/main.go:128:2", 484: "01-inventory/main.go:128:2", 485: "01-inventory/main.go:128:2", 486: "01-inventory/main.go:128:2", 487: "01-inventory/main.go:128:2", 488: "01-inventory/main.go:128:2", 489: "01-inventory/main.go:128:2", 490: "01-inventory/main.go:128:2", 491: "01-inventory/main.go:128:2", 492: "01-inventory/main.go:128:2", 493: "01-inventory/main.go:128:2", 494: "01-inventory/main.go:128:2", 495: "01-inventory/main.go:128:2", 496: "01-inventory/main.go:128:2", 497: "01-inventory/main.go:128:2", 498: "01-inventory/main.go:128:2", 499: "01-inventory/main.go:128:2", 500: "01-inventory/main.go:128:2", 501: "01-inventory/main.go:128:2", 502: "01-inventory/main.go:128:2", 503: "01-inventory/main.go:128:2", 504: "01-inventory/main.go:128:2", 505: "01-inventory/main.go:128:2", 506: "01-inventory/main.go:128:2", 507: "01-inventory/main.go:128:2", 508: "01-inventory/main.go:132:3", 509: "01-inventory/main.go:132:13", 511: "01-inventory/main.go:133:4", 513: "01-inventory/main.go:135:4", 517: "01-inventory/main.go:138:61", 518: "01-inventory/main.go:138:61", 519: "01-inventory/main.go:138:61", 520: "01-inventory/main.go:138:61", 521: "01-inventory/main.go:138:61", 522: "01-inventory/main.go:138:61", 523: "01-inventory/main.go:138:61", 524: "01-inventory/main.go:138:61", 525: "01-inventory/main.go:140:2", 526: "01-inventory/main.go:141:2", 527: "01-inventory/main.go:142:3", 533: "01-inventory/main.go:144:4", 539: "01-inventory/main.go:146:4", 544: "01-inventory/main.go:148:4", 550: "01-inventory/main.go:151:2"});

class Item {}

go2jsRegisterInterface("Item", ["Describe", "Name", "Price", "SKU"]);

class Base {
    sku;
    name;
    price;
    constructor() {
        this.sku = "";
        this.name = "";
        this.price = 0;
    }
}

go2jsRegisterTypeName(
    Base,
    "main.Base",
    ["Describe", "Name", "Price", "SKU"],
    [
        {Name: "sku", Tag: "", PkgPath: "main", Anonymous: false, Type: "string", Kind: "string"},
        {Name: "name", Tag: "", PkgPath: "main", Anonymous: false, Type: "string", Kind: "string"},
        {Name: "price", Tag: "", PkgPath: "main", Anonymous: false, Type: "float64", Kind: "float64"}
    ],
    "struct"
);
go2jsRegisterValueMethods("main.Base", ["Describe", "Name", "Price", "SKU"]);

Base.prototype.SKU$ = function() {
    return this.sku;
};
Base.prototype.SKU = function*() {
    return Base.prototype.SKU$.call(this);
};
go2jsRegisterMethod("Base.SKU", function*(receiver, ...args) {
    return yield* Base.prototype.SKU.call(receiver, ...args);
}, 1);

Base.prototype.Name$ = function() {
    return this.name;
};
Base.prototype.Name = function*() {
    return Base.prototype.Name$.call(this);
};
go2jsRegisterMethod("Base.Name", function*(receiver, ...args) {
    return yield* Base.prototype.Name.call(receiver, ...args);
}, 1);

Base.prototype.Price$ = function() {
    return this.price;
};
Base.prototype.Price = function*() {
    return Base.prototype.Price$.call(this);
};
go2jsRegisterMethod("Base.Price", function*(receiver, ...args) {
    return yield* Base.prototype.Price.call(receiver, ...args);
}, 1);

Base.prototype.Describe$ = function() {
    return go2jsSprintf("%s (%s)", this.name, this.sku);
};
Base.prototype.Describe = function*() {
    return Base.prototype.Describe$.call(this);
};
go2jsRegisterMethod("Base.Describe", function*(receiver, ...args) {
    return yield* Base.prototype.Describe.call(receiver, ...args);
}, 1);

class Book {
    Base;
    Author;
    Pages;
    constructor() {
        this.Base = new Base();
        this.Author = "";
        this.Pages = 0;
        return go2jsEmbedProxy(this, ["Base"]);
    }
}

go2jsRegisterTypeName(
    Book,
    "main.Book",
    ["Describe", "Name", "Price", "SKU"],
    [
        {Name: "Base", Tag: "", PkgPath: "", Anonymous: true, Type: "main.Base", Kind: "struct"},
        {Name: "Author", Tag: "", PkgPath: "", Anonymous: false, Type: "string", Kind: "string"},
        {Name: "Pages", Tag: "", PkgPath: "", Anonymous: false, Type: "int", Kind: "int"}
    ],
    "struct"
);
go2jsRegisterValueMethods("main.Book", ["Describe", "Name", "Price", "SKU"]);

Book.prototype.Describe = function*() {
    return go2jsSprintf("%s by %s, %d pages", yield* this.Name(), this.Author, this.Pages);
}
go2jsRegisterMethod("Book.Describe", function*(receiver, ...args) {
    return yield* Book.prototype.Describe.call(receiver, ...args);
}, 1);

class Gadget {
    Base;
    Watts;
    constructor() {
        this.Base = new Base();
        this.Watts = 0;
        return go2jsEmbedProxy(this, ["Base"]);
    }
}

go2jsRegisterTypeName(
    Gadget,
    "main.Gadget",
    ["Describe", "Name", "Price", "SKU"],
    [
        {Name: "Base", Tag: "", PkgPath: "", Anonymous: true, Type: "main.Base", Kind: "struct"},
        {Name: "Watts", Tag: "", PkgPath: "", Anonymous: false, Type: "int", Kind: "int"}
    ],
    "struct"
);
go2jsRegisterValueMethods("main.Gadget", ["Describe", "Name", "Price", "SKU"]);

Gadget.prototype.Describe = function*() {
    return go2jsSprintf("%s drawing %dW", yield* this.Name(), this.Watts);
}
go2jsRegisterMethod("Gadget.Describe", function*(receiver, ...args) {
    return yield* Gadget.prototype.Describe.call(receiver, ...args);
}, 1);

class Grocery {
    Base;
    Perishable;
    constructor() {
        this.Base = new Base();
        this.Perishable = false;
        return go2jsEmbedProxy(this, ["Base"]);
    }
}

go2jsRegisterTypeName(
    Grocery,
    "main.Grocery",
    ["Describe", "Name", "Price", "SKU"],
    [
        {Name: "Base", Tag: "", PkgPath: "", Anonymous: true, Type: "main.Base", Kind: "struct"},
        {Name: "Perishable", Tag: "", PkgPath: "", Anonymous: false, Type: "bool", Kind: "bool"}
    ],
    "struct"
);
go2jsRegisterValueMethods("main.Grocery", ["Describe", "Name", "Price", "SKU"]);

class Stock {
    item;
    qty;
    constructor() {
        this.item = null;
        this.qty = 0;
    }
}

go2jsRegisterTypeName(
    Stock,
    "main.Stock",
    [],
    [
        {Name: "item", Tag: "", PkgPath: "main", Anonymous: false, Type: "main.Item", Kind: "interface"},
        {Name: "qty", Tag: "", PkgPath: "main", Anonymous: false, Type: "int", Kind: "int"}
    ],
    "struct"
);

class Inventory {
    stock;
    constructor() {
        this.stock = null;
    }
}

go2jsRegisterTypeName(
    Inventory,
    "main.Inventory",
    ["Add", "Remove", "Sorted", "Total"],
    [
        {
            Name: "stock",
            Tag: "",
            PkgPath: "main",
            Anonymous: false,
            Type: "map[string]*main.Stock",
            Kind: "map"
        }
    ],
    "struct"
);
go2jsRegisterValueMethods("main.Inventory", []);

function NewInventory$() {
    return go2jsNew(
        Object.assign(new Inventory(), {stock: go2jsMapTyped("map[string]*main.Stock", go2jsMap([]))})
    );
}

function* NewInventory() {
    return NewInventory$();
}

Inventory.prototype.Add = function*(item, qty) {
    {
        let [s, ok] = go2jsMapGetOK(this.stock, yield* go2jsInterfaceCall(item, "SKU"), null)
        if (ok) {
            s.qty = go2jsWideAdd(s.qty, qty, "int");
            return;
        }
    }
    const lhs1 = yield* go2jsInterfaceCall(item, "SKU");
    go2jsMapSet(this.stock, lhs1, go2jsNew(Object.assign(new Stock(), {item: item, qty: qty})));
}
go2jsRegisterMethod("*Inventory.Add", function*(receiver, ...args) {
    return yield* Inventory.prototype.Add.call(receiver, ...args);
}, 0);

Inventory.prototype.Remove$ = function(sku, qty) {
    let [s, ok] = go2jsMapGetOK(this.stock, sku, null);
    if (!ok) {
        return go2jsWrapError("unknown sku %q", sku);
    }
    if (s.qty < qty) {
        return go2jsWrapError("only %d of %s left, wanted %d", s.qty, sku, qty);
    }
    s.qty = go2jsWideSub(s.qty, qty, "int");
    if (s.qty == 0) {
        go2jsMapDelete(this.stock, sku);
    }
    return null;
};
Inventory.prototype.Remove = function*(sku, qty) {
    return Inventory.prototype.Remove$.call(this, sku, qty);
};
go2jsRegisterMethod("*Inventory.Remove", function*(receiver, ...args) {
    return yield* Inventory.prototype.Remove.call(receiver, ...args);
}, 1);

Inventory.prototype.Sorted$ = function(less) {
    const go2jsReceiver_inv = this;
    let list = go2jsSliceMake(0, go2jsLen(this.stock), null, true);
    for (const s of go2jsRangeMap(this.stock).values()) {
        list = go2jsSliceAppend(list, s);
    }
    go2jsSortSlice(go2jsNilInterface(list, "[]*main.Stock", "slice", true), function*(i, j) {
        return yield* go2jsCall(less, null, [go2jsIndex(list, i), go2jsIndex(list, j)]);
    });
    return list;
};
Inventory.prototype.Sorted = function*(less) {
    return Inventory.prototype.Sorted$.call(this, less);
};
go2jsRegisterMethod("*Inventory.Sorted", function*(receiver, ...args) {
    return yield* Inventory.prototype.Sorted.call(receiver, ...args);
}, 1);

Inventory.prototype.Total = function*() {
    let total = 0.0;
    for (const s of go2jsRangeMap(this.stock).values()) {
        total += (yield* go2jsInterfaceCall(s.item, "Price")) * Number(s.qty);
    }
    return total;
}
go2jsRegisterMethod("*Inventory.Total", function*(receiver, ...args) {
    return yield* Inventory.prototype.Total.call(receiver, ...args);
}, 1);

function* report(title, list) {
    go2jsPrintln(title);
    go2jsPrintln(go2jsStringsRepeat("-", 58));
    for (const s of go2jsRangeSequence(list)) {
        go2jsWriteOut(
            process.stdout,
            go2jsSprintf(
                "%-8s %-34s %4d x %7.2f\n",
                yield* go2jsInterfaceCall(s.item, "SKU"),
                yield* go2jsInterfaceCall(s.item, "Describe"),
                s.qty,
                go2jsTyped(yield* go2jsInterfaceCall(s.item, "Price"), "float64")
            )
        );
    }
    go2jsPrintln();
}

function* main() {
    let inv = NewInventory$();
    yield* go2jsMethodOnYield(
        inv,
        Inventory,
        "Add",
        [
            go2jsInterface(
                Object.assign(
                    new Book(),
                    {
                        Base: Object.assign(new Base(), {sku: "B-100", name: "Go in Practice", price: 39.9}),
                        Author: "A. Smith",
                        Pages: 320
                    }
                ),
                "Book",
                "main.Book"
            ),
            4
        ]
    );
    yield* go2jsMethodOnYield(
        inv,
        Inventory,
        "Add",
        [
            go2jsInterface(
                Object.assign(
                    new Book(),
                    {
                        Base: Object.assign(
                            new Base(),
                            {sku: "B-200", name: "Types and Programming", price: 54.5}
                        ),
                        Author: "B. Jones",
                        Pages: 623
                    }
                ),
                "Book",
                "main.Book"
            ),
            2
        ]
    );
    yield* go2jsMethodOnYield(
        inv,
        Inventory,
        "Add",
        [
            go2jsInterface(
                Object.assign(
                    new Gadget(),
                    {
                        Base: Object.assign(new Base(), {sku: "G-001", name: "Desk lamp", price: 24.0}),
                        Watts: 12
                    }
                ),
                "Gadget",
                "main.Gadget"
            ),
            10
        ]
    );
    yield* go2jsMethodOnYield(
        inv,
        Inventory,
        "Add",
        [
            go2jsInterface(
                Object.assign(
                    new Gadget(),
                    {
                        Base: Object.assign(new Base(), {sku: "G-002", name: "Kettle", price: 31.25}),
                        Watts: 1800
                    }
                ),
                "Gadget",
                "main.Gadget"
            ),
            3
        ]
    );
    yield* go2jsMethodOnYield(
        inv,
        Inventory,
        "Add",
        [
            go2jsInterface(
                Object.assign(
                    new Grocery(),
                    {
                        Base: Object.assign(new Base(), {sku: "F-900", name: "Coffee beans", price: 12.75}),
                        Perishable: false
                    }
                ),
                "Grocery",
                "main.Grocery"
            ),
            25
        ]
    );
    yield* go2jsMethodOnYield(
        inv,
        Inventory,
        "Add",
        [
            go2jsInterface(
                Object.assign(
                    new Grocery(),
                    {
                        Base: Object.assign(new Base(), {sku: "F-901", name: "Fresh milk", price: 1.2}),
                        Perishable: true
                    }
                ),
                "Grocery",
                "main.Grocery"
            ),
            40
        ]
    );
    yield* go2jsMethodOnYield(
        inv,
        Inventory,
        "Add",
        [
            go2jsInterface(
                Object.assign(
                    new Gadget(),
                    {
                        Base: Object.assign(new Base(), {sku: "G-001", name: "Desk lamp", price: 24.0}),
                        Watts: 12
                    }
                ),
                "Gadget",
                "main.Gadget"
            ),
            5
        ]
    );
    yield* report("By SKU", go2jsMethodOn(inv, Inventory, "Sorted$", [function*(a, b) {
        return (
            go2jsStrCompare(
                yield* go2jsInterfaceCall(a.item, "SKU"),
                yield* go2jsInterfaceCall(b.item, "SKU")
            ) < 0
        );
    }]));
    yield* report("By value, highest first", go2jsMethodOn(inv, Inventory, "Sorted$", [function*(a, b) {
        return (yield* go2jsInterfaceCall(a.item, "Price")) * Number(a.qty) > (
            yield* go2jsInterfaceCall(b.item, "Price")
        ) * Number(b.qty);
    }]));
    go2jsWriteOut(
        process.stdout,
        go2jsSprintf(
            "Total stock value: %.2f\n",
            go2jsTyped(yield* go2jsMethodOnYield(inv, Inventory, "Total", []), "float64")
        )
    );
    for (const step of go2jsRangeCopies(go2jsRangeSequence(go2jsSliceTyped("[]struct{sku string; qty int}", [(() => {
        const __v = {"sku": "G-001", "qty": 15};
        Object.defineProperty(__v, "__go2js_type_name", {value: "struct { sku string; qty int }"});
        Object.defineProperty(
            __v,
            "__go2js_struct_fields",
            {
                value: [
                    {Name: "sku", Type: "string", Anonymous: false},
                    {Name: "qty", Type: "int", Anonymous: false}
                ]
            }
        );
        return __v;
    })(), (() => {
        const __v = {"sku": "F-901", "qty": 41};
        Object.defineProperty(__v, "__go2js_type_name", {value: "struct { sku string; qty int }"});
        Object.defineProperty(
            __v,
            "__go2js_struct_fields",
            {
                value: [
                    {Name: "sku", Type: "string", Anonymous: false},
                    {Name: "qty", Type: "int", Anonymous: false}
                ]
            }
        );
        return __v;
    })(), (() => {
        const __v = {"sku": "X-000", "qty": 1};
        Object.defineProperty(__v, "__go2js_type_name", {value: "struct { sku string; qty int }"});
        Object.defineProperty(
            __v,
            "__go2js_struct_fields",
            {
                value: [
                    {Name: "sku", Type: "string", Anonymous: false},
                    {Name: "qty", Type: "int", Anonymous: false}
                ]
            }
        );
        return __v;
    })(), (() => {
        const __v = {"sku": "B-200", "qty": 2};
        Object.defineProperty(__v, "__go2js_type_name", {value: "struct { sku string; qty int }"});
        Object.defineProperty(
            __v,
            "__go2js_struct_fields",
            {
                value: [
                    {Name: "sku", Type: "string", Anonymous: false},
                    {Name: "qty", Type: "int", Anonymous: false}
                ]
            }
        );
        return __v;
    })()])))) {
        {
            let err = go2jsMethodOn(inv, Inventory, "Remove$", [step.sku, step.qty])
            if (go2jsEqual(err, null) === false) {
                go2jsPrintln("error:", err);
            } else {
                go2jsWriteOut(process.stdout, go2jsSprintf("removed %d of %s\n", step.qty, step.sku));
            }
        }
    }
    go2jsWriteOut(
        process.stdout,
        go2jsSprintf(
            "Items left: %d, value %.2f\n",
            go2jsLen(inv.stock),
            go2jsTyped(yield* go2jsMethodOnYield(inv, Inventory, "Total", []), "float64")
        )
    );
    let kinds = go2jsMapTyped("map[string]int", go2jsMap([]));
    for (const s of go2jsRangeMap(inv.stock).values()) {
        {
            let switch1 = s.item;
            switch (go2jsSwitchCaseIndex(switch1, ["Book", "Gadget"])) {
                case 0:

                    {
                    go2jsMapUpdate(kinds, "book", 0, (current) => go2jsWideAdd(current, 1, "int"));
                }
                    break;
                case 1:

                    {
                    go2jsMapUpdate(kinds, "gadget", 0, (current) => go2jsWideAdd(current, 1, "int"));
                }
                    break;
                default:
                    {
                    go2jsMapUpdate(kinds, "other", 0, (current) => go2jsWideAdd(current, 1, "int"));
                }
                    break;
            }
        }
    }
    go2jsPrintln(go2jsTyped(kinds, "map[string]int", "", "map"));
}

try {
    go2jsRunMain((function*() {
        return yield* main();
    })());
} catch (e) {
    go2jsReportUncaughtPanic(e);
}
