package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// global vars
var (

	// config file - expected to be in the present working directory
	configFile = "config.toml"

	// files for signing - we use these filenames in the local working directory and in the remote bucket
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
						Name:    "sequence",
						Aliases: []string{"s"},
						Value:   0,
						Usage:   "sequence number for the tx",
						// Required: true,
					},
					&cli.IntFlag{
						Name:    "account",
						Aliases: []string{"a"},
						Value:   0,
						Usage:   "account number for the tx",
						//Required: true,
					},
					&cli.StringFlag{
						Name:    "node",
						Aliases: []string{"n"},
						Value:   "",
						Usage:   "tendermint rpc node to get sequence and account number from",
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
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "all",
						Value: false,
						Usage: "list files for all chains and keys",
					},
				},
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
					&cli.StringFlag{
						Name:     "description",
						Aliases:  []string{"d"},
						Value:    "",
						Usage:    "description of the tx to be logged",
						Required: true,
					},
				},
			},
			{
				Name:  "raw",
				Usage: "raw operations on the s3 bucket",
				// Action:      cmdRaw,
				Subcommands: []*cli.Command{
					{
						Name:      "bech32",
						Usage:     "convert a bech32 string to a different prefix",
						ArgsUsage: "<bech32 string> <new prefix>",
						Action:    cmdRawBech32,
					},
					{
						Name:      "cat",
						Usage:     "dump the contents of all files in a directory",
						ArgsUsage: "<chain name> <key name>",
						Action:    cmdRawCat,
					},
					{
						Name:      "up",
						Usage:     "upload a local file to a path in the s3 bucket",
						ArgsUsage: "<source filepath> <destination filepath>",
						Action:    cmdRawUp,
					},
					{
						Name:      "down",
						Usage:     "download a file or directory from the s3 bucket",
						UsageText: "if the path ends in a '/' it will attempt to download all files in that directory",
						ArgsUsage: "<source filepath> <destination filepath>",
						Action:    cmdRawDown,
					},
					{
						Name:      "mkdir",
						Usage:     "create a directory in the s3 bucket - must end with a '/'",
						UsageText: "note there are no directories in s3, just empty objects that end with a '/'",
						ArgsUsage: "<directory path>",
						Action:    cmdRawMkdir,
					},
					{
						Name:      "delete",
						Usage:     "delete a file from the s3 bucket",
						ArgsUsage: "<filepath>",
						Action:    cmdRawDelete,
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

func cmdRawBech32(c *cli.Context) error {
	args := c.Args()
	first, tail := args.First(), args.Tail()
	if len(tail) < 1 {
		fmt.Println("must specify args:", c.Command.ArgsUsage)
		return nil
	}

	bech32String := first
	bech32Prefix := tail[0]
	newbech32String, err := bech32ify(bech32String, bech32Prefix)
	if err != nil {
		return err
	}
	fmt.Println(newbech32String)
	return nil
}

// copy a local file to the bucket
func cmdRawUp(c *cli.Context) error {
	args := c.Args()
	first, tail := args.First(), args.Tail()
	if len(tail) < 1 {
		fmt.Println("must specify args:", c.Command.ArgsUsage)
		return nil
	}

	local := first
	remote := tail[0]

	conf, err := loadConfig(configFile)
	if err != nil {
		return err
	}
	sess := awsSession(conf.AWS.Pub, conf.AWS.Priv)

	// read the local file
	localBytes, err := ioutil.ReadFile(local)
	if err != nil {
		return err
	}

	// upload it
	dir := filepath.Dir(remote)
	fileName := filepath.Base(remote)
	if err := awsUpload(sess, conf.AWS.Bucket, dir, fileName, localBytes); err != nil {
		return err
	}
	return nil
}

// copy a local file to the bucket
func cmdRawDown(c *cli.Context) error {
	args := c.Args()
	first, tail := args.First(), args.Tail()
	if len(tail) < 1 {
		fmt.Println("must specify args:", c.Command.ArgsUsage)
		return nil
	}

	remote := first
	local := tail[0]

	conf, err := loadConfig(configFile)
	if err != nil {
		return err
	}
	sess := awsSession(conf.AWS.Pub, conf.AWS.Priv)

	// if remote ends in /, fetch the whole directory and return
	if strings.HasSuffix(remote, "/") {
		if err := os.Mkdir(local, 0777); err != nil {
			return err
		}
		if err := os.Chdir(local); err != nil {
			return err
		}
		_, err = fetchFilesInDir(sess, conf.AWS.Bucket, remote)
		return err
	}

	// otherwise, just download the one file
	dir := filepath.Dir(remote)
	fileName := filepath.Base(remote)
	file, err := awsDownload(sess, conf.AWS.Bucket, dir, fileName)
	if err != nil {
		return err
	}

	// rename if necessary
	oldName := file.Name()
	newName := local
	if oldName != newName {
		return os.Rename(oldName, newName)
	}
	return nil
}

// dump content of all files in a dir
func cmdRawCat(c *cli.Context) error {
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

	txDir := filepath.Join(chainName, keyName)

	files, err := fetchFilesInDir(sess, conf.AWS.Bucket, txDir)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		fmt.Println("No files in", txDir)
		return nil
	}

	fmt.Println("") // for spacing
	for _, f := range files {
		// cat the file
		b, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}
		fmt.Printf("---------- %s ----------\n", f)
		fmt.Println("")
		fmt.Println(string(b))
		fmt.Println("")
		os.Remove(f)
	}

	return nil
}

// copy a local file to the bucket
func cmdRawDelete(c *cli.Context) error {
	args := c.Args()
	filePath := args.First()
	if filePath == "" {
		fmt.Println("must specify args:", c.Command.ArgsUsage)
		return nil
	}

	conf, err := loadConfig(configFile)
	if err != nil {
		return err
	}

	sess := awsSession(conf.AWS.Pub, conf.AWS.Priv)
	awsDelete(sess, conf.AWS.Bucket, filePath)
	return nil
}

// create an empty object with the given name
func cmdRawMkdir(c *cli.Context) error {
	args := c.Args()
	dirName := args.First()
	if dirName == "" {
		fmt.Println("must specify args:", c.Command.ArgsUsage)
		return nil
	} else if !strings.HasSuffix(dirName, "/") {
		fmt.Println("directory paths must end with a '/'")
		return nil
	}

	conf, err := loadConfig(configFile)
	if err != nil {
		return err
	}

	sess := awsSession(conf.AWS.Pub, conf.AWS.Priv)
	awsMkdir(sess, conf.AWS.Bucket, dirName)
	return nil
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

	//-----------------------------------
	// find account and sequence numbers
	// either from a node and/or from CLI
	//------------------------------------

	nodeAddress := chain.Node
	if c.String("node") != "" {
		nodeAddress = c.String("node")
	}

	isAccSet := c.IsSet("account")
	isSeqSet := c.IsSet("sequence")

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
		accountNum = c.Int("account")
	}
	if isSeqSet {
		sequenceNum = c.Int("sequence")
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
	if c.Bool("all") {
		return listAll(c)
	}
	return listDir(c)
}

func listAll(c *cli.Context) error {
	conf, err := loadConfig(configFile)
	if err != nil {
		return err
	}
	sess := awsSession(conf.AWS.Pub, conf.AWS.Priv)
	svc := s3.New(sess)

	// list all items in bucket
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(conf.AWS.Bucket)})
	if err != nil {
		return err
	}

	files := []string{}
	for _, item := range resp.Contents {
		key := *item.Key
		files = append(files, key)
	}

	last := ""
	sep := "---------------------------------"
	fmt.Println(sep)
	for _, f := range files {
		fDir := filepath.Dir(f)
		if fDir != filepath.Dir(last) && !strings.HasPrefix(fDir, last) {
			fmt.Println(sep)
		}
		fmt.Println(f)
		last = f
	}
	fmt.Println(sep)

	return nil

}

