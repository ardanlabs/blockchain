package signature_test

import (
	"testing"

	"github.com/ardanlabs/blockchain/foundation/blockchain/signature"
	"github.com/ethereum/go-ethereum/crypto"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
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

	t.Log("Given the need to the validate signatures.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a specific private/public key.", testID)
		{
			pk, err := crypto.HexToECDSA(pkHexKey)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to generate a private key: %s", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to generate a private key.", success, testID)

			v, r, s, err := signature.Sign(value, pk)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to sign data: %s", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to sign data.", success, testID)

			if err := signature.VerifySignature(value, v, r, s); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to verify the signature: %s", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to verify the signature.", success, testID)

			addr, err := signature.FromAddress(value, v, r, s)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to generate from address: %s", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to generate from address.", success, testID)

			if from != addr {
				t.Logf("\t%s\tTest %d:\tgot: %s", failed, testID, addr)
				t.Logf("\t%s\tTest %d:\texp: %s", failed, testID, from)
				t.Fatalf("\t%s\tTest %d:\tShould get back the right address.", failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the right address: %s", success, testID, addr[:10])

			str := signature.SignatureString(v, r, s)
			if from != addr {
				t.Logf("\t%s\tTest %d:\tgot: %s", failed, testID, str[:10])
				t.Logf("\t%s\tTest %d:\texp: %s", failed, testID, sigStr[:10])
				t.Fatalf("\t%s\tTest %d:\tShould get back the right signature string.", failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the right signature string: %s", success, testID, str[:6])
		}
	}
}

func Test_Hash(t *testing.T) {
	value := struct {
		Name string
	}{
		Name: "Bill",
	}
	hash := "0x0f6887ac85101d6d6425a617edf35bd721b5f619fb92c36c3d2224e3bdb0ee5a"

	t.Log("Given the need to verify the hash function is working.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a specific value.", testID)
		{
			h := signature.Hash(value)
			if h != hash {
				t.Logf("\t%s\tTest %d:\tgot: %s", failed, testID, h)
				t.Logf("\t%s\tTest %d:\texp: %s", failed, testID, hash)
				t.Fatalf("\t%s\tTest %d:\tShould get back the right hash: %s", failed, testID, h[:6])
			}
			t.Logf("\t%s\tTest %d:\tShould get back the right hash: %s", success, testID, h[:6])

			h = signature.Hash(value)
			if h != hash {
				t.Logf("\t%s\tTest %d:\tgot: %s", failed, testID, h)
				t.Logf("\t%s\tTest %d:\texp: %s", failed, testID, hash)
				t.Fatalf("\t%s\tTest %d:\tShould get back the same hash twice.", failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same hash twice: %s", success, testID, h[:6])
		}
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

	t.Log("Given the need to verify signatures are working.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a set of transaction.", testID)
		{
			pk, err := crypto.HexToECDSA(pkHexKey)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to generate a private key: %s", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to generate a private key.", success, testID)

			v1, r1, s1, err := signature.Sign(value1, pk)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to sign data: %s", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to sign data.", success, testID)

			addr1, err := signature.FromAddress(value1, v1, r1, s1)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to generate an address: %s", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to generate an address.", success, testID)

			v2, r2, s2, err := signature.Sign(value2, pk)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to sign data: %s", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to sign data.", success, testID)

			addr2, err := signature.FromAddress(value2, v2, r2, s2)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to generate an address: %s", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to generate an address.", success, testID)

			if addr1 != addr2 {
				t.Errorf("\t%s\tTest %d:\tGot: %s", failed, testID, addr1)
				t.Errorf("\t%s\tTest %d:\tGot: %s", failed, testID, addr2)
				t.Fatalf("\t%s\tTest %d:\tShould have the same address.", failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould have the same address.", success, testID)
		}
	}
}
