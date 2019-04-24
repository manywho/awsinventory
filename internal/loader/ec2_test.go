package loader_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"

	"github.com/itmecho/awsinventory/internal/inventory"
	. "github.com/itmecho/awsinventory/internal/loader"
)

var testEC2Rows = []inventory.Row{
	{
		ID:           "i-12345678",
		AssetType:    "EC2 Instance",
		Location:     "test-region",
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
		Location:     "test-region",
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
		Location:     "test-region",
		CreationDate: time.Now().AddDate(0, 0, -3),
		Application:  "test app 3",
		Hardware:     "r4.medium",
		Baseline:     "ami-a1b2c3d4",
		InternalIP:   "10.6.7.8",
		ExternalIP:   "203.0.113.30",
		VPCID:        "vpc-a1b2c3d4",
	},
}

type EC2Mock struct {
	ec2iface.EC2API
}

func (e EC2Mock) DescribeInstances(cfg *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return &ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			{
				Instances: []*ec2.Instance{
					{
						InstanceId:       aws.String(testEC2Rows[0].ID),
						InstanceType:     aws.String(testEC2Rows[0].Hardware),
						ImageId:          aws.String(testEC2Rows[0].Baseline),
						LaunchTime:       aws.Time(testEC2Rows[0].CreationDate),
						PublicIpAddress:  aws.String(testEC2Rows[0].ExternalIP),
						PrivateIpAddress: aws.String(testEC2Rows[0].InternalIP),
						VpcId:            aws.String(testEC2Rows[0].VPCID),
						Tags: []*ec2.Tag{
							{
								Key:   aws.String("Name"),
								Value: aws.String(testEC2Rows[0].Application),
							},
							{
								Key:   aws.String("extra tag"),
								Value: aws.String("testval"),
							},
						},
					},
					{
						InstanceId:       aws.String(testEC2Rows[1].ID),
						InstanceType:     aws.String(testEC2Rows[1].Hardware),
						ImageId:          aws.String(testEC2Rows[1].Baseline),
						LaunchTime:       aws.Time(testEC2Rows[1].CreationDate),
						PublicIpAddress:  aws.String(testEC2Rows[1].ExternalIP),
						PrivateIpAddress: aws.String(testEC2Rows[1].InternalIP),
						VpcId:            aws.String(testEC2Rows[1].VPCID),
						Tags: []*ec2.Tag{
							{
								Key:   aws.String("Name"),
								Value: aws.String(testEC2Rows[1].Application),
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
						InstanceId:       aws.String(testEC2Rows[2].ID),
						InstanceType:     aws.String(testEC2Rows[2].Hardware),
						ImageId:          aws.String(testEC2Rows[2].Baseline),
						LaunchTime:       aws.Time(testEC2Rows[2].CreationDate),
						PublicIpAddress:  aws.String(testEC2Rows[2].ExternalIP),
						PrivateIpAddress: aws.String(testEC2Rows[2].InternalIP),
						VpcId:            aws.String(testEC2Rows[2].VPCID),
						Tags: []*ec2.Tag{
							{
								Key:   aws.String("Name"),
								Value: aws.String(testEC2Rows[2].Application),
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
	}, nil
}

func TestCanLoadEC2Instances(t *testing.T) {
	l := NewLoader()

	l.LoadEC2Instances(EC2Mock{}, testEC2Rows[0].Location)

	require.Len(t, l.Data, 3, "got more than 3 instances")
	require.Equal(t, testEC2Rows, l.Data, "didn't get expected data")
}

var testError = errors.New("test aws error")

type EC2ErrorMock struct {
	ec2iface.EC2API
}

func (e EC2ErrorMock) DescribeInstances(cfg *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return &ec2.DescribeInstancesOutput{}, testError
}

func TestLoadEC2InstancesSendsErrorToChan(t *testing.T) {
	l := NewLoader()

	l.LoadEC2Instances(EC2ErrorMock{}, testEC2Rows[0].Location)

	require.Len(t, l.Errors, 1, "didn't send error to Errors channel")

	select {
	case e := <-l.Errors:
		require.Equal(t, testError, e, "didn't get expected error")
	default:
		t.Fatal("should have recieved an error")
	}
}
