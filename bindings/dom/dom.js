// JavaScript behind the library "dom", written by go2js bind.
// It is not loaded: what it declares is global.
function dom_load() {
	return globalThis;
}

// One JavaScript object is always the same Go value, so two of them compare as they should.
const dom_wrapped = new WeakMap();
function dom_unwrap(value) {
	value = dom_open(value);

	return value === null || value === undefined ? null : value.handle;
}

// whether an object of the library is the class of that name, or a kind of it
function dom_is(value, name) {
	const kind = value.handle && value.handle.constructor ? dom_kinds[value.handle.constructor.name] : undefined;

	return kind !== undefined && kind.lineage.includes(name);
}

// a value Go holds in an interface is boxed; the library gets what is inside
function dom_open(value) {
	return value !== null && typeof value === "object" && value.__go2js_interface === true ? value.value : value;
}

// a value of any type that is one of the library's objects as Go holds it goes back as the object itself
function dom_unwrapAny(value) {
	value = dom_open(value);

	if (value !== null && typeof value === "object" && value.handle !== undefined && value.handle !== null && dom_wrapped.get(value.handle) === value) {
		return value.handle;
	}

	return value;
}

function dom_wrap(Class, handle) {
	if (handle === null || handle === undefined) {
		return null;
	}

	let value = dom_wrapped.get(handle);

	if (value === undefined) {
		// a Mesh the library gives where an Object3D was declared is a Mesh
		const kind = handle.constructor ? dom_kinds[handle.constructor.name] : undefined;

		if (kind !== undefined && kind.lineage.includes(Class.name.slice(6))) {
			Class = kind.Class();
		}

		value = new Class();
		value.handle = handle;
		dom_wrapped.set(handle, value);
	}

	return value;
}

// waiting for a promise: the goroutine parks on a channel that the promise answers
let dom_pending = 0;
let dom_resume = null;
let dom_keep = null;
let dom_previous = null;

function dom_idle(resume) {
	if (resume === null) {
		return dom_previous !== null ? dom_previous(null) : false;
	}

	if (dom_pending > 0) {
		dom_resume = resume;

		return true;
	}

	return dom_previous !== null ? dom_previous(resume) : false;
}

function* dom_await(start) {
	const channel = go2jsChannel(1);
	let outcome = null;

	if (go2jsIdleHook !== dom_idle) {
		dom_previous = go2jsIdleHook;
		go2jsIdleHook = dom_idle;
	}

	dom_pending++;

	if (dom_keep === null && typeof setInterval === "function") {
		dom_keep = setInterval(function() {}, 1 << 30);
	}

	const answer = (result) => {
		outcome = result;
		dom_pending--;

		if (dom_pending === 0 && dom_keep !== null) {
			clearInterval(dom_keep);
			dom_keep = null;
		}

		go2jsChanTrySend(channel, true);

		if (dom_resume !== null) {
			const resume = dom_resume;

			dom_resume = null;
			resume();
		}
	};

	let promise;

	try {
		promise = Promise.resolve(start());
	} catch (error) {
		promise = Promise.reject(error);
	}

	promise.then(
		(value) => answer({ok: true, value: value}),
		(error) => answer({ok: false, error: go2jsErrorsNew(String(error !== null && error !== undefined && error.message !== undefined ? error.message : error))})
	);

	yield* go2jsChanRecv(channel, false);

	return outcome;
}

// the object the library gets for a AssignedNodesOptions
function dom_toAssignedNodesOptions(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("Flatten") || !(value.Flatten === false)) {
		out["flatten"] = value.Flatten;
	}

	return out;
}

// a AssignedNodesOptions of what the library gives
function dom_fromAssignedNodesOptions(object) {
	const value = new dom.AssignedNodesOptions();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["flatten"] !== undefined && object["flatten"] !== null) {
		value.Flatten = object["flatten"];
	}

	return value;
}

// the object the library gets for a AudioTimestamp
function dom_toAudioTimestamp(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("ContextTime") || !(value.ContextTime === 0)) {
		out["contextTime"] = value.ContextTime;
	}
	if (explicit.has("PerformanceTime") || !(value.PerformanceTime === 0)) {
		out["performanceTime"] = value.PerformanceTime;
	}

	return out;
}

// a AudioTimestamp of what the library gives
function dom_fromAudioTimestamp(object) {
	const value = new dom.AudioTimestamp();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["contextTime"] !== undefined && object["contextTime"] !== null) {
		value.ContextTime = object["contextTime"];
	}
	if (object["performanceTime"] !== undefined && object["performanceTime"] !== null) {
		value.PerformanceTime = object["performanceTime"];
	}

	return value;
}

// the object the library gets for a CanvasRenderingContext2DSettings
function dom_toCanvasRenderingContext2DSettings(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("Alpha") || !(value.Alpha === false)) {
		out["alpha"] = value.Alpha;
	}
	if (explicit.has("ColorSpace") || !(value.ColorSpace === "")) {
		out["colorSpace"] = value.ColorSpace;
	}
	if (explicit.has("Desynchronized") || !(value.Desynchronized === false)) {
		out["desynchronized"] = value.Desynchronized;
	}
	if (explicit.has("WillReadFrequently") || !(value.WillReadFrequently === false)) {
		out["willReadFrequently"] = value.WillReadFrequently;
	}

	return out;
}

// a CanvasRenderingContext2DSettings of what the library gives
function dom_fromCanvasRenderingContext2DSettings(object) {
	const value = new dom.CanvasRenderingContext2DSettings();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["alpha"] !== undefined && object["alpha"] !== null) {
		value.Alpha = object["alpha"];
	}
	if (object["colorSpace"] !== undefined && object["colorSpace"] !== null) {
		value.ColorSpace = object["colorSpace"];
	}
	if (object["desynchronized"] !== undefined && object["desynchronized"] !== null) {
		value.Desynchronized = object["desynchronized"];
	}
	if (object["willReadFrequently"] !== undefined && object["willReadFrequently"] !== null) {
		value.WillReadFrequently = object["willReadFrequently"];
	}

	return value;
}

// the object the library gets for a CheckVisibilityOptions
function dom_toCheckVisibilityOptions(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("CheckOpacity") || !(value.CheckOpacity === false)) {
		out["checkOpacity"] = value.CheckOpacity;
	}
	if (explicit.has("CheckVisibilityCSS") || !(value.CheckVisibilityCSS === false)) {
		out["checkVisibilityCSS"] = value.CheckVisibilityCSS;
	}
	if (explicit.has("ContentVisibilityAuto") || !(value.ContentVisibilityAuto === false)) {
		out["contentVisibilityAuto"] = value.ContentVisibilityAuto;
	}
	if (explicit.has("OpacityProperty") || !(value.OpacityProperty === false)) {
		out["opacityProperty"] = value.OpacityProperty;
	}
	if (explicit.has("VisibilityProperty") || !(value.VisibilityProperty === false)) {
		out["visibilityProperty"] = value.VisibilityProperty;
	}

	return out;
}

// a CheckVisibilityOptions of what the library gives
function dom_fromCheckVisibilityOptions(object) {
	const value = new dom.CheckVisibilityOptions();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["checkOpacity"] !== undefined && object["checkOpacity"] !== null) {
		value.CheckOpacity = object["checkOpacity"];
	}
	if (object["checkVisibilityCSS"] !== undefined && object["checkVisibilityCSS"] !== null) {
		value.CheckVisibilityCSS = object["checkVisibilityCSS"];
	}
	if (object["contentVisibilityAuto"] !== undefined && object["contentVisibilityAuto"] !== null) {
		value.ContentVisibilityAuto = object["contentVisibilityAuto"];
	}
	if (object["opacityProperty"] !== undefined && object["opacityProperty"] !== null) {
		value.OpacityProperty = object["opacityProperty"];
	}
	if (object["visibilityProperty"] !== undefined && object["visibilityProperty"] !== null) {
		value.VisibilityProperty = object["visibilityProperty"];
	}

	return value;
}

// the object the library gets for a DOMMatrix2DInit
function dom_toDOMMatrix2DInit(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("A") || !(value.A === 0)) {
		out["a"] = value.A;
	}
	if (explicit.has("B") || !(value.B === 0)) {
		out["b"] = value.B;
	}
	if (explicit.has("C") || !(value.C === 0)) {
		out["c"] = value.C;
	}
	if (explicit.has("D") || !(value.D === 0)) {
		out["d"] = value.D;
	}
	if (explicit.has("E") || !(value.E === 0)) {
		out["e"] = value.E;
	}
	if (explicit.has("F") || !(value.F === 0)) {
		out["f"] = value.F;
	}
	if (explicit.has("M11") || !(value.M11 === 0)) {
		out["m11"] = value.M11;
	}
	if (explicit.has("M12") || !(value.M12 === 0)) {
		out["m12"] = value.M12;
	}
	if (explicit.has("M21") || !(value.M21 === 0)) {
		out["m21"] = value.M21;
	}
	if (explicit.has("M22") || !(value.M22 === 0)) {
		out["m22"] = value.M22;
	}
	if (explicit.has("M41") || !(value.M41 === 0)) {
		out["m41"] = value.M41;
	}
	if (explicit.has("M42") || !(value.M42 === 0)) {
		out["m42"] = value.M42;
	}

	return out;
}

// a DOMMatrix2DInit of what the library gives
function dom_fromDOMMatrix2DInit(object) {
	const value = new dom.DOMMatrix2DInit();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["a"] !== undefined && object["a"] !== null) {
		value.A = object["a"];
	}
	if (object["b"] !== undefined && object["b"] !== null) {
		value.B = object["b"];
	}
	if (object["c"] !== undefined && object["c"] !== null) {
		value.C = object["c"];
	}
	if (object["d"] !== undefined && object["d"] !== null) {
		value.D = object["d"];
	}
	if (object["e"] !== undefined && object["e"] !== null) {
		value.E = object["e"];
	}
	if (object["f"] !== undefined && object["f"] !== null) {
		value.F = object["f"];
	}
	if (object["m11"] !== undefined && object["m11"] !== null) {
		value.M11 = object["m11"];
	}
	if (object["m12"] !== undefined && object["m12"] !== null) {
		value.M12 = object["m12"];
	}
	if (object["m21"] !== undefined && object["m21"] !== null) {
		value.M21 = object["m21"];
	}
	if (object["m22"] !== undefined && object["m22"] !== null) {
		value.M22 = object["m22"];
	}
	if (object["m41"] !== undefined && object["m41"] !== null) {
		value.M41 = object["m41"];
	}
	if (object["m42"] !== undefined && object["m42"] !== null) {
		value.M42 = object["m42"];
	}

	return value;
}

// the object the library gets for a DOMMatrixInit
function dom_toDOMMatrixInit(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("Is2D") || !(value.Is2D === false)) {
		out["is2D"] = value.Is2D;
	}
	if (explicit.has("M13") || !(value.M13 === 0)) {
		out["m13"] = value.M13;
	}
	if (explicit.has("M14") || !(value.M14 === 0)) {
		out["m14"] = value.M14;
	}
	if (explicit.has("M23") || !(value.M23 === 0)) {
		out["m23"] = value.M23;
	}
	if (explicit.has("M24") || !(value.M24 === 0)) {
		out["m24"] = value.M24;
	}
	if (explicit.has("M31") || !(value.M31 === 0)) {
		out["m31"] = value.M31;
	}
	if (explicit.has("M32") || !(value.M32 === 0)) {
		out["m32"] = value.M32;
	}
	if (explicit.has("M33") || !(value.M33 === 0)) {
		out["m33"] = value.M33;
	}
	if (explicit.has("M34") || !(value.M34 === 0)) {
		out["m34"] = value.M34;
	}
	if (explicit.has("M43") || !(value.M43 === 0)) {
		out["m43"] = value.M43;
	}
	if (explicit.has("M44") || !(value.M44 === 0)) {
		out["m44"] = value.M44;
	}
	if (explicit.has("A") || !(value.A === 0)) {
		out["a"] = value.A;
	}
	if (explicit.has("B") || !(value.B === 0)) {
		out["b"] = value.B;
	}
	if (explicit.has("C") || !(value.C === 0)) {
		out["c"] = value.C;
	}
	if (explicit.has("D") || !(value.D === 0)) {
		out["d"] = value.D;
	}
	if (explicit.has("E") || !(value.E === 0)) {
		out["e"] = value.E;
	}
	if (explicit.has("F") || !(value.F === 0)) {
		out["f"] = value.F;
	}
	if (explicit.has("M11") || !(value.M11 === 0)) {
		out["m11"] = value.M11;
	}
	if (explicit.has("M12") || !(value.M12 === 0)) {
		out["m12"] = value.M12;
	}
	if (explicit.has("M21") || !(value.M21 === 0)) {
		out["m21"] = value.M21;
	}
	if (explicit.has("M22") || !(value.M22 === 0)) {
		out["m22"] = value.M22;
	}
	if (explicit.has("M41") || !(value.M41 === 0)) {
		out["m41"] = value.M41;
	}
	if (explicit.has("M42") || !(value.M42 === 0)) {
		out["m42"] = value.M42;
	}

	return out;
}

// a DOMMatrixInit of what the library gives
function dom_fromDOMMatrixInit(object) {
	const value = new dom.DOMMatrixInit();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["is2D"] !== undefined && object["is2D"] !== null) {
		value.Is2D = object["is2D"];
	}
	if (object["m13"] !== undefined && object["m13"] !== null) {
		value.M13 = object["m13"];
	}
	if (object["m14"] !== undefined && object["m14"] !== null) {
		value.M14 = object["m14"];
	}
	if (object["m23"] !== undefined && object["m23"] !== null) {
		value.M23 = object["m23"];
	}
	if (object["m24"] !== undefined && object["m24"] !== null) {
		value.M24 = object["m24"];
	}
	if (object["m31"] !== undefined && object["m31"] !== null) {
		value.M31 = object["m31"];
	}
	if (object["m32"] !== undefined && object["m32"] !== null) {
		value.M32 = object["m32"];
	}
	if (object["m33"] !== undefined && object["m33"] !== null) {
		value.M33 = object["m33"];
	}
	if (object["m34"] !== undefined && object["m34"] !== null) {
		value.M34 = object["m34"];
	}
	if (object["m43"] !== undefined && object["m43"] !== null) {
		value.M43 = object["m43"];
	}
	if (object["m44"] !== undefined && object["m44"] !== null) {
		value.M44 = object["m44"];
	}
	if (object["a"] !== undefined && object["a"] !== null) {
		value.A = object["a"];
	}
	if (object["b"] !== undefined && object["b"] !== null) {
		value.B = object["b"];
	}
	if (object["c"] !== undefined && object["c"] !== null) {
		value.C = object["c"];
	}
	if (object["d"] !== undefined && object["d"] !== null) {
		value.D = object["d"];
	}
	if (object["e"] !== undefined && object["e"] !== null) {
		value.E = object["e"];
	}
	if (object["f"] !== undefined && object["f"] !== null) {
		value.F = object["f"];
	}
	if (object["m11"] !== undefined && object["m11"] !== null) {
		value.M11 = object["m11"];
	}
	if (object["m12"] !== undefined && object["m12"] !== null) {
		value.M12 = object["m12"];
	}
	if (object["m21"] !== undefined && object["m21"] !== null) {
		value.M21 = object["m21"];
	}
	if (object["m22"] !== undefined && object["m22"] !== null) {
		value.M22 = object["m22"];
	}
	if (object["m41"] !== undefined && object["m41"] !== null) {
		value.M41 = object["m41"];
	}
	if (object["m42"] !== undefined && object["m42"] !== null) {
		value.M42 = object["m42"];
	}

	return value;
}

// the object the library gets for a DOMPointInit
function dom_toDOMPointInit(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("W") || !(value.W === 0)) {
		out["w"] = value.W;
	}
	if (explicit.has("X") || !(value.X === 0)) {
		out["x"] = value.X;
	}
	if (explicit.has("Y") || !(value.Y === 0)) {
		out["y"] = value.Y;
	}
	if (explicit.has("Z") || !(value.Z === 0)) {
		out["z"] = value.Z;
	}

	return out;
}

// a DOMPointInit of what the library gives
function dom_fromDOMPointInit(object) {
	const value = new dom.DOMPointInit();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["w"] !== undefined && object["w"] !== null) {
		value.W = object["w"];
	}
	if (object["x"] !== undefined && object["x"] !== null) {
		value.X = object["x"];
	}
	if (object["y"] !== undefined && object["y"] !== null) {
		value.Y = object["y"];
	}
	if (object["z"] !== undefined && object["z"] !== null) {
		value.Z = object["z"];
	}

	return value;
}

// the object the library gets for a ElementCreationOptions
function dom_toElementCreationOptions(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("Is") || !(value.Is === "")) {
		out["is"] = value.Is;
	}

	return out;
}

// a ElementCreationOptions of what the library gives
function dom_fromElementCreationOptions(object) {
	const value = new dom.ElementCreationOptions();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["is"] !== undefined && object["is"] !== null) {
		value.Is = object["is"];
	}

	return value;
}

// the object the library gets for a FocusOptions
function dom_toFocusOptions(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("PreventScroll") || !(value.PreventScroll === false)) {
		out["preventScroll"] = value.PreventScroll;
	}

	return out;
}

// a FocusOptions of what the library gives
function dom_fromFocusOptions(object) {
	const value = new dom.FocusOptions();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["preventScroll"] !== undefined && object["preventScroll"] !== null) {
		value.PreventScroll = object["preventScroll"];
	}

	return value;
}

// the object the library gets for a FullscreenOptions
function dom_toFullscreenOptions(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("NavigationUI") || !(value.NavigationUI === "")) {
		out["navigationUI"] = value.NavigationUI;
	}

	return out;
}

// a FullscreenOptions of what the library gives
function dom_fromFullscreenOptions(object) {
	const value = new dom.FullscreenOptions();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["navigationUI"] !== undefined && object["navigationUI"] !== null) {
		value.NavigationUI = object["navigationUI"];
	}

	return value;
}

// the object the library gets for a GetAnimationsOptions
function dom_toGetAnimationsOptions(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("Subtree") || !(value.Subtree === false)) {
		out["subtree"] = value.Subtree;
	}

	return out;
}

// a GetAnimationsOptions of what the library gives
function dom_fromGetAnimationsOptions(object) {
	const value = new dom.GetAnimationsOptions();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["subtree"] !== undefined && object["subtree"] !== null) {
		value.Subtree = object["subtree"];
	}

	return value;
}

// the object the library gets for a GetHTMLOptions
function dom_toGetHTMLOptions(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("SerializableShadowRoots") || !(value.SerializableShadowRoots === false)) {
		out["serializableShadowRoots"] = value.SerializableShadowRoots;
	}
	if (explicit.has("ShadowRoots") || !(value.ShadowRoots === null || value.ShadowRoots === undefined || value.ShadowRoots.length === 0)) {
		out["shadowRoots"] = Array.from(value.ShadowRoots || [], (item) => dom_unwrap(item));
	}

	return out;
}

// a GetHTMLOptions of what the library gives
function dom_fromGetHTMLOptions(object) {
	const value = new dom.GetHTMLOptions();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["serializableShadowRoots"] !== undefined && object["serializableShadowRoots"] !== null) {
		value.SerializableShadowRoots = object["serializableShadowRoots"];
	}
	if (object["shadowRoots"] !== undefined && object["shadowRoots"] !== null) {
		value.ShadowRoots = Array.from(object["shadowRoots"] || [], (item) => dom_wrap(dom.ShadowRoot, item));
	}

	return value;
}

// the object the library gets for a GetRootNodeOptions
function dom_toGetRootNodeOptions(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("Composed") || !(value.Composed === false)) {
		out["composed"] = value.Composed;
	}

	return out;
}

// a GetRootNodeOptions of what the library gives
function dom_fromGetRootNodeOptions(object) {
	const value = new dom.GetRootNodeOptions();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["composed"] !== undefined && object["composed"] !== null) {
		value.Composed = object["composed"];
	}

	return value;
}

// the object the library gets for a IdleRequestOptions
function dom_toIdleRequestOptions(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("Timeout") || !(value.Timeout === 0)) {
		out["timeout"] = value.Timeout;
	}

	return out;
}

// a IdleRequestOptions of what the library gives
function dom_fromIdleRequestOptions(object) {
	const value = new dom.IdleRequestOptions();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["timeout"] !== undefined && object["timeout"] !== null) {
		value.Timeout = object["timeout"];
	}

	return value;
}

// the object the library gets for a ImageBitmapOptions
function dom_toImageBitmapOptions(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("ColorSpaceConversion") || !(value.ColorSpaceConversion === "")) {
		out["colorSpaceConversion"] = value.ColorSpaceConversion;
	}
	if (explicit.has("ImageOrientation") || !(value.ImageOrientation === "")) {
		out["imageOrientation"] = value.ImageOrientation;
	}
	if (explicit.has("PremultiplyAlpha") || !(value.PremultiplyAlpha === "")) {
		out["premultiplyAlpha"] = value.PremultiplyAlpha;
	}
	if (explicit.has("ResizeHeight") || !(value.ResizeHeight === 0)) {
		out["resizeHeight"] = value.ResizeHeight;
	}
	if (explicit.has("ResizeQuality") || !(value.ResizeQuality === "")) {
		out["resizeQuality"] = value.ResizeQuality;
	}
	if (explicit.has("ResizeWidth") || !(value.ResizeWidth === 0)) {
		out["resizeWidth"] = value.ResizeWidth;
	}

	return out;
}

// a ImageBitmapOptions of what the library gives
function dom_fromImageBitmapOptions(object) {
	const value = new dom.ImageBitmapOptions();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["colorSpaceConversion"] !== undefined && object["colorSpaceConversion"] !== null) {
		value.ColorSpaceConversion = object["colorSpaceConversion"];
	}
	if (object["imageOrientation"] !== undefined && object["imageOrientation"] !== null) {
		value.ImageOrientation = object["imageOrientation"];
	}
	if (object["premultiplyAlpha"] !== undefined && object["premultiplyAlpha"] !== null) {
		value.PremultiplyAlpha = object["premultiplyAlpha"];
	}
	if (object["resizeHeight"] !== undefined && object["resizeHeight"] !== null) {
		value.ResizeHeight = object["resizeHeight"];
	}
	if (object["resizeQuality"] !== undefined && object["resizeQuality"] !== null) {
		value.ResizeQuality = object["resizeQuality"];
	}
	if (object["resizeWidth"] !== undefined && object["resizeWidth"] !== null) {
		value.ResizeWidth = object["resizeWidth"];
	}

	return value;
}

// the object the library gets for a ImageDataSettings
function dom_toImageDataSettings(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("ColorSpace") || !(value.ColorSpace === "")) {
		out["colorSpace"] = value.ColorSpace;
	}

	return out;
}

// a ImageDataSettings of what the library gives
function dom_fromImageDataSettings(object) {
	const value = new dom.ImageDataSettings();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["colorSpace"] !== undefined && object["colorSpace"] !== null) {
		value.ColorSpace = object["colorSpace"];
	}

	return value;
}

// the object the library gets for a ImageEncodeOptions
function dom_toImageEncodeOptions(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("Quality") || !(value.Quality === 0)) {
		out["quality"] = value.Quality;
	}
	if (explicit.has("Type") || !(value.Type === "")) {
		out["type"] = value.Type;
	}

	return out;
}

// a ImageEncodeOptions of what the library gives
function dom_fromImageEncodeOptions(object) {
	const value = new dom.ImageEncodeOptions();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["quality"] !== undefined && object["quality"] !== null) {
		value.Quality = object["quality"];
	}
	if (object["type"] !== undefined && object["type"] !== null) {
		value.Type = object["type"];
	}

	return value;
}

// the object the library gets for a MIDIOptions
function dom_toMIDIOptions(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("Software") || !(value.Software === false)) {
		out["software"] = value.Software;
	}
	if (explicit.has("Sysex") || !(value.Sysex === false)) {
		out["sysex"] = value.Sysex;
	}

	return out;
}

// a MIDIOptions of what the library gives
function dom_fromMIDIOptions(object) {
	const value = new dom.MIDIOptions();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["software"] !== undefined && object["software"] !== null) {
		value.Software = object["software"];
	}
	if (object["sysex"] !== undefined && object["sysex"] !== null) {
		value.Sysex = object["sysex"];
	}

	return value;
}

// the object the library gets for a MediaKeySystemConfiguration
function dom_toMediaKeySystemConfiguration(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("AudioCapabilities") || !(value.AudioCapabilities === null || value.AudioCapabilities === undefined || value.AudioCapabilities.length === 0)) {
		out["audioCapabilities"] = Array.from(value.AudioCapabilities || [], (item) => dom_unwrapAny(item));
	}
	if (explicit.has("DistinctiveIdentifier") || !(value.DistinctiveIdentifier === "")) {
		out["distinctiveIdentifier"] = value.DistinctiveIdentifier;
	}
	if (explicit.has("InitDataTypes") || !(value.InitDataTypes === null || value.InitDataTypes === undefined || value.InitDataTypes.length === 0)) {
		out["initDataTypes"] = Array.from(value.InitDataTypes || []);
	}
	if (explicit.has("Label") || !(value.Label === "")) {
		out["label"] = value.Label;
	}
	if (explicit.has("PersistentState") || !(value.PersistentState === "")) {
		out["persistentState"] = value.PersistentState;
	}
	if (explicit.has("SessionTypes") || !(value.SessionTypes === null || value.SessionTypes === undefined || value.SessionTypes.length === 0)) {
		out["sessionTypes"] = Array.from(value.SessionTypes || []);
	}
	if (explicit.has("VideoCapabilities") || !(value.VideoCapabilities === null || value.VideoCapabilities === undefined || value.VideoCapabilities.length === 0)) {
		out["videoCapabilities"] = Array.from(value.VideoCapabilities || [], (item) => dom_unwrapAny(item));
	}

	return out;
}

// a MediaKeySystemConfiguration of what the library gives
function dom_fromMediaKeySystemConfiguration(object) {
	const value = new dom.MediaKeySystemConfiguration();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["audioCapabilities"] !== undefined && object["audioCapabilities"] !== null) {
		value.AudioCapabilities = Array.from(object["audioCapabilities"] || [], (item) => item);
	}
	if (object["distinctiveIdentifier"] !== undefined && object["distinctiveIdentifier"] !== null) {
		value.DistinctiveIdentifier = object["distinctiveIdentifier"];
	}
	if (object["initDataTypes"] !== undefined && object["initDataTypes"] !== null) {
		value.InitDataTypes = Array.from(object["initDataTypes"] || []);
	}
	if (object["label"] !== undefined && object["label"] !== null) {
		value.Label = object["label"];
	}
	if (object["persistentState"] !== undefined && object["persistentState"] !== null) {
		value.PersistentState = object["persistentState"];
	}
	if (object["sessionTypes"] !== undefined && object["sessionTypes"] !== null) {
		value.SessionTypes = Array.from(object["sessionTypes"] || []);
	}
	if (object["videoCapabilities"] !== undefined && object["videoCapabilities"] !== null) {
		value.VideoCapabilities = Array.from(object["videoCapabilities"] || [], (item) => item);
	}

	return value;
}

// the object the library gets for a MediaKeysPolicy
function dom_toMediaKeysPolicy(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("MinHdcpVersion") || !(value.MinHdcpVersion === "")) {
		out["minHdcpVersion"] = value.MinHdcpVersion;
	}

	return out;
}

// a MediaKeysPolicy of what the library gives
function dom_fromMediaKeysPolicy(object) {
	const value = new dom.MediaKeysPolicy();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["minHdcpVersion"] !== undefined && object["minHdcpVersion"] !== null) {
		value.MinHdcpVersion = object["minHdcpVersion"];
	}

	return value;
}

// the object the library gets for a PerformanceMarkOptions
function dom_toPerformanceMarkOptions(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("Detail") || !(value.Detail === null || value.Detail === undefined)) {
		out["detail"] = dom_unwrapAny(value.Detail);
	}
	if (explicit.has("StartTime") || !(value.StartTime === 0)) {
		out["startTime"] = value.StartTime;
	}

	return out;
}

// a PerformanceMarkOptions of what the library gives
function dom_fromPerformanceMarkOptions(object) {
	const value = new dom.PerformanceMarkOptions();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["detail"] !== undefined && object["detail"] !== null) {
		value.Detail = object["detail"];
	}
	if (object["startTime"] !== undefined && object["startTime"] !== null) {
		value.StartTime = object["startTime"];
	}

	return value;
}

// the object the library gets for a PeriodicWaveConstraints
function dom_toPeriodicWaveConstraints(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("DisableNormalization") || !(value.DisableNormalization === false)) {
		out["disableNormalization"] = value.DisableNormalization;
	}

	return out;
}

// a PeriodicWaveConstraints of what the library gives
function dom_fromPeriodicWaveConstraints(object) {
	const value = new dom.PeriodicWaveConstraints();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["disableNormalization"] !== undefined && object["disableNormalization"] !== null) {
		value.DisableNormalization = object["disableNormalization"];
	}

	return value;
}

// the object the library gets for a PointerLockOptions
function dom_toPointerLockOptions(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("UnadjustedMovement") || !(value.UnadjustedMovement === false)) {
		out["unadjustedMovement"] = value.UnadjustedMovement;
	}

	return out;
}

// a PointerLockOptions of what the library gives
function dom_fromPointerLockOptions(object) {
	const value = new dom.PointerLockOptions();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["unadjustedMovement"] !== undefined && object["unadjustedMovement"] !== null) {
		value.UnadjustedMovement = object["unadjustedMovement"];
	}

	return value;
}

// the object the library gets for a RequestInit
function dom_toRequestInit(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("Body") || !(value.Body === null || value.Body === undefined)) {
		out["body"] = dom_unwrapAny(value.Body);
	}
	if (explicit.has("Cache") || !(value.Cache === "")) {
		out["cache"] = value.Cache;
	}
	if (explicit.has("Credentials") || !(value.Credentials === "")) {
		out["credentials"] = value.Credentials;
	}
	if (explicit.has("Headers") || !(value.Headers === null || value.Headers === undefined)) {
		out["headers"] = dom_unwrapAny(value.Headers);
	}
	if (explicit.has("Integrity") || !(value.Integrity === "")) {
		out["integrity"] = value.Integrity;
	}
	if (explicit.has("Keepalive") || !(value.Keepalive === false)) {
		out["keepalive"] = value.Keepalive;
	}
	if (explicit.has("Method") || !(value.Method === "")) {
		out["method"] = value.Method;
	}
	if (explicit.has("Mode") || !(value.Mode === "")) {
		out["mode"] = value.Mode;
	}
	if (explicit.has("Priority") || !(value.Priority === "")) {
		out["priority"] = value.Priority;
	}
	if (explicit.has("Redirect") || !(value.Redirect === "")) {
		out["redirect"] = value.Redirect;
	}
	if (explicit.has("Referrer") || !(value.Referrer === "")) {
		out["referrer"] = value.Referrer;
	}
	if (explicit.has("ReferrerPolicy") || !(value.ReferrerPolicy === "")) {
		out["referrerPolicy"] = value.ReferrerPolicy;
	}
	if (explicit.has("Signal") || !(value.Signal === null || value.Signal === undefined)) {
		out["signal"] = dom_unwrap(value.Signal);
	}

	return out;
}

// a RequestInit of what the library gives
function dom_fromRequestInit(object) {
	const value = new dom.RequestInit();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["body"] !== undefined && object["body"] !== null) {
		value.Body = object["body"];
	}
	if (object["cache"] !== undefined && object["cache"] !== null) {
		value.Cache = object["cache"];
	}
	if (object["credentials"] !== undefined && object["credentials"] !== null) {
		value.Credentials = object["credentials"];
	}
	if (object["headers"] !== undefined && object["headers"] !== null) {
		value.Headers = object["headers"];
	}
	if (object["integrity"] !== undefined && object["integrity"] !== null) {
		value.Integrity = object["integrity"];
	}
	if (object["keepalive"] !== undefined && object["keepalive"] !== null) {
		value.Keepalive = object["keepalive"];
	}
	if (object["method"] !== undefined && object["method"] !== null) {
		value.Method = object["method"];
	}
	if (object["mode"] !== undefined && object["mode"] !== null) {
		value.Mode = object["mode"];
	}
	if (object["priority"] !== undefined && object["priority"] !== null) {
		value.Priority = object["priority"];
	}
	if (object["redirect"] !== undefined && object["redirect"] !== null) {
		value.Redirect = object["redirect"];
	}
	if (object["referrer"] !== undefined && object["referrer"] !== null) {
		value.Referrer = object["referrer"];
	}
	if (object["referrerPolicy"] !== undefined && object["referrerPolicy"] !== null) {
		value.ReferrerPolicy = object["referrerPolicy"];
	}
	if (object["signal"] !== undefined && object["signal"] !== null) {
		value.Signal = dom_wrap(dom.AbortSignal, object["signal"]);
	}

	return value;
}

// the object the library gets for a ScrollToOptions
function dom_toScrollToOptions(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("Left") || !(value.Left === 0)) {
		out["left"] = value.Left;
	}
	if (explicit.has("Top") || !(value.Top === 0)) {
		out["top"] = value.Top;
	}
	if (explicit.has("Behavior") || !(value.Behavior === "")) {
		out["behavior"] = value.Behavior;
	}

	return out;
}

// a ScrollToOptions of what the library gives
function dom_fromScrollToOptions(object) {
	const value = new dom.ScrollToOptions();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["left"] !== undefined && object["left"] !== null) {
		value.Left = object["left"];
	}
	if (object["top"] !== undefined && object["top"] !== null) {
		value.Top = object["top"];
	}
	if (object["behavior"] !== undefined && object["behavior"] !== null) {
		value.Behavior = object["behavior"];
	}

	return value;
}

// the object the library gets for a ShadowRootInit
function dom_toShadowRootInit(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("DelegatesFocus") || !(value.DelegatesFocus === false)) {
		out["delegatesFocus"] = value.DelegatesFocus;
	}
	if (explicit.has("Mode") || !(value.Mode === "")) {
		out["mode"] = value.Mode;
	}
	if (explicit.has("Serializable") || !(value.Serializable === false)) {
		out["serializable"] = value.Serializable;
	}
	if (explicit.has("SlotAssignment") || !(value.SlotAssignment === "")) {
		out["slotAssignment"] = value.SlotAssignment;
	}

	return out;
}

// a ShadowRootInit of what the library gives
function dom_fromShadowRootInit(object) {
	const value = new dom.ShadowRootInit();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["delegatesFocus"] !== undefined && object["delegatesFocus"] !== null) {
		value.DelegatesFocus = object["delegatesFocus"];
	}
	if (object["mode"] !== undefined && object["mode"] !== null) {
		value.Mode = object["mode"];
	}
	if (object["serializable"] !== undefined && object["serializable"] !== null) {
		value.Serializable = object["serializable"];
	}
	if (object["slotAssignment"] !== undefined && object["slotAssignment"] !== null) {
		value.SlotAssignment = object["slotAssignment"];
	}

	return value;
}

// the object the library gets for a ShareData
function dom_toShareData(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("Files") || !(value.Files === null || value.Files === undefined || value.Files.length === 0)) {
		out["files"] = Array.from(value.Files || [], (item) => dom_unwrapAny(item));
	}
	if (explicit.has("Text") || !(value.Text === "")) {
		out["text"] = value.Text;
	}
	if (explicit.has("Title") || !(value.Title === "")) {
		out["title"] = value.Title;
	}
	if (explicit.has("Url") || !(value.Url === "")) {
		out["url"] = value.Url;
	}

	return out;
}

// a ShareData of what the library gives
function dom_fromShareData(object) {
	const value = new dom.ShareData();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["files"] !== undefined && object["files"] !== null) {
		value.Files = Array.from(object["files"] || [], (item) => item);
	}
	if (object["text"] !== undefined && object["text"] !== null) {
		value.Text = object["text"];
	}
	if (object["title"] !== undefined && object["title"] !== null) {
		value.Title = object["title"];
	}
	if (object["url"] !== undefined && object["url"] !== null) {
		value.Url = object["url"];
	}

	return value;
}

// the object the library gets for a StructuredSerializeOptions
function dom_toStructuredSerializeOptions(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("Transfer") || !(value.Transfer === null || value.Transfer === undefined || value.Transfer.length === 0)) {
		out["transfer"] = Array.from(value.Transfer || [], (item) => dom_unwrapAny(item));
	}

	return out;
}

// a StructuredSerializeOptions of what the library gives
function dom_fromStructuredSerializeOptions(object) {
	const value = new dom.StructuredSerializeOptions();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["transfer"] !== undefined && object["transfer"] !== null) {
		value.Transfer = Array.from(object["transfer"] || [], (item) => item);
	}

	return value;
}

// the object the library gets for a ValidityStateFlags
function dom_toValidityStateFlags(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("BadInput") || !(value.BadInput === false)) {
		out["badInput"] = value.BadInput;
	}
	if (explicit.has("CustomError") || !(value.CustomError === false)) {
		out["customError"] = value.CustomError;
	}
	if (explicit.has("PatternMismatch") || !(value.PatternMismatch === false)) {
		out["patternMismatch"] = value.PatternMismatch;
	}
	if (explicit.has("RangeOverflow") || !(value.RangeOverflow === false)) {
		out["rangeOverflow"] = value.RangeOverflow;
	}
	if (explicit.has("RangeUnderflow") || !(value.RangeUnderflow === false)) {
		out["rangeUnderflow"] = value.RangeUnderflow;
	}
	if (explicit.has("StepMismatch") || !(value.StepMismatch === false)) {
		out["stepMismatch"] = value.StepMismatch;
	}
	if (explicit.has("TooLong") || !(value.TooLong === false)) {
		out["tooLong"] = value.TooLong;
	}
	if (explicit.has("TooShort") || !(value.TooShort === false)) {
		out["tooShort"] = value.TooShort;
	}
	if (explicit.has("TypeMismatch") || !(value.TypeMismatch === false)) {
		out["typeMismatch"] = value.TypeMismatch;
	}
	if (explicit.has("ValueMissing") || !(value.ValueMissing === false)) {
		out["valueMissing"] = value.ValueMissing;
	}

	return out;
}

// a ValidityStateFlags of what the library gives
function dom_fromValidityStateFlags(object) {
	const value = new dom.ValidityStateFlags();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["badInput"] !== undefined && object["badInput"] !== null) {
		value.BadInput = object["badInput"];
	}
	if (object["customError"] !== undefined && object["customError"] !== null) {
		value.CustomError = object["customError"];
	}
	if (object["patternMismatch"] !== undefined && object["patternMismatch"] !== null) {
		value.PatternMismatch = object["patternMismatch"];
	}
	if (object["rangeOverflow"] !== undefined && object["rangeOverflow"] !== null) {
		value.RangeOverflow = object["rangeOverflow"];
	}
	if (object["rangeUnderflow"] !== undefined && object["rangeUnderflow"] !== null) {
		value.RangeUnderflow = object["rangeUnderflow"];
	}
	if (object["stepMismatch"] !== undefined && object["stepMismatch"] !== null) {
		value.StepMismatch = object["stepMismatch"];
	}
	if (object["tooLong"] !== undefined && object["tooLong"] !== null) {
		value.TooLong = object["tooLong"];
	}
	if (object["tooShort"] !== undefined && object["tooShort"] !== null) {
		value.TooShort = object["tooShort"];
	}
	if (object["typeMismatch"] !== undefined && object["typeMismatch"] !== null) {
		value.TypeMismatch = object["typeMismatch"];
	}
	if (object["valueMissing"] !== undefined && object["valueMissing"] !== null) {
		value.ValueMissing = object["valueMissing"];
	}

	return value;
}

// the object the library gets for a WorkletOptions
function dom_toWorkletOptions(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("Credentials") || !(value.Credentials === "")) {
		out["credentials"] = value.Credentials;
	}

	return out;
}

// a WorkletOptions of what the library gives
function dom_fromWorkletOptions(object) {
	const value = new dom.WorkletOptions();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["credentials"] !== undefined && object["credentials"] !== null) {
		value.Credentials = object["credentials"];
	}

	return value;
}

// the object the library gets for a AnimationTimeline
function dom_toAnimationTimeline(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("CurrentTime") || !(value.CurrentTime === null || value.CurrentTime === undefined)) {
		out["currentTime"] = dom_unwrapAny(value.CurrentTime);
	}

	return out;
}

// a AnimationTimeline of what the library gives
function dom_fromAnimationTimeline(object) {
	const value = new dom.AnimationTimeline();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["currentTime"] !== undefined && object["currentTime"] !== null) {
		value.CurrentTime = object["currentTime"];
	}

	return value;
}

// the object the library gets for a BarProp
function dom_toBarProp(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("Visible") || !(value.Visible === false)) {
		out["visible"] = value.Visible;
	}

	return out;
}

// a BarProp of what the library gives
function dom_fromBarProp(object) {
	const value = new dom.BarProp();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["visible"] !== undefined && object["visible"] !== null) {
		value.Visible = object["visible"];
	}

	return value;
}

// the object the library gets for a DOMException
function dom_toDOMException(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("Code") || !(value.Code === 0)) {
		out["code"] = value.Code;
	}
	if (explicit.has("Message") || !(value.Message === "")) {
		out["message"] = value.Message;
	}
	if (explicit.has("Name") || !(value.Name === "")) {
		out["name"] = value.Name;
	}
	if (explicit.has("INDEX_SIZE_ERR") || !(value.INDEX_SIZE_ERR === 0)) {
		out["INDEX_SIZE_ERR"] = value.INDEX_SIZE_ERR;
	}
	if (explicit.has("DOMSTRING_SIZE_ERR") || !(value.DOMSTRING_SIZE_ERR === 0)) {
		out["DOMSTRING_SIZE_ERR"] = value.DOMSTRING_SIZE_ERR;
	}
	if (explicit.has("HIERARCHY_REQUEST_ERR") || !(value.HIERARCHY_REQUEST_ERR === 0)) {
		out["HIERARCHY_REQUEST_ERR"] = value.HIERARCHY_REQUEST_ERR;
	}
	if (explicit.has("WRONG_DOCUMENT_ERR") || !(value.WRONG_DOCUMENT_ERR === 0)) {
		out["WRONG_DOCUMENT_ERR"] = value.WRONG_DOCUMENT_ERR;
	}
	if (explicit.has("INVALID_CHARACTER_ERR") || !(value.INVALID_CHARACTER_ERR === 0)) {
		out["INVALID_CHARACTER_ERR"] = value.INVALID_CHARACTER_ERR;
	}
	if (explicit.has("NO_DATA_ALLOWED_ERR") || !(value.NO_DATA_ALLOWED_ERR === 0)) {
		out["NO_DATA_ALLOWED_ERR"] = value.NO_DATA_ALLOWED_ERR;
	}
	if (explicit.has("NO_MODIFICATION_ALLOWED_ERR") || !(value.NO_MODIFICATION_ALLOWED_ERR === 0)) {
		out["NO_MODIFICATION_ALLOWED_ERR"] = value.NO_MODIFICATION_ALLOWED_ERR;
	}
	if (explicit.has("NOT_FOUND_ERR") || !(value.NOT_FOUND_ERR === 0)) {
		out["NOT_FOUND_ERR"] = value.NOT_FOUND_ERR;
	}
	if (explicit.has("NOT_SUPPORTED_ERR") || !(value.NOT_SUPPORTED_ERR === 0)) {
		out["NOT_SUPPORTED_ERR"] = value.NOT_SUPPORTED_ERR;
	}
	if (explicit.has("INUSE_ATTRIBUTE_ERR") || !(value.INUSE_ATTRIBUTE_ERR === 0)) {
		out["INUSE_ATTRIBUTE_ERR"] = value.INUSE_ATTRIBUTE_ERR;
	}
	if (explicit.has("INVALID_STATE_ERR") || !(value.INVALID_STATE_ERR === 0)) {
		out["INVALID_STATE_ERR"] = value.INVALID_STATE_ERR;
	}
	if (explicit.has("SYNTAX_ERR") || !(value.SYNTAX_ERR === 0)) {
		out["SYNTAX_ERR"] = value.SYNTAX_ERR;
	}
	if (explicit.has("INVALID_MODIFICATION_ERR") || !(value.INVALID_MODIFICATION_ERR === 0)) {
		out["INVALID_MODIFICATION_ERR"] = value.INVALID_MODIFICATION_ERR;
	}
	if (explicit.has("NAMESPACE_ERR") || !(value.NAMESPACE_ERR === 0)) {
		out["NAMESPACE_ERR"] = value.NAMESPACE_ERR;
	}
	if (explicit.has("INVALID_ACCESS_ERR") || !(value.INVALID_ACCESS_ERR === 0)) {
		out["INVALID_ACCESS_ERR"] = value.INVALID_ACCESS_ERR;
	}
	if (explicit.has("VALIDATION_ERR") || !(value.VALIDATION_ERR === 0)) {
		out["VALIDATION_ERR"] = value.VALIDATION_ERR;
	}
	if (explicit.has("TYPE_MISMATCH_ERR") || !(value.TYPE_MISMATCH_ERR === 0)) {
		out["TYPE_MISMATCH_ERR"] = value.TYPE_MISMATCH_ERR;
	}
	if (explicit.has("SECURITY_ERR") || !(value.SECURITY_ERR === 0)) {
		out["SECURITY_ERR"] = value.SECURITY_ERR;
	}
	if (explicit.has("NETWORK_ERR") || !(value.NETWORK_ERR === 0)) {
		out["NETWORK_ERR"] = value.NETWORK_ERR;
	}
	if (explicit.has("ABORT_ERR") || !(value.ABORT_ERR === 0)) {
		out["ABORT_ERR"] = value.ABORT_ERR;
	}
	if (explicit.has("URL_MISMATCH_ERR") || !(value.URL_MISMATCH_ERR === 0)) {
		out["URL_MISMATCH_ERR"] = value.URL_MISMATCH_ERR;
	}
	if (explicit.has("QUOTA_EXCEEDED_ERR") || !(value.QUOTA_EXCEEDED_ERR === 0)) {
		out["QUOTA_EXCEEDED_ERR"] = value.QUOTA_EXCEEDED_ERR;
	}
	if (explicit.has("TIMEOUT_ERR") || !(value.TIMEOUT_ERR === 0)) {
		out["TIMEOUT_ERR"] = value.TIMEOUT_ERR;
	}
	if (explicit.has("INVALID_NODE_TYPE_ERR") || !(value.INVALID_NODE_TYPE_ERR === 0)) {
		out["INVALID_NODE_TYPE_ERR"] = value.INVALID_NODE_TYPE_ERR;
	}
	if (explicit.has("DATA_CLONE_ERR") || !(value.DATA_CLONE_ERR === 0)) {
		out["DATA_CLONE_ERR"] = value.DATA_CLONE_ERR;
	}

	return out;
}

// a DOMException of what the library gives
function dom_fromDOMException(object) {
	const value = new dom.DOMException();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["code"] !== undefined && object["code"] !== null) {
		value.Code = object["code"];
	}
	if (object["message"] !== undefined && object["message"] !== null) {
		value.Message = object["message"];
	}
	if (object["name"] !== undefined && object["name"] !== null) {
		value.Name = object["name"];
	}
	if (object["INDEX_SIZE_ERR"] !== undefined && object["INDEX_SIZE_ERR"] !== null) {
		value.INDEX_SIZE_ERR = object["INDEX_SIZE_ERR"];
	}
	if (object["DOMSTRING_SIZE_ERR"] !== undefined && object["DOMSTRING_SIZE_ERR"] !== null) {
		value.DOMSTRING_SIZE_ERR = object["DOMSTRING_SIZE_ERR"];
	}
	if (object["HIERARCHY_REQUEST_ERR"] !== undefined && object["HIERARCHY_REQUEST_ERR"] !== null) {
		value.HIERARCHY_REQUEST_ERR = object["HIERARCHY_REQUEST_ERR"];
	}
	if (object["WRONG_DOCUMENT_ERR"] !== undefined && object["WRONG_DOCUMENT_ERR"] !== null) {
		value.WRONG_DOCUMENT_ERR = object["WRONG_DOCUMENT_ERR"];
	}
	if (object["INVALID_CHARACTER_ERR"] !== undefined && object["INVALID_CHARACTER_ERR"] !== null) {
		value.INVALID_CHARACTER_ERR = object["INVALID_CHARACTER_ERR"];
	}
	if (object["NO_DATA_ALLOWED_ERR"] !== undefined && object["NO_DATA_ALLOWED_ERR"] !== null) {
		value.NO_DATA_ALLOWED_ERR = object["NO_DATA_ALLOWED_ERR"];
	}
	if (object["NO_MODIFICATION_ALLOWED_ERR"] !== undefined && object["NO_MODIFICATION_ALLOWED_ERR"] !== null) {
		value.NO_MODIFICATION_ALLOWED_ERR = object["NO_MODIFICATION_ALLOWED_ERR"];
	}
	if (object["NOT_FOUND_ERR"] !== undefined && object["NOT_FOUND_ERR"] !== null) {
		value.NOT_FOUND_ERR = object["NOT_FOUND_ERR"];
	}
	if (object["NOT_SUPPORTED_ERR"] !== undefined && object["NOT_SUPPORTED_ERR"] !== null) {
		value.NOT_SUPPORTED_ERR = object["NOT_SUPPORTED_ERR"];
	}
	if (object["INUSE_ATTRIBUTE_ERR"] !== undefined && object["INUSE_ATTRIBUTE_ERR"] !== null) {
		value.INUSE_ATTRIBUTE_ERR = object["INUSE_ATTRIBUTE_ERR"];
	}
	if (object["INVALID_STATE_ERR"] !== undefined && object["INVALID_STATE_ERR"] !== null) {
		value.INVALID_STATE_ERR = object["INVALID_STATE_ERR"];
	}
	if (object["SYNTAX_ERR"] !== undefined && object["SYNTAX_ERR"] !== null) {
		value.SYNTAX_ERR = object["SYNTAX_ERR"];
	}
	if (object["INVALID_MODIFICATION_ERR"] !== undefined && object["INVALID_MODIFICATION_ERR"] !== null) {
		value.INVALID_MODIFICATION_ERR = object["INVALID_MODIFICATION_ERR"];
	}
	if (object["NAMESPACE_ERR"] !== undefined && object["NAMESPACE_ERR"] !== null) {
		value.NAMESPACE_ERR = object["NAMESPACE_ERR"];
	}
	if (object["INVALID_ACCESS_ERR"] !== undefined && object["INVALID_ACCESS_ERR"] !== null) {
		value.INVALID_ACCESS_ERR = object["INVALID_ACCESS_ERR"];
	}
	if (object["VALIDATION_ERR"] !== undefined && object["VALIDATION_ERR"] !== null) {
		value.VALIDATION_ERR = object["VALIDATION_ERR"];
	}
	if (object["TYPE_MISMATCH_ERR"] !== undefined && object["TYPE_MISMATCH_ERR"] !== null) {
		value.TYPE_MISMATCH_ERR = object["TYPE_MISMATCH_ERR"];
	}
	if (object["SECURITY_ERR"] !== undefined && object["SECURITY_ERR"] !== null) {
		value.SECURITY_ERR = object["SECURITY_ERR"];
	}
	if (object["NETWORK_ERR"] !== undefined && object["NETWORK_ERR"] !== null) {
		value.NETWORK_ERR = object["NETWORK_ERR"];
	}
	if (object["ABORT_ERR"] !== undefined && object["ABORT_ERR"] !== null) {
		value.ABORT_ERR = object["ABORT_ERR"];
	}
	if (object["URL_MISMATCH_ERR"] !== undefined && object["URL_MISMATCH_ERR"] !== null) {
		value.URL_MISMATCH_ERR = object["URL_MISMATCH_ERR"];
	}
	if (object["QUOTA_EXCEEDED_ERR"] !== undefined && object["QUOTA_EXCEEDED_ERR"] !== null) {
		value.QUOTA_EXCEEDED_ERR = object["QUOTA_EXCEEDED_ERR"];
	}
	if (object["TIMEOUT_ERR"] !== undefined && object["TIMEOUT_ERR"] !== null) {
		value.TIMEOUT_ERR = object["TIMEOUT_ERR"];
	}
	if (object["INVALID_NODE_TYPE_ERR"] !== undefined && object["INVALID_NODE_TYPE_ERR"] !== null) {
		value.INVALID_NODE_TYPE_ERR = object["INVALID_NODE_TYPE_ERR"];
	}
	if (object["DATA_CLONE_ERR"] !== undefined && object["DATA_CLONE_ERR"] !== null) {
		value.DATA_CLONE_ERR = object["DATA_CLONE_ERR"];
	}

	return value;
}

// the object the library gets for a DocumentTimeline
function dom_toDocumentTimeline(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("CurrentTime") || !(value.CurrentTime === null || value.CurrentTime === undefined)) {
		out["currentTime"] = dom_unwrapAny(value.CurrentTime);
	}

	return out;
}

// a DocumentTimeline of what the library gives
function dom_fromDocumentTimeline(object) {
	const value = new dom.DocumentTimeline();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["currentTime"] !== undefined && object["currentTime"] !== null) {
		value.CurrentTime = object["currentTime"];
	}

	return value;
}

// the object the library gets for a Gamepad
function dom_toGamepad(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("Axes") || !(value.Axes === null || value.Axes === undefined || value.Axes.length === 0)) {
		out["axes"] = Array.from(value.Axes || []);
	}
	if (explicit.has("Buttons") || !(value.Buttons === null || value.Buttons === undefined || value.Buttons.length === 0)) {
		out["buttons"] = Array.from(value.Buttons || [], (item) => dom_unwrapAny(item));
	}
	if (explicit.has("Connected") || !(value.Connected === false)) {
		out["connected"] = value.Connected;
	}
	if (explicit.has("Id") || !(value.Id === "")) {
		out["id"] = value.Id;
	}
	if (explicit.has("Index") || !(value.Index === 0)) {
		out["index"] = value.Index;
	}
	if (explicit.has("Mapping") || !(value.Mapping === "")) {
		out["mapping"] = value.Mapping;
	}
	if (explicit.has("Timestamp") || !(value.Timestamp === 0)) {
		out["timestamp"] = value.Timestamp;
	}
	if (explicit.has("VibrationActuator") || !(value.VibrationActuator === null || value.VibrationActuator === undefined)) {
		out["vibrationActuator"] = dom_unwrapAny(value.VibrationActuator);
	}

	return out;
}

// a Gamepad of what the library gives
function dom_fromGamepad(object) {
	const value = new dom.Gamepad();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["axes"] !== undefined && object["axes"] !== null) {
		value.Axes = Array.from(object["axes"] || []);
	}
	if (object["buttons"] !== undefined && object["buttons"] !== null) {
		value.Buttons = Array.from(object["buttons"] || [], (item) => item);
	}
	if (object["connected"] !== undefined && object["connected"] !== null) {
		value.Connected = object["connected"];
	}
	if (object["id"] !== undefined && object["id"] !== null) {
		value.Id = object["id"];
	}
	if (object["index"] !== undefined && object["index"] !== null) {
		value.Index = object["index"];
	}
	if (object["mapping"] !== undefined && object["mapping"] !== null) {
		value.Mapping = object["mapping"];
	}
	if (object["timestamp"] !== undefined && object["timestamp"] !== null) {
		value.Timestamp = object["timestamp"];
	}
	if (object["vibrationActuator"] !== undefined && object["vibrationActuator"] !== null) {
		value.VibrationActuator = object["vibrationActuator"];
	}

	return value;
}

// the object the library gets for a ImageData
function dom_toImageData(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("ColorSpace") || !(value.ColorSpace === "")) {
		out["colorSpace"] = value.ColorSpace;
	}
	if (explicit.has("Data") || !(value.Data === null || value.Data === undefined || value.Data.length === 0)) {
		out["data"] = Uint8ClampedArray.from(value.Data || []);
	}
	if (explicit.has("Height") || !(value.Height === 0)) {
		out["height"] = value.Height;
	}
	if (explicit.has("Width") || !(value.Width === 0)) {
		out["width"] = value.Width;
	}

	return out;
}

// a ImageData of what the library gives
function dom_fromImageData(object) {
	const value = new dom.ImageData();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["colorSpace"] !== undefined && object["colorSpace"] !== null) {
		value.ColorSpace = object["colorSpace"];
	}
	if (object["data"] !== undefined && object["data"] !== null) {
		value.Data = Array.from(object["data"] || []);
	}
	if (object["height"] !== undefined && object["height"] !== null) {
		value.Height = object["height"];
	}
	if (object["width"] !== undefined && object["width"] !== null) {
		value.Width = object["width"];
	}

	return value;
}

// the object the library gets for a MediaError
function dom_toMediaError(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("Code") || !(value.Code === 0)) {
		out["code"] = value.Code;
	}
	if (explicit.has("Message") || !(value.Message === "")) {
		out["message"] = value.Message;
	}
	if (explicit.has("MEDIA_ERR_ABORTED") || !(value.MEDIA_ERR_ABORTED === 0)) {
		out["MEDIA_ERR_ABORTED"] = value.MEDIA_ERR_ABORTED;
	}
	if (explicit.has("MEDIA_ERR_NETWORK") || !(value.MEDIA_ERR_NETWORK === 0)) {
		out["MEDIA_ERR_NETWORK"] = value.MEDIA_ERR_NETWORK;
	}
	if (explicit.has("MEDIA_ERR_DECODE") || !(value.MEDIA_ERR_DECODE === 0)) {
		out["MEDIA_ERR_DECODE"] = value.MEDIA_ERR_DECODE;
	}
	if (explicit.has("MEDIA_ERR_SRC_NOT_SUPPORTED") || !(value.MEDIA_ERR_SRC_NOT_SUPPORTED === 0)) {
		out["MEDIA_ERR_SRC_NOT_SUPPORTED"] = value.MEDIA_ERR_SRC_NOT_SUPPORTED;
	}

	return out;
}

// a MediaError of what the library gives
function dom_fromMediaError(object) {
	const value = new dom.MediaError();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["code"] !== undefined && object["code"] !== null) {
		value.Code = object["code"];
	}
	if (object["message"] !== undefined && object["message"] !== null) {
		value.Message = object["message"];
	}
	if (object["MEDIA_ERR_ABORTED"] !== undefined && object["MEDIA_ERR_ABORTED"] !== null) {
		value.MEDIA_ERR_ABORTED = object["MEDIA_ERR_ABORTED"];
	}
	if (object["MEDIA_ERR_NETWORK"] !== undefined && object["MEDIA_ERR_NETWORK"] !== null) {
		value.MEDIA_ERR_NETWORK = object["MEDIA_ERR_NETWORK"];
	}
	if (object["MEDIA_ERR_DECODE"] !== undefined && object["MEDIA_ERR_DECODE"] !== null) {
		value.MEDIA_ERR_DECODE = object["MEDIA_ERR_DECODE"];
	}
	if (object["MEDIA_ERR_SRC_NOT_SUPPORTED"] !== undefined && object["MEDIA_ERR_SRC_NOT_SUPPORTED"] !== null) {
		value.MEDIA_ERR_SRC_NOT_SUPPORTED = object["MEDIA_ERR_SRC_NOT_SUPPORTED"];
	}

	return value;
}

// the object the library gets for a Screen
function dom_toScreen(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("AvailHeight") || !(value.AvailHeight === 0)) {
		out["availHeight"] = value.AvailHeight;
	}
	if (explicit.has("AvailWidth") || !(value.AvailWidth === 0)) {
		out["availWidth"] = value.AvailWidth;
	}
	if (explicit.has("ColorDepth") || !(value.ColorDepth === 0)) {
		out["colorDepth"] = value.ColorDepth;
	}
	if (explicit.has("Height") || !(value.Height === 0)) {
		out["height"] = value.Height;
	}
	if (explicit.has("Orientation") || !(value.Orientation === null || value.Orientation === undefined)) {
		out["orientation"] = dom_unwrapAny(value.Orientation);
	}
	if (explicit.has("PixelDepth") || !(value.PixelDepth === 0)) {
		out["pixelDepth"] = value.PixelDepth;
	}
	if (explicit.has("Width") || !(value.Width === 0)) {
		out["width"] = value.Width;
	}

	return out;
}

// a Screen of what the library gives
function dom_fromScreen(object) {
	const value = new dom.Screen();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["availHeight"] !== undefined && object["availHeight"] !== null) {
		value.AvailHeight = object["availHeight"];
	}
	if (object["availWidth"] !== undefined && object["availWidth"] !== null) {
		value.AvailWidth = object["availWidth"];
	}
	if (object["colorDepth"] !== undefined && object["colorDepth"] !== null) {
		value.ColorDepth = object["colorDepth"];
	}
	if (object["height"] !== undefined && object["height"] !== null) {
		value.Height = object["height"];
	}
	if (object["orientation"] !== undefined && object["orientation"] !== null) {
		value.Orientation = object["orientation"];
	}
	if (object["pixelDepth"] !== undefined && object["pixelDepth"] !== null) {
		value.PixelDepth = object["pixelDepth"];
	}
	if (object["width"] !== undefined && object["width"] !== null) {
		value.Width = object["width"];
	}

	return value;
}

// the object the library gets for a StaticRange
function dom_toStaticRange(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("Collapsed") || !(value.Collapsed === false)) {
		out["collapsed"] = value.Collapsed;
	}
	if (explicit.has("EndContainer") || !(value.EndContainer === null || value.EndContainer === undefined)) {
		out["endContainer"] = dom_unwrap(value.EndContainer);
	}
	if (explicit.has("EndOffset") || !(value.EndOffset === 0)) {
		out["endOffset"] = value.EndOffset;
	}
	if (explicit.has("StartContainer") || !(value.StartContainer === null || value.StartContainer === undefined)) {
		out["startContainer"] = dom_unwrap(value.StartContainer);
	}
	if (explicit.has("StartOffset") || !(value.StartOffset === 0)) {
		out["startOffset"] = value.StartOffset;
	}

	return out;
}

// a StaticRange of what the library gives
function dom_fromStaticRange(object) {
	const value = new dom.StaticRange();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["collapsed"] !== undefined && object["collapsed"] !== null) {
		value.Collapsed = object["collapsed"];
	}
	if (object["endContainer"] !== undefined && object["endContainer"] !== null) {
		value.EndContainer = dom_wrap(dom.Node, object["endContainer"]);
	}
	if (object["endOffset"] !== undefined && object["endOffset"] !== null) {
		value.EndOffset = object["endOffset"];
	}
	if (object["startContainer"] !== undefined && object["startContainer"] !== null) {
		value.StartContainer = dom_wrap(dom.Node, object["startContainer"]);
	}
	if (object["startOffset"] !== undefined && object["startOffset"] !== null) {
		value.StartOffset = object["startOffset"];
	}

	return value;
}

// the object the library gets for a TextMetrics
function dom_toTextMetrics(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("ActualBoundingBoxAscent") || !(value.ActualBoundingBoxAscent === 0)) {
		out["actualBoundingBoxAscent"] = value.ActualBoundingBoxAscent;
	}
	if (explicit.has("ActualBoundingBoxDescent") || !(value.ActualBoundingBoxDescent === 0)) {
		out["actualBoundingBoxDescent"] = value.ActualBoundingBoxDescent;
	}
	if (explicit.has("ActualBoundingBoxLeft") || !(value.ActualBoundingBoxLeft === 0)) {
		out["actualBoundingBoxLeft"] = value.ActualBoundingBoxLeft;
	}
	if (explicit.has("ActualBoundingBoxRight") || !(value.ActualBoundingBoxRight === 0)) {
		out["actualBoundingBoxRight"] = value.ActualBoundingBoxRight;
	}
	if (explicit.has("AlphabeticBaseline") || !(value.AlphabeticBaseline === 0)) {
		out["alphabeticBaseline"] = value.AlphabeticBaseline;
	}
	if (explicit.has("EmHeightAscent") || !(value.EmHeightAscent === 0)) {
		out["emHeightAscent"] = value.EmHeightAscent;
	}
	if (explicit.has("EmHeightDescent") || !(value.EmHeightDescent === 0)) {
		out["emHeightDescent"] = value.EmHeightDescent;
	}
	if (explicit.has("FontBoundingBoxAscent") || !(value.FontBoundingBoxAscent === 0)) {
		out["fontBoundingBoxAscent"] = value.FontBoundingBoxAscent;
	}
	if (explicit.has("FontBoundingBoxDescent") || !(value.FontBoundingBoxDescent === 0)) {
		out["fontBoundingBoxDescent"] = value.FontBoundingBoxDescent;
	}
	if (explicit.has("HangingBaseline") || !(value.HangingBaseline === 0)) {
		out["hangingBaseline"] = value.HangingBaseline;
	}
	if (explicit.has("IdeographicBaseline") || !(value.IdeographicBaseline === 0)) {
		out["ideographicBaseline"] = value.IdeographicBaseline;
	}
	if (explicit.has("Width") || !(value.Width === 0)) {
		out["width"] = value.Width;
	}

	return out;
}

// a TextMetrics of what the library gives
function dom_fromTextMetrics(object) {
	const value = new dom.TextMetrics();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["actualBoundingBoxAscent"] !== undefined && object["actualBoundingBoxAscent"] !== null) {
		value.ActualBoundingBoxAscent = object["actualBoundingBoxAscent"];
	}
	if (object["actualBoundingBoxDescent"] !== undefined && object["actualBoundingBoxDescent"] !== null) {
		value.ActualBoundingBoxDescent = object["actualBoundingBoxDescent"];
	}
	if (object["actualBoundingBoxLeft"] !== undefined && object["actualBoundingBoxLeft"] !== null) {
		value.ActualBoundingBoxLeft = object["actualBoundingBoxLeft"];
	}
	if (object["actualBoundingBoxRight"] !== undefined && object["actualBoundingBoxRight"] !== null) {
		value.ActualBoundingBoxRight = object["actualBoundingBoxRight"];
	}
	if (object["alphabeticBaseline"] !== undefined && object["alphabeticBaseline"] !== null) {
		value.AlphabeticBaseline = object["alphabeticBaseline"];
	}
	if (object["emHeightAscent"] !== undefined && object["emHeightAscent"] !== null) {
		value.EmHeightAscent = object["emHeightAscent"];
	}
	if (object["emHeightDescent"] !== undefined && object["emHeightDescent"] !== null) {
		value.EmHeightDescent = object["emHeightDescent"];
	}
	if (object["fontBoundingBoxAscent"] !== undefined && object["fontBoundingBoxAscent"] !== null) {
		value.FontBoundingBoxAscent = object["fontBoundingBoxAscent"];
	}
	if (object["fontBoundingBoxDescent"] !== undefined && object["fontBoundingBoxDescent"] !== null) {
		value.FontBoundingBoxDescent = object["fontBoundingBoxDescent"];
	}
	if (object["hangingBaseline"] !== undefined && object["hangingBaseline"] !== null) {
		value.HangingBaseline = object["hangingBaseline"];
	}
	if (object["ideographicBaseline"] !== undefined && object["ideographicBaseline"] !== null) {
		value.IdeographicBaseline = object["ideographicBaseline"];
	}
	if (object["width"] !== undefined && object["width"] !== null) {
		value.Width = object["width"];
	}

	return value;
}

// the object the library gets for a Touch
function dom_toTouch(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("ClientX") || !(value.ClientX === 0)) {
		out["clientX"] = value.ClientX;
	}
	if (explicit.has("ClientY") || !(value.ClientY === 0)) {
		out["clientY"] = value.ClientY;
	}
	if (explicit.has("Force") || !(value.Force === 0)) {
		out["force"] = value.Force;
	}
	if (explicit.has("Identifier") || !(value.Identifier === 0)) {
		out["identifier"] = value.Identifier;
	}
	if (explicit.has("PageX") || !(value.PageX === 0)) {
		out["pageX"] = value.PageX;
	}
	if (explicit.has("PageY") || !(value.PageY === 0)) {
		out["pageY"] = value.PageY;
	}
	if (explicit.has("RadiusX") || !(value.RadiusX === 0)) {
		out["radiusX"] = value.RadiusX;
	}
	if (explicit.has("RadiusY") || !(value.RadiusY === 0)) {
		out["radiusY"] = value.RadiusY;
	}
	if (explicit.has("RotationAngle") || !(value.RotationAngle === 0)) {
		out["rotationAngle"] = value.RotationAngle;
	}
	if (explicit.has("ScreenX") || !(value.ScreenX === 0)) {
		out["screenX"] = value.ScreenX;
	}
	if (explicit.has("ScreenY") || !(value.ScreenY === 0)) {
		out["screenY"] = value.ScreenY;
	}
	if (explicit.has("Target") || !(value.Target === null || value.Target === undefined)) {
		out["target"] = dom_unwrap(value.Target);
	}

	return out;
}

// a Touch of what the library gives
function dom_fromTouch(object) {
	const value = new dom.Touch();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["clientX"] !== undefined && object["clientX"] !== null) {
		value.ClientX = object["clientX"];
	}
	if (object["clientY"] !== undefined && object["clientY"] !== null) {
		value.ClientY = object["clientY"];
	}
	if (object["force"] !== undefined && object["force"] !== null) {
		value.Force = object["force"];
	}
	if (object["identifier"] !== undefined && object["identifier"] !== null) {
		value.Identifier = object["identifier"];
	}
	if (object["pageX"] !== undefined && object["pageX"] !== null) {
		value.PageX = object["pageX"];
	}
	if (object["pageY"] !== undefined && object["pageY"] !== null) {
		value.PageY = object["pageY"];
	}
	if (object["radiusX"] !== undefined && object["radiusX"] !== null) {
		value.RadiusX = object["radiusX"];
	}
	if (object["radiusY"] !== undefined && object["radiusY"] !== null) {
		value.RadiusY = object["radiusY"];
	}
	if (object["rotationAngle"] !== undefined && object["rotationAngle"] !== null) {
		value.RotationAngle = object["rotationAngle"];
	}
	if (object["screenX"] !== undefined && object["screenX"] !== null) {
		value.ScreenX = object["screenX"];
	}
	if (object["screenY"] !== undefined && object["screenY"] !== null) {
		value.ScreenY = object["screenY"];
	}
	if (object["target"] !== undefined && object["target"] !== null) {
		value.Target = dom_wrap(dom.EventTarget, object["target"]);
	}

	return value;
}

// the object the library gets for a UserActivation
function dom_toUserActivation(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("HasBeenActive") || !(value.HasBeenActive === false)) {
		out["hasBeenActive"] = value.HasBeenActive;
	}
	if (explicit.has("IsActive") || !(value.IsActive === false)) {
		out["isActive"] = value.IsActive;
	}

	return out;
}

// a UserActivation of what the library gives
function dom_fromUserActivation(object) {
	const value = new dom.UserActivation();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["hasBeenActive"] !== undefined && object["hasBeenActive"] !== null) {
		value.HasBeenActive = object["hasBeenActive"];
	}
	if (object["isActive"] !== undefined && object["isActive"] !== null) {
		value.IsActive = object["isActive"];
	}

	return value;
}

// the object the library gets for a ValidityState
function dom_toValidityState(value) {
	const out = {};

	if (value === null || value === undefined) {
		return out;
	}

	const explicit = new Set(value.Explicit || []);

	if (explicit.has("BadInput") || !(value.BadInput === false)) {
		out["badInput"] = value.BadInput;
	}
	if (explicit.has("CustomError") || !(value.CustomError === false)) {
		out["customError"] = value.CustomError;
	}
	if (explicit.has("PatternMismatch") || !(value.PatternMismatch === false)) {
		out["patternMismatch"] = value.PatternMismatch;
	}
	if (explicit.has("RangeOverflow") || !(value.RangeOverflow === false)) {
		out["rangeOverflow"] = value.RangeOverflow;
	}
	if (explicit.has("RangeUnderflow") || !(value.RangeUnderflow === false)) {
		out["rangeUnderflow"] = value.RangeUnderflow;
	}
	if (explicit.has("StepMismatch") || !(value.StepMismatch === false)) {
		out["stepMismatch"] = value.StepMismatch;
	}
	if (explicit.has("TooLong") || !(value.TooLong === false)) {
		out["tooLong"] = value.TooLong;
	}
	if (explicit.has("TooShort") || !(value.TooShort === false)) {
		out["tooShort"] = value.TooShort;
	}
	if (explicit.has("TypeMismatch") || !(value.TypeMismatch === false)) {
		out["typeMismatch"] = value.TypeMismatch;
	}
	if (explicit.has("Valid") || !(value.Valid === false)) {
		out["valid"] = value.Valid;
	}
	if (explicit.has("ValueMissing") || !(value.ValueMissing === false)) {
		out["valueMissing"] = value.ValueMissing;
	}

	return out;
}

// a ValidityState of what the library gives
function dom_fromValidityState(object) {
	const value = new dom.ValidityState();

	if (object === null || typeof object !== "object") {
		return value;
	}

	if (object["badInput"] !== undefined && object["badInput"] !== null) {
		value.BadInput = object["badInput"];
	}
	if (object["customError"] !== undefined && object["customError"] !== null) {
		value.CustomError = object["customError"];
	}
	if (object["patternMismatch"] !== undefined && object["patternMismatch"] !== null) {
		value.PatternMismatch = object["patternMismatch"];
	}
	if (object["rangeOverflow"] !== undefined && object["rangeOverflow"] !== null) {
		value.RangeOverflow = object["rangeOverflow"];
	}
	if (object["rangeUnderflow"] !== undefined && object["rangeUnderflow"] !== null) {
		value.RangeUnderflow = object["rangeUnderflow"];
	}
	if (object["stepMismatch"] !== undefined && object["stepMismatch"] !== null) {
		value.StepMismatch = object["stepMismatch"];
	}
	if (object["tooLong"] !== undefined && object["tooLong"] !== null) {
		value.TooLong = object["tooLong"];
	}
	if (object["tooShort"] !== undefined && object["tooShort"] !== null) {
		value.TooShort = object["tooShort"];
	}
	if (object["typeMismatch"] !== undefined && object["typeMismatch"] !== null) {
		value.TypeMismatch = object["typeMismatch"];
	}
	if (object["valid"] !== undefined && object["valid"] !== null) {
		value.Valid = object["valid"];
	}
	if (object["valueMissing"] !== undefined && object["valueMissing"] !== null) {
		value.ValueMissing = object["valueMissing"];
	}

	return value;
}

class dom_base {
	AsAbortSignal() {
		return dom_is(this, "AbortSignal") ? this : null;
	}
	AsAnalyserNode() {
		return dom_is(this, "AnalyserNode") ? this : null;
	}
	AsAnimation() {
		return dom_is(this, "Animation") ? this : null;
	}
	AsAnimationEvent() {
		return dom_is(this, "AnimationEvent") ? this : null;
	}
	AsAnimationPlaybackEvent() {
		return dom_is(this, "AnimationPlaybackEvent") ? this : null;
	}
	AsAttr() {
		return dom_is(this, "Attr") ? this : null;
	}
	AsAudioBufferSourceNode() {
		return dom_is(this, "AudioBufferSourceNode") ? this : null;
	}
	AsAudioContext() {
		return dom_is(this, "AudioContext") ? this : null;
	}
	AsAudioDestinationNode() {
		return dom_is(this, "AudioDestinationNode") ? this : null;
	}
	AsAudioNode() {
		return dom_is(this, "AudioNode") ? this : null;
	}
	AsAudioProcessingEvent() {
		return dom_is(this, "AudioProcessingEvent") ? this : null;
	}
	AsAudioScheduledSourceNode() {
		return dom_is(this, "AudioScheduledSourceNode") ? this : null;
	}
	AsAudioWorklet() {
		return dom_is(this, "AudioWorklet") ? this : null;
	}
	AsBaseAudioContext() {
		return dom_is(this, "BaseAudioContext") ? this : null;
	}
	AsBeforeUnloadEvent() {
		return dom_is(this, "BeforeUnloadEvent") ? this : null;
	}
	AsBiquadFilterNode() {
		return dom_is(this, "BiquadFilterNode") ? this : null;
	}
	AsCDATASection() {
		return dom_is(this, "CDATASection") ? this : null;
	}
	AsCSSStyleSheet() {
		return dom_is(this, "CSSStyleSheet") ? this : null;
	}
	AsCanvasRenderingContext2D() {
		return dom_is(this, "CanvasRenderingContext2D") ? this : null;
	}
	AsChannelMergerNode() {
		return dom_is(this, "ChannelMergerNode") ? this : null;
	}
	AsChannelSplitterNode() {
		return dom_is(this, "ChannelSplitterNode") ? this : null;
	}
	AsCharacterData() {
		return dom_is(this, "CharacterData") ? this : null;
	}
	AsChildNode() {
		return dom_is(this, "ChildNode") ? this : null;
	}
	AsClipboard() {
		return dom_is(this, "Clipboard") ? this : null;
	}
	AsClipboardEvent() {
		return dom_is(this, "ClipboardEvent") ? this : null;
	}
	AsComment() {
		return dom_is(this, "Comment") ? this : null;
	}
	AsCompositionEvent() {
		return dom_is(this, "CompositionEvent") ? this : null;
	}
	AsConstantSourceNode() {
		return dom_is(this, "ConstantSourceNode") ? this : null;
	}
	AsConvolverNode() {
		return dom_is(this, "ConvolverNode") ? this : null;
	}
	AsDOMMatrix() {
		return dom_is(this, "DOMMatrix") ? this : null;
	}
	AsDOMPoint() {
		return dom_is(this, "DOMPoint") ? this : null;
	}
	AsDOMRect() {
		return dom_is(this, "DOMRect") ? this : null;
	}
	AsDelayNode() {
		return dom_is(this, "DelayNode") ? this : null;
	}
	AsDeviceMotionEvent() {
		return dom_is(this, "DeviceMotionEvent") ? this : null;
	}
	AsDeviceOrientationEvent() {
		return dom_is(this, "DeviceOrientationEvent") ? this : null;
	}
	AsDocument() {
		return dom_is(this, "Document") ? this : null;
	}
	AsDocumentFragment() {
		return dom_is(this, "DocumentFragment") ? this : null;
	}
	AsDocumentType() {
		return dom_is(this, "DocumentType") ? this : null;
	}
	AsDragEvent() {
		return dom_is(this, "DragEvent") ? this : null;
	}
	AsDynamicsCompressorNode() {
		return dom_is(this, "DynamicsCompressorNode") ? this : null;
	}
	AsElement() {
		return dom_is(this, "Element") ? this : null;
	}
	AsElementInternals() {
		return dom_is(this, "ElementInternals") ? this : null;
	}
	AsErrorEvent() {
		return dom_is(this, "ErrorEvent") ? this : null;
	}
	AsFocusEvent() {
		return dom_is(this, "FocusEvent") ? this : null;
	}
	AsFontFaceSet() {
		return dom_is(this, "FontFaceSet") ? this : null;
	}
	AsFormDataEvent() {
		return dom_is(this, "FormDataEvent") ? this : null;
	}
	AsGainNode() {
		return dom_is(this, "GainNode") ? this : null;
	}
	AsGamepadEvent() {
		return dom_is(this, "GamepadEvent") ? this : null;
	}
	AsHTMLAudioElement() {
		return dom_is(this, "HTMLAudioElement") ? this : null;
	}
	AsHTMLCanvasElement() {
		return dom_is(this, "HTMLCanvasElement") ? this : null;
	}
	AsHTMLCollection() {
		return dom_is(this, "HTMLCollection") ? this : null;
	}
	AsHTMLCollectionOf() {
		return dom_is(this, "HTMLCollectionOf") ? this : null;
	}
	AsHTMLElement() {
		return dom_is(this, "HTMLElement") ? this : null;
	}
	AsHTMLFormElement() {
		return dom_is(this, "HTMLFormElement") ? this : null;
	}
	AsHTMLHeadElement() {
		return dom_is(this, "HTMLHeadElement") ? this : null;
	}
	AsHTMLImageElement() {
		return dom_is(this, "HTMLImageElement") ? this : null;
	}
	AsHTMLMediaElement() {
		return dom_is(this, "HTMLMediaElement") ? this : null;
	}
	AsHTMLSlotElement() {
		return dom_is(this, "HTMLSlotElement") ? this : null;
	}
	AsHashChangeEvent() {
		return dom_is(this, "HashChangeEvent") ? this : null;
	}
	AsIIRFilterNode() {
		return dom_is(this, "IIRFilterNode") ? this : null;
	}
	AsInputEvent() {
		return dom_is(this, "InputEvent") ? this : null;
	}
	AsKeyboardEvent() {
		return dom_is(this, "KeyboardEvent") ? this : null;
	}
	AsMIDIAccess() {
		return dom_is(this, "MIDIAccess") ? this : null;
	}
	AsMediaDevices() {
		return dom_is(this, "MediaDevices") ? this : null;
	}
	AsMediaElementAudioSourceNode() {
		return dom_is(this, "MediaElementAudioSourceNode") ? this : null;
	}
	AsMediaEncryptedEvent() {
		return dom_is(this, "MediaEncryptedEvent") ? this : null;
	}
	AsMediaKeySession() {
		return dom_is(this, "MediaKeySession") ? this : null;
	}
	AsMediaQueryList() {
		return dom_is(this, "MediaQueryList") ? this : null;
	}
	AsMediaStream() {
		return dom_is(this, "MediaStream") ? this : null;
	}
	AsMediaStreamAudioDestinationNode() {
		return dom_is(this, "MediaStreamAudioDestinationNode") ? this : null;
	}
	AsMediaStreamAudioSourceNode() {
		return dom_is(this, "MediaStreamAudioSourceNode") ? this : null;
	}
	AsMediaStreamTrack() {
		return dom_is(this, "MediaStreamTrack") ? this : null;
	}
	AsMediaStreamTrackEvent() {
		return dom_is(this, "MediaStreamTrackEvent") ? this : null;
	}
	AsMessageEvent() {
		return dom_is(this, "MessageEvent") ? this : null;
	}
	AsMouseEvent() {
		return dom_is(this, "MouseEvent") ? this : null;
	}
	AsNavigator() {
		return dom_is(this, "Navigator") ? this : null;
	}
	AsNode() {
		return dom_is(this, "Node") ? this : null;
	}
	AsNodeListOf() {
		return dom_is(this, "NodeListOf") ? this : null;
	}
	AsOffscreenCanvas() {
		return dom_is(this, "OffscreenCanvas") ? this : null;
	}
	AsOffscreenCanvasRenderingContext2D() {
		return dom_is(this, "OffscreenCanvasRenderingContext2D") ? this : null;
	}
	AsOscillatorNode() {
		return dom_is(this, "OscillatorNode") ? this : null;
	}
	AsPageTransitionEvent() {
		return dom_is(this, "PageTransitionEvent") ? this : null;
	}
	AsPannerNode() {
		return dom_is(this, "PannerNode") ? this : null;
	}
	AsParentNode() {
		return dom_is(this, "ParentNode") ? this : null;
	}
	AsPath2D() {
		return dom_is(this, "Path2D") ? this : null;
	}
	AsPerformance() {
		return dom_is(this, "Performance") ? this : null;
	}
	AsPerformanceMark() {
		return dom_is(this, "PerformanceMark") ? this : null;
	}
	AsPerformanceMeasure() {
		return dom_is(this, "PerformanceMeasure") ? this : null;
	}
	AsPointerEvent() {
		return dom_is(this, "PointerEvent") ? this : null;
	}
	AsPopStateEvent() {
		return dom_is(this, "PopStateEvent") ? this : null;
	}
	AsProcessingInstruction() {
		return dom_is(this, "ProcessingInstruction") ? this : null;
	}
	AsProgressEvent() {
		return dom_is(this, "ProgressEvent") ? this : null;
	}
	AsPromiseRejectionEvent() {
		return dom_is(this, "PromiseRejectionEvent") ? this : null;
	}
	AsRange() {
		return dom_is(this, "Range") ? this : null;
	}
	AsRemotePlayback() {
		return dom_is(this, "RemotePlayback") ? this : null;
	}
	AsResponse() {
		return dom_is(this, "Response") ? this : null;
	}
	AsSVGElement() {
		return dom_is(this, "SVGElement") ? this : null;
	}
	AsSVGGraphicsElement() {
		return dom_is(this, "SVGGraphicsElement") ? this : null;
	}
	AsSVGSVGElement() {
		return dom_is(this, "SVGSVGElement") ? this : null;
	}
	AsScriptProcessorNode() {
		return dom_is(this, "ScriptProcessorNode") ? this : null;
	}
	AsSecurityPolicyViolationEvent() {
		return dom_is(this, "SecurityPolicyViolationEvent") ? this : null;
	}
	AsServiceWorkerContainer() {
		return dom_is(this, "ServiceWorkerContainer") ? this : null;
	}
	AsShadowRoot() {
		return dom_is(this, "ShadowRoot") ? this : null;
	}
	AsSpeechSynthesis() {
		return dom_is(this, "SpeechSynthesis") ? this : null;
	}
	AsStereoPannerNode() {
		return dom_is(this, "StereoPannerNode") ? this : null;
	}
	AsStorageEvent() {
		return dom_is(this, "StorageEvent") ? this : null;
	}
	AsStylePropertyMap() {
		return dom_is(this, "StylePropertyMap") ? this : null;
	}
	AsSubmitEvent() {
		return dom_is(this, "SubmitEvent") ? this : null;
	}
	AsText() {
		return dom_is(this, "Text") ? this : null;
	}
	AsTextTrack() {
		return dom_is(this, "TextTrack") ? this : null;
	}
	AsTextTrackCue() {
		return dom_is(this, "TextTrackCue") ? this : null;
	}
	AsTextTrackList() {
		return dom_is(this, "TextTrackList") ? this : null;
	}
	AsTouchEvent() {
		return dom_is(this, "TouchEvent") ? this : null;
	}
	AsTrackEvent() {
		return dom_is(this, "TrackEvent") ? this : null;
	}
	AsTransitionEvent() {
		return dom_is(this, "TransitionEvent") ? this : null;
	}
	AsUIEvent() {
		return dom_is(this, "UIEvent") ? this : null;
	}
	AsVisualViewport() {
		return dom_is(this, "VisualViewport") ? this : null;
	}
	AsWaveShaperNode() {
		return dom_is(this, "WaveShaperNode") ? this : null;
	}
	AsWheelEvent() {
		return dom_is(this, "WheelEvent") ? this : null;
	}
	AsWindow() {
		return dom_is(this, "Window") ? this : null;
	}
}

class dom_C_ARIAMixin extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfARIAMixin() {
			
		}

		AsElement() {
			return dom_is(this, "Element") ? this : null;
		}

		AsElementInternals() {
			return dom_is(this, "ElementInternals") ? this : null;
		}

		AsHTMLAudioElement() {
			return dom_is(this, "HTMLAudioElement") ? this : null;
		}

		AsHTMLCanvasElement() {
			return dom_is(this, "HTMLCanvasElement") ? this : null;
		}

		AsHTMLElement() {
			return dom_is(this, "HTMLElement") ? this : null;
		}

		AsHTMLFormElement() {
			return dom_is(this, "HTMLFormElement") ? this : null;
		}

		AsHTMLHeadElement() {
			return dom_is(this, "HTMLHeadElement") ? this : null;
		}

		AsHTMLImageElement() {
			return dom_is(this, "HTMLImageElement") ? this : null;
		}

		AsHTMLMediaElement() {
			return dom_is(this, "HTMLMediaElement") ? this : null;
		}

		AsHTMLSlotElement() {
			return dom_is(this, "HTMLSlotElement") ? this : null;
		}

		AsSVGElement() {
			return dom_is(this, "SVGElement") ? this : null;
		}

		AsSVGGraphicsElement() {
			return dom_is(this, "SVGGraphicsElement") ? this : null;
		}

		AsSVGSVGElement() {
			return dom_is(this, "SVGSVGElement") ? this : null;
		}

		AriaAtomic() {
			return this.handle.ariaAtomic;
		}

		SetAriaAtomic(value) {
			this.handle.ariaAtomic = value;
		}

		AriaAutoComplete() {
			return this.handle.ariaAutoComplete;
		}

		SetAriaAutoComplete(value) {
			this.handle.ariaAutoComplete = value;
		}

		AriaBrailleLabel() {
			return this.handle.ariaBrailleLabel;
		}

		SetAriaBrailleLabel(value) {
			this.handle.ariaBrailleLabel = value;
		}

		AriaBrailleRoleDescription() {
			return this.handle.ariaBrailleRoleDescription;
		}

		SetAriaBrailleRoleDescription(value) {
			this.handle.ariaBrailleRoleDescription = value;
		}

		AriaBusy() {
			return this.handle.ariaBusy;
		}

		SetAriaBusy(value) {
			this.handle.ariaBusy = value;
		}

		AriaChecked() {
			return this.handle.ariaChecked;
		}

		SetAriaChecked(value) {
			this.handle.ariaChecked = value;
		}

		AriaColCount() {
			return this.handle.ariaColCount;
		}

		SetAriaColCount(value) {
			this.handle.ariaColCount = value;
		}

		AriaColIndex() {
			return this.handle.ariaColIndex;
		}

		SetAriaColIndex(value) {
			this.handle.ariaColIndex = value;
		}

		AriaColSpan() {
			return this.handle.ariaColSpan;
		}

		SetAriaColSpan(value) {
			this.handle.ariaColSpan = value;
		}

		AriaCurrent() {
			return this.handle.ariaCurrent;
		}

		SetAriaCurrent(value) {
			this.handle.ariaCurrent = value;
		}

		AriaDescription() {
			return this.handle.ariaDescription;
		}

		SetAriaDescription(value) {
			this.handle.ariaDescription = value;
		}

		AriaDisabled() {
			return this.handle.ariaDisabled;
		}

		SetAriaDisabled(value) {
			this.handle.ariaDisabled = value;
		}

		AriaExpanded() {
			return this.handle.ariaExpanded;
		}

		SetAriaExpanded(value) {
			this.handle.ariaExpanded = value;
		}

		AriaHasPopup() {
			return this.handle.ariaHasPopup;
		}

		SetAriaHasPopup(value) {
			this.handle.ariaHasPopup = value;
		}

		AriaHidden() {
			return this.handle.ariaHidden;
		}

		SetAriaHidden(value) {
			this.handle.ariaHidden = value;
		}

		AriaInvalid() {
			return this.handle.ariaInvalid;
		}

		SetAriaInvalid(value) {
			this.handle.ariaInvalid = value;
		}

		AriaKeyShortcuts() {
			return this.handle.ariaKeyShortcuts;
		}

		SetAriaKeyShortcuts(value) {
			this.handle.ariaKeyShortcuts = value;
		}

		AriaLabel() {
			return this.handle.ariaLabel;
		}

		SetAriaLabel(value) {
			this.handle.ariaLabel = value;
		}

		AriaLevel() {
			return this.handle.ariaLevel;
		}

		SetAriaLevel(value) {
			this.handle.ariaLevel = value;
		}

		AriaLive() {
			return this.handle.ariaLive;
		}

		SetAriaLive(value) {
			this.handle.ariaLive = value;
		}

		AriaModal() {
			return this.handle.ariaModal;
		}

		SetAriaModal(value) {
			this.handle.ariaModal = value;
		}

		AriaMultiLine() {
			return this.handle.ariaMultiLine;
		}

		SetAriaMultiLine(value) {
			this.handle.ariaMultiLine = value;
		}

		AriaMultiSelectable() {
			return this.handle.ariaMultiSelectable;
		}

		SetAriaMultiSelectable(value) {
			this.handle.ariaMultiSelectable = value;
		}

		AriaOrientation() {
			return this.handle.ariaOrientation;
		}

		SetAriaOrientation(value) {
			this.handle.ariaOrientation = value;
		}

		AriaPlaceholder() {
			return this.handle.ariaPlaceholder;
		}

		SetAriaPlaceholder(value) {
			this.handle.ariaPlaceholder = value;
		}

		AriaPosInSet() {
			return this.handle.ariaPosInSet;
		}

		SetAriaPosInSet(value) {
			this.handle.ariaPosInSet = value;
		}

		AriaPressed() {
			return this.handle.ariaPressed;
		}

		SetAriaPressed(value) {
			this.handle.ariaPressed = value;
		}

		AriaReadOnly() {
			return this.handle.ariaReadOnly;
		}

		SetAriaReadOnly(value) {
			this.handle.ariaReadOnly = value;
		}

		AriaRequired() {
			return this.handle.ariaRequired;
		}

		SetAriaRequired(value) {
			this.handle.ariaRequired = value;
		}

		AriaRoleDescription() {
			return this.handle.ariaRoleDescription;
		}

		SetAriaRoleDescription(value) {
			this.handle.ariaRoleDescription = value;
		}

		AriaRowCount() {
			return this.handle.ariaRowCount;
		}

		SetAriaRowCount(value) {
			this.handle.ariaRowCount = value;
		}

		AriaRowIndex() {
			return this.handle.ariaRowIndex;
		}

		SetAriaRowIndex(value) {
			this.handle.ariaRowIndex = value;
		}

		AriaRowSpan() {
			return this.handle.ariaRowSpan;
		}

		SetAriaRowSpan(value) {
			this.handle.ariaRowSpan = value;
		}

		AriaSelected() {
			return this.handle.ariaSelected;
		}

		SetAriaSelected(value) {
			this.handle.ariaSelected = value;
		}

		AriaSetSize() {
			return this.handle.ariaSetSize;
		}

		SetAriaSetSize(value) {
			this.handle.ariaSetSize = value;
		}

		AriaSort() {
			return this.handle.ariaSort;
		}

		SetAriaSort(value) {
			this.handle.ariaSort = value;
		}

		AriaValueMax() {
			return this.handle.ariaValueMax;
		}

		SetAriaValueMax(value) {
			this.handle.ariaValueMax = value;
		}

		AriaValueMin() {
			return this.handle.ariaValueMin;
		}

		SetAriaValueMin(value) {
			this.handle.ariaValueMin = value;
		}

		AriaValueNow() {
			return this.handle.ariaValueNow;
		}

		SetAriaValueNow(value) {
			this.handle.ariaValueNow = value;
		}

		AriaValueText() {
			return this.handle.ariaValueText;
		}

		SetAriaValueText(value) {
			this.handle.ariaValueText = value;
		}

		Role() {
			return this.handle.role;
		}

		SetRole(value) {
			this.handle.role = value;
		}
}

class dom_C_EventTarget extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfEventTarget() {
			
		}

		AsAbortSignal() {
			return dom_is(this, "AbortSignal") ? this : null;
		}

		AsAnalyserNode() {
			return dom_is(this, "AnalyserNode") ? this : null;
		}

		AsAnimation() {
			return dom_is(this, "Animation") ? this : null;
		}

		AsAttr() {
			return dom_is(this, "Attr") ? this : null;
		}

		AsAudioBufferSourceNode() {
			return dom_is(this, "AudioBufferSourceNode") ? this : null;
		}

		AsAudioContext() {
			return dom_is(this, "AudioContext") ? this : null;
		}

		AsAudioDestinationNode() {
			return dom_is(this, "AudioDestinationNode") ? this : null;
		}

		AsAudioNode() {
			return dom_is(this, "AudioNode") ? this : null;
		}

		AsAudioScheduledSourceNode() {
			return dom_is(this, "AudioScheduledSourceNode") ? this : null;
		}

		AsBaseAudioContext() {
			return dom_is(this, "BaseAudioContext") ? this : null;
		}

		AsBiquadFilterNode() {
			return dom_is(this, "BiquadFilterNode") ? this : null;
		}

		AsCDATASection() {
			return dom_is(this, "CDATASection") ? this : null;
		}

		AsChannelMergerNode() {
			return dom_is(this, "ChannelMergerNode") ? this : null;
		}

		AsChannelSplitterNode() {
			return dom_is(this, "ChannelSplitterNode") ? this : null;
		}

		AsCharacterData() {
			return dom_is(this, "CharacterData") ? this : null;
		}

		AsChildNode() {
			return dom_is(this, "ChildNode") ? this : null;
		}

		AsClipboard() {
			return dom_is(this, "Clipboard") ? this : null;
		}

		AsComment() {
			return dom_is(this, "Comment") ? this : null;
		}

		AsConstantSourceNode() {
			return dom_is(this, "ConstantSourceNode") ? this : null;
		}

		AsConvolverNode() {
			return dom_is(this, "ConvolverNode") ? this : null;
		}

		AsDelayNode() {
			return dom_is(this, "DelayNode") ? this : null;
		}

		AsDocument() {
			return dom_is(this, "Document") ? this : null;
		}

		AsDocumentType() {
			return dom_is(this, "DocumentType") ? this : null;
		}

		AsDocumentFragment() {
			return dom_is(this, "DocumentFragment") ? this : null;
		}

		AsDynamicsCompressorNode() {
			return dom_is(this, "DynamicsCompressorNode") ? this : null;
		}

		AsElement() {
			return dom_is(this, "Element") ? this : null;
		}

		AsFontFaceSet() {
			return dom_is(this, "FontFaceSet") ? this : null;
		}

		AsGainNode() {
			return dom_is(this, "GainNode") ? this : null;
		}

		AsHTMLAudioElement() {
			return dom_is(this, "HTMLAudioElement") ? this : null;
		}

		AsHTMLCanvasElement() {
			return dom_is(this, "HTMLCanvasElement") ? this : null;
		}

		AsHTMLElement() {
			return dom_is(this, "HTMLElement") ? this : null;
		}

		AsHTMLFormElement() {
			return dom_is(this, "HTMLFormElement") ? this : null;
		}

		AsHTMLHeadElement() {
			return dom_is(this, "HTMLHeadElement") ? this : null;
		}

		AsHTMLImageElement() {
			return dom_is(this, "HTMLImageElement") ? this : null;
		}

		AsHTMLMediaElement() {
			return dom_is(this, "HTMLMediaElement") ? this : null;
		}

		AsHTMLSlotElement() {
			return dom_is(this, "HTMLSlotElement") ? this : null;
		}

		AsIIRFilterNode() {
			return dom_is(this, "IIRFilterNode") ? this : null;
		}

		AsMIDIAccess() {
			return dom_is(this, "MIDIAccess") ? this : null;
		}

		AsMediaDevices() {
			return dom_is(this, "MediaDevices") ? this : null;
		}

		AsMediaElementAudioSourceNode() {
			return dom_is(this, "MediaElementAudioSourceNode") ? this : null;
		}

		AsMediaKeySession() {
			return dom_is(this, "MediaKeySession") ? this : null;
		}

		AsMediaQueryList() {
			return dom_is(this, "MediaQueryList") ? this : null;
		}

		AsMediaStream() {
			return dom_is(this, "MediaStream") ? this : null;
		}

		AsMediaStreamAudioDestinationNode() {
			return dom_is(this, "MediaStreamAudioDestinationNode") ? this : null;
		}

		AsMediaStreamAudioSourceNode() {
			return dom_is(this, "MediaStreamAudioSourceNode") ? this : null;
		}

		AsMediaStreamTrack() {
			return dom_is(this, "MediaStreamTrack") ? this : null;
		}

		AsNode() {
			return dom_is(this, "Node") ? this : null;
		}

		AsOffscreenCanvas() {
			return dom_is(this, "OffscreenCanvas") ? this : null;
		}

		AsOscillatorNode() {
			return dom_is(this, "OscillatorNode") ? this : null;
		}

		AsPannerNode() {
			return dom_is(this, "PannerNode") ? this : null;
		}

		AsParentNode() {
			return dom_is(this, "ParentNode") ? this : null;
		}

		AsPerformance() {
			return dom_is(this, "Performance") ? this : null;
		}

		AsProcessingInstruction() {
			return dom_is(this, "ProcessingInstruction") ? this : null;
		}

		AsRemotePlayback() {
			return dom_is(this, "RemotePlayback") ? this : null;
		}

		AsSVGElement() {
			return dom_is(this, "SVGElement") ? this : null;
		}

		AsSVGGraphicsElement() {
			return dom_is(this, "SVGGraphicsElement") ? this : null;
		}

		AsSVGSVGElement() {
			return dom_is(this, "SVGSVGElement") ? this : null;
		}

		AsScriptProcessorNode() {
			return dom_is(this, "ScriptProcessorNode") ? this : null;
		}

		AsServiceWorkerContainer() {
			return dom_is(this, "ServiceWorkerContainer") ? this : null;
		}

		AsShadowRoot() {
			return dom_is(this, "ShadowRoot") ? this : null;
		}

		AsSpeechSynthesis() {
			return dom_is(this, "SpeechSynthesis") ? this : null;
		}

		AsStereoPannerNode() {
			return dom_is(this, "StereoPannerNode") ? this : null;
		}

		AsText() {
			return dom_is(this, "Text") ? this : null;
		}

		AsTextTrack() {
			return dom_is(this, "TextTrack") ? this : null;
		}

		AsTextTrackCue() {
			return dom_is(this, "TextTrackCue") ? this : null;
		}

		AsTextTrackList() {
			return dom_is(this, "TextTrackList") ? this : null;
		}

		AsVisualViewport() {
			return dom_is(this, "VisualViewport") ? this : null;
		}

		AsWaveShaperNode() {
			return dom_is(this, "WaveShaperNode") ? this : null;
		}

		AsWindow() {
			return dom_is(this, "Window") ? this : null;
		}

		AddEventListener(p0, p1, ...p2) {
			this.handle.addEventListener(p0, (p1 === null || p1 === undefined ? null : (...a) => go2jsCallNow(p1, null, [dom_wrap(dom.Event, a[0])])), ...p2.map((item) => dom_unwrapAny(item)));
		}

		DispatchEvent(p0) {
			return this.handle.dispatchEvent(dom_unwrap(p0));
		}

		RemoveEventListener(p0, p1, ...p2) {
			this.handle.removeEventListener(p0, (p1 === null || p1 === undefined ? null : (...a) => go2jsCallNow(p1, null, [dom_wrap(dom.Event, a[0])])), ...p2.map((item) => dom_unwrapAny(item)));
		}
}

class dom_C_AbortSignal extends dom_C_EventTarget {

		KindOfAbortSignal() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		ThrowIfAborted() {
			this.handle.throwIfAborted();
		}

		OnAbort(p0) {
			this.handle.addEventListener("abort", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		Onabort(p0) {
			return this.handle.onabort(dom_unwrap(p0));
		}

		Aborted() {
			return this.handle.aborted;
		}

		Reason() {
			return this.handle.reason;
		}
}

class dom_C_AbstractRange extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfAbstractRange() {
			
		}

		AsRange() {
			return dom_is(this, "Range") ? this : null;
		}

		Collapsed() {
			return this.handle.collapsed;
		}

		EndContainer() {
			return dom_wrap(dom.Node, this.handle.endContainer);
		}

		EndOffset() {
			return this.handle.endOffset;
		}

		StartContainer() {
			return dom_wrap(dom.Node, this.handle.startContainer);
		}

		StartOffset() {
			return this.handle.startOffset;
		}
}

class dom_C_AudioNode extends dom_C_EventTarget {

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		AsAnalyserNode() {
			return dom_is(this, "AnalyserNode") ? this : null;
		}

		AsAudioBufferSourceNode() {
			return dom_is(this, "AudioBufferSourceNode") ? this : null;
		}

		AsAudioDestinationNode() {
			return dom_is(this, "AudioDestinationNode") ? this : null;
		}

		AsAudioScheduledSourceNode() {
			return dom_is(this, "AudioScheduledSourceNode") ? this : null;
		}

		AsBiquadFilterNode() {
			return dom_is(this, "BiquadFilterNode") ? this : null;
		}

		AsChannelMergerNode() {
			return dom_is(this, "ChannelMergerNode") ? this : null;
		}

		AsChannelSplitterNode() {
			return dom_is(this, "ChannelSplitterNode") ? this : null;
		}

		AsConstantSourceNode() {
			return dom_is(this, "ConstantSourceNode") ? this : null;
		}

		AsConvolverNode() {
			return dom_is(this, "ConvolverNode") ? this : null;
		}

		AsDelayNode() {
			return dom_is(this, "DelayNode") ? this : null;
		}

		AsDynamicsCompressorNode() {
			return dom_is(this, "DynamicsCompressorNode") ? this : null;
		}

		AsGainNode() {
			return dom_is(this, "GainNode") ? this : null;
		}

		AsIIRFilterNode() {
			return dom_is(this, "IIRFilterNode") ? this : null;
		}

		AsMediaElementAudioSourceNode() {
			return dom_is(this, "MediaElementAudioSourceNode") ? this : null;
		}

		AsMediaStreamAudioDestinationNode() {
			return dom_is(this, "MediaStreamAudioDestinationNode") ? this : null;
		}

		AsMediaStreamAudioSourceNode() {
			return dom_is(this, "MediaStreamAudioSourceNode") ? this : null;
		}

		AsOscillatorNode() {
			return dom_is(this, "OscillatorNode") ? this : null;
		}

		AsPannerNode() {
			return dom_is(this, "PannerNode") ? this : null;
		}

		AsScriptProcessorNode() {
			return dom_is(this, "ScriptProcessorNode") ? this : null;
		}

		AsStereoPannerNode() {
			return dom_is(this, "StereoPannerNode") ? this : null;
		}

		AsWaveShaperNode() {
			return dom_is(this, "WaveShaperNode") ? this : null;
		}

		Connect(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			return dom_wrap(dom.AudioNode, this.handle.connect(dom_unwrap(p0), ...p1.map((item) => item)));
		}

		Disconnect() {
			this.handle.disconnect();
		}

		ChannelCount() {
			return this.handle.channelCount;
		}

		SetChannelCount(value) {
			this.handle.channelCount = value;
		}

		ChannelCountMode() {
			return this.handle.channelCountMode;
		}

		SetChannelCountMode(value) {
			this.handle.channelCountMode = value;
		}

		ChannelInterpretation() {
			return this.handle.channelInterpretation;
		}

		SetChannelInterpretation(value) {
			this.handle.channelInterpretation = value;
		}

		Context() {
			return dom_wrap(dom.BaseAudioContext, this.handle.context);
		}

		NumberOfInputs() {
			return this.handle.numberOfInputs;
		}

		NumberOfOutputs() {
			return this.handle.numberOfOutputs;
		}
}

class dom_C_AnalyserNode extends dom_C_AudioNode {

		KindOfAnalyserNode() {
			
		}

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsAudioNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		GetByteFrequencyData(p0) {
			this.handle.getByteFrequencyData(Uint8Array.from(p0 || []));
		}

		GetByteTimeDomainData(p0) {
			this.handle.getByteTimeDomainData(Uint8Array.from(p0 || []));
		}

		GetFloatFrequencyData(p0) {
			this.handle.getFloatFrequencyData(Float32Array.from(p0 || []));
		}

		GetFloatTimeDomainData(p0) {
			this.handle.getFloatTimeDomainData(Float32Array.from(p0 || []));
		}

		FftSize() {
			return this.handle.fftSize;
		}

		SetFftSize(value) {
			this.handle.fftSize = value;
		}

		FrequencyBinCount() {
			return this.handle.frequencyBinCount;
		}

		MaxDecibels() {
			return this.handle.maxDecibels;
		}

		SetMaxDecibels(value) {
			this.handle.maxDecibels = value;
		}

		MinDecibels() {
			return this.handle.minDecibels;
		}

		SetMinDecibels(value) {
			this.handle.minDecibels = value;
		}

		SmoothingTimeConstant() {
			return this.handle.smoothingTimeConstant;
		}

		SetSmoothingTimeConstant(value) {
			this.handle.smoothingTimeConstant = value;
		}
}

class dom_C_Animatable extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfAnimatable() {
			
		}

		AsElement() {
			return dom_is(this, "Element") ? this : null;
		}

		AsHTMLAudioElement() {
			return dom_is(this, "HTMLAudioElement") ? this : null;
		}

		AsHTMLCanvasElement() {
			return dom_is(this, "HTMLCanvasElement") ? this : null;
		}

		AsHTMLElement() {
			return dom_is(this, "HTMLElement") ? this : null;
		}

		AsHTMLFormElement() {
			return dom_is(this, "HTMLFormElement") ? this : null;
		}

		AsHTMLHeadElement() {
			return dom_is(this, "HTMLHeadElement") ? this : null;
		}

		AsHTMLImageElement() {
			return dom_is(this, "HTMLImageElement") ? this : null;
		}

		AsHTMLMediaElement() {
			return dom_is(this, "HTMLMediaElement") ? this : null;
		}

		AsHTMLSlotElement() {
			return dom_is(this, "HTMLSlotElement") ? this : null;
		}

		AsSVGElement() {
			return dom_is(this, "SVGElement") ? this : null;
		}

		AsSVGGraphicsElement() {
			return dom_is(this, "SVGGraphicsElement") ? this : null;
		}

		AsSVGSVGElement() {
			return dom_is(this, "SVGSVGElement") ? this : null;
		}

		Animate(p0, ...p1) {
			return dom_wrap(dom.Animation, this.handle.animate(dom_unwrapAny(p0), ...p1.map((item) => dom_unwrapAny(item))));
		}

		GetAnimations(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return Array.from(this.handle.getAnimations(...p0.map((item) => dom_toGetAnimationsOptions(item))) || [], (item) => dom_wrap(dom.Animation, item));
		}
}

class dom_C_Animation extends dom_C_EventTarget {

		KindOfAnimation() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		Cancel() {
			this.handle.cancel();
		}

		CommitStyles() {
			this.handle.commitStyles();
		}

		Finish() {
			this.handle.finish();
		}

		Pause() {
			this.handle.pause();
		}

		Persist() {
			this.handle.persist();
		}

		Play() {
			this.handle.play();
		}

		Reverse() {
			this.handle.reverse();
		}

		UpdatePlaybackRate(p0) {
			this.handle.updatePlaybackRate(p0);
		}

		OnCancel(p0) {
			this.handle.addEventListener("cancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationPlaybackEvent, a[0])])));
		}

		OnFinish(p0) {
			this.handle.addEventListener("finish", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationPlaybackEvent, a[0])])));
		}

		OnRemove(p0) {
			this.handle.addEventListener("remove", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		Oncancel(p0) {
			return this.handle.oncancel(dom_unwrap(p0));
		}

		Onfinish(p0) {
			return this.handle.onfinish(dom_unwrap(p0));
		}

		Onremove(p0) {
			return this.handle.onremove(dom_unwrap(p0));
		}

		CurrentTime() {
			return this.handle.currentTime;
		}

		SetCurrentTime(value) {
			this.handle.currentTime = dom_unwrapAny(value);
		}

		Effect() {
			return dom_wrap(dom.AnimationEffect, this.handle.effect);
		}

		SetEffect(value) {
			this.handle.effect = dom_unwrap(value);
		}

		Id() {
			return this.handle.id;
		}

		SetId(value) {
			this.handle.id = value;
		}

		Pending() {
			return this.handle.pending;
		}

		PlayState() {
			return this.handle.playState;
		}

		PlaybackRate() {
			return this.handle.playbackRate;
		}

		SetPlaybackRate(value) {
			this.handle.playbackRate = value;
		}

		ReplaceState() {
			return this.handle.replaceState;
		}

		StartTime() {
			return this.handle.startTime;
		}

		SetStartTime(value) {
			this.handle.startTime = dom_unwrapAny(value);
		}

		Timeline() {
			return dom_fromAnimationTimeline(this.handle.timeline);
		}

		SetTimeline(value) {
			this.handle.timeline = dom_toAnimationTimeline(value);
		}
}

class dom_C_AnimationEffect extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfAnimationEffect() {
			
		}

		GetComputedTiming() {
			return this.handle.getComputedTiming();
		}

		GetTiming() {
			return this.handle.getTiming();
		}

		UpdateTiming(...p0) {
			this.handle.updateTiming(...p0.map((item) => dom_unwrapAny(item)));
		}
}

class dom_C_Event extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfEvent() {
			
		}

		AsAnimationEvent() {
			return dom_is(this, "AnimationEvent") ? this : null;
		}

		AsAnimationPlaybackEvent() {
			return dom_is(this, "AnimationPlaybackEvent") ? this : null;
		}

		AsAudioProcessingEvent() {
			return dom_is(this, "AudioProcessingEvent") ? this : null;
		}

		AsBeforeUnloadEvent() {
			return dom_is(this, "BeforeUnloadEvent") ? this : null;
		}

		AsClipboardEvent() {
			return dom_is(this, "ClipboardEvent") ? this : null;
		}

		AsCompositionEvent() {
			return dom_is(this, "CompositionEvent") ? this : null;
		}

		AsDeviceMotionEvent() {
			return dom_is(this, "DeviceMotionEvent") ? this : null;
		}

		AsDeviceOrientationEvent() {
			return dom_is(this, "DeviceOrientationEvent") ? this : null;
		}

		AsDragEvent() {
			return dom_is(this, "DragEvent") ? this : null;
		}

		AsErrorEvent() {
			return dom_is(this, "ErrorEvent") ? this : null;
		}

		AsFocusEvent() {
			return dom_is(this, "FocusEvent") ? this : null;
		}

		AsFormDataEvent() {
			return dom_is(this, "FormDataEvent") ? this : null;
		}

		AsGamepadEvent() {
			return dom_is(this, "GamepadEvent") ? this : null;
		}

		AsHashChangeEvent() {
			return dom_is(this, "HashChangeEvent") ? this : null;
		}

		AsInputEvent() {
			return dom_is(this, "InputEvent") ? this : null;
		}

		AsKeyboardEvent() {
			return dom_is(this, "KeyboardEvent") ? this : null;
		}

		AsMediaEncryptedEvent() {
			return dom_is(this, "MediaEncryptedEvent") ? this : null;
		}

		AsMediaStreamTrackEvent() {
			return dom_is(this, "MediaStreamTrackEvent") ? this : null;
		}

		AsMessageEvent() {
			return dom_is(this, "MessageEvent") ? this : null;
		}

		AsMouseEvent() {
			return dom_is(this, "MouseEvent") ? this : null;
		}

		AsPageTransitionEvent() {
			return dom_is(this, "PageTransitionEvent") ? this : null;
		}

		AsPointerEvent() {
			return dom_is(this, "PointerEvent") ? this : null;
		}

		AsPopStateEvent() {
			return dom_is(this, "PopStateEvent") ? this : null;
		}

		AsProgressEvent() {
			return dom_is(this, "ProgressEvent") ? this : null;
		}

		AsPromiseRejectionEvent() {
			return dom_is(this, "PromiseRejectionEvent") ? this : null;
		}

		AsSecurityPolicyViolationEvent() {
			return dom_is(this, "SecurityPolicyViolationEvent") ? this : null;
		}

		AsStorageEvent() {
			return dom_is(this, "StorageEvent") ? this : null;
		}

		AsSubmitEvent() {
			return dom_is(this, "SubmitEvent") ? this : null;
		}

		AsTouchEvent() {
			return dom_is(this, "TouchEvent") ? this : null;
		}

		AsTrackEvent() {
			return dom_is(this, "TrackEvent") ? this : null;
		}

		AsTransitionEvent() {
			return dom_is(this, "TransitionEvent") ? this : null;
		}

		AsUIEvent() {
			return dom_is(this, "UIEvent") ? this : null;
		}

		AsWheelEvent() {
			return dom_is(this, "WheelEvent") ? this : null;
		}

		ComposedPath() {
			return Array.from(this.handle.composedPath() || [], (item) => dom_wrap(dom.EventTarget, item));
		}

		InitEvent(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			this.handle.initEvent(p0, ...p1.map((item) => item));
		}

		PreventDefault() {
			this.handle.preventDefault();
		}

		StopImmediatePropagation() {
			this.handle.stopImmediatePropagation();
		}

		StopPropagation() {
			this.handle.stopPropagation();
		}

		Bubbles() {
			return this.handle.bubbles;
		}

		CancelBubble() {
			return this.handle.cancelBubble;
		}

		SetCancelBubble(value) {
			this.handle.cancelBubble = value;
		}

		Cancelable() {
			return this.handle.cancelable;
		}

		Composed() {
			return this.handle.composed;
		}

		CurrentTarget() {
			return dom_wrap(dom.EventTarget, this.handle.currentTarget);
		}

		DefaultPrevented() {
			return this.handle.defaultPrevented;
		}

		EventPhase() {
			return this.handle.eventPhase;
		}

		IsTrusted() {
			return this.handle.isTrusted;
		}

		ReturnValue() {
			return this.handle.returnValue;
		}

		SetReturnValue(value) {
			this.handle.returnValue = value;
		}

		SrcElement() {
			return dom_wrap(dom.EventTarget, this.handle.srcElement);
		}

		Target() {
			return dom_wrap(dom.EventTarget, this.handle.target);
		}

		TimeStamp() {
			return this.handle.timeStamp;
		}

		Type() {
			return this.handle.type;
		}

		NONE() {
			return this.handle.NONE;
		}

		CAPTURING_PHASE() {
			return this.handle.CAPTURING_PHASE;
		}

		AT_TARGET() {
			return this.handle.AT_TARGET;
		}

		BUBBLING_PHASE() {
			return this.handle.BUBBLING_PHASE;
		}
}

class dom_C_AnimationEvent extends dom_C_Event {

		KindOfAnimationEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		AnimationName() {
			return this.handle.animationName;
		}

		ElapsedTime() {
			return this.handle.elapsedTime;
		}

		PseudoElement() {
			return this.handle.pseudoElement;
		}
}

class dom_C_AnimationFrameProvider extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfAnimationFrameProvider() {
			
		}

		AsWindow() {
			return dom_is(this, "Window") ? this : null;
		}

		CancelAnimationFrame(p0) {
			this.handle.cancelAnimationFrame(p0);
		}

		RequestAnimationFrame(p0) {
			return this.handle.requestAnimationFrame((p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [a[0]])));
		}
}

class dom_C_AnimationPlaybackEvent extends dom_C_Event {

		KindOfAnimationPlaybackEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		CurrentTime() {
			return this.handle.currentTime;
		}

		TimelineTime() {
			return this.handle.timelineTime;
		}
}

class dom_C_Node extends dom_C_EventTarget {

		KindOfNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		AsAttr() {
			return dom_is(this, "Attr") ? this : null;
		}

		AsCDATASection() {
			return dom_is(this, "CDATASection") ? this : null;
		}

		AsCharacterData() {
			return dom_is(this, "CharacterData") ? this : null;
		}

		AsChildNode() {
			return dom_is(this, "ChildNode") ? this : null;
		}

		AsComment() {
			return dom_is(this, "Comment") ? this : null;
		}

		AsDocument() {
			return dom_is(this, "Document") ? this : null;
		}

		AsDocumentType() {
			return dom_is(this, "DocumentType") ? this : null;
		}

		AsDocumentFragment() {
			return dom_is(this, "DocumentFragment") ? this : null;
		}

		AsElement() {
			return dom_is(this, "Element") ? this : null;
		}

		AsHTMLAudioElement() {
			return dom_is(this, "HTMLAudioElement") ? this : null;
		}

		AsHTMLCanvasElement() {
			return dom_is(this, "HTMLCanvasElement") ? this : null;
		}

		AsHTMLElement() {
			return dom_is(this, "HTMLElement") ? this : null;
		}

		AsHTMLFormElement() {
			return dom_is(this, "HTMLFormElement") ? this : null;
		}

		AsHTMLHeadElement() {
			return dom_is(this, "HTMLHeadElement") ? this : null;
		}

		AsHTMLImageElement() {
			return dom_is(this, "HTMLImageElement") ? this : null;
		}

		AsHTMLMediaElement() {
			return dom_is(this, "HTMLMediaElement") ? this : null;
		}

		AsHTMLSlotElement() {
			return dom_is(this, "HTMLSlotElement") ? this : null;
		}

		AsParentNode() {
			return dom_is(this, "ParentNode") ? this : null;
		}

		AsProcessingInstruction() {
			return dom_is(this, "ProcessingInstruction") ? this : null;
		}

		AsSVGElement() {
			return dom_is(this, "SVGElement") ? this : null;
		}

		AsSVGGraphicsElement() {
			return dom_is(this, "SVGGraphicsElement") ? this : null;
		}

		AsSVGSVGElement() {
			return dom_is(this, "SVGSVGElement") ? this : null;
		}

		AsShadowRoot() {
			return dom_is(this, "ShadowRoot") ? this : null;
		}

		AsText() {
			return dom_is(this, "Text") ? this : null;
		}

		AppendChild(p0) {
			return dom_wrap(dom.Node, this.handle.appendChild(dom_unwrap(p0)));
		}

		CloneNode(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.Node, this.handle.cloneNode(...p0.map((item) => item)));
		}

		CompareDocumentPosition(p0) {
			return this.handle.compareDocumentPosition(dom_unwrap(p0));
		}

		Contains(p0) {
			return this.handle.contains(dom_unwrap(p0));
		}

		GetRootNode(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.Node, this.handle.getRootNode(...p0.map((item) => dom_toGetRootNodeOptions(item))));
		}

		HasChildNodes() {
			return this.handle.hasChildNodes();
		}

		InsertBefore(p0, p1) {
			return dom_wrap(dom.Node, this.handle.insertBefore(dom_unwrap(p0), dom_unwrap(p1)));
		}

		IsDefaultNamespace(p0) {
			return this.handle.isDefaultNamespace(p0);
		}

		IsEqualNode(p0) {
			return this.handle.isEqualNode(dom_unwrap(p0));
		}

		IsSameNode(p0) {
			return this.handle.isSameNode(dom_unwrap(p0));
		}

		LookupNamespaceURI(p0) {
			return this.handle.lookupNamespaceURI(p0);
		}

		LookupPrefix(p0) {
			return this.handle.lookupPrefix(p0);
		}

		Normalize() {
			this.handle.normalize();
		}

		RemoveChild(p0) {
			return dom_wrap(dom.Node, this.handle.removeChild(dom_unwrap(p0)));
		}

		ReplaceChild(p0, p1) {
			return dom_wrap(dom.Node, this.handle.replaceChild(dom_unwrap(p0), dom_unwrap(p1)));
		}

		BaseURI() {
			return this.handle.baseURI;
		}

		ChildNodes() {
			return dom_wrap(dom.NodeListOf, this.handle.childNodes);
		}

		FirstChild() {
			return dom_wrap(dom.ChildNode, this.handle.firstChild);
		}

		IsConnected() {
			return this.handle.isConnected;
		}

		LastChild() {
			return dom_wrap(dom.ChildNode, this.handle.lastChild);
		}

		NextSibling() {
			return dom_wrap(dom.ChildNode, this.handle.nextSibling);
		}

		NodeName() {
			return this.handle.nodeName;
		}

		NodeType() {
			return this.handle.nodeType;
		}

		NodeValue() {
			return this.handle.nodeValue;
		}

		SetNodeValue(value) {
			this.handle.nodeValue = value;
		}

		OwnerDocument() {
			return dom_wrap(dom.Document, this.handle.ownerDocument);
		}

		ParentElement() {
			return dom_wrap(dom.HTMLElement, this.handle.parentElement);
		}

		ParentNode() {
			return dom_wrap(dom.ParentNode, this.handle.parentNode);
		}

		PreviousSibling() {
			return dom_wrap(dom.ChildNode, this.handle.previousSibling);
		}

		TextContent() {
			return this.handle.textContent;
		}

		SetTextContent(value) {
			this.handle.textContent = value;
		}

		ELEMENT_NODE() {
			return this.handle.ELEMENT_NODE;
		}

		ATTRIBUTE_NODE() {
			return this.handle.ATTRIBUTE_NODE;
		}

		TEXT_NODE() {
			return this.handle.TEXT_NODE;
		}

		CDATA_SECTION_NODE() {
			return this.handle.CDATA_SECTION_NODE;
		}

		ENTITY_REFERENCE_NODE() {
			return this.handle.ENTITY_REFERENCE_NODE;
		}

		ENTITY_NODE() {
			return this.handle.ENTITY_NODE;
		}

		PROCESSING_INSTRUCTION_NODE() {
			return this.handle.PROCESSING_INSTRUCTION_NODE;
		}

		COMMENT_NODE() {
			return this.handle.COMMENT_NODE;
		}

		DOCUMENT_NODE() {
			return this.handle.DOCUMENT_NODE;
		}

		DOCUMENT_TYPE_NODE() {
			return this.handle.DOCUMENT_TYPE_NODE;
		}

		DOCUMENT_FRAGMENT_NODE() {
			return this.handle.DOCUMENT_FRAGMENT_NODE;
		}

		NOTATION_NODE() {
			return this.handle.NOTATION_NODE;
		}

		DOCUMENT_POSITION_DISCONNECTED() {
			return this.handle.DOCUMENT_POSITION_DISCONNECTED;
		}

		DOCUMENT_POSITION_PRECEDING() {
			return this.handle.DOCUMENT_POSITION_PRECEDING;
		}

		DOCUMENT_POSITION_FOLLOWING() {
			return this.handle.DOCUMENT_POSITION_FOLLOWING;
		}

		DOCUMENT_POSITION_CONTAINS() {
			return this.handle.DOCUMENT_POSITION_CONTAINS;
		}

		DOCUMENT_POSITION_CONTAINED_BY() {
			return this.handle.DOCUMENT_POSITION_CONTAINED_BY;
		}

		DOCUMENT_POSITION_IMPLEMENTATION_SPECIFIC() {
			return this.handle.DOCUMENT_POSITION_IMPLEMENTATION_SPECIFIC;
		}
}

class dom_C_Attr extends dom_C_Node {

		KindOfAttr() {
			
		}

		KindOfNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		LocalName() {
			return this.handle.localName;
		}

		Name() {
			return this.handle.name;
		}

		NamespaceURI() {
			return this.handle.namespaceURI;
		}

		OwnerElement() {
			return dom_wrap(dom.Element, this.handle.ownerElement);
		}

		Prefix() {
			return this.handle.prefix;
		}

		Specified() {
			return this.handle.specified;
		}

		Value() {
			return this.handle.value;
		}

		SetValue(value) {
			this.handle.value = value;
		}
}

class dom_C_AudioBuffer extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfAudioBuffer() {
			
		}

		CopyFromChannel(p0, p1, ...p2) {
			if (p2.length === 1 && Array.isArray(p2[0]) && p2[0].length === 0) {
				p2 = [];
			}
			
			this.handle.copyFromChannel(Float32Array.from(p0 || []), p1, ...p2.map((item) => item));
		}

		CopyToChannel(p0, p1, ...p2) {
			if (p2.length === 1 && Array.isArray(p2[0]) && p2[0].length === 0) {
				p2 = [];
			}
			
			this.handle.copyToChannel(Float32Array.from(p0 || []), p1, ...p2.map((item) => item));
		}

		GetChannelData(p0) {
			return Array.from(this.handle.getChannelData(p0) || []);
		}

		Duration() {
			return this.handle.duration;
		}

		Length() {
			return this.handle.length;
		}

		NumberOfChannels() {
			return this.handle.numberOfChannels;
		}

		SampleRate() {
			return this.handle.sampleRate;
		}
}

class dom_C_AudioScheduledSourceNode extends dom_C_AudioNode {

		KindOfAudioScheduledSourceNode() {
			
		}

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsAudioNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		AsAudioBufferSourceNode() {
			return dom_is(this, "AudioBufferSourceNode") ? this : null;
		}

		AsConstantSourceNode() {
			return dom_is(this, "ConstantSourceNode") ? this : null;
		}

		AsOscillatorNode() {
			return dom_is(this, "OscillatorNode") ? this : null;
		}

		Start(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.start(...p0.map((item) => item));
		}

		Stop(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.stop(...p0.map((item) => item));
		}

		OnEnded(p0) {
			this.handle.addEventListener("ended", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		Onended(p0) {
			return this.handle.onended(dom_unwrap(p0));
		}
}

class dom_C_AudioBufferSourceNode extends dom_C_AudioScheduledSourceNode {

		KindOfAudioBufferSourceNode() {
			
		}

		KindOfAudioScheduledSourceNode() {
			
		}

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsAudioScheduledSourceNode() {
			return this;
		}

		AsAudioNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		Buffer() {
			return dom_wrap(dom.AudioBuffer, this.handle.buffer);
		}

		SetBuffer(value) {
			this.handle.buffer = dom_unwrap(value);
		}

		Detune() {
			return dom_wrap(dom.AudioParam, this.handle.detune);
		}

		Loop() {
			return this.handle.loop;
		}

		SetLoop(value) {
			this.handle.loop = value;
		}

		LoopEnd() {
			return this.handle.loopEnd;
		}

		SetLoopEnd(value) {
			this.handle.loopEnd = value;
		}

		LoopStart() {
			return this.handle.loopStart;
		}

		SetLoopStart(value) {
			this.handle.loopStart = value;
		}

		PlaybackRate() {
			return dom_wrap(dom.AudioParam, this.handle.playbackRate);
		}
}

class dom_C_BaseAudioContext extends dom_C_EventTarget {

		KindOfBaseAudioContext() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		AsAudioContext() {
			return dom_is(this, "AudioContext") ? this : null;
		}

		CreateAnalyser() {
			return dom_wrap(dom.AnalyserNode, this.handle.createAnalyser());
		}

		CreateBiquadFilter() {
			return dom_wrap(dom.BiquadFilterNode, this.handle.createBiquadFilter());
		}

		CreateBuffer(p0, p1, p2) {
			return dom_wrap(dom.AudioBuffer, this.handle.createBuffer(p0, p1, p2));
		}

		CreateBufferSource() {
			return dom_wrap(dom.AudioBufferSourceNode, this.handle.createBufferSource());
		}

		CreateChannelMerger(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.ChannelMergerNode, this.handle.createChannelMerger(...p0.map((item) => item)));
		}

		CreateChannelSplitter(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.ChannelSplitterNode, this.handle.createChannelSplitter(...p0.map((item) => item)));
		}

		CreateConstantSource() {
			return dom_wrap(dom.ConstantSourceNode, this.handle.createConstantSource());
		}

		CreateConvolver() {
			return dom_wrap(dom.ConvolverNode, this.handle.createConvolver());
		}

		CreateDelay(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DelayNode, this.handle.createDelay(...p0.map((item) => item)));
		}

		CreateDynamicsCompressor() {
			return dom_wrap(dom.DynamicsCompressorNode, this.handle.createDynamicsCompressor());
		}

		CreateGain() {
			return dom_wrap(dom.GainNode, this.handle.createGain());
		}

		CreateIIRFilter(p0, p1) {
			return dom_wrap(dom.IIRFilterNode, this.handle.createIIRFilter(Array.from(p0 || []), Array.from(p1 || [])));
		}

		CreateOscillator() {
			return dom_wrap(dom.OscillatorNode, this.handle.createOscillator());
		}

		CreatePanner() {
			return dom_wrap(dom.PannerNode, this.handle.createPanner());
		}

		CreatePeriodicWave(p0, p1, ...p2) {
			if (p2.length === 1 && Array.isArray(p2[0]) && p2[0].length === 0) {
				p2 = [];
			}
			
			return this.handle.createPeriodicWave(dom_unwrapAny(p0), dom_unwrapAny(p1), ...p2.map((item) => dom_toPeriodicWaveConstraints(item)));
		}

		CreateScriptProcessor(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.ScriptProcessorNode, this.handle.createScriptProcessor(...p0.map((item) => item)));
		}

		CreateStereoPanner() {
			return dom_wrap(dom.StereoPannerNode, this.handle.createStereoPanner());
		}

		CreateWaveShaper() {
			return dom_wrap(dom.WaveShaperNode, this.handle.createWaveShaper());
		}

		*DecodeAudioData(p0, ...p1) {
			const outcome = yield* dom_await(() => this.handle.decodeAudioData(Uint8Array.from(p0 || []), ...p1.map((item) => dom_unwrapAny(item))));
			
			return outcome.ok ? [dom_wrap(dom.AudioBuffer, outcome.value), null] : [null, outcome.error];
		}

		OnStatechange(p0) {
			this.handle.addEventListener("statechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		Onstatechange(p0) {
			return this.handle.onstatechange(dom_unwrap(p0));
		}

		AudioWorklet() {
			return dom_wrap(dom.AudioWorklet, this.handle.audioWorklet);
		}

		CurrentTime() {
			return this.handle.currentTime;
		}

		Destination() {
			return dom_wrap(dom.AudioDestinationNode, this.handle.destination);
		}

		Listener() {
			return dom_wrap(dom.AudioListener, this.handle.listener);
		}

		SampleRate() {
			return this.handle.sampleRate;
		}

		State() {
			return this.handle.state;
		}
}

class dom_C_AudioContext extends dom_C_BaseAudioContext {

		KindOfAudioContext() {
			
		}

		KindOfBaseAudioContext() {
			
		}

		KindOfEventTarget() {
			
		}

		AsBaseAudioContext() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		*Close() {
			const outcome = yield* dom_await(() => this.handle.close());
			
			return outcome.ok ? null : outcome.error;
		}

		CreateMediaElementSource(p0) {
			return dom_wrap(dom.MediaElementAudioSourceNode, this.handle.createMediaElementSource(dom_unwrap(p0)));
		}

		CreateMediaStreamDestination() {
			return dom_wrap(dom.MediaStreamAudioDestinationNode, this.handle.createMediaStreamDestination());
		}

		CreateMediaStreamSource(p0) {
			return dom_wrap(dom.MediaStreamAudioSourceNode, this.handle.createMediaStreamSource(dom_unwrap(p0)));
		}

		GetOutputTimestamp() {
			return dom_fromAudioTimestamp(this.handle.getOutputTimestamp());
		}

		*Resume() {
			const outcome = yield* dom_await(() => this.handle.resume());
			
			return outcome.ok ? null : outcome.error;
		}

		*Suspend() {
			const outcome = yield* dom_await(() => this.handle.suspend());
			
			return outcome.ok ? null : outcome.error;
		}

		BaseLatency() {
			return this.handle.baseLatency;
		}

		OutputLatency() {
			return this.handle.outputLatency;
		}
}

class dom_C_AudioContextType extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfAudioContextType() {
			
		}

		Prototype() {
			return dom_wrap(dom.AudioContext, this.handle.prototype);
		}

		SetPrototype(value) {
			this.handle.prototype = dom_unwrap(value);
		}
}

class dom_C_AudioDestinationNode extends dom_C_AudioNode {

		KindOfAudioDestinationNode() {
			
		}

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsAudioNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		MaxChannelCount() {
			return this.handle.maxChannelCount;
		}
}

class dom_C_AudioListener extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfAudioListener() {
			
		}

		SetOrientation(p0, p1, p2, p3, p4, p5) {
			this.handle.setOrientation(p0, p1, p2, p3, p4, p5);
		}

		SetPosition(p0, p1, p2) {
			this.handle.setPosition(p0, p1, p2);
		}

		ForwardX() {
			return dom_wrap(dom.AudioParam, this.handle.forwardX);
		}

		ForwardY() {
			return dom_wrap(dom.AudioParam, this.handle.forwardY);
		}

		ForwardZ() {
			return dom_wrap(dom.AudioParam, this.handle.forwardZ);
		}

		PositionX() {
			return dom_wrap(dom.AudioParam, this.handle.positionX);
		}

		PositionY() {
			return dom_wrap(dom.AudioParam, this.handle.positionY);
		}

		PositionZ() {
			return dom_wrap(dom.AudioParam, this.handle.positionZ);
		}

		UpX() {
			return dom_wrap(dom.AudioParam, this.handle.upX);
		}

		UpY() {
			return dom_wrap(dom.AudioParam, this.handle.upY);
		}

		UpZ() {
			return dom_wrap(dom.AudioParam, this.handle.upZ);
		}
}

class dom_C_AudioParam extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfAudioParam() {
			
		}

		CancelAndHoldAtTime(p0) {
			return dom_wrap(dom.AudioParam, this.handle.cancelAndHoldAtTime(p0));
		}

		CancelScheduledValues(p0) {
			return dom_wrap(dom.AudioParam, this.handle.cancelScheduledValues(p0));
		}

		ExponentialRampToValueAtTime(p0, p1) {
			return dom_wrap(dom.AudioParam, this.handle.exponentialRampToValueAtTime(p0, p1));
		}

		LinearRampToValueAtTime(p0, p1) {
			return dom_wrap(dom.AudioParam, this.handle.linearRampToValueAtTime(p0, p1));
		}

		SetTargetAtTime(p0, p1, p2) {
			return dom_wrap(dom.AudioParam, this.handle.setTargetAtTime(p0, p1, p2));
		}

		SetValueAtTime(p0, p1) {
			return dom_wrap(dom.AudioParam, this.handle.setValueAtTime(p0, p1));
		}

		SetValueCurveAtTime(p0, p1, p2) {
			return dom_wrap(dom.AudioParam, this.handle.setValueCurveAtTime(dom_unwrapAny(p0), p1, p2));
		}

		AutomationRate() {
			return this.handle.automationRate;
		}

		SetAutomationRate(value) {
			this.handle.automationRate = value;
		}

		DefaultValue() {
			return this.handle.defaultValue;
		}

		MaxValue() {
			return this.handle.maxValue;
		}

		MinValue() {
			return this.handle.minValue;
		}

		Value() {
			return this.handle.value;
		}

		SetValue(value) {
			this.handle.value = value;
		}
}

class dom_C_AudioProcessingEvent extends dom_C_Event {

		KindOfAudioProcessingEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		InputBuffer() {
			return dom_wrap(dom.AudioBuffer, this.handle.inputBuffer);
		}

		OutputBuffer() {
			return dom_wrap(dom.AudioBuffer, this.handle.outputBuffer);
		}

		PlaybackTime() {
			return this.handle.playbackTime;
		}
}

class dom_C_Worklet extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfWorklet() {
			
		}

		AsAudioWorklet() {
			return dom_is(this, "AudioWorklet") ? this : null;
		}

		*AddModule(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			const outcome = yield* dom_await(() => this.handle.addModule(dom_unwrapAny(p0), ...p1.map((item) => dom_toWorkletOptions(item))));
			
			return outcome.ok ? null : outcome.error;
		}
}

class dom_C_AudioWorklet extends dom_C_Worklet {

		KindOfAudioWorklet() {
			
		}

		KindOfWorklet() {
			
		}

		AsWorklet() {
			return this;
		}
}

class dom_C_BeforeUnloadEvent extends dom_C_Event {

		KindOfBeforeUnloadEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		SetReturnValue(value) {
			this.handle.returnValue = dom_unwrapAny(value);
		}
}

class dom_C_BiquadFilterNode extends dom_C_AudioNode {

		KindOfBiquadFilterNode() {
			
		}

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsAudioNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		GetFrequencyResponse(p0, p1, p2) {
			this.handle.getFrequencyResponse(Float32Array.from(p0 || []), Float32Array.from(p1 || []), Float32Array.from(p2 || []));
		}

		Q() {
			return dom_wrap(dom.AudioParam, this.handle.Q);
		}

		Detune() {
			return dom_wrap(dom.AudioParam, this.handle.detune);
		}

		Frequency() {
			return dom_wrap(dom.AudioParam, this.handle.frequency);
		}

		Gain() {
			return dom_wrap(dom.AudioParam, this.handle.gain);
		}

		Type() {
			return this.handle.type;
		}

		SetType(value) {
			this.handle.type = value;
		}
}

class dom_C_Blob extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfBlob() {
			
		}

		*ArrayBuffer() {
			const outcome = yield* dom_await(() => this.handle.arrayBuffer());
			
			return outcome.ok ? [Array.from(outcome.value || []), null] : [[], outcome.error];
		}

		Slice(...p0) {
			return dom_wrap(dom.Blob, this.handle.slice(...p0.map((item) => dom_unwrapAny(item))));
		}

		Stream() {
			return dom_wrap(dom.ReadableStream, this.handle.stream());
		}

		*Text() {
			const outcome = yield* dom_await(() => this.handle.text());
			
			return outcome.ok ? [outcome.value, null] : ["", outcome.error];
		}

		Size() {
			return this.handle.size;
		}

		Type() {
			return this.handle.type;
		}
}

class dom_C_Body extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfBody() {
			
		}

		AsResponse() {
			return dom_is(this, "Response") ? this : null;
		}

		*ArrayBuffer() {
			const outcome = yield* dom_await(() => this.handle.arrayBuffer());
			
			return outcome.ok ? [Array.from(outcome.value || []), null] : [[], outcome.error];
		}

		*Blob() {
			const outcome = yield* dom_await(() => this.handle.blob());
			
			return outcome.ok ? [dom_wrap(dom.Blob, outcome.value), null] : [null, outcome.error];
		}

		*FormData() {
			const outcome = yield* dom_await(() => this.handle.formData());
			
			return outcome.ok ? [dom_wrap(dom.FormData, outcome.value), null] : [null, outcome.error];
		}

		*Json() {
			const outcome = yield* dom_await(() => this.handle.json());
			
			return outcome.ok ? [outcome.value, null] : [null, outcome.error];
		}

		*Text() {
			const outcome = yield* dom_await(() => this.handle.text());
			
			return outcome.ok ? [outcome.value, null] : ["", outcome.error];
		}

		Body() {
			return dom_wrap(dom.ReadableStream, this.handle.body);
		}

		BodyUsed() {
			return this.handle.bodyUsed;
		}
}

class dom_C_ChildNode extends dom_C_Node {

		KindOfChildNode() {
			
		}

		KindOfNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		AsCDATASection() {
			return dom_is(this, "CDATASection") ? this : null;
		}

		AsCharacterData() {
			return dom_is(this, "CharacterData") ? this : null;
		}

		AsComment() {
			return dom_is(this, "Comment") ? this : null;
		}

		AsDocumentType() {
			return dom_is(this, "DocumentType") ? this : null;
		}

		AsElement() {
			return dom_is(this, "Element") ? this : null;
		}

		AsHTMLAudioElement() {
			return dom_is(this, "HTMLAudioElement") ? this : null;
		}

		AsHTMLCanvasElement() {
			return dom_is(this, "HTMLCanvasElement") ? this : null;
		}

		AsHTMLElement() {
			return dom_is(this, "HTMLElement") ? this : null;
		}

		AsHTMLFormElement() {
			return dom_is(this, "HTMLFormElement") ? this : null;
		}

		AsHTMLHeadElement() {
			return dom_is(this, "HTMLHeadElement") ? this : null;
		}

		AsHTMLImageElement() {
			return dom_is(this, "HTMLImageElement") ? this : null;
		}

		AsHTMLMediaElement() {
			return dom_is(this, "HTMLMediaElement") ? this : null;
		}

		AsHTMLSlotElement() {
			return dom_is(this, "HTMLSlotElement") ? this : null;
		}

		AsProcessingInstruction() {
			return dom_is(this, "ProcessingInstruction") ? this : null;
		}

		AsSVGElement() {
			return dom_is(this, "SVGElement") ? this : null;
		}

		AsSVGGraphicsElement() {
			return dom_is(this, "SVGGraphicsElement") ? this : null;
		}

		AsSVGSVGElement() {
			return dom_is(this, "SVGSVGElement") ? this : null;
		}

		AsText() {
			return dom_is(this, "Text") ? this : null;
		}

		After(...p0) {
			this.handle.after(...p0.map((item) => dom_unwrapAny(item)));
		}

		Before(...p0) {
			this.handle.before(...p0.map((item) => dom_unwrapAny(item)));
		}

		Remove() {
			this.handle.remove();
		}

		ReplaceWith(...p0) {
			this.handle.replaceWith(...p0.map((item) => dom_unwrapAny(item)));
		}
}

class dom_C_NonDocumentTypeChildNode extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfNonDocumentTypeChildNode() {
			
		}

		AsCDATASection() {
			return dom_is(this, "CDATASection") ? this : null;
		}

		AsCharacterData() {
			return dom_is(this, "CharacterData") ? this : null;
		}

		AsComment() {
			return dom_is(this, "Comment") ? this : null;
		}

		AsElement() {
			return dom_is(this, "Element") ? this : null;
		}

		AsHTMLAudioElement() {
			return dom_is(this, "HTMLAudioElement") ? this : null;
		}

		AsHTMLCanvasElement() {
			return dom_is(this, "HTMLCanvasElement") ? this : null;
		}

		AsHTMLElement() {
			return dom_is(this, "HTMLElement") ? this : null;
		}

		AsHTMLFormElement() {
			return dom_is(this, "HTMLFormElement") ? this : null;
		}

		AsHTMLHeadElement() {
			return dom_is(this, "HTMLHeadElement") ? this : null;
		}

		AsHTMLImageElement() {
			return dom_is(this, "HTMLImageElement") ? this : null;
		}

		AsHTMLMediaElement() {
			return dom_is(this, "HTMLMediaElement") ? this : null;
		}

		AsHTMLSlotElement() {
			return dom_is(this, "HTMLSlotElement") ? this : null;
		}

		AsProcessingInstruction() {
			return dom_is(this, "ProcessingInstruction") ? this : null;
		}

		AsSVGElement() {
			return dom_is(this, "SVGElement") ? this : null;
		}

		AsSVGGraphicsElement() {
			return dom_is(this, "SVGGraphicsElement") ? this : null;
		}

		AsSVGSVGElement() {
			return dom_is(this, "SVGSVGElement") ? this : null;
		}

		AsText() {
			return dom_is(this, "Text") ? this : null;
		}

		NextElementSibling() {
			return dom_wrap(dom.Element, this.handle.nextElementSibling);
		}

		PreviousElementSibling() {
			return dom_wrap(dom.Element, this.handle.previousElementSibling);
		}
}

class dom_C_CharacterData extends dom_C_Node {

		KindOfCharacterData() {
			
		}

		KindOfNode() {
			
		}

		KindOfChildNode() {
			
		}

		KindOfNonDocumentTypeChildNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsNode() {
			return this;
		}

		AsChildNode() {
			return this;
		}

		AsNonDocumentTypeChildNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		AsCDATASection() {
			return dom_is(this, "CDATASection") ? this : null;
		}

		AsComment() {
			return dom_is(this, "Comment") ? this : null;
		}

		AsProcessingInstruction() {
			return dom_is(this, "ProcessingInstruction") ? this : null;
		}

		AsText() {
			return dom_is(this, "Text") ? this : null;
		}

		AppendData(p0) {
			this.handle.appendData(p0);
		}

		DeleteData(p0, p1) {
			this.handle.deleteData(p0, p1);
		}

		InsertData(p0, p1) {
			this.handle.insertData(p0, p1);
		}

		ReplaceData(p0, p1, p2) {
			this.handle.replaceData(p0, p1, p2);
		}

		SubstringData(p0, p1) {
			return this.handle.substringData(p0, p1);
		}

		After(...p0) {
			this.handle.after(...p0.map((item) => dom_unwrapAny(item)));
		}

		Before(...p0) {
			this.handle.before(...p0.map((item) => dom_unwrapAny(item)));
		}

		Remove() {
			this.handle.remove();
		}

		ReplaceWith(...p0) {
			this.handle.replaceWith(...p0.map((item) => dom_unwrapAny(item)));
		}

		Data() {
			return this.handle.data;
		}

		SetData(value) {
			this.handle.data = value;
		}

		Length() {
			return this.handle.length;
		}

		NextElementSibling() {
			return dom_wrap(dom.Element, this.handle.nextElementSibling);
		}

		PreviousElementSibling() {
			return dom_wrap(dom.Element, this.handle.previousElementSibling);
		}
}

class dom_C_Slottable extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfSlottable() {
			
		}

		AsCDATASection() {
			return dom_is(this, "CDATASection") ? this : null;
		}

		AsElement() {
			return dom_is(this, "Element") ? this : null;
		}

		AsHTMLAudioElement() {
			return dom_is(this, "HTMLAudioElement") ? this : null;
		}

		AsHTMLCanvasElement() {
			return dom_is(this, "HTMLCanvasElement") ? this : null;
		}

		AsHTMLElement() {
			return dom_is(this, "HTMLElement") ? this : null;
		}

		AsHTMLFormElement() {
			return dom_is(this, "HTMLFormElement") ? this : null;
		}

		AsHTMLHeadElement() {
			return dom_is(this, "HTMLHeadElement") ? this : null;
		}

		AsHTMLImageElement() {
			return dom_is(this, "HTMLImageElement") ? this : null;
		}

		AsHTMLMediaElement() {
			return dom_is(this, "HTMLMediaElement") ? this : null;
		}

		AsHTMLSlotElement() {
			return dom_is(this, "HTMLSlotElement") ? this : null;
		}

		AsSVGElement() {
			return dom_is(this, "SVGElement") ? this : null;
		}

		AsSVGGraphicsElement() {
			return dom_is(this, "SVGGraphicsElement") ? this : null;
		}

		AsSVGSVGElement() {
			return dom_is(this, "SVGSVGElement") ? this : null;
		}

		AsText() {
			return dom_is(this, "Text") ? this : null;
		}

		AssignedSlot() {
			return dom_wrap(dom.HTMLSlotElement, this.handle.assignedSlot);
		}
}

class dom_C_Text extends dom_C_CharacterData {

		KindOfText() {
			
		}

		KindOfCharacterData() {
			
		}

		KindOfSlottable() {
			
		}

		KindOfNode() {
			
		}

		KindOfChildNode() {
			
		}

		KindOfNonDocumentTypeChildNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsCharacterData() {
			return this;
		}

		AsSlottable() {
			return this;
		}

		AsNode() {
			return this;
		}

		AsChildNode() {
			return this;
		}

		AsNonDocumentTypeChildNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		AsCDATASection() {
			return dom_is(this, "CDATASection") ? this : null;
		}

		SplitText(p0) {
			return dom_wrap(dom.Text, this.handle.splitText(p0));
		}

		WholeText() {
			return this.handle.wholeText;
		}

		AssignedSlot() {
			return dom_wrap(dom.HTMLSlotElement, this.handle.assignedSlot);
		}
}

class dom_C_CDATASection extends dom_C_Text {

		KindOfCDATASection() {
			
		}

		KindOfText() {
			
		}

		KindOfCharacterData() {
			
		}

		KindOfSlottable() {
			
		}

		KindOfNode() {
			
		}

		KindOfChildNode() {
			
		}

		KindOfNonDocumentTypeChildNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsText() {
			return this;
		}

		AsCharacterData() {
			return this;
		}

		AsSlottable() {
			return this;
		}

		AsNode() {
			return this;
		}

		AsChildNode() {
			return this;
		}

		AsNonDocumentTypeChildNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}
}

class dom_C_CSSRule extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCSSRule() {
			
		}

		CssText() {
			return this.handle.cssText;
		}

		SetCssText(value) {
			this.handle.cssText = value;
		}

		ParentRule() {
			return dom_wrap(dom.CSSRule, this.handle.parentRule);
		}

		ParentStyleSheet() {
			return dom_wrap(dom.CSSStyleSheet, this.handle.parentStyleSheet);
		}

		Type() {
			return this.handle.type;
		}

		STYLE_RULE() {
			return this.handle.STYLE_RULE;
		}

		CHARSET_RULE() {
			return this.handle.CHARSET_RULE;
		}

		IMPORT_RULE() {
			return this.handle.IMPORT_RULE;
		}

		MEDIA_RULE() {
			return this.handle.MEDIA_RULE;
		}

		FONT_FACE_RULE() {
			return this.handle.FONT_FACE_RULE;
		}

		PAGE_RULE() {
			return this.handle.PAGE_RULE;
		}

		NAMESPACE_RULE() {
			return this.handle.NAMESPACE_RULE;
		}

		KEYFRAMES_RULE() {
			return this.handle.KEYFRAMES_RULE;
		}

		KEYFRAME_RULE() {
			return this.handle.KEYFRAME_RULE;
		}

		SUPPORTS_RULE() {
			return this.handle.SUPPORTS_RULE;
		}

		COUNTER_STYLE_RULE() {
			return this.handle.COUNTER_STYLE_RULE;
		}

		FONT_FEATURE_VALUES_RULE() {
			return this.handle.FONT_FEATURE_VALUES_RULE;
		}
}

class dom_C_CSSStyleDeclaration extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCSSStyleDeclaration() {
			
		}

		GetPropertyPriority(p0) {
			return this.handle.getPropertyPriority(p0);
		}

		GetPropertyValue(p0) {
			return this.handle.getPropertyValue(p0);
		}

		Item(p0) {
			return this.handle.item(p0);
		}

		RemoveProperty(p0) {
			return this.handle.removeProperty(p0);
		}

		SetProperty(p0, p1, ...p2) {
			if (p2.length === 1 && Array.isArray(p2[0]) && p2[0].length === 0) {
				p2 = [];
			}
			
			this.handle.setProperty(p0, p1, ...p2.map((item) => item));
		}

		AccentColor() {
			return this.handle.accentColor;
		}

		SetAccentColor(value) {
			this.handle.accentColor = value;
		}

		AlignContent() {
			return this.handle.alignContent;
		}

		SetAlignContent(value) {
			this.handle.alignContent = value;
		}

		AlignItems() {
			return this.handle.alignItems;
		}

		SetAlignItems(value) {
			this.handle.alignItems = value;
		}

		AlignSelf() {
			return this.handle.alignSelf;
		}

		SetAlignSelf(value) {
			this.handle.alignSelf = value;
		}

		AlignmentBaseline() {
			return this.handle.alignmentBaseline;
		}

		SetAlignmentBaseline(value) {
			this.handle.alignmentBaseline = value;
		}

		All() {
			return this.handle.all;
		}

		SetAll(value) {
			this.handle.all = value;
		}

		Animation() {
			return this.handle.animation;
		}

		SetAnimation(value) {
			this.handle.animation = value;
		}

		AnimationComposition() {
			return this.handle.animationComposition;
		}

		SetAnimationComposition(value) {
			this.handle.animationComposition = value;
		}

		AnimationDelay() {
			return this.handle.animationDelay;
		}

		SetAnimationDelay(value) {
			this.handle.animationDelay = value;
		}

		AnimationDirection() {
			return this.handle.animationDirection;
		}

		SetAnimationDirection(value) {
			this.handle.animationDirection = value;
		}

		AnimationDuration() {
			return this.handle.animationDuration;
		}

		SetAnimationDuration(value) {
			this.handle.animationDuration = value;
		}

		AnimationFillMode() {
			return this.handle.animationFillMode;
		}

		SetAnimationFillMode(value) {
			this.handle.animationFillMode = value;
		}

		AnimationIterationCount() {
			return this.handle.animationIterationCount;
		}

		SetAnimationIterationCount(value) {
			this.handle.animationIterationCount = value;
		}

		AnimationName() {
			return this.handle.animationName;
		}

		SetAnimationName(value) {
			this.handle.animationName = value;
		}

		AnimationPlayState() {
			return this.handle.animationPlayState;
		}

		SetAnimationPlayState(value) {
			this.handle.animationPlayState = value;
		}

		AnimationTimingFunction() {
			return this.handle.animationTimingFunction;
		}

		SetAnimationTimingFunction(value) {
			this.handle.animationTimingFunction = value;
		}

		Appearance() {
			return this.handle.appearance;
		}

		SetAppearance(value) {
			this.handle.appearance = value;
		}

		AspectRatio() {
			return this.handle.aspectRatio;
		}

		SetAspectRatio(value) {
			this.handle.aspectRatio = value;
		}

		BackdropFilter() {
			return this.handle.backdropFilter;
		}

		SetBackdropFilter(value) {
			this.handle.backdropFilter = value;
		}

		BackfaceVisibility() {
			return this.handle.backfaceVisibility;
		}

		SetBackfaceVisibility(value) {
			this.handle.backfaceVisibility = value;
		}

		Background() {
			return this.handle.background;
		}

		SetBackground(value) {
			this.handle.background = value;
		}

		BackgroundAttachment() {
			return this.handle.backgroundAttachment;
		}

		SetBackgroundAttachment(value) {
			this.handle.backgroundAttachment = value;
		}

		BackgroundBlendMode() {
			return this.handle.backgroundBlendMode;
		}

		SetBackgroundBlendMode(value) {
			this.handle.backgroundBlendMode = value;
		}

		BackgroundClip() {
			return this.handle.backgroundClip;
		}

		SetBackgroundClip(value) {
			this.handle.backgroundClip = value;
		}

		BackgroundColor() {
			return this.handle.backgroundColor;
		}

		SetBackgroundColor(value) {
			this.handle.backgroundColor = value;
		}

		BackgroundImage() {
			return this.handle.backgroundImage;
		}

		SetBackgroundImage(value) {
			this.handle.backgroundImage = value;
		}

		BackgroundOrigin() {
			return this.handle.backgroundOrigin;
		}

		SetBackgroundOrigin(value) {
			this.handle.backgroundOrigin = value;
		}

		BackgroundPosition() {
			return this.handle.backgroundPosition;
		}

		SetBackgroundPosition(value) {
			this.handle.backgroundPosition = value;
		}

		BackgroundPositionX() {
			return this.handle.backgroundPositionX;
		}

		SetBackgroundPositionX(value) {
			this.handle.backgroundPositionX = value;
		}

		BackgroundPositionY() {
			return this.handle.backgroundPositionY;
		}

		SetBackgroundPositionY(value) {
			this.handle.backgroundPositionY = value;
		}

		BackgroundRepeat() {
			return this.handle.backgroundRepeat;
		}

		SetBackgroundRepeat(value) {
			this.handle.backgroundRepeat = value;
		}

		BackgroundSize() {
			return this.handle.backgroundSize;
		}

		SetBackgroundSize(value) {
			this.handle.backgroundSize = value;
		}

		BaselineShift() {
			return this.handle.baselineShift;
		}

		SetBaselineShift(value) {
			this.handle.baselineShift = value;
		}

		BaselineSource() {
			return this.handle.baselineSource;
		}

		SetBaselineSource(value) {
			this.handle.baselineSource = value;
		}

		BlockSize() {
			return this.handle.blockSize;
		}

		SetBlockSize(value) {
			this.handle.blockSize = value;
		}

		Border() {
			return this.handle.border;
		}

		SetBorder(value) {
			this.handle.border = value;
		}

		BorderBlock() {
			return this.handle.borderBlock;
		}

		SetBorderBlock(value) {
			this.handle.borderBlock = value;
		}

		BorderBlockColor() {
			return this.handle.borderBlockColor;
		}

		SetBorderBlockColor(value) {
			this.handle.borderBlockColor = value;
		}

		BorderBlockEnd() {
			return this.handle.borderBlockEnd;
		}

		SetBorderBlockEnd(value) {
			this.handle.borderBlockEnd = value;
		}

		BorderBlockEndColor() {
			return this.handle.borderBlockEndColor;
		}

		SetBorderBlockEndColor(value) {
			this.handle.borderBlockEndColor = value;
		}

		BorderBlockEndStyle() {
			return this.handle.borderBlockEndStyle;
		}

		SetBorderBlockEndStyle(value) {
			this.handle.borderBlockEndStyle = value;
		}

		BorderBlockEndWidth() {
			return this.handle.borderBlockEndWidth;
		}

		SetBorderBlockEndWidth(value) {
			this.handle.borderBlockEndWidth = value;
		}

		BorderBlockStart() {
			return this.handle.borderBlockStart;
		}

		SetBorderBlockStart(value) {
			this.handle.borderBlockStart = value;
		}

		BorderBlockStartColor() {
			return this.handle.borderBlockStartColor;
		}

		SetBorderBlockStartColor(value) {
			this.handle.borderBlockStartColor = value;
		}

		BorderBlockStartStyle() {
			return this.handle.borderBlockStartStyle;
		}

		SetBorderBlockStartStyle(value) {
			this.handle.borderBlockStartStyle = value;
		}

		BorderBlockStartWidth() {
			return this.handle.borderBlockStartWidth;
		}

		SetBorderBlockStartWidth(value) {
			this.handle.borderBlockStartWidth = value;
		}

		BorderBlockStyle() {
			return this.handle.borderBlockStyle;
		}

		SetBorderBlockStyle(value) {
			this.handle.borderBlockStyle = value;
		}

		BorderBlockWidth() {
			return this.handle.borderBlockWidth;
		}

		SetBorderBlockWidth(value) {
			this.handle.borderBlockWidth = value;
		}

		BorderBottom() {
			return this.handle.borderBottom;
		}

		SetBorderBottom(value) {
			this.handle.borderBottom = value;
		}

		BorderBottomColor() {
			return this.handle.borderBottomColor;
		}

		SetBorderBottomColor(value) {
			this.handle.borderBottomColor = value;
		}

		BorderBottomLeftRadius() {
			return this.handle.borderBottomLeftRadius;
		}

		SetBorderBottomLeftRadius(value) {
			this.handle.borderBottomLeftRadius = value;
		}

		BorderBottomRightRadius() {
			return this.handle.borderBottomRightRadius;
		}

		SetBorderBottomRightRadius(value) {
			this.handle.borderBottomRightRadius = value;
		}

		BorderBottomStyle() {
			return this.handle.borderBottomStyle;
		}

		SetBorderBottomStyle(value) {
			this.handle.borderBottomStyle = value;
		}

		BorderBottomWidth() {
			return this.handle.borderBottomWidth;
		}

		SetBorderBottomWidth(value) {
			this.handle.borderBottomWidth = value;
		}

		BorderCollapse() {
			return this.handle.borderCollapse;
		}

		SetBorderCollapse(value) {
			this.handle.borderCollapse = value;
		}

		BorderColor() {
			return this.handle.borderColor;
		}

		SetBorderColor(value) {
			this.handle.borderColor = value;
		}

		BorderEndEndRadius() {
			return this.handle.borderEndEndRadius;
		}

		SetBorderEndEndRadius(value) {
			this.handle.borderEndEndRadius = value;
		}

		BorderEndStartRadius() {
			return this.handle.borderEndStartRadius;
		}

		SetBorderEndStartRadius(value) {
			this.handle.borderEndStartRadius = value;
		}

		BorderImage() {
			return this.handle.borderImage;
		}

		SetBorderImage(value) {
			this.handle.borderImage = value;
		}

		BorderImageOutset() {
			return this.handle.borderImageOutset;
		}

		SetBorderImageOutset(value) {
			this.handle.borderImageOutset = value;
		}

		BorderImageRepeat() {
			return this.handle.borderImageRepeat;
		}

		SetBorderImageRepeat(value) {
			this.handle.borderImageRepeat = value;
		}

		BorderImageSlice() {
			return this.handle.borderImageSlice;
		}

		SetBorderImageSlice(value) {
			this.handle.borderImageSlice = value;
		}

		BorderImageSource() {
			return this.handle.borderImageSource;
		}

		SetBorderImageSource(value) {
			this.handle.borderImageSource = value;
		}

		BorderImageWidth() {
			return this.handle.borderImageWidth;
		}

		SetBorderImageWidth(value) {
			this.handle.borderImageWidth = value;
		}

		BorderInline() {
			return this.handle.borderInline;
		}

		SetBorderInline(value) {
			this.handle.borderInline = value;
		}

		BorderInlineColor() {
			return this.handle.borderInlineColor;
		}

		SetBorderInlineColor(value) {
			this.handle.borderInlineColor = value;
		}

		BorderInlineEnd() {
			return this.handle.borderInlineEnd;
		}

		SetBorderInlineEnd(value) {
			this.handle.borderInlineEnd = value;
		}

		BorderInlineEndColor() {
			return this.handle.borderInlineEndColor;
		}

		SetBorderInlineEndColor(value) {
			this.handle.borderInlineEndColor = value;
		}

		BorderInlineEndStyle() {
			return this.handle.borderInlineEndStyle;
		}

		SetBorderInlineEndStyle(value) {
			this.handle.borderInlineEndStyle = value;
		}

		BorderInlineEndWidth() {
			return this.handle.borderInlineEndWidth;
		}

		SetBorderInlineEndWidth(value) {
			this.handle.borderInlineEndWidth = value;
		}

		BorderInlineStart() {
			return this.handle.borderInlineStart;
		}

		SetBorderInlineStart(value) {
			this.handle.borderInlineStart = value;
		}

		BorderInlineStartColor() {
			return this.handle.borderInlineStartColor;
		}

		SetBorderInlineStartColor(value) {
			this.handle.borderInlineStartColor = value;
		}

		BorderInlineStartStyle() {
			return this.handle.borderInlineStartStyle;
		}

		SetBorderInlineStartStyle(value) {
			this.handle.borderInlineStartStyle = value;
		}

		BorderInlineStartWidth() {
			return this.handle.borderInlineStartWidth;
		}

		SetBorderInlineStartWidth(value) {
			this.handle.borderInlineStartWidth = value;
		}

		BorderInlineStyle() {
			return this.handle.borderInlineStyle;
		}

		SetBorderInlineStyle(value) {
			this.handle.borderInlineStyle = value;
		}

		BorderInlineWidth() {
			return this.handle.borderInlineWidth;
		}

		SetBorderInlineWidth(value) {
			this.handle.borderInlineWidth = value;
		}

		BorderLeft() {
			return this.handle.borderLeft;
		}

		SetBorderLeft(value) {
			this.handle.borderLeft = value;
		}

		BorderLeftColor() {
			return this.handle.borderLeftColor;
		}

		SetBorderLeftColor(value) {
			this.handle.borderLeftColor = value;
		}

		BorderLeftStyle() {
			return this.handle.borderLeftStyle;
		}

		SetBorderLeftStyle(value) {
			this.handle.borderLeftStyle = value;
		}

		BorderLeftWidth() {
			return this.handle.borderLeftWidth;
		}

		SetBorderLeftWidth(value) {
			this.handle.borderLeftWidth = value;
		}

		BorderRadius() {
			return this.handle.borderRadius;
		}

		SetBorderRadius(value) {
			this.handle.borderRadius = value;
		}

		BorderRight() {
			return this.handle.borderRight;
		}

		SetBorderRight(value) {
			this.handle.borderRight = value;
		}

		BorderRightColor() {
			return this.handle.borderRightColor;
		}

		SetBorderRightColor(value) {
			this.handle.borderRightColor = value;
		}

		BorderRightStyle() {
			return this.handle.borderRightStyle;
		}

		SetBorderRightStyle(value) {
			this.handle.borderRightStyle = value;
		}

		BorderRightWidth() {
			return this.handle.borderRightWidth;
		}

		SetBorderRightWidth(value) {
			this.handle.borderRightWidth = value;
		}

		BorderSpacing() {
			return this.handle.borderSpacing;
		}

		SetBorderSpacing(value) {
			this.handle.borderSpacing = value;
		}

		BorderStartEndRadius() {
			return this.handle.borderStartEndRadius;
		}

		SetBorderStartEndRadius(value) {
			this.handle.borderStartEndRadius = value;
		}

		BorderStartStartRadius() {
			return this.handle.borderStartStartRadius;
		}

		SetBorderStartStartRadius(value) {
			this.handle.borderStartStartRadius = value;
		}

		BorderStyle() {
			return this.handle.borderStyle;
		}

		SetBorderStyle(value) {
			this.handle.borderStyle = value;
		}

		BorderTop() {
			return this.handle.borderTop;
		}

		SetBorderTop(value) {
			this.handle.borderTop = value;
		}

		BorderTopColor() {
			return this.handle.borderTopColor;
		}

		SetBorderTopColor(value) {
			this.handle.borderTopColor = value;
		}

		BorderTopLeftRadius() {
			return this.handle.borderTopLeftRadius;
		}

		SetBorderTopLeftRadius(value) {
			this.handle.borderTopLeftRadius = value;
		}

		BorderTopRightRadius() {
			return this.handle.borderTopRightRadius;
		}

		SetBorderTopRightRadius(value) {
			this.handle.borderTopRightRadius = value;
		}

		BorderTopStyle() {
			return this.handle.borderTopStyle;
		}

		SetBorderTopStyle(value) {
			this.handle.borderTopStyle = value;
		}

		BorderTopWidth() {
			return this.handle.borderTopWidth;
		}

		SetBorderTopWidth(value) {
			this.handle.borderTopWidth = value;
		}

		BorderWidth() {
			return this.handle.borderWidth;
		}

		SetBorderWidth(value) {
			this.handle.borderWidth = value;
		}

		Bottom() {
			return this.handle.bottom;
		}

		SetBottom(value) {
			this.handle.bottom = value;
		}

		BoxShadow() {
			return this.handle.boxShadow;
		}

		SetBoxShadow(value) {
			this.handle.boxShadow = value;
		}

		BoxSizing() {
			return this.handle.boxSizing;
		}

		SetBoxSizing(value) {
			this.handle.boxSizing = value;
		}

		BreakAfter() {
			return this.handle.breakAfter;
		}

		SetBreakAfter(value) {
			this.handle.breakAfter = value;
		}

		BreakBefore() {
			return this.handle.breakBefore;
		}

		SetBreakBefore(value) {
			this.handle.breakBefore = value;
		}

		BreakInside() {
			return this.handle.breakInside;
		}

		SetBreakInside(value) {
			this.handle.breakInside = value;
		}

		CaptionSide() {
			return this.handle.captionSide;
		}

		SetCaptionSide(value) {
			this.handle.captionSide = value;
		}

		CaretColor() {
			return this.handle.caretColor;
		}

		SetCaretColor(value) {
			this.handle.caretColor = value;
		}

		Clear() {
			return this.handle.clear;
		}

		SetClear(value) {
			this.handle.clear = value;
		}

		Clip() {
			return this.handle.clip;
		}

		SetClip(value) {
			this.handle.clip = value;
		}

		ClipPath() {
			return this.handle.clipPath;
		}

		SetClipPath(value) {
			this.handle.clipPath = value;
		}

		ClipRule() {
			return this.handle.clipRule;
		}

		SetClipRule(value) {
			this.handle.clipRule = value;
		}

		Color() {
			return this.handle.color;
		}

		SetColor(value) {
			this.handle.color = value;
		}

		ColorInterpolation() {
			return this.handle.colorInterpolation;
		}

		SetColorInterpolation(value) {
			this.handle.colorInterpolation = value;
		}

		ColorInterpolationFilters() {
			return this.handle.colorInterpolationFilters;
		}

		SetColorInterpolationFilters(value) {
			this.handle.colorInterpolationFilters = value;
		}

		ColorScheme() {
			return this.handle.colorScheme;
		}

		SetColorScheme(value) {
			this.handle.colorScheme = value;
		}

		ColumnCount() {
			return this.handle.columnCount;
		}

		SetColumnCount(value) {
			this.handle.columnCount = value;
		}

		ColumnFill() {
			return this.handle.columnFill;
		}

		SetColumnFill(value) {
			this.handle.columnFill = value;
		}

		ColumnGap() {
			return this.handle.columnGap;
		}

		SetColumnGap(value) {
			this.handle.columnGap = value;
		}

		ColumnRule() {
			return this.handle.columnRule;
		}

		SetColumnRule(value) {
			this.handle.columnRule = value;
		}

		ColumnRuleColor() {
			return this.handle.columnRuleColor;
		}

		SetColumnRuleColor(value) {
			this.handle.columnRuleColor = value;
		}

		ColumnRuleStyle() {
			return this.handle.columnRuleStyle;
		}

		SetColumnRuleStyle(value) {
			this.handle.columnRuleStyle = value;
		}

		ColumnRuleWidth() {
			return this.handle.columnRuleWidth;
		}

		SetColumnRuleWidth(value) {
			this.handle.columnRuleWidth = value;
		}

		ColumnSpan() {
			return this.handle.columnSpan;
		}

		SetColumnSpan(value) {
			this.handle.columnSpan = value;
		}

		ColumnWidth() {
			return this.handle.columnWidth;
		}

		SetColumnWidth(value) {
			this.handle.columnWidth = value;
		}

		Columns() {
			return this.handle.columns;
		}

		SetColumns(value) {
			this.handle.columns = value;
		}

		Contain() {
			return this.handle.contain;
		}

		SetContain(value) {
			this.handle.contain = value;
		}

		ContainIntrinsicBlockSize() {
			return this.handle.containIntrinsicBlockSize;
		}

		SetContainIntrinsicBlockSize(value) {
			this.handle.containIntrinsicBlockSize = value;
		}

		ContainIntrinsicHeight() {
			return this.handle.containIntrinsicHeight;
		}

		SetContainIntrinsicHeight(value) {
			this.handle.containIntrinsicHeight = value;
		}

		ContainIntrinsicInlineSize() {
			return this.handle.containIntrinsicInlineSize;
		}

		SetContainIntrinsicInlineSize(value) {
			this.handle.containIntrinsicInlineSize = value;
		}

		ContainIntrinsicSize() {
			return this.handle.containIntrinsicSize;
		}

		SetContainIntrinsicSize(value) {
			this.handle.containIntrinsicSize = value;
		}

		ContainIntrinsicWidth() {
			return this.handle.containIntrinsicWidth;
		}

		SetContainIntrinsicWidth(value) {
			this.handle.containIntrinsicWidth = value;
		}

		Container() {
			return this.handle.container;
		}

		SetContainer(value) {
			this.handle.container = value;
		}

		ContainerName() {
			return this.handle.containerName;
		}

		SetContainerName(value) {
			this.handle.containerName = value;
		}

		ContainerType() {
			return this.handle.containerType;
		}

		SetContainerType(value) {
			this.handle.containerType = value;
		}

		Content() {
			return this.handle.content;
		}

		SetContent(value) {
			this.handle.content = value;
		}

		ContentVisibility() {
			return this.handle.contentVisibility;
		}

		SetContentVisibility(value) {
			this.handle.contentVisibility = value;
		}

		CounterIncrement() {
			return this.handle.counterIncrement;
		}

		SetCounterIncrement(value) {
			this.handle.counterIncrement = value;
		}

		CounterReset() {
			return this.handle.counterReset;
		}

		SetCounterReset(value) {
			this.handle.counterReset = value;
		}

		CounterSet() {
			return this.handle.counterSet;
		}

		SetCounterSet(value) {
			this.handle.counterSet = value;
		}

		CssFloat() {
			return this.handle.cssFloat;
		}

		SetCssFloat(value) {
			this.handle.cssFloat = value;
		}

		CssText() {
			return this.handle.cssText;
		}

		SetCssText(value) {
			this.handle.cssText = value;
		}

		Cursor() {
			return this.handle.cursor;
		}

		SetCursor(value) {
			this.handle.cursor = value;
		}

		Cx() {
			return this.handle.cx;
		}

		SetCx(value) {
			this.handle.cx = value;
		}

		Cy() {
			return this.handle.cy;
		}

		SetCy(value) {
			this.handle.cy = value;
		}

		D() {
			return this.handle.d;
		}

		SetD(value) {
			this.handle.d = value;
		}

		Direction() {
			return this.handle.direction;
		}

		SetDirection(value) {
			this.handle.direction = value;
		}

		Display() {
			return this.handle.display;
		}

		SetDisplay(value) {
			this.handle.display = value;
		}

		DominantBaseline() {
			return this.handle.dominantBaseline;
		}

		SetDominantBaseline(value) {
			this.handle.dominantBaseline = value;
		}

		EmptyCells() {
			return this.handle.emptyCells;
		}

		SetEmptyCells(value) {
			this.handle.emptyCells = value;
		}

		Fill() {
			return this.handle.fill;
		}

		SetFill(value) {
			this.handle.fill = value;
		}

		FillOpacity() {
			return this.handle.fillOpacity;
		}

		SetFillOpacity(value) {
			this.handle.fillOpacity = value;
		}

		FillRule() {
			return this.handle.fillRule;
		}

		SetFillRule(value) {
			this.handle.fillRule = value;
		}

		Filter() {
			return this.handle.filter;
		}

		SetFilter(value) {
			this.handle.filter = value;
		}

		Flex() {
			return this.handle.flex;
		}

		SetFlex(value) {
			this.handle.flex = value;
		}

		FlexBasis() {
			return this.handle.flexBasis;
		}

		SetFlexBasis(value) {
			this.handle.flexBasis = value;
		}

		FlexDirection() {
			return this.handle.flexDirection;
		}

		SetFlexDirection(value) {
			this.handle.flexDirection = value;
		}

		FlexFlow() {
			return this.handle.flexFlow;
		}

		SetFlexFlow(value) {
			this.handle.flexFlow = value;
		}

		FlexGrow() {
			return this.handle.flexGrow;
		}

		SetFlexGrow(value) {
			this.handle.flexGrow = value;
		}

		FlexShrink() {
			return this.handle.flexShrink;
		}

		SetFlexShrink(value) {
			this.handle.flexShrink = value;
		}

		FlexWrap() {
			return this.handle.flexWrap;
		}

		SetFlexWrap(value) {
			this.handle.flexWrap = value;
		}

		Float() {
			return this.handle.float;
		}

		SetFloat(value) {
			this.handle.float = value;
		}

		FloodColor() {
			return this.handle.floodColor;
		}

		SetFloodColor(value) {
			this.handle.floodColor = value;
		}

		FloodOpacity() {
			return this.handle.floodOpacity;
		}

		SetFloodOpacity(value) {
			this.handle.floodOpacity = value;
		}

		Font() {
			return this.handle.font;
		}

		SetFont(value) {
			this.handle.font = value;
		}

		FontFamily() {
			return this.handle.fontFamily;
		}

		SetFontFamily(value) {
			this.handle.fontFamily = value;
		}

		FontFeatureSettings() {
			return this.handle.fontFeatureSettings;
		}

		SetFontFeatureSettings(value) {
			this.handle.fontFeatureSettings = value;
		}

		FontKerning() {
			return this.handle.fontKerning;
		}

		SetFontKerning(value) {
			this.handle.fontKerning = value;
		}

		FontOpticalSizing() {
			return this.handle.fontOpticalSizing;
		}

		SetFontOpticalSizing(value) {
			this.handle.fontOpticalSizing = value;
		}

		FontPalette() {
			return this.handle.fontPalette;
		}

		SetFontPalette(value) {
			this.handle.fontPalette = value;
		}

		FontSize() {
			return this.handle.fontSize;
		}

		SetFontSize(value) {
			this.handle.fontSize = value;
		}

		FontSizeAdjust() {
			return this.handle.fontSizeAdjust;
		}

		SetFontSizeAdjust(value) {
			this.handle.fontSizeAdjust = value;
		}

		FontStretch() {
			return this.handle.fontStretch;
		}

		SetFontStretch(value) {
			this.handle.fontStretch = value;
		}

		FontStyle() {
			return this.handle.fontStyle;
		}

		SetFontStyle(value) {
			this.handle.fontStyle = value;
		}

		FontSynthesis() {
			return this.handle.fontSynthesis;
		}

		SetFontSynthesis(value) {
			this.handle.fontSynthesis = value;
		}

		FontSynthesisSmallCaps() {
			return this.handle.fontSynthesisSmallCaps;
		}

		SetFontSynthesisSmallCaps(value) {
			this.handle.fontSynthesisSmallCaps = value;
		}

		FontSynthesisStyle() {
			return this.handle.fontSynthesisStyle;
		}

		SetFontSynthesisStyle(value) {
			this.handle.fontSynthesisStyle = value;
		}

		FontSynthesisWeight() {
			return this.handle.fontSynthesisWeight;
		}

		SetFontSynthesisWeight(value) {
			this.handle.fontSynthesisWeight = value;
		}

		FontVariant() {
			return this.handle.fontVariant;
		}

		SetFontVariant(value) {
			this.handle.fontVariant = value;
		}

		FontVariantAlternates() {
			return this.handle.fontVariantAlternates;
		}

		SetFontVariantAlternates(value) {
			this.handle.fontVariantAlternates = value;
		}

		FontVariantCaps() {
			return this.handle.fontVariantCaps;
		}

		SetFontVariantCaps(value) {
			this.handle.fontVariantCaps = value;
		}

		FontVariantEastAsian() {
			return this.handle.fontVariantEastAsian;
		}

		SetFontVariantEastAsian(value) {
			this.handle.fontVariantEastAsian = value;
		}

		FontVariantLigatures() {
			return this.handle.fontVariantLigatures;
		}

		SetFontVariantLigatures(value) {
			this.handle.fontVariantLigatures = value;
		}

		FontVariantNumeric() {
			return this.handle.fontVariantNumeric;
		}

		SetFontVariantNumeric(value) {
			this.handle.fontVariantNumeric = value;
		}

		FontVariantPosition() {
			return this.handle.fontVariantPosition;
		}

		SetFontVariantPosition(value) {
			this.handle.fontVariantPosition = value;
		}

		FontVariationSettings() {
			return this.handle.fontVariationSettings;
		}

		SetFontVariationSettings(value) {
			this.handle.fontVariationSettings = value;
		}

		FontWeight() {
			return this.handle.fontWeight;
		}

		SetFontWeight(value) {
			this.handle.fontWeight = value;
		}

		ForcedColorAdjust() {
			return this.handle.forcedColorAdjust;
		}

		SetForcedColorAdjust(value) {
			this.handle.forcedColorAdjust = value;
		}

		Gap() {
			return this.handle.gap;
		}

		SetGap(value) {
			this.handle.gap = value;
		}

		Grid() {
			return this.handle.grid;
		}

		SetGrid(value) {
			this.handle.grid = value;
		}

		GridArea() {
			return this.handle.gridArea;
		}

		SetGridArea(value) {
			this.handle.gridArea = value;
		}

		GridAutoColumns() {
			return this.handle.gridAutoColumns;
		}

		SetGridAutoColumns(value) {
			this.handle.gridAutoColumns = value;
		}

		GridAutoFlow() {
			return this.handle.gridAutoFlow;
		}

		SetGridAutoFlow(value) {
			this.handle.gridAutoFlow = value;
		}

		GridAutoRows() {
			return this.handle.gridAutoRows;
		}

		SetGridAutoRows(value) {
			this.handle.gridAutoRows = value;
		}

		GridColumn() {
			return this.handle.gridColumn;
		}

		SetGridColumn(value) {
			this.handle.gridColumn = value;
		}

		GridColumnEnd() {
			return this.handle.gridColumnEnd;
		}

		SetGridColumnEnd(value) {
			this.handle.gridColumnEnd = value;
		}

		GridColumnGap() {
			return this.handle.gridColumnGap;
		}

		SetGridColumnGap(value) {
			this.handle.gridColumnGap = value;
		}

		GridColumnStart() {
			return this.handle.gridColumnStart;
		}

		SetGridColumnStart(value) {
			this.handle.gridColumnStart = value;
		}

		GridGap() {
			return this.handle.gridGap;
		}

		SetGridGap(value) {
			this.handle.gridGap = value;
		}

		GridRow() {
			return this.handle.gridRow;
		}

		SetGridRow(value) {
			this.handle.gridRow = value;
		}

		GridRowEnd() {
			return this.handle.gridRowEnd;
		}

		SetGridRowEnd(value) {
			this.handle.gridRowEnd = value;
		}

		GridRowGap() {
			return this.handle.gridRowGap;
		}

		SetGridRowGap(value) {
			this.handle.gridRowGap = value;
		}

		GridRowStart() {
			return this.handle.gridRowStart;
		}

		SetGridRowStart(value) {
			this.handle.gridRowStart = value;
		}

		GridTemplate() {
			return this.handle.gridTemplate;
		}

		SetGridTemplate(value) {
			this.handle.gridTemplate = value;
		}

		GridTemplateAreas() {
			return this.handle.gridTemplateAreas;
		}

		SetGridTemplateAreas(value) {
			this.handle.gridTemplateAreas = value;
		}

		GridTemplateColumns() {
			return this.handle.gridTemplateColumns;
		}

		SetGridTemplateColumns(value) {
			this.handle.gridTemplateColumns = value;
		}

		GridTemplateRows() {
			return this.handle.gridTemplateRows;
		}

		SetGridTemplateRows(value) {
			this.handle.gridTemplateRows = value;
		}

		Height() {
			return this.handle.height;
		}

		SetHeight(value) {
			this.handle.height = value;
		}

		HyphenateCharacter() {
			return this.handle.hyphenateCharacter;
		}

		SetHyphenateCharacter(value) {
			this.handle.hyphenateCharacter = value;
		}

		Hyphens() {
			return this.handle.hyphens;
		}

		SetHyphens(value) {
			this.handle.hyphens = value;
		}

		ImageOrientation() {
			return this.handle.imageOrientation;
		}

		SetImageOrientation(value) {
			this.handle.imageOrientation = value;
		}

		ImageRendering() {
			return this.handle.imageRendering;
		}

		SetImageRendering(value) {
			this.handle.imageRendering = value;
		}

		InlineSize() {
			return this.handle.inlineSize;
		}

		SetInlineSize(value) {
			this.handle.inlineSize = value;
		}

		Inset() {
			return this.handle.inset;
		}

		SetInset(value) {
			this.handle.inset = value;
		}

		InsetBlock() {
			return this.handle.insetBlock;
		}

		SetInsetBlock(value) {
			this.handle.insetBlock = value;
		}

		InsetBlockEnd() {
			return this.handle.insetBlockEnd;
		}

		SetInsetBlockEnd(value) {
			this.handle.insetBlockEnd = value;
		}

		InsetBlockStart() {
			return this.handle.insetBlockStart;
		}

		SetInsetBlockStart(value) {
			this.handle.insetBlockStart = value;
		}

		InsetInline() {
			return this.handle.insetInline;
		}

		SetInsetInline(value) {
			this.handle.insetInline = value;
		}

		InsetInlineEnd() {
			return this.handle.insetInlineEnd;
		}

		SetInsetInlineEnd(value) {
			this.handle.insetInlineEnd = value;
		}

		InsetInlineStart() {
			return this.handle.insetInlineStart;
		}

		SetInsetInlineStart(value) {
			this.handle.insetInlineStart = value;
		}

		Isolation() {
			return this.handle.isolation;
		}

		SetIsolation(value) {
			this.handle.isolation = value;
		}

		JustifyContent() {
			return this.handle.justifyContent;
		}

		SetJustifyContent(value) {
			this.handle.justifyContent = value;
		}

		JustifyItems() {
			return this.handle.justifyItems;
		}

		SetJustifyItems(value) {
			this.handle.justifyItems = value;
		}

		JustifySelf() {
			return this.handle.justifySelf;
		}

		SetJustifySelf(value) {
			this.handle.justifySelf = value;
		}

		Left() {
			return this.handle.left;
		}

		SetLeft(value) {
			this.handle.left = value;
		}

		Length() {
			return this.handle.length;
		}

		LetterSpacing() {
			return this.handle.letterSpacing;
		}

		SetLetterSpacing(value) {
			this.handle.letterSpacing = value;
		}

		LightingColor() {
			return this.handle.lightingColor;
		}

		SetLightingColor(value) {
			this.handle.lightingColor = value;
		}

		LineBreak() {
			return this.handle.lineBreak;
		}

		SetLineBreak(value) {
			this.handle.lineBreak = value;
		}

		LineHeight() {
			return this.handle.lineHeight;
		}

		SetLineHeight(value) {
			this.handle.lineHeight = value;
		}

		ListStyle() {
			return this.handle.listStyle;
		}

		SetListStyle(value) {
			this.handle.listStyle = value;
		}

		ListStyleImage() {
			return this.handle.listStyleImage;
		}

		SetListStyleImage(value) {
			this.handle.listStyleImage = value;
		}

		ListStylePosition() {
			return this.handle.listStylePosition;
		}

		SetListStylePosition(value) {
			this.handle.listStylePosition = value;
		}

		ListStyleType() {
			return this.handle.listStyleType;
		}

		SetListStyleType(value) {
			this.handle.listStyleType = value;
		}

		Margin() {
			return this.handle.margin;
		}

		SetMargin(value) {
			this.handle.margin = value;
		}

		MarginBlock() {
			return this.handle.marginBlock;
		}

		SetMarginBlock(value) {
			this.handle.marginBlock = value;
		}

		MarginBlockEnd() {
			return this.handle.marginBlockEnd;
		}

		SetMarginBlockEnd(value) {
			this.handle.marginBlockEnd = value;
		}

		MarginBlockStart() {
			return this.handle.marginBlockStart;
		}

		SetMarginBlockStart(value) {
			this.handle.marginBlockStart = value;
		}

		MarginBottom() {
			return this.handle.marginBottom;
		}

		SetMarginBottom(value) {
			this.handle.marginBottom = value;
		}

		MarginInline() {
			return this.handle.marginInline;
		}

		SetMarginInline(value) {
			this.handle.marginInline = value;
		}

		MarginInlineEnd() {
			return this.handle.marginInlineEnd;
		}

		SetMarginInlineEnd(value) {
			this.handle.marginInlineEnd = value;
		}

		MarginInlineStart() {
			return this.handle.marginInlineStart;
		}

		SetMarginInlineStart(value) {
			this.handle.marginInlineStart = value;
		}

		MarginLeft() {
			return this.handle.marginLeft;
		}

		SetMarginLeft(value) {
			this.handle.marginLeft = value;
		}

		MarginRight() {
			return this.handle.marginRight;
		}

		SetMarginRight(value) {
			this.handle.marginRight = value;
		}

		MarginTop() {
			return this.handle.marginTop;
		}

		SetMarginTop(value) {
			this.handle.marginTop = value;
		}

		Marker() {
			return this.handle.marker;
		}

		SetMarker(value) {
			this.handle.marker = value;
		}

		MarkerEnd() {
			return this.handle.markerEnd;
		}

		SetMarkerEnd(value) {
			this.handle.markerEnd = value;
		}

		MarkerMid() {
			return this.handle.markerMid;
		}

		SetMarkerMid(value) {
			this.handle.markerMid = value;
		}

		MarkerStart() {
			return this.handle.markerStart;
		}

		SetMarkerStart(value) {
			this.handle.markerStart = value;
		}

		Mask() {
			return this.handle.mask;
		}

		SetMask(value) {
			this.handle.mask = value;
		}

		MaskClip() {
			return this.handle.maskClip;
		}

		SetMaskClip(value) {
			this.handle.maskClip = value;
		}

		MaskComposite() {
			return this.handle.maskComposite;
		}

		SetMaskComposite(value) {
			this.handle.maskComposite = value;
		}

		MaskImage() {
			return this.handle.maskImage;
		}

		SetMaskImage(value) {
			this.handle.maskImage = value;
		}

		MaskMode() {
			return this.handle.maskMode;
		}

		SetMaskMode(value) {
			this.handle.maskMode = value;
		}

		MaskOrigin() {
			return this.handle.maskOrigin;
		}

		SetMaskOrigin(value) {
			this.handle.maskOrigin = value;
		}

		MaskPosition() {
			return this.handle.maskPosition;
		}

		SetMaskPosition(value) {
			this.handle.maskPosition = value;
		}

		MaskRepeat() {
			return this.handle.maskRepeat;
		}

		SetMaskRepeat(value) {
			this.handle.maskRepeat = value;
		}

		MaskSize() {
			return this.handle.maskSize;
		}

		SetMaskSize(value) {
			this.handle.maskSize = value;
		}

		MaskType() {
			return this.handle.maskType;
		}

		SetMaskType(value) {
			this.handle.maskType = value;
		}

		MathDepth() {
			return this.handle.mathDepth;
		}

		SetMathDepth(value) {
			this.handle.mathDepth = value;
		}

		MathStyle() {
			return this.handle.mathStyle;
		}

		SetMathStyle(value) {
			this.handle.mathStyle = value;
		}

		MaxBlockSize() {
			return this.handle.maxBlockSize;
		}

		SetMaxBlockSize(value) {
			this.handle.maxBlockSize = value;
		}

		MaxHeight() {
			return this.handle.maxHeight;
		}

		SetMaxHeight(value) {
			this.handle.maxHeight = value;
		}

		MaxInlineSize() {
			return this.handle.maxInlineSize;
		}

		SetMaxInlineSize(value) {
			this.handle.maxInlineSize = value;
		}

		MaxWidth() {
			return this.handle.maxWidth;
		}

		SetMaxWidth(value) {
			this.handle.maxWidth = value;
		}

		MinBlockSize() {
			return this.handle.minBlockSize;
		}

		SetMinBlockSize(value) {
			this.handle.minBlockSize = value;
		}

		MinHeight() {
			return this.handle.minHeight;
		}

		SetMinHeight(value) {
			this.handle.minHeight = value;
		}

		MinInlineSize() {
			return this.handle.minInlineSize;
		}

		SetMinInlineSize(value) {
			this.handle.minInlineSize = value;
		}

		MinWidth() {
			return this.handle.minWidth;
		}

		SetMinWidth(value) {
			this.handle.minWidth = value;
		}

		MixBlendMode() {
			return this.handle.mixBlendMode;
		}

		SetMixBlendMode(value) {
			this.handle.mixBlendMode = value;
		}

		ObjectFit() {
			return this.handle.objectFit;
		}

		SetObjectFit(value) {
			this.handle.objectFit = value;
		}

		ObjectPosition() {
			return this.handle.objectPosition;
		}

		SetObjectPosition(value) {
			this.handle.objectPosition = value;
		}

		Offset() {
			return this.handle.offset;
		}

		SetOffset(value) {
			this.handle.offset = value;
		}

		OffsetAnchor() {
			return this.handle.offsetAnchor;
		}

		SetOffsetAnchor(value) {
			this.handle.offsetAnchor = value;
		}

		OffsetDistance() {
			return this.handle.offsetDistance;
		}

		SetOffsetDistance(value) {
			this.handle.offsetDistance = value;
		}

		OffsetPath() {
			return this.handle.offsetPath;
		}

		SetOffsetPath(value) {
			this.handle.offsetPath = value;
		}

		OffsetPosition() {
			return this.handle.offsetPosition;
		}

		SetOffsetPosition(value) {
			this.handle.offsetPosition = value;
		}

		OffsetRotate() {
			return this.handle.offsetRotate;
		}

		SetOffsetRotate(value) {
			this.handle.offsetRotate = value;
		}

		Opacity() {
			return this.handle.opacity;
		}

		SetOpacity(value) {
			this.handle.opacity = value;
		}

		Order() {
			return this.handle.order;
		}

		SetOrder(value) {
			this.handle.order = value;
		}

		Orphans() {
			return this.handle.orphans;
		}

		SetOrphans(value) {
			this.handle.orphans = value;
		}

		Outline() {
			return this.handle.outline;
		}

		SetOutline(value) {
			this.handle.outline = value;
		}

		OutlineColor() {
			return this.handle.outlineColor;
		}

		SetOutlineColor(value) {
			this.handle.outlineColor = value;
		}

		OutlineOffset() {
			return this.handle.outlineOffset;
		}

		SetOutlineOffset(value) {
			this.handle.outlineOffset = value;
		}

		OutlineStyle() {
			return this.handle.outlineStyle;
		}

		SetOutlineStyle(value) {
			this.handle.outlineStyle = value;
		}

		OutlineWidth() {
			return this.handle.outlineWidth;
		}

		SetOutlineWidth(value) {
			this.handle.outlineWidth = value;
		}

		Overflow() {
			return this.handle.overflow;
		}

		SetOverflow(value) {
			this.handle.overflow = value;
		}

		OverflowAnchor() {
			return this.handle.overflowAnchor;
		}

		SetOverflowAnchor(value) {
			this.handle.overflowAnchor = value;
		}

		OverflowClipMargin() {
			return this.handle.overflowClipMargin;
		}

		SetOverflowClipMargin(value) {
			this.handle.overflowClipMargin = value;
		}

		OverflowWrap() {
			return this.handle.overflowWrap;
		}

		SetOverflowWrap(value) {
			this.handle.overflowWrap = value;
		}

		OverflowX() {
			return this.handle.overflowX;
		}

		SetOverflowX(value) {
			this.handle.overflowX = value;
		}

		OverflowY() {
			return this.handle.overflowY;
		}

		SetOverflowY(value) {
			this.handle.overflowY = value;
		}

		OverscrollBehavior() {
			return this.handle.overscrollBehavior;
		}

		SetOverscrollBehavior(value) {
			this.handle.overscrollBehavior = value;
		}

		OverscrollBehaviorBlock() {
			return this.handle.overscrollBehaviorBlock;
		}

		SetOverscrollBehaviorBlock(value) {
			this.handle.overscrollBehaviorBlock = value;
		}

		OverscrollBehaviorInline() {
			return this.handle.overscrollBehaviorInline;
		}

		SetOverscrollBehaviorInline(value) {
			this.handle.overscrollBehaviorInline = value;
		}

		OverscrollBehaviorX() {
			return this.handle.overscrollBehaviorX;
		}

		SetOverscrollBehaviorX(value) {
			this.handle.overscrollBehaviorX = value;
		}

		OverscrollBehaviorY() {
			return this.handle.overscrollBehaviorY;
		}

		SetOverscrollBehaviorY(value) {
			this.handle.overscrollBehaviorY = value;
		}

		Padding() {
			return this.handle.padding;
		}

		SetPadding(value) {
			this.handle.padding = value;
		}

		PaddingBlock() {
			return this.handle.paddingBlock;
		}

		SetPaddingBlock(value) {
			this.handle.paddingBlock = value;
		}

		PaddingBlockEnd() {
			return this.handle.paddingBlockEnd;
		}

		SetPaddingBlockEnd(value) {
			this.handle.paddingBlockEnd = value;
		}

		PaddingBlockStart() {
			return this.handle.paddingBlockStart;
		}

		SetPaddingBlockStart(value) {
			this.handle.paddingBlockStart = value;
		}

		PaddingBottom() {
			return this.handle.paddingBottom;
		}

		SetPaddingBottom(value) {
			this.handle.paddingBottom = value;
		}

		PaddingInline() {
			return this.handle.paddingInline;
		}

		SetPaddingInline(value) {
			this.handle.paddingInline = value;
		}

		PaddingInlineEnd() {
			return this.handle.paddingInlineEnd;
		}

		SetPaddingInlineEnd(value) {
			this.handle.paddingInlineEnd = value;
		}

		PaddingInlineStart() {
			return this.handle.paddingInlineStart;
		}

		SetPaddingInlineStart(value) {
			this.handle.paddingInlineStart = value;
		}

		PaddingLeft() {
			return this.handle.paddingLeft;
		}

		SetPaddingLeft(value) {
			this.handle.paddingLeft = value;
		}

		PaddingRight() {
			return this.handle.paddingRight;
		}

		SetPaddingRight(value) {
			this.handle.paddingRight = value;
		}

		PaddingTop() {
			return this.handle.paddingTop;
		}

		SetPaddingTop(value) {
			this.handle.paddingTop = value;
		}

		Page() {
			return this.handle.page;
		}

		SetPage(value) {
			this.handle.page = value;
		}

		PageBreakAfter() {
			return this.handle.pageBreakAfter;
		}

		SetPageBreakAfter(value) {
			this.handle.pageBreakAfter = value;
		}

		PageBreakBefore() {
			return this.handle.pageBreakBefore;
		}

		SetPageBreakBefore(value) {
			this.handle.pageBreakBefore = value;
		}

		PageBreakInside() {
			return this.handle.pageBreakInside;
		}

		SetPageBreakInside(value) {
			this.handle.pageBreakInside = value;
		}

		PaintOrder() {
			return this.handle.paintOrder;
		}

		SetPaintOrder(value) {
			this.handle.paintOrder = value;
		}

		ParentRule() {
			return dom_wrap(dom.CSSRule, this.handle.parentRule);
		}

		Perspective() {
			return this.handle.perspective;
		}

		SetPerspective(value) {
			this.handle.perspective = value;
		}

		PerspectiveOrigin() {
			return this.handle.perspectiveOrigin;
		}

		SetPerspectiveOrigin(value) {
			this.handle.perspectiveOrigin = value;
		}

		PlaceContent() {
			return this.handle.placeContent;
		}

		SetPlaceContent(value) {
			this.handle.placeContent = value;
		}

		PlaceItems() {
			return this.handle.placeItems;
		}

		SetPlaceItems(value) {
			this.handle.placeItems = value;
		}

		PlaceSelf() {
			return this.handle.placeSelf;
		}

		SetPlaceSelf(value) {
			this.handle.placeSelf = value;
		}

		PointerEvents() {
			return this.handle.pointerEvents;
		}

		SetPointerEvents(value) {
			this.handle.pointerEvents = value;
		}

		Position() {
			return this.handle.position;
		}

		SetPosition(value) {
			this.handle.position = value;
		}

		PrintColorAdjust() {
			return this.handle.printColorAdjust;
		}

		SetPrintColorAdjust(value) {
			this.handle.printColorAdjust = value;
		}

		Quotes() {
			return this.handle.quotes;
		}

		SetQuotes(value) {
			this.handle.quotes = value;
		}

		R() {
			return this.handle.r;
		}

		SetR(value) {
			this.handle.r = value;
		}

		Resize() {
			return this.handle.resize;
		}

		SetResize(value) {
			this.handle.resize = value;
		}

		Right() {
			return this.handle.right;
		}

		SetRight(value) {
			this.handle.right = value;
		}

		Rotate() {
			return this.handle.rotate;
		}

		SetRotate(value) {
			this.handle.rotate = value;
		}

		RowGap() {
			return this.handle.rowGap;
		}

		SetRowGap(value) {
			this.handle.rowGap = value;
		}

		RubyPosition() {
			return this.handle.rubyPosition;
		}

		SetRubyPosition(value) {
			this.handle.rubyPosition = value;
		}

		Rx() {
			return this.handle.rx;
		}

		SetRx(value) {
			this.handle.rx = value;
		}

		Ry() {
			return this.handle.ry;
		}

		SetRy(value) {
			this.handle.ry = value;
		}

		Scale() {
			return this.handle.scale;
		}

		SetScale(value) {
			this.handle.scale = value;
		}

		ScrollBehavior() {
			return this.handle.scrollBehavior;
		}

		SetScrollBehavior(value) {
			this.handle.scrollBehavior = value;
		}

		ScrollMargin() {
			return this.handle.scrollMargin;
		}

		SetScrollMargin(value) {
			this.handle.scrollMargin = value;
		}

		ScrollMarginBlock() {
			return this.handle.scrollMarginBlock;
		}

		SetScrollMarginBlock(value) {
			this.handle.scrollMarginBlock = value;
		}

		ScrollMarginBlockEnd() {
			return this.handle.scrollMarginBlockEnd;
		}

		SetScrollMarginBlockEnd(value) {
			this.handle.scrollMarginBlockEnd = value;
		}

		ScrollMarginBlockStart() {
			return this.handle.scrollMarginBlockStart;
		}

		SetScrollMarginBlockStart(value) {
			this.handle.scrollMarginBlockStart = value;
		}

		ScrollMarginBottom() {
			return this.handle.scrollMarginBottom;
		}

		SetScrollMarginBottom(value) {
			this.handle.scrollMarginBottom = value;
		}

		ScrollMarginInline() {
			return this.handle.scrollMarginInline;
		}

		SetScrollMarginInline(value) {
			this.handle.scrollMarginInline = value;
		}

		ScrollMarginInlineEnd() {
			return this.handle.scrollMarginInlineEnd;
		}

		SetScrollMarginInlineEnd(value) {
			this.handle.scrollMarginInlineEnd = value;
		}

		ScrollMarginInlineStart() {
			return this.handle.scrollMarginInlineStart;
		}

		SetScrollMarginInlineStart(value) {
			this.handle.scrollMarginInlineStart = value;
		}

		ScrollMarginLeft() {
			return this.handle.scrollMarginLeft;
		}

		SetScrollMarginLeft(value) {
			this.handle.scrollMarginLeft = value;
		}

		ScrollMarginRight() {
			return this.handle.scrollMarginRight;
		}

		SetScrollMarginRight(value) {
			this.handle.scrollMarginRight = value;
		}

		ScrollMarginTop() {
			return this.handle.scrollMarginTop;
		}

		SetScrollMarginTop(value) {
			this.handle.scrollMarginTop = value;
		}

		ScrollPadding() {
			return this.handle.scrollPadding;
		}

		SetScrollPadding(value) {
			this.handle.scrollPadding = value;
		}

		ScrollPaddingBlock() {
			return this.handle.scrollPaddingBlock;
		}

		SetScrollPaddingBlock(value) {
			this.handle.scrollPaddingBlock = value;
		}

		ScrollPaddingBlockEnd() {
			return this.handle.scrollPaddingBlockEnd;
		}

		SetScrollPaddingBlockEnd(value) {
			this.handle.scrollPaddingBlockEnd = value;
		}

		ScrollPaddingBlockStart() {
			return this.handle.scrollPaddingBlockStart;
		}

		SetScrollPaddingBlockStart(value) {
			this.handle.scrollPaddingBlockStart = value;
		}

		ScrollPaddingBottom() {
			return this.handle.scrollPaddingBottom;
		}

		SetScrollPaddingBottom(value) {
			this.handle.scrollPaddingBottom = value;
		}

		ScrollPaddingInline() {
			return this.handle.scrollPaddingInline;
		}

		SetScrollPaddingInline(value) {
			this.handle.scrollPaddingInline = value;
		}

		ScrollPaddingInlineEnd() {
			return this.handle.scrollPaddingInlineEnd;
		}

		SetScrollPaddingInlineEnd(value) {
			this.handle.scrollPaddingInlineEnd = value;
		}

		ScrollPaddingInlineStart() {
			return this.handle.scrollPaddingInlineStart;
		}

		SetScrollPaddingInlineStart(value) {
			this.handle.scrollPaddingInlineStart = value;
		}

		ScrollPaddingLeft() {
			return this.handle.scrollPaddingLeft;
		}

		SetScrollPaddingLeft(value) {
			this.handle.scrollPaddingLeft = value;
		}

		ScrollPaddingRight() {
			return this.handle.scrollPaddingRight;
		}

		SetScrollPaddingRight(value) {
			this.handle.scrollPaddingRight = value;
		}

		ScrollPaddingTop() {
			return this.handle.scrollPaddingTop;
		}

		SetScrollPaddingTop(value) {
			this.handle.scrollPaddingTop = value;
		}

		ScrollSnapAlign() {
			return this.handle.scrollSnapAlign;
		}

		SetScrollSnapAlign(value) {
			this.handle.scrollSnapAlign = value;
		}

		ScrollSnapStop() {
			return this.handle.scrollSnapStop;
		}

		SetScrollSnapStop(value) {
			this.handle.scrollSnapStop = value;
		}

		ScrollSnapType() {
			return this.handle.scrollSnapType;
		}

		SetScrollSnapType(value) {
			this.handle.scrollSnapType = value;
		}

		ScrollbarColor() {
			return this.handle.scrollbarColor;
		}

		SetScrollbarColor(value) {
			this.handle.scrollbarColor = value;
		}

		ScrollbarGutter() {
			return this.handle.scrollbarGutter;
		}

		SetScrollbarGutter(value) {
			this.handle.scrollbarGutter = value;
		}

		ScrollbarWidth() {
			return this.handle.scrollbarWidth;
		}

		SetScrollbarWidth(value) {
			this.handle.scrollbarWidth = value;
		}

		ShapeImageThreshold() {
			return this.handle.shapeImageThreshold;
		}

		SetShapeImageThreshold(value) {
			this.handle.shapeImageThreshold = value;
		}

		ShapeMargin() {
			return this.handle.shapeMargin;
		}

		SetShapeMargin(value) {
			this.handle.shapeMargin = value;
		}

		ShapeOutside() {
			return this.handle.shapeOutside;
		}

		SetShapeOutside(value) {
			this.handle.shapeOutside = value;
		}

		ShapeRendering() {
			return this.handle.shapeRendering;
		}

		SetShapeRendering(value) {
			this.handle.shapeRendering = value;
		}

		StopColor() {
			return this.handle.stopColor;
		}

		SetStopColor(value) {
			this.handle.stopColor = value;
		}

		StopOpacity() {
			return this.handle.stopOpacity;
		}

		SetStopOpacity(value) {
			this.handle.stopOpacity = value;
		}

		Stroke() {
			return this.handle.stroke;
		}

		SetStroke(value) {
			this.handle.stroke = value;
		}

		StrokeDasharray() {
			return this.handle.strokeDasharray;
		}

		SetStrokeDasharray(value) {
			this.handle.strokeDasharray = value;
		}

		StrokeDashoffset() {
			return this.handle.strokeDashoffset;
		}

		SetStrokeDashoffset(value) {
			this.handle.strokeDashoffset = value;
		}

		StrokeLinecap() {
			return this.handle.strokeLinecap;
		}

		SetStrokeLinecap(value) {
			this.handle.strokeLinecap = value;
		}

		StrokeLinejoin() {
			return this.handle.strokeLinejoin;
		}

		SetStrokeLinejoin(value) {
			this.handle.strokeLinejoin = value;
		}

		StrokeMiterlimit() {
			return this.handle.strokeMiterlimit;
		}

		SetStrokeMiterlimit(value) {
			this.handle.strokeMiterlimit = value;
		}

		StrokeOpacity() {
			return this.handle.strokeOpacity;
		}

		SetStrokeOpacity(value) {
			this.handle.strokeOpacity = value;
		}

		StrokeWidth() {
			return this.handle.strokeWidth;
		}

		SetStrokeWidth(value) {
			this.handle.strokeWidth = value;
		}

		TabSize() {
			return this.handle.tabSize;
		}

		SetTabSize(value) {
			this.handle.tabSize = value;
		}

		TableLayout() {
			return this.handle.tableLayout;
		}

		SetTableLayout(value) {
			this.handle.tableLayout = value;
		}

		TextAlign() {
			return this.handle.textAlign;
		}

		SetTextAlign(value) {
			this.handle.textAlign = value;
		}

		TextAlignLast() {
			return this.handle.textAlignLast;
		}

		SetTextAlignLast(value) {
			this.handle.textAlignLast = value;
		}

		TextAnchor() {
			return this.handle.textAnchor;
		}

		SetTextAnchor(value) {
			this.handle.textAnchor = value;
		}

		TextCombineUpright() {
			return this.handle.textCombineUpright;
		}

		SetTextCombineUpright(value) {
			this.handle.textCombineUpright = value;
		}

		TextDecoration() {
			return this.handle.textDecoration;
		}

		SetTextDecoration(value) {
			this.handle.textDecoration = value;
		}

		TextDecorationColor() {
			return this.handle.textDecorationColor;
		}

		SetTextDecorationColor(value) {
			this.handle.textDecorationColor = value;
		}

		TextDecorationLine() {
			return this.handle.textDecorationLine;
		}

		SetTextDecorationLine(value) {
			this.handle.textDecorationLine = value;
		}

		TextDecorationSkipInk() {
			return this.handle.textDecorationSkipInk;
		}

		SetTextDecorationSkipInk(value) {
			this.handle.textDecorationSkipInk = value;
		}

		TextDecorationStyle() {
			return this.handle.textDecorationStyle;
		}

		SetTextDecorationStyle(value) {
			this.handle.textDecorationStyle = value;
		}

		TextDecorationThickness() {
			return this.handle.textDecorationThickness;
		}

		SetTextDecorationThickness(value) {
			this.handle.textDecorationThickness = value;
		}

		TextEmphasis() {
			return this.handle.textEmphasis;
		}

		SetTextEmphasis(value) {
			this.handle.textEmphasis = value;
		}

		TextEmphasisColor() {
			return this.handle.textEmphasisColor;
		}

		SetTextEmphasisColor(value) {
			this.handle.textEmphasisColor = value;
		}

		TextEmphasisPosition() {
			return this.handle.textEmphasisPosition;
		}

		SetTextEmphasisPosition(value) {
			this.handle.textEmphasisPosition = value;
		}

		TextEmphasisStyle() {
			return this.handle.textEmphasisStyle;
		}

		SetTextEmphasisStyle(value) {
			this.handle.textEmphasisStyle = value;
		}

		TextIndent() {
			return this.handle.textIndent;
		}

		SetTextIndent(value) {
			this.handle.textIndent = value;
		}

		TextOrientation() {
			return this.handle.textOrientation;
		}

		SetTextOrientation(value) {
			this.handle.textOrientation = value;
		}

		TextOverflow() {
			return this.handle.textOverflow;
		}

		SetTextOverflow(value) {
			this.handle.textOverflow = value;
		}

		TextRendering() {
			return this.handle.textRendering;
		}

		SetTextRendering(value) {
			this.handle.textRendering = value;
		}

		TextShadow() {
			return this.handle.textShadow;
		}

		SetTextShadow(value) {
			this.handle.textShadow = value;
		}

		TextTransform() {
			return this.handle.textTransform;
		}

		SetTextTransform(value) {
			this.handle.textTransform = value;
		}

		TextUnderlineOffset() {
			return this.handle.textUnderlineOffset;
		}

		SetTextUnderlineOffset(value) {
			this.handle.textUnderlineOffset = value;
		}

		TextUnderlinePosition() {
			return this.handle.textUnderlinePosition;
		}

		SetTextUnderlinePosition(value) {
			this.handle.textUnderlinePosition = value;
		}

		TextWrap() {
			return this.handle.textWrap;
		}

		SetTextWrap(value) {
			this.handle.textWrap = value;
		}

		TextWrapMode() {
			return this.handle.textWrapMode;
		}

		SetTextWrapMode(value) {
			this.handle.textWrapMode = value;
		}

		TextWrapStyle() {
			return this.handle.textWrapStyle;
		}

		SetTextWrapStyle(value) {
			this.handle.textWrapStyle = value;
		}

		Top() {
			return this.handle.top;
		}

		SetTop(value) {
			this.handle.top = value;
		}

		TouchAction() {
			return this.handle.touchAction;
		}

		SetTouchAction(value) {
			this.handle.touchAction = value;
		}

		Transform() {
			return this.handle.transform;
		}

		SetTransform(value) {
			this.handle.transform = value;
		}

		TransformBox() {
			return this.handle.transformBox;
		}

		SetTransformBox(value) {
			this.handle.transformBox = value;
		}

		TransformOrigin() {
			return this.handle.transformOrigin;
		}

		SetTransformOrigin(value) {
			this.handle.transformOrigin = value;
		}

		TransformStyle() {
			return this.handle.transformStyle;
		}

		SetTransformStyle(value) {
			this.handle.transformStyle = value;
		}

		Transition() {
			return this.handle.transition;
		}

		SetTransition(value) {
			this.handle.transition = value;
		}

		TransitionBehavior() {
			return this.handle.transitionBehavior;
		}

		SetTransitionBehavior(value) {
			this.handle.transitionBehavior = value;
		}

		TransitionDelay() {
			return this.handle.transitionDelay;
		}

		SetTransitionDelay(value) {
			this.handle.transitionDelay = value;
		}

		TransitionDuration() {
			return this.handle.transitionDuration;
		}

		SetTransitionDuration(value) {
			this.handle.transitionDuration = value;
		}

		TransitionProperty() {
			return this.handle.transitionProperty;
		}

		SetTransitionProperty(value) {
			this.handle.transitionProperty = value;
		}

		TransitionTimingFunction() {
			return this.handle.transitionTimingFunction;
		}

		SetTransitionTimingFunction(value) {
			this.handle.transitionTimingFunction = value;
		}

		Translate() {
			return this.handle.translate;
		}

		SetTranslate(value) {
			this.handle.translate = value;
		}

		UnicodeBidi() {
			return this.handle.unicodeBidi;
		}

		SetUnicodeBidi(value) {
			this.handle.unicodeBidi = value;
		}

		UserSelect() {
			return this.handle.userSelect;
		}

		SetUserSelect(value) {
			this.handle.userSelect = value;
		}

		VectorEffect() {
			return this.handle.vectorEffect;
		}

		SetVectorEffect(value) {
			this.handle.vectorEffect = value;
		}

		VerticalAlign() {
			return this.handle.verticalAlign;
		}

		SetVerticalAlign(value) {
			this.handle.verticalAlign = value;
		}

		ViewTransitionName() {
			return this.handle.viewTransitionName;
		}

		SetViewTransitionName(value) {
			this.handle.viewTransitionName = value;
		}

		Visibility() {
			return this.handle.visibility;
		}

		SetVisibility(value) {
			this.handle.visibility = value;
		}

		WebkitAlignContent() {
			return this.handle.webkitAlignContent;
		}

		SetWebkitAlignContent(value) {
			this.handle.webkitAlignContent = value;
		}

		WebkitAlignItems() {
			return this.handle.webkitAlignItems;
		}

		SetWebkitAlignItems(value) {
			this.handle.webkitAlignItems = value;
		}

		WebkitAlignSelf() {
			return this.handle.webkitAlignSelf;
		}

		SetWebkitAlignSelf(value) {
			this.handle.webkitAlignSelf = value;
		}

		WebkitAnimation() {
			return this.handle.webkitAnimation;
		}

		SetWebkitAnimation(value) {
			this.handle.webkitAnimation = value;
		}

		WebkitAnimationDelay() {
			return this.handle.webkitAnimationDelay;
		}

		SetWebkitAnimationDelay(value) {
			this.handle.webkitAnimationDelay = value;
		}

		WebkitAnimationDirection() {
			return this.handle.webkitAnimationDirection;
		}

		SetWebkitAnimationDirection(value) {
			this.handle.webkitAnimationDirection = value;
		}

		WebkitAnimationDuration() {
			return this.handle.webkitAnimationDuration;
		}

		SetWebkitAnimationDuration(value) {
			this.handle.webkitAnimationDuration = value;
		}

		WebkitAnimationFillMode() {
			return this.handle.webkitAnimationFillMode;
		}

		SetWebkitAnimationFillMode(value) {
			this.handle.webkitAnimationFillMode = value;
		}

		WebkitAnimationIterationCount() {
			return this.handle.webkitAnimationIterationCount;
		}

		SetWebkitAnimationIterationCount(value) {
			this.handle.webkitAnimationIterationCount = value;
		}

		WebkitAnimationName() {
			return this.handle.webkitAnimationName;
		}

		SetWebkitAnimationName(value) {
			this.handle.webkitAnimationName = value;
		}

		WebkitAnimationPlayState() {
			return this.handle.webkitAnimationPlayState;
		}

		SetWebkitAnimationPlayState(value) {
			this.handle.webkitAnimationPlayState = value;
		}

		WebkitAnimationTimingFunction() {
			return this.handle.webkitAnimationTimingFunction;
		}

		SetWebkitAnimationTimingFunction(value) {
			this.handle.webkitAnimationTimingFunction = value;
		}

		WebkitAppearance() {
			return this.handle.webkitAppearance;
		}

		SetWebkitAppearance(value) {
			this.handle.webkitAppearance = value;
		}

		WebkitBackfaceVisibility() {
			return this.handle.webkitBackfaceVisibility;
		}

		SetWebkitBackfaceVisibility(value) {
			this.handle.webkitBackfaceVisibility = value;
		}

		WebkitBackgroundClip() {
			return this.handle.webkitBackgroundClip;
		}

		SetWebkitBackgroundClip(value) {
			this.handle.webkitBackgroundClip = value;
		}

		WebkitBackgroundOrigin() {
			return this.handle.webkitBackgroundOrigin;
		}

		SetWebkitBackgroundOrigin(value) {
			this.handle.webkitBackgroundOrigin = value;
		}

		WebkitBackgroundSize() {
			return this.handle.webkitBackgroundSize;
		}

		SetWebkitBackgroundSize(value) {
			this.handle.webkitBackgroundSize = value;
		}

		WebkitBorderBottomLeftRadius() {
			return this.handle.webkitBorderBottomLeftRadius;
		}

		SetWebkitBorderBottomLeftRadius(value) {
			this.handle.webkitBorderBottomLeftRadius = value;
		}

		WebkitBorderBottomRightRadius() {
			return this.handle.webkitBorderBottomRightRadius;
		}

		SetWebkitBorderBottomRightRadius(value) {
			this.handle.webkitBorderBottomRightRadius = value;
		}

		WebkitBorderRadius() {
			return this.handle.webkitBorderRadius;
		}

		SetWebkitBorderRadius(value) {
			this.handle.webkitBorderRadius = value;
		}

		WebkitBorderTopLeftRadius() {
			return this.handle.webkitBorderTopLeftRadius;
		}

		SetWebkitBorderTopLeftRadius(value) {
			this.handle.webkitBorderTopLeftRadius = value;
		}

		WebkitBorderTopRightRadius() {
			return this.handle.webkitBorderTopRightRadius;
		}

		SetWebkitBorderTopRightRadius(value) {
			this.handle.webkitBorderTopRightRadius = value;
		}

		WebkitBoxAlign() {
			return this.handle.webkitBoxAlign;
		}

		SetWebkitBoxAlign(value) {
			this.handle.webkitBoxAlign = value;
		}

		WebkitBoxFlex() {
			return this.handle.webkitBoxFlex;
		}

		SetWebkitBoxFlex(value) {
			this.handle.webkitBoxFlex = value;
		}

		WebkitBoxOrdinalGroup() {
			return this.handle.webkitBoxOrdinalGroup;
		}

		SetWebkitBoxOrdinalGroup(value) {
			this.handle.webkitBoxOrdinalGroup = value;
		}

		WebkitBoxOrient() {
			return this.handle.webkitBoxOrient;
		}

		SetWebkitBoxOrient(value) {
			this.handle.webkitBoxOrient = value;
		}

		WebkitBoxPack() {
			return this.handle.webkitBoxPack;
		}

		SetWebkitBoxPack(value) {
			this.handle.webkitBoxPack = value;
		}

		WebkitBoxShadow() {
			return this.handle.webkitBoxShadow;
		}

		SetWebkitBoxShadow(value) {
			this.handle.webkitBoxShadow = value;
		}

		WebkitBoxSizing() {
			return this.handle.webkitBoxSizing;
		}

		SetWebkitBoxSizing(value) {
			this.handle.webkitBoxSizing = value;
		}

		WebkitFilter() {
			return this.handle.webkitFilter;
		}

		SetWebkitFilter(value) {
			this.handle.webkitFilter = value;
		}

		WebkitFlex() {
			return this.handle.webkitFlex;
		}

		SetWebkitFlex(value) {
			this.handle.webkitFlex = value;
		}

		WebkitFlexBasis() {
			return this.handle.webkitFlexBasis;
		}

		SetWebkitFlexBasis(value) {
			this.handle.webkitFlexBasis = value;
		}

		WebkitFlexDirection() {
			return this.handle.webkitFlexDirection;
		}

		SetWebkitFlexDirection(value) {
			this.handle.webkitFlexDirection = value;
		}

		WebkitFlexFlow() {
			return this.handle.webkitFlexFlow;
		}

		SetWebkitFlexFlow(value) {
			this.handle.webkitFlexFlow = value;
		}

		WebkitFlexGrow() {
			return this.handle.webkitFlexGrow;
		}

		SetWebkitFlexGrow(value) {
			this.handle.webkitFlexGrow = value;
		}

		WebkitFlexShrink() {
			return this.handle.webkitFlexShrink;
		}

		SetWebkitFlexShrink(value) {
			this.handle.webkitFlexShrink = value;
		}

		WebkitFlexWrap() {
			return this.handle.webkitFlexWrap;
		}

		SetWebkitFlexWrap(value) {
			this.handle.webkitFlexWrap = value;
		}

		WebkitJustifyContent() {
			return this.handle.webkitJustifyContent;
		}

		SetWebkitJustifyContent(value) {
			this.handle.webkitJustifyContent = value;
		}

		WebkitLineClamp() {
			return this.handle.webkitLineClamp;
		}

		SetWebkitLineClamp(value) {
			this.handle.webkitLineClamp = value;
		}

		WebkitMask() {
			return this.handle.webkitMask;
		}

		SetWebkitMask(value) {
			this.handle.webkitMask = value;
		}

		WebkitMaskBoxImage() {
			return this.handle.webkitMaskBoxImage;
		}

		SetWebkitMaskBoxImage(value) {
			this.handle.webkitMaskBoxImage = value;
		}

		WebkitMaskBoxImageOutset() {
			return this.handle.webkitMaskBoxImageOutset;
		}

		SetWebkitMaskBoxImageOutset(value) {
			this.handle.webkitMaskBoxImageOutset = value;
		}

		WebkitMaskBoxImageRepeat() {
			return this.handle.webkitMaskBoxImageRepeat;
		}

		SetWebkitMaskBoxImageRepeat(value) {
			this.handle.webkitMaskBoxImageRepeat = value;
		}

		WebkitMaskBoxImageSlice() {
			return this.handle.webkitMaskBoxImageSlice;
		}

		SetWebkitMaskBoxImageSlice(value) {
			this.handle.webkitMaskBoxImageSlice = value;
		}

		WebkitMaskBoxImageSource() {
			return this.handle.webkitMaskBoxImageSource;
		}

		SetWebkitMaskBoxImageSource(value) {
			this.handle.webkitMaskBoxImageSource = value;
		}

		WebkitMaskBoxImageWidth() {
			return this.handle.webkitMaskBoxImageWidth;
		}

		SetWebkitMaskBoxImageWidth(value) {
			this.handle.webkitMaskBoxImageWidth = value;
		}

		WebkitMaskClip() {
			return this.handle.webkitMaskClip;
		}

		SetWebkitMaskClip(value) {
			this.handle.webkitMaskClip = value;
		}

		WebkitMaskComposite() {
			return this.handle.webkitMaskComposite;
		}

		SetWebkitMaskComposite(value) {
			this.handle.webkitMaskComposite = value;
		}

		WebkitMaskImage() {
			return this.handle.webkitMaskImage;
		}

		SetWebkitMaskImage(value) {
			this.handle.webkitMaskImage = value;
		}

		WebkitMaskOrigin() {
			return this.handle.webkitMaskOrigin;
		}

		SetWebkitMaskOrigin(value) {
			this.handle.webkitMaskOrigin = value;
		}

		WebkitMaskPosition() {
			return this.handle.webkitMaskPosition;
		}

		SetWebkitMaskPosition(value) {
			this.handle.webkitMaskPosition = value;
		}

		WebkitMaskRepeat() {
			return this.handle.webkitMaskRepeat;
		}

		SetWebkitMaskRepeat(value) {
			this.handle.webkitMaskRepeat = value;
		}

		WebkitMaskSize() {
			return this.handle.webkitMaskSize;
		}

		SetWebkitMaskSize(value) {
			this.handle.webkitMaskSize = value;
		}

		WebkitOrder() {
			return this.handle.webkitOrder;
		}

		SetWebkitOrder(value) {
			this.handle.webkitOrder = value;
		}

		WebkitPerspective() {
			return this.handle.webkitPerspective;
		}

		SetWebkitPerspective(value) {
			this.handle.webkitPerspective = value;
		}

		WebkitPerspectiveOrigin() {
			return this.handle.webkitPerspectiveOrigin;
		}

		SetWebkitPerspectiveOrigin(value) {
			this.handle.webkitPerspectiveOrigin = value;
		}

		WebkitTextFillColor() {
			return this.handle.webkitTextFillColor;
		}

		SetWebkitTextFillColor(value) {
			this.handle.webkitTextFillColor = value;
		}

		WebkitTextSizeAdjust() {
			return this.handle.webkitTextSizeAdjust;
		}

		SetWebkitTextSizeAdjust(value) {
			this.handle.webkitTextSizeAdjust = value;
		}

		WebkitTextStroke() {
			return this.handle.webkitTextStroke;
		}

		SetWebkitTextStroke(value) {
			this.handle.webkitTextStroke = value;
		}

		WebkitTextStrokeColor() {
			return this.handle.webkitTextStrokeColor;
		}

		SetWebkitTextStrokeColor(value) {
			this.handle.webkitTextStrokeColor = value;
		}

		WebkitTextStrokeWidth() {
			return this.handle.webkitTextStrokeWidth;
		}

		SetWebkitTextStrokeWidth(value) {
			this.handle.webkitTextStrokeWidth = value;
		}

		WebkitTransform() {
			return this.handle.webkitTransform;
		}

		SetWebkitTransform(value) {
			this.handle.webkitTransform = value;
		}

		WebkitTransformOrigin() {
			return this.handle.webkitTransformOrigin;
		}

		SetWebkitTransformOrigin(value) {
			this.handle.webkitTransformOrigin = value;
		}

		WebkitTransformStyle() {
			return this.handle.webkitTransformStyle;
		}

		SetWebkitTransformStyle(value) {
			this.handle.webkitTransformStyle = value;
		}

		WebkitTransition() {
			return this.handle.webkitTransition;
		}

		SetWebkitTransition(value) {
			this.handle.webkitTransition = value;
		}

		WebkitTransitionDelay() {
			return this.handle.webkitTransitionDelay;
		}

		SetWebkitTransitionDelay(value) {
			this.handle.webkitTransitionDelay = value;
		}

		WebkitTransitionDuration() {
			return this.handle.webkitTransitionDuration;
		}

		SetWebkitTransitionDuration(value) {
			this.handle.webkitTransitionDuration = value;
		}

		WebkitTransitionProperty() {
			return this.handle.webkitTransitionProperty;
		}

		SetWebkitTransitionProperty(value) {
			this.handle.webkitTransitionProperty = value;
		}

		WebkitTransitionTimingFunction() {
			return this.handle.webkitTransitionTimingFunction;
		}

		SetWebkitTransitionTimingFunction(value) {
			this.handle.webkitTransitionTimingFunction = value;
		}

		WebkitUserSelect() {
			return this.handle.webkitUserSelect;
		}

		SetWebkitUserSelect(value) {
			this.handle.webkitUserSelect = value;
		}

		WhiteSpace() {
			return this.handle.whiteSpace;
		}

		SetWhiteSpace(value) {
			this.handle.whiteSpace = value;
		}

		WhiteSpaceCollapse() {
			return this.handle.whiteSpaceCollapse;
		}

		SetWhiteSpaceCollapse(value) {
			this.handle.whiteSpaceCollapse = value;
		}

		Widows() {
			return this.handle.widows;
		}

		SetWidows(value) {
			this.handle.widows = value;
		}

		Width() {
			return this.handle.width;
		}

		SetWidth(value) {
			this.handle.width = value;
		}

		WillChange() {
			return this.handle.willChange;
		}

		SetWillChange(value) {
			this.handle.willChange = value;
		}

		WordBreak() {
			return this.handle.wordBreak;
		}

		SetWordBreak(value) {
			this.handle.wordBreak = value;
		}

		WordSpacing() {
			return this.handle.wordSpacing;
		}

		SetWordSpacing(value) {
			this.handle.wordSpacing = value;
		}

		WordWrap() {
			return this.handle.wordWrap;
		}

		SetWordWrap(value) {
			this.handle.wordWrap = value;
		}

		WritingMode() {
			return this.handle.writingMode;
		}

		SetWritingMode(value) {
			this.handle.writingMode = value;
		}

		X() {
			return this.handle.x;
		}

		SetX(value) {
			this.handle.x = value;
		}

		Y() {
			return this.handle.y;
		}

		SetY(value) {
			this.handle.y = value;
		}

		ZIndex() {
			return this.handle.zIndex;
		}

		SetZIndex(value) {
			this.handle.zIndex = value;
		}

		Zoom() {
			return this.handle.zoom;
		}

		SetZoom(value) {
			this.handle.zoom = value;
		}
}

class dom_C_StyleSheet extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfStyleSheet() {
			
		}

		AsCSSStyleSheet() {
			return dom_is(this, "CSSStyleSheet") ? this : null;
		}

		Disabled() {
			return this.handle.disabled;
		}

		SetDisabled(value) {
			this.handle.disabled = value;
		}

		Href() {
			return this.handle.href;
		}

		Media() {
			return this.handle.media;
		}

		OwnerNode() {
			return this.handle.ownerNode;
		}

		ParentStyleSheet() {
			return dom_wrap(dom.CSSStyleSheet, this.handle.parentStyleSheet);
		}

		Title() {
			return this.handle.title;
		}

		Type() {
			return this.handle.type;
		}
}

class dom_C_CSSStyleSheet extends dom_C_StyleSheet {

		KindOfCSSStyleSheet() {
			
		}

		KindOfStyleSheet() {
			
		}

		AsStyleSheet() {
			return this;
		}

		AddRule(...p0) {
			return this.handle.addRule(...p0.map((item) => dom_unwrapAny(item)));
		}

		DeleteRule(p0) {
			this.handle.deleteRule(p0);
		}

		InsertRule(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			return this.handle.insertRule(p0, ...p1.map((item) => item));
		}

		RemoveRule(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.removeRule(...p0.map((item) => item));
		}

		*Replace(p0) {
			const outcome = yield* dom_await(() => this.handle.replace(p0));
			
			return outcome.ok ? [dom_wrap(dom.CSSStyleSheet, outcome.value), null] : [null, outcome.error];
		}

		ReplaceSync(p0) {
			this.handle.replaceSync(p0);
		}

		CssRules() {
			return this.handle.cssRules;
		}

		OwnerRule() {
			return dom_wrap(dom.CSSRule, this.handle.ownerRule);
		}

		Rules() {
			return this.handle.rules;
		}
}

class dom_C_CSSStyleValue extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCSSStyleValue() {
			
		}

		ToString() {
			return this.handle.toString();
		}
}

class dom_C_CacheStorage extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCacheStorage() {
			
		}

		*Delete(p0) {
			const outcome = yield* dom_await(() => this.handle.delete(p0));
			
			return outcome.ok ? [outcome.value, null] : [false, outcome.error];
		}

		*Has(p0) {
			const outcome = yield* dom_await(() => this.handle.has(p0));
			
			return outcome.ok ? [outcome.value, null] : [false, outcome.error];
		}

		*Keys() {
			const outcome = yield* dom_await(() => this.handle.keys());
			
			return outcome.ok ? [Array.from(outcome.value || []), null] : [[], outcome.error];
		}

		*Match(p0, ...p1) {
			const outcome = yield* dom_await(() => this.handle.match(dom_unwrapAny(p0), ...p1.map((item) => dom_unwrapAny(item))));
			
			return outcome.ok ? [dom_wrap(dom.Response, outcome.value), null] : [null, outcome.error];
		}

		*Open(p0) {
			const outcome = yield* dom_await(() => this.handle.open(p0));
			
			return outcome.ok ? [outcome.value, null] : [null, outcome.error];
		}
}

class dom_C_CanvasCompositing extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCanvasCompositing() {
			
		}

		AsCanvasRenderingContext2D() {
			return dom_is(this, "CanvasRenderingContext2D") ? this : null;
		}

		AsOffscreenCanvasRenderingContext2D() {
			return dom_is(this, "OffscreenCanvasRenderingContext2D") ? this : null;
		}

		GlobalAlpha() {
			return this.handle.globalAlpha;
		}

		SetGlobalAlpha(value) {
			this.handle.globalAlpha = value;
		}

		GlobalCompositeOperation() {
			return this.handle.globalCompositeOperation;
		}

		SetGlobalCompositeOperation(value) {
			this.handle.globalCompositeOperation = value;
		}
}

class dom_C_CanvasDrawImage extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCanvasDrawImage() {
			
		}

		AsCanvasRenderingContext2D() {
			return dom_is(this, "CanvasRenderingContext2D") ? this : null;
		}

		AsOffscreenCanvasRenderingContext2D() {
			return dom_is(this, "OffscreenCanvasRenderingContext2D") ? this : null;
		}

		DrawImage(p0, p1, p2) {
			this.handle.drawImage(dom_unwrapAny(p0), p1, p2);
		}
}

class dom_C_CanvasDrawPath extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCanvasDrawPath() {
			
		}

		AsCanvasRenderingContext2D() {
			return dom_is(this, "CanvasRenderingContext2D") ? this : null;
		}

		AsOffscreenCanvasRenderingContext2D() {
			return dom_is(this, "OffscreenCanvasRenderingContext2D") ? this : null;
		}

		BeginPath() {
			this.handle.beginPath();
		}

		Clip(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.clip(...p0.map((item) => item));
		}

		Fill(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.fill(...p0.map((item) => item));
		}

		IsPointInPath(p0, p1, ...p2) {
			if (p2.length === 1 && Array.isArray(p2[0]) && p2[0].length === 0) {
				p2 = [];
			}
			
			return this.handle.isPointInPath(p0, p1, ...p2.map((item) => item));
		}

		IsPointInStroke(p0, p1) {
			return this.handle.isPointInStroke(p0, p1);
		}

		Stroke() {
			this.handle.stroke();
		}
}

class dom_C_CanvasFillStrokeStyles extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCanvasFillStrokeStyles() {
			
		}

		AsCanvasRenderingContext2D() {
			return dom_is(this, "CanvasRenderingContext2D") ? this : null;
		}

		AsOffscreenCanvasRenderingContext2D() {
			return dom_is(this, "OffscreenCanvasRenderingContext2D") ? this : null;
		}

		CreateConicGradient(p0, p1, p2) {
			return dom_wrap(dom.CanvasGradient, this.handle.createConicGradient(p0, p1, p2));
		}

		CreateLinearGradient(p0, p1, p2, p3) {
			return dom_wrap(dom.CanvasGradient, this.handle.createLinearGradient(p0, p1, p2, p3));
		}

		CreatePattern(p0, p1) {
			return dom_wrap(dom.CanvasPattern, this.handle.createPattern(dom_unwrapAny(p0), p1));
		}

		CreateRadialGradient(p0, p1, p2, p3, p4, p5) {
			return dom_wrap(dom.CanvasGradient, this.handle.createRadialGradient(p0, p1, p2, p3, p4, p5));
		}

		FillStyle() {
			return this.handle.fillStyle;
		}

		SetFillStyle(value) {
			this.handle.fillStyle = dom_unwrapAny(value);
		}

		StrokeStyle() {
			return this.handle.strokeStyle;
		}

		SetStrokeStyle(value) {
			this.handle.strokeStyle = dom_unwrapAny(value);
		}
}

class dom_C_CanvasFilters extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCanvasFilters() {
			
		}

		AsCanvasRenderingContext2D() {
			return dom_is(this, "CanvasRenderingContext2D") ? this : null;
		}

		AsOffscreenCanvasRenderingContext2D() {
			return dom_is(this, "OffscreenCanvasRenderingContext2D") ? this : null;
		}

		Filter() {
			return this.handle.filter;
		}

		SetFilter(value) {
			this.handle.filter = value;
		}
}

class dom_C_CanvasGradient extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCanvasGradient() {
			
		}

		AddColorStop(p0, p1) {
			this.handle.addColorStop(p0, p1);
		}
}

class dom_C_CanvasImageData extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCanvasImageData() {
			
		}

		AsCanvasRenderingContext2D() {
			return dom_is(this, "CanvasRenderingContext2D") ? this : null;
		}

		AsOffscreenCanvasRenderingContext2D() {
			return dom_is(this, "OffscreenCanvasRenderingContext2D") ? this : null;
		}

		CreateImageData(p0, p1, ...p2) {
			if (p2.length === 1 && Array.isArray(p2[0]) && p2[0].length === 0) {
				p2 = [];
			}
			
			return dom_fromImageData(this.handle.createImageData(p0, p1, ...p2.map((item) => dom_toImageDataSettings(item))));
		}

		GetImageData(p0, p1, p2, p3, ...p4) {
			if (p4.length === 1 && Array.isArray(p4[0]) && p4[0].length === 0) {
				p4 = [];
			}
			
			return dom_fromImageData(this.handle.getImageData(p0, p1, p2, p3, ...p4.map((item) => dom_toImageDataSettings(item))));
		}

		PutImageData(p0, p1, p2) {
			this.handle.putImageData(dom_toImageData(p0), p1, p2);
		}
}

class dom_C_CanvasImageSmoothing extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCanvasImageSmoothing() {
			
		}

		AsCanvasRenderingContext2D() {
			return dom_is(this, "CanvasRenderingContext2D") ? this : null;
		}

		AsOffscreenCanvasRenderingContext2D() {
			return dom_is(this, "OffscreenCanvasRenderingContext2D") ? this : null;
		}

		ImageSmoothingEnabled() {
			return this.handle.imageSmoothingEnabled;
		}

		SetImageSmoothingEnabled(value) {
			this.handle.imageSmoothingEnabled = value;
		}

		ImageSmoothingQuality() {
			return this.handle.imageSmoothingQuality;
		}

		SetImageSmoothingQuality(value) {
			this.handle.imageSmoothingQuality = value;
		}
}

class dom_C_CanvasPath extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCanvasPath() {
			
		}

		AsCanvasRenderingContext2D() {
			return dom_is(this, "CanvasRenderingContext2D") ? this : null;
		}

		AsOffscreenCanvasRenderingContext2D() {
			return dom_is(this, "OffscreenCanvasRenderingContext2D") ? this : null;
		}

		AsPath2D() {
			return dom_is(this, "Path2D") ? this : null;
		}

		Arc(p0, p1, p2, p3, p4, ...p5) {
			if (p5.length === 1 && Array.isArray(p5[0]) && p5[0].length === 0) {
				p5 = [];
			}
			
			this.handle.arc(p0, p1, p2, p3, p4, ...p5.map((item) => item));
		}

		ArcTo(p0, p1, p2, p3, p4) {
			this.handle.arcTo(p0, p1, p2, p3, p4);
		}

		BezierCurveTo(p0, p1, p2, p3, p4, p5) {
			this.handle.bezierCurveTo(p0, p1, p2, p3, p4, p5);
		}

		ClosePath() {
			this.handle.closePath();
		}

		Ellipse(p0, p1, p2, p3, p4, p5, p6, ...p7) {
			if (p7.length === 1 && Array.isArray(p7[0]) && p7[0].length === 0) {
				p7 = [];
			}
			
			this.handle.ellipse(p0, p1, p2, p3, p4, p5, p6, ...p7.map((item) => item));
		}

		LineTo(p0, p1) {
			this.handle.lineTo(p0, p1);
		}

		MoveTo(p0, p1) {
			this.handle.moveTo(p0, p1);
		}

		QuadraticCurveTo(p0, p1, p2, p3) {
			this.handle.quadraticCurveTo(p0, p1, p2, p3);
		}

		Rect(p0, p1, p2, p3) {
			this.handle.rect(p0, p1, p2, p3);
		}

		RoundRect(p0, p1, p2, p3, ...p4) {
			this.handle.roundRect(p0, p1, p2, p3, ...p4.map((item) => dom_unwrapAny(item)));
		}
}

class dom_C_CanvasPathDrawingStyles extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCanvasPathDrawingStyles() {
			
		}

		AsCanvasRenderingContext2D() {
			return dom_is(this, "CanvasRenderingContext2D") ? this : null;
		}

		AsOffscreenCanvasRenderingContext2D() {
			return dom_is(this, "OffscreenCanvasRenderingContext2D") ? this : null;
		}

		GetLineDash() {
			return Array.from(this.handle.getLineDash() || []);
		}

		SetLineDash(p0) {
			this.handle.setLineDash(Array.from(p0 || []));
		}

		LineCap() {
			return this.handle.lineCap;
		}

		SetLineCap(value) {
			this.handle.lineCap = value;
		}

		LineDashOffset() {
			return this.handle.lineDashOffset;
		}

		SetLineDashOffset(value) {
			this.handle.lineDashOffset = value;
		}

		LineJoin() {
			return this.handle.lineJoin;
		}

		SetLineJoin(value) {
			this.handle.lineJoin = value;
		}

		LineWidth() {
			return this.handle.lineWidth;
		}

		SetLineWidth(value) {
			this.handle.lineWidth = value;
		}

		MiterLimit() {
			return this.handle.miterLimit;
		}

		SetMiterLimit(value) {
			this.handle.miterLimit = value;
		}
}

class dom_C_CanvasPattern extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCanvasPattern() {
			
		}

		SetTransform(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.setTransform(...p0.map((item) => dom_toDOMMatrix2DInit(item)));
		}
}

class dom_C_CanvasRect extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCanvasRect() {
			
		}

		AsCanvasRenderingContext2D() {
			return dom_is(this, "CanvasRenderingContext2D") ? this : null;
		}

		AsOffscreenCanvasRenderingContext2D() {
			return dom_is(this, "OffscreenCanvasRenderingContext2D") ? this : null;
		}

		ClearRect(p0, p1, p2, p3) {
			this.handle.clearRect(p0, p1, p2, p3);
		}

		FillRect(p0, p1, p2, p3) {
			this.handle.fillRect(p0, p1, p2, p3);
		}

		StrokeRect(p0, p1, p2, p3) {
			this.handle.strokeRect(p0, p1, p2, p3);
		}
}

class dom_C_CanvasShadowStyles extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCanvasShadowStyles() {
			
		}

		AsCanvasRenderingContext2D() {
			return dom_is(this, "CanvasRenderingContext2D") ? this : null;
		}

		AsOffscreenCanvasRenderingContext2D() {
			return dom_is(this, "OffscreenCanvasRenderingContext2D") ? this : null;
		}

		ShadowBlur() {
			return this.handle.shadowBlur;
		}

		SetShadowBlur(value) {
			this.handle.shadowBlur = value;
		}

		ShadowColor() {
			return this.handle.shadowColor;
		}

		SetShadowColor(value) {
			this.handle.shadowColor = value;
		}

		ShadowOffsetX() {
			return this.handle.shadowOffsetX;
		}

		SetShadowOffsetX(value) {
			this.handle.shadowOffsetX = value;
		}

		ShadowOffsetY() {
			return this.handle.shadowOffsetY;
		}

		SetShadowOffsetY(value) {
			this.handle.shadowOffsetY = value;
		}
}

class dom_C_CanvasState extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCanvasState() {
			
		}

		AsCanvasRenderingContext2D() {
			return dom_is(this, "CanvasRenderingContext2D") ? this : null;
		}

		AsOffscreenCanvasRenderingContext2D() {
			return dom_is(this, "OffscreenCanvasRenderingContext2D") ? this : null;
		}

		IsContextLost() {
			return this.handle.isContextLost();
		}

		Reset() {
			this.handle.reset();
		}

		Restore() {
			this.handle.restore();
		}

		Save() {
			this.handle.save();
		}
}

class dom_C_CanvasText extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCanvasText() {
			
		}

		AsCanvasRenderingContext2D() {
			return dom_is(this, "CanvasRenderingContext2D") ? this : null;
		}

		AsOffscreenCanvasRenderingContext2D() {
			return dom_is(this, "OffscreenCanvasRenderingContext2D") ? this : null;
		}

		FillText(p0, p1, p2, ...p3) {
			if (p3.length === 1 && Array.isArray(p3[0]) && p3[0].length === 0) {
				p3 = [];
			}
			
			this.handle.fillText(p0, p1, p2, ...p3.map((item) => item));
		}

		MeasureText(p0) {
			return dom_fromTextMetrics(this.handle.measureText(p0));
		}

		StrokeText(p0, p1, p2, ...p3) {
			if (p3.length === 1 && Array.isArray(p3[0]) && p3[0].length === 0) {
				p3 = [];
			}
			
			this.handle.strokeText(p0, p1, p2, ...p3.map((item) => item));
		}
}

class dom_C_CanvasTextDrawingStyles extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCanvasTextDrawingStyles() {
			
		}

		AsCanvasRenderingContext2D() {
			return dom_is(this, "CanvasRenderingContext2D") ? this : null;
		}

		AsOffscreenCanvasRenderingContext2D() {
			return dom_is(this, "OffscreenCanvasRenderingContext2D") ? this : null;
		}

		Direction() {
			return this.handle.direction;
		}

		SetDirection(value) {
			this.handle.direction = value;
		}

		Font() {
			return this.handle.font;
		}

		SetFont(value) {
			this.handle.font = value;
		}

		FontKerning() {
			return this.handle.fontKerning;
		}

		SetFontKerning(value) {
			this.handle.fontKerning = value;
		}

		FontStretch() {
			return this.handle.fontStretch;
		}

		SetFontStretch(value) {
			this.handle.fontStretch = value;
		}

		FontVariantCaps() {
			return this.handle.fontVariantCaps;
		}

		SetFontVariantCaps(value) {
			this.handle.fontVariantCaps = value;
		}

		LetterSpacing() {
			return this.handle.letterSpacing;
		}

		SetLetterSpacing(value) {
			this.handle.letterSpacing = value;
		}

		TextAlign() {
			return this.handle.textAlign;
		}

		SetTextAlign(value) {
			this.handle.textAlign = value;
		}

		TextBaseline() {
			return this.handle.textBaseline;
		}

		SetTextBaseline(value) {
			this.handle.textBaseline = value;
		}

		TextRendering() {
			return this.handle.textRendering;
		}

		SetTextRendering(value) {
			this.handle.textRendering = value;
		}

		WordSpacing() {
			return this.handle.wordSpacing;
		}

		SetWordSpacing(value) {
			this.handle.wordSpacing = value;
		}
}

class dom_C_CanvasTransform extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCanvasTransform() {
			
		}

		AsCanvasRenderingContext2D() {
			return dom_is(this, "CanvasRenderingContext2D") ? this : null;
		}

		AsOffscreenCanvasRenderingContext2D() {
			return dom_is(this, "OffscreenCanvasRenderingContext2D") ? this : null;
		}

		GetTransform() {
			return dom_wrap(dom.DOMMatrix, this.handle.getTransform());
		}

		ResetTransform() {
			this.handle.resetTransform();
		}

		Rotate(p0) {
			this.handle.rotate(p0);
		}

		Scale(p0, p1) {
			this.handle.scale(p0, p1);
		}

		SetTransform(p0, p1, p2, p3, p4, p5) {
			this.handle.setTransform(p0, p1, p2, p3, p4, p5);
		}

		Transform(p0, p1, p2, p3, p4, p5) {
			this.handle.transform(p0, p1, p2, p3, p4, p5);
		}

		Translate(p0, p1) {
			this.handle.translate(p0, p1);
		}
}

class dom_C_CanvasUserInterface extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCanvasUserInterface() {
			
		}

		AsCanvasRenderingContext2D() {
			return dom_is(this, "CanvasRenderingContext2D") ? this : null;
		}

		DrawFocusIfNeeded(p0) {
			this.handle.drawFocusIfNeeded(dom_unwrap(p0));
		}
}

class dom_C_CanvasRenderingContext2D extends dom_C_CanvasCompositing {

		KindOfCanvasRenderingContext2D() {
			
		}

		KindOfCanvasCompositing() {
			
		}

		KindOfCanvasDrawImage() {
			
		}

		KindOfCanvasDrawPath() {
			
		}

		KindOfCanvasFillStrokeStyles() {
			
		}

		KindOfCanvasFilters() {
			
		}

		KindOfCanvasImageData() {
			
		}

		KindOfCanvasImageSmoothing() {
			
		}

		KindOfCanvasPath() {
			
		}

		KindOfCanvasPathDrawingStyles() {
			
		}

		KindOfCanvasRect() {
			
		}

		KindOfCanvasShadowStyles() {
			
		}

		KindOfCanvasState() {
			
		}

		KindOfCanvasText() {
			
		}

		KindOfCanvasTextDrawingStyles() {
			
		}

		KindOfCanvasTransform() {
			
		}

		KindOfCanvasUserInterface() {
			
		}

		AsCanvasCompositing() {
			return this;
		}

		AsCanvasDrawImage() {
			return this;
		}

		AsCanvasDrawPath() {
			return this;
		}

		AsCanvasFillStrokeStyles() {
			return this;
		}

		AsCanvasFilters() {
			return this;
		}

		AsCanvasImageData() {
			return this;
		}

		AsCanvasImageSmoothing() {
			return this;
		}

		AsCanvasPath() {
			return this;
		}

		AsCanvasPathDrawingStyles() {
			return this;
		}

		AsCanvasRect() {
			return this;
		}

		AsCanvasShadowStyles() {
			return this;
		}

		AsCanvasState() {
			return this;
		}

		AsCanvasText() {
			return this;
		}

		AsCanvasTextDrawingStyles() {
			return this;
		}

		AsCanvasTransform() {
			return this;
		}

		AsCanvasUserInterface() {
			return this;
		}

		GetContextAttributes() {
			return dom_fromCanvasRenderingContext2DSettings(this.handle.getContextAttributes());
		}

		DrawImage(p0, p1, p2) {
			this.handle.drawImage(dom_unwrapAny(p0), p1, p2);
		}

		BeginPath() {
			this.handle.beginPath();
		}

		Clip(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.clip(...p0.map((item) => item));
		}

		Fill(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.fill(...p0.map((item) => item));
		}

		IsPointInPath(p0, p1, ...p2) {
			if (p2.length === 1 && Array.isArray(p2[0]) && p2[0].length === 0) {
				p2 = [];
			}
			
			return this.handle.isPointInPath(p0, p1, ...p2.map((item) => item));
		}

		IsPointInStroke(p0, p1) {
			return this.handle.isPointInStroke(p0, p1);
		}

		Stroke() {
			this.handle.stroke();
		}

		CreateConicGradient(p0, p1, p2) {
			return dom_wrap(dom.CanvasGradient, this.handle.createConicGradient(p0, p1, p2));
		}

		CreateLinearGradient(p0, p1, p2, p3) {
			return dom_wrap(dom.CanvasGradient, this.handle.createLinearGradient(p0, p1, p2, p3));
		}

		CreatePattern(p0, p1) {
			return dom_wrap(dom.CanvasPattern, this.handle.createPattern(dom_unwrapAny(p0), p1));
		}

		CreateRadialGradient(p0, p1, p2, p3, p4, p5) {
			return dom_wrap(dom.CanvasGradient, this.handle.createRadialGradient(p0, p1, p2, p3, p4, p5));
		}

		CreateImageData(p0, p1, ...p2) {
			if (p2.length === 1 && Array.isArray(p2[0]) && p2[0].length === 0) {
				p2 = [];
			}
			
			return dom_fromImageData(this.handle.createImageData(p0, p1, ...p2.map((item) => dom_toImageDataSettings(item))));
		}

		GetImageData(p0, p1, p2, p3, ...p4) {
			if (p4.length === 1 && Array.isArray(p4[0]) && p4[0].length === 0) {
				p4 = [];
			}
			
			return dom_fromImageData(this.handle.getImageData(p0, p1, p2, p3, ...p4.map((item) => dom_toImageDataSettings(item))));
		}

		PutImageData(p0, p1, p2) {
			this.handle.putImageData(dom_toImageData(p0), p1, p2);
		}

		Arc(p0, p1, p2, p3, p4, ...p5) {
			if (p5.length === 1 && Array.isArray(p5[0]) && p5[0].length === 0) {
				p5 = [];
			}
			
			this.handle.arc(p0, p1, p2, p3, p4, ...p5.map((item) => item));
		}

		ArcTo(p0, p1, p2, p3, p4) {
			this.handle.arcTo(p0, p1, p2, p3, p4);
		}

		BezierCurveTo(p0, p1, p2, p3, p4, p5) {
			this.handle.bezierCurveTo(p0, p1, p2, p3, p4, p5);
		}

		ClosePath() {
			this.handle.closePath();
		}

		Ellipse(p0, p1, p2, p3, p4, p5, p6, ...p7) {
			if (p7.length === 1 && Array.isArray(p7[0]) && p7[0].length === 0) {
				p7 = [];
			}
			
			this.handle.ellipse(p0, p1, p2, p3, p4, p5, p6, ...p7.map((item) => item));
		}

		LineTo(p0, p1) {
			this.handle.lineTo(p0, p1);
		}

		MoveTo(p0, p1) {
			this.handle.moveTo(p0, p1);
		}

		QuadraticCurveTo(p0, p1, p2, p3) {
			this.handle.quadraticCurveTo(p0, p1, p2, p3);
		}

		Rect(p0, p1, p2, p3) {
			this.handle.rect(p0, p1, p2, p3);
		}

		RoundRect(p0, p1, p2, p3, ...p4) {
			this.handle.roundRect(p0, p1, p2, p3, ...p4.map((item) => dom_unwrapAny(item)));
		}

		GetLineDash() {
			return Array.from(this.handle.getLineDash() || []);
		}

		SetLineDash(p0) {
			this.handle.setLineDash(Array.from(p0 || []));
		}

		ClearRect(p0, p1, p2, p3) {
			this.handle.clearRect(p0, p1, p2, p3);
		}

		FillRect(p0, p1, p2, p3) {
			this.handle.fillRect(p0, p1, p2, p3);
		}

		StrokeRect(p0, p1, p2, p3) {
			this.handle.strokeRect(p0, p1, p2, p3);
		}

		IsContextLost() {
			return this.handle.isContextLost();
		}

		Reset() {
			this.handle.reset();
		}

		Restore() {
			this.handle.restore();
		}

		Save() {
			this.handle.save();
		}

		FillText(p0, p1, p2, ...p3) {
			if (p3.length === 1 && Array.isArray(p3[0]) && p3[0].length === 0) {
				p3 = [];
			}
			
			this.handle.fillText(p0, p1, p2, ...p3.map((item) => item));
		}

		MeasureText(p0) {
			return dom_fromTextMetrics(this.handle.measureText(p0));
		}

		StrokeText(p0, p1, p2, ...p3) {
			if (p3.length === 1 && Array.isArray(p3[0]) && p3[0].length === 0) {
				p3 = [];
			}
			
			this.handle.strokeText(p0, p1, p2, ...p3.map((item) => item));
		}

		GetTransform() {
			return dom_wrap(dom.DOMMatrix, this.handle.getTransform());
		}

		ResetTransform() {
			this.handle.resetTransform();
		}

		Rotate(p0) {
			this.handle.rotate(p0);
		}

		Scale(p0, p1) {
			this.handle.scale(p0, p1);
		}

		SetTransform(p0, p1, p2, p3, p4, p5) {
			this.handle.setTransform(p0, p1, p2, p3, p4, p5);
		}

		Transform(p0, p1, p2, p3, p4, p5) {
			this.handle.transform(p0, p1, p2, p3, p4, p5);
		}

		Translate(p0, p1) {
			this.handle.translate(p0, p1);
		}

		DrawFocusIfNeeded(p0) {
			this.handle.drawFocusIfNeeded(dom_unwrap(p0));
		}

		Canvas() {
			return dom_wrap(dom.HTMLCanvasElement, this.handle.canvas);
		}

		FillStyle() {
			return this.handle.fillStyle;
		}

		SetFillStyle(value) {
			this.handle.fillStyle = dom_unwrapAny(value);
		}

		StrokeStyle() {
			return this.handle.strokeStyle;
		}

		SetStrokeStyle(value) {
			this.handle.strokeStyle = dom_unwrapAny(value);
		}

		Filter() {
			return this.handle.filter;
		}

		SetFilter(value) {
			this.handle.filter = value;
		}

		ImageSmoothingEnabled() {
			return this.handle.imageSmoothingEnabled;
		}

		SetImageSmoothingEnabled(value) {
			this.handle.imageSmoothingEnabled = value;
		}

		ImageSmoothingQuality() {
			return this.handle.imageSmoothingQuality;
		}

		SetImageSmoothingQuality(value) {
			this.handle.imageSmoothingQuality = value;
		}

		LineCap() {
			return this.handle.lineCap;
		}

		SetLineCap(value) {
			this.handle.lineCap = value;
		}

		LineDashOffset() {
			return this.handle.lineDashOffset;
		}

		SetLineDashOffset(value) {
			this.handle.lineDashOffset = value;
		}

		LineJoin() {
			return this.handle.lineJoin;
		}

		SetLineJoin(value) {
			this.handle.lineJoin = value;
		}

		LineWidth() {
			return this.handle.lineWidth;
		}

		SetLineWidth(value) {
			this.handle.lineWidth = value;
		}

		MiterLimit() {
			return this.handle.miterLimit;
		}

		SetMiterLimit(value) {
			this.handle.miterLimit = value;
		}

		ShadowBlur() {
			return this.handle.shadowBlur;
		}

		SetShadowBlur(value) {
			this.handle.shadowBlur = value;
		}

		ShadowColor() {
			return this.handle.shadowColor;
		}

		SetShadowColor(value) {
			this.handle.shadowColor = value;
		}

		ShadowOffsetX() {
			return this.handle.shadowOffsetX;
		}

		SetShadowOffsetX(value) {
			this.handle.shadowOffsetX = value;
		}

		ShadowOffsetY() {
			return this.handle.shadowOffsetY;
		}

		SetShadowOffsetY(value) {
			this.handle.shadowOffsetY = value;
		}

		Direction() {
			return this.handle.direction;
		}

		SetDirection(value) {
			this.handle.direction = value;
		}

		Font() {
			return this.handle.font;
		}

		SetFont(value) {
			this.handle.font = value;
		}

		FontKerning() {
			return this.handle.fontKerning;
		}

		SetFontKerning(value) {
			this.handle.fontKerning = value;
		}

		FontStretch() {
			return this.handle.fontStretch;
		}

		SetFontStretch(value) {
			this.handle.fontStretch = value;
		}

		FontVariantCaps() {
			return this.handle.fontVariantCaps;
		}

		SetFontVariantCaps(value) {
			this.handle.fontVariantCaps = value;
		}

		LetterSpacing() {
			return this.handle.letterSpacing;
		}

		SetLetterSpacing(value) {
			this.handle.letterSpacing = value;
		}

		TextAlign() {
			return this.handle.textAlign;
		}

		SetTextAlign(value) {
			this.handle.textAlign = value;
		}

		TextBaseline() {
			return this.handle.textBaseline;
		}

		SetTextBaseline(value) {
			this.handle.textBaseline = value;
		}

		TextRendering() {
			return this.handle.textRendering;
		}

		SetTextRendering(value) {
			this.handle.textRendering = value;
		}

		WordSpacing() {
			return this.handle.wordSpacing;
		}

		SetWordSpacing(value) {
			this.handle.wordSpacing = value;
		}
}

class dom_C_CanvasRenderingContext2DType extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCanvasRenderingContext2DType() {
			
		}

		Prototype() {
			return dom_wrap(dom.CanvasRenderingContext2D, this.handle.prototype);
		}

		SetPrototype(value) {
			this.handle.prototype = dom_unwrap(value);
		}
}

class dom_C_ChannelMergerNode extends dom_C_AudioNode {

		KindOfChannelMergerNode() {
			
		}

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsAudioNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}
}

class dom_C_ChannelSplitterNode extends dom_C_AudioNode {

		KindOfChannelSplitterNode() {
			
		}

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsAudioNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}
}

class dom_C_Clipboard extends dom_C_EventTarget {

		KindOfClipboard() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		*Read() {
			const outcome = yield* dom_await(() => this.handle.read());
			
			return outcome.ok ? [Array.from(outcome.value || [], (item) => item), null] : [[], outcome.error];
		}

		*ReadText() {
			const outcome = yield* dom_await(() => this.handle.readText());
			
			return outcome.ok ? [outcome.value, null] : ["", outcome.error];
		}

		*Write(p0) {
			const outcome = yield* dom_await(() => this.handle.write(Array.from(p0 || [], (item) => dom_unwrapAny(item))));
			
			return outcome.ok ? null : outcome.error;
		}

		*WriteText(p0) {
			const outcome = yield* dom_await(() => this.handle.writeText(p0));
			
			return outcome.ok ? null : outcome.error;
		}
}

class dom_C_ClipboardEvent extends dom_C_Event {

		KindOfClipboardEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		ClipboardData() {
			return dom_wrap(dom.DataTransfer, this.handle.clipboardData);
		}
}

class dom_C_Comment extends dom_C_CharacterData {

		KindOfComment() {
			
		}

		KindOfCharacterData() {
			
		}

		KindOfNode() {
			
		}

		KindOfChildNode() {
			
		}

		KindOfNonDocumentTypeChildNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsCharacterData() {
			return this;
		}

		AsNode() {
			return this;
		}

		AsChildNode() {
			return this;
		}

		AsNonDocumentTypeChildNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}
}

class dom_C_UIEvent extends dom_C_Event {

		KindOfUIEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		AsCompositionEvent() {
			return dom_is(this, "CompositionEvent") ? this : null;
		}

		AsDragEvent() {
			return dom_is(this, "DragEvent") ? this : null;
		}

		AsFocusEvent() {
			return dom_is(this, "FocusEvent") ? this : null;
		}

		AsInputEvent() {
			return dom_is(this, "InputEvent") ? this : null;
		}

		AsKeyboardEvent() {
			return dom_is(this, "KeyboardEvent") ? this : null;
		}

		AsMouseEvent() {
			return dom_is(this, "MouseEvent") ? this : null;
		}

		AsPointerEvent() {
			return dom_is(this, "PointerEvent") ? this : null;
		}

		AsTouchEvent() {
			return dom_is(this, "TouchEvent") ? this : null;
		}

		AsWheelEvent() {
			return dom_is(this, "WheelEvent") ? this : null;
		}

		InitUIEvent(p0, ...p1) {
			this.handle.initUIEvent(p0, ...p1.map((item) => dom_unwrapAny(item)));
		}

		Detail() {
			return this.handle.detail;
		}

		View() {
			return dom_wrap(dom.Window, this.handle.view);
		}

		Which() {
			return this.handle.which;
		}
}

class dom_C_CompositionEvent extends dom_C_UIEvent {

		KindOfCompositionEvent() {
			
		}

		KindOfUIEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsUIEvent() {
			return this;
		}

		AsEvent() {
			return this;
		}

		InitCompositionEvent(p0, ...p1) {
			this.handle.initCompositionEvent(p0, ...p1.map((item) => dom_unwrapAny(item)));
		}

		Data() {
			return this.handle.data;
		}
}

class dom_C_ConstantSourceNode extends dom_C_AudioScheduledSourceNode {

		KindOfConstantSourceNode() {
			
		}

		KindOfAudioScheduledSourceNode() {
			
		}

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsAudioScheduledSourceNode() {
			return this;
		}

		AsAudioNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		Offset() {
			return dom_wrap(dom.AudioParam, this.handle.offset);
		}
}

class dom_C_ConvolverNode extends dom_C_AudioNode {

		KindOfConvolverNode() {
			
		}

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsAudioNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		Buffer() {
			return dom_wrap(dom.AudioBuffer, this.handle.buffer);
		}

		SetBuffer(value) {
			this.handle.buffer = dom_unwrap(value);
		}

		Normalize() {
			return this.handle.normalize;
		}

		SetNormalize(value) {
			this.handle.normalize = value;
		}
}

class dom_C_CredentialsContainer extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCredentialsContainer() {
			
		}

		*Create(...p0) {
			const outcome = yield* dom_await(() => this.handle.create(...p0.map((item) => dom_unwrapAny(item))));
			
			return outcome.ok ? [outcome.value, null] : [null, outcome.error];
		}

		*Get(...p0) {
			const outcome = yield* dom_await(() => this.handle.get(...p0.map((item) => dom_unwrapAny(item))));
			
			return outcome.ok ? [outcome.value, null] : [null, outcome.error];
		}

		*PreventSilentAccess() {
			const outcome = yield* dom_await(() => this.handle.preventSilentAccess());
			
			return outcome.ok ? null : outcome.error;
		}

		*Store(p0) {
			const outcome = yield* dom_await(() => this.handle.store(dom_unwrapAny(p0)));
			
			return outcome.ok ? null : outcome.error;
		}
}

class dom_C_Crypto extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCrypto() {
			
		}

		GetRandomValues(p0) {
			return this.handle.getRandomValues(dom_unwrapAny(p0));
		}

		RandomUUID() {
			return this.handle.randomUUID();
		}

		Subtle() {
			return this.handle.subtle;
		}
}

class dom_C_CustomElementRegistry extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCustomElementRegistry() {
			
		}

		Define(p0, p1, ...p2) {
			this.handle.define(p0, dom_unwrapAny(p1), ...p2.map((item) => dom_unwrapAny(item)));
		}

		Get(p0) {
			return this.handle.get(p0);
		}

		GetName(p0) {
			return this.handle.getName(dom_unwrapAny(p0));
		}

		Upgrade(p0) {
			this.handle.upgrade(dom_unwrap(p0));
		}

		*WhenDefined(p0) {
			const outcome = yield* dom_await(() => this.handle.whenDefined(p0));
			
			return outcome.ok ? [outcome.value, null] : [null, outcome.error];
		}
}

class dom_C_CustomStateSet extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfCustomStateSet() {
			
		}

		ForEach(p0, ...p1) {
			this.handle.forEach((p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [a[0], a[1], dom_wrap(dom.CustomStateSet, a[2])])), ...p1.map((item) => dom_unwrapAny(item)));
		}
}

class dom_C_DOMImplementation extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfDOMImplementation() {
			
		}

		CreateDocument(p0, p1, ...p2) {
			if (p2.length === 1 && Array.isArray(p2[0]) && p2[0].length === 0) {
				p2 = [];
			}
			
			return this.handle.createDocument(p0, p1, ...p2.map((item) => dom_unwrap(item)));
		}

		CreateDocumentType(p0, p1, p2) {
			return dom_wrap(dom.DocumentType, this.handle.createDocumentType(p0, p1, p2));
		}

		CreateHTMLDocument(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.Document, this.handle.createHTMLDocument(...p0.map((item) => item)));
		}

		HasFeature(...p0) {
			return this.handle.hasFeature(...p0.map((item) => dom_unwrapAny(item)));
		}
}

class dom_C_DOMMatrixReadOnly extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfDOMMatrixReadOnly() {
			
		}

		AsDOMMatrix() {
			return dom_is(this, "DOMMatrix") ? this : null;
		}

		FlipX() {
			return dom_wrap(dom.DOMMatrix, this.handle.flipX());
		}

		FlipY() {
			return dom_wrap(dom.DOMMatrix, this.handle.flipY());
		}

		Inverse() {
			return dom_wrap(dom.DOMMatrix, this.handle.inverse());
		}

		Multiply(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMMatrix, this.handle.multiply(...p0.map((item) => dom_toDOMMatrixInit(item))));
		}

		Rotate(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMMatrix, this.handle.rotate(...p0.map((item) => item)));
		}

		RotateAxisAngle(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMMatrix, this.handle.rotateAxisAngle(...p0.map((item) => item)));
		}

		RotateFromVector(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMMatrix, this.handle.rotateFromVector(...p0.map((item) => item)));
		}

		Scale(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMMatrix, this.handle.scale(...p0.map((item) => item)));
		}

		Scale3d(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMMatrix, this.handle.scale3d(...p0.map((item) => item)));
		}

		ScaleNonUniform(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMMatrix, this.handle.scaleNonUniform(...p0.map((item) => item)));
		}

		SkewX(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMMatrix, this.handle.skewX(...p0.map((item) => item)));
		}

		SkewY(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMMatrix, this.handle.skewY(...p0.map((item) => item)));
		}

		ToFloat32Array() {
			return Array.from(this.handle.toFloat32Array() || []);
		}

		ToFloat64Array() {
			return Array.from(this.handle.toFloat64Array() || []);
		}

		ToJSON() {
			return this.handle.toJSON();
		}

		TransformPoint(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMPoint, this.handle.transformPoint(...p0.map((item) => dom_toDOMPointInit(item))));
		}

		Translate(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMMatrix, this.handle.translate(...p0.map((item) => item)));
		}

		ToString() {
			return this.handle.toString();
		}

		A() {
			return this.handle.a;
		}

		B() {
			return this.handle.b;
		}

		C() {
			return this.handle.c;
		}

		D() {
			return this.handle.d;
		}

		E() {
			return this.handle.e;
		}

		F() {
			return this.handle.f;
		}

		Is2D() {
			return this.handle.is2D;
		}

		IsIdentity() {
			return this.handle.isIdentity;
		}

		M11() {
			return this.handle.m11;
		}

		M12() {
			return this.handle.m12;
		}

		M13() {
			return this.handle.m13;
		}

		M14() {
			return this.handle.m14;
		}

		M21() {
			return this.handle.m21;
		}

		M22() {
			return this.handle.m22;
		}

		M23() {
			return this.handle.m23;
		}

		M24() {
			return this.handle.m24;
		}

		M31() {
			return this.handle.m31;
		}

		M32() {
			return this.handle.m32;
		}

		M33() {
			return this.handle.m33;
		}

		M34() {
			return this.handle.m34;
		}

		M41() {
			return this.handle.m41;
		}

		M42() {
			return this.handle.m42;
		}

		M43() {
			return this.handle.m43;
		}

		M44() {
			return this.handle.m44;
		}
}

class dom_C_DOMMatrix extends dom_C_DOMMatrixReadOnly {

		KindOfDOMMatrix() {
			
		}

		KindOfDOMMatrixReadOnly() {
			
		}

		AsDOMMatrixReadOnly() {
			return this;
		}

		InvertSelf() {
			return dom_wrap(dom.DOMMatrix, this.handle.invertSelf());
		}

		MultiplySelf(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMMatrix, this.handle.multiplySelf(...p0.map((item) => dom_toDOMMatrixInit(item))));
		}

		PreMultiplySelf(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMMatrix, this.handle.preMultiplySelf(...p0.map((item) => dom_toDOMMatrixInit(item))));
		}

		RotateAxisAngleSelf(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMMatrix, this.handle.rotateAxisAngleSelf(...p0.map((item) => item)));
		}

		RotateFromVectorSelf(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMMatrix, this.handle.rotateFromVectorSelf(...p0.map((item) => item)));
		}

		RotateSelf(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMMatrix, this.handle.rotateSelf(...p0.map((item) => item)));
		}

		Scale3dSelf(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMMatrix, this.handle.scale3dSelf(...p0.map((item) => item)));
		}

		ScaleSelf(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMMatrix, this.handle.scaleSelf(...p0.map((item) => item)));
		}

		SetMatrixValue(p0) {
			return dom_wrap(dom.DOMMatrix, this.handle.setMatrixValue(p0));
		}

		SkewXSelf(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMMatrix, this.handle.skewXSelf(...p0.map((item) => item)));
		}

		SkewYSelf(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMMatrix, this.handle.skewYSelf(...p0.map((item) => item)));
		}

		TranslateSelf(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMMatrix, this.handle.translateSelf(...p0.map((item) => item)));
		}

		SetA(value) {
			this.handle.a = value;
		}

		SetB(value) {
			this.handle.b = value;
		}

		SetC(value) {
			this.handle.c = value;
		}

		SetD(value) {
			this.handle.d = value;
		}

		SetE(value) {
			this.handle.e = value;
		}

		SetF(value) {
			this.handle.f = value;
		}

		SetM11(value) {
			this.handle.m11 = value;
		}

		SetM12(value) {
			this.handle.m12 = value;
		}

		SetM13(value) {
			this.handle.m13 = value;
		}

		SetM14(value) {
			this.handle.m14 = value;
		}

		SetM21(value) {
			this.handle.m21 = value;
		}

		SetM22(value) {
			this.handle.m22 = value;
		}

		SetM23(value) {
			this.handle.m23 = value;
		}

		SetM24(value) {
			this.handle.m24 = value;
		}

		SetM31(value) {
			this.handle.m31 = value;
		}

		SetM32(value) {
			this.handle.m32 = value;
		}

		SetM33(value) {
			this.handle.m33 = value;
		}

		SetM34(value) {
			this.handle.m34 = value;
		}

		SetM41(value) {
			this.handle.m41 = value;
		}

		SetM42(value) {
			this.handle.m42 = value;
		}

		SetM43(value) {
			this.handle.m43 = value;
		}

		SetM44(value) {
			this.handle.m44 = value;
		}
}

class dom_C_DOMPointReadOnly extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfDOMPointReadOnly() {
			
		}

		AsDOMPoint() {
			return dom_is(this, "DOMPoint") ? this : null;
		}

		MatrixTransform(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.DOMPoint, this.handle.matrixTransform(...p0.map((item) => dom_toDOMMatrixInit(item))));
		}

		ToJSON() {
			return this.handle.toJSON();
		}

		W() {
			return this.handle.w;
		}

		X() {
			return this.handle.x;
		}

		Y() {
			return this.handle.y;
		}

		Z() {
			return this.handle.z;
		}
}

class dom_C_DOMPoint extends dom_C_DOMPointReadOnly {

		KindOfDOMPoint() {
			
		}

		KindOfDOMPointReadOnly() {
			
		}

		AsDOMPointReadOnly() {
			return this;
		}

		SetW(value) {
			this.handle.w = value;
		}

		SetX(value) {
			this.handle.x = value;
		}

		SetY(value) {
			this.handle.y = value;
		}

		SetZ(value) {
			this.handle.z = value;
		}
}

class dom_C_DOMRectReadOnly extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfDOMRectReadOnly() {
			
		}

		AsDOMRect() {
			return dom_is(this, "DOMRect") ? this : null;
		}

		ToJSON() {
			return this.handle.toJSON();
		}

		Bottom() {
			return this.handle.bottom;
		}

		Height() {
			return this.handle.height;
		}

		Left() {
			return this.handle.left;
		}

		Right() {
			return this.handle.right;
		}

		Top() {
			return this.handle.top;
		}

		Width() {
			return this.handle.width;
		}

		X() {
			return this.handle.x;
		}

		Y() {
			return this.handle.y;
		}
}

class dom_C_DOMRect extends dom_C_DOMRectReadOnly {

		KindOfDOMRect() {
			
		}

		KindOfDOMRectReadOnly() {
			
		}

		AsDOMRectReadOnly() {
			return this;
		}

		SetHeight(value) {
			this.handle.height = value;
		}

		SetWidth(value) {
			this.handle.width = value;
		}

		SetX(value) {
			this.handle.x = value;
		}

		SetY(value) {
			this.handle.y = value;
		}
}

class dom_C_DOMRectList extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfDOMRectList() {
			
		}

		Item(p0) {
			return dom_wrap(dom.DOMRect, this.handle.item(p0));
		}

		Length() {
			return this.handle.length;
		}
}

class dom_C_DOMStringList extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfDOMStringList() {
			
		}

		Contains(p0) {
			return this.handle.contains(p0);
		}

		Item(p0) {
			return this.handle.item(p0);
		}

		Length() {
			return this.handle.length;
		}
}

class dom_C_DOMTokenList extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfDOMTokenList() {
			
		}

		ToString() {
			return this.handle.toString();
		}

		Add(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.add(...p0.map((item) => item));
		}

		Contains(p0) {
			return this.handle.contains(p0);
		}

		Item(p0) {
			return this.handle.item(p0);
		}

		Remove(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.remove(...p0.map((item) => item));
		}

		Replace(p0, p1) {
			return this.handle.replace(p0, p1);
		}

		Supports(p0) {
			return this.handle.supports(p0);
		}

		Toggle(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			return this.handle.toggle(p0, ...p1.map((item) => item));
		}

		ForEach(p0, ...p1) {
			this.handle.forEach((p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [a[0], a[1], dom_wrap(dom.DOMTokenList, a[2])])), ...p1.map((item) => dom_unwrapAny(item)));
		}

		Length() {
			return this.handle.length;
		}

		Value() {
			return this.handle.value;
		}

		SetValue(value) {
			this.handle.value = value;
		}
}

class dom_C_DataTransfer extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfDataTransfer() {
			
		}

		ClearData(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.clearData(...p0.map((item) => item));
		}

		GetData(p0) {
			return this.handle.getData(p0);
		}

		SetData(p0, p1) {
			this.handle.setData(p0, p1);
		}

		SetDragImage(p0, p1, p2) {
			this.handle.setDragImage(dom_unwrap(p0), p1, p2);
		}

		DropEffect() {
			return this.handle.dropEffect;
		}

		SetDropEffect(value) {
			this.handle.dropEffect = value;
		}

		EffectAllowed() {
			return this.handle.effectAllowed;
		}

		SetEffectAllowed(value) {
			this.handle.effectAllowed = value;
		}

		Files() {
			return this.handle.files;
		}

		Items() {
			return this.handle.items;
		}

		Types() {
			return Array.from(this.handle.types || []);
		}
}

class dom_C_DelayNode extends dom_C_AudioNode {

		KindOfDelayNode() {
			
		}

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsAudioNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		DelayTime() {
			return dom_wrap(dom.AudioParam, this.handle.delayTime);
		}
}

class dom_C_DeviceMotionEvent extends dom_C_Event {

		KindOfDeviceMotionEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		Acceleration() {
			return this.handle.acceleration;
		}

		AccelerationIncludingGravity() {
			return this.handle.accelerationIncludingGravity;
		}

		Interval() {
			return this.handle.interval;
		}

		RotationRate() {
			return this.handle.rotationRate;
		}
}

class dom_C_DeviceOrientationEvent extends dom_C_Event {

		KindOfDeviceOrientationEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		Absolute() {
			return this.handle.absolute;
		}

		Alpha() {
			return this.handle.alpha;
		}

		Beta() {
			return this.handle.beta;
		}

		Gamma() {
			return this.handle.gamma;
		}
}

class dom_C_DocumentOrShadowRoot extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfDocumentOrShadowRoot() {
			
		}

		AsDocument() {
			return dom_is(this, "Document") ? this : null;
		}

		AsShadowRoot() {
			return dom_is(this, "ShadowRoot") ? this : null;
		}

		ElementFromPoint(p0, p1) {
			return dom_wrap(dom.Element, this.handle.elementFromPoint(p0, p1));
		}

		ElementsFromPoint(p0, p1) {
			return Array.from(this.handle.elementsFromPoint(p0, p1) || [], (item) => dom_wrap(dom.Element, item));
		}

		GetAnimations() {
			return Array.from(this.handle.getAnimations() || [], (item) => dom_wrap(dom.Animation, item));
		}

		ActiveElement() {
			return dom_wrap(dom.Element, this.handle.activeElement);
		}

		AdoptedStyleSheets() {
			return Array.from(this.handle.adoptedStyleSheets || [], (item) => dom_wrap(dom.CSSStyleSheet, item));
		}

		SetAdoptedStyleSheets(value) {
			this.handle.adoptedStyleSheets = Array.from(value || [], (item) => dom_unwrap(item));
		}

		FullscreenElement() {
			return dom_wrap(dom.Element, this.handle.fullscreenElement);
		}

		PictureInPictureElement() {
			return dom_wrap(dom.Element, this.handle.pictureInPictureElement);
		}

		PointerLockElement() {
			return dom_wrap(dom.Element, this.handle.pointerLockElement);
		}

		StyleSheets() {
			return dom_wrap(dom.StyleSheetList, this.handle.styleSheets);
		}
}

class dom_C_FontFaceSource extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfFontFaceSource() {
			
		}

		AsDocument() {
			return dom_is(this, "Document") ? this : null;
		}

		Fonts() {
			return dom_wrap(dom.FontFaceSet, this.handle.fonts);
		}
}

class dom_C_GlobalEventHandlers extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfGlobalEventHandlers() {
			
		}

		AsDocument() {
			return dom_is(this, "Document") ? this : null;
		}

		AsHTMLAudioElement() {
			return dom_is(this, "HTMLAudioElement") ? this : null;
		}

		AsHTMLCanvasElement() {
			return dom_is(this, "HTMLCanvasElement") ? this : null;
		}

		AsHTMLElement() {
			return dom_is(this, "HTMLElement") ? this : null;
		}

		AsHTMLFormElement() {
			return dom_is(this, "HTMLFormElement") ? this : null;
		}

		AsHTMLHeadElement() {
			return dom_is(this, "HTMLHeadElement") ? this : null;
		}

		AsHTMLImageElement() {
			return dom_is(this, "HTMLImageElement") ? this : null;
		}

		AsHTMLMediaElement() {
			return dom_is(this, "HTMLMediaElement") ? this : null;
		}

		AsHTMLSlotElement() {
			return dom_is(this, "HTMLSlotElement") ? this : null;
		}

		AsSVGElement() {
			return dom_is(this, "SVGElement") ? this : null;
		}

		AsSVGGraphicsElement() {
			return dom_is(this, "SVGGraphicsElement") ? this : null;
		}

		AsSVGSVGElement() {
			return dom_is(this, "SVGSVGElement") ? this : null;
		}

		AsWindow() {
			return dom_is(this, "Window") ? this : null;
		}

		AddEventListener(p0, p1, ...p2) {
			this.handle.addEventListener(p0, (p1 === null || p1 === undefined ? null : (...a) => go2jsCallNow(p1, null, [dom_wrap(dom.Event, a[0])])), ...p2.map((item) => dom_unwrapAny(item)));
		}

		RemoveEventListener(p0, p1, ...p2) {
			this.handle.removeEventListener(p0, (p1 === null || p1 === undefined ? null : (...a) => go2jsCallNow(p1, null, [dom_wrap(dom.Event, a[0])])), ...p2.map((item) => dom_unwrapAny(item)));
		}

		OnAbort(p0) {
			this.handle.addEventListener("abort", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.UIEvent, a[0])])));
		}

		OnAnimationcancel(p0) {
			this.handle.addEventListener("animationcancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationEvent, a[0])])));
		}

		OnAnimationend(p0) {
			this.handle.addEventListener("animationend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationEvent, a[0])])));
		}

		OnAnimationiteration(p0) {
			this.handle.addEventListener("animationiteration", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationEvent, a[0])])));
		}

		OnAnimationstart(p0) {
			this.handle.addEventListener("animationstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationEvent, a[0])])));
		}

		OnAuxclick(p0) {
			this.handle.addEventListener("auxclick", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnBeforeinput(p0) {
			this.handle.addEventListener("beforeinput", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.InputEvent, a[0])])));
		}

		OnBeforetoggle(p0) {
			this.handle.addEventListener("beforetoggle", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnBlur(p0) {
			this.handle.addEventListener("blur", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FocusEvent, a[0])])));
		}

		OnCancel(p0) {
			this.handle.addEventListener("cancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCanplay(p0) {
			this.handle.addEventListener("canplay", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCanplaythrough(p0) {
			this.handle.addEventListener("canplaythrough", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnChange(p0) {
			this.handle.addEventListener("change", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnClick(p0) {
			this.handle.addEventListener("click", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnClose(p0) {
			this.handle.addEventListener("close", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCompositionend(p0) {
			this.handle.addEventListener("compositionend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.CompositionEvent, a[0])])));
		}

		OnCompositionstart(p0) {
			this.handle.addEventListener("compositionstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.CompositionEvent, a[0])])));
		}

		OnCompositionupdate(p0) {
			this.handle.addEventListener("compositionupdate", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.CompositionEvent, a[0])])));
		}

		OnContextlost(p0) {
			this.handle.addEventListener("contextlost", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnContextmenu(p0) {
			this.handle.addEventListener("contextmenu", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnContextrestored(p0) {
			this.handle.addEventListener("contextrestored", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCopy(p0) {
			this.handle.addEventListener("copy", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ClipboardEvent, a[0])])));
		}

		OnCuechange(p0) {
			this.handle.addEventListener("cuechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCut(p0) {
			this.handle.addEventListener("cut", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ClipboardEvent, a[0])])));
		}

		OnDblclick(p0) {
			this.handle.addEventListener("dblclick", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnDrag(p0) {
			this.handle.addEventListener("drag", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragend(p0) {
			this.handle.addEventListener("dragend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragenter(p0) {
			this.handle.addEventListener("dragenter", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragleave(p0) {
			this.handle.addEventListener("dragleave", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragover(p0) {
			this.handle.addEventListener("dragover", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragstart(p0) {
			this.handle.addEventListener("dragstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDrop(p0) {
			this.handle.addEventListener("drop", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDurationchange(p0) {
			this.handle.addEventListener("durationchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnEmptied(p0) {
			this.handle.addEventListener("emptied", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnEnded(p0) {
			this.handle.addEventListener("ended", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnError(p0) {
			this.handle.addEventListener("error", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ErrorEvent, a[0])])));
		}

		OnFocus(p0) {
			this.handle.addEventListener("focus", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FocusEvent, a[0])])));
		}

		OnFocusin(p0) {
			this.handle.addEventListener("focusin", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FocusEvent, a[0])])));
		}

		OnFocusout(p0) {
			this.handle.addEventListener("focusout", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FocusEvent, a[0])])));
		}

		OnFormdata(p0) {
			this.handle.addEventListener("formdata", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FormDataEvent, a[0])])));
		}

		OnGotpointercapture(p0) {
			this.handle.addEventListener("gotpointercapture", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnInput(p0) {
			this.handle.addEventListener("input", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnInvalid(p0) {
			this.handle.addEventListener("invalid", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnKeydown(p0) {
			this.handle.addEventListener("keydown", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.KeyboardEvent, a[0])])));
		}

		OnKeypress(p0) {
			this.handle.addEventListener("keypress", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.KeyboardEvent, a[0])])));
		}

		OnKeyup(p0) {
			this.handle.addEventListener("keyup", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.KeyboardEvent, a[0])])));
		}

		OnLoad(p0) {
			this.handle.addEventListener("load", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLoadeddata(p0) {
			this.handle.addEventListener("loadeddata", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLoadedmetadata(p0) {
			this.handle.addEventListener("loadedmetadata", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLoadstart(p0) {
			this.handle.addEventListener("loadstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLostpointercapture(p0) {
			this.handle.addEventListener("lostpointercapture", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnMousedown(p0) {
			this.handle.addEventListener("mousedown", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseenter(p0) {
			this.handle.addEventListener("mouseenter", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseleave(p0) {
			this.handle.addEventListener("mouseleave", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMousemove(p0) {
			this.handle.addEventListener("mousemove", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseout(p0) {
			this.handle.addEventListener("mouseout", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseover(p0) {
			this.handle.addEventListener("mouseover", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseup(p0) {
			this.handle.addEventListener("mouseup", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnPaste(p0) {
			this.handle.addEventListener("paste", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ClipboardEvent, a[0])])));
		}

		OnPause(p0) {
			this.handle.addEventListener("pause", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnPlay(p0) {
			this.handle.addEventListener("play", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnPlaying(p0) {
			this.handle.addEventListener("playing", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnPointercancel(p0) {
			this.handle.addEventListener("pointercancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerdown(p0) {
			this.handle.addEventListener("pointerdown", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerenter(p0) {
			this.handle.addEventListener("pointerenter", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerleave(p0) {
			this.handle.addEventListener("pointerleave", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointermove(p0) {
			this.handle.addEventListener("pointermove", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerout(p0) {
			this.handle.addEventListener("pointerout", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerover(p0) {
			this.handle.addEventListener("pointerover", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerup(p0) {
			this.handle.addEventListener("pointerup", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnProgress(p0) {
			this.handle.addEventListener("progress", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ProgressEvent, a[0])])));
		}

		OnRatechange(p0) {
			this.handle.addEventListener("ratechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnReset(p0) {
			this.handle.addEventListener("reset", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnResize(p0) {
			this.handle.addEventListener("resize", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.UIEvent, a[0])])));
		}

		OnScroll(p0) {
			this.handle.addEventListener("scroll", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnScrollend(p0) {
			this.handle.addEventListener("scrollend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSecuritypolicyviolation(p0) {
			this.handle.addEventListener("securitypolicyviolation", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.SecurityPolicyViolationEvent, a[0])])));
		}

		OnSeeked(p0) {
			this.handle.addEventListener("seeked", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSeeking(p0) {
			this.handle.addEventListener("seeking", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSelect(p0) {
			this.handle.addEventListener("select", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSelectionchange(p0) {
			this.handle.addEventListener("selectionchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSelectstart(p0) {
			this.handle.addEventListener("selectstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSlotchange(p0) {
			this.handle.addEventListener("slotchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnStalled(p0) {
			this.handle.addEventListener("stalled", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSubmit(p0) {
			this.handle.addEventListener("submit", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.SubmitEvent, a[0])])));
		}

		OnSuspend(p0) {
			this.handle.addEventListener("suspend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnTimeupdate(p0) {
			this.handle.addEventListener("timeupdate", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnToggle(p0) {
			this.handle.addEventListener("toggle", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnTouchcancel(p0) {
			this.handle.addEventListener("touchcancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TouchEvent, a[0])])));
		}

		OnTouchend(p0) {
			this.handle.addEventListener("touchend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TouchEvent, a[0])])));
		}

		OnTouchmove(p0) {
			this.handle.addEventListener("touchmove", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TouchEvent, a[0])])));
		}

		OnTouchstart(p0) {
			this.handle.addEventListener("touchstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TouchEvent, a[0])])));
		}

		OnTransitioncancel(p0) {
			this.handle.addEventListener("transitioncancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TransitionEvent, a[0])])));
		}

		OnTransitionend(p0) {
			this.handle.addEventListener("transitionend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TransitionEvent, a[0])])));
		}

		OnTransitionrun(p0) {
			this.handle.addEventListener("transitionrun", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TransitionEvent, a[0])])));
		}

		OnTransitionstart(p0) {
			this.handle.addEventListener("transitionstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TransitionEvent, a[0])])));
		}

		OnVolumechange(p0) {
			this.handle.addEventListener("volumechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWaiting(p0) {
			this.handle.addEventListener("waiting", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWebkitanimationend(p0) {
			this.handle.addEventListener("webkitanimationend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWebkitanimationiteration(p0) {
			this.handle.addEventListener("webkitanimationiteration", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWebkitanimationstart(p0) {
			this.handle.addEventListener("webkitanimationstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWebkittransitionend(p0) {
			this.handle.addEventListener("webkittransitionend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWheel(p0) {
			this.handle.addEventListener("wheel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.WheelEvent, a[0])])));
		}

		Onabort(p0) {
			return this.handle.onabort(dom_unwrap(p0));
		}

		Onanimationcancel(p0) {
			return this.handle.onanimationcancel(dom_unwrap(p0));
		}

		Onanimationend(p0) {
			return this.handle.onanimationend(dom_unwrap(p0));
		}

		Onanimationiteration(p0) {
			return this.handle.onanimationiteration(dom_unwrap(p0));
		}

		Onanimationstart(p0) {
			return this.handle.onanimationstart(dom_unwrap(p0));
		}

		Onauxclick(p0) {
			return this.handle.onauxclick(dom_unwrap(p0));
		}

		Onbeforeinput(p0) {
			return this.handle.onbeforeinput(dom_unwrap(p0));
		}

		Onbeforetoggle(p0) {
			return this.handle.onbeforetoggle(dom_unwrap(p0));
		}

		Onblur(p0) {
			return this.handle.onblur(dom_unwrap(p0));
		}

		Oncancel(p0) {
			return this.handle.oncancel(dom_unwrap(p0));
		}

		Oncanplay(p0) {
			return this.handle.oncanplay(dom_unwrap(p0));
		}

		Oncanplaythrough(p0) {
			return this.handle.oncanplaythrough(dom_unwrap(p0));
		}

		Onchange(p0) {
			return this.handle.onchange(dom_unwrap(p0));
		}

		Onclick(p0) {
			return this.handle.onclick(dom_unwrap(p0));
		}

		Onclose(p0) {
			return this.handle.onclose(dom_unwrap(p0));
		}

		Oncontextlost(p0) {
			return this.handle.oncontextlost(dom_unwrap(p0));
		}

		Oncontextmenu(p0) {
			return this.handle.oncontextmenu(dom_unwrap(p0));
		}

		Oncontextrestored(p0) {
			return this.handle.oncontextrestored(dom_unwrap(p0));
		}

		Oncopy(p0) {
			return this.handle.oncopy(dom_unwrap(p0));
		}

		Oncuechange(p0) {
			return this.handle.oncuechange(dom_unwrap(p0));
		}

		Oncut(p0) {
			return this.handle.oncut(dom_unwrap(p0));
		}

		Ondblclick(p0) {
			return this.handle.ondblclick(dom_unwrap(p0));
		}

		Ondrag(p0) {
			return this.handle.ondrag(dom_unwrap(p0));
		}

		Ondragend(p0) {
			return this.handle.ondragend(dom_unwrap(p0));
		}

		Ondragenter(p0) {
			return this.handle.ondragenter(dom_unwrap(p0));
		}

		Ondragleave(p0) {
			return this.handle.ondragleave(dom_unwrap(p0));
		}

		Ondragover(p0) {
			return this.handle.ondragover(dom_unwrap(p0));
		}

		Ondragstart(p0) {
			return this.handle.ondragstart(dom_unwrap(p0));
		}

		Ondrop(p0) {
			return this.handle.ondrop(dom_unwrap(p0));
		}

		Ondurationchange(p0) {
			return this.handle.ondurationchange(dom_unwrap(p0));
		}

		Onemptied(p0) {
			return this.handle.onemptied(dom_unwrap(p0));
		}

		Onended(p0) {
			return this.handle.onended(dom_unwrap(p0));
		}

		Onfocus(p0) {
			return this.handle.onfocus(dom_unwrap(p0));
		}

		Onformdata(p0) {
			return this.handle.onformdata(dom_unwrap(p0));
		}

		Ongotpointercapture(p0) {
			return this.handle.ongotpointercapture(dom_unwrap(p0));
		}

		Oninput(p0) {
			return this.handle.oninput(dom_unwrap(p0));
		}

		Oninvalid(p0) {
			return this.handle.oninvalid(dom_unwrap(p0));
		}

		Onkeydown(p0) {
			return this.handle.onkeydown(dom_unwrap(p0));
		}

		Onkeypress(p0) {
			return this.handle.onkeypress(dom_unwrap(p0));
		}

		Onkeyup(p0) {
			return this.handle.onkeyup(dom_unwrap(p0));
		}

		Onload(p0) {
			return this.handle.onload(dom_unwrap(p0));
		}

		Onloadeddata(p0) {
			return this.handle.onloadeddata(dom_unwrap(p0));
		}

		Onloadedmetadata(p0) {
			return this.handle.onloadedmetadata(dom_unwrap(p0));
		}

		Onloadstart(p0) {
			return this.handle.onloadstart(dom_unwrap(p0));
		}

		Onlostpointercapture(p0) {
			return this.handle.onlostpointercapture(dom_unwrap(p0));
		}

		Onmousedown(p0) {
			return this.handle.onmousedown(dom_unwrap(p0));
		}

		Onmouseenter(p0) {
			return this.handle.onmouseenter(dom_unwrap(p0));
		}

		Onmouseleave(p0) {
			return this.handle.onmouseleave(dom_unwrap(p0));
		}

		Onmousemove(p0) {
			return this.handle.onmousemove(dom_unwrap(p0));
		}

		Onmouseout(p0) {
			return this.handle.onmouseout(dom_unwrap(p0));
		}

		Onmouseover(p0) {
			return this.handle.onmouseover(dom_unwrap(p0));
		}

		Onmouseup(p0) {
			return this.handle.onmouseup(dom_unwrap(p0));
		}

		Onpaste(p0) {
			return this.handle.onpaste(dom_unwrap(p0));
		}

		Onpause(p0) {
			return this.handle.onpause(dom_unwrap(p0));
		}

		Onplay(p0) {
			return this.handle.onplay(dom_unwrap(p0));
		}

		Onplaying(p0) {
			return this.handle.onplaying(dom_unwrap(p0));
		}

		Onpointercancel(p0) {
			return this.handle.onpointercancel(dom_unwrap(p0));
		}

		Onpointerdown(p0) {
			return this.handle.onpointerdown(dom_unwrap(p0));
		}

		Onpointerenter(p0) {
			return this.handle.onpointerenter(dom_unwrap(p0));
		}

		Onpointerleave(p0) {
			return this.handle.onpointerleave(dom_unwrap(p0));
		}

		Onpointermove(p0) {
			return this.handle.onpointermove(dom_unwrap(p0));
		}

		Onpointerout(p0) {
			return this.handle.onpointerout(dom_unwrap(p0));
		}

		Onpointerover(p0) {
			return this.handle.onpointerover(dom_unwrap(p0));
		}

		Onpointerup(p0) {
			return this.handle.onpointerup(dom_unwrap(p0));
		}

		Onprogress(p0) {
			return this.handle.onprogress(dom_unwrap(p0));
		}

		Onratechange(p0) {
			return this.handle.onratechange(dom_unwrap(p0));
		}

		Onreset(p0) {
			return this.handle.onreset(dom_unwrap(p0));
		}

		Onresize(p0) {
			return this.handle.onresize(dom_unwrap(p0));
		}

		Onscroll(p0) {
			return this.handle.onscroll(dom_unwrap(p0));
		}

		Onscrollend(p0) {
			return this.handle.onscrollend(dom_unwrap(p0));
		}

		Onsecuritypolicyviolation(p0) {
			return this.handle.onsecuritypolicyviolation(dom_unwrap(p0));
		}

		Onseeked(p0) {
			return this.handle.onseeked(dom_unwrap(p0));
		}

		Onseeking(p0) {
			return this.handle.onseeking(dom_unwrap(p0));
		}

		Onselect(p0) {
			return this.handle.onselect(dom_unwrap(p0));
		}

		Onselectionchange(p0) {
			return this.handle.onselectionchange(dom_unwrap(p0));
		}

		Onselectstart(p0) {
			return this.handle.onselectstart(dom_unwrap(p0));
		}

		Onslotchange(p0) {
			return this.handle.onslotchange(dom_unwrap(p0));
		}

		Onstalled(p0) {
			return this.handle.onstalled(dom_unwrap(p0));
		}

		Onsubmit(p0) {
			return this.handle.onsubmit(dom_unwrap(p0));
		}

		Onsuspend(p0) {
			return this.handle.onsuspend(dom_unwrap(p0));
		}

		Ontimeupdate(p0) {
			return this.handle.ontimeupdate(dom_unwrap(p0));
		}

		Ontoggle(p0) {
			return this.handle.ontoggle(dom_unwrap(p0));
		}

		Ontouchcancel(p0) {
			return this.handle.ontouchcancel(dom_unwrap(p0));
		}

		Ontouchend(p0) {
			return this.handle.ontouchend(dom_unwrap(p0));
		}

		Ontouchmove(p0) {
			return this.handle.ontouchmove(dom_unwrap(p0));
		}

		Ontouchstart(p0) {
			return this.handle.ontouchstart(dom_unwrap(p0));
		}

		Ontransitioncancel(p0) {
			return this.handle.ontransitioncancel(dom_unwrap(p0));
		}

		Ontransitionend(p0) {
			return this.handle.ontransitionend(dom_unwrap(p0));
		}

		Ontransitionrun(p0) {
			return this.handle.ontransitionrun(dom_unwrap(p0));
		}

		Ontransitionstart(p0) {
			return this.handle.ontransitionstart(dom_unwrap(p0));
		}

		Onvolumechange(p0) {
			return this.handle.onvolumechange(dom_unwrap(p0));
		}

		Onwaiting(p0) {
			return this.handle.onwaiting(dom_unwrap(p0));
		}

		Onwebkitanimationend(p0) {
			return this.handle.onwebkitanimationend(dom_unwrap(p0));
		}

		Onwebkitanimationiteration(p0) {
			return this.handle.onwebkitanimationiteration(dom_unwrap(p0));
		}

		Onwebkitanimationstart(p0) {
			return this.handle.onwebkitanimationstart(dom_unwrap(p0));
		}

		Onwebkittransitionend(p0) {
			return this.handle.onwebkittransitionend(dom_unwrap(p0));
		}

		Onwheel(p0) {
			return this.handle.onwheel(dom_unwrap(p0));
		}

		Onerror() {
			return this.handle.onerror;
		}

		SetOnerror(value) {
			this.handle.onerror = (value === null || value === undefined ? null : (...a) => dom_unwrapAny(go2jsCallNow(value, null, [a[0], a[1], a[2], a[3], a[4]])));
		}
}

class dom_C_NonElementParentNode extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfNonElementParentNode() {
			
		}

		AsDocument() {
			return dom_is(this, "Document") ? this : null;
		}

		AsDocumentFragment() {
			return dom_is(this, "DocumentFragment") ? this : null;
		}

		AsShadowRoot() {
			return dom_is(this, "ShadowRoot") ? this : null;
		}

		GetElementById(p0) {
			return dom_wrap(dom.Element, this.handle.getElementById(p0));
		}
}

class dom_C_ParentNode extends dom_C_Node {

		KindOfParentNode() {
			
		}

		KindOfNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		AsDocument() {
			return dom_is(this, "Document") ? this : null;
		}

		AsDocumentFragment() {
			return dom_is(this, "DocumentFragment") ? this : null;
		}

		AsElement() {
			return dom_is(this, "Element") ? this : null;
		}

		AsHTMLAudioElement() {
			return dom_is(this, "HTMLAudioElement") ? this : null;
		}

		AsHTMLCanvasElement() {
			return dom_is(this, "HTMLCanvasElement") ? this : null;
		}

		AsHTMLElement() {
			return dom_is(this, "HTMLElement") ? this : null;
		}

		AsHTMLFormElement() {
			return dom_is(this, "HTMLFormElement") ? this : null;
		}

		AsHTMLHeadElement() {
			return dom_is(this, "HTMLHeadElement") ? this : null;
		}

		AsHTMLImageElement() {
			return dom_is(this, "HTMLImageElement") ? this : null;
		}

		AsHTMLMediaElement() {
			return dom_is(this, "HTMLMediaElement") ? this : null;
		}

		AsHTMLSlotElement() {
			return dom_is(this, "HTMLSlotElement") ? this : null;
		}

		AsSVGElement() {
			return dom_is(this, "SVGElement") ? this : null;
		}

		AsSVGGraphicsElement() {
			return dom_is(this, "SVGGraphicsElement") ? this : null;
		}

		AsSVGSVGElement() {
			return dom_is(this, "SVGSVGElement") ? this : null;
		}

		AsShadowRoot() {
			return dom_is(this, "ShadowRoot") ? this : null;
		}

		Append(...p0) {
			this.handle.append(...p0.map((item) => dom_unwrapAny(item)));
		}

		Prepend(...p0) {
			this.handle.prepend(...p0.map((item) => dom_unwrapAny(item)));
		}

		QuerySelector(p0) {
			return dom_wrap(dom.Element, this.handle.querySelector(p0));
		}

		QuerySelectorAll(p0) {
			return dom_wrap(dom.NodeListOf, this.handle.querySelectorAll(p0));
		}

		ReplaceChildren(...p0) {
			this.handle.replaceChildren(...p0.map((item) => dom_unwrapAny(item)));
		}

		ChildElementCount() {
			return this.handle.childElementCount;
		}

		Children() {
			return dom_wrap(dom.HTMLCollection, this.handle.children);
		}

		FirstElementChild() {
			return dom_wrap(dom.Element, this.handle.firstElementChild);
		}

		LastElementChild() {
			return dom_wrap(dom.Element, this.handle.lastElementChild);
		}
}

class dom_C_XPathEvaluatorBase extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfXPathEvaluatorBase() {
			
		}

		AsDocument() {
			return dom_is(this, "Document") ? this : null;
		}

		CreateExpression(p0, ...p1) {
			return dom_wrap(dom.XPathExpression, this.handle.createExpression(p0, ...p1.map((item) => dom_unwrapAny(item))));
		}

		CreateNSResolver(p0) {
			return dom_wrap(dom.Node, this.handle.createNSResolver(dom_unwrap(p0)));
		}

		Evaluate(p0, p1, ...p2) {
			return dom_wrap(dom.XPathResult, this.handle.evaluate(p0, dom_unwrap(p1), ...p2.map((item) => dom_unwrapAny(item))));
		}
}

class dom_C_Document extends dom_C_Node {

		KindOfDocument() {
			
		}

		KindOfNode() {
			
		}

		KindOfDocumentOrShadowRoot() {
			
		}

		KindOfFontFaceSource() {
			
		}

		KindOfGlobalEventHandlers() {
			
		}

		KindOfNonElementParentNode() {
			
		}

		KindOfParentNode() {
			
		}

		KindOfXPathEvaluatorBase() {
			
		}

		KindOfEventTarget() {
			
		}

		AsNode() {
			return this;
		}

		AsDocumentOrShadowRoot() {
			return this;
		}

		AsFontFaceSource() {
			return this;
		}

		AsGlobalEventHandlers() {
			return this;
		}

		AsNonElementParentNode() {
			return this;
		}

		AsParentNode() {
			return this;
		}

		AsXPathEvaluatorBase() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		AdoptNode(p0) {
			return dom_wrap(dom.Node, this.handle.adoptNode(dom_unwrap(p0)));
		}

		CaptureEvents() {
			this.handle.captureEvents();
		}

		CaretRangeFromPoint(p0, p1) {
			return dom_wrap(dom.Range, this.handle.caretRangeFromPoint(p0, p1));
		}

		Clear() {
			this.handle.clear();
		}

		Close() {
			this.handle.close();
		}

		CreateAttribute(p0) {
			return dom_wrap(dom.Attr, this.handle.createAttribute(p0));
		}

		CreateAttributeNS(p0, p1) {
			return dom_wrap(dom.Attr, this.handle.createAttributeNS(p0, p1));
		}

		CreateCDATASection(p0) {
			return dom_wrap(dom.CDATASection, this.handle.createCDATASection(p0));
		}

		CreateComment(p0) {
			return dom_wrap(dom.Comment, this.handle.createComment(p0));
		}

		CreateDocumentFragment() {
			return dom_wrap(dom.DocumentFragment, this.handle.createDocumentFragment());
		}

		CreateElement(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			return dom_wrap(dom.HTMLElement, this.handle.createElement(p0, ...p1.map((item) => dom_toElementCreationOptions(item))));
		}

		CreateElementNS(p0, p1) {
			return dom_wrap(dom.HTMLElement, this.handle.createElementNS(p0, p1));
		}

		CreateEvent(p0) {
			return dom_wrap(dom.AnimationEvent, this.handle.createEvent(p0));
		}

		CreateNodeIterator(p0, ...p1) {
			return dom_wrap(dom.NodeIterator, this.handle.createNodeIterator(dom_unwrap(p0), ...p1.map((item) => dom_unwrapAny(item))));
		}

		CreateProcessingInstruction(p0, p1) {
			return dom_wrap(dom.ProcessingInstruction, this.handle.createProcessingInstruction(p0, p1));
		}

		CreateRange() {
			return dom_wrap(dom.Range, this.handle.createRange());
		}

		CreateTextNode(p0) {
			return dom_wrap(dom.Text, this.handle.createTextNode(p0));
		}

		CreateTreeWalker(p0, ...p1) {
			return dom_wrap(dom.TreeWalker, this.handle.createTreeWalker(dom_unwrap(p0), ...p1.map((item) => dom_unwrapAny(item))));
		}

		ExecCommand(p0, ...p1) {
			return this.handle.execCommand(p0, ...p1.map((item) => dom_unwrapAny(item)));
		}

		*ExitFullscreen() {
			const outcome = yield* dom_await(() => this.handle.exitFullscreen());
			
			return outcome.ok ? null : outcome.error;
		}

		*ExitPictureInPicture() {
			const outcome = yield* dom_await(() => this.handle.exitPictureInPicture());
			
			return outcome.ok ? null : outcome.error;
		}

		ExitPointerLock() {
			this.handle.exitPointerLock();
		}

		GetElementById(p0) {
			return dom_wrap(dom.HTMLElement, this.handle.getElementById(p0));
		}

		GetElementsByClassName(p0) {
			return dom_wrap(dom.HTMLCollectionOf, this.handle.getElementsByClassName(p0));
		}

		GetElementsByName(p0) {
			return dom_wrap(dom.NodeListOf, this.handle.getElementsByName(p0));
		}

		GetElementsByTagName(p0) {
			return dom_wrap(dom.HTMLCollectionOf, this.handle.getElementsByTagName(p0));
		}

		GetElementsByTagNameNS(p0, p1) {
			return dom_wrap(dom.HTMLCollectionOf, this.handle.getElementsByTagNameNS(p0, p1));
		}

		GetSelection() {
			return dom_wrap(dom.Selection, this.handle.getSelection());
		}

		HasFocus() {
			return this.handle.hasFocus();
		}

		*HasStorageAccess() {
			const outcome = yield* dom_await(() => this.handle.hasStorageAccess());
			
			return outcome.ok ? [outcome.value, null] : [false, outcome.error];
		}

		ImportNode(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			return dom_wrap(dom.Node, this.handle.importNode(dom_unwrap(p0), ...p1.map((item) => item)));
		}

		Open(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.Document, this.handle.open(...p0.map((item) => item)));
		}

		QueryCommandEnabled(p0) {
			return this.handle.queryCommandEnabled(p0);
		}

		QueryCommandIndeterm(p0) {
			return this.handle.queryCommandIndeterm(p0);
		}

		QueryCommandState(p0) {
			return this.handle.queryCommandState(p0);
		}

		QueryCommandSupported(p0) {
			return this.handle.queryCommandSupported(p0);
		}

		QueryCommandValue(p0) {
			return this.handle.queryCommandValue(p0);
		}

		ReleaseEvents() {
			this.handle.releaseEvents();
		}

		*RequestStorageAccess() {
			const outcome = yield* dom_await(() => this.handle.requestStorageAccess());
			
			return outcome.ok ? null : outcome.error;
		}

		StartViewTransition(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.ViewTransition, this.handle.startViewTransition(...p0.map((item) => (item === null || item === undefined ? null : (...a) => dom_unwrapAny(go2jsCallNow(item, null, []))))));
		}

		Write(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.write(...p0.map((item) => item));
		}

		Writeln(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.writeln(...p0.map((item) => item));
		}

		ElementFromPoint(p0, p1) {
			return dom_wrap(dom.Element, this.handle.elementFromPoint(p0, p1));
		}

		ElementsFromPoint(p0, p1) {
			return Array.from(this.handle.elementsFromPoint(p0, p1) || [], (item) => dom_wrap(dom.Element, item));
		}

		GetAnimations() {
			return Array.from(this.handle.getAnimations() || [], (item) => dom_wrap(dom.Animation, item));
		}

		Append(...p0) {
			this.handle.append(...p0.map((item) => dom_unwrapAny(item)));
		}

		Prepend(...p0) {
			this.handle.prepend(...p0.map((item) => dom_unwrapAny(item)));
		}

		QuerySelector(p0) {
			return dom_wrap(dom.Element, this.handle.querySelector(p0));
		}

		QuerySelectorAll(p0) {
			return dom_wrap(dom.NodeListOf, this.handle.querySelectorAll(p0));
		}

		ReplaceChildren(...p0) {
			this.handle.replaceChildren(...p0.map((item) => dom_unwrapAny(item)));
		}

		CreateExpression(p0, ...p1) {
			return dom_wrap(dom.XPathExpression, this.handle.createExpression(p0, ...p1.map((item) => dom_unwrapAny(item))));
		}

		CreateNSResolver(p0) {
			return dom_wrap(dom.Node, this.handle.createNSResolver(dom_unwrap(p0)));
		}

		Evaluate(p0, p1, ...p2) {
			return dom_wrap(dom.XPathResult, this.handle.evaluate(p0, dom_unwrap(p1), ...p2.map((item) => dom_unwrapAny(item))));
		}

		OnDOMContentLoaded(p0) {
			this.handle.addEventListener("DOMContentLoaded", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnFullscreenchange(p0) {
			this.handle.addEventListener("fullscreenchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnFullscreenerror(p0) {
			this.handle.addEventListener("fullscreenerror", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnPointerlockchange(p0) {
			this.handle.addEventListener("pointerlockchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnPointerlockerror(p0) {
			this.handle.addEventListener("pointerlockerror", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnReadystatechange(p0) {
			this.handle.addEventListener("readystatechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnVisibilitychange(p0) {
			this.handle.addEventListener("visibilitychange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnAbort(p0) {
			this.handle.addEventListener("abort", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.UIEvent, a[0])])));
		}

		OnAnimationcancel(p0) {
			this.handle.addEventListener("animationcancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationEvent, a[0])])));
		}

		OnAnimationend(p0) {
			this.handle.addEventListener("animationend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationEvent, a[0])])));
		}

		OnAnimationiteration(p0) {
			this.handle.addEventListener("animationiteration", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationEvent, a[0])])));
		}

		OnAnimationstart(p0) {
			this.handle.addEventListener("animationstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationEvent, a[0])])));
		}

		OnAuxclick(p0) {
			this.handle.addEventListener("auxclick", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnBeforeinput(p0) {
			this.handle.addEventListener("beforeinput", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.InputEvent, a[0])])));
		}

		OnBeforetoggle(p0) {
			this.handle.addEventListener("beforetoggle", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnBlur(p0) {
			this.handle.addEventListener("blur", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FocusEvent, a[0])])));
		}

		OnCancel(p0) {
			this.handle.addEventListener("cancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCanplay(p0) {
			this.handle.addEventListener("canplay", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCanplaythrough(p0) {
			this.handle.addEventListener("canplaythrough", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnChange(p0) {
			this.handle.addEventListener("change", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnClick(p0) {
			this.handle.addEventListener("click", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnClose(p0) {
			this.handle.addEventListener("close", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCompositionend(p0) {
			this.handle.addEventListener("compositionend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.CompositionEvent, a[0])])));
		}

		OnCompositionstart(p0) {
			this.handle.addEventListener("compositionstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.CompositionEvent, a[0])])));
		}

		OnCompositionupdate(p0) {
			this.handle.addEventListener("compositionupdate", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.CompositionEvent, a[0])])));
		}

		OnContextlost(p0) {
			this.handle.addEventListener("contextlost", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnContextmenu(p0) {
			this.handle.addEventListener("contextmenu", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnContextrestored(p0) {
			this.handle.addEventListener("contextrestored", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCopy(p0) {
			this.handle.addEventListener("copy", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ClipboardEvent, a[0])])));
		}

		OnCuechange(p0) {
			this.handle.addEventListener("cuechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCut(p0) {
			this.handle.addEventListener("cut", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ClipboardEvent, a[0])])));
		}

		OnDblclick(p0) {
			this.handle.addEventListener("dblclick", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnDrag(p0) {
			this.handle.addEventListener("drag", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragend(p0) {
			this.handle.addEventListener("dragend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragenter(p0) {
			this.handle.addEventListener("dragenter", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragleave(p0) {
			this.handle.addEventListener("dragleave", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragover(p0) {
			this.handle.addEventListener("dragover", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragstart(p0) {
			this.handle.addEventListener("dragstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDrop(p0) {
			this.handle.addEventListener("drop", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDurationchange(p0) {
			this.handle.addEventListener("durationchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnEmptied(p0) {
			this.handle.addEventListener("emptied", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnEnded(p0) {
			this.handle.addEventListener("ended", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnError(p0) {
			this.handle.addEventListener("error", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ErrorEvent, a[0])])));
		}

		OnFocus(p0) {
			this.handle.addEventListener("focus", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FocusEvent, a[0])])));
		}

		OnFocusin(p0) {
			this.handle.addEventListener("focusin", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FocusEvent, a[0])])));
		}

		OnFocusout(p0) {
			this.handle.addEventListener("focusout", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FocusEvent, a[0])])));
		}

		OnFormdata(p0) {
			this.handle.addEventListener("formdata", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FormDataEvent, a[0])])));
		}

		OnGotpointercapture(p0) {
			this.handle.addEventListener("gotpointercapture", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnInput(p0) {
			this.handle.addEventListener("input", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnInvalid(p0) {
			this.handle.addEventListener("invalid", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnKeydown(p0) {
			this.handle.addEventListener("keydown", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.KeyboardEvent, a[0])])));
		}

		OnKeypress(p0) {
			this.handle.addEventListener("keypress", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.KeyboardEvent, a[0])])));
		}

		OnKeyup(p0) {
			this.handle.addEventListener("keyup", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.KeyboardEvent, a[0])])));
		}

		OnLoad(p0) {
			this.handle.addEventListener("load", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLoadeddata(p0) {
			this.handle.addEventListener("loadeddata", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLoadedmetadata(p0) {
			this.handle.addEventListener("loadedmetadata", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLoadstart(p0) {
			this.handle.addEventListener("loadstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLostpointercapture(p0) {
			this.handle.addEventListener("lostpointercapture", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnMousedown(p0) {
			this.handle.addEventListener("mousedown", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseenter(p0) {
			this.handle.addEventListener("mouseenter", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseleave(p0) {
			this.handle.addEventListener("mouseleave", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMousemove(p0) {
			this.handle.addEventListener("mousemove", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseout(p0) {
			this.handle.addEventListener("mouseout", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseover(p0) {
			this.handle.addEventListener("mouseover", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseup(p0) {
			this.handle.addEventListener("mouseup", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnPaste(p0) {
			this.handle.addEventListener("paste", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ClipboardEvent, a[0])])));
		}

		OnPause(p0) {
			this.handle.addEventListener("pause", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnPlay(p0) {
			this.handle.addEventListener("play", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnPlaying(p0) {
			this.handle.addEventListener("playing", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnPointercancel(p0) {
			this.handle.addEventListener("pointercancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerdown(p0) {
			this.handle.addEventListener("pointerdown", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerenter(p0) {
			this.handle.addEventListener("pointerenter", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerleave(p0) {
			this.handle.addEventListener("pointerleave", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointermove(p0) {
			this.handle.addEventListener("pointermove", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerout(p0) {
			this.handle.addEventListener("pointerout", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerover(p0) {
			this.handle.addEventListener("pointerover", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerup(p0) {
			this.handle.addEventListener("pointerup", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnProgress(p0) {
			this.handle.addEventListener("progress", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ProgressEvent, a[0])])));
		}

		OnRatechange(p0) {
			this.handle.addEventListener("ratechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnReset(p0) {
			this.handle.addEventListener("reset", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnResize(p0) {
			this.handle.addEventListener("resize", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.UIEvent, a[0])])));
		}

		OnScroll(p0) {
			this.handle.addEventListener("scroll", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnScrollend(p0) {
			this.handle.addEventListener("scrollend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSecuritypolicyviolation(p0) {
			this.handle.addEventListener("securitypolicyviolation", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.SecurityPolicyViolationEvent, a[0])])));
		}

		OnSeeked(p0) {
			this.handle.addEventListener("seeked", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSeeking(p0) {
			this.handle.addEventListener("seeking", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSelect(p0) {
			this.handle.addEventListener("select", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSelectionchange(p0) {
			this.handle.addEventListener("selectionchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSelectstart(p0) {
			this.handle.addEventListener("selectstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSlotchange(p0) {
			this.handle.addEventListener("slotchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnStalled(p0) {
			this.handle.addEventListener("stalled", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSubmit(p0) {
			this.handle.addEventListener("submit", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.SubmitEvent, a[0])])));
		}

		OnSuspend(p0) {
			this.handle.addEventListener("suspend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnTimeupdate(p0) {
			this.handle.addEventListener("timeupdate", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnToggle(p0) {
			this.handle.addEventListener("toggle", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnTouchcancel(p0) {
			this.handle.addEventListener("touchcancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TouchEvent, a[0])])));
		}

		OnTouchend(p0) {
			this.handle.addEventListener("touchend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TouchEvent, a[0])])));
		}

		OnTouchmove(p0) {
			this.handle.addEventListener("touchmove", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TouchEvent, a[0])])));
		}

		OnTouchstart(p0) {
			this.handle.addEventListener("touchstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TouchEvent, a[0])])));
		}

		OnTransitioncancel(p0) {
			this.handle.addEventListener("transitioncancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TransitionEvent, a[0])])));
		}

		OnTransitionend(p0) {
			this.handle.addEventListener("transitionend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TransitionEvent, a[0])])));
		}

		OnTransitionrun(p0) {
			this.handle.addEventListener("transitionrun", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TransitionEvent, a[0])])));
		}

		OnTransitionstart(p0) {
			this.handle.addEventListener("transitionstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TransitionEvent, a[0])])));
		}

		OnVolumechange(p0) {
			this.handle.addEventListener("volumechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWaiting(p0) {
			this.handle.addEventListener("waiting", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWebkitanimationend(p0) {
			this.handle.addEventListener("webkitanimationend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWebkitanimationiteration(p0) {
			this.handle.addEventListener("webkitanimationiteration", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWebkitanimationstart(p0) {
			this.handle.addEventListener("webkitanimationstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWebkittransitionend(p0) {
			this.handle.addEventListener("webkittransitionend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWheel(p0) {
			this.handle.addEventListener("wheel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.WheelEvent, a[0])])));
		}

		Onfullscreenchange(p0) {
			return this.handle.onfullscreenchange(dom_unwrap(p0));
		}

		Onfullscreenerror(p0) {
			return this.handle.onfullscreenerror(dom_unwrap(p0));
		}

		Onpointerlockchange(p0) {
			return this.handle.onpointerlockchange(dom_unwrap(p0));
		}

		Onpointerlockerror(p0) {
			return this.handle.onpointerlockerror(dom_unwrap(p0));
		}

		Onreadystatechange(p0) {
			return this.handle.onreadystatechange(dom_unwrap(p0));
		}

		Onvisibilitychange(p0) {
			return this.handle.onvisibilitychange(dom_unwrap(p0));
		}

		Onabort(p0) {
			return this.handle.onabort(dom_unwrap(p0));
		}

		Onanimationcancel(p0) {
			return this.handle.onanimationcancel(dom_unwrap(p0));
		}

		Onanimationend(p0) {
			return this.handle.onanimationend(dom_unwrap(p0));
		}

		Onanimationiteration(p0) {
			return this.handle.onanimationiteration(dom_unwrap(p0));
		}

		Onanimationstart(p0) {
			return this.handle.onanimationstart(dom_unwrap(p0));
		}

		Onauxclick(p0) {
			return this.handle.onauxclick(dom_unwrap(p0));
		}

		Onbeforeinput(p0) {
			return this.handle.onbeforeinput(dom_unwrap(p0));
		}

		Onbeforetoggle(p0) {
			return this.handle.onbeforetoggle(dom_unwrap(p0));
		}

		Onblur(p0) {
			return this.handle.onblur(dom_unwrap(p0));
		}

		Oncancel(p0) {
			return this.handle.oncancel(dom_unwrap(p0));
		}

		Oncanplay(p0) {
			return this.handle.oncanplay(dom_unwrap(p0));
		}

		Oncanplaythrough(p0) {
			return this.handle.oncanplaythrough(dom_unwrap(p0));
		}

		Onchange(p0) {
			return this.handle.onchange(dom_unwrap(p0));
		}

		Onclick(p0) {
			return this.handle.onclick(dom_unwrap(p0));
		}

		Onclose(p0) {
			return this.handle.onclose(dom_unwrap(p0));
		}

		Oncontextlost(p0) {
			return this.handle.oncontextlost(dom_unwrap(p0));
		}

		Oncontextmenu(p0) {
			return this.handle.oncontextmenu(dom_unwrap(p0));
		}

		Oncontextrestored(p0) {
			return this.handle.oncontextrestored(dom_unwrap(p0));
		}

		Oncopy(p0) {
			return this.handle.oncopy(dom_unwrap(p0));
		}

		Oncuechange(p0) {
			return this.handle.oncuechange(dom_unwrap(p0));
		}

		Oncut(p0) {
			return this.handle.oncut(dom_unwrap(p0));
		}

		Ondblclick(p0) {
			return this.handle.ondblclick(dom_unwrap(p0));
		}

		Ondrag(p0) {
			return this.handle.ondrag(dom_unwrap(p0));
		}

		Ondragend(p0) {
			return this.handle.ondragend(dom_unwrap(p0));
		}

		Ondragenter(p0) {
			return this.handle.ondragenter(dom_unwrap(p0));
		}

		Ondragleave(p0) {
			return this.handle.ondragleave(dom_unwrap(p0));
		}

		Ondragover(p0) {
			return this.handle.ondragover(dom_unwrap(p0));
		}

		Ondragstart(p0) {
			return this.handle.ondragstart(dom_unwrap(p0));
		}

		Ondrop(p0) {
			return this.handle.ondrop(dom_unwrap(p0));
		}

		Ondurationchange(p0) {
			return this.handle.ondurationchange(dom_unwrap(p0));
		}

		Onemptied(p0) {
			return this.handle.onemptied(dom_unwrap(p0));
		}

		Onended(p0) {
			return this.handle.onended(dom_unwrap(p0));
		}

		Onfocus(p0) {
			return this.handle.onfocus(dom_unwrap(p0));
		}

		Onformdata(p0) {
			return this.handle.onformdata(dom_unwrap(p0));
		}

		Ongotpointercapture(p0) {
			return this.handle.ongotpointercapture(dom_unwrap(p0));
		}

		Oninput(p0) {
			return this.handle.oninput(dom_unwrap(p0));
		}

		Oninvalid(p0) {
			return this.handle.oninvalid(dom_unwrap(p0));
		}

		Onkeydown(p0) {
			return this.handle.onkeydown(dom_unwrap(p0));
		}

		Onkeypress(p0) {
			return this.handle.onkeypress(dom_unwrap(p0));
		}

		Onkeyup(p0) {
			return this.handle.onkeyup(dom_unwrap(p0));
		}

		Onload(p0) {
			return this.handle.onload(dom_unwrap(p0));
		}

		Onloadeddata(p0) {
			return this.handle.onloadeddata(dom_unwrap(p0));
		}

		Onloadedmetadata(p0) {
			return this.handle.onloadedmetadata(dom_unwrap(p0));
		}

		Onloadstart(p0) {
			return this.handle.onloadstart(dom_unwrap(p0));
		}

		Onlostpointercapture(p0) {
			return this.handle.onlostpointercapture(dom_unwrap(p0));
		}

		Onmousedown(p0) {
			return this.handle.onmousedown(dom_unwrap(p0));
		}

		Onmouseenter(p0) {
			return this.handle.onmouseenter(dom_unwrap(p0));
		}

		Onmouseleave(p0) {
			return this.handle.onmouseleave(dom_unwrap(p0));
		}

		Onmousemove(p0) {
			return this.handle.onmousemove(dom_unwrap(p0));
		}

		Onmouseout(p0) {
			return this.handle.onmouseout(dom_unwrap(p0));
		}

		Onmouseover(p0) {
			return this.handle.onmouseover(dom_unwrap(p0));
		}

		Onmouseup(p0) {
			return this.handle.onmouseup(dom_unwrap(p0));
		}

		Onpaste(p0) {
			return this.handle.onpaste(dom_unwrap(p0));
		}

		Onpause(p0) {
			return this.handle.onpause(dom_unwrap(p0));
		}

		Onplay(p0) {
			return this.handle.onplay(dom_unwrap(p0));
		}

		Onplaying(p0) {
			return this.handle.onplaying(dom_unwrap(p0));
		}

		Onpointercancel(p0) {
			return this.handle.onpointercancel(dom_unwrap(p0));
		}

		Onpointerdown(p0) {
			return this.handle.onpointerdown(dom_unwrap(p0));
		}

		Onpointerenter(p0) {
			return this.handle.onpointerenter(dom_unwrap(p0));
		}

		Onpointerleave(p0) {
			return this.handle.onpointerleave(dom_unwrap(p0));
		}

		Onpointermove(p0) {
			return this.handle.onpointermove(dom_unwrap(p0));
		}

		Onpointerout(p0) {
			return this.handle.onpointerout(dom_unwrap(p0));
		}

		Onpointerover(p0) {
			return this.handle.onpointerover(dom_unwrap(p0));
		}

		Onpointerup(p0) {
			return this.handle.onpointerup(dom_unwrap(p0));
		}

		Onprogress(p0) {
			return this.handle.onprogress(dom_unwrap(p0));
		}

		Onratechange(p0) {
			return this.handle.onratechange(dom_unwrap(p0));
		}

		Onreset(p0) {
			return this.handle.onreset(dom_unwrap(p0));
		}

		Onresize(p0) {
			return this.handle.onresize(dom_unwrap(p0));
		}

		Onscroll(p0) {
			return this.handle.onscroll(dom_unwrap(p0));
		}

		Onscrollend(p0) {
			return this.handle.onscrollend(dom_unwrap(p0));
		}

		Onsecuritypolicyviolation(p0) {
			return this.handle.onsecuritypolicyviolation(dom_unwrap(p0));
		}

		Onseeked(p0) {
			return this.handle.onseeked(dom_unwrap(p0));
		}

		Onseeking(p0) {
			return this.handle.onseeking(dom_unwrap(p0));
		}

		Onselect(p0) {
			return this.handle.onselect(dom_unwrap(p0));
		}

		Onselectionchange(p0) {
			return this.handle.onselectionchange(dom_unwrap(p0));
		}

		Onselectstart(p0) {
			return this.handle.onselectstart(dom_unwrap(p0));
		}

		Onslotchange(p0) {
			return this.handle.onslotchange(dom_unwrap(p0));
		}

		Onstalled(p0) {
			return this.handle.onstalled(dom_unwrap(p0));
		}

		Onsubmit(p0) {
			return this.handle.onsubmit(dom_unwrap(p0));
		}

		Onsuspend(p0) {
			return this.handle.onsuspend(dom_unwrap(p0));
		}

		Ontimeupdate(p0) {
			return this.handle.ontimeupdate(dom_unwrap(p0));
		}

		Ontoggle(p0) {
			return this.handle.ontoggle(dom_unwrap(p0));
		}

		Ontouchcancel(p0) {
			return this.handle.ontouchcancel(dom_unwrap(p0));
		}

		Ontouchend(p0) {
			return this.handle.ontouchend(dom_unwrap(p0));
		}

		Ontouchmove(p0) {
			return this.handle.ontouchmove(dom_unwrap(p0));
		}

		Ontouchstart(p0) {
			return this.handle.ontouchstart(dom_unwrap(p0));
		}

		Ontransitioncancel(p0) {
			return this.handle.ontransitioncancel(dom_unwrap(p0));
		}

		Ontransitionend(p0) {
			return this.handle.ontransitionend(dom_unwrap(p0));
		}

		Ontransitionrun(p0) {
			return this.handle.ontransitionrun(dom_unwrap(p0));
		}

		Ontransitionstart(p0) {
			return this.handle.ontransitionstart(dom_unwrap(p0));
		}

		Onvolumechange(p0) {
			return this.handle.onvolumechange(dom_unwrap(p0));
		}

		Onwaiting(p0) {
			return this.handle.onwaiting(dom_unwrap(p0));
		}

		Onwebkitanimationend(p0) {
			return this.handle.onwebkitanimationend(dom_unwrap(p0));
		}

		Onwebkitanimationiteration(p0) {
			return this.handle.onwebkitanimationiteration(dom_unwrap(p0));
		}

		Onwebkitanimationstart(p0) {
			return this.handle.onwebkitanimationstart(dom_unwrap(p0));
		}

		Onwebkittransitionend(p0) {
			return this.handle.onwebkittransitionend(dom_unwrap(p0));
		}

		Onwheel(p0) {
			return this.handle.onwheel(dom_unwrap(p0));
		}

		URL() {
			return this.handle.URL;
		}

		AlinkColor() {
			return this.handle.alinkColor;
		}

		SetAlinkColor(value) {
			this.handle.alinkColor = value;
		}

		All() {
			return dom_wrap(dom.HTMLAllCollection, this.handle.all);
		}

		Anchors() {
			return dom_wrap(dom.HTMLCollectionOf, this.handle.anchors);
		}

		Applets() {
			return dom_wrap(dom.HTMLCollection, this.handle.applets);
		}

		BgColor() {
			return this.handle.bgColor;
		}

		SetBgColor(value) {
			this.handle.bgColor = value;
		}

		Body() {
			return dom_wrap(dom.HTMLElement, this.handle.body);
		}

		SetBody(value) {
			this.handle.body = dom_unwrap(value);
		}

		CharacterSet() {
			return this.handle.characterSet;
		}

		Charset() {
			return this.handle.charset;
		}

		CompatMode() {
			return this.handle.compatMode;
		}

		ContentType() {
			return this.handle.contentType;
		}

		Cookie() {
			return this.handle.cookie;
		}

		SetCookie(value) {
			this.handle.cookie = value;
		}

		CurrentScript() {
			return this.handle.currentScript;
		}

		DefaultView() {
			return dom_wrap(dom.Window, this.handle.defaultView);
		}

		DesignMode() {
			return this.handle.designMode;
		}

		SetDesignMode(value) {
			this.handle.designMode = value;
		}

		Dir() {
			return this.handle.dir;
		}

		SetDir(value) {
			this.handle.dir = value;
		}

		Doctype() {
			return dom_wrap(dom.DocumentType, this.handle.doctype);
		}

		DocumentElement() {
			return dom_wrap(dom.HTMLElement, this.handle.documentElement);
		}

		DocumentURI() {
			return this.handle.documentURI;
		}

		Domain() {
			return this.handle.domain;
		}

		SetDomain(value) {
			this.handle.domain = value;
		}

		Embeds() {
			return dom_wrap(dom.HTMLCollectionOf, this.handle.embeds);
		}

		FgColor() {
			return this.handle.fgColor;
		}

		SetFgColor(value) {
			this.handle.fgColor = value;
		}

		Forms() {
			return dom_wrap(dom.HTMLCollectionOf, this.handle.forms);
		}

		Fullscreen() {
			return this.handle.fullscreen;
		}

		FullscreenEnabled() {
			return this.handle.fullscreenEnabled;
		}

		Head() {
			return dom_wrap(dom.HTMLHeadElement, this.handle.head);
		}

		Hidden() {
			return this.handle.hidden;
		}

		Images() {
			return dom_wrap(dom.HTMLCollectionOf, this.handle.images);
		}

		Implementation() {
			return dom_wrap(dom.DOMImplementation, this.handle.implementation);
		}

		InputEncoding() {
			return this.handle.inputEncoding;
		}

		LastModified() {
			return this.handle.lastModified;
		}

		LinkColor() {
			return this.handle.linkColor;
		}

		SetLinkColor(value) {
			this.handle.linkColor = value;
		}

		Links() {
			return dom_wrap(dom.HTMLCollectionOf, this.handle.links);
		}

		Location() {
			return dom_wrap(dom.Location, this.handle.location);
		}

		SetLocation(value) {
			this.handle.location = dom_unwrap(value);
		}

		PictureInPictureEnabled() {
			return this.handle.pictureInPictureEnabled;
		}

		Plugins() {
			return dom_wrap(dom.HTMLCollectionOf, this.handle.plugins);
		}

		ReadyState() {
			return this.handle.readyState;
		}

		Referrer() {
			return this.handle.referrer;
		}

		RootElement() {
			return dom_wrap(dom.SVGSVGElement, this.handle.rootElement);
		}

		Scripts() {
			return dom_wrap(dom.HTMLCollectionOf, this.handle.scripts);
		}

		ScrollingElement() {
			return dom_wrap(dom.Element, this.handle.scrollingElement);
		}

		Timeline() {
			return dom_fromDocumentTimeline(this.handle.timeline);
		}

		Title() {
			return this.handle.title;
		}

		SetTitle(value) {
			this.handle.title = value;
		}

		VisibilityState() {
			return this.handle.visibilityState;
		}

		VlinkColor() {
			return this.handle.vlinkColor;
		}

		SetVlinkColor(value) {
			this.handle.vlinkColor = value;
		}

		ActiveElement() {
			return dom_wrap(dom.Element, this.handle.activeElement);
		}

		AdoptedStyleSheets() {
			return Array.from(this.handle.adoptedStyleSheets || [], (item) => dom_wrap(dom.CSSStyleSheet, item));
		}

		SetAdoptedStyleSheets(value) {
			this.handle.adoptedStyleSheets = Array.from(value || [], (item) => dom_unwrap(item));
		}

		FullscreenElement() {
			return dom_wrap(dom.Element, this.handle.fullscreenElement);
		}

		PictureInPictureElement() {
			return dom_wrap(dom.Element, this.handle.pictureInPictureElement);
		}

		PointerLockElement() {
			return dom_wrap(dom.Element, this.handle.pointerLockElement);
		}

		StyleSheets() {
			return dom_wrap(dom.StyleSheetList, this.handle.styleSheets);
		}

		Fonts() {
			return dom_wrap(dom.FontFaceSet, this.handle.fonts);
		}

		Onerror() {
			return this.handle.onerror;
		}

		SetOnerror(value) {
			this.handle.onerror = (value === null || value === undefined ? null : (...a) => dom_unwrapAny(go2jsCallNow(value, null, [a[0], a[1], a[2], a[3], a[4]])));
		}

		ChildElementCount() {
			return this.handle.childElementCount;
		}

		Children() {
			return dom_wrap(dom.HTMLCollection, this.handle.children);
		}

		FirstElementChild() {
			return dom_wrap(dom.Element, this.handle.firstElementChild);
		}

		LastElementChild() {
			return dom_wrap(dom.Element, this.handle.lastElementChild);
		}
}

class dom_C_DocumentType extends dom_C_Node {

		KindOfDocumentType() {
			
		}

		KindOfNode() {
			
		}

		KindOfChildNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsNode() {
			return this;
		}

		AsChildNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		ParseHTMLUnsafe(p0) {
			return dom_wrap(dom.Document, this.handle.parseHTMLUnsafe(p0));
		}

		After(...p0) {
			this.handle.after(...p0.map((item) => dom_unwrapAny(item)));
		}

		Before(...p0) {
			this.handle.before(...p0.map((item) => dom_unwrapAny(item)));
		}

		Remove() {
			this.handle.remove();
		}

		ReplaceWith(...p0) {
			this.handle.replaceWith(...p0.map((item) => dom_unwrapAny(item)));
		}

		Prototype() {
			return dom_wrap(dom.Document, this.handle.prototype);
		}

		SetPrototype(value) {
			this.handle.prototype = dom_unwrap(value);
		}

		Name() {
			return this.handle.name;
		}

		PublicId() {
			return this.handle.publicId;
		}

		SystemId() {
			return this.handle.systemId;
		}
}

class dom_C_DocumentFragment extends dom_C_Node {

		KindOfDocumentFragment() {
			
		}

		KindOfNode() {
			
		}

		KindOfNonElementParentNode() {
			
		}

		KindOfParentNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsNode() {
			return this;
		}

		AsNonElementParentNode() {
			return this;
		}

		AsParentNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		AsShadowRoot() {
			return dom_is(this, "ShadowRoot") ? this : null;
		}

		GetElementById(p0) {
			return dom_wrap(dom.HTMLElement, this.handle.getElementById(p0));
		}

		Append(...p0) {
			this.handle.append(...p0.map((item) => dom_unwrapAny(item)));
		}

		Prepend(...p0) {
			this.handle.prepend(...p0.map((item) => dom_unwrapAny(item)));
		}

		QuerySelector(p0) {
			return dom_wrap(dom.Element, this.handle.querySelector(p0));
		}

		QuerySelectorAll(p0) {
			return dom_wrap(dom.NodeListOf, this.handle.querySelectorAll(p0));
		}

		ReplaceChildren(...p0) {
			this.handle.replaceChildren(...p0.map((item) => dom_unwrapAny(item)));
		}

		ChildElementCount() {
			return this.handle.childElementCount;
		}

		Children() {
			return dom_wrap(dom.HTMLCollection, this.handle.children);
		}

		FirstElementChild() {
			return dom_wrap(dom.Element, this.handle.firstElementChild);
		}

		LastElementChild() {
			return dom_wrap(dom.Element, this.handle.lastElementChild);
		}
}

class dom_C_MouseEvent extends dom_C_UIEvent {

		KindOfMouseEvent() {
			
		}

		KindOfUIEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsUIEvent() {
			return this;
		}

		AsEvent() {
			return this;
		}

		AsDragEvent() {
			return dom_is(this, "DragEvent") ? this : null;
		}

		AsPointerEvent() {
			return dom_is(this, "PointerEvent") ? this : null;
		}

		AsWheelEvent() {
			return dom_is(this, "WheelEvent") ? this : null;
		}

		GetModifierState(p0) {
			return this.handle.getModifierState(p0);
		}

		InitMouseEvent(p0, p1, p2, p3, p4, p5, p6, p7, p8, p9, p10, p11, p12, p13, p14) {
			this.handle.initMouseEvent(p0, p1, p2, dom_unwrap(p3), p4, p5, p6, p7, p8, p9, p10, p11, p12, p13, dom_unwrap(p14));
		}

		AltKey() {
			return this.handle.altKey;
		}

		Button() {
			return this.handle.button;
		}

		Buttons() {
			return this.handle.buttons;
		}

		ClientX() {
			return this.handle.clientX;
		}

		ClientY() {
			return this.handle.clientY;
		}

		CtrlKey() {
			return this.handle.ctrlKey;
		}

		LayerX() {
			return this.handle.layerX;
		}

		LayerY() {
			return this.handle.layerY;
		}

		MetaKey() {
			return this.handle.metaKey;
		}

		MovementX() {
			return this.handle.movementX;
		}

		MovementY() {
			return this.handle.movementY;
		}

		OffsetX() {
			return this.handle.offsetX;
		}

		OffsetY() {
			return this.handle.offsetY;
		}

		PageX() {
			return this.handle.pageX;
		}

		PageY() {
			return this.handle.pageY;
		}

		RelatedTarget() {
			return dom_wrap(dom.EventTarget, this.handle.relatedTarget);
		}

		ScreenX() {
			return this.handle.screenX;
		}

		ScreenY() {
			return this.handle.screenY;
		}

		ShiftKey() {
			return this.handle.shiftKey;
		}

		X() {
			return this.handle.x;
		}

		Y() {
			return this.handle.y;
		}
}

class dom_C_DragEvent extends dom_C_MouseEvent {

		KindOfDragEvent() {
			
		}

		KindOfMouseEvent() {
			
		}

		KindOfUIEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsMouseEvent() {
			return this;
		}

		AsUIEvent() {
			return this;
		}

		AsEvent() {
			return this;
		}

		DataTransfer() {
			return dom_wrap(dom.DataTransfer, this.handle.dataTransfer);
		}
}

class dom_C_DynamicsCompressorNode extends dom_C_AudioNode {

		KindOfDynamicsCompressorNode() {
			
		}

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsAudioNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		Attack() {
			return dom_wrap(dom.AudioParam, this.handle.attack);
		}

		Knee() {
			return dom_wrap(dom.AudioParam, this.handle.knee);
		}

		Ratio() {
			return dom_wrap(dom.AudioParam, this.handle.ratio);
		}

		Reduction() {
			return this.handle.reduction;
		}

		Release() {
			return dom_wrap(dom.AudioParam, this.handle.release);
		}

		Threshold() {
			return dom_wrap(dom.AudioParam, this.handle.threshold);
		}
}

class dom_C_Element extends dom_C_Node {

		KindOfElement() {
			
		}

		KindOfNode() {
			
		}

		KindOfARIAMixin() {
			
		}

		KindOfAnimatable() {
			
		}

		KindOfChildNode() {
			
		}

		KindOfNonDocumentTypeChildNode() {
			
		}

		KindOfParentNode() {
			
		}

		KindOfSlottable() {
			
		}

		KindOfEventTarget() {
			
		}

		AsNode() {
			return this;
		}

		AsARIAMixin() {
			return this;
		}

		AsAnimatable() {
			return this;
		}

		AsChildNode() {
			return this;
		}

		AsNonDocumentTypeChildNode() {
			return this;
		}

		AsParentNode() {
			return this;
		}

		AsSlottable() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		AsHTMLAudioElement() {
			return dom_is(this, "HTMLAudioElement") ? this : null;
		}

		AsHTMLCanvasElement() {
			return dom_is(this, "HTMLCanvasElement") ? this : null;
		}

		AsHTMLElement() {
			return dom_is(this, "HTMLElement") ? this : null;
		}

		AsHTMLFormElement() {
			return dom_is(this, "HTMLFormElement") ? this : null;
		}

		AsHTMLHeadElement() {
			return dom_is(this, "HTMLHeadElement") ? this : null;
		}

		AsHTMLImageElement() {
			return dom_is(this, "HTMLImageElement") ? this : null;
		}

		AsHTMLMediaElement() {
			return dom_is(this, "HTMLMediaElement") ? this : null;
		}

		AsHTMLSlotElement() {
			return dom_is(this, "HTMLSlotElement") ? this : null;
		}

		AsSVGElement() {
			return dom_is(this, "SVGElement") ? this : null;
		}

		AsSVGGraphicsElement() {
			return dom_is(this, "SVGGraphicsElement") ? this : null;
		}

		AsSVGSVGElement() {
			return dom_is(this, "SVGSVGElement") ? this : null;
		}

		AttachShadow(p0) {
			return dom_wrap(dom.ShadowRoot, this.handle.attachShadow(dom_toShadowRootInit(p0)));
		}

		CheckVisibility(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return this.handle.checkVisibility(...p0.map((item) => dom_toCheckVisibilityOptions(item)));
		}

		Closest(p0) {
			return dom_wrap(dom.Element, this.handle.closest(p0));
		}

		ComputedStyleMap() {
			return dom_wrap(dom.StylePropertyMapReadOnly, this.handle.computedStyleMap());
		}

		GetAttribute(p0) {
			return this.handle.getAttribute(p0);
		}

		GetAttributeNS(p0, p1) {
			return this.handle.getAttributeNS(p0, p1);
		}

		GetAttributeNames() {
			return Array.from(this.handle.getAttributeNames() || []);
		}

		GetAttributeNode(p0) {
			return dom_wrap(dom.Attr, this.handle.getAttributeNode(p0));
		}

		GetAttributeNodeNS(p0, p1) {
			return dom_wrap(dom.Attr, this.handle.getAttributeNodeNS(p0, p1));
		}

		GetBoundingClientRect() {
			return dom_wrap(dom.DOMRect, this.handle.getBoundingClientRect());
		}

		GetClientRects() {
			return dom_wrap(dom.DOMRectList, this.handle.getClientRects());
		}

		GetElementsByClassName(p0) {
			return dom_wrap(dom.HTMLCollectionOf, this.handle.getElementsByClassName(p0));
		}

		GetElementsByTagName(p0) {
			return dom_wrap(dom.HTMLCollectionOf, this.handle.getElementsByTagName(p0));
		}

		GetElementsByTagNameNS(p0, p1) {
			return dom_wrap(dom.HTMLCollectionOf, this.handle.getElementsByTagNameNS(p0, p1));
		}

		GetHTML(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return this.handle.getHTML(...p0.map((item) => dom_toGetHTMLOptions(item)));
		}

		HasAttribute(p0) {
			return this.handle.hasAttribute(p0);
		}

		HasAttributeNS(p0, p1) {
			return this.handle.hasAttributeNS(p0, p1);
		}

		HasAttributes() {
			return this.handle.hasAttributes();
		}

		HasPointerCapture(p0) {
			return this.handle.hasPointerCapture(p0);
		}

		InsertAdjacentElement(p0, p1) {
			return dom_wrap(dom.Element, this.handle.insertAdjacentElement(p0, dom_unwrap(p1)));
		}

		InsertAdjacentHTML(p0, p1) {
			this.handle.insertAdjacentHTML(p0, p1);
		}

		InsertAdjacentText(p0, p1) {
			this.handle.insertAdjacentText(p0, p1);
		}

		Matches(p0) {
			return this.handle.matches(p0);
		}

		ReleasePointerCapture(p0) {
			this.handle.releasePointerCapture(p0);
		}

		RemoveAttribute(p0) {
			this.handle.removeAttribute(p0);
		}

		RemoveAttributeNS(p0, p1) {
			this.handle.removeAttributeNS(p0, p1);
		}

		RemoveAttributeNode(p0) {
			return dom_wrap(dom.Attr, this.handle.removeAttributeNode(dom_unwrap(p0)));
		}

		*RequestFullscreen(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			const outcome = yield* dom_await(() => this.handle.requestFullscreen(...p0.map((item) => dom_toFullscreenOptions(item))));
			
			return outcome.ok ? null : outcome.error;
		}

		*RequestPointerLock(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			const outcome = yield* dom_await(() => this.handle.requestPointerLock(...p0.map((item) => dom_toPointerLockOptions(item))));
			
			return outcome.ok ? null : outcome.error;
		}

		Scroll(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.scroll(...p0.map((item) => dom_toScrollToOptions(item)));
		}

		ScrollBy(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.scrollBy(...p0.map((item) => dom_toScrollToOptions(item)));
		}

		ScrollIntoView(...p0) {
			this.handle.scrollIntoView(...p0.map((item) => dom_unwrapAny(item)));
		}

		ScrollTo(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.scrollTo(...p0.map((item) => dom_toScrollToOptions(item)));
		}

		SetAttribute(p0, p1) {
			this.handle.setAttribute(p0, p1);
		}

		SetAttributeNS(p0, p1, p2) {
			this.handle.setAttributeNS(p0, p1, p2);
		}

		SetAttributeNode(p0) {
			return dom_wrap(dom.Attr, this.handle.setAttributeNode(dom_unwrap(p0)));
		}

		SetAttributeNodeNS(p0) {
			return dom_wrap(dom.Attr, this.handle.setAttributeNodeNS(dom_unwrap(p0)));
		}

		SetHTMLUnsafe(p0) {
			this.handle.setHTMLUnsafe(p0);
		}

		SetPointerCapture(p0) {
			this.handle.setPointerCapture(p0);
		}

		ToggleAttribute(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			return this.handle.toggleAttribute(p0, ...p1.map((item) => item));
		}

		WebkitMatchesSelector(p0) {
			return this.handle.webkitMatchesSelector(p0);
		}

		Animate(p0, ...p1) {
			return dom_wrap(dom.Animation, this.handle.animate(dom_unwrapAny(p0), ...p1.map((item) => dom_unwrapAny(item))));
		}

		GetAnimations(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return Array.from(this.handle.getAnimations(...p0.map((item) => dom_toGetAnimationsOptions(item))) || [], (item) => dom_wrap(dom.Animation, item));
		}

		After(...p0) {
			this.handle.after(...p0.map((item) => dom_unwrapAny(item)));
		}

		Before(...p0) {
			this.handle.before(...p0.map((item) => dom_unwrapAny(item)));
		}

		Remove() {
			this.handle.remove();
		}

		ReplaceWith(...p0) {
			this.handle.replaceWith(...p0.map((item) => dom_unwrapAny(item)));
		}

		Append(...p0) {
			this.handle.append(...p0.map((item) => dom_unwrapAny(item)));
		}

		Prepend(...p0) {
			this.handle.prepend(...p0.map((item) => dom_unwrapAny(item)));
		}

		QuerySelector(p0) {
			return dom_wrap(dom.Element, this.handle.querySelector(p0));
		}

		QuerySelectorAll(p0) {
			return dom_wrap(dom.NodeListOf, this.handle.querySelectorAll(p0));
		}

		ReplaceChildren(...p0) {
			this.handle.replaceChildren(...p0.map((item) => dom_unwrapAny(item)));
		}

		OnFullscreenchange(p0) {
			this.handle.addEventListener("fullscreenchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnFullscreenerror(p0) {
			this.handle.addEventListener("fullscreenerror", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		Onfullscreenchange(p0) {
			return this.handle.onfullscreenchange(dom_unwrap(p0));
		}

		Onfullscreenerror(p0) {
			return this.handle.onfullscreenerror(dom_unwrap(p0));
		}

		Attributes() {
			return dom_wrap(dom.NamedNodeMap, this.handle.attributes);
		}

		ClassList() {
			return dom_wrap(dom.DOMTokenList, this.handle.classList);
		}

		ClassName() {
			return this.handle.className;
		}

		SetClassName(value) {
			this.handle.className = value;
		}

		ClientHeight() {
			return this.handle.clientHeight;
		}

		ClientLeft() {
			return this.handle.clientLeft;
		}

		ClientTop() {
			return this.handle.clientTop;
		}

		ClientWidth() {
			return this.handle.clientWidth;
		}

		Id() {
			return this.handle.id;
		}

		SetId(value) {
			this.handle.id = value;
		}

		InnerHTML() {
			return this.handle.innerHTML;
		}

		SetInnerHTML(value) {
			this.handle.innerHTML = value;
		}

		LocalName() {
			return this.handle.localName;
		}

		NamespaceURI() {
			return this.handle.namespaceURI;
		}

		OuterHTML() {
			return this.handle.outerHTML;
		}

		SetOuterHTML(value) {
			this.handle.outerHTML = value;
		}

		Part() {
			return dom_wrap(dom.DOMTokenList, this.handle.part);
		}

		Prefix() {
			return this.handle.prefix;
		}

		ScrollHeight() {
			return this.handle.scrollHeight;
		}

		ScrollLeft() {
			return this.handle.scrollLeft;
		}

		SetScrollLeft(value) {
			this.handle.scrollLeft = value;
		}

		ScrollTop() {
			return this.handle.scrollTop;
		}

		SetScrollTop(value) {
			this.handle.scrollTop = value;
		}

		ScrollWidth() {
			return this.handle.scrollWidth;
		}

		ShadowRoot() {
			return dom_wrap(dom.ShadowRoot, this.handle.shadowRoot);
		}

		Slot() {
			return this.handle.slot;
		}

		SetSlot(value) {
			this.handle.slot = value;
		}

		TagName() {
			return this.handle.tagName;
		}

		AriaAtomic() {
			return this.handle.ariaAtomic;
		}

		SetAriaAtomic(value) {
			this.handle.ariaAtomic = value;
		}

		AriaAutoComplete() {
			return this.handle.ariaAutoComplete;
		}

		SetAriaAutoComplete(value) {
			this.handle.ariaAutoComplete = value;
		}

		AriaBrailleLabel() {
			return this.handle.ariaBrailleLabel;
		}

		SetAriaBrailleLabel(value) {
			this.handle.ariaBrailleLabel = value;
		}

		AriaBrailleRoleDescription() {
			return this.handle.ariaBrailleRoleDescription;
		}

		SetAriaBrailleRoleDescription(value) {
			this.handle.ariaBrailleRoleDescription = value;
		}

		AriaBusy() {
			return this.handle.ariaBusy;
		}

		SetAriaBusy(value) {
			this.handle.ariaBusy = value;
		}

		AriaChecked() {
			return this.handle.ariaChecked;
		}

		SetAriaChecked(value) {
			this.handle.ariaChecked = value;
		}

		AriaColCount() {
			return this.handle.ariaColCount;
		}

		SetAriaColCount(value) {
			this.handle.ariaColCount = value;
		}

		AriaColIndex() {
			return this.handle.ariaColIndex;
		}

		SetAriaColIndex(value) {
			this.handle.ariaColIndex = value;
		}

		AriaColSpan() {
			return this.handle.ariaColSpan;
		}

		SetAriaColSpan(value) {
			this.handle.ariaColSpan = value;
		}

		AriaCurrent() {
			return this.handle.ariaCurrent;
		}

		SetAriaCurrent(value) {
			this.handle.ariaCurrent = value;
		}

		AriaDescription() {
			return this.handle.ariaDescription;
		}

		SetAriaDescription(value) {
			this.handle.ariaDescription = value;
		}

		AriaDisabled() {
			return this.handle.ariaDisabled;
		}

		SetAriaDisabled(value) {
			this.handle.ariaDisabled = value;
		}

		AriaExpanded() {
			return this.handle.ariaExpanded;
		}

		SetAriaExpanded(value) {
			this.handle.ariaExpanded = value;
		}

		AriaHasPopup() {
			return this.handle.ariaHasPopup;
		}

		SetAriaHasPopup(value) {
			this.handle.ariaHasPopup = value;
		}

		AriaHidden() {
			return this.handle.ariaHidden;
		}

		SetAriaHidden(value) {
			this.handle.ariaHidden = value;
		}

		AriaInvalid() {
			return this.handle.ariaInvalid;
		}

		SetAriaInvalid(value) {
			this.handle.ariaInvalid = value;
		}

		AriaKeyShortcuts() {
			return this.handle.ariaKeyShortcuts;
		}

		SetAriaKeyShortcuts(value) {
			this.handle.ariaKeyShortcuts = value;
		}

		AriaLabel() {
			return this.handle.ariaLabel;
		}

		SetAriaLabel(value) {
			this.handle.ariaLabel = value;
		}

		AriaLevel() {
			return this.handle.ariaLevel;
		}

		SetAriaLevel(value) {
			this.handle.ariaLevel = value;
		}

		AriaLive() {
			return this.handle.ariaLive;
		}

		SetAriaLive(value) {
			this.handle.ariaLive = value;
		}

		AriaModal() {
			return this.handle.ariaModal;
		}

		SetAriaModal(value) {
			this.handle.ariaModal = value;
		}

		AriaMultiLine() {
			return this.handle.ariaMultiLine;
		}

		SetAriaMultiLine(value) {
			this.handle.ariaMultiLine = value;
		}

		AriaMultiSelectable() {
			return this.handle.ariaMultiSelectable;
		}

		SetAriaMultiSelectable(value) {
			this.handle.ariaMultiSelectable = value;
		}

		AriaOrientation() {
			return this.handle.ariaOrientation;
		}

		SetAriaOrientation(value) {
			this.handle.ariaOrientation = value;
		}

		AriaPlaceholder() {
			return this.handle.ariaPlaceholder;
		}

		SetAriaPlaceholder(value) {
			this.handle.ariaPlaceholder = value;
		}

		AriaPosInSet() {
			return this.handle.ariaPosInSet;
		}

		SetAriaPosInSet(value) {
			this.handle.ariaPosInSet = value;
		}

		AriaPressed() {
			return this.handle.ariaPressed;
		}

		SetAriaPressed(value) {
			this.handle.ariaPressed = value;
		}

		AriaReadOnly() {
			return this.handle.ariaReadOnly;
		}

		SetAriaReadOnly(value) {
			this.handle.ariaReadOnly = value;
		}

		AriaRequired() {
			return this.handle.ariaRequired;
		}

		SetAriaRequired(value) {
			this.handle.ariaRequired = value;
		}

		AriaRoleDescription() {
			return this.handle.ariaRoleDescription;
		}

		SetAriaRoleDescription(value) {
			this.handle.ariaRoleDescription = value;
		}

		AriaRowCount() {
			return this.handle.ariaRowCount;
		}

		SetAriaRowCount(value) {
			this.handle.ariaRowCount = value;
		}

		AriaRowIndex() {
			return this.handle.ariaRowIndex;
		}

		SetAriaRowIndex(value) {
			this.handle.ariaRowIndex = value;
		}

		AriaRowSpan() {
			return this.handle.ariaRowSpan;
		}

		SetAriaRowSpan(value) {
			this.handle.ariaRowSpan = value;
		}

		AriaSelected() {
			return this.handle.ariaSelected;
		}

		SetAriaSelected(value) {
			this.handle.ariaSelected = value;
		}

		AriaSetSize() {
			return this.handle.ariaSetSize;
		}

		SetAriaSetSize(value) {
			this.handle.ariaSetSize = value;
		}

		AriaSort() {
			return this.handle.ariaSort;
		}

		SetAriaSort(value) {
			this.handle.ariaSort = value;
		}

		AriaValueMax() {
			return this.handle.ariaValueMax;
		}

		SetAriaValueMax(value) {
			this.handle.ariaValueMax = value;
		}

		AriaValueMin() {
			return this.handle.ariaValueMin;
		}

		SetAriaValueMin(value) {
			this.handle.ariaValueMin = value;
		}

		AriaValueNow() {
			return this.handle.ariaValueNow;
		}

		SetAriaValueNow(value) {
			this.handle.ariaValueNow = value;
		}

		AriaValueText() {
			return this.handle.ariaValueText;
		}

		SetAriaValueText(value) {
			this.handle.ariaValueText = value;
		}

		Role() {
			return this.handle.role;
		}

		SetRole(value) {
			this.handle.role = value;
		}

		NextElementSibling() {
			return dom_wrap(dom.Element, this.handle.nextElementSibling);
		}

		PreviousElementSibling() {
			return dom_wrap(dom.Element, this.handle.previousElementSibling);
		}

		ChildElementCount() {
			return this.handle.childElementCount;
		}

		Children() {
			return dom_wrap(dom.HTMLCollection, this.handle.children);
		}

		FirstElementChild() {
			return dom_wrap(dom.Element, this.handle.firstElementChild);
		}

		LastElementChild() {
			return dom_wrap(dom.Element, this.handle.lastElementChild);
		}

		AssignedSlot() {
			return dom_wrap(dom.HTMLSlotElement, this.handle.assignedSlot);
		}
}

class dom_C_ElementCSSInlineStyle extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfElementCSSInlineStyle() {
			
		}

		AsHTMLAudioElement() {
			return dom_is(this, "HTMLAudioElement") ? this : null;
		}

		AsHTMLCanvasElement() {
			return dom_is(this, "HTMLCanvasElement") ? this : null;
		}

		AsHTMLElement() {
			return dom_is(this, "HTMLElement") ? this : null;
		}

		AsHTMLFormElement() {
			return dom_is(this, "HTMLFormElement") ? this : null;
		}

		AsHTMLHeadElement() {
			return dom_is(this, "HTMLHeadElement") ? this : null;
		}

		AsHTMLImageElement() {
			return dom_is(this, "HTMLImageElement") ? this : null;
		}

		AsHTMLMediaElement() {
			return dom_is(this, "HTMLMediaElement") ? this : null;
		}

		AsHTMLSlotElement() {
			return dom_is(this, "HTMLSlotElement") ? this : null;
		}

		AsSVGElement() {
			return dom_is(this, "SVGElement") ? this : null;
		}

		AsSVGGraphicsElement() {
			return dom_is(this, "SVGGraphicsElement") ? this : null;
		}

		AsSVGSVGElement() {
			return dom_is(this, "SVGSVGElement") ? this : null;
		}

		AttributeStyleMap() {
			return dom_wrap(dom.StylePropertyMap, this.handle.attributeStyleMap);
		}

		Style() {
			return dom_wrap(dom.CSSStyleDeclaration, this.handle.style);
		}
}

class dom_C_ElementContentEditable extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfElementContentEditable() {
			
		}

		AsHTMLAudioElement() {
			return dom_is(this, "HTMLAudioElement") ? this : null;
		}

		AsHTMLCanvasElement() {
			return dom_is(this, "HTMLCanvasElement") ? this : null;
		}

		AsHTMLElement() {
			return dom_is(this, "HTMLElement") ? this : null;
		}

		AsHTMLFormElement() {
			return dom_is(this, "HTMLFormElement") ? this : null;
		}

		AsHTMLHeadElement() {
			return dom_is(this, "HTMLHeadElement") ? this : null;
		}

		AsHTMLImageElement() {
			return dom_is(this, "HTMLImageElement") ? this : null;
		}

		AsHTMLMediaElement() {
			return dom_is(this, "HTMLMediaElement") ? this : null;
		}

		AsHTMLSlotElement() {
			return dom_is(this, "HTMLSlotElement") ? this : null;
		}

		ContentEditable() {
			return this.handle.contentEditable;
		}

		SetContentEditable(value) {
			this.handle.contentEditable = value;
		}

		EnterKeyHint() {
			return this.handle.enterKeyHint;
		}

		SetEnterKeyHint(value) {
			this.handle.enterKeyHint = value;
		}

		InputMode() {
			return this.handle.inputMode;
		}

		SetInputMode(value) {
			this.handle.inputMode = value;
		}

		IsContentEditable() {
			return this.handle.isContentEditable;
		}
}

class dom_C_ElementInternals extends dom_C_ARIAMixin {

		KindOfElementInternals() {
			
		}

		KindOfARIAMixin() {
			
		}

		AsARIAMixin() {
			return this;
		}

		CheckValidity() {
			return this.handle.checkValidity();
		}

		ReportValidity() {
			return this.handle.reportValidity();
		}

		SetFormValue(p0, ...p1) {
			this.handle.setFormValue(dom_unwrapAny(p0), ...p1.map((item) => dom_unwrapAny(item)));
		}

		SetValidity(...p0) {
			this.handle.setValidity(...p0.map((item) => dom_unwrapAny(item)));
		}

		Form() {
			return dom_wrap(dom.HTMLFormElement, this.handle.form);
		}

		Labels() {
			return dom_wrap(dom.NodeList, this.handle.labels);
		}

		ShadowRoot() {
			return dom_wrap(dom.ShadowRoot, this.handle.shadowRoot);
		}

		States() {
			return dom_wrap(dom.CustomStateSet, this.handle.states);
		}

		ValidationMessage() {
			return this.handle.validationMessage;
		}

		Validity() {
			return dom_fromValidityState(this.handle.validity);
		}

		WillValidate() {
			return this.handle.willValidate;
		}
}

class dom_C_ErrorEvent extends dom_C_Event {

		KindOfErrorEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		Colno() {
			return this.handle.colno;
		}

		Error() {
			return this.handle.error;
		}

		Filename() {
			return this.handle.filename;
		}

		Lineno() {
			return this.handle.lineno;
		}

		Message() {
			return this.handle.message;
		}
}

class dom_C_EventCounts extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfEventCounts() {
			
		}

		ForEach(p0, ...p1) {
			this.handle.forEach((p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [a[0], a[1], dom_wrap(dom.EventCounts, a[2])])), ...p1.map((item) => dom_unwrapAny(item)));
		}
}

class dom_C_External extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfExternal() {
			
		}

		AddSearchProvider() {
			this.handle.AddSearchProvider();
		}

		IsSearchProviderInstalled() {
			this.handle.IsSearchProviderInstalled();
		}
}

class dom_C_FocusEvent extends dom_C_UIEvent {

		KindOfFocusEvent() {
			
		}

		KindOfUIEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsUIEvent() {
			return this;
		}

		AsEvent() {
			return this;
		}

		RelatedTarget() {
			return dom_wrap(dom.EventTarget, this.handle.relatedTarget);
		}
}

class dom_C_FontFaceSet extends dom_C_EventTarget {

		KindOfFontFaceSet() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		Check(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			return this.handle.check(p0, ...p1.map((item) => item));
		}

		*Load(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			const outcome = yield* dom_await(() => this.handle.load(p0, ...p1.map((item) => item)));
			
			return outcome.ok ? [Array.from(outcome.value || [], (item) => item), null] : [[], outcome.error];
		}

		ForEach(p0, ...p1) {
			this.handle.forEach((p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [a[0], a[1], dom_wrap(dom.FontFaceSet, a[2])])), ...p1.map((item) => dom_unwrapAny(item)));
		}

		OnLoading(p0) {
			this.handle.addEventListener("loading", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLoadingdone(p0) {
			this.handle.addEventListener("loadingdone", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLoadingerror(p0) {
			this.handle.addEventListener("loadingerror", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		Onloading(p0) {
			return this.handle.onloading(dom_unwrap(p0));
		}

		Onloadingdone(p0) {
			return this.handle.onloadingdone(dom_unwrap(p0));
		}

		Onloadingerror(p0) {
			return this.handle.onloadingerror(dom_unwrap(p0));
		}

		Status() {
			return this.handle.status;
		}
}

class dom_C_FormData extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfFormData() {
			
		}

		Append(p0, p1) {
			this.handle.append(p0, p1);
		}

		Delete(p0) {
			this.handle.delete(p0);
		}

		Get(p0) {
			return this.handle.get(p0);
		}

		GetAll(p0) {
			return Array.from(this.handle.getAll(p0) || [], (item) => item);
		}

		Has(p0) {
			return this.handle.has(p0);
		}

		Set(p0, p1) {
			this.handle.set(p0, p1);
		}

		ForEach(p0, ...p1) {
			this.handle.forEach((p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [a[0], a[1], dom_wrap(dom.FormData, a[2])])), ...p1.map((item) => dom_unwrapAny(item)));
		}
}

class dom_C_FormDataEvent extends dom_C_Event {

		KindOfFormDataEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		FormData() {
			return dom_wrap(dom.FormData, this.handle.formData);
		}
}

class dom_C_GainNode extends dom_C_AudioNode {

		KindOfGainNode() {
			
		}

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsAudioNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		Gain() {
			return dom_wrap(dom.AudioParam, this.handle.gain);
		}
}

class dom_C_GamepadEvent extends dom_C_Event {

		KindOfGamepadEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		Gamepad() {
			return dom_fromGamepad(this.handle.gamepad);
		}
}

class dom_C_Geolocation extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfGeolocation() {
			
		}

		ClearWatch(p0) {
			this.handle.clearWatch(p0);
		}

		GetCurrentPosition(p0, ...p1) {
			this.handle.getCurrentPosition((p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [a[0]])), ...p1.map((item) => dom_unwrapAny(item)));
		}

		WatchPosition(p0, ...p1) {
			return this.handle.watchPosition((p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [a[0]])), ...p1.map((item) => dom_unwrapAny(item)));
		}
}

class dom_C_HTMLAllCollection extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfHTMLAllCollection() {
			
		}

		Item(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return this.handle.item(...p0.map((item) => item));
		}

		NamedItem(p0) {
			return this.handle.namedItem(p0);
		}

		Length() {
			return this.handle.length;
		}
}

class dom_C_HTMLOrSVGElement extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfHTMLOrSVGElement() {
			
		}

		AsHTMLAudioElement() {
			return dom_is(this, "HTMLAudioElement") ? this : null;
		}

		AsHTMLCanvasElement() {
			return dom_is(this, "HTMLCanvasElement") ? this : null;
		}

		AsHTMLElement() {
			return dom_is(this, "HTMLElement") ? this : null;
		}

		AsHTMLFormElement() {
			return dom_is(this, "HTMLFormElement") ? this : null;
		}

		AsHTMLHeadElement() {
			return dom_is(this, "HTMLHeadElement") ? this : null;
		}

		AsHTMLImageElement() {
			return dom_is(this, "HTMLImageElement") ? this : null;
		}

		AsHTMLMediaElement() {
			return dom_is(this, "HTMLMediaElement") ? this : null;
		}

		AsHTMLSlotElement() {
			return dom_is(this, "HTMLSlotElement") ? this : null;
		}

		AsSVGElement() {
			return dom_is(this, "SVGElement") ? this : null;
		}

		AsSVGGraphicsElement() {
			return dom_is(this, "SVGGraphicsElement") ? this : null;
		}

		AsSVGSVGElement() {
			return dom_is(this, "SVGSVGElement") ? this : null;
		}

		Blur() {
			this.handle.blur();
		}

		Focus(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.focus(...p0.map((item) => dom_toFocusOptions(item)));
		}

		Autofocus() {
			return this.handle.autofocus;
		}

		SetAutofocus(value) {
			this.handle.autofocus = value;
		}

		Dataset() {
			return this.handle.dataset;
		}

		Nonce() {
			return this.handle.nonce;
		}

		SetNonce(value) {
			this.handle.nonce = value;
		}

		TabIndex() {
			return this.handle.tabIndex;
		}

		SetTabIndex(value) {
			this.handle.tabIndex = value;
		}
}

class dom_C_HTMLElement extends dom_C_Element {

		KindOfHTMLElement() {
			
		}

		KindOfElement() {
			
		}

		KindOfElementCSSInlineStyle() {
			
		}

		KindOfElementContentEditable() {
			
		}

		KindOfGlobalEventHandlers() {
			
		}

		KindOfHTMLOrSVGElement() {
			
		}

		KindOfNode() {
			
		}

		KindOfARIAMixin() {
			
		}

		KindOfAnimatable() {
			
		}

		KindOfChildNode() {
			
		}

		KindOfNonDocumentTypeChildNode() {
			
		}

		KindOfParentNode() {
			
		}

		KindOfSlottable() {
			
		}

		KindOfEventTarget() {
			
		}

		AsElement() {
			return this;
		}

		AsElementCSSInlineStyle() {
			return this;
		}

		AsElementContentEditable() {
			return this;
		}

		AsGlobalEventHandlers() {
			return this;
		}

		AsHTMLOrSVGElement() {
			return this;
		}

		AsNode() {
			return this;
		}

		AsARIAMixin() {
			return this;
		}

		AsAnimatable() {
			return this;
		}

		AsChildNode() {
			return this;
		}

		AsNonDocumentTypeChildNode() {
			return this;
		}

		AsParentNode() {
			return this;
		}

		AsSlottable() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		AsHTMLAudioElement() {
			return dom_is(this, "HTMLAudioElement") ? this : null;
		}

		AsHTMLCanvasElement() {
			return dom_is(this, "HTMLCanvasElement") ? this : null;
		}

		AsHTMLFormElement() {
			return dom_is(this, "HTMLFormElement") ? this : null;
		}

		AsHTMLHeadElement() {
			return dom_is(this, "HTMLHeadElement") ? this : null;
		}

		AsHTMLImageElement() {
			return dom_is(this, "HTMLImageElement") ? this : null;
		}

		AsHTMLMediaElement() {
			return dom_is(this, "HTMLMediaElement") ? this : null;
		}

		AsHTMLSlotElement() {
			return dom_is(this, "HTMLSlotElement") ? this : null;
		}

		AttachInternals() {
			return dom_wrap(dom.ElementInternals, this.handle.attachInternals());
		}

		Click() {
			this.handle.click();
		}

		HidePopover() {
			this.handle.hidePopover();
		}

		ShowPopover() {
			this.handle.showPopover();
		}

		TogglePopover(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return this.handle.togglePopover(...p0.map((item) => item));
		}

		Blur() {
			this.handle.blur();
		}

		Focus(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.focus(...p0.map((item) => dom_toFocusOptions(item)));
		}

		OnAbort(p0) {
			this.handle.addEventListener("abort", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.UIEvent, a[0])])));
		}

		OnAnimationcancel(p0) {
			this.handle.addEventListener("animationcancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationEvent, a[0])])));
		}

		OnAnimationend(p0) {
			this.handle.addEventListener("animationend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationEvent, a[0])])));
		}

		OnAnimationiteration(p0) {
			this.handle.addEventListener("animationiteration", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationEvent, a[0])])));
		}

		OnAnimationstart(p0) {
			this.handle.addEventListener("animationstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationEvent, a[0])])));
		}

		OnAuxclick(p0) {
			this.handle.addEventListener("auxclick", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnBeforeinput(p0) {
			this.handle.addEventListener("beforeinput", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.InputEvent, a[0])])));
		}

		OnBeforetoggle(p0) {
			this.handle.addEventListener("beforetoggle", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnBlur(p0) {
			this.handle.addEventListener("blur", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FocusEvent, a[0])])));
		}

		OnCancel(p0) {
			this.handle.addEventListener("cancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCanplay(p0) {
			this.handle.addEventListener("canplay", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCanplaythrough(p0) {
			this.handle.addEventListener("canplaythrough", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnChange(p0) {
			this.handle.addEventListener("change", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnClick(p0) {
			this.handle.addEventListener("click", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnClose(p0) {
			this.handle.addEventListener("close", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCompositionend(p0) {
			this.handle.addEventListener("compositionend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.CompositionEvent, a[0])])));
		}

		OnCompositionstart(p0) {
			this.handle.addEventListener("compositionstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.CompositionEvent, a[0])])));
		}

		OnCompositionupdate(p0) {
			this.handle.addEventListener("compositionupdate", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.CompositionEvent, a[0])])));
		}

		OnContextlost(p0) {
			this.handle.addEventListener("contextlost", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnContextmenu(p0) {
			this.handle.addEventListener("contextmenu", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnContextrestored(p0) {
			this.handle.addEventListener("contextrestored", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCopy(p0) {
			this.handle.addEventListener("copy", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ClipboardEvent, a[0])])));
		}

		OnCuechange(p0) {
			this.handle.addEventListener("cuechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCut(p0) {
			this.handle.addEventListener("cut", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ClipboardEvent, a[0])])));
		}

		OnDblclick(p0) {
			this.handle.addEventListener("dblclick", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnDrag(p0) {
			this.handle.addEventListener("drag", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragend(p0) {
			this.handle.addEventListener("dragend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragenter(p0) {
			this.handle.addEventListener("dragenter", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragleave(p0) {
			this.handle.addEventListener("dragleave", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragover(p0) {
			this.handle.addEventListener("dragover", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragstart(p0) {
			this.handle.addEventListener("dragstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDrop(p0) {
			this.handle.addEventListener("drop", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDurationchange(p0) {
			this.handle.addEventListener("durationchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnEmptied(p0) {
			this.handle.addEventListener("emptied", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnEnded(p0) {
			this.handle.addEventListener("ended", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnError(p0) {
			this.handle.addEventListener("error", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ErrorEvent, a[0])])));
		}

		OnFocus(p0) {
			this.handle.addEventListener("focus", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FocusEvent, a[0])])));
		}

		OnFocusin(p0) {
			this.handle.addEventListener("focusin", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FocusEvent, a[0])])));
		}

		OnFocusout(p0) {
			this.handle.addEventListener("focusout", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FocusEvent, a[0])])));
		}

		OnFormdata(p0) {
			this.handle.addEventListener("formdata", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FormDataEvent, a[0])])));
		}

		OnGotpointercapture(p0) {
			this.handle.addEventListener("gotpointercapture", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnInput(p0) {
			this.handle.addEventListener("input", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnInvalid(p0) {
			this.handle.addEventListener("invalid", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnKeydown(p0) {
			this.handle.addEventListener("keydown", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.KeyboardEvent, a[0])])));
		}

		OnKeypress(p0) {
			this.handle.addEventListener("keypress", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.KeyboardEvent, a[0])])));
		}

		OnKeyup(p0) {
			this.handle.addEventListener("keyup", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.KeyboardEvent, a[0])])));
		}

		OnLoad(p0) {
			this.handle.addEventListener("load", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLoadeddata(p0) {
			this.handle.addEventListener("loadeddata", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLoadedmetadata(p0) {
			this.handle.addEventListener("loadedmetadata", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLoadstart(p0) {
			this.handle.addEventListener("loadstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLostpointercapture(p0) {
			this.handle.addEventListener("lostpointercapture", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnMousedown(p0) {
			this.handle.addEventListener("mousedown", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseenter(p0) {
			this.handle.addEventListener("mouseenter", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseleave(p0) {
			this.handle.addEventListener("mouseleave", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMousemove(p0) {
			this.handle.addEventListener("mousemove", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseout(p0) {
			this.handle.addEventListener("mouseout", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseover(p0) {
			this.handle.addEventListener("mouseover", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseup(p0) {
			this.handle.addEventListener("mouseup", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnPaste(p0) {
			this.handle.addEventListener("paste", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ClipboardEvent, a[0])])));
		}

		OnPause(p0) {
			this.handle.addEventListener("pause", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnPlay(p0) {
			this.handle.addEventListener("play", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnPlaying(p0) {
			this.handle.addEventListener("playing", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnPointercancel(p0) {
			this.handle.addEventListener("pointercancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerdown(p0) {
			this.handle.addEventListener("pointerdown", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerenter(p0) {
			this.handle.addEventListener("pointerenter", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerleave(p0) {
			this.handle.addEventListener("pointerleave", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointermove(p0) {
			this.handle.addEventListener("pointermove", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerout(p0) {
			this.handle.addEventListener("pointerout", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerover(p0) {
			this.handle.addEventListener("pointerover", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerup(p0) {
			this.handle.addEventListener("pointerup", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnProgress(p0) {
			this.handle.addEventListener("progress", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ProgressEvent, a[0])])));
		}

		OnRatechange(p0) {
			this.handle.addEventListener("ratechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnReset(p0) {
			this.handle.addEventListener("reset", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnResize(p0) {
			this.handle.addEventListener("resize", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.UIEvent, a[0])])));
		}

		OnScroll(p0) {
			this.handle.addEventListener("scroll", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnScrollend(p0) {
			this.handle.addEventListener("scrollend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSecuritypolicyviolation(p0) {
			this.handle.addEventListener("securitypolicyviolation", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.SecurityPolicyViolationEvent, a[0])])));
		}

		OnSeeked(p0) {
			this.handle.addEventListener("seeked", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSeeking(p0) {
			this.handle.addEventListener("seeking", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSelect(p0) {
			this.handle.addEventListener("select", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSelectionchange(p0) {
			this.handle.addEventListener("selectionchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSelectstart(p0) {
			this.handle.addEventListener("selectstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSlotchange(p0) {
			this.handle.addEventListener("slotchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnStalled(p0) {
			this.handle.addEventListener("stalled", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSubmit(p0) {
			this.handle.addEventListener("submit", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.SubmitEvent, a[0])])));
		}

		OnSuspend(p0) {
			this.handle.addEventListener("suspend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnTimeupdate(p0) {
			this.handle.addEventListener("timeupdate", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnToggle(p0) {
			this.handle.addEventListener("toggle", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnTouchcancel(p0) {
			this.handle.addEventListener("touchcancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TouchEvent, a[0])])));
		}

		OnTouchend(p0) {
			this.handle.addEventListener("touchend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TouchEvent, a[0])])));
		}

		OnTouchmove(p0) {
			this.handle.addEventListener("touchmove", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TouchEvent, a[0])])));
		}

		OnTouchstart(p0) {
			this.handle.addEventListener("touchstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TouchEvent, a[0])])));
		}

		OnTransitioncancel(p0) {
			this.handle.addEventListener("transitioncancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TransitionEvent, a[0])])));
		}

		OnTransitionend(p0) {
			this.handle.addEventListener("transitionend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TransitionEvent, a[0])])));
		}

		OnTransitionrun(p0) {
			this.handle.addEventListener("transitionrun", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TransitionEvent, a[0])])));
		}

		OnTransitionstart(p0) {
			this.handle.addEventListener("transitionstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TransitionEvent, a[0])])));
		}

		OnVolumechange(p0) {
			this.handle.addEventListener("volumechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWaiting(p0) {
			this.handle.addEventListener("waiting", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWebkitanimationend(p0) {
			this.handle.addEventListener("webkitanimationend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWebkitanimationiteration(p0) {
			this.handle.addEventListener("webkitanimationiteration", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWebkitanimationstart(p0) {
			this.handle.addEventListener("webkitanimationstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWebkittransitionend(p0) {
			this.handle.addEventListener("webkittransitionend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWheel(p0) {
			this.handle.addEventListener("wheel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.WheelEvent, a[0])])));
		}

		Onabort(p0) {
			return this.handle.onabort(dom_unwrap(p0));
		}

		Onanimationcancel(p0) {
			return this.handle.onanimationcancel(dom_unwrap(p0));
		}

		Onanimationend(p0) {
			return this.handle.onanimationend(dom_unwrap(p0));
		}

		Onanimationiteration(p0) {
			return this.handle.onanimationiteration(dom_unwrap(p0));
		}

		Onanimationstart(p0) {
			return this.handle.onanimationstart(dom_unwrap(p0));
		}

		Onauxclick(p0) {
			return this.handle.onauxclick(dom_unwrap(p0));
		}

		Onbeforeinput(p0) {
			return this.handle.onbeforeinput(dom_unwrap(p0));
		}

		Onbeforetoggle(p0) {
			return this.handle.onbeforetoggle(dom_unwrap(p0));
		}

		Onblur(p0) {
			return this.handle.onblur(dom_unwrap(p0));
		}

		Oncancel(p0) {
			return this.handle.oncancel(dom_unwrap(p0));
		}

		Oncanplay(p0) {
			return this.handle.oncanplay(dom_unwrap(p0));
		}

		Oncanplaythrough(p0) {
			return this.handle.oncanplaythrough(dom_unwrap(p0));
		}

		Onchange(p0) {
			return this.handle.onchange(dom_unwrap(p0));
		}

		Onclick(p0) {
			return this.handle.onclick(dom_unwrap(p0));
		}

		Onclose(p0) {
			return this.handle.onclose(dom_unwrap(p0));
		}

		Oncontextlost(p0) {
			return this.handle.oncontextlost(dom_unwrap(p0));
		}

		Oncontextmenu(p0) {
			return this.handle.oncontextmenu(dom_unwrap(p0));
		}

		Oncontextrestored(p0) {
			return this.handle.oncontextrestored(dom_unwrap(p0));
		}

		Oncopy(p0) {
			return this.handle.oncopy(dom_unwrap(p0));
		}

		Oncuechange(p0) {
			return this.handle.oncuechange(dom_unwrap(p0));
		}

		Oncut(p0) {
			return this.handle.oncut(dom_unwrap(p0));
		}

		Ondblclick(p0) {
			return this.handle.ondblclick(dom_unwrap(p0));
		}

		Ondrag(p0) {
			return this.handle.ondrag(dom_unwrap(p0));
		}

		Ondragend(p0) {
			return this.handle.ondragend(dom_unwrap(p0));
		}

		Ondragenter(p0) {
			return this.handle.ondragenter(dom_unwrap(p0));
		}

		Ondragleave(p0) {
			return this.handle.ondragleave(dom_unwrap(p0));
		}

		Ondragover(p0) {
			return this.handle.ondragover(dom_unwrap(p0));
		}

		Ondragstart(p0) {
			return this.handle.ondragstart(dom_unwrap(p0));
		}

		Ondrop(p0) {
			return this.handle.ondrop(dom_unwrap(p0));
		}

		Ondurationchange(p0) {
			return this.handle.ondurationchange(dom_unwrap(p0));
		}

		Onemptied(p0) {
			return this.handle.onemptied(dom_unwrap(p0));
		}

		Onended(p0) {
			return this.handle.onended(dom_unwrap(p0));
		}

		Onfocus(p0) {
			return this.handle.onfocus(dom_unwrap(p0));
		}

		Onformdata(p0) {
			return this.handle.onformdata(dom_unwrap(p0));
		}

		Ongotpointercapture(p0) {
			return this.handle.ongotpointercapture(dom_unwrap(p0));
		}

		Oninput(p0) {
			return this.handle.oninput(dom_unwrap(p0));
		}

		Oninvalid(p0) {
			return this.handle.oninvalid(dom_unwrap(p0));
		}

		Onkeydown(p0) {
			return this.handle.onkeydown(dom_unwrap(p0));
		}

		Onkeypress(p0) {
			return this.handle.onkeypress(dom_unwrap(p0));
		}

		Onkeyup(p0) {
			return this.handle.onkeyup(dom_unwrap(p0));
		}

		Onload(p0) {
			return this.handle.onload(dom_unwrap(p0));
		}

		Onloadeddata(p0) {
			return this.handle.onloadeddata(dom_unwrap(p0));
		}

		Onloadedmetadata(p0) {
			return this.handle.onloadedmetadata(dom_unwrap(p0));
		}

		Onloadstart(p0) {
			return this.handle.onloadstart(dom_unwrap(p0));
		}

		Onlostpointercapture(p0) {
			return this.handle.onlostpointercapture(dom_unwrap(p0));
		}

		Onmousedown(p0) {
			return this.handle.onmousedown(dom_unwrap(p0));
		}

		Onmouseenter(p0) {
			return this.handle.onmouseenter(dom_unwrap(p0));
		}

		Onmouseleave(p0) {
			return this.handle.onmouseleave(dom_unwrap(p0));
		}

		Onmousemove(p0) {
			return this.handle.onmousemove(dom_unwrap(p0));
		}

		Onmouseout(p0) {
			return this.handle.onmouseout(dom_unwrap(p0));
		}

		Onmouseover(p0) {
			return this.handle.onmouseover(dom_unwrap(p0));
		}

		Onmouseup(p0) {
			return this.handle.onmouseup(dom_unwrap(p0));
		}

		Onpaste(p0) {
			return this.handle.onpaste(dom_unwrap(p0));
		}

		Onpause(p0) {
			return this.handle.onpause(dom_unwrap(p0));
		}

		Onplay(p0) {
			return this.handle.onplay(dom_unwrap(p0));
		}

		Onplaying(p0) {
			return this.handle.onplaying(dom_unwrap(p0));
		}

		Onpointercancel(p0) {
			return this.handle.onpointercancel(dom_unwrap(p0));
		}

		Onpointerdown(p0) {
			return this.handle.onpointerdown(dom_unwrap(p0));
		}

		Onpointerenter(p0) {
			return this.handle.onpointerenter(dom_unwrap(p0));
		}

		Onpointerleave(p0) {
			return this.handle.onpointerleave(dom_unwrap(p0));
		}

		Onpointermove(p0) {
			return this.handle.onpointermove(dom_unwrap(p0));
		}

		Onpointerout(p0) {
			return this.handle.onpointerout(dom_unwrap(p0));
		}

		Onpointerover(p0) {
			return this.handle.onpointerover(dom_unwrap(p0));
		}

		Onpointerup(p0) {
			return this.handle.onpointerup(dom_unwrap(p0));
		}

		Onprogress(p0) {
			return this.handle.onprogress(dom_unwrap(p0));
		}

		Onratechange(p0) {
			return this.handle.onratechange(dom_unwrap(p0));
		}

		Onreset(p0) {
			return this.handle.onreset(dom_unwrap(p0));
		}

		Onresize(p0) {
			return this.handle.onresize(dom_unwrap(p0));
		}

		Onscroll(p0) {
			return this.handle.onscroll(dom_unwrap(p0));
		}

		Onscrollend(p0) {
			return this.handle.onscrollend(dom_unwrap(p0));
		}

		Onsecuritypolicyviolation(p0) {
			return this.handle.onsecuritypolicyviolation(dom_unwrap(p0));
		}

		Onseeked(p0) {
			return this.handle.onseeked(dom_unwrap(p0));
		}

		Onseeking(p0) {
			return this.handle.onseeking(dom_unwrap(p0));
		}

		Onselect(p0) {
			return this.handle.onselect(dom_unwrap(p0));
		}

		Onselectionchange(p0) {
			return this.handle.onselectionchange(dom_unwrap(p0));
		}

		Onselectstart(p0) {
			return this.handle.onselectstart(dom_unwrap(p0));
		}

		Onslotchange(p0) {
			return this.handle.onslotchange(dom_unwrap(p0));
		}

		Onstalled(p0) {
			return this.handle.onstalled(dom_unwrap(p0));
		}

		Onsubmit(p0) {
			return this.handle.onsubmit(dom_unwrap(p0));
		}

		Onsuspend(p0) {
			return this.handle.onsuspend(dom_unwrap(p0));
		}

		Ontimeupdate(p0) {
			return this.handle.ontimeupdate(dom_unwrap(p0));
		}

		Ontoggle(p0) {
			return this.handle.ontoggle(dom_unwrap(p0));
		}

		Ontouchcancel(p0) {
			return this.handle.ontouchcancel(dom_unwrap(p0));
		}

		Ontouchend(p0) {
			return this.handle.ontouchend(dom_unwrap(p0));
		}

		Ontouchmove(p0) {
			return this.handle.ontouchmove(dom_unwrap(p0));
		}

		Ontouchstart(p0) {
			return this.handle.ontouchstart(dom_unwrap(p0));
		}

		Ontransitioncancel(p0) {
			return this.handle.ontransitioncancel(dom_unwrap(p0));
		}

		Ontransitionend(p0) {
			return this.handle.ontransitionend(dom_unwrap(p0));
		}

		Ontransitionrun(p0) {
			return this.handle.ontransitionrun(dom_unwrap(p0));
		}

		Ontransitionstart(p0) {
			return this.handle.ontransitionstart(dom_unwrap(p0));
		}

		Onvolumechange(p0) {
			return this.handle.onvolumechange(dom_unwrap(p0));
		}

		Onwaiting(p0) {
			return this.handle.onwaiting(dom_unwrap(p0));
		}

		Onwebkitanimationend(p0) {
			return this.handle.onwebkitanimationend(dom_unwrap(p0));
		}

		Onwebkitanimationiteration(p0) {
			return this.handle.onwebkitanimationiteration(dom_unwrap(p0));
		}

		Onwebkitanimationstart(p0) {
			return this.handle.onwebkitanimationstart(dom_unwrap(p0));
		}

		Onwebkittransitionend(p0) {
			return this.handle.onwebkittransitionend(dom_unwrap(p0));
		}

		Onwheel(p0) {
			return this.handle.onwheel(dom_unwrap(p0));
		}

		AccessKey() {
			return this.handle.accessKey;
		}

		SetAccessKey(value) {
			this.handle.accessKey = value;
		}

		AccessKeyLabel() {
			return this.handle.accessKeyLabel;
		}

		Autocapitalize() {
			return this.handle.autocapitalize;
		}

		SetAutocapitalize(value) {
			this.handle.autocapitalize = value;
		}

		Dir() {
			return this.handle.dir;
		}

		SetDir(value) {
			this.handle.dir = value;
		}

		Draggable() {
			return this.handle.draggable;
		}

		SetDraggable(value) {
			this.handle.draggable = value;
		}

		Hidden() {
			return this.handle.hidden;
		}

		SetHidden(value) {
			this.handle.hidden = value;
		}

		Inert() {
			return this.handle.inert;
		}

		SetInert(value) {
			this.handle.inert = value;
		}

		InnerText() {
			return this.handle.innerText;
		}

		SetInnerText(value) {
			this.handle.innerText = value;
		}

		Lang() {
			return this.handle.lang;
		}

		SetLang(value) {
			this.handle.lang = value;
		}

		OffsetHeight() {
			return this.handle.offsetHeight;
		}

		OffsetLeft() {
			return this.handle.offsetLeft;
		}

		OffsetParent() {
			return dom_wrap(dom.Element, this.handle.offsetParent);
		}

		OffsetTop() {
			return this.handle.offsetTop;
		}

		OffsetWidth() {
			return this.handle.offsetWidth;
		}

		OuterText() {
			return this.handle.outerText;
		}

		SetOuterText(value) {
			this.handle.outerText = value;
		}

		Popover() {
			return this.handle.popover;
		}

		SetPopover(value) {
			this.handle.popover = value;
		}

		Spellcheck() {
			return this.handle.spellcheck;
		}

		SetSpellcheck(value) {
			this.handle.spellcheck = value;
		}

		Title() {
			return this.handle.title;
		}

		SetTitle(value) {
			this.handle.title = value;
		}

		Translate() {
			return this.handle.translate;
		}

		SetTranslate(value) {
			this.handle.translate = value;
		}

		AttributeStyleMap() {
			return dom_wrap(dom.StylePropertyMap, this.handle.attributeStyleMap);
		}

		Style() {
			return dom_wrap(dom.CSSStyleDeclaration, this.handle.style);
		}

		ContentEditable() {
			return this.handle.contentEditable;
		}

		SetContentEditable(value) {
			this.handle.contentEditable = value;
		}

		EnterKeyHint() {
			return this.handle.enterKeyHint;
		}

		SetEnterKeyHint(value) {
			this.handle.enterKeyHint = value;
		}

		InputMode() {
			return this.handle.inputMode;
		}

		SetInputMode(value) {
			this.handle.inputMode = value;
		}

		IsContentEditable() {
			return this.handle.isContentEditable;
		}

		Onerror() {
			return this.handle.onerror;
		}

		SetOnerror(value) {
			this.handle.onerror = (value === null || value === undefined ? null : (...a) => dom_unwrapAny(go2jsCallNow(value, null, [a[0], a[1], a[2], a[3], a[4]])));
		}

		Autofocus() {
			return this.handle.autofocus;
		}

		SetAutofocus(value) {
			this.handle.autofocus = value;
		}

		Dataset() {
			return this.handle.dataset;
		}

		Nonce() {
			return this.handle.nonce;
		}

		SetNonce(value) {
			this.handle.nonce = value;
		}

		TabIndex() {
			return this.handle.tabIndex;
		}

		SetTabIndex(value) {
			this.handle.tabIndex = value;
		}
}

class dom_C_HTMLMediaElement extends dom_C_HTMLElement {

		KindOfHTMLMediaElement() {
			
		}

		KindOfHTMLElement() {
			
		}

		KindOfElement() {
			
		}

		KindOfElementCSSInlineStyle() {
			
		}

		KindOfElementContentEditable() {
			
		}

		KindOfGlobalEventHandlers() {
			
		}

		KindOfHTMLOrSVGElement() {
			
		}

		KindOfNode() {
			
		}

		KindOfARIAMixin() {
			
		}

		KindOfAnimatable() {
			
		}

		KindOfChildNode() {
			
		}

		KindOfNonDocumentTypeChildNode() {
			
		}

		KindOfParentNode() {
			
		}

		KindOfSlottable() {
			
		}

		KindOfEventTarget() {
			
		}

		AsHTMLElement() {
			return this;
		}

		AsElement() {
			return this;
		}

		AsElementCSSInlineStyle() {
			return this;
		}

		AsElementContentEditable() {
			return this;
		}

		AsGlobalEventHandlers() {
			return this;
		}

		AsHTMLOrSVGElement() {
			return this;
		}

		AsNode() {
			return this;
		}

		AsARIAMixin() {
			return this;
		}

		AsAnimatable() {
			return this;
		}

		AsChildNode() {
			return this;
		}

		AsNonDocumentTypeChildNode() {
			return this;
		}

		AsParentNode() {
			return this;
		}

		AsSlottable() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		AsHTMLAudioElement() {
			return dom_is(this, "HTMLAudioElement") ? this : null;
		}

		AddTextTrack(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			return dom_wrap(dom.TextTrack, this.handle.addTextTrack(p0, ...p1.map((item) => item)));
		}

		CanPlayType(p0) {
			return this.handle.canPlayType(p0);
		}

		FastSeek(p0) {
			this.handle.fastSeek(p0);
		}

		Load() {
			this.handle.load();
		}

		Pause() {
			this.handle.pause();
		}

		*Play() {
			const outcome = yield* dom_await(() => this.handle.play());
			
			return outcome.ok ? null : outcome.error;
		}

		*SetMediaKeys(p0) {
			const outcome = yield* dom_await(() => this.handle.setMediaKeys(dom_unwrap(p0)));
			
			return outcome.ok ? null : outcome.error;
		}

		*SetSinkId(p0) {
			const outcome = yield* dom_await(() => this.handle.setSinkId(p0));
			
			return outcome.ok ? null : outcome.error;
		}

		OnEncrypted(p0) {
			this.handle.addEventListener("encrypted", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MediaEncryptedEvent, a[0])])));
		}

		OnWaitingforkey(p0) {
			this.handle.addEventListener("waitingforkey", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		Onencrypted(p0) {
			return this.handle.onencrypted(dom_unwrap(p0));
		}

		Onwaitingforkey(p0) {
			return this.handle.onwaitingforkey(dom_unwrap(p0));
		}

		Autoplay() {
			return this.handle.autoplay;
		}

		SetAutoplay(value) {
			this.handle.autoplay = value;
		}

		Buffered() {
			return dom_wrap(dom.TimeRanges, this.handle.buffered);
		}

		Controls() {
			return this.handle.controls;
		}

		SetControls(value) {
			this.handle.controls = value;
		}

		CrossOrigin() {
			return this.handle.crossOrigin;
		}

		SetCrossOrigin(value) {
			this.handle.crossOrigin = value;
		}

		CurrentSrc() {
			return this.handle.currentSrc;
		}

		CurrentTime() {
			return this.handle.currentTime;
		}

		SetCurrentTime(value) {
			this.handle.currentTime = value;
		}

		DefaultMuted() {
			return this.handle.defaultMuted;
		}

		SetDefaultMuted(value) {
			this.handle.defaultMuted = value;
		}

		DefaultPlaybackRate() {
			return this.handle.defaultPlaybackRate;
		}

		SetDefaultPlaybackRate(value) {
			this.handle.defaultPlaybackRate = value;
		}

		DisableRemotePlayback() {
			return this.handle.disableRemotePlayback;
		}

		SetDisableRemotePlayback(value) {
			this.handle.disableRemotePlayback = value;
		}

		Duration() {
			return this.handle.duration;
		}

		Ended() {
			return this.handle.ended;
		}

		Error() {
			return dom_fromMediaError(this.handle.error);
		}

		Loop() {
			return this.handle.loop;
		}

		SetLoop(value) {
			this.handle.loop = value;
		}

		MediaKeys() {
			return dom_wrap(dom.MediaKeys, this.handle.mediaKeys);
		}

		Muted() {
			return this.handle.muted;
		}

		SetMuted(value) {
			this.handle.muted = value;
		}

		NetworkState() {
			return this.handle.networkState;
		}

		Paused() {
			return this.handle.paused;
		}

		PlaybackRate() {
			return this.handle.playbackRate;
		}

		SetPlaybackRate(value) {
			this.handle.playbackRate = value;
		}

		Played() {
			return dom_wrap(dom.TimeRanges, this.handle.played);
		}

		Preload() {
			return this.handle.preload;
		}

		SetPreload(value) {
			this.handle.preload = value;
		}

		PreservesPitch() {
			return this.handle.preservesPitch;
		}

		SetPreservesPitch(value) {
			this.handle.preservesPitch = value;
		}

		ReadyState() {
			return this.handle.readyState;
		}

		Remote() {
			return dom_wrap(dom.RemotePlayback, this.handle.remote);
		}

		Seekable() {
			return dom_wrap(dom.TimeRanges, this.handle.seekable);
		}

		Seeking() {
			return this.handle.seeking;
		}

		SinkId() {
			return this.handle.sinkId;
		}

		Src() {
			return this.handle.src;
		}

		SetSrc(value) {
			this.handle.src = value;
		}

		SrcObject() {
			return this.handle.srcObject;
		}

		SetSrcObject(value) {
			this.handle.srcObject = dom_unwrapAny(value);
		}

		TextTracks() {
			return dom_wrap(dom.TextTrackList, this.handle.textTracks);
		}

		Volume() {
			return this.handle.volume;
		}

		SetVolume(value) {
			this.handle.volume = value;
		}

		NETWORK_EMPTY() {
			return this.handle.NETWORK_EMPTY;
		}

		NETWORK_IDLE() {
			return this.handle.NETWORK_IDLE;
		}

		NETWORK_LOADING() {
			return this.handle.NETWORK_LOADING;
		}

		NETWORK_NO_SOURCE() {
			return this.handle.NETWORK_NO_SOURCE;
		}

		HAVE_NOTHING() {
			return this.handle.HAVE_NOTHING;
		}

		HAVE_METADATA() {
			return this.handle.HAVE_METADATA;
		}

		HAVE_CURRENT_DATA() {
			return this.handle.HAVE_CURRENT_DATA;
		}

		HAVE_FUTURE_DATA() {
			return this.handle.HAVE_FUTURE_DATA;
		}

		HAVE_ENOUGH_DATA() {
			return this.handle.HAVE_ENOUGH_DATA;
		}
}

class dom_C_HTMLAudioElement extends dom_C_HTMLMediaElement {

		KindOfHTMLAudioElement() {
			
		}

		KindOfHTMLMediaElement() {
			
		}

		KindOfHTMLElement() {
			
		}

		KindOfElement() {
			
		}

		KindOfElementCSSInlineStyle() {
			
		}

		KindOfElementContentEditable() {
			
		}

		KindOfGlobalEventHandlers() {
			
		}

		KindOfHTMLOrSVGElement() {
			
		}

		KindOfNode() {
			
		}

		KindOfARIAMixin() {
			
		}

		KindOfAnimatable() {
			
		}

		KindOfChildNode() {
			
		}

		KindOfNonDocumentTypeChildNode() {
			
		}

		KindOfParentNode() {
			
		}

		KindOfSlottable() {
			
		}

		KindOfEventTarget() {
			
		}

		AsHTMLMediaElement() {
			return this;
		}

		AsHTMLElement() {
			return this;
		}

		AsElement() {
			return this;
		}

		AsElementCSSInlineStyle() {
			return this;
		}

		AsElementContentEditable() {
			return this;
		}

		AsGlobalEventHandlers() {
			return this;
		}

		AsHTMLOrSVGElement() {
			return this;
		}

		AsNode() {
			return this;
		}

		AsARIAMixin() {
			return this;
		}

		AsAnimatable() {
			return this;
		}

		AsChildNode() {
			return this;
		}

		AsNonDocumentTypeChildNode() {
			return this;
		}

		AsParentNode() {
			return this;
		}

		AsSlottable() {
			return this;
		}

		AsEventTarget() {
			return this;
		}
}

class dom_C_HTMLAudioElementType extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfHTMLAudioElementType() {
			
		}

		Prototype() {
			return dom_wrap(dom.HTMLAudioElement, this.handle.prototype);
		}

		SetPrototype(value) {
			this.handle.prototype = dom_unwrap(value);
		}
}

class dom_C_HTMLCanvasElement extends dom_C_HTMLElement {

		KindOfHTMLCanvasElement() {
			
		}

		KindOfHTMLElement() {
			
		}

		KindOfElement() {
			
		}

		KindOfElementCSSInlineStyle() {
			
		}

		KindOfElementContentEditable() {
			
		}

		KindOfGlobalEventHandlers() {
			
		}

		KindOfHTMLOrSVGElement() {
			
		}

		KindOfNode() {
			
		}

		KindOfARIAMixin() {
			
		}

		KindOfAnimatable() {
			
		}

		KindOfChildNode() {
			
		}

		KindOfNonDocumentTypeChildNode() {
			
		}

		KindOfParentNode() {
			
		}

		KindOfSlottable() {
			
		}

		KindOfEventTarget() {
			
		}

		AsHTMLElement() {
			return this;
		}

		AsElement() {
			return this;
		}

		AsElementCSSInlineStyle() {
			return this;
		}

		AsElementContentEditable() {
			return this;
		}

		AsGlobalEventHandlers() {
			return this;
		}

		AsHTMLOrSVGElement() {
			return this;
		}

		AsNode() {
			return this;
		}

		AsARIAMixin() {
			return this;
		}

		AsAnimatable() {
			return this;
		}

		AsChildNode() {
			return this;
		}

		AsNonDocumentTypeChildNode() {
			return this;
		}

		AsParentNode() {
			return this;
		}

		AsSlottable() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		CaptureStream(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.MediaStream, this.handle.captureStream(...p0.map((item) => item)));
		}

		GetContext(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			return dom_wrap(dom.CanvasRenderingContext2D, this.handle.getContext(p0, ...p1.map((item) => dom_toCanvasRenderingContext2DSettings(item))));
		}

		ToBlob(p0, ...p1) {
			this.handle.toBlob((p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Blob, a[0])])), ...p1.map((item) => dom_unwrapAny(item)));
		}

		ToDataURL(...p0) {
			return this.handle.toDataURL(...p0.map((item) => dom_unwrapAny(item)));
		}

		TransferControlToOffscreen() {
			return dom_wrap(dom.OffscreenCanvas, this.handle.transferControlToOffscreen());
		}

		Height() {
			return this.handle.height;
		}

		SetHeight(value) {
			this.handle.height = value;
		}

		Width() {
			return this.handle.width;
		}

		SetWidth(value) {
			this.handle.width = value;
		}
}

class dom_C_HTMLCanvasElementType extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfHTMLCanvasElementType() {
			
		}

		Prototype() {
			return dom_wrap(dom.HTMLCanvasElement, this.handle.prototype);
		}

		SetPrototype(value) {
			this.handle.prototype = dom_unwrap(value);
		}
}

class dom_C_HTMLCollectionBase extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfHTMLCollectionBase() {
			
		}

		AsHTMLCollection() {
			return dom_is(this, "HTMLCollection") ? this : null;
		}

		AsHTMLCollectionOf() {
			return dom_is(this, "HTMLCollectionOf") ? this : null;
		}

		Item(p0) {
			return dom_wrap(dom.Element, this.handle.item(p0));
		}

		Length() {
			return this.handle.length;
		}
}

class dom_C_HTMLCollection extends dom_C_HTMLCollectionBase {

		KindOfHTMLCollection() {
			
		}

		KindOfHTMLCollectionBase() {
			
		}

		AsHTMLCollectionBase() {
			return this;
		}

		NamedItem(p0) {
			return dom_wrap(dom.Element, this.handle.namedItem(p0));
		}
}

class dom_C_HTMLCollectionOf extends dom_C_HTMLCollectionBase {

		KindOfHTMLCollectionOf() {
			
		}

		KindOfHTMLCollectionBase() {
			
		}

		AsHTMLCollectionBase() {
			return this;
		}

		NamedItem(p0) {
			return dom_wrap(dom.Element, this.handle.namedItem(p0));
		}
}

class dom_C_HTMLFormElement extends dom_C_HTMLElement {

		KindOfHTMLFormElement() {
			
		}

		KindOfHTMLElement() {
			
		}

		KindOfElement() {
			
		}

		KindOfElementCSSInlineStyle() {
			
		}

		KindOfElementContentEditable() {
			
		}

		KindOfGlobalEventHandlers() {
			
		}

		KindOfHTMLOrSVGElement() {
			
		}

		KindOfNode() {
			
		}

		KindOfARIAMixin() {
			
		}

		KindOfAnimatable() {
			
		}

		KindOfChildNode() {
			
		}

		KindOfNonDocumentTypeChildNode() {
			
		}

		KindOfParentNode() {
			
		}

		KindOfSlottable() {
			
		}

		KindOfEventTarget() {
			
		}

		AsHTMLElement() {
			return this;
		}

		AsElement() {
			return this;
		}

		AsElementCSSInlineStyle() {
			return this;
		}

		AsElementContentEditable() {
			return this;
		}

		AsGlobalEventHandlers() {
			return this;
		}

		AsHTMLOrSVGElement() {
			return this;
		}

		AsNode() {
			return this;
		}

		AsARIAMixin() {
			return this;
		}

		AsAnimatable() {
			return this;
		}

		AsChildNode() {
			return this;
		}

		AsNonDocumentTypeChildNode() {
			return this;
		}

		AsParentNode() {
			return this;
		}

		AsSlottable() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		CheckValidity() {
			return this.handle.checkValidity();
		}

		ReportValidity() {
			return this.handle.reportValidity();
		}

		RequestSubmit(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.requestSubmit(...p0.map((item) => dom_unwrap(item)));
		}

		Reset() {
			this.handle.reset();
		}

		Submit() {
			this.handle.submit();
		}

		AcceptCharset() {
			return this.handle.acceptCharset;
		}

		SetAcceptCharset(value) {
			this.handle.acceptCharset = value;
		}

		Action() {
			return this.handle.action;
		}

		SetAction(value) {
			this.handle.action = value;
		}

		Autocomplete() {
			return this.handle.autocomplete;
		}

		SetAutocomplete(value) {
			this.handle.autocomplete = value;
		}

		Elements() {
			return this.handle.elements;
		}

		Encoding() {
			return this.handle.encoding;
		}

		SetEncoding(value) {
			this.handle.encoding = value;
		}

		Enctype() {
			return this.handle.enctype;
		}

		SetEnctype(value) {
			this.handle.enctype = value;
		}

		Length() {
			return this.handle.length;
		}

		Method() {
			return this.handle.method;
		}

		SetMethod(value) {
			this.handle.method = value;
		}

		Name() {
			return this.handle.name;
		}

		SetName(value) {
			this.handle.name = value;
		}

		NoValidate() {
			return this.handle.noValidate;
		}

		SetNoValidate(value) {
			this.handle.noValidate = value;
		}

		Rel() {
			return this.handle.rel;
		}

		SetRel(value) {
			this.handle.rel = value;
		}

		RelList() {
			return dom_wrap(dom.DOMTokenList, this.handle.relList);
		}

		Target() {
			return this.handle.target;
		}

		SetTarget(value) {
			this.handle.target = value;
		}
}

class dom_C_HTMLHeadElement extends dom_C_HTMLElement {

		KindOfHTMLHeadElement() {
			
		}

		KindOfHTMLElement() {
			
		}

		KindOfElement() {
			
		}

		KindOfElementCSSInlineStyle() {
			
		}

		KindOfElementContentEditable() {
			
		}

		KindOfGlobalEventHandlers() {
			
		}

		KindOfHTMLOrSVGElement() {
			
		}

		KindOfNode() {
			
		}

		KindOfARIAMixin() {
			
		}

		KindOfAnimatable() {
			
		}

		KindOfChildNode() {
			
		}

		KindOfNonDocumentTypeChildNode() {
			
		}

		KindOfParentNode() {
			
		}

		KindOfSlottable() {
			
		}

		KindOfEventTarget() {
			
		}

		AsHTMLElement() {
			return this;
		}

		AsElement() {
			return this;
		}

		AsElementCSSInlineStyle() {
			return this;
		}

		AsElementContentEditable() {
			return this;
		}

		AsGlobalEventHandlers() {
			return this;
		}

		AsHTMLOrSVGElement() {
			return this;
		}

		AsNode() {
			return this;
		}

		AsARIAMixin() {
			return this;
		}

		AsAnimatable() {
			return this;
		}

		AsChildNode() {
			return this;
		}

		AsNonDocumentTypeChildNode() {
			return this;
		}

		AsParentNode() {
			return this;
		}

		AsSlottable() {
			return this;
		}

		AsEventTarget() {
			return this;
		}
}

class dom_C_HTMLImageElement extends dom_C_HTMLElement {

		KindOfHTMLImageElement() {
			
		}

		KindOfHTMLElement() {
			
		}

		KindOfElement() {
			
		}

		KindOfElementCSSInlineStyle() {
			
		}

		KindOfElementContentEditable() {
			
		}

		KindOfGlobalEventHandlers() {
			
		}

		KindOfHTMLOrSVGElement() {
			
		}

		KindOfNode() {
			
		}

		KindOfARIAMixin() {
			
		}

		KindOfAnimatable() {
			
		}

		KindOfChildNode() {
			
		}

		KindOfNonDocumentTypeChildNode() {
			
		}

		KindOfParentNode() {
			
		}

		KindOfSlottable() {
			
		}

		KindOfEventTarget() {
			
		}

		AsHTMLElement() {
			return this;
		}

		AsElement() {
			return this;
		}

		AsElementCSSInlineStyle() {
			return this;
		}

		AsElementContentEditable() {
			return this;
		}

		AsGlobalEventHandlers() {
			return this;
		}

		AsHTMLOrSVGElement() {
			return this;
		}

		AsNode() {
			return this;
		}

		AsARIAMixin() {
			return this;
		}

		AsAnimatable() {
			return this;
		}

		AsChildNode() {
			return this;
		}

		AsNonDocumentTypeChildNode() {
			return this;
		}

		AsParentNode() {
			return this;
		}

		AsSlottable() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		*Decode() {
			const outcome = yield* dom_await(() => this.handle.decode());
			
			return outcome.ok ? null : outcome.error;
		}

		Align() {
			return this.handle.align;
		}

		SetAlign(value) {
			this.handle.align = value;
		}

		Alt() {
			return this.handle.alt;
		}

		SetAlt(value) {
			this.handle.alt = value;
		}

		Border() {
			return this.handle.border;
		}

		SetBorder(value) {
			this.handle.border = value;
		}

		Complete() {
			return this.handle.complete;
		}

		CrossOrigin() {
			return this.handle.crossOrigin;
		}

		SetCrossOrigin(value) {
			this.handle.crossOrigin = value;
		}

		CurrentSrc() {
			return this.handle.currentSrc;
		}

		Decoding() {
			return this.handle.decoding;
		}

		SetDecoding(value) {
			this.handle.decoding = value;
		}

		FetchPriority() {
			return this.handle.fetchPriority;
		}

		SetFetchPriority(value) {
			this.handle.fetchPriority = value;
		}

		Height() {
			return this.handle.height;
		}

		SetHeight(value) {
			this.handle.height = value;
		}

		Hspace() {
			return this.handle.hspace;
		}

		SetHspace(value) {
			this.handle.hspace = value;
		}

		IsMap() {
			return this.handle.isMap;
		}

		SetIsMap(value) {
			this.handle.isMap = value;
		}

		Loading() {
			return this.handle.loading;
		}

		SetLoading(value) {
			this.handle.loading = value;
		}

		LongDesc() {
			return this.handle.longDesc;
		}

		SetLongDesc(value) {
			this.handle.longDesc = value;
		}

		Lowsrc() {
			return this.handle.lowsrc;
		}

		SetLowsrc(value) {
			this.handle.lowsrc = value;
		}

		Name() {
			return this.handle.name;
		}

		SetName(value) {
			this.handle.name = value;
		}

		NaturalHeight() {
			return this.handle.naturalHeight;
		}

		NaturalWidth() {
			return this.handle.naturalWidth;
		}

		ReferrerPolicy() {
			return this.handle.referrerPolicy;
		}

		SetReferrerPolicy(value) {
			this.handle.referrerPolicy = value;
		}

		Sizes() {
			return this.handle.sizes;
		}

		SetSizes(value) {
			this.handle.sizes = value;
		}

		Src() {
			return this.handle.src;
		}

		SetSrc(value) {
			this.handle.src = value;
		}

		Srcset() {
			return this.handle.srcset;
		}

		SetSrcset(value) {
			this.handle.srcset = value;
		}

		UseMap() {
			return this.handle.useMap;
		}

		SetUseMap(value) {
			this.handle.useMap = value;
		}

		Vspace() {
			return this.handle.vspace;
		}

		SetVspace(value) {
			this.handle.vspace = value;
		}

		Width() {
			return this.handle.width;
		}

		SetWidth(value) {
			this.handle.width = value;
		}

		X() {
			return this.handle.x;
		}

		Y() {
			return this.handle.y;
		}
}

class dom_C_HTMLImageElementType extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfHTMLImageElementType() {
			
		}

		Prototype() {
			return dom_wrap(dom.HTMLImageElement, this.handle.prototype);
		}

		SetPrototype(value) {
			this.handle.prototype = dom_unwrap(value);
		}
}

class dom_C_HTMLSlotElement extends dom_C_HTMLElement {

		KindOfHTMLSlotElement() {
			
		}

		KindOfHTMLElement() {
			
		}

		KindOfElement() {
			
		}

		KindOfElementCSSInlineStyle() {
			
		}

		KindOfElementContentEditable() {
			
		}

		KindOfGlobalEventHandlers() {
			
		}

		KindOfHTMLOrSVGElement() {
			
		}

		KindOfNode() {
			
		}

		KindOfARIAMixin() {
			
		}

		KindOfAnimatable() {
			
		}

		KindOfChildNode() {
			
		}

		KindOfNonDocumentTypeChildNode() {
			
		}

		KindOfParentNode() {
			
		}

		KindOfSlottable() {
			
		}

		KindOfEventTarget() {
			
		}

		AsHTMLElement() {
			return this;
		}

		AsElement() {
			return this;
		}

		AsElementCSSInlineStyle() {
			return this;
		}

		AsElementContentEditable() {
			return this;
		}

		AsGlobalEventHandlers() {
			return this;
		}

		AsHTMLOrSVGElement() {
			return this;
		}

		AsNode() {
			return this;
		}

		AsARIAMixin() {
			return this;
		}

		AsAnimatable() {
			return this;
		}

		AsChildNode() {
			return this;
		}

		AsNonDocumentTypeChildNode() {
			return this;
		}

		AsParentNode() {
			return this;
		}

		AsSlottable() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		Assign(...p0) {
			this.handle.assign(...p0.map((item) => dom_unwrapAny(item)));
		}

		AssignedElements(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return Array.from(this.handle.assignedElements(...p0.map((item) => dom_toAssignedNodesOptions(item))) || [], (item) => dom_wrap(dom.Element, item));
		}

		AssignedNodes(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return Array.from(this.handle.assignedNodes(...p0.map((item) => dom_toAssignedNodesOptions(item))) || [], (item) => dom_wrap(dom.Node, item));
		}

		Name() {
			return this.handle.name;
		}

		SetName(value) {
			this.handle.name = value;
		}
}

class dom_C_HashChangeEvent extends dom_C_Event {

		KindOfHashChangeEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		NewURL() {
			return this.handle.newURL;
		}

		OldURL() {
			return this.handle.oldURL;
		}
}

class dom_C_Headers extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfHeaders() {
			
		}

		Append(p0, p1) {
			this.handle.append(p0, p1);
		}

		Delete(p0) {
			this.handle.delete(p0);
		}

		Get(p0) {
			return this.handle.get(p0);
		}

		GetSetCookie() {
			return Array.from(this.handle.getSetCookie() || []);
		}

		Has(p0) {
			return this.handle.has(p0);
		}

		Set(p0, p1) {
			this.handle.set(p0, p1);
		}

		ForEach(p0, ...p1) {
			this.handle.forEach((p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [a[0], a[1], dom_wrap(dom.Headers, a[2])])), ...p1.map((item) => dom_unwrapAny(item)));
		}
}

class dom_C_History extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfHistory() {
			
		}

		Back() {
			this.handle.back();
		}

		Forward() {
			this.handle.forward();
		}

		Go(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.go(...p0.map((item) => item));
		}

		PushState(p0, p1, ...p2) {
			this.handle.pushState(dom_unwrapAny(p0), p1, ...p2.map((item) => dom_unwrapAny(item)));
		}

		ReplaceState(p0, p1, ...p2) {
			this.handle.replaceState(dom_unwrapAny(p0), p1, ...p2.map((item) => dom_unwrapAny(item)));
		}

		Length() {
			return this.handle.length;
		}

		ScrollRestoration() {
			return this.handle.scrollRestoration;
		}

		SetScrollRestoration(value) {
			this.handle.scrollRestoration = value;
		}

		State() {
			return this.handle.state;
		}
}

class dom_C_IDBFactory extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfIDBFactory() {
			
		}

		Cmp(p0, p1) {
			return this.handle.cmp(dom_unwrapAny(p0), dom_unwrapAny(p1));
		}

		*Databases() {
			const outcome = yield* dom_await(() => this.handle.databases());
			
			return outcome.ok ? [Array.from(outcome.value || [], (item) => item), null] : [[], outcome.error];
		}

		DeleteDatabase(p0) {
			return this.handle.deleteDatabase(p0);
		}

		Open(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			return this.handle.open(p0, ...p1.map((item) => item));
		}
}

class dom_C_IIRFilterNode extends dom_C_AudioNode {

		KindOfIIRFilterNode() {
			
		}

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsAudioNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		GetFrequencyResponse(p0, p1, p2) {
			this.handle.getFrequencyResponse(Float32Array.from(p0 || []), Float32Array.from(p1 || []), Float32Array.from(p2 || []));
		}
}

class dom_C_IdleDeadline extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfIdleDeadline() {
			
		}

		TimeRemaining() {
			return this.handle.timeRemaining();
		}

		DidTimeout() {
			return this.handle.didTimeout;
		}
}

class dom_C_ImageBitmap extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfImageBitmap() {
			
		}

		Close() {
			this.handle.close();
		}

		Height() {
			return this.handle.height;
		}

		Width() {
			return this.handle.width;
		}
}

class dom_C_ImageDataType extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfImageDataType() {
			
		}

		Prototype() {
			return dom_fromImageData(this.handle.prototype);
		}

		SetPrototype(value) {
			this.handle.prototype = dom_toImageData(value);
		}
}

class dom_C_InputEvent extends dom_C_UIEvent {

		KindOfInputEvent() {
			
		}

		KindOfUIEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsUIEvent() {
			return this;
		}

		AsEvent() {
			return this;
		}

		GetTargetRanges() {
			return Array.from(this.handle.getTargetRanges() || [], (item) => dom_fromStaticRange(item));
		}

		Data() {
			return this.handle.data;
		}

		DataTransfer() {
			return dom_wrap(dom.DataTransfer, this.handle.dataTransfer);
		}

		InputType() {
			return this.handle.inputType;
		}

		IsComposing() {
			return this.handle.isComposing;
		}
}

class dom_C_KeyboardEvent extends dom_C_UIEvent {

		KindOfKeyboardEvent() {
			
		}

		KindOfUIEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsUIEvent() {
			return this;
		}

		AsEvent() {
			return this;
		}

		GetModifierState(p0) {
			return this.handle.getModifierState(p0);
		}

		InitKeyboardEvent(p0, ...p1) {
			this.handle.initKeyboardEvent(p0, ...p1.map((item) => dom_unwrapAny(item)));
		}

		AltKey() {
			return this.handle.altKey;
		}

		CharCode() {
			return this.handle.charCode;
		}

		Code() {
			return this.handle.code;
		}

		CtrlKey() {
			return this.handle.ctrlKey;
		}

		IsComposing() {
			return this.handle.isComposing;
		}

		Key() {
			return this.handle.key;
		}

		KeyCode() {
			return this.handle.keyCode;
		}

		Location() {
			return this.handle.location;
		}

		MetaKey() {
			return this.handle.metaKey;
		}

		Repeat() {
			return this.handle.repeat;
		}

		ShiftKey() {
			return this.handle.shiftKey;
		}

		DOM_KEY_LOCATION_STANDARD() {
			return this.handle.DOM_KEY_LOCATION_STANDARD;
		}

		DOM_KEY_LOCATION_LEFT() {
			return this.handle.DOM_KEY_LOCATION_LEFT;
		}

		DOM_KEY_LOCATION_RIGHT() {
			return this.handle.DOM_KEY_LOCATION_RIGHT;
		}

		DOM_KEY_LOCATION_NUMPAD() {
			return this.handle.DOM_KEY_LOCATION_NUMPAD;
		}
}

class dom_C_KeyboardEventType extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfKeyboardEventType() {
			
		}

		Prototype() {
			return dom_wrap(dom.KeyboardEvent, this.handle.prototype);
		}

		SetPrototype(value) {
			this.handle.prototype = dom_unwrap(value);
		}

		DOM_KEY_LOCATION_STANDARD() {
			return this.handle.DOM_KEY_LOCATION_STANDARD;
		}

		DOM_KEY_LOCATION_LEFT() {
			return this.handle.DOM_KEY_LOCATION_LEFT;
		}

		DOM_KEY_LOCATION_RIGHT() {
			return this.handle.DOM_KEY_LOCATION_RIGHT;
		}

		DOM_KEY_LOCATION_NUMPAD() {
			return this.handle.DOM_KEY_LOCATION_NUMPAD;
		}
}

class dom_C_LinkStyle extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfLinkStyle() {
			
		}

		AsProcessingInstruction() {
			return dom_is(this, "ProcessingInstruction") ? this : null;
		}

		Sheet() {
			return dom_wrap(dom.CSSStyleSheet, this.handle.sheet);
		}
}

class dom_C_Location extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfLocation() {
			
		}

		ToString() {
			return this.handle.toString();
		}

		Assign(p0) {
			this.handle.assign(dom_unwrapAny(p0));
		}

		Reload() {
			this.handle.reload();
		}

		Replace(p0) {
			this.handle.replace(dom_unwrapAny(p0));
		}

		AncestorOrigins() {
			return dom_wrap(dom.DOMStringList, this.handle.ancestorOrigins);
		}

		Hash() {
			return this.handle.hash;
		}

		SetHash(value) {
			this.handle.hash = value;
		}

		Host() {
			return this.handle.host;
		}

		SetHost(value) {
			this.handle.host = value;
		}

		Hostname() {
			return this.handle.hostname;
		}

		SetHostname(value) {
			this.handle.hostname = value;
		}

		Href() {
			return this.handle.href;
		}

		SetHref(value) {
			this.handle.href = value;
		}

		Origin() {
			return this.handle.origin;
		}

		Pathname() {
			return this.handle.pathname;
		}

		SetPathname(value) {
			this.handle.pathname = value;
		}

		Port() {
			return this.handle.port;
		}

		SetPort(value) {
			this.handle.port = value;
		}

		Protocol() {
			return this.handle.protocol;
		}

		SetProtocol(value) {
			this.handle.protocol = value;
		}

		Search() {
			return this.handle.search;
		}

		SetSearch(value) {
			this.handle.search = value;
		}
}

class dom_C_LockManager extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfLockManager() {
			
		}

		*Query() {
			const outcome = yield* dom_await(() => this.handle.query());
			
			return outcome.ok ? [outcome.value, null] : [null, outcome.error];
		}

		*Request(p0, p1) {
			const outcome = yield* dom_await(() => this.handle.request(p0, (p1 === null || p1 === undefined ? null : (...a) => dom_unwrapAny(go2jsCallNow(p1, null, [a[0]])))));
			
			return outcome.ok ? [outcome.value, null] : [null, outcome.error];
		}
}

class dom_C_MIDIAccess extends dom_C_EventTarget {

		KindOfMIDIAccess() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		OnStatechange(p0) {
			this.handle.addEventListener("statechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		Onstatechange(p0) {
			return this.handle.onstatechange(dom_unwrap(p0));
		}

		Inputs() {
			return this.handle.inputs;
		}

		Outputs() {
			return this.handle.outputs;
		}

		SysexEnabled() {
			return this.handle.sysexEnabled;
		}
}

class dom_C_MediaCapabilities extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfMediaCapabilities() {
			
		}

		*DecodingInfo(p0) {
			const outcome = yield* dom_await(() => this.handle.decodingInfo(dom_unwrapAny(p0)));
			
			return outcome.ok ? [outcome.value, null] : [null, outcome.error];
		}

		*EncodingInfo(p0) {
			const outcome = yield* dom_await(() => this.handle.encodingInfo(dom_unwrapAny(p0)));
			
			return outcome.ok ? [outcome.value, null] : [null, outcome.error];
		}
}

class dom_C_MediaDevices extends dom_C_EventTarget {

		KindOfMediaDevices() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		*EnumerateDevices() {
			const outcome = yield* dom_await(() => this.handle.enumerateDevices());
			
			return outcome.ok ? [Array.from(outcome.value || [], (item) => item), null] : [[], outcome.error];
		}

		*GetDisplayMedia(...p0) {
			const outcome = yield* dom_await(() => this.handle.getDisplayMedia(...p0.map((item) => dom_unwrapAny(item))));
			
			return outcome.ok ? [dom_wrap(dom.MediaStream, outcome.value), null] : [null, outcome.error];
		}

		GetSupportedConstraints() {
			return this.handle.getSupportedConstraints();
		}

		*GetUserMedia(...p0) {
			const outcome = yield* dom_await(() => this.handle.getUserMedia(...p0.map((item) => dom_unwrapAny(item))));
			
			return outcome.ok ? [dom_wrap(dom.MediaStream, outcome.value), null] : [null, outcome.error];
		}

		OnDevicechange(p0) {
			this.handle.addEventListener("devicechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		Ondevicechange(p0) {
			return this.handle.ondevicechange(dom_unwrap(p0));
		}
}

class dom_C_MediaElementAudioSourceNode extends dom_C_AudioNode {

		KindOfMediaElementAudioSourceNode() {
			
		}

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsAudioNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		MediaElement() {
			return dom_wrap(dom.HTMLMediaElement, this.handle.mediaElement);
		}
}

class dom_C_MediaEncryptedEvent extends dom_C_Event {

		KindOfMediaEncryptedEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		InitData() {
			return Array.from(this.handle.initData || []);
		}

		InitDataType() {
			return this.handle.initDataType;
		}
}

class dom_C_MediaKeySession extends dom_C_EventTarget {

		KindOfMediaKeySession() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		*Close() {
			const outcome = yield* dom_await(() => this.handle.close());
			
			return outcome.ok ? null : outcome.error;
		}

		*GenerateRequest(p0, p1) {
			const outcome = yield* dom_await(() => this.handle.generateRequest(p0, dom_unwrapAny(p1)));
			
			return outcome.ok ? null : outcome.error;
		}

		*Load(p0) {
			const outcome = yield* dom_await(() => this.handle.load(p0));
			
			return outcome.ok ? [outcome.value, null] : [false, outcome.error];
		}

		*Remove() {
			const outcome = yield* dom_await(() => this.handle.remove());
			
			return outcome.ok ? null : outcome.error;
		}

		*Update(p0) {
			const outcome = yield* dom_await(() => this.handle.update(dom_unwrapAny(p0)));
			
			return outcome.ok ? null : outcome.error;
		}

		OnKeystatuseschange(p0) {
			this.handle.addEventListener("keystatuseschange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnMessage(p0) {
			this.handle.addEventListener("message", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [a[0]])));
		}

		Onkeystatuseschange(p0) {
			return this.handle.onkeystatuseschange(dom_unwrap(p0));
		}

		Onmessage(p0) {
			return this.handle.onmessage(dom_unwrapAny(p0));
		}

		Expiration() {
			return this.handle.expiration;
		}

		KeyStatuses() {
			return this.handle.keyStatuses;
		}

		SessionId() {
			return this.handle.sessionId;
		}
}

class dom_C_MediaKeySessionType extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfMediaKeySessionType() {
			
		}

		Prototype() {
			return dom_wrap(dom.MediaKeySession, this.handle.prototype);
		}

		SetPrototype(value) {
			this.handle.prototype = dom_unwrap(value);
		}
}

class dom_C_MediaKeySystemAccess extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfMediaKeySystemAccess() {
			
		}

		*CreateMediaKeys() {
			const outcome = yield* dom_await(() => this.handle.createMediaKeys());
			
			return outcome.ok ? [dom_wrap(dom.MediaKeys, outcome.value), null] : [null, outcome.error];
		}

		GetConfiguration() {
			return dom_fromMediaKeySystemConfiguration(this.handle.getConfiguration());
		}

		KeySystem() {
			return this.handle.keySystem;
		}
}

class dom_C_MediaKeys extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfMediaKeys() {
			
		}

		CreateSession(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return dom_wrap(dom.MediaKeySession, this.handle.createSession(...p0.map((item) => dom_unwrap(item))));
		}

		*GetStatusForPolicy(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			const outcome = yield* dom_await(() => this.handle.getStatusForPolicy(...p0.map((item) => dom_toMediaKeysPolicy(item))));
			
			return outcome.ok ? [outcome.value, null] : ["", outcome.error];
		}

		*SetServerCertificate(p0) {
			const outcome = yield* dom_await(() => this.handle.setServerCertificate(dom_unwrapAny(p0)));
			
			return outcome.ok ? [outcome.value, null] : [false, outcome.error];
		}
}

class dom_C_MediaQueryList extends dom_C_EventTarget {

		KindOfMediaQueryList() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		AddListener(p0) {
			this.handle.addListener((p0 === null || p0 === undefined ? null : (...a) => dom_unwrapAny(go2jsCallNow(p0, null, [a[0]]))));
		}

		RemoveListener(p0) {
			this.handle.removeListener((p0 === null || p0 === undefined ? null : (...a) => dom_unwrapAny(go2jsCallNow(p0, null, [a[0]]))));
		}

		OnChange(p0) {
			this.handle.addEventListener("change", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [a[0]])));
		}

		Onchange(p0) {
			return this.handle.onchange(dom_unwrapAny(p0));
		}

		Matches() {
			return this.handle.matches;
		}

		Media() {
			return this.handle.media;
		}
}

class dom_C_MediaSession extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfMediaSession() {
			
		}

		SetActionHandler(p0, p1) {
			this.handle.setActionHandler(p0, (p1 === null || p1 === undefined ? null : (...a) => go2jsCallNow(p1, null, [a[0]])));
		}

		SetPositionState(...p0) {
			this.handle.setPositionState(...p0.map((item) => dom_unwrapAny(item)));
		}

		Metadata() {
			return this.handle.metadata;
		}

		SetMetadata(value) {
			this.handle.metadata = dom_unwrapAny(value);
		}

		PlaybackState() {
			return this.handle.playbackState;
		}

		SetPlaybackState(value) {
			this.handle.playbackState = value;
		}
}

class dom_C_MediaStream extends dom_C_EventTarget {

		KindOfMediaStream() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		AddTrack(p0) {
			this.handle.addTrack(dom_unwrap(p0));
		}

		Clone() {
			return dom_wrap(dom.MediaStream, this.handle.clone());
		}

		GetAudioTracks() {
			return Array.from(this.handle.getAudioTracks() || [], (item) => dom_wrap(dom.MediaStreamTrack, item));
		}

		GetTrackById(p0) {
			return dom_wrap(dom.MediaStreamTrack, this.handle.getTrackById(p0));
		}

		GetTracks() {
			return Array.from(this.handle.getTracks() || [], (item) => dom_wrap(dom.MediaStreamTrack, item));
		}

		GetVideoTracks() {
			return Array.from(this.handle.getVideoTracks() || [], (item) => dom_wrap(dom.MediaStreamTrack, item));
		}

		RemoveTrack(p0) {
			this.handle.removeTrack(dom_unwrap(p0));
		}

		OnAddtrack(p0) {
			this.handle.addEventListener("addtrack", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MediaStreamTrackEvent, a[0])])));
		}

		OnRemovetrack(p0) {
			this.handle.addEventListener("removetrack", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MediaStreamTrackEvent, a[0])])));
		}

		Onaddtrack(p0) {
			return this.handle.onaddtrack(dom_unwrap(p0));
		}

		Onremovetrack(p0) {
			return this.handle.onremovetrack(dom_unwrap(p0));
		}

		Active() {
			return this.handle.active;
		}

		Id() {
			return this.handle.id;
		}
}

class dom_C_MediaStreamAudioDestinationNode extends dom_C_AudioNode {

		KindOfMediaStreamAudioDestinationNode() {
			
		}

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsAudioNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		Stream() {
			return dom_wrap(dom.MediaStream, this.handle.stream);
		}
}

class dom_C_MediaStreamAudioSourceNode extends dom_C_AudioNode {

		KindOfMediaStreamAudioSourceNode() {
			
		}

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsAudioNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		MediaStream() {
			return dom_wrap(dom.MediaStream, this.handle.mediaStream);
		}
}

class dom_C_MediaStreamTrack extends dom_C_EventTarget {

		KindOfMediaStreamTrack() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		*ApplyConstraints(...p0) {
			const outcome = yield* dom_await(() => this.handle.applyConstraints(...p0.map((item) => dom_unwrapAny(item))));
			
			return outcome.ok ? null : outcome.error;
		}

		Clone() {
			return dom_wrap(dom.MediaStreamTrack, this.handle.clone());
		}

		GetCapabilities() {
			return this.handle.getCapabilities();
		}

		GetConstraints() {
			return this.handle.getConstraints();
		}

		GetSettings() {
			return this.handle.getSettings();
		}

		Stop() {
			this.handle.stop();
		}

		OnEnded(p0) {
			this.handle.addEventListener("ended", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnMute(p0) {
			this.handle.addEventListener("mute", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnUnmute(p0) {
			this.handle.addEventListener("unmute", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		Onended(p0) {
			return this.handle.onended(dom_unwrap(p0));
		}

		Onmute(p0) {
			return this.handle.onmute(dom_unwrap(p0));
		}

		Onunmute(p0) {
			return this.handle.onunmute(dom_unwrap(p0));
		}

		ContentHint() {
			return this.handle.contentHint;
		}

		SetContentHint(value) {
			this.handle.contentHint = value;
		}

		Enabled() {
			return this.handle.enabled;
		}

		SetEnabled(value) {
			this.handle.enabled = value;
		}

		Id() {
			return this.handle.id;
		}

		Kind() {
			return this.handle.kind;
		}

		Label() {
			return this.handle.label;
		}

		Muted() {
			return this.handle.muted;
		}

		ReadyState() {
			return this.handle.readyState;
		}
}

class dom_C_MediaStreamTrackEvent extends dom_C_Event {

		KindOfMediaStreamTrackEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		Track() {
			return dom_wrap(dom.MediaStreamTrack, this.handle.track);
		}
}

class dom_C_MessageEvent extends dom_C_Event {

		KindOfMessageEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		InitMessageEvent(p0, ...p1) {
			this.handle.initMessageEvent(p0, ...p1.map((item) => dom_unwrapAny(item)));
		}

		Data() {
			return this.handle.data;
		}

		LastEventId() {
			return this.handle.lastEventId;
		}

		Origin() {
			return this.handle.origin;
		}

		Ports() {
			return Array.from(this.handle.ports || [], (item) => item);
		}

		Source() {
			return this.handle.source;
		}
}

class dom_C_MimeTypeArray extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfMimeTypeArray() {
			
		}

		Item(p0) {
			return this.handle.item(p0);
		}

		NamedItem(p0) {
			return this.handle.namedItem(p0);
		}

		Length() {
			return this.handle.length;
		}
}

class dom_C_MouseEventType extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfMouseEventType() {
			
		}

		Prototype() {
			return dom_wrap(dom.MouseEvent, this.handle.prototype);
		}

		SetPrototype(value) {
			this.handle.prototype = dom_unwrap(value);
		}
}

class dom_C_NamedNodeMap extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfNamedNodeMap() {
			
		}

		GetNamedItem(p0) {
			return dom_wrap(dom.Attr, this.handle.getNamedItem(p0));
		}

		GetNamedItemNS(p0, p1) {
			return dom_wrap(dom.Attr, this.handle.getNamedItemNS(p0, p1));
		}

		Item(p0) {
			return dom_wrap(dom.Attr, this.handle.item(p0));
		}

		RemoveNamedItem(p0) {
			return dom_wrap(dom.Attr, this.handle.removeNamedItem(p0));
		}

		RemoveNamedItemNS(p0, p1) {
			return dom_wrap(dom.Attr, this.handle.removeNamedItemNS(p0, p1));
		}

		SetNamedItem(p0) {
			return dom_wrap(dom.Attr, this.handle.setNamedItem(dom_unwrap(p0)));
		}

		SetNamedItemNS(p0) {
			return dom_wrap(dom.Attr, this.handle.setNamedItemNS(dom_unwrap(p0)));
		}

		Length() {
			return this.handle.length;
		}
}

class dom_C_NavigatorAutomationInformation extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfNavigatorAutomationInformation() {
			
		}

		AsNavigator() {
			return dom_is(this, "Navigator") ? this : null;
		}

		Webdriver() {
			return this.handle.webdriver;
		}
}

class dom_C_NavigatorBadge extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfNavigatorBadge() {
			
		}

		AsNavigator() {
			return dom_is(this, "Navigator") ? this : null;
		}

		*ClearAppBadge() {
			const outcome = yield* dom_await(() => this.handle.clearAppBadge());
			
			return outcome.ok ? null : outcome.error;
		}

		*SetAppBadge(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			const outcome = yield* dom_await(() => this.handle.setAppBadge(...p0.map((item) => item)));
			
			return outcome.ok ? null : outcome.error;
		}
}

class dom_C_NavigatorConcurrentHardware extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfNavigatorConcurrentHardware() {
			
		}

		AsNavigator() {
			return dom_is(this, "Navigator") ? this : null;
		}

		HardwareConcurrency() {
			return this.handle.hardwareConcurrency;
		}
}

class dom_C_NavigatorContentUtils extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfNavigatorContentUtils() {
			
		}

		AsNavigator() {
			return dom_is(this, "Navigator") ? this : null;
		}

		RegisterProtocolHandler(p0, p1) {
			this.handle.registerProtocolHandler(p0, dom_unwrapAny(p1));
		}
}

class dom_C_NavigatorCookies extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfNavigatorCookies() {
			
		}

		AsNavigator() {
			return dom_is(this, "Navigator") ? this : null;
		}

		CookieEnabled() {
			return this.handle.cookieEnabled;
		}
}

class dom_C_NavigatorID extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfNavigatorID() {
			
		}

		AsNavigator() {
			return dom_is(this, "Navigator") ? this : null;
		}

		AppCodeName() {
			return this.handle.appCodeName;
		}

		AppName() {
			return this.handle.appName;
		}

		AppVersion() {
			return this.handle.appVersion;
		}

		Platform() {
			return this.handle.platform;
		}

		Product() {
			return this.handle.product;
		}

		ProductSub() {
			return this.handle.productSub;
		}

		UserAgent() {
			return this.handle.userAgent;
		}

		Vendor() {
			return this.handle.vendor;
		}

		VendorSub() {
			return this.handle.vendorSub;
		}
}

class dom_C_NavigatorLanguage extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfNavigatorLanguage() {
			
		}

		AsNavigator() {
			return dom_is(this, "Navigator") ? this : null;
		}

		Language() {
			return this.handle.language;
		}

		Languages() {
			return Array.from(this.handle.languages || []);
		}
}

class dom_C_NavigatorLocks extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfNavigatorLocks() {
			
		}

		AsNavigator() {
			return dom_is(this, "Navigator") ? this : null;
		}

		Locks() {
			return dom_wrap(dom.LockManager, this.handle.locks);
		}
}

class dom_C_NavigatorOnLine extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfNavigatorOnLine() {
			
		}

		AsNavigator() {
			return dom_is(this, "Navigator") ? this : null;
		}

		OnLine() {
			return this.handle.onLine;
		}
}

class dom_C_NavigatorPlugins extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfNavigatorPlugins() {
			
		}

		AsNavigator() {
			return dom_is(this, "Navigator") ? this : null;
		}

		JavaEnabled() {
			return this.handle.javaEnabled();
		}

		MimeTypes() {
			return dom_wrap(dom.MimeTypeArray, this.handle.mimeTypes);
		}

		PdfViewerEnabled() {
			return this.handle.pdfViewerEnabled;
		}

		Plugins() {
			return dom_wrap(dom.PluginArray, this.handle.plugins);
		}
}

class dom_C_NavigatorStorage extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfNavigatorStorage() {
			
		}

		AsNavigator() {
			return dom_is(this, "Navigator") ? this : null;
		}

		Storage() {
			return dom_wrap(dom.StorageManager, this.handle.storage);
		}
}

class dom_C_Navigator extends dom_C_NavigatorAutomationInformation {

		KindOfNavigator() {
			
		}

		KindOfNavigatorAutomationInformation() {
			
		}

		KindOfNavigatorBadge() {
			
		}

		KindOfNavigatorConcurrentHardware() {
			
		}

		KindOfNavigatorContentUtils() {
			
		}

		KindOfNavigatorCookies() {
			
		}

		KindOfNavigatorID() {
			
		}

		KindOfNavigatorLanguage() {
			
		}

		KindOfNavigatorLocks() {
			
		}

		KindOfNavigatorOnLine() {
			
		}

		KindOfNavigatorPlugins() {
			
		}

		KindOfNavigatorStorage() {
			
		}

		AsNavigatorAutomationInformation() {
			return this;
		}

		AsNavigatorBadge() {
			return this;
		}

		AsNavigatorConcurrentHardware() {
			return this;
		}

		AsNavigatorContentUtils() {
			return this;
		}

		AsNavigatorCookies() {
			return this;
		}

		AsNavigatorID() {
			return this;
		}

		AsNavigatorLanguage() {
			return this;
		}

		AsNavigatorLocks() {
			return this;
		}

		AsNavigatorOnLine() {
			return this;
		}

		AsNavigatorPlugins() {
			return this;
		}

		AsNavigatorStorage() {
			return this;
		}

		CanShare(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return this.handle.canShare(...p0.map((item) => dom_toShareData(item)));
		}

		GetGamepads() {
			return Array.from(this.handle.getGamepads() || [], (item) => dom_fromGamepad(item));
		}

		*RequestMIDIAccess(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			const outcome = yield* dom_await(() => this.handle.requestMIDIAccess(...p0.map((item) => dom_toMIDIOptions(item))));
			
			return outcome.ok ? [dom_wrap(dom.MIDIAccess, outcome.value), null] : [null, outcome.error];
		}

		*RequestMediaKeySystemAccess(p0, p1) {
			const outcome = yield* dom_await(() => this.handle.requestMediaKeySystemAccess(p0, Array.from(p1 || [], (item) => dom_toMediaKeySystemConfiguration(item))));
			
			return outcome.ok ? [dom_wrap(dom.MediaKeySystemAccess, outcome.value), null] : [null, outcome.error];
		}

		SendBeacon(p0, ...p1) {
			return this.handle.sendBeacon(dom_unwrapAny(p0), ...p1.map((item) => dom_unwrapAny(item)));
		}

		*Share(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			const outcome = yield* dom_await(() => this.handle.share(...p0.map((item) => dom_toShareData(item))));
			
			return outcome.ok ? null : outcome.error;
		}

		Vibrate(p0) {
			return this.handle.vibrate(dom_unwrapAny(p0));
		}

		*ClearAppBadge() {
			const outcome = yield* dom_await(() => this.handle.clearAppBadge());
			
			return outcome.ok ? null : outcome.error;
		}

		*SetAppBadge(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			const outcome = yield* dom_await(() => this.handle.setAppBadge(...p0.map((item) => item)));
			
			return outcome.ok ? null : outcome.error;
		}

		RegisterProtocolHandler(p0, p1) {
			this.handle.registerProtocolHandler(p0, dom_unwrapAny(p1));
		}

		JavaEnabled() {
			return this.handle.javaEnabled();
		}

		Clipboard() {
			return dom_wrap(dom.Clipboard, this.handle.clipboard);
		}

		Credentials() {
			return dom_wrap(dom.CredentialsContainer, this.handle.credentials);
		}

		DoNotTrack() {
			return this.handle.doNotTrack;
		}

		Geolocation() {
			return dom_wrap(dom.Geolocation, this.handle.geolocation);
		}

		MaxTouchPoints() {
			return this.handle.maxTouchPoints;
		}

		MediaCapabilities() {
			return dom_wrap(dom.MediaCapabilities, this.handle.mediaCapabilities);
		}

		MediaDevices() {
			return dom_wrap(dom.MediaDevices, this.handle.mediaDevices);
		}

		MediaSession() {
			return dom_wrap(dom.MediaSession, this.handle.mediaSession);
		}

		Permissions() {
			return dom_wrap(dom.Permissions, this.handle.permissions);
		}

		ServiceWorker() {
			return dom_wrap(dom.ServiceWorkerContainer, this.handle.serviceWorker);
		}

		UserActivation() {
			return dom_fromUserActivation(this.handle.userActivation);
		}

		WakeLock() {
			return dom_wrap(dom.WakeLock, this.handle.wakeLock);
		}

		HardwareConcurrency() {
			return this.handle.hardwareConcurrency;
		}

		CookieEnabled() {
			return this.handle.cookieEnabled;
		}

		AppCodeName() {
			return this.handle.appCodeName;
		}

		AppName() {
			return this.handle.appName;
		}

		AppVersion() {
			return this.handle.appVersion;
		}

		Platform() {
			return this.handle.platform;
		}

		Product() {
			return this.handle.product;
		}

		ProductSub() {
			return this.handle.productSub;
		}

		UserAgent() {
			return this.handle.userAgent;
		}

		Vendor() {
			return this.handle.vendor;
		}

		VendorSub() {
			return this.handle.vendorSub;
		}

		Language() {
			return this.handle.language;
		}

		Languages() {
			return Array.from(this.handle.languages || []);
		}

		Locks() {
			return dom_wrap(dom.LockManager, this.handle.locks);
		}

		OnLine() {
			return this.handle.onLine;
		}

		MimeTypes() {
			return dom_wrap(dom.MimeTypeArray, this.handle.mimeTypes);
		}

		PdfViewerEnabled() {
			return this.handle.pdfViewerEnabled;
		}

		Plugins() {
			return dom_wrap(dom.PluginArray, this.handle.plugins);
		}

		Storage() {
			return dom_wrap(dom.StorageManager, this.handle.storage);
		}
}

class dom_C_NodeIterator extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfNodeIterator() {
			
		}

		Detach() {
			this.handle.detach();
		}

		NextNode() {
			return dom_wrap(dom.Node, this.handle.nextNode());
		}

		PreviousNode() {
			return dom_wrap(dom.Node, this.handle.previousNode());
		}

		Filter() {
			return this.handle.filter;
		}

		PointerBeforeReferenceNode() {
			return this.handle.pointerBeforeReferenceNode;
		}

		ReferenceNode() {
			return dom_wrap(dom.Node, this.handle.referenceNode);
		}

		Root() {
			return dom_wrap(dom.Node, this.handle.root);
		}

		WhatToShow() {
			return this.handle.whatToShow;
		}
}

class dom_C_NodeList extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfNodeList() {
			
		}

		AsNodeListOf() {
			return dom_is(this, "NodeListOf") ? this : null;
		}

		Item(p0) {
			return dom_wrap(dom.Node, this.handle.item(p0));
		}

		ForEach(p0, ...p1) {
			this.handle.forEach((p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Node, a[0]), a[1], dom_wrap(dom.NodeList, a[2])])), ...p1.map((item) => dom_unwrapAny(item)));
		}

		Length() {
			return this.handle.length;
		}
}

class dom_C_NodeListOf extends dom_C_NodeList {

		KindOfNodeListOf() {
			
		}

		KindOfNodeList() {
			
		}

		AsNodeList() {
			return this;
		}

		ForEach(p0, ...p1) {
			this.handle.forEach((p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Node, a[0]), a[1], dom_wrap(dom.NodeListOf, a[2])])), ...p1.map((item) => dom_unwrapAny(item)));
		}
}

class dom_C_OffscreenCanvas extends dom_C_EventTarget {

		KindOfOffscreenCanvas() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		*ConvertToBlob(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			const outcome = yield* dom_await(() => this.handle.convertToBlob(...p0.map((item) => dom_toImageEncodeOptions(item))));
			
			return outcome.ok ? [dom_wrap(dom.Blob, outcome.value), null] : [null, outcome.error];
		}

		GetContext(p0, ...p1) {
			return dom_wrap(dom.OffscreenCanvasRenderingContext2D, this.handle.getContext(p0, ...p1.map((item) => dom_unwrapAny(item))));
		}

		TransferToImageBitmap() {
			return dom_wrap(dom.ImageBitmap, this.handle.transferToImageBitmap());
		}

		OnContextlost(p0) {
			this.handle.addEventListener("contextlost", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnContextrestored(p0) {
			this.handle.addEventListener("contextrestored", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		Oncontextlost(p0) {
			return this.handle.oncontextlost(dom_unwrap(p0));
		}

		Oncontextrestored(p0) {
			return this.handle.oncontextrestored(dom_unwrap(p0));
		}

		Height() {
			return this.handle.height;
		}

		SetHeight(value) {
			this.handle.height = value;
		}

		Width() {
			return this.handle.width;
		}

		SetWidth(value) {
			this.handle.width = value;
		}
}

class dom_C_OffscreenCanvasRenderingContext2D extends dom_C_CanvasCompositing {

		KindOfOffscreenCanvasRenderingContext2D() {
			
		}

		KindOfCanvasCompositing() {
			
		}

		KindOfCanvasDrawImage() {
			
		}

		KindOfCanvasDrawPath() {
			
		}

		KindOfCanvasFillStrokeStyles() {
			
		}

		KindOfCanvasFilters() {
			
		}

		KindOfCanvasImageData() {
			
		}

		KindOfCanvasImageSmoothing() {
			
		}

		KindOfCanvasPath() {
			
		}

		KindOfCanvasPathDrawingStyles() {
			
		}

		KindOfCanvasRect() {
			
		}

		KindOfCanvasShadowStyles() {
			
		}

		KindOfCanvasState() {
			
		}

		KindOfCanvasText() {
			
		}

		KindOfCanvasTextDrawingStyles() {
			
		}

		KindOfCanvasTransform() {
			
		}

		AsCanvasCompositing() {
			return this;
		}

		AsCanvasDrawImage() {
			return this;
		}

		AsCanvasDrawPath() {
			return this;
		}

		AsCanvasFillStrokeStyles() {
			return this;
		}

		AsCanvasFilters() {
			return this;
		}

		AsCanvasImageData() {
			return this;
		}

		AsCanvasImageSmoothing() {
			return this;
		}

		AsCanvasPath() {
			return this;
		}

		AsCanvasPathDrawingStyles() {
			return this;
		}

		AsCanvasRect() {
			return this;
		}

		AsCanvasShadowStyles() {
			return this;
		}

		AsCanvasState() {
			return this;
		}

		AsCanvasText() {
			return this;
		}

		AsCanvasTextDrawingStyles() {
			return this;
		}

		AsCanvasTransform() {
			return this;
		}

		DrawImage(p0, p1, p2) {
			this.handle.drawImage(dom_unwrapAny(p0), p1, p2);
		}

		BeginPath() {
			this.handle.beginPath();
		}

		Clip(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.clip(...p0.map((item) => item));
		}

		Fill(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.fill(...p0.map((item) => item));
		}

		IsPointInPath(p0, p1, ...p2) {
			if (p2.length === 1 && Array.isArray(p2[0]) && p2[0].length === 0) {
				p2 = [];
			}
			
			return this.handle.isPointInPath(p0, p1, ...p2.map((item) => item));
		}

		IsPointInStroke(p0, p1) {
			return this.handle.isPointInStroke(p0, p1);
		}

		Stroke() {
			this.handle.stroke();
		}

		CreateConicGradient(p0, p1, p2) {
			return dom_wrap(dom.CanvasGradient, this.handle.createConicGradient(p0, p1, p2));
		}

		CreateLinearGradient(p0, p1, p2, p3) {
			return dom_wrap(dom.CanvasGradient, this.handle.createLinearGradient(p0, p1, p2, p3));
		}

		CreatePattern(p0, p1) {
			return dom_wrap(dom.CanvasPattern, this.handle.createPattern(dom_unwrapAny(p0), p1));
		}

		CreateRadialGradient(p0, p1, p2, p3, p4, p5) {
			return dom_wrap(dom.CanvasGradient, this.handle.createRadialGradient(p0, p1, p2, p3, p4, p5));
		}

		CreateImageData(p0, p1, ...p2) {
			if (p2.length === 1 && Array.isArray(p2[0]) && p2[0].length === 0) {
				p2 = [];
			}
			
			return dom_fromImageData(this.handle.createImageData(p0, p1, ...p2.map((item) => dom_toImageDataSettings(item))));
		}

		GetImageData(p0, p1, p2, p3, ...p4) {
			if (p4.length === 1 && Array.isArray(p4[0]) && p4[0].length === 0) {
				p4 = [];
			}
			
			return dom_fromImageData(this.handle.getImageData(p0, p1, p2, p3, ...p4.map((item) => dom_toImageDataSettings(item))));
		}

		PutImageData(p0, p1, p2) {
			this.handle.putImageData(dom_toImageData(p0), p1, p2);
		}

		Arc(p0, p1, p2, p3, p4, ...p5) {
			if (p5.length === 1 && Array.isArray(p5[0]) && p5[0].length === 0) {
				p5 = [];
			}
			
			this.handle.arc(p0, p1, p2, p3, p4, ...p5.map((item) => item));
		}

		ArcTo(p0, p1, p2, p3, p4) {
			this.handle.arcTo(p0, p1, p2, p3, p4);
		}

		BezierCurveTo(p0, p1, p2, p3, p4, p5) {
			this.handle.bezierCurveTo(p0, p1, p2, p3, p4, p5);
		}

		ClosePath() {
			this.handle.closePath();
		}

		Ellipse(p0, p1, p2, p3, p4, p5, p6, ...p7) {
			if (p7.length === 1 && Array.isArray(p7[0]) && p7[0].length === 0) {
				p7 = [];
			}
			
			this.handle.ellipse(p0, p1, p2, p3, p4, p5, p6, ...p7.map((item) => item));
		}

		LineTo(p0, p1) {
			this.handle.lineTo(p0, p1);
		}

		MoveTo(p0, p1) {
			this.handle.moveTo(p0, p1);
		}

		QuadraticCurveTo(p0, p1, p2, p3) {
			this.handle.quadraticCurveTo(p0, p1, p2, p3);
		}

		Rect(p0, p1, p2, p3) {
			this.handle.rect(p0, p1, p2, p3);
		}

		RoundRect(p0, p1, p2, p3, ...p4) {
			this.handle.roundRect(p0, p1, p2, p3, ...p4.map((item) => dom_unwrapAny(item)));
		}

		GetLineDash() {
			return Array.from(this.handle.getLineDash() || []);
		}

		SetLineDash(p0) {
			this.handle.setLineDash(Array.from(p0 || []));
		}

		ClearRect(p0, p1, p2, p3) {
			this.handle.clearRect(p0, p1, p2, p3);
		}

		FillRect(p0, p1, p2, p3) {
			this.handle.fillRect(p0, p1, p2, p3);
		}

		StrokeRect(p0, p1, p2, p3) {
			this.handle.strokeRect(p0, p1, p2, p3);
		}

		IsContextLost() {
			return this.handle.isContextLost();
		}

		Reset() {
			this.handle.reset();
		}

		Restore() {
			this.handle.restore();
		}

		Save() {
			this.handle.save();
		}

		FillText(p0, p1, p2, ...p3) {
			if (p3.length === 1 && Array.isArray(p3[0]) && p3[0].length === 0) {
				p3 = [];
			}
			
			this.handle.fillText(p0, p1, p2, ...p3.map((item) => item));
		}

		MeasureText(p0) {
			return dom_fromTextMetrics(this.handle.measureText(p0));
		}

		StrokeText(p0, p1, p2, ...p3) {
			if (p3.length === 1 && Array.isArray(p3[0]) && p3[0].length === 0) {
				p3 = [];
			}
			
			this.handle.strokeText(p0, p1, p2, ...p3.map((item) => item));
		}

		GetTransform() {
			return dom_wrap(dom.DOMMatrix, this.handle.getTransform());
		}

		ResetTransform() {
			this.handle.resetTransform();
		}

		Rotate(p0) {
			this.handle.rotate(p0);
		}

		Scale(p0, p1) {
			this.handle.scale(p0, p1);
		}

		SetTransform(p0, p1, p2, p3, p4, p5) {
			this.handle.setTransform(p0, p1, p2, p3, p4, p5);
		}

		Transform(p0, p1, p2, p3, p4, p5) {
			this.handle.transform(p0, p1, p2, p3, p4, p5);
		}

		Translate(p0, p1) {
			this.handle.translate(p0, p1);
		}

		Canvas() {
			return dom_wrap(dom.OffscreenCanvas, this.handle.canvas);
		}

		FillStyle() {
			return this.handle.fillStyle;
		}

		SetFillStyle(value) {
			this.handle.fillStyle = dom_unwrapAny(value);
		}

		StrokeStyle() {
			return this.handle.strokeStyle;
		}

		SetStrokeStyle(value) {
			this.handle.strokeStyle = dom_unwrapAny(value);
		}

		Filter() {
			return this.handle.filter;
		}

		SetFilter(value) {
			this.handle.filter = value;
		}

		ImageSmoothingEnabled() {
			return this.handle.imageSmoothingEnabled;
		}

		SetImageSmoothingEnabled(value) {
			this.handle.imageSmoothingEnabled = value;
		}

		ImageSmoothingQuality() {
			return this.handle.imageSmoothingQuality;
		}

		SetImageSmoothingQuality(value) {
			this.handle.imageSmoothingQuality = value;
		}

		LineCap() {
			return this.handle.lineCap;
		}

		SetLineCap(value) {
			this.handle.lineCap = value;
		}

		LineDashOffset() {
			return this.handle.lineDashOffset;
		}

		SetLineDashOffset(value) {
			this.handle.lineDashOffset = value;
		}

		LineJoin() {
			return this.handle.lineJoin;
		}

		SetLineJoin(value) {
			this.handle.lineJoin = value;
		}

		LineWidth() {
			return this.handle.lineWidth;
		}

		SetLineWidth(value) {
			this.handle.lineWidth = value;
		}

		MiterLimit() {
			return this.handle.miterLimit;
		}

		SetMiterLimit(value) {
			this.handle.miterLimit = value;
		}

		ShadowBlur() {
			return this.handle.shadowBlur;
		}

		SetShadowBlur(value) {
			this.handle.shadowBlur = value;
		}

		ShadowColor() {
			return this.handle.shadowColor;
		}

		SetShadowColor(value) {
			this.handle.shadowColor = value;
		}

		ShadowOffsetX() {
			return this.handle.shadowOffsetX;
		}

		SetShadowOffsetX(value) {
			this.handle.shadowOffsetX = value;
		}

		ShadowOffsetY() {
			return this.handle.shadowOffsetY;
		}

		SetShadowOffsetY(value) {
			this.handle.shadowOffsetY = value;
		}

		Direction() {
			return this.handle.direction;
		}

		SetDirection(value) {
			this.handle.direction = value;
		}

		Font() {
			return this.handle.font;
		}

		SetFont(value) {
			this.handle.font = value;
		}

		FontKerning() {
			return this.handle.fontKerning;
		}

		SetFontKerning(value) {
			this.handle.fontKerning = value;
		}

		FontStretch() {
			return this.handle.fontStretch;
		}

		SetFontStretch(value) {
			this.handle.fontStretch = value;
		}

		FontVariantCaps() {
			return this.handle.fontVariantCaps;
		}

		SetFontVariantCaps(value) {
			this.handle.fontVariantCaps = value;
		}

		LetterSpacing() {
			return this.handle.letterSpacing;
		}

		SetLetterSpacing(value) {
			this.handle.letterSpacing = value;
		}

		TextAlign() {
			return this.handle.textAlign;
		}

		SetTextAlign(value) {
			this.handle.textAlign = value;
		}

		TextBaseline() {
			return this.handle.textBaseline;
		}

		SetTextBaseline(value) {
			this.handle.textBaseline = value;
		}

		TextRendering() {
			return this.handle.textRendering;
		}

		SetTextRendering(value) {
			this.handle.textRendering = value;
		}

		WordSpacing() {
			return this.handle.wordSpacing;
		}

		SetWordSpacing(value) {
			this.handle.wordSpacing = value;
		}
}

class dom_C_OscillatorNode extends dom_C_AudioScheduledSourceNode {

		KindOfOscillatorNode() {
			
		}

		KindOfAudioScheduledSourceNode() {
			
		}

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsAudioScheduledSourceNode() {
			return this;
		}

		AsAudioNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		SetPeriodicWave(p0) {
			this.handle.setPeriodicWave(dom_unwrapAny(p0));
		}

		Detune() {
			return dom_wrap(dom.AudioParam, this.handle.detune);
		}

		Frequency() {
			return dom_wrap(dom.AudioParam, this.handle.frequency);
		}

		Type() {
			return this.handle.type;
		}

		SetType(value) {
			this.handle.type = value;
		}
}

class dom_C_PageTransitionEvent extends dom_C_Event {

		KindOfPageTransitionEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		Persisted() {
			return this.handle.persisted;
		}
}

class dom_C_PannerNode extends dom_C_AudioNode {

		KindOfPannerNode() {
			
		}

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsAudioNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		SetOrientation(p0, p1, p2) {
			this.handle.setOrientation(p0, p1, p2);
		}

		SetPosition(p0, p1, p2) {
			this.handle.setPosition(p0, p1, p2);
		}

		ConeInnerAngle() {
			return this.handle.coneInnerAngle;
		}

		SetConeInnerAngle(value) {
			this.handle.coneInnerAngle = value;
		}

		ConeOuterAngle() {
			return this.handle.coneOuterAngle;
		}

		SetConeOuterAngle(value) {
			this.handle.coneOuterAngle = value;
		}

		ConeOuterGain() {
			return this.handle.coneOuterGain;
		}

		SetConeOuterGain(value) {
			this.handle.coneOuterGain = value;
		}

		DistanceModel() {
			return this.handle.distanceModel;
		}

		SetDistanceModel(value) {
			this.handle.distanceModel = value;
		}

		MaxDistance() {
			return this.handle.maxDistance;
		}

		SetMaxDistance(value) {
			this.handle.maxDistance = value;
		}

		OrientationX() {
			return dom_wrap(dom.AudioParam, this.handle.orientationX);
		}

		OrientationY() {
			return dom_wrap(dom.AudioParam, this.handle.orientationY);
		}

		OrientationZ() {
			return dom_wrap(dom.AudioParam, this.handle.orientationZ);
		}

		PanningModel() {
			return this.handle.panningModel;
		}

		SetPanningModel(value) {
			this.handle.panningModel = value;
		}

		PositionX() {
			return dom_wrap(dom.AudioParam, this.handle.positionX);
		}

		PositionY() {
			return dom_wrap(dom.AudioParam, this.handle.positionY);
		}

		PositionZ() {
			return dom_wrap(dom.AudioParam, this.handle.positionZ);
		}

		RefDistance() {
			return this.handle.refDistance;
		}

		SetRefDistance(value) {
			this.handle.refDistance = value;
		}

		RolloffFactor() {
			return this.handle.rolloffFactor;
		}

		SetRolloffFactor(value) {
			this.handle.rolloffFactor = value;
		}
}

class dom_C_Path2D extends dom_C_CanvasPath {

		KindOfPath2D() {
			
		}

		KindOfCanvasPath() {
			
		}

		AsCanvasPath() {
			return this;
		}

		AddPath(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			this.handle.addPath(dom_unwrap(p0), ...p1.map((item) => dom_toDOMMatrix2DInit(item)));
		}
}

class dom_C_Path2DType extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfPath2DType() {
			
		}

		Prototype() {
			return dom_wrap(dom.Path2D, this.handle.prototype);
		}

		SetPrototype(value) {
			this.handle.prototype = dom_unwrap(value);
		}
}

class dom_C_Performance extends dom_C_EventTarget {

		KindOfPerformance() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		ClearMarks(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.clearMarks(...p0.map((item) => item));
		}

		ClearMeasures(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.clearMeasures(...p0.map((item) => item));
		}

		ClearResourceTimings() {
			this.handle.clearResourceTimings();
		}

		GetEntries() {
			return Array.from(this.handle.getEntries() || [], (item) => dom_wrap(dom.PerformanceEntry, item));
		}

		GetEntriesByName(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			return Array.from(this.handle.getEntriesByName(p0, ...p1.map((item) => item)) || [], (item) => dom_wrap(dom.PerformanceEntry, item));
		}

		GetEntriesByType(p0) {
			return Array.from(this.handle.getEntriesByType(p0) || [], (item) => dom_wrap(dom.PerformanceEntry, item));
		}

		Mark(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			return dom_wrap(dom.PerformanceMark, this.handle.mark(p0, ...p1.map((item) => dom_toPerformanceMarkOptions(item))));
		}

		Measure(p0, ...p1) {
			return dom_wrap(dom.PerformanceMeasure, this.handle.measure(p0, ...p1.map((item) => dom_unwrapAny(item))));
		}

		Now() {
			return this.handle.now();
		}

		SetResourceTimingBufferSize(p0) {
			this.handle.setResourceTimingBufferSize(p0);
		}

		ToJSON() {
			return this.handle.toJSON();
		}

		OnResourcetimingbufferfull(p0) {
			this.handle.addEventListener("resourcetimingbufferfull", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		Onresourcetimingbufferfull(p0) {
			return this.handle.onresourcetimingbufferfull(dom_unwrap(p0));
		}

		EventCounts() {
			return dom_wrap(dom.EventCounts, this.handle.eventCounts);
		}

		Navigation() {
			return dom_wrap(dom.PerformanceNavigation, this.handle.navigation);
		}

		TimeOrigin() {
			return this.handle.timeOrigin;
		}

		Timing() {
			return dom_wrap(dom.PerformanceTiming, this.handle.timing);
		}
}

class dom_C_PerformanceEntry extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfPerformanceEntry() {
			
		}

		AsPerformanceMark() {
			return dom_is(this, "PerformanceMark") ? this : null;
		}

		AsPerformanceMeasure() {
			return dom_is(this, "PerformanceMeasure") ? this : null;
		}

		ToJSON() {
			return this.handle.toJSON();
		}

		Duration() {
			return this.handle.duration;
		}

		EntryType() {
			return this.handle.entryType;
		}

		Name() {
			return this.handle.name;
		}

		StartTime() {
			return this.handle.startTime;
		}
}

class dom_C_PerformanceMark extends dom_C_PerformanceEntry {

		KindOfPerformanceMark() {
			
		}

		KindOfPerformanceEntry() {
			
		}

		AsPerformanceEntry() {
			return this;
		}

		Detail() {
			return this.handle.detail;
		}
}

class dom_C_PerformanceMeasure extends dom_C_PerformanceEntry {

		KindOfPerformanceMeasure() {
			
		}

		KindOfPerformanceEntry() {
			
		}

		AsPerformanceEntry() {
			return this;
		}

		Detail() {
			return this.handle.detail;
		}
}

class dom_C_PerformanceNavigation extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfPerformanceNavigation() {
			
		}

		ToJSON() {
			return this.handle.toJSON();
		}

		RedirectCount() {
			return this.handle.redirectCount;
		}

		Type() {
			return this.handle.type;
		}

		TYPE_NAVIGATE() {
			return this.handle.TYPE_NAVIGATE;
		}

		TYPE_RELOAD() {
			return this.handle.TYPE_RELOAD;
		}

		TYPE_BACK_FORWARD() {
			return this.handle.TYPE_BACK_FORWARD;
		}

		TYPE_RESERVED() {
			return this.handle.TYPE_RESERVED;
		}
}

class dom_C_PerformanceTiming extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfPerformanceTiming() {
			
		}

		ToJSON() {
			return this.handle.toJSON();
		}

		ConnectEnd() {
			return this.handle.connectEnd;
		}

		ConnectStart() {
			return this.handle.connectStart;
		}

		DomComplete() {
			return this.handle.domComplete;
		}

		DomContentLoadedEventEnd() {
			return this.handle.domContentLoadedEventEnd;
		}

		DomContentLoadedEventStart() {
			return this.handle.domContentLoadedEventStart;
		}

		DomInteractive() {
			return this.handle.domInteractive;
		}

		DomLoading() {
			return this.handle.domLoading;
		}

		DomainLookupEnd() {
			return this.handle.domainLookupEnd;
		}

		DomainLookupStart() {
			return this.handle.domainLookupStart;
		}

		FetchStart() {
			return this.handle.fetchStart;
		}

		LoadEventEnd() {
			return this.handle.loadEventEnd;
		}

		LoadEventStart() {
			return this.handle.loadEventStart;
		}

		NavigationStart() {
			return this.handle.navigationStart;
		}

		RedirectEnd() {
			return this.handle.redirectEnd;
		}

		RedirectStart() {
			return this.handle.redirectStart;
		}

		RequestStart() {
			return this.handle.requestStart;
		}

		ResponseEnd() {
			return this.handle.responseEnd;
		}

		ResponseStart() {
			return this.handle.responseStart;
		}

		SecureConnectionStart() {
			return this.handle.secureConnectionStart;
		}

		UnloadEventEnd() {
			return this.handle.unloadEventEnd;
		}

		UnloadEventStart() {
			return this.handle.unloadEventStart;
		}
}

class dom_C_Permissions extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfPermissions() {
			
		}

		*Query(p0) {
			const outcome = yield* dom_await(() => this.handle.query(dom_unwrapAny(p0)));
			
			return outcome.ok ? [outcome.value, null] : [null, outcome.error];
		}
}

class dom_C_PluginArray extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfPluginArray() {
			
		}

		Item(p0) {
			return this.handle.item(p0);
		}

		NamedItem(p0) {
			return this.handle.namedItem(p0);
		}

		Refresh() {
			this.handle.refresh();
		}

		Length() {
			return this.handle.length;
		}
}

class dom_C_PointerEvent extends dom_C_MouseEvent {

		KindOfPointerEvent() {
			
		}

		KindOfMouseEvent() {
			
		}

		KindOfUIEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsMouseEvent() {
			return this;
		}

		AsUIEvent() {
			return this;
		}

		AsEvent() {
			return this;
		}

		GetCoalescedEvents() {
			return Array.from(this.handle.getCoalescedEvents() || [], (item) => dom_wrap(dom.PointerEvent, item));
		}

		GetPredictedEvents() {
			return Array.from(this.handle.getPredictedEvents() || [], (item) => dom_wrap(dom.PointerEvent, item));
		}

		Height() {
			return this.handle.height;
		}

		IsPrimary() {
			return this.handle.isPrimary;
		}

		PointerId() {
			return this.handle.pointerId;
		}

		PointerType() {
			return this.handle.pointerType;
		}

		Pressure() {
			return this.handle.pressure;
		}

		TangentialPressure() {
			return this.handle.tangentialPressure;
		}

		TiltX() {
			return this.handle.tiltX;
		}

		TiltY() {
			return this.handle.tiltY;
		}

		Twist() {
			return this.handle.twist;
		}

		Width() {
			return this.handle.width;
		}
}

class dom_C_PointerEventType extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfPointerEventType() {
			
		}

		Prototype() {
			return dom_wrap(dom.PointerEvent, this.handle.prototype);
		}

		SetPrototype(value) {
			this.handle.prototype = dom_unwrap(value);
		}
}

class dom_C_PopStateEvent extends dom_C_Event {

		KindOfPopStateEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		HasUAVisualTransition() {
			return this.handle.hasUAVisualTransition;
		}

		State() {
			return this.handle.state;
		}
}

class dom_C_ProcessingInstruction extends dom_C_CharacterData {

		KindOfProcessingInstruction() {
			
		}

		KindOfCharacterData() {
			
		}

		KindOfLinkStyle() {
			
		}

		KindOfNode() {
			
		}

		KindOfChildNode() {
			
		}

		KindOfNonDocumentTypeChildNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsCharacterData() {
			return this;
		}

		AsLinkStyle() {
			return this;
		}

		AsNode() {
			return this;
		}

		AsChildNode() {
			return this;
		}

		AsNonDocumentTypeChildNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		Target() {
			return this.handle.target;
		}

		Sheet() {
			return dom_wrap(dom.CSSStyleSheet, this.handle.sheet);
		}
}

class dom_C_ProgressEvent extends dom_C_Event {

		KindOfProgressEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		LengthComputable() {
			return this.handle.lengthComputable;
		}

		Loaded() {
			return this.handle.loaded;
		}

		Total() {
			return this.handle.total;
		}
}

class dom_C_PromiseRejectionEvent extends dom_C_Event {

		KindOfPromiseRejectionEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		Reason() {
			return this.handle.reason;
		}
}

class dom_C_Range extends dom_C_AbstractRange {

		KindOfRange() {
			
		}

		KindOfAbstractRange() {
			
		}

		AsAbstractRange() {
			return this;
		}

		CloneContents() {
			return dom_wrap(dom.DocumentFragment, this.handle.cloneContents());
		}

		CloneRange() {
			return dom_wrap(dom.Range, this.handle.cloneRange());
		}

		Collapse(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.collapse(...p0.map((item) => item));
		}

		CompareBoundaryPoints(p0, p1) {
			return this.handle.compareBoundaryPoints(p0, dom_unwrap(p1));
		}

		ComparePoint(p0, p1) {
			return this.handle.comparePoint(dom_unwrap(p0), p1);
		}

		CreateContextualFragment(p0) {
			return dom_wrap(dom.DocumentFragment, this.handle.createContextualFragment(p0));
		}

		DeleteContents() {
			this.handle.deleteContents();
		}

		Detach() {
			this.handle.detach();
		}

		ExtractContents() {
			return dom_wrap(dom.DocumentFragment, this.handle.extractContents());
		}

		GetBoundingClientRect() {
			return dom_wrap(dom.DOMRect, this.handle.getBoundingClientRect());
		}

		GetClientRects() {
			return dom_wrap(dom.DOMRectList, this.handle.getClientRects());
		}

		InsertNode(p0) {
			this.handle.insertNode(dom_unwrap(p0));
		}

		IntersectsNode(p0) {
			return this.handle.intersectsNode(dom_unwrap(p0));
		}

		IsPointInRange(p0, p1) {
			return this.handle.isPointInRange(dom_unwrap(p0), p1);
		}

		SelectNode(p0) {
			this.handle.selectNode(dom_unwrap(p0));
		}

		SelectNodeContents(p0) {
			this.handle.selectNodeContents(dom_unwrap(p0));
		}

		SetEnd(p0, p1) {
			this.handle.setEnd(dom_unwrap(p0), p1);
		}

		SetEndAfter(p0) {
			this.handle.setEndAfter(dom_unwrap(p0));
		}

		SetEndBefore(p0) {
			this.handle.setEndBefore(dom_unwrap(p0));
		}

		SetStart(p0, p1) {
			this.handle.setStart(dom_unwrap(p0), p1);
		}

		SetStartAfter(p0) {
			this.handle.setStartAfter(dom_unwrap(p0));
		}

		SetStartBefore(p0) {
			this.handle.setStartBefore(dom_unwrap(p0));
		}

		SurroundContents(p0) {
			this.handle.surroundContents(dom_unwrap(p0));
		}

		ToString() {
			return this.handle.toString();
		}

		CommonAncestorContainer() {
			return dom_wrap(dom.Node, this.handle.commonAncestorContainer);
		}

		START_TO_START() {
			return this.handle.START_TO_START;
		}

		START_TO_END() {
			return this.handle.START_TO_END;
		}

		END_TO_END() {
			return this.handle.END_TO_END;
		}

		END_TO_START() {
			return this.handle.END_TO_START;
		}
}

class dom_C_ReadableStream extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfReadableStream() {
			
		}

		*Cancel(...p0) {
			const outcome = yield* dom_await(() => this.handle.cancel(...p0.map((item) => dom_unwrapAny(item))));
			
			return outcome.ok ? null : outcome.error;
		}

		GetReader(p0) {
			return this.handle.getReader(dom_unwrapAny(p0));
		}

		PipeThrough(p0, ...p1) {
			return dom_wrap(dom.ReadableStream, this.handle.pipeThrough(dom_unwrapAny(p0), ...p1.map((item) => dom_unwrapAny(item))));
		}

		*PipeTo(p0, ...p1) {
			const outcome = yield* dom_await(() => this.handle.pipeTo(dom_unwrapAny(p0), ...p1.map((item) => dom_unwrapAny(item))));
			
			return outcome.ok ? null : outcome.error;
		}

		Tee() {
			return this.handle.tee();
		}

		Locked() {
			return this.handle.locked;
		}
}

class dom_C_RemotePlayback extends dom_C_EventTarget {

		KindOfRemotePlayback() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		*CancelWatchAvailability(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			const outcome = yield* dom_await(() => this.handle.cancelWatchAvailability(...p0.map((item) => item)));
			
			return outcome.ok ? null : outcome.error;
		}

		*Prompt() {
			const outcome = yield* dom_await(() => this.handle.prompt());
			
			return outcome.ok ? null : outcome.error;
		}

		*WatchAvailability(p0) {
			const outcome = yield* dom_await(() => this.handle.watchAvailability((p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [a[0]]))));
			
			return outcome.ok ? [outcome.value, null] : [0, outcome.error];
		}

		OnConnect(p0) {
			this.handle.addEventListener("connect", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnConnecting(p0) {
			this.handle.addEventListener("connecting", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnDisconnect(p0) {
			this.handle.addEventListener("disconnect", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		Onconnect(p0) {
			return this.handle.onconnect(dom_unwrap(p0));
		}

		Onconnecting(p0) {
			return this.handle.onconnecting(dom_unwrap(p0));
		}

		Ondisconnect(p0) {
			return this.handle.ondisconnect(dom_unwrap(p0));
		}

		State() {
			return this.handle.state;
		}
}

class dom_C_Response extends dom_C_Body {

		KindOfResponse() {
			
		}

		KindOfBody() {
			
		}

		AsBody() {
			return this;
		}

		Clone() {
			return dom_wrap(dom.Response, this.handle.clone());
		}

		Headers() {
			return dom_wrap(dom.Headers, this.handle.headers);
		}

		Ok() {
			return this.handle.ok;
		}

		Redirected() {
			return this.handle.redirected;
		}

		Status() {
			return this.handle.status;
		}

		StatusText() {
			return this.handle.statusText;
		}

		Type() {
			return dom_wrap(dom.ResponseType, this.handle.type);
		}

		Url() {
			return this.handle.url;
		}
}

class dom_C_ResponseType extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfResponseType() {
			
		}

		Error() {
			return dom_wrap(dom.Response, this.handle.error());
		}

		Json(p0, ...p1) {
			return dom_wrap(dom.Response, this.handle.json(dom_unwrapAny(p0), ...p1.map((item) => dom_unwrapAny(item))));
		}

		Redirect(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			return dom_wrap(dom.Response, this.handle.redirect(dom_unwrapAny(p0), ...p1.map((item) => item)));
		}

		Prototype() {
			return dom_wrap(dom.Response, this.handle.prototype);
		}

		SetPrototype(value) {
			this.handle.prototype = dom_unwrap(value);
		}
}

class dom_C_SVGElement extends dom_C_Element {

		KindOfSVGElement() {
			
		}

		KindOfElement() {
			
		}

		KindOfElementCSSInlineStyle() {
			
		}

		KindOfGlobalEventHandlers() {
			
		}

		KindOfHTMLOrSVGElement() {
			
		}

		KindOfNode() {
			
		}

		KindOfARIAMixin() {
			
		}

		KindOfAnimatable() {
			
		}

		KindOfChildNode() {
			
		}

		KindOfNonDocumentTypeChildNode() {
			
		}

		KindOfParentNode() {
			
		}

		KindOfSlottable() {
			
		}

		KindOfEventTarget() {
			
		}

		AsElement() {
			return this;
		}

		AsElementCSSInlineStyle() {
			return this;
		}

		AsGlobalEventHandlers() {
			return this;
		}

		AsHTMLOrSVGElement() {
			return this;
		}

		AsNode() {
			return this;
		}

		AsARIAMixin() {
			return this;
		}

		AsAnimatable() {
			return this;
		}

		AsChildNode() {
			return this;
		}

		AsNonDocumentTypeChildNode() {
			return this;
		}

		AsParentNode() {
			return this;
		}

		AsSlottable() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		AsSVGGraphicsElement() {
			return dom_is(this, "SVGGraphicsElement") ? this : null;
		}

		AsSVGSVGElement() {
			return dom_is(this, "SVGSVGElement") ? this : null;
		}

		Blur() {
			this.handle.blur();
		}

		Focus(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.focus(...p0.map((item) => dom_toFocusOptions(item)));
		}

		OnAbort(p0) {
			this.handle.addEventListener("abort", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.UIEvent, a[0])])));
		}

		OnAnimationcancel(p0) {
			this.handle.addEventListener("animationcancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationEvent, a[0])])));
		}

		OnAnimationend(p0) {
			this.handle.addEventListener("animationend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationEvent, a[0])])));
		}

		OnAnimationiteration(p0) {
			this.handle.addEventListener("animationiteration", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationEvent, a[0])])));
		}

		OnAnimationstart(p0) {
			this.handle.addEventListener("animationstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationEvent, a[0])])));
		}

		OnAuxclick(p0) {
			this.handle.addEventListener("auxclick", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnBeforeinput(p0) {
			this.handle.addEventListener("beforeinput", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.InputEvent, a[0])])));
		}

		OnBeforetoggle(p0) {
			this.handle.addEventListener("beforetoggle", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnBlur(p0) {
			this.handle.addEventListener("blur", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FocusEvent, a[0])])));
		}

		OnCancel(p0) {
			this.handle.addEventListener("cancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCanplay(p0) {
			this.handle.addEventListener("canplay", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCanplaythrough(p0) {
			this.handle.addEventListener("canplaythrough", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnChange(p0) {
			this.handle.addEventListener("change", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnClick(p0) {
			this.handle.addEventListener("click", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnClose(p0) {
			this.handle.addEventListener("close", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCompositionend(p0) {
			this.handle.addEventListener("compositionend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.CompositionEvent, a[0])])));
		}

		OnCompositionstart(p0) {
			this.handle.addEventListener("compositionstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.CompositionEvent, a[0])])));
		}

		OnCompositionupdate(p0) {
			this.handle.addEventListener("compositionupdate", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.CompositionEvent, a[0])])));
		}

		OnContextlost(p0) {
			this.handle.addEventListener("contextlost", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnContextmenu(p0) {
			this.handle.addEventListener("contextmenu", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnContextrestored(p0) {
			this.handle.addEventListener("contextrestored", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCopy(p0) {
			this.handle.addEventListener("copy", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ClipboardEvent, a[0])])));
		}

		OnCuechange(p0) {
			this.handle.addEventListener("cuechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCut(p0) {
			this.handle.addEventListener("cut", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ClipboardEvent, a[0])])));
		}

		OnDblclick(p0) {
			this.handle.addEventListener("dblclick", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnDrag(p0) {
			this.handle.addEventListener("drag", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragend(p0) {
			this.handle.addEventListener("dragend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragenter(p0) {
			this.handle.addEventListener("dragenter", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragleave(p0) {
			this.handle.addEventListener("dragleave", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragover(p0) {
			this.handle.addEventListener("dragover", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragstart(p0) {
			this.handle.addEventListener("dragstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDrop(p0) {
			this.handle.addEventListener("drop", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDurationchange(p0) {
			this.handle.addEventListener("durationchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnEmptied(p0) {
			this.handle.addEventListener("emptied", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnEnded(p0) {
			this.handle.addEventListener("ended", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnError(p0) {
			this.handle.addEventListener("error", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ErrorEvent, a[0])])));
		}

		OnFocus(p0) {
			this.handle.addEventListener("focus", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FocusEvent, a[0])])));
		}

		OnFocusin(p0) {
			this.handle.addEventListener("focusin", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FocusEvent, a[0])])));
		}

		OnFocusout(p0) {
			this.handle.addEventListener("focusout", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FocusEvent, a[0])])));
		}

		OnFormdata(p0) {
			this.handle.addEventListener("formdata", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FormDataEvent, a[0])])));
		}

		OnGotpointercapture(p0) {
			this.handle.addEventListener("gotpointercapture", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnInput(p0) {
			this.handle.addEventListener("input", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnInvalid(p0) {
			this.handle.addEventListener("invalid", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnKeydown(p0) {
			this.handle.addEventListener("keydown", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.KeyboardEvent, a[0])])));
		}

		OnKeypress(p0) {
			this.handle.addEventListener("keypress", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.KeyboardEvent, a[0])])));
		}

		OnKeyup(p0) {
			this.handle.addEventListener("keyup", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.KeyboardEvent, a[0])])));
		}

		OnLoad(p0) {
			this.handle.addEventListener("load", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLoadeddata(p0) {
			this.handle.addEventListener("loadeddata", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLoadedmetadata(p0) {
			this.handle.addEventListener("loadedmetadata", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLoadstart(p0) {
			this.handle.addEventListener("loadstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLostpointercapture(p0) {
			this.handle.addEventListener("lostpointercapture", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnMousedown(p0) {
			this.handle.addEventListener("mousedown", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseenter(p0) {
			this.handle.addEventListener("mouseenter", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseleave(p0) {
			this.handle.addEventListener("mouseleave", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMousemove(p0) {
			this.handle.addEventListener("mousemove", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseout(p0) {
			this.handle.addEventListener("mouseout", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseover(p0) {
			this.handle.addEventListener("mouseover", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseup(p0) {
			this.handle.addEventListener("mouseup", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnPaste(p0) {
			this.handle.addEventListener("paste", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ClipboardEvent, a[0])])));
		}

		OnPause(p0) {
			this.handle.addEventListener("pause", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnPlay(p0) {
			this.handle.addEventListener("play", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnPlaying(p0) {
			this.handle.addEventListener("playing", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnPointercancel(p0) {
			this.handle.addEventListener("pointercancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerdown(p0) {
			this.handle.addEventListener("pointerdown", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerenter(p0) {
			this.handle.addEventListener("pointerenter", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerleave(p0) {
			this.handle.addEventListener("pointerleave", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointermove(p0) {
			this.handle.addEventListener("pointermove", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerout(p0) {
			this.handle.addEventListener("pointerout", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerover(p0) {
			this.handle.addEventListener("pointerover", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerup(p0) {
			this.handle.addEventListener("pointerup", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnProgress(p0) {
			this.handle.addEventListener("progress", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ProgressEvent, a[0])])));
		}

		OnRatechange(p0) {
			this.handle.addEventListener("ratechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnReset(p0) {
			this.handle.addEventListener("reset", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnResize(p0) {
			this.handle.addEventListener("resize", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.UIEvent, a[0])])));
		}

		OnScroll(p0) {
			this.handle.addEventListener("scroll", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnScrollend(p0) {
			this.handle.addEventListener("scrollend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSecuritypolicyviolation(p0) {
			this.handle.addEventListener("securitypolicyviolation", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.SecurityPolicyViolationEvent, a[0])])));
		}

		OnSeeked(p0) {
			this.handle.addEventListener("seeked", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSeeking(p0) {
			this.handle.addEventListener("seeking", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSelect(p0) {
			this.handle.addEventListener("select", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSelectionchange(p0) {
			this.handle.addEventListener("selectionchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSelectstart(p0) {
			this.handle.addEventListener("selectstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSlotchange(p0) {
			this.handle.addEventListener("slotchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnStalled(p0) {
			this.handle.addEventListener("stalled", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSubmit(p0) {
			this.handle.addEventListener("submit", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.SubmitEvent, a[0])])));
		}

		OnSuspend(p0) {
			this.handle.addEventListener("suspend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnTimeupdate(p0) {
			this.handle.addEventListener("timeupdate", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnToggle(p0) {
			this.handle.addEventListener("toggle", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnTouchcancel(p0) {
			this.handle.addEventListener("touchcancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TouchEvent, a[0])])));
		}

		OnTouchend(p0) {
			this.handle.addEventListener("touchend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TouchEvent, a[0])])));
		}

		OnTouchmove(p0) {
			this.handle.addEventListener("touchmove", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TouchEvent, a[0])])));
		}

		OnTouchstart(p0) {
			this.handle.addEventListener("touchstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TouchEvent, a[0])])));
		}

		OnTransitioncancel(p0) {
			this.handle.addEventListener("transitioncancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TransitionEvent, a[0])])));
		}

		OnTransitionend(p0) {
			this.handle.addEventListener("transitionend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TransitionEvent, a[0])])));
		}

		OnTransitionrun(p0) {
			this.handle.addEventListener("transitionrun", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TransitionEvent, a[0])])));
		}

		OnTransitionstart(p0) {
			this.handle.addEventListener("transitionstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TransitionEvent, a[0])])));
		}

		OnVolumechange(p0) {
			this.handle.addEventListener("volumechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWaiting(p0) {
			this.handle.addEventListener("waiting", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWebkitanimationend(p0) {
			this.handle.addEventListener("webkitanimationend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWebkitanimationiteration(p0) {
			this.handle.addEventListener("webkitanimationiteration", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWebkitanimationstart(p0) {
			this.handle.addEventListener("webkitanimationstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWebkittransitionend(p0) {
			this.handle.addEventListener("webkittransitionend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWheel(p0) {
			this.handle.addEventListener("wheel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.WheelEvent, a[0])])));
		}

		Onabort(p0) {
			return this.handle.onabort(dom_unwrap(p0));
		}

		Onanimationcancel(p0) {
			return this.handle.onanimationcancel(dom_unwrap(p0));
		}

		Onanimationend(p0) {
			return this.handle.onanimationend(dom_unwrap(p0));
		}

		Onanimationiteration(p0) {
			return this.handle.onanimationiteration(dom_unwrap(p0));
		}

		Onanimationstart(p0) {
			return this.handle.onanimationstart(dom_unwrap(p0));
		}

		Onauxclick(p0) {
			return this.handle.onauxclick(dom_unwrap(p0));
		}

		Onbeforeinput(p0) {
			return this.handle.onbeforeinput(dom_unwrap(p0));
		}

		Onbeforetoggle(p0) {
			return this.handle.onbeforetoggle(dom_unwrap(p0));
		}

		Onblur(p0) {
			return this.handle.onblur(dom_unwrap(p0));
		}

		Oncancel(p0) {
			return this.handle.oncancel(dom_unwrap(p0));
		}

		Oncanplay(p0) {
			return this.handle.oncanplay(dom_unwrap(p0));
		}

		Oncanplaythrough(p0) {
			return this.handle.oncanplaythrough(dom_unwrap(p0));
		}

		Onchange(p0) {
			return this.handle.onchange(dom_unwrap(p0));
		}

		Onclick(p0) {
			return this.handle.onclick(dom_unwrap(p0));
		}

		Onclose(p0) {
			return this.handle.onclose(dom_unwrap(p0));
		}

		Oncontextlost(p0) {
			return this.handle.oncontextlost(dom_unwrap(p0));
		}

		Oncontextmenu(p0) {
			return this.handle.oncontextmenu(dom_unwrap(p0));
		}

		Oncontextrestored(p0) {
			return this.handle.oncontextrestored(dom_unwrap(p0));
		}

		Oncopy(p0) {
			return this.handle.oncopy(dom_unwrap(p0));
		}

		Oncuechange(p0) {
			return this.handle.oncuechange(dom_unwrap(p0));
		}

		Oncut(p0) {
			return this.handle.oncut(dom_unwrap(p0));
		}

		Ondblclick(p0) {
			return this.handle.ondblclick(dom_unwrap(p0));
		}

		Ondrag(p0) {
			return this.handle.ondrag(dom_unwrap(p0));
		}

		Ondragend(p0) {
			return this.handle.ondragend(dom_unwrap(p0));
		}

		Ondragenter(p0) {
			return this.handle.ondragenter(dom_unwrap(p0));
		}

		Ondragleave(p0) {
			return this.handle.ondragleave(dom_unwrap(p0));
		}

		Ondragover(p0) {
			return this.handle.ondragover(dom_unwrap(p0));
		}

		Ondragstart(p0) {
			return this.handle.ondragstart(dom_unwrap(p0));
		}

		Ondrop(p0) {
			return this.handle.ondrop(dom_unwrap(p0));
		}

		Ondurationchange(p0) {
			return this.handle.ondurationchange(dom_unwrap(p0));
		}

		Onemptied(p0) {
			return this.handle.onemptied(dom_unwrap(p0));
		}

		Onended(p0) {
			return this.handle.onended(dom_unwrap(p0));
		}

		Onfocus(p0) {
			return this.handle.onfocus(dom_unwrap(p0));
		}

		Onformdata(p0) {
			return this.handle.onformdata(dom_unwrap(p0));
		}

		Ongotpointercapture(p0) {
			return this.handle.ongotpointercapture(dom_unwrap(p0));
		}

		Oninput(p0) {
			return this.handle.oninput(dom_unwrap(p0));
		}

		Oninvalid(p0) {
			return this.handle.oninvalid(dom_unwrap(p0));
		}

		Onkeydown(p0) {
			return this.handle.onkeydown(dom_unwrap(p0));
		}

		Onkeypress(p0) {
			return this.handle.onkeypress(dom_unwrap(p0));
		}

		Onkeyup(p0) {
			return this.handle.onkeyup(dom_unwrap(p0));
		}

		Onload(p0) {
			return this.handle.onload(dom_unwrap(p0));
		}

		Onloadeddata(p0) {
			return this.handle.onloadeddata(dom_unwrap(p0));
		}

		Onloadedmetadata(p0) {
			return this.handle.onloadedmetadata(dom_unwrap(p0));
		}

		Onloadstart(p0) {
			return this.handle.onloadstart(dom_unwrap(p0));
		}

		Onlostpointercapture(p0) {
			return this.handle.onlostpointercapture(dom_unwrap(p0));
		}

		Onmousedown(p0) {
			return this.handle.onmousedown(dom_unwrap(p0));
		}

		Onmouseenter(p0) {
			return this.handle.onmouseenter(dom_unwrap(p0));
		}

		Onmouseleave(p0) {
			return this.handle.onmouseleave(dom_unwrap(p0));
		}

		Onmousemove(p0) {
			return this.handle.onmousemove(dom_unwrap(p0));
		}

		Onmouseout(p0) {
			return this.handle.onmouseout(dom_unwrap(p0));
		}

		Onmouseover(p0) {
			return this.handle.onmouseover(dom_unwrap(p0));
		}

		Onmouseup(p0) {
			return this.handle.onmouseup(dom_unwrap(p0));
		}

		Onpaste(p0) {
			return this.handle.onpaste(dom_unwrap(p0));
		}

		Onpause(p0) {
			return this.handle.onpause(dom_unwrap(p0));
		}

		Onplay(p0) {
			return this.handle.onplay(dom_unwrap(p0));
		}

		Onplaying(p0) {
			return this.handle.onplaying(dom_unwrap(p0));
		}

		Onpointercancel(p0) {
			return this.handle.onpointercancel(dom_unwrap(p0));
		}

		Onpointerdown(p0) {
			return this.handle.onpointerdown(dom_unwrap(p0));
		}

		Onpointerenter(p0) {
			return this.handle.onpointerenter(dom_unwrap(p0));
		}

		Onpointerleave(p0) {
			return this.handle.onpointerleave(dom_unwrap(p0));
		}

		Onpointermove(p0) {
			return this.handle.onpointermove(dom_unwrap(p0));
		}

		Onpointerout(p0) {
			return this.handle.onpointerout(dom_unwrap(p0));
		}

		Onpointerover(p0) {
			return this.handle.onpointerover(dom_unwrap(p0));
		}

		Onpointerup(p0) {
			return this.handle.onpointerup(dom_unwrap(p0));
		}

		Onprogress(p0) {
			return this.handle.onprogress(dom_unwrap(p0));
		}

		Onratechange(p0) {
			return this.handle.onratechange(dom_unwrap(p0));
		}

		Onreset(p0) {
			return this.handle.onreset(dom_unwrap(p0));
		}

		Onresize(p0) {
			return this.handle.onresize(dom_unwrap(p0));
		}

		Onscroll(p0) {
			return this.handle.onscroll(dom_unwrap(p0));
		}

		Onscrollend(p0) {
			return this.handle.onscrollend(dom_unwrap(p0));
		}

		Onsecuritypolicyviolation(p0) {
			return this.handle.onsecuritypolicyviolation(dom_unwrap(p0));
		}

		Onseeked(p0) {
			return this.handle.onseeked(dom_unwrap(p0));
		}

		Onseeking(p0) {
			return this.handle.onseeking(dom_unwrap(p0));
		}

		Onselect(p0) {
			return this.handle.onselect(dom_unwrap(p0));
		}

		Onselectionchange(p0) {
			return this.handle.onselectionchange(dom_unwrap(p0));
		}

		Onselectstart(p0) {
			return this.handle.onselectstart(dom_unwrap(p0));
		}

		Onslotchange(p0) {
			return this.handle.onslotchange(dom_unwrap(p0));
		}

		Onstalled(p0) {
			return this.handle.onstalled(dom_unwrap(p0));
		}

		Onsubmit(p0) {
			return this.handle.onsubmit(dom_unwrap(p0));
		}

		Onsuspend(p0) {
			return this.handle.onsuspend(dom_unwrap(p0));
		}

		Ontimeupdate(p0) {
			return this.handle.ontimeupdate(dom_unwrap(p0));
		}

		Ontoggle(p0) {
			return this.handle.ontoggle(dom_unwrap(p0));
		}

		Ontouchcancel(p0) {
			return this.handle.ontouchcancel(dom_unwrap(p0));
		}

		Ontouchend(p0) {
			return this.handle.ontouchend(dom_unwrap(p0));
		}

		Ontouchmove(p0) {
			return this.handle.ontouchmove(dom_unwrap(p0));
		}

		Ontouchstart(p0) {
			return this.handle.ontouchstart(dom_unwrap(p0));
		}

		Ontransitioncancel(p0) {
			return this.handle.ontransitioncancel(dom_unwrap(p0));
		}

		Ontransitionend(p0) {
			return this.handle.ontransitionend(dom_unwrap(p0));
		}

		Ontransitionrun(p0) {
			return this.handle.ontransitionrun(dom_unwrap(p0));
		}

		Ontransitionstart(p0) {
			return this.handle.ontransitionstart(dom_unwrap(p0));
		}

		Onvolumechange(p0) {
			return this.handle.onvolumechange(dom_unwrap(p0));
		}

		Onwaiting(p0) {
			return this.handle.onwaiting(dom_unwrap(p0));
		}

		Onwebkitanimationend(p0) {
			return this.handle.onwebkitanimationend(dom_unwrap(p0));
		}

		Onwebkitanimationiteration(p0) {
			return this.handle.onwebkitanimationiteration(dom_unwrap(p0));
		}

		Onwebkitanimationstart(p0) {
			return this.handle.onwebkitanimationstart(dom_unwrap(p0));
		}

		Onwebkittransitionend(p0) {
			return this.handle.onwebkittransitionend(dom_unwrap(p0));
		}

		Onwheel(p0) {
			return this.handle.onwheel(dom_unwrap(p0));
		}

		OwnerSVGElement() {
			return dom_wrap(dom.SVGSVGElement, this.handle.ownerSVGElement);
		}

		ViewportElement() {
			return dom_wrap(dom.SVGElement, this.handle.viewportElement);
		}

		AttributeStyleMap() {
			return dom_wrap(dom.StylePropertyMap, this.handle.attributeStyleMap);
		}

		Style() {
			return dom_wrap(dom.CSSStyleDeclaration, this.handle.style);
		}

		Onerror() {
			return this.handle.onerror;
		}

		SetOnerror(value) {
			this.handle.onerror = (value === null || value === undefined ? null : (...a) => dom_unwrapAny(go2jsCallNow(value, null, [a[0], a[1], a[2], a[3], a[4]])));
		}

		Autofocus() {
			return this.handle.autofocus;
		}

		SetAutofocus(value) {
			this.handle.autofocus = value;
		}

		Dataset() {
			return this.handle.dataset;
		}

		Nonce() {
			return this.handle.nonce;
		}

		SetNonce(value) {
			this.handle.nonce = value;
		}

		TabIndex() {
			return this.handle.tabIndex;
		}

		SetTabIndex(value) {
			this.handle.tabIndex = value;
		}
}

class dom_C_SVGFitToViewBox extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfSVGFitToViewBox() {
			
		}

		AsSVGSVGElement() {
			return dom_is(this, "SVGSVGElement") ? this : null;
		}

		PreserveAspectRatio() {
			return this.handle.preserveAspectRatio;
		}

		ViewBox() {
			return this.handle.viewBox;
		}
}

class dom_C_SVGTests extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfSVGTests() {
			
		}

		AsSVGGraphicsElement() {
			return dom_is(this, "SVGGraphicsElement") ? this : null;
		}

		AsSVGSVGElement() {
			return dom_is(this, "SVGSVGElement") ? this : null;
		}

		RequiredExtensions() {
			return this.handle.requiredExtensions;
		}

		SystemLanguage() {
			return this.handle.systemLanguage;
		}
}

class dom_C_SVGGraphicsElement extends dom_C_SVGElement {

		KindOfSVGGraphicsElement() {
			
		}

		KindOfSVGElement() {
			
		}

		KindOfSVGTests() {
			
		}

		KindOfElement() {
			
		}

		KindOfElementCSSInlineStyle() {
			
		}

		KindOfGlobalEventHandlers() {
			
		}

		KindOfHTMLOrSVGElement() {
			
		}

		KindOfNode() {
			
		}

		KindOfARIAMixin() {
			
		}

		KindOfAnimatable() {
			
		}

		KindOfChildNode() {
			
		}

		KindOfNonDocumentTypeChildNode() {
			
		}

		KindOfParentNode() {
			
		}

		KindOfSlottable() {
			
		}

		KindOfEventTarget() {
			
		}

		AsSVGElement() {
			return this;
		}

		AsSVGTests() {
			return this;
		}

		AsElement() {
			return this;
		}

		AsElementCSSInlineStyle() {
			return this;
		}

		AsGlobalEventHandlers() {
			return this;
		}

		AsHTMLOrSVGElement() {
			return this;
		}

		AsNode() {
			return this;
		}

		AsARIAMixin() {
			return this;
		}

		AsAnimatable() {
			return this;
		}

		AsChildNode() {
			return this;
		}

		AsNonDocumentTypeChildNode() {
			return this;
		}

		AsParentNode() {
			return this;
		}

		AsSlottable() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		AsSVGSVGElement() {
			return dom_is(this, "SVGSVGElement") ? this : null;
		}

		GetBBox(...p0) {
			return dom_wrap(dom.DOMRect, this.handle.getBBox(...p0.map((item) => dom_unwrapAny(item))));
		}

		GetCTM() {
			return dom_wrap(dom.DOMMatrix, this.handle.getCTM());
		}

		GetScreenCTM() {
			return dom_wrap(dom.DOMMatrix, this.handle.getScreenCTM());
		}

		Transform() {
			return this.handle.transform;
		}

		RequiredExtensions() {
			return this.handle.requiredExtensions;
		}

		SystemLanguage() {
			return this.handle.systemLanguage;
		}
}

class dom_C_WindowEventHandlers extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfWindowEventHandlers() {
			
		}

		AsSVGSVGElement() {
			return dom_is(this, "SVGSVGElement") ? this : null;
		}

		AsWindow() {
			return dom_is(this, "Window") ? this : null;
		}

		AddEventListener(p0, p1, ...p2) {
			this.handle.addEventListener(p0, (p1 === null || p1 === undefined ? null : (...a) => go2jsCallNow(p1, null, [dom_wrap(dom.Event, a[0])])), ...p2.map((item) => dom_unwrapAny(item)));
		}

		RemoveEventListener(p0, p1, ...p2) {
			this.handle.removeEventListener(p0, (p1 === null || p1 === undefined ? null : (...a) => go2jsCallNow(p1, null, [dom_wrap(dom.Event, a[0])])), ...p2.map((item) => dom_unwrapAny(item)));
		}

		OnAfterprint(p0) {
			this.handle.addEventListener("afterprint", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnBeforeprint(p0) {
			this.handle.addEventListener("beforeprint", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnBeforeunload(p0) {
			this.handle.addEventListener("beforeunload", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.BeforeUnloadEvent, a[0])])));
		}

		OnGamepadconnected(p0) {
			this.handle.addEventListener("gamepadconnected", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.GamepadEvent, a[0])])));
		}

		OnGamepaddisconnected(p0) {
			this.handle.addEventListener("gamepaddisconnected", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.GamepadEvent, a[0])])));
		}

		OnHashchange(p0) {
			this.handle.addEventListener("hashchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.HashChangeEvent, a[0])])));
		}

		OnLanguagechange(p0) {
			this.handle.addEventListener("languagechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnMessage(p0) {
			this.handle.addEventListener("message", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MessageEvent, a[0])])));
		}

		OnMessageerror(p0) {
			this.handle.addEventListener("messageerror", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MessageEvent, a[0])])));
		}

		OnOffline(p0) {
			this.handle.addEventListener("offline", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnOnline(p0) {
			this.handle.addEventListener("online", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnPagehide(p0) {
			this.handle.addEventListener("pagehide", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PageTransitionEvent, a[0])])));
		}

		OnPageshow(p0) {
			this.handle.addEventListener("pageshow", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PageTransitionEvent, a[0])])));
		}

		OnPopstate(p0) {
			this.handle.addEventListener("popstate", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PopStateEvent, a[0])])));
		}

		OnRejectionhandled(p0) {
			this.handle.addEventListener("rejectionhandled", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PromiseRejectionEvent, a[0])])));
		}

		OnStorage(p0) {
			this.handle.addEventListener("storage", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.StorageEvent, a[0])])));
		}

		OnUnhandledrejection(p0) {
			this.handle.addEventListener("unhandledrejection", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PromiseRejectionEvent, a[0])])));
		}

		OnUnload(p0) {
			this.handle.addEventListener("unload", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		Onafterprint(p0) {
			return this.handle.onafterprint(dom_unwrap(p0));
		}

		Onbeforeprint(p0) {
			return this.handle.onbeforeprint(dom_unwrap(p0));
		}

		Onbeforeunload(p0) {
			return this.handle.onbeforeunload(dom_unwrap(p0));
		}

		Ongamepadconnected(p0) {
			return this.handle.ongamepadconnected(dom_unwrap(p0));
		}

		Ongamepaddisconnected(p0) {
			return this.handle.ongamepaddisconnected(dom_unwrap(p0));
		}

		Onhashchange(p0) {
			return this.handle.onhashchange(dom_unwrap(p0));
		}

		Onlanguagechange(p0) {
			return this.handle.onlanguagechange(dom_unwrap(p0));
		}

		Onmessage(p0) {
			return this.handle.onmessage(dom_unwrap(p0));
		}

		Onmessageerror(p0) {
			return this.handle.onmessageerror(dom_unwrap(p0));
		}

		Onoffline(p0) {
			return this.handle.onoffline(dom_unwrap(p0));
		}

		Ononline(p0) {
			return this.handle.ononline(dom_unwrap(p0));
		}

		Onpagehide(p0) {
			return this.handle.onpagehide(dom_unwrap(p0));
		}

		Onpageshow(p0) {
			return this.handle.onpageshow(dom_unwrap(p0));
		}

		Onpopstate(p0) {
			return this.handle.onpopstate(dom_unwrap(p0));
		}

		Onrejectionhandled(p0) {
			return this.handle.onrejectionhandled(dom_unwrap(p0));
		}

		Onstorage(p0) {
			return this.handle.onstorage(dom_unwrap(p0));
		}

		Onunhandledrejection(p0) {
			return this.handle.onunhandledrejection(dom_unwrap(p0));
		}

		Onunload(p0) {
			return this.handle.onunload(dom_unwrap(p0));
		}
}

class dom_C_SVGSVGElement extends dom_C_SVGGraphicsElement {

		KindOfSVGSVGElement() {
			
		}

		KindOfSVGGraphicsElement() {
			
		}

		KindOfSVGFitToViewBox() {
			
		}

		KindOfWindowEventHandlers() {
			
		}

		KindOfSVGElement() {
			
		}

		KindOfSVGTests() {
			
		}

		KindOfElement() {
			
		}

		KindOfElementCSSInlineStyle() {
			
		}

		KindOfGlobalEventHandlers() {
			
		}

		KindOfHTMLOrSVGElement() {
			
		}

		KindOfNode() {
			
		}

		KindOfARIAMixin() {
			
		}

		KindOfAnimatable() {
			
		}

		KindOfChildNode() {
			
		}

		KindOfNonDocumentTypeChildNode() {
			
		}

		KindOfParentNode() {
			
		}

		KindOfSlottable() {
			
		}

		KindOfEventTarget() {
			
		}

		AsSVGGraphicsElement() {
			return this;
		}

		AsSVGFitToViewBox() {
			return this;
		}

		AsWindowEventHandlers() {
			return this;
		}

		AsSVGElement() {
			return this;
		}

		AsSVGTests() {
			return this;
		}

		AsElement() {
			return this;
		}

		AsElementCSSInlineStyle() {
			return this;
		}

		AsGlobalEventHandlers() {
			return this;
		}

		AsHTMLOrSVGElement() {
			return this;
		}

		AsNode() {
			return this;
		}

		AsARIAMixin() {
			return this;
		}

		AsAnimatable() {
			return this;
		}

		AsChildNode() {
			return this;
		}

		AsNonDocumentTypeChildNode() {
			return this;
		}

		AsParentNode() {
			return this;
		}

		AsSlottable() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		AnimationsPaused() {
			return this.handle.animationsPaused();
		}

		CheckEnclosure(p0, p1) {
			return this.handle.checkEnclosure(dom_unwrap(p0), dom_unwrap(p1));
		}

		CheckIntersection(p0, p1) {
			return this.handle.checkIntersection(dom_unwrap(p0), dom_unwrap(p1));
		}

		CreateSVGAngle() {
			return this.handle.createSVGAngle();
		}

		CreateSVGLength() {
			return this.handle.createSVGLength();
		}

		CreateSVGMatrix() {
			return dom_wrap(dom.DOMMatrix, this.handle.createSVGMatrix());
		}

		CreateSVGNumber() {
			return this.handle.createSVGNumber();
		}

		CreateSVGPoint() {
			return dom_wrap(dom.DOMPoint, this.handle.createSVGPoint());
		}

		CreateSVGRect() {
			return dom_wrap(dom.DOMRect, this.handle.createSVGRect());
		}

		CreateSVGTransform() {
			return this.handle.createSVGTransform();
		}

		CreateSVGTransformFromMatrix(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return this.handle.createSVGTransformFromMatrix(...p0.map((item) => dom_toDOMMatrix2DInit(item)));
		}

		DeselectAll() {
			this.handle.deselectAll();
		}

		ForceRedraw() {
			this.handle.forceRedraw();
		}

		GetCurrentTime() {
			return this.handle.getCurrentTime();
		}

		GetElementById(p0) {
			return dom_wrap(dom.Element, this.handle.getElementById(p0));
		}

		GetEnclosureList(p0, p1) {
			return dom_wrap(dom.NodeListOf, this.handle.getEnclosureList(dom_unwrap(p0), dom_unwrap(p1)));
		}

		GetIntersectionList(p0, p1) {
			return dom_wrap(dom.NodeListOf, this.handle.getIntersectionList(dom_unwrap(p0), dom_unwrap(p1)));
		}

		PauseAnimations() {
			this.handle.pauseAnimations();
		}

		SetCurrentTime(p0) {
			this.handle.setCurrentTime(p0);
		}

		SuspendRedraw(p0) {
			return this.handle.suspendRedraw(p0);
		}

		UnpauseAnimations() {
			this.handle.unpauseAnimations();
		}

		UnsuspendRedraw(p0) {
			this.handle.unsuspendRedraw(p0);
		}

		UnsuspendRedrawAll() {
			this.handle.unsuspendRedrawAll();
		}

		OnAfterprint(p0) {
			this.handle.addEventListener("afterprint", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnBeforeprint(p0) {
			this.handle.addEventListener("beforeprint", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnBeforeunload(p0) {
			this.handle.addEventListener("beforeunload", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.BeforeUnloadEvent, a[0])])));
		}

		OnGamepadconnected(p0) {
			this.handle.addEventListener("gamepadconnected", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.GamepadEvent, a[0])])));
		}

		OnGamepaddisconnected(p0) {
			this.handle.addEventListener("gamepaddisconnected", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.GamepadEvent, a[0])])));
		}

		OnHashchange(p0) {
			this.handle.addEventListener("hashchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.HashChangeEvent, a[0])])));
		}

		OnLanguagechange(p0) {
			this.handle.addEventListener("languagechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnMessage(p0) {
			this.handle.addEventListener("message", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MessageEvent, a[0])])));
		}

		OnMessageerror(p0) {
			this.handle.addEventListener("messageerror", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MessageEvent, a[0])])));
		}

		OnOffline(p0) {
			this.handle.addEventListener("offline", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnOnline(p0) {
			this.handle.addEventListener("online", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnPagehide(p0) {
			this.handle.addEventListener("pagehide", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PageTransitionEvent, a[0])])));
		}

		OnPageshow(p0) {
			this.handle.addEventListener("pageshow", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PageTransitionEvent, a[0])])));
		}

		OnPopstate(p0) {
			this.handle.addEventListener("popstate", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PopStateEvent, a[0])])));
		}

		OnRejectionhandled(p0) {
			this.handle.addEventListener("rejectionhandled", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PromiseRejectionEvent, a[0])])));
		}

		OnStorage(p0) {
			this.handle.addEventListener("storage", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.StorageEvent, a[0])])));
		}

		OnUnhandledrejection(p0) {
			this.handle.addEventListener("unhandledrejection", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PromiseRejectionEvent, a[0])])));
		}

		OnUnload(p0) {
			this.handle.addEventListener("unload", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		Onafterprint(p0) {
			return this.handle.onafterprint(dom_unwrap(p0));
		}

		Onbeforeprint(p0) {
			return this.handle.onbeforeprint(dom_unwrap(p0));
		}

		Onbeforeunload(p0) {
			return this.handle.onbeforeunload(dom_unwrap(p0));
		}

		Ongamepadconnected(p0) {
			return this.handle.ongamepadconnected(dom_unwrap(p0));
		}

		Ongamepaddisconnected(p0) {
			return this.handle.ongamepaddisconnected(dom_unwrap(p0));
		}

		Onhashchange(p0) {
			return this.handle.onhashchange(dom_unwrap(p0));
		}

		Onlanguagechange(p0) {
			return this.handle.onlanguagechange(dom_unwrap(p0));
		}

		Onmessage(p0) {
			return this.handle.onmessage(dom_unwrap(p0));
		}

		Onmessageerror(p0) {
			return this.handle.onmessageerror(dom_unwrap(p0));
		}

		Onoffline(p0) {
			return this.handle.onoffline(dom_unwrap(p0));
		}

		Ononline(p0) {
			return this.handle.ononline(dom_unwrap(p0));
		}

		Onpagehide(p0) {
			return this.handle.onpagehide(dom_unwrap(p0));
		}

		Onpageshow(p0) {
			return this.handle.onpageshow(dom_unwrap(p0));
		}

		Onpopstate(p0) {
			return this.handle.onpopstate(dom_unwrap(p0));
		}

		Onrejectionhandled(p0) {
			return this.handle.onrejectionhandled(dom_unwrap(p0));
		}

		Onstorage(p0) {
			return this.handle.onstorage(dom_unwrap(p0));
		}

		Onunhandledrejection(p0) {
			return this.handle.onunhandledrejection(dom_unwrap(p0));
		}

		Onunload(p0) {
			return this.handle.onunload(dom_unwrap(p0));
		}

		CurrentScale() {
			return this.handle.currentScale;
		}

		SetCurrentScale(value) {
			this.handle.currentScale = value;
		}

		CurrentTranslate() {
			return dom_wrap(dom.DOMPointReadOnly, this.handle.currentTranslate);
		}

		Height() {
			return this.handle.height;
		}

		Width() {
			return this.handle.width;
		}

		X() {
			return this.handle.x;
		}

		Y() {
			return this.handle.y;
		}

		PreserveAspectRatio() {
			return this.handle.preserveAspectRatio;
		}

		ViewBox() {
			return this.handle.viewBox;
		}
}

class dom_C_ScriptProcessorNode extends dom_C_AudioNode {

		KindOfScriptProcessorNode() {
			
		}

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsAudioNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		OnAudioprocess(p0) {
			this.handle.addEventListener("audioprocess", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AudioProcessingEvent, a[0])])));
		}

		Onaudioprocess(p0) {
			return this.handle.onaudioprocess(dom_unwrap(p0));
		}

		BufferSize() {
			return this.handle.bufferSize;
		}
}

class dom_C_SecurityPolicyViolationEvent extends dom_C_Event {

		KindOfSecurityPolicyViolationEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		BlockedURI() {
			return this.handle.blockedURI;
		}

		ColumnNumber() {
			return this.handle.columnNumber;
		}

		Disposition() {
			return this.handle.disposition;
		}

		DocumentURI() {
			return this.handle.documentURI;
		}

		EffectiveDirective() {
			return this.handle.effectiveDirective;
		}

		LineNumber() {
			return this.handle.lineNumber;
		}

		OriginalPolicy() {
			return this.handle.originalPolicy;
		}

		Referrer() {
			return this.handle.referrer;
		}

		Sample() {
			return this.handle.sample;
		}

		SourceFile() {
			return this.handle.sourceFile;
		}

		StatusCode() {
			return this.handle.statusCode;
		}

		ViolatedDirective() {
			return this.handle.violatedDirective;
		}
}

class dom_C_Selection extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfSelection() {
			
		}

		AddRange(p0) {
			this.handle.addRange(dom_unwrap(p0));
		}

		Collapse(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			this.handle.collapse(dom_unwrap(p0), ...p1.map((item) => item));
		}

		CollapseToEnd() {
			this.handle.collapseToEnd();
		}

		CollapseToStart() {
			this.handle.collapseToStart();
		}

		ContainsNode(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			return this.handle.containsNode(dom_unwrap(p0), ...p1.map((item) => item));
		}

		DeleteFromDocument() {
			this.handle.deleteFromDocument();
		}

		Empty() {
			this.handle.empty();
		}

		Extend(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			this.handle.extend(dom_unwrap(p0), ...p1.map((item) => item));
		}

		GetRangeAt(p0) {
			return dom_wrap(dom.Range, this.handle.getRangeAt(p0));
		}

		Modify(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.modify(...p0.map((item) => item));
		}

		RemoveAllRanges() {
			this.handle.removeAllRanges();
		}

		RemoveRange(p0) {
			this.handle.removeRange(dom_unwrap(p0));
		}

		SelectAllChildren(p0) {
			this.handle.selectAllChildren(dom_unwrap(p0));
		}

		SetBaseAndExtent(p0, p1, p2, p3) {
			this.handle.setBaseAndExtent(dom_unwrap(p0), p1, dom_unwrap(p2), p3);
		}

		SetPosition(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			this.handle.setPosition(dom_unwrap(p0), ...p1.map((item) => item));
		}

		ToString() {
			return this.handle.toString();
		}

		AnchorNode() {
			return dom_wrap(dom.Node, this.handle.anchorNode);
		}

		AnchorOffset() {
			return this.handle.anchorOffset;
		}

		Direction() {
			return this.handle.direction;
		}

		FocusNode() {
			return dom_wrap(dom.Node, this.handle.focusNode);
		}

		FocusOffset() {
			return this.handle.focusOffset;
		}

		IsCollapsed() {
			return this.handle.isCollapsed;
		}

		RangeCount() {
			return this.handle.rangeCount;
		}

		Type() {
			return this.handle.type;
		}
}

class dom_C_ServiceWorkerContainer extends dom_C_EventTarget {

		KindOfServiceWorkerContainer() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		*GetRegistration(...p0) {
			const outcome = yield* dom_await(() => this.handle.getRegistration(...p0.map((item) => dom_unwrapAny(item))));
			
			return outcome.ok ? [outcome.value, null] : [null, outcome.error];
		}

		*GetRegistrations() {
			const outcome = yield* dom_await(() => this.handle.getRegistrations());
			
			return outcome.ok ? [Array.from(outcome.value || [], (item) => item), null] : [[], outcome.error];
		}

		*Register(p0, ...p1) {
			const outcome = yield* dom_await(() => this.handle.register(dom_unwrapAny(p0), ...p1.map((item) => dom_unwrapAny(item))));
			
			return outcome.ok ? [outcome.value, null] : [null, outcome.error];
		}

		StartMessages() {
			this.handle.startMessages();
		}

		OnControllerchange(p0) {
			this.handle.addEventListener("controllerchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnMessage(p0) {
			this.handle.addEventListener("message", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MessageEvent, a[0])])));
		}

		OnMessageerror(p0) {
			this.handle.addEventListener("messageerror", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MessageEvent, a[0])])));
		}

		Oncontrollerchange(p0) {
			return this.handle.oncontrollerchange(dom_unwrap(p0));
		}

		Onmessage(p0) {
			return this.handle.onmessage(dom_unwrap(p0));
		}

		Onmessageerror(p0) {
			return this.handle.onmessageerror(dom_unwrap(p0));
		}

		Controller() {
			return this.handle.controller;
		}
}

class dom_C_ShadowRoot extends dom_C_DocumentFragment {

		KindOfShadowRoot() {
			
		}

		KindOfDocumentFragment() {
			
		}

		KindOfDocumentOrShadowRoot() {
			
		}

		KindOfNode() {
			
		}

		KindOfNonElementParentNode() {
			
		}

		KindOfParentNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsDocumentFragment() {
			return this;
		}

		AsDocumentOrShadowRoot() {
			return this;
		}

		AsNode() {
			return this;
		}

		AsNonElementParentNode() {
			return this;
		}

		AsParentNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		GetHTML(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return this.handle.getHTML(...p0.map((item) => dom_toGetHTMLOptions(item)));
		}

		SetHTMLUnsafe(p0) {
			this.handle.setHTMLUnsafe(p0);
		}

		ElementFromPoint(p0, p1) {
			return dom_wrap(dom.Element, this.handle.elementFromPoint(p0, p1));
		}

		ElementsFromPoint(p0, p1) {
			return Array.from(this.handle.elementsFromPoint(p0, p1) || [], (item) => dom_wrap(dom.Element, item));
		}

		GetAnimations() {
			return Array.from(this.handle.getAnimations() || [], (item) => dom_wrap(dom.Animation, item));
		}

		OnSlotchange(p0) {
			this.handle.addEventListener("slotchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		Onslotchange(p0) {
			return this.handle.onslotchange(dom_unwrap(p0));
		}

		Clonable() {
			return this.handle.clonable;
		}

		DelegatesFocus() {
			return this.handle.delegatesFocus;
		}

		Host() {
			return dom_wrap(dom.Element, this.handle.host);
		}

		InnerHTML() {
			return this.handle.innerHTML;
		}

		SetInnerHTML(value) {
			this.handle.innerHTML = value;
		}

		Mode() {
			return this.handle.mode;
		}

		Serializable() {
			return this.handle.serializable;
		}

		SlotAssignment() {
			return this.handle.slotAssignment;
		}

		ActiveElement() {
			return dom_wrap(dom.Element, this.handle.activeElement);
		}

		AdoptedStyleSheets() {
			return Array.from(this.handle.adoptedStyleSheets || [], (item) => dom_wrap(dom.CSSStyleSheet, item));
		}

		SetAdoptedStyleSheets(value) {
			this.handle.adoptedStyleSheets = Array.from(value || [], (item) => dom_unwrap(item));
		}

		FullscreenElement() {
			return dom_wrap(dom.Element, this.handle.fullscreenElement);
		}

		PictureInPictureElement() {
			return dom_wrap(dom.Element, this.handle.pictureInPictureElement);
		}

		PointerLockElement() {
			return dom_wrap(dom.Element, this.handle.pointerLockElement);
		}

		StyleSheets() {
			return dom_wrap(dom.StyleSheetList, this.handle.styleSheets);
		}
}

class dom_C_SpeechSynthesis extends dom_C_EventTarget {

		KindOfSpeechSynthesis() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		Cancel() {
			this.handle.cancel();
		}

		GetVoices() {
			return Array.from(this.handle.getVoices() || [], (item) => item);
		}

		Pause() {
			this.handle.pause();
		}

		Resume() {
			this.handle.resume();
		}

		Speak(p0) {
			this.handle.speak(dom_unwrapAny(p0));
		}

		OnVoiceschanged(p0) {
			this.handle.addEventListener("voiceschanged", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		Onvoiceschanged(p0) {
			return this.handle.onvoiceschanged(dom_unwrap(p0));
		}

		Paused() {
			return this.handle.paused;
		}

		Pending() {
			return this.handle.pending;
		}

		Speaking() {
			return this.handle.speaking;
		}
}

class dom_C_StereoPannerNode extends dom_C_AudioNode {

		KindOfStereoPannerNode() {
			
		}

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsAudioNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		Pan() {
			return dom_wrap(dom.AudioParam, this.handle.pan);
		}
}

class dom_C_Storage extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfStorage() {
			
		}

		Clear() {
			this.handle.clear();
		}

		GetItem(p0) {
			return this.handle.getItem(p0);
		}

		Key(p0) {
			return this.handle.key(p0);
		}

		RemoveItem(p0) {
			this.handle.removeItem(p0);
		}

		SetItem(p0, p1) {
			this.handle.setItem(p0, p1);
		}

		Length() {
			return this.handle.length;
		}
}

class dom_C_StorageEvent extends dom_C_Event {

		KindOfStorageEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		InitStorageEvent(p0, ...p1) {
			this.handle.initStorageEvent(p0, ...p1.map((item) => dom_unwrapAny(item)));
		}

		Key() {
			return this.handle.key;
		}

		NewValue() {
			return this.handle.newValue;
		}

		OldValue() {
			return this.handle.oldValue;
		}

		StorageArea() {
			return dom_wrap(dom.Storage, this.handle.storageArea);
		}

		Url() {
			return this.handle.url;
		}
}

class dom_C_StorageManager extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfStorageManager() {
			
		}

		*Estimate() {
			const outcome = yield* dom_await(() => this.handle.estimate());
			
			return outcome.ok ? [outcome.value, null] : [null, outcome.error];
		}

		*GetDirectory() {
			const outcome = yield* dom_await(() => this.handle.getDirectory());
			
			return outcome.ok ? [outcome.value, null] : [null, outcome.error];
		}

		*Persist() {
			const outcome = yield* dom_await(() => this.handle.persist());
			
			return outcome.ok ? [outcome.value, null] : [false, outcome.error];
		}

		*Persisted() {
			const outcome = yield* dom_await(() => this.handle.persisted());
			
			return outcome.ok ? [outcome.value, null] : [false, outcome.error];
		}
}

class dom_C_StylePropertyMapReadOnly extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfStylePropertyMapReadOnly() {
			
		}

		AsStylePropertyMap() {
			return dom_is(this, "StylePropertyMap") ? this : null;
		}

		Get(p0) {
			return dom_wrap(dom.CSSStyleValue, this.handle.get(p0));
		}

		GetAll(p0) {
			return Array.from(this.handle.getAll(p0) || [], (item) => dom_wrap(dom.CSSStyleValue, item));
		}

		Has(p0) {
			return this.handle.has(p0);
		}

		ForEach(p0, ...p1) {
			this.handle.forEach((p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [Array.from(a[0] || [], (item) => dom_wrap(dom.CSSStyleValue, item)), a[1], dom_wrap(dom.StylePropertyMapReadOnly, a[2])])), ...p1.map((item) => dom_unwrapAny(item)));
		}

		Size() {
			return this.handle.size;
		}
}

class dom_C_StylePropertyMap extends dom_C_StylePropertyMapReadOnly {

		KindOfStylePropertyMap() {
			
		}

		KindOfStylePropertyMapReadOnly() {
			
		}

		AsStylePropertyMapReadOnly() {
			return this;
		}

		Append(p0, ...p1) {
			this.handle.append(p0, ...p1.map((item) => dom_unwrapAny(item)));
		}

		Clear() {
			this.handle.clear();
		}

		Delete(p0) {
			this.handle.delete(p0);
		}

		Set(p0, ...p1) {
			this.handle.set(p0, ...p1.map((item) => dom_unwrapAny(item)));
		}
}

class dom_C_StyleSheetList extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfStyleSheetList() {
			
		}

		Item(p0) {
			return dom_wrap(dom.CSSStyleSheet, this.handle.item(p0));
		}

		Length() {
			return this.handle.length;
		}
}

class dom_C_SubmitEvent extends dom_C_Event {

		KindOfSubmitEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		Submitter() {
			return dom_wrap(dom.HTMLElement, this.handle.submitter);
		}
}

class dom_C_TextTrack extends dom_C_EventTarget {

		KindOfTextTrack() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		AddCue(p0) {
			this.handle.addCue(dom_unwrap(p0));
		}

		RemoveCue(p0) {
			this.handle.removeCue(dom_unwrap(p0));
		}

		OnCuechange(p0) {
			this.handle.addEventListener("cuechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		Oncuechange(p0) {
			return this.handle.oncuechange(dom_unwrap(p0));
		}

		ActiveCues() {
			return dom_wrap(dom.TextTrackCueList, this.handle.activeCues);
		}

		Cues() {
			return dom_wrap(dom.TextTrackCueList, this.handle.cues);
		}

		Id() {
			return this.handle.id;
		}

		InBandMetadataTrackDispatchType() {
			return this.handle.inBandMetadataTrackDispatchType;
		}

		Kind() {
			return this.handle.kind;
		}

		Label() {
			return this.handle.label;
		}

		Language() {
			return this.handle.language;
		}

		Mode() {
			return this.handle.mode;
		}

		SetMode(value) {
			this.handle.mode = value;
		}
}

class dom_C_TextTrackCue extends dom_C_EventTarget {

		KindOfTextTrackCue() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		OnEnter(p0) {
			this.handle.addEventListener("enter", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnExit(p0) {
			this.handle.addEventListener("exit", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		Onenter(p0) {
			return this.handle.onenter(dom_unwrap(p0));
		}

		Onexit(p0) {
			return this.handle.onexit(dom_unwrap(p0));
		}

		EndTime() {
			return this.handle.endTime;
		}

		SetEndTime(value) {
			this.handle.endTime = value;
		}

		Id() {
			return this.handle.id;
		}

		SetId(value) {
			this.handle.id = value;
		}

		PauseOnExit() {
			return this.handle.pauseOnExit;
		}

		SetPauseOnExit(value) {
			this.handle.pauseOnExit = value;
		}

		StartTime() {
			return this.handle.startTime;
		}

		SetStartTime(value) {
			this.handle.startTime = value;
		}

		Track() {
			return dom_wrap(dom.TextTrack, this.handle.track);
		}
}

class dom_C_TextTrackCueList extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfTextTrackCueList() {
			
		}

		GetCueById(p0) {
			return dom_wrap(dom.TextTrackCue, this.handle.getCueById(p0));
		}

		Length() {
			return this.handle.length;
		}
}

class dom_C_TextTrackList extends dom_C_EventTarget {

		KindOfTextTrackList() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		GetTrackById(p0) {
			return dom_wrap(dom.TextTrack, this.handle.getTrackById(p0));
		}

		OnAddtrack(p0) {
			this.handle.addEventListener("addtrack", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TrackEvent, a[0])])));
		}

		OnChange(p0) {
			this.handle.addEventListener("change", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnRemovetrack(p0) {
			this.handle.addEventListener("removetrack", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TrackEvent, a[0])])));
		}

		Onaddtrack(p0) {
			return this.handle.onaddtrack(dom_unwrap(p0));
		}

		Onchange(p0) {
			return this.handle.onchange(dom_unwrap(p0));
		}

		Onremovetrack(p0) {
			return this.handle.onremovetrack(dom_unwrap(p0));
		}

		Length() {
			return this.handle.length;
		}
}

class dom_C_TimeRanges extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfTimeRanges() {
			
		}

		End(p0) {
			return this.handle.end(p0);
		}

		Start(p0) {
			return this.handle.start(p0);
		}

		Length() {
			return this.handle.length;
		}
}

class dom_C_TouchEvent extends dom_C_UIEvent {

		KindOfTouchEvent() {
			
		}

		KindOfUIEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsUIEvent() {
			return this;
		}

		AsEvent() {
			return this;
		}

		AltKey() {
			return this.handle.altKey;
		}

		ChangedTouches() {
			return dom_wrap(dom.TouchList, this.handle.changedTouches);
		}

		CtrlKey() {
			return this.handle.ctrlKey;
		}

		MetaKey() {
			return this.handle.metaKey;
		}

		ShiftKey() {
			return this.handle.shiftKey;
		}

		TargetTouches() {
			return dom_wrap(dom.TouchList, this.handle.targetTouches);
		}

		Touches() {
			return dom_wrap(dom.TouchList, this.handle.touches);
		}
}

class dom_C_TouchEventType extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfTouchEventType() {
			
		}

		Prototype() {
			return dom_wrap(dom.TouchEvent, this.handle.prototype);
		}

		SetPrototype(value) {
			this.handle.prototype = dom_unwrap(value);
		}
}

class dom_C_TouchList extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfTouchList() {
			
		}

		Item(p0) {
			return dom_fromTouch(this.handle.item(p0));
		}

		Length() {
			return this.handle.length;
		}
}

class dom_C_TrackEvent extends dom_C_Event {

		KindOfTrackEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		Track() {
			return dom_wrap(dom.TextTrack, this.handle.track);
		}
}

class dom_C_TransitionEvent extends dom_C_Event {

		KindOfTransitionEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsEvent() {
			return this;
		}

		ElapsedTime() {
			return this.handle.elapsedTime;
		}

		PropertyName() {
			return this.handle.propertyName;
		}

		PseudoElement() {
			return this.handle.pseudoElement;
		}
}

class dom_C_TreeWalker extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfTreeWalker() {
			
		}

		FirstChild() {
			return dom_wrap(dom.Node, this.handle.firstChild());
		}

		LastChild() {
			return dom_wrap(dom.Node, this.handle.lastChild());
		}

		NextNode() {
			return dom_wrap(dom.Node, this.handle.nextNode());
		}

		NextSibling() {
			return dom_wrap(dom.Node, this.handle.nextSibling());
		}

		ParentNode() {
			return dom_wrap(dom.Node, this.handle.parentNode());
		}

		PreviousNode() {
			return dom_wrap(dom.Node, this.handle.previousNode());
		}

		PreviousSibling() {
			return dom_wrap(dom.Node, this.handle.previousSibling());
		}

		CurrentNode() {
			return dom_wrap(dom.Node, this.handle.currentNode);
		}

		SetCurrentNode(value) {
			this.handle.currentNode = dom_unwrap(value);
		}

		Filter() {
			return this.handle.filter;
		}

		Root() {
			return dom_wrap(dom.Node, this.handle.root);
		}

		WhatToShow() {
			return this.handle.whatToShow;
		}
}

class dom_C_ViewTransition extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfViewTransition() {
			
		}

		SkipTransition() {
			this.handle.skipTransition();
		}
}

class dom_C_VisualViewport extends dom_C_EventTarget {

		KindOfVisualViewport() {
			
		}

		KindOfEventTarget() {
			
		}

		AsEventTarget() {
			return this;
		}

		OnResize(p0) {
			this.handle.addEventListener("resize", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnScroll(p0) {
			this.handle.addEventListener("scroll", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		Onresize(p0) {
			return this.handle.onresize(dom_unwrap(p0));
		}

		Onscroll(p0) {
			return this.handle.onscroll(dom_unwrap(p0));
		}

		Height() {
			return this.handle.height;
		}

		OffsetLeft() {
			return this.handle.offsetLeft;
		}

		OffsetTop() {
			return this.handle.offsetTop;
		}

		PageLeft() {
			return this.handle.pageLeft;
		}

		PageTop() {
			return this.handle.pageTop;
		}

		Scale() {
			return this.handle.scale;
		}

		Width() {
			return this.handle.width;
		}
}

class dom_C_WakeLock extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfWakeLock() {
			
		}

		*Request(...p0) {
			const outcome = yield* dom_await(() => this.handle.request(...p0.map((item) => dom_unwrapAny(item))));
			
			return outcome.ok ? [outcome.value, null] : [null, outcome.error];
		}
}

class dom_C_WaveShaperNode extends dom_C_AudioNode {

		KindOfWaveShaperNode() {
			
		}

		KindOfAudioNode() {
			
		}

		KindOfEventTarget() {
			
		}

		AsAudioNode() {
			return this;
		}

		AsEventTarget() {
			return this;
		}

		Curve() {
			return Array.from(this.handle.curve || []);
		}

		SetCurve(value) {
			this.handle.curve = Float32Array.from(value || []);
		}

		Oversample() {
			return this.handle.oversample;
		}

		SetOversample(value) {
			this.handle.oversample = value;
		}
}

class dom_C_WheelEvent extends dom_C_MouseEvent {

		KindOfWheelEvent() {
			
		}

		KindOfMouseEvent() {
			
		}

		KindOfUIEvent() {
			
		}

		KindOfEvent() {
			
		}

		AsMouseEvent() {
			return this;
		}

		AsUIEvent() {
			return this;
		}

		AsEvent() {
			return this;
		}

		DeltaMode() {
			return this.handle.deltaMode;
		}

		DeltaX() {
			return this.handle.deltaX;
		}

		DeltaY() {
			return this.handle.deltaY;
		}

		DeltaZ() {
			return this.handle.deltaZ;
		}

		DOM_DELTA_PIXEL() {
			return this.handle.DOM_DELTA_PIXEL;
		}

		DOM_DELTA_LINE() {
			return this.handle.DOM_DELTA_LINE;
		}

		DOM_DELTA_PAGE() {
			return this.handle.DOM_DELTA_PAGE;
		}
}

class dom_C_WheelEventType extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfWheelEventType() {
			
		}

		Prototype() {
			return dom_wrap(dom.WheelEvent, this.handle.prototype);
		}

		SetPrototype(value) {
			this.handle.prototype = dom_unwrap(value);
		}

		DOM_DELTA_PIXEL() {
			return this.handle.DOM_DELTA_PIXEL;
		}

		DOM_DELTA_LINE() {
			return this.handle.DOM_DELTA_LINE;
		}

		DOM_DELTA_PAGE() {
			return this.handle.DOM_DELTA_PAGE;
		}
}

class dom_C_WindowLocalStorage extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfWindowLocalStorage() {
			
		}

		AsWindow() {
			return dom_is(this, "Window") ? this : null;
		}

		LocalStorage() {
			return dom_wrap(dom.Storage, this.handle.localStorage);
		}
}

class dom_C_WindowOrWorkerGlobalScope extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfWindowOrWorkerGlobalScope() {
			
		}

		AsWindow() {
			return dom_is(this, "Window") ? this : null;
		}

		Atob(p0) {
			return this.handle.atob(p0);
		}

		Btoa(p0) {
			return this.handle.btoa(p0);
		}

		ClearInterval(p0) {
			this.handle.clearInterval(p0);
		}

		ClearTimeout(p0) {
			this.handle.clearTimeout(p0);
		}

		*CreateImageBitmap(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			const outcome = yield* dom_await(() => this.handle.createImageBitmap(dom_unwrapAny(p0), ...p1.map((item) => dom_toImageBitmapOptions(item))));
			
			return outcome.ok ? [dom_wrap(dom.ImageBitmap, outcome.value), null] : [null, outcome.error];
		}

		*Fetch(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			const outcome = yield* dom_await(() => this.handle.fetch(dom_unwrapAny(p0), ...p1.map((item) => dom_toRequestInit(item))));
			
			return outcome.ok ? [dom_wrap(dom.Response, outcome.value), null] : [null, outcome.error];
		}

		QueueMicrotask(p0) {
			this.handle.queueMicrotask((p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [])));
		}

		ReportError(p0) {
			this.handle.reportError(dom_unwrapAny(p0));
		}

		SetInterval(p0, ...p1) {
			return this.handle.setInterval(dom_unwrapAny(p0), ...p1.map((item) => dom_unwrapAny(item)));
		}

		SetTimeout(p0, ...p1) {
			return this.handle.setTimeout(dom_unwrapAny(p0), ...p1.map((item) => dom_unwrapAny(item)));
		}

		StructuredClone(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			return this.handle.structuredClone(dom_unwrapAny(p0), ...p1.map((item) => dom_toStructuredSerializeOptions(item)));
		}

		Caches() {
			return dom_wrap(dom.CacheStorage, this.handle.caches);
		}

		CrossOriginIsolated() {
			return this.handle.crossOriginIsolated;
		}

		Crypto() {
			return dom_wrap(dom.Crypto, this.handle.crypto);
		}

		IndexedDB() {
			return dom_wrap(dom.IDBFactory, this.handle.indexedDB);
		}

		IsSecureContext() {
			return this.handle.isSecureContext;
		}

		Origin() {
			return this.handle.origin;
		}

		Performance() {
			return dom_wrap(dom.Performance, this.handle.performance);
		}
}

class dom_C_WindowSessionStorage extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfWindowSessionStorage() {
			
		}

		AsWindow() {
			return dom_is(this, "Window") ? this : null;
		}

		SessionStorage() {
			return dom_wrap(dom.Storage, this.handle.sessionStorage);
		}
}

class dom_C_Window extends dom_C_EventTarget {

		KindOfWindow() {
			
		}

		KindOfEventTarget() {
			
		}

		KindOfAnimationFrameProvider() {
			
		}

		KindOfGlobalEventHandlers() {
			
		}

		KindOfWindowEventHandlers() {
			
		}

		KindOfWindowLocalStorage() {
			
		}

		KindOfWindowOrWorkerGlobalScope() {
			
		}

		KindOfWindowSessionStorage() {
			
		}

		AsEventTarget() {
			return this;
		}

		AsAnimationFrameProvider() {
			return this;
		}

		AsGlobalEventHandlers() {
			return this;
		}

		AsWindowEventHandlers() {
			return this;
		}

		AsWindowLocalStorage() {
			return this;
		}

		AsWindowOrWorkerGlobalScope() {
			return this;
		}

		AsWindowSessionStorage() {
			return this;
		}

		Alert(...p0) {
			this.handle.alert(...p0.map((item) => dom_unwrapAny(item)));
		}

		Blur() {
			this.handle.blur();
		}

		CancelIdleCallback(p0) {
			this.handle.cancelIdleCallback(p0);
		}

		CaptureEvents() {
			this.handle.captureEvents();
		}

		Close() {
			this.handle.close();
		}

		Confirm(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return this.handle.confirm(...p0.map((item) => item));
		}

		Focus() {
			this.handle.focus();
		}

		GetComputedStyle(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			return dom_wrap(dom.CSSStyleDeclaration, this.handle.getComputedStyle(dom_unwrap(p0), ...p1.map((item) => item)));
		}

		GetSelection() {
			return dom_wrap(dom.Selection, this.handle.getSelection());
		}

		MatchMedia(p0) {
			return dom_wrap(dom.MediaQueryList, this.handle.matchMedia(p0));
		}

		MoveBy(p0, p1) {
			this.handle.moveBy(p0, p1);
		}

		MoveTo(p0, p1) {
			this.handle.moveTo(p0, p1);
		}

		Open(...p0) {
			return dom_wrap(dom.Window, this.handle.open(...p0.map((item) => dom_unwrapAny(item))));
		}

		PostMessage(p0, p1, ...p2) {
			this.handle.postMessage(dom_unwrapAny(p0), p1, ...p2.map((item) => Array.from(item || [], (item) => dom_unwrapAny(item))));
		}

		Print() {
			this.handle.print();
		}

		Prompt(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			return this.handle.prompt(...p0.map((item) => item));
		}

		ReleaseEvents() {
			this.handle.releaseEvents();
		}

		RequestIdleCallback(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			return this.handle.requestIdleCallback((p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.IdleDeadline, a[0])])), ...p1.map((item) => dom_toIdleRequestOptions(item)));
		}

		ResizeBy(p0, p1) {
			this.handle.resizeBy(p0, p1);
		}

		ResizeTo(p0, p1) {
			this.handle.resizeTo(p0, p1);
		}

		Scroll(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.scroll(...p0.map((item) => dom_toScrollToOptions(item)));
		}

		ScrollBy(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.scrollBy(...p0.map((item) => dom_toScrollToOptions(item)));
		}

		ScrollTo(...p0) {
			if (p0.length === 1 && Array.isArray(p0[0]) && p0[0].length === 0) {
				p0 = [];
			}
			
			this.handle.scrollTo(...p0.map((item) => dom_toScrollToOptions(item)));
		}

		Stop() {
			this.handle.stop();
		}

		CancelAnimationFrame(p0) {
			this.handle.cancelAnimationFrame(p0);
		}

		RequestAnimationFrame(p0) {
			return this.handle.requestAnimationFrame((p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [a[0]])));
		}

		Atob(p0) {
			return this.handle.atob(p0);
		}

		Btoa(p0) {
			return this.handle.btoa(p0);
		}

		ClearInterval(p0) {
			this.handle.clearInterval(p0);
		}

		ClearTimeout(p0) {
			this.handle.clearTimeout(p0);
		}

		*CreateImageBitmap(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			const outcome = yield* dom_await(() => this.handle.createImageBitmap(dom_unwrapAny(p0), ...p1.map((item) => dom_toImageBitmapOptions(item))));
			
			return outcome.ok ? [dom_wrap(dom.ImageBitmap, outcome.value), null] : [null, outcome.error];
		}

		*Fetch(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			const outcome = yield* dom_await(() => this.handle.fetch(dom_unwrapAny(p0), ...p1.map((item) => dom_toRequestInit(item))));
			
			return outcome.ok ? [dom_wrap(dom.Response, outcome.value), null] : [null, outcome.error];
		}

		QueueMicrotask(p0) {
			this.handle.queueMicrotask((p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [])));
		}

		ReportError(p0) {
			this.handle.reportError(dom_unwrapAny(p0));
		}

		SetInterval(p0, ...p1) {
			return this.handle.setInterval(dom_unwrapAny(p0), ...p1.map((item) => dom_unwrapAny(item)));
		}

		SetTimeout(p0, ...p1) {
			return this.handle.setTimeout(dom_unwrapAny(p0), ...p1.map((item) => dom_unwrapAny(item)));
		}

		StructuredClone(p0, ...p1) {
			if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
				p1 = [];
			}
			
			return this.handle.structuredClone(dom_unwrapAny(p0), ...p1.map((item) => dom_toStructuredSerializeOptions(item)));
		}

		OnDOMContentLoaded(p0) {
			this.handle.addEventListener("DOMContentLoaded", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnDevicemotion(p0) {
			this.handle.addEventListener("devicemotion", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DeviceMotionEvent, a[0])])));
		}

		OnDeviceorientation(p0) {
			this.handle.addEventListener("deviceorientation", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DeviceOrientationEvent, a[0])])));
		}

		OnDeviceorientationabsolute(p0) {
			this.handle.addEventListener("deviceorientationabsolute", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DeviceOrientationEvent, a[0])])));
		}

		OnGamepadconnected(p0) {
			this.handle.addEventListener("gamepadconnected", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.GamepadEvent, a[0])])));
		}

		OnGamepaddisconnected(p0) {
			this.handle.addEventListener("gamepaddisconnected", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.GamepadEvent, a[0])])));
		}

		OnOrientationchange(p0) {
			this.handle.addEventListener("orientationchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnAbort(p0) {
			this.handle.addEventListener("abort", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.UIEvent, a[0])])));
		}

		OnAnimationcancel(p0) {
			this.handle.addEventListener("animationcancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationEvent, a[0])])));
		}

		OnAnimationend(p0) {
			this.handle.addEventListener("animationend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationEvent, a[0])])));
		}

		OnAnimationiteration(p0) {
			this.handle.addEventListener("animationiteration", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationEvent, a[0])])));
		}

		OnAnimationstart(p0) {
			this.handle.addEventListener("animationstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.AnimationEvent, a[0])])));
		}

		OnAuxclick(p0) {
			this.handle.addEventListener("auxclick", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnBeforeinput(p0) {
			this.handle.addEventListener("beforeinput", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.InputEvent, a[0])])));
		}

		OnBeforetoggle(p0) {
			this.handle.addEventListener("beforetoggle", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnBlur(p0) {
			this.handle.addEventListener("blur", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FocusEvent, a[0])])));
		}

		OnCancel(p0) {
			this.handle.addEventListener("cancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCanplay(p0) {
			this.handle.addEventListener("canplay", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCanplaythrough(p0) {
			this.handle.addEventListener("canplaythrough", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnChange(p0) {
			this.handle.addEventListener("change", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnClick(p0) {
			this.handle.addEventListener("click", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnClose(p0) {
			this.handle.addEventListener("close", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCompositionend(p0) {
			this.handle.addEventListener("compositionend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.CompositionEvent, a[0])])));
		}

		OnCompositionstart(p0) {
			this.handle.addEventListener("compositionstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.CompositionEvent, a[0])])));
		}

		OnCompositionupdate(p0) {
			this.handle.addEventListener("compositionupdate", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.CompositionEvent, a[0])])));
		}

		OnContextlost(p0) {
			this.handle.addEventListener("contextlost", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnContextmenu(p0) {
			this.handle.addEventListener("contextmenu", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnContextrestored(p0) {
			this.handle.addEventListener("contextrestored", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCopy(p0) {
			this.handle.addEventListener("copy", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ClipboardEvent, a[0])])));
		}

		OnCuechange(p0) {
			this.handle.addEventListener("cuechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnCut(p0) {
			this.handle.addEventListener("cut", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ClipboardEvent, a[0])])));
		}

		OnDblclick(p0) {
			this.handle.addEventListener("dblclick", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnDrag(p0) {
			this.handle.addEventListener("drag", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragend(p0) {
			this.handle.addEventListener("dragend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragenter(p0) {
			this.handle.addEventListener("dragenter", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragleave(p0) {
			this.handle.addEventListener("dragleave", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragover(p0) {
			this.handle.addEventListener("dragover", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDragstart(p0) {
			this.handle.addEventListener("dragstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDrop(p0) {
			this.handle.addEventListener("drop", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.DragEvent, a[0])])));
		}

		OnDurationchange(p0) {
			this.handle.addEventListener("durationchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnEmptied(p0) {
			this.handle.addEventListener("emptied", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnEnded(p0) {
			this.handle.addEventListener("ended", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnError(p0) {
			this.handle.addEventListener("error", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ErrorEvent, a[0])])));
		}

		OnFocus(p0) {
			this.handle.addEventListener("focus", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FocusEvent, a[0])])));
		}

		OnFocusin(p0) {
			this.handle.addEventListener("focusin", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FocusEvent, a[0])])));
		}

		OnFocusout(p0) {
			this.handle.addEventListener("focusout", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FocusEvent, a[0])])));
		}

		OnFormdata(p0) {
			this.handle.addEventListener("formdata", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.FormDataEvent, a[0])])));
		}

		OnGotpointercapture(p0) {
			this.handle.addEventListener("gotpointercapture", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnInput(p0) {
			this.handle.addEventListener("input", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnInvalid(p0) {
			this.handle.addEventListener("invalid", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnKeydown(p0) {
			this.handle.addEventListener("keydown", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.KeyboardEvent, a[0])])));
		}

		OnKeypress(p0) {
			this.handle.addEventListener("keypress", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.KeyboardEvent, a[0])])));
		}

		OnKeyup(p0) {
			this.handle.addEventListener("keyup", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.KeyboardEvent, a[0])])));
		}

		OnLoad(p0) {
			this.handle.addEventListener("load", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLoadeddata(p0) {
			this.handle.addEventListener("loadeddata", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLoadedmetadata(p0) {
			this.handle.addEventListener("loadedmetadata", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLoadstart(p0) {
			this.handle.addEventListener("loadstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnLostpointercapture(p0) {
			this.handle.addEventListener("lostpointercapture", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnMousedown(p0) {
			this.handle.addEventListener("mousedown", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseenter(p0) {
			this.handle.addEventListener("mouseenter", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseleave(p0) {
			this.handle.addEventListener("mouseleave", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMousemove(p0) {
			this.handle.addEventListener("mousemove", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseout(p0) {
			this.handle.addEventListener("mouseout", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseover(p0) {
			this.handle.addEventListener("mouseover", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnMouseup(p0) {
			this.handle.addEventListener("mouseup", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MouseEvent, a[0])])));
		}

		OnPaste(p0) {
			this.handle.addEventListener("paste", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ClipboardEvent, a[0])])));
		}

		OnPause(p0) {
			this.handle.addEventListener("pause", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnPlay(p0) {
			this.handle.addEventListener("play", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnPlaying(p0) {
			this.handle.addEventListener("playing", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnPointercancel(p0) {
			this.handle.addEventListener("pointercancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerdown(p0) {
			this.handle.addEventListener("pointerdown", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerenter(p0) {
			this.handle.addEventListener("pointerenter", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerleave(p0) {
			this.handle.addEventListener("pointerleave", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointermove(p0) {
			this.handle.addEventListener("pointermove", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerout(p0) {
			this.handle.addEventListener("pointerout", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerover(p0) {
			this.handle.addEventListener("pointerover", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnPointerup(p0) {
			this.handle.addEventListener("pointerup", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PointerEvent, a[0])])));
		}

		OnProgress(p0) {
			this.handle.addEventListener("progress", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.ProgressEvent, a[0])])));
		}

		OnRatechange(p0) {
			this.handle.addEventListener("ratechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnReset(p0) {
			this.handle.addEventListener("reset", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnResize(p0) {
			this.handle.addEventListener("resize", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.UIEvent, a[0])])));
		}

		OnScroll(p0) {
			this.handle.addEventListener("scroll", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnScrollend(p0) {
			this.handle.addEventListener("scrollend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSecuritypolicyviolation(p0) {
			this.handle.addEventListener("securitypolicyviolation", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.SecurityPolicyViolationEvent, a[0])])));
		}

		OnSeeked(p0) {
			this.handle.addEventListener("seeked", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSeeking(p0) {
			this.handle.addEventListener("seeking", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSelect(p0) {
			this.handle.addEventListener("select", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSelectionchange(p0) {
			this.handle.addEventListener("selectionchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSelectstart(p0) {
			this.handle.addEventListener("selectstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSlotchange(p0) {
			this.handle.addEventListener("slotchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnStalled(p0) {
			this.handle.addEventListener("stalled", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnSubmit(p0) {
			this.handle.addEventListener("submit", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.SubmitEvent, a[0])])));
		}

		OnSuspend(p0) {
			this.handle.addEventListener("suspend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnTimeupdate(p0) {
			this.handle.addEventListener("timeupdate", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnToggle(p0) {
			this.handle.addEventListener("toggle", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnTouchcancel(p0) {
			this.handle.addEventListener("touchcancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TouchEvent, a[0])])));
		}

		OnTouchend(p0) {
			this.handle.addEventListener("touchend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TouchEvent, a[0])])));
		}

		OnTouchmove(p0) {
			this.handle.addEventListener("touchmove", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TouchEvent, a[0])])));
		}

		OnTouchstart(p0) {
			this.handle.addEventListener("touchstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TouchEvent, a[0])])));
		}

		OnTransitioncancel(p0) {
			this.handle.addEventListener("transitioncancel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TransitionEvent, a[0])])));
		}

		OnTransitionend(p0) {
			this.handle.addEventListener("transitionend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TransitionEvent, a[0])])));
		}

		OnTransitionrun(p0) {
			this.handle.addEventListener("transitionrun", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TransitionEvent, a[0])])));
		}

		OnTransitionstart(p0) {
			this.handle.addEventListener("transitionstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.TransitionEvent, a[0])])));
		}

		OnVolumechange(p0) {
			this.handle.addEventListener("volumechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWaiting(p0) {
			this.handle.addEventListener("waiting", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWebkitanimationend(p0) {
			this.handle.addEventListener("webkitanimationend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWebkitanimationiteration(p0) {
			this.handle.addEventListener("webkitanimationiteration", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWebkitanimationstart(p0) {
			this.handle.addEventListener("webkitanimationstart", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWebkittransitionend(p0) {
			this.handle.addEventListener("webkittransitionend", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnWheel(p0) {
			this.handle.addEventListener("wheel", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.WheelEvent, a[0])])));
		}

		OnAfterprint(p0) {
			this.handle.addEventListener("afterprint", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnBeforeprint(p0) {
			this.handle.addEventListener("beforeprint", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnBeforeunload(p0) {
			this.handle.addEventListener("beforeunload", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.BeforeUnloadEvent, a[0])])));
		}

		OnHashchange(p0) {
			this.handle.addEventListener("hashchange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.HashChangeEvent, a[0])])));
		}

		OnLanguagechange(p0) {
			this.handle.addEventListener("languagechange", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnMessage(p0) {
			this.handle.addEventListener("message", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MessageEvent, a[0])])));
		}

		OnMessageerror(p0) {
			this.handle.addEventListener("messageerror", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.MessageEvent, a[0])])));
		}

		OnOffline(p0) {
			this.handle.addEventListener("offline", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnOnline(p0) {
			this.handle.addEventListener("online", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		OnPagehide(p0) {
			this.handle.addEventListener("pagehide", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PageTransitionEvent, a[0])])));
		}

		OnPageshow(p0) {
			this.handle.addEventListener("pageshow", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PageTransitionEvent, a[0])])));
		}

		OnPopstate(p0) {
			this.handle.addEventListener("popstate", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PopStateEvent, a[0])])));
		}

		OnRejectionhandled(p0) {
			this.handle.addEventListener("rejectionhandled", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PromiseRejectionEvent, a[0])])));
		}

		OnStorage(p0) {
			this.handle.addEventListener("storage", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.StorageEvent, a[0])])));
		}

		OnUnhandledrejection(p0) {
			this.handle.addEventListener("unhandledrejection", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.PromiseRejectionEvent, a[0])])));
		}

		OnUnload(p0) {
			this.handle.addEventListener("unload", (p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [dom_wrap(dom.Event, a[0])])));
		}

		Ondevicemotion(p0) {
			return this.handle.ondevicemotion(dom_unwrap(p0));
		}

		Ondeviceorientation(p0) {
			return this.handle.ondeviceorientation(dom_unwrap(p0));
		}

		Ondeviceorientationabsolute(p0) {
			return this.handle.ondeviceorientationabsolute(dom_unwrap(p0));
		}

		Onorientationchange(p0) {
			return this.handle.onorientationchange(dom_unwrap(p0));
		}

		Onabort(p0) {
			return this.handle.onabort(dom_unwrap(p0));
		}

		Onanimationcancel(p0) {
			return this.handle.onanimationcancel(dom_unwrap(p0));
		}

		Onanimationend(p0) {
			return this.handle.onanimationend(dom_unwrap(p0));
		}

		Onanimationiteration(p0) {
			return this.handle.onanimationiteration(dom_unwrap(p0));
		}

		Onanimationstart(p0) {
			return this.handle.onanimationstart(dom_unwrap(p0));
		}

		Onauxclick(p0) {
			return this.handle.onauxclick(dom_unwrap(p0));
		}

		Onbeforeinput(p0) {
			return this.handle.onbeforeinput(dom_unwrap(p0));
		}

		Onbeforetoggle(p0) {
			return this.handle.onbeforetoggle(dom_unwrap(p0));
		}

		Onblur(p0) {
			return this.handle.onblur(dom_unwrap(p0));
		}

		Oncancel(p0) {
			return this.handle.oncancel(dom_unwrap(p0));
		}

		Oncanplay(p0) {
			return this.handle.oncanplay(dom_unwrap(p0));
		}

		Oncanplaythrough(p0) {
			return this.handle.oncanplaythrough(dom_unwrap(p0));
		}

		Onchange(p0) {
			return this.handle.onchange(dom_unwrap(p0));
		}

		Onclick(p0) {
			return this.handle.onclick(dom_unwrap(p0));
		}

		Onclose(p0) {
			return this.handle.onclose(dom_unwrap(p0));
		}

		Oncontextlost(p0) {
			return this.handle.oncontextlost(dom_unwrap(p0));
		}

		Oncontextmenu(p0) {
			return this.handle.oncontextmenu(dom_unwrap(p0));
		}

		Oncontextrestored(p0) {
			return this.handle.oncontextrestored(dom_unwrap(p0));
		}

		Oncopy(p0) {
			return this.handle.oncopy(dom_unwrap(p0));
		}

		Oncuechange(p0) {
			return this.handle.oncuechange(dom_unwrap(p0));
		}

		Oncut(p0) {
			return this.handle.oncut(dom_unwrap(p0));
		}

		Ondblclick(p0) {
			return this.handle.ondblclick(dom_unwrap(p0));
		}

		Ondrag(p0) {
			return this.handle.ondrag(dom_unwrap(p0));
		}

		Ondragend(p0) {
			return this.handle.ondragend(dom_unwrap(p0));
		}

		Ondragenter(p0) {
			return this.handle.ondragenter(dom_unwrap(p0));
		}

		Ondragleave(p0) {
			return this.handle.ondragleave(dom_unwrap(p0));
		}

		Ondragover(p0) {
			return this.handle.ondragover(dom_unwrap(p0));
		}

		Ondragstart(p0) {
			return this.handle.ondragstart(dom_unwrap(p0));
		}

		Ondrop(p0) {
			return this.handle.ondrop(dom_unwrap(p0));
		}

		Ondurationchange(p0) {
			return this.handle.ondurationchange(dom_unwrap(p0));
		}

		Onemptied(p0) {
			return this.handle.onemptied(dom_unwrap(p0));
		}

		Onended(p0) {
			return this.handle.onended(dom_unwrap(p0));
		}

		Onfocus(p0) {
			return this.handle.onfocus(dom_unwrap(p0));
		}

		Onformdata(p0) {
			return this.handle.onformdata(dom_unwrap(p0));
		}

		Ongotpointercapture(p0) {
			return this.handle.ongotpointercapture(dom_unwrap(p0));
		}

		Oninput(p0) {
			return this.handle.oninput(dom_unwrap(p0));
		}

		Oninvalid(p0) {
			return this.handle.oninvalid(dom_unwrap(p0));
		}

		Onkeydown(p0) {
			return this.handle.onkeydown(dom_unwrap(p0));
		}

		Onkeypress(p0) {
			return this.handle.onkeypress(dom_unwrap(p0));
		}

		Onkeyup(p0) {
			return this.handle.onkeyup(dom_unwrap(p0));
		}

		Onload(p0) {
			return this.handle.onload(dom_unwrap(p0));
		}

		Onloadeddata(p0) {
			return this.handle.onloadeddata(dom_unwrap(p0));
		}

		Onloadedmetadata(p0) {
			return this.handle.onloadedmetadata(dom_unwrap(p0));
		}

		Onloadstart(p0) {
			return this.handle.onloadstart(dom_unwrap(p0));
		}

		Onlostpointercapture(p0) {
			return this.handle.onlostpointercapture(dom_unwrap(p0));
		}

		Onmousedown(p0) {
			return this.handle.onmousedown(dom_unwrap(p0));
		}

		Onmouseenter(p0) {
			return this.handle.onmouseenter(dom_unwrap(p0));
		}

		Onmouseleave(p0) {
			return this.handle.onmouseleave(dom_unwrap(p0));
		}

		Onmousemove(p0) {
			return this.handle.onmousemove(dom_unwrap(p0));
		}

		Onmouseout(p0) {
			return this.handle.onmouseout(dom_unwrap(p0));
		}

		Onmouseover(p0) {
			return this.handle.onmouseover(dom_unwrap(p0));
		}

		Onmouseup(p0) {
			return this.handle.onmouseup(dom_unwrap(p0));
		}

		Onpaste(p0) {
			return this.handle.onpaste(dom_unwrap(p0));
		}

		Onpause(p0) {
			return this.handle.onpause(dom_unwrap(p0));
		}

		Onplay(p0) {
			return this.handle.onplay(dom_unwrap(p0));
		}

		Onplaying(p0) {
			return this.handle.onplaying(dom_unwrap(p0));
		}

		Onpointercancel(p0) {
			return this.handle.onpointercancel(dom_unwrap(p0));
		}

		Onpointerdown(p0) {
			return this.handle.onpointerdown(dom_unwrap(p0));
		}

		Onpointerenter(p0) {
			return this.handle.onpointerenter(dom_unwrap(p0));
		}

		Onpointerleave(p0) {
			return this.handle.onpointerleave(dom_unwrap(p0));
		}

		Onpointermove(p0) {
			return this.handle.onpointermove(dom_unwrap(p0));
		}

		Onpointerout(p0) {
			return this.handle.onpointerout(dom_unwrap(p0));
		}

		Onpointerover(p0) {
			return this.handle.onpointerover(dom_unwrap(p0));
		}

		Onpointerup(p0) {
			return this.handle.onpointerup(dom_unwrap(p0));
		}

		Onprogress(p0) {
			return this.handle.onprogress(dom_unwrap(p0));
		}

		Onratechange(p0) {
			return this.handle.onratechange(dom_unwrap(p0));
		}

		Onreset(p0) {
			return this.handle.onreset(dom_unwrap(p0));
		}

		Onresize(p0) {
			return this.handle.onresize(dom_unwrap(p0));
		}

		Onscroll(p0) {
			return this.handle.onscroll(dom_unwrap(p0));
		}

		Onscrollend(p0) {
			return this.handle.onscrollend(dom_unwrap(p0));
		}

		Onsecuritypolicyviolation(p0) {
			return this.handle.onsecuritypolicyviolation(dom_unwrap(p0));
		}

		Onseeked(p0) {
			return this.handle.onseeked(dom_unwrap(p0));
		}

		Onseeking(p0) {
			return this.handle.onseeking(dom_unwrap(p0));
		}

		Onselect(p0) {
			return this.handle.onselect(dom_unwrap(p0));
		}

		Onselectionchange(p0) {
			return this.handle.onselectionchange(dom_unwrap(p0));
		}

		Onselectstart(p0) {
			return this.handle.onselectstart(dom_unwrap(p0));
		}

		Onslotchange(p0) {
			return this.handle.onslotchange(dom_unwrap(p0));
		}

		Onstalled(p0) {
			return this.handle.onstalled(dom_unwrap(p0));
		}

		Onsubmit(p0) {
			return this.handle.onsubmit(dom_unwrap(p0));
		}

		Onsuspend(p0) {
			return this.handle.onsuspend(dom_unwrap(p0));
		}

		Ontimeupdate(p0) {
			return this.handle.ontimeupdate(dom_unwrap(p0));
		}

		Ontoggle(p0) {
			return this.handle.ontoggle(dom_unwrap(p0));
		}

		Ontouchcancel(p0) {
			return this.handle.ontouchcancel(dom_unwrap(p0));
		}

		Ontouchend(p0) {
			return this.handle.ontouchend(dom_unwrap(p0));
		}

		Ontouchmove(p0) {
			return this.handle.ontouchmove(dom_unwrap(p0));
		}

		Ontouchstart(p0) {
			return this.handle.ontouchstart(dom_unwrap(p0));
		}

		Ontransitioncancel(p0) {
			return this.handle.ontransitioncancel(dom_unwrap(p0));
		}

		Ontransitionend(p0) {
			return this.handle.ontransitionend(dom_unwrap(p0));
		}

		Ontransitionrun(p0) {
			return this.handle.ontransitionrun(dom_unwrap(p0));
		}

		Ontransitionstart(p0) {
			return this.handle.ontransitionstart(dom_unwrap(p0));
		}

		Onvolumechange(p0) {
			return this.handle.onvolumechange(dom_unwrap(p0));
		}

		Onwaiting(p0) {
			return this.handle.onwaiting(dom_unwrap(p0));
		}

		Onwebkitanimationend(p0) {
			return this.handle.onwebkitanimationend(dom_unwrap(p0));
		}

		Onwebkitanimationiteration(p0) {
			return this.handle.onwebkitanimationiteration(dom_unwrap(p0));
		}

		Onwebkitanimationstart(p0) {
			return this.handle.onwebkitanimationstart(dom_unwrap(p0));
		}

		Onwebkittransitionend(p0) {
			return this.handle.onwebkittransitionend(dom_unwrap(p0));
		}

		Onwheel(p0) {
			return this.handle.onwheel(dom_unwrap(p0));
		}

		Onafterprint(p0) {
			return this.handle.onafterprint(dom_unwrap(p0));
		}

		Onbeforeprint(p0) {
			return this.handle.onbeforeprint(dom_unwrap(p0));
		}

		Onbeforeunload(p0) {
			return this.handle.onbeforeunload(dom_unwrap(p0));
		}

		Ongamepadconnected(p0) {
			return this.handle.ongamepadconnected(dom_unwrap(p0));
		}

		Ongamepaddisconnected(p0) {
			return this.handle.ongamepaddisconnected(dom_unwrap(p0));
		}

		Onhashchange(p0) {
			return this.handle.onhashchange(dom_unwrap(p0));
		}

		Onlanguagechange(p0) {
			return this.handle.onlanguagechange(dom_unwrap(p0));
		}

		Onmessage(p0) {
			return this.handle.onmessage(dom_unwrap(p0));
		}

		Onmessageerror(p0) {
			return this.handle.onmessageerror(dom_unwrap(p0));
		}

		Onoffline(p0) {
			return this.handle.onoffline(dom_unwrap(p0));
		}

		Ononline(p0) {
			return this.handle.ononline(dom_unwrap(p0));
		}

		Onpagehide(p0) {
			return this.handle.onpagehide(dom_unwrap(p0));
		}

		Onpageshow(p0) {
			return this.handle.onpageshow(dom_unwrap(p0));
		}

		Onpopstate(p0) {
			return this.handle.onpopstate(dom_unwrap(p0));
		}

		Onrejectionhandled(p0) {
			return this.handle.onrejectionhandled(dom_unwrap(p0));
		}

		Onstorage(p0) {
			return this.handle.onstorage(dom_unwrap(p0));
		}

		Onunhandledrejection(p0) {
			return this.handle.onunhandledrejection(dom_unwrap(p0));
		}

		Onunload(p0) {
			return this.handle.onunload(dom_unwrap(p0));
		}

		ClientInformation() {
			return dom_wrap(dom.Navigator, this.handle.clientInformation);
		}

		Closed() {
			return this.handle.closed;
		}

		CustomElements() {
			return dom_wrap(dom.CustomElementRegistry, this.handle.customElements);
		}

		DevicePixelRatio() {
			return this.handle.devicePixelRatio;
		}

		Document() {
			return dom_wrap(dom.Document, this.handle.document);
		}

		Event() {
			return dom_wrap(dom.Event, this.handle.event);
		}

		External() {
			return dom_wrap(dom.External, this.handle.external);
		}

		FrameElement() {
			return dom_wrap(dom.Element, this.handle.frameElement);
		}

		Frames() {
			return dom_wrap(dom.Window, this.handle.frames);
		}

		History() {
			return dom_wrap(dom.History, this.handle.history);
		}

		InnerHeight() {
			return this.handle.innerHeight;
		}

		InnerWidth() {
			return this.handle.innerWidth;
		}

		Length() {
			return this.handle.length;
		}

		Location() {
			return dom_wrap(dom.Location, this.handle.location);
		}

		SetLocation(value) {
			this.handle.location = dom_unwrap(value);
		}

		Locationbar() {
			return dom_fromBarProp(this.handle.locationbar);
		}

		Menubar() {
			return dom_fromBarProp(this.handle.menubar);
		}

		Name() {
			return this.handle.name;
		}

		SetName(value) {
			this.handle.name = value;
		}

		Navigator() {
			return dom_wrap(dom.Navigator, this.handle.navigator);
		}

		Opener() {
			return this.handle.opener;
		}

		SetOpener(value) {
			this.handle.opener = dom_unwrapAny(value);
		}

		Orientation() {
			return this.handle.orientation;
		}

		OuterHeight() {
			return this.handle.outerHeight;
		}

		OuterWidth() {
			return this.handle.outerWidth;
		}

		PageXOffset() {
			return this.handle.pageXOffset;
		}

		PageYOffset() {
			return this.handle.pageYOffset;
		}

		Parent() {
			return dom_wrap(dom.Window, this.handle.parent);
		}

		Personalbar() {
			return dom_fromBarProp(this.handle.personalbar);
		}

		Screen() {
			return dom_fromScreen(this.handle.screen);
		}

		ScreenLeft() {
			return this.handle.screenLeft;
		}

		ScreenTop() {
			return this.handle.screenTop;
		}

		ScreenX() {
			return this.handle.screenX;
		}

		ScreenY() {
			return this.handle.screenY;
		}

		ScrollX() {
			return this.handle.scrollX;
		}

		ScrollY() {
			return this.handle.scrollY;
		}

		Scrollbars() {
			return dom_fromBarProp(this.handle.scrollbars);
		}

		Self() {
			return dom_wrap(dom.Window, this.handle.self);
		}

		SpeechSynthesis() {
			return dom_wrap(dom.SpeechSynthesis, this.handle.speechSynthesis);
		}

		Status() {
			return this.handle.status;
		}

		SetStatus(value) {
			this.handle.status = value;
		}

		Statusbar() {
			return dom_fromBarProp(this.handle.statusbar);
		}

		Toolbar() {
			return dom_fromBarProp(this.handle.toolbar);
		}

		Top() {
			return dom_wrap(dom.Window, this.handle.top);
		}

		VisualViewport() {
			return dom_wrap(dom.VisualViewport, this.handle.visualViewport);
		}

		Window() {
			return dom_wrap(dom.Window, this.handle.window);
		}

		Onerror() {
			return this.handle.onerror;
		}

		SetOnerror(value) {
			this.handle.onerror = (value === null || value === undefined ? null : (...a) => dom_unwrapAny(go2jsCallNow(value, null, [a[0], a[1], a[2], a[3], a[4]])));
		}

		LocalStorage() {
			return dom_wrap(dom.Storage, this.handle.localStorage);
		}

		Caches() {
			return dom_wrap(dom.CacheStorage, this.handle.caches);
		}

		CrossOriginIsolated() {
			return this.handle.crossOriginIsolated;
		}

		Crypto() {
			return dom_wrap(dom.Crypto, this.handle.crypto);
		}

		IndexedDB() {
			return dom_wrap(dom.IDBFactory, this.handle.indexedDB);
		}

		IsSecureContext() {
			return this.handle.isSecureContext;
		}

		Origin() {
			return this.handle.origin;
		}

		Performance() {
			return dom_wrap(dom.Performance, this.handle.performance);
		}

		SessionStorage() {
			return dom_wrap(dom.Storage, this.handle.sessionStorage);
		}
}

class dom_C_XPathExpression extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfXPathExpression() {
			
		}

		Evaluate(p0, ...p1) {
			return dom_wrap(dom.XPathResult, this.handle.evaluate(dom_unwrap(p0), ...p1.map((item) => dom_unwrapAny(item))));
		}
}

class dom_C_XPathResult extends dom_base {
	constructor() {
		super();
		this.handle = null;
	}

		KindOfXPathResult() {
			
		}

		IterateNext() {
			return dom_wrap(dom.Node, this.handle.iterateNext());
		}

		SnapshotItem(p0) {
			return dom_wrap(dom.Node, this.handle.snapshotItem(p0));
		}

		BooleanValue() {
			return this.handle.booleanValue;
		}

		InvalidIteratorState() {
			return this.handle.invalidIteratorState;
		}

		NumberValue() {
			return this.handle.numberValue;
		}

		ResultType() {
			return this.handle.resultType;
		}

		SingleNodeValue() {
			return dom_wrap(dom.Node, this.handle.singleNodeValue);
		}

		SnapshotLength() {
			return this.handle.snapshotLength;
		}

		StringValue() {
			return this.handle.stringValue;
		}

		ANY_TYPE() {
			return this.handle.ANY_TYPE;
		}

		NUMBER_TYPE() {
			return this.handle.NUMBER_TYPE;
		}

		STRING_TYPE() {
			return this.handle.STRING_TYPE;
		}

		BOOLEAN_TYPE() {
			return this.handle.BOOLEAN_TYPE;
		}

		UNORDERED_NODE_ITERATOR_TYPE() {
			return this.handle.UNORDERED_NODE_ITERATOR_TYPE;
		}

		ORDERED_NODE_ITERATOR_TYPE() {
			return this.handle.ORDERED_NODE_ITERATOR_TYPE;
		}

		UNORDERED_NODE_SNAPSHOT_TYPE() {
			return this.handle.UNORDERED_NODE_SNAPSHOT_TYPE;
		}

		ORDERED_NODE_SNAPSHOT_TYPE() {
			return this.handle.ORDERED_NODE_SNAPSHOT_TYPE;
		}

		ANY_UNORDERED_NODE_TYPE() {
			return this.handle.ANY_UNORDERED_NODE_TYPE;
		}

		FIRST_ORDERED_NODE_TYPE() {
			return this.handle.FIRST_ORDERED_NODE_TYPE;
		}
}

const dom = {
	AssignedNodesOptions: class {
		constructor() {
			this.Flatten = false;
			this.Explicit = null;
		}
	},
	AudioTimestamp: class {
		constructor() {
			this.ContextTime = 0;
			this.PerformanceTime = 0;
			this.Explicit = null;
		}
	},
	CanvasRenderingContext2DSettings: class {
		constructor() {
			this.Alpha = false;
			this.ColorSpace = "";
			this.Desynchronized = false;
			this.WillReadFrequently = false;
			this.Explicit = null;
		}
	},
	CheckVisibilityOptions: class {
		constructor() {
			this.CheckOpacity = false;
			this.CheckVisibilityCSS = false;
			this.ContentVisibilityAuto = false;
			this.OpacityProperty = false;
			this.VisibilityProperty = false;
			this.Explicit = null;
		}
	},
	DOMMatrix2DInit: class {
		constructor() {
			this.A = 0;
			this.B = 0;
			this.C = 0;
			this.D = 0;
			this.E = 0;
			this.F = 0;
			this.M11 = 0;
			this.M12 = 0;
			this.M21 = 0;
			this.M22 = 0;
			this.M41 = 0;
			this.M42 = 0;
			this.Explicit = null;
		}
	},
	DOMMatrixInit: class {
		constructor() {
			this.Is2D = false;
			this.M13 = 0;
			this.M14 = 0;
			this.M23 = 0;
			this.M24 = 0;
			this.M31 = 0;
			this.M32 = 0;
			this.M33 = 0;
			this.M34 = 0;
			this.M43 = 0;
			this.M44 = 0;
			this.A = 0;
			this.B = 0;
			this.C = 0;
			this.D = 0;
			this.E = 0;
			this.F = 0;
			this.M11 = 0;
			this.M12 = 0;
			this.M21 = 0;
			this.M22 = 0;
			this.M41 = 0;
			this.M42 = 0;
			this.Explicit = null;
		}
	},
	DOMPointInit: class {
		constructor() {
			this.W = 0;
			this.X = 0;
			this.Y = 0;
			this.Z = 0;
			this.Explicit = null;
		}
	},
	ElementCreationOptions: class {
		constructor() {
			this.Is = "";
			this.Explicit = null;
		}
	},
	FocusOptions: class {
		constructor() {
			this.PreventScroll = false;
			this.Explicit = null;
		}
	},
	FullscreenOptions: class {
		constructor() {
			this.NavigationUI = "";
			this.Explicit = null;
		}
	},
	GetAnimationsOptions: class {
		constructor() {
			this.Subtree = false;
			this.Explicit = null;
		}
	},
	GetHTMLOptions: class {
		constructor() {
			this.SerializableShadowRoots = false;
			this.ShadowRoots = [];
			this.Explicit = null;
		}
	},
	GetRootNodeOptions: class {
		constructor() {
			this.Composed = false;
			this.Explicit = null;
		}
	},
	IdleRequestOptions: class {
		constructor() {
			this.Timeout = 0;
			this.Explicit = null;
		}
	},
	ImageBitmapOptions: class {
		constructor() {
			this.ColorSpaceConversion = "";
			this.ImageOrientation = "";
			this.PremultiplyAlpha = "";
			this.ResizeHeight = 0;
			this.ResizeQuality = "";
			this.ResizeWidth = 0;
			this.Explicit = null;
		}
	},
	ImageDataSettings: class {
		constructor() {
			this.ColorSpace = "";
			this.Explicit = null;
		}
	},
	ImageEncodeOptions: class {
		constructor() {
			this.Quality = 0;
			this.Type = "";
			this.Explicit = null;
		}
	},
	MIDIOptions: class {
		constructor() {
			this.Software = false;
			this.Sysex = false;
			this.Explicit = null;
		}
	},
	MediaKeySystemConfiguration: class {
		constructor() {
			this.AudioCapabilities = [];
			this.DistinctiveIdentifier = "";
			this.InitDataTypes = [];
			this.Label = "";
			this.PersistentState = "";
			this.SessionTypes = [];
			this.VideoCapabilities = [];
			this.Explicit = null;
		}
	},
	MediaKeysPolicy: class {
		constructor() {
			this.MinHdcpVersion = "";
			this.Explicit = null;
		}
	},
	PerformanceMarkOptions: class {
		constructor() {
			this.Detail = null;
			this.StartTime = 0;
			this.Explicit = null;
		}
	},
	PeriodicWaveConstraints: class {
		constructor() {
			this.DisableNormalization = false;
			this.Explicit = null;
		}
	},
	PointerLockOptions: class {
		constructor() {
			this.UnadjustedMovement = false;
			this.Explicit = null;
		}
	},
	RequestInit: class {
		constructor() {
			this.Body = null;
			this.Cache = "";
			this.Credentials = "";
			this.Headers = null;
			this.Integrity = "";
			this.Keepalive = false;
			this.Method = "";
			this.Mode = "";
			this.Priority = "";
			this.Redirect = "";
			this.Referrer = "";
			this.ReferrerPolicy = "";
			this.Signal = null;
			this.Explicit = null;
		}
	},
	ScrollToOptions: class {
		constructor() {
			this.Left = 0;
			this.Top = 0;
			this.Behavior = "";
			this.Explicit = null;
		}
	},
	ShadowRootInit: class {
		constructor() {
			this.DelegatesFocus = false;
			this.Mode = "";
			this.Serializable = false;
			this.SlotAssignment = "";
			this.Explicit = null;
		}
	},
	ShareData: class {
		constructor() {
			this.Files = [];
			this.Text = "";
			this.Title = "";
			this.Url = "";
			this.Explicit = null;
		}
	},
	StructuredSerializeOptions: class {
		constructor() {
			this.Transfer = [];
			this.Explicit = null;
		}
	},
	ValidityStateFlags: class {
		constructor() {
			this.BadInput = false;
			this.CustomError = false;
			this.PatternMismatch = false;
			this.RangeOverflow = false;
			this.RangeUnderflow = false;
			this.StepMismatch = false;
			this.TooLong = false;
			this.TooShort = false;
			this.TypeMismatch = false;
			this.ValueMissing = false;
			this.Explicit = null;
		}
	},
	WorkletOptions: class {
		constructor() {
			this.Credentials = "";
			this.Explicit = null;
		}
	},
	AnimationTimeline: class {
		constructor() {
			this.CurrentTime = null;
			this.Explicit = null;
		}
	},
	BarProp: class {
		constructor() {
			this.Visible = false;
			this.Explicit = null;
		}
	},
	DOMException: class {
		constructor() {
			this.Code = 0;
			this.Message = "";
			this.Name = "";
			this.INDEX_SIZE_ERR = 0;
			this.DOMSTRING_SIZE_ERR = 0;
			this.HIERARCHY_REQUEST_ERR = 0;
			this.WRONG_DOCUMENT_ERR = 0;
			this.INVALID_CHARACTER_ERR = 0;
			this.NO_DATA_ALLOWED_ERR = 0;
			this.NO_MODIFICATION_ALLOWED_ERR = 0;
			this.NOT_FOUND_ERR = 0;
			this.NOT_SUPPORTED_ERR = 0;
			this.INUSE_ATTRIBUTE_ERR = 0;
			this.INVALID_STATE_ERR = 0;
			this.SYNTAX_ERR = 0;
			this.INVALID_MODIFICATION_ERR = 0;
			this.NAMESPACE_ERR = 0;
			this.INVALID_ACCESS_ERR = 0;
			this.VALIDATION_ERR = 0;
			this.TYPE_MISMATCH_ERR = 0;
			this.SECURITY_ERR = 0;
			this.NETWORK_ERR = 0;
			this.ABORT_ERR = 0;
			this.URL_MISMATCH_ERR = 0;
			this.QUOTA_EXCEEDED_ERR = 0;
			this.TIMEOUT_ERR = 0;
			this.INVALID_NODE_TYPE_ERR = 0;
			this.DATA_CLONE_ERR = 0;
			this.Explicit = null;
		}
	},
	DocumentTimeline: class {
		constructor() {
			this.CurrentTime = null;
			this.Explicit = null;
		}
	},
	Gamepad: class {
		constructor() {
			this.Axes = [];
			this.Buttons = [];
			this.Connected = false;
			this.Id = "";
			this.Index = 0;
			this.Mapping = "";
			this.Timestamp = 0;
			this.VibrationActuator = null;
			this.Explicit = null;
		}
	},
	ImageData: class {
		constructor() {
			this.ColorSpace = "";
			this.Data = [];
			this.Height = 0;
			this.Width = 0;
			this.Explicit = null;
		}
	},
	MediaError: class {
		constructor() {
			this.Code = 0;
			this.Message = "";
			this.MEDIA_ERR_ABORTED = 0;
			this.MEDIA_ERR_NETWORK = 0;
			this.MEDIA_ERR_DECODE = 0;
			this.MEDIA_ERR_SRC_NOT_SUPPORTED = 0;
			this.Explicit = null;
		}
	},
	Screen: class {
		constructor() {
			this.AvailHeight = 0;
			this.AvailWidth = 0;
			this.ColorDepth = 0;
			this.Height = 0;
			this.Orientation = null;
			this.PixelDepth = 0;
			this.Width = 0;
			this.Explicit = null;
		}
	},
	StaticRange: class {
		constructor() {
			this.Collapsed = false;
			this.EndContainer = null;
			this.EndOffset = 0;
			this.StartContainer = null;
			this.StartOffset = 0;
			this.Explicit = null;
		}
	},
	TextMetrics: class {
		constructor() {
			this.ActualBoundingBoxAscent = 0;
			this.ActualBoundingBoxDescent = 0;
			this.ActualBoundingBoxLeft = 0;
			this.ActualBoundingBoxRight = 0;
			this.AlphabeticBaseline = 0;
			this.EmHeightAscent = 0;
			this.EmHeightDescent = 0;
			this.FontBoundingBoxAscent = 0;
			this.FontBoundingBoxDescent = 0;
			this.HangingBaseline = 0;
			this.IdeographicBaseline = 0;
			this.Width = 0;
			this.Explicit = null;
		}
	},
	Touch: class {
		constructor() {
			this.ClientX = 0;
			this.ClientY = 0;
			this.Force = 0;
			this.Identifier = 0;
			this.PageX = 0;
			this.PageY = 0;
			this.RadiusX = 0;
			this.RadiusY = 0;
			this.RotationAngle = 0;
			this.ScreenX = 0;
			this.ScreenY = 0;
			this.Target = null;
			this.Explicit = null;
		}
	},
	UserActivation: class {
		constructor() {
			this.HasBeenActive = false;
			this.IsActive = false;
			this.Explicit = null;
		}
	},
	ValidityState: class {
		constructor() {
			this.BadInput = false;
			this.CustomError = false;
			this.PatternMismatch = false;
			this.RangeOverflow = false;
			this.RangeUnderflow = false;
			this.StepMismatch = false;
			this.TooLong = false;
			this.TooShort = false;
			this.TypeMismatch = false;
			this.Valid = false;
			this.ValueMissing = false;
			this.Explicit = null;
		}
	},
	get AudioContextValue() {
		return dom_wrap(dom.AudioContextType, dom_load().AudioContext);
	},
	get CanvasRenderingContext2DValue() {
		return dom_wrap(dom.CanvasRenderingContext2DType, dom_load().CanvasRenderingContext2D);
	},
	get HTMLAudioElementValue() {
		return dom_wrap(dom.HTMLAudioElementType, dom_load().HTMLAudioElement);
	},
	get HTMLCanvasElementValue() {
		return dom_wrap(dom.HTMLCanvasElementType, dom_load().HTMLCanvasElement);
	},
	get HTMLImageElementValue() {
		return dom_wrap(dom.HTMLImageElementType, dom_load().HTMLImageElement);
	},
	get ImageDataValue() {
		return dom_wrap(dom.ImageDataType, dom_load().ImageData);
	},
	get KeyboardEventValue() {
		return dom_wrap(dom.KeyboardEventType, dom_load().KeyboardEvent);
	},
	get MouseEventValue() {
		return dom_wrap(dom.MouseEventType, dom_load().MouseEvent);
	},
	get Path2DValue() {
		return dom_wrap(dom.Path2DType, dom_load().Path2D);
	},
	get PointerEventValue() {
		return dom_wrap(dom.PointerEventType, dom_load().PointerEvent);
	},
	get TouchEventValue() {
		return dom_wrap(dom.TouchEventType, dom_load().TouchEvent);
	},
	get WheelEventValue() {
		return dom_wrap(dom.WheelEventType, dom_load().WheelEvent);
	},
	get DocumentValue() {
		return dom_wrap(dom.Document, dom_load().document);
	},
	get LocationValue() {
		return dom_wrap(dom.Location, dom_load().location);
	},
	get NavigatorValue() {
		return dom_wrap(dom.Navigator, dom_load().navigator);
	},
	get WindowValue() {
		return dom_wrap(dom.Window, dom_load().window);
	},
	get LocalStorage() {
		return dom_wrap(dom.Storage, dom_load().localStorage);
	},
	get PerformanceValue() {
		return dom_wrap(dom.Performance, dom_load().performance);
	},
	CancelAnimationFrame(p0) {
		dom_load().cancelAnimationFrame(p0);
	},
	RequestAnimationFrame(p0) {
		return dom_load().requestAnimationFrame((p0 === null || p0 === undefined ? null : (...a) => go2jsCallNow(p0, null, [a[0]])));
	},
	*Fetch(p0, ...p1) {
		if (p1.length === 1 && Array.isArray(p1[0]) && p1[0].length === 0) {
			p1 = [];
		}
		
		const outcome = yield* dom_await(() => dom_load().fetch(dom_unwrapAny(p0), ...p1.map((item) => dom_toRequestInit(item))));
		
		return outcome.ok ? [dom_wrap(dom.Response, outcome.value), null] : [null, outcome.error];
	},
	ARIAMixin: dom_C_ARIAMixin,
	EventTarget: dom_C_EventTarget,
	AbortSignal: dom_C_AbortSignal,
	AbstractRange: dom_C_AbstractRange,
	AudioNode: dom_C_AudioNode,
	AnalyserNode: dom_C_AnalyserNode,
	Animatable: dom_C_Animatable,
	Animation: dom_C_Animation,
	AnimationEffect: dom_C_AnimationEffect,
	Event: dom_C_Event,
	AnimationEvent: dom_C_AnimationEvent,
	AnimationFrameProvider: dom_C_AnimationFrameProvider,
	AnimationPlaybackEvent: dom_C_AnimationPlaybackEvent,
	Node: dom_C_Node,
	Attr: dom_C_Attr,
	AudioBuffer: dom_C_AudioBuffer,
	AudioScheduledSourceNode: dom_C_AudioScheduledSourceNode,
	AudioBufferSourceNode: dom_C_AudioBufferSourceNode,
	BaseAudioContext: dom_C_BaseAudioContext,
	AudioContext: dom_C_AudioContext,
	AudioContextType: dom_C_AudioContextType,
	AudioDestinationNode: dom_C_AudioDestinationNode,
	AudioListener: dom_C_AudioListener,
	AudioParam: dom_C_AudioParam,
	AudioProcessingEvent: dom_C_AudioProcessingEvent,
	Worklet: dom_C_Worklet,
	AudioWorklet: dom_C_AudioWorklet,
	BeforeUnloadEvent: dom_C_BeforeUnloadEvent,
	BiquadFilterNode: dom_C_BiquadFilterNode,
	Blob: dom_C_Blob,
	Body: dom_C_Body,
	ChildNode: dom_C_ChildNode,
	NonDocumentTypeChildNode: dom_C_NonDocumentTypeChildNode,
	CharacterData: dom_C_CharacterData,
	Slottable: dom_C_Slottable,
	Text: dom_C_Text,
	CDATASection: dom_C_CDATASection,
	CSSRule: dom_C_CSSRule,
	CSSStyleDeclaration: dom_C_CSSStyleDeclaration,
	StyleSheet: dom_C_StyleSheet,
	CSSStyleSheet: dom_C_CSSStyleSheet,
	CSSStyleValue: dom_C_CSSStyleValue,
	CacheStorage: dom_C_CacheStorage,
	CanvasCompositing: dom_C_CanvasCompositing,
	CanvasDrawImage: dom_C_CanvasDrawImage,
	CanvasDrawPath: dom_C_CanvasDrawPath,
	CanvasFillStrokeStyles: dom_C_CanvasFillStrokeStyles,
	CanvasFilters: dom_C_CanvasFilters,
	CanvasGradient: dom_C_CanvasGradient,
	CanvasImageData: dom_C_CanvasImageData,
	CanvasImageSmoothing: dom_C_CanvasImageSmoothing,
	CanvasPath: dom_C_CanvasPath,
	CanvasPathDrawingStyles: dom_C_CanvasPathDrawingStyles,
	CanvasPattern: dom_C_CanvasPattern,
	CanvasRect: dom_C_CanvasRect,
	CanvasShadowStyles: dom_C_CanvasShadowStyles,
	CanvasState: dom_C_CanvasState,
	CanvasText: dom_C_CanvasText,
	CanvasTextDrawingStyles: dom_C_CanvasTextDrawingStyles,
	CanvasTransform: dom_C_CanvasTransform,
	CanvasUserInterface: dom_C_CanvasUserInterface,
	CanvasRenderingContext2D: dom_C_CanvasRenderingContext2D,
	CanvasRenderingContext2DType: dom_C_CanvasRenderingContext2DType,
	ChannelMergerNode: dom_C_ChannelMergerNode,
	ChannelSplitterNode: dom_C_ChannelSplitterNode,
	Clipboard: dom_C_Clipboard,
	ClipboardEvent: dom_C_ClipboardEvent,
	Comment: dom_C_Comment,
	UIEvent: dom_C_UIEvent,
	CompositionEvent: dom_C_CompositionEvent,
	ConstantSourceNode: dom_C_ConstantSourceNode,
	ConvolverNode: dom_C_ConvolverNode,
	CredentialsContainer: dom_C_CredentialsContainer,
	Crypto: dom_C_Crypto,
	CustomElementRegistry: dom_C_CustomElementRegistry,
	CustomStateSet: dom_C_CustomStateSet,
	DOMImplementation: dom_C_DOMImplementation,
	DOMMatrixReadOnly: dom_C_DOMMatrixReadOnly,
	DOMMatrix: dom_C_DOMMatrix,
	DOMPointReadOnly: dom_C_DOMPointReadOnly,
	DOMPoint: dom_C_DOMPoint,
	DOMRectReadOnly: dom_C_DOMRectReadOnly,
	DOMRect: dom_C_DOMRect,
	DOMRectList: dom_C_DOMRectList,
	DOMStringList: dom_C_DOMStringList,
	DOMTokenList: dom_C_DOMTokenList,
	DataTransfer: dom_C_DataTransfer,
	DelayNode: dom_C_DelayNode,
	DeviceMotionEvent: dom_C_DeviceMotionEvent,
	DeviceOrientationEvent: dom_C_DeviceOrientationEvent,
	DocumentOrShadowRoot: dom_C_DocumentOrShadowRoot,
	FontFaceSource: dom_C_FontFaceSource,
	GlobalEventHandlers: dom_C_GlobalEventHandlers,
	NonElementParentNode: dom_C_NonElementParentNode,
	ParentNode: dom_C_ParentNode,
	XPathEvaluatorBase: dom_C_XPathEvaluatorBase,
	Document: dom_C_Document,
	DocumentType: dom_C_DocumentType,
	DocumentFragment: dom_C_DocumentFragment,
	MouseEvent: dom_C_MouseEvent,
	DragEvent: dom_C_DragEvent,
	DynamicsCompressorNode: dom_C_DynamicsCompressorNode,
	Element: dom_C_Element,
	ElementCSSInlineStyle: dom_C_ElementCSSInlineStyle,
	ElementContentEditable: dom_C_ElementContentEditable,
	ElementInternals: dom_C_ElementInternals,
	ErrorEvent: dom_C_ErrorEvent,
	EventCounts: dom_C_EventCounts,
	External: dom_C_External,
	FocusEvent: dom_C_FocusEvent,
	FontFaceSet: dom_C_FontFaceSet,
	FormData: dom_C_FormData,
	FormDataEvent: dom_C_FormDataEvent,
	GainNode: dom_C_GainNode,
	GamepadEvent: dom_C_GamepadEvent,
	Geolocation: dom_C_Geolocation,
	HTMLAllCollection: dom_C_HTMLAllCollection,
	HTMLOrSVGElement: dom_C_HTMLOrSVGElement,
	HTMLElement: dom_C_HTMLElement,
	HTMLMediaElement: dom_C_HTMLMediaElement,
	HTMLAudioElement: dom_C_HTMLAudioElement,
	HTMLAudioElementType: dom_C_HTMLAudioElementType,
	HTMLCanvasElement: dom_C_HTMLCanvasElement,
	HTMLCanvasElementType: dom_C_HTMLCanvasElementType,
	HTMLCollectionBase: dom_C_HTMLCollectionBase,
	HTMLCollection: dom_C_HTMLCollection,
	HTMLCollectionOf: dom_C_HTMLCollectionOf,
	HTMLFormElement: dom_C_HTMLFormElement,
	HTMLHeadElement: dom_C_HTMLHeadElement,
	HTMLImageElement: dom_C_HTMLImageElement,
	HTMLImageElementType: dom_C_HTMLImageElementType,
	HTMLSlotElement: dom_C_HTMLSlotElement,
	HashChangeEvent: dom_C_HashChangeEvent,
	Headers: dom_C_Headers,
	History: dom_C_History,
	IDBFactory: dom_C_IDBFactory,
	IIRFilterNode: dom_C_IIRFilterNode,
	IdleDeadline: dom_C_IdleDeadline,
	ImageBitmap: dom_C_ImageBitmap,
	ImageDataType: dom_C_ImageDataType,
	InputEvent: dom_C_InputEvent,
	KeyboardEvent: dom_C_KeyboardEvent,
	KeyboardEventType: dom_C_KeyboardEventType,
	LinkStyle: dom_C_LinkStyle,
	Location: dom_C_Location,
	LockManager: dom_C_LockManager,
	MIDIAccess: dom_C_MIDIAccess,
	MediaCapabilities: dom_C_MediaCapabilities,
	MediaDevices: dom_C_MediaDevices,
	MediaElementAudioSourceNode: dom_C_MediaElementAudioSourceNode,
	MediaEncryptedEvent: dom_C_MediaEncryptedEvent,
	MediaKeySession: dom_C_MediaKeySession,
	MediaKeySessionType: dom_C_MediaKeySessionType,
	MediaKeySystemAccess: dom_C_MediaKeySystemAccess,
	MediaKeys: dom_C_MediaKeys,
	MediaQueryList: dom_C_MediaQueryList,
	MediaSession: dom_C_MediaSession,
	MediaStream: dom_C_MediaStream,
	MediaStreamAudioDestinationNode: dom_C_MediaStreamAudioDestinationNode,
	MediaStreamAudioSourceNode: dom_C_MediaStreamAudioSourceNode,
	MediaStreamTrack: dom_C_MediaStreamTrack,
	MediaStreamTrackEvent: dom_C_MediaStreamTrackEvent,
	MessageEvent: dom_C_MessageEvent,
	MimeTypeArray: dom_C_MimeTypeArray,
	MouseEventType: dom_C_MouseEventType,
	NamedNodeMap: dom_C_NamedNodeMap,
	NavigatorAutomationInformation: dom_C_NavigatorAutomationInformation,
	NavigatorBadge: dom_C_NavigatorBadge,
	NavigatorConcurrentHardware: dom_C_NavigatorConcurrentHardware,
	NavigatorContentUtils: dom_C_NavigatorContentUtils,
	NavigatorCookies: dom_C_NavigatorCookies,
	NavigatorID: dom_C_NavigatorID,
	NavigatorLanguage: dom_C_NavigatorLanguage,
	NavigatorLocks: dom_C_NavigatorLocks,
	NavigatorOnLine: dom_C_NavigatorOnLine,
	NavigatorPlugins: dom_C_NavigatorPlugins,
	NavigatorStorage: dom_C_NavigatorStorage,
	Navigator: dom_C_Navigator,
	NodeIterator: dom_C_NodeIterator,
	NodeList: dom_C_NodeList,
	NodeListOf: dom_C_NodeListOf,
	OffscreenCanvas: dom_C_OffscreenCanvas,
	OffscreenCanvasRenderingContext2D: dom_C_OffscreenCanvasRenderingContext2D,
	OscillatorNode: dom_C_OscillatorNode,
	PageTransitionEvent: dom_C_PageTransitionEvent,
	PannerNode: dom_C_PannerNode,
	Path2D: dom_C_Path2D,
	Path2DType: dom_C_Path2DType,
	Performance: dom_C_Performance,
	PerformanceEntry: dom_C_PerformanceEntry,
	PerformanceMark: dom_C_PerformanceMark,
	PerformanceMeasure: dom_C_PerformanceMeasure,
	PerformanceNavigation: dom_C_PerformanceNavigation,
	PerformanceTiming: dom_C_PerformanceTiming,
	Permissions: dom_C_Permissions,
	PluginArray: dom_C_PluginArray,
	PointerEvent: dom_C_PointerEvent,
	PointerEventType: dom_C_PointerEventType,
	PopStateEvent: dom_C_PopStateEvent,
	ProcessingInstruction: dom_C_ProcessingInstruction,
	ProgressEvent: dom_C_ProgressEvent,
	PromiseRejectionEvent: dom_C_PromiseRejectionEvent,
	Range: dom_C_Range,
	ReadableStream: dom_C_ReadableStream,
	RemotePlayback: dom_C_RemotePlayback,
	Response: dom_C_Response,
	ResponseType: dom_C_ResponseType,
	SVGElement: dom_C_SVGElement,
	SVGFitToViewBox: dom_C_SVGFitToViewBox,
	SVGTests: dom_C_SVGTests,
	SVGGraphicsElement: dom_C_SVGGraphicsElement,
	WindowEventHandlers: dom_C_WindowEventHandlers,
	SVGSVGElement: dom_C_SVGSVGElement,
	ScriptProcessorNode: dom_C_ScriptProcessorNode,
	SecurityPolicyViolationEvent: dom_C_SecurityPolicyViolationEvent,
	Selection: dom_C_Selection,
	ServiceWorkerContainer: dom_C_ServiceWorkerContainer,
	ShadowRoot: dom_C_ShadowRoot,
	SpeechSynthesis: dom_C_SpeechSynthesis,
	StereoPannerNode: dom_C_StereoPannerNode,
	Storage: dom_C_Storage,
	StorageEvent: dom_C_StorageEvent,
	StorageManager: dom_C_StorageManager,
	StylePropertyMapReadOnly: dom_C_StylePropertyMapReadOnly,
	StylePropertyMap: dom_C_StylePropertyMap,
	StyleSheetList: dom_C_StyleSheetList,
	SubmitEvent: dom_C_SubmitEvent,
	TextTrack: dom_C_TextTrack,
	TextTrackCue: dom_C_TextTrackCue,
	TextTrackCueList: dom_C_TextTrackCueList,
	TextTrackList: dom_C_TextTrackList,
	TimeRanges: dom_C_TimeRanges,
	TouchEvent: dom_C_TouchEvent,
	TouchEventType: dom_C_TouchEventType,
	TouchList: dom_C_TouchList,
	TrackEvent: dom_C_TrackEvent,
	TransitionEvent: dom_C_TransitionEvent,
	TreeWalker: dom_C_TreeWalker,
	ViewTransition: dom_C_ViewTransition,
	VisualViewport: dom_C_VisualViewport,
	WakeLock: dom_C_WakeLock,
	WaveShaperNode: dom_C_WaveShaperNode,
	WheelEvent: dom_C_WheelEvent,
	WheelEventType: dom_C_WheelEventType,
	WindowLocalStorage: dom_C_WindowLocalStorage,
	WindowOrWorkerGlobalScope: dom_C_WindowOrWorkerGlobalScope,
	WindowSessionStorage: dom_C_WindowSessionStorage,
	Window: dom_C_Window,
	XPathExpression: dom_C_XPathExpression,
	XPathResult: dom_C_XPathResult,
};

// the classes of the library by the name its objects carry, with the classes each is a kind of
const dom_kinds = {
	"ARIAMixin": {Class: () => dom.ARIAMixin, lineage: ["ARIAMixin"]},
	"AbortSignal": {Class: () => dom.AbortSignal, lineage: ["AbortSignal", "EventTarget"]},
	"AbstractRange": {Class: () => dom.AbstractRange, lineage: ["AbstractRange"]},
	"AnalyserNode": {Class: () => dom.AnalyserNode, lineage: ["AnalyserNode", "AudioNode", "EventTarget"]},
	"Animatable": {Class: () => dom.Animatable, lineage: ["Animatable"]},
	"Animation": {Class: () => dom.Animation, lineage: ["Animation", "EventTarget"]},
	"AnimationEffect": {Class: () => dom.AnimationEffect, lineage: ["AnimationEffect"]},
	"AnimationEvent": {Class: () => dom.AnimationEvent, lineage: ["AnimationEvent", "Event"]},
	"AnimationFrameProvider": {Class: () => dom.AnimationFrameProvider, lineage: ["AnimationFrameProvider"]},
	"AnimationPlaybackEvent": {Class: () => dom.AnimationPlaybackEvent, lineage: ["AnimationPlaybackEvent", "Event"]},
	"Attr": {Class: () => dom.Attr, lineage: ["Attr", "Node", "EventTarget"]},
	"AudioBuffer": {Class: () => dom.AudioBuffer, lineage: ["AudioBuffer"]},
	"AudioBufferSourceNode": {Class: () => dom.AudioBufferSourceNode, lineage: ["AudioBufferSourceNode", "AudioScheduledSourceNode", "AudioNode", "EventTarget"]},
	"AudioContext": {Class: () => dom.AudioContext, lineage: ["AudioContext", "BaseAudioContext", "EventTarget"]},
	"AudioContextType": {Class: () => dom.AudioContextType, lineage: ["AudioContextType"]},
	"AudioDestinationNode": {Class: () => dom.AudioDestinationNode, lineage: ["AudioDestinationNode", "AudioNode", "EventTarget"]},
	"AudioListener": {Class: () => dom.AudioListener, lineage: ["AudioListener"]},
	"AudioNode": {Class: () => dom.AudioNode, lineage: ["AudioNode", "EventTarget"]},
	"AudioParam": {Class: () => dom.AudioParam, lineage: ["AudioParam"]},
	"AudioProcessingEvent": {Class: () => dom.AudioProcessingEvent, lineage: ["AudioProcessingEvent", "Event"]},
	"AudioScheduledSourceNode": {Class: () => dom.AudioScheduledSourceNode, lineage: ["AudioScheduledSourceNode", "AudioNode", "EventTarget"]},
	"AudioWorklet": {Class: () => dom.AudioWorklet, lineage: ["AudioWorklet", "Worklet"]},
	"BaseAudioContext": {Class: () => dom.BaseAudioContext, lineage: ["BaseAudioContext", "EventTarget"]},
	"BeforeUnloadEvent": {Class: () => dom.BeforeUnloadEvent, lineage: ["BeforeUnloadEvent", "Event"]},
	"BiquadFilterNode": {Class: () => dom.BiquadFilterNode, lineage: ["BiquadFilterNode", "AudioNode", "EventTarget"]},
	"Blob": {Class: () => dom.Blob, lineage: ["Blob"]},
	"Body": {Class: () => dom.Body, lineage: ["Body"]},
	"CDATASection": {Class: () => dom.CDATASection, lineage: ["CDATASection", "Text", "CharacterData", "Slottable", "Node", "ChildNode", "NonDocumentTypeChildNode", "EventTarget"]},
	"CSSRule": {Class: () => dom.CSSRule, lineage: ["CSSRule"]},
	"CSSStyleDeclaration": {Class: () => dom.CSSStyleDeclaration, lineage: ["CSSStyleDeclaration"]},
	"CSSStyleSheet": {Class: () => dom.CSSStyleSheet, lineage: ["CSSStyleSheet", "StyleSheet"]},
	"CSSStyleValue": {Class: () => dom.CSSStyleValue, lineage: ["CSSStyleValue"]},
	"CacheStorage": {Class: () => dom.CacheStorage, lineage: ["CacheStorage"]},
	"CanvasCompositing": {Class: () => dom.CanvasCompositing, lineage: ["CanvasCompositing"]},
	"CanvasDrawImage": {Class: () => dom.CanvasDrawImage, lineage: ["CanvasDrawImage"]},
	"CanvasDrawPath": {Class: () => dom.CanvasDrawPath, lineage: ["CanvasDrawPath"]},
	"CanvasFillStrokeStyles": {Class: () => dom.CanvasFillStrokeStyles, lineage: ["CanvasFillStrokeStyles"]},
	"CanvasFilters": {Class: () => dom.CanvasFilters, lineage: ["CanvasFilters"]},
	"CanvasGradient": {Class: () => dom.CanvasGradient, lineage: ["CanvasGradient"]},
	"CanvasImageData": {Class: () => dom.CanvasImageData, lineage: ["CanvasImageData"]},
	"CanvasImageSmoothing": {Class: () => dom.CanvasImageSmoothing, lineage: ["CanvasImageSmoothing"]},
	"CanvasPath": {Class: () => dom.CanvasPath, lineage: ["CanvasPath"]},
	"CanvasPathDrawingStyles": {Class: () => dom.CanvasPathDrawingStyles, lineage: ["CanvasPathDrawingStyles"]},
	"CanvasPattern": {Class: () => dom.CanvasPattern, lineage: ["CanvasPattern"]},
	"CanvasRect": {Class: () => dom.CanvasRect, lineage: ["CanvasRect"]},
	"CanvasRenderingContext2D": {Class: () => dom.CanvasRenderingContext2D, lineage: ["CanvasRenderingContext2D", "CanvasCompositing", "CanvasDrawImage", "CanvasDrawPath", "CanvasFillStrokeStyles", "CanvasFilters", "CanvasImageData", "CanvasImageSmoothing", "CanvasPath", "CanvasPathDrawingStyles", "CanvasRect", "CanvasShadowStyles", "CanvasState", "CanvasText", "CanvasTextDrawingStyles", "CanvasTransform", "CanvasUserInterface"]},
	"CanvasRenderingContext2DType": {Class: () => dom.CanvasRenderingContext2DType, lineage: ["CanvasRenderingContext2DType"]},
	"CanvasShadowStyles": {Class: () => dom.CanvasShadowStyles, lineage: ["CanvasShadowStyles"]},
	"CanvasState": {Class: () => dom.CanvasState, lineage: ["CanvasState"]},
	"CanvasText": {Class: () => dom.CanvasText, lineage: ["CanvasText"]},
	"CanvasTextDrawingStyles": {Class: () => dom.CanvasTextDrawingStyles, lineage: ["CanvasTextDrawingStyles"]},
	"CanvasTransform": {Class: () => dom.CanvasTransform, lineage: ["CanvasTransform"]},
	"CanvasUserInterface": {Class: () => dom.CanvasUserInterface, lineage: ["CanvasUserInterface"]},
	"ChannelMergerNode": {Class: () => dom.ChannelMergerNode, lineage: ["ChannelMergerNode", "AudioNode", "EventTarget"]},
	"ChannelSplitterNode": {Class: () => dom.ChannelSplitterNode, lineage: ["ChannelSplitterNode", "AudioNode", "EventTarget"]},
	"CharacterData": {Class: () => dom.CharacterData, lineage: ["CharacterData", "Node", "ChildNode", "NonDocumentTypeChildNode", "EventTarget"]},
	"ChildNode": {Class: () => dom.ChildNode, lineage: ["ChildNode", "Node", "EventTarget"]},
	"Clipboard": {Class: () => dom.Clipboard, lineage: ["Clipboard", "EventTarget"]},
	"ClipboardEvent": {Class: () => dom.ClipboardEvent, lineage: ["ClipboardEvent", "Event"]},
	"Comment": {Class: () => dom.Comment, lineage: ["Comment", "CharacterData", "Node", "ChildNode", "NonDocumentTypeChildNode", "EventTarget"]},
	"CompositionEvent": {Class: () => dom.CompositionEvent, lineage: ["CompositionEvent", "UIEvent", "Event"]},
	"ConstantSourceNode": {Class: () => dom.ConstantSourceNode, lineage: ["ConstantSourceNode", "AudioScheduledSourceNode", "AudioNode", "EventTarget"]},
	"ConvolverNode": {Class: () => dom.ConvolverNode, lineage: ["ConvolverNode", "AudioNode", "EventTarget"]},
	"CredentialsContainer": {Class: () => dom.CredentialsContainer, lineage: ["CredentialsContainer"]},
	"Crypto": {Class: () => dom.Crypto, lineage: ["Crypto"]},
	"CustomElementRegistry": {Class: () => dom.CustomElementRegistry, lineage: ["CustomElementRegistry"]},
	"CustomStateSet": {Class: () => dom.CustomStateSet, lineage: ["CustomStateSet"]},
	"DOMImplementation": {Class: () => dom.DOMImplementation, lineage: ["DOMImplementation"]},
	"DOMMatrix": {Class: () => dom.DOMMatrix, lineage: ["DOMMatrix", "DOMMatrixReadOnly"]},
	"DOMMatrixReadOnly": {Class: () => dom.DOMMatrixReadOnly, lineage: ["DOMMatrixReadOnly"]},
	"DOMPoint": {Class: () => dom.DOMPoint, lineage: ["DOMPoint", "DOMPointReadOnly"]},
	"DOMPointReadOnly": {Class: () => dom.DOMPointReadOnly, lineage: ["DOMPointReadOnly"]},
	"DOMRect": {Class: () => dom.DOMRect, lineage: ["DOMRect", "DOMRectReadOnly"]},
	"DOMRectList": {Class: () => dom.DOMRectList, lineage: ["DOMRectList"]},
	"DOMRectReadOnly": {Class: () => dom.DOMRectReadOnly, lineage: ["DOMRectReadOnly"]},
	"DOMStringList": {Class: () => dom.DOMStringList, lineage: ["DOMStringList"]},
	"DOMTokenList": {Class: () => dom.DOMTokenList, lineage: ["DOMTokenList"]},
	"DataTransfer": {Class: () => dom.DataTransfer, lineage: ["DataTransfer"]},
	"DelayNode": {Class: () => dom.DelayNode, lineage: ["DelayNode", "AudioNode", "EventTarget"]},
	"DeviceMotionEvent": {Class: () => dom.DeviceMotionEvent, lineage: ["DeviceMotionEvent", "Event"]},
	"DeviceOrientationEvent": {Class: () => dom.DeviceOrientationEvent, lineage: ["DeviceOrientationEvent", "Event"]},
	"Document": {Class: () => dom.Document, lineage: ["Document", "Node", "DocumentOrShadowRoot", "FontFaceSource", "GlobalEventHandlers", "NonElementParentNode", "ParentNode", "XPathEvaluatorBase", "EventTarget"]},
	"DocumentType": {Class: () => dom.DocumentType, lineage: ["DocumentType", "Node", "ChildNode", "EventTarget"]},
	"DocumentFragment": {Class: () => dom.DocumentFragment, lineage: ["DocumentFragment", "Node", "NonElementParentNode", "ParentNode", "EventTarget"]},
	"DocumentOrShadowRoot": {Class: () => dom.DocumentOrShadowRoot, lineage: ["DocumentOrShadowRoot"]},
	"DragEvent": {Class: () => dom.DragEvent, lineage: ["DragEvent", "MouseEvent", "UIEvent", "Event"]},
	"DynamicsCompressorNode": {Class: () => dom.DynamicsCompressorNode, lineage: ["DynamicsCompressorNode", "AudioNode", "EventTarget"]},
	"Element": {Class: () => dom.Element, lineage: ["Element", "Node", "ARIAMixin", "Animatable", "ChildNode", "NonDocumentTypeChildNode", "ParentNode", "Slottable", "EventTarget"]},
	"ElementCSSInlineStyle": {Class: () => dom.ElementCSSInlineStyle, lineage: ["ElementCSSInlineStyle"]},
	"ElementContentEditable": {Class: () => dom.ElementContentEditable, lineage: ["ElementContentEditable"]},
	"ElementInternals": {Class: () => dom.ElementInternals, lineage: ["ElementInternals", "ARIAMixin"]},
	"ErrorEvent": {Class: () => dom.ErrorEvent, lineage: ["ErrorEvent", "Event"]},
	"Event": {Class: () => dom.Event, lineage: ["Event"]},
	"EventCounts": {Class: () => dom.EventCounts, lineage: ["EventCounts"]},
	"EventTarget": {Class: () => dom.EventTarget, lineage: ["EventTarget"]},
	"External": {Class: () => dom.External, lineage: ["External"]},
	"FocusEvent": {Class: () => dom.FocusEvent, lineage: ["FocusEvent", "UIEvent", "Event"]},
	"FontFaceSet": {Class: () => dom.FontFaceSet, lineage: ["FontFaceSet", "EventTarget"]},
	"FontFaceSource": {Class: () => dom.FontFaceSource, lineage: ["FontFaceSource"]},
	"FormData": {Class: () => dom.FormData, lineage: ["FormData"]},
	"FormDataEvent": {Class: () => dom.FormDataEvent, lineage: ["FormDataEvent", "Event"]},
	"GainNode": {Class: () => dom.GainNode, lineage: ["GainNode", "AudioNode", "EventTarget"]},
	"GamepadEvent": {Class: () => dom.GamepadEvent, lineage: ["GamepadEvent", "Event"]},
	"Geolocation": {Class: () => dom.Geolocation, lineage: ["Geolocation"]},
	"GlobalEventHandlers": {Class: () => dom.GlobalEventHandlers, lineage: ["GlobalEventHandlers"]},
	"HTMLAllCollection": {Class: () => dom.HTMLAllCollection, lineage: ["HTMLAllCollection"]},
	"HTMLAudioElement": {Class: () => dom.HTMLAudioElement, lineage: ["HTMLAudioElement", "HTMLMediaElement", "HTMLElement", "Element", "ElementCSSInlineStyle", "ElementContentEditable", "GlobalEventHandlers", "HTMLOrSVGElement", "Node", "ARIAMixin", "Animatable", "ChildNode", "NonDocumentTypeChildNode", "ParentNode", "Slottable", "EventTarget"]},
	"HTMLAudioElementType": {Class: () => dom.HTMLAudioElementType, lineage: ["HTMLAudioElementType"]},
	"HTMLCanvasElement": {Class: () => dom.HTMLCanvasElement, lineage: ["HTMLCanvasElement", "HTMLElement", "Element", "ElementCSSInlineStyle", "ElementContentEditable", "GlobalEventHandlers", "HTMLOrSVGElement", "Node", "ARIAMixin", "Animatable", "ChildNode", "NonDocumentTypeChildNode", "ParentNode", "Slottable", "EventTarget"]},
	"HTMLCanvasElementType": {Class: () => dom.HTMLCanvasElementType, lineage: ["HTMLCanvasElementType"]},
	"HTMLCollectionBase": {Class: () => dom.HTMLCollectionBase, lineage: ["HTMLCollectionBase"]},
	"HTMLCollection": {Class: () => dom.HTMLCollection, lineage: ["HTMLCollection", "HTMLCollectionBase"]},
	"HTMLCollectionOf": {Class: () => dom.HTMLCollectionOf, lineage: ["HTMLCollectionOf", "HTMLCollectionBase"]},
	"HTMLElement": {Class: () => dom.HTMLElement, lineage: ["HTMLElement", "Element", "ElementCSSInlineStyle", "ElementContentEditable", "GlobalEventHandlers", "HTMLOrSVGElement", "Node", "ARIAMixin", "Animatable", "ChildNode", "NonDocumentTypeChildNode", "ParentNode", "Slottable", "EventTarget"]},
	"HTMLFormElement": {Class: () => dom.HTMLFormElement, lineage: ["HTMLFormElement", "HTMLElement", "Element", "ElementCSSInlineStyle", "ElementContentEditable", "GlobalEventHandlers", "HTMLOrSVGElement", "Node", "ARIAMixin", "Animatable", "ChildNode", "NonDocumentTypeChildNode", "ParentNode", "Slottable", "EventTarget"]},
	"HTMLHeadElement": {Class: () => dom.HTMLHeadElement, lineage: ["HTMLHeadElement", "HTMLElement", "Element", "ElementCSSInlineStyle", "ElementContentEditable", "GlobalEventHandlers", "HTMLOrSVGElement", "Node", "ARIAMixin", "Animatable", "ChildNode", "NonDocumentTypeChildNode", "ParentNode", "Slottable", "EventTarget"]},
	"HTMLImageElement": {Class: () => dom.HTMLImageElement, lineage: ["HTMLImageElement", "HTMLElement", "Element", "ElementCSSInlineStyle", "ElementContentEditable", "GlobalEventHandlers", "HTMLOrSVGElement", "Node", "ARIAMixin", "Animatable", "ChildNode", "NonDocumentTypeChildNode", "ParentNode", "Slottable", "EventTarget"]},
	"HTMLImageElementType": {Class: () => dom.HTMLImageElementType, lineage: ["HTMLImageElementType"]},
	"HTMLMediaElement": {Class: () => dom.HTMLMediaElement, lineage: ["HTMLMediaElement", "HTMLElement", "Element", "ElementCSSInlineStyle", "ElementContentEditable", "GlobalEventHandlers", "HTMLOrSVGElement", "Node", "ARIAMixin", "Animatable", "ChildNode", "NonDocumentTypeChildNode", "ParentNode", "Slottable", "EventTarget"]},
	"HTMLOrSVGElement": {Class: () => dom.HTMLOrSVGElement, lineage: ["HTMLOrSVGElement"]},
	"HTMLSlotElement": {Class: () => dom.HTMLSlotElement, lineage: ["HTMLSlotElement", "HTMLElement", "Element", "ElementCSSInlineStyle", "ElementContentEditable", "GlobalEventHandlers", "HTMLOrSVGElement", "Node", "ARIAMixin", "Animatable", "ChildNode", "NonDocumentTypeChildNode", "ParentNode", "Slottable", "EventTarget"]},
	"HashChangeEvent": {Class: () => dom.HashChangeEvent, lineage: ["HashChangeEvent", "Event"]},
	"Headers": {Class: () => dom.Headers, lineage: ["Headers"]},
	"History": {Class: () => dom.History, lineage: ["History"]},
	"IDBFactory": {Class: () => dom.IDBFactory, lineage: ["IDBFactory"]},
	"IIRFilterNode": {Class: () => dom.IIRFilterNode, lineage: ["IIRFilterNode", "AudioNode", "EventTarget"]},
	"IdleDeadline": {Class: () => dom.IdleDeadline, lineage: ["IdleDeadline"]},
	"ImageBitmap": {Class: () => dom.ImageBitmap, lineage: ["ImageBitmap"]},
	"ImageDataType": {Class: () => dom.ImageDataType, lineage: ["ImageDataType"]},
	"InputEvent": {Class: () => dom.InputEvent, lineage: ["InputEvent", "UIEvent", "Event"]},
	"KeyboardEvent": {Class: () => dom.KeyboardEvent, lineage: ["KeyboardEvent", "UIEvent", "Event"]},
	"KeyboardEventType": {Class: () => dom.KeyboardEventType, lineage: ["KeyboardEventType"]},
	"LinkStyle": {Class: () => dom.LinkStyle, lineage: ["LinkStyle"]},
	"Location": {Class: () => dom.Location, lineage: ["Location"]},
	"LockManager": {Class: () => dom.LockManager, lineage: ["LockManager"]},
	"MIDIAccess": {Class: () => dom.MIDIAccess, lineage: ["MIDIAccess", "EventTarget"]},
	"MediaCapabilities": {Class: () => dom.MediaCapabilities, lineage: ["MediaCapabilities"]},
	"MediaDevices": {Class: () => dom.MediaDevices, lineage: ["MediaDevices", "EventTarget"]},
	"MediaElementAudioSourceNode": {Class: () => dom.MediaElementAudioSourceNode, lineage: ["MediaElementAudioSourceNode", "AudioNode", "EventTarget"]},
	"MediaEncryptedEvent": {Class: () => dom.MediaEncryptedEvent, lineage: ["MediaEncryptedEvent", "Event"]},
	"MediaKeySession": {Class: () => dom.MediaKeySession, lineage: ["MediaKeySession", "EventTarget"]},
	"MediaKeySessionType": {Class: () => dom.MediaKeySessionType, lineage: ["MediaKeySessionType"]},
	"MediaKeySystemAccess": {Class: () => dom.MediaKeySystemAccess, lineage: ["MediaKeySystemAccess"]},
	"MediaKeys": {Class: () => dom.MediaKeys, lineage: ["MediaKeys"]},
	"MediaQueryList": {Class: () => dom.MediaQueryList, lineage: ["MediaQueryList", "EventTarget"]},
	"MediaSession": {Class: () => dom.MediaSession, lineage: ["MediaSession"]},
	"MediaStream": {Class: () => dom.MediaStream, lineage: ["MediaStream", "EventTarget"]},
	"MediaStreamAudioDestinationNode": {Class: () => dom.MediaStreamAudioDestinationNode, lineage: ["MediaStreamAudioDestinationNode", "AudioNode", "EventTarget"]},
	"MediaStreamAudioSourceNode": {Class: () => dom.MediaStreamAudioSourceNode, lineage: ["MediaStreamAudioSourceNode", "AudioNode", "EventTarget"]},
	"MediaStreamTrack": {Class: () => dom.MediaStreamTrack, lineage: ["MediaStreamTrack", "EventTarget"]},
	"MediaStreamTrackEvent": {Class: () => dom.MediaStreamTrackEvent, lineage: ["MediaStreamTrackEvent", "Event"]},
	"MessageEvent": {Class: () => dom.MessageEvent, lineage: ["MessageEvent", "Event"]},
	"MimeTypeArray": {Class: () => dom.MimeTypeArray, lineage: ["MimeTypeArray"]},
	"MouseEvent": {Class: () => dom.MouseEvent, lineage: ["MouseEvent", "UIEvent", "Event"]},
	"MouseEventType": {Class: () => dom.MouseEventType, lineage: ["MouseEventType"]},
	"NamedNodeMap": {Class: () => dom.NamedNodeMap, lineage: ["NamedNodeMap"]},
	"Navigator": {Class: () => dom.Navigator, lineage: ["Navigator", "NavigatorAutomationInformation", "NavigatorBadge", "NavigatorConcurrentHardware", "NavigatorContentUtils", "NavigatorCookies", "NavigatorID", "NavigatorLanguage", "NavigatorLocks", "NavigatorOnLine", "NavigatorPlugins", "NavigatorStorage"]},
	"NavigatorAutomationInformation": {Class: () => dom.NavigatorAutomationInformation, lineage: ["NavigatorAutomationInformation"]},
	"NavigatorBadge": {Class: () => dom.NavigatorBadge, lineage: ["NavigatorBadge"]},
	"NavigatorConcurrentHardware": {Class: () => dom.NavigatorConcurrentHardware, lineage: ["NavigatorConcurrentHardware"]},
	"NavigatorContentUtils": {Class: () => dom.NavigatorContentUtils, lineage: ["NavigatorContentUtils"]},
	"NavigatorCookies": {Class: () => dom.NavigatorCookies, lineage: ["NavigatorCookies"]},
	"NavigatorID": {Class: () => dom.NavigatorID, lineage: ["NavigatorID"]},
	"NavigatorLanguage": {Class: () => dom.NavigatorLanguage, lineage: ["NavigatorLanguage"]},
	"NavigatorLocks": {Class: () => dom.NavigatorLocks, lineage: ["NavigatorLocks"]},
	"NavigatorOnLine": {Class: () => dom.NavigatorOnLine, lineage: ["NavigatorOnLine"]},
	"NavigatorPlugins": {Class: () => dom.NavigatorPlugins, lineage: ["NavigatorPlugins"]},
	"NavigatorStorage": {Class: () => dom.NavigatorStorage, lineage: ["NavigatorStorage"]},
	"Node": {Class: () => dom.Node, lineage: ["Node", "EventTarget"]},
	"NodeIterator": {Class: () => dom.NodeIterator, lineage: ["NodeIterator"]},
	"NodeList": {Class: () => dom.NodeList, lineage: ["NodeList"]},
	"NodeListOf": {Class: () => dom.NodeListOf, lineage: ["NodeListOf", "NodeList"]},
	"NonDocumentTypeChildNode": {Class: () => dom.NonDocumentTypeChildNode, lineage: ["NonDocumentTypeChildNode"]},
	"NonElementParentNode": {Class: () => dom.NonElementParentNode, lineage: ["NonElementParentNode"]},
	"OffscreenCanvas": {Class: () => dom.OffscreenCanvas, lineage: ["OffscreenCanvas", "EventTarget"]},
	"OffscreenCanvasRenderingContext2D": {Class: () => dom.OffscreenCanvasRenderingContext2D, lineage: ["OffscreenCanvasRenderingContext2D", "CanvasCompositing", "CanvasDrawImage", "CanvasDrawPath", "CanvasFillStrokeStyles", "CanvasFilters", "CanvasImageData", "CanvasImageSmoothing", "CanvasPath", "CanvasPathDrawingStyles", "CanvasRect", "CanvasShadowStyles", "CanvasState", "CanvasText", "CanvasTextDrawingStyles", "CanvasTransform"]},
	"OscillatorNode": {Class: () => dom.OscillatorNode, lineage: ["OscillatorNode", "AudioScheduledSourceNode", "AudioNode", "EventTarget"]},
	"PageTransitionEvent": {Class: () => dom.PageTransitionEvent, lineage: ["PageTransitionEvent", "Event"]},
	"PannerNode": {Class: () => dom.PannerNode, lineage: ["PannerNode", "AudioNode", "EventTarget"]},
	"ParentNode": {Class: () => dom.ParentNode, lineage: ["ParentNode", "Node", "EventTarget"]},
	"Path2D": {Class: () => dom.Path2D, lineage: ["Path2D", "CanvasPath"]},
	"Path2DType": {Class: () => dom.Path2DType, lineage: ["Path2DType"]},
	"Performance": {Class: () => dom.Performance, lineage: ["Performance", "EventTarget"]},
	"PerformanceEntry": {Class: () => dom.PerformanceEntry, lineage: ["PerformanceEntry"]},
	"PerformanceMark": {Class: () => dom.PerformanceMark, lineage: ["PerformanceMark", "PerformanceEntry"]},
	"PerformanceMeasure": {Class: () => dom.PerformanceMeasure, lineage: ["PerformanceMeasure", "PerformanceEntry"]},
	"PerformanceNavigation": {Class: () => dom.PerformanceNavigation, lineage: ["PerformanceNavigation"]},
	"PerformanceTiming": {Class: () => dom.PerformanceTiming, lineage: ["PerformanceTiming"]},
	"Permissions": {Class: () => dom.Permissions, lineage: ["Permissions"]},
	"PluginArray": {Class: () => dom.PluginArray, lineage: ["PluginArray"]},
	"PointerEvent": {Class: () => dom.PointerEvent, lineage: ["PointerEvent", "MouseEvent", "UIEvent", "Event"]},
	"PointerEventType": {Class: () => dom.PointerEventType, lineage: ["PointerEventType"]},
	"PopStateEvent": {Class: () => dom.PopStateEvent, lineage: ["PopStateEvent", "Event"]},
	"ProcessingInstruction": {Class: () => dom.ProcessingInstruction, lineage: ["ProcessingInstruction", "CharacterData", "LinkStyle", "Node", "ChildNode", "NonDocumentTypeChildNode", "EventTarget"]},
	"ProgressEvent": {Class: () => dom.ProgressEvent, lineage: ["ProgressEvent", "Event"]},
	"PromiseRejectionEvent": {Class: () => dom.PromiseRejectionEvent, lineage: ["PromiseRejectionEvent", "Event"]},
	"Range": {Class: () => dom.Range, lineage: ["Range", "AbstractRange"]},
	"ReadableStream": {Class: () => dom.ReadableStream, lineage: ["ReadableStream"]},
	"RemotePlayback": {Class: () => dom.RemotePlayback, lineage: ["RemotePlayback", "EventTarget"]},
	"Response": {Class: () => dom.Response, lineage: ["Response", "Body"]},
	"ResponseType": {Class: () => dom.ResponseType, lineage: ["ResponseType"]},
	"SVGElement": {Class: () => dom.SVGElement, lineage: ["SVGElement", "Element", "ElementCSSInlineStyle", "GlobalEventHandlers", "HTMLOrSVGElement", "Node", "ARIAMixin", "Animatable", "ChildNode", "NonDocumentTypeChildNode", "ParentNode", "Slottable", "EventTarget"]},
	"SVGFitToViewBox": {Class: () => dom.SVGFitToViewBox, lineage: ["SVGFitToViewBox"]},
	"SVGGraphicsElement": {Class: () => dom.SVGGraphicsElement, lineage: ["SVGGraphicsElement", "SVGElement", "SVGTests", "Element", "ElementCSSInlineStyle", "GlobalEventHandlers", "HTMLOrSVGElement", "Node", "ARIAMixin", "Animatable", "ChildNode", "NonDocumentTypeChildNode", "ParentNode", "Slottable", "EventTarget"]},
	"SVGSVGElement": {Class: () => dom.SVGSVGElement, lineage: ["SVGSVGElement", "SVGGraphicsElement", "SVGFitToViewBox", "WindowEventHandlers", "SVGElement", "SVGTests", "Element", "ElementCSSInlineStyle", "GlobalEventHandlers", "HTMLOrSVGElement", "Node", "ARIAMixin", "Animatable", "ChildNode", "NonDocumentTypeChildNode", "ParentNode", "Slottable", "EventTarget"]},
	"SVGTests": {Class: () => dom.SVGTests, lineage: ["SVGTests"]},
	"ScriptProcessorNode": {Class: () => dom.ScriptProcessorNode, lineage: ["ScriptProcessorNode", "AudioNode", "EventTarget"]},
	"SecurityPolicyViolationEvent": {Class: () => dom.SecurityPolicyViolationEvent, lineage: ["SecurityPolicyViolationEvent", "Event"]},
	"Selection": {Class: () => dom.Selection, lineage: ["Selection"]},
	"ServiceWorkerContainer": {Class: () => dom.ServiceWorkerContainer, lineage: ["ServiceWorkerContainer", "EventTarget"]},
	"ShadowRoot": {Class: () => dom.ShadowRoot, lineage: ["ShadowRoot", "DocumentFragment", "DocumentOrShadowRoot", "Node", "NonElementParentNode", "ParentNode", "EventTarget"]},
	"Slottable": {Class: () => dom.Slottable, lineage: ["Slottable"]},
	"SpeechSynthesis": {Class: () => dom.SpeechSynthesis, lineage: ["SpeechSynthesis", "EventTarget"]},
	"StereoPannerNode": {Class: () => dom.StereoPannerNode, lineage: ["StereoPannerNode", "AudioNode", "EventTarget"]},
	"Storage": {Class: () => dom.Storage, lineage: ["Storage"]},
	"StorageEvent": {Class: () => dom.StorageEvent, lineage: ["StorageEvent", "Event"]},
	"StorageManager": {Class: () => dom.StorageManager, lineage: ["StorageManager"]},
	"StylePropertyMap": {Class: () => dom.StylePropertyMap, lineage: ["StylePropertyMap", "StylePropertyMapReadOnly"]},
	"StylePropertyMapReadOnly": {Class: () => dom.StylePropertyMapReadOnly, lineage: ["StylePropertyMapReadOnly"]},
	"StyleSheet": {Class: () => dom.StyleSheet, lineage: ["StyleSheet"]},
	"StyleSheetList": {Class: () => dom.StyleSheetList, lineage: ["StyleSheetList"]},
	"SubmitEvent": {Class: () => dom.SubmitEvent, lineage: ["SubmitEvent", "Event"]},
	"Text": {Class: () => dom.Text, lineage: ["Text", "CharacterData", "Slottable", "Node", "ChildNode", "NonDocumentTypeChildNode", "EventTarget"]},
	"TextTrack": {Class: () => dom.TextTrack, lineage: ["TextTrack", "EventTarget"]},
	"TextTrackCue": {Class: () => dom.TextTrackCue, lineage: ["TextTrackCue", "EventTarget"]},
	"TextTrackCueList": {Class: () => dom.TextTrackCueList, lineage: ["TextTrackCueList"]},
	"TextTrackList": {Class: () => dom.TextTrackList, lineage: ["TextTrackList", "EventTarget"]},
	"TimeRanges": {Class: () => dom.TimeRanges, lineage: ["TimeRanges"]},
	"TouchEvent": {Class: () => dom.TouchEvent, lineage: ["TouchEvent", "UIEvent", "Event"]},
	"TouchEventType": {Class: () => dom.TouchEventType, lineage: ["TouchEventType"]},
	"TouchList": {Class: () => dom.TouchList, lineage: ["TouchList"]},
	"TrackEvent": {Class: () => dom.TrackEvent, lineage: ["TrackEvent", "Event"]},
	"TransitionEvent": {Class: () => dom.TransitionEvent, lineage: ["TransitionEvent", "Event"]},
	"TreeWalker": {Class: () => dom.TreeWalker, lineage: ["TreeWalker"]},
	"UIEvent": {Class: () => dom.UIEvent, lineage: ["UIEvent", "Event"]},
	"ViewTransition": {Class: () => dom.ViewTransition, lineage: ["ViewTransition"]},
	"VisualViewport": {Class: () => dom.VisualViewport, lineage: ["VisualViewport", "EventTarget"]},
	"WakeLock": {Class: () => dom.WakeLock, lineage: ["WakeLock"]},
	"WaveShaperNode": {Class: () => dom.WaveShaperNode, lineage: ["WaveShaperNode", "AudioNode", "EventTarget"]},
	"WheelEvent": {Class: () => dom.WheelEvent, lineage: ["WheelEvent", "MouseEvent", "UIEvent", "Event"]},
	"WheelEventType": {Class: () => dom.WheelEventType, lineage: ["WheelEventType"]},
	"Window": {Class: () => dom.Window, lineage: ["Window", "EventTarget", "AnimationFrameProvider", "GlobalEventHandlers", "WindowEventHandlers", "WindowLocalStorage", "WindowOrWorkerGlobalScope", "WindowSessionStorage"]},
	"WindowEventHandlers": {Class: () => dom.WindowEventHandlers, lineage: ["WindowEventHandlers"]},
	"WindowLocalStorage": {Class: () => dom.WindowLocalStorage, lineage: ["WindowLocalStorage"]},
	"WindowOrWorkerGlobalScope": {Class: () => dom.WindowOrWorkerGlobalScope, lineage: ["WindowOrWorkerGlobalScope"]},
	"WindowSessionStorage": {Class: () => dom.WindowSessionStorage, lineage: ["WindowSessionStorage"]},
	"Worklet": {Class: () => dom.Worklet, lineage: ["Worklet"]},
	"XPathEvaluatorBase": {Class: () => dom.XPathEvaluatorBase, lineage: ["XPathEvaluatorBase"]},
	"XPathExpression": {Class: () => dom.XPathExpression, lineage: ["XPathExpression"]},
	"XPathResult": {Class: () => dom.XPathResult, lineage: ["XPathResult"]},
};
