package src

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type SuccessCloser struct{}

func (handle *SuccessCloser) Read(p []byte) (int, error) {
	return cap(p), io.EOF
}

func (handle *SuccessCloser) Close() error {
	return nil
}

type ClientError struct {
	Message string
}

func (error ClientError) Error() string {
	return error.Message
}

type SuccessMockS3 struct{}

func (c *SuccessMockS3) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return &s3.PutObjectOutput{}, nil
}

func (c *SuccessMockS3) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return &s3.GetObjectOutput{
		Body:          &SuccessCloser{},
		ContentLength: aws.Int64(5),
	}, nil
}

func (c *SuccessMockS3) DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	return &s3.DeleteObjectOutput{}, nil
}

type FailMockS3 struct{}

func (c FailMockS3) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return &s3.PutObjectOutput{}, ClientError{Message: "Put Object Error"}
}

func (c *FailMockS3) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return nil, ClientError{Message: "Get Object Error"}
}

func (c *FailMockS3) DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	return nil, ClientError{Message: "Delete Object Error"}
}
