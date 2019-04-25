package data_test

import (
	"bufio"
	"bytes"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	. "github.com/itmecho/awsinventory/internal/data"
	"github.com/itmecho/awsinventory/internal/inventory"
)

var testEBSVolumeRows = []inventory.Row{
	{
		ID:           "vol-12345678",
		AssetType:    AssetTypeEBSVolume,
		Location:     "test-region-1a",
		CreationDate: time.Now().AddDate(0, 0, -1),
		Application:  "test app 1",
		Hardware:     "gp2 (100GB)",
	},
	{
		ID:           "vol-abcdefgh",
		AssetType:    AssetTypeEBSVolume,
		Location:     "test-region-1b",
		CreationDate: time.Now().AddDate(0, 0, -1),
		Hardware:     "gp2 (50GB)",
	},
	{
		ID:           "vol-a1b2c3d4",
		AssetType:    AssetTypeEBSVolume,
		Location:     "test-region-1c",
		CreationDate: time.Now().AddDate(0, 0, -1),
		Hardware:     "gp2 (20GB)",
	},
}

// Test Data
var testEBSVolumesOutput = &ec2.DescribeVolumesOutput{
	Volumes: []*ec2.Volume{
		{
			VolumeId:         aws.String(testEBSVolumeRows[0].ID),
			VolumeType:       aws.String("gp2"),
			CreateTime:       aws.Time(testEBSVolumeRows[0].CreationDate),
			AvailabilityZone: aws.String(testEBSVolumeRows[0].Location),
			Size:             aws.Int64(100),
			Tags: []*ec2.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String(testEBSVolumeRows[0].Application),
				},
				{
					Key:   aws.String("extra tag"),
					Value: aws.String("testval"),
				},
			},
		},
		{

			VolumeId:         aws.String(testEBSVolumeRows[1].ID),
			VolumeType:       aws.String("gp2"),
			CreateTime:       aws.Time(testEBSVolumeRows[1].CreationDate),
			AvailabilityZone: aws.String(testEBSVolumeRows[1].Location),
			Size:             aws.Int64(50),
		},
		{

			VolumeId:         aws.String(testEBSVolumeRows[2].ID),
			VolumeType:       aws.String("gp2"),
			CreateTime:       aws.Time(testEBSVolumeRows[2].CreationDate),
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

	d.Load([]string{"test-region"}, []string{ServiceEBS})

	var count int
	d.MapRows(func(row inventory.Row) error {
		require.Equal(t, testEBSVolumeRows[count], row)
		count++
		return nil
	})
	require.Equal(t, 3, count)
}

func TestLoadEBSVolumesLogsError(t *testing.T) {
	var output bytes.Buffer
	buf := bufio.NewWriter(&output)

	logger := logrus.New()
	logger.SetOutput(buf)

	d := New(logger, TestClients{EC2: EBSErrorMock{}})

	d.Load([]string{"test-region"}, []string{ServiceEBS})

	buf.Flush()
	require.Contains(t, output.String(), testError.Error())
}
