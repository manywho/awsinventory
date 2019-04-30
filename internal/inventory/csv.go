package inventory

import (
	"encoding/csv"
	"io"
)

var csvHeaders = []string{
	"Unique Asset Identifier",
	"IPv4 or IPv6 Address",
	"Virtual",
	"Public",
	"DNS Name or URL",
	"NetBIOS Name",
	"MAC Address",
	"Authenticated Scan",
	"Baseline Configuration Name",
	"OS Name and Version",
	"Location",
	"Asset Type",
	"Hardware Make/Model",
	"In Latest Scan",
	"Software/Database Vendor",
	"Patch Level",
	"Function",
	"Comments",
	"Serial #/Asset Tag #",
	"VLAN/Network ID",
	"System Administrator/Owner",
	"ApplicationAdministrator/Owner",
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
	return c.writer.Write(r.StringSlice())
}

// Flush flushes the buffer to the writer
func (c CSV) Flush() {
	c.writer.Flush()
}
