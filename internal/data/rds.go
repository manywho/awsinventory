package data

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/itmecho/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeRDSInstance is the value used in the AssetType field when fetching RDS instances
	AssetTypeRDSInstance string = "RDS Instance"

	// ServiceRDS is the key for the RDS service
	ServiceRDS string = "rds"
)

func (d *Data) loadRDSInstances(rdsSvc rdsiface.RDSAPI, region string) {
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

		var public string
		if aws.BoolValue(i.PubliclyAccessible) {
			public = "publicly accessible"
		}

		d.results <- result{
			Row: inventory.Row{
				ID:           aws.StringValue(i.DBInstanceIdentifier),
				AssetType:    AssetTypeRDSInstance,
				Location:     region,
				CreationDate: aws.TimeValue(i.InstanceCreateTime),
				Application:  fmt.Sprintf("%s %s", aws.StringValue(i.Engine), aws.StringValue(i.EngineVersion)),
				Hardware:     aws.StringValue(i.DBInstanceClass),
				InternalIP:   aws.StringValue(i.Endpoint.Address),
				ExternalIP:   public,
				VPCID:        aws.StringValue(i.DBSubnetGroup.VpcId),
			},
		}
	}

	log.Info("finished processing data")
}
