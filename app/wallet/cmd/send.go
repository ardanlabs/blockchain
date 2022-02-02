package cmd

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ardanlabs/blockchain/app/services/node/handlers/v1/public"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

var (
	url   string
	to    string
	value uint
	tip   uint
	data  string
	file  string
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

		if file == "" {
			sendWithDetails(privateKey)
			return
		}

		sendWithFile(privateKey)
	},
}

func sendWithFile(privateKey *ecdsa.PrivateKey) {
}

func sendWithDetails(privateKey *ecdsa.PrivateKey) {
	tx := public.Tx{
		To:    to,
		Value: value,
		Tip:   tip,
		Data:  data,
	}
	data, err := json.Marshal(tx)
	if err != nil {
		log.Fatal(err)
	}
	hash := crypto.Keccak256Hash(data)
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	payload := []public.SignedTx{
		{
			Transaction: tx,
			Signature:   signature,
		},
	}
	data, err = json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.Post(fmt.Sprintf("%s/v1/tx/send", url), "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
}

func init() {
	rootCmd.AddCommand(sendCmd)
	sendCmd.Flags().StringVarP(&url, "url", "u", "http://localhost:8080", "Url of the node.")
	sendCmd.Flags().StringVarP(&to, "to", "t", "", "Url of the node.")
	sendCmd.MarkFlagRequired("to")
	sendCmd.Flags().UintVarP(&value, "value", "v", 0, "Value to send.")
	sendCmd.Flags().UintVarP(&tip, "tip", "c", 0, "Tip to send.")
	sendCmd.Flags().StringVarP(&data, "data", "d", "", "Data to send.")
	sendCmd.Flags().StringVarP(&data, "file", "f", "", "File to read for transactions.")
}
