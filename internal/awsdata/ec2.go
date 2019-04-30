package awsdata

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/manywho/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeEC2Instance is the value used in the AssetType field when fetching EC2 instances
	AssetTypeEC2Instance string = "EC2 Instance"

	// ServiceEC2 is the key for the EC2 service
	ServiceEC2 string = "ec2"
)

func (d *AWSData) loadEC2Instances(ec2Svc ec2iface.EC2API, region string) {
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

			var ips []string
			var macAddresses []string

			if aws.StringValue(i.PublicIpAddress) != "" {
				ips = append(ips, aws.StringValue(i.PublicIpAddress))
			}

			for _, networkInterface := range i.NetworkInterfaces {
				ips = append(ips, aws.StringValue(networkInterface.PrivateIpAddress))
				macAddresses = append(macAddresses, aws.StringValue(networkInterface.MacAddress))
			}

			d.results <- result{
				Row: inventory.Row{
					UniqueAssetIdentifier: aws.StringValue(i.InstanceId),
					IPv4orIPv6Address:     strings.Join(ips, "\n"),
					Virtual:               true,
					Public:                aws.StringValue(i.PublicIpAddress) != "",
					// TODO DNSNameOrURL
					MACAddress:                strings.Join(macAddresses, "\n"),
					BaselineConfigurationName: aws.StringValue(i.ImageId),
					Location:                  region,
					AssetType:                 AssetTypeEC2Instance,
					HardwareMakeModel:         aws.StringValue(i.InstanceType),
					Function:                  name,
					VLANNetworkID:             aws.StringValue(i.VpcId),
				},
			}
		}
	}

	log.Info("finished processing instance data")
}
