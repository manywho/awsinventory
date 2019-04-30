package awsdata

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/manywho/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeEBSVolume is the value used in the AssetType field when fetching EBS volumes
	AssetTypeEBSVolume string = "EBS Volume"

	// ServiceEBS is the key for the EBS service
	ServiceEBS string = "ebs"
)

func (d *Data) loadEBSVolumes(ec2Svc ec2iface.EC2API, region string) {
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
				ID:           aws.StringValue(v.VolumeId),
				AssetType:    AssetTypeEBSVolume,
				Location:     aws.StringValue(v.AvailabilityZone),
				CreationDate: aws.TimeValue(v.CreateTime),
				Application:  name,
				Hardware:     fmt.Sprintf("%s (%dGB)", aws.StringValue(v.VolumeType), aws.Int64Value(v.Size)),
			},
		}
	}

	log.Info("finished processing data")
}
