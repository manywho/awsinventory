package awsdata

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/manywho/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeELB is the value used in the AssetType field when fetching ELBs
	AssetTypeELB string = "ELB"

	// ServiceELB is the key for the ELB service
	ServiceELB string = "elb"
)

func (d *AWSData) loadELBs(region string) {
	elbSvc := d.clients.GetELBClient(region)

	log := d.log.WithFields(logrus.Fields{
		"region":  region,
		"service": ServiceELB,
	})
	d.wg.Add(1)
	defer d.wg.Done()
	log.Info("loading data")
	out, err := elbSvc.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})
	if err != nil {
		d.results <- result{Err: err}
		return
	}

	log.Info("processing data")
	for _, l := range out.LoadBalancerDescriptions {
		d.results <- result{
			Row: inventory.Row{
				UniqueAssetIdentifier: aws.StringValue(l.LoadBalancerName),
				Virtual:               true,
				DNSNameOrURL:          aws.StringValue(l.DNSName),
				Location:              region,
				AssetType:             AssetTypeELB,
				Function:              aws.StringValue(l.CanonicalHostedZoneName),
				VLANNetworkID:         aws.StringValue(l.VPCId),
			},
		}
	}

	log.Info("finished processing data")
}
