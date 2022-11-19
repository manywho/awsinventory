package awsdata

import (
	//"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeRDSInstance is the value used in the AssetType field when fetching RDS instances
	AssetTypeCloudFormationInstance string = "CloudFormation Stack"

	// ServiceRDS is the key for the RDS service
	ServiceCloudFormation string = "cloudformation"
)

func (d *AWSData) loadCloudFormationsInstances(region string) {
	defer d.wg.Done()

	CloudFormationSvc := d.clients.GetCloudFormationClient(region)

	log := d.log.WithFields(logrus.Fields{
		"region":  region,
		"service": ServiceCloudFormation,
	})

	log.Info("loading data")

	var CloudFormationsItems []*cloudformation.Stack
	done := false
	params := &cloudformation.DescribeStacksInput{}
	for !done {
		out, err := CloudFormationSvc.DescribeStacks(params)

		if err != nil {
			log.Errorf("failed to get CloudFormations Stacks: %s", err)
			return
		}

		CloudFormationsItems = append(CloudFormationsItems, out.Stacks...)

		if out.NextToken == nil {
			done = true
		} else {
			params.NextToken = out.NextToken
		}
	}

	log.Info("processing data")

	for _, i := range CloudFormationsItems {
		d.rows <- inventory.Row{
			UniqueAssetIdentifier:          aws.StringValue(i.StackId),
			Virtual:                        true,
			Public:                         false,
			Location:                       region,
			AssetType:                      AssetTypeCloudFormationInstance,
			Comments:                       aws.StringValue(i.StackName),
			SerialAssetTagNumber:           aws.StringValue(i.StackId),
		}
	}

	log.Info("finished processing data")
}
