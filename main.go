package main

import (
	"cosmossdk.io/math"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// global vars
var (

	// config file - expected to be in the present working directory
	defaultLocalConfigFile  = "config.toml"
	defaultGlobalConfigFile = ".multisig/config.toml"

	// files for signing - we use these filenames in the local working directory and in the remote bucket
	unsignedJSON = "unsigned.json"
	signedJSON   = "signed.json"
	signDataJSON = "signdata.json"
)

// Data we need for signers to sign a tx (eg. without access to a node)
type SignData struct {
	Account     int    `json:"account"`
	Sequence    int    `json:"sequence"`
	ChainID     string `json:"chain-id"`
	Description string `json:"description"`
}

func main() {
	// cmds defined in cmd.go
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func cmdDelete(cobraCmd *cobra.Command, args []string) error {
	chainName := args[0]
	keyName := args[1]

	conf, err := loadConfig(flagConfigPath)
	if err != nil {
		return err
	}

	txIndex := flagTxIndex
	txDir := filepath.Join(chainName, keyName, fmt.Sprintf("%d", txIndex))

	err = deleteAllFilesInPath(txDir, conf)
	if err != nil {
		return err
	}

	return nil
}

// Generates a [binary] `tx distribution withdraw-all-rewards` transaction
func cmdWithdraw(cmd *cobra.Command, args []string) error {
	chainName := args[0]
	keyName := args[1]

	conf, err := loadConfig(flagConfigPath)
	if err != nil {
		return err
	}

	chain, found := conf.GetChain(chainName)
	if !found {
		return fmt.Errorf("chain %s not found in config", chainName)
	}
	key, found := conf.GetKey(keyName)
	if !found {
		return fmt.Errorf("key %s not found in config", keyName)
	}

	nodeAddress := chain.Node
	if flagNode != "" {
		nodeAddress = flagNode
	}

	binary := chain.Binary
	address, err := bech32ify(key.Address, chain.Prefix)
	if err != nil {
		return err
	}

	// Get fees
	fees, err := getFeesParameter(cmd)
	if err != nil {
		return err
	}

	// [binary] tx distribution withdraw-all-rewards
	cmdArgs := []string{"tx", "distribution", "withdraw-all-rewards",
		"--from", address,
		"--fees", fmt.Sprintf("%s%s", fees.Amount.String(), fees.Denom),
		"--gas", fmt.Sprintf("%d", getGas(conf)),
		"--generate-only",
		"--chain-id", fmt.Sprintf("%s", chain.ID),
	}

	if nodeAddress != "" {
		cmdArgs = append(cmdArgs, "--node", nodeAddress)
	}

	execCmd := exec.Command(binary, cmdArgs...)
	fmt.Println(execCmd)
	unsignedBytes, err := execCmd.CombinedOutput()
	if err != nil {
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println("call failed")
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println(execCmd)
		fmt.Println(string(unsignedBytes))
		return err
	}
	fmt.Println(string(unsignedBytes))

	return pushTx(chainName, keyName, unsignedBytes, cmd)
}

// Generates a [binary] `tx staking delegate` transaction
func cmdDelegate(cmd *cobra.Command, args []string) error {
	chainName := args[0]
	keyName := args[1]
	validator := args[2]
	amount := args[3]

	conf, err := loadConfig(flagConfigPath)
	if err != nil {
		return err
	}

	chain, found := conf.GetChain(chainName)
	if !found {
		return fmt.Errorf("chain %s not found in config", chainName)
	}
	key, found := conf.GetKey(keyName)
	if !found {
		return fmt.Errorf("key %s not found in config", keyName)
	}

	// Use denom from flag if specified, if not, then try
	// to retrieve it from the config, if not in the config
	// try to retrieve from the chain registry.
	var denom string
	isDenomSet := cmd.Flags().Changed("denom")
	if isDenomSet {
		denom = flagDenom
	} else {
		denom, err = getDenom(conf, chainName)
		if err != nil {
			return fmt.Errorf("denom not found in config or chain registry: %s", err)
		}
	}

	nodeAddress := chain.Node
	if flagNode != "" {
		nodeAddress = flagNode
	}

	binary := chain.Binary
	address, err := bech32ify(key.Address, chain.Prefix)
	if err != nil {
		return err
	}

	// Safe check for amount
	amountDecCoin, err := sdk.ParseDecCoin(amount)
	if err != nil {
		return fmt.Errorf("error parsing the amount to delegate, plesae specify amount and denom, e.g. 100uatom")
	}

	// Get fees
	fees, err := getFeesParameter(cmd)
	if err != nil {
		return err
	}

	// Check if amount + fee < available balance
	balance, err := getAccountBalance(address, denom, chain)
	if err != nil {
		fmt.Println("error getting account balance, skipping check to validate enough balance to delegate")
	} else {
		amountFee := amountDecCoin.Add(fees)
		if sdk.NewDecFromBigInt(balance.BigInt()).RoundInt().LT(amountFee.Amount.RoundInt()) {
			return fmt.Errorf("the balance available (%s) is less than the amount (%s) plus the fee (%s), transaction will fail", balance, balance.String(), amountFee.Amount.RoundInt().String())
		}
	}

	cmdArgs := []string{"tx", "staking", "delegate",
		validator,
		amount,
		"--from", address,
		"--fees", fmt.Sprintf("%s%s", fees.Amount.String(), fees.Denom),
		"--gas", fmt.Sprintf("%d", getGas(conf)),
		"--generate-only",
		"--chain-id", fmt.Sprintf("%s", chain.ID),
	}

	if nodeAddress != "" {
		cmdArgs = append(cmdArgs, "--node", nodeAddress)
	}

	execCmd := exec.Command(binary, cmdArgs...)
	fmt.Println(execCmd)
	unsignedBytes, err := execCmd.CombinedOutput()
	if err != nil {
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println("call failed")
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println(execCmd)
		fmt.Println(string(unsignedBytes))
		return err
	}
	fmt.Println(string(unsignedBytes))

	return pushTx(chainName, keyName, unsignedBytes, cmd)
}

func cmdClaimValidator(cmd *cobra.Command, args []string) error {
	chainName := args[0]
	keyName := args[1]
	valAddress := args[2]

	conf, err := loadConfig(flagConfigPath)
	if err != nil {
		return err
	}

	chain, found := conf.GetChain(chainName)
	if !found {
		return fmt.Errorf("chain %s not found in config", chainName)
	}
	key, found := conf.GetKey(keyName)
	if !found {
		return fmt.Errorf("key %s not found in config", keyName)
	}

	nodeAddress := chain.Node
	if flagNode != "" {
		nodeAddress = flagNode
	}

	binary := chain.Binary
	address, err := bech32ify(key.Address, chain.Prefix)
	if err != nil {
		return err
	}

	// Get fees
	fees, err := getFeesParameter(cmd)
	if err != nil {
		return err
	}

	// [binary] tx distribution withdraw-rewards [validator-addr]
	cmdArgs := []string{"tx", "distribution", "withdraw-rewards",
		valAddress,
		"--commission",
		"--from", address,
		"--fees", fmt.Sprintf("%s%s", fees.Amount.String(), fees.Denom),
		"--gas", fmt.Sprintf("%d", getGas(conf)),
		"--generate-only",
		"--chain-id", fmt.Sprintf("%s", chain.ID),
	}

	if nodeAddress != "" {
		cmdArgs = append(cmdArgs, "--node", nodeAddress)
	}

	// Append keyring if specified in the config
	if conf.KeyringBackend != "" {
		cmdArgs = append(cmdArgs, "--keyring-backend", conf.KeyringBackend)
	}

	execCmd := exec.Command(binary, cmdArgs...)
	fmt.Println(execCmd)
	unsignedBytes, err := execCmd.CombinedOutput()
	if err != nil {
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println("call failed")
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println(execCmd)
		fmt.Println(string(unsignedBytes))
		return err
	}
	fmt.Println(string(unsignedBytes))

	return pushTx(chainName, keyName, unsignedBytes, cmd)
}

func cmdGrantAuthz(cmd *cobra.Command, args []string) error {
	chainName := args[0]
	keyName := args[1]
	grantee := args[2]

	msgType := args[3]
	// Parse message-type parameter and generate proper tx msg-type
	// Only support the messages we need for now (withdraw, delegate, commission, vote)
	var cosmosMsg string
	switch msgType {
	case "withdraw":
		cosmosMsg = "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward"
	case "delegate":
		cosmosMsg = "/cosmos.staking.v1beta1.MsgDelegate"
	case "commission":
		cosmosMsg = "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission"
	case "vote":
		cosmosMsg = "/cosmos.gov.v1beta1.MsgVote"
	case "unbond":
		cosmosMsg = "/cosmos.staking.v1beta1.MsgUndelegate"
	case "redelegate":
		cosmosMsg = "/cosmos.staking.v1beta1.MsgBeginRedelegate"
	default:
		return fmt.Errorf("message type %s not supported", msgType)
	}

	daysToExpiration := args[4]
	expiration, err := strconv.Atoi(daysToExpiration)
	if err != nil {
		return fmt.Errorf("invalid days to expiration %s. Only specify the number of days to expire e.g. 30 (for 30 days)", daysToExpiration)
	}

	// Expiration from days to timestamp
	expireTimestamp := time.Now().AddDate(0, 0, expiration).Unix()

	conf, err := loadConfig(flagConfigPath)
	if err != nil {
		return err
	}

	chain, found := conf.GetChain(chainName)
	if !found {
		return fmt.Errorf("chain %s not found in config", chainName)
	}
	key, found := conf.GetKey(keyName)
	if !found {
		return fmt.Errorf("key %s not found in config", keyName)
	}

	nodeAddress := chain.Node
	if flagNode != "" {
		nodeAddress = flagNode
	}

	binary := chain.Binary
	address, err := bech32ify(key.Address, chain.Prefix)
	if err != nil {
		return err
	}

	// Get fees
	fees, err := getFeesParameter(cmd)
	if err != nil {
		return err
	}

	// gaiad tx authz grant
	cmdArgs := []string{"tx", "authz", "grant", grantee, "generic",
		"--expiration", fmt.Sprintf("%d", expireTimestamp),
		"--msg-type", cosmosMsg,
		"--from", address,
		"--fees", fmt.Sprintf("%s%s", fees.Amount.String(), fees.Denom),
		"--gas", fmt.Sprintf("%d", getGas(conf)),
		"--generate-only",
		"--chain-id", fmt.Sprintf("%s", chain.ID),
	}

	if nodeAddress != "" {
		cmdArgs = append(cmdArgs, "--node", nodeAddress)
	}

	execCmd := exec.Command(binary, cmdArgs...)
	fmt.Println(execCmd)
	unsignedBytes, err := execCmd.CombinedOutput()
	if err != nil {
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println("call failed")
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println(execCmd)
		fmt.Println(string(unsignedBytes))
		return err
	}
	fmt.Println(string(unsignedBytes))

	return pushTx(chainName, keyName, unsignedBytes, cmd)
}

func cmdRevokeAuthz(cmd *cobra.Command, args []string) error {
	chainName := args[0]
	keyName := args[1]
	grantee := args[2]

	msgType := args[3]
	// Parse message-type parameter and generate proper tx msg-type
	// Only support the messages we need for now (withdraw, delegate, commission, vote)
	var cosmosMsg string
	switch msgType {
	case "withdraw":
		cosmosMsg = "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward"
	case "delegate":
		cosmosMsg = "/cosmos.staking.v1beta1.MsgDelegate"
	case "commission":
		cosmosMsg = "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission"
	case "vote":
		cosmosMsg = "/cosmos.gov.v1beta1.MsgVote"
	case "unbond":
		cosmosMsg = "/cosmos.staking.v1beta1.MsgUndelegate"
	case "redelegate":
		cosmosMsg = "/cosmos.staking.v1beta1.MsgBeginRedelegate"
	default:
		return fmt.Errorf("message type %s not supported", msgType)
	}

	conf, err := loadConfig(flagConfigPath)
	if err != nil {
		return err
	}

	chain, found := conf.GetChain(chainName)
	if !found {
		return fmt.Errorf("chain %s not found in config", chainName)
	}
	key, found := conf.GetKey(keyName)
	if !found {
		return fmt.Errorf("key %s not found in config", keyName)
	}

	nodeAddress := chain.Node
	if flagNode != "" {
		nodeAddress = flagNode
	}

	binary := chain.Binary
	address, err := bech32ify(key.Address, chain.Prefix)
	if err != nil {
		return err
	}

	// Get fees
	fees, err := getFeesParameter(cmd)
	if err != nil {
		return err
	}

	// gaiad tx authz grant
	cmdArgs := []string{"tx", "authz", "revoke", grantee, cosmosMsg,
		"--from", address,
		"--fees", fmt.Sprintf("%s%s", fees.Amount.String(), fees.Denom),
		"--gas", fmt.Sprintf("%d", getGas(conf)),
		"--generate-only",
		"--chain-id", fmt.Sprintf("%s", chain.ID),
	}

	if nodeAddress != "" {
		cmdArgs = append(cmdArgs, "--node", nodeAddress)
	}

	execCmd := exec.Command(binary, cmdArgs...)
	fmt.Println(execCmd)
	unsignedBytes, err := execCmd.CombinedOutput()
	if err != nil {
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println("call failed")
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println(execCmd)
		fmt.Println(string(unsignedBytes))
		return err
	}
	fmt.Println(string(unsignedBytes))

	return pushTx(chainName, keyName, unsignedBytes, cmd)
}

func cmdVote(cmd *cobra.Command, args []string) error {
	chainName := args[0]
	keyName := args[1]
	propID := args[2]
	voteOption := args[3]

	conf, err := loadConfig(flagConfigPath)
	if err != nil {
		return err
	}

	chain, found := conf.GetChain(chainName)
	if !found {
		return fmt.Errorf("chain %s not found in config", chainName)
	}
	key, found := conf.GetKey(keyName)
	if !found {
		return fmt.Errorf("key %s not found in config", keyName)
	}

	nodeAddress := chain.Node
	if flagNode != "" {
		nodeAddress = flagNode
	}

	binary := chain.Binary
	address, err := bech32ify(key.Address, chain.Prefix)
	if err != nil {
		return err
	}

	// Get fees
	fees, err := getFeesParameter(cmd)
	if err != nil {
		return err
	}

	// gaiad tx gov vote <prop id> <option> --from <from> --generate-only
	cmdArgs := []string{"tx", "gov", "vote", propID, voteOption,
		"--from", address,
		"--fees", fmt.Sprintf("%s%s", fees.Amount.String(), fees.Denom),
		"--gas", fmt.Sprintf("%d", getGas(conf)),
		"--generate-only",
		"--chain-id", fmt.Sprintf("%s", chain.ID),
	}

	if nodeAddress != "" {
		cmdArgs = append(cmdArgs, "--node", nodeAddress)
	}

	execCmd := exec.Command(binary, cmdArgs...)
	fmt.Println(execCmd)
	unsignedBytes, err := execCmd.CombinedOutput()
	if err != nil {
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println("call failed")
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println(execCmd)
		fmt.Println(string(unsignedBytes))
		return err
	}
	fmt.Println(string(unsignedBytes))

	return pushTx(chainName, keyName, unsignedBytes, cmd)
}

func cmdPush(cmd *cobra.Command, args []string) error {
	txFile := args[0]
	chainName := args[1]
	keyName := args[2]

	unsignedBytes, err := ioutil.ReadFile(txFile)
	if err != nil {
		return err
	}

	conf, err := loadConfig(flagConfigPath)
	if err != nil {
		return err
	}

	// Logic to emit a warning if the denoms don't match
	denomInJson, err := parseDenomFromJson(unsignedBytes)
	if err == nil {
		denomConfig, err := getDenom(conf, chainName)
		if err == nil {
			if denomInJson != denomConfig {
				fmt.Printf("WARNING: Denom '%s' in the unsigned json is different from the denom '%s' in the config or registry!\n", denomInJson, denomConfig)
			}
		}
	}

	return pushTx(chainName, keyName, unsignedBytes, cmd)
}

func pushTx(chainName, keyName string, unsignedTxBytes []byte, cmd *cobra.Command) error {

	if flagForce && flagAdditional {
		return fmt.Errorf("Cannot specify both --force and --additional")
	}

	conf, err := loadConfig(flagConfigPath)
	if err != nil {
		return err
	}

	chain, found := conf.GetChain(chainName)
	if !found {
		return fmt.Errorf("chain %s not found in config", chainName)
	}
	key, found := conf.GetKey(keyName)
	if !found {
		return fmt.Errorf("key %s not found in config", keyName)
	}

	//-----------------------------------
	// find account and sequence numbers
	// either from a node and/or from CLI
	//------------------------------------

	nodeAddress := chain.Node
	if flagNode != "" {
		nodeAddress = flagNode
	}

	isAccSet := cmd.Flags().Changed("account")
	isSeqSet := cmd.Flags().Changed("sequence")

	// if both account and sequence are not set, the node must be set in the config or CLI
	noAccOrSeq := !(isAccSet && isSeqSet)
	noNode := nodeAddress == ""
	if noAccOrSeq && noNode {
		fmt.Println("if the --account and --sequence are not provided, a node must be specified in the config or with --node")
		return nil
	}

	var (
		accountNum  int
		sequenceNum int
	)

	// if both account and sequence are not set, get them from the node
	if noAccOrSeq {
		var err error
		binary := chain.Binary
		address, err := bech32ify(key.Address, chain.Prefix)
		if err != nil {
			return err
		}
		accountNum, sequenceNum, err = getAccSeq(binary, address, nodeAddress)
		if err != nil {
			return err
		}
	}

	// if the acc or seq flags are set, overwrite the node
	if isAccSet {
		accountNum = flagAccount
	}
	if isSeqSet {
		sequenceNum = flagSequence
	}

	txDir := filepath.Join(chainName, keyName)

	sess := awsSession(conf.AWS)

	// check if a file already exists
	files, err := awsListFilesInDir(sess, conf.AWS, chainName, keyName)
	if err != nil {
		return err
	}

	// if there is already files there, and we don't specify -f or -x, return
	if len(files) > 0 && !(flagForce || flagAdditional) {
		return fmt.Errorf("Files already exist for %s/%s. Use -f to force overwrite or -x to add additional txs", chainName, keyName)
	} else if len(files) == 0 && (flagForce || flagAdditional) {
		return fmt.Errorf("Path %s/%s is empty, Cannot specify --force or --additional", chainName, keyName)
	}

	// now, either:
	// it is empty, so push files
	// it is not empty, overwrite files (--force)
	// it is not empty, add additional files (--additional)

	// we always start paths with 0, to support multiple txs per chain/key pair
	N := 0

	// if we're pushing additional files, figure out what the highest number is and increment,
	// and add that to the sequence number
	if flagAdditional {
		// figure out what highest number in the files is
		// files should be either "filename.json" or "n/filename.json"
		for _, fullPathFile := range files {
			f := strings.TrimPrefix(fullPathFile, txDir+"/")
			spl := strings.Split(f, "/")
			if len(spl) == 1 {
				continue
			}
			nString := spl[0]
			n, err := strconv.Atoi(nString)
			if err != nil {
				return fmt.Errorf("failed to read number after %s in path %s", txDir, fullPathFile)
			}
			if n > N {
				N = n
			}
		}

		N += 1

		if !isSeqSet {
			sequenceNum += N
		}
	}
	txDir = filepath.Join(txDir, fmt.Sprintf("%d", N))

	// Delete existing files in the path
	err = deleteAllFilesInPath(txDir, conf)
	if err != nil {
		return err
	}

	// create and marshal the sign data
	signData := SignData{
		Account:     accountNum,
		Sequence:    sequenceNum,
		ChainID:     chain.ID,
		Description: flagDescription,
	}
	signDataBytes, err := json.Marshal(signData)
	if err != nil {
		return err
	}

	// upload the unsigned tx
	if err := awsUpload(sess, conf.AWS, txDir, unsignedJSON, unsignedTxBytes); err != nil {
		return err
	}

	// upload the sign data
	if err := awsUpload(sess, conf.AWS, txDir, signDataJSON, signDataBytes); err != nil {
		return err
	}

	fmt.Printf("pushed %s and %s files to %s\n", unsignedJSON, signedJSON, txDir)
	return nil
}

func cmdSign(cobraCmd *cobra.Command, args []string) error {
	/*
		fetch the unsigned tx and signdata
		display the tx and sign data and ask for confirmation from the user
		run the appropriate tx sign command with the right binary using the unsigned tx and metadata
		upload the signature to the right bucket
	*/

	chainName := args[0]
	keyName := args[1]

	from := flagFrom
	txIndex := flagTxIndex

	conf, err := loadConfig(flagConfigPath)
	if err != nil {
		return err
	}

	chain, found := conf.GetChain(chainName)
	if !found {
		return fmt.Errorf("chain %s not found in config", chainName)
	}

	key, found := conf.GetKey(keyName)
	if !found {
		return fmt.Errorf("key %s not found in config", keyName)
	}

	txDir := filepath.Join(chainName, keyName, fmt.Sprintf("%d", txIndex))

	sess := awsSession(conf.AWS)
	downloader := s3manager.NewDownloader(sess)

	// Make a file for the unsigned.json, download it

	unsignedFile, err := ioutil.TempFile("", "temp")
	if err != nil {
		return err
	}
	defer os.Remove(unsignedFile.Name())

	unsignedPath := filepath.Join(txDir, unsignedJSON)
	numBytes, err := downloader.Download(unsignedFile,
		&s3.GetObjectInput{
			Bucket: aws.String(conf.AWS.Bucket),
			Key:    aws.String(unsignedPath),
		})
	if err != nil {
		return err
	}
	_ = numBytes

	// Make a file for sign data, download it

	signDataFile, err := ioutil.TempFile("", "temp")
	if err != nil {
		return err
	}
	defer os.Remove(signDataFile.Name())

	signDataPath := filepath.Join(txDir, signDataJSON)
	numBytes, err = downloader.Download(signDataFile,
		&s3.GetObjectInput{
			Bucket: aws.String(conf.AWS.Bucket),
			Key:    aws.String(signDataPath),
		})
	if err != nil {
		return err
	}
	_ = numBytes

	// TODO: pretty print and confirm the unsigned tx
	unsignedBytes, _ := io.ReadAll(unsignedFile)
	signDataBytes, _ := io.ReadAll(signDataFile)
	fmt.Println("You are signing the following tx:")
	fmt.Println(string(unsignedBytes))
	fmt.Println("With the following sign data:")
	fmt.Println(string(signDataBytes))

	var signData SignData
	if err := json.Unmarshal(signDataBytes, &signData); err != nil {
		return err
	}

	address, err := bech32ify(key.Address, chain.Prefix)
	if err != nil {
		return err
	}
	binary := chain.Binary
	accNum := fmt.Sprintf("%d", signData.Account)
	seqNum := fmt.Sprintf("%d", signData.Sequence)
	chainID := signData.ChainID
	unsignedFileName := unsignedFile.Name()
	backend := conf.KeyringBackend
	user := conf.User

	// gaiad tx sign unsigned.json --multisig <address> --from <from> --account-number <acc> --sequence <seq> --chain-id <id> --offline
	cmdArgs := []string{"tx", "sign", unsignedFileName, "--multisig", address, "--from", from,
		"--account-number", accNum, "--sequence", seqNum, "--chain-id", chainID,
		"--sign-mode", "amino-json",
		"--offline",
	}
	cmdArgs = append(cmdArgs, "--keyring-backend", backend)
	if flagHomePath != "" {
		cmdArgs = append(cmdArgs, "--home", flagHomePath)
	}
	cmd := exec.Command(binary, cmdArgs...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println("call failed")
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println(cmd)
		fmt.Println(string(b))
		return err
	}
	fmt.Println(cmd)
	fmt.Println(string(b))

	// upload the signature as <user>.json
	if err := awsUpload(sess, conf.AWS, txDir, fmt.Sprintf("%s.json", user), b); err != nil {
		return err
	}

	return nil
}

func cmdBroadcast(cobraCmd *cobra.Command, args []string) error {
	chainName := args[0]
	keyName := args[1]

	conf, err := loadConfig(flagConfigPath)
	if err != nil {
		return err
	}

	chain, found := conf.GetChain(chainName)
	if !found {
		return fmt.Errorf("chain %s not found in config", chainName)
	}

	key, found := conf.GetKey(keyName)
	if !found {
		return fmt.Errorf("key %s not found in config", keyName)
	}

	txIndex := flagTxIndex
	txDir := filepath.Join(chainName, keyName, fmt.Sprintf("%d", txIndex))

	sess := awsSession(conf.AWS)
	svc := s3.New(sess)

	// list all items in bucket
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(conf.AWS.Bucket)})
	if err != nil {
		return err
	}

	//--------------------------------
	// txIndex specified must be smallest index for this chainName/keyName pair,
	// otherwise error

	files, err := awsListFilesInDir(sess, conf.AWS, chainName, keyName)
	if err != nil {
		return err
	}

	// see if any indices are smaller than txIndex, and if so, quit
	for _, fullPathFile := range files {
		dirPrefix := filepath.Join(chainName, keyName)
		f := strings.TrimPrefix(fullPathFile, dirPrefix+"/")
		spl := strings.Split(f, "/")
		if len(spl) == 1 {
			continue
		}
		nString := spl[0]
		n, err := strconv.Atoi(nString)
		if err != nil {
			return fmt.Errorf("failed to read number after %s in path %s", txDir, fullPathFile)
		}
		if n < txIndex {
			return fmt.Errorf("found index %d smaller than specified txIndex %d. txs must be broadcast in order", n, txIndex)
		}
	}
	//--------------------------------

	fileNames := []string{}
	for _, item := range resp.Contents {
		itemKey := *item.Key
		if strings.HasPrefix(itemKey, txDir) && !strings.HasSuffix(itemKey, "/") {
			base := filepath.Base(itemKey)

			// sanity check
			if len(base) == 0 {
				return fmt.Errorf("%s had empty base", itemKey)
			}

			fileNames = append(fileNames, base)
		}
	}

	for _, f := range fileNames {
		fmt.Println(f)

		_, err := awsDownload(sess, conf.AWS, txDir, f)
		if err != nil {
			return err
		}
	}

	// get the names of the signatures (everything except unsigned.json and signdata.json)
	sigFileNames := []string{}
	for _, f := range fileNames {
		if f == unsignedJSON || f == signDataJSON {
			continue
		}
		sigFileNames = append(sigFileNames, f)
	}

	// TODO: add this to the key config so its not hardcoded to 2.
	// can default to 2 tho
	threshold := 2
	if len(sigFileNames) < threshold {
		return fmt.Errorf("Insufficient signatures for broadcast. Requires %d, got %d", threshold, len(sigFileNames))
	}

	// read and unmarshal the sign data
	signDataBytes, err := os.ReadFile(signDataJSON)
	if err != nil {
		return err
	}
	var signData SignData
	if err := json.Unmarshal(signDataBytes, &signData); err != nil {
		return err
	}

	// setup for the `tx multisign` command
	binary := chain.Binary
	accNum := fmt.Sprintf("%d", signData.Account)
	seqNum := fmt.Sprintf("%d", signData.Sequence)
	chainID := signData.ChainID
	unsignedFileName := unsignedJSON
	backend := conf.KeyringBackend

	// Check if the local multisig key name was passed as a parameter,
	// if not then check the config file, if also not in the file
	// then return an error
	var localMultisigName string
	if cobraCmd.Flags().Changed("key") {
		localMultisigName = flagMultisigKey
	} else {
		localMultisigName = key.LocalName
	}
	if localMultisigName == "" {
		return fmt.Errorf("localname property for %s key entry in the configuration file is empty", key.Name)
	}

	// gaiad tx multisign unsigned.json <local multisig name> <sig 1> <sig 2> ...  --account-number <acc> --sequence <seq> --chain-id <id>
	cmdArgs := []string{"tx", "multisign", unsignedFileName, localMultisigName}
	for _, sig := range sigFileNames {
		cmdArgs = append(cmdArgs, sig)
	}
	cmdArgs = append(cmdArgs, "--account-number", accNum, "--sequence", seqNum, "--chain-id", chainID, "--offline")
	cmdArgs = append(cmdArgs, "--keyring-backend", backend) // sigh

	nodeAddress := chain.Node
	if flagNode != "" {
		nodeAddress = flagNode
	}
	if nodeAddress != "" {
		cmdArgs = append(cmdArgs, "--node", nodeAddress)
	}
	if flagHomePath != "" {
		cmdArgs = append(cmdArgs, "--home", flagHomePath)
	}

	cmd := exec.Command(binary, cmdArgs...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println("call failed")
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println(cmd)
		fmt.Println(string(b))
		return err
	}
	fmt.Println(cmd)
	fmt.Println(string(b))

	if err := os.WriteFile(signedJSON, b, 0666); err != nil {
		return err
	}

	// broadcast tx
	// TODO: use --broadcast-mode block ?
	// 	  otherwise the tx might still fail when it gets executed but we will delete it
	cmdArgs = []string{"tx", "broadcast", signedJSON, "--node", nodeAddress}
	cmd = exec.Command(binary, cmdArgs...)
	b, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println("call failed")
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println(cmd)
		fmt.Println(string(b))
		return err
	}
	fmt.Println(cmd)
	fmt.Println(string(b))

	code, hash, err := parseTxResult(b)
	if err != nil {
		return err
	}

	// TODO: write the result in a log file
	_, _ = code, hash

	// cleanup txDir in the bucket by deleting everything
	for _, f := range fileNames {
		awsString := aws.String(filepath.Join(txDir, f))
		_, err = svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(conf.AWS.Bucket),
			Key:    awsString,
		})
		if err != nil {
			return err
		}

		err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
			Bucket: aws.String(conf.AWS.Bucket),
			Key:    awsString,
		})
	}

	// Remove all downloaded files and the signed.json
	for _, f := range fileNames {
		err := os.Remove(f)
		if err != nil {
			return err
		}
	}
	if err := os.Remove(signedJSON); err != nil {
		return err
	}

	return nil
}

