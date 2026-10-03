package javascript

// The math/bits package works on the bits of a number rather than on its value.
// JavaScript has no such numbers, so each function here reads the width it was
// asked about and works out the same count Go works out. Every function takes
// its arguments in the width its own name says: a 32 bit function reads two
// arguments, and an unsized one reads as many as the call holds.
var bitsFuncs = map[string]string{
	"OnesCount":       "go2jsBitsOnesCount",
	"OnesCount8":      "go2jsBitsOnesCount8",
	"OnesCount16":     "go2jsBitsOnesCount16",
	"OnesCount32":     "go2jsBitsOnesCount32",
	"OnesCount64":     "go2jsBitsOnesCount64",
	"LeadingZeros":    "go2jsBitsLeadingZeros",
	"LeadingZeros8":   "go2jsBitsLeadingZeros8",
	"LeadingZeros16":  "go2jsBitsLeadingZeros16",
	"LeadingZeros32":  "go2jsBitsLeadingZeros32",
	"LeadingZeros64":  "go2jsBitsLeadingZeros64",
	"TrailingZeros":   "go2jsBitsTrailingZeros",
	"TrailingZeros8":  "go2jsBitsTrailingZeros8",
	"TrailingZeros16": "go2jsBitsTrailingZeros16",
	"TrailingZeros32": "go2jsBitsTrailingZeros32",
	"TrailingZeros64": "go2jsBitsTrailingZeros64",
	"Len":             "go2jsBitsLen",
	"Len8":            "go2jsBitsLen8",
	"Len16":           "go2jsBitsLen16",
	"Len32":           "go2jsBitsLen32",
	"Len64":           "go2jsBitsLen64",
	"Reverse":         "go2jsBitsReverse",
	"Reverse8":        "go2jsBitsReverse8",
	"Reverse16":       "go2jsBitsReverse16",
	"Reverse32":       "go2jsBitsReverse32",
	"Reverse64":       "go2jsBitsReverse64",
	"RotateLeft":      "go2jsBitsRotateLeft",
	"RotateLeft8":     "go2jsBitsRotateLeft8",
	"RotateLeft16":    "go2jsBitsRotateLeft16",
	"RotateLeft32":    "go2jsBitsRotateLeft32",
	"RotateLeft64":    "go2jsBitsRotateLeft64",
	"Rotate":          "go2jsBitsRotate",
	"Rotate8":         "go2jsBitsRotate8",
	"Rotate16":        "go2jsBitsRotate16",
	"Rotate32":        "go2jsBitsRotate32",
	"Rotate64":        "go2jsBitsRotate64",
	"Add":             "go2jsBitsAdd",
	"Sub":             "go2jsBitsSub",
	"Mul":             "go2jsBitsMul",
	"Div":             "go2jsBitsDiv",
	"Rem":             "go2jsBitsRem",
	"Uint":            "go2jsBitsUint",
}

