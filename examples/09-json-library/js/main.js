"use strict";

const {
    DurationHours,
    go2jsBytesToString,
    go2jsDuration,
    go2jsEqual,
    go2jsFprint,
    go2jsIndex,
    go2jsInterface,
    go2jsJSONMarshal,
    go2jsJSONUnmarshal,
    go2jsLen,
    go2jsMap,
    go2jsMapGet,
    go2jsMapTyped,
    go2jsNew,
    go2jsNilInterface,
    go2jsOSStdout,
    go2jsPanic,
    go2jsPrintln,
    go2jsPtr,
    go2jsRangeCopies,
    go2jsRangeMap,
    go2jsRangeSequence,
    go2jsRegisterTypeName,
    go2jsReportUncaughtPanic,
    go2jsRunMain,
    go2jsSetPackageName,
    go2jsSetSourceLines,
    go2jsSliceAppend,
    go2jsSliceMake,
    go2jsSliceRange,
    go2jsSliceTyped,
    go2jsSortSlice,
    go2jsSortStrings,
    go2jsSprintf,
    go2jsStringToBytes,
    go2jsStringsJoin,
    go2jsStructCopy,
    go2jsTabwriterFlush,
    go2jsTabwriterNewWriter,
    go2jsTimeDate,
    go2jsTimeLocation,
    go2jsTimeParse,
    go2jsTyped,
    go2jsWideMul,
    go2jsWideToSigned,
    go2jsWriteOut,
} = require("./main.runtime.js");

