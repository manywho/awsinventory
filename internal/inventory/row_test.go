package inventory

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var testRow = Row{
	ID:            "ID",
	AssetType:     "AssetType",
	Location:      "Location",
	Application:   "Application",
	Hardware:      "Hardware",
	Baseline:      "Baseline",
	OSNameVersion: "OSNameVersion",
	InternalIP:    "InternalIP",
	ExternalIP:    "ExternalIP",
	VPCID:         "VPCID",
	DNSName:       "DNSName",
}

func TestRowCanReturnSliceOfStrings(t *testing.T) {
	actual := testRow.StringSlice()

	expected := []string{
		testRow.ID,
		testRow.AssetType,
		testRow.Location,
		testRow.Application,
		testRow.Hardware,
		testRow.Baseline,
		testRow.OSNameVersion,
		testRow.InternalIP,
		testRow.ExternalIP,
		testRow.VPCID,
		testRow.DNSName,
	}

	require.Equal(t, expected, actual)
}
