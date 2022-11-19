package awsdata_test

import (
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/sudoinclabs/awsinventory/internal/awsdata"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
)

var testELBV2Rows = []inventory.Row{
	{
		UniqueAssetIdentifier: "12345678abcdefgh",
		IPv4orIPv6Address:     "1.2.3.4",
		Virtual:               true,
		Public:                true,
		DNSNameOrURL:          "12345678abcdefgh.us-east-1.elb.amazonaws.com",
		Location:              DefaultRegion,
		AssetType:             AssetTypeALB,
		SerialAssetTagNumber:  "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/12345678abcdefgh/50dc6c495c0c9188",
		VLANNetworkID:         "vpc-abcdefgh",
	},
	{
		UniqueAssetIdentifier: "a1b2c3d4e5f6g7h8",
		IPv4orIPv6Address:     "2.4.6.8\n2001:0db8:85a3:0000:0000:8a2e:0370:7334\n10.33.44.55",
		Virtual:               true,
		Public:                true,
		DNSNameOrURL:          "a1b2c3d4e5f6g7h8.us-east-1.elb.amazonaws.com",
		Location:              DefaultRegion,
		AssetType:             AssetTypeNLB,
		SerialAssetTagNumber:  "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/net/a1b2c3d4e5f6g7h8/b6dabd72b1aef5a2",
		VLANNetworkID:         "vpc-12345678",
	},
	{
		UniqueAssetIdentifier: "abcdefgh12345678",
		IPv4orIPv6Address:     "10.22.33.44",
		Virtual:               true,
		Public:                false,
		DNSNameOrURL:          "abcdefgh12345678.us-east-1.elb.amazonaws.com",
		Location:              DefaultRegion,
		AssetType:             AssetTypeALB,
		SerialAssetTagNumber:  "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/abcdefgh12345678s/573c812d8a4526b7",
		VLANNetworkID:         "vpc-a1b2c3d4",
	},
	{
		UniqueAssetIdentifier: "h9g8f7e6d5c4b3a2",
		IPv4orIPv6Address:     "10.99.88.77",
		Virtual:               true,
		Public:                false,
		DNSNameOrURL:          "h9g8f7e6d5c4b3a2.us-east-1.elb.amazonaws.com",
		Location:              DefaultRegion,
		AssetType:             AssetTypeGLB,
		SerialAssetTagNumber:  "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/gateway/h9g8f7e6d5c4b3a2/d9a04f08d67f6442",
		VLANNetworkID:         "vpc-e9f8g7h6",
	},
}

// Test Data
var testELBV2DescribeLoadBalancersOutputPage1 = &elbv2.DescribeLoadBalancersOutput{
	LoadBalancers: []*elbv2.LoadBalancer{
		{
			AvailabilityZones: []*elbv2.AvailabilityZone{
				{
					LoadBalancerAddresses: []*elbv2.LoadBalancerAddress{
						{
							IpAddress: aws.String("1.2.3.4"),
						},
					},
				},
			},
			LoadBalancerName: aws.String(testELBV2Rows[0].UniqueAssetIdentifier),
			LoadBalancerArn:  aws.String(testELBV2Rows[0].SerialAssetTagNumber),
			DNSName:          aws.String(testELBV2Rows[0].DNSNameOrURL),
			Type:             aws.String("application"),
			Scheme:           aws.String("internet-facing"),
			VpcId:            aws.String(testELBV2Rows[0].VLANNetworkID),
		},
		{
			AvailabilityZones: []*elbv2.AvailabilityZone{
				{
					LoadBalancerAddresses: []*elbv2.LoadBalancerAddress{
						{
							IpAddress:          aws.String("2.4.6.8"),
							IPv6Address:        aws.String("2001:0db8:85a3:0000:0000:8a2e:0370:7334"),
							PrivateIPv4Address: aws.String("10.33.44.55"),
						},
					},
				},
			},
			LoadBalancerName: aws.String(testELBV2Rows[1].UniqueAssetIdentifier),
			LoadBalancerArn:  aws.String(testELBV2Rows[1].SerialAssetTagNumber),
			DNSName:          aws.String(testELBV2Rows[1].DNSNameOrURL),
			Type:             aws.String("network"),
			Scheme:           aws.String("internet-facing"),
			VpcId:            aws.String(testELBV2Rows[1].VLANNetworkID),
		},
	},
	NextMarker: aws.String(testELBV2Rows[1].UniqueAssetIdentifier),
}

var testELBV2DescribeLoadBalancersOutputPage2 = &elbv2.DescribeLoadBalancersOutput{
	LoadBalancers: []*elbv2.LoadBalancer{
		{
			AvailabilityZones: []*elbv2.AvailabilityZone{
				{
					LoadBalancerAddresses: []*elbv2.LoadBalancerAddress{
						{
							IpAddress: aws.String("10.22.33.44"),
						},
					},
				},
			},
			LoadBalancerName: aws.String(testELBV2Rows[2].UniqueAssetIdentifier),
			LoadBalancerArn:  aws.String(testELBV2Rows[2].SerialAssetTagNumber),
			DNSName:          aws.String(testELBV2Rows[2].DNSNameOrURL),
			Type:             aws.String("application"),
			Scheme:           aws.String("internal"),
			VpcId:            aws.String(testELBV2Rows[2].VLANNetworkID),
		},
		{
			AvailabilityZones: []*elbv2.AvailabilityZone{
				{
					LoadBalancerAddresses: []*elbv2.LoadBalancerAddress{
						{
							IpAddress: aws.String("10.99.88.77"),
						},
					},
				},
			},
			LoadBalancerName: aws.String(testELBV2Rows[3].UniqueAssetIdentifier),
			LoadBalancerArn:  aws.String(testELBV2Rows[3].SerialAssetTagNumber),
			DNSName:          aws.String(testELBV2Rows[3].DNSNameOrURL),
			Type:             aws.String("gateway"),
			Scheme:           aws.String("internal"),
			VpcId:            aws.String(testELBV2Rows[3].VLANNetworkID),
		},
	},
}

// Mocks
type ELBV2Mock struct {
	elbv2iface.ELBV2API
}

func (e ELBV2Mock) DescribeLoadBalancers(cfg *elbv2.DescribeLoadBalancersInput) (*elbv2.DescribeLoadBalancersOutput, error) {
	if cfg.Marker == nil {
		return testELBV2DescribeLoadBalancersOutputPage1, nil
	}

	return testELBV2DescribeLoadBalancersOutputPage2, nil
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

	var rows []inventory.Row
	d.Load([]string{DefaultRegion}, []string{ServiceELBV2}, func(row inventory.Row) error {
		rows = append(rows, row)
		return nil
	})

	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].UniqueAssetIdentifier < rows[j].UniqueAssetIdentifier
	})

	require.Equal(t, 4, len(rows))

	for i, row := range rows {
		require.Equal(t, testELBV2Rows[i], row)
	}
}

func TestLoadELBV2sLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{ELBV2: ELBV2ErrorMock{}})

	d.Load([]string{DefaultRegion}, []string{ServiceELBV2}, nil)

	assertTestErrorWasLogged(t, hook.Entries)
}
