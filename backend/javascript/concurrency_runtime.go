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
		senders: [],
		waitingReceivers: 0
	};
}

// A channel that was never made is not a channel, but it is still a value: a
// send on it and a receive from it are both never ready, and a select that
// offers one of them takes whichever other case it was given. Only the
// operations that ask for an answer, such as the length of one, have to treat
// it as a channel of no capacity.
function go2jsChannelIsNil(channel) {
	return channel === null || channel === undefined;
}

function go2jsChanLen(channel) {
	return go2jsChannelIsNil(channel) ? 0 : channel.buffer.length;
}

function go2jsChanCap(channel) {
	return go2jsChannelIsNil(channel) ? 0 : channel.capacity;
}

function go2jsChannelClosed(channel) {
	return !go2jsChannelIsNil(channel) && channel.closed && channel.buffer.length === 0;
}

function go2jsChannelClosedPanic(operation) {
	throw new Error("send on closed channel");
}

// go2jsTimers holds the channels a timer is waiting on, so that a select with
// nothing else to do can wait for the moment one of them comes due rather than
// declaring every goroutine asleep.
const go2jsTimers = new Map();

// go2jsTimerDue says whether a timer has come due. A channel that stands for a
// timer is not ready until the moment it names, which is what lets a select
// choose the work that finished over the one that merely has a deadline.
function go2jsTimerDue(channel) {
	if (channel.timerDeadline === undefined) {
		return false;
	}

	if (Date.now() < channel.timerDeadline) {
		return false;
	}

	go2jsTimers.delete(channel);
	channel.timerDeadline = undefined;

	// A timer that runs a function has no value to send: the function is what
	// the timer was asked for, and the moment it fires at is not a result.
	if (typeof channel.timerCallback === "function") {
		const run = channel.timerCallback;

		delete channel.timerCallback;
		run();

		return true;
	}

	channel.buffer.push(go2jsTimeValue(new Date(channel.timerDue)));

	// A ticker sends again once its period is up, so the deadline it leaves
	// behind is the next one rather than none at all. A ticker that was stopped
	// has no period left and is done.
	if (channel.timerPeriod !== null && channel.timerPeriod !== undefined) {
		go2jsTimerArm(channel);
	}

	return true;
}

// go2jsTimerArm gives a channel a deadline a period on from now, in the two
// forms the schedule keeps it in: the moment a send carries, and the millis-
// second moment a select compares against the clock.
function go2jsTimerArm(channel) {
	channel.timerDeadline = Date.now() + channel.timerPeriod / 1000000;
	channel.timerDue = go2jsTimeValue(new Date(channel.timerDeadline));
	go2jsTimers.set(channel, channel.timerDue);
}

function go2jsWaitForEarliestTimer() {
	let earliest = Infinity;

	for (const channel of go2jsTimers.keys()) {
		if (channel.timerDeadline < earliest) {
			earliest = channel.timerDeadline;
		}
	}

	if (!isFinite(earliest)) {
		return false;
	}

	const wait = earliest - Date.now();

	if (wait > 0) {
		try {
			require("child_process").execFileSync("sleep", [String(wait / 1000)]);
		} catch (err) {
			// A wait that could not be taken still leaves the deadline to be
			// reached, and the loop comes back round to it.
		}
	}

	return true;
}

function go2jsChanRecvPair(channel) {
	// A receive that finds nothing waits, and a send to a channel of no room
	// is a handover that only completes once somebody has taken what was put
	// into it. The count of how many receives are waiting is what tells a send
	// there is somebody to hand over to, so it is kept for as long as the
	// waiting lasts and given back when the receive is answered or gives up.
	channel.waitingReceivers++;

	try {
		while (true) {
			if (channel.buffer.length > 0) {
				const value = channel.buffer.shift();
				go2jsChannelPump(channel);
				return [value, true];
			}

			if (go2jsTimerDue(channel)) {
				const value = channel.buffer.shift();
				go2jsChannelPump(channel);
				return [value, true];
			}

			if (channel.closed) {
				return [go2jsChannelZero, false];
			}

			if (!go2jsProgress() && !go2jsWaitForEarliestTimer()) {
				throw go2jsFatalError("all goroutines are asleep - deadlock!");
			}
		}
	} finally {
		channel.waitingReceivers--;
	}
}

