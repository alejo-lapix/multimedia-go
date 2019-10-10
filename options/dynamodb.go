package options

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type DynamoDBPageOptionRepository struct {
	DynamoDB  *dynamodb.DynamoDB
	tableName string
}

func (repository *DynamoDBPageOptionRepository) Store(option *PageOption) error {
	item, err := dynamodbattribute.MarshalMap(option)

	if err != nil {
		return err
	}

	_, err = repository.DynamoDB.PutItem(&dynamodb.PutItemInput{
		Item:      item,
		TableName: &repository.tableName,
	})

	return err
}

func (repository *DynamoDBPageOptionRepository) FindByName(name string) (*PageOption, error) {
	item := &PageOption{}
	output, err := repository.DynamoDB.GetItem(&dynamodb.GetItemInput{
		Key:       map[string]*dynamodb.AttributeValue{"name": {S: &name}},
		TableName: &repository.tableName,
	})

	if err != nil {
		return nil, err
	}

	if err = dynamodbattribute.UnmarshalMap(output.Item, item); err != nil {
		return nil, err
	}

	return item, nil
}
