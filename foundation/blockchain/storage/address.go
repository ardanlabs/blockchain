// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package storage

import "errors"

// Address represents an account address in the system.
type Address string

// ToAddress converts a hex-encoded string to an address and validates the
// hex-encoded string is formatted correctly.
func ToAddress(hex string) (Address, error) {
	a := Address(hex)
	if !a.IsAddress() {
		return "", errors.New("invalid address format")
	}

	return a, nil
}

// IsAddress verifies whether the specified string represents a valid
// hex-encoded address.
func (a Address) IsAddress() bool {
	const addressLength = 20

	if has0xPrefix(a) {
		a = a[2:]
	}
	return len(a) == 2*addressLength && isHex(a)
}

// =============================================================================

func has0xPrefix(a Address) bool {
	return len(a) >= 2 && a[0] == '0' && (a[1] == 'x' || a[1] == 'X')
}

// isHex validates whether each byte is valid hexadecimal string.
func isHex(a Address) bool {
	if len(a)%2 != 0 {
		return false
	}
	for _, c := range []byte(a) {
		if !isHexCharacter(c) {
			return false
		}
	}
	return true
}

// isHexCharacter returns bool of c being a valid hexadecimal.
func isHexCharacter(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}
