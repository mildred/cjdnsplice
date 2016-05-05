'use strict';
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

const reEncode = (label, type) => {
    return Cjdnsplice.reEncode(label, Cjdnsplice.SCHEME_358, type);
};

const cannonicalize = (label) => {
    return reEncode(label, Cjdnsplice.FORM_CANNONICAL);
};

var reEncodeTest = () => {
    for (let i = 0; i < 256; i++) {
        if (i === 0) { continue; }
        const bits = (i).toString(2).split('').map((x)=>(Number(x)));
        let formNum;
        if (bits.length <= 3) {
            while (bits.length < 3) { bits.unshift(0); }
            bits.push(1);
            formNum = 2;
        } else if (bits.length <= 5) {
            while (bits.length < 5) { bits.unshift(0); }
            bits.push(1,0);
            formNum = 1;
        } else if (bits.length <= 8) {
            while (bits.length < 8) { bits.unshift(0); }
            bits.push(0,0);
            formNum = 0;
        } else {
            throw new Error(bits.length);
        }
        bits.unshift(1);
        while (bits.length < 64) { bits.unshift(0); }
        const label = Cjdnsplice._bitsToLabel(bits);
        let label2 = "";
        let label3 = "";
        if (formNum > 0) {
            label2 = reEncode(label, 0);
            if (reEncode(label2, formNum) !== label) { throw new Error(); }
            if (cannonicalize(label2) !== label) { throw new Error(); }
            if (formNum > 1) {
                label3 = reEncode(label, 1);
                if (reEncode(label2, 1) !== label3) { throw new Error(); }
                if (reEncode(label3, 0) !== label2) { throw new Error(); }
                if (reEncode(label3, 2) !== label) { throw new Error(); }
                if (cannonicalize(label3) !== label) { throw new Error(); }
            }
        }
        //console.log(label + " " + label2 + " " + label3);
    }
};

cannonicalize("0000.0000.0000.009e");
cannonicalize("0000.0000.0000.041c");
buildLabelTest();
spliceTest();
testBitsToLabel();
reEncodeTest();
