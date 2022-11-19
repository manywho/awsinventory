package awsdata_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/sudoinclabs/awsinventory/internal/awsdata"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
)

var testEBSVolumeRows = []inventory.Row{
	{
		UniqueAssetIdentifier: "vol-12345678",
		Virtual:               true,
		Location:              DefaultRegion,
		AssetType:             AssetTypeEBSVolume,
		HardwareMakeModel:     "gp2 (100GB)",
		Function:              "test app 1",
		SerialAssetTagNumber:  "arn:aws:ec2:us-east-1:012345678910:volume/vol-12345678",
	},
	{
		UniqueAssetIdentifier: "vol-abcdefgh",
		Virtual:               true,
		Location:              DefaultRegion,
		AssetType:             AssetTypeEBSVolume,
		HardwareMakeModel:     "gp2 (50GB)",
		SerialAssetTagNumber:  "arn:aws:ec2:us-east-1:012345678910:volume/vol-abcdefgh",
	},
	{
		UniqueAssetIdentifier: "vol-a1b2c3d4",
		Virtual:               true,
		Location:              DefaultRegion,
		AssetType:             AssetTypeEBSVolume,
		HardwareMakeModel:     "gp2 (20GB)",
		SerialAssetTagNumber:  "arn:aws:ec2:us-east-1:012345678910:volume/vol-a1b2c3d4",
	},
}

// Test Data
var testEBSDescribeVolumesOutputPage1 = &ec2.DescribeVolumesOutput{
	NextToken: aws.String(testEBSVolumeRows[1].UniqueAssetIdentifier),
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
	},
}

var testEBSDescribeVolumesOutputPage2 = &ec2.DescribeVolumesOutput{
	Volumes: []*ec2.Volume{
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

func (e EBSMock) DescribeSecurityGroups(cfg *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
	return testEC2DescribeSecurityGroupsOutput, nil
}

func (e EBSMock) DescribeVolumes(cfg *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
	if cfg.NextToken == nil {
		return testEBSDescribeVolumesOutputPage1, nil
	}

	return testEBSDescribeVolumesOutputPage2, nil
}

type EBSErrorMock struct {
	ec2iface.EC2API
}

func (e EBSErrorMock) DescribeSecurityGroups(cfg *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
	return &ec2.DescribeSecurityGroupsOutput{}, testError
}

func (e EBSErrorMock) DescribeVolumes(cfg *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
	return &ec2.DescribeVolumesOutput{}, testError
}

// Tests
func TestCanLoadEBSVolumes(t *testing.T) {
	d := New(logrus.New(), TestClients{EC2: EBSMock{}})

	var count int
	d.Load([]string{DefaultRegion}, []string{ServiceEBS}, func(row inventory.Row) error {
		require.Equal(t, testEBSVolumeRows[count], row)
		count++
		return nil
	})
	require.Equal(t, 3, count)
}

func TestLoadEBSVolumesLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{EC2: EBSErrorMock{}})

	d.Load([]string{DefaultRegion}, []string{ServiceEBS}, nil)

	assertTestErrorWasLogged(t, hook.Entries)
}