go2jsSetPackageName("main");
go2jsSetSourceLines({148: "09-json-library/main.go:47:6", 149: "09-json-library/main.go:48:12", 150: "09-json-library/main.go:49:2", 151: "09-json-library/main.go:50:3", 153: "09-json-library/main.go:52:2", 160: "09-json-library/main.go:55:6", 164: "09-json-library/main.go:56:2", 165: "09-json-library/main.go:57:2", 166: "09-json-library/main.go:57:27", 167: "09-json-library/main.go:57:27", 168: "09-json-library/main.go:57:27", 169: "09-json-library/main.go:57:27", 170: "09-json-library/main.go:57:27", 171: "09-json-library/main.go:57:27", 172: "09-json-library/main.go:57:27", 173: "09-json-library/main.go:57:27", 174: "09-json-library/main.go:57:27", 175: "09-json-library/main.go:57:27", 176: "09-json-library/main.go:57:27", 177: "09-json-library/main.go:57:27", 178: "09-json-library/main.go:57:27", 179: "09-json-library/main.go:57:27", 180: "09-json-library/main.go:57:27", 181: "09-json-library/main.go:57:27", 182: "09-json-library/main.go:57:27", 183: "09-json-library/main.go:57:27", 184: "09-json-library/main.go:57:27", 185: "09-json-library/main.go:57:27", 186: "09-json-library/main.go:57:27", 187: "09-json-library/main.go:57:27", 188: "09-json-library/main.go:57:27", 189: "09-json-library/main.go:57:27", 190: "09-json-library/main.go:57:27", 191: "09-json-library/main.go:57:27", 192: "09-json-library/main.go:57:27", 193: "09-json-library/main.go:57:27", 194: "09-json-library/main.go:57:27", 195: "09-json-library/main.go:57:27", 196: "09-json-library/main.go:57:27", 197: "09-json-library/main.go:57:27", 198: "09-json-library/main.go:57:27", 199: "09-json-library/main.go:57:27", 200: "09-json-library/main.go:57:27", 201: "09-json-library/main.go:57:27", 202: "09-json-library/main.go:57:27", 203: "09-json-library/main.go:57:27", 204: "09-json-library/main.go:57:27", 205: "09-json-library/main.go:57:27", 206: "09-json-library/main.go:57:27", 207: "09-json-library/main.go:57:27", 208: "09-json-library/main.go:57:27", 209: "09-json-library/main.go:57:27", 210: "09-json-library/main.go:57:27", 211: "09-json-library/main.go:57:27", 212: "09-json-library/main.go:57:27", 213: "09-json-library/main.go:57:27", 214: "09-json-library/main.go:57:27", 215: "09-json-library/main.go:57:27", 216: "09-json-library/main.go:57:27", 217: "09-json-library/main.go:57:27", 218: "09-json-library/main.go:57:27", 219: "09-json-library/main.go:57:27", 220: "09-json-library/main.go:57:27", 221: "09-json-library/main.go:57:27", 222: "09-json-library/main.go:57:27", 223: "09-json-library/main.go:57:27", 224: "09-json-library/main.go:57:27", 225: "09-json-library/main.go:57:27", 226: "09-json-library/main.go:57:27", 227: "09-json-library/main.go:57:27", 228: "09-json-library/main.go:57:27", 229: "09-json-library/main.go:57:27", 230: "09-json-library/main.go:57:27", 231: "09-json-library/main.go:57:27", 232: "09-json-library/main.go:57:27", 233: "09-json-library/main.go:57:27", 234: "09-json-library/main.go:57:27", 235: "09-json-library/main.go:57:27", 236: "09-json-library/main.go:57:27", 237: "09-json-library/main.go:57:27", 238: "09-json-library/main.go:57:27", 239: "09-json-library/main.go:57:27", 241: "09-json-library/main.go:58:3", 242: "09-json-library/main.go:59:3", 245: "09-json-library/main.go:61:24", 246: "09-json-library/main.go:63:8", 247: "09-json-library/main.go:64:2", 248: "09-json-library/main.go:65:2", 249: "09-json-library/main.go:66:3", 250: "09-json-library/main.go:67:3", 251: "09-json-library/main.go:68:12", 253: "09-json-library/main.go:70:3", 254: "09-json-library/main.go:71:3", 255: "09-json-library/main.go:72:4", 257: "09-json-library/main.go:74:92", 258: "09-json-library/main.go:74:92", 259: "09-json-library/main.go:74:92", 260: "09-json-library/main.go:74:92", 261: "09-json-library/main.go:74:92", 262: "09-json-library/main.go:74:92", 263: "09-json-library/main.go:74:92", 264: "09-json-library/main.go:74:92", 265: "09-json-library/main.go:74:92", 266: "09-json-library/main.go:74:92", 267: "09-json-library/main.go:74:92", 268: "09-json-library/main.go:74:92", 269: "09-json-library/main.go:74:92", 270: "09-json-library/main.go:74:92", 271: "09-json-library/main.go:74:92", 272: "09-json-library/main.go:74:92", 274: "09-json-library/main.go:76:2", 275: "09-json-library/main.go:78:2", 276: "09-json-library/main.go:78:46", 278: "09-json-library/main.go:79:2", 279: "09-json-library/main.go:81:2", 280: "09-json-library/main.go:82:2", 281: "09-json-library/main.go:82:2", 282: "09-json-library/main.go:82:2", 283: "09-json-library/main.go:82:2", 284: "09-json-library/main.go:83:14", 285: "09-json-library/main.go:83:14", 286: "09-json-library/main.go:83:14", 287: "09-json-library/main.go:83:14", 288: "09-json-library/main.go:83:14", 289: "09-json-library/main.go:83:14", 290: "09-json-library/main.go:83:14", 291: "09-json-library/main.go:83:14", 292: "09-json-library/main.go:83:14", 293: "09-json-library/main.go:84:14", 294: "09-json-library/main.go:84:14", 295: "09-json-library/main.go:84:14", 296: "09-json-library/main.go:84:14", 297: "09-json-library/main.go:84:14", 298: "09-json-library/main.go:84:14", 299: "09-json-library/main.go:84:14", 300: "09-json-library/main.go:84:14", 301: "09-json-library/main.go:84:14", 302: "09-json-library/main.go:84:14", 303: "09-json-library/main.go:84:14", 304: "09-json-library/main.go:84:14", 305: "09-json-library/main.go:84:14", 306: "09-json-library/main.go:84:14", 307: "09-json-library/main.go:84:14", 308: "09-json-library/main.go:84:14", 309: "09-json-library/main.go:84:14", 310: "09-json-library/main.go:84:14", 311: "09-json-library/main.go:84:14", 312: "09-json-library/main.go:84:14", 313: "09-json-library/main.go:84:14", 314: "09-json-library/main.go:84:14", 315: "09-json-library/main.go:84:14", 316: "09-json-library/main.go:84:14", 317: "09-json-library/main.go:84:14", 318: "09-json-library/main.go:84:14", 319: "09-json-library/main.go:84:14", 320: "09-json-library/main.go:84:14", 321: "09-json-library/main.go:84:14", 322: "09-json-library/main.go:84:14", 323: "09-json-library/main.go:84:14", 324: "09-json-library/main.go:84:14", 325: "09-json-library/main.go:84:14", 326: "09-json-library/main.go:84:14", 327: "09-json-library/main.go:84:14", 328: "09-json-library/main.go:84:14", 329: "09-json-library/main.go:84:14", 330: "09-json-library/main.go:84:14", 331: "09-json-library/main.go:84:14", 332: "09-json-library/main.go:84:14", 333: "09-json-library/main.go:84:14", 334: "09-json-library/main.go:84:14", 335: "09-json-library/main.go:84:14", 336: "09-json-library/main.go:84:14", 337: "09-json-library/main.go:84:14", 338: "09-json-library/main.go:84:14", 339: "09-json-library/main.go:84:14", 340: "09-json-library/main.go:84:14", 341: "09-json-library/main.go:84:14", 342: "09-json-library/main.go:84:14", 343: "09-json-library/main.go:84:14", 344: "09-json-library/main.go:84:14", 345: "09-json-library/main.go:84:14", 346: "09-json-library/main.go:84:14", 347: "09-json-library/main.go:84:14", 348: "09-json-library/main.go:84:14", 349: "09-json-library/main.go:84:14", 350: "09-json-library/main.go:84:14", 351: "09-json-library/main.go:84:14", 352: "09-json-library/main.go:84:14", 353: "09-json-library/main.go:84:14", 354: "09-json-library/main.go:84:14", 355: "09-json-library/main.go:84:14", 356: "09-json-library/main.go:84:14", 357: "09-json-library/main.go:85:14", 358: "09-json-library/main.go:87:16", 359: "09-json-library/main.go:87:16", 360: "09-json-library/main.go:87:16", 361: "09-json-library/main.go:87:16", 362: "09-json-library/main.go:87:16", 363: "09-json-library/main.go:87:16", 364: "09-json-library/main.go:87:16", 365: "09-json-library/main.go:87:16", 366: "09-json-library/main.go:87:16", 367: "09-json-library/main.go:87:16", 368: "09-json-library/main.go:87:16", 369: "09-json-library/main.go:87:16", 370: "09-json-library/main.go:87:16", 371: "09-json-library/main.go:87:16", 372: "09-json-library/main.go:87:16", 373: "09-json-library/main.go:87:16", 374: "09-json-library/main.go:87:16", 375: "09-json-library/main.go:87:16", 376: "09-json-library/main.go:87:16", 377: "09-json-library/main.go:87:16", 378: "09-json-library/main.go:88:14", 379: "09-json-library/main.go:90:2", 380: "09-json-library/main.go:91:17", 381: "09-json-library/main.go:91:17", 382: "09-json-library/main.go:91:17", 383: "09-json-library/main.go:91:17", 384: "09-json-library/main.go:91:17", 385: "09-json-library/main.go:91:17", 386: "09-json-library/main.go:91:17", 387: "09-json-library/main.go:92:10", 388: "09-json-library/main.go:93:2", 389: "09-json-library/main.go:94:10", 391: "09-json-library/main.go:96:2", 392: "09-json-library/main.go:97:2", 393: "09-json-library/main.go:98:3", 394: "09-json-library/main.go:98:3", 395: "09-json-library/main.go:98:3", 396: "09-json-library/main.go:98:3", 397: "09-json-library/main.go:98:3", 398: "09-json-library/main.go:98:3", 399: "09-json-library/main.go:98:3", 400: "09-json-library/main.go:98:3", 401: "09-json-library/main.go:98:3", 403: "09-json-library/main.go:101:2", 404: "09-json-library/main.go:102:29", 405: "09-json-library/main.go:102:29", 406: "09-json-library/main.go:102:29", 407: "09-json-library/main.go:102:29", 408: "09-json-library/main.go:102:29", 409: "09-json-library/main.go:102:29", 410: "09-json-library/main.go:102:29", 411: "09-json-library/main.go:102:29", 412: "09-json-library/main.go:102:29", 413: "09-json-library/main.go:102:29", 414: "09-json-library/main.go:102:29", 415: "09-json-library/main.go:102:29", 416: "09-json-library/main.go:102:29", 417: "09-json-library/main.go:102:29", 418: "09-json-library/main.go:102:29", 419: "09-json-library/main.go:102:29", 420: "09-json-library/main.go:102:29", 421: "09-json-library/main.go:102:29", 422: "09-json-library/main.go:102:29", 423: "09-json-library/main.go:102:29", 424: "09-json-library/main.go:102:29", 425: "09-json-library/main.go:102:29", 426: "09-json-library/main.go:102:29", 427: "09-json-library/main.go:102:29", 428: "09-json-library/main.go:102:29", 429: "09-json-library/main.go:102:29", 430: "09-json-library/main.go:102:29", 431: "09-json-library/main.go:102:29", 432: "09-json-library/main.go:102:29", 433: "09-json-library/main.go:102:29", 434: "09-json-library/main.go:102:29", 435: "09-json-library/main.go:102:29", 436: "09-json-library/main.go:102:29", 437: "09-json-library/main.go:102:29", 438: "09-json-library/main.go:102:29", 439: "09-json-library/main.go:102:29", 440: "09-json-library/main.go:102:29", 441: "09-json-library/main.go:102:29", 442: "09-json-library/main.go:102:29", 443: "09-json-library/main.go:102:29", 444: "09-json-library/main.go:102:29", 445: "09-json-library/main.go:102:29", 446: "09-json-library/main.go:102:29", 447: "09-json-library/main.go:102:29", 448: "09-json-library/main.go:102:29", 449: "09-json-library/main.go:102:29", 450: "09-json-library/main.go:102:29", 451: "09-json-library/main.go:102:29", 452: "09-json-library/main.go:102:29", 453: "09-json-library/main.go:102:29", 454: "09-json-library/main.go:102:29", 455: "09-json-library/main.go:102:29", 456: "09-json-library/main.go:102:29", 457: "09-json-library/main.go:102:29", 458: "09-json-library/main.go:102:29", 459: "09-json-library/main.go:102:29", 460: "09-json-library/main.go:102:29", 461: "09-json-library/main.go:102:29", 462: "09-json-library/main.go:102:29", 463: "09-json-library/main.go:102:29", 464: "09-json-library/main.go:102:29", 465: "09-json-library/main.go:102:29", 466: "09-json-library/main.go:102:29", 467: "09-json-library/main.go:102:29", 468: "09-json-library/main.go:102:29", 469: "09-json-library/main.go:102:29", 470: "09-json-library/main.go:102:29", 471: "09-json-library/main.go:102:29", 472: "09-json-library/main.go:103:29", 473: "09-json-library/main.go:103:29", 474: "09-json-library/main.go:103:29", 475: "09-json-library/main.go:103:29", 476: "09-json-library/main.go:103:29", 477: "09-json-library/main.go:103:29", 478: "09-json-library/main.go:103:29", 479: "09-json-library/main.go:103:29", 480: "09-json-library/main.go:103:29", 481: "09-json-library/main.go:103:29", 482: "09-json-library/main.go:103:29", 483: "09-json-library/main.go:103:29", 484: "09-json-library/main.go:103:29", 485: "09-json-library/main.go:103:29", 486: "09-json-library/main.go:103:29", 487: "09-json-library/main.go:103:29", 488: "09-json-library/main.go:103:29", 489: "09-json-library/main.go:103:29", 490: "09-json-library/main.go:103:29", 491: "09-json-library/main.go:103:29", 492: "09-json-library/main.go:103:29", 493: "09-json-library/main.go:103:29", 494: "09-json-library/main.go:103:29", 495: "09-json-library/main.go:103:29", 496: "09-json-library/main.go:103:29", 497: "09-json-library/main.go:103:29", 498: "09-json-library/main.go:103:29", 499: "09-json-library/main.go:103:29", 500: "09-json-library/main.go:103:29", 501: "09-json-library/main.go:103:29", 502: "09-json-library/main.go:103:29", 503: "09-json-library/main.go:103:29", 504: "09-json-library/main.go:103:29", 505: "09-json-library/main.go:103:29", 506: "09-json-library/main.go:103:29", 507: "09-json-library/main.go:103:29", 508: "09-json-library/main.go:103:29", 509: "09-json-library/main.go:103:29", 510: "09-json-library/main.go:103:29", 511: "09-json-library/main.go:103:29", 512: "09-json-library/main.go:103:29", 513: "09-json-library/main.go:103:29", 514: "09-json-library/main.go:103:29", 515: "09-json-library/main.go:103:29", 516: "09-json-library/main.go:103:29", 517: "09-json-library/main.go:103:29", 518: "09-json-library/main.go:103:29", 519: "09-json-library/main.go:103:29", 520: "09-json-library/main.go:103:29", 521: "09-json-library/main.go:103:29", 522: "09-json-library/main.go:103:29", 523: "09-json-library/main.go:103:29", 524: "09-json-library/main.go:103:29", 525: "09-json-library/main.go:103:29", 526: "09-json-library/main.go:103:29", 527: "09-json-library/main.go:103:29", 528: "09-json-library/main.go:103:29", 529: "09-json-library/main.go:103:29", 530: "09-json-library/main.go:103:29", 531: "09-json-library/main.go:103:29", 532: "09-json-library/main.go:103:29", 533: "09-json-library/main.go:103:29", 534: "09-json-library/main.go:103:29", 535: "09-json-library/main.go:103:29", 536: "09-json-library/main.go:103:29", 537: "09-json-library/main.go:103:29", 538: "09-json-library/main.go:103:29", 539: "09-json-library/main.go:103:29", 540: "09-json-library/main.go:105:51", 541: "09-json-library/main.go:105:51", 542: "09-json-library/main.go:105:51", 543: "09-json-library/main.go:105:51", 544: "09-json-library/main.go:106:7", 545: "09-json-library/main.go:107:67", 546: "09-json-library/main.go:107:67", 547: "09-json-library/main.go:107:67", 548: "09-json-library/main.go:107:67", 549: "09-json-library/main.go:107:67", 550: "09-json-library/main.go:108:8", 551: "09-json-library/main.go:109:99", 552: "09-json-library/main.go:109:99", 553: "09-json-library/main.go:109:99", 554: "09-json-library/main.go:109:99", 555: "09-json-library/main.go:109:99", 556: "09-json-library/main.go:109:99", 557: "09-json-library/main.go:109:99", 558: "09-json-library/main.go:109:99", 559: "09-json-library/main.go:109:99", 560: "09-json-library/main.go:109:99", 561: "09-json-library/main.go:109:99"});

