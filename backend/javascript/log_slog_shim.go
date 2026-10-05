package javascript

import (
	"go/ast"
	"strconv"
	"strings"

	gotypesstd "go/types"
)

// slogFuncs maps the package level functions of log/slog onto the runtime
// helpers that stand in for them.
var slogFuncs = map[string]string{
	"Any":               "go2jsSlogAny",
	"Bool":              "go2jsSlogBool",
	"Duration":          "go2jsSlogDuration",
	"Float64":           "go2jsSlogFloat64",
	"Group":             "go2jsSlogGroup",
	"Int":               "go2jsSlogInt",
	"Int64":             "go2jsSlogInt64",
	"String":            "go2jsSlogString",
	"Time":              "go2jsSlogTime",
	"Uint64":            "go2jsSlogUint64",
	"AnyValue":          "go2jsSlogAnyValue",
	"BoolValue":         "go2jsSlogBoolValue",
	"DurationValue":     "go2jsSlogDurationValue",
	"Float64Value":      "go2jsSlogFloat64Value",
	"GroupValue":        "go2jsSlogGroupValue",
	"Int64Value":        "go2jsSlogInt64Value",
	"IntValue":          "go2jsSlogIntValue",
	"StringValue":       "go2jsSlogStringValue",
	"TimeValue":         "go2jsSlogTimeValue",
	"Uint64Value":       "go2jsSlogUint64Value",
	"NewRecord":         "go2jsSlogNewRecord",
	"NewJSONHandler":    "go2jsSlogNewJSONHandler",
	"NewTextHandler":    "go2jsSlogNewTextHandler",
	"NewLogLogger":      "go2jsSlogNewLogLogger",
	"New":               "go2jsSlogNew",
	"With":              "go2jsSlogWith",
	"Default":           "go2jsSlogDefault",
	"SetDefault":        "go2jsSlogSetDefault",
	"SetLogLoggerLevel": "go2jsSlogSetLogLoggerLevel",
	"Debug":             "go2jsSlogDebug",
	"DebugContext":      "go2jsSlogDebugContext",
	"Info":              "go2jsSlogInfo",
	"InfoContext":       "go2jsSlogInfoContext",
	"Warn":              "go2jsSlogWarn",
	"WarnContext":       "go2jsSlogWarnContext",
	"Error":             "go2jsSlogError",
	"ErrorContext":      "go2jsSlogErrorContext",
	"Log":               "go2jsSlogLog",
	"LogAttrs":          "go2jsSlogLogAttrs",
}

// slogMethods maps the methods of the types of log/slog onto the runtime
// helpers that stand in for them. The key is the path of the package, the name
// of the type and the name of the method, which is how a call on a value of a
// type from a package is told apart from a call on a value of the program.
var slogMethods = map[string]string{
	"log/slog.Attr.String": "go2jsSlogAttrStringMethod",
	"log/slog.Attr.Equal":  "go2jsSlogAttrEqualMethod",
	"log/slog.Attr.Value":  "go2jsSlogAttrValueMethod",

	"log/slog.Kind.String": "go2jsSlogKindName",

	"log/slog.Level.String":        "go2jsSlogLevelStringMethod",
	"log/slog.Level.MarshalText":   "go2jsSlogLevelStringMethod",
	"log/slog.Level.MarshalJSON":   "go2jsSlogLevelMarshalJSONMethod",
	"log/slog.Level.UnmarshalText": "go2jsSlogLevelUnmarshalTextMethod",
	"log/slog.Level.UnmarshalJSON": "go2jsSlogLevelUnmarshalJSONMethod",

	"log/slog.LevelVar.Level":          "go2jsSlogLevelVarLevel",
	"log/slog.LevelVar.Set":            "go2jsSlogLevelVarSet",
	"log/slog.LevelVar.String":         "go2jsSlogLevelVarString",
	"log/slog.LevelVar.Swap":           "go2jsSlogLevelVarSwap",
	"log/slog.LevelVar.CompareAndSwap": "go2jsSlogLevelVarCompareAndSwap",

	"log/slog.Value.Kind":       "go2jsSlogValueKindMethod",
	"log/slog.Value.Any":        "go2jsSlogValueAnyMethod",
	"log/slog.Value.Bool":       "go2jsSlogValueBoolMethod",
	"log/slog.Value.Duration":   "go2jsSlogValueDurationMethod",
	"log/slog.Value.Float64":    "go2jsSlogValueFloat64Method",
	"log/slog.Value.Int64":      "go2jsSlogValueInt64Method",
	"log/slog.Value.Uint64":     "go2jsSlogValueUint64Method",
	"log/slog.Value.Time":       "go2jsSlogValueTimeMethod",
	"log/slog.Value.Group":      "go2jsSlogValueGroupMethod",
	"log/slog.Value.GroupValue": "go2jsSlogValueGroupValueMethod",
	"log/slog.Value.LogValue":   "go2jsSlogValueLogValueMethod",
	"log/slog.Value.String":     "go2jsSlogValueTextMethod",
	"log/slog.Value.Equal":      "go2jsSlogValueEqualMethod",
	"log/slog.Value.Resolve":    "go2jsSlogValueResolveMethod",

	"log/slog.Record.Clone":    "go2jsSlogRecordCloneMethod",
	"log/slog.Record.NumAttrs": "go2jsSlogRecordNumAttrsMethod",
	"log/slog.Record.Attrs":    "go2jsSlogRecordAttrsMethod",
	"log/slog.Record.Add":      "go2jsSlogRecordAdd",
	"log/slog.Record.AddAttrs": "go2jsSlogRecordAddAttrs",

	"log/slog.Logger.Handler":      "go2jsSlogLoggerHandlerMethod",
	"log/slog.Logger.Enabled":      "go2jsSlogLoggerEnabledMethod",
	"log/slog.Logger.WithGroup":    "go2jsSlogLoggerWithGroupMethod",
	"log/slog.Logger.LogAttrs":     "go2jsSlogLoggerLogAttrs",
	"log/slog.Logger.Log":          "go2jsSlogLoggerLog",
	"log/slog.Logger.With":         "go2jsSlogLoggerWith",
	"log/slog.Logger.Debug":        "go2jsSlogLoggerDebug",
	"log/slog.Logger.DebugContext": "go2jsSlogLoggerDebugContext",
	"log/slog.Logger.Info":         "go2jsSlogLoggerInfo",
	"log/slog.Logger.InfoContext":  "go2jsSlogLoggerInfoContext",
	"log/slog.Logger.Warn":         "go2jsSlogLoggerWarn",
	"log/slog.Logger.WarnContext":  "go2jsSlogLoggerWarnContext",
	"log/slog.Logger.Error":        "go2jsSlogLoggerError",
	"log/slog.Logger.ErrorContext": "go2jsSlogLoggerErrorContext",

	"log/slog.Source.String": "go2jsSlogSourceString",
}

// slogKindNameMethods are the methods whose trailing arguments are values rather
// than attributes, so the kind of each of them is known from the program and not
// from what the value turns out to look like at run time. The number is how
// many arguments come before the ones that are named here.
var slogKindNameMethods = map[string]int{
	"log/slog.Logger.Log":          3,
	"log/slog.Logger.Debug":        1,
	"log/slog.Logger.DebugContext": 2,
	"log/slog.Logger.Info":         1,
	"log/slog.Logger.InfoContext":  2,
	"log/slog.Logger.Warn":         1,
	"log/slog.Logger.WarnContext":  2,
	"log/slog.Logger.Error":        1,
	"log/slog.Logger.ErrorContext": 2,
	"log/slog.Logger.With":         0,
	"log/slog.Record.Add":          0,
}

// slogKindNameFuncs are the package level functions whose trailing arguments
// are values rather than attributes, with the number of arguments that come
// before them.
var slogKindNameFuncs = map[string]int{
	"Debug":        1,
	"DebugContext": 2,
	"Info":         1,
	"InfoContext":  2,
	"Warn":         1,
	"WarnContext":  2,
	"Error":        1,
	"ErrorContext": 2,
	"Log":          3,
	"With":         0,
	"Group":        1,
	"AnyValue":     0,
}

// slogSourcedMethods are the methods that write a record, and so are told where
// the program stands, which is where a handler asked for the source of a record
// says it was made.
var slogSourcedMethods = map[string]bool{
	"log/slog.Logger.Log":          true,
	"log/slog.Logger.LogAttrs":     true,
	"log/slog.Logger.Debug":        true,
	"log/slog.Logger.DebugContext": true,
	"log/slog.Logger.Info":         true,
	"log/slog.Logger.InfoContext":  true,
	"log/slog.Logger.Warn":         true,
	"log/slog.Logger.WarnContext":  true,
	"log/slog.Logger.Error":        true,
	"log/slog.Logger.ErrorContext": true,
}

// slogSourcedFuncs are the package level functions that write a record.
var slogSourcedFuncs = map[string]bool{
	"Debug":        true,
	"DebugContext": true,
	"Info":         true,
	"InfoContext":  true,
	"Warn":         true,
	"WarnContext":  true,
	"Error":        true,
	"ErrorContext": true,
	"Log":          true,
	"LogAttrs":     true,
}

func init() {
	stdlibFuncMaps["log/slog"] = slogFuncs
	supportedStdlibPackages["log/slog"] = funcSet(slogFuncs)

	// A type of a package is found by the name of the package and the name of
	// the type itself, which for this package are slog.Logger and the rest
	// rather than anything written out of the path it is imported by.
	packageTypes["slog.Logger"] = "go2jsSlogLogger"
	packageTypes["slog.Record"] = "go2jsSlogRecord"
	packageTypes["slog.Value"] = "go2jsSlogValue"
	packageTypes["slog.Attr"] = "go2jsSlogAttr"
	packageTypes["slog.LevelVar"] = "go2jsSlogLevelVar"
	packageTypes["slog.HandlerOptions"] = "go2jsSlogHandlerOptions"
	packageTypeNew["slog.Record"] = true
	packageTypeNew["slog.Value"] = true
	packageTypeNew["slog.Attr"] = true
	packageTypeNew["slog.LevelVar"] = true
	packageTypeNew["slog.HandlerOptions"] = true

	for name, helper := range slogMethods {
		shimValueMethods[name] = helper
	}

	// A level is a whole number with a name, and the names are written where
	// both the path of the package and the name of the package can find them.
	slogConstants := map[string]string{
		"LevelDebug":    "-4",
		"LevelInfo":     "0",
		"LevelWarn":     "4",
		"LevelError":    "8",
		"KindAny":       "0",
		"KindBool":      "1",
		"KindDuration":  "2",
		"KindFloat64":   "3",
		"KindInt64":     "4",
		"KindString":    "5",
		"KindTime":      "6",
		"KindUint64":    "7",
		"KindGroup":     "8",
		"KindLogValuer": "9",
		"TimeKey":       strconv.Quote("time"),
		"LevelKey":      strconv.Quote("level"),
		"MessageKey":    strconv.Quote("msg"),
		"SourceKey":     strconv.Quote("source"),
	}

	for name, value := range slogConstants {
		packageConstants["slog."+name] = value
		packageConstants["log/slog."+name] = value
	}
}

