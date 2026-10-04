package javascript

import "strings"

// sha512Funcs are the functions of crypto/sha512. Go has one function for the
// family and a constructor for each of the four sizes, so all of them are here and
// the size a program asks for by name is the size it says.
var sha512Funcs = map[string]string{
	"Sum512":     "go2jsSHA512Sum512",
	"Sum384":     "go2jsSHA512Sum384",
	"Sum512_224": "go2jsSHA512Sum224",
	"Sum512_256": "go2jsSHA512Sum256",
	"New":        "go2jsSHA512New",
	"New384":     "go2jsSHA384New",
	"New512_224": "go2jsSHA512224New",
	"New512_256": "go2jsSHA512256New",
}

var sha512Constants = map[string]string{
	"Size":      "64",
	"Size224":   "28",
	"Size256":   "32",
	"Size384":   "48",
	"BlockSize": "128",
}

// crc32Funcs are the functions of hash/crc32, which are the checksum and the
// running hash of the two polynomials the standard library holds tables for.
var crc32Funcs = map[string]string{
	"Checksum":     "go2jsCRC32Checksum",
	"ChecksumIEEE": "go2jsCRC32ChecksumIEEE",
	"New":          "go2jsHashCRC32",
	"NewIEEE":      "go2jsHashCRC32IEEE",
	"Update":       "go2jsCRC32Update",
	"MakeTable":    "go2jsCRC32MakeTable",
}

var crc64Funcs = map[string]string{
	"Checksum":  "go2jsCRC64Checksum",
	"New":       "go2jsHashCRC64",
	"Update":    "go2jsCRC64Update",
	"MakeTable": "go2jsCRC64MakeTable",
}

// crcSizes are the widths of the checksums of both kinds of CRC, which a program
// reads from the package rather than from the number of bytes it wrote out.
var crc32Constants = map[string]string{"Size": "4"}
var crc64Constants = map[string]string{"Size": "8"}

// fnvFuncs are the functions of hash/fnv, which are the constructors of the family:
// four widths times two orders, each told apart by whether the multiply comes
// before the exclusive or or after it.
var fnvFuncs = map[string]string{
	"New32":   "go2jsFNVNew32",
	"New32a":  "go2jsFNVNew32a",
	"New64":   "go2jsFNVNew64",
	"New64a":  "go2jsFNVNew64a",
	"New128":  "go2jsFNVNew128",
	"New128a": "go2jsFNVNew128a",
}

// adler32 has no package of its own: it is reached through hash/adler32, whose one
// function answers with the checksum of a message.
var adler32Funcs = map[string]string{
	"Checksum": "go2jsAdler32Checksum",
	"New":      "go2jsHashAdler32",
}

