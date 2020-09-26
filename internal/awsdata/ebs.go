package awsdata

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
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
	defer d.wg.Done()

	ec2Svc := d.clients.GetEC2Client(region)

	log := d.log.WithFields(logrus.Fields{
		"region":  region,
		"service": ServiceEBS,
	})

	log.Info("loading data")

	var partition string
	if p, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), region); ok {
		partition = p.ID()
	}

	var accountId string
	out, err := ec2Svc.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		MaxResults: aws.Int64(5),
	})
	if err != nil {
		d.results <- result{Err: err}
		return
	} else if len(out.SecurityGroups) > 0 {
		accountId = aws.StringValue(out.SecurityGroups[0].OwnerId)
	}

	var volumes []*ec2.Volume
	done := false
	params := &ec2.DescribeVolumesInput{}
	for !done {
		out, err := ec2Svc.DescribeVolumes(params)
		if err != nil {
			d.results <- result{Err: err}
			return
		}

		volumes = append(volumes, out.Volumes...)

		if out.NextToken == nil {
			done = true
		} else {
			params.NextToken = out.NextToken
		}
	}

	log.Info("processing data")

	for _, v := range volumes {
		var name string
		for _, t := range v.Tags {
			if aws.StringValue(t.Key) == "Name" {
				name = aws.StringValue(t.Value)
			}
		}

		d.results <- result{
			Row: inventory.Row{
				UniqueAssetIdentifier: aws.StringValue(v.VolumeId),
				Virtual:               true,
				Location:              region,
				AssetType:             AssetTypeEBSVolume,
				HardwareMakeModel:     fmt.Sprintf("%s (%dGB)", aws.StringValue(v.VolumeType), aws.Int64Value(v.Size)),
				Function:              name,
				SerialAssetTagNumber:  fmt.Sprintf("arn:%s:ec2:%s:%s:volume/%s", partition, region, accountId, aws.StringValue(v.VolumeId)),
			},
		}
	}

	log.Info("finished processing data")
}
