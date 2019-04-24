package loader_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/stretchr/testify/require"

	"github.com/itmecho/awsinventory/internal/inventory"
	. "github.com/itmecho/awsinventory/internal/loader"
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

type IAMMock struct {
	iamiface.IAMAPI
}

func (i IAMMock) ListUsers(cfg *iam.ListUsersInput) (*iam.ListUsersOutput, error) {
	return &iam.ListUsersOutput{
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
	}, nil
}

func TestCanLoadIAMUsers(t *testing.T) {
	l := NewLoader()

	l.LoadIAMUsers(IAMMock{})

	require.Len(t, l.Data, 3, "got more than 3 bucket")
	require.Equal(t, testIAMRows, l.Data, "didn't get expected data")
}

type IAMErrorMock struct {
	iamiface.IAMAPI
}

func (i IAMErrorMock) ListUsers(cfg *iam.ListUsersInput) (*iam.ListUsersOutput, error) {
	return &iam.ListUsersOutput{}, testError
}

func TestLoadIAMUsersSendsErrorToChan(t *testing.T) {
	l := NewLoader()

	l.LoadIAMUsers(IAMErrorMock{})

	require.Len(t, l.Errors, 1, "didn't send error to Errors channel")

	select {
	case e := <-l.Errors:
		require.Equal(t, testError, e, "didn't get expected error")
	default:
		t.Fatal("should have received an error")
	}
}
