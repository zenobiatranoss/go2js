package javascript

// ebitenRuntimeSource is what runs behind the ebiten packages of a program: a
// canvas in the page stands for the window, the keyboard and the mouse of the
// page for its input, and the game is stepped sixty times a second the way the
// library steps it. Under Node there is no page, so a program is stepped for as
// many frames as GO2JS_EBITEN_FRAMES says and draws to nothing, which is what
// lets the logic of a game be run and tested without a screen.
func ebitenRuntimeSource() string {
	return `
function go2jsColorRGBA() {
	return new go2jsColorRGBAValue();
}

function go2jsColorNRGBA() {
	return new go2jsColorNRGBAValue();
}

function go2jsColorGray() {
	return new go2jsColorGrayValue();
}

function go2jsEbitenGeoM() {
	return new go2jsEbitenGeoMValue();
}

function go2jsEbitenDrawImageOptions() {
	return new go2jsEbitenDrawImageOptionsValue();
}

class go2jsColorRGBAValue {
	constructor() {
		this.R = 0;
		this.G = 0;
		this.B = 0;
		this.A = 0;
	}

	RGBA() {
		return [this.R * 257, this.G * 257, this.B * 257, this.A * 257];
	}
}

class go2jsColorNRGBAValue {
	constructor() {
		this.R = 0;
		this.G = 0;
		this.B = 0;
		this.A = 0;
	}

	RGBA() {
		const a = this.A * 257;

		return [Math.floor(this.R * 257 * a / 65535), Math.floor(this.G * 257 * a / 65535), Math.floor(this.B * 257 * a / 65535), a];
	}
}

class go2jsColorGrayValue {
	constructor() {
		this.Y = 0;
	}

	RGBA() {
		return [this.Y * 257, this.Y * 257, this.Y * 257, 65535];
	}
}

// go2jsEbitenCSS turns a color the way a program holds it, an RGBA that already
// carries its alpha in its other channels or an NRGBA that does not, into the
// text a canvas takes.
function go2jsEbitenCSS(color) {
	let c = color;

	if (c !== null && c !== undefined && c.__go2js_interface === true) {
		c = c.value;
	}

	if (c === null || c === undefined) {
		return "rgba(0,0,0,0)";
	}

	if (c instanceof go2jsColorNRGBAValue) {
		return "rgba(" + c.R + "," + c.G + "," + c.B + "," + (c.A / 255) + ")";
	}

	if (c instanceof go2jsColorGrayValue) {
		return "rgb(" + c.Y + "," + c.Y + "," + c.Y + ")";
	}

	if (c instanceof go2jsColorRGBAValue) {
		if (c.A === 0) {
			return "rgba(0,0,0,0)";
		}

		const r = Math.min(255, Math.round(c.R * 255 / c.A));
		const g = Math.min(255, Math.round(c.G * 255 / c.A));
		const b = Math.min(255, Math.round(c.B * 255 / c.A));

		return "rgba(" + r + "," + g + "," + b + "," + (c.A / 255) + ")";
	}

	if (typeof c.RGBA === "function") {
		const parts = go2jsInvokeMethod(c, "RGBA", []);
		const a = parts[3] / 65535;

		if (a === 0) {
			return "rgba(0,0,0,0)";
		}

		return "rgba(" + Math.round(parts[0] / parts[3] * 255) + "," + Math.round(parts[1] / parts[3] * 255) + "," + Math.round(parts[2] / parts[3] * 255) + "," + a + ")";
	}

	return "rgba(0,0,0,1)";
}

class go2jsEbitenGeoMValue {
	constructor() {
		this.m = [1, 0, 0, 1, 0, 0];
	}

	// each step is applied after the ones before it, the way a matrix chain is
	apply(a, b, c, d, e, f) {
		const m = this.m;

		this.m = [
			a * m[0] + c * m[1], b * m[0] + d * m[1],
			a * m[2] + c * m[3], b * m[2] + d * m[3],
			a * m[4] + c * m[5] + e, b * m[4] + d * m[5] + f,
		];
	}

	Translate(tx, ty) {
		this.apply(1, 0, 0, 1, tx, ty);
	}

	Scale(x, y) {
		this.apply(x, 0, 0, y, 0, 0);
	}

	Rotate(theta) {
		this.apply(Math.cos(theta), Math.sin(theta), -Math.sin(theta), Math.cos(theta), 0, 0);
	}
}

class go2jsEbitenDrawImageOptionsValue {
	constructor() {
		this.GeoM = new go2jsEbitenGeoMValue();
	}
}

class go2jsEbitenImage {
	constructor(width, height, canvas) {
		this.width = width;
		this.height = height;
		this.canvas = canvas || null;
		this.ctx = this.canvas && this.canvas.getContext ? this.canvas.getContext("2d") : null;
	}

	Fill(color) {
		if (this.ctx === null) {
			return;
		}

		this.ctx.save();
		this.ctx.setTransform(1, 0, 0, 1, 0, 0);
		this.ctx.globalCompositeOperation = "copy";
		this.ctx.fillStyle = go2jsEbitenCSS(color);
		this.ctx.fillRect(0, 0, this.width, this.height);
		this.ctx.restore();
	}

	Clear() {
		if (this.ctx !== null) {
			this.ctx.clearRect(0, 0, this.width, this.height);
		}
	}

	Size() {
		return [this.width, this.height];
	}

	Set(x, y, color) {
		if (this.ctx !== null) {
			this.ctx.fillStyle = go2jsEbitenCSS(color);
			this.ctx.fillRect(x, y, 1, 1);
		}
	}

	DrawImage(image, options) {
		if (this.ctx === null || image === null || image.canvas === null) {
			return;
		}

		const m = options && options.GeoM ? options.GeoM.m : [1, 0, 0, 1, 0, 0];

		this.ctx.save();
		this.ctx.setTransform(m[0], m[1], m[2], m[3], m[4], m[5]);
		this.ctx.drawImage(image.canvas, 0, 0);
		this.ctx.restore();
	}
}

const go2jsEbitenKeyNames = [
	"KeyA", "KeyB", "KeyC", "KeyD", "KeyE", "KeyF", "KeyG", "KeyH", "KeyI", "KeyJ", "KeyK", "KeyL", "KeyM",
	"KeyN", "KeyO", "KeyP", "KeyQ", "KeyR", "KeyS", "KeyT", "KeyU", "KeyV", "KeyW", "KeyX", "KeyY", "KeyZ",
	"KeyDigit0", "KeyDigit1", "KeyDigit2", "KeyDigit3", "KeyDigit4", "KeyDigit5", "KeyDigit6", "KeyDigit7", "KeyDigit8", "KeyDigit9",
	"KeyArrowDown", "KeyArrowLeft", "KeyArrowRight", "KeyArrowUp", "KeyEnter", "KeyEscape", "KeySpace", "KeyTab", "KeyBackspace",
	"KeyShiftLeft", "KeyShiftRight", "KeyControlLeft", "KeyControlRight", "KeyAltLeft", "KeyAltRight",
	"KeyF1", "KeyF2", "KeyF3", "KeyF4", "KeyF5", "KeyF6", "KeyF7", "KeyF8", "KeyF9", "KeyF10", "KeyF11", "KeyF12",
];

// go2jsEbitenCodeIndex is the number of a key that the page names by its code.
const go2jsEbitenCodeIndex = go2jsEbitenKeyNames.reduce((index, name, i) => {
	index[name.slice(3)] = i;

	return index;
}, {});

const go2jsEbitenState = {
	windowWidth: 640,
	windowHeight: 480,
	title: "",
	tps: 60,
	actualTPS: 0,
	actualFPS: 0,
	down: new Set(),
	tapped: new Set(),
	held: new Map(),
	released: new Set(),
	mouseDown: new Set(),
	mouseTapped: new Set(),
	mouseHeld: new Map(),
	mouseJustReleased: new Set(),
	cursor: [0, 0],
	wheel: [0, 0],
	canvas: null,
	screen: null,
	running: false,
};

function go2jsEbitenKeyEvent(event, pressed) {
	let name = event.code || "";

	if (name.startsWith("Digit")) {
		name = name;
	} else if (name.startsWith("Key")) {
		name = name.slice(3);
	}

	const index = go2jsEbitenCodeIndex[name];

	if (index === undefined) {
		return false;
	}

	if (pressed) {
		go2jsEbitenState.down.add(index);
		go2jsEbitenState.tapped.add(index);
	} else {
		go2jsEbitenState.down.delete(index);
	}

	return true;
}

// go2jsEbitenPoll is what the library does before each step of the game: the
// keys and buttons that were down since the last step are counted, so that a
// key pressed and let go between two steps is still seen by one of them.
function go2jsEbitenPoll() {
	const state = go2jsEbitenState;

	for (const [store, now, tapped, released] of [
		[state.held, state.down, state.tapped, state.released],
		[state.mouseHeld, state.mouseDown, state.mouseTapped, state.mouseJustReleased],
	]) {
		released.clear();

		for (const key of new Set([...store.keys(), ...now, ...tapped])) {
			const before = store.get(key) || 0;
			const isDown = now.has(key) || tapped.has(key);

			if (isDown) {
				store.set(key, before + 1);
			} else {
				store.delete(key);

				if (before > 0) {
					released.add(key);
				}
			}
		}

		tapped.clear();
	}
}

function go2jsEbitenRunHeadless(game, frames) {
	const state = go2jsEbitenState;
	const size = go2jsInvokeMethod(game, "Layout", [state.windowWidth, state.windowHeight]);
	const screen = new go2jsEbitenImage(size[0], size[1], null);

	for (let i = 0; i < frames; i++) {
		go2jsEbitenPoll();

		const failure = go2jsInvokeMethod(game, "Update", []);

		if (failure !== null && failure !== undefined) {
			return failure === ebiten.Termination ? null : failure;
		}

		go2jsInvokeMethod(game, "Draw", [screen]);
	}

	return null;
}

function go2jsEbitenRunPage(game) {
	const state = go2jsEbitenState;
	const size = go2jsInvokeMethod(game, "Layout", [state.windowWidth, state.windowHeight]);
	const canvas = document.createElement("canvas");

	canvas.width = size[0];
	canvas.height = size[1];
	canvas.tabIndex = 0;
	canvas.style.cssText = "display:block;margin:auto;background:#000;image-rendering:pixelated;outline:none;max-width:100vw;max-height:100vh";
	canvas.style.width = state.windowWidth + "px";
	canvas.style.height = state.windowHeight + "px";

	document.body.style.cssText = "margin:0;background:#111;display:flex;min-height:100vh;align-items:center;justify-content:center";

	document.body.appendChild(canvas);

	if (state.title !== "") {
		document.title = state.title;
	}

	state.canvas = canvas;

	const screen = new go2jsEbitenImage(size[0], size[1], canvas);

	state.screen = screen;

	window.addEventListener("keydown", (event) => {
		if (go2jsEbitenKeyEvent(event, true) && event.code.startsWith("Arrow") || event.code === "Space") {
			event.preventDefault();
		}
	});
	window.addEventListener("keyup", (event) => {
		go2jsEbitenKeyEvent(event, false);
	});
	window.addEventListener("blur", () => {
		state.down.clear();
	});

	const toCanvas = (event) => {
		const rect = canvas.getBoundingClientRect();

		return [Math.floor((event.clientX - rect.left) * canvas.width / rect.width), Math.floor((event.clientY - rect.top) * canvas.height / rect.height)];
	};

	canvas.addEventListener("mousemove", (event) => {
		state.cursor = toCanvas(event);
	});
	canvas.addEventListener("mousedown", (event) => {
		state.cursor = toCanvas(event);
		state.mouseDown.add(event.button);
		state.mouseTapped.add(event.button);
		canvas.focus();
	});
	window.addEventListener("mouseup", (event) => {
		state.mouseDown.delete(event.button);
	});
	canvas.addEventListener("wheel", (event) => {
		state.wheel = [event.deltaX, event.deltaY];
		event.preventDefault();
	}, {passive: false});
	canvas.addEventListener("contextmenu", (event) => {
		event.preventDefault();
	});

	canvas.focus();
	state.running = true;

	let last = performance.now();
	let owed = 0;
	let frames = 0;
	let steps = 0;
	let counted = last;

	const frame = (now) => {
		if (!state.running) {
			return;
		}

		owed += Math.min(now - last, 250);
		last = now;

		const step = 1000 / state.tps;

		while (owed >= step) {
			owed -= step;
			go2jsEbitenPoll();

			let failure;

			try {
				failure = go2jsInvokeMethod(game, "Update", []);
			} catch (thrown) {
				state.running = false;
				throw thrown;
			}

			steps++;
			state.wheel = [0, 0];

			if (failure !== null && failure !== undefined) {
				state.running = false;

				if (failure !== ebiten.Termination) {
					console.error(go2jsStringify(failure));
				}

				return;
			}
		}

		go2jsInvokeMethod(game, "Draw", [screen]);
		frames++;

		if (now - counted >= 1000) {
			state.actualFPS = frames * 1000 / (now - counted);
			state.actualTPS = steps * 1000 / (now - counted);
			frames = 0;
			steps = 0;
			counted = now;
		}

		requestAnimationFrame(frame);
	};

	requestAnimationFrame(frame);

	return null;
}

// go2jsEbitenBuild adds the keys, which are numbered in the order the library
// names them, and the older names that stand for some of them.
function go2jsEbitenBuild(library) {
	for (let i = 0; i < go2jsEbitenKeyNames.length; i++) {
		library[go2jsEbitenKeyNames[i]] = i;
	}

	return Object.assign(library, {
		Key0: library.KeyDigit0, Key1: library.KeyDigit1, Key2: library.KeyDigit2, Key3: library.KeyDigit3, Key4: library.KeyDigit4,
		Key5: library.KeyDigit5, Key6: library.KeyDigit6, Key7: library.KeyDigit7, Key8: library.KeyDigit8, Key9: library.KeyDigit9,
		KeyDown: library.KeyArrowDown, KeyLeft: library.KeyArrowLeft, KeyRight: library.KeyArrowRight, KeyUp: library.KeyArrowUp,
	});
}

const ebiten = go2jsEbitenBuild({
	DefaultTPS: 60,
	Termination: go2jsErrorsNew("termination"),
	WindowResizingModeDisabled: 0,
	WindowResizingModeOnlyFullscreenEnabled: 1,
	WindowResizingModeEnabled: 2,
	CursorModeVisible: 0,
	CursorModeHidden: 1,
	CursorModeCaptured: 2,
	MouseButtonLeft: 0,
	MouseButtonRight: 2,
	MouseButtonMiddle: 1,
	NewImage(width, height) {
		if (typeof document === "undefined") {
			return new go2jsEbitenImage(width, height, null);
		}

		const canvas = document.createElement("canvas");

		canvas.width = width;
		canvas.height = height;

		return new go2jsEbitenImage(width, height, canvas);
	},
	RunGame(game) {
		if (typeof document === "undefined") {
			const frames = typeof process !== "undefined" && process.env ? Number(process.env.GO2JS_EBITEN_FRAMES || 0) : 0;

			if (frames > 0) {
				return go2jsEbitenRunHeadless(game, frames);
			}

			return go2jsErrorsNew("ebiten: no display; run the program in a page, or set GO2JS_EBITEN_FRAMES to step it without one");
		}

		return go2jsEbitenRunPage(game);
	},
	SetWindowSize(width, height) {
		go2jsEbitenState.windowWidth = width;
		go2jsEbitenState.windowHeight = height;

		if (go2jsEbitenState.canvas !== null) {
			go2jsEbitenState.canvas.style.width = width + "px";
			go2jsEbitenState.canvas.style.height = height + "px";
		}
	},
	WindowSize() {
		return [go2jsEbitenState.windowWidth, go2jsEbitenState.windowHeight];
	},
	SetWindowTitle(title) {
		go2jsEbitenState.title = title;

		if (typeof document !== "undefined") {
			document.title = title;
		}
	},
	SetWindowResizingMode() {},
	SetFullscreen() {},
	SetVsyncEnabled() {},
	SetCursorMode() {},
	SetTPS(tps) {
		go2jsEbitenState.tps = tps > 0 ? tps : 60;
	},
	SetMaxTPS(tps) {
		go2jsEbitenState.tps = tps > 0 ? tps : 60;
	},
	MaxTPS() {
		return go2jsEbitenState.tps;
	},
	ActualTPS() {
		return go2jsEbitenState.actualTPS;
	},
	ActualFPS() {
		return go2jsEbitenState.actualFPS;
	},
	IsKeyPressed(key) {
		return go2jsEbitenState.down.has(key);
	},
	IsMouseButtonPressed(button) {
		return go2jsEbitenState.mouseDown.has(button);
	},
	CursorPosition() {
		return go2jsEbitenState.cursor.slice();
	},
	Wheel() {
		return go2jsEbitenState.wheel.slice();
	},
});


const inpututil = {
	IsKeyJustPressed(key) {
		return go2jsEbitenState.held.get(key) === 1;
	},
	IsKeyJustReleased(key) {
		return go2jsEbitenState.released.has(key);
	},
	KeyPressDuration(key) {
		return go2jsEbitenState.held.get(key) || 0;
	},
	IsMouseButtonJustPressed(button) {
		return go2jsEbitenState.mouseHeld.get(button) === 1;
	},
	IsMouseButtonJustReleased(button) {
		return go2jsEbitenState.mouseJustReleased.has(button);
	},
	MouseButtonPressDuration(button) {
		return go2jsEbitenState.mouseHeld.get(button) || 0;
	},
};

const vector = {
	DrawFilledRect(dst, x, y, width, height, color) {
		if (dst.ctx === null) {
			return;
		}

		dst.ctx.fillStyle = go2jsEbitenCSS(color);
		dst.ctx.fillRect(x, y, width, height);
	},
	StrokeRect(dst, x, y, width, height, strokeWidth, color) {
		if (dst.ctx === null) {
			return;
		}

		dst.ctx.strokeStyle = go2jsEbitenCSS(color);
		dst.ctx.lineWidth = strokeWidth;
		dst.ctx.strokeRect(x, y, width, height);
	},
	StrokeLine(dst, x0, y0, x1, y1, strokeWidth, color) {
		if (dst.ctx === null) {
			return;
		}

		dst.ctx.strokeStyle = go2jsEbitenCSS(color);
		dst.ctx.lineWidth = strokeWidth;
		dst.ctx.beginPath();
		dst.ctx.moveTo(x0, y0);
		dst.ctx.lineTo(x1, y1);
		dst.ctx.stroke();
	},
	DrawFilledCircle(dst, cx, cy, r, color) {
		if (dst.ctx === null) {
			return;
		}

		dst.ctx.fillStyle = go2jsEbitenCSS(color);
		dst.ctx.beginPath();
		dst.ctx.arc(cx, cy, r, 0, Math.PI * 2);
		dst.ctx.fill();
	},
	StrokeCircle(dst, cx, cy, r, strokeWidth, color) {
		if (dst.ctx === null) {
			return;
		}

		dst.ctx.strokeStyle = go2jsEbitenCSS(color);
		dst.ctx.lineWidth = strokeWidth;
		dst.ctx.beginPath();
		dst.ctx.arc(cx, cy, r, 0, Math.PI * 2);
		dst.ctx.stroke();
	},
};

const ebitenutil = {
	DebugPrintAt(image, text, x, y) {
		if (image.ctx === null) {
			return;
		}

		const lines = String(text).split("\n");
		const ctx = image.ctx;

		ctx.save();
		ctx.font = "11px monospace";
		ctx.textBaseline = "top";
		ctx.fillStyle = "#fff";

		for (let row = 0; row < lines.length; row++) {
			const chars = Array.from(lines[row]);

			for (let col = 0; col < chars.length; col++) {
				ctx.fillText(chars[col], x + col * 6, y + row * 16 + 2);
			}
		}

		ctx.restore();
	},
	DebugPrint(image, text) {
		ebitenutil.DebugPrintAt(image, text, 0, 0);
	},
};
`
}

