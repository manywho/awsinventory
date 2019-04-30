package main

import (
	"fmt"
	"os"

	"github.com/manywho/awsinventory/internal/awsdata"

	"github.com/manywho/awsinventory/internal/inventory"
	"github.com/spf13/pflag"
)

var (
	outputFile        string
	regions, services []string
	logLevel          string
	printRegions      bool
	printVersion      bool

	version, build string
)

func init() {
	pflag.StringVarP(&outputFile, "output-file", "o", "inventory.csv", "path to the output file")
	pflag.StringSliceVarP(&regions, "regions", "r", []string{}, "regions to gather data from")
	pflag.StringSliceVarP(&services, "services", "s", awsdata.ValidServices, "services to gather data from")
	pflag.BoolVar(&printRegions, "print-regions", false, "prints the available AWS regions")
	pflag.StringVarP(&logLevel, "log-level", "l", "warning", "set the level of log output")
	pflag.BoolVarP(&printVersion, "version", "v", false, "prints the version information")
	pflag.Parse()

	if printRegions {
		for _, r := range awsdata.ValidRegions {
			println(r)
		}
		os.Exit(0)
	}

	if printVersion {
		fmt.Printf("awsinventory %s+%s\n", version, build)
		os.Exit(0)
	}

	initLogger()
}

func main() {
	awsData := awsdata.New(logger, nil)
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
