package loader

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/itmecho/awsinventory/internal/inventory"
)

// LoadELBs loads the elb data form the given region into the Loader's data
func (l *Loader) LoadELBs(elbSvc elbiface.ELBAPI, region string) {
	out, err := elbSvc.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})
	if err != nil {
		l.Errors <- err
		return
	}

	results := make([]inventory.Row, 0)

	for _, l := range out.LoadBalancerDescriptions {
		results = append(results, inventory.Row{
			ID:           aws.StringValue(l.LoadBalancerName),
			AssetType:    "ELB",
			Location:     region,
			CreationDate: aws.TimeValue(l.CreatedTime),
			Application:  aws.StringValue(l.CanonicalHostedZoneName),
			DNSName:      aws.StringValue(l.DNSName),
			VPCID:        aws.StringValue(l.VPCId),
		})
	}

	l.appendData(results)
}
