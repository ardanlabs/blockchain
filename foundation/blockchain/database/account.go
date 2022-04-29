package database

import (
	"crypto/ecdsa"
	"errors"

	"github.com/ethereum/go-ethereum/crypto"
)

// Account represents information stored in the database for an individual account.
type Account struct {
	AccountID AccountID
	Nonce     uint64
	Balance   uint64
}

// newAccount constructs a new account value for use.
func newAccount(accountID AccountID, balance uint64) Account {
	return Account{
		AccountID: accountID,
		Balance:   balance,
	}
}

// =============================================================================

// AccountID represents an account id that is used to sign transactions and is
// associated with transactions on the blockchain.
type AccountID string

// ToAccountID converts a hex-encoded string to an account and validates the
// hex-encoded string is formatted correctly.
func ToAccountID(hex string) (AccountID, error) {
	a := AccountID(hex)
	if !a.IsAccountID() {
		return "", errors.New("invalid account format")
	}

	return a, nil
}

// PublicKeyToAccountID converts the public key to an account value.
func PublicKeyToAccountID(pk ecdsa.PublicKey) AccountID {
	return AccountID(crypto.PubkeyToAddress(pk).String())
}

// IsAccountID verifies whether the underlying data represents a valid
// hex-encoded account.
func (a AccountID) IsAccountID() bool {
	const addressLength = 20

	if has0xPrefix(a) {
		a = a[2:]
	}

	return len(a) == 2*addressLength && isHex(a)
}

// =============================================================================

// has0xPrefix validates the account starts with a 0x.
func has0xPrefix(a AccountID) bool {
	return len(a) >= 2 && a[0] == '0' && (a[1] == 'x' || a[1] == 'X')
}

// isHex validates whether each byte is valid hexadecimal string.
func isHex(a AccountID) bool {
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

// =============================================================================

// byAccount provides sorting support by the account id value.
type byAccount []Account

// Len returns the number of transactions in the list.
func (ba byAccount) Len() int {
	return len(ba)
}

// Less helps to sort the list by account id in ascending order to keep the
// accounts in the right order of processing.
func (ba byAccount) Less(i, j int) bool {
	return ba[i].AccountID < ba[j].AccountID
}

// Swap moves accounts in the order of the account id value.
func (ba byAccount) Swap(i, j int) {
	ba[i], ba[j] = ba[j], ba[i]
}
