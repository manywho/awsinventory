package awsdata_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/manywho/awsinventory/internal/awsdata"
	"github.com/manywho/awsinventory/internal/inventory"
)

var testECRImageRows = []inventory.Row{
	{
		UniqueAssetIdentifier: "TestRepo1-sha256:7692c3ad3540bb803c020b3aee66cd8887123234ea0c6e7143c0add73ff431ed",
		Virtual:               true,
		Public:                false,
		DNSNameOrURL:          "012345678910.dkr.ecr.us-east-1.amazonaws.com/TestRepo1",
		Location:              DefaultRegion,
		AssetType:             AssetTypeECRImage,
		Function:              "latest",
		Comments:              "200.6 MB",
		SerialAssetTagNumber:  "sha256:7692c3ad3540bb803c020b3aee66cd8887123234ea0c6e7143c0add73ff431ed",
	},
	{
		UniqueAssetIdentifier: "TestRepo1-sha256:3fc4ccfe745870e2c0d99f71f30ff0656c8dedd41cc1d7d3d376b0dbe685e2f3",
		Virtual:               true,
		Public:                false,
		DNSNameOrURL:          "012345678910.dkr.ecr.us-east-1.amazonaws.com/TestRepo1",
		Location:              DefaultRegion,
		AssetType:             AssetTypeECRImage,
		Function:              "previous,last",
		Comments:              "800.3 MB",
		SerialAssetTagNumber:  "sha256:3fc4ccfe745870e2c0d99f71f30ff0656c8dedd41cc1d7d3d376b0dbe685e2f3",
	},
	{
		UniqueAssetIdentifier: "TestRepo1-sha256:8b5b9db0c13db24256c829aa364aa90c6d2eba318b9232a4ab9313b954d3555f",
		Virtual:               true,
		Public:                false,
		DNSNameOrURL:          "012345678910.dkr.ecr.us-east-1.amazonaws.com/TestRepo1",
		Location:              DefaultRegion,
		AssetType:             AssetTypeECRImage,
		Comments:              "1.1 GB",
		SerialAssetTagNumber:  "sha256:8b5b9db0c13db24256c829aa364aa90c6d2eba318b9232a4ab9313b954d3555f",
	},
}

// Test Data
var testECRDescribeRepositoriesOutput = &ecr.DescribeRepositoriesOutput{
	NextToken: aws.String("arn:aws:ecr:region:012345678910:repository/TestRepo1"),
	Repositories: []*ecr.Repository{
		{
			RepositoryArn:  aws.String("arn:aws:ecr:region:012345678910:repository/TestRepo1"),
			RepositoryName: aws.String("TestRepo1"),
			RepositoryUri:  aws.String(testECRImageRows[0].DNSNameOrURL),
		},
	},
}

var testECRDescribeImagesOutputPage1 = &ecr.DescribeImagesOutput{
	ImageDetails: []*ecr.ImageDetail{
		{
			ImageDigest:      aws.String(testECRImageRows[0].SerialAssetTagNumber),
			ImageSizeInBytes: aws.Int64(210344346),
			ImageTags: []*string{
				aws.String("latest"),
			},
			RepositoryName: aws.String("TestRepo1"),
		},
		{
			ImageDigest:      aws.String(testECRImageRows[1].SerialAssetTagNumber),
			ImageSizeInBytes: aws.Int64(839175373),
			ImageTags: []*string{
				aws.String("previous"),
				aws.String("last"),
			},
			RepositoryName: aws.String("TestRepo1"),
		},
	},
	NextToken: aws.String(testECRImageRows[1].SerialAssetTagNumber),
}

var testECRDescribeImagesOutputPage2 = &ecr.DescribeImagesOutput{
	ImageDetails: []*ecr.ImageDetail{
		{
			ImageDigest:      aws.String(testECRImageRows[2].SerialAssetTagNumber),
			ImageSizeInBytes: aws.Int64(1181115402),
			RepositoryName:   aws.String("TestRepo1"),
		},
	},
}

// Mocks
type ECRMock struct {
	ecriface.ECRAPI
}

func (e ECRMock) DescribeRepositories(cfg *ecr.DescribeRepositoriesInput) (*ecr.DescribeRepositoriesOutput, error) {
	if cfg.NextToken == testECRDescribeRepositoriesOutput.NextToken {
		return &ecr.DescribeRepositoriesOutput{}, nil
	}

	return testECRDescribeRepositoriesOutput, nil
}

func (e ECRMock) DescribeImages(cfg *ecr.DescribeImagesInput) (*ecr.DescribeImagesOutput, error) {
	if cfg.NextToken == nil {
		return testECRDescribeImagesOutputPage1, nil
	}

	return testECRDescribeImagesOutputPage2, nil
}

type ECRErrorMock struct {
	ecriface.ECRAPI
}

func (e ECRErrorMock) DescribeRepositories(cfg *ecr.DescribeRepositoriesInput) (*ecr.DescribeRepositoriesOutput, error) {
	return &ecr.DescribeRepositoriesOutput{}, testError
}

func (e ECRErrorMock) DescribeImages(cfg *ecr.DescribeImagesInput) (*ecr.DescribeImagesOutput, error) {
	return &ecr.DescribeImagesOutput{}, testError
}

// Tests
func TestCanLoadECRImages(t *testing.T) {
	d := New(logrus.New(), TestClients{ECR: ECRMock{}})

	d.Load([]string{DefaultRegion}, []string{ServiceECR})

	var count int
	d.MapRows(func(row inventory.Row) error {
		require.Equal(t, testECRImageRows[count], row)
		count++
		return nil
	})
	require.Equal(t, 3, count)
}

func TestLoadECRImagesLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{ECR: ECRErrorMock{}})

	d.Load([]string{DefaultRegion}, []string{ServiceECR})

	require.Contains(t, hook.LastEntry().Message, testError.Error())
}
