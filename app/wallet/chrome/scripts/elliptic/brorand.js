var _nodeResolve_empty = {};
var _nodeResolve_empty$1 = /* @__PURE__ */ Object.freeze({
  __proto__: null,
  default: _nodeResolve_empty
});
function getDefaultExportFromNamespaceIfNotNamed(n) {
  return n && Object.prototype.hasOwnProperty.call(n, "default") && Object.keys(n).length === 1 ? n["default"] : n;
}
var require$$0 = /* @__PURE__ */ getDefaultExportFromNamespaceIfNotNamed(_nodeResolve_empty$1);
var r;
var brorand = function rand(len) {
  if (!r)
    r = new Rand(null);
  return r.generate(len);
};
function Rand(rand2) {
  this.rand = rand2;
}
var Rand_1 = Rand;
Rand.prototype.generate = function generate(len) {
  return this._rand(len);
};
Rand.prototype._rand = function _rand(n) {
  if (this.rand.getBytes)
    return this.rand.getBytes(n);
  var res = new Uint8Array(n);
  for (var i = 0; i < res.length; i++)
    res[i] = this.rand.getByte();
  return res;
};
if (typeof self === "object") {
  if (self.crypto && self.crypto.getRandomValues) {
    Rand.prototype._rand = function _rand2(n) {
      var arr = new Uint8Array(n);
      self.crypto.getRandomValues(arr);
      return arr;
    };
  } else if (self.msCrypto && self.msCrypto.getRandomValues) {
    Rand.prototype._rand = function _rand2(n) {
      var arr = new Uint8Array(n);
      self.msCrypto.getRandomValues(arr);
      return arr;
    };
  } else if (typeof window === "object") {
    Rand.prototype._rand = function() {
      throw new Error("Not implemented yet");
    };
  }
} else {
  try {
    var crypto = require$$0;
    if (typeof crypto.randomBytes !== "function")
      throw new Error("Not supported");
    Rand.prototype._rand = function _rand2(n) {
      return crypto.randomBytes(n);
    };
  } catch (e) {
  }
}
brorand.Rand = Rand_1;
export default brorand;
export {Rand_1 as Rand, brorand as __moduleExports};