// emitSlogCall writes a call on a type of log/slog. The kinds of the values a
// call carries as its trailing arguments are told to the runtime, since a number
// in JavaScript is a number whichever of the kinds Go gave it the program meant,
// and a moment or a span of time is a value of a kind JavaScript has none of.
func (e *emitter) emitSlogCall(call *ast.CallExpr, selector *ast.SelectorExpr) (bool, error) {
	key, ok := shimMethodKey(selector, e.analyzedType(selector.X))
	if !ok {
		return false, nil
	}

	helper, known := slogMethods[key]
	if !known {
		return false, nil
	}

	variadic := false
	skip := 0
	usesKinds := false

	if kinds, ok := slogKindNameMethods[key]; ok {
		variadic = true
		skip = kinds
		usesKinds = true
	}

	if skip > len(call.Args) {
		return false, nil
	}

	e.needsRuntime = true

	// Special handling for LogAttrs and AddAttrs
	// The level and the message are named by the program, and what comes after
	// them is a list of arguments of its own length, so it is written as one.
	if key == "log/slog.Logger.Log" {
		e.write(helper)
		e.write("(")

		if err := e.emitExpr(selector.X); err != nil {
			return true, err
		}

		for i := 0; i < 3 && i < len(call.Args); i++ {
			e.write(", ")

			if err := e.emitSlogArgument(call.Args[i]); err != nil {
				return true, err
			}
		}

		e.write(", ")

		if len(call.Args) > 3 {
			e.write("[")

			for i := 3; i < len(call.Args); i++ {
				if i > 3 {
					e.write(", ")
				}

				if err := e.emitSlogArgument(call.Args[i]); err != nil {
					return true, err
				}
			}

			e.write("]")
		} else {
			e.write("[]")
		}

		e.write(", ")

		if err := e.emitSlogValueTypes(call.Args[3:]); err != nil {
			return true, err
		}

		e.write(", ")
		e.write(strconv.Quote(e.slogSourceOf(call)))
		e.write(")")
		return true, nil
	}

	if key == "log/slog.Logger.LogAttrs" {
		e.write(helper)
		e.write("(")
		if err := e.emitExpr(selector.X); err != nil {
			return true, err
		}
		// receiver, ctx, level, msg, then attrs...
		for i := 0; i < 3 && i < len(call.Args); i++ {
			e.write(", ")
			if err := e.emitSlogArgument(call.Args[i]); err != nil {
				return true, err
			}
		}
		if len(call.Args) > 3 {
			e.write(", [")
			for i := 3; i < len(call.Args); i++ {
				if i > 3 {
					e.write(", ")
				}
				if err := e.emitSlogArgument(call.Args[i]); err != nil {
					return true, err
				}
			}
			e.write("]")
		} else {
			e.write(", []")
		}
		if slogSourcedMethods[key] {
			e.write(", ")
			e.write(strconv.Quote(e.slogSourceOf(call)))
		}
		e.write(")")
		return true, nil
	}

	if key == "log/slog.Record.AddAttrs" {
		e.write(helper)
		e.write("(")
		if err := e.emitExpr(selector.X); err != nil {
			return true, err
		}
		if len(call.Args) > 0 {
			e.write(", [")
			for i := 0; i < len(call.Args); i++ {
				if i > 0 {
					e.write(", ")
				}
				if err := e.emitSlogArgument(call.Args[i]); err != nil {
					return true, err
				}
			}
			e.write("]")
		} else {
			e.write(", []")
		}
		e.write(")")
		return true, nil
	}

	e.write(helper)
	e.write("(")

	if err := e.emitExpr(selector.X); err != nil {
		return true, err
	}

	for _, arg := range call.Args {
		e.write(", ")

		if err := e.emitSlogArgument(arg); err != nil {
			return true, err
		}
	}

	if variadic && usesKinds {
		if len(call.Args) > 0 {
			e.write(", ")
		}

		if err := e.emitSlogValueTypes(call.Args[skip:]); err != nil {
			return true, err
		}
	}

	if slogSourcedMethods[key] {
		e.write(", ")
		e.write(strconv.Quote(e.slogSourceOf(call)))
	}

	e.write(")")

	return true, nil
}

// emitSlogPackageCall writes a call on a function of log/slog.
func (e *emitter) emitSlogPackageCall(call *ast.CallExpr, selector *ast.SelectorExpr) (bool, error) {
	pkg, ok := e.packageImportPath(selector)
	if !ok || pkg != "log/slog" {
		return false, nil
	}

	name := selector.Sel.Name

	helper, known := slogFuncs[name]
	if !known {
		return false, nil
	}

	skip, variadic := slogKindNameFuncs[name]
	if !variadic && !slogSourcedFuncs[name] {
		return false, nil
	}

	if skip > len(call.Args) {
		return false, nil
	}

	e.needsRuntime = true

	e.write(helper)
	e.write("(")

	for index, arg := range call.Args {
		if index > 0 {
			e.write(", ")
		}

		if err := e.emitSlogArgument(arg); err != nil {
			return true, err
		}
	}

	if variadic {
		if len(call.Args) > 0 {
			e.write(", ")
		}

		if err := e.emitSlogValueTypes(call.Args[skip:]); err != nil {
			return true, err
		}
	}

	if slogSourcedFuncs[name] {
		e.write(", ")
		e.write(strconv.Quote(e.slogSourceOf(call)))
	}

	e.write(")")

	return true, nil
}

// emitSlogArgument writes one argument of a call on log/slog. A value of a type
// the runtime cannot read off the value itself is wrapped in a box naming that
// type, since the way a value is written out later depends on what Go would
// have written it as rather than on what it happens to look like.
func (e *emitter) emitSlogArgument(arg ast.Expr) error {
	return e.emitTypedValue(arg)
}

// emitSlogValueTypes writes the kinds of the values a call carries as its
// trailing arguments, as the names of the types the program gave them.
func (e *emitter) emitSlogValueTypes(args []ast.Expr) error {
	e.write("[")

	for index, arg := range args {
		if index > 0 {
			e.write(", ")
		}

		e.write(strconv.Quote(slogTypeName(e.analyzedType(arg))))
	}

	e.write("]")

	return nil
}

// slogTypeName is the name Go gives a type, which is the name the runtime knows
// a kind by.
func slogTypeName(value gotypesstd.Type) string {
	if value == nil {
		return ""
	}

	return value.String()
}

// slogSourceOf is where the program stands when it asks for a record to be
// written, which is what a handler told to add the source of a record says.
func (e *emitter) slogSourceOf(call *ast.CallExpr) string {
	if e.semantic == nil {
		return ""
	}

	position := e.semantic.Position(call.Pos())
	if position.Filename == "" {
		return ""
	}

	return strings.TrimPrefix(position.Filename, "./") + ":" + strconv.Itoa(position.Line)
}

