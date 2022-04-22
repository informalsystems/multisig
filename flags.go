package main

import "github.com/spf13/cobra"

// addTxCmdCommonFlags defines common flags to be reused across tx commands
func addTxCmdCommonFlags(cmd *cobra.Command) {
	cmd.Flags().IntVarP(&flagSequence, "sequence", "s", 0, "sequence number for the tx")
	cmd.Flags().IntVarP(&flagAccount, "account", "a", 0, "account number for the tx")
	cmd.Flags().StringVarP(&flagNode, "node", "n", "", "tendermint rpc node to get sequence and account number from")
	cmd.Flags().BoolVarP(&flagForce, "force", "f", false, "overwrite files already there")
	cmd.Flags().BoolVarP(&flagAdditional, "additional", "x", false, "add additional txs with higher sequence number")
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
	// broacastCmd.Flags().StringVarP(&flagDescription, "description", "d", "", "description of the tx to be logged")
}
