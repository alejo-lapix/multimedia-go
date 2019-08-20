package files

import (
	"io/ioutil"
	"os"
	"reflect"
	"runtime"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var runIntegrationTest = os.Getenv("INTEGRATION_TEST")
var awsRegion = os.Getenv("AWS_REGION")
var awsBucket = os.Getenv("AWS_BUCKET")

var _, filenameToStore, _, _ = runtime.Caller(0)
var uploadedFileName = "provider_integration_test.go"

func TestIntegrationAwsProvider_Store(t *testing.T) {
	if shouldSkip() {
		t.Skip("Integration test skipped")
		return
	}

	provider, err := awsProvider()

	if err != nil {
		t.Errorf("Can not create an AWS Provider error = %v", err)
		return
	}

	t.Logf("Reading File %v", filenameToStore)
	got, err := provider.Store(&filenameToStore, &uploadedFileName)

	if err != nil {
		t.Errorf("Store() error = %+v", err)
		return
	}
	if !reflect.DeepEqual(got, true) {
		t.Errorf("Store() got = %+v, want %t", got, true)
	}
}

func TestIntegrationAwsProvider_Read(t *testing.T) {
	if shouldSkip() {
		t.Skip("Integration test skipped")
		return
	}

	provider, err := awsProvider()

	if err != nil {
		t.Errorf("Can not create an AWS Provider error = %v", err)
		return
	}

	got, err := provider.Read(&uploadedFileName)

	if err != nil {
		t.Errorf("Read() error = %+v", err)
		return
	}

	buffer, err := ioutil.ReadFile(filenameToStore)

	if err != nil {
		t.Errorf("Can not open the file to compare due to = %+v", err)
		return
	}

	if !reflect.DeepEqual(string(got), string(buffer)) {
		t.Errorf("Read() got = %v, want %v", string(got), string(buffer))
	}
}

func awsProvider() (*AWSProvider, error) {
	sess, err := newSession()

	if err != nil {
		return nil, err
	}

	return NewAWSProvider(&awsBucket, s3.New(sess)), nil
}

func newSession() (*session.Session, error) {
	return session.NewSession(&aws.Config{Region: &awsRegion})
}

func shouldSkip() bool {
	return runIntegrationTest != "true"
}
