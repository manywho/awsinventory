package awsdata

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/manywho/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeDynamoDBTable is the value used in the AssetType field when fetching DynamoDB tables
	AssetTypeDynamoDBTable string = "DynamoDB Table"

	// ServiceDynamoDB is the key for the DynamoDB service
	ServiceDynamoDB string = "dynamodb"
)

func (d *AWSData) loadDynamoDBTables(region string) {
	defer d.wg.Done()

	dynamodbSvc := d.clients.GetDynamoDBClient(region)

	log := d.log.WithFields(logrus.Fields{
		"region":  region,
		"service": ServiceDynamoDB,
	})

	log.Info("loading data")

	var tables []*string
	done := false
	params := &dynamodb.ListTablesInput{}
	for !done {
		out, err := dynamodbSvc.ListTables(params)

		if err != nil {
			d.results <- result{Err: err}
			return
		}

		tables = append(tables, out.TableNames...)

		if out.LastEvaluatedTableName == nil {
			done = true
		} else {
			params.ExclusiveStartTableName = out.LastEvaluatedTableName
		}
	}

	log.Info("processing data")

	for _, t := range tables {
		out, err := dynamodbSvc.DescribeTable(&dynamodb.DescribeTableInput{
			TableName: t,
		})
		if err != nil {
			d.results <- result{Err: err}
			return
		}

		d.results <- result{
			Row: inventory.Row{
				UniqueAssetIdentifier:  aws.StringValue(out.Table.TableName),
				Virtual:                true,
				Location:               region,
				AssetType:              AssetTypeDynamoDBTable,
				SoftwareDatabaseVendor: "Amazon",
				Comments:               humanReadableBytes(aws.Int64Value(out.Table.TableSizeBytes)),
				SerialAssetTagNumber:   aws.StringValue(out.Table.TableArn),
			},
		}
	}

	log.Info("finished processing data")
}
