package awsdata_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/sudoinclabs/awsinventory/internal/awsdata"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
)

var testCloudFormationRows = []inventory.Row{
	{
		UniqueAssetIdentifier:          "arn:aws:cloudformation:us-east-1:123456789012:stack/SUDO-Access/ba019810-8fd7-11ec-be38-0a732e992c23",
		Virtual:                        true,
		Public:                         false,
		Location:                       DefaultRegion,
		AssetType:                      "CloudFormation Stack",
		Comments:                       "Test Stack",
		SerialAssetTagNumber:           "arn:aws:rds:us-east-1:123456789012:db:test-db-1",
	},
	{
		UniqueAssetIdentifier:          "arn:aws:cloudformation:us-east-1:123456789012:stack/SUDO-Access/ba019810-8fd7-11ec-be38-0a732e992c23",,
		Virtual:                        true,
		Public:                         false,
		Location:                       DefaultRegion,
		AssetType:                      "CloudFormation Stack",
		Comments:                       "Test Stack",
		SerialAssetTagNumber:           "arn:aws:cloudformation:us-east-1:123456789012:stack/SUDO-Access/ba019810-8fd7-11ec-be38-0a732e992c23",,
	},
}

// Test Data
var testCloudFormationDescribeOutputPage1 = &CloudFormation.DescribeCloudFormationOutput{
	CloudFormation: []*CloudFormation.DescribeStacksInput{
		{
			StackId:        aws.String(testCloudFormationRows[0].SerialAssetTagNumber),
			StackName:               aws.String(testCloudFormationRows[0].Comments),
		},
	},
	NextToken: aws.String(testCloudFormationRows[0].UniqueAssetIdentifier),
}

var testCloudFormationDescribeOutputPage2 = &CloudFormation.DescribeCloudFormationOutput{
	CloudFormation: []*CloudFormation.Workspace{
		{
			StackId:        aws.String(testCloudFormationRows[1].SerialAssetTagNumber),
			StackName:               aws.String(testCloudFormationRows[1].Comments),
		},
	},
}

// Mocks
type CloudFormationMock struct {
	CloudFormationiface.CloudFormationAPI
}

func (e CloudFormationMock) DescribeStacks(cfg *CloudFormation.DescribeStacksInput) (*CloudFormation.DescribeCloudFormationOutput, error) {
	if cfg.NextToken == nil {
		return testCloudFormationDescribeOutputPage1, nil
	}

	return testCloudFormationDescribeOutputPage2, nil
}

type CloudFormationErrorMock struct {
	CloudFormationiface.CloudFormationAPI
}

func (e CloudFormationErrorMock) DescribeStacks(cfg *CloudFormation.DescribeStacksInput) (*CloudFormation.DescribeCloudFormationOutput, error) {
	return &CloudFormation.DescribeCloudFormationOutput{}, testError
}

// Tests
func TestCanLoadCloudFormation(t *testing.T) {
	d := New(logrus.New(), TestClients{WorkSpace: CloudFormationMock{}})

	var count int
	d.Load([]string{DefaultRegion}, []string{ServiceCloudFormation}, func(row inventory.Row) error {
		require.Equal(t, testCloudFormationRows[count], row)
		count++
		return nil
	})
	require.Equal(t, 2, count)
}

func TestLoadCloudFormationLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{WorkSpace: CloudFormationErrorMock{}})

	d.Load([]string{DefaultRegion}, []string{ServiceCloudFormation}, nil)

	assertTestErrorWasLogged(t, hook.Entries)
}
