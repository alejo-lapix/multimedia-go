package persitence

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

var runIntegrationTest = os.Getenv("INTEGRATION_TEST")
var awsRegion = os.Getenv("AWS_REGION")
var tableName = "example"

func TestIntegrationAWSPersistenceManager_Store(t *testing.T) {
	if shouldSkip() {
		return
	}

	bucketName := "http://localhost:8080/"
	filename := "example.com"
	fileType := SOUND
	provider, err := NewDynamoDBRepository(&tableName, dynamodb.New(newSession()))

	if err != nil {
		t.Errorf("Store() error = %+v", err)
		return
	}

	item, err := NewMultimediaItem(&bucketName, &filename, &fileType)

	if err != nil {
		t.Errorf("Store() error = %+v", err)
		return
	}

	err = provider.Store(item)

	if err != nil {
		t.Errorf("Store() error = %+v", err.Error())
		return
	}
}

func TestIntegrationAWSPersistenceManager_Remove(t *testing.T) {
	if shouldSkip() {
		return
	}

	uuid := "de86e755-a76f-459b-a3d4-f3feac373617"
	provider, err := NewDynamoDBRepository(&tableName, dynamodb.New(newSession()))

	if err != nil {
		t.Errorf("Remove() error = %+v", err)
		return
	}

	err = provider.Remove(&uuid)

	if err != nil {
		t.Errorf("Remove() error = %+v", err.Error())
	}
}

func newSession() *session.Session {
	sess, err := session.NewSession(&aws.Config{Region: &awsRegion})

	if err != nil {
		panic(err)
	}

	return sess
}

func shouldSkip() bool {
	return runIntegrationTest != "true"
}
