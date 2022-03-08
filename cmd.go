package main

import (
	"github.com/spf13/cobra"
)

// TODO: something more intelligent
// Remember to change this every time ...
const VERSION = "0.1.0"

var rootCmd = &cobra.Command{
	Use:     "multisig",
	Short:   "manage multisig transactions",
	Version: VERSION,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate <chain name> <key name>",
	Short: "generate a new unsigned tx",
	Long:  "if a tx already exists for this chain and key, it will start using prefixes",
	Args:  cobra.ExactArgs(2),
	RunE:  cmdGenerate,
}

var signCmd = &cobra.Command{
	Use:   "sign <chain name> <key name>",
	Short: "sign a tx",
	Args:  cobra.ExactArgs(2),
	RunE:  cmdSign,
}

var listCmd = &cobra.Command{
	Use:   "list <chain name> <key name>",
	Short: "list items in a directory",
	RunE:  cmdList,
}

var broadcastCmd = &cobra.Command{
	Use:   "broadcast <chain name> <key name>",
	Short: "broadcast a tx",
	Args:  cobra.ExactArgs(2),
	RunE:  cmdBroadcast,
}

var rawCmd = &cobra.Command{
	Use:   "raw <cmd>",
	Short: "raw operations on the s3 bucket",
}

var rawBech32Cmd = &cobra.Command{
	Use:   "bech32 <bech32 string> <new prefix>",
	Short: "convert a bech32 string to a different prefix",
	Args:  cobra.ExactArgs(2),
	RunE:  cmdRawBech32,
}

var rawCatCmd = &cobra.Command{
	Use:   "cat <chain name> <key name>",
	Short: "dump the contents of all files in a directory",
	Args:  cobra.ExactArgs(2),
	RunE:  cmdRawCat,
}

var rawUpCmd = &cobra.Command{
	Use:   "up <source filepath> <destination filepath>",
	Short: "upload a local file to a path in the s3 bucket",
	Args:  cobra.ExactArgs(2),
	RunE:  cmdRawUp,
}

var rawDownCmd = &cobra.Command{
	Use:   "down <source filepath> <destination filepath>",
	Short: "download a file or directory from the s3 bucket",
	Long:  "if the path ends in a '/' it will attempt to download all files in that directory",
	Args:  cobra.ExactArgs(2),
	RunE:  cmdRawDown,
}

var rawMkdirCmd = &cobra.Command{
	Use:   "mkdir <directory path>",
	Short: "create a directory in the s3 bucket - must end with a '/'",
	Long:  "if the path ends in a '/' it will attempt to download all files in that directory",
	Args:  cobra.ExactArgs(1),
	RunE:  cmdRawMkdir,
}

var rawDeleteCmd = &cobra.Command{
	Use:   "delete <filepath>",
	Short: "delete a file from the s3 bucket",
	Args:  cobra.ExactArgs(1),
	RunE:  cmdRawDelete,
}

var (
	flagTx          string
	flagSequence    int
	flagAccount     int
	flagNode        string
	flagFrom        string
	flagAll         bool
	flagForce       bool
	flagAdditional  bool
	flagDescription string
)

func init() {
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(signCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(broadcastCmd)
	rootCmd.AddCommand(rawCmd)

	rawCmd.AddCommand(rawBech32Cmd)
	rawCmd.AddCommand(rawCatCmd)
	rawCmd.AddCommand(rawUpCmd)
	rawCmd.AddCommand(rawDownCmd)
	rawCmd.AddCommand(rawMkdirCmd)
	rawCmd.AddCommand(rawDeleteCmd)

	generateCmd.Flags().StringVarP(&flagTx, "tx", "t", "", "unsigned tx file")
	generateCmd.MarkFlagRequired("tx")
	generateCmd.Flags().IntVarP(&flagSequence, "sequence", "s", 0, "sequence number for the tx")
	generateCmd.Flags().IntVarP(&flagAccount, "account", "a", 0, "account number for the tx")
	generateCmd.Flags().StringVarP(&flagNode, "node", "n", "", "tendermint rpc node to get sequence and account number from")
	generateCmd.Flags().BoolVarP(&flagForce, "force", "f", false, "overwrite files already there")
	generateCmd.Flags().BoolVarP(&flagAdditional, "additional", "x", false, "add additional txs with higher sequence number")

	signCmd.Flags().StringVarP(&flagFrom, "from", "f", "", "name of your local key to sign with")
	signCmd.MarkFlagRequired("from")

	listCmd.Flags().BoolVarP(&flagAll, "all", "a", false, "list files for all chains and keys")

	broadcastCmd.Flags().StringVarP(&flagNode, "node", "n", "", "node address to broadcast too. flag overrides config")
	// broacastCmd.Flags().StringVarP(&flagDescription, "description", "d", "", "description of the tx to be logged")
}
