package awsdata

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
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
			log.Errorf("failed to list tables: %s", err)
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
		d.wg.Add(1)
		go d.processDynamoDBTable(log, dynamodbSvc, t, region)
	}

	log.Info("finished processing data")
}

func (d *AWSData) processDynamoDBTable(log *logrus.Entry, dynamodbSvc dynamodbiface.DynamoDBAPI, table *string, region string) {
	defer d.wg.Done()

	out, err := dynamodbSvc.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: table,
	})
	if err != nil {
		log.Errorf("failed to describe table %s: %s", aws.StringValue(table), err)
		return
	}

	d.rows <- inventory.Row{
		UniqueAssetIdentifier:          aws.StringValue(out.Table.TableName),
		Virtual:                        true,
		Public:                         false,
		Location:                       region,
		AssetType:                      AssetTypeDynamoDBTable,
		SoftwareDatabaseVendor:         "Amazon",
		SoftwareDatabaseNameAndVersion: "DynamoDB",
		Comments:                       humanReadableBytes(aws.Int64Value(out.Table.TableSizeBytes)),
		SerialAssetTagNumber:           aws.StringValue(out.Table.TableArn),
	}
}
