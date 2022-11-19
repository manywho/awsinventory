package awsdata

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
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
		log.Errorf("Failed to list buckets: %s", err)
		return
	}

	log.Info("processing data")

	for _, b := range out.Buckets {
		d.wg.Add(1)
		go d.processS3Bucket(log, s3Svc, b, partition, region)
	}

	log.Info("finished processing data")
}

func (d *AWSData) processS3Bucket(log *logrus.Entry, s3Svc s3iface.S3API, bucket *s3.Bucket, partition string, region string) {
	defer d.wg.Done()

	outLocation, err := s3Svc.GetBucketLocation(&s3.GetBucketLocationInput{
		Bucket: bucket.Name,
	})
	if err != nil {
		log.Errorf("failed to get bucket location for %s: %s", aws.StringValue(bucket.Name), err)
		return
	}

	// Only include buckets located in the region selected
	if s3.NormalizeBucketLocation(aws.StringValue(outLocation.LocationConstraint)) != region {
		return
	}

	d.rows <- inventory.Row{
		UniqueAssetIdentifier: aws.StringValue(bucket.Name),
		Virtual:               true,
		Location:              region,
		AssetType:             AssetTypeS3Bucket,
		SerialAssetTagNumber:  fmt.Sprintf("arn:%s:s3:::%s", partition, aws.StringValue(bucket.Name)),
	}
}