// TODO can we get a result from the tx without having to parse the tx response? use the API instead of CLI
// Need to parse out the return code and the tx hash
func parseTxResult(txResultBytes []byte) (int, string, error) {
	spl := strings.Split(string(txResultBytes), "\n")
	var (
		code      int
		codeFound bool
		txhash    string
		err       error
	)
	for _, s := range spl {
		if strings.Contains(s, "code:") {
			c := strings.TrimPrefix(s, "code: ")
			code, err = strconv.Atoi(c)
			if err != nil {
				return 0, "", fmt.Errorf("code in tx response is not an integer")
			}
			codeFound = true

		} else if strings.Contains(s, "txhash:") {
			txhash = strings.TrimPrefix(s, "txhash: ")
		}
	}

	if !codeFound {
		return 0, "", fmt.Errorf("couldn't find code in tx response")
	} else if txhash == "" {
		return 0, "", fmt.Errorf("couldn't find txhash in tx response")
	}
	return code, txhash, nil
}

// Parse out the account and sequence number
// Return: accountNumber, sequenceNumber, error
func parseAccountQuery(queryResponseBytes []byte) (int, int, error) {
	var (
		accInt   int = 0
		seqInt   int = 0
		acctType AccountType
	)

	json.Unmarshal(queryResponseBytes, &acctType)

	if acctType.Type == "/cosmos.auth.v1beta1.BaseAccount" {
		var ba BaseAccount
		err := json.Unmarshal(queryResponseBytes, &ba)
		if err != nil {
			return 0, 0, fmt.Errorf("error un-marshalling base account")
		}
		accInt, err := strconv.Atoi(ba.AccountNumber)
		if err != nil {
			return 0, 0, fmt.Errorf("account number is not an integer")
		}
		seqInt, err := strconv.Atoi(ba.Sequence)
		if err != nil {
			return 0, 0, fmt.Errorf("sequence number is not an integer")
		}
		return accInt, seqInt, nil
	} else if acctType.Type == "/cosmos.vesting.v1beta1.PeriodicVestingAccount" {
		var pva PeriodicVestingAccount
		err := json.Unmarshal(queryResponseBytes, &pva)
		if err != nil {
			return 0, 0, fmt.Errorf("error un-marshalling periodic vesting account")
		}
		accInt, err := strconv.Atoi(pva.BaseVestingAccount.BaseAccount.AccountNumber)
		if err != nil {
			return 0, 0, fmt.Errorf("account number is not an integer")
		}
		seqInt, err := strconv.Atoi(pva.BaseVestingAccount.BaseAccount.Sequence)
		if err != nil {
			return 0, 0, fmt.Errorf("sequence number is not an integer")
		}
		return accInt, seqInt, nil
	} else if acctType.Type == "/cosmos.vesting.v1beta1.ContinuousVestingAccount" {
		var cva ContinuousVestingAccount
		err := json.Unmarshal(queryResponseBytes, &cva)
		if err != nil {
			return 0, 0, fmt.Errorf("error un-marshalling continuous vesting account")
		}
		accInt, err := strconv.Atoi(cva.BaseVestingAccount.BaseAccount.AccountNumber)
		if err != nil {
			return 0, 0, fmt.Errorf("account number is not an integer")
		}
		seqInt, err := strconv.Atoi(cva.BaseVestingAccount.BaseAccount.Sequence)
		if err != nil {
			return 0, 0, fmt.Errorf("sequence number is not an integer")
		}
		return accInt, seqInt, nil
	} else if acctType.Type == "/stride.vesting.StridePeriodicVestingAccount" {
		var spva StridePeriodicVestingAccount
		err := json.Unmarshal(queryResponseBytes, &spva)
		if err != nil {
			return 0, 0, fmt.Errorf("error un-marshalling Stride Periodic Vesting Account account")
		}
		accInt, err := strconv.Atoi(spva.BaseVestingAccount.BaseAccount.AccountNumber)
		if err != nil {
			return 0, 0, fmt.Errorf("account number is not an integer")
		}
		seqInt, err := strconv.Atoi(spva.BaseVestingAccount.BaseAccount.Sequence)
		if err != nil {
			return 0, 0, fmt.Errorf("sequence number is not an integer")
		}
		return accInt, seqInt, nil
	} else if strings.Contains(acctType.Type, "EthAccount") {
		var eth EthAccount
		err := json.Unmarshal(queryResponseBytes, &eth)
		if err != nil {
			return 0, 0, fmt.Errorf("error un-marshalling ethermint account")
		}
		accInt, err := strconv.Atoi(eth.BaseAccount.AccountNumber)
		if err != nil {
			return 0, 0, fmt.Errorf("account number is not an integer")
		}
		seqInt, err := strconv.Atoi(eth.BaseAccount.Sequence)
		if err != nil {
			return 0, 0, fmt.Errorf("sequence number is not an integer")
		}
		return accInt, seqInt, nil
	} else {
		return accInt, seqInt, fmt.Errorf("cannot parse account type")
	}
}

