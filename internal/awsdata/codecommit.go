package awsdata

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeCodeCommitRepository is the value used in the AssetType field when fetching CodeCommit repositories
	AssetTypeCodeCommitRepository string = "CodeCommit Repository"

	// ServiceCodeCommit is the key for the CodeCommit service
	ServiceCodeCommit string = "codecommit"
)

func (d *AWSData) loadCodeCommitRepositories(region string) {
	defer d.wg.Done()

	codecommitSvc := d.clients.GetCodeCommitClient(region)

	log := d.log.WithFields(logrus.Fields{
		"region":  region,
		"service": ServiceCodeCommit,
	})

	log.Info("loading data")

	var repositories []string
	done := false
	params := &codecommit.ListRepositoriesInput{}
	for !done {
		out, err := codecommitSvc.ListRepositories(params)

		if err != nil {
			log.Errorf("failed to list repositories: %s", err)
			return
		}

		for _, repository := range out.Repositories {
			repositories = append(repositories, aws.StringValue(repository.RepositoryName))
		}

		if out.NextToken == nil {
			done = true
		} else {
			params.NextToken = out.NextToken
		}
	}

	log.Info("processing data")

	if len(repositories) == 0 {
		log.Info("no data found; bailing early")
		return
	}

	// TODO: API call can only handle 100 repository names at a time
	out, err := codecommitSvc.BatchGetRepositories(&codecommit.BatchGetRepositoriesInput{
		RepositoryNames: aws.StringSlice(repositories),
	})
	if err != nil {
		log.Errorf("failed to get repositories: %s", err)
		return
	}

	for _, r := range out.Repositories {
		d.rows <- inventory.Row{
			UniqueAssetIdentifier: fmt.Sprintf("%s-%s", aws.StringValue(r.RepositoryName), aws.StringValue(r.RepositoryId)),
			Virtual:               true,
			DNSNameOrURL:          aws.StringValue(r.CloneUrlHttp),
			Location:              region,
			AssetType:             AssetTypeCodeCommitRepository,
			SerialAssetTagNumber:  aws.StringValue(r.Arn),
			Function:              aws.StringValue(r.RepositoryDescription),
		}
	}

	log.Info("finished processing data")
}
