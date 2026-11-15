package javascript

import "strings"

// The range tables of the unicode package. Go keeps each as a list of ranges;
// the host already knows what the standard names mean, so a table here is the
// name it is known by, and a rune is asked of it by the host's own property
// escapes.

var unicodeTableList = []struct{ name, kind, key string }{
	{"ASCII_Hex_Digit", "prop", "ASCII_Hex_Digit"},
	{"Adlam", "script", "Adlam"},
	{"Ahom", "script", "Ahom"},
	{"Anatolian_Hieroglyphs", "script", "Anatolian_Hieroglyphs"},
	{"Arabic", "script", "Arabic"},
	{"Armenian", "script", "Armenian"},
	{"Avestan", "script", "Avestan"},
	{"Balinese", "script", "Balinese"},
	{"Bamum", "script", "Bamum"},
	{"Bassa_Vah", "script", "Bassa_Vah"},
	{"Batak", "script", "Batak"},
	{"Bengali", "script", "Bengali"},
	{"Bhaiksuki", "script", "Bhaiksuki"},
	{"Bidi_Control", "prop", "Bidi_Control"},
	{"Bopomofo", "script", "Bopomofo"},
	{"Brahmi", "script", "Brahmi"},
	{"Braille", "script", "Braille"},
	{"Buginese", "script", "Buginese"},
	{"Buhid", "script", "Buhid"},
	{"C", "cat", "C"},
	{"Canadian_Aboriginal", "script", "Canadian_Aboriginal"},
	{"Carian", "script", "Carian"},
	{"Caucasian_Albanian", "script", "Caucasian_Albanian"},
	{"Cc", "cat", "Cc"},
	{"Cf", "cat", "Cf"},
	{"Chakma", "script", "Chakma"},
	{"Cham", "script", "Cham"},
	{"Cherokee", "script", "Cherokee"},
	{"Chorasmian", "script", "Chorasmian"},
	{"Co", "cat", "Co"},
	{"Common", "script", "Common"},
	{"Coptic", "script", "Coptic"},
	{"Cs", "cat", "Cs"},
	{"Cuneiform", "script", "Cuneiform"},
	{"Cypriot", "script", "Cypriot"},
	{"Cypro_Minoan", "script", "Cypro_Minoan"},
	{"Cyrillic", "script", "Cyrillic"},
	{"Dash", "prop", "Dash"},
	{"Deprecated", "prop", "Deprecated"},
	{"Deseret", "script", "Deseret"},
	{"Devanagari", "script", "Devanagari"},
	{"Diacritic", "prop", "Diacritic"},
	{"Digit", "cat", "Nd"},
	{"Dives_Akuru", "script", "Dives_Akuru"},
	{"Dogra", "script", "Dogra"},
	{"Duployan", "script", "Duployan"},
	{"Egyptian_Hieroglyphs", "script", "Egyptian_Hieroglyphs"},
	{"Elbasan", "script", "Elbasan"},
	{"Elymaic", "script", "Elymaic"},
	{"Ethiopic", "script", "Ethiopic"},
	{"Extender", "prop", "Extender"},
	{"Georgian", "script", "Georgian"},
	{"Glagolitic", "script", "Glagolitic"},
	{"Gothic", "script", "Gothic"},
	{"Grantha", "script", "Grantha"},
	{"Greek", "script", "Greek"},
	{"Gujarati", "script", "Gujarati"},
	{"Gunjala_Gondi", "script", "Gunjala_Gondi"},
	{"Gurmukhi", "script", "Gurmukhi"},
	{"Han", "script", "Han"},
	{"Hangul", "script", "Hangul"},
	{"Hanifi_Rohingya", "script", "Hanifi_Rohingya"},
	{"Hanunoo", "script", "Hanunoo"},
	{"Hatran", "script", "Hatran"},
	{"Hebrew", "script", "Hebrew"},
	{"Hex_Digit", "prop", "Hex_Digit"},
	{"Hiragana", "script", "Hiragana"},
	{"Hyphen", "prop", "Hyphen"},
	{"IDS_Binary_Operator", "prop", "IDS_Binary_Operator"},
	{"IDS_Trinary_Operator", "prop", "IDS_Trinary_Operator"},
	{"Ideographic", "prop", "Ideographic"},
	{"Imperial_Aramaic", "script", "Imperial_Aramaic"},
	{"Inherited", "script", "Inherited"},
	{"Inscriptional_Pahlavi", "script", "Inscriptional_Pahlavi"},
	{"Inscriptional_Parthian", "script", "Inscriptional_Parthian"},
	{"Javanese", "script", "Javanese"},
	{"Join_Control", "prop", "Join_Control"},
	{"Kaithi", "script", "Kaithi"},
	{"Kannada", "script", "Kannada"},
	{"Katakana", "script", "Katakana"},
	{"Kawi", "script", "Kawi"},
	{"Kayah_Li", "script", "Kayah_Li"},
	{"Kharoshthi", "script", "Kharoshthi"},
	{"Khitan_Small_Script", "script", "Khitan_Small_Script"},
	{"Khmer", "script", "Khmer"},
	{"Khojki", "script", "Khojki"},
	{"Khudawadi", "script", "Khudawadi"},
	{"L", "cat", "L"},
	{"Lao", "script", "Lao"},
	{"Latin", "script", "Latin"},
	{"Lepcha", "script", "Lepcha"},
	{"Letter", "cat", "L"},
	{"Limbu", "script", "Limbu"},
	{"Linear_A", "script", "Linear_A"},
	{"Linear_B", "script", "Linear_B"},
	{"Lisu", "script", "Lisu"},
	{"Ll", "cat", "Ll"},
	{"Lm", "cat", "Lm"},
	{"Lo", "cat", "Lo"},
	{"Logical_Order_Exception", "prop", "Logical_Order_Exception"},
	{"Lower", "cat", "Ll"},
	{"Lt", "cat", "Lt"},
	{"Lu", "cat", "Lu"},
	{"Lycian", "script", "Lycian"},
	{"Lydian", "script", "Lydian"},
	{"M", "cat", "M"},
	{"Mahajani", "script", "Mahajani"},
	{"Makasar", "script", "Makasar"},
	{"Malayalam", "script", "Malayalam"},
	{"Mandaic", "script", "Mandaic"},
	{"Manichaean", "script", "Manichaean"},
	{"Marchen", "script", "Marchen"},
	{"Mark", "cat", "M"},
	{"Masaram_Gondi", "script", "Masaram_Gondi"},
	{"Mc", "cat", "Mc"},
	{"Me", "cat", "Me"},
	{"Medefaidrin", "script", "Medefaidrin"},
	{"Meetei_Mayek", "script", "Meetei_Mayek"},
	{"Mende_Kikakui", "script", "Mende_Kikakui"},
	{"Meroitic_Cursive", "script", "Meroitic_Cursive"},
	{"Meroitic_Hieroglyphs", "script", "Meroitic_Hieroglyphs"},
	{"Miao", "script", "Miao"},
	{"Mn", "cat", "Mn"},
	{"Modi", "script", "Modi"},
	{"Mongolian", "script", "Mongolian"},
	{"Mro", "script", "Mro"},
	{"Multani", "script", "Multani"},
	{"Myanmar", "script", "Myanmar"},
	{"N", "cat", "N"},
	{"Nabataean", "script", "Nabataean"},
	{"Nag_Mundari", "script", "Nag_Mundari"},
	{"Nandinagari", "script", "Nandinagari"},
	{"Nd", "cat", "Nd"},
	{"New_Tai_Lue", "script", "New_Tai_Lue"},
	{"Newa", "script", "Newa"},
	{"Nko", "script", "Nko"},
	{"Nl", "cat", "Nl"},
	{"No", "cat", "No"},
	{"Noncharacter_Code_Point", "prop", "Noncharacter_Code_Point"},
	{"Number", "cat", "N"},
	{"Nushu", "script", "Nushu"},
	{"Nyiakeng_Puachue_Hmong", "script", "Nyiakeng_Puachue_Hmong"},
	{"Ogham", "script", "Ogham"},
	{"Ol_Chiki", "script", "Ol_Chiki"},
	{"Old_Hungarian", "script", "Old_Hungarian"},
	{"Old_Italic", "script", "Old_Italic"},
	{"Old_North_Arabian", "script", "Old_North_Arabian"},
	{"Old_Permic", "script", "Old_Permic"},
	{"Old_Persian", "script", "Old_Persian"},
	{"Old_Sogdian", "script", "Old_Sogdian"},
	{"Old_South_Arabian", "script", "Old_South_Arabian"},
	{"Old_Turkic", "script", "Old_Turkic"},
	{"Old_Uyghur", "script", "Old_Uyghur"},
	{"Oriya", "script", "Oriya"},
	{"Osage", "script", "Osage"},
	{"Osmanya", "script", "Osmanya"},
	{"Other", "cat", "C"},
	{"Other_Alphabetic", "prop", "Other_Alphabetic"},
	{"Other_Default_Ignorable_Code_Point", "prop", "Other_Default_Ignorable_Code_Point"},
	{"Other_Grapheme_Extend", "prop", "Other_Grapheme_Extend"},
	{"Other_ID_Continue", "prop", "Other_ID_Continue"},
	{"Other_ID_Start", "prop", "Other_ID_Start"},
	{"Other_Lowercase", "prop", "Other_Lowercase"},
	{"Other_Math", "prop", "Other_Math"},
	{"Other_Uppercase", "prop", "Other_Uppercase"},
	{"P", "cat", "P"},
	{"Pahawh_Hmong", "script", "Pahawh_Hmong"},
	{"Palmyrene", "script", "Palmyrene"},
	{"Pattern_Syntax", "prop", "Pattern_Syntax"},
	{"Pattern_White_Space", "prop", "Pattern_White_Space"},
	{"Pau_Cin_Hau", "script", "Pau_Cin_Hau"},
	{"Pc", "cat", "Pc"},
	{"Pd", "cat", "Pd"},
	{"Pe", "cat", "Pe"},
	{"Pf", "cat", "Pf"},
	{"Phags_Pa", "script", "Phags_Pa"},
	{"Phoenician", "script", "Phoenician"},
	{"Pi", "cat", "Pi"},
	{"Po", "cat", "Po"},
	{"Prepended_Concatenation_Mark", "prop", "Prepended_Concatenation_Mark"},
	{"Ps", "cat", "Ps"},
	{"Psalter_Pahlavi", "script", "Psalter_Pahlavi"},
	{"Punct", "cat", "P"},
	{"Quotation_Mark", "prop", "Quotation_Mark"},
	{"Radical", "prop", "Radical"},
	{"Regional_Indicator", "prop", "Regional_Indicator"},
	{"Rejang", "script", "Rejang"},
	{"Runic", "script", "Runic"},
	{"S", "cat", "S"},
	{"STerm", "prop", "STerm"},
	{"Samaritan", "script", "Samaritan"},
	{"Saurashtra", "script", "Saurashtra"},
	{"Sc", "cat", "Sc"},
	{"Sentence_Terminal", "prop", "Sentence_Terminal"},
	{"Sharada", "script", "Sharada"},
	{"Shavian", "script", "Shavian"},
	{"Siddham", "script", "Siddham"},
	{"SignWriting", "script", "SignWriting"},
	{"Sinhala", "script", "Sinhala"},
	{"Sk", "cat", "Sk"},
	{"Sm", "cat", "Sm"},
	{"So", "cat", "So"},
	{"Soft_Dotted", "prop", "Soft_Dotted"},
	{"Sogdian", "script", "Sogdian"},
	{"Sora_Sompeng", "script", "Sora_Sompeng"},
	{"Soyombo", "script", "Soyombo"},
	{"Space", "cat", "Z"},
	{"Sundanese", "script", "Sundanese"},
	{"Syloti_Nagri", "script", "Syloti_Nagri"},
	{"Symbol", "cat", "S"},
	{"Syriac", "script", "Syriac"},
	{"Tagalog", "script", "Tagalog"},
	{"Tagbanwa", "script", "Tagbanwa"},
	{"Tai_Le", "script", "Tai_Le"},
	{"Tai_Tham", "script", "Tai_Tham"},
	{"Tai_Viet", "script", "Tai_Viet"},
	{"Takri", "script", "Takri"},
	{"Tamil", "script", "Tamil"},
	{"Tangsa", "script", "Tangsa"},
	{"Tangut", "script", "Tangut"},
	{"Telugu", "script", "Telugu"},
	{"Terminal_Punctuation", "prop", "Terminal_Punctuation"},
	{"Thaana", "script", "Thaana"},
	{"Thai", "script", "Thai"},
	{"Tibetan", "script", "Tibetan"},
	{"Tifinagh", "script", "Tifinagh"},
	{"Tirhuta", "script", "Tirhuta"},
	{"Title", "cat", "Lt"},
	{"Toto", "script", "Toto"},
	{"Ugaritic", "script", "Ugaritic"},
	{"Unified_Ideograph", "prop", "Unified_Ideograph"},
	{"Upper", "cat", "Lu"},
	{"Vai", "script", "Vai"},
	{"Variation_Selector", "prop", "Variation_Selector"},
	{"Vithkuqi", "script", "Vithkuqi"},
	{"Wancho", "script", "Wancho"},
	{"Warang_Citi", "script", "Warang_Citi"},
	{"White_Space", "prop", "White_Space"},
	{"Yezidi", "script", "Yezidi"},
	{"Yi", "script", "Yi"},
	{"Z", "cat", "Z"},
	{"Zanabazar_Square", "script", "Zanabazar_Square"},
	{"Zl", "cat", "Zl"},
	{"Zp", "cat", "Zp"},
	{"Zs", "cat", "Zs"},
}

