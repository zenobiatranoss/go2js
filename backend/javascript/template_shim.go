package javascript

var templateFuncs = map[string]string{
	"New":  "go2jsTemplateNew",
	"Must": "go2jsTemplateMust",
}

var templateTypes = map[string]string{
	"FuncMap":  "go2jsTemplateFuncMap",
	"Template": "go2jsTemplateNew",
	"HTML":     "go2jsTemplate",
	"JS":       "go2jsTemplate",
	"URL":      "go2jsTemplate",
	"CSS":      "go2jsTemplate",
	"HTMLAttr": "go2jsTemplate",
}

func init() {
	extendedStdlibFuncs()
	moreStdlibFuncs()

	functions := make(map[string]string, len(templateFuncs)+len(templateTypes))

	for name, value := range templateFuncs {
		functions[name] = value
	}

	for name := range templateTypes {
		functions[name] = ""
	}

	supported := funcSet(functions)

	stdlibFuncMaps["text/template"] = functions
	stdlibFuncMaps["html/template"] = functions
	supportedStdlibPackages["text/template"] = supported
	supportedStdlibPackages["html/template"] = supported

	for name, value := range templateTypes {
		packageTypes["template."+name] = value
	}

	for _, alias := range []string{"htmltemplate", "texttemplate"} {
		supportedStdlibPackages[alias] = supported
	}
}

