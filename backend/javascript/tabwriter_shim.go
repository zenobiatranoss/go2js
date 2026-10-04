package javascript

// tabwriterFuncs are the functions of text/tabwriter, which is one constructor: a
// writer is given its output and the widths it aligns to, and everything after
// that is written through it.
var tabwriterFuncs = map[string]string{
	"NewWriter": "go2jsTabwriterNewWriter",
}

// tabwriterMethods are the methods of a writer, written out by the type they hang
// off so that a value answers to them the way it does in Go.
var tabwriterMethods = map[string]string{
	"Write": "go2jsTabwriterWrite",
	"Flush": "go2jsTabwriterFlush",
	"Init":  "go2jsTabwriterInit",
}

// tabwriterConstants are the flags that say how a writer formats, each of them a
// different bit of the one number that carries them all, and the character a
// program brackets text with to keep it from being read as columns.
var tabwriterConstants = map[string]string{
	"FilterHTML":          "1",
	"StripEscape":         "2",
	"AlignRight":          "4",
	"DiscardEmptyColumns": "8",
	"TabIndent":           "16",
	"Debug":               "32",
	"Escape":              "255",
}

func init() {
	stdlibFuncMaps["text/tabwriter"] = tabwriterFuncs
	stdlibPkgAliases["text/tabwriter"] = []string{"tabwriter"}
	supportedStdlibPackages["text/tabwriter"] = funcSetWith(tabwriterConstants, tabwriterFuncs)

	// A writer is made by the package rather than by constructing a value of its
	// type, and the type is named here under both the name of the package and the
	// path it is imported by, since one is how a value is declared and the other
	// is how the name of the type is looked up.
	for _, key := range []string{"tabwriter.Writer", "text/tabwriter.Writer"} {
		packageTypes[key] = "go2jsTabwriterWriterType"
	}

	for name, value := range tabwriterConstants {
		for _, key := range []string{"text/tabwriter." + name, "tabwriter." + name} {
			packageConstants[key] = value
			packageVarValues[key] = value
			packageVarTypes[key] = "int"
		}
	}

	// The methods of a writer hang off the type they belong to, which is how a
	// method of a type of the standard library is written out.
	for _, key := range []string{"tabwriter.Writer", "text/tabwriter.Writer"} {
		helpers := syncMethodHelpers[key]

		if helpers == nil {
			helpers = map[string]string{}
			syncMethodHelpers[key] = helpers
		}

		for method, helper := range tabwriterMethods {
			helpers[method] = helper
		}
	}
}

