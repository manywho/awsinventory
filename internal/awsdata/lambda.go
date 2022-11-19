package awsdata

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeLambdaFunction is the value used in the AssetType field when fetching Lambda functions
	AssetTypeLambdaFunction string = "Lambda Function"

	// ServiceLambda is the key for the Lambda service
	ServiceLambda string = "lambda"
)

func (d *AWSData) loadLambdaFunctions(region string) {
	defer d.wg.Done()

	lambdaSvc := d.clients.GetLambdaClient(region)

	log := d.log.WithFields(logrus.Fields{
		"region":  region,
		"service": ServiceLambda,
	})

	log.Info("loading data")

	var functions []*lambda.FunctionConfiguration
	done := false
	params := &lambda.ListFunctionsInput{}
	for !done {
		out, err := lambdaSvc.ListFunctions(params)
		if err != nil {
			log.Errorf("failed to list functions: %s", err)
			return
		}

		functions = append(functions, out.Functions...)

		if out.NextMarker == nil {
			done = true
		} else {
			params.Marker = out.NextMarker
		}
	}

	log.Info("processing data")

	for _, f := range functions {
		var vpcID string
		if f.VpcConfig != nil {
			vpcID = aws.StringValue(f.VpcConfig.VpcId)
		}

		d.rows <- inventory.Row{
			UniqueAssetIdentifier:          aws.StringValue(f.FunctionName),
			Virtual:                        true,
			BaselineConfigurationName:      aws.StringValue(f.Version),
			OSNameAndVersion:               "Amazon Linux",
			Location:                       region,
			AssetType:                      AssetTypeLambdaFunction,
			SoftwareDatabaseNameAndVersion: aws.StringValue(f.Runtime),
			Function:                       aws.StringValue(f.Description),
			Comments:                       fmt.Sprintf("%ds, %dMB", aws.Int64Value(f.Timeout), aws.Int64Value(f.MemorySize)),
			SerialAssetTagNumber:           aws.StringValue(f.FunctionArn),
			VLANNetworkID:                  vpcID,
		}
	}

	log.Info("finished processing data")
}
