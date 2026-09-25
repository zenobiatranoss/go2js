function go2jsLen(value) {
        if (value instanceof Map) {
                return value.size;
        }
        return value.length;
}

function go2jsCap(value) {
        return value.length;
}

function go2jsAppend(value, ...items) {
        return value.concat(items);
}

function go2jsMake(type, size) {
        if (typeof size === "number") {
                return new Array(size);
        }

        return [];
}

function go2jsMakeMap() {
        return new Map();
}

function go2jsMap(entries) {
        const map = new Map();
        for (const entry of entries) {
                map.set(entry[0], entry[1]);
        }
        return map;
}

function go2jsMapGet(map, key) {
        return map.get(key);
}

function go2jsMapSet(map, key, value) {
        map.set(key, value);
}

function go2jsMapDelete(map, key) {
        map.delete(key);
}

class User {
    Name;
    Age;
}

function add(a, b) {
    return a + b;
}

function main() {
    let result = add(10, 20);
    if (result > 20) {
        console.log("result:", result);
    }
    for (let i = 0; i < 3; i++) {
        console.log(i);
    }
    let values = [10, 20, 30];
    console.log(go2jsLen(values));
}

main();
