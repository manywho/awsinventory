package awsdata

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/manywho/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeALB is the value used in the AssetType field when fetching ALBs
	AssetTypeALB string = "ALB"

	// AssetTypeNLB is the value used in the AssetType field when fetching NLBs
	AssetTypeNLB string = "NLB"

	// ServiceELBv2 is the key for the ELBV2 service
	ServiceELBV2 string = "elbv2"
)

func (d *AWSData) loadELBV2s(region string) {
	defer d.wg.Done()

	elbv2Svc := d.clients.GetELBV2Client(region)

	log := d.log.WithFields(logrus.Fields{
		"region":  region,
		"service": ServiceELBV2,
	})

	log.Info("loading data")

	var loadBalancers []*elbv2.LoadBalancer
	done := false
	params := &elbv2.DescribeLoadBalancersInput{}
	for !done {
		out, err := elbv2Svc.DescribeLoadBalancers(params)

		if err != nil {
			d.results <- result{Err: err}
			return
		}

		loadBalancers = append(loadBalancers, out.LoadBalancers...)

		if out.NextMarker == nil {
			done = true
		} else {
			params.Marker = out.NextMarker
		}
	}

	log.Info("processing data")

	for _, l := range loadBalancers {
		var assettype string
		if aws.StringValue(l.Type) == "application" {
			assettype = AssetTypeALB
		} else if aws.StringValue(l.Type) == "network" {
			assettype = AssetTypeNLB
		}

		var public bool
		if aws.StringValue(l.Scheme) == "internet-facing" {
			public = true
		} else if aws.StringValue(l.Scheme) == "internal" {
			public = false
		}

		d.results <- result{
			Row: inventory.Row{
				UniqueAssetIdentifier: aws.StringValue(l.LoadBalancerName),
				Virtual:               true,
				Public:                public,
				DNSNameOrURL:          aws.StringValue(l.DNSName),
				Location:              region,
				AssetType:             assettype,
				SerialAssetTagNumber:  aws.StringValue(l.LoadBalancerArn),
				VLANNetworkID:         aws.StringValue(l.VpcId),
			},
		}
	}

	log.Info("finished processing data")
}
