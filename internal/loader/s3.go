package loader

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/itmecho/awsinventory/internal/inventory"
)

// LoadS3Buckets loads the s3 bucket data from the given region into the Loader's data
func (l *Loader) LoadS3Buckets(s3Svc s3iface.S3API) {
	out, err := s3Svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		l.Errors <- err
		return
	}

	results := make([]inventory.Row, 0)

	for _, b := range out.Buckets {
		results = append(results, inventory.Row{
			ID:           aws.StringValue(b.Name),
			AssetType:    "S3 Bucket",
			Location:     "global",
			CreationDate: aws.TimeValue(b.CreationDate),
		})
	}

	l.appendData(results)
}
