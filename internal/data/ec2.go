package data

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/itmecho/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeEC2Instance is the value used in the AssetType field when fetching EC2 instances
	AssetTypeEC2Instance string = "EC2 Instance"

	// ServiceEC2 is the keyfor the EC2 service
	ServiceEC2 string = "ec2"
)

func (d *Data) loadEC2Instances(ec2Svc ec2iface.EC2API, region string) {
	log := d.log.WithFields(logrus.Fields{
		"region":  region,
		"service": ServiceEC2,
	})
	d.wg.Add(1)
	defer d.wg.Done()
	log.Info("loading instance data")
	out, err := ec2Svc.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err != nil {
		d.results <- result{Err: err}
		return
	}

	log.Info("processing instance data")
	for _, r := range out.Reservations {
		for _, i := range r.Instances {
			var name string
			for _, tag := range i.Tags {
				if *tag.Key == "Name" {
					name = aws.StringValue(tag.Value)
				}
			}

			d.results <- result{
				Row: inventory.Row{
					ID:           aws.StringValue(i.InstanceId),
					AssetType:    AssetTypeEC2Instance,
					Location:     region,
					CreationDate: aws.TimeValue(i.LaunchTime),
					Application:  name,
					Hardware:     aws.StringValue(i.InstanceType),
					Baseline:     aws.StringValue(i.ImageId),
					InternalIP:   aws.StringValue(i.PrivateIpAddress),
					ExternalIP:   aws.StringValue(i.PublicIpAddress),
					VPCID:        aws.StringValue(i.VpcId),
				},
			}
		}
	}

	log.Info("finished processing instance data")
}
