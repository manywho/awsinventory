package awsdata_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/sudoinclabs/awsinventory/internal/awsdata"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
)

var testIAMRows = []inventory.Row{
	{
		UniqueAssetIdentifier: "test-user-1",
		Virtual:               true,
		AssetType:             AssetTypeIAMUser,
		SerialAssetTagNumber:  "arn:aws:iam::123456789012:user/test-user-1",
	},
	{
		UniqueAssetIdentifier: "test-user-2",
		Virtual:               true,
		AssetType:             AssetTypeIAMUser,
		SerialAssetTagNumber:  "arn:aws:iam::123456789012:user/test-user-2",
	},
	{
		UniqueAssetIdentifier: "test-user-3",
		Virtual:               true,
		AssetType:             AssetTypeIAMUser,
		SerialAssetTagNumber:  "arn:aws:iam::123456789012:user/test-user-3",
	},
}

// Test Data
var testIAMListUsersOutputPage1 = &iam.ListUsersOutput{
	IsTruncated: aws.Bool(true),
	Marker:      aws.String(testIAMRows[1].UniqueAssetIdentifier),
	Users: []*iam.User{
		{
			UserName: aws.String(testIAMRows[0].UniqueAssetIdentifier),
			Arn:      aws.String(testIAMRows[0].SerialAssetTagNumber),
		},
		{
			UserName: aws.String(testIAMRows[1].UniqueAssetIdentifier),
			Arn:      aws.String(testIAMRows[1].SerialAssetTagNumber),
		},
	},
}

var testIAMListUsersOutputPage2 = &iam.ListUsersOutput{
	Users: []*iam.User{
		{
			UserName: aws.String(testIAMRows[2].UniqueAssetIdentifier),
			Arn:      aws.String(testIAMRows[2].SerialAssetTagNumber),
		},
	},
}

// Mocks
type IAMMock struct {
	iamiface.IAMAPI
}

func (e IAMMock) ListUsers(cfg *iam.ListUsersInput) (*iam.ListUsersOutput, error) {
	if cfg.Marker == nil {
		return testIAMListUsersOutputPage1, nil
	}

	return testIAMListUsersOutputPage2, nil
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

	var count int
	d.Load([]string{}, []string{ServiceIAM}, func(row inventory.Row) error {
		require.Equal(t, testIAMRows[count], row)
		count++
		return nil
	})
	require.Equal(t, 3, count)
}

func TestLoadIAMUsersLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{IAM: IAMErrorMock{}})

	d.Load([]string{DefaultRegion}, []string{ServiceIAM}, nil)

	assertTestErrorWasLogged(t, hook.Entries)
}
