package awsdata_test

import (
	"testing"
	"time"

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
		ID:           "test-db-1",
		AssetType:    "RDS Instance",
		Location:     ValidRegions[0],
		CreationDate: time.Now().AddDate(0, 0, -1),
		Application:  "mysql 5.7",
		Hardware:     "db.t2.medium",
		InternalIP:   "test-db-1.rds.aws.amaozn.com",
		VPCID:        "vpc-12345678",
	},
	{
		ID:           "test-db-2",
		AssetType:    "RDS Instance",
		Location:     ValidRegions[0],
		CreationDate: time.Now().AddDate(0, 0, -1),
		Application:  "postgres 9.6",
		Hardware:     "db.t2.small",
		InternalIP:   "test-db-2.rds.aws.amaozn.com",
		VPCID:        "vpc-abcdefgh",
	},
	{
		ID:           "test-db-3",
		AssetType:    "RDS Instance",
		Location:     ValidRegions[0],
		CreationDate: time.Now().AddDate(0, 0, -1),
		Application:  "postgres 10.0",
		Hardware:     "db.m4.large",
		InternalIP:   "test-db-3.rds.aws.amaozn.com",
		ExternalIP:   "publicly accessible",
		VPCID:        "vpc-a1b2c3d4",
	},
}

// Test Data
var testRDSInstanceOutput = &rds.DescribeDBInstancesOutput{
	DBInstances: []*rds.DBInstance{
		{
			DBInstanceIdentifier: aws.String(testRDSInstanceRows[0].ID),
			InstanceCreateTime:   aws.Time(testRDSInstanceRows[0].CreationDate),
			Engine:               aws.String("mysql"),
			EngineVersion:        aws.String("5.7"),
			DBInstanceClass:      aws.String("db.t2.medium"),
			Endpoint: &rds.Endpoint{
				Address: aws.String(testRDSInstanceRows[0].InternalIP),
			},
			PubliclyAccessible: aws.Bool(false),
			DBSubnetGroup: &rds.DBSubnetGroup{
				VpcId: aws.String(testRDSInstanceRows[0].VPCID),
			},
		},
		{
			DBInstanceIdentifier: aws.String(testRDSInstanceRows[1].ID),
			InstanceCreateTime:   aws.Time(testRDSInstanceRows[1].CreationDate),
			Engine:               aws.String("postgres"),
			EngineVersion:        aws.String("9.6"),
			DBInstanceClass:      aws.String("db.t2.small"),
			Endpoint: &rds.Endpoint{
				Address: aws.String(testRDSInstanceRows[1].InternalIP),
			},
			PubliclyAccessible: aws.Bool(false),
			DBSubnetGroup: &rds.DBSubnetGroup{
				VpcId: aws.String(testRDSInstanceRows[1].VPCID),
			},
		},
		{
			DBInstanceIdentifier: aws.String(testRDSInstanceRows[2].ID),
			InstanceCreateTime:   aws.Time(testRDSInstanceRows[2].CreationDate),
			Engine:               aws.String("postgres"),
			EngineVersion:        aws.String("10.0"),
			DBInstanceClass:      aws.String("db.m4.large"),
			Endpoint: &rds.Endpoint{
				Address: aws.String(testRDSInstanceRows[2].InternalIP),
			},
			PubliclyAccessible: aws.Bool(true),
			DBSubnetGroup: &rds.DBSubnetGroup{
				VpcId: aws.String(testRDSInstanceRows[2].VPCID),
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
