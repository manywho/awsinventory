package loader_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"

	"github.com/itmecho/awsinventory/internal/inventory"
	. "github.com/itmecho/awsinventory/internal/loader"
)

var testS3Rows = []inventory.Row{
	inventory.Row{
		ID:           "test-bucket-1",
		AssetType:    "S3 Bucket",
		Location:     "global",
		CreationDate: time.Now().AddDate(0, 0, -1),
	},
	inventory.Row{
		ID:           "test-bucket-2",
		AssetType:    "S3 Bucket",
		Location:     "global",
		CreationDate: time.Now().AddDate(0, 0, -2),
	},
	inventory.Row{
		ID:           "test-bucket-3",
		AssetType:    "S3 Bucket",
		Location:     "global",
		CreationDate: time.Now().AddDate(0, 0, -3),
	},
}

type S3Mock struct {
	s3iface.S3API
}

func (s S3Mock) ListBuckets(cfg *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
	return &s3.ListBucketsOutput{
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
	}, nil
}

func TestCanLoadS3Buckets(t *testing.T) {
	l := NewLoader()

	l.LoadS3Buckets(S3Mock{})

	require.Len(t, l.Data, 3, "got more than 3 bucket")
	require.Equal(t, testS3Rows, l.Data, "didn't get expected data")
}

type S3ErrorMock struct {
	s3iface.S3API
}

func (s S3ErrorMock) ListBuckets(cfg *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
	return &s3.ListBucketsOutput{}, testError
}

func TestLoadS3BucketsSendsErrorToChan(t *testing.T) {
	l := NewLoader()

	l.LoadS3Buckets(S3ErrorMock{})

	require.Len(t, l.Errors, 1, "didn't send error to Errors channel")

	select {
	case e := <-l.Errors:
		require.Equal(t, testError, e, "didn't get expected error")
	default:
		t.Fatal("should have received an error")
	}
}
