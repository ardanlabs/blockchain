var minimalisticAssert = assert;
function assert(val, msg) {
  if (!val)
    throw new Error(msg || "Assertion failed");
}
assert.equal = function assertEqual(l, r, msg) {
  if (l != r)
    throw new Error(msg || "Assertion failed: " + l + " != " + r);
};
export default minimalisticAssert;
var equal = minimalisticAssert.equal;
export {minimalisticAssert as __moduleExports, equal};
