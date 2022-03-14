package cmd

import (
	"log"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate new key pair",
	Run:   generateRun,
}

func init() {
	rootCmd.AddCommand(generateCmd)
}

func generateRun(cmd *cobra.Command, args []string) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	if err := crypto.SaveECDSA(getPrivateKeyPath(), privateKey); err != nil {
		log.Fatal(err)
	}
}
