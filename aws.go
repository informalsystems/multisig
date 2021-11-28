package main

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// setup a new aws session
func awsSession(pub, priv string) *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("ca-central-1"),
		Credentials: credentials.NewStaticCredentials(pub, priv, ""),
	})
	if err != nil {
		// TODO
		panic(err)
	}
	return sess
}

func awsDownload(sess *session.Session, bucketName, txDir, name string) (*os.File, error) {

	file, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	downloader := s3manager.NewDownloader(sess)

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(filepath.Join(txDir, name)),
		})
	if err != nil {
		return nil, err
	}
	_ = numBytes

	return file, nil
}

func awsUpload(sess *session.Session, bucketName, txDir, name string, dataBytes []byte) error {
	uploader := s3manager.NewUploader(sess)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filepath.Join(txDir, name)),
		Body:   bytes.NewBuffer(dataBytes),
	})

	return err
}
