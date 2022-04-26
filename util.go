package main

import (
	"errors"
	"fmt"
)

// getDenom get the denom to be used in transaction fees
// This method first will try to retrieve the denom from the configuration '[[chains]] denom'
// if the deno	m is not available in the configuration then it will try to retrieve it from
// the chain-registry (https://github.com/cosmos/chain-registry). But in order for the chain
// registry retrieval to work, the chain name in the configuration file has to match the
// chain_name property in the chain.json. For example
// https://github.com/cosmos/chain-registry/blob/5ebdb2cf8bf0a6a14d602d4e63fd046f66895cbb/cosmoshub/chain.json#L3
func getDenom(conf *Config, chainName string) (string, error) {
	chain, found := conf.GetChain(chainName)
	if !found {
		return "", errors.New(fmt.Sprintf("chain %s not found in config", chainName))
	} else {
		if chain.Denom != "" {
			return chain.Denom, nil
		} else {
			// Try chain registry
			return "", errors.New(fmt.Sprintf("cannot find denom in the config or registry"))
		}
	}
}