function go2jsChanRecv(channel) {
	return go2jsChanRecvPair(channel)[0];
}

function go2jsChanTryRecv(channel) {
	if (go2jsChannelIsNil(channel)) {
		return null;
	}

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

function go2jsChanRecvReady(channel) {
	if (go2jsChannelIsNil(channel)) {
		return false;
	}

	return channel.buffer.length > 0 || channel.closed || go2jsTimerDue(channel);
}

function go2jsChanSendReady(channel) {
	if (go2jsChannelIsNil(channel) || channel.closed) {
		return false;
	}

	// A channel of no room is a handover rather than a store, so a send to one
	// is ready only while a goroutine is standing there to take it. Answering
	// yes for one nobody is reading from is what lets a loop of sends fill a
	// channel that can never hold anything.
	if (channel.capacity === 0) {
		return channel.waitingReceivers > 0;
	}

	return channel.buffer.length < channel.capacity;
}

function go2jsChanSend(channel, value) {
	while (true) {
		if (channel.closed) {
			go2jsChannelClosedPanic("send");
		}

				// A channel of no room is a handover: the value is not sent until the
		// goroutine that was waiting for it has taken it, and that goroutine
		// runs before the send is said to be done. A channel with room is not
		// a handover, so a send to one carries on without waiting for whoever
		// reads it later.
		if (channel.capacity === 0 || channel.buffer.length < channel.capacity) {
			channel.buffer.push(value);
			go2jsChannelPump(channel);

			if (channel.capacity === 0) {
				go2jsRunTasks();
			}

			return;
		}

		if (!go2jsProgress()) {
			throw go2jsFatalError("all goroutines are asleep - deadlock!");
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
					// Ranging over a channel is receiving from it until it is
					// closed, and a receive waits for whatever will send on it,
					// a goroutine or the moment a timer comes due.
					if (channel === null || channel === undefined) {
						throw go2jsFatalError("all goroutines are asleep - deadlock!");
					}

					const result = go2jsChanRecvPair(channel);

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

		try {
			task();
		} catch (thrown) {
			// A goroutine that was told to end stops there, and the rest of them
			// carry on. Anything else is a failure and belongs to the program, so
			// it is left to travel up.
			if (!go2jsIsGoexit(thrown)) {
				go2jsDraining = false;
				throw thrown;
			}
		}

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

// go2jsLetGoroutinesRun lets the goroutines that are waiting to start go. A Go
// program that has reached something which waits on another process has already
// had the goroutines written above it scheduled, so a listener opened in one of
// them is open by then; waiting for a process is one of the moments at which
// that has to be so, since nothing else here would ever hand the turn over.
function go2jsLetGoroutinesRun() {
	if (go2jsTasks.length === 0) {
		return 0;
	}

	return go2jsRunTasks();
}

// go2jsProgress lets the goroutines that are runnable run, and says whether any
// of them did. A goroutine waiting on a timer is runnable once its moment comes,
// so a select with nothing else to do waits for the earliest one rather than
// reporting that every goroutine is asleep.
function go2jsProgress() {
	if (go2jsTasks.length > 0 && go2jsRunTasks() > 0) {
		return true;
	}

	return go2jsWaitForEarliestTimer();
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
	return { entries: new go2jsNativeMap() };
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
	if (go2jsPointerAccessor(target, go2jsPointerGet) && go2jsPointerAccessor(target, go2jsPointerSet)) {
		return {
			get value() {
				return target[go2jsPointerGet]();
			},
			set value(next) {
				target[go2jsPointerSet](next);
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

// go2jsErrorMethodCall invokes an error chain method such as Is or Unwrap on a
// value, a pointer box, an interface holder, or an embedded promotion proxy.
function go2jsErrorMethodCall(err, name, ...args) {
	if (err === null || err === undefined) {
		return undefined;
	}

	let receiver = err;

	if (err.__go2js_pointer === true) {
		receiver = err[go2jsPointerGet]();

		if (receiver === null || receiver === undefined) {
			return undefined;
		}
	}

	if (err.__go2js_interface === true) {
		if (typeof err.type === "string") {
			const fn = go2jsMethodTable[err.type + "." + name];

			if (typeof fn === "function") {
				return fn.apply(null, [err.value, ...args]);
			}
		}

		receiver = err.value;
	}

	if (receiver === null || receiver === undefined) {
		return undefined;
	}

	if (typeof receiver === "object" || typeof receiver === "function") {
		const fn = receiver[name];

		if (typeof fn === "function") {
			return fn.apply(receiver, args);
		}
	}

	return undefined;
}

// go2jsErrorUnwrapAll returns the direct children of an error chain node, which
// covers cause fields, errors.Join results, and Unwrap() returning one or many
// errors.
function go2jsErrorUnwrapAll(err) {
	if (err === null || err === undefined) {
		return [];
	}

	const unwrapped = go2jsErrorMethodCall(err, "Unwrap");

	if (unwrapped !== undefined && unwrapped !== null) {
		return Array.isArray(unwrapped) ? unwrapped.filter(item => item !== null && item !== undefined) : [unwrapped];
	}

	if (err.cause !== undefined && err.cause !== null) {
		return [err.cause];
	}

	if (Array.isArray(err.joined)) {
		return err.joined.filter(item => item !== null && item !== undefined);
	}

	return [];
}

function go2jsErrorsIs(err, target) {
	if (err === target) {
		return true;
	}

	if (err === null || err === undefined || target === null || target === undefined) {
		return err === target;
	}

	if (go2jsErrorMethodCall(err, "Is", target) === true) {
		return true;
	}

	return go2jsErrorUnwrapAll(err).some(part => go2jsErrorsIs(part, target));
}

function go2jsWrapError(format, ...args) {
	const error = new Error(go2jsSprintfWrapping(format, args));
	// fmt.Errorf names its result after how many errors it wraps: one gives a
	// wrapError, more than one gives a wrapErrors that wraps them all.
	const wrapped = (String(format).match(/%[+#0 -.]*[0-9.]*w/g) || []).length;

	go2jsNameError(error, wrapped === 0 ? "*errors.errorString" : wrapped === 1 ? "*fmt.wrapError" : "*fmt.wrapErrors");

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

	const unwrapped = go2jsErrorUnwrapAll(err);

	// errors.Unwrap reports a single error, so a joined chain unwraps to nil.
	if (unwrapped.length !== 1) {
		return null;
	}

	return unwrapped[0];
}

function go2jsErrorsJoin(...errs) {
	const parts = errs.flat().filter(item => item !== null && item !== undefined);

	if (parts.length === 0) {
		return null;
	}

	if (parts.length === 1) {
		return parts[0];
	}

	const error = new Error(parts.map(part => go2jsErrorMessage(part)).join("\n"));
	error.joined = parts;

	return go2jsNameError(error, "*errors.joinError");
}

function go2jsErrorMessage(err) {
	if (err === null || err === undefined) {
		return "<nil>";
	}

	// An error is asked of itself what it says, which a value carrying an Error
	// method of its own is asked the same way, so that joining errors says what
	// each of them says rather than what it is built from.
	return go2jsErrorString(go2jsUnwrap(err));
}

function go2jsDuration(nanoseconds) {
	// a span of time that is already one keeps the shape it has, since giving
	// it a second would bury the first inside it
	if (go2jsIsDuration(nanoseconds)) {
		return nanoseconds;
	}

	return {
		nanoseconds: nanoseconds,
		// the name of the type is kept where a verb asking for it can find it,
		// since the object is its own kind rather than something wrapped
		__go2js_type_name: "time.Duration",
		valueOf() {
			return this.nanoseconds;
		},
		String() {
			return go2jsDurationString(this.nanoseconds);
		},
		Nanoseconds() {
			return this.nanoseconds;
		},
		Microseconds() {
			return Math.trunc(this.nanoseconds / 1000);
		},
		Milliseconds() {
			return Math.trunc(this.nanoseconds / 1000000);
		},
		Seconds() {
			return this.nanoseconds / 1000000000;
		},
		Minutes() {
			return this.nanoseconds / 60000000000;
		},
		Hours() {
			return this.nanoseconds / 3600000000000;
		},
		Abs() {
			return DurationAbs(this);
		},
		Truncate(m) {
			return DurationTruncate(this, m);
		},
		Round(m) {
			return DurationRound(this, m);
		}
	};
}

// go2jsDurationNanos accepts either a duration object or a raw nanosecond count,
// and gives the count back as it is held, since a count with more digits than a
// double keeps is held as digits and must not be brought down to a double here.
function go2jsDurationNanos(value) {
	if (go2jsIsDuration(value)) {
		return value.nanoseconds;
	}

	if (typeof value === "bigint") {
		return value;
	}

	return Number(value);
}

// go2jsDurationNanosAsNumber gives the count back as a double, which is what
// the helpers that divide the count and answer with a fraction need.
function go2jsDurationNanosAsNumber(value) {
	return Number(go2jsDurationNanos(value));
}

// go2jsDurationCmp answers which of two spans of time comes first, as JavaScript
// cannot be asked the question itself: the two counts may be held as a plain
// number or as a wide one, and a wide number is never equal to a plain one
// however alike the two look.
function go2jsDurationCmp(left, right) {
	let a = go2jsDurationNanos(left);
	let b = go2jsDurationNanos(right);

	if (typeof a === "bigint" || typeof b === "bigint") {
		a = BigInt(a);
		b = BigInt(b);
	} else {
		a = Number(a);
		b = Number(b);
	}

	if (a < b) {
		return -1;
	}

	return a > b ? 1 : 0;
}

// Methods on the named scalar type time.Duration are emitted as flat functions.
function DurationString(d) {
	return go2jsDurationString(go2jsDurationNanos(d));
}

function DurationNanoseconds(d) {
	return go2jsDurationNanos(d);
}

function DurationMicroseconds(d) {
	const span = go2jsDurationNanos(d);

	if (typeof span === "bigint") {
		return go2jsNarrow(span / 1000n);
	}

	return Math.trunc(span / 1000);
}

function DurationMilliseconds(d) {
	const span = go2jsDurationNanos(d);

	if (typeof span === "bigint") {
		return go2jsNarrow(span / 1000000n);
	}

	return Math.trunc(span / 1000000);
}

function DurationSeconds(d) {
	return go2jsDurationNanosAsNumber(d) / 1000000000;
}

function DurationMinutes(d) {
	return go2jsDurationNanosAsNumber(d) / 60000000000;
}

function DurationHours(d) {
	return go2jsDurationNanosAsNumber(d) / 3600000000000;
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

// A span of time too many digits to be a double is put in words the same way,
// working the digits themselves rather than rounding them away.
function go2jsDurationStringBig(nanoseconds) {
	if (nanoseconds === 0n) {
		return "0s";
	}

	const sign = nanoseconds < 0n ? "-" : "";
	let rest = nanoseconds < 0n ? -nanoseconds : nanoseconds;
	const billion = 1000000000n;
	const minute = 60000000000n;
	const hour = 3600000000000n;

	if (rest < billion) {
		let scale = 1n;
		let unit = "ns";

		if (rest >= 1000000n) {
			scale = 1000000n;
			unit = "ms";
		} else if (rest >= 1000n) {
			scale = 1000n;
			unit = "\u00b5s";
		}

		return sign + go2jsFractionDigits(rest, scale) + unit;
	}

	const hours = rest / hour;
	rest -= hours * hour;

	const minutes = rest / minute;
	rest -= minutes * minute;

	let text = "";

	if (hours > 0n) {
		text += hours + "h";
	}

	if (hours > 0n || minutes > 0n) {
		text += minutes + "m";
	}

	return sign + text + go2jsFractionDigits(rest, billion) + "s";
}

// go2jsFractionDigits writes a fraction of a whole part out the way Go writes
// one, in digits, with the zeros on the right left off and none at all when
// there is nothing to say. The unit is left to the caller, which is the one
// that knows what is being measured.
function go2jsFractionDigits(numerator, denominator) {
	const whole = numerator / denominator;
	const fraction = (numerator % denominator).toString().padStart(denominator.toString().length - 1, "0").replace(/0+$/, "");

	if (fraction === "") {
		return String(whole);
	}

	return String(whole) + "." + fraction;
}

function go2jsDurationString(nanoseconds) {
	if (typeof nanoseconds === "bigint") {
		return go2jsDurationStringBig(nanoseconds);
	}

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

function go2jsStdoutWrite(values, suffix, separator) {
	go2jsWriteOut(process.stdout, go2jsOutputText(values, suffix, separator));
}

function go2jsStderrWrite(values, suffix, separator) {
	go2jsWriteOut(process.stderr, go2jsOutputText(values, suffix, separator));
}

function go2jsOperandIsString(value) {
	return typeof value === "string";
}

function go2jsOutputText(value, suffix, separator) {
	if (value === undefined) {
		return "";
	}

	if (Array.isArray(value)) {
		if (separator === undefined || separator === null || separator === "") {
			let out = "";

			for (let index = 0; index < value.length; index++) {
				if (index > 0 && !go2jsOperandIsString(value[index - 1]) && !go2jsOperandIsString(value[index])) {
					out += " ";
				}

				out += go2jsErrorString(value[index]);
			}

			return out + suffix;
		}

		return value.map(item => go2jsErrorString(item)).join(separator) + suffix;
	}

	return go2jsErrorString(value) + suffix;
}

function go2jsExit(code) {
	process.exit(code === undefined ? 0 : Number(code));
}

function DurationAbs(d) {
	const span = go2jsDurationNanos(d);

	if (span >= 0) {
		return go2jsDuration(span);
	}

	// the smallest count there is has no positive counterpart inside the width
	// it is held in, so its own absolute value is the largest there is
	if (span === -go2jsSignBit64) {
		return go2jsDuration(go2jsSignBit64 - 1n);
	}

	return go2jsDuration(-span);
}

function DurationTruncate(d, multiple) {
	const step = go2jsDurationNanosAsNumber(multiple);

	if (step === 0) {
		return go2jsDuration(go2jsDurationNanos(d));
	}

	return go2jsDuration(Math.trunc(go2jsDurationNanosAsNumber(d) / step) * step);
}

function DurationRound(d, multiple) {
	const step = go2jsDurationNanosAsNumber(multiple);

	if (step === 0) {
		return go2jsDuration(go2jsDurationNanos(d));
	}

	const value = go2jsDurationNanosAsNumber(d);
	const half = step / 2;
	let offset = value % step;

	if (offset < 0) {
		offset += step;
	}

	const rounded = offset >= half ? value + (step - offset) : value - offset;

	return go2jsDuration(rounded);
}

function go2jsParseDuration(text) {
	const source = String(text);
	const scale = {
		ns: 1,
		us: 1000,
		"\u00b5s": 1000,
		"\u03bcs": 1000,
		ms: 1000000,
		s: 1000000000,
		m: 60000000000,
		h: 3600000000000
	};

	const pattern = /([+-]?(?:\d+(?:\.\d*)?|\.\d+))(ns|us|\u00b5s|\u03bcs|ms|s|m|h)/g;

	let total = 0;
	let matched = false;
	let position = 0;
	let match;

	while ((match = pattern.exec(source)) !== null) {
		if (match.index !== position) {
			return [go2jsDuration(0), go2jsSentinelError("time: invalid duration " + JSON.stringify(source))];
		}

		position = pattern.lastIndex;
		matched = true;

		let value = Number(match[1]);
		if (value < 0) {
			value = -value;
			total -= value * scale[match[2]];
		} else {
			total += value * scale[match[2]];
		}
	}

	if (!matched || position !== source.replace(/^\s+|\s+$/g, "").length) {
		return [go2jsDuration(0), go2jsSentinelError("time: invalid duration " + JSON.stringify(source))];
	}

	return [go2jsDuration(Math.round(total)), null];
}

`
}
