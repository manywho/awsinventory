package inventory

import "time"

// Row represents a row in the report
type Row struct {
	ID            string
	AssetType     string
	Location      string
	CreationDate  time.Time
	Application   string
	Hardware      string
	Baseline      string
	OSNameVersion string
	InternalIP    string
	ExternalIP    string
	VPCID         string
	DNSName       string
}

// StringSlice returns a slice of strings representing the fields on the Row
func (r Row) StringSlice() []string {
	var record []string

	record = append(record, r.ID)
	record = append(record, r.AssetType)
	record = append(record, r.Location)
	record = append(record, r.CreationDate.String())
	record = append(record, r.Application)
	record = append(record, r.Hardware)
	record = append(record, r.Baseline)
	record = append(record, r.OSNameVersion)
	record = append(record, r.InternalIP)
	record = append(record, r.ExternalIP)
	record = append(record, r.VPCID)
	record = append(record, r.DNSName)

	return record
}
