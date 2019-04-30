package data_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/manywho/awsinventory/internal/data"
	"github.com/manywho/awsinventory/internal/inventory"
)

var testELBRows = []inventory.Row{
	{
		ID:           "abcdefgh12345678",
		AssetType:    "ELB",
		Location:     ValidRegions[0],
		CreationDate: time.Now().AddDate(0, 0, -1),
		Application:  "mydomain.com",
		DNSName:      "abcdefgh12345678.ValidRegions[0].elb.amazonaws.com",
		VPCID:        "vpc-abcdefgh",
	},
	{
		ID:           "12345678abcdefgh",
		AssetType:    "ELB",
		Location:     ValidRegions[0],
		CreationDate: time.Now().AddDate(0, 0, -2),
		Application:  "another.com",
		DNSName:      "12345678abcdefgh.ValidRegions[0].elb.amazonaws.com",
		VPCID:        "vpc-12345678",
	},
	{
		ID:           "a1b2c3d4e5f6g7h8",
		AssetType:    "ELB",
		Location:     ValidRegions[0],
		CreationDate: time.Now().AddDate(0, 0, -1),
		Application:  "yetanother.com",
		DNSName:      "a1b2c3d4e5f6g7h8.ValidRegions[0].elb.amazonaws.com",
		VPCID:        "vpc-a1b2c3d4",
	},
}

// Test Data
var testELBOutput = &elb.DescribeLoadBalancersOutput{
	LoadBalancerDescriptions: []*elb.LoadBalancerDescription{
		{
			LoadBalancerName:        aws.String(testELBRows[0].ID),
			CreatedTime:             aws.Time(testELBRows[0].CreationDate),
			CanonicalHostedZoneName: aws.String(testELBRows[0].Application),
			DNSName:                 aws.String(testELBRows[0].DNSName),
			VPCId:                   aws.String(testELBRows[0].VPCID),
		},
		{
			LoadBalancerName:        aws.String(testELBRows[1].ID),
			CreatedTime:             aws.Time(testELBRows[1].CreationDate),
			CanonicalHostedZoneName: aws.String(testELBRows[1].Application),
			DNSName:                 aws.String(testELBRows[1].DNSName),
			VPCId:                   aws.String(testELBRows[1].VPCID),
		},
		{
			LoadBalancerName:        aws.String(testELBRows[2].ID),
			CreatedTime:             aws.Time(testELBRows[2].CreationDate),
			CanonicalHostedZoneName: aws.String(testELBRows[2].Application),
			DNSName:                 aws.String(testELBRows[2].DNSName),
			VPCId:                   aws.String(testELBRows[2].VPCID),
		},
	},
}

// Mocks
type ELBMock struct {
	elbiface.ELBAPI
}

func (e ELBMock) DescribeLoadBalancers(cfg *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {
	return testELBOutput, nil
}

type ELBErrorMock struct {
	elbiface.ELBAPI
}

func (e ELBErrorMock) DescribeLoadBalancers(cfg *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {
	return &elb.DescribeLoadBalancersOutput{}, testError
}

// Tests
func TestCanLoadELBs(t *testing.T) {
	d := New(logrus.New(), TestClients{ELB: ELBMock{}})

	d.Load([]string{ValidRegions[0]}, []string{ServiceELB})

	var count int
	d.MapRows(func(row inventory.Row) error {
		require.Equal(t, testELBRows[count], row)
		count++
		return nil
	})
	require.Equal(t, 3, count)
}

func TestLoadELBsLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{ELB: ELBErrorMock{}})

	d.Load([]string{ValidRegions[0]}, []string{ServiceELB})

	require.Contains(t, hook.LastEntry().Message, testError.Error())
}