func listDir(c *cli.Context) error {
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

	// setup for the `tx multisign` command
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
	// TODO: use --broadcast-mode block ?
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

// TODO can we get a programmatic result from the tx?
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

// TODO can we get this more programmatically ?
// Need to parse out the account and sequence number
// Return: accountNumber, sequenceNumber, error
func parseAccountQuery(queryResponseBytes []byte) (int, int, error) {
	spl := strings.Split(string(queryResponseBytes), "\n")
	var (
		acc string
		seq string
	)
	for _, s := range spl {
		if strings.Contains(s, "account_number:") {
			c := strings.TrimPrefix(s, `account_number:`)
			c = strings.TrimSpace(c)
			c = strings.TrimPrefix(c, `"`)
			c = strings.TrimSuffix(c, `"`)
			acc = c
		} else if strings.Contains(s, "sequence:") {
			c := strings.TrimPrefix(s, `sequence:`)
			c = strings.TrimSpace(c)
			c = strings.TrimPrefix(c, `"`)
			c = strings.TrimSuffix(c, `"`)
			seq = c
		}
	}

	accInt, err := strconv.Atoi(acc)
	if err != nil {
		return 0, 0, fmt.Errorf("account number in query response is not an integer")
	}

	seqInt, err := strconv.Atoi(seq)
	if err != nil {
		return 0, 0, fmt.Errorf("sequence in query response is not an integer")
	}

	return accInt, seqInt, nil
}

// Return: accountNumber, sequenceNumber, error
func getAccSeq(binary, addr, node string) (int, int, error) {
	cmdArgs := []string{"query", "--node", node, "account", addr}
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