// tabwriterRuntimeSource is the filter of text/tabwriter carried whole, since the
// widths of a column are a question about the lines around it rather than about
// the line it is on: a writer buffers what it is given and answers it out when it
// is flushed or when a line settles enough of its own.
func tabwriterRuntimeSource() string {
	return `
// The flags of a writer are one number, and each of them says one thing about how
// the columns are formatted, so each is one bit of that number.
const go2jsTabwriterFilterHTML = 1;
const go2jsTabwriterStripEscape = 2;
const go2jsTabwriterAlignRight = 4;
const go2jsTabwriterDiscardEmptyColumns = 8;
const go2jsTabwriterTabIndent = 16;
const go2jsTabwriterDebug = 32;

// go2jsTabwriterEscape is the character a program brackets text with so that the
// tabs and the line breaks inside it are text rather than columns, which is a byte
// that cannot be part of a character of UTF-8.
const go2jsTabwriterEscape = 255;

// go2jsTabwriterTab is what a tab of the output is written as, since padding
// written out one tab at a time needs several of them to stand for one cell.
const go2jsTabwriterTabs = [9, 9, 9, 9, 9, 9, 9, 9];

// go2jsTabwriterCell is one piece of a line: how much text it stands for, how wide
// that text is, and whether a tab stands after it, since a cell a tab stands
// after belongs to a column and one that does not is the rest of the line.
function go2jsTabwriterCell() {
	return {size: 0, width: 0, htab: false};
}

function go2jsTabwriterWriterType() {
	const writer = go2jsTabwriterNewWriter({Write: () => [0, null]}, 0, 0, 0, 32, 0);

	writer.output = null;

	return writer;
}

// go2jsTabwriterNewWriter is a writer of columns, which is given where to write
// to, the least width a cell may have, the width of a tab, the room to leave
// beside the text of a cell, what to pad with, and the flags.
function go2jsTabwriterNewWriter(output, minwidth, tabwidth, padding, padchar, flags) {
	const writer = {};

	writer.Write = (buf) => go2jsTabwriterWrite(writer, buf);
	writer.Flush = () => go2jsTabwriterFlush(writer);
	writer.Init = (output, minwidth, tabwidth, padding, padchar, flags) =>
		go2jsTabwriterInit(writer, output, minwidth, tabwidth, padding, padchar, flags);

	return go2jsTabwriterInit(writer, output, minwidth, tabwidth, padding, padchar, flags);
}

// go2jsTabwriterInit is Init: a writer that is given its widths again is a writer
// that has written nothing, since its state is what was written to it so far.
function go2jsTabwriterInit(writer, output, minwidth, tabwidth, padding, padchar, flags) {
	if (Number(minwidth) < 0 || Number(tabwidth) < 0 || Number(padding) < 0) {
		panic("tabwriter: negative minwidth, tabwidth, or padding");
	}

	writer.output = output;
	writer.minwidth = Number(minwidth);
	writer.tabwidth = Number(tabwidth);
	writer.padding = Number(padding);
	writer.padchar = Number(padchar) & 255;
	writer.flags = Number(flags) & 0xffffffff;

	// Padding with a tab leaves every cell aligned to the left whatever the flags
	// say, because a tab of the host cannot be placed to the right of anything.
	if (writer.padchar === 9) {
		writer.flags &= ~go2jsTabwriterAlignRight;
	}

	go2jsTabwriterReset(writer);

	return writer;
}

// go2jsTabwriterReset is the state of a writer that has written nothing: no text
// kept, no cell half written, and one empty line to write into.
function go2jsTabwriterReset(writer) {
	writer.buf = [];
	writer.pos = 0;
	writer.cell = go2jsTabwriterCell();
	writer.endChar = 0;
	writer.lines = [];
	writer.widths = [];

	go2jsTabwriterAddLine(writer);
}

// go2jsTabwriterAddLine is a line to write into, and a form feed says the line
// before it is not a fair account of what the next one holds.
function go2jsTabwriterAddLine(writer) {
	writer.lines.push([]);
}

// go2jsTabwriterAppend is text written into the cell being written, which is held
// apart from the tabs and the line breaks that end the cells.
function go2jsTabwriterAppend(writer, bytes) {
	for (const byte of bytes) {
		writer.buf.push(Number(byte) & 255);
	}

	writer.cell.size += bytes.length;
}

// go2jsTabwriterRuneCount is the width of text in characters, and a character of
// UTF-8 is the byte that does not continue the one before it.
function go2jsTabwriterRuneCount(bytes) {
	let count = 0;

	for (const byte of bytes) {
		if ((Number(byte) & 0xc0) !== 0x80) {
			count++;
		}
	}

	return count;
}

// go2jsTabwriterUpdateWidth is the width of what has been written to the cell
// being written, which is counted as it goes since the text of a tag or of an
// escaped piece of it is not text as far as the width of the cell is concerned.
function go2jsTabwriterUpdateWidth(writer) {
	writer.cell.width += go2jsTabwriterRuneCount(writer.buf.slice(writer.pos));
	writer.pos = writer.buf.length;
}

// go2jsTabwriterStartEscape is the beginning of text that is not read as columns,
// which ends at the character that closes it: the escape itself, the end of a
// tag, or the semicolon that ends an entity.
function go2jsTabwriterStartEscape(writer, ch) {
	if (ch === go2jsTabwriterEscape) {
		writer.endChar = go2jsTabwriterEscape;
	} else if (ch === 60) {
		writer.endChar = 62;
	} else if (ch === 38) {
		writer.endChar = 59;
	}
}

// go2jsTabwriterEndEscape is the end of text that was not read as columns: what
// it was is a piece of text whose width is its own, a tag of no width at all, or
// an entity that stands for one character.
function go2jsTabwriterEndEscape(writer) {
	if (writer.endChar === go2jsTabwriterEscape) {
		go2jsTabwriterUpdateWidth(writer);

		if ((writer.flags & go2jsTabwriterStripEscape) === 0) {
			writer.cell.width -= 2;
		}
	} else if (writer.endChar === 59) {
		writer.cell.width++;
	}

	writer.pos = writer.buf.length;
	writer.endChar = 0;
}

// go2jsTabwriterTerminateCell is the end of the cell being written, which is a
// cell of the line being written, and answers how many cells the line holds.
function go2jsTabwriterTerminateCell(writer, htab) {
	writer.cell.htab = htab;

	writer.lines[writer.lines.length - 1].push(writer.cell);
	writer.cell = go2jsTabwriterCell();

	return writer.lines[writer.lines.length - 1].length;
}

// go2jsTabwriterWrite0 is what is written out, which is a fault of the writer
// underneath rather than of the one asking for it.
function go2jsTabwriterWrite0(writer, bytes) {
	// What is written to is whatever the writer was given, which may be reached
	// through an interface or through a pointer, so the method of it is looked up
	// the way a writer's method is looked up anywhere else.
	const write = go2jsWriterMethod(writer.output, "Write");

	if (write === null) {
		throw new TypeError("tabwriter: output does not implement Write");
	}

	const written = write(bytes);

	if (typeof written === "number") {
		if (written < bytes.length) {
			panic("short write");
		}

		return;
	}

	if (Array.isArray(written) && written.length === 2) {
		const fault = written[1];

		if (fault !== null && fault !== undefined) {
			throw fault;
		}

		if (Number(written[0]) < bytes.length) {
			panic("short write");
		}
	}
}

// go2jsTabwriterWriteN is the first n bytes of some text written out, which is
// the padding of a cell written one small piece at a time.
function go2jsTabwriterWriteN(writer, src, n) {
	let left = n;

	while (left > src.length) {
		go2jsTabwriterWrite0(writer, src);
		left -= src.length;
	}

	go2jsTabwriterWrite0(writer, src.slice(0, left));
}

// go2jsTabwriterWritePadding is the room left beside the text of a cell, which is
// written with tabs when tabs are what a cell is padded with.
function go2jsTabwriterWritePadding(writer, textw, cellw, useTabs) {
	if (writer.padchar === 9 || useTabs) {
		if (writer.tabwidth === 0) {
			return;
		}

		cellw = Math.ceil(cellw / writer.tabwidth) * writer.tabwidth;

		const wanted = cellw - textw;

		if (wanted < 0) {
			panic("tabwriter: internal error");
		}

		go2jsTabwriterWriteN(writer, go2jsTabwriterTabs, Math.ceil(wanted / writer.tabwidth));

		return;
	}

	go2jsTabwriterWriteN(writer, [writer.padchar], cellw - textw);
}

// go2jsTabwriterWriteLines is lines written out as they stand, which is what a
// line of which nothing more is known looks like, and answers where in the text
// kept the lines it wrote leave off.
function go2jsTabwriterWriteLines(writer, pos, line0, line1) {
	for (let i = line0; i < line1; i++) {
		const line = writer.lines[i];
		let useTabs = (writer.flags & go2jsTabwriterTabIndent) !== 0;

		for (let j = 0; j < line.length; j++) {
			const cell = line[j];

			if (j > 0 && (writer.flags & go2jsTabwriterDebug) !== 0) {
				go2jsTabwriterWrite0(writer, [124]);
			}

			if (cell.size === 0) {
				if (j < writer.widths.length) {
					go2jsTabwriterWritePadding(writer, cell.width, writer.widths[j], useTabs);
				}
			} else {
				useTabs = false;

				if ((writer.flags & go2jsTabwriterAlignRight) === 0) {
					go2jsTabwriterWrite0(writer, writer.buf.slice(pos, pos + cell.size));
					pos += cell.size;

					if (j < writer.widths.length) {
						go2jsTabwriterWritePadding(writer, cell.width, writer.widths[j], false);
					}
				} else {
					if (j < writer.widths.length) {
						go2jsTabwriterWritePadding(writer, cell.width, writer.widths[j], false);
					}

					go2jsTabwriterWrite0(writer, writer.buf.slice(pos, pos + cell.size));
					pos += cell.size;
				}
			}
		}

		if (i + 1 === writer.lines.length) {
			go2jsTabwriterWrite0(writer, writer.buf.slice(pos, pos + writer.cell.size));
			pos += writer.cell.size;
		} else {
			go2jsTabwriterWrite0(writer, [10]);
		}
	}

	return pos;
}

// go2jsTabwriterFormat is the lines written out aligned, one column at a time: the
// width of a column is the width of the widest cell of it beside the room to leave
// around it, and a column holds the lines that reach as far as it does.
function go2jsTabwriterFormat(writer, pos, line0, line1) {
	const column = writer.widths.length;

	for (let at = line0; at < line1; at++) {
		if (column >= writer.lines[at].length - 1) {
			continue;
		}

		pos = go2jsTabwriterWriteLines(writer, pos, line0, at);
		line0 = at;

		let width = writer.minwidth;
		let discardable = true;

		for (; at < line1; at++) {
			const line = writer.lines[at];

			if (column >= line.length - 1) {
				break;
			}

			const cell = line[column];

			if (cell.width + writer.padding > width) {
				width = cell.width + writer.padding;
			}

			if (cell.width > 0 || cell.htab) {
				discardable = false;
			}
		}

		if (discardable && (writer.flags & go2jsTabwriterDiscardEmptyColumns) !== 0) {
			width = 0;
		}

		writer.widths.push(width);
		pos = go2jsTabwriterFormat(writer, pos, line0, at);
		writer.widths.pop();
		line0 = at;
	}

	return go2jsTabwriterWriteLines(writer, pos, line0, line1);
}

// go2jsTabwriterFlushNoDefers is what a writer holds written out, which is the
// whole of what it has been given: a writer is asked for it by name rather than
// once more for each line, since a line may be waiting on the ones after it.
function go2jsTabwriterFlushNoDefers(writer) {
	if (writer.cell.size > 0) {
		if (writer.endChar !== 0) {
			go2jsTabwriterEndEscape(writer);
		}

		go2jsTabwriterTerminateCell(writer, false);
	}

	go2jsTabwriterFormat(writer, 0, 0, writer.lines.length);
	go2jsTabwriterReset(writer);
}

// go2jsTabwriterFlush is Flush: everything the writer holds goes out aligned, and
// the writer is empty afterwards, as though nothing had been written to it.
function go2jsTabwriterFlush(writer) {
	go2jsTabwriterFlushNoDefers(writer);

	return null;
}

// go2jsTabwriterWrite is Write: the text given is cut into cells at the tabs and
// the line breaks in it, kept, and answered in the number of bytes taken rather
// than in what is written out, which waits for the columns to be settled.
function go2jsTabwriterWrite(writer, buf) {
	const bytes = Array.from(go2jsToArray(buf), item => Number(item) & 255);

	let n = 0;

	for (let i = 0; i < bytes.length; i++) {
		const ch = bytes[i];

		if (writer.endChar === 0) {
			if (ch === 9 || ch === 11 || ch === 10 || ch === 12) {
				go2jsTabwriterAppend(writer, bytes.slice(n, i));
				go2jsTabwriterUpdateWidth(writer);
				n = i + 1;

				const ncells = go2jsTabwriterTerminateCell(writer, ch === 9);

				if (ch === 10 || ch === 12) {
					go2jsTabwriterAddLine(writer);

					// A line of one cell says nothing of the columns to come, and
					// a form feed ends the columns of the line it ends, so either
					// of them settles what has been written so far.
					if (ch === 12 || ncells === 1) {
						go2jsTabwriterFlushNoDefers(writer);

						if (ch === 12 && (writer.flags & go2jsTabwriterDebug) !== 0) {
							go2jsTabwriterWrite0(writer, [45, 45, 45, 10]);
						}
					}
				}
			} else if (ch === go2jsTabwriterEscape) {
				go2jsTabwriterAppend(writer, bytes.slice(n, i));
				go2jsTabwriterUpdateWidth(writer);
				n = i;

				if ((writer.flags & go2jsTabwriterStripEscape) !== 0) {
					n++;
				}

				go2jsTabwriterStartEscape(writer, go2jsTabwriterEscape);
			} else if ((ch === 60 || ch === 38) && (writer.flags & go2jsTabwriterFilterHTML) !== 0) {
				go2jsTabwriterAppend(writer, bytes.slice(n, i));
				go2jsTabwriterUpdateWidth(writer);
				n = i;
				go2jsTabwriterStartEscape(writer, ch);
			}
		} else if (ch === writer.endChar) {
			let j = i + 1;

			if (ch === go2jsTabwriterEscape && (writer.flags & go2jsTabwriterStripEscape) !== 0) {
				j = i;
			}

			go2jsTabwriterAppend(writer, bytes.slice(n, j));
			n = i + 1;
			go2jsTabwriterEndEscape(writer);
		}
	}

	go2jsTabwriterAppend(writer, bytes.slice(n));

	return [bytes.length, null];
}
`
}

// tabwriterShimTypes are the names a program reaches the writer through, which are
// the type of it and the methods of that type.
