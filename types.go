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
