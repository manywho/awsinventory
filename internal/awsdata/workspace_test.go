package awsdata_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/aws/aws-sdk-go/service/workspaces/workspacesiface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/manywho/awsinventory/internal/awsdata"
	"github.com/manywho/awsinventory/internal/inventory"
)

var testWorkSpacesRows = []inventory.Row{
	{
		UniqueAssetIdentifier:          "test-db-1",
		IPv4orIPv6Address:              "1.1.1.1",
		Virtual:                        true,
		Public:                         false,
		DNSNameOrURL:                   "test-db-1.rds.aws.amazon.com",
		Location:                       DefaultRegion,
		AssetType:                      "Workspace",
		HardwareMakeModel:              "db.t2.medium",
		SoftwareDatabaseVendor:         "mysql",
		Comments:                       "mysql 5.7",
		SerialAssetTagNumber:           "arn:aws:rds:us-east-1:123456789012:db:test-db-1",
		VLANNetworkID:                  "vpc-12345678",
	},
	{
		UniqueAssetIdentifier:          "test-db-1",
		IPv4orIPv6Address:              "1.1.1.1",
		Virtual:                        true,
		Public:                         false,
		DNSNameOrURL:                   "test-db-1.rds.aws.amazon.com",
		Location:                       DefaultRegion,
		AssetType:                      "Workspace",
		HardwareMakeModel:              "db.t2.medium",
		SoftwareDatabaseVendor:         "mysql",
		Comments:                       "mysql 5.7",
		SerialAssetTagNumber:           "arn:aws:rds:us-east-1:123456789012:db:test-db-1",
		VLANNetworkID:                  "vpc-12345678",
	},
}

// Test Data
var testWorkSpacesDescribeOutputPage1 = &workspaces.DescribeWorkspacesOutput{
	Workspaces: []*workspaces.Workspace{
		{
			WorkspaceId:        aws.String(testWorkSpacesRows[0].SerialAssetTagNumber),
			BundleId:               aws.String(testWorkSpacesRows[0].SoftwareDatabaseVendor),
			WorkspaceProperties: &workspaces.WorkspaceProperties {
				ComputeTypeName: aws.String(testWorkSpacesRows[0].HardwareMakeModel),
				RootVolumeSizeGib:aws.Int64(10),
				RunningMode:aws.String("AUTO_STOP"),
				UserVolumeSizeGib:aws.Int64(64),
			},
			IpAddress: aws.String(testWorkSpacesRows[0].IPv4orIPv6Address),
			DirectoryId: aws.String("d-234sdfsd"),
			ComputerName: aws.String(testWorkSpacesRows[0].DNSNameOrURL),
		},
	},
	NextToken: aws.String(testWorkSpacesRows[0].UniqueAssetIdentifier),
}

var testWorkSpacesDescribeOutputPage2 = &workspaces.DescribeWorkspacesOutput{
	Workspaces: []*workspaces.Workspace{
		{
			WorkspaceId:        aws.String(testWorkSpacesRows[1].SerialAssetTagNumber),
			BundleId:               aws.String(testWorkSpacesRows[1].SoftwareDatabaseVendor),
			WorkspaceProperties: &workspaces.WorkspaceProperties {
				ComputeTypeName: aws.String(testWorkSpacesRows[1].HardwareMakeModel),
				RootVolumeSizeGib:aws.Int64(10),
				RunningMode:aws.String("AUTO_STOP"),
				UserVolumeSizeGib:aws.Int64(64),
			},
			IpAddress: aws.String(testWorkSpacesRows[1].IPv4orIPv6Address),
			DirectoryId: aws.String("d-234sdfsd"),
			ComputerName: aws.String(testWorkSpacesRows[1].DNSNameOrURL),
		},
	},
}

// Mocks
type WorkSpacesMock struct {
	workspacesiface.WorkSpacesAPI
}

func (e WorkSpacesMock) DescribeWorkspaces(cfg *workspaces.DescribeWorkspacesInput) (*workspaces.DescribeWorkspacesOutput, error) {
	if cfg.NextToken == nil {
		return testWorkSpacesDescribeOutputPage1, nil
	}

	return testWorkSpacesDescribeOutputPage2, nil
}

type WorkSpacesErrorMock struct {
	workspacesiface.WorkSpacesAPI
}

func (e WorkSpacesErrorMock) DescribeWorkspaces(cfg *workspaces.DescribeWorkspacesInput) (*workspaces.DescribeWorkspacesOutput, error) {
	return &workspaces.DescribeWorkspacesOutput{}, testError
}

// Tests
func TestCanLoadWorkSpaces(t *testing.T) {
	d := New(logrus.New(), TestClients{WorkSpace: WorkSpacesMock{}})

	var count int
	d.Load([]string{DefaultRegion}, []string{ServiceWorkSpace}, func(row inventory.Row) error {
		require.Equal(t, testWorkSpacesRows[count], row)
		count++
		return nil
	})
	require.Equal(t, 2, count)
}

func TestLoadWorkSpacesLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{WorkSpace: WorkSpacesErrorMock{}})

	d.Load([]string{DefaultRegion}, []string{ServiceWorkSpace}, nil)

	assertTestErrorWasLogged(t, hook.Entries)
}
