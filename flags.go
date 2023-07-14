package main

import "github.com/spf13/cobra"

// addTxCmdCommonFlags defines common flags to be reused across tx commands
func addTxCmdCommonFlags(cmd *cobra.Command) {
	cmd.Flags().IntVarP(&flagSequence, "sequence", "s", 0, "sequence number for the tx")
	cmd.Flags().IntVarP(&flagAccount, "account", "a", 0, "account number for the tx")
	cmd.Flags().StringVarP(&flagNode, "node", "n", "", "tendermint rpc node to get sequence and account number from")
	cmd.Flags().BoolVarP(&flagForce, "force", "f", false, "overwrite files already there")
	cmd.Flags().BoolVarP(&flagAdditional, "additional", "x", false, "add additional txs with higher sequence number")
	cmd.Flags().StringVarP(&flagDescription, "description", "i", "", "information about the transaction")
}

// addTxCmdGasFeesFlags defines flags for gas and fees to be used in transactions
func addTxCmdGasFeesFlags(cmd *cobra.Command) {
	cmd.Flags().IntVarP(&flagGas, "gas", "g", 0, "gas limit for the transaction, e.g. 200000")
	cmd.Flags().StringVarP(&flagFees, "fees", "", "", "fees to pay for the transaction, e.g. 10uatom")
}

// addDenomFlags defines a denom flag to be reused across commands
func addDenomFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&flagDenom, "denom", "d", "", "fee denom, for offline creation")
}

// addSignCmdFlags defines common flags to be used in the sign command
func addSignCmdFlags(cmd *cobra.Command) {
	cmd.Flags().IntVarP(&flagTxIndex, "index", "i", 0, "index of the tx to sign")
	cmd.Flags().StringVarP(&flagFrom, "from", "f", "", "name of your local key to sign with")
	cmd.MarkFlagRequired("from")
}

// addListCmdFlags defines common flags to be used in the list command
func addListCmdFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&flagAll, "all", "a", false, "list files for all chains and keys")
}

// addBroadcastCmdFlags defines common flags to be used in the broadcast command
func addBroadcastCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&flagNode, "node", "n", "", "node address to broadcast too. flag overrides config")
	cmd.Flags().IntVarP(&flagTxIndex, "index", "i", 0, "index of the tx to broadcast")
	cmd.Flags().StringVarP(&flagMultisigKey, "key", "k", "", "name of the local multisig key name, flag overrides the config")
}

// addDeleteCmdFlags defines common flags to be used in the delete command
func addDeleteCmdFlags(cmd *cobra.Command) {
	cmd.Flags().IntVarP(&flagTxIndex, "index", "i", 0, "index of the tx to delete")
}

// addGlobalFlags defines flags to be used regardless of the command used
func addGlobalFlags(cmd *cobra.Command) {
	rootCmd.PersistentFlags().StringVarP(&flagConfigPath, "config", "c", "", "custom config path")
	rootCmd.PersistentFlags().StringVarP(&flagHomePath, "home", "", "", "custom home path for keystore for sign and broadcast cmds")
}