func slogRuntimeSource() string {
	return `
// The kinds of value a record can carry are the kinds Go names, in the order it
// names them, where any comes first because a value nobody told what it holds
// holds nothing in particular.
const go2jsSlogKind = {
	Any: 0,
	Bool: 1,
	Duration: 2,
	Float64: 3,
	Int64: 4,
	String: 5,
	Time: 6,
	Uint64: 7,
	Group: 8,
	LogValuer: 9,
};

function go2jsSlogKindName(kind) {
	switch (kind) {
	case go2jsSlogKind.Any: return "Any";
	case go2jsSlogKind.Bool: return "Bool";
	case go2jsSlogKind.Duration: return "Duration";
	case go2jsSlogKind.Float64: return "Float64";
	case go2jsSlogKind.Int64: return "Int64";
	case go2jsSlogKind.String: return "String";
	case go2jsSlogKind.Time: return "Time";
	case go2jsSlogKind.Uint64: return "Uint64";
	case go2jsSlogKind.Group: return "Group";
	case go2jsSlogKind.LogValuer: return "LogValuer";
	}

	return "<unknown slog.Kind>";
}

// go2jsSlogLevelName is the name a level is written under, which is the name of
// the level itself when it is one of the four Go names and the name of the level
// it stands beside followed by how far it stands from it when it is not, the way
// Go writes DEBUG+1 for a level one step above DEBUG.
function go2jsSlogLevelName(level) {
	const held = Number(level) | 0;
	const named = function(base, from) {
		const off = held - from;

		return off === 0 ? base : base + (off < 0 ? "" : "+") + off;
	};

	if (held < go2jsSlogKind.Any) {
		return named("DEBUG", -4);
	}

	if (held < 4) {
		return named("INFO", 0);
	}

	if (held < 8) {
		return named("WARN", 4);
	}

	return named("ERROR", 8);
}

// go2jsSlogLevelParse is the level a name stands for, where a name is one of the
// four Go names with the number of steps beside it, written the way Go writes
// it, and anything else is not a level at all.
function go2jsSlogLevelParse(text) {
	const held = String(text);
	const at = held.search(/[+-]/);
	const name = (at < 0 ? held : held.slice(0, at)).toUpperCase();
	const rest = at < 0 ? "" : held.slice(at);
	const step = rest === "" ? 0 : Number(rest);

	if (rest !== "" && !Number.isInteger(step)) {
		return null;
	}

	switch (name) {
	case "DEBUG": return -4 + step;
	case "INFO": return 0 + step;
	case "WARN": return 4 + step;
	case "ERROR": return 8 + step;
	}

	return null;
}

// go2jsSlogLevelOf is the level a holder of one reports, which is what a handler
// asked for the lowest level it writes asks for.
function go2jsSlogLevelOf(leveler) {
	leveler = go2jsUnwrap(go2jsUntyped(leveler));
	if (leveler !== null && leveler !== undefined && typeof leveler.Level === "function") {
		return leveler.Level() | 0;
	}

	return Number(leveler) | 0;
}

// go2jsSlogLevelNumber is a level as the number behind its name, read through
// whatever it was wrapped in on the way here.
function go2jsSlogLevelNumber(level) {
	return Number(go2jsUnwrap(go2jsUntyped(level))) | 0;
}

// go2jsSlogValue holds one value of a record together with the kind of value it
// is, since a number in JavaScript is a number whichever kind Go gave it, and a
// moment or a span of time is a value JavaScript has no kind of its own for.
function go2jsSlogValue(kind, num, str, any) {
	// A value built with nothing said to it is a value of no kind in particular,
	// which is what a value of the program that was never filled in is.
	this.kind = kind === undefined ? go2jsSlogKind.Any : kind;
	this.num = num === undefined ? 0 : num;
	this.str = str === undefined ? "" : str;
	this.any = any === undefined ? null : any;
}

go2jsSlogValue.prototype.Kind = function() {
	return this.kind;
};

go2jsSlogValue.prototype.Any = function() {
	switch (this.kind) {
	case go2jsSlogKind.String:
		return this.str;
	case go2jsSlogKind.Int64:
	case go2jsSlogKind.Uint64:
	case go2jsSlogKind.Duration:
		return Number(this.num);
	case go2jsSlogKind.Float64:
		return Number(this.num);
	case go2jsSlogKind.Bool:
		return this.num === true || this.num === 1;
	case go2jsSlogKind.Time:
		return this.any;
	case go2jsSlogKind.Group:
		return this.any.slice();
	}

	return this.any;
};

// A reader that is handed a value of another kind than the one it reads is a
// reader that was asked the wrong question, which Go answers by panicking.
go2jsSlogValue.prototype.Bool = function() {
	go2jsSlogValueWantKind(this, go2jsSlogKind.Bool);

	return this.num === true || this.num === 1;
};

go2jsSlogValue.prototype.Duration = function() {
	go2jsSlogValueWantKind(this, go2jsSlogKind.Duration);

	return go2jsDuration(this.num);
};

go2jsSlogValue.prototype.Float64 = function() {
	go2jsSlogValueWantKind(this, go2jsSlogKind.Float64);

	return Number(this.num);
};

go2jsSlogValue.prototype.Int64 = function() {
	go2jsSlogValueWantKind(this, go2jsSlogKind.Int64);

	return Number(this.num);
};

go2jsSlogValue.prototype.Uint64 = function() {
	go2jsSlogValueWantKind(this, go2jsSlogKind.Uint64);

	return Number(this.num);
};

go2jsSlogValue.prototype.Time = function() {
	go2jsSlogValueWantKind(this, go2jsSlogKind.Time);

	return this.any;
};

go2jsSlogValue.prototype.Group = function() {
	go2jsSlogValueWantKind(this, go2jsSlogKind.Group);

	return this.any;
};

// go2jsSlogValueWantKind is the complaint of a reader that was handed a value of
// another kind than the one it reads.
function go2jsSlogValueWantKind(value, wanted) {
	if (value.kind !== wanted) {
		go2jsPanic("Value kind is " + go2jsSlogKindName(value.kind) + ", not " +
			go2jsSlogKindName(wanted));
	}
}

go2jsSlogValue.prototype.GroupValue = function() {
	return this.any.slice();
};

go2jsSlogValue.prototype.LogValue = function() {
	return this.any;
};

go2jsSlogValue.prototype.String = function() {
	return go2jsSlogValueText(this);
};

go2jsSlogValue.prototype.Equal = function(other) {
	other = go2jsSlogUnwrapValue(other);

	if (other === null || other === undefined || other.kind !== this.kind) {
		return false;
	}

	if (this.kind === go2jsSlogKind.String) {
		return this.str === other.str;
	}

	if (this.kind === go2jsSlogKind.Group) {
		if (this.any.length !== other.any.length) {
			return false;
		}

		return this.any.every((attr, index) => go2jsSlogAttrEqual(attr, other.any[index]));
	}

	if (this.kind === go2jsSlogKind.Time) {
		return go2jsSlogTimeText(this.any) === go2jsSlogTimeText(other.any);
	}

	return go2jsSlogValueText(this) === go2jsSlogValueText(other);
};

go2jsSlogValue.prototype.Resolve = function() {
	let value = this;

	// A value the program asked to be asked again is asked again and again, and
	// a value that keeps asking stops at the point where Go stops asking too, so
	// that a value which never settles is answered rather than waited on.
	for (let count = 0; count < 100; count++) {
		if (value.kind !== go2jsSlogKind.LogValuer) {
			return value;
		}

		value = go2jsSlogLogValue(value.any);
	}

	return go2jsSlogAnyValueText("LogValue called too many times on Value of type " + go2jsSlogTypeNameOf(this.any));
};

// go2jsSlogLogValue asks a value of the program for the value it stands for.
function go2jsSlogLogValue(valuer) {
	try {
		valuer = go2jsUnwrap(go2jsUntyped(valuer));

		return go2jsSlogAsValue(go2jsSlogInvoke(valuer.LogValue, valuer, []));
	} catch (err) {
		// What the value said on its way down is not what is reported, since
		// what it did not reach says more than what it said.
		return go2jsSlogAnyValueText("LogValue panicked\n" + go2jsSlogPanicSite(valuer));
	}
}

// go2jsSlogPanicSite is where a value that would not answer was asked, since a
// fault a value raised says nothing of the fault that raised it.
function go2jsSlogPanicSite(valuer) {
	const name = go2jsSlogTypeNameOf(valuer);

	if (name === "<nil>" || name === "") {
		return "";
	}

	return "called from " + name;
}

// go2jsSlogAsValue reads a value that answers for being one, which is how a
// value the program made with the constructor of a Value of the package comes
// back holding what it holds.
function go2jsSlogAsValue(value) {
	if (value instanceof go2jsSlogValue) {
		return value;
	}

	return go2jsSlogAnyValue(value, null);
}

// go2jsSlogTypeNameOf is the name Go gives a value, which is the name a value
// that writes itself as text is asked for under.
function go2jsSlogTypeNameOf(value) {
	if (value === null || value === undefined) {
		return "<nil>";
	}

	const named = go2jsGoTypeName(value);

	if (named !== undefined && named !== null && named !== "") {
		return String(named);
	}

	return typeof value;
}

// go2jsSlogValueText is a value written the way Go writes one with fmt.Sprint,
// which is how a value of a kind of its own is written inside a group.
function go2jsSlogValueText(value) {
	switch (value.kind) {
	case go2jsSlogKind.String:
		return value.str;
	case go2jsSlogKind.Int64:
	case go2jsSlogKind.Uint64:
		return String(value.num);
	case go2jsSlogKind.Float64:
		return go2jsSprintf("%v", Number(value.num));
	case go2jsSlogKind.Bool:
		return value.num === true || value.num === 1 ? "true" : "false";
	case go2jsSlogKind.Duration:
		return go2jsDurationStringBig(BigInt(value.num));
	case go2jsSlogKind.Time:
		return go2jsSlogTimeText(value.any);
	case go2jsSlogKind.Group:
		return "[" + value.any.map(attr => attr.String()).join(" ") + "]";
	case go2jsSlogKind.Any:
	case go2jsSlogKind.LogValuer:
		return go2jsSprintf("%v", value.any);
	}

	return "<unknown slog.Value>";
}

function go2jsSlogStringValue(value) {
	return new go2jsSlogValue(go2jsSlogKind.String, 0,
		go2jsRawText(go2jsSlogTextOf(value)), null);
}

function go2jsSlogIntValue(value) {
	return new go2jsSlogValue(go2jsSlogKind.Int64, go2jsSlogBigInt(value), "", null);
}

function go2jsSlogInt64Value(value) {
	return new go2jsSlogValue(go2jsSlogKind.Int64, go2jsSlogBigInt(go2jsUnwrap(go2jsUntyped(value))), "", null);
}

function go2jsSlogUint64Value(value) {
	return new go2jsSlogValue(go2jsSlogKind.Uint64, go2jsSlogBigInt(go2jsUnwrap(go2jsUntyped(value))), "", null);
}

function go2jsSlogFloat64Value(value) {
	return new go2jsSlogValue(go2jsSlogKind.Float64, Number(go2jsUnwrap(go2jsUntyped(value))), "", null);
}

function go2jsSlogBoolValue(value) {
	value = go2jsUnwrap(go2jsUntyped(value));

	return new go2jsSlogValue(go2jsSlogKind.Bool, value === true || value === 1, "", null);
}

function go2jsSlogTimeValue(value) {
	return new go2jsSlogValue(go2jsSlogKind.Time, 0, "",
		go2jsUnwrap(go2jsUntyped(value)));
}

function go2jsSlogDurationValue(value) {
	return new go2jsSlogValue(go2jsSlogKind.Duration, go2jsSlogBigInt(go2jsSlogNumberOf(value)), "", null);
}

// The attributes of a group of values are given one after another, or as the
// one list of them the program already had, since either says the same thing.
function go2jsSlogGroupValue(...attrs) {
	return new go2jsSlogValue(go2jsSlogKind.Group, 0, "", go2jsSlogFlattenAttrs(attrs));
}

function go2jsSlogFlattenAttrs(attrs) {
	const flat = [];

	for (const attr of attrs) {
		if (Array.isArray(attr)) {
			flat.push(...attr);
			continue;
		}

		if (attr !== undefined) {
			flat.push(attr);
		}
	}

	return flat;
}

// go2jsSlogAnyValue is a value read out of one of Go's own shapes, which the
// kinds of the program are told apart by, and which a value nothing was said
// about is read the way the shape it has reads itself.
function go2jsSlogAnyValue(value, types) {
	const type = types === null || types === undefined ? "" : String(types);

	switch (type) {
	case "string":
		return go2jsSlogStringValue(value);
	case "bool":
		return go2jsSlogBoolValue(value);
	case "int":
	case "int8":
	case "int16":
	case "int32":
	case "int64":
	case "untyped int":
	case "untyped rune":
		return go2jsSlogInt64Value(value);
	case "uint":
	case "uint8":
	case "uint16":
	case "uint32":
	case "uint64":
	case "uintptr":
	case "untyped uint":
		return go2jsSlogUint64Value(value);
	case "float32":
	case "float64":
	case "untyped float":
	case "untyped complex":
		return go2jsSlogFloat64Value(value);
	case "time.Duration":
		return go2jsSlogDurationValue(value);
	case "time.Time":
		return go2jsSlogTimeValue(value);
	case "slog.Value":
	case "log/slog.Value":
		value = go2jsUnwrap(go2jsUntyped(value));

		return value instanceof go2jsSlogValue ? value : go2jsSlogAnyValueText(String(value));
	case "slog.Attr":
	case "log/slog.Attr":
		// An attribute handed over as a value is kept whole, since what the
		// words would have written of it is written as the pair it stands for.
		value = go2jsUnwrap(go2jsUntyped(value));

		return value instanceof go2jsSlogAttr
			? new go2jsSlogValue(go2jsSlogKind.Any, 0, "", value)
			: go2jsSlogAnyValueText(String(value));
	case "[]slog.Attr":
	case "[]log/slog.Attr":
		return go2jsSlogGroupValue(go2jsSlogAttrsOf(go2jsUnwrap(go2jsUntyped(value))));
	}

	// A value of a shape the runtime has of its own is a number, and a number
	// that counts without a remainder is the whole number Go would have written
	// it as, while one that carries a remainder is a length that is not whole.
	if (type === "") {
		if (typeof value === "string") {
			return go2jsSlogStringValue(value);
		}

		if (typeof value === "boolean") {
			return go2jsSlogBoolValue(value);
		}

		if (typeof value === "bigint") {
			return go2jsSlogInt64Value(value);
		}

		if (typeof value === "number") {
			return Number.isInteger(value) ? go2jsSlogInt64Value(value) : go2jsSlogFloat64Value(value);
		}
	}

	// A value the program gave a LogValue method to is asked what it stands for
	// before it is written, and the kind of it is the kind that asking gives it.
	if (go2jsSlogIsLogValuer(value)) {
		return new go2jsSlogValue(go2jsSlogKind.LogValuer, 0, "", go2jsUnwrap(go2jsUntyped(value)));
	}

	return new go2jsSlogValue(go2jsSlogKind.Any, 0, "", value);
}

// go2jsSlogIsLogValuer is whether a value answers for being one, since a value
// the program gave a LogValue method to is asked what it stands for before it is
// written, rather than being written as the value it holds.
function go2jsSlogIsLogValuer(value) {
	value = go2jsUnwrap(go2jsUntyped(value));

	return value !== null && value !== undefined && typeof value === "object" &&
		typeof value.LogValue === "function";
}

// go2jsSlogAnyValueText is a value the program holds that stands for a fault,
// which Go keeps as a value of its own so that writing it out says what is
// wrong rather than that something is.
function go2jsSlogAnyValueText(text) {
	return new go2jsSlogValue(go2jsSlogKind.Any, 0, "", go2jsSlogErrorValue(text));
}

// go2jsSlogBigInt is a whole number of any width, since a number of more digits
// than a double holds is still a whole number and must not be rounded to fit.
function go2jsSlogBigInt(value) {
	if (typeof value === "bigint") {
		return value;
	}

	return BigInt(Math.trunc(Number(value) || 0));
}

// go2jsSlogNumberOf is the number a value stands for, however the runtime was
// handed it, since a typed or boxed value is a number written in a wrapper.
function go2jsSlogNumberOf(value) {
	return Number(go2jsUnwrap(go2jsUntyped(value)));
}

// go2jsSlogTextOf is the text a value stands for, whether it was written as one
// already or is something that says what it is.
function go2jsSlogTextOf(value) {
	value = go2jsUnwrap(go2jsUntyped(value));

	if (value === null || value === undefined) {
		return value === null ? "<nil>" : "<nil>";
	}

	if (typeof value === "string") {
		return value;
	}

	if (typeof value === "object" && typeof value.String === "function") {
		return String(go2jsSlogInvoke(value.String, value, []));
	}

	return String(value);
}

// go2jsSlogAttr holds a key with the value it stands for, which is the pair a
// record carries beside its message.
function go2jsSlogAttr(key, value) {
	this.Key = key === undefined ? "" : go2jsRawText(key);
	this.Value = value instanceof go2jsSlogValue ? value : new go2jsSlogValue(go2jsSlogKind.Any, 0, "", value);
}

go2jsSlogAttr.prototype.String = function() {
	return this.Key + "=" + this.Value.String();
};

go2jsSlogAttr.prototype.Equal = function(other) {
	return go2jsSlogAttrEqual(this, other);
};

go2jsSlogAttr.prototype.Value = function() {
	return this.Value;
};

function go2jsSlogAttrEqual(one, other) {
	one = go2jsSlogUnwrapAttr(one);
	other = go2jsSlogUnwrapAttr(other);

	if (one === null || one === undefined || other === null || other === undefined) {
		return false;
	}

	return one.Key === other.Key && one.Value.Equal(other.Value);
}

// An attribute or a value of the program can arrive wrapped in whatever it was
// typed as, and what a comparison is asked about is what is inside.
function go2jsSlogUnwrapAttr(attr) {
	attr = go2jsUnwrap(go2jsUntyped(attr));

	return attr instanceof go2jsSlogAttr ? attr : null;
}

function go2jsSlogUnwrapValue(value) {
	value = go2jsUnwrap(go2jsUntyped(value));

	return value instanceof go2jsSlogValue ? value : null;
}

// go2jsSlogAttrEmpty is an attribute nobody said anything about, which is an
// empty key with a value that holds nothing, and which is written nowhere.
function go2jsSlogAttrEmpty(attr) {
	if (attr.Key !== "") {
		return false;
	}

	const any = attr.Value.any;

	return attr.Value.kind === go2jsSlogKind.Any &&
		(any === null || any === undefined || any === go2jsInterface(null));
}

// go2jsSlogEmptyGroup is a group with nothing in it, which Go leaves out of a
// record because a group nobody filled says nothing a reader would want.
function go2jsSlogEmptyGroup(value) {
	return value.kind === go2jsSlogKind.Group && value.Group().length === 0;
}

function go2jsSlogString(key, value) {
	return new go2jsSlogAttr(key, go2jsSlogStringValue(value));
}

function go2jsSlogInt(key, value) {
	return new go2jsSlogAttr(key, go2jsSlogInt64Value(value));
}

function go2jsSlogInt64(key, value) {
	return new go2jsSlogAttr(key, go2jsSlogInt64Value(value));
}

function go2jsSlogUint64(key, value) {
	return new go2jsSlogAttr(key, go2jsSlogUint64Value(value));
}

function go2jsSlogFloat64(key, value) {
	return new go2jsSlogAttr(key, go2jsSlogFloat64Value(value));
}

function go2jsSlogBool(key, value) {
	return new go2jsSlogAttr(key, go2jsSlogBoolValue(value));
}

function go2jsSlogTime(key, value) {
	return new go2jsSlogAttr(key, go2jsSlogTimeValue(value));
}

function go2jsSlogDuration(key, value) {
	return new go2jsSlogAttr(key, go2jsSlogDurationValue(value));
}

function go2jsSlogAny(key, value, type) {
	return new go2jsSlogAttr(key, go2jsSlogAnyValue(value, type));
}

// go2jsSlogLevelAttr is the level of a record as the attribute a program may
// rewrite, which carries the name the level is written under rather than the
// number behind it, since that is what a level writes as.
function go2jsSlogLevelAttr(key, level) {
	return new go2jsSlogAttr(key, go2jsSlogAnyValue(go2jsSlogLevelHolder(level), "log/slog.Level"));
}

// go2jsSlogLevelHolder is a level as a value of its own, which writes as the
// name it is known by however it is written: as text, as JSON, or as itself.
function go2jsSlogLevelHolder(level) {
	const name = go2jsSlogLevelName(level);

	return {
		level: Number(level) | 0,
		MarshalText: function() {
			return name;
		},
		MarshalJSON: function() {
			return "\"" + go2jsSlogEscapeJSON(name) + "\"";
		},
		String: function() {
			return name;
		},
	};
}

function go2jsSlogGroup(key, ...rest) {
	let types = null;
	let args = [];
	if (rest.length > 0) {
		if (Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}

	return new go2jsSlogAttr(key, go2jsSlogGroupValue(go2jsSlogArgsToAttrs(args, types)));
}

// go2jsSlogAttrsOf reads the attributes a slice of them holds, since a slice the
// program made is a list the runtime walks the same way it walks one of its own.
function go2jsSlogAttrsOf(value) {
	if (Array.isArray(value)) {
		return value.map(attr => go2jsSlogAsAttr(attr));
	}

	return [];
}

function go2jsSlogAsAttr(value) {
	value = go2jsUntyped(value);
	if (value instanceof go2jsSlogAttr) {
		return value;
	}

	return new go2jsSlogAttr("!BADKEY", value);
}

// go2jsSlogArgsToAttrs reads the trailing arguments of a log call as the pairs
// they stand for, where a name in front of a value makes the pair and anything
// else stands for a value whose name was left out.
function go2jsSlogArgsToAttrs(args, types) {
	const attrs = [];
	let index = 0;

	while (index < args.length) {
		const value = args[index];

		if (typeof value === "string") {
			if (index + 1 === args.length) {
				attrs.push(new go2jsSlogAttr("!BADKEY", go2jsSlogStringValue(value)));
				break;
			}

			attrs.push(new go2jsSlogAttr(value, go2jsSlogAnyValue(args[index + 1],
				types === null || types === undefined ? null : types[index + 1])));

			index += 2;
			continue;
		}

		const v2 = go2jsUntyped(value);
		if (v2 instanceof go2jsSlogAttr) {
			attrs.push(v2);
			index++;
			continue;
		}

		attrs.push(new go2jsSlogAttr("!BADKEY",
			go2jsSlogAnyValue(value, types === null || types === undefined ? null : types[index])));

		index++;
	}

	return attrs;
}

// go2jsSlogRecord is one record of what a program logged: when it happened, how
// loud it was, what was said and what was said beside it.
function go2jsSlogRecord(time, level, message, pc) {
	this.Time = time === undefined || time === null ? go2jsTimeZero() : time;
	this.Message = message === undefined ? "" : go2jsRawText(message);
	this.Level = Number(level) | 0;
	this.PC = pc === undefined || pc === null ? 0 : Number(pc);

	// The attributes of a record are kept as one list: Go keeps the first few
	// inside the record and the rest beside it, which is an arrangement a list
	// does not need.
	this.__go2js_attrs = [];
	this.__go2js_source = "";
}

go2jsSlogRecord.prototype.NumAttrs = function() {
	return this.__go2js_attrs.length;
};

go2jsSlogRecord.prototype.Attrs = function(visit) {
	// A visitor is a function of the program behind whatever it was handed
	// as, and it is called the way a callback is called: run to its end here,
	// waiting where it waits.
	const seen = go2jsUnwrap(go2jsUntyped(visit));

	for (const attr of this.__go2js_attrs.slice()) {
		if (go2jsCallNow(seen, null, [attr]) === false) {
			return;
		}
	}
};

go2jsSlogRecord.prototype.Clone = function() {
	const clone = new go2jsSlogRecord(this.Time, this.Level, this.Message, this.PC);

	clone.__go2js_attrs = this.__go2js_attrs.slice();
	clone.__go2js_source = this.__go2js_source;

	return clone;
};

go2jsSlogRecord.prototype.AddAttrs = function(attrs) {
	for (const attr of go2jsSlogAttrsOf(attrs === undefined ? [] : attrs)) {
		if (!go2jsSlogEmptyGroup(attr.Value)) {
			this.__go2js_attrs.push(attr);
		}
	}
};

go2jsSlogRecord.prototype.Add = function(args, types) {
	for (const attr of go2jsSlogArgsToAttrs(args === undefined ? [] : args, types)) {
		if (!go2jsSlogEmptyGroup(attr.Value)) {
			this.__go2js_attrs.push(attr);
		}
	}
};

function go2jsSlogNewRecord(time, level, message, pc) {
	return new go2jsSlogRecord(time, level, message, pc);
}

// go2jsSlogRecordAdd is a record told what the program said beside its message.
function go2jsSlogRecordAdd(record, ...rest) {
	let types = null;
	let args = [];
	if (rest.length > 0) {
		if (Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}
	record.Add(args, types);
}

// go2jsSlogRecordAddAttrs is a record told the pairs the program made itself.
function go2jsSlogRecordAddAttrs(record, attrs) {
	record.AddAttrs(attrs);
}

function go2jsSlogRecordCloneMethod(record) {
	return record.Clone();
}

function go2jsSlogRecordNumAttrsMethod(record) {
	return record.NumAttrs();
}

function go2jsSlogRecordAttrsMethod(record, visit) {
	return record.Attrs(visit);
}

// go2jsSlogLevelVar holds a level that can be changed while a program runs,
// which is how the lowest level a handler writes is moved without stopping.
function go2jsSlogLevelVar(level) {
	this.level = go2jsSlogNumberOf(level) | 0;
}

go2jsSlogLevelVar.prototype.Level = function() {
	return this.level | 0;
};

go2jsSlogLevelVar.prototype.Set = function(level) {
	this.level = go2jsSlogNumberOf(level) | 0;
};

go2jsSlogLevelVar.prototype.String = function() {
	return go2jsSlogLevelName(this.level);
};

go2jsSlogLevelVar.prototype.Swap = function(level) {
	const held = this.level | 0;

	this.level = Number(level) | 0;

	return held;
};

go2jsSlogLevelVar.prototype.CompareAndSwap = function(wanted, level) {
	if ((this.level | 0) !== (Number(wanted) | 0)) {
		return false;
	}

	this.level = Number(level) | 0;

	return true;
};

function go2jsSlogLevelVarLevel(level) {
	return level.Level();
}

function go2jsSlogLevelVarSet(level, wanted) {
	level.Set(wanted);
}

function go2jsSlogLevelVarString(level) {
	return level.String();
}

function go2jsSlogLevelVarSwap(level, wanted) {
	return level.Swap(wanted);
}

function go2jsSlogLevelVarCompareAndSwap(level, wanted, level2) {
	return level.CompareAndSwap(wanted, level2);
}

// A level is a named number, so a verb or a writer that asks what a value is
// asks the level the name it is written under, whichever of the two names of
// the package the value is holding was written.
go2jsRegisterMethod("slog.Level.String", go2jsSlogLevelStringMethod);
go2jsRegisterMethod("log/slog.Level.String", go2jsSlogLevelStringMethod);
go2jsRegisterMethod("slog.Level.MarshalText", go2jsSlogLevelStringMethod);
go2jsRegisterMethod("log/slog.Level.MarshalText", go2jsSlogLevelStringMethod);
go2jsRegisterMethod("slog.Level.MarshalJSON", go2jsSlogLevelMarshalJSONMethod);
go2jsRegisterMethod("log/slog.Level.MarshalJSON", go2jsSlogLevelMarshalJSONMethod);

// A kind is a named number too, and fmt writes the name of the kind rather
// than the number behind one.
go2jsRegisterMethod("slog.Kind.String", go2jsSlogKindName);
go2jsRegisterMethod("log/slog.Kind.String", go2jsSlogKindName);

function go2jsSlogLevelStringMethod(level) {
	return go2jsSlogLevelName(level);
}

function go2jsSlogLevelMarshalJSONMethod(level) {
	return go2jsSlogStringValue(go2jsSlogLevelName(level));
}

// go2jsSlogLevelUnmarshalTextMethod is a level read out of a name, which is
// written into the level it stands for where the program can see it.
function go2jsSlogLevelUnmarshalTextMethod(level, text) {
	const parsed = go2jsSlogLevelParse(go2jsBytesToString(text));

	if (parsed === null) {
		go2jsSlogStoreLevel(level, 0);
		return go2jsSlogErrorValue("slog: level string " + JSON.stringify(go2jsBytesToString(text)) + ": unknown name");
	}

	go2jsSlogStoreLevel(level, parsed);

	return null;
}

// go2jsSlogLevelUnmarshalJSONMethod is a level read out of the text of a JSON
// value, which is the text of a name.
function go2jsSlogLevelUnmarshalJSONMethod(level, text) {
	const parsed = go2jsSlogLevelParse(go2jsRawText(text));

	if (parsed === null) {
		go2jsSlogStoreLevel(level, 0);
		return go2jsSlogErrorValue("slog: level string " + JSON.stringify(go2jsRawText(text)) + ": unknown name");
	}

	go2jsSlogStoreLevel(level, parsed);

	return null;
}

function go2jsSlogStoreLevel(level, wanted) {
	if (level === null || typeof level !== "object") {
		return;
	}

	if (go2jsPointerAccessor(level, go2jsPointerSet)) {
		level[go2jsPointerSet](wanted);
		return;
	}

	level.level = wanted;
}

function go2jsSlogValueKindMethod(value) {
	return value.Kind();
}

function go2jsSlogValueAnyMethod(value) {
	return value.Any();
}

function go2jsSlogValueBoolMethod(value) {
	return value.Bool();
}

function go2jsSlogValueDurationMethod(value) {
	return value.Duration();
}

function go2jsSlogValueFloat64Method(value) {
	return value.Float64();
}

function go2jsSlogValueInt64Method(value) {
	return value.Int64();
}

function go2jsSlogValueUint64Method(value) {
	return value.Uint64();
}

function go2jsSlogValueTimeMethod(value) {
	return value.Time();
}

function go2jsSlogValueGroupMethod(value) {
	return value.Group();
}

function go2jsSlogValueGroupValueMethod(value) {
	return value.GroupValue();
}

function go2jsSlogValueLogValueMethod(value) {
	return value.LogValue();
}

function go2jsSlogValueTextMethod(value) {
	return value.String();
}

function go2jsSlogValueEqualMethod(value, other) {
	return value.Equal(other);
}

function go2jsSlogValueResolveMethod(value) {
	return value.Resolve();
}

function go2jsSlogAttrStringMethod(attr) {
	return attr.String();
}

function go2jsSlogAttrEqualMethod(attr, other) {
	return attr.Equal(other);
}

function go2jsSlogAttrValueMethod(attr) {
	return attr.Value;
}
`
}

