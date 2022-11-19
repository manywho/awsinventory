package awsdata_test

import (
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/sudoinclabs/awsinventory/internal/awsdata"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
)

var testEC2InstanceRows = []inventory.Row{
	{
		UniqueAssetIdentifier:     "i-11111111",
		IPv4orIPv6Address:         "203.0.113.10\n10.0.1.2\n10.0.2.2",
		Virtual:                   true,
		Public:                    true,
		DNSNameOrURL:              "test.mydomain.com",
		BaselineConfigurationName: "ami-12345678",
		OSNameAndVersion:          "debian-stretch-2019-01-01",
		Location:                  DefaultRegion,
		AssetType:                 AssetTypeEC2Instance,
		MACAddress:                "00:00:00:00:00:00\n11:11:11:11:11:11",
		HardwareMakeModel:         "m4.large",
		Function:                  "test app 1",
		SerialAssetTagNumber:      "arn:aws:ec2:us-east-1:012345678910:instance/i-11111111",
		VLANNetworkID:             "vpc-12345678",
	},
	{
		UniqueAssetIdentifier:     "i-22222222",
		IPv4orIPv6Address:         "10.0.1.3",
		Virtual:                   true,
		Public:                    true,
		DNSNameOrURL:              "ec2-54-194-252-215.us-east-1.compute.amazonaws.com\nip-192-168-1-88.us-east-1.compute.internal",
		BaselineConfigurationName: "ami-abcdefgh",
		OSNameAndVersion:          "ubuntu-trusty-2019-01-01",
		Location:                  DefaultRegion,
		AssetType:                 AssetTypeEC2Instance,
		HardwareMakeModel:         "t2.medium",
		Function:                  "test app 2",
		SerialAssetTagNumber:      "arn:aws:ec2:us-east-1:012345678910:instance/i-22222222",
		VLANNetworkID:             "vpc-abcdefgh",
	},
	{
		UniqueAssetIdentifier:     "i-33333333",
		IPv4orIPv6Address:         "10.0.1.4",
		Virtual:                   true,
		Public:                    false,
		BaselineConfigurationName: "ami-a1b2c3d4",
		OSNameAndVersion:          "ubuntu-xenial-2019-01-01",
		Location:                  DefaultRegion,
		AssetType:                 AssetTypeEC2Instance,
		HardwareMakeModel:         "t2.small",
		Function:                  "test app 3",
		SerialAssetTagNumber:      "arn:aws:ec2:us-east-1:012345678910:instance/i-33333333",
		VLANNetworkID:             "vpc-a1b2c3d4",
	},
}

// Test Data
var testEC2Route53HostedZonesOutput = &route53.ListHostedZonesOutput{
	HostedZones: []*route53.HostedZone{
		{
			Id: aws.String("ABCDEFGH"),
		},
	},
	IsTruncated: aws.Bool(true),
	NextMarker:  aws.String("ABCDEFGH"),
}

var testEC2Route53RecordSetsOutput = &route53.ListResourceRecordSetsOutput{
	IsTruncated:          aws.Bool(true),
	NextRecordIdentifier: nil,
	NextRecordName:       aws.String(testEC2InstanceRows[0].DNSNameOrURL),
	NextRecordType:       aws.String("A"),
	ResourceRecordSets: []*route53.ResourceRecordSet{
		{
			Type: aws.String("A"),
			Name: aws.String(testEC2InstanceRows[0].DNSNameOrURL),
			ResourceRecords: []*route53.ResourceRecord{
				{
					Value: testEC2DescribeInstancesOutputPage1.Reservations[0].Instances[0].PublicIpAddress,
				},
			},
		},
	},
}

var testEC2DescribeSecurityGroupsOutput = &ec2.DescribeSecurityGroupsOutput{
	SecurityGroups: []*ec2.SecurityGroup{
		{
			OwnerId: aws.String("012345678910"),
		},
	},
}

var testEC2DescribeInstancesOutputPage1 = &ec2.DescribeInstancesOutput{
	NextToken: aws.String(testEC2InstanceRows[1].UniqueAssetIdentifier),
	Reservations: []*ec2.Reservation{
		{
			Instances: []*ec2.Instance{
				{
					InstanceId:      aws.String(testEC2InstanceRows[0].UniqueAssetIdentifier),
					InstanceType:    aws.String(testEC2InstanceRows[0].HardwareMakeModel),
					ImageId:         aws.String(testEC2InstanceRows[0].BaselineConfigurationName),
					PublicIpAddress: aws.String("203.0.113.10"),
					NetworkInterfaces: []*ec2.InstanceNetworkInterface{
						{
							PrivateIpAddress: aws.String("10.0.1.2"),
							PrivateIpAddresses: []*ec2.InstancePrivateIpAddress{
								{
									Primary:          aws.Bool(true),
									PrivateIpAddress: aws.String("10.0.1.2"),
								},
								{
									PrivateIpAddress: aws.String("10.0.1.3"),
								},
							},
							MacAddress: aws.String("00:00:00:00:00:00"),
						},
						{
							PrivateIpAddress: aws.String("10.0.2.2"),
							MacAddress:       aws.String("11:11:11:11:11:11"),
						},
					},
					VpcId: aws.String(testEC2InstanceRows[0].VLANNetworkID),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(testEC2InstanceRows[0].Function),
						},
						{
							Key:   aws.String("extra tag"),
							Value: aws.String("testval"),
						},
					},
				},
				{
					InstanceId:   aws.String(testEC2InstanceRows[1].UniqueAssetIdentifier),
					InstanceType: aws.String(testEC2InstanceRows[1].HardwareMakeModel),
					ImageId:      aws.String(testEC2InstanceRows[1].BaselineConfigurationName),
					NetworkInterfaces: []*ec2.InstanceNetworkInterface{
						{
							PrivateIpAddress: aws.String(testEC2InstanceRows[1].IPv4orIPv6Address),
						},
					},
					PrivateDnsName: aws.String("ip-192-168-1-88.us-east-1.compute.internal"),
					PublicDnsName:  aws.String("ec2-54-194-252-215.us-east-1.compute.amazonaws.com"),
					VpcId:          aws.String(testEC2InstanceRows[1].VLANNetworkID),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(testEC2InstanceRows[1].Function),
						},
						{
							Key:   aws.String("extra tag"),
							Value: aws.String("testval"),
						},
					},
				},
			},
		},
	},
}