// browserProcessSource gives a page the few Node globals the runtime reaches for,
// so that a program that was compiled for Node also runs in a page: what it
// writes to the standard output or error is written to the console, a line at a
// time, and the arguments and the environment are empty.
func browserProcessSource() string {
	return `
(function() {
	if (typeof process !== "undefined" || typeof globalThis === "undefined") {
		return null;
	}

	const decoder = typeof TextDecoder === "function" ? new TextDecoder() : null;

	const lineWriter = function(log) {
		let pending = "";

		return {
			write(value) {
				pending += typeof value === "string" ? value : (decoder !== null ? decoder.decode(value) : String(value));

				let at = pending.indexOf("\n");

				while (at >= 0) {
					log(pending.slice(0, at));
					pending = pending.slice(at + 1);
					at = pending.indexOf("\n");
				}

				return true;
			},
		};
	};

	globalThis.process = {
		stdout: lineWriter((line) => console.log(line)),
		stderr: lineWriter((line) => console.warn(line)),
		env: {},
		argv: ["go2js"],
		platform: "browser",
		exit(code) {
			throw new Error("exit status " + code);
		},
		on() {},
		cwd() {
			return "/";
		},
		hrtime: {bigint() { return BigInt(Math.round(performance.now() * 1e6)); }},
	};

	return globalThis.process;
})();
`
}
