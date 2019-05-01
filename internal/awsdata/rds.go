package awsdata

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/manywho/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeRDSInstance is the value used in the AssetType field when fetching RDS instances
	AssetTypeRDSInstance string = "RDS Instance"

	// ServiceRDS is the key for the RDS service
	ServiceRDS string = "rds"
)

func (d *AWSData) loadRDSInstances(region string) {
	rdsSvc := d.clients.GetRDSClient(region)

	log := d.log.WithFields(logrus.Fields{
		"region":  region,
		"service": ServiceRDS,
	})
	d.wg.Add(1)
	defer d.wg.Done()
	log.Info("loading data")
	out, err := rdsSvc.DescribeDBInstances(&rds.DescribeDBInstancesInput{})
	if err != nil {
		d.results <- result{Err: err}
		return
	}

	log.Info("processing data")
	for _, i := range out.DBInstances {
		d.results <- result{
			Row: inventory.Row{
				UniqueAssetIdentifier: aws.StringValue(i.DBInstanceIdentifier),
				Virtual:               true,
				Public:                aws.BoolValue(i.PubliclyAccessible),
				DNSNameOrURL:          aws.StringValue(i.Endpoint.Address),
				Location:              region,
				AssetType:             AssetTypeRDSInstance,
				HardwareMakeModel:     aws.StringValue(i.DBInstanceClass),
				// TODO: SoftwareDatabaseVendor
				SoftwareDatabaseNameAndVersion: fmt.Sprintf("%s %s", aws.StringValue(i.Engine), aws.StringValue(i.EngineVersion)),
				VLANNetworkID:                  aws.StringValue(i.DBSubnetGroup.VpcId),
			},
		}
	}

	log.Info("finished processing data")
}
