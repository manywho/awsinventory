package awsdata_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/manywho/awsinventory/internal/awsdata"
	"github.com/manywho/awsinventory/internal/inventory"
)

var testRDSInstanceRows = []inventory.Row{
	{
		UniqueAssetIdentifier:          "test-db-1",
		AssetType:                      "RDS Instance",
		Location:                       ValidRegions[0],
		SoftwareDatabaseNameAndVersion: "mysql 5.7",
		HardwareMakeModel:              "db.t2.medium",
		DNSNameOrURL:                   "test-db-1.rds.aws.amazon.com",
		VLANNetworkID:                  "vpc-12345678",
	},
	{
		UniqueAssetIdentifier:          "test-db-2",
		AssetType:                      "RDS Instance",
		Location:                       ValidRegions[0],
		SoftwareDatabaseNameAndVersion: "postgres 9.6",
		HardwareMakeModel:              "db.t2.small",
		DNSNameOrURL:                   "test-db-2.rds.aws.amazon.com",
		VLANNetworkID:                  "vpc-abcdefgh",
	},
	{
		UniqueAssetIdentifier:          "test-db-3",
		AssetType:                      "RDS Instance",
		Location:                       ValidRegions[0],
		SoftwareDatabaseNameAndVersion: "postgres 10.0",
		HardwareMakeModel:              "db.m4.large",
		DNSNameOrURL:                   "test-db-3.rds.aws.amazon.com",
		Public:                         true,
		VLANNetworkID:                  "vpc-a1b2c3d4",
	},
}

// Test Data
var testRDSInstanceOutput = &rds.DescribeDBInstancesOutput{
	DBInstances: []*rds.DBInstance{
		{
			DBInstanceIdentifier: aws.String(testRDSInstanceRows[0].UniqueAssetIdentifier),
			Engine:               aws.String("mysql"),
			EngineVersion:        aws.String("5.7"),
			DBInstanceClass:      aws.String("db.t2.medium"),
			Endpoint: &rds.Endpoint{
				Address: aws.String(testRDSInstanceRows[0].DNSNameOrURL),
			},
			PubliclyAccessible: aws.Bool(false),
			DBSubnetGroup: &rds.DBSubnetGroup{
				VpcId: aws.String(testRDSInstanceRows[0].VLANNetworkID),
			},
		},
		{
			DBInstanceIdentifier: aws.String(testRDSInstanceRows[1].UniqueAssetIdentifier),
			Engine:               aws.String("postgres"),
			EngineVersion:        aws.String("9.6"),
			DBInstanceClass:      aws.String("db.t2.small"),
			Endpoint: &rds.Endpoint{
				Address: aws.String(testRDSInstanceRows[1].DNSNameOrURL),
			},
			PubliclyAccessible: aws.Bool(false),
			DBSubnetGroup: &rds.DBSubnetGroup{
				VpcId: aws.String(testRDSInstanceRows[1].VLANNetworkID),
			},
		},
		{
			DBInstanceIdentifier: aws.String(testRDSInstanceRows[2].UniqueAssetIdentifier),
			Engine:               aws.String("postgres"),
			EngineVersion:        aws.String("10.0"),
			DBInstanceClass:      aws.String("db.m4.large"),
			Endpoint: &rds.Endpoint{
				Address: aws.String(testRDSInstanceRows[2].DNSNameOrURL),
			},
			PubliclyAccessible: aws.Bool(true),
			DBSubnetGroup: &rds.DBSubnetGroup{
				VpcId: aws.String(testRDSInstanceRows[2].VLANNetworkID),
			},
		},
	},
}

// Mocks
type RDSMock struct {
	rdsiface.RDSAPI
}

func (e RDSMock) DescribeDBInstances(cfg *rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error) {
	return testRDSInstanceOutput, nil
}

type RDSErrorMock struct {
	rdsiface.RDSAPI
}

func (e RDSErrorMock) DescribeDBInstances(cfg *rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error) {
	return &rds.DescribeDBInstancesOutput{}, testError
}

// Tests
func TestCanLoadRDSInstances(t *testing.T) {
	d := New(logrus.New(), TestClients{RDS: RDSMock{}})

	d.Load([]string{ValidRegions[0]}, []string{ServiceRDS})

	var count int
	d.MapRows(func(row inventory.Row) error {
		require.Equal(t, testRDSInstanceRows[count], row)
		count++
		return nil
	})
	require.Equal(t, 3, count)
}

func TestLoadRDSInstancesLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{RDS: RDSErrorMock{}})

	d.Load([]string{ValidRegions[0]}, []string{ServiceRDS})

	require.Contains(t, hook.LastEntry().Message, testError.Error())
}
