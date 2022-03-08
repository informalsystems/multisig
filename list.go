package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func cmdList(cmd *cobra.Command, args []string) error {
	if flagAll {
		return listAll()
	}
	return listDir(args)
}

func listAll() error {
	conf, err := loadConfig(configFile)
	if err != nil {
		return err
	}
	sess := awsSession(conf.AWS)
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

func listDir(args []string) error {
	if len(args) != 2 {
		fmt.Println("must specify args: <chain name> <key name>")
		return nil
	}
	chainName := args[0]
	keyName := args[1]

	conf, err := loadConfig(configFile)
	if err != nil {
		return err
	}
	sess := awsSession(conf.AWS)
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
