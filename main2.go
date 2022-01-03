package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "multisig",
	Short: "manage multisig transactions",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate <chain name> <key name>",
	Short: "sign a tx",
	Args:  cobra.ExactArgs(2),
	RunE:  cmdGenerate,
}

var signCmd = &cobra.Command{
	Use:   "sign <chain name> <key name>",
	Short: "generate a new unsigned tx",
	Args:  cobra.ExactArgs(2),
	// RunE:  cmdSign,
}

var listCmd = &cobra.Command{
	Use:   "list <chain name> <key name>",
	Short: "list items in a directory",
	RunE:  cmdList,
}

var (
	flagTx          string
	flagSequence    int
	flagAccount     int
	flagNode        string
	flagFrom        string
	flagAll         bool
	flagDescription string
)

func init() {
	rootCmd.AddCommand(generateCmd)
	//	rootCmd.AddCommand(signCmd)
	rootCmd.AddCommand(listCmd)

	generateCmd.Flags().StringVarP(&flagTx, "tx", "t", "", "unsigned tx file")
	generateCmd.MarkFlagRequired("tx")

	generateCmd.Flags().IntVarP(&flagSequence, "sequence", "s", 0, "sequence number for the tx")
	generateCmd.Flags().IntVarP(&flagAccount, "account", "a", 0, "account number for the tx")
	generateCmd.Flags().StringVarP(&flagNode, "node", "n", "", "tendermint rpc node to get sequence and account number from")

	signCmd.Flags().StringVarP(&flagFrom, "from", "f", "", "name of your local key to sign with")
	signCmd.MarkFlagRequired("from")

	listCmd.Flags().BoolVarP(&flagAll, "all", "a", false, "list files for all chains and keys")
}