class Author {
    Name;
    Born;
    constructor() {
        this.Name = "";
        this.Born = 0;
    }
}

go2jsRegisterTypeName(
    Author,
    "main.Author",
    [],
    [
        {Name: "Name", Tag: "json:\"name\"", PkgPath: "", Anonymous: false, Type: "string", Kind: "string"},
        {
            Name: "Born",
            Tag: "json:\"born,omitempty\"",
            PkgPath: "",
            Anonymous: false,
            Type: "int",
            Kind: "int"
        }
    ],
    "struct"
);

class Book {
    ID;
    Title;
    Authors;
    Tags;
    Rating;
    Available;
    Published;
    Extra;
    secret;
    constructor() {
        this.ID = 0;
        this.Title = "";
        this.Authors = null;
        this.Tags = null;
        this.Rating = 0;
        this.Available = false;
        this.Published = "";
        this.Extra = null;
        this.secret = "";
    }
}

go2jsRegisterTypeName(
    Book,
    "main.Book",
    [],
    [
        {Name: "ID", Tag: "json:\"id\"", PkgPath: "", Anonymous: false, Type: "int", Kind: "int"},
        {
            Name: "Title",
            Tag: "json:\"title\"",
            PkgPath: "",
            Anonymous: false,
            Type: "string",
            Kind: "string"
        },
        {
            Name: "Authors",
            Tag: "json:\"authors\"",
            PkgPath: "",
            Anonymous: false,
            Type: "[]main.Author",
            Kind: "slice"
        },
        {
            Name: "Tags",
            Tag: "json:\"tags,omitempty\"",
            PkgPath: "",
            Anonymous: false,
            Type: "[]string",
            Kind: "slice"
        },
        {
            Name: "Rating",
            Tag: "json:\"rating\"",
            PkgPath: "",
            Anonymous: false,
            Type: "float64",
            Kind: "float64"
        },
        {
            Name: "Available",
            Tag: "json:\"available\"",
            PkgPath: "",
            Anonymous: false,
            Type: "bool",
            Kind: "bool"
        },
        {
            Name: "Published",
            Tag: "json:\"published\"",
            PkgPath: "",
            Anonymous: false,
            Type: "string",
            Kind: "string"
        },
        {
            Name: "Extra",
            Tag: "json:\"extra,omitempty\"",
            PkgPath: "",
            Anonymous: false,
            Type: "map[string]string",
            Kind: "map"
        },
        {Name: "secret", Tag: "", PkgPath: "main", Anonymous: false, Type: "string", Kind: "string"}
    ],
    "struct"
);