func init() {
	stdlibFuncMaps["crypto/sha512"] = sha512Funcs
	stdlibFuncMaps["hash/crc32"] = crc32Funcs
	stdlibFuncMaps["hash/crc64"] = crc64Funcs
	stdlibFuncMaps["hash/fnv"] = fnvFuncs
	stdlibFuncMaps["hash/adler32"] = adler32Funcs

	for path, alias := range map[string]string{
		"crypto/sha512": "sha512",
		"hash/crc32":    "crc32",
		"hash/crc64":    "crc64",
		"hash/fnv":      "fnv",
		"hash/adler32":  "adler32",
	} {
		stdlibPkgAliases[path] = []string{alias}
		supportedStdlibPackages[path] = funcSet(stdlibFuncMaps[path])
	}

	supportedStdlibPackages["crypto/sha512"] = funcSetWith(sha512Constants, sha512Funcs)

	// A constant is looked up by the name the program imported the package under,
	// which is the last part of the path most of the time, so the constants of
	// these packages are registered under both names.
	for name, value := range sha512Constants {
		for _, key := range []string{"crypto/sha512." + name, "sha512." + name} {
			packageConstants[key] = value
			packageVarValues[key] = value
			packageVarTypes[key] = "int"
		}
	}

	// A polynomial is a number and a table is a value, and a program passes the
	// one to MakeTable and the other to New, so they are told apart here.
	polynomials := map[string]string{
		"hash/crc32.Size":       "4",
		"hash/crc64.Size":       "8",
		"hash/crc32.IEEE":       "0xedb88320",
		"hash/crc32.Castagnoli": "0x82f63b78",
		"hash/crc32.Koopman":    "0xeb31d82e",
		"hash/crc64.ISO":        "0xd800000000000000n",
		"hash/crc64.ECMA":       "0xc96c5795d7870f42n",
	}

	for key, value := range polynomials {
		short := key[strings.Index(key, "/")+1:]

		for _, name := range []string{key, short} {
			packageConstants[name] = value
			packageVarValues[name] = value
			packageVarTypes[name] = crcVarType(name)
		}
	}

	// The tables of the standard library are values rather than constants, since a
	// program hands one to New or to Checksum as a table and not as a number.
	tables := map[string]string{
		"hash/crc32.IEEETable": "go2jsCRC32IEEETableValue",
	}

	for key, value := range tables {
		short := key[strings.Index(key, "/")+1:]

		for _, name := range []string{key, short} {
			packageVarValues[name] = value
			packageVarTypes[name] = "*" + strings.Split(short, ".")[0] + ".Table"
		}
	}

	for _, path := range []string{"hash/crc32", "hash/crc64"} {
		for _, name := range []string{"IEEE", "Castagnoli", "Koopman", "ISO", "ECMA", "IEEETable", "Size"} {
			key := path + "." + name

			if _, ok := packageConstants[key]; !ok {
				continue
			}

			supportedStdlibPackages[path][name] = true
		}
	}
}

// funcSetWith is the set of names a package answers to, which is what it has as
// functions together with what it declares as constants and as variables.
func funcSetWith(constants, functions map[string]string) map[string]bool {
	set := funcSet(functions)

	for name := range constants {
		set[name] = true
	}

	return set
}

// crcVarType is the type of a value of one of the two checksum packages: the width
// of a checksum is a plain number, a polynomial of the 32-bit kind is a uint32 and
// one of the 64-bit kind is a uint64, which is a number JavaScript cannot hold.
func crcVarType(name string) string {
	switch {
	case strings.HasSuffix(name, ".Size"):
		return "int"
	case strings.HasPrefix(name, "hash/crc64."), strings.HasPrefix(name, "crc64."):
		return "uint64"
	default:
		return "uint32"
	}
}

