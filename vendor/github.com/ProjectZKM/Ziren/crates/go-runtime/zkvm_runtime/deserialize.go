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

func deserializeData(data []byte, v reflect.Value, index int) (int, error) {
	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(data[index] == 1)
		return index + 1, nil
	case reflect.Int8:
		v.SetInt(int64(int8(data[index])))
		return index + 1, nil
	case reflect.Uint8:
		v.SetUint(uint64(data[index]))
		return index + 1, nil
	case reflect.Int16:
		b := []byte{data[index], data[index+1]}
		a := binary.LittleEndian.Uint16(b)
		v.SetInt(int64(int16(a)))
		return index + 2, nil
	case reflect.Uint16:
		b := []byte{data[index], data[index+1]}
		a := binary.LittleEndian.Uint16(b)
		v.SetUint(uint64(a))
		return index + 2, nil
	case reflect.Int32:
		b := []byte{data[index], data[index+1], data[index+2], data[index+3]}
		a := binary.LittleEndian.Uint32(b)
		v.SetInt(int64(int32(a)))
		return index + 4, nil
	case reflect.Uint32:
		b := []byte{data[index], data[index+1], data[index+2], data[index+3]}
		a := binary.LittleEndian.Uint32(b)
		v.SetUint(uint64(a))
		return index + 4, nil
	case reflect.Int64:
		b := []byte{data[index], data[index+1], data[index+2], data[index+3],
			data[index+4], data[index+5], data[index+6], data[index+7]}
		a := binary.LittleEndian.Uint64(b)
		v.SetInt(int64(a))
		return index + 8, nil
	case reflect.Uint64:
		b := []byte{data[index], data[index+1], data[index+2], data[index+3],
			data[index+4], data[index+5], data[index+6], data[index+7]}
		a := binary.LittleEndian.Uint64(b)
		v.SetUint(a)
		return index + 8, nil
	case reflect.Slice:
		b := []byte{data[index], data[index+1], data[index+2], data[index+3],
			data[index+4], data[index+5], data[index+6], data[index+7]}

		length := binary.LittleEndian.Uint64(b)
		index += 8
		switch v.Type().Elem().Kind() {
		case reflect.Uint8:
			bytes := data[index : index+int(length)]
			v.SetBytes(bytes)
			return index + int(length), nil
		}
		return index, fmt.Errorf("unsupport type: %v, elem: %v", v.Kind(), v.Elem().Kind())
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
		b := []byte{data[index], data[index+1], data[index+2], data[index+3],
			data[index+4], data[index+5], data[index+6], data[index+7]}
		l := binary.LittleEndian.Uint64(b)
		index += 8
		length := int(l)
		str := make([]byte, length)
		copy(str[:], data[index:index+length])
		v.SetString(string(str))
		return index + length, nil
	case reflect.Ptr:
		if data[index] == 0 {
			v.SetZero()
			return index + 1, nil
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
	return index, fmt.Errorf("unsupport type: %v", v.Kind())
}
