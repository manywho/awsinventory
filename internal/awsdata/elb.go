package awsdata

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeELB is the value used in the AssetType field when fetching ELBs
	AssetTypeELB string = "ELB"

	// ServiceELB is the key for the ELB service
	ServiceELB string = "elb"
)

func (d *AWSData) loadELBs(region string) {
	defer d.wg.Done()

	ec2Svc := d.clients.GetEC2Client(region)
	elbSvc := d.clients.GetELBClient(region)

	log := d.log.WithFields(logrus.Fields{
		"region":  region,
		"service": ServiceELB,
	})

	log.Info("loading data")

	var partition string
	if p, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), region); ok {
		partition = p.ID()
	}

	var accountID string
	out, err := ec2Svc.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		MaxResults: aws.Int64(5),
	})
	if err != nil {
		log.Errorf("failed to load account if from security groups: %s", err)
		return
	} else if len(out.SecurityGroups) > 0 {
		accountID = aws.StringValue(out.SecurityGroups[0].OwnerId)
	}

	var loadBalancers []*elb.LoadBalancerDescription
	done := false
	params := &elb.DescribeLoadBalancersInput{}
	for !done {
		out, err := elbSvc.DescribeLoadBalancers(params)

		if err != nil {
			log.Errorf("failed to describe load balancers: %s", err)
			return
		}

		loadBalancers = append(loadBalancers, out.LoadBalancerDescriptions...)

		if out.NextMarker == nil {
			done = true
		} else {
			params.Marker = out.NextMarker
		}
	}

	log.Info("processing data")

	for _, l := range loadBalancers {
		var public bool
		if aws.StringValue(l.Scheme) == "internet-facing" {
			public = true
		} else if aws.StringValue(l.Scheme) == "internal" {
			public = false
		}

		d.rows <- inventory.Row{
			UniqueAssetIdentifier: aws.StringValue(l.LoadBalancerName),
			Virtual:               true,
			Public:                public,
			DNSNameOrURL:          aws.StringValue(l.DNSName),
			Location:              region,
			AssetType:             AssetTypeELB,
			Function:              aws.StringValue(l.CanonicalHostedZoneName),
			SerialAssetTagNumber:  fmt.Sprintf("arn:%s:elasticloadbalancing:%s:%s:loadbalancer/%s", partition, region, accountID, aws.StringValue(l.LoadBalancerName)),
			VLANNetworkID:         aws.StringValue(l.VPCId),
		}
	}

	log.Info("finished processing data")
}
