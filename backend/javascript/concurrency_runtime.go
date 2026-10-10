package javascript

func concurrencyRuntimeSource() string {
	return `
function go2jsChannel(capacity) {
	return {
		__go2js_channel: true,
		buffer: [],
		capacity: typeof capacity === "number" && capacity > 0 ? Math.trunc(capacity) : 0,
		closed: false,
		blocked: [],
		handoff: []
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
	// the timer was asked for, and the moment it fires at is not a result. The
	// function runs as a goroutine of its own, which is what the timer is
	// written in Go to mean.
	if (typeof channel.timerCallback === "function") {
		const run = channel.timerCallback;

		delete channel.timerCallback;
		go2jsGo(run);

		return true;
	}

	channel.buffer.push(go2jsTimeValue(new Date(channel.timerDue)));
	go2jsWake(channel);

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

// go2jsPark suspends the goroutine that is running until one of the channels
// it names changes in a way that might let its operation through. The task is
// registered on every one of them, so a change to any of them makes it
// runnable again, and the operation is then tried once more from the top.
const go2jsParked = {};

function go2jsPark(channels, kind, value) {
	const task = go2jsRunning;

	if (task === null || task === undefined) {
		throw new Error("go2js: blocking operation outside a goroutine");
	}

	for (const channel of channels) {
		channel.blocked.push(task);
		task.waiting.push(channel);
	}

	task.parked = kind;
	task.parkedValue = value;

	// A goroutine that parks to send or to receive makes a case of a select
	// that is waiting on the same channel ready, so any select parked there is
	// woken to look again. Go hands a value straight from one goroutine to
	// another without either of them waiting again, and a select is a
	// goroutine that is waiting to be given one.
	if (kind === "receiver" || kind === "sender") {
		const selects = [];

		for (const channel of channels) {
			for (const entry of channel.blocked) {
				if (entry.parked === "select") {
					selects.push(entry);
				}
			}
		}

		for (const entry of selects) {
			go2jsRunnable(entry);
		}
	}

	return go2jsParked;
}

// go2jsNeverReady parks a goroutine on nothing at all, which is what an
// operation on a channel that was never made waits for: it never becomes
// ready, so the goroutine waits for ever.
function go2jsNeverReady() {
	return go2jsPark([], "never", undefined);
}

// go2jsWake makes every goroutine waiting on the channel runnable, since the
// channel it was waiting on has just changed.
function go2jsWake(channel) {
	if (channel.blocked.length === 0) {
		return;
	}

	const waiting = channel.blocked;

	channel.blocked = [];

	for (const task of waiting) {
		go2jsRunnable(task);
	}
}

function go2jsRunnable(task) {
	if (task.finished || go2jsReady.indexOf(task) >= 0) {
		return;
	}

	for (const channel of task.waiting) {
		const index = channel.blocked.indexOf(task);

		if (index >= 0) {
			channel.blocked.splice(index, 1);
		}
	}

	task.waiting = [];
	task.parked = null;
	go2jsReady.push(task);
}

// go2jsTakeBlocked takes a goroutine that is waiting on the channel for the
// kind of operation named, so that a send can be handed to a receiver or a
// receive answered by a sender without either of them waiting again.
function go2jsTakeBlocked(channel, kind) {
	for (let index = 0; index < channel.blocked.length; index++) {
		const task = channel.blocked[index];

		if (task.parked === kind) {
			channel.blocked.splice(index, 1);
			task.waiting = task.waiting.filter((entry) => entry !== channel);
			return task;
		}
	}

	return null;
}

function go2jsHasBlocked(channel, kind) {
	for (const task of channel.blocked) {
		if (task.parked === kind) {
			return true;
		}
	}

	return false;
}

// A channel of no room keeps a value that has been handed to a goroutine which
// has not taken it yet, since a goroutine that was parked waiting for a value
// does not hold it until it runs again. Whoever runs next and receives from the
// channel takes the value that is waiting there.
function go2jsChanHandoff(channel, value) {
	if (channel.handoff === undefined) {
		channel.handoff = [];
	}

	channel.handoff.push(value);
}

function go2jsChanHandoffWaiting(channel) {
	return channel.handoff !== undefined && channel.handoff.length > 0;
}

// A Go function is written as a generator, so calling one is stepping it: the
// call is delegated with yield*, which runs the callee as part of the caller's
// own goroutine and lets either of them wait. A value the compiler could not
// tell the shape of is called through go2jsCall instead, which runs a generator
// and passes through anything else unchanged.

// go2jsGenerator says whether a value handed back from a call is a goroutine
// that has been started but not run.
function go2jsGenerator(value) {
	return value !== null && typeof value === "object" && typeof value.next === "function";
}

function* go2jsCall(fn, self, args) {
	const result = fn.apply(self, args);

	if (go2jsGenerator(result)) {
		return yield* result;
	}

	return result;
}

// go2jsCallNow calls a value the runtime was handed as a callback. The
// callback runs to its end here, and if it waits on something then the
// goroutines that can be run are run until it is through, which is what a
// callback such as a comparison or a visitor has to be given: nothing here
// waits with it, so the wait it makes is one the scheduler answers.
function go2jsCallNow(fn, self, args) {
	const result = fn.apply(self, args);

	if (go2jsGenerator(result)) {
		return go2jsRunNow(result);
	}

	return result;
}

// go2jsCallFromJavaScript calls a function of the program from outside it, which
// is what a program that hands a function over to a module is asked for. It runs
// as a goroutine of the program would, since such a function may wait on another
// goroutine of the program to answer it, and the scheduler is what answers it.
// What the call comes to is what the caller is handed, and a fault in it is
// thrown at the caller the way a fault in the program is thrown.
function go2jsCallFromJavaScript(fn, self, args) {
	let result;

	const task = go2jsStart((function* () {
		result = yield* go2jsCall(fn, self, args);
	})());

	while (go2jsLiveTasks > 0) {
		if (go2jsWaitForWork()) {
			continue;
		}

		// Nothing is runnable and nothing is coming due. A call that has already
		// come back has ended, as a program that has returned has, and the
		// goroutines still waiting are left where they are; a call still going is a
		// program with every goroutine asleep, which Go calls a deadlock.
		if (task.finished) {
			go2jsReady.length = 0;
			break;
		}

		throw go2jsFatalError("all goroutines are asleep - deadlock!");
	}

	return result;
}

function go2jsRunNow(generator) {
	// A thunk the compiler could not see the shape of may well turn out not to
	// be a generator, and then there is nothing to run: the value it produced is
	// the answer.
	if (!go2jsGenerator(generator)) {
		return generator;
	}

	for (;;) {
		const step = generator.next();

		if (step.done) {
			return step.value;
		}

		if (!go2jsWaitForWork()) {
			throw go2jsFatalError("all goroutines are asleep - deadlock!");
		}
	}
}

function* go2jsChanRecvPair(channel, zero) {
	// A receive from a channel that was never made waits for ever, the same
	// way it does in Go, and no send can ever answer it.
	if (go2jsChannelIsNil(channel)) {
		yield go2jsNeverReady();

		return [zero, false];
	}

	while (true) {
		if (channel.buffer.length > 0) {
			const value = channel.buffer.shift();
			go2jsWake(channel);

			return [value, true];
		}

		if (go2jsTimerDue(channel)) {
			const value = channel.buffer.shift();

			return [value, true];
		}

		// A value handed over to this goroutine while it was parked is taken
		// here, before the close of the channel is looked at, since a value
		// that was sent is still there to be read after the close.
		if (go2jsChanHandoffWaiting(channel)) {
			const value = channel.handoff.shift();

			return [value, true];
		}

		if (channel.closed) {
			return [zero, false];
		}

		// A goroutine already waiting to send on this channel has the value
		// in hand, so the receive takes it from there and lets the send finish
		// rather than storing it for a moment and taking it back out again.
		const sender = go2jsTakeBlocked(channel, "sender");

		if (sender !== null) {
			const value = sender.parkedValue;

			sender.parked = null;
			sender.parkedValue = undefined;
			sender.answered = true;
			go2jsRunnable(sender);

			return [value, true];
		}

		yield go2jsPark([channel], "receiver", undefined);
	}
}

function* go2jsChanRecv(channel, zero) {
	const pair = yield* go2jsChanRecvPair(channel, zero);

	return pair[0];
}

function go2jsChanTryRecv(channel, zero) {
	if (go2jsChannelIsNil(channel)) {
		return null;
	}

	if (channel.buffer.length > 0) {
		const value = channel.buffer.shift();
		go2jsWake(channel);

		return [value, true];
	}

	if (channel.closed) {
		return [zero, false];
	}

	return null;
}

function go2jsChanRecvReady(channel) {
	if (go2jsChannelIsNil(channel)) {
		return false;
	}

	if (channel.buffer.length > 0 || channel.closed || go2jsTimerDue(channel)) {
		return true;
	}

	if (go2jsChanHandoffWaiting(channel)) {
		return true;
	}

	// A receive from a channel of no room is answered by whoever is holding a
	// value for it, so a send already waiting on this channel is what makes it
	// ready.
	return go2jsHasBlocked(channel, "sender");
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
		return go2jsHasBlocked(channel, "receiver");
	}

	return channel.buffer.length < channel.capacity;
}

function* go2jsChanSend(channel, value) {
	if (go2jsChannelIsNil(channel)) {
		yield go2jsNeverReady();

		return;
	}

	while (true) {
		if (channel.closed) {
			go2jsChannelClosedPanic("send");
		}

		// A channel of no room is a handover: the value is not sent until the
		// goroutine that was waiting for it has taken it, so the value goes
		// straight into the hands of a receiver that is already waiting.
		if (channel.capacity === 0) {
			// A goroutine that was parked in a select is waiting for a value as
			// surely as one parked in a receive, so it is handed the value the
			// same way and looks at the channel again when it runs.
			const receiver = go2jsTakeBlocked(channel, "receiver") ||
				go2jsTakeBlocked(channel, "select");

			if (receiver !== null) {
				go2jsChanHandoff(channel, value);
				go2jsRunnable(receiver);

				return;
			}
		} else if (channel.buffer.length < channel.capacity) {
			channel.buffer.push(value);
			go2jsWake(channel);

			return;
		}

		yield go2jsPark([channel], "sender", value);

		// The receive that took the value off this goroutine has already made
		// the send complete, so there is nothing left to hand over.
		if (go2jsRunning !== null && go2jsRunning.answered) {
			go2jsRunning.answered = false;

			return;
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
	go2jsWake(channel);

	return true;
}

function go2jsChannelClose(channel) {
	if (channel === null || channel === undefined) {
		throw new Error("close of nil channel");
	}

	if (channel.closed) {
		throw new Error("close of closed channel");
	}

	channel.closed = true;

	// Closing a channel is what lets everything waiting on it go again: a
	// receive reads the zero value, a send finds the channel closed under it,
	// and a select chooses whichever of its cases the close made ready.
	go2jsWake(channel);
}

// go2jsSelectWait waits for one of the channels a select was written over to
// become ready. It is what a select that had nothing to choose from waits on,
// and it parks on every channel the select mentioned, so whichever one changes
// first brings the select back to look at all of them again.
function* go2jsSelectWait(channels) {
	yield go2jsPark(channels.filter((channel) => !go2jsChannelIsNil(channel)), "select", undefined);

	if (channels.length === 0) {
		yield go2jsNeverReady();
	}
}

function go2jsWaitGroup() {
	return {
		__go2js_channel: true,
		blocked: [],
		capacity: 0,
		buffer: [],
		closed: false,
		count: 0,
		handoff: []
	};
}

function go2jsWaitGroupAdd(group, delta) {
	group.count += delta;

	if (group.count < 0) {
		throw new Error("sync: negative WaitGroup counter");
	}

	// A wait group is reached at zero, which is the moment every goroutine
	// waiting on it is free to carry on.
	if (group.count === 0) {
		go2jsWake(group);
	}
}

function go2jsWaitGroupDone(group) {
	go2jsWaitGroupAdd(group, -1);
}

// A task is one goroutine. The body of every Go function is written as a
// generator, so a task is that generator together with the channels its
// current operation is waiting on. Running the task means giving the generator
// its next step; where it stops is where it waits until a channel changes.
function go2jsTask(generator) {
	return {
		generator,
		waiting: [],
		parked: null,
		parkedValue: undefined,
		answered: false,
		finished: false
	};
}

// go2jsReady holds the goroutines that can be stepped right now, and
// go2jsRunning holds the one being stepped, so that a goroutine that is waiting
// knows what it is waiting as.
const go2jsReady = [];
let go2jsRunning = null;
let go2jsLiveTasks = 0;

function go2jsStep() {
	let steps = 0;

	while (go2jsReady.length > 0) {
		if (steps >= 1000000) {
			go2jsReady.length = 0;
			throw new Error("go2js: goroutine task limit exceeded");
		}

		const task = go2jsReady.shift();

		go2jsRunning = task;

		let step;

		try {
			step = task.generator.next();
		} catch (thrown) {
			go2jsRunning = null;
			task.finished = true;
			go2jsLiveTasks--;

			// A goroutine that was told to end stops there, and the rest of them
			// carry on. Anything else is a failure and belongs to the program, so
			// it is left to travel up.
			if (go2jsIsGoexit(thrown)) {
				steps++;
				continue;
			}

			go2jsReady.length = 0;
			throw thrown;
		}

		go2jsRunning = null;
		steps++;

		if (step.done) {
			task.finished = true;
			go2jsLiveTasks--;
			continue;
		}

		// A step that stopped without parking its goroutine asked for the turn
		// to be handed to another one, so it goes back into the queue.
		if (step.value !== go2jsParked) {
			go2jsReady.push(task);
		}
	}

	return steps;
}

// go2jsWakeDueTimers wakes what is waiting on a timer that has come due. A
// timer is noticed by the goroutine that is waiting on it when that goroutine
// runs, so reaching the moment a timer names is what makes it runnable again.
function go2jsWakeDueTimers() {
	const now = Date.now();

	for (const channel of go2jsTimers.keys()) {
		if (channel.timerDeadline <= now) {
			// A timer that runs a function has nothing waiting on the channel it
			// stands for, so it is fired from here rather than left for a
			// goroutine that may never come to look at it.
			if (typeof channel.timerCallback === "function") {
				go2jsTimerDue(channel);
			} else {
				go2jsWake(channel);
			}
		}
	}
}

// go2jsWaitForWork runs the goroutines that can be run, and if none of them can
// then waits for the moment the earliest timer comes due. It says whether
// there was anything at all to wait for.
function go2jsWaitForWork() {
	if (go2jsStep() > 0) {
		return true;
	}

	if (!go2jsWaitForEarliestTimer()) {
		return false;
	}

	go2jsWakeDueTimers();

	return true;
}

function go2jsStart(generator) {
	const task = go2jsTask(generator);

	go2jsLiveTasks++;
	go2jsReady.push(task);

	return task;
}

// go2jsGo writes "go f(x)": f runs as a goroutine of its own, starting from the
// moment it was written rather than straight away. What it is handed is the
// generator function holding the body, which is started and then left to be
// stepped by the scheduler like any other goroutine.
function go2jsGo(body) {
	go2jsStart(body());
}

// go2jsRunMain runs main as a goroutine too, so that everything else in the
// program is written the same way and the scheduler has one place it all goes.
function go2jsRunMain(generator) {
	const main = go2jsStart(generator);

	while (go2jsLiveTasks > 0) {
		if (go2jsWaitForWork()) {
			continue;
		}

		// Nothing is runnable and nothing is coming due. A program whose main
		// has already returned has ended, as a Go program does, and the
		// goroutines still waiting are left where they are; a program still
		// inside main has every goroutine asleep, which Go calls a deadlock and
		// says so.
		if (main.finished) {
			go2jsReady.length = 0;
			break;
		}

		throw go2jsFatalError("all goroutines are asleep - deadlock!");
	}
}

// go2jsLetGoroutinesRun hands the turn to the goroutines that can be run. A Go
// program that has reached something which waits on another process has already
// had the goroutines written above it scheduled, so a listener opened in one of
// them is open by then; waiting for a process is one of the moments at which
// that has to be so, since nothing else here would ever hand the turn over.
function* go2jsLetGoroutinesRun() {
	yield;
}

function* go2jsWaitGroupWait(group) {
	// Waiting on a wait group is parking on it, which the group wakes when the
	// count it holds comes back to zero.
	while (group.count > 0) {
		yield go2jsPark([group], "waitgroup", undefined);
	}
}

function go2jsMutex() {
	return { locked: false, blocked: [] };
}

// go2jsMutexLock takes a mutex, waiting for it when another goroutine is
// holding it. The wait is a park on the mutex itself, which the unlock wakes,
// so a goroutine that has to wait for a lock goes to the back of the queue with
// the goroutines that were waiting before it rather than finding the lock held
// under it.
function* go2jsMutexLock(mutex) {
	while (mutex.locked) {
		yield go2jsPark([mutex], "mutex", undefined);

		// The unlock that woke this goroutine gave the lock to it as it went, so
		// it holds the lock already and has nothing left to wait for.
		if (go2jsRunning.answered) {
			go2jsRunning.answered = false;

			return;
		}
	}

	mutex.locked = true;
}

function go2jsMutexUnlock(mutex) {
	if (!mutex.locked) {
		throw new Error("sync: unlock of unlocked mutex");
	}

	mutex.locked = false;

	// The goroutine that was waiting first for the lock is given it as the lock
	// is let go, so it goes straight from one holder to the next without a
	// moment of it free for anyone else to take.
	const next = go2jsTakeBlocked(mutex, "mutex");

	if (next !== null) {
		mutex.locked = true;
		go2jsAnswered(next);
	}
}

function go2jsRWMutex() {
	return { readers: 0, writer: false, blocked: [] };
}

// go2jsRWMutexRLock takes a lock to read, which waits only for the writers. A
// writer waiting for the lock asks for it to be given to no reader until it is
// through, which is what Go does to keep a reader from starving it.
function* go2jsRWMutexRLock(mutex) {
	while (mutex.writer || go2jsHasBlocked(mutex, "mutexwriter")) {
		yield go2jsPark([mutex], "mutexreader", undefined);

		if (go2jsRunning.answered) {
			go2jsRunning.answered = false;

			return;
		}
	}

	mutex.readers++;
}

function go2jsRWMutexRUnlock(mutex) {
	if (mutex.readers === 0) {
		throw new Error("sync: RUnlock of unlocked RWMutex");
	}

	mutex.readers--;
	go2jsRWMutexPromote(mutex);
}

function* go2jsRWMutexLock(mutex) {
	while (mutex.writer || mutex.readers > 0) {
		yield go2jsPark([mutex], "mutexwriter", undefined);

		if (go2jsRunning.answered) {
			go2jsRunning.answered = false;

			return;
		}
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

// go2jsRWMutexPromote gives the lock to a writer waiting for it, or else to
// every reader waiting for it once no writer holds it.
function go2jsRWMutexPromote(mutex) {
	if (mutex.writer || mutex.readers > 0) {
		return;
	}

	const writer = go2jsTakeBlocked(mutex, "mutexwriter");

	if (writer !== null) {
		mutex.writer = true;
		go2jsAnswered(writer);

		return;
	}

	while (go2jsHasBlocked(mutex, "mutexreader")) {
		mutex.readers++;
		go2jsAnswered(go2jsTakeBlocked(mutex, "mutexreader"));
	}
}

// go2jsAnswered lets a goroutine that was waiting be run again with what it was
// waiting for already granted, so it carries on from where it stopped rather
// than looking for it a second time.
function go2jsAnswered(task) {
	task.parked = null;
	task.parkedValue = undefined;
	task.answered = true;
	go2jsRunnable(task);
}

function go2jsOnce() {
	return { done: false };
}

function go2jsOnceDo(once, fn) {
	if (once.done) {
		return;
	}

	once.done = true;
	go2jsCallNow(fn, null, []);
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
		go2jsCallNow(fn, null, [key, value]);
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
				return go2jsCallNow(fn, null, [err.value, ...args]);
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
			return go2jsCallNow(fn, receiver, args);
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
		return go2jsCallNow(value.Error, value, []);
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
