//go:build mipsle
// +build mipsle

package zkvm_runtime

import (
	"encoding/binary"
	"unsafe"
)

// Sha256 computes the SHA-256 hash of data using zkVM precompiles.
func Sha256(data []byte) [32]byte {
	// SHA-256 initial hash values
	state := [8]uint32{
		0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a,
		0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19,
	}

	// Pad message per SHA-256 spec
	msgLen := len(data)
	bitLen := uint64(msgLen) * 8

	// Append 0x80 byte
	data = append(data, 0x80)
	// Pad to 56 mod 64
	for len(data)%64 != 56 {
		data = append(data, 0x00)
	}
	// Append original length in bits as big-endian uint64
	var lenBuf [8]byte
	binary.BigEndian.PutUint64(lenBuf[:], bitLen)
	data = append(data, lenBuf[:]...)

	// Process each 64-byte block
	for offset := 0; offset < len(data); offset += 64 {
		var w [64]uint32

		// Load 16 words from block (big-endian as per SHA-256 spec)
		for i := 0; i < 16; i++ {
			w[i] = binary.BigEndian.Uint32(data[offset+i*4 : offset+i*4+4])
		}

		// Extend to 64 words using precompile
		SyscallSha256Extend(unsafe.Pointer(&w[0]))

		// Compress using precompile
		SyscallSha256Compress(unsafe.Pointer(&w[0]), unsafe.Pointer(&state[0]))
	}

	// Convert state to big-endian bytes
	var result [32]byte
	for i := 0; i < 8; i++ {
		binary.BigEndian.PutUint32(result[i*4:i*4+4], state[i])
	}
	return result
}

// Bn254G1Point represents a BN254 G1 affine point as [16]uint32 in little-endian.
type Bn254G1Point [16]uint32

// Bn254G1Add computes p = p + q on BN254 G1.
func Bn254G1Add(p, q *Bn254G1Point) {
	SyscallBn254Add(unsafe.Pointer(p), unsafe.Pointer(q))
}

// Bn254G1Double computes p = 2*p on BN254 G1.
func Bn254G1Double(p *Bn254G1Point) {
	SyscallBn254Double(unsafe.Pointer(p), unsafe.Pointer(nil))
}

// Bn254G1ScalarMul computes result = scalar * base using double-and-add.
// scalar is a big-endian 32-byte value.
func Bn254G1ScalarMul(base *Bn254G1Point, scalar []byte) Bn254G1Point {
	var result Bn254G1Point // identity (all zeros)
	temp := *base

	for i := len(scalar) - 1; i >= 0; i-- {
		b := scalar[i]
		for bit := 0; bit < 8; bit++ {
			if (b>>uint(bit))&1 == 1 {
				bn254AddSafe(&result, &temp)
			}
			Bn254G1Double(&temp)
		}
	}
	return result
}

func bn254AddSafe(p, q *Bn254G1Point) {
	if isZeroPoint(p[:]) {
		*p = *q
		return
	}
	if isZeroPoint(q[:]) {
		return
	}
	if *p == *q {
		Bn254G1Double(p)
		return
	}
	Bn254G1Add(p, q)
}

// Bls12381G1Point represents a BLS12-381 G1 affine point as [24]uint32 in little-endian.
type Bls12381G1Point [24]uint32

// Bls12381G1Add computes p = p + q on BLS12-381 G1.
func Bls12381G1Add(p, q *Bls12381G1Point) {
	SyscallBls12381Add(unsafe.Pointer(p), unsafe.Pointer(q))
}

// Bls12381G1Double computes p = 2*p on BLS12-381 G1.
func Bls12381G1Double(p *Bls12381G1Point) {
	SyscallBls12381Double(unsafe.Pointer(p), unsafe.Pointer(nil))
}

// Bls12381G1ScalarMul computes result = scalar * base using double-and-add.
// scalar is a big-endian 32-byte value.
func Bls12381G1ScalarMul(base *Bls12381G1Point, scalar []byte) Bls12381G1Point {
	var result Bls12381G1Point
	temp := *base

	for i := len(scalar) - 1; i >= 0; i-- {
		b := scalar[i]
		for bit := 0; bit < 8; bit++ {
			if (b>>uint(bit))&1 == 1 {
				bls12381AddSafe(&result, &temp)
			}
			Bls12381G1Double(&temp)
		}
	}
	return result
}

func bls12381AddSafe(p, q *Bls12381G1Point) {
	if isZeroPoint(p[:]) {
		*p = *q
		return
	}
	if isZeroPoint(q[:]) {
		return
	}
	if *p == *q {
		Bls12381G1Double(p)
		return
	}
	Bls12381G1Add(p, q)
}

func isZeroPoint(limbs []uint32) bool {
	for _, v := range limbs {
		if v != 0 {
			return false
		}
	}
	return true
}

// BeToLeU32 converts a big-endian byte slice to little-endian [N]uint32.
// Each 4 bytes of big-endian input becomes one uint32, stored in reverse order.
func BeToLeU32(be []byte, limbs []uint32) {
	n := len(limbs)
	for i := 0; i < n; i++ {
		// Read from the end of BE bytes
		off := len(be) - (i+1)*4
		if off >= 0 {
			limbs[i] = binary.LittleEndian.Uint32([]byte{be[off+3], be[off+2], be[off+1], be[off]})
		}
	}
}

// LeU32ToBe converts little-endian [N]uint32 back to big-endian bytes.
func LeU32ToBe(limbs []uint32, be []byte) {
	n := len(limbs)
	for i := range be {
		be[i] = 0
	}
	for i := 0; i < n; i++ {
		off := len(be) - (i+1)*4
		if off >= 0 {
			v := limbs[i]
			be[off] = byte(v >> 24)
			be[off+1] = byte(v >> 16)
			be[off+2] = byte(v >> 8)
			be[off+3] = byte(v)
		}
	}
}
