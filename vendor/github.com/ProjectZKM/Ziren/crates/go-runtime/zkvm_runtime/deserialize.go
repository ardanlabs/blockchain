package zkvm_runtime

import (
	"encoding/binary"
	"fmt"
	"reflect"
)

func DeserializeData(data []byte, e any) {
	if e == nil {
		return
	}
	value := reflect.ValueOf(e)
	// If e represents a value as opposed to a pointer, the answer won't
	// get back to the caller. Make sure it's a pointer.
	if value.Type().Kind() != reflect.Pointer {
		panic("attempt to deserialize into a non-pointer")
	}

	if value.IsValid() {
		if value.Kind() == reflect.Pointer && !value.IsNil() {
			// That's okay, we'll store through the pointer.
		} else if !value.CanSet() {
			panic("gob: DecodeValue of unassignable value")
		}
	}

	index, err := deserializeData(data, value.Elem(), 0)
	if err != nil {
		panic(err)
	}
	if index != len(data) {
		panic("deserialize failed")
	}
}

func readBytes(data []byte, index, size int) ([]byte, error) {
	if index < 0 || size < 0 || index > len(data) || size > len(data)-index {
		return nil, fmt.Errorf("deserialize failed: need %d bytes at offset %d, have %d", size, index, len(data)-index)
	}
	return data[index : index+size], nil
}

func maxInt() int {
	return int(^uint(0) >> 1)
}

func deserializeData(data []byte, v reflect.Value, index int) (int, error) {
	switch v.Kind() {
	case reflect.Bool:
		b, err := readBytes(data, index, 1)
		if err != nil {
			return index, err
		}
		v.SetBool(b[0] == 1)
		return index + 1, nil
	case reflect.Int8:
		b, err := readBytes(data, index, 1)
		if err != nil {
			return index, err
		}
		v.SetInt(int64(int8(b[0])))
		return index + 1, nil
	case reflect.Uint8:
		b, err := readBytes(data, index, 1)
		if err != nil {
			return index, err
		}
		v.SetUint(uint64(b[0]))
		return index + 1, nil
	case reflect.Int16:
		b, err := readBytes(data, index, 2)
		if err != nil {
			return index, err
		}
		a := binary.LittleEndian.Uint16(b)
		v.SetInt(int64(int16(a)))
		return index + 2, nil
	case reflect.Uint16:
		b, err := readBytes(data, index, 2)
		if err != nil {
			return index, err
		}
		a := binary.LittleEndian.Uint16(b)
		v.SetUint(uint64(a))
		return index + 2, nil
	case reflect.Int32:
		b, err := readBytes(data, index, 4)
		if err != nil {
			return index, err
		}
		a := binary.LittleEndian.Uint32(b)
		v.SetInt(int64(int32(a)))
		return index + 4, nil
	case reflect.Uint32:
		b, err := readBytes(data, index, 4)
		if err != nil {
			return index, err
		}
		a := binary.LittleEndian.Uint32(b)
		v.SetUint(uint64(a))
		return index + 4, nil
	case reflect.Int64:
		b, err := readBytes(data, index, 8)
		if err != nil {
			return index, err
		}
		a := binary.LittleEndian.Uint64(b)
		v.SetInt(int64(a))
		return index + 8, nil
	case reflect.Uint64:
		b, err := readBytes(data, index, 8)
		if err != nil {
			return index, err
		}
		a := binary.LittleEndian.Uint64(b)
		v.SetUint(a)
		return index + 8, nil
	case reflect.Slice:
		const maxSliceLen = 1_000_000
		b, err := readBytes(data, index, 8)
		if err != nil {
			return index, err
		}

		length := binary.LittleEndian.Uint64(b)
		index += 8
		elemKind := v.Type().Elem().Kind()
		if length > uint64(maxInt()) {
			return index, fmt.Errorf("deserialize failed: slice length %d exceeds int range", length)
		}
		l := int(length)
		if elemKind == reflect.Uint8 {
			if l > len(data)-index {
				return index, fmt.Errorf("deserialize failed: []byte length %d exceeds remaining %d", l, len(data)-index)
			}
			v.SetBytes(data[index : index+l])
			return index + l, nil
		}

		if l > maxSliceLen {
			return index, fmt.Errorf("deserialize failed: slice length %d exceeds max %d", l, maxSliceLen)
		}
		slice := reflect.MakeSlice(v.Type(), l, l)
		for i := 0; i < l; i++ {
			var err error
			index, err = deserializeData(data, slice.Index(i), index)
			if err != nil {
				return index, err
			}
		}
		v.Set(slice)
		return index, nil
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			var err error
			index, err = deserializeData(data, v.Index(i), index)
			if err != nil {
				return index, err
			}
		}
		return index, nil
	case reflect.String:
		b, err := readBytes(data, index, 8)
		if err != nil {
			return index, err
		}
		l := binary.LittleEndian.Uint64(b)
		index += 8
		if l > uint64(maxInt()) {
			return index, fmt.Errorf("deserialize failed: string length %d exceeds int range", l)
		}
		length := int(l)
		if length > len(data)-index {
			return index, fmt.Errorf("deserialize failed: string length %d exceeds remaining %d", length, len(data)-index)
		}
		str := make([]byte, length)
		copy(str[:], data[index:index+length])
		v.SetString(string(str))
		return index + length, nil
	case reflect.Ptr:
		b, err := readBytes(data, index, 1)
		if err != nil {
			return index, err
		}
		if b[0] == 0 {
			v.SetZero()
			return index + 1, nil
		}
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return deserializeData(data, v.Elem(), index+1)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			var err error
			index, err = deserializeData(data, field, index)
			if err != nil {
				return index, err
			}
		}
		return index, nil
	}
	return index, fmt.Errorf("unsupported type: %v", v.Kind())
}
