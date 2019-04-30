package inventory

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var testRow = Row{
	UniqueAssetIdentifier:          "UniqueAssetIdentifier",
	IPv4orIPv6Address:              "IPv4orIPv6Address",
	Virtual:                        true,
	Public:                         true,
	DNSNameOrURL:                   "DNSNameOrURL",
	NetBIOSName:                    "NetBIOSName",
	MACAddress:                     "MACAddress",
	AuthenticatedScan:              false,
	BaselineConfigurationName:      "BaselineConfigurationName",
	OSNameandVersion:               "OSNameandVersion",
	Location:                       "Location",
	AssetType:                      "AssetType",
	HardwareMakeModel:              "HardwareMakeModel",
	InLatestScan:                   true,
	SoftwareDatabaseVendor:         "SoftwareDatabaseVendor",
	SoftwareDatabaseNameAndVersion: "SoftwareDatabaseNameAndVersion",
	PatchLevel:                     "PatchLevel",
	Function:                       "Function",
	Comments:                       "Comments",
	SerialAssetTagNumber:           "SerialAssetTagNumber",
	VLANNetworkID:                  "VLANNetworkID",
	SystemAdministratorOwner:       "SystemAdministratorOwner",
	ApplicationAdministratorOwner:  "ApplicationAdministratorOwner",
}

func TestRowCanReturnSliceOfStrings(t *testing.T) {
	actual := testRow.StringSlice()

	expected := []string{
		testRow.UniqueAssetIdentifier,
		testRow.IPv4orIPv6Address,
		"Yes",
		"Yes",
		testRow.DNSNameOrURL,
		testRow.NetBIOSName,
		testRow.MACAddress,
		"No",
		testRow.BaselineConfigurationName,
		testRow.OSNameandVersion,
		testRow.Location,
		testRow.AssetType,
		testRow.HardwareMakeModel,
		"Yes",
		testRow.SoftwareDatabaseVendor,
		testRow.SoftwareDatabaseNameAndVersion,
		testRow.PatchLevel,
		testRow.Function,
		testRow.Comments,
		testRow.SerialAssetTagNumber,
		testRow.VLANNetworkID,
		testRow.SystemAdministratorOwner,
		testRow.ApplicationAdministratorOwner,
	}

	require.Equal(t, expected, actual)
}
