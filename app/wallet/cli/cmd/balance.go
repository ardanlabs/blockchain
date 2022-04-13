package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

type balance struct {
	Account string `json:"account"`
	Balance uint   `json:"balance"`
}

type balances struct {
	LastestBlock string    `json:"lastest_block"`
	Uncommitted  int       `json:"uncommitted"`
	Balances     []balance `json:"balances"`
}

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Print your balance.",
	Run:   balanceRun,
}

func init() {
	rootCmd.AddCommand(balanceCmd)
	balanceCmd.Flags().StringVarP(&url, "url", "u", "http://localhost:8080", "Url of the node.")
}

func balanceRun(cmd *cobra.Command, args []string) {
	privateKey, err := crypto.LoadECDSA(getPrivateKeyPath())
	if err != nil {
		log.Fatal(err)
	}

	accountID := database.PublicKeyToAccountID(privateKey.PublicKey)
	fmt.Println("For Account:", accountID)

	resp, err := http.Get(fmt.Sprintf("%s/v1/balances/list/%s", url, accountID))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var balances balances
	if err := decoder.Decode(&balances); err != nil {
		log.Fatal(err)
	}

	if len(balances.Balances) > 0 {
		fmt.Println(balances.Balances[0].Balance)
	}
}