class Catalogue {
    Name;
    Books;
    constructor() {
        this.Name = "";
        this.Books = null;
    }
}

go2jsRegisterTypeName(
    Catalogue,
    "main.Catalogue",
    [],
    [
        {Name: "Name", Tag: "json:\"name\"", PkgPath: "", Anonymous: false, Type: "string", Kind: "string"},
        {
            Name: "Books",
            Tag: "json:\"books\"",
            PkgPath: "",
            Anonymous: false,
            Type: "[]*main.Book",
            Kind: "slice"
        }
    ],
    "struct"
);

const raw = "{\n  \"name\": \"City Library\",\n  \"books\": [\n    {\"id\": 1, \"title\": \"The Go Programming Language\", \"authors\": [{\"name\": \"Alan Donovan\", \"born\": 1970}, {\"name\": \"Brian Kernighan\", \"born\": 1942}],\n     \"tags\": [\"programming\", \"go\"], \"rating\": 4.7, \"available\": true, \"published\": \"2015-11-16T00:00:00Z\"},\n    {\"id\": 2, \"title\": \"Dune\", \"authors\": [{\"name\": \"Frank Herbert\", \"born\": 1920}], \"tags\": [\"scifi\"],\n     \"rating\": 4.5, \"available\": false, \"published\": \"1965-08-01T00:00:00Z\", \"extra\": {\"series\": \"Dune\", \"volume\": \"1\"}},\n    {\"id\": 3, \"title\": \"Poems\", \"authors\": [{\"name\": \"Anonymous\"}], \"rating\": 3.2, \"available\": true, \"published\": \"1850-01-01T00:00:00Z\"}\n  ]\n}";

