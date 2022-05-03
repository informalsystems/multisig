package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

// TODO: something more intelligent
// Remember to change this every time ...
const VERSION = "0.2.0"

var rootCmd = &cobra.Command{
	Use:     "multisig",
	Short:   "manage multisig transactions",
	Version: VERSION,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var txCmd = &cobra.Command{
	Use:   "tx",
	Short: "generate a new unsigned tx",
	// Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var pushCmd = &cobra.Command{
	Use:   "push <unsigned tx file> <chain name> <key name>",
	Short: "push the given unsigned tx with associated signing metadata",
	Long:  "if a tx already exists for this chain and key, it will start using prefixes",
	Args:  cobra.ExactArgs(3),
	RunE:  cmdPush,
}

var authzCmd = &cobra.Command{
	Use:   "authz",
	Short: "generate an unsigned authz tx grant",
	Args:  cobra.NoArgs, // print long help from custom verification
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var authzGrantCmd = &cobra.Command{
	Use:   "grant <chain name> <key name> <grantee address> <withdraw|commission|delegate|vote> <expiration days>",
	Short: "generate an authz grant tx and push it",
	Long: "\nThis commands allows you to generate an unsigned tx to grant authorization " +
		"to a 'grantee' address that will be able to execute transactions as specified in " +
		"the '<message-type>' parameter. The grant authz is the first step in order to " +
		"authorize, after the grant tx is signed, then an 'authz exec' command will need to " +
		"be signed and executed in order to enable the authorization on chain.\n" +
		"Example: Grant withdraw authz permissions to a grantee (cosmos1add... address) for 30 days\n" +
		"multisig tx grant cosmoshub my-key cosmos1adggsadfsadfffredffdssdf withdraw 30",
	Args: func(cmd *cobra.Command, args []string) error {
		numArgs := 5 // Update the number of arguments if command use changes
		if len(args) != numArgs {
			cmd.Help()
			return fmt.Errorf("\n accepts %d arg(s), received %d", numArgs, len(args))
		}
		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          cmdGrantAuthz,
}

var voteCmd = &cobra.Command{
	Use:   "vote <chain name> <key name> <proposal number> <vote option (yes/no/veto/abstain)>",
	Short: "generate a vote tx and push it",
	Args:  cobra.ExactArgs(4),
	RunE:  cmdVote,
}

var withdrawCmd = &cobra.Command{
	Use:   "withdraw <chain name> <key name>",
	Short: "generate a withdraw-all-rewards tx and push it",
	Args:  cobra.ExactArgs(2),
	RunE:  cmdWithdraw,
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
	Args:  cobra.ExactArgs(2),
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
	flagDenom       string
	flagTxIndex     int
)

func init() {
	// Main commands
	rootCmd.AddCommand(txCmd)
	rootCmd.AddCommand(signCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(broadcastCmd)
	rootCmd.AddCommand(rawCmd)

	// Raw commands
	rawCmd.AddCommand(rawBech32Cmd)
	rawCmd.AddCommand(rawCatCmd)
	rawCmd.AddCommand(rawUpCmd)
	rawCmd.AddCommand(rawDownCmd)
	rawCmd.AddCommand(rawMkdirCmd)
	rawCmd.AddCommand(rawDeleteCmd)

	// Tx subcommands
	txCmd.AddCommand(pushCmd)
	txCmd.AddCommand(voteCmd)
	txCmd.AddCommand(withdrawCmd)
	txCmd.AddCommand(authzCmd)

	// Authz subcommands
	authzCmd.AddCommand(authzGrantCmd)

	// Add flags to commands
	addTxCmdCommonFlags(pushCmd)

	addTxCmdCommonFlags(voteCmd)
	addDenomFlags(voteCmd)

	addTxCmdCommonFlags(withdrawCmd)
	addDenomFlags(withdrawCmd)

	addTxCmdCommonFlags(authzGrantCmd)
	addDenomFlags(authzGrantCmd)

	addSignCmdFlags(signCmd)

	addListCmdFlags(listCmd)

	addBroadcastCmdFlags(broadcastCmd)
}
