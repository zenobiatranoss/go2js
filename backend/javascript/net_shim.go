package javascript

import (
	"go/ast"

	gotypesstd "go/types"
)

var netFuncs = map[string]string{
	"ParseIP":   "go2jsNetParseIP",
	"ParseCIDR": "go2jsNetParseCIDR",
	"IPv4Mask":  "go2jsNetIPv4Mask",
	"CIDRMask":  "go2jsNetCIDRMask",
	"IPv4":      "go2jsNetIPv4",
}

var netTypes = map[string]string{
	"IP":     "go2jsNetIP",
	"IPMask": "go2jsNetIPMask",
}

var netPlainTypes = []string{
	"IPNet",
	"IPAddr",
}

var netConstants = map[string]string{
	"IPv4len": "4",
	"IPv6len": "16",
}

var shimValueMethods = map[string]string{
	"net.IP.String":        "go2jsNetIPString",
	"net.IP.Equal":         "go2jsNetIPEqual",
	"net.IP.To4":           "go2jsNetIPTo4",
	"net.IP.To16":          "go2jsNetIPTo16",
	"net.IP.IsUnspecified": "go2jsNetIPIsUnspecified",
	"net.IP.IsLoopback":    "go2jsNetIPIsLoopback",
	"net.IP.IsPrivate":     "go2jsNetIPIsPrivate",
	"net.IPMask.String":    "go2jsNetIPMaskString",
	"net.IPNet.String":     "go2jsNetIPNetString",
	"net.IPNet.Contains":   "go2jsNetIPNetContains",
	"net.IPAddr.String":    "go2jsNetIPAddrString",

	"sort.StringSlice.Sort":  "go2jsSortStringSlice",
	"sort.IntSlice.Sort":     "go2jsSortIntSlice",
	"sort.Float64Slice.Sort": "go2jsSortFloat64Slice",
}

