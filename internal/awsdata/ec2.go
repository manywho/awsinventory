package awsdata

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
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

	log.Info("loading data")

	var partition string
	if p, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), region); ok {
		partition = p.ID()
	}

	var accountID string
	out, err := ec2Svc.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		MaxResults: aws.Int64(5),
	})
	if err != nil {
		log.Errorf("failed to load account id from security groups: %s", err)
		return
	} else if len(out.SecurityGroups) > 0 {
		accountID = aws.StringValue(out.SecurityGroups[0].OwnerId)
	}

	var reservations []*ec2.Reservation
	done := false
	params := &ec2.DescribeInstancesInput{
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
	}
	for !done {
		out, err := ec2Svc.DescribeInstances(params)
		if err != nil {
			log.Errorf("failed to describe instances: %s", err)
			return
		}

		reservations = append(reservations, out.Reservations...)

		if out.NextToken == nil {
			done = true
		} else {
			params.NextToken = out.NextToken
		}
	}

	log.Info("processing data")

	for _, r := range reservations {
		for _, i := range r.Instances {
			d.wg.Add(1)
			go d.processEC2Instance(log, ec2Svc, i, accountID, region, partition)
		}
	}

	log.Info("finished processing data")
}

func (d *AWSData) processEC2Instance(log *logrus.Entry, ec2Svc ec2iface.EC2API, instance *ec2.Instance, accountID string, region string, partition string) {
	defer d.wg.Done()

	var name string
	for _, tag := range instance.Tags {
		if *tag.Key == "Name" {
			name = aws.StringValue(tag.Value)
		}
	}

	var public = false
	var ips []string
	var macAddresses []string
	var dnsNames []string

	if aws.StringValue(instance.PublicIpAddress) != "" {
		ips = append(ips, aws.StringValue(instance.PublicIpAddress))
		public = true
	}

	for _, networkInterface := range instance.NetworkInterfaces {
		ips = append(ips, aws.StringValue(networkInterface.PrivateIpAddress))
		for _, ipSet := range networkInterface.PrivateIpAddresses {
			if aws.BoolValue(ipSet.Primary) {
				ips = appendIfMissing(ips, aws.StringValue(ipSet.PrivateIpAddress))
			}
		}
		macAddresses = append(macAddresses, aws.StringValue(networkInterface.MacAddress))
	}

	dnsNames = append(dnsNames, d.route53Cache.FindRecordsForInstance(instance)...)

	if aws.StringValue(instance.PublicDnsName) != "" {
		dnsNames = appendIfMissing(dnsNames, aws.StringValue(instance.PublicDnsName))
		public = true
	}

	if aws.StringValue(instance.PrivateDnsName) != "" {
		dnsNames = appendIfMissing(dnsNames, aws.StringValue(instance.PrivateDnsName))
	}

	var amiName string
	images, err := ec2Svc.DescribeImages(&ec2.DescribeImagesInput{ImageIds: []*string{
		instance.ImageId,
	}})
	if err != nil {
		log.Warningf("failed to load ami for %s: %s", aws.StringValue(instance.InstanceId), err)
	} else if len(images.Images) > 0 {
		amiName = aws.StringValue(images.Images[0].Name)
	}

	d.rows <- inventory.Row{
		UniqueAssetIdentifier:     aws.StringValue(instance.InstanceId),
		IPv4orIPv6Address:         strings.Join(ips, "\n"),
		Virtual:                   true,
		Public:                    public,
		DNSNameOrURL:              strings.Join(dnsNames, "\n"),
		MACAddress:                strings.Join(macAddresses, "\n"),
		BaselineConfigurationName: aws.StringValue(instance.ImageId),
		OSNameAndVersion:          amiName,
		Location:                  region,
		AssetType:                 AssetTypeEC2Instance,
		HardwareMakeModel:         aws.StringValue(instance.InstanceType),
		Function:                  name,
		SerialAssetTagNumber:      fmt.Sprintf("arn:%s:ec2:%s:%s:instance/%s", partition, region, accountID, aws.StringValue(instance.InstanceId)),
		VLANNetworkID:             aws.StringValue(instance.VpcId),
	}
}
