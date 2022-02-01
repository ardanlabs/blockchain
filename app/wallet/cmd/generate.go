package cmd

import (
	"log"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate new key pair",
	Run: func(cmd *cobra.Command, args []string) {
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			log.Fatal(err)
		}
		if err := crypto.SaveECDSA("private.ecdsa", privateKey); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
