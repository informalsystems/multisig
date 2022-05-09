package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// ChainInfo simplified version of the chain information from the registry
type ChainInfo struct {
	Schema       string   `json:"$schema"`
	ChainName    string   `json:"chain_name"`
	ChainID      string   `json:"chain_id"`
	PrettyName   string   `json:"pretty_name"`
	Status       string   `json:"status"`
	NetworkType  string   `json:"network_type"`
	Bech32Prefix string   `json:"bech32_prefix"`
	DaemonName   string   `json:"daemon_name"`
	NodeHome     string   `json:"node_home"`
	KeyAlgos     []string `json:"key_algos"`
	Slip44       int      `json:"slip44"`
	Fees         Fees     `json:"fees"`
}

type FeeTokens struct {
	Denom            string `json:"denom"`
	FixedMinGasPrice int    `json:"fixed_min_gas_price"`
}
type Fees struct {
	FeeTokens []FeeTokens `json:"fee_tokens"`
}

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
			denom, err := getDenomFromRegistry(chainName)
			if err != nil {
				return "", errors.New(fmt.Sprintf("cannot find denom in the config or registry: %s", err))
			} else {
				return denom, nil
			}
		}
	}
}

func getDenomFromRegistry(chainName string) (string, error) {

	chain := ChainInfo{}
	denom := ""
	url := fmt.Sprintf("https://raw.githubusercontent.com/cosmos/chain-registry/master/%s/chain.json", chainName)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
		return denom, err
	}
	res, err := client.Do(req)
	if (err != nil) || (res.StatusCode == 404) {
		fmt.Println(err)
		return denom, errors.New(fmt.Sprintf("cannot find denom in the chain registry, please ensure the chain name in the configuration file matches the folder name in the registry (https://github.com/cosmos/chain-registry)"))
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return denom, err
	}
	err = json.Unmarshal(body, &chain)
	if err != nil {
		fmt.Println(err)
		return denom, err
	}

	// Assumption that the first fee token is the one used for paying
	// the fees and the fee information is in the registry
	if chain.Fees.FeeTokens != nil {
		if chain.Fees.FeeTokens[0].Denom != "" {
			return chain.Fees.FeeTokens[0].Denom, nil
		}
	}

	return "", errors.New(fmt.Sprintf("cannot find denom fee information for the %s chain in the registry", chainName))
}

// Quick way to parse the denom from the json
// without worrying on the different message types
// that the unsigned may have. In the future this
// might be improved parsing the right msg type
func parseDenomFromJson(tx []byte) (string, error) {
	var anyJson map[string]interface{}
	err := json.Unmarshal(tx, &anyJson)
	if err != nil {
		return "", err
	}
	authJson := anyJson["auth_info"]
	if authJson != nil {
		auth := authJson.(map[string]interface{})
		feeJson := auth["fee"]
		if feeJson != nil {
			fee := feeJson.(map[string]interface{})
			amountJson := fee["amount"]
			if amountJson != nil {
				amount := amountJson.([]interface{})
				if len(amount) >= 1 {
					first := amount[0].(map[string]interface{})
					firstJson := first["denom"]
					if firstJson != nil {
						return firstJson.(string), nil
					} else {
						return "", fmt.Errorf("cannot parse tx json")
					}
				} else {
					return "", fmt.Errorf("cannot parse tx json")
				}
			} else {
				return "", fmt.Errorf("cannot parse tx json")
			}
		} else {
			return "", fmt.Errorf("cannot parse tx json")
		}

	} else {
		return "", fmt.Errorf("cannot parse tx json")
	}
}
