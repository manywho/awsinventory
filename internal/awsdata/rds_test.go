package awsdata_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/sudoinclabs/awsinventory/internal/awsdata"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
)

var testRDSInstanceRows = []inventory.Row{
	{
		UniqueAssetIdentifier:          "test-db-1",
		Virtual:                        true,
		Public:                         false,
		DNSNameOrURL:                   "test-db-1.rds.aws.amazon.com",
		Location:                       DefaultRegion,
		AssetType:                      "RDS Instance",
		HardwareMakeModel:              "db.t2.medium",
		SoftwareDatabaseVendor:         "mysql",
		SoftwareDatabaseNameAndVersion: "mysql 5.7",
		SerialAssetTagNumber:           "arn:aws:rds:us-east-1:123456789012:db:test-db-1",
		VLANNetworkID:                  "vpc-12345678",
	},
	{
		UniqueAssetIdentifier:          "test-db-2",
		Virtual:                        true,
		Public:                         false,
		DNSNameOrURL:                   "test-db-2.rds.aws.amazon.com",
		Location:                       DefaultRegion,
		AssetType:                      "RDS Instance",
		HardwareMakeModel:              "db.t2.small",
		SoftwareDatabaseVendor:         "postgres",
		SoftwareDatabaseNameAndVersion: "postgres 9.6",
		SerialAssetTagNumber:           "arn:aws:rds:us-east-1:123456789012:db:test-db-2",
		VLANNetworkID:                  "vpc-abcdefgh",
	},
	{
		UniqueAssetIdentifier:          "test-db-3",
		Virtual:                        true,
		Public:                         true,
		DNSNameOrURL:                   "test-db-3.rds.aws.amazon.com",
		Location:                       DefaultRegion,
		AssetType:                      "RDS Instance",
		HardwareMakeModel:              "db.m4.large",
		SoftwareDatabaseVendor:         "postgres",
		SoftwareDatabaseNameAndVersion: "postgres 10.0",
		SerialAssetTagNumber:           "arn:aws:rds:us-east-1:123456789012:db:test-db-3",
		VLANNetworkID:                  "vpc-a1b2c3d4",
	},
}

// Test Data
var testRDSDescribeDBInstancesOutputPage1 = &rds.DescribeDBInstancesOutput{
	DBInstances: []*rds.DBInstance{
		{
			DBInstanceIdentifier: aws.String(testRDSInstanceRows[0].UniqueAssetIdentifier),
			DBInstanceArn:        aws.String(testRDSInstanceRows[0].SerialAssetTagNumber),
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
			DBInstanceArn:        aws.String(testRDSInstanceRows[1].SerialAssetTagNumber),
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
	},
	Marker: aws.String(testRDSInstanceRows[1].UniqueAssetIdentifier),
}

var testRDSDescribeDBInstancesOutputPage2 = &rds.DescribeDBInstancesOutput{
	DBInstances: []*rds.DBInstance{
		{
			DBInstanceIdentifier: aws.String(testRDSInstanceRows[2].UniqueAssetIdentifier),
			DBInstanceArn:        aws.String(testRDSInstanceRows[2].SerialAssetTagNumber),
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
	if cfg.Marker == nil {
		return testRDSDescribeDBInstancesOutputPage1, nil
	}

	return testRDSDescribeDBInstancesOutputPage2, nil
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

	var count int
	d.Load([]string{DefaultRegion}, []string{ServiceRDS}, func(row inventory.Row) error {
		require.Equal(t, testRDSInstanceRows[count], row)
		count++
		return nil
	})
	require.Equal(t, 3, count)
}

func TestLoadRDSInstancesLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{RDS: RDSErrorMock{}})

	d.Load([]string{DefaultRegion}, []string{ServiceRDS}, nil)

	assertTestErrorWasLogged(t, hook.Entries)
}
