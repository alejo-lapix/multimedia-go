package persitence

import (
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func TestNewItem(t *testing.T) {
	type args struct {
		bucket   *string
		filename *string
		fileType *string
	}
	tests := []struct {
		name    string
		args    args
		want    *MultimediaItem
		wantErr bool
	}{
		{
			name: "Must fail if the inputs are empty",
			args: args{
				bucket:   aws.String(""),
				filename: aws.String(""),
				fileType: aws.String(""),
			},
			wantErr: true,
		},
		{
			name: "Must fail if the file type is not valid",
			args: args{
				bucket:   aws.String("http://example.com"),
				filename: aws.String("example"),
				fileType: aws.String("no valid"),
			},
			wantErr: true,
		},
		{
			name: "Must fail if the bucket is not valid",
			args: args{
				bucket:   aws.String("example"),
				filename: aws.String("example"),
				fileType: aws.String(SOUND),
			},
			wantErr: true,
		},
		{
			name: "Returns a MultimediaItem",
			args: args{
				bucket:   aws.String("http://example.com"),
				filename: aws.String("example"),
				fileType: aws.String(SOUND),
			},
			want: &MultimediaItem{
				Bucket:    aws.String("http://example.com"),
				Filename:  aws.String("example"),
				Type:      aws.String(SOUND),
				CreatedAt: aws.String(time.Now().Format(time.RFC3339)),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMultimediaItem(tt.args.bucket, tt.args.filename, tt.args.fileType)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMultimediaItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMultimediaItem() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAWSPersistenceManager_Store(t *testing.T) {
	type fields struct {
		DynamoDB  DynamoDBRepository
		TableName *string
	}
	type args struct {
		item *MultimediaItem
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Must return an error if DynamoDB Fails",
			fields: fields{
				DynamoDB:  &DynamoDBFail{},
				TableName: aws.String("example"),
			},
			args:    args{item: &MultimediaItem{}},
			wantErr: true,
		},
		{
			name: "Must to return a error with nil value",
			fields: fields{
				DynamoDB:  &DynamoDBSuccess{},
				TableName: aws.String("example"),
			},
			args: args{
				item: &MultimediaItem{
					Bucket:   aws.String("http://example.com"),
					Filename: aws.String("example.pdf"),
					Type:     aws.String(PDF),
				},
			},
			wantErr: false,
		},
		{
			name: "Must return an error if DynamoDB Fails",
			fields: fields{
				DynamoDB:  &DynamoDBSuccess{},
				TableName: aws.String("example"),
			},
			args: args{item: &MultimediaItem{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &AWSPersistenceManager{
				DynamoDB:  tt.fields.DynamoDB,
				TableName: tt.fields.TableName,
			}
			if err := manager.Store(tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != (tt.args.item.ID == nil) {
				t.Errorf("Store() if it does not expect an error the ID's value must not be nil")
			}
		})
	}
}

func TestAWSPersistenceManager_Remove(t *testing.T) {
	type fields struct {
		DynamoDB  DynamoDBRepository
		TableName *string
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
			name: "Must return error if DynamoDB Fails",
			fields: fields{
				DynamoDB:  &DynamoDBFail{},
				TableName: aws.String("Error"),
			},
			args:    args{ID: aws.String("any-uuid")},
			wantErr: true,
		},
		{
			name: "Must not return error",
			fields: fields{
				DynamoDB:  &DynamoDBSuccess{},
				TableName: aws.String("example"),
			},
			args:    args{ID: aws.String("any-uuid")},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &AWSPersistenceManager{
				DynamoDB:  tt.fields.DynamoDB,
				TableName: tt.fields.TableName,
			}
			if err := manager.Remove(tt.args.ID); (err != nil) != tt.wantErr {
				t.Errorf("Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type InvalidArguments struct {
	Message *string
}

func (err InvalidArguments) Error() string {
	return *err.Message
}

type DynamoDBSuccess struct{}

func (dynamo *DynamoDBSuccess) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	if input.Item["id"] == nil || input.Item["bucket"] == nil || input.Item["filename"] == nil || input.Item["created_at"] == nil {
		return nil, InvalidArguments{Message: aws.String("Some Attribute were not send to dynamo")}
	}

	return &dynamodb.PutItemOutput{}, nil
}

func (dynamo *DynamoDBSuccess) DeleteItem(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	return &dynamodb.DeleteItemOutput{}, nil
}

type InternalServerError struct{}

func (err InternalServerError) Error() string {
	return "Internal Server InternalServerError"
}

type DynamoDBFail struct{}

func (dynamo *DynamoDBFail) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return nil, InternalServerError{}
}

func (dynamo *DynamoDBFail) DeleteItem(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	return nil, InternalServerError{}
}
