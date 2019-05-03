package awsdata

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/manywho/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeIAMUser is the value used in the AssetType field when fetching IAM users
	AssetTypeIAMUser string = "IAM User"

	// ServiceIAM is the key for the IAM service
	ServiceIAM string = "iam"
)

func (d *AWSData) loadIAMUsers() {
	defer d.wg.Done()

	iamSvc := d.clients.GetIAMClient(ValidRegions[0])

	log := d.log.WithFields(logrus.Fields{
		"region":  "global",
		"service": ServiceIAM,
	})
	log.Info("loading data")
	out, err := iamSvc.ListUsers(&iam.ListUsersInput{})
	if err != nil {
		d.results <- result{Err: err}
		return
	}

	log.Info("processing data")
	for _, u := range out.Users {
		d.results <- result{
			Row: inventory.Row{
				UniqueAssetIdentifier: aws.StringValue(u.UserName),
				Virtual:               true,
				AssetType:             AssetTypeIAMUser,
			},
		}
	}

	log.Info("finished processing data")
}