func cryptoHashRuntimeSource() string {
	return `
// go2jsBigUint is a 64-bit number as a BigInt, since a JavaScript number holds 53
// bits exactly and every hash of 64 bits and every hash value read in as one is
// wider than that.
function go2jsBigUint(value) {
	return BigInt(value);
}

const go2jsMask64 = 0xffffffffffffffffn;
const go2jsMask32 = 0xffffffffn;

// go2jsBigBytes is a BigInt as the bytes of a number Go writes out, little end
// first, since that is the order a number of the host is held in.
function go2jsBigBytes(value, width) {
	const bytes = [];
	let rest = BigInt(value) & go2jsMask64;

	for (let index = 0; index < width; index++) {
		bytes.push(Number(rest & 0xffn));
		rest >>= 8n;
	}

	return bytes;
}

// go2jsBigFromBytes reads a number out of bytes, little end first, which is how a
// checksum that was written out as bytes is read back into the number it stands
// for.
function go2jsBigFromBytes(bytes) {
	let value = 0n;

	for (let index = bytes.length - 1; index >= 0; index--) {
		value = (value << 8n) | BigInt(Number(bytes[index]) & 255);
	}

	return value;
}

// go2jsSHA512K is the round constants of SHA-512, which are the first sixty-four
// bits of the fractional parts of the cube roots of the first eighty primes.
const go2jsSHA512K = [
	"428a2f98d728ae22", "7137449123ef65cd", "b5c0fbcfec4d3b2f", "e9b5dba58189dbbc",
	"3956c25bf348b538", "59f111f1b605d019", "923f82a4af194f9b", "ab1c5ed5da6d8118",
	"d807aa98a3030242", "12835b0145706fbe", "243185be4ee4b28c", "550c7dc3d5ffb4e2",
	"72be5d74f27b896f", "80deb1fe3b1696b1", "9bdc06a725c71235", "c19bf174cf692694",
	"e49b69c19ef14ad2", "efbe4786384f25e3", "0fc19dc68b8cd5b5", "240ca1cc77ac9c65",
	"2de92c6f592b0275", "4a7484aa6ea6e483", "5cb0a9dcbd41fbd4", "76f988da831153b5",
	"983e5152ee66dfab", "a831c66d2db43210", "b00327c898fb213f", "bf597fc7beef0ee4",
	"c6e00bf33da88fc2", "d5a79147930aa725", "06ca6351e003826f", "142929670a0e6e70",
	"27b70a8546d22ffc", "2e1b21385c26c926", "4d2c6dfc5ac42aed", "53380d139d95b3df",
	"650a73548baf63de", "766a0abb3c77b2a8", "81c2c92e47edaee6", "92722c851482353b",
	"a2bfe8a14cf10364", "a81a664bbc423001", "c24b8b70d0f89791", "c76c51a30654be30",
	"d192e819d6ef5218", "d69906245565a910", "f40e35855771202a", "106aa07032bbd1b8",
	"19a4c116b8d2d0c8", "1e376c085141ab53", "2748774cdf8eeb99", "34b0bcb5e19b48a8",
	"391c0cb3c5c95a63", "4ed8aa4ae3418acb", "5b9cca4f7763e373", "682e6ff3d6b2b8a3",
	"748f82ee5defb2fc", "78a5636f43172f60", "84c87814a1f0ab72", "8cc702081a6439ec",
	"90befffa23631e28", "a4506cebde82bde9", "bef9a3f7b2c67915", "c67178f2e372532b",
	"ca273eceea26619c", "d186b8c721c0c207", "eada7dd6cde0eb1e", "f57d4f7fee6ed178",
	"06f067aa72176fba", "0a637dc5a2c898a6", "113f9804bef90dae", "1b710b35131c471b",
	"28db77f523047d84", "32caab7b40c72493", "3c9ebe0a15c9bebc", "431d67c49c100d4c",
	"4cc5d4becb3e42b6", "597f299cfc657e2a", "5fcb6fab3ad6faec", "6c44198c4a475817"
];

// go2jsSHA512IV is the state SHA-512 starts from: the first sixty-four bits of the
// fractional parts of the square roots of the first eight primes.
const go2jsSHA512IV = [
	"6a09e667f3bcc908", "bb67ae8584caa73b", "3c6ef372fe94f82b", "a54ff53a5f1d36f1",
	"510e527fade682d1", "9b05688c2b3e6c1f", "1f83d9abfb41bd6b", "5be0cd19137e2179"
];

// go2jsSHA384IV is the state SHA-384 starts from, which is another eight words of
// the same fractions as SHA-512's, taken one further along: a hash of a narrower
// width starts from its own values and not from a wider hash cut down.
const go2jsSHA384IV = [
	"cbbb9d5dc1059ed8", "629a292a367cd507", "9159015a3070dd17", "152fecd8f70e5939",
	"67332667ffc00b31", "8eb44a8768581511", "db0c2e0d64f98fa7", "47b5481dbefa4fa4"
];

// go2jsSHA512224IV and go2jsSHA512256IV start the two truncations of SHA-512 that
// are 224 and 256 bits wide, each of which a standard says is a hash of its own
// rather than a wider one with bytes cut off the end of it.
const go2jsSHA512224IV = [
	"8c3d37c819544da2", "73e1996689dcd4d6", "1dfab7ae32ff9c82", "679dd514582f9fcf",
	"0f6d2b697bd44da8", "77e36f7304c48942", "3f9d85a86a1d36c8", "1112e6ad91d692a1"
];

const go2jsSHA512256IV = [
	"22312194fc2bf72c", "9f555fa3c84c64c2", "2393b86b6f53b151", "963877195940eabd",
	"96283ee2a88effe3", "be5e1e2553863992", "2b0199fc2c85b8aa", "0eb72ddc81c52ca2"
];

function go2jsShaRor(value, bits) {
	return ((value >> BigInt(bits)) | (value << BigInt(64 - bits))) & go2jsMask64;
}

function go2jsShaShr(value, bits) {
	return (value >> BigInt(bits)) & go2jsMask64;
}

// go2jsBigFromBytesBE reads a number out of bytes most significant byte first,
// which is the order SHA-2 reads its message in and the order it writes its digest
// in, unlike every other number here which is held little end first.
function go2jsBigFromBytesBE(bytes) {
	let value = 0n;

	for (let index = 0; index < bytes.length; index++) {
		value = (value << 8n) | BigInt(Number(bytes[index]) & 255);
	}

	return value;
}

// go2jsBigBytesBE is a number written out most significant byte first.
function go2jsBigBytesBE(value, width) {
	const bytes = [];

	for (let index = width - 1; index >= 0; index--) {
		bytes.push(Number((BigInt(value) >> BigInt(index * 8)) & 0xffn));
	}

	return bytes;
}

// go2jsSHA512Block folds one block of 128 bytes into the state. A word of SHA-512
// is 64 bits and a JavaScript number cannot hold one exactly, so every word is a
// BigInt and every result is masked back to 64 bits, which is what the arithmetic
// of the standard is on.
function go2jsSHA512Block(state, block) {
	const w = [];

	for (let index = 0; index < 16; index++) {
		w.push(go2jsBigFromBytesBE(block.slice(index * 8, index * 8 + 8)));
	}

	for (let index = 16; index < 80; index++) {
		const s0 = go2jsShaRor(w[index - 15], 1) ^ go2jsShaRor(w[index - 15], 8) ^ go2jsShaShr(w[index - 15], 7);
		const s1 = go2jsShaRor(w[index - 2], 19) ^ go2jsShaRor(w[index - 2], 61) ^ go2jsShaShr(w[index - 2], 6);

		w.push((w[index - 16] + s0 + w[index - 7] + s1) & go2jsMask64);
	}

	let a = state[0];
	let b = state[1];
	let c = state[2];
	let d = state[3];
	let e = state[4];
	let f = state[5];
	let g = state[6];
	let h = state[7];

	for (let index = 0; index < 80; index++) {
		const S1 = go2jsShaRor(e, 14) ^ go2jsShaRor(e, 18) ^ go2jsShaRor(e, 41);
		const ch = (e & f) ^ ((~e & go2jsMask64) & g);
		const temp1 = (h + S1 + ch + BigInt("0x" + go2jsSHA512K[index]) + w[index]) & go2jsMask64;
		const S0 = go2jsShaRor(a, 28) ^ go2jsShaRor(a, 34) ^ go2jsShaRor(a, 39);
		const maj = (a & b) ^ (a & c) ^ (b & c);
		const temp2 = (S0 + maj) & go2jsMask64;

		h = g;
		g = f;
		f = e;
		e = (d + temp1) & go2jsMask64;
		d = c;
		c = b;
		b = a;
		a = (temp1 + temp2) & go2jsMask64;
	}

	return [
		(state[0] + a) & go2jsMask64, (state[1] + b) & go2jsMask64,
		(state[2] + c) & go2jsMask64, (state[3] + d) & go2jsMask64,
		(state[4] + e) & go2jsMask64, (state[5] + f) & go2jsMask64,
		(state[6] + g) & go2jsMask64, (state[7] + h) & go2jsMask64
	];
}

// go2jsSHA512Digest is the hash of a whole message: the state after the last block,
// with the length of the message written into the end of the last one, cut down to
// the width the hash is. The two sizes wider than 384 bits are the state read a
// half of at a time rather than as one word, which is the truncation Go applies.
function go2jsSHA512Digest(bytes, state, length, size) {
	const message = (bytes || []).slice();
	const total = message.length + length;
	const bits = BigInt(total) * 8n;

	message.push(0x80);

	while (message.length % 128 !== 112) {
		message.push(0);
	}

	// The length of the message goes at the end of the last block as a 128-bit
	// number written most significant byte first: the high half of the counter
	// comes before the low half of it, and both of them are written the way every
	// other part of a message is.
	for (let index = 0; index < 8; index++) {
		message.push(0);
	}

	message.push(...go2jsBigBytesBE(bits, 8));

	let hashed = state || go2jsSHA512Init();

	for (let offset = 0; offset < message.length; offset += 128) {
		hashed = go2jsSHA512Block(hashed, message.slice(offset, offset + 128));
	}

	const out = [];

	for (const word of hashed) {
		out.push(...go2jsBigBytesBE(word, 8));
	}

	return out.slice(0, size);
}

function go2jsSHA512Init(iv) {
	return (iv || go2jsSHA512IV).map(value => BigInt("0x" + value));
}

function go2jsSHA512Hash(data, size, iv) {
	return go2jsSHA512Digest(go2jsToArray(data).map(item => Number(item) & 255), go2jsSHA512Init(iv), 0, size);
}

function go2jsSHA512Sum512(data) {
	return go2jsSHA512Hash(data, 64, go2jsSHA512IV);
}

function go2jsSHA512Sum384(data) {
	return go2jsSHA512Hash(data, 48, go2jsSHA384IV);
}

function go2jsSHA512Sum224(data) {
	return go2jsSHA512Hash(data, 28, go2jsSHA512224IV);
}

function go2jsSHA512Sum256(data) {
	return go2jsSHA512Hash(data, 32, go2jsSHA512256IV);
}

// go2jsSHA512Hasher is a hash that keeps its state between the writes to it, which
// is the state itself, what has been written but is not yet a whole block, and how
// many bytes have been written in all.
function go2jsSHA512Hasher(size, iv) {
	const state = {hashed: go2jsSHA512Init(iv), buffer: [], length: 0, size: size};

	const write = (data) => {
		const bytes = Array.from(go2jsToArray(data), item => Number(item) & 255);
		const combined = state.buffer.concat(bytes);

		for (let offset = 0; offset + 128 <= combined.length; offset += 128) {
			state.hashed = go2jsSHA512Block(state.hashed, combined.slice(offset, offset + 128));
			state.length += 128;
		}

		state.buffer = combined.slice(state.length);
		return bytes.length;
	};

	const sum = (target, width) => go2jsToArray(target).concat(
		go2jsSHA512Digest(state.buffer.slice(), state.hashed, state.length, width));

	return go2jsInterface({
		Size() {
			return state.size;
		},
		BlockSize() {
			return 128;
		},
		Reset() {
			state.hashed = go2jsSHA512Init(iv);
			state.buffer = [];
			state.length = 0;
		},
		Write: write,
		Sum(target) {
			return sum(target, state.size);
		},
		Sum512(target) {
			return sum(target, 64);
		},
		Sum384(target) {
			return sum(target, 48);
		},
		Size512() {
			return 64;
		}
	}, "hash.Hash");
}

function go2jsSHA512New() {
	return go2jsSHA512Hasher(64, go2jsSHA512IV);
}

function go2jsSHA384New() {
	return go2jsSHA512Hasher(48, go2jsSHA384IV);
}

function go2jsSHA512224New() {
	return go2jsSHA512Hasher(28, go2jsSHA512224IV);
}

function go2jsSHA512256New() {
	return go2jsSHA512Hasher(32, go2jsSHA512256IV);
}

// A checksum table is the whole running value written out one byte at a time: the
// entry for a byte is what folding that byte in gives, so the fold is a lookup
// rather than a walk over the bits of the byte.
function go2jsCRCTable(polynomial) {
	const table = [];

	for (let index = 0; index < 256; index++) {
		let value = index >>> 0;

		for (let bit = 0; bit < 8; bit++) {
			value = (value & 1) !== 0 ? ((value >>> 1) ^ polynomial) >>> 0 : (value >>> 1) >>> 0;
		}

		table.push(value >>> 0);
	}

	return table;
}

let go2jsCRC32IEEETable = null;
let go2jsCRC32CastagnoliTable = null;
let go2jsCRC32KoopmanTable = null;

// go2jsCRC32TableValue is the table a program asks for by the name of its
// polynomial, which is a table of 256 numbers rather than a polynomial.
function go2jsCRC32TableValue(which) {
	if (which === "Castagnoli") {
		if (go2jsCRC32CastagnoliTable === null) {
			go2jsCRC32CastagnoliTable = go2jsCRCTable(0x82f63b78);
		}

		return go2jsCRC32CastagnoliTable.slice();
	}

	if (which === "Koopman") {
		if (go2jsCRC32KoopmanTable === null) {
			go2jsCRC32KoopmanTable = go2jsCRCTable(0xeb31d82e);
		}

		return go2jsCRC32KoopmanTable.slice();
	}

	if (go2jsCRC32IEEETable === null) {
		go2jsCRC32IEEETable = go2jsCRCTable(0xedb88320);
	}

	return go2jsCRC32IEEETable.slice();
}

// go2jsCRC32Update folds bytes into a running CRC of the 32-bit kind. The value is
// held complemented, which is what the table was worked out for, so it is
// complemented going in and coming out.
function go2jsCRC32Update(crc, table, bytes) {
	const entries = go2jsCRC32TableOf(table);
	let value = (Number(crc) ^ 0xffffffff) >>> 0;

	for (const byte of bytes) {
		value = (value >>> 8) ^ entries[(value ^ Number(byte)) & 255];
	}

	return (value ^ 0xffffffff) >>> 0;
}

function go2jsCRC32ChecksumIEEE(data) {
	return go2jsCRC32Update(0, go2jsCRC32TableValue("IEEE"), go2jsToArray(data));
}

function go2jsCRC32Checksum(data, table) {
	if (table === undefined || table === null) {
		return go2jsCRC32ChecksumIEEE(data);
	}

	return go2jsCRC32Update(0, go2jsCRC32TableOf(table), go2jsToArray(data));
}

// go2jsCRC32TableOf is the table of entries a table value stands for, since the
// table of a program of this runtime is a value carrying the entries rather than
// the entries themselves.
// go2jsCRC32IEEETableValue is the table of the IEEE polynomial, which the standard
// library holds as a value of its own for a program to hand to New.
function go2jsCRC32IEEETableValue() {
	return go2jsCRC32TableValue("IEEE");
}

function go2jsCRC32TableOf(table) {
	if (Array.isArray(table)) {
		return table;
	}

	if (table !== null && typeof table === "object" && Array.isArray(table.table)) {
		return table.table;
	}

	return go2jsCRC32TableValue("IEEE");
}


function go2jsCRC32MakeTable(polynomial) {
	return {table: go2jsCRC32TableFor(Number(polynomial)), which: "custom"};
}

// go2jsCRC32TableFor is the table of the polynomial a number is, which is one of
// the three the standard library holds and a table of its own for any other.
function go2jsCRC32TableFor(polynomial) {
	const value = Number(polynomial) >>> 0;

	if (value === 0xedb88320) {
		return go2jsCRC32TableValue("IEEE");
	}

	if (value === 0x82f63b78) {
		return go2jsCRC32TableValue("Castagnoli");
	}

	if (value === 0xeb31d82e) {
		return go2jsCRC32TableValue("Koopman");
	}

	return go2jsCRCTable(value);
}

function go2jsHashCRC32(table) {
	return go2jsCRCHasher32(table);
}

function go2jsHashCRC32IEEE() {
	return go2jsCRCHasher32(go2jsCRC32TableValue("IEEE"));
}

// go2jsCRCHasher32 is a running CRC kept between the writes to it, which is the
// value so far and the table the next byte is folded in through.
function go2jsCRCHasher32(table) {
	const state = {value: 0};
	const entries = go2jsCRC32TableOf(table);

	return go2jsInterface({
		Size() {
			return 4;
		},
		BlockSize() {
			return 1;
		},
		Reset() {
			state.value = 0;
		},
		Write(data) {
			const bytes = Array.from(go2jsToArray(data), item => Number(item) & 255);

			state.value = go2jsCRC32Update(state.value, entries, bytes);
			return bytes.length;
		},
		Sum(target) {
			return go2jsToArray(target).concat(go2jsBigBytesBE(BigInt(state.value >>> 0), 4));
		},
		Sum32() {
			return state.value >>> 0;
		}
	}, "hash.Hash32");
}

// go2jsCRC64Polynomial is the reflected polynomial of a CRC of the 64-bit kind,
// which is the form Go hands to MakeTable and the form the fold below uses.
function go2jsCRC64Polynomial(which) {
	if (which === "ECMA") {
		return 0xc96c5795d7870f42n;
	}

	return 0xd800000000000000n;
}

// go2jsCRC64Fold folds bytes into a running CRC of the 64-bit kind. The value is
// held as a BigInt because 64 bits are more than a JavaScript number holds, and it
// is complemented either side of the fold for the same reason the 32-bit one is.
function go2jsCRC64Fold(crc, polynomial, bytes) {
	let value = (BigInt(crc) ^ go2jsMask64) & go2jsMask64;

	for (const byte of bytes) {
		value ^= BigInt(Number(byte) & 255);

		for (let bit = 0; bit < 8; bit++) {
			value = (value & 1n) !== 0n ? ((value >> 1n) ^ polynomial) & go2jsMask64 : (value >> 1n) & go2jsMask64;
		}
	}

	return (value ^ go2jsMask64) & go2jsMask64;
}

function go2jsCRC64Table(which) {
	return {which: which, polynomial: go2jsCRC64Polynomial(which)};
}

function go2jsCRC64Checksum(data, table) {
	return go2jsCRC64Fold(0n, go2jsCRC64TablePolynomial(table), go2jsToArray(data));
}

function go2jsCRC64Update(crc, table, bytes) {
	return go2jsCRC64Fold(crc, go2jsCRC64TablePolynomial(table), bytes);
}

function go2jsCRC64MakeTable(polynomial) {
	return {which: "custom", polynomial: BigInt(polynomial) & go2jsMask64};
}

// go2jsCRC64TablePolynomial is the polynomial a table stands for, which is what a
// table of a program that made its own is asked for.
function go2jsCRC64TablePolynomial(table) {
	if (table === undefined || table === null) {
		return go2jsCRC64Polynomial("ISO");
	}

	if (typeof table === "object" && table.polynomial !== undefined) {
		return BigInt(table.polynomial) & go2jsMask64;
	}

	return go2jsCRC64Polynomial("ISO");
}

function go2jsCRCHasher64(table) {
	const polynomial = go2jsCRC64TablePolynomial(table);
	const state = {value: 0n};

	return go2jsInterface({
		Size() {
			return 8;
		},
		BlockSize() {
			return 1;
		},
		Reset() {
			state.value = 0n;
		},
		Write(data) {
			const bytes = Array.from(go2jsToArray(data), item => Number(item) & 255);

			state.value = go2jsCRC64Fold(state.value, polynomial, bytes);
			return bytes.length;
		},
		Sum(target) {
			return go2jsToArray(target).concat(go2jsBigBytesBE(state.value, 8));
		},
		Sum64() {
			return state.value;
		}
	}, "hash.Hash64");
}

function go2jsHashCRC64(table) {
	return go2jsCRCHasher64(table);
}

// FNV is a multiply and then an exclusive or, or the other way round: the two
// orders give different hashes of the same bytes and are named apart, and what
// separates the 32-bit family from the 64-bit one is the width of the multiply,
// which a JavaScript number holds for 32 bits and a BigInt holds for 64.
const go2jsFNVPrime32 = 16777619;
const go2jsFNVPrime64 = 1099511628211n;
const go2jsFNVPrime128 = 309485009821345068724781371n;

// go2jsFNVBasis is where a hash of a given width starts, which is the offset basis
// of the family and also the hash of an empty message.
function go2jsFNVBasis(width) {
	if (width === 32) {
		return 0x811c9dc5 >>> 0;
	}

	if (width === 64) {
		return 0xcbf29ce484222325n;
	}

	return 0x6c62272e07bb014262b821756295c58dn;
}

// go2jsFNVPrime is the prime a hash of a given width multiplies by, which is a
// power of two with a little added: 16777619 for 32 bits, the 64th power of two
// plus 2 for 64 and the 88th plus 315 for 128.
function go2jsFNVPrime(width) {
	if (width === 32) {
		return go2jsFNVPrime32;
	}

	if (width === 64) {
		return go2jsFNVPrime64;
	}

	return go2jsFNVPrime128;
}

function go2jsFNVMask(width) {
	if (width === 32) {
		return 0xffffffffn;
	}

	if (width === 64) {
		return go2jsMask64;
	}

	return (1n << 128n) - 1n;
}

// go2jsFNVFold is one byte folded into the running value, which is a multiply and
// an exclusive or in one order or the other: FNV-1 multiplies first and FNV-1a
// multiplies last.
function go2jsFNVFold(value, byte, width, primeFirst) {
	if (width === 32) {
		return primeFirst
			? (Math.imul(value, go2jsFNVPrime32) ^ byte) >>> 0
			: Math.imul(value ^ byte, go2jsFNVPrime32) >>> 0;
	}

	const mask = go2jsFNVMask(width);
	const prime = go2jsFNVPrime(width);

	return primeFirst
		? ((value * prime) ^ BigInt(byte)) & mask
		: ((value ^ BigInt(byte)) * prime) & mask;
}

// go2jsFNVBytes is a hash value as the bytes a hash writes out: a number of 32
// bits held in one number, and a wider one little end first, which is the order Go
// writes the bytes of a number wider than a word in.
function go2jsFNVBytes(value, width) {
	if (width === 32) {
		return go2jsBigBytesBE(BigInt(value >>> 0), 4);
	}

	return go2jsBigBytesBE(BigInt(value) & go2jsFNVMask(width), width / 8);
}


// go2jsFNVHasher is an FNV hash kept between the writes to it, which is the value
// so far and whether the multiply comes before the exclusive or.
function go2jsFNVHasher(width, primeFirst) {
	const basis = go2jsFNVBasis(width);
	const state = {value: basis};

	const fold = (bytes) => {
		for (const byte of bytes) {
			state.value = go2jsFNVFold(state.value, Number(byte) & 255, width, primeFirst);
		}
	};

	return go2jsInterface({
		Size() {
			return width / 8;
		},
		BlockSize() {
			return 1;
		},
		Reset() {
			state.value = basis;
		},
		Write(data) {
			const bytes = Array.from(go2jsToArray(data), item => Number(item) & 255);

			fold(bytes);
			return bytes.length;
		},
		Sum(target) {
			return go2jsToArray(target).concat(go2jsFNVBytes(state.value, width));
		},
		Sum32() {
			return Number(state.value) >>> 0;
		},
		Sum64() {
			return BigInt(state.value) & go2jsMask64;
		}
	}, "hash.Hash");
}

function go2jsFNVNew32() {
	return go2jsFNVHasher(32, true);
}

function go2jsFNVNew32a() {
	return go2jsFNVHasher(32, false);
}

function go2jsFNVNew64() {
	return go2jsFNVHasher(64, true);
}

function go2jsFNVNew64a() {
	return go2jsFNVHasher(64, false);
}

function go2jsFNVNew128() {
	return go2jsFNVHasher(128, true);
}

function go2jsFNVNew128a() {
	return go2jsFNVHasher(128, false);
}

// Adler-32 is the checksum zlib made: one running total of the bytes and one of
// those totals, which between them stand for the whole message.
function go2jsAdler32Checksum(data) {
	let low = 1;
	let high = 0;

	for (const byte of go2jsToArray(data)) {
		low = (low + (Number(byte) & 255)) % 65521;
		high = (high + low) % 65521;
	}

	return (high * 65536 + low) >>> 0;
}

function go2jsHashAdler32() {
	const state = {value: 1};

	return go2jsInterface({
		Size() {
			return 4;
		},
		BlockSize() {
			return 1;
		},
		Reset() {
			state.value = 1;
		},
		Write(data) {
			const bytes = Array.from(go2jsToArray(data), item => Number(item) & 255);

			let low = state.value & 0xffff;
			let high = (state.value >>> 16) & 0xffff;

			for (const byte of bytes) {
				low = (low + byte) % 65521;
				high = (high + low) % 65521;
			}

			state.value = ((high * 65536 + low) >>> 0);
			return bytes.length;
		},
		Sum(target) {
			return go2jsToArray(target).concat(go2jsBigBytesBE(BigInt(state.value >>> 0), 4));
		},
		Sum32() {
			return state.value >>> 0;
		}
	}, "hash.Hash32");
}
`
}
