package javascript

// The math/big package holds whole numbers of no width at all, which is what
// JavaScript calls a BigInt, so the value of an Int is held as a BigInt and
// nothing else. The holder is a class of its own because every method of an Int
// is written on a pointer to it and hands back the very pointer it was called
// on, the way Go does: a program holds pointers to Ints rather than Ints, and
// the pointer is what the methods are written on.
var bigFuncs = map[string]string{
	"NewInt": "go2jsBigNewInt",
}

func mathBigRuntimeSource() string {
	return `
// go2jsBigInt is the holder of a whole number of the math/big package. Go says an
// Int is a value of its own and never a name for one, so a program that copies a
// pointer to an Int has two pointers to the one value, which is what two holders
// standing for one number cannot say; a pointer is therefore held as the holder
// itself and a copy of a pointer is a copy of that holder.
function go2jsBigInt() {
	// The holder says it holds a whole number of its own, which is what tells the
	// writing of one to count the number rather than to name the holder, the way
	// a type of Go that answers for itself is written rather than described.
	this.__go2js_whole_number = true;
	this.value = 0n;
}

// A whole number of no width is the one number rather than a copy of it, since
// every method of an Int is written on a pointer to it and the holder is that
// pointer, so a copy of a pointer is a copy of the pointer and not a second Int.
go2jsBigInt.prototype.__go2js_shared = true;

// go2jsBigIntDigits answers the value of a digit written as a letter, which is
// how the digits above nine are written from ten up, and -1 for anything that is
// no digit at all.
function go2jsBigIntDigit(character) {
	if (character >= "0" && character <= "9") {
		return character.charCodeAt(0) - 48;
	}

	const lowered = character.toLowerCase();
	if (lowered >= "a" && lowered <= "z") {
		return lowered.charCodeAt(0) - 97 + 10;
	}

	return -1;
}

// go2jsBigIntValue is the number inside a value of the package, however it was
// handed over. A value that has been through an interface is held in a box, and
// the box is a wrapper around the number rather than the number itself, and a
// value of nothing is a number of no value at all, which is as good as zero.
function go2jsBigIntValue(value) {
	if (value !== null && value !== undefined && value.__go2js_interface === true) {
		value = value.value;
	}

	if (value === null || value === undefined) {
		return 0n;
	}

	if (value !== null && value !== undefined && typeof value === "object" &&
		typeof value.value === "bigint") {
		return value.value;
	}

	return BigInt(value);
}

// go2jsBigIntOf is the number a pointer to an Int holds. A method on a pointer to
// nothing is a fault in Go rather than an answer, which is what reaching for the
// number behind nothing is.
function go2jsBigIntOf(receiver) {
	if (receiver === null || receiver === undefined) {
		throw new TypeError(go2jsRuntimeErrorPrefix + "invalid memory address or nil pointer dereference");
	}

	return go2jsBigIntValue(receiver);
}

// go2jsBigIntDivisor is the number a division is by, and dividing by zero is a
// fault in Go rather than an answer.
function go2jsBigIntDivisor(divisor) {
	const value = go2jsBigIntOf(divisor);

	if (value === 0n) {
		throw new go2jsNativeRangeError(go2jsRuntimeErrorPrefix + "division by zero");
	}

	return value;
}

// go2jsBigIntEuclidParts is what a division leaves three numbers of: what was
// divided, the divisor it was divided by, and the remainder left over, which is
// never negative and never as large as the divisor whatever the sign of either.
// JavaScript leaves the remainder with the sign of what was divided, so it is
// taken in the width of the divisor and brought up out of the negative side.
function go2jsBigIntEuclidParts(x, y) {
	const dividend = go2jsBigIntOf(x);
	const divisor = go2jsBigIntDivisor(y);
	const width = divisor < 0n ? -divisor : divisor;
	const remainder = ((dividend % width) + width) % width;

	return [dividend, divisor, remainder];
}

// go2jsBigIntLowBits is what a machine number of a width answers for a number
// wider than that width: the bits it holds at the bottom of the number, brought
// round about the ends of the width where they run past the top of it.
function go2jsBigIntLowBits(value, half, span) {
	const held = ((value % span) + span) % span;

	return held > half ? held - span : held;
}

// go2jsBigIntAbs is a number with nothing under it, which is a number below
// nothing standing the other way up rather than the number of no value at all.
function go2jsBigIntAbs(value) {
	return value < 0n ? -value : value;
}

// go2jsBigIntMod is a number counted off another as many times as it takes, which
// is the count left over in the width of the other however either is written, and
// where the other is no value at all there is nothing to count off and the count
// is the number itself.
function go2jsBigIntMod(value, modulus) {
	if (modulus === 0n) {
		return value;
	}

	return go2jsBigIntEuclidParts(value, {value: modulus})[2];
}

// go2jsBigIntGiven is whether a pointer to an Int was handed over at all, which
// is what tells a pointer to nothing from a pointer to a number of no value: the
// first is nothing to write into and the second is written like any other.
function go2jsBigIntGiven(value) {
	return value !== null && value !== undefined;
}

// go2jsBigIntInto is a number written into the Int a pointer points at, and a
// pointer to nothing is a fault in Go rather than an answer, the same as reading
// through one is.
function go2jsBigIntInto(pointer, held) {
	if (!go2jsBigIntGiven(pointer)) {
		throw new TypeError(go2jsRuntimeErrorPrefix + "invalid memory address or nil pointer dereference");
	}

	pointer.value = held;

	return pointer;
}

// go2jsBigIntBezout is a number that divides two without a remainder together with
// the count of the first of them that says so, both taken from the sizes of the
// two, since a count of one number belongs to it however it is written: a*x + b*y
// comes to the divisor for some y, and the count of the first is what is left of
// the divisor when the first multiplied by it is taken away, divided by the second.
function go2jsBigIntBezout(a, b) {
	const left = go2jsBigIntAbs(a);
	const right = go2jsBigIntAbs(b);
	let held = left;
	let other = right;
	let count = 1n;
	let step = 0n;

	// The two are brought down to nothing by the division that leaves nothing
	// behind, keeping beside the count how many times the first was taken in to
	// reach what is held.
	while (other !== 0n) {
		const quotient = held / other;

		[held, other] = [other, held - quotient * other];
		[count, step] = [step, count - quotient * step];
	}

	return [held, count];
}

// go2jsBigIntCount is a count of places or bits, which Go counts in a machine
// word and hands over to the package as a number, and JavaScript has no count
// wider than a double, so the count is taken as the number it stands for.
function go2jsBigIntCount(value) {
	const held = typeof value === "bigint" ? value : BigInt(Math.trunc(Number(value)));

	// A count below nothing is a fault in Go, since there is no place below the
	// first one to count from and no bit below the bottom one to count up to.
	if (held < 0n) {
		throw new go2jsNativeRangeError("negative count");
	}

	return held;
}

// go2jsBigIntSign is -1, 0 or 1 as a number is below, equal to or above nothing.
function go2jsBigIntSign(value) {
	return value < 0n ? -1 : (value > 0n ? 1 : 0);
}

// go2jsBigIntWidth is how wide a number has to be written for two of them to hold
// between them every bit either has, and a width is a number of bits rather than
// a number, so the count is a plain one.
function go2jsBigIntWidth(x, y) {
	const left = go2jsBigIntValue(x) < 0n ? -go2jsBigIntValue(x) : go2jsBigIntValue(x);
	const right = go2jsBigIntValue(y) < 0n ? -go2jsBigIntValue(y) : go2jsBigIntValue(y);

	return Number(left.toString(2).length > right.toString(2).length
		? left.toString(2).length
		: right.toString(2).length) + 1;
}

// go2jsBigIntBits is two numbers put together bit by bit. A number below nothing
// has no bits of its own in JavaScript, but Go writes a number below nothing in
// twos complement as far as the width it is written to, and both numbers are
// therefore written at a width that holds every bit of either before they are put
// together, so that a bit set in a number below nothing is set in the result the
// way Go sets it.
function go2jsBigIntBits(x, y, width, combine) {
	const held = BigInt.asUintN(width, go2jsBigIntValue(x));

	return BigInt.asIntN(width, combine(held, BigInt.asUintN(width, go2jsBigIntValue(y))));
}

go2jsBigInt.prototype.Set = function(value) {
	this.value = go2jsBigIntValue(value);

	return this;
};

go2jsBigInt.prototype.SetInt64 = function(value) {
	this.value = go2jsBigIntValue(value);

	return this;
};

go2jsBigInt.prototype.SetUint64 = function(value) {
	const held = go2jsBigIntValue(value);

	// A count of no width has no negative side and no upper one beyond the width
	// it is written to, so a number outside those is not a count at all.
	if (held < 0n || held > 18446744073709551615n) {
		throw new go2jsNativeRangeError("argument to SetUint64 is not a uint64");
	}

	this.value = held;

	return this;
};

go2jsBigInt.prototype.Add = function(x, y) {
	this.value = go2jsBigIntOf(x) + go2jsBigIntOf(y);

	return this;
};

go2jsBigInt.prototype.Sub = function(x, y) {
	this.value = go2jsBigIntOf(x) - go2jsBigIntOf(y);

	return this;
};

go2jsBigInt.prototype.Mul = function(x, y) {
	this.value = go2jsBigIntOf(x) * go2jsBigIntOf(y);

	return this;
};

go2jsBigInt.prototype.Quo = function(x, y) {
	// Quo is the quotient as it falls, cut off at the point and no further, so
	// the remainder is left holding the sign of what was divided.
	this.value = go2jsBigIntOf(x) / go2jsBigIntDivisor(y);

	return this;
};

go2jsBigInt.prototype.Rem = function(x, y) {
	this.value = go2jsBigIntOf(x) % go2jsBigIntDivisor(y);

	return this;
};

go2jsBigInt.prototype.Div = function(x, y) {
	// Div is a division that leaves nothing behind but a remainder no smaller than
	// nothing and no larger than the divisor however the divisor is written, so it
	// counts down to the divisor rather than off at the point: the remainder is
	// what is left of the dividend in the width of the divisor and the quotient is
	// whatever takes the rest of it away.
	const [dividend, divisor, remainder] = go2jsBigIntEuclidParts(x, y);

	this.value = (dividend - remainder) / divisor;

	return this;
};

go2jsBigInt.prototype.Mod = function(x, y) {
	// Mod is what such a division leaves over, so it too is never negative and
	// never as large as the divisor, whatever the sign of what was divided.
	const [, , remainder] = go2jsBigIntEuclidParts(x, y);

	this.value = remainder;

	return this;
};

go2jsBigInt.prototype.DivMod = function(x, y, modulus) {
	// DivMod is a division that leaves nothing behind written twice: the quotient
	// into the number it is called on and the remainder into the one given.
	const [dividend, divisor, remainder] = go2jsBigIntEuclidParts(x, y);

	this.value = (dividend - remainder) / divisor;

	go2jsBigIntInto(modulus, remainder);

	return [this, modulus];
};

go2jsBigInt.prototype.QuoRem = function(x, y, remainder) {
	// QuoRem is the division that falls at the point, written twice, where the
	// remainder is left holding the sign of what was divided.
	const dividend = go2jsBigIntOf(x);
	const divisor = go2jsBigIntDivisor(y);

	this.value = dividend / divisor;

	go2jsBigIntInto(remainder, dividend % divisor);

	return [this, remainder];
};

go2jsBigInt.prototype.Exp = function(x, count, modulus) {
	// Exp is the number raised to a power, and where a modulus is given it is the
	// number left after raising it as many times as the power says, counted off
	// the modulus each time so that nothing wider than the modulus is ever held.
	// A modulus of no value at all is no modulus, and a power of no value or less
	// leaves the number of one, since anything raised to nothing is one.
	const base = go2jsBigIntOf(x);
	const power = go2jsBigIntValue(count);
	const held = go2jsBigIntValue(modulus);
	let width = held < 0n ? -held : held;

	if (width < 1n) {
		width = 0n;
	}

	if (power <= 0n) {
		this.value = width === 0n ? 1n : 1n % width;

		return this;
	}

	// The power is walked from its top bit down, doubling what is held and
	// multiplying it by the base wherever the bit stands for one, since a power
	// of no width may be far wider than anything counted in one go.
	let held2 = width === 0n ? 1n : 1n % width;
	let square = width === 0n ? base : base % width;
	let left = power;

	while (left > 0n) {
		if ((left & 1n) === 1n) {
			held2 = width === 0n ? held2 * square : (held2 * square) % width;
		}

		square = width === 0n ? square * square : (square * square) % width;
		left >>= 1n;
	}

	// What is left of a power counted off a modulus is counted off in the width of
	// that modulus, so a power below nothing comes out above nothing, which is
	// where Go puts it as well.
	this.value = width === 0n ? held2 : (held2 < 0n ? held2 + width : held2);

	return this;
};

go2jsBigInt.prototype.GCD = function(x, y, a, b) {
	// GCD is the largest number that divides both without a remainder, and it is
	// written as no value at all or above however the two are written. x and y are
	// written so that the divisor is a*x + b*y, which is what makes a division of
	// a by the divisor come out with nothing in it, and where one of the two given
	// is no value at all the other answers for itself beside a count of one or
	// nothing, since nothing is a divisor of everything.
	const left = go2jsBigIntOf(a);
	const right = go2jsBigIntOf(b);

	if (go2jsBigIntAbs(left) === 0n || go2jsBigIntAbs(right) === 0n) {
		this.value = go2jsBigIntAbs(left) === 0n ? go2jsBigIntAbs(right) : go2jsBigIntAbs(left);

		if (go2jsBigIntGiven(x)) {
			go2jsBigIntInto(x, go2jsBigIntAbs(left) === 0n ? 0n : BigInt(go2jsBigIntSign(left)));
		}

		if (go2jsBigIntGiven(y)) {
			go2jsBigIntInto(y, go2jsBigIntAbs(right) === 0n ? 0n : BigInt(go2jsBigIntSign(right)));
		}

		return this;
	}

	// The two counts are found from the sizes of the two, since the count of one
	// belongs to one whichever way it is written, and the count beside the other is
	// then whatever takes the divisor off the whole of it.
	const [divisor, count] = go2jsBigIntBezout(left, right);

	this.value = divisor;

	if (go2jsBigIntGiven(x)) {
		go2jsBigIntInto(x, left < 0n ? -count : count);
	}

	if (go2jsBigIntGiven(y)) {
		const written = left < 0n ? -count : count;
		const rest = divisor - left * written;

		go2jsBigIntInto(y, rest / right);
	}

	return this;
};

go2jsBigInt.prototype.ModInverse = function(value, modulus) {
	// ModInverse is the count that undoes the number, which multiplies it into one
	// left over after being counted off the modulus as many times as it takes. A
	// modulus is counted in its own width, whatever its sign, and a number below
	// nothing is counted off the modulus first, so that what is counted off is the
	// number as it stands among the numbers of that width. Two numbers that share
	// something have no such count at all, and Go hands back nothing rather than
	// a count, leaving the number it was called on as it was.
	const held = go2jsBigIntOf(value);
	const written = go2jsBigIntDivisor(modulus);
	const width = written < 0n ? -written : written;
	const number = held < 0n ? go2jsBigIntMod(held, width) : held;
	const [divisor, count] = go2jsBigIntBezout(number, width);

	if (divisor !== 1n) {
		return null;
	}

	// The count comes out of the division wherever it lands, and what matters is
	// the count left over in the width of the modulus, so a count below nothing is
	// brought up by the width of the modulus once, which is as much as it takes.
	this.value = count < 0n ? count + width : count;

	return this;
};

go2jsBigInt.prototype.Neg = function(x) {
	this.value = -go2jsBigIntOf(x);

	return this;
};

go2jsBigInt.prototype.Abs = function(x) {
	const value = go2jsBigIntOf(x);

	this.value = value < 0n ? -value : value;

	return this;
};

go2jsBigInt.prototype.Not = function(x) {
	// Not is every bit turned over, which for a number written in twos complement
	// is the number one below the number itself, and JavaScript writes numbers
	// below nothing the same way, so the two agree without a width.
	this.value = ~go2jsBigIntOf(x);

	return this;
};

go2jsBigInt.prototype.And = function(x, y) {
	const width = go2jsBigIntWidth(x, y);

	this.value = go2jsBigIntBits(x, y, width, (left, right) => left & right);

	return this;
};

go2jsBigInt.prototype.Or = function(x, y) {
	const width = go2jsBigIntWidth(x, y);

	this.value = go2jsBigIntBits(x, y, width, (left, right) => left | right);

	return this;
};

go2jsBigInt.prototype.Xor = function(x, y) {
	const width = go2jsBigIntWidth(x, y);

	this.value = go2jsBigIntBits(x, y, width, (left, right) => left ^ right);

	return this;
};

go2jsBigInt.prototype.AndNot = function(x, y) {
	// AndNot is the bits of the first that the second does not have set, which
	// is the first with the bits of the second turned over put through it.
	const width = go2jsBigIntWidth(x, y);

	this.value = go2jsBigIntBits(x, y, width, (left, right) => left & ~right);

	return this;
};

go2jsBigInt.prototype.Lsh = function(x, count) {
	// Lsh is the number with that many bits of nothing standing on the right of
	// it, which is the number multiplied by two as many times.
	this.value = go2jsBigIntOf(x) << go2jsBigIntCount(count);

	return this;
};

go2jsBigInt.prototype.Rsh = function(x, count) {
	// Rsh is the number with that many bits of it falling off the right, and the
	// bits that fall are made up by the sign repeated, so a number below nothing
	// divided by two as many times still counts down rather than off at nothing.
	this.value = go2jsBigIntOf(x) >> go2jsBigIntCount(count);

	return this;
};

go2jsBigInt.prototype.Bit = function(index) {
	// Bit is one bit of the number at a place counted from the bottom, and a
	// number below nothing is written with as many bits above it as it has below,
	// so every bit of it above its width stands as it stands in twos complement,
	// which is what a shift of a number below nothing does in JavaScript as well.
	return Number((this.value >> go2jsBigIntCount(index)) & 1n);
};

go2jsBigInt.prototype.TrailingZeroBits = function() {
	// The trailing bits of nothing are no bits of the number at all, so a number
	// of no value has as many of them as a machine can count in one go.
	const value = this.value < 0n ? -this.value : this.value;

	// A number of no value at all has no bits of its own at all, so none of them
	// stand at the bottom of it and none is to be counted off the bottom.
	if (value === 0n) {
		return 0;
	}

	let held = 0n;

	while (((value >> held) & 1n) === 0n) {
		held += 1n;
	}

	return Number(held);
};

go2jsBigInt.prototype.Cmp = function(x) {
	// Cmp says which is the greater, so the number held is set against the one
	// given rather than the other way round.
	return go2jsBigIntSign(this.value - go2jsBigIntOf(x));
};

go2jsBigInt.prototype.CmpAbs = function(x) {
	// CmpAbs sets the two numbers by their size rather than by where they stand,
	// so two numbers the same size apart are neither above one another however
	// they are written.
	const held = go2jsBigIntAbs(this.value);
	const other = go2jsBigIntAbs(go2jsBigIntOf(x));

	return go2jsBigIntSign(held - other);
};

go2jsBigInt.prototype.Sign = function() {
	return go2jsBigIntSign(this.value);
};

go2jsBigInt.prototype.String = function() {
	return this.value.toString();
};

go2jsBigInt.prototype.Text = function(base) {
	const radix = Number(base);

	// A base is a place to count in, and a place of one, of none or of more than
	// the thirty six digits there are letters for is no place to count in at all.
	if (!Number.isInteger(radix) || radix < 2 || radix > 36) {
		throw new go2jsNativeRangeError("invalid base");
	}

	return this.value.toString(radix);
};

go2jsBigInt.prototype.SetString = function(text, base) {
	const value = go2jsBigIntParse(String(text), Number(base));

	// Text that is no number at all hands back nothing at all, because there is no
	// number in it to point to and what would have been read into cannot be told
	// from one number to another, and Go says what is left in a number that was
	// not read is not to be relied on.
	if (value === null) {
		return [null, false];
	}

	this.value = value;

	return [this, true];
};

go2jsBigInt.prototype.Int64 = function() {
	// A machine number of sixty four bits is read off the bottom of the number and
	// wraps about the ends of that width where the number is wider than it, which
	// is what the machine holds of such a number, and the sign of the number is
	// put back afterwards.
	const value = go2jsBigIntLowBits(this.value, 9223372036854775807n, 18446744073709551616n);

	return go2jsNarrow(value);
};

go2jsBigInt.prototype.Uint64 = function() {
	// A count of no width is read off the bottom of the number with the sign left
	// off, so a number below nothing answers with what is under it and a number
	// wider than the sixty four bits answers with what is left at the bottom.
	const value = this.value < 0n ? -this.value : this.value;

	return go2jsNarrow(value & 18446744073709551615n);
};

go2jsBigInt.prototype.BitLen = function() {
	// The bits a number holds are counted from the top down rather than written
	// out, because writing a number wider than a double as text costs a character
	// for every bit of it and a shift of thirty two at a time costs none for the
	// bits below where it stops.
	const value = this.value < 0n ? -this.value - 1n : this.value;
	let held = 0n;

	while (value >> BigInt(held) > 0xffffffffn) {
		held += 32n;
	}

	return Number(held) + 32 - Math.clz32(Number(value >> held));
};

go2jsBigInt.prototype.IsInt64 = function() {
	return this.value >= -9223372036854775808n && this.value <= 9223372036854775807n;
};

go2jsBigInt.prototype.IsUint64 = function() {
	return this.value >= 0n && this.value <= 18446744073709551615n;
};

go2jsBigInt.prototype.Bytes = function() {
	// Bytes is the number written as bytes from the heaviest to the lightest,
	// which is how a number travels between machines and out of a program, and it
	// is the number under nothing rather than the number itself, so a number below
	// nothing is written as what is under it.
	const value = go2jsBigIntAbs(this.value);

	if (value === 0n) {
		return [];
	}

	const held = value.toString(16);
	const written = held.length % 2 === 0 ? held : "0" + held;
	const answered = [];

	for (let i = 0; i < written.length; i += 2) {
		answered.push(parseInt(written.slice(i, i + 2), 16));
	}

	return answered;
};

go2jsBigInt.prototype.SetBytes = function(bytes) {
	// SetBytes is the number read back out of bytes written from the heaviest to
	// the lightest, with nothing above nothing counted as a byte of no value.
	const source = bytes === null || bytes === undefined ? [] : Array.from(bytes);
	let held = 0n;

	for (const item of source) {
		const byte = BigInt(item);

		if (byte < 0n || byte > 255n) {
			throw new go2jsNativeRangeError("byte out of range in SetBytes");
		}

		held = (held << 8n) | byte;
	}

	this.value = held;

	return this;
};

go2jsBigInt.prototype.Float64 = function() {
	// Float64 is the double standing nearest the number, with the way it stands
	// said as well: below where the double falls short of it, above where it goes
	// past it, and exact where it lands on the number itself.
	const held = Number(this.value);

	if (!Number.isFinite(held)) {
		return [held, "Below"];
	}

	const below = BigInt(held);
	const above = held < 0 ? below - 1n : below + 1n;

	if (below === this.value) {
		return [held, "Exact"];
	}

	return [held, below > this.value ? "Above" : "Below"];
};

// go2jsBigIntParse reads a number written as text in a given base, and answers
// nothing at all rather than a number when the text is not one: a base of zero is
// read off the front of the text rather than named, and every base has to be one
// there are digits for.
function go2jsBigIntParse(text, base) {
	// A base is a place to count in, and there are digits for the places from two
	// to the thirty six and no letters for any other, so a base outside those is a
	// fault in Go rather than a number that is read differently.
	if (base !== 0 && (base < 2 || base > 36)) {
		throw new go2jsNativeRangeError("invalid number base " + base);
	}

	if (!Number.isInteger(base)) {
		return null;
	}

	// A sign is a part of the number and not of the base it is written in, so it
	// is taken off before the rest is read and put back afterwards.
	let sign = 1n;
	let rest = text;

	if (rest.startsWith("+") || rest.startsWith("-")) {
		sign = rest.startsWith("-") ? -1n : 1n;
		rest = rest.slice(1);
	}

	let value = null;

	if (base === 0) {
		const lowered = rest.toLowerCase();

		if (lowered.startsWith("0b")) {
			value = go2jsBigIntDigits(go2jsBigIntAfterPrefix(lowered.slice(2)), 2, true);
		} else if (lowered.startsWith("0o")) {
			value = go2jsBigIntDigits(go2jsBigIntAfterPrefix(lowered.slice(2)), 8, true);
		} else if (lowered.startsWith("0x")) {
			value = go2jsBigIntDigits(go2jsBigIntAfterPrefix(lowered.slice(2)), 16, true);
		} else if (lowered.length > 1 && lowered.startsWith("0")) {
			value = go2jsBigIntDigits(go2jsBigIntAfterPrefix(lowered.slice(1)), 8, true);
		} else {
			// No prefix was written, so there is no base standing between an
			// underscore and the first digit, and one standing first is no number.
			value = go2jsBigIntDigits(lowered, 10, true);
		}
	} else {
		value = go2jsBigIntDigits(rest, base, false);
	}

	return value === null ? null : sign * value;
}

// go2jsBigIntAfterPrefix takes the digits of a number whose base was read off its
// front, where an underscore may stand between the base and the first digit, the
// way Go writes a number apart from the base it is in.
function go2jsBigIntAfterPrefix(text) {
	return text.startsWith("_") ? text.slice(1) : text;
}

// go2jsBigIntDigits reads the digits of a number in a base that was settled on
// already, and answers nothing at all rather than a number when a digit is not
// one of the base or an underscore stands where Go would not have it stand.
function go2jsBigIntDigits(text, base, underscoresAllowed) {
	if (base < 2 || base > 36 || text === "") {
		return null;
	}

	let written = "";

	for (let index = 0; index < text.length; index++) {
		const character = text[index];

		if (character === "_") {
			// An underscore stands for nothing at all, and is written between two
			// digits and only where the base was read off the front of the number
			// rather than named outright, so one at either end of the digits or two
			// of them running together is not a number.
			if (!underscoresAllowed || index === 0 || index === text.length - 1 ||
				text[index - 1] === "_") {
				return null;
			}

			continue;
		}

		const digit = go2jsBigIntDigit(character);

		if (digit < 0 || digit >= base) {
			return null;
		}

		written += character;
	}

	if (written === "") {
		return null;
	}

	const radix = BigInt(base);
	let value = 0n;

	for (let index = 0; index < written.length; index++) {
		value = value * radix + BigInt(go2jsBigIntDigit(written[index]));
	}

	return value;
}

// go2jsBigNewInt is a pointer to a whole number of no width, which Go names after
// the machine number it was given, which is a number it can hold.
function go2jsBigNewInt(value) {
	const held = new go2jsBigInt();

	held.value = go2jsBigIntValue(value);

	return held;
}
`
}

func init() {
	stdlibFuncMaps["math/big"] = bigFuncs
	supportedStdlibPackages["math/big"] = funcSet(bigFuncs)
	// A type of a package is found by the name of the package it belongs to and
	// the name of the type itself, which for this package is big.Int rather than
	// anything written out of the path it is imported by.
	packageTypes["big.Int"] = "go2jsBigInt"
	packageTypeNew["big.Int"] = true
}
