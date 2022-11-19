package awsdata_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/aws/aws-sdk-go/service/codecommit/codecommitiface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/sudoinclabs/awsinventory/internal/awsdata"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
)

var testCodeCommitRepositoryRows = []inventory.Row{
	{
		UniqueAssetIdentifier: "TestRepository1-2eb7bb38-90f4-4b84-8960-c2650bb8a774",
		Virtual:               true,
		DNSNameOrURL:          "https://git-codecommit.us-east-1.amazonaws.com/v1/repos/TestRepository1",
		Location:              DefaultRegion,
		AssetType:             AssetTypeCodeCommitRepository,
		SerialAssetTagNumber:  "arn:aws:codecommit:us-east-1:123456789012:TestRepository1",
		Function:              "Test repository 1",
	},
	{
		UniqueAssetIdentifier: "TestRepository2-204e8c06-c92d-4c11-b151-e4903ef1c9b5",
		Virtual:               true,
		DNSNameOrURL:          "https://git-codecommit.us-east-1.amazonaws.com/v1/repos/TestRepository2",
		Location:              DefaultRegion,
		AssetType:             AssetTypeCodeCommitRepository,
		SerialAssetTagNumber:  "arn:aws:codecommit:us-east-1:123456789012:TestRepository2",
		Function:              "Test repository 2",
	},
	{
		UniqueAssetIdentifier: "TestRepository3-9dfee785-239e-4eb0-82bf-cc4a1f641db0",
		Virtual:               true,
		DNSNameOrURL:          "https://git-codecommit.us-east-1.amazonaws.com/v1/repos/TestRepository3",
		Location:              DefaultRegion,
		AssetType:             AssetTypeCodeCommitRepository,
		SerialAssetTagNumber:  "arn:aws:codecommit:us-east-1:123456789012:TestRepository3",
		Function:              "Test repository 3",
	},
}

// Test Data
var testCodeCommitListRepositoriesOutputPage1 = &codecommit.ListRepositoriesOutput{
	NextToken: aws.String("204e8c06-c92d-4c11-b151-e4903ef1c9b5"),
	Repositories: []*codecommit.RepositoryNameIdPair{
		{
			RepositoryName: aws.String("TestRepository1"),
		},
		{
			RepositoryName: aws.String("TestRepository2"),
		},
	},
}

var testCodeCommitListRepositoriesOutputPage2 = &codecommit.ListRepositoriesOutput{
	Repositories: []*codecommit.RepositoryNameIdPair{
		{
			RepositoryName: aws.String("TestRepository3"),
		},
	},
}

var testCodeCommitBatchGetRepositoriesOutput = &codecommit.BatchGetRepositoriesOutput{
	Repositories: []*codecommit.RepositoryMetadata{
		{
			Arn:                   aws.String(testCodeCommitRepositoryRows[0].SerialAssetTagNumber),
			CloneUrlHttp:          aws.String(testCodeCommitRepositoryRows[0].DNSNameOrURL),
			RepositoryDescription: aws.String(testCodeCommitRepositoryRows[0].Function),
			RepositoryId:          aws.String("2eb7bb38-90f4-4b84-8960-c2650bb8a774"),
			RepositoryName:        aws.String("TestRepository1"),
		},
		{
			Arn:                   aws.String(testCodeCommitRepositoryRows[1].SerialAssetTagNumber),
			CloneUrlHttp:          aws.String(testCodeCommitRepositoryRows[1].DNSNameOrURL),
			RepositoryDescription: aws.String(testCodeCommitRepositoryRows[1].Function),
			RepositoryId:          aws.String("204e8c06-c92d-4c11-b151-e4903ef1c9b5"),
			RepositoryName:        aws.String("TestRepository2"),
		},
		{
			Arn:                   aws.String(testCodeCommitRepositoryRows[2].SerialAssetTagNumber),
			CloneUrlHttp:          aws.String(testCodeCommitRepositoryRows[2].DNSNameOrURL),
			RepositoryDescription: aws.String(testCodeCommitRepositoryRows[2].Function),
			RepositoryId:          aws.String("9dfee785-239e-4eb0-82bf-cc4a1f641db0"),
			RepositoryName:        aws.String("TestRepository3"),
		},
	},
}

// Mocks
type CodeCommitMock struct {
	codecommitiface.CodeCommitAPI
}

func (e CodeCommitMock) ListRepositories(cfg *codecommit.ListRepositoriesInput) (*codecommit.ListRepositoriesOutput, error) {
	if cfg.NextToken == nil {
		return testCodeCommitListRepositoriesOutputPage1, nil
	}

	return testCodeCommitListRepositoriesOutputPage2, nil
}

func (e CodeCommitMock) BatchGetRepositories(cfg *codecommit.BatchGetRepositoriesInput) (*codecommit.BatchGetRepositoriesOutput, error) {
	return testCodeCommitBatchGetRepositoriesOutput, nil
}

type CodeCommitErrorMock struct {
	codecommitiface.CodeCommitAPI
}

func (e CodeCommitErrorMock) ListRepositories(cfg *codecommit.ListRepositoriesInput) (*codecommit.ListRepositoriesOutput, error) {
	return &codecommit.ListRepositoriesOutput{}, testError
}

func (e CodeCommitErrorMock) BatchGetRepositories(cfg *codecommit.BatchGetRepositoriesInput) (*codecommit.BatchGetRepositoriesOutput, error) {
	return &codecommit.BatchGetRepositoriesOutput{}, testError
}

// Tests
func TestCanLoadCodeCommitRepositories(t *testing.T) {
	d := New(logrus.New(), TestClients{CodeCommit: CodeCommitMock{}})

	var count int
	d.Load([]string{DefaultRegion}, []string{ServiceCodeCommit}, func(row inventory.Row) error {
		require.Equal(t, testCodeCommitRepositoryRows[count], row)
		count++
		return nil
	})
	require.Equal(t, 3, count)
}

func TestLoadCodeCommitRepositoriesLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{CodeCommit: CodeCommitErrorMock{}})

	d.Load([]string{DefaultRegion}, []string{ServiceCodeCommit}, nil)

	assertTestErrorWasLogged(t, hook.Entries)
}
