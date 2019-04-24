package loader_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/itmecho/awsinventory/internal/inventory"
	. "github.com/itmecho/awsinventory/internal/loader"
)

var testEC2VolumeRows = []inventory.Row{
	{
		ID:           "vol-12345678",
		AssetType:    "EC2 Volume",
		Location:     "test-region-1a",
		CreationDate: time.Now().AddDate(0, 0, -1),
		Application:  "test app 1",
		Hardware:     "gp2",
	},
	{
		ID:           "vol-abcdefgh",
		AssetType:    "EC2 Volume",
		Location:     "test-region-1b",
		CreationDate: time.Now().AddDate(0, 0, -1),
		Hardware:     "gp2",
	},
	{
		ID:           "vol-a1b2c3d4",
		AssetType:    "EC2 Volume",
		Location:     "test-region-1c",
		CreationDate: time.Now().AddDate(0, 0, -1),
		Hardware:     "gp2",
	},
}

func (e EC2Mock) DescribeVolumes(cfg *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
	return &ec2.DescribeVolumesOutput{
		Volumes: []*ec2.Volume{
			{
				VolumeId:         aws.String(testEC2VolumeRows[0].ID),
				VolumeType:       aws.String(testEC2VolumeRows[0].Hardware),
				CreateTime:       aws.Time(testEC2VolumeRows[0].CreationDate),
				AvailabilityZone: aws.String(testEC2VolumeRows[0].Location),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(testEC2VolumeRows[0].Application),
					},
					{
						Key:   aws.String("extra tag"),
						Value: aws.String("testval"),
					},
				},
			},
			{

				VolumeId:         aws.String(testEC2VolumeRows[1].ID),
				VolumeType:       aws.String(testEC2VolumeRows[1].Hardware),
				CreateTime:       aws.Time(testEC2VolumeRows[1].CreationDate),
				AvailabilityZone: aws.String(testEC2VolumeRows[1].Location),
			},
			{

				VolumeId:         aws.String(testEC2VolumeRows[2].ID),
				VolumeType:       aws.String(testEC2VolumeRows[2].Hardware),
				CreateTime:       aws.Time(testEC2VolumeRows[2].CreationDate),
				AvailabilityZone: aws.String(testEC2VolumeRows[2].Location),
			},
		},
	}, nil
}

func TestCanLoadEC2Volumes(t *testing.T) {
	l := NewLoader()

	l.LoadEC2Volumes(EC2Mock{}, "test-region")

	require.Len(t, l.Data, 3, "got more than 3 instances")
	require.Equal(t, testEC2VolumeRows, l.Data, "didn't get expected data")
}

func (e EC2ErrorMock) DescribeVolumes(cfg *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
	return &ec2.DescribeVolumesOutput{}, testError
}

func TestLoadEC2VolumesSendsErrorToChan(t *testing.T) {
	l := NewLoader()

	l.LoadEC2Volumes(EC2ErrorMock{}, "test-region")

	require.Len(t, l.Errors, 1, "didn't send error to Errors channel")

	select {
	case e := <-l.Errors:
		require.Equal(t, testError, e, "didn't get expected error")
	default:
		t.Fatal("should have received an error")
	}
}
