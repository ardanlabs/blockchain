package signature_test

import (
	"testing"

	"github.com/ardanlabs/blockchain/foundation/blockchain/signature"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
	"github.com/ethereum/go-ethereum/crypto"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

func TestSigning(t *testing.T) {
	value := struct {
		Name string
	}{
		Name: "Bill",
	}

	pkHexKey := "fae85851bdf5c9f49923722ce38f3c1defcfd3619ef5453230a58ad805499959"
	from := "0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4"
	sigStr := "0x3fc1a5adca72b01479c92856f2498296975448a208413c8f5a66a79ac75503d4434bac60b5fd40ac51ad61235b208a8d52c6a615c7f9ee92b2d8ce2fbb855a7c1e"

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
				t.Fatalf("\t%s\tTest %d:\tShould be able to generate a private key: %s", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to generate a private key.", success, testID)

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

func TestHash(t *testing.T) {
	value := struct {
		Name string
	}{
		Name: "Bill",
	}
	hash := "0f6887ac85101d6d6425a617edf35bd721b5f619fb92c36c3d2224e3bdb0ee5a"

	t.Log("Given the need to the hash function is working.")
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

// Found a bug with converting signatures from the slice of bytes to [R|S|V].
// The S value for this data produces a leading 0 in the slice of bytes. The
// big package truncates that 0, causing a signature issue.
func TestFromAddress(t *testing.T) {
	t.Log("Given the need to validate FromAddress validates signatures.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a specific value.", testID)
		{
			tx, err := sign(storage.UserTx{Nonce: 2, To: "space", Tip: 75}, 0)
			if err != nil {
				t.Fatalf("\t%s \tShould be able to sign transaction: %s", failed, err)
			}
			t.Logf("\t%s \tShould be able to sign transaction.", success)

			if _, err = tx.FromAddress(); err == nil {
				t.Fatalf("\t%s \tShould be able to catch bad signature.", failed)
			}
			t.Logf("\t%s \tShould be able to to catch bad signature: %s", success, err)
		}
	}
}

func sign(tx storage.UserTx, gas uint) (storage.BlockTx, error) {
	pk, err := crypto.HexToECDSA("9f332e3700d8fc2446eaf6d15034cf96e0c2745e40353deef032a5dbf1dfed93")
	if err != nil {
		return storage.BlockTx{}, err
	}

	signedTx, err := tx.Sign(pk)
	if err != nil {
		return storage.BlockTx{}, err
	}

	return storage.NewBlockTx(signedTx, gas), nil
}
