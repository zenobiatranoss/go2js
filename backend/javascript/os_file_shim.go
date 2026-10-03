package javascript

var osFileFuncs = map[string]string{
	"Open":       "go2jsOSOpen",
	"OpenFile":   "go2jsOSOpenFile",
	"Create":     "go2jsOSCreate",
	"ReadFile":   "go2jsOSReadFile",
	"WriteFile":  "go2jsOSWriteFile",
	"Mkdir":      "go2jsOSMkdir",
	"MkdirAll":   "go2jsOSMkdirAll",
	"Remove":     "go2jsOSRemove",
	"RemoveAll":  "go2jsOSRemoveAll",
	"Rename":     "go2jsOSRename",
	"CreateTemp": "go2jsOSCreateTemp",
	"MkdirTemp":  "go2jsOSMkdirTemp",
	"TempDir":    "go2jsOSTempDir",
	"Environ":    "go2jsOSEnviron",
	"Getenv":     "go2jsOSGetenv",
}

var osFileConstants = map[string]string{
	"O_RDONLY": "0",
	"O_WRONLY": "1",
	"O_RDWR":   "2",
	"O_APPEND": "1024",
	"O_CREATE": "64",
	"O_EXCL":   "128",
	"O_SYNC":   "1052672",
	"O_TRUNC":  "512",
}

var osFileMethods = map[string]string{
	"os.File.Read":        "go2jsOSFileRead",
	"os.File.ReadAt":      "go2jsOSFileReadAt",
	"os.File.ReadFile":    "go2jsOSFileReadFile",
	"os.File.Seek":        "go2jsOSFileSeek",
	"os.File.Sync":        "go2jsOSFileSync",
	"os.File.Write":       "go2jsOSFileWrite",
	"os.File.WriteAt":     "go2jsOSFileWriteAt",
	"os.File.WriteString": "go2jsOSFileWriteString",
	"os.File.Name":        "go2jsOSFileName",
	"os.File.Close":       "go2jsOSFileClose",
	"os.File.Fd":          "go2jsOSFileFd",
	"os.File.Stat":        "go2jsOSFileStat",
}

func init() {
	extendedStdlibFuncs()
	moreStdlibFuncs()

	functions := make(map[string]string, len(osFileFuncs)+len(osFileConstants))

	for name, value := range osFileFuncs {
		functions[name] = value
		shimValueMethods["os."+name] = value
	}

	for name, value := range osFileConstants {
		functions[name] = ""
		packageConstants["os."+name] = value
	}

	supportedStdlibPackages["os"] = funcSet(functions)

	for name, value := range osFileMethods {
		shimValueMethods[name] = value
	}
}

