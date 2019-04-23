package inventory

import (
	"encoding/csv"
	"io"
)

var csvHeaders = []string{
	"Asset Identifier",
	"Asset Type",
	"Location",
	"Creation Date",
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
	return c.writer.Write(r.StringSlice())
}

// Flush flushes the buffer to the writer
func (c CSV) Flush() {
	c.writer.Flush()
}
