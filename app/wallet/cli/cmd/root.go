// Package cmd contains wallet app
package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	accountName string
	accountPath string
)

const (
	keyExtenstion = ".ecdsa"
)

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringVarP(&accountName, "account", "a", "private.ecdsa", "Path to the private key.")
	rootCmd.PersistentFlags().StringVarP(&accountPath, "account-path", "p", "zblock/accounts/", "Path to the directory with private keys.")
}

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "You simple wallet",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func getPrivateKeyPath() string {
	if !strings.HasSuffix(accountName, keyExtenstion) {
		accountName += keyExtenstion
	}

	return filepath.Join(accountPath, accountName)
}
