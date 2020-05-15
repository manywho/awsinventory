package awsdata_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/manywho/awsinventory/internal/awsdata"
	"github.com/manywho/awsinventory/internal/inventory"
)

var testIAMRows = []inventory.Row{
	{
		UniqueAssetIdentifier: "test-user-1",
		Virtual:               true,
		AssetType:             AssetTypeIAMUser,
	},
	{
		UniqueAssetIdentifier: "test-user-2",
		Virtual:               true,
		AssetType:             AssetTypeIAMUser,
	},
	{
		UniqueAssetIdentifier: "test-user-3",
		Virtual:               true,
		AssetType:             AssetTypeIAMUser,
	},
}

// Test Data
var testIAMOutput = &iam.ListUsersOutput{
	Users: []*iam.User{
		{
			UserName: aws.String(testIAMRows[0].UniqueAssetIdentifier),
		},
		{
			UserName: aws.String(testIAMRows[1].UniqueAssetIdentifier),
		},
		{
			UserName: aws.String(testIAMRows[2].UniqueAssetIdentifier),
		},
	},
}

// Mocks
type IAMMock struct {
	iamiface.IAMAPI
}

func (e IAMMock) ListUsers(cfg *iam.ListUsersInput) (*iam.ListUsersOutput, error) {
	return testIAMOutput, nil
}

type IAMErrorMock struct {
	iamiface.IAMAPI
}

func (e IAMErrorMock) ListUsers(cfg *iam.ListUsersInput) (*iam.ListUsersOutput, error) {
	return &iam.ListUsersOutput{}, testError
}

// Tests
func TestCanLoadIAMUsers(t *testing.T) {
	d := New(logrus.New(), TestClients{IAM: IAMMock{}})

	d.Load([]string{}, []string{ServiceIAM})

	var count int
	d.MapRows(func(row inventory.Row) error {
		require.Equal(t, testIAMRows[count], row)
		count++
		return nil
	})
	require.Equal(t, 3, count)
}

func TestLoadIAMUsersLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{IAM: IAMErrorMock{}})

	d.Load([]string{ValidRegions[0]}, []string{ServiceIAM})

	require.Contains(t, hook.LastEntry().Message, testError.Error())
	hook.Reset()
}