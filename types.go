package main

type ContinuousVestingAccount struct {
	Type               string `json:"@type"`
	BaseVestingAccount struct {
		BaseAccount struct {
			Address string `json:"address"`
			PubKey  struct {
				Type       string `json:"@type"`
				Threshold  int    `json:"threshold"`
				PublicKeys []struct {
					Type string `json:"@type"`
					Key  string `json:"key"`
				} `json:"public_keys"`
			} `json:"pub_key"`
			AccountNumber string `json:"account_number"`
			Sequence      string `json:"sequence"`
		} `json:"base_account"`
		OriginalVesting []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"original_vesting"`
		DelegatedFree []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"delegated_free"`
		DelegatedVesting []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"delegated_vesting"`
		EndTime string `json:"end_time"`
	} `json:"base_vesting_account"`
	StartTime string `json:"start_time"`
}

type DelayedVestingAccount struct {
	Type               string `json:"@type"`
	BaseVestingAccount struct {
		BaseAccount struct {
			Address string `json:"address"`
			PubKey  struct {
				Type       string `json:"@type"`
				Threshold  int    `json:"threshold"`
				PublicKeys []struct {
					Type string `json:"@type"`
					Key  string `json:"key"`
				} `json:"public_keys"`
			} `json:"pub_key"`
			AccountNumber string `json:"account_number"`
			Sequence      string `json:"sequence"`
		} `json:"base_account"`
		OriginalVesting []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"original_vesting"`
		DelegatedFree []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"delegated_free"`
		DelegatedVesting []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"delegated_vesting"`
		EndTime string `json:"end_time"`
	} `json:"base_vesting_account"`
}

type PeriodicVestingAccount struct {
	Type               string `json:"@type"`
	BaseVestingAccount struct {
		BaseAccount struct {
			Address string `json:"address"`
			PubKey  struct {
				Type       string `json:"@type"`
				Threshold  int    `json:"threshold"`
				PublicKeys []struct {
					Type string `json:"@type"`
					Key  string `json:"key"`
				} `json:"public_keys"`
			} `json:"pub_key"`
			AccountNumber string `json:"account_number"`
			Sequence      string `json:"sequence"`
		} `json:"base_account"`
		OriginalVesting []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"original_vesting"`
		DelegatedFree []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"delegated_free"`
		DelegatedVesting []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"delegated_vesting"`
		EndTime string `json:"end_time"`
	} `json:"base_vesting_account"`
	StartTime      string `json:"start_time"`
	VestingPeriods []struct {
		Length string `json:"length"`
		Amount []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"amount"`
	} `json:"vesting_periods"`
}

type EthAccount struct {
	Type        string `json:"@type"`
	BaseAccount struct {
		Address string `json:"address"`
		PubKey  struct {
			Type       string `json:"@type"`
			Threshold  int    `json:"threshold"`
			PublicKeys []struct {
				Type string `json:"@type"`
				Key  string `json:"key"`
			} `json:"public_keys"`
		} `json:"pub_key"`
		AccountNumber string `json:"account_number"`
		Sequence      string `json:"sequence"`
	} `json:"base_account"`
	CodeHash string `json:"code_hash"`
}

type StridePeriodicVestingAccount struct {
	Type               string `json:"@type"`
	BaseVestingAccount struct {
		BaseAccount struct {
			Address       string      `json:"address"`
			PubKey        interface{} `json:"pub_key"`
			AccountNumber string      `json:"account_number"`
			Sequence      string      `json:"sequence"`
		} `json:"base_account"`
		OriginalVesting []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"original_vesting"`
		DelegatedFree    []interface{} `json:"delegated_free"`
		DelegatedVesting []interface{} `json:"delegated_vesting"`
		EndTime          string        `json:"end_time"`
	} `json:"base_vesting_account"`
	VestingPeriods []struct {
		StartTime string `json:"start_time"`
		Length    string `json:"length"`
		Amount    []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"amount"`
		ActionType int `json:"action_type"`
	} `json:"vesting_periods"`
}

type BaseAccount struct {
	Type    string `json:"@type"`
	Address string `json:"address"`
	PubKey  struct {
		Type       string `json:"@type"`
		Threshold  int    `json:"threshold"`
		PublicKeys []struct {
			Type string `json:"@type"`
			Key  string `json:"key"`
		} `json:"public_keys"`
	} `json:"pub_key"`
	AccountNumber string `json:"account_number"`
	Sequence      string `json:"sequence"`
}

type AccountType struct {
	Type string `json:"@type"`
}

type AccountBalance struct {
	Balances []struct {
		Denom  string `json:"denom"`
		Amount string `json:"amount"`
	} `json:"balances"`
	Pagination struct {
		NextKey interface{} `json:"next_key"`
		Total   string      `json:"total"`
	} `json:"pagination"`
}

type ApplicationVersion struct {
	Name      string `json:"name"`
	AppName   string `json:"app_name"`
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildTags string `json:"build_tags"`
	GoVersion string `json:"go_version"`
	BuildDeps []struct {
		Path    string `json:"path"`
		Version string `json:"version"`
		Sum     string `json:"sum"`
	} `json:"build_deps"`
	CosmosSdkVersion string `json:"cosmos_sdk_version"`
}

type NodeInfo struct {
	DefaultNodeInfo struct {
		ProtocolVersion struct {
			P2P   string `json:"p2p"`
			Block string `json:"block"`
			App   string `json:"app"`
		} `json:"protocol_version"`
		DefaultNodeId string `json:"default_node_id"`
		ListenAddress string `json:"listen_address"`
		Network       string `json:"network"`
		Version       string `json:"version"`
		Channels      string `json:"channels"`
		Moniker       string `json:"moniker"`
		Other         struct {
			TxIndex    string `json:"tx_index"`
			RpcAddress string `json:"rpc_address"`
		} `json:"other"`
	} `json:"default_node_info"`
	ApplicationVersion ApplicationVersion `json:"application_version"`
}
