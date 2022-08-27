package signature_test

import (
	"testing"

	"github.com/ardanlabs/blockchain/foundation/blockchain/signature"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	pkHexKey = "fae85851bdf5c9f49923722ce38f3c1defcfd3619ef5453230a58ad805499959"
	from     = "0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4"
	sigStr   = "0x3fc1a5adca72b01479c92856f2498296975448a208413c8f5a66a79ac75503d4434bac60b5fd40ac51ad61235b208a8d52c6a615c7f9ee92b2d8ce2fbb855a7c1e"
)

// =============================================================================

func Test_Signing(t *testing.T) {
	value := struct {
		Name string
	}{
		Name: "Bill",
	}

	pk, err := crypto.HexToECDSA(pkHexKey)
	if err != nil {
		t.Fatalf("Should be able to generate a private key: %s", err)
	}

	v, r, s, err := signature.Sign(value, pk)
	if err != nil {
		t.Fatalf("Should be able to sign data: %s", err)
	}

	if err := signature.VerifySignature(v, r, s); err != nil {
		t.Fatalf("Should be able to verify the signature: %s", err)
	}

	addr, err := signature.FromAddress(value, v, r, s)
	if err != nil {
		t.Fatalf("Should be able to generate from address: %s", err)
	}

	if from != addr {
		t.Logf("got: %s", addr)
		t.Logf("exp: %s", from)
		t.Fatalf("Should get back the right address.")
	}

	str := signature.SignatureString(v, r, s)
	if from != addr {
		t.Logf("got: %s", str[:10])
		t.Logf("exp: %s", sigStr[:10])
		t.Fatalf("Should get back the right signature string.")
	}
}

func Test_Hash(t *testing.T) {
	value := struct {
		Name string
	}{
		Name: "Bill",
	}
	hash := "0x0f6887ac85101d6d6425a617edf35bd721b5f619fb92c36c3d2224e3bdb0ee5a"

	h := signature.Hash(value)
	if h != hash {
		t.Logf("got: %s", h)
		t.Logf("exp: %s", hash)
		t.Fatalf("Should get back the right hash: %s", h[:6])
	}

	h = signature.Hash(value)
	if h != hash {
		t.Logf("got: %s", h)
		t.Logf("exp: %s", hash)
		t.Fatalf("Should get back the same hash twice.")
	}
}

func Test_SignConsistency(t *testing.T) {
	value1 := struct {
		Name string
	}{
		Name: "Bill",
	}
	value2 := struct {
		Name string
	}{
		Name: "Jill",
	}

	pk, err := crypto.HexToECDSA(pkHexKey)
	if err != nil {
		t.Fatalf("Should be able to generate a private key: %s", err)
	}

	v1, r1, s1, err := signature.Sign(value1, pk)
	if err != nil {
		t.Fatalf("Should be able to sign data: %s", err)
	}

	addr1, err := signature.FromAddress(value1, v1, r1, s1)
	if err != nil {
		t.Fatalf("Should be able to generate an address: %s", err)
	}

	v2, r2, s2, err := signature.Sign(value2, pk)
	if err != nil {
		t.Fatalf("Should be able to sign data: %s", err)
	}

	addr2, err := signature.FromAddress(value2, v2, r2, s2)
	if err != nil {
		t.Fatalf("Should be able to generate an address: %s", err)
	}

	if addr1 != addr2 {
		t.Errorf("Got: %s", addr1)
		t.Errorf("Got: %s", addr2)
		t.Fatalf("Should have the same address.")
	}
}
