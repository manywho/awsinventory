package awsdata

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
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
			log.Errorf("failed to describe clusters: %s", err)
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
		d.wg.Add(1)
		go d.processElastiCacheCacheCluster(log, elasticacheSvc, c, region)
	}

	log.Info("finished processing data")
}

func (d *AWSData) processElastiCacheCacheCluster(log *logrus.Entry, elasticacheSvc elasticacheiface.ElastiCacheAPI, cacheCluster *elasticache.CacheCluster, region string) {
	defer d.wg.Done()

	var vpcID string
	groups, err := elasticacheSvc.DescribeCacheSubnetGroups(&elasticache.DescribeCacheSubnetGroupsInput{
		CacheSubnetGroupName: cacheCluster.CacheSubnetGroupName,
	})
	if err != nil {
		log.Warningf("failed to describe cache subnet groups for %s: %s", aws.StringValue(cacheCluster.CacheClusterId), err)
	} else if len(groups.CacheSubnetGroups) > 0 {
		vpcID = aws.StringValue(groups.CacheSubnetGroups[0].VpcId)
	}

	for _, n := range cacheCluster.CacheNodes {
		d.rows <- inventory.Row{
			UniqueAssetIdentifier:          fmt.Sprintf("%s-%s", aws.StringValue(cacheCluster.CacheClusterId), aws.StringValue(n.CacheNodeId)),
			Virtual:                        true,
			Public:                         false,
			DNSNameOrURL:                   aws.StringValue(n.Endpoint.Address),
			BaselineConfigurationName:      aws.StringValue(cacheCluster.CacheParameterGroup.CacheParameterGroupName),
			Location:                       region,
			AssetType:                      AssetTypeElastiCacheNode,
			HardwareMakeModel:              aws.StringValue(cacheCluster.CacheNodeType),
			SoftwareDatabaseVendor:         aws.StringValue(cacheCluster.Engine),
			SoftwareDatabaseNameAndVersion: fmt.Sprintf("%s %s", aws.StringValue(cacheCluster.Engine), aws.StringValue(cacheCluster.EngineVersion)),
			SerialAssetTagNumber:           aws.StringValue(cacheCluster.ARN),
			VLANNetworkID:                  vpcID,
		}
	}
}