func netRuntimeSource() string {
	return `function go2jsNetIP(bytes) {
	return go2jsNetBytes(bytes, 4);
}

function go2jsNetIPMask(bytes) {
	return go2jsNetBytes(bytes, 16);
}

function go2jsNetBytes(bytes, size) {
	const raw = go2jsToArray(bytes === undefined || bytes === null ? null : go2jsUntyped(bytes));
	const padded = [];

	for (let index = 0; index < raw.length; index++) {
		padded.push(Number(raw[index]) & 255);
	}

	while (padded.length < size) {
		padded.push(0);
	}

	if (padded.length > size && size === 4) {
		return padded.slice(padded.length - 4);
	}

	return padded.slice(0, 16);
}

function go2jsNetIPAddr(ip, zone) {
	return {ip: go2jsNetIP(ip), zone: zone || ""};
}

function go2jsNetIsIPv4(bytes) {
	if (bytes.length === 4) {
		return true;
	}

	if (bytes.length !== 16) {
		return false;
	}

	for (let index = 0; index < 10; index++) {
		if (bytes[index] !== 0) {
			return false;
		}
	}

	return bytes[10] === 255 && bytes[11] === 255;
}

function go2jsNetIPString(value) {
	const bytes = go2jsNetToArray(value);

	if (bytes === null) {
		return "<nil>";
	}

	if (bytes.length === 0) {
		return "<nil>";
	}

	if (go2jsNetIsIPv4(bytes)) {
		const start = bytes.length === 16 ? 12 : 0;
		return bytes.slice(start).join(".");
	}

	const groups = [];

	for (let index = 0; index < 16; index += 2) {
		groups.push(((bytes[index] << 8) | bytes[index + 1]).toString(16));
	}

	let address = groups.join(":").replace(/:(0:)+/g, ":").replace(/^0:/, "").replace(/:$/, "");

	if (address.indexOf("::") === -1) {
		address = address.replace(/:(0:)+/g, ":").replace(/:$/, ":0");
	}

	if (address.indexOf("::") !== -1 && address.indexOf("::", 1) === -1) {
		address = address + "0:";
	}

	return address;
}

function go2jsNetIPEqual(left, right) {
	const a = go2jsNetToArray(left);
	const b = go2jsNetToArray(right);

	if (a === null || b === null) {
		return a === b;
	}

	const a4 = go2jsNetIsIPv4(a) ? a.slice(a.length - 4) : a;
	const b4 = go2jsNetIsIPv4(b) ? b.slice(b.length - 4) : b;

	if (a4.length !== b4.length) {
		return false;
	}

	for (let index = 0; index < a4.length; index++) {
		if (a4[index] !== b4[index]) {
			return false;
		}
	}

	return true;
}

function go2jsNetIPTo4(value) {
	const bytes = go2jsNetToArray(value);

	if (bytes === null || !go2jsNetIsIPv4(bytes)) {
		return null;
	}

	return bytes.slice(bytes.length - 4);
}

function go2jsNetIPTo16(value) {
	const bytes = go2jsNetToArray(value);

	if (bytes === null) {
		return null;
	}

	return go2jsNetIsIPv4(bytes) ? go2jsNetIPv4To16(bytes) : bytes.slice(0, 16);
}

function go2jsNetIPv4To16(bytes) {
	const padded = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 255];

	return padded.concat(bytes.slice(bytes.length - 4));
}

function go2jsNetIPIsUnspecified(value) {
	const bytes = go2jsNetToArray(value);

	if (bytes === null) {
		return false;
	}

	for (let index = 0; index < bytes.length; index++) {
		if (bytes[index] !== 0) {
			return false;
		}
	}

	return true;
}

function go2jsNetIPIsLoopback(value) {
	const bytes = go2jsNetIPTo4(value);

	return bytes !== null && bytes[0] === 127;
}

function go2jsNetIPIsPrivate(value) {
	const bytes = go2jsNetIPTo4(value);

	if (bytes === null) {
		return false;
	}

	return bytes[0] === 10 ||
		(bytes[0] === 172 && bytes[1] >= 16 && bytes[1] <= 31) ||
		(bytes[0] === 192 && bytes[1] === 168);
}

function go2jsNetIPMaskString(value) {
	const bytes = go2jsNetToArray(value);

	if (bytes === null) {
		return "<nil>";
	}

	let text = "";

	for (let index = 0; index < bytes.length; index++) {
		text += (Number(bytes[index]) & 255).toString(16).padStart(2, "0");
	}

	return text;
}

function go2jsNetToArray(value) {
	if (value === null || value === undefined) {
		return null;
	}

	if (value !== null && typeof value === "object" && Array.isArray(value.ip)) {
		return value.ip;
	}

	return go2jsToArray(go2jsUntyped(value));
}

function go2jsNetIPNetString(value) {
	if (value === null || value === undefined) {
		return "<nil>";
	}

	const address = go2jsNetIPString(value.ip);
	const mask = go2jsNetMaskBits(value.mask);

	return address + "/" + String(mask);
}

function go2jsNetMaskBits(mask) {
	const bytes = go2jsNetToArray(mask);

	if (bytes === null) {
		return 0;
	}

	let bits = 0;

	for (let index = 0; index < bytes.length; index++) {
		let byte = Number(bytes[index]) & 255;

		for (let shift = 7; shift >= 0; shift--) {
			if (((byte >> shift) & 1) === 0) {
				return bits;
			}

			bits++;
		}
	}

	return bits;
}

function go2jsNetIPNetContains(network, address) {
	const wanted = go2jsNetIPTo4(address);
	const network4 = go2jsNetIPTo4(network.ip);
	const mask = go2jsNetToArray(network.mask);

	if (wanted === null || network4 === null || mask === null) {
		return false;
	}

	const start = mask.length === 16 ? 12 : 0;

	for (let index = 0; index < 4; index++) {
		if ((wanted[index] & (mask[start + index] || 0)) !== (network4[index] & mask[start + index])) {
			return false;
		}
	}

	return true;
}

function go2jsNetIPAddrString(value) {
	if (value === null || value === undefined) {
		return "<nil>";
	}

	if (!value.zone) {
		return go2jsNetIPString(value.ip);
	}

	return go2jsNetIPString(value.ip) + "%" + value.zone;
}

function go2jsNetIPv4Mask(a, b, c, d) {
	return go2jsNetBytes([Number(a) & 255, Number(b) & 255, Number(c) & 255, Number(d) & 255], 4);
}

function go2jsNetCIDRMask(ones, bits) {
	if (bits !== 32) {
		if (bits !== 128) {
			return null;
		}

		const wide = new Array(16).fill(0);
		let left = ones;

		for (let index = 0; index < 16 && left > 0; index++) {
			const take = Math.min(8, left);
			wide[index] = (0xff << (8 - take)) & 255;
			left -= take;
		}

		return go2jsNetIPMask(wide);
	}

	return go2jsNetIPMask(go2jsNetParseCIDRMask(String(ones)));
}

function go2jsNetIPv4(a, b, c, d) {
	return go2jsNetIPv4To16([Number(a) & 255, Number(b) & 255, Number(c) & 255, Number(d) & 255]);
}

function go2jsNetParseIPv4(text) {
	const parts = text.split(".");

	if (parts.length !== 4) {
		return null;
	}

	const bytes = [];

	for (let index = 0; index < 4; index++) {
		const part = parts[index];

		if (part === "" || part.length > 3 || !/^[0-9]+$/.test(part)) {
			return null;
		}

		const value = Number(part);

		if (value > 255) {
			return null;
		}

		if (part.length > 1 && part[0] === "0") {
			return null;
		}

		bytes.push(value);
	}

	return bytes;
}

function go2jsNetParseIPv6(text) {
	let tail = text;
	let head = [];

	const double = tail.indexOf("::");

	if (double !== -1) {
		head = tail.slice(0, double) === "" ? [] : tail.slice(0, double).split(":");
		tail = tail.slice(double + 2) === "" ? [] : tail.slice(double + 2).split(":");

		if (head.length + tail.length > 7) {
			return null;
		}

		const fill = new Array(8 - head.length - tail.length).fill("0");
		head = head.concat(fill).concat(tail);
	} else {
		head = tail.split(":");

		if (head.length !== 8) {
			return null;
		}
	}

	const bytes = [];

	for (let index = 0; index < 8; index++) {
		const part = head[index];

		if (part.indexOf(".") !== -1) {
			if (index !== 7) {
				return null;
			}

			const embedded = go2jsNetParseIPv4(part);

			if (embedded === null) {
				return null;
			}

			bytes.push(embedded[0], embedded[1], embedded[2], embedded[3]);
			continue;
		}

		if (part === "" || part.length > 4 || !/^[0-9A-Fa-f]+$/.test(part)) {
			return null;
		}

		const value = parseInt(part, 16);
		bytes.push((value >> 8) & 255, value & 255);
	}

	if (bytes.length !== 16) {
		return null;
	}

	return bytes;
}

function go2jsNetParseIP(text) {
	if (typeof text !== "string") {
		return null;
	}

	const value = text.trim();

	if (value === "") {
		return null;
	}

	if (value.indexOf(":") !== -1) {
		const parsed = go2jsNetParseIPv6(value);
		return parsed === null ? null : go2jsNetIP(parsed);
	}

	const parsed = go2jsNetParseIPv4(value);
	return parsed === null ? null : go2jsNetIP(parsed);
}

function go2jsNetParseCIDRMask(text) {
	if (typeof text !== "string" || !/^[0-9]+$/.test(text)) {
		return null;
	}

	const bits = Number(text);

	if (bits > 32) {
		return null;
	}

	const bytes = new Array(4).fill(0);
	let left = bits;

	for (let index = 0; index < 4 && left > 0; index++) {
		const take = Math.min(8, left);
		bytes[index] = (0xff << (8 - take)) & 255;
		left -= take;
	}

	return bytes;
}

function go2jsNetParseCIDR(text) {
	if (typeof text !== "string") {
		return [null, null, new Error("invalid CIDR address: " + String(text))];
	}

	const value = text.trim();
	const slash = value.indexOf("/");

	if (slash === -1) {
		return [null, null, new Error("invalid CIDR address: " + value)];
	}

	const address = go2jsNetParseIP(value.slice(0, slash));

	if (address === null) {
		return [null, null, new Error("invalid CIDR address: " + value)];
	}

	const isV4 = go2jsNetIsIPv4(address);
	const mask = go2jsNetParseCIDRMask(value.slice(slash + 1));

	if (mask === null) {
		return [null, null, new Error("invalid CIDR address: " + value)];
	}

	let bits = 0;

	for (let index = 0; index < 4; index++) {
		let byte = mask[index];

		for (let shift = 7; shift >= 0; shift--) {
			if (((byte >> shift) & 1) === 0) {
				break;
			}

			bits++;
		}
	}

	const network = new Array(4);

	for (let index = 0; index < 4; index++) {
		network[index] = address[address.length - 4 + index] & mask[index];
	}

	if (isV4) {
		return [go2jsNetIP(address), {ip: network, mask: mask.slice(0)}, null];
	}

	const wide = new Array(12).fill(0).concat(network);
	const wideMask = new Array(12).fill(0).concat(mask);

	return [go2jsNetIP(address), {ip: wide, mask: wideMask}, null];
}
`
}

