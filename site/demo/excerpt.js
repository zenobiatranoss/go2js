class Rect {
    W;
    H;
    constructor() {
        this.W = 0;
        this.H = 0;
    }
}

Rect.prototype.Area$ = function() {
    return this.W * this.H;
};

function* Sum(__go2js_type0, xs) {
    let total = null;
    for (const x of go2jsRangeSequence(xs)) {
        total += x;
    }
    return total;
}
