package persistence

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"

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
	ID        *string
	Bucket    *string `validate:"required,url"`
	Filename  *string `validate:"required"`
	Type      *string `validate:"required,oneof=sound image pdf"`
	CreatedAt *string
}

func NewMultimediaItem(bucket, filename, fileType *string) (*MultimediaItem, error) {
	multimediaItem := &MultimediaItem{
		Bucket:   bucket,
		Filename: filename,
		Type:     fileType,
	}

	err := validator.New().Struct(multimediaItem)

	if err != nil {
		return nil, err
	}

	currentDate := time.Now().Format(time.RFC3339)
	multimediaItem.CreatedAt = &currentDate

	return multimediaItem, nil
}

type Storable interface {
	Store(item *MultimediaItem) error
}

type Removable interface {
	Remove(ID *string) error
}

type Findable interface {
	Find(ID *string) (*MultimediaItem, error)
	FindMany([]*string) ([]*MultimediaItem, error)
}

type BasicRepository interface {
	Storable
	Removable
	Findable
}

type DynamoDBRepository interface {
	PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error)
	DeleteItem(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error)
	GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
	Query(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error)
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
			"id":        &dynamodb.AttributeValue{S: &ID},
			"bucket":    &dynamodb.AttributeValue{S: item.Bucket},
			"filename":  &dynamodb.AttributeValue{S: item.Filename},
			"type":      &dynamodb.AttributeValue{S: item.Type},
			"createdAt": &dynamodb.AttributeValue{S: item.CreatedAt},
		},
	}

	_, err := manager.DynamoDB.PutItem(&input)

	if err == nil {
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

func (manager *AWSPersistenceManager) Find(ID *string) (*MultimediaItem, error) {
	output, err := manager.DynamoDB.GetItem(&dynamodb.GetItemInput{
		Key:       map[string]*dynamodb.AttributeValue{"id": &dynamodb.AttributeValue{S: ID}},
		TableName: manager.TableName,
	})

	if err != nil {
		return nil, err
	}

	if output.Item == nil {
		return nil, nil
	}

	return mapItemOutput(output.Item), nil
}

func mapItemOutput(output map[string]*dynamodb.AttributeValue) *MultimediaItem {
	return &MultimediaItem{
		ID:        output["id"].S,
		Bucket:    output["bucket"].S,
		Filename:  output["filename"].S,
		Type:      output["type"].S,
		CreatedAt: output["createdAt"].S,
	}
}

func (manager *AWSPersistenceManager) FindMany(ids []*string) ([]*MultimediaItem, error) {
	attributeValues := make(map[string]*dynamodb.AttributeValue, len(ids))
	conditionExpression := make([]string, len(ids))

	for index, id := range ids {
		queryValue := fmt.Sprintf(":v%v", index+1)

		attributeValues[queryValue] = &dynamodb.AttributeValue{S: id}
		conditionExpression[index] = queryValue
	}

	output, err := manager.DynamoDB.Query(&dynamodb.QueryInput{
		TableName:                 manager.TableName,
		FilterExpression:          aws.String(fmt.Sprintf("id IN (%v)", strings.Join(conditionExpression, ","))),
		ExpressionAttributeValues: attributeValues,
	})

	if err != nil {
		return nil, err
	}

	result := make([]*MultimediaItem, len(output.Items))

	for index, item := range output.Items {
		result[index] = mapItemOutput(item)
	}

	return result, nil
}