// Return: accountNumber, sequenceNumber, error
func getAccSeq(binary, addr, node string) (int, int, error) {
	cmdArgs := []string{"query", "--node", node, "account", addr, "--output", "json"}
	cmd := exec.Command(binary, cmdArgs...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println("call failed")
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println(cmd)
		fmt.Println(string(b))
		return 0, 0, err
	}
	fmt.Println(cmd)
	fmt.Println(string(b))

	return parseAccountQuery(b)
}

// Get account balance for a particular denom
func getAccountBalance(address string, denom string, chain Chain) (math.Int, error) {

	// [binary] query account balances <account> --output json
	cmdArgs := []string{"query", "bank", "balances",
		"--output", "json",
		"--node", chain.Node,
		address,
	}

	execCmd := exec.Command(chain.Binary, cmdArgs...)
	fmt.Println(execCmd)
	unsignedBytes, err := execCmd.CombinedOutput()
	if err != nil {
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println("call failed")
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println(execCmd)
		fmt.Println(string(unsignedBytes))
		return math.ZeroInt(), err
	}

	var ab AccountBalance
	err = json.Unmarshal(unsignedBytes, &ab)
	if err != nil {
		return math.ZeroInt(), fmt.Errorf("cannot parse balance")
	}
	for _, balance := range ab.Balances {
		if strings.ToLower(balance.Denom) == strings.ToLower(denom) {
			amount, ok := math.NewIntFromString(balance.Amount)
			if !ok {
				return math.ZeroInt(), fmt.Errorf("cannot parse balance for %s denom", denom)
			}
			return amount, nil
		}
	}
	return math.ZeroInt(), fmt.Errorf("cannot find balance for %s denom", denom)
}

