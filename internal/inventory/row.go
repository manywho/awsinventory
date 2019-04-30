package inventory

// Row represents a row in the report
type Row struct {
	UniqueAssetIdentifier          string
	IPv4orIPv6Address              string
	Virtual                        bool
	Public                         bool
	DNSNameOrURL                   string
	NetBIOSName                    string
	MACAddress                     string
	AuthenticatedScan              bool
	BaselineConfigurationName      string
	OSNameandVersion               string
	Location                       string
	AssetType                      string
	HardwareMakeModel              string
	InLatestScan                   bool
	SoftwareDatabaseVendor         string
	SoftwareDatabaseNameAndVersion string
	PatchLevel                     string
	Function                       string
	Comments                       string
	SerialAssetTagNumber           string
	VLANNetworkID                  string
	SystemAdministratorOwner       string
	ApplicationAdministratorOwner  string
}

// StringSlice returns a slice of strings representing the fields on the Row
func (r Row) StringSlice() []string {
	var record []string

	record = append(record, r.UniqueAssetIdentifier)
	record = append(record, r.IPv4orIPv6Address)
	record = append(record, getBoolString(r.Virtual))
	record = append(record, getBoolString(r.Public))
	record = append(record, r.DNSNameOrURL)
	record = append(record, r.NetBIOSName)
	record = append(record, r.MACAddress)
	record = append(record, getBoolString(r.AuthenticatedScan))
	record = append(record, r.BaselineConfigurationName)
	record = append(record, r.OSNameandVersion)
	record = append(record, r.Location)
	record = append(record, r.AssetType)
	record = append(record, r.HardwareMakeModel)
	record = append(record, getBoolString(r.InLatestScan))
	record = append(record, r.SoftwareDatabaseVendor)
	record = append(record, r.SoftwareDatabaseNameAndVersion)
	record = append(record, r.PatchLevel)
	record = append(record, r.Function)
	record = append(record, r.Comments)
	record = append(record, r.SerialAssetTagNumber)
	record = append(record, r.VLANNetworkID)
	record = append(record, r.SystemAdministratorOwner)
	record = append(record, r.ApplicationAdministratorOwner)

	return record
}

func getBoolString(b bool) string {
	if b {
		return "Yes"
	}

	return "No"
}
