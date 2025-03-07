package service

import (
	"fmt"

	"github.com/alejo-lapix/multimedia-go/files"
	"github.com/alejo-lapix/multimedia-go/persistence"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

type Uploader interface {
	// Upload uploads a file and store a file to the given provider
	Upload(filename *string, destination *string) (*persistence.MultimediaItem, error)
	// Delete deletes a file from the given provider and database
	Delete(ID *string) error
}

type AWSUploader struct {
	Bucket     *string
	Region     *string
	Repository persistence.BasicRepository
	Storage    files.Provider
}

type InvalidArgumentError struct {
	Message string
}

func (err InvalidArgumentError) Error() string {
	return err.Message
}

func NewAWSUploader(tableName, bucket, region *string) (*AWSUploader, error) {
	if *tableName == "" || *bucket == "" || *region == "" {
		return nil, InvalidArgumentError{Message: "TableName, Bucket or Region can no be empty"}
	}

	sess, err := session.NewSession(&aws.Config{Region: region})

	if err != nil {
		return nil, err
	}

	repository, err := persistence.NewDynamoDBRepository(tableName, dynamodb.New(sess))

	if err != nil {
		return nil, err
	}

	storage := files.NewAWSProvider(bucket, s3.New(sess))

	return &AWSUploader{
		Bucket:     bucket,
		Region:     region,
		Repository: repository,
		Storage:    storage,
	}, nil
}

func (uploader *AWSUploader) Upload(filename, destination *string) (*persistence.MultimediaItem, error) {
	urlRegion := ""

	if *uploader.Region != "us-east-1" {
		urlRegion = fmt.Sprintf("-%v", *uploader.Region)
	}

	bucket := fmt.Sprintf("https://%v.s3%v.amazonaws.com", *uploader.Bucket, urlRegion)
	fileType, err := getFileType(filename)

	if err != nil {
		return nil, err
	}

	item, err := persistence.NewMultimediaItem(&bucket, destination, fileType)

	if err != nil {
		return nil, err
	}

	err = uploader.Storage.Store(filename, destination)

	if err != nil {
		return nil, err
	}

	return item, nil
}

func getFileType(filename *string) (*string, error) {
	image := persistence.IMAGE

	return &image, nil
}

type NotFoundError struct {
	Message string
}

func (err NotFoundError) Error() string {
	return err.Message
}

func (uploader *AWSUploader) Delete(ID *string) error {
	return uploader.Storage.Remove(ID)
}
