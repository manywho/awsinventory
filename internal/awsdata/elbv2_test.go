package awsdata_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/manywho/awsinventory/internal/awsdata"
	"github.com/manywho/awsinventory/internal/inventory"
)

var testELBV2Rows = []inventory.Row{
	{
		UniqueAssetIdentifier: "abcdefgh12345678",
		Virtual:               true,
		Public:                true,
		DNSNameOrURL:          "abcdefgh12345678.ValidRegions[0].elb.amazonaws.com",
		Location:              ValidRegions[0],
		AssetType:             AssetTypeALB,
		VLANNetworkID:         "vpc-abcdefgh",
	},
	{
		UniqueAssetIdentifier: "12345678abcdefgh",
		Virtual:               true,
		Public:                true,
		DNSNameOrURL:          "12345678abcdefgh.ValidRegions[0].elb.amazonaws.com",
		Location:              ValidRegions[0],
		AssetType:             AssetTypeNLB,
		VLANNetworkID:         "vpc-12345678",
	},
	{
		UniqueAssetIdentifier: "a1b2c3d4e5f6g7h8",
		Virtual:               true,
		Public:                false,
		DNSNameOrURL:          "a1b2c3d4e5f6g7h8.ValidRegions[0].elb.amazonaws.com",
		Location:              ValidRegions[0],
		AssetType:             AssetTypeALB,
		VLANNetworkID:         "vpc-a1b2c3d4",
	},
}

// Test Data
var testELBV2Output = &elbv2.DescribeLoadBalancersOutput{
	LoadBalancers: []*elbv2.LoadBalancer{
		{
			LoadBalancerName:        aws.String(testELBV2Rows[0].UniqueAssetIdentifier),
			DNSName:                 aws.String(testELBV2Rows[0].DNSNameOrURL),
			Type:                    aws.String("application"),
			Scheme:                  aws.String("internet-facing"),
			VpcId:                   aws.String(testELBV2Rows[0].VLANNetworkID),
		},
		{
			LoadBalancerName:        aws.String(testELBV2Rows[1].UniqueAssetIdentifier),
			DNSName:                 aws.String(testELBV2Rows[1].DNSNameOrURL),
			Type:                    aws.String("network"),
			Scheme:                  aws.String("internet-facing"),
			VpcId:                   aws.String(testELBV2Rows[1].VLANNetworkID),
		},
		{
			LoadBalancerName:        aws.String(testELBV2Rows[2].UniqueAssetIdentifier),
			DNSName:                 aws.String(testELBV2Rows[2].DNSNameOrURL),
			Type:                    aws.String("application"),
			Scheme:                  aws.String("internal"),
			VpcId:                   aws.String(testELBV2Rows[2].VLANNetworkID),
		},
	},
}

// Mocks
type ELBV2Mock struct {
	elbv2iface.ELBV2API
}

func (e ELBV2Mock) DescribeLoadBalancers(cfg *elbv2.DescribeLoadBalancersInput) (*elbv2.DescribeLoadBalancersOutput, error) {
	return testELBV2Output, nil
}

type ELBV2ErrorMock struct {
	elbv2iface.ELBV2API
}

func (e ELBV2ErrorMock) DescribeLoadBalancers(cfg *elbv2.DescribeLoadBalancersInput) (*elbv2.DescribeLoadBalancersOutput, error) {
	return &elbv2.DescribeLoadBalancersOutput{}, testError
}

// Tests
func TestCanLoadELBV2s(t *testing.T) {
	d := New(logrus.New(), TestClients{ELBV2: ELBV2Mock{}})

	d.Load([]string{ValidRegions[0]}, []string{ServiceELBV2})

	var count int
	d.MapRows(func(row inventory.Row) error {
		require.Equal(t, testELBV2Rows[count], row)
		count++
		return nil
	})
	require.Equal(t, 3, count)
}

func TestLoadELBV2sLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{ELBV2: ELBV2ErrorMock{}})

	d.Load([]string{ValidRegions[0]}, []string{ServiceELBV2})

	require.Contains(t, hook.LastEntry().Message, testError.Error())
}
