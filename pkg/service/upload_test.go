package service

import (
	"multimedia/pkg/files"
	persistence "multimedia/pkg/persistence"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestAWSUploader_Delete(t *testing.T) {
	type fields struct {
		Bucket     *string
		Region     *string
		Repository persistence.BasicRepository
		Storage    files.Provider
	}
	type args struct {
		ID *string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Should return error if dynamo fails",
			wantErr: true,
			fields: fields{
				Bucket:     aws.String("any-bucket"),
				Region:     aws.String("any-region"),
				Repository: &ServerErrorRepository{},
				Storage:    nil,
			},
		},
		{
			name:    "Should return error if the given files does not exists",
			wantErr: true,
			fields: fields{
				Bucket:     aws.String("any-bucket"),
				Region:     aws.String("any-region"),
				Repository: &EmptyRepository{},
				Storage:    nil,
			},
		},
		{
			name:    "Should return error if can not remove the file",
			wantErr: true,
			fields: fields{
				Bucket:     aws.String("any-bucket"),
				Region:     aws.String("any-region"),
				Repository: &SuccessRepository{},
				Storage:    &FailProvider{},
			},
		},
		{
			name:    "Should not return error if the repository can remove or not",
			wantErr: false,
			fields: fields{
				Bucket:     aws.String("any-bucket"),
				Region:     aws.String("any-region"),
				Repository: &FailRemovingRepository{},
				Storage:    &SuccessProvider{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uploader := &AWSUploader{
				Bucket:     tt.fields.Bucket,
				Region:     tt.fields.Region,
				Repository: tt.fields.Repository,
				Storage:    tt.fields.Storage,
			}
			if err := uploader.Delete(tt.args.ID); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAWSUploader_Upload(t *testing.T) {
	type fields struct {
		Bucket     *string
		Region     *string
		Repository persistence.BasicRepository
		Storage    files.Provider
	}
	type args struct {
		filename    *string
		destination *string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *persistence.MultimediaItem
		wantErr bool
	}{
		/*
			TODO This is a pretty weird case, Struct Validation does not return error even if the required values are empty
			{
				name:    "Should return error if the filename is empty",
				wantErr: true,
				fields: fields{
					Bucket: aws.String("any-bucket"),
					Region: aws.String("any-region"),
				},
				args: args{
					filename:    aws.String(""),
					destination: aws.String(""),
				},
			},
		*/
		{
			name:    "Should return error if the repository can remove or not",
			wantErr: true,
			fields: fields{
				Bucket:  aws.String("any-bucket"),
				Region:  aws.String("any-region"),
				Storage: &FailProvider{},
			},
			args: args{
				filename:    aws.String("filename"),
				destination: aws.String("destination"),
			},
		},
		{
			name:    "Should return a MultimediaItem",
			wantErr: false,
			fields: fields{
				Bucket:  aws.String("any-bucket"),
				Region:  aws.String("any-region"),
				Storage: &SuccessProvider{},
			},
			args: args{
				filename:    aws.String("filename"),
				destination: aws.String("destination"),
			},
			want: &persistence.MultimediaItem{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uploader := &AWSUploader{
				Bucket:     tt.fields.Bucket,
				Region:     tt.fields.Region,
				Repository: tt.fields.Repository,
				Storage:    tt.fields.Storage,
			}
			got, err := uploader.Upload(tt.args.filename, tt.args.destination)
			if (err != nil) != tt.wantErr {
				t.Errorf("Upload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("Upload() got = nil")
			}
		})
	}
}

func TestNewAWSUploader(t *testing.T) {
	type args struct {
		tableName *string
		bucket    *string
		region    *string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Error if parameters are empty",
			args: args{
				tableName: aws.String(""),
				bucket:    aws.String(""),
				region:    aws.String(""),
			},
			wantErr: true,
		},
		{
			name: "Should not return errors",
			args: args{
				tableName: aws.String("example"),
				bucket:    aws.String("example"),
				region:    aws.String("us-west-2"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAWSUploader(tt.args.tableName, tt.args.bucket, tt.args.region)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAWSUploader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("NewAWSUploader() got = nil")
			}
		})
	}
}

type InternalServerError struct{}

func (err InternalServerError) Error() string {
	return "Internal Server Error"
}

type ServerErrorRepository struct{}

func (repository ServerErrorRepository) Store(item *persistence.MultimediaItem) error {
	return InternalServerError{}
}
func (repository ServerErrorRepository) Find(ID *string) (*persistence.MultimediaItem, error) {
	return nil, InternalServerError{}
}
func (repository ServerErrorRepository) Remove(ID *string) error {
	return InternalServerError{}
}

type SuccessRepository struct{}

func (repository SuccessRepository) Store(item *persistence.MultimediaItem) error {
	return nil
}
func (repository SuccessRepository) Find(ID *string) (*persistence.MultimediaItem, error) {
	return &persistence.MultimediaItem{ID: ID}, nil
}
func (repository SuccessRepository) Remove(ID *string) error {
	return nil
}

type EmptyRepository struct{}

func (repository EmptyRepository) Store(item *persistence.MultimediaItem) error {
	return InternalServerError{}
}
func (repository EmptyRepository) Find(ID *string) (*persistence.MultimediaItem, error) {
	return nil, nil
}
func (repository EmptyRepository) Remove(ID *string) error {
	return InternalServerError{}
}

type FailRemovingRepository struct{}

func (repository FailRemovingRepository) Store(item *persistence.MultimediaItem) error {
	return nil
}
func (repository FailRemovingRepository) Find(ID *string) (*persistence.MultimediaItem, error) {
	return &persistence.MultimediaItem{ID: ID}, nil
}
func (repository FailRemovingRepository) Remove(ID *string) error {
	return InternalServerError{}
}

type FailProvider struct{}

func (provider FailProvider) Store(currentPath *string, newPath *string) error {
	return InternalServerError{}
}
func (provider FailProvider) Read(path *string) ([]byte, error) {
	return nil, InternalServerError{}
}
func (provider FailProvider) Remove(filename *string) error {
	return InternalServerError{}
}

type SuccessProvider struct{}

func (provider SuccessProvider) Store(currentPath *string, newPath *string) error {
	return nil
}
func (provider SuccessProvider) Read(path *string) ([]byte, error) {
	return []byte("Example content"), nil
}
func (provider SuccessProvider) Remove(filename *string) error {
	return nil
}
