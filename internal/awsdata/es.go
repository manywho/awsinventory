package awsdata

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeElasticsearchDomain is the value used in the AssetType field when fetching Elasticsearch domains
	AssetTypeElasticsearchDomain string = "Elasticsearch Domain"

	// ServiceElasticsearchService is the key for the ElasticsearchService service
	ServiceElasticsearchService string = "es"
)

func (d *AWSData) loadElasticsearchDomains(region string) {
	defer d.wg.Done()

	elasticsearchserviceSvc := d.clients.GetElasticsearchServiceClient(region)

	log := d.log.WithFields(logrus.Fields{
		"region":  region,
		"service": ServiceElasticsearchService,
	})

	log.Info("loading data")

	out, err := elasticsearchserviceSvc.ListDomainNames(&elasticsearchservice.ListDomainNamesInput{})
	if err != nil {
		log.Errorf("failed to list domain names: %s", err)
		return
	}

	var domains []string
	for _, domain := range out.DomainNames {
		domains = append(domains, aws.StringValue(domain.DomainName))
	}

	log.Info("processing data")

	if len(domains) == 0 {
		log.Info("no data found; bailing early")
		return
	}

	// API call only accepts 5 domains at a time
	for i := 0; i+1 < len(domains); i += 5 {
		var j int = i + 5
		if j > len(domains) {
			j = len(domains)
		}

		out, err := elasticsearchserviceSvc.DescribeElasticsearchDomains(&elasticsearchservice.DescribeElasticsearchDomainsInput{
			DomainNames: aws.StringSlice(domains[i:j]),
		})
		if err != nil {
			log.Errorf("failed to describe elasticsearch domains: %s", err)
			continue
		}
		for _, c := range out.DomainStatusList {
			d.rows <- inventory.Row{
				UniqueAssetIdentifier:          aws.StringValue(c.DomainName),
				Virtual:                        true,
				Public:                         false,
				DNSNameOrURL:                   aws.StringValue(c.Endpoints["vpc"]),
				Location:                       region,
				AssetType:                      AssetTypeElasticsearchDomain,
				HardwareMakeModel:              aws.StringValue(c.ElasticsearchClusterConfig.InstanceType),
				SoftwareDatabaseVendor:         "Elastic",
				SoftwareDatabaseNameAndVersion: fmt.Sprintf("Elasticsearch %s", aws.StringValue(c.ElasticsearchVersion)),
				SerialAssetTagNumber:           aws.StringValue(c.ARN),
				VLANNetworkID:                  aws.StringValue(c.VPCOptions.VPCId),
			}
		}
	}

	log.Info("finished processing data")
}
