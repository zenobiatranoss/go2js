package javascript

// syncFuncs are the functions of the sync package that stand on their own rather
// than hang off one of its types.
var syncFuncs = map[string]string{
	"NewCond": "go2jsSyncNewCond",
}

// syncTypes are the types of the sync package the runtime answers with a class of
// its own, so a value of one is built by constructing that class rather than by
// calling a function that answers with an object.
var syncTypes = map[string]string{
	"sync.Pool": "go2jsSyncPool",

	// A condition is not made by constructing a value of its type but by asking
	// the package for one, and it is named here for the same reason a pool is: so
	// that a method hanging off it is written out by this runtime rather than
	// left as a property of an object that has no answer for it.
	"sync.Cond": "go2jsSyncCond",
}

// syncMethods are the methods of the types of the sync package, written out by
// the type they hang off, so that a value answers to them the way it does in Go
// and so that a method which has to wait is a delegation.
var syncMethods = map[string]map[string]string{
	"sync.Cond": {
		"Wait":      "go2jsCondWait",
		"Signal":    "go2jsCondSignal",
		"Broadcast": "go2jsCondBroadcast",
	},
	"sync.Pool": {
		"Get": "go2jsSyncPoolGet",
		"Put": "go2jsSyncPoolPut",
	},
}

func init() {
	// The sync package is one the runtime already answers part of, so what is
	// added here is folded into what it has rather than standing beside it: a
	// package is supported when everything in it is.
	supported := supportedStdlibPackages["sync"]

	if supported == nil {
		supported = map[string]bool{}
		supportedStdlibPackages["sync"] = supported
	}

	for name := range syncFuncs {
		supported[name] = true
	}

	for name, value := range syncTypes {
		packageTypes[name] = value
		supported[name] = true
	}

	for typeName, methods := range syncMethods {
		helpers := syncMethodHelpers[typeName]

		if helpers == nil {
			helpers = map[string]string{}
			syncMethodHelpers[typeName] = helpers
		}

		for method, helper := range methods {
			helpers[method] = helper
		}

		for method := range methods {
			supported[method] = true
		}
	}
}

func syncShimRuntimeSource() string {
	return `
// A pool of values is a pool of what was put into it and nothing else: a value
// that was never put into it is nothing, since that is what a pool holds before
// anything is put in. What a pool holds is not promised to still be there, since
// the host is free to empty it between one turn and the next, so the values are
// held in the pool's own store and anything else a pool could be holding is
// nothing.
function go2jsSyncPool() {
	return {items: [], New: null, victim: null};
}

// go2jsSyncPoolGet takes a value out of the pool. A pool with nothing of its own
// in it asks the program for one, and a program that named no way to make one
// gets nothing, which is the same as the value of a pool before anything is put
// into it.
function go2jsSyncPoolGet(pool) {
	for (let i = pool.items.length - 1; i >= 0; i--) {
		const item = pool.items[i];

		if (item !== null && item !== undefined) {
			pool.items.splice(i, 1);
			return item;
		}
	}

	// The way of making one is a Go function, which is written out as a generator
	// because a Go function may wait. Making a value for a pool does not wait,
	// so it is run to its end here rather than waited on.
	if (typeof pool.New === "function") {
		return go2jsCallNow(pool.New, null, []);
	}

	return null;
}

// go2jsSyncPoolPut leaves a value in the pool for another goroutine to ask for.
// A value that is nothing is not left in the pool at all, which is what a pool
// holding nothing is worth and what a program putting nothing into one means.
function go2jsSyncPoolPut(pool, value) {
	if (value === null || value === undefined) {
		return;
	}

	if (go2jsIsInterfaceValue(value) && go2jsInterfaceValue(value) === null) {
		return;
	}

	pool.items.push(value);
}

// A condition is a way of saying that something has become true to whoever is
// waiting to hear it, which is a wait the program can be woken from without
// anything being handed to it.
function go2jsSyncNewCond(locker) {
	return go2jsSyncCond(locker);
}

function go2jsSyncCond(locker) {
	return {locker: locker, waiting: [], parked: null, standIn: null};
}

// go2jsCondWait waits to be woken, and comes back holding the lock again, which
// is what the caller must already be holding when it waits: a wait on a condition
// releases the lock while it waits and takes it back when it is woken, since the
// thing it waits for is said by something else.
function* go2jsCondWait(cond) {
	const locker = go2jsCondLocker(cond);

	go2jsMutexUnlock(locker);

	const task = go2jsRunning;

	cond.waiting.push(task);

	// Waking is not handing over: the goroutine is runnable, and it takes the
	// lock back for itself when it is next run.
	yield go2jsCondPark(cond);

	yield* go2jsMutexLock(locker);
}

function go2jsCondPark(cond) {
	const task = go2jsRunning;

	if (task === null || task === undefined) {
		throw new Error("go2js: blocking operation outside a goroutine");
	}

	cond.parked = task;
	task.parked = "cond";

	return go2jsParked;
}

// go2jsCondSignal wakes one goroutine waiting on the condition, which is the one
// that has been waiting longest. Nobody in particular is woken when there is
// nobody waiting, since a condition said to nobody is said and forgotten.
function go2jsCondSignal(cond) {
	if (cond.waiting.length === 0) {
		return;
	}

	const task = cond.waiting.shift();

	go2jsRunnable(task);
}

// go2jsCondBroadcast wakes every goroutine waiting on the condition, which is
// what a change that several of them are waiting for is said with.
function go2jsCondBroadcast(cond) {
	const waiting = cond.waiting;

	cond.waiting = [];

	for (const task of waiting) {
		go2jsRunnable(task);
	}
}

// go2jsCondLocker is the lock a condition is waited on with. A condition is made
// over a lock behind an interface and behind a pointer, so the lock itself is
// found by looking through both: what a wait has to release is the lock the
// caller locked, not the box it was named through.
function go2jsCondLocker(cond) {
	const locker = go2jsCondUnderlying(cond.locker);

	if (locker !== null && locker !== undefined && typeof locker.locked === "boolean" && locker.blocked !== undefined) {
		return locker;
	}

	// A condition made over a lock of a kind the runtime has no lock for is waited
	// on by a lock of its own, which stands in for the lock of the program: what
	// a wait does is let others in and take the lock back, and that is what the
	// stand-in is used for.
	if (!cond.standIn) {
		cond.standIn = go2jsMutex();
	}

	return cond.standIn;
}

// go2jsCondUnderlying looks through an interface and a pointer to what they are
// holding, which is how a lock reaches a condition.
function go2jsCondUnderlying(value) {
	let current = value;

	for (let depth = 0; current !== null && current !== undefined && typeof current === "object" && depth < 16; depth++) {
		if (typeof current.locked === "boolean" && current.blocked !== undefined) {
			return current;
		}

		if (current.__go2js_interface === true && current.value !== undefined) {
			current = current.value;
			continue;
		}

		if (current.__go2js_pointer === true) {
			current = current[go2jsPointerGet]();
			continue;
		}

		break;
	}

	return current;
}

function go2jsIsInterfaceValue(value) {
	return value !== null && typeof value === "object" && value.__go2js_interface === true;
}
`
}
