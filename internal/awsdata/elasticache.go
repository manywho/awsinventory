package awsdata

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/manywho/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeElastiCacheNode is the value used in the AssetType field when fetching ElastiCache nodes
	AssetTypeElastiCacheNode string = "ElastiCache Node"

	// ServiceElastiCache is the key for the ElastiCache service
	ServiceElastiCache string = "elasticache"
)

func (d *AWSData) loadElastiCacheNodes(region string) {
	defer d.wg.Done()

	elasticacheSvc := d.clients.GetElastiCacheClient(region)

	log := d.log.WithFields(logrus.Fields{
		"region":  region,
		"service": ServiceElastiCache,
	})

	log.Info("loading data")

	var cacheClusters []*elasticache.CacheCluster
	done := false
	params := &elasticache.DescribeCacheClustersInput{
		ShowCacheNodeInfo: aws.Bool(true),
	}
	for !done {
		out, err := elasticacheSvc.DescribeCacheClusters(params)
		if err != nil {
			d.results <- result{Err: err}
			return
		}

		cacheClusters = append(cacheClusters, out.CacheClusters...)

		if out.Marker == nil {
			done = true
		} else {
			params.Marker = out.Marker
		}
	}

	log.Info("processing data")

	for _, c := range cacheClusters {
		var vpcId string
		groups, err := elasticacheSvc.DescribeCacheSubnetGroups(&elasticache.DescribeCacheSubnetGroupsInput{
			CacheSubnetGroupName: c.CacheSubnetGroupName,
		})
		if err != nil {
			d.results <- result{Err: err}
		} else if len(groups.CacheSubnetGroups) > 0 {
			vpcId = aws.StringValue(groups.CacheSubnetGroups[0].VpcId)
		}

		for _, n := range c.CacheNodes {
			d.results <- result{
				Row: inventory.Row{
					UniqueAssetIdentifier:          fmt.Sprintf("%s-%s", aws.StringValue(c.CacheClusterId), aws.StringValue(n.CacheNodeId)),
					Virtual:                        true,
					Public:                         false,
					DNSNameOrURL:                   aws.StringValue(n.Endpoint.Address),
					BaselineConfigurationName:      aws.StringValue(c.CacheParameterGroup.CacheParameterGroupName),
					Location:                       region,
					AssetType:                      AssetTypeElastiCacheNode,
					HardwareMakeModel:              aws.StringValue(c.CacheNodeType),
					SoftwareDatabaseVendor:         aws.StringValue(c.Engine),
					SoftwareDatabaseNameAndVersion: fmt.Sprintf("%s %s", aws.StringValue(c.Engine), aws.StringValue(c.EngineVersion)),
					SerialAssetTagNumber:           aws.StringValue(c.ARN),
					VLANNetworkID:                  vpcId,
				},
			}
		}
	}

	log.Info("finished processing data")
}
