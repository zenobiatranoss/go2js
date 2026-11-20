package javascript

// syscall offers the two calls a program makes to signal itself: Getpid and
// Kill. A signal sent to the host process is passed on to it, except SIGCHLD,
// which Go ignores by default and so does nothing.
func syscallRuntimeSource() string {
	return `function go2jsSyscallGetpid() {
	return typeof process !== "undefined" ? process.pid : 1;
}

function go2jsSyscallKill(pid, signal) {
	const number = go2jsSignalNumberOf(signal);

	if (number === 17 || number === 0) {
		return null;
	}

	const name = go2jsSignalHostName(number);

	if (name !== null && typeof process !== "undefined") {
		process.kill(Number(pid), name);
	}

	return null;
}
`
}
