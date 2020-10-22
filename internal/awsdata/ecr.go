package awsdata

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/manywho/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeECRImage is the value used in the AssetType field when fetching ECR images
	AssetTypeECRImage string = "ECR Image"

	// ServiceECR is the key for the ECR service
	ServiceECR string = "ecr"
)

func (d *AWSData) loadECRImages(region string) {
	defer d.wg.Done()

	ecrSvc := d.clients.GetECRClient(region)

	log := d.log.WithFields(logrus.Fields{
		"region":  region,
		"service": ServiceECR,
	})

	log.Info("loading data")

	var repositories []*ecr.Repository
	done := false
	params := &ecr.DescribeRepositoriesInput{}
	for !done {
		out, err := ecrSvc.DescribeRepositories(params)

		if err != nil {
			d.results <- result{Err: err}
			return
		}

		repositories = append(repositories, out.Repositories...)

		if out.NextToken == nil {
			done = true
		} else {
			params.NextToken = out.NextToken
		}
	}

	log.Info("processing data")

	for _, r := range repositories {
		var images []*ecr.ImageDetail
		done := false
		params := &ecr.DescribeImagesInput{
			RepositoryName: r.RepositoryName,
		}
		for !done {
			out, err := ecrSvc.DescribeImages(params)

			if err != nil {
				d.results <- result{Err: err}
				return
			}

			images = append(images, out.ImageDetails...)

			if out.NextToken == nil {
				done = true
			} else {
				params.NextToken = out.NextToken
			}
		}

		for _, i := range images {
			d.results <- result{
				Row: inventory.Row{
					UniqueAssetIdentifier: fmt.Sprintf("%s-%s", aws.StringValue(i.RepositoryName), aws.StringValue(i.ImageDigest)),
					Virtual:               true,
					Public:                false,
					DNSNameOrURL:          aws.StringValue(r.RepositoryUri),
					Location:              region,
					AssetType:             AssetTypeECRImage,
					Function:              strings.Join(aws.StringValueSlice(i.ImageTags), ","),
					Comments:              humanReadableBytes(aws.Int64Value(i.ImageSizeInBytes)),
					SerialAssetTagNumber:  aws.StringValue(i.ImageDigest),
				},
			}
		}
	}

	log.Info("finished processing data")
}
