// Package cmd contains wallet app
package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	privateKeyName string
	walletPath     string
)

const (
	keyExtenstion = ".ecdsa"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "You simple wallet",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringVarP(&privateKeyName, "wallet", "w", "private.ecdsa", "Path to the private key.")
	rootCmd.PersistentFlags().StringVarP(&walletPath, "wallet-path", "p", "zblock/wallets/", "Path to the directory with private keys.")
}

func getPrivateKeyPath() string {
	if !strings.HasSuffix(privateKeyName, keyExtenstion) {
		privateKeyName += keyExtenstion
	}
	return filepath.Join(walletPath, privateKeyName)
}
