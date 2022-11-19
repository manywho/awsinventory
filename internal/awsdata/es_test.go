package awsdata_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/aws/aws-sdk-go/service/elasticsearchservice/elasticsearchserviceiface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/sudoinclabs/awsinventory/internal/awsdata"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
)

var testElasticsearchDomainRows = []inventory.Row{
	{
		UniqueAssetIdentifier:          "test-domain-1",
		Virtual:                        true,
		Public:                         false,
		DNSNameOrURL:                   "vpc-test-domain-1-abcdefghijk.us-east-2.es.amazonaws.com",
		Location:                       DefaultRegion,
		AssetType:                      "Elasticsearch Domain",
		HardwareMakeModel:              "c4.medium.elasticsearch",
		SoftwareDatabaseVendor:         "Elastic",
		SoftwareDatabaseNameAndVersion: "Elasticsearch 7.1",
		SerialAssetTagNumber:           "arn:aws:es:us-east-2:123456789012:domain/test-domain-1",
		VLANNetworkID:                  "vpc-12345678",
	},
	{
		UniqueAssetIdentifier:          "test-domain-2",
		Virtual:                        true,
		Public:                         false,
		DNSNameOrURL:                   "vpc-test-domain-2-jkhsdjhfghjsfghj.us-east-2.es.amazonaws.com",
		Location:                       DefaultRegion,
		AssetType:                      "Elasticsearch Domain",
		HardwareMakeModel:              "c4.large.elasticsearch",
		SoftwareDatabaseVendor:         "Elastic",
		SoftwareDatabaseNameAndVersion: "Elasticsearch 7.7",
		SerialAssetTagNumber:           "arn:aws:es:us-east-2:123456789012:domain/test-domain-2",
		VLANNetworkID:                  "vpc-abcdefgh",
	},
	{
		UniqueAssetIdentifier:          "test-domain-3",
		Virtual:                        true,
		Public:                         false,
		DNSNameOrURL:                   "vpc-test-domain-3-rtuiewruincdfgs.us-east-2.es.amazonaws.com",
		Location:                       DefaultRegion,
		AssetType:                      "Elasticsearch Domain",
		HardwareMakeModel:              "m4.large.elasticsearch",
		SoftwareDatabaseVendor:         "Elastic",
		SoftwareDatabaseNameAndVersion: "Elasticsearch 6.2",
		SerialAssetTagNumber:           "arn:aws:es:us-east-2:123456789012:domain/test-domain-3",
		VLANNetworkID:                  "vpc-a1b2c3d4",
	},
}

// Test Data
var testElasticsearchListDomainNamesOutput = &elasticsearchservice.ListDomainNamesOutput{
	DomainNames: []*elasticsearchservice.DomainInfo{
		{
			DomainName: aws.String(testElasticsearchDomainRows[0].UniqueAssetIdentifier),
		},
		{
			DomainName: aws.String(testElasticsearchDomainRows[1].UniqueAssetIdentifier),
		},
		{
			DomainName: aws.String(testElasticsearchDomainRows[2].UniqueAssetIdentifier),
		},
	},
}

