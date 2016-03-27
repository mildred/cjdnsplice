var Cjdnsplice = require('./index.js');

const PATH = [
    {"labelN":"0000.0000.0000.0015","encodingScheme":Cjdnsplice.SCHEME_358},
    {"labelN":"0000.0000.0000.008e","labelP":"0000.0000.0000.009e","encodingScheme":Cjdnsplice.SCHEME_358},
    {"labelN":"0000.0000.0000.00a2","labelP":"0000.0000.0000.0013","encodingScheme":Cjdnsplice.SCHEME_358},
    {"labelN":"0000.0000.0000.001d","labelP":"0000.0000.0000.001b","encodingScheme":Cjdnsplice.SCHEME_358},
    {"labelN":"0000.0000.0000.001b","labelP":"0000.0000.0000.00ee","encodingScheme":Cjdnsplice.SCHEME_358},
    {"labelN":"0000.0000.0000.001b","labelP":"0000.0000.0000.0019","encodingScheme":Cjdnsplice.SCHEME_358},
    {"labelP":"0000.0000.0000.0013","encodingScheme":Cjdnsplice.SCHEME_358}
];

const buildLabelTest = () => {
    const builtPath = Cjdnsplice.buildLabel(PATH);
    if (builtPath.label !== '0000.0003.64b5.10e5') { throw new Error("incorrect label"); }
};

const spliceTest = () => {
    const result = Cjdnsplice.splice.apply(null, [
        "0000.0000.0000.0015",
        "0000.0000.0000.008e",
        "0000.0000.0000.00a2",
        "0000.0000.0000.001d",
        "0000.0000.0000.0414",
        "0000.0000.0000.001b"
    ].reverse());
    if (result !== "0000.001b.0535.10e5") { throw new Error(); }
};

var randLabel = function () {
    var out = [];
    for (var i = 0; i < 4; i++) {
        var x = Math.random().toString(16);
        if (x.length < 6) { i--; continue; }
        out.push(x.substring(x.length-4));
    }
    return out.join('.');
};

var testBitsToLabel = function () {
    for (var i = 0; i < 1000; i++) {
        var x = randLabel();
        if (Cjdnsplice._bitsToLabel(Cjdnsplice._labelToBits(x)) !== x) {
            throw new Error(x);
        }
    }
};

buildLabelTest();
spliceTest();
testBitsToLabel();
