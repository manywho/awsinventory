package loader_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"

	"github.com/itmecho/awsinventory/internal/inventory"
	. "github.com/itmecho/awsinventory/internal/loader"
)

var testELBRows = []inventory.Row{
	inventory.Row{
		ID:           "abcdefgh12345678",
		AssetType:    "ELB",
		Location:     "test-region",
		CreationDate: time.Now().AddDate(0, 0, -1),
		Application:  "mydomain.com",
		DNSName:      "abcdefgh12345678.test-region.elb.amazonaws.com",
		VPCID:        "vpc-abcdefgh",
	},
	inventory.Row{
		ID:           "12345678abcdefgh",
		AssetType:    "ELB",
		Location:     "test-region",
		CreationDate: time.Now().AddDate(0, 0, -2),
		Application:  "another.com",
		DNSName:      "12345678abcdefgh.test-region.elb.amazonaws.com",
		VPCID:        "vpc-12345678",
	},
	inventory.Row{
		ID:           "a1b2c3d4e5f6g7h8",
		AssetType:    "ELB",
		Location:     "test-region",
		CreationDate: time.Now().AddDate(0, 0, -1),
		Application:  "yetanother.com",
		DNSName:      "a1b2c3d4e5f6g7h8.test-region.elb.amazonaws.com",
		VPCID:        "vpc-a1b2c3d4",
	},
}

type ELBMock struct {
	elbiface.ELBAPI
}

func (e ELBMock) DescribeLoadBalancers(cfg *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {
	return &elb.DescribeLoadBalancersOutput{
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
	}, nil
}

func TestLoadELBs(t *testing.T) {
	l := NewLoader()

	l.LoadELBs(ELBMock{}, testELBRows[0].Location)

	require.Len(t, l.Data, 3, "got more than 3 instances")
	require.Equal(t, testELBRows, l.Data, "didn't get expected data")
}

type ELBErrorMock struct {
	elbiface.ELBAPI
}

func (e ELBErrorMock) DescribeLoadBalancers(cfg *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {
	return &elb.DescribeLoadBalancersOutput{}, testError
}

func TestLoadELBsSendsErrorToChan(t *testing.T) {
	l := NewLoader()

	l.LoadELBs(ELBErrorMock{}, testELBRows[0].Location)

	require.Len(t, l.Errors, 1, "didn't send error to Errors channel")

	select {
	case e := <-l.Errors:
		require.Equal(t, testError, e, "didn't get expected error")
	default:
		t.Fatal("should have received an error")
	}
}