function mustDate$(s) {
    let [t, err] = go2jsTimeParse("2006-01-02T15:04:05Z07:00", s);
    if (go2jsEqual(err, null) === false) {
        go2jsPanic(err);
    }
    return go2jsStructCopy(t);
}

function* mustDate(s) {
    return mustDate$(s);
}

function* main() {
    let addr1 = null;
    let addr2 = null;
    let addr3 = null;
    let cat = new Catalogue();
    {
        let err = go2jsJSONUnmarshal(
            go2jsStringToBytes(raw),
            (addr2 ??= go2jsPtr(() => cat, value => cat = value)),
            {"Name": "name", "Books": "books"},
            {
                "kind": "ptr",
                "elem": {
                    "kind": "struct",
                    "name": "main.Catalogue",
                    "go": "main.Catalogue",
                    "fields": {"Name": "name", "Books": "books"},
                    "string": null,
                    "types": {
                        "Name": {"kind": "string", "go": "string"},
                        "Books": {
                            "kind": "slice",
                            "elem": {
                                "kind": "ptr",
                                "elem": {
                                    "kind": "struct",
                                    "name": "main.Book",
                                    "go": "main.Book",
                                    "fields": {
                                        "ID": "id",
                                        "Title": "title",
                                        "Authors": "authors",
                                        "Tags": "tags",
                                        "Rating": "rating",
                                        "Available": "available",
                                        "Published": "published",
                                        "Extra": "extra"
                                    },
                                    "string": null,
                                    "types": {
                                        "ID": {"kind": "number", "go": "int"},
                                        "Title": {"kind": "string", "go": "string"},
                                        "Authors": {
                                            "kind": "slice",
                                            "elem": {
                                                "kind": "struct",
                                                "name": "main.Author",
                                                "go": "main.Author",
                                                "fields": {"Name": "name", "Born": "born"},
                                                "string": null,
                                                "types": {
                                                    "Name": {"kind": "string", "go": "string"},
                                                    "Born": {"kind": "number", "go": "int"}
                                                }
                                            },
                                            "go": "[]main.Author"
                                        },
                                        "Tags": {
                                            "kind": "slice",
                                            "elem": {"kind": "string", "go": "string"},
                                            "go": "[]string"
                                        },
                                        "Rating": {"kind": "number", "go": "float64"},
                                        "Available": {"kind": "bool", "go": "bool"},
                                        "Published": {"kind": "string", "go": "string"},
                                        "Extra": {
                                            "kind": "map",
                                            "elem": {"kind": "string", "go": "string"},
                                            "go": "map[string]string"
                                        }
                                    }
                                }
                            },
                            "go": "[]*main.Book"
                        }
                    }
                }
            },
            null
        )
        if (go2jsEqual(err, null) === false) {
            go2jsPrintln("decode error:", err);
            return;
        }
    }
    go2jsPrintln(cat.Name, go2jsLen(cat.Books), "books");
    let tw = go2jsTabwriterNewWriter(go2jsInterface(go2jsOSStdout(), "*File", "*os.File"), 0, 8, 2, 32, 0);
    go2jsFprint(tw, ["ID\tTITLE\tAUTHORS\tYEAR\tRATING\tIN"], "\n", " ");
    for (const b of go2jsRangeSequence(cat.Books)) {
        let names = null;
        for (const a of go2jsRangeCopies(go2jsRangeSequence(b.Authors))) {
            names = go2jsSliceAppend(names, a.Name);
        }
        let in$go2js = "no";
        if (b.Available) {
            in$go2js = "yes";
        }
        go2jsFprint(
            tw,
            [
                go2jsSprintf(
                    "%d\t%s\t%s\t%d\t%.1f\t%s\n",
                    b.ID,
                    b.Title,
                    go2jsStringsJoin(names, " & "),
                    mustDate$(b.Published).Year(),
                    go2jsTyped(b.Rating, "float64"),
                    in$go2js
                )
            ],
            "",
            ""
        );
    }
    go2jsTabwriterFlush(tw);
    go2jsSortSlice(go2jsNilInterface(cat.Books, "[]*main.Book", "slice", true), function*(i, j) {
        return go2jsIndex(cat.Books, i).Rating > go2jsIndex(cat.Books, j).Rating;
    });
    go2jsPrintln("best:", go2jsIndex(cat.Books, 0).Title);
    cat.Books[1].Available = true;
    cat.Books[2].Extra = go2jsMapTyped(
        "map[string]string",
        go2jsMap([["note", "loaned out"], ["branch", "north"]])
    );
    cat.Books = go2jsSliceAppend(
        cat.Books,
        go2jsNew(
            Object.assign(
                new Book(),
                {ID: 4, Title: "New Arrival", Rating: 0, Published: "2024-02-29T12:30:00Z", secret: "x"}
            )
        )
    );
    let [out, err] = go2jsJSONMarshal(
        go2jsSliceRange(cat.Books, 2, undefined, undefined, false),
        null,
        null,
        null,
        null,
        "",
        "  ",
        {
            "kind": "slice",
            "elem": {
                "kind": "ptr",
                "elem": {
                    "kind": "struct",
                    "name": "main.Book",
                    "go": "main.Book",
                    "fields": {
                        "ID": "id",
                        "Title": "title",
                        "Authors": "authors",
                        "Tags": "tags",
                        "Rating": "rating",
                        "Available": "available",
                        "Published": "published",
                        "Extra": "extra"
                    },
                    "string": null,
                    "types": {
                        "ID": {"kind": "number", "go": "int"},
                        "Title": {"kind": "string", "go": "string"},
                        "Authors": {
                            "kind": "slice",
                            "elem": {
                                "kind": "struct",
                                "name": "main.Author",
                                "go": "main.Author",
                                "fields": {"Name": "name", "Born": "born"},
                                "string": null,
                                "types": {
                                    "Name": {"kind": "string", "go": "string"},
                                    "Born": {"kind": "number", "go": "int"}
                                }
                            },
                            "go": "[]main.Author"
                        },
                        "Tags": {
                            "kind": "slice",
                            "elem": {"kind": "string", "go": "string"},
                            "go": "[]string"
                        },
                        "Rating": {"kind": "number", "go": "float64"},
                        "Available": {"kind": "bool", "go": "bool"},
                        "Published": {"kind": "string", "go": "string"},
                        "Extra": {
                            "kind": "map",
                            "elem": {"kind": "string", "go": "string"},
                            "go": "map[string]string"
                        }
                    }
                }
            },
            "go": "[]*main.Book"
        }
    );
    go2jsPrintln(go2jsBytesToString(out), err);
    let [compact] = go2jsJSONMarshal(
        go2jsMapTyped(
            "map[string]any",
            go2jsMap(
                [
                    ["b", go2jsNilInterface(go2jsSliceTyped("[]int", [1, 2]), "[]int", "slice", true)],
                    ["a", null],
                    ["c", go2jsInterface(1.5, "float64", "float64")],
                    ["d", "x<y"]
                ]
            )
        ),
        null,
        null,
        null,
        null,
        "",
        "",
        {"kind": "map", "elem": {"kind": "any"}, "go": "map[string]any"}
    );
    go2jsPrintln(go2jsBytesToString(compact));
    let generic = null;
    go2jsJSONUnmarshal(
        go2jsStringToBytes("{\"n\": 3, \"list\": [1, \"two\", true, null], \"obj\": {\"k\": \"v\"}}"),
        (addr3 ??= go2jsPtr(() => generic, value => generic = value)),
        null,
        {"kind": "ptr", "elem": {"kind": "map", "elem": {"kind": "any"}, "go": "map[string]any"}},
        null
    );
    let keys = go2jsSliceMake(0, 0, "", true);
    for (const k of go2jsRangeMap(generic).keys()) {
        keys = go2jsSliceAppend(keys, k);
    }
    go2jsSortStrings(keys);
    for (const k of go2jsRangeSequence(keys)) {
        go2jsWriteOut(
            process.stdout,
            go2jsSprintf(
                "%s: %v (%T)\n",
                k,
                go2jsTyped(go2jsMapGet(generic, k, null), "interface {}"),
                go2jsTyped(go2jsMapGet(generic, k, null), "interface {}")
            )
        );
    }
    let bad = new Book();
    go2jsPrintln(
        go2jsJSONUnmarshal(
            go2jsStringToBytes("{\"id\": \"seven\"}"),
            (addr1 ??= go2jsPtr(() => bad, value => bad = value)),
            {
                "ID": "id",
                "Title": "title",
                "Authors": "authors",
                "Tags": "tags",
                "Rating": "rating",
                "Available": "available",
                "Published": "published",
                "Extra": "extra"
            },
            {
                "kind": "ptr",
                "elem": {
                    "kind": "struct",
                    "name": "main.Book",
                    "go": "main.Book",
                    "fields": {
                        "ID": "id",
                        "Title": "title",
                        "Authors": "authors",
                        "Tags": "tags",
                        "Rating": "rating",
                        "Available": "available",
                        "Published": "published",
                        "Extra": "extra"
                    },
                    "string": null,
                    "types": {
                        "ID": {"kind": "number", "go": "int"},
                        "Title": {"kind": "string", "go": "string"},
                        "Authors": {
                            "kind": "slice",
                            "elem": {
                                "kind": "struct",
                                "name": "main.Author",
                                "go": "main.Author",
                                "fields": {"Name": "name", "Born": "born"},
                                "string": null,
                                "types": {
                                    "Name": {"kind": "string", "go": "string"},
                                    "Born": {"kind": "number", "go": "int"}
                                }
                            },
                            "go": "[]main.Author"
                        },
                        "Tags": {
                            "kind": "slice",
                            "elem": {"kind": "string", "go": "string"},
                            "go": "[]string"
                        },
                        "Rating": {"kind": "number", "go": "float64"},
                        "Available": {"kind": "bool", "go": "bool"},
                        "Published": {"kind": "string", "go": "string"},
                        "Extra": {
                            "kind": "map",
                            "elem": {"kind": "string", "go": "string"},
                            "go": "map[string]string"
                        }
                    }
                }
            },
            null
        )
    );
    go2jsPrintln(
        go2jsJSONUnmarshal(
            go2jsStringToBytes("{\"id\": 7,"),
            (addr1 ??= go2jsPtr(() => bad, value => bad = value)),
            {
                "ID": "id",
                "Title": "title",
                "Authors": "authors",
                "Tags": "tags",
                "Rating": "rating",
                "Available": "available",
                "Published": "published",
                "Extra": "extra"
            },
            {
                "kind": "ptr",
                "elem": {
                    "kind": "struct",
                    "name": "main.Book",
                    "go": "main.Book",
                    "fields": {
                        "ID": "id",
                        "Title": "title",
                        "Authors": "authors",
                        "Tags": "tags",
                        "Rating": "rating",
                        "Available": "available",
                        "Published": "published",
                        "Extra": "extra"
                    },
                    "string": null,
                    "types": {
                        "ID": {"kind": "number", "go": "int"},
                        "Title": {"kind": "string", "go": "string"},
                        "Authors": {
                            "kind": "slice",
                            "elem": {
                                "kind": "struct",
                                "name": "main.Author",
                                "go": "main.Author",
                                "fields": {"Name": "name", "Born": "born"},
                                "string": null,
                                "types": {
                                    "Name": {"kind": "string", "go": "string"},
                                    "Born": {"kind": "number", "go": "int"}
                                }
                            },
                            "go": "[]main.Author"
                        },
                        "Tags": {
                            "kind": "slice",
                            "elem": {"kind": "string", "go": "string"},
                            "go": "[]string"
                        },
                        "Rating": {"kind": "number", "go": "float64"},
                        "Available": {"kind": "bool", "go": "bool"},
                        "Published": {"kind": "string", "go": "string"},
                        "Extra": {
                            "kind": "map",
                            "elem": {"kind": "string", "go": "string"},
                            "go": "map[string]string"
                        }
                    }
                }
            },
            null
        )
    );
    let [first, last] = [
        go2jsStructCopy(mustDate$(go2jsIndex(cat.Books, 0).Published)),
        go2jsStructCopy(mustDate$(go2jsIndex(cat.Books, 3).Published))
    ];
    let d = first.Sub(go2jsStructCopy(last));
    go2jsPrintln(
        first.Format("Mon Jan 2 2006"),
        go2jsTyped(first.Weekday(), "time.Weekday", "int"),
        go2jsWideToSigned(DurationHours(d) / 24 / 365)
    );
    let t0 = go2jsStructCopy(go2jsTimeDate(2024, 1, 31, 0, 0, 0, 0, go2jsTimeLocation("UTC")));
    go2jsPrintln(
        t0.AddDate(0, 1, 0).Format("2006-01-02"),
        t0.Add(go2jsDuration(go2jsWideMul(36, go2jsDuration(3600000000000), "int64"))).Format(
            "2006-01-02T15:04:05Z07:00"
        ),
        go2jsTyped(
            go2jsDuration(go2jsWideMul(go2jsDuration(5400), go2jsDuration(1000000000), "int64")),
            "time.Duration",
            "int64"
        )
    );
}

try {
    go2jsRunMain((function*() {
        return yield* main();
    })());
} catch (e) {
    go2jsReportUncaughtPanic(e);
}
