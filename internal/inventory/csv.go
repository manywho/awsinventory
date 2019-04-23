package inventory

import (
	"encoding/csv"
	"io"
)

var csvHeaders = []string{
	"Asset Identifier",
	"Asset Type",
	"Location",
	"Application",
	"Hardware",
	"Baseline",
	"OS Name/Version",
	"Internal IP",
	"External IP",
	"VPC ID",
	"DNS Name",
}

// CSV handles a csv format inventory
type CSV struct {
	writer *csv.Writer
}

// NewCSV returns a new csv object ready to have rows written to it
func NewCSV(writer io.Writer) (c CSV, err error) {
	c = CSV{
		writer: csv.NewWriter(writer),
	}

	err = c.writeHeaders()

	return
}

// writeHeaders writes the column headings to a csv writer
func (c CSV) writeHeaders() error {
	return c.writer.Write(csvHeaders)
}

// WriteRow writes the row to the csv writer
func (c CSV) WriteRow(r Row) error {
	var record []string

	record = append(record, r.ID)
	record = append(record, r.AssetType)
	record = append(record, r.Location)
	record = append(record, r.Application)
	record = append(record, r.Hardware)
	record = append(record, r.Baseline)
	record = append(record, r.OSNameVersion)
	record = append(record, r.InternalIP)
	record = append(record, r.ExternalIP)
	record = append(record, r.VPCID)
	record = append(record, r.DNSName)

	return c.writer.Write(record)
}

// Flush flushes the buffer to the writer
func (c CSV) Flush() {
	c.writer.Flush()
}
