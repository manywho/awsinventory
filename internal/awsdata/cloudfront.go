package awsdata

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeCloudFrontDistribution is the value used in the AssetType field when fetching CloudFront distributions
	AssetTypeCloudFrontDistribution string = "CloudFront Distribution"

	// ServiceCloudFront is the key for the CloudFront service
	ServiceCloudFront string = "cloudfront"
)

func (d *AWSData) loadCloudFrontDistributions() {
	defer d.wg.Done()

	cloudfrontSvc := d.clients.GetCloudFrontClient(DefaultRegion)

	log := d.log.WithFields(logrus.Fields{
		"region":  "global",
		"service": ServiceCloudFront,
	})

	log.Info("loading data")

	var distributions []*cloudfront.DistributionSummary
	done := false
	params := &cloudfront.ListDistributionsInput{}
	for !done {
		out, err := cloudfrontSvc.ListDistributions(params)

		if err != nil {
			log.Errorf("failed to list distributions: %s", err)
			return
		}

		distributions = append(distributions, out.DistributionList.Items...)

		if aws.BoolValue(out.DistributionList.IsTruncated) {
			params.Marker = out.DistributionList.NextMarker
		} else {
			done = true
		}
	}

	log.Info("processing data")

	for _, dist := range distributions {
		var domainNames []string
		var origins []string

		domainNames = append(domainNames, aws.StringValue(dist.DomainName))

		domainNames = append(domainNames, aws.StringValueSlice(dist.Aliases.Items)...)

		for _, origin := range dist.Origins.Items {
			origins = append(origins, aws.StringValue(origin.DomainName))
		}

		d.rows <- inventory.Row{
			UniqueAssetIdentifier:     aws.StringValue(dist.Id),
			Virtual:                   true,
			Public:                    true,
			DNSNameOrURL:              strings.Join(domainNames, "\n"),
			BaselineConfigurationName: strings.Join(origins, "\n"),
			AssetType:                 AssetTypeCloudFrontDistribution,
			Function:                  aws.StringValue(dist.Comment),
			SerialAssetTagNumber:      aws.StringValue(dist.ARN),
		}
	}

	log.Info("finished processing data")
}
