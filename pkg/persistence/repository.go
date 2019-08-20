package persitence

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
	"gopkg.in/go-playground/validator.v9"
)

const (
	SOUND = "sound"
	IMAGE = "image"
	PDF   = "pdf"
	VIDEO = "video"
)

type MultimediaItem struct {
	ID       *string
	Bucket   *string `validate:"required,url"`
	Filename *string `validate:"required"`
	Type     *string `validate:"required,oneof=sound image pdf"`
}

func NewMultimediaItem(bucket *string, filename *string, fileType *string) (*MultimediaItem, error) {
	multimediaItem := &MultimediaItem{
		Bucket:   bucket,
		Filename: filename,
		Type:     fileType,
	}

	err := validator.New().Struct(multimediaItem)

	if err != nil {
		return nil, err
	}

	return multimediaItem, nil
}

type Storable interface {
	Store(item *MultimediaItem) error
}

type Removable interface {
	Remove(ID *string) error
}

type DynamoDBRepository interface {
	PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error)
	DeleteItem(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error)
}

type AWSPersistenceManager struct {
	DynamoDB  DynamoDBRepository
	TableName *string `validate:"required"`
}

func NewDynamoDBRepository(tableName *string, repository DynamoDBRepository) (*AWSPersistenceManager, error) {
	manager := AWSPersistenceManager{
		DynamoDB:  repository,
		TableName: tableName,
	}

	err := validator.New().Struct(manager)

	if err != nil {
		return nil, err
	}

	return &manager, nil
}

func (manager *AWSPersistenceManager) Store(item *MultimediaItem) error {
	ID := uuid.New().String()

	input := dynamodb.PutItemInput{
		TableName: manager.TableName,
		Item: map[string]*dynamodb.AttributeValue{
			"id":       &dynamodb.AttributeValue{S: &ID},
			"bucket":   &dynamodb.AttributeValue{S: item.Bucket},
			"filename": &dynamodb.AttributeValue{S: item.Filename},
		},
	}

	_, err := manager.DynamoDB.PutItem(&input)

	if err != nil {
		// TODO make a copy of the real object and returns it
		item.ID = &ID
	}

	return err
}

func (manager *AWSPersistenceManager) Remove(ID *string) error {
	_, err := manager.DynamoDB.DeleteItem(&dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": &dynamodb.AttributeValue{S: ID},
		},
		TableName: manager.TableName,
	})

	return err
}
