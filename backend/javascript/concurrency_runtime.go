package javascript

func concurrencyRuntimeSource() string {
	return `
function go2jsChannel(capacity) {
	return {
		__go2js_channel: true,
		buffer: [],
		capacity: typeof capacity === "number" && capacity > 0 ? Math.trunc(capacity) : 0,
		closed: false,
		receivers: [],
		senders: []
	};
}

function go2jsChanLen(channel) {
	return channel.buffer.length;
}

function go2jsChanCap(channel) {
	return channel.capacity;
}

function go2jsChannelClosed(channel) {
	return channel.closed && channel.buffer.length === 0;
}

function go2jsChannelClosedPanic(operation) {
	throw new Error("send on closed channel");
}

function go2jsChanRecvPair(channel) {
	while (true) {
		if (channel.buffer.length > 0) {
			const value = channel.buffer.shift();
			go2jsChannelPump(channel);
			return [value, true];
		}

		if (channel.closed) {
			return [go2jsChannelZero, false];
		}

		if (!go2jsProgress()) {
			throw new Error("go2js: no goroutine can unblock this channel receive");
		}
	}
}

function go2jsChanRecv(channel) {
	return go2jsChanRecvPair(channel)[0];
}

function go2jsChanTryRecv(channel) {
	if (channel.buffer.length > 0) {
		const value = channel.buffer.shift();
		go2jsChannelPump(channel);
		return [value, true];
	}

	if (channel.closed) {
		return [go2jsChannelZero, false];
	}

	return null;
}

function go2jsChanSend(channel, value) {
	while (true) {
		if (channel.closed) {
			go2jsChannelClosedPanic("send");
		}

		if (channel.capacity === 0 || channel.buffer.length < channel.capacity) {
			channel.buffer.push(value);
			go2jsChannelPump(channel);
			return;
		}

		if (!go2jsProgress()) {
			throw new Error("go2js: no goroutine can unblock this channel send");
		}
	}
}

function go2jsChanTrySend(channel, value) {
	if (channel.closed) {
		go2jsChannelClosedPanic("send");
	}

	if (channel.capacity === 0) {
		return false;
	}

	if (channel.buffer.length >= channel.capacity) {
		return false;
	}

	channel.buffer.push(value);
	go2jsChannelPump(channel);
	return true;
}

function go2jsChannelPump(channel) {
	if (channel.buffer.length === 0) {
		return;
	}

	const receiver = channel.receivers.shift();

	if (receiver === undefined) {
		return;
	}

	receiver.resolve({ channel, value: channel.buffer.shift() });
}

function go2jsChannelClose(channel) {
	if (channel.closed) {
		throw new Error("close of closed channel");
	}

	channel.closed = true;

	const waiting = channel.receivers.splice(0, channel.receivers.length);

	for (const receiver of waiting) {
		receiver.resolve({ channel, value: go2jsChannelZero, closed: true });
	}

	const blocked = channel.senders.splice(0, channel.senders.length);

	for (const sender of blocked) {
		sender.reject(new Error("send on closed channel"));
	}
}

function go2jsChannelZero() {
	return undefined;
}

function go2jsChannelRange(channel) {
	return {
		[Symbol.iterator]() {
			return {
				next() {
					const result = go2jsChanTryRecv(channel);

					if (result === null) {
						if (channel.closed) {
							return { done: true, value: undefined };
						}
						return { done: false, value: undefined };
					}

					if (result[1] === false) {
						return { done: true, value: undefined };
					}

					return { done: false, value: result[0] };
				}
			};
		}
	};
}

function go2jsWaitGroup() {
	return {
		count: 0,
		waiters: []
	};
}

const go2jsTasks = [];
let go2jsDraining = false;

function go2jsWaitGroupAdd(group, delta) {
	group.count += delta;

	if (group.count < 0) {
		throw new Error("sync: negative WaitGroup counter");
	}

	if (group.count === 0) {
		const waiting = group.waiters.splice(0, group.waiters.length);

		for (const waiter of waiting) {
			waiter.resolve();
		}
	}
}

function go2jsWaitGroupDone(group) {
	go2jsWaitGroupAdd(group, -1);
}

function go2jsRunTasks() {
	if (go2jsDraining) {
		return 0;
	}

	go2jsDraining = true;

	let executed = 0;

	try {
		while (go2jsTasks.length > 0) {
			if (executed >= 1000000) {
				throw new Error("go2js: goroutine task limit exceeded");
			}

			const task = go2jsTasks.shift();
			task();
			executed++;
		}
	} finally {
		go2jsDraining = false;
	}

	return executed;
}

function go2jsGo(task) {
	go2jsTasks.push(task);
	queueMicrotask(() => {
		go2jsRunTasks();
	});
}

function go2jsProgress() {
	if (go2jsTasks.length === 0) {
		return false;
	}

	return go2jsRunTasks() > 0;
}

function go2jsWaitGroupWait(group) {
	let budget = 1000000;

	while (group.count > 0) {
		if (budget-- === 0) {
			throw new Error("go2js: wait group did not reach zero");
		}

		go2jsRunTasks();
	}
}

function go2jsMutex() {
	return { locked: false, waiters: [] };
}

function go2jsMutexLock(mutex) {
	if (!mutex.locked) {
		mutex.locked = true;
		return;
	}

	throw new Error("sync.Mutex: already locked");
}

function go2jsMutexUnlock(mutex) {
	if (!mutex.locked) {
		throw new Error("sync: unlock of unlocked mutex");
	}

	mutex.locked = false;

	const waiting = mutex.waiters.shift();

	if (waiting !== undefined) {
		mutex.locked = true;
		waiting.resolve();
	}
}

function go2jsRWMutex() {
	return { readers: 0, writer: false, readersWaiting: [], writersWaiting: [] };
}

function go2jsRWMutexRLock(mutex) {
	if (mutex.writer) {
		throw new Error("sync: RLock of write-locked mutex");
	}

	mutex.readers++;
}

function go2jsRWMutexRUnlock(mutex) {
	if (mutex.readers === 0) {
		throw new Error("sync: RUnlock of unlocked RWMutex");
	}

	mutex.readers--;

	if (mutex.readers === 0) {
		go2jsRWMutexPromote(mutex);
	}
}

function go2jsRWMutexLock(mutex) {
	if (mutex.writer || mutex.readers > 0) {
		throw new Error("sync: Lock already held");
	}

	mutex.writer = true;
}

function go2jsRWMutexUnlock(mutex) {
	if (!mutex.writer) {
		throw new Error("sync: unlock of unlocked RWMutex");
	}

	mutex.writer = false;
	go2jsRWMutexPromote(mutex);
}

function go2jsRWMutexPromote(mutex) {
	if (mutex.writer || mutex.readers > 0) {
		return;
	}

	if (mutex.writersWaiting.length > 0) {
		const writer = mutex.writersWaiting.shift();
		mutex.writer = true;
		writer.resolve();
		return;
	}

	while (mutex.readersWaiting.length > 0) {
		mutex.readers++;
		mutex.readersWaiting.shift().resolve();
	}
}

function go2jsOnce() {
	return { done: false };
}

function go2jsOnceDo(once, fn) {
	if (once.done) {
		return;
	}

	once.done = true;
	fn();
}

function go2jsSyncMap() {
	return { entries: new Map() };
}

function go2jsSyncMapStore(store, key, value) {
	store.entries.set(key, value);
}

function go2jsSyncMapLoad(store, key) {
	if (!store.entries.has(key)) {
		return [null, false];
	}

	return [store.entries.get(key), true];
}

function go2jsSyncMapLoadOrStore(store, key, value) {
	if (store.entries.has(key)) {
		return [store.entries.get(key), true];
	}

	store.entries.set(key, value);
	return [value, false];
}

function go2jsSyncMapLoadAndDelete(store, key) {
	if (!store.entries.has(key)) {
		return [null, false];
	}

	const value = store.entries.get(key);
	store.entries.delete(key);
	return [value, true];
}

function go2jsSyncMapDelete(store, key) {
	store.entries.delete(key);
}

function go2jsSyncMapSwap(store, key, value) {
	const previous = store.entries.get(key);
	store.entries.set(key, value);
	return previous;
}

function go2jsSyncMapCompareAndSwap(store, key, old, next) {
	if (!store.entries.has(key)) {
		if (old !== undefined && old !== null) {
			return false;
		}
	}

	if (store.entries.get(key) !== old) {
		return false;
	}

	store.entries.set(key, next);
	return true;
}

function go2jsSyncMapRange(store, fn) {
	for (const [key, value] of store.entries) {
		fn(key, value);
	}
}

function go2jsSyncMapLen(store) {
	return store.entries.size;
}

function go2jsAtomicCell(target) {
	if (target !== null && target !== undefined &&
		typeof target.get === "function" && typeof target.set === "function") {
		return {
			get value() {
				return target.get();
			},
			set value(next) {
				target.set(next);
			}
		};
	}

	return { value: target };
}

function go2jsAtomicLoad(cell) {
	return cell.value;
}

function go2jsAtomicStore(cell, value) {
	cell.value = value;
}

function go2jsAtomicAdd(cell, delta) {
	cell.value = cell.value + delta;
	return cell.value;
}

function go2jsAtomicSwap(cell, value) {
	const previous = cell.value;
	cell.value = value;
	return previous;
}

function go2jsAtomicCompareAndSwap(cell, expected, next) {
	if (cell.value !== expected) {
		return false;
	}

	cell.value = next;
	return true;
}

function go2jsErrorsIs(err, target) {
	if (err === target) {
		return true;
	}

	if (err === null || err === undefined || target === null || target === undefined) {
		return err === target;
	}

	if (err instanceof Error && target instanceof Error && err.message === target.message) {
		return true;
	}

	if (err === null || err === undefined || err.cause === undefined) {
		return false;
	}

	return go2jsErrorsIs(err.cause, target);
}

function go2jsWrapError(format, ...args) {
	const error = new Error(go2jsSprintf(format, ...args));

	for (let i = 0; i < args.length; i++) {
		if (format.includes("%w")) {
			const verbs = format.match(/%[+#0 -.]*[0-9.]*[a-zA-Z]/g) || [];
			const position = verbs.indexOf("%w");

			if (position === i) {
				error.cause = args[i];
				break;
			}
		}
	}

	return error;
}

function go2jsErrorsUnwrap(err) {
	if (err === null || err === undefined) {
		return null;
	}

	if (err.cause !== undefined) {
		return err.cause;
	}

	return null;
}

function go2jsErrorsAs(err, target) {
	if (typeof target !== "function") {
		return false;
	}

	return err instanceof target;
}

function go2jsErrorsJoin(errs) {
	if (!Array.isArray(errs) || errs.length === 0) {
		return null;
	}

	const parts = errs.filter(item => item !== null && item !== undefined);

	if (parts.length === 0) {
		return null;
	}

	if (parts.length === 1) {
		return parts[0];
	}

	return new Error(parts.map(part => part.message).join("\n"));
}

function go2jsContextBackground() {
	return {
		__go2js_context: true,
		done: false,
		err: null,
		value: undefined
	};
}

function go2jsDuration(nanoseconds) {
	return {
		nanoseconds: nanoseconds,
		String() {
			return go2jsDurationString(this.nanoseconds);
		}
	};
}

function go2jsDurationDecimal(value) {
	if (Number.isInteger(value)) {
		return String(value);
	}

	let text = value.toFixed(9);

	while (text.endsWith("0")) {
		text = text.slice(0, -1);
	}

	if (text.endsWith(".")) {
		text = text.slice(0, -1);
	}

	return text;
}

function go2jsDurationString(nanoseconds) {
	if (nanoseconds === 0) {
		return "0s";
	}

	const sign = nanoseconds < 0 ? "-" : "";
	let rest = Math.abs(nanoseconds);

	if (rest < 1000000000) {
		let scale = 1;
		let unit = "ns";

		if (rest >= 1000000) {
			scale = 1000000;
			unit = "ms";
		} else if (rest >= 1000) {
			scale = 1000;
			unit = "\u00b5s";
		}

		return sign + go2jsDurationDecimal(rest / scale) + unit;
	}

	const hours = Math.floor(rest / 3600000000000);
	rest -= hours * 3600000000000;

	const minutes = Math.floor(rest / 60000000000);
	rest -= minutes * 60000000000;

	let text = "";

	if (hours > 0) {
		text += hours + "h";
	}

	if (hours > 0 || minutes > 0) {
		text += minutes + "m";
	}

	return sign + text + go2jsDurationDecimal(rest / 1000000000) + "s";
}

function go2jsErrorString(value) {
	if (value === null || value === undefined) {
		return "<nil>";
	}

	if (value instanceof Error) {
		return value.message;
	}

	if (value.__go2js_interface === true) {
		return go2jsErrorString(value.value);
	}

	if (typeof value === "object" && typeof value.Error === "function") {
		return value.Error();
	}

	return go2jsStringify(value);
}

function go2jsStdoutWrite(values, suffix) {
	process.stdout.write(go2jsOutputText(values, suffix));
}

function go2jsStderrWrite(values, suffix) {
	process.stderr.write(go2jsOutputText(values, suffix));
}

function go2jsOutputText(value, suffix) {
	if (value === undefined) {
		return "";
	}

	if (Array.isArray(value)) {
		return value.map(item => go2jsErrorString(item)).join("") + suffix;
	}

	return go2jsErrorString(value) + suffix;
}

function go2jsExit(code) {
	process.exit(code === undefined ? 0 : Number(code));
}
`
}
