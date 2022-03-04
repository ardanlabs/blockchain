package cmd

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

var (
	url   string
	nonce uint
	to    string
	value uint
	tip   uint
	data  []byte
)

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send transaction",
	Run: func(cmd *cobra.Command, args []string) {
		privateKey, err := crypto.LoadECDSA(getPrivateKeyPath())
		if err != nil {
			log.Fatal(err)
		}

		sendWithDetails(privateKey)
	},
}

func sendWithDetails(privateKey *ecdsa.PrivateKey) {
	toAddr, err := storage.ToAddress(to)
	if err != nil {
		log.Fatal(err)
	}

	userTx, err := storage.NewUserTx(nonce, toAddr, value, tip, data)
	if err != nil {
		log.Fatal(err)
	}

	signedTx, err := userTx.Sign(privateKey)
	if err != nil {
		log.Fatal(err)
	}

	data, err := json.Marshal(signedTx)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.Post(fmt.Sprintf("%s/v1/tx/submit", url), "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
}

func init() {
	rootCmd.AddCommand(sendCmd)
	sendCmd.Flags().StringVarP(&url, "url", "u", "http://localhost:8080", "Url of the node.")
	sendCmd.Flags().UintVarP(&nonce, "nonce", "n", 0, "id for the transaction.")
	sendCmd.Flags().StringVarP(&to, "to", "t", "", "Url of the node.")
	sendCmd.Flags().UintVarP(&value, "value", "v", 0, "Value to send.")
	sendCmd.Flags().UintVarP(&tip, "tip", "c", 0, "Tip to send.")
	sendCmd.Flags().BytesHexVarP(&data, "data", "d", nil, "Data to send.")
}
