package awsdata_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/sudoinclabs/awsinventory/internal/awsdata"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
)

var testLambdaFunctionRows = []inventory.Row{
	{
		UniqueAssetIdentifier:          "function-1",
		Virtual:                        true,
		BaselineConfigurationName:      "2",
		OSNameAndVersion:               "Amazon Linux",
		Location:                       DefaultRegion,
		AssetType:                      AssetTypeLambdaFunction,
		SoftwareDatabaseNameAndVersion: "python2.7",
		Function:                       "lambda function 1",
		Comments:                       "10s, 128MB",
		SerialAssetTagNumber:           "arn:aws:lambda:us-east-1:012345678910:function:function-1",
		VLANNetworkID:                  "vpc-12345678",
	},
	{
		UniqueAssetIdentifier:          "function-2",
		Virtual:                        true,
		BaselineConfigurationName:      "4",
		OSNameAndVersion:               "Amazon Linux",
		Location:                       DefaultRegion,
		AssetType:                      AssetTypeLambdaFunction,
		SoftwareDatabaseNameAndVersion: "nodejs12.x",
		Function:                       "lambda function 2",
		Comments:                       "5s, 256MB",
		SerialAssetTagNumber:           "arn:aws:lambda:us-east-1:012345678910:function:function-2",
		VLANNetworkID:                  "vpc-abcdefgh",
	},
	{
		UniqueAssetIdentifier:          "function-3",
		Virtual:                        true,
		BaselineConfigurationName:      "LATEST",
		OSNameAndVersion:               "Amazon Linux",
		Location:                       DefaultRegion,
		AssetType:                      AssetTypeLambdaFunction,
		SoftwareDatabaseNameAndVersion: "go1.x",
		Function:                       "lambda function 3",
		Comments:                       "25s, 512MB",
		SerialAssetTagNumber:           "arn:aws:lambda:us-east-1:012345678910:function:function-3",
		VLANNetworkID:                  "vpc-a1b2c3d4",
	},
}

// Test Data
var testLambdaListFunctionsOutputPage1 = &lambda.ListFunctionsOutput{
	Functions: []*lambda.FunctionConfiguration{
		{
			Description:  aws.String("lambda function 1"),
			FunctionArn:  aws.String(testLambdaFunctionRows[0].SerialAssetTagNumber),
			FunctionName: aws.String(testLambdaFunctionRows[0].UniqueAssetIdentifier),
			MemorySize:   aws.Int64(128),
			Runtime:      aws.String("python2.7"),
			Timeout:      aws.Int64(10),
			Version:      aws.String(testLambdaFunctionRows[0].BaselineConfigurationName),
			VpcConfig: &lambda.VpcConfigResponse{
				VpcId: aws.String(testLambdaFunctionRows[0].VLANNetworkID),
			},
		},
		{
			Description:  aws.String("lambda function 2"),
			FunctionArn:  aws.String(testLambdaFunctionRows[1].SerialAssetTagNumber),
			FunctionName: aws.String(testLambdaFunctionRows[1].UniqueAssetIdentifier),
			MemorySize:   aws.Int64(256),
			Runtime:      aws.String("nodejs12.x"),
			Timeout:      aws.Int64(5),
			Version:      aws.String(testLambdaFunctionRows[1].BaselineConfigurationName),
			VpcConfig: &lambda.VpcConfigResponse{
				VpcId: aws.String(testLambdaFunctionRows[1].VLANNetworkID),
			},
		},
	},
	NextMarker: aws.String(testLambdaFunctionRows[1].UniqueAssetIdentifier),
}

var testLambdaListFunctionsOutputPage2 = &lambda.ListFunctionsOutput{
	Functions: []*lambda.FunctionConfiguration{
		{
			Description:  aws.String("lambda function 3"),
			FunctionArn:  aws.String(testLambdaFunctionRows[2].SerialAssetTagNumber),
			FunctionName: aws.String(testLambdaFunctionRows[2].UniqueAssetIdentifier),
			MemorySize:   aws.Int64(512),
			Runtime:      aws.String("go1.x"),
			Timeout:      aws.Int64(25),
			Version:      aws.String(testLambdaFunctionRows[2].BaselineConfigurationName),
			VpcConfig: &lambda.VpcConfigResponse{
				VpcId: aws.String(testLambdaFunctionRows[2].VLANNetworkID),
			},
		},
	},
}

// Mocks
type LambdaMock struct {
	lambdaiface.LambdaAPI
}

func (e LambdaMock) ListFunctions(cfg *lambda.ListFunctionsInput) (*lambda.ListFunctionsOutput, error) {
	if cfg.Marker == nil {
		return testLambdaListFunctionsOutputPage1, nil
	}

	return testLambdaListFunctionsOutputPage2, nil
}

type LambdaErrorMock struct {
	lambdaiface.LambdaAPI
}

func (e LambdaErrorMock) ListFunctions(cfg *lambda.ListFunctionsInput) (*lambda.ListFunctionsOutput, error) {
	return &lambda.ListFunctionsOutput{}, testError
}

// Tests
func TestCanLoadLambdaFunctions(t *testing.T) {
	d := New(logrus.New(), TestClients{Lambda: LambdaMock{}})

	var count int
	d.Load([]string{DefaultRegion}, []string{ServiceLambda}, func(row inventory.Row) error {
		require.Equal(t, testLambdaFunctionRows[count], row)
		count++
		return nil
	})
	require.Equal(t, 3, count)
}

func TestLoadLambdaFunctionsLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{Lambda: LambdaErrorMock{}})

	d.Load([]string{DefaultRegion}, []string{ServiceLambda}, nil)

	assertTestErrorWasLogged(t, hook.Entries)
}
