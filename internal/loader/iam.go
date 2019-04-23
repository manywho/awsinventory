package loader

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/itmecho/awsinventory/internal/inventory"
)

// LoadIAMUsers loads the iam user data into the Loader's data
func (l *Loader) LoadIAMUsers(iamSvc iamiface.IAMAPI) {
	out, err := iamSvc.ListUsers(&iam.ListUsersInput{})
	if err != nil {
		l.Errors <- err
		return
	}

	results := make([]inventory.Row, 0)

	for _, u := range out.Users {
		results = append(results, inventory.Row{
			ID:               aws.StringValue(u.UserId),
			AssetType:        "IAM User",
			Location:         "global",
			CreationDate:     aws.TimeValue(u.CreateDate),
			Application:      aws.StringValue(u.UserName),
			PasswordLastUsed: aws.TimeValue(u.PasswordLastUsed),
		})
	}

	l.appendData(results)
}
