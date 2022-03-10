import BN from './bnjs.js';
import minAssert from "./minimalistic-assert.js";
import minUtils from "./minimalistic-crypto-utils.js";
import rand from "./brorand.js";
import inherits2 from "./inherits.js";
import hash from "./hashjs.js";
import HmacDRBG from "./hmac-drbg.js";

function createCommonjsModule(fn, basedir, module) {
  return module = {
    path: basedir,
    exports: {},
    require: function(path, base2) {
      return commonjsRequire(path, base2 === void 0 || base2 === null ? module.path : base2);
    }
  }, fn(module, module.exports), module.exports;
}
function commonjsRequire() {
  throw new Error("Dynamic requires are not currently supported by @rollup/plugin-commonjs");
}
const name = "elliptic";
const version = "6.5.4";
const description = "EC cryptography";
const main = "lib/elliptic.js";
const files = [
  "lib"
];
const scripts = {
  lint: "eslint lib test",
  "lint:fix": "npm run lint -- --fix",
  unit: "istanbul test _mocha --reporter=spec test/index.js",
  test: "npm run lint && npm run unit",
  version: "grunt dist && git add dist/"
};
const repository = {
  type: "git",
  url: "git@github.com:indutny/elliptic"
};
const keywords = [
  "EC",
  "Elliptic",
  "curve",
  "Cryptography"
];
const author = "Fedor Indutny <fedor@indutny.com>";
const license = "MIT";
const bugs = {
  url: "https://github.com/indutny/elliptic/issues"
};
const homepage = "https://github.com/indutny/elliptic";
const devDependencies = {
  brfs: "^2.0.2",
  coveralls: "^3.1.0",
  eslint: "^7.6.0",
  grunt: "^1.2.1",
  "grunt-browserify": "^5.3.0",
  "grunt-cli": "^1.3.2",
  "grunt-contrib-connect": "^3.0.0",
  "grunt-contrib-copy": "^1.0.0",
  "grunt-contrib-uglify": "^5.0.0",
  "grunt-mocha-istanbul": "^5.0.2",
  "grunt-saucelabs": "^9.0.1",
  istanbul: "^0.4.5",
  mocha: "^8.0.1"
};
const dependencies = {
  "bn.js": "^4.11.9",
  brorand: "^1.1.0",
  "hash.js": "^1.0.0",
  "hmac-drbg": "^1.0.1",
  inherits: "^2.0.4",
  "minimalistic-assert": "^1.0.1",
  "minimalistic-crypto-utils": "^1.0.1"
};
var require$$0 = {
  name,
  version,
  description,
  main,
  files,
  scripts,
  repository,
  keywords,
  author,
  license,
  bugs,
  homepage,
  devDependencies,
  dependencies
};
var utils_1 = createCommonjsModule(function(module, exports) {
  var utils = exports;
  utils.assert = minAssert;
  utils.toArray = minUtils.toArray;
  utils.zero2 = minUtils.zero2;
  utils.toHex = minUtils.toHex;
  utils.encode = minUtils.encode;
  function getNAF2(num, w, bits) {
    var naf = new Array(Math.max(num.bitLength(), bits) + 1);
    naf.fill(0);
    var ws = 1 << w + 1;
    var k = num.clone();
    for (var i = 0; i < naf.length; i++) {
      var z;
      var mod = k.andln(ws - 1);
      if (k.isOdd()) {
        if (mod > (ws >> 1) - 1)
          z = (ws >> 1) - mod;
        else
          z = mod;
        k.isubn(z);
      } else {
        z = 0;
      }
      naf[i] = z;
      k.iushrn(1);
    }
    return naf;
  }
  utils.getNAF = getNAF2;
  function getJSF2(k1, k2) {
    var jsf = [
      [],
      []
    ];
    k1 = k1.clone();
    k2 = k2.clone();
    var d1 = 0;
    var d2 = 0;
    var m8;
    while (k1.cmpn(-d1) > 0 || k2.cmpn(-d2) > 0) {
      var m14 = k1.andln(3) + d1 & 3;
      var m24 = k2.andln(3) + d2 & 3;
      if (m14 === 3)
        m14 = -1;
      if (m24 === 3)
        m24 = -1;
      var u1;
      if ((m14 & 1) === 0) {
        u1 = 0;
      } else {
        m8 = k1.andln(7) + d1 & 7;
        if ((m8 === 3 || m8 === 5) && m24 === 2)
          u1 = -m14;
        else
          u1 = m14;
      }
      jsf[0].push(u1);
      var u2;
      if ((m24 & 1) === 0) {
        u2 = 0;
      } else {
        m8 = k2.andln(7) + d2 & 7;
        if ((m8 === 3 || m8 === 5) && m14 === 2)
          u2 = -m24;
        else
          u2 = m24;
      }
      jsf[1].push(u2);
      if (2 * d1 === u1 + 1)
        d1 = 1 - d1;
      if (2 * d2 === u2 + 1)
        d2 = 1 - d2;
      k1.iushrn(1);
      k2.iushrn(1);
    }
    return jsf;
  }
  utils.getJSF = getJSF2;
  function cachedProperty2(obj, name2, computer) {
    var key2 = "_" + name2;
    obj.prototype[name2] = function cachedProperty3() {
      return this[key2] !== void 0 ? this[key2] : this[key2] = computer.call(this);
    };
  }
  utils.cachedProperty = cachedProperty2;
  function parseBytes2(bytes) {
    return typeof bytes === "string" ? utils.toArray(bytes, "hex") : bytes;
  }
  utils.parseBytes = parseBytes2;
  function intFromLE(bytes) {
    return new BN(bytes, "hex", "le");
  }
  utils.intFromLE = intFromLE;
});
var getNAF = utils_1.getNAF;
var getJSF = utils_1.getJSF;
var assert = utils_1.assert;
function BaseCurve(type, conf) {
  this.type = type;
  this.p = new BN(conf.p, 16);
  this.red = conf.prime ? BN.red(conf.prime) : BN.mont(this.p);
  this.zero = new BN(0).toRed(this.red);
  this.one = new BN(1).toRed(this.red);
  this.two = new BN(2).toRed(this.red);
  this.n = conf.n && new BN(conf.n, 16);
  this.g = conf.g && this.pointFromJSON(conf.g, conf.gRed);
  this._wnafT1 = new Array(4);
  this._wnafT2 = new Array(4);
  this._wnafT3 = new Array(4);
  this._wnafT4 = new Array(4);
  this._bitLength = this.n ? this.n.bitLength() : 0;
  var adjustCount = this.n && this.p.div(this.n);
  if (!adjustCount || adjustCount.cmpn(100) > 0) {
    this.redN = null;
  } else {
    this._maxwellTrick = true;
    this.redN = this.n.toRed(this.red);
  }
}
var base = BaseCurve;
BaseCurve.prototype.point = function point() {
  throw new Error("Not implemented");
};
BaseCurve.prototype.validate = function validate() {
  throw new Error("Not implemented");
};
BaseCurve.prototype._fixedNafMul = function _fixedNafMul(p, k) {
  assert(p.precomputed);
  var doubles = p._getDoubles();
  var naf = getNAF(k, 1, this._bitLength);
  var I = (1 << doubles.step + 1) - (doubles.step % 2 === 0 ? 2 : 1);
  I /= 3;
  var repr = [];
  var j;
  var nafW;
  for (j = 0; j < naf.length; j += doubles.step) {
    nafW = 0;
    for (var l = j + doubles.step - 1; l >= j; l--)
      nafW = (nafW << 1) + naf[l];
    repr.push(nafW);
  }
  var a = this.jpoint(null, null, null);
  var b = this.jpoint(null, null, null);
  for (var i = I; i > 0; i--) {
    for (j = 0; j < repr.length; j++) {
      nafW = repr[j];
      if (nafW === i)
        b = b.mixedAdd(doubles.points[j]);
      else if (nafW === -i)
        b = b.mixedAdd(doubles.points[j].neg());
    }
    a = a.add(b);
  }
  return a.toP();
};
BaseCurve.prototype._wnafMul = function _wnafMul(p, k) {
  var w = 4;
  var nafPoints = p._getNAFPoints(w);
  w = nafPoints.wnd;
  var wnd = nafPoints.points;
  var naf = getNAF(k, w, this._bitLength);
  var acc = this.jpoint(null, null, null);
  for (var i = naf.length - 1; i >= 0; i--) {
    for (var l = 0; i >= 0 && naf[i] === 0; i--)
      l++;
    if (i >= 0)
      l++;
    acc = acc.dblp(l);
    if (i < 0)
      break;
    var z = naf[i];
    assert(z !== 0);
    if (p.type === "affine") {
      if (z > 0)
        acc = acc.mixedAdd(wnd[z - 1 >> 1]);
      else
        acc = acc.mixedAdd(wnd[-z - 1 >> 1].neg());
    } else {
      if (z > 0)
        acc = acc.add(wnd[z - 1 >> 1]);
      else
        acc = acc.add(wnd[-z - 1 >> 1].neg());
    }
  }
  return p.type === "affine" ? acc.toP() : acc;
};
BaseCurve.prototype._wnafMulAdd = function _wnafMulAdd(defW, points, coeffs, len, jacobianResult) {
  var wndWidth = this._wnafT1;
  var wnd = this._wnafT2;
  var naf = this._wnafT3;
  var max = 0;
  var i;
  var j;
  var p;
  for (i = 0; i < len; i++) {
    p = points[i];
    var nafPoints = p._getNAFPoints(defW);
    wndWidth[i] = nafPoints.wnd;
    wnd[i] = nafPoints.points;
  }
  for (i = len - 1; i >= 1; i -= 2) {
    var a = i - 1;
    var b = i;
    if (wndWidth[a] !== 1 || wndWidth[b] !== 1) {
      naf[a] = getNAF(coeffs[a], wndWidth[a], this._bitLength);
      naf[b] = getNAF(coeffs[b], wndWidth[b], this._bitLength);
      max = Math.max(naf[a].length, max);
      max = Math.max(naf[b].length, max);
      continue;
    }
    var comb = [
      points[a],
      null,
      null,
      points[b]
    ];
    if (points[a].y.cmp(points[b].y) === 0) {
      comb[1] = points[a].add(points[b]);
      comb[2] = points[a].toJ().mixedAdd(points[b].neg());
    } else if (points[a].y.cmp(points[b].y.redNeg()) === 0) {
      comb[1] = points[a].toJ().mixedAdd(points[b]);
      comb[2] = points[a].add(points[b].neg());
    } else {
      comb[1] = points[a].toJ().mixedAdd(points[b]);
      comb[2] = points[a].toJ().mixedAdd(points[b].neg());
    }
    var index = [
      -3,
      -1,
      -5,
      -7,
      0,
      7,
      5,
      1,
      3
    ];
    var jsf = getJSF(coeffs[a], coeffs[b]);
    max = Math.max(jsf[0].length, max);
    naf[a] = new Array(max);
    naf[b] = new Array(max);
    for (j = 0; j < max; j++) {
      var ja = jsf[0][j] | 0;
      var jb = jsf[1][j] | 0;
      naf[a][j] = index[(ja + 1) * 3 + (jb + 1)];
      naf[b][j] = 0;
      wnd[a] = comb;
    }
  }
  var acc = this.jpoint(null, null, null);
  var tmp = this._wnafT4;
  for (i = max; i >= 0; i--) {
    var k = 0;
    while (i >= 0) {
      var zero = true;
      for (j = 0; j < len; j++) {
        tmp[j] = naf[j][i] | 0;
        if (tmp[j] !== 0)
          zero = false;
      }
      if (!zero)
        break;
      k++;
      i--;
    }
    if (i >= 0)
      k++;
    acc = acc.dblp(k);
    if (i < 0)
      break;
    for (j = 0; j < len; j++) {
      var z = tmp[j];
      if (z === 0)
        continue;
      else if (z > 0)
        p = wnd[j][z - 1 >> 1];
      else if (z < 0)
        p = wnd[j][-z - 1 >> 1].neg();
      if (p.type === "affine")
        acc = acc.mixedAdd(p);
      else
        acc = acc.add(p);
    }
  }
  for (i = 0; i < len; i++)
    wnd[i] = null;
  if (jacobianResult)
    return acc;
  else
    return acc.toP();
};
function BasePoint(curve, type) {
  this.curve = curve;
  this.type = type;
  this.precomputed = null;
}
BaseCurve.BasePoint = BasePoint;
BasePoint.prototype.eq = function eq() {
  throw new Error("Not implemented");
};
BasePoint.prototype.validate = function validate2() {
  return this.curve.validate(this);
};
BaseCurve.prototype.decodePoint = function decodePoint(bytes, enc) {
  bytes = utils_1.toArray(bytes, enc);
  var len = this.p.byteLength();
  if ((bytes[0] === 4 || bytes[0] === 6 || bytes[0] === 7) && bytes.length - 1 === 2 * len) {
    if (bytes[0] === 6)
      assert(bytes[bytes.length - 1] % 2 === 0);
    else if (bytes[0] === 7)
      assert(bytes[bytes.length - 1] % 2 === 1);
    var res = this.point(bytes.slice(1, 1 + len), bytes.slice(1 + len, 1 + 2 * len));
    return res;
  } else if ((bytes[0] === 2 || bytes[0] === 3) && bytes.length - 1 === len) {
    return this.pointFromX(bytes.slice(1, 1 + len), bytes[0] === 3);
  }
  throw new Error("Unknown point format");
};
BasePoint.prototype.encodeCompressed = function encodeCompressed(enc) {
  return this.encode(enc, true);
};
BasePoint.prototype._encode = function _encode(compact) {
  var len = this.curve.p.byteLength();
  var x = this.getX().toArray("be", len);
  if (compact)
    return [this.getY().isEven() ? 2 : 3].concat(x);
  return [4].concat(x, this.getY().toArray("be", len));
};
BasePoint.prototype.encode = function encode(enc, compact) {
  return utils_1.encode(this._encode(compact), enc);
};
BasePoint.prototype.precompute = function precompute(power) {
  if (this.precomputed)
    return this;
  var precomputed = {
    doubles: null,
    naf: null,
    beta: null
  };
  precomputed.naf = this._getNAFPoints(8);
  precomputed.doubles = this._getDoubles(4, power);
  precomputed.beta = this._getBeta();
  this.precomputed = precomputed;
  return this;
};
BasePoint.prototype._hasDoubles = function _hasDoubles(k) {
  if (!this.precomputed)
    return false;
  var doubles = this.precomputed.doubles;
  if (!doubles)
    return false;
  return doubles.points.length >= Math.ceil((k.bitLength() + 1) / doubles.step);
};
BasePoint.prototype._getDoubles = function _getDoubles(step, power) {
  if (this.precomputed && this.precomputed.doubles)
    return this.precomputed.doubles;
  var doubles = [this];
  var acc = this;
  for (var i = 0; i < power; i += step) {
    for (var j = 0; j < step; j++)
      acc = acc.dbl();
    doubles.push(acc);
  }
  return {
    step,
    points: doubles
  };
};
BasePoint.prototype._getNAFPoints = function _getNAFPoints(wnd) {
  if (this.precomputed && this.precomputed.naf)
    return this.precomputed.naf;
  var res = [this];
  var max = (1 << wnd) - 1;
  var dbl5 = max === 1 ? null : this.dbl();
  for (var i = 1; i < max; i++)
    res[i] = res[i - 1].add(dbl5);
  return {
    wnd,
    points: res
  };
};
BasePoint.prototype._getBeta = function _getBeta() {
  return null;
};
BasePoint.prototype.dblp = function dblp(k) {
  var r = this;
  for (var i = 0; i < k; i++)
    r = r.dbl();
  return r;
};
var assert$1 = utils_1.assert;
function ShortCurve(conf) {
  base.call(this, "short", conf);
  this.a = new BN(conf.a, 16).toRed(this.red);
  this.b = new BN(conf.b, 16).toRed(this.red);
  this.tinv = this.two.redInvm();
  this.zeroA = this.a.fromRed().cmpn(0) === 0;
  this.threeA = this.a.fromRed().sub(this.p).cmpn(-3) === 0;
  this.endo = this._getEndomorphism(conf);
  this._endoWnafT1 = new Array(4);
  this._endoWnafT2 = new Array(4);
}
inherits2(ShortCurve, base);
var short_1 = ShortCurve;
ShortCurve.prototype._getEndomorphism = function _getEndomorphism(conf) {
  if (!this.zeroA || !this.g || !this.n || this.p.modn(3) !== 1)
    return;
  var beta;
  var lambda;
  if (conf.beta) {
    beta = new BN(conf.beta, 16).toRed(this.red);
  } else {
    var betas = this._getEndoRoots(this.p);
    beta = betas[0].cmp(betas[1]) < 0 ? betas[0] : betas[1];
    beta = beta.toRed(this.red);
  }
  if (conf.lambda) {
    lambda = new BN(conf.lambda, 16);
  } else {
    var lambdas = this._getEndoRoots(this.n);
    if (this.g.mul(lambdas[0]).x.cmp(this.g.x.redMul(beta)) === 0) {
      lambda = lambdas[0];
    } else {
      lambda = lambdas[1];
      assert$1(this.g.mul(lambda).x.cmp(this.g.x.redMul(beta)) === 0);
    }
  }
  var basis;
  if (conf.basis) {
    basis = conf.basis.map(function(vec) {
      return {
        a: new BN(vec.a, 16),
        b: new BN(vec.b, 16)
      };
    });
  } else {
    basis = this._getEndoBasis(lambda);
  }
  return {
    beta,
    lambda,
    basis
  };
};
ShortCurve.prototype._getEndoRoots = function _getEndoRoots(num) {
  var red = num === this.p ? this.red : BN.mont(num);
  var tinv = new BN(2).toRed(red).redInvm();
  var ntinv = tinv.redNeg();
  var s = new BN(3).toRed(red).redNeg().redSqrt().redMul(tinv);
  var l1 = ntinv.redAdd(s).fromRed();
  var l2 = ntinv.redSub(s).fromRed();
  return [l1, l2];
};
ShortCurve.prototype._getEndoBasis = function _getEndoBasis(lambda) {
  var aprxSqrt = this.n.ushrn(Math.floor(this.n.bitLength() / 2));
  var u = lambda;
  var v = this.n.clone();
  var x1 = new BN(1);
  var y1 = new BN(0);
  var x2 = new BN(0);
  var y2 = new BN(1);
  var a0;
  var b0;
  var a1;
  var b1;
  var a2;
  var b2;
  var prevR;
  var i = 0;
  var r;
  var x;
  while (u.cmpn(0) !== 0) {
    var q = v.div(u);
    r = v.sub(q.mul(u));
    x = x2.sub(q.mul(x1));
    var y = y2.sub(q.mul(y1));
    if (!a1 && r.cmp(aprxSqrt) < 0) {
      a0 = prevR.neg();
      b0 = x1;
      a1 = r.neg();
      b1 = x;
    } else if (a1 && ++i === 2) {
      break;
    }
    prevR = r;
    v = u;
    u = r;
    x2 = x1;
    x1 = x;
    y2 = y1;
    y1 = y;
  }
  a2 = r.neg();
  b2 = x;
  var len1 = a1.sqr().add(b1.sqr());
  var len2 = a2.sqr().add(b2.sqr());
  if (len2.cmp(len1) >= 0) {
    a2 = a0;
    b2 = b0;
  }
  if (a1.negative) {
    a1 = a1.neg();
    b1 = b1.neg();
  }
  if (a2.negative) {
    a2 = a2.neg();
    b2 = b2.neg();
  }
  return [
    {a: a1, b: b1},
    {a: a2, b: b2}
  ];
};
ShortCurve.prototype._endoSplit = function _endoSplit(k) {
  var basis = this.endo.basis;
  var v1 = basis[0];
  var v2 = basis[1];
  var c1 = v2.b.mul(k).divRound(this.n);
  var c2 = v1.b.neg().mul(k).divRound(this.n);
  var p1 = c1.mul(v1.a);
  var p2 = c2.mul(v2.a);
  var q1 = c1.mul(v1.b);
  var q2 = c2.mul(v2.b);
  var k1 = k.sub(p1).sub(p2);
  var k2 = q1.add(q2).neg();
  return {k1, k2};
};
ShortCurve.prototype.pointFromX = function pointFromX(x, odd) {
  x = new BN(x, 16);
  if (!x.red)
    x = x.toRed(this.red);
  var y2 = x.redSqr().redMul(x).redIAdd(x.redMul(this.a)).redIAdd(this.b);
  var y = y2.redSqrt();
  if (y.redSqr().redSub(y2).cmp(this.zero) !== 0)
    throw new Error("invalid point");
  var isOdd = y.fromRed().isOdd();
  if (odd && !isOdd || !odd && isOdd)
    y = y.redNeg();
  return this.point(x, y);
};
ShortCurve.prototype.validate = function validate3(point5) {
  if (point5.inf)
    return true;
  var x = point5.x;
  var y = point5.y;
  var ax = this.a.redMul(x);
  var rhs = x.redSqr().redMul(x).redIAdd(ax).redIAdd(this.b);
  return y.redSqr().redISub(rhs).cmpn(0) === 0;
};
ShortCurve.prototype._endoWnafMulAdd = function _endoWnafMulAdd(points, coeffs, jacobianResult) {
  var npoints = this._endoWnafT1;
  var ncoeffs = this._endoWnafT2;
  for (var i = 0; i < points.length; i++) {
    var split = this._endoSplit(coeffs[i]);
    var p = points[i];
    var beta = p._getBeta();
    if (split.k1.negative) {
      split.k1.ineg();
      p = p.neg(true);
    }
    if (split.k2.negative) {
      split.k2.ineg();
      beta = beta.neg(true);
    }
    npoints[i * 2] = p;
    npoints[i * 2 + 1] = beta;
    ncoeffs[i * 2] = split.k1;
    ncoeffs[i * 2 + 1] = split.k2;
  }
  var res = this._wnafMulAdd(1, npoints, ncoeffs, i * 2, jacobianResult);
  for (var j = 0; j < i * 2; j++) {
    npoints[j] = null;
    ncoeffs[j] = null;
  }
  return res;
};
function Point(curve, x, y, isRed) {
  base.BasePoint.call(this, curve, "affine");
  if (x === null && y === null) {
    this.x = null;
    this.y = null;
    this.inf = true;
  } else {
    this.x = new BN(x, 16);
    this.y = new BN(y, 16);
    if (isRed) {
      this.x.forceRed(this.curve.red);
      this.y.forceRed(this.curve.red);
    }
    if (!this.x.red)
      this.x = this.x.toRed(this.curve.red);
    if (!this.y.red)
      this.y = this.y.toRed(this.curve.red);
    this.inf = false;
  }
}
inherits2(Point, base.BasePoint);
ShortCurve.prototype.point = function point2(x, y, isRed) {
  return new Point(this, x, y, isRed);
};
ShortCurve.prototype.pointFromJSON = function pointFromJSON(obj, red) {
  return Point.fromJSON(this, obj, red);
};
Point.prototype._getBeta = function _getBeta2() {
  if (!this.curve.endo)
    return;
  var pre = this.precomputed;
  if (pre && pre.beta)
    return pre.beta;
  var beta = this.curve.point(this.x.redMul(this.curve.endo.beta), this.y);
  if (pre) {
    var curve = this.curve;
    var endoMul = function(p) {
      return curve.point(p.x.redMul(curve.endo.beta), p.y);
    };
    pre.beta = beta;
    beta.precomputed = {
      beta: null,
      naf: pre.naf && {
        wnd: pre.naf.wnd,
        points: pre.naf.points.map(endoMul)
      },
      doubles: pre.doubles && {
        step: pre.doubles.step,
        points: pre.doubles.points.map(endoMul)
      }
    };
  }
  return beta;
};
Point.prototype.toJSON = function toJSON() {
  if (!this.precomputed)
    return [this.x, this.y];
  return [this.x, this.y, this.precomputed && {
    doubles: this.precomputed.doubles && {
      step: this.precomputed.doubles.step,
      points: this.precomputed.doubles.points.slice(1)
    },
    naf: this.precomputed.naf && {
      wnd: this.precomputed.naf.wnd,
      points: this.precomputed.naf.points.slice(1)
    }
  }];
};
Point.fromJSON = function fromJSON(curve, obj, red) {
  if (typeof obj === "string")
    obj = JSON.parse(obj);
  var res = curve.point(obj[0], obj[1], red);
  if (!obj[2])
    return res;
  function obj2point(obj2) {
    return curve.point(obj2[0], obj2[1], red);
  }
  var pre = obj[2];
  res.precomputed = {
    beta: null,
    doubles: pre.doubles && {
      step: pre.doubles.step,
      points: [res].concat(pre.doubles.points.map(obj2point))
    },
    naf: pre.naf && {
      wnd: pre.naf.wnd,
      points: [res].concat(pre.naf.points.map(obj2point))
    }
  };
  return res;
};
Point.prototype.inspect = function inspect() {
  if (this.isInfinity())
    return "<EC Point Infinity>";
  return "<EC Point x: " + this.x.fromRed().toString(16, 2) + " y: " + this.y.fromRed().toString(16, 2) + ">";
};
Point.prototype.isInfinity = function isInfinity() {
  return this.inf;
};
Point.prototype.add = function add(p) {
  if (this.inf)
    return p;
  if (p.inf)
    return this;
  if (this.eq(p))
    return this.dbl();
  if (this.neg().eq(p))
    return this.curve.point(null, null);
  if (this.x.cmp(p.x) === 0)
    return this.curve.point(null, null);
  var c = this.y.redSub(p.y);
  if (c.cmpn(0) !== 0)
    c = c.redMul(this.x.redSub(p.x).redInvm());
  var nx = c.redSqr().redISub(this.x).redISub(p.x);
  var ny = c.redMul(this.x.redSub(nx)).redISub(this.y);
  return this.curve.point(nx, ny);
};
Point.prototype.dbl = function dbl() {
  if (this.inf)
    return this;
  var ys1 = this.y.redAdd(this.y);
  if (ys1.cmpn(0) === 0)
    return this.curve.point(null, null);
  var a = this.curve.a;
  var x2 = this.x.redSqr();
  var dyinv = ys1.redInvm();
  var c = x2.redAdd(x2).redIAdd(x2).redIAdd(a).redMul(dyinv);
  var nx = c.redSqr().redISub(this.x.redAdd(this.x));
  var ny = c.redMul(this.x.redSub(nx)).redISub(this.y);
  return this.curve.point(nx, ny);
};
Point.prototype.getX = function getX() {
  return this.x.fromRed();
};
Point.prototype.getY = function getY() {
  return this.y.fromRed();
};
Point.prototype.mul = function mul(k) {
  k = new BN(k, 16);
  if (this.isInfinity())
    return this;
  else if (this._hasDoubles(k))
    return this.curve._fixedNafMul(this, k);
  else if (this.curve.endo)
    return this.curve._endoWnafMulAdd([this], [k]);
  else
    return this.curve._wnafMul(this, k);
};
Point.prototype.mulAdd = function mulAdd(k1, p2, k2) {
  var points = [this, p2];
  var coeffs = [k1, k2];
  if (this.curve.endo)
    return this.curve._endoWnafMulAdd(points, coeffs);
  else
    return this.curve._wnafMulAdd(1, points, coeffs, 2);
};
Point.prototype.jmulAdd = function jmulAdd(k1, p2, k2) {
  var points = [this, p2];
  var coeffs = [k1, k2];
  if (this.curve.endo)
    return this.curve._endoWnafMulAdd(points, coeffs, true);
  else
    return this.curve._wnafMulAdd(1, points, coeffs, 2, true);
};
Point.prototype.eq = function eq2(p) {
  return this === p || this.inf === p.inf && (this.inf || this.x.cmp(p.x) === 0 && this.y.cmp(p.y) === 0);
};
Point.prototype.neg = function neg(_precompute) {
  if (this.inf)
    return this;
  var res = this.curve.point(this.x, this.y.redNeg());
  if (_precompute && this.precomputed) {
    var pre = this.precomputed;
    var negate = function(p) {
      return p.neg();
    };
    res.precomputed = {
      naf: pre.naf && {
        wnd: pre.naf.wnd,
        points: pre.naf.points.map(negate)
      },
      doubles: pre.doubles && {
        step: pre.doubles.step,
        points: pre.doubles.points.map(negate)
      }
    };
  }
  return res;
};
Point.prototype.toJ = function toJ() {
  if (this.inf)
    return this.curve.jpoint(null, null, null);
  var res = this.curve.jpoint(this.x, this.y, this.curve.one);
  return res;
};
function JPoint(curve, x, y, z) {
  base.BasePoint.call(this, curve, "jacobian");
  if (x === null && y === null && z === null) {
    this.x = this.curve.one;
    this.y = this.curve.one;
    this.z = new BN(0);
  } else {
    this.x = new BN(x, 16);
    this.y = new BN(y, 16);
    this.z = new BN(z, 16);
  }
  if (!this.x.red)
    this.x = this.x.toRed(this.curve.red);
  if (!this.y.red)
    this.y = this.y.toRed(this.curve.red);
  if (!this.z.red)
    this.z = this.z.toRed(this.curve.red);
  this.zOne = this.z === this.curve.one;
}
inherits2(JPoint, base.BasePoint);
ShortCurve.prototype.jpoint = function jpoint(x, y, z) {
  return new JPoint(this, x, y, z);
};
JPoint.prototype.toP = function toP() {
  if (this.isInfinity())
    return this.curve.point(null, null);
  var zinv = this.z.redInvm();
  var zinv2 = zinv.redSqr();
  var ax = this.x.redMul(zinv2);
  var ay = this.y.redMul(zinv2).redMul(zinv);
  return this.curve.point(ax, ay);
};
JPoint.prototype.neg = function neg2() {
  return this.curve.jpoint(this.x, this.y.redNeg(), this.z);
};
JPoint.prototype.add = function add2(p) {
  if (this.isInfinity())
    return p;
  if (p.isInfinity())
    return this;
  var pz2 = p.z.redSqr();
  var z2 = this.z.redSqr();
  var u1 = this.x.redMul(pz2);
  var u2 = p.x.redMul(z2);
  var s1 = this.y.redMul(pz2.redMul(p.z));
  var s2 = p.y.redMul(z2.redMul(this.z));
  var h = u1.redSub(u2);
  var r = s1.redSub(s2);
  if (h.cmpn(0) === 0) {
    if (r.cmpn(0) !== 0)
      return this.curve.jpoint(null, null, null);
    else
      return this.dbl();
  }
  var h2 = h.redSqr();
  var h3 = h2.redMul(h);
  var v = u1.redMul(h2);
  var nx = r.redSqr().redIAdd(h3).redISub(v).redISub(v);
  var ny = r.redMul(v.redISub(nx)).redISub(s1.redMul(h3));
  var nz = this.z.redMul(p.z).redMul(h);
  return this.curve.jpoint(nx, ny, nz);
};
JPoint.prototype.mixedAdd = function mixedAdd(p) {
  if (this.isInfinity())
    return p.toJ();
  if (p.isInfinity())
    return this;
  var z2 = this.z.redSqr();
  var u1 = this.x;
  var u2 = p.x.redMul(z2);
  var s1 = this.y;
  var s2 = p.y.redMul(z2).redMul(this.z);
  var h = u1.redSub(u2);
  var r = s1.redSub(s2);
  if (h.cmpn(0) === 0) {
    if (r.cmpn(0) !== 0)
      return this.curve.jpoint(null, null, null);
    else
      return this.dbl();
  }
  var h2 = h.redSqr();
  var h3 = h2.redMul(h);
  var v = u1.redMul(h2);
  var nx = r.redSqr().redIAdd(h3).redISub(v).redISub(v);
  var ny = r.redMul(v.redISub(nx)).redISub(s1.redMul(h3));
  var nz = this.z.redMul(h);
  return this.curve.jpoint(nx, ny, nz);
};
JPoint.prototype.dblp = function dblp2(pow) {
  if (pow === 0)
    return this;
  if (this.isInfinity())
    return this;
  if (!pow)
    return this.dbl();
  var i;
  if (this.curve.zeroA || this.curve.threeA) {
    var r = this;
    for (i = 0; i < pow; i++)
      r = r.dbl();
    return r;
  }
  var a = this.curve.a;
  var tinv = this.curve.tinv;
  var jx = this.x;
  var jy = this.y;
  var jz = this.z;
  var jz4 = jz.redSqr().redSqr();
  var jyd = jy.redAdd(jy);
  for (i = 0; i < pow; i++) {
    var jx2 = jx.redSqr();
    var jyd2 = jyd.redSqr();
    var jyd4 = jyd2.redSqr();
    var c = jx2.redAdd(jx2).redIAdd(jx2).redIAdd(a.redMul(jz4));
    var t1 = jx.redMul(jyd2);
    var nx = c.redSqr().redISub(t1.redAdd(t1));
    var t2 = t1.redISub(nx);
    var dny = c.redMul(t2);
    dny = dny.redIAdd(dny).redISub(jyd4);
    var nz = jyd.redMul(jz);
    if (i + 1 < pow)
      jz4 = jz4.redMul(jyd4);
    jx = nx;
    jz = nz;
    jyd = dny;
  }
  return this.curve.jpoint(jx, jyd.redMul(tinv), jz);
};
JPoint.prototype.dbl = function dbl2() {
  if (this.isInfinity())
    return this;
  if (this.curve.zeroA)
    return this._zeroDbl();
  else if (this.curve.threeA)
    return this._threeDbl();
  else
    return this._dbl();
};
JPoint.prototype._zeroDbl = function _zeroDbl() {
  var nx;
  var ny;
  var nz;
  if (this.zOne) {
    var xx = this.x.redSqr();
    var yy = this.y.redSqr();
    var yyyy = yy.redSqr();
    var s = this.x.redAdd(yy).redSqr().redISub(xx).redISub(yyyy);
    s = s.redIAdd(s);
    var m = xx.redAdd(xx).redIAdd(xx);
    var t = m.redSqr().redISub(s).redISub(s);
    var yyyy8 = yyyy.redIAdd(yyyy);
    yyyy8 = yyyy8.redIAdd(yyyy8);
    yyyy8 = yyyy8.redIAdd(yyyy8);
    nx = t;
    ny = m.redMul(s.redISub(t)).redISub(yyyy8);
    nz = this.y.redAdd(this.y);
  } else {
    var a = this.x.redSqr();
    var b = this.y.redSqr();
    var c = b.redSqr();
    var d = this.x.redAdd(b).redSqr().redISub(a).redISub(c);
    d = d.redIAdd(d);
    var e = a.redAdd(a).redIAdd(a);
    var f = e.redSqr();
    var c8 = c.redIAdd(c);
    c8 = c8.redIAdd(c8);
    c8 = c8.redIAdd(c8);
    nx = f.redISub(d).redISub(d);
    ny = e.redMul(d.redISub(nx)).redISub(c8);
    nz = this.y.redMul(this.z);
    nz = nz.redIAdd(nz);
  }
  return this.curve.jpoint(nx, ny, nz);
};
JPoint.prototype._threeDbl = function _threeDbl() {
  var nx;
  var ny;
  var nz;
  if (this.zOne) {
    var xx = this.x.redSqr();
    var yy = this.y.redSqr();
    var yyyy = yy.redSqr();
    var s = this.x.redAdd(yy).redSqr().redISub(xx).redISub(yyyy);
    s = s.redIAdd(s);
    var m = xx.redAdd(xx).redIAdd(xx).redIAdd(this.curve.a);
    var t = m.redSqr().redISub(s).redISub(s);
    nx = t;
    var yyyy8 = yyyy.redIAdd(yyyy);
    yyyy8 = yyyy8.redIAdd(yyyy8);
    yyyy8 = yyyy8.redIAdd(yyyy8);
    ny = m.redMul(s.redISub(t)).redISub(yyyy8);
    nz = this.y.redAdd(this.y);
  } else {
    var delta = this.z.redSqr();
    var gamma = this.y.redSqr();
    var beta = this.x.redMul(gamma);
    var alpha = this.x.redSub(delta).redMul(this.x.redAdd(delta));
    alpha = alpha.redAdd(alpha).redIAdd(alpha);
    var beta4 = beta.redIAdd(beta);
    beta4 = beta4.redIAdd(beta4);
    var beta8 = beta4.redAdd(beta4);
    nx = alpha.redSqr().redISub(beta8);
    nz = this.y.redAdd(this.z).redSqr().redISub(gamma).redISub(delta);
    var ggamma8 = gamma.redSqr();
    ggamma8 = ggamma8.redIAdd(ggamma8);
    ggamma8 = ggamma8.redIAdd(ggamma8);
    ggamma8 = ggamma8.redIAdd(ggamma8);
    ny = alpha.redMul(beta4.redISub(nx)).redISub(ggamma8);
  }
  return this.curve.jpoint(nx, ny, nz);
};
JPoint.prototype._dbl = function _dbl() {
  var a = this.curve.a;
  var jx = this.x;
  var jy = this.y;
  var jz = this.z;
  var jz4 = jz.redSqr().redSqr();
  var jx2 = jx.redSqr();
  var jy2 = jy.redSqr();
  var c = jx2.redAdd(jx2).redIAdd(jx2).redIAdd(a.redMul(jz4));
  var jxd4 = jx.redAdd(jx);
  jxd4 = jxd4.redIAdd(jxd4);
  var t1 = jxd4.redMul(jy2);
  var nx = c.redSqr().redISub(t1.redAdd(t1));
  var t2 = t1.redISub(nx);
  var jyd8 = jy2.redSqr();
  jyd8 = jyd8.redIAdd(jyd8);
  jyd8 = jyd8.redIAdd(jyd8);
  jyd8 = jyd8.redIAdd(jyd8);
  var ny = c.redMul(t2).redISub(jyd8);
  var nz = jy.redAdd(jy).redMul(jz);
  return this.curve.jpoint(nx, ny, nz);
};
JPoint.prototype.trpl = function trpl() {
  if (!this.curve.zeroA)
    return this.dbl().add(this);
  var xx = this.x.redSqr();
  var yy = this.y.redSqr();
  var zz = this.z.redSqr();
  var yyyy = yy.redSqr();
  var m = xx.redAdd(xx).redIAdd(xx);
  var mm = m.redSqr();
  var e = this.x.redAdd(yy).redSqr().redISub(xx).redISub(yyyy);
  e = e.redIAdd(e);
  e = e.redAdd(e).redIAdd(e);
  e = e.redISub(mm);
  var ee = e.redSqr();
  var t = yyyy.redIAdd(yyyy);
  t = t.redIAdd(t);
  t = t.redIAdd(t);
  t = t.redIAdd(t);
  var u = m.redIAdd(e).redSqr().redISub(mm).redISub(ee).redISub(t);
  var yyu4 = yy.redMul(u);
  yyu4 = yyu4.redIAdd(yyu4);
  yyu4 = yyu4.redIAdd(yyu4);
  var nx = this.x.redMul(ee).redISub(yyu4);
  nx = nx.redIAdd(nx);
  nx = nx.redIAdd(nx);
  var ny = this.y.redMul(u.redMul(t.redISub(u)).redISub(e.redMul(ee)));
  ny = ny.redIAdd(ny);
  ny = ny.redIAdd(ny);
  ny = ny.redIAdd(ny);
  var nz = this.z.redAdd(e).redSqr().redISub(zz).redISub(ee);
  return this.curve.jpoint(nx, ny, nz);
};
JPoint.prototype.mul = function mul2(k, kbase) {
  k = new BN(k, kbase);
  return this.curve._wnafMul(this, k);
};
JPoint.prototype.eq = function eq3(p) {
  if (p.type === "affine")
    return this.eq(p.toJ());
  if (this === p)
    return true;
  var z2 = this.z.redSqr();
  var pz2 = p.z.redSqr();
  if (this.x.redMul(pz2).redISub(p.x.redMul(z2)).cmpn(0) !== 0)
    return false;
  var z3 = z2.redMul(this.z);
  var pz3 = pz2.redMul(p.z);
  return this.y.redMul(pz3).redISub(p.y.redMul(z3)).cmpn(0) === 0;
};
JPoint.prototype.eqXToP = function eqXToP(x) {
  var zs = this.z.redSqr();
  var rx = x.toRed(this.curve.red).redMul(zs);
  if (this.x.cmp(rx) === 0)
    return true;
  var xc = x.clone();
  var t = this.curve.redN.redMul(zs);
  for (; ; ) {
    xc.iadd(this.curve.n);
    if (xc.cmp(this.curve.p) >= 0)
      return false;
    rx.redIAdd(t);
    if (this.x.cmp(rx) === 0)
      return true;
  }
};
JPoint.prototype.inspect = function inspect2() {
  if (this.isInfinity())
    return "<EC JPoint Infinity>";
  return "<EC JPoint x: " + this.x.toString(16, 2) + " y: " + this.y.toString(16, 2) + " z: " + this.z.toString(16, 2) + ">";
};
JPoint.prototype.isInfinity = function isInfinity2() {
  return this.z.cmpn(0) === 0;
};
function MontCurve(conf) {
  base.call(this, "mont", conf);
  this.a = new BN(conf.a, 16).toRed(this.red);
  this.b = new BN(conf.b, 16).toRed(this.red);
  this.i4 = new BN(4).toRed(this.red).redInvm();
  this.two = new BN(2).toRed(this.red);
  this.a24 = this.i4.redMul(this.a.redAdd(this.two));
}
inherits2(MontCurve, base);
var mont = MontCurve;
MontCurve.prototype.validate = function validate4(point5) {
  var x = point5.normalize().x;
  var x2 = x.redSqr();
  var rhs = x2.redMul(x).redAdd(x2.redMul(this.a)).redAdd(x);
  var y = rhs.redSqrt();
  return y.redSqr().cmp(rhs) === 0;
};
function Point$1(curve, x, z) {
  base.BasePoint.call(this, curve, "projective");
  if (x === null && z === null) {
    this.x = this.curve.one;
    this.z = this.curve.zero;
  } else {
    this.x = new BN(x, 16);
    this.z = new BN(z, 16);
    if (!this.x.red)
      this.x = this.x.toRed(this.curve.red);
    if (!this.z.red)
      this.z = this.z.toRed(this.curve.red);
  }
}
inherits2(Point$1, base.BasePoint);
MontCurve.prototype.decodePoint = function decodePoint2(bytes, enc) {
  return this.point(utils_1.toArray(bytes, enc), 1);
};
MontCurve.prototype.point = function point3(x, z) {
  return new Point$1(this, x, z);
};
MontCurve.prototype.pointFromJSON = function pointFromJSON2(obj) {
  return Point$1.fromJSON(this, obj);
};
Point$1.prototype.precompute = function precompute2() {
};
Point$1.prototype._encode = function _encode2() {
  return this.getX().toArray("be", this.curve.p.byteLength());
};
Point$1.fromJSON = function fromJSON2(curve, obj) {
  return new Point$1(curve, obj[0], obj[1] || curve.one);
};
Point$1.prototype.inspect = function inspect3() {
  if (this.isInfinity())
    return "<EC Point Infinity>";
  return "<EC Point x: " + this.x.fromRed().toString(16, 2) + " z: " + this.z.fromRed().toString(16, 2) + ">";
};
Point$1.prototype.isInfinity = function isInfinity3() {
  return this.z.cmpn(0) === 0;
};
Point$1.prototype.dbl = function dbl3() {
  var a = this.x.redAdd(this.z);
  var aa = a.redSqr();
  var b = this.x.redSub(this.z);
  var bb = b.redSqr();
  var c = aa.redSub(bb);
  var nx = aa.redMul(bb);
  var nz = c.redMul(bb.redAdd(this.curve.a24.redMul(c)));
  return this.curve.point(nx, nz);
};
Point$1.prototype.add = function add3() {
  throw new Error("Not supported on Montgomery curve");
};
Point$1.prototype.diffAdd = function diffAdd(p, diff) {
  var a = this.x.redAdd(this.z);
  var b = this.x.redSub(this.z);
  var c = p.x.redAdd(p.z);
  var d = p.x.redSub(p.z);
  var da = d.redMul(a);
  var cb = c.redMul(b);
  var nx = diff.z.redMul(da.redAdd(cb).redSqr());
  var nz = diff.x.redMul(da.redISub(cb).redSqr());
  return this.curve.point(nx, nz);
};
Point$1.prototype.mul = function mul3(k) {
  var t = k.clone();
  var a = this;
  var b = this.curve.point(null, null);
  var c = this;
  for (var bits = []; t.cmpn(0) !== 0; t.iushrn(1))
    bits.push(t.andln(1));
  for (var i = bits.length - 1; i >= 0; i--) {
    if (bits[i] === 0) {
      a = a.diffAdd(b, c);
      b = b.dbl();
    } else {
      b = a.diffAdd(b, c);
      a = a.dbl();
    }
  }
  return b;
};
Point$1.prototype.mulAdd = function mulAdd2() {
  throw new Error("Not supported on Montgomery curve");
};
Point$1.prototype.jumlAdd = function jumlAdd() {
  throw new Error("Not supported on Montgomery curve");
};
Point$1.prototype.eq = function eq4(other) {
  return this.getX().cmp(other.getX()) === 0;
};
Point$1.prototype.normalize = function normalize() {
  this.x = this.x.redMul(this.z.redInvm());
  this.z = this.curve.one;
  return this;
};
Point$1.prototype.getX = function getX2() {
  this.normalize();
  return this.x.fromRed();
};
var assert$2 = utils_1.assert;
function EdwardsCurve(conf) {
  this.twisted = (conf.a | 0) !== 1;
  this.mOneA = this.twisted && (conf.a | 0) === -1;
  this.extended = this.mOneA;
  base.call(this, "edwards", conf);
  this.a = new BN(conf.a, 16).umod(this.red.m);
  this.a = this.a.toRed(this.red);
  this.c = new BN(conf.c, 16).toRed(this.red);
  this.c2 = this.c.redSqr();
  this.d = new BN(conf.d, 16).toRed(this.red);
  this.dd = this.d.redAdd(this.d);
  assert$2(!this.twisted || this.c.fromRed().cmpn(1) === 0);
  this.oneC = (conf.c | 0) === 1;
}
inherits2(EdwardsCurve, base);
var edwards = EdwardsCurve;
EdwardsCurve.prototype._mulA = function _mulA(num) {
  if (this.mOneA)
    return num.redNeg();
  else
    return this.a.redMul(num);
};
EdwardsCurve.prototype._mulC = function _mulC(num) {
  if (this.oneC)
    return num;
  else
    return this.c.redMul(num);
};
EdwardsCurve.prototype.jpoint = function jpoint2(x, y, z, t) {
  return this.point(x, y, z, t);
};
EdwardsCurve.prototype.pointFromX = function pointFromX2(x, odd) {
  x = new BN(x, 16);
  if (!x.red)
    x = x.toRed(this.red);
  var x2 = x.redSqr();
  var rhs = this.c2.redSub(this.a.redMul(x2));
  var lhs = this.one.redSub(this.c2.redMul(this.d).redMul(x2));
  var y2 = rhs.redMul(lhs.redInvm());
  var y = y2.redSqrt();
  if (y.redSqr().redSub(y2).cmp(this.zero) !== 0)
    throw new Error("invalid point");
  var isOdd = y.fromRed().isOdd();
  if (odd && !isOdd || !odd && isOdd)
    y = y.redNeg();
  return this.point(x, y);
};
EdwardsCurve.prototype.pointFromY = function pointFromY(y, odd) {
  y = new BN(y, 16);
  if (!y.red)
    y = y.toRed(this.red);
  var y2 = y.redSqr();
  var lhs = y2.redSub(this.c2);
  var rhs = y2.redMul(this.d).redMul(this.c2).redSub(this.a);
  var x2 = lhs.redMul(rhs.redInvm());
  if (x2.cmp(this.zero) === 0) {
    if (odd)
      throw new Error("invalid point");
    else
      return this.point(this.zero, y);
  }
  var x = x2.redSqrt();
  if (x.redSqr().redSub(x2).cmp(this.zero) !== 0)
    throw new Error("invalid point");
  if (x.fromRed().isOdd() !== odd)
    x = x.redNeg();
  return this.point(x, y);
};
EdwardsCurve.prototype.validate = function validate5(point5) {
  if (point5.isInfinity())
    return true;
  point5.normalize();
  var x2 = point5.x.redSqr();
  var y2 = point5.y.redSqr();
  var lhs = x2.redMul(this.a).redAdd(y2);
  var rhs = this.c2.redMul(this.one.redAdd(this.d.redMul(x2).redMul(y2)));
  return lhs.cmp(rhs) === 0;
};
function Point$2(curve, x, y, z, t) {
  base.BasePoint.call(this, curve, "projective");
  if (x === null && y === null && z === null) {
    this.x = this.curve.zero;
    this.y = this.curve.one;
    this.z = this.curve.one;
    this.t = this.curve.zero;
    this.zOne = true;
  } else {
    this.x = new BN(x, 16);
    this.y = new BN(y, 16);
    this.z = z ? new BN(z, 16) : this.curve.one;
    this.t = t && new BN(t, 16);
    if (!this.x.red)
      this.x = this.x.toRed(this.curve.red);
    if (!this.y.red)
      this.y = this.y.toRed(this.curve.red);
    if (!this.z.red)
      this.z = this.z.toRed(this.curve.red);
    if (this.t && !this.t.red)
      this.t = this.t.toRed(this.curve.red);
    this.zOne = this.z === this.curve.one;
    if (this.curve.extended && !this.t) {
      this.t = this.x.redMul(this.y);
      if (!this.zOne)
        this.t = this.t.redMul(this.z.redInvm());
    }
  }
}
inherits2(Point$2, base.BasePoint);
EdwardsCurve.prototype.pointFromJSON = function pointFromJSON3(obj) {
  return Point$2.fromJSON(this, obj);
};
EdwardsCurve.prototype.point = function point4(x, y, z, t) {
  return new Point$2(this, x, y, z, t);
};
Point$2.fromJSON = function fromJSON3(curve, obj) {
  return new Point$2(curve, obj[0], obj[1], obj[2]);
};
Point$2.prototype.inspect = function inspect4() {
  if (this.isInfinity())
    return "<EC Point Infinity>";
  return "<EC Point x: " + this.x.fromRed().toString(16, 2) + " y: " + this.y.fromRed().toString(16, 2) + " z: " + this.z.fromRed().toString(16, 2) + ">";
};
Point$2.prototype.isInfinity = function isInfinity4() {
  return this.x.cmpn(0) === 0 && (this.y.cmp(this.z) === 0 || this.zOne && this.y.cmp(this.curve.c) === 0);
};
Point$2.prototype._extDbl = function _extDbl() {
  var a = this.x.redSqr();
  var b = this.y.redSqr();
  var c = this.z.redSqr();
  c = c.redIAdd(c);
  var d = this.curve._mulA(a);
  var e = this.x.redAdd(this.y).redSqr().redISub(a).redISub(b);
  var g = d.redAdd(b);
  var f = g.redSub(c);
  var h = d.redSub(b);
  var nx = e.redMul(f);
  var ny = g.redMul(h);
  var nt = e.redMul(h);
  var nz = f.redMul(g);
  return this.curve.point(nx, ny, nz, nt);
};
Point$2.prototype._projDbl = function _projDbl() {
  var b = this.x.redAdd(this.y).redSqr();
  var c = this.x.redSqr();
  var d = this.y.redSqr();
  var nx;
  var ny;
  var nz;
  var e;
  var h;
  var j;
  if (this.curve.twisted) {
    e = this.curve._mulA(c);
    var f = e.redAdd(d);
    if (this.zOne) {
      nx = b.redSub(c).redSub(d).redMul(f.redSub(this.curve.two));
      ny = f.redMul(e.redSub(d));
      nz = f.redSqr().redSub(f).redSub(f);
    } else {
      h = this.z.redSqr();
      j = f.redSub(h).redISub(h);
      nx = b.redSub(c).redISub(d).redMul(j);
      ny = f.redMul(e.redSub(d));
      nz = f.redMul(j);
    }
  } else {
    e = c.redAdd(d);
    h = this.curve._mulC(this.z).redSqr();
    j = e.redSub(h).redSub(h);
    nx = this.curve._mulC(b.redISub(e)).redMul(j);
    ny = this.curve._mulC(e).redMul(c.redISub(d));
    nz = e.redMul(j);
  }
  return this.curve.point(nx, ny, nz);
};
Point$2.prototype.dbl = function dbl4() {
  if (this.isInfinity())
    return this;
  if (this.curve.extended)
    return this._extDbl();
  else
    return this._projDbl();
};
Point$2.prototype._extAdd = function _extAdd(p) {
  var a = this.y.redSub(this.x).redMul(p.y.redSub(p.x));
  var b = this.y.redAdd(this.x).redMul(p.y.redAdd(p.x));
  var c = this.t.redMul(this.curve.dd).redMul(p.t);
  var d = this.z.redMul(p.z.redAdd(p.z));
  var e = b.redSub(a);
  var f = d.redSub(c);
  var g = d.redAdd(c);
  var h = b.redAdd(a);
  var nx = e.redMul(f);
  var ny = g.redMul(h);
  var nt = e.redMul(h);
  var nz = f.redMul(g);
  return this.curve.point(nx, ny, nz, nt);
};
Point$2.prototype._projAdd = function _projAdd(p) {
  var a = this.z.redMul(p.z);
  var b = a.redSqr();
  var c = this.x.redMul(p.x);
  var d = this.y.redMul(p.y);
  var e = this.curve.d.redMul(c).redMul(d);
  var f = b.redSub(e);
  var g = b.redAdd(e);
  var tmp = this.x.redAdd(this.y).redMul(p.x.redAdd(p.y)).redISub(c).redISub(d);
  var nx = a.redMul(f).redMul(tmp);
  var ny;
  var nz;
  if (this.curve.twisted) {
    ny = a.redMul(g).redMul(d.redSub(this.curve._mulA(c)));
    nz = f.redMul(g);
  } else {
    ny = a.redMul(g).redMul(d.redSub(c));
    nz = this.curve._mulC(f).redMul(g);
  }
  return this.curve.point(nx, ny, nz);
};
Point$2.prototype.add = function add4(p) {
  if (this.isInfinity())
    return p;
  if (p.isInfinity())
    return this;
  if (this.curve.extended)
    return this._extAdd(p);
  else
    return this._projAdd(p);
};
Point$2.prototype.mul = function mul4(k) {
  if (this._hasDoubles(k))
    return this.curve._fixedNafMul(this, k);
  else
    return this.curve._wnafMul(this, k);
};
Point$2.prototype.mulAdd = function mulAdd3(k1, p, k2) {
  return this.curve._wnafMulAdd(1, [this, p], [k1, k2], 2, false);
};
Point$2.prototype.jmulAdd = function jmulAdd2(k1, p, k2) {
  return this.curve._wnafMulAdd(1, [this, p], [k1, k2], 2, true);
};
Point$2.prototype.normalize = function normalize2() {
  if (this.zOne)
    return this;
  var zi = this.z.redInvm();
  this.x = this.x.redMul(zi);
  this.y = this.y.redMul(zi);
  if (this.t)
    this.t = this.t.redMul(zi);
  this.z = this.curve.one;
  this.zOne = true;
  return this;
};
Point$2.prototype.neg = function neg3() {
  return this.curve.point(this.x.redNeg(), this.y, this.z, this.t && this.t.redNeg());
};
Point$2.prototype.getX = function getX3() {
  this.normalize();
  return this.x.fromRed();
};
Point$2.prototype.getY = function getY2() {
  this.normalize();
  return this.y.fromRed();
};
Point$2.prototype.eq = function eq5(other) {
  return this === other || this.getX().cmp(other.getX()) === 0 && this.getY().cmp(other.getY()) === 0;
};
Point$2.prototype.eqXToP = function eqXToP2(x) {
  var rx = x.toRed(this.curve.red).redMul(this.z);
  if (this.x.cmp(rx) === 0)
    return true;
  var xc = x.clone();
  var t = this.curve.redN.redMul(this.z);
  for (; ; ) {
    xc.iadd(this.curve.n);
    if (xc.cmp(this.curve.p) >= 0)
      return false;
    rx.redIAdd(t);
    if (this.x.cmp(rx) === 0)
      return true;
  }
};
Point$2.prototype.toP = Point$2.prototype.normalize;
Point$2.prototype.mixedAdd = Point$2.prototype.add;
var curve_1 = createCommonjsModule(function(module, exports) {
  var curve = exports;
  curve.base = base;
  curve.short = short_1;
  curve.mont = mont;
  curve.edwards = edwards;
});
var secp256k1 = {
  doubles: {
    step: 4,
    points: [
      [
        "e60fce93b59e9ec53011aabc21c23e97b2a31369b87a5ae9c44ee89e2a6dec0a",
        "f7e3507399e595929db99f34f57937101296891e44d23f0be1f32cce69616821"
      ],
      [
        "8282263212c609d9ea2a6e3e172de238d8c39cabd5ac1ca10646e23fd5f51508",
        "11f8a8098557dfe45e8256e830b60ace62d613ac2f7b17bed31b6eaff6e26caf"
      ],
      [
        "175e159f728b865a72f99cc6c6fc846de0b93833fd2222ed73fce5b551e5b739",
        "d3506e0d9e3c79eba4ef97a51ff71f5eacb5955add24345c6efa6ffee9fed695"
      ],
      [
        "363d90d447b00c9c99ceac05b6262ee053441c7e55552ffe526bad8f83ff4640",
        "4e273adfc732221953b445397f3363145b9a89008199ecb62003c7f3bee9de9"
      ],
      [
        "8b4b5f165df3c2be8c6244b5b745638843e4a781a15bcd1b69f79a55dffdf80c",
        "4aad0a6f68d308b4b3fbd7813ab0da04f9e336546162ee56b3eff0c65fd4fd36"
      ],
      [
        "723cbaa6e5db996d6bf771c00bd548c7b700dbffa6c0e77bcb6115925232fcda",
        "96e867b5595cc498a921137488824d6e2660a0653779494801dc069d9eb39f5f"
      ],
      [
        "eebfa4d493bebf98ba5feec812c2d3b50947961237a919839a533eca0e7dd7fa",
        "5d9a8ca3970ef0f269ee7edaf178089d9ae4cdc3a711f712ddfd4fdae1de8999"
      ],
      [
        "100f44da696e71672791d0a09b7bde459f1215a29b3c03bfefd7835b39a48db0",
        "cdd9e13192a00b772ec8f3300c090666b7ff4a18ff5195ac0fbd5cd62bc65a09"
      ],
      [
        "e1031be262c7ed1b1dc9227a4a04c017a77f8d4464f3b3852c8acde6e534fd2d",
        "9d7061928940405e6bb6a4176597535af292dd419e1ced79a44f18f29456a00d"
      ],
      [
        "feea6cae46d55b530ac2839f143bd7ec5cf8b266a41d6af52d5e688d9094696d",
        "e57c6b6c97dce1bab06e4e12bf3ecd5c981c8957cc41442d3155debf18090088"
      ],
      [
        "da67a91d91049cdcb367be4be6ffca3cfeed657d808583de33fa978bc1ec6cb1",
        "9bacaa35481642bc41f463f7ec9780e5dec7adc508f740a17e9ea8e27a68be1d"
      ],
      [
        "53904faa0b334cdda6e000935ef22151ec08d0f7bb11069f57545ccc1a37b7c0",
        "5bc087d0bc80106d88c9eccac20d3c1c13999981e14434699dcb096b022771c8"
      ],
      [
        "8e7bcd0bd35983a7719cca7764ca906779b53a043a9b8bcaeff959f43ad86047",
        "10b7770b2a3da4b3940310420ca9514579e88e2e47fd68b3ea10047e8460372a"
      ],
      [
        "385eed34c1cdff21e6d0818689b81bde71a7f4f18397e6690a841e1599c43862",
        "283bebc3e8ea23f56701de19e9ebf4576b304eec2086dc8cc0458fe5542e5453"
      ],
      [
        "6f9d9b803ecf191637c73a4413dfa180fddf84a5947fbc9c606ed86c3fac3a7",
        "7c80c68e603059ba69b8e2a30e45c4d47ea4dd2f5c281002d86890603a842160"
      ],
      [
        "3322d401243c4e2582a2147c104d6ecbf774d163db0f5e5313b7e0e742d0e6bd",
        "56e70797e9664ef5bfb019bc4ddaf9b72805f63ea2873af624f3a2e96c28b2a0"
      ],
      [
        "85672c7d2de0b7da2bd1770d89665868741b3f9af7643397721d74d28134ab83",
        "7c481b9b5b43b2eb6374049bfa62c2e5e77f17fcc5298f44c8e3094f790313a6"
      ],
      [
        "948bf809b1988a46b06c9f1919413b10f9226c60f668832ffd959af60c82a0a",
        "53a562856dcb6646dc6b74c5d1c3418c6d4dff08c97cd2bed4cb7f88d8c8e589"
      ],
      [
        "6260ce7f461801c34f067ce0f02873a8f1b0e44dfc69752accecd819f38fd8e8",
        "bc2da82b6fa5b571a7f09049776a1ef7ecd292238051c198c1a84e95b2b4ae17"
      ],
      [
        "e5037de0afc1d8d43d8348414bbf4103043ec8f575bfdc432953cc8d2037fa2d",
        "4571534baa94d3b5f9f98d09fb990bddbd5f5b03ec481f10e0e5dc841d755bda"
      ],
      [
        "e06372b0f4a207adf5ea905e8f1771b4e7e8dbd1c6a6c5b725866a0ae4fce725",
        "7a908974bce18cfe12a27bb2ad5a488cd7484a7787104870b27034f94eee31dd"
      ],
      [
        "213c7a715cd5d45358d0bbf9dc0ce02204b10bdde2a3f58540ad6908d0559754",
        "4b6dad0b5ae462507013ad06245ba190bb4850f5f36a7eeddff2c27534b458f2"
      ],
      [
        "4e7c272a7af4b34e8dbb9352a5419a87e2838c70adc62cddf0cc3a3b08fbd53c",
        "17749c766c9d0b18e16fd09f6def681b530b9614bff7dd33e0b3941817dcaae6"
      ],
      [
        "fea74e3dbe778b1b10f238ad61686aa5c76e3db2be43057632427e2840fb27b6",
        "6e0568db9b0b13297cf674deccb6af93126b596b973f7b77701d3db7f23cb96f"
      ],
      [
        "76e64113f677cf0e10a2570d599968d31544e179b760432952c02a4417bdde39",
        "c90ddf8dee4e95cf577066d70681f0d35e2a33d2b56d2032b4b1752d1901ac01"
      ],
      [
        "c738c56b03b2abe1e8281baa743f8f9a8f7cc643df26cbee3ab150242bcbb891",
        "893fb578951ad2537f718f2eacbfbbbb82314eef7880cfe917e735d9699a84c3"
      ],
      [
        "d895626548b65b81e264c7637c972877d1d72e5f3a925014372e9f6588f6c14b",
        "febfaa38f2bc7eae728ec60818c340eb03428d632bb067e179363ed75d7d991f"
      ],
      [
        "b8da94032a957518eb0f6433571e8761ceffc73693e84edd49150a564f676e03",
        "2804dfa44805a1e4d7c99cc9762808b092cc584d95ff3b511488e4e74efdf6e7"
      ],
      [
        "e80fea14441fb33a7d8adab9475d7fab2019effb5156a792f1a11778e3c0df5d",
        "eed1de7f638e00771e89768ca3ca94472d155e80af322ea9fcb4291b6ac9ec78"
      ],
      [
        "a301697bdfcd704313ba48e51d567543f2a182031efd6915ddc07bbcc4e16070",
        "7370f91cfb67e4f5081809fa25d40f9b1735dbf7c0a11a130c0d1a041e177ea1"
      ],
      [
        "90ad85b389d6b936463f9d0512678de208cc330b11307fffab7ac63e3fb04ed4",
        "e507a3620a38261affdcbd9427222b839aefabe1582894d991d4d48cb6ef150"
      ],
      [
        "8f68b9d2f63b5f339239c1ad981f162ee88c5678723ea3351b7b444c9ec4c0da",
        "662a9f2dba063986de1d90c2b6be215dbbea2cfe95510bfdf23cbf79501fff82"
      ],
      [
        "e4f3fb0176af85d65ff99ff9198c36091f48e86503681e3e6686fd5053231e11",
        "1e63633ad0ef4f1c1661a6d0ea02b7286cc7e74ec951d1c9822c38576feb73bc"
      ],
      [
        "8c00fa9b18ebf331eb961537a45a4266c7034f2f0d4e1d0716fb6eae20eae29e",
        "efa47267fea521a1a9dc343a3736c974c2fadafa81e36c54e7d2a4c66702414b"
      ],
      [
        "e7a26ce69dd4829f3e10cec0a9e98ed3143d084f308b92c0997fddfc60cb3e41",
        "2a758e300fa7984b471b006a1aafbb18d0a6b2c0420e83e20e8a9421cf2cfd51"
      ],
      [
        "b6459e0ee3662ec8d23540c223bcbdc571cbcb967d79424f3cf29eb3de6b80ef",
        "67c876d06f3e06de1dadf16e5661db3c4b3ae6d48e35b2ff30bf0b61a71ba45"
      ],
      [
        "d68a80c8280bb840793234aa118f06231d6f1fc67e73c5a5deda0f5b496943e8",
        "db8ba9fff4b586d00c4b1f9177b0e28b5b0e7b8f7845295a294c84266b133120"
      ],
      [
        "324aed7df65c804252dc0270907a30b09612aeb973449cea4095980fc28d3d5d",
        "648a365774b61f2ff130c0c35aec1f4f19213b0c7e332843967224af96ab7c84"
      ],
      [
        "4df9c14919cde61f6d51dfdbe5fee5dceec4143ba8d1ca888e8bd373fd054c96",
        "35ec51092d8728050974c23a1d85d4b5d506cdc288490192ebac06cad10d5d"
      ],
      [
        "9c3919a84a474870faed8a9c1cc66021523489054d7f0308cbfc99c8ac1f98cd",
        "ddb84f0f4a4ddd57584f044bf260e641905326f76c64c8e6be7e5e03d4fc599d"
      ],
      [
        "6057170b1dd12fdf8de05f281d8e06bb91e1493a8b91d4cc5a21382120a959e5",
        "9a1af0b26a6a4807add9a2daf71df262465152bc3ee24c65e899be932385a2a8"
      ],
      [
        "a576df8e23a08411421439a4518da31880cef0fba7d4df12b1a6973eecb94266",
        "40a6bf20e76640b2c92b97afe58cd82c432e10a7f514d9f3ee8be11ae1b28ec8"
      ],
      [
        "7778a78c28dec3e30a05fe9629de8c38bb30d1f5cf9a3a208f763889be58ad71",
        "34626d9ab5a5b22ff7098e12f2ff580087b38411ff24ac563b513fc1fd9f43ac"
      ],
      [
        "928955ee637a84463729fd30e7afd2ed5f96274e5ad7e5cb09eda9c06d903ac",
        "c25621003d3f42a827b78a13093a95eeac3d26efa8a8d83fc5180e935bcd091f"
      ],
      [
        "85d0fef3ec6db109399064f3a0e3b2855645b4a907ad354527aae75163d82751",
        "1f03648413a38c0be29d496e582cf5663e8751e96877331582c237a24eb1f962"
      ],
      [
        "ff2b0dce97eece97c1c9b6041798b85dfdfb6d8882da20308f5404824526087e",
        "493d13fef524ba188af4c4dc54d07936c7b7ed6fb90e2ceb2c951e01f0c29907"
      ],
      [
        "827fbbe4b1e880ea9ed2b2e6301b212b57f1ee148cd6dd28780e5e2cf856e241",
        "c60f9c923c727b0b71bef2c67d1d12687ff7a63186903166d605b68baec293ec"
      ],
      [
        "eaa649f21f51bdbae7be4ae34ce6e5217a58fdce7f47f9aa7f3b58fa2120e2b3",
        "be3279ed5bbbb03ac69a80f89879aa5a01a6b965f13f7e59d47a5305ba5ad93d"
      ],
      [
        "e4a42d43c5cf169d9391df6decf42ee541b6d8f0c9a137401e23632dda34d24f",
        "4d9f92e716d1c73526fc99ccfb8ad34ce886eedfa8d8e4f13a7f7131deba9414"
      ],
      [
        "1ec80fef360cbdd954160fadab352b6b92b53576a88fea4947173b9d4300bf19",
        "aeefe93756b5340d2f3a4958a7abbf5e0146e77f6295a07b671cdc1cc107cefd"
      ],
      [
        "146a778c04670c2f91b00af4680dfa8bce3490717d58ba889ddb5928366642be",
        "b318e0ec3354028add669827f9d4b2870aaa971d2f7e5ed1d0b297483d83efd0"
      ],
      [
        "fa50c0f61d22e5f07e3acebb1aa07b128d0012209a28b9776d76a8793180eef9",
        "6b84c6922397eba9b72cd2872281a68a5e683293a57a213b38cd8d7d3f4f2811"
      ],
      [
        "da1d61d0ca721a11b1a5bf6b7d88e8421a288ab5d5bba5220e53d32b5f067ec2",
        "8157f55a7c99306c79c0766161c91e2966a73899d279b48a655fba0f1ad836f1"
      ],
      [
        "a8e282ff0c9706907215ff98e8fd416615311de0446f1e062a73b0610d064e13",
        "7f97355b8db81c09abfb7f3c5b2515888b679a3e50dd6bd6cef7c73111f4cc0c"
      ],
      [
        "174a53b9c9a285872d39e56e6913cab15d59b1fa512508c022f382de8319497c",
        "ccc9dc37abfc9c1657b4155f2c47f9e6646b3a1d8cb9854383da13ac079afa73"
      ],
      [
        "959396981943785c3d3e57edf5018cdbe039e730e4918b3d884fdff09475b7ba",
        "2e7e552888c331dd8ba0386a4b9cd6849c653f64c8709385e9b8abf87524f2fd"
      ],
      [
        "d2a63a50ae401e56d645a1153b109a8fcca0a43d561fba2dbb51340c9d82b151",
        "e82d86fb6443fcb7565aee58b2948220a70f750af484ca52d4142174dcf89405"
      ],
      [
        "64587e2335471eb890ee7896d7cfdc866bacbdbd3839317b3436f9b45617e073",
        "d99fcdd5bf6902e2ae96dd6447c299a185b90a39133aeab358299e5e9faf6589"
      ],
      [
        "8481bde0e4e4d885b3a546d3e549de042f0aa6cea250e7fd358d6c86dd45e458",
        "38ee7b8cba5404dd84a25bf39cecb2ca900a79c42b262e556d64b1b59779057e"
      ],
      [
        "13464a57a78102aa62b6979ae817f4637ffcfed3c4b1ce30bcd6303f6caf666b",
        "69be159004614580ef7e433453ccb0ca48f300a81d0942e13f495a907f6ecc27"
      ],
      [
        "bc4a9df5b713fe2e9aef430bcc1dc97a0cd9ccede2f28588cada3a0d2d83f366",
        "d3a81ca6e785c06383937adf4b798caa6e8a9fbfa547b16d758d666581f33c1"
      ],
      [
        "8c28a97bf8298bc0d23d8c749452a32e694b65e30a9472a3954ab30fe5324caa",
        "40a30463a3305193378fedf31f7cc0eb7ae784f0451cb9459e71dc73cbef9482"
      ],
      [
        "8ea9666139527a8c1dd94ce4f071fd23c8b350c5a4bb33748c4ba111faccae0",
        "620efabbc8ee2782e24e7c0cfb95c5d735b783be9cf0f8e955af34a30e62b945"
      ],
      [
        "dd3625faef5ba06074669716bbd3788d89bdde815959968092f76cc4eb9a9787",
        "7a188fa3520e30d461da2501045731ca941461982883395937f68d00c644a573"
      ],
      [
        "f710d79d9eb962297e4f6232b40e8f7feb2bc63814614d692c12de752408221e",
        "ea98e67232d3b3295d3b535532115ccac8612c721851617526ae47a9c77bfc82"
      ]
    ]
  },
  naf: {
    wnd: 7,
    points: [
      [
        "f9308a019258c31049344f85f89d5229b531c845836f99b08601f113bce036f9",
        "388f7b0f632de8140fe337e62a37f3566500a99934c2231b6cb9fd7584b8e672"
      ],
      [
        "2f8bde4d1a07209355b4a7250a5c5128e88b84bddc619ab7cba8d569b240efe4",
        "d8ac222636e5e3d6d4dba9dda6c9c426f788271bab0d6840dca87d3aa6ac62d6"
      ],
      [
        "5cbdf0646e5db4eaa398f365f2ea7a0e3d419b7e0330e39ce92bddedcac4f9bc",
        "6aebca40ba255960a3178d6d861a54dba813d0b813fde7b5a5082628087264da"
      ],
      [
        "acd484e2f0c7f65309ad178a9f559abde09796974c57e714c35f110dfc27ccbe",
        "cc338921b0a7d9fd64380971763b61e9add888a4375f8e0f05cc262ac64f9c37"
      ],
      [
        "774ae7f858a9411e5ef4246b70c65aac5649980be5c17891bbec17895da008cb",
        "d984a032eb6b5e190243dd56d7b7b365372db1e2dff9d6a8301d74c9c953c61b"
      ],
      [
        "f28773c2d975288bc7d1d205c3748651b075fbc6610e58cddeeddf8f19405aa8",
        "ab0902e8d880a89758212eb65cdaf473a1a06da521fa91f29b5cb52db03ed81"
      ],
      [
        "d7924d4f7d43ea965a465ae3095ff41131e5946f3c85f79e44adbcf8e27e080e",
        "581e2872a86c72a683842ec228cc6defea40af2bd896d3a5c504dc9ff6a26b58"
      ],
      [
        "defdea4cdb677750a420fee807eacf21eb9898ae79b9768766e4faa04a2d4a34",
        "4211ab0694635168e997b0ead2a93daeced1f4a04a95c0f6cfb199f69e56eb77"
      ],
      [
        "2b4ea0a797a443d293ef5cff444f4979f06acfebd7e86d277475656138385b6c",
        "85e89bc037945d93b343083b5a1c86131a01f60c50269763b570c854e5c09b7a"
      ],
      [
        "352bbf4a4cdd12564f93fa332ce333301d9ad40271f8107181340aef25be59d5",
        "321eb4075348f534d59c18259dda3e1f4a1b3b2e71b1039c67bd3d8bcf81998c"
      ],
      [
        "2fa2104d6b38d11b0230010559879124e42ab8dfeff5ff29dc9cdadd4ecacc3f",
        "2de1068295dd865b64569335bd5dd80181d70ecfc882648423ba76b532b7d67"
      ],
      [
        "9248279b09b4d68dab21a9b066edda83263c3d84e09572e269ca0cd7f5453714",
        "73016f7bf234aade5d1aa71bdea2b1ff3fc0de2a887912ffe54a32ce97cb3402"
      ],
      [
        "daed4f2be3a8bf278e70132fb0beb7522f570e144bf615c07e996d443dee8729",
        "a69dce4a7d6c98e8d4a1aca87ef8d7003f83c230f3afa726ab40e52290be1c55"
      ],
      [
        "c44d12c7065d812e8acf28d7cbb19f9011ecd9e9fdf281b0e6a3b5e87d22e7db",
        "2119a460ce326cdc76c45926c982fdac0e106e861edf61c5a039063f0e0e6482"
      ],
      [
        "6a245bf6dc698504c89a20cfded60853152b695336c28063b61c65cbd269e6b4",
        "e022cf42c2bd4a708b3f5126f16a24ad8b33ba48d0423b6efd5e6348100d8a82"
      ],
      [
        "1697ffa6fd9de627c077e3d2fe541084ce13300b0bec1146f95ae57f0d0bd6a5",
        "b9c398f186806f5d27561506e4557433a2cf15009e498ae7adee9d63d01b2396"
      ],
      [
        "605bdb019981718b986d0f07e834cb0d9deb8360ffb7f61df982345ef27a7479",
        "2972d2de4f8d20681a78d93ec96fe23c26bfae84fb14db43b01e1e9056b8c49"
      ],
      [
        "62d14dab4150bf497402fdc45a215e10dcb01c354959b10cfe31c7e9d87ff33d",
        "80fc06bd8cc5b01098088a1950eed0db01aa132967ab472235f5642483b25eaf"
      ],
      [
        "80c60ad0040f27dade5b4b06c408e56b2c50e9f56b9b8b425e555c2f86308b6f",
        "1c38303f1cc5c30f26e66bad7fe72f70a65eed4cbe7024eb1aa01f56430bd57a"
      ],
      [
        "7a9375ad6167ad54aa74c6348cc54d344cc5dc9487d847049d5eabb0fa03c8fb",
        "d0e3fa9eca8726909559e0d79269046bdc59ea10c70ce2b02d499ec224dc7f7"
      ],
      [
        "d528ecd9b696b54c907a9ed045447a79bb408ec39b68df504bb51f459bc3ffc9",
        "eecf41253136e5f99966f21881fd656ebc4345405c520dbc063465b521409933"
      ],
      [
        "49370a4b5f43412ea25f514e8ecdad05266115e4a7ecb1387231808f8b45963",
        "758f3f41afd6ed428b3081b0512fd62a54c3f3afbb5b6764b653052a12949c9a"
      ],
      [
        "77f230936ee88cbbd73df930d64702ef881d811e0e1498e2f1c13eb1fc345d74",
        "958ef42a7886b6400a08266e9ba1b37896c95330d97077cbbe8eb3c7671c60d6"
      ],
      [
        "f2dac991cc4ce4b9ea44887e5c7c0bce58c80074ab9d4dbaeb28531b7739f530",
        "e0dedc9b3b2f8dad4da1f32dec2531df9eb5fbeb0598e4fd1a117dba703a3c37"
      ],
      [
        "463b3d9f662621fb1b4be8fbbe2520125a216cdfc9dae3debcba4850c690d45b",
        "5ed430d78c296c3543114306dd8622d7c622e27c970a1de31cb377b01af7307e"
      ],
      [
        "f16f804244e46e2a09232d4aff3b59976b98fac14328a2d1a32496b49998f247",
        "cedabd9b82203f7e13d206fcdf4e33d92a6c53c26e5cce26d6579962c4e31df6"
      ],
      [
        "caf754272dc84563b0352b7a14311af55d245315ace27c65369e15f7151d41d1",
        "cb474660ef35f5f2a41b643fa5e460575f4fa9b7962232a5c32f908318a04476"
      ],
      [
        "2600ca4b282cb986f85d0f1709979d8b44a09c07cb86d7c124497bc86f082120",
        "4119b88753c15bd6a693b03fcddbb45d5ac6be74ab5f0ef44b0be9475a7e4b40"
      ],
      [
        "7635ca72d7e8432c338ec53cd12220bc01c48685e24f7dc8c602a7746998e435",
        "91b649609489d613d1d5e590f78e6d74ecfc061d57048bad9e76f302c5b9c61"
      ],
      [
        "754e3239f325570cdbbf4a87deee8a66b7f2b33479d468fbc1a50743bf56cc18",
        "673fb86e5bda30fb3cd0ed304ea49a023ee33d0197a695d0c5d98093c536683"
      ],
      [
        "e3e6bd1071a1e96aff57859c82d570f0330800661d1c952f9fe2694691d9b9e8",
        "59c9e0bba394e76f40c0aa58379a3cb6a5a2283993e90c4167002af4920e37f5"
      ],
      [
        "186b483d056a033826ae73d88f732985c4ccb1f32ba35f4b4cc47fdcf04aa6eb",
        "3b952d32c67cf77e2e17446e204180ab21fb8090895138b4a4a797f86e80888b"
      ],
      [
        "df9d70a6b9876ce544c98561f4be4f725442e6d2b737d9c91a8321724ce0963f",
        "55eb2dafd84d6ccd5f862b785dc39d4ab157222720ef9da217b8c45cf2ba2417"
      ],
      [
        "5edd5cc23c51e87a497ca815d5dce0f8ab52554f849ed8995de64c5f34ce7143",
        "efae9c8dbc14130661e8cec030c89ad0c13c66c0d17a2905cdc706ab7399a868"
      ],
      [
        "290798c2b6476830da12fe02287e9e777aa3fba1c355b17a722d362f84614fba",
        "e38da76dcd440621988d00bcf79af25d5b29c094db2a23146d003afd41943e7a"
      ],
      [
        "af3c423a95d9f5b3054754efa150ac39cd29552fe360257362dfdecef4053b45",
        "f98a3fd831eb2b749a93b0e6f35cfb40c8cd5aa667a15581bc2feded498fd9c6"
      ],
      [
        "766dbb24d134e745cccaa28c99bf274906bb66b26dcf98df8d2fed50d884249a",
        "744b1152eacbe5e38dcc887980da38b897584a65fa06cedd2c924f97cbac5996"
      ],
      [
        "59dbf46f8c94759ba21277c33784f41645f7b44f6c596a58ce92e666191abe3e",
        "c534ad44175fbc300f4ea6ce648309a042ce739a7919798cd85e216c4a307f6e"
      ],
      [
        "f13ada95103c4537305e691e74e9a4a8dd647e711a95e73cb62dc6018cfd87b8",
        "e13817b44ee14de663bf4bc808341f326949e21a6a75c2570778419bdaf5733d"
      ],
      [
        "7754b4fa0e8aced06d4167a2c59cca4cda1869c06ebadfb6488550015a88522c",
        "30e93e864e669d82224b967c3020b8fa8d1e4e350b6cbcc537a48b57841163a2"
      ],
      [
        "948dcadf5990e048aa3874d46abef9d701858f95de8041d2a6828c99e2262519",
        "e491a42537f6e597d5d28a3224b1bc25df9154efbd2ef1d2cbba2cae5347d57e"
      ],
      [
        "7962414450c76c1689c7b48f8202ec37fb224cf5ac0bfa1570328a8a3d7c77ab",
        "100b610ec4ffb4760d5c1fc133ef6f6b12507a051f04ac5760afa5b29db83437"
      ],
      [
        "3514087834964b54b15b160644d915485a16977225b8847bb0dd085137ec47ca",
        "ef0afbb2056205448e1652c48e8127fc6039e77c15c2378b7e7d15a0de293311"
      ],
      [
        "d3cc30ad6b483e4bc79ce2c9dd8bc54993e947eb8df787b442943d3f7b527eaf",
        "8b378a22d827278d89c5e9be8f9508ae3c2ad46290358630afb34db04eede0a4"
      ],
      [
        "1624d84780732860ce1c78fcbfefe08b2b29823db913f6493975ba0ff4847610",
        "68651cf9b6da903e0914448c6cd9d4ca896878f5282be4c8cc06e2a404078575"
      ],
      [
        "733ce80da955a8a26902c95633e62a985192474b5af207da6df7b4fd5fc61cd4",
        "f5435a2bd2badf7d485a4d8b8db9fcce3e1ef8e0201e4578c54673bc1dc5ea1d"
      ],
      [
        "15d9441254945064cf1a1c33bbd3b49f8966c5092171e699ef258dfab81c045c",
        "d56eb30b69463e7234f5137b73b84177434800bacebfc685fc37bbe9efe4070d"
      ],
      [
        "a1d0fcf2ec9de675b612136e5ce70d271c21417c9d2b8aaaac138599d0717940",
        "edd77f50bcb5a3cab2e90737309667f2641462a54070f3d519212d39c197a629"
      ],
      [
        "e22fbe15c0af8ccc5780c0735f84dbe9a790badee8245c06c7ca37331cb36980",
        "a855babad5cd60c88b430a69f53a1a7a38289154964799be43d06d77d31da06"
      ],
      [
        "311091dd9860e8e20ee13473c1155f5f69635e394704eaa74009452246cfa9b3",
        "66db656f87d1f04fffd1f04788c06830871ec5a64feee685bd80f0b1286d8374"
      ],
      [
        "34c1fd04d301be89b31c0442d3e6ac24883928b45a9340781867d4232ec2dbdf",
        "9414685e97b1b5954bd46f730174136d57f1ceeb487443dc5321857ba73abee"
      ],
      [
        "f219ea5d6b54701c1c14de5b557eb42a8d13f3abbcd08affcc2a5e6b049b8d63",
        "4cb95957e83d40b0f73af4544cccf6b1f4b08d3c07b27fb8d8c2962a400766d1"
      ],
      [
        "d7b8740f74a8fbaab1f683db8f45de26543a5490bca627087236912469a0b448",
        "fa77968128d9c92ee1010f337ad4717eff15db5ed3c049b3411e0315eaa4593b"
      ],
      [
        "32d31c222f8f6f0ef86f7c98d3a3335ead5bcd32abdd94289fe4d3091aa824bf",
        "5f3032f5892156e39ccd3d7915b9e1da2e6dac9e6f26e961118d14b8462e1661"
      ],
      [
        "7461f371914ab32671045a155d9831ea8793d77cd59592c4340f86cbc18347b5",
        "8ec0ba238b96bec0cbdddcae0aa442542eee1ff50c986ea6b39847b3cc092ff6"
      ],
      [
        "ee079adb1df1860074356a25aa38206a6d716b2c3e67453d287698bad7b2b2d6",
        "8dc2412aafe3be5c4c5f37e0ecc5f9f6a446989af04c4e25ebaac479ec1c8c1e"
      ],
      [
        "16ec93e447ec83f0467b18302ee620f7e65de331874c9dc72bfd8616ba9da6b5",
        "5e4631150e62fb40d0e8c2a7ca5804a39d58186a50e497139626778e25b0674d"
      ],
      [
        "eaa5f980c245f6f038978290afa70b6bd8855897f98b6aa485b96065d537bd99",
        "f65f5d3e292c2e0819a528391c994624d784869d7e6ea67fb18041024edc07dc"
      ],
      [
        "78c9407544ac132692ee1910a02439958ae04877151342ea96c4b6b35a49f51",
        "f3e0319169eb9b85d5404795539a5e68fa1fbd583c064d2462b675f194a3ddb4"
      ],
      [
        "494f4be219a1a77016dcd838431aea0001cdc8ae7a6fc688726578d9702857a5",
        "42242a969283a5f339ba7f075e36ba2af925ce30d767ed6e55f4b031880d562c"
      ],
      [
        "a598a8030da6d86c6bc7f2f5144ea549d28211ea58faa70ebf4c1e665c1fe9b5",
        "204b5d6f84822c307e4b4a7140737aec23fc63b65b35f86a10026dbd2d864e6b"
      ],
      [
        "c41916365abb2b5d09192f5f2dbeafec208f020f12570a184dbadc3e58595997",
        "4f14351d0087efa49d245b328984989d5caf9450f34bfc0ed16e96b58fa9913"
      ],
      [
        "841d6063a586fa475a724604da03bc5b92a2e0d2e0a36acfe4c73a5514742881",
        "73867f59c0659e81904f9a1c7543698e62562d6744c169ce7a36de01a8d6154"
      ],
      [
        "5e95bb399a6971d376026947f89bde2f282b33810928be4ded112ac4d70e20d5",
        "39f23f366809085beebfc71181313775a99c9aed7d8ba38b161384c746012865"
      ],
      [
        "36e4641a53948fd476c39f8a99fd974e5ec07564b5315d8bf99471bca0ef2f66",
        "d2424b1b1abe4eb8164227b085c9aa9456ea13493fd563e06fd51cf5694c78fc"
      ],
      [
        "336581ea7bfbbb290c191a2f507a41cf5643842170e914faeab27c2c579f726",
        "ead12168595fe1be99252129b6e56b3391f7ab1410cd1e0ef3dcdcabd2fda224"
      ],
      [
        "8ab89816dadfd6b6a1f2634fcf00ec8403781025ed6890c4849742706bd43ede",
        "6fdcef09f2f6d0a044e654aef624136f503d459c3e89845858a47a9129cdd24e"
      ],
      [
        "1e33f1a746c9c5778133344d9299fcaa20b0938e8acff2544bb40284b8c5fb94",
        "60660257dd11b3aa9c8ed618d24edff2306d320f1d03010e33a7d2057f3b3b6"
      ],
      [
        "85b7c1dcb3cec1b7ee7f30ded79dd20a0ed1f4cc18cbcfcfa410361fd8f08f31",
        "3d98a9cdd026dd43f39048f25a8847f4fcafad1895d7a633c6fed3c35e999511"
      ],
      [
        "29df9fbd8d9e46509275f4b125d6d45d7fbe9a3b878a7af872a2800661ac5f51",
        "b4c4fe99c775a606e2d8862179139ffda61dc861c019e55cd2876eb2a27d84b"
      ],
      [
        "a0b1cae06b0a847a3fea6e671aaf8adfdfe58ca2f768105c8082b2e449fce252",
        "ae434102edde0958ec4b19d917a6a28e6b72da1834aff0e650f049503a296cf2"
      ],
      [
        "4e8ceafb9b3e9a136dc7ff67e840295b499dfb3b2133e4ba113f2e4c0e121e5",
        "cf2174118c8b6d7a4b48f6d534ce5c79422c086a63460502b827ce62a326683c"
      ],
      [
        "d24a44e047e19b6f5afb81c7ca2f69080a5076689a010919f42725c2b789a33b",
        "6fb8d5591b466f8fc63db50f1c0f1c69013f996887b8244d2cdec417afea8fa3"
      ],
      [
        "ea01606a7a6c9cdd249fdfcfacb99584001edd28abbab77b5104e98e8e3b35d4",
        "322af4908c7312b0cfbfe369f7a7b3cdb7d4494bc2823700cfd652188a3ea98d"
      ],
      [
        "af8addbf2b661c8a6c6328655eb96651252007d8c5ea31be4ad196de8ce2131f",
        "6749e67c029b85f52a034eafd096836b2520818680e26ac8f3dfbcdb71749700"
      ],
      [
        "e3ae1974566ca06cc516d47e0fb165a674a3dabcfca15e722f0e3450f45889",
        "2aeabe7e4531510116217f07bf4d07300de97e4874f81f533420a72eeb0bd6a4"
      ],
      [
        "591ee355313d99721cf6993ffed1e3e301993ff3ed258802075ea8ced397e246",
        "b0ea558a113c30bea60fc4775460c7901ff0b053d25ca2bdeee98f1a4be5d196"
      ],
      [
        "11396d55fda54c49f19aa97318d8da61fa8584e47b084945077cf03255b52984",
        "998c74a8cd45ac01289d5833a7beb4744ff536b01b257be4c5767bea93ea57a4"
      ],
      [
        "3c5d2a1ba39c5a1790000738c9e0c40b8dcdfd5468754b6405540157e017aa7a",
        "b2284279995a34e2f9d4de7396fc18b80f9b8b9fdd270f6661f79ca4c81bd257"
      ],
      [
        "cc8704b8a60a0defa3a99a7299f2e9c3fbc395afb04ac078425ef8a1793cc030",
        "bdd46039feed17881d1e0862db347f8cf395b74fc4bcdc4e940b74e3ac1f1b13"
      ],
      [
        "c533e4f7ea8555aacd9777ac5cad29b97dd4defccc53ee7ea204119b2889b197",
        "6f0a256bc5efdf429a2fb6242f1a43a2d9b925bb4a4b3a26bb8e0f45eb596096"
      ],
      [
        "c14f8f2ccb27d6f109f6d08d03cc96a69ba8c34eec07bbcf566d48e33da6593",
        "c359d6923bb398f7fd4473e16fe1c28475b740dd098075e6c0e8649113dc3a38"
      ],
      [
        "a6cbc3046bc6a450bac24789fa17115a4c9739ed75f8f21ce441f72e0b90e6ef",
        "21ae7f4680e889bb130619e2c0f95a360ceb573c70603139862afd617fa9b9f"
      ],
      [
        "347d6d9a02c48927ebfb86c1359b1caf130a3c0267d11ce6344b39f99d43cc38",
        "60ea7f61a353524d1c987f6ecec92f086d565ab687870cb12689ff1e31c74448"
      ],
      [
        "da6545d2181db8d983f7dcb375ef5866d47c67b1bf31c8cf855ef7437b72656a",
        "49b96715ab6878a79e78f07ce5680c5d6673051b4935bd897fea824b77dc208a"
      ],
      [
        "c40747cc9d012cb1a13b8148309c6de7ec25d6945d657146b9d5994b8feb1111",
        "5ca560753be2a12fc6de6caf2cb489565db936156b9514e1bb5e83037e0fa2d4"
      ],
      [
        "4e42c8ec82c99798ccf3a610be870e78338c7f713348bd34c8203ef4037f3502",
        "7571d74ee5e0fb92a7a8b33a07783341a5492144cc54bcc40a94473693606437"
      ],
      [
        "3775ab7089bc6af823aba2e1af70b236d251cadb0c86743287522a1b3b0dedea",
        "be52d107bcfa09d8bcb9736a828cfa7fac8db17bf7a76a2c42ad961409018cf7"
      ],
      [
        "cee31cbf7e34ec379d94fb814d3d775ad954595d1314ba8846959e3e82f74e26",
        "8fd64a14c06b589c26b947ae2bcf6bfa0149ef0be14ed4d80f448a01c43b1c6d"
      ],
      [
        "b4f9eaea09b6917619f6ea6a4eb5464efddb58fd45b1ebefcdc1a01d08b47986",
        "39e5c9925b5a54b07433a4f18c61726f8bb131c012ca542eb24a8ac07200682a"
      ],
      [
        "d4263dfc3d2df923a0179a48966d30ce84e2515afc3dccc1b77907792ebcc60e",
        "62dfaf07a0f78feb30e30d6295853ce189e127760ad6cf7fae164e122a208d54"
      ],
      [
        "48457524820fa65a4f8d35eb6930857c0032acc0a4a2de422233eeda897612c4",
        "25a748ab367979d98733c38a1fa1c2e7dc6cc07db2d60a9ae7a76aaa49bd0f77"
      ],
      [
        "dfeeef1881101f2cb11644f3a2afdfc2045e19919152923f367a1767c11cceda",
        "ecfb7056cf1de042f9420bab396793c0c390bde74b4bbdff16a83ae09a9a7517"
      ],
      [
        "6d7ef6b17543f8373c573f44e1f389835d89bcbc6062ced36c82df83b8fae859",
        "cd450ec335438986dfefa10c57fea9bcc521a0959b2d80bbf74b190dca712d10"
      ],
      [
        "e75605d59102a5a2684500d3b991f2e3f3c88b93225547035af25af66e04541f",
        "f5c54754a8f71ee540b9b48728473e314f729ac5308b06938360990e2bfad125"
      ],
      [
        "eb98660f4c4dfaa06a2be453d5020bc99a0c2e60abe388457dd43fefb1ed620c",
        "6cb9a8876d9cb8520609af3add26cd20a0a7cd8a9411131ce85f44100099223e"
      ],
      [
        "13e87b027d8514d35939f2e6892b19922154596941888336dc3563e3b8dba942",
        "fef5a3c68059a6dec5d624114bf1e91aac2b9da568d6abeb2570d55646b8adf1"
      ],
      [
        "ee163026e9fd6fe017c38f06a5be6fc125424b371ce2708e7bf4491691e5764a",
        "1acb250f255dd61c43d94ccc670d0f58f49ae3fa15b96623e5430da0ad6c62b2"
      ],
      [
        "b268f5ef9ad51e4d78de3a750c2dc89b1e626d43505867999932e5db33af3d80",
        "5f310d4b3c99b9ebb19f77d41c1dee018cf0d34fd4191614003e945a1216e423"
      ],
      [
        "ff07f3118a9df035e9fad85eb6c7bfe42b02f01ca99ceea3bf7ffdba93c4750d",
        "438136d603e858a3a5c440c38eccbaddc1d2942114e2eddd4740d098ced1f0d8"
      ],
      [
        "8d8b9855c7c052a34146fd20ffb658bea4b9f69e0d825ebec16e8c3ce2b526a1",
        "cdb559eedc2d79f926baf44fb84ea4d44bcf50fee51d7ceb30e2e7f463036758"
      ],
      [
        "52db0b5384dfbf05bfa9d472d7ae26dfe4b851ceca91b1eba54263180da32b63",
        "c3b997d050ee5d423ebaf66a6db9f57b3180c902875679de924b69d84a7b375"
      ],
      [
        "e62f9490d3d51da6395efd24e80919cc7d0f29c3f3fa48c6fff543becbd43352",
        "6d89ad7ba4876b0b22c2ca280c682862f342c8591f1daf5170e07bfd9ccafa7d"
      ],
      [
        "7f30ea2476b399b4957509c88f77d0191afa2ff5cb7b14fd6d8e7d65aaab1193",
        "ca5ef7d4b231c94c3b15389a5f6311e9daff7bb67b103e9880ef4bff637acaec"
      ],
      [
        "5098ff1e1d9f14fb46a210fada6c903fef0fb7b4a1dd1d9ac60a0361800b7a00",
        "9731141d81fc8f8084d37c6e7542006b3ee1b40d60dfe5362a5b132fd17ddc0"
      ],
      [
        "32b78c7de9ee512a72895be6b9cbefa6e2f3c4ccce445c96b9f2c81e2778ad58",
        "ee1849f513df71e32efc3896ee28260c73bb80547ae2275ba497237794c8753c"
      ],
      [
        "e2cb74fddc8e9fbcd076eef2a7c72b0ce37d50f08269dfc074b581550547a4f7",
        "d3aa2ed71c9dd2247a62df062736eb0baddea9e36122d2be8641abcb005cc4a4"
      ],
      [
        "8438447566d4d7bedadc299496ab357426009a35f235cb141be0d99cd10ae3a8",
        "c4e1020916980a4da5d01ac5e6ad330734ef0d7906631c4f2390426b2edd791f"
      ],
      [
        "4162d488b89402039b584c6fc6c308870587d9c46f660b878ab65c82c711d67e",
        "67163e903236289f776f22c25fb8a3afc1732f2b84b4e95dbda47ae5a0852649"
      ],
      [
        "3fad3fa84caf0f34f0f89bfd2dcf54fc175d767aec3e50684f3ba4a4bf5f683d",
        "cd1bc7cb6cc407bb2f0ca647c718a730cf71872e7d0d2a53fa20efcdfe61826"
      ],
      [
        "674f2600a3007a00568c1a7ce05d0816c1fb84bf1370798f1c69532faeb1a86b",
        "299d21f9413f33b3edf43b257004580b70db57da0b182259e09eecc69e0d38a5"
      ],
      [
        "d32f4da54ade74abb81b815ad1fb3b263d82d6c692714bcff87d29bd5ee9f08f",
        "f9429e738b8e53b968e99016c059707782e14f4535359d582fc416910b3eea87"
      ],
      [
        "30e4e670435385556e593657135845d36fbb6931f72b08cb1ed954f1e3ce3ff6",
        "462f9bce619898638499350113bbc9b10a878d35da70740dc695a559eb88db7b"
      ],
      [
        "be2062003c51cc3004682904330e4dee7f3dcd10b01e580bf1971b04d4cad297",
        "62188bc49d61e5428573d48a74e1c655b1c61090905682a0d5558ed72dccb9bc"
      ],
      [
        "93144423ace3451ed29e0fb9ac2af211cb6e84a601df5993c419859fff5df04a",
        "7c10dfb164c3425f5c71a3f9d7992038f1065224f72bb9d1d902a6d13037b47c"
      ],
      [
        "b015f8044f5fcbdcf21ca26d6c34fb8197829205c7b7d2a7cb66418c157b112c",
        "ab8c1e086d04e813744a655b2df8d5f83b3cdc6faa3088c1d3aea1454e3a1d5f"
      ],
      [
        "d5e9e1da649d97d89e4868117a465a3a4f8a18de57a140d36b3f2af341a21b52",
        "4cb04437f391ed73111a13cc1d4dd0db1693465c2240480d8955e8592f27447a"
      ],
      [
        "d3ae41047dd7ca065dbf8ed77b992439983005cd72e16d6f996a5316d36966bb",
        "bd1aeb21ad22ebb22a10f0303417c6d964f8cdd7df0aca614b10dc14d125ac46"
      ],
      [
        "463e2763d885f958fc66cdd22800f0a487197d0a82e377b49f80af87c897b065",
        "bfefacdb0e5d0fd7df3a311a94de062b26b80c61fbc97508b79992671ef7ca7f"
      ],
      [
        "7985fdfd127c0567c6f53ec1bb63ec3158e597c40bfe747c83cddfc910641917",
        "603c12daf3d9862ef2b25fe1de289aed24ed291e0ec6708703a5bd567f32ed03"
      ],
      [
        "74a1ad6b5f76e39db2dd249410eac7f99e74c59cb83d2d0ed5ff1543da7703e9",
        "cc6157ef18c9c63cd6193d83631bbea0093e0968942e8c33d5737fd790e0db08"
      ],
      [
        "30682a50703375f602d416664ba19b7fc9bab42c72747463a71d0896b22f6da3",
        "553e04f6b018b4fa6c8f39e7f311d3176290d0e0f19ca73f17714d9977a22ff8"
      ],
      [
        "9e2158f0d7c0d5f26c3791efefa79597654e7a2b2464f52b1ee6c1347769ef57",
        "712fcdd1b9053f09003a3481fa7762e9ffd7c8ef35a38509e2fbf2629008373"
      ],
      [
        "176e26989a43c9cfeba4029c202538c28172e566e3c4fce7322857f3be327d66",
        "ed8cc9d04b29eb877d270b4878dc43c19aefd31f4eee09ee7b47834c1fa4b1c3"
      ],
      [
        "75d46efea3771e6e68abb89a13ad747ecf1892393dfc4f1b7004788c50374da8",
        "9852390a99507679fd0b86fd2b39a868d7efc22151346e1a3ca4726586a6bed8"
      ],
      [
        "809a20c67d64900ffb698c4c825f6d5f2310fb0451c869345b7319f645605721",
        "9e994980d9917e22b76b061927fa04143d096ccc54963e6a5ebfa5f3f8e286c1"
      ],
      [
        "1b38903a43f7f114ed4500b4eac7083fdefece1cf29c63528d563446f972c180",
        "4036edc931a60ae889353f77fd53de4a2708b26b6f5da72ad3394119daf408f9"
      ]
    ]
  }
};
var curves_1 = createCommonjsModule(function(module, exports) {
  var curves = exports;
  var assert2 = utils_1.assert;
  function PresetCurve(options) {
    if (options.type === "short")
      this.curve = new curve_1.short(options);
    else if (options.type === "edwards")
      this.curve = new curve_1.edwards(options);
    else
      this.curve = new curve_1.mont(options);
    this.g = this.curve.g;
    this.n = this.curve.n;
    this.hash = options.hash;
    assert2(this.g.validate(), "Invalid curve");
    assert2(this.g.mul(this.n).isInfinity(), "Invalid curve, G*N != O");
  }
  curves.PresetCurve = PresetCurve;
  function defineCurve(name2, options) {
    Object.defineProperty(curves, name2, {
      configurable: true,
      enumerable: true,
      get: function() {
        var curve = new PresetCurve(options);
        Object.defineProperty(curves, name2, {
          configurable: true,
          enumerable: true,
          value: curve
        });
        return curve;
      }
    });
  }
  defineCurve("p192", {
    type: "short",
    prime: "p192",
    p: "ffffffff ffffffff ffffffff fffffffe ffffffff ffffffff",
    a: "ffffffff ffffffff ffffffff fffffffe ffffffff fffffffc",
    b: "64210519 e59c80e7 0fa7e9ab 72243049 feb8deec c146b9b1",
    n: "ffffffff ffffffff ffffffff 99def836 146bc9b1 b4d22831",
    hash: hash.sha256,
    gRed: false,
    g: [
      "188da80e b03090f6 7cbf20eb 43a18800 f4ff0afd 82ff1012",
      "07192b95 ffc8da78 631011ed 6b24cdd5 73f977a1 1e794811"
    ]
  });
  defineCurve("p224", {
    type: "short",
    prime: "p224",
    p: "ffffffff ffffffff ffffffff ffffffff 00000000 00000000 00000001",
    a: "ffffffff ffffffff ffffffff fffffffe ffffffff ffffffff fffffffe",
    b: "b4050a85 0c04b3ab f5413256 5044b0b7 d7bfd8ba 270b3943 2355ffb4",
    n: "ffffffff ffffffff ffffffff ffff16a2 e0b8f03e 13dd2945 5c5c2a3d",
    hash: hash.sha256,
    gRed: false,
    g: [
      "b70e0cbd 6bb4bf7f 321390b9 4a03c1d3 56c21122 343280d6 115c1d21",
      "bd376388 b5f723fb 4c22dfe6 cd4375a0 5a074764 44d58199 85007e34"
    ]
  });
  defineCurve("p256", {
    type: "short",
    prime: null,
    p: "ffffffff 00000001 00000000 00000000 00000000 ffffffff ffffffff ffffffff",
    a: "ffffffff 00000001 00000000 00000000 00000000 ffffffff ffffffff fffffffc",
    b: "5ac635d8 aa3a93e7 b3ebbd55 769886bc 651d06b0 cc53b0f6 3bce3c3e 27d2604b",
    n: "ffffffff 00000000 ffffffff ffffffff bce6faad a7179e84 f3b9cac2 fc632551",
    hash: hash.sha256,
    gRed: false,
    g: [
      "6b17d1f2 e12c4247 f8bce6e5 63a440f2 77037d81 2deb33a0 f4a13945 d898c296",
      "4fe342e2 fe1a7f9b 8ee7eb4a 7c0f9e16 2bce3357 6b315ece cbb64068 37bf51f5"
    ]
  });
  defineCurve("p384", {
    type: "short",
    prime: null,
    p: "ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff fffffffe ffffffff 00000000 00000000 ffffffff",
    a: "ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff fffffffe ffffffff 00000000 00000000 fffffffc",
    b: "b3312fa7 e23ee7e4 988e056b e3f82d19 181d9c6e fe814112 0314088f 5013875a c656398d 8a2ed19d 2a85c8ed d3ec2aef",
    n: "ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff c7634d81 f4372ddf 581a0db2 48b0a77a ecec196a ccc52973",
    hash: hash.sha384,
    gRed: false,
    g: [
      "aa87ca22 be8b0537 8eb1c71e f320ad74 6e1d3b62 8ba79b98 59f741e0 82542a38 5502f25d bf55296c 3a545e38 72760ab7",
      "3617de4a 96262c6f 5d9e98bf 9292dc29 f8f41dbd 289a147c e9da3113 b5f0b8c0 0a60b1ce 1d7e819d 7a431d7c 90ea0e5f"
    ]
  });
  defineCurve("p521", {
    type: "short",
    prime: null,
    p: "000001ff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff",
    a: "000001ff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff fffffffc",
    b: "00000051 953eb961 8e1c9a1f 929a21a0 b68540ee a2da725b 99b315f3 b8b48991 8ef109e1 56193951 ec7e937b 1652c0bd 3bb1bf07 3573df88 3d2c34f1 ef451fd4 6b503f00",
    n: "000001ff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff fffffffa 51868783 bf2f966b 7fcc0148 f709a5d0 3bb5c9b8 899c47ae bb6fb71e 91386409",
    hash: hash.sha512,
    gRed: false,
    g: [
      "000000c6 858e06b7 0404e9cd 9e3ecb66 2395b442 9c648139 053fb521 f828af60 6b4d3dba a14b5e77 efe75928 fe1dc127 a2ffa8de 3348b3c1 856a429b f97e7e31 c2e5bd66",
      "00000118 39296a78 9a3bc004 5c8a5fb4 2c7d1bd9 98f54449 579b4468 17afbd17 273e662c 97ee7299 5ef42640 c550b901 3fad0761 353c7086 a272c240 88be9476 9fd16650"
    ]
  });
  defineCurve("curve25519", {
    type: "mont",
    prime: "p25519",
    p: "7fffffffffffffff ffffffffffffffff ffffffffffffffff ffffffffffffffed",
    a: "76d06",
    b: "1",
    n: "1000000000000000 0000000000000000 14def9dea2f79cd6 5812631a5cf5d3ed",
    hash: hash.sha256,
    gRed: false,
    g: [
      "9"
    ]
  });
  defineCurve("ed25519", {
    type: "edwards",
    prime: "p25519",
    p: "7fffffffffffffff ffffffffffffffff ffffffffffffffff ffffffffffffffed",
    a: "-1",
    c: "1",
    d: "52036cee2b6ffe73 8cc740797779e898 00700a4d4141d8ab 75eb4dca135978a3",
    n: "1000000000000000 0000000000000000 14def9dea2f79cd6 5812631a5cf5d3ed",
    hash: hash.sha256,
    gRed: false,
    g: [
      "216936d3cd6e53fec0a4e231fdd6dc5c692cc7609525a7b2c9562d608f25d51a",
      "6666666666666666666666666666666666666666666666666666666666666658"
    ]
  });
  var pre;
  try {
    pre = secp256k1;
  } catch (e) {
    pre = void 0;
  }
  defineCurve("secp256k1", {
    type: "short",
    prime: "k256",
    p: "ffffffff ffffffff ffffffff ffffffff ffffffff ffffffff fffffffe fffffc2f",
    a: "0",
    b: "7",
    n: "ffffffff ffffffff ffffffff fffffffe baaedce6 af48a03b bfd25e8c d0364141",
    h: "1",
    hash: hash.sha256,
    beta: "7ae96a2b657c07106e64479eac3434e99cf0497512f58995c1396c28719501ee",
    lambda: "5363ad4cc05c30e0a5261c028812645a122e22ea20816678df02967c1b23bd72",
    basis: [
      {
        a: "3086d221a7d46bcde86c90e49284eb15",
        b: "-e4437ed6010e88286f547fa90abfe4c3"
      },
      {
        a: "114ca50f7a8e2f3f657c1108d9d44cfd8",
        b: "3086d221a7d46bcde86c90e49284eb15"
      }
    ],
    gRed: false,
    g: [
      "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798",
      "483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8",
      pre
    ]
  });
});
var assert$3 = utils_1.assert;
function KeyPair(ec2, options) {
  this.ec = ec2;
  this.priv = null;
  this.pub = null;
  if (options.priv)
    this._importPrivate(options.priv, options.privEnc);
  if (options.pub)
    this._importPublic(options.pub, options.pubEnc);
}
var key = KeyPair;
KeyPair.fromPublic = function fromPublic(ec2, pub2, enc) {
  if (pub2 instanceof KeyPair)
    return pub2;
  return new KeyPair(ec2, {
    pub: pub2,
    pubEnc: enc
  });
};
KeyPair.fromPrivate = function fromPrivate(ec2, priv2, enc) {
  if (priv2 instanceof KeyPair)
    return priv2;
  return new KeyPair(ec2, {
    priv: priv2,
    privEnc: enc
  });
};
KeyPair.prototype.validate = function validate6() {
  var pub2 = this.getPublic();
  if (pub2.isInfinity())
    return {result: false, reason: "Invalid public key"};
  if (!pub2.validate())
    return {result: false, reason: "Public key is not a point"};
  if (!pub2.mul(this.ec.curve.n).isInfinity())
    return {result: false, reason: "Public key * N != O"};
  return {result: true, reason: null};
};
KeyPair.prototype.getPublic = function getPublic(compact, enc) {
  if (typeof compact === "string") {
    enc = compact;
    compact = null;
  }
  if (!this.pub)
    this.pub = this.ec.g.mul(this.priv);
  if (!enc)
    return this.pub;
  return this.pub.encode(enc, compact);
};
KeyPair.prototype.getPrivate = function getPrivate(enc) {
  if (enc === "hex")
    return this.priv.toString(16, 2);
  else
    return this.priv;
};
KeyPair.prototype._importPrivate = function _importPrivate(key2, enc) {
  this.priv = new BN(key2, enc || 16);
  this.priv = this.priv.umod(this.ec.curve.n);
};
KeyPair.prototype._importPublic = function _importPublic(key2, enc) {
  if (key2.x || key2.y) {
    if (this.ec.curve.type === "mont") {
      assert$3(key2.x, "Need x coordinate");
    } else if (this.ec.curve.type === "short" || this.ec.curve.type === "edwards") {
      assert$3(key2.x && key2.y, "Need both x and y coordinate");
    }
    this.pub = this.ec.curve.point(key2.x, key2.y);
    return;
  }
  this.pub = this.ec.curve.decodePoint(key2, enc);
};
KeyPair.prototype.derive = function derive(pub2) {
  if (!pub2.validate()) {
    assert$3(pub2.validate(), "public point not validated");
  }
  return pub2.mul(this.priv).getX();
};
KeyPair.prototype.sign = function sign(msg, enc, options) {
  return this.ec.sign(msg, this, enc, options);
};
KeyPair.prototype.verify = function verify(msg, signature2) {
  return this.ec.verify(msg, signature2, this);
};
KeyPair.prototype.inspect = function inspect5() {
  return "<Key priv: " + (this.priv && this.priv.toString(16, 2)) + " pub: " + (this.pub && this.pub.inspect()) + " >";
};
var assert$4 = utils_1.assert;
function Signature(options, enc) {
  if (options instanceof Signature)
    return options;
  if (this._importDER(options, enc))
    return;
  assert$4(options.r && options.s, "Signature without r or s");
  this.r = new BN(options.r, 16);
  this.s = new BN(options.s, 16);
  if (options.recoveryParam === void 0)
    this.recoveryParam = null;
  else
    this.recoveryParam = options.recoveryParam;
}
var signature = Signature;
function Position() {
  this.place = 0;
}
function getLength(buf, p) {
  var initial = buf[p.place++];
  if (!(initial & 128)) {
    return initial;
  }
  var octetLen = initial & 15;
  if (octetLen === 0 || octetLen > 4) {
    return false;
  }
  var val = 0;
  for (var i = 0, off = p.place; i < octetLen; i++, off++) {
    val <<= 8;
    val |= buf[off];
    val >>>= 0;
  }
  if (val <= 127) {
    return false;
  }
  p.place = off;
  return val;
}
function rmPadding(buf) {
  var i = 0;
  var len = buf.length - 1;
  while (!buf[i] && !(buf[i + 1] & 128) && i < len) {
    i++;
  }
  if (i === 0) {
    return buf;
  }
  return buf.slice(i);
}
Signature.prototype._importDER = function _importDER(data, enc) {
  data = utils_1.toArray(data, enc);
  var p = new Position();
  if (data[p.place++] !== 48) {
    return false;
  }
  var len = getLength(data, p);
  if (len === false) {
    return false;
  }
  if (len + p.place !== data.length) {
    return false;
  }
  if (data[p.place++] !== 2) {
    return false;
  }
  var rlen = getLength(data, p);
  if (rlen === false) {
    return false;
  }
  var r = data.slice(p.place, rlen + p.place);
  p.place += rlen;
  if (data[p.place++] !== 2) {
    return false;
  }
  var slen = getLength(data, p);
  if (slen === false) {
    return false;
  }
  if (data.length !== slen + p.place) {
    return false;
  }
  var s = data.slice(p.place, slen + p.place);
  if (r[0] === 0) {
    if (r[1] & 128) {
      r = r.slice(1);
    } else {
      return false;
    }
  }
  if (s[0] === 0) {
    if (s[1] & 128) {
      s = s.slice(1);
    } else {
      return false;
    }
  }
  this.r = new BN(r);
  this.s = new BN(s);
  this.recoveryParam = null;
  return true;
};
function constructLength(arr, len) {
  if (len < 128) {
    arr.push(len);
    return;
  }
  var octets = 1 + (Math.log(len) / Math.LN2 >>> 3);
  arr.push(octets | 128);
  while (--octets) {
    arr.push(len >>> (octets << 3) & 255);
  }
  arr.push(len);
}
Signature.prototype.toDER = function toDER(enc) {
  var r = this.r.toArray();
  var s = this.s.toArray();
  if (r[0] & 128)
    r = [0].concat(r);
  if (s[0] & 128)
    s = [0].concat(s);
  r = rmPadding(r);
  s = rmPadding(s);
  while (!s[0] && !(s[1] & 128)) {
    s = s.slice(1);
  }
  var arr = [2];
  constructLength(arr, r.length);
  arr = arr.concat(r);
  arr.push(2);
  constructLength(arr, s.length);
  var backHalf = arr.concat(s);
  var res = [48];
  constructLength(res, backHalf.length);
  res = res.concat(backHalf);
  return utils_1.encode(res, enc);
};
var assert$5 = utils_1.assert;
function EC(options) {
  if (!(this instanceof EC))
    return new EC(options);
  if (typeof options === "string") {
    assert$5(Object.prototype.hasOwnProperty.call(curves_1, options), "Unknown curve " + options);
    options = curves_1[options];
  }
  if (options instanceof curves_1.PresetCurve)
    options = {curve: options};
  this.curve = options.curve.curve;
  this.n = this.curve.n;
  this.nh = this.n.ushrn(1);
  this.g = this.curve.g;
  this.g = options.curve.g;
  this.g.precompute(options.curve.n.bitLength() + 1);
  this.hash = options.hash || options.curve.hash;
}
var ec = EC;
EC.prototype.keyPair = function keyPair(options) {
  return new key(this, options);
};
EC.prototype.keyFromPrivate = function keyFromPrivate(priv2, enc) {
  return key.fromPrivate(this, priv2, enc);
};
EC.prototype.keyFromPublic = function keyFromPublic(pub2, enc) {
  return key.fromPublic(this, pub2, enc);
};
EC.prototype.genKeyPair = function genKeyPair(options) {
  if (!options)
    options = {};
  var drbg = new HmacDRBG({
    hash: this.hash,
    pers: options.pers,
    persEnc: options.persEnc || "utf8",
    entropy: options.entropy || rand(this.hash.hmacStrength),
    entropyEnc: options.entropy && options.entropyEnc || "utf8",
    nonce: this.n.toArray()
  });
  var bytes = this.n.byteLength();
  var ns2 = this.n.sub(new BN(2));
  for (; ; ) {
    var priv2 = new BN(drbg.generate(bytes));
    if (priv2.cmp(ns2) > 0)
      continue;
    priv2.iaddn(1);
    return this.keyFromPrivate(priv2);
  }
};
EC.prototype._truncateToN = function _truncateToN(msg, truncOnly) {
  var delta = msg.byteLength() * 8 - this.n.bitLength();
  if (delta > 0)
    msg = msg.ushrn(delta);
  if (!truncOnly && msg.cmp(this.n) >= 0)
    return msg.sub(this.n);
  else
    return msg;
};
EC.prototype.sign = function sign2(msg, key2, enc, options) {
  if (typeof enc === "object") {
    options = enc;
    enc = null;
  }
  if (!options)
    options = {};
  key2 = this.keyFromPrivate(key2, enc);
  msg = this._truncateToN(new BN(msg, 16));
  var bytes = this.n.byteLength();
  var bkey = key2.getPrivate().toArray("be", bytes);
  var nonce = msg.toArray("be", bytes);
  var drbg = new HmacDRBG({
    hash: this.hash,
    entropy: bkey,
    nonce,
    pers: options.pers,
    persEnc: options.persEnc || "utf8"
  });
  var ns1 = this.n.sub(new BN(1));
  for (var iter = 0; ; iter++) {
    var k = options.k ? options.k(iter) : new BN(drbg.generate(this.n.byteLength()));
    k = this._truncateToN(k, true);
    if (k.cmpn(1) <= 0 || k.cmp(ns1) >= 0)
      continue;
    var kp = this.g.mul(k);
    if (kp.isInfinity())
      continue;
    var kpX = kp.getX();
    var r = kpX.umod(this.n);
    if (r.cmpn(0) === 0)
      continue;
    var s = k.invm(this.n).mul(r.mul(key2.getPrivate()).iadd(msg));
    s = s.umod(this.n);
    if (s.cmpn(0) === 0)
      continue;
    var recoveryParam = (kp.getY().isOdd() ? 1 : 0) | (kpX.cmp(r) !== 0 ? 2 : 0);
    if (options.canonical && s.cmp(this.nh) > 0) {
      s = this.n.sub(s);
      recoveryParam ^= 1;
    }
    return new signature({r, s, recoveryParam});
  }
};
EC.prototype.verify = function verify2(msg, signature$12, key2, enc) {
  msg = this._truncateToN(new BN(msg, 16));
  key2 = this.keyFromPublic(key2, enc);
  signature$12 = new signature(signature$12, "hex");
  var r = signature$12.r;
  var s = signature$12.s;
  if (r.cmpn(1) < 0 || r.cmp(this.n) >= 0)
    return false;
  if (s.cmpn(1) < 0 || s.cmp(this.n) >= 0)
    return false;
  var sinv = s.invm(this.n);
  var u1 = sinv.mul(msg).umod(this.n);
  var u2 = sinv.mul(r).umod(this.n);
  var p;
  if (!this.curve._maxwellTrick) {
    p = this.g.mulAdd(u1, key2.getPublic(), u2);
    if (p.isInfinity())
      return false;
    return p.getX().umod(this.n).cmp(r) === 0;
  }
  p = this.g.jmulAdd(u1, key2.getPublic(), u2);
  if (p.isInfinity())
    return false;
  return p.eqXToP(r);
};
EC.prototype.recoverPubKey = function(msg, signature$12, j, enc) {
  assert$5((3 & j) === j, "The recovery param is more than two bits");
  signature$12 = new signature(signature$12, enc);
  var n = this.n;
  var e = new BN(msg);
  var r = signature$12.r;
  var s = signature$12.s;
  var isYOdd = j & 1;
  var isSecondKey = j >> 1;
  if (r.cmp(this.curve.p.umod(this.curve.n)) >= 0 && isSecondKey)
    throw new Error("Unable to find sencond key candinate");
  if (isSecondKey)
    r = this.curve.pointFromX(r.add(this.curve.n), isYOdd);
  else
    r = this.curve.pointFromX(r, isYOdd);
  var rInv = signature$12.r.invm(n);
  var s1 = n.sub(e).mul(rInv).umod(n);
  var s2 = s.mul(rInv).umod(n);
  return this.g.mulAdd(s1, r, s2);
};
EC.prototype.getKeyRecoveryParam = function(e, signature$12, Q, enc) {
  signature$12 = new signature(signature$12, enc);
  if (signature$12.recoveryParam !== null)
    return signature$12.recoveryParam;
  for (var i = 0; i < 4; i++) {
    var Qprime;
    try {
      Qprime = this.recoverPubKey(e, signature$12, i);
    } catch (e2) {
      continue;
    }
    if (Qprime.eq(Q))
      return i;
  }
  throw new Error("Unable to find valid recovery factor");
};
var assert$6 = utils_1.assert;
var parseBytes = utils_1.parseBytes;
var cachedProperty = utils_1.cachedProperty;
function KeyPair$1(eddsa2, params) {
  this.eddsa = eddsa2;
  this._secret = parseBytes(params.secret);
  if (eddsa2.isPoint(params.pub))
    this._pub = params.pub;
  else
    this._pubBytes = parseBytes(params.pub);
}
KeyPair$1.fromPublic = function fromPublic2(eddsa2, pub2) {
  if (pub2 instanceof KeyPair$1)
    return pub2;
  return new KeyPair$1(eddsa2, {pub: pub2});
};
KeyPair$1.fromSecret = function fromSecret(eddsa2, secret2) {
  if (secret2 instanceof KeyPair$1)
    return secret2;
  return new KeyPair$1(eddsa2, {secret: secret2});
};
KeyPair$1.prototype.secret = function secret() {
  return this._secret;
};
cachedProperty(KeyPair$1, "pubBytes", function pubBytes() {
  return this.eddsa.encodePoint(this.pub());
});
cachedProperty(KeyPair$1, "pub", function pub() {
  if (this._pubBytes)
    return this.eddsa.decodePoint(this._pubBytes);
  return this.eddsa.g.mul(this.priv());
});
cachedProperty(KeyPair$1, "privBytes", function privBytes() {
  var eddsa2 = this.eddsa;
  var hash3 = this.hash();
  var lastIx = eddsa2.encodingLength - 1;
  var a = hash3.slice(0, eddsa2.encodingLength);
  a[0] &= 248;
  a[lastIx] &= 127;
  a[lastIx] |= 64;
  return a;
});
cachedProperty(KeyPair$1, "priv", function priv() {
  return this.eddsa.decodeInt(this.privBytes());
});
cachedProperty(KeyPair$1, "hash", function hash2() {
  return this.eddsa.hash().update(this.secret()).digest();
});
cachedProperty(KeyPair$1, "messagePrefix", function messagePrefix() {
  return this.hash().slice(this.eddsa.encodingLength);
});
KeyPair$1.prototype.sign = function sign3(message) {
  assert$6(this._secret, "KeyPair can only verify");
  return this.eddsa.sign(message, this);
};
KeyPair$1.prototype.verify = function verify3(message, sig) {
  return this.eddsa.verify(message, sig, this);
};
KeyPair$1.prototype.getSecret = function getSecret(enc) {
  assert$6(this._secret, "KeyPair is public only");
  return utils_1.encode(this.secret(), enc);
};
KeyPair$1.prototype.getPublic = function getPublic2(enc) {
  return utils_1.encode(this.pubBytes(), enc);
};
var key$1 = KeyPair$1;
var assert$7 = utils_1.assert;
var cachedProperty$1 = utils_1.cachedProperty;
var parseBytes$1 = utils_1.parseBytes;
function Signature$1(eddsa2, sig) {
  this.eddsa = eddsa2;
  if (typeof sig !== "object")
    sig = parseBytes$1(sig);
  if (Array.isArray(sig)) {
    sig = {
      R: sig.slice(0, eddsa2.encodingLength),
      S: sig.slice(eddsa2.encodingLength)
    };
  }
  assert$7(sig.R && sig.S, "Signature without R or S");
  if (eddsa2.isPoint(sig.R))
    this._R = sig.R;
  if (sig.S instanceof BN)
    this._S = sig.S;
  this._Rencoded = Array.isArray(sig.R) ? sig.R : sig.Rencoded;
  this._Sencoded = Array.isArray(sig.S) ? sig.S : sig.Sencoded;
}
cachedProperty$1(Signature$1, "S", function S() {
  return this.eddsa.decodeInt(this.Sencoded());
});
cachedProperty$1(Signature$1, "R", function R() {
  return this.eddsa.decodePoint(this.Rencoded());
});
cachedProperty$1(Signature$1, "Rencoded", function Rencoded() {
  return this.eddsa.encodePoint(this.R());
});
cachedProperty$1(Signature$1, "Sencoded", function Sencoded() {
  return this.eddsa.encodeInt(this.S());
});
Signature$1.prototype.toBytes = function toBytes() {
  return this.Rencoded().concat(this.Sencoded());
};
Signature$1.prototype.toHex = function toHex() {
  return utils_1.encode(this.toBytes(), "hex").toUpperCase();
};
var signature$1 = Signature$1;
var assert$8 = utils_1.assert;
var parseBytes$2 = utils_1.parseBytes;
function EDDSA(curve) {
  assert$8(curve === "ed25519", "only tested with ed25519 so far");
  if (!(this instanceof EDDSA))
    return new EDDSA(curve);
  curve = curves_1[curve].curve;
  this.curve = curve;
  this.g = curve.g;
  this.g.precompute(curve.n.bitLength() + 1);
  this.pointClass = curve.point().constructor;
  this.encodingLength = Math.ceil(curve.n.bitLength() / 8);
  this.hash = hash.sha512;
}
var eddsa = EDDSA;
EDDSA.prototype.sign = function sign4(message, secret2) {
  message = parseBytes$2(message);
  var key2 = this.keyFromSecret(secret2);
  var r = this.hashInt(key2.messagePrefix(), message);
  var R2 = this.g.mul(r);
  var Rencoded2 = this.encodePoint(R2);
  var s_ = this.hashInt(Rencoded2, key2.pubBytes(), message).mul(key2.priv());
  var S2 = r.add(s_).umod(this.curve.n);
  return this.makeSignature({R: R2, S: S2, Rencoded: Rencoded2});
};
EDDSA.prototype.verify = function verify4(message, sig, pub2) {
  message = parseBytes$2(message);
  sig = this.makeSignature(sig);
  var key2 = this.keyFromPublic(pub2);
  var h = this.hashInt(sig.Rencoded(), key2.pubBytes(), message);
  var SG = this.g.mul(sig.S());
  var RplusAh = sig.R().add(key2.pub().mul(h));
  return RplusAh.eq(SG);
};
EDDSA.prototype.hashInt = function hashInt() {
  var hash3 = this.hash();
  for (var i = 0; i < arguments.length; i++)
    hash3.update(arguments[i]);
  return utils_1.intFromLE(hash3.digest()).umod(this.curve.n);
};
EDDSA.prototype.keyFromPublic = function keyFromPublic2(pub2) {
  return key$1.fromPublic(this, pub2);
};
EDDSA.prototype.keyFromSecret = function keyFromSecret(secret2) {
  return key$1.fromSecret(this, secret2);
};
EDDSA.prototype.makeSignature = function makeSignature(sig) {
  if (sig instanceof signature$1)
    return sig;
  return new signature$1(this, sig);
};
EDDSA.prototype.encodePoint = function encodePoint(point5) {
  var enc = point5.getY().toArray("le", this.encodingLength);
  enc[this.encodingLength - 1] |= point5.getX().isOdd() ? 128 : 0;
  return enc;
};
EDDSA.prototype.decodePoint = function decodePoint3(bytes) {
  bytes = utils_1.parseBytes(bytes);
  var lastIx = bytes.length - 1;
  var normed = bytes.slice(0, lastIx).concat(bytes[lastIx] & ~128);
  var xIsOdd = (bytes[lastIx] & 128) !== 0;
  var y = utils_1.intFromLE(normed);
  return this.curve.pointFromY(y, xIsOdd);
};
EDDSA.prototype.encodeInt = function encodeInt(num) {
  return num.toArray("le", this.encodingLength);
};
EDDSA.prototype.decodeInt = function decodeInt(bytes) {
  return utils_1.intFromLE(bytes);
};
EDDSA.prototype.isPoint = function isPoint(val) {
  return val instanceof this.pointClass;
};
var elliptic_1 = createCommonjsModule(function(module, exports) {
  var elliptic = exports;
  elliptic.version = require$$0.version;
  elliptic.utils = utils_1;
  elliptic.rand = rand;
  elliptic.curve = curve_1;
  elliptic.curves = curves_1;
  elliptic.ec = ec;
  elliptic.eddsa = eddsa;
});
export default elliptic_1;
