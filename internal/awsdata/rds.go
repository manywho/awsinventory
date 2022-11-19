package awsdata

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeRDSInstance is the value used in the AssetType field when fetching RDS instances
	AssetTypeRDSInstance string = "RDS Instance"

	// ServiceRDS is the key for the RDS service
	ServiceRDS string = "rds"
)

func (d *AWSData) loadRDSInstances(region string) {
	defer d.wg.Done()

	rdsSvc := d.clients.GetRDSClient(region)

	log := d.log.WithFields(logrus.Fields{
		"region":  region,
		"service": ServiceRDS,
	})

	log.Info("loading data")

	var dbInstances []*rds.DBInstance
	done := false
	params := &rds.DescribeDBInstancesInput{}
	for !done {
		out, err := rdsSvc.DescribeDBInstances(params)

		if err != nil {
			log.Errorf("failed to describe db instances: %s", err)
			return
		}

		dbInstances = append(dbInstances, out.DBInstances...)

		if out.Marker == nil {
			done = true
		} else {
			params.Marker = out.Marker
		}
	}

	log.Info("processing data")

	for _, i := range dbInstances {
		d.rows <- inventory.Row{
			UniqueAssetIdentifier:          aws.StringValue(i.DBInstanceIdentifier),
			Virtual:                        true,
			Public:                         aws.BoolValue(i.PubliclyAccessible),
			DNSNameOrURL:                   aws.StringValue(i.Endpoint.Address),
			Location:                       region,
			AssetType:                      AssetTypeRDSInstance,
			HardwareMakeModel:              aws.StringValue(i.DBInstanceClass),
			SoftwareDatabaseVendor:         aws.StringValue(i.Engine),
			SoftwareDatabaseNameAndVersion: fmt.Sprintf("%s %s", aws.StringValue(i.Engine), aws.StringValue(i.EngineVersion)),
			SerialAssetTagNumber:           aws.StringValue(i.DBInstanceArn),
			VLANNetworkID:                  aws.StringValue(i.DBSubnetGroup.VpcId),
		}
	}

	log.Info("finished processing data")
}
