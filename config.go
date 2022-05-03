package main

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"

	"github.com/cosmos/cosmos-sdk/types/bech32"
)

// A chain we sign txs on
type Chain struct {
	Name   string // chain name
	Binary string // binary to use for signing
	Prefix string // bech32 address prefix
	ID     string // chain id for signing
	Node   string // node to broadcast signed txs to
	Denom  string // denom used for fees
}

// A key we sign txs with
type Key struct {
	Name      string
	Address   string
	LocalName string
}

// Credentials for AWS
type AWS struct {
	Bucket       string
	BucketRegion string
	Pub          string
	Priv         string
}

// Config file
type Config struct {
	User           string
	KeyringBackend string // TODO: how to support snake_case?

	AWS    AWS
	Keys   []Key
	Chains []Chain
}

func (c *Config) GetChain(name string) (Chain, bool) {
	for _, chain := range c.Chains {
		if chain.Name == name {
			return chain, true
		}
	}
	return Chain{}, false
}

func (c *Config) GetKey(name string) (Key, bool) {
	for _, key := range c.Keys {
		if key.Name == name {
			return key, true
		}
	}
	return Key{}, false
}

// load toml config
func loadConfig(filename string) (*Config, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	err = toml.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}

	if c.AWS.BucketRegion == "" {
		c.AWS.BucketRegion = defaultBucketRegion
	}

	return c, nil
}

// convert the prefix on a bech32 address
func bech32ify(addrBech, prefix string) (string, error) {
	hrp, addrBytes, err := bech32.DecodeAndConvert(addrBech)
	if err != nil {
		return "", err
	}
	_ = hrp

	newAddrBech, err := bech32.ConvertAndEncode(prefix, addrBytes)
	if err != nil {
		return "", err
	}
	return newAddrBech, nil
}
