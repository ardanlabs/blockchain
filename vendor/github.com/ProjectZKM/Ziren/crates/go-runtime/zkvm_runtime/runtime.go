//go:build mipsle
// +build mipsle

package zkvm_runtime

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
	"reflect"
	"unsafe"
)

func SyscallWrite(fd int, write_buf []byte, nbytes int) int
func SyscallHintLen() int
func SyscallHintRead(ptr []byte, len int)
func SyscallCommit(index int, word uint32)
func SyscallExit(code int)
func SyscallKeccakSponge(input unsafe.Pointer, result unsafe.Pointer)

var PublicValuesHasher hash.Hash = sha256.New()

const EMBEDDED_RESERVED_INPUT_REGION_SIZE int = 1024 * 1024 * 1024
const MAX_MEMORY int = 0x7f000000

var RESERVED_INPUT_PTR int = MAX_MEMORY - EMBEDDED_RESERVED_INPUT_REGION_SIZE

func Read[T any]() T {
	len := SyscallHintLen()
	var value []byte
	capacity := (len + 3) / 4 * 4
	addr := RESERVED_INPUT_PTR
	RESERVED_INPUT_PTR += capacity
	ptr := unsafe.Pointer(uintptr(addr))
	value = unsafe.Slice((*byte)(ptr), capacity)
	var result T
	SyscallHintRead(value, len)
	DeserializeData(value[0:len], &result)
	return result
}

func Commit[T any](value T) {
	bytes := MustSerializeData(value)
	length := len(bytes)
	if (length & 3) != 0 {
		d := make([]byte, 4-(length&3))
		bytes = append(bytes, d...)
	}

	_, _ = PublicValuesHasher.Write(bytes)

	SyscallWrite(13, bytes, length)
}

//go:linkname RuntimeExit zkvm.RuntimeExit
func RuntimeExit(code int) {
	hashBytes := PublicValuesHasher.Sum(nil)

	// 2. COMMIT each u32 word
	for i := 0; i < 8; i++ {
		word := binary.LittleEndian.Uint32(hashBytes[i*4 : (i+1)*4])
		SyscallCommit(i, word)
	}

	SyscallExit(code)
}

func Keccak256(data []byte) [32]byte {
	var result [32]byte
	length := len(data)

	if length == 0 {
		return [32]byte{
			0xC5, 0xD2, 0x46, 0x01, 0x86, 0xF7, 0x23, 0x3C, 0x92, 0x7E, 0x7D, 0xB2, 0xDC, 0xC7,
			0x03, 0xC0, 0xE5, 0, 0xB6, 0x53, 0xCA, 0x82, 0x27, 0x3B, 0x7B, 0xFA, 0xD8, 0x04, 0x5D,
			0x85, 0xA4, 0x70,
		}
	}

	// Padding input to reach the required size
	finalBlockLen := length % 136
	paddedLen := length - finalBlockLen + 136

	paddedData := make([]byte, paddedLen)
	copy(paddedData, data)

	if length%136 == 135 {
		paddedData[paddedLen-1] = 0b10000001
	} else {
		paddedData[length] = 1
		paddedData[paddedLen-1] = 0b10000000
	}

	// Convert to u32 to align the memory
	u32Array := make([]uint32, 0, paddedLen/4+(paddedLen/136)*2)
	count := 0
	for i := 0; i < paddedLen; i += 4 {
		// Little-endian conversion
		u32Value := uint32(paddedData[i]) |
			uint32(paddedData[i+1])<<8 |
			uint32(paddedData[i+2])<<16 |
			uint32(paddedData[i+3])<<24
		u32Array = append(u32Array, u32Value)
		count++
		if count == 34 {
			// Add padding for sponge structure
			u32Array = append(u32Array, 0, 0)
			count = 0
		}
	}

	var generalResult [17]uint32
	generalResult[16] = uint32(len(u32Array))

	inputPtr := unsafe.Pointer(&u32Array[0])
	resultPtr := unsafe.Pointer(&generalResult[0])
	SyscallKeccakSponge(inputPtr, resultPtr)

	for i := 0; i < 8; i++ {
		val := generalResult[i]
		result[i*4] = byte(val)
		result[i*4+1] = byte(val >> 8)
		result[i*4+2] = byte(val >> 16)
		result[i*4+3] = byte(val >> 24)
	}

	return result
}

func init() {
	// Explicit reference, prevent optimization
	_ = reflect.ValueOf(RuntimeExit)
}
