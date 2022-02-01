package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ardanlabs/blockchain/app/services/node/handlers/v1/public"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

// balanceCmd represents the balance command
var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Print your balance.",
	Run: func(cmd *cobra.Command, args []string) {
		privateKey, err := crypto.LoadECDSA("private.ecdsa")
		if err != nil {
			log.Fatal(err)
		}
		account := crypto.PubkeyToAddress(privateKey.PublicKey)
		fmt.Println("For Account:", account)
		resp, err := http.Get(fmt.Sprintf("%s/v1/balances/list/%s", url, account))
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		var balances public.Balances
		if err := decoder.Decode(&balances); err != nil {
			log.Fatal(err)
		}
		if len(balances.Balances) > 0 {
			fmt.Println(balances.Balances[0].Balance)
		}
	},
}

func init() {
	rootCmd.AddCommand(balanceCmd)
	balanceCmd.Flags().StringVarP(&url, "url", "u", "http://localhost:8080", "Url of the node.")
}
