package awsdata_test

import (
	"sort"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/sudoinclabs/awsinventory/internal/awsdata"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
)

var testKMSKeyRows = []inventory.Row{
	{
		UniqueAssetIdentifier:     "1234abcd-12ab-34cd-56ef-1234567890ab",
		Virtual:                   true,
		Public:                    false,
		BaselineConfigurationName: "AWS_KMS",
		Location:                  DefaultRegion,
		AssetType:                 "KMS Key",
		Comments:                  "AWS, SYMMETRIC_DEFAULT\nCreated at: 2019-10-10T20:00:00Z",
		SerialAssetTagNumber:      "arn:aws:kms:us-east-1:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab",
		Function:                  "Test key 1",
	},
	{
		UniqueAssetIdentifier:     "b735f38f-81b2-462b-899a-a281ae69b844",
		Virtual:                   true,
		Public:                    false,
		BaselineConfigurationName: "AWS_KMS",
		Location:                  DefaultRegion,
		AssetType:                 "KMS Key",
		Comments:                  "CUSTOMER, SYMMETRIC_DEFAULT\nCreated at: 2020-01-02T12:00:00Z",
		SerialAssetTagNumber:      "arn:aws:kms:us-east-1:123456789012:key/b735f38f-81b2-462b-899a-a281ae69b844",
		Function:                  "Test key 2",
	},
	{
		UniqueAssetIdentifier:     "e602e5cb-e5c4-4ed3-aaad-980db0ec8fdd",
		Virtual:                   true,
		Public:                    false,
		BaselineConfigurationName: "EXTERNAL",
		Location:                  DefaultRegion,
		AssetType:                 "KMS Key",
		Comments:                  "CUSTOMER, RSA_4096\nCreated at: 2020-09-01T04:00:00Z\nValid to: 2021-08-31T23:00:00Z",
		SerialAssetTagNumber:      "arn:aws:kms:us-east-1:123456789012:key/e602e5cb-e5c4-4ed3-aaad-980db0ec8fdd",
		Function:                  "Test key 3",
	},
}

// Test Data
var testKMSListKeysOutputPage1 = &kms.ListKeysOutput{
	Keys: []*kms.KeyListEntry{
		{
			KeyId: aws.String(testKMSKeyRows[0].UniqueAssetIdentifier),
		},
		{
			KeyId: aws.String(testKMSKeyRows[1].UniqueAssetIdentifier),
		},
	},
	NextMarker: aws.String(testKMSKeyRows[1].UniqueAssetIdentifier),
	Truncated:  aws.Bool(true),
}

var testKMSListKeysOutputPage2 = &kms.ListKeysOutput{
	Keys: []*kms.KeyListEntry{
		{
			KeyId: aws.String(testKMSKeyRows[2].UniqueAssetIdentifier),
		},
	},
}

// Mocks
type KMSMock struct {
	kmsiface.KMSAPI
}

func (e KMSMock) ListKeys(cfg *kms.ListKeysInput) (*kms.ListKeysOutput, error) {
	if cfg.Marker == nil {
		return testKMSListKeysOutputPage1, nil
	}

	return testKMSListKeysOutputPage2, nil
}

func (e KMSMock) DescribeKey(cfg *kms.DescribeKeyInput) (*kms.DescribeKeyOutput, error) {
	var customerMasterKeySpec, keyManager string
	var creationDate, validTo time.Time
	var row int
	switch aws.StringValue(cfg.KeyId) {
	case testKMSKeyRows[0].UniqueAssetIdentifier:
		row = 0
		creationDate = time.Date(2019, time.October, 10, 20, 0, 0, 0, time.UTC)
		customerMasterKeySpec = "SYMMETRIC_DEFAULT"
		keyManager = "AWS"
	case testKMSKeyRows[1].UniqueAssetIdentifier:
		row = 1
		creationDate = time.Date(2020, time.January, 2, 12, 0, 0, 0, time.UTC)
		customerMasterKeySpec = "SYMMETRIC_DEFAULT"
		keyManager = "CUSTOMER"
	case testKMSKeyRows[2].UniqueAssetIdentifier:
		row = 2
		creationDate = time.Date(2020, time.September, 1, 4, 0, 0, 0, time.UTC)
		customerMasterKeySpec = "RSA_4096"
		keyManager = "CUSTOMER"
		validTo = time.Date(2021, time.August, 31, 23, 0, 0, 0, time.UTC)
	}
	return &kms.DescribeKeyOutput{
		KeyMetadata: &kms.KeyMetadata{
			Arn:                   aws.String(testKMSKeyRows[row].SerialAssetTagNumber),
			CreationDate:          aws.Time(creationDate),
			CustomerMasterKeySpec: aws.String(customerMasterKeySpec),
			Description:           aws.String(testKMSKeyRows[row].Function),
			KeyId:                 aws.String(testKMSKeyRows[row].UniqueAssetIdentifier),
			KeyManager:            aws.String(keyManager),
			Origin:                aws.String(testKMSKeyRows[row].BaselineConfigurationName),
			ValidTo:               aws.Time(validTo),
		},
	}, nil
}

type KMSErrorMock struct {
	kmsiface.KMSAPI
}

func (e KMSErrorMock) ListKeys(cfg *kms.ListKeysInput) (*kms.ListKeysOutput, error) {
	return &kms.ListKeysOutput{}, testError
}

func (e KMSErrorMock) DescribeKey(cfg *kms.DescribeKeyInput) (*kms.DescribeKeyOutput, error) {
	return &kms.DescribeKeyOutput{}, testError
}

// Tests
func TestCanLoadKMSKeys(t *testing.T) {
	d := New(logrus.New(), TestClients{KMS: KMSMock{}})

	var rows []inventory.Row
	d.Load([]string{DefaultRegion}, []string{ServiceKMS}, func(row inventory.Row) error {
		rows = append(rows, row)
		return nil
	})

	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].UniqueAssetIdentifier < rows[j].UniqueAssetIdentifier
	})

	require.Equal(t, 3, len(rows))

	for i := range rows {
		require.Equal(t, testKMSKeyRows[i], rows[i])
	}
}

func TestLoadKMSKeysLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{KMS: KMSErrorMock{}})

	d.Load([]string{DefaultRegion}, []string{ServiceKMS}, nil)

	assertTestErrorWasLogged(t, hook.Entries)
}
