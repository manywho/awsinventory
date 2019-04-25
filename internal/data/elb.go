package data

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/itmecho/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeELB is the value used in the AssetType field when fetching ELBs
	AssetTypeELB string = "ELB"

	// ServiceELB is the key for the S3 service
	ServiceELB string = "elb"
)

func (d *Data) loadELBs(elbSvc elbiface.ELBAPI, region string) {
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
				ID:           aws.StringValue(l.LoadBalancerName),
				AssetType:    "ELB",
				Location:     region,
				CreationDate: aws.TimeValue(l.CreatedTime),
				Application:  aws.StringValue(l.CanonicalHostedZoneName),
				DNSName:      aws.StringValue(l.DNSName),
				VPCID:        aws.StringValue(l.VPCId),
			},
		}
	}

	log.Info("finished processing data")
}
