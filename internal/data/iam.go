package data

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/manywho/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeIAMUser is the value used in the AssetType field when fetching IAM users
	AssetTypeIAMUser string = "IAM User"

	// ServiceIAM is the key for the IAM service
	ServiceIAM string = "iam"
)

func (d *Data) loadIAMUsers(iamSvc iamiface.IAMAPI) {
	log := d.log.WithFields(logrus.Fields{
		"region":  "global",
		"service": ServiceIAM,
	})
	d.wg.Add(1)
	defer d.wg.Done()
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
				ID:               aws.StringValue(u.UserId),
				AssetType:        AssetTypeIAMUser,
				Location:         "global",
				CreationDate:     aws.TimeValue(u.CreateDate),
				Application:      aws.StringValue(u.UserName),
				PasswordLastUsed: aws.TimeValue(u.PasswordLastUsed),
			},
		}
	}

	log.Info("finished processing data")
}
