package awsdata

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/manywho/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeRDSInstance is the value used in the AssetType field when fetching RDS instances
	AssetTypeWorkSpaceInstance string = "Workspace"

	// ServiceRDS is the key for the RDS service
	ServiceWorkSpace string = "workspaces"
)

func (d *AWSData) loadWorkSpacesInstances(region string) {
	defer d.wg.Done()

	workspaceSvc := d.clients.GetWorkSpaceClient(region)

	log := d.log.WithFields(logrus.Fields{
		"region":  region,
		"service": ServiceWorkSpace,
	})

	log.Info("loading data")

	var workspacesItems []*workspaces.Workspace
	done := false
	params := &workspaces.DescribeWorkspacesInput{}
	for !done {
		out, err := workspaceSvc.DescribeWorkspaces(params)

		if err != nil {
			log.Errorf("failed to workspacess: %s", err)
			return
		}

		workspacesItems = append(workspacesItems, out.Workspaces...)

		if out.NextToken == nil {
			done = true
		} else {
			params.NextToken = out.NextToken
		}
	}

	log.Info("processing data")

	for _, i := range workspacesItems {
		d.rows <- inventory.Row{
			UniqueAssetIdentifier:          aws.StringValue(i.WorkspaceId),
			IPv4orIPv6Address:              aws.StringValue(i.IpAddress),
			Virtual:                        true,
			Public:                         true,
			NetBIOSName:                    aws.StringValue(i.ComputerName),
			BaselineConfigurationName:      aws.StringValue(i.BundleId),
			Location:                       region,
			AssetType:                      AssetTypeWorkSpaceInstance,
			HardwareMakeModel:              aws.StringValue(i.WorkspaceProperties.ComputeTypeName),
			Comments:                       fmt.Sprintf("UserName: %s, DirectoryID: %s, RunningMode %s, RootVolumeSize: %dGB, UserVolumeSizeGib: %dGB", aws.StringValue(i.UserName),aws.StringValue(i.DirectoryId),aws.StringValue(i.WorkspaceProperties.RunningMode),aws.Int64Value(i.WorkspaceProperties.RootVolumeSizeGib),aws.Int64Value(i.WorkspaceProperties.UserVolumeSizeGib)),
			SerialAssetTagNumber:           aws.StringValue(i.WorkspaceId),
			VLANNetworkID:                  aws.StringValue(i.SubnetId),
		}
	}

	log.Info("finished processing data")
}
