package awsdata_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/sudoinclabs/awsinventory/internal/awsdata"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
)

var testELBRows = []inventory.Row{
	{
		UniqueAssetIdentifier: "abcdefgh12345678",
		Virtual:               true,
		Public:                true,
		DNSNameOrURL:          "abcdefgh12345678.us-east-1.elb.amazonaws.com",
		Location:              DefaultRegion,
		AssetType:             AssetTypeELB,
		Function:              "mydomain.com",
		SerialAssetTagNumber:  "arn:aws:elasticloadbalancing:us-east-1:012345678910:loadbalancer/abcdefgh12345678",
		VLANNetworkID:         "vpc-abcdefgh",
	},
	{
		UniqueAssetIdentifier: "12345678abcdefgh",
		Virtual:               true,
		Public:                true,
		DNSNameOrURL:          "12345678abcdefgh.us-east-1.elb.amazonaws.com",
		Location:              DefaultRegion,
		AssetType:             AssetTypeELB,
		Function:              "another.com",
		SerialAssetTagNumber:  "arn:aws:elasticloadbalancing:us-east-1:012345678910:loadbalancer/12345678abcdefgh",
		VLANNetworkID:         "vpc-12345678",
	},
	{
		UniqueAssetIdentifier: "a1b2c3d4e5f6g7h8",
		Virtual:               true,
		Public:                false,
		DNSNameOrURL:          "a1b2c3d4e5f6g7h8.us-east-1.elb.amazonaws.com",
		Location:              DefaultRegion,
		AssetType:             AssetTypeELB,
		Function:              "yetanother.com",
		SerialAssetTagNumber:  "arn:aws:elasticloadbalancing:us-east-1:012345678910:loadbalancer/a1b2c3d4e5f6g7h8",
		VLANNetworkID:         "vpc-a1b2c3d4",
	},
}

// Test Data
var testELBDescribeLoadBalancersOutputPage1 = &elb.DescribeLoadBalancersOutput{
	LoadBalancerDescriptions: []*elb.LoadBalancerDescription{
		{
			LoadBalancerName:        aws.String(testELBRows[0].UniqueAssetIdentifier),
			CanonicalHostedZoneName: aws.String(testELBRows[0].Function),
			DNSName:                 aws.String(testELBRows[0].DNSNameOrURL),
			Scheme:                  aws.String("internet-facing"),
			VPCId:                   aws.String(testELBRows[0].VLANNetworkID),
		},
		{
			LoadBalancerName:        aws.String(testELBRows[1].UniqueAssetIdentifier),
			CanonicalHostedZoneName: aws.String(testELBRows[1].Function),
			DNSName:                 aws.String(testELBRows[1].DNSNameOrURL),
			Scheme:                  aws.String("internet-facing"),
			VPCId:                   aws.String(testELBRows[1].VLANNetworkID),
		},
	},
	NextMarker: aws.String(testELBRows[1].UniqueAssetIdentifier),
}

var testELBDescribeLoadBalancersOutputPage2 = &elb.DescribeLoadBalancersOutput{
	LoadBalancerDescriptions: []*elb.LoadBalancerDescription{
		{
			LoadBalancerName:        aws.String(testELBRows[2].UniqueAssetIdentifier),
			CanonicalHostedZoneName: aws.String(testELBRows[2].Function),
			DNSName:                 aws.String(testELBRows[2].DNSNameOrURL),
			Scheme:                  aws.String("internal"),
			VPCId:                   aws.String(testELBRows[2].VLANNetworkID),
		},
	},
}

// Mocks
type ELBMock struct {
	elbiface.ELBAPI
}

func (e ELBMock) DescribeLoadBalancers(cfg *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {
	if cfg.Marker == nil {
		return testELBDescribeLoadBalancersOutputPage1, nil
	}

	return testELBDescribeLoadBalancersOutputPage2, nil
}

type ELBErrorMock struct {
	elbiface.ELBAPI
}

func (e ELBErrorMock) DescribeLoadBalancers(cfg *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {
	return &elb.DescribeLoadBalancersOutput{}, testError
}

// Tests
func TestCanLoadELBs(t *testing.T) {
	d := New(logrus.New(), TestClients{EC2: EC2Mock{}, ELB: ELBMock{}})

	var count int
	d.Load([]string{DefaultRegion}, []string{ServiceELB}, func(row inventory.Row) error {
		require.Equal(t, testELBRows[count], row)
		count++
		return nil
	})
	require.Equal(t, 3, count)
}

func TestLoadELBsLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{EC2: EC2ErrorMock{}, ELB: ELBErrorMock{}})

	d.Load([]string{DefaultRegion}, []string{ServiceELB}, nil)

	assertTestErrorWasLogged(t, hook.Entries)
}
