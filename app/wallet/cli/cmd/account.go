package cmd

import (
	"fmt"
	"log"

	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Print account for the specific wallet",
	Run:   accountRun,
}

func init() {
	rootCmd.AddCommand(accountCmd)
}

func accountRun(cmd *cobra.Command, args []string) {
	privateKey, err := crypto.LoadECDSA(getPrivateKeyPath())
	if err != nil {
		log.Fatal(err)
	}

	account := storage.PublicKeyToAccount(privateKey.PublicKey)
	fmt.Println(account)
}
