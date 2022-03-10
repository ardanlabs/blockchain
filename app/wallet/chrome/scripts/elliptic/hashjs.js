import assert from "./minimalistic-assert.js";
import inherits2 from "./inherits.js";

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
var inherits_1 = inherits2;
function isSurrogatePair(msg, i) {
  if ((msg.charCodeAt(i) & 64512) !== 55296) {
    return false;
  }
  if (i < 0 || i + 1 >= msg.length) {
    return false;
  }
  return (msg.charCodeAt(i + 1) & 64512) === 56320;
}
function toArray(msg, enc) {
  if (Array.isArray(msg))
    return msg.slice();
  if (!msg)
    return [];
  var res = [];
  if (typeof msg === "string") {
    if (!enc) {
      var p = 0;
      for (var i = 0; i < msg.length; i++) {
        var c = msg.charCodeAt(i);
        if (c < 128) {
          res[p++] = c;
        } else if (c < 2048) {
          res[p++] = c >> 6 | 192;
          res[p++] = c & 63 | 128;
        } else if (isSurrogatePair(msg, i)) {
          c = 65536 + ((c & 1023) << 10) + (msg.charCodeAt(++i) & 1023);
          res[p++] = c >> 18 | 240;
          res[p++] = c >> 12 & 63 | 128;
          res[p++] = c >> 6 & 63 | 128;
          res[p++] = c & 63 | 128;
        } else {
          res[p++] = c >> 12 | 224;
          res[p++] = c >> 6 & 63 | 128;
          res[p++] = c & 63 | 128;
        }
      }
    } else if (enc === "hex") {
      msg = msg.replace(/[^a-z0-9]+/ig, "");
      if (msg.length % 2 !== 0)
        msg = "0" + msg;
      for (i = 0; i < msg.length; i += 2)
        res.push(parseInt(msg[i] + msg[i + 1], 16));
    }
  } else {
    for (i = 0; i < msg.length; i++)
      res[i] = msg[i] | 0;
  }
  return res;
}
var toArray_1 = toArray;
function toHex(msg) {
  var res = "";
  for (var i = 0; i < msg.length; i++)
    res += zero2(msg[i].toString(16));
  return res;
}
var toHex_1 = toHex;
function htonl(w) {
  var res = w >>> 24 | w >>> 8 & 65280 | w << 8 & 16711680 | (w & 255) << 24;
  return res >>> 0;
}
var htonl_1 = htonl;
function toHex32(msg, endian) {
  var res = "";
  for (var i = 0; i < msg.length; i++) {
    var w = msg[i];
    if (endian === "little")
      w = htonl(w);
    res += zero8(w.toString(16));
  }
  return res;
}
var toHex32_1 = toHex32;
function zero2(word) {
  if (word.length === 1)
    return "0" + word;
  else
    return word;
}
var zero2_1 = zero2;
function zero8(word) {
  if (word.length === 7)
    return "0" + word;
  else if (word.length === 6)
    return "00" + word;
  else if (word.length === 5)
    return "000" + word;
  else if (word.length === 4)
    return "0000" + word;
  else if (word.length === 3)
    return "00000" + word;
  else if (word.length === 2)
    return "000000" + word;
  else if (word.length === 1)
    return "0000000" + word;
  else
    return word;
}
var zero8_1 = zero8;
function join32(msg, start, end, endian) {
  var len = end - start;
  assert(len % 4 === 0);
  var res = new Array(len / 4);
  for (var i = 0, k = start; i < res.length; i++, k += 4) {
    var w;
    if (endian === "big")
      w = msg[k] << 24 | msg[k + 1] << 16 | msg[k + 2] << 8 | msg[k + 3];
    else
      w = msg[k + 3] << 24 | msg[k + 2] << 16 | msg[k + 1] << 8 | msg[k];
    res[i] = w >>> 0;
  }
  return res;
}
var join32_1 = join32;
function split32(msg, endian) {
  var res = new Array(msg.length * 4);
  for (var i = 0, k = 0; i < msg.length; i++, k += 4) {
    var m = msg[i];
    if (endian === "big") {
      res[k] = m >>> 24;
      res[k + 1] = m >>> 16 & 255;
      res[k + 2] = m >>> 8 & 255;
      res[k + 3] = m & 255;
    } else {
      res[k + 3] = m >>> 24;
      res[k + 2] = m >>> 16 & 255;
      res[k + 1] = m >>> 8 & 255;
      res[k] = m & 255;
    }
  }
  return res;
}
var split32_1 = split32;
function rotr32(w, b) {
  return w >>> b | w << 32 - b;
}
var rotr32_1 = rotr32;
function rotl32(w, b) {
  return w << b | w >>> 32 - b;
}
var rotl32_1 = rotl32;
function sum32(a, b) {
  return a + b >>> 0;
}
var sum32_1 = sum32;
function sum32_3(a, b, c) {
  return a + b + c >>> 0;
}
var sum32_3_1 = sum32_3;
function sum32_4(a, b, c, d) {
  return a + b + c + d >>> 0;
}
var sum32_4_1 = sum32_4;
function sum32_5(a, b, c, d, e) {
  return a + b + c + d + e >>> 0;
}
var sum32_5_1 = sum32_5;
function sum64(buf, pos, ah, al) {
  var bh = buf[pos];
  var bl = buf[pos + 1];
  var lo = al + bl >>> 0;
  var hi = (lo < al ? 1 : 0) + ah + bh;
  buf[pos] = hi >>> 0;
  buf[pos + 1] = lo;
}
var sum64_1 = sum64;
function sum64_hi(ah, al, bh, bl) {
  var lo = al + bl >>> 0;
  var hi = (lo < al ? 1 : 0) + ah + bh;
  return hi >>> 0;
}
var sum64_hi_1 = sum64_hi;
function sum64_lo(ah, al, bh, bl) {
  var lo = al + bl;
  return lo >>> 0;
}
var sum64_lo_1 = sum64_lo;
function sum64_4_hi(ah, al, bh, bl, ch, cl, dh, dl) {
  var carry = 0;
  var lo = al;
  lo = lo + bl >>> 0;
  carry += lo < al ? 1 : 0;
  lo = lo + cl >>> 0;
  carry += lo < cl ? 1 : 0;
  lo = lo + dl >>> 0;
  carry += lo < dl ? 1 : 0;
  var hi = ah + bh + ch + dh + carry;
  return hi >>> 0;
}
var sum64_4_hi_1 = sum64_4_hi;
function sum64_4_lo(ah, al, bh, bl, ch, cl, dh, dl) {
  var lo = al + bl + cl + dl;
  return lo >>> 0;
}
var sum64_4_lo_1 = sum64_4_lo;
function sum64_5_hi(ah, al, bh, bl, ch, cl, dh, dl, eh, el) {
  var carry = 0;
  var lo = al;
  lo = lo + bl >>> 0;
  carry += lo < al ? 1 : 0;
  lo = lo + cl >>> 0;
  carry += lo < cl ? 1 : 0;
  lo = lo + dl >>> 0;
  carry += lo < dl ? 1 : 0;
  lo = lo + el >>> 0;
  carry += lo < el ? 1 : 0;
  var hi = ah + bh + ch + dh + eh + carry;
  return hi >>> 0;
}
var sum64_5_hi_1 = sum64_5_hi;
function sum64_5_lo(ah, al, bh, bl, ch, cl, dh, dl, eh, el) {
  var lo = al + bl + cl + dl + el;
  return lo >>> 0;
}
var sum64_5_lo_1 = sum64_5_lo;
function rotr64_hi(ah, al, num) {
  var r2 = al << 32 - num | ah >>> num;
  return r2 >>> 0;
}
var rotr64_hi_1 = rotr64_hi;
function rotr64_lo(ah, al, num) {
  var r2 = ah << 32 - num | al >>> num;
  return r2 >>> 0;
}
var rotr64_lo_1 = rotr64_lo;
function shr64_hi(ah, al, num) {
  return ah >>> num;
}
var shr64_hi_1 = shr64_hi;
function shr64_lo(ah, al, num) {
  var r2 = ah << 32 - num | al >>> num;
  return r2 >>> 0;
}
var shr64_lo_1 = shr64_lo;
var utils = {
  inherits: inherits_1,
  toArray: toArray_1,
  toHex: toHex_1,
  htonl: htonl_1,
  toHex32: toHex32_1,
  zero2: zero2_1,
  zero8: zero8_1,
  join32: join32_1,
  split32: split32_1,
  rotr32: rotr32_1,
  rotl32: rotl32_1,
  sum32: sum32_1,
  sum32_3: sum32_3_1,
  sum32_4: sum32_4_1,
  sum32_5: sum32_5_1,
  sum64: sum64_1,
  sum64_hi: sum64_hi_1,
  sum64_lo: sum64_lo_1,
  sum64_4_hi: sum64_4_hi_1,
  sum64_4_lo: sum64_4_lo_1,
  sum64_5_hi: sum64_5_hi_1,
  sum64_5_lo: sum64_5_lo_1,
  rotr64_hi: rotr64_hi_1,
  rotr64_lo: rotr64_lo_1,
  shr64_hi: shr64_hi_1,
  shr64_lo: shr64_lo_1
};
function BlockHash() {
  this.pending = null;
  this.pendingTotal = 0;
  this.blockSize = this.constructor.blockSize;
  this.outSize = this.constructor.outSize;
  this.hmacStrength = this.constructor.hmacStrength;
  this.padLength = this.constructor.padLength / 8;
  this.endian = "big";
  this._delta8 = this.blockSize / 8;
  this._delta32 = this.blockSize / 32;
}
var BlockHash_1 = BlockHash;
BlockHash.prototype.update = function update(msg, enc) {
  msg = utils.toArray(msg, enc);
  if (!this.pending)
    this.pending = msg;
  else
    this.pending = this.pending.concat(msg);
  this.pendingTotal += msg.length;
  if (this.pending.length >= this._delta8) {
    msg = this.pending;
    var r2 = msg.length % this._delta8;
    this.pending = msg.slice(msg.length - r2, msg.length);
    if (this.pending.length === 0)
      this.pending = null;
    msg = utils.join32(msg, 0, msg.length - r2, this.endian);
    for (var i = 0; i < msg.length; i += this._delta32)
      this._update(msg, i, i + this._delta32);
  }
  return this;
};
BlockHash.prototype.digest = function digest(enc) {
  this.update(this._pad());
  assert(this.pending === null);
  return this._digest(enc);
};
BlockHash.prototype._pad = function pad() {
  var len = this.pendingTotal;
  var bytes = this._delta8;
  var k = bytes - (len + this.padLength) % bytes;
  var res = new Array(k + this.padLength);
  res[0] = 128;
  for (var i = 1; i < k; i++)
    res[i] = 0;
  len <<= 3;
  if (this.endian === "big") {
    for (var t = 8; t < this.padLength; t++)
      res[i++] = 0;
    res[i++] = 0;
    res[i++] = 0;
    res[i++] = 0;
    res[i++] = 0;
    res[i++] = len >>> 24 & 255;
    res[i++] = len >>> 16 & 255;
    res[i++] = len >>> 8 & 255;
    res[i++] = len & 255;
  } else {
    res[i++] = len & 255;
    res[i++] = len >>> 8 & 255;
    res[i++] = len >>> 16 & 255;
    res[i++] = len >>> 24 & 255;
    res[i++] = 0;
    res[i++] = 0;
    res[i++] = 0;
    res[i++] = 0;
    for (t = 8; t < this.padLength; t++)
      res[i++] = 0;
  }
  return res;
};
var common = {
  BlockHash: BlockHash_1
};
var rotr32$1 = utils.rotr32;
function ft_1(s2, x, y, z) {
  if (s2 === 0)
    return ch32(x, y, z);
  if (s2 === 1 || s2 === 3)
    return p32(x, y, z);
  if (s2 === 2)
    return maj32(x, y, z);
}
var ft_1_1 = ft_1;
function ch32(x, y, z) {
  return x & y ^ ~x & z;
}
var ch32_1 = ch32;
function maj32(x, y, z) {
  return x & y ^ x & z ^ y & z;
}
var maj32_1 = maj32;
function p32(x, y, z) {
  return x ^ y ^ z;
}
var p32_1 = p32;
function s0_256(x) {
  return rotr32$1(x, 2) ^ rotr32$1(x, 13) ^ rotr32$1(x, 22);
}
var s0_256_1 = s0_256;
function s1_256(x) {
  return rotr32$1(x, 6) ^ rotr32$1(x, 11) ^ rotr32$1(x, 25);
}
var s1_256_1 = s1_256;
function g0_256(x) {
  return rotr32$1(x, 7) ^ rotr32$1(x, 18) ^ x >>> 3;
}
var g0_256_1 = g0_256;
function g1_256(x) {
  return rotr32$1(x, 17) ^ rotr32$1(x, 19) ^ x >>> 10;
}
var g1_256_1 = g1_256;
var common$1 = {
  ft_1: ft_1_1,
  ch32: ch32_1,
  maj32: maj32_1,
  p32: p32_1,
  s0_256: s0_256_1,
  s1_256: s1_256_1,
  g0_256: g0_256_1,
  g1_256: g1_256_1
};
var rotl32$1 = utils.rotl32;
var sum32$1 = utils.sum32;
var sum32_5$1 = utils.sum32_5;
var ft_1$1 = common$1.ft_1;
var BlockHash$1 = common.BlockHash;
var sha1_K = [
  1518500249,
  1859775393,
  2400959708,
  3395469782
];
function SHA1() {
  if (!(this instanceof SHA1))
    return new SHA1();
  BlockHash$1.call(this);
  this.h = [
    1732584193,
    4023233417,
    2562383102,
    271733878,
    3285377520
  ];
  this.W = new Array(80);
}
utils.inherits(SHA1, BlockHash$1);
var _1 = SHA1;
SHA1.blockSize = 512;
SHA1.outSize = 160;
SHA1.hmacStrength = 80;
SHA1.padLength = 64;
SHA1.prototype._update = function _update(msg, start) {
  var W = this.W;
  for (var i = 0; i < 16; i++)
    W[i] = msg[start + i];
  for (; i < W.length; i++)
    W[i] = rotl32$1(W[i - 3] ^ W[i - 8] ^ W[i - 14] ^ W[i - 16], 1);
  var a = this.h[0];
  var b = this.h[1];
  var c = this.h[2];
  var d = this.h[3];
  var e = this.h[4];
  for (i = 0; i < W.length; i++) {
    var s2 = ~~(i / 20);
    var t = sum32_5$1(rotl32$1(a, 5), ft_1$1(s2, b, c, d), e, W[i], sha1_K[s2]);
    e = d;
    d = c;
    c = rotl32$1(b, 30);
    b = a;
    a = t;
  }
  this.h[0] = sum32$1(this.h[0], a);
  this.h[1] = sum32$1(this.h[1], b);
  this.h[2] = sum32$1(this.h[2], c);
  this.h[3] = sum32$1(this.h[3], d);
  this.h[4] = sum32$1(this.h[4], e);
};
SHA1.prototype._digest = function digest2(enc) {
  if (enc === "hex")
    return utils.toHex32(this.h, "big");
  else
    return utils.split32(this.h, "big");
};
var sum32$2 = utils.sum32;
var sum32_4$1 = utils.sum32_4;
var sum32_5$2 = utils.sum32_5;
var ch32$1 = common$1.ch32;
var maj32$1 = common$1.maj32;
var s0_256$1 = common$1.s0_256;
var s1_256$1 = common$1.s1_256;
var g0_256$1 = common$1.g0_256;
var g1_256$1 = common$1.g1_256;
var BlockHash$2 = common.BlockHash;
var sha256_K = [
  1116352408,
  1899447441,
  3049323471,
  3921009573,
  961987163,
  1508970993,
  2453635748,
  2870763221,
  3624381080,
  310598401,
  607225278,
  1426881987,
  1925078388,
  2162078206,
  2614888103,
  3248222580,
  3835390401,
  4022224774,
  264347078,
  604807628,
  770255983,
  1249150122,
  1555081692,
  1996064986,
  2554220882,
  2821834349,
  2952996808,
  3210313671,
  3336571891,
  3584528711,
  113926993,
  338241895,
  666307205,
  773529912,
  1294757372,
  1396182291,
  1695183700,
  1986661051,
  2177026350,
  2456956037,
  2730485921,
  2820302411,
  3259730800,
  3345764771,
  3516065817,
  3600352804,
  4094571909,
  275423344,
  430227734,
  506948616,
  659060556,
  883997877,
  958139571,
  1322822218,
  1537002063,
  1747873779,
  1955562222,
  2024104815,
  2227730452,
  2361852424,
  2428436474,
  2756734187,
  3204031479,
  3329325298
];
function SHA256() {
  if (!(this instanceof SHA256))
    return new SHA256();
  BlockHash$2.call(this);
  this.h = [
    1779033703,
    3144134277,
    1013904242,
    2773480762,
    1359893119,
    2600822924,
    528734635,
    1541459225
  ];
  this.k = sha256_K;
  this.W = new Array(64);
}
utils.inherits(SHA256, BlockHash$2);
var _256 = SHA256;
SHA256.blockSize = 512;
SHA256.outSize = 256;
SHA256.hmacStrength = 192;
SHA256.padLength = 64;
SHA256.prototype._update = function _update2(msg, start) {
  var W = this.W;
  for (var i = 0; i < 16; i++)
    W[i] = msg[start + i];
  for (; i < W.length; i++)
    W[i] = sum32_4$1(g1_256$1(W[i - 2]), W[i - 7], g0_256$1(W[i - 15]), W[i - 16]);
  var a = this.h[0];
  var b = this.h[1];
  var c = this.h[2];
  var d = this.h[3];
  var e = this.h[4];
  var f2 = this.h[5];
  var g = this.h[6];
  var h = this.h[7];
  assert(this.k.length === W.length);
  for (i = 0; i < W.length; i++) {
    var T1 = sum32_5$2(h, s1_256$1(e), ch32$1(e, f2, g), this.k[i], W[i]);
    var T2 = sum32$2(s0_256$1(a), maj32$1(a, b, c));
    h = g;
    g = f2;
    f2 = e;
    e = sum32$2(d, T1);
    d = c;
    c = b;
    b = a;
    a = sum32$2(T1, T2);
  }
  this.h[0] = sum32$2(this.h[0], a);
  this.h[1] = sum32$2(this.h[1], b);
  this.h[2] = sum32$2(this.h[2], c);
  this.h[3] = sum32$2(this.h[3], d);
  this.h[4] = sum32$2(this.h[4], e);
  this.h[5] = sum32$2(this.h[5], f2);
  this.h[6] = sum32$2(this.h[6], g);
  this.h[7] = sum32$2(this.h[7], h);
};
SHA256.prototype._digest = function digest3(enc) {
  if (enc === "hex")
    return utils.toHex32(this.h, "big");
  else
    return utils.split32(this.h, "big");
};
function SHA224() {
  if (!(this instanceof SHA224))
    return new SHA224();
  _256.call(this);
  this.h = [
    3238371032,
    914150663,
    812702999,
    4144912697,
    4290775857,
    1750603025,
    1694076839,
    3204075428
  ];
}
utils.inherits(SHA224, _256);
var _224 = SHA224;
SHA224.blockSize = 512;
SHA224.outSize = 224;
SHA224.hmacStrength = 192;
SHA224.padLength = 64;
SHA224.prototype._digest = function digest4(enc) {
  if (enc === "hex")
    return utils.toHex32(this.h.slice(0, 7), "big");
  else
    return utils.split32(this.h.slice(0, 7), "big");
};
var rotr64_hi$1 = utils.rotr64_hi;
var rotr64_lo$1 = utils.rotr64_lo;
var shr64_hi$1 = utils.shr64_hi;
var shr64_lo$1 = utils.shr64_lo;
var sum64$1 = utils.sum64;
var sum64_hi$1 = utils.sum64_hi;
var sum64_lo$1 = utils.sum64_lo;
var sum64_4_hi$1 = utils.sum64_4_hi;
var sum64_4_lo$1 = utils.sum64_4_lo;
var sum64_5_hi$1 = utils.sum64_5_hi;
var sum64_5_lo$1 = utils.sum64_5_lo;
var BlockHash$3 = common.BlockHash;
var sha512_K = [
  1116352408,
  3609767458,
  1899447441,
  602891725,
  3049323471,
  3964484399,
  3921009573,
  2173295548,
  961987163,
  4081628472,
  1508970993,
  3053834265,
  2453635748,
  2937671579,
  2870763221,
  3664609560,
  3624381080,
  2734883394,
  310598401,
  1164996542,
  607225278,
  1323610764,
  1426881987,
  3590304994,
  1925078388,
  4068182383,
  2162078206,
  991336113,
  2614888103,
  633803317,
  3248222580,
  3479774868,
  3835390401,
  2666613458,
  4022224774,
  944711139,
  264347078,
  2341262773,
  604807628,
  2007800933,
  770255983,
  1495990901,
  1249150122,
  1856431235,
  1555081692,
  3175218132,
  1996064986,
  2198950837,
  2554220882,
  3999719339,
  2821834349,
  766784016,
  2952996808,
  2566594879,
  3210313671,
  3203337956,
  3336571891,
  1034457026,
  3584528711,
  2466948901,
  113926993,
  3758326383,
  338241895,
  168717936,
  666307205,
  1188179964,
  773529912,
  1546045734,
  1294757372,
  1522805485,
  1396182291,
  2643833823,
  1695183700,
  2343527390,
  1986661051,
  1014477480,
  2177026350,
  1206759142,
  2456956037,
  344077627,
  2730485921,
  1290863460,
  2820302411,
  3158454273,
  3259730800,
  3505952657,
  3345764771,
  106217008,
  3516065817,
  3606008344,
  3600352804,
  1432725776,
  4094571909,
  1467031594,
  275423344,
  851169720,
  430227734,
  3100823752,
  506948616,
  1363258195,
  659060556,
  3750685593,
  883997877,
  3785050280,
  958139571,
  3318307427,
  1322822218,
  3812723403,
  1537002063,
  2003034995,
  1747873779,
  3602036899,
  1955562222,
  1575990012,
  2024104815,
  1125592928,
  2227730452,
  2716904306,
  2361852424,
  442776044,
  2428436474,
  593698344,
  2756734187,
  3733110249,
  3204031479,
  2999351573,
  3329325298,
  3815920427,
  3391569614,
  3928383900,
  3515267271,
  566280711,
  3940187606,
  3454069534,
  4118630271,
  4000239992,
  116418474,
  1914138554,
  174292421,
  2731055270,
  289380356,
  3203993006,
  460393269,
  320620315,
  685471733,
  587496836,
  852142971,
  1086792851,
  1017036298,
  365543100,
  1126000580,
  2618297676,
  1288033470,
  3409855158,
  1501505948,
  4234509866,
  1607167915,
  987167468,
  1816402316,
  1246189591
];
function SHA512() {
  if (!(this instanceof SHA512))
    return new SHA512();
  BlockHash$3.call(this);
  this.h = [
    1779033703,
    4089235720,
    3144134277,
    2227873595,
    1013904242,
    4271175723,
    2773480762,
    1595750129,
    1359893119,
    2917565137,
    2600822924,
    725511199,
    528734635,
    4215389547,
    1541459225,
    327033209
  ];
  this.k = sha512_K;
  this.W = new Array(160);
}
utils.inherits(SHA512, BlockHash$3);
var _512 = SHA512;
SHA512.blockSize = 1024;
SHA512.outSize = 512;
SHA512.hmacStrength = 192;
SHA512.padLength = 128;
SHA512.prototype._prepareBlock = function _prepareBlock(msg, start) {
  var W = this.W;
  for (var i = 0; i < 32; i++)
    W[i] = msg[start + i];
  for (; i < W.length; i += 2) {
    var c0_hi = g1_512_hi(W[i - 4], W[i - 3]);
    var c0_lo = g1_512_lo(W[i - 4], W[i - 3]);
    var c1_hi = W[i - 14];
    var c1_lo = W[i - 13];
    var c2_hi = g0_512_hi(W[i - 30], W[i - 29]);
    var c2_lo = g0_512_lo(W[i - 30], W[i - 29]);
    var c3_hi = W[i - 32];
    var c3_lo = W[i - 31];
    W[i] = sum64_4_hi$1(c0_hi, c0_lo, c1_hi, c1_lo, c2_hi, c2_lo, c3_hi, c3_lo);
    W[i + 1] = sum64_4_lo$1(c0_hi, c0_lo, c1_hi, c1_lo, c2_hi, c2_lo, c3_hi, c3_lo);
  }
};
SHA512.prototype._update = function _update3(msg, start) {
  this._prepareBlock(msg, start);
  var W = this.W;
  var ah = this.h[0];
  var al = this.h[1];
  var bh = this.h[2];
  var bl = this.h[3];
  var ch = this.h[4];
  var cl = this.h[5];
  var dh = this.h[6];
  var dl = this.h[7];
  var eh = this.h[8];
  var el = this.h[9];
  var fh = this.h[10];
  var fl = this.h[11];
  var gh = this.h[12];
  var gl = this.h[13];
  var hh = this.h[14];
  var hl = this.h[15];
  assert(this.k.length === W.length);
  for (var i = 0; i < W.length; i += 2) {
    var c0_hi = hh;
    var c0_lo = hl;
    var c1_hi = s1_512_hi(eh, el);
    var c1_lo = s1_512_lo(eh, el);
    var c2_hi = ch64_hi(eh, el, fh, fl, gh);
    var c2_lo = ch64_lo(eh, el, fh, fl, gh, gl);
    var c3_hi = this.k[i];
    var c3_lo = this.k[i + 1];
    var c4_hi = W[i];
    var c4_lo = W[i + 1];
    var T1_hi = sum64_5_hi$1(c0_hi, c0_lo, c1_hi, c1_lo, c2_hi, c2_lo, c3_hi, c3_lo, c4_hi, c4_lo);
    var T1_lo = sum64_5_lo$1(c0_hi, c0_lo, c1_hi, c1_lo, c2_hi, c2_lo, c3_hi, c3_lo, c4_hi, c4_lo);
    c0_hi = s0_512_hi(ah, al);
    c0_lo = s0_512_lo(ah, al);
    c1_hi = maj64_hi(ah, al, bh, bl, ch);
    c1_lo = maj64_lo(ah, al, bh, bl, ch, cl);
    var T2_hi = sum64_hi$1(c0_hi, c0_lo, c1_hi, c1_lo);
    var T2_lo = sum64_lo$1(c0_hi, c0_lo, c1_hi, c1_lo);
    hh = gh;
    hl = gl;
    gh = fh;
    gl = fl;
    fh = eh;
    fl = el;
    eh = sum64_hi$1(dh, dl, T1_hi, T1_lo);
    el = sum64_lo$1(dl, dl, T1_hi, T1_lo);
    dh = ch;
    dl = cl;
    ch = bh;
    cl = bl;
    bh = ah;
    bl = al;
    ah = sum64_hi$1(T1_hi, T1_lo, T2_hi, T2_lo);
    al = sum64_lo$1(T1_hi, T1_lo, T2_hi, T2_lo);
  }
  sum64$1(this.h, 0, ah, al);
  sum64$1(this.h, 2, bh, bl);
  sum64$1(this.h, 4, ch, cl);
  sum64$1(this.h, 6, dh, dl);
  sum64$1(this.h, 8, eh, el);
  sum64$1(this.h, 10, fh, fl);
  sum64$1(this.h, 12, gh, gl);
  sum64$1(this.h, 14, hh, hl);
};
SHA512.prototype._digest = function digest5(enc) {
  if (enc === "hex")
    return utils.toHex32(this.h, "big");
  else
    return utils.split32(this.h, "big");
};
function ch64_hi(xh, xl, yh, yl, zh) {
  var r2 = xh & yh ^ ~xh & zh;
  if (r2 < 0)
    r2 += 4294967296;
  return r2;
}
function ch64_lo(xh, xl, yh, yl, zh, zl) {
  var r2 = xl & yl ^ ~xl & zl;
  if (r2 < 0)
    r2 += 4294967296;
  return r2;
}
function maj64_hi(xh, xl, yh, yl, zh) {
  var r2 = xh & yh ^ xh & zh ^ yh & zh;
  if (r2 < 0)
    r2 += 4294967296;
  return r2;
}
function maj64_lo(xh, xl, yh, yl, zh, zl) {
  var r2 = xl & yl ^ xl & zl ^ yl & zl;
  if (r2 < 0)
    r2 += 4294967296;
  return r2;
}
function s0_512_hi(xh, xl) {
  var c0_hi = rotr64_hi$1(xh, xl, 28);
  var c1_hi = rotr64_hi$1(xl, xh, 2);
  var c2_hi = rotr64_hi$1(xl, xh, 7);
  var r2 = c0_hi ^ c1_hi ^ c2_hi;
  if (r2 < 0)
    r2 += 4294967296;
  return r2;
}
function s0_512_lo(xh, xl) {
  var c0_lo = rotr64_lo$1(xh, xl, 28);
  var c1_lo = rotr64_lo$1(xl, xh, 2);
  var c2_lo = rotr64_lo$1(xl, xh, 7);
  var r2 = c0_lo ^ c1_lo ^ c2_lo;
  if (r2 < 0)
    r2 += 4294967296;
  return r2;
}
function s1_512_hi(xh, xl) {
  var c0_hi = rotr64_hi$1(xh, xl, 14);
  var c1_hi = rotr64_hi$1(xh, xl, 18);
  var c2_hi = rotr64_hi$1(xl, xh, 9);
  var r2 = c0_hi ^ c1_hi ^ c2_hi;
  if (r2 < 0)
    r2 += 4294967296;
  return r2;
}
function s1_512_lo(xh, xl) {
  var c0_lo = rotr64_lo$1(xh, xl, 14);
  var c1_lo = rotr64_lo$1(xh, xl, 18);
  var c2_lo = rotr64_lo$1(xl, xh, 9);
  var r2 = c0_lo ^ c1_lo ^ c2_lo;
  if (r2 < 0)
    r2 += 4294967296;
  return r2;
}
function g0_512_hi(xh, xl) {
  var c0_hi = rotr64_hi$1(xh, xl, 1);
  var c1_hi = rotr64_hi$1(xh, xl, 8);
  var c2_hi = shr64_hi$1(xh, xl, 7);
  var r2 = c0_hi ^ c1_hi ^ c2_hi;
  if (r2 < 0)
    r2 += 4294967296;
  return r2;
}
function g0_512_lo(xh, xl) {
  var c0_lo = rotr64_lo$1(xh, xl, 1);
  var c1_lo = rotr64_lo$1(xh, xl, 8);
  var c2_lo = shr64_lo$1(xh, xl, 7);
  var r2 = c0_lo ^ c1_lo ^ c2_lo;
  if (r2 < 0)
    r2 += 4294967296;
  return r2;
}
function g1_512_hi(xh, xl) {
  var c0_hi = rotr64_hi$1(xh, xl, 19);
  var c1_hi = rotr64_hi$1(xl, xh, 29);
  var c2_hi = shr64_hi$1(xh, xl, 6);
  var r2 = c0_hi ^ c1_hi ^ c2_hi;
  if (r2 < 0)
    r2 += 4294967296;
  return r2;
}
function g1_512_lo(xh, xl) {
  var c0_lo = rotr64_lo$1(xh, xl, 19);
  var c1_lo = rotr64_lo$1(xl, xh, 29);
  var c2_lo = shr64_lo$1(xh, xl, 6);
  var r2 = c0_lo ^ c1_lo ^ c2_lo;
  if (r2 < 0)
    r2 += 4294967296;
  return r2;
}
function SHA384() {
  if (!(this instanceof SHA384))
    return new SHA384();
  _512.call(this);
  this.h = [
    3418070365,
    3238371032,
    1654270250,
    914150663,
    2438529370,
    812702999,
    355462360,
    4144912697,
    1731405415,
    4290775857,
    2394180231,
    1750603025,
    3675008525,
    1694076839,
    1203062813,
    3204075428
  ];
}
utils.inherits(SHA384, _512);
var _384 = SHA384;
SHA384.blockSize = 1024;
SHA384.outSize = 384;
SHA384.hmacStrength = 192;
SHA384.padLength = 128;
SHA384.prototype._digest = function digest6(enc) {
  if (enc === "hex")
    return utils.toHex32(this.h.slice(0, 12), "big");
  else
    return utils.split32(this.h.slice(0, 12), "big");
};
var sha1 = _1;
var sha224 = _224;
var sha256 = _256;
var sha384 = _384;
var sha512 = _512;
var sha = {
  sha1,
  sha224,
  sha256,
  sha384,
  sha512
};
var rotl32$2 = utils.rotl32;
var sum32$3 = utils.sum32;
var sum32_3$1 = utils.sum32_3;
var sum32_4$2 = utils.sum32_4;
var BlockHash$4 = common.BlockHash;
function RIPEMD160() {
  if (!(this instanceof RIPEMD160))
    return new RIPEMD160();
  BlockHash$4.call(this);
  this.h = [1732584193, 4023233417, 2562383102, 271733878, 3285377520];
  this.endian = "little";
}
utils.inherits(RIPEMD160, BlockHash$4);
var ripemd160 = RIPEMD160;
RIPEMD160.blockSize = 512;
RIPEMD160.outSize = 160;
RIPEMD160.hmacStrength = 192;
RIPEMD160.padLength = 64;
RIPEMD160.prototype._update = function update2(msg, start) {
  var A = this.h[0];
  var B = this.h[1];
  var C = this.h[2];
  var D = this.h[3];
  var E = this.h[4];
  var Ah = A;
  var Bh = B;
  var Ch = C;
  var Dh = D;
  var Eh = E;
  for (var j = 0; j < 80; j++) {
    var T = sum32$3(rotl32$2(sum32_4$2(A, f(j, B, C, D), msg[r[j] + start], K(j)), s[j]), E);
    A = E;
    E = D;
    D = rotl32$2(C, 10);
    C = B;
    B = T;
    T = sum32$3(rotl32$2(sum32_4$2(Ah, f(79 - j, Bh, Ch, Dh), msg[rh[j] + start], Kh(j)), sh[j]), Eh);
    Ah = Eh;
    Eh = Dh;
    Dh = rotl32$2(Ch, 10);
    Ch = Bh;
    Bh = T;
  }
  T = sum32_3$1(this.h[1], C, Dh);
  this.h[1] = sum32_3$1(this.h[2], D, Eh);
  this.h[2] = sum32_3$1(this.h[3], E, Ah);
  this.h[3] = sum32_3$1(this.h[4], A, Bh);
  this.h[4] = sum32_3$1(this.h[0], B, Ch);
  this.h[0] = T;
};
RIPEMD160.prototype._digest = function digest7(enc) {
  if (enc === "hex")
    return utils.toHex32(this.h, "little");
  else
    return utils.split32(this.h, "little");
};
function f(j, x, y, z) {
  if (j <= 15)
    return x ^ y ^ z;
  else if (j <= 31)
    return x & y | ~x & z;
  else if (j <= 47)
    return (x | ~y) ^ z;
  else if (j <= 63)
    return x & z | y & ~z;
  else
    return x ^ (y | ~z);
}
function K(j) {
  if (j <= 15)
    return 0;
  else if (j <= 31)
    return 1518500249;
  else if (j <= 47)
    return 1859775393;
  else if (j <= 63)
    return 2400959708;
  else
    return 2840853838;
}
function Kh(j) {
  if (j <= 15)
    return 1352829926;
  else if (j <= 31)
    return 1548603684;
  else if (j <= 47)
    return 1836072691;
  else if (j <= 63)
    return 2053994217;
  else
    return 0;
}
var r = [
  0,
  1,
  2,
  3,
  4,
  5,
  6,
  7,
  8,
  9,
  10,
  11,
  12,
  13,
  14,
  15,
  7,
  4,
  13,
  1,
  10,
  6,
  15,
  3,
  12,
  0,
  9,
  5,
  2,
  14,
  11,
  8,
  3,
  10,
  14,
  4,
  9,
  15,
  8,
  1,
  2,
  7,
  0,
  6,
  13,
  11,
  5,
  12,
  1,
  9,
  11,
  10,
  0,
  8,
  12,
  4,
  13,
  3,
  7,
  15,
  14,
  5,
  6,
  2,
  4,
  0,
  5,
  9,
  7,
  12,
  2,
  10,
  14,
  1,
  3,
  8,
  11,
  6,
  15,
  13
];
var rh = [
  5,
  14,
  7,
  0,
  9,
  2,
  11,
  4,
  13,
  6,
  15,
  8,
  1,
  10,
  3,
  12,
  6,
  11,
  3,
  7,
  0,
  13,
  5,
  10,
  14,
  15,
  8,
  12,
  4,
  9,
  1,
  2,
  15,
  5,
  1,
  3,
  7,
  14,
  6,
  9,
  11,
  8,
  12,
  2,
  10,
  0,
  4,
  13,
  8,
  6,
  4,
  1,
  3,
  11,
  15,
  0,
  5,
  12,
  2,
  13,
  9,
  7,
  10,
  14,
  12,
  15,
  10,
  4,
  1,
  5,
  8,
  7,
  6,
  2,
  13,
  14,
  0,
  3,
  9,
  11
];
var s = [
  11,
  14,
  15,
  12,
  5,
  8,
  7,
  9,
  11,
  13,
  14,
  15,
  6,
  7,
  9,
  8,
  7,
  6,
  8,
  13,
  11,
  9,
  7,
  15,
  7,
  12,
  15,
  9,
  11,
  7,
  13,
  12,
  11,
  13,
  6,
  7,
  14,
  9,
  13,
  15,
  14,
  8,
  13,
  6,
  5,
  12,
  7,
  5,
  11,
  12,
  14,
  15,
  14,
  15,
  9,
  8,
  9,
  14,
  5,
  6,
  8,
  6,
  5,
  12,
  9,
  15,
  5,
  11,
  6,
  8,
  13,
  12,
  5,
  12,
  13,
  14,
  11,
  8,
  5,
  6
];
var sh = [
  8,
  9,
  9,
  11,
  13,
  15,
  15,
  5,
  7,
  7,
  8,
  11,
  14,
  14,
  12,
  6,
  9,
  13,
  15,
  7,
  12,
  8,
  9,
  11,
  7,
  7,
  12,
  7,
  6,
  15,
  13,
  11,
  9,
  7,
  15,
  11,
  8,
  6,
  6,
  14,
  12,
  13,
  5,
  14,
  13,
  13,
  7,
  5,
  15,
  5,
  8,
  11,
  14,
  14,
  6,
  14,
  6,
  9,
  12,
  9,
  12,
  5,
  15,
  8,
  8,
  5,
  12,
  9,
  12,
  5,
  14,
  6,
  8,
  13,
  6,
  5,
  15,
  13,
  11,
  11
];
var ripemd = {
  ripemd160
};
function Hmac(hash, key, enc) {
  if (!(this instanceof Hmac))
    return new Hmac(hash, key, enc);
  this.Hash = hash;
  this.blockSize = hash.blockSize / 8;
  this.outSize = hash.outSize / 8;
  this.inner = null;
  this.outer = null;
  this._init(utils.toArray(key, enc));
}
var hmac = Hmac;
Hmac.prototype._init = function init(key) {
  if (key.length > this.blockSize)
    key = new this.Hash().update(key).digest();
  assert(key.length <= this.blockSize);
  for (var i = key.length; i < this.blockSize; i++)
    key.push(0);
  for (i = 0; i < key.length; i++)
    key[i] ^= 54;
  this.inner = new this.Hash().update(key);
  for (i = 0; i < key.length; i++)
    key[i] ^= 106;
  this.outer = new this.Hash().update(key);
};
Hmac.prototype.update = function update3(msg, enc) {
  this.inner.update(msg, enc);
  return this;
};
Hmac.prototype.digest = function digest8(enc) {
  this.outer.update(this.inner.digest());
  return this.outer.digest(enc);
};
var hash_1 = createCommonjsModule(function(module, exports) {
  var hash = exports;
  hash.utils = utils;
  hash.common = common;
  hash.sha = sha;
  hash.ripemd = ripemd;
  hash.hmac = hmac;
  hash.sha1 = hash.sha.sha1;
  hash.sha256 = hash.sha.sha256;
  hash.sha224 = hash.sha.sha224;
  hash.sha384 = hash.sha.sha384;
  hash.sha512 = hash.sha.sha512;
  hash.ripemd160 = hash.ripemd.ripemd160;
});
export default hash_1;
