package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// global vars
var (

	// config file
	configFile = "config.toml"

	// files for signing
	unsignedJSON = "unsigned.json"
	signedJSON   = "signed.json"
	signDataJSON = "signdata.json"
)

// Data we need for signers to sign a tx (eg. without access to a node)
type SignData struct {
	Account  int    `json:"account"`
	Sequence int    `json:"sequence"`
	ChainID  string `json:"chain-id"`
}

func main() {
	app := &cli.App{
		Name:   "multisig",
		Usage:  "manage multisig transactions",
		Action: cli.ShowAppHelp,
		Commands: []*cli.Command{
			{
				Name:  "generate",
				Usage: "generate a new unsigned tx",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tx",
						Value:    "",
						Usage:    "unsigned tx file",
						Required: true,
					},
					&cli.IntFlag{
						Name:     "sequence",
						Aliases:  []string{"s"},
						Value:    0,
						Usage:    "sequence number for the tx",
						Required: true,
					},
					&cli.IntFlag{
						Name:     "account",
						Aliases:  []string{"a"},
						Value:    0,
						Usage:    "account number for the tx",
						Required: true,
					},
				},
				Action:    cmdGenerate,
				ArgsUsage: "<chain name> <key name>",
			},
			{
				Name:      "sign",
				Usage:     "sign a tx",
				Action:    cmdSign,
				ArgsUsage: "<chain name> <key name>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "from",
						Value:    "",
						Usage:    "name of your local key to sign with",
						Required: true,
					},
				},
			},
			{
				Name:      "list",
				Usage:     "list items in a directory",
				Action:    cmdList,
				ArgsUsage: "<chain name> <key name>",
			},
			{
				Name:      "broadcast",
				Usage:     "broadcast a tx",
				Action:    cmdBroadcast,
				ArgsUsage: "<chain name> <key name>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "node",
						Value: "",
						Usage: "node address to broadcast too. flag overrides config",
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

func cmdGenerate(c *cli.Context) error {
	args := c.Args()
	first, tail := args.First(), args.Tail()
	if len(tail) < 1 {
		fmt.Println("must specify args:", c.Command.ArgsUsage)
		return nil
	}
	chainName := first
	keyName := tail[0]

	accountNum := c.Int("account")
	sequenceNum := c.Int("sequence")

	conf, err := loadConfig(configFile)
	if err != nil {
		return err
	}

	chain, found := conf.GetChain(chainName)
	if !found {
		return fmt.Errorf("chain %s not found in config", chainName)
	}

	// read the unsigned tx file
	txFile := c.String("tx")
	unsignedBytes, err := ioutil.ReadFile(txFile)
	if err != nil {
		return err
	}

	// create and marshal the sign data
	signData := SignData{
		Account:  accountNum,
		Sequence: sequenceNum,
		ChainID:  chain.ID,
	}
	signDataBytes, err := json.Marshal(signData)
	if err != nil {
		return err
	}
	txDir := filepath.Join(chainName, keyName)

	sess := awsSession(conf.AWS.Pub, conf.AWS.Priv)

	// upload the unsigned tx
	if err := awsUpload(sess, conf.AWS.Bucket, txDir, unsignedJSON, unsignedBytes); err != nil {
		return err
	}

	// upload the sign data
	if err := awsUpload(sess, conf.AWS.Bucket, txDir, signDataJSON, signDataBytes); err != nil {
		return err
	}

	return nil
}

func cmdList(c *cli.Context) error {
	args := c.Args()
	first, tail := args.First(), args.Tail()
	if len(tail) < 1 {
		fmt.Println("must specify args:", c.Command.ArgsUsage)
		return nil
	}
	chainName := first
	keyName := tail[0]

	conf, err := loadConfig(configFile)
	if err != nil {
		return err
	}
	sess := awsSession(conf.AWS.Pub, conf.AWS.Priv)
	svc := s3.New(sess)
	filePath := filepath.Join(chainName, keyName)

	// list all items in bucket
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(conf.AWS.Bucket)})
	if err != nil {
		return err
	}

	files := []string{}
	for _, item := range resp.Contents {
		key := *item.Key
		if strings.HasPrefix(key, filePath) {
			files = append(files, key)
		}
	}

	for _, f := range files {
		fmt.Println(f)
	}

	return nil
}

