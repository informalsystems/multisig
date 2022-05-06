package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func cmdRawBech32(cobraCmd *cobra.Command, args []string) error {
	bech32String := args[0]
	bech32Prefix := args[1]
	newbech32String, err := bech32ify(bech32String, bech32Prefix)
	if err != nil {
		return err
	}
	fmt.Println(newbech32String)
	return nil
}

// copy a local file to the bucket
func cmdRawUp(cobraCmd *cobra.Command, args []string) error {
	local := args[0]
	remote := args[1]

	conf, err := loadConfig(configFile)
	if err != nil {
		return err
	}
	sess := awsSession(conf.AWS)

	// read the local file
	localBytes, err := ioutil.ReadFile(local)
	if err != nil {
		return err
	}

	// upload it
	dir := filepath.Dir(remote)
	fileName := filepath.Base(remote)
	if err := awsUpload(sess, conf.AWS, dir, fileName, localBytes); err != nil {
		return err
	}
	return nil
}

// copy a file from the bucket to the local machine
func cmdRawDown(cobraCmd *cobra.Command, args []string) error {
	remote := args[0]
	local := args[1]

	conf, err := loadConfig(configFile)
	if err != nil {
		return err
	}
	sess := awsSession(conf.AWS)

	// if remote ends in /, fetch the whole directory and return
	if strings.HasSuffix(remote, "/") {
		if err := os.Mkdir(local, 0777); err != nil {
			return err
		}
		if err := os.Chdir(local); err != nil {
			return err
		}
		_, err = awsDownloadFilesInDir(sess, conf.AWS, remote)
		return err
	}

	// otherwise, just download the one file
	dir := filepath.Dir(remote)
	fileName := filepath.Base(remote)
	file, err := awsDownload(sess, conf.AWS, dir, fileName)
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
func cmdRawCat(cobraCmd *cobra.Command, args []string) error {
	chainName := args[0]
	keyName := args[1]

	conf, err := loadConfig(configFile)
	if err != nil {
		return err
	}

	sess := awsSession(conf.AWS)

	txDir := filepath.Join(chainName, keyName)

	// Get all the files in S3 bucket and display the contents
	// of the files
	filesInS3, err := awsListFilesInDir(sess, conf.AWS, chainName, keyName)
	if err != nil {
		return err
	} else {
		if len(filesInS3) == 0 {
			fmt.Println("No files in", txDir)
			return nil
		}
		for _, file := range filesInS3 {
			parts := strings.Split(file, "/")
			path := strings.Join([]string{parts[0], parts[1], parts[2]}, "/")
			filename := parts[3]
			f, err := awsDownload(sess, conf.AWS, path, filename)
			if err != nil {
				return err
			} else {
				b, err := ioutil.ReadFile(f.Name())
				if err != nil {
					return err
				}
				fmt.Printf("\n|------------| %s |------------|", file)
				fmt.Println("\n")
				fmt.Println(string(b))
				os.Remove(f.Name())
			}
		}
	}

	return nil
}

// delete a file from the bucket
func cmdRawDelete(cobraCmd *cobra.Command, args []string) error {
	filePath := args[0]

	conf, err := loadConfig(configFile)
	if err != nil {
		return err
	}

	sess := awsSession(conf.AWS)
	awsDelete(sess, conf.AWS, filePath)
	return nil
}

// create an empty object with the given name
func cmdRawMkdir(cobraCmd *cobra.Command, args []string) error {
	dirName := args[0]
	if !strings.HasSuffix(dirName, "/") {
		fmt.Println("directory paths must end with a '/'")
		return nil
	}

	conf, err := loadConfig(configFile)
	if err != nil {
		return err
	}

	sess := awsSession(conf.AWS)
	awsMkdir(sess, conf.AWS, dirName)
	return nil
}
