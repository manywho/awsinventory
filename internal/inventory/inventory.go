package inventory

// Row represents a row in the report
type Row struct {
	ID            string
	AssetType     string
	Location      string
	Application   string
	Hardware      string
	Baseline      string
	OSNameVersion string
	InternalIP    string
	ExternalIP    string
	VPCID         string
	DNSName       string
}
