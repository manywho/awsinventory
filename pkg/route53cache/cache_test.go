package route53cache_test

import (
	"testing"

	. "github.com/sudoinclabs/awsinventory/pkg/route53cache"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"

	"github.com/stretchr/testify/require"
)

var testDomains = []string{
	"public-dns.testdomain.com",
	"private-dns.testdomain.com",
	"public-ip.testdomain.com",
	"private-ip.testdomain.com",
}

var testRecords = []*route53.ResourceRecordSet{
	{
		Type: aws.String("CNAME"),
		Name: aws.String(testDomains[0]),
		ResourceRecords: []*route53.ResourceRecord{
			{
				Value: testInstance.PublicDnsName,
			},
		},
	},
	{
		Type: aws.String("CNAME"),
		Name: aws.String(testDomains[1]),
		ResourceRecords: []*route53.ResourceRecord{
			{
				Value: testInstance.PrivateDnsName,
			},
		},
	},
	{
		Type: aws.String("A"),
		Name: aws.String(testDomains[2]),
		ResourceRecords: []*route53.ResourceRecord{
			{
				Value: testInstance.PublicIpAddress,
			},
		},
	},
	{
		Type: aws.String("A"),
		Name: aws.String(testDomains[3]),
		ResourceRecords: []*route53.ResourceRecord{
			{
				Value: testInstance.PrivateIpAddress,
			},
		},
	},
	{
		Type: aws.String("A"),
		Name: aws.String("should-not-be-here.com"),
		ResourceRecords: []*route53.ResourceRecord{
			{
				Value: aws.String("10.20.30.40"),
			},
		},
	},
}

var testInstance = &ec2.Instance{
	PublicDnsName:    aws.String("instance-1-public.ec2.aws.amazon.com"),
	PrivateDnsName:   aws.String("instance-1-private.ec2.aws.amazon.com"),
	PublicIpAddress:  aws.String("203.0.113.10"),
	PrivateIpAddress: aws.String("10.1.2.3"),
}

func TestSearchReturnsCorrectDomains(t *testing.T) {
	cache := New(testRecords)

	actual := cache.FindRecordsForInstance(testInstance)

	require.Equal(t, testDomains, actual)
}