func getGas(config *Config) int64 {
	if flagGas > 0 {
		return int64(flagGas)
	} else {
		if config.DefaultGas > 0 {
			return config.DefaultGas
		} else {
			return int64(defaultGas)
		}

	}
}

func getFeesParameter(cmd *cobra.Command) (sdk.DecCoin, error) {
	if cmd.Flags().Changed("fees") {
		decCoin, err := sdk.ParseDecCoin(flagFees)
		if err != nil {
			return sdk.DecCoin{}, fmt.Errorf("error parsing the '--fees' parameter, please specify the amount and denom, e.g. 100uatom")
		}
		return decCoin, nil
	} else {
		return sdk.DecCoin{}, fmt.Errorf("please specify the '--fees' parameter")
	}
}

func listFilesInPath(svc *s3.S3, conf *Config, txDir string) ([]string, error) {
	// list all items in bucket
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(conf.AWS.Bucket)})
	if err != nil {
		return nil, err
	}

	fileNames := []string{}
	for _, item := range resp.Contents {
		itemKey := *item.Key
		if strings.HasPrefix(itemKey, txDir) && !strings.HasSuffix(itemKey, "/") {
			base := filepath.Base(itemKey)

			// sanity check
			if len(base) == 0 {
				return nil, fmt.Errorf("%s had empty base", itemKey)
			}

			fileNames = append(fileNames, base)
		}
	}
	return fileNames, nil
}

func deleteAllFilesInPath(txDir string, conf *Config) error {

	sess := awsSession(conf.AWS)
	svc := s3.New(sess)

	fileNames, err := listFilesInPath(svc, conf, txDir)
	if err != nil {
		return err
	}

	// Check if there is anything in the bucket, if not then return
	if len(fileNames) == 0 {
		fmt.Printf("no existing files in %s, nothing will be deleted\n", txDir)
		return nil
	} else {
		fmt.Printf("found %d files in %s, deleting files...\n", len(fileNames), txDir)
	}

	sep := "---------------------------------------------------------------------"
	fmt.Println(sep)

	// cleanup txDir in the bucket by deleting everything
	for _, f := range fileNames {
		awsString := aws.String(filepath.Join(txDir, f))
		_, err = svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(conf.AWS.Bucket),
			Key:    awsString,
		})
		if err != nil {
			return err
		} else {
			fmt.Printf("%v will be deleted...\n", *awsString)
		}

		err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
			Bucket: aws.String(conf.AWS.Bucket),
			Key:    awsString,
		})
		if err != nil {
			return err
		} else {
			fmt.Printf("deleted %v !\n", *awsString)
			fmt.Println(sep)
		}
	}
	return nil
}
