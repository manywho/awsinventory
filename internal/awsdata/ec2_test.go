package awsdata_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/manywho/awsinventory/internal/awsdata"
	"github.com/manywho/awsinventory/internal/inventory"
)

var testEC2InstanceRows = []inventory.Row{
	{
		UniqueAssetIdentifier:     "i-12345678",
		AssetType:                 AssetTypeEC2Instance,
		Virtual:                   true,
		Location:                  ValidRegions[0],
		Function:                  "test app 1",
		HardwareMakeModel:         "m4.large",
		BaselineConfigurationName: "ami-12345678",
		IPv4orIPv6Address:         "203.0.113.10\n10.0.1.2",
		Public:                    true,
		VLANNetworkID:             "vpc-12345678",
	},
	{
		UniqueAssetIdentifier:     "i-abcdefgh",
		AssetType:                 AssetTypeEC2Instance,
		Virtual:                   true,
		Location:                  ValidRegions[0],
		Function:                  "test app 2",
		HardwareMakeModel:         "t2.medium",
		BaselineConfigurationName: "ami-abcdefgh",
		IPv4orIPv6Address:         "10.0.1.3",
		Public:                    false,
		VLANNetworkID:             "vpc-abcdefgh",
	},
	{
		UniqueAssetIdentifier:     "i-a1b2c3d4",
		AssetType:                 AssetTypeEC2Instance,
		Virtual:                   true,
		Location:                  ValidRegions[0],
		Function:                  "test app 3",
		HardwareMakeModel:         "t2.small",
		BaselineConfigurationName: "ami-a1b2c3d4",
		IPv4orIPv6Address:         "10.0.1.4",
		Public:                    false,
		VLANNetworkID:             "vpc-a1b2c3d4",
	},
}

// Test Data
var testEC2InstanceOutput = &ec2.DescribeInstancesOutput{
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
					VpcId: aws.String(testEC2InstanceRows[1].VLANNetworkID),
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

func (e EC2Mock) DescribeInstances(cfg *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return testEC2InstanceOutput, nil
}

type EC2ErrorMock struct {
	ec2iface.EC2API
}

func (e EC2ErrorMock) DescribeInstances(cfg *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return &ec2.DescribeInstancesOutput{}, testError
}

// Tests
func TestCanLoadEC2Instances(t *testing.T) {
	d := New(logrus.New(), TestClients{EC2: EC2Mock{}})

	d.Load([]string{ValidRegions[0]}, []string{"ec2"})

	var count int
	d.MapRows(func(row inventory.Row) error {
		require.Equal(t, testEC2InstanceRows[count], row)
		count++
		return nil
	})
	require.Equal(t, 3, count)
}

func TestLoadEC2InstancesLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{EC2: EC2ErrorMock{}})

	d.Load([]string{ValidRegions[0]}, []string{"ec2"})

	require.Contains(t, hook.LastEntry().Message, testError.Error())
}
