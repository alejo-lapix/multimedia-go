package files

import (
	"bytes"
	"io"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Client interface {
	PutObject(*s3.PutObjectInput) (*s3.PutObjectOutput, error)
	GetObject(*s3.GetObjectInput) (*s3.GetObjectOutput, error)
	DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
}

type Provider interface {
	Store(currentPath *string, newPath *string) error
	Read(path *string) ([]byte, error)
	Remove(filename *string) error
}

type FileOpener interface {
	Open(string) (*os.File, error)
}

type AWSProvider struct {
	S3     S3Client
	Opener FileOpener
	Bucket *string
}

// NewProvider return a new AWSProvider
func NewAWSProvider(bucket *string, s3 *s3.S3) *AWSProvider {
	return &AWSProvider{
		S3:     s3,
		Opener: &OSFileOpener{},
		Bucket: bucket,
	}
}

type OSFileOpener struct{}

func (opener *OSFileOpener) Open(filename string) (*os.File, error) {
	return os.Open(filename)
}

// Store put an object in the given S3 Bucket
func (provider *AWSProvider) Store(filename *string, destination *string) error {
	file, err := provider.Opener.Open(*filename)

	if err != nil {
		return err
	}

	defer file.Close()

	fileInfo, err := file.Stat()

	if err != nil {
		return err
	}

	var size = fileInfo.Size()
	buffer := make([]byte, size)
	_, err = file.Read(buffer)

	if err != nil {
		return err
	}

	_, err = provider.S3.PutObject(&s3.PutObjectInput{
		Bucket:        provider.Bucket,
		Key:           destination,
		Body:          bytes.NewReader(buffer),
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(http.DetectContentType(buffer)),
		// TODO This parameters must be dynamic, maybe permissions
		ACL:                  aws.String("public-read"),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})

	return err
}

// Read reads an element from aws
func (provider *AWSProvider) Read(path *string) ([]byte, error) {
	output, err := provider.S3.GetObject(&s3.GetObjectInput{
		Bucket: provider.Bucket,
		Key:    path,
	})

	if err != nil {
		return nil, err
	}

	buffer := make([]byte, *output.ContentLength)
	_, err = output.Body.Read(buffer)

	if err == io.EOF || err == nil {
		return buffer, nil
	}

	buffer = nil

	return buffer, err
}

func (provider *AWSProvider) Remove(filename *string) error {
	_, err := provider.S3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: provider.Bucket,
		Key:    filename,
	})

	return err
}
