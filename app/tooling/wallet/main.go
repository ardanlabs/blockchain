package main

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
)

func main() {
	password := getPassPhrase("Please enter a password to decrypt the wallet:", false)

	ks := keystore.NewKeyStore("./", keystore.StandardScryptN, keystore.StandardScryptP)
	acc, err := ks.NewAccount(password)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("New account created: %s\n", acc.Address.Hex())
}

func getPassPhrase(prompt string, confirmation bool) string {
	return utils.GetPassPhrase(prompt, confirmation)
}
