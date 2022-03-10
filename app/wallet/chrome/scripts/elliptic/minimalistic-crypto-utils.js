function createCommonjsModule(fn, basedir, module) {
  return module = {
    path: basedir,
    exports: {},
    require: function(path, base) {
      return commonjsRequire(path, base === void 0 || base === null ? module.path : base);
    }
  }, fn(module, module.exports), module.exports;
}
function commonjsRequire() {
  throw new Error("Dynamic requires are not currently supported by @rollup/plugin-commonjs");
}
var utils_1 = createCommonjsModule(function(module, exports) {
  var utils = exports;
  function toArray2(msg, enc) {
    if (Array.isArray(msg))
      return msg.slice();
    if (!msg)
      return [];
    var res = [];
    if (typeof msg !== "string") {
      for (var i = 0; i < msg.length; i++)
        res[i] = msg[i] | 0;
      return res;
    }
    if (enc === "hex") {
      msg = msg.replace(/[^a-z0-9]+/ig, "");
      if (msg.length % 2 !== 0)
        msg = "0" + msg;
      for (var i = 0; i < msg.length; i += 2)
        res.push(parseInt(msg[i] + msg[i + 1], 16));
    } else {
      for (var i = 0; i < msg.length; i++) {
        var c = msg.charCodeAt(i);
        var hi = c >> 8;
        var lo = c & 255;
        if (hi)
          res.push(hi, lo);
        else
          res.push(lo);
      }
    }
    return res;
  }
  utils.toArray = toArray2;
  function zero22(word) {
    if (word.length === 1)
      return "0" + word;
    else
      return word;
  }
  utils.zero2 = zero22;
  function toHex2(msg) {
    var res = "";
    for (var i = 0; i < msg.length; i++)
      res += zero22(msg[i].toString(16));
    return res;
  }
  utils.toHex = toHex2;
  utils.encode = function encode2(arr, enc) {
    if (enc === "hex")
      return toHex2(arr);
    else
      return arr;
  };
});
export default utils_1;
var encode = utils_1.encode;
var toArray = utils_1.toArray;
var toHex = utils_1.toHex;
var zero2 = utils_1.zero2;
export {utils_1 as __moduleExports, encode, toArray, toHex, zero2};
