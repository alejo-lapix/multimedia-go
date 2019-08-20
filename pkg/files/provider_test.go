package files

import (
	"multimedia/pkg/files/testdata/src"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestAwsProvider_Read(t *testing.T) {
	type fields struct {
		S3     S3Client
		Bucket *string
	}
	type args struct {
		path *string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Error if file is not found",
			fields: fields{
				S3:     &src.FailMockS3{},
				Bucket: aws.String("Example"),
			},
			args:    args{aws.String("/example/implementation")},
			want:    nil,
			wantErr: true,
		},
		{
			name: "File exists",
			fields: fields{
				S3:     &src.SuccessMockS3{},
				Bucket: aws.String("example"),
			},
			args:    args{aws.String("/example")},
			want:    make([]byte, 5),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &AWSProvider{
				S3:     tt.fields.S3,
				Bucket: tt.fields.Bucket,
			}
			got, err := provider.Read(tt.args.path)

			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Read() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAwsProvider_Store(t *testing.T) {
	type fields struct {
		S3     S3Client
		Bucket *string
	}
	type args struct {
		currentPath *string
		newPath     *string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "File to upload does not exists",
			fields: fields{
				S3:     &src.SuccessMockS3{},
				Bucket: aws.String("Example"),
			},
			args: args{
				currentPath: aws.String("/example/implementation"),
				newPath:     aws.String(""),
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &AWSProvider{
				S3:     tt.fields.S3,
				Bucket: tt.fields.Bucket,
				Opener: &OSFileOpener{},
			}
			got, err := provider.Store(tt.args.currentPath, tt.args.newPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Store() got = %v, want %v", got, tt.want)
			}
		})
	}
}