func mathBitsRuntimeSource() string {
	return `
// go2jsBitsWide turns a number into the whole number JavaScript keeps for counts
// wider than a double holds, so the bits of a count stay the bits it was given.
function go2jsBitsWide(value) {
	if (typeof value === "bigint") {
		return value;
	}

	return BigInt(Math.trunc(Number(value)));
}

// go2jsBitsNarrow turns a whole number back into a plain number for the widths a
// double holds exactly, which is every width short of sixty four bits. A width
// of sixty four bits is left as a whole number of its own kind, because a double
// cannot hold the top of that range exactly.
function go2jsBitsNarrow(value) {
	return typeof value === "bigint" && value <= 9007199254740991n && value >= -9007199254740991n
		? Number(value)
		: value;
}

// go2jsBitsMask keeps only the low width bits of a count, which is what a count
// of that width holds and what every answer below is worked out from.
function go2jsBitsMask(value, width) {
	const masked = go2jsBitsWide(value) & ((1n << BigInt(width)) - 1n);

	if (masked < 0n) {
		return masked + (1n << BigInt(width));
	}

	return masked;
}

// go2jsBitsSigned reads a count as a signed number of the width it was asked
// about, so a count whose top bit is set stands for a negative number.
function go2jsBitsSigned(value, width) {
	const masked = go2jsBitsMask(value, width);

	if (masked >= 1n << BigInt(width - 1)) {
		return masked - (1n << BigInt(width));
	}

	return masked;
}

function go2jsBitsOnesCount(value) {
	let remaining = go2jsBitsMask(value, 64);
	let count = 0;

	while (remaining !== 0n) {
		count += Number(remaining & 1n);
		remaining >>= 1n;
	}

	return count;
}

function go2jsBitsOnesCount8(value) {
	return go2jsBitsOnesCount(go2jsBitsMask(value, 8));
}

function go2jsBitsOnesCount16(value) {
	return go2jsBitsOnesCount(go2jsBitsMask(value, 16));
}

function go2jsBitsOnesCount32(value) {
	return go2jsBitsOnesCount(go2jsBitsMask(value, 32));
}

function go2jsBitsOnesCount64(value) {
	return go2jsBitsOnesCount(go2jsBitsMask(value, 64));
}

function go2jsBitsLeadingZeros(value, width) {
	const masked = go2jsBitsMask(value, width);

	if (masked === 0n) {
		return width;
	}

	let count = 0;

	for (let bit = width - 1; bit >= 0; bit--) {
		if ((masked >> BigInt(bit)) & 1n) {
			break;
		}

		count++;
	}

	return count;
}

function go2jsBitsLeadingZeros8(value) {
	return go2jsBitsLeadingZeros(value, 8);
}

function go2jsBitsLeadingZeros16(value) {
	return go2jsBitsLeadingZeros(value, 16);
}

function go2jsBitsLeadingZeros32(value) {
	return go2jsBitsLeadingZeros(value, 32);
}

function go2jsBitsLeadingZeros64(value) {
	return go2jsBitsLeadingZeros(value, 64);
}

function go2jsBitsTrailingZeros(value, width) {
	const masked = go2jsBitsMask(value, width);

	if (masked === 0n) {
		return width;
	}

	let count = 0;

	// The count of trailing zeros is the number of low places the count does
	// not have a one in, so the walk goes on until it finds one.
	while (((masked >> BigInt(count)) & 1n) === 0n) {
		count++;
	}

	return count;
}

function go2jsBitsTrailingZeros8(value) {
	return go2jsBitsTrailingZeros(value, 8);
}

function go2jsBitsTrailingZeros16(value) {
	return go2jsBitsTrailingZeros(value, 16);
}

function go2jsBitsTrailingZeros32(value) {
	return go2jsBitsTrailingZeros(value, 32);
}

function go2jsBitsTrailingZeros64(value) {
	return go2jsBitsTrailingZeros(value, 64);
}

// go2jsBitsLen is the count of bits a number needs before the bit above its
// highest one, and a count of zero needs no bits at all.
function go2jsBitsLen(value) {
	const masked = go2jsBitsMask(value, 64);

	if (masked === 0n) {
		return 0;
	}

	return go2jsBitsLen64(masked);
}

function go2jsBitsLen8(value) {
	const masked = go2jsBitsMask(value, 8);

	return masked === 0n ? 0 : Number(masked.toString(2).length);
}

function go2jsBitsLen16(value) {
	const masked = go2jsBitsMask(value, 16);

	return masked === 0n ? 0 : Number(masked.toString(2).length);
}

function go2jsBitsLen32(value) {
	const masked = go2jsBitsMask(value, 32);

	return masked === 0n ? 0 : Number(masked.toString(2).length);
}

function go2jsBitsLen64(value) {
	const masked = go2jsBitsMask(value, 64);

	return masked === 0n ? 0 : Number(masked.toString(2).length);
}

// go2jsBitsReverse reads the bits of a count from the low end to the high one.
function go2jsBitsReverse(value, width) {
	const masked = go2jsBitsMask(value, width);
	let out = 0n;

	for (let bit = 0; bit < width; bit++) {
		out = (out << 1n) | ((masked >> BigInt(bit)) & 1n);
	}

	return out;
}

function go2jsBitsReverse8(value) {
	return go2jsBitsReverse(value, 8);
}

function go2jsBitsReverse16(value) {
	return go2jsBitsReverse(value, 16);
}

function go2jsBitsReverse32(value) {
	return go2jsBitsReverse(value, 32);
}

function go2jsBitsReverse64(value) {
	return go2jsBitsReverse(value, 64);
}

// go2jsBitsRotateLeft turns the bits of a count around by a number of places,
// which is the same as turning the other way by the places that are left over.
function go2jsBitsRotateLeft(value, count, width) {
	const shift = go2jsBitsSigned(count, width);
	const masked = go2jsBitsMask(value, width);
	const places = ((shift % BigInt(width)) + BigInt(width)) % BigInt(width);

	if (places === 0n) {
		return go2jsBitsNarrow(masked);
	}

	const left = (masked << places) & ((1n << BigInt(width)) - 1n);

	return go2jsBitsNarrow(left | (masked >> (BigInt(width) - places)));
}

// go2jsBitsWide64 turns a number into a whole number that keeps sixty four bits
// exactly, which is the width the counts that overflow a double are read at.
function go2jsBitsWide64(value) {
	if (typeof value === "bigint") {
		return value;
	}

	return BigInt(Math.trunc(Number(value)));
}

function go2jsBitsRotateLeft8(value, count) {
	return go2jsBitsRotateLeft(value, count, 8);
}

function go2jsBitsRotateLeft16(value, count) {
	return go2jsBitsRotateLeft(value, count, 16);
}

function go2jsBitsRotateLeft32(value, count) {
	return go2jsBitsRotateLeft(value, count, 32);
}

function go2jsBitsRotateLeft64(value, count) {
	return go2jsBitsRotateLeft(value, count, 64);
}

function go2jsBitsRotate(value, count) {
	const wide = go2jsBitsWide(value);
	const places = go2jsBitsWide(count);

	// The width of an unsized count is the width of the count itself, which
	// JavaScript has already had to work out to hold it.
	const width = typeof value === "bigint"
		? (value < 0n || value > 4294967295n ? 64 : 32)
		: (value < 0 || value > 4294967295 ? 64 : 32);

	if (places === 0n) {
		return go2jsBitsNarrow(wide);
	}

	const shift = ((places % BigInt(width)) + BigInt(width)) % BigInt(width);
	const masked = go2jsBitsMask(wide, width);
	const left = (masked << shift) & ((1n << BigInt(width)) - 1n);

	return go2jsBitsNarrow(left | (masked >> (BigInt(width) - shift)));
}

function go2jsBitsRotate8(value, count) {
	return go2jsBitsRotateSigned(value, count, 8);
}

function go2jsBitsRotate16(value, count) {
	return go2jsBitsRotateSigned(value, count, 16);
}

function go2jsBitsRotate32(value, count) {
	return go2jsBitsRotateSigned(value, count, 32);
}

function go2jsBitsRotate64(value, count) {
	return go2jsBitsRotateSigned(value, count, 64);
}

function go2jsBitsRotateSigned(value, count, width) {
	const shift = go2jsBitsSigned(count, width);
	const masked = go2jsBitsMask(value, width);
	const places = ((shift % BigInt(width)) + BigInt(width)) % BigInt(width);

	if (places === 0n) {
		return go2jsBitsNarrow(masked);
	}

	const right = (masked >> places) | (masked << (BigInt(width) - places));

	return go2jsBitsNarrow(go2jsBitsMask(right, width));
}

// go2jsBitsAdd adds two counts and the carry into them, and hands back the sum
// and whether it carried out past the width it was asked about.
function go2jsBitsAdd(x, y, carry) {
	const total = go2jsBitsWide64(x) + go2jsBitsWide64(y) + go2jsBitsWide64(carry);

	// A sum that does not fit the width it was asked about is one that carried
	// out of it, which is what the second answer reports. A sum that fits is one
	// inside the width it was asked about, which is where Go looks for a carry.
	const width = 64;
	const limit = 1n << BigInt(width);
	const carried = total < 0n || total >= limit;

	return [go2jsBitsNarrow(go2jsBitsMask(total, width)), carried ? 1 : 0];
}

function go2jsBitsSub(x, y, borrow) {
	const left = go2jsBitsWide(x);
	const right = go2jsBitsWide(y) + go2jsBitsWide(borrow);
	const total = left - right;

	return [go2jsBitsNarrow(total), left < right ? 1 : 0];
}

// go2jsBitsMul multiplies two counts and hands back the high half of the
// product and the low half of it, in that order.
function go2jsBitsMul(x, y) {
	const product = go2jsBitsWide(x) * go2jsBitsWide(y);

	return [
		go2jsBitsNarrow(go2jsBitsMask(product >> 64n, 64)),
		go2jsBitsNarrow(go2jsBitsMask(product, 64)),
	];
}

// go2jsBitsDiv divides one count by another, and the remainder of the division
// is worked out from the two together rather than asked for on its own.
function go2jsBitsDiv(hi, lo, y) {
	const divisor = go2jsBitsWide(y);

	if (divisor === 0n) {
		throw go2jsStdlibError("integer divide by zero");
	}

	const high = go2jsBitsWide(hi);
	const low = go2jsBitsWide(lo);

	// The count being divided is the high half above the low one, so the
	// division goes over both of them at once.
	const numerator = (high << 64n) | low;
	const quotient = numerator / divisor;
	const remainder = numerator % divisor;

	return [go2jsBitsNarrow(go2jsBitsMask(quotient, 64)), go2jsBitsNarrow(remainder)];
}

// go2jsBitsRem is the remainder of a division, worked out from the same three
// numbers the division itself is worked out from.
function go2jsBitsRem(hi, lo, y) {
	const divisor = go2jsBitsWide(y);

	if (divisor === 0n) {
		throw go2jsStdlibError("integer divide by zero");
	}

	const numerator = (go2jsBitsWide(hi) << 64n) | go2jsBitsWide(lo);

	return go2jsBitsNarrow(numerator % divisor);
}

// go2jsBitsUint reads a signed count of sixty four bits as an unsigned one, and
// a signed count of fewer bits as an unsigned count of sixty four.
function go2jsBitsUint(x) {
	if (typeof x === "bigint") {
		return go2jsBitsNarrow(go2jsBitsMask(x, 64));
	}

	const value = Math.trunc(Number(x));

	if (value < 0) {
		return go2jsBitsNarrow(BigInt(value) & 18446744073709551615n);
	}

	return value;
}
`
}

func init() {
	stdlibFuncMaps["math/bits"] = bitsFuncs
	supportedStdlibPackages["math/bits"] = funcSet(bitsFuncs)
}
