package awsdata_test

import (
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/sudoinclabs/awsinventory/internal/awsdata"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
)

var testSQSQueueRows = []inventory.Row{
	{
		UniqueAssetIdentifier: "TestQueue1",
		Virtual:               true,
		DNSNameOrURL:          "https://sqs.us-east-1.amazonaws.com/123456789012/TestQueue1",
		Location:              DefaultRegion,
		AssetType:             AssetTypeSQSQueue,
		Comments:              "100, 0",
		SerialAssetTagNumber:  "arn:aws:sqs:us-east-1:123456789012:TestQueue1",
	},
	{
		UniqueAssetIdentifier: "TestQueue2",
		Virtual:               true,
		DNSNameOrURL:          "https://sqs.us-east-1.amazonaws.com/123456789012/TestQueue2",
		Location:              DefaultRegion,
		AssetType:             AssetTypeSQSQueue,
		Comments:              "0, 0",
		SerialAssetTagNumber:  "arn:aws:sqs:us-east-1:123456789012:TestQueue2",
	},
	{
		UniqueAssetIdentifier: "TestQueue3",
		Virtual:               true,
		DNSNameOrURL:          "https://sqs.us-east-1.amazonaws.com/123456789012/TestQueue3",
		Location:              DefaultRegion,
		AssetType:             AssetTypeSQSQueue,
		Comments:              "200, 100",
		SerialAssetTagNumber:  "arn:aws:sqs:us-east-1:123456789012:TestQueue3",
	},
}

// Test Data
var testSQSListQueuesOutputPage1 = &sqs.ListQueuesOutput{
	NextToken: aws.String(testSQSQueueRows[1].DNSNameOrURL),
	QueueUrls: []*string{
		aws.String(testSQSQueueRows[0].DNSNameOrURL),
		aws.String(testSQSQueueRows[1].DNSNameOrURL),
	},
}

var testSQSListQueuesOutputPage2 = &sqs.ListQueuesOutput{
	QueueUrls: []*string{
		aws.String(testSQSQueueRows[2].DNSNameOrURL),
	},
}

// Mocks
type SQSMock struct {
	sqsiface.SQSAPI
}

func (e SQSMock) ListQueues(cfg *sqs.ListQueuesInput) (*sqs.ListQueuesOutput, error) {
	if cfg.NextToken == nil {
		return testSQSListQueuesOutputPage1, nil
	}

	return testSQSListQueuesOutputPage2, nil
}

func (e SQSMock) GetQueueAttributes(cfg *sqs.GetQueueAttributesInput) (*sqs.GetQueueAttributesOutput, error) {
	var row int
	var numMessages, numMessagesNotVisible string
	switch aws.StringValue(cfg.QueueUrl) {
	case testSQSQueueRows[0].DNSNameOrURL:
		row = 0
		numMessages = "100"
		numMessagesNotVisible = "0"
	case testSQSQueueRows[1].DNSNameOrURL:
		row = 1
		numMessages = "0"
		numMessagesNotVisible = "0"
	case testSQSQueueRows[2].DNSNameOrURL:
		row = 2
		numMessages = "200"
		numMessagesNotVisible = "100"
	}
	return &sqs.GetQueueAttributesOutput{
		Attributes: map[string]*string{
			sqs.QueueAttributeNameApproximateNumberOfMessages:           aws.String(numMessages),
			sqs.QueueAttributeNameApproximateNumberOfMessagesNotVisible: aws.String(numMessagesNotVisible),
			sqs.QueueAttributeNameQueueArn:                              aws.String(testSQSQueueRows[row].SerialAssetTagNumber),
		},
	}, nil
}

type SQSErrorMock struct {
	sqsiface.SQSAPI
}

func (e SQSErrorMock) ListQueues(cfg *sqs.ListQueuesInput) (*sqs.ListQueuesOutput, error) {
	return &sqs.ListQueuesOutput{}, testError
}

func (e SQSErrorMock) GetQueueAttributes(cfg *sqs.GetQueueAttributesInput) (*sqs.GetQueueAttributesOutput, error) {
	return &sqs.GetQueueAttributesOutput{}, testError
}

// Tests
func TestCanLoadSQSQueues(t *testing.T) {
	d := New(logrus.New(), TestClients{SQS: SQSMock{}})

	var rows []inventory.Row
	d.Load([]string{DefaultRegion}, []string{ServiceSQS}, func(row inventory.Row) error {
		rows = append(rows, row)
		return nil
	})

	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].UniqueAssetIdentifier < rows[j].UniqueAssetIdentifier
	})

	require.Equal(t, 3, len(rows))

	for i := range rows {
		require.Equal(t, testSQSQueueRows[i], rows[i])
	}
}

func TestLoadSQSQueuesLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{SQS: SQSErrorMock{}})

	d.Load([]string{DefaultRegion}, []string{ServiceSQS}, nil)

	assertTestErrorWasLogged(t, hook.Entries)
}