func osFileRuntimeSource() string {
	return `// An errno the host reports is said the way Go says it, rather than in the
// wording of the host, which names the same condition a different way.
const go2jsErrnoMessages = new Map([
	["ENOENT", "no such file or directory"],
	["EEXIST", "file exists"],
	["EACCES", "permission denied"],
	["EPERM", "operation not permitted"],
	["EISDIR", "is a directory"],
	["ENOTDIR", "not a directory"],
	["ENOTEMPTY", "directory not empty"],
	["EMFILE", "too many open files"],
	["ELOOP", "too many levels of symbolic links"]
]);

// An error the host reports is written as an error of a path: what was being
// done, which path, and the condition that ended it, which is what a Go program
// reading one is given.
function go2jsOSHostError(err, syscall, path) {
	if (err === null || err === undefined) {
		return null;
	}

	const named = go2jsErrnoMessages.get(err.code);

	if (named === undefined) {
		return err;
	}

	return go2jsNameError(go2jsOSError(named, syscall, path), "*fs.PathError");
}

// An error from a file is told by what was being done to which path, which is
// how one error is told from another that reads the same.
function go2jsOSError(message, syscall, path, target) {
	const text = String(message);
	const notExist = /no such file or directory|not exist|ENOENT/i.test(text);
	const error = new Error(String(syscall) + " " + String(path) + ": " + text);

	error.__go2js_errno = true;
	error.code = notExist ? "ENOENT" : text;
	error.syscall = syscall;
	error.path = String(path);
	error.errno = notExist ? 2 : 0;

	// An error of a path carries the errno that ended it, which is what a chain
	// unwraps to and what tells errors.Is that the condition it stands for is
	// the one a sentinel of the os package names.
	const errno = go2jsErrnoSentinel(error.code);

	if (errno !== null) {
		error.cause = errno;
	}

	return error;
}

function go2jsOSWrapFile(fd, path, appending) {
	// A file is a writer and a reader before it is anything else, and it says
	// so here rather than only through the method of its own type, because what
	// writes to a file is rarely written as that type.
	//
	// Where reading and writing stand is kept on the file rather than left to
	// the descriptor underneath, because a read at a place is a read from that
	// place and the two have to agree with each other.
	const file = {
		__go2js_osfile: true,
		fd: fd,
		path: String(path),
		closed: false,
		position: 0,
		append: appending === true
	};

	// The methods a file answers are on the file itself rather than reached
	// through the file type, so that a file handed to something that takes a
	// reader or a writer is read from and written to the way it is directly.
	file.Read = function (buffer) {
		return go2jsOSFileRead(file, buffer);
	};

	file.ReadAt = function (buffer, offset) {
		return go2jsOSFileReadAt(file, buffer, offset);
	};

	file.Seek = function (offset, whence) {
		return go2jsOSFileSeek(file, offset, whence);
	};

	file.Sync = function () {
		return go2jsOSFileSync(file);
	};

	file.Stat = function () {
		return go2jsOSFileStat(file);
	};

	file.Write = function (buffer) {
		return go2jsOSFileWrite(file, buffer);
	};

	file.WriteAt = function (buffer, offset) {
		return go2jsOSFileWriteAt(file, buffer, offset);
	};

	file.WriteString = function (text) {
		return go2jsOSFileWriteString(file, text);
	};

	file.WriteByte = function (value) {
		go2jsOSFileWrite(file, [value & 255]);

		return null;
	};

	// A reader that is given a file reads from where the file is, which is what
	// the whole of what is left of it says.
	file.__go2js_readAll = function () {
		return go2jsOSFileReadAll(file);
	};

	return file;
}

// A file that was never opened is not a file, and everything asked of it is
// refused the same way Go refuses it: with the error that stands for an
// argument that is not one.
function go2jsOSFileInvalid() {
	const error = new Error("invalid argument");

	error.__go2js_errno = true;
	error.errno = 0;

	return error;
}

function go2jsOSFileOf(value) {
	if (value !== null && value !== undefined && value.__go2js_osfile === true) {
		return value;
	}

	return null;
}

function go2jsOSOpenFile(path, flags, perm) {
	const fs = require("fs");
	const target = String(path);
	const access = flags & 3;
	const append = (flags & 1024) !== 0;
	const truncate = (flags & 512) !== 0;
	const create = (flags & 64) !== 0;
	const exclusive = (flags & 128) !== 0;
	let mode;

	if (access === 0) {
		mode = "r";
	} else if (append) {
		mode = access === 2 ? "a+" : "a";
	} else if (truncate) {
		mode = access === 2 ? "w+" : "w";
	} else {
		mode = "r+";
	}

	if (exclusive && create && access !== 0) {
		mode += "x";
	}

	try {
		const handle = fs.openSync(target, mode, perm);

		return [go2jsOSWrapFile(handle, target, append), null];
	} catch (err) {
		if (err.code === "ENOENT" && create && access !== 0) {
			try {
				return [go2jsOSWrapFile(fs.openSync(target, access === 2 ? "w+" : "w", perm), target, append), null];
			} catch (retry) {
				return [null, go2jsOSError(String(retry.message), "open", target)];
			}
		}

		return [null, go2jsOSHostError(err, "open", target)];
	}
}

function go2jsOSOpen(path) {
	return go2jsOSOpenFile(path, 0, 0);
}

function go2jsOSCreate(path) {
	const fs = require("fs");
	const target = String(path);

	try {
		return [go2jsOSWrapFile(fs.openSync(target, "w"), target), null];
	} catch (err) {
		return [null, go2jsOSError(String(err.message), "open", target)];
	}
}

function go2jsOSFileStream(value) {
	if (value === process.stdout) {
		return process.stdout;
	}

	if (value === process.stderr) {
		return process.stderr;
	}

	return null;
}

function go2jsOSFileWrite(file, buffer) {
	const value = go2jsUnwrap(file);
	const stream = go2jsOSFileStream(value);
	const bytes = go2jsToArray(buffer);

	if (stream !== null) {
		stream.write(go2jsBytesToString(buffer));
		return [bytes.length, null];
	}

	const target = go2jsOSFileOf(value);

	if (target === null) {
		// A file that is not there is refused, where the streams of the process
		// take what they are given when no file was named at all.
		if (file !== null && file !== undefined) {
			return [0, go2jsOSFileInvalid()];
		}

		go2jsWriteOut(process.stdout, go2jsBytesToString(buffer));

		return [bytes.length, null];
	}

	const result = go2jsOSFileWriteAt(target, buffer, target.position);

	// A write that went through moves the file along with it, and one that did
	// not leaves it where it was.
	if (result[1] === null) {
		target.position += bytes.length;
	}

	return result;
}

function go2jsOSFileSync(file) {
	const target = go2jsOSFileOf(go2jsUnwrap(file));

	if (target === null) {
		return go2jsOSFileInvalid();
	}

	try {
		require("fs").fsyncSync(target.fd);
	} catch (err) {
		return go2jsOSError(String(err.message), "fsync", target.path);
	}

	return null;
}

function go2jsOSFileWriteString(file, value) {
	return go2jsOSFileWrite(file, go2jsStringToBytes(value));
}

function go2jsOSFileName(file) {
	const value = go2jsUnwrap(file);

	if (value === process.stdout) {
		return "/dev/stdout";
	}

	if (value === process.stderr) {
		return "/dev/stderr";
	}

	if (value === process.stdin) {
		return "/dev/stdin";
	}

	const target = go2jsOSFileOf(value);

	return target === null ? "/dev/stdout" : target.path;
}

function go2jsOSFileFd(file) {
	const target = go2jsOSFileOf(go2jsUnwrap(file));

	return target === null ? 1 : target.fd;
}

function go2jsOSFileClose(file) {
	const target = go2jsOSFileOf(go2jsUnwrap(file));

	if (target === null) {
		return go2jsOSFileInvalid();
	}

	if (target.closed) {
		return go2jsOSError("file already closed", "close", target.path);
	}

	target.closed = true;

	try {
		require("fs").closeSync(target.fd);
	} catch (err) {
		return go2jsOSError(String(err.message), "close", target.path);
	}

	return null;
}

// go2jsOSFileReadAt reads into the bytes the program gave at a place, and says
// how many of them were filled. A file that has nothing left hands back no bytes
// at all, which is what tells a read at the end of a file from one that filled
// what it was given.
function go2jsOSFileReadAt(file, buffer, offset) {
	const target = go2jsOSFileOf(go2jsUnwrap(file));

	if (target === null || target.closed === true) {
		return [0, go2jsOSFileInvalid()];
	}

	const size = buffer === undefined || buffer === null ? 512 : go2jsLen(buffer);

	if (size === 0) {
		return [0, null];
	}

	const chunk = Buffer.alloc(size);

	let read = 0;

	try {
		read = require("fs").readSync(target.fd, chunk, 0, size, Number(offset));
	} catch (err) {
		return [0, go2jsOSHostError(err, "read", target.path)];
	}

	for (let index = 0; index < read; index++) {
		go2jsIndexSet(buffer, index, chunk[index]);
	}

	// A file that has nothing left is the end of the text rather than a failure
	// of a read, which is the one thing a read reports it as.
	return [read, read === 0 ? go2jsIOEOF() : null];
}

function go2jsOSFileRead(file, buffer) {
	const target = go2jsOSFileOf(go2jsUnwrap(file));

	if (target === null || target.closed === true) {
		return [0, go2jsOSFileInvalid()];
	}

	const result = go2jsOSFileReadAt(target, buffer, target.position);

	target.position += result[0];

	return result;
}

// go2jsOSStreamFd is the descriptor one of the streams of the process stands
// for, which is what a seek or a stat of one is carried out on.
function go2jsOSStreamFd(value) {
	if (value === process.stdin) {
		return 0;
	}

	if (value === process.stdout) {
		return 1;
	}

	if (value === process.stderr) {
		return 2;
	}

	return null;
}

// go2jsOSStreamPosition is where reading one of the streams of the process
// stands, which is kept on the stream rather than left to the descriptor
// underneath, so that a seek of one can put it back.
function go2jsOSStreamPosition(stream) {
	return typeof stream.__go2js_position === "number" ? stream.__go2js_position : 0;
}

function go2jsOSStreamSetPosition(stream, position) {
	stream.__go2js_position = position;
}

// A file is put at a place, which is where reading and writing after it stand
// and is what a write at a place of its own leaves alone.
function go2jsOSFileSeek(file, offset, whence) {
	const value = go2jsUnwrap(file);
	const target = go2jsOSFileOf(value);

	if (target !== null && target.closed === true) {
		return [0, go2jsOSFileInvalid()];
	}

	if (target === null) {
		const fd = go2jsOSStreamFd(value);

		if (fd === null) {
			return [0, go2jsOSFileInvalid()];
		}

		const where = Number(whence) || 0;
		let stream = go2jsOSStreamPosition(value);

		if (where === 1) {
			stream += Number(offset);
		} else if (where === 2) {
			try {
				stream = require("fs").fstatSync(fd).size + Number(offset);
			} catch (err) {
				return [0, go2jsOSHostError(err, "seek", go2jsOSFileName(value))];
			}
		} else if (where !== 0) {
			return [0, go2jsOSError("invalid whence", "seek", go2jsOSFileName(value))];
		} else {
			stream = Number(offset);
		}

		if (stream < 0) {
			return [0, go2jsOSError("invalid argument", "seek", go2jsOSFileName(value))];
		}

		go2jsOSStreamSetPosition(value, stream);

		return [stream, null];
	}

	const from = Number(whence) || 0;
	let position = 0;

	if (from === 0) {
		position = Number(offset);
	} else if (from === 1) {
		position = target.position + Number(offset);
	} else if (from === 2) {
		try {
			position = require("fs").fstatSync(target.fd).size + Number(offset);
		} catch (err) {
			return [0, go2jsOSHostError(err, "seek", target.path)];
		}
	} else {
		return [0, go2jsOSError("invalid whence", "seek", target.path)];
	}

	if (position < 0) {
		return [0, go2jsOSError("invalid argument", "seek", target.path)];
	}

	target.position = position;

	return [position, null];
}

// A write at a place leaves where the file stands alone, which is what tells it
// from a write that moves it along.
function go2jsOSFileWriteAt(file, buffer, offset) {
	const target = go2jsOSFileOf(go2jsUnwrap(file));

	if (target === null || target.closed === true) {
		return [0, go2jsOSFileInvalid()];
	}

	const bytes = go2jsToArray(buffer);
	const at = Number(offset);

	try {
		require("fs").writeSync(target.fd, Buffer.from(bytes), 0, bytes.length, target.append ? null : at);
	} catch (err) {
		return [0, go2jsOSHostError(err, "write", target.path)];
	}

	return [bytes.length, null];
}

function go2jsOSFileReadFile(file) {
	const target = go2jsOSFileOf(go2jsUnwrap(file));

	if (target === null) {
		return [null, go2jsOSFileInvalid()];
	}

	try {
		return [go2jsStringToBytes(require("fs").readFileSync(target.fd, "utf8")), null];
	} catch (err) {
		return [null, go2jsOSError(String(err.message), "read", target.path)];
	}
}

function go2jsOSFileReadAll(file) {
	const target = go2jsOSFileOf(go2jsUnwrap(file));

	if (target === null) {
		return "";
	}


	const fs = require("fs");
	const parts = [];
	const chunk = Buffer.alloc(4096);

	for (;;) {
		const read = fs.readSync(target.fd, chunk, 0, chunk.length, null);

		if (read === 0) {
			break;
		}

		parts.push(Buffer.from(chunk.subarray(0, read)).toString("utf8"));
	}

	return parts.join("");
}

// go2jsOSFileInfo is what a stat answers with: a set of methods rather than the
// fields the host happens to keep the same information in, so that a program
// asks a file what it wants to know the way it asks one in Go.
function go2jsOSFileInfo(path, stats) {
	const text = String(path);
	const info = {
		__go2js_type: "*os.fileStat",
		Name: function () {
			return text.slice(text.lastIndexOf("/") + 1);
		},
		Size: function () {
			return Number(stats.size);
		},
		IsDir: function () {
			return typeof stats.isDirectory === "function" ? stats.isDirectory() : false;
		},
		Mode: function () {
			return stats.mode === undefined ? 0 : stats.mode;
		},
		ModTime: function () {
			const when = stats.mtime instanceof Date ? stats.mtime : new Date(Number(stats.mtime));

			return go2jsTimeValue(isNaN(when.getTime()) ? new Date(0) : when);
		}
	};

	return info;
}

function go2jsOSFileStat(file) {
	const target = go2jsOSFileOf(go2jsUnwrap(file));
	const path = target === null ? "/dev/stdout" : target.path;

	try {
		return [go2jsOSFileInfo(path, require("fs").fstatSync(target === null ? 1 : target.fd)), null];
	} catch (err) {
		return [null, go2jsOSHostError(err, "stat", path)];
	}
}

function go2jsOSReadFile(path) {
	try {
		return [go2jsStringToBytes(require("fs").readFileSync(String(path), "utf8")), null];
	} catch (err) {
		return [null, go2jsOSHostError(err, "open", String(path))];
	}
}

function go2jsOSWriteFile(path, data, perm) {
	try {
		require("fs").writeFileSync(String(path), go2jsBytesToString(data));
	} catch (err) {
		return go2jsOSError(String(err.message), "open", String(path));
	}

	return null;
}

function go2jsOSMkdir(path, perm) {
	try {
		require("fs").mkdirSync(String(path));
	} catch (err) {
		return go2jsOSHostError(err, "mkdir", String(path));
	}

	return null;
}

function go2jsOSMkdirAll(path, perm) {
	try {
		require("fs").mkdirSync(String(path), {recursive: true});
	} catch (err) {
		return go2jsOSError(String(err.message), "mkdir", String(path));
	}

	return null;
}

function go2jsOSRemove(path) {
	try {
		require("fs").unlinkSync(String(path));
	} catch (err) {
		return go2jsOSHostError(err, "remove", String(path));
	}

	return null;
}

function go2jsOSRemoveAll(path) {
	try {
		require("fs").rmSync(String(path), {recursive: true, force: true});
	} catch (err) {
		return go2jsOSError(String(err.message), "remove", String(path));
	}

	return null;
}

function go2jsOSRename(from, to) {
	try {
		require("fs").renameSync(String(from), String(to));
	} catch (err) {
		return go2jsOSError(String(err.message), "rename", String(from));
	}

	return null;
}

function go2jsOSTempDir() {
	return require("os").tmpdir();
}

function go2jsOSTempCandidate(dir, pattern, attempt) {
	const base = dir === undefined || dir === null || dir === "" ? go2jsOSTempDir() : String(dir);
	const wanted = pattern === undefined || pattern === "" ? "*" : String(pattern);
	const suffix = String(Date.now()) + String(process.pid) + String(attempt);
	const name = wanted.includes("*") ? wanted.replace("*", suffix) : wanted + suffix;

	return base.replace(/\/+$/, "") + "/" + name;
}

function go2jsOSCreateTemp(dir, pattern) {
	const fs = require("fs");

	for (let attempt = 0; attempt < 100; attempt++) {
		const target = go2jsOSTempCandidate(dir, pattern, attempt);

		try {
			return [go2jsOSWrapFile(fs.openSync(target, "wx"), target), null];
		} catch (err) {
			if (err.code !== "EEXIST") {
				return [null, go2jsOSError(String(err.message), "open", target)];
			}
		}
	}

	return [null, go2jsOSError("cannot create temporary file", "open", String(dir))];
}

function go2jsOSMkdirTemp(dir, pattern) {
	const fs = require("fs");

	for (let attempt = 0; attempt < 100; attempt++) {
		const target = go2jsOSTempCandidate(dir, pattern, attempt);

		try {
			fs.mkdirSync(target);

			return [target, null];
		} catch (err) {
			if (err.code !== "EEXIST") {
				return [null, go2jsOSError(String(err.message), "mkdir", target)];
			}
		}
	}

	return [null, go2jsOSError("cannot create temporary directory", "mkdir", String(dir))];
}
`
}
