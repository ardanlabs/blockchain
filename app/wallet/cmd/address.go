package cmd

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

// addressCmd represents the address command
var addressCmd = &cobra.Command{
	Use:   "address",
	Short: "Print address for the specific wallet",
	Run: func(cmd *cobra.Command, args []string) {
		privateKey, err := crypto.LoadECDSA(getPrivateKeyPath())
		if err != nil {
			log.Fatal(err)
		}
		account := crypto.PubkeyToAddress(privateKey.PublicKey)
		fmt.Println(account)
	},
}

func init() {
	rootCmd.AddCommand(addressCmd)
}