func slogHandlerRuntimeSource() string {
	return `
// go2jsSlogNewTextHandler is a handler that writes a record as key=value pairs
// separated by spaces, which is what a program reads when it is not reading
// anything but the words themselves.
function go2jsSlogNewTextHandler(writer, opts) {
	return go2jsSlogHandler(false, writer, opts);
}

// go2jsSlogNewJSONHandler is a handler that writes a record as one JSON object,
// which is what a program writes when something else is the one reading.
function go2jsSlogNewJSONHandler(writer, opts) {
	return go2jsSlogHandler(true, writer, opts);
}

// go2jsSlogHandler is one of the two handlers the package builds itself: what it
// writes, where it writes it and how it was asked to write it, together with the
// attributes and groups a program built it up with before the record arrived.
function go2jsSlogHandler(json, writer, opts) {
	const handler = {
		__go2js_slog_builtin: true,
		json: json,
		writer: writer,
		opts: opts === undefined || opts === null ? {} : opts,
		preformatted: "",
		groupPrefix: "",
		groups: [],
		nOpenGroups: 0,
	};

	handler.Enabled = function(ctx, level) {
		void ctx;
		return go2jsSlogHandlerEnabled(this, level);
	};

	handler.Handle = function(ctx, record) {
		void ctx;
		return go2jsSlogHandlerHandle(this, record);
	};

	handler.WithAttrs = function(attrs) {
		return go2jsSlogHandlerWithAttrs(this, attrs);
	};

	handler.WithGroup = function(name) {
		return go2jsSlogHandlerWithGroup(this, name);
	};

	return handler;
}

// go2jsSlogHandlerEnabled is whether a record of this level is one the handler
// writes, which it is once it is as loud as the lowest level it was asked for.
function go2jsSlogHandlerEnabled(handler, level) {
	const asked = handler.opts === undefined || handler.opts === null ? null : handler.opts;

	if (asked === null || asked.Level === undefined || asked.Level === null) {
		return (Number(level) | 0) >= 0;
	}

	return (Number(level) | 0) >= go2jsSlogLevelOf(asked.Level);
}

// go2jsSlogHandlerClone is a copy of a handler that can be given attributes of
// its own, which is how a logger is built up without disturbing the one it came
// from.
function go2jsSlogHandlerClone(handler) {
	const clone = go2jsSlogHandler(handler.json, handler.writer, handler.opts);

	clone.preformatted = handler.preformatted;
	clone.groupPrefix = handler.groupPrefix;
	clone.groups = handler.groups.slice();
	clone.nOpenGroups = handler.nOpenGroups;

	return clone;
}

// go2jsSlogHandlerWithAttrs is a handler that writes the attributes it already
// wrote before the ones it is given now, written once and for all, since they
// are the same on every record that follows.
function go2jsSlogHandlerWithAttrs(handler, attrs) {
	const held = go2jsSlogAttrsOf(attrs);

	// Attributes that are all groups nobody filled write nothing, so a list of
	// nothing but them leaves the handler standing as it does.
	if (go2jsSlogCountEmptyGroups(held) === held.length) {
		return handler;
	}

	const next = go2jsSlogHandlerClone(handler);
	const state = go2jsSlogState(next, held !== null ? next.preformatted : "");
	state.prefix.push(next.groupPrefix);
	go2jsSlogOpenGroups(state);

	const mark = state.buf.length;

	if (go2jsSlogAppendAttrs(state, held)) {
		next.preformatted = state.buf.join("");
		next.groupPrefix = state.prefix.join("");
		next.nOpenGroups = next.groups.length;
	} else {
		next.preformatted = state.buf.slice(0, mark).join("");
	}

	return next;
}

// go2jsSlogHandlerWithGroup is a handler that writes everything it is given
// under the name of one more group.
function go2jsSlogHandlerWithGroup(handler, name) {
	const next = go2jsSlogHandlerClone(handler);

	next.groups.push(go2jsRawText(name));

	return next;
}

// go2jsSlogHandlerHandle writes one record, which is one line however many
// attributes the record carries, since a line that ran on would be read as a
// record of its own.
function go2jsSlogHandlerHandle(handler, record) {
	const state = go2jsSlogState(handler, "");

	if (handler.json) {
		go2jsSlogWrite(state, "{");
	}

	const replace = go2jsSlogReplaceOf(handler);
	const groups = state.groups;

	// The attributes a handler writes about a record are not inside any group,
	// so a program asked to rewrite one of them is told of no group at all.
	state.groups = null;

	if (!go2jsSlogTimeIsZero(record.Time)) {
		if (replace === null) {
			go2jsSlogAppendKey(state, "time");
			go2jsSlogAppendTime(state, record.Time);
		} else {
			go2jsSlogAppendAttr(state, go2jsSlogTime("time", record.Time));
		}
	}

	if (replace === null) {
		go2jsSlogAppendKey(state, "level");
		go2jsSlogAppendString(state, go2jsSlogLevelName(record.Level));
	} else {
		go2jsSlogAppendAttr(state, go2jsSlogLevelAttr("level", record.Level));
	}

	if (handler.opts !== undefined && handler.opts !== null && handler.opts.AddSource === true) {
		go2jsSlogAppendAttr(state, go2jsSlogAny("source", go2jsSlogSourceOf(record.__go2js_source), ""));
	}

	if (replace === null) {
		go2jsSlogAppendKey(state, "msg");
		go2jsSlogAppendString(state, record.Message);
	} else {
		go2jsSlogAppendAttr(state, go2jsSlogString("msg", record.Message));
	}

	state.groups = groups;
	go2jsSlogAppendNonBuiltIns(state, record);
	go2jsSlogWrite(state, "\n");

	return go2jsSlogWriteLine(handler, state.buf.join(""));
}

// go2jsSlogReplaceOf is the function a program gave for rewriting attributes,
// which is none at all unless it gave one.
function go2jsSlogReplaceOf(handler) {
	const opts = handler.opts === undefined || handler.opts === null ? null : handler.opts;

	if (opts === null || typeof opts.ReplaceAttr !== "function") {
		return null;
	}

	return opts.ReplaceAttr;
}

// go2jsSlogState is the writing of one record as it stands, which is gathered as
// pieces so that a group found to hold nothing can be taken back off the end.
function go2jsSlogState(handler, written) {
	const opts = handler.opts === undefined || handler.opts === null ? null : handler.opts;
	const state = {
		h: handler,
		buf: written === "" ? [] : [written],
		sep: "",
		prefix: [],
		groups: null,
	};

	if (opts !== null && typeof opts.ReplaceAttr === "function") {
		state.groups = handler.groups.slice(0, handler.nOpenGroups);
	}

	if (written !== "") {
		state.sep = handler.json ? "," : " ";

		if (handler.json && written.endsWith("{")) {
			state.sep = "";
		}
	}

	return state;
}

function go2jsSlogWrite(state, text) {
	state.buf.push(text);
}

function go2jsSlogOpenGroups(state) {
	for (const name of state.h.groups.slice(state.h.nOpenGroups)) {
		go2jsSlogOpenGroup(state, name);
	}
}

function go2jsSlogOpenGroup(state, name) {
	if (state.h.json) {
		go2jsSlogAppendKey(state, name);
		go2jsSlogWrite(state, "{");
		state.sep = "";
	} else {
		state.prefix.push(name, ".");
	}

	if (state.groups !== null) {
		state.groups.push(name);
	}
}

function go2jsSlogCloseGroup(state, name) {
	if (state.h.json) {
		go2jsSlogWrite(state, "}");
	} else {
		state.prefix.splice(state.prefix.length - 2, 2);
	}

	state.sep = state.h.json ? "," : " ";

	if (state.groups !== null) {
		state.groups.pop();
	}

	void name;
}

// go2jsSlogAppendAttrs writes a list of attributes and says whether anything of
// them was written, since a list of nothing but groups nobody filled writes
// nothing at all.
function go2jsSlogAppendAttrs(state, attrs) {
	let written = false;

	for (const attr of attrs) {
		if (go2jsSlogAppendAttr(state, attr)) {
			written = true;
		}
	}

	return written;
}

// go2jsSlogAppendAttr writes one attribute, after asking the program whether it
// wants it written another way, and says whether anything was written for it.
function go2jsSlogAppendAttr(state, attr) {
	let held = new go2jsSlogAttr(attr.Key, attr.Value.Resolve());
	const replace = go2jsSlogReplaceOf(state.h);

	if (replace !== null && held.Value.kind !== go2jsSlogKind.Group) {
		held = go2jsSlogAsAttr(go2jsSlogInvoke(replace, null,
			[state.groups === null ? null : state.groups.slice(), held]));
		held = new go2jsSlogAttr(held.Key, held.Value.Resolve());
	}

	if (go2jsSlogAttrEmpty(held)) {
		return false;
	}

	// The place a record was made is a value of its own, written as the place it
	// was made rather than as the fields it is made of, unless the record is
	// being written as JSON, where those fields are the shape of the writing.
	if (held.Value.kind === go2jsSlogKind.Any) {
		const source = go2jsSlogSourceParts(held.Value.any);

		if (source !== null) {
			held = new go2jsSlogAttr(held.Key, state.h.json
				? go2jsSlogGroupValue([
					go2jsSlogString("function", source.Function),
					go2jsSlogString("file", source.File),
					go2jsSlogInt("line", source.Line),
				])
				: go2jsSlogStringValue(source.File + ":" + source.Line));
		}
	}

	if (held.Value.kind === go2jsSlogKind.Group) {
		const attrs = held.Value.Group();

		// A group with nothing in it is not written, and a group that turns out
		// to hold nothing once the program has had its say about it is taken back
		// off the end of what was written.
		if (attrs.length > 0) {
			const mark = state.buf.length;

			if (held.Key !== "") {
				go2jsSlogOpenGroup(state, held.Key);
			}

			if (!go2jsSlogAppendAttrs(state, attrs)) {
				state.buf.length = mark;
				return false;
			}

			if (held.Key !== "") {
				go2jsSlogCloseGroup(state, held.Key);
			}
		}
	} else {
		go2jsSlogAppendKey(state, held.Key);
		go2jsSlogAppendValue(state, held.Value);
	}

	return true;
}

// go2jsSlogAppendNonBuiltIns writes the attributes a program gave before the
// record was made, which were written once and for all, and then those the record
// carries itself, which are inside the groups the handler was given.
function go2jsSlogAppendNonBuiltIns(state, record) {
	const handler = state.h;

	if (handler.preformatted !== "") {
		go2jsSlogWrite(state, state.sep);
		go2jsSlogWrite(state, handler.preformatted);
		state.sep = handler.json ? "," : " ";

		if (handler.json && handler.preformatted.endsWith("{")) {
			state.sep = "";
		}
	}

	// A record with no attributes writes no group at all, since a group nobody
	// put anything in is not part of what the record said.
	let open = handler.nOpenGroups;

	if (record.NumAttrs() > 0) {
		state.prefix.push(handler.groupPrefix);

		const mark = state.buf.length;

		go2jsSlogOpenGroups(state);
		open = handler.groups.length;

		let empty = true;

		for (const attr of record.__go2js_attrs) {
			if (go2jsSlogAppendAttr(state, attr)) {
				empty = false;
			}
		}

		if (empty) {
			state.buf.length = mark;
			open = handler.nOpenGroups;
		}
	}

	if (handler.json) {
		for (let count = 0; count < open; count++) {
			go2jsSlogWrite(state, "}");
		}

		go2jsSlogWrite(state, "}");
	}
}

function go2jsSlogAppendKey(state, key) {
	go2jsSlogWrite(state, state.sep);
	go2jsSlogAppendString(state, state.prefix.join("") + key);

	if (state.h.json) {
		go2jsSlogWrite(state, ":");
	} else {
		go2jsSlogWrite(state, "=");
	}

	state.sep = state.h.json ? "," : " ";
}

function go2jsSlogAppendString(state, text) {
	go2jsSlogWrite(state, state.h.json
		? "\"" + go2jsSlogEscapeJSON(text) + "\""
		: go2jsSlogQuoted(text));
}

// go2jsSlogAppendValue writes the value of an attribute, which is written the
// way the handler was built to write values.
function go2jsSlogAppendValue(state, value) {
	try {
		if (state.h.json) {
			go2jsSlogAppendJSONValue(state, value);
		} else {
			go2jsSlogAppendTextValue(state, value);
		}
	} catch (err) {
		go2jsSlogAppendString(state, "!ERROR:" + go2jsSlogErrorText(err));
	}
}

// go2jsSlogAppendTextValue writes a value the way Go writes one into a line of
// key=value pairs: a string and a moment are written as themselves, a value of
// no kind of its own is written the way the words would write it, and a value
// that has a kind is written as that kind writes it.
// go2jsSlogIsByteSlice is whether a value is a run of bytes, which a handler
// writes as the words it stands for rather than as the numbers behind them.
function go2jsSlogIsByteSlice(value) {
	const name = go2jsGoTypeNameRaw(value);

	return (name === "[]byte" || name === "[]uint8" || name === "[]uint8 " ||
		name === "[]byte ") && go2jsSlogBytesOf(go2jsUnwrap(go2jsUntyped(value))) !== null;
}

// go2jsSlogBytesOf is the numbers a run of bytes holds, whether the runtime
// keeps them as a list or as the text they were written from.
function go2jsSlogBytesOf(value) {
	if (value === null || value === undefined) {
		return null;
	}

	if (Array.isArray(value)) {
		return value;
	}

	if (typeof value === "string") {
		return go2jsStringToBytes(value);
	}

	if (typeof value === "object") {
		if (Array.isArray(value.data)) {
			return value.data.slice(value.offset, value.offset + value.length);
		}

		if (typeof value.valueOf === "function") {
			const held = value.valueOf();

			if (Array.isArray(held)) {
				return held;
			}
		}
	}

	return null;
}

function go2jsSlogAppendTextValue(state, value) {
	switch (value.kind) {
	case go2jsSlogKind.String:
		go2jsSlogAppendString(state, value.str);
		return;
	case go2jsSlogKind.Time:
		go2jsSlogAppendTime(state, value.any);
		return;
	case go2jsSlogKind.Any: {
		const any = go2jsUnwrap(go2jsUntyped(value.any));

		// An attribute held as a value is written as the pair it stands for,
		// since that is what it says of itself when it is shown.
		if (any instanceof go2jsSlogAttr) {
			go2jsSlogAppendString(state, go2jsSlogInvoke(any.String, any, []));
			return;
		}

		// A run of bytes is written as the words it stands for, since that is
		// what a run of bytes is for, and writing out the numbers behind them
		// says nothing that a reader wanted.
		if (go2jsSlogIsByteSlice(value.any)) {
			go2jsSlogWrite(state, go2jsSprintf("%q", go2jsBytesToString(go2jsSlogBytesOf(any))));
			return;
		}

		const written = any !== null && any !== undefined &&
			typeof any === "object" && typeof any.MarshalText === "function"
			? go2jsSlogInvoke(any.MarshalText, any, [])
			: null;

		if (written !== null) {
			go2jsSlogAppendString(state, go2jsRawText(written));
			return;
		}

		go2jsSlogAppendString(state, go2jsSprintf("%+v", any));
		return;
	}
	}

	go2jsSlogWrite(state, go2jsSlogValueText(value));
}

function go2jsSlogAppendJSONValue(state, value) {
	switch (value.kind) {
	case go2jsSlogKind.String:
		go2jsSlogAppendString(state, value.str);
		return;
	case go2jsSlogKind.Int64:
	case go2jsSlogKind.Uint64:
		go2jsSlogWrite(state, String(value.num));
		return;
	case go2jsSlogKind.Duration:
		go2jsSlogWrite(state, String(value.num));
		return;
	case go2jsSlogKind.Float64:
		go2jsSlogWrite(state, go2jsSlogJSONFloat(Number(value.num)));
		return;
	case go2jsSlogKind.Bool:
		go2jsSlogWrite(state, value.num === true || value.num === 1 ? "true" : "false");
		return;
	case go2jsSlogKind.Time:
		go2jsSlogWrite(state, "\"" + go2jsSlogTimeText(value.any, "2006-01-02T15:04:05.999999999Z07:00") + "\"");
		return;
	case go2jsSlogKind.Any: {
		const any = value.any;

		// An error an attribute holds is written as the text it says, since that
		// is what the words of a fault are, unless it says how to write itself as
		// something else first.
		if (go2jsIsErrorValue(any) && typeof any.MarshalJSON !== "function") {
			go2jsSlogAppendString(state, go2jsRawText(go2jsSlogErrorText(any)));
			return;
		}

		if (any !== null && any !== undefined && typeof any === "object" &&
			typeof any.MarshalJSON === "function") {
			go2jsSlogWrite(state, go2jsRawText(go2jsSlogInvoke(any.MarshalJSON, any, [])));
			return;
		}

		go2jsSlogWrite(state, go2jsSlogJSONText(any));
		return;
	}
	}

	go2jsSlogAppendString(state, "!ERROR:bad kind: " + go2jsSlogKindName(value.kind));
}

// go2jsSlogAppendTime writes a moment the way RFC 3339 writes it with the
// thousandths of a second counted out in full.
function go2jsSlogAppendTime(state, time) {
	go2jsSlogWrite(state, go2jsSlogHandlerOf(state).json
		? "\"" + go2jsSlogTimeText(time, "2006-01-02T15:04:05.999999999Z07:00") + "\""
		: go2jsSlogTimeText(time, "2006-01-02T15:04:05.000Z07:00"));
}

function go2jsSlogHandlerOf(state) {
	return state.h;
}

// go2jsSlogJSONFloat is a length written the way encoding/json writes one,
// which is in full until the number is so large or so small that an exponent is
// shorter than the digits it stands for.
function go2jsSlogJSONFloat(value) {
	if (!isFinite(value)) {
		return go2jsSlogQuoted("!ERROR:json: unsupported value: " +
			(Number.isNaN(value) ? "NaN" : (value > 0 ? "+Inf" : "-Inf")));
	}

	const held = Math.abs(value);

	if (held === 0) {
		return Object.is(value, -0) ? "-0" : "0";
	}

	if (held < 1e-6 || held >= 1e21) {
		const written = value.toExponential().split("e");
		const sign = written[1][0] === "-" ? "-" : "+";
		const step = written[1].slice(1).replace(/^0+/, "");

		return written[0] + "e" + sign + step;
	}

	return String(value);
}

// go2jsSlogJSONText is a value of no kind of its own written as JSON writes it,
// which is what a handler told to write JSON writes for a value it knows nothing
// about.
function go2jsSlogJSONText(value) {
	if (value === null || value === undefined) {
		return "null";
	}

	switch (typeof value) {
	case "string":
		return "\"" + go2jsSlogEscapeJSON(value) + "\"";
	case "boolean":
		return value ? "true" : "false";
	case "bigint":
		return String(value);
	case "number":
		return go2jsSlogJSONFloat(value);
	case "function":
		return go2jsSlogQuoted("!ERROR:json: unsupported type: function");
	case "symbol":
		return go2jsSlogQuoted("!ERROR:json: unsupported type: symbol");
	}

	if (value.value instanceof Date) {
		return "\"" + go2jsSlogTimeText(value) + "\"";
	}

	if (Array.isArray(value)) {
		return "[" + value.map(item => go2jsSlogJSONText(item)).join(",") + "]";
	}

	if (value instanceof go2jsNativeMap) {
		const parts = [];

		for (const [key, item] of value.entries()) {
			parts.push("\"" + go2jsSlogEscapeJSON(go2jsStringify(key)) + "\":" + go2jsSlogJSONText(item));
		}

		return "{" + parts.join(",") + "}";
	}

	if (value instanceof go2jsSlogAttr) {
		return "\"" + go2jsSlogEscapeJSON(value.Key) + "\":" + go2jsSlogJSONText(value.Value.Any());
	}

	const parts = [];

	for (const key of Object.keys(value)) {
		if (key.startsWith("__go2js")) {
			continue;
		}

		parts.push("\"" + go2jsSlogEscapeJSON(key) + "\":" + go2jsSlogJSONText(value[key]));
	}

	return "{" + parts.join(",") + "}";
}

// go2jsSlogQuoted is a piece of a line written as it is when nothing in it needs
// quoting, and quoted the way strconv quotes when something does.
function go2jsSlogQuoted(text) {
	return go2jsSlogNeedsQuoting(text) ? go2jsSprintf("%q", text) : text;
}

// go2jsSlogNeedsQuoting is whether a piece of a line carries something that would
// be read as more or less than the piece itself, which is what a space, a mark of
// equality, a mark of quotation or anything no page shows would.
function go2jsSlogNeedsQuoting(text) {
	if (text === "") {
		return true;
	}

	return go2jsSlogUnprintable.test(text);
}

const go2jsSlogUnprintable = /[\x00-\x1f\x7f"'\\= \p{Cc}\p{Cf}\p{Cs}\p{Co}\p{Cn}\p{Zl}\p{Zp}]/u;

// go2jsSlogEscapeJSON is a string written as JSON writes one, where the marks
// that end a string and the one that escapes the character after it are the only
// ones written in full, and the two characters that end a line inside a string
// are written out even though JSON does not ask for it, since JSON is often
// read as JavaScript.
function go2jsSlogEscapeJSON(text) {
	const hex = "0123456789abcdef";
	let out = "";
	let start = 0;

	for (let index = 0; index < text.length; index++) {
		const held = text[index];
		const code = text.charCodeAt(index);

		if (code < 0x80) {
			// A character a page shows is written as it is, save for the two that
			// end a string and the one that stands for the character after it.
			if (code >= 0x20 && code <= 0x7e && held !== "\"" && held !== "\\") {
				continue;
			}

			out += text.slice(start, index);

			switch (held) {
			case "\\":
			case "\"":
				out += "\\" + held;
				break;
			case "\n":
				out += "\\n";
				break;
			case "\r":
				out += "\\r";
				break;
			case "\t":
				out += "\\t";
				break;
			default:
				out += "\\u00" + hex[code >> 4] + hex[code & 0xf];
			}

			start = index + 1;
			continue;
		}

		if (code >= 0xd800 && code <= 0xdbff && index + 1 < text.length) {
			const next = text.charCodeAt(index + 1);

			if (next >= 0xdc00 && next <= 0xdfff) {
				continue;
			}
		}

		if (code === 0x2028 || code === 0x2029) {
			out += text.slice(start, index) + "\\u202" + hex[code & 0xf];
			start = index + 1;
		}
	}

	return start < text.length ? out + text.slice(start) : out;
}

// go2jsSlogWriteLine is a line handed to the writer a program gave, which is one
// call to Write however many attributes the record carried.
function go2jsSlogWriteLine(handler, text) {
	// The line of a record written through no handler of the program's is left
	// to the logger of the log package, since that is what writes it there.
	if (handler.__go2js_slog_default === true) {
		go2jsLogOutput(go2jsLogDecorate(text));

		return null;
	}

	const target = go2jsUnwrap(handler.writer);

	if (target === null || target === undefined) {
		return go2jsSlogErrorValue("slog: writer is nil");
	}

	if (typeof target.Write !== "function") {
		return go2jsSlogErrorValue("slog: writer does not implement io.Writer");
	}

	const written = go2jsCallNow(target.Write, target, [go2jsStringToBytes(text)]);

	if (Array.isArray(written)) {
		return written[1] === null || written[1] === undefined ? null : written[1];
	}

	return written === undefined ? null : written;
}

// go2jsSlogTimeText is a moment written as a layout writes it, where a layout
// named for a moment is how Go writes one.
function go2jsSlogTimeText(time, layout) {
	const wanted = layout === undefined || layout === null ? "2006-01-02 15:04:05.999999999 -0700 MST" : layout;

	if (time === null || time === undefined) {
		return "";
	}

	if (time.value instanceof Date) {
		return go2jsTimeFormat(time.value, wanted, time);
	}

	if (typeof time.Format === "function") {
		return go2jsRawText(time.Format(wanted));
	}

	return go2jsRawText(time);
}

function go2jsSlogTimeIsZero(time) {
	if (time === null || time === undefined) {
		return true;
	}

	if (typeof time.IsZero === "function") {
		return time.IsZero() === true;
	}

	return time.value instanceof Date && time.__go2js_time_zero === true;
}

// go2jsSlogSourceOf is the place a program stood when it asked for a record to
// be written, read back into the parts Go keeps a place in.
function go2jsSlogSourceOf(text) {
	const held = String(text === null || text === undefined ? "" : text);

	if (held === "") {
		return null;
	}

	const at = held.lastIndexOf(":");

	if (at < 0) {
		return {Function: "", File: held, Line: 0};
	}

	return {Function: "", File: held.slice(0, at), Line: Number(held.slice(at + 1))};
}

// go2jsSlogSourceParts reads a place out of a value, whether the program made one
// or the runtime read one back out of the text it was given.
function go2jsSlogSourceParts(value) {
	if (value === null || value === undefined || typeof value !== "object") {
		return null;
	}

	if (typeof value.File === "string" && typeof value.Line === "number") {
		return {Function: String(value.Function === undefined ? "" : value.Function), File: value.File, Line: value.Line};
	}

	if (typeof value.__go2js_file === "string" && typeof value.__go2js_line === "number") {
		return {Function: "", File: value.__go2js_file, Line: value.__go2js_line};
	}

	return null;
}

function go2jsSlogCountEmptyGroups(attrs) {
	let count = 0;

	for (const attr of attrs) {
		if (go2jsSlogEmptyGroup(attr.Value)) {
			count++;
		}
	}

	return count;
}

// go2jsSlogInvoke asks a value of the program to do something and reads what it
// answered, since a method the program declared may be one that waits and a
// method of two answers comes back as a pair of them.
function go2jsSlogInvoke(target, self, args) {
	if (target === null || target === undefined || typeof target !== "function") {
		return null;
	}

	const answered = go2jsCallNow(target, self, args === undefined ? [] : args);

	return Array.isArray(answered) ? answered[0] : answered;
}

// go2jsSlogError is a fault the package stands for on its own account, which is
// what a record says went wrong when the program said so with words.
function go2jsSlogErrorValue(text) {
	return go2jsErrorsNew(String(text));
}

function go2jsSlogErrorText(err) {
	return go2jsErrorMessage(err);
}
`
}

