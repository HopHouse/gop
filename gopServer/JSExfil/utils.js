const getCircularReplacer = () => {
    const seen = new WeakSet();
    return (key, value) => {
        if (typeof value === "object" && value !== null) {
            if (seen.has(value)) {
                return;
            }
            seen.add(value);
        }
        return value;
    };
};

var objs = []; // we'll store the object references in this array

function walkTheObject(obj) {
    walkTheObjectForDepth(obj, 8)

    return JSON.stringify(objs, getCircularReplacer());
}

function walkTheObjectForDepth(obj, depth) {
    if (depth == 0) {
        return "Recursion stopped"
    }

    var keys = Object.keys(obj); // get all own property names of the object

    keys.forEach(function (key) {

        var value = obj[key]; // get property value

        // if the property value is an object...
        if (value && typeof value === 'object') {
            // if we don't have this reference...

            if (objs.indexOf(value) < 0) {
                objs.push(value); // store the reference
                walkTheObjectForDepth(value, depth - 1); // traverse all its own properties
            }
        }

    });
}