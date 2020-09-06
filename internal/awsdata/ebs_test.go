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

var testEBSVolumeRows = []inventory.Row{
	{
		UniqueAssetIdentifier: "vol-12345678",
		Location:              DefaultRegion + "-1a",
		AssetType:             AssetTypeEBSVolume,
		HardwareMakeModel:     "gp2 (100GB)",
		Function:              "test app 1",
	},
	{
		UniqueAssetIdentifier: "vol-abcdefgh",
		Location:              DefaultRegion + "-1b",
		AssetType:             AssetTypeEBSVolume,
		HardwareMakeModel:     "gp2 (50GB)",
	},
	{
		UniqueAssetIdentifier: "vol-a1b2c3d4",
		Location:              DefaultRegion + "-1c",
		AssetType:             AssetTypeEBSVolume,
		HardwareMakeModel:     "gp2 (20GB)",
	},
}

// Test Data
var testEBSVolumesOutput = &ec2.DescribeVolumesOutput{
	Volumes: []*ec2.Volume{
		{
			VolumeId:         aws.String(testEBSVolumeRows[0].UniqueAssetIdentifier),
			VolumeType:       aws.String("gp2"),
			AvailabilityZone: aws.String(testEBSVolumeRows[0].Location),
			Size:             aws.Int64(100),
			Tags: []*ec2.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String(testEBSVolumeRows[0].Function),
				},
				{
					Key:   aws.String("extra tag"),
					Value: aws.String("testval"),
				},
			},
		},
		{

			VolumeId:         aws.String(testEBSVolumeRows[1].UniqueAssetIdentifier),
			VolumeType:       aws.String("gp2"),
			AvailabilityZone: aws.String(testEBSVolumeRows[1].Location),
			Size:             aws.Int64(50),
		},
		{

			VolumeId:         aws.String(testEBSVolumeRows[2].UniqueAssetIdentifier),
			VolumeType:       aws.String("gp2"),
			AvailabilityZone: aws.String(testEBSVolumeRows[2].Location),
			Size:             aws.Int64(20),
		},
	},
}

// Mocks
type EBSMock struct {
	ec2iface.EC2API
}

func (e EBSMock) DescribeVolumes(cfg *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
	return testEBSVolumesOutput, nil
}

type EBSErrorMock struct {
	ec2iface.EC2API
}

func (e EBSErrorMock) DescribeVolumes(cfg *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
	return &ec2.DescribeVolumesOutput{}, testError
}

// Tests
func TestCanLoadEBSVolumes(t *testing.T) {
	d := New(logrus.New(), TestClients{EC2: EBSMock{}})

	d.Load([]string{DefaultRegion}, []string{ServiceEBS})

	var count int
	d.MapRows(func(row inventory.Row) error {
		require.Equal(t, testEBSVolumeRows[count], row)
		count++
		return nil
	})
	require.Equal(t, 3, count)
}

func TestLoadEBSVolumesLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{EC2: EBSErrorMock{}})

	d.Load([]string{DefaultRegion}, []string{ServiceEBS})

	require.Contains(t, hook.LastEntry().Message, testError.Error())
}
