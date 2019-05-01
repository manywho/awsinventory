package awsdata

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/manywho/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeEBSVolume is the value used in the AssetType field when fetching EBS volumes
	AssetTypeEBSVolume string = "EBS Volume"

	// ServiceEBS is the key for the EBS service
	ServiceEBS string = "ebs"
)

func (d *AWSData) loadEBSVolumes(region string) {
	ec2Svc := d.clients.GetEC2Client(region)

	log := d.log.WithFields(logrus.Fields{
		"region":  region,
		"service": ServiceEBS,
	})
	d.wg.Add(1)
	defer d.wg.Done()
	log.Info("loading data")
	out, err := ec2Svc.DescribeVolumes(&ec2.DescribeVolumesInput{})
	if err != nil {
		d.results <- result{Err: err}
		return
	}

	log.Info("processing data")
	for _, v := range out.Volumes {
		var name string
		for _, t := range v.Tags {
			if aws.StringValue(t.Key) == "Name" {
				name = aws.StringValue(t.Value)
			}
		}

		d.results <- result{
			Row: inventory.Row{
				UniqueAssetIdentifier: aws.StringValue(v.VolumeId),
				Location:              aws.StringValue(v.AvailabilityZone),
				AssetType:             AssetTypeEBSVolume,
				HardwareMakeModel:     fmt.Sprintf("%s (%dGB)", aws.StringValue(v.VolumeType), aws.Int64Value(v.Size)),
				Function:              name,
			},
		}
	}

	log.Info("finished processing data")
}
