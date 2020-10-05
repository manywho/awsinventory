package awsdata

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/manywho/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeS3Bucket is the value used in the AssetType field when fetching S3 Buckets
	AssetTypeS3Bucket string = "S3 Bucket"

	// ServiceS3 is the key for the S3 service
	ServiceS3 string = "s3"
)

func (d *AWSData) loadS3Buckets(region string) {
	defer d.wg.Done()

	s3Svc := d.clients.GetS3Client(region)

	log := d.log.WithFields(logrus.Fields{
		"region":  region,
		"service": ServiceS3,
	})

	log.Info("loading data")

	var partition string
	if p, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), region); ok {
		partition = p.ID()
	}

	out, err := s3Svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		d.results <- result{Err: err}
		return
	}

	log.Info("processing data")

	for _, b := range out.Buckets {
		outLocation, err := s3Svc.GetBucketLocation(&s3.GetBucketLocationInput{
			Bucket: b.Name,
		})
		if err != nil {
			d.results <- result{Err: err}
			return
		}

		// Only include buckets located in the region selected
		if s3.NormalizeBucketLocation(aws.StringValue(outLocation.LocationConstraint)) != region {
			continue
		}

		d.results <- result{
			Row: inventory.Row{
				UniqueAssetIdentifier: aws.StringValue(b.Name),
				Virtual:               true,
				Location:              region,
				AssetType:             AssetTypeS3Bucket,
				SerialAssetTagNumber:  fmt.Sprintf("arn:%s:s3:::%s", partition, aws.StringValue(b.Name)),
			},
		}
	}

	log.Info("finished processing data")
}