var testElasticsearchDescribeElasticsearchDomainsOutput = &elasticsearchservice.DescribeElasticsearchDomainsOutput{
	DomainStatusList: []*elasticsearchservice.ElasticsearchDomainStatus{
		{
			DomainName: aws.String(testElasticsearchDomainRows[0].UniqueAssetIdentifier),
			ARN:        aws.String(testElasticsearchDomainRows[0].SerialAssetTagNumber),
			Endpoints: map[string]*string{
				"vpc": aws.String(testElasticsearchDomainRows[0].DNSNameOrURL),
			},
			ElasticsearchVersion: aws.String("7.1"),
			ElasticsearchClusterConfig: &elasticsearchservice.ElasticsearchClusterConfig{
				InstanceType: aws.String(testElasticsearchDomainRows[0].HardwareMakeModel),
			},
			VPCOptions: &elasticsearchservice.VPCDerivedInfo{
				VPCId: aws.String(testElasticsearchDomainRows[0].VLANNetworkID),
			},
		},
		{
			DomainName: aws.String(testElasticsearchDomainRows[1].UniqueAssetIdentifier),
			ARN:        aws.String(testElasticsearchDomainRows[1].SerialAssetTagNumber),
			Endpoints: map[string]*string{
				"vpc": aws.String(testElasticsearchDomainRows[1].DNSNameOrURL),
			},
			ElasticsearchVersion: aws.String("7.7"),
			ElasticsearchClusterConfig: &elasticsearchservice.ElasticsearchClusterConfig{
				InstanceType: aws.String(testElasticsearchDomainRows[1].HardwareMakeModel),
			},
			VPCOptions: &elasticsearchservice.VPCDerivedInfo{
				VPCId: aws.String(testElasticsearchDomainRows[1].VLANNetworkID),
			},
		},
		{
			DomainName: aws.String(testElasticsearchDomainRows[2].UniqueAssetIdentifier),
			ARN:        aws.String(testElasticsearchDomainRows[2].SerialAssetTagNumber),
			Endpoints: map[string]*string{
				"vpc": aws.String(testElasticsearchDomainRows[2].DNSNameOrURL),
			},
			ElasticsearchVersion: aws.String("6.2"),
			ElasticsearchClusterConfig: &elasticsearchservice.ElasticsearchClusterConfig{
				InstanceType: aws.String(testElasticsearchDomainRows[2].HardwareMakeModel),
			},
			VPCOptions: &elasticsearchservice.VPCDerivedInfo{
				VPCId: aws.String(testElasticsearchDomainRows[2].VLANNetworkID),
			},
		},
	},
}

// Mocks
type ElasticsearchServiceMock struct {
	elasticsearchserviceiface.ElasticsearchServiceAPI
}

func (e ElasticsearchServiceMock) ListDomainNames(cfg *elasticsearchservice.ListDomainNamesInput) (*elasticsearchservice.ListDomainNamesOutput, error) {
	return testElasticsearchListDomainNamesOutput, nil
}

func (e ElasticsearchServiceMock) DescribeElasticsearchDomains(cfg *elasticsearchservice.DescribeElasticsearchDomainsInput) (*elasticsearchservice.DescribeElasticsearchDomainsOutput, error) {
	return testElasticsearchDescribeElasticsearchDomainsOutput, nil
}

type ElasticsearchServiceErrorMock struct {
	elasticsearchserviceiface.ElasticsearchServiceAPI
}

func (e ElasticsearchServiceErrorMock) ListDomainNames(cfg *elasticsearchservice.ListDomainNamesInput) (*elasticsearchservice.ListDomainNamesOutput, error) {
	return &elasticsearchservice.ListDomainNamesOutput{}, testError
}

func (e ElasticsearchServiceErrorMock) DescribeElasticsearchDomains(cfg *elasticsearchservice.DescribeElasticsearchDomainsInput) (*elasticsearchservice.DescribeElasticsearchDomainsOutput, error) {
	return &elasticsearchservice.DescribeElasticsearchDomainsOutput{}, testError
}

// Tests
func TestCanLoadElasticsearchDomains(t *testing.T) {
	d := New(logrus.New(), TestClients{ElasticsearchService: ElasticsearchServiceMock{}})

	var count int
	d.Load([]string{DefaultRegion}, []string{ServiceElasticsearchService}, func(row inventory.Row) error {
		require.Equal(t, testElasticsearchDomainRows[count], row)
		count++
		return nil
	})
	require.Equal(t, 3, count)
}

func TestLoadElasticsearchDomainsLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{ElasticsearchService: ElasticsearchServiceErrorMock{}})

	d.Load([]string{DefaultRegion}, []string{ServiceElasticsearchService}, nil)

	assertTestErrorWasLogged(t, hook.Entries)
}
