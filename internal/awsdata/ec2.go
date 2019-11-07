package awsdata

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/manywho/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeEC2Instance is the value used in the AssetType field when fetching EC2 instances
	AssetTypeEC2Instance string = "EC2 Instance"

	// ServiceEC2 is the key for the EC2 service
	ServiceEC2 string = "ec2"
)

func (d *AWSData) loadEC2Instances(region string) {
	defer d.wg.Done()

	ec2Svc := d.clients.GetEC2Client(region)

	log := d.log.WithFields(logrus.Fields{
		"region":  region,
		"service": ServiceEC2,
	})
	log.Info("loading instance data")
	out, err := ec2Svc.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("instance-state-name"),
				Values: []*string{
					aws.String("running"),
					aws.String("stopping"),
					aws.String("stopped"),
				},
			},
		},
	})
	if err != nil {
		d.results <- result{Err: err}
		return
	}

	log.Info("processing instance data")
	for _, r := range out.Reservations {
		for _, i := range r.Instances {
			d.wg.Add(1)
			go d.processEC2Instance(i, region)
		}
	}

	log.Info("finished processing instance data")
}

func (d *AWSData) processEC2Instance(i *ec2.Instance, region string) {
	defer d.wg.Done()

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

	ec2Svc := d.clients.GetEC2Client(region)
	var amiName string
	images, err := ec2Svc.DescribeImages(&ec2.DescribeImagesInput{ImageIds: []*string{
		i.ImageId,
	}})
	if err != nil {
		d.results <- result{Err: err}
	} else if len(images.Images) > 0 {
		amiName = aws.StringValue(images.Images[0].Name)
	}

	d.results <- result{
		Row: inventory.Row{
			UniqueAssetIdentifier: aws.StringValue(i.InstanceId),
			IPv4orIPv6Address:     strings.Join(ips, "\n"),
			Virtual:               true,
			// TODO find a better way of checking if the instance is publicly accessible
			Public:                    aws.StringValue(i.PublicIpAddress) != "",
			DNSNameOrURL:              strings.Join(d.route53Cache.FindRecordsForInstance(i), "\n"),
			MACAddress:                strings.Join(macAddresses, "\n"),
			BaselineConfigurationName: aws.StringValue(i.ImageId),
			OSNameAndVersion:          amiName,
			Location:                  region,
			AssetType:                 AssetTypeEC2Instance,
			HardwareMakeModel:         aws.StringValue(i.InstanceType),
			Function:                  name,
			VLANNetworkID:             aws.StringValue(i.VpcId),
		},
	}
}