func slogLoggerRuntimeSource() string {
	return `
// go2jsSlogLogger writes records to the handler it was given, which is the one
// place a logger decides what a record looks like rather than what it says.
function go2jsSlogLogger(handler) {
	return {
		handler: handler === undefined || handler === null ? go2jsSlogDefaultHandler() : handler,
		Handler: function() {
			return this.handler;
		},
		Enabled: function(ctx, level) {
			return go2jsSlogLoggerEnabled(this, ctx, level);
		},
		With: function(args, types) {
			return go2jsSlogLoggerWith(this, args, types);
		},
		WithGroup: function(name) {
			return go2jsSlogLoggerWithGroup(this, name);
		},
		Log: function(ctx, level, message, args, types, source) {
			return go2jsSlogWriteRecord(this, ctx, level, message,
				go2jsSlogArgsToAttrs(args === undefined ? [] : args, types), source);
		},
		LogAttrs: function(ctx, level, message, attrs, source) {
			return go2jsSlogWriteRecord(this, ctx, level, message, attrs, source);
		},
		Debug: function(message, args, types, source) {
			return this.Log(null, -4, message, args, types, source);
		},
		Info: function(message, args, types, source) {
			return this.Log(null, 0, message, args, types, source);
		},
		Warn: function(message, args, types, source) {
			return this.Log(null, 4, message, args, types, source);
		},
		Error: function(message, args, types, source) {
			return this.Log(null, 8, message, args, types, source);
		},
	};
}

function go2jsSlogNew(handler) {
	return new go2jsSlogLogger(handler);
}

// go2jsSlogLoggerEnabled is whether the handler behind a logger writes a record
// of this level, which is the handler's own answer rather than the logger's.
function go2jsSlogLoggerEnabled(logger, ctx, level) {
	const handler = go2jsUnwrap(logger.handler);

	if (handler === null || handler === undefined || typeof handler.Enabled !== "function") {
		return true;
	}

	return go2jsSlogInvoke(handler.Enabled, handler, [ctx, Number(level) | 0]) !== false;
}

// go2jsSlogLoggerWith is a logger that writes what this one writes and then
// writes these attributes as well, which the handler remembers rather than the
// logger, since they are the same on every record that follows.
function go2jsSlogLoggerWith(logger, ...rest) {
	let types = null;
	let args = [];
	if (rest.length > 0) {
		if (Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}
	const handler = go2jsUnwrap(logger.handler);
	const attrs = go2jsSlogArgsToAttrs(args, types);

	if (handler === null || handler === undefined || typeof handler.WithAttrs !== "function") {
		return new go2jsSlogLogger(handler);
	}

	return new go2jsSlogLogger(go2jsSlogInvoke(handler.WithAttrs, handler, [attrs]));
}

// go2jsSlogLoggerWithGroup is a logger that writes everything it is given under
// one more group name.
function go2jsSlogLoggerWithGroup(logger, name) {
	const handler = go2jsUnwrap(logger.handler);

	if (handler === null || handler === undefined || typeof handler.WithGroup !== "function") {
		return new go2jsSlogLogger(handler);
	}

	return new go2jsSlogLogger(go2jsSlogInvoke(handler.WithGroup, handler, [go2jsRawText(name)]));
}

// go2jsSlogWriteRecord writes one record, which is first asked of the handler
// whether it is one the handler writes at all, since a record of a level the
// handler drops is not written and never reaches the writer.
function go2jsSlogWriteRecord(logger, ctx, level, message, attrs, source) {
	const handler = go2jsUnwrap(logger.handler);

	if (handler === null || handler === undefined || typeof handler.Handle !== "function") {
		return;
	}

	// A level is a named number, so it is read through whatever it was wrapped
	// in on the way here.
	level = go2jsSlogLevelNumber(level);

	if (!go2jsSlogLoggerEnabled(logger, ctx, level)) {
		return;
	}

	const record = new go2jsSlogRecord(go2jsTimeNow(), level, message, 0);

	record.__go2js_source = source === undefined || source === null ? "" : String(source);
	record.AddAttrs(attrs === undefined || attrs === null ? [] : attrs);

	go2jsSlogInvoke(handler.Handle, handler, [ctx, record]);
}

function go2jsSlogLoggerHandlerMethod(logger) {
	return logger.handler;
}

function go2jsSlogLoggerEnabledMethod(logger, ctx, level) {
	return go2jsSlogLoggerEnabled(logger, ctx, level);
}

function go2jsSlogLoggerWithMethod(logger, args, types) {
	return go2jsSlogLoggerWith(logger, args, types);
}

function go2jsSlogLoggerWithGroupMethod(logger, name) {
	return go2jsSlogLoggerWithGroup(logger, name);
}

function go2jsSlogLoggerLog(logger, ctx, level, message, args, types, source) {
	return go2jsSlogWriteRecord(logger, ctx, level, message,
		go2jsSlogArgsToAttrs(args === undefined ? [] : args, types), source);
}

function go2jsSlogLoggerLogAttrs(logger, ctx, level, message, attrs, source) {
	const src = (typeof source === "string" && source.indexOf(":") >= 0) ? source : undefined;
	const attrsArr = Array.isArray(attrs) ? attrs : (attrs === undefined || attrs === null ? [] : [attrs]);
	return go2jsSlogWriteRecord(logger, ctx, level, message, attrsArr, src);
}

function go2jsSlogLoggerDebug(logger, message, ...rest) {
	let types = null;
	let source = undefined;
	let args = [];
	if (rest.length > 0) {
		const last = rest[rest.length - 1];
		if (typeof last === "string" && last.indexOf(":") >= 0) {
			source = last;
			rest = rest.slice(0, rest.length - 1);
		}
		if (rest.length > 0 && Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}
	return go2jsSlogWriteRecord(logger, null, -4, message, go2jsSlogArgsToAttrs(args, types), source);
}

function go2jsSlogLoggerInfo(logger, message, ...rest) {
	let types = null;
	let source = undefined;
	let args = [];
	if (rest.length > 0) {
		const last = rest[rest.length - 1];
		if (typeof last === "string" && last.indexOf(":") >= 0 && rest.length >= 2) {
			source = last;
			rest = rest.slice(0, rest.length - 1);
		}
		if (rest.length > 0 && Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}
	return go2jsSlogWriteRecord(logger, null, 0, message, go2jsSlogArgsToAttrs(args, types), source);
}

function go2jsSlogLoggerWarn(logger, message, ...rest) {
	let types = null;
	let source = undefined;
	let args = [];
	if (rest.length > 0) {
		const last = rest[rest.length - 1];
		if (typeof last === "string" && last.indexOf(":") >= 0) {
			source = last;
			rest = rest.slice(0, rest.length - 1);
		}
		if (rest.length > 0 && Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}
	return go2jsSlogWriteRecord(logger, null, 4, message, go2jsSlogArgsToAttrs(args, types), source);
}

function go2jsSlogLoggerError(logger, message, ...rest) {
	let types = null;
	let source = undefined;
	let args = [];
	if (rest.length > 0) {
		const last = rest[rest.length - 1];
		if (typeof last === "string" && last.indexOf(":") >= 0) {
			source = last;
			rest = rest.slice(0, rest.length - 1);
		}
		if (rest.length > 0 && Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}
	return go2jsSlogWriteRecord(logger, null, 8, message, go2jsSlogArgsToAttrs(args, types), source);
}

function go2jsSlogLoggerDebugContext(logger, ctx, message, ...rest) {
	let types = null;
	let source = undefined;
	let args = [];
	if (rest.length > 0) {
		const last = rest[rest.length - 1];
		if (typeof last === "string" && last.indexOf(":") >= 0) {
			source = last;
			rest = rest.slice(0, rest.length - 1);
		}
		if (rest.length > 0 && Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}
	return go2jsSlogWriteRecord(logger, ctx, -4, message, go2jsSlogArgsToAttrs(args, types), source);
}

function go2jsSlogLoggerInfoContext(logger, ctx, message, ...rest) {
	let types = null;
	let source = undefined;
	let args = [];
	if (rest.length > 0) {
		const last = rest[rest.length - 1];
		if (typeof last === "string" && last.indexOf(":") >= 0) {
			source = last;
			rest = rest.slice(0, rest.length - 1);
		}
		if (rest.length > 0 && Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}
	return go2jsSlogWriteRecord(logger, ctx, 0, message, go2jsSlogArgsToAttrs(args, types), source);
}

function go2jsSlogLoggerWarnContext(logger, ctx, message, ...rest) {
	let types = null;
	let source = undefined;
	let args = [];
	if (rest.length > 0) {
		const last = rest[rest.length - 1];
		if (typeof last === "string" && last.indexOf(":") >= 0) {
			source = last;
			rest = rest.slice(0, rest.length - 1);
		}
		if (rest.length > 0 && Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}
	return go2jsSlogWriteRecord(logger, ctx, 4, message, go2jsSlogArgsToAttrs(args, types), source);
}

function go2jsSlogLoggerErrorContext(logger, ctx, message, ...rest) {
	let types = null;
	let source = undefined;
	let args = [];
	if (rest.length > 0) {
		const last = rest[rest.length - 1];
		if (typeof last === "string" && last.indexOf(":") >= 0) {
			source = last;
			rest = rest.slice(0, rest.length - 1);
		}
		if (rest.length > 0 && Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}
	return go2jsSlogWriteRecord(logger, ctx, 8, message, go2jsSlogArgsToAttrs(args, types), source);
}

// go2jsSlogDefaultLogger is the logger a program logs through when it names no
// logger of its own, which writes through the logger of the log package and so
// is written where that one is written.
const go2jsSlogDefaultLogger = new go2jsSlogLogger(go2jsSlogDefaultHandler());

// go2jsSlogLogLoggerLevel is the lowest level the default logger writes, which
// the log package decides and a program may move.
let go2jsSlogLogLoggerLevel = 0;

// go2jsSlogDefaultHandler is the handler behind the default logger, which is not
// one of the two the package builds, since it leaves the writing of the line to
// the log package and only gathers what goes on it.
function go2jsSlogDefaultHandler() {
	const handler = {
		__go2js_slog_default: true,
		opts: {},
		preformatted: "",
		groupPrefix: "",
		groups: [],
		nOpenGroups: 0,
	};

	handler.Enabled = function(ctx, level) {
		void ctx;
		return (Number(level) | 0) >= go2jsSlogLogLoggerLevel;
	};

	handler.Handle = function(ctx, record) {
		void ctx;

		const state = go2jsSlogState(this, go2jsSlogLevelName(record.Level) + " " + record.Message);

		state.sep = " ";
		go2jsSlogAppendNonBuiltIns(state, record);

		return go2jsSlogWriteLine(this, state.buf.join("") + "\n");
	};

	handler.WithAttrs = function(attrs) {
		return go2jsSlogHandlerWithAttrs(this, attrs);
	};

	handler.WithGroup = function(name) {
		return go2jsSlogHandlerWithGroup(this, name);
	};

	return handler;
}

function go2jsSlogDefault() {
	return go2jsSlogDefaultLogger;
}

// go2jsSlogSetDefault is the logger a program logs through when it names none,
// which is the one it was given.
function go2jsSlogSetDefault(logger) {
	if (logger === null || logger === undefined) {
		return;
	}

	go2jsSlogSet(logger);
}

// go2jsSlogSet is where the default logger stands, which is a name rather than a
// constant so that a program which was given a logger of its own logs through
// that one from then on.
let go2jsSlogDefaultHeld = go2jsSlogDefaultLogger;

function go2jsSlogSet(logger) {
	go2jsSlogDefaultHeld = logger;
}

function go2jsSlogWith(args, types) {
	return go2jsSlogLoggerWith(go2jsSlogDefaultHeld, args, types);
}

function go2jsSlogDebug(message, ...rest) {
	let types = null;
	let source = undefined;
	let args = [];
	if (rest.length > 0) {
		const last = rest[rest.length - 1];
		if (typeof last === "string" && last.indexOf(":") >= 0) {
			source = last;
			rest = rest.slice(0, rest.length - 1);
		}
		if (rest.length > 0 && Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}
	return go2jsSlogLoggerDebug(go2jsSlogDefaultHeld, message, ...args, types, source);
}

function go2jsSlogInfo(message, ...rest) {
	let types = null;
	let source = undefined;
	let args = [];
	if (rest.length > 0) {
		const last = rest[rest.length - 1];
		if (typeof last === "string" && last.indexOf(":") >= 0) {
			source = last;
			rest = rest.slice(0, rest.length - 1);
		}
		if (rest.length > 0 && Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}
	return go2jsSlogLoggerInfo(go2jsSlogDefaultHeld, message, ...args, types, source);
}

function go2jsSlogWarn(message, ...rest) {
	let types = null;
	let source = undefined;
	let args = [];
	if (rest.length > 0) {
		const last = rest[rest.length - 1];
		if (typeof last === "string" && last.indexOf(":") >= 0) {
			source = last;
			rest = rest.slice(0, rest.length - 1);
		}
		if (rest.length > 0 && Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}
	return go2jsSlogLoggerWarn(go2jsSlogDefaultHeld, message, ...args, types, source);
}

function go2jsSlogError(message, ...rest) {
	let types = null;
	let source = undefined;
	let args = [];
	if (rest.length > 0) {
		const last = rest[rest.length - 1];
		if (typeof last === "string" && last.indexOf(":") >= 0) {
			source = last;
			rest = rest.slice(0, rest.length - 1);
		}
		if (rest.length > 0 && Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}
	return go2jsSlogLoggerError(go2jsSlogDefaultHeld, message, ...args, types, source);
}

function go2jsSlogDebugContext(ctx, message, ...rest) {
	let types = null;
	let source = undefined;
	let args = [];
	if (rest.length > 0) {
		const last = rest[rest.length - 1];
		if (typeof last === "string" && last.indexOf(":") >= 0) {
			source = last;
			rest = rest.slice(0, rest.length - 1);
		}
		if (rest.length > 0 && Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}
	return go2jsSlogLoggerDebugContext(go2jsSlogDefaultHeld, ctx, message, ...args, types, source);
}

function go2jsSlogInfoContext(ctx, message, ...rest) {
	let types = null;
	let source = undefined;
	let args = [];
	if (rest.length > 0) {
		const last = rest[rest.length - 1];
		if (typeof last === "string" && last.indexOf(":") >= 0) {
			source = last;
			rest = rest.slice(0, rest.length - 1);
		}
		if (rest.length > 0 && Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}
	return go2jsSlogLoggerInfoContext(go2jsSlogDefaultHeld, ctx, message, ...args, types, source);
}

function go2jsSlogWarnContext(ctx, message, ...rest) {
	let types = null;
	let source = undefined;
	let args = [];
	if (rest.length > 0) {
		const last = rest[rest.length - 1];
		if (typeof last === "string" && last.indexOf(":") >= 0) {
			source = last;
			rest = rest.slice(0, rest.length - 1);
		}
		if (rest.length > 0 && Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}
	return go2jsSlogLoggerWarnContext(go2jsSlogDefaultHeld, ctx, message, ...args, types, source);
}

function go2jsSlogErrorContext(ctx, message, ...rest) {
	let types = null;
	let source = undefined;
	let args = [];
	if (rest.length > 0) {
		const last = rest[rest.length - 1];
		if (typeof last === "string" && last.indexOf(":") >= 0) {
			source = last;
			rest = rest.slice(0, rest.length - 1);
		}
		if (rest.length > 0 && Array.isArray(rest[rest.length - 1])) {
			types = rest[rest.length - 1];
			args = rest.slice(0, rest.length - 1);
		} else {
			args = rest;
		}
	}
	return go2jsSlogLoggerErrorContext(go2jsSlogDefaultHeld, ctx, message, ...args, types, source);
}

function go2jsSlogLog(ctx, level, message, args, types, source) {
	return go2jsSlogWriteRecord(go2jsSlogDefaultHeld, ctx, level, message,
		go2jsSlogArgsToAttrs(args === undefined ? [] : args, types), source);
}

function go2jsSlogLogAttrs(ctx, level, message, attrs, source) {
	return go2jsSlogWriteRecord(go2jsSlogDefaultHeld, ctx, level, message, attrs, source);
}

// go2jsSlogSetLogLoggerLevel is the lowest level the default logger writes, which
// is moved from here and is answered with the level it stood at before.
function go2jsSlogSetLogLoggerLevel(level) {
	const held = go2jsSlogLogLoggerLevel;

	go2jsSlogLogLoggerLevel = Number(level) | 0;

	return held;
}

// go2jsSlogNewLogLogger is a logger of the log package that writes through a
// handler, which is how a program that logs the old way is written in the way
// the handler says rather than the way the log package says.
function go2jsSlogNewLogLogger(handler, level) {
	const held = go2jsUnwrap(go2jsUntyped(handler));
	const logger = {handler: held};
	const writer = {
		Write: function(bytes) {
			const text = go2jsBytesToString(bytes);
			const line = go2jsSlogTrimNewline(text);

			// A message of the log package is one message, so a line of its own
			// ends it, and a blank line at the end of it is not written.
			if (line !== "") {
				go2jsSlogWriteRecord(logger, null, level | 0, line, [], "");
			}

			return [bytes.length, null];
		},
	};

	return go2jsLogNew(writer, "", 0);
}

// go2jsSlogTrimNewline is a message without the line ending the log package
// puts at the end of every line it writes.
function go2jsSlogTrimNewline(text) {
	let end = text.length;

	if (end > 0 && text[end - 1] === "\n") {
		end--;
	}

	if (end > 0 && text[end - 1] === "\r") {
		end--;
	}

	return text.slice(0, end);
}

function go2jsSlogSourceString(source) {
	const parts = go2jsSlogSourceParts(source);

	return parts === null ? "" : parts.File + ":" + parts.Line;
}

function go2jsSlogHandlerOptions() {}
go2jsRegisterTypeName(go2jsSlogHandlerOptions, "slog.HandlerOptions");
go2jsSlogHandlerOptions.prototype.Level = undefined;
go2jsSlogHandlerOptions.prototype.AddSource = false;
go2jsSlogHandlerOptions.prototype.ReplaceAttr = undefined;
`
}
