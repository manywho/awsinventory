package awsdata_test

import (
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/sudoinclabs/awsinventory/internal/awsdata"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
)

var testS3Rows = []inventory.Row{
	{
		UniqueAssetIdentifier: "test-bucket-1",
		Virtual:               true,
		Location:              DefaultRegion,
		AssetType:             AssetTypeS3Bucket,
		SerialAssetTagNumber:  "arn:aws:s3:::test-bucket-1",
	},
	{
		UniqueAssetIdentifier: "test-bucket-2",
		Virtual:               true,
		Location:              DefaultRegion,
		AssetType:             AssetTypeS3Bucket,
		SerialAssetTagNumber:  "arn:aws:s3:::test-bucket-2",
	},
	{
		UniqueAssetIdentifier: "test-bucket-3",
		Virtual:               true,
		Location:              DefaultRegion,
		AssetType:             AssetTypeS3Bucket,
		SerialAssetTagNumber:  "arn:aws:s3:::test-bucket-3",
	},
}

// Test Data
var testS3ListBucketsOutput = &s3.ListBucketsOutput{
	Buckets: []*s3.Bucket{
		{
			Name: aws.String(testS3Rows[0].UniqueAssetIdentifier),
		},
		{
			Name: aws.String(testS3Rows[1].UniqueAssetIdentifier),
		},
		{
			Name: aws.String(testS3Rows[2].UniqueAssetIdentifier),
		},
	},
}

var testS3GetBucketLocationOutput = &s3.GetBucketLocationOutput{
	LocationConstraint: nil,
}

// Mocks
type S3Mock struct {
	s3iface.S3API
}

func (e S3Mock) ListBuckets(cfg *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
	return testS3ListBucketsOutput, nil
}

func (e S3Mock) GetBucketLocation(cfg *s3.GetBucketLocationInput) (*s3.GetBucketLocationOutput, error) {
	return testS3GetBucketLocationOutput, nil
}

type S3ErrorMock struct {
	s3iface.S3API
}

func (e S3ErrorMock) ListBuckets(cfg *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
	return &s3.ListBucketsOutput{}, testError
}

func (e S3ErrorMock) GetBucketLocation(cfg *s3.GetBucketLocationInput) (*s3.GetBucketLocationOutput, error) {
	return &s3.GetBucketLocationOutput{}, testError
}

// Tests
func TestCanLoadS3Buckets(t *testing.T) {
	d := New(logrus.New(), TestClients{S3: S3Mock{}})

	var rows []inventory.Row
	d.Load([]string{DefaultRegion}, []string{ServiceS3}, func(row inventory.Row) error {
		rows = append(rows, row)
		return nil
	})

	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].UniqueAssetIdentifier < rows[j].UniqueAssetIdentifier
	})

	require.Equal(t, 3, len(rows))

	for i, row := range rows {
		require.Equal(t, testS3Rows[i], row)
	}
}

func TestLoadS3BucketsLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{S3: S3ErrorMock{}})

	d.Load([]string{DefaultRegion}, []string{ServiceS3}, nil)

	assertTestErrorWasLogged(t, hook.Entries)
}