func cmdSign(c *cli.Context) error {
	/*
		fetch the unsigned tx and signdata
		display the tx and sign data and ask for confirmation from the user
		run the appropriate tx sign command with the right binary using the unsigned tx and metadata
		upload the signature to the right bucket
	*/

	args := c.Args()
	first, tail := args.First(), args.Tail()
	if len(tail) < 1 {
		fmt.Println("must specify args:", c.Command.ArgsUsage)
		return nil
	}
	chainName := first
	keyName := tail[0]

	from := c.String("from")

	conf, err := loadConfig(configFile)
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

	txDir := filepath.Join(chainName, keyName)

	sess := awsSession(conf.AWS.Pub, conf.AWS.Priv)
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
	unsignedBytes, _ := ioutil.ReadAll(unsignedFile)
	signDataBytes, _ := ioutil.ReadAll(signDataFile)
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
	if err := awsUpload(sess, conf.AWS.Bucket, txDir, fmt.Sprintf("%s.json", user), b); err != nil {
		return err
	}

	return nil
}

func cmdBroadcast(c *cli.Context) error {
	args := c.Args()
	first, tail := args.First(), args.Tail()
	if len(tail) < 1 {
		fmt.Println("must specify args:", c.Command.ArgsUsage)
		return nil
	}
	chainName := first
	keyName := tail[0]

	conf, err := loadConfig(configFile)
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

	sess := awsSession(conf.AWS.Pub, conf.AWS.Priv)
	svc := s3.New(sess)
	txDir := filepath.Join(chainName, keyName)

	// list all items in bucket
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(conf.AWS.Bucket)})
	if err != nil {
		return err
	}

	fileNames := []string{}
	for _, item := range resp.Contents {
		itemKey := *item.Key
		if strings.HasPrefix(itemKey, txDir) && !strings.HasSuffix(itemKey, "/") {
			fileNames = append(fileNames, filepath.Base(itemKey))
		}
	}

	for _, f := range fileNames {
		fmt.Println(f)

		_, err := awsDownload(sess, conf.AWS.Bucket, txDir, f)
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

	// read and unmarshal the sign data
	signDataBytes, err := ioutil.ReadFile(signDataJSON)
	if err != nil {
		return err
	}
	var signData SignData
	if err := json.Unmarshal(signDataBytes, &signData); err != nil {
		return err
	}

	binary := chain.Binary
	accNum := fmt.Sprintf("%d", signData.Account)
	seqNum := fmt.Sprintf("%d", signData.Sequence)
	chainID := signData.ChainID
	unsignedFileName := unsignedJSON
	backend := conf.KeyringBackend
	// TODO: can I used the address directly here instead ?
	localMultisigName := key.LocalName

	// gaiad tx multisign unsigned.json <local multisig name> <sig 1> <sig 2> ...  --account-number <acc> --sequence <seq> --chain-id <id>
	cmdArgs := []string{"tx", "multisign", unsignedFileName, localMultisigName}
	for _, sig := range sigFileNames {
		cmdArgs = append(cmdArgs, sig)
	}
	cmdArgs = append(cmdArgs, "--account-number", accNum, "--sequence", seqNum, "--chain-id", chainID, "--offline")
	cmdArgs = append(cmdArgs, "--keyring-backend", backend) // sigh
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

	if err := ioutil.WriteFile(signedJSON, b, 0666); err != nil {
		return err
	}

	nodeAddress := chain.Node
	if c.String("node") != "" {
		nodeAddress = c.String("node")
	}

	// broadcast tx
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
