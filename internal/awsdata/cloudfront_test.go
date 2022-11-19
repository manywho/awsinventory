package awsdata_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/cloudfront/cloudfrontiface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/sudoinclabs/awsinventory/internal/awsdata"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
)

var testCloudFrontDistributionRows = []inventory.Row{
	{
		UniqueAssetIdentifier:     "EDFDVBD632BHDS5",
		Virtual:                   true,
		Public:                    true,
		DNSNameOrURL:              "d111111abcdef8.cloudfront.net\ntest-1.example.com",
		BaselineConfigurationName: "awsexamplebucket.s3.us-west-2.amazonaws.com",
		AssetType:                 "CloudFront Distribution",
		Function:                  "Test distribution 1",
		SerialAssetTagNumber:      "arn:aws:cloudfront::123456789012:distribution/EDFDVBD632BHDS5",
	},
	{
		UniqueAssetIdentifier:     "EMLARXS9EXAMPLE",
		Virtual:                   true,
		Public:                    true,
		DNSNameOrURL:              "d222222abcdef7.cloudfront.net\ntest-2.example.com",
		BaselineConfigurationName: "example-load-balancer-1234567890.us-west-2.elb.amazonaws.com",
		AssetType:                 "CloudFront Distribution",
		Function:                  "Test distribution 2",
		SerialAssetTagNumber:      "arn:aws:cloudfront::123456789012:distribution/EMLARXS9EXAMPLE",
	},
	{
		UniqueAssetIdentifier:     "E7GGTQ8UCFC4G",
		Virtual:                   true,
		Public:                    true,
		DNSNameOrURL:              "d333333abcdef6.cloudfront.net",
		BaselineConfigurationName: "https://example3a.com\nhttps://example3b.com",
		AssetType:                 "CloudFront Distribution",
		Function:                  "Test distribution 3",
		SerialAssetTagNumber:      "arn:aws:cloudfront::123456789012:distribution/E7GGTQ8UCFC4G",
	},
}

// Test Data
var testCloudFrontListDistributionsOutputPage1 = &cloudfront.ListDistributionsOutput{
	DistributionList: &cloudfront.DistributionList{
		IsTruncated: aws.Bool(true),
		Items: []*cloudfront.DistributionSummary{
			{
				Aliases: &cloudfront.Aliases{
					Items: []*string{
						aws.String("test-1.example.com"),
					},
				},
				ARN:        aws.String(testCloudFrontDistributionRows[0].SerialAssetTagNumber),
				Comment:    aws.String(testCloudFrontDistributionRows[0].Function),
				DomainName: aws.String("d111111abcdef8.cloudfront.net"),
				Id:         aws.String(testCloudFrontDistributionRows[0].UniqueAssetIdentifier),
				Origins: &cloudfront.Origins{
					Items: []*cloudfront.Origin{
						{
							DomainName: aws.String("awsexamplebucket.s3.us-west-2.amazonaws.com"),
						},
					},
				},
			},
			{
				Aliases: &cloudfront.Aliases{
					Items: []*string{
						aws.String("test-2.example.com"),
					},
				},
				ARN:        aws.String(testCloudFrontDistributionRows[1].SerialAssetTagNumber),
				Comment:    aws.String(testCloudFrontDistributionRows[1].Function),
				DomainName: aws.String("d222222abcdef7.cloudfront.net"),
				Id:         aws.String(testCloudFrontDistributionRows[1].UniqueAssetIdentifier),
				Origins: &cloudfront.Origins{
					Items: []*cloudfront.Origin{
						{
							DomainName: aws.String("example-load-balancer-1234567890.us-west-2.elb.amazonaws.com"),
						},
					},
				},
			},
		},
		NextMarker: aws.String(testCloudFrontDistributionRows[1].UniqueAssetIdentifier),
	},
}

var testCloudFrontListDistributionsOutputPage2 = &cloudfront.ListDistributionsOutput{
	DistributionList: &cloudfront.DistributionList{
		Items: []*cloudfront.DistributionSummary{
			{
				Aliases: &cloudfront.Aliases{
					Items: []*string{},
				},
				ARN:        aws.String(testCloudFrontDistributionRows[2].SerialAssetTagNumber),
				Comment:    aws.String(testCloudFrontDistributionRows[2].Function),
				DomainName: aws.String("d333333abcdef6.cloudfront.net"),
				Id:         aws.String(testCloudFrontDistributionRows[2].UniqueAssetIdentifier),
				Origins: &cloudfront.Origins{
					Items: []*cloudfront.Origin{
						{
							DomainName: aws.String("https://example3a.com"),
						},
						{
							DomainName: aws.String("https://example3b.com"),
						},
					},
				},
			},
		},
		Marker: aws.String(testCloudFrontDistributionRows[1].UniqueAssetIdentifier),
	},
}

// Mocks
type CloudFrontMock struct {
	cloudfrontiface.CloudFrontAPI
}

func (e CloudFrontMock) ListDistributions(cfg *cloudfront.ListDistributionsInput) (*cloudfront.ListDistributionsOutput, error) {
	if cfg.Marker == nil {
		return testCloudFrontListDistributionsOutputPage1, nil
	}

	return testCloudFrontListDistributionsOutputPage2, nil
}

type CloudFrontErrorMock struct {
	cloudfrontiface.CloudFrontAPI
}

func (e CloudFrontErrorMock) ListDistributions(cfg *cloudfront.ListDistributionsInput) (*cloudfront.ListDistributionsOutput, error) {
	return &cloudfront.ListDistributionsOutput{}, testError
}

// Tests
func TestCanLoadCloudFrontDistributions(t *testing.T) {
	d := New(logrus.New(), TestClients{CloudFront: CloudFrontMock{}})

	var count int
	d.Load([]string{}, []string{ServiceCloudFront}, func(row inventory.Row) error {
		require.Equal(t, testCloudFrontDistributionRows[count], row)
		count++
		return nil
	})
	require.Equal(t, 3, count)
}

func TestLoadCloudFrontDistributionsLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{CloudFront: CloudFrontErrorMock{}})

	d.Load([]string{}, []string{ServiceCloudFront}, nil)

	assertTestErrorWasLogged(t, hook.Entries)
}