func init() {
	extendedStdlibFuncs()
	moreStdlibFuncs()

	functions := make(map[string]string, len(netFuncs)+len(netTypes)+len(netConstants))

	for name, value := range netFuncs {
		functions[name] = value
	}

	for name := range netTypes {
		functions[name] = ""
	}

	for _, name := range netPlainTypes {
		functions[name] = ""
	}

	for name := range netConstants {
		functions[name] = ""
	}

	stdlibFuncMaps["net"] = functions
	supportedStdlibPackages["net"] = funcSet(functions)

	for name, value := range netTypes {
		packageTypes["net."+name] = value
	}

	for name, value := range netConstants {
		packageConstants["net."+name] = value
	}
}

func shimMethodKey(selector *ast.SelectorExpr, receiver gotypesstd.Type) (string, bool) {
	if selector == nil || selector.Sel == nil || receiver == nil {
		return "", false
	}

	if pointer, ok := receiver.(*gotypesstd.Pointer); ok {
		receiver = pointer.Elem()
	}

	named, ok := receiver.(*gotypesstd.Named)
	if !ok {
		return "", false
	}

	object := named.Obj()
	if object == nil || object.Pkg() == nil {
		return "", false
	}

	return object.Pkg().Path() + "." + object.Name() + "." + selector.Sel.Name, true
}

func (e *emitter) emitShimValueMethodCall(call *ast.CallExpr, selector *ast.SelectorExpr) (bool, error) {
	key, ok := shimMethodKey(selector, e.analyzedType(selector.X))
	if !ok {
		return false, nil
	}

	helper, ok := shimValueMethods[key]
	if !ok {
		return false, nil
	}

	e.needsRuntime = true

	if osFileMultiReturn[key] {
		e.write("[")
	}

	e.write(helper)
	e.write("(")

	if err := e.emitExpr(selector.X); err != nil {
		return true, err
	}

	for _, arg := range call.Args {
		e.write(", ")

		if err := e.emitExpr(arg); err != nil {
			return true, err
		}
	}

	e.write(")")

	if osFileMultiReturn[key] {
		e.write("]")
	}

	return true, nil
}
