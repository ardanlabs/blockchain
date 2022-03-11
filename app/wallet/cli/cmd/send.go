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
	toAccount, err := storage.ToAccount(to)
	if err != nil {
		log.Fatal(err)
	}

	userTx, err := storage.NewUserTx(nonce, toAccount, value, tip, data)
	if err != nil {
		log.Fatal(err)
	}

	signedTx, err := userTx.Sign(privateKey)
	if err != nil {
		log.Fatal(err)
	}

	type walletTx struct {
		Nonce uint   `json:"nonce"` // Unique id for the transaction supplied by the user.
		To    string `json:"to"`    // Account receiving the benefit of the transaction.
		Value uint   `json:"value"` // Monetary value received from this transaction.
		Tip   uint   `json:"tip"`   // Tip offered by the sender as an incentive to mine this transaction.
		Data  []byte `json:"data"`  // Extra data related to the transaction.
		Sig   string `json:"sig"`   // Raw signature of the account who signed the transaction.
	}

	w := walletTx{
		Nonce: signedTx.Nonce,
		To:    string(signedTx.To),
		Value: signedTx.Value,
		Tip:   signedTx.Tip,
		Data:  signedTx.Data,
		Sig:   signedTx.SignatureString(),
	}

	fmt.Println(w.Sig)

	data, err := json.Marshal(w)
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