var unicodeExtraFuncs = map[string]string{
	"Is":         "go2jsUnicodeIs",
	"In":         "go2jsUnicodeIn",
	"IsOneOf":    "go2jsUnicodeIsOneOf",
	"SimpleFold": "go2jsUnicodeSimpleFold",
}

func init() {
	for name, value := range unicodeExtraFuncs {
		unicodeFuncs[name] = value
	}

	for _, table := range unicodeTableList {
		unicodeFuncs[table.name] = ""
		packageVarValues["unicode."+table.name] = "go2jsUnicodeTable_" + table.name
		packageVarTypes["unicode."+table.name] = "*unicode.RangeTable"
	}

	supportedStdlibPackages["unicode"] = funcSet(unicodeFuncs)
}

func unicodeTableRuntimeSource() string {
	var b strings.Builder

	for _, table := range unicodeTableList {
		b.WriteString("function go2jsUnicodeTable_" + table.name + "() {\n\treturn go2jsUnicodeTableOf(\"" + table.kind + "\", \"" + table.key + "\");\n}\n\n")
	}

	b.WriteString(`const go2jsUnicodeTableCache = {};

function go2jsUnicodeTableOf(kind, key) {
	const id = kind + ":" + key;

	if (go2jsUnicodeTableCache[id] === undefined) {
		let expression = null;

		try {
			if (kind === "script") {
				expression = new RegExp("^\\p{Script=" + key + "}$", "u");
			} else {
				expression = new RegExp("^\\p{" + key + "}$", "u");
			}
		} catch (e) {
			expression = null;
		}

		go2jsUnicodeTableCache[id] = { expression: expression, __go2js_type_name: "*unicode.RangeTable" };
	}

	return go2jsUnicodeTableCache[id];
}

function go2jsUnicodeIs(table, value) {
	const ch = go2jsUnicodeRune(value);
	const inner = go2jsUnwrap(table);

	if (ch === "" || inner === null || inner === undefined || inner.expression === null || inner.expression === undefined) {
		return false;
	}

	return inner.expression.test(ch);
}

function go2jsUnicodeIn(value, ...tables) {
	for (const table of tables) {
		if (go2jsUnicodeIs(table, value)) {
			return true;
		}
	}

	return false;
}

function go2jsUnicodeIsOneOf(tables, value) {
	for (const table of go2jsToArray(tables)) {
		if (go2jsUnicodeIs(table, value)) {
			return true;
		}
	}

	return false;
}

// The orbits of runes that fold into one another, found once from what the
// host knows of upper and lower case, and kept as a ring through each of them.
let go2jsUnicodeFoldNext = null;

function go2jsUnicodeFoldBuild() {
	const parent = new Map();
	const find = function(x) {
		while (parent.has(x) && parent.get(x) !== x) {
			parent.set(x, parent.get(parent.get(x)));
			x = parent.get(x);
		}

		return x;
	};
	const join = function(a, b) {
		if (!parent.has(a)) {
			parent.set(a, a);
		}
		if (!parent.has(b)) {
			parent.set(b, b);
		}
		const ra = find(a);
		const rb = find(b);
		if (ra !== rb) {
			parent.set(ra, rb);
		}
	};

	for (let r = 0; r <= 0x1ffff; r++) {
		if (r >= 0xd800 && r <= 0xdfff) {
			continue;
		}

		const ch = String.fromCodePoint(r);
		const lower = ch.toLowerCase();
		const upper = ch.toUpperCase();

		if (lower !== ch && [...lower].length === 1) {
			join(r, lower.codePointAt(0));
		}

		if (upper !== ch && [...upper].length === 1) {
			join(r, upper.codePointAt(0));
		}
	}

	const groups = new Map();

	for (const r of parent.keys()) {
		const root = find(r);

		if (!groups.has(root)) {
			groups.set(root, []);
		}

		groups.get(root).push(r);
	}

	const next = new Map();

	for (const members of groups.values()) {
		members.sort(function(a, b) { return a - b; });

		for (let i = 0; i < members.length; i++) {
			next.set(members[i], members[(i + 1) % members.length]);
		}
	}

	return next;
}

function go2jsUnicodeSimpleFold(value) {
	const r = go2jsUnicodeCodePoint(value);

	if (r < 0 || r > 0x10ffff) {
		return r;
	}

	if (go2jsUnicodeFoldNext === null) {
		go2jsUnicodeFoldNext = go2jsUnicodeFoldBuild();
	}

	return go2jsUnicodeFoldNext.has(r) ? go2jsUnicodeFoldNext.get(r) : r;
}
`)

	return b.String()
}
