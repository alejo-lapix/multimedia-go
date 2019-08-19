package files

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"net/http"
	"os"
)

type S3Client interface {
	PutObject(*s3.PutObjectInput) (*s3.PutObjectOutput, error)
	GetObject(*s3.GetObjectInput) (*s3.GetObjectOutput, error)
}

type Provider interface {
	Store(currentPath *string, newPath *string) (bool, error)
	Read(path *string) ([]byte, error)
}

type FileOpener interface {
	Open(string) (*os.File, error)
}

type AwsProvider struct {
	S3     S3Client
	Opener FileOpener
	Bucket *string
}

// NewProvider return a new AwsProvider
func NewAwsProvider(bucket *string, s3 *s3.S3) *AwsProvider {
	return &AwsProvider{
		S3:     s3,
		Opener: &OsFileOpener{},
		Bucket: bucket,
	}
}

type OsFileOpener struct{}

func (opener *OsFileOpener) Open(filename string) (*os.File, error) {
	return os.Open(filename)
}

// Store put an object in the given S3 Bucket
func (provider *AwsProvider) Store(filename *string, destination *string) (bool, error) {
	file, err := provider.Opener.Open(*filename)

	if err != nil {
		return false, err
	}

	defer file.Close()

	fileInfo, err := file.Stat()

	if err != nil {
		return false, err
	}

	var size = fileInfo.Size()
	buffer := make([]byte, size)
	_, err = file.Read(buffer)

	if err != nil {
		return false, err
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

	return err == nil, err
}

// Read reads an element from aws
func (provider *AwsProvider) Read(path *string) ([]byte, error) {
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
