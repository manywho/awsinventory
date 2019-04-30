package data_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/manywho/awsinventory/internal/data"
	"github.com/manywho/awsinventory/internal/inventory"
)

var testEC2InstanceRows = []inventory.Row{
	{
		ID:           "i-12345678",
		AssetType:    "EC2 Instance",
		Location:     ValidRegions[0],
		CreationDate: time.Now().AddDate(0, 0, -1),
		Application:  "test app 1",
		Hardware:     "m4.large",
		Baseline:     "ami-12345678",
		InternalIP:   "10.0.1.2",
		ExternalIP:   "203.0.113.10",
		VPCID:        "vpc-12345678",
	},
	{
		ID:           "i-abcdefgh",
		AssetType:    "EC2 Instance",
		Location:     ValidRegions[0],
		CreationDate: time.Now().AddDate(0, 0, -2),
		Application:  "test app 2",
		Hardware:     "t2.medium",
		Baseline:     "ami-abcdefgh",
		InternalIP:   "10.3.4.5",
		ExternalIP:   "203.0.113.20",
		VPCID:        "vpc-abcdefgh",
	},
	{
		ID:           "i-a1b2c3d4",
		AssetType:    "EC2 Instance",
		Location:     ValidRegions[0],
		CreationDate: time.Now().AddDate(0, 0, -3),
		Application:  "test app 3",
		Hardware:     "r4.medium",
		Baseline:     "ami-a1b2c3d4",
		InternalIP:   "10.6.7.8",
		ExternalIP:   "203.0.113.30",
		VPCID:        "vpc-a1b2c3d4",
	},
}

// Test Data
var testEC2InstanceOutput = &ec2.DescribeInstancesOutput{
	Reservations: []*ec2.Reservation{
		{
			Instances: []*ec2.Instance{
				{
					InstanceId:       aws.String(testEC2InstanceRows[0].ID),
					InstanceType:     aws.String(testEC2InstanceRows[0].Hardware),
					ImageId:          aws.String(testEC2InstanceRows[0].Baseline),
					LaunchTime:       aws.Time(testEC2InstanceRows[0].CreationDate),
					PublicIpAddress:  aws.String(testEC2InstanceRows[0].ExternalIP),
					PrivateIpAddress: aws.String(testEC2InstanceRows[0].InternalIP),
					VpcId:            aws.String(testEC2InstanceRows[0].VPCID),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(testEC2InstanceRows[0].Application),
						},
						{
							Key:   aws.String("extra tag"),
							Value: aws.String("testval"),
						},
					},
				},
				{
					InstanceId:       aws.String(testEC2InstanceRows[1].ID),
					InstanceType:     aws.String(testEC2InstanceRows[1].Hardware),
					ImageId:          aws.String(testEC2InstanceRows[1].Baseline),
					LaunchTime:       aws.Time(testEC2InstanceRows[1].CreationDate),
					PublicIpAddress:  aws.String(testEC2InstanceRows[1].ExternalIP),
					PrivateIpAddress: aws.String(testEC2InstanceRows[1].InternalIP),
					VpcId:            aws.String(testEC2InstanceRows[1].VPCID),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(testEC2InstanceRows[1].Application),
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
					InstanceId:       aws.String(testEC2InstanceRows[2].ID),
					InstanceType:     aws.String(testEC2InstanceRows[2].Hardware),
					ImageId:          aws.String(testEC2InstanceRows[2].Baseline),
					LaunchTime:       aws.Time(testEC2InstanceRows[2].CreationDate),
					PublicIpAddress:  aws.String(testEC2InstanceRows[2].ExternalIP),
					PrivateIpAddress: aws.String(testEC2InstanceRows[2].InternalIP),
					VpcId:            aws.String(testEC2InstanceRows[2].VPCID),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(testEC2InstanceRows[2].Application),
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
