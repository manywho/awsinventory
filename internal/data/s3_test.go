package data_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/itmecho/awsinventory/internal/data"
	"github.com/itmecho/awsinventory/internal/inventory"
)

var testS3Rows = []inventory.Row{
	{
		ID:           "test-bucket-1",
		AssetType:    AssetTypeS3Bucket,
		Location:     "global",
		CreationDate: time.Now().AddDate(0, 0, -1),
	},
	{
		ID:           "test-bucket-2",
		AssetType:    AssetTypeS3Bucket,
		Location:     "global",
		CreationDate: time.Now().AddDate(0, 0, -2),
	},
	{
		ID:           "test-bucket-3",
		AssetType:    AssetTypeS3Bucket,
		Location:     "global",
		CreationDate: time.Now().AddDate(0, 0, -3),
	},
}

// Test Data
var testS3Output = &s3.ListBucketsOutput{
	Buckets: []*s3.Bucket{
		{
			Name:         aws.String(testS3Rows[0].ID),
			CreationDate: aws.Time(testS3Rows[0].CreationDate),
		},
		{
			Name:         aws.String(testS3Rows[1].ID),
			CreationDate: aws.Time(testS3Rows[1].CreationDate),
		},
		{
			Name:         aws.String(testS3Rows[2].ID),
			CreationDate: aws.Time(testS3Rows[2].CreationDate),
		},
	},
}

// Mocks
type S3Mock struct {
	s3iface.S3API
}

func (e S3Mock) ListBuckets(cfg *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
	return testS3Output, nil
}

type S3ErrorMock struct {
	s3iface.S3API
}

func (e S3ErrorMock) ListBuckets(cfg *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
	return &s3.ListBucketsOutput{}, testError
}

// Tests
func TestCanLoadS3Buckets(t *testing.T) {
	d := New(logrus.New(), TestClients{S3: S3Mock{}})

	d.Load([]string{ValidRegions[0]}, []string{ServiceS3})

	var count int
	d.MapRows(func(row inventory.Row) error {
		require.Equal(t, testS3Rows[count], row)
		count++
		return nil
	})
	require.Equal(t, 3, count)
}

func TestLoadS3BucketsLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{S3: S3ErrorMock{}})

	d.Load([]string{ValidRegions[0]}, []string{ServiceS3})

	require.Contains(t, hook.LastEntry().Message, testError.Error())
}
