package awsdata

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeKMSKey is the value used in the AssetType field when fetching KMS keys
	AssetTypeKMSKey string = "KMS Key"

	// ServiceKMS is the key for the KMS service
	ServiceKMS string = "kms"
)

func (d *AWSData) loadKMSKeys(region string) {
	defer d.wg.Done()

	kmsSvc := d.clients.GetKMSClient(region)

	log := d.log.WithFields(logrus.Fields{
		"region":  region,
		"service": ServiceKMS,
	})

	log.Info("loading data")

	var keys []*kms.KeyListEntry
	done := false
	params := &kms.ListKeysInput{}
	for !done {
		out, err := kmsSvc.ListKeys(params)

		if err != nil {
			log.Errorf("failed to list keys: %s", err)
			return
		}

		keys = append(keys, out.Keys...)

		if aws.BoolValue(out.Truncated) {
			params.Marker = out.NextMarker
		} else {
			done = true
		}
	}

	log.Info("processing data")

	for _, k := range keys {
		d.wg.Add(1)
		go d.processKMSKey(log, kmsSvc, k, region)
	}

	log.Info("finished processing data")
}

func (d *AWSData) processKMSKey(log *logrus.Entry, kmsSvc kmsiface.KMSAPI, key *kms.KeyListEntry, region string) {
	defer d.wg.Done()

	out, err := kmsSvc.DescribeKey(&kms.DescribeKeyInput{
		KeyId: key.KeyId,
	})
	if err != nil {
		log.Errorf("failed to describe key %s: %s", aws.StringValue(key.KeyId), err)
		return
	}

	var comments []string
	comments = append(comments, fmt.Sprintf("%s, %s", aws.StringValue(out.KeyMetadata.KeyManager), aws.StringValue(out.KeyMetadata.CustomerMasterKeySpec)))
	comments = append(comments, "Created at: "+aws.TimeValue(out.KeyMetadata.CreationDate).Format(time.RFC3339))
	if out.KeyMetadata.ValidTo != nil && !out.KeyMetadata.ValidTo.IsZero() {
		comments = append(comments, "Valid to: "+aws.TimeValue(out.KeyMetadata.ValidTo).Format(time.RFC3339))
	}

	d.rows <- inventory.Row{
		UniqueAssetIdentifier:     aws.StringValue(out.KeyMetadata.KeyId),
		Virtual:                   true,
		Public:                    false,
		BaselineConfigurationName: aws.StringValue(out.KeyMetadata.Origin),
		Location:                  region,
		AssetType:                 AssetTypeKMSKey,
		Comments:                  strings.Join(comments, "\n"),
		SerialAssetTagNumber:      aws.StringValue(out.KeyMetadata.Arn),
		Function:                  aws.StringValue(out.KeyMetadata.Description),
	}
}
