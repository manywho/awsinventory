package main

import (
	"os"

	"github.com/manywho/awsinventory/internal/data"

	"github.com/manywho/awsinventory/internal/inventory"
	"github.com/spf13/pflag"
)

var (
	outputFile        string
	regions, services []string
	logLevel          string
)

func init() {
	pflag.StringVarP(&outputFile, "output-file", "o", "inventory.csv", "path to the output file")
	pflag.StringSliceVarP(&regions, "regions", "r", data.ValidRegions, "regions to gather data from")
	pflag.StringSliceVarP(&services, "services", "s", data.ValidServices, "services to gather data from")
	pflag.StringVarP(&logLevel, "log-level", "l", "warning", "set the level of log output")
	pflag.Parse()

	initLogger()
}

func main() {
	awsData := data.New(logger, nil)
	awsData.Load(regions, services)

	f, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		logger.Fatal(err)
	}
	defer f.Close()

	// Create new csv inventory
	csv, err := inventory.NewCSV(f)
	if err != nil {
		logger.Fatal(err)
	}

	// Write stored rows to csv inventory
	var count int
	awsData.MapRows(func(row inventory.Row) error {
		count++
		return csv.WriteRow(row)
	})

	// Write file to disk
	logger.Infof("writing %d rows to %s", count, outputFile)
	csv.Flush()
}
