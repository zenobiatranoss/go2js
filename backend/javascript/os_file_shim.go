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
	"os.File.ReadFile":    "go2jsOSFileReadFile",
	"os.File.Sync":        "go2jsOSFileSync",
	"os.File.Write":       "go2jsOSFileWrite",
	"os.File.WriteString": "go2jsOSFileWriteString",
	"os.File.Name":        "go2jsOSFileName",
	"os.File.Close":       "go2jsOSFileClose",
	"os.File.Fd":          "go2jsOSFileFd",
	"os.File.Stat":        "go2jsOSFileStat",
}

var osFileMultiReturn = map[string]bool{
	"os.File.Read": true,
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
	return `function go2jsOSError(message, syscall, path, target) {
	const text = String(message);
	const notExist = /no such file or directory|not exist|ENOENT/i.test(text);
	const error = new Error("open " + String(path) + ": " + text);

	error.__go2js_errno = true;
	error.code = notExist ? "ENOENT" : text;
	error.syscall = syscall;
	error.path = String(path);
	error.errno = notExist ? 2 : 0;

	return error;
}

function go2jsOSWrapFile(fd, path) {
	return {__go2js_osfile: true, fd: fd, path: String(path), closed: false};
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

		return [go2jsOSWrapFile(handle, target), null];
	} catch (err) {
		if (err.code === "ENOENT" && create && access !== 0) {
			try {
				return [go2jsOSWrapFile(fs.openSync(target, access === 2 ? "w+" : "w", perm), target), null];
			} catch (retry) {
				return [null, go2jsOSError(String(retry.message), "open", target)];
			}
		}

		return [null, go2jsOSError(err.code === "ENOENT" ? "no such file or directory" : String(err.message), "open", target)];
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
		process.stdout.write(go2jsBytesToString(buffer));
		return [bytes.length, null];
	}

	require("fs").writeSync(target.fd, Buffer.from(bytes));

	return [bytes.length, null];
}

function go2jsOSFileSync(file) {
	const target = go2jsOSFileOf(go2jsUnwrap(file));

	if (target === null) {
		return null;
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

	const target = go2jsOSFileOf(value);

	return target === null ? "/dev/stdout" : target.path;
}

function go2jsOSFileFd(file) {
	const target = go2jsOSFileOf(go2jsUnwrap(file));

	return target === null ? 1 : target.fd;
}

function go2jsOSFileClose(file) {
	const target = go2jsOSFileOf(go2jsUnwrap(file));

	if (target === null || target.closed) {
		return null;
	}

	target.closed = true;

	try {
		require("fs").closeSync(target.fd);
	} catch (err) {
		return go2jsOSError(String(err.message), "close", target.path);
	}

	return null;
}

function go2jsOSFileRead(file, buffer) {
	const target = go2jsOSFileOf(go2jsUnwrap(file));

	if (target === null) {
		return [0, null];
	}

	const size = buffer === undefined || buffer === null ? 512 : go2jsToArray(buffer).length;
	const chunk = Buffer.alloc(size);
	const read = require("fs").readSync(target.fd, chunk, 0, size, null);
	const bytes = Array.prototype.slice.call(chunk.subarray(0, read));
	const view = go2jsToArray(buffer);

	for (let index = 0; index < view.length; index++) {
		view[index] = bytes[index] === undefined ? 0 : bytes[index];
	}

	return [read, read === 0 ? go2jsOSError("EOF", "read", target.path) : null];
}

function go2jsOSFileReadFile(file) {
	const target = go2jsOSFileOf(go2jsUnwrap(file));

	if (target === null) {
		return [null, go2jsOSError("invalid file", "read", "/dev/stdin")];
	}

	try {
		return [go2jsStringToBytes(require("fs").readFileSync(target.fd, "utf8")), null];
	} catch (err) {
		return [null, go2jsOSError(String(err.message), "read", target.path)];
	}
}

function go2jsOSFileStat(file) {
	const target = go2jsOSFileOf(go2jsUnwrap(file));

	try {
		const stat = require("fs").fstatSync(target === null ? 1 : target.fd);
		const info = {
			name: target === null ? "/dev/stdout" : target.path,
			size: stat.size,
			isDir: stat.isDirectory()
		};

		return [info, null];
	} catch (err) {
		return [null, go2jsOSError(String(err.message), "stat", target === null ? "" : target.path)];
	}
}

function go2jsOSReadFile(path) {
	try {
		return [go2jsStringToBytes(require("fs").readFileSync(String(path), "utf8")), null];
	} catch (err) {
		return [null, go2jsOSError(err.code === "ENOENT" ? "no such file or directory" : String(err.message), "open", String(path))];
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
		return go2jsOSError(err.code === "EEXIST" ? "file exists" : String(err.message), "mkdir", String(path));
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
		return go2jsOSError(err.code === "ENOENT" ? "no such file or directory" : String(err.message), "remove", String(path));
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