var testEC2DescribeInstancesOutputPage2 = &ec2.DescribeInstancesOutput{
	Reservations: []*ec2.Reservation{
		{
			Instances: []*ec2.Instance{
				{
					InstanceId:   aws.String(testEC2InstanceRows[2].UniqueAssetIdentifier),
					InstanceType: aws.String(testEC2InstanceRows[2].HardwareMakeModel),
					ImageId:      aws.String(testEC2InstanceRows[2].BaselineConfigurationName),
					NetworkInterfaces: []*ec2.InstanceNetworkInterface{
						{
							PrivateIpAddress: aws.String(testEC2InstanceRows[2].IPv4orIPv6Address),
						},
					},
					VpcId: aws.String(testEC2InstanceRows[2].VLANNetworkID),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(testEC2InstanceRows[2].Function),
						},
						{
							Key:   aws.String("extra tag"),
							Value: aws.String("testval"),
						},
					},
				},
			},
		},
	},
}

// Mocks
type EC2Mock struct {
	ec2iface.EC2API
}

func (e EC2Mock) DescribeSecurityGroups(cfg *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
	return testEC2DescribeSecurityGroupsOutput, nil
}

func (e EC2Mock) DescribeInstances(cfg *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	if cfg.NextToken == nil {
		return testEC2DescribeInstancesOutputPage1, nil
	}

	return testEC2DescribeInstancesOutputPage2, nil
}

func (e EC2Mock) DescribeImages(cfg *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	var name string
	switch aws.StringValue(cfg.ImageIds[0]) {
	case testEC2InstanceRows[0].BaselineConfigurationName:
		name = testEC2InstanceRows[0].OSNameAndVersion
	case testEC2InstanceRows[1].BaselineConfigurationName:
		name = testEC2InstanceRows[1].OSNameAndVersion
	case testEC2InstanceRows[2].BaselineConfigurationName:
		name = testEC2InstanceRows[2].OSNameAndVersion
	}
	return &ec2.DescribeImagesOutput{
		Images: []*ec2.Image{
			{
				Name: aws.String(name),
			},
		},
	}, nil
}

type EC2Route53Mock struct {
	route53iface.Route53API
}

func (e EC2Route53Mock) ListHostedZones(cfg *route53.ListHostedZonesInput) (*route53.ListHostedZonesOutput, error) {
	if cfg.Marker == testEC2Route53HostedZonesOutput.HostedZones[0].Id {
		return &route53.ListHostedZonesOutput{}, nil
	}

	return testEC2Route53HostedZonesOutput, nil
}

func (e EC2Route53Mock) ListResourceRecordSets(cfg *route53.ListResourceRecordSetsInput) (*route53.ListResourceRecordSetsOutput, error) {
	if cfg.StartRecordName == aws.String(testEC2InstanceRows[0].DNSNameOrURL) {
		return &route53.ListResourceRecordSetsOutput{}, nil
	}

	return testEC2Route53RecordSetsOutput, nil
}

type EC2ErrorMock struct {
	ec2iface.EC2API
}

func (e EC2ErrorMock) DescribeSecurityGroups(cfg *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
	return &ec2.DescribeSecurityGroupsOutput{}, testError
}

func (e EC2ErrorMock) DescribeInstances(cfg *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return &ec2.DescribeInstancesOutput{}, testError
}

func (e EC2ErrorMock) DescribeImages(cfg *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	return &ec2.DescribeImagesOutput{}, testError
}

// Tests
func TestCanLoadEC2Instances(t *testing.T) {
	d := New(logrus.New(), TestClients{EC2: EC2Mock{}, Route53: EC2Route53Mock{}})

	var rows []inventory.Row

	d.Load([]string{DefaultRegion}, []string{ServiceEC2}, func(row inventory.Row) error {
		rows = append(rows, row)
		return nil
	})

	require.Equal(t, 3, len(rows))

	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].UniqueAssetIdentifier < rows[j].UniqueAssetIdentifier
	})

	for i := range rows {
		require.Equal(t, testEC2InstanceRows[i], rows[i])
	}
}

func TestLoadEC2InstancesLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{EC2: EC2ErrorMock{}, Route53: EC2Route53Mock{}})

	d.Load([]string{DefaultRegion}, []string{ServiceEC2}, nil)

	assertTestErrorWasLogged(t, hook.Entries)
}
