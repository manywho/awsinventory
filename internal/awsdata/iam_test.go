package awsdata_test

import (
	"testing"
	"time"

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
		ID:               "ABCDEFGH",
		AssetType:        "IAM User",
		Location:         "global",
		Application:      "test-user-1",
		CreationDate:     time.Now().AddDate(0, 0, -10),
		PasswordLastUsed: time.Now().AddDate(0, 0, -1),
	},
	{
		ID:               "12345678",
		AssetType:        "IAM User",
		Location:         "global",
		Application:      "test-user-2",
		CreationDate:     time.Now().AddDate(0, 0, -20),
		PasswordLastUsed: time.Now().AddDate(0, 0, -2),
	},
	{
		ID:               "A1B2C3D4E",
		AssetType:        "IAM User",
		Location:         "global",
		CreationDate:     time.Now().AddDate(0, 0, -30),
		Application:      "test-user-3",
		PasswordLastUsed: time.Now().AddDate(0, 0, -3),
	},
}

// Test Data
var testIAMOutput = &iam.ListUsersOutput{
	Users: []*iam.User{
		{
			UserId:           aws.String(testIAMRows[0].ID),
			UserName:         aws.String(testIAMRows[0].Application),
			CreateDate:       aws.Time(testIAMRows[0].CreationDate),
			PasswordLastUsed: aws.Time(testIAMRows[0].PasswordLastUsed),
		},
		{
			UserId:           aws.String(testIAMRows[1].ID),
			UserName:         aws.String(testIAMRows[1].Application),
			CreateDate:       aws.Time(testIAMRows[1].CreationDate),
			PasswordLastUsed: aws.Time(testIAMRows[1].PasswordLastUsed),
		},
		{
			UserId:           aws.String(testIAMRows[2].ID),
			UserName:         aws.String(testIAMRows[2].Application),
			CreateDate:       aws.Time(testIAMRows[2].CreationDate),
			PasswordLastUsed: aws.Time(testIAMRows[2].PasswordLastUsed),
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

	d.Load([]string{ValidRegions[0]}, []string{ServiceIAM})

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
