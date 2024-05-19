package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func GetNodeInfo(chain Chain, client *http.Client) (response NodeInfo, err error) {
	var body []byte

	// TODO: replace these hacks with a more elegant handling of the node endpoints and protocol
	httpUrl := strings.ReplaceAll(chain.Node, "tcp://", "http://")

	url := strings.ReplaceAll(httpUrl, "rpc", "rest") + "/cosmos/base/tendermint/v1beta1/node_info"
	body, err = HttpGet(url, client)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return
	}

	return
}