func templateRuntimeSource() string {
	return `function go2jsTemplateFuncMap() {
	return {};
}

function go2jsTemplate(name) {
	const template = {
		__go2js_template: true,
		name: name === undefined || name === null ? "" : String(name),
		tree: null,
		funcs: {}
	};

	template.Parse = function(text) {
		return [template, go2jsTemplateParse(template, text)];
	};

	template.Execute = function(writer, data) {
		return go2jsTemplateExecute(template, writer, data);
	};

	template.ExecuteTemplate = function(name, writer, data) {
		return go2jsTemplateExecuteTemplate(template, name, writer, data);
	};

	template.Funcs = function(funcs) {
		return go2jsTemplateFuncs(template, funcs);
	};

	return template;
}

function go2jsTemplateNew(name) {
	return go2jsTemplate(name);
}

function go2jsTemplateFuncs(template, funcs) {
	if (funcs !== null && funcs !== undefined) {
		for (const key of Object.keys(funcs)) {
			template.funcs[key] = funcs[key];
		}
	}

	return template;
}

function go2jsTemplateMust(result) {
	const values = Array.isArray(result) ? result : [result, null];
	const template = values[0];
	const err = values[1];

	if (err !== null && err !== undefined) {
		throw err;
	}

	return template;
}

function go2jsTemplateParse(template, text) {
	template.tree = go2jsTemplateParseNodes(String(text), null, 0).nodes;

	return null;
}

function go2jsTemplateExecute(template, writer, data) {
	try {
		const out = go2jsTemplateRender(template.tree, data, data, template.funcs, {$: data});
		go2jsTemplateWrite(writer, out);
	} catch (err) {
		return err;
	}

	return null;
}

function go2jsTemplateExecuteTemplate(template, name, writer, data) {
	return go2jsTemplateExecute(template, writer, data);
}

function go2jsTemplateWrite(writer, text) {
	if (writer === null || writer === undefined) {
		return;
	}

	const writeString = go2jsWriterMethod(writer, "WriteString");

	if (writeString !== null) {
		writeString(text);
		return;
	}

	const write = go2jsWriterMethod(writer, "Write");

	if (write === null) {
		throw new TypeError("template: writer does not implement io.Writer");
	}

	write(go2jsStringToBytes(text));
}

function go2jsTemplateParseNodes(text, stops, offset) {
	const nodes = [];
	let index = offset;

	for (;;) {
		const open = text.indexOf("{{", index);

		if (open === -1) {
			if (index < text.length) {
				nodes.push({t: "text", v: text.slice(index)});
			}

			return {nodes: nodes, stop: "", action: "", index: text.length};
		}

		let literal = text.slice(index, open);
		let cursor = open + 2;
		let trimLeft = false;

		if (text[cursor] === "-" && (text[cursor + 1] === " " || text[cursor + 1] === "}")) {
			trimLeft = true;
			cursor++;
		}

		const close = text.indexOf("}}", cursor);

		if (close === -1) {
			throw new Error("template: unclosed action");
		}

		let body = text.slice(cursor, close).trim();
		let trimRight = false;

		if (body.length > 0 && body[body.length - 1] === "-") {
			trimRight = true;
			body = body.slice(0, -1).trim();
		}

		if (trimLeft) {
			literal = literal.replace(/\s+$/, "");
		}

		if (literal !== "") {
			nodes.push({t: "text", v: literal});
		}

		index = close + 2;

		if (trimRight) {
			while (index < text.length && /\s/.test(text[index])) {
				index++;
			}
		}

		if (body === "") {
			continue;
		}

		const keyword = body.split(/[\s(]/)[0];

		if (stops !== null && stops.indexOf(keyword) !== -1) {
			return {nodes: nodes, stop: keyword, action: body.slice(keyword.length).trim(), index: index};
		}

		if (keyword === "end" || keyword === "else") {
			if (stops !== null) {
				return {nodes: nodes, stop: keyword, action: body.slice(keyword.length).trim(), index: index};
			}

			throw new Error("template: unexpected " + keyword);
		}

		// A block action opens a body of its own, and what that body is decided
		// by is the rest of the action rather than the body that follows it.
		if ((keyword === "if" || keyword === "range" || keyword === "with" ||
			keyword === "define" || keyword === "block" || keyword === "template")
			&& body.length > keyword.length) {
			const block = go2jsTemplateParseBlock(text, index, keyword, body.slice(keyword.length).trim());

			nodes.push(block.node);
			index = block.index;

			continue;
		}

		nodes.push({t: "action", action: body});
	}
}

// go2jsTemplateParseBlock reads the body of a block action and the branch that
// follows it, up to the end that closes it. The action itself is read by the
// caller, because the body of a block begins after the action rather than
// inside it.
function go2jsTemplateParseBlock(text, index, keyword, action) {
	const head = go2jsTemplateParseNodes(text, ["else", "end"], index);
	const node = {t: keyword, action: action, body: head.nodes, alt: [], stop: ""};

	if (head.stop === "end") {
		return {node: node, index: head.index};
	}

	const branch = go2jsTemplateParseNodes(text, ["else", "end"], head.index);

	node.alt = branch.nodes;

	if (branch.stop === "end") {
		node.stop = "end";
		return {node: node, index: branch.index};
	}

	// An else that opens another if asks a further question rather than naming a
	// branch of its own, so it is read as the block it is and stands in for the
	// branch. Anything else after an else is a pipeline the else is given.
	const rest = branch.action.trim();
	const nested = rest.startsWith("if")
		? go2jsTemplateParseBlock(text, branch.index, "if", rest.slice(2).trim())
		: go2jsTemplateParseBlock(text, branch.index, "if", rest);

	node.alt = [nested.node];

	return {node: node, index: nested.index};
}

function go2jsTemplateRender(nodes, dot, root, funcs, vars) {
	let out = "";

	for (const node of nodes) {
		switch (node.t) {
		case "text":
			out += node.v;
			break;

		case "action":
			out += go2jsTemplateString(go2jsTemplateEvalAction(node.action, dot, root, funcs, vars));
			break;

		case "if": {
			const value = go2jsTemplateEval(node.action, dot, root, funcs, vars);

			if (go2jsTemplateTruth(value)) {
				out += go2jsTemplateRender(node.body, dot, root, funcs, vars);
			} else {
				out += go2jsTemplateRender(node.alt, dot, root, funcs, vars);
			}

			break;
		}

		case "with": {
			const value = go2jsTemplateEval(node.action, dot, root, funcs, vars);

			if (go2jsTemplateTruth(value)) {
				out += go2jsTemplateRender(node.body, value, root, funcs, go2jsTemplateVars(vars, {}));
			} else {
				out += go2jsTemplateRender(node.alt, dot, root, funcs, vars);
			}

			break;
		}

		case "range": {
			out += go2jsTemplateRenderRange(node, dot, root, funcs, vars);
			break;
		}

		case "define":
		case "template":
		case "block":
			out += go2jsTemplateRender(node.body, dot, root, funcs, vars);
			break;
		}
	}

	return out;
}

function go2jsTemplateVars(vars, extra) {
	const merged = {};

	for (const key of Object.keys(vars)) {
		merged[key] = vars[key];
	}

	for (const key of Object.keys(extra)) {
		merged[key] = extra[key];
	}

	return merged;
}

function go2jsTemplateRenderRange(node, dot, root, funcs, vars) {
	const declared = go2jsTemplateRangeVars(node.action);
	const pipeline = go2jsTemplateStripRangeVars(declared, node.action);
	const value = go2jsTemplateEval(pipeline, dot, root, funcs, vars);
	let out = "";

	if (value === null || value === undefined) {
		return go2jsTemplateRender(node.alt, dot, root, funcs, vars);
	}

	if (Array.isArray(value)) {
		if (value.length === 0) {
			return go2jsTemplateRender(node.alt, dot, root, funcs, vars);
		}

		for (let index = 0; index < value.length; index++) {
			out += go2jsTemplateRenderItem(node, value[index], index, declared, dot, root, funcs, vars);
		}

		return out;
	}

	if (typeof value === "string") {
		for (let index = 0; index < value.length; index++) {
			out += go2jsTemplateRenderItem(node, value[index], index, declared, dot, root, funcs, vars);
		}

		return out;
	}

	if (value instanceof go2jsNativeMap) {
		if (value.size === 0) {
			return go2jsTemplateRender(node.alt, dot, root, funcs, vars);
		}

		const keys = go2jsMapKeys(value);
		let last;

		for (const key of keys) {
			last = [key, go2jsMapIndex(value, key)];
		}

		for (const pair of keys.map((key) => [key, go2jsMapIndex(value, key)])) {
			out += go2jsTemplateRenderItem(node, pair[1], pair[0], declared, dot, root, funcs, vars);
		}

		return out;
	}

	if (typeof value === "object") {
		const keys = Object.keys(value);

		if (keys.length === 0) {
			return go2jsTemplateRender(node.alt, dot, root, funcs, vars);
		}

		for (const key of keys) {
			out += go2jsTemplateRenderItem(node, value[key], key, declared, dot, root, funcs, vars);
		}

		return out;
	}

	out += go2jsTemplateRender(node.body, dot, root, funcs, vars);

	return out;
}

function go2jsTemplateRenderItem(node, item, key, declared, dot, root, funcs, vars) {
	const extra = {};

	if (declared.length === 1) {
		extra[declared[0]] = item;
	} else if (declared.length >= 2) {
		extra[declared[0]] = key;
		extra[declared[1]] = item;
	}

	return go2jsTemplateRender(node.body, item, root, funcs, go2jsTemplateVars(vars, extra));
}

function go2jsTemplateRangeVars(action) {
	const assign = action.indexOf(":=");
	const declare = action.indexOf(" = ");

	let cut = -1;

	if (assign !== -1) {
		cut = assign;
	} else if (declare !== -1) {
		cut = declare;
	}

	if (cut === -1) {
		return [];
	}

	// The names are written in a list, so each one but the last carries the comma
	// that separates it from the next.
	return action.slice(0, cut).trim().split(/[\s,]+/).filter((name) => name.startsWith("$"));
}

function go2jsTemplateStripRangeVars(declared, action) {
	if (declared.length === 0) {
		return action;
	}

	const assign = action.indexOf(":=");
	const declare = action.indexOf(" = ");
	const cut = assign !== -1 ? assign : declare;

	return action.slice(cut + (assign !== -1 ? 2 : 3)).trim();
}

function go2jsTemplateEvalAction(action, dot, root, funcs, vars) {
	const assign = action.indexOf(":=");
	const equals = action.indexOf(" = ");

	if (assign !== -1) {
		const name = action.slice(0, assign).trim();

		if (name.startsWith("$")) {
			vars[name] = go2jsTemplateEval(action.slice(assign + 2).trim(), dot, root, funcs, vars);
			return "";
		}
	}

	if (equals !== -1 && !action.includes("==") && !action.includes("!=")) {
		const name = action.slice(0, equals).trim();

		if (name.startsWith("$")) {
			vars[name] = go2jsTemplateEval(action.slice(equals + 3).trim(), dot, root, funcs, vars);
			return "";
		}
	}

	return go2jsTemplateEval(action, dot, root, funcs, vars);
}

function go2jsTemplateEval(pipeline, dot, root, funcs, vars) {
	const stages = go2jsTemplateSplit(pipeline, "|");

	let value = go2jsTemplateEvalCommand(stages[0], dot, root, funcs, vars);

	for (let index = 1; index < stages.length; index++) {
		const fn = go2jsTemplateEvalCommand(stages[index], dot, root, funcs, vars);

		if (typeof fn !== "function") {
			throw new Error("template: non-function in pipeline");
		}

		value = go2jsCallNow(fn, null, [value]);
	}

	return value;
}

function go2jsTemplateSplit(pipeline, separator) {
	const parts = [];
	let depth = 0;
	let quote = "";
	let current = "";

	for (const char of pipeline) {
		if (quote !== "") {
			current += char;

			if (char === quote) {
				quote = "";
			}

			continue;
		}

		if (char === '"' || char === "'") {
			quote = char;
			current += char;
			continue;
		}

		if (char === "(") {
			depth++;
		} else if (char === ")") {
			depth--;
		}

		if (char === separator && depth === 0) {
			parts.push(current);
			current = "";
			continue;
		}

		current += char;
	}

	parts.push(current);

	return parts;
}

function go2jsTemplateEvalCommand(command, dot, root, funcs, vars) {
	const tokens = go2jsTemplateTokenize(command);

	if (tokens.length === 0) {
		return "";
	}

	const args = [];

	for (let index = 1; index < tokens.length; index++) {
		args.push(go2jsTemplateEvalToken(tokens[index], dot, root, funcs, vars));
	}

	return go2jsTemplateResolve(tokens[0], args, dot, root, funcs, vars);
}

function go2jsTemplateTokenize(command) {
	const tokens = [];
	let index = 0;

	while (index < command.length) {
		const char = command[index];

		if (/\s/.test(char)) {
			index++;
			continue;
		}

		if (char === "(") {
			let depth = 1;
			let cursor = index + 1;
			let body = "";

			while (cursor < command.length && depth > 0) {
				if (command[cursor] === "(") {
					depth++;
				} else if (command[cursor] === ")") {
					depth--;

					if (depth === 0) {
						break;
					}
				}

				body += command[cursor];
				cursor++;
			}

			tokens.push({k: "group", v: body});
			index = cursor + 1;
			continue;
		}

		if (char === '"' || char === "'") {
			let cursor = index + 1;
			let body = "";

			while (cursor < command.length && command[cursor] !== char) {
				if (char === '"' && command[cursor] === "\\") {
					cursor++;
				}

				body += command[cursor];
				cursor++;
			}

			tokens.push({k: "string", v: body});
			index = cursor + 1;
			continue;
		}

		let word = "";

		while (index < command.length && !/\s/.test(command[index]) && command[index] !== "(") {
			word += command[index];
			index++;
		}

		if (word !== "") {
			tokens.push({k: "word", v: word});
		}
	}

	return tokens;
}

function go2jsTemplateEvalToken(token, dot, root, funcs, vars) {
	if (token.k === "group") {
		return go2jsTemplateEval(token.v, dot, root, funcs, vars);
	}

	if (token.k === "string") {
		return token.v;
	}

	return go2jsTemplateResolve(token, [], dot, root, funcs, vars);
}

function go2jsTemplateResolve(token, args, dot, root, funcs, vars) {
	if (token.k === "group") {
		return go2jsTemplateEval(token.v, dot, root, funcs, vars);
	}

	if (token.k === "string") {
		return token.v;
	}

	const word = token.v;

	if (word === ".") {
		return go2jsTemplateCall(dot, args);
	}

	// A variable is named up to the first dot, and whatever follows that dot is
	// a field of the value the variable holds rather than part of its name.
	if (word.startsWith("$")) {
		const dot2 = word.indexOf(".");
		const name = dot2 === -1 ? word : word.slice(0, dot2);
		const base = name === "$" ? root : vars[name];

		return go2jsTemplateField(word.slice(name.length), base, args, dot);
	}

	if (word.startsWith(".")) {
		return go2jsTemplateField(word, dot, args, dot);
	}

	if (word === "true") {
		return true;
	}

	if (word === "false") {
		return false;
	}

	if (word === "nil") {
		return null;
	}

	if (/^-?[0-9]+$/.test(word)) {
		return parseInt(word, 10);
	}

	if (/^-?[0-9]*\.[0-9]+$/.test(word)) {
		return parseFloat(word);
	}

	const builtin = go2jsTemplateBuiltin(word, funcs, vars);

	if (builtin !== undefined) {
		return go2jsCallNow(builtin, null, args);
	}

	if (funcs[word] !== undefined) {
		const fn = funcs[word];

		return typeof fn === "function" ? go2jsTemplateCall(fn, args) : fn;
	}

	if (args.length === 0) {
		return "";
	}

	throw new Error("template: function " + word + " not defined");
}

function go2jsTemplateField(word, base, args, dot) {
	if (word === ".") {
		return go2jsTemplateCall(base, args);
	}

	const parts = word.split(".").slice(1);
	let value = base;

	for (const part of parts) {
		if (value === null || value === undefined) {
			return null;
		}

		value = go2jsTemplateProperty(value, part);
	}

	if (value === null || value === undefined) {
		return null;
	}

	if (typeof value === "function" && (args.length > 0 || !go2jsTemplateHasField(base, parts))) {
		return go2jsTemplateCall(value, args);
	}

	if (typeof value === "function" && args.length === 0) {
		return go2jsTemplateCall(value, args);
	}

	return value;
}

function go2jsTemplateHasField(base, parts) {
	if (base === null || base === undefined) {
		return false;
	}

	const last = parts[parts.length - 1];
	let value = base;

	for (const part of parts) {
		if (value === null || value === undefined) {
			return false;
		}

		value = go2jsTemplateProperty(value, part);
	}

	return !(typeof value === "function") && last !== undefined;
}

function go2jsTemplateProperty(value, name) {
	if (value === null || value === undefined) {
		return null;
	}

	if (name.startsWith("$")) {
		return null;
	}

	const direct = value[name];

	if (typeof direct === "function") {
		return direct.bind(value);
	}

	return direct;
}

function go2jsTemplateCall(fn, args) {
	if (typeof fn !== "function") {
		return fn;
	}

	return go2jsCallNow(fn, null, args);
}

function go2jsTemplateBuiltin(word, funcs, vars) {
	switch (word) {
	case "and":
		return function(...args) {
			let last = args[0];

			for (const value of args) {
				if (!go2jsTemplateTruth(value)) {
					return value;
				}

				last = value;
			}

			return last;
		};

	case "or":
		return function(...args) {
			for (const value of args) {
				if (go2jsTemplateTruth(value)) {
					return value;
				}
			}

			return args.length > 0 ? args[args.length - 1] : undefined;
		};

	case "not":
		return (value) => !go2jsTemplateTruth(value);

	case "eq":
		return (...args) => args.slice(1).some((value) => go2jsTemplateEqual(args[0], value));

	case "ne":
		return (left, right) => !go2jsTemplateEqual(left, right);

	case "lt":
		return (left, right) => go2jsTemplateCompare(left, right) < 0;

	case "le":
		return (left, right) => go2jsTemplateCompare(left, right) <= 0;

	case "gt":
		return (left, right) => go2jsTemplateCompare(left, right) > 0;

	case "ge":
		return (left, right) => go2jsTemplateCompare(left, right) >= 0;

	case "len":
		return (value) => go2jsTemplateLength(value);

	case "index":
		return (value, ...keys) => go2jsTemplateIndex(value, keys);

	case "printf":
		return (format, ...args) => go2jsSprintf(String(format), ...args);

	case "print":
		return (...args) => args.map((value) => go2jsTemplateString(value)).join("");

	case "println":
		return (...args) => args.map((value) => go2jsTemplateString(value)).join(" ") + "\n";

	case "slice":
		return (value, ...bounds) => go2jsTemplateIndex(value, bounds);
	}

	return undefined;
}

function go2jsTemplateIndex(value, keys) {
	let current = value;

	for (const key of keys) {
		if (current === null || current === undefined) {
			return null;
		}

		if (Array.isArray(current)) {
			current = current[Number(key)];
			continue;
		}

		if (current instanceof go2jsNativeMap) {
			current = go2jsMapIndex(current, key);
			continue;
		}

		current = current[key];
	}

	return current === undefined ? null : current;
}

function go2jsTemplateLength(value) {
	if (value === null || value === undefined) {
		return 0;
	}

	if (typeof value === "string" || Array.isArray(value)) {
		return value.length;
	}

	if (value instanceof go2jsNativeMap) {
		return value.size;
	}

	if (typeof value === "object") {
		return Object.keys(value).length;
	}

	return 0;
}

function go2jsTemplateEqual(left, right) {
	if (left === right) {
		return true;
	}

	if (left === null || left === undefined || right === null || right === undefined) {
		return false;
	}

	if (typeof left === "number" && typeof right === "number") {
		return left === right;
	}

	if (typeof left === "string" && typeof right === "string") {
		return left === right;
	}

	return go2jsEqual(left, right) === true;
}

function go2jsTemplateCompare(left, right) {
	if (typeof left === "string" && typeof right === "string") {
		return left < right ? -1 : left > right ? 1 : 0;
	}

	const a = Number(left);
	const b = Number(right);

	return a < b ? -1 : a > b ? 1 : 0;
}

function go2jsTemplateTruth(value) {
	if (value === null || value === undefined || value === false) {
		return false;
	}

	if (value === true) {
		return true;
	}

	if (typeof value === "number") {
		return value !== 0;
	}

	if (typeof value === "string") {
		return value.length > 0;
	}

	if (Array.isArray(value)) {
		return value.length > 0;
	}

	if (value instanceof go2jsNativeMap) {
		return value.size > 0;
	}

	if (typeof value === "object") {
		return Object.keys(value).length > 0;
	}

	return true;
}

// A number reaches a template as whatever arithmetic left behind, so it is
// written the way the language writes one: without a trailing fraction, and in
// the exponent form once the number is too large or too small to sit on one
// line.
function go2jsTemplateNumber(value) {
	if (typeof go2jsFormatFloatDefault === "function") {
		return go2jsFormatFloatDefault(value);
	}

	return String(value);
}

function go2jsTemplateString(value) {
	if (value === null || value === undefined) {
		return "";
	}

	if (typeof value === "string") {
		return value;
	}

	if (typeof value === "function") {
		return "";
	}

	if (typeof value === "boolean") {
		return value ? "true" : "false";
	}

	if (typeof value === "number") {
		return go2jsTemplateNumber(value);
	}

	if (value.__go2js_reflectValue === true) {
		return go2jsTemplateString(value.v);
	}

	return go2jsFormat(value);
}

function go2jsTemplateRenderError(err) {
	return err === null || err === undefined ? "" : String(err && err.message ? err.message : err);
}
`
}
